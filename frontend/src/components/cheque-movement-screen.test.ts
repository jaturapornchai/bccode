import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { ChequeMovementScreen } from "./cheque-movement-screen";

describe("ChequeMovementScreen (Champ Parity 60000+ / Bank Cheque Utilities)", () => {
  it("exports component function", () => {
    expect(typeof ChequeMovementScreen).toBe("function");
  });

  it("has active route for cheque-movement in menu data hierarchy under bank-utilities", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "cheque-movement");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/banking/cheque-movement");
    expect(item?.label.th).toContain("ข้อมูลการเคลื่อนไหวเช็ค");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/banking/cheque-movement")).toBe(false);
  });
});
