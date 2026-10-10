"use client";

import { useMemo, useRef, useState } from "react";
import {
  Boxes,
  CheckCircle2,
  FolderTree,
  MapPin,
  Package,
  Plus,
  RotateCcw,
  Save,
  Search,
  Sliders,
  Warehouse,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export interface WarehouseItem {
  id: string;
  code: string;
  name: string;
  type: string;
  branchCode: string;
  isActive: boolean;
}

export interface ProductWarehouseRule {
  id: string;
  warehouseCode: string;
  productCode: string;
  productName: string;
  category: string;
  unit: string;
  shelfBin: string;
  minQty: number;
  maxQty: number;
  reorderPoint: number;
  isPrimary: boolean;
  isEnabled: boolean;
}

const INITIAL_WAREHOUSES: WarehouseItem[] = [
  { id: "wh-1", code: "WH-01", name: "คลังสินค้าสำเร็จรูปหลัก (HQ)", type: "สำเร็จรูป", branchCode: "00000", isActive: true },
  { id: "wh-2", code: "WH-02", name: "คลังวัตถุดิบและอุปกรณ์ (HQ)", type: "วัตถุดิบ", branchCode: "00000", isActive: true },
  { id: "wh-3", code: "WH-03", name: "คลังสินค้าระหว่างผลิต (WIP)", type: "งานระหว่างทำ", branchCode: "00000", isActive: true },
  { id: "wh-4", code: "WH-04", name: "คลังสินค้าชำรุด/รอทำลาย", type: "ของเสีย", branchCode: "00000", isActive: true },
];

const INITIAL_RULES: ProductWarehouseRule[] = [
  {
    id: "r-01",
    warehouseCode: "WH-01",
    productCode: "MAT-CM-01",
    productName: "ปูนซีเมนต์ผสมสำเร็จ ปอร์ตแลนด์ประเภท 1 (ถุง 50 กก.)",
    category: "วัสดุโครงสร้าง",
    unit: "ถุง",
    shelfBin: "A-01-01",
    minQty: 50,
    maxQty: 500,
    reorderPoint: 100,
    isPrimary: true,
    isEnabled: true,
  },
  {
    id: "r-02",
    warehouseCode: "WH-01",
    productCode: "MAT-ST-02",
    productName: "เหล็กเส้นข้ออ้อย SD40 ขนาด 16 มม. (ยาว 10 ม.)",
    category: "วัสดุโครงสร้าง",
    unit: "เส้น",
    shelfBin: "B-02-04",
    minQty: 30,
    maxQty: 300,
    reorderPoint: 80,
    isPrimary: true,
    isEnabled: true,
  },
  {
    id: "r-03",
    warehouseCode: "WH-01",
    productCode: "FIN-TL-05",
    productName: "กระเบื้องเซรามิกปูพื้น แกรนิตโต้ 60x60 ซม. สีครีม",
    category: "กระเบื้องและสุขภัณฑ์",
    unit: "กล่อง",
    shelfBin: "C-01-02",
    minQty: 20,
    maxQty: 200,
    reorderPoint: 40,
    isPrimary: true,
    isEnabled: true,
  },
  {
    id: "r-04",
    warehouseCode: "WH-02",
    productCode: "RAW-SD-01",
    productName: "ทรายละเอียดคัดพิเศษสำหรับงานฉาบ",
    category: "วัตถุดิบ",
    unit: "คิว",
    shelfBin: "YARD-01",
    minQty: 10,
    maxQty: 100,
    reorderPoint: 25,
    isPrimary: true,
    isEnabled: true,
  },
];

export interface SystemProductByWarehouseScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemProductByWarehouseScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: SystemProductByWarehouseScreenProps) {
  const isTh = language === "th";
  const [warehouses] = useState<WarehouseItem[]>(INITIAL_WAREHOUSES);
  const [selectedWhCode, setSelectedWhCode] = useState<string>(INITIAL_WAREHOUSES[0].code);
  const [rules, setRules] = useState<ProductWarehouseRule[]>(INITIAL_RULES);
  const [searchQuery, setSearchQuery] = useState("");
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 35,
    min: 25,
    max: 55,
    storageKey: "bcai_split_product_warehouse",
  });

  const selectedWarehouse = useMemo(() => {
    return warehouses.find((w) => w.code === selectedWhCode) || warehouses[0];
  }, [warehouses, selectedWhCode]);

  const currentRules = useMemo(() => {
    return rules.filter((r) => r.warehouseCode === selectedWhCode);
  }, [rules, selectedWhCode]);

  const filteredRules = useMemo(() => {
    const q = searchQuery.toLowerCase().trim();
    if (!q) return currentRules;
    return currentRules.filter(
      (r) =>
        r.productCode.toLowerCase().includes(q) ||
        r.productName.toLowerCase().includes(q) ||
        r.shelfBin.toLowerCase().includes(q)
    );
  }, [currentRules, searchQuery]);

  const handleUpdateRule = <K extends keyof ProductWarehouseRule>(
    id: string,
    field: K,
    val: ProductWarehouseRule[K]
  ) => {
    setRules((prev) =>
      prev.map((r) => (r.id === id ? { ...r, [field]: val } : r))
    );
  };

  const handleSave = () => {
    setNotification(
      isTh
        ? `บันทึกการกำหนดสินค้าตามคลัง (${selectedWarehouse.code} - ${selectedWarehouse.name}) เรียบร้อยแล้ว`
        : `Product warehouse rules for (${selectedWarehouse.code}) saved successfully.`
    );
    setTimeout(() => setNotification(null), 3500);
  };

  const handleReset = () => {
    setRules(INITIAL_RULES);
    setNotification(isTh ? "คืนค่าการตั้งค่าเริ่มต้นเรียบร้อย" : "Reset rules to defaults");
    setTimeout(() => setNotification(null), 3000);
  };

  return (
    <div className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 font-sans text-xs">
      {/* Header Bar */}
      <div className="flex-none p-3.5 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-wrap items-center justify-between gap-3 shadow-xs">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-indigo-100 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
            <Warehouse className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-semibold text-slate-900 dark:text-slate-100">
                {isTh ? "กำหนดสินค้าตามคลัง" : "Products by Warehouse Configuration"}
              </h1>
              <span className="px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-[11px] text-slate-500 dark:text-slate-400">
              {isTh
                ? "จัดการระบบ › กำหนดตำแหน่งจัดเก็บ (Shelf/Bin), ปริมาณ Min/Max และจุดสั่งซื้อแยกตามคลังสินค้า"
                : "System Settings › Setup shelf/bin location, Min/Max stock thresholds and reorder points per warehouse"}
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
        {/* Left Column: Warehouse List */}
        <div
          style={{ width: `${splitPercent}%` }}
          className="h-full flex flex-col border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900"
        >
          <div className="p-3 border-b border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/50 flex items-center justify-between">
            <span className="font-semibold text-slate-700 dark:text-slate-200">
              {isTh ? "คลังสินค้าทั้งหมด" : "Warehouses"} ({warehouses.length})
            </span>
            <span className="text-[11px] text-slate-500 font-mono">Select Warehouse</span>
          </div>

          <div className="flex-1 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800">
            {warehouses.map((wh) => {
              const isSelected = wh.code === selectedWarehouse.code;
              const count = rules.filter((r) => r.warehouseCode === wh.code).length;
              return (
                <div
                  key={wh.id}
                  onClick={() => setSelectedWhCode(wh.code)}
                  className={`p-3.5 cursor-pointer transition-colors ${
                    isSelected
                      ? "bg-indigo-50/80 dark:bg-indigo-950/40 border-l-4 border-indigo-600"
                      : "hover:bg-slate-50 dark:hover:bg-slate-800/40"
                  }`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-mono font-bold text-indigo-600 dark:text-indigo-400 text-sm">
                      {wh.code}
                    </span>
                    <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
                      {wh.type}
                    </span>
                  </div>
                  <p className="text-slate-800 dark:text-slate-200 font-medium">{wh.name}</p>
                  <div className="mt-2 flex items-center justify-between text-[11px] text-slate-500">
                    <span>{isTh ? "สาขา: " : "Branch: "} {wh.branchCode}</span>
                    <span className="font-mono bg-indigo-50 dark:bg-indigo-950 px-2 py-0.5 rounded text-indigo-600 dark:text-indigo-400">
                      {count} {isTh ? "รายการสินค้า" : "items"}
                    </span>
                  </div>
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

        {/* Right Column: Products in Warehouse Table */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="h-full flex flex-col bg-white dark:bg-slate-900 overflow-hidden"
        >
          {/* Action and Search Toolbar */}
          <div className="p-3 border-b border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3 bg-slate-50/50 dark:bg-slate-900/50">
            <div className="flex items-center gap-2">
              <span className="font-semibold text-slate-800 dark:text-slate-100 text-sm">
                {selectedWarehouse.code} - {selectedWarehouse.name}
              </span>
              <span className="text-[11px] text-slate-500 font-mono">
                ({filteredRules.length} {isTh ? "รายการ" : "items"})
              </span>
            </div>

            <div className="flex items-center gap-2">
              <div className="relative w-64">
                <Search className="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400" />
                <Input
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder={isTh ? "ค้นหารหัส, ชื่อสินค้า, ช่องวาง..." : "Search code, name, shelf..."}
                  className="h-8 pl-8 text-xs"
                />
              </div>
            </div>
          </div>

          {/* Table Container */}
          <div className="flex-1 overflow-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead className="sticky top-0 bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 font-medium border-b border-slate-200 dark:border-slate-700 z-10">
                <tr>
                  <th className="py-2.5 px-3 w-10 text-center">{isTh ? "ใช้" : "Active"}</th>
                  <th className="py-2.5 px-3 w-32">{isTh ? "รหัสสินค้า" : "Code"}</th>
                  <th className="py-2.5 px-3 min-w-[200px]">{isTh ? "ชื่อสินค้า / หมวดหมู่" : "Product Name"}</th>
                  <th className="py-2.5 px-3 w-28">{isTh ? "ที่เก็บ (Shelf/Bin)" : "Shelf/Bin"}</th>
                  <th className="py-2.5 px-3 w-24 text-right">{isTh ? "Min Stock" : "Min Qty"}</th>
                  <th className="py-2.5 px-3 w-24 text-right">{isTh ? "Max Stock" : "Max Qty"}</th>
                  <th className="py-2.5 px-3 w-28 text-right">{isTh ? "จุดสั่งซื้อ (ROP)" : "Reorder Point"}</th>
                  <th className="py-2.5 px-3 w-20 text-center">{isTh ? "คลังหลัก" : "Primary"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                {filteredRules.length === 0 ? (
                  <tr>
                    <td colSpan={8} className="py-12 text-center text-slate-400">
                      <Package className="w-8 h-8 mx-auto mb-2 opacity-40" />
                      <p>{isTh ? "ไม่พบสินค้าในคลังนี้ตามเงื่อนไขค้นหา" : "No products found in this warehouse"}</p>
                    </td>
                  </tr>
                ) : (
                  filteredRules.map((rule) => (
                    <tr
                      key={rule.id}
                      className={`hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors ${
                        !rule.isEnabled ? "opacity-50" : ""
                      }`}
                    >
                      <td className="py-2 px-3 text-center">
                        <input
                          type="checkbox"
                          checked={rule.isEnabled}
                          onChange={(e) => handleUpdateRule(rule.id, "isEnabled", e.target.checked)}
                          className="rounded text-indigo-600 focus:ring-indigo-500 w-3.5 h-3.5"
                        />
                      </td>
                      <td className="py-2 px-3 font-mono font-bold text-indigo-600 dark:text-indigo-400">
                        {rule.productCode}
                      </td>
                      <td className="py-2 px-3">
                        <p className="font-medium text-slate-800 dark:text-slate-100">{rule.productName}</p>
                        <p className="text-[10px] text-slate-400">
                          {rule.category} • หน่วย: {rule.unit}
                        </p>
                      </td>
                      <td className="py-2 px-3">
                        <Input
                          value={rule.shelfBin}
                          onChange={(e) => handleUpdateRule(rule.id, "shelfBin", e.target.value.toUpperCase())}
                          className="h-7 text-xs font-mono uppercase"
                          placeholder="A-01"
                        />
                      </td>
                      <td className="py-2 px-3 text-right">
                        <Input
                          type="number"
                          value={rule.minQty}
                          onChange={(e) => handleUpdateRule(rule.id, "minQty", Number(e.target.value))}
                          className="h-7 text-xs font-mono text-right w-20 ml-auto"
                        />
                      </td>
                      <td className="py-2 px-3 text-right">
                        <Input
                          type="number"
                          value={rule.maxQty}
                          onChange={(e) => handleUpdateRule(rule.id, "maxQty", Number(e.target.value))}
                          className="h-7 text-xs font-mono text-right w-20 ml-auto"
                        />
                      </td>
                      <td className="py-2 px-3 text-right">
                        <Input
                          type="number"
                          value={rule.reorderPoint}
                          onChange={(e) => handleUpdateRule(rule.id, "reorderPoint", Number(e.target.value))}
                          className="h-7 text-xs font-mono text-right w-20 ml-auto"
                        />
                      </td>
                      <td className="py-2 px-3 text-center">
                        <input
                          type="checkbox"
                          checked={rule.isPrimary}
                          onChange={(e) => handleUpdateRule(rule.id, "isPrimary", e.target.checked)}
                          className="rounded text-amber-500 focus:ring-amber-500 w-3.5 h-3.5"
                          title={isTh ? "คลังเก็บหลักสำหรับสินค้านี้" : "Primary warehouse"}
                        />
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
