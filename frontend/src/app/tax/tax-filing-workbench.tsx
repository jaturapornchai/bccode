"use client";

import { useState, useMemo, useEffect, useCallback } from "react";
import {
  useBackendDictionary,
  useBackendText,
} from "@/components/backend-text-provider";
import {
  getThaiTaxConfig,
  taxText,
  fetchVatRegister,
  fetchPp30Summary,
  fetchWhtReport,
  type CompanyHeader,
  type ThaiTaxRecord,
  type Pp30Summary,
  type VatRegisterSummary,
  type WhtReportRow,
  type WhtReportSummary,
} from "@/lib/thai-tax";
import { formatAmount } from "@/lib/general-ledger";
import type { LanguageCode } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChoiceSelect } from "@/components/ui/select";
import { Card, CardContent } from "@/components/ui/card";
import { WhtCertificatePanel } from "./wht-certificate-panel";
import {
  FileText, Printer, Download, Calculator, Building2, Calendar, CheckCircle2, Receipt, Search,
  AlertCircle, Loader2, ShieldCheck,
} from "lucide-react";

interface TaxFilingWorkbenchProps {
  route: string;
  embedded?: boolean;
  language?: LanguageCode;
  holdingcode?: string;
  businesscode?: string;
}

// ยอดเงินมาจาก backend เป็น string ทศนิยม — จอนี้จัดรูปแบบแสดงผลอย่างเดียว ไม่คำนวณภาษีเอง
const money = (value: string | undefined) => formatAmount(value ?? "0.00", 2);
const isZeroMoney = (value: string) => /^-?0*(\.0*)?$/.test(value.trim());
// ช่องกรอกภาษีชำระเกินยกมา: ตัวเลขไม่ติดลบ ทศนิยมไม่เกิน 2 ตำแหน่ง (backend ตรวจซ้ำ)
const CREDIT_INPUT = /^\d{0,13}(\.\d{0,2})?$/;
// ขอทั้งงวดในหน้าเดียวตามเพดาน backend (taxRegisterMaxLimit) — ยอดรวมท้ายตารางมาจาก backend ทั้งงวดเสมอ
const REGISTER_PAGE_LIMIT = 500;

