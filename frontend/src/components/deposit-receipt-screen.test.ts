import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { DepositReceiptScreen } from "./deposit-receipt-screen";

describe("DepositReceiptScreen (Champ Parity 30006 - BILLFrmARDeposit)", () => {
  it("exports component function", () => {
    expect(typeof DepositReceiptScreen).toBe("function");
  });

  it("has active route for receive-deposit in menu data hierarchy", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "receive-deposit");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/transaction/receivedeposit");
    expect(item?.label.th).toContain("บันทึกใบรับเงินมัดจำ");
  });

  it("is active and not marked as pending", () => {
    expect(isMenuScreenPending("/transaction/receivedeposit")).toBe(false);
  });
});
