"use client";

import { useEffect, useState } from "react";
import { Check, Type, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useBackendText } from "@/components/backend-text-provider";
import { type LanguageCode } from "@/lib/i18n";
import {
  appFonts,
  applyAppFont,
  defaultFontId,
  ensureFontLoaded,
  fontStorageKey,
  normalizeAppFont,
  writeFontCookie,
  type FontId,
} from "@/lib/font-data";

interface ScreenFontSettingsProps {
  language?: LanguageCode;
}

/**
 * เมนู หน้าต่าง › เลือกตัวอักษรสำหรับหน้าจอทั่วไป (/settings/font) — ฟอนต์หน้าจอมีผลจริงทันทีผ่าน cookie + CSS (applyAppFont).
 * ฟอนต์รายงาน 5 เมนูเดิมเก็บใน localStorage แต่ไม่มีโค้ดใดอ่านไปใช้ จึงถอดเป็น "รอพัฒนา" จนกว่าจะเก็บค่าต่อบริษัทใน backend
 * (docs/kms/champ/design-2026-10-11/facade-screens.md D-6).
 */
export function ScreenFontSettings({ language: _language = "th" }: ScreenFontSettingsProps) {
  const tr = useBackendText();
  const [selectedFont, setSelectedFont] = useState<FontId>(defaultFontId);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const saved = localStorage.getItem(fontStorageKey);
    setSelectedFont(saved ? normalizeAppFont(saved) : defaultFontId);
    setReady(true);
    appFonts.forEach((font) => ensureFontLoaded(font));
  }, []);

  const handleSelectFont = (fontId: FontId) => {
    setSelectedFont(fontId);
    localStorage.setItem(fontStorageKey, fontId);
    applyAppFont(fontId);
    writeFontCookie(fontId);
  };

  const handleResetDefault = () => {
    handleSelectFont(defaultFontId);
  };

  const [activeTheme, setActiveTheme] = useState<"light" | "dark" | "system">("light");

  useEffect(() => {
    const isDark = document.documentElement.classList.contains("dark");
    setActiveTheme(isDark ? "dark" : "light");
  }, []);

  const handleSetTheme = (theme: "light" | "dark" | "system") => {
    setActiveTheme(theme);
    if (theme === "dark") {
      document.documentElement.classList.add("dark");
      localStorage.setItem("bc_theme", "dark");
    } else if (theme === "light") {
      document.documentElement.classList.remove("dark");
      localStorage.setItem("bc_theme", "light");
    } else {
      const prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
      document.documentElement.classList.toggle("dark", prefersDark);
      localStorage.setItem("bc_theme", "system");
    }
  };

  return (
    <div className="flex-1 min-h-0 flex flex-col gap-4 p-4 lg:p-6 overflow-auto">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-5 shadow-xs">
        <div className="flex items-center gap-3.5">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-inner">
            <Type className="h-4 w-4" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">
              {tr("window_font_screen", "เลือกตัวอักษรสำหรับหน้าจอทั่วไป")}
            </h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              {tr(
                "font_screen_description",
                "เลือกรูปแบบตัวอักษร (Typography) สำหรับเมนู ฟอร์ม และหน้าจอการทำงานทั่วไปในระบบ การตั้งค่ามีผลทันที",
              )}
            </p>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleResetDefault}
            className="gap-2 text-xs"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            {tr("reset_default_font", "คืนค่าเริ่มต้น")}
          </Button>
        </div>
      </div>

      {/* Theme Quick Bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 px-4 py-2.5 rounded-xl border border-border/80 bg-card/60 text-xs">
        <div className="flex items-center gap-2">
          <span className="font-semibold text-foreground">{tr("font_theme_mode_label", "ธีมสีของหน้าจอ (Theme Mode):")}</span>
          <span className="text-muted-foreground">{tr("font_theme_mode_hint", "ปรับโทนสีแสดงผลควบคู่กับตัวอักษร")}</span>
        </div>
        <div className="flex items-center gap-1.5">
          <button
            onClick={() => handleSetTheme("light")}
            className={`px-3 py-1 rounded-md font-medium transition-colors ${
              activeTheme === "light"
                ? "bg-primary text-primary-foreground shadow-xs"
                : "bg-muted text-muted-foreground hover:text-foreground"
            }`}
          >
            {tr("font_theme_light", "สว่าง (Light)")}
          </button>
          <button
            onClick={() => handleSetTheme("dark")}
            className={`px-3 py-1 rounded-md font-medium transition-colors ${
              activeTheme === "dark"
                ? "bg-primary text-primary-foreground shadow-xs"
                : "bg-muted text-muted-foreground hover:text-foreground"
            }`}
          >
            {tr("font_theme_dark", "มืด (Dark)")}
          </button>
          <button
            onClick={() => handleSetTheme("system")}
            className={`px-3 py-1 rounded-md font-medium transition-colors ${
              activeTheme === "system"
                ? "bg-primary text-primary-foreground shadow-xs"
                : "bg-muted text-muted-foreground hover:text-foreground"
            }`}
          >
            {tr("font_theme_system", "ตามระบบ (Auto)")}
          </button>
        </div>
      </div>

      {/* Font Cards Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {appFonts.map((font) => {
          const isSelected = ready && selectedFont === font.id;
          return (
            <div
              key={font.id}
              onClick={() => handleSelectFont(font.id)}
              className={`group relative flex flex-col justify-between rounded-2xl border p-5 transition-all cursor-pointer select-none ${
                isSelected
                  ? "border-primary bg-primary/5 shadow-md ring-2 ring-primary/20"
                  : "border-border/70 bg-card hover:border-primary/50 hover:bg-muted/40 shadow-xs"
              }`}
            >
              <div>
                <div className="flex items-start justify-between gap-2 mb-3">
                  <div>
                    <h3 className="font-semibold text-base text-foreground flex items-center gap-1.5">
                      {font.short}
                      {isSelected && (
                        <span className="inline-flex items-center rounded-full bg-primary px-2 py-0.5 text-[11px] font-medium text-primary-foreground shadow-xs">
                          {tr("active_badge", "ใช้งานอยู่")}
                        </span>
                      )}
                    </h3>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {font.name.split("—")[1]?.trim() || ""}
                    </p>
                  </div>
                  {isSelected && (
                    <div className="flex h-5 w-5 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-xs">
                      <Check className="h-3 w-3 stroke-[3]" />
                    </div>
                  )}
                </div>

                {/* Live Preview Box — ข้อความตัวอย่างไทย/อังกฤษ/ตัวเลขตั้งใจคงเป็นตัวอักษรตรง ๆ เพื่อโชว์รูปอักษรของฟอนต์ ไม่ใช่ข้อความ UI */}
                <div
                  style={{ fontFamily: font.family }}
                  className="rounded-xl border border-border/50 bg-background/80 p-3.5 text-foreground transition-colors group-hover:border-border"
                >
                  <div className="text-sm font-medium leading-relaxed">
                    ระบบบัญชีแยกประเภทและการเงิน
                  </div>
                  <div className="text-xs text-muted-foreground mt-1">
                    General Ledger &amp; Financial ERP
                  </div>
                  <div className="text-xs font-semibold text-primary mt-2">
                    1,250,800.50 บาท (THB)
                  </div>
                </div>
              </div>

              <div className="mt-4 flex items-center justify-between pt-2 border-t border-border/40 text-[11px] text-muted-foreground">
                <span>{tr("font_compat_note", "ความเข้ากันได้: ไทย + สากล")}</span>
                <span className="font-mono">{font.id}</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
