"use client";

import React, { useState, useEffect, useCallback, useMemo } from "react";
import {
  Search,
  Calendar,
  CalendarDays,
  Copy,
  CheckCircle2,
  AlertCircle,
  Undo2,
  Save,
  Building2,
  FileSpreadsheet,
  RefreshCw,
} from "lucide-react";
import { useBackendText, useBackendDictionary } from "@/components/backend-text-provider";
import { backendText } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";
import {
  fetchDeliveryDateInvoices,
  saveDeliveryDates,
  type DeliveryDateItem,
} from "@/lib/erp-transaction";

export interface DeliveryDateOverhaulProps {
  mode: "customer" | "vendor";
  language?: LanguageCode;
  holdingcode?: string;
  onClose?: () => void;
}

function addDaysToDate(dateStr: string, days: number): string {
  if (!dateStr) return "";
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) return dateStr;
  d.setDate(d.getDate() + Number(days || 0));
  return d.toISOString().split("T")[0];
}

export function DeliveryDateOverhaulScreen({
  mode,
  language = "th",
  onClose,
}: DeliveryDateOverhaulProps) {
  const tr = useBackendText();
  const dictionary = useBackendDictionary();

  const isVendor = mode === "vendor";
  const moduleName = isVendor ? "purchase" : "sale-invoice";

  const [partnerCode, setPartnerCode] = useState<string>("");
  const [partnerName, setPartnerName] = useState<string>("");
  const [items, setItems] = useState<DeliveryDateItem[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [submitting, setSubmitting] = useState<boolean>(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);
  const [hasChanges, setHasChanges] = useState<boolean>(false);

  // Load invoices
  const handleLoadInvoices = useCallback(async () => {
    setLoading(true);
    setErrorMsg(null);
    setSuccessMsg(null);
    try {
      const res = await fetchDeliveryDateInvoices(moduleName, partnerCode.trim(), language);
      if (res.success) {
        setItems(res.items);
        if (res.items.length > 0 && res.items[0].custname && !partnerName) {
          setPartnerName(res.items[0].custname);
        }
        if (res.items.length === 0) {
          setErrorMsg(backendText(dictionary, "ops_no_invoices_found", "ไม่พบรายการบิลค้างชำระ"));
        }
        setHasChanges(false);
      } else {
        setErrorMsg(backendText(dictionary, res.error || "load_failed", "โหลดรายการเอกสารไม่สำเร็จ"));
      }
    } catch {
      setErrorMsg(backendText(dictionary, "connection_error", "เกิดข้อผิดพลาดในการเชื่อมต่อ"));
    } finally {
      setLoading(false);
    }
  }, [moduleName, partnerCode, partnerName, language, dictionary]);

  // Recalculate credit days for all rows (ปุ่ม: ปรับปรุงเครดิตนับจากวันส่งของ)
  const handleOverhaulAllCreditDays = () => {
    if (items.length === 0) return;
    const updated = items.map((item) => {
      const newDue = addDaysToDate(item.delivery_date, item.credit_delivery);
      return {
        ...item,
        new_duedate: newDue,
      };
    });
    setItems(updated);
    setHasChanges(true);
    setSuccessMsg(backendText(dictionary, "ops_applied_all_credit_days", "คำนวณวันครบกำหนดใหม่ทุกรายการแล้ว"));
  };

  // Copy first row's due date to all rows (ปุ่ม: ปรับปรุงวันที่ครบกำหนด เหมือนบรรทัดแรกทุกรายการ)
  const handleApplyFirstRowDueDate = () => {
    if (items.length === 0) return;
    const firstRowDueDate = items[0].new_duedate || items[0].old_duedate;
    if (!firstRowDueDate) return;

    const updated = items.map((item) => ({
      ...item,
      new_duedate: firstRowDueDate,
    }));
    setItems(updated);
    setHasChanges(true);
    setSuccessMsg(backendText(dictionary, "ops_applied_first_row_due_date", "กำหนดวันครบกำหนดเหมือนบรรทัดแรกทุกรายการแล้ว"));
  };

  // Row field update
  const handleRowChange = (index: number, field: "delivery_date" | "credit_delivery" | "new_duedate", value: string | number) => {
    setItems((prev) => {
      const copy = [...prev];
      const target = { ...copy[index] };

      if (field === "delivery_date") {
        target.delivery_date = String(value);
        target.new_duedate = addDaysToDate(target.delivery_date, target.credit_delivery);
      } else if (field === "credit_delivery") {
        target.credit_delivery = Number(value) || 0;
        target.new_duedate = addDaysToDate(target.delivery_date, target.credit_delivery);
      } else if (field === "new_duedate") {
        target.new_duedate = String(value);
      }

      copy[index] = target;
      return copy;
    });
    setHasChanges(true);
  };

  // Save changes (ปุ่ม F11 = ตกลง)
  const handleSave = async () => {
    if (items.length === 0) return;
    setSubmitting(true);
    setErrorMsg(null);
    setSuccessMsg(null);

    const payloadItems = items.map((item) => ({
      docno: item.docno,
      delivery_date: item.delivery_date,
      credit_delivery: item.credit_delivery,
      new_duedate: item.new_duedate,
    }));

    try {
      const res = await saveDeliveryDates(moduleName, payloadItems, language);
      if (res.success) {
        setSuccessMsg(
          backendText(dictionary, "ops_save_delivery_dates_success", `บันทึกการปรับปรุงวันครบกำหนดสำเร็จ (${res.count ?? items.length} รายการ)`)
        );
        setHasChanges(false);
      } else {
        setErrorMsg(backendText(dictionary, res.message || "action_failed", "บันทึกการปรับปรุงวันที่ไม่สำเร็จ"));
      }
    } catch {
      setErrorMsg(backendText(dictionary, "connection_error", "เกิดข้อผิดพลาดในการเชื่อมต่อ"));
    } finally {
      setSubmitting(false);
    }
  };

  // Cancel action (ปุ่ม Esc = ยกเลิก)
  const handleCancel = () => {
    if (onClose) {
      onClose();
    } else {
      setItems([]);
      setPartnerCode("");
      setPartnerName("");
      setHasChanges(false);
      setErrorMsg(null);
      setSuccessMsg(null);
    }
  };

  // Hotkeys: F11 for OK/Save, Esc for Cancel
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "F11") {
        e.preventDefault();
        void handleSave();
      } else if (e.key === "Escape") {
        e.preventDefault();
        handleCancel();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  });

  const screenTitle = isVendor
    ? backendText(dictionary, "menu_purchase_receipt_date_adjust", "ปรับปรุงวันที่รับสินค้าจากเจ้าหนี้")
    : backendText(dictionary, "menu_sale_delivery_date", "ปรับปรุงวันที่ส่งของให้ลูกค้า");

  const partnerLabel = isVendor
    ? backendText(dictionary, "field_vendor_code", "รหัสเจ้าหนี้")
    : backendText(dictionary, "field_customer_code", "รหัสลูกค้า");

  const deliveryColLabel = isVendor
    ? backendText(dictionary, "col_billing_date", "วันที่วางบิล")
    : backendText(dictionary, "col_delivery_date", "วันที่ส่งของ");

  const creditColLabel = isVendor
    ? backendText(dictionary, "col_credit_days_simple", "เครดิต(วัน)")
    : backendText(dictionary, "col_credit_from_delivery_date", "เครดิตนับจากวันส่งของ");

  return (
    <div className="flex flex-col h-full bg-slate-50 dark:bg-slate-950 p-4 sm:p-6 overflow-y-auto">
      {/* Header card matching Champ layout */}
      <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl shadow-sm p-4 sm:p-5 mb-4">
        {/* Title bar */}
        <div className="flex items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800 mb-4">
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-lg bg-primary/10 text-primary">
              <CalendarDays className="w-5 h-5" />
            </div>
            <div>
              <h1 className="text-lg font-bold text-slate-900 dark:text-slate-100">
                {screenTitle}
              </h1>
              <p className="text-xs text-slate-500 dark:text-slate-400">
                {isVendor
                  ? backendText(dictionary, "ops_desc_vendor_receipt_date", "ปรับปรุงวันที่รับสินค้าหรือวันที่วางบิล และคำนวณวันครบกำหนดชำระเงินใหม่")
                  : backendText(dictionary, "ops_desc_customer_delivery_date", "ปรับปรุงวันที่ส่งของจริงให้ลูกค้า และคำนวณวันครบกำหนดชำระหนี้ใหม่")}
              </p>
            </div>
          </div>

          {/* Action buttons (F11 / Esc) */}
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleSave}
              disabled={submitting || items.length === 0}
              className="inline-flex items-center gap-1.5 px-4 py-2 text-sm font-semibold rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-emerald-500/50"
              title="กดแป้น F11 เพื่อตกลงบันทึก"
            >
              {submitting ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                <Save className="w-4 h-4" />
              )}
              {backendText(dictionary, "ops_btn_f11_confirm", "F11=ตกลง")}
            </button>

            <button
              type="button"
              onClick={handleCancel}
              className="inline-flex items-center gap-1.5 px-3.5 py-2 text-sm font-medium rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-all focus:outline-none focus:ring-2 focus:ring-slate-400/50"
              title="กดแป้น Esc เพื่อยกเลิก"
            >
              <Undo2 className="w-4 h-4" />
              {backendText(dictionary, "ops_btn_esc_cancel", "Esc=ยกเลิก")}
            </button>
          </div>
        </div>

        {/* Input filters row */}
        <div className="grid grid-cols-1 md:grid-cols-12 gap-3 items-center mb-4">
          <label className="md:col-span-2 text-sm font-medium text-slate-700 dark:text-slate-300">
            {partnerLabel}
          </label>
          <div className="md:col-span-4 flex items-center gap-1.5">
            <input
              type="text"
              value={partnerCode}
              onChange={(e) => setPartnerCode(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  void handleLoadInvoices();
                }
              }}
              placeholder={isVendor ? "เช่น V001 หรือเว้นว่าง" : "เช่น C001 หรือเว้นว่าง"}
              className="w-full px-3 py-1.5 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary font-mono"
            />
            <button
              type="button"
              onClick={() => void handleLoadInvoices()}
              className="p-2 rounded-lg border border-slate-300 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 transition-colors"
              title="ค้นหา"
            >
              <Search className="w-4 h-4" />
            </button>
          </div>
          <div className="md:col-span-6">
            <input
              type="text"
              readOnly
              value={partnerName}
              placeholder={backendText(dictionary, "field_partner_name_placeholder", "ชื่อกิจการ / คู่ค้า")}
              className="w-full px-3 py-1.5 text-sm bg-slate-100 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700/60 rounded-lg text-slate-600 dark:text-slate-400 cursor-default"
            />
          </div>
        </div>

        {/* Toolbar Buttons matching Champ */}
        <div className="flex flex-wrap items-center gap-2 pt-2 border-t border-slate-100 dark:border-slate-800">
          <button
            type="button"
            onClick={() => void handleLoadInvoices()}
            disabled={loading}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg bg-blue-50 dark:bg-blue-950/40 text-blue-700 dark:text-blue-300 border border-blue-200 dark:border-blue-900/60 hover:bg-blue-100 dark:hover:bg-blue-900/40 transition-colors disabled:opacity-50"
          >
            {loading ? (
              <RefreshCw className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <FileSpreadsheet className="w-3.5 h-3.5" />
            )}
            {backendText(dictionary, "ops_select_bills", "เลือกบิล")}
          </button>

          <button
            type="button"
            onClick={handleOverhaulAllCreditDays}
            disabled={items.length === 0}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg bg-indigo-50 dark:bg-indigo-950/40 text-indigo-700 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-900/60 hover:bg-indigo-100 dark:hover:bg-indigo-900/40 transition-colors disabled:opacity-50"
          >
            <Calendar className="w-3.5 h-3.5" />
            {backendText(dictionary, "ops_overhaul_by_delivery_date", "ปรับปรุงเครดิตนับจากวันส่งของ")}
          </button>

          <button
            type="button"
            onClick={handleApplyFirstRowDueDate}
            disabled={items.length === 0}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg bg-purple-50 dark:bg-purple-950/40 text-purple-700 dark:text-purple-300 border border-purple-200 dark:border-purple-900/60 hover:bg-purple-100 dark:hover:bg-purple-900/40 transition-colors disabled:opacity-50"
          >
            <Copy className="w-3.5 h-3.5" />
            {backendText(dictionary, "ops_apply_first_row_due_date", "ปรับปรุงวันที่ครบกำหนด เหมือนบรรทัดแรกทุกรายการ")}
          </button>
        </div>
      </div>

      {/* Notifications */}
      {errorMsg && (
        <div className="flex items-center gap-2 p-3 mb-3 text-sm text-red-700 dark:text-red-400 bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900/50 rounded-lg">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{errorMsg}</span>
        </div>
      )}

      {successMsg && (
        <div className="flex items-center gap-2 p-3 mb-3 text-sm text-emerald-700 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-900/50 rounded-lg">
          <CheckCircle2 className="w-4 h-4 shrink-0" />
          <span>{successMsg}</span>
        </div>
      )}

      {/* Data Table */}
      <div className="flex-1 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl shadow-sm overflow-hidden flex flex-col">
        <div className="overflow-x-auto flex-1">
          <table className="w-full text-left border-collapse text-xs sm:text-sm">
            <thead>
              <tr className="bg-slate-100 dark:bg-slate-800/80 text-slate-700 dark:text-slate-300 font-semibold border-b border-slate-200 dark:border-slate-700">
                <th className="py-2.5 px-3 w-12 text-center">#</th>
                <th className="py-2.5 px-3 font-semibold">{backendText(dictionary, "col_docno", "เลขที่เอกสาร")}</th>
                <th className="py-2.5 px-3 text-center">{backendText(dictionary, "col_docdate", "วันที่เอกสาร")}</th>
                <th className="py-2.5 px-3 text-right">{backendText(dictionary, "col_credit_of_bill", "เครดิตตามออกบิล")}</th>
                <th className="py-2.5 px-3 text-center">{backendText(dictionary, "col_old_duedate", "วันครบกำหนดเดิม")}</th>
                <th className="py-2.5 px-3 text-center bg-blue-50/60 dark:bg-blue-950/30 font-semibold text-blue-900 dark:text-blue-200">
                  {deliveryColLabel} *
                </th>
                <th className="py-2.5 px-3 text-right bg-blue-50/60 dark:bg-blue-950/30 font-semibold text-blue-900 dark:text-blue-200">
                  {creditColLabel} *
                </th>
                <th className="py-2.5 px-3 text-center bg-indigo-50/60 dark:bg-indigo-950/30 font-semibold text-indigo-900 dark:text-indigo-200">
                  {backendText(dictionary, "col_new_duedate", "วันครบกำหนดใหม่")} *
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
              {items.length === 0 ? (
                <tr>
                  <td colSpan={8} className="py-12 text-center text-slate-400 dark:text-slate-500">
                    <FileSpreadsheet className="w-10 h-10 mx-auto mb-2 opacity-40" />
                    <p className="text-sm font-medium">
                      {backendText(
                        dictionary,
                        "ops_empty_delivery_list",
                        "ยังไม่มีรายการบิลที่เลือก — กรุณาระบุรหัสคู่ค้าแล้วกดปุ่ม [เลือกบิล]"
                      )}
                    </p>
                  </td>
                </tr>
              ) : (
                items.map((item, idx) => (
                  <tr
                    key={item.docno}
                    className="hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
                  >
                    <td className="py-2 px-3 text-center text-slate-400 font-mono text-xs">
                      {idx + 1}
                    </td>
                    <td className="py-2 px-3 font-mono font-medium text-slate-900 dark:text-slate-100">
                      {item.docno}
                      {item.custname && (
                        <span className="block text-xs font-normal text-slate-500 dark:text-slate-400 truncate max-w-xs">
                          {item.custname}
                        </span>
                      )}
                    </td>
                    <td className="py-2 px-3 text-center text-slate-600 dark:text-slate-400 font-mono text-xs">
                      {item.docdate}
                    </td>
                    <td className="py-2 px-3 text-right font-mono text-slate-600 dark:text-slate-400">
                      {item.credit_of_bill}
                    </td>
                    <td className="py-2 px-3 text-center text-slate-500 dark:text-slate-400 font-mono text-xs">
                      {item.old_duedate}
                    </td>
                    <td className="py-1.5 px-2 bg-blue-50/20 dark:bg-blue-950/10">
                      <input
                        type="date"
                        value={item.delivery_date}
                        onChange={(e) => handleRowChange(idx, "delivery_date", e.target.value)}
                        className="w-full px-2 py-1 text-xs sm:text-sm font-mono text-center bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded focus:outline-none focus:ring-1 focus:ring-primary"
                      />
                    </td>
                    <td className="py-1.5 px-2 bg-blue-50/20 dark:bg-blue-950/10">
                      <input
                        type="number"
                        min="0"
                        max="999"
                        value={item.credit_delivery}
                        onChange={(e) => handleRowChange(idx, "credit_delivery", e.target.value)}
                        className="w-20 px-2 py-1 text-xs sm:text-sm font-mono text-right bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded focus:outline-none focus:ring-1 focus:ring-primary ml-auto block"
                      />
                    </td>
                    <td className="py-1.5 px-2 bg-indigo-50/20 dark:bg-indigo-950/10">
                      <input
                        type="date"
                        value={item.new_duedate}
                        onChange={(e) => handleRowChange(idx, "new_duedate", e.target.value)}
                        className="w-full px-2 py-1 text-xs sm:text-sm font-mono text-center font-semibold text-indigo-700 dark:text-indigo-300 bg-white dark:bg-slate-800 border border-indigo-300 dark:border-indigo-700 rounded focus:outline-none focus:ring-1 focus:ring-primary"
                      />
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Footer summary */}
        {items.length > 0 && (
          <div className="py-2.5 px-4 bg-slate-50 dark:bg-slate-800/60 border-t border-slate-200 dark:border-slate-700 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
            <span>
              {backendText(dictionary, "ops_total_invoices_count", `จำนวนเอกสารทั้งหมด: ${items.length} รายการ`)}
            </span>
            {hasChanges && (
              <span className="text-amber-600 dark:text-amber-400 font-medium">
                {backendText(dictionary, "ops_unsaved_changes_hint", "● มีข้อมูลที่ยังไม่ได้บันทึก (กด F11 เพื่อตกลง)")}
              </span>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
