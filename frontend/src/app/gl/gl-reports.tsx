"use client";

import { useEffect, useMemo, useState } from "react";
import { Download, RefreshCw, FileText, ExternalLink, X, CheckCircle2, AlertTriangle, Printer } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Combobox } from "@/components/ui/select";
import { accountTypeLabels, displayAmountUnits, findJournalBook, journalBookName, type GLDimension, type GLJournalBook, labelText, type GLLabel, type GLTextFn, formatAmount, reportCsv, type GLReport, type GLJournal, journalTotals, amountString, GL_MENU_ITEMS, accountName, type GLAccount, normalizeDimensionCodes, reportCodeList, type GLMaster, periodRangeDates } from "@/lib/general-ledger";
import type { LanguageCode } from "@/lib/i18n";
import { formatAppDate } from "@/lib/date-time";
import { glAllRecords, glRequest } from "@/lib/general-ledger-api";
import { DimensionCodeMultiSelect, useDimensionRegisters } from "./gl-dimension-select";
import { AccountSelect, Check, Field, Notice, Pager, YearSelect, actionClass, control, downloadText, panel, useReferences, useRowDensity, useGLLanguage, useGLText } from "./gl-common";
import { yearPeriods } from "./gl-periods";
import { AccountLedgerModal } from "./account-ledger-drilldown";
import { useReportPreferences } from "@/hooks/use-report-preferences";
import { ReportDisplayToolbar } from "@/components/report-display-toolbar";
import { GLVoucherPrint, ReportPrintConditions, ReportPrintHeading, printCompanyName, printedAtText, printOrientation, printUserName, useGLPrint, visiblePrintColumns, printAmountCell } from "./gl-print";
import { KeyHint, ReportFilterSummary, ReportFilterToggle, useReportShortcuts } from "./report-filter-bar";

// accountfrom/accountto = Champ "จากรหัสบัญชี–ถึงรหัสบัญชี"; ตัวเลือกเป็น "true" หรือ "" (ส่งเป็น query ตรง ๆ)
export type ReportFilters = { fiscalyear: string; from: string; to: string; periodfrom: string; periodto: string; accountcode: string; accountfrom: string; accountto: string; branchcode: string; departmentcode: string; projectcode: string; bookcode: string; budgetcode: string; budgetcodeto: string; hideopening: string; hidezero: string; controltotals: string; summary: string; gross: string; daytotals: string; bydepartment: string; byproject: string; dimtotals: string };
export const emptyReportFilters: ReportFilters = { fiscalyear: "", from: "", to: "", periodfrom: "", periodto: "", accountcode: "", accountfrom: "", accountto: "", branchcode: "", departmentcode: "", projectcode: "", bookcode: "", budgetcode: "", budgetcodeto: "", hideopening: "", hidezero: "", controltotals: "", summary: "", gross: "", daytotals: "", bydepartment: "", byproject: "", dimtotals: "" };
// ตัวเลือกแบบ Champ ที่แต่ละรายงานรับ (docs/kms/champ/09-gl.md §5.2-5.6) — รายงานอื่นไม่ส่ง เพื่อให้ URL ตรงกับสิ่งที่จอแสดง; fetchReport ยังใช้กับเจาะบัญชี งวด/ปิดบัญชี และพิมพ์จากจอข้อมูลหลัก
// แยกแผนก / แยกโครงการ / มียอดรวม (Champ 5506/5499/5535): หนึ่งแถวต่อบัญชี × รหัสมิติ — แยกประเภทใช้เฉพาะแบบสรุป
const BREAKDOWN_KEYS: readonly (keyof ReportFilters)[] = ["bydepartment", "byproject", "dimtotals"];
const REPORT_OPTIONS: Readonly<Record<string, readonly (keyof ReportFilters)[]>> = {
  ledger: ["summary", "hideopening", "controltotals", "daytotals", ...BREAKDOWN_KEYS],
  trialbalance: ["hidezero", "controltotals", "gross", ...BREAKDOWN_KEYS], // gross = Champ 5507 (trial balance 2): kept for the legacy endpoint; the menu now opens trial-balance-2 from the central registry
  workingpaper: ["hidezero", "controltotals", ...BREAKDOWN_KEYS],
  budgetdetail: ["budgetcode", "budgetcodeto"], // Champ 5525 print conditions: budget code range (from alone = that code)
};
const OPTION_KEYS: readonly (keyof ReportFilters)[] = ["budgetcode", "budgetcodeto", "hideopening", "hidezero", "controltotals", "summary", "gross", "daytotals", ...BREAKDOWN_KEYS];
// periodfrom/periodto = รหัสงวดที่เลือกบนจอ (Champ "พิมพ์จากงวด") ใช้แสดงหัวรายงานเท่านั้น — backend รับเฉพาะ from/to ที่แปลงจากงวดแล้ว จึงไม่ส่งไป
const PERIOD_PICK_KEYS: readonly (keyof ReportFilters)[] = ["periodfrom", "periodto"];
export function reportOptionOn(name: string, filters: ReportFilters, key: keyof ReportFilters) {
  return filters[key] === "true" && (REPORT_OPTIONS[name] ?? []).includes(key);
}
/** Query of one report: empty values and options the report does not take are dropped.
 * Ledger: ไม่แสดงยอดยกมา ใช้กับแบบละเอียด (5534); ผลรวมบัญชีคุมและยอด ณ วันที่ ใช้กับแบบสรุป (5535). */
