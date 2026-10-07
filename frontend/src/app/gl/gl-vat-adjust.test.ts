import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { readFileSync } from "node:fs";
import { describe, expect, it, vi } from "vitest";
import type { GLVatAdjustRow } from "@/lib/general-ledger";
import { EMPTY_VAT_ADJUST_FILTER, GLVatAdjust, sameVatAdjustFilter, setVatAdjustEdit, vatAdjustFilterPayload, vatAdjustHasDate, vatAdjustInvoiceDateCommand, vatAdjustItems, vatAdjustKey, vatAdjustMergedRow, vatAdjustProblems, vatAdjustQuery, vatAdjustRowsCommand, type VatAdjustEdit } from "./gl-vat-adjust";

vi.mock("./gl-common", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./gl-common")>();
  return { ...actual, useGLCommand: vi.fn().mockReturnValue({ busy: false, execute: vi.fn() }) };
});

const row = (over: Partial<GLVatAdjustRow> = {}): GLVatAdjustRow => ({
  journalid: "j1", version: 3, docno: "PV-0001", docdate: "2026-01-20", id: "v1", tax_type: 1, document_type: 1,
  tax_invoice_no: "IV-001", tax_invoice_date: "2026-01-15", partner_code: "AP001", partner_name: "บริษัท ปูนซีเมนต์ไทย จำกัด",
  base_amount: "1000.00", zero_rate_amount: "0.00", exempt_amount: "0.00", vat_rate: "7", vat_amount: "70.00",
  tax_period_year: 2026, tax_period_month: 1, claim_status: 1, ...over,
});
const none = new Map<string, VatAdjustEdit>();

