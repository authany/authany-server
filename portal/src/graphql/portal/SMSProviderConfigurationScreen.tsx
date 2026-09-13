import cn from "classnames";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import authgear from "@authgear/web";
import {
  AppSecretKey,
  SmsProviderConfigurationInput,
  SmsProviderConfigurationTwilioInput,
  TwilioCredentialType,
} from "./globalTypes.generated";
import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import FormContainer from "../../FormContainer";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
  SMSProvider,
  SMSProviderAliyunCredentials,
  SMSProviderAliyunMASCredentials,
  SMSProviderGatewayAPICredentials,
  SMSProviderSmsAeroCredentials,
  SMSProviderSmsbaoCredentials,
  SMSProviderTencentCredentials,
  SMSProviderTwilioCredentials,
  SMSProviderYunpianCredentials,
  getHookKind,
} from "../../types";
import { produce } from "immer";
import { FormattedMessage, Context as MessageContext } from "../../intl";
import ScreenContent from "../../ScreenContent";
import styles from "./SMSProviderConfigurationScreen.module.css";
import logoTwilio from "../../images/twilio_logo.svg";
import logoAliyun from "../../images/aliyun_logo.svg";
import logoAliyunMAS from "../../images/aliyun_mas_logo.svg";
import logoTencentCloud from "../../images/tencent_cloud_logo.svg";
import logoYunpian from "../../images/yunpian_logo.svg";
import logoSMSBao from "../../images/smsbao_logo.svg";
import logoGatewayAPI from "../../images/gatewayapi_logo.svg";
import logoSMSAero from "../../images/smsaero_logo.svg";
import logoWebhook from "../../images/webhook_logo.svg";
import logoAuthany from "../../images/authany_logo.svg";
import { startReauthentication } from "./Authenticated";
import { CodeField } from "../../components/common/CodeField";
import CodeEditor from "../../CodeEditor";
import { useResourceForm } from "../../hook/useResourceForm";
import {
  Resource,
  ResourceSpecifier,
  ResourcesDiffResult,
  getDenoScriptPathFromURL,
  makeDenoScriptSpecifier,
} from "../../util/resource";
import { DENO_TYPES_URL } from "../../util/deno";
import { genRandomHexadecimalString } from "../../util/random";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import { useSendTestSMSMutation } from "./mutations/sendTestSMS";
import { useCheckDenoHookMutation } from "./mutations/checkDenoHook";
import { FeatureDisabledCallout } from "../../components/v2/FeatureDisabledCallout/FeatureDisabledCallout";
import { ErrorParseRule, makeLocalErrorParseRule } from "../../error/parse";
import { APIError, LocalError } from "../../error/error";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { TestSMSDialog } from "../../components/sms-provider/TestSMSDialog";
import {
  AliyunForm,
  AliyunFormState,
} from "../../components/sms-provider/AliyunForm";
import {
  AliyunMASForm,
  AliyunMASFormState,
} from "../../components/sms-provider/AliyunMASForm";
import {
  TencentForm,
  TencentFormState,
} from "../../components/sms-provider/TencentForm";
import {
  YunpianForm,
  YunpianFormState,
} from "../../components/sms-provider/YunpianForm";
import {
  SMSBaoForm,
  SMSBaoFormState,
} from "../../components/sms-provider/SMSBaoForm";
import {
  DEFAULT_GATEWAY_API_ENDPOINT,
  GatewayAPIForm,
  GatewayAPIFormState,
} from "../../components/sms-provider/GatewayAPIForm";
import {
  SMSAeroForm,
  SMSAeroFormState,
} from "../../components/sms-provider/SMSAeroForm";
import {
  SMSTemplateCodes,
  localErrorSignNameRequired,
  localErrorTemplateCodeRequired,
  parseSMSTemplateCodes,
  serializeSMSTemplateCodes,
} from "../../components/sms-provider/TemplateCodeFields";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { RedMessageBar_RemindConfigureSMSProviderInSMSProviderScreen } from "../../RedMessageBar";
import ExternalLink from "../../ExternalLink";
import {
  IconRadioCards,
  IconRadioCardOption,
} from "../../components/v2/IconRadioCards/IconRadioCards";
import { TextField } from "../../components/v2/TextField/TextField";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import {
  Flex,
  Heading,
  IconButton as RadixIconButton,
  RadioGroup,
  Text,
  Tooltip as RadixTooltip,
} from "@radix-ui/themes";
import { CodeIcon, EyeOpenIcon, InfoCircledIcon } from "@radix-ui/react-icons";
import { FormField } from "../../components/v2/FormField/FormField";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { useFormContainerBaseContext } from "../../FormContainerBase";

const SECRETS = [AppSecretKey.SmsProviderSecrets, AppSecretKey.WebhookSecret];

interface LocationState {
  isRevealSecrets: boolean;
}

function isLocationState(raw: unknown): raw is LocationState {
  return (
    raw != null &&
    typeof raw === "object" &&
    (raw as Partial<LocationState>).isRevealSecrets != null
  );
}

export type FormModel = Omit<
  AppSecretConfigFormModel<FormState>,
  "saveWithState"
>;

enum SMSProviderType {
  Authgear = "authgear",
  Twilio = "twilio",
  Aliyun = "aliyun",
  AliyunMAS = "aliyun_mas",
  Tencent = "tencent",
  Yunpian = "yunpian",
  SMSBao = "smsbao",
  GatewayAPI = "gatewayapi",
  SMSAero = "smsaero",
  Webhook = "webhook",
  Deno = "deno",
}

enum TwilioSenderType {
  MessagingServiceSID = "MessagingServiceSID",
  From = "From",
}

const MASK = "********";

// Matches v2 IconRadioCards storybook inner icon size (SquareIcon iconSize).
const PROVIDER_RADIO_ICON_SIZE = "1.375rem";

interface ConfigFormState
  extends AliyunFormState,
    AliyunMASFormState,
    TencentFormState,
    YunpianFormState,
    SMSBaoFormState,
    GatewayAPIFormState,
    SMSAeroFormState {
  enabled: boolean;
  providerType: SMSProviderType;
  webhookSecretKey: string | null;

  // twilio
  twilioCredentialType: TwilioCredentialType;
  twilioSID: string;
  twilioAuthToken: string | null;
  twilioAPIKeySID: string;
  twilioAPIKeySecret: string | null;
  twilioSenderType: TwilioSenderType;
  twilioMessagingServiceSID: string;
  twilioFrom: string;

  // webhook
  webhookURL: string;
  webhookTimeout: number;

  // deno
  denoHookURL: string;
  denoHookTimeout: number;
}

interface FormState extends ConfigFormState {
  resources: Resource[];
  diff: ResourcesDiffResult | null;

  isSMSRequiredForSomeEnabledFeatures: boolean;
  smsProviderConfigured: boolean;
}

