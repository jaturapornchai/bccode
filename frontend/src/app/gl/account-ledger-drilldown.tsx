"use client";

import { useEffect, useState, useMemo } from "react";
import { BookOpen, ExternalLink, RefreshCw, X, FileText, ArrowRight, Printer, Download, Eye } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Combobox } from "@/components/ui/select";
import { formatAmount, type GLAccount, type GLFiscalYear, type GLJournalBook, type GLReport, journalBookName, findJournalBook } from "@/lib/general-ledger";
import { formatAppDate } from "@/lib/date-time";
import { actionClass, control, downloadText, useGLLanguage, useGLText } from "./gl-common";
import { fetchReport, JournalDrillDownModal, emptyReportFilters, type ReportFilters } from "./gl-reports";

export function AccountBalanceCard({
  account,
  years,
  books = [],
  onOpenLedger,
}: {
  account: GLAccount;
  years: GLFiscalYear[];
  books?: GLJournalBook[];
  onOpenLedger: (fiscalYear: string) => void;
}) {
  const tr = useGLText();
  const activeYear = useMemo(() => {
    return years.find((y) => y.isactive && !y.closed)?.code ?? years[0]?.code ?? "";
  }, [years]);

  const [selectedYear, setSelectedYear] = useState(activeYear);
  const [balance, setBalance] = useState<{
    opening: string;
    debit: string;
    credit: string;
    balance: string;
    totalRows: number;
  } | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!selectedYear && activeYear) {
      setSelectedYear(activeYear);
    }
  }, [activeYear, selectedYear]);

  useEffect(() => {
    if (!account?.accountcode || !selectedYear) return;
    let active = true;
    setLoading(true);

    fetchReport(
      "ledger",
      { ...emptyReportFilters, fiscalyear: selectedYear, accountcode: account.accountcode },
      1,
      1
    )
      .then((res) => {
        if (!active) return;
        setBalance({
          opening: res.totals?.opening ?? "0.00",
          debit: res.totals?.debit ?? "0.00",
          credit: res.totals?.credit ?? "0.00",
          balance: res.totals?.balance ?? "0.00",
          totalRows: res.totalrows ?? 0,
        });
      })
      .catch(() => {
        if (active) {
          setBalance({
            opening: "0.00",
            debit: "0.00",
            credit: "0.00",
            balance: "0.00",
            totalRows: 0,
          });
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [account?.accountcode, selectedYear]);

  const netNum = balance ? parseFloat(balance.balance) : 0;
  const isDebitBalance = netNum > 0;
  const isCreditBalance = netNum < 0;

  return (
    <div className="rounded-xl border border-primary/20 bg-primary/5 p-2 flex flex-col gap-1.5 shrink-0">
      <div className="flex flex-wrap items-center justify-between gap-1.5 border-b border-primary/10 pb-1.5">
        <div className="flex items-center gap-1.5">
          <BookOpen className="size-3.5 text-primary" />
          <span className="text-xs font-bold text-foreground">
            {tr("gl_account_balance_summary", "สรุปยอดยกมาและการเคลื่อนไหว (คำนวณสดจากข้อมูลรายวัน)")}
          </span>
        </div>
        <div className="flex items-center gap-1.5">
          {years.length > 1 && (
            <div className="flex items-center gap-1 text-xs">
              <span className="text-muted-foreground">{tr("gl_fiscal_year", "ปีบัญชี")}:</span>
              <Combobox
                className="h-6 text-xs py-0 px-2 min-w-20"
                value={selectedYear}
                onChange={setSelectedYear}
              >
                {years.map((y) => (
                  <option key={y.code} value={y.code}>
                    {y.code} {y.closed ? `(${tr("gl_closed", "ปิดแล้ว")})` : ""}
                  </option>
                ))}
              </Combobox>
            </div>
          )}
          <Button
            type="button"
            variant="outline"
            className="h-6 px-2 text-xs font-medium text-primary hover:bg-primary/15 border-primary/30"
            onClick={() => onOpenLedger(selectedYear)}
            disabled={loading}
            title={tr("gl_drill_to_ledger", "เปิดดูสมุดบัญชีแยกประเภทของบัญชีนี้")}
          >
            <Eye className="size-3 mr-1" />
            {tr("gl_drill_to_ledger", "ดูสมุดแยกประเภท")}
            {balance && balance.totalRows > 0 && (
              <span className="ml-1 px-1 py-0 rounded-full bg-primary/20 text-[10px] font-bold">
                {balance.totalRows.toLocaleString("th-TH")}
              </span>
            )}
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-1.5">
        {/* ยอดยกมา */}
        <div className="rounded-lg bg-background/80 border border-border/60 px-2.5 py-1 flex items-center justify-between min-w-0">
          <span className="text-[10.5px] font-medium text-muted-foreground truncate mr-1">
            {tr("gl_opening_balance_2", "ยอดยกมา")}
          </span>
          <span className="text-xs font-mono font-bold text-foreground truncate">
            {loading ? "…" : formatAmount(balance?.opening ?? "0.00")}
          </span>
        </div>

        {/* รวมเดบิต */}
        <div className="rounded-lg bg-background/80 border border-border/60 px-2.5 py-1 flex items-center justify-between min-w-0">
          <span className="text-[10.5px] font-medium text-muted-foreground truncate mr-1">
            {tr("gl_total_debit", "รวมเดบิต")}
          </span>
          <span className="text-xs font-mono font-bold text-primary truncate">
            {loading ? "…" : formatAmount(balance?.debit ?? "0.00")}
          </span>
        </div>

        {/* รวมเครดิต */}
        <div className="rounded-lg bg-background/80 border border-border/60 px-2.5 py-1 flex items-center justify-between min-w-0">
          <span className="text-[10.5px] font-medium text-muted-foreground truncate mr-1">
            {tr("gl_total_credit", "รวมเครดิต")}
          </span>
          <span className="text-xs font-mono font-bold text-primary truncate">
            {loading ? "…" : formatAmount(balance?.credit ?? "0.00")}
          </span>
        </div>

        {/* ยอดคงเหลือสุทธิ */}
        <div className="rounded-lg bg-background/80 border border-primary/30 px-2.5 py-1 flex items-center justify-between min-w-0">
          <div className="flex items-center gap-1 min-w-0">
            <span className="text-[10.5px] font-bold text-foreground truncate">
              {tr("gl_balance", "คงเหลือ")}
            </span>
            {!loading && balance && netNum !== 0 && (
              <span
                className={`text-[9px] font-extrabold px-1 rounded ${
                  isDebitBalance
                    ? "bg-blue-500/10 text-blue-600 dark:text-blue-400"
                    : "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                }`}
              >
                {isDebitBalance ? "Dr." : "Cr."}
              </span>
            )}
          </div>
          <span className="text-xs font-mono font-extrabold text-foreground truncate">
            {loading ? "…" : formatAmount(balance ? Math.abs(netNum).toFixed(2) : "0.00")}
          </span>
        </div>
      </div>
    </div>
  );
}

export function AccountLedgerModal({
  open,
  onClose,
  account,
  fiscalYear,
  books = [],
}: {
  open: boolean;
  onClose: () => void;
  account: GLAccount | null;
  fiscalYear: string;
  books?: GLJournalBook[];
}) {
  const tr = useGLText();
  const language = useGLLanguage();
  const [report, setReport] = useState<GLReport | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [drillDoc, setDrillDoc] = useState<{ docno: string; journalId: string } | null>(null);

  useEffect(() => {
    if (!open || !account?.accountcode || !fiscalYear) {
      setReport(null);
      setError("");
      return;
    }
    let active = true;
    setLoading(true);
    setError("");

    fetchReport(
      "ledger",
      { ...emptyReportFilters, fiscalyear: fiscalYear, accountcode: account.accountcode },
      1,
      200
    )
      .then((res) => {
        if (active) setReport(res);
      })
      .catch((err) => {
        if (active) setError((err as Error).message || tr("gl_load_error", "โหลดข้อมูลไม่สำเร็จ"));
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [open, account?.accountcode, fiscalYear, tr]);

  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, onClose]);

  if (!open || !account) return null;

  const rows = report?.rows ?? [];
  const thName = (account.names || []).find((n) => n.code === "th")?.name ?? account.accountcode;

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-200">
      <div className="max-w-5xl w-full max-h-[90vh] bg-card border border-border rounded-2xl shadow-2xl flex flex-col overflow-hidden text-card-foreground">
        {/* Header */}
        <div className="p-4 border-b border-border bg-muted/30 flex items-center justify-between gap-3 shrink-0">
          <div className="flex items-center gap-3 min-w-0">
            <div className="p-2 rounded-xl bg-primary/10 text-primary shrink-0">
              <BookOpen className="size-5" />
            </div>
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <h3 className="text-base font-bold text-foreground">
                  {tr("gl_ledger_drill_account", "สมุดบัญชีแยกประเภท: {0}").replace("{0}", account.accountcode)}
                </h3>
                <span className="font-mono text-xs px-2 py-0.5 rounded-md bg-primary/10 border border-primary/20 font-bold text-primary">
                  {account.accountcode}
                </span>
                <span className="text-xs px-2 py-0.5 rounded-md bg-muted border border-border font-medium text-muted-foreground">
                  {tr("gl_fiscal_year", "ปีบัญชี")} {fiscalYear}
                </span>
              </div>
              <p className="text-xs text-muted-foreground truncate">{thName}</p>
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-8 rounded-lg text-muted-foreground hover:text-foreground"
              onClick={onClose}
              aria-label={tr("gl_close", "ปิด")}
              title={tr("gl_close", "ปิด")}
            >
              <X className="size-4" />
            </Button>
          </div>
        </div>

        {/* Toolbar & Summary */}
        <div className="px-4 py-2.5 bg-muted/20 border-b border-border/70 flex flex-wrap items-center justify-between gap-2 shrink-0">
          <div className="flex flex-wrap items-center gap-4 text-xs">
            <span>
              {tr("gl_opening_balance_2", "ยอดยกมา")}:{" "}
              <strong className="font-mono font-bold">{formatAmount(report?.totals?.opening ?? "0.00")}</strong>
            </span>
            <span className="text-border">|</span>
            <span>
              {tr("gl_total_debit", "รวมเดบิต")}:{" "}
              <strong className="font-mono font-bold text-primary">{formatAmount(report?.totals?.debit ?? "0.00")}</strong>
            </span>
            <span className="text-border">|</span>
            <span>
              {tr("gl_total_credit", "รวมเครดิต")}:{" "}
              <strong className="font-mono font-bold text-primary">{formatAmount(report?.totals?.credit ?? "0.00")}</strong>
            </span>
            <span className="text-border">|</span>
            <span>
              {tr("gl_balance", "ยอดคงเหลือ")}:{" "}
              <strong className="font-mono font-bold">{formatAmount(report?.totals?.balance ?? "0.00")}</strong>
            </span>
          </div>

          <span className="text-xs text-muted-foreground font-medium">
            {tr("gl_all_records", "รายการทั้งหมด")}: {rows.length.toLocaleString("th-TH")}{" "}
            {tr("gl_items", "รายการ")}
          </span>
        </div>

        {/* Table Content */}
        <div className="flex-1 min-h-[350px] overflow-auto scrollbar-thin p-1">
          {loading ? (
            <div className="flex flex-col items-center justify-center p-12 text-muted-foreground gap-2">
              <RefreshCw className="size-6 animate-spin text-primary" />
              <span className="text-sm font-medium">{tr("gl_loading_data", "กำลังโหลดข้อมูล…")}</span>
            </div>
          ) : error ? (
            <div className="p-8 text-center text-sm text-destructive">{error}</div>
          ) : rows.length === 0 ? (
            <div className="flex flex-col items-center justify-center p-12 text-muted-foreground gap-2">
              <FileText className="size-8 text-muted-foreground/60" />
              <span className="text-sm font-medium">{tr("gl_no_records_found_criteria", "ไม่มีรายการเคลื่อนไหวในรอบระยะเวลานี้")}</span>
            </div>
          ) : (
            <table className="w-full text-left border-collapse text-xs">
              <thead className="sticky top-0 bg-card border-b border-border text-[0.7rem] font-bold text-muted-foreground uppercase z-10">
                <tr>
                  <th className="py-2 px-2.5 w-24">{tr("gl_date", "วันที่")}</th>
                  <th className="py-2 px-2 w-32">{tr("gl_document_no", "เลขที่เอกสาร")}</th>
                  <th className="py-2 px-2 w-20">{tr("gl_journal", "สมุดรายวัน")}</th>
                  <th className="py-2 px-2 min-w-44">{tr("gl_description", "คำอธิบาย")}</th>
                  <th className="py-2 px-2 w-16 text-center">{tr("gl_branch_code", "สาขา")}</th>
                  <th className="py-2 px-2 w-16 text-center">{tr("gl_department", "แผนก")}</th>
                  <th className="py-2 px-2 w-16 text-center">{tr("gl_project", "โครงการ")}</th>
                  <th className="py-2 px-2 w-28 text-right">{tr("gl_debit", "เดบิต")}</th>
                  <th className="py-2 px-2 w-28 text-right">{tr("gl_credit", "เครดิต")}</th>
                  <th className="py-2 pr-2.5 pl-2 w-28 text-right">{tr("gl_balance", "ยอดคงเหลือ")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40 font-mono text-[0.75rem]">
                {rows.map((row, index) => {
                  const bCode = row.bookcode || "";
                  const bName = journalBookName(findJournalBook(books, bCode), bCode, language);
                  return (
                    <tr
                      key={row.journalid ? `${row.journalid}-${index}` : index}
                      className="hover:bg-muted/40 transition-colors"
                    >
                      <td className="py-1.5 px-2.5 whitespace-nowrap text-muted-foreground font-sans">
                        {row.date ? formatAppDate(row.date, language) : "-"}
                      </td>
                      <td className="py-1.5 px-2 whitespace-nowrap">
                        {row.docno ? (
                          <button
                            type="button"
                            className="font-bold text-primary hover:underline hover:text-primary/80 cursor-pointer inline-flex items-center gap-1"
                            onClick={() => setDrillDoc({ docno: row.docno, journalId: row.journalid || "" })}
                            title={tr("gl_view_voucher_detail", "คลิกเพื่อดูใบสำคัญรายวัน")}
                          >
                            <span>{row.docno}</span>
                            <ExternalLink className="size-3 opacity-60" />
                          </button>
                        ) : (
                          "-"
                        )}
                      </td>
                      <td className="py-1.5 px-2 whitespace-nowrap font-sans text-muted-foreground" title={bName}>
                        {bCode || "-"}
                      </td>
                      <td className="py-1.5 px-2 font-sans truncate max-w-xs text-foreground" title={row.description}>
                        {row.description || "-"}
                      </td>
                      <td className="py-1.5 px-2 text-center text-muted-foreground">
                        {row.branchcode || "-"}
                      </td>
                      <td className="py-1.5 px-2 text-center text-muted-foreground">
                        {row.departmentcode || "-"}
                      </td>
                      <td className="py-1.5 px-2 text-center text-muted-foreground">
                        {row.projectcode || "-"}
                      </td>
                      <td className="py-1.5 px-2 text-right font-medium text-foreground">
                        {row.debit && row.debit !== "0.00" ? formatAmount(row.debit) : "-"}
                      </td>
                      <td className="py-1.5 px-2 text-right font-medium text-foreground">
                        {row.credit && row.credit !== "0.00" ? formatAmount(row.credit) : "-"}
                      </td>
                      <td className="py-1.5 pr-2.5 pl-2 text-right font-bold text-foreground">
                        {row.balance ? formatAmount(row.balance) : "-"}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>

        {/* Footer */}
        <div className="p-3 border-t border-border bg-muted/30 flex items-center justify-end gap-2 shrink-0">
          <Button type="button" variant="outline" className={actionClass} onClick={onClose}>
            {tr("gl_close", "ปิด")}
          </Button>
        </div>
      </div>

      {/* Nested Voucher Drilldown */}
      {drillDoc && (
        <JournalDrillDownModal
          docno={drillDoc.docno}
          journalId={drillDoc.journalId}
          open={Boolean(drillDoc)}
          onClose={() => setDrillDoc(null)}
          books={books}
        />
      )}
    </div>
  );
}
