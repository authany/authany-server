package tencent

import (
	"encoding/json"
)

// SendRequest is the request body of the SendSms action.
// https://cloud.tencent.com/document/api/382/55981
type SendRequest struct {
	PhoneNumberSet   []string `json:"PhoneNumberSet"`
	SmsSdkAppId      string   `json:"SmsSdkAppId"`
	SignName         string   `json:"SignName"`
	TemplateId       string   `json:"TemplateId"`
	TemplateParamSet []string `json:"TemplateParamSet"`
}

// ResponseError is the request level error.
type ResponseError struct {
	Code    string `json:"Code,omitempty"`
	Message string `json:"Message,omitempty"`
}

// SendStatus is the per phone number send level status.
type SendStatus struct {
	SerialNo       string `json:"SerialNo,omitempty"`
	PhoneNumber    string `json:"PhoneNumber,omitempty"`
	Fee            int    `json:"Fee,omitempty"`
	SessionContext string `json:"SessionContext,omitempty"`
	Code           string `json:"Code,omitempty"`
	Message        string `json:"Message,omitempty"`
	IsoCode        string `json:"IsoCode,omitempty"`
}

// SendStatusCodeOK is the send level code of a successfully accepted message.
const SendStatusCodeOK = "Ok"

type SendResponseBody struct {
	SendStatusSet []SendStatus   `json:"SendStatusSet,omitempty"`
	RequestId     string         `json:"RequestId,omitempty"`
	Error         *ResponseError `json:"Error,omitempty"`
}

type SendResponse struct {
	Response SendResponseBody `json:"Response"`
}

func ParseSendResponse(jsonData []byte) (*SendResponse, error) {
	var response SendResponse
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
