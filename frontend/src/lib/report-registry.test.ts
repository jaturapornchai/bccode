import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  PRINT_ROWS_LANDSCAPE, PRINT_ROWS_PORTRAIT, REPORT_PAGE_SIZE, REPORT_ROUTES, ReportError, buildRunParams, buildRunRequest, countTotalKeys, currentFiscalYear, defaultParamValues, drillFiscalYear,
  exportReportCsv, fetchReportRegistry, fillBlankValues, filenameFromDisposition, isReportRegistryRoute, normalizeRun, printPages, printRowsPerPage, reportErrorFrom, reportMenuIdForRoute,
  runReport, toGLReport, type ReportRun, type ReportView,
} from "./report-registry";

vi.mock("./client-auth-session", () => ({
  apiFetch: (path: string, init?: RequestInit) => fetch(path, { ...init, headers: { Authorization: "Bearer test-token", "x-bc-backend-url": "http://localhost:8888", ...init?.headers } }),
}));

const years = [
  { code: "2568", startdate: "2025-01-01", enddate: "2025-12-31" },
  { code: "2569", startdate: "2026-01-01", enddate: "2026-12-31" },
];

const ledger: ReportView = {
  menu_id: "general-ledger", champ_menu_id: "5534", code: "gl_ledger", module: "gl", family: "ledger", status: "ready", name: "รายงานบัญชีแยกประเภท", route: "/report/ledger",
  landscape: false, exports: ["csv"], drill: "journal", default_sort: "date",
  sorts: [{ value: "date", label: "วันที่" }, { value: "docno", label: "เลขที่เอกสาร" }],
  params: [
    { key: "fiscal_year", type: "year", label: "ปีบัญชี", required: true, default: "", lookup: "fiscal_year", options: [] },
    { key: "date_from", type: "date", label: "ตั้งแต่วันที่", required: false, default: "", lookup: "", options: [] },
    { key: "date_to", type: "date", label: "ถึงวันที่", required: false, default: "", lookup: "", options: [] },
    { key: "account_from", type: "code", label: "จากรหัสบัญชี", required: false, default: "", lookup: "account", options: [] },
    { key: "book_codes", type: "codes", label: "สมุดรายวัน", required: false, default: "", lookup: "book", options: [] },
    { key: "hide_opening", type: "bool", label: "ไม่แสดงยอดยกมา", required: false, default: "", lookup: "", options: [] },
  ],
};

describe("report routes", () => {
  it("maps every menu route to its registry menu id, ignoring the query string", () => {
    expect(reportMenuIdForRoute("/report/ledger")).toBe("general-ledger");
    expect(reportMenuIdForRoute("/report/ledger?tab=1")).toBe("general-ledger");
    expect(reportMenuIdForRoute("/gl/workingpaper")).toBe("working-paper");
    expect(isReportRegistryRoute("/report/pnl")).toBe(false);
    expect(reportMenuIdForRoute("/report/toString")).toBeUndefined();
  });

  it("stays in step with the backend's golden registry list (the statement designer keeps its own screen)", () => {
    const golden = JSON.parse(readFileSync(resolve(process.cwd(), "..", "backend", "internal", "report", "testdata", "registry.json"), "utf8")) as { menu_id: string; route: string; status: string }[];
    const ready = Object.fromEntries(golden.filter((entry) => entry.status === "ready").map((entry) => [entry.route, entry.menu_id]));
    expect(REPORT_ROUTES).toEqual(ready);
    expect(golden.filter((entry) => entry.status === "planned").map((entry) => entry.route)).toEqual(["/gl/statement-designer"]);
    expect(isReportRegistryRoute("/gl/statement-designer")).toBe(false);
  });
});

