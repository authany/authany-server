package aliyun

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/util/uuid"
)

// This file implements the Aliyun RPC style signature (SignatureVersion 1.0).
// It is shared by the Aliyun products using the RPC style API, for example
// dysmsapi (Aliyun SMS) and dypnsapi (Aliyun SMS verification service).
//
// Written against
// https://www.alibabacloud.com/help/zh/sms/signature-method

const (
	SignatureMethod  = "HMAC-SHA1"
	SignatureVersion = "1.0"
	Format           = "JSON"
	// TimestampFormat is the format of the Timestamp common parameter.
	// The timestamp must be in UTC.
	TimestampFormat = "2006-01-02T15:04:05Z"
)

// PercentEncode encodes s according to RFC 3986, as required by the signature.
func PercentEncode(s string) string {
	encoded := url.QueryEscape(s)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}

// CanonicalizedQueryString sorts the parameters by key,
// and joins the percent encoded key-value pairs with "&".
// The Signature parameter, if any, is excluded.
func CanonicalizedQueryString(params url.Values) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key == "Signature" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, PercentEncode(key)+"="+PercentEncode(params.Get(key)))
	}
	return strings.Join(pairs, "&")
}

// StringToSign builds the string to sign of the given HTTP method and parameters.
func StringToSign(httpMethod string, params url.Values) string {
	return httpMethod + "&" + PercentEncode("/") + "&" + PercentEncode(CanonicalizedQueryString(params))
}

// SignRPCRequest returns the value of the Signature parameter.
// The key of the HMAC is the access key secret with a "&" appended.
func SignRPCRequest(httpMethod string, params url.Values, accessKeySecret string) string {
	mac := hmac.New(sha1.New, []byte(accessKeySecret+"&"))
	// hash.Hash never returns an error.
	_, _ = mac.Write([]byte(StringToSign(httpMethod, params)))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// CommonParams are the common parameters of an RPC style request.
type CommonParams struct {
	AccessKeyID string
	Action      string
	Version     string
	RegionID    string
	// Timestamp is optional. time.Now() is used when it is zero.
	Timestamp time.Time
	// SignatureNonce is optional. A new uuid is used when it is empty.
	SignatureNonce string
}

// BuildCommonParams builds the common parameters, without the Signature parameter.
func BuildCommonParams(p CommonParams) url.Values {
	timestamp := p.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	signatureNonce := p.SignatureNonce
	if signatureNonce == "" {
		signatureNonce = uuid.New()
	}

	values := url.Values{}
	values.Set("AccessKeyId", p.AccessKeyID)
	values.Set("Action", p.Action)
	values.Set("Version", p.Version)
	values.Set("Format", Format)
	values.Set("SignatureMethod", SignatureMethod)
	values.Set("SignatureVersion", SignatureVersion)
	values.Set("SignatureNonce", signatureNonce)
	values.Set("Timestamp", timestamp.UTC().Format(TimestampFormat))
	if p.RegionID != "" {
		values.Set("RegionId", p.RegionID)
	}
	return values
}
