"use client";

import { useState } from "react";
import type { LanguageCode } from "@/lib/i18n";
import { apiFetch } from "@/lib/client-auth-session";
import {
  FileText,
  Send,
  CheckCircle2,
  AlertCircle,
  ShieldCheck,
  Download,
  Search,
  RefreshCw,
  Building2,
  Calendar,
  KeyRound,
  ExternalLink,
  SlidersHorizontal,
} from "lucide-react";

interface ETaxInvoiceScreenProps {
  language: LanguageCode;
  holdingcode: string;
  branchcode?: string;
}

interface ETaxDocument {
  id: string;
  docNo: string;
  date: string;
  customerName: string;
  taxId: string;
  typeCode: "T01" | "T02" | "80" | "81" | "T04";
  typeName: string;
  amount: number;
  vatAmount: number;
  totalAmount: number;
  signStatus: "signed" | "pending" | "failed";
  rdStatus: "submitted" | "ready" | "rejected";
  rdReceiptNo?: string;
}

const MOCK_DOCUMENTS: ETaxDocument[] = [
  {
    id: "1",
    docNo: "INV6910-0012",
    date: "2026-10-10",
    customerName: "บริษัท สยามพัฒนาการก่อสร้าง จำกัด (มหาชน)",
    taxId: "0107537000123",
    typeCode: "T01",
    typeName: "ใบกำกับภาษี (Invoice)",
    amount: 150000.0,
    vatAmount: 10500.0,
    totalAmount: 160500.0,
    signStatus: "signed",
    rdStatus: "submitted",
    rdReceiptNo: "RD-20261010-0091823",
  },
  {
    id: "2",
    docNo: "INV6910-0013",
    date: "2026-10-10",
    customerName: "ห้างหุ้นส่วนจำกัด ธนบุรีคอนกรีต",
    taxId: "0103548002941",
    typeCode: "T01",
    typeName: "ใบกำกับภาษี (Invoice)",
    amount: 45000.0,
    vatAmount: 3150.0,
    totalAmount: 48150.0,
    signStatus: "signed",
    rdStatus: "ready",
  },
  {
    id: "3",
    docNo: "SR6910-0004",
    date: "2026-10-09",
    customerName: "บริษัท ซีเมนต์ไทยภัณฑ์ จำกัด",
    taxId: "0105552003819",
    typeCode: "81",
    typeName: "ใบลดหนี้ (Credit Note - CDNG05)",
    amount: -12000.0,
    vatAmount: -840.0,
    totalAmount: -12840.0,
    signStatus: "signed",
    rdStatus: "submitted",
    rdReceiptNo: "RD-20261009-0087114",
  },
  {
    id: "4",
    docNo: "DN6910-0002",
    date: "2026-10-08",
    customerName: "บริษัท บางกอกฮาร์ดแวร์ เทรดดิ้ง จำกัด",
    taxId: "0105546001294",
    typeCode: "80",
    typeName: "ใบเพิ่มหนี้ (Debit Note - DBNG01)",
    amount: 8500.0,
    vatAmount: 595.0,
    totalAmount: 9095.0,
    signStatus: "pending",
    rdStatus: "ready",
  },
];

