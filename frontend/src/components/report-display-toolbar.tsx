"use client";

import { Contrast } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { ReportFontSize } from "@/hooks/use-report-preferences";

export interface ReportDisplayToolbarProps {
  fontSize: ReportFontSize;
  onFontSizeChange: (size: ReportFontSize) => void;
  highContrast: boolean;
  onToggleHighContrast: () => void;
}

/**
 * Reusable Display Controls for Financial Reports and Ledger Tables:
 * - Font Size toggles: A (Normal), A+ (Medium), A++ (Large)
 * - High Contrast mode toggle: dark/light high contrast with clear borders and bold tabular numbers
 */
export function ReportDisplayToolbar({
  fontSize,
  onFontSizeChange,
  highContrast,
  onToggleHighContrast,
}: ReportDisplayToolbarProps) {
  return (
    <div className="flex flex-wrap items-center gap-1.5 text-xs">
      {/* Font Size Adjusters */}
      <div className="inline-flex items-center rounded-lg border border-border bg-muted/40 p-0.5 shadow-xs">
        <button
          type="button"
          onClick={() => onFontSizeChange("normal")}
          className={`px-2 py-1 rounded text-xs font-medium transition-colors ${
            fontSize === "normal"
              ? "bg-background text-foreground shadow-xs font-semibold"
              : "text-muted-foreground hover:text-foreground"
          }`}
          title="ขนาดตัวอักษรปกติ (0.95rem)"
          aria-label="ขนาดตัวอักษรปกติ"
        >
          A
        </button>
        <button
          type="button"
          onClick={() => onFontSizeChange("medium")}
          className={`px-2 py-1 rounded text-xs font-medium transition-colors ${
            fontSize === "medium"
              ? "bg-background text-foreground shadow-xs font-semibold"
              : "text-muted-foreground hover:text-foreground"
          }`}
          title="ขนาดตัวอักษรปานกลาง (1.05rem)"
          aria-label="ขนาดตัวอักษรปานกลาง"
        >
          A+
        </button>
        <button
          type="button"
          onClick={() => onFontSizeChange("large")}
          className={`px-2 py-1 rounded text-xs font-medium transition-colors ${
            fontSize === "large"
              ? "bg-background text-foreground shadow-xs font-semibold"
              : "text-muted-foreground hover:text-foreground"
          }`}
          title="ขนาดตัวอักษรใหญ่พิเศษ (1.2rem)"
          aria-label="ขนาดตัวอักษรใหญ่พิเศษ"
        >
          A++
        </button>
      </div>

      {/* High Contrast Toggle */}
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={onToggleHighContrast}
        className={`h-7 px-2.5 text-xs rounded-lg gap-1.5 transition-colors ${
          highContrast
            ? "bg-foreground text-background hover:bg-foreground/90 font-semibold border-foreground"
            : "text-muted-foreground hover:text-foreground"
        }`}
        title="เปิด/ปิด โหมดคอนทราสต์สูง สำหรับตรวจเช็กตัวเลขงบการเงิน"
        aria-pressed={highContrast}
      >
        <Contrast className="size-3.5" />
        <span>{highContrast ? "คอนทราสต์สูง: เปิด" : "คอนทราสต์สูง"}</span>
      </Button>
    </div>
  );
}
