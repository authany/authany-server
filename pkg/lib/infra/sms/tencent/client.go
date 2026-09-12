package tencent

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/util/clock"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
)

const (
	Host        = "sms.tencentcloudapi.com"
	Service     = "sms"
	Action      = "SendSms"
	Version     = "2021-01-11"
	Algorithm   = "TC3-HMAC-SHA256"
	ContentType = "application/json; charset=utf-8"
	// SignedHeaders is the set of headers participating in the signature.
	// It follows the example of https://cloud.tencent.com/document/api/213/30654
	SignedHeaders = "content-type;host;x-tc-action"
)

// TemplateNameForgotPasswordSMS is the only SMS template that carries a link
// instead of a code. Tencent Cloud SMS can only send pre-registered templates,
// so this template is not supported.
const TemplateNameForgotPasswordSMS = "forgot_password_sms.txt"

type TencentClient struct {
	Client             *http.Client
	Clock              clock.Clock
	TencentCredentials *config.TencentCredentials
	// Endpoint is the base URL of the API. It is only overridden in tests.
	Endpoint string
}

func NewTencentClient(c *config.TencentCredentials) *TencentClient {
	if c == nil {
		return nil
	}

	return &TencentClient{
		Client:             utilhttputil.NewExternalClient(5 * time.Second),
		Clock:              clock.NewSystemClock(),
		TencentCredentials: c,
		Endpoint:           "https://" + Host + "/",
	}
}

func sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

func hmacsha256(key []byte, s string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(s))
	return h.Sum(nil)
}

// The following signing functions are written against
// https://cloud.tencent.com/document/api/213/30654

func makeCanonicalRequest(host string, action string, payload string) string {
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-tc-action:%s\n", ContentType, host, strings.ToLower(action))
	return strings.Join([]string{
		"POST",
		"/",
		"",
		canonicalHeaders,
		SignedHeaders,
		sha256hex(payload),
	}, "\n")
}

func makeCredentialScope(date string, service string) string {
	return fmt.Sprintf("%s/%s/tc3_request", date, service)
}

func makeStringToSign(timestamp int64, credentialScope string, canonicalRequest string) string {
	return strings.Join([]string{
		Algorithm,
		strconv.FormatInt(timestamp, 10),
		credentialScope,
		sha256hex(canonicalRequest),
	}, "\n")
}

func makeSignature(secretKey string, date string, service string, stringToSign string) string {
	secretDate := hmacsha256([]byte("TC3"+secretKey), date)
	secretService := hmacsha256(secretDate, service)
	secretSigning := hmacsha256(secretService, "tc3_request")
	return hex.EncodeToString(hmacsha256(secretSigning, stringToSign))
}

// makeAuthorization computes the TC3-HMAC-SHA256 Authorization header.
func makeAuthorization(secretID string, secretKey string, host string, service string, action string, timestamp int64, payload string) string {
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	credentialScope := makeCredentialScope(date, service)
	canonicalRequest := makeCanonicalRequest(host, action, payload)
	stringToSign := makeStringToSign(timestamp, credentialScope, canonicalRequest)
	signature := makeSignature(secretKey, date, service, stringToSign)

	return fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		Algorithm, secretID, credentialScope, SignedHeaders, signature)
}

