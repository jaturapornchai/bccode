"use client";

import { useEffect, useState } from "react";
import { sanitizeBackendLanguageDictionary, sanitizeBackendLanguageText } from "./backend-language-sanitize";
import { persistLanguagePreferenceCookies } from "./backend-language-preload";
import type { LanguageCode } from "./i18n";

export type BackendLanguageDictionary = Record<string, string>;

export type BackendLanguageState = {
  dictionary: BackendLanguageDictionary;
  error: Error | null;
  hasCache: boolean;
  isLoading: boolean;
  isReady: boolean;
};

const readyMarker = "__backendLanguageReady";
const loadingMarker = "__backendLanguageLoading";
const hasCacheMarker = "__backendLanguageHasCache";
const memoryCache = new Map<string, BackendLanguageDictionary>();

function cacheKeyFor(language: LanguageCode, backendUrl: string): string {
  return `${backendUrl}::${language}`;
}

function storageKeyFor(cacheKey: string): string {
  return `bc_backend_language::${cacheKey}`;
}

function canUseBrowserStorage(): boolean {
  return typeof window !== "undefined" && Boolean(window.localStorage);
}

function readStoredDictionary(cacheKey: string): BackendLanguageDictionary | null {
  if (!canUseBrowserStorage()) return null;
  try {
    const raw = window.localStorage.getItem(storageKeyFor(cacheKey));
    if (!raw) return null;
    const parsed = JSON.parse(raw) as unknown;
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) return null;
    return sanitizeBackendLanguageDictionary(parsed as BackendLanguageDictionary);
  } catch {
    return null;
  }
}

function writeStoredDictionary(cacheKey: string, dictionary: BackendLanguageDictionary) {
  if (!canUseBrowserStorage()) return;
  try {
    window.localStorage.setItem(storageKeyFor(cacheKey), JSON.stringify(sanitizeBackendLanguageDictionary(dictionary)));
  } catch {
    // Cache writes are best-effort; the UI can still use the in-memory dictionary.
  }
}

function withLanguageMeta(
  dictionary: BackendLanguageDictionary,
  meta: Pick<BackendLanguageState, "hasCache" | "isLoading" | "isReady">,
): BackendLanguageDictionary {
  Object.defineProperties(dictionary, {
    [readyMarker]: { configurable: true, enumerable: false, value: String(meta.isReady) },
    [loadingMarker]: { configurable: true, enumerable: false, value: String(meta.isLoading) },
    [hasCacheMarker]: { configurable: true, enumerable: false, value: String(meta.hasCache) },
  });
  return dictionary;
}

function stateFromDictionary(
  dictionary: BackendLanguageDictionary,
  meta: Pick<BackendLanguageState, "hasCache" | "isLoading" | "isReady">,
  error: Error | null = null,
): BackendLanguageState {
  return {
    dictionary: withLanguageMeta(dictionary, meta),
    error,
    hasCache: meta.hasCache,
    isLoading: meta.isLoading,
    isReady: meta.isReady,
  };
}

function cachedDictionary(language: LanguageCode, backendUrl: string | undefined): BackendLanguageDictionary | null {
  if (!backendUrl) return null;
  const cacheKey = cacheKeyFor(language, backendUrl);
  const memory = memoryCache.get(cacheKey);
  if (memory) return { ...memory };
  const stored = readStoredDictionary(cacheKey);
  if (!stored) return null;
  memoryCache.set(cacheKey, sanitizeBackendLanguageDictionary(stored));
  return { ...stored };
}

function initialState(
  language: LanguageCode,
  backendUrl: string | undefined,
  initialDictionary?: BackendLanguageDictionary,
): BackendLanguageState {
  if (initialDictionary && Object.keys(initialDictionary).length > 0) {
    const sanitized = sanitizeBackendLanguageDictionary(initialDictionary);
    if (backendUrl) memoryCache.set(cacheKeyFor(language, backendUrl), sanitized);
    return stateFromDictionary({ ...sanitized }, { hasCache: true, isLoading: false, isReady: true });
  }
  const cached = cachedDictionary(language, backendUrl);
  if (cached) return stateFromDictionary(cached, { hasCache: true, isLoading: false, isReady: true });
  return stateFromDictionary({}, { hasCache: false, isLoading: false, isReady: false });
}

export function useBackendLanguageState(
  language: LanguageCode,
  backendUrl: string | undefined,
  initialDictionary?: BackendLanguageDictionary,
): BackendLanguageState {
  const [state, setState] = useState<BackendLanguageState>(() => initialState(language, backendUrl, initialDictionary));

  useEffect(() => {
    persistLanguagePreferenceCookies(language, backendUrl);

    if (!backendUrl) {
      setState(stateFromDictionary({}, { hasCache: false, isLoading: false, isReady: false }));
      return;
    }

    const cacheKey = cacheKeyFor(language, backendUrl);
    const cached = cachedDictionary(language, backendUrl);
    setState((current) => {
      if (cached) return stateFromDictionary(cached, { hasCache: true, isLoading: true, isReady: true });
      if (current.isReady && current.hasCache) {
        return stateFromDictionary({ ...current.dictionary }, { hasCache: true, isLoading: true, isReady: true });
      }
      return stateFromDictionary({}, { hasCache: false, isLoading: true, isReady: false });
    });

    const controller = new AbortController();
    void fetch(`/api/language/${language}?backendUrl=${encodeURIComponent(backendUrl)}`, {
      cache: "no-store",
      signal: controller.signal,
    })
      .then((response) => {
        if (!response.ok) throw new Error("language load failed");
        return response.json() as Promise<BackendLanguageDictionary>;
      })
      .then((data) => {
        const sanitized = sanitizeBackendLanguageDictionary(data);
        memoryCache.set(cacheKey, sanitized);
        writeStoredDictionary(cacheKey, sanitized);
        setState(stateFromDictionary({ ...sanitized }, { hasCache: true, isLoading: false, isReady: true }));
      })
      .catch((error: unknown) => {
        if (controller.signal.aborted) return;
        const nextError = error instanceof Error ? error : new Error("language load failed");
        setState((current) => {
          if (current.isReady) {
            return stateFromDictionary({ ...current.dictionary }, { hasCache: current.hasCache, isLoading: false, isReady: true }, nextError);
          }
          return stateFromDictionary({}, { hasCache: false, isLoading: false, isReady: false }, nextError);
        });
      });

    return () => controller.abort();
  }, [backendUrl, language]);

  return state;
}

export function useBackendLanguage(
  language: LanguageCode,
  backendUrl: string | undefined,
  initialDictionary?: BackendLanguageDictionary,
): BackendLanguageDictionary {
  return useBackendLanguageState(language, backendUrl, initialDictionary).dictionary;
}

export function isBackendLanguageReady(dictionary: BackendLanguageDictionary | undefined): boolean {
  return dictionary?.[readyMarker] === "true";
}

export function isBackendLanguageLoading(dictionary: BackendLanguageDictionary | undefined): boolean {
  return dictionary?.[loadingMarker] === "true";
}

export function backendText(dictionary: BackendLanguageDictionary, key: string, fallback?: string): string {
  const text = dictionary[key];
  if (text) return sanitizeBackendLanguageText(text);
  if (fallback !== undefined) return sanitizeBackendLanguageText(fallback);
  if (!isBackendLanguageReady(dictionary)) return "";
  return key;
}
