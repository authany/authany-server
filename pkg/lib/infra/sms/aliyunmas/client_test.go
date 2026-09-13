package aliyunmas

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
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/aliyun"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

// TestSignatureSDKVector asserts that the shared RPC style signature
// reproduces, byte for byte, the Signature parameter of a SendSmsVerifyCode
// request sent by the official SDK
// github.com/aliyun/alibaba-cloud-sdk-go/services/dypnsapi (core 1.63.22).
//
// The request was recorded by pointing the SDK at a local httptest server.
// The credentials are fake. The SDK additionally sends an empty SignatureType
// parameter, and puts the parameters in the query string instead of the
// request body; both are accepted by the RPC style API. The recorded
// parameters are fed to SignRPCRequest as-is here.
func TestSignatureSDKVector(t *testing.T) {
	Convey("official SDK recorded request", t, func() {
		values := url.Values{}
		values.Set("AccessKeyId", "LTAI5tTESTACCESSKEYID")
		values.Set("Action", "SendSmsVerifyCode")
		values.Set("CountryCode", "86")
		values.Set("Format", "JSON")
		values.Set("PhoneNumber", "13000000000")
		values.Set("RegionId", "cn-hangzhou")
		values.Set("SignName", "Authany 测试")
		values.Set("SignatureMethod", "HMAC-SHA1")
		values.Set("SignatureNonce", "d8525ea95c4287bbb3c3c76898d44272")
		values.Set("SignatureType", "")
		values.Set("SignatureVersion", "1.0")
		values.Set("TemplateCode", "SMS_987654321")
		values.Set("TemplateParam", `{"code":"123456"}`)
		values.Set("Timestamp", "2026-09-13T08:56:57Z")
		values.Set("Version", "2017-05-25")

		So(aliyun.SignRPCRequest("POST", values, "TESTACCESSKEYSECRETTESTACCESSKE"), ShouldEqual, "7uejeSsPry0Tscoup4tjn2Ay/Nk=")
	})
}

