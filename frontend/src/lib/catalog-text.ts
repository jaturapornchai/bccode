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

// Some screens keep a whole `{ th: {...}, en: {...} }` table of their own. This
// resolves every property in one pass so the screen keeps using `text.title`
// exactly as before, but the words now come from languages.tsv.
export function resolveTextTable<T extends Record<string, string>>(
  tables: { th: T; en: Record<string, string> },
  keys: Record<string, string>,
  language: LanguageCode,
  dictionary?: BackendLanguageDictionary,
): Record<keyof T, string> {
  const resolved: Record<string, string> = {};
  for (const prop of Object.keys(tables.th)) {
    resolved[prop] = catalogText(
      keys,
      prop,
      { th: tables.th[prop] ?? "", en: tables.en[prop] ?? "" },
      language,
      dictionary,
    );
  }
  return resolved as Record<keyof T, string>;
}
