import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";

describe("Customer Address Label Print (Champ parity)", () => {
  it("has active route /customer/label-print in menu hierarchy", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "customer-label-print");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/customer/label-print");
    expect(item?.label.th).toContain("พิมพ์ป้ายผนึกชื่อลูกค้า-สมาชิก");
  });

  it("is recognized as an active connected screen (not pending)", () => {
    expect(isMenuScreenPending("/customer/label-print")).toBe(false);
  });
});

describe("Customer Address Label Print API wiring", () => {
  const source = readFileSync(resolve(process.cwd(), "src/components/customer-label-print.tsx"), "utf8");

  it("posts through the goapi BFF (/api/goapi/api/...) instead of a bare /api/customer path that has no Next route", () => {
    expect(source).toContain('apiFetch("/api/goapi/api/customer/labels/query"');
    expect(source).not.toContain('apiFetch("/api/customer/labels/query"');
  });
});
