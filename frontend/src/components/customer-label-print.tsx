"use client";

import React, { useState } from "react";
import {
  Printer,
  Search,
  CheckSquare,
  Square,
  AlertCircle,
  FileSpreadsheet,
  Settings,
  Users,
  MapPin,
  Phone,
  Building,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useBackendDictionary } from "@/components/backend-text-provider";
import { backendText } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";
import { apiFetch } from "@/lib/client-auth-session";

interface CustomerLabelItem {
  customer_code: string;
  name_th: string;
  name_en: string;
  tax_id: string;
  branch_no: string;
  billing_address: string;
  shipping_address: string;
  district: string;
  province: string;
  postcode: string;
  phone: string;
  total_sales: string;
  doc_count: number;
}

interface CustomerLabelPrintProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function CustomerLabelPrintScreen({
  language: _language = "th",
  holdingcode = "",
}: CustomerLabelPrintProps) {
  const dictionary = useBackendDictionary();

  // Filters - Left Column
  const [fromCode, setFromCode] = useState("");
  const [fromBirthday, setFromBirthday] = useState("");
  const [fromType, setFromType] = useState("");
  const [fromGroup, setFromGroup] = useState("");
  const [fromSalesperson, setFromSalesperson] = useState("");
  const [fromArea, setFromArea] = useState("");
  const [fromDate, setFromDate] = useState("");
  const [fromAmount, setFromAmount] = useState("");

  // Filters - Right Column
  const [toCode, setToCode] = useState("");
  const [toBirthday, setToBirthday] = useState("");
  const [toType, setToType] = useState("");
  const [toGroup, setToGroup] = useState("");
  const [toSalesperson, setToSalesperson] = useState("");
  const [toArea, setToArea] = useState("");
  const [toDate, setToDate] = useState("");
  const [toAmount, setToAmount] = useState("");

  // Label configuration
  const [labelsPerPage, setLabelsPerPage] = useState<16 | 24 | 30>(24);
  const [addressSource, setAddressSource] = useState<"shipping" | "billing">("shipping");

  // State
  const [items, setItems] = useState<CustomerLabelItem[]>([]);
  const [selectedCodes, setSelectedCodes] = useState<Set<string>>(new Set());
  const [isLoading, setIsLoading] = useState(false);
  const [notice, setNotice] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);

  // Fetch customer list
  async function handleShowData() {
    setIsLoading(true);
    setNotice(null);
    try {
      const payload = {
        holdingcode,
        from_code: fromCode,
        to_code: toCode,
        from_date: fromDate,
        to_date: toDate,
        from_amount: fromAmount,
        to_amount: toAmount,
        limit: 1000,
      };

      const res = await apiFetch("/api/customer/labels/query", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        throw new Error("HTTP " + res.status);
      }

      const resData = await res.json();
      const list: CustomerLabelItem[] = resData.data || [];
      setItems(list);
      setSelectedCodes(new Set(list.map((c) => c.customer_code)));

      if (list.length === 0) {
        setNotice({
          type: "info",
          message: backendText(dictionary, "ops_msg_no_customer_found", "ไม่พบข้อมูลลูกค้าตามเงื่อนไขที่ระบุ"),
        });
      }
    } catch (err: unknown) {
      setNotice({
        type: "error",
        message: err instanceof Error ? err.message : "ดึงข้อมูลลูกค้าไม่สำเร็จ",
      });
    } finally {
      setIsLoading(false);
    }
  }

  // Selection
  function toggleCustomer(code: string) {
    setSelectedCodes((prev) => {
      const next = new Set(prev);
      if (next.has(code)) {
        next.delete(code);
      } else {
        next.add(code);
      }
      return next;
    });
  }

  function handleSelectAll() {
    setSelectedCodes(new Set(items.map((c) => c.customer_code)));
  }

  function handleDeselectAll() {
    setSelectedCodes(new Set());
  }

  // Print Handler
  function handlePrint() {
    if (selectedCodes.size === 0) {
      setNotice({
        type: "error",
        message: backendText(dictionary, "customer_label_err_no_select", "กรุณาเลือกรายชื่อลูกค้าที่ต้องการพิมพ์ป้ายผนึก"),
      });
      return;
    }

    document.body.classList.add("customer-address-label-printing");
    window.setTimeout(() => {
      document.body.classList.remove("customer-address-label-printing");
    }, 1500);

    window.print();
  }

  const selectedList = items.filter((item) => selectedCodes.has(item.customer_code));

  return (
    <div className="flex h-full w-full flex-col gap-4 overflow-y-auto p-4 md:p-6">
      {/* Screen Header */}
      <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-foreground sm:text-2xl">
              {backendText(dictionary, "customer_label_print_title", "พิมพ์ป้ายผนึกชื่อลูกค้า-สมาชิก")}
            </h1>
            <Badge variant="outline" className="border-primary/30 text-primary">
              Champ Parity
            </Badge>
          </div>
          <p className="text-sm text-muted-foreground">
            {backendText(
              dictionary,
              "customer_label_print_desc",
              "พิมพ์สติกเกอร์จ่าหน้าซองและกล่องพัสดุสำหรับส่งเอกสารและสินค้าให้ลูกค้าและสมาชิก"
            )}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="secondary" className="px-3 py-1 font-mono text-xs">
            {selectedList.length} / {items.length} รายการที่เลือก
          </Badge>
        </div>
      </div>

      {/* Filter Card (2 Columns matching Champ Spec) */}
      <div className="rounded-2xl border border-border bg-card p-4 shadow-sm md:p-5">
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          {/* Column 1: Left Filters */}
          <div className="flex flex-col gap-3">
            <h3 className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
              เงื่อนไขเริ่มต้น (From)
            </h3>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_from_code", "จากรหัสลูกค้า")}
                </label>
                <Input
                  placeholder="CUST-001"
                  value={fromCode}
                  onChange={(e) => setFromCode(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_from_birthday", "พิมพ์จากวันเกิด")}
                </label>
                <Input
                  type="date"
                  value={fromBirthday}
                  onChange={(e) => setFromBirthday(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>

              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_from_type", "พิมพ์จากประเภทลูกค้า")}
                </label>
                <Input
                  placeholder="ทั่วไป / นิติบุคคล"
                  value={fromType}
                  onChange={(e) => setFromType(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_from_group", "พิมพ์จากกลุ่มลูกค้า")}
                </label>
                <Input
                  placeholder="กลุ่มค้าปลีก / ค้าส่ง"
                  value={fromGroup}
                  onChange={(e) => setFromGroup(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>

              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_from_salesperson", "พิมพ์จากพนักงานขาย")}
                </label>
                <Input
                  placeholder="พนักงานขายเริ่มต้น"
                  value={fromSalesperson}
                  onChange={(e) => setFromSalesperson(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_from_area", "พิมพ์จากเขตการขาย")}
                </label>
                <Input
                  placeholder="กรุงเทพฯ / ภาคกลาง"
                  value={fromArea}
                  onChange={(e) => setFromArea(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>

              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_from_sale_date", "พิมพ์จากวันที่ขาย")}
                </label>
                <Input
                  type="date"
                  value={fromDate}
                  onChange={(e) => setFromDate(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_from_sales_amt", "พิมพ์จากยอดขาย")}
                </label>
                <Input
                  type="number"
                  placeholder="0.00"
                  value={fromAmount}
                  onChange={(e) => setFromAmount(e.target.value)}
                  className="h-9 text-sm font-mono"
                />
              </div>
            </div>
          </div>

          {/* Column 2: Right Filters */}
          <div className="flex flex-col gap-3">
            <h3 className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
              เงื่อนไขสิ้นสุดและรูปแบบสติกเกอร์ (To & Template)
            </h3>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_to_code", "ถึงรหัสลูกค้า")}
                </label>
                <Input
                  placeholder="CUST-999"
                  value={toCode}
                  onChange={(e) => setToCode(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_to_birthday", "พิมพ์ถึงวันเกิด")}
                </label>
                <Input
                  type="date"
                  value={toBirthday}
                  onChange={(e) => setToBirthday(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>

              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_to_type", "พิมพ์ถึงประเภทลูกค้า")}
                </label>
                <Input
                  placeholder="ทั่วไป / นิติบุคคล"
                  value={toType}
                  onChange={(e) => setToType(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_to_group", "พิมพ์ถึงกลุ่มลูกค้า")}
                </label>
                <Input
                  placeholder="กลุ่มค้าปลีก / ค้าส่ง"
                  value={toGroup}
                  onChange={(e) => setToGroup(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>

              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_to_salesperson", "พิมพ์ถึงพนักงานขาย")}
                </label>
                <Input
                  placeholder="พนักงานขายสิ้นสุด"
                  value={toSalesperson}
                  onChange={(e) => setToSalesperson(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_to_area", "พิมพ์ถึงเขตการขาย")}
                </label>
                <Input
                  placeholder="กรุงเทพฯ / ภาคกลาง"
                  value={toArea}
                  onChange={(e) => setToArea(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>

              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_to_sale_date", "พิมพ์ถึงวันที่ขาย")}
                </label>
                <Input
                  type="date"
                  value={toDate}
                  onChange={(e) => setToDate(e.target.value)}
                  className="h-9 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_to_sales_amt", "พิมพ์ถึงยอดขาย")}
                </label>
                <Input
                  type="number"
                  placeholder="9999999.00"
                  value={toAmount}
                  onChange={(e) => setToAmount(e.target.value)}
                  className="h-9 text-sm font-mono"
                />
              </div>
            </div>

            {/* Template & Labels Per Page */}
            <div className="mt-2 grid grid-cols-1 gap-3 rounded-xl bg-muted/40 p-3 sm:grid-cols-2">
              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  {backendText(dictionary, "customer_label_labels_per_page", "จำนวนรายการ Label/หน้า")}
                </label>
                <select
                  value={labelsPerPage}
                  onChange={(e) => setLabelsPerPage(Number(e.target.value) as 16 | 24 | 30)}
                  className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
                >
                  <option value={16}>16 ป้าย / หน้า (A4 2x8)</option>
                  <option value={24}>24 ป้าย / หน้า (A4 3x8 - มาตรฐาน)</option>
                  <option value={30}>30 ป้าย / หน้า (A4 3x10)</option>
                </select>
              </div>

              <div>
                <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                  ที่อยู่ที่ใช้พิมพ์
                </label>
                <select
                  value={addressSource}
                  onChange={(e) => setAddressSource(e.target.value as "shipping" | "billing")}
                  className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
                >
                  <option value="shipping">ที่อยู่จัดส่ง / ปัจจุบัน (แนะนำ)</option>
                  <option value="billing">ที่อยู่ออกใบกำกับภาษี</option>
                </select>
              </div>
            </div>
          </div>
        </div>

        {/* Buttons Bar */}
        <div className="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-border/60 pt-4">
          <div className="flex items-center gap-3">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleSelectAll}
              disabled={items.length === 0}
              className="h-9 gap-1.5"
            >
              <CheckSquare className="h-4 w-4" />
              เลือกทั้งหมด
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleDeselectAll}
              disabled={items.length === 0}
              className="h-9 gap-1.5"
            >
              <Square className="h-4 w-4" />
              ไม่เลือกทั้งหมด
            </Button>
          </div>

          <div className="flex items-center gap-2.5">
            <Button
              type="button"
              variant="outline"
              onClick={handleShowData}
              disabled={isLoading}
              className="h-10 min-w-[120px] gap-2 border-border shadow-xs hover:border-primary/50"
            >
              {isLoading ? (
                <RefreshCw className="h-4 w-4 animate-spin text-primary" />
              ) : (
                <Search className="h-4 w-4 text-primary" />
              )}
              {backendText(dictionary, "customer_label_btn_preview", "แสดงข้อมูล")}
            </Button>

            <Button
              type="button"
              variant="default"
              onClick={handlePrint}
              disabled={selectedList.length === 0}
              className="h-10 min-w-[130px] gap-2 shadow-xs"
            >
              <Printer className="h-4 w-4" />
              {backendText(dictionary, "customer_label_btn_print", "พิมพ์ป้ายผนึก")}
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

      {/* Customer Data Table */}
      <div className="flex-1 overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
        <div className="h-full overflow-auto">
          <table className="w-full border-collapse text-left text-sm">
            <thead className="sticky top-0 z-10 border-b border-border bg-muted/60 text-xs font-semibold text-muted-foreground backdrop-blur-sm">
              <tr>
                <th className="w-10 px-3 py-3 text-center">
                  <span className="sr-only">เลือก</span>
                </th>
                <th className="px-3 py-3">รหัสลูกค้า</th>
                <th className="px-3 py-3">ชื่อลูกค้า / บริษัท</th>
                <th className="px-3 py-3">ชื่อภาษาอังกฤษ</th>
                <th className="px-3 py-3">ที่อยู่ออกบิล</th>
                <th className="px-3 py-3">ที่อยู่ปัจจุบัน / จัดส่ง</th>
                <th className="px-3 py-3 text-right">ยอดขายรวม</th>
                <th className="px-3 py-3 text-center">จำนวนบิล</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {items.length === 0 ? (
                <tr>
                  <td colSpan={8} className="py-12 text-center text-muted-foreground">
                    {isLoading
                      ? "กำลังค้นหาข้อมูลลูกค้า..."
                      : backendText(dictionary, "ops_msg_click_preview_to_show_customers", "กดปุ่ม [แสดงข้อมูล] เพื่อสืบค้นรายชื่อลูกค้าสำหรับพิมพ์ป้ายผนึก")}
                  </td>
                </tr>
              ) : (
                items.map((item) => {
                  const isChecked = selectedCodes.has(item.customer_code);
                  return (
                    <tr
                      key={item.customer_code}
                      onClick={() => toggleCustomer(item.customer_code)}
                      className={`cursor-pointer transition-colors hover:bg-muted/40 ${
                        isChecked ? "bg-primary/5 dark:bg-primary/10" : ""
                      }`}
                    >
                      <td className="px-3 py-3 text-center" onClick={(e) => e.stopPropagation()}>
                        <button
                          type="button"
                          onClick={() => toggleCustomer(item.customer_code)}
                          className="text-primary hover:opacity-80"
                        >
                          {isChecked ? (
                            <CheckSquare className="h-4 w-4" />
                          ) : (
                            <Square className="h-4 w-4 text-muted-foreground" />
                          )}
                        </button>
                      </td>
                      <td className="px-3 py-3 font-medium text-foreground">{item.customer_code}</td>
                      <td className="max-w-[200px] truncate px-3 py-3 text-foreground" title={item.name_th}>
                        {item.name_th}
                      </td>
                      <td className="max-w-[180px] truncate px-3 py-3 text-muted-foreground" title={item.name_en}>
                        {item.name_en || "-"}
                      </td>
                      <td className="max-w-[220px] truncate px-3 py-3 text-muted-foreground" title={item.billing_address}>
                        {item.billing_address || "-"}
                      </td>
                      <td className="max-w-[220px] truncate px-3 py-3 text-foreground" title={item.shipping_address}>
                        {item.shipping_address || item.billing_address || "-"}
                      </td>
                      <td className="px-3 py-3 text-right font-mono text-foreground">
                        {Number(item.total_sales).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="px-3 py-3 text-center font-mono text-muted-foreground">
                        {item.doc_count}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Print-Only Area: A4 Label Sheet for Sticker Printing */}
      <div className="hidden print:block print:w-full print:bg-white print:text-black">
        <style jsx global>{`
          @media print {
            body * {
              visibility: hidden !important;
            }
            .customer-address-label-sheet,
            .customer-address-label-sheet * {
              visibility: visible !important;
            }
            .customer-address-label-sheet {
              position: absolute !important;
              left: 0 !important;
              top: 0 !important;
              width: 100% !important;
              display: grid !important;
              background: white !important;
              color: black !important;
              box-sizing: border-box !important;
            }
            @page {
              size: A4 portrait;
              margin: 8mm 6mm;
            }
          }
        `}</style>

        <div
          className="customer-address-label-sheet"
          style={{
            gridTemplateColumns:
              labelsPerPage === 16
                ? "repeat(2, 99mm)"
                : labelsPerPage === 30
                ? "repeat(3, 66mm)"
                : "repeat(3, 66mm)",
            gap: "2mm",
          }}
        >
          {selectedList.map((item) => {
            const rawAddr =
              addressSource === "shipping"
                ? item.shipping_address || item.billing_address
                : item.billing_address || item.shipping_address;

            const fullAddr = [rawAddr, item.district, item.province, item.postcode]
              .filter(Boolean)
              .join(" ");

            return (
              <div
                key={item.customer_code}
                style={{
                  minHeight: labelsPerPage === 16 ? "35mm" : labelsPerPage === 30 ? "26mm" : "34mm",
                  border: "1px dashed #bbb",
                  padding: "3mm 4mm",
                  display: "flex",
                  flexDirection: "column",
                  justifyContent: "space-between",
                  boxSizing: "border-box",
                  pageBreakInside: "avoid",
                  fontFamily: "Sarabun, Tahoma, sans-serif",
                }}
              >
                <div>
                  <div style={{ fontSize: "8pt", color: "#666", marginBottom: "1mm" }}>
                    กรุณาส่ง: [{item.customer_code}]
                  </div>
                  <div style={{ fontSize: "10pt", fontWeight: "bold", lineHeight: 1.25, color: "#111" }}>
                    {item.name_th}
                  </div>
                  {item.name_en && (
                    <div style={{ fontSize: "8.5pt", color: "#444", marginTop: "0.5mm" }}>
                      {item.name_en}
                    </div>
                  )}
                  <div style={{ fontSize: "8.5pt", lineHeight: 1.35, marginTop: "1.5mm", color: "#222" }}>
                    {fullAddr || "ไม่มีข้อมูลที่อยู่"}
                  </div>
                </div>

                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginTop: "2mm" }}>
                  {item.phone && (
                    <div style={{ fontSize: "8pt", color: "#555" }}>
                      โทร: {item.phone}
                    </div>
                  )}
                  {item.postcode && (
                    <div style={{ fontSize: "11pt", fontWeight: "bold", letterSpacing: "1px", color: "#000" }}>
                      {item.postcode}
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
