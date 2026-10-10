import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { afterEach, describe, expect, it, vi } from "vitest";
import { accountTypes, reportCsv, type GLJournalBook, type GLReport } from "@/lib/general-ledger";
import { GET } from "../api/gl/[...glPath]/route";
import { ReportGrid, controlRowText, emptyReportFilters, isSummaryRow, labelledReportRows, reportQuery } from "./gl-reports";

function render(patch: Partial<GLReport>) {
  const report: GLReport = { sequence: 1, columns: [], rows: [], totals: {}, totalrows: 0, warnings: [], asof: "2026-09-11", ...patch };
  return renderToStaticMarkup(createElement(ReportGrid, { report }));
}

describe("Thai general ledger report presentation", () => {
  it.each([
    ["accounttype", accountTypes],
    ["status", { draft: "ฉบับร่าง", posted: "ผ่านรายการแล้ว", reversed: "กลับรายการแล้ว", void: "ยกเลิกร่าง" }],
    ["direction", { in: "เงินเข้า", out: "เงินออก" }],
    ["category", { operating: "ดำเนินงาน", investing: "ลงทุน", financing: "จัดหาเงิน", unclassified: "ยังไม่ระบุ" }],
  ] as const)("renders %s as Thai labels", (key, labels) => {
    const html = render({ columns: [{ key, label: "รายการ", amount: false }], rows: Object.keys(labels).map(value => ({ [key]: value })) });
    for (const [value, label] of Object.entries(labels)) {
      expect(html).toContain(`>${label}</td>`);
      expect(html).not.toContain(`>${value}</td>`);
    }
  });

  it("renders book codes with the company's journal-book names (unknown codes stay as codes)", () => {
    const books: GLJournalBook[] = [
      { id: "b1", code: "JV", name: "สมุดรายวันทั่วไป", booktype: 1, isactive: true },
      { id: "b2", code: "POS1", name: "สมุดขายหน้าร้าน", booktype: 4, isactive: false },
    ];
    const report: GLReport = { sequence: 1, columns: [{ key: "bookcode", label: "สมุด", amount: false }], rows: [{ bookcode: "JV" }, { bookcode: "POS1" }, { bookcode: "ZZ" }], totals: {}, totalrows: 3, warnings: [], asof: "2026-09-11" };
    const html = renderToStaticMarkup(createElement(ReportGrid, { report, books }));
    expect(html).toContain(">สมุดรายวันทั่วไป</td>");
    expect(html).toContain(">สมุดขายหน้าร้าน</td>");
    expect(html).toContain(">ZZ</td>");
  });

  it("labels supplemental financial totals and shows a line count without money decimals", () => {
    const html = render({ totals: { currentearnings: "52000.00000000", cash: "341395.00000000", unclassifiedlines: "0" } });
    expect(html).toContain("กำไรขาดทุนที่ยังไม่ปิด");
    expect(html).toContain("เงินสดและเงินฝากธนาคาร");
    expect(html).toContain("บรรทัดที่ยังไม่ระบุประเภท");
    expect(html).toContain("52,000.00");
    expect(html).toContain("341,395.00");
    expect(html).toContain("0 รายการ");
    expect(html).not.toContain("ยอดรวม");
  });

  it("hides only the computed account key while preserving source data and exact CSV amounts", () => {
    const report: GLReport = { sequence: 1, columns: [{ key: "accountcode", label: "รหัสบัญชี", amount: false }, { key: "accountname", label: "ชื่อบัญชี", amount: false }, { key: "accounttype", label: "หมวดบัญชี", amount: false }, { key: "amount", label: "จำนวนเงิน", amount: true }], rows: [{ accountcode: "__current_earnings__", accountname: "กำไรขาดทุนที่ยังไม่ปิดเข้ากำไรสะสม", accounttype: "equity", amount: "0.30000001" }, { accountcode: "BM69-310201", accountname: "กำไรสะสม", accounttype: "equity", amount: "100000.00000000" }], totals: {}, totalrows: 2, warnings: [], asof: "2026-09-11" };
    const before = JSON.stringify(report), csv = reportCsv(report);
    const html = render(report);
    expect(html).toContain(">—</td>");
    expect(html).not.toContain("__current_earnings__");
    expect(html).toContain("BM69-310201");
    expect(html).toContain("กำไรขาดทุนที่ยังไม่ปิดเข้ากำไรสะสม");
    expect(html).toContain("0.30000001");
    expect(JSON.stringify(report)).toBe(before);
    expect(reportCsv(report)).toBe(csv);
    expect(csv).toContain("__current_earnings__");
    expect(csv).toContain("0.30000001");
  });

  it("does not translate business codes, descriptions or unknown enum values", () => {
    const html = render({ columns: [{ key: "accountcode", label: "รหัส", amount: false }, { key: "description", label: "รายละเอียด", amount: false }, { key: "status", label: "สถานะ", amount: false }], rows: [{ accountcode: "asset", description: "in", status: "future-status" }] });
    expect(html).toContain(">asset</td>");
    expect(html).toContain(">in</td>");
    expect(html).toContain(">future-status</td>");
  });

  it("renders interactive drill-down buttons when callbacks are provided", () => {
    const report: GLReport = {
      sequence: 1,
      columns: [
        { key: "accountcode", label: "รหัสบัญชี", amount: false },
        { key: "docno", label: "เลขที่เอกสาร", amount: false },
      ],
      rows: [
        { accountcode: "1111-01", docno: "JV202609001", journalid: "journal-1" },
      ],
      totals: {},
      totalrows: 1,
      warnings: [],
      asof: "2026-09-18",
    };

    const html = renderToStaticMarkup(createElement(ReportGrid, {
      report,
      onDrillAccount: () => {},
      onDrillDocNo: () => {},
    }));

    expect(html).toContain("JV202609001");
    expect(html).toContain("1111-01");
    expect(html).toContain("ดูรายงานแยกประเภท");
    expect(html).toContain("รายละเอียดใบสำคัญรายวัน");
  });

  it("renders budget comparison and gl journal totals properly", () => {
    const html = render({
      totals: {
        budgetamount: "500000.00",
        actualamount: "320000.00",
        variance: "180000.00",
      },
    });
    expect(html).toContain("งบประมาณ");
    expect(html).toContain("ยอดจริง");
    expect(html).toContain("ผลต่าง");
    expect(html).toContain("500,000.00");
    expect(html).toContain("320,000.00");
    expect(html).toContain("180,000.00");
  });
});


