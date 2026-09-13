package gatewayapi

import (
	"encoding/json"
)

// Recipient is an entry of the recipients array.
// https://gatewayapi.com/docs/apis/rest/
type Recipient struct {
	// MSISDN is the full mobile phone number without the plus sign.
	MSISDN string `json:"msisdn"`
}

// SendRequest is the request body of POST /rest/mtsms.
type SendRequest struct {
	Sender     string      `json:"sender,omitempty"`
	Message    string      `json:"message"`
	Recipients []Recipient `json:"recipients"`
}

// SendResponse is the response body of a successful POST /rest/mtsms.
type SendResponse struct {
	IDs []json.Number `json:"ids,omitempty"`
}

// ErrorResponse is the response body of a failed POST /rest/mtsms.
// code and variables are omitted by the server when they are empty.
type ErrorResponse struct {
	Code         string   `json:"code,omitempty"`
	IncidentUUID string   `json:"incident_uuid,omitempty"`
	Message      string   `json:"message,omitempty"`
	Variables    []string `json:"variables,omitempty"`
}

func ParseSendResponse(jsonData []byte) (*SendResponse, error) {
	var response SendResponse
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func ParseErrorResponse(jsonData []byte) (*ErrorResponse, error) {
	var response ErrorResponse
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
