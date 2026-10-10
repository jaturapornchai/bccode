import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { SystemPaymentTermsScreen } from "./system-payment-terms-screen";

describe("SystemPaymentTermsScreen (Champ Parity S2 / Payment Terms Configuration)", () => {
  it("exports component function", () => {
    expect(typeof SystemPaymentTermsScreen).toBe("function");
  });

  it("has active route for system-payment-terms in menu data hierarchy under system-setup", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "system-payment-terms");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/settings/payment-terms");
    expect(item?.label.th).toContain("เงื่อนไขการชำระเงิน");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/settings/payment-terms")).toBe(false);
  });
});
