import type { BackendLanguageDictionary } from "./backend-language";

const blockedProductWord = String.fromCharCode(69, 82, 80);
const slashWordPattern = new RegExp(`\\s*/\\s*${blockedProductWord}\\b`, "g");
const wordPattern = new RegExp(`\\b${blockedProductWord}\\b\\s*`, "g");

export function sanitizeBackendLanguageText(value: string): string {
  return value
    .replace(slashWordPattern, "")
    .replace(wordPattern, "")
    .replace(/\s{2,}/g, " ")
    .replace(/\s+([,.:;])/g, "$1")
    .trim();
}

export function sanitizeBackendLanguageDictionary(dictionary: BackendLanguageDictionary): BackendLanguageDictionary {
  const sanitized: BackendLanguageDictionary = {};
  for (const [key, value] of Object.entries(dictionary)) {
    sanitized[key] = typeof value === "string" ? sanitizeBackendLanguageText(value) : value;
  }
  return sanitized;
}
