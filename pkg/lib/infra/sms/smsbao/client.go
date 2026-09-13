package smsbao

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

const (
	// DefaultEndpoint is the base URL of the SMSBao API.
	DefaultEndpoint = "https://api.smsbao.com"
	// PathDomestic is the path of the mainland China SMS API.
	PathDomestic = "/sms"
	// PathInternational is the path of the international SMS API.
	PathInternational = "/wsms"

	// mainlandChinaCountryCallingCode is the country calling code of mainland China.
	mainlandChinaCountryCallingCode = "86"
)

type SmsbaoClient struct {
	Client *http.Client
	// Endpoint is the base URL of the API, without the path.
	Endpoint          string
	SmsbaoCredentials *config.SmsbaoCredentials
}

func NewSmsbaoClient(c *config.SmsbaoCredentials) *SmsbaoClient {
	if c == nil {
		return nil
	}

	return &SmsbaoClient{
		Client:            newHTTPClient(5 * time.Second),
		Endpoint:          DefaultEndpoint,
		SmsbaoCredentials: c,
	}
}

// newHTTPClient returns a client that keeps the credentials, which the API
// accepts in the query only, out of the telemetry recorded by otelhttp.
// otelhttp records url.full, which strips the userinfo but keeps the query,
// so the query is taken out of the URL before the request reaches otelhttp,
// and is put back by the inner RoundTripper afterwards.
func newHTTPClient(timeout time.Duration) *http.Client {
	client := utilhttputil.NewExternalClientWithOptions(timeout, utilhttputil.ExternalClientOptions{
		Transport: restoreQueryRoundTripper{},
	})
	client.Transport = stripQueryRoundTripper{Base: client.Transport}
	return client
}

// resolvePhoneNumber converts an E.164 phone number into the format expected by SMSBao.
// A mainland China number is sent without the country calling code to the domestic API,
// while any other number is sent in E.164 format, with the plus sign, to the
// international API.
// See https://www.smsbao.com/openapi/213.html and https://www.smsbao.com/openapi/299.html
func (s *SmsbaoClient) resolvePhoneNumber(to string) (phoneNumber string, isMainlandChina bool, err error) {
	parsed, err := phone.ParsePhoneNumberWithUserInput(to)
	if err != nil {
		return "", false, err
	}

	if parsed.CountryCallingCodeWithoutPlusSign == mainlandChinaCountryCallingCode {
		return parsed.NationalNumberWithoutFormatting, true, nil
	}
	return parsed.E164, false, nil
}

func (s *SmsbaoClient) send0(ctx context.Context, opts smsapi.SendOptions) ([]byte, []byte, error) {
	// Written against https://www.smsbao.com/openapi/213.html
	phoneNumber, isMainlandChina, err := s.resolvePhoneNumber(opts.To)
	if err != nil {
		return nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderSmsbao,
			APIErrorKind: &smsapi.ErrKindInvalidPhoneNumber,
		})
	}

	values := url.Values{}
	values.Set("u", s.SmsbaoCredentials.Username)
	// p is either the MD5 of the login password, or the API key.
	// The two are indistinguishable, so the configured value is sent as-is,
	// and the console tells the user to configure the MD5 of the password.
	values.Set("p", s.SmsbaoCredentials.PasswordOrAPIKey)
	values.Set("m", phoneNumber)
	values.Set("c", opts.Body)
	// g selects a dedicated channel product. It is documented by the domestic
	// API only, so it is not sent to the international API.
	if isMainlandChina && s.SmsbaoCredentials.GoodsID != "" {
		values.Set("g", s.SmsbaoCredentials.GoodsID)
	}

	path := PathInternational
	if isMainlandChina {
		path = PathDomestic
	}

	// The API accepts GET only, with every parameter in the query.
	// See https://www.smsbao.com/openapi/213.html and https://www.smsbao.com/openapi/299.html
	req, err := http.NewRequestWithContext(ctx, "GET", s.Endpoint+path+"?"+values.Encode(), nil)
	if err != nil {
		return nil, nil, redactError(err)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, nil, s.makeTransportError(err)
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType:   config.SMSProviderSmsbao,
			DumpedResponse: dumpedResponse,
		})
	}

	return bodyBytes, dumpedResponse, nil
}

