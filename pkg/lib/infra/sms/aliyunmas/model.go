package aliyunmas

import (
	"encoding/json"
)

// SendResponse is the response of the SendSmsVerifyCode action.
// See https://help.aliyun.com/zh/pnvs/developer-reference/api-dypnsapi-2017-05-25-sendsmsverifycode
// An error response has the same shape, with Code being an error code.
type SendResponse struct {
	Code      string             `json:"Code,omitempty"`
	Message   string             `json:"Message,omitempty"`
	RequestID string             `json:"RequestId,omitempty"`
	Success   bool               `json:"Success,omitempty"`
	Model     *SendResponseModel `json:"Model,omitempty"`
}

// SendResponseModel is the Model field of SendResponse.
type SendResponseModel struct {
	RequestID string `json:"RequestId,omitempty"`
	BizID     string `json:"BizId,omitempty"`
	OutID     string `json:"OutId,omitempty"`
	// VerifyCode is returned only when ReturnVerifyCode is true, which we never set.
	VerifyCode string `json:"VerifyCode,omitempty"`
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