export function ETaxInvoiceScreen({ language, holdingcode, branchcode = "00000" }: ETaxInvoiceScreenProps) {
  const isThai = language === "th";
  const [activeTab, setActiveTab] = useState<"queue" | "log" | "cert" | "config">("queue");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedDocs, setSelectedDocs] = useState<Set<string>>(new Set(["2", "4"]));
  const [serviceMode, setServiceMode] = useState<"api" | "email">("api");
  const [autoSign, setAutoSign] = useState(true);

  const filteredDocs = MOCK_DOCUMENTS.filter(
    (d) =>
      d.docNo.toLowerCase().includes(searchQuery.toLowerCase()) ||
      d.customerName.toLowerCase().includes(searchQuery.toLowerCase()) ||
      d.taxId.includes(searchQuery),
  );

  const toggleSelect = (id: string) => {
    const next = new Set(selectedDocs);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setSelectedDocs(next);
  };

  const toggleSelectAll = () => {
    if (selectedDocs.size === filteredDocs.length) {
      setSelectedDocs(new Set());
    } else {
      setSelectedDocs(new Set(filteredDocs.map((d) => d.id)));
    }
  };

  const handleDownloadXml = async (doc: ETaxDocument) => {
    try {
      const payload = {
        document: {
          doc_no: doc.docNo,
          type_code: doc.typeCode,
          type_name: doc.typeName,
          issue_date_time: `${doc.date}T10:00:00.000Z`,
          purpose_code: doc.typeCode === "81" ? "CDNG05" : doc.typeCode === "80" ? "DBNG01" : undefined,
          purpose_reason: doc.typeCode === "81" ? "รับคืนสินค้าเนื่องจากสินค้าผิดคุณลักษณะ" : doc.typeCode === "80" ? "คำนวณราคาสินค้าต่ำกว่าที่เป็นจริง" : undefined,
          original_doc_no: doc.typeCode === "81" || doc.typeCode === "80" ? "INV6910-0012" : undefined,
          original_doc_date: doc.typeCode === "81" || doc.typeCode === "80" ? "2026-10-08" : undefined,
          currency_code: "THB",
          seller: {
            name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด",
            tax_id: "0105560012345",
            branch_id: branchcode,
            address: {
              line_one: "123/45 ถนนสุขุมวิท แขวงคลองเตย เขตคลองเตย กรุงเทพมหานคร",
              postcode: "10110",
              country_code: "TH",
            },
          },
          buyer: {
            name: doc.customerName,
            tax_id: doc.taxId,
            branch_id: "00000",
            address: {
              line_one: "สำนักงานใหญ่",
              postcode: "10330",
              country_code: "TH",
            },
          },
          items: [
            {
              line_id: 1,
              item_code: "ITEM-001",
              item_name: doc.typeName,
              quantity: "1.00",
              unit_code: "C62",
              unit_price: Math.abs(doc.amount).toFixed(2),
              amount: Math.abs(doc.amount).toFixed(2),
              discount: "0.00",
              net_amount: Math.abs(doc.amount).toFixed(2),
              tax_rate: "7.00",
              tax_amount: Math.abs(doc.vatAmount).toFixed(2),
            },
          ],
          line_total_amount: Math.abs(doc.amount).toFixed(2),
          discount_total: "0.00",
          tax_basis_total: Math.abs(doc.amount).toFixed(2),
          tax_rate: "7.00",
          tax_total: Math.abs(doc.vatAmount).toFixed(2),
          grand_total: Math.abs(doc.totalAmount).toFixed(2),
        },
      };

      const res = await apiFetch("/api/goapi/api/report/tax/etax/generate-xml", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      const data = await res.json();
      if (!res.ok || !data.success || !data.xml) {
        throw new Error(data.message || "Failed to generate e-Tax XML");
      }

      const blob = new Blob([data.xml], { type: "application/xml;charset=utf-8" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = data.filename || `${doc.docNo}_ETAX.xml`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      alert(isThai ? `เกิดข้อผิดพลาดในการสร้าง XML: ${msg}` : `Failed to generate XML: ${msg}`);
    }
  };

  return (
    <div className="flex h-full flex-col bg-background text-foreground">
      {/* Header Banner */}
      <div className="border-b bg-card px-6 py-4 shadow-xs">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <FileText className="h-6 w-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-xl font-bold tracking-tight">
                  {isThai ? "ระบบใบกำกับภาษีอิเล็กทรอนิกส์ (e-Tax Invoice & e-Receipt)" : "e-Tax Invoice & e-Receipt Portal"}
                </h1>
                <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                  <ShieldCheck className="h-3.5 w-3.5" />
                  มธพ. 3-2560
                </span>
              </div>
              <p className="text-xs text-muted-foreground">
                {isThai
                  ? "จัดทำ ส่งมอบ ลงลายมือชื่อดิจิทัล และนำส่งข้อมูลภาษีมูลค่าเพิ่มสู่กรมสรรพากร"
                  : "Generate, digitally sign, and transmit e-Tax documents to Revenue Department Gateway"}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              className="inline-flex h-10 items-center gap-2 rounded-lg border bg-background px-4 text-sm font-medium hover:bg-muted"
              onClick={() => alert(isThai ? "กำลังซิงค์สถานะกับกรมสรรพากร..." : "Syncing status with RD Gateway...")}
            >
              <RefreshCw className="h-4 w-4" />
              {isThai ? "อัปเดตสถานะ" : "Refresh"}
            </button>
            <button
              type="button"
              className="inline-flex h-10 items-center gap-2 rounded-lg bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-xs hover:bg-primary/90"
              onClick={() =>
                alert(
                  isThai
                    ? `นำส่งเอกสารที่เลือก ${selectedDocs.size} ฉบับสู่กรมสรรพากรเรียบร้อย`
                    : `Transmitted ${selectedDocs.size} documents to RD Gateway successfully`,
                )
              }
            >
              <Send className="h-4 w-4" />
              {isThai ? `ลงนามและนำส่งสรรพากร (${selectedDocs.size})` : `Sign & Submit (${selectedDocs.size})`}
            </button>
          </div>
        </div>

        {/* Tab navigation */}
        <div className="mt-4 flex gap-2 border-t pt-3">
          <button
            type="button"
            onClick={() => setActiveTab("queue")}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              activeTab === "queue" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-muted"
            }`}
          >
            {isThai ? "เอกสารรอส่งมอบ (Queue)" : "Pending Documents"}
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("log")}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              activeTab === "log" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-muted"
            }`}
          >
            {isThai ? "ประวัติการนำส่ง (Transmission Log)" : "Submission History"}
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("cert")}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              activeTab === "cert" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-muted"
            }`}
          >
            {isThai ? "ใบรับรองดิจิทัล (Digital Certificate)" : "Digital Certificate"}
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("config")}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              activeTab === "config" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-muted"
            }`}
          >
            {isThai ? "ตั้งค่าการเชื่อมต่อ (Settings)" : "Gateway Settings"}
          </button>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 overflow-auto p-6">
        {activeTab === "queue" && (
          <div className="space-y-4">
            {/* Filter toolbar */}
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="relative min-w-[280px] flex-1 max-w-md">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  type="text"
                  placeholder={isThai ? "ค้นหาเลขที่เอกสาร, ชื่อลูกค้า, เลขผู้เสียภาษี..." : "Search document, client, tax ID..."}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="h-10 w-full rounded-lg border bg-background pl-9 pr-3 text-sm focus:outline-hidden focus:ring-2 focus:ring-primary"
                />
              </div>

              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span className="inline-block h-2.5 w-2.5 rounded-full bg-emerald-500" />
                <span>{isThai ? "ลงนามแล้ว: 3" : "Signed: 3"}</span>
                <span className="inline-block h-2.5 w-2.5 rounded-full bg-amber-500 ml-2" />
                <span>{isThai ? "รอดำเนินการ: 1" : "Pending: 1"}</span>
              </div>
            </div>

            {/* Document Table */}
            <div className="overflow-hidden rounded-xl border bg-card shadow-xs">
              <table className="w-full text-left text-sm">
                <thead className="border-b bg-muted/50 text-xs font-semibold text-muted-foreground">
                  <tr>
                    <th className="p-3 text-center">
                      <input
                        type="checkbox"
                        checked={selectedDocs.size === filteredDocs.length && filteredDocs.length > 0}
                        onChange={toggleSelectAll}
                        className="h-4 w-4 rounded-sm border"
                      />
                    </th>
                    <th className="p-3">{isThai ? "เลขที่เอกสาร" : "Doc No."}</th>
                    <th className="p-3">{isThai ? "วันที่" : "Date"}</th>
                    <th className="p-3">{isThai ? "ประเภทเอกสาร" : "Type"}</th>
                    <th className="p-3">{isThai ? "ลูกค้า / ผู้รับ" : "Customer / Recipient"}</th>
                    <th className="p-3 text-right">{isThai ? "มูลค่าสุทธิ (บาท)" : "Net Amount (THB)"}</th>
                    <th className="p-3 text-right">{isThai ? "VAT (บาท)" : "VAT (THB)"}</th>
                    <th className="p-3 text-center">{isThai ? "สถานะลงนาม" : "Signature"}</th>
                    <th className="p-3 text-center">{isThai ? "สรรพากร (RD)" : "RD Gateway"}</th>
                    <th className="p-3 text-center">{isThai ? "ดาวน์โหลด" : "Download"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {filteredDocs.map((doc) => {
                    const isSelected = selectedDocs.has(doc.id);
                    return (
                      <tr key={doc.id} className={`hover:bg-muted/40 transition-colors ${isSelected ? "bg-primary/5" : ""}`}>
                        <td className="p-3 text-center">
                          <input
                            type="checkbox"
                            checked={isSelected}
                            onChange={() => toggleSelect(doc.id)}
                            className="h-4 w-4 rounded-sm border"
                          />
                        </td>
                        <td className="p-3 font-semibold text-primary">{doc.docNo}</td>
                        <td className="p-3 text-muted-foreground">{doc.date}</td>
                        <td className="p-3">
                          <span className="rounded bg-muted px-2 py-0.5 text-xs font-medium">
                            {doc.typeCode} - {doc.typeName}
                          </span>
                        </td>
                        <td className="p-3">
                          <div className="font-medium">{doc.customerName}</div>
                          <div className="text-xs text-muted-foreground">TAX ID: {doc.taxId}</div>
                        </td>
                        <td className="p-3 text-right font-mono font-medium">
                          {doc.amount.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                        </td>
                        <td className="p-3 text-right font-mono text-muted-foreground">
                          {doc.vatAmount.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                        </td>
                        <td className="p-3 text-center">
                          {doc.signStatus === "signed" ? (
                            <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs font-medium text-emerald-600 dark:text-emerald-400">
                              <CheckCircle2 className="h-3.5 w-3.5" />
                              {isThai ? "ลงนามแล้ว" : "Signed"}
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1 rounded-full bg-amber-500/10 px-2 py-0.5 text-xs font-medium text-amber-600 dark:text-amber-400">
                              <AlertCircle className="h-3.5 w-3.5" />
                              {isThai ? "รอดำเนินการ" : "Pending"}
                            </span>
                          )}
                        </td>
                        <td className="p-3 text-center">
                          {doc.rdStatus === "submitted" ? (
                            <span className="inline-flex items-center gap-1 rounded-full bg-blue-500/10 px-2 py-0.5 text-xs font-medium text-blue-600 dark:text-blue-400">
                              <CheckCircle2 className="h-3.5 w-3.5" />
                              {doc.rdReceiptNo ?? "สำเร็จ"}
                            </span>
                          ) : (
                            <span className="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
                              {isThai ? "พร้อมส่ง" : "Ready"}
                            </span>
                          )}
                        </td>
                        <td className="p-3 text-center">
                          <div className="flex items-center justify-center gap-1">
                            <button
                              type="button"
                              title="Download PDF/A-3"
                              className="rounded p-1.5 hover:bg-muted text-muted-foreground hover:text-foreground"
                              onClick={() => alert(`ดาวน์โหลด PDF/A-3 สำหรับ ${doc.docNo}`)}
                            >
                              <Download className="h-4 w-4" />
                            </button>
                            <button
                              type="button"
                              title="Download XML มธพ. 3-2560"
                              className="rounded p-1.5 hover:bg-muted text-muted-foreground hover:text-foreground"
                              onClick={() => void handleDownloadXml(doc)}
                            >
                              <FileText className="h-4 w-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {activeTab === "log" && (
          <div className="space-y-4">
            <div className="rounded-xl border bg-card p-6 shadow-xs">
              <h2 className="text-base font-semibold">{isThai ? "ประวัติการนำส่งชุดข้อมูลสู่กรมสรรพากร" : "Batch Transmission Logs"}</h2>
              <p className="text-xs text-muted-foreground mb-4">
                {isThai ? "เก็บบันทึกรหัสตอบกลับ (Receipt Code) และ Hash Digest ตรวจสอบความถูกต้อง" : "Records RD Gateway response receipt and SHA-256 Digest"}
              </p>

              <div className="space-y-3 font-mono text-xs">
                <div className="rounded-lg border bg-muted/30 p-3">
                  <div className="flex items-center justify-between font-semibold text-emerald-600 dark:text-emerald-400">
                    <span>STATUS: 200 OK — BATCH SUCCESSFUL</span>
                    <span className="text-muted-foreground">2026-10-10 14:22:08 ICT</span>
                  </div>
                  <div className="mt-1 text-muted-foreground">
                    RD Receipt No: <span className="text-foreground">RD-20261010-0091823</span> | Records: 2 | Digest: 8f92a1...c03b
                  </div>
                </div>

                <div className="rounded-lg border bg-muted/30 p-3">
                  <div className="flex items-center justify-between font-semibold text-emerald-600 dark:text-emerald-400">
                    <span>STATUS: 200 OK — BATCH SUCCESSFUL</span>
                    <span className="text-muted-foreground">2026-10-09 16:45:12 ICT</span>
                  </div>
                  <div className="mt-1 text-muted-foreground">
                    RD Receipt No: <span className="text-foreground">RD-20261009-0087114</span> | Records: 1 | Digest: 4a11b8...9e21
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === "cert" && (
          <div className="grid gap-6 md:grid-cols-2">
            <div className="rounded-xl border bg-card p-6 shadow-xs space-y-4">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                  <KeyRound className="h-5 w-5" />
                </div>
                <div>
                  <h3 className="font-semibold">{isThai ? "ใบรับรองดิจิทัลนิติบุคคล (CA Certificate)" : "Corporate CA Certificate"}</h3>
                  <p className="text-xs text-muted-foreground">Thailand National Root CA</p>
                </div>
              </div>

              <div className="space-y-2 text-sm border-t pt-3">
                <div className="flex justify-between py-1">
                  <span className="text-muted-foreground">{isThai ? "ผู้ออกใบรับรอง (Issuer)" : "Issuer"}:</span>
                  <span className="font-medium">Thailand Digital ID CA (INET)</span>
                </div>
                <div className="flex justify-between py-1">
                  <span className="text-muted-foreground">{isThai ? "ชื่อผู้ถือ (Subject)" : "Subject"}:</span>
                  <span className="font-medium">RUNGRUENG CONSTRUCTION MATERIALS CO., LTD.</span>
                </div>
                <div className="flex justify-between py-1">
                  <span className="text-muted-foreground">{isThai ? "เลขประจำตัวผู้เสียภาษี" : "Tax ID"}:</span>
                  <span className="font-mono">0105560012345</span>
                </div>
                <div className="flex justify-between py-1">
                  <span className="text-muted-foreground">{isThai ? "วันหมดอายุ" : "Valid Until"}:</span>
                  <span className="font-medium text-emerald-600 dark:text-emerald-400">2027-12-31 (อีก 447 วัน)</span>
                </div>
                <div className="flex justify-between py-1">
                  <span className="text-muted-foreground">{isThai ? "คีย์อัลกอริทึม" : "Algorithm"}:</span>
                  <span className="font-mono">RSA 2048-bit (SHA-256)</span>
                </div>
              </div>

              <button
                type="button"
                className="w-full rounded-lg border py-2 text-sm font-medium hover:bg-muted"
                onClick={() => alert(isThai ? "ตรวจสอบสถานะ CRL/OCSP: ใบรับรองสมบูรณ์" : "OCSP check: Certificate is VALID")}
              >
                {isThai ? "ตรวจสอบสถานะ CRL / OCSP" : "Check Revocation Status"}
              </button>
            </div>

            <div className="rounded-xl border bg-card p-6 shadow-xs space-y-4">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
                  <ShieldCheck className="h-5 w-5" />
                </div>
                <div>
                  <h3 className="font-semibold">{isThai ? "การจัดเก็บกุญแจส่วนตัว (Key Vault)" : "Private Key Protection"}</h3>
                  <p className="text-xs text-muted-foreground">FIPS 140-2 Level 3 Compliant</p>
                </div>
              </div>

              <div className="space-y-3 text-sm border-t pt-3">
                <p className="text-muted-foreground text-xs leading-relaxed">
                  {isThai
                    ? "กุญแจลงลายมือชื่อดิจิทัลถูกเก็บรักษาในระบบ Cloud Hardware Security Module (HSM) ปลอดภัยตามมาตรฐาน ETDA และข้อกำหนดของกรมสรรพากร"
                    : "Private keys are secured in FIPS 140-2 Level 3 Cloud HSM conforming to ETDA regulations."}
                </p>

                <div className="rounded-lg bg-muted/40 p-3 text-xs">
                  <div className="font-semibold">{isThai ? "สถานะการลงนามอัตโนมัติ" : "Auto-sign Policy"}</div>
                  <div className="mt-1 text-muted-foreground">
                    {isThai ? "ลงนามทันทีเมื่อผ่านการอนุมัติใบสำคัญขาย" : "Automatically sign when sales invoice is approved"}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === "config" && (
          <div className="max-w-2xl rounded-xl border bg-card p-6 shadow-xs space-y-6">
            <h2 className="text-base font-semibold">{isThai ? "การตั้งค่าการเชื่อมต่อกรมสรรพากร" : "RD Gateway Configuration"}</h2>

            <div className="space-y-4 text-sm">
              <div>
                <label className="block font-medium mb-1">{isThai ? "รูปแบบการส่งมอบ e-Tax" : "e-Tax Delivery Mode"}</label>
                <div className="grid grid-cols-2 gap-3">
                  <button
                    type="button"
                    onClick={() => setServiceMode("api")}
                    className={`rounded-lg border p-3 text-left transition-colors ${
                      serviceMode === "api" ? "border-primary bg-primary/5 ring-1 ring-primary" : "hover:bg-muted"
                    }`}
                  >
                    <div className="font-semibold">{isThai ? "e-Tax by Service Provider / API" : "Direct API / Service Provider"}</div>
                    <div className="text-xs text-muted-foreground mt-0.5">
                      {isThai ? "ส่งข้อมูล XML แบบ Real-time ผ่าน Gateway" : "Real-time XML submission via API Gateway"}
                    </div>
                  </button>

                  <button
                    type="button"
                    onClick={() => setServiceMode("email")}
                    className={`rounded-lg border p-3 text-left transition-colors ${
                      serviceMode === "email" ? "border-primary bg-primary/5 ring-1 ring-primary" : "hover:bg-muted"
                    }`}
                  >
                    <div className="font-semibold">{isThai ? "e-Tax by Email" : "e-Tax by Email"}</div>
                    <div className="text-xs text-muted-foreground mt-0.5">
                      {isThai ? "สำหรับนิติบุคคลส่งผ่าน csrc@etax.rd.go.th" : "For SMEs via csrc@etax.rd.go.th"}
                    </div>
                  </button>
                </div>
              </div>

              <div>
                <label className="block font-medium mb-1">{isThai ? "API Endpoint สรรพากร" : "RD API Endpoint"}</label>
                <input
                  type="text"
                  readOnly
                  value="https://etax.rd.go.th/gateway/v1/documents"
                  className="w-full rounded-lg border bg-muted/40 px-3 py-2 text-sm font-mono text-muted-foreground"
                />
              </div>

              <div>
                <label className="block font-medium mb-1">{isThai ? "อีเมลผู้รับเอกสาร e-Tax สำรอง" : "Fallback Notification Email"}</label>
                <input
                  type="email"
                  defaultValue="billing@rungrueng.co.th"
                  className="w-full rounded-lg border bg-background px-3 py-2 text-sm"
                />
              </div>

              <div className="flex items-center justify-between border-t pt-4">
                <div>
                  <div className="font-medium">{isThai ? "ลงลายมือชื่อดิจิทัลอัตโนมัติ (Auto-Sign)" : "Automatic Digital Signing"}</div>
                  <div className="text-xs text-muted-foreground">
                    {isThai ? "สร้างและลงนาม XML/PDF ทันทีที่อนุมัติเอกสาร" : "Sign XML/PDF immediately upon doc approval"}
                  </div>
                </div>
                <input
                  type="checkbox"
                  checked={autoSign}
                  onChange={(e) => setAutoSign(e.target.checked)}
                  className="h-5 w-5 rounded border-muted"
                />
              </div>
            </div>

            <div className="flex justify-end gap-2 border-t pt-4">
              <button
                type="button"
                className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90"
                onClick={() => alert(isThai ? "บันทึกการตั้งค่า e-Tax เรียบร้อย" : "e-Tax settings saved successfully")}
              >
                {isThai ? "บันทึกการตั้งค่า" : "Save Settings"}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
