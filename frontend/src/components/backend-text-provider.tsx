"use client";

import { createContext, useCallback, useContext, type ReactNode } from "react";
import { backendText, type BackendLanguageDictionary } from "@/lib/backend-language";

// Screen text must follow the selected language (AGENTS.md rule 2026-09-14): code carries
// English keys, the texts live in backend/assets/language/languages.tsv. Screens that already
// hold the dictionary (system-settings-screen.tsx) wrap their child views with this provider
// so nested components can call `const tr = useBackendText()` without prop drilling.
export type BackendTextFn = (key: string, fallback: string) => string;

const BackendTextContext = createContext<BackendLanguageDictionary>({});

export function BackendTextProvider({ dictionary, children }: { dictionary: BackendLanguageDictionary; children: ReactNode }) {
  return <BackendTextContext.Provider value={dictionary}>{children}</BackendTextContext.Provider>;
}

export function useBackendText(): BackendTextFn {
  const dictionary = useContext(BackendTextContext);
  return useCallback((key: string, fallback: string) => backendText(dictionary, key, fallback), [dictionary]);
}

// For helpers that already take a dictionary (thailandAddressUi, postalAddressHint)
// and are shared between the settings screen and the company/branch tree view.
export function useBackendDictionary(): BackendLanguageDictionary {
  return useContext(BackendTextContext);
}
