import { describe, expect, it } from "vitest";
import { emptyJournal, emptyLine, type GLJournal, type GLJournalTemplate } from "./general-ledger";
import type { GLDetailVat } from "./gl-journal-details";
import { newVatRow, nextJournalAfterSave } from "./gl-journal-entry";
import {
  JOURNAL_TEMPLATE_NAME_MAX, journalFromTemplate, journalJumpKey, journalTemplateName, journalTemplateNameProblem, jumpRowIndex,
  templateBodyFromJournal,
} from "./gl-journal-templates";

const rentJournal = (patch: Partial<GLJournal> = {}): GLJournal => ({
  ...emptyJournal("JV", "manual", [], "00000"), id: "J1", version: 4, docno: "JV6901-0007", date: "2026-01-31", fiscalyear: "2026", status: "posted",
  bookcode: "JV", reference: "สัญญาเช่า-01", refdate: "2026-01-01", description: "ค่าเช่าสำนักงานเดือนมกราคม",
  lines: [
    { ...emptyLine(), accountcode: " 5300 ", description: " ค่าเช่า ", departmentcode: "D01", projectcode: "P01", debit: "15000.00" },
    { ...emptyLine(), accountcode: "1100", description: "จ่ายเงินสด", credit: "15000.00" },
    { ...emptyLine() },
  ],
  ...patch,
});
const purchaseVat = (patch: Partial<GLDetailVat> = {}): GLDetailVat => ({ ...newVatRow("2026-01-31", 1), tax_invoice_no: "IV-001", base_amount: "15000.00", vat_amount: "1050.00", ...patch });
const template = (patch: Partial<GLJournalTemplate["template"]> = {}): GLJournalTemplate => ({
  id: "T1", version: 1, code: "T0001", name: "ค่าเช่าสำนักงานรายเดือน", isactive: true,
  template: { bookcode: "JV", reference: "สัญญาเช่า-01", refdate: "2026-01-01", description: "ค่าเช่าสำนักงาน", lines: [{ accountcode: "5300", description: "ค่าเช่า", departmentcode: "D01", projectcode: "P01", cashflow: "" }, { accountcode: "1100" }], ...patch },
});

describe("template name (Champ Recur Name dialog)", () => {
  it("is trimmed, required and at most 40 characters like BCTrans.RecurName VARCHAR(40)", () => {
    expect(JOURNAL_TEMPLATE_NAME_MAX).toBe(40);
    expect(journalTemplateName("  ค่าเช่าสำนักงาน ")).toBe("ค่าเช่าสำนักงาน");
    expect(journalTemplateNameProblem("   ")).toEqual({ problem: "required" });
    expect(journalTemplateNameProblem("ก".repeat(40))).toBeNull();
    expect(journalTemplateNameProblem("ก".repeat(41))).toEqual({ problem: "too_long", length: 41 });
  });
  it("counts Thai tone marks and vowels like the backend counts runes (code points, not characters)", () => {
    expect(journalTemplateNameProblem("กิ".repeat(20))).toBeNull();
    expect(journalTemplateNameProblem("กิ".repeat(21))).toEqual({ problem: "too_long", length: 42 });
  });
});

describe("templateBodyFromJournal (Recur button)", () => {
  it("keeps book, reference, description and the account/detail/dimension of the lines that have an account", () => {
    expect(templateBodyFromJournal(rentJournal())).toEqual({
      bookcode: "JV", reference: "สัญญาเช่า-01", refdate: "2026-01-01", description: "ค่าเช่าสำนักงานเดือนมกราคม",
      lines: [
        { accountcode: "5300", description: "ค่าเช่า", departmentcode: "D01", projectcode: "P01", cashflow: "" },
        { accountcode: "1100", description: "จ่ายเงินสด", departmentcode: "", projectcode: "", cashflow: "" },
      ],
    });
  });
  it("never stores the document number, dates of the voucher, amounts or invoice figures", () => {
    const stored = JSON.stringify(templateBodyFromJournal(rentJournal({ details: { vats: [purchaseVat()] } })));
    for (const forbidden of ["JV6901-0007", "2026-01-31", "15000", "1050", "IV-001", "debit", "credit", "docno"]) expect(stored).not.toContain(forbidden);
  });
  it("records the tax type only when every invoice of the voucher has the same one", () => {
    expect(templateBodyFromJournal(rentJournal({ details: { vats: [purchaseVat()] } })).taxtype).toBe(1);
    expect(templateBodyFromJournal(rentJournal({ details: { vats: [purchaseVat({ tax_type: 2 })] } })).taxtype).toBe(2);
    expect(templateBodyFromJournal(rentJournal({ details: { vats: [purchaseVat(), purchaseVat({ tax_type: 2 })] } })).taxtype).toBeUndefined();
    expect(templateBodyFromJournal(rentJournal()).taxtype).toBeUndefined();
    expect("taxtype" in templateBodyFromJournal(rentJournal())).toBe(false);
  });
  it("has no lines for a voucher without account codes (the screen refuses to save it)", () => {
    expect(templateBodyFromJournal(emptyJournal("JV")).lines).toEqual([]);
  });
});

