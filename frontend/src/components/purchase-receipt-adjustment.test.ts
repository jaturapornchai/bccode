import { readFileSync } from "node:fs";
import { resolve } from "node:path";
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

describe("Purchase Receipt Adjustment API wiring", () => {
  const source = readFileSync(resolve(process.cwd(), "src/components/purchase-receipt-adjustment.tsx"), "utf8");

  it("calls the BFF route /api/erp-transaction/goodsreceive (a bare /transaction URL has no Next route and 404s)", () => {
    expect(source).toContain("/api/erp-transaction/goodsreceive/receipt-adjustments");
    expect(source).not.toMatch(/apiFetch\((`|")\/transaction\//);
  });

  it("treats a success:false body (HTTP 200 from the BFF) as a failure instead of counting it as adjusted", () => {
    expect(source).toContain("result.success === false");
  });
});
