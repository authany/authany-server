package model

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

// PortalFeatureConfig is the portal API representation of FeatureConfig.
// It contains only the fields relevant to the portal UI.
type PortalFeatureConfig struct {
	Identity         *config.IdentityFeatureConfig         `json:"identity,omitempty"`
	Authentication   *config.AuthenticationFeatureConfig   `json:"authentication,omitempty"`
	Authenticator    *config.AuthenticatorFeatureConfig    `json:"authenticator,omitempty"`
	CustomDomain     *config.CustomDomainFeatureConfig     `json:"custom_domain,omitempty"`
	UI               *config.UIFeatureConfig               `json:"ui,omitempty"`
	OAuth            *config.OAuthFeatureConfig            `json:"oauth,omitempty"`
	Hook             *config.HookFeatureConfig             `json:"hook,omitempty"`
	AuditLog         *config.AuditLogFeatureConfig         `json:"audit_log,omitempty"`
	GoogleTagManager *config.GoogleTagManagerFeatureConfig `json:"google_tag_manager,omitempty"`
	Messaging        *PortalMessagingFeatureConfig         `json:"messaging,omitempty"`
	Collaborator     *config.CollaboratorFeatureConfig     `json:"collaborator,omitempty"`
	FraudProtection  *config.FraudProtectionFeatureConfig  `json:"fraud_protection,omitempty"`
}

// PortalMessagingFeatureConfig exposes only the messaging feature flags
// relevant to the portal UI (rate limits and usage fields are excluded).
type PortalMessagingFeatureConfig struct {
	TemplateCustomizationDisabled *bool `json:"template_customization_disabled,omitempty"`
	CustomSMTPDisabled            *bool `json:"custom_smtp_disabled,omitempty"`
	CustomSMSProviderDisabled     *bool `json:"custom_sms_provider_disabled,omitempty"`
}

func NewPortalFeatureConfig(c *config.FeatureConfig) *PortalFeatureConfig {
	if c == nil {
		return nil
	}
	out := &PortalFeatureConfig{
		Identity:         c.Identity,
		Authentication:   c.Authentication,
		Authenticator:    c.Authenticator,
		CustomDomain:     c.CustomDomain,
		UI:               c.UI,
		OAuth:            c.OAuth,
		Hook:             c.Hook,
		AuditLog:         c.AuditLog,
		GoogleTagManager: c.GoogleTagManager,
		Collaborator:     c.Collaborator,
		FraudProtection:  c.FraudProtection,
	}
	if c.Messaging != nil {
		out.Messaging = &PortalMessagingFeatureConfig{
			TemplateCustomizationDisabled: c.Messaging.TemplateCustomizationDisabled,
			CustomSMTPDisabled:            c.Messaging.CustomSMTPDisabled,
			CustomSMSProviderDisabled:     c.Messaging.CustomSMSProviderDisabled,
		}
	}
	return out
}

type AppListItem struct {
	AppID        string `json:"appID,omitempty"`
	PublicOrigin string `json:"publicOrigin,omitempty"`
}

type App struct {
	ID      string
	Context *config.AppContext
}

type AppResource struct {
	DescriptedPath resource.DescriptedPath
	Context        *config.AppContext
}

type WebhookSecret struct {
	Secret *string `json:"secret,omitempty"`
}

type OAuthSSOProviderClientSecret struct {
	Alias        string  `json:"alias,omitempty"`
	ClientSecret *string `json:"clientSecret,omitempty"`
}

type AdminAPISecret struct {
	KeyID         string     `json:"keyID,omitempty"`
	CreatedAt     *time.Time `json:"createdAt,omitempty"`
	PublicKeyPEM  string     `json:"publicKeyPEM,omitempty"`
	PrivateKeyPEM *string    `json:"privateKeyPEM,omitempty"`
}

type SMTPSecret struct {
	Host     string  `json:"host,omitempty"`
	Port     int     `json:"port,omitempty"`
	Username string  `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	Sender   string  `json:"sender,omitempty"`
}

type OAuthClientSecretKey struct {
	KeyID     string     `json:"keyID,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	Key       string     `json:"key,omitempty"`
}

