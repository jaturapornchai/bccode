"use client";

import { useState, useMemo, useEffect, useCallback, useRef, useSyncExternalStore } from "react";
import {
  useBackendDictionary,
  useBackendText,
} from "@/components/backend-text-provider";
import {
  getThaiTaxConfig,
  taxText,
  fetchVatRegister,
  fetchWhtReport,
  countTaxIdIssues,
  taxCompanyLabel,
  whtRegisterRecords,
  type CompanyHeader,
  type ThaiTaxRecord,
  type VatRegisterSummary,
  type WhtReportRow,
  type WhtReportSummary,
} from "@/lib/thai-tax";
import { formatAmount } from "@/lib/general-ledger";
import { formatAppDate } from "@/lib/date-time";
import type { LanguageCode } from "@/lib/i18n";
import { WORKSPACE_CHANGED_EVENT, workspaceStorageKeys } from "@/lib/workspace-models";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChoiceSelect } from "@/components/ui/select";
import { Card, CardContent } from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { WhtCertificatePanel, WhtReceivedCertificateDetails, leaveWhtCertificate } from "./wht-certificate-panel";
import { useDirtyGuard } from "@/app/gl/gl-common";
import {
  FileText, Printer, Download, Building2, Calendar, CheckCircle2, Receipt, Search,
  AlertCircle, AlertTriangle, Loader2,
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
// ขอทั้งงวดในหน้าเดียวตามเพดาน backend (taxRegisterMaxLimit) — ยอดรวมท้ายตารางมาจาก backend ทั้งงวดเสมอ
const REGISTER_PAGE_LIMIT = 500;

// บริษัทที่เลือกใน workspace ของ session (ทะเบียน VAT ไม่ส่งหัวบริษัท) — อ่านผ่าน useSyncExternalStore ให้ SSR ได้ค่าว่างและตามการสลับบริษัท
function subscribeWorkspace(onChange: () => void): () => void {
  window.addEventListener("storage", onChange);
  window.addEventListener(WORKSPACE_CHANGED_EVENT, onChange);
  return () => {
    window.removeEventListener("storage", onChange);
    window.removeEventListener(WORKSPACE_CHANGED_EVENT, onChange);
  };
}
const readWorkspaceJson = () => window.localStorage.getItem(workspaceStorageKeys.workspace) ?? "";
const noWorkspaceJson = () => "";

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
  const [activeTab, setActiveTab] = useState<"table" | "50twi">(() => (config.formType === "50twi" ? "50twi" : "table"));

  const [selectedRecord, setSelectedRecord] = useState<ThaiTaxRecord | null>(null);
  const [selectedWhtRow, setSelectedWhtRow] = useState<WhtReportRow | null>(null);
  // 50 ทวิ: ข้อมูลผู้จ่าย/ผู้รับเงินที่แก้แต่ยังไม่บันทึกลงใบสำคัญ — ถามก่อนเปลี่ยนแท็บหรืองวด (แผงถูกถอดแล้วค่าที่แก้หาย)
  const whtDirtyRef = useRef(false);
  const [whtDirty, setWhtDirty] = useState(false);
  const reportWhtDirty = useCallback((dirty: boolean) => { whtDirtyRef.current = dirty; setWhtDirty(dirty); }, []);
  // ปิดแท็บงาน/เปลี่ยนบริษัท/ออกจากระบบ/รีเฟรชหน้า ขณะมีค่าที่ยังไม่บันทึก ต้องถามก่อนเหมือนจอบัญชีแยกประเภท
  // (เดิมรู้แค่การเปลี่ยนแท็บ/งวดภายในจอนี้ — ปิดแท็บของเมนูหลักแล้วค่าที่แก้หายเงียบ ๆ; review 2026-09-24)
  useDirtyGuard(route, whtDirty);
  const { confirm: confirmLeave, confirmationDialog: leaveDialog } = useConfirmDialog();
  const leaveWhtEdits = () => leaveWhtCertificate(whtDirtyRef.current, confirmLeave, tr);

  const [records, setRecords] = useState<ThaiTaxRecord[]>([]);
  const [salesSummary, setSalesSummary] = useState<{ total: number; summary: VatRegisterSummary } | null>(null);
  const [purchaseSummary, setPurchaseSummary] = useState<{ total: number; summary: VatRegisterSummary } | null>(null);
  const [whtRows, setWhtRows] = useState<WhtReportRow[]>([]);
  const [whtSummary, setWhtSummary] = useState<{ total: number; summary: WhtReportSummary; note: string } | null>(null);
  const [company, setCompany] = useState<CompanyHeader | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [errorKey, setErrorKey] = useState<string | null>(null);

  // รองรับทั้ง VAT และ WHT ทุกประเภท
  const isVatType = config.formType === "vat_sale" || config.formType === "vat_buy";
  const isWhtType = config.formType === "50twi" || config.formType === "wht_received" || config.formType === "wht_summary";
  // ภาษีถูกหัก: ลูกค้า (ผู้จ่ายเงิน) เป็นผู้ออก 50 ทวิ ให้เรา — จอนี้ดูรายละเอียดที่บันทึกได้อย่างเดียว ห้ามออกใบแทนลูกค้า
  const isReceivedWht = config.formType === "wht_received";
  // ภาษาที่เลือกในแอปส่งไปกับทุกคำขอรายงาน (ข้อความประเภทเงินได้/หมายเหตุจาก backend) — เก็บใน ref ไม่ใส่ใน deps ของ loadData
  // เพราะโหลดใหม่ตอนสลับภาษาจะได้แถวชุดใหม่ แล้วแผง 50 ทวิ ล้างข้อมูลผู้จ่าย/ผู้รับเงินที่แก้ค้างอยู่ทิ้ง; งวดถัดไปได้ภาษาใหม่เอง
  const languageRef = useRef(language);
  useEffect(() => { languageRef.current = language; }, [language]);

  const loadData = useCallback(async () => {
    setLoading(true);
    setErrorKey(null);

    try {
      if (isVatType) {
        const period = { holdingcode, businesscode, year: selectedYear, month: selectedMonth, limit: REGISTER_PAGE_LIMIT };
        const registerType: "sale" | "purchase" =
          config.formType === "vat_buy" ? "purchase" : "sale";
        const registerResult = await fetchVatRegister({ ...period, type: registerType, language: languageRef.current });
        const periodSummary = { total: registerResult.total, summary: registerResult.summary };

        setRecords(registerResult.records);
        if (registerType === "sale") {
          setSalesSummary(periodSummary);
        } else {
          setPurchaseSummary(periodSummary);
        }
        setErrorKey(registerResult.error ?? null);
      } else if (isWhtType) {
        // ข้อมูลภาษีหัก ณ ที่จ่ายจากบัญชีแยกประเภทที่ผ่านรายการจริง (backend /api/report/tax/wht)
        const whtDirection: "paid" | "received" = config.formType === "wht_received" ? "received" : "paid";
        const whtResult = await fetchWhtReport({
          holdingcode,
          businesscode,
          year: selectedYear,
          month: selectedMonth,
          direction: whtDirection,
          limit: REGISTER_PAGE_LIMIT,
          language: languageRef.current,
        });

        setWhtRows(whtResult.rows);
        setSelectedWhtRow(whtResult.rows[0] ?? null);
        setWhtSummary(whtResult.error ? null : { total: whtResult.total, summary: whtResult.summary, note: whtResult.note });
        setCompany(whtResult.error ? null : whtResult.company);

        setRecords(whtRegisterRecords(whtResult.rows));
        setErrorKey(whtResult.error ?? null);
      }
    } catch {
      setRecords([]);
      setWhtRows([]);
      setSelectedWhtRow(null);
      setSalesSummary(null);
      setPurchaseSummary(null);
      setWhtSummary(null);
      setErrorKey("connection_error");
    } finally {
      setLoading(false);
    }
  }, [config.formType, isVatType, isWhtType, holdingcode, businesscode, selectedYear, selectedMonth]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const filteredRecords = useMemo(() => {
    if (!searchTerm.trim()) return records;
    const term = searchTerm.toLowerCase();
    return records.filter(
      (r) =>
        r.taxinvoiceno.toLowerCase().includes(term) ||
        r.counterpartyname.toLowerCase().includes(term) ||
        r.taxid.includes(term),
    );
  }, [records, searchTerm]);

  // ยอดรวมทั้งงวดจาก backend (decimal) — ไม่บวกเลขบน browser
  const totals = useMemo(() => {
    if (isWhtType) {
      const summary = whtSummary?.summary;
      return {
        count: whtSummary?.total ?? 0,
        beforeVat: summary?.basetotal ?? "0.00",
        tax: summary?.whttotal ?? "0.00",
        total: summary?.nettotal ?? "0.00",
        duplicates: 0,
      };
    }
    const register = config.formType === "vat_buy" ? purchaseSummary : salesSummary;
    return {
      count: register?.total ?? 0,
      beforeVat: register?.summary.amountbeforevat ?? "0.00",
      tax: register?.summary.vatamount ?? "0.00",
      total: register?.summary.totalamount ?? "0.00",
      duplicates: register?.summary.duplicatecount ?? 0,
    };
  }, [isWhtType, whtSummary, config.formType, purchaseSummary, salesSummary]);

  const notSpecified = tr("tax_not_specified", "ยังไม่ระบุ");
  const workspaceJson = useSyncExternalStore(subscribeWorkspace, readWorkspaceJson, noWorkspaceJson);
  const companyName = taxCompanyLabel(company, workspaceJson, holdingcode, businesscode) || notSpecified;
  // ใบกำกับฉบับเดียวกันถูกบันทึกในใบสำคัญอื่นด้วย (backend นับทั้งงวด) — เตือนก่อนยื่น เพราะใบกำกับหนึ่งฉบับใช้ได้ครั้งเดียว (ม.82/5)
  const showDuplicateNotice = isVatType && !loading && !errorKey && totals.duplicates > 0;

  const canExport = !loading && !errorKey && filteredRecords.length > 0;

  // backend ส่งแถวไม่เกิน REGISTER_PAGE_LIMIT แต่ total/summary เป็นของทั้งงวด — ต้องบอกเมื่อแถวที่เห็นไม่ครบ
  const isTruncated = !loading && !errorKey && records.length < totals.count;

  // สถานะตรวจจากแถวที่โหลดจริง: ตรวจได้แค่เลขผู้เสียภาษี 13 หลัก จึงห้ามขึ้นว่า "พร้อมยื่นแบบ" ลอย ๆ
  // (สื่อสถานะด้วยข้อความ + ไอคอน ไม่ใช่สีอย่างเดียว)
  const taxIdIssues = useMemo(() => countTaxIdIssues(records), [records]);
  const checkedScope = tr("tax_check_scope", "ตรวจจาก {0} รายการที่โหลด").replace("{0}", String(records.length));
  const taxIdCheck = loading
    ? { tone: "text-muted-foreground", Icon: Loader2, text: tr("tax_check_loading", "กำลังตรวจสอบ..."), detail: "" }
    : errorKey
      ? { tone: "text-destructive", Icon: AlertCircle, text: tr("tax_check_failed", "ตรวจสอบไม่ได้ เพราะโหลดข้อมูลไม่สำเร็จ"), detail: "" }
      : records.length === 0
        ? { tone: "text-muted-foreground", Icon: FileText, text: tr("tax_check_no_rows", "ไม่มีรายการในงวดนี้"), detail: "" }
        : taxIdIssues > 0
          ? {
              tone: "text-destructive",
              Icon: AlertTriangle,
              text: tr("tax_check_taxid_issues", "{0} รายการ เลขประจำตัวผู้เสียภาษีไม่ครบ 13 หลัก — กรุณาตรวจ").replace("{0}", String(taxIdIssues)),
              detail: checkedScope,
            }
          : { tone: "text-primary", Icon: CheckCircle2, text: tr("tax_check_taxid_ok", "เลขประจำตัวผู้เสียภาษีครบ 13 หลักทุกรายการ"), detail: checkedScope };

  const partyHeader = isReceivedWht
    ? tr("tax_register_col_withholder", "ชื่อผู้หักภาษี (ผู้จ่ายเงิน)")
    : isWhtType
      ? tr("tax_register_col_payee", "ชื่อผู้ถูกหักภาษี (ผู้รับเงิน)")
      : tr("tax_register_col_party", "ชื่อผู้ซื้อ/ผู้ขาย");
  const viewDetails = tr("tax_register_view_details", "ดูรายละเอียด");
  const closeLabel = tr("close", "ปิด");
  const baht = tr("baht", "บาท");

  // Escape ปิดหน้าต่างรายละเอียด (แบบเดียวกับกรอบตัวอย่างแบบยื่น) — ผู้ใช้คีย์บอร์ดไม่ต้องหาปุ่มปิด
  useEffect(() => {
    if (!selectedRecord) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setSelectedRecord(null);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [selectedRecord]);

  const selectRecord = (item: ThaiTaxRecord) => {
    setSelectedRecord(item);
    if (isWhtType) {
      const matchedWht = whtRows.find((w) => w.rowid === item.id);
      if (matchedWht) setSelectedWhtRow(matchedWht);
    }
  };

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
          {/* Action buttons for WHT (PND.3 / PND.53 / 50 Twi) */}
          {isWhtType && (
            <>
              <Button
                variant={activeTab === "50twi" ? "default" : "outline"}
                onClick={() => setActiveTab("50twi")}
                className="gap-2 shadow-sm font-medium"
              >
                <FileText className="h-4 w-4" />
                {isReceivedWht
                  ? tr("wht_received_ui_tab", "รายละเอียด 50 ทวิ ที่ได้รับ")
                  : tr("tax_print_50twi", "หนังสือรับรอง 50 ทวิ")}
              </Button>
              <Button
                variant={activeTab === "table" ? "default" : "outline"}
                onClick={() => void leaveWhtEdits().then((ok) => { if (ok) setActiveTab("table"); })}
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
              disabled={!canExport}
              onClick={() => window.print()}
              className="gap-2 shadow-sm"
            >
              <Printer className="h-4 w-4" />
              {tr("print_report", "พิมพ์รายงาน")}
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
              onChange={(e) => { const month = Number(e.target.value); void leaveWhtEdits().then((ok) => { if (ok) setSelectedMonth(month); }); }}
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
              onChange={(val) => { const year = Number(val); void leaveWhtEdits().then((ok) => { if (ok) setSelectedYear(year); }); }}
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
              : tr("tax_ui_vat_total", "ยอดภาษีมูลค่าเพิ่มรวม")}
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
          <div role="status" className={`mt-1 flex items-start gap-1.5 ${taxIdCheck.tone}`} data-field="taxid-check">
            <taxIdCheck.Icon className={`mt-0.5 h-5 w-5 shrink-0 ${loading ? "animate-spin" : ""}`} aria-hidden />
            <span className="text-[0.95rem] font-semibold leading-normal">{taxIdCheck.text}</span>
          </div>
          {taxIdCheck.detail && <span className="text-xs text-muted-foreground">{taxIdCheck.detail}</span>}
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

      {/* 4. แท็บหนังสือรับรองการหักภาษี ณ ที่จ่าย (ใบ 50 ทวิ) — backend สร้าง PDF บนแบบฟอร์มกรมสรรพากร */}
      {isWhtType && activeTab === "50twi" && selectedWhtRow && (isReceivedWht ? (
        <WhtReceivedCertificateDetails row={selectedWhtRow} company={company} />
      ) : (
        <WhtCertificatePanel
          key={selectedWhtRow.rowid}
          row={selectedWhtRow}
          company={company}
          holdingcode={holdingcode}
          businesscode={businesscode}
          language={language}
          onDirtyChange={reportWhtDirty}
        />
      ))}
      {leaveDialog}

      {/* ตารางรายงานภาษี (Tax Register) */}
      {activeTab === "table" && (
        <Card className="overflow-hidden shadow-sm">
          <div className="border-b bg-muted/40 p-4 flex flex-wrap items-center justify-between gap-2">
            <div>
              <h2 className="text-base font-bold text-foreground">
                {isWhtType ? "ทะเบียนรายการภาษีเงินได้หัก ณ ที่จ่าย" : "ทะเบียนรายการใบกำกับภาษี"}
              </h2>
              <p className="text-xs text-muted-foreground">
                ประจำเดือน {monthNamesTh[selectedMonth - 1]} พ.ศ. {selectedYear + 543} (จำนวน {filteredRecords.length} รายการ)
              </p>
              {isTruncated && (
                <p role="status" className="mt-1 flex items-start gap-1.5 text-[0.9rem] font-semibold leading-normal text-primary" data-field="register-truncated">
                  <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden />
                  <span>
                    {tr("tax_register_truncated", "แสดง {0} จาก {1} รายการ — ยอดในการ์ดและท้ายตารางเป็นยอดรวมทั้งงวด")
                      .replace("{0}", String(records.length))
                      .replace("{1}", String(totals.count))}
                  </span>
                </p>
              )}
              {showDuplicateNotice && (
                <p role="status" className="mt-2 flex items-start gap-1.5 rounded-lg border border-destructive/40 bg-destructive/10 px-3 py-2 text-[0.95rem] leading-[1.5] text-foreground" data-field="register-duplicates">
                  <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-destructive" aria-hidden />
                  <span>
                    {tr("vat_ui_duplicate_notice", "{0} รายการในงวดนี้เป็นใบกำกับภาษีฉบับเดียวกับที่บันทึกในใบสำคัญอื่นด้วย (ผู้ออกใบกำกับและเลขที่เดียวกัน) — ใบกำกับหนึ่งฉบับใช้ได้ครั้งเดียว ตรวจแถวที่มีป้าย “ใบกำกับซ้ำ” แล้วลบรายละเอียดภาษีที่บันทึกเกิน หรือแก้เลขที่/เลขผู้เสียภาษีให้ตรงใบจริง; เลขที่เดียวกันแต่ต่างผู้ออกใบกำกับไม่นับว่าซ้ำ")
                      .replace("{0}", String(totals.duplicates))}
                  </span>
                </p>
              )}
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
                  <th className="px-3 py-3 w-12 text-center">{tr("tax_register_col_seq", "ลำดับ")}</th>
                  <th className="px-3 py-3 w-28">{tr("tax_register_col_date", "วันที่")}</th>
                  <th className="px-3 py-3 w-36">{tr("tax_register_col_docno", "เลขที่เอกสาร/ใบกำกับ")}</th>
                  <th className="px-4 py-3">{partyHeader}</th>
                  <th className="px-3 py-3 w-36">{tr("tax_register_col_taxid", "เลขประจำตัวผู้เสียภาษี 13 หลัก")}</th>
                  <th className="px-3 py-3 w-20 text-center">{tr("tax_register_col_branch", "สาขา")}</th>
                  <th className="px-4 py-3 text-right">{tr("tax_register_col_base", "มูลค่าก่อนภาษี")}</th>
                  <th className="px-4 py-3 text-right">{tr("tax_register_col_tax", "ภาษี")}</th>
                  <th className="px-4 py-3 text-right">{tr("tax_register_col_net", "ยอดสุทธิ")}</th>
                  <th className="px-3 py-3 w-20 text-center print:hidden">{tr("tax_register_col_actions", "จัดการ")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredRecords.map((item, idx) => (
                  <tr
                    key={item.id}
                    className="hover:bg-muted/30 transition-colors cursor-pointer"
                    onClick={() => selectRecord(item)}
                  >
                    <td className="px-3 py-2.5 text-center text-muted-foreground font-mono">{idx + 1}</td>
                    <td className="px-3 py-2.5 text-xs whitespace-nowrap">{formatAppDate(item.docdate, language)}</td>
                    <td className="px-3 py-2.5 font-medium text-primary">
                      {item.taxinvoiceno}
                      <DuplicateInvoiceBadge docnos={item.duplicatedocnos} tr={tr} />
                    </td>
                    <td className="px-4 py-2.5">
                      <div className="font-medium text-foreground">{item.counterpartyname}</div>
                      {item.incometype && (
                        <div className="text-xs text-muted-foreground">{item.incometype}</div>
                      )}
                      {item.remark && (
                        <div role="note" className="mt-0.5 flex items-start gap-1 text-sm leading-relaxed text-foreground">
                          <AlertTriangle className="mt-1 size-3.5 shrink-0 text-primary" aria-hidden />
                          <span>{item.remark}</span>
                        </div>
                      )}
                    </td>
                    <td className="px-3 py-2.5 font-mono text-xs text-muted-foreground">{item.taxid}</td>
                    <td className="px-3 py-2.5 text-center">
                      {item.branchno && (
                        <span className="rounded bg-muted px-1.5 py-0.5 text-xs font-mono">
                          {item.branchno}
                        </span>
                      )}
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
                          selectRecord(item);
                        }}
                        aria-label={viewDetails}
                        title={viewDetails}
                        className="h-8 px-2 text-xs"
                      >
                        <FileText className="h-4 w-4" aria-hidden />
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
        // คลิกพื้นหลังหรือกด Escape ปิดได้ (useEffect ด้านบน) — ข้อความทุกคำมาจาก languages.tsv
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 print:hidden"
          onClick={(e) => {
            if (e.target === e.currentTarget) setSelectedRecord(null);
          }}
        >
          <Card className="w-full max-w-xl shadow-2xl" role="dialog" aria-modal="true" aria-labelledby="tax-record-detail-title">
            <div className="flex items-center justify-between border-b p-4">
              <div className="flex items-center gap-2">
                <FileText className="h-5 w-5 text-primary" />
                <h3 id="tax-record-detail-title" className="font-bold text-foreground">
                  {tr("tax_detail_title", "รายละเอียดเอกสาร: {docno}").replace("{docno}", selectedRecord.taxinvoiceno)}
                </h3>
              </div>
              <Button size="sm" variant="ghost" onClick={() => setSelectedRecord(null)} aria-label={closeLabel} title={closeLabel}>
                ✕
              </Button>
            </div>
            <CardContent className="space-y-3 p-5 text-sm">
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <span className="text-xs text-muted-foreground">{tr("tax_register_col_date", "วันที่")}</span>
                  <p className="font-medium">{formatAppDate(selectedRecord.docdate, language)}</p>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground">{tr("tax_register_col_docno", "เลขที่เอกสาร/ใบกำกับ")}</span>
                  <p className="font-medium text-primary">{selectedRecord.taxinvoiceno}</p>
                  <DuplicateInvoiceBadge docnos={selectedRecord.duplicatedocnos} tr={tr} />
                </div>
                <div className="col-span-2">
                  <span className="text-xs text-muted-foreground">{partyHeader}</span>
                  <p className="font-medium">{selectedRecord.counterpartyname}</p>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground">{tr("tax_register_col_taxid", "เลขประจำตัวผู้เสียภาษี 13 หลัก")}</span>
                  <p className="font-mono font-medium">{selectedRecord.taxid}</p>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground">{tr("tax_register_col_branch", "สาขา")}</span>
                  <p className="font-mono">{selectedRecord.branchno} {selectedRecord.isheadoffice ? `(${tr("head_office", "สำนักงานใหญ่")})` : ""}</p>
                </div>
                {selectedRecord.incometype && (
                  <div className="col-span-2">
                    <span className="text-xs text-muted-foreground">{tr("tax_detail_income_type", "ประเภทเงินได้")}</span>
                    <p className="font-medium text-amber-600 dark:text-amber-400">{selectedRecord.incometype}</p>
                  </div>
                )}
              </div>

              <div className="rounded-xl border border-border bg-muted/40 p-3 space-y-1.5 font-mono">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{tr("tax_register_col_base", "มูลค่าก่อนภาษี")}</span>
                  <span>{money(selectedRecord.amountbeforevat)} {baht}</span>
                </div>
                <div className="flex justify-between text-primary font-semibold">
                  {/* ภาษีมูลค่าเพิ่มไม่ใส่อัตราตายตัว: รายการส่งออก 0% / ยกเว้น ก็เปิดดูจากตารางเดียวกัน (UAT V29 2026-09-24) */}
                  <span>
                    {selectedRecord.whtamount !== undefined
                      ? `${tr("tax_detail_wht_amount", "ภาษีหัก ณ ที่จ่าย")}${selectedRecord.taxrate ? ` (${selectedRecord.taxrate}%)` : ""}`
                      : tr("tax_detail_vat_amount", "ภาษีมูลค่าเพิ่ม")}
                  </span>
                  <span>{money(selectedRecord.whtamount ?? selectedRecord.vatamount)} {baht}</span>
                </div>
                <div className="flex justify-between border-t pt-1 text-base font-bold">
                  <span>{tr("tax_register_col_net", "ยอดสุทธิ")}</span>
                  {/* ยอดสุทธิว่าง = backend ยังไม่รู้ฐานภาษี (แถวประมาณที่แบ่งฐานไม่ได้) — ไม่แสดงเลขติดลบ */}
                  <span>{selectedRecord.totalamount ? `${money(selectedRecord.totalamount)} ${baht}` : "—"}</span>
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <Button variant="outline" onClick={() => setSelectedRecord(null)}>
                  {closeLabel}
                </Button>
                {/* 50 ทวิ ออกจากแผง PDF ของ backend เท่านั้น (window.print() พิมพ์หน้าทะเบียน ไม่ใช่หนังสือรับรอง);
                    ภาษีถูกหัก = ผู้จ่ายเงินเป็นผู้ออกใบให้เรา และภาษีมูลค่าเพิ่มไม่มีหนังสือรับรอง จึงไม่มีปุ่มนี้ */}
                {isWhtType && !isReceivedWht && selectedWhtRow?.rowid === selectedRecord.id && (
                  <Button
                    variant="default"
                    onClick={() => {
                      setActiveTab("50twi");
                      setSelectedRecord(null);
                    }}
                    className="gap-1.5"
                  >
                    <FileText className="h-4 w-4" aria-hidden />
                    {tr("tax_ui_open_50twi", "ออกหนังสือรับรอง 50 ทวิ ของรายการนี้")}
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      )}

    </div>
  );
}

// ป้ายใบกำกับซ้ำ: ไอคอน + ข้อความ + เลขที่ใบสำคัญอื่น (ห้ามสื่อด้วยสีอย่างเดียว — กฎ 40+ ข้อ 5)
function DuplicateInvoiceBadge({ docnos, tr }: { docnos: string[]; tr: (key: string, fallback: string) => string }) {
  if (docnos.length === 0) return null;
  return (
    <div className="mt-1 grid gap-0.5" data-field="duplicate-invoice">
      <span className="inline-flex w-fit items-center gap-1 rounded-md border border-destructive/40 bg-destructive/10 px-1.5 py-0.5 text-[0.85rem] font-semibold leading-normal text-destructive">
        <AlertTriangle className="h-3.5 w-3.5 shrink-0" aria-hidden />
        {tr("vat_ui_duplicate_badge", "ใบกำกับซ้ำ")}
      </span>
      <span className="text-[0.85rem] font-normal leading-normal text-foreground [overflow-wrap:anywhere]">
        {tr("vat_ui_duplicate_docnos", "บันทึกซ้ำในใบสำคัญ {0}").replace("{0}", docnos.join(", "))}
      </span>
    </div>
  );
}
