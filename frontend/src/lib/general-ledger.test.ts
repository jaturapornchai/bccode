import { describe, expect, it } from "vitest";
import { activeJournalBooks, amountString, amountUnits, csvCell, defaultJournalBookCode, fiscalYearForDate, journalBookPayload, journalBookProblem, journalBookTypeLabels, normalizeJournalLines, untypedJournalBooks, validateJournalBook, workspaceBranchCode, type GLJournalBook, emptyAccount, emptyFiscalYear, emptyJournal, emptyLine, formatAmount, GL_MENU_ITEMS, isGeneralLedgerRoute, journalTotals, reportCsv, validateJournal, evaluateStatementFormula, generateStarterTemplates } from "./general-ledger";

const accounts = [ { ...emptyAccount(), accountcode: "A", names: [{ code: "th", name: "เงินสด" }] }, { ...emptyAccount(), accountcode: "B", names: [{ code: "th", name: "ทุน" }] } ];
const year = { ...emptyFiscalYear(), code: "FY", startdate: "2026-01-01", enddate: "2026-12-31", currency: "THB", scale: 2 };
const journal = { ...emptyJournal(), docno: "JV-1", date: "2026-09-11", fiscalyear: "FY", currency: "THB", description: "ตรวจยอด", lines: [{ ...emptyLine(), accountcode: "A", debit: "0.3" }, { ...emptyLine(), accountcode: "B", credit: "0.3" }] };
describe("general ledger exact accounting helpers", () => {
  it("adds 0.1 + 0.2 exactly and preserves JSON decimal strings", () => {
    const sum = amountUnits("0.1") + amountUnits("0.2");
    expect(sum).toBe(amountUnits("0.3"));
    expect(amountString(sum, 2)).toBe("0.30");
    expect(JSON.parse(JSON.stringify(journal)).lines[0].debit).toBe("0.3");
  });
  it("preserves very large values and all eight decimals without rounding", () => {
    const value = "99999999999999999999999999.99999999";
    expect(amountString(amountUnits(value))).toBe(value);
    expect(formatAmount("0.12345678", 2)).toBe("0.12345678");
    expect(amountString(amountUnits("-0.01"), 2)).toBe("-0.01");
    expect(formatAmount("1234567.89")).toBe("1,234,567.89");
    expect(formatAmount("999999999999999999999999999999.99")).toBe("999,999,999,999,999,999,999,999,999,999.99");
  });
  it.each(["", "1,000", "1e2", "NaN", "Infinity", "01", "1.123456789", "+1", " 1", "1."])("rejects unsafe or ambiguous amount %s", (value) => { expect(() => amountUnits(value)).toThrow(); });
  it("validates both sides, precision boundary and fiscal year", () => {
    expect(validateJournal(journal, year, accounts)).toBeNull();
    expect(journalTotals(journal.lines).difference).toBe(0n);
    expect(validateJournal({ ...journal, lines: [{ ...journal.lines[0], credit: "1" }, journal.lines[1]] }, year, accounts)).toContain("ด้านเดียว");
    expect(validateJournal({ ...journal, lines: [{ ...journal.lines[0], debit: "0.301" }, { ...journal.lines[1], credit: "0.301" }] }, year, accounts)).toContain("ทศนิยม");
    expect(validateJournal({ ...journal, date: "2027-01-01" }, year, accounts)).toContain("วันที่ภายในปีบัญชี");
    expect(validateJournal(journal, { ...year, closed: true }, accounts)).toContain("ปีบัญชี");
    expect(validateJournal(journal, year, [{ ...accounts[0], allowposting: false }, accounts[1]])).toContain("ลงรายการได้");
    expect(validateJournal({ ...journal, lines: [journal.lines[0], { ...journal.lines[1], credit: "0.2" }] }, year, accounts)).toContain("เท่ากัน");
  });
  it("protects CSV formulas and exports amounts as numeric values without apostrophes", () => {
    expect(csvCell("=HYPERLINK(\"https://bad\")")).toBe('"\'=HYPERLINK(""https://bad"")"');
    expect(csvCell("\t+cmd")).toBe('"\'\t+cmd"');
    const csv = reportCsv({ columns: [{ key: "name", label: "ชื่อ" }, { key: "amount", label: "จำนวนเงิน", amount: true }], rows: [{ name: "=1+1", amount: "999999999999999999.01" }, { name: "ติดลบ", amount: "-1234.50" }], totals: {}, totalrows: 2, warnings: [], asof: "", sequence: 1 });
    expect(csv).toContain('"\'=1+1","999999999999999999.01"');
    expect(csv).toContain('"ติดลบ","-1234.50"');
    expect(csv).not.toContain("'-1234.50");
  });
  it("provides default level 1 and supports 1 to 12 in account models", () => {
    const defaultAcc = emptyAccount();
    expect(defaultAcc.level).toBe(1);
    const childAcc = { ...defaultAcc, accountcode: "1101", parentaccountcode: "1100", level: 2 };
    expect(childAcc.level).toBe(2);
  });
  it("covers the exact 23 GL menu routes (Champ parity 2026-09-19)", () => {
    expect(GL_MENU_ITEMS).toHaveLength(23);
    expect(new Set(GL_MENU_ITEMS.map((item) => item.route)).size).toBe(23);
    expect(GL_MENU_ITEMS.every((item) => isGeneralLedgerRoute(item.route))).toBe(true);
    expect(isGeneralLedgerRoute("/report/ledger?accountcode=A")).toBe(true);
    expect(isGeneralLedgerRoute("/gl/journals")).toBe(true);
    expect(isGeneralLedgerRoute("/gl/unposting")).toBe(true);
    expect(isGeneralLedgerRoute("/gl/journal/jv")).toBe(true);
    expect(isGeneralLedgerRoute("/employee")).toBe(false);
  });

  it("evaluates statement formulas accurately with range sums and arithmetic", () => {
    const rowValues = new Map<number, bigint>([
      [10, 10000000000n], // 100.00
      [20, 5000000000n],  // 50.00
      [30, 2500000000n],  // 25.00
    ]);
    expect(evaluateStatementFormula("R10 + R20", rowValues)).toBe(15000000000n);
    expect(evaluateStatementFormula("R10 - R20", rowValues)).toBe(5000000000n);
    expect(evaluateStatementFormula("SUM(R10:R30)", rowValues)).toBe(17500000000n);
    expect(evaluateStatementFormula("SUM(10..30)", rowValues)).toBe(17500000000n);
    expect(evaluateStatementFormula("(R10 - R20) * 2", rowValues)).toBe(10000000000n);
    // Prevents self-reference recursion
    expect(evaluateStatementFormula("R10", rowValues, 10)).toBe(0n);
  });

  it("provides 4 standard starter templates with valid rows and structure", () => {
    const templates = generateStarterTemplates();
    expect(templates).toHaveLength(4);
    const bs = templates.find((t) => t.statementtype === "balance_sheet")!;
    expect(bs).toBeDefined();
    expect(bs.rows.length).toBeGreaterThan(15);
    const pnl = templates.find((t) => t.statementtype === "pnl")!;
    expect(pnl).toBeDefined();
    expect(pnl.rows.length).toBeGreaterThan(10);
  });
});

