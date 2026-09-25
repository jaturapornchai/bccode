"use client";

import { Fragment, useEffect, useRef, useState } from "react";
import { AlertCircle, Calendar, FileText, Loader2, Printer, Scale } from "lucide-react";
import { useBackendText } from "@/components/backend-text-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { formatAmount } from "@/lib/general-ledger";
import { formatAppDate } from "@/lib/date-time";
import type { LanguageCode } from "@/lib/i18n";
import { TAX_MONTH_KEYS, fetchVatSummary, type VatSummaryResult, type VatSummarySort } from "@/lib/thai-tax";

const SORTS: VatSummarySort[] = ["date", "taxno", "docno"];
const SORT_FALLBACK: Record<VatSummarySort, string> = { date: "วันที่ใบกำกับ", taxno: "เลขที่ใบกำกับ", docno: "เลขที่ใบสำคัญ" };
const selectClass =
  "min-h-[2.6em] rounded-lg border border-border bg-background px-3 text-[0.95rem] font-medium text-foreground shadow-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20";
// ยอดเงินเป็น string ทศนิยมจาก backend — จอจัดรูปแบบอย่างเดียว ไม่คำนวณ (ui-scale-polish §8.29)
const money = (value: string) => formatAmount(value, 2);

interface Props {
  language?: LanguageCode;
  holdingcode?: string;
  businesscode?: string;
}

/**
 * รายงานสรุปยอดภาษี (Champ เมนู 5539): ภาษีซื้อที่ใช้สิทธิ + ภาษีขายของงวดภาษี เรียงรวมกันพร้อมยอดคงเหลือสะสม
 * รวมรายวัน และยอดที่ตรงกับ ภ.พ.30 ข้อ 8/9 — ทุกยอดคำนวณที่ backend (POST /api/report/tax/vat-summary)
 */
