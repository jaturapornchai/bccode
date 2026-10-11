"use client";

import React, { useState } from "react";
import {
  Printer,
  Search,
  CheckSquare,
  Square,
  AlertCircle,
  RotateCcw,
  Calendar,
  FileText,
  UserCheck,
  Building2,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useBackendDictionary } from "@/components/backend-text-provider";
import { backendText } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";
import { apiFetch } from "@/lib/client-auth-session";

interface BillingNotePrintItem {
  docno: string;
  docdate: string;
  custcode: string;
  custname: string;
  total_amount: string;
  remark: string;
  sale_code: string;
  inv_count: number;
}

interface BillingNoteAutoPrintProps {
  language?: LanguageCode;
  holdingcode?: string;
  companyName?: string;
}

export function BillingNoteAutoPrintScreen({
  language: _language = "th",
  holdingcode: _holdingcode = "",
  companyName = "บริษัท ตัวอย่าง จำกัด",
}: BillingNoteAutoPrintProps) {
  const dictionary = useBackendDictionary();

  // Filters
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const [fromDocNo, setFromDocNo] = useState("");
  const [toDocNo, setToDocNo] = useState("");
  const [fromCust, setFromCust] = useState("");
  const [toCust, setToCust] = useState("");
  const [saleCode, setSaleCode] = useState("");

  // Options
  const [reprint, setReprint] = useState(false);
  const [selectedDocs, setSelectedDocs] = useState<Set<string>>(new Set());

  // Data & Status
  const [items, setItems] = useState<BillingNotePrintItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [notice, setNotice] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);

  // Query billing notes
  async function handleShowBill() {
    setIsLoading(true);
    setNotice(null);
    try {
      const params = new URLSearchParams();
      if (fromDate) params.append("from_date", fromDate);
      if (toDate) params.append("to_date", toDate);
      if (fromDocNo) params.append("from_docno", fromDocNo);
      if (toDocNo) params.append("to_docno", toDocNo);
      if (fromCust) params.append("from_cust", fromCust);
      if (toCust) params.append("to_cust", toCust);
      if (saleCode) params.append("sale_code", saleCode);

      const res = await apiFetch(`/api/erp-transaction/billingnote/auto-print?${params.toString()}`);
      if (!res.ok) {
        throw new Error("HTTP " + res.status);
      }
      const data = await res.json();
      const list: BillingNotePrintItem[] = data.data || [];
      setItems(list);
      setSelectedDocs(new Set(list.map((i) => i.docno)));

      if (list.length === 0) {
        setNotice({
          type: "info",
          message: backendText(dictionary, "ops_msg_no_billing_notes", "ไม่พบรายการใบวางบิลตามเงื่อนไขที่ระบุ"),
        });
      }
    } catch (err: unknown) {
      setNotice({
        type: "error",
        message: err instanceof Error ? err.message : "โหลดข้อมูลใบวางบิลไม่สำเร็จ",
      });
    } finally {
      setIsLoading(false);
    }
  }

  // Selection
  function handleSelectAll() {
    setSelectedDocs(new Set(items.map((i) => i.docno)));
  }

  function handleDeselectAll() {
    setSelectedDocs(new Set());
  }

  function toggleDoc(docno: string) {
    setSelectedDocs((prev) => {
      const next = new Set(prev);
      if (next.has(docno)) next.delete(docno);
      else next.add(docno);
      return next;
    });
  }

  function handleClear() {
    setItems([]);
    setSelectedDocs(new Set());
    setFromDocNo("");
    setToDocNo("");
    setFromCust("");
    setToCust("");
    setFromDate("");
    setToDate("");
    setSaleCode("");
    setNotice(null);
  }

  function handlePrint() {
    if (selectedDocs.size === 0) {
      setNotice({
        type: "error",
        message: backendText(dictionary, "billing_note_select_required", "กรุณาเลือกใบวางบิลที่ต้องการพิมพ์อย่างน้อยหนึ่งรายการ"),
      });
      return;
    }

    document.body.classList.add("billing-note-batch-printing");
    window.setTimeout(() => {
      document.body.classList.remove("billing-note-batch-printing");
    }, 1500);

    window.print();
  }

  const selectedList = items.filter((item) => selectedDocs.has(item.docno));

  return (
    <div className="flex h-full w-full flex-col gap-4 overflow-y-auto p-4 md:p-6">
      {/* Header */}
      <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-foreground sm:text-2xl">
              {backendText(dictionary, "billing_note_auto_print_title", "พิมพ์ใบวางบิลอัตโนมัติ (Batch Billing Note Print)")}
            </h1>
            <Badge variant="outline" className="border-primary/30 text-primary">
              Champ Parity
            </Badge>
          </div>
          <p className="text-sm text-muted-foreground">
            {backendText(
              dictionary,
              "billing_note_auto_print_desc",
              "สืบค้นและพิมพ์ใบวางบิลเป็นชุดตามรอบวางบิลและกลุ่มลูกหนี้"
            )}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="secondary" className="px-3 py-1 font-mono text-xs">
            {selectedList.length} / {items.length} ใบวางบิลที่เลือก
          </Badge>
        </div>
      </div>

      {/* Filter Card */}
      <div className="rounded-2xl border border-border bg-card p-4 shadow-sm md:p-5">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {/* Date Range */}
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <Calendar className="h-3.5 w-3.5" />
              {backendText(dictionary, "purchase_receipt_adj_from_date", "จากวันที่")}
            </label>
            <Input
              type="date"
              value={fromDate}
              onChange={(e) => setFromDate(e.target.value)}
              className="h-10 text-sm"
            />
          </div>
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <Calendar className="h-3.5 w-3.5" />
              {backendText(dictionary, "purchase_receipt_adj_to_date", "ถึงวันที่")}
            </label>
            <Input
              type="date"
              value={toDate}
              onChange={(e) => setToDate(e.target.value)}
              className="h-10 text-sm"
            />
          </div>

          {/* Doc No Range */}
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <FileText className="h-3.5 w-3.5" />
              จากเลขที่ใบวางบิล
            </label>
            <Input
              placeholder="BL..."
              value={fromDocNo}
              onChange={(e) => setFromDocNo(e.target.value)}
              className="h-10 text-sm"
            />
          </div>
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <FileText className="h-3.5 w-3.5" />
              ถึงเลขที่ใบวางบิล
            </label>
            <Input
              placeholder="BL..."
              value={toDocNo}
              onChange={(e) => setToDocNo(e.target.value)}
              className="h-10 text-sm"
            />
          </div>

          {/* Customer Range */}
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <UserCheck className="h-3.5 w-3.5" />
              จากรหัสลูกหนี้
            </label>
            <Input
              placeholder="AR..."
              value={fromCust}
              onChange={(e) => setFromCust(e.target.value)}
              className="h-10 text-sm"
            />
          </div>
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <UserCheck className="h-3.5 w-3.5" />
              ถึงรหัสลูกหนี้
            </label>
            <Input
              placeholder="AR..."
              value={toCust}
              onChange={(e) => setToCust(e.target.value)}
              className="h-10 text-sm"
            />
          </div>

          {/* Salesperson & Options */}
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <Building2 className="h-3.5 w-3.5" />
              พนักงานขาย
            </label>
            <Input
              placeholder="SALE..."
              value={saleCode}
              onChange={(e) => setSaleCode(e.target.value)}
              className="h-10 text-sm"
            />
          </div>

          <div className="flex items-center gap-4 pt-5 sm:col-span-2">
            <label className="flex cursor-pointer items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={reprint}
                onChange={(e) => setReprint(e.target.checked)}
                className="h-4 w-4 rounded accent-primary"
              />
              <span className="text-foreground">พิมพ์ซ้ำ (RePrint)</span>
            </label>
            <label className="flex cursor-pointer items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={items.length > 0 && selectedDocs.size === items.length}
                onChange={(e) => {
                  if (e.target.checked) handleSelectAll();
                  else handleDeselectAll();
                }}
                className="h-4 w-4 rounded accent-primary"
              />
              <span className="text-foreground">เลือกทั้งหมด (Select All)</span>
            </label>
          </div>
        </div>

        {/* Action Controls */}
        <div className="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-border/60 pt-4">
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleClear}
              className="h-9 gap-1.5 text-muted-foreground"
            >
              <RotateCcw className="h-3.5 w-3.5" />
              ล้างหน้าจอ
            </Button>
          </div>

          <div className="flex items-center gap-2.5">
            <Button
              type="button"
              variant="outline"
              onClick={handleShowBill}
              disabled={isLoading}
              className="h-10 min-w-[120px] gap-2 border-border shadow-xs hover:border-primary/50"
            >
              {isLoading ? (
                <RefreshCw className="h-4 w-4 animate-spin text-primary" />
              ) : (
                <Search className="h-4 w-4 text-primary" />
              )}
              แสดงข้อมูล
            </Button>

            <Button
              type="button"
              variant="default"
              onClick={handlePrint}
              disabled={selectedList.length === 0}
              className="h-10 min-w-[130px] gap-2 shadow-xs"
            >
              <Printer className="h-4 w-4" />
              พิมพ์ใบวางบิล
              {selectedList.length > 0 && (
                <span className="rounded-full bg-primary-foreground/20 px-1.5 py-0.5 text-xs font-bold">
                  {selectedList.length}
                </span>
              )}
            </Button>
          </div>
        </div>
      </div>

      {/* Notice Message */}
      {notice && (
        <div
          className={`flex items-center gap-2.5 rounded-xl border p-3.5 text-sm ${
            notice.type === "success"
              ? "border-emerald-500/30 bg-emerald-50 text-emerald-800 dark:bg-emerald-950/30 dark:text-emerald-300"
              : notice.type === "error"
              ? "border-rose-500/30 bg-rose-50 text-rose-800 dark:bg-rose-950/30 dark:text-rose-300"
              : "border-blue-500/30 bg-blue-50 text-blue-800 dark:bg-blue-950/30 dark:text-blue-300"
          }`}
        >
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{notice.message}</span>
        </div>
      )}

      {/* Data Table */}
      <div className="flex-1 overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
        <div className="h-full overflow-auto">
          <table className="w-full border-collapse text-left text-sm">
            <thead className="sticky top-0 z-10 border-b border-border bg-muted/60 text-xs font-semibold text-muted-foreground backdrop-blur-sm">
              <tr>
                <th className="w-10 px-3 py-3 text-center">
                  <span className="sr-only">เลือก</span>
                </th>
                <th className="px-3 py-3">วันที่วางบิล</th>
                <th className="px-3 py-3">เลขที่ใบวางบิล</th>
                <th className="px-3 py-3">รหัสลูกหนี้</th>
                <th className="px-3 py-3">ชื่อลูกหนี้</th>
                <th className="px-3 py-3 text-right">ยอดเงินวางบิล</th>
                <th className="px-3 py-3 text-center">จำนวนบิล</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {items.length === 0 ? (
                <tr>
                  <td colSpan={7} className="py-12 text-center text-muted-foreground">
                    {isLoading
                      ? "กำลังโหลดข้อมูล..."
                      : "กดปุ่ม [แสดงข้อมูล] เพื่อสืบค้นรายการใบวางบิลสำหรับพิมพ์"}
                  </td>
                </tr>
              ) : (
                items.map((item) => {
                  const isChecked = selectedDocs.has(item.docno);
                  return (
                    <tr
                      key={item.docno}
                      onClick={() => toggleDoc(item.docno)}
                      className={`cursor-pointer transition-colors hover:bg-muted/40 ${
                        isChecked ? "bg-primary/5 dark:bg-primary/10" : ""
                      }`}
                    >
                      <td className="px-3 py-3 text-center" onClick={(e) => e.stopPropagation()}>
                        <button
                          type="button"
                          onClick={() => toggleDoc(item.docno)}
                          className="text-primary hover:opacity-80"
                        >
                          {isChecked ? (
                            <CheckSquare className="h-4 w-4" />
                          ) : (
                            <Square className="h-4 w-4 text-muted-foreground" />
                          )}
                        </button>
                      </td>
                      <td className="px-3 py-3 text-muted-foreground">{item.docdate}</td>
                      <td className="px-3 py-3 font-medium text-foreground">{item.docno}</td>
                      <td className="px-3 py-3 text-muted-foreground">{item.custcode}</td>
                      <td className="max-w-[240px] truncate px-3 py-3 text-foreground" title={item.custname}>
                        {item.custname}
                      </td>
                      <td className="px-3 py-3 text-right font-mono text-foreground">
                        {Number(item.total_amount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="px-3 py-3 text-center font-mono text-muted-foreground">
                        {item.inv_count}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Print-Only View: Official Thai Billing Note Sheet */}
      <div className="hidden print:block print:w-full print:bg-white print:text-black">
        <style jsx global>{`
          @media print {
            body * {
              visibility: hidden !important;
            }
            .billing-note-batch-sheet,
            .billing-note-batch-sheet * {
              visibility: visible !important;
            }
            .billing-note-batch-sheet {
              position: absolute !important;
              left: 0 !important;
              top: 0 !important;
              width: 100% !important;
              background: white !important;
              color: black !important;
            }
            .billing-note-page {
              page-break-after: always;
              min-height: 270mm;
              padding: 12mm 15mm;
              box-sizing: border-box;
              font-family: Sarabun, Tahoma, sans-serif;
            }
            @page {
              size: A4 portrait;
              margin: 0;
            }
          }
        `}</style>

        <div className="billing-note-batch-sheet">
          {selectedList.map((item) => (
            <div key={item.docno} className="billing-note-page">
              {/* Header */}
              <div className="flex justify-between border-b pb-4">
                <div>
                  <h2 className="text-xl font-bold">{companyName}</h2>
                  <p className="text-xs text-gray-600">เอกสารออกโดยระบบบัญชี BC Ai Account</p>
                </div>
                <div className="text-right">
                  <h1 className="text-2xl font-bold text-gray-900">ใบวางบิล</h1>
                  <p className="text-sm font-semibold tracking-wider text-gray-700">BILLING NOTE</p>
                </div>
              </div>

              {/* Doc & Customer Details */}
              <div className="mt-4 grid grid-cols-2 gap-4 rounded border p-3 text-sm">
                <div>
                  <p><span className="font-semibold">รหัสลูกค้า:</span> {item.custcode}</p>
                  <p><span className="font-semibold">ชื่อลูกค้า:</span> {item.custname}</p>
                  {item.remark && <p><span className="font-semibold">หมายเหตุ:</span> {item.remark}</p>}
                </div>
                <div className="text-right">
                  <p><span className="font-semibold">เลขที่เอกสาร:</span> {item.docno}</p>
                  <p><span className="font-semibold">วันที่วางบิล:</span> {item.docdate}</p>
                  <p><span className="font-semibold">พนักงานขาย:</span> {item.sale_code || "-"}</p>
                </div>
              </div>

              {/* Bill Details Table */}
              <div className="mt-6">
                <table className="w-full border-collapse text-sm">
                  <thead>
                    <tr className="border-b-2 border-t-2 border-gray-800 bg-gray-50 text-left text-xs">
                      <th className="py-2 pl-2">ลำดับ</th>
                      <th className="py-2">รายการบิล / ใบกำกับภาษี</th>
                      <th className="py-2 text-center">วันที่เอกสาร</th>
                      <th className="py-2 text-right pr-2">จำนวนเงิน (บาท)</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-b border-gray-300">
                      <td className="py-3 pl-2">1</td>
                      <td className="py-3 font-medium">รวมยอดวางบิลตามเอกสาร {item.docno} ({item.inv_count} ฉบับ)</td>
                      <td className="py-3 text-center">{item.docdate}</td>
                      <td className="py-3 text-right pr-2 font-mono">
                        {Number(item.total_amount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                    </tr>
                  </tbody>
                  <tfoot>
                    <tr className="border-b-2 border-t-2 border-gray-800 font-bold">
                      <td colSpan={3} className="py-2 pl-2 text-right">ยอดรวมวางบิลทั้งสิ้น:</td>
                      <td className="py-2 pr-2 text-right font-mono text-base">
                        {Number(item.total_amount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                    </tr>
                  </tfoot>
                </table>
              </div>

              {/* Signatures */}
              <div className="mt-20 grid grid-cols-3 gap-6 text-center text-xs">
                <div className="rounded border p-4 pt-12">
                  <div className="border-t border-dashed pt-2">ผู้ส่งวางบิล</div>
                  <div className="mt-1 text-gray-500">วันที่: ____/____/________</div>
                </div>
                <div className="rounded border p-4 pt-12">
                  <div className="border-t border-dashed pt-2">ผู้รับวางบิล</div>
                  <div className="mt-1 text-gray-500">วันที่: ____/____/________</div>
                </div>
                <div className="rounded border p-4 pt-12">
                  <div className="border-t border-dashed pt-2">นัดรับเช็ค / ชำระเงิน</div>
                  <div className="mt-1 text-gray-500">วันที่: ____/____/________</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
