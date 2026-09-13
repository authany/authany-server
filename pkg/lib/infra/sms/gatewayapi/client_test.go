package gatewayapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
)

func newTestClient(server *httptest.Server) *GatewayAPIClient {
	return &GatewayAPIClient{
		Client: server.Client(),
		GatewayAPICredentials: &config.GatewayAPICredentials{
			Endpoint: server.URL,
			APIToken: "api-token",
			Sender:   "MySender",
		},
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

func TestGatewayAPIClientSend(t *testing.T) {
	ctx := context.Background()

	Convey("GatewayAPIClient.Send", t, func() {
		Convey("success", func() {
			var receivedBody []byte
			var receivedHeader http.Header
			var receivedPath string
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				receivedBody, _ = io.ReadAll(r.Body)
				receivedHeader = r.Header.Clone()
				receivedPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				// The example response of https://gatewayapi.com/docs/apis/rest/
				_, _ = w.Write([]byte(`{"ids":[421332671],"usage":{"countries":{"DK":2},"currency":"DKK","total_cost":0.30}}`))
			})
			defer server.Close()

			err := newTestClient(server).Send(ctx, testSendOptions())
			So(err, ShouldBeNil)

			So(receivedPath, ShouldEqual, Path)

			var request SendRequest
			So(json.Unmarshal(receivedBody, &request), ShouldBeNil)
			So(request, ShouldResemble, SendRequest{
				Sender:     "MySender",
				Message:    "123456 is your code",
				Recipients: []Recipient{{MSISDN: "85298765432"}},
			})

			So(receivedHeader.Get("Content-Type"), ShouldEqual, ContentType)
			So(receivedHeader.Get("Accept"), ShouldEqual, Accept)
			So(receivedHeader.Get("Authorization"), ShouldEqual, "Token api-token")
		})

		Convey("invalid phone number", func() {
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				panic("unexpected request")
			})
			defer server.Close()

			opts := testSendOptions()
			opts.To = "not-a-phone-number"
			err := newTestClient(server).Send(ctx, opts)
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindInvalidPhoneNumber)
		})

		assertErrorKind := func(statusCode int, responseBody string, expected *apierrors.Kind, expectedCode string) {
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(statusCode)
				_, _ = w.Write([]byte(responseBody))
			})
			defer server.Close()

			err := newTestClient(server).Send(ctx, testSendOptions())
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderGatewayAPI)
			So(sendError.ProviderErrorCode, ShouldEqual, expectedCode)
			So(sendError.APIErrorKind, ShouldEqual, expected)
			So(sendError.DumpedResponse, ShouldNotBeEmpty)
		}

		Convey("authentication failed", func() {
			assertErrorKind(
				http.StatusUnauthorized,
				`{"code":"0x0229","incident_uuid":"incident-1","message":"Invalid token"}`,
				&smsapi.ErrKindAuthenticationFailed,
				"0x0229",
			)
		})

		Convey("unauthorized IP-address", func() {
			// The failed request example of https://gatewayapi.com/docs/apis/rest/
			assertErrorKind(
				http.StatusForbidden,
				`{"code": "0x0213","incident_uuid": "d8127429-fa0c-4316-b1f2-e610c3958f43","message": "Unauthorized IP-address: %1","variables": ["1.2.3.4"]}`,
				&smsapi.ErrKindAuthenticationFailed,
				"0x0213",
			)
		})

		Convey("delivery rejected", func() {
			assertErrorKind(
				http.StatusForbidden,
				`{"code":"0x0216","message":"Insufficient credit"}`,
				&smsapi.ErrKindDeliveryRejected,
				"0x0216",
			)
		})

		Convey("blacklisted MSISDN", func() {
			assertErrorKind(
				http.StatusForbidden,
				`{"code":"0x0215","message":"MSISDN 4512345678 blacklisted"}`,
				&smsapi.ErrKindDeliveryRejected,
				"0x0215",
			)
		})

		Convey("rate limited", func() {
			assertErrorKind(
				http.StatusTooManyRequests,
				`{"message":"Too many requests"}`,
				&smsapi.ErrKindRateLimited,
				"",
			)
		})

		Convey("undocumented error code", func() {
			assertErrorKind(
				http.StatusUnprocessableEntity,
				`{"code":"0x0318","message":"undocumented"}`,
				nil,
				"0x0318",
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
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderGatewayAPI)
			So(sendError.ProviderErrorCode, ShouldEqual, "")
			So(sendError.APIErrorKind, ShouldBeNil)
			So(sendError.DumpedResponse, ShouldNotBeEmpty)
		})

		Convey("non JSON success response", func() {
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("<html>ok</html>"))
			})
			defer server.Close()

			err := newTestClient(server).Send(ctx, testSendOptions())
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.APIErrorKind, ShouldBeNil)
			So(sendError.DumpedResponse, ShouldNotBeEmpty)
		})

		Convey("empty ids", func() {
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"ids":[]}`))
			})
			defer server.Close()

			err := newTestClient(server).Send(ctx, testSendOptions())
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.APIErrorKind, ShouldBeNil)
		})

		Convey("endpoint with a trailing slash", func() {
			var receivedPath string
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path
				_, _ = w.Write([]byte(`{"ids":[41008]}`))
			})
			defer server.Close()

			client := newTestClient(server)
			client.GatewayAPICredentials.Endpoint = server.URL + "/"
			err := client.Send(ctx, testSendOptions())
			So(err, ShouldBeNil)
			So(receivedPath, ShouldEqual, Path)
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
