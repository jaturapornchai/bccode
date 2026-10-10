import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { ChequeIssuedPrepareScreen } from "./cheque-issued-prepare-screen";

describe("ChequeIssuedPrepareScreen (Champ Parity 60022 - CQFrmCheqOut)", () => {
  it("exports component function", () => {
    expect(typeof ChequeIssuedPrepareScreen).toBe("function");
  });

  it("has active route for cheque-issued-prepare in menu data hierarchy under cash-bank section", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "cheque-issued-prepare");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/banking/cheque-issued-prepare");
    expect(item?.label.th).toContain("บันทึกเช็คเตรียมจ่าย");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/banking/cheque-issued-prepare")).toBe(false);
  });
});