function constructFormState(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): ConfigFormState {
  let enabled: boolean;
  let providerType: SMSProviderType;

  // This implementation only handles the new sms_gateway config and ignores the old sms_provider config
  const isSMSGatewayIsTwilio =
    config.messaging?.sms_gateway?.provider === "twilio";
  const hasCustomTwilioCredentials =
    secrets.smsProviderSecrets?.twilioCredentials != null;

  const isSMSGatewayIsAliyun =
    config.messaging?.sms_gateway?.provider === "aliyun";
  const hasAliyunCredentials =
    secrets.smsProviderSecrets?.aliyunCredentials != null;

  const isSMSGatewayIsAliyunMAS =
    config.messaging?.sms_gateway?.provider === "aliyun_mas";
  const hasAliyunMASCredentials =
    secrets.smsProviderSecrets?.aliyunMASCredentials != null;

  const isSMSGatewayIsTencent =
    config.messaging?.sms_gateway?.provider === "tencent";
  const hasTencentCredentials =
    secrets.smsProviderSecrets?.tencentCredentials != null;

  const isSMSGatewayIsYunpian =
    config.messaging?.sms_gateway?.provider === "yunpian";
  const hasYunpianCredentials =
    secrets.smsProviderSecrets?.yunpianCredentials != null;

  const isSMSGatewayIsSMSBao =
    config.messaging?.sms_gateway?.provider === "smsbao";
  const hasSMSBaoCredentials =
    secrets.smsProviderSecrets?.smsbaoCredentials != null;

  const isSMSGatewayIsGatewayAPI =
    config.messaging?.sms_gateway?.provider === "gatewayapi";
  const hasGatewayAPICredentials =
    secrets.smsProviderSecrets?.gatewayAPICredentials != null;

  const isSMSGatewayIsSMSAero =
    config.messaging?.sms_gateway?.provider === "smsaero";
  const hasSMSAeroCredentials =
    secrets.smsProviderSecrets?.smsAeroCredentials != null;

  const isSMSGatewayIsCustom =
    config.messaging?.sms_gateway?.provider === "custom";
  const hasCustomProviderSecrets =
    secrets.smsProviderSecrets?.customSMSProviderCredentials != null;

  if (isSMSGatewayIsTwilio && hasCustomTwilioCredentials) {
    enabled = true;
    providerType = SMSProviderType.Twilio;
  } else if (isSMSGatewayIsAliyun && hasAliyunCredentials) {
    enabled = true;
    providerType = SMSProviderType.Aliyun;
  } else if (isSMSGatewayIsAliyunMAS && hasAliyunMASCredentials) {
    enabled = true;
    providerType = SMSProviderType.AliyunMAS;
  } else if (isSMSGatewayIsTencent && hasTencentCredentials) {
    enabled = true;
    providerType = SMSProviderType.Tencent;
  } else if (isSMSGatewayIsYunpian && hasYunpianCredentials) {
    enabled = true;
    providerType = SMSProviderType.Yunpian;
  } else if (isSMSGatewayIsSMSBao && hasSMSBaoCredentials) {
    enabled = true;
    providerType = SMSProviderType.SMSBao;
  } else if (isSMSGatewayIsGatewayAPI && hasGatewayAPICredentials) {
    enabled = true;
    providerType = SMSProviderType.GatewayAPI;
  } else if (isSMSGatewayIsSMSAero && hasSMSAeroCredentials) {
    enabled = true;
    providerType = SMSProviderType.SMSAero;
  } else if (isSMSGatewayIsCustom && hasCustomProviderSecrets) {
    enabled = true;
    if (
      getHookKind(
        secrets.smsProviderSecrets!.customSMSProviderCredentials!.url
      ) === "denohook"
    ) {
    }
    providerType =
      getHookKind(
        secrets.smsProviderSecrets!.customSMSProviderCredentials!.url
      ) === "denohook"
        ? SMSProviderType.Deno
        : SMSProviderType.Webhook;
  } else {
    enabled = false;
    providerType = SMSProviderType.Authgear;
  }

  let twilioCredentialType: TwilioCredentialType = TwilioCredentialType.ApiKey;
  let twilioSID = "";
  let twilioAPIKeySID = "";
  let twilioAuthToken: string | null = "";
  let twilioAPIKeySecret: string | null = "";
  let twilioSenderType: TwilioSenderType = TwilioSenderType.From;
  let twilioMessagingServiceSID = "";
  let twilioFrom = "";

  if (enabled && providerType === SMSProviderType.Twilio) {
    twilioSID = secrets.smsProviderSecrets?.twilioCredentials?.accountSID ?? "";
    twilioCredentialType =
      secrets.smsProviderSecrets?.twilioCredentials?.credentialType ??
      TwilioCredentialType.AuthToken;
    switch (twilioCredentialType) {
      case TwilioCredentialType.AuthToken:
        twilioAuthToken =
          secrets.smsProviderSecrets?.twilioCredentials != null
            ? secrets.smsProviderSecrets.twilioCredentials.authToken ?? null
            : "";
        break;
      case TwilioCredentialType.ApiKey:
        twilioAPIKeySID =
          secrets.smsProviderSecrets?.twilioCredentials?.apiKeySID ?? "";
        twilioAPIKeySecret =
          secrets.smsProviderSecrets?.twilioCredentials != null
            ? secrets.smsProviderSecrets.twilioCredentials.apiKeySecret ?? null
            : "";
    }

    if (secrets.smsProviderSecrets?.twilioCredentials?.messagingServiceSID) {
      twilioSenderType = TwilioSenderType.MessagingServiceSID;
      twilioMessagingServiceSID =
        secrets.smsProviderSecrets.twilioCredentials.messagingServiceSID;
    } else if (secrets.smsProviderSecrets?.twilioCredentials?.from) {
      twilioSenderType = TwilioSenderType.From;
      twilioFrom = secrets.smsProviderSecrets.twilioCredentials.from;
    }
  }

  let aliyunAccessKeyID = "";
  let aliyunAccessKeySecret: string | null = "";
  let aliyunSignName = "";
  let aliyunTemplateCode = "";
  let aliyunTemplateCodes: SMSTemplateCodes = {};
  let aliyunOverseasTemplateCode = "";

  if (enabled && providerType === SMSProviderType.Aliyun) {
    const credentials = secrets.smsProviderSecrets?.aliyunCredentials;
    aliyunAccessKeyID = credentials?.accessKeyID ?? "";
    aliyunAccessKeySecret =
      credentials != null ? credentials.accessKeySecret ?? null : "";
    aliyunSignName = credentials?.signName ?? "";
    aliyunTemplateCode = credentials?.templateCode ?? "";
    aliyunTemplateCodes = parseSMSTemplateCodes(credentials?.templateCodes);
    aliyunOverseasTemplateCode = credentials?.overseasTemplateCode ?? "";
  }

  let aliyunMASAccessKeyID = "";
  let aliyunMASAccessKeySecret: string | null = "";
  let aliyunMASSignName = "";
  let aliyunMASTemplateCode = "";
  let aliyunMASTemplateCodes: SMSTemplateCodes = {};

  if (enabled && providerType === SMSProviderType.AliyunMAS) {
    const credentials = secrets.smsProviderSecrets?.aliyunMASCredentials;
    aliyunMASAccessKeyID = credentials?.accessKeyID ?? "";
    aliyunMASAccessKeySecret =
      credentials != null ? credentials.accessKeySecret ?? null : "";
    aliyunMASSignName = credentials?.signName ?? "";
    aliyunMASTemplateCode = credentials?.templateCode ?? "";
    aliyunMASTemplateCodes = parseSMSTemplateCodes(credentials?.templateCodes);
  }

  let tencentSecretID = "";
  let tencentSecretKey: string | null = "";
  let tencentSDKAppID = "";
  let tencentRegion = "";
  let tencentSignName = "";
  let tencentTemplateCode = "";
  let tencentTemplateCodes: SMSTemplateCodes = {};

  if (enabled && providerType === SMSProviderType.Tencent) {
    const credentials = secrets.smsProviderSecrets?.tencentCredentials;
    tencentSecretID = credentials?.secretID ?? "";
    tencentSecretKey = credentials != null ? credentials.secretKey ?? null : "";
    tencentSDKAppID = credentials?.sdkAppID ?? "";
    tencentRegion = credentials?.region ?? "";
    tencentSignName = credentials?.signName ?? "";
    tencentTemplateCode = credentials?.templateCode ?? "";
    tencentTemplateCodes = parseSMSTemplateCodes(credentials?.templateCodes);
  }

  let yunpianAPIKey: string | null = "";

  if (enabled && providerType === SMSProviderType.Yunpian) {
    const credentials = secrets.smsProviderSecrets?.yunpianCredentials;
    yunpianAPIKey = credentials != null ? credentials.apiKey ?? null : "";
  }

  let smsbaoUsername = "";
  let smsbaoPasswordOrAPIKey: string | null = "";
  let smsbaoGoodsID = "";

  if (enabled && providerType === SMSProviderType.SMSBao) {
    const credentials = secrets.smsProviderSecrets?.smsbaoCredentials;
    smsbaoUsername = credentials?.username ?? "";
    smsbaoPasswordOrAPIKey =
      credentials != null ? credentials.passwordOrAPIKey ?? null : "";
    smsbaoGoodsID = credentials?.goodsID ?? "";
  }

  let gatewayAPIEndpoint: string = DEFAULT_GATEWAY_API_ENDPOINT;
  let gatewayAPIAPIToken: string | null = "";
  let gatewayAPISender = "";

  if (enabled && providerType === SMSProviderType.GatewayAPI) {
    const credentials = secrets.smsProviderSecrets?.gatewayAPICredentials;
    gatewayAPIEndpoint =
      credentials?.endpoint != null && credentials.endpoint !== ""
        ? credentials.endpoint
        : DEFAULT_GATEWAY_API_ENDPOINT;
    gatewayAPIAPIToken =
      credentials != null ? credentials.apiToken ?? null : "";
    gatewayAPISender = credentials?.sender ?? "";
  }

  let smsAeroEmail = "";
  let smsAeroAPIKey: string | null = "";
  let smsAeroSenderName = "";

  if (enabled && providerType === SMSProviderType.SMSAero) {
    const credentials = secrets.smsProviderSecrets?.smsAeroCredentials;
    smsAeroEmail = credentials?.email ?? "";
    smsAeroAPIKey = credentials != null ? credentials.apiKey ?? null : "";
    smsAeroSenderName = credentials?.senderName ?? "";
  }

  let webhookURL = "";
  let webhookTimeout = 30;

  let denoHookURL = "";
  let denoHookTimeout = 30;

  if (
    enabled &&
    (providerType === SMSProviderType.Webhook ||
      providerType === SMSProviderType.Deno) &&
    secrets.smsProviderSecrets?.customSMSProviderCredentials != null
  ) {
    if (
      getHookKind(
        secrets.smsProviderSecrets.customSMSProviderCredentials.url
      ) === "denohook"
    ) {
      denoHookURL = secrets.smsProviderSecrets.customSMSProviderCredentials.url;
    } else {
      webhookURL = secrets.smsProviderSecrets.customSMSProviderCredentials.url;
    }
    if (
      secrets.smsProviderSecrets.customSMSProviderCredentials.timeout != null
    ) {
      denoHookTimeout =
        secrets.smsProviderSecrets.customSMSProviderCredentials.timeout;
      webhookTimeout =
        secrets.smsProviderSecrets.customSMSProviderCredentials.timeout;
    }
  }
  return {
    enabled,
    providerType,
    webhookSecretKey: secrets.webhookSecret?.secret ?? null,

    twilioCredentialType,
    twilioSID,
    twilioAuthToken,
    twilioAPIKeySID,
    twilioAPIKeySecret,
    twilioSenderType,
    twilioMessagingServiceSID,
    twilioFrom,

    aliyunAccessKeyID,
    aliyunAccessKeySecret,
    aliyunSignName,
    aliyunTemplateCode,
    aliyunTemplateCodes,
    aliyunOverseasTemplateCode,

    aliyunMASAccessKeyID,
    aliyunMASAccessKeySecret,
    aliyunMASSignName,
    aliyunMASTemplateCode,
    aliyunMASTemplateCodes,

    tencentSecretID,
    tencentSecretKey,
    tencentSDKAppID,
    tencentRegion,
    tencentSignName,
    tencentTemplateCode,
    tencentTemplateCodes,

    yunpianAPIKey,

    smsbaoUsername,
    smsbaoPasswordOrAPIKey,
    smsbaoGoodsID,

    gatewayAPIEndpoint,
    gatewayAPIAPIToken,
    gatewayAPISender,

    smsAeroEmail,
    smsAeroAPIKey,
    smsAeroSenderName,

    webhookURL,
    webhookTimeout,

    denoHookURL,
    denoHookTimeout,
  } satisfies ConfigFormState;
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _initialState: ConfigFormState,
  currentState: ConfigFormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  const newConfig = produce(config, (config) => {
    config.messaging ??= {};
    if (!currentState.enabled) {
      config.messaging.sms_gateway = undefined;
      config.messaging.sms_provider = undefined;
    } else {
      config.messaging.sms_provider = undefined;

      let newProvider: SMSProvider;
      switch (currentState.providerType) {
        case SMSProviderType.Authgear:
          config.messaging.sms_gateway = undefined;
          config.messaging.sms_provider = undefined;
          return;
        case SMSProviderType.Twilio:
          newProvider = "twilio";
          break;
        case SMSProviderType.Aliyun:
          newProvider = "aliyun";
          break;
        case SMSProviderType.AliyunMAS:
          newProvider = "aliyun_mas";
          break;
        case SMSProviderType.Tencent:
          newProvider = "tencent";
          break;
        case SMSProviderType.Yunpian:
          newProvider = "yunpian";
          break;
        case SMSProviderType.SMSBao:
          newProvider = "smsbao";
          break;
        case SMSProviderType.GatewayAPI:
          newProvider = "gatewayapi";
          break;
        case SMSProviderType.SMSAero:
          newProvider = "smsaero";
          break;
        case SMSProviderType.Deno:
          newProvider = "custom";
          break;
        case SMSProviderType.Webhook:
          newProvider = "custom";
          break;
      }

      config.messaging.sms_gateway = {
        provider: newProvider,
        use_config_from: "authgear.secrets.yaml",
      };
    }
  });

  const newSecrets = produce(secrets, (secrets) => {
    if (!currentState.enabled) {
      secrets.smsProviderSecrets = null;
    } else {
      switch (currentState.providerType) {
        case SMSProviderType.Authgear:
          secrets.smsProviderSecrets = null;
          break;
        case SMSProviderType.Twilio: {
          const twilioCredentials: SMSProviderTwilioCredentials = {
            credentialType: currentState.twilioCredentialType,
            accountSID: currentState.twilioSID,
          };
          switch (currentState.twilioCredentialType) {
            case TwilioCredentialType.ApiKey:
              twilioCredentials.apiKeySID = currentState.twilioAPIKeySID;
              twilioCredentials.apiKeySecret = currentState.twilioAPIKeySecret;
              break;
            case TwilioCredentialType.AuthToken:
              twilioCredentials.authToken = currentState.twilioAuthToken;
              break;
          }
          switch (currentState.twilioSenderType) {
            case TwilioSenderType.From:
              twilioCredentials.from = currentState.twilioFrom;
              break;
            case TwilioSenderType.MessagingServiceSID:
              twilioCredentials.messagingServiceSID =
                currentState.twilioMessagingServiceSID;
              break;
          }
          secrets.smsProviderSecrets = { twilioCredentials: twilioCredentials };
          break;
        }
        case SMSProviderType.Aliyun: {
          const aliyunCredentials: SMSProviderAliyunCredentials = {
            accessKeyID: currentState.aliyunAccessKeyID,
            accessKeySecret: currentState.aliyunAccessKeySecret,
            signName: currentState.aliyunSignName,
            templateCode: currentState.aliyunTemplateCode,
            templateCodes: serializeSMSTemplateCodes(
              currentState.aliyunTemplateCodes
            ),
            overseasTemplateCode: currentState.aliyunOverseasTemplateCode,
          };
          secrets.smsProviderSecrets = { aliyunCredentials: aliyunCredentials };
          break;
        }
        case SMSProviderType.AliyunMAS: {
          const aliyunMASCredentials: SMSProviderAliyunMASCredentials = {
            accessKeyID: currentState.aliyunMASAccessKeyID,
            accessKeySecret: currentState.aliyunMASAccessKeySecret,
            signName: currentState.aliyunMASSignName,
            templateCode: currentState.aliyunMASTemplateCode,
            templateCodes: serializeSMSTemplateCodes(
              currentState.aliyunMASTemplateCodes
            ),
          };
          secrets.smsProviderSecrets = {
            aliyunMASCredentials: aliyunMASCredentials,
          };
          break;
        }
        case SMSProviderType.Tencent: {
          const tencentCredentials: SMSProviderTencentCredentials = {
            secretID: currentState.tencentSecretID,
            secretKey: currentState.tencentSecretKey,
            sdkAppID: currentState.tencentSDKAppID,
            region: currentState.tencentRegion,
            signName: currentState.tencentSignName,
            templateCode: currentState.tencentTemplateCode,
            templateCodes: serializeSMSTemplateCodes(
              currentState.tencentTemplateCodes
            ),
          };
          secrets.smsProviderSecrets = {
            tencentCredentials: tencentCredentials,
          };
          break;
        }
        case SMSProviderType.Yunpian: {
          const yunpianCredentials: SMSProviderYunpianCredentials = {
            apiKey: currentState.yunpianAPIKey,
          };
          secrets.smsProviderSecrets = {
            yunpianCredentials: yunpianCredentials,
          };
          break;
        }
        case SMSProviderType.SMSBao: {
          const smsbaoCredentials: SMSProviderSmsbaoCredentials = {
            username: currentState.smsbaoUsername,
            passwordOrAPIKey: currentState.smsbaoPasswordOrAPIKey,
            goodsID: currentState.smsbaoGoodsID,
          };
          secrets.smsProviderSecrets = {
            smsbaoCredentials: smsbaoCredentials,
          };
          break;
        }
        case SMSProviderType.GatewayAPI: {
          const gatewayAPICredentials: SMSProviderGatewayAPICredentials = {
            endpoint: currentState.gatewayAPIEndpoint,
            apiToken: currentState.gatewayAPIAPIToken,
            sender: currentState.gatewayAPISender,
          };
          secrets.smsProviderSecrets = {
            gatewayAPICredentials: gatewayAPICredentials,
          };
          break;
        }
        case SMSProviderType.SMSAero: {
          const smsAeroCredentials: SMSProviderSmsAeroCredentials = {
            email: currentState.smsAeroEmail,
            apiKey: currentState.smsAeroAPIKey,
            senderName: currentState.smsAeroSenderName,
          };
          secrets.smsProviderSecrets = {
            smsAeroCredentials: smsAeroCredentials,
          };
          break;
        }
        case SMSProviderType.Webhook:
          secrets.smsProviderSecrets = {
            customSMSProviderCredentials: {
              url: currentState.webhookURL,
              timeout: currentState.webhookTimeout,
            },
          };
          break;
        case SMSProviderType.Deno:
          secrets.smsProviderSecrets = {
            customSMSProviderCredentials: {
              url: currentState.denoHookURL,
              timeout: currentState.denoHookTimeout,
            },
          };
          break;
      }
    }
  });
  return [newConfig, newSecrets];
}

