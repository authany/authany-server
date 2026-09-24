import { IntlProvider } from "react-intl";
import React, { useCallback, useEffect, useMemo, useState } from "react";
import DEFAULT_MESSAGES from "../../locale-data/en.json";
import { SystemConfig } from "../../system-config";
import { IntlContextProvider } from "../../intl";
import {
  LocaleContext,
  pickLocale,
  readStoredLocale,
  writeStoredLocale,
} from "../../locale";

const defaultRichTextElements = {
  br: () => <br />,
  b: (children: React.ReactNode) => <b>{children}</b>,
  strong: (children: React.ReactNode) => <strong>{children}</strong>,
  code: (children: React.ReactNode) => (
    <code className="inline-code">{children}</code>
  ),
  pre: (children: React.ReactNode) => <pre>{children}</pre>,
  small: (children: React.ReactNode) => <small>{children}</small>,
};

// This is to support legacy API that uses @oursky/react-messageformat
function _LegacyAPIContextProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  return <IntlContextProvider>{children}</IntlContextProvider>;
}

export function AppLocaleProvider({
  children,
  systemConfig,
}: {
  systemConfig?: SystemConfig;
  children: React.ReactNode;
}): React.ReactElement {
  const translations = systemConfig?.translations;

  // Authany i18n: locales available = "en" + the locales in translations.json
  // (the built-in resources/portal/translations.json, or a copy in
  // PORTAL_CUSTOM_RESOURCE_DIRECTORY, which replaces it whole).
  const availableLocales = useMemo(() => {
    const keys =
      translations != null
        ? Object.keys(translations).filter((k) => translations[k] != null)
        : [];
    return Array.from(new Set(["en", ...keys]));
  }, [translations]);

  // Only the explicit choice (localStorage) is state. The effective locale is
  // derived from it, so it re-resolves by itself once the system config (and
  // thus the locale list) arrives, without syncing state in an effect.
  const [chosenLocale, setChosenLocale] = useState(readStoredLocale);

  const locale = useMemo(
    () => pickLocale(availableLocales, navigator.languages, chosenLocale),
    [availableLocales, chosenLocale]
  );

  const setLocale = useCallback((next: string) => {
    writeStoredLocale(next);
    setChosenLocale(next);
  }, []);

  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);

  // Fallback chain: bundled en.json → server-side en overrides → chosen locale.
  const messages = useMemo(() => {
    const en = translations?.en ?? {};
    const chosen = locale === "en" ? {} : translations?.[locale] ?? {};
    return { ...DEFAULT_MESSAGES, ...en, ...chosen };
  }, [translations, locale]);

  const contextValue = useMemo(
    () => ({ locale, availableLocales, setLocale }),
    [locale, availableLocales, setLocale]
  );

  return (
    <LocaleContext.Provider value={contextValue}>
      <IntlProvider
        locale={locale}
        defaultLocale="en"
        messages={messages}
        defaultRichTextElements={defaultRichTextElements}
      >
        <_LegacyAPIContextProvider>{children}</_LegacyAPIContextProvider>
      </IntlProvider>
    </LocaleContext.Provider>
  );
}
