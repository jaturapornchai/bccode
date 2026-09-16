"use client";

import { useMemo } from "react";

import { useBackendDictionary } from "@/components/backend-text-provider";
import { getBarcodeText, type BarcodeText } from "@/lib/product-barcode/language";
import type { LanguageCode } from "@/lib/i18n";

// `getBarcodeText(language)` without a dictionary only knows Thai and English, so every
// other language silently read the English column even though languages.tsv carries all
// twelve (AGENTS.md rule 2026-09-14). Screens wrap their tree in <BackendTextProvider>
// and the components below just call this hook — no prop drilling through the tabs.
export function useBarcodeText(language: LanguageCode | string | undefined): BarcodeText {
  const dictionary = useBackendDictionary();
  return useMemo(() => getBarcodeText(language, dictionary), [language, dictionary]);
}
