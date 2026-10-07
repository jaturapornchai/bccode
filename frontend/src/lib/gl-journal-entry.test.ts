import { describe, expect, it } from "vitest";
import { emptyJournal, emptyLine, type GLJournal, type GLLine } from "./general-ledger";
import type { GLDetailVat } from "./gl-journal-details";
import {
  JOURNAL_SAVE_OPTIONS, dropBlankTaxBoxRow, isBlankVatRow, journalFunctionKey, journalSaveOptionStorageKey, linesHaveAmounts,
  newVatRow, nextJournalAfterSave, parseJournalSaveOption, patchTaxBox, setTaxBoxType, shouldOfferNewDocNo, taxBoxCalcInput, taxBoxView, taxLinesFromCalc, vatErrorTarget,
} from "./gl-journal-entry";

const vat = (patch: Partial<GLDetailVat> = {}): GLDetailVat => ({
  id: "V1", tax_type: 1, document_type: 1, tax_invoice_no: "IV6909-001", tax_invoice_date: "2026-09-10", tax_period_year: 2026, tax_period_month: 9,
  partner_name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", base_amount: "10000.00", zero_rate_amount: "0", exempt_amount: "0", vat_rate: "7", vat_amount: "700.00",
  claim_status: 1, remark: "ซื้อปูนซีเมนต์ปอร์ตแลนด์", ...patch,
});
const line = (accountcode: string, debit: string, credit: string, description = ""): GLLine => ({ ...emptyLine(), accountcode, debit, credit, description, departmentcode: "D01", projectcode: "P01" });

describe("journal save option (Champ combo)", () => {
  it("lists the two Champ options in Champ order with the exact Thai texts", () => {
    expect(JOURNAL_SAVE_OPTIONS.map((option) => [option.value, option.th])).toEqual([["keep", "บันทึกแล้วไม่ล้างหน้าจอ"], ["new", "บันทึกแล้วเพิ่มรายการใหม่"]]);
  });
  it("remembers the option per user and falls back to Champ's first option", () => {
    expect(journalSaveOptionStorageKey(" Somchai ")).toBe("bc_gl_journal_save_option:somchai");
    expect(journalSaveOptionStorageKey(undefined)).toBe("bc_gl_journal_save_option:anonymous");
    expect(parseJournalSaveOption("new")).toBe("new");
    expect(parseJournalSaveOption("keep")).toBe("keep");
    expect(parseJournalSaveOption(null)).toBe("keep");
    expect(parseJournalSaveOption("stay")).toBe("keep");
  });
});

describe("nextJournalAfterSave", () => {
  const saved: GLJournal = {
    ...emptyJournal("PV", "manual", [], "00000"), id: "J1", version: 3, docno: "PV6909-0007", date: "2026-09-10", fiscalyear: "2569", status: "draft",
    reference: "IV6909-001", refdate: "2026-09-09", description: "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. จำนวน 500 ถุง เป็นเงินเชื่อ", branchcode: "00001",
    lines: [line("11310", "10000.00", "0", "ปูนซีเมนต์"), line("11410", "700.00", "0"), line("21110", "0", "10700.00")],
    details: { vats: [vat()] },
  };
  const blank = emptyJournal("JV", "manual", [], "00000");

  it("'new' returns the blank voucher untouched", () => {
    expect(nextJournalAfterSave(saved, "new", blank)).toBe(blank);
  });

  it("'keep' keeps header, accounts and line details but clears doc number, amounts and invoice", () => {
    const next = nextJournalAfterSave(saved, "keep", blank);
    expect(next.id).toBeUndefined();
    expect(next.version).toBeUndefined();
    expect(next.docno).toBe("");
    expect(next.status).toBe(blank.status);
    expect([next.bookcode, next.date, next.fiscalyear, next.reference, next.refdate, next.branchcode]).toEqual(["PV", "2026-09-10", "2569", "IV6909-001", "2026-09-09", "00001"]);
    expect(next.description).toBe(saved.description);
    expect(next.lines.map((item) => [item.accountcode, item.description, item.departmentcode, item.projectcode, item.debit, item.credit])).toEqual([
      ["11310", "ปูนซีเมนต์", "D01", "P01", "0", "0"], ["11410", "", "D01", "P01", "0", "0"], ["21110", "", "D01", "P01", "0", "0"],
    ]);
    const [kept] = next.details?.vats ?? [];
    expect(kept.id).not.toBe("V1");
    expect([kept.tax_type, kept.tax_invoice_date, kept.vat_rate, kept.remark, kept.claim_status]).toEqual([1, "2026-09-10", "7", "ซื้อปูนซีเมนต์ปอร์ตแลนด์", 1]);
    expect([kept.tax_invoice_no, kept.base_amount, kept.vat_amount, kept.partner_name]).toEqual(["", "0", undefined, ""]);
    expect(isBlankVatRow(kept)).toBe(true);
  });

  it("'keep' does not carry a multi-invoice tax table or the other detail sections", () => {
    const next = nextJournalAfterSave({ ...saved, details: { vats: [vat(), vat({ id: "V2" })], withholdings: [] } }, "keep", blank);
    expect(next.details).toBeUndefined();
  });
});