function constructSecretUpdateInstruction(
  _config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  currentState: ConfigFormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  if (!currentState.enabled || !secrets.smsProviderSecrets) {
    // Remove all existing secrets
    return {
      smsProviderSecrets: {
        action: "set",
        setData: {},
      },
    };
  }

  switch (currentState.providerType) {
    case SMSProviderType.Authgear:
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {},
        },
      };
    case SMSProviderType.Twilio:
      if (secrets.smsProviderSecrets.twilioCredentials == null) {
        console.error("unexpected null twilioCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            twilioCredentials: {
              credentialType:
                secrets.smsProviderSecrets.twilioCredentials.credentialType,
              accountSID:
                secrets.smsProviderSecrets.twilioCredentials.accountSID,
              authToken: secrets.smsProviderSecrets.twilioCredentials.authToken,
              apiKeySID: secrets.smsProviderSecrets.twilioCredentials.apiKeySID,
              apiKeySecret:
                secrets.smsProviderSecrets.twilioCredentials.apiKeySecret,
              messagingServiceSID:
                secrets.smsProviderSecrets.twilioCredentials
                  .messagingServiceSID,
              from: secrets.smsProviderSecrets.twilioCredentials.from,
            },
          },
        },
      };
    case SMSProviderType.Aliyun:
      if (secrets.smsProviderSecrets.aliyunCredentials == null) {
        console.error("unexpected null aliyunCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            aliyunCredentials: {
              accessKeyID:
                secrets.smsProviderSecrets.aliyunCredentials.accessKeyID,
              accessKeySecret:
                secrets.smsProviderSecrets.aliyunCredentials.accessKeySecret,
              signName: secrets.smsProviderSecrets.aliyunCredentials.signName,
              templateCode:
                secrets.smsProviderSecrets.aliyunCredentials.templateCode,
              templateCodes:
                secrets.smsProviderSecrets.aliyunCredentials.templateCodes,
              overseasTemplateCode:
                secrets.smsProviderSecrets.aliyunCredentials
                  .overseasTemplateCode,
            },
          },
        },
      };
    case SMSProviderType.AliyunMAS:
      if (secrets.smsProviderSecrets.aliyunMASCredentials == null) {
        console.error("unexpected null aliyunMASCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            aliyunMASCredentials: {
              accessKeyID:
                secrets.smsProviderSecrets.aliyunMASCredentials.accessKeyID,
              accessKeySecret:
                secrets.smsProviderSecrets.aliyunMASCredentials.accessKeySecret,
              signName:
                secrets.smsProviderSecrets.aliyunMASCredentials.signName,
              templateCode:
                secrets.smsProviderSecrets.aliyunMASCredentials.templateCode,
              templateCodes:
                secrets.smsProviderSecrets.aliyunMASCredentials.templateCodes,
            },
          },
        },
      };
    case SMSProviderType.Tencent:
      if (secrets.smsProviderSecrets.tencentCredentials == null) {
        console.error("unexpected null tencentCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            tencentCredentials: {
              secretID: secrets.smsProviderSecrets.tencentCredentials.secretID,
              secretKey:
                secrets.smsProviderSecrets.tencentCredentials.secretKey,
              sdkAppID: secrets.smsProviderSecrets.tencentCredentials.sdkAppID,
              region: secrets.smsProviderSecrets.tencentCredentials.region,
              signName: secrets.smsProviderSecrets.tencentCredentials.signName,
              templateCode:
                secrets.smsProviderSecrets.tencentCredentials.templateCode,
              templateCodes:
                secrets.smsProviderSecrets.tencentCredentials.templateCodes,
            },
          },
        },
      };
    case SMSProviderType.Yunpian:
      if (secrets.smsProviderSecrets.yunpianCredentials == null) {
        console.error("unexpected null yunpianCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            yunpianCredentials: {
              apiKey: secrets.smsProviderSecrets.yunpianCredentials.apiKey,
            },
          },
        },
      };
    case SMSProviderType.SMSBao:
      if (secrets.smsProviderSecrets.smsbaoCredentials == null) {
        console.error("unexpected null smsbaoCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            smsbaoCredentials: {
              username: secrets.smsProviderSecrets.smsbaoCredentials.username,
              passwordOrAPIKey:
                secrets.smsProviderSecrets.smsbaoCredentials.passwordOrAPIKey,
              goodsID: secrets.smsProviderSecrets.smsbaoCredentials.goodsID,
            },
          },
        },
      };
    case SMSProviderType.GatewayAPI:
      if (secrets.smsProviderSecrets.gatewayAPICredentials == null) {
        console.error("unexpected null gatewayAPICredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            gatewayAPICredentials: {
              endpoint:
                secrets.smsProviderSecrets.gatewayAPICredentials.endpoint,
              apiToken:
                secrets.smsProviderSecrets.gatewayAPICredentials.apiToken,
              sender: secrets.smsProviderSecrets.gatewayAPICredentials.sender,
            },
          },
        },
      };
    case SMSProviderType.SMSAero:
      if (secrets.smsProviderSecrets.smsAeroCredentials == null) {
        console.error("unexpected null smsAeroCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            smsAeroCredentials: {
              email: secrets.smsProviderSecrets.smsAeroCredentials.email,
              apiKey: secrets.smsProviderSecrets.smsAeroCredentials.apiKey,
              senderName:
                secrets.smsProviderSecrets.smsAeroCredentials.senderName,
            },
          },
        },
      };
    case SMSProviderType.Webhook:
      if (secrets.smsProviderSecrets.customSMSProviderCredentials == null) {
        console.error("unexpected null customSMSProviderCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            customSMSProviderCredentials: {
              url: secrets.smsProviderSecrets.customSMSProviderCredentials.url,
              timeout:
                secrets.smsProviderSecrets.customSMSProviderCredentials.timeout,
            },
          },
        },
      };
    case SMSProviderType.Deno:
      if (secrets.smsProviderSecrets.customSMSProviderCredentials == null) {
        console.error("unexpected null customSMSProviderCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            customSMSProviderCredentials: {
              url: secrets.smsProviderSecrets.customSMSProviderCredentials.url,
              timeout:
                secrets.smsProviderSecrets.customSMSProviderCredentials.timeout,
            },
          },
        },
      };
  }
}

