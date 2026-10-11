import { describe, expect, it, vi } from "vitest";
import type { GLAccount, GLJournalBook, GLPage } from "@/lib/general-ledger";
import { emptyReportFilters, reportQuery } from "./gl-reports";
import { DEFAULT_BUDGET_PRINT, DEFAULT_CHART_PRINT, budgetPrintConditionLines, budgetPrintFilters, budgetPrintReport, budgetPrintReversed, chartPrintConditionLines, chartPrintReversed, codeRangeLines, codeRangePrintRows, codeRangeReversed, ensureReportRight, filterChartForPrint } from "./gl-print-conditions";

const tr = (_key: string, fallback: string) => fallback;
const account = (accountcode: string, accounttype: GLAccount["accounttype"], extra: Partial<GLAccount> = {}): GLAccount =>
  ({ id: `id-${accountcode}`, version: 1, accountcode, names: [{ code: "th", name: `บัญชี ${accountcode}` }], accounttype, parentaccountcode: null, normalbalance: "debit", allowposting: true, isactive: true, iscash: false, ...extra }) as GLAccount;
const chart = [
  account("1100", "asset"), account("1200", "asset"), account("2100", "liability"), account("3100", "equity"),
  account("4100", "income"), account("5100", "expense"), account("5200", "expense", { isdeleted: true } as Partial<GLAccount>),
];
const codes = (list: GLAccount[]) => list.map((item) => item.accountcode);

describe("codeRangePrintRows (Champ code-range print conditions)", () => {
  const book = (code: string): GLJournalBook => ({ id: `b-${code}`, code, name: `สมุด ${code}`, isactive: true }) as GLJournalBook;
  const page = (items: GLJournalBook[], total = items.length): GLPage<GLJournalBook> => ({ items, total, page: 1, limit: 500, sequence: 1 });
  const url = (request: { mock: { calls: unknown[][] } }, call = 0) => request.mock.calls[call][0] as string;

  it("sends the trimmed range to the backend list of the given master", async () => {
    const request = vi.fn(async () => page([book("JV"), book("PV")]));
    const result = await codeRangePrintRows<GLJournalBook>("journal-books", " JV ", " PV ", request as never);
    expect(url(request)).toBe("journal-books?page=1&limit=500&codefrom=JV&codeto=PV&purpose=print");
    expect(result.complete).toBe(true);
    expect(result.rows.map((item) => item.code)).toEqual(["JV", "PV"]);
  });
  it("'from' alone is the exact-code filter and 'to' alone is ignored, like Champ", async () => {
    const request = vi.fn(async () => page([book("JV")]));
    await codeRangePrintRows<GLJournalBook>("journal-books", "JV", "", request as never);
    await codeRangePrintRows<GLJournalBook>("journal-books", "", "PV", request as never);
    expect(url(request, 0)).toBe("journal-books?page=1&limit=500&codefrom=JV&purpose=print");
    expect(url(request, 1)).toBe("journal-books?page=1&limit=500&purpose=print");
  });
  it("reads every page and flags a short read", async () => {
    const pages = [page([book("A"), book("B")], 3), page([book("C")], 3)];
    const all = await codeRangePrintRows<GLJournalBook>("journal-books", "", "", vi.fn(async () => pages.shift()!) as never);
    expect(all).toMatchObject({ complete: true, rows: [{ code: "A" }, { code: "B" }, { code: "C" }] });
    const short = await codeRangePrintRows<GLJournalBook>("journal-books", "", "", vi.fn(async () => page([], 5)) as never);
    expect(short).toEqual({ rows: [], complete: false });
  });
  // Champ checks the right on click (BaseReportView.cpp:733-746, 2402-2416): a register already on screen is checked
  // with one purpose-marked read; the backend's refusal reaches the dialog as the thrown message
  it("checks the print / export right of data already on screen with one purpose-marked read", async () => {
    const request = vi.fn(async () => page([]));
    await ensureReportRight("accounts", "export", request as never);
    expect(url(request)).toBe("accounts?page=1&limit=1&purpose=export");
    const refused = vi.fn(async () => { throw new Error("ไม่มีสิทธิ์พิมพ์รายงาน"); });
    await expect(ensureReportRight("accounts", "print", refused as never)).rejects.toThrow("ไม่มีสิทธิ์พิมพ์รายงาน");
  });
});