describe("parameter values", () => {
  it("starts from the current fiscal year and its first / last day; booleans and codes stay blank", () => {
    expect(defaultParamValues(ledger, years, "2026-10-11")).toEqual({ fiscal_year: "2569", date_from: "2026-01-01", date_to: "2026-12-31", account_from: "", book_codes: "", hide_opening: "" });
    expect(currentFiscalYear(years, "2030-05-05")?.code).toBe("2569"); // outside every year: the latest one
    expect(defaultParamValues(ledger, [], "2026-10-11").fiscal_year).toBe("");
  });

  it("keeps a registry default instead of the computed one", () => {
    const withDefault: ReportView = { ...ledger, params: ledger.params.map((param) => (param.key === "fiscal_year" ? { ...param, default: "2568" } : param)) };
    expect(defaultParamValues(withDefault, years, "2026-10-11").fiscal_year).toBe("2568");
  });

  it("fills only blank values when the years arrive late, never overwriting what the user typed", () => {
    const filled = fillBlankValues({ fiscal_year: "", date_from: "2026-03-01", account_from: "11000" }, { fiscal_year: "2569", date_from: "2026-01-01", account_from: "" });
    expect(filled).toEqual({ fiscal_year: "2569", date_from: "2026-03-01", account_from: "11000" });
  });

  it("sends offered, non-empty values as strings; a boolean always travels as true / false", () => {
    const params = buildRunParams(ledger, { fiscal_year: " 2569 ", date_from: "", account_from: "11000", book_codes: "JV,PV", hide_opening: "", not_offered: "x" });
    expect(params).toEqual({ fiscal_year: "2569", account_from: "11000", book_codes: "JV,PV", hide_opening: "false" });
    expect(buildRunParams(ledger, { hide_opening: "true" }).hide_opening).toBe("true");
  });

  it("builds a page request; the sort goes only when the report offers a choice", () => {
    expect(buildRunRequest(ledger, { fiscal_year: "2569" }, "docno", 3)).toEqual({ menu_id: "general-ledger", params: { fiscal_year: "2569", hide_opening: "false" }, sort: "docno", page: 3, limit: REPORT_PAGE_SIZE });
    expect(buildRunRequest({ ...ledger, sorts: [ledger.sorts[0]] }, {}, "date", 1)).not.toHaveProperty("sort");
    expect(buildRunRequest(ledger, {}, "", 1)).not.toHaveProperty("sort");
    expect(buildRunRequest(ledger, {}, "date", 1, { print: true, limit: 1000, rows_per_page: 28 })).toMatchObject({ print: true, limit: 1000, rows_per_page: 28 });
  });

  it("opens an account's ledger for the chosen year, else the year that holds the end date", () => {
    expect(drillFiscalYear(ledger, { fiscal_year: "2568" }, years)).toBe("2568");
    const noYear = { params: ledger.params.filter((param) => param.type !== "year") };
    expect(drillFiscalYear(noYear, { date_to: "2026-03-31" }, years)).toBe("2569");
    expect(drillFiscalYear(noYear, { date_to: "2031-03-31" }, years)).toBe("");
    expect(drillFiscalYear(noYear, {}, years)).toBe("");
  });

  it("picks the print page size from the paper orientation", () => {
    expect(printRowsPerPage({ landscape: false })).toBe(PRINT_ROWS_PORTRAIT);
    expect(printRowsPerPage({ landscape: true })).toBe(PRINT_ROWS_LANDSCAPE);
  });
});

function run(partial: Partial<ReportRun> = {}): ReportRun {
  return normalizeRun({
    columns: [
      { key: "accountcode", label: "รหัสบัญชี", type: "text", scale: 0, group: "" },
      { key: "date", label: "วันที่", type: "date", scale: 0, group: "" },
      { key: "debit", label: "เดบิต", type: "amount", scale: 2, group: "เคลื่อนไหว" },
      { key: "qty", label: "จำนวน", type: "qty", scale: 4, group: "" },
      { key: "linecount", label: "รายการ", type: "int", scale: 0, group: "" },
    ],
    totals: { debit: "100.00", linecount: "5", qty: "3" },
    ...partial,
  });
}

describe("run -> screen shape", () => {
  it("treats a total as a count only when it is an integer string of a non-amount column", () => {
    // `qty` is an amount-like column even though its total has no decimals; `linecount` is a real count
    expect(countTotalKeys(run())).toEqual(["linecount"]);
    expect(countTotalKeys(run({ totals: { header_count: "12", detail_count: "40" } }))).toEqual(["header_count", "detail_count"]);
  });

  it("adapts columns for the shared grid: amount / qty are amount columns with their scale, groups survive", () => {
    const report = toGLReport(run({ total_rows: 7, as_of: "2026-10-11", totals_labels: { linecount: "จำนวนรายการ" } }));
    expect(report.columns.map((column) => [column.key, !!column.amount, column.scale])).toEqual([
      ["accountcode", false, undefined], ["date", false, undefined], ["debit", true, 2], ["qty", true, 4], ["linecount", false, undefined],
    ]);
    expect(report.columns.find((column) => column.key === "debit")).toMatchObject({ group: "เคลื่อนไหว" });
    expect(report).toMatchObject({ totalrows: 7, asof: "2026-10-11", totalLabels: { linecount: "จำนวนรายการ" }, countTotals: ["linecount"] });
  });

  it("normalizes a sparse response so the screen can always iterate", () => {
    const empty = normalizeRun({});
    expect(empty).toMatchObject({ rows: [], columns: [], footer: [], page_totals: [], warnings: [], totals: {}, total_rows: 0 });
    expect(empty.header.conditions).toEqual([]);
    expect(normalizeRun({ header: { company_name: "บริษัท รุ่งเรือง จำกัด", branch_name: "", title: "", conditions: undefined as never, printed_by: "", printed_at: "" } }).header.conditions).toEqual([]);
  });
});

