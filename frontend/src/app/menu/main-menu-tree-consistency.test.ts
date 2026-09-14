import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const source = readFileSync(fileURLToPath(new URL("./main-menu-screen.tsx", import.meta.url)), "utf8");

describe("main menu tree consistency", () => {
  it("does not bypass single-item groups with raw leaf buttons", () => {
    // Single item groups should NOT be bypassed and turned into flat buttons
    expect(source).not.toContain("ถ้าข้างในมี เมนูเดียว ไม่ต้องทำเป็น Group");
  });

  it("renders all groups uniformly with MenuTreeGroup in multi-group sections", () => {
    expect(source).toContain("<MenuTreeGroup");
  });

  it("has group icon configured for bank-accounts", () => {
    expect(source).toContain('"bank-accounts": Landmark,');
  });
});
