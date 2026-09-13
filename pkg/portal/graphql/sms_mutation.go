package graphql

import (
	"context"
	"encoding/json"

	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

var sendTestSMSInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SendTestSMSInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "App ID to test.",
		},
		"to": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The recipient phone number.",
		},
		"providerConfiguration": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(smsProviderConfigurationInput),
			Description: "The SMS provider configuration.",
		},
	},
})

var smsProviderConfigurationInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"twilio": &graphql.InputObjectFieldConfig{
			Type:        smsProviderConfigurationTwilioInput,
			Description: "Twilio configuration",
		},
		"aliyun": &graphql.InputObjectFieldConfig{
			Type:        smsProviderConfigurationAliyunInput,
			Description: "Aliyun configuration",
		},
		"aliyunMAS": &graphql.InputObjectFieldConfig{
			Type:        smsProviderConfigurationAliyunMASInput,
			Description: "Aliyun MAS configuration",
		},
		"tencent": &graphql.InputObjectFieldConfig{
			Type:        smsProviderConfigurationTencentInput,
			Description: "Tencent Cloud configuration",
		},
		"yunpian": &graphql.InputObjectFieldConfig{
			Type:        smsProviderConfigurationYunpianInput,
			Description: "Yunpian configuration",
		},
		"smsbao": &graphql.InputObjectFieldConfig{
			Type:        smsProviderConfigurationSmsbaoInput,
			Description: "SMSBao configuration",
		},
		"gatewayAPI": &graphql.InputObjectFieldConfig{
			Type:        smsProviderConfigurationGatewayAPIInput,
			Description: "GatewayAPI configuration",
		},
		"smsAero": &graphql.InputObjectFieldConfig{
			Type:        smsProviderConfigurationSmsAeroInput,
			Description: "SMS Aero configuration",
		},
		"webhook": &graphql.InputObjectFieldConfig{
			Type:        smsProviderConfigurationWebhookInput,
			Description: "Webhook Configuration",
		},
		"deno": &graphql.InputObjectFieldConfig{
			Type:        smsProviderConfigurationDenoInput,
			Description: "Deno hook configuration",
		},
	},
})

var smsProviderConfigurationTwilioInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationTwilioInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"credentialType": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(twilioCredentialType),
		},
		"accountSID": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"authToken": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"apiKeySID": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"apiKeySecret": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"messagingServiceSID": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"from": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

var smsProviderConfigurationAliyunInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationAliyunInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"accessKeyID": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"accessKeySecret": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"signName": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"templateCode": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"templateCodes": &graphql.InputObjectFieldConfig{
			Type: SMSTemplateCodes,
		},
		"overseasTemplateCode": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

var smsProviderConfigurationTencentInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationTencentInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"secretID": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"secretKey": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"sdkAppID": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"region": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"signName": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"templateCode": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"templateCodes": &graphql.InputObjectFieldConfig{
			Type: SMSTemplateCodes,
		},
	},
})

var smsProviderConfigurationAliyunMASInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationAliyunMASInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"accessKeyID": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"accessKeySecret": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"signName": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"templateCode": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"templateCodes": &graphql.InputObjectFieldConfig{
			Type: SMSTemplateCodes,
		},
	},
})

var smsProviderConfigurationYunpianInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationYunpianInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"apiKey": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var smsProviderConfigurationSmsbaoInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationSmsbaoInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"username": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"passwordOrAPIKey": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"goodsID": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

var smsProviderConfigurationGatewayAPIInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationGatewayAPIInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"endpoint": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"apiToken": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"sender": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

var smsProviderConfigurationSmsAeroInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationSmsAeroInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"email": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"apiKey": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"senderName": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

var smsProviderConfigurationWebhookInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationWebhookInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"url": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"timeout": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var smsProviderConfigurationDenoInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationDenoInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"script": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"timeout": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var _ = registerMutationField(
	"sendTestSMSConfiguration",
	&graphql.Field{
		Description: "Send a SMS to test the configuration",
		Type:        graphql.Boolean,
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(sendTestSMSInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			input := p.Args["input"].(map[string]any)
			appNodeID := input["appID"].(string)
			to := input["to"].(string)
			providerConfigurationInput := input["providerConfiguration"]

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			// Access control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(ctx, appID)
			if err != nil {
				return nil, err
			}

			app, err := gqlCtx.AppService.Get(ctx, appID)
			if err != nil {
				return nil, err
			}

			providerConfigJSON, err := json.Marshal(providerConfigurationInput)
			if err != nil {
				return nil, err
			}
			var providerConfig model.SMSProviderConfigurationInput
			err = json.Unmarshal(providerConfigJSON, &providerConfig)
			if err != nil {
				return nil, err
			}

			webhookSecretLoader := func(ctx context.Context) (*config.WebhookKeyMaterials, error) {
				return gqlCtx.AppService.LoadAppWebhookSecretMaterials(ctx, app)
			}

			err = gqlCtx.SMSService.SendTestSMS(
				ctx,
				app,
				to,
				webhookSecretLoader,
				providerConfig,
			)
			if err != nil {
				return nil, err
			}

			return nil, nil
		},
	},
)
