"use client";

import { formatAppDate } from "@/lib/date-time";
import { formatAmount, type GLJournalBook, type GLReport, type GLTextFn } from "@/lib/general-ledger";
import type { LanguageCode } from "@/lib/i18n";
import { countTotalKeys, printPages, toGLReport, type ReportCondition, type ReportRow, type ReportRun, type ReportView } from "@/lib/report-registry";
import { ReportHeadRows, accountIndent, balanceText, captionSpan, controlRowText, isSummaryRow, repeatsAccount, reportText } from "@/app/gl/gl-reports";
import { ReportPrintConditions, ReportPrintHeading, printAmountCell, printCompanyName, printUserName, printedAtText, visiblePrintColumns } from "@/app/gl/gl-print";

type PrintColumn = GLReport["columns"][number];
const ISO_DATE = /^\d{4}-\d{2}-\d{2}$/;

/** header.printed_at is the server's timestamp; an unreadable value falls back to the browser clock. */
export function printedAtFromHeader(value: string, language: LanguageCode): string {
  const parsed = new Date(value);
  return value && !Number.isNaN(parsed.getTime()) ? printedAtText(language, parsed) : printedAtText(language);
}

function conditionLine(condition: ReportCondition, language: LanguageCode): string {
  const value = ISO_DATE.test(condition.value) ? formatAppDate(condition.value, language) : condition.value;
  return `${condition.label} ${value}`.trim();
}

/** The date range goes to the heading's "details" line (Champ layout); every other condition is listed after the report. */
export function splitConditions(conditions: ReportCondition[], language: LanguageCode): { description: string[]; conditions: string[] } {
  const isDate = (condition: ReportCondition) => condition.type === "date" || ISO_DATE.test(condition.value);
  return {
    description: conditions.filter(isDate).map((condition) => conditionLine(condition, language)),
    conditions: conditions.filter((condition) => !isDate(condition) && condition.value.trim() !== "").map((condition) => conditionLine(condition, language)),
  };
}

function printTotalText(run: ReportRun, counts: string[], key: string, value: string): string {
  return formatAmount(value, counts.includes(key) ? 0 : run.columns.find((column) => column.key === key)?.scale);
}

type Props = { run: ReportRun; report: ReportView; books: GLJournalBook[]; tr: GLTextFn; language: LanguageCode };

/**
 * Print layout of a registry run: one table per printed page (server page_totals), each with the repeated report heading,
 * the page total and the carried-forward total; the last page closes with the grand totals. Every table after the first
 * starts a new sheet (`gl-print-pagebreak`).
 */
