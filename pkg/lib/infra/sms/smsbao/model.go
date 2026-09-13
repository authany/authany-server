package smsbao

import (
	"fmt"
	"strings"
)

// ResponseCode is the plain text status code of a send response.
// SMSBao does not return JSON. The response body is the status code itself.
// See https://www.smsbao.com/openapi/213.html
type ResponseCode string

const (
	// ResponseCodeOK means the SMS was accepted.
	ResponseCodeOK ResponseCode = "0"
	// ResponseCodeWrongPassword means the password or the API key is wrong.
	ResponseCodeWrongPassword ResponseCode = "30"
	// ResponseCodeAccountNotFound means the username does not exist.
	ResponseCodeAccountNotFound ResponseCode = "40"
	// ResponseCodeInsufficientBalance means the account is out of balance.
	ResponseCodeInsufficientBalance ResponseCode = "41"
	// ResponseCodeIPRestricted means the caller IP address is not allowed.
	ResponseCodeIPRestricted ResponseCode = "43"
	// ResponseCodeSensitiveContent means the body contains a forbidden word.
	ResponseCodeSensitiveContent ResponseCode = "50"
	// ResponseCodeInvalidPhoneNumber means the phone number is malformed.
	ResponseCodeInvalidPhoneNumber ResponseCode = "51"
)

// maxResponseCodeLength guards against treating an arbitrary long body as a code.
const maxResponseCodeLength = 8

// ParseSendResponse parses the plain text status code out of the response body.
// The status code is on the first line, as documented by the balance query API
// which returns extra lines after the status code.
func ParseSendResponse(body []byte) (ResponseCode, error) {
	firstLine := string(body)
	if idx := strings.IndexAny(firstLine, "\r\n"); idx >= 0 {
		firstLine = firstLine[:idx]
	}
	firstLine = strings.TrimSpace(firstLine)

	if firstLine == "" || len(firstLine) > maxResponseCodeLength {
		return "", fmt.Errorf("smsbao: unexpected response body")
	}
	for _, r := range firstLine {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("smsbao: unexpected response body")
		}
	}

	return ResponseCode(firstLine), nil
}
