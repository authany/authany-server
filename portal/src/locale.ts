import React, { useContext } from "react";

// Authany i18n: locale selection for the portal UI.
// Priority: explicit choice (localStorage) > browser languages > "en".

export const LOCALE_STORAGE_KEY = "authany.portal.locale";

export const LOCALE_DISPLAY_NAMES: Record<string, string> = {
  en: "English",
  "zh-CN": "简体中文",
  "zh-TW": "繁體中文",
  "zh-HK": "繁體中文（香港）",
  ja: "日本語",
  ko: "한국어",
};

function normalize(tag: string): string {
  const t = tag.trim();
  const lower = t.toLowerCase();
  if (lower.startsWith("zh")) {
    if (/hant|tw|hk|mo/.test(lower)) {
      return lower.includes("hk") ? "zh-HK" : "zh-TW";
    }
    return "zh-CN";
  }
  return lower.split("-")[0];
}

export function pickLocale(
  available: string[],
  browserLanguages: readonly string[],
  stored: string | null
): string {
  if (stored != null && available.includes(stored)) {
    return stored;
  }
  for (const lang of browserLanguages) {
    if (available.includes(lang)) return lang;
    const n = normalize(lang);
    if (available.includes(n)) return n;
    // zh-TW ↔ zh-HK are mutually acceptable
    if (n === "zh-TW" && available.includes("zh-HK")) return "zh-HK";
    if (n === "zh-HK" && available.includes("zh-TW")) return "zh-TW";
    const base = n.split("-")[0];
    const sameBase = available.find((a) => a.split("-")[0] === base);
    if (sameBase != null) return sameBase;
  }
  return "en";
}

export function readStoredLocale(): string | null {
  try {
    return window.localStorage.getItem(LOCALE_STORAGE_KEY);
  } catch {
    return null;
  }
}

export function writeStoredLocale(locale: string): void {
  try {
    window.localStorage.setItem(LOCALE_STORAGE_KEY, locale);
  } catch {
    // ignore
  }
}

export interface LocaleContextValue {
  locale: string;
  availableLocales: string[];
  setLocale: (locale: string) => void;
}

export const LocaleContext = React.createContext<LocaleContextValue>({
  locale: "en",
  availableLocales: ["en"],
  setLocale: () => {},
});

export function useLocale(): LocaleContextValue {
  return useContext(LocaleContext);
}
