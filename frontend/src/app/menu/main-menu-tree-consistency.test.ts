import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const source = readFileSync(fileURLToPath(new URL("./main-menu-screen.tsx", import.meta.url)), "utf8");

describe("main menu tree consistency", () => {
  it("does not bypass single-item groups with raw leaf buttons", () => {
    // Single item groups should NOT be bypassed and turned into flat buttons
    expect(source).not.toContain("ถ้าข้างในมี เมนูเดียว ไม่ต้องทำเป็น Group");
  });

  // 2026-10-07: the left tree is the same layout as the Champ-style top bar (src/lib/menu-bars.ts):
  // one accordion per head, items in Champ order, submenus as folders — no per-group tree any more.
  it("renders the left tree from the same MENU_BARS as the top bar", () => {
    expect(source).toContain("TREE_BARS.map((bar)");
    expect(source).toContain("<ChampMenuBar");
    expect(source).not.toContain("TopMenuChrome");
    expect(source).not.toContain("<MenuTreeGroup");
  });
});
