"use client";

import { useMemo, useRef, useState } from "react";
import {
  AlertCircle,
  CheckCircle2,
  FileSpreadsheet,
  FileText,
  Filter,
  Layers,
  Plus,
  RotateCcw,
  Save,
  Search,
  Sliders,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export type NoteType = "credit_note" | "debit_note" | "both";
export type StockImpact = "return_to_stock" | "value_only" | "deduct_stock";

export interface CreditDebitReasonItem {
  id: string;
  code: string;
  nameTh: string;
  nameEn: string;
  noteType: NoteType;
  stockImpact: StockImpact;
  etaxCode: string; // e-Tax Invoice Revenue Dept Reason Code
  isActive: boolean;
}

const DEFAULT_REASONS: CreditDebitReasonItem[] = [
  {
    id: "r-cn-01",
    code: "CN-01",
    nameTh: "สินค้าชำรุดเสียหาย แตกหัก ไม่สมบูรณ์",
    nameEn: "Damaged / Defective Goods",
    noteType: "credit_note",
    stockImpact: "return_to_stock",
    etaxCode: "CDNG01",
    isActive: true,
  },
  {
    id: "r-cn-02",
    code: "CN-02",
    nameTh: "สินค้าไม่ตรงตามคุณลักษณะหรือตัวอย่างที่ตกลง",
    nameEn: "Goods Not Matching Sample / Specification",
    noteType: "credit_note",
    stockImpact: "return_to_stock",
    etaxCode: "CDNG02",
    isActive: true,
  },
  {
    id: "r-cn-03",
    code: "CN-03",
    nameTh: "คำนวณราคาสินค้าหรือบริการสูงเกินความเป็นจริง",
    nameEn: "Price Overcharged / Pricing Error",
    noteType: "credit_note",
    stockImpact: "value_only",
    etaxCode: "CDNG03",
    isActive: true,
  },
  {
    id: "r-cn-04",
    code: "CN-04",
    nameTh: "ส่งคืนสินค้าเนื่องจากยกเลิกสัญญาหรือคำสั่งซื้อ",
    nameEn: "Order / Contract Cancellation",
    noteType: "credit_note",
    stockImpact: "return_to_stock",
    etaxCode: "CDNG04",
    isActive: true,
  },
  {
    id: "r-cn-05",
    code: "CN-05",
    nameTh: "ให้ส่วนลดการค้าภายหลังการส่งมอบสินค้า",
    nameEn: "Post-Sale Trade Discount",
    noteType: "credit_note",
    stockImpact: "value_only",
    etaxCode: "CDNG05",
    isActive: true,
  },
  {
    id: "r-dn-01",
    code: "DN-01",
    nameTh: "คำนวณราคาสินค้าหรือบริการต่ำกว่าความเป็นจริง",
    nameEn: "Price Undercharged / Billing Error",
    noteType: "debit_note",
    stockImpact: "value_only",
    etaxCode: "DBNG01",
    isActive: true,
  },
  {
    id: "r-dn-02",
    code: "DN-02",
    nameTh: "ส่งมอบสินค้าเกินกว่าจำนวนในใบสั่งซื้อและลูกค้าตกลงรับ",
    nameEn: "Extra Quantity Delivered Accepted",
    noteType: "debit_note",
    stockImpact: "deduct_stock",
    etaxCode: "DBNG02",
    isActive: true,
  },
  {
    id: "r-dn-03",
    code: "DN-03",
    nameTh: "เพิ่มค่าขนส่งหรือบริการเพิ่มเติมที่ไม่ได้เรียกเก็บในตอนแรก",
    nameEn: "Additional Freight / Service Charge",
    noteType: "debit_note",
    stockImpact: "value_only",
    etaxCode: "DBNG03",
    isActive: true,
  },
];

export interface SystemCreditDebitReasonsScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemCreditDebitReasonsScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: SystemCreditDebitReasonsScreenProps) {
  const isTh = language === "th";
  const [reasons, setReasons] = useState<CreditDebitReasonItem[]>(DEFAULT_REASONS);
  const [selectedId, setSelectedId] = useState<string>(DEFAULT_REASONS[0].id);
  const [typeFilter, setTypeFilter] = useState<"all" | NoteType>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 40,
    min: 25,
    max: 60,
    storageKey: "bcai_split_credit_debit_reasons",
  });

  const selectedReason = useMemo(() => {
    return reasons.find((r) => r.id === selectedId) || reasons[0];
  }, [reasons, selectedId]);

  const filteredReasons = useMemo(() => {
    return reasons.filter((r) => {
      const matchType = typeFilter === "all" || r.noteType === typeFilter;
      const q = searchQuery.toLowerCase().trim();
      const matchSearch =
        !q ||
        r.code.toLowerCase().includes(q) ||
        r.nameTh.toLowerCase().includes(q) ||
        r.nameEn.toLowerCase().includes(q) ||
        r.etaxCode.toLowerCase().includes(q);
      return matchType && matchSearch;
    });
  }, [reasons, typeFilter, searchQuery]);

  const handleUpdate = <K extends keyof CreditDebitReasonItem>(
    field: K,
    val: CreditDebitReasonItem[K]
  ) => {
    setReasons((prev) =>
      prev.map((r) => (r.id === selectedReason.id ? { ...r, [field]: val } : r))
    );
  };

  const handleAddReason = () => {
    const newReason: CreditDebitReasonItem = {
      id: `reason-${Date.now()}`,
      code: "CN-NEW",
      nameTh: "สาเหตุการลดหนี้/เพิ่มหนี้ใหม่",
      nameEn: "New Reason",
      noteType: "credit_note",
      stockImpact: "value_only",
      etaxCode: "CDNG99",
      isActive: true,
    };
    setReasons((prev) => [...prev, newReason]);
    setSelectedId(newReason.id);
  };

  const handleSave = () => {
    setNotification(
      isTh
        ? `บันทึกสาเหตุการลดหนี้/เพิ่มหนี้ (${selectedReason.code} - ${selectedReason.nameTh}) เรียบร้อยแล้ว`
        : `Reason (${selectedReason.code}) saved successfully.`
    );
    setTimeout(() => setNotification(null), 3500);
  };

  const handleReset = () => {
    setReasons(DEFAULT_REASONS);
    setNotification(isTh ? "คืนค่าการตั้งค่าเริ่มต้นเรียบร้อย" : "Reset reasons to defaults");
    setTimeout(() => setNotification(null), 3000);
  };

  return (
    <div className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 font-sans text-xs">
      {/* Header Bar */}
      <div className="flex-none p-3.5 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-wrap items-center justify-between gap-3 shadow-xs">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-indigo-100 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
            <FileSpreadsheet className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-semibold text-slate-900 dark:text-slate-100">
                {isTh ? "กำหนดสาเหตุการลดหนี้/เพิ่มหนี้" : "Credit & Debit Note Reasons Configuration"}
              </h1>
              <span className="px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-[11px] text-slate-500 dark:text-slate-400">
              {isTh
                ? "จัดการระบบ › กำหนดรหัสเหตุผลในการออกใบลดหนี้/เพิ่มหนี้ ผลกระทบต่อสต็อก และรหัส e-Tax Invoice สรรพากร"
                : "System Settings › Setup credit/debit adjustment reasons, stock movements, and Revenue Department e-Tax codes"}
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
        {/* Left Column: Reasons List */}
        <div
          style={{ width: `${splitPercent}%` }}
          className="h-full flex flex-col border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900"
        >
          {/* Filter Toolbar */}
          <div className="p-3 border-b border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/50 space-y-2">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-slate-700 dark:text-slate-200">
                {isTh ? "สาเหตุทั้งหมด" : "Reasons"} ({filteredReasons.length})
              </span>
              <Button
                size="sm"
                variant="ghost"
                onClick={handleAddReason}
                className="h-7 text-xs gap-1 text-indigo-600 hover:text-indigo-700 hover:bg-indigo-50"
              >
                <Plus className="w-3.5 h-3.5" />
                {isTh ? "เพิ่มสาเหตุ" : "Add"}
              </Button>
            </div>

            <div className="flex items-center gap-1.5">
              <div className="relative flex-1">
                <Search className="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400" />
                <Input
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder={isTh ? "ค้นหารหัส หรือชื่อสาเหตุ..." : "Search reason..."}
                  className="h-7 pl-8 text-xs"
                />
              </div>

              <select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value as "all" | NoteType)}
                className="h-7 px-2 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-[11px]"
              >
                <option value="all">{isTh ? "ทุกประเภท" : "All Types"}</option>
                <option value="credit_note">{isTh ? "ใบลดหนี้ (CN)" : "Credit Note"}</option>
                <option value="debit_note">{isTh ? "ใบเพิ่มหนี้ (DN)" : "Debit Note"}</option>
              </select>
            </div>
          </div>

          <div className="flex-1 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800">
            {filteredReasons.map((reason) => {
              const isSelected = reason.id === selectedReason.id;
              const isCN = reason.noteType === "credit_note";
              return (
                <div
                  key={reason.id}
                  onClick={() => setSelectedId(reason.id)}
                  className={`p-3.5 cursor-pointer transition-colors ${
                    isSelected
                      ? "bg-indigo-50/80 dark:bg-indigo-950/40 border-l-4 border-indigo-600"
                      : "hover:bg-slate-50 dark:hover:bg-slate-800/40"
                  }`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-mono font-bold text-indigo-600 dark:text-indigo-400 text-sm">
                      {reason.code}
                    </span>
                    <div className="flex items-center gap-1.5">
                      <span
                        className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                          isCN
                            ? "bg-rose-100 text-rose-700 dark:bg-rose-950/60 dark:text-rose-400"
                            : "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-400"
                        }`}
                      >
                        {isCN ? (isTh ? "ลดหนี้ (CN)" : "CN") : (isTh ? "เพิ่มหนี้ (DN)" : "DN")}
                      </span>
                      <span className="font-mono text-[10px] bg-slate-100 dark:bg-slate-800 text-slate-500 px-1.5 py-0.5 rounded">
                        {reason.etaxCode}
                      </span>
                    </div>
                  </div>
                  <p className="text-slate-800 dark:text-slate-200 font-medium">{reason.nameTh}</p>
                  <p className="mt-1 text-[11px] text-slate-400">
                    {reason.stockImpact === "return_to_stock"
                      ? (isTh ? "• รับคืนเข้าสต็อกสินค้า" : "• Returns goods to stock")
                      : reason.stockImpact === "deduct_stock"
                      ? (isTh ? "• ตัดจ่ายสินค้าออกจากสต็อก" : "• Deducts goods from stock")
                      : (isTh ? "• ปรับมูลค่าเงิน ไม่แตะต้องสต็อก" : "• Price adjustment only")}
                  </p>
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

        {/* Right Column: Reason Detail Form */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="h-full flex flex-col bg-slate-50/50 dark:bg-slate-950 overflow-y-auto p-4 gap-4"
        >
          {/* Form Card */}
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5 shadow-xs">
            <div className="flex items-center gap-2 pb-3 mb-4 border-b border-slate-100 dark:border-slate-800">
              <Sliders className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
              <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
                {isTh ? "รายละเอียดสาเหตุการออกเอกสาร" : "Reason Specifications"}
              </h2>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "รหัสสาเหตุ (Reason Code)" : "Reason Code"}
                </label>
                <Input
                  value={selectedReason.code}
                  onChange={(e) => handleUpdate("code", e.target.value.toUpperCase())}
                  className="font-mono uppercase font-bold text-indigo-600 dark:text-indigo-400"
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "รหัสมาตรฐาน e-Tax กรมสรรพากร" : "Revenue Dept e-Tax Code"}
                </label>
                <Input
                  value={selectedReason.etaxCode}
                  onChange={(e) => handleUpdate("etaxCode", e.target.value.toUpperCase())}
                  className="font-mono uppercase font-bold text-slate-700 dark:text-slate-200"
                  placeholder="CDNG01"
                />
                <p className="mt-1 text-[10px] text-slate-400">
                  {isTh ? "รหัสตามโครงสร้าง XML e-Tax Invoice สรรพากร" : "Official XML reason code"}
                </p>
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ชื่อสาเหตุภาษาไทย" : "Thai Reason Name"}
                </label>
                <Input
                  value={selectedReason.nameTh}
                  onChange={(e) => handleUpdate("nameTh", e.target.value)}
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ชื่อสาเหตุภาษาอังกฤษ" : "English Reason Name"}
                </label>
                <Input
                  value={selectedReason.nameEn}
                  onChange={(e) => handleUpdate("nameEn", e.target.value)}
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ประเภทเอกสารที่อนุญาตให้ใช้" : "Applicable Document Type"}
                </label>
                <select
                  value={selectedReason.noteType}
                  onChange={(e) => handleUpdate("noteType", e.target.value as NoteType)}
                  className="w-full h-8 px-2.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs"
                >
                  <option value="credit_note">{isTh ? "ใบลดหนี้ (Credit Note Only)" : "Credit Note Only"}</option>
                  <option value="debit_note">{isTh ? "ใบเพิ่มหนี้ (Debit Note Only)" : "Debit Note Only"}</option>
                  <option value="both">{isTh ? "ใช้ได้ทั้งลดหนี้และเพิ่มหนี้ (Both)" : "Both"}</option>
                </select>
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ผลกระทบต่อสต็อกสินค้า" : "Inventory Stock Movement Impact"}
                </label>
                <select
                  value={selectedReason.stockImpact}
                  onChange={(e) => handleUpdate("stockImpact", e.target.value as StockImpact)}
                  className="w-full h-8 px-2.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs"
                >
                  <option value="value_only">{isTh ? "ปรับมูลค่าเงินอย่างเดียว ไม่แตะต้องสต็อก" : "Price/Value adjustment only (no stock change)"}</option>
                  <option value="return_to_stock">{isTh ? "รับคืนสินค้ากลับเข้าสต็อก (Return to Stock)" : "Return goods to stock"}</option>
                  <option value="deduct_stock">{isTh ? "ตัดสินค้าออกจากสต็อกเพิ่ม (Deduct from Stock)" : "Deduct extra goods from stock"}</option>
                </select>
              </div>
            </div>

            <div className="mt-5 pt-4 border-t border-slate-100 dark:border-slate-800">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedReason.isActive}
                  onChange={(e) => handleUpdate("isActive", e.target.checked)}
                  className="rounded text-indigo-600 focus:ring-indigo-500 w-4 h-4"
                />
                <span className="text-slate-700 dark:text-slate-300">
                  {isTh ? "เปิดใช้งานสาเหตุนี้ในระบบ (Active)" : "Enable this reason in document workflows"}
                </span>
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