describe("journalFromTemplate (the '...' picker)", () => {
  const base = { ...emptyJournal("GV", "manual", [], "00001"), date: "2026-02-28", fiscalyear: "2026", docno: "typed-by-mistake", branchcode: "00001" };

  it("fills book, reference, reference date, description and lines; document number empty and every amount 0", () => {
    const filled = journalFromTemplate(template(), base, ["JV", "GV"]);
    expect(filled).toMatchObject({ docno: "", bookcode: "JV", reference: "สัญญาเช่า-01", refdate: "2026-01-01", description: "ค่าเช่าสำนักงาน" });
    expect(filled.lines).toEqual([
      { ...emptyLine(), accountcode: "5300", description: "ค่าเช่า", departmentcode: "D01", projectcode: "P01" },
      { ...emptyLine(), accountcode: "1100" },
    ]);
    expect(filled.lines.every((line) => line.debit === "0" && line.credit === "0")).toBe(true);
  });
  it("keeps the document date, fiscal year and branch of the screen (Champ does not load the template's date)", () => {
    const filled = journalFromTemplate(template(), base, ["JV"]);
    expect([filled.date, filled.fiscalyear, filled.branchcode, filled.kind, filled.status]).toEqual(["2026-02-28", "2026", "00001", "manual", "draft"]);
    expect(filled.id).toBeUndefined();
  });
  it("keeps the screen's book when the template book is no longer usable, and the screen's book is never invented", () => {
    expect(journalFromTemplate(template(), base, ["GV"]).bookcode).toBe("GV");
    expect(journalFromTemplate(template({ bookcode: "" }), base, ["JV", "GV"]).bookcode).toBe("GV");
  });
  it("opens the tax box with the template's tax type, empty invoice number and no amount; none when the template has no tax", () => {
    const withTax = journalFromTemplate(template({ taxtype: 1 }), { ...base, details: { vats: [purchaseVat({ tax_type: 2 })] } }, ["JV"]);
    expect(withTax.details?.vats).toHaveLength(1);
    expect(withTax.details?.vats?.[0]).toMatchObject({ tax_type: 1, tax_invoice_no: "", tax_invoice_date: "2026-02-28", base_amount: "0", vat_rate: "" });
    const sale = journalFromTemplate(template({ taxtype: 2 }), base, ["JV"]);
    expect(sale.details?.vats?.[0]).toMatchObject({ tax_type: 2, tax_invoice_no: "" });
    expect(journalFromTemplate(template(), { ...base, details: { vats: [purchaseVat()] } }, ["JV"]).details).toBeUndefined();
  });
  it("always leaves at least two lines to key", () => {
    expect(journalFromTemplate(template({ lines: [{ accountcode: "5300" }] }), base, ["JV"]).lines.map((line) => line.accountcode)).toEqual(["5300", ""]);
    expect(journalFromTemplate(template({ lines: [] }), base, ["JV"]).lines).toHaveLength(2);
  });
  it("is the same pattern the post-save 'keep screen' option leaves: accounts, details and dimensions kept, amounts cleared", () => {
    const saved = rentJournal();
    const kept = nextJournalAfterSave(saved, "keep", emptyJournal("JV"));
    const viaTemplate = journalFromTemplate({ ...template(), template: templateBodyFromJournal(saved) }, emptyJournal("JV"), ["JV"]);
    const pattern = (journal: GLJournal) => journal.lines.filter((line) => line.accountcode).map(({ accountcode, description, departmentcode, projectcode, debit, credit }) => ({ accountcode, description, departmentcode, projectcode, debit, credit }));
    expect(pattern(viaTemplate)).toEqual(pattern(kept).map((line) => ({ ...line, accountcode: line.accountcode.trim(), description: line.description.trim() })));
  });
});

describe("journalJumpKey (OnToDebit / OnToCredit)", () => {
  it("is Alt+D for debit and Alt+C for credit, also on a Thai keyboard layout (key is a Thai letter, code is KeyD/KeyC)", () => {
    expect(journalJumpKey({ key: "d", altKey: true })).toBe("debit");
    expect(journalJumpKey({ key: "C", altKey: true })).toBe("credit");
    expect(journalJumpKey({ key: "ก", code: "KeyD", altKey: true })).toBe("debit");
    expect(journalJumpKey({ key: "แ", code: "KeyC", altKey: true })).toBe("credit");
  });
  it("ignores plain letters, other modifiers (Ctrl+Alt is AltGr), other Alt keys and handled events", () => {
    expect(journalJumpKey({ key: "d" })).toBeNull();
    expect(journalJumpKey({ key: "d", altKey: true, ctrlKey: true })).toBeNull();
    expect(journalJumpKey({ key: "d", altKey: true, metaKey: true })).toBeNull();
    expect(journalJumpKey({ key: "d", altKey: true, shiftKey: true })).toBeNull();
    expect(journalJumpKey({ key: "n", altKey: true, code: "KeyN" })).toBeNull();
    expect(journalJumpKey({ key: "d", altKey: true, defaultPrevented: true })).toBeNull();
  });
});

describe("jumpRowIndex", () => {
  it("uses the focused row, else the last row worked in, and never leaves the table", () => {
    expect(jumpRowIndex(2, 0, 5)).toBe(2);
    expect(jumpRowIndex(null, 3, 5)).toBe(3);
    expect(jumpRowIndex(null, 9, 5)).toBe(4);
    expect(jumpRowIndex(null, -1, 5)).toBe(0);
    expect(jumpRowIndex(null, Number.NaN, 5)).toBe(0);
    expect(jumpRowIndex(null, 0, 0)).toBe(0);
  });
});