const localErrorFromRequired: LocalError = {
  errorName: "__local",
  reason: "__local",
  info: {
    error: {
      messageID: "errors.validation.required",
    },
  },
};

const fromErrorRules: ErrorParseRule[] = [
  makeLocalErrorParseRule(
    localErrorFromRequired,
    localErrorFromRequired.info.error
  ),
];

const localErrorMessagingServiceSIDRequired: LocalError = {
  errorName: "__local",
  reason: "__local",
  info: {
    error: {
      messageID: "errors.validation.required",
    },
  },
};

const messagingServiceSIDErrorRules: ErrorParseRule[] = [
  makeLocalErrorParseRule(
    localErrorMessagingServiceSIDRequired,
    localErrorMessagingServiceSIDRequired.info.error
  ),
];

function makeSpecifiersFromState(state: ConfigFormState): ResourceSpecifier[] {
  const specifiers: ResourceSpecifier[] = [];
  if (state.denoHookURL) {
    specifiers.push(makeDenoScriptSpecifier(state.denoHookURL));
  }
  return specifiers;
}

function makeNewDenoScriptURL(): string {
  const rand = genRandomHexadecimalString();
  return `authgeardeno:///deno/sms.${rand}.ts`;
}

const DEFAULT_SMS_SCRIPT_TEMPLATE = `// This custom script will be executed when a message is triggered
// Sample script:
import { CustomSMSGatewayPayload, CustomSMSGatewayResponse } from "${DENO_TYPES_URL}";

export default async function (e: CustomSMSGatewayPayload): Promise<CustomSMSGatewayResponse> {
  const body = JSON.stringify(e);
  const response = await fetch("https://some.sms.gateway", {
    method: "POST",
    body: body,
  });

  if (!response.ok) {
    return {
      code: "delivery_rejected",
    }
  }

  return {
    code: "ok",
  }
}
`;

const CODE_EDITOR_OPTIONS = {
  minimap: {
    enabled: false,
  },
};

function useDenoScriptResourceIndex(state: FormState) {
  const resourceIdx = useMemo(() => {
    if (state.denoHookURL === "") {
      return -1;
    }
    const path = getDenoScriptPathFromURL(state.denoHookURL);
    for (const [idx, r] of state.resources.entries()) {
      if (r.path === path && r.nullableValue != null) {
        return idx;
      }
    }
    return -1;
  }, [state.denoHookURL, state.resources]);
  return resourceIdx;
}

