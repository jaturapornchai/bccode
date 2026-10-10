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
