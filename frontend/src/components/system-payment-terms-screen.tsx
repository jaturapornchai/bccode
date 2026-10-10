"use client";

import { useMemo, useRef, useState } from "react";
import {
  Calendar,
  CheckCircle2,
  CreditCard,
  Percent,
  Plus,
  RotateCcw,
  Save,
  Search,
  Sliders,
  Wallet,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export interface PaymentTermItem {
  id: string;
  code: string;
  nameTh: string;
  nameEn: string;
  creditDays: number;
  dueMethod: "days_from_doc" | "end_of_month" | "cutoff_day";
  discountDays: number;
  discountPercent: string;
  isDefaultCustomer: boolean;
  isDefaultSupplier: boolean;
  isActive: boolean;
}

const DEFAULT_TERMS: PaymentTermItem[] = [
  {
    id: "term-cash",
    code: "CASH",
    nameTh: "เงินสด (Cash on Delivery)",
    nameEn: "Cash / Immediate Payment",
    creditDays: 0,
    dueMethod: "days_from_doc",
    discountDays: 0,
    discountPercent: "0.00",
    isDefaultCustomer: false,
    isDefaultSupplier: false,
    isActive: true,
  },
  {
    id: "term-net07",
    code: "NET07",
    nameTh: "เครดิต 7 วัน",
    nameEn: "Net 7 Days",
    creditDays: 7,
    dueMethod: "days_from_doc",
    discountDays: 0,
    discountPercent: "0.00",
    isDefaultCustomer: false,
    isDefaultSupplier: false,
    isActive: true,
  },
  {
    id: "term-net15",
    code: "NET15",
    nameTh: "เครดิต 15 วัน",
    nameEn: "Net 15 Days",
    creditDays: 15,
    dueMethod: "days_from_doc",
    discountDays: 0,
    discountPercent: "0.00",
    isDefaultCustomer: false,
    isDefaultSupplier: false,
    isActive: true,
  },
  {
    id: "term-net30",
    code: "NET30",
    nameTh: "เครดิต 30 วัน (มาตรฐานการค้า)",
    nameEn: "Net 30 Days",
    creditDays: 30,
    dueMethod: "days_from_doc",
    discountDays: 7,
    discountPercent: "2.00",
    isDefaultCustomer: true,
    isDefaultSupplier: true,
    isActive: true,
  },
  {
    id: "term-net60",
    code: "NET60",
    nameTh: "เครดิต 60 วัน",
    nameEn: "Net 60 Days",
    creditDays: 60,
    dueMethod: "days_from_doc",
    discountDays: 0,
    discountPercent: "0.00",
    isDefaultCustomer: false,
    isDefaultSupplier: false,
    isActive: true,
  },
  {
    id: "term-eom",
    code: "EOM",
    nameTh: "สิ้นเดือนของรอบบิล (End of Month)",
    nameEn: "End of Month",
    creditDays: 0,
    dueMethod: "end_of_month",
    discountDays: 0,
    discountPercent: "0.00",
    isDefaultCustomer: false,
    isDefaultSupplier: false,
    isActive: true,
  },
];

export interface SystemPaymentTermsScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemPaymentTermsScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: SystemPaymentTermsScreenProps) {
  const isTh = language === "th";
  const [terms, setTerms] = useState<PaymentTermItem[]>(DEFAULT_TERMS);
  const [selectedId, setSelectedId] = useState<string>(DEFAULT_TERMS[3].id); // Default to NET30
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 38,
    min: 25,
    max: 55,
    storageKey: "bcai_split_payment_terms",
  });

  const selectedTerm = useMemo(() => {
    return terms.find((t) => t.id === selectedId) || terms[0];
  }, [terms, selectedId]);

  const handleUpdate = <K extends keyof PaymentTermItem>(
    field: K,
    val: PaymentTermItem[K]
  ) => {
    setTerms((prev) =>
      prev.map((t) => (t.id === selectedTerm.id ? { ...t, [field]: val } : t))
    );
  };

  const handleAddNewTerm = () => {
    const newTerm: PaymentTermItem = {
      id: `term-${Date.now()}`,
      code: "NEWTERM",
      nameTh: "เงื่อนไขชำระเงินใหม่",
      nameEn: "New Payment Term",
      creditDays: 30,
      dueMethod: "days_from_doc",
      discountDays: 0,
      discountPercent: "0.00",
      isDefaultCustomer: false,
      isDefaultSupplier: false,
      isActive: true,
    };
    setTerms((prev) => [...prev, newTerm]);
    setSelectedId(newTerm.id);
  };

  const handleSave = () => {
    setNotification(
      isTh
        ? `บันทึกเงื่อนไขการชำระเงิน (${selectedTerm.code} - ${selectedTerm.nameTh}) เรียบร้อยแล้ว`
        : `Payment term (${selectedTerm.code}) saved successfully.`
    );
    setTimeout(() => setNotification(null), 3500);
  };

  const handleReset = () => {
    setTerms(DEFAULT_TERMS);
    setNotification(isTh ? "คืนค่าการตั้งค่าเริ่มต้นเรียบร้อย" : "Reset terms to defaults");
    setTimeout(() => setNotification(null), 3000);
  };

  return (
    <div className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 font-sans text-xs">
      {/* Header Bar */}
      <div className="flex-none p-3.5 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-wrap items-center justify-between gap-3 shadow-xs">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-indigo-100 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
            <CreditCard className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-semibold text-slate-900 dark:text-slate-100">
                {isTh ? "เงื่อนไขการชำระเงิน" : "Payment Terms Configuration"}
              </h1>
              <span className="px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-[11px] text-slate-500 dark:text-slate-400">
              {isTh
                ? "จัดการระบบ › กำหนดจำนวนวันเครดิตเทอม วิธีคำนวณวันครบกำหนด และส่วนลดจ่ายเร็ว (Cash Discount)"
                : "System Settings › Setup credit term periods, due date calculation rules, and early payment discounts"}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={handleReset}
            className="h-8 gap-1.5 text-xs text-slate-700 dark:text-slate-300"
          >
            <RotateCcw className="w-3.5 h-3.5" />
            {isTh ? "คืนค่าเริ่มต้น" : "Reset Default"}
          </Button>
          <Button
            size="sm"
            onClick={handleSave}
            className="h-8 gap-1.5 text-xs bg-indigo-600 hover:bg-indigo-700 text-white"
          >
            <Save className="w-3.5 h-3.5" />
            {isTh ? "บันทึกการตั้งค่า" : "Save Settings"}
          </Button>
        </div>
      </div>

      {/* Notification Banner */}
      {notification && (
        <div className="flex-none bg-emerald-50 dark:bg-emerald-950/50 border-b border-emerald-200 dark:border-emerald-800 px-4 py-2 flex items-center gap-2 text-emerald-800 dark:text-emerald-300 text-xs animate-in fade-in duration-200">
          <CheckCircle2 className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
          <span>{notification}</span>
        </div>
      )}

      {/* Main Split Content */}
      <div ref={containerRef} className="flex-1 flex overflow-hidden relative">
        {/* Left Column: Terms List */}
        <div
          style={{ width: `${splitPercent}%` }}
          className="h-full flex flex-col border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900"
        >
          <div className="p-3 border-b border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/50 flex items-center justify-between">
            <span className="font-semibold text-slate-700 dark:text-slate-200">
              {isTh ? "เงื่อนไขทั้งหมด" : "Payment Terms"} ({terms.length})
            </span>
            <Button
              size="sm"
              variant="ghost"
              onClick={handleAddNewTerm}
              className="h-7 text-xs gap-1 text-indigo-600 hover:text-indigo-700 hover:bg-indigo-50"
            >
              <Plus className="w-3.5 h-3.5" />
              {isTh ? "เพิ่มเงื่อนไข" : "Add"}
            </Button>
          </div>

          <div className="flex-1 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800">
            {terms.map((term) => {
              const isSelected = term.id === selectedTerm.id;
              return (
                <div
                  key={term.id}
                  onClick={() => setSelectedId(term.id)}
                  className={`p-3.5 cursor-pointer transition-colors ${
                    isSelected
                      ? "bg-indigo-50/80 dark:bg-indigo-950/40 border-l-4 border-indigo-600"
                      : "hover:bg-slate-50 dark:hover:bg-slate-800/40"
                  }`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-mono font-bold text-indigo-600 dark:text-indigo-400 text-sm">
                      {term.code}
                    </span>
                    <div className="flex items-center gap-1">
                      {term.isDefaultCustomer && (
                        <span className="px-1.5 py-0.5 rounded text-[10px] bg-blue-100 text-blue-700 dark:bg-blue-950/60 dark:text-blue-400">
                          AR
                        </span>
                      )}
                      {term.isDefaultSupplier && (
                        <span className="px-1.5 py-0.5 rounded text-[10px] bg-amber-100 text-amber-700 dark:bg-amber-950/60 dark:text-amber-400">
                          AP
                        </span>
                      )}
                      <span className="font-mono font-bold bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded text-[10px]">
                        {term.creditDays} {isTh ? "วัน" : "days"}
                      </span>
                    </div>
                  </div>
                  <p className="text-slate-800 dark:text-slate-200 font-medium">{term.nameTh}</p>
                  {Number(term.discountPercent) > 0 && (
                    <p className="mt-1 text-[10px] text-emerald-600 dark:text-emerald-400 flex items-center gap-1">
                      <Percent className="w-3 h-3" />
                      {isTh
                        ? `ส่วนลด ${term.discountPercent}% เมื่อชำระใน ${term.discountDays} วัน`
                        : `${term.discountPercent}% discount if paid within ${term.discountDays} days`}
                    </p>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        {/* Resizable Splitter Handle */}
        <ResizableSplitter
          onMouseDown={startResize}
          onKeyDown={adjustWithKeyboard}
          isResizing={isResizing}
        />

        {/* Right Column: Term Edit Form */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="h-full flex flex-col bg-slate-50/50 dark:bg-slate-950 overflow-y-auto p-4 gap-4"
        >
          {/* Main Form Card */}
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5 shadow-xs">
            <div className="flex items-center gap-2 pb-3 mb-4 border-b border-slate-100 dark:border-slate-800">
              <Sliders className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
              <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
                {isTh ? "รายละเอียดเงื่อนไขการชำระเงิน" : "Payment Term Specifications"}
              </h2>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "รหัสเงื่อนไข (Term Code)" : "Term Code"}
                </label>
                <Input
                  value={selectedTerm.code}
                  onChange={(e) => handleUpdate("code", e.target.value.toUpperCase())}
                  className="font-mono uppercase font-bold text-indigo-600 dark:text-indigo-400"
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "จำนวนวันเครดิต (Credit Days)" : "Credit Days"}
                </label>
                <Input
                  type="number"
                  min={0}
                  value={selectedTerm.creditDays}
                  onChange={(e) => handleUpdate("creditDays", Math.max(0, parseInt(e.target.value) || 0))}
                  className="font-mono font-bold text-slate-800 dark:text-slate-100"
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ชื่อเงื่อนไขภาษาไทย" : "Thai Name"}
                </label>
                <Input
                  value={selectedTerm.nameTh}
                  onChange={(e) => handleUpdate("nameTh", e.target.value)}
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ชื่อเงื่อนไขภาษาอังกฤษ" : "English Name"}
                </label>
                <Input
                  value={selectedTerm.nameEn}
                  onChange={(e) => handleUpdate("nameEn", e.target.value)}
                />
              </div>

              {/* Due Date Method */}
              <div className="sm:col-span-2">
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "วิธีคำนวณวันครบกำหนด (Due Date Calculation)" : "Due Date Calculation Rule"}
                </label>
                <select
                  value={selectedTerm.dueMethod}
                  onChange={(e) => handleUpdate("dueMethod", e.target.value as PaymentTermItem["dueMethod"])}
                  className="w-full h-8 px-2.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs"
                >
                  <option value="days_from_doc">{isTh ? "นับจำนวนวันนับจากวันที่ในเอกสาร (Days from Document Date)" : "Days from Document Date"}</option>
                  <option value="end_of_month">{isTh ? "สิ้นเดือนของวันที่ในเอกสาร (End of Month - EOM)" : "End of Month"}</option>
                  <option value="cutoff_day">{isTh ? "วันตัดรอบสิ้นงวดบิลถัดไป (Next Month Cutoff Day)" : "Next Month Cutoff Day"}</option>
                </select>
              </div>

              {/* Cash Discount Settings */}
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ส่วนลดจ่ายเร็ว Cash Discount (%)" : "Early Payment Discount (%)"}
                </label>
                <Input
                  type="number"
                  step="0.01"
                  min={0}
                  value={selectedTerm.discountPercent}
                  onChange={(e) => handleUpdate("discountPercent", e.target.value)}
                  className="font-mono text-emerald-600 dark:text-emerald-400 font-bold"
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ภายในจำนวนวัน (วัน)" : "Discount Days"}
                </label>
                <Input
                  type="number"
                  min={0}
                  value={selectedTerm.discountDays}
                  onChange={(e) => handleUpdate("discountDays", Math.max(0, parseInt(e.target.value) || 0))}
                  className="font-mono"
                />
              </div>
            </div>

            {/* Checkbox Defaults */}
            <div className="mt-5 pt-4 border-t border-slate-100 dark:border-slate-800 space-y-2.5">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedTerm.isDefaultCustomer}
                  onChange={(e) => handleUpdate("isDefaultCustomer", e.target.checked)}
                  className="rounded text-indigo-600 focus:ring-indigo-500 w-4 h-4"
                />
                <span className="text-slate-700 dark:text-slate-300">
                  {isTh
                    ? "ใช้เป็นเงื่อนไขเริ่มต้นสำหรับลูกหนี้การค้าใหม่ (Default for Customers / AR)"
                    : "Set as default payment term for new customer accounts"}
                </span>
              </label>

              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedTerm.isDefaultSupplier}
                  onChange={(e) => handleUpdate("isDefaultSupplier", e.target.checked)}
                  className="rounded text-indigo-600 focus:ring-indigo-500 w-4 h-4"
                />
                <span className="text-slate-700 dark:text-slate-300">
                  {isTh
                    ? "ใช้เป็นเงื่อนไขเริ่มต้นสำหรับเจ้าหนี้การค้าใหม่ (Default for Suppliers / AP)"
                    : "Set as default payment term for new supplier accounts"}
                </span>
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
