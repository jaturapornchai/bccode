"use client";

import { useState } from "react";
import { Download, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { accountTypes, books, displayAmountUnits, formatAmount, reportCsv, type GLReport } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { AccountSelect, Field, Notice, Pager, YearSelect, actionClass, control, downloadText, panel, useReferences, useRowDensity } from "./gl-common";

export type ReportFilters = { fiscalyear: string; from: string; to: string; accountcode: string; branchcode: string; departmentcode: string; projectcode: string; bookcode: string };
export const emptyReportFilters: ReportFilters = { fiscalyear: "", from: "", to: "", accountcode: "", branchcode: "", departmentcode: "", projectcode: "", bookcode: "" };
export async function fetchReport(name: string, filters: ReportFilters, page = 1, limit = 50, snapshot?: number) {
  return glRequest<GLReport>(`reports/${name}?${new URLSearchParams({ ...filters, page: String(page), limit: String(limit), ...(snapshot === undefined ? {} : { snapshot: String(snapshot) }) })}`);
}
function reportRows(report: GLReport) { return report.rows ?? []; }
const reportTextLabels: Record<string, Record<string, string>> = {
  accounttype: accountTypes,
  bookcode: books,
  status: { draft: "ฉบับร่าง", posted: "ผ่านรายการแล้ว", reversed: "กลับรายการแล้ว", void: "ยกเลิกร่าง" },
  direction: { in: "เงินเข้า", out: "เงินออก" },
  category: { operating: "ดำเนินงาน", investing: "ลงทุน", financing: "จัดหาเงิน", unclassified: "ยังไม่ระบุ" },
};
function reportText(key: string, value = "") {
  // The computed earnings row is not a chart-of-accounts entry. Keep its source key for exports.
  if (key === "accountcode" && value === "__current_earnings__") return "—";
  return reportTextLabels[key]?.[value] ?? value;
}
export function ReportGrid({ report, graphs = false }: { report: GLReport; graphs?: boolean }) {
  const density = useRowDensity();
  const amountColumns = report.columns.filter((column) => column.amount);
  const amounts = reportRows(report).map((row) => { try { return displayAmountUnits(row[amountColumns[0]?.key] ?? "0"); } catch { return 0n; } });
  const max = amounts.reduce((largest, amount) => (amount < 0n ? -amount : amount) > largest ? (amount < 0n ? -amount : amount) : largest, 0n);
  return <div className="flex flex-col flex-1 min-h-0 gap-3">
    <div className="shrink-0 flex flex-col gap-2">{(report.warnings ?? []).map((warning, index) => <Notice key={index} text={warning} />)}</div>
    {!!Object.keys(report.totals ?? {}).length && <div className="shrink-0 flex flex-wrap gap-2">{Object.entries(report.totals).map(([key, value]) => <div key={key} className="min-w-36 flex-1 rounded-xl border border-border bg-muted/40 p-3"><div className="text-[0.9rem] text-muted-foreground">{report.columns.find((column) => column.key === key)?.label ?? totalLabels[key] ?? "ยอดรวม"}</div><div className="text-lg font-semibold tabular-nums">{key === "unclassifiedlines" ? `${value} รายการ` : formatAmount(value)}</div></div>)}</div>}
    {graphs && max > 0n && <div className="shrink-0 grid gap-2 rounded-xl border border-border p-3" aria-label="กราฟเปรียบเทียบยอดบัญชี">{reportRows(report).slice(0, 20).map((row, index) => <div key={index} className="grid gap-1"><div className="flex flex-wrap justify-between gap-2 text-[0.95rem]"><span>{row.name ?? row.accountname ?? row.month ?? row[report.columns.find((column) => !column.amount)?.key ?? ""] ?? `รายการ ${index + 1}`}</span><strong>{formatAmount(row[amountColumns[0]?.key] ?? "0")}</strong></div><div className="h-3 rounded-full bg-muted"><div className="h-3 rounded-full bg-primary" style={{ width: `${((amounts[index] < 0n ? -amounts[index] : amounts[index]) * 100n / max).toString()}%` }} /></div></div>)}</div>}
    <div className="shrink-0 flex justify-end"><Button type="button" variant="outline" className={actionClass} aria-pressed={density.compact} onClick={density.toggle}>{density.compact ? "ขยายบรรทัด" : "ย่อบรรทัด"}</Button></div>
    <div className="flex-1 min-h-[300px] overflow-auto rounded-xl border border-border"><table className={`w-full text-left text-[0.95rem] leading-normal ${density.tableClass}`}><thead className="sticky top-0 bg-muted"><tr>{report.columns.map((column) => <th key={column.key} className={`whitespace-nowrap p-2 ${column.amount ? "text-right" : "text-left"}`}>{column.label}</th>)}</tr></thead><tbody>{reportRows(report).map((row, index) => <tr key={index} className="border-t border-border hover:bg-accent">{report.columns.map((column) => <td key={column.key} className={`p-2 ${column.amount ? "whitespace-nowrap text-right tabular-nums" : "min-w-28"}`}>{column.amount ? formatAmount(row[column.key] ?? "0") : reportText(column.key, row[column.key])}</td>)}</tr>)}{!reportRows(report).length && <tr><td className="p-5 text-center text-muted-foreground" colSpan={report.columns.length || 1}>ไม่พบรายการตามเงื่อนไขที่เลือก</td></tr>}</tbody></table></div>
    <p className="shrink-0 text-[0.9rem] text-muted-foreground">ข้อมูล ณ {report.asof || "เวลาประมวลผลล่าสุด"}</p>
  </div>;
}
const totalLabels: Record<string, string> = { debit: "เดบิตรวม", credit: "เครดิตรวม", balance: "ยอดคงเหลือ", income: "รายได้", revenue: "รายได้", expense: "ค่าใช้จ่าย", profit: "กำไรสุทธิ", assets: "สินทรัพย์", liabilities: "หนี้สิน", equity: "ส่วนของเจ้าของ", difference: "ผลต่าง", budget: "งบประมาณ", actual: "ยอดจริง", opening: "ยอดยกมา", closing: "ยอดยกไป", cashin: "เงินเข้า", cashout: "เงินออก", netcash: "เงินสดสุทธิ", cash: "เงินสดและเงินฝากธนาคาร", currentearnings: "กำไรขาดทุนที่ยังไม่ปิด", unclassifiedlines: "บรรทัดที่ยังไม่ระบุประเภท" };
export function GLReports({ name, heading }: { name: string; heading?: string }) {
  const refs = useReferences();
  const [filters, setFilters] = useState<ReportFilters>({ ...emptyReportFilters }), [applied, setApplied] = useState<ReportFilters | null>(null);
  const [report, setReport] = useState<GLReport | null>(null), [page, setPage] = useState(1);
  const [busy, setBusy] = useState(false), [error, setError] = useState("");
  const set = (key: keyof ReportFilters, value: string) => setFilters((current) => ({ ...current, [key]: value }));
  async function load(nextPage = 1, selected = filters) {
    if (!selected.fiscalyear || busy) return;
    if (selected.from && selected.to && selected.from > selected.to) { setError("วันเริ่มต้นต้องไม่เกินวันสิ้นสุด"); return; }
    setBusy(true); setError("");
    try { const result = await fetchReport(name, selected, nextPage, 50, nextPage > 1 && report && JSON.stringify(selected) === JSON.stringify(applied) ? report.sequence : undefined); setReport(result); setApplied({ ...selected }); setPage(nextPage); }
    catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  async function exportCsv() {
    if (!applied || !report || busy) return;
    setBusy(true); setError("");
    try {
      const complete = await fetchReport(name, applied, 1, 500);
      if (complete.totalrows > 100000) throw new Error("ข้อมูลเกิน 100,000 รายการ กรุณาจำกัดช่วงวันที่ก่อนส่งออก");
      complete.rows = complete.rows ?? [];
      for (let next = 2; complete.rows.length < complete.totalrows; next++) {
        const chunk = await fetchReport(name, applied, next, 500, complete.sequence);
        if (chunk.totalrows !== complete.totalrows || chunk.sequence !== complete.sequence || !chunk.rows?.length) throw new Error("ข้อมูลเปลี่ยนแปลงระหว่างส่งออก กรุณาลองใหม่");
        complete.rows.push(...chunk.rows);
      }
      downloadText(`บัญชี-${name}-${applied.fiscalyear}.csv`, reportCsv(complete), "text/csv;charset=utf-8");
    } catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  return <section className={`${panel} flex flex-col flex-1 min-h-0 gap-3`}>
    {heading && <h2 className="shrink-0 text-lg font-semibold">{heading}</h2>}
    <Notice error text={error || refs.error} />
    <form className="shrink-0 grid gap-3" onSubmit={(event) => { event.preventDefault(); void load(); }}>
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <Field label="ปีบัญชี"><YearSelect years={refs.years} value={filters.fiscalyear} onChange={(value) => { const year = refs.years.find((item) => item.code === value); setFilters((current) => ({ ...current, fiscalyear: value, from: year?.startdate ?? "", to: year?.enddate ?? "" })); }} /></Field>
        <Field label="ตั้งแต่วันที่"><input className={control} type="date" value={filters.from} onChange={(e) => set("from", e.target.value)} /></Field>
        <Field label="ถึงวันที่"><input className={control} type="date" value={filters.to} onChange={(e) => set("to", e.target.value)} /></Field>
        <Field label="บัญชี"><AccountSelect accounts={refs.accounts} all value={filters.accountcode} onChange={(value) => set("accountcode", value)} /></Field>
        <Field label="รหัสสาขา"><input className={control} value={filters.branchcode} onChange={(e) => set("branchcode", e.target.value)} placeholder="ทุกสาขา" /></Field>
        <Field label="รหัสแผนก"><input className={control} value={filters.departmentcode} onChange={(e) => set("departmentcode", e.target.value)} placeholder="ทุกแผนก" /></Field>
        <Field label="รหัสโครงการ"><input className={control} value={filters.projectcode} onChange={(e) => set("projectcode", e.target.value)} placeholder="ทุกโครงการ" /></Field>
        <div className="flex flex-wrap items-end gap-2"><Button type="submit" className={actionClass} disabled={busy || !filters.fiscalyear}><RefreshCw />{busy ? "กำลังประมวลผล…" : "แสดงรายงาน"}</Button><Button type="button" variant="outline" className={actionClass} disabled={busy || !report} onClick={() => void exportCsv()}><Download />ส่งออกตาราง</Button></div>
      </div>
    </form>
    {report ? <><ReportGrid report={report} graphs={["financialgraphs", "dashboard", "executivesummary"].includes(name)} /><div className="shrink-0"><Pager page={page} total={report.totalrows} onPage={(next) => void load(next, applied!)} loading={busy} limit={50} /></div></> : <div className="rounded-xl border border-dashed border-border p-6 text-center text-muted-foreground">เลือกปีบัญชีและเงื่อนไข แล้วกดแสดงรายงาน</div>}
  </section>;
}