describe("tax box", () => {
  it("new VAT rows never get a default rate", () => {
    const row = newVatRow("2026-09-10", 2);
    expect(row.vat_rate).toBe("");
    expect(row.claim_status).toBeUndefined();
    expect([row.tax_period_year, row.tax_period_month]).toEqual([2026, 9]);
    expect(newVatRow("2026-09-10", 1).claim_status).toBe(1);
  });

  it("views none / single / multi", () => {
    expect(taxBoxView(undefined)).toEqual({ mode: "none" });
    expect(taxBoxView({ vats: [] })).toEqual({ mode: "none" });
    expect(taxBoxView({ vats: [vat()] })).toMatchObject({ mode: "single", row: { id: "V1" } });
    expect(taxBoxView({ vats: [vat(), vat({ id: "V2" })] })).toEqual({ mode: "multi", count: 2 });
  });

  it("choosing a type creates one row; 'no tax' removes it; multi is left to the tax table", () => {
    const created = setTaxBoxType(undefined, "2", "2026-09-10");
    expect(created?.vats).toHaveLength(1);
    expect(created?.vats?.[0]).toMatchObject({ tax_type: 2, vat_rate: "", tax_invoice_date: "2026-09-10" });
    expect(setTaxBoxType(created, "", "2026-09-10")?.vats).toEqual([]);
    expect(setTaxBoxType(undefined, "", "2026-09-10")).toBeUndefined();
    const multi = { vats: [vat(), vat({ id: "V2" })] };
    expect(setTaxBoxType(multi, "", "2026-09-10")).toBe(multi);
  });

  it("switching type keeps the row and applies the tax-table side effects", () => {
    const sale = setTaxBoxType({ vats: [vat({ claim_reason: "x" })] }, "2", "2026-10-01")?.vats?.[0];
    expect(sale).toMatchObject({ id: "V1", tax_type: 2, tax_invoice_no: "IV6909-001", tax_period_year: 2026, tax_period_month: 9 });
    expect(sale?.claim_status).toBeUndefined();
    expect(sale?.claim_reason).toBeUndefined();
    const purchase = setTaxBoxType({ vats: [vat({ tax_type: 2, claim_status: undefined })] }, "1", "2026-10-01")?.vats?.[0];
    expect(purchase).toMatchObject({ tax_type: 1, claim_status: 1 });
  });

  it("editing base or rate clears the stored tax so the backend recomputes; an explicit tax stays", () => {
    expect(patchTaxBox({ vats: [vat()] }, { base_amount: "20000.00" })?.vats?.[0].vat_amount).toBeUndefined();
    expect(patchTaxBox({ vats: [vat()] }, { vat_rate: "10" })?.vats?.[0].vat_amount).toBeUndefined();
    expect(patchTaxBox({ vats: [vat()] }, { base_amount: "100.00", vat_amount: "7.00" })?.vats?.[0]).toMatchObject({ base_amount: "100.00", vat_amount: "7.00" });
    expect(patchTaxBox({ vats: [vat()] }, { tax_invoice_no: "IV2" })?.vats?.[0].vat_amount).toBe("700.00");
    const multi = { vats: [vat(), vat({ id: "V2" })] };
    expect(patchTaxBox(multi, { remark: "x" })).toBe(multi);
  });

  it("a type-only box is blank and is dropped before saving; a filled one is kept", () => {
    const typeOnly = setTaxBoxType(undefined, "1", "2026-09-10");
    expect(isBlankVatRow(typeOnly!.vats![0])).toBe(true);
    expect(dropBlankTaxBoxRow(typeOnly)?.vats).toEqual([]);
    expect(dropBlankTaxBoxRow({ vats: [vat()] })?.vats).toHaveLength(1);
    expect(isBlankVatRow(vat({ tax_invoice_no: "", base_amount: "0.00", vat_amount: undefined, partner_name: "" }))).toBe(true);
    expect(isBlankVatRow(vat({ tax_invoice_no: "", base_amount: "0.01", vat_amount: undefined, partner_name: "" }))).toBe(false);
    expect(dropBlankTaxBoxRow(undefined)).toBeUndefined();
  });

  it("calc input needs a rate and a non-zero amount (no 7% default)", () => {
    expect(taxBoxCalcInput(undefined)).toEqual({ ok: false, problem: "no_row" });
    expect(taxBoxCalcInput({ vats: [vat({ vat_rate: " " })] })).toEqual({ ok: false, problem: "rate" });
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "0.00" })] })).toEqual({ ok: false, problem: "amount" });
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "10700.00", tax_type: 2 })] })).toEqual({ ok: true, amount: "10700.00", rate: "7", taxType: 2 });
  });
});

