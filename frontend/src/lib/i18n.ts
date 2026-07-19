// BC Ai Account frontend i18n.
// Source of truth: src/locales/th.json. Other languages live in src/locales/<lang>.json.
// Adding a new string: edit src/locales/th.json first (Thai-First No Auto-Translate Iron Rule),
// then copy the Thai value as a placeholder into the other locale files. Jead commands translation passes.

import th from "@/locales/th.json";
import en from "@/locales/en.json";
import cn from "@/locales/cn.json";
import ja from "@/locales/ja.json";
import ko from "@/locales/ko.json";
import lo from "@/locales/lo.json";
import my from "@/locales/my.json";
import km from "@/locales/km.json";
import vi from "@/locales/vi.json";
import ms from "@/locales/ms.json";
import id from "@/locales/id.json";
import fil from "@/locales/fil.json";

export const LANGUAGES = [
  { code: "th", name: "ภาษาไทย" },
  { code: "en", name: "English" },
  { code: "cn", name: "中文" },
  { code: "ja", name: "日本語" },
  { code: "ko", name: "한국어" },
  { code: "lo", name: "ພາສາລາວ" },
  { code: "my", name: "မြန်မာဘာသာ" },
  { code: "km", name: "ភាសាខ្មែរ" },
  { code: "vi", name: "Tiếng Việt" },
  { code: "ms", name: "Bahasa Melayu" },
  { code: "id", name: "Bahasa Indonesia" },
  { code: "fil", name: "Filipino" },
] as const;

export type LanguageCode = (typeof LANGUAGES)[number]["code"];

// TranslationKey is derived from the Thai dictionary (source of truth) so there is no
// hand-maintained union to drift out of sync. Adding a key to th.json makes it available
// to t() across all languages; the drift test in i18n.test.ts catches missing keys.
export type TranslationKey = keyof typeof th;

const translations: Record<LanguageCode, Partial<Record<TranslationKey, string>>> = {
  th,
  en,
  cn,
  ja,
  ko,
  lo,
  my,
  km,
  vi,
  ms,
  id,
  fil,
};

const defaultLanguage: LanguageCode = "th";

export function normalizeLanguage(value: string | null | undefined): LanguageCode {
  const normalized = (value ?? "").toLowerCase().split("-")[0];
  if (normalized === "zh") return "cn";
  if (normalized === "tl") return "fil";
  return LANGUAGES.some((language) => language.code === normalized)
    ? (normalized as LanguageCode)
    : defaultLanguage;
}

export function t(language: LanguageCode, key: TranslationKey): string {
  return translations[language]?.[key] ?? translations[defaultLanguage]?.[key] ?? key;
}