type OAuthClientSecret struct {
	ClientID string                 `json:"clientID,omitempty"`
	Keys     []OAuthClientSecretKey `json:"keys,omitempty"`
}

type BotProtectionProviderSecret struct {
	Type      config.BotProtectionProviderType `json:"type,omitempty"`
	SecretKey *string                          `json:"secretKey,omitempty"`
}

type SAMLIdpSigningCertificate struct {
	KeyID                  string `json:"keyID,omitempty"`
	CertificateFingerprint string `json:"certificateFingerprint,omitempty"`
	CertificatePEM         string `json:"certificatePEM,omitempty"`
}

type SAMLIdpSigningSecrets struct {
	Certificates []SAMLIdpSigningCertificate `json:"certificates,omitempty"`
}

type SAMLSpSigningCertificate struct {
	CertificateFingerprint string `json:"certificateFingerprint,omitempty"`
	CertificatePEM         string `json:"certificatePEM,omitempty"`
}

type SAMLSpSigningSecrets struct {
	ClientID     string                     `json:"clientID,omitempty"`
	Certificates []SAMLSpSigningCertificate `json:"certificates,omitempty"`
}

type SMSProviderTwilioCredentials struct {
	CredentialType      config.TwilioCredentialType `json:"credentialType,omitempty"`
	AccountSID          string                      `json:"accountSID,omitempty"`
	AuthToken           *string                     `json:"authToken,omitempty"`
	APIKeySID           string                      `json:"apiKeySID,omitempty"`
	APIKeySecret        *string                     `json:"apiKeySecret,omitempty"`
	MessagingServiceSID string                      `json:"messagingServiceSID,omitempty"`
	From                string                      `json:"from,omitempty"`
}

type SMSProviderCustomSMSProviderConfigs struct {
	URL     string `json:"url,omitempty"`
	Timeout *int   `json:"timeout,omitempty"`
}

type SMSProviderAliyunCredentials struct {
	AccessKeyID          string            `json:"accessKeyID,omitempty"`
	AccessKeySecret      *string           `json:"accessKeySecret,omitempty"`
	SignName             string            `json:"signName,omitempty"`
	TemplateCode         string            `json:"templateCode,omitempty"`
	TemplateCodes        map[string]string `json:"templateCodes,omitempty"`
	OverseasTemplateCode string            `json:"overseasTemplateCode,omitempty"`
}

type SMSProviderTencentCredentials struct {
	SecretID      string            `json:"secretID,omitempty"`
	SecretKey     *string           `json:"secretKey,omitempty"`
	SDKAppID      string            `json:"sdkAppID,omitempty"`
	Region        string            `json:"region,omitempty"`
	SignName      string            `json:"signName,omitempty"`
	TemplateCode  string            `json:"templateCode,omitempty"`
	TemplateCodes map[string]string `json:"templateCodes,omitempty"`
}

type SMSProviderAliyunMASCredentials struct {
	AccessKeyID     string            `json:"accessKeyID,omitempty"`
	AccessKeySecret *string           `json:"accessKeySecret,omitempty"`
	SignName        string            `json:"signName,omitempty"`
	TemplateCode    string            `json:"templateCode,omitempty"`
	TemplateCodes   map[string]string `json:"templateCodes,omitempty"`
}

type SMSProviderYunpianCredentials struct {
	APIKey *string `json:"apiKey,omitempty"`
}

type SMSProviderSmsbaoCredentials struct {
	Username         string  `json:"username,omitempty"`
	PasswordOrAPIKey *string `json:"passwordOrAPIKey,omitempty"`
	GoodsID          string  `json:"goodsID,omitempty"`
}

type SMSProviderGatewayAPICredentials struct {
	Endpoint string  `json:"endpoint,omitempty"`
	APIToken *string `json:"apiToken,omitempty"`
	Sender   string  `json:"sender,omitempty"`
}

type SMSProviderSmsAeroCredentials struct {
	Email      string  `json:"email,omitempty"`
	APIKey     *string `json:"apiKey,omitempty"`
	SenderName string  `json:"senderName,omitempty"`
}

