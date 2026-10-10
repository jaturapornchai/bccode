import { describe, expect, it } from "vitest";
import { getOperationsConfig } from "@/lib/erp-operations";
import { isMenuScreenPending } from "@/lib/menu-screen-status";

describe("Purchase Receipt Adjustment (POFrmAdjGRStatus parity)", () => {
  it("has operations config registered for purchase receipt adjustment", () => {
    const config = getOperationsConfig("/transaction/purchasereceiptadjustment");
    expect(config).toBeDefined();
    expect(config?.code).toBe("purchase_receipt_adjustment");
    expect(config?.category).toBe("procurement");
    expect(config?.title.th).toContain("ปรับปรุงใบรับสินค้า");
  });

  it("is marked as active connected screen (not pending)", () => {
    expect(isMenuScreenPending("/transaction/purchasereceiptadjustment")).toBe(false);
  });
});
