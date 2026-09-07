import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const source = readFileSync(fileURLToPath(new URL("./main-menu-screen.tsx", import.meta.url)), "utf8");

describe("main menu sidebar resize behavior", () => {
  it("defines proper bounds and storage key for sidebar width", () => {
    expect(source).toContain('const SIDEBAR_MIN_WIDTH = 260;');
    expect(source).toContain('const SIDEBAR_MAX_WIDTH = 540;');
    expect(source).toContain('const SIDEBAR_DEFAULT_WIDTH = 320;');
    expect(source).toContain('const SIDEBAR_WIDTH_STORAGE_KEY = "bc_menu_sidebar_width";');
  });

  it("persists and hydrates sidebar width from localStorage", () => {
    expect(source).toContain('localStorage.getItem(SIDEBAR_WIDTH_STORAGE_KEY)');
    expect(source).toContain('localStorage.setItem(SIDEBAR_WIDTH_STORAGE_KEY');
  });

  it("applies dynamic CSS variable for grid columns without hardcoded width", () => {
    expect(source).toContain('"lg:grid-cols-[var(--menu-sidebar-width)_minmax(0,1fr)]"');
    expect(source).toContain('--menu-sidebar-width');
    expect(source).not.toContain('lg:grid-cols-[280px_minmax(0,1fr)]');
  });

  it("provides accessible drag handle with role=separator and keyboard controls", () => {
    expect(source).toContain('role="separator"');
    expect(source).toContain('aria-orientation="vertical"');
    expect(source).toContain('onPointerDown={handleSidebarResizeStart}');
    expect(source).toContain('onDoubleClick={handleSidebarResizeReset}');
    expect(source).toContain('onKeyDown={handleSidebarKeyDown}');
  });
});
