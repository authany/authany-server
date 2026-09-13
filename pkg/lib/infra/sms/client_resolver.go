package sms

import (
	"fmt"
	"strconv"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/aliyun"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/aliyunmas"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/custom"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/gatewayapi"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/nexmo"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsaero"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsbao"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/tencent"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/twilio"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/yunpian"
)

func NewTwilioClientCredentialsFromSecrets(secret *config.TwilioCredentials) *TwilioClientCredentials {
	return &TwilioClientCredentials{
		CredentialType:      secret.GetCredentialType(),
		AccountSID:          secret.AccountSID,
		AuthToken:           secret.AuthToken,
		APIKeySID:           secret.APIKeySID,
		APIKeySecret:        secret.APIKeySecret,
		MessagingServiceSID: secret.MessagingServiceSID,
	}
}

type TwilioClientCredentials struct {
	CredentialType      config.TwilioCredentialType
	AccountSID          string
	AuthToken           string
	APIKeySID           string
	APIKeySecret        string
	MessagingServiceSID string
	From                string
}

func (c *TwilioClientCredentials) toSecret() *config.TwilioCredentials {
	if c == nil {
		return nil
	}
	return &config.TwilioCredentials{
		CredentialType_WriteOnly: &c.CredentialType,
		AccountSID:               c.AccountSID,
		AuthToken:                c.AuthToken,
		APIKeySID:                c.APIKeySID,
		APIKeySecret:             c.APIKeySecret,
		MessagingServiceSID:      c.MessagingServiceSID,
		From:                     c.From,
	}
}

func (TwilioClientCredentials) smsClientCredentials() {}

type NexmoClientCredentials struct {
	APIKey    string
	APISecret string
}

func (NexmoClientCredentials) smsClientCredentials() {}

type CustomClientCredentials struct {
	URL     string
	Timeout *config.DurationSeconds
}

func (CustomClientCredentials) smsClientCredentials() {}

type AliyunClientCredentials struct {
	AccessKeyID          string
	AccessKeySecret      string
	SignName             string
	TemplateCode         string
	TemplateCodes        map[string]string
	OverseasTemplateCode string
}

func (AliyunClientCredentials) smsClientCredentials() {}

type TencentClientCredentials struct {
	SecretID      string
	SecretKey     string
	SDKAppID      string
	Region        string
	SignName      string
	TemplateCode  string
	TemplateCodes map[string]string
}

func (TencentClientCredentials) smsClientCredentials() {}

type AliyunMASClientCredentials struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
	TemplateCodes   map[string]string
}

func (AliyunMASClientCredentials) smsClientCredentials() {}

type YunpianClientCredentials struct {
	APIKey string
}

func (YunpianClientCredentials) smsClientCredentials() {}

type SmsbaoClientCredentials struct {
	Username         string
	PasswordOrAPIKey string
	GoodsID          string
}

func (SmsbaoClientCredentials) smsClientCredentials() {}

type GatewayAPIClientCredentials struct {
	Endpoint string
	APIToken string
	Sender   string
}

func (GatewayAPIClientCredentials) smsClientCredentials() {}

type SmsAeroClientCredentials struct {
	Email      string
	APIKey     string
	SenderName string
}

func (SmsAeroClientCredentials) smsClientCredentials() {}

// rawClients is the set of clients resolved from a single configuration source,
// together with the credentials each client was constructed from.
type rawClients struct {
	nexmo                 *nexmo.NexmoClient
	nexmoCredentials      *NexmoClientCredentials
	twilio                *twilio.TwilioClient
	twilioCredentials     *TwilioClientCredentials
	custom                *custom.CustomClient
	customCredentials     *CustomClientCredentials
	aliyun                *aliyun.AliyunClient
	aliyunCredentials     *AliyunClientCredentials
	aliyunMAS             *aliyunmas.AliyunMASClient
	aliyunMASCredentials  *AliyunMASClientCredentials
	tencent               *tencent.TencentClient
	tencentCredentials    *TencentClientCredentials
	yunpian               *yunpian.YunpianClient
	yunpianCredentials    *YunpianClientCredentials
	smsbao                *smsbao.SmsbaoClient
	smsbaoCredentials     *SmsbaoClientCredentials
	gatewayAPI            *gatewayapi.GatewayAPIClient
	gatewayAPICredentials *GatewayAPIClientCredentials
	smsAero               *smsaero.SmsAeroClient
	smsAeroCredentials    *SmsAeroClientCredentials
}

