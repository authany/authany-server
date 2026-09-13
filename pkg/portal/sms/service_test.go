package sms

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

func TestSendTestSMSGatewayAPIEndpoint(t *testing.T) {
	Convey("SendTestSMS rejects a GatewayAPI endpoint not in the allowlist", t, func() {
		called := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		s := &Service{}

		test := func(endpoint string) error {
			return s.SendTestSMS(
				context.Background(),
				&model.App{},
				"+85298887766",
				nil,
				model.SMSProviderConfigurationInput{
					GatewayAPI: &model.SMSProviderConfigurationGatewayAPIInput{
						Endpoint: endpoint,
						APIToken: "token",
						Sender:   "Authgear",
					},
				},
			)
		}

		for _, endpoint := range []string{
			server.URL,
			"https://evil.example.com",
			"https://gatewayapi.com.evil.example.com",
			"https://gatewayapi.com/",
			"",
		} {
			err := test(endpoint)
			So(err, ShouldNotBeNil)
			So(apierrors.IsAPIError(err), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "endpoint must be one of")
		}

		So(called, ShouldBeFalse)
	})
}
