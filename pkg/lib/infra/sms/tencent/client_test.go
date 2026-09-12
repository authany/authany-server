package tencent

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

// The example of https://cloud.tencent.com/document/api/213/30654
// The document redacts the SecretId and the SecretKey,
// so only the values that do not depend on them are asserted here.
func TestSignatureExampleVector(t *testing.T) {
	const exampleHost = "cvm.tencentcloudapi.com"
	const exampleAction = "DescribeInstances"
	const examplePayload = `{"Limit": 1, "Filters": [{"Values": ["\u672a\u547d\u540d"], "Name": "instance-name"}]}`
	const exampleTimestamp int64 = 1551113065

	Convey("TC3-HMAC-SHA256", t, func() {
		Convey("hashed request payload", func() {
			So(sha256hex(examplePayload), ShouldEqual, "35e9c5b0e3ae67532d3c9f17ead6c90222632e5b1ff7f6e89887f1398934f064")
		})

		Convey("canonical request", func() {
			canonicalRequest := makeCanonicalRequest(exampleHost, exampleAction, examplePayload)
			So(canonicalRequest, ShouldEqual, strings.Join([]string{
				"POST",
				"/",
				"",
				"content-type:application/json; charset=utf-8",
				"host:cvm.tencentcloudapi.com",
				"x-tc-action:describeinstances",
				"",
				"content-type;host;x-tc-action",
				"35e9c5b0e3ae67532d3c9f17ead6c90222632e5b1ff7f6e89887f1398934f064",
			}, "\n"))
			So(sha256hex(canonicalRequest), ShouldEqual, "7019a55be8395899b900fb5564e4200d984910f34794a27cb3fb7d10ff6a1e84")
		})

		Convey("string to sign", func() {
			date := time.Unix(exampleTimestamp, 0).UTC().Format("2006-01-02")
			So(date, ShouldEqual, "2019-02-25")

			credentialScope := makeCredentialScope(date, "cvm")
			So(credentialScope, ShouldEqual, "2019-02-25/cvm/tc3_request")

			canonicalRequest := makeCanonicalRequest(exampleHost, exampleAction, examplePayload)
			So(makeStringToSign(exampleTimestamp, credentialScope, canonicalRequest), ShouldEqual, strings.Join([]string{
				"TC3-HMAC-SHA256",
				"1551113065",
				"2019-02-25/cvm/tc3_request",
				"7019a55be8395899b900fb5564e4200d984910f34794a27cb3fb7d10ff6a1e84",
			}, "\n"))
		})

		// The SecretKey of the example is redacted in the document,
		// so the signature is asserted against a golden value instead.
		Convey("authorization", func() {
			authorization := makeAuthorization("AKIDEXAMPLE", "SECRETKEYEXAMPLE", exampleHost, "cvm", exampleAction, exampleTimestamp, examplePayload)
			So(authorization, ShouldEqual, "TC3-HMAC-SHA256 Credential=AKIDEXAMPLE/2019-02-25/cvm/tc3_request, SignedHeaders=content-type;host;x-tc-action, Signature=8cd2b6c0a8c5b1e6aa358096ca9df41f14b09d4b3a5f85014a31f8c0947410f5")
		})
	})
}

func newTestClient(server *httptest.Server) *TencentClient {
	return &TencentClient{
		Client: server.Client(),
		Clock:  clock.NewMockClockAt("2019-02-25T08:44:25Z"),
		TencentCredentials: &config.TencentCredentials{
			SecretID:  "secret-id",
			SecretKey: "secret-key",
			SDKAppID:  "1400000000",
			Region:    "ap-guangzhou",
			SMSTemplateCodeConfig: config.SMSTemplateCodeConfig{
				SignName:     "MySign",
				TemplateCode: "1000",
				TemplateCodes: map[string]string{
					"setup_primary_oob_sms.txt": "2000",
				},
			},
		},
		Endpoint: server.URL,
	}
}

func testSendOptions() smsapi.SendOptions {
	return smsapi.SendOptions{
		To:                "+85298765432",
		Body:              "123456 is your code",
		TemplateName:      "verification_sms.txt",
		LanguageTag:       "en",
		TemplateVariables: &smsapi.TemplateVariables{Code: "123456"},
	}
}

func newServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestTencentClientSend(t *testing.T) {
	ctx := context.Background()

	Convey("TencentClient.Send", t, func() {
		Convey("success", func() {
			var receivedBody []byte
			var receivedHeader http.Header
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				receivedBody, _ = io.ReadAll(r.Body)
				receivedHeader = r.Header.Clone()
				receivedHeader.Set("Host", r.Host)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"Response":{"SendStatusSet":[{"SerialNo":"1","PhoneNumber":"+85298765432","Code":"Ok","Message":"send success"}],"RequestId":"req-1"}}`))
			})
			defer server.Close()

			client := newTestClient(server)
			err := client.Send(ctx, testSendOptions())
			So(err, ShouldBeNil)

			var request SendRequest
			So(json.Unmarshal(receivedBody, &request), ShouldBeNil)
			So(request, ShouldResemble, SendRequest{
				PhoneNumberSet:   []string{"+85298765432"},
				SmsSdkAppId:      "1400000000",
				SignName:         "MySign",
				TemplateId:       "1000",
				TemplateParamSet: []string{"123456"},
			})

			So(receivedHeader.Get("Host"), ShouldEqual, "sms.tencentcloudapi.com")
			So(receivedHeader.Get("Content-Type"), ShouldEqual, "application/json; charset=utf-8")
			So(receivedHeader.Get("X-TC-Action"), ShouldEqual, "SendSms")
			So(receivedHeader.Get("X-TC-Version"), ShouldEqual, "2021-01-11")
			So(receivedHeader.Get("X-TC-Region"), ShouldEqual, "ap-guangzhou")
			So(receivedHeader.Get("X-TC-Timestamp"), ShouldEqual, "1551084265")
			So(receivedHeader.Get("Authorization"), ShouldEqual, makeAuthorization(
				"secret-id", "secret-key", Host, Service, Action, 1551084265, string(receivedBody),
			))
		})

		Convey("template code override", func() {
			var receivedBody []byte
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				receivedBody, _ = io.ReadAll(r.Body)
				_, _ = w.Write([]byte(`{"Response":{"SendStatusSet":[{"Code":"Ok"}]}}`))
			})
			defer server.Close()

			opts := testSendOptions()
			opts.TemplateName = "setup_primary_oob_sms.txt"
			err := newTestClient(server).Send(ctx, opts)
			So(err, ShouldBeNil)

			var request SendRequest
			So(json.Unmarshal(receivedBody, &request), ShouldBeNil)
			So(request.TemplateId, ShouldEqual, "2000")
		})

		Convey("the link template is unsupported", func() {
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				panic("unexpected request")
			})
			defer server.Close()

			opts := testSendOptions()
			opts.TemplateName = TemplateNameForgotPasswordSMS
			err := newTestClient(server).Send(ctx, opts)
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindUnsupportedRequest)
		})

		assertErrorKind := func(responseBody string, expected *apierrors.Kind, expectedCode string) {
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(responseBody))
			})
			defer server.Close()

			err := newTestClient(server).Send(ctx, testSendOptions())
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderTencent)
			So(sendError.ProviderErrorCode, ShouldEqual, expectedCode)
			So(sendError.APIErrorKind, ShouldEqual, expected)
			So(sendError.DumpedResponse, ShouldNotBeEmpty)
		}

		Convey("rate limited", func() {
			assertErrorKind(
				`{"Response":{"SendStatusSet":[{"Code":"LimitExceeded.PhoneNumberDailyLimit","Message":"limit"}]}}`,
				&smsapi.ErrKindRateLimited,
				"LimitExceeded.PhoneNumberDailyLimit",
			)
		})

		Convey("authentication failed", func() {
			assertErrorKind(
				`{"Response":{"Error":{"Code":"AuthFailure.SignatureFailure","Message":"bad signature"},"RequestId":"req-1"}}`,
				&smsapi.ErrKindAuthenticationFailed,
				"AuthFailure.SignatureFailure",
			)
		})

		Convey("invalid phone number", func() {
			assertErrorKind(
				`{"Response":{"SendStatusSet":[{"Code":"InvalidParameterValue.IncorrectPhoneNumber","Message":"bad number"}]}}`,
				&smsapi.ErrKindInvalidPhoneNumber,
				"InvalidParameterValue.IncorrectPhoneNumber",
			)
		})

		Convey("delivery rejected", func() {
			assertErrorKind(
				`{"Response":{"Error":{"Code":"FailedOperation.TemplateIncorrectOrUnapproved","Message":"bad template"}}}`,
				&smsapi.ErrKindDeliveryRejected,
				"FailedOperation.TemplateIncorrectOrUnapproved",
			)
		})

		Convey("unknown error code", func() {
			assertErrorKind(
				`{"Response":{"SendStatusSet":[{"Code":"UnknownError","Message":"unknown"}]}}`,
				nil,
				"UnknownError",
			)
		})

		Convey("non JSON response", func() {
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte("<html>bad gateway</html>"))
			})
			defer server.Close()

			err := newTestClient(server).Send(ctx, testSendOptions())
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderTencent)
			So(sendError.APIErrorKind, ShouldBeNil)
			So(sendError.DumpedResponse, ShouldNotBeEmpty)
		})

		Convey("empty SendStatusSet", func() {
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"Response":{"RequestId":"req-1"}}`))
			})
			defer server.Close()

			err := newTestClient(server).Send(ctx, testSendOptions())
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.APIErrorKind, ShouldBeNil)
		})

		Convey("timeout", func() {
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
			})
			defer server.Close()

			client := newTestClient(server)
			client.Client = &http.Client{Timeout: 10 * time.Millisecond}
			err := client.Send(ctx, testSendOptions())
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindTimeout)
		})
	})
}
