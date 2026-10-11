import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { appFonts, defaultFontId, normalizeAppFont } from "@/lib/font-data";

describe("Screen Font Settings (Champ parity - Window Fonts)", () => {
  it("keeps /settings/font working and sets the 5 report-font menus to รอพัฒนา (2026-10-11)", () => {
    const items = flattenMenuItems();
    const screen = items.find((i) => i.id === "window-font-screen");
    expect(screen?.route).toBe("/settings/font");
    expect(screen?.label.th).toContain("เลือกตัวอักษรสำหรับหน้าจอทั่วไป");
    expect(isMenuScreenPending("/settings/font")).toBe(false);

    // report fonts were saved in localStorage but no code read them -> pending until stored per company in the backend
    const reportFonts = [
      { id: "window-font-report", route: "/settings/font-report", label: "เลือกตัวอักษรสำหรับรายงานทั่วไป" },
      { id: "window-font-report-title", route: "/settings/font-report-title", label: "เลือกตัวอักษรสำหรับชื่อของรายงาน" },
      { id: "window-font-report-column", route: "/settings/font-report-column", label: "เลือกตัวอักษรสำหรับคอลัมน์ของรายงาน" },
      { id: "window-font-report-detail", route: "/settings/font-report-detail", label: "เลือกตัวอักษรสำหรับรายการย่อยของรายงาน" },
      { id: "window-font-report-company", route: "/settings/font-report-company", label: "เลือกตัวอักษรสำหรับชื่อบริษัทในรายงาน" },
    ];
    for (const t of reportFonts) {
      const item = items.find((i) => i.id === t.id);
      expect(item, t.id).toBeDefined();
      expect(item?.route).toBe(t.route);
      expect(item?.label.th).toContain(t.label);
      expect(isMenuScreenPending(t.route), t.route).toBe(true);
    }
  });

  it("provides available fonts with default Thai typography", () => {
    expect(appFonts.length).toBeGreaterThan(0);
    expect(defaultFontId).toBe("inter");
    expect(normalizeAppFont("sarabun")).toBe("sarabun");
    expect(normalizeAppFont("non-existent-font")).toBe(defaultFontId);
  });

  it("offers only the screen font: no target tabs, no company-default button, no report-font storage", () => {
    const source = readFileSync(resolve(process.cwd(), "src/components/screen-font-settings.tsx"), "utf8");
    expect(source).not.toContain("FONT_TARGET_CONFIGS");
    expect(source).not.toContain("handleSaveCompanyDefault");
    expect(source).not.toContain("bc_company_font_theme_settings");
    expect(source).not.toContain("bc_app_font_report");
    // heading and description come from language keys, not hard-coded text (no "(Font for General Screens)" subtitle)
    expect(source).toContain('tr("window_font_screen"');
    expect(source).toContain('"font_screen_description"');
    expect(source).not.toContain("Font for General Screens");
    // theme bar + footer text also come from language keys (Thai left on screen in en mode = bug)
    for (const key of ["font_theme_mode_label", "font_theme_mode_hint", "font_theme_light", "font_theme_dark", "font_theme_system", "font_compat_note"]) {
      expect(source, key).toContain(`tr("${key}"`);
    }
    expect(source).not.toMatch(/>\s*(สว่าง|มืด|ตามระบบ)\s*\(/);
    // the screen font still takes effect immediately (cookie + CSS variable)
    expect(source).toContain("applyAppFont(fontId)");
    expect(source).toContain("writeFontCookie(fontId)");
  });
});
