package gatewayapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

// Path is the path of the send SMS endpoint, relative to the configured endpoint.
const Path = "/rest/mtsms"

// Accept follows the examples of https://gatewayapi.com/docs/message/overview/
const (
	Accept      = "application/json, text/javascript"
	ContentType = "application/json"
)

type GatewayAPIClient struct {
	Client                *http.Client
	GatewayAPICredentials *config.GatewayAPICredentials
}

func NewGatewayAPIClient(c *config.GatewayAPICredentials) *GatewayAPIClient {
	if c == nil {
		return nil
	}

	return &GatewayAPIClient{
		Client:                utilhttputil.NewExternalClient(5 * time.Second),
		GatewayAPICredentials: c,
	}
}

// resolveMSISDN converts an E.164 phone number into a MSISDN,
// that is, the country calling code and the national number, without the plus sign.
func resolveMSISDN(to string) (string, error) {
	parsed, err := phone.ParsePhoneNumberWithUserInput(to)
	if err != nil {
		return "", err
	}
	return parsed.CountryCallingCodeWithoutPlusSign + parsed.NationalNumberWithoutFormatting, nil
}

func (g *GatewayAPIClient) endpoint() string {
	return strings.TrimSuffix(g.GatewayAPICredentials.Endpoint, "/") + Path
}

func (g *GatewayAPIClient) send0(ctx context.Context, opts smsapi.SendOptions) (int, []byte, []byte, error) {
	// Written against https://gatewayapi.com/docs/message/overview/
	msisdn, err := resolveMSISDN(opts.To)
	if err != nil {
		return 0, nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderGatewayAPI,
			APIErrorKind: &smsapi.ErrKindInvalidPhoneNumber,
		})
	}

	requestBody, err := json.Marshal(&SendRequest{
		Sender:     g.GatewayAPICredentials.Sender,
		Message:    opts.Body,
		Recipients: []Recipient{{MSISDN: msisdn}},
	})
	if err != nil {
		return 0, nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.endpoint(), bytes.NewReader(requestBody))
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Content-Type", ContentType)
	req.Header.Set("Accept", Accept)
	req.Header.Set("Authorization", "Token "+g.GatewayAPICredentials.APIToken)

	resp, err := g.Client.Do(req)
	if err != nil {
		return 0, nil, nil, g.makeTransportError(err)
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return 0, nil, nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType:   config.SMSProviderGatewayAPI,
			DumpedResponse: dumpedResponse,
		})
	}

	return resp.StatusCode, bodyBytes, dumpedResponse, nil
}

func (g *GatewayAPIClient) makeTransportError(err error) error {
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderGatewayAPI,
			APIErrorKind: &smsapi.ErrKindTimeout,
		})
	}
	return err
}

func (g *GatewayAPIClient) Send(ctx context.Context, opts smsapi.SendOptions) error {
	statusCode, bodyBytes, dumpedResponse, err := g.send0(ctx, opts)
	if err != nil {
		return err
	}

	if statusCode < 200 || statusCode > 299 {
		errorCode := ""
		// The error body is not guaranteed to be JSON, for example when an
		// intermediary returns the error instead of GatewayAPI.
		if errorResponse, err := ParseErrorResponse(bodyBytes); err == nil {
			errorCode = errorResponse.Code
		}
		return g.makeError(statusCode, errorCode, dumpedResponse)
	}

	sendResponse, err := ParseSendResponse(bodyBytes)
	if err != nil {
		return errors.Join(err, &smsapi.SendError{
			ProviderType:   config.SMSProviderGatewayAPI,
			DumpedResponse: dumpedResponse,
		})
	}

	// A successful response always carries the IDs of the accepted messages.
	if len(sendResponse.IDs) == 0 {
		return &smsapi.SendError{
			ProviderType:   config.SMSProviderGatewayAPI,
			DumpedResponse: dumpedResponse,
		}
	}

	return nil
}

// errorCodeKinds maps the documented GatewayAPI error codes to the API error kinds.
// See https://gatewayapi.com/docs/error-codes/
var errorCodeKinds = map[string]*apierrors.Kind{
	// 401 Invalid username
	"0X0210": &smsapi.ErrKindAuthenticationFailed,
	// 401 Invalid password
	"0X0211": &smsapi.ErrKindAuthenticationFailed,
	// 401 Invalid IP-address
	"0X0212": &smsapi.ErrKindAuthenticationFailed,
	// 403 Unauthorized IP-address
	"0X0213": &smsapi.ErrKindAuthenticationFailed,
	// 403 Temporary blacklist for MSISDN
	"0X0214": &smsapi.ErrKindDeliveryRejected,
	// 403 MSISDN blacklisted
	"0X0215": &smsapi.ErrKindDeliveryRejected,
	// 403 Insufficient credit
	"0X0216": &smsapi.ErrKindDeliveryRejected,
	// 403 Unauthorized destination: country
	"0X0217": &smsapi.ErrKindDeliveryRejected,
	// 403 SMS not enabled for account
	"0X021D": &smsapi.ErrKindDeliveryRejected,
	// 401 Invalid Consumer Key
	"0X0223": &smsapi.ErrKindAuthenticationFailed,
	// 401 Invalid signature
	"0X0224": &smsapi.ErrKindAuthenticationFailed,
	// 401 Expired timestamp
	"0X0225": &smsapi.ErrKindAuthenticationFailed,
	// 401 Invalid / used nonce
	"0X0226": &smsapi.ErrKindAuthenticationFailed,
	// 401 Invalid token
	"0X0229": &smsapi.ErrKindAuthenticationFailed,
	// 403 Account frozen, contact support
	"0X022A": &smsapi.ErrKindDeliveryRejected,
	// 422 Messages filtered based on content
	"0X0312": &smsapi.ErrKindDeliveryRejected,
	// 422 A message recipient belongs to a blocked country
	"0X0313": &smsapi.ErrKindDeliveryRejected,
}

func (g *GatewayAPIClient) makeError(statusCode int, errorCode string, dumpedResponse []byte) error {
	err := &smsapi.SendError{
		DumpedResponse:    dumpedResponse,
		ProviderType:      config.SMSProviderGatewayAPI,
		ProviderErrorCode: errorCode,
	}

	if kind, ok := errorCodeKinds[strings.ToUpper(strings.TrimSpace(errorCode))]; ok {
		err.APIErrorKind = kind
		return err
	}

	// The error code is absent or undocumented, fall back to the HTTP status code.
	switch statusCode {
	case http.StatusTooManyRequests:
		err.APIErrorKind = &smsapi.ErrKindRateLimited
	case http.StatusUnauthorized:
		err.APIErrorKind = &smsapi.ErrKindAuthenticationFailed
	}

	return err
}

var _ smsapi.Client = &GatewayAPIClient{}
