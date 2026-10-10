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
