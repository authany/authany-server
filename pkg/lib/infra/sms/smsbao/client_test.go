package smsbao

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
)

func TestParseSendResponse(t *testing.T) {
	Convey("ParseSendResponse", t, func() {
		code, err := ParseSendResponse([]byte("0"))
		So(err, ShouldBeNil)
		So(code, ShouldEqual, ResponseCodeOK)

		// The balance query API documents extra lines after the status code.
		code, err = ParseSendResponse([]byte("0\r\n100,200\r\n"))
		So(err, ShouldBeNil)
		So(code, ShouldEqual, ResponseCodeOK)

		code, err = ParseSendResponse([]byte(" 51 \n"))
		So(err, ShouldBeNil)
		So(code, ShouldEqual, ResponseCodeInvalidPhoneNumber)

		_, err = ParseSendResponse([]byte(""))
		So(err, ShouldNotBeNil)

		_, err = ParseSendResponse([]byte("<html>bad gateway</html>"))
		So(err, ShouldNotBeNil)
	})
}

func TestSmsbaoClient(t *testing.T) {
	newClient := func(handler http.HandlerFunc, credentials *config.SmsbaoCredentials) (*SmsbaoClient, func()) {
		server := httptest.NewServer(handler)
		client := &SmsbaoClient{
			Client:            server.Client(),
			Endpoint:          server.URL,
			SmsbaoCredentials: credentials,
		}
		return client, server.Close
	}

	credentials := func() *config.SmsbaoCredentials {
		return &config.SmsbaoCredentials{
			Username:         "tom",
			PasswordOrAPIKey: "9b11127a9701975c734b8aee81ee3526",
			GoodsID:          "123456",
		}
	}

	// The response body is the plain text status code documented by
	// https://www.smsbao.com/openapi/213.html and
	// https://www.smsbao.com/openapi/299.html, where "0" means success.
	respondText := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = io.WriteString(w, body)
		}
	}

	send := func(client *SmsbaoClient, to string, templateName string) error {
		return client.Send(context.Background(), smsapi.SendOptions{
			To:                to,
			Body:              "【短信宝】您的验证码是123456",
			TemplateName:      templateName,
			TemplateVariables: &smsapi.TemplateVariables{Code: "123456"},
		})
	}

	Convey("SmsbaoClient.Send", t, func() {
		Convey("success with a mainland China number", func() {
			var path string
			var query url.Values
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				path = r.URL.Path
				query = r.URL.Query()
				respondText("0")(w, r)
			}, credentials())
			defer closeServer()

			err := send(client, "+8613800138000", "verification_sms.txt")
			So(err, ShouldBeNil)
			So(path, ShouldEqual, PathDomestic)
			So(query.Get("u"), ShouldEqual, "tom")
			So(query.Get("p"), ShouldEqual, "9b11127a9701975c734b8aee81ee3526")
			// The country calling code is stripped.
			So(query.Get("m"), ShouldEqual, "13800138000")
			So(query.Get("c"), ShouldEqual, "【短信宝】您的验证码是123456")
			So(query.Get("g"), ShouldEqual, "123456")
		})

		Convey("success with an international number", func() {
			var path string
			var query url.Values
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				path = r.URL.Path
				query = r.URL.Query()
				respondText("0")(w, r)
			}, credentials())
			defer closeServer()

			err := send(client, "+60901234567", "verification_sms.txt")
			So(err, ShouldBeNil)
			So(path, ShouldEqual, PathInternational)
			// The international API expects the plus sign.
			So(query.Get("m"), ShouldEqual, "+60901234567")
			// g is documented by the domestic API only.
			So(query.Get("g"), ShouldEqual, "")
		})

		Convey("goods_id is absent", func() {
			var query url.Values
			c := credentials()
			c.GoodsID = ""
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				query = r.URL.Query()
				respondText("0")(w, r)
			}, c)
			defer closeServer()

			err := send(client, "+8613800138000", "verification_sms.txt")
			So(err, ShouldBeNil)
			So(query.Get("g"), ShouldEqual, "")
		})

		Convey("the link SMS template is supported", func() {
			var query url.Values
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				query = r.URL.Query()
				respondText("0")(w, r)
			}, credentials())
			defer closeServer()

			err := client.Send(context.Background(), smsapi.SendOptions{
				To:           "+8613800138000",
				Body:         "【短信宝】请点击 https://example.com 重设密码",
				TemplateName: "forgot_password_sms.txt",
			})
			So(err, ShouldBeNil)
			So(query.Get("c"), ShouldEqual, "【短信宝】请点击 https://example.com 重设密码")
		})

		Convey("invalid phone number", func() {
			client, closeServer := newClient(respondText("0"), credentials())
			defer closeServer()

			err := send(client, "not-a-phone-number", "verification_sms.txt")
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderSmsbao)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindInvalidPhoneNumber)
		})

		Convey("error codes", func() {
			// The 错误代码列表 of https://www.smsbao.com/openapi/213.html
			cases := []struct {
				Code string
				Kind *apierrors.Kind
			}{
				{"30", &smsapi.ErrKindAuthenticationFailed},
				{"40", &smsapi.ErrKindAuthenticationFailed},
				{"43", &smsapi.ErrKindAuthenticationFailed},
				{"41", &smsapi.ErrKindDeliveryRejected},
				{"50", &smsapi.ErrKindDeliveryRejected},
				{"51", &smsapi.ErrKindInvalidPhoneNumber},
				// SMSBao does not document a rate limit code,
				// so an undocumented code is not mapped.
				{"99", nil},
			}
			for _, c := range cases {
				client, closeServer := newClient(respondText(c.Code), credentials())

				err := send(client, "+8613800138000", "verification_sms.txt")
				So(err, ShouldNotBeNil)

				var sendError *smsapi.SendError
				So(errors.As(err, &sendError), ShouldBeTrue)
				So(sendError.ProviderType, ShouldEqual, config.SMSProviderSmsbao)
				So(sendError.ProviderErrorCode, ShouldEqual, c.Code)
				So(sendError.APIErrorKind, ShouldEqual, c.Kind)
				So(len(sendError.DumpedResponse), ShouldBeGreaterThan, 0)

				closeServer()
			}
		})

		Convey("unexpected response body", func() {
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				_, _ = io.WriteString(w, "<html>bad gateway</html>")
			}, credentials())
			defer closeServer()

			err := send(client, "+8613800138000", "verification_sms.txt")
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderSmsbao)
			So(sendError.APIErrorKind, ShouldBeNil)
			So(len(sendError.DumpedResponse), ShouldBeGreaterThan, 0)
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
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderSmsbao)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindTimeout)

			shouldNotContainCredentials(err)
		})

		Convey("the error of a failed connection does not contain the credentials", func() {
			client, closeServer := newClient(respondText("0"), credentials())
			// The server is closed so that the connection fails.
			closeServer()

			err := send(client, "+8613800138000", "verification_sms.txt")
			So(err, ShouldNotBeNil)

			shouldNotContainCredentials(err)
		})

		Convey("the credentials do not reach the instrumentation layer", func() {
			var query url.Values
			client, closeServer := newClient(func(w http.ResponseWriter, r *http.Request) {
				query = r.URL.Query()
				respondText("0")(w, r)
			}, credentials())
			defer closeServer()

			// recording sits where otelhttp sits, that is,
			// between the two RoundTrippers of the pair.
			recording := &recordingRoundTripper{
				Base: restoreQueryRoundTripper{Base: client.Client.Transport},
			}
			client.Client = &http.Client{Transport: stripQueryRoundTripper{Base: recording}}

			err := send(client, "+8613800138000", "verification_sms.txt")
			So(err, ShouldBeNil)

			So(recording.URLs, ShouldHaveLength, 1)
			So(recording.URLs[0], ShouldEndWith, PathDomestic)
			So(recording.URLs[0], ShouldNotContainSubstring, "?")
			So(recording.URLs[0], ShouldNotContainSubstring, "tom")
			So(recording.URLs[0], ShouldNotContainSubstring, "9b11127a9701975c734b8aee81ee3526")

			// The request actually sent still carries the query.
			So(query.Get("u"), ShouldEqual, "tom")
			So(query.Get("p"), ShouldEqual, "9b11127a9701975c734b8aee81ee3526")
			So(query.Get("m"), ShouldEqual, "13800138000")
		})
	})
}

func TestNewSmsbaoClient(t *testing.T) {
	Convey("NewSmsbaoClient", t, func() {
		client := NewSmsbaoClient(&config.SmsbaoCredentials{
			Username:         "tom",
			PasswordOrAPIKey: "9b11127a9701975c734b8aee81ee3526",
		})
		// The query, which carries the credentials, is taken out of the URL
		// before the request reaches otelhttp.
		_, ok := client.Client.Transport.(stripQueryRoundTripper)
		So(ok, ShouldBeTrue)
	})
}

func shouldNotContainCredentials(err error) {
	message := err.Error()
	// The path is kept, so the assertions below are not vacuous.
	So(message, ShouldContainSubstring, PathDomestic)
	So(message, ShouldNotContainSubstring, "u=")
	So(message, ShouldNotContainSubstring, "p=")
	So(message, ShouldNotContainSubstring, "tom")
	So(message, ShouldNotContainSubstring, "9b11127a9701975c734b8aee81ee3526")
}

type recordingRoundTripper struct {
	Base http.RoundTripper
	URLs []string
}

func (t *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	t.URLs = append(t.URLs, req.URL.String())
	return t.Base.RoundTrip(req)
}
