import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { formatAppDate } from "@/lib/date-time";
import type { GLDimension } from "@/lib/general-ledger";
import { normalizeRun, type ReportParam, type ReportRun, type ReportRunColumn, type ReportView } from "@/lib/report-registry";
import { PRINT_LIMIT, applyParamChange, canDrillAccount } from "./report-center";
import { ReportCenterPrint, printedAtFromHeader, splitConditions } from "./report-center-print";
import { EMPTY_REPORT_LOOKUPS, ReportParamField, singleCodeOptions, type ReportLookups } from "./report-param-field";

const tr = (_key: string, fallback: string) => fallback;
const dim = (code: string, name: string): GLDimension => ({ id: `id-${code}`, code, name, isactive: true });
const param = (partial: Partial<ReportParam> & Pick<ReportParam, "key" | "type">): ReportParam => ({ label: partial.key, required: false, default: "", lookup: "", options: [], ...partial });

const years = [
  { code: "2568", startdate: "2025-01-01", enddate: "2025-12-31" },
  { code: "2569", startdate: "2026-01-01", enddate: "2026-12-31" },
];
const ledgerParams = [
  param({ key: "fiscal_year", type: "year", lookup: "fiscal_year" }),
  param({ key: "date_from", type: "date" }),
  param({ key: "date_to", type: "date" }),
  param({ key: "account_from", type: "code", lookup: "account" }),
];

describe("applyParamChange", () => {
  const values = { fiscal_year: "2569", date_from: "2026-03-01", date_to: "2026-03-31", account_from: "11000" };

  it("moves the date range to the chosen fiscal year, leaving the other parameters alone", () => {
    expect(applyParamChange({ params: ledgerParams }, values, "fiscal_year", "2568", years)).toEqual({ fiscal_year: "2568", date_from: "2025-01-01", date_to: "2025-12-31", account_from: "11000" });
  });

  it("only edits the one parameter for anything that is not a year, and keeps the dates for an unknown year", () => {
    expect(applyParamChange({ params: ledgerParams }, values, "account_from", "12000", years)).toEqual({ ...values, account_from: "12000" });
    expect(applyParamChange({ params: ledgerParams }, values, "fiscal_year", "2599", years)).toEqual({ ...values, fiscal_year: "2599" });
    expect(applyParamChange({ params: ledgerParams }, values, "date_to", "2026-04-30", years).date_from).toBe("2026-03-01");
  });

  it("treats a parameter with the fiscal_year lookup as the year even when its type is code", () => {
    const params = [param({ key: "budget_year", type: "code", lookup: "fiscal_year" }), param({ key: "from", type: "date" }), param({ key: "to", type: "date" })];
    expect(applyParamChange({ params }, { budget_year: "", from: "", to: "" }, "budget_year", "2569", years)).toEqual({ budget_year: "2569", from: "2026-01-01", to: "2026-12-31" });
  });
});

describe("canDrillAccount", () => {
  it("opens an account's ledger only for account-based families that have a year to open it for", () => {
    expect(canDrillAccount({ family: "trial_balance" }, "2569")).toBe(true);
    expect(canDrillAccount({ family: "working_paper" }, "2569")).toBe(true);
    expect(canDrillAccount({ family: "trial_balance" }, "")).toBe(false);
    for (const family of ["journal", "ledger", "budget", "statement", "master"]) expect(canDrillAccount({ family }, "2569"), family).toBe(false);
  });

  it("asks for the most rows the backend allows when printing", () => {
    expect(PRINT_LIMIT).toBe(1000);
  });
});

describe("singleCodeOptions", () => {
  const lookups: ReportLookups = {
    ...EMPTY_REPORT_LOOKUPS,
    departments: [dim("D01", "ฝ่ายขาย"), dim("D02", "")],
    projects: [dim("P01", "โครงการคลังสินค้า")],
    books: [dim("JV", "สมุดรายวันทั่วไป")],
    budgets: [{ code: "B68", name: "งบปี 2568", fiscalyear: "2568" }, { code: "B69", name: "งบปี 2569", fiscalyear: "2569" }],
  };

  it("offers a leading 'all' option and code · name entries (a nameless code shows the code alone)", () => {
    expect(singleCodeOptions({ type: "code", lookup: "department" }, lookups, "", tr)).toEqual([
      { value: "", label: "ทุกแผนก" }, { value: "D01", label: "D01 · ฝ่ายขาย" }, { value: "D02", label: "D02" },
    ]);
    expect(singleCodeOptions({ type: "code", lookup: "project" }, lookups, "", tr)[0]).toEqual({ value: "", label: "ทุกโครงการ" });
    expect(singleCodeOptions({ type: "code", lookup: "book" }, lookups, "", tr).map((option) => option.value)).toEqual(["", "JV"]);
  });

  it("offers only the budgets of the chosen fiscal year", () => {
    expect(singleCodeOptions({ type: "code", lookup: "budget" }, lookups, "2569", tr).map((option) => option.value)).toEqual(["", "B69"]);
    expect(singleCodeOptions({ type: "code", lookup: "budget" }, lookups, "", tr).map((option) => option.value)).toEqual(["", "B68", "B69"]);
  });

  it("offers nothing for parameters that are not a single-code register lookup", () => {
    expect(singleCodeOptions({ type: "codes", lookup: "department" }, lookups, "", tr)).toEqual([]);
    expect(singleCodeOptions({ type: "code", lookup: "account" }, lookups, "", tr)).toEqual([]);
    expect(singleCodeOptions({ type: "code", lookup: "" }, lookups, "", tr)).toEqual([]);
  });
});

