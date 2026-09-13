package smsaero

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

func newTestClient(server *httptest.Server) *SmsAeroClient {
	return &SmsAeroClient{
		Client: server.Client(),
		SmsAeroCredentials: &config.SmsAeroCredentials{
			Email:      "user@example.com",
			APIKey:     "api-key",
			SenderName: "SMS Aero",
		},
		Endpoint: server.URL,
	}
}

func testSendOptions() smsapi.SendOptions {
	return smsapi.SendOptions{
		To:                "+79990000000",
		Body:              "123456 is your code",
		TemplateName:      "verification_sms.txt",
		LanguageTag:       "en",
		TemplateVariables: &smsapi.TemplateVariables{Code: "123456"},
	}
}

func newServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestSmsAeroClientSend(t *testing.T) {
	ctx := context.Background()

	Convey("SmsAeroClient.Send", t, func() {
		Convey("success", func() {
			var receivedBody []byte
			var receivedHeader http.Header
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				receivedBody, _ = io.ReadAll(r.Body)
				receivedHeader = r.Header.Clone()
				w.Header().Set("Content-Type", "application/json")
				// The sms/send response example of
				// https://smsaero.ru/integration/documentation/api/
				_, _ = w.Write([]byte(`{"success": true,"data": [{"id": 1,"from": "SMS Aero","number": "79990000000","text": "your text","status": 0,"extendStatus": "queue","channel": "FREE SIGN","cost": 1.95,"dateCreate": 1510656981,"dateSend": 1510656981}],"message": null}`))
			})
			defer server.Close()

			err := newTestClient(server).Send(ctx, testSendOptions())
			So(err, ShouldBeNil)

			var request SendRequest
			So(json.Unmarshal(receivedBody, &request), ShouldBeNil)
			So(request, ShouldResemble, SendRequest{
				Number: "79990000000",
				Text:   "123456 is your code",
				Sign:   "SMS Aero",
			})

			So(receivedHeader.Get("Content-Type"), ShouldEqual, "application/json")
			username, password, ok := (&http.Request{Header: receivedHeader}).BasicAuth()
			So(ok, ShouldBeTrue)
			So(username, ShouldEqual, "user@example.com")
			So(password, ShouldEqual, "api-key")
		})

		Convey("success with a single data object", func() {
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"success":true,"data":{"id":5,"number":"79990000000"},"message":null}`))
			})
			defer server.Close()

			So(newTestClient(server).Send(ctx, testSendOptions()), ShouldBeNil)
		})

		Convey("invalid phone number in the request", func() {
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
				w.WriteHeader(statusCode)
				_, _ = w.Write([]byte(responseBody))
			})
			defer server.Close()

			err := newTestClient(server).Send(ctx, testSendOptions())
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderSmsAero)
			So(sendError.ProviderErrorCode, ShouldEqual, expectedCode)
			So(sendError.APIErrorKind, ShouldEqual, expected)
			So(sendError.DumpedResponse, ShouldNotBeEmpty)
		}

		Convey("rate limited by HTTP status", func() {
			assertErrorKind(
				http.StatusTooManyRequests,
				`{"success":false,"data":null,"message":"Too many requests"}`,
				&smsapi.ErrKindRateLimited,
				"Too many requests",
			)
		})

		Convey("rate limited by message", func() {
			assertErrorKind(
				http.StatusOK,
				`{"success":false,"data":null,"message":"Daily limit exceeded"}`,
				&smsapi.ErrKindRateLimited,
				"Daily limit exceeded",
			)
		})

		Convey("authentication failed by HTTP status", func() {
			assertErrorKind(
				http.StatusUnauthorized,
				`{"success":false,"data":null,"message":"Unauthorized"}`,
				&smsapi.ErrKindAuthenticationFailed,
				"Unauthorized",
			)
		})

		Convey("authentication failed by message", func() {
			assertErrorKind(
				http.StatusOK,
				`{"success":false,"data":null,"message":"Incorrect email or api_key"}`,
				&smsapi.ErrKindAuthenticationFailed,
				"Incorrect email or api_key",
			)
		})

		Convey("invalid phone number in the response", func() {
			assertErrorKind(
				http.StatusOK,
				`{"success":false,"data":null,"message":"Incorrect number"}`,
				&smsapi.ErrKindInvalidPhoneNumber,
				"Incorrect number",
			)
		})

		Convey("delivery rejected", func() {
			// "Not enough money" with HTTP 402 is the documented error of
			// https://smsaero.ru/integration/documentation/api/#param-errors
			assertErrorKind(
				http.StatusPaymentRequired,
				`{"success": false,"data": null,"message": "Not enough money"}`,
				&smsapi.ErrKindDeliveryRejected,
				"Not enough money",
			)
		})

		Convey("unapproved sign", func() {
			assertErrorKind(
				http.StatusOK,
				`{"success":false,"data":null,"message":"Sign is not approved"}`,
				&smsapi.ErrKindDeliveryRejected,
				"Sign is not approved",
			)
		})

		Convey("unknown message", func() {
			assertErrorKind(
				http.StatusOK,
				`{"success":false,"data":null,"message":"Something went wrong"}`,
				nil,
				"Something went wrong",
			)
		})

		Convey("validation errors object as the message", func() {
			assertErrorKind(
				http.StatusBadRequest,
				`{"success":false,"data":null,"message":{"number":["Invalid"]}}`,
				&smsapi.ErrKindInvalidPhoneNumber,
				"number: Invalid",
			)
		})

		Convey("exceeded limit of numbers is not an invalid phone number", func() {
			assertErrorKind(
				http.StatusOK,
				`{"success":false,"data":null,"message":"Exceeded limit of numbers"}`,
				&smsapi.ErrKindRateLimited,
				"Exceeded limit of numbers",
			)
		})

		Convey("assigned is not an unapproved sign", func() {
			assertErrorKind(
				http.StatusOK,
				`{"success":false,"data":null,"message":"No tariff assigned"}`,
				nil,
				"No tariff assigned",
			)
		})

		Convey("non 2xx status with success true in the body", func() {
			assertErrorKind(
				http.StatusInternalServerError,
				`{"success":true,"data":[],"message":null}`,
				nil,
				"",
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
			So(sendError.ProviderType, ShouldEqual, config.SMSProviderSmsAero)
			So(sendError.APIErrorKind, ShouldBeNil)
			So(sendError.DumpedResponse, ShouldNotBeEmpty)
		})

		Convey("non JSON response with an unauthorized status", func() {
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("<html>unauthorized</html>"))
			})
			defer server.Close()

			err := newTestClient(server).Send(ctx, testSendOptions())
			So(err, ShouldNotBeNil)

			var sendError *smsapi.SendError
			So(errors.As(err, &sendError), ShouldBeTrue)
			So(sendError.APIErrorKind, ShouldEqual, &smsapi.ErrKindAuthenticationFailed)
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

func TestParseSendResponse(t *testing.T) {
	Convey("ParseSendResponse", t, func() {
		Convey("data is an array", func() {
			response, err := ParseSendResponse([]byte(`{"success":true,"data":[{"id":5,"number":"79990000000","cost":2.2}],"message":null}`))
			So(err, ShouldBeNil)
			So(response.Success, ShouldBeTrue)
			So(response.NormalizedMessage(), ShouldEqual, "")
		})

		Convey("data is an object", func() {
			response, err := ParseSendResponse([]byte(`{"success":true,"data":{"id":5,"number":"79990000000"},"message":null}`))
			So(err, ShouldBeNil)
			So(response.Success, ShouldBeTrue)
		})

		Convey("message is a string", func() {
			response, err := ParseSendResponse([]byte(`{"success":false,"data":null,"message":"Incorrect number"}`))
			So(err, ShouldBeNil)
			So(response.Success, ShouldBeFalse)
			So(response.NormalizedMessage(), ShouldEqual, "Incorrect number")
		})

		Convey("message is a validation error object", func() {
			response, err := ParseSendResponse([]byte(`{"success":false,"data":null,"message":{"number":["Invalid"]}}`))
			So(err, ShouldBeNil)
			So(response.NormalizedMessage(), ShouldEqual, "number: Invalid")
		})

		Convey("message is a validation error object of many fields", func() {
			response, err := ParseSendResponse([]byte(`{"success":false,"message":{"sign":["Not approved"],"number":["Invalid","Too long"]}}`))
			So(err, ShouldBeNil)
			So(response.NormalizedMessage(), ShouldEqual, "number: Invalid; Too long; sign: Not approved")
		})

		Convey("message is absent", func() {
			response, err := ParseSendResponse([]byte(`{"success":false}`))
			So(err, ShouldBeNil)
			So(response.NormalizedMessage(), ShouldEqual, "")
		})

		Convey("message is an array of strings", func() {
			response, err := ParseSendResponse([]byte(`{"success":false,"message":["Incorrect number","Sign is not approved"]}`))
			So(err, ShouldBeNil)
			So(response.NormalizedMessage(), ShouldEqual, "Incorrect number; Sign is not approved")
		})

		Convey("not JSON", func() {
			_, err := ParseSendResponse([]byte(`<html>bad gateway</html>`))
			So(err, ShouldNotBeNil)
		})
	})
}