export function ReportCenterPrint({ run, report, books, tr, language }: Props) {
  const glReport = toGLReport(run);
  const columns = visiblePrintColumns(glReport.columns, run.rows);
  const pages = printPages(run);
  const counts = countTotalKeys(run);
  const header = run.header;
  const split = splitConditions(header.conditions, language);
  const description = [header.branch_name, ...split.description].filter(Boolean);
  const company = header.company_name || printCompanyName();
  const title = header.title || report.name;
  const printedBy = header.printed_by || printUserName();
  const printedAt = printedAtFromHeader(header.printed_at, language);
  const inline = Object.entries(run.totals).filter(([key]) => columns.some((column) => column.amount && column.key === key));
  const extra = Object.entries(run.totals).filter(([key]) => !inline.some(([inlineKey]) => inlineKey === key));
  const footerRows = run.footer.filter((row) => row.label);
  const dateKeys = new Set(run.columns.filter((column) => column.type === "date").map((column) => column.key));

  const cell = (row: ReportRow, column: PrintColumn, index: number) => {
    if (column.amount) return balanceText(column.key, printAmountCell(column.key, row[column.key] ?? "", column.scale));
    if (repeatsAccount(row, column.key, index === 0)) return "";
    const value = controlRowText(row, column.key, row[column.key] ?? "", tr);
    return dateKeys.has(column.key) && ISO_DATE.test(value) ? formatAppDate(value, language) : reportText(column.key, value, tr, books, language);
  };
  const rowCells = (row: ReportRow, index: number) => {
    const caption = captionSpan(row._label, row, columns);
    return columns.map((column, columnIndex) => {
      if (caption && columnIndex > 0 && columnIndex < caption) return null;
      if (caption && columnIndex === 0) return <td key={column.key} colSpan={caption}>{row._label}</td>;
      return <td key={column.key} style={accountIndent(row, column.key)} className={column.amount ? "gl-print-num" : ""}>{cell(row, column, index)}</td>;
    });
  };
  /** A row of sums under the amount columns that have a value in `values` (page total, brought / carried forward). */
  const sumRow = (label: string, values: Record<string, string>, key: string) => (
    <tr key={key} className="gl-print-total">
      {columns.map((column, index) => <td key={column.key} className={column.amount ? "gl-print-num" : ""}>{column.amount ? (values[column.key] !== undefined ? formatAmount(values[column.key], column.scale) : "") : index === 0 ? label : ""}</td>)}
    </tr>
  );

  return <>
    {pages.map((page, pageIndex) => {
      const last = pageIndex === pages.length - 1;
      const paged = pages.length > 1 && Object.keys(page.sums).length > 0;
      return <table key={pageIndex} className={pageIndex > 0 ? "gl-print-pagebreak" : undefined}>
        <thead>
          <tr className="gl-print-head"><th colSpan={columns.length || 1}><ReportPrintHeading company={company} title={title} printedBy={printedBy} printedAt={printedAt} description={description} tr={tr} /></th></tr>
          <ReportHeadRows columns={columns} cellClass={(column) => column.amount ? "gl-print-num" : ""} />
        </thead>
        <tbody>
          {paged && pageIndex > 0 && sumRow(tr("gl_opening_balance_2", "ยอดยกมา"), page.brought, "brought")}
          {page.rows.map((row, index) => <tr key={index} className={isSummaryRow(row) ? "gl-print-total" : undefined}>{rowCells(row, index)}</tr>)}
          {!run.rows.length && <tr><td colSpan={columns.length || 1}>{tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")}</td></tr>}
          {paged && sumRow(tr("report_center_page_total", "รวมหน้านี้"), page.sums, "sums")}
          {paged && !last && sumRow(tr("gl_closing_balance", "ยอดยกไป"), page.carried, "carried")}
          {last && inline.length > 0 && <tr className="gl-print-total">{columns.map((column, index) => { const total = inline.find(([key]) => key === column.key); return <td key={column.key} className={column.amount ? "gl-print-num" : ""}>{total ? formatAmount(total[1], column.scale) : index === 0 ? tr("gl_report_grand_total", "รวมทั้งสิ้น") : ""}</td>; })}</tr>}
          {last && footerRows.map((row, rowIndex) => {
            const caption = captionSpan(row.label, row, columns);
            return <tr key={`footer-${rowIndex}`} className="gl-print-total">{columns.map((column, columnIndex) => {
              if (columnIndex > 0 && columnIndex < caption) return null;
              if (columnIndex === 0) return <td key={column.key} colSpan={caption}>{row.label}</td>;
              return <td key={column.key} className={column.amount ? "gl-print-num" : ""}>{column.amount ? formatAmount(row[column.key] ?? "", column.scale) : row[column.key] ?? ""}</td>;
            })}</tr>;
          })}
        </tbody>
      </table>;
    })}
    {extra.length > 0 && <dl className="gl-print-totals">{extra.map(([key, value]) => <div key={key}><dt>{run.totals_labels[key] ?? key}</dt><dd>{printTotalText(run, counts, key, value)}</dd></div>)}</dl>}
    <ReportPrintConditions lines={split.conditions} tr={tr} />
  </>;
}