describe("ReportParamField", () => {
  const render = (props: Partial<Parameters<typeof ReportParamField>[0]> & Pick<Parameters<typeof ReportParamField>[0], "param">) => renderToStaticMarkup(
    createElement(ReportParamField, { value: "", invalid: false, lookups: EMPTY_REPORT_LOOKUPS, fiscalYear: "", onChange: () => undefined, ...props }),
  );

  it("draws a date editor with the label, and marks the field by its key for the error highlight", () => {
    const html = render({ param: param({ key: "date_from", type: "date", label: "ตั้งแต่วันที่", required: true }), value: "2026-01-01" });
    expect(html).toContain("ตั้งแต่วันที่");
    expect(html).toContain('data-field="date_from"');
    expect(html).toContain('type="date"');
    expect(html).toContain('value="2026-01-01"');
    expect(html).toContain("required");
    expect(html).not.toContain("data-invalid");
  });

  it("outlines the parameter the backend named, so the problem is not shown by colour alone being a message away", () => {
    const html = render({ param: param({ key: "date_to", type: "date", label: "ถึงวันที่" }), invalid: true });
    expect(html).toContain('data-invalid="true"');
    expect(html).toContain("ring-destructive");
  });

  it("draws a boolean as a checkbox that follows the 'true' value", () => {
    const flag = param({ key: "hide_opening", type: "bool", label: "ไม่แสดงยอดยกมา" });
    expect(render({ param: flag, value: "true" })).toContain("checked");
    expect(render({ param: flag, value: "false" })).not.toContain("checked");
    expect(render({ param: flag, value: "true" })).toContain("ไม่แสดงยอดยกมา");
  });

  it("draws the fiscal-year list for a year parameter", () => {
    const lookups: ReportLookups = { ...EMPTY_REPORT_LOOKUPS, years: [{ id: "y1", code: "2569", startdate: "2026-01-01", enddate: "2026-12-31", scale: 2, retainedearningsaccount: "", profitlossaccount: "", isactive: true, closed: false }] };
    const html = render({ param: param({ key: "fiscal_year", type: "year", label: "ปีบัญชี", lookup: "fiscal_year" }), lookups, value: "2569" });
    expect(html).toContain("2569");
    expect(html).toContain('data-field="fiscal_year"');
  });

  it("draws a closed register picker showing 'all' for a single-code department lookup", () => {
    const lookups: ReportLookups = { ...EMPTY_REPORT_LOOKUPS, departments: [dim("D01", "ฝ่ายขาย")] };
    const html = render({ param: param({ key: "department_code", type: "code", label: "แผนก", lookup: "department" }), lookups });
    expect(html).toContain('role="combobox"');
    expect(html).toContain('aria-label="แผนก"');
    expect(html).toContain("ทุกแผนก");
  });

  it("shows the picked register entry on the closed picker", () => {
    const lookups: ReportLookups = { ...EMPTY_REPORT_LOOKUPS, departments: [dim("D01", "ฝ่ายขาย")] };
    expect(render({ param: param({ key: "department_code", type: "code", label: "แผนก", lookup: "department" }), lookups, value: "D01" })).toContain("D01 · ฝ่ายขาย");
  });

  it("falls back to a text box for a plain code, and a branch box says 'all branches' when empty", () => {
    const plain = render({ param: param({ key: "docno_from", type: "code", label: "จากเลขที่เอกสาร" }), value: "JV6910-0001" });
    expect(plain).toContain('value="JV6910-0001"');
    expect(plain).toContain('maxLength="30"');
    const branch = render({ param: param({ key: "branch_code", type: "code", label: "สาขา", lookup: "branch" }) });
    expect(branch).toContain("ทุกสาขา");
  });
});