export function VatSummaryReport({ language = "th", holdingcode = "", businesscode = "" }: Props) {
  const tr = useBackendText();
  const today = new Date();
  const [year, setYear] = useState(today.getFullYear());
  const [month, setMonth] = useState(today.getMonth() + 1);
  const [sort, setSort] = useState<VatSummarySort>("date");
  const [result, setResult] = useState<VatSummaryResult | null>(null);
  const [loading, setLoading] = useState(false);
  // เปลี่ยนงวดเร็ว ๆ แล้วคำตอบเก่ามาทีหลัง ห้ามทับผลของงวดล่าสุด
  const requestRef = useRef(0);

  useEffect(() => {
    const request = ++requestRef.current;
    setLoading(true);
    void fetchVatSummary({ holdingcode, businesscode, year, month, sort, language }).then((loaded) => {
      if (request !== requestRef.current) return;
      setResult(loaded);
      setLoading(false);
    });
  }, [holdingcode, businesscode, year, month, sort, language]);

  const years = [today.getFullYear() - 3, today.getFullYear() - 2, today.getFullYear() - 1, today.getFullYear(), today.getFullYear() + 1];
  const errorText = result && !result.ok
    ? result.message ||
      (result.error === "company_required"
        ? tr("company_required", "กรุณาเลือกบริษัทก่อน")
        : result.error === "unauthorized"
          ? tr("unauthorized", "ไม่มีสิทธิ์เข้าถึงข้อมูล")
          : result.error === "connection_error"
            ? tr("connection_error", "เชื่อมต่อไม่สำเร็จ")
            : tr("tax_vat_summary_load_failed", "โหลดรายงานสรุปยอดภาษีไม่สำเร็จ กรุณาลองใหม่"))
    : "";
  const data = result?.ok ? result : null;
  const dayByDate = new Map((data?.days ?? []).map((d) => [d.date, d]));
  const monthLabel = tr(`month_${TAX_MONTH_KEYS[month - 1]}`, String(month));
  const yearLabel = (y: number) => String(language === "th" ? y + 543 : y);

  return (
    <div className="flex flex-col gap-4 p-4 lg:p-6 print:gap-2 print:p-0" data-screen="vat-summary">
      <div className="flex flex-wrap items-start justify-between gap-4 rounded-2xl border border-border bg-card p-4 shadow-sm">
        <div className="flex items-start gap-3">
          <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-inner print:hidden">
            <Scale className="h-6 w-6" aria-hidden />
          </div>
          <div className="space-y-1">
            <h1 className="text-xl font-bold text-foreground">
              {tr("tax_vat_summary_title", "รายงานสรุปยอดภาษี")} — {monthLabel} {yearLabel(year)}
            </h1>
            <p className="max-w-3xl text-[0.9rem] leading-relaxed text-muted-foreground print:hidden">
              {tr("tax_vat_summary_hint", "ภาษีซื้อที่ใช้สิทธิและภาษีขายของงวดภาษีนี้ (ชุดเดียวกับ ภ.พ.30) เรียงรวมกัน พร้อมยอดคงเหลือสะสม = ภาษีซื้อ − ภาษีขาย")}
            </p>
          </div>
        </div>
        <Button variant="default" disabled={!data || data.rows.length === 0} onClick={() => window.print()} className="gap-2 shadow-sm print:hidden">
          <Printer className="h-4 w-4" aria-hidden />
          {tr("print_report", "พิมพ์รายงาน")}
        </Button>
      </div>

      <Card className="shadow-sm print:hidden">
        <CardContent className="flex flex-wrap items-center gap-3 p-4">
          <span className="flex items-center gap-2 text-[0.95rem] font-medium text-muted-foreground">
            <Calendar className="h-4 w-4 text-primary" aria-hidden />
            {tr("ops_tax_period", "งวดภาษี:")}
          </span>
          <select aria-label={tr("ops_tax_period", "งวดภาษี:")} value={month} onChange={(e) => setMonth(Number(e.target.value))} className={selectClass} data-field="month">
            {TAX_MONTH_KEYS.map((key, i) => (
              <option key={key} value={i + 1}>{tr(`month_${key}`, String(i + 1))}</option>
            ))}
          </select>
          <select aria-label={tr("tax_vat_summary_year", "ปี")} value={year} onChange={(e) => setYear(Number(e.target.value))} className={selectClass} data-field="year">
            {years.map((y) => <option key={y} value={y}>{yearLabel(y)}</option>)}
          </select>
          <label className="flex items-center gap-2 text-[0.95rem] font-medium text-muted-foreground">
            {tr("tax_vat_summary_sort", "เรียงตาม")}
            <select value={sort} onChange={(e) => setSort(e.target.value as VatSummarySort)} className={selectClass} data-field="sort">
              {SORTS.map((s) => <option key={s} value={s}>{tr(`tax_vat_summary_sort_${s}`, SORT_FALLBACK[s])}</option>)}
            </select>
          </label>
        </CardContent>
      </Card>

      {loading && !data ? (
        <div role="status" className="flex items-center gap-2 rounded-xl border border-border bg-muted/30 px-4 py-3 text-[0.95rem]">
          <Loader2 className="h-4 w-4 animate-spin" aria-hidden />
          {tr("tax_vat_summary_loading", "กำลังโหลดรายงานสรุปยอดภาษี...")}
        </div>
      ) : errorText ? (
        <div role="alert" className="flex items-start gap-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-[0.95rem] text-destructive">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden />
          <span>{errorText}</span>
        </div>
      ) : data && data.rows.length === 0 ? (
        <div role="status" className="flex items-center gap-2 rounded-xl border border-border bg-muted/30 px-4 py-3 text-[0.95rem] text-muted-foreground">
          <FileText className="h-4 w-4" aria-hidden />
          {tr("tax_check_no_rows", "ไม่มีรายการในงวดนี้")}
        </div>
      ) : data ? (
        <>
          <div className="overflow-x-auto rounded-2xl border border-border bg-card shadow-sm" aria-busy={loading}>
            <table className="w-full text-left text-[0.95rem]">
              <thead className="border-b bg-muted/60 text-[0.9rem] font-semibold text-muted-foreground">
                <tr>
                  <th className="w-12 px-3 py-3 text-center">{tr("tax_register_col_seq", "ลำดับ")}</th>
                  <th className="px-3 py-3 whitespace-nowrap">{tr("tax_vat_summary_col_taxdate", "วันที่ใบกำกับ")}</th>
                  <th className="px-3 py-3">{tr("tax_vat_summary_col_taxno", "เลขที่ใบกำกับ")}</th>
                  <th className="px-3 py-3">{tr("tax_vat_summary_col_docno", "เลขที่ใบสำคัญ")}</th>
                  <th className="px-3 py-3">{tr("tax_vat_summary_col_desc", "รายละเอียด")}</th>
                  <th className="px-3 py-3 text-right">{tr("tax_vat_summary_col_taxin", "ภาษีซื้อ")}</th>
                  <th className="px-3 py-3 text-right">{tr("tax_vat_summary_col_taxout", "ภาษีขาย")}</th>
                  <th className="px-3 py-3 text-right">{tr("tax_vat_summary_col_balance", "คงเหลือ (ซื้อ − ขาย)")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {data.rows.map((row, i) => {
                  // รวมรายวันต่อท้ายแถวสุดท้ายของวันนั้น (backend ส่งเฉพาะเมื่อเรียงตามวันที่)
                  const day = data.rows[i + 1]?.taxdate === row.taxdate ? undefined : dayByDate.get(row.taxdate);
                  return (
                    <Fragment key={`${row.no}-${row.journalid}`}>
                      <tr className="hover:bg-muted/30" data-row="item">
                        <td className="px-3 py-2 text-center text-muted-foreground">{row.no}</td>
                        <td className="px-3 py-2 whitespace-nowrap">{formatAppDate(row.taxdate, language)}</td>
                        <td className="px-3 py-2 font-mono [overflow-wrap:anywhere]">{row.taxinvoiceno}</td>
                        <td className="px-3 py-2 font-mono whitespace-nowrap">{row.docno}</td>
                        <td className="px-3 py-2 leading-normal [overflow-wrap:anywhere]">{row.description}</td>
                        <td className="px-3 py-2 text-right font-mono">{row.taxin ? money(row.taxin) : ""}</td>
                        <td className="px-3 py-2 text-right font-mono">{row.taxout ? money(row.taxout) : ""}</td>
                        <td className="px-3 py-2 text-right font-mono font-semibold">{money(row.balance)}</td>
                      </tr>
                      {day && (
                        <tr className="bg-muted/40 font-semibold" data-row="day-total">
                          <td colSpan={5} className="px-3 py-2 text-right">
                            {tr("tax_vat_summary_day_total", "รวมวันที่ {date} ({count} รายการ)")
                              .replace("{date}", formatAppDate(day.date, language))
                              .replace("{count}", String(day.count))}
                          </td>
                          <td className="px-3 py-2 text-right font-mono">{money(day.taxin)}</td>
                          <td className="px-3 py-2 text-right font-mono">{money(day.taxout)}</td>
                          <td className="px-3 py-2 text-right font-mono">{money(day.net)}</td>
                        </tr>
                      )}
                    </Fragment>
                  );
                })}
              </tbody>
              <tfoot className="border-t-2 border-border bg-primary/10 font-bold text-foreground">
                <tr data-row="grand-total">
                  <td colSpan={5} className="px-3 py-3 text-right">
                    {tr("tax_vat_summary_grand_total", "รวมทั้งงวด ({count} รายการ)").replace("{count}", String(data.totals.count))}
                  </td>
                  <td className="px-3 py-3 text-right font-mono">{money(data.totals.taxin)}</td>
                  <td className="px-3 py-3 text-right font-mono">{money(data.totals.taxout)}</td>
                  <td className="px-3 py-3 text-right font-mono">{money(data.totals.net)}</td>
                </tr>
              </tfoot>
            </table>
          </div>
          <div className="grid gap-3 md:grid-cols-2">
            <div className="rounded-2xl border border-border bg-card p-4 shadow-sm" data-field="taxpayable">
              <p className="text-[0.95rem] text-muted-foreground">{tr("tax_vat_summary_payable", "ภาษีที่ต้องชำระเดือนนี้ (ภ.พ.30 ข้อ 8)")}</p>
              <p className="font-mono text-2xl font-bold text-primary">{money(data.totals.taxpayable)}</p>
            </div>
            <div className="rounded-2xl border border-border bg-card p-4 shadow-sm" data-field="taxexcess">
              <p className="text-[0.95rem] text-muted-foreground">{tr("tax_vat_summary_excess", "ภาษีที่ชำระเกินเดือนนี้ (ภ.พ.30 ข้อ 9)")}</p>
              <p className="font-mono text-2xl font-bold text-foreground">{money(data.totals.taxexcess)}</p>
            </div>
          </div>
        </>
      ) : null}
    </div>
  );
}
