import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { INITIAL_PETTY_CASH_LIMITS } from "./petty-cash-limit-screen";

describe("Petty Cash Limit Screen (Champ parity 70001)", () => {
  it("has active route /banking/petty-cash-limit in menu hierarchy", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "petty-cash-limit");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/banking/petty-cash-limit");
    expect(item?.label.th).toContain("กำหนดวงเงินสดย่อย");
  });

  it("is recognized as an active connected screen (not pending)", () => {
    expect(isMenuScreenPending("/banking/petty-cash-limit")).toBe(false);
  });

  it("has realistic Thai SME petty cash initial limits adhering to Rule 20 (decimal string)", () => {
    expect(INITIAL_PETTY_CASH_LIMITS.length).toBeGreaterThan(0);
    for (const record of INITIAL_PETTY_CASH_LIMITS) {
      expect(record.code).toMatch(/^PC-\d{2}$/);
      expect(record.name).toBeTruthy();
      expect(record.custodian_name).toBeTruthy();
      expect(record.gl_account_code).toMatch(/^1111\d$/);
      // TFRS Rule 20: format string with 2 decimal places
      expect(record.limit_amount).toMatch(/^\d+\.\d{2}$/);
      expect(record.begin_amount).toMatch(/^\d+\.\d{2}$/);
      expect(record.balance).toMatch(/^\d+\.\d{2}$/);
    }
  });
});
