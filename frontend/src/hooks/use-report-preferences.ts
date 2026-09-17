"use client";

import { useState, useEffect, useCallback } from "react";

export type ReportFontSize = "normal" | "medium" | "large";

export interface ReportPreferences {
  fontSize: ReportFontSize;
  highContrast: boolean;
}

const STORAGE_KEY = "bc_report_pref";

export function getReportFontSizeClass(size: ReportFontSize): string {
  switch (size) {
    case "medium":
      return "text-[1.05rem] [&_td]:py-3 [&_th]:py-3.5";
    case "large":
      return "text-[1.2rem] [&_td]:py-3.5 [&_th]:py-4";
    case "normal":
    default:
      return "text-[0.95rem] [&_td]:py-2 [&_th]:py-2.5";
  }
}

export function getReportContrastClass(highContrast: boolean): string {
  if (!highContrast) return "";
  return "bc-high-contrast border-2 border-foreground/70 [&_th]:bg-muted [&_th]:text-foreground [&_th]:border-b-2 [&_th]:border-foreground/80 [&_td]:border-b [&_td]:border-foreground/30 [&_td]:text-foreground [&_.tabular-nums]:font-bold [&_.tabular-nums]:text-foreground [&_tr:nth-child(even)]:bg-muted/40";
}

/**
 * Hook for managing report viewer accessibility preferences:
 * - Font sizing: normal (0.95rem), medium (1.05rem), large (1.2rem)
 * - High Contrast: crisp solid borders, bold numbers, enhanced contrast
 * Persisted in localStorage ("bc_report_pref")
 */
export function useReportPreferences() {
  const [fontSize, setFontSizeState] = useState<ReportFontSize>("normal");
  const [highContrast, setHighContrastState] = useState<boolean>(false);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    if (typeof window === "undefined") return;
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (raw) {
        const parsed = JSON.parse(raw) as Partial<ReportPreferences>;
        if (parsed.fontSize) setFontSizeState(parsed.fontSize);
        if (typeof parsed.highContrast === "boolean") setHighContrastState(parsed.highContrast);
      }
    } catch {
      // ignore storage errors
    }
  }, []);

  const setFontSize = useCallback((size: ReportFontSize) => {
    setFontSizeState(size);
    if (typeof window === "undefined") return;
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      const current = raw ? JSON.parse(raw) : {};
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ ...current, fontSize: size }));
    } catch {
      // ignore storage errors
    }
  }, []);

  const toggleHighContrast = useCallback(() => {
    setHighContrastState((prev) => {
      const next = !prev;
      if (typeof window !== "undefined") {
        try {
          const raw = localStorage.getItem(STORAGE_KEY);
          const current = raw ? JSON.parse(raw) : {};
          localStorage.setItem(STORAGE_KEY, JSON.stringify({ ...current, highContrast: next }));
        } catch {
          // ignore storage errors
        }
      }
      return next;
    });
  }, []);

  return {
    fontSize,
    setFontSize,
    highContrast,
    toggleHighContrast,
    mounted,
    fontSizeClass: getReportFontSizeClass(fontSize),
    contrastClass: getReportContrastClass(highContrast),
  };
}
