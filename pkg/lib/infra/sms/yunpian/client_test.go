package yunpian

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
)

func TestYunpianClient(t *testing.T) {
	newClient := func(handler http.HandlerFunc) (*YunpianClient, func()) {
		server := httptest.NewServer(handler)
		client := &YunpianClient{
			Client:   server.Client(),
			Endpoint: server.URL,
			YunpianCredentials: &config.YunpianCredentials{
				APIKey: "testapikey",
			},
		}
		return client, server.Close
	}

	respondJSON := func(statusCode int, body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			_, _ = io.WriteString(w, body)
		}
	}

	send := func(client *YunpianClient, to string) error {
		return client.Send(context.Background(), smsapi.SendOptions{
			To:           to,
			Body:         "【云片网】您的验证码是 123456",
			TemplateName: "verification_sms.txt",
		})
	}

	Convey("YunpianClient.Send", t, func() {
		Convey("mainland China number", func() {
			var form url.Values
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				_ = r.ParseForm()
				form = r.PostForm
				// The Json 返回示例 of
				// https://www.yunpian.com/official/document/sms/zh_CN/domestic_single_send
				respondJSON(http.StatusOK, `{"code":0,"msg":"发送成功","count":1,"fee":0.05,"unit":"RMB","mobile":"13200000000","sid":3310228982}`)(w, r)
			})
			defer closeServer()

			err := send(client, "+8613800138000")
			So(err, ShouldBeNil)
			So(form.Get("apikey"), ShouldEqual, "testapikey")
			// The country calling code is stripped.
			So(form.Get("mobile"), ShouldEqual, "13800138000")
			// The rendered body is sent as-is.
			So(form.Get("text"), ShouldEqual, "【云片网】您的验证码是 123456")
		})

		Convey("international number", func() {
			var form url.Values
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				_ = r.ParseForm()
				form = r.PostForm
				// The Json 返回示例 of
				// https://www.yunpian.com/official/document/sms/zh_CN/intl_single_send
				respondJSON(http.StatusOK, `{"code":0,"msg":"发送成功","count":1,"fee":0.05,"unit":"RMB","mobile":"+93701234567","sid":3310228982}`)(w, r)
			})
			defer closeServer()

			err := send(client, "+85298765432")
			So(err, ShouldBeNil)
			// An international number keeps the plus sign.
			So(form.Get("mobile"), ShouldEqual, "+85298765432")
		})

		Convey("invalid phone number", func() {
			client, closeServer := newClient(respondJSON(http.StatusOK, `{"code":0}`))
			defer closeServer()

			err := send(client, "not-a-phone-number")
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderYunpian)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindInvalidPhoneNumber)
		})

		Convey("documented error response", func() {
			// The API 调用失败，返回错误结果示例 of
			// https://www.yunpian.com/official/document/sms/zh_CN/returnvalue_example
			client, closeServer := newClient(respondJSON(
				http.StatusBadRequest,
				`{"http_status_code": 400,"code": 3,"msg": "账户余额不足","detail": "账户需要充值，请充值后重试"}`,
			))
			defer closeServer()

			err := send(client, "+8613800138000")
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderYunpian)
			So(sendError.ProviderErrorCode, ShouldEqual, "3")
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindDeliveryRejected)
			So(len(sendError.DumpedResponse), ShouldBeGreaterThan, 0)
		})

		// The return codes are the ones of
		// https://www.yunpian.com/official/document/sms/zh_CN/returnvalue_common
		Convey("error codes", func() {
			cases := []struct {
				Code       int
				StatusCode int
				Kind       *apierrors.Kind
			}{
				{53, http.StatusOK, &smsapi.ErrKindRateLimited},
				{-5, http.StatusBadRequest, &smsapi.ErrKindRateLimited},
				{-1, http.StatusBadRequest, &smsapi.ErrKindAuthenticationFailed},
				{56, http.StatusOK, &smsapi.ErrKindInvalidPhoneNumber},
				{3, http.StatusOK, &smsapi.ErrKindDeliveryRejected},
				{-6, http.StatusOK, &smsapi.ErrKindUnsupportedRequest},
				// An unmapped return code is not mapped to any kind.
				{-50, http.StatusOK, nil},
			}
			for _, c := range cases {
				client, closeServer := newClient(respondJSON(
					c.StatusCode,
					`{"code":`+strconv.Itoa(c.Code)+`,"msg":"message","detail":"detail"}`,
				))

				err := send(client, "+8613800138000")
				So(err, ShouldNotBeNil)

				var sendError *smsapi.SendError
				So(errors.As(err, &sendError), ShouldBeTrue)
				So(sendError.ProviderType, ShouldEqual, config.SMSProviderYunpian)
				So(sendError.ProviderErrorCode, ShouldEqual, strconv.Itoa(c.Code))
				So(sendError.APIErrorKind, ShouldEqual, c.Kind)
				So(len(sendError.DumpedResponse), ShouldBeGreaterThan, 0)

				closeServer()
			}
		})

		Convey("timeout", func() {
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
			})
			defer closeServer()

			client.Client = &http.Client{Timeout: 10 * time.Millisecond}

			err := send(client, "+8613800138000")
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderYunpian)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindTimeout)
		})

		Convey("non-JSON response", func() {
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				_, _ = io.WriteString(w, "<html>bad gateway</html>")
			})
			defer closeServer()

			err := send(client, "+8613800138000")
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderYunpian)
			So(sendError.APIErrorKind, ShouldBeNil)
			So(len(sendError.DumpedResponse), ShouldBeGreaterThan, 0)
		})
	})
}