func (s *SmsbaoClient) makeTransportError(err error) error {
	// The credentials are in the query, so the URL must never be reported as-is.
	redacted := redactError(err)

	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return errors.Join(redacted, &smsapi.SendError{
			ProviderType: config.SMSProviderSmsbao,
			APIErrorKind: &smsapi.ErrKindTimeout,
		})
	}
	return redacted
}

// redactError rewrites the URL reported by a *url.Error so that the
// credentials in the query do not reach logs or the console.
func redactError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return &url.Error{
			Op:  urlErr.Op,
			URL: redactURL(urlErr.URL),
			Err: urlErr.Err,
		}
	}
	return err
}

// redactURL keeps the scheme, the host and the path only.
func redactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	u.User = nil
	u.RawQuery = ""
	u.ForceQuery = false
	u.Fragment = ""
	u.RawFragment = ""
	return u.String()
}

func (s *SmsbaoClient) Send(ctx context.Context, opts smsapi.SendOptions) error {
	// SMSBao sends the body we give, so every SMS template is supported,
	// including the one carrying a link.
	bodyBytes, dumpedResponse, err := s.send0(ctx, opts)
	if err != nil {
		return err
	}

	responseCode, err := ParseSendResponse(bodyBytes)
	if err != nil {
		// The response is not something we can understand,
		// return an error with the dumped response.
		return errors.Join(err, &smsapi.SendError{
			ProviderType:   config.SMSProviderSmsbao,
			DumpedResponse: dumpedResponse,
		})
	}

	if responseCode != ResponseCodeOK {
		return s.makeError(responseCode, dumpedResponse)
	}

	return nil
}

func (s *SmsbaoClient) makeError(responseCode ResponseCode, dumpedResponse []byte) error {
	err := &smsapi.SendError{
		DumpedResponse:    dumpedResponse,
		ProviderType:      config.SMSProviderSmsbao,
		ProviderErrorCode: string(responseCode),
	}

	// See https://www.smsbao.com/openapi/213.html
	// The documented codes do not include a rate limit code,
	// so no code is mapped to ErrKindRateLimited.
	switch responseCode {
	case ResponseCodeWrongPassword:
		fallthrough
	case ResponseCodeAccountNotFound:
		fallthrough
	case ResponseCodeIPRestricted:
		// The IP address of the caller is not in the allowlist configured at SMSBao.
		// It is a credential-like problem rather than a rejected message.
		err.APIErrorKind = &smsapi.ErrKindAuthenticationFailed
	case ResponseCodeInsufficientBalance:
		fallthrough
	case ResponseCodeSensitiveContent:
		err.APIErrorKind = &smsapi.ErrKindDeliveryRejected
	case ResponseCodeInvalidPhoneNumber:
		err.APIErrorKind = &smsapi.ErrKindInvalidPhoneNumber
	}

	return err
}

var _ smsapi.Client = &SmsbaoClient{}

type queryContextKeyType struct{}

var queryContextKey = queryContextKeyType{}

// stripQueryRoundTripper moves the query out of the URL and into the request
// context, so that the layers below it, otelhttp in particular, do not see the
// credentials. restoreQueryRoundTripper puts the query back.
// The pair is symmetric, so a client without either of them keeps working.
type stripQueryRoundTripper struct {
	Base http.RoundTripper
}

func (t stripQueryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL != nil && req.URL.RawQuery != "" {
		rawQuery := req.URL.RawQuery
		req = req.Clone(context.WithValue(req.Context(), queryContextKey, rawQuery))
		req.URL.RawQuery = ""
	}
	return baseRoundTripper(t.Base).RoundTrip(req)
}

// restoreQueryRoundTripper puts back the query taken out by
// stripQueryRoundTripper, so that the request actually sent carries it.
type restoreQueryRoundTripper struct {
	Base http.RoundTripper
}

func (t restoreQueryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rawQuery, ok := req.Context().Value(queryContextKey).(string)
	if ok && rawQuery != "" && req.URL != nil && req.URL.RawQuery == "" {
		req = req.Clone(req.Context())
		req.URL.RawQuery = rawQuery
	}
	return baseRoundTripper(t.Base).RoundTrip(req)
}

func baseRoundTripper(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		return http.DefaultTransport
	}
	return rt
}
