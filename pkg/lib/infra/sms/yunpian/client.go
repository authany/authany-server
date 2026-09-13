package yunpian

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

const (
	// DefaultEndpoint is the endpoint of the Yunpian single send API.
	// The same endpoint serves both mainland China and international numbers;
	// us.yunpian.com is an overseas mirror of the same API.
	DefaultEndpoint = "https://sms.yunpian.com/v2/sms/single_send.json"

	// mainlandChinaCountryCallingCode is the country calling code of mainland China.
	mainlandChinaCountryCallingCode = "86"
)

type YunpianClient struct {
	Client             *http.Client
	Endpoint           string
	YunpianCredentials *config.YunpianCredentials
}

func NewYunpianClient(c *config.YunpianCredentials) *YunpianClient {
	if c == nil {
		return nil
	}

	return &YunpianClient{
		Client:             utilhttputil.NewExternalClient(5 * time.Second),
		Endpoint:           DefaultEndpoint,
		YunpianCredentials: c,
	}
}

// resolveMobile converts an E.164 phone number into the format expected by Yunpian.
// A mainland China number is sent as the national number alone,
// while any other number is sent in E.164, that is, with the plus sign.
func resolveMobile(to string) (string, error) {
	parsed, err := phone.ParsePhoneNumberWithUserInput(to)
	if err != nil {
		return "", err
	}

	if parsed.CountryCallingCodeWithoutPlusSign == mainlandChinaCountryCallingCode {
		return parsed.NationalNumberWithoutFormatting, nil
	}
	return parsed.E164, nil
}

func (y *YunpianClient) send0(ctx context.Context, opts smsapi.SendOptions) ([]byte, []byte, error) {
	// Written against
	// https://www.yunpian.com/official/document/sms/zh_CN/domestic_single_send
	// https://www.yunpian.com/official/document/sms/zh_CN/intl_single_send
	mobile, err := resolveMobile(opts.To)
	if err != nil {
		return nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderYunpian,
			APIErrorKind: &smsapi.ErrKindInvalidPhoneNumber,
		})
	}

	values := url.Values{}
	values.Set("apikey", y.YunpianCredentials.APIKey)
	values.Set("mobile", mobile)
	values.Set("text", opts.Body)

	req, err := http.NewRequestWithContext(ctx, "POST", y.Endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	req.Header.Set("Accept", "application/json;charset=utf-8")

	resp, err := y.Client.Do(req)
	if err != nil {
		return nil, nil, y.makeTransportError(err)
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Join(err, &smsapi.SendError{
			ProviderType:   config.SMSProviderYunpian,
			DumpedResponse: dumpedResponse,
		})
	}

	return bodyBytes, dumpedResponse, nil
}

func (y *YunpianClient) makeTransportError(err error) error {
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return errors.Join(err, &smsapi.SendError{
			ProviderType: config.SMSProviderYunpian,
			APIErrorKind: &smsapi.ErrKindTimeout,
		})
	}
	return err
}

func (y *YunpianClient) Send(ctx context.Context, opts smsapi.SendOptions) error {
	bodyBytes, dumpedResponse, err := y.send0(ctx, opts)
	if err != nil {
		return err
	}

	// An error is reported in the response body, with any HTTP status code.
	sendResponse, err := ParseSendResponse(bodyBytes)
	if err != nil {
		// The response is not something we can understand,
		// return an error with the dumped response.
		return errors.Join(err, &smsapi.SendError{
			ProviderType:   config.SMSProviderYunpian,
			DumpedResponse: dumpedResponse,
		})
	}

	if sendResponse.Code != ResponseCodeOK {
		return y.makeError(sendResponse.Code, dumpedResponse)
	}

	return nil
}

// errorCodeKinds maps the documented Yunpian return codes to the API error kinds.
// See https://www.yunpian.com/official/document/sms/zh_CN/returnvalue_common
var errorCodeKinds = map[int]*apierrors.Kind{
	// 账户余额不足
	3: &smsapi.ErrKindDeliveryRejected,
	// 未自动匹配到合适的模板
	5: &smsapi.ErrKindDeliveryRejected,
	// 模板不可用
	7: &smsapi.ErrKindDeliveryRejected,
	// 手机号防骚扰名单过滤
	10: &smsapi.ErrKindDeliveryRejected,
	// 用户号码防骚扰
	11: &smsapi.ErrKindDeliveryRejected,
	// 签名不匹配
	15: &smsapi.ErrKindDeliveryRejected,
	// 签名格式不正确
	16: &smsapi.ErrKindDeliveryRejected,
	// 签名校验失败
	18: &smsapi.ErrKindDeliveryRejected,
	// 暂不支持的国家地区
	20: &smsapi.ErrKindDeliveryRejected,
	// 号码归属地不在模板可发送的地区内
	23: &smsapi.ErrKindDeliveryRejected,
	// 手机号接收超过频率限制
	53: &smsapi.ErrKindRateLimited,
	// 服务已下线
	55: &smsapi.ErrKindDeliveryRejected,
	// 手机号码格式不正确
	56: &smsapi.ErrKindInvalidPhoneNumber,
	// 短信内容变量参数格式错误
	58: &smsapi.ErrKindDeliveryRejected,
	// 模版已失效，请重新提交模版
	60: &smsapi.ErrKindDeliveryRejected,
	// 非法的 apikey
	-1: &smsapi.ErrKindAuthenticationFailed,
	// API 没有权限
	-2: &smsapi.ErrKindAuthenticationFailed,
	// IP 没有权限
	-3: &smsapi.ErrKindAuthenticationFailed,
	// 访问次数超限
	-4: &smsapi.ErrKindRateLimited,
	// 访问频率超限
	-5: &smsapi.ErrKindRateLimited,
	// 不支持批量发送
	-6: &smsapi.ErrKindUnsupportedRequest,
	// 用户未开启该功能
	-11: &smsapi.ErrKindDeliveryRejected,
	// 无可用签名
	-57: &smsapi.ErrKindDeliveryRejected,
	// 无效资质信息
	-62: &smsapi.ErrKindDeliveryRejected,
}

func (y *YunpianClient) makeError(errorCode int, dumpedResponse []byte) error {
	err := &smsapi.SendError{
		DumpedResponse:    dumpedResponse,
		ProviderType:      config.SMSProviderYunpian,
		ProviderErrorCode: strconv.Itoa(errorCode),
	}

	if kind, ok := errorCodeKinds[errorCode]; ok {
		err.APIErrorKind = kind
	}

	return err
}

var _ smsapi.Client = &YunpianClient{}