export function reportQuery(name: string, filters: ReportFilters): Record<string, string> {
  const ledgerSummary = name === "ledger" && reportOptionOn(name, filters, "summary");
  const query: Record<string, string> = {};
  for (const [key, value] of Object.entries(filters) as [keyof ReportFilters, string][]) {
    if (!value || PERIOD_PICK_KEYS.includes(key) || OPTION_KEYS.includes(key) && !(REPORT_OPTIONS[name] ?? []).includes(key)) continue;
    if (name === "ledger" && (ledgerSummary ? key === "hideopening" || key === "daytotals" || key === "from" : key === "controltotals" || BREAKDOWN_KEYS.includes(key))) continue;
    if (key === "dimtotals" && filters.bydepartment !== "true") continue; // the "with totals" row follows the department split only (Champ m_iSumDepart)
    query[key] = value;
  }
  return query;
}
// purpose marks a whole-report pull for print / export: the backend checks report:print / report:export (Champ สิทธิ์เฉพาะ)
export async function fetchReport(name: string, filters: ReportFilters, page = 1, limit = 50, snapshot?: number, companywide = false, purpose?: "print" | "export") {
  return glRequest<GLReport>(`reports/${name}?${new URLSearchParams({ ...reportQuery(name, filters), ...(companywide ? { companywide: "true" } : {}), page: String(page), limit: String(limit), ...(snapshot === undefined ? {} : { snapshot: String(snapshot) }), ...(purpose ? { purpose } : {}) })}`);
}
// แถวพิเศษจาก backend: control = รวมบัญชีคุม (accountBalanceReport), opening/carry/total = แยกประเภท (ledger)
// แถวพิเศษของรายงานแยกประเภท (Champ 5534, backend ledger) — backend ส่งคำอธิบายว่าง ข้อความแปลที่นี่
const LEDGER_ROW_LABELS: Record<string, GLLabel> = { opening: ["gl_opening_balance_2", "ยอดยกมา"], carry: ["gl_closing_balance", "ยอดยกไป"], total: ["gl_total", "ยอดรวม"], daytotal: ["gl_report_day_total", "รวมรายวัน"] };
export function controlRowText(row: Record<string, string>, key: string, value: string, tr: GLTextFn) {
  // แถวรวมบัญชีคุมแบบ Champ: ไม่พิมพ์รหัส มีแต่ชื่อบัญชีคุม (ตัวหนา) — หัวบัญชีคุม (rowtype=header) พิมพ์รหัส + ชื่อก่อนบัญชีลูกแล้ว
  if (row.rowtype === "control" && key === "accountcode") return "";
  // รวมของเอกสารในรายงานข้อมูลรายวัน (Champ 5532: "รวมทั้งสิ้น n รายการ" ต่อเอกสาร)
  if (row.rowtype === "doctotal" && key === "description") return tr("gl_report_doc_total", "รวมทั้งสิ้น {0} รายการ").replace("{0}", row.linecount ?? "");
  // รวมของบัญชีในรายงานเปรียบเทียบงบประมาณ (Champ 5530)
  if (row.rowtype === "budgettotal" && key === "budgetname") return tr("gl_report_total", "รวม");
  // รวมรายวันของรายงานข้อมูลรายวัน (Champ 5532/5533) บอกจำนวนรายการของวันนั้นด้วย; แยกประเภทไม่ส่ง linecount
  if (row.rowtype === "daytotal" && key === "description" && row.linecount) return tr("gl_report_day_total_count", "รวมรายวัน {0} รายการ").replace("{0}", row.linecount);
  const ledgerLabel = LEDGER_ROW_LABELS[row.rowtype ?? ""];
  return ledgerLabel && key === "description" ? tr(...ledgerLabel) : value;
}
/** Rows for file export carry the same labels as the screen (opening/carry/day totals, document totals, control rows). */
export function labelledReportRows(rows: Record<string, string>[], tr: GLTextFn) {
  return rows.map((row) => Object.fromEntries(Object.entries(row).map(([key, value]) => [key, controlRowText(row, key, value ?? "", tr)])));
}
type ReportColumn = GLReport["columns"][number];
/** ชื่อเต็มของคอลัมน์ที่มีหัวกลุ่ม เช่น "ยอดยกมา เดบิต" — ใช้กับการ์ดยอดรวมและไฟล์ส่งออก */
export function columnTitle(column: ReportColumn | undefined) {
  return column && (column.group ? `${column.group} ${column.label}` : column.label);
}
/** หัวตาราง 2 ชั้นแบบ Champ (ยอดยกมา คร่อม เดบิต | เครดิต); คอลัมน์ที่ไม่มีกลุ่มกินสองแถว */
export function ReportHeadRows({ columns, cellClass }: { columns: ReportColumn[]; cellClass: (column: ReportColumn) => string }) {
  if (!columns.some((column) => column.group)) return <tr>{columns.map((column) => <th key={column.key} className={cellClass(column)}>{column.label}</th>)}</tr>;
  const bands: { column: ReportColumn; span: number }[] = [];
  for (const column of columns) {
    const last = bands.at(-1);
    if (column.group && last?.column.group === column.group) last.span++; else bands.push({ column, span: 1 });
  }
  return <>
    <tr>{bands.map(({ column, span }) => column.group ? <th key={column.key} colSpan={span} className={cellClass(column)} style={{ textAlign: "center" }}>{column.group}</th> : <th key={column.key} rowSpan={2} className={cellClass(column)}>{column.label}</th>)}</tr>
    <tr>{columns.filter((column) => column.group).map((column) => <th key={column.key} className={cellClass(column)}>{column.label}</th>)}</tr>
  </>;
}
/** ยอดคงเหลือ/ผลต่างงบประมาณที่ติดลบแสดงในวงเล็บแบบ Champ เช่น (1,000.00) */
export function balanceText(key: string, text: string) {
  return (key === "balance" || key === "variance") && text.startsWith("-") ? `(${text.slice(1)})` : text;
}
/** แถวรวม/ยอดยกของรายงาน พิมพ์ตัวหนาแบบ Champ */
// _kind/_bold = row markers of the central registry: group and control-account totals, and detail rows the registry asks to print bold (chart of accounts control accounts)
const SUMMARY_KINDS: ReadonlySet<string> = new Set(["group_total", "control_total"]);
export function isSummaryRow(row: Record<string, string>) {
  return row.rowtype === "control" || row.rowtype === "document" || row.rowtype === "doctotal" || row.rowtype === "budgetaccount" || row.rowtype === "budgettotal" || row.rowtype in LEDGER_ROW_LABELS || row._bold === "true" || SUMMARY_KINDS.has(row._kind ?? "");
}
// รายงานข้อมูลรายวันแบบ Champ: วันที่/สมุด/เลขที่/เอกสารอ้างอิง อยู่ที่หัวเอกสารแล้ว
const DOCUMENT_HEADER_KEYS: ReadonlySet<string> = new Set(["date", "bookcode", "docno", "referenceno"]);
/** รายงานแยกประเภทแสดงรหัส/ชื่อบัญชีครั้งเดียวที่แถวแรกของบัญชี (หัวบัญชีแบบ Champ) และแถวแรกของหน้า */
export function repeatsAccount(row: Record<string, string>, key: string, firstOnPage: boolean) {
  // แถวรวมของเอกสารไม่พิมพ์หัวเอกสารซ้ำ; บรรทัดรายการพิมพ์หัวเอกสารเฉพาะแถวแรกของหน้า (หน้าที่เริ่มกลางเอกสาร)
  if (row.rowtype === "doctotal" && DOCUMENT_HEADER_KEYS.has(key)) return true;
  if (row.rowtype === "docline" && DOCUMENT_HEADER_KEYS.has(key)) return !firstOnPage;
  // เปรียบเทียบงบประมาณ: รหัส/ชื่อผังบัญชีอยู่ที่หัวบัญชี (budgetaccount) แล้ว
  const accountKey = key === "accountcode" || key === "accountname";
  if (row.rowtype === "budgettotal" && accountKey) return true;
  if (row.rowtype === "budgetline" && accountKey) return !firstOnPage;
  return !firstOnPage && row.accountfirst === "" && (key === "accountcode" || key === "accountname");
}
export function accountIndent(row: Record<string, string>, key: string) {
  const level = Number(row.accountlevel);
  return key === "accountname" && level > 1 ? { paddingInlineStart: `calc(0.5rem + ${level - 1}em)` } : undefined;
}
function reportRows(report: GLReport) { return report.rows ?? []; }
// การ์ดยอดรวมเรียงตามลำดับคอลัมน์ (ผ่านบัญชี → ตัดยอด → คงเหลือ) ไม่ใช่ตามตัวอักษรของ key; key ที่ไม่ใช่คอลัมน์ต่อท้าย
function totalRank(report: GLReport, key: string) { const index = report.columns.findIndex((column) => column.key === key); return index < 0 ? report.columns.length : index; }
// ยอดคงค้างลูกหนี้/เจ้าหนี้/Statement เป็นยอด ณ วันที่ (backend อ่านแค่ to + branchcode) จึงไม่ต้องเลือกปีบัญชี
const AS_OF_REPORTS: ReadonlySet<string> = new Set(["ar-outstanding", "ap-outstanding", "bank-unmatched"]);
function localToday() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
}
const reportTextLabels: Record<string, Record<string, GLLabel>> = {
  accounttype: accountTypeLabels,
  status: { draft: ["gl_draft", "ฉบับร่าง"], posted: ["gl_posted", "ผ่านรายการแล้ว"], reversed: ["gl_reversed", "กลับรายการแล้ว"], void: ["gl_cancel_draft", "ยกเลิกร่าง"] },
  direction: { in: ["gl_money_in", "เงินเข้า"], out: ["gl_money_out", "เงินออก"], 1: ["gl_money_in", "เงินเข้า"], 2: ["gl_money_out", "เงินออก"] },
  document_kind: { 1: ["gl_details_ui_1", "ตั้งหนี้"], 2: ["gl_details_ui_2", "ผลชำระที่เกิดแล้ว"], 3: ["gl_details_ui_3", "เพิ่มหนี้"], 4: ["gl_details_ui_4", "ลดหนี้"], 5: ["gl_opening_balance_2", "ยอดยกมา"] },
  category: { operating: ["gl_operating", "ดำเนินงาน"], investing: ["gl_investing", "ลงทุน"], financing: ["gl_raise_funds", "จัดหาเงิน"], unclassified: ["gl_not_specified", "ยังไม่ระบุ"] },
};
export function reportText(key: string, value = "", tr: GLTextFn, books: GLJournalBook[] = [], language = "th") {
  // The computed earnings row is not a chart-of-accounts entry. Keep its source key for exports.
  if (key === "accountcode" && value === "__current_earnings__") return "—";
  // สมุดรายวันเป็นข้อมูลหลักของบริษัท: แสดงชื่อสมุดจริง (ไม่รู้จักรหัส = แสดงรหัส)
  if (key === "bookcode") return journalBookName(findJournalBook(books, value), value, language);
  return reportTextLabels[key] ? labelText(reportTextLabels[key], value, tr) : value;
}

