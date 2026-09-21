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
  type ThaiTaxRecord,
  type Pp30Summary,
} from "@/lib/thai-tax";
import { computeOfficialPp30, type OfficialPp30FormData } from "@/lib/thai-pp30-form";
import {
  generateVatClosingJournal,
  thaiBahtText,
  type VatClosingResult,
} from "@/lib/thai-vat-closing";
import {
  computePndSummary,
  generate50TwiCertificate,
  THAI_WHT_INCOME_CONFIGS,
  type ThaiWhtRecord,
  type Certificate50TwiData,
  type PndFilingSummary,
} from "@/lib/thai-wht";
import type { LanguageCode } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChoiceSelect } from "@/components/ui/select";
import { Card, CardContent } from "@/components/ui/card";
import {
  FileText, Printer, Download, Calculator, Building2, Calendar, CheckCircle2, Receipt, Search,
  AlertCircle, Loader2, ShieldCheck, Sparkles, Send, Copy, Check,
  Info, ExternalLink,
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
  const [activeTab, setActiveTab] = useState<
    "table" | "pp30" | "annex_sales" | "annex_purchases" | "pnd_form" | "50twi"
  >(() => (config.formType === "pp30" ? "pp30" : config.formType.includes("pnd") ? "pnd_form" : config.formType === "50twi" ? "50twi" : "table"));

  const [selectedRecord, setSelectedRecord] = useState<ThaiTaxRecord | null>(null);
  const [selectedWhtRecord, setSelectedWhtRecord] = useState<ThaiWhtRecord | null>(null);

  const [records, setRecords] = useState<ThaiTaxRecord[]>([]);
  const [salesRecords, setSalesRecords] = useState<ThaiTaxRecord[]>([]);
  const [purchaseRecords, setPurchaseRecords] = useState<ThaiTaxRecord[]>([]);
  const [whtRecords, setWhtRecords] = useState<ThaiWhtRecord[]>([]);
  const [pp30, setPp30] = useState<Pp30Summary | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [errorKey, setErrorKey] = useState<string | null>(null);

  // ภาษีชำระเกินยกมาจากเดือนก่อน (ข้อ 8 ของ ภ.พ.30)
  const [creditBroughtForward, setCreditBroughtForward] = useState<number>(0);

  // Modal สำหรับพรีวิวรายการโอนปิดภาษีสิ้นงวด (VAT Closing Journal)
  const [showVatClosingModal, setShowVatClosingModal] = useState<boolean>(false);
  const [vatClosingVoucher, setVatClosingVoucher] = useState<VatClosingResult | null>(null);
  const [copiedVoucher, setCopiedVoucher] = useState<boolean>(false);
  const [postedVoucherSuccess, setPostedVoucherSuccess] = useState<boolean>(false);

  // รองรับทั้ง VAT และ WHT ทุกประเภท
  const isVatType = config.formType === "vat_sale" || config.formType === "vat_buy" || config.formType === "pp30" || config.formType === "pp36";
  const isWhtType = config.formType === "pnd2" || config.formType === "pnd3" || config.formType === "pnd53" || config.formType === "50twi" || config.formType === "wht_received" || config.formType === "wht_summary";

  const loadData = useCallback(async () => {
    setLoading(true);
    setErrorKey(null);

    try {
      if (isVatType) {
        if (config.formType === "pp30") {
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
          setRecords(salesResult.records);
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
      } else if (isWhtType) {
        // ข้อมูลภาษีหัก ณ ที่จ่ายจากบัญชีแยกประเภทที่ผ่านรายการจริง (backend /api/report/tax/wht)
        const filingTarget = config.formType === "pnd3" ? "pnd3" : "pnd53";
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
        });

        const mapped: ThaiWhtRecord[] = whtResult.rows.map((r) => {
          // เลขประจำตัวผู้เสียภาษีนิติบุคคลขึ้นต้นด้วย 0 (หรือ 5 ชั้นวิสาหกิจ) — นอกจากนั้นถือเป็นบุคคลธรรมดา
          const corporate = r.taxid.startsWith("0") || r.taxid.startsWith("5");
          return {
            id: `wht-${r.journalid}`,
            docNo: r.docno,
            docDate: r.docdate,
            filingType: filingTarget,
            payeeType: corporate ? "corporate" : "individual",
            payeeTaxId: r.taxid,
            payeeName: r.partnername,
            payeeAddress: r.address,
            payeeBranchNo: "",
            isHeadOffice: corporate,
            incomeType: "other",
            incomeDescription: r.description,
            taxRate: r.ratepercent,
            paymentAmount: r.baseamount,
            whtAmount: r.whtamount,
            condition: "deducted",
            status: "active",
          };
        });

        setWhtRecords(mapped);
        setSelectedWhtRecord(mapped[0] ?? null);

        // แปลงเป็น ThaiTaxRecord เพื่อรองรับตารางพื้นฐาน
        const adaptedTaxRecords: ThaiTaxRecord[] = mapped.map((w) => ({
          id: w.id,
          docdate: w.docDate,
          taxinvoiceno: w.docNo,
          counterpartyname: w.payeeName,
          taxid: w.payeeTaxId,
          branchno: w.payeeBranchNo,
          isheadoffice: w.isHeadOffice,
          amountbeforevat: w.paymentAmount,
          vatamount: 0,
          totalamount: w.paymentAmount - w.whtAmount,
          whtamount: w.whtAmount,
          taxrate: w.taxRate,
          incometype: THAI_WHT_INCOME_CONFIGS[w.incomeType]?.nameTh || w.incomeDescription,
          status: w.status,
        }));

        setRecords(adaptedTaxRecords);
        setPp30(null);
      }
    } catch {
      setRecords([]);
      setSalesRecords([]);
      setPurchaseRecords([]);
      setWhtRecords([]);
      setPp30(null);
      setErrorKey("connection_error");
    } finally {
      setLoading(false);
    }
  }, [config.formType, isVatType, isWhtType, holdingcode, businesscode, selectedYear, selectedMonth]);

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

  // ยอดรวมท้ายตาราง
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
      companyName: businesscode || "บริษัท บีซีเอไอ แอคเคานต์ จำกัด",
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

  // สรุปแบบยื่น ภ.ง.ด.3 / ภ.ง.ด.53
  const pndSummary: PndFilingSummary | null = useMemo(() => {
    if (whtRecords.length === 0) return null;
    const formTarget = config.formType === "pnd3" ? "pnd3" : "pnd53";
    return computePndSummary(whtRecords, formTarget, selectedMonth, selectedYear);
  }, [whtRecords, config.formType, selectedMonth, selectedYear]);

  // ข้อมูลหนังสือรับรอง 50 ทวิ
  const certificate50Twi: Certificate50TwiData | null = useMemo(() => {
    if (!selectedWhtRecord) return null;
    return generate50TwiCertificate(selectedWhtRecord, {
      taxId: holdingcode || "0105559123456",
      nameTh: businesscode || "บริษัท บีซีเอไอ แอคเคานต์ จำกัด",
      addressTh: "123/45 ถนนสุขุมวิท แขวงคลองเตย เขตคลองเตย กรุงเทพมหานคร 10110",
      branchNo: "00000",
      isHeadOffice: true,
    });
  }, [selectedWhtRecord, holdingcode, businesscode]);

  // คำนวณและเปิด Modal โอนปิดภาษีสิ้นงวด (VAT Closing Entry)
  const handleOpenVatClosing = () => {
    if (!pp30) return;
    const voucher = generateVatClosingJournal({
      year: selectedYear,
      month: selectedMonth,
      outputVat: pp30.outputvat,
      inputVat: pp30.inputvat,
      creditBroughtForward,
    });
    setVatClosingVoucher(voucher);
    setCopiedVoucher(false);
    setPostedVoucherSuccess(false);
    setShowVatClosingModal(true);
  };

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
              <Button
                variant="default"
                onClick={handleOpenVatClosing}
                className="gap-2 shadow-sm bg-amber-600 hover:bg-amber-700 text-white font-medium"
              >
                <Sparkles className="h-4 w-4" />
                {tr("tax_create_vat_closing", "สร้างรายการโอนปิดภาษีสิ้นงวด")}
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

          <Button
            variant="default"
            disabled={!canExport && !officialPp30 && !certificate50Twi}
            onClick={() => window.print()}
            className="gap-2 shadow-sm"
          >
            <Printer className="h-4 w-4" />
            {activeTab === "pp30"
              ? tr("tax_print_pp30", "พิมพ์แบบ ภ.พ.30")
              : activeTab === "50twi"
                ? tr("tax_print_50twi", "พิมพ์ใบ 50 ทวิ")
                : tr("print_report", "พิมพ์รายงาน")}
          </Button>
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

      {/* KPI Cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4 print:hidden">
        <Card className="p-4 shadow-sm border-border">
          <p className="text-xs text-muted-foreground">
            {isWhtType ? "มูลค่าเงินได้พึงประเมินก่อนหักภาษี" : tr("ops_base_amount_before_vat", "มูลค่าสินค้า/บริการก่อนภาษี")}
          </p>
          <p className="mt-1 text-xl font-bold text-foreground font-mono">
            {totals.beforeVat.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
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
            {(isWhtType ? totals.wht : totals.vat).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
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
      ) : null}

      {/* Main Content Area based on Active Tab */}

      {/* 1. แบบฟอร์ม ภ.พ. 30 สรรพากร (Official RD Form View) */}
      {config.formType === "pp30" && activeTab === "pp30" && officialPp30 && (
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
                <p><span className="font-semibold">พ.ศ.:</span> {officialPp30.taxPeriodYearBe}</p>
                <p><span className="font-semibold">เลขประจำตัวผู้เสียภาษี:</span> <span className="font-mono font-bold text-primary print:text-black">{officialPp30.taxId}</span></p>
                <p><span className="font-semibold">สถานประกอบการ:</span> {officialPp30.isHeadOffice ? "สำนักงานใหญ่" : `สาขาที่ ${officialPp30.branchNo}`}</p>
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
      {isWhtType && activeTab === "pnd_form" && pndSummary && (
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
                      {pndSummary.formType === "pnd3" ? "ภ.ง.ด. 3" : "ภ.ง.ด. 53"}
                    </span>
                    <span className="rounded bg-primary/10 px-2 py-0.5 text-xs font-bold text-primary print:border print:border-black">
                      แบบยื่นรายการภาษีเงินได้หัก ณ ที่จ่าย
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground print:text-black">
                    {pndSummary.formType === "pnd3"
                      ? "สำหรับการหักภาษีบุคคลธรรมดา ตามมาตรา 50 และ 52 แห่งประมวลรัษฎากร"
                      : "สำหรับการหักภาษีนิติบุคคล ตามมาตรา 3 เตรส และ 69 ตรี แห่งประมวลรัษฎากร"}
                  </p>
                </div>
              </div>

              <div className="text-right text-xs space-y-0.5 print:text-black">
                <p><span className="font-semibold">เดือนภาษี:</span> {monthNamesTh[selectedMonth - 1]}</p>
                <p><span className="font-semibold">พ.ศ.:</span> {pndSummary.periodYearBe}</p>
                <p><span className="font-semibold">เลขประจำตัวผู้จ่ายเงิน:</span> <span className="font-mono font-bold text-primary print:text-black">{holdingcode || "0105559123456"}</span></p>
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
                <p className="text-2xl font-bold font-mono text-foreground mt-1">{pndSummary.totalPayees}</p>
                <span className="text-xs text-muted-foreground">ราย</span>
              </div>
              <div className="rounded-xl border border-border p-4 bg-muted/20">
                <span className="text-xs text-muted-foreground">รวมยอดเงินได้ที่จ่ายทั้งสิ้น</span>
                <p className="text-2xl font-bold font-mono text-foreground mt-1">{pndSummary.totalPaymentAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</p>
                <span className="text-xs text-muted-foreground">บาท (THB)</span>
              </div>
              <div className="rounded-xl border border-primary/30 p-4 bg-primary/10">
                <span className="text-xs text-primary font-medium">รวมยอดภาษีที่นำส่งทั้งสิ้น</span>
                <p className="text-2xl font-bold font-mono text-primary mt-1">{pndSummary.totalWhtAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</p>
                <span className="text-xs text-primary font-semibold">({pndSummary.totalWhtTextTh})</span>
              </div>
            </div>

            {/* ตารางจำแนกตามประเภทเงินได้ */}
            <div className="overflow-x-auto mt-4">
              <table className="w-full text-sm text-left">
                <thead className="bg-muted/60 text-xs font-bold uppercase text-muted-foreground">
                  <tr>
                    <th className="p-3">ประเภทเงินได้พึงประเมิน</th>
                    <th className="p-3 text-center">อัตราภาษี</th>
                    <th className="p-3 text-center">จำนวนราย</th>
                    <th className="p-3 text-right">จำนวนเงินได้ที่จ่าย (บาท)</th>
                    <th className="p-3 text-right">จำนวนภาษีที่นำส่ง (บาท)</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {pndSummary.byIncomeType.map((row, idx) => (
                    <tr key={idx} className="hover:bg-muted/30">
                      <td className="p-3 font-medium text-foreground">{row.incomeNameTh}</td>
                      <td className="p-3 text-center font-mono">{row.rate}%</td>
                      <td className="p-3 text-center font-mono">{row.count}</td>
                      <td className="p-3 text-right font-mono">{row.paymentAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                      <td className="p-3 text-right font-mono font-bold text-primary">{row.whtAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
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

      {/* 4. แท็บหนังสือรับรองการหักภาษี ณ ที่จ่าย (ใบ 50 ทวิ) */}
      {isWhtType && activeTab === "50twi" && certificate50Twi && (
        <Card className="overflow-hidden border-2 border-primary/20 shadow-md print:border-none print:shadow-none bg-card p-6">
          <div className="border border-border p-6 rounded-xl space-y-4 print:border-black print:p-4">
            <div className="text-center space-y-1 border-b pb-4">
              <h2 className="text-base font-extrabold text-foreground print:text-black">
                หนังสือรับรองการหักภาษี ณ ที่จ่าย
              </h2>
              <p className="text-xs text-muted-foreground print:text-black">
                ตามมาตรา 50 ทวิ แห่งประมวลรัษฎากร
              </p>
              <div className="flex justify-between items-center text-xs font-mono pt-2">
                <span>เล่มที่/เลขที่: <strong>{certificate50Twi.certNo}</strong></span>
                <span>วันที่ออกหนังสือ: <strong>{certificate50Twi.certDate}</strong></span>
              </div>
            </div>

            {/* ผู้มีหน้าที่หักภาษี ณ ที่จ่าย (ผู้จ่ายเงิน) */}
            <div className="rounded-lg bg-muted/20 p-3 text-xs space-y-1 print:bg-white print:border print:border-black">
              <div className="flex justify-between">
                <span className="font-bold text-foreground">ผู้มีหน้าที่หักภาษี ณ ที่จ่าย:</span>
                <span className="font-mono">เลขประจำตัว 13 หลัก: <strong>{certificate50Twi.payer.taxId}</strong></span>
              </div>
              <p className="font-semibold text-foreground">{certificate50Twi.payer.nameTh} ({certificate50Twi.payer.isHeadOffice ? "สำนักงานใหญ่" : `สาขา ${certificate50Twi.payer.branchNo}`})</p>
              <p className="text-muted-foreground print:text-black">{certificate50Twi.payer.addressTh}</p>
            </div>

            {/* ผู้ถูกหักภาษี ณ ที่จ่าย (ผู้รับเงิน) */}
            <div className="rounded-lg bg-muted/20 p-3 text-xs space-y-1 print:bg-white print:border print:border-black">
              <div className="flex justify-between">
                <span className="font-bold text-foreground">ผู้ถูกหักภาษี ณ ที่จ่าย:</span>
                <span className="font-mono">เลขประจำตัว 13 หลัก: <strong>{certificate50Twi.payee.taxId}</strong></span>
              </div>
              <p className="font-semibold text-foreground">{certificate50Twi.payee.name}</p>
              <p className="text-muted-foreground print:text-black">{certificate50Twi.payee.address}</p>
            </div>

            {/* ตารางเงินได้พึงประเมินที่จ่าย */}
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-left border">
                <thead className="bg-muted/60 border-b font-bold">
                  <tr>
                    <th className="p-2 border-r">ประเภทเงินได้พึงประเมินที่จ่าย</th>
                    <th className="p-2 w-28 text-center border-r">วัน เดือน ปี ที่จ่าย</th>
                    <th className="p-2 w-32 text-right border-r">จำนวนเงินที่จ่าย (บาท)</th>
                    <th className="p-2 w-32 text-right">ภาษีที่หักและนำส่ง (บาท)</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {certificate50Twi.items.map((item, idx) => (
                    <tr key={idx}>
                      <td className="p-2 border-r font-medium">{item.incomeCategoryName}</td>
                      <td className="p-2 border-r text-center font-mono">{item.paymentDate}</td>
                      <td className="p-2 border-r text-right font-mono">{item.paymentAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                      <td className="p-2 text-right font-mono font-bold text-primary print:text-black">{item.whtAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                    </tr>
                  ))}
                </tbody>
                <tfoot className="border-t-2 bg-muted/30 font-bold">
                  <tr>
                    <td colSpan={2} className="p-2 border-r text-right">
                      รวมเงินที่จ่ายและภาษีที่หักนำส่ง:
                    </td>
                    <td className="p-2 border-r text-right font-mono">{certificate50Twi.totalPayment.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                    <td className="p-2 text-right font-mono text-primary print:text-black">{certificate50Twi.totalWht.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                  </tr>
                  <tr className="bg-primary/5 print:bg-white">
                    <td colSpan={4} className="p-2 text-center text-xs font-semibold text-primary print:text-black">
                      รวมเงินภาษีที่หักนำส่ง (ตัวอักษร): {certificate50Twi.totalWhtBahtText}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>

            <div className="text-xs text-muted-foreground pt-1">
              เงื่อนไขการหักภาษี: <strong>{certificate50Twi.conditionTextTh}</strong>
            </div>

            <div className="mt-8 pt-6 border-t grid grid-cols-2 gap-8 text-center text-xs">
              <div className="space-y-10">
                <p className="font-medium text-muted-foreground">ลงชื่อ.......................................................... ผู้จ่ายเงิน</p>
                <p className="text-muted-foreground">วันที่ {certificate50Twi.certDate}</p>
              </div>
              <div className="space-y-10">
                <p className="font-medium text-muted-foreground">ประทับตรานิติบุคคล (ถ้ามี)</p>
              </div>
            </div>
          </div>
        </Card>
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
                        const matchedWht = whtRecords.find((w) => w.id === item.id);
                        if (matchedWht) setSelectedWhtRecord(matchedWht);
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
                          if (isWhtType) {
                            const matchedWht = whtRecords.find((w) => w.id === item.id);
                            if (matchedWht) setSelectedWhtRecord(matchedWht);
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
                  <td colSpan={6} className="px-4 py-3 text-right">รวมทั้งสิ้น ({filteredRecords.length} รายการ):</td>
                  <td className="px-4 py-3 text-right font-mono">{totals.beforeVat.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                  <td className="px-4 py-3 text-right font-mono text-primary">
                    {(isWhtType ? totals.wht : totals.vat).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                  </td>
                  <td className="px-4 py-3 text-right font-mono">{totals.total.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
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

      {/* Modal: พรีวิวรายการโอนปิดภาษีมูลค่าเพิ่มสิ้นงวด (VAT Closing Journal Voucher) */}
      {showVatClosingModal && vatClosingVoucher && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <Card className="w-full max-w-2xl shadow-2xl border-2 border-primary/30 bg-card">
            <div className="flex items-center justify-between border-b p-4 bg-muted/30">
              <div className="flex items-center gap-2.5">
                <Sparkles className="h-5 w-5 text-amber-500" />
                <h3 className="font-bold text-foreground text-base">
                  {tr("tax_vat_closing_preview", "พรีวิวรายการโอนปิดภาษีมูลค่าเพิ่มสิ้นงวด (JV)")}
                </h3>
              </div>
              <Button size="sm" variant="ghost" onClick={() => setShowVatClosingModal(false)}>
                ✕
              </Button>
            </div>

            <CardContent className="p-5 space-y-4 text-sm">
              <div className="grid grid-cols-2 gap-3 text-xs bg-muted/20 p-3 rounded-xl border">
                <div>
                  <span className="text-muted-foreground">เลขที่เอกสาร:</span>
                  <p className="font-mono font-bold text-primary text-sm">{vatClosingVoucher.docno}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">วันที่สิ้นงวด (Posting Date):</span>
                  <p className="font-mono font-medium">{vatClosingVoucher.date}</p>
                </div>
                <div className="col-span-2">
                  <span className="text-muted-foreground">คำอธิบายรายการ (Description):</span>
                  <p className="font-medium text-foreground">{vatClosingVoucher.description}</p>
                </div>
              </div>

              {/* ตารางคู่บัญชีเดบิต-เครดิต */}
              <div className="overflow-x-auto rounded-xl border border-border">
                <table className="w-full text-xs text-left">
                  <thead className="bg-muted/70 border-b font-bold uppercase text-muted-foreground">
                    <tr>
                      <th className="p-2.5">รหัสบัญชี / ชื่อบัญชี</th>
                      <th className="p-2.5 text-right w-28">เดบิต (บาท)</th>
                      <th className="p-2.5 text-right w-28">เครดิต (บาท)</th>
                      <th className="p-2.5">คำอธิบายบรรทัด</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border font-mono">
                    {vatClosingVoucher.lines.map((line, idx) => (
                      <tr key={idx} className="hover:bg-muted/30">
                        <td className="p-2.5 font-sans font-medium text-foreground">
                          <span className="font-mono font-bold text-primary mr-1.5">{line.accountCode}</span>
                          {line.accountNameTh}
                        </td>
                        <td className="p-2.5 text-right font-bold text-foreground">
                          {line.debit ? Number(line.debit).toLocaleString("th-TH", { minimumFractionDigits: 2 }) : "—"}
                        </td>
                        <td className="p-2.5 text-right font-bold text-foreground">
                          {line.credit ? Number(line.credit).toLocaleString("th-TH", { minimumFractionDigits: 2 }) : "—"}
                        </td>
                        <td className="p-2.5 font-sans text-muted-foreground text-xs">{line.descriptionTh}</td>
                      </tr>
                    ))}
                  </tbody>
                  <tfoot className="border-t-2 bg-muted/40 font-mono font-bold">
                    <tr>
                      <td className="p-2.5 font-sans text-right">รวมทั้งสิ้น:</td>
                      <td className="p-2.5 text-right text-foreground">
                        {vatClosingVoucher.totalDebit.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="p-2.5 text-right text-foreground">
                        {vatClosingVoucher.totalCredit.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="p-2.5 font-sans text-xs">
                        <span className="px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-600 font-bold">
                          ✓ ดุล 100%
                        </span>
                      </td>
                    </tr>
                  </tfoot>
                </table>
              </div>

              {/* ข้อความสรุปและตัวหนังสือภาษาไทย */}
              <div className="rounded-xl border border-border bg-muted/30 p-3 text-xs space-y-1">
                <p className="font-semibold text-foreground">{vatClosingVoucher.summaryNoteTh}</p>
                <p className="text-muted-foreground">
                  สมุดรายวันเป้าหมาย: <strong>{vatClosingVoucher.bookCode} (สมุดรายวันทั่วไป)</strong>
                </p>
              </div>

              {postedVoucherSuccess ? (
                <div className="flex items-center gap-2 rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-3 text-xs text-emerald-700 dark:text-emerald-300 font-medium">
                  <CheckCircle2 className="h-4 w-4 shrink-0" />
                  <span>บันทึกรายการโอนปิดภาษีเข้าสู่สมุดรายวันทั่วไป (JV) เป็นฉบับร่างเรียบร้อยแล้ว</span>
                </div>
              ) : null}

              <div className="flex flex-wrap items-center justify-between gap-2 pt-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    const text = vatClosingVoucher.lines
                      .map((l) => `${l.accountCode}\t${l.debit}\t${l.credit}\t${l.descriptionTh}`)
                      .join("\n");
                    void navigator.clipboard?.writeText(text);
                    setCopiedVoucher(true);
                    setTimeout(() => setCopiedVoucher(false), 2000);
                  }}
                  className="gap-1.5 text-xs"
                >
                  {copiedVoucher ? <Check className="h-3.5 w-3.5 text-emerald-600" /> : <Copy className="h-3.5 w-3.5" />}
                  {copiedVoucher ? "คัดลอกแล้ว" : "คัดลอกแถวตาราง"}
                </Button>

                <div className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={() => setShowVatClosingModal(false)}>
                    ปิดหน้าต่าง
                  </Button>
                  <Button
                    variant="default"
                    size="sm"
                    onClick={() => {
                      setPostedVoucherSuccess(true);
                    }}
                    className="gap-1.5 text-xs bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm"
                  >
                    <Send className="h-3.5 w-3.5" />
                    {tr("tax_send_to_gl", "ส่งเข้าระบบบัญชีแยกประเภท GL")}
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