describe("purchase VAT adjustment helpers (Champ 5498)", () => {
  it("sends only the filled filter fields, trimmed", () => {
    expect(vatAdjustFilterPayload(EMPTY_VAT_ADJUST_FILTER)).toEqual({});
    expect(vatAdjustFilterPayload({ ...EMPTY_VAT_ADJUST_FILTER, invdatefrom: "2026-01-01", invdateto: " 2026-01-31 ", partnerfrom: " AP001 ", partnerto: "  " }))
      .toEqual({ invdatefrom: "2026-01-01", invdateto: "2026-01-31", partnerfrom: "AP001" });
  });

  it("requires a document-date or invoice-date bound before processing, like Champ", () => {
    expect(vatAdjustHasDate(EMPTY_VAT_ADJUST_FILTER)).toBe(false);
    expect(vatAdjustHasDate({ ...EMPTY_VAT_ADJUST_FILTER, partnerfrom: "AP001" })).toBe(false);
    expect(vatAdjustHasDate({ ...EMPTY_VAT_ADJUST_FILTER, docdateto: "2026-01-31" })).toBe(true);
    expect(vatAdjustHasDate({ ...EMPTY_VAT_ADJUST_FILTER, invdatefrom: "2026-01-01" })).toBe(true);
  });

  it("builds the list query with the filter, page and page size", () => {
    const query = new URLSearchParams(vatAdjustQuery({ ...EMPTY_VAT_ADJUST_FILTER, invdatefrom: "2026-01-01", partnerfrom: "AP001" }, 2));
    expect(Object.fromEntries(query)).toEqual({ invdatefrom: "2026-01-01", partnerfrom: "AP001", page: "2", limit: "100" });
  });

  it("compares filters by their effective values", () => {
    const base = { ...EMPTY_VAT_ADJUST_FILTER, invdatefrom: "2026-01-01" };
    expect(sameVatAdjustFilter(base, { ...base, partnerto: "  " })).toBe(true);
    expect(sameVatAdjustFilter(base, { ...base, invdateto: "2026-01-31" })).toBe(false);
  });

  it("tracks edits per journal row and drops a row that returns to its loaded values", () => {
    const first = row(), second = row({ journalid: "j2", id: "v9", version: 5 });
    let edits = setVatAdjustEdit(none, first, { period: "2026-03" });
    expect(edits.get(vatAdjustKey(first))).toMatchObject({ year: 2026, month: 3, claim: 1 });
    edits = setVatAdjustEdit(edits, second, { claim: 3 });
    expect(edits.size).toBe(2);
    edits = setVatAdjustEdit(edits, first, { period: "2026-01" });
    expect(edits.has(vatAdjustKey(first))).toBe(false);
    expect(edits.has(vatAdjustKey(second))).toBe(true);
    expect(none.size).toBe(0);
  });

  it("clears the period when the empty option is chosen and keeps the other field", () => {
    const edits = setVatAdjustEdit(none, row(), { period: "" });
    expect(edits.get("j1:v1")).toMatchObject({ year: 0, month: 0, claim: 1 });
    expect(vatAdjustMergedRow(row(), edits.get("j1:v1"))).toMatchObject({ tax_period_year: undefined, tax_period_month: undefined, claim_status: 1 });
  });

  it("sends the loaded version of each row with the chosen period and claim status", () => {
    const edits = setVatAdjustEdit(setVatAdjustEdit(none, row(), { period: "2026-02" }), row({ journalid: "j2", id: "v9", version: 5 }), { claim: 2 });
    expect(vatAdjustItems(edits)).toEqual([
      { journalid: "j1", version: 3, vatid: "v1", taxperiodyear: 2026, taxperiodmonth: 2, claimstatus: 1 },
      { journalid: "j2", version: 5, vatid: "v9", taxperiodyear: 2026, taxperiodmonth: 1, claimstatus: 2 },
    ]);
    expect(vatAdjustRowsCommand(edits, "  ย้ายงวดตามใบกำกับ ")).toMatchObject({ resource: "vat-adjustments", action: "update", reason: "ย้ายงวดตามใบกำกับ", vatadjust: { mode: "rows" } });
  });

  it("flags rows the backend would refuse before saving", () => {
    const late = setVatAdjustEdit(none, row(), { period: "2026-09" });
    expect(vatAdjustProblems(late).map((item) => item.kind)).toEqual(["exceeded"]);
    const before = setVatAdjustEdit(none, row({ tax_invoice_date: "2026-03-10" }), { period: "2026-02" });
    expect(vatAdjustProblems(before).map((item) => item.kind)).toEqual(["before"]);
    const missing = setVatAdjustEdit(none, row({ claim_status: 3, tax_period_year: undefined, tax_period_month: undefined }), { claim: 1 });
    expect(vatAdjustProblems(missing).map((item) => item.kind)).toEqual(["missing_period"]);
    const fine = setVatAdjustEdit(none, row(), { period: "2026-03" });
    expect(vatAdjustProblems(fine)).toEqual([]);
  });

  it("limits the invoice-date button to the shown filter and its row count", () => {
    const filter = { ...EMPTY_VAT_ADJUST_FILTER, invdatefrom: "2026-01-01", invdateto: "2026-01-31", partnerfrom: " AP001" };
    expect(vatAdjustInvoiceDateCommand(filter, 12, " ตั้งงวดตามวันที่ใบกำกับ ")).toEqual({
      resource: "vat-adjustments", action: "update", reason: "ตั้งงวดตามวันที่ใบกำกับ",
      vatadjust: { mode: "invoice_date", filter: { invdatefrom: "2026-01-01", invdateto: "2026-01-31", partnerfrom: "AP001" }, expectedcount: 12 },
    });
  });

  it("renders the filters, shortcuts and the first-step hint before any search", () => {
    const html = renderToStaticMarkup(createElement(GLVatAdjust, { route: "/gl/purchasevatadjust" }));
    for (const text of ["วันที่เอกสาร", "วันที่ใบกำกับภาษี", "รหัสเจ้าหนี้", "ประมวลผล (F7)", "บันทึก (F11)", "ตั้งวันที่ยื่นภาษี = วันที่ใบกำกับภาษี", "ล้างการแก้ไข"]) expect(html).toContain(text);
    expect(html).toContain("aria-required=\"true\"");
    expect(html).not.toContain("<table");
  });

  it("has no hard-coded Thai outside tr() fallbacks and comments", () => {
    const source = readFileSync(new URL("./gl-vat-adjust.tsx", import.meta.url), "utf8")
      .replace(/\/\*[\s\S]*?\*\//g, "").replace(/\/\/.*$/gm, "")
      .replace(/tr\("[a-z0-9_]+", "[^"]*"\)/g, "");
    expect(source).not.toMatch(/[฀-๿]/);
  });
});
