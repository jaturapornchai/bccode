import { describe, expect, it } from "vitest";
import { accountReferenceRows } from "./gl-account-references";

const tr = (_key: string, fallback: string) => fallback;

describe("accountReferenceRows (Champ 5502 'รายการที่ใช้บัญชีนี้')", () => {
  it("lists the kinds in a fixed order, whatever order the backend sends", () => {
    const rows = accountReferenceRows([
      { kind: "fiscal-years", count: 1 }, { kind: "journals", count: 12 }, { kind: "statement-templates", count: 2 }, { kind: "budgets", count: 3 },
    ], tr);
    expect(rows.map((row) => row.kind)).toEqual(["journals", "budgets", "statement-templates", "fiscal-years"]);
    expect(rows[0]).toEqual({ kind: "journals", label: "ใบสำคัญรายวัน", count: 12 });
  });
  it("drops zero counts, so an unused account has no rows", () => {
    expect(accountReferenceRows([{ kind: "journals", count: 0 }, { kind: "budgets", count: 0 }], tr)).toEqual([]);
    expect(accountReferenceRows([], tr)).toEqual([]);
  });
  it("shows a kind the backend adds later last, with a generic label, and never hides it", () => {
    const rows = accountReferenceRows([{ kind: "zeta-kind", count: 1 }, { kind: "alpha-kind", count: 4 }, { kind: "accounts", count: 2 }], tr);
    expect(rows.map((row) => [row.kind, row.label])).toEqual([["accounts", "บัญชีลูกที่อยู่ใต้บัญชีนี้"], ["alpha-kind", "ข้อมูลอื่น (alpha-kind)"], ["zeta-kind", "ข้อมูลอื่น (zeta-kind)"]]);
  });
  it("does not mutate the backend list", () => {
    const items = [{ kind: "budgets", count: 1 }, { kind: "journals", count: 2 }];
    accountReferenceRows(items, tr);
    expect(items.map((item) => item.kind)).toEqual(["budgets", "journals"]);
  });
});