describe("print pages (server page_totals: 1-based inclusive row indexes)", () => {
  const rows = [1, 2, 3, 4, 5, 6, 7].map((n) => ({ n: String(n) }));
  const ids = (page: { rows: Record<string, string>[] }) => page.rows.map((row) => Number(row.n));

  it("splits the rows by page, chains the brought-forward totals and keeps rows no range covers", () => {
    const pages = printPages({
      rows,
      page_totals: [
        { print_page: 1, first_row: 1, last_row: 3, sums: { debit: "30.00" }, carried: { debit: "30.00" } },
        { print_page: 2, first_row: 4, last_row: 6, sums: { debit: "30.00" }, carried: { debit: "60.00" } },
      ],
    });
    expect(pages.map(ids)).toEqual([[1, 2, 3], [4, 5, 6, 7]]);
    expect(pages[0].brought).toEqual({});
    expect(pages[1].brought).toEqual({ debit: "30.00" });
    expect(pages[1].carried).toEqual({ debit: "60.00" });
  });

  it("is one page without page totals, and clamps ranges that run past the rows", () => {
    expect(printPages({ rows, page_totals: [] }).map(ids)).toEqual([[1, 2, 3, 4, 5, 6, 7]]);
    const clamped = printPages({ rows, page_totals: [{ print_page: 1, first_row: 1, last_row: 99, sums: {}, carried: {} }] });
    expect(clamped.map(ids)).toEqual([[1, 2, 3, 4, 5, 6, 7]]);
    expect(printPages({ rows: [], page_totals: [] })).toEqual([{ rows: [], sums: {}, carried: {}, brought: {} }]);
  });
});

describe("errors and file names", () => {
  it("reads the translated message, error code and offending field of a backend error", () => {
    const error = reportErrorFrom({ success: false, errorcode: "REPORT_DATE_RANGE_INVALID", code: "REPORT_DATE_RANGE_INVALID", message: "วันที่เริ่มต้นต้องไม่เกินวันที่สิ้นสุด", field: "date_to" }, 422, "ผิดพลาด");
    expect(error).toBeInstanceOf(ReportError);
    expect(error).toMatchObject({ message: "วันที่เริ่มต้นต้องไม่เกินวันที่สิ้นสุด", errorcode: "REPORT_DATE_RANGE_INVALID", field: "date_to", status: 422 });
  });

  it("falls back to the Thai text, then to the caller's fallback, and tolerates non-object bodies", () => {
    expect(reportErrorFrom({ message: "", message_th: "ไม่มีสิทธิ์" }, 403, "x").message).toBe("ไม่มีสิทธิ์");
    expect(reportErrorFrom({}, 500, "ระบบขัดข้อง").message).toBe("ระบบขัดข้อง");
    expect(reportErrorFrom(null, 502, "ระบบขัดข้อง")).toMatchObject({ message: "ระบบขัดข้อง", field: "", errorcode: "" });
  });

  it("reads the file name from Content-Disposition (RFC 5987 first)", () => {
    expect(filenameFromDisposition("attachment; filename*=UTF-8''%E0%B8%A3%E0%B8%B2%E0%B8%A2%E0%B8%87%E0%B8%B2%E0%B8%99.csv; filename=\"x.csv\"")).toBe("รายงาน.csv");
    expect(filenameFromDisposition('attachment; filename="general-ledger.csv"')).toBe("general-ledger.csv");
    expect(filenameFromDisposition(null)).toBe("");
  });
});

