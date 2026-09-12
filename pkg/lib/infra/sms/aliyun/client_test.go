package aliyun

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func TestSignRPCRequest(t *testing.T) {
	Convey("SignRPCRequest", t, func() {
		// The example in
		// https://www.alibabacloud.com/help/en/sdk/product-overview/rpc-mechanism
		values := url.Values{}
		values.Set("AccessKeyId", "testid")
		values.Set("Action", "DescribeDedicatedHosts")
		values.Set("Format", "JSON")
		values.Set("RegionId", "cn-beijing")
		values.Set("SignatureMethod", "HMAC-SHA1")
		values.Set("SignatureNonce", "edb2b34af0af9a6d14deaf7c1a5315eb")
		values.Set("SignatureVersion", "1.0")
		values.Set("Timestamp", "2023-03-13T08:34:30Z")
		values.Set("Version", "2014-05-26")

		So(CanonicalizedQueryString(values), ShouldEqual, "AccessKeyId=testid&Action=DescribeDedicatedHosts&Format=JSON&RegionId=cn-beijing&SignatureMethod=HMAC-SHA1&SignatureNonce=edb2b34af0af9a6d14deaf7c1a5315eb&SignatureVersion=1.0&Timestamp=2023-03-13T08%3A34%3A30Z&Version=2014-05-26")

		So(StringToSign("GET", values), ShouldEqual, "GET&%2F&AccessKeyId%3Dtestid%26Action%3DDescribeDedicatedHosts%26Format%3DJSON%26RegionId%3Dcn-beijing%26SignatureMethod%3DHMAC-SHA1%26SignatureNonce%3Dedb2b34af0af9a6d14deaf7c1a5315eb%26SignatureVersion%3D1.0%26Timestamp%3D2023-03-13T08%253A34%253A30Z%26Version%3D2014-05-26")

		So(SignRPCRequest("GET", values, "testsecret"), ShouldEqual, "9NaGiOspFP5UPcwX8Iwt2YJXXuk=")
	})

	Convey("PercentEncode", t, func() {
		So(PercentEncode(" "), ShouldEqual, "%20")
		So(PercentEncode("*"), ShouldEqual, "%2A")
		So(PercentEncode("~"), ShouldEqual, "~")
		So(PercentEncode("-_."), ShouldEqual, "-_.")
	})

	Convey("BuildCommonParams", t, func() {
		values := BuildCommonParams(CommonParams{
			AccessKeyID:    "testid",
			Action:         ActionSendSms,
			Version:        APIVersion,
			RegionID:       DefaultRegionID,
			Timestamp:      time.Date(2023, 3, 13, 8, 34, 30, 0, time.UTC),
			SignatureNonce: "nonce",
		})
		So(values.Get("AccessKeyId"), ShouldEqual, "testid")
		So(values.Get("Action"), ShouldEqual, "SendSms")
		So(values.Get("Version"), ShouldEqual, "2017-05-25")
		So(values.Get("RegionId"), ShouldEqual, "cn-hangzhou")
		So(values.Get("Format"), ShouldEqual, "JSON")
		So(values.Get("SignatureMethod"), ShouldEqual, "HMAC-SHA1")
		So(values.Get("SignatureVersion"), ShouldEqual, "1.0")
		So(values.Get("SignatureNonce"), ShouldEqual, "nonce")
		So(values.Get("Timestamp"), ShouldEqual, "2023-03-13T08:34:30Z")
	})
}

