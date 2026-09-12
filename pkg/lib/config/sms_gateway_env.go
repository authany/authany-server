package config

type SMSGatewayEnvironmentTwilioCredentials struct {
	AccountSID          string `envconfig:"ACCOUNT_SID"`
	AuthToken           string `envconfig:"AUTH_TOKEN"`
	MessagingServiceSID string `envconfig:"MESSAGING_SERVICE_SID"`
}

type SMSGatewayEnvironmentNexmoCredentials struct {
	APIKey    string `envconfig:"API_KEY"`
	APISecret string `envconfig:"API_SECRET"`
}

type SMSGatewayEnvironmentCustomSMSProviderConfig struct {
	URL     string `envconfig:"URL"`
	Timeout string `envconfig:"TIMEOUT"`
}

type SMSGatewayEnvironmentDefaultUseConfigFrom string

const (
	SMSGatewayEnvironmentDefaultUseConfigFromEnvironmentVariable SMSGatewayEnvironmentDefaultUseConfigFrom = "environment_variable"
	SMSGatewayEnvironmentDefaultUseConfigFromAuthgearSecretsYAML SMSGatewayEnvironmentDefaultUseConfigFrom = "authgear.secrets.yaml"
)

type SMSGatewayEnvironmentDefaultProvider string

const (
	SMSGatewayEnvironmentDefaultProviderNexmo  SMSGatewayEnvironmentDefaultProvider = "nexmo"
	SMSGatewayEnvironmentDefaultProviderTwilio SMSGatewayEnvironmentDefaultProvider = "twilio"
	SMSGatewayEnvironmentDefaultProviderCustom SMSGatewayEnvironmentDefaultProvider = "custom"
)

type SMSGatewayEnvironmentDefaultConfig struct {
	UseConfigFrom SMSGatewayEnvironmentDefaultUseConfigFrom `envconfig:"USE_CONFIG_FROM"`
	Provider      SMSGatewayEnvironmentDefaultProvider      `envconfig:"PROVIDER"`
}

// SMSGatewayEnvironmentConfig intentionally only covers twilio, nexmo and
// custom. The SMS providers added after them are configured per project in
// authgear.secrets.yaml only, because a deployment-wide default is meaningful
// only for the providers that the deployment operator, instead of the project
// owner, has an account of.
type SMSGatewayEnvironmentConfig struct {
	Twilio  SMSGatewayEnvironmentTwilioCredentials       `envconfig:"TWILIO"`
	Nexmo   SMSGatewayEnvironmentNexmoCredentials        `envconfig:"NEXMO"`
	Custom  SMSGatewayEnvironmentCustomSMSProviderConfig `envconfig:"CUSTOM"`
	Default SMSGatewayEnvironmentDefaultConfig           `envconfig:"DEFAULT"`
}