export function JournalDrillDownModal({
  docno,
  journalId,
  open,
  onClose,
  books = [],
}: {
  docno: string | null;
  journalId: string | null;
  open: boolean;
  onClose: () => void;
  books?: GLJournalBook[];
}) {
  const tr = useGLText();
  const language = useGLLanguage();
  const voucherPrint = useGLPrint();
  const [journal, setJournal] = useState<GLJournal | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open || !docno) {
      setJournal(null);
      setError("");
      return;
    }
    let active = true;
    setLoading(true);
    setError("");

    setJournal(null);
    if (!journalId) { setError(tr("gl_err_not_found", "ไม่พบรายการบัญชี")); setLoading(false); return; }
    glRequest<GLJournal>(`journals/${encodeURIComponent(journalId)}`)
      .then((value) => { if (active) setJournal(value); })
      .catch((err) => {
        if (active) setError((err as Error).message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [open, docno, journalId, tr]);

  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, onClose]);

  if (!open || !docno) return null;

  const totals = journal?.lines ? journalTotals(journal.lines) : null;
  const isBalanced = totals ? totals.debit === totals.credit : false;

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-200">
      <div className="max-w-4xl w-full max-h-[90vh] bg-card border border-border rounded-2xl shadow-2xl flex flex-col overflow-hidden text-card-foreground">
        {/* Header */}
        <div className="p-4 border-b border-border bg-muted/30 flex items-center justify-between gap-3 shrink-0">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-primary/10 text-primary">
              <FileText className="size-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-lg font-bold text-foreground">{tr("gl_voucher_detail", "รายละเอียดใบสำคัญรายวัน")}</h3>
                <span className="font-mono text-sm px-2 py-0.5 rounded-md bg-muted border border-border font-bold text-primary">
                  {docno}
                </span>
                {journal?.bookcode && (
                  <span className="text-xs px-2 py-0.5 rounded-md bg-primary/10 text-primary border border-primary/20 font-medium">
                    {journal.bookcode}
                  </span>
                )}
                {journal?.status && (
                  <span className={`text-xs px-2 py-0.5 rounded-md font-medium border ${
                    journal.status === "posted"
                      ? "bg-primary/10 text-primary border-primary/20"
                      : "bg-muted text-muted-foreground border-border"
                  }`}>
                    {journal.status === "posted" ? tr("gl_posted", "ผ่านรายการแล้ว") : journal.status === "reversed" ? tr("gl_reversed", "กลับรายการแล้ว") : tr("gl_draft", "ฉบับร่าง")}
                  </span>
                )}
              </div>
              <p className="text-xs text-muted-foreground mt-0.5">
                {journal?.date ? `${tr("gl_date", "วันที่")} ${journal.date}` : tr("gl_voucher_detail", "รายละเอียดใบสำคัญรายวัน")}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-1.5">
            <Button type="button" variant="outline" className={actionClass} disabled={!journal} onClick={() => journal && voucherPrint.print(<GLVoucherPrint journal={journal} books={books} company={printCompanyName()} tr={tr} language={language} />)}>
              <Printer className="size-4 mr-1.5" />{tr("gl_print_voucher", "พิมพ์ใบสำคัญ")}
            </Button>
            <Button variant="ghost" size="icon" className="size-8 rounded-lg" onClick={onClose} aria-label={tr("gl_close_view_dialog", "ปิดหน้าต่างแสดงข้อมูล")} title={tr("gl_close_view_dialog", "ปิดหน้าต่างแสดงข้อมูล")}>
              <X className="size-4" />
            </Button>
          </div>
          {voucherPrint.portal}
        </div>

        {/* Body */}
        <div className="flex-1 min-h-0 overflow-auto p-4 flex flex-col gap-4">
          {loading && (
            <div className="p-8 text-center text-muted-foreground flex flex-col items-center gap-2">
              <RefreshCw className="size-6 animate-spin text-primary" />
              <span>{tr("gl_processing", "กำลังประมวลผล…")}</span>
            </div>
          )}
          {error && !loading && <Notice error text={error} />}

          {journal && !loading && (
            <>
              {/* Metadata Cards */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs">
                <div className="p-2.5 rounded-xl border border-border bg-muted/30">
                  <span className="text-muted-foreground block">{tr("gl_fiscal_year", "ปีบัญชี")}</span>
                  <span className="font-semibold text-foreground text-sm font-mono">{journal.fiscalyear || "—"}</span>
                </div>
                <div className="p-2.5 rounded-xl border border-border bg-muted/30">
                  <span className="text-muted-foreground block">{tr("gl_branch_code", "รหัสสาขา")}</span>
                  <span className="font-semibold text-foreground text-sm font-mono">{journal.branchcode || tr("gl_head_office", "สำนักงานใหญ่")}</span>
                </div>
                <div className="p-2.5 rounded-xl border border-border bg-muted/30 col-span-2">
                  <span className="text-muted-foreground block">{tr("gl_reference", "เอกสารอ้างอิง")}</span>
                  <span className="font-semibold text-foreground text-sm">{journal.reference || "—"}</span>
                </div>
                {journal.description && (
                  <div className="p-2.5 rounded-xl border border-border bg-muted/30 col-span-2 sm:col-span-4">
                    <span className="text-muted-foreground block">{tr("gl_description", "คำอธิบายรายการ")}</span>
                    <span className="text-foreground text-sm">{journal.description}</span>
                  </div>
                )}
              </div>

              {/* Journal Lines Table */}
              <div className="rounded-xl border border-border overflow-hidden shadow-sm">
                <table className="w-full text-left text-xs leading-normal">
                  <thead className="bg-muted text-muted-foreground font-medium border-b border-border">
                    <tr>
                      <th className="p-2.5 w-10 text-center">#</th>
                      <th className="p-2.5 w-36">{tr("gl_account_code", "รหัสบัญชี")}</th>
                      <th className="p-2.5 min-w-36">{tr("gl_account_name", "ชื่อบัญชี")}</th>
                      <th className="p-2.5 min-w-40">{tr("gl_description", "คำอธิบาย")}</th>
                      <th className="p-2.5 text-right w-28">{tr("gl_debit", "เดบิต")}</th>
                      <th className="p-2.5 text-right w-28">{tr("gl_credit", "เครดิต")}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {journal.lines.map((line, idx) => {
                      const hasDebit = line.debit && line.debit !== "0";
                      const hasCredit = line.credit && line.credit !== "0";
                      return (
                        <tr key={idx} className="hover:bg-accent/40">
                          <td className="p-2 text-center text-muted-foreground">{idx + 1}</td>
                          <td className="p-2 font-mono font-bold text-primary">{line.accountcode}</td>
                          <td className="p-2 text-foreground">{line.accountname || "—"}</td>
                          <td className="p-2 text-muted-foreground truncate max-w-48" title={line.description}>{line.description || "—"}</td>
                          <td className={`p-2 text-right font-mono tabular-nums ${hasDebit ? "font-semibold text-foreground" : "text-muted-foreground"}`}>
                            {hasDebit ? formatAmount(line.debit) : "—"}
                          </td>
                          <td className={`p-2 text-right font-mono tabular-nums ${hasCredit ? "font-semibold text-foreground" : "text-muted-foreground"}`}>
                            {hasCredit ? formatAmount(line.credit) : "—"}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                  {totals && (
                    <tfoot className="bg-muted/50 border-t border-border font-semibold text-foreground">
                      <tr>
                        <td colSpan={4} className="p-2.5 text-right">{tr("gl_total", "ยอดรวม")}</td>
                        <td className="p-2.5 text-right font-mono tabular-nums text-foreground">{formatAmount(amountString(totals.debit))}</td>
                        <td className="p-2.5 text-right font-mono tabular-nums text-foreground">{formatAmount(amountString(totals.credit))}</td>
                      </tr>
                    </tfoot>
                  )}
                </table>
              </div>

              {/* Balanced Check Status */}
              <div className="flex items-center justify-between p-2.5 rounded-xl border border-border bg-muted/20 text-xs">
                <span className="text-muted-foreground">{tr("gl_accounting_check", "การตรวจสอบสมดุลบัญชี")}:</span>
                {isBalanced ? (
                  <span className="inline-flex items-center gap-1.5 font-semibold text-primary bg-primary/10 px-2.5 py-1 rounded-lg border border-primary/20">
                    <CheckCircle2 className="size-4" /> {tr("gl_debit_equals_credit", "เดบิต = เครดิต สมดุล 100%")}
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1.5 font-semibold text-destructive bg-destructive/10 px-2.5 py-1 rounded-lg border border-destructive/30">
                    <AlertTriangle className="size-4" /> {tr("gl_debit_not_equals_credit", "เดบิตและเครดิตไม่สมดุล")}
                  </span>
                )}
              </div>
            </>
          )}
        </div>

        {/* Footer */}
        <div className="p-3 border-t border-border bg-muted/30 flex justify-end shrink-0">
          <Button variant="outline" size="sm" onClick={onClose}>
            {tr("gl_close", "ปิด")}
          </Button>
        </div>
      </div>
    </div>
  );
}

/** Leading cells a caption row spans (master-list totals `_label`, footer `label`): the first column plus the empty cells after it. 0 = not a caption row. */
export function captionSpan(caption: string | undefined, row: Record<string, string>, columns: { key: string }[]): number {
  if (!caption) return 0;
  let span = 1;
  while (span < columns.length && !(row[columns[span].key] ?? "").trim()) span++;
  return span;
}
/** Text of one total card: counts are plain integers, amounts follow the scale of their column (default 2). */
function totalText(report: GLReport, key: string, value: string, tr: GLTextFn) {
  if (key === "unclassifiedlines") return tr("gl_x_items", "{0} รายการ").replace("{0}", String(value));
  if (report.countTotals?.includes(key)) return formatAmount(value, 0);
  return formatAmount(value, report.columns.find((column) => column.key === key)?.scale);
}
function totalTitle(report: GLReport, key: string, tr: GLTextFn) {
  return report.totalLabels?.[key] ?? columnTitle(report.columns.find((column) => column.key === key)) ?? labelText(totalLabels, key, tr, tr("gl_total", "ยอดรวม"));
}

export function ReportGrid({
  report,
  graphs = false,
  onDrillDocNo,
  onDrillAccount,
  books = [],
}: {
  report: GLReport;
  graphs?: boolean;
  books?: GLJournalBook[];
  onDrillDocNo?: (docNo: string, journalId: string) => void;
  onDrillAccount?: (accountCode: string) => void;
}) {
  const tr = useGLText();
  const language = useGLLanguage();
  const density = useRowDensity();
  const reportPref = useReportPreferences();
  const amountColumns = report.columns.filter((column) => column.amount);
  const amounts = reportRows(report).map((row) => { try { return displayAmountUnits(row[amountColumns[0]?.key] ?? "0"); } catch { return 0n; } });
  const max = amounts.reduce((largest, amount) => (amount < 0n ? -amount : amount) > largest ? (amount < 0n ? -amount : amount) : largest, 0n);
  // footer rows of the central registry carry a ready-made `label`; the legacy working-paper footer (`labelkey`) is print-only
  const footerRows = (report.footer ?? []).filter((row) => row.label);
  const cells = (row: Record<string, string>, index: number) => {
    const caption = captionSpan(row._label, row, report.columns);
    return report.columns.map((column, columnIndex) => {
      if (caption && columnIndex > 0 && columnIndex < caption) return null;
      if (caption && columnIndex === 0) return <td key={column.key} colSpan={caption} className="p-2">{row._label}</td>;
      const val = repeatsAccount(row, column.key, index === 0) ? "" : row[column.key] ?? "";
      const isDocNoCol = (column.key === "docno" || column.key === "documentno" || column.key === "reference") && val && val.trim() !== "";
      // แถวรวมบัญชีคุมไม่มีรายการของตัวเอง จึงไม่เปิดรายงานแยกประเภท
      const isAccountCodeCol = column.key === "accountcode" && val && val !== "__current_earnings__" && row.rowtype !== "control" && row.rowtype !== "header" && (!row._kind || row._kind === "detail");

      let cellContent;
      if (column.amount) {
        const amountText = balanceText(column.key, formatAmount(val, column.scale));
        // bank-unmatched: a partly matched statement line keeps a smaller remaining amount than its full amount
        const isPartial = val !== row.amount && val !== "0";
        cellContent = column.key === "remaining_amount" && row.amount ? (
          <div className="inline-flex items-center justify-end gap-1.5">
            <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[11px] font-medium ${isPartial ? "bg-primary/15 text-primary" : "bg-muted text-muted-foreground border border-border"}`}>
              {isPartial ? tr("gl_bank_unmatched_status_partial", "จับคู่บางส่วน") : tr("gl_bank_unmatched_status_unmatched", "ยังไม่จับคู่")}
            </span>
            <span>{amountText}</span>
          </div>
        ) : amountText;
      } else if (isDocNoCol && onDrillDocNo && (row.journalid || row.id)) {
        cellContent = (
          <button
            type="button"
            className="group inline-flex items-center gap-1 font-mono font-semibold text-primary hover:text-primary/80 hover:underline cursor-pointer"
            onClick={(e) => { e.stopPropagation(); onDrillDocNo(val, row.journalid || row.id); }}
            title={tr("gl_voucher_detail", "รายละเอียดใบสำคัญรายวัน")}
          >
            <FileText className="size-3.5 opacity-70 group-hover:opacity-100 shrink-0" />
            <span>{val}</span>
          </button>
        );
      } else if (isAccountCodeCol && onDrillAccount) {
        cellContent = (
          <div className="flex items-center justify-between gap-1.5 group">
            <span className="font-mono font-medium">{val}</span>
            <button
              type="button"
              className="opacity-0 group-hover:opacity-100 transition-opacity text-xs text-primary hover:text-primary/80 inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-primary/10 hover:bg-primary/20 cursor-pointer"
              onClick={(e) => { e.stopPropagation(); onDrillAccount(val); }}
              title={tr("gl_drill_to_ledger", "ดูรายงานแยกประเภท")}
            >
              <ExternalLink className="size-3 shrink-0" />
              <span className="text-[11px] font-medium">{tr("gl_drill_to_ledger", "ดูรายงานแยกประเภท")}</span>
            </button>
          </div>
        );
      } else {
        cellContent = reportText(column.key, controlRowText(row, column.key, val, tr), tr, books, language);
      }

      return (
        <td key={column.key} style={accountIndent(row, column.key)} className={`p-2 ${column.amount ? "whitespace-nowrap text-right tabular-nums" : "min-w-28"}`}>
          {cellContent}
        </td>
      );
    });
  };
  return <div className="flex flex-col flex-1 min-h-0 gap-3">
    <div className="shrink-0 flex flex-col gap-2">{(report.warnings ?? []).map((warning, index) => <Notice key={index} text={warning} />)}</div>
    {!!Object.keys(report.totals ?? {}).length && <div className="shrink-0 flex flex-wrap gap-2">{Object.entries(report.totals).sort(([left], [right]) => totalRank(report, left) - totalRank(report, right)).map(([key, value]) => <div key={key} className="min-w-36 flex-1 rounded-xl border border-border bg-muted/40 p-3 shadow-sm"><div className="text-[0.9rem] text-muted-foreground">{totalTitle(report, key, tr)}</div><div className="text-lg font-semibold tabular-nums">{totalText(report, key, value, tr)}</div></div>)}</div>}
    {graphs && max > 0n && <div className="shrink-0 grid gap-2 rounded-xl border border-border p-3 shadow-sm" aria-label={tr("gl_balance_comparison_chart", "กราฟเปรียบเทียบยอดบัญชี")}>{reportRows(report).slice(0, 20).map((row, index) => <div key={index} className="grid gap-1"><div className="flex flex-wrap justify-between gap-2 text-[0.95rem]"><span>{row.name ?? row.accountname ?? row.month ?? row[report.columns.find((column) => !column.amount)?.key ?? ""] ?? tr("gl_item_x", "รายการ {0}").replace("{0}", String(index + 1))}</span><strong>{formatAmount(row[amountColumns[0]?.key] ?? "0")}</strong></div><div className="h-3 rounded-full bg-muted"><div className="h-3 rounded-full bg-primary" style={{ width: `${((amounts[index] < 0n ? -amounts[index] : amounts[index]) * 100n / max).toString()}%` }} /></div></div>)}</div>}
    <div className="shrink-0 flex justify-end">
      <ReportDisplayToolbar
        fontSize={reportPref.fontSize}
        onFontSizeChange={reportPref.setFontSize}
        highContrast={reportPref.highContrast}
        onToggleHighContrast={reportPref.toggleHighContrast}
      />
    </div>
    <div className={`flex-1 min-h-[220px] overflow-auto rounded-xl border border-border shadow-sm ${reportPref.contrastClass}`}>
      <table className={`w-full text-left leading-normal ${density.tableClass} ${reportPref.fontSizeClass}`}>
        <thead className="sticky top-0 bg-muted z-10">
          <ReportHeadRows columns={report.columns} cellClass={(column) => `whitespace-nowrap p-2 ${column.amount ? "text-right" : "text-left"}`} />
        </thead>
        <tbody>
          {reportRows(report).map((row, index) => (
            <tr key={index} className={`border-t border-border hover:bg-accent/60 transition-colors ${isSummaryRow(row) ? "font-semibold bg-muted/40" : ""}`}>
              {cells(row, index)}
            </tr>
          ))}
          {footerRows.map((row, index) => {
            const caption = captionSpan(row.label, row, report.columns);
            return <tr key={`footer-${index}`} className="border-t border-border font-semibold bg-muted/40">
              {report.columns.map((column, columnIndex) => {
                if (columnIndex > 0 && columnIndex < caption) return null;
                if (columnIndex === 0) return <td key={column.key} colSpan={caption} className="p-2">{row.label}</td>;
                return <td key={column.key} className={`p-2 ${column.amount ? "whitespace-nowrap text-right tabular-nums" : ""}`}>{column.amount ? formatAmount(row[column.key] ?? "", column.scale) : row[column.key] ?? ""}</td>;
              })}
            </tr>;
          })}
          {!reportRows(report).length && <tr><td className="p-5 text-center text-muted-foreground" colSpan={report.columns.length || 1}>{tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")}</td></tr>}
        </tbody>
      </table>
    </div>
    <p className="shrink-0 text-[0.9rem] text-muted-foreground">{tr("gl_data_as_of", "ข้อมูล ณ {0}").replace("{0}", String(report.asof || tr("gl_last_processing_time", "เวลาประมวลผลล่าสุด")))}</p>
  </div>;
}
const totalLabels: Record<string, GLLabel> = { debit: ["gl_total_debit", "รวมเดบิต"], credit: ["gl_total_credit", "รวมเครดิต"], balance: ["gl_balance", "ยอดคงเหลือ"], income: ["gl_revenue", "รายได้"], revenue: ["gl_revenue", "รายได้"], expense: ["gl_expense", "ค่าใช้จ่าย"], profit: ["gl_net_profit", "กำไรสุทธิ"], assets: ["gl_asset", "สินทรัพย์"], liabilities: ["gl_liability", "หนี้สิน"], equity: ["gl_equity", "ส่วนของเจ้าของ"], difference: ["gl_difference", "ผลต่าง"], budget: ["gl_budget_2", "งบประมาณ"], actual: ["gl_actual", "ยอดจริง"], budgetamount: ["gl_budget_2", "งบประมาณ"], actualamount: ["gl_actual", "ยอดจริง"], variance: ["gl_difference", "ผลต่าง"], opening: ["gl_opening_balance_2", "ยอดยกมา"], closing: ["gl_closing_balance", "ยอดยกไป"], cashin: ["gl_money_in", "เงินเข้า"], cashout: ["gl_money_out", "เงินออก"], netcash: ["gl_net_cash", "เงินสดสุทธิ"], cash: ["gl_cash_and_bank", "เงินสดและเงินฝากธนาคาร"], currentearnings: ["gl_unclosed_profit_loss", "กำไรขาดทุนที่ยังไม่ปิด"], unclassifiedlines: ["gl_unclassified_line", "บรรทัดที่ยังไม่ระบุประเภท"] };
// Print of the reports that stay on the legacy engine (profit and loss, financial position, outstanding receivables/payables, unmatched statement lines); the Champ-parity reports print from the central registry (app/report/report-center-print.tsx)
// สมุดบัญชีตามประกาศกรมทะเบียนการค้า 2544 ข้อ 5: หัวทุกหน้า (thead ซ้ำทุกหน้า) = ชื่อผู้จัดทำบัญชี + ชนิดของบัญชี + ช่วงเวลา; เลขหน้าจาก gl-print
const PRINT_MAX_ROWS = 10000;
function printCell(key: string, value: string, tr: GLTextFn, books: GLJournalBook[], language: LanguageCode) {
  if (/(^|_)date$/.test(key) && /^\d{4}-\d{2}-\d{2}$/.test(value)) return formatAppDate(value, language);
  return reportText(key, value, tr, books, language);
}
function GLReportPrint({ report, title, meta, company, books, tr, language }: { report: GLReport; title: string; meta: ReportPrintMeta; company: string; books: GLJournalBook[]; tr: GLTextFn; language: LanguageCode }) {
  const rows = reportRows(report);
  const columns = visiblePrintColumns(report.columns, rows);
  const totals = Object.entries(report.totals ?? {}).sort(([left], [right]) => totalRank(report, left) - totalRank(report, right));
  const inline = totals.filter(([key]) => columns.some((column) => column.amount && column.key === key));
  const extra = totals.filter(([key]) => !inline.some(([inlineKey]) => inlineKey === key));
  return <>
    <table>
      <thead>
        <tr className="gl-print-head"><th colSpan={columns.length || 1}><ReportPrintHeading company={company} title={title} printedBy={meta.printedBy} printedAt={meta.printedAt} description={meta.description} tr={tr} /></th></tr>
        <ReportHeadRows columns={columns} cellClass={(column) => column.amount ? "gl-print-num" : ""} />
      </thead>
      <tbody>
        {rows.map((row, index) => <tr key={index} className={isSummaryRow(row) ? "gl-print-total" : undefined}>{columns.map((column) => <td key={column.key} style={accountIndent(row, column.key)} className={column.amount ? "gl-print-num" : ""}>{column.amount ? balanceText(column.key, printAmountCell(column.key, row[column.key] ?? "")) : repeatsAccount(row, column.key, index === 0) ? "" : printCell(column.key, controlRowText(row, column.key, row[column.key] ?? "", tr), tr, books, language)}</td>)}</tr>)}
        {!rows.length && <tr><td colSpan={columns.length || 1}>{tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")}</td></tr>}
        {inline.length > 0 && <tr className="gl-print-total">{columns.map((column, index) => { const total = inline.find(([key]) => key === column.key); return <td key={column.key} className={column.amount ? "gl-print-num" : ""}>{total ? formatAmount(total[1]) : index === 0 ? tr("gl_report_grand_total", "รวมทั้งสิ้น") : ""}</td>; })}</tr>}
      </tbody>
    </table>
    {extra.length > 0 && <dl className="gl-print-totals">{extra.map(([key, value]) => <div key={key}><dt>{totalTitle(report, key, tr)}</dt><dd>{totalText(report, key, value, tr)}</dd></div>)}</dl>}
    <ReportPrintConditions lines={meta.conditions} tr={tr} />
  </>;
}
function reportPrintTitle(name: string, filters: ReportFilters, tr: GLTextFn, books: GLJournalBook[], accounts: GLAccount[], language: LanguageCode) {
  const item = GL_MENU_ITEMS.find((menu) => menu.route === `/report/${name}`);
  const base = item ? (item.label.key ? tr(item.label.key, item.label.th) : item.label.th) : name;
  // บัญชีเดียว (ช่วงที่ต้นเท่ากับท้าย) แสดงชื่อบัญชีในหัวรายงาน; ช่วงหลายบัญชีอยู่ในบรรทัดเงื่อนไข
  const single = filters.accountcode || (filters.accountfrom && filters.accountfrom === filters.accountto ? filters.accountfrom : "");
  const account = single ? accounts.find((row) => row.accountcode === single) : undefined;
  return [
    base,
    reportCodeList(filters.bookcode).map((code) => `${code} ${journalBookName(findJournalBook(books, code), code, language)}`).join(", "),
    single ? `${single} ${account ? accountName(account) : ""}`.trim() : "",
  ].filter(Boolean).join(" — ");
}
// กรอบหน้ารายงานแบบ Champ: หัว = ช่วงวันที่ (รายละเอียด) + พิมพ์โดย/วันที่พิมพ์; ตัวกรองอื่นไปอยู่ "เงื่อนไขการพิมพ์รายงาน" ท้ายรายงาน
type ReportPrintMeta = { description: string[]; conditions: string[]; printedBy: string; printedAt: string };
// หัวรายงานเมื่อเลือกตามงวด (Champ GLRepTrialBalanceDlg.cpp:565-580: "ตั้งแต่งวดที่ n ถึงงวดที่ m", งวดเดียวพิมพ์ "งวดที่ n"); null = เลือกตามวันที่
function periodHeading(filters: ReportFilters, cumulative: boolean, date: (value: string) => string, tr: GLTextFn): string | null {
  const { periodfrom: from, periodto: to } = filters;
  if (!from || !to) return null;
  if (cumulative) return tr("gl_for_period_as_of", "สำหรับงวดบัญชี {0} (ณ วันที่ {1})").replace("{0}", to).replace("{1}", date(filters.to));
  if (from === to) return tr("gl_print_period_single", "งวดที่ {0}").replace("{0}", from);
  return tr("gl_print_period_range", "ตั้งแต่งวดที่ {0} ถึงงวดที่ {1}").replace("{0}", from).replace("{1}", to);
}
function reportPrintMeta(filters: ReportFilters, asOf: boolean, tr: GLTextFn, language: LanguageCode): ReportPrintMeta {
  const date = (value: string) => formatAppDate(value, language);
  const range = !filters.accountcode && filters.accountfrom !== filters.accountto;
  return {
    description: [
      periodHeading(filters, asOf, date, tr)
        ?? (asOf ? tr("gl_print_as_of", "ณ วันที่ {0}").replace("{0}", date(filters.to))
        : filters.from && filters.to ? tr("gl_print_period", "ตั้งแต่วันที่ {0} ถึงวันที่ {1}").replace("{0}", date(filters.from)).replace("{1}", date(filters.to))
        : `${tr("gl_fiscal_year", "ปีบัญชี")} ${filters.fiscalyear}`),
    ],
    conditions: [
      range && filters.accountfrom ? `${tr("gl_account_from", "จากรหัสบัญชี")} ${filters.accountfrom}` : "",
      range && filters.accountto ? `${tr("gl_account_to", "ถึงรหัสบัญชี")} ${filters.accountto}` : "",
      filters.branchcode ? `${tr("gl_branch_code", "รหัสสาขา")} ${filters.branchcode}` : "",
      filters.departmentcode ? `${tr("gl_department_code", "รหัสแผนก")} ${reportCodeList(filters.departmentcode).join(", ")}` : "",
      filters.projectcode ? `${tr("gl_project_code", "รหัสโครงการ")} ${reportCodeList(filters.projectcode).join(", ")}` : "",
    ],
    printedBy: printUserName(),
    printedAt: printedAtText(language),
  };
}
/** Screen for the reports that are not in the central registry (name = last route segment: pnl, balancesheet, ar-outstanding, ap-outstanding, bank-unmatched). */
export function GLReports({ name, heading }: { name: string; heading?: string }) {
  const tr = useGLText();
  const refs = useReferences();
  const dims = useDimensionRegisters();
  const language = useGLLanguage();
  const [drillDocument, setDrillDocument] = useState<{ docno: string; journalId: string } | null>(null);
  const [drillAccount, setDrillAccount] = useState<GLAccount | null>(null);

  const asOf = AS_OF_REPORTS.has(name);
  const [filters, setFilters] = useState<ReportFilters>(() => ({ ...emptyReportFilters, ...(asOf ? { to: localToday() } : {}) })), [applied, setApplied] = useState<ReportFilters | null>(null);
  const [report, setReport] = useState<GLReport | null>(null), [page, setPage] = useState(1);
  const [busy, setBusy] = useState(false), [error, setError] = useState("");
  const [filterCollapsed, setFilterCollapsed] = useState(false);
  // ทะเบียนงวด ใช้เลือกช่วงรายงานตามงวดแบบ Champ (ไม่ติ๊ก "ดูจากวันที่เอกสาร" = ตามงวด); โหลดไม่ได้หรือปียังไม่มีงวด = ใช้ช่วงวันที่ตามเดิม
  const [periods, setPeriods] = useState<GLMaster[]>([]);
  const [byDocDate, setByDocDate] = useState(false);
  const reportPrint = useGLPrint();
  const set = (key: keyof ReportFilters, value: string) => setFilters((current) => ({ ...current, [key]: value }));
  const yearPeriodList = yearPeriods(periods, filters.fiscalyear);
  // the journal-book filter is picked like departments/projects: the company's books shaped as register entries (reports read closed books too)
  const bookRegister = useMemo<GLDimension[]>(() => (refs.books ?? []).map((book) => ({ ...book, name: journalBookName(book, book.code, language), isactive: true })), [refs.books, language]);
  const byPeriod = !asOf && !byDocDate && yearPeriodList.length > 0;
  const firstPeriod = yearPeriodList[0]?.code ?? "", lastPeriod = yearPeriodList.at(-1)?.code ?? "";

  useEffect(() => {
    if (asOf) return;
    let active = true;
    glAllRecords<GLMaster>("periods", "", 1000).then((items) => { if (active) setPeriods(items); }).catch(() => { if (active) setPeriods([]); });
    return () => { active = false; };
  }, [asOf]);

  // ตามงวด: ช่วงวันที่มาจากทะเบียนงวด (วันเริ่มของงวดแรก–วันสิ้นสุดของงวดสุดท้าย แบบ Champ OnGetPeriodDate); ช่องว่าง = งวดแรก/งวดสุดท้ายของปี
  function submit() {
    if (!byPeriod) { void load(); return; }
    const range = periodRangeDates(yearPeriodList, filters.periodfrom, filters.periodto);
    if (!range) { setError(tr("gl_start_period_not_after_end", "งวดเริ่มต้นต้องไม่เกินงวดสิ้นสุด")); return; }
    void load(1, { ...filters, from: range.from, to: range.to, periodfrom: range.fromcode, periodto: range.tocode });
  }

  // ติ๊ก "ดูจากวันที่เอกสาร" = เริ่มจากช่วงวันที่ของงวดที่เลือกอยู่ แล้วแก้เป็นวันใดก็ได้ (Champ OnChkDate)
  function changeByDocDate(on: boolean) {
    setByDocDate(on);
    const range = on ? periodRangeDates(yearPeriodList, filters.periodfrom, filters.periodto) : null;
    setFilters((current) => ({ ...current, periodfrom: "", periodto: "", ...(range ? { from: range.from, to: range.to } : {}) }));
  }
  const periodSelect = (label: string, value: string, onChange: (code: string) => void) => <Combobox aria-label={label} value={value} onChange={(picked: string | number) => onChange(String(picked))}>{yearPeriodList.map((period) => <option key={period.code} value={period.code}>{period.code}{period.name ? ` · ${period.name}` : ""}</option>)}</Combobox>;

  async function load(nextPage = 1, selected = filters) {
    if (!(asOf ? selected.to : selected.fiscalyear) || busy) return;
    if (selected.from && selected.to && selected.from > selected.to) { setError(tr("gl_start_date_not_after_end", "วันเริ่มต้นต้องไม่เกินวันสิ้นสุด")); return; }
    if (selected.accountfrom && selected.accountto && selected.accountfrom > selected.accountto) { setError(tr("gl_err_report_account_range_invalid", "รหัสบัญชีเริ่มต้นต้องไม่มากกว่ารหัสบัญชีสิ้นสุด")); return; }
    setBusy(true); setError("");
    try {
      const result = await fetchReport(name, selected, nextPage, 50, nextPage > 1 && report && JSON.stringify(selected) === JSON.stringify(applied) ? report.sequence : undefined);
      setReport(result);
      setApplied({ ...selected });
      setPage(nextPage);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  // เจาะบัญชีไปรายงานแยกประเภทในหน้าต่างซ้อน (ไม่ต้องออกจากรายงานเดิม); ยอด ณ วันที่ไม่มีปีบัญชี จึงไม่เจาะ
  const handleDrillAccount = (accountCode: string) => {
    if (!applied || busy || !accountCode || accountCode === "__current_earnings__") return;
    setDrillAccount(refs.accounts.find((account) => account.accountcode === accountCode) ?? null);
  };

  // ส่งออก/พิมพ์ต้องได้ทุกแถวจาก snapshot เดียวกัน (sequence) ไม่ใช่แค่หน้าที่แสดงอยู่
  // the backend refuses print / export without the right (Thai message shown via setError, buttons stay visible like Champ)
  async function completeReport(selected: ReportFilters, maxRows: number, tooMany: string, purpose: "print" | "export") {
    const complete = await fetchReport(name, selected, 1, 500, undefined, false, purpose);
    if (complete.totalrows > maxRows) throw new Error(tooMany);
    complete.rows = complete.rows ?? [];
    for (let next = 2; complete.rows.length < complete.totalrows; next++) {
      const chunk = await fetchReport(name, selected, next, 500, complete.sequence, false, purpose);
      if (chunk.totalrows !== complete.totalrows || chunk.sequence !== complete.sequence || !chunk.rows?.length) throw new Error(tr("gl_data_changed_during_export_retry", "ข้อมูลเปลี่ยนแปลงระหว่างส่งออก กรุณาลองใหม่"));
      complete.rows.push(...chunk.rows);
    }
    return complete;
  }

  async function printReport() {
    if (!applied || !report || busy) return;
    setBusy(true); setError("");
    try {
      const complete = await completeReport(applied, PRINT_MAX_ROWS, tr("gl_print_too_many", "รายการเกิน {0} รายการ กรุณาจำกัดช่วงวันที่ หรือเลือกสมุดรายวัน/บัญชีก่อนพิมพ์").replace("{0}", PRINT_MAX_ROWS.toLocaleString("en-US")), "print");
      const books = refs.books ?? [];
      const columns = visiblePrintColumns(complete.columns, complete.rows ?? []);
      reportPrint.print(<GLReportPrint report={complete} title={reportPrintTitle(name, applied, tr, books, refs.accounts, language)} meta={reportPrintMeta(applied, asOf, tr, language)} company={printCompanyName()} books={books} tr={tr} language={language} />, printOrientation(columns.length));
    } catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }

  async function exportCsv() {
    if (!applied || !report || busy) return;
    setBusy(true); setError("");
    try {
      const complete = await completeReport(applied, 100000, tr("gl_data_exceeds_100k_limit_date_range", "ข้อมูลเกิน 100,000 รายการ กรุณาจำกัดช่วงวันที่ก่อนส่งออก"), "export");
      downloadText(tr("gl_accounts_csv", "บัญชี-{0}-{1}.csv").replace("{0}", name).replace("{1}", String(asOf ? applied.to : applied.fiscalyear)), reportCsv({ ...complete, rows: labelledReportRows(complete.rows ?? [], tr) }), "text/csv;charset=utf-8");
    } catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }

  const resultReady = drillDocument === null && !!applied && !!report && !busy;
  useReportShortcuts({
    print: resultReady ? () => void printReport() : null,
    exportCsv: resultReady ? () => void exportCsv() : null,
    refresh: drillDocument === null && !busy && (asOf ? filters.to : filters.fiscalyear) ? submit : null,
  });

  const filterSummaryText = useMemo(() => {
    const parts: string[] = [];
    if (filters.fiscalyear) parts.push(`${tr("gl_fiscal_year", "ปีบัญชี")} ${filters.fiscalyear}`);
    if (filters.from && filters.to) parts.push(`${formatAppDate(filters.from, language)} - ${formatAppDate(filters.to, language)}`);
    else if (filters.to) parts.push(`${tr("gl_as_of_date", "ณ วันที่")} ${formatAppDate(filters.to, language)}`);
    if (filters.branchcode) parts.push(`${tr("gl_branch_code", "สาขา")} ${filters.branchcode}`);
    if (filters.accountfrom && filters.accountto) parts.push(`${tr("gl_account_code", "รหัสบัญชี")} ${filters.accountfrom} - ${filters.accountto}`);
    else if (filters.accountcode) parts.push(`${tr("gl_account_code", "รหัสบัญชี")} ${filters.accountcode}`);
    return parts.join(" · ");
  }, [filters, tr, language]);

  return <section className={`${panel} flex flex-col flex-1 min-h-0 gap-3`}>
    {(heading || report) && <div className="shrink-0 flex items-center justify-between gap-3">
      <div className="flex items-center gap-2.5 min-w-0">
        {heading && <h2 className="text-lg font-semibold truncate">{heading}</h2>}
        {report && <ReportFilterToggle collapsed={filterCollapsed} onToggle={() => setFilterCollapsed(!filterCollapsed)} tr={tr} />}
      </div>
    </div>}
    <Notice error text={error || refs.error || dims.error} />
    {report && filterCollapsed ? (
      <ReportFilterSummary summary={filterSummaryText} busy={busy} onExpand={() => setFilterCollapsed(false)} onRun={submit} onExport={() => void exportCsv()} onPrint={() => void printReport()} tr={tr} />
    ) : (
      <form className="shrink-0 grid gap-3" onSubmit={(event) => { event.preventDefault(); submit(); }}>
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {asOf ? <>
        <Field label={tr("gl_as_of_date", "ณ วันที่")}><input className={control} type="date" required value={filters.to} onChange={(e) => set("to", e.target.value)} /></Field>
        <Field label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} value={filters.branchcode} onChange={(e) => set("branchcode", e.target.value)} placeholder={tr("gl_all_branches", "ทุกสาขา")} /></Field>
        </> : <>
        <Field label={tr("gl_fiscal_year", "ปีบัญชี")}><YearSelect years={refs.years} value={filters.fiscalyear} onChange={(value) => { const year = refs.years.find((item) => item.code === value); setFilters((current) => ({ ...current, fiscalyear: value, from: year?.startdate ?? "", to: year?.enddate ?? "", periodfrom: "", periodto: "" })); }} /></Field>
        {yearPeriodList.length > 0 && <div className="flex items-end"><Check label={tr("gl_report_by_doc_date", "ดูจากวันที่เอกสาร")} checked={byDocDate} onChange={changeByDocDate} /></div>}
        {byPeriod ? <>
          <Field label={tr("gl_report_from_period", "จากงวด")}>{periodSelect(tr("gl_report_from_period", "จากงวด"), filters.periodfrom || firstPeriod, (code) => set("periodfrom", code))}</Field>
          <Field label={tr("gl_report_to_period", "ถึงงวด")}>{periodSelect(tr("gl_report_to_period", "ถึงงวด"), filters.periodto || lastPeriod, (code) => set("periodto", code))}</Field>
        </> : <>
        <Field label={tr("gl_from_date", "ตั้งแต่วันที่")}><input className={control} type="date" value={filters.from} onChange={(e) => set("from", e.target.value)} /></Field>
        <Field label={tr("gl_to_date", "ถึงวันที่")}><input className={control} type="date" value={filters.to} onChange={(e) => set("to", e.target.value)} /></Field>
        </>}
        <Field label={tr("gl_account_from", "จากรหัสบัญชี")}><AccountSelect accounts={refs.accounts} all label={tr("gl_account_from", "จากรหัสบัญชี")} value={filters.accountfrom} onChange={(value) => setFilters((current) => ({ ...current, accountcode: "", accountfrom: value }))} /></Field>
        <Field label={tr("gl_account_to", "ถึงรหัสบัญชี")}><AccountSelect accounts={refs.accounts} all label={tr("gl_account_to", "ถึงรหัสบัญชี")} value={filters.accountto} onChange={(value) => setFilters((current) => ({ ...current, accountcode: "", accountto: value }))} /></Field>
        <Field label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} value={filters.branchcode} onChange={(e) => set("branchcode", e.target.value)} placeholder={tr("gl_all_branches", "ทุกสาขา")} /></Field>
        <DimensionCodeMultiSelect report resource="departments" value={reportCodeList(filters.departmentcode)} register={dims.departments} onChange={(codes) => set("departmentcode", normalizeDimensionCodes(codes).join(","))} />
        <DimensionCodeMultiSelect report resource="projects" value={reportCodeList(filters.projectcode)} register={dims.projects} onChange={(codes) => set("projectcode", normalizeDimensionCodes(codes).join(","))} />
        <DimensionCodeMultiSelect report resource="books" value={reportCodeList(filters.bookcode)} register={bookRegister} onChange={(codes) => set("bookcode", normalizeDimensionCodes(codes).join(","))} />
        </>}
        <div className="flex flex-wrap items-end gap-2">
          <Button type="submit" className={actionClass} title="Enter / Alt+R" disabled={busy || !(asOf ? filters.to : filters.fiscalyear)}>
            <RefreshCw />
            {busy ? tr("gl_processing", "กำลังประมวลผล…") : tr("gl_show_report", "แสดงรายงาน")}
            <KeyHint keys="Enter" onPrimary />
          </Button>
          <Button type="button" variant="outline" className={actionClass} title="Alt+E" disabled={busy || !report} onClick={() => void exportCsv()}>
            <Download />
            {tr("gl_export_table", "ส่งออกตาราง")}
            <KeyHint keys="Alt+E" />
          </Button>
          <Button type="button" variant="outline" className={actionClass} title="Alt+P" disabled={busy || !report} onClick={() => void printReport()}>
            <Printer />
            {tr("gl_print", "พิมพ์")}
            <KeyHint keys="Alt+P" />
          </Button>
          {name === "bank-unmatched" && (
            <Button
              type="button"
              variant="outline"
              className={actionClass}
              onClick={() => { window.location.href = "/gl/journals"; }}
              title={tr("gl_bank_unmatched_goto_journal", "ไปที่สมุดรายวันรับ/จ่ายเงิน")}
            >
              <ExternalLink className="size-4 shrink-0" />
              {tr("gl_bank_unmatched_goto_journal", "ไปที่สมุดรายวันรับ/จ่ายเงิน")}
            </Button>
          )}
        </div>
      </div>
    </form>
    )}
    {report ? (
      <>
        <ReportGrid
          report={report}
          graphs={["financialgraphs", "dashboard", "executivesummary"].includes(name)}
          onDrillDocNo={(docno, journalId) => setDrillDocument({ docno, journalId })}
          onDrillAccount={asOf ? undefined : handleDrillAccount}
          books={refs.books ?? []}
        />
        <div className="shrink-0">
          <Pager page={page} total={report.totalrows} onPage={(next) => void load(next, applied!)} loading={busy} limit={50} />
        </div>
      </>
    ) : (
      <div className="rounded-xl border border-dashed border-border p-6 text-center text-muted-foreground">
        {asOf ? tr("gl_select_date_show_report", "เลือกวันที่ แล้วกดแสดงรายงาน") : tr("gl_select_fy_and_conditions_show_report", "เลือกปีบัญชีและเงื่อนไข แล้วกดแสดงรายงาน")}
      </div>
    )}

    {/* Drill Down Voucher Detail Modal */}
    <JournalDrillDownModal
      docno={drillDocument?.docno ?? null}
      journalId={drillDocument?.journalId ?? null}
      open={drillDocument !== null}
      onClose={() => setDrillDocument(null)}
      books={refs.books ?? []}
    />
    <AccountLedgerModal open={drillAccount !== null} onClose={() => setDrillAccount(null)} account={drillAccount} fiscalYear={applied?.fiscalyear ?? filters.fiscalyear} books={refs.books ?? []} />
    {reportPrint.portal}
  </section>;
}
