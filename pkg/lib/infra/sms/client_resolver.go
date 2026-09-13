package sms

import (
	"fmt"
	"strconv"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/aliyun"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/custom"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/nexmo"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/tencent"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/twilio"
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

// rawClients is the set of clients resolved from a single configuration source,
// together with the credentials each client was constructed from.
type rawClients struct {
	nexmo              *nexmo.NexmoClient
	nexmoCredentials   *NexmoClientCredentials
	twilio             *twilio.TwilioClient
	twilioCredentials  *TwilioClientCredentials
	custom             *custom.CustomClient
	customCredentials  *CustomClientCredentials
	aliyun             *aliyun.AliyunClient
	aliyunCredentials  *AliyunClientCredentials
	tencent            *tencent.TencentClient
	tencentCredentials *TencentClientCredentials
}

type ClientResolver struct {
	AuthgearYAMLSMSProvider config.SMSProvider
	AuthgearYAMLSMSGateway  *config.SMSGatewayConfig

	AuthgearSecretsYAMLNexmoCredentials        *config.NexmoCredentials
	AuthgearSecretsYAMLTwilioCredentials       *config.TwilioCredentials
	AuthgearSecretsYAMLCustomSMSProviderConfig *config.CustomSMSProviderConfig
	AuthgearSecretsYAMLAliyunCredentials       *config.AliyunCredentials
	AuthgearSecretsYAMLTencentCredentials      *config.TencentCredentials

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

	type availableClient struct {
		RawClient            smsapi.Client
		SMSClientCredentials SMSClientCredentials
	}

	var client smsapi.Client
	var smsClientCredentials SMSClientCredentials
	switch provider {
	case config.SMSProviderNexmo:
		if raw.nexmo == nil {
			return nil, nil, smsapi.ErrNoAvailableClient
		}
		client = raw.nexmo
		smsClientCredentials = raw.nexmoCredentials
	case config.SMSProviderTwilio:
		if raw.twilio == nil {
			return nil, nil, smsapi.ErrNoAvailableClient
		}
		client = raw.twilio
		smsClientCredentials = raw.twilioCredentials
	case config.SMSProviderCustom:
		if raw.custom == nil {
			return nil, nil, smsapi.ErrNoAvailableClient
		}
		client = raw.custom
		smsClientCredentials = raw.customCredentials
	case config.SMSProviderAliyun:
		if raw.aliyun == nil {
			return nil, nil, smsapi.ErrNoAvailableClient
		}
		client = raw.aliyun
		smsClientCredentials = raw.aliyunCredentials
	case config.SMSProviderTencent:
		if raw.tencent == nil {
			return nil, nil, smsapi.ErrNoAvailableClient
		}
		client = raw.tencent
		smsClientCredentials = raw.tencentCredentials
	default:
		var availableClients []availableClient = []availableClient{}

		if raw.nexmo != nil {
			availableClients = append(availableClients, availableClient{
				RawClient:            raw.nexmo,
				SMSClientCredentials: raw.nexmoCredentials,
			})
		}
		if raw.twilio != nil {
			availableClients = append(availableClients, availableClient{
				RawClient:            raw.twilio,
				SMSClientCredentials: raw.twilioCredentials,
			})
		}
		if raw.custom != nil {
			availableClients = append(availableClients, availableClient{
				RawClient:            raw.custom,
				SMSClientCredentials: raw.customCredentials,
			})
		}
		if raw.aliyun != nil {
			availableClients = append(availableClients, availableClient{
				RawClient:            raw.aliyun,
				SMSClientCredentials: raw.aliyunCredentials,
			})
		}
		if raw.tencent != nil {
			availableClients = append(availableClients, availableClient{
				RawClient:            raw.tencent,
				SMSClientCredentials: raw.tencentCredentials,
			})
		}
		if len(availableClients) == 0 {
			return nil, nil, smsapi.ErrNoAvailableClient
		}
		if len(availableClients) > 1 {
			return nil, nil, smsapi.ErrAmbiguousClient
		}
		client = availableClients[0].RawClient
		smsClientCredentials = availableClients[0].SMSClientCredentials
	}
	return client, smsClientCredentials, nil
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
	var tencentClientCredentials *TencentClientCredentials

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

	return rawClients{
		nexmo:              nexmo.NewNexmoClient(r.AuthgearSecretsYAMLNexmoCredentials),
		nexmoCredentials:   nexmoClientCredentials,
		twilio:             twilio.NewTwilioClient(r.AuthgearSecretsYAMLTwilioCredentials),
		twilioCredentials:  twilioClientCredentials,
		custom:             custom.NewCustomClient(r.AuthgearSecretsYAMLCustomSMSProviderConfig, r.SMSDenoHook, r.SMSWebHook),
		customCredentials:  customClientCredentials,
		aliyun:             aliyun.NewAliyunClient(r.AuthgearSecretsYAMLAliyunCredentials),
		aliyunCredentials:  aliyunClientCredentials,
		tencent:            tencent.NewTencentClient(r.AuthgearSecretsYAMLTencentCredentials),
		tencentCredentials: tencentClientCredentials,
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