function useTestSMSConfig(
  state: FormState
): SmsProviderConfigurationInput | null {
  const denoResourceIdx = useDenoScriptResourceIndex(state);

  return useMemo((): SmsProviderConfigurationInput | null => {
    if (!state.enabled) {
      return null;
    }
    switch (state.providerType) {
      case SMSProviderType.Authgear:
        return null;
      case SMSProviderType.Twilio: {
        if (!state.twilioSID) {
          return null;
        }
        const twilio: SmsProviderConfigurationTwilioInput = {
          credentialType: state.twilioCredentialType,
          accountSID: state.twilioSID,
          authToken: state.twilioAuthToken ?? "",
          apiKeySID: state.twilioAPIKeySID,
          apiKeySecret: state.twilioAPIKeySecret ?? "",
        };
        switch (state.twilioCredentialType) {
          case TwilioCredentialType.ApiKey:
            twilio.apiKeySID = state.twilioAPIKeySID;
            twilio.apiKeySecret = state.twilioAPIKeySecret;
            break;
          case TwilioCredentialType.AuthToken:
            twilio.authToken = state.twilioAuthToken;
            break;
        }
        switch (state.twilioSenderType) {
          case TwilioSenderType.From:
            twilio.from = state.twilioFrom;
            break;
          case TwilioSenderType.MessagingServiceSID:
            twilio.messagingServiceSID = state.twilioMessagingServiceSID;
            break;
        }
        return {
          twilio: {
            credentialType: state.twilioCredentialType,
            accountSID: state.twilioSID,
            authToken: state.twilioAuthToken ?? "",
            apiKeySID: state.twilioAPIKeySID,
            apiKeySecret: state.twilioAPIKeySecret ?? "",
            messagingServiceSID: state.twilioMessagingServiceSID,
            from: state.twilioFrom,
          },
        };
      }
      case SMSProviderType.Aliyun: {
        if (
          !state.aliyunAccessKeyID ||
          !state.aliyunAccessKeySecret ||
          !state.aliyunSignName ||
          !state.aliyunTemplateCode
        ) {
          return null;
        }
        return {
          aliyun: {
            accessKeyID: state.aliyunAccessKeyID,
            accessKeySecret: state.aliyunAccessKeySecret,
            signName: state.aliyunSignName,
            templateCode: state.aliyunTemplateCode,
            templateCodes: serializeSMSTemplateCodes(state.aliyunTemplateCodes),
            overseasTemplateCode: state.aliyunOverseasTemplateCode,
          },
        };
      }
      case SMSProviderType.AliyunMAS: {
        if (
          !state.aliyunMASAccessKeyID ||
          !state.aliyunMASAccessKeySecret ||
          !state.aliyunMASSignName ||
          !state.aliyunMASTemplateCode
        ) {
          return null;
        }
        return {
          aliyunMAS: {
            accessKeyID: state.aliyunMASAccessKeyID,
            accessKeySecret: state.aliyunMASAccessKeySecret,
            signName: state.aliyunMASSignName,
            templateCode: state.aliyunMASTemplateCode,
            templateCodes: serializeSMSTemplateCodes(
              state.aliyunMASTemplateCodes
            ),
          },
        };
      }
      case SMSProviderType.Tencent: {
        if (
          !state.tencentSecretID ||
          !state.tencentSecretKey ||
          !state.tencentSDKAppID ||
          !state.tencentSignName ||
          !state.tencentTemplateCode
        ) {
          return null;
        }
        return {
          tencent: {
            secretID: state.tencentSecretID,
            secretKey: state.tencentSecretKey,
            sdkAppID: state.tencentSDKAppID,
            region: state.tencentRegion,
            signName: state.tencentSignName,
            templateCode: state.tencentTemplateCode,
            templateCodes: serializeSMSTemplateCodes(
              state.tencentTemplateCodes
            ),
          },
        };
      }
      case SMSProviderType.Yunpian: {
        if (!state.yunpianAPIKey) {
          return null;
        }
        return {
          yunpian: {
            apiKey: state.yunpianAPIKey,
          },
        };
      }
      case SMSProviderType.SMSBao: {
        if (!state.smsbaoUsername || !state.smsbaoPasswordOrAPIKey) {
          return null;
        }
        return {
          smsbao: {
            username: state.smsbaoUsername,
            passwordOrAPIKey: state.smsbaoPasswordOrAPIKey,
            goodsID: state.smsbaoGoodsID,
          },
        };
      }
      case SMSProviderType.GatewayAPI: {
        if (!state.gatewayAPIAPIToken || !state.gatewayAPISender) {
          return null;
        }
        return {
          gatewayAPI: {
            endpoint: state.gatewayAPIEndpoint,
            apiToken: state.gatewayAPIAPIToken,
            sender: state.gatewayAPISender,
          },
        };
      }
      case SMSProviderType.SMSAero: {
        if (
          !state.smsAeroEmail ||
          !state.smsAeroAPIKey ||
          !state.smsAeroSenderName
        ) {
          return null;
        }
        return {
          smsAero: {
            email: state.smsAeroEmail,
            apiKey: state.smsAeroAPIKey,
            senderName: state.smsAeroSenderName,
          },
        };
      }
      case SMSProviderType.Webhook:
        if (!state.webhookURL) {
          return null;
        }
        return {
          webhook: {
            url: state.webhookURL,
            timeout: state.webhookTimeout,
          },
        };
      case SMSProviderType.Deno: {
        if (denoResourceIdx === -1) {
          return null;
        }
        const script = state.resources[denoResourceIdx].nullableValue ?? "";
        if (!script) {
          return null;
        }
        return {
          deno: {
            script: script,
            timeout: state.denoHookTimeout,
          },
        };
      }
    }
  }, [
    denoResourceIdx,
    state.aliyunAccessKeyID,
    state.aliyunAccessKeySecret,
    state.aliyunOverseasTemplateCode,
    state.aliyunSignName,
    state.aliyunTemplateCode,
    state.aliyunTemplateCodes,
    state.aliyunMASAccessKeyID,
    state.aliyunMASAccessKeySecret,
    state.aliyunMASSignName,
    state.aliyunMASTemplateCode,
    state.aliyunMASTemplateCodes,
    state.denoHookTimeout,
    state.enabled,
    state.gatewayAPIAPIToken,
    state.gatewayAPIEndpoint,
    state.gatewayAPISender,
    state.providerType,
    state.resources,
    state.smsAeroAPIKey,
    state.smsAeroEmail,
    state.smsAeroSenderName,
    state.smsbaoGoodsID,
    state.smsbaoPasswordOrAPIKey,
    state.smsbaoUsername,
    state.tencentRegion,
    state.tencentSDKAppID,
    state.tencentSecretID,
    state.tencentSecretKey,
    state.tencentSignName,
    state.tencentTemplateCode,
    state.tencentTemplateCodes,
    state.twilioAPIKeySID,
    state.twilioAPIKeySecret,
    state.twilioAuthToken,
    state.twilioCredentialType,
    state.twilioFrom,
    state.twilioMessagingServiceSID,
    state.twilioSID,
    state.twilioSenderType,
    state.webhookTimeout,
    state.webhookURL,
    state.yunpianAPIKey,
  ]);
}

function computeIsSecretMasked(state: FormState): boolean {
  if (!state.enabled) {
    return false;
  }
  switch (state.providerType) {
    case SMSProviderType.Authgear:
      return false;
    case SMSProviderType.Twilio:
      switch (state.twilioCredentialType) {
        case TwilioCredentialType.ApiKey:
          return state.twilioAPIKeySecret == null;
        case TwilioCredentialType.AuthToken:
          return state.twilioAuthToken == null;
      }
      throw new Error("unreachable code");
    case SMSProviderType.Aliyun:
      return state.aliyunAccessKeySecret == null;
    case SMSProviderType.AliyunMAS:
      return state.aliyunMASAccessKeySecret == null;
    case SMSProviderType.Tencent:
      return state.tencentSecretKey == null;
    case SMSProviderType.Yunpian:
      return state.yunpianAPIKey == null;
    case SMSProviderType.SMSBao:
      return state.smsbaoPasswordOrAPIKey == null;
    case SMSProviderType.GatewayAPI:
      return state.gatewayAPIAPIToken == null;
    case SMSProviderType.SMSAero:
      return state.smsAeroAPIKey == null;
    case SMSProviderType.Webhook:
      return state.webhookSecretKey == null;
    case SMSProviderType.Deno:
      return false;
  }
  throw new Error("unreachable code");
}

const SMSProviderConfigurationScreen: React.VFC =
  function SMSProviderConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const location = useLocation();
    const [shouldRefreshToken] = useState<boolean>(() => {
      const { state } = location;
      if (isLocationState(state) && state.isRevealSecrets) {
        return true;
      }
      if (!authgear.canReauthenticate()) {
        return true;
      }
      return false;
    });
    useLocationEffect<LocationState>(() => {
      // Pop the location state if exist
    });
    const { token, loading, error, retry } = useAppSecretVisitToken(
      appID,
      SECRETS,
      shouldRefreshToken
    );

    if (error) {
      return <ShowError error={error} onRetry={retry} />;
    }

    if (loading || token === undefined) {
      return <ShowLoading />;
    }

    return (
      <SMSProviderConfigurationScreen1 appID={appID} secretToken={token} />
    );
  };

export default SMSProviderConfigurationScreen;