describe("code ranges", () => {
  it("is reversed only when both ends are filled and 'from' is greater", () => {
    expect(codeRangeReversed("B", "A")).toBe(true);
    expect(codeRangeReversed("A", "B")).toBe(false);
    expect(codeRangeReversed("A", "A")).toBe(false);
    expect(codeRangeReversed("B", "")).toBe(false);
    expect(codeRangeReversed("", "A")).toBe(false);
  });
  it("prints nothing without 'from', 'equals' for one code and a range for from + to", () => {
    expect(codeRangeLines("รหัสบัญชี", "", "1999", tr)).toEqual([]);
    expect(codeRangeLines("รหัสบัญชี", " 1100 ", "", tr)).toEqual(["จากรหัสบัญชี เท่ากับ 1100"]);
    expect(codeRangeLines("รหัสบัญชี", "1100", " 1999 ", tr)).toEqual(["จากรหัสบัญชี 1100 ถึง 1999"]);
  });
});

describe("chart of accounts print conditions (Champ 5520)", () => {
  it("prints every account of every category by default, never a deleted one", () => {
    expect(codes(filterChartForPrint(chart, DEFAULT_CHART_PRINT))).toEqual(["1100", "1200", "2100", "3100", "4100", "5100"]);
  });
  it("code 'from' alone keeps only that code, from + to keeps the inclusive range", () => {
    expect(codes(filterChartForPrint(chart, { ...DEFAULT_CHART_PRINT, codefrom: "1200" }))).toEqual(["1200"]);
    expect(codes(filterChartForPrint(chart, { ...DEFAULT_CHART_PRINT, codefrom: "1200", codeto: "3100" }))).toEqual(["1200", "2100", "3100"]);
    expect(codes(filterChartForPrint(chart, { ...DEFAULT_CHART_PRINT, codefrom: "", codeto: "1100" }))).toHaveLength(6);
  });
  it("keeps the category range in the Champ order asset → liability → equity → income → expense", () => {
    expect(codes(filterChartForPrint(chart, { ...DEFAULT_CHART_PRINT, categoryfrom: "liability", categoryto: "income" }))).toEqual(["2100", "3100", "4100"]);
    expect(codes(filterChartForPrint(chart, { ...DEFAULT_CHART_PRINT, categoryfrom: "expense", categoryto: "expense" }))).toEqual(["5100"]);
  });
  it("combines the code and category conditions", () => {
    expect(codes(filterChartForPrint(chart, { ...DEFAULT_CHART_PRINT, codefrom: "1100", codeto: "4100", categoryfrom: "equity", categoryto: "income" }))).toEqual(["3100", "4100"]);
  });
  it("refuses a reversed code range or category range", () => {
    expect(chartPrintReversed(DEFAULT_CHART_PRINT)).toBe(false);
    expect(chartPrintReversed({ ...DEFAULT_CHART_PRINT, codefrom: "2000", codeto: "1000" })).toBe(true);
    expect(chartPrintReversed({ ...DEFAULT_CHART_PRINT, categoryfrom: "income", categoryto: "asset" })).toBe(true);
  });
  it("lists the code range (when given), always the category range and the name display", () => {
    expect(chartPrintConditionLines(DEFAULT_CHART_PRINT, tr)).toEqual(["จากหมวด สินทรัพย์ ถึง ค่าใช้จ่าย", "แสดงชื่อบัญชี: ทั้งชื่อบัญชี 1 และ 2"]);
    expect(chartPrintConditionLines({ ...DEFAULT_CHART_PRINT, codefrom: "1100", codeto: "1999", namemode: "name2" }, tr))
      .toEqual(["จากรหัสบัญชี 1100 ถึง 1999", "จากหมวด สินทรัพย์ ถึง ค่าใช้จ่าย", "แสดงชื่อบัญชี: ชื่อบัญชี 2"]);
  });
});

