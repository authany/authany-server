package smsaero

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// SendRequest is the request body of POST /v2/sms/send.
// https://smsaero.ru/integration/documentation/api/
type SendRequest struct {
	// Number is the phone number in digits only, without the plus sign.
	Number string `json:"number"`
	Text   string `json:"text"`
	// Sign is the sender name registered at SMS Aero.
	Sign string `json:"sign"`
}

// SendResponseData is the data object of a successful response.
type SendResponseData struct {
	ID           json.Number `json:"id,omitempty"`
	From         string      `json:"from,omitempty"`
	Number       string      `json:"number,omitempty"`
	Text         string      `json:"text,omitempty"`
	Status       int         `json:"status,omitempty"`
	ExtendStatus string      `json:"extendStatus,omitempty"`
	Channel      string      `json:"channel,omitempty"`
	Cost         json.Number `json:"cost,omitempty"`
	DateCreate   int64       `json:"dateCreate,omitempty"`
	DateSend     int64       `json:"dateSend,omitempty"`
}

// SendResponse is the response body of POST /v2/sms/send.
// SMS Aero does not return a machine readable error code,
// it only returns a human readable message.
type SendResponse struct {
	Success bool `json:"success"`
	// Data is an array of SendResponseData on success, but some endpoints
	// document it as a single object. Send only consumes Success, so the
	// shape is left unparsed.
	Data json.RawMessage `json:"data,omitempty"`
	// Message is a string, or an object of validation errors keyed by field,
	// for example {"number":["Invalid"]}. Use NormalizedMessage to read it.
	Message json.RawMessage `json:"message,omitempty"`
}

// NormalizedMessage renders Message as a human readable string.
// A string is returned as is, an object or an array of validation errors
// is flattened into "key: msg; key2: msg".
func (r *SendResponse) NormalizedMessage() string {
	return normalizeMessage(r.Message)
}

func normalizeMessage(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return ""
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString
	}

	var asObject map[string]json.RawMessage
	if err := json.Unmarshal(raw, &asObject); err == nil {
		keys := make([]string, 0, len(asObject))
		for key := range asObject {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, fmt.Sprintf("%s: %s", key, normalizeMessage(asObject[key])))
		}
		return strings.Join(parts, "; ")
	}

	var asArray []json.RawMessage
	if err := json.Unmarshal(raw, &asArray); err == nil {
		parts := make([]string, 0, len(asArray))
		for _, element := range asArray {
			parts = append(parts, normalizeMessage(element))
		}
		return strings.Join(parts, "; ")
	}

	// Any other JSON value, for example a number or a boolean.
	return trimmed
}

func ParseSendResponse(jsonData []byte) (*SendResponse, error) {
	var response SendResponse
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