describe("report source identifiers", () => {
  it("keeps hidden journal IDs out of CSV and never makes a missing ID clickable", () => {
    const report: GLReport = { sequence: 1, columns: [{key: "docno", label: "เลขที่", amount: false}], rows: [{docno: "JV1", journalid: "hidden-id"}, {docno: "JV2"}], totals: {}, totalrows: 2, warnings: [], asof: "2026-09-20" };
    const html = renderToStaticMarkup(createElement(ReportGrid, {report, onDrillDocNo: () => {}}));
    expect((html.match(/title="รายละเอียดใบสำคัญรายวัน"/g) ?? []).length).toBe(1);
    expect(reportCsv(report)).not.toContain("hidden-id");
    expect(reportCsv(report)).toContain("JV2");
  });
});

// สีต้องตามพาเลตที่ผู้ใช้เลือก (10 พาเลต + dark) — สถานะใช้ token primary/destructive/muted ตาม ui-scale-polish 8.27
describe("GL report and statement-note colours follow the theme tokens", () => {
  it.each(["gl-reports.tsx", "gl-statement-notes.tsx"])("%s has no hard-coded palette colour classes", (file) => {
    const source = readFileSync(resolve(process.cwd(), "src", "app", "gl", file), "utf8");
    expect(source).not.toMatch(/\b(amber|emerald|purple|red|green)-\d/);
  });
});

