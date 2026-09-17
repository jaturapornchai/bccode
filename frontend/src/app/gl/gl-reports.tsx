"use client";

import { useState } from "react";
import { Download, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { accountTypeLabels, bookLabels, displayAmountUnits, labelText, type GLLabel, type GLTextFn, formatAmount, reportCsv, type GLReport } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { AccountSelect, Field, Notice, Pager, YearSelect, actionClass, control, downloadText, panel, useReferences, useRowDensity, useGLText } from "./gl-common";
import { useReportPreferences } from "@/hooks/use-report-preferences";
import { ReportDisplayToolbar } from "@/components/report-display-toolbar";

export type ReportFilters = { fiscalyear: string; from: string; to: string; accountcode: string; branchcode: string; departmentcode: string; projectcode: string; bookcode: string };
export const emptyReportFilters: ReportFilters = { fiscalyear: "", from: "", to: "", accountcode: "", branchcode: "", departmentcode: "", projectcode: "", bookcode: "" };
export async function fetchReport(name: string, filters: ReportFilters, page = 1, limit = 50, snapshot?: number) {
  return glRequest<GLReport>(`reports/${name}?${new URLSearchParams({ ...filters, page: String(page), limit: String(limit), ...(snapshot === undefined ? {} : { snapshot: String(snapshot) }) })}`);
}
function reportRows(report: GLReport) { return report.rows ?? []; }
const reportTextLabels: Record<string, Record<string, GLLabel>> = {
  accounttype: accountTypeLabels,
  bookcode: bookLabels,
  status: { draft: ["gl_draft", "ฉบับร่าง"], posted: ["gl_posted", "ผ่านรายการแล้ว"], reversed: ["gl_reversed", "กลับรายการแล้ว"], void: ["gl_cancel_draft", "ยกเลิกร่าง"] },
  direction: { in: ["gl_money_in", "เงินเข้า"], out: ["gl_money_out", "เงินออก"] },
  category: { operating: ["gl_operating", "ดำเนินงาน"], investing: ["gl_investing", "ลงทุน"], financing: ["gl_raise_funds", "จัดหาเงิน"], unclassified: ["gl_not_specified", "ยังไม่ระบุ"] },
};
function reportText(key: string, value = "", tr: GLTextFn) {
  // The computed earnings row is not a chart-of-accounts entry. Keep its source key for exports.
  if (key === "accountcode" && value === "__current_earnings__") return "—";
  return reportTextLabels[key] ? labelText(reportTextLabels[key], value, tr) : value;
}
export function ReportGrid({ report, graphs = false }: { report: GLReport; graphs?: boolean }) {
  const tr = useGLText();
  const density = useRowDensity();
  const reportPref = useReportPreferences();
  const amountColumns = report.columns.filter((column) => column.amount);
  const amounts = reportRows(report).map((row) => { try { return displayAmountUnits(row[amountColumns[0]?.key] ?? "0"); } catch { return 0n; } });
  const max = amounts.reduce((largest, amount) => (amount < 0n ? -amount : amount) > largest ? (amount < 0n ? -amount : amount) : largest, 0n);
  return <div className="flex flex-col flex-1 min-h-0 gap-3">
    <div className="shrink-0 flex flex-col gap-2">{(report.warnings ?? []).map((warning, index) => <Notice key={index} text={warning} />)}</div>
    {!!Object.keys(report.totals ?? {}).length && <div className="shrink-0 flex flex-wrap gap-2">{Object.entries(report.totals).map(([key, value]) => <div key={key} className="min-w-36 flex-1 rounded-xl border border-border bg-muted/40 p-3 shadow-sm"><div className="text-[0.9rem] text-muted-foreground">{report.columns.find((column) => column.key === key)?.label ?? labelText(totalLabels, key, tr, tr("gl_total", "ยอดรวม"))}</div><div className="text-lg font-semibold tabular-nums">{key === "unclassifiedlines" ? tr("gl_x_items", "{0} รายการ").replace("{0}", String(value)) : formatAmount(value)}</div></div>)}</div>}
    {graphs && max > 0n && <div className="shrink-0 grid gap-2 rounded-xl border border-border p-3 shadow-sm" aria-label={tr("gl_balance_comparison_chart", "กราฟเปรียบเทียบยอดบัญชี")}>{reportRows(report).slice(0, 20).map((row, index) => <div key={index} className="grid gap-1"><div className="flex flex-wrap justify-between gap-2 text-[0.95rem]"><span>{row.name ?? row.accountname ?? row.month ?? row[report.columns.find((column) => !column.amount)?.key ?? ""] ?? tr("gl_item_x", "รายการ {0}").replace("{0}", String(index + 1))}</span><strong>{formatAmount(row[amountColumns[0]?.key] ?? "0")}</strong></div><div className="h-3 rounded-full bg-muted"><div className="h-3 rounded-full bg-primary" style={{ width: `${((amounts[index] < 0n ? -amounts[index] : amounts[index]) * 100n / max).toString()}%` }} /></div></div>)}</div>}
    <div className="shrink-0 flex justify-end">
      <ReportDisplayToolbar
        fontSize={reportPref.fontSize}
        onFontSizeChange={reportPref.setFontSize}
        highContrast={reportPref.highContrast}
        onToggleHighContrast={reportPref.toggleHighContrast}
        compact={density.compact}
        onToggleCompact={density.toggle}
        compactLabel={{ compact: tr("gl_collapse_row", "ย่อบรรทัด"), expand: tr("gl_expand_row", "ขยายบรรทัด") }}
      />
    </div>
    <div className={`flex-1 min-h-[300px] overflow-auto rounded-xl border border-border shadow-sm ${reportPref.contrastClass}`}><table className={`w-full text-left leading-normal ${density.tableClass} ${reportPref.fontSizeClass}`}><thead className="sticky top-0 bg-muted"><tr>{report.columns.map((column) => <th key={column.key} className={`whitespace-nowrap p-2 ${column.amount ? "text-right" : "text-left"}`}>{column.label}</th>)}</tr></thead><tbody>{reportRows(report).map((row, index) => <tr key={index} className="border-t border-border hover:bg-accent">{report.columns.map((column) => <td key={column.key} className={`p-2 ${column.amount ? "whitespace-nowrap text-right tabular-nums" : "min-w-28"}`}>{column.amount ? formatAmount(row[column.key] ?? "0") : reportText(column.key, row[column.key], tr)}</td>)}</tr>)}{!reportRows(report).length && <tr><td className="p-5 text-center text-muted-foreground" colSpan={report.columns.length || 1}>{tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")}</td></tr>}</tbody></table></div>
    <p className="shrink-0 text-[0.9rem] text-muted-foreground">{tr("gl_data_as_of", "ข้อมูล ณ {0}").replace("{0}", String(report.asof || tr("gl_last_processing_time", "เวลาประมวลผลล่าสุด")))}</p>
  </div>;
}
const totalLabels: Record<string, GLLabel> = { debit: ["gl_total_debit", "รวมเดบิต"], credit: ["gl_total_credit", "รวมเครดิต"], balance: ["gl_balance", "ยอดคงเหลือ"], income: ["gl_revenue", "รายได้"], revenue: ["gl_revenue", "รายได้"], expense: ["gl_expense", "ค่าใช้จ่าย"], profit: ["gl_net_profit", "กำไรสุทธิ"], assets: ["gl_asset", "สินทรัพย์"], liabilities: ["gl_liability", "หนี้สิน"], equity: ["gl_equity", "ส่วนของเจ้าของ"], difference: ["gl_difference", "ผลต่าง"], budget: ["gl_budget_2", "งบประมาณ"], actual: ["gl_actual", "ยอดจริง"], opening: ["gl_opening_balance_2", "ยอดยกมา"], closing: ["gl_closing_balance", "ยอดยกไป"], cashin: ["gl_money_in", "เงินเข้า"], cashout: ["gl_money_out", "เงินออก"], netcash: ["gl_net_cash", "เงินสดสุทธิ"], cash: ["gl_cash_and_bank", "เงินสดและเงินฝากธนาคาร"], currentearnings: ["gl_unclosed_profit_loss", "กำไรขาดทุนที่ยังไม่ปิด"], unclassifiedlines: ["gl_unclassified_line", "บรรทัดที่ยังไม่ระบุประเภท"] };
export function GLReports({ name, heading }: { name: string; heading?: string }) {
  const tr = useGLText();
  const refs = useReferences();
  const [filters, setFilters] = useState<ReportFilters>({ ...emptyReportFilters }), [applied, setApplied] = useState<ReportFilters | null>(null);
  const [report, setReport] = useState<GLReport | null>(null), [page, setPage] = useState(1);
  const [busy, setBusy] = useState(false), [error, setError] = useState("");
  const set = (key: keyof ReportFilters, value: string) => setFilters((current) => ({ ...current, [key]: value }));
  async function load(nextPage = 1, selected = filters) {
    if (!selected.fiscalyear || busy) return;
    if (selected.from && selected.to && selected.from > selected.to) { setError(tr("gl_start_date_not_after_end", "วันเริ่มต้นต้องไม่เกินวันสิ้นสุด")); return; }
    setBusy(true); setError("");
    try { const result = await fetchReport(name, selected, nextPage, 50, nextPage > 1 && report && JSON.stringify(selected) === JSON.stringify(applied) ? report.sequence : undefined); setReport(result); setApplied({ ...selected }); setPage(nextPage); }
    catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  async function exportCsv() {
    if (!applied || !report || busy) return;
    setBusy(true); setError("");
    try {
      const complete = await fetchReport(name, applied, 1, 500);
      if (complete.totalrows > 100000) throw new Error(tr("gl_data_exceeds_100k_limit_date_range", "ข้อมูลเกิน 100,000 รายการ กรุณาจำกัดช่วงวันที่ก่อนส่งออก"));
      complete.rows = complete.rows ?? [];
      for (let next = 2; complete.rows.length < complete.totalrows; next++) {
        const chunk = await fetchReport(name, applied, next, 500, complete.sequence);
        if (chunk.totalrows !== complete.totalrows || chunk.sequence !== complete.sequence || !chunk.rows?.length) throw new Error(tr("gl_data_changed_during_export_retry", "ข้อมูลเปลี่ยนแปลงระหว่างส่งออก กรุณาลองใหม่"));
        complete.rows.push(...chunk.rows);
      }
      downloadText(tr("gl_accounts_csv", "บัญชี-{0}-{1}.csv").replace("{0}", String(name)).replace("{1}", String(applied.fiscalyear)), reportCsv(complete), "text/csv;charset=utf-8");
    } catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  return <section className={`${panel} flex flex-col flex-1 min-h-0 gap-3`}>
    {heading && <h2 className="shrink-0 text-lg font-semibold">{heading}</h2>}
    <Notice error text={error || refs.error} />
    <form className="shrink-0 grid gap-3" onSubmit={(event) => { event.preventDefault(); void load(); }}>
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <Field label={tr("gl_fiscal_year", "ปีบัญชี")}><YearSelect years={refs.years} value={filters.fiscalyear} onChange={(value) => { const year = refs.years.find((item) => item.code === value); setFilters((current) => ({ ...current, fiscalyear: value, from: year?.startdate ?? "", to: year?.enddate ?? "" })); }} /></Field>
        <Field label={tr("gl_from_date", "ตั้งแต่วันที่")}><input className={control} type="date" value={filters.from} onChange={(e) => set("from", e.target.value)} /></Field>
        <Field label={tr("gl_to_date", "ถึงวันที่")}><input className={control} type="date" value={filters.to} onChange={(e) => set("to", e.target.value)} /></Field>
        <Field label={tr("gl_account", "บัญชี")}><AccountSelect accounts={refs.accounts} all value={filters.accountcode} onChange={(value) => set("accountcode", value)} /></Field>
        <Field label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} value={filters.branchcode} onChange={(e) => set("branchcode", e.target.value)} placeholder={tr("gl_all_branches", "ทุกสาขา")} /></Field>
        <Field label={tr("gl_department_code", "รหัสแผนก")}><input className={control} value={filters.departmentcode} onChange={(e) => set("departmentcode", e.target.value)} placeholder={tr("gl_all_departments", "ทุกแผนก")} /></Field>
        <Field label={tr("gl_project_code", "รหัสโครงการ")}><input className={control} value={filters.projectcode} onChange={(e) => set("projectcode", e.target.value)} placeholder={tr("gl_all_projects", "ทุกโครงการ")} /></Field>
        <div className="flex flex-wrap items-end gap-2"><Button type="submit" className={actionClass} disabled={busy || !filters.fiscalyear}><RefreshCw />{busy ? tr("gl_processing", "กำลังประมวลผล…") : tr("gl_show_report", "แสดงรายงาน")}</Button><Button type="button" variant="outline" className={actionClass} disabled={busy || !report} onClick={() => void exportCsv()}><Download />{tr("gl_export_table", "ส่งออกตาราง")}</Button></div>
      </div>
    </form>
    {report ? <><ReportGrid report={report} graphs={["financialgraphs", "dashboard", "executivesummary"].includes(name)} /><div className="shrink-0"><Pager page={page} total={report.totalrows} onPage={(next) => void load(next, applied!)} loading={busy} limit={50} /></div></> : <div className="rounded-xl border border-dashed border-border p-6 text-center text-muted-foreground">{tr("gl_select_fy_and_conditions_show_report", "เลือกปีบัญชีและเงื่อนไข แล้วกดแสดงรายงาน")}</div>}
  </section>;
}