// availableClient is a resolved client together with the credentials it was
// constructed from. RawClient is nil when the client is not available.
type availableClient struct {
	RawClient            smsapi.Client
	SMSClientCredentials SMSClientCredentials
}

// providerEntry associates a provider name with its resolved client.
type providerEntry struct {
	provider config.SMSProvider
	client   availableClient
}

// newAvailableClient keeps RawClient nil when the concrete client pointer is
// nil, so that the interface value stays nil.
func newAvailableClient[T interface {
	smsapi.Client
	comparable
}](client T, credentials SMSClientCredentials) availableClient {
	var zero T
	if client == zero {
		return availableClient{}
	}
	return availableClient{
		RawClient:            client,
		SMSClientCredentials: credentials,
	}
}

// entries lists every provider in the order used to detect ambiguity.
func (c rawClients) entries() []providerEntry {
	return []providerEntry{
		{config.SMSProviderNexmo, newAvailableClient(c.nexmo, c.nexmoCredentials)},
		{config.SMSProviderTwilio, newAvailableClient(c.twilio, c.twilioCredentials)},
		{config.SMSProviderCustom, newAvailableClient(c.custom, c.customCredentials)},
		{config.SMSProviderAliyun, newAvailableClient(c.aliyun, c.aliyunCredentials)},
		{config.SMSProviderTencent, newAvailableClient(c.tencent, c.tencentCredentials)},
		{config.SMSProviderAliyunMAS, newAvailableClient(c.aliyunMAS, c.aliyunMASCredentials)},
		{config.SMSProviderYunpian, newAvailableClient(c.yunpian, c.yunpianCredentials)},
		{config.SMSProviderSmsbao, newAvailableClient(c.smsbao, c.smsbaoCredentials)},
		{config.SMSProviderGatewayAPI, newAvailableClient(c.gatewayAPI, c.gatewayAPICredentials)},
		{config.SMSProviderSmsAero, newAvailableClient(c.smsAero, c.smsAeroCredentials)},
	}
}

// byProvider returns the client of the given provider.
// known is false when provider is not a known provider name.
func (c rawClients) byProvider(provider config.SMSProvider) (resolved availableClient, known bool) {
	for _, entry := range c.entries() {
		if entry.provider == provider {
			return entry.client, true
		}
	}
	return availableClient{}, false
}

// availableClients returns the clients that are available.
func (c rawClients) availableClients() []availableClient {
	var availableClients []availableClient = []availableClient{}
	for _, entry := range c.entries() {
		if entry.client.RawClient != nil {
			availableClients = append(availableClients, entry.client)
		}
	}
	return availableClients
}

