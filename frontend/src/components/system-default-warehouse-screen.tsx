"use client";

import { useMemo, useState } from "react";
import {
  Boxes,
  Building,
  CheckCircle2,
  Lock,
  Package,
  RotateCcw,
  Save,
  Sliders,
  Truck,
  Warehouse,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { type LanguageCode } from "@/lib/i18n";

export interface TransactionWarehouseMapping {
  id: string;
  moduleKey: string;
  moduleNameTh: string;
  moduleNameEn: string;
  category: "purchase" | "sale" | "inventory" | "production";
  defaultWarehouseCode: string;
  defaultShelf: string;
  allowOverride: boolean;
  notes: string;
}

const AVAILABLE_WAREHOUSES = [
  { code: "WH-01", name: "คลังสินค้าสำเร็จรูปหลัก (HQ)" },
  { code: "WH-02", name: "คลังวัตถุดิบและอุปกรณ์ (HQ)" },
  { code: "WH-03", name: "คลังสินค้าระหว่างผลิต (WIP)" },
  { code: "WH-04", name: "คลังสินค้าชำรุด/รอทำลาย" },
  { code: "WH-05", name: "คลังสินค้าฝากขาย / รอตรวจสอบ" },
];

const DEFAULT_MAPPINGS: TransactionWarehouseMapping[] = [
  {
    id: "m-01",
    moduleKey: "purchase_receipt",
    moduleNameTh: "งานจัดซื้อ / รับสินค้าเข้า (Purchase Receipt)",
    moduleNameEn: "Purchase Goods Receipt",
    category: "purchase",
    defaultWarehouseCode: "WH-02",
    defaultShelf: "IN-01",
    allowOverride: true,
    notes: "รับสินค้าและวัตถุดิบเข้าคลังวัตถุดิบเป็นค่าตั้งต้น",
  },
  {
    id: "m-02",
    moduleKey: "purchase_return",
    moduleNameTh: "งานจัดซื้อ / ส่งคืนสินค้าให้ผู้ขาย (Purchase Return)",
    moduleNameEn: "Purchase Goods Return",
    category: "purchase",
    defaultWarehouseCode: "WH-02",
    defaultShelf: "RT-01",
    allowOverride: true,
    notes: "ตัดสินค้าคืนออกจากคลังวัตถุดิบ",
  },
  {
    id: "m-03",
    moduleKey: "sale_dispatch",
    moduleNameTh: "งานจัดขาย / ส่งมอบสินค้าให้ลูกค้า (Sale Invoice & Dispatch)",
    moduleNameEn: "Sales Goods Delivery",
    category: "sale",
    defaultWarehouseCode: "WH-01",
    defaultShelf: "OUT-01",
    allowOverride: true,
    notes: "ตัดสต็อกสินค้าสำเร็จรูปออกจากคลังหลัก",
  },
  {
    id: "m-04",
    moduleKey: "sale_return",
    moduleNameTh: "งานจัดขาย / รับคืนสินค้าจากลูกค้า (Sale Return)",
    moduleNameEn: "Sales Goods Return",
    category: "sale",
    defaultWarehouseCode: "WH-05",
    defaultShelf: "QC-01",
    allowOverride: true,
    notes: "รับคืนสินค้าเข้าคลังรอตรวจสอบสภาพก่อนปรับยอด",
  },
  {
    id: "m-05",
    moduleKey: "stock_pickup",
    moduleNameTh: "งานคลัง / เบิกใช้สินค้าและวัตถุดิบ (Material Issue / Pickup)",
    moduleNameEn: "Stock Issue / Pickup",
    category: "inventory",
    defaultWarehouseCode: "WH-02",
    defaultShelf: "USE-01",
    allowOverride: true,
    notes: "ตัดยอดเบิกใช้วัตถุดิบในการดำเนินงาน",
  },
  {
    id: "m-06",
    moduleKey: "stock_return",
    moduleNameTh: "งานคลัง / รับคืนจากการเบิกใช้ (Material Return)",
    moduleNameEn: "Stock Pickup Return",
    category: "inventory",
    defaultWarehouseCode: "WH-02",
    defaultShelf: "IN-02",
    allowOverride: true,
    notes: "รับคืนของเหลือจากการเบิกใช้เข้าคลังเดิม",
  },
  {
    id: "m-07",
    moduleKey: "stock_receive_finish",
    moduleNameTh: "งานคลัง / รับสินค้าสำเร็จรูปจากการผลิต (Finish Goods)",
    moduleNameEn: "Finish Goods Receive",
    category: "production",
    defaultWarehouseCode: "WH-01",
    defaultShelf: "FG-01",
    allowOverride: true,
    notes: "เพิ่มยอดสินค้าสำเร็จรูปเข้าคลังหลักพร้อมจำหน่าย",
  },
  {
    id: "m-08",
    moduleKey: "stock_adjust",
    moduleNameTh: "งานตรวจนับ / ปรับปรุงยอดสินค้า (Stock Adjustment)",
    moduleNameEn: "Stock Count Variance Adjustment",
    category: "inventory",
    defaultWarehouseCode: "WH-01",
    defaultShelf: "ADJ-01",
    allowOverride: true,
    notes: "ปรับยอดผลต่างสต็อกจากงานตรวจนับประจำงวด",
  },
];

export interface SystemDefaultWarehouseScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemDefaultWarehouseScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: SystemDefaultWarehouseScreenProps) {
  const isTh = language === "th";
  const [mappings, setMappings] = useState<TransactionWarehouseMapping[]>(DEFAULT_MAPPINGS);
  const [notification, setNotification] = useState<string | null>(null);

