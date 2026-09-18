"use client";

import { useEffect, useState, useMemo } from "react";
import { Download, RefreshCw, FileText, ArrowLeft, ExternalLink, X, CheckCircle2, AlertTriangle, Sparkles, ShieldCheck, TrendingUp, Calendar } from "lucide-react";
import { Button } from "@/components/ui/button";
import { accountTypeLabels, bookLabels, displayAmountUnits, labelText, type GLLabel, type GLTextFn, formatAmount, reportCsv, type GLReport, type GLJournal, journalTotals, amountString } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { AccountSelect, Field, Notice, Pager, YearSelect, actionClass, control, downloadText, panel, useReferences, useRowDensity, useGLText } from "./gl-common";
import { useReportPreferences } from "@/hooks/use-report-preferences";
import { ReportDisplayToolbar } from "@/components/report-display-toolbar";
import { GLHealthAuditModal } from "./gl-health-audit-modal";
import { ComparativeReportView, MonthlyTrendMatrixView } from "./gl-comparative-view";
import { buildComparativeReport, pivotAnnualBalances, type ComparativeReportResult } from "@/lib/gl-comparative-report";

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

export function JournalDrillDownModal({
  docno,
  open,
  onClose,
}: {
  docno: string | null;
  open: boolean;
  onClose: () => void;
}) {
  const tr = useGLText();
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

    glRequest<{ items?: GLJournal[] }>(`journals?q=${encodeURIComponent(docno)}&limit=10`)
      .then(async (res) => {
        if (!active) return;
        const match = res.items?.find((j) => j.docno.toLowerCase() === docno.toLowerCase()) ?? res.items?.[0];
        if (!match) {
          setError(tr("gl_err_not_found", "ไม่พบรายการบัญชี"));
          return;
        }
        if (match.id) {
          try {
            const full = await glRequest<GLJournal>(`journals/${encodeURIComponent(match.id)}`);
            if (active) setJournal(full);
          } catch {
            if (active) setJournal(match);
          }
        } else {
          setJournal(match);
        }
      })
      .catch((err) => {
        if (active) setError((err as Error).message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [open, docno, tr]);

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
                      ? "bg-emerald-500/10 text-emerald-600 border-emerald-500/20"
                      : journal.status === "reversed"
                        ? "bg-purple-500/10 text-purple-600 border-purple-500/20"
                        : "bg-amber-500/10 text-amber-600 border-amber-500/20"
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
          <Button variant="ghost" size="icon" className="size-8 rounded-lg" onClick={onClose}>
            <X className="size-4" />
          </Button>
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
                  <span className="inline-flex items-center gap-1.5 font-semibold text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 px-2.5 py-1 rounded-lg border border-emerald-500/20">
                    <CheckCircle2 className="size-4" /> {tr("gl_debit_equals_credit", "เดบิต = เครดิต สมดุล 100%")}
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1.5 font-semibold text-amber-600 dark:text-amber-400 bg-amber-500/10 px-2.5 py-1 rounded-lg border border-amber-500/20">
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

export function ReportGrid({
  report,
  graphs = false,
  onDrillDocNo,
  onDrillAccount,
}: {
  report: GLReport;
  graphs?: boolean;
  onDrillDocNo?: (docNo: string) => void;
  onDrillAccount?: (accountCode: string) => void;
}) {
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
    <div className={`flex-1 min-h-[300px] overflow-auto rounded-xl border border-border shadow-sm ${reportPref.contrastClass}`}>
      <table className={`w-full text-left leading-normal ${density.tableClass} ${reportPref.fontSizeClass}`}>
        <thead className="sticky top-0 bg-muted z-10">
          <tr>{report.columns.map((column) => <th key={column.key} className={`whitespace-nowrap p-2 ${column.amount ? "text-right" : "text-left"}`}>{column.label}</th>)}</tr>
        </thead>
        <tbody>
          {reportRows(report).map((row, index) => (
            <tr key={index} className="border-t border-border hover:bg-accent/60 transition-colors">
              {report.columns.map((column) => {
                const val = row[column.key] ?? "";
                const isDocNoCol = (column.key === "docno" || column.key === "documentno" || column.key === "reference") && val && val.trim() !== "";
                const isAccountCodeCol = column.key === "accountcode" && val && val !== "__current_earnings__";

                let cellContent;
                if (column.amount) {
                  cellContent = formatAmount(val);
                } else if (isDocNoCol && onDrillDocNo) {
                  cellContent = (
                    <button
                      type="button"
                      className="group inline-flex items-center gap-1 font-mono font-semibold text-primary hover:text-primary/80 hover:underline cursor-pointer"
                      onClick={(e) => { e.stopPropagation(); onDrillDocNo(val); }}
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
                  cellContent = reportText(column.key, val, tr);
                }

                return (
                  <td key={column.key} className={`p-2 ${column.amount ? "whitespace-nowrap text-right tabular-nums" : "min-w-28"}`}>
                    {cellContent}
                  </td>
                );
              })}
            </tr>
          ))}
          {!reportRows(report).length && <tr><td className="p-5 text-center text-muted-foreground" colSpan={report.columns.length || 1}>{tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")}</td></tr>}
        </tbody>
      </table>
    </div>
    <p className="shrink-0 text-[0.9rem] text-muted-foreground">{tr("gl_data_as_of", "ข้อมูล ณ {0}").replace("{0}", String(report.asof || tr("gl_last_processing_time", "เวลาประมวลผลล่าสุด")))}</p>
  </div>;
}
const totalLabels: Record<string, GLLabel> = { debit: ["gl_total_debit", "รวมเดบิต"], credit: ["gl_total_credit", "รวมเครดิต"], balance: ["gl_balance", "ยอดคงเหลือ"], income: ["gl_revenue", "รายได้"], revenue: ["gl_revenue", "รายได้"], expense: ["gl_expense", "ค่าใช้จ่าย"], profit: ["gl_net_profit", "กำไรสุทธิ"], assets: ["gl_asset", "สินทรัพย์"], liabilities: ["gl_liability", "หนี้สิน"], equity: ["gl_equity", "ส่วนของเจ้าของ"], difference: ["gl_difference", "ผลต่าง"], budget: ["gl_budget_2", "งบประมาณ"], actual: ["gl_actual", "ยอดจริง"], opening: ["gl_opening_balance_2", "ยอดยกมา"], closing: ["gl_closing_balance", "ยอดยกไป"], cashin: ["gl_money_in", "เงินเข้า"], cashout: ["gl_money_out", "เงินออก"], netcash: ["gl_net_cash", "เงินสดสุทธิ"], cash: ["gl_cash_and_bank", "เงินสดและเงินฝากธนาคาร"], currentearnings: ["gl_unclosed_profit_loss", "กำไรขาดทุนที่ยังไม่ปิด"], unclassifiedlines: ["gl_unclassified_line", "บรรทัดที่ยังไม่ระบุประเภท"] };
export function GLReports({ name, heading }: { name: string; heading?: string }) {
  const tr = useGLText();
  const refs = useReferences();
  const [activeReportName, setActiveReportName] = useState(name);
  const [drilledFrom, setDrilledFrom] = useState<{ report: string; accountcode: string } | null>(null);
  const [drillDocNo, setDrillDocNo] = useState<string | null>(null);

  const [filters, setFilters] = useState<ReportFilters>({ ...emptyReportFilters }), [applied, setApplied] = useState<ReportFilters | null>(null);
  const [report, setReport] = useState<GLReport | null>(null), [page, setPage] = useState(1);
  const [busy, setBusy] = useState(false), [error, setError] = useState("");
  const [healthAuditOpen, setHealthAuditOpen] = useState(false);
  const [comparativeMode, setComparativeMode] = useState(false);
  const [comparativeData, setComparativeData] = useState<ComparativeReportResult | null>(null);
  const [annualPivotMode, setAnnualPivotMode] = useState(false);
  const set = (key: keyof ReportFilters, value: string) => setFilters((current) => ({ ...current, [key]: value }));

  const monthlyPivot = useMemo(() => {
    if (activeReportName !== "annual-balances" || !report?.rows) return null;
    return pivotAnnualBalances(report.rows);
  }, [activeReportName, report?.rows]);

  async function handleToggleComparative() {
    if (comparativeMode) {
      setComparativeMode(false);
      setComparativeData(null);
      return;
    }

    if (!report || !applied?.fiscalyear) return;

    const currentIndex = refs.years.findIndex((y) => y.code === applied.fiscalyear);
    let priorYear: (typeof refs.years)[number] | undefined = refs.years[currentIndex + 1];
    if (!priorYear && refs.years.length > 1) {
      priorYear = refs.years.find((y) => y.code !== applied.fiscalyear);
    }

    if (!priorYear) {
      setError(tr("gl_no_prior_year_found", "ไม่พบปีบัญชีก่อนหน้าสำหรับเปรียบเทียบ"));
      return;
    }

    setBusy(true);
    setError("");
    try {
      const priorResult = await fetchReport(
        activeReportName,
        {
          ...applied,
          fiscalyear: priorYear.code,
          from: priorYear.startdate,
          to: priorYear.enddate,
        },
        1,
        500
      );

      const comp = buildComparativeReport({
        period1Label: `${tr("gl_fiscal_year", "ปี")} ${applied.fiscalyear}`,
        period1Rows: report.rows ?? [],
        period2Label: `${tr("gl_fiscal_year", "ปี")} ${priorYear.code}`,
        period2Rows: priorResult.rows ?? [],
        accounts: refs.accounts,
      });

      setComparativeData(comp);
      setComparativeMode(true);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  async function load(nextPage = 1, selected = filters, targetReport = activeReportName) {
    if (!selected.fiscalyear || busy) return;
    if (selected.from && selected.to && selected.from > selected.to) { setError(tr("gl_start_date_not_after_end", "วันเริ่มต้นต้องไม่เกินวันสิ้นสุด")); return; }
    setBusy(true); setError("");
    try {
      const result = await fetchReport(targetReport, selected, nextPage, 50, nextPage > 1 && report && JSON.stringify(selected) === JSON.stringify(applied) && targetReport === activeReportName ? report.sequence : undefined);
      setReport(result);
      setApplied({ ...selected });
      setPage(nextPage);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  const handleDrillAccount = (accountCode: string) => {
    if (!accountCode || accountCode === "__current_earnings__") return;
    setDrilledFrom({ report: activeReportName, accountcode: filters.accountcode });
    setActiveReportName("ledger");
    const updatedFilters = { ...filters, accountcode: accountCode };
    setFilters(updatedFilters);
    void load(1, updatedFilters, "ledger");
  };

  const handleBackFromDrill = () => {
    if (!drilledFrom) return;
    const prevReport = drilledFrom.report;
    const prevAccount = drilledFrom.accountcode;
    setDrilledFrom(null);
    setActiveReportName(prevReport);
    const updatedFilters = { ...filters, accountcode: prevAccount };
    setFilters(updatedFilters);
    void load(1, updatedFilters, prevReport);
  };

  async function exportCsv() {
    if (!applied || !report || busy) return;
    setBusy(true); setError("");
    try {
      const complete = await fetchReport(activeReportName, applied, 1, 500);
      if (complete.totalrows > 100000) throw new Error(tr("gl_data_exceeds_100k_limit_date_range", "ข้อมูลเกิน 100,000 รายการ กรุณาจำกัดช่วงวันที่ก่อนส่งออก"));
      complete.rows = complete.rows ?? [];
      for (let next = 2; complete.rows.length < complete.totalrows; next++) {
        const chunk = await fetchReport(activeReportName, applied, next, 500, complete.sequence);
        if (chunk.totalrows !== complete.totalrows || chunk.sequence !== complete.sequence || !chunk.rows?.length) throw new Error(tr("gl_data_changed_during_export_retry", "ข้อมูลเปลี่ยนแปลงระหว่างส่งออก กรุณาลองใหม่"));
        complete.rows.push(...chunk.rows);
      }
      downloadText(tr("gl_accounts_csv", "บัญชี-{0}-{1}.csv").replace("{0}", String(activeReportName)).replace("{1}", String(applied.fiscalyear)), reportCsv(complete), "text/csv;charset=utf-8");
    } catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }

  return <section className={`${panel} flex flex-col flex-1 min-h-0 gap-3`}>
    <div className="shrink-0 flex items-center justify-between gap-3">
      {heading && <h2 className="text-lg font-semibold">{heading}</h2>}
      {drilledFrom && (
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="bg-primary/10 text-primary border-primary/30 hover:bg-primary/20"
          onClick={handleBackFromDrill}
        >
          <ArrowLeft className="size-4 mr-1.5" />
          {tr("gl_back_to_report", "กลับไปรายงานก่อนหน้า")}
        </Button>
      )}
    </div>

    {drilledFrom && (
      <div className="shrink-0 flex items-center justify-between gap-3 p-3 rounded-xl bg-primary/10 border border-primary/20 text-primary">
        <div className="flex items-center gap-2 text-sm font-medium">
          <Sparkles className="size-4 shrink-0" />
          <span>
            {tr("gl_ledger_drill_account", "เจาะลึกแยกประเภทบัญชี: {0}").replace("{0}", filters.accountcode)}
          </span>
        </div>
      </div>
    )}

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
        <div className="flex flex-wrap items-end gap-2">
          <Button type="submit" className={actionClass} disabled={busy || !filters.fiscalyear}>
            <RefreshCw />
            {busy ? tr("gl_processing", "กำลังประมวลผล…") : tr("gl_show_report", "แสดงรายงาน")}
          </Button>
          <Button type="button" variant="outline" className={actionClass} disabled={busy || !report} onClick={() => void exportCsv()}>
            <Download />
            {tr("gl_export_table", "ส่งออกตาราง")}
          </Button>
          <Button
            type="button"
            variant="outline"
            className={`${actionClass} border-emerald-500/30 text-emerald-700 dark:text-emerald-400 bg-emerald-500/10 hover:bg-emerald-500/20 shadow-sm`}
            disabled={busy || !report}
            onClick={() => setHealthAuditOpen(true)}
            title={tr("gl_health_audit_btn_hint", "ตรวจสอบความผิดปกติและสมดุลผังบัญชี")}
          >
            <ShieldCheck className="size-4 mr-1 text-emerald-600 dark:text-emerald-400" />
            {tr("gl_audit_health", "ตรวจสุขภาพบัญชี")}
          </Button>
          {refs.years.length > 1 && (
            <Button
              type="button"
              variant="outline"
              className={`${actionClass} ${comparativeMode ? "bg-primary/15 text-primary border-primary font-semibold" : ""}`}
              disabled={busy || !report}
              onClick={() => void handleToggleComparative()}
              title={tr("gl_compare_prior_year_hint", "เปรียบเทียบตัวเลขกับปีก่อนหน้า")}
            >
              <TrendingUp className="size-4 mr-1 text-primary" />
              {comparativeMode ? tr("gl_normal_mode", "มุมมองปกติ") : tr("gl_compare_prior_year", "เปรียบเทียบปีก่อนหน้า")}
            </Button>
          )}
          {activeReportName === "annual-balances" && (
            <Button
              type="button"
              variant="outline"
              className={`${actionClass} ${annualPivotMode ? "bg-primary/15 text-primary border-primary font-semibold" : ""}`}
              disabled={busy || !report}
              onClick={() => setAnnualPivotMode(!annualPivotMode)}
              title={tr("gl_monthly_pivot_hint", "แสดงตารางเปรียบเทียบ 12 เดือน")}
            >
              <Calendar className="size-4 mr-1 text-primary" />
              {annualPivotMode ? tr("gl_list_mode", "ตารางปกติ") : tr("gl_monthly_pivot", "แนวโน้ม 12 เดือน")}
            </Button>
          )}
        </div>
      </div>
    </form>
    {report ? (
      <>
        {comparativeMode && comparativeData ? (
          <ComparativeReportView
            data={comparativeData}
            onDrillAccount={handleDrillAccount}
          />
        ) : annualPivotMode && monthlyPivot ? (
          <MonthlyTrendMatrixView
            matrix={monthlyPivot}
            onDrillAccount={handleDrillAccount}
          />
        ) : (
          <>
            <ReportGrid
              report={report}
              graphs={["financialgraphs", "dashboard", "executivesummary"].includes(activeReportName)}
              onDrillDocNo={(docno) => setDrillDocNo(docno)}
              onDrillAccount={handleDrillAccount}
            />
            <div className="shrink-0">
              <Pager page={page} total={report.totalrows} onPage={(next) => void load(next, applied!)} loading={busy} limit={50} />
            </div>
          </>
        )}
      </>
    ) : (
      <div className="rounded-xl border border-dashed border-border p-6 text-center text-muted-foreground">
        {tr("gl_select_fy_and_conditions_show_report", "เลือกปีบัญชีและเงื่อนไข แล้วกดแสดงรายงาน")}
      </div>
    )}

    {/* Drill Down Voucher Detail Modal */}
    <JournalDrillDownModal
      docno={drillDocNo}
      open={drillDocNo !== null}
      onClose={() => setDrillDocNo(null)}
    />

    {/* GL Health Audit Modal */}
    <GLHealthAuditModal
      open={healthAuditOpen}
      onClose={() => setHealthAuditOpen(false)}
      report={report}
      accounts={refs.accounts}
      onDrillAccount={handleDrillAccount}
    />
  </section>;
}