type ClientResolver struct {
	AuthgearYAMLSMSProvider config.SMSProvider
	AuthgearYAMLSMSGateway  *config.SMSGatewayConfig

	AuthgearSecretsYAMLNexmoCredentials        *config.NexmoCredentials
	AuthgearSecretsYAMLTwilioCredentials       *config.TwilioCredentials
	AuthgearSecretsYAMLCustomSMSProviderConfig *config.CustomSMSProviderConfig
	AuthgearSecretsYAMLAliyunCredentials       *config.AliyunCredentials
	AuthgearSecretsYAMLAliyunMASCredentials    *config.AliyunMASCredentials
	AuthgearSecretsYAMLTencentCredentials      *config.TencentCredentials
	AuthgearSecretsYAMLYunpianCredentials      *config.YunpianCredentials
	AuthgearSecretsYAMLSmsbaoCredentials       *config.SmsbaoCredentials
	AuthgearSecretsYAMLGatewayAPICredentials   *config.GatewayAPICredentials
	AuthgearSecretsYAMLSmsAeroCredentials      *config.SmsAeroCredentials

	EnvironmentDefaultProvider      config.SMSGatewayEnvironmentDefaultProvider
	EnvironmentDefaultUseConfigFrom config.SMSGatewayEnvironmentDefaultUseConfigFrom

	EnvironmentNexmoCredentials        config.SMSGatewayEnvironmentNexmoCredentials
	EnvironmentTwilioCredentials       config.SMSGatewayEnvironmentTwilioCredentials
	EnvironmentCustomSMSProviderConfig config.SMSGatewayEnvironmentCustomSMSProviderConfig

	SMSDenoHook custom.SMSDenoHook
	SMSWebHook  custom.SMSWebHook
}

func (r *ClientResolver) ResolveClient() (smsapi.Client, SMSClientCredentials, error) {
	raw := r.resolveRawClients()
	provider := r.resolveProvider()

	if resolved, known := raw.byProvider(provider); known {
		if resolved.RawClient == nil {
			return nil, nil, smsapi.ErrNoAvailableClient
		}
		return resolved.RawClient, resolved.SMSClientCredentials, nil
	}

	// provider is not a known provider name.
	// It is determined from application logic instead.
	availableClients := raw.availableClients()
	if len(availableClients) == 0 {
		return nil, nil, smsapi.ErrNoAvailableClient
	}
	if len(availableClients) > 1 {
		return nil, nil, smsapi.ErrAmbiguousClient
	}
	return availableClients[0].RawClient, availableClients[0].SMSClientCredentials, nil
}

func (r *ClientResolver) resolveProvider() config.SMSProvider {
	if r.AuthgearYAMLSMSGateway != nil {
		// Use sms gateway config. See Table 3
		return r.resolveProviderFromAuthgearYAMLAndAuthgearSecretsYAML()
	}
	if r.AuthgearYAMLSMSProvider != "" {
		// Use `messaging.sms_provider` from `authgear.yaml`. Read config from `sms.{messaging.sms_provider}` from `authgear.secrets.yaml`
		return r.AuthgearYAMLSMSProvider
	}
	// sms_provider == "" and sms_gateway == nil
	// See table 2
	return r.resolveProviderFromEnv()
}

// Table 2
func (r *ClientResolver) resolveProviderFromEnv() config.SMSProvider {
	if r.EnvironmentDefaultUseConfigFrom == "" {
		// `provider` will be determined from application logic. Read config from `sms.{provider}` from `authgear.secrets.yaml`
		return ""
	}
	switch r.EnvironmentDefaultUseConfigFrom {
	case config.SMSGatewayEnvironmentDefaultUseConfigFromEnvironmentVariable:
		if r.EnvironmentDefaultProvider == "" {
			// `provider` will be determined from application logic. Read config from `SMS_GATEWAY_{provider}_*` from environment variables
			return ""
		}
		// Use `SMS_GATEWAY_DEFAULT_PROVIDER` as provider. Will read config from `SMS_GATEWAY_{SMS_GATEWAY_DEFAULT_PROVIDER}_*` environment variables
		return config.SMSProvider(r.EnvironmentDefaultProvider)
	case config.SMSGatewayEnvironmentDefaultUseConfigFromAuthgearSecretsYAML:
		// `provider` will be determined from application logic. Read config from `sms.{provider}` from `authgear.secrets.yaml`
		return ""
	default:
		panic(fmt.Errorf("Invalid DEFAULT_USE_CONFIG_FROM %v", r.EnvironmentDefaultUseConfigFrom))
	}
}