describe("budget detail print conditions (Champ 5525)", () => {
  it("sends nothing for the whole report", () => {
    expect(reportQuery("budgetdetail", budgetPrintFilters(DEFAULT_BUDGET_PRINT))).toEqual({});
  });
  it("budget code: 'from' alone is exact (no budgetcodeto), from + to is a range, 'to' alone is ignored", () => {
    expect(reportQuery("budgetdetail", budgetPrintFilters({ ...DEFAULT_BUDGET_PRINT, budgetfrom: " BG-1 " }))).toEqual({ budgetcode: "BG-1" });
    expect(reportQuery("budgetdetail", budgetPrintFilters({ ...DEFAULT_BUDGET_PRINT, budgetfrom: "BG-1", budgetto: "BG-3" }))).toEqual({ budgetcode: "BG-1", budgetcodeto: "BG-3" });
    expect(reportQuery("budgetdetail", budgetPrintFilters({ ...DEFAULT_BUDGET_PRINT, budgetto: "BG-3" }))).toEqual({});
  });
  it("account: 'from' alone maps to one account, from + to is a range", () => {
    expect(reportQuery("budgetdetail", budgetPrintFilters({ ...DEFAULT_BUDGET_PRINT, accountfrom: "5100" }))).toEqual({ accountfrom: "5100", accountto: "5100" });
    expect(reportQuery("budgetdetail", budgetPrintFilters({ ...DEFAULT_BUDGET_PRINT, accountfrom: "5100", accountto: "5199" }))).toEqual({ accountfrom: "5100", accountto: "5199" });
  });
  it("sends the branch and the picked departments / projects as one comma-separated value each", () => {
    expect(reportQuery("budgetdetail", budgetPrintFilters({ ...DEFAULT_BUDGET_PRINT, branchcode: " 00001 ", departments: ["D1", "D2", "D1"], projects: ["P1"] })))
      .toEqual({ branchcode: "00001", departmentcode: "D1,D2", projectcode: "P1" });
  });
  it("other reports drop the budget code and its range (only budgetdetail takes them; budget comparison runs from the central registry)", () => {
    const filters = { ...emptyReportFilters, budgetcode: "BG-1", budgetcodeto: "BG-3" };
    expect(reportQuery("budgetdetail", filters)).toMatchObject({ budgetcode: "BG-1", budgetcodeto: "BG-3" });
    expect(reportQuery("budgetcomparison", filters)).toEqual({});
    expect(reportQuery("trialbalance", filters)).toEqual({});
  });
  it("refuses a reversed budget or account range", () => {
    expect(budgetPrintReversed(DEFAULT_BUDGET_PRINT)).toBe(false);
    expect(budgetPrintReversed({ ...DEFAULT_BUDGET_PRINT, budgetfrom: "BG-9", budgetto: "BG-1" })).toBe(true);
    expect(budgetPrintReversed({ ...DEFAULT_BUDGET_PRINT, accountfrom: "5200", accountto: "5100" })).toBe(true);
  });
  it("lists every given condition on the printed report", () => {
    expect(budgetPrintConditionLines(DEFAULT_BUDGET_PRINT, tr)).toEqual([]);
    expect(budgetPrintConditionLines({ ...DEFAULT_BUDGET_PRINT, budgetfrom: "BG-1", budgetto: "BG-3", accountfrom: "5100", accountto: "", branchcode: "00001", departments: ["D1", "D2"], projects: ["P1"] }, tr))
      .toEqual(["จากรหัสงบประมาณ BG-1 ถึง BG-3", "จากรหัสบัญชี เท่ากับ 5100", "รหัสสาขา 00001", "รหัสแผนก D1, D2", "รหัสโครงการ P1"]);
  });
  it("reads every page of the report at one snapshot and aborts when the data changes between pages", async () => {
    const report = (rows: number[], totalrows: number, sequence = 7) => ({ rows: rows.map((n) => ({ budgetcode: `B${n}` })), totalrows, sequence, totals: { amount: "9" } });
    const pages = [report([1, 2], 3), report([3], 3)];
    const request = vi.fn(async () => pages.shift()!);
    const full = await budgetPrintReport(emptyReportFilters, "changed", request as never);
    expect(full.rows.map((row) => row.budgetcode)).toEqual(["B1", "B2", "B3"]);
    expect(full.totals).toEqual({ amount: "9" });
    expect(request.mock.calls[0]).toEqual(["budgetdetail", emptyReportFilters, 1, 1000, undefined, false, "print"]);
    expect(request.mock.calls[1]).toEqual(["budgetdetail", emptyReportFilters, 2, 1000, 7, false, "print"]);
    const changed = [report([1, 2], 3), report([3], 4)];
    await expect(budgetPrintReport(emptyReportFilters, "changed", vi.fn(async () => changed.shift()!) as never)).rejects.toThrow("changed");
  });
});