// Champ report options (docs/kms/champ/09-gl.md §5.2-5.6): each report sends only the options it takes
describe("Champ report filters and options", () => {
  afterEach(() => vi.unstubAllGlobals());
  const all = { ...emptyReportFilters, fiscalyear: "2569", from: "2569-01-01", to: "2569-12-31", accountfrom: "11000", accountto: "11999", budgetcode: "BG-1", hideopening: "true", hidezero: "true", controltotals: "true", summary: "true", gross: "true", daytotals: "true" };
  it.each([
    ["trialbalance", { accountfrom: "11000", accountto: "11999", hidezero: "true", controltotals: "true", gross: "true" }, ["budgetcode", "hideopening", "summary", "daytotals"]],
    ["workingpaper", { hidezero: "true", controltotals: "true" }, ["budgetcode", "hideopening", "summary", "gross", "daytotals"]],
    ["gljournal", { summary: "true", accountto: "11999" }, ["budgetcode", "hideopening", "hidezero", "controltotals", "gross", "daytotals"]],
    ["budgetcomparison", { budgetcode: "BG-1", accountfrom: "11000" }, ["hideopening", "hidezero", "controltotals", "summary", "gross", "daytotals"]],
    ["pnl", { accountfrom: "11000" }, ["budgetcode", "hideopening", "hidezero", "controltotals", "summary", "gross", "daytotals"]],
  ] as const)("%s keeps its own options and drops the rest", (name, kept, dropped) => {
    const query = reportQuery(name, all);
    expect(query).toMatchObject(kept);
    for (const key of dropped) expect(query).not.toHaveProperty(key);
    expect(reportQuery(name, emptyReportFilters)).toEqual({});
  });
  it("sends the picked departments and projects as one comma-separated value each (Champ multi-select), nothing when none is picked", () => {
    const picked = { ...emptyReportFilters, fiscalyear: "2569", departmentcode: "D1,D2", projectcode: "P1" };
    for (const name of ["ledger", "trialbalance", "gljournal", "budgetcomparison", "budgetdetail", "statement"]) {
      expect(reportQuery(name, picked)).toMatchObject({ departmentcode: "D1,D2", projectcode: "P1" });
      expect(reportQuery(name, { ...picked, departmentcode: "", projectcode: "" })).not.toHaveProperty("departmentcode");
    }
  });
  it("ledger: hide-opening for the detail report, control totals and as-of date for the summary (5535)", () => {
    const detail = reportQuery("ledger", { ...all, summary: "" });
    expect(detail).toMatchObject({ hideopening: "true", daytotals: "true", from: "2569-01-01" });
    expect(detail).not.toHaveProperty("controltotals");
    const summary = reportQuery("ledger", all);
    expect(summary).toMatchObject({ summary: "true", controltotals: "true", to: "2569-12-31" });
    expect(summary).not.toHaveProperty("hideopening");
    expect(summary).not.toHaveProperty("daytotals");
    expect(summary).not.toHaveProperty("from");
  });
  it("ledger day totals print as bold summary rows labelled รวมรายวัน with the day sums and no account code, date or balance", () => {
    const columns = [{ key: "accountcode", label: "ผังบัญชี", amount: false }, { key: "date", label: "วันที่", amount: false }, { key: "description", label: "คำอธิบาย", amount: false }, { key: "debit", label: "เดบิต", amount: true }, { key: "credit", label: "เครดิต", amount: true }, { key: "balance", label: "ยอดคงเหลือ", amount: true }];
    const rows = [
      { accountcode: "11110", date: "2569-01-05", description: "ขายสด", debit: "50", credit: "0", balance: "50", rowtype: "line", accountfirst: "1" },
      { accountcode: "11110", date: "2569-01-05", description: "ขายสด", debit: "25", credit: "0", balance: "75", rowtype: "line", accountfirst: "" },
      { accountcode: "11110", date: "", description: "", debit: "75", credit: "0", balance: "", rowtype: "daytotal", accountfirst: "" },
    ];
    const html = renderToStaticMarkup(createElement(ReportGrid, { report: { sequence: 1, columns, rows, totals: { debit: "75", credit: "0" }, totalrows: 3, warnings: [], asof: "2569-12-31" } as GLReport, onDrillAccount: () => {} }));
    expect(html).toContain("รวมรายวัน");
    expect(html.match(/11110/g)?.length).toBe(1); // รหัสบัญชีพิมพ์ที่แถวแรกของบัญชีเท่านั้น
    expect(isSummaryRow({ rowtype: "daytotal" })).toBe(true);
    expect(controlRowText({ rowtype: "daytotal" }, "description", "", (_key: string, fallback: string) => fallback)).toBe("รวมรายวัน");
  });
  it("shows the journal report as Champ document blocks: header, lines without repeating the header fields, then the document total", () => {
    const report: GLReport = { sequence: 1, columns: [{ key: "date", label: "วันที่", amount: false }, { key: "docno", label: "เลขที่เอกสาร", amount: false }, { key: "accountcode", label: "ผังบัญชี", amount: false }, { key: "description", label: "คำอธิบาย", amount: false }, { key: "debit", label: "เดบิต", amount: true }, { key: "credit", label: "เครดิต", amount: true }], rows: [
      { date: "2569-01-05", docno: "JV6901-0001", accountcode: "", description: "ขายวัสดุก่อสร้าง รับชำระเงินสดหน้าร้าน", debit: "", credit: "", rowtype: "document", journalid: "j1" },
      { date: "2569-01-05", docno: "JV6901-0001", accountcode: "11110", description: "", debit: "107", credit: "0", rowtype: "docline", journalid: "j1" },
      { date: "2569-01-05", docno: "JV6901-0001", accountcode: "41100", description: "", debit: "0", credit: "107", rowtype: "docline", journalid: "j1" },
      { date: "2569-01-05", docno: "JV6901-0001", accountcode: "", description: "", debit: "107", credit: "107", rowtype: "doctotal", linecount: "2", journalid: "j1" },
    ], totals: { debit: "107", credit: "107" }, totalrows: 4, warnings: [], asof: "2569-12-31" };
    const html = renderToStaticMarkup(createElement(ReportGrid, { report, onDrillAccount: () => {} }));
    expect(html.match(/JV6901-0001/g)?.length).toBe(1); // เลขที่เอกสารพิมพ์ที่หัวเอกสารเท่านั้น
    expect(html).toContain("ขายวัสดุก่อสร้าง รับชำระเงินสดหน้าร้าน");
    expect(html).toContain("รวมทั้งสิ้น 2 รายการ");
  });

  it("prints Champ control-account header rows before the children and name-only bold subtotals after, neither drilling to the ledger", () => {
    const report: GLReport = { sequence: 1, columns: [{ key: "accountcode", label: "รหัสบัญชี", amount: false }, { key: "accountname", label: "ชื่อบัญชี", amount: false }, { key: "debit", label: "เดบิต", amount: true }], rows: [{ accountcode: "11100", accountname: "เงินสดและรายการเทียบเท่าเงินสด", debit: "", rowtype: "header", accountlevel: "3" }, { accountcode: "11110", accountname: "เงินสดในมือ", debit: "100", rowtype: "account", accountlevel: "4" }, { accountcode: "11100", accountname: "เงินสดและรายการเทียบเท่าเงินสด", debit: "100", rowtype: "control", accountlevel: "3" }], totals: { debit: "100" }, totalrows: 3, warnings: [], asof: "2569-12-31" };
    const html = renderToStaticMarkup(createElement(ReportGrid, { report, onDrillAccount: () => {} }));
    expect(html.match(/11100/g)?.length).toBe(1); // รหัสพิมพ์ที่หัวบัญชีคุมเท่านั้น
    expect(html.match(/เงินสดและรายการเทียบเท่าเงินสด/g)?.length).toBe(2);
    expect(html).not.toContain("รวม เงินสด");
    expect(html).toContain("font-semibold");
    expect((html.match(/title="ดูรายงานแยกประเภท"/g) ?? []).length).toBe(1);
    expect(html).toContain("padding-inline-start:calc(0.5rem + 3em)");
  });
  it("the BFF forwards the account range and option flags to the backend", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    const headers = { Authorization: "Bearer test-token", "x-bc-backend-url": "http://localhost:8888" };
    await GET(new Request("http://localhost/api/gl/reports/trialbalance?fiscalyear=2569&accountfrom=11000&accountto=11999&hideopening=true&hidezero=true&controltotals=true&summary=true&gross=true&budgetcode=BG-1", { headers }), { params: Promise.resolve({ glPath: ["reports", "trialbalance"] }) });
    const forwarded = new URL(String(fetchMock.mock.calls[0][0])).searchParams;
    for (const key of ["accountfrom", "accountto", "hideopening", "hidezero", "controltotals", "summary", "gross", "budgetcode"]) expect(forwarded.has(key)).toBe(true);
  });
});

