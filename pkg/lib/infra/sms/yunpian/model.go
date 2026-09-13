package yunpian

import (
	"encoding/json"
)

// SendResponse is the response of the single send API.
// See https://www.yunpian.com/official/document/sms/zh_CN/domestic_single_send
// An error response has the same shape, with Code being an error code,
// and Detail carrying a human readable explanation.
type SendResponse struct {
	Code   int     `json:"code"`
	Msg    string  `json:"msg,omitempty"`
	Detail string  `json:"detail,omitempty"`
	Count  int     `json:"count,omitempty"`
	Fee    float64 `json:"fee,omitempty"`
	Unit   string  `json:"unit,omitempty"`
	Mobile string  `json:"mobile,omitempty"`
	SID    int64   `json:"sid,omitempty"`
}

// ResponseCodeOK is the Code of a successful response.
const ResponseCodeOK = 0

func ParseSendResponse(jsonData []byte) (*SendResponse, error) {
	var response SendResponse
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