describe("journal amounts: blank means zero, errors name the line", () => {
  const base = { ...emptyJournal(), docno: "JV-2", date: "2026-09-11", fiscalyear: "FY", description: "บันทึกค่าไฟฟ้า" };
  it("accepts a line whose untouched side is blank or 0 (never fails on an untouched zero)", () => {
    expect(validateJournal({ ...base, lines: [{ ...emptyLine(), accountcode: "A", debit: "1500.00", credit: "" }, { ...emptyLine(), accountcode: "B", debit: "0", credit: "1500.00" }] }, year, accounts)).toBeNull();
    expect(normalizeJournalLines([{ ...emptyLine(), debit: "", credit: " " }])[0]).toMatchObject({ debit: "0", credit: "0" });
  });
  it("rejects a negative amount with the line number and the fix", () => {
    const problem = validateJournal({ ...base, lines: [{ ...emptyLine(), accountcode: "A", debit: "1500", credit: "0" }, { ...emptyLine(), accountcode: "B", debit: "-1500", credit: "0" }] }, year, accounts);
    expect(problem).toContain("บรรทัดที่ 2");
    expect(problem).toContain("เครดิต");
  });
  it("rejects text in an amount with the line number and the column", () => {
    const problem = validateJournal({ ...base, lines: [{ ...emptyLine(), accountcode: "A", debit: "1,500", credit: "0" }, { ...emptyLine(), accountcode: "B", debit: "0", credit: "1500" }] }, year, accounts);
    expect(problem).toContain("บรรทัดที่ 1");
    expect(problem).toContain("เดบิต");
  });
});