describe("labelledReportRows — CSV carries the screen labels", () => {
  it("labels ledger summary rows and document totals, leaves posting lines untouched", () => {
    const tr = (_key: string, fallback: string) => fallback;
    const rows = labelledReportRows([
      { rowtype: "daytotal", description: "", debit: "100.00" },
      { rowtype: "doctotal", description: "", linecount: "2" },
      { rowtype: "line", description: "ขายเงินสด", debit: "100.00" },
    ], tr);
    expect(rows[0]).toMatchObject({ description: "รวมรายวัน", debit: "100.00" });
    expect(rows[1].description).toBe("รวมทั้งสิ้น 2 รายการ");
    expect(rows[2].description).toBe("ขายเงินสด");
  });
});

describe("report range chosen by period", () => {
  it("sends the dates resolved from the periods and never the picked period codes", () => {
    const query = reportQuery("trialbalance", { ...emptyReportFilters, fiscalyear: "2026", from: "2026-02-01", to: "2026-03-31", periodfrom: "2026-02", periodto: "2026-03" });
    expect(query).toEqual({ fiscalyear: "2026", from: "2026-02-01", to: "2026-03-31" });
  });
});

// Champ 5532/5533/5506/5499/5535 report details: document-no range + sort, several books, per-department/project split
describe("Champ report details: documents, books and dimension split", () => {
  afterEach(() => vi.unstubAllGlobals());
  const base = { ...emptyReportFilters, fiscalyear: "2569" };
  const tr = (_key: string, fallback: string) => fallback;
  it("sends several journal books as one comma-separated value on every report that filters by book", () => {
    for (const name of ["gljournal", "ledger", "trialbalance"]) expect(reportQuery(name, { ...base, bookcode: "JV,PV" })).toMatchObject({ bookcode: "JV,PV" });
  });
  it("gljournal keeps the document range and sort order; day totals only for the detail report sorted by date", () => {
    const picked = { ...base, docnofrom: "JV6901-0001", docnoto: "JV6901-0009", orderby: "date", daytotals: "true" };
    expect(reportQuery("gljournal", picked)).toMatchObject({ docnofrom: "JV6901-0001", docnoto: "JV6901-0009", orderby: "date", daytotals: "true" });
    expect(reportQuery("gljournal", { ...picked, orderby: "docno" })).not.toHaveProperty("daytotals");
    expect(reportQuery("gljournal", { ...picked, summary: "true" })).not.toHaveProperty("daytotals");
    expect(reportQuery("gljournal", { ...picked, summary: "true" })).toMatchObject({ summary: "true", docnofrom: "JV6901-0001", orderby: "date" });
    for (const name of ["ledger", "trialbalance", "workingpaper", "pnl"]) {
      const query = reportQuery(name, picked);
      for (const key of ["docnofrom", "docnoto", "orderby"]) expect(query).not.toHaveProperty(key);
    }
  });
  it("trial balance, working paper and the ledger summary send the department/project split; the total row only with the department split", () => {
    const split = { ...base, bydepartment: "true", byproject: "true", dimtotals: "true" };
    for (const name of ["trialbalance", "workingpaper"]) expect(reportQuery(name, split)).toMatchObject({ bydepartment: "true", byproject: "true", dimtotals: "true" });
    expect(reportQuery("ledger", { ...split, summary: "true" })).toMatchObject({ bydepartment: "true", byproject: "true", dimtotals: "true" });
    expect(reportQuery("trialbalance", { ...split, bydepartment: "" })).not.toHaveProperty("dimtotals");
    expect(reportQuery("trialbalance", { ...split, bydepartment: "", dimtotals: "" })).toMatchObject({ byproject: "true" });
    for (const key of ["bydepartment", "byproject", "dimtotals"]) {
      expect(reportQuery("ledger", split)).not.toHaveProperty(key); // detail ledger has no split
      expect(reportQuery("gljournal", split)).not.toHaveProperty(key);
    }
  });
  it("day totals of the journal report say how many lines or documents the day holds; the ledger keeps its plain label", () => {
    expect(controlRowText({ rowtype: "daytotal", linecount: "3" }, "description", "", tr)).toBe("รวมรายวัน 3 รายการ");
    expect(controlRowText({ rowtype: "daytotal" }, "description", "", tr)).toBe("รวมรายวัน");
    expect(controlRowText({ rowtype: "daytotal", linecount: "3" }, "debit", "50", tr)).toBe("50");
  });
  it("prints the tax invoice date and number on the document header row only", () => {
    const columns = [{ key: "docno", label: "เลขที่เอกสาร", amount: false }, { key: "taxinvoicedate", label: "วันที่ใบกำกับภาษี", amount: false }, { key: "taxinvoiceno", label: "เลขที่ใบกำกับภาษี", amount: false }, { key: "description", label: "คำอธิบาย", amount: false }, { key: "debit", label: "เดบิต", amount: true }];
    const rows = [
      { docno: "PV6901-0007", taxinvoicedate: "2569-01-05", taxinvoiceno: "IV-0042", description: "ซื้อวัสดุสำนักงาน", debit: "", rowtype: "document", journalid: "j1" },
      { docno: "PV6901-0007", taxinvoicedate: "", taxinvoiceno: "", description: "ค่าวัสดุสำนักงาน", debit: "100", rowtype: "docline", journalid: "j1" },
    ];
    const html = renderToStaticMarkup(createElement(ReportGrid, { report: { sequence: 1, columns, rows, totals: {}, totalrows: 2, warnings: [], asof: "2569-12-31" } as GLReport }));
    expect(html).toContain("วันที่ใบกำกับภาษี");
    expect(html.match(/IV-0042/g)?.length).toBe(1);
    expect(html).toContain("ค่าวัสดุสำนักงาน");
  });
  it("the BFF forwards the new report query keys to the backend", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    const headers = { Authorization: "Bearer test-token", "x-bc-backend-url": "http://localhost:8888" };
    await GET(new Request("http://localhost/api/gl/reports/gljournal?fiscalyear=2569&bookcode=JV,PV&docnofrom=A1&docnoto=A9&orderby=date&daytotals=true&bydepartment=true&byproject=true&dimtotals=true", { headers }), { params: Promise.resolve({ glPath: ["reports", "gljournal"] }) });
    const forwarded = new URL(String(fetchMock.mock.calls[0][0])).searchParams;
    for (const key of ["bookcode", "docnofrom", "docnoto", "orderby", "daytotals", "bydepartment", "byproject", "dimtotals"]) expect(forwarded.has(key)).toBe(true);
    expect(forwarded.get("bookcode")).toBe("JV,PV");
  });
  it("provides Desktop ERP keyboard shortcuts and hints for Print (Alt+P), Export (Alt+E), and Submit (Enter / Alt+R)", () => {
    const source = readFileSync(resolve(process.cwd(), "src/app/gl/gl-reports.tsx"), "utf8");
    expect(source).toContain("Alt+P");
    expect(source).toContain("Alt+E");
    expect(source).toContain("Alt+R");
    expect(source).toContain("handleKeyDown");
    expect(source).toContain("<kbd className=\"hidden sm:inline-block ml-1 text-[10px] font-sans font-normal opacity-70 bg-black/15 dark:bg-white/15 px-1 rounded\">Enter</kbd>");
    expect(source).toContain("<kbd className=\"hidden sm:inline-block ml-1 text-[10px] font-sans font-normal opacity-70 bg-muted px-1 rounded\">Alt+E</kbd>");
    expect(source).toContain("<kbd className=\"hidden sm:inline-block ml-1 text-[10px] font-sans font-normal opacity-70 bg-muted px-1 rounded\">Alt+P</kbd>");
  });
});