type SMSProviderSecrets struct {
	TwilioCredentials            *SMSProviderTwilioCredentials        `json:"twilioCredentials,omitempty"`
	AliyunCredentials            *SMSProviderAliyunCredentials        `json:"aliyunCredentials,omitempty"`
	AliyunMASCredentials         *SMSProviderAliyunMASCredentials     `json:"aliyunMASCredentials,omitempty"`
	TencentCredentials           *SMSProviderTencentCredentials       `json:"tencentCredentials,omitempty"`
	YunpianCredentials           *SMSProviderYunpianCredentials       `json:"yunpianCredentials,omitempty"`
	SmsbaoCredentials            *SMSProviderSmsbaoCredentials        `json:"smsbaoCredentials,omitempty"`
	GatewayAPICredentials        *SMSProviderGatewayAPICredentials    `json:"gatewayAPICredentials,omitempty"`
	SmsAeroCredentials           *SMSProviderSmsAeroCredentials       `json:"smsAeroCredentials,omitempty"`
	CustomSMSProviderCredentials *SMSProviderCustomSMSProviderConfigs `json:"customSMSProviderCredentials,omitempty"`
}

type SecretConfig struct {
	OAuthSSOProviderClientSecrets []OAuthSSOProviderClientSecret `json:"oauthSSOProviderClientSecrets,omitempty"`
	WebhookSecret                 *WebhookSecret                 `json:"webhookSecret,omitempty"`
	AdminAPISecrets               []AdminAPISecret               `json:"adminAPISecrets,omitempty"`
	SMTPSecret                    *SMTPSecret                    `json:"smtpSecret,omitempty"`
	OAuthClientSecrets            []OAuthClientSecret            `json:"oauthClientSecrets,omitempty"`
	BotProtectionProviderSecret   *BotProtectionProviderSecret   `json:"botProtectionProviderSecret,omitempty"`
	SAMLIdpSigningSecrets         *SAMLIdpSigningSecrets         `json:"samlIdpSigningSecrets,omitempty"`
	SAMLSpSigningSecrets          []SAMLSpSigningSecrets         `json:"samlSpSigningSecrets,omitempty"`
	SMSProviderSecrets            *SMSProviderSecrets            `json:"smsProviderSecrets,omitempty"`
}

type EffectiveSecretConfig struct {
	OAuthSSOProviderDemoSecrets []OAuthSSOProviderDemoSecretItem `json:"oauthSSOProviderDemoSecrets"`
}

type OAuthSSOProviderDemoSecretItem struct {
	Type string `json:"type"`
}

