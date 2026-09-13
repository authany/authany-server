package model

import "github.com/authgear/authgear-server/pkg/lib/config"

type SMSProviderConfigurationInput struct {
	Twilio  *SMSProviderConfigurationTwilioInput  `json:"twilio,omitempty"`
	Aliyun  *SMSProviderConfigurationAliyunInput  `json:"aliyun,omitempty"`
	Tencent *SMSProviderConfigurationTencentInput `json:"tencent,omitempty"`
	Webhook *SMSProviderConfigurationWebhookInput `json:"webhook,omitempty"`
	Deno    *SMSProviderConfigurationDenoInput    `json:"deno,omitempty"`
}

type SMSProviderConfigurationAliyunInput struct {
	AccessKeyID          string            `json:"accessKeyID,omitempty"`
	AccessKeySecret      string            `json:"accessKeySecret,omitempty"`
	SignName             string            `json:"signName,omitempty"`
	TemplateCode         string            `json:"templateCode,omitempty"`
	TemplateCodes        map[string]string `json:"templateCodes,omitempty"`
	OverseasTemplateCode string            `json:"overseasTemplateCode,omitempty"`
}

type SMSProviderConfigurationTencentInput struct {
	SecretID      string            `json:"secretID,omitempty"`
	SecretKey     string            `json:"secretKey,omitempty"`
	SDKAppID      string            `json:"sdkAppID,omitempty"`
	Region        string            `json:"region,omitempty"`
	SignName      string            `json:"signName,omitempty"`
	TemplateCode  string            `json:"templateCode,omitempty"`
	TemplateCodes map[string]string `json:"templateCodes,omitempty"`
}

type SMSProviderConfigurationTwilioInput struct {
	CredentialType      config.TwilioCredentialType `json:"credentialType,omitempty"`
	AccountSID          string                      `json:"accountSID,omitempty"`
	AuthToken           string                      `json:"authToken,omitempty"`
	APIKeySID           string                      `json:"apiKeySID,omitempty"`
	APIKeySecret        string                      `json:"apiKeySecret,omitempty"`
	MessagingServiceSID string                      `json:"messagingServiceSID,omitempty"`
	From                string                      `json:"from,omitempty"`
}

type SMSProviderConfigurationWebhookInput struct {
	URL     string `json:"url,omitempty"`
	Timeout *int   `json:"timeout,omitempty"`
}

type SMSProviderConfigurationDenoInput struct {
	Script  string `json:"script,omitempty"`
	Timeout *int   `json:"timeout,omitempty"`
}