func TestAliyunClient(t *testing.T) {
	newClient := func(handler http.HandlerFunc, credentials *config.AliyunCredentials) (*AliyunClient, func()) {
		server := httptest.NewServer(handler)
		client := &AliyunClient{
			Client:            server.Client(),
			Clock:             clock.NewMockClockAt("2023-03-13T08:34:30Z"),
			Endpoint:          server.URL,
			AliyunCredentials: credentials,
		}
		return client, server.Close
	}

	credentials := func() *config.AliyunCredentials {
		return &config.AliyunCredentials{
			AccessKeyID:     "testid",
			AccessKeySecret: "testsecret",
			SMSTemplateCodeConfig: config.SMSTemplateCodeConfig{
				SignName:     "sign",
				TemplateCode: "SMS_DEFAULT",
				TemplateCodes: map[string]string{
					"setup_primary_oob_sms.txt": "SMS_SETUP",
				},
			},
			OverseasTemplateCode: "SMS_OVERSEAS",
		}
	}

	respondJSON := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, body)
		}
	}

	send := func(client *AliyunClient, to string, templateName string) error {
		return client.Send(context.Background(), smsapi.SendOptions{
			To:                to,
			TemplateName:      templateName,
			TemplateVariables: &smsapi.TemplateVariables{Code: "123456"},
		})
	}

	Convey("AliyunClient.Send", t, func() {
		Convey("success", func() {
			var form url.Values
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				_ = r.ParseForm()
				form = r.PostForm
				respondJSON(`{"Code":"OK","Message":"OK","RequestId":"req","BizId":"biz"}`)(w, r)
			}, credentials())
			defer closeServer()

			err := send(client, "+85298765432", "setup_primary_oob_sms.txt")
			So(err, ShouldBeNil)
			So(form.Get("PhoneNumbers"), ShouldEqual, "85298765432")
			So(form.Get("SignName"), ShouldEqual, "sign")
			// A non +86 number uses the overseas template code.
			So(form.Get("TemplateCode"), ShouldEqual, "SMS_OVERSEAS")
			So(form.Get("TemplateParam"), ShouldEqual, `{"code":"123456"}`)
			So(form.Get("Action"), ShouldEqual, "SendSms")
			// The timestamp comes from the injected clock.
			So(form.Get("Timestamp"), ShouldEqual, "2023-03-13T08:34:30Z")

			signedValues := url.Values{}
			for key, value := range form {
				if key != "Signature" {
					signedValues[key] = value
				}
			}
			So(form.Get("Signature"), ShouldEqual, SignRPCRequest("POST", signedValues, "testsecret"))
		})

		Convey("mainland China number", func() {
			var form url.Values
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				_ = r.ParseForm()
				form = r.PostForm
				respondJSON(`{"Code":"OK"}`)(w, r)
			}, credentials())
			defer closeServer()

			err := send(client, "+8613800138000", "setup_primary_oob_sms.txt")
			So(err, ShouldBeNil)
			// The country calling code is stripped.
			So(form.Get("PhoneNumbers"), ShouldEqual, "13800138000")
			// The per template name override takes precedence.
			So(form.Get("TemplateCode"), ShouldEqual, "SMS_SETUP")
		})

		Convey("overseas_template_code is absent", func() {
			var form url.Values
			c := credentials()
			c.OverseasTemplateCode = ""
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				_ = r.ParseForm()
				form = r.PostForm
				respondJSON(`{"Code":"OK"}`)(w, r)
			}, c)
			defer closeServer()

			err := send(client, "+85298765432", "verification_sms.txt")
			So(err, ShouldBeNil)
			So(form.Get("TemplateCode"), ShouldEqual, "SMS_DEFAULT")
		})

		Convey("link SMS is unsupported", func() {
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				respondJSON(`{"Code":"OK"}`)(w, r)
			}, credentials())
			defer closeServer()

			err := send(client, "+8613800138000", "forgot_password_sms.txt")
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindUnsupportedRequest)
		})

		Convey("invalid phone number", func() {
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				respondJSON(`{"Code":"OK"}`)(w, r)
			}, credentials())
			defer closeServer()

			err := send(client, "not-a-phone-number", "verification_sms.txt")
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindInvalidPhoneNumber)
		})

		Convey("error codes", func() {
			cases := []struct {
				Code string
				Kind *apierrors.Kind
			}{
				{"isv.BUSINESS_LIMIT_CONTROL", &smsapi.ErrKindRateLimited},
				{"isv.MOBILE_NUMBER_ILLEGAL", &smsapi.ErrKindInvalidPhoneNumber},
				{"InvalidAccessKeyId.NotFound", &smsapi.ErrKindAuthenticationFailed},
				{"SignatureDoesNotMatch", &smsapi.ErrKindAuthenticationFailed},
				{"isv.SMS_SIGNATURE_ILLEGAL", &smsapi.ErrKindDeliveryRejected},
				{"isv.SMS_TEMPLATE_ILLEGAL", &smsapi.ErrKindDeliveryRejected},
				// An unknown error code is not mapped.
				{"SomeUnknownErrorCode", nil},
			}
			for _, c := range cases {
				client, closeServer := newClient(
					respondJSON(`{"Code":"`+c.Code+`","Message":"message","RequestId":"req"}`),
					credentials(),
				)

				err := send(client, "+8613800138000", "verification_sms.txt")
				So(err, ShouldNotBeNil)

				var sendError *smsapi.SendError
				So(errors.As(err, &sendError), ShouldBeTrue)
				So(sendError.ProviderType, ShouldEqual, config.SMSProviderAliyun)
				So(sendError.ProviderErrorCode, ShouldEqual, c.Code)
				So(sendError.APIErrorKind, ShouldEqual, c.Kind)
				So(len(sendError.DumpedResponse), ShouldBeGreaterThan, 0)

				closeServer()
			}
		})

		Convey("timeout", func() {
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
			}, credentials())
			defer closeServer()

			client.Client = &http.Client{Timeout: 10 * time.Millisecond}

			err := send(client, "+8613800138000", "verification_sms.txt")
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderAliyun)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindTimeout)
		})

		Convey("non-JSON response", func() {
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				_, _ = io.WriteString(w, "<html>bad gateway</html>")
			}, credentials())
			defer closeServer()

			err := send(client, "+8613800138000", "verification_sms.txt")
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.APIErrorKind, ShouldBeNil)
			So(len(sendError.DumpedResponse), ShouldBeGreaterThan, 0)
		})
	})
}
