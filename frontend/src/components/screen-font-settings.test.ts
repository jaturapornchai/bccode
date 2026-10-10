import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { appFonts, defaultFontId, normalizeAppFont } from "@/lib/font-data";

describe("Screen Font Settings (Champ parity - Window Font)", () => {
  it("has active route /settings/font in menu hierarchy under window head", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "window-font-screen");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/settings/font");
    expect(item?.label.th).toContain("เลือกตัวอักษรสำหรับหน้าจอทั่วไป");
  });

  it("is recognized as an active connected screen (not pending)", () => {
    expect(isMenuScreenPending("/settings/font")).toBe(false);
  });

  it("provides available fonts with default Thai typography", () => {
    expect(appFonts.length).toBeGreaterThan(0);
    expect(defaultFontId).toBe("inter");
    expect(normalizeAppFont("sarabun")).toBe("sarabun");
    expect(normalizeAppFont("non-existent-font")).toBe(defaultFontId);
  });
});
