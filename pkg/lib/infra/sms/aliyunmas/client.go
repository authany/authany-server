package aliyunmas

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
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/aliyun"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/util/clock"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

const (
	// DefaultEndpoint is the endpoint of the Aliyun SMS Authentication Service.
	DefaultEndpoint = "https://dypnsapi.aliyuncs.com/"
	// APIVersion is the version of the SendSmsVerifyCode action.
	APIVersion = "2017-05-25"
	// DefaultRegionID is the region of the Aliyun SMS Authentication Service.
	DefaultRegionID = "cn-hangzhou"
	// ActionSendSmsVerifyCode is the action of sending a verification code by SMS.
	ActionSendSmsVerifyCode = "SendSmsVerifyCode"

	// mainlandChinaCountryCallingCode is the country calling code of mainland China.
	// The service supports mainland China phone numbers only.
	mainlandChinaCountryCallingCode = "86"

	// linkSMSTemplateName is the only SMS template carrying a link instead of a code.
	linkSMSTemplateName = "forgot_password_sms.txt"
)

type AliyunMASClient struct {
	Client               *http.Client
	Clock                clock.Clock
	Endpoint             string
	AliyunMASCredentials *config.AliyunMASCredentials
}

func NewAliyunMASClient(c *config.AliyunMASCredentials) *AliyunMASClient {
	if c == nil {
		return nil
	}

	return &AliyunMASClient{
		Client:               utilhttputil.NewExternalClient(5 * time.Second),
		Clock:                clock.NewSystemClock(),
		Endpoint:             DefaultEndpoint,
		AliyunMASCredentials: c,
	}
}

// resolvePhoneNumber converts an E.164 phone number into the format expected by
// the Aliyun SMS Authentication Service, that is, a mainland China number
// without the country calling code. Any other number is rejected.
func (a *AliyunMASClient) resolvePhoneNumber(to string) (string, error) {
	parsed, err := phone.ParsePhoneNumberWithUserInput(to)
	if err != nil {
		return "", err
	}

	if parsed.CountryCallingCodeWithoutPlusSign != mainlandChinaCountryCallingCode {
		return "", errors.New("aliyun_mas: only mainland China phone numbers are supported")
	}
	return parsed.NationalNumberWithoutFormatting, nil
}

func (a *AliyunMASClient) send0(ctx context.Context, opts smsapi.SendOptions) ([]byte, []byte, error) {
	// Written against
	// https://help.aliyun.com/zh/pnvs/developer-reference/api-dypnsapi-2017-05-25-sendsmsverifycode

	phoneNumber, err := a.resolvePhoneNumber(opts.To)
	if err != nil {
		return nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderAliyunMAS,
			APIErrorKind: &smsapi.ErrKindInvalidPhoneNumber,
		})
	}

	var code string
	if opts.TemplateVariables != nil {
		code = opts.TemplateVariables.Code
	}
	// We generate the code ourselves and send it as a literal value,
	// so the verification code is not verified at the Aliyun side.
	templateParam, err := json.Marshal(&aliyun.TemplateParam{Code: code})
	if err != nil {
		return nil, nil, err
	}

	values := aliyun.BuildCommonParams(aliyun.CommonParams{
		AccessKeyID: a.AliyunMASCredentials.AccessKeyID,
		Action:      ActionSendSmsVerifyCode,
		Version:     APIVersion,
		RegionID:    DefaultRegionID,
		Timestamp:   a.Clock.NowUTC(),
	})
	values.Set("PhoneNumber", phoneNumber)
	values.Set("CountryCode", mainlandChinaCountryCallingCode)
	values.Set("SignName", a.AliyunMASCredentials.SignName)
	values.Set("TemplateCode", a.AliyunMASCredentials.ResolveTemplateCode(opts.TemplateName))
	values.Set("TemplateParam", string(templateParam))
	values.Set("Signature", aliyun.SignRPCRequest("POST", values, a.AliyunMASCredentials.AccessKeySecret))

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
			ProviderType:   config.SMSProviderAliyunMAS,
			DumpedResponse: dumpedResponse,
		})
	}

	return bodyBytes, dumpedResponse, nil
}

func (a *AliyunMASClient) makeTransportError(err error) error {
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderAliyunMAS,
			APIErrorKind: &smsapi.ErrKindTimeout,
		})
	}
	return err
}

func (a *AliyunMASClient) Send(ctx context.Context, opts smsapi.SendOptions) error {
	// The Aliyun SMS Authentication Service does not allow sending an arbitrary body.
	// The body is defined by the template registered at Aliyun,
	// and the only template variable we send is the code.
	// Therefore the SMS template carrying a link is not supported.
	if opts.TemplateName == linkSMSTemplateName {
		return errors.Join(
			errors.New("aliyun_mas: sending a link by SMS is not supported; use the code-based flow to reset password"),
			&smsapi.SendError{
				ProviderType: config.SMSProviderAliyunMAS,
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
			ProviderType:   config.SMSProviderAliyunMAS,
			DumpedResponse: dumpedResponse,
		})
	}

	if sendResponse.Code != ResponseCodeOK {
		return a.makeError(sendResponse.Code, dumpedResponse)
	}

	return nil
}

func (a *AliyunMASClient) makeError(
	errorCode string,
	dumpedResponse []byte,
) error {
	err := &smsapi.SendError{
		DumpedResponse:    dumpedResponse,
		ProviderType:      config.SMSProviderAliyunMAS,
		ProviderErrorCode: errorCode,
	}

	// The action specific error codes are documented at
	// https://help.aliyun.com/zh/pnvs/developer-reference/api-dypnsapi-2017-05-25-sendsmsverifycode
	// The authentication error codes are the common ones shared by
	// every Aliyun RPC style API.
	switch errorCode {
	case "BUSINESS_LIMIT_CONTROL": // The daily sending limit of the receiver is exceeded
		fallthrough
	case "FREQUENCY_FAIL": // The sending interval is not respected
		err.APIErrorKind = &smsapi.ErrKindRateLimited
	case "MOBILE_NUMBER_ILLEGAL": // Invalid phone number
		err.APIErrorKind = &smsapi.ErrKindInvalidPhoneNumber
	case "InvalidAccessKeyId.NotFound":
		fallthrough
	case "SignatureDoesNotMatch":
		fallthrough
	case "Forbidden.AccessKeyDisabled":
		err.APIErrorKind = &smsapi.ErrKindAuthenticationFailed
	case "FUNCTION_NOT_OPENED": // The SMS authentication service is not enabled
		fallthrough
	// The signature and the template are provided by Aliyun.
	// Using the ones of the Aliyun SMS service instead results in these codes.
	// See https://help.aliyun.com/zh/pnvs/user-guide/sms-authentication-service
	case "isv.SMS_SIGNATURE_ILLEGAL":
		fallthrough
	case "isv.SMS_TEMPLATE_ILLEGAL":
		err.APIErrorKind = &smsapi.ErrKindDeliveryRejected
	}

	return err
}

var _ smsapi.Client = &AliyunMASClient{}
