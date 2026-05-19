import { validateBackendUrl } from "./backend-url";
import { normalizeLanguage, type LanguageCode } from "./i18n";
import type { BackendLanguageDictionary } from "./backend-language";

export const languagePreferenceCookie = "user_language";
export const backendUrlPreferenceCookie = "backend_url";

const preferenceMaxAgeSeconds = 60 * 60 * 24 * 365;

type Fetcher = typeof fetch;

export function serializePreferenceCookie(name: string, value: string): string {
  return `${name}=${encodeURIComponent(value)}; Max-Age=${preferenceMaxAgeSeconds}; Path=/; SameSite=Lax`;
}

export function persistLanguagePreferenceCookies(language: LanguageCode, backendUrl?: string) {
  if (typeof document === "undefined") return;
  document.cookie = serializePreferenceCookie(languagePreferenceCookie, language);
  if (backendUrl?.trim()) {
    document.cookie = serializePreferenceCookie(backendUrlPreferenceCookie, backendUrl.trim());
  }
}

export async function loadBackendLanguageDictionary(
  language: LanguageCode,
  backendUrl: string | undefined,
  fetcher: Fetcher = fetch,
): Promise<BackendLanguageDictionary> {
  if (!backendUrl?.trim()) return {};

  let normalizedGoApiUrl: string;
  try {
    normalizedGoApiUrl = validateBackendUrl(backendUrl).normalizedGoApiUrl;
  } catch {
    return {};
  }

  const normalizedLanguage = normalizeLanguage(language);
  try {
    const response = await fetcher(`${normalizedGoApiUrl}/api/language/${encodeURIComponent(normalizedLanguage)}`, {
      cache: "no-store",
    });
    const payload = await response.json().catch(() => ({})) as unknown;
    if (!response.ok || !isDictionaryPayload(payload)) return {};
    return payload;
  } catch {
    return {};
  }
}

function isDictionaryPayload(payload: unknown): payload is BackendLanguageDictionary {
  return Boolean(payload) && typeof payload === "object" && !Array.isArray(payload);
}
