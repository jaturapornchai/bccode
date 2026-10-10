import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { SystemCreditDebitReasonsScreen } from "./system-credit-debit-reasons-screen";

describe("SystemCreditDebitReasonsScreen (Champ Parity S2 / Credit & Debit Note Reasons)", () => {
  it("exports component function", () => {
    expect(typeof SystemCreditDebitReasonsScreen).toBe("function");
  });

  it("has active route for system-credit-debit-reasons in menu data hierarchy under system-setup", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "system-credit-debit-reasons");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/settings/credit-debit-reasons");
    expect(item?.label.th).toContain("กำหนดสาเหตุการลดหนี้/เพิ่มหนี้");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/settings/credit-debit-reasons")).toBe(false);
  });
});
