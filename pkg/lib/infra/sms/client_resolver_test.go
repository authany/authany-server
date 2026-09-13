package sms

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/kelseyhightower/envconfig"
	. "github.com/smartystreets/goconvey/convey"
	goyaml "go.yaml.in/yaml/v2"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

// nolint: gocognit
func TestClientResolver(t *testing.T) {
	Convey("resolve client", t, func() {
		f, err := os.Open("testdata/client_resolver_tests.yaml")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		type AuthgearYAML struct {
			Messaging *config.MessagingConfig `json:"messaging"`
		}

		type AuthgesrSecretsYAML struct {
			Nexmo      *config.NexmoCredentials        `json:"nexmo"`
			Twilio     *config.TwilioCredentials       `json:"twilio"`
			Custom     *config.CustomSMSProviderConfig `json:"custom"`
			Aliyun     *config.AliyunCredentials       `json:"aliyun"`
			AliyunMAS  *config.AliyunMASCredentials    `json:"aliyun_mas"`
			Tencent    *config.TencentCredentials      `json:"tencent"`
			Yunpian    *config.YunpianCredentials      `json:"yunpian"`
			Smsbao     *config.SmsbaoCredentials       `json:"smsbao"`
			GatewayAPI *config.GatewayAPICredentials   `json:"gatewayapi"`
			SmsAero    *config.SmsAeroCredentials      `json:"smsaero"`
		}

		type EnvConfig struct {
			SMSGateway config.SMSGatewayEnvironmentConfig `envconfig:"SMS_GATEWAY"`
		}

		type TestCase struct {
			Name                 string  `yaml:"name"`
			AuthgearYAML         any     `yaml:"authgear.yaml"`
			AuthgearSecretsYAML  any     `yaml:"authgear.secrets.yaml"`
			EnvironmentVariables *string `yaml:"environment_variables"`
			Result               any     `yaml:"result"`
			Error                string  `yaml:"error"`
		}

		decoder := goyaml.NewDecoder(f)
		for {
			var testCase TestCase
			err := decoder.Decode(&testCase)
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				panic(err)
			}

			Convey(testCase.Name, func() {
				authgearYAMLData, err := goyaml.Marshal(testCase.AuthgearYAML)
				if err != nil {
					panic(err)
				}
				authgearYAMLData, err = yaml.YAMLToJSON(authgearYAMLData)
				if err != nil {
					panic(err)
				}
				var authgearYAML *AuthgearYAML
				err = json.Unmarshal(authgearYAMLData, &authgearYAML)
				if err != nil {
					panic(err)
				}

				authgearSecretsYAMLData, err := goyaml.Marshal(testCase.AuthgearSecretsYAML)
				if err != nil {
					panic(err)
				}
				authgearSecretsYAMLData, err = yaml.YAMLToJSON(authgearSecretsYAMLData)
				if err != nil {
					panic(err)
				}
				var authgearSecretsYAML *AuthgesrSecretsYAML
				err = json.Unmarshal(authgearSecretsYAMLData, &authgearSecretsYAML)
				if err != nil {
					panic(err)
				}

				var messagingConfig *config.MessagingConfig
				if authgearYAML != nil {
					messagingConfig = authgearYAML.Messaging
				}
				var authgearSecretsYAMLNexmo *config.NexmoCredentials
				var authgearSecretsYAMLTwilio *config.TwilioCredentials
				var authgearSecretsYAMLCustom *config.CustomSMSProviderConfig
				var authgearSecretsYAMLAliyun *config.AliyunCredentials
				var authgearSecretsYAMLAliyunMAS *config.AliyunMASCredentials
				var authgearSecretsYAMLTencent *config.TencentCredentials
				var authgearSecretsYAMLYunpian *config.YunpianCredentials
				var authgearSecretsYAMLSmsbao *config.SmsbaoCredentials
				var authgearSecretsYAMLGatewayAPI *config.GatewayAPICredentials
				var authgearSecretsYAMLSmsAero *config.SmsAeroCredentials
				if authgearSecretsYAML != nil {
					authgearSecretsYAMLNexmo = authgearSecretsYAML.Nexmo
					authgearSecretsYAMLTwilio = authgearSecretsYAML.Twilio
					authgearSecretsYAMLCustom = authgearSecretsYAML.Custom
					authgearSecretsYAMLAliyun = authgearSecretsYAML.Aliyun
					authgearSecretsYAMLAliyunMAS = authgearSecretsYAML.AliyunMAS
					authgearSecretsYAMLTencent = authgearSecretsYAML.Tencent
					authgearSecretsYAMLYunpian = authgearSecretsYAML.Yunpian
					authgearSecretsYAMLSmsbao = authgearSecretsYAML.Smsbao
					authgearSecretsYAMLGatewayAPI = authgearSecretsYAML.GatewayAPI
					authgearSecretsYAMLSmsAero = authgearSecretsYAML.SmsAero
				}

				var smsGatewayEnvironmentConfig config.SMSGatewayEnvironmentConfig
				if testCase.EnvironmentVariables != nil {
					for ln := range strings.SplitSeq(*testCase.EnvironmentVariables, "\n") {
						var keyval = strings.Split(ln, "=")
						if len(keyval) < 2 {
							continue
						}
						t.Setenv(keyval[0], keyval[1])
					}

					cfg := &EnvConfig{}
					err = envconfig.Process("", cfg)
					if err != nil {
						panic(err)
					}

					smsGatewayEnvironmentConfig = cfg.SMSGateway
				}

				resultData, err := goyaml.Marshal(testCase.Result)
				if err != nil {
					panic(err)
				}
				resultData, err = yaml.YAMLToJSON(resultData)
				if err != nil {
					panic(err)
				}
				var result any
				err = json.Unmarshal(resultData, &result)
				if err != nil {
					panic(err)
				}

				var authgearYAMLSMSProvider config.SMSProvider
				var authgearYAMLSMSGateway *config.SMSGatewayConfig
				if messagingConfig != nil {
					authgearYAMLSMSProvider = messagingConfig.Deprecated_SMSProvider
					authgearYAMLSMSGateway = messagingConfig.SMSGateway
				}

				var environmentDefaultProvider config.SMSGatewayEnvironmentDefaultProvider
				var environmentDefaultUseConfigFrom config.SMSGatewayEnvironmentDefaultUseConfigFrom
				var environmentNexmoCredentials config.SMSGatewayEnvironmentNexmoCredentials
				var environmentTwilioCredentials config.SMSGatewayEnvironmentTwilioCredentials
				var environmentCustomSMSProviderConfig config.SMSGatewayEnvironmentCustomSMSProviderConfig
				environmentDefaultProvider = smsGatewayEnvironmentConfig.Default.Provider
				environmentDefaultUseConfigFrom = smsGatewayEnvironmentConfig.Default.UseConfigFrom
				environmentNexmoCredentials = smsGatewayEnvironmentConfig.Nexmo
				environmentTwilioCredentials = smsGatewayEnvironmentConfig.Twilio
				environmentCustomSMSProviderConfig = smsGatewayEnvironmentConfig.Custom

				clientResolver := ClientResolver{
					AuthgearYAMLSMSProvider:                    authgearYAMLSMSProvider,
					AuthgearYAMLSMSGateway:                     authgearYAMLSMSGateway,
					AuthgearSecretsYAMLNexmoCredentials:        authgearSecretsYAMLNexmo,
					AuthgearSecretsYAMLTwilioCredentials:       authgearSecretsYAMLTwilio,
					AuthgearSecretsYAMLCustomSMSProviderConfig: authgearSecretsYAMLCustom,
					AuthgearSecretsYAMLAliyunCredentials:       authgearSecretsYAMLAliyun,
					AuthgearSecretsYAMLAliyunMASCredentials:    authgearSecretsYAMLAliyunMAS,
					AuthgearSecretsYAMLTencentCredentials:      authgearSecretsYAMLTencent,
					AuthgearSecretsYAMLYunpianCredentials:      authgearSecretsYAMLYunpian,
					AuthgearSecretsYAMLSmsbaoCredentials:       authgearSecretsYAMLSmsbao,
					AuthgearSecretsYAMLGatewayAPICredentials:   authgearSecretsYAMLGatewayAPI,
					AuthgearSecretsYAMLSmsAeroCredentials:      authgearSecretsYAMLSmsAero,
					EnvironmentDefaultProvider:                 environmentDefaultProvider,
					EnvironmentDefaultUseConfigFrom:            environmentDefaultUseConfigFrom,
					EnvironmentNexmoCredentials:                environmentNexmoCredentials,
					EnvironmentTwilioCredentials:               environmentTwilioCredentials,
					EnvironmentCustomSMSProviderConfig:         environmentCustomSMSProviderConfig,
				}
				_, cred, err := clientResolver.ResolveClient()
				if testCase.Error != "" {
					So(err.Error(), ShouldEqual, testCase.Error)
				} else {
					So(err, ShouldBeNil)
				}
				if result != nil {
					So(toMap(cred), ShouldEqual, result)
				} else {
					So(cred, ShouldBeNil)
				}

			})
		}
	})
}

