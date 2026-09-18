"use client";

import { ExternalLink, TrendingUp, TrendingDown, Minus, Calendar } from "lucide-react";
import { formatAmount } from "@/lib/general-ledger";
import { type ComparativeReportResult, type MonthlyPivotResult } from "@/lib/gl-comparative-report";
import { useRowDensity, useGLText } from "./gl-common";
import { useReportPreferences } from "@/hooks/use-report-preferences";
import { ReportDisplayToolbar } from "@/components/report-display-toolbar";

export function ComparativeReportView({
  data,
  onDrillAccount,
}: {
  data: ComparativeReportResult;
  onDrillAccount?: (accountCode: string) => void;
}) {
  const tr = useGLText();
  const density = useRowDensity();
  const reportPref = useReportPreferences();

  return (
    <div className="flex flex-col flex-1 min-h-0 gap-3">
      {/* Comparative KPI Header */}
      <div className="shrink-0 grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div className="p-3 rounded-xl border border-border bg-muted/40 shadow-sm">
          <div className="text-xs text-muted-foreground">{data.period1Label}</div>
          <div className="text-lg font-mono font-bold tabular-nums text-foreground mt-0.5">
            {data.totals.period1Total}
          </div>
        </div>
        <div className="p-3 rounded-xl border border-border bg-muted/40 shadow-sm">
          <div className="text-xs text-muted-foreground">{data.period2Label}</div>
          <div className="text-lg font-mono font-bold tabular-nums text-foreground mt-0.5">
            {data.totals.period2Total}
          </div>
        </div>
        <div className="p-3 rounded-xl border border-border bg-muted/40 shadow-sm">
          <div className="text-xs text-muted-foreground">{tr("gl_variance_total", "ผลต่างรวม (Variance)")}</div>
          <div className="text-lg font-mono font-bold tabular-nums text-foreground mt-0.5">
            {data.totals.varianceTotal}
          </div>
        </div>
        <div className="p-3 rounded-xl border border-border bg-muted/40 shadow-sm">
          <div className="text-xs text-muted-foreground">{tr("gl_growth_percent", "อัตราการเปลี่ยนแปลง")}</div>
          <div className="flex items-center gap-1.5 mt-0.5">
            {data.totals.direction === "increase" ? (
              <TrendingUp className="size-4 text-emerald-500 shrink-0" />
            ) : data.totals.direction === "decrease" ? (
              <TrendingDown className="size-4 text-rose-500 shrink-0" />
            ) : (
              <Minus className="size-4 text-muted-foreground shrink-0" />
            )}
            <span
              className={`text-lg font-mono font-bold tabular-nums ${
                data.totals.direction === "increase"
                  ? "text-emerald-600 dark:text-emerald-400"
                  : data.totals.direction === "decrease"
                  ? "text-rose-600 dark:text-rose-400"
                  : "text-foreground"
              }`}
            >
              {data.totals.percentChangeTotal}
            </span>
          </div>
        </div>
      </div>

      {/* Display Toolbar */}
      <div className="shrink-0 flex justify-end">
        <ReportDisplayToolbar
          fontSize={reportPref.fontSize}
          onFontSizeChange={reportPref.setFontSize}
          highContrast={reportPref.highContrast}
          onToggleHighContrast={reportPref.toggleHighContrast}
          compact={density.compact}
          onToggleCompact={density.toggle}
          compactLabel={{
            compact: tr("gl_collapse_row", "ย่อบรรทัด"),
            expand: tr("gl_expand_row", "ขยายบรรทัด"),
          }}
        />
      </div>

      {/* Comparative Table */}
      <div className={`flex-1 min-h-[300px] overflow-auto rounded-xl border border-border shadow-sm ${reportPref.contrastClass}`}>
        <table className={`w-full text-left leading-normal ${density.tableClass} ${reportPref.fontSizeClass}`}>
          <thead className="sticky top-0 bg-muted z-10 text-xs uppercase font-semibold text-muted-foreground">
            <tr>
              <th className="p-2.5">{tr("gl_account_code", "รหัสบัญชี")}</th>
              <th className="p-2.5">{tr("gl_account_name", "ชื่อบัญชี")}</th>
              <th className="p-2.5 text-right">{data.period1Label}</th>
              <th className="p-2.5 text-right">{data.period2Label}</th>
              <th className="p-2.5 text-right">{tr("gl_variance", "ผลต่าง (Variance)")}</th>
              <th className="p-2.5 text-right">{tr("gl_percent_change", "% เปลี่ยนแปลง")}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border bg-card">
            {data.rows.map((row, idx) => {
              const isInc = row.direction === "increase";
              const isDec = row.direction === "decrease";

              return (
                <tr key={idx} className="hover:bg-accent/60 transition-colors">
                  <td className="p-2.5 font-mono text-xs font-bold text-primary whitespace-nowrap">
                    <div className="flex items-center justify-between gap-1 group">
                      <span>{row.accountcode}</span>
                      {onDrillAccount && (
                        <button
                          type="button"
                          className="opacity-0 group-hover:opacity-100 transition-opacity text-xs text-primary hover:text-primary/80 inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-primary/10 hover:bg-primary/20 cursor-pointer"
                          onClick={() => onDrillAccount(row.accountcode)}
                          title={tr("gl_drill_to_ledger", "ดูรายงานแยกประเภท")}
                        >
                          <ExternalLink className="size-3" />
                        </button>
                      )}
                    </div>
                  </td>
                  <td className="p-2.5 text-xs font-medium text-foreground">
                    {row.accountname}
                  </td>
                  <td className="p-2.5 text-right font-mono tabular-nums text-xs font-semibold text-foreground">
                    {row.period1Amount}
                  </td>
                  <td className="p-2.5 text-right font-mono tabular-nums text-xs text-muted-foreground">
                    {row.period2Amount}
                  </td>
                  <td className="p-2.5 text-right font-mono tabular-nums text-xs font-bold text-foreground">
                    {row.varianceAmount}
                  </td>
                  <td className="p-2.5 text-right whitespace-nowrap">
                    <span
                      className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-md font-mono text-xs font-bold ${
                        row.isNew
                          ? "bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20"
                          : isInc
                          ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20"
                          : isDec
                          ? "bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/20"
                          : "text-muted-foreground"
                      }`}
                    >
                      {isInc && !row.isNew && <TrendingUp className="size-3" />}
                      {isDec && <TrendingDown className="size-3" />}
                      {row.percentChange}
                    </span>
                  </td>
                </tr>
              );
            })}
            {data.rows.length === 0 && (
              <tr>
                <td colSpan={6} className="p-6 text-center text-muted-foreground">
                  {tr("gl_no_comparative_data", "ไม่พบข้อมูลสำหรับเปรียบเทียบในงวดที่เลือก")}
                </td>
              </tr>
            )}
          </tbody>
          <tfoot className="bg-muted/70 border-t-2 border-border font-bold text-foreground">
            <tr>
              <td colSpan={2} className="p-2.5 text-right text-xs">
                {tr("gl_total", "ยอดรวม")}
              </td>
              <td className="p-2.5 text-right font-mono tabular-nums text-xs">
                {data.totals.period1Total}
              </td>
              <td className="p-2.5 text-right font-mono tabular-nums text-xs text-muted-foreground">
                {data.totals.period2Total}
              </td>
              <td className="p-2.5 text-right font-mono tabular-nums text-xs font-bold">
                {data.totals.varianceTotal}
              </td>
              <td className="p-2.5 text-right font-mono tabular-nums text-xs font-bold">
                {data.totals.percentChangeTotal}
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  );
}

export function MonthlyTrendMatrixView({
  matrix,
  onDrillAccount,
}: {
  matrix: MonthlyPivotResult;
  onDrillAccount?: (accountCode: string) => void;
}) {
  const tr = useGLText();
  const density = useRowDensity();
  const reportPref = useReportPreferences();

  return (
    <div className="flex flex-col flex-1 min-h-0 gap-3">
      {/* Matrix Toolbar */}
      <div className="shrink-0 flex items-center justify-between">
        <div className="flex items-center gap-2 text-xs font-semibold text-muted-foreground">
          <Calendar className="size-4 text-primary" />
          <span>{tr("gl_monthly_trend_desc", "ตารางเปรียบเทียบแนวโน้มการเคลื่อนไหวรายเดือน 12 เดือน (Jan - Dec)")}</span>
        </div>
        <ReportDisplayToolbar
          fontSize={reportPref.fontSize}
          onFontSizeChange={reportPref.setFontSize}
          highContrast={reportPref.highContrast}
          onToggleHighContrast={reportPref.toggleHighContrast}
          compact={density.compact}
          onToggleCompact={density.toggle}
          compactLabel={{
            compact: tr("gl_collapse_row", "ย่อบรรทัด"),
            expand: tr("gl_expand_row", "ขยายบรรทัด"),
          }}
        />
      </div>

      {/* Matrix Table */}
      <div className={`flex-1 min-h-[300px] overflow-auto rounded-xl border border-border shadow-sm ${reportPref.contrastClass}`}>
        <table className={`w-full text-left leading-normal ${density.tableClass} ${reportPref.fontSizeClass}`}>
          <thead className="sticky top-0 bg-muted z-10 text-[11px] uppercase font-semibold text-muted-foreground">
            <tr>
              {matrix.columns.map((col) => (
                <th
                  key={col.key}
                  className={`p-2 whitespace-nowrap ${col.amount ? "text-right" : "text-left"}`}
                >
                  {col.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-border bg-card">
            {matrix.rows.map((row, idx) => (
              <tr key={idx} className="hover:bg-accent/60 transition-colors">
                <td className="p-2 font-mono text-xs font-bold text-primary whitespace-nowrap">
                  <div className="flex items-center justify-between gap-1 group">
                    <span>{row.accountcode}</span>
                    {onDrillAccount && (
                      <button
                        type="button"
                        className="opacity-0 group-hover:opacity-100 transition-opacity text-xs text-primary hover:text-primary/80 inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-primary/10 hover:bg-primary/20 cursor-pointer"
                        onClick={() => onDrillAccount(row.accountcode)}
                        title={tr("gl_drill_to_ledger", "ดูรายงานแยกประเภท")}
                      >
                        <ExternalLink className="size-3" />
                      </button>
                    )}
                  </div>
                </td>
                <td className="p-2 text-xs font-medium text-foreground whitespace-nowrap min-w-36">
                  {row.accountname}
                </td>
                {Array.from({ length: 12 }, (_, i) => `m${i + 1}`).map((mKey) => (
                  <td
                    key={mKey}
                    className="p-2 text-right font-mono tabular-nums text-xs text-foreground whitespace-nowrap"
                  >
                    {row[mKey] === "0.00" ? "—" : row[mKey]}
                  </td>
                ))}
                <td className="p-2 text-right font-mono tabular-nums text-xs font-bold text-primary whitespace-nowrap bg-primary/5">
                  {row.total}
                </td>
              </tr>
            ))}
            {matrix.rows.length === 0 && (
              <tr>
                <td colSpan={15} className="p-6 text-center text-muted-foreground">
                  {tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")}
                </td>
              </tr>
            )}
          </tbody>
          <tfoot className="bg-muted/70 border-t-2 border-border font-bold text-foreground">
            <tr>
              <td colSpan={2} className="p-2 text-right text-xs">
                {tr("gl_total", "ยอดรวม")}
              </td>
              {Array.from({ length: 12 }, (_, i) => `m${i + 1}`).map((mKey) => (
                <td key={mKey} className="p-2 text-right font-mono tabular-nums text-xs">
                  {matrix.monthTotals[mKey] || "0.00"}
                </td>
              ))}
              <td className="p-2 text-right font-mono tabular-nums text-xs font-bold text-primary bg-primary/10">
                {matrix.grandTotal}
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  );
}
