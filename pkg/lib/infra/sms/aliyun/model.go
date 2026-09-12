package aliyun

import (
	"encoding/json"
)

// SendResponse is the response of the SendSms action.
// See https://help.aliyun.com/zh/sms/developer-reference/api-dysmsapi-2017-05-25-sendsms
// An error response has the same shape, with Code being an error code.
type SendResponse struct {
	Code      string `json:"Code,omitempty"`
	Message   string `json:"Message,omitempty"`
	RequestID string `json:"RequestId,omitempty"`
	BizID     string `json:"BizId,omitempty"`
}

// ResponseCodeOK is the Code of a successful response.
const ResponseCodeOK = "OK"

func ParseSendResponse(jsonData []byte) (*SendResponse, error) {
	var response SendResponse
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// TemplateParam is the value of the TemplateParam parameter.
// The registered template must contain exactly one variable, namely ${code}.
type TemplateParam struct {
	Code string `json:"code"`
}
