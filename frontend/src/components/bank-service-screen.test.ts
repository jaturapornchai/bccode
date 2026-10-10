import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { BankServiceScreen } from "./bank-service-screen";

describe("BankServiceScreen (Champ Parity 40000+ / Bank Accounts Services)", () => {
  it("exports component function", () => {
    expect(typeof BankServiceScreen).toBe("function");
  });

  it("has active route for bank-service in menu data hierarchy under bank-accounts", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "bank-service");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/banking/bank-service");
    expect(item?.label.th).toContain("บริการธนาคาร");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/banking/bank-service")).toBe(false);
  });
});
