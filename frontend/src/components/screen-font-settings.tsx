"use client";

import { useEffect, useState } from "react";
import { Check, Type, RefreshCw, FileText, Table, List, Building, Sparkles } from "lucide-react";
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

export type FontTarget =
  | "screen"
  | "report"
  | "report-title"
  | "report-column"
  | "report-detail"
  | "report-company";

export interface FontTargetConfig {
  id: FontTarget;
  route: string;
  menuId: string;
  titleTh: string;
  titleEn: string;
  descTh: string;
  storageKey: string;
  defaultFont: FontId;
}

export const FONT_TARGET_CONFIGS: Record<FontTarget, FontTargetConfig> = {
  screen: {
    id: "screen",
    route: "/settings/font",
    menuId: "window-font-screen",
    titleTh: "เลือกตัวอักษรสำหรับหน้าจอทั่วไป",
    titleEn: "Font for General Screens",
    descTh: "เลือกรูปแบบตัวอักษร (Typography) สำหรับเมนู ฟอร์ม และหน้าจอการทำงานทั่วไปในระบบ การตั้งค่ามีผลทันที",
    storageKey: fontStorageKey,
    defaultFont: "inter",
  },
  report: {
    id: "report",
    route: "/settings/font-report",
    menuId: "window-font-report",
    titleTh: "เลือกตัวอักษรสำหรับรายงานทั่วไป",
    titleEn: "Font for General Reports",
    descTh: "เลือกรูปแบบตัวอักษรสำหรับเนื้อหาตารางและตัวเลขในรายงานทางการเงินและบัญชีแยกประเภท",
    storageKey: "bc_app_font_report",
    defaultFont: "sarabun",
  },
  "report-title": {
    id: "report-title",
    route: "/settings/font-report-title",
    menuId: "window-font-report-title",
    titleTh: "เลือกตัวอักษรสำหรับชื่อของรายงาน",
    titleEn: "Font for Report Titles",
    descTh: "เลือกรูปแบบตัวอักษรสำหรับหัวเรื่องและชื่อรายงานด้านบนสุดของเอกสาร",
    storageKey: "bc_app_font_report_title",
    defaultFont: "kanit",
  },
  "report-column": {
    id: "report-column",
    route: "/settings/font-report-column",
    menuId: "window-font-report-column",
    titleTh: "เลือกตัวอักษรสำหรับคอลัมน์ของรายงาน",
    titleEn: "Font for Report Columns",
    descTh: "เลือกรูปแบบตัวอักษรสำหรับหัวคอลัมน์ (Table Headers) ในรายงานทุกประเภท",
    storageKey: "bc_app_font_report_column",
    defaultFont: "prompt",
  },
  "report-detail": {
    id: "report-detail",
    route: "/settings/font-report-detail",
    menuId: "window-font-report-detail",
    titleTh: "เลือกตัวอักษรสำหรับรายการย่อยของรายงาน",
    titleEn: "Font for Report Details",
    descTh: "เลือกรูปแบบตัวอักษรสำหรับคำอธิบายรายการย่อย ยอดรวมย่อย และหมายเหตุกำกับ",
    storageKey: "bc_app_font_report_detail",
    defaultFont: "sarabun",
  },
  "report-company": {
    id: "report-company",
    route: "/settings/font-report-company",
    menuId: "window-font-report-company",
    titleTh: "เลือกตัวอักษรสำหรับชื่อบริษัทในรายงาน",
    titleEn: "Font for Company Name in Reports",
    descTh: "เลือกรูปแบบตัวอักษรสำหรับชื่อบริษัท/องค์กรและสาขาบนหัวรายงานทุกฉบับ",
    storageKey: "bc_app_font_report_company",
    defaultFont: "ibm-plex-sans-thai",
  },
};

export function targetFromRoute(route?: string): FontTarget {
  if (!route) return "screen";
  if (route.includes("font-report-title")) return "report-title";
  if (route.includes("font-report-column")) return "report-column";
  if (route.includes("font-report-detail")) return "report-detail";
  if (route.includes("font-report-company")) return "report-company";
  if (route.includes("font-report")) return "report";
  return "screen";
}

interface ScreenFontSettingsProps {
  language?: LanguageCode;
  route?: string;
  initialTarget?: FontTarget;
}

