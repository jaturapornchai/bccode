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

export function ScreenFontSettings({ language: _language = "th" }: ScreenFontSettingsProps) {
  const tr = useBackendText();
  const [currentFontId, setCurrentFontId] = useState<FontId>(defaultFontId);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const saved = normalizeAppFont(localStorage.getItem(fontStorageKey));
    setCurrentFontId(saved);
    setReady(true);
    appFonts.forEach((font) => ensureFontLoaded(font));
  }, []);

  const handleSelectFont = (fontId: FontId) => {
    setCurrentFontId(fontId);
    applyAppFont(fontId);
    localStorage.setItem(fontStorageKey, fontId);
    writeFontCookie(fontId);
  };

  const handleResetDefault = () => {
    handleSelectFont(defaultFontId);
  };

  return (
    <div className="flex-1 min-h-0 flex flex-col gap-5 p-4 lg:p-6 overflow-auto">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-5 shadow-sm">
        <div className="flex items-center gap-3.5">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-inner">
            <Type className="h-6 w-6" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">
              {tr("window_font_screen", "เลือกตัวอักษรสำหรับหน้าจอทั่วไป")}
            </h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              {tr("window_font_screen_desc", "เลือกรูปแบบตัวอักษร (Typography) ที่สบายตาและเหมาะสมกับสายตาของคุณ การตั้งค่ามีผลกับทุกหน้าจอในระบบทันที")}
            </p>
          </div>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={handleResetDefault}
          className="gap-2"
        >
          <RefreshCw className="h-4 w-4" />
          {tr("reset_default_font", "คืนค่าเริ่มต้น")}
        </Button>
      </div>

      {/* Font Cards Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {appFonts.map((font) => {
          const isSelected = ready && currentFontId === font.id;
          return (
            <div
              key={font.id}
              onClick={() => handleSelectFont(font.id)}
              className={`group relative flex flex-col justify-between rounded-2xl border p-5 transition-all cursor-pointer select-none ${
                isSelected
                  ? "border-primary bg-primary/5 shadow-md ring-2 ring-primary/20"
                  : "border-border/70 bg-card hover:border-primary/50 hover:bg-muted/40 shadow-sm"
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
                    <p className="text-xs text-muted-foreground mt-0.5">{font.name.split("—")[1]?.trim() || ""}</p>
                  </div>

                  <div
                    className={`flex size-6 shrink-0 items-center justify-center rounded-full border transition-colors ${
                      isSelected
                        ? "border-primary bg-primary text-primary-foreground"
                        : "border-border/80 group-hover:border-primary/60"
                    }`}
                  >
                    {isSelected && <Check className="size-3.5 stroke-[3]" />}
                  </div>
                </div>

                {/* Live Preview Box rendered in this font */}
                <div
                  className="rounded-xl border border-border/40 bg-background/80 p-3.5 mt-2 space-y-2 text-foreground"
                  style={{ fontFamily: font.family }}
                >
                  <div className="text-sm font-semibold tracking-wide">
                    ระบบบัญชีคู่ TFRS สำหรับ SMEs
                  </div>
                  <div className="text-xs text-muted-foreground leading-relaxed">
                    ผังบัญชี · สมุดรายวันทั่วไป · ภาษีมูลค่าเพิ่ม ภ.พ.30
                  </div>
                  <div className="text-sm font-bold tabular-nums text-primary pt-1 border-t border-border/30 flex justify-between items-center">
                    <span className="text-xs font-normal text-muted-foreground">ยอดเงินคงเหลือ:</span>
                    <span>฿ 1,234,567.89</span>
                  </div>
                </div>
              </div>

              <div className="mt-4 pt-3 border-t border-border/40 flex items-center justify-between text-xs text-muted-foreground">
                <span className="font-mono text-[11px]">font-sans</span>
                <span className="group-hover:text-primary transition-colors font-medium">
                  {isSelected ? tr("current_selected", "เลือกแล้ว") : tr("click_to_select", "คลิกเพื่อเลือกใช้")}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
