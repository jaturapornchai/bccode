import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { ApPaymentPrepareScreen } from "./ap-payment-prepare-screen";

describe("ApPaymentPrepareScreen (Champ Parity 20000+ / AP Payment Prepare)", () => {
  it("exports component function", () => {
    expect(typeof ApPaymentPrepareScreen).toBe("function");
  });

  it("has active route for ap-payment-prepare in menu data hierarchy under ap-adjustments", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "ap-payment-prepare");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/ap/payment-prepare");
    expect(item?.label.th).toContain("บันทึกเตรียมจ่ายชำระหนี้");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/ap/payment-prepare")).toBe(false);
  });
});
