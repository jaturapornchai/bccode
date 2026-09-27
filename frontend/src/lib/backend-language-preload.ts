import { serverGoApiBase } from "./backend-url";
import { normalizeLanguage, type LanguageCode } from "./i18n";
import { sanitizeBackendLanguageDictionary } from "./backend-language-sanitize";
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

// Server-side (SSR preload) only. The host is always BCAI_LOCAL_BACKEND_URL, never the client's
// backend_url cookie: a cookie-chosen host let any visitor point this server fetch at an arbitrary
// http(s) URL (SSRF) and hairpinned it through the public /backend proxy. Missing env = no preload
// ({}), the client hook then loads the dictionary through the /api/language BFF route.
export async function loadBackendLanguageDictionary(
  language: LanguageCode,
  fetcher: Fetcher = fetch,
): Promise<BackendLanguageDictionary> {
  let goApiBase: string;
  try {
    goApiBase = serverGoApiBase();
  } catch {
    return {};
  }

  const normalizedLanguage = normalizeLanguage(language);
  try {
    const response = await fetcher(`${goApiBase}/api/language/${encodeURIComponent(normalizedLanguage)}`, {
      cache: "no-store",
    });
    const payload = await response.json().catch(() => ({})) as unknown;
    if (!response.ok || !isDictionaryPayload(payload)) return {};
    return sanitizeBackendLanguageDictionary(payload);
  } catch {
    return {};
  }
}

function isDictionaryPayload(payload: unknown): payload is BackendLanguageDictionary {
  return Boolean(payload) && typeof payload === "object" && !Array.isArray(payload);
}
