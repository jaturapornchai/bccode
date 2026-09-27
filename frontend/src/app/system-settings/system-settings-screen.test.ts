import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";

const source = readFileSync(
  join(process.cwd(), "src/app/system-settings/system-settings-screen.tsx"),
  "utf8",
);

describe("SystemSettingsScreen layout contract", () => {
  it("keeps standalone system setting pages full width without hiding horizontal overflow", () => {
    expect(source).toContain(
      '<main className="min-h-dvh w-full max-w-none bg-background p-2 text-foreground sm:p-3">',
    );
    expect(source).not.toContain(
      '<main className="min-h-dvh w-full overflow-x-hidden bg-background p-2 text-foreground sm:p-3">',
    );
  });

  // Standalone /<setting> routes (src/app/[systemSetting]/page.tsx) must show the same "รอพัฒนา" card as the
  // menu for a screen whose backend was removed, and never call that missing API.
  it("renders the menu's planned card for a retired backend and skips loading its records", () => {
    const guard = source.indexOf("if (isMenuBackendRetired(config.route)) {");
    expect(guard).toBeGreaterThan(source.indexOf("if (!config) {"));
    expect(guard).toBeLessThan(source.indexOf("const content = ("));
    expect(source.slice(guard, guard + 600)).toContain("<MenuPlannedCard");
    expect(source).toContain("if (isMenuBackendRetired(currentConfig.route)) return;");
    // The warehouse screen called /warehouse/* (removed with MongoDB); it comes back with a PostgreSQL API.
    expect(source).not.toContain("WarehouseTreeView");
  });
});
