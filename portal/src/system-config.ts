import MESSAGES from "./locale-data/en.json";
import { DEFAULT_TEMPLATE_LOCALE } from "./resources";

export interface SystemConfig {
  authgearAppID: string;
  authgearClientID: string;
  authgearEndpoint: string;
  authgearWebSDKSessionType: "cookie" | "refresh_token";
  isAuthgearOnce: boolean;
  authgearOnceLicenseKey: string;
  authgearOnceLicenseExpireAt: string;
  authgearOnceLicenseeEmail: string;
  sentryDSN: string;
  appHostSuffix: string;
  availableLanguages: string[];
  builtinLanguages: string[];
  translations: SystemConfigTranslations;
  searchEnabled: boolean;
  auditLogEnabled: boolean;
  gitCommitHash: string;
  analyticEnabled: boolean;
  analyticEpoch: string;
  gtmContainerID: string;
  uiImplementation: string;
  uiSettingsImplemenation: string;
}

// Authany i18n: en is always present; other locales (zh-CN, ...) come from
// PORTAL_CUSTOM_RESOURCE_DIRECTORY/translations.json and are merged over en at runtime.
export interface SystemConfigTranslations {
  en: Record<string, string>;
  [locale: string]: Record<string, string> | undefined;
}

export interface PartialSystemConfig
  extends Partial<Omit<SystemConfig, "translations">> {
  translations?: Partial<SystemConfigTranslations>;
}

export const defaultSystemConfig: PartialSystemConfig = {
  translations: {
    en: MESSAGES,
  },
};

export function mergeSystemConfig(
  baseConfig: PartialSystemConfig,
  overlayConfig: PartialSystemConfig
): PartialSystemConfig {
  return {
    ...baseConfig,
    ...overlayConfig,
    translations: mergeTranslations(
      baseConfig.translations,
      overlayConfig.translations
    ),
  };
}

function mergeTranslations(
  base?: Partial<SystemConfigTranslations>,
  overlay?: Partial<SystemConfigTranslations>
): SystemConfigTranslations {
  const locales = new Set<string>(["en"]);
  for (const k of Object.keys(base ?? {})) locales.add(k);
  for (const k of Object.keys(overlay ?? {})) locales.add(k);
  const out: SystemConfigTranslations = { en: {} };
  for (const locale of locales) {
    out[locale] = {
      ...(base?.[locale] ?? {}),
      ...(overlay?.[locale] ?? {}),
    };
  }
  return out;
}

export function instantiateSystemConfig(
  config: PartialSystemConfig
): SystemConfig {
  return {
    authgearAppID: config.authgearAppID ?? "",
    authgearClientID: config.authgearClientID ?? "",
    authgearEndpoint: config.authgearEndpoint ?? "",
    authgearWebSDKSessionType: config.authgearWebSDKSessionType ?? "cookie",
    isAuthgearOnce: config.isAuthgearOnce ?? false,
    authgearOnceLicenseKey: config.authgearOnceLicenseKey ?? "",
    authgearOnceLicenseExpireAt: config.authgearOnceLicenseExpireAt ?? "",
    authgearOnceLicenseeEmail: config.authgearOnceLicenseeEmail ?? "",
    sentryDSN: config.sentryDSN ?? "",
    appHostSuffix: config.appHostSuffix ?? "",
    availableLanguages: config.availableLanguages ?? [DEFAULT_TEMPLATE_LOCALE],
    builtinLanguages: config.builtinLanguages ?? [DEFAULT_TEMPLATE_LOCALE],
    translations: mergeTranslations({ en: {} }, config.translations),
    searchEnabled: config.searchEnabled ?? false,
    auditLogEnabled: config.auditLogEnabled ?? false,
    gitCommitHash: config.gitCommitHash ?? "",
    analyticEnabled: config.analyticEnabled ?? false,
    analyticEpoch: config.analyticEpoch ?? "",
    gtmContainerID: config.gtmContainerID ?? "",
    uiImplementation: config.uiImplementation ?? "authflowv2",
    uiSettingsImplemenation: config.uiSettingsImplemenation ?? "v2",
  };
}