function SMSProviderConfigurationScreen1({
  appID,
  secretToken,
}: {
  appID: string;
  secretToken: string | null;
}) {
  const {
    effectiveAppConfig,
    isLoading: loadingAppConfig,
    loadError: appConfigError,
    refetch: refetchAppConfig,
    secretConfig,
  } = useAppAndSecretConfigQuery(appID, secretToken);
  const configForm = useAppSecretConfigForm({
    appID,
    secretVisitToken: secretToken,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });
  const featureConfig = useAppFeatureConfigQuery(appID);
  const specifiers = useMemo(() => {
    return makeSpecifiersFromState(configForm.state);
  }, [configForm.state]);
  const resources = useResourceForm(
    appID,
    specifiers,
    (resources) => resources,
    (resources) => resources
  );
  const sendTestSMSHandle = useSendTestSMSMutation(appID);
  const checkDenoHookHandle = useCheckDenoHookMutation(appID);

  const [localError, setLocalError] = useState<APIError | null>(null);

  // eslint-disable-next-line react-hooks/preserve-manual-memoization
  const state = useMemo<FormState>(() => {
    return {
      ...configForm.state,
      resources: resources.state,
      diff: resources.diff,

      isSMSRequiredForSomeEnabledFeatures:
        // primary authentication uses SMS.
        effectiveAppConfig?.authentication?.primary_authenticators?.includes(
          "oob_otp_sms"
        ) === true ||
        // secondary authenticatoin uses SMS AND secondary authentication is enabled.
        (effectiveAppConfig?.authentication?.secondary_authenticators?.includes(
          "oob_otp_sms"
        ) === true &&
          (effectiveAppConfig.authentication.secondary_authentication_mode ===
            "if_exists" ||
            effectiveAppConfig.authentication.secondary_authentication_mode ===
              "required")) ||
        // phone verification enabled.
        effectiveAppConfig?.verification?.claims?.phone_number?.enabled ===
          true,

      smsProviderConfigured:
        secretConfig?.smsProviderSecrets?.twilioCredentials != null ||
        secretConfig?.smsProviderSecrets?.aliyunCredentials != null ||
        secretConfig?.smsProviderSecrets?.aliyunMASCredentials != null ||
        secretConfig?.smsProviderSecrets?.tencentCredentials != null ||
        secretConfig?.smsProviderSecrets?.yunpianCredentials != null ||
        secretConfig?.smsProviderSecrets?.smsbaoCredentials != null ||
        secretConfig?.smsProviderSecrets?.gatewayAPICredentials != null ||
        secretConfig?.smsProviderSecrets?.smsAeroCredentials != null ||
        secretConfig?.smsProviderSecrets?.customSMSProviderCredentials != null,
    };
  }, [
    configForm.state,
    resources.state,
    resources.diff,
    effectiveAppConfig?.authentication?.primary_authenticators,
    effectiveAppConfig?.authentication?.secondary_authenticators,
    effectiveAppConfig?.authentication?.secondary_authentication_mode,
    effectiveAppConfig?.verification?.claims?.phone_number?.enabled,
    secretConfig?.smsProviderSecrets?.twilioCredentials,
    secretConfig?.smsProviderSecrets?.aliyunCredentials,
    secretConfig?.smsProviderSecrets?.aliyunMASCredentials,
    secretConfig?.smsProviderSecrets?.tencentCredentials,
    secretConfig?.smsProviderSecrets?.yunpianCredentials,
    secretConfig?.smsProviderSecrets?.smsbaoCredentials,
    secretConfig?.smsProviderSecrets?.gatewayAPICredentials,
    secretConfig?.smsProviderSecrets?.smsAeroCredentials,
    secretConfig?.smsProviderSecrets?.customSMSProviderCredentials,
  ]);

  const form: FormModel = {
    isLoading: configForm.isLoading || resources.isLoading,
    isUpdating: configForm.isUpdating || resources.isUpdating,
    getIsDirty: () => configForm.getIsDirty() || resources.getIsDirty(),
    loadError: configForm.loadError ?? resources.loadError,
    updateError: configForm.updateError ?? resources.updateError,
    state,
    setState: (fn) => {
      const newState = fn(state);
      const { resources: newResources, ...configState } = newState;
      configForm.setState(() => ({
        ...configState,
      }));
      resources.setState(() => newResources);
    },
    reload: () => {
      resources.reload();
      configForm.reload();
    },
    reset: () => {
      resources.reset();
      configForm.reset();
    },
    save: async (ignoreConflict: boolean = false) => {
      await resources.save(ignoreConflict);
      await configForm.save(ignoreConflict);
    },
  };

  const validateTemplateCodeFields = useCallback(
    (signName: string, templateCode: string) => {
      if (!signName) {
        setLocalError(localErrorSignNameRequired);
        throw new Error("sign name is required");
      }
      if (!templateCode) {
        setLocalError(localErrorTemplateCodeRequired);
        throw new Error("template code is required");
      }
    },
    []
  );

  const validateForm = useCallback(async () => {
    setLocalError(null);
    if (!form.state.enabled) {
      return;
    }
    switch (form.state.providerType) {
      case SMSProviderType.Twilio:
        switch (form.state.twilioSenderType) {
          case TwilioSenderType.From:
            if (!form.state.twilioFrom) {
              setLocalError(localErrorFromRequired);
              throw new Error("twilioFrom is required");
            }
            break;
          case TwilioSenderType.MessagingServiceSID:
            if (!form.state.twilioMessagingServiceSID) {
              setLocalError(localErrorMessagingServiceSIDRequired);
              throw new Error("twilioMessagingServiceSID is required");
            }
            break;
        }
        break;
      case SMSProviderType.Aliyun:
        validateTemplateCodeFields(
          form.state.aliyunSignName,
          form.state.aliyunTemplateCode
        );
        break;
      case SMSProviderType.AliyunMAS:
        validateTemplateCodeFields(
          form.state.aliyunMASSignName,
          form.state.aliyunMASTemplateCode
        );
        break;
      case SMSProviderType.Tencent:
        validateTemplateCodeFields(
          form.state.tencentSignName,
          form.state.tencentTemplateCode
        );
        break;
      default:
        break;
    }
  }, [
    form.state.aliyunMASSignName,
    form.state.aliyunMASTemplateCode,
    form.state.aliyunSignName,
    form.state.aliyunTemplateCode,
    form.state.enabled,
    form.state.providerType,
    form.state.tencentSignName,
    form.state.tencentTemplateCode,
    form.state.twilioFrom,
    form.state.twilioMessagingServiceSID,
    form.state.twilioSenderType,
    validateTemplateCodeFields,
  ]);

  if (loadingAppConfig || form.isLoading || featureConfig.isLoading) {
    return <ShowLoading />;
  }

  if (appConfigError ?? form.loadError ?? featureConfig.loadError) {
    return (
      <ShowError
        error={form.loadError ?? featureConfig.loadError}
        onRetry={() => {
          refetchAppConfig().finally(() => {});
          form.reload();
          featureConfig.refetch().finally(() => {});
        }}
      />
    );
  }

  return (
    <FormContainer
      form={form}
      beforeSave={validateForm}
      localError={
        checkDenoHookHandle.error ?? sendTestSMSHandle.error ?? localError
      }
    >
      <SMSProviderConfigurationContent
        form={form}
        effectiveAppConfig={effectiveAppConfig ?? undefined}
        sendTestSMSHandle={sendTestSMSHandle}
        checkDenoHookHandle={checkDenoHookHandle}
        isCustomSMSProviderDisabled={
          featureConfig.effectiveFeatureConfig?.messaging
            ?.custom_sms_provider_disabled ?? false
        }
      />
    </FormContainer>
  );
}