describe("taxLinesFromCalc (Champ OnTaxAuto amounts, no account codes)", () => {
  const calc = { base_amount: "10000.00", vat_amount: "700.00", total_amount: "10700.00" };
  it("purchase: Dr base · Dr tax · Cr total and clears every other amount", () => {
    const lines = [line("51110", "5.00", "0"), line("11410", "0", "0"), line("21110", "0", "3.00"), line("11120", "9.00", "0")];
    const next = taxLinesFromCalc(lines, 1, calc);
    expect(next.map((item) => [item.accountcode, item.debit, item.credit])).toEqual([["51110", "10000.00", "0"], ["11410", "700.00", "0"], ["21110", "0", "10700.00"], ["11120", "0", "0"]]);
  });
  it("sale: Dr total · Cr base · Cr tax, padding to three lines without guessing accounts", () => {
    const next = taxLinesFromCalc([line("11210", "0", "0")], 2, calc);
    expect(next.map((item) => [item.accountcode, item.debit, item.credit])).toEqual([["11210", "10700.00", "0"], ["", "0", "10000.00"], ["", "0", "700.00"]]);
  });
  it("detects existing amounts without floats", () => {
    expect(linesHaveAmounts([line("1", "0", "0"), line("2", "0.00", "")])).toBe(false);
    expect(linesHaveAmounts([line("1", "0", "0.01")])).toBe(true);
  });
});

describe("vatErrorTarget", () => {
  it("points a backend vat_ error at the single invoice row (fields the box lacks live in the F3 table)", () => {
    expect(vatErrorTarget({ vats: [vat()] }, "vat_partner_name_invalid", "partner_name")).toEqual({ section: "vats", row: 1, field: "partner_name" });
  });
  it("does not guess with several invoices, other error codes or no field", () => {
    expect(vatErrorTarget({ vats: [vat(), vat({ id: "V2" })] }, "vat_partner_name_invalid", "partner_name")).toBeNull();
    expect(vatErrorTarget({ vats: [vat()] }, "wht_payee_name_invalid", "partner_name")).toBeNull();
    expect(vatErrorTarget({ vats: [vat()] }, "vat_partner_name_invalid", "")).toBeNull();
    expect(vatErrorTarget(undefined, "vat_partner_name_invalid", "partner_name")).toBeNull();
  });
});

describe("shouldOfferNewDocNo (Champ: Document No is duplicated, run again?)", () => {
  const ctx = { isNew: true, typedDocNo: "JV6910-0001", bookFound: true };
  it("offers a new running number only for a typed number that the backend rejected as duplicate on a new voucher", () => {
    expect(shouldOfferNewDocNo("duplicate_code", ctx)).toBe(true);
  });
  it("keeps the plain error for other codes, edits, blank or unknown-book cases", () => {
    expect(shouldOfferNewDocNo("validation_failed", ctx)).toBe(false);
    expect(shouldOfferNewDocNo(undefined, ctx)).toBe(false);
    expect(shouldOfferNewDocNo("duplicate_code", { ...ctx, isNew: false })).toBe(false);
    expect(shouldOfferNewDocNo("duplicate_code", { ...ctx, typedDocNo: "  " })).toBe(false);
    expect(shouldOfferNewDocNo("duplicate_code", { ...ctx, bookFound: false })).toBe(false);
  });
});

describe("journalFunctionKey", () => {
  it("maps F2 / F3 / F5 without modifiers", () => {
    expect(journalFunctionKey({ key: "F2" })).toBe("list");
    expect(journalFunctionKey({ key: "F3" })).toBe("tax");
    expect(journalFunctionKey({ key: "F5" })).toBe("details");
  });
  it("leaves other keys, modified keys and handled events to the browser", () => {
    expect(journalFunctionKey({ key: "F9" })).toBeNull();
    expect(journalFunctionKey({ key: "F11" })).toBeNull();
    expect(journalFunctionKey({ key: "F5", ctrlKey: true })).toBeNull();
    expect(journalFunctionKey({ key: "F5", shiftKey: true })).toBeNull();
    expect(journalFunctionKey({ key: "F3", altKey: true })).toBeNull();
    expect(journalFunctionKey({ key: "F3", metaKey: true })).toBeNull();
    expect(journalFunctionKey({ key: "F2", defaultPrevented: true })).toBeNull();
  });
});
