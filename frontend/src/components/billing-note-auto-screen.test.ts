import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { BillingNoteAutoScreen } from "./billing-note-auto-screen";

describe("BillingNoteAutoScreen (Champ Parity 30005 / AR Auto Billing)", () => {
  it("exports component function", () => {
    expect(typeof BillingNoteAutoScreen).toBe("function");
  });

  it("has active route for billing-note-auto in menu data hierarchy under sales-ar section", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "billing-note-auto");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/ar/billing-note-auto");
    expect(item?.label.th).toContain("บันทึกใบวางบิลอัตโนมัติ");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/ar/billing-note-auto")).toBe(false);
  });
});
