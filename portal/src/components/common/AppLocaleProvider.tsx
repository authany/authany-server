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
  // Authany i18n: locales available = "en" + whatever translations.json provides.
  const availableLocales = useMemo(() => {
    const keys = Object.keys(systemConfig?.translations ?? {}).filter(
      (k) => (systemConfig?.translations[k] ?? undefined) != null
    );
    return Array.from(new Set(["en", ...keys]));
  }, [systemConfig]);

  const [locale, setLocaleState] = useState<string>(() =>
    pickLocale(availableLocales, navigator.languages ?? [], readStoredLocale())
  );

  useEffect(() => {
    // Re-evaluate once the system config (and thus the locale list) arrives.
    setLocaleState((current) =>
      availableLocales.includes(current)
        ? current
        : pickLocale(availableLocales, navigator.languages ?? [], readStoredLocale())
    );
  }, [availableLocales]);

  const setLocale = useCallback((next: string) => {
    writeStoredLocale(next);
    setLocaleState(next);
  }, []);

  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);

  // Fallback chain: bundled en.json → server-side en overrides → chosen locale.
  const messages = useMemo(() => {
    const en = systemConfig?.translations.en ?? {};
    const chosen = locale === "en" ? {} : systemConfig?.translations[locale] ?? {};
    return { ...DEFAULT_MESSAGES, ...en, ...chosen };
  }, [systemConfig, locale]);

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