// Table 3
func (r *ClientResolver) resolveProviderFromAuthgearYAMLAndAuthgearSecretsYAML() config.SMSProvider {
	AuthgearYAMLUseConfigFrom := r.AuthgearYAMLSMSGateway.UseConfigFrom
	switch AuthgearYAMLUseConfigFrom {
	case config.SMSGatewayUseConfigFromEnvironmentVariable:
		if r.AuthgearYAMLSMSGateway.Provider == "" {
			if r.EnvironmentDefaultProvider == "" {
				// provider` will be determined from application logic. Read config from `SMS_GATEWAY_{provider}_*` from environment variables
				return ""
			}
			// Use `SMS_GATEWAY_DEFAULT_PROVIDER` as provider. Will read config from `SMS_GATEWAY_{SMS_GATEWAY_DEFAULT_PROVIDER}_*` environment variables
			return config.SMSProvider(r.EnvironmentDefaultProvider)
		}
		// Use `sms_gateway.provider` as provider. Will read config from `SMS_GATEWAY_{sms_gateway.provider}_*` environment variables
		return r.AuthgearYAMLSMSGateway.Provider
	case config.SMSGatewayUseConfigFromAuthgearSecretsYAML:
		// `sms_gateway.provider` is required
		// Use provider configs from `authgear.yaml`. Will read config from `sms.{sms_gateway.provider}` from `authgear.secrets.yaml`
		return r.AuthgearYAMLSMSGateway.Provider
	default:
		panic(fmt.Errorf("Invalid sms_gateway.use_config_from %v", AuthgearYAMLUseConfigFrom))
	}
}