func (t *TencentClient) send0(ctx context.Context, opts smsapi.SendOptions) ([]byte, []byte, error) {
	// Written against
	// https://cloud.tencent.com/document/api/382/55981
	templateID := t.TencentCredentials.ResolveTemplateCode(opts.TemplateName)
	code := ""
	if opts.TemplateVariables != nil {
		code = opts.TemplateVariables.Code
	}

	requestBody, err := json.Marshal(&SendRequest{
		// Tencent Cloud expects the phone number in E.164 format.
		PhoneNumberSet:   []string{opts.To},
		SmsSdkAppId:      t.TencentCredentials.SDKAppID,
		SignName:         t.TencentCredentials.SignName,
		TemplateId:       templateID,
		TemplateParamSet: []string{code},
	})
	if err != nil {
		return nil, nil, err
	}

	timestamp := t.Clock.NowUTC().Unix()
	req, err := http.NewRequestWithContext(ctx, "POST", t.Endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, nil, err
	}
	req.Host = Host
	req.Header.Set("Content-Type", ContentType)
	req.Header.Set("X-TC-Action", Action)
	req.Header.Set("X-TC-Version", Version)
	req.Header.Set("X-TC-Region", t.TencentCredentials.Region)
	req.Header.Set("X-TC-Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("Authorization", makeAuthorization(
		t.TencentCredentials.SecretID,
		t.TencentCredentials.SecretKey,
		Host,
		Service,
		Action,
		timestamp,
		string(requestBody),
	))

	resp, err := t.Client.Do(req)
	if err != nil {
		return nil, nil, t.makeTransportError(err)
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType:   config.SMSProviderTencent,
			DumpedResponse: dumpedResponse,
		})
	}

	return bodyBytes, dumpedResponse, nil
}

func (t *TencentClient) makeTransportError(err error) error {
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderTencent,
			APIErrorKind: &smsapi.ErrKindTimeout,
		})
	}
	return err
}

func (t *TencentClient) Send(ctx context.Context, opts smsapi.SendOptions) error {
	if opts.TemplateName == TemplateNameForgotPasswordSMS {
		return errors.Join(
			errors.New("tencent: sending a link by SMS is not supported; use the code based password reset instead"),
			&smsapi.SendError{
				ProviderType: config.SMSProviderTencent,
				APIErrorKind: &smsapi.ErrKindUnsupportedRequest,
			},
		)
	}

	bodyBytes, dumpedResponse, err := t.send0(ctx, opts)
	if err != nil {
		return err
	}

	sendResponse, err := ParseSendResponse(bodyBytes)
	if err != nil {
		return errors.Join(err, &smsapi.SendError{
			ProviderType:   config.SMSProviderTencent,
			DumpedResponse: dumpedResponse,
		})
	}

	// A request level error.
	if respErr := sendResponse.Response.Error; respErr != nil {
		return t.makeError(respErr.Code, dumpedResponse)
	}

	// A send level status.
	if len(sendResponse.Response.SendStatusSet) == 0 {
		return &smsapi.SendError{
			ProviderType:   config.SMSProviderTencent,
			DumpedResponse: dumpedResponse,
		}
	}

	sendStatus := sendResponse.Response.SendStatusSet[0]
	if sendStatus.Code != SendStatusCodeOK {
		return t.makeError(sendStatus.Code, dumpedResponse)
	}

	return nil
}

func (t *TencentClient) makeError(errorCode string, dumpedResponse []byte) error {
	err := &smsapi.SendError{
		DumpedResponse:    dumpedResponse,
		ProviderType:      config.SMSProviderTencent,
		ProviderErrorCode: errorCode,
	}

	// See https://cloud.tencent.com/document/api/382/55981
	switch {
	case strings.HasPrefix(errorCode, "LimitExceeded"), errorCode == "RequestLimitExceeded":
		err.APIErrorKind = &smsapi.ErrKindRateLimited
	case errorCode == "InvalidParameterValue.IncorrectPhoneNumber":
		err.APIErrorKind = &smsapi.ErrKindInvalidPhoneNumber
	case strings.HasPrefix(errorCode, "AuthFailure"):
		err.APIErrorKind = &smsapi.ErrKindAuthenticationFailed
	case errorCode == "FailedOperation.SignatureIncorrectOrUnapproved",
		errorCode == "FailedOperation.TemplateIncorrectOrUnapproved":
		err.APIErrorKind = &smsapi.ErrKindDeliveryRejected
	}

	return err
}

var _ smsapi.Client = &TencentClient{}
