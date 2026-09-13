package smsaero

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

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

// DefaultEndpoint is the endpoint of the SMS Aero send API.
const DefaultEndpoint = "https://gate.smsaero.ru/v2/sms/send"

type SmsAeroClient struct {
	Client             *http.Client
	SmsAeroCredentials *config.SmsAeroCredentials
	// Endpoint is the URL of the send API. It is only overridden in tests.
	Endpoint string
}

func NewSmsAeroClient(c *config.SmsAeroCredentials) *SmsAeroClient {
	if c == nil {
		return nil
	}

	return &SmsAeroClient{
		Client:             utilhttputil.NewExternalClient(5 * time.Second),
		SmsAeroCredentials: c,
		Endpoint:           DefaultEndpoint,
	}
}

// resolvePhoneNumber converts an E.164 phone number into digits only,
// which is the format expected by SMS Aero.
func resolvePhoneNumber(to string) (string, error) {
	parsed, err := phone.ParsePhoneNumberWithUserInput(to)
	if err != nil {
		return "", err
	}
	return parsed.CountryCallingCodeWithoutPlusSign + parsed.NationalNumberWithoutFormatting, nil
}

func (c *SmsAeroClient) send0(ctx context.Context, opts smsapi.SendOptions) (int, []byte, []byte, error) {
	// Written against
	// https://smsaero.ru/integration/documentation/api/

	number, err := resolvePhoneNumber(opts.To)
	if err != nil {
		return 0, nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderSmsAero,
			APIErrorKind: &smsapi.ErrKindInvalidPhoneNumber,
		})
	}

	requestBody, err := json.Marshal(&SendRequest{
		Number: number,
		Text:   opts.Body,
		Sign:   c.SmsAeroCredentials.SenderName,
	})
	if err != nil {
		return 0, nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// SMS Aero authenticates with HTTP Basic, the username is the email
	// and the password is the API key.
	req.SetBasicAuth(c.SmsAeroCredentials.Email, c.SmsAeroCredentials.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return 0, nil, nil, c.makeTransportError(err)
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return 0, nil, nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType:   config.SMSProviderSmsAero,
			DumpedResponse: dumpedResponse,
		})
	}

	return resp.StatusCode, bodyBytes, dumpedResponse, nil
}

func (c *SmsAeroClient) makeTransportError(err error) error {
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderSmsAero,
			APIErrorKind: &smsapi.ErrKindTimeout,
		})
	}
	return err
}

func (c *SmsAeroClient) Send(ctx context.Context, opts smsapi.SendOptions) error {
	statusCode, bodyBytes, dumpedResponse, err := c.send0(ctx, opts)
	if err != nil {
		return err
	}

	if statusCode < 200 || statusCode > 299 {
		message := ""
		// The error body is not guaranteed to be JSON, for example when an
		// intermediary returns the error instead of SMS Aero.
		if sendResponse, err := ParseSendResponse(bodyBytes); err == nil {
			message = sendResponse.NormalizedMessage()
		}
		return c.makeError(statusCode, message, dumpedResponse)
	}

	sendResponse, err := ParseSendResponse(bodyBytes)
	if err != nil {
		// The response is not something we can understand,
		// the HTTP status is still mapped when it is meaningful.
		return errors.Join(err, c.makeError(statusCode, "", dumpedResponse))
	}

	if !sendResponse.Success {
		return c.makeError(statusCode, sendResponse.NormalizedMessage(), dumpedResponse)
	}

	return nil
}

func (c *SmsAeroClient) makeError(statusCode int, message string, dumpedResponse []byte) error {
	err := &smsapi.SendError{
		DumpedResponse: dumpedResponse,
		ProviderType:   config.SMSProviderSmsAero,
		// SMS Aero does not return an error code, the message takes its place.
		ProviderErrorCode: message,
	}

	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		err.APIErrorKind = &smsapi.ErrKindAuthenticationFailed
		return err
	case http.StatusTooManyRequests:
		err.APIErrorKind = &smsapi.ErrKindRateLimited
		return err
	}

	// The message is the only information SMS Aero gives about the failure.
	// The quota and the balance are checked first, because their messages
	// mention numbers, for example "Exceeded limit of numbers".
	m := strings.ToLower(message)
	switch {
	case strings.Contains(m, "authoriz"), strings.Contains(m, "api_key"):
		err.APIErrorKind = &smsapi.ErrKindAuthenticationFailed
	case strings.Contains(m, "limit"):
		err.APIErrorKind = &smsapi.ErrKindRateLimited
	case strings.Contains(m, "credit"), strings.Contains(m, "money"):
		err.APIErrorKind = &smsapi.ErrKindDeliveryRejected
	// "number:" is the prefix of a flattened validation error on the number field.
	case strings.Contains(m, "incorrect number"), strings.Contains(m, "invalid number"),
		strings.Contains(m, "number:"), strings.Contains(m, "phone"):
		err.APIErrorKind = &smsapi.ErrKindInvalidPhoneNumber
	// "sign " and "signature not" do not match unrelated words such as "assigned".
	case strings.Contains(m, "sign "), strings.Contains(m, "signature not"):
		err.APIErrorKind = &smsapi.ErrKindDeliveryRejected
	}

	return err
}

var _ smsapi.Client = &SmsAeroClient{}
