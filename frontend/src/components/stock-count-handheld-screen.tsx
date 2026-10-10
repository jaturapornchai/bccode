"use client";

import { useMemo, useRef, useState } from "react";
import {
  CheckCircle2,
  Package,
  Printer,
  QrCode,
  Search,
  Smartphone,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export interface HandheldCountItem {
  id: string;
  itemCode: string;
  itemName: string;
  barcode: string;
  shelf: string;
  systemQty: number;
  countedQty: number;
  unit: string;
}

export interface HandheldBatchDoc {
  id: string;
  batchNo: string;
  countDate: string;
  warehouseCode: string;
  warehouseName: string;
  deviceModel: string;
  deviceId: string;
  countedBy: string;
  status: "in_progress" | "verified" | "adjusted";
  items: HandheldCountItem[];
  remark: string;
}

const INITIAL_BATCHES: HandheldBatchDoc[] = [
  {
    id: "batch-01",
    batchNo: "HH-6910-001",
    countDate: "2026-10-09",
    warehouseCode: "WH-01",
    warehouseName: "คลังใหญ่พระประแดง (สินค้าโครงสร้าง)",
    deviceModel: "Zebra TC26 Android Barcode Scanner",
    deviceId: "HH-DEV-04",
    countedBy: "นายเกรียงไกร คลังดี",
    status: "verified",
    remark: "ตรวจนับประจำสัปดาห์ โซน A-01 ถึง A-05",
    items: [
      {
        id: "item-01",
        itemCode: "MAT-CM-01",
        itemName: "ปูนซีเมนต์ผสมสำเร็จ ปอร์ตแลนด์ประเภท 1 (ถุง 50 กก.)",
        barcode: "8850123450012",
        shelf: "A-01-04",
        systemQty: 850,
        countedQty: 850,
        unit: "ถุง",
      },
      {
        id: "item-02",
        itemCode: "MAT-CM-02",
        itemName: "ปูนซีเมนต์ผสมเสร็จ ก่อฉาบ สูตรพิเศษ (ถุง 50 กก.)",
        barcode: "8850123450029",
        shelf: "A-01-06",
        systemQty: 420,
        countedQty: 415,
        unit: "ถุง",
      },
      {
        id: "item-03",
        itemCode: "MAT-ST-02",
        itemName: "หินก่อสร้าง เบอร์ 1 ผสมคอนกรีต (คิว)",
        barcode: "8850123450036",
        shelf: "YARD-02",
        systemQty: 250,
        countedQty: 250,
        unit: "คิว",
      },
    ],
  },
  {
    id: "batch-02",
    batchNo: "HH-6910-002",
    countDate: "2026-10-10",
    warehouseCode: "WH-02",
    warehouseName: "คลังเหล็กและโลหะภัณฑ์",
    deviceModel: "Honeywell EDA51 Handheld Terminal",
    deviceId: "HH-DEV-02",
    countedBy: "นายสมชาย ตาไว",
    status: "in_progress",
    remark: "ตรวจนับเหล็กเส้นข้ออ้อยและเหล็กรูปพรรณ ล็อตนำเข้าใหม่",
    items: [
      {
        id: "item-04",
        itemCode: "STL-DB12",
        itemName: "เหล็กเส้นข้ออ้อย มอก. SD40 ขนาด DB12 ยาว 10 ม.",
        barcode: "8850123488123",
        shelf: "BAY-03",
        systemQty: 1400,
        countedQty: 1400,
        unit: "เส้น",
      },
      {
        id: "item-05",
        itemCode: "STL-DB16",
        itemName: "เหล็กเส้นข้ออ้อย มอก. SD40 ขนาด DB16 ยาว 10 ม.",
        barcode: "8850123488161",
        shelf: "BAY-04",
        systemQty: 800,
        countedQty: 804,
        unit: "เส้น",
      },
    ],
  },
  {
    id: "batch-03",
    batchNo: "HH-6910-003",
    countDate: "2026-09-30",
    warehouseCode: "WH-03",
    warehouseName: "คลังกระเบื้องและสุขภัณฑ์",
    deviceModel: "Datalogic Memor 10 Terminal",
    deviceId: "HH-DEV-01",
    countedBy: "นางสาวสมร มีทรัพย์",
    status: "adjusted",
    remark: "ตรวจนับสิ้นไตรมาส 3 ปรับปรุงยอดเข้าสต็อกเรียบร้อย",
    items: [
      {
        id: "item-06",
        itemCode: "FLR-GRN60",
        itemName: "กระเบื้องปูพื้นแกรนิตโต้ นาโนไวท์ 60x60 ซม. (เกรด A)",
        barcode: "8850123499214",
        shelf: "PALLET-12",
        systemQty: 320,
        countedQty: 320,
        unit: "กล่อง",
      },
    ],
  },
];

export interface StockCountHandheldScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function StockCountHandheldScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: StockCountHandheldScreenProps) {
  const isTh = language === "th";
  const [batches, setBatches] = useState<HandheldBatchDoc[]>(INITIAL_BATCHES);
  const [selectedId, setSelectedId] = useState<string>(INITIAL_BATCHES[0].id);
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [barcodeInput, setBarcodeInput] = useState("");
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 38,
    min: 25,
    max: 60,
    storageKey: "bcai_split_stock_count_hh",
  });

  const selectedBatch = useMemo(() => {
    return batches.find((b) => b.id === selectedId) || batches[0];
  }, [batches, selectedId]);

  const filteredBatches = useMemo(() => {
    return batches.filter((b) => {
      const matchStatus = statusFilter === "all" || b.status === statusFilter;
      const q = searchQuery.toLowerCase().trim();
      const matchSearch =
        !q ||
        b.batchNo.toLowerCase().includes(q) ||
        b.warehouseName.toLowerCase().includes(q) ||
        b.deviceModel.toLowerCase().includes(q) ||
        b.countedBy.toLowerCase().includes(q);
      return matchStatus && matchSearch;
    });
  }, [batches, statusFilter, searchQuery]);

  // Barcode scan quick add
  const handleQuickScan = (e: React.FormEvent) => {
    e.preventDefault();
    if (!barcodeInput.trim()) return;

    const matchedItem = selectedBatch.items.find(
      (item) => item.barcode === barcodeInput.trim() || item.itemCode === barcodeInput.trim()
    );

    if (matchedItem) {
      setBatches((prev) =>
        prev.map((b) => {
          if (b.id === selectedBatch.id) {
            return {
              ...b,
              items: b.items.map((it) =>
                it.id === matchedItem.id ? { ...it, countedQty: it.countedQty + 1 } : it
              ),
            };
          }
          return b;
        })
      );
      setNotification(
        isTh
          ? `สแกนบาร์โค้ด ${matchedItem.itemCode} สำเร็จ (+1 ยอดตรวจนับ)`
          : `Scanned ${matchedItem.itemCode} (+1 count)`
      );
    } else {
      setNotification(
        isTh ? `ไม่พบสินค้าบาร์โค้ด "${barcodeInput}" ในชุดตรวจนับนี้` : `Barcode "${barcodeInput}" not in batch`
      );
    }
    setBarcodeInput("");
  };

  const handleVerifyBatch = () => {
    setBatches((prev) =>
      prev.map((b) => {
        if (b.id === selectedBatch.id) {
          return { ...b, status: "verified" };
        }
        return b;
      })
    );
    setNotification(isTh ? `ยืนยันผลการตรวจนับ ${selectedBatch.batchNo} เรียบร้อย` : `Batch ${selectedBatch.batchNo} verified`);
  };

  const totalVariance = useMemo(() => {
    return selectedBatch.items.reduce((acc, it) => acc + (it.countedQty - it.systemQty), 0);
  }, [selectedBatch.items]);

  return (
    <div
      ref={containerRef}
      className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 text-slate-800 dark:text-slate-100 overflow-hidden font-sans"
    >
      {/* Top Header Bar */}
      <div className="flex flex-wrap items-center justify-between px-4 py-2.5 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 gap-3">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 border border-indigo-200/60 dark:border-indigo-800/40">
            <Smartphone className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold tracking-tight text-slate-900 dark:text-white">
                {isTh ? "ตรวจนับสต็อกโดย Hand Held (Mobile Count)" : "Hand Held Stock Count"}
              </h1>
              <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800">
                Champ Parity 50000+
              </span>
            </div>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {isTh
                ? "นำเข้าผลตรวจนับสินค้าจากอุปกรณ์ Barcode Handheld / Mobile Terminal เปรียบเทียบผลต่างสต็อกและปรับปรุงยอด"
                : "Import stock counts from mobile barcode terminals, reconcile variance & update inventory"}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {selectedBatch.status === "in_progress" && (
            <Button
              size="sm"
              onClick={handleVerifyBatch}
              className="text-xs h-8 gap-1.5 bg-emerald-600 hover:bg-emerald-700 text-white font-medium shadow-sm"
            >
              <CheckCircle2 className="w-3.5 h-3.5" />
              <span>{isTh ? "ยืนยันผลตรวจนับ" : "Verify Batch"}</span>
            </Button>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={() => window.print()}
            className="text-xs h-8 gap-1.5"
          >
            <Printer className="w-3.5 h-3.5" />
            <span>{isTh ? "พิมพ์ผลตรวจนับ" : "Print"}</span>
          </Button>
        </div>
      </div>

      {/* Notification Banner */}
      {notification && (
        <div className="flex items-center justify-between px-4 py-1.5 bg-indigo-50 dark:bg-indigo-950/40 text-indigo-800 dark:text-indigo-200 text-xs border-b border-indigo-200 dark:border-indigo-900">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 text-indigo-600" />
            <span>{notification}</span>
          </div>
          <button onClick={() => setNotification(null)} className="text-slate-400 hover:text-slate-600">✕</button>
        </div>
      )}

      {/* Main Split Layout */}
      <div className="flex flex-1 overflow-hidden relative">
        {/* Left Pane: Filter & Batch List */}
        <div
          style={{ width: `${splitPercent}%` }}
          className="flex flex-col h-full bg-white dark:bg-slate-900 border-r border-slate-200 dark:border-slate-800 overflow-hidden"
        >
          {/* Search & Filter Header */}
          <div className="p-3 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50 space-y-2">
            <div className="relative">
              <Search className="w-3.5 h-3.5 absolute left-2.5 top-2.5 text-slate-400" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={isTh ? "ค้นหาชุดตรวจนับ / คลัง / อุปกรณ์..." : "Search batch, WH or device..."}
                className="pl-8 h-8 text-xs bg-white dark:bg-slate-950"
              />
            </div>

            {/* Status Pills */}
            <div className="flex flex-wrap gap-1 items-center">
              {[
                { id: "all", th: "ทั้งหมด", en: "All" },
                { id: "in_progress", th: "กำลังนับ", en: "Counting" },
                { id: "verified", th: "ยืนยันแล้ว", en: "Verified" },
                { id: "adjusted", th: "ปรับสต็อกแล้ว", en: "Adjusted" },
              ].map((st) => (
                <button
                  key={st.id}
                  onClick={() => setStatusFilter(st.id)}
                  className={`text-[10px] px-2 py-0.5 rounded-full border transition-all ${
                    statusFilter === st.id
                      ? "bg-indigo-600 text-white border-indigo-600 font-semibold"
                      : "bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 border-slate-200 dark:border-slate-800 hover:bg-slate-100"
                  }`}
                >
                  {isTh ? st.th : st.en}
                </button>
              ))}
            </div>
          </div>

          {/* List Items */}
          <div className="flex-1 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800/60 p-1.5 space-y-1">
            {filteredBatches.map((b) => {
              const isSelected = b.id === selectedBatch.id;
              return (
                <div
                  key={b.id}
                  onClick={() => setSelectedId(b.id)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") setSelectedId(b.id);
                  }}
                  tabIndex={0}
                  role="button"
                  className={`flex flex-col p-2.5 rounded-lg cursor-pointer transition-all border text-left ${
                    isSelected
                      ? "bg-indigo-50/90 dark:bg-indigo-950/40 border-indigo-300 dark:border-indigo-700 shadow-sm"
                      : "border-transparent hover:bg-slate-100/70 dark:hover:bg-slate-800/60"
                  }`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-mono text-xs font-bold text-slate-900 dark:text-slate-100">
                      {b.batchNo}
                    </span>
                    <span
                      className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${
                        b.status === "adjusted"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-300/50"
                          : b.status === "verified"
                          ? "bg-blue-100 text-blue-800 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-300/50"
                          : "bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-300 border border-amber-300/50"
                      }`}
                    >
                      {b.status === "adjusted"
                        ? isTh ? "ปรับสต็อกแล้ว" : "Adjusted"
                        : b.status === "verified"
                        ? isTh ? "ยืนยันผลแล้ว" : "Verified"
                        : isTh ? "กำลังตรวจนับ" : "Counting"}
                    </span>
                  </div>

                  <p className="text-xs font-semibold text-slate-800 dark:text-slate-200 mt-1 truncate">
                    {b.warehouseName}
                  </p>

                  <div className="flex items-center justify-between mt-1 text-[11px] text-slate-500">
                    <span className="truncate max-w-[140px]">{b.deviceModel}</span>
                    <span>{b.countDate}</span>
                  </div>
                </div>
              );
            })}
          </div>

          {/* List Footer */}
          <div className="p-3 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 flex justify-between items-center">
            <span>{isTh ? `รวม ${filteredBatches.length} ชุดตรวจนับ` : `${filteredBatches.length} batches`}</span>
            <span className="font-semibold text-slate-700 dark:text-slate-300">
              {filteredBatches.reduce((acc, b) => acc + b.items.length, 0)} รายการสินค้า
            </span>
          </div>
        </div>

        {/* Resizable Splitter */}
        <ResizableSplitter
          isResizing={isResizing}
          onMouseDown={startResize}
          onKeyDown={adjustWithKeyboard}
        />

        {/* Right Pane: Batch Details, Quick Scanner & Reconciliation Table */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="flex flex-col h-full bg-slate-50 dark:bg-slate-950 overflow-y-auto p-4 space-y-4"
        >
          {/* Card 1: Batch Info */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-3">
            <div className="flex flex-wrap items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800 gap-2">
              <div className="flex items-center gap-2">
                <Smartphone className="w-5 h-5 text-indigo-600" />
                <div>
                  <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                    <span>{selectedBatch.batchNo}</span>
                    <span className="text-xs text-slate-500 font-normal">({selectedBatch.warehouseName})</span>
                  </h2>
                  <p className="text-xs text-slate-400">อุปกรณ์: {selectedBatch.deviceModel} (ID: {selectedBatch.deviceId}) • ผู้ตรวจ: {selectedBatch.countedBy}</p>
                </div>
              </div>

              {/* Barcode Quick Scanner Form */}
              <form onSubmit={handleQuickScan} className="flex items-center gap-1.5">
                <div className="relative">
                  <QrCode className="w-3.5 h-3.5 absolute left-2 top-2.5 text-slate-400" />
                  <Input
                    value={barcodeInput}
                    onChange={(e) => setBarcodeInput(e.target.value)}
                    placeholder={isTh ? "สแกนบาร์โค้ดสินค้าที่นี่..." : "Scan barcode here..."}
                    className="pl-7 h-8 text-xs font-mono w-52 bg-white dark:bg-slate-950"
                  />
                </div>
                <Button type="submit" size="sm" className="text-xs h-8 bg-indigo-600 hover:bg-indigo-700 text-white">
                  {isTh ? "สแกน" : "Scan"}
                </Button>
              </form>
            </div>

            {/* Reconciliation Variance Summary */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
              <div className="p-2.5 bg-slate-50 dark:bg-slate-950 rounded-lg border border-slate-100 dark:border-slate-800">
                <span className="text-[11px] text-slate-500">{isTh ? "จำนวนรายการที่นับ" : "Total Lines"}</span>
                <p className="font-mono text-sm font-bold text-slate-800 dark:text-slate-200 mt-0.5">
                  {selectedBatch.items.length} รายการ
                </p>
              </div>

              <div className="p-2.5 bg-slate-50 dark:bg-slate-950 rounded-lg border border-slate-100 dark:border-slate-800">
                <span className="text-[11px] text-slate-500">{isTh ? "ยอดรวมตามระบบ" : "System Total"}</span>
                <p className="font-mono text-sm font-bold text-slate-800 dark:text-slate-200 mt-0.5">
                  {selectedBatch.items.reduce((acc, it) => acc + it.systemQty, 0).toLocaleString()}
                </p>
              </div>

              <div className="p-2.5 bg-slate-50 dark:bg-slate-950 rounded-lg border border-slate-100 dark:border-slate-800">
                <span className="text-[11px] text-slate-500">{isTh ? "ยอดตรวจนับจริง" : "Counted Total"}</span>
                <p className="font-mono text-sm font-bold text-indigo-600 dark:text-indigo-400 mt-0.5">
                  {selectedBatch.items.reduce((acc, it) => acc + it.countedQty, 0).toLocaleString()}
                </p>
              </div>

              <div className="p-2.5 bg-slate-50 dark:bg-slate-950 rounded-lg border border-slate-100 dark:border-slate-800">
                <span className="text-[11px] text-slate-500">{isTh ? "ผลต่างรวม (Diff)" : "Total Variance"}</span>
                <p
                  className={`font-mono text-sm font-bold mt-0.5 ${
                    totalVariance === 0
                      ? "text-emerald-600"
                      : totalVariance > 0
                      ? "text-blue-600"
                      : "text-rose-600"
                  }`}
                >
                  {totalVariance > 0 ? `+${totalVariance}` : totalVariance}
                </p>
              </div>
            </div>
          </div>

          {/* Card 2: Counted Items Table */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-3">
            <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
              <div className="flex items-center gap-2">
                <Package className="w-4 h-4 text-indigo-500" />
                <h3 className="text-xs font-bold text-slate-900 dark:text-white">
                  {isTh ? "ตารางเปรียบเทียบยอดสต็อกและผลต่าง (Reconciliation Lines)" : "Stock Count Reconciliation"}
                </h3>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 bg-slate-50/70 dark:bg-slate-950/70">
                    <th className="py-2 px-3 font-semibold">{isTh ? "รหัสสินค้า" : "Item Code"}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "ชื่อสินค้า" : "Description"}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "บาร์โค้ด / ที่เก็บ" : "Barcode / Shelf"}</th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "ยอดในระบบ" : "System"}</th>
                    <th className="py-2 px-3 text-right font-semibold text-indigo-600 dark:text-indigo-400">
                      {isTh ? "นับจริง (HH)" : "Counted"}
                    </th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "ผลต่าง" : "Diff"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono text-[11px]">
                  {selectedBatch.items.map((it) => {
                    const diff = it.countedQty - it.systemQty;
                    return (
                      <tr key={it.id} className="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                        <td className="py-2 px-3 font-bold text-indigo-600 dark:text-indigo-400 whitespace-nowrap">
                          {it.itemCode}
                        </td>
                        <td className="py-2 px-3 font-sans text-slate-800 dark:text-slate-200 max-w-[240px] truncate" title={it.itemName}>
                          {it.itemName}
                        </td>
                        <td className="py-2 px-3 text-slate-500">
                          <span>{it.barcode}</span>
                          <span className="ml-1 text-[10px] font-sans px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800">
                            {it.shelf}
                          </span>
                        </td>
                        <td className="py-2 px-3 text-right text-slate-600">
                          {it.systemQty} {it.unit}
                        </td>
                        <td className="py-2 px-3 text-right font-bold text-indigo-600 dark:text-indigo-400">
                          {it.countedQty} {it.unit}
                        </td>
                        <td
                          className={`py-2 px-3 text-right font-bold ${
                            diff === 0
                              ? "text-emerald-600"
                              : diff > 0
                              ? "text-blue-600"
                              : "text-rose-600"
                          }`}
                        >
                          {diff > 0 ? `+${diff}` : diff}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
