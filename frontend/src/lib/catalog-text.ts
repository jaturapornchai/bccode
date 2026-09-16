// Catalog strings (work-tab screen titles, report columns, tool steps) ship as
// `{ th, en }` literals next to their route config. This resolves them through
// languages.tsv so all twelve languages work, keeping the literal as the
// offline fallback — same shape as `fieldLabel()` in system-settings/utils.ts.

import { backendText, type BackendLanguageDictionary } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";

export type BilingualText = { th: string; en: string };

export function catalogText(
  keys: Record<string, string>,
  id: string,
  text: BilingualText | undefined,
  language: LanguageCode,
  dictionary?: BackendLanguageDictionary,
): string {
  const fallback = text ? (language === "th" ? text.th : text.en) : "";
  const backendKey = keys[id];
  if (!backendKey || !dictionary) return fallback;
  const value = backendText(dictionary, backendKey, fallback);
  return value === backendKey ? fallback : value;
}