export function ScreenFontSettings({
  language: _language = "th",
  route,
  initialTarget,
}: ScreenFontSettingsProps) {
  const tr = useBackendText();
  const [activeTarget, setActiveTarget] = useState<FontTarget>(
    initialTarget ?? targetFromRoute(route)
  );

  const currentConfig = FONT_TARGET_CONFIGS[activeTarget];
  const [selectedFonts, setSelectedFonts] = useState<Record<FontTarget, FontId>>({
    screen: defaultFontId,
    report: "sarabun",
    "report-title": "kanit",
    "report-column": "prompt",
    "report-detail": "sarabun",
    "report-company": "ibm-plex-sans-thai",
  });
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (route) {
      setActiveTarget(targetFromRoute(route));
    }
  }, [route]);

  useEffect(() => {
    const loaded: Record<FontTarget, FontId> = {} as Record<FontTarget, FontId>;
    for (const [key, cfg] of Object.entries(FONT_TARGET_CONFIGS)) {
      const targetKey = key as FontTarget;
      const saved = localStorage.getItem(cfg.storageKey);
      loaded[targetKey] = saved ? normalizeAppFont(saved) : cfg.defaultFont;
    }
    setSelectedFonts(loaded);
    setReady(true);
    appFonts.forEach((font) => ensureFontLoaded(font));
  }, []);

  const handleSelectFont = (fontId: FontId) => {
    const next = { ...selectedFonts, [activeTarget]: fontId };
    setSelectedFonts(next);
    localStorage.setItem(currentConfig.storageKey, fontId);

    // If changing screen font, immediately update live runtime and cookie
    if (activeTarget === "screen") {
      applyAppFont(fontId);
      writeFontCookie(fontId);
    }
  };

  const handleResetDefault = () => {
    handleSelectFont(currentConfig.defaultFont);
  };

  const getTargetIcon = (target: FontTarget) => {
    switch (target) {
      case "screen":
        return <Type className="h-4 w-4" />;
      case "report":
        return <FileText className="h-4 w-4" />;
      case "report-title":
        return <FileText className="h-4 w-4" />;
      case "report-column":
        return <Table className="h-4 w-4" />;
      case "report-detail":
        return <List className="h-4 w-4" />;
      case "report-company":
        return <Building className="h-4 w-4" />;
    }
  };

  const currentFontId = selectedFonts[activeTarget];

  return (
    <div className="flex-1 min-h-0 flex flex-col gap-4 p-4 lg:p-6 overflow-auto">
      {/* Target Selector Tabs */}
      <div className="flex flex-wrap items-center gap-1.5 p-1.5 rounded-xl border border-border/70 bg-muted/40 shadow-xs">
        {(Object.keys(FONT_TARGET_CONFIGS) as FontTarget[]).map((key) => {
          const cfg = FONT_TARGET_CONFIGS[key];
          const isCurrentTab = activeTarget === key;
          return (
            <button
              key={key}
              onClick={() => setActiveTarget(key)}
              className={`flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold transition-all ${
                isCurrentTab
                  ? "bg-card text-foreground shadow-xs border border-border/80"
                  : "text-muted-foreground hover:text-foreground hover:bg-card/40"
              }`}
            >
              {getTargetIcon(key)}
              <span>{cfg.titleTh}</span>
              <span className="text-[10px] opacity-60 font-mono bg-muted px-1.5 py-0.5 rounded">
                {selectedFonts[key]}
              </span>
            </button>
          );
        })}
      </div>

      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-5 shadow-xs">
        <div className="flex items-center gap-3.5">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-inner">
            {getTargetIcon(activeTarget)}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">{currentConfig.titleTh}</h1>
              <span className="text-xs text-muted-foreground font-mono">({currentConfig.titleEn})</span>
            </div>
            <p className="text-sm text-muted-foreground mt-0.5">{currentConfig.descTh}</p>
          </div>
        </div>

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

                {/* Live Preview Box */}
                <div
                  style={{ fontFamily: font.family }}
                  className="rounded-xl border border-border/50 bg-background/80 p-3.5 text-foreground transition-colors group-hover:border-border"
                >
                  {activeTarget === "screen" && (
                    <>
                      <div className="text-sm font-medium leading-relaxed">
                        ระบบบัญชีแยกประเภทและการเงิน
                      </div>
                      <div className="text-xs text-muted-foreground mt-1">
                        General Ledger &amp; Financial ERP
                      </div>
                      <div className="text-xs font-semibold text-primary mt-2">
                        1,250,800.50 บาท (THB)
                      </div>
                    </>
                  )}

                  {activeTarget === "report" && (
                    <>
                      <div className="text-sm font-medium leading-relaxed">
                        รายงานงบทดลองและบัญชีแยกประเภท
                      </div>
                      <div className="text-xs text-muted-foreground mt-1">
                        Trial Balance &amp; General Ledger
                      </div>
                      <div className="text-xs font-mono font-medium text-emerald-600 mt-2">
                        เดบิต 350,000.00 | เครดิต 350,000.00
                      </div>
                    </>
                  )}

                  {activeTarget === "report-title" && (
                    <>
                      <div className="text-base font-bold text-foreground">
                        งบแสดงฐานะการเงินและงบกำไรขาดทุน
                      </div>
                      <div className="text-xs text-muted-foreground mt-1">
                        ณ วันที่ 31 ธันวาคม 2569
                      </div>
                    </>
                  )}

                  {activeTarget === "report-column" && (
                    <>
                      <div className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
                        รหัสบัญชี | ชื่อบัญชี | ยอดยกมา | เดบิต | เครดิต | ยอดยกไป
                      </div>
                      <div className="text-xs text-foreground mt-1.5 font-medium">
                        11110 เงินสดในมือ | 50,000.00 บาท
                      </div>
                    </>
                  )}

                  {activeTarget === "report-detail" && (
                    <>
                      <div className="text-xs text-muted-foreground leading-relaxed">
                        หมายเหตุ: รายการนี้ได้รับการตรวจสอบความถูกต้องและผ่านการตรวจสอบโดยผู้สอบบัญชีแล้ว
                      </div>
                    </>
                  )}

                  {activeTarget === "report-company" && (
                    <>
                      <div className="text-sm font-bold text-foreground">
                        บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด (สำนักงานใหญ่)
                      </div>
                      <div className="text-xs text-muted-foreground mt-1">
                        เลขประจำตัวผู้เสียภาษีอากร 0105558123456
                      </div>
                    </>
                  )}
                </div>
              </div>

              <div className="mt-4 flex items-center justify-between pt-2 border-t border-border/40 text-[11px] text-muted-foreground">
                <span>ความเข้ากันได้: ไทย + สากล</span>
                <span className="font-mono">{font.id}</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
