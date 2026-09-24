"use client";

// Standard app header controls — shared across every screen so users always
// see the same Font / Theme / Language buttons in the same place.
// All three are 42x42 icon buttons. Language shows only the current flag.
// Drop <AppHeaderControls language={language} onLanguageChange={setLanguage} />
// into any screen header.

import { type LanguageCode } from "@/lib/i18n";
import { FontPicker } from "./font-picker";
import { ThemeToggle } from "./theme-toggle";
import { LanguageDialog } from "./language-dialog";

export function AppHeaderControls({
  language,
  onLanguageChange,
}: {
  language: LanguageCode;
  onLanguageChange: (next: LanguageCode) => void;
}) {
  return (
    <div className="header-controls">
      <FontPicker language={language} />
      <ThemeToggle language={language} />
      <LanguageDialog language={language} onLanguageChange={onLanguageChange} />
    </div>
  );
}