//nolint:gocognit
func NewSecretConfig(secretConfig *config.SecretConfig, unmaskedSecrets []config.SecretKey, now time.Time) (*SecretConfig, error) {
	out := &SecretConfig{}
	var unmaskedSecretsSet map[config.SecretKey]any = map[config.SecretKey]any{}
	for _, s := range unmaskedSecrets {
		unmaskedSecretsSet[s] = s
	}

	if oauthSSOProviderCredentials, ok := secretConfig.LookupData(config.OAuthSSOProviderCredentialsKey).(*config.OAuthSSOProviderCredentials); ok {
		for _, item := range oauthSSOProviderCredentials.Items {
			var clientSecret *string = nil
			if _, exist := unmaskedSecretsSet[config.OAuthSSOProviderCredentialsKey]; exist {
				s := item.ClientSecret
				clientSecret = &s
			}
			out.OAuthSSOProviderClientSecrets = append(out.OAuthSSOProviderClientSecrets, OAuthSSOProviderClientSecret{
				Alias:        item.Alias,
				ClientSecret: clientSecret,
			})
		}
	}

	if webhook, ok := secretConfig.LookupData(config.WebhookKeyMaterialsKey).(*config.WebhookKeyMaterials); ok {
		if webhook.Set.Len() == 1 {
			if jwkKey, ok := webhook.Set.Key(0); ok {
				if sKey, ok := jwkKey.(jwk.SymmetricKey); ok {
					var secret *string
					if _, exist := unmaskedSecretsSet[config.WebhookKeyMaterialsKey]; exist {
						octets := sKey.Octets()
						octetsStr := string(octets)
						secret = &octetsStr
					}
					out.WebhookSecret = &WebhookSecret{
						Secret: secret,
					}
				}
			}
		}
	}

	if adminAPI, ok := secretConfig.LookupData(config.AdminAPIAuthKeyKey).(*config.AdminAPIAuthKey); ok {
		for i := 0; i < adminAPI.Set.Len(); i++ {
			if jwkKey, ok := adminAPI.Set.Key(i); ok {
				var createdAt *time.Time
				if anyCreatedAt, ok := jwkKey.Get("created_at"); ok {
					if fCreatedAt, ok := anyCreatedAt.(float64); ok {
						t := time.Unix(int64(fCreatedAt), 0).UTC()
						createdAt = &t
					}
				}
				set := jwk.NewSet()
				_ = set.AddKey(jwkKey)
				publicKeyPEMBytes, err := jwkutil.PublicPEM(set)
				if err != nil {
					return nil, err
				}

				var privateKeyPEM *string
				if _, exist := unmaskedSecretsSet[config.AdminAPIAuthKeyKey]; exist {
					privateKeyPEMBytes, err := jwkutil.PrivatePublicPEM(set)
					if err != nil {
						return nil, err
					}
					privateKeyPEMStr := string(privateKeyPEMBytes)
					privateKeyPEM = &privateKeyPEMStr
				}

				out.AdminAPISecrets = append(out.AdminAPISecrets, AdminAPISecret{
					KeyID:         jwkKey.KeyID(),
					CreatedAt:     createdAt,
					PublicKeyPEM:  string(publicKeyPEMBytes),
					PrivateKeyPEM: privateKeyPEM,
				})
			}
		}
	}

	if smtp, ok := secretConfig.LookupData(config.SMTPServerCredentialsKey).(*config.SMTPServerCredentials); ok {
		smtpSecret := &SMTPSecret{
			Host:     smtp.Host,
			Port:     smtp.Port,
			Username: smtp.Username,
			Sender:   smtp.Sender,
		}
		if _, exist := unmaskedSecretsSet[config.SMTPServerCredentialsKey]; exist {
			smtpSecret.Password = &smtp.Password
		}
		out.SMTPSecret = smtpSecret
	}

	if oauthClientSecrets, ok := secretConfig.LookupData(config.OAuthClientCredentialsKey).(*config.OAuthClientCredentials); ok {
		for _, item := range oauthClientSecrets.Items {

			keys := []OAuthClientSecretKey{}

			for i := 0; i < item.Set.Len(); i++ {
				if jwkKey, ok := item.Set.Key(i); ok {
					newlyCreated := false
					var createdAt *time.Time
					if anyCreatedAt, ok := jwkKey.Get("created_at"); ok {
						if fCreatedAt, ok := anyCreatedAt.(float64); ok {
							t := time.Unix(int64(fCreatedAt), 0).UTC()
							createdAt = &t
							elapsed := now.Sub(*createdAt)
							newlyCreated = elapsed < 5*time.Minute
						}
					}
					var bytes []byte
					err := jwkKey.Raw(&bytes)
					if err != nil {
						return nil, err
					}

					clientSecret := ""
					_, unmask := unmaskedSecretsSet[config.OAuthClientCredentialsKey]
					if unmask || newlyCreated {
						clientSecret = string(bytes)
					}

					keys = append(keys, OAuthClientSecretKey{
						KeyID:     jwkKey.KeyID(),
						Key:       clientSecret,
						CreatedAt: createdAt,
					})
				}
			}

			out.OAuthClientSecrets = append(out.OAuthClientSecrets, OAuthClientSecret{
				ClientID: item.ClientID,
				Keys:     keys,
			})
		}
	}

	if botProtectionProviderSecrets, ok := secretConfig.LookupData(config.BotProtectionProviderCredentialsKey).(*config.BotProtectionProviderCredentials); ok {
		bpSecret := &BotProtectionProviderSecret{
			Type: botProtectionProviderSecrets.Type,
		}

		if _, exist := unmaskedSecretsSet[config.BotProtectionProviderCredentialsKey]; exist {
			bpSecret.SecretKey = &botProtectionProviderSecrets.SecretKey
		}

		out.BotProtectionProviderSecret = bpSecret
	}

	if samlIdpSigningSecrets, ok := secretConfig.LookupData(config.SAMLIdpSigningMaterialsKey).(*config.SAMLIdpSigningMaterials); ok {
		out.SAMLIdpSigningSecrets = toPortalSAMLIdpSigningSecrets(samlIdpSigningSecrets)
	}

	if samlSpSigningSecrets, ok := secretConfig.LookupData(config.SAMLSpSigningMaterialsKey).(*config.SAMLSpSigningMaterials); ok {
		out.SAMLSpSigningSecrets = toPortalSAMLSpSigningSecrets(samlSpSigningSecrets)
	}

	smsProviderSecrets := &SMSProviderSecrets{}
	if twilioCredentials, ok := secretConfig.LookupData(config.TwilioCredentialsKey).(*config.TwilioCredentials); ok {
		smsProviderSecrets.TwilioCredentials = &SMSProviderTwilioCredentials{
			CredentialType:      twilioCredentials.GetCredentialType(),
			AccountSID:          twilioCredentials.AccountSID,
			APIKeySID:           twilioCredentials.APIKeySID,
			MessagingServiceSID: twilioCredentials.MessagingServiceSID,
			From:                twilioCredentials.From,
		}
		if _, exist := unmaskedSecretsSet[config.TwilioCredentialsKey]; exist {
			smsProviderSecrets.TwilioCredentials.AuthToken = &twilioCredentials.AuthToken
			smsProviderSecrets.TwilioCredentials.APIKeySecret = &twilioCredentials.APIKeySecret
		}

	}
	if aliyunCredentials, ok := secretConfig.LookupData(config.AliyunCredentialsKey).(*config.AliyunCredentials); ok {
		smsProviderSecrets.AliyunCredentials = &SMSProviderAliyunCredentials{
			AccessKeyID:          aliyunCredentials.AccessKeyID,
			SignName:             aliyunCredentials.SignName,
			TemplateCode:         aliyunCredentials.TemplateCode,
			TemplateCodes:        aliyunCredentials.TemplateCodes,
			OverseasTemplateCode: aliyunCredentials.OverseasTemplateCode,
		}
		if _, exist := unmaskedSecretsSet[config.AliyunCredentialsKey]; exist {
			smsProviderSecrets.AliyunCredentials.AccessKeySecret = &aliyunCredentials.AccessKeySecret
		}
	}
	if tencentCredentials, ok := secretConfig.LookupData(config.TencentCredentialsKey).(*config.TencentCredentials); ok {
		smsProviderSecrets.TencentCredentials = &SMSProviderTencentCredentials{
			SecretID:      tencentCredentials.SecretID,
			SDKAppID:      tencentCredentials.SDKAppID,
			Region:        tencentCredentials.Region,
			SignName:      tencentCredentials.SignName,
			TemplateCode:  tencentCredentials.TemplateCode,
			TemplateCodes: tencentCredentials.TemplateCodes,
		}
		if _, exist := unmaskedSecretsSet[config.TencentCredentialsKey]; exist {
			smsProviderSecrets.TencentCredentials.SecretKey = &tencentCredentials.SecretKey
		}
	}
	if aliyunMASCredentials, ok := secretConfig.LookupData(config.AliyunMASCredentialsKey).(*config.AliyunMASCredentials); ok {
		smsProviderSecrets.AliyunMASCredentials = &SMSProviderAliyunMASCredentials{
			AccessKeyID:   aliyunMASCredentials.AccessKeyID,
			SignName:      aliyunMASCredentials.SignName,
			TemplateCode:  aliyunMASCredentials.TemplateCode,
			TemplateCodes: aliyunMASCredentials.TemplateCodes,
		}
		if _, exist := unmaskedSecretsSet[config.AliyunMASCredentialsKey]; exist {
			smsProviderSecrets.AliyunMASCredentials.AccessKeySecret = &aliyunMASCredentials.AccessKeySecret
		}
	}
	if yunpianCredentials, ok := secretConfig.LookupData(config.YunpianCredentialsKey).(*config.YunpianCredentials); ok {
		smsProviderSecrets.YunpianCredentials = &SMSProviderYunpianCredentials{}
		if _, exist := unmaskedSecretsSet[config.YunpianCredentialsKey]; exist {
			smsProviderSecrets.YunpianCredentials.APIKey = &yunpianCredentials.APIKey
		}
	}
	if smsbaoCredentials, ok := secretConfig.LookupData(config.SmsbaoCredentialsKey).(*config.SmsbaoCredentials); ok {
		smsProviderSecrets.SmsbaoCredentials = &SMSProviderSmsbaoCredentials{
			Username: smsbaoCredentials.Username,
			GoodsID:  smsbaoCredentials.GoodsID,
		}
		if _, exist := unmaskedSecretsSet[config.SmsbaoCredentialsKey]; exist {
			smsProviderSecrets.SmsbaoCredentials.PasswordOrAPIKey = &smsbaoCredentials.PasswordOrAPIKey
		}
	}
	if gatewayAPICredentials, ok := secretConfig.LookupData(config.GatewayAPICredentialsKey).(*config.GatewayAPICredentials); ok {
		smsProviderSecrets.GatewayAPICredentials = &SMSProviderGatewayAPICredentials{
			Endpoint: gatewayAPICredentials.Endpoint,
			Sender:   gatewayAPICredentials.Sender,
		}
		if _, exist := unmaskedSecretsSet[config.GatewayAPICredentialsKey]; exist {
			smsProviderSecrets.GatewayAPICredentials.APIToken = &gatewayAPICredentials.APIToken
		}
	}
	if smsAeroCredentials, ok := secretConfig.LookupData(config.SmsAeroCredentialsKey).(*config.SmsAeroCredentials); ok {
		smsProviderSecrets.SmsAeroCredentials = &SMSProviderSmsAeroCredentials{
			Email:      smsAeroCredentials.Email,
			SenderName: smsAeroCredentials.SenderName,
		}
		if _, exist := unmaskedSecretsSet[config.SmsAeroCredentialsKey]; exist {
			smsProviderSecrets.SmsAeroCredentials.APIKey = &smsAeroCredentials.APIKey
		}
	}
	if customSMSProviderConfig, ok := secretConfig.LookupData(config.CustomSMSProviderConfigKey).(*config.CustomSMSProviderConfig); ok {
		smsProviderSecrets.CustomSMSProviderCredentials = &SMSProviderCustomSMSProviderConfigs{
			URL:     customSMSProviderConfig.URL,
			Timeout: (*int)(customSMSProviderConfig.Timeout),
		}
	}
	out.SMSProviderSecrets = smsProviderSecrets

	return out, nil
}