func (r *ClientResolver) clientsFromAuthgearSecretsYAML() rawClients {
	var nexmoClientCredentials *NexmoClientCredentials
	var twilioClientCredentials *TwilioClientCredentials
	var customClientCredentials *CustomClientCredentials
	var aliyunClientCredentials *AliyunClientCredentials
	var aliyunMASClientCredentials *AliyunMASClientCredentials
	var tencentClientCredentials *TencentClientCredentials
	var yunpianClientCredentials *YunpianClientCredentials
	var smsbaoClientCredentials *SmsbaoClientCredentials
	var gatewayAPIClientCredentials *GatewayAPIClientCredentials
	var smsAeroClientCredentials *SmsAeroClientCredentials

	if r.AuthgearSecretsYAMLNexmoCredentials != nil {
		nexmoClientCredentials = &NexmoClientCredentials{
			APIKey:    r.AuthgearSecretsYAMLNexmoCredentials.APIKey,
			APISecret: r.AuthgearSecretsYAMLNexmoCredentials.APISecret,
		}
	}

	if r.AuthgearSecretsYAMLTwilioCredentials != nil {
		credtyp := r.AuthgearSecretsYAMLTwilioCredentials.GetCredentialType()
		twilioClientCredentials = &TwilioClientCredentials{
			CredentialType:      credtyp,
			AccountSID:          r.AuthgearSecretsYAMLTwilioCredentials.AccountSID,
			AuthToken:           r.AuthgearSecretsYAMLTwilioCredentials.AuthToken,
			APIKeySID:           r.AuthgearSecretsYAMLTwilioCredentials.APIKeySID,
			APIKeySecret:        r.AuthgearSecretsYAMLTwilioCredentials.APIKeySecret,
			MessagingServiceSID: r.AuthgearSecretsYAMLTwilioCredentials.MessagingServiceSID,
			From:                r.AuthgearSecretsYAMLTwilioCredentials.From,
		}
	}

	if r.AuthgearSecretsYAMLCustomSMSProviderConfig != nil {
		customClientCredentials = &CustomClientCredentials{
			URL:     r.AuthgearSecretsYAMLCustomSMSProviderConfig.URL,
			Timeout: r.AuthgearSecretsYAMLCustomSMSProviderConfig.Timeout,
		}
	}

	if r.AuthgearSecretsYAMLAliyunCredentials != nil {
		aliyunClientCredentials = &AliyunClientCredentials{
			AccessKeyID:          r.AuthgearSecretsYAMLAliyunCredentials.AccessKeyID,
			AccessKeySecret:      r.AuthgearSecretsYAMLAliyunCredentials.AccessKeySecret,
			SignName:             r.AuthgearSecretsYAMLAliyunCredentials.SignName,
			TemplateCode:         r.AuthgearSecretsYAMLAliyunCredentials.TemplateCode,
			TemplateCodes:        r.AuthgearSecretsYAMLAliyunCredentials.TemplateCodes,
			OverseasTemplateCode: r.AuthgearSecretsYAMLAliyunCredentials.OverseasTemplateCode,
		}
	}

	if r.AuthgearSecretsYAMLTencentCredentials != nil {
		tencentClientCredentials = &TencentClientCredentials{
			SecretID:      r.AuthgearSecretsYAMLTencentCredentials.SecretID,
			SecretKey:     r.AuthgearSecretsYAMLTencentCredentials.SecretKey,
			SDKAppID:      r.AuthgearSecretsYAMLTencentCredentials.SDKAppID,
			Region:        r.AuthgearSecretsYAMLTencentCredentials.Region,
			SignName:      r.AuthgearSecretsYAMLTencentCredentials.SignName,
			TemplateCode:  r.AuthgearSecretsYAMLTencentCredentials.TemplateCode,
			TemplateCodes: r.AuthgearSecretsYAMLTencentCredentials.TemplateCodes,
		}
	}

	if r.AuthgearSecretsYAMLAliyunMASCredentials != nil {
		aliyunMASClientCredentials = &AliyunMASClientCredentials{
			AccessKeyID:     r.AuthgearSecretsYAMLAliyunMASCredentials.AccessKeyID,
			AccessKeySecret: r.AuthgearSecretsYAMLAliyunMASCredentials.AccessKeySecret,
			SignName:        r.AuthgearSecretsYAMLAliyunMASCredentials.SignName,
			TemplateCode:    r.AuthgearSecretsYAMLAliyunMASCredentials.TemplateCode,
			TemplateCodes:   r.AuthgearSecretsYAMLAliyunMASCredentials.TemplateCodes,
		}
	}

	if r.AuthgearSecretsYAMLYunpianCredentials != nil {
		yunpianClientCredentials = &YunpianClientCredentials{
			APIKey: r.AuthgearSecretsYAMLYunpianCredentials.APIKey,
		}
	}

	if r.AuthgearSecretsYAMLSmsbaoCredentials != nil {
		smsbaoClientCredentials = &SmsbaoClientCredentials{
			Username:         r.AuthgearSecretsYAMLSmsbaoCredentials.Username,
			PasswordOrAPIKey: r.AuthgearSecretsYAMLSmsbaoCredentials.PasswordOrAPIKey,
			GoodsID:          r.AuthgearSecretsYAMLSmsbaoCredentials.GoodsID,
		}
	}

	if r.AuthgearSecretsYAMLGatewayAPICredentials != nil {
		gatewayAPIClientCredentials = &GatewayAPIClientCredentials{
			Endpoint: r.AuthgearSecretsYAMLGatewayAPICredentials.Endpoint,
			APIToken: r.AuthgearSecretsYAMLGatewayAPICredentials.APIToken,
			Sender:   r.AuthgearSecretsYAMLGatewayAPICredentials.Sender,
		}
	}

	if r.AuthgearSecretsYAMLSmsAeroCredentials != nil {
		smsAeroClientCredentials = &SmsAeroClientCredentials{
			Email:      r.AuthgearSecretsYAMLSmsAeroCredentials.Email,
			APIKey:     r.AuthgearSecretsYAMLSmsAeroCredentials.APIKey,
			SenderName: r.AuthgearSecretsYAMLSmsAeroCredentials.SenderName,
		}
	}

	return rawClients{
		nexmo:                 nexmo.NewNexmoClient(r.AuthgearSecretsYAMLNexmoCredentials),
		nexmoCredentials:      nexmoClientCredentials,
		twilio:                twilio.NewTwilioClient(r.AuthgearSecretsYAMLTwilioCredentials),
		twilioCredentials:     twilioClientCredentials,
		custom:                custom.NewCustomClient(r.AuthgearSecretsYAMLCustomSMSProviderConfig, r.SMSDenoHook, r.SMSWebHook),
		customCredentials:     customClientCredentials,
		aliyun:                aliyun.NewAliyunClient(r.AuthgearSecretsYAMLAliyunCredentials),
		aliyunCredentials:     aliyunClientCredentials,
		aliyunMAS:             aliyunmas.NewAliyunMASClient(r.AuthgearSecretsYAMLAliyunMASCredentials),
		aliyunMASCredentials:  aliyunMASClientCredentials,
		tencent:               tencent.NewTencentClient(r.AuthgearSecretsYAMLTencentCredentials),
		tencentCredentials:    tencentClientCredentials,
		yunpian:               yunpian.NewYunpianClient(r.AuthgearSecretsYAMLYunpianCredentials),
		yunpianCredentials:    yunpianClientCredentials,
		smsbao:                smsbao.NewSmsbaoClient(r.AuthgearSecretsYAMLSmsbaoCredentials),
		smsbaoCredentials:     smsbaoClientCredentials,
		gatewayAPI:            gatewayapi.NewGatewayAPIClient(r.AuthgearSecretsYAMLGatewayAPICredentials),
		gatewayAPICredentials: gatewayAPIClientCredentials,
		smsAero:               smsaero.NewSmsAeroClient(r.AuthgearSecretsYAMLSmsAeroCredentials),
		smsAeroCredentials:    smsAeroClientCredentials,
	}
}

