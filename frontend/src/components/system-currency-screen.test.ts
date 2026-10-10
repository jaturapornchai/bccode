import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { SystemCurrencyScreen } from "./system-currency-screen";

describe("SystemCurrencyScreen (Champ Parity S2 / Currency Configuration)", () => {
  it("exports component function", () => {
    expect(typeof SystemCurrencyScreen).toBe("function");
  });

  it("has active route for system-currency in menu data hierarchy under system-setup", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "system-currency");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/settings/currency");
    expect(item?.label.th).toContain("กำหนดสกุลเงิน");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/settings/currency")).toBe(false);
  });
});