func toPortalSAMLIdpSigningSecrets(cfg *config.SAMLIdpSigningMaterials) *SAMLIdpSigningSecrets {
	result := &SAMLIdpSigningSecrets{
		Certificates: []SAMLIdpSigningCertificate{},
	}

	for _, cfgCert := range cfg.Certificates {
		cert := SAMLIdpSigningCertificate{
			KeyID:                  cfgCert.Key.KeyID(),
			CertificateFingerprint: cfgCert.Certificate.Fingerprint(),
			CertificatePEM:         string(cfgCert.Certificate.Pem),
		}
		result.Certificates = append(result.Certificates, cert)
	}

	return result
}

func toPortalSAMLSpSigningSecrets(cfg *config.SAMLSpSigningMaterials) []SAMLSpSigningSecrets {
	result := []SAMLSpSigningSecrets{}

	for _, cfgItem := range *cfg {
		item := SAMLSpSigningSecrets{
			ClientID:     cfgItem.ServiceProviderID,
			Certificates: []SAMLSpSigningCertificate{},
		}

		for _, cfgCert := range cfgItem.Certificates {
			item.Certificates = append(item.Certificates, SAMLSpSigningCertificate{
				CertificateFingerprint: cfgCert.Fingerprint(),
				CertificatePEM:         string(cfgCert.Pem),
			})
		}

		result = append(result, item)
	}

	return result
}
