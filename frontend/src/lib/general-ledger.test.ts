import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { BUDGET_PERIODS, budgetPeriodStarts, budgetPeriodsTotal } from "./general-ledger";
import { GL_RESOURCES, STATEMENT_NOTES_NPAES_BASIS, newStatementNote, starterStatementNotes, statementNoteHasText, statementNotesAfterReload, statementNotesFromRecord, statementNotesNeedReload } from "./general-ledger";
import { activeJournalBooks, amountString, amountUnits, csvCell, defaultJournalBookCode, fiscalYearForDate, journalBookPayload, journalBookProblem, journalBookTypeLabels, normalizeJournalLines, untypedJournalBooks, validateJournalBook, workspaceBranchCode, type GLJournalBook, emptyAccount, emptyFiscalYear, emptyJournal, emptyLine, formatAmount, GL_MENU_ITEMS, isGeneralLedgerRoute, journalTotals, reportCsv, validateJournal, generateStarterTemplates, emptyStatementTemplate, statementStarterReplaceNeedsConfirm, statementStarterReplacedCode, statementTemplateFromStarter, type StatementRow } from "./general-ledger";
import { statementIsPeriodic, statementSetDefaultSelection, statementSetRank, statementSetTemplates, type GLStatementTemplate, type StatementType } from "./general-ledger";

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
  it("lists budget period starts like the backend fiscalPeriodStarts (full, mid-month and short years)", () => {
    expect(budgetPeriodStarts({ startdate: "2026-01-01", enddate: "2026-12-31" })).toEqual(["2026-01-01", "2026-02-01", "2026-03-01", "2026-04-01", "2026-05-01", "2026-06-01", "2026-07-01", "2026-08-01", "2026-09-01", "2026-10-01", "2026-11-01", "2026-12-01"]);
    // ปีบัญชีเริ่ม ต.ค. ข้ามปีปฏิทิน และเริ่มกลางเดือน: งวด 1 = วันเริ่มจริง งวดต่อไปวันที่ 1
    expect(budgetPeriodStarts({ startdate: "2025-10-15", enddate: "2026-09-30" }).slice(0, 4)).toEqual(["2025-10-15", "2025-11-01", "2025-12-01", "2026-01-01"]);
    expect(budgetPeriodStarts({ startdate: "2025-10-15", enddate: "2026-09-30" })).toHaveLength(BUDGET_PERIODS);
    // ปีแรกสั้น (เริ่ม เม.ย.) มี 9 งวด
    expect(budgetPeriodStarts({ startdate: "2026-04-01", enddate: "2026-12-31" })).toHaveLength(9);
    expect(budgetPeriodStarts(undefined)).toEqual([]);
    expect(budgetPeriodStarts({ startdate: "", enddate: "" })).toEqual([]);
  });

  it("totals budget periods exactly with blanks as zero (no float drift)", () => {
    expect(amountString(budgetPeriodsTotal(["0.1", "0.2", "", "33333.33"]), 2)).toBe("33333.63");
    expect(budgetPeriodsTotal(Array(12).fill("8333.33"))).toBe(amountUnits("99999.96"));
    expect(() => budgetPeriodsTotal(["1,000"])).toThrow();
  });

  it("covers the exact 26 GL menu routes (Champ parity 2026-09-19 + 3 subledger reports 2026-09-25)", () => {
    expect(GL_MENU_ITEMS).toHaveLength(26);
    expect(new Set(GL_MENU_ITEMS.map((item) => item.route)).size).toBe(26);
    expect(GL_MENU_ITEMS.every((item) => isGeneralLedgerRoute(item.route))).toBe(true);
    expect(isGeneralLedgerRoute("/report/ledger?accountcode=A")).toBe(true);
    expect(isGeneralLedgerRoute("/gl/journals")).toBe(true);
    expect(isGeneralLedgerRoute("/gl/unposting")).toBe(true);
    expect(isGeneralLedgerRoute("/gl/journal/jv")).toBe(true);
    expect(isGeneralLedgerRoute("/employee")).toBe(false);
  });

  it("provides 4 standard starter templates with valid rows and structure", () => {
    const templates = generateStarterTemplates();
    expect(templates).toHaveLength(5);
    const bs = templates.find((t) => t.statementtype === "balance_sheet")!;
    expect(bs).toBeDefined();
    expect(bs.rows.length).toBeGreaterThan(15);
    const pnl = templates.find((t) => t.statementtype === "pnl")!;
    expect(pnl).toBeDefined();
    expect(pnl.rows.length).toBeGreaterThan(10);
  });

  // งบคำนวณที่ backend (statements_test.go); แม่แบบต้องไม่เดารหัสบัญชี (ผังบัญชีแต่ละกิจการไม่เหมือนกัน) และสูตรต้องอ้างแถวที่มีจริง
  it("starter templates carry no guessed account codes and formulas reference existing rows", () => {
    for (const template of generateStarterTemplates()) {
      const rownos = new Set(template.rows.map((row) => row.rowno));
      expect(rownos.size).toBe(template.rows.length);
      for (const row of template.rows) {
        for (const code of row.accountcodes ?? []) expect(code).toBe("__current_earnings__");
        for (const ref of (row.formula ?? "").match(/\d+/g) ?? []) expect(rownos.has(Number(ref))).toBe(true);
      }
    }
    const [bs, pnl] = generateStarterTemplates();
    expect([bs.name, bs.globalstyle.comparisontype, pnl.name, pnl.globalstyle.comparisontype]).toEqual(["งบฐานะการเงิน", "previous_year", "งบกำไรขาดทุน", "previous_year"]);
    expect(bs.rows.find((row) => row.rowno === 690)).toMatchObject({ rowtype: "subtotal", formula: "R540 + R680", style: { underline: "double" } });
  });

  // ต้นงวด/ปลายงวดต้องเป็นยอดคงเหลือจริง (backend statementBasisValue) ไม่ใช่ยอดเคลื่อนไหว; งบส่วนของผู้ถือหุ้นตามแบบ 2 หน้า 2-21
  it("cost, cash-flow and equity starters use explicit opening/closing bases", () => {
    const byCode = Object.fromEntries(generateStarterTemplates().map((template) => [template.code, template]));
    const basis = (code: string, rowno: number) => byCode[code].rows.find((row) => row.rowno === rowno)?.amountbasis;
    expect([20, 40, 90, 100, 120, 130].map((rowno) => basis("COGS-STMT", rowno))).toEqual(["opening", "closing", "opening", "closing", "opening", "closing"]);
    expect(basis("COGS-STMT", 30)).toBeUndefined();
    expect(basis("CASH-FLOW-IND", 170)).toBe("opening");
    expect(byCode["CASH-FLOW-IND"].rows.find((row) => row.rowno === 180)).toMatchObject({ formula: "R160 + R170", style: { underline: "double" } });
    const equity = byCode["EQ-DBD"];
    expect(equity.statementtype).toBe("equity");
    expect(equity.columns?.map((column) => column.title)).toEqual(["ทุนที่ชำระแล้ว", "ส่วนเกินมูลค่าหุ้น", "ส่วนเกิน (ต่ำกว่า) ทุนอื่น", "กำไร (ขาดทุน) สะสม", "ส่วนได้เสีย - ทุนอื่น", "องค์ประกอบอื่นของส่วนของผู้ถือหุ้น"]);
    expect(equity.columns?.flatMap((column) => column.accountcodes ?? [])).toEqual(["__current_earnings__"]);
    expect([10, 130, 140].map((rowno) => basis("EQ-DBD", rowno))).toEqual(["opening", "other", "closing"]);
    expect(equity.rows.find((row) => row.rowno === 100)?.accountcodes).toEqual(["__current_earnings__"]);
  });

  // ข้อ 7 ประกาศกรมพัฒนาธุรกิจการค้า เรื่อง กำหนดรายการย่อที่ต้องมีในงบการเงิน พ.ศ. 2566: แม่แบบซ่อนรายการที่ไม่มียอด (backend statements_hidezero.go);
  // แม่แบบเปล่าไม่ตั้งไว้; บังคับแสดงศูนย์เฉพาะบรรทัดผลลัพธ์/ยอดรวมใหญ่ของงบ (ข้อ 7 ไม่ได้ให้ละบรรทัดเหล่านี้)
  it("starter templates hide items with no amount by default", () => {
    const templates = generateStarterTemplates();
    expect(templates.map((template) => [template.code, template.globalstyle.hidezerorows])).toEqual([
      ["BS-DBD", true], ["PNL-DBD", true], ["COGS-STMT", true], ["CASH-FLOW-IND", true], ["EQ-DBD", true],
    ]);
    expect(Object.fromEntries(templates.map((template) => [template.code, template.rows.filter((row) => row.showzero).map((row) => row.rowno)]))).toEqual({
      "BS-DBD": [300, 690], "PNL-DBD": [140], "COGS-STMT": [140], "CASH-FLOW-IND": [180], "EQ-DBD": [140],
    });
    expect(emptyStatementTemplate().globalstyle.hidezerorows).toBeUndefined();
  });

  // backend hideZeroStatementBlock: หัวข้อแบบซ้อนจบที่แถวแรกที่ย่อหน้าไม่ลึกกว่าหัวข้อ — ถ้าแถวนั้นเป็นยอดรวมย่อยที่ตามด้วยหัวข้อย่อยที่ลึกกว่า
  // (ยอดรวมกลางส่วน) หัวข้อย่อยนั้นหลุดจากส่วน และหัวข้อแม่หายทั้งที่ยอดรวมท้ายส่วนยังพิมพ์ (เคยเกิดกับ 320 "หนี้สินและส่วนของผู้ถือหุ้น"
  // เมื่อ 540 "รวมหนี้สิน" ย่อหน้า 0); แถวบัญชีที่ต่อจากยอดรวม (เช่น ค่าแรงทางตรงต่อจากวัตถุดิบใช้ไปใน COGS-STMT) เป็นยอดสะสม ไม่ใช่ส่วนของหัวข้อ
  it("starter template headers keep every deeper row in their hide-zero section", () => {
    const indent = (row: StatementRow) => row.style?.indent ?? 0;
    const broken: string[] = [];
    for (const template of generateStarterTemplates()) {
      const rows = template.rows.filter((row) => row.rowtype !== "blank" && row.rowtype !== "divider");
      rows.forEach((header, i) => {
        const level = indent(header);
        if (header.rowtype !== "header" || !rows[i + 1] || indent(rows[i + 1]) <= level) return;
        const end = rows.findIndex((row, j) => j > i && indent(row) <= level);
        const total = rows[end], after = rows[end + 1];
        if (total && (total.rowtype === "subtotal" || total.rowtype === "formula") && indent(total) === level && after?.rowtype === "header" && indent(after) > level) {
          broken.push(`${template.code} ${header.rowno} ends at ${total.rowno}`);
        }
      });
    }
    expect(broken).toEqual([]);
    const bs = generateStarterTemplates().find((template) => template.code === "BS-DBD");
    expect(bs?.rows.find((row) => row.rowno === 540)?.style?.indent).toBe(1);
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

// หมายเหตุประกอบงบการเงิน: หัวข้อเริ่มต้นตามแบบ 2 ข้อ 5.1–5.6 (ประกาศกรมพัฒนาธุรกิจการค้า พ.ศ. 2566 หน้า 2-30..2-31)
describe("statement notes (แบบ 2 ข้อ 5)", () => {
  it("starts with the six Form 2 headings numbered 1..6 and only the 5.2.1 sentence prefilled", () => {
    const notes = starterStatementNotes();
    expect(notes.map((note) => [note.noteno, note.title])).toEqual([
      ["1", "ข้อมูลทั่วไป"], ["2", "เกณฑ์ในการจัดทำและนำเสนองบการเงิน"], ["3", "สรุปนโยบายการบัญชี"],
      ["4", "ประมาณการทางบัญชี"], ["5", "ข้อผิดพลาดในงวดก่อน"], ["6", "ข้อมูลเพิ่มเติมอื่น ๆ"],
    ]);
    expect(notes[1].body).toBe(STATEMENT_NOTES_NPAES_BASIS);
    expect(STATEMENT_NOTES_NPAES_BASIS).toBe("งบการเงินฉบับนี้จัดทำขึ้นตามมาตรฐานการรายงานทางการเงินสำหรับกิจการที่ไม่มีส่วนได้เสียสาธารณะ (TFRS for NPAEs)");
    expect(notes.filter((note) => note.body !== "")).toHaveLength(1);
    expect(new Set(notes.map((note) => note.id)).size).toBe(6);
    // สร้างใหม่ทุกครั้ง: แก้ชุดหนึ่งไม่กระทบอีกชุด
    notes[0].title = "แก้แล้ว";
    expect(starterStatementNotes()[0].title).toBe("ข้อมูลทั่วไป");
  });
  it("numbers a new note after the largest plain number and ignores 5.1-style numbers", () => {
    expect(newStatementNote([]).noteno).toBe("1");
    const next = newStatementNote([{ id: "a", noteno: " 7 ", title: "x", body: "" }, { id: "b", noteno: "5.1", title: "y", body: "" }, { id: "c", noteno: "ก", title: "z", body: "" }]);
    expect(next).toMatchObject({ noteno: "8", title: "", body: "" });
    expect(next.id).not.toBe("");
  });
  it("fills missing API fields with empty strings", () => {
    expect(statementNotesFromRecord(null)).toEqual([]);
    expect(statementNotesFromRecord({ notes: null })).toEqual([]);
    const [note] = statementNotesFromRecord({ notes: [{ id: "n1", noteno: "1", title: "ข้อมูลทั่วไป" }] });
    expect(note).toEqual({ id: "n1", noteno: "1", title: "ข้อมูลทั่วไป", body: "" });
    expect(statementNotesFromRecord({ notes: [{ noteno: "2" }] })[0].id).not.toBe("");
  });
  it("keeps text typed while a save was in flight instead of replacing it with the reloaded copy", () => {
    const sent = [{ id: "n1", noteno: " 1 ", title: "ข้อมูลทั่วไป ", body: "บริษัท" }];
    const server = [{ id: "n1", noteno: "1", title: "ข้อมูลทั่วไป", body: "บริษัท" }];
    const typedMore = [{ ...sent[0], body: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด" }];
    expect(statementNotesAfterReload(sent, JSON.stringify(sent), server)).toBe(server);
    expect(statementNotesAfterReload(typedMore, JSON.stringify(sent), server)).toBe(typedMore);
    expect(statementNotesAfterReload(null, null, server)).toBe(server);
    expect(statementNotesAfterReload(typedMore, null, null)).toBeNull();
  });
  // บันทึกชนกับฉบับที่คนอื่นบันทึกก่อน (version เก่า / สร้างปีเดียวกันก่อน / ถูกลบไปแล้ว): ส่งซ้ำไม่มีวันผ่าน จอต้องมีปุ่มโหลดฉบับล่าสุด
  it("offers loading the latest notes after a save or delete conflicts with the stored copy", () => {
    expect(["stale_version", "duplicate_code", "not_found"].map(statementNotesNeedReload)).toEqual([true, true, true]);
    expect(["validation_failed", "statement_note_no_duplicate", "unavailable", ""].map(statementNotesNeedReload)).toEqual([false, false, false, false]);
    const editor = readFileSync(resolve(process.cwd(), "src", "app", "gl", "gl-statement-notes.tsx"), "utf8");
    expect(editor).toMatch(/setReloadNeeded\(statementNotesNeedReload\(/);
    expect(editor).toMatch(/onClick=\{\(\) => void reloadLatest\(\)\}/);
    expect(editor).toContain('"gl_statement_notes_reload_latest"');
  });
  it("asks before deleting a note only when it has a title or text", () => {
    expect(statementNoteHasText({ id: "a", noteno: "1", title: "", body: "  " })).toBe(false);
    expect(statementNoteHasText({ id: "a", noteno: "1", title: " ", body: "เนื้อหา" })).toBe(true);
    expect(statementNoteHasText({ id: "a", noteno: "1", title: "ข้อมูลทั่วไป", body: "" })).toBe(true);
  });
  it("gives a repeated note id a new id so editing one note never edits another", () => {
    const notes = statementNotesFromRecord({ notes: [{ id: "n1", noteno: "1", title: "ก" }, { id: "n1", noteno: "2", title: "ข" }, { id: "n2", noteno: "3", title: "ค" }] });
    expect(notes[0].id).toBe("n1");
    expect(notes[2].id).toBe("n2");
    expect(new Set(notes.map((note) => note.id)).size).toBe(3);
  });
  it("is a GL resource so both BFF allowlists forward it", () => {
    expect(GL_RESOURCES).toContain("statement-notes");
  });
});

// ใช้แม่แบบมาตรฐาน = แทนที่ทั้งแม่แบบ (ประเภทงบ ชื่อ บรรทัด คอลัมน์ รูปแบบ; รหัสด้วยถ้ายังไม่บันทึก): ต้องถามก่อนเมื่อผู้ใช้จะเสียสิ่งที่ทำไว้
describe("standard statement template replaces the open template only after confirmation", () => {
  const starter = generateStarterTemplates()[0];
  it("asks only when something would be lost", () => {
    expect(statementStarterReplaceNeedsConfirm(null, false)).toBe(false);
    expect(statementStarterReplaceNeedsConfirm(emptyStatementTemplate(), false)).toBe(false);
    expect(statementStarterReplaceNeedsConfirm({ ...emptyStatementTemplate(), name: "งบแสดงฐานะการเงิน" }, true)).toBe(true);
    expect(statementStarterReplaceNeedsConfirm({ ...starter, id: "t1", version: 3 }, false)).toBe(true);
    expect(statementStarterReplaceNeedsConfirm({ ...emptyStatementTemplate(), statementtype: "equity", columns: [{ id: "c1", title: "ทุนที่ออกและชำระแล้ว", accountcodes: [] }] }, false)).toBe(true);
    expect(statementStarterReplaceNeedsConfirm({ ...emptyStatementTemplate(), rows: null as unknown as StatementRow[] }, false)).toBe(false);
  });
  it("keeps id/version, keeps the code only for a saved template, and reports a typed code that will be replaced", () => {
    const other = generateStarterTemplates().find((item) => item.statementtype !== starter.statementtype)!;
    const saved = { ...starter, id: "t1", version: 3, code: "BS-01", name: "งบของบริษัท" };
    const fromSaved = statementTemplateFromStarter(saved, other);
    expect(fromSaved).toMatchObject({ id: "t1", version: 3, code: "BS-01", name: other.name, statementtype: other.statementtype, rows: other.rows });
    expect(statementStarterReplacedCode(saved, fromSaved)).toBe("");
    const typed = { ...emptyStatementTemplate(), code: " BS-01 ", name: "งบของบริษัท" };
    const fromTyped = statementTemplateFromStarter(typed, other);
    expect(fromTyped.id).toBeUndefined();
    expect(fromTyped.code).toBe(other.code);
    expect(statementStarterReplacedCode(typed, fromTyped)).toBe("BS-01");
    expect(statementStarterReplacedCode({ ...emptyStatementTemplate(), code: `${other.code} ` }, fromTyped)).toBe("");
    expect(statementStarterReplacedCode(emptyStatementTemplate(), statementTemplateFromStarter(emptyStatementTemplate(), other))).toBe("");
    expect(statementStarterReplacedCode(null, statementTemplateFromStarter(null, other))).toBe("");
  });
  it("the designer awaits the confirmation before replacing, and a fresh template from the empty state is not dirty", () => {
    const screen = readFileSync(resolve(process.cwd(), "src", "app", "gl", "gl-statement-designer.tsx"), "utf8");
    expect(screen).toMatch(/async function applyStarterTemplate\(starter: GLStatementTemplate\) \{\s*const fresh = statementTemplateFromStarter\(template, starter\);\s*if \(statementStarterReplaceNeedsConfirm\(template, dirty\)\) \{\s*const replacedCode = statementStarterReplacedCode\(template, fresh\);[\s\S]*?const replaced = await confirm\(\{[\s\S]*?"gl_starter_replace_title"[\s\S]*?\{replacedCode && <p[^>]*>\{tr\("gl_starter_replace_code_changes"[\s\S]*?if \(!replaced\) return;\s*\}\s*setTemplate\(fresh\);/);
    expect(screen).toMatch(/onClick=\{\(\) => void applyStarterTemplate\(starter\)\}/);
    expect(screen).toMatch(/const fresh = emptyStatementTemplate\(\);\s*setTemplate\(fresh\);\s*setOriginal\(JSON\.stringify\(fresh\)\);\s*setStarterModalOpen\(true\);/);
  });
});

// พิมพ์ชุดงบการเงิน: ลำดับตามแบบ 2 ประกาศกรมพัฒนาธุรกิจการค้า พ.ศ. 2566 (งบฐานะการเงิน → กำไรขาดทุน → ส่วนของผู้ถือหุ้น → กระแสเงินสด → งบอื่น → หมายเหตุ)
describe("financial statement set", () => {
  const template = (code: string, statementtype: StatementType, extra: Partial<GLStatementTemplate> = {}): GLStatementTemplate => ({ ...emptyStatementTemplate(), code, name: code, statementtype, isactive: true, ...extra });
  const items = [
    template("Z-CUSTOM", "custom"),
    template("CF-01", "cash_flow"),
    template("PC-01", "production_cost"),
    template("EQ-01", "equity"),
    template("PL-02", "pnl"),
    template("PL-01", "pnl"),
    template("BS-02", "balance_sheet", { isactive: false }),
    template("BS-01", "balance_sheet"),
    template("BS-00", "balance_sheet", { isdeleted: true }),
  ];
  it("orders active, non-deleted templates by DBD Form 2 and then by code", () => {
    expect(statementSetTemplates(items).map((item) => item.code)).toEqual(["BS-01", "PL-01", "PL-02", "EQ-01", "CF-01", "PC-01", "Z-CUSTOM"]);
    expect(items.map((item) => item.code)[0]).toBe("Z-CUSTOM");
  });
  it("ranks the four DBD statements first and puts other statement types after the cash flow statement", () => {
    expect(["balance_sheet", "pnl", "equity", "cash_flow", "production_cost", "custom"].map(statementSetRank)).toEqual([0, 1, 2, 3, 4, 4]);
  });
  it("selects the first template of each DBD statement by default, never production cost or custom", () => {
    expect(statementSetDefaultSelection(statementSetTemplates(items))).toEqual(["BS-01", "PL-01", "EQ-01", "CF-01"]);
    expect(statementSetDefaultSelection(statementSetTemplates([template("PC-01", "production_cost"), template("X-1", "custom")]))).toEqual([]);
  });
  it("marks period statements exactly like the backend statementPeriodic", () => {
    expect(["pnl", "production_cost", "cash_flow", "equity", "balance_sheet", "custom"].map(statementIsPeriodic)).toEqual([true, true, true, true, false, false]);
  });
  it("is opened from the designer, knows about unsaved edits, and prints from backend reports only", () => {
    const screen = readFileSync(resolve(process.cwd(), "src", "app", "gl", "gl-statement-designer.tsx"), "utf8");
    expect(screen).toContain("<GLStatementSetDialog open={statementSetOpen} onClose={() => setStatementSetOpen(false)} unsavedChanges={dirty || notesDirty} />");
    expect(screen).toMatch(/onClick=\{\(\) => setStatementSetOpen\(true\)\}[^>]*>\s*<Printer[^>]*\/> \{tr\("gl_statement_set_print"/);
    expect(screen).not.toContain('["pnl", "production_cost", "cash_flow", "equity"]');
    const dialog = readFileSync(resolve(process.cwd(), "src", "app", "gl", "gl-statement-set.tsx"), "utf8");
    // portal อยู่นอก {open && …} จึงพิมพ์ต่อได้แม้ dialog ปิด; งบทุกรายการคำนวณพร้อมกันที่ backend และทิ้งผลที่ตอบกลับช้า
    expect(dialog).toMatch(/\)\}\s*\{print\.portal\}\s*<\/>/);
    expect(dialog).toContain('glAllRecords<GLStatementTemplate>("statement-templates"');
    expect(dialog).toContain('fetchReport("statement", query)');
    expect(dialog).toContain("Promise.allSettled(");
    expect(dialog).toContain("if (request !== requestRef.current) return;");
  });
});