func toMap(c SMSClientCredentials) map[string]any {
	switch v := c.(type) {
	case *TwilioClientCredentials:
		return map[string]any{
			"account_sid":         v.AccountSID,
			"auth_token":          v.AuthToken,
			"message_service_sid": v.MessagingServiceSID,
		}
	case *NexmoClientCredentials:
		return map[string]any{
			"api_key":    v.APIKey,
			"api_secret": v.APISecret,
		}
	case *CustomClientCredentials:
		return map[string]any{
			"url":     v.URL,
			"timeout": float64(*v.Timeout),
		}
	case *AliyunClientCredentials:
		return map[string]any{
			"access_key_id":     v.AccessKeyID,
			"access_key_secret": v.AccessKeySecret,
			"sign_name":         v.SignName,
			"template_code":     v.TemplateCode,
		}
	case *TencentClientCredentials:
		return map[string]any{
			"secret_id":     v.SecretID,
			"secret_key":    v.SecretKey,
			"sdk_app_id":    v.SDKAppID,
			"region":        v.Region,
			"sign_name":     v.SignName,
			"template_code": v.TemplateCode,
		}
	case *AliyunMASClientCredentials:
		return map[string]any{
			"access_key_id":     v.AccessKeyID,
			"access_key_secret": v.AccessKeySecret,
			"sign_name":         v.SignName,
			"template_code":     v.TemplateCode,
		}
	case *YunpianClientCredentials:
		return map[string]any{
			"apikey": v.APIKey,
		}
	case *SmsbaoClientCredentials:
		return map[string]any{
			"username":            v.Username,
			"password_or_api_key": v.PasswordOrAPIKey,
			"goods_id":            v.GoodsID,
		}
	case *GatewayAPIClientCredentials:
		return map[string]any{
			"endpoint":  v.Endpoint,
			"api_token": v.APIToken,
			"sender":    v.Sender,
		}
	case *SmsAeroClientCredentials:
		return map[string]any{
			"email":       v.Email,
			"api_key":     v.APIKey,
			"sender_name": v.SenderName,
		}
	}
	return nil
}