describe("new journal defaults", () => {
  const years = [
    { ...emptyFiscalYear(), code: "2568", startdate: "2025-01-01", enddate: "2025-12-31", closed: true },
    { ...emptyFiscalYear(), code: "2569", startdate: "2026-01-01", enddate: "2026-12-31" },
    { ...emptyFiscalYear(), code: "2569-OLD", startdate: "2026-01-01", enddate: "2026-12-31", isactive: false },
  ];
  it("pre-selects the active, open fiscal year that contains the date", () => {
    expect(fiscalYearForDate(years, "2026-09-24")).toBe("2569");
    expect(fiscalYearForDate(years, "2025-06-01")).toBe("");
    expect(fiscalYearForDate(years, "2027-01-01")).toBe("");
  });
  it("emptyJournal fills the fiscal year for today's date when one is open", () => {
    const today = emptyJournal("JV", "manual", [{ ...emptyFiscalYear(), code: "ALL", startdate: "2000-01-01", enddate: "2999-12-31" }]);
    expect(today.fiscalyear).toBe("ALL");
    expect(today.lines.every((line) => line.debit === "0" && line.credit === "0")).toBe(true);
  });
  // adversarial review 2026-09-24: a company-wide session rejects a blank branch (journal_branch_required)
  it("new journals start in the workspace branch instead of a blank branch", () => {
    const workspace = JSON.stringify({ shop: { holdingcode: "rungrueng" }, company: { code: "01" }, branch: { guidfixed: "00000", code: "00000" } });
    expect(workspaceBranchCode(workspace)).toBe("00000");
    expect(workspaceBranchCode(JSON.stringify({ branch: { guidfixed: "B02" } }))).toBe("B02");
    expect(workspaceBranchCode(JSON.stringify({ branch: null }))).toBe("");
    expect(workspaceBranchCode(null)).toBe("");
    expect(workspaceBranchCode("not json")).toBe("");
    expect(emptyJournal("JV", "manual", [], workspaceBranchCode(workspace)).branchcode).toBe("00000");
    expect(emptyJournal("JV").branchcode).toBe("");
  });
});

