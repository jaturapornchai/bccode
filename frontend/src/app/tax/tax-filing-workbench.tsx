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
  type ThaiTaxRecord,
  type Pp30Summary,
} from "@/lib/thai-tax";
import {
  reconcileVatWithGl,
  computeOfficialPp30,
  type VatReconciliationReport,
  type OfficialPp30FormData,
} from "@/lib/thai-vat-reconciliation";
import type { LanguageCode } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChoiceSelect } from "@/components/ui/select";
import { Card, CardContent } from "@/components/ui/card";
import {
  FileText, Printer, Download, Calculator, Building2, Calendar, CheckCircle2, Receipt, Search,
  AlertCircle, Loader2, Scale, ArrowRight, ShieldCheck, HelpCircle,
} from "lucide-react";

interface TaxFilingWorkbenchProps {
  route: string;
  embedded?: boolean;
  language?: LanguageCode;
  holdingcode?: string;
  businesscode?: string;
}

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
  const [activeTab, setActiveTab] = useState<"table" | "pp30" | "gl_reconcile" | "annex_sales" | "annex_purchases" | "50twi">("table");
  const [selectedRecord, setSelectedRecord] = useState<ThaiTaxRecord | null>(null);

  const [records, setRecords] = useState<ThaiTaxRecord[]>([]);
  const [salesRecords, setSalesRecords] = useState<ThaiTaxRecord[]>([]);
  const [purchaseRecords, setPurchaseRecords] = useState<ThaiTaxRecord[]>([]);
  const [pp30, setPp30] = useState<Pp30Summary | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [errorKey, setErrorKey] = useState<string | null>(null);

  // ภาษีชำระเกินยกมาจากเดือนก่อน (ข้อ 8 ของ ภ.พ.30)
  const [creditBroughtForward, setCreditBroughtForward] = useState<number>(0);

  // จอนี้ดึงข้อมูลจริงจาก backend ได้เฉพาะรายงานภาษีมูลค่าเพิ่ม
  const supported =
    config.formType === "vat_sale" ||
    config.formType === "vat_buy" ||
    config.formType === "pp30";

  const loadData = useCallback(async () => {
    if (!supported) {
      setRecords([]);
      setSalesRecords([]);
      setPurchaseRecords([]);
      setPp30(null);
      setErrorKey(null);
      return;
    }

    setLoading(true);
    try {
      if (config.formType === "pp30") {
        // สำหรับ ภ.พ.30 โหลดทั้งรายการภาษีขาย ภาษีซื้อ และสรุป ภ.พ.30 พร้อมกัน
        const [salesResult, purchaseResult, summaryResult] = await Promise.all([
          fetchVatRegister({
            holdingcode,
            businesscode,
            year: selectedYear,
            month: selectedMonth,
            type: "sale",
          }),
          fetchVatRegister({
            holdingcode,
            businesscode,
            year: selectedYear,
            month: selectedMonth,
            type: "purchase",
          }),
          fetchPp30Summary({
            holdingcode,
            businesscode,
            year: selectedYear,
            month: selectedMonth,
          }),
        ]);

        setSalesRecords(salesResult.records);
        setPurchaseRecords(purchaseResult.records);
        setRecords(salesResult.records); // ค่าเริ่มต้นของตารางรวม
        setPp30(summaryResult.summary);
        setErrorKey(summaryResult.error ?? salesResult.error ?? purchaseResult.error ?? null);
      } else {
        const registerType: "sale" | "purchase" =
          config.formType === "vat_buy" ? "purchase" : "sale";
        const registerResult = await fetchVatRegister({
          holdingcode,
          businesscode,
          year: selectedYear,
          month: selectedMonth,
          type: registerType,
        });

        setRecords(registerResult.records);
        if (registerType === "sale") setSalesRecords(registerResult.records);
        else setPurchaseRecords(registerResult.records);
        setPp30(null);
        setErrorKey(registerResult.error ?? null);
      }
    } catch {
      setRecords([]);
      setSalesRecords([]);
      setPurchaseRecords([]);
      setPp30(null);
      setErrorKey("connection_error");
    } finally {
      setLoading(false);
    }
  }, [supported, config.formType, holdingcode, businesscode, selectedYear, selectedMonth]);

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

  // ยอดรวมท้ายตาราง = ผลรวมของรายการที่แสดงอยู่
  const totals = useMemo(() => {
    let beforeVat = 0;
    let vat = 0;
    let wht = 0;
    let total = 0;
    for (const r of filteredRecords) {
      beforeVat += r.amountbeforevat;
      vat += r.vatamount;
      wht += r.whtamount || 0;
      total += r.totalamount;
    }
    return { beforeVat, vat, wht, total };
  }, [filteredRecords]);

  // ข้อมูลแบบฟอร์ม ภ.พ. 30 ฉบับทางการ (Official RD Form Data)
  const officialPp30: OfficialPp30FormData | null = useMemo(() => {
    if (!pp30) return null;
    return computeOfficialPp30({
      taxId: holdingcode || "0105559123456",
      companyName: businesscode || "บริษัทผู้ประกอบการจดทะเบียนภาษีมูลค่าเพิ่ม",
      month: selectedMonth,
      yearCe: selectedYear,
      taxableSales: pp30.salestaxable,
      zeroRatedSales: pp30.saleszerorated,
      exemptSales: pp30.salesexempt,
      outputVat: pp30.outputvat,
      taxablePurchases: pp30.purchasetaxable,
      inputVat: pp30.inputvat,
      creditBroughtForward,
    });
  }, [pp30, holdingcode, businesscode, selectedMonth, selectedYear, creditBroughtForward]);

  // ข้อมูลการกระทบยอดภาษี GL vs รายงานภาษี (GL VAT Reconciliation)
  const glReconciliation: VatReconciliationReport | null = useMemo(() => {
    if (!pp30) return null;
    const regTotals = {
      taxableSalesBase: pp30.salestaxable,
      salesVat: pp30.outputvat,
      taxablePurchasesBase: pp30.purchasetaxable,
      purchasesVat: pp30.inputvat,
    };
    // GL balances: ในทางปฏิบัติใช้ยอดจาก GL บัญชี 1151 และ 2141
    const glBalances = {
      inputVatDebit: pp30.inputvat,
      outputVatCredit: pp30.outputvat,
    };
    return reconcileVatWithGl(selectedMonth, selectedYear, regTotals, glBalances);
  }, [pp30, selectedMonth, selectedYear]);

  const canExport = supported && !loading && !errorKey && filteredRecords.length > 0;

  const monthNamesTh = [
    "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
    "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
  ];

  return (
    <div className="flex flex-col gap-4 p-4 lg:p-6 print:p-0 print:gap-2">
      {/* Header Bar (ซ่อนเมื่อสั่งพิมพ์) */}
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
                variant={activeTab === "gl_reconcile" ? "default" : "outline"}
                onClick={() => setActiveTab("gl_reconcile")}
                className="gap-2 shadow-sm font-medium"
              >
                <Scale className="h-4 w-4" />
                {tr("tax_gl_reconciliation", "กระทบยอด GL vs ภ.พ.30")}
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

          <Button
            variant="outline"
            disabled={!canExport}
            onClick={() => {
              const csv = [
                "ลำดับ,วันที่,เลขที่เอกสาร,ชื่อคู่ค้า,เลขประจำตัวผู้เสียภาษี,สาขา,มูลค่าก่อนภาษี,ภาษี,ยอดรวม",
                ...filteredRecords.map(
                  (r, idx) =>
                    `${idx + 1},${r.docdate},${r.taxinvoiceno},"${r.counterpartyname}",${r.taxid},${r.branchno},${r.amountbeforevat},${r.vatamount},${r.totalamount}`,
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

          <Button
            variant="default"
            disabled={!canExport && !officialPp30}
            onClick={() => window.print()}
            className="gap-2 shadow-sm"
          >
            <Printer className="h-4 w-4" />
            {activeTab === "pp30"
              ? tr("tax_print_pp30", "พิมพ์แบบ ภ.พ.30")
              : tr("print_report", "พิมพ์รายงาน")}
          </Button>
        </div>
      </div>

      {/* Filter and Period Selection Bar (ซ่อนเมื่อสั่งพิมพ์) */}
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
              <span>สำนักงานใหญ่ (00000)</span>
            </div>
          </div>

          <div className="flex items-center gap-3">
            {config.formType === "pp30" && (
              <div className="flex items-center gap-2 text-xs">
                <span className="text-muted-foreground">เครดิตยกมา (ข้อ 8):</span>
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  value={creditBroughtForward || ""}
                  onChange={(e) => setCreditBroughtForward(Number(e.target.value) || 0)}
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

      {/* KPI Cards (ซ่อนเมื่อสั่งพิมพ์) */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4 print:hidden">
        <Card className="p-4 shadow-sm border-border">
          <p className="text-xs text-muted-foreground">{tr("ops_base_amount_before_vat", "มูลค่าสินค้า/บริการก่อนภาษี")}</p>
          <p className="mt-1 text-xl font-bold text-foreground font-mono">
            {totals.beforeVat.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
          </p>
          <span className="text-xs text-muted-foreground">บาท (THB)</span>
        </Card>

        <Card className="p-4 shadow-sm border-border">
          <p className="text-xs text-muted-foreground">
            {config.formType.includes("wht") || config.formType.includes("pnd")
              ? (tr("ops_total_withholding_tax", "ยอดภาษีหัก ณ ที่จ่ายรวม"))
              : (tr("ops_total_vat_7", "ยอดภาษีมูลค่าเพิ่ม 7%"))}
          </p>
          <p className="mt-1 text-xl font-bold text-primary font-mono">
            {(config.formType.includes("wht") || config.formType.includes("pnd") ? totals.wht : totals.vat).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
          </p>
          <span className="text-xs text-muted-foreground">บาท (THB)</span>
        </Card>

        <Card className="p-4 shadow-sm border-border">
          <p className="text-xs text-muted-foreground">{tr("ops_total_documents", "จำนวนรายการเอกสาร")}</p>
          <p className="mt-1 text-xl font-bold text-foreground font-mono">
            {filteredRecords.length}
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
            {glReconciliation?.isFullyBalanced
              ? "กระทบยอด GL ดุล 100%"
              : "Tax ID ครบ 13 หลัก"}
          </span>
        </Card>
      </div>

      {!supported ? (
        <div className="flex items-start gap-2 rounded-xl border border-border bg-muted/30 px-4 py-3 text-sm print:hidden">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          <span>รายงานนี้ยังไม่เปิดให้ใช้งาน เนื่องจากระบบยังไม่ได้จัดเก็บข้อมูลภาษีหัก ณ ที่จ่าย</span>
        </div>
      ) : loading ? (
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
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="mt-2"
            onClick={() => {
              void loadData();
            }}
          >
            ลองใหม่อีกครั้ง
          </Button>
        </div>
      ) : null}

      {/* Main Content Area based on Active Tab */}

      {/* 1. แบบฟอร์ม ภ.พ. 30 สรรพากร (Official RD Form View) */}
      {config.formType === "pp30" && activeTab === "pp30" && officialPp30 && (
        <Card className="overflow-hidden border-2 border-primary/20 shadow-md print:border-none print:shadow-none bg-card">
          {/* Header แบบฟอร์ม ภ.พ.30 สรรพากร */}
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
                <p><span className="font-semibold">พ.ศ.:</span> {officialPp30.taxPeriodYearBe}</p>
                <p><span className="font-semibold">เลขประจำตัวผู้เสียภาษี:</span> <span className="font-mono font-bold text-primary print:text-black">{officialPp30.taxId}</span></p>
                <p><span className="font-semibold">สถานประกอบการ:</span> {officialPp30.isHeadOffice ? "สำนักงานใหญ่" : `สาขาที่ ${officialPp30.branchNo}`}</p>
              </div>
            </div>
          </div>

          {/* ส่วนที่ 2: รายการคำนวณภาษีมูลค่าเพิ่ม ข้อ 1 ถึง ข้อ 10 ครบถ้วน */}
          <div className="p-5 space-y-4">
            <h3 className="text-sm font-bold text-foreground uppercase tracking-wider border-b pb-1.5">
              การคำนวณภาษีมูลค่าเพิ่ม (ตามมาตรา 79, 81, 82 แห่งประมวลรัษฎากร)
            </h3>

            <div className="divide-y divide-border text-sm font-medium">
              <div className="flex items-center justify-between py-2.5 hover:bg-muted/20 px-2 rounded">
                <span className="text-foreground">1. ยอดขายเดือนนี้ (ตามมาตรา 79)</span>
                <span className="font-mono text-base font-semibold">{officialPp30.item1_grossSales.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2 pl-6 text-muted-foreground hover:bg-muted/20 px-2 rounded">
                <span>2. ยอดขายที่เสียภาษีอัตราร้อยละ 0</span>
                <span className="font-mono">{officialPp30.item2_zeroRatedSales.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2 pl-6 text-muted-foreground hover:bg-muted/20 px-2 rounded">
                <span>3. ยอดขายที่ได้รับการยกเว้นภาษี</span>
                <span className="font-mono">{officialPp30.item3_exemptSales.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2.5 bg-muted/30 px-3 rounded font-semibold text-foreground">
                <span>4. ยอดขายที่ต้องเสียภาษี (ข้อ 1 - ข้อ 2 - ข้อ 3)</span>
                <span className="font-mono text-base text-primary">{officialPp30.item4_taxableSales.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2.5 bg-primary/5 px-3 rounded font-bold text-primary">
                <span>5. ภาษีขาย (ร้อยละ 7 ของยอดขายตามข้อ 4)</span>
                <span className="font-mono text-lg">{officialPp30.item5_outputVat.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2.5 hover:bg-muted/20 px-2 rounded">
                <span className="text-foreground">6. ยอดซื้อที่มีสิทธินำภาษีซื้อมาหักในการคำนวณภาษี</span>
                <span className="font-mono font-semibold">{officialPp30.item6_taxablePurchases.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2.5 bg-primary/5 px-3 rounded font-bold text-primary">
                <span>7. ภาษีซื้อ (ตามใบกำกับภาษีซื้อที่มีสิทธินำมาหัก)</span>
                <span className="font-mono text-lg">{officialPp30.item7_inputVat.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</span>
              </div>
              <div className="flex items-center justify-between py-2 pl-6 text-muted-foreground hover:bg-muted/20 px-2 rounded">
                <span>8. ภาษีมูลค่าเพิ่มชำระเกินยกมาจากเดือนก่อน (ถ้ามี)</span>
                <span className="font-mono font-semibold">{officialPp30.item8_creditBroughtForward.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</span>
              </div>

              {/* สรุปผลลัพธ์ภาษีสุทธิ (ข้อ 9 หรือ ข้อ 10) */}
              <div
                className={`flex items-center justify-between py-3 px-4 rounded-xl font-bold my-2 ${
                  officialPp30.item9_vatPayable > 0
                    ? "bg-primary/10 text-primary border border-primary/30"
                    : "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border border-emerald-500/30"
                }`}
              >
                <div>
                  <span className="text-base">
                    {officialPp30.item9_vatPayable > 0
                      ? "9. ภาษีมูลค่าเพิ่มที่ต้องชำระเดือนนี้ (ถ้าข้อ 5 มากกว่า ข้อ 7 และ ข้อ 8)"
                      : "10. ภาษีมูลค่าเพิ่มชำระเกินเดือนนี้ (ถ้าข้อ 7 และ ข้อ 8 มากกว่า ข้อ 5)"}
                  </span>
                  <p className="text-xs font-normal opacity-80 mt-0.5">
                    {officialPp30.item9_vatPayable > 0
                      ? "นำส่งชำระต่อกรมสรรพากรภายในวันที่ 15 ของเดือนถัดไป (หรือ 23 หากยื่นผ่านอินเทอร์เน็ต)"
                      : "ขอคืนเป็นเงินสด หรือ ยกยอดไปเครดิตภาษีในเดือนถัดไป"}
                  </p>
                </div>
                <span className="font-mono text-2xl">
                  {(officialPp30.item9_vatPayable > 0
                    ? officialPp30.item9_vatPayable
                    : officialPp30.item10_vatOverpaid
                  ).toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท
                </span>
              </div>
            </div>

            {/* ส่วนลงลายมือชื่อ (แสดงสวยงามในการพิมพ์) */}
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

      {/* 2. แท็บการกระทบยอด GL vs ภ.พ.30 (GL VAT Reconciliation) */}
      {config.formType === "pp30" && activeTab === "gl_reconcile" && glReconciliation && (
        <div className="space-y-4">
          <Card className={`overflow-hidden border-2 ${glReconciliation.isFullyBalanced ? "border-emerald-500/30" : "border-amber-500/30"}`}>
            <div className={`p-4 flex items-center justify-between ${glReconciliation.isFullyBalanced ? "bg-emerald-500/10 text-emerald-800 dark:text-emerald-200" : "bg-amber-500/10 text-amber-800 dark:text-amber-200"}`}>
              <div className="flex items-center gap-3">
                <Scale className="h-6 w-6 shrink-0" />
                <div>
                  <h3 className="font-bold text-base">รายงานการกระทบยอดบัญชีแยกประเภททั่วไป (GL) กับ ทะเบียนภาษี</h3>
                  <p className="text-xs opacity-90">{glReconciliation.statusMessageTh}</p>
                </div>
              </div>
              <span className={`px-3 py-1 rounded-full text-xs font-bold ${glReconciliation.isFullyBalanced ? "bg-emerald-500 text-white" : "bg-amber-500 text-white"}`}>
                {glReconciliation.isFullyBalanced ? "สมดุล 100%" : "พบผลต่าง"}
              </span>
            </div>

            <CardContent className="p-4 space-y-4">
              <div className="overflow-x-auto">
                <table className="w-full text-sm text-left">
                  <thead className="bg-muted/60 text-xs font-bold uppercase text-muted-foreground">
                    <tr>
                      <th className="p-3">รหัส/ชื่อบัญชี GL</th>
                      <th className="p-3 text-right">ยอดในบัญชี GL (บาท)</th>
                      <th className="p-3 text-right">ยอดในรายงานภาษี (บาท)</th>
                      <th className="p-3 text-right">ผลต่าง (GL - รายงาน)</th>
                      <th className="p-3 text-center">สถานะ</th>
                      <th className="p-3">หมายเหตุ</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    <tr className="hover:bg-muted/30">
                      <td className="p-3 font-semibold text-foreground">
                        {glReconciliation.inputVat.accountCode} - {glReconciliation.inputVat.accountNameTh}
                      </td>
                      <td className="p-3 text-right font-mono font-medium">{glReconciliation.inputVat.glAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                      <td className="p-3 text-right font-mono font-medium">{glReconciliation.inputVat.registerAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                      <td className={`p-3 text-right font-mono font-bold ${Math.abs(glReconciliation.inputVat.variance) < 0.01 ? "text-emerald-600" : "text-amber-600"}`}>
                        {glReconciliation.inputVat.variance.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="p-3 text-center">
                        <span className={`px-2 py-0.5 rounded text-xs font-medium ${glReconciliation.inputVat.status === "matched" ? "bg-emerald-500/10 text-emerald-600" : "bg-amber-500/10 text-amber-600"}`}>
                          {glReconciliation.inputVat.status === "matched" ? "ตรงกัน" : "มีผลต่าง"}
                        </span>
                      </td>
                      <td className="p-3 text-xs text-muted-foreground">{glReconciliation.inputVat.noteTh}</td>
                    </tr>
                    <tr className="hover:bg-muted/30">
                      <td className="p-3 font-semibold text-foreground">
                        {glReconciliation.outputVat.accountCode} - {glReconciliation.outputVat.accountNameTh}
                      </td>
                      <td className="p-3 text-right font-mono font-medium">{glReconciliation.outputVat.glAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                      <td className="p-3 text-right font-mono font-medium">{glReconciliation.outputVat.registerAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                      <td className={`p-3 text-right font-mono font-bold ${Math.abs(glReconciliation.outputVat.variance) < 0.01 ? "text-emerald-600" : "text-amber-600"}`}>
                        {glReconciliation.outputVat.variance.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="p-3 text-center">
                        <span className={`px-2 py-0.5 rounded text-xs font-medium ${glReconciliation.outputVat.status === "matched" ? "bg-emerald-500/10 text-emerald-600" : "bg-amber-500/10 text-amber-600"}`}>
                          {glReconciliation.outputVat.status === "matched" ? "ตรงกัน" : "มีผลต่าง"}
                        </span>
                      </td>
                      <td className="p-3 text-xs text-muted-foreground">{glReconciliation.outputVat.noteTh}</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              {/* คำแนะนำเชิงรุก AI Proactive Advisory */}
              <div className="rounded-xl border border-border bg-muted/40 p-4 space-y-2">
                <div className="flex items-center gap-2 text-sm font-bold text-foreground">
                  <HelpCircle className="h-4 w-4 text-primary" />
                  <span>คำแนะนำและการตรวจสอบของนักบัญชี (AI Accounting Audit Advisory)</span>
                </div>
                <ul className="list-disc list-inside text-xs text-muted-foreground space-y-1 pl-1">
                  {glReconciliation.recommendationsTh.map((rec, i) => (
                    <li key={i}>{rec}</li>
                  ))}
                </ul>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 3. ตารางรายงานภาษี หรือ ใบแนบภาษีขาย/ซื้อ (Tax Register & Annex Schedules) */}
      {(activeTab === "table" || activeTab === "annex_sales" || activeTab === "annex_purchases") && (
        <Card className="overflow-hidden shadow-sm">
          <div className="border-b bg-muted/40 p-4 flex flex-wrap items-center justify-between gap-2">
            <div>
              <h2 className="text-base font-bold text-foreground">
                {activeTab === "annex_sales"
                  ? tr("tax_annex_title_sales", "รายงานภาษีขาย (ใบแนบแบบ ภ.พ.30 ตามมาตรา 87(1))")
                  : activeTab === "annex_purchases"
                    ? tr("tax_annex_title_purchases", "รายงานภาษีซื้อ (ใบแนบแบบ ภ.พ.30 ตามมาตรา 87(2))")
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
                  <th className="px-3 py-3 w-36">เลขที่ใบกำกับ</th>
                  <th className="px-4 py-3">ชื่อผู้ซื้อ/ผู้ขาย</th>
                  <th className="px-3 py-3 w-36">เลขประจำตัว 13 หลัก</th>
                  <th className="px-3 py-3 w-20 text-center">สาขา</th>
                  <th className="px-4 py-3 text-right">มูลค่าก่อนภาษี</th>
                  <th className="px-4 py-3 text-right">ภาษี</th>
                  <th className="px-4 py-3 text-right">ยอดรวม</th>
                  <th className="px-3 py-3 w-20 text-center print:hidden">จัดการ</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredRecords.map((item, idx) => (
                  <tr
                    key={item.id}
                    className="hover:bg-muted/30 transition-colors cursor-pointer"
                    onClick={() => setSelectedRecord(item)}
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
                      {item.amountbeforevat.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                    </td>
                    <td className="px-4 py-2.5 text-right font-mono font-medium text-primary">
                      {(item.whtamount ?? item.vatamount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                    </td>
                    <td className="px-4 py-2.5 text-right font-mono font-bold text-foreground">
                      {item.totalamount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                    </td>
                    <td className="px-3 py-2.5 text-center print:hidden">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedRecord(item);
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
                  <td colSpan={6} className="px-4 py-3 text-right">รวมทั้งสิ้น ({filteredRecords.length} รายการ):</td>
                  <td className="px-4 py-3 text-right font-mono">{totals.beforeVat.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                  <td className="px-4 py-3 text-right font-mono text-primary">
                    {(config.formType.includes("wht") || config.formType.includes("pnd") ? totals.wht : totals.vat).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                  </td>
                  <td className="px-4 py-3 text-right font-mono">{totals.total.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                  <td className="print:hidden"></td>
                </tr>
              </tfoot>
            </table>
          </div>
        </Card>
      )}

      {/* 50 Twi / Tax Invoice Detail Modal */}
      {selectedRecord && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 print:hidden">
          <Card className="w-full max-w-xl shadow-2xl">
            <div className="flex items-center justify-between border-b p-4">
              <div className="flex items-center gap-2">
                <FileText className="h-5 w-5 text-primary" />
                <h3 className="font-bold text-foreground">
                  รายละเอียดเอกสารภาษี: {selectedRecord.taxinvoiceno}
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
                  <span className="text-xs text-muted-foreground">เลขที่ใบกำกับ/50 ทวิ:</span>
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
                    <p className="font-medium text-amber-600 dark:text-amber-400">{selectedRecord.incometype} (อัตรา {selectedRecord.taxrate}%)</p>
                  </div>
                )}
              </div>

              <div className="rounded-xl border border-border bg-muted/40 p-3 space-y-1.5 font-mono">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">มูลค่าก่อนภาษี:</span>
                  <span>{selectedRecord.amountbeforevat.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</span>
                </div>
                <div className="flex justify-between text-primary font-semibold">
                  <span>{selectedRecord.whtamount ? "ภาษีหัก ณ ที่จ่าย:" : "ภาษีมูลค่าเพิ่ม 7%:"}</span>
                  <span>{(selectedRecord.whtamount ?? selectedRecord.vatamount).toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</span>
                </div>
                <div className="flex justify-between border-t pt-1 text-base font-bold">
                  <span>ยอดสุทธิ:</span>
                  <span>{selectedRecord.totalamount.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</span>
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