func TestAliyunMASClient(t *testing.T) {
	newClient := func(handler http.HandlerFunc, credentials *config.AliyunMASCredentials) (*AliyunMASClient, func()) {
		server := httptest.NewServer(handler)
		client := &AliyunMASClient{
			Client:               server.Client(),
			Clock:                clock.NewMockClockAt("2023-03-13T08:34:30Z"),
			Endpoint:             server.URL,
			AliyunMASCredentials: credentials,
		}
		return client, server.Close
	}

	credentials := func() *config.AliyunMASCredentials {
		return &config.AliyunMASCredentials{
			AccessKeyID:     "testid",
			AccessKeySecret: "testsecret",
			SMSTemplateCodeConfig: config.SMSTemplateCodeConfig{
				SignName:     "sign",
				TemplateCode: "100001",
				TemplateCodes: map[string]string{
					"setup_primary_oob_sms.txt": "100002",
				},
			},
		}
	}

	respondJSON := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, body)
		}
	}

	send := func(client *AliyunMASClient, to string, templateName string) error {
		return client.Send(context.Background(), smsapi.SendOptions{
			To:                to,
			TemplateName:      templateName,
			TemplateVariables: &smsapi.TemplateVariables{Code: "123456"},
		})
	}

	Convey("AliyunMASClient.Send", t, func() {
		Convey("success", func() {
			var form url.Values
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				_ = r.ParseForm()
				form = r.PostForm
				respondJSON(`{"Code":"OK","Message":"成功","Success":true,"RequestId":"req","Model":{"BizId":"biz","RequestId":"req"}}`)(w, r)
			}, credentials())
			defer closeServer()

			err := send(client, "+8613800138000", "verification_sms.txt")
			So(err, ShouldBeNil)
			// The country calling code is stripped.
			So(form.Get("PhoneNumber"), ShouldEqual, "13800138000")
			So(form.Get("CountryCode"), ShouldEqual, "86")
			So(form.Get("SignName"), ShouldEqual, "sign")
			So(form.Get("TemplateCode"), ShouldEqual, "100001")
			So(form.Get("TemplateParam"), ShouldEqual, `{"code":"123456"}`)
			So(form.Get("Action"), ShouldEqual, "SendSmsVerifyCode")
			So(form.Get("Version"), ShouldEqual, "2017-05-25")
			// The timestamp comes from the injected clock.
			So(form.Get("Timestamp"), ShouldEqual, "2023-03-13T08:34:30Z")

			signedValues := url.Values{}
			for key, value := range form {
				if key != "Signature" {
					signedValues[key] = value
				}
			}
			So(form.Get("Signature"), ShouldEqual, aliyun.SignRPCRequest("POST", signedValues, "testsecret"))
		})

		Convey("per template name override", func() {
			var form url.Values
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				_ = r.ParseForm()
				form = r.PostForm
				respondJSON(`{"Code":"OK","Success":true}`)(w, r)
			}, credentials())
			defer closeServer()

			err := send(client, "+8613800138000", "setup_primary_oob_sms.txt")
			So(err, ShouldBeNil)
			So(form.Get("TemplateCode"), ShouldEqual, "100002")
		})

		Convey("link SMS is unsupported", func() {
			client, closeServer := newClient(respondJSON(`{"Code":"OK","Success":true}`), credentials())
			defer closeServer()

			err := send(client, "+8613800138000", "forgot_password_sms.txt")
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderAliyunMAS)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindUnsupportedRequest)
		})

		Convey("invalid phone number", func() {
			cases := []string{
				// Unparsable.
				"not-a-phone-number",
				// The service supports mainland China phone numbers only.
				"+85298765432",
			}
			for _, to := range cases {
				client, closeServer := newClient(respondJSON(`{"Code":"OK","Success":true}`), credentials())

				err := send(client, to, "verification_sms.txt")
				So(err, ShouldNotBeNil)

				var sendError *smsapi.SendError
				So(errors.As(err, &sendError), ShouldBeTrue)
				So(sendError.ProviderType, ShouldEqual, config.SMSProviderAliyunMAS)
				So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindInvalidPhoneNumber)

				closeServer()
			}
		})

		Convey("error codes", func() {
			cases := []struct {
				Code string
				Kind *apierrors.Kind
			}{
				{"BUSINESS_LIMIT_CONTROL", &smsapi.ErrKindRateLimited},
				{"FREQUENCY_FAIL", &smsapi.ErrKindRateLimited},
				{"MOBILE_NUMBER_ILLEGAL", &smsapi.ErrKindInvalidPhoneNumber},
				{"InvalidAccessKeyId.NotFound", &smsapi.ErrKindAuthenticationFailed},
				{"SignatureDoesNotMatch", &smsapi.ErrKindAuthenticationFailed},
				{"Forbidden.AccessKeyDisabled", &smsapi.ErrKindAuthenticationFailed},
				{"FUNCTION_NOT_OPENED", &smsapi.ErrKindDeliveryRejected},
				{"isv.SMS_SIGNATURE_ILLEGAL", &smsapi.ErrKindDeliveryRejected},
				{"isv.SMS_TEMPLATE_ILLEGAL", &smsapi.ErrKindDeliveryRejected},
				// INVALID_PARAMETERS is too generic to be mapped.
				{"INVALID_PARAMETERS", nil},
				// An unknown error code is not mapped.
				{"SomeUnknownErrorCode", nil},
			}
			for _, c := range cases {
				client, closeServer := newClient(
					respondJSON(`{"Code":"`+c.Code+`","Message":"message","Success":false,"RequestId":"req"}`),
					credentials(),
				)

				err := send(client, "+8613800138000", "verification_sms.txt")
				So(err, ShouldNotBeNil)

				var sendError *smsapi.SendError
				So(errors.As(err, &sendError), ShouldBeTrue)
				So(sendError.ProviderType, ShouldEqual, config.SMSProviderAliyunMAS)
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
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderAliyunMAS)
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
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderAliyunMAS)
			So(sendError.APIErrorKind, ShouldBeNil)
			So(len(sendError.DumpedResponse), ShouldBeGreaterThan, 0)
		})
	})
}
