import { cookies } from "next/headers";
import { normalizeLanguage, type LanguageCode } from "./i18n";
import type { BackendLanguageDictionary } from "./backend-language";
import {
  backendUrlPreferenceCookie,
  languagePreferenceCookie,
  loadBackendLanguageDictionary,
} from "./backend-language-preload";

export type InitialBackendLanguage = {
  initialBackendLanguage: BackendLanguageDictionary;
  initialBackendUrl?: string;
  initialLanguage: LanguageCode;
};

export async function getInitialBackendLanguage(): Promise<InitialBackendLanguage> {
  const cookieStore = await cookies();
  const initialLanguage = normalizeLanguage(cookieStore.get(languagePreferenceCookie)?.value ?? "th");
  // Handed back to the client screens only (their fallback backend URL); the server never fetches it.
  const initialBackendUrl = cookieStore.get(backendUrlPreferenceCookie)?.value;
  const initialBackendLanguage = await loadBackendLanguageDictionary(initialLanguage);

  return {
    initialBackendLanguage,
    initialBackendUrl,
    initialLanguage,
  };
}