  const handleUpdate = <K extends keyof TransactionWarehouseMapping>(
    id: string,
    field: K,
    val: TransactionWarehouseMapping[K]
  ) => {
    setMappings((prev) =>
      prev.map((m) => (m.id === id ? { ...m, [field]: val } : m))
    );
  };

  const handleSave = () => {
    setNotification(
      isTh
        ? "บันทึกการกำหนดคลังสินค้าเริ่มต้นตามประเภทเอกสารเรียบร้อยแล้ว"
        : "Default warehouse mappings saved successfully."
    );
    setTimeout(() => setNotification(null), 3500);
  };

  const handleReset = () => {
    setMappings(DEFAULT_MAPPINGS);
    setNotification(isTh ? "คืนค่าการตั้งค่าเริ่มต้นเรียบร้อย" : "Reset mappings to defaults");
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
                {isTh ? "กำหนดคลังที่เก็บสินค้าเริ่มต้น" : "Default Warehouse Configuration"}
              </h1>
              <span className="px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-[11px] text-slate-500 dark:text-slate-400">
              {isTh
                ? "จัดการระบบ › กำหนดคลังสินค้าและช่องจัดเก็บเริ่มต้นแยกตามแต่ละประเภทเอกสารธุรกรรม"
                : "System Settings › Setup default destination/source warehouse and shelf per transaction type"}
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

      {/* Main Content List Table */}
      <div className="flex-1 overflow-auto p-4 sm:p-6 max-w-5xl mx-auto w-full">
        <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/50 flex items-center justify-between">
            <span className="font-semibold text-slate-800 dark:text-slate-100 text-sm">
              {isTh ? "ตารางกำหนดคลังสินค้าเริ่มต้นตามประเภทงาน" : "Default Warehouse Mapping Table"}
            </span>
            <span className="text-[11px] text-slate-500 font-mono">
              {mappings.length} {isTh ? "ประเภทธุรกรรม" : "transaction types"}
            </span>
          </div>

          <table className="w-full text-left border-collapse text-xs">
            <thead className="bg-slate-100/80 dark:bg-slate-800/80 text-slate-600 dark:text-slate-300 font-medium border-b border-slate-200 dark:border-slate-700">
              <tr>
                <th className="py-2.5 px-4 w-12 text-center">#</th>
                <th className="py-2.5 px-4 min-w-[260px]">{isTh ? "ประเภทธุรกรรม / โมดูล" : "Transaction Module"}</th>
                <th className="py-2.5 px-4 w-60">{isTh ? "คลังสินค้าเริ่มต้น" : "Default Warehouse"}</th>
                <th className="py-2.5 px-4 w-32">{isTh ? "ที่เก็บเริ่มต้น" : "Default Shelf"}</th>
                <th className="py-2.5 px-4 w-28 text-center">{isTh ? "ให้เปลี่ยนคลังได้" : "Allow Override"}</th>
                <th className="py-2.5 px-4">{isTh ? "หมายเหตุ / วัตถุประสงค์" : "Notes"}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
              {mappings.map((m, idx) => (
                <tr key={m.id} className="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors">
                  <td className="py-3 px-4 text-center font-mono text-slate-400">{idx + 1}</td>
                  <td className="py-3 px-4">
                    <p className="font-semibold text-slate-800 dark:text-slate-100">
                      {isTh ? m.moduleNameTh : m.moduleNameEn}
                    </p>
                    <span className="text-[10px] font-mono text-indigo-600 dark:text-indigo-400">
                      {m.moduleKey}
                    </span>
                  </td>
                  <td className="py-3 px-4">
                    <select
                      value={m.defaultWarehouseCode}
                      onChange={(e) => handleUpdate(m.id, "defaultWarehouseCode", e.target.value)}
                      className="w-full h-8 px-2.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-sans"
                    >
                      {AVAILABLE_WAREHOUSES.map((wh) => (
                        <option key={wh.code} value={wh.code}>
                          {wh.code} - {wh.name}
                        </option>
                      ))}
                    </select>
                  </td>
                  <td className="py-3 px-4">
                    <Input
                      value={m.defaultShelf}
                      onChange={(e) => handleUpdate(m.id, "defaultShelf", e.target.value.toUpperCase())}
                      className="h-8 text-xs font-mono uppercase"
                      placeholder="BIN-01"
                    />
                  </td>
                  <td className="py-3 px-4 text-center">
                    <input
                      type="checkbox"
                      checked={m.allowOverride}
                      onChange={(e) => handleUpdate(m.id, "allowOverride", e.target.checked)}
                      className="rounded text-indigo-600 focus:ring-indigo-500 w-4 h-4"
                      title={isTh ? "อนุญาตให้ผู้ใช้เลือกคลังอื่นในเอกสารได้" : "Allow user to select another warehouse"}
                    />
                  </td>
                  <td className="py-3 px-4 text-slate-500 dark:text-slate-400 text-[11px]">
                    {m.notes}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
