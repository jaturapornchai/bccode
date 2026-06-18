"use client";

// Standard app header controls — shared across every screen so users always
// see the same Font / Theme / Language / Settings buttons in the same place.
// All four are 42x42 icon buttons. Language shows only the current flag.
// Drop <AppHeaderControls language={language} onLanguageChange={setLanguage} />
// into any screen header; pass showSettings=false to hide the settings link
// on screens that are already the settings screen.

import Link from "next/link";
import { Settings as SettingsIcon } from "lucide-react";
import { t, type LanguageCode } from "@/lib/i18n";
import { FontPicker } from "./font-picker";
import { ThemeToggle } from "./theme-toggle";
import { LanguageDialog } from "./language-dialog";

export function AppHeaderControls({
  language,
  onLanguageChange,
  showSettings = true,
}: {
  language: LanguageCode;
  onLanguageChange: (next: LanguageCode) => void;
  showSettings?: boolean;
}) {
  return (
    <div className="header-controls">
      <FontPicker language={language} />
      <ThemeToggle language={language} />
      <LanguageDialog language={language} onLanguageChange={onLanguageChange} />
      {showSettings ? (
        <Link
          href="/settings"
          className="header-control-button"
          title={t(language, "settings")}
          aria-label={t(language, "settings")}
        >
          <SettingsIcon aria-hidden="true" size={18} />
        </Link>
      ) : null}
    </div>
  );
}
