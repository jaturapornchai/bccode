"use client";

import React, { useState } from "react";
import {
  FileCheck2,
  RefreshCw,
  Search,
  CheckSquare,
  Square,
  AlertCircle,
  Building2,
  Calendar,
  FileText,
  UserCheck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useBackendDictionary } from "@/components/backend-text-provider";
import { backendText } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";
import { apiFetch } from "@/lib/client-auth-session";

interface ReceiptAdjustmentItem {
  docno: string;
  docdate: string;
  vendor_code: string;
  vendor_name: string;
  ir_ref_no: string;
  ir_ref_date: string;
  gr_qty: string;
  ir_qty: string;
  is_matched: boolean;
  is_closed: boolean;
}

interface PurchaseReceiptAdjustmentProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function PurchaseReceiptAdjustmentScreen({
  language: _language = "th",
  holdingcode = "",
}: PurchaseReceiptAdjustmentProps) {
  const dictionary = useBackendDictionary();

  // Filters
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const [fromDocNo, setFromDocNo] = useState("");
  const [toDocNo, setToDocNo] = useState("");
  const [fromVendor, setFromVendor] = useState("");
  const [toVendor, setToVendor] = useState("");
  const [fromDept, setFromDept] = useState("");
  const [toDept, setToDept] = useState("");

  // Data & Selection
  const [items, setItems] = useState<ReceiptAdjustmentItem[]>([]);
  const [selectedDocs, setSelectedDocs] = useState<Set<string>>(new Set());
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [notice, setNotice] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);
  const [showConfirm, setShowConfirm] = useState(false);

  // Query / Process function
  async function handleProcess() {
    setIsLoading(true);
    setNotice(null);
    try {
      const params = new URLSearchParams();
      if (fromDate) params.append("from_date", fromDate);
      if (toDate) params.append("to_date", toDate);
      if (fromDocNo) params.append("from_docno", fromDocNo);
      if (toDocNo) params.append("to_docno", toDocNo);
      if (fromVendor) params.append("from_vendor", fromVendor);
      if (toVendor) params.append("to_vendor", toVendor);
      if (fromDept) params.append("from_dept", fromDept);
      if (toDept) params.append("to_dept", toDept);

      const res = await apiFetch(`/api/erp-transaction/goodsreceive/receipt-adjustments?${params.toString()}`);
      if (!res.ok) {
        throw new Error("HTTP " + res.status);
      }
      const data = await res.json();
      const list: ReceiptAdjustmentItem[] = data.data || data.items || [];
      setItems(list);
      setSelectedDocs(new Set());

      if (list.length === 0) {
        setNotice({
          type: "info",
          message: backendText(dictionary, "ops_msg_no_pending_documents", "ไม่พบรายการใบรับสินค้าที่ตรงตามเงื่อนไข"),
        });
      }
    } catch (err: unknown) {
      setNotice({
        type: "error",
        message: err instanceof Error ? err.message : "โหลดข้อมูลไม่สำเร็จ",
      });
    } finally {
      setIsLoading(false);
    }
  }

  // Selection handlers
  function handleSelectAll() {
    setSelectedDocs(new Set(items.map((i) => i.docno)));
  }

  function handleDeselectAll() {
    setSelectedDocs(new Set());
  }

  function toggleDoc(docno: string) {
    setSelectedDocs((prev) => {
      const next = new Set(prev);
      if (next.has(docno)) {
        next.delete(docno);
      } else {
        next.add(docno);
      }
      return next;
    });
  }

  // Save / Adjust function
  async function handleAdjustSave() {
    if (selectedDocs.size === 0) {
      setNotice({
        type: "error",
        message: backendText(dictionary, "purchase_receipt_adj_no_selection", "กรุณาเลือกรายการใบรับสินค้าที่ต้องการปรับปรุง"),
      });
      return;
    }

    setIsSaving(true);
    setNotice(null);
    setShowConfirm(false);

    try {
      const res = await apiFetch("/api/erp-transaction/goodsreceive/receipt-adjustments", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          docnos: Array.from(selectedDocs),
        }),
      });

      if (!res.ok) {
        throw new Error("HTTP " + res.status);
      }

      const result = await res.json();
      // BFF ส่งต่อเหตุผลที่ผู้ใช้แก้ได้เป็น HTTP 200 + success:false — ห้ามนับเป็นสำเร็จ
      if (result.success === false) {
        throw new Error(
          (typeof result.message === "string" && result.message.trim()) ||
            backendText(dictionary, "ops_msg_action_failed_please_try_again", "ดำเนินการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"),
        );
      }
      const count = result.adjusted_count ?? selectedDocs.size;
      setNotice({
        type: "success",
        message: backendText(
          dictionary,
          "ops_msg_adjust_success",
          `ปรับปรุงสถานะใบรับสินค้าสำเร็จแล้ว ${count} รายการ`
        ),
      });

      // Clear table and form according to Champ POFrmAdjGRStatus.cpp
      setItems([]);
      setSelectedDocs(new Set());
      setFromDocNo("");
      setToDocNo("");
      setFromVendor("");
      setToVendor("");
    } catch (err: unknown) {
      setNotice({
        type: "error",
        message: err instanceof Error ? err.message : "ปรับปรุงสถานะไม่สำเร็จ",
      });
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <div className="flex h-full w-full flex-col gap-4 overflow-y-auto p-4 md:p-6">
      {/* Header */}
      <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-foreground sm:text-2xl">
              {backendText(dictionary, "purchase_receipt_adj_title", "ปรับปรุงใบรับสินค้า (GR Status Adjustment)")}
            </h1>
            <Badge variant="outline" className="border-primary/30 text-primary">
              Champ Parity
            </Badge>
          </div>
          <p className="text-sm text-muted-foreground">
            {backendText(
              dictionary,
              "purchase_receipt_adj_desc",
              "สืบค้นและปรับปรุงสถานะใบรับสินค้า (GR) ที่มีการตั้งหนี้หรือรับบิลครบแล้วให้เป็นสถานะปิดบิล"
            )}
          </p>
        </div>
      </div>

      {/* Filter Card */}
      <div className="rounded-2xl border border-border bg-card p-4 shadow-sm md:p-5">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {/* Dates */}
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

          {/* Doc Nos */}
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <FileText className="h-3.5 w-3.5" />
              {backendText(dictionary, "purchase_receipt_adj_from_doc", "จากเลขที่เอกสาร")}
            </label>
            <Input
              placeholder="GR..."
              value={fromDocNo}
              onChange={(e) => setFromDocNo(e.target.value)}
              className="h-10 text-sm"
            />
          </div>
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <FileText className="h-3.5 w-3.5" />
              {backendText(dictionary, "purchase_receipt_adj_to_doc", "ถึงเลขที่เอกสาร")}
            </label>
            <Input
              placeholder="GR..."
              value={toDocNo}
              onChange={(e) => setToDocNo(e.target.value)}
              className="h-10 text-sm"
            />
          </div>

          {/* Vendors */}
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <UserCheck className="h-3.5 w-3.5" />
              {backendText(dictionary, "purchase_receipt_adj_from_vendor", "จากรหัสผู้จำหน่าย")}
            </label>
            <Input
              placeholder="AP..."
              value={fromVendor}
              onChange={(e) => setFromVendor(e.target.value)}
              className="h-10 text-sm"
            />
          </div>
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <UserCheck className="h-3.5 w-3.5" />
              {backendText(dictionary, "purchase_receipt_adj_to_vendor", "ถึงรหัสผู้จำหน่าย")}
            </label>
            <Input
              placeholder="AP..."
              value={toVendor}
              onChange={(e) => setToVendor(e.target.value)}
              className="h-10 text-sm"
            />
          </div>

          {/* Departments */}
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <Building2 className="h-3.5 w-3.5" />
              {backendText(dictionary, "purchase_receipt_adj_from_dept", "จากแผนก")}
            </label>
            <Input
              placeholder="001"
              value={fromDept}
              onChange={(e) => setFromDept(e.target.value)}
              className="h-10 text-sm"
            />
          </div>
          <div>
            <label className="mb-1.5 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
              <Building2 className="h-3.5 w-3.5" />
              {backendText(dictionary, "purchase_receipt_adj_to_dept", "ถึงแผนก")}
            </label>
            <Input
              placeholder="001"
              value={toDept}
              onChange={(e) => setToDept(e.target.value)}
              className="h-10 text-sm"
            />
          </div>
        </div>

        {/* Action Controls & Radio */}
        <div className="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-border/60 pt-4">
          <div className="flex items-center gap-4 text-sm font-medium">
            <label className="flex cursor-pointer items-center gap-2">
              <input
                type="radio"
                name="select_mode"
                className="h-4 w-4 accent-primary"
                onChange={handleSelectAll}
                checked={items.length > 0 && selectedDocs.size === items.length}
              />
              <span>{backendText(dictionary, "purchase_receipt_adj_select_all", "เลือกทั้งหมด")}</span>
            </label>
            <label className="flex cursor-pointer items-center gap-2">
              <input
                type="radio"
                name="select_mode"
                className="h-4 w-4 accent-primary"
                onChange={handleDeselectAll}
                checked={selectedDocs.size === 0}
              />
              <span>{backendText(dictionary, "purchase_receipt_adj_deselect_all", "ไม่เลือกทั้งหมด")}</span>
            </label>
          </div>

          <div className="flex items-center gap-2.5">
            <Button
              type="button"
              variant="outline"
              onClick={handleProcess}
              disabled={isLoading}
              className="h-10 min-w-[120px] gap-2 border-border shadow-xs hover:border-primary/50"
            >
              {isLoading ? (
                <RefreshCw className="h-4 w-4 animate-spin text-primary" />
              ) : (
                <Search className="h-4 w-4 text-primary" />
              )}
              {backendText(dictionary, "purchase_receipt_adj_btn_process", "ประมวลผล")}
            </Button>

            <Button
              type="button"
              variant="default"
              onClick={() => setShowConfirm(true)}
              disabled={isSaving || selectedDocs.size === 0}
              className="h-10 min-w-[120px] gap-2 shadow-xs"
            >
              {isSaving ? (
                <RefreshCw className="h-4 w-4 animate-spin" />
              ) : (
                <FileCheck2 className="h-4 w-4" />
              )}
              {backendText(dictionary, "purchase_receipt_adj_btn_adjust", "ปรับปรุง")}
              {selectedDocs.size > 0 && (
                <span className="rounded-full bg-primary-foreground/20 px-1.5 py-0.5 text-xs font-bold">
                  {selectedDocs.size}
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
                  <span className="sr-only">Select</span>
                </th>
                <th className="px-3 py-3">{backendText(dictionary, "ops_col_docno", "เลขที่เอกสาร")}</th>
                <th className="px-3 py-3">{backendText(dictionary, "ops_col_docdate", "วันที่เอกสาร")}</th>
                <th className="px-3 py-3">{backendText(dictionary, "purchase_receipt_adj_from_vendor", "รหัสผู้จำหน่าย")}</th>
                <th className="px-3 py-3">{backendText(dictionary, "label_vendor_name", "ชื่อผู้จำหน่าย")}</th>
                <th className="px-3 py-3">{backendText(dictionary, "purchase_receipt_adj_ir_ref", "เลขที่อ้างอิง IR")}</th>
                <th className="px-3 py-3">{backendText(dictionary, "purchase_receipt_adj_ir_date", "วันที่อ้างอิง IR")}</th>
                <th className="px-3 py-3 text-right">{backendText(dictionary, "purchase_receipt_adj_gr_qty", "จำนวนรับ (GR)")}</th>
                <th className="px-3 py-3 text-right">{backendText(dictionary, "purchase_receipt_adj_ir_qty", "จำนวนตั้งหนี้ (IR)")}</th>
                <th className="px-3 py-3 text-center">{backendText(dictionary, "purchase_receipt_adj_status", "สถานะการเทียบยอด")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {items.length === 0 ? (
                <tr>
                  <td colSpan={10} className="py-12 text-center text-muted-foreground">
                    {isLoading
                      ? "กำลังโหลดข้อมูล..."
                      : backendText(dictionary, "ops_msg_no_items_click_process", "กดปุ่ม [ประมวลผล] เพื่อแสดงรายการใบรับสินค้า")}
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
                      <td className="px-3 py-3 font-medium text-foreground">{item.docno}</td>
                      <td className="px-3 py-3 text-muted-foreground">{item.docdate}</td>
                      <td className="px-3 py-3 text-muted-foreground">{item.vendor_code}</td>
                      <td className="max-w-[200px] truncate px-3 py-3 text-foreground" title={item.vendor_name}>
                        {item.vendor_name}
                      </td>
                      <td className="px-3 py-3 text-muted-foreground">
                        {item.ir_ref_no || "-"}
                      </td>
                      <td className="px-3 py-3 text-muted-foreground">
                        {item.ir_ref_date || "-"}
                      </td>
                      <td className="px-3 py-3 text-right font-mono text-foreground">{item.gr_qty}</td>
                      <td className="px-3 py-3 text-right font-mono text-foreground">{item.ir_qty}</td>
                      <td className="px-3 py-3 text-center">
                        {item.is_matched ? (
                          <Badge variant="outline" className="border-emerald-500/30 bg-emerald-50 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-400">
                            ตั้งหนี้ครบแล้ว
                          </Badge>
                        ) : (
                          <Badge variant="outline" className="border-amber-500/30 bg-amber-50 text-amber-700 dark:bg-amber-950/30 dark:text-amber-400">
                            ยังตั้งหนี้ไม่ครบ
                          </Badge>
                        )}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Confirmation Modal */}
      {showConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-xs">
          <div className="w-full max-w-md rounded-2xl border border-border bg-card p-6 shadow-xl">
            <h3 className="text-lg font-bold text-foreground">
              {backendText(dictionary, "purchase_receipt_adj_confirm_title", "ยืนยันการปรับปรุงใบรับสินค้า")}
            </h3>
            <p className="mt-2 text-sm text-muted-foreground">
              {backendText(
                dictionary,
                "purchase_receipt_adj_confirm_msg",
                `คุณต้องการปรับปรุงใบรับสินค้า (GR) ที่เลือกจำนวน ${selectedDocs.size} รายการใช่หรือไม่? เอกสารจะถูกบันทึกเป็นปิดบิล`
              )}
            </p>
            <div className="mt-6 flex justify-end gap-3">
              <Button
                type="button"
                variant="outline"
                onClick={() => setShowConfirm(false)}
                className="h-10"
              >
                {backendText(dictionary, "btn_cancel", "ยกเลิก")}
              </Button>
              <Button
                type="button"
                variant="default"
                onClick={handleAdjustSave}
                disabled={isSaving}
                className="h-10"
              >
                {isSaving && <RefreshCw className="mr-2 h-4 w-4 animate-spin" />}
                {backendText(dictionary, "btn_confirm", "ยืนยันการปรับปรุง")}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