describe("printedAtFromHeader", () => {
  it("prints the server's timestamp as the Thai (Buddhist-era) date and time", () => {
    expect(printedAtFromHeader("2026-10-11T10:30:00", "th")).toBe(`${formatAppDate("2026-10-11", "th")} 10:30`);
    expect(printedAtFromHeader("2026-10-11T10:30:00", "th")).toContain("2569");
  });

  it("uses the browser clock when the server gave nothing readable", () => {
    expect(printedAtFromHeader("", "th")).toMatch(/\d{1,2} \S+ 25\d{2} \d{2}:\d{2}$/);
    expect(printedAtFromHeader("not a date", "th")).toMatch(/\d{1,2} \S+ 25\d{2} \d{2}:\d{2}$/);
  });
});

describe("splitConditions", () => {
  const conditions = [
    { key: "date_from", label: "ตั้งแต่วันที่", value: "2026-01-01", type: "date" },
    { key: "date_to", label: "ถึงวันที่", value: "2026-12-31", type: "date" },
    { key: "account_from", label: "จากรหัสบัญชี", value: "11000", type: "code" },
    { key: "book_codes", label: "สมุดรายวัน", value: "  ", type: "codes" },
    { key: "as_of", label: "ณ วันที่", value: "2026-06-30", type: "text" },
  ];

  it("puts the dates on the heading's details line and the other filters in the conditions list; blank filters are not printed", () => {
    const split = splitConditions(conditions, "th");
    expect(split.description).toEqual([
      `ตั้งแต่วันที่ ${formatAppDate("2026-01-01", "th")}`, `ถึงวันที่ ${formatAppDate("2026-12-31", "th")}`, `ณ วันที่ ${formatAppDate("2026-06-30", "th")}`,
    ]);
    expect(split.conditions).toEqual(["จากรหัสบัญชี 11000"]);
  });

  it("returns nothing for a report without conditions", () => {
    expect(splitConditions([], "th")).toEqual({ description: [], conditions: [] });
  });
});

const columns: ReportRunColumn[] = [
  { key: "accountcode", label: "รหัสบัญชี", type: "text", scale: 0, group: "" },
  { key: "accountname", label: "ชื่อบัญชี", type: "text", scale: 0, group: "" },
  { key: "debit", label: "เดบิต", type: "amount", scale: 2, group: "" },
  { key: "credit", label: "เครดิต", type: "amount", scale: 2, group: "" },
];
const report: ReportView = {
  menu_id: "trial-balance", champ_menu_id: "5536", code: "gl_trial_balance", module: "gl", family: "trial_balance", status: "ready", name: "งบทดลอง", route: "/report/trialbalance",
  landscape: false, exports: ["csv"], drill: "", params: [], sorts: [], default_sort: "",
};
const detail = (code: string, name: string, debit: string, credit: string) => ({ accountcode: code, accountname: name, debit, credit });

function printRun(partial: Partial<ReportRun> = {}): ReportRun {
  return normalizeRun({
    menu_id: "trial-balance",
    columns,
    rows: [detail("11000", "เงินสด", "100.00", "0.00"), detail("11100", "เงินฝากธนาคาร", "200.00", "0.00"), detail("21000", "เจ้าหนี้การค้า", "0.00", "250.00"), detail("41000", "รายได้จากการขาย", "0.00", "50.00")],
    totals: { debit: "300.00", credit: "300.00", count: "4" },
    totals_labels: { count: "จำนวนบัญชี" },
    total_rows: 4,
    header: {
      company_name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", branch_name: "สำนักงานใหญ่", title: "รายงานงบทดลอง",
      conditions: [{ key: "date_to", label: "ถึงวันที่", value: "2026-12-31", type: "date" }, { key: "account_from", label: "จากรหัสบัญชี", value: "11000", type: "code" }],
      printed_by: "สมศรี ใจดี", printed_at: "2026-10-11T10:30:00",
    },
    ...partial,
  });
}
const renderPrint = (run: ReportRun) => renderToStaticMarkup(createElement(ReportCenterPrint, { run, report, books: [], tr, language: "th" }));
const count = (html: string, text: string) => html.split(text).length - 1;

