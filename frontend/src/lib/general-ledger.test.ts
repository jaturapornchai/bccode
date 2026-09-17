import { describe, expect, it } from "vitest";
import { amountString, amountUnits, csvCell, emptyAccount, emptyFiscalYear, emptyJournal, emptyLine, formatAmount, GL_MENU_ITEMS, isGeneralLedgerRoute, journalTotals, reportCsv, validateJournal, evaluateStatementFormula, generateStarterTemplates } from "./general-ledger";

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
  it("covers the exact 38 existing GL menu routes", () => {
    expect(GL_MENU_ITEMS).toHaveLength(38);
    expect(new Set(GL_MENU_ITEMS.map((item) => item.route)).size).toBe(38);
    expect(GL_MENU_ITEMS.every((item) => isGeneralLedgerRoute(item.route))).toBe(true);
    expect(isGeneralLedgerRoute("/report/ledger?accountcode=A")).toBe(true);
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
