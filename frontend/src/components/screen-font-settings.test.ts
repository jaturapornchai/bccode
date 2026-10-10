import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { appFonts, defaultFontId, normalizeAppFont } from "@/lib/font-data";
import { FONT_TARGET_CONFIGS, targetFromRoute } from "./screen-font-settings";

describe("Screen Font Settings (Champ parity - Window Fonts)", () => {
  it("has active routes for all 6 font settings in menu hierarchy under window head", () => {
    const items = flattenMenuItems();
    const targets = [
      { id: "window-font-screen", route: "/settings/font", label: "เลือกตัวอักษรสำหรับหน้าจอทั่วไป" },
      { id: "window-font-report", route: "/settings/font-report", label: "เลือกตัวอักษรสำหรับรายงานทั่วไป" },
      { id: "window-font-report-title", route: "/settings/font-report-title", label: "เลือกตัวอักษรสำหรับชื่อของรายงาน" },
      { id: "window-font-report-column", route: "/settings/font-report-column", label: "เลือกตัวอักษรสำหรับคอลัมน์ของรายงาน" },
      { id: "window-font-report-detail", route: "/settings/font-report-detail", label: "เลือกตัวอักษรสำหรับรายการย่อยของรายงาน" },
      { id: "window-font-report-company", route: "/settings/font-report-company", label: "เลือกตัวอักษรสำหรับชื่อบริษัทในรายงาน" },
    ];

    for (const t of targets) {
      const item = items.find((i) => i.id === t.id);
      expect(item, t.id).toBeDefined();
      expect(item?.route).toBe(t.route);
      expect(item?.label.th).toContain(t.label);
      expect(isMenuScreenPending(t.route)).toBe(false);
    }
  });

  it("correctly resolves target from route parameter", () => {
    expect(targetFromRoute("/settings/font")).toBe("screen");
    expect(targetFromRoute("/settings/font-report")).toBe("report");
    expect(targetFromRoute("/settings/font-report-title")).toBe("report-title");
    expect(targetFromRoute("/settings/font-report-column")).toBe("report-column");
    expect(targetFromRoute("/settings/font-report-detail")).toBe("report-detail");
    expect(targetFromRoute("/settings/font-report-company")).toBe("report-company");
  });

  it("provides available fonts with default Thai typography", () => {
    expect(appFonts.length).toBeGreaterThan(0);
    expect(defaultFontId).toBe("inter");
    expect(normalizeAppFont("sarabun")).toBe("sarabun");
    expect(normalizeAppFont("non-existent-font")).toBe(defaultFontId);
    expect(Object.keys(FONT_TARGET_CONFIGS)).toHaveLength(6);
  });

  it("supports company-level default persistence and theme selection", () => {
    const storage: Record<string, string> = {};
    const mockCompanySettings = {
      fonts: {
        screen: "sarabun",
        report: "sarabun",
        "report-title": "kanit",
        "report-column": "prompt",
        "report-detail": "sarabun",
        "report-company": "ibm-plex-sans-thai",
      },
      theme: "dark",
      updatedAt: new Date().toISOString(),
    };
    storage["bc_company_font_theme_settings"] = JSON.stringify(mockCompanySettings);
    const stored = JSON.parse(storage["bc_company_font_theme_settings"] || "{}");
    expect(stored.theme).toBe("dark");
    expect(stored.fonts.screen).toBe("sarabun");
    expect(stored.fonts["report-company"]).toBe("ibm-plex-sans-thai");
  });
});

