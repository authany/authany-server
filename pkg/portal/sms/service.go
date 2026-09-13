package sms

import (
	"context"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/aliyun"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/aliyunmas"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/custom"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/gatewayapi"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsaero"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsbao"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/tencent"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/twilio"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/yunpian"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

type Service struct {
	DenoEndpoint config.DenoEndpoint
}

const TEST_OTP = "000000"
const TEST_APP_NAME = "Test"

// TestSMSTemplateName is the template used to resolve the template code of
// template code based providers when sending a test SMS.
const TestSMSTemplateName = "verification_sms.txt"

func makeTestSMSBody(appName string, code string) string {
	return fmt.Sprintf("[%s] Your one-time password is %s", appName, code)
}

func (s *Service) sendByTwilio(
	ctx context.Context,
	app *model.App,
	to string,
	cfg model.SMSProviderConfigurationTwilioInput,
) error {
	twilioClient := twilio.NewTwilioClient(&config.TwilioCredentials{
		CredentialType_WriteOnly: &cfg.CredentialType,
		AccountSID:               cfg.AccountSID,
		AuthToken:                cfg.AuthToken,
		APIKeySID:                cfg.APIKeySID,
		APIKeySecret:             cfg.APIKeySecret,
		MessagingServiceSID:      cfg.MessagingServiceSID,
		From:                     cfg.From,
	})

	translationService := NewTranslationService(app)
	sender, err := translationService.GetSenderForTestSMS(ctx)
	if err != nil {
		return err
	}

	return twilioClient.Send(ctx, smsapi.SendOptions{
		Sender: sender,
		To:     to,
		Body:   makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
}

func (s *Service) sendByAliyun(
	ctx context.Context,
	to string,
	cfg model.SMSProviderConfigurationAliyunInput,
) error {
	aliyunClient := aliyun.NewAliyunClient(&config.AliyunCredentials{
		AccessKeyID:     cfg.AccessKeyID,
		AccessKeySecret: cfg.AccessKeySecret,
		SMSTemplateCodeConfig: config.SMSTemplateCodeConfig{
			SignName:      cfg.SignName,
			TemplateCode:  cfg.TemplateCode,
			TemplateCodes: cfg.TemplateCodes,
		},
		OverseasTemplateCode: cfg.OverseasTemplateCode,
	})

	return aliyunClient.Send(ctx, smsapi.SendOptions{
		To:           to,
		Body:         makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateName: TestSMSTemplateName,
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
}

func (s *Service) sendByTencent(
	ctx context.Context,
	to string,
	cfg model.SMSProviderConfigurationTencentInput,
) error {
	credentials := &config.TencentCredentials{
		SecretID:  cfg.SecretID,
		SecretKey: cfg.SecretKey,
		SDKAppID:  cfg.SDKAppID,
		Region:    cfg.Region,
		SMSTemplateCodeConfig: config.SMSTemplateCodeConfig{
			SignName:      cfg.SignName,
			TemplateCode:  cfg.TemplateCode,
			TemplateCodes: cfg.TemplateCodes,
		},
	}
	credentials.SetDefaults()
	tencentClient := tencent.NewTencentClient(credentials)

	return tencentClient.Send(ctx, smsapi.SendOptions{
		To:           to,
		Body:         makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateName: TestSMSTemplateName,
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
}

func (s *Service) sendByAliyunMAS(
	ctx context.Context,
	to string,
	cfg model.SMSProviderConfigurationAliyunMASInput,
) error {
	aliyunMASClient := aliyunmas.NewAliyunMASClient(&config.AliyunMASCredentials{
		AccessKeyID:     cfg.AccessKeyID,
		AccessKeySecret: cfg.AccessKeySecret,
		SMSTemplateCodeConfig: config.SMSTemplateCodeConfig{
			SignName:      cfg.SignName,
			TemplateCode:  cfg.TemplateCode,
			TemplateCodes: cfg.TemplateCodes,
		},
	})

	return aliyunMASClient.Send(ctx, smsapi.SendOptions{
		To:           to,
		Body:         makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateName: TestSMSTemplateName,
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
}

func (s *Service) sendByYunpian(
	ctx context.Context,
	to string,
	cfg model.SMSProviderConfigurationYunpianInput,
) error {
	yunpianClient := yunpian.NewYunpianClient(&config.YunpianCredentials{
		APIKey: cfg.APIKey,
	})

	return yunpianClient.Send(ctx, smsapi.SendOptions{
		To:   to,
		Body: makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
}

func (s *Service) sendBySmsbao(
	ctx context.Context,
	to string,
	cfg model.SMSProviderConfigurationSmsbaoInput,
) error {
	smsbaoClient := smsbao.NewSmsbaoClient(&config.SmsbaoCredentials{
		Username:         cfg.Username,
		PasswordOrAPIKey: cfg.PasswordOrAPIKey,
		GoodsID:          cfg.GoodsID,
	})

	return smsbaoClient.Send(ctx, smsapi.SendOptions{
		To:   to,
		Body: makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
}

func (s *Service) sendByGatewayAPI(
	ctx context.Context,
	to string,
	cfg model.SMSProviderConfigurationGatewayAPIInput,
) error {
	credentials := &config.GatewayAPICredentials{
		Endpoint: cfg.Endpoint,
		APIToken: cfg.APIToken,
		Sender:   cfg.Sender,
	}
	// The secret schema restricts the endpoint to the official base URLs,
	// but this input does not go through the schema.
	if err := credentials.ValidateEndpoint(); err != nil {
		return apierrors.NewInvalid(err.Error())
	}
	gatewayAPIClient := gatewayapi.NewGatewayAPIClient(credentials)

	return gatewayAPIClient.Send(ctx, smsapi.SendOptions{
		To:   to,
		Body: makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
}

func (s *Service) sendBySmsAero(
	ctx context.Context,
	to string,
	cfg model.SMSProviderConfigurationSmsAeroInput,
) error {
	smsAeroClient := smsaero.NewSmsAeroClient(&config.SmsAeroCredentials{
		Email:      cfg.Email,
		APIKey:     cfg.APIKey,
		SenderName: cfg.SenderName,
	})

	return smsAeroClient.Send(ctx, smsapi.SendOptions{
		To:   to,
		Body: makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
}

func (s *Service) sendByWebhook(
	ctx context.Context,
	secret *config.WebhookKeyMaterials,
	to string,
	cfg model.SMSProviderConfigurationWebhookInput,
) error {
	webHookImpl := &hook.WebHookImpl{
		Secret: secret,
	}
	webhook := custom.NewSMSWebHook(webHookImpl, &config.CustomSMSProviderConfig{
		URL:     cfg.URL,
		Timeout: (*config.DurationSeconds)(cfg.Timeout),
	})

	url, err := url.Parse(cfg.URL)
	if err != nil {
		return err
	}

	err = webhook.Call(ctx, url, custom.SendOptions{
		To:   to,
		Body: makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) sendByDeno(
	ctx context.Context,
	app *model.App,
	to string,
	cfg model.SMSProviderConfigurationDenoInput,
) error {

	deno := custom.NewSMSDenoHookForTest(s.DenoEndpoint, &config.CustomSMSProviderConfig{
		// URL is not important here, we execute the script with a string
		URL:     "",
		Timeout: (*config.DurationSeconds)(cfg.Timeout),
	})

	err := deno.Test(ctx, cfg.Script, custom.SendOptions{
		To:   to,
		Body: makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) SendTestSMS(
	ctx context.Context,
	app *model.App,
	to string,
	webhookSecretLoader func(ctx context.Context) (*config.WebhookKeyMaterials, error),
	input model.SMSProviderConfigurationInput) error {
	if input.Twilio != nil {
		return s.sendByTwilio(ctx, app, to, *input.Twilio)

	} else if input.Aliyun != nil {
		return s.sendByAliyun(ctx, to, *input.Aliyun)

	} else if input.AliyunMAS != nil {
		return s.sendByAliyunMAS(ctx, to, *input.AliyunMAS)

	} else if input.Tencent != nil {
		return s.sendByTencent(ctx, to, *input.Tencent)

	} else if input.Yunpian != nil {
		return s.sendByYunpian(ctx, to, *input.Yunpian)

	} else if input.Smsbao != nil {
		return s.sendBySmsbao(ctx, to, *input.Smsbao)

	} else if input.GatewayAPI != nil {
		return s.sendByGatewayAPI(ctx, to, *input.GatewayAPI)

	} else if input.SmsAero != nil {
		return s.sendBySmsAero(ctx, to, *input.SmsAero)

	} else if input.Webhook != nil {
		webhookSecret, err := webhookSecretLoader(ctx)
		if err != nil {
			return err
		}
		return s.sendByWebhook(ctx, webhookSecret, to, *input.Webhook)

	} else if input.Deno != nil {
		return s.sendByDeno(ctx, app, to, *input.Deno)
	}
	return apierrors.NewInvalid("no provider config given")
}