describe("journal books are user-defined master data (booktype drives behaviour, never the code)", () => {
  const books: GLJournalBook[] = [
    { id: "1", code: "UV", name: "สมุดรายวันซื้อ", booktype: 5, isactive: true },
    { id: "2", code: "สมุดขาย1", name: "สมุดรายวันขายหน้าร้าน", booktype: 4, isactive: true },
    { id: "3", code: "JV", name: "สมุดรายวันทั่วไป", booktype: 1, isactive: true },
    { id: "4", code: "OLD", name: "สมุดเลิกใช้", booktype: 1, isactive: false },
    { id: "5", code: "NOTYPE", name: "สมุดยังไม่กำหนดประเภท", booktype: 0, isactive: true },
  ];
  it("lists only active books with a type, ordered by type", () => {
    expect(activeJournalBooks(books).map((book) => book.code)).toEqual(["JV", "สมุดขาย1", "UV"]);
  });
  // ป้ายเตือนหน้ากำหนดสมุด: สมุดเดิมที่เปิดใช้แต่ยังไม่มีประเภท (สมุดปิดใช้/ถูกลบไม่ต้องเตือน)
  it("lists active books without a type for the journal-books notice", () => {
    expect(untypedJournalBooks([...books, { id: "6", code: "GONE", name: "สมุดที่ลบแล้ว", booktype: 0, isactive: true, isdeleted: true }, { id: "7", code: "A1", name: "สมุดเก่า", isactive: true }]).map((book) => book.code)).toEqual(["A1", "NOTYPE"]);
  });
  it("defaults to the requested book when usable, else the first general book", () => {
    expect(defaultJournalBookCode(books, "UV")).toBe("UV");
    expect(defaultJournalBookCode(books, "OLD")).toBe("JV");
    expect(defaultJournalBookCode([], "JV")).toBe("");
  });
  it("requires an existing active book with a type for new journals, but keeps an unchanged book on edit", () => {
    expect(journalBookProblem("", books, undefined)).toContain("กรุณาเลือกสมุดรายวัน");
    expect(journalBookProblem("XX", books, undefined)).toContain("กรุณาเลือกสมุดรายวัน");
    expect(journalBookProblem("OLD", books, undefined)).toContain("ปิดใช้งาน");
    expect(journalBookProblem("NOTYPE", books, undefined)).toContain("ยังไม่กำหนดประเภท");
    expect(journalBookProblem("OLD", books, "OLD")).toBeNull();
    expect(journalBookProblem("OLD", books, "JV")).toContain("ปิดใช้งาน");
    expect(journalBookProblem("UV", books, undefined)).toBeNull();
  });
  it("validates the book form per field (code ≤ 15 characters counted as runes, Thai name, type 1–6)", () => {
    const ok: GLJournalBook = { code: "สมุดขาย1", name: "สมุดรายวันขาย", nameen: "", booktype: 4, isactive: true };
    expect(validateJournalBook(ok)).toBeNull();
    expect(validateJournalBook({ ...ok, code: " " })?.field).toBe("code");
    expect(validateJournalBook({ ...ok, code: "ก".repeat(16) })?.field).toBe("code");
    expect(validateJournalBook({ ...ok, code: "ก".repeat(15) })).toBeNull();
    expect(validateJournalBook({ ...ok, name: "" })?.field).toBe("name");
    expect(validateJournalBook({ ...ok, booktype: 0 })?.field).toBe("booktype");
    expect(validateJournalBook({ ...ok, booktype: 7 })?.field).toBe("booktype");
  });
  it("lets a legacy book without a type be renamed or deactivated as it is (backend allowUntyped)", () => {
    const legacy: GLJournalBook = { id: "b-ap", code: "AP", name: "สมุดรายวันซื้อเชื่อ", nameen: "", booktype: 0, isactive: false };
    expect(validateJournalBook(legacy, undefined, true)).toBeNull();
    expect(validateJournalBook(legacy)?.field).toBe("booktype");
    expect(validateJournalBook({ ...legacy, booktype: 7 }, undefined, true)?.field).toBe("booktype");
    expect(validateJournalBook({ ...legacy, name: "" }, undefined, true)?.field).toBe("name");
  });
  it("sends only the journal-book contract fields", () => {
    const payload = journalBookPayload({ id: "b1", version: 3, code: " PV ", name: " สมุดรายวันจ่ายเงิน ", nameen: " Payments ", booktype: 2, isactive: false, bookcode: "x" } as GLJournalBook);
    expect(payload).toEqual({ id: "b1", version: 3, code: "PV", name: "สมุดรายวันจ่ายเงิน", nameen: "Payments", booktype: 2, isactive: false });
  });
  it("labels the six book types from the spec", () => {
    expect(Object.keys(journalBookTypeLabels)).toEqual(["1", "2", "3", "4", "5", "6"]);
  });
});
