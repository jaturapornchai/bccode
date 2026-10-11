import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { flattenMenuItems } from "@/lib/menu-data";
import { isMenuScreenPending } from "@/lib/menu-screen-status";

describe("Billing Note Auto Print (ARFrmPrintBillAutoView parity)", () => {
  it("has active route for billing-note-auto-print in menu hierarchy", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "billing-note-auto-print");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/ar/billing-note-auto-print");
    expect(item?.label.th).toContain("พิมพ์ใบวางบิลอัตโนมัติ");
  });

  it("is recognized as an active connected screen (not pending)", () => {
    expect(isMenuScreenPending("/ar/billing-note-auto-print")).toBe(false);
  });
});

describe("Billing Note Auto Print API wiring", () => {
  const source = readFileSync(resolve(process.cwd(), "src/components/billing-note-auto-print.tsx"), "utf8");

  it("calls the BFF route /api/erp-transaction/billingnote/auto-print with the filter names the BFF allows", () => {
    expect(source).toContain("/api/erp-transaction/billingnote/auto-print");
    expect(source).not.toMatch(/apiFetch\((`|")\/transaction\//);
    for (const name of ["from_date", "to_date", "from_docno", "to_docno", "from_cust", "to_cust", "sale_code"]) {
      expect(source, name).toContain(`params.append("${name}"`);
    }
  });
});