func (r *ClientResolver) clientsFromEnv() rawClients {
	var nexmoClientCredentials *NexmoClientCredentials
	var twilioClientCredentials *TwilioClientCredentials
	var customClientCredentials *CustomClientCredentials

	if r.EnvironmentNexmoCredentials != (config.SMSGatewayEnvironmentNexmoCredentials{}) {
		nexmoClientCredentials = &NexmoClientCredentials{
			APIKey:    r.EnvironmentNexmoCredentials.APIKey,
			APISecret: r.EnvironmentNexmoCredentials.APISecret,
		}
	}

	if r.EnvironmentTwilioCredentials != (config.SMSGatewayEnvironmentTwilioCredentials{}) {
		credtyp := config.TwilioCredentialTypeAuthToken
		twilioClientCredentials = &TwilioClientCredentials{
			CredentialType:      credtyp,
			AccountSID:          r.EnvironmentTwilioCredentials.AccountSID,
			AuthToken:           r.EnvironmentTwilioCredentials.AuthToken,
			APIKeySID:           "",
			APIKeySecret:        "",
			MessagingServiceSID: r.EnvironmentTwilioCredentials.MessagingServiceSID,
			From:                "",
		}
	}

	if r.EnvironmentCustomSMSProviderConfig != (config.SMSGatewayEnvironmentCustomSMSProviderConfig{}) {
		timeoutInt, _ := strconv.Atoi(r.EnvironmentCustomSMSProviderConfig.Timeout)
		var timeout *config.DurationSeconds
		timeout = new(config.DurationSeconds)
		*timeout = config.DurationSeconds(timeoutInt)
		customClientCredentials = &CustomClientCredentials{
			URL:     r.EnvironmentCustomSMSProviderConfig.URL,
			Timeout: timeout,
		}
	}

	// SMSGatewayEnvironmentConfig intentionally does not cover aliyun and tencent.
	return rawClients{
		nexmo:             nexmo.NewNexmoClient((*config.NexmoCredentials)(nexmoClientCredentials)),
		nexmoCredentials:  nexmoClientCredentials,
		twilio:            twilio.NewTwilioClient(twilioClientCredentials.toSecret()),
		twilioCredentials: twilioClientCredentials,
		custom:            custom.NewCustomClient((*config.CustomSMSProviderConfig)(customClientCredentials), r.SMSDenoHook, r.SMSWebHook),
		customCredentials: customClientCredentials,
	}
}