export function TaxFilingWorkbench({
  route,
  embedded: _embedded = false,
  language = "th",
  holdingcode = "",
  businesscode = "",
}: TaxFilingWorkbenchProps) {
  const tr = useBackendText();
  const dictionary = useBackendDictionary();
  const config = getThaiTaxConfig(route) || {
    route,
    code: "tax_filing",
    title: { th: "รายงานและแบบยื่นภาษี", en: "Tax Filing & Reports" },
    formType: "vat_sale" as const,
    description: { th: "ระบบภาษีมูลค่าเพิ่มและภาษีหัก ณ ที่จ่าย", en: "VAT & WHT System" },
    revenueDepartmentFormCode: "สรรพากร",
  };

  const [selectedYear, setSelectedYear] = useState<number>(() => new Date().getFullYear());
  const [selectedMonth, setSelectedMonth] = useState<number>(() => new Date().getMonth() + 1);
  const [searchTerm, setSearchTerm] = useState<string>("");
  const [activeTab, setActiveTab] = useState<
    "table" | "pp30" | "annex_sales" | "annex_purchases" | "pnd_form" | "50twi"
  >(() => (config.formType === "pp30" ? "pp30" : config.formType.includes("pnd") ? "pnd_form" : config.formType === "50twi" ? "50twi" : "table"));

  const [selectedRecord, setSelectedRecord] = useState<ThaiTaxRecord | null>(null);
  const [selectedWhtRow, setSelectedWhtRow] = useState<WhtReportRow | null>(null);

  const [records, setRecords] = useState<ThaiTaxRecord[]>([]);
  const [salesRecords, setSalesRecords] = useState<ThaiTaxRecord[]>([]);
  const [purchaseRecords, setPurchaseRecords] = useState<ThaiTaxRecord[]>([]);
  const [salesSummary, setSalesSummary] = useState<{ total: number; summary: VatRegisterSummary } | null>(null);
  const [purchaseSummary, setPurchaseSummary] = useState<{ total: number; summary: VatRegisterSummary } | null>(null);
  const [whtRows, setWhtRows] = useState<WhtReportRow[]>([]);
  const [whtSummary, setWhtSummary] = useState<{ total: number; summary: WhtReportSummary; note: string } | null>(null);
  const [company, setCompany] = useState<CompanyHeader | null>(null);
  const [pp30, setPp30] = useState<Pp30Summary | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [errorKey, setErrorKey] = useState<string | null>(null);

  // ภาษีชำระเกินยกมาจากเดือนก่อน (ข้อ 8 ของ ภ.พ.30) — ส่งให้ backend คำนวณข้อ 9/10 เมื่อกรอกเสร็จ
  const [creditInput, setCreditInput] = useState<string>("");
  const [creditBroughtForward, setCreditBroughtForward] = useState<string>("");

  // รองรับทั้ง VAT และ WHT ทุกประเภท
  const isVatType = config.formType === "vat_sale" || config.formType === "vat_buy" || config.formType === "pp30" || config.formType === "pp36";
  const isWhtType = config.formType === "pnd2" || config.formType === "pnd3" || config.formType === "pnd53" || config.formType === "50twi" || config.formType === "wht_received" || config.formType === "wht_summary";

  const loadData = useCallback(async () => {
    setLoading(true);
    setErrorKey(null);

    try {
      if (isVatType) {
        const period = { holdingcode, businesscode, year: selectedYear, month: selectedMonth, limit: REGISTER_PAGE_LIMIT };
        if (config.formType === "pp30") {
          const [salesResult, purchaseResult, summaryResult] = await Promise.all([
            fetchVatRegister({ ...period, type: "sale" }),
            fetchVatRegister({ ...period, type: "purchase" }),
            fetchPp30Summary({ holdingcode, businesscode, year: selectedYear, month: selectedMonth, creditbroughtforward: creditBroughtForward }),
          ]);

          setSalesRecords(salesResult.records);
          setPurchaseRecords(purchaseResult.records);
          setSalesSummary({ total: salesResult.total, summary: salesResult.summary });
          setPurchaseSummary({ total: purchaseResult.total, summary: purchaseResult.summary });
          setRecords(salesResult.records);
          setPp30(summaryResult.summary);
          setCompany(summaryResult.summary?.company ?? null);
          setErrorKey(summaryResult.error ?? salesResult.error ?? purchaseResult.error ?? null);
        } else {
          const registerType: "sale" | "purchase" =
            config.formType === "vat_buy" ? "purchase" : "sale";
          const registerResult = await fetchVatRegister({ ...period, type: registerType });
          const periodSummary = { total: registerResult.total, summary: registerResult.summary };

          setRecords(registerResult.records);
          if (registerType === "sale") {
            setSalesRecords(registerResult.records);
            setSalesSummary(periodSummary);
          } else {
            setPurchaseRecords(registerResult.records);
            setPurchaseSummary(periodSummary);
          }
          setPp30(null);
          setErrorKey(registerResult.error ?? null);
        }
      } else if (isWhtType) {
        // ข้อมูลภาษีหัก ณ ที่จ่ายจากบัญชีแยกประเภทที่ผ่านรายการจริง (backend /api/report/tax/wht)
        // เลือกทิศทางและแบบยื่นตามจอ: ภ.ง.ด.2 = ดอกเบี้ย/ปันผล (บัญชีแยกแบบยื่นไว้แล้วในผังบัญชี)
        const whtDirection: "paid" | "received" = config.formType === "wht_received" ? "received" : "paid";
        const whtForms = config.formType === "pnd2" ? ["2"] : undefined;
        const whtResult = await fetchWhtReport({
          holdingcode,
          businesscode,
          year: selectedYear,
          month: selectedMonth,
          direction: whtDirection,
          forms: whtForms,
          limit: REGISTER_PAGE_LIMIT,
        });

        setWhtRows(whtResult.rows);
        setSelectedWhtRow(whtResult.rows[0] ?? null);
        setWhtSummary(whtResult.error ? null : { total: whtResult.total, summary: whtResult.summary, note: whtResult.note });
        setCompany(whtResult.error ? null : whtResult.company);

        // แปลงเป็น ThaiTaxRecord เพื่อใช้ตารางทะเบียนร่วมกับภาษีมูลค่าเพิ่ม (ยอดทุกช่องมาจาก backend)
        setRecords(whtResult.rows.map((row) => ({
          id: `wht-${row.journalid}`,
          docdate: row.docdate,
          taxinvoiceno: row.docno,
          counterpartyname: row.partnername,
          taxid: row.taxid,
          branchno: "",
          isheadoffice: false,
          amountbeforevat: row.baseamount,
          vatamount: "0.00",
          totalamount: row.netamount,
          whtamount: row.whtamount,
          taxrate: row.ratepercent,
          incometype: row.description,
          status: "active",
        })));
        setPp30(null);
        setErrorKey(whtResult.error ?? null);
      }
    } catch {
      setRecords([]);
      setSalesRecords([]);
      setPurchaseRecords([]);
      setWhtRows([]);
      setSalesSummary(null);
      setPurchaseSummary(null);
      setWhtSummary(null);
      setPp30(null);
      setErrorKey("connection_error");
    } finally {
      setLoading(false);
    }
  }, [config.formType, isVatType, isWhtType, holdingcode, businesscode, selectedYear, selectedMonth, creditBroughtForward]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  // กำหนดเรคคอร์ดที่ต้องแสดงตามแท็บ
  const displayRecords = useMemo(() => {
    if (activeTab === "annex_sales") return salesRecords;
    if (activeTab === "annex_purchases") return purchaseRecords;
    return records;
  }, [activeTab, salesRecords, purchaseRecords, records]);

  const filteredRecords = useMemo(() => {
    if (!searchTerm.trim()) return displayRecords;
    const term = searchTerm.toLowerCase();
    return displayRecords.filter(
      (r) =>
        r.taxinvoiceno.toLowerCase().includes(term) ||
        r.counterpartyname.toLowerCase().includes(term) ||
        r.taxid.includes(term),
    );
  }, [displayRecords, searchTerm]);

  // ยอดรวมทั้งงวดจาก backend (decimal) — ไม่บวกเลขบน browser
  const totals = useMemo(() => {
    if (isWhtType) {
      const summary = whtSummary?.summary;
      return {
        count: whtSummary?.total ?? 0,
        beforeVat: summary?.basetotal ?? "0.00",
        tax: summary?.whttotal ?? "0.00",
        total: summary?.nettotal ?? "0.00",
      };
    }
    const register = activeTab === "annex_purchases" || config.formType === "vat_buy" ? purchaseSummary : salesSummary;
    return {
      count: register?.total ?? 0,
      beforeVat: register?.summary.amountbeforevat ?? "0.00",
      tax: register?.summary.vatamount ?? "0.00",
      total: register?.summary.totalamount ?? "0.00",
    };
  }, [isWhtType, whtSummary, activeTab, config.formType, purchaseSummary, salesSummary]);

  const applyCreditBroughtForward = () => {
    const value = creditInput.trim();
    if (value !== creditBroughtForward) setCreditBroughtForward(value);
  };

  const notSpecified = tr("tax_not_specified", "ยังไม่ระบุ");
  const notInCompanyRegister = tr("tax_not_in_company_register", "ยังไม่ระบุในทะเบียนบริษัท");
  const companyName = company?.name || notSpecified;
  const companyTaxId = company?.taxid || notSpecified;

  const canExport = !loading && !errorKey && filteredRecords.length > 0;

  const monthNamesTh = [
    "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
    "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
  ];

  return (
    <div className="flex flex-col gap-4 p-4 lg:p-6 print:p-0 print:gap-2">
      {/* Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-4 shadow-sm print:hidden">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-inner">
            <Receipt className="h-6 w-6" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">
                {taxText(config, "title", language, dictionary)}
              </h1>
              <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary">
                {config.revenueDepartmentFormCode}
              </span>
            </div>
            <p className="text-sm text-muted-foreground">
              {taxText(config, "description", language, dictionary)}
            </p>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {/* Action buttons for PP.30 */}
          {config.formType === "pp30" && (
            <>
              <Button
                variant={activeTab === "pp30" ? "default" : "outline"}
                onClick={() => setActiveTab("pp30")}
                className="gap-2 shadow-sm font-medium"
              >
                <Calculator className="h-4 w-4" />
                {tr("tax_pp30_official", "แบบฟอร์ม ภ.พ.30 สรรพากร")}
              </Button>
              <Button
                variant={activeTab === "annex_sales" ? "default" : "outline"}
                onClick={() => setActiveTab("annex_sales")}
                className="gap-2 shadow-sm font-medium"
              >
                <FileText className="h-4 w-4" />
                {tr("tax_annex_sales", "ใบแนบภาษีขาย")}
              </Button>
              <Button
                variant={activeTab === "annex_purchases" ? "default" : "outline"}
                onClick={() => setActiveTab("annex_purchases")}
                className="gap-2 shadow-sm font-medium"
              >
                <FileText className="h-4 w-4" />
                {tr("tax_annex_purchases", "ใบแนบภาษีซื้อ")}
              </Button>
            </>
          )}

          {/* Action buttons for WHT (PND.3 / PND.53 / 50 Twi) */}
          {isWhtType && (
            <>
              <Button
                variant={activeTab === "pnd_form" ? "default" : "outline"}
                onClick={() => setActiveTab("pnd_form")}
                className="gap-2 shadow-sm font-medium"
              >
                <ShieldCheck className="h-4 w-4" />
                {tr("tax_pnd_official", "แบบยื่นสรรพากรทางการ")}
              </Button>
              <Button
                variant={activeTab === "50twi" ? "default" : "outline"}
                onClick={() => setActiveTab("50twi")}
                className="gap-2 shadow-sm font-medium"
              >
                <FileText className="h-4 w-4" />
                {tr("tax_print_50twi", "หนังสือรับรอง 50 ทวิ")}
              </Button>
              <Button
                variant={activeTab === "table" ? "default" : "outline"}
                onClick={() => setActiveTab("table")}
                className="gap-2 shadow-sm font-medium"
              >
                <Receipt className="h-4 w-4" />
                {tr("ops_table", "ทะเบียนภาษี")}
              </Button>
            </>
          )}

          <Button
            variant="outline"
            disabled={!canExport}
            onClick={() => {
              const csv = [
                "ลำดับ,วันที่,เลขที่เอกสาร,ชื่อคู่ค้า,เลขประจำตัวผู้เสียภาษี,สาขา,มูลค่าก่อนภาษี,ภาษี,ยอดรวม",
                ...filteredRecords.map(
                  (r, idx) =>
                    `${idx + 1},${r.docdate},${r.taxinvoiceno},"${r.counterpartyname}",${r.taxid},${r.branchno},${r.amountbeforevat},${r.whtamount ?? r.vatamount},${r.totalamount}`,
                ),
              ].join("\n");
              const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
              const url = URL.createObjectURL(blob);
              const link = document.createElement("a");
              link.href = url;
              link.setAttribute("download", `${config.code}_${activeTab}_${selectedYear}_${selectedMonth}.csv`);
              document.body.appendChild(link);
              link.click();
              document.body.removeChild(link);
            }}
            className="gap-2"
          >
            <Download className="h-4 w-4" />
            {tr("gl_export_csv", "ส่งออก CSV")}
          </Button>

          {/* ใบ 50 ทวิ พิมพ์/ดาวน์โหลดจากหน้าต่าง PDF ที่ backend สร้าง ไม่ใช่ window.print() ของหน้าเว็บ */}
          {activeTab !== "50twi" && (
            <Button
              variant="default"
              disabled={!canExport && !pp30}
              onClick={() => window.print()}
              className="gap-2 shadow-sm"
            >
              <Printer className="h-4 w-4" />
              {activeTab === "pp30" ? tr("tax_print_pp30", "พิมพ์แบบ ภ.พ.30") : tr("print_report", "พิมพ์รายงาน")}
            </Button>
          )}
        </div>
      </div>

      {/* Filter and Period Selection Bar */}
      <Card className="print:hidden shadow-sm">
        <CardContent className="flex flex-wrap items-center justify-between gap-4 p-4">
          <div className="flex flex-wrap items-center gap-3">
            <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
              <Calendar className="h-4 w-4 text-primary" />
              <span>{tr("ops_tax_period", "งวดภาษี:")}</span>
            </div>
            <select
              value={selectedMonth}
              onChange={(e) => setSelectedMonth(Number(e.target.value))}
              className="rounded-lg border border-border bg-background px-3 py-1.5 text-sm font-medium text-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary shadow-sm"
            >
              {monthNamesTh.map((name, i) => (
                <option key={i + 1} value={i + 1}>
                  {name}
                </option>
              ))}
            </select>
            <ChoiceSelect
              value={selectedYear}
              onChange={(val) => setSelectedYear(Number(val))}
              layout="flex"
              className="w-auto"
              radioClassName="w-auto px-3 min-h-[2.2em] text-xs font-semibold"
              options={[
                { value: 2026, label: "พ.ศ. 2569 (2026)" },
                { value: 2025, label: "พ.ศ. 2568 (2025)" },
              ]}
            />

            <div className="flex items-center gap-1.5 rounded-lg bg-muted px-3 py-1.5 text-xs text-muted-foreground">
              <Building2 className="h-3.5 w-3.5" />
              <span>{companyName}</span>
            </div>
          </div>

          <div className="flex items-center gap-3">
            {config.formType === "pp30" && (
              <div className="flex items-center gap-2 text-xs">
                <span className="text-muted-foreground">เครดิตยกมา (ข้อ 8):</span>
                <input
                  type="text"
                  inputMode="decimal"
                  value={creditInput}
                  onChange={(e) => {
                    if (CREDIT_INPUT.test(e.target.value)) setCreditInput(e.target.value);
                  }}
                  onBlur={applyCreditBroughtForward}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") applyCreditBroughtForward();
                  }}
                  placeholder="0.00"
                  className="w-24 rounded-lg border border-border bg-background px-2 py-1 text-right font-mono text-xs focus:border-primary focus:outline-none"
                />
                <span className="text-muted-foreground">บาท</span>
              </div>
            )}

            <div className="relative w-full max-w-xs">
              <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder={tr("ops_search_2", "ค้นหาเลขที่, ชื่อคู่ค้า, เลขประจำตัว...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-9"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* KPI Cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4 print:hidden">
        <Card className="p-4 shadow-sm border-border">
          <p className="text-xs text-muted-foreground">
            {isWhtType ? "มูลค่าเงินได้พึงประเมินก่อนหักภาษี" : tr("ops_base_amount_before_vat", "มูลค่าสินค้า/บริการก่อนภาษี")}
          </p>
          <p className="mt-1 text-xl font-bold text-foreground font-mono">
            {money(totals.beforeVat)}
          </p>
          <span className="text-xs text-muted-foreground">บาท (THB)</span>
        </Card>

        <Card className="p-4 shadow-sm border-border">
          <p className="text-xs text-muted-foreground">
            {isWhtType
              ? tr("ops_total_withholding_tax", "ยอดภาษีหัก ณ ที่จ่ายรวม")
              : tr("ops_total_vat_7", "ยอดภาษีมูลค่าเพิ่ม 7%")}
          </p>
          <p className="mt-1 text-xl font-bold text-primary font-mono">
            {money(totals.tax)}
          </p>
          <span className="text-xs text-muted-foreground">บาท (THB)</span>
        </Card>

        <Card className="p-4 shadow-sm border-border">
          <p className="text-xs text-muted-foreground">{tr("ops_total_documents", "จำนวนรายการเอกสาร")}</p>
          <p className="mt-1 text-xl font-bold text-foreground font-mono">
            {totals.count}
          </p>
          <span className="text-xs text-muted-foreground">รายการ (Docs)</span>
        </Card>

        <Card className="p-4 shadow-sm border-border">
          <p className="text-xs text-muted-foreground">{tr("ops_compliance_status", "สถานะการตรวจสอบ")}</p>
          <div className="mt-1 flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400">
            <CheckCircle2 className="h-5 w-5" />
            <span className="text-base font-semibold">{tr("ops_ready_to_file", "ถูกต้อง พร้อมยื่นแบบ")}</span>
          </div>
          <span className="text-xs text-muted-foreground">
            Tax ID ครบ 13 หลัก
          </span>
        </Card>
      </div>

      {loading ? (
        <div className="flex items-center gap-2 rounded-xl border border-border bg-muted/30 px-4 py-3 text-sm print:hidden">
          <Loader2 className="h-4 w-4 shrink-0 animate-spin" />
          <span>กำลังโหลดข้อมูลภาษีและสมุดรายวัน...</span>
        </div>
      ) : errorKey ? (
        <div
          role="alert"
          className="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive print:hidden"
        >
          <div className="flex items-start gap-2">
            <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
            <span>
              {errorKey === "company_required"
                ? "กรุณาเลือกบริษัทก่อน"
                : errorKey === "unauthorized"
                  ? "ไม่มีสิทธิ์เข้าถึงข้อมูล"
                  : errorKey === "connection_error"
                    ? "เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ กรุณาตรวจสอบอินเทอร์เน็ต"
                    : "โหลดข้อมูลไม่สำเร็จ"}
            </span>
          </div>
        </div>
      ) : whtSummary?.note ? (
        <div role="status" className="flex items-start gap-2 rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-800 dark:text-amber-200 print:hidden">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          <span>{whtSummary.note}</span>
        </div>
      ) : null}

      {/* Main Content Area based on Active Tab */}

      {/* 1. แบบฟอร์ม ภ.พ. 30 สรรพากร (Official RD Form View) */}
      {config.formType === "pp30" && activeTab === "pp30" && pp30 && (
        <Card className="overflow-hidden border-2 border-primary/20 shadow-md print:border-none print:shadow-none bg-card">
          <div className="border-b bg-muted/40 p-5 print:bg-white print:border-b-2 print:border-black">
            <div className="flex flex-wrap items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary print:hidden">
                  <ShieldCheck className="h-7 w-7" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="text-xl font-extrabold text-foreground tracking-tight">ภ.พ. 30</span>
                    <span className="rounded bg-primary/10 px-2 py-0.5 text-xs font-bold text-primary print:border print:border-black">
                      แบบแสดงรายการภาษีมูลค่าเพิ่ม
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground print:text-black">
                    ตามประมวลรัษฎากร กรมสรรพากร กระทรวงการคลัง
                  </p>
                </div>
              </div>

              <div className="text-right text-xs space-y-0.5 print:text-black">
                <p><span className="font-semibold">เดือนภาษี:</span> {monthNamesTh[selectedMonth - 1]}</p>
                <p><span className="font-semibold">พ.ศ.:</span> {selectedYear + 543}</p>
                <p><span className="font-semibold">ผู้ประกอบการ:</span> {companyName}</p>
                <p><span className="font-semibold">เลขประจำตัวผู้เสียภาษี:</span> <span className="font-mono font-bold text-primary print:text-black">{companyTaxId}</span></p>
                <p><span className="font-semibold">สถานประกอบการ:</span> {notInCompanyRegister}</p>
              </div>
            </div>
          </div>

          <div className="p-5 space-y-4">
            <h3 className="text-sm font-bold text-foreground uppercase tracking-wider border-b pb-1.5">
              การคำนวณภาษีมูลค่าเพิ่ม (ตามมาตรา 79, 81, 82 แห่งประมวลรัษฎากร)
            </h3>

            <div className="divide-y divide-border text-sm font-medium">
              <div className="flex items-center justify-between py-2.5 hover:bg-muted/20 px-2 rounded">
                <span className="text-foreground">1. ยอดขายเดือนนี้ (ตามมาตรา 79)</span>
                <span className="font-mono text-base font-semibold">{money(pp30.salesgross)} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2 pl-6 text-muted-foreground hover:bg-muted/20 px-2 rounded">
                <span>2. ยอดขายที่เสียภาษีอัตราร้อยละ 0</span>
                <span className="font-mono">{money(pp30.saleszerorated)} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2 pl-6 text-muted-foreground hover:bg-muted/20 px-2 rounded">
                <span>3. ยอดขายที่ได้รับการยกเว้นภาษี</span>
                <span className="font-mono">{money(pp30.salesexempt)} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2.5 bg-muted/30 px-3 rounded font-semibold text-foreground">
                <span>4. ยอดขายที่ต้องเสียภาษี (ข้อ 1 - ข้อ 2 - ข้อ 3)</span>
                <span className="font-mono text-base text-primary">{money(pp30.salestaxable)} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2.5 bg-primary/5 px-3 rounded font-bold text-primary">
                <span>5. ภาษีขาย (ร้อยละ 7 ของยอดขายตามข้อ 4)</span>
                <span className="font-mono text-lg">{money(pp30.outputvat)} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2.5 hover:bg-muted/20 px-2 rounded">
                <span className="text-foreground">6. ยอดซื้อที่มีสิทธินำภาษีซื้อมาหักในการคำนวณภาษี</span>
                <span className="font-mono font-semibold">{money(pp30.purchasetaxable)} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2.5 bg-primary/5 px-3 rounded font-bold text-primary">
                <span>7. ภาษีซื้อ (ตามใบกำกับภาษีซื้อที่มีสิทธินำมาหัก)</span>
                <span className="font-mono text-lg">{money(pp30.inputvat)} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2 pl-6 text-muted-foreground hover:bg-muted/20 px-2 rounded">
                <span>8. ภาษีมูลค่าเพิ่มชำระเกินยกมาจากเดือนก่อน (ถ้ามี)</span>
                <span className="font-mono font-semibold">{money(pp30.creditbroughtforward)} บาท</span>
              </div>

              <div
                className={`flex items-center justify-between py-3 px-4 rounded-xl font-bold my-2 ${
                  !isZeroMoney(pp30.payable)
                    ? "bg-primary/10 text-primary border border-primary/30"
                    : "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border border-emerald-500/30"
                }`}
              >
                <div>
                  <span className="text-base">
                    {!isZeroMoney(pp30.payable)
                      ? "9. ภาษีมูลค่าเพิ่มที่ต้องชำระเดือนนี้ (ถ้าข้อ 5 มากกว่า ข้อ 7 และ ข้อ 8)"
                      : "10. ภาษีมูลค่าเพิ่มชำระเกินเดือนนี้ (ถ้าข้อ 7 และ ข้อ 8 มากกว่า ข้อ 5)"}
                  </span>
                  <p className="text-xs font-normal opacity-80 mt-0.5">
                    {!isZeroMoney(pp30.payable)
                      ? "นำส่งชำระต่อกรมสรรพากรภายในวันที่ 15 ของเดือนถัดไป (หรือ 23 หากยื่นผ่านอินเทอร์เน็ต)"
                      : "ขอคืนเป็นเงินสด หรือ ยกยอดไปเครดิตภาษีในเดือนถัดไป"}
                  </p>
                </div>
                <span className="font-mono text-2xl">
                  {money(isZeroMoney(pp30.payable) ? pp30.creditable : pp30.payable)} บาท
                </span>
              </div>
            </div>

            <div className="mt-8 pt-6 border-t grid grid-cols-2 gap-8 text-center text-xs">
              <div className="space-y-12">
                <p className="font-medium text-muted-foreground">ลงชื่อ.......................................................... ผู้จ่ายเงิน/ผู้มีอำนาจลงนาม</p>
                <p className="text-muted-foreground">(..........................................................)</p>
                <p className="text-muted-foreground">วันที่ ......./......./.......</p>
              </div>
              <div className="space-y-12">
                <p className="font-medium text-muted-foreground">ลงชื่อ.......................................................... ผู้ทำบัญชี/ผู้ตรวจสอบ</p>
                <p className="text-muted-foreground">(..........................................................)</p>
                <p className="text-muted-foreground">วันที่ ......./......./.......</p>
              </div>
            </div>
          </div>
        </Card>
      )}

      {/* 3. แท็บแบบยื่นภาษีหัก ณ ที่จ่าย ภ.ง.ด. 3 หรือ ภ.ง.ด. 53 สรรพากร (Official PND Form) */}
      {isWhtType && activeTab === "pnd_form" && whtSummary && (
        <Card className="overflow-hidden border-2 border-primary/20 shadow-md print:border-none print:shadow-none bg-card">
          <div className="border-b bg-muted/40 p-5 print:bg-white print:border-b-2 print:border-black">
            <div className="flex flex-wrap items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary print:hidden">
                  <ShieldCheck className="h-7 w-7" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="text-xl font-extrabold text-foreground tracking-tight">
                      {config.formType === "pnd3" ? "ภ.ง.ด. 3" : "ภ.ง.ด. 53"}
                    </span>
                    <span className="rounded bg-primary/10 px-2 py-0.5 text-xs font-bold text-primary print:border print:border-black">
                      แบบยื่นรายการภาษีเงินได้หัก ณ ที่จ่าย
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground print:text-black">
                    {config.formType === "pnd3"
                      ? "สำหรับการหักภาษีบุคคลธรรมดา ตามมาตรา 50 และ 52 แห่งประมวลรัษฎากร"
                      : "สำหรับการหักภาษีนิติบุคคล ตามมาตรา 3 เตรส และ 69 ตรี แห่งประมวลรัษฎากร"}
                  </p>
                </div>
              </div>

              <div className="text-right text-xs space-y-0.5 print:text-black">
                <p><span className="font-semibold">เดือนภาษี:</span> {monthNamesTh[selectedMonth - 1]}</p>
                <p><span className="font-semibold">พ.ศ.:</span> {selectedYear + 543}</p>
                <p><span className="font-semibold">ผู้จ่ายเงิน:</span> {companyName}</p>
                <p><span className="font-semibold">เลขประจำตัวผู้จ่ายเงิน:</span> <span className="font-mono font-bold text-primary print:text-black">{companyTaxId}</span></p>
              </div>
            </div>
          </div>

          <div className="p-5 space-y-4">
            <h3 className="text-sm font-bold text-foreground uppercase tracking-wider border-b pb-1.5">
              สรุปรายการภาษีเงินได้หัก ณ ที่จ่ายที่นำส่งเดือนนี้
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="rounded-xl border border-border p-4 bg-muted/20">
                <span className="text-xs text-muted-foreground">จำนวนผู้มีเงินได้ (ราย)</span>
                <p className="text-2xl font-bold font-mono text-foreground mt-1">{whtSummary.summary.payeecount}</p>
                <span className="text-xs text-muted-foreground">ราย</span>
              </div>
              <div className="rounded-xl border border-border p-4 bg-muted/20">
                <span className="text-xs text-muted-foreground">รวมยอดเงินได้ที่จ่ายทั้งสิ้น</span>
                <p className="text-2xl font-bold font-mono text-foreground mt-1">{money(whtSummary.summary.basetotal)}</p>
                <span className="text-xs text-muted-foreground">บาท (THB)</span>
              </div>
              <div className="rounded-xl border border-primary/30 p-4 bg-primary/10">
                <span className="text-xs text-primary font-medium">รวมยอดภาษีที่นำส่งทั้งสิ้น</span>
                <p className="text-2xl font-bold font-mono text-primary mt-1">{money(whtSummary.summary.whttotal)}</p>
                <span className="text-xs text-primary font-semibold">({whtSummary.summary.whttotaltext})</span>
              </div>
            </div>

            {/* ตารางจำแนกตามประเภทเงินได้ */}
            <div className="overflow-x-auto mt-4">
              <table className="w-full text-sm text-left">
                <thead className="bg-muted/60 text-xs font-bold uppercase text-muted-foreground">
                  <tr>
                    <th className="p-3">อัตราภาษี</th>
                    <th className="p-3 text-center">จำนวนราย</th>
                    <th className="p-3 text-right">จำนวนเงินได้ที่จ่าย (บาท)</th>
                    <th className="p-3 text-right">จำนวนภาษีที่นำส่ง (บาท)</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {whtSummary.summary.byrate.map((row) => (
                    <tr key={row.ratepercent} className="hover:bg-muted/30">
                      <td className="p-3 font-mono font-medium text-foreground">{row.ratepercent ? `${row.ratepercent}%` : notSpecified}</td>
                      <td className="p-3 text-center font-mono">{row.count}</td>
                      <td className="p-3 text-right font-mono">{money(row.baseamount)}</td>
                      <td className="p-3 text-right font-mono font-bold text-primary">{money(row.whtamount)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="mt-8 pt-6 border-t grid grid-cols-2 gap-8 text-center text-xs">
              <div className="space-y-12">
                <p className="font-medium text-muted-foreground">ลงชื่อ.......................................................... ผู้มีหน้าที่หักภาษี ณ ที่จ่าย</p>
                <p className="text-muted-foreground">(..........................................................)</p>
                <p className="text-muted-foreground">วันที่ ......./......./.......</p>
              </div>
              <div className="space-y-12">
                <p className="font-medium text-muted-foreground">ประทับตรานิติบุคคล (ถ้ามี)</p>
              </div>
            </div>
          </div>
        </Card>
      )}

      {/* 4. แท็บหนังสือรับรองการหักภาษี ณ ที่จ่าย (ใบ 50 ทวิ) — backend สร้าง PDF บนแบบฟอร์มกรมสรรพากร */}
      {isWhtType && activeTab === "50twi" && selectedWhtRow && (
        <WhtCertificatePanel
          key={selectedWhtRow.journalid}
          row={selectedWhtRow}
          company={company}
          holdingcode={holdingcode}
          businesscode={businesscode}
          formType={config.formType}
          language={language}
        />
      )}

      {/* 5. ตารางรายงานภาษี หรือ ใบแนบภาษีขาย/ซื้อ (Tax Register & Annex Schedules) */}
      {(activeTab === "table" || activeTab === "annex_sales" || activeTab === "annex_purchases") && (
        <Card className="overflow-hidden shadow-sm">
          <div className="border-b bg-muted/40 p-4 flex flex-wrap items-center justify-between gap-2">
            <div>
              <h2 className="text-base font-bold text-foreground">
                {activeTab === "annex_sales"
                  ? tr("tax_annex_title_sales", "รายงานภาษีขาย (ใบแนบแบบ ภ.พ.30 ตามมาตรา 87(1))")
                  : activeTab === "annex_purchases"
                    ? tr("tax_annex_title_purchases", "รายงานภาษีซื้อ (ใบแนบแบบ ภ.พ.30 ตามมาตรา 87(2))")
                    : isWhtType
                      ? "ทะเบียนรายการภาษีเงินได้หัก ณ ที่จ่าย"
                      : "ทะเบียนรายการใบกำกับภาษี"}
              </h2>
              <p className="text-xs text-muted-foreground">
                ประจำเดือน {monthNamesTh[selectedMonth - 1]} พ.ศ. {selectedYear + 543} (จำนวน {filteredRecords.length} รายการ)
              </p>
            </div>

            <div className="flex items-center gap-2 print:hidden">
              <Button
                size="sm"
                variant="outline"
                onClick={() => window.print()}
                className="gap-1.5 text-xs"
              >
                <Printer className="h-3.5 w-3.5" />
                {tr("tax_print_annex", "พิมพ์ใบแนบ")}
              </Button>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="border-b bg-muted/60 text-xs font-semibold text-muted-foreground uppercase">
                <tr>
                  <th className="px-3 py-3 w-12 text-center">ลำดับ</th>
                  <th className="px-3 py-3 w-28">วันที่</th>
                  <th className="px-3 py-3 w-36">เลขที่เอกสาร/ใบกำกับ</th>
                  <th className="px-4 py-3">ชื่อผู้ซื้อ/ผู้ขาย/ผู้รับเงิน</th>
                  <th className="px-3 py-3 w-36">เลขประจำตัว 13 หลัก</th>
                  <th className="px-3 py-3 w-20 text-center">สาขา</th>
                  <th className="px-4 py-3 text-right">มูลค่าก่อนภาษี</th>
                  <th className="px-4 py-3 text-right">ภาษี</th>
                  <th className="px-4 py-3 text-right">ยอดสุทธิ</th>
                  <th className="px-3 py-3 w-20 text-center print:hidden">จัดการ</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredRecords.map((item, idx) => (
                  <tr
                    key={item.id}
                    className="hover:bg-muted/30 transition-colors cursor-pointer"
                    onClick={() => {
                      setSelectedRecord(item);
                      if (isWhtType) {
                        const matchedWht = whtRows.find((w) => `wht-${w.journalid}` === item.id);
                        if (matchedWht) setSelectedWhtRow(matchedWht);
                      }
                    }}
                  >
                    <td className="px-3 py-2.5 text-center text-muted-foreground font-mono">{idx + 1}</td>
                    <td className="px-3 py-2.5 font-mono text-xs">{item.docdate}</td>
                    <td className="px-3 py-2.5 font-medium text-primary">{item.taxinvoiceno}</td>
                    <td className="px-4 py-2.5">
                      <div className="font-medium text-foreground">{item.counterpartyname}</div>
                      {item.incometype && (
                        <div className="text-xs text-muted-foreground">{item.incometype}</div>
                      )}
                    </td>
                    <td className="px-3 py-2.5 font-mono text-xs text-muted-foreground">{item.taxid}</td>
                    <td className="px-3 py-2.5 text-center">
                      <span className="rounded bg-muted px-1.5 py-0.5 text-xs font-mono">
                        {item.branchno}
                      </span>
                    </td>
                    <td className="px-4 py-2.5 text-right font-mono font-medium">
                      {money(item.amountbeforevat)}
                    </td>
                    <td className="px-4 py-2.5 text-right font-mono font-medium text-primary">
                      {money(item.whtamount ?? item.vatamount)}
                    </td>
                    <td className="px-4 py-2.5 text-right font-mono font-bold text-foreground">
                      {money(item.totalamount)}
                    </td>
                    <td className="px-3 py-2.5 text-center print:hidden">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedRecord(item);
                          if (isWhtType) {
                            const matchedWht = whtRows.find((w) => `wht-${w.journalid}` === item.id);
                            if (matchedWht) setSelectedWhtRow(matchedWht);
                          }
                        }}
                        className="h-8 px-2 text-xs"
                      >
                        <FileText className="h-4 w-4" />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
              <tfoot className="border-t-2 bg-muted/40 font-bold">
                <tr>
                  <td colSpan={6} className="px-4 py-3 text-right">{tr("tax_period_total", "รวมทั้งงวด")} ({totals.count} รายการ):</td>
                  <td className="px-4 py-3 text-right font-mono">{money(totals.beforeVat)}</td>
                  <td className="px-4 py-3 text-right font-mono text-primary">
                    {money(totals.tax)}
                  </td>
                  <td className="px-4 py-3 text-right font-mono">{money(totals.total)}</td>
                  <td className="print:hidden"></td>
                </tr>
              </tfoot>
            </table>
          </div>
        </Card>
      )}

      {/* Modal: รายละเอียดเอกสารภาษี */}
      {selectedRecord && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 print:hidden">
          <Card className="w-full max-w-xl shadow-2xl">
            <div className="flex items-center justify-between border-b p-4">
              <div className="flex items-center gap-2">
                <FileText className="h-5 w-5 text-primary" />
                <h3 className="font-bold text-foreground">
                  รายละเอียดเอกสาร: {selectedRecord.taxinvoiceno}
                </h3>
              </div>
              <Button size="sm" variant="ghost" onClick={() => setSelectedRecord(null)}>
                ✕
              </Button>
            </div>
            <CardContent className="space-y-3 p-5 text-sm">
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <span className="text-xs text-muted-foreground">วันที่เอกสาร:</span>
                  <p className="font-medium">{selectedRecord.docdate}</p>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground">เลขที่เอกสาร:</span>
                  <p className="font-medium text-primary">{selectedRecord.taxinvoiceno}</p>
                </div>
                <div className="col-span-2">
                  <span className="text-xs text-muted-foreground">ชื่อคู่ค้า / ผู้เสียภาษี:</span>
                  <p className="font-medium">{selectedRecord.counterpartyname}</p>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground">เลขประจำตัว 13 หลัก:</span>
                  <p className="font-mono font-medium">{selectedRecord.taxid}</p>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground">สาขา:</span>
                  <p className="font-mono">{selectedRecord.branchno} {selectedRecord.isheadoffice ? "(สำนักงานใหญ่)" : ""}</p>
                </div>
                {selectedRecord.incometype && (
                  <div className="col-span-2">
                    <span className="text-xs text-muted-foreground">ประเภทเงินได้:</span>
                    <p className="font-medium text-amber-600 dark:text-amber-400">{selectedRecord.incometype}</p>
                  </div>
                )}
              </div>

              <div className="rounded-xl border border-border bg-muted/40 p-3 space-y-1.5 font-mono">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">มูลค่าก่อนภาษี:</span>
                  <span>{money(selectedRecord.amountbeforevat)} บาท</span>
                </div>
                <div className="flex justify-between text-primary font-semibold">
                  <span>{selectedRecord.whtamount ? "ภาษีหัก ณ ที่จ่าย:" : "ภาษีมูลค่าเพิ่ม 7%:"}</span>
                  <span>{money(selectedRecord.whtamount ?? selectedRecord.vatamount)} บาท</span>
                </div>
                <div className="flex justify-between border-t pt-1 text-base font-bold">
                  <span>ยอดสุทธิ:</span>
                  <span>{money(selectedRecord.totalamount)} บาท</span>
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <Button variant="outline" onClick={() => setSelectedRecord(null)}>
                  ปิดหน้าต่าง
                </Button>
                <Button variant="default" onClick={() => window.print()} className="gap-1.5">
                  <Printer className="h-4 w-4" />
                  พิมพ์เอกสารรับรอง
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

    </div>
  );
}