describe("requests", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("loads the registry of a module in the screen language and fills missing lists", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: { can_print: true, can_export: false, reports: [{ menu_id: "general-ledger", params: null, sorts: null, exports: null, drill: null }] } }));
    vi.stubGlobal("fetch", fetchMock);
    const registry = await fetchReportRegistry("th", "ผิดพลาด");
    expect(String(fetchMock.mock.calls[0][0])).toBe("/api/report/registry?module=gl&lang=th");
    expect((fetchMock.mock.calls[0][1] as RequestInit).headers).toMatchObject({ "Accept-Language": "th", Authorization: "Bearer test-token" });
    expect(registry).toMatchObject({ can_print: true, can_export: false });
    expect(registry.reports[0]).toMatchObject({ params: [], sorts: [], exports: [], drill: "", default_sort: "" });
  });

  it("raises the backend's translated message for a failed registry call, even when it arrives as HTTP 200 + success:false", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async () => Response.json({ success: false, errorcode: "REPORT_FORBIDDEN", code: "REPORT_FORBIDDEN", message: "ไม่มีสิทธิ์ดูรายงาน" })));
    await expect(fetchReportRegistry("th", "ผิดพลาด")).rejects.toMatchObject({ message: "ไม่มีสิทธิ์ดูรายงาน", errorcode: "REPORT_FORBIDDEN" });
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async () => new Response("<html>", { status: 502 })));
    await expect(fetchReportRegistry("th", "ระบบขัดข้อง")).rejects.toMatchObject({ message: "ระบบขัดข้อง", status: 502 });
  });

  it("runs a report, posting the request as JSON, and returns the normalized run", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: { menu_id: "general-ledger", rows: [{ accountcode: "11000" }], total_rows: 1 } }));
    vi.stubGlobal("fetch", fetchMock);
    const request = buildRunRequest(ledger, { fiscal_year: "2569" }, "date", 1);
    const result = await runReport(request, "en", "x");
    expect(String(fetchMock.mock.calls[0][0])).toBe("/api/report/run");
    expect(JSON.parse(String((fetchMock.mock.calls[0][1] as RequestInit).body))).toEqual(request);
    expect((fetchMock.mock.calls[0][1] as RequestInit).headers).toMatchObject({ "Accept-Language": "en" });
    expect(result).toMatchObject({ total_rows: 1, footer: [], page_totals: [] });
    expect(result.rows).toEqual([{ accountcode: "11000" }]);
  });

  it("names the parameter to highlight when a run is refused", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async () => Response.json({ success: false, errorcode: "REPORT_PARAM_REQUIRED", code: "REPORT_PARAM_REQUIRED", message: "กรุณาเลือกปีบัญชี", field: "fiscal_year" })));
    await expect(runReport(buildRunRequest(ledger, {}, "", 1), "th", "x")).rejects.toMatchObject({ field: "fiscal_year", errorcode: "REPORT_PARAM_REQUIRED", message: "กรุณาเลือกปีบัญชี" });
  });

  it("exports csv: asks for the csv format and returns the file with the server's name", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => new Response("a,b\r\n", { headers: { "Content-Type": "text/csv; charset=utf-8", "Content-Disposition": 'attachment; filename="general-ledger-2569.csv"' } }));
    vi.stubGlobal("fetch", fetchMock);
    const file = await exportReportCsv({ menu_id: "general-ledger", params: { fiscal_year: "2569" } }, "th", "x");
    expect(String(fetchMock.mock.calls[0][0])).toBe("/api/report/export");
    expect(JSON.parse(String((fetchMock.mock.calls[0][1] as RequestInit).body))).toEqual({ menu_id: "general-ledger", params: { fiscal_year: "2569" }, format: "csv" });
    expect(file.filename).toBe("general-ledger-2569.csv");
    expect(await file.blob.text()).toBe("a,b\r\n");
  });

  it("exports csv: names the file after the report when the server sends no name, and raises a JSON error instead of saving it", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async () => new Response("x", { headers: { "Content-Type": "text/csv" } })));
    expect((await exportReportCsv({ menu_id: "trial-balance", params: {} }, "th", "x")).filename).toBe("trial-balance.csv");
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async () => Response.json({ success: false, errorcode: "REPORT_EXPORT_TOO_LARGE", code: "REPORT_EXPORT_TOO_LARGE", message: "ข้อมูลมากเกินไป กรุณาจำกัดช่วงข้อมูล" })));
    await expect(exportReportCsv({ menu_id: "trial-balance", params: {} }, "th", "x")).rejects.toMatchObject({ errorcode: "REPORT_EXPORT_TOO_LARGE", message: "ข้อมูลมากเกินไป กรุณาจำกัดช่วงข้อมูล" });
  });
});