describe("ReportCenterPrint", () => {
  const paged = printRun({
    page_totals: [
      { print_page: 1, first_row: 1, last_row: 2, sums: { debit: "300.00", credit: "0.00" }, carried: { debit: "300.00", credit: "0.00" } },
      { print_page: 2, first_row: 3, last_row: 4, sums: { debit: "0.00", credit: "300.00" }, carried: { debit: "300.00", credit: "300.00" } },
    ],
  });

  it("repeats the report heading on every printed page and lists the filters once at the end", () => {
    const html = renderPrint(paged);
    expect(count(html, "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด")).toBe(2);
    expect(count(html, "รายงานงบทดลอง")).toBe(2);
    expect(html).toContain("พิมพ์โดย : สมศรี ใจดี");
    expect(html).toContain(`วันที่พิมพ์ : ${formatAppDate("2026-10-11", "th")} 10:30`);
    expect(html).toContain(`รายละเอียด : สำนักงานใหญ่ · ถึงวันที่ ${formatAppDate("2026-12-31", "th")}`);
    expect(count(html, "เงื่อนไขการพิมพ์รายงาน")).toBe(1);
    expect(html).toContain("จากรหัสบัญชี 11000");
  });

  it("starts every table after the first on a new sheet, with the page total, brought-forward and carried-forward rows", () => {
    const html = renderPrint(paged);
    expect(count(html, "<table")).toBe(2);
    expect(count(html, "gl-print-pagebreak")).toBe(1);
    expect(html.indexOf("gl-print-pagebreak")).toBeGreaterThan(html.indexOf("</table>"));
    expect(count(html, "รวมหน้านี้")).toBe(2);
    expect(count(html, "ยอดยกไป")).toBe(1); // carried forward only at the end of a page that is not the last
    expect(count(html, "ยอดยกมา")).toBe(1); // brought forward only at the top of a page after the first
    expect(count(html, "รวมทั้งสิ้น")).toBe(1);
    const second = html.slice(html.indexOf("gl-print-pagebreak"), html.lastIndexOf("</table>"));
    expect(second).toContain("ยอดยกมา");
    expect(second).toContain("21000");
    expect(second).not.toContain("11000");
  });

  it("closes the last page with the grand total and the count totals with their labels (counts without decimals)", () => {
    const html = renderPrint(paged);
    expect(html).toContain("300.00");
    expect(html).toContain('class="gl-print-totals"');
    expect(html).toContain("<dt>จำนวนบัญชี</dt><dd>4</dd>");
  });

  it("prints one table without page rows when the report has no page totals", () => {
    const html = renderPrint(printRun());
    expect(count(html, "<table")).toBe(1);
    expect(html).not.toContain("gl-print-pagebreak");
    expect(html).not.toContain("รวมหน้านี้");
    expect(html).toContain("รวมทั้งสิ้น");
    for (const code of ["11000", "11100", "21000", "41000"]) expect(html).toContain(code);
  });

  it("prints a caption row across the leading columns and bolds group totals", () => {
    const html = renderPrint(printRun({
      rows: [detail("11000", "เงินสด", "100.00", "0.00"), { accountcode: "", accountname: "", debit: "100.00", credit: "0.00", _kind: "group_total", _label: "รวมสินทรัพย์" }],
      total_rows: 2,
    }));
    expect(html).toMatch(/<td colSpan="2">รวมสินทรัพย์<\/td>/);
    expect(html).toMatch(/<tr class="gl-print-total"><td colSpan="2">รวมสินทรัพย์/);
  });

  it("prints the footer rows with their label in the first column, spanning the empty ones", () => {
    const html = renderPrint(printRun({ footer: [{ label: "ยอดสุทธิหลังปรับปรุง", debit: "300.00" }, { label: "", debit: "9.99" }] }));
    expect(html).toMatch(/<td colSpan="2">ยอดสุทธิหลังปรับปรุง<\/td>/);
    expect(html).not.toContain("9.99");
  });

  it("says so when the run found no rows, instead of printing an empty frame", () => {
    const html = renderPrint(printRun({ rows: [], totals: {}, total_rows: 0 }));
    expect(html).toContain("ไม่พบรายการตามเงื่อนไขที่เลือก");
  });
});

describe("print / export rights (Champ สิทธิ์เฉพาะ)", () => {
  // a role without the right still sees the buttons and is told how to get it (Champ BaseReportView.cpp:736-745, 2406-2415),
  // instead of the buttons vanishing for roles that printed before report:print / report:export existed
  it("keeps the buttons and refuses with the message naming the right to tick", () => {
    const source = readFileSync(resolve(process.cwd(), "src/app/report/report-center.tsx"), "utf8");
    expect(source).not.toMatch(/\{can(Print|Export) && <Button/);
    expect(source).toContain(`refused(registry?.can_print, "report_error_print_forbidden"`);
    expect(source).toContain(`refused(registry?.can_export, "report_error_export_forbidden"`);
    const tsv = readFileSync(resolve(process.cwd(), "..", "backend", "assets", "language", "languages.tsv"), "utf8");
    const th = (key: string) => tsv.split(/\r?\n/).find((line) => line.startsWith(`${key}\t`))?.split("\t")[1] ?? "";
    expect(th("report_error_print_forbidden")).toContain(th("perm_special_report_print"));
    expect(th("report_error_export_forbidden")).toContain(th("perm_special_report_export"));
  });
});