func (r *ClientResolver) resolveRawClients() rawClients {
	if r.AuthgearYAMLSMSGateway != nil {
		// Use sms gateway config. See Table 3
		return r.resolveConfigFromAuthgearYAMLAndAuthgearSecretsYAML()
	}
	if r.AuthgearYAMLSMSProvider != "" {
		// Use `messaging.sms_provider` from `authgear.yaml`. Read config from `sms.{messaging.sms_provider}` from `authgear.secrets.yaml`
		return r.clientsFromAuthgearSecretsYAML()
	}
	// sms_provider == "" and sms_gateway == nil
	// See table 2
	return r.resolveConfigFromEnv()
}

// Table 2
func (r *ClientResolver) resolveConfigFromEnv() rawClients {
	if r.EnvironmentDefaultUseConfigFrom == "" {
		// `provider` will be determined from application logic. Read config from `sms.{provider}` from `authgear.secrets.yaml`
		return r.clientsFromAuthgearSecretsYAML()
	}
	switch r.EnvironmentDefaultUseConfigFrom {
	case config.SMSGatewayEnvironmentDefaultUseConfigFromEnvironmentVariable:
		if r.EnvironmentDefaultProvider == "" {
			// `provider` will be determined from application logic. Read config from `SMS_GATEWAY_{provider}_*` from environment variables
			return r.clientsFromEnv()
		}
		// Use `SMS_GATEWAY_DEFAULT_PROVIDER` as provider. Will read config from `SMS_GATEWAY_{SMS_GATEWAY_DEFAULT_PROVIDER}_*` environment variables
		return r.clientsFromEnv()
	case config.SMSGatewayEnvironmentDefaultUseConfigFromAuthgearSecretsYAML:
		// `provider` will be determined from application logic. Read config from `sms.{provider}` from `authgear.secrets.yaml`
		return r.clientsFromAuthgearSecretsYAML()
	default:
		panic(fmt.Errorf("Invalid DEFAULT_USE_CONFIG_FROM %v", r.EnvironmentDefaultUseConfigFrom))
	}
}

// Table 3
func (r *ClientResolver) resolveConfigFromAuthgearYAMLAndAuthgearSecretsYAML() rawClients {
	switch r.AuthgearYAMLSMSGateway.UseConfigFrom {
	case config.SMSGatewayUseConfigFromEnvironmentVariable:
		if r.AuthgearYAMLSMSGateway.Provider == "" {
			if r.EnvironmentDefaultProvider == "" {
				// provider` will be determined from application logic. Read config from `SMS_GATEWAY_{provider}_*` from environment variables
				return r.clientsFromEnv()
			}
			// Use `SMS_GATEWAY_DEFAULT_PROVIDER` as provider. Will read config from `SMS_GATEWAY_{SMS_GATEWAY_DEFAULT_PROVIDER}_*` environment variables
			return r.clientsFromEnv()
		}
		// Use `sms_gateway.provider` as provider. Will read config from `SMS_GATEWAY_{sms_gateway.provider}_*` environment variables
		return r.clientsFromEnv()
	case config.SMSGatewayUseConfigFromAuthgearSecretsYAML:
		// `sms_gateway.provider` is required
		// Use provider configs from `authgear.yaml`. Will read config from `sms.{sms_gateway.provider}` from `authgear.secrets.yaml`
		return r.clientsFromAuthgearSecretsYAML()
	default:
		panic(fmt.Errorf("Invalid sms_gateway.use_config_from %v", r.AuthgearYAMLSMSGateway.UseConfigFrom))
	}
}