function SMSProviderConfigurationContent(props: {
  form: FormModel;
  effectiveAppConfig: PortalAPIAppConfig | undefined;
  sendTestSMSHandle: ReturnType<typeof useSendTestSMSMutation>;
  checkDenoHookHandle: ReturnType<typeof useCheckDenoHookMutation>;
  isCustomSMSProviderDisabled: boolean;
}) {
  const {
    form,
    effectiveAppConfig,
    checkDenoHookHandle,
    isCustomSMSProviderDisabled,
  } = props;
  const { isAuthgearOnce } = useSystemConfig();
  const { appID } = useParams() as { appID: string };
  const { state, setState } = form;
  const { isSMSRequiredForSomeEnabledFeatures, smsProviderConfigured } = state;
  const { getIsDirty } = useFormContainerBaseContext();
  const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
  const navigate = useNavigate();
  const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

  const [isReauthDialogOpen, setIsReauthDialogOpen] = useState(false);
  const [isTestSMSDialogHidden, setIsTestSMSDialogHidden] = useState(true);

  const { checkDenoHook, loading: checkDenoHookLoading } = checkDenoHookHandle;

  const isSecretMasked = useMemo(
    () => computeIsSecretMasked(form.state),
    [form.state]
  );

  const onChangeProviderType = useCallback(
    (value: SMSProviderType) => {
      if (isSecretMasked) {
        setIsReauthDialogOpen(true);
        return;
      }
      setState((s) => ({
        ...s,
        enabled: value !== SMSProviderType.Authgear,
        providerType: value,
      }));
    },
    [isSecretMasked, setState]
  );

  const triggerReauth = useCallback(() => {
    // We are going to leave, reset the form so that the confirmation dialog won't appear
    form.reset();

    startReauthentication<LocationState>(navigate, {
      isRevealSecrets: true,
    }).catch((e) => {
      // Normally there should not be any error.
      console.error(e);
    });
  }, [navigate, form]);

  const onRevealSecrets = useCallback(() => {
    setIsReauthDialogOpen(true);
  }, []);

  const testConfig = useTestSMSConfig(form.state);

  const onTestSMS = useCallback(async () => {
    if (isSecretMasked) {
      setIsReauthDialogOpen(true);
      return;
    }
    if (form.state.providerType === SMSProviderType.Deno) {
      await checkDenoHook(testConfig?.deno?.script ?? "");
    }
    setIsTestSMSDialogHidden(false);
  }, [
    checkDenoHook,
    form.state.providerType,
    isSecretMasked,
    testConfig?.deno?.script,
  ]);

  const onCancelTestSMS = useCallback(() => {
    setIsTestSMSDialogHidden(true);
  }, []);

  const providerOptions = useMemo(
    (): IconRadioCardOption<SMSProviderType>[] => [
      {
        value: SMSProviderType.Authgear,
        icon: (
          <img
            src={logoAuthany}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.authgear" />
        ),
      },
      {
        value: SMSProviderType.Twilio,
        icon: (
          <img
            src={logoTwilio}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.twilio" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
      {
        value: SMSProviderType.Aliyun,
        icon: (
          <img
            src={logoAliyun}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.aliyun" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
      {
        value: SMSProviderType.AliyunMAS,
        icon: (
          <img
            src={logoAliyunMAS}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.aliyunMAS" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
      {
        value: SMSProviderType.Tencent,
        icon: (
          <img
            src={logoTencentCloud}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.tencent" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
      {
        value: SMSProviderType.Yunpian,
        icon: (
          <img
            src={logoYunpian}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.yunpian" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
      {
        value: SMSProviderType.SMSBao,
        icon: (
          <img
            src={logoSMSBao}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.smsbao" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
      {
        value: SMSProviderType.GatewayAPI,
        icon: (
          <img
            src={logoGatewayAPI}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.gatewayapi" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
      {
        value: SMSProviderType.SMSAero,
        icon: (
          <img
            src={logoSMSAero}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.smsaero" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
      {
        value: SMSProviderType.Webhook,
        icon: (
          <img
            src={logoWebhook}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.webhook" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
      {
        value: SMSProviderType.Deno,
        icon: (
          <CodeIcon
            width={PROVIDER_RADIO_ICON_SIZE}
            height={PROVIDER_RADIO_ICON_SIZE}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.deno" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
    ],
    [isCustomSMSProviderDisabled]
  );

  const providerDescription = useMemo(() => {
    switch (state.providerType) {
      case SMSProviderType.Authgear:
        return (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.authgear.description" />
        );
      case SMSProviderType.Twilio:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.twilio.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://docs.authgear.com/customization/custom-providers/twilio">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      case SMSProviderType.Aliyun:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.aliyun.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://help.aliyun.com/zh/sms/">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      case SMSProviderType.AliyunMAS:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.aliyunMAS.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://help.aliyun.com/zh/pnvs/">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      case SMSProviderType.Tencent:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.tencent.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://cloud.tencent.com/document/product/382">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      case SMSProviderType.Yunpian:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.yunpian.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://www.yunpian.com/official/document/sms/zh_CN/introduction_api_domains">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      case SMSProviderType.SMSBao:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.smsbao.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://www.smsbao.com/openapi/">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      case SMSProviderType.GatewayAPI:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.gatewayapi.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://gatewayapi.com/docs/apis/rest/">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      case SMSProviderType.SMSAero:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.smsaero.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://smsaero.ru/integration/api/">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      case SMSProviderType.Webhook:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.webhook.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://docs.authgear.com/customization/custom-providers/webhook-custom-script">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      case SMSProviderType.Deno:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.deno.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://docs.authgear.com/customization/custom-providers/webhook-custom-script">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      default:
        return null;
    }
  }, [state.providerType]);

  const showSettings = state.providerType !== SMSProviderType.Authgear;

  return (
    <>
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
            {isAuthgearOnce ? (
              <FormattedMessage id="SMSProviderConfigurationScreen.title--authgearonce" />
            ) : (
              <FormattedMessage id="SMSProviderConfigurationScreen.title" />
            )}
          </Heading>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="SMSProviderConfigurationScreen.description" />
          </Text>
        </div>
        {isCustomSMSProviderDisabled ? (
          <FeatureDisabledCallout
            className={styles.widget}
            messageID="FeatureConfig.custom-sms-provider.disabled"
          />
        ) : null}
        {isAuthgearOnce &&
        isSMSRequiredForSomeEnabledFeatures &&
        !smsProviderConfigured ? (
          <div className={cn(styles.widget, "flex flex-col")}>
            <RedMessageBar_RemindConfigureSMSProviderInSMSProviderScreen className="self-start w-fit" />
          </div>
        ) : null}

        <div className={cn(styles.widget, styles.providerSelector)}>
          <IconRadioCards
            size="3"
            value={state.providerType}
            onValueChange={onChangeProviderType}
            options={providerOptions}
            itemFillSpaces={true}
          />
          {providerDescription != null ? (
            <Text
              as="p"
              size="1"
              color="gray"
              className={styles.providerDescription}
            >
              {providerDescription}
            </Text>
          ) : null}
        </div>

        {showSettings ? (
          <SettingsSectionCard
            className={styles.widget}
            contentClassName="gap-4"
            title={
              <FormattedMessage id="SMSProviderConfigurationScreen.settings.label" />
            }
          >
            <FormSection form={form} onRevealSecrets={onRevealSecrets} />
            {isSecretMasked ? (
              <div>
                <PrimaryButton
                  size="3"
                  disabled={isCustomSMSProviderDisabled}
                  onClick={onRevealSecrets}
                  text={<FormattedMessage id="edit" />}
                />
              </div>
            ) : (
              <div>
                <SecondaryButton
                  size="2"
                  // eslint-disable-next-line @typescript-eslint/strict-void-return
                  onClick={onTestSMS}
                  disabled={testConfig == null || checkDenoHookLoading}
                  text={
                    <FormattedMessage id="SMSProviderConfigurationScreen.testSMS" />
                  }
                />
              </div>
            )}
          </SettingsSectionCard>
        ) : null}

        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
      <ConfirmationDialog
        open={isReauthDialogOpen}
        onOpenChange={(open) => {
          if (!open) {
            setIsReauthDialogOpen(false);
          }
        }}
        title={<FormattedMessage id="ReauthDialog.title" />}
        description={<FormattedMessage id="ReauthDialog.description" />}
        confirmText={<FormattedMessage id="confirm" />}
        cancelText={<FormattedMessage id="cancel" />}
        confirmColor="indigo"
        onConfirm={triggerReauth}
        onCancel={useCallback(() => {
          setIsReauthDialogOpen(false);
        }, [])}
      />
      {testConfig != null ? (
        <TestSMSDialog
          appID={appID}
          isHidden={isTestSMSDialogHidden}
          effectiveAppConfig={effectiveAppConfig}
          input={testConfig}
          onDismiss={onCancelTestSMS}
        />
      ) : null}
    </>
  );
}

function FormSection({
  form,
  onRevealSecrets,
}: {
  form: FormModel;
  onRevealSecrets: () => void;
}) {
  switch (form.state.providerType) {
    case SMSProviderType.Authgear:
      return null;
    case SMSProviderType.Twilio:
      return <TwilioForm form={form} />;
    case SMSProviderType.Aliyun:
      return <AliyunForm state={form.state} setState={form.setState} />;
    case SMSProviderType.AliyunMAS:
      return <AliyunMASForm state={form.state} setState={form.setState} />;
    case SMSProviderType.Tencent:
      return <TencentForm state={form.state} setState={form.setState} />;
    case SMSProviderType.Yunpian:
      return <YunpianForm state={form.state} setState={form.setState} />;
    case SMSProviderType.SMSBao:
      return <SMSBaoForm state={form.state} setState={form.setState} />;
    case SMSProviderType.GatewayAPI:
      return <GatewayAPIForm state={form.state} setState={form.setState} />;
    case SMSProviderType.SMSAero:
      return <SMSAeroForm state={form.state} setState={form.setState} />;
    case SMSProviderType.Webhook:
      return <WebhookForm form={form} onRevealSecrets={onRevealSecrets} />;
    case SMSProviderType.Deno:
      return <DenoHookForm form={form} />;
  }
}

function TwilioForm({ form }: { form: FormModel }) {
  const { renderToString } = useContext(MessageContext);

  const onStringChangeCallbacks = useMemo(() => {
    const callbackFactory = (
      key:
        | "twilioSID"
        | "twilioAuthToken"
        | "twilioAPIKeySID"
        | "twilioAPIKeySecret"
        | "twilioMessagingServiceSID"
        | "twilioFrom"
    ) => {
      return (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        form.setState((prevState) => {
          const s: FormState = {
            ...prevState,
          };
          s[key] = value;
          return s;
        });
      };
    };
    return {
      twilioSID: callbackFactory("twilioSID"),
      twilioAuthToken: callbackFactory("twilioAuthToken"),
      twilioAPIKeySID: callbackFactory("twilioAPIKeySID"),
      twilioAPIKeySecret: callbackFactory("twilioAPIKeySecret"),
      twilioMessagingServiceSID: callbackFactory("twilioMessagingServiceSID"),
      twilioFrom: callbackFactory("twilioFrom"),
    };
  }, [form]);

  const isTwilioSecretMasked =
    form.state.twilioCredentialType === TwilioCredentialType.AuthToken
      ? form.state.twilioAuthToken == null
      : form.state.twilioAPIKeySecret == null;

  const onCredentialTypeChange = useCallback(
    (value: string) => {
      form.setState((prev) => ({
        ...prev,
        twilioCredentialType: value as TwilioCredentialType,
      }));
    },
    [form]
  );

  const onSenderTypeChange = useCallback(
    (value: string) => {
      form.setState((prev) => ({
        ...prev,
        twilioSenderType: value as TwilioSenderType,
      }));
    },
    [form]
  );

  return (
    <div className="flex flex-col gap-y-4">
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.twilioSID" />
        }
        value={form.state.twilioSID}
        required={true}
        onChange={onStringChangeCallbacks.twilioSID}
        disabled={isTwilioSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="account_sid"
      />
      <div className="flex flex-col gap-3">
        <FormField
          size="2"
          labelSize="2"
          label={
            <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.sender" />
          }
          required={true}
          labelSpace="1"
        >
          <RadioGroup.Root
            value={form.state.twilioSenderType}
            onValueChange={onSenderTypeChange}
            disabled={isTwilioSecretMasked}
          >
            <Flex direction="column" gap="2">
              <Text as="label" size="2">
                <Flex gap="2" align="center">
                  <RadioGroup.Item
                    value={TwilioSenderType.MessagingServiceSID}
                  />
                  <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID" />
                  <Tooltip
                    content={
                      <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID.tooltip" />
                    }
                  >
                    <InfoCircledIcon className={styles.senderInfoIcon} />
                  </Tooltip>
                </Flex>
              </Text>
              <Text as="label" size="2">
                <Flex gap="2" align="center">
                  <RadioGroup.Item value={TwilioSenderType.From} />
                  <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.twilioFrom" />
                  <Tooltip
                    content={
                      <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.twilioFrom.tooltip" />
                    }
                  >
                    <InfoCircledIcon className={styles.senderInfoIcon} />
                  </Tooltip>
                </Flex>
              </Text>
            </Flex>
          </RadioGroup.Root>
        </FormField>
        {form.state.twilioSenderType ===
        TwilioSenderType.MessagingServiceSID ? (
          <div className="flex flex-col">
            <TextField
              size="2"
              labelSize="2"
              type="text"
              placeholder={renderToString(
                "SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID.placeholder"
              )}
              value={form.state.twilioMessagingServiceSID}
              onChange={onStringChangeCallbacks.twilioMessagingServiceSID}
              disabled={isTwilioSecretMasked}
              errorRules={messagingServiceSIDErrorRules}
              parentJSONPointer={/\/secrets\/\d+\/data/}
              fieldName="message_service_sid"
            />
            <Text as="p" size="1" color="gray">
              <FormattedMessage
                id="SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID.hint"
                values={{
                  // eslint-disable-next-line react/no-unstable-nested-components
                  ExternalLink: (chunks: React.ReactNode) => (
                    <ExternalLink href="https://www.twilio.com/docs/messaging/services">
                      {chunks}
                    </ExternalLink>
                  ),
                }}
              />
            </Text>
          </div>
        ) : (
          <TextField
            size="2"
            labelSize="2"
            type="text"
            placeholder={renderToString(
              "SMSProviderConfigurationScreen.form.twilio.twilioFrom.placeholder"
            )}
            value={form.state.twilioFrom}
            onChange={onStringChangeCallbacks.twilioFrom}
            disabled={isTwilioSecretMasked}
            parentJSONPointer={/\/secrets\/\d+\/data/}
            fieldName="from"
            errorRules={fromErrorRules}
          />
        )}
      </div>
      <div className="flex flex-col gap-4">
        <FormField
          size="2"
          labelSize="2"
          label={
            <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.credentialType" />
          }
          required={true}
          labelSpace="1"
        >
          <RadioGroup.Root
            value={form.state.twilioCredentialType}
            onValueChange={onCredentialTypeChange}
            disabled={isTwilioSecretMasked}
          >
            <Flex direction="column" gap="2">
              <Text as="label" size="2">
                <Flex gap="2" align="center">
                  <RadioGroup.Item value={TwilioCredentialType.AuthToken} />
                  <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.credentialType.options.authToken" />
                </Flex>
              </Text>
              <Text as="label" size="2">
                <Flex gap="2" align="center">
                  <RadioGroup.Item value={TwilioCredentialType.ApiKey} />
                  <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.credentialType.options.apiKey" />
                </Flex>
              </Text>
            </Flex>
          </RadioGroup.Root>
        </FormField>
        {form.state.twilioCredentialType === TwilioCredentialType.AuthToken ? (
          <TextField
            size="2"
            labelSize="2"
            type="text"
            placeholder={renderToString(
              "SMSProviderConfigurationScreen.form.twilio.credentialType.options.authToken"
            )}
            value={form.state.twilioAuthToken ?? MASK}
            onChange={onStringChangeCallbacks.twilioAuthToken}
            disabled={isTwilioSecretMasked}
            parentJSONPointer={/\/secrets\/\d+\/data/}
            fieldName="auth_token"
          />
        ) : null}
        {form.state.twilioCredentialType === TwilioCredentialType.ApiKey ? (
          <>
            <TextField
              size="2"
              labelSize="2"
              type="text"
              label={
                <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.apiKeySID" />
              }
              value={form.state.twilioAPIKeySID}
              onChange={onStringChangeCallbacks.twilioAPIKeySID}
              disabled={isTwilioSecretMasked}
              parentJSONPointer={/\/secrets\/\d+\/data/}
              fieldName="api_key_sid"
            />
            <TextField
              size="2"
              labelSize="2"
              type="text"
              label={
                <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.apiKeySecret" />
              }
              value={form.state.twilioAPIKeySecret ?? MASK}
              onChange={onStringChangeCallbacks.twilioAPIKeySecret}
              disabled={isTwilioSecretMasked}
              parentJSONPointer={/\/secrets\/\d+\/data/}
              fieldName="api_key_secret"
            />
          </>
        ) : null}
      </div>
    </div>
  );
}

function RevealIconButton({
  onClick,
}: {
  onClick: () => void;
}): React.ReactElement {
  const { renderToString } = useContext(MessageContext);

  return (
    <RadixTooltip content={renderToString("reveal")}>
      <RadixIconButton
        type="button"
        variant="ghost"
        color="gray"
        size="1"
        aria-label={renderToString("reveal")}
        onClick={onClick}
        className={styles.copyIconButton}
      >
        <EyeOpenIcon width="1rem" height="1rem" />
      </RadixIconButton>
    </RadixTooltip>
  );
}

function WebhookForm({
  form,
  onRevealSecrets,
}: {
  form: FormModel;
  onRevealSecrets: () => void;
}) {
  const { renderToString } = useContext(MessageContext);

  const onURLChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      form.setState((prevState) => {
        return {
          ...prevState,
          webhookURL: value,
        } satisfies FormState;
      });
    },
    [form]
  );

  const onTimeoutChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = parseInt(e.target.value, 10);
      if (isNaN(value)) {
        return;
      }
      form.setState((prevState) => {
        return {
          ...prevState,
          webhookTimeout: value,
        } satisfies FormState;
      });
    },
    [form]
  );

  const isWebhookSecretMasked = form.state.webhookSecretKey == null;
  const webhookSecretKey = form.state.webhookSecretKey ?? "";

  return (
    <div className="flex flex-col gap-y-4">
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.webhook.url" />
        }
        value={form.state.webhookURL}
        required={true}
        onChange={onURLChange}
        disabled={isWebhookSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="url"
      />
      <CodeField
        label={renderToString(
          "SMSProviderConfigurationScreen.form.webhook.payload"
        )}
        description={renderToString(
          "SMSProviderConfigurationScreen.form.webhook.payload.description"
        )}
      >
        {`{
  "to": "+85298765432",
  "body": "You OTP is 123456"
}`}
      </CodeField>
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.webhook.signatureKey" />
        }
        value={isWebhookSecretMasked ? MASK : webhookSecretKey}
        readOnly={true}
        suffixPlain={true}
        suffix={
          isWebhookSecretMasked ? (
            <RevealIconButton onClick={onRevealSecrets} />
          ) : webhookSecretKey.length > 0 ? (
            <CopyIconButton textToCopy={webhookSecretKey} />
          ) : undefined
        }
        hint={
          <FormattedMessage
            id="SMSProviderConfigurationScreen.form.webhook.signatureKey.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://docs.authgear.com/customization/events-hooks/webhooks#verifying-signature">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        }
      />
      <TextField
        size="2"
        labelSize="2"
        type="number"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.webhook.timeout" />
        }
        hint={renderToString(
          "SMSProviderConfigurationScreen.form.webhook.timeout.description"
        )}
        value={String(form.state.webhookTimeout)}
        onChange={onTimeoutChange}
        disabled={isWebhookSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="timeout"
      />
    </div>
  );
}

function DenoHookForm({ form }: { form: FormModel }) {
  const { renderToString } = useContext(MessageContext);
  const { state, setState } = form;

  const onTimeoutChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = parseInt(e.target.value, 10);
      if (isNaN(value)) {
        return;
      }
      setState((prevState) => {
        return {
          ...prevState,
          denoHookTimeout: value,
        } satisfies FormState;
      });
    },
    [setState]
  );

  const resourceIdx = useDenoScriptResourceIndex(form.state);

  // Generate a new script resource if one does not exist
  useEffect(() => {
    if (state.providerType !== SMSProviderType.Deno || resourceIdx !== -1) {
      return;
    }
    setState((prevState) => {
      return produce(prevState, (prevState) => {
        prevState.denoHookURL = makeNewDenoScriptURL();
        const path = getDenoScriptPathFromURL(prevState.denoHookURL);
        const specifier = makeDenoScriptSpecifier(prevState.denoHookURL);
        const r = prevState.resources.find((r) => r.path === path);
        if (r == null) {
          prevState.resources.push({
            path,
            specifier,
            nullableValue: DEFAULT_SMS_SCRIPT_TEMPLATE,
          });
        }
      });
    });
  }, [
    resourceIdx,
    setState,
    state.denoHookURL,
    state.providerType,
    state.resources,
  ]);

  const onChangeCode = useCallback(
    (newValue?: string) => {
      if (newValue == null) {
        return;
      }
      if (resourceIdx === -1) {
        return;
      }
      setState((prevState) =>
        produce(prevState, (prevState) => {
          prevState.resources[resourceIdx].nullableValue = newValue;
        })
      );
    },
    [resourceIdx, setState]
  );

  return (
    <div className="flex flex-col gap-y-4">
      <div className="border border-[var(--gray-5)] rounded-xl overflow-hidden">
        <div className="px-6 py-4 border-b border-[var(--gray-5)]">
          <Text as="p" size="3" weight="medium">
            <FormattedMessage id="SMSProviderConfigurationScreen.form.deno.script" />
          </Text>
        </div>
        <div className="px-0 py-0">
          <CodeEditor
            className="block h-[412px]"
            language="typescript"
            value={
              resourceIdx !== -1
                ? state.resources[resourceIdx].nullableValue ?? ""
                : ""
            }
            onChange={onChangeCode}
            options={CODE_EDITOR_OPTIONS}
          />
        </div>
      </div>
      <TextField
        size="2"
        labelSize="2"
        type="number"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.deno.timeout" />
        }
        hint={renderToString(
          "SMSProviderConfigurationScreen.form.deno.timeout.description"
        )}
        value={String(form.state.denoHookTimeout)}
        onChange={onTimeoutChange}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="timeout"
      />
    </div>
  );
}
