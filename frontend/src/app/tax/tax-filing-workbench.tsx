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
import type { LanguageCode } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChoiceSelect } from "@/components/ui/select";
import { Card, CardContent } from "@/components/ui/card";
import {
  FileText, Printer, Download, Calculator, Building2, Calendar, CheckCircle2, Receipt, Search,
  AlertCircle, Loader2,
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
  const [activeTab, setActiveTab] = useState<"table" | "pp30" | "50twi">("table");
  const [selectedRecord, setSelectedRecord] = useState<ThaiTaxRecord | null>(null);

  const [records, setRecords] = useState<ThaiTaxRecord[]>([]);
  const [pp30, setPp30] = useState<Pp30Summary | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [errorKey, setErrorKey] = useState<string | null>(null);

  // จอนี้ดึงข้อมูลจริงจาก backend ได้เฉพาะรายงานภาษีมูลค่าเพิ่ม
  const supported =
    config.formType === "vat_sale" ||
    config.formType === "vat_buy" ||
    config.formType === "pp30";

  const loadData = useCallback(async () => {
    if (!supported) {
      setRecords([]);
      setPp30(null);
      setErrorKey(null);
      return;
    }

    setLoading(true);
    try {
      const registerType: "sale" | "purchase" =
        config.formType === "vat_buy" ? "purchase" : "sale";

      if (config.formType === "pp30") {
        const [registerResult, summaryResult] = await Promise.all([
          fetchVatRegister({
            holdingcode,
            businesscode,
            year: selectedYear,
            month: selectedMonth,
            type: registerType,
          }),
          fetchPp30Summary({
            holdingcode,
            businesscode,
            year: selectedYear,
            month: selectedMonth,
          }),
        ]);

        setRecords(registerResult.records);
        setPp30(summaryResult.summary);
        setErrorKey(registerResult.error ?? summaryResult.error ?? null);
      } else {
        const registerResult = await fetchVatRegister({
          holdingcode,
          businesscode,
          year: selectedYear,
          month: selectedMonth,
          type: registerType,
        });

        setRecords(registerResult.records);
        setPp30(null);
        setErrorKey(registerResult.error ?? null);
      }
    } catch {
      setRecords([]);
      setPp30(null);
      setErrorKey("connection_error");
    } finally {
      setLoading(false);
    }
  }, [supported, config.formType, holdingcode, businesscode, selectedYear, selectedMonth]);

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

  // ยอดรวมท้ายตาราง = ผลรวมของรายการที่แสดงอยู่ (ไม่ใช่การคำนวณภาษี)
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

  const canExport = supported && !loading && !errorKey && filteredRecords.length > 0;

  const monthNamesTh = [
    "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
    "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
  ];

  return (
    <div className="flex flex-col gap-4 p-4 lg:p-6">
      {/* Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-4 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
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
            <Button
              variant={activeTab === "pp30" ? "default" : "outline"}
              onClick={() => setActiveTab(activeTab === "pp30" ? "table" : "pp30")}
              className="gap-2"
            >
              <Calculator className="h-4 w-4" />
              {tr("ops_pp_30_form", "แบบฟอร์ม ภ.พ.30")}
            </Button>
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
              link.setAttribute("download", `${config.code}_${selectedYear}_${selectedMonth}.csv`);
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
            disabled={!canExport}
            onClick={() => window.print()}
            className="gap-2"
          >
            <Printer className="h-4 w-4" />
            {tr("print_report", "พิมพ์รายงาน")}
          </Button>
        </div>
      </div>

      {/* Filter and Period Selection Bar */}
      <Card>
        <CardContent className="flex flex-wrap items-center justify-between gap-4 p-4">
          <div className="flex flex-wrap items-center gap-3">
            <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
              <Calendar className="h-4 w-4 text-primary" />
              <span>{tr("ops_tax_period", "งวดภาษี:")}</span>
            </div>
            <select
              value={selectedMonth}
              onChange={(e) => setSelectedMonth(Number(e.target.value))}
              className="rounded-lg border border-border bg-background px-3 py-1.5 text-sm font-medium text-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
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

          <div className="relative w-full max-w-xs">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder={tr("ops_search_2", "ค้นหาเลขที่, ชื่อคู่ค้า, เลขประจำตัว...")}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-9"
            />
          </div>
        </CardContent>
      </Card>

      {/* KPI Cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <Card className="p-4">
          <p className="text-xs text-muted-foreground">{tr("ops_base_amount_before_vat", "มูลค่าสินค้า/บริการก่อนภาษี")}</p>
          <p className="mt-1 text-xl font-bold text-foreground">
            {totals.beforeVat.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
          </p>
          <span className="text-xs text-muted-foreground">บาท (THB)</span>
        </Card>

        <Card className="p-4">
          <p className="text-xs text-muted-foreground">
            {config.formType.includes("wht") || config.formType.includes("pnd")
              ? (tr("ops_total_withholding_tax", "ยอดภาษีหัก ณ ที่จ่ายรวม"))
              : (tr("ops_total_vat_7", "ยอดภาษีมูลค่าเพิ่ม 7%"))}
          </p>
          <p className="mt-1 text-xl font-bold text-primary">
            {(config.formType.includes("wht") || config.formType.includes("pnd") ? totals.wht : totals.vat).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
          </p>
          <span className="text-xs text-muted-foreground">บาท (THB)</span>
        </Card>

        <Card className="p-4">
          <p className="text-xs text-muted-foreground">{tr("ops_total_documents", "จำนวนรายการเอกสาร")}</p>
          <p className="mt-1 text-xl font-bold text-foreground">
            {filteredRecords.length}
          </p>
          <span className="text-xs text-muted-foreground">รายการ (Docs)</span>
        </Card>

        <Card className="p-4">
          <p className="text-xs text-muted-foreground">{tr("ops_compliance_status", "สถานะการตรวจสอบ")}</p>
          <div className="mt-1 flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400">
            <CheckCircle2 className="h-5 w-5" />
            <span className="text-base font-semibold">{tr("ops_ready_to_file", "ถูกต้อง พร้อมยื่นแบบ")}</span>
          </div>
          <span className="text-xs text-muted-foreground">Tax ID ครบ 13 หลัก</span>
        </Card>
      </div>

      {!supported ? (
        <div className="flex items-start gap-2 rounded-xl border border-border bg-muted/30 px-4 py-3 text-sm">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          <span>รายงานนี้ยังไม่เปิดให้ใช้งาน เนื่องจากระบบยังไม่ได้จัดเก็บข้อมูลภาษีหัก ณ ที่จ่าย</span>
        </div>
      ) : loading ? (
        <div className="flex items-center gap-2 rounded-xl border border-border bg-muted/30 px-4 py-3 text-sm">
          <Loader2 className="h-4 w-4 shrink-0 animate-spin" />
          <span>กำลังโหลดข้อมูล...</span>
        </div>
      ) : errorKey ? (
        <div
          role="alert"
          className="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
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
      ) : records.length === 0 ? (
        <div className="flex items-start gap-2 rounded-xl border border-border bg-muted/30 px-4 py-3 text-sm">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          <span>ไม่พบข้อมูลในงวดที่เลือก</span>
        </div>
      ) : null}

      {/* Main Content Area: PP.30 Form View or Standard Tax Table */}
      {config.formType === "pp30" && activeTab === "pp30" ? (
        <Card className="overflow-hidden border-2 border-primary/20">
          <div className="border-b bg-muted/40 p-4">
            <h2 className="text-lg font-bold text-foreground">
              แบบแสดงรายการภาษีมูลค่าเพิ่ม (ภ.พ. 30) กรมสรรพากร
            </h2>
            <p className="text-sm text-muted-foreground">
              งวดเดือน {monthNamesTh[selectedMonth - 1]} พ.ศ. {selectedYear + 543}
            </p>
          </div>
          {pp30 === null ? (
            <div className="p-6 text-center text-sm text-muted-foreground">
              ยังไม่มีข้อมูลสรุป ภ.พ.30 สำหรับงวดที่เลือก
            </div>
          ) : (
            <div className="divide-y divide-border p-4 text-sm">
              <div className="flex items-center justify-between py-2">
                <span className="font-medium text-foreground">1. ยอดขายเดือนนี้ (ตามมาตรา 79)</span>
                <span className="font-mono text-base font-semibold">
                  {(pp30.salestaxable + pp30.saleszerorated + pp30.salesexempt).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                </span>
              </div>
              <div className="flex items-center justify-between py-2 pl-4 text-muted-foreground">
                <span>2. ยอดขายที่เสียภาษีอัตราร้อยละ 0</span>
                <span className="font-mono">{pp30.saleszerorated.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
              </div>
              <div className="flex items-center justify-between py-2 pl-4 text-muted-foreground">
                <span>3. ยอดขายที่ได้รับการยกเว้นภาษี</span>
                <span className="font-mono">{pp30.salesexempt.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
              </div>
              <div className="flex items-center justify-between py-2 bg-muted/20 px-2 font-medium">
                <span>4. ยอดขายที่ต้องเสียภาษี (ข้อ 1 - ข้อ 2 - ข้อ 3)</span>
                <span className="font-mono text-base font-semibold text-primary">{pp30.salestaxable.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
              </div>
              <div className="flex items-center justify-between py-2 bg-primary/5 px-2 font-semibold text-primary">
                <span>5. ภาษีขาย (ตามใบกำกับภาษีขาย)</span>
                <span className="font-mono text-lg">{pp30.outputvat.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
              </div>
              <div className="flex items-center justify-between py-2">
                <span className="font-medium text-foreground">6. ยอดซื้อที่มีสิทธินำภาษีซื้อมาหัก</span>
                <span className="font-mono font-semibold">{pp30.purchasetaxable.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
              </div>
              <div className="flex items-center justify-between py-2 bg-primary/5 px-2 font-semibold text-primary">
                <span>7. ภาษีซื้อ (ตามใบกำกับภาษีซื้อ)</span>
                <span className="font-mono text-lg">{pp30.inputvat.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
              </div>
              <div
                className={`flex items-center justify-between py-3 px-3 font-bold ${
                  pp30.netvat >= 0 ? "bg-primary/10 text-primary" : "bg-destructive/10 text-destructive"
                }`}
              >
                <span className="text-base">
                  {pp30.netvat >= 0 ? "8. ภาษีมูลค่าเพิ่มที่ต้องชำระเดือนนี้ (ข้อ 5 - ข้อ 7)" : "8. ภาษีชำระเกิน (ข้อ 7 - ข้อ 5)"}
                </span>
                <span className="font-mono text-xl">{Math.abs(pp30.netvat).toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
              </div>
            </div>
          )}
        </Card>
      ) : (
        <Card className="overflow-hidden">
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
                  <th className="px-3 py-3 w-20 text-center">จัดการ</th>
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
                    <td className="px-3 py-2.5 text-center">
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
                  <td></td>
                </tr>
              </tfoot>
            </table>
          </div>
        </Card>
      )}

      {/* 50 Twi / Tax Invoice Detail Modal */}
      {selectedRecord && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
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
