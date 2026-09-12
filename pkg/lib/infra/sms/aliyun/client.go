package aliyun

import (
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
	"github.com/authgear/authgear-server/pkg/util/clock"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

const (
	// DefaultEndpoint is the endpoint of the Aliyun SMS service.
	DefaultEndpoint = "https://dysmsapi.aliyuncs.com/"
	// APIVersion is the version of the SendSms action.
	APIVersion = "2017-05-25"
	// DefaultRegionID is the region of the Aliyun SMS service.
	DefaultRegionID = "cn-hangzhou"
	// ActionSendSms is the action of sending an SMS.
	ActionSendSms = "SendSms"

	// mainlandChinaCountryCallingCode is the country calling code of mainland China.
	mainlandChinaCountryCallingCode = "86"

	// linkSMSTemplateName is the only SMS template carrying a link instead of a code.
	linkSMSTemplateName = "forgot_password_sms.txt"
)

type AliyunClient struct {
	Client            *http.Client
	Clock             clock.Clock
	Endpoint          string
	AliyunCredentials *config.AliyunCredentials
}

func NewAliyunClient(c *config.AliyunCredentials) *AliyunClient {
	if c == nil {
		return nil
	}

	return &AliyunClient{
		Client:            utilhttputil.NewExternalClient(5 * time.Second),
		Clock:             clock.NewSystemClock(),
		Endpoint:          DefaultEndpoint,
		AliyunCredentials: c,
	}
}

// resolvePhoneNumber converts an E.164 phone number into the format expected by Aliyun.
// A mainland China number is sent without the country calling code,
// while any other number is sent with the country calling code, without the plus sign.
func (a *AliyunClient) resolvePhoneNumber(to string) (phoneNumber string, isMainlandChina bool, err error) {
	parsed, err := phone.ParsePhoneNumberWithUserInput(to)
	if err != nil {
		return "", false, err
	}

	if parsed.CountryCallingCodeWithoutPlusSign == mainlandChinaCountryCallingCode {
		return parsed.NationalNumberWithoutFormatting, true, nil
	}
	return parsed.CountryCallingCodeWithoutPlusSign + parsed.NationalNumberWithoutFormatting, false, nil
}

func (a *AliyunClient) resolveTemplateCode(templateName string, isMainlandChina bool) string {
	if !isMainlandChina && a.AliyunCredentials.OverseasTemplateCode != "" {
		return a.AliyunCredentials.OverseasTemplateCode
	}
	return a.AliyunCredentials.ResolveTemplateCode(templateName)
}

func (a *AliyunClient) send0(ctx context.Context, opts smsapi.SendOptions) ([]byte, []byte, error) {
	// Written against
	// https://help.aliyun.com/zh/sms/developer-reference/api-dysmsapi-2017-05-25-sendsms

	phoneNumber, isMainlandChina, err := a.resolvePhoneNumber(opts.To)
	if err != nil {
		return nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderAliyun,
			APIErrorKind: &smsapi.ErrKindInvalidPhoneNumber,
		})
	}

	var code string
	if opts.TemplateVariables != nil {
		code = opts.TemplateVariables.Code
	}
	templateParam, err := json.Marshal(&TemplateParam{Code: code})
	if err != nil {
		return nil, nil, err
	}

	values := BuildCommonParams(CommonParams{
		AccessKeyID: a.AliyunCredentials.AccessKeyID,
		Action:      ActionSendSms,
		Version:     APIVersion,
		RegionID:    DefaultRegionID,
		Timestamp:   a.Clock.NowUTC(),
	})
	values.Set("PhoneNumbers", phoneNumber)
	values.Set("SignName", a.AliyunCredentials.SignName)
	values.Set("TemplateCode", a.resolveTemplateCode(opts.TemplateName, isMainlandChina))
	values.Set("TemplateParam", string(templateParam))
	values.Set("Signature", SignRPCRequest("POST", values, a.AliyunCredentials.AccessKeySecret))

	requestBody := values.Encode()
	req, err := http.NewRequestWithContext(ctx, "POST", a.Endpoint, strings.NewReader(requestBody))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, nil, a.makeTransportError(err)
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType:   config.SMSProviderAliyun,
			DumpedResponse: dumpedResponse,
		})
	}

	return bodyBytes, dumpedResponse, nil
}

func (a *AliyunClient) makeTransportError(err error) error {
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderAliyun,
			APIErrorKind: &smsapi.ErrKindTimeout,
		})
	}
	return err
}

func (a *AliyunClient) Send(ctx context.Context, opts smsapi.SendOptions) error {
	// Aliyun does not allow sending an arbitrary body.
	// The body is defined by the template registered at Aliyun,
	// and the only template variable we send is the code.
	// Therefore the SMS template carrying a link is not supported.
	if opts.TemplateName == linkSMSTemplateName {
		return errors.Join(
			errors.New("aliyun: sending a link by SMS is not supported; use the code-based flow to reset password"),
			&smsapi.SendError{
				ProviderType: config.SMSProviderAliyun,
				APIErrorKind: &smsapi.ErrKindUnsupportedRequest,
			},
		)
	}

	bodyBytes, dumpedResponse, err := a.send0(ctx, opts)
	if err != nil {
		return err
	}

	sendResponse, err := ParseSendResponse(bodyBytes)
	if err != nil {
		// The response is not something we can understand,
		// return an error with the dumped response.
		return errors.Join(err, &smsapi.SendError{
			ProviderType:   config.SMSProviderAliyun,
			DumpedResponse: dumpedResponse,
		})
	}

	if sendResponse.Code != ResponseCodeOK {
		return a.makeError(sendResponse.Code, dumpedResponse)
	}

	return nil
}

func (a *AliyunClient) makeError(
	errorCode string,
	dumpedResponse []byte,
) error {
	err := &smsapi.SendError{
		DumpedResponse:    dumpedResponse,
		ProviderType:      config.SMSProviderAliyun,
		ProviderErrorCode: errorCode,
	}

	// See https://help.aliyun.com/zh/sms/developer-reference/api-error-codes
	switch errorCode {
	case "isv.BUSINESS_LIMIT_CONTROL": // The sending limit of the receiver is exceeded
		fallthrough
	case "isv.DAY_LIMIT_CONTROL": // The daily sending limit is exceeded
		err.APIErrorKind = &smsapi.ErrKindRateLimited
	case "isv.MOBILE_NUMBER_ILLEGAL": // Invalid phone number
		err.APIErrorKind = &smsapi.ErrKindInvalidPhoneNumber
	case "InvalidAccessKeyId.NotFound":
		fallthrough
	case "SignatureDoesNotMatch":
		fallthrough
	case "Forbidden.AccessKeyDisabled":
		err.APIErrorKind = &smsapi.ErrKindAuthenticationFailed
	case "isv.SMS_SIGNATURE_ILLEGAL": // The signature is not approved
		fallthrough
	case "isv.SMS_TEMPLATE_ILLEGAL": // The template is not approved
		fallthrough
	case "isv.TEMPLATE_MISSING_PARAMETERS": // The template variables do not match
		fallthrough
	case "isv.OUT_OF_SERVICE": // The service is suspended
		fallthrough
	case "isv.AMOUNT_NOT_ENOUGH": // Out of balance
		err.APIErrorKind = &smsapi.ErrKindDeliveryRejected
	}

	return err
}

var _ smsapi.Client = &AliyunClient{}
