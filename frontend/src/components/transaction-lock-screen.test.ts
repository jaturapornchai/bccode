import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { TransactionLockScreen } from "./transaction-lock-screen";

describe("TransactionLockScreen (Champ Parity 80000+ / Tools Transaction Lock)", () => {
  it("exports component function", () => {
    expect(typeof TransactionLockScreen).toBe("function");
  });

  it("has active route for transaction-lock in menu data hierarchy under ic-utilities", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "transaction-lock");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/tools/transaction-lock");
    expect(item?.label.th).toContain("Transaction Lock");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/tools/transaction-lock")).toBe(false);
  });
});
