"use client";

import { useMemo, useRef, useState } from "react";
import {
  CheckCircle2,
  FileCheck,
  Package,
  Search,
  Star,
  Truck,
  Users,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export interface ProductSupplierRelation {
  id: string;
  supplierCode: string;
  supplierName: string;
  isPrimary: boolean;
  supplierPartNo: string;
  lastPurchasePrice: string;
  leadTimeDays: number;
  minOrderQty: number;
  contactPerson: string;
  phone: string;
}

export interface ProductSupplierMaster {
  id: string;
  productCode: string;
  productName: string;
  category: string;
  standardCost: string;
  unit: string;
  suppliers: ProductSupplierRelation[];
}

const INITIAL_PRODUCTS: ProductSupplierMaster[] = [
  {
    id: "prod-01",
    productCode: "MAT-CM-01",
    productName: "ปูนซีเมนต์ผสมสำเร็จ ปอร์ตแลนด์ประเภท 1 (ถุง 50 กก.)",
    category: "วัสดุโครงสร้าง",
    standardCost: "145.00",
    unit: "ถุง",
    suppliers: [
      {
        id: "rel-01",
        supplierCode: "AP-001",
        supplierName: "บริษัท ปูนซีเมนต์ไทย (ลำปาง) จำกัด",
        isPrimary: true,
        supplierPartNo: "SCG-PORTLAND-01",
        lastPurchasePrice: "142.50",
        leadTimeDays: 3,
        minOrderQty: 100,
        contactPerson: "นายประวิทย์ มิตรภาพ",
        phone: "054-221-889",
      },
      {
        id: "rel-02",
        supplierCode: "AP-004",
        supplierName: "บริษัท อิตาเลียนไทย บิลดิ้ง ซัพพลาย จำกัด",
        isPrimary: false,
        supplierPartNo: "ITB-CEM-P1",
        lastPurchasePrice: "148.00",
        leadTimeDays: 5,
        minOrderQty: 50,
        contactPerson: "นางสาวรัตนา วาณิชย์",
        phone: "02-716-1234",
      },
    ],
  },
  {
    id: "prod-02",
    productCode: "STL-DB12",
    productName: "เหล็กเส้นข้ออ้อย มอก. SD40 ขนาด DB12 ยาว 10 ม.",
    category: "เหล็กและโลหะภัณฑ์",
    standardCost: "195.00",
    unit: "เส้น",
    suppliers: [
      {
        id: "rel-03",
        supplierCode: "AP-002",
        supplierName: "บริษัท สยามสตีลอินดัสตรี จำกัด (มหาชน)",
        isPrimary: true,
        supplierPartNo: "SSI-SD40-DB12",
        lastPurchasePrice: "192.00",
        leadTimeDays: 7,
        minOrderQty: 200,
        contactPerson: "นายอานนท์ เหล็กกล้า",
        phone: "02-394-5511",
      },
    ],
  },
  {
    id: "prod-03",
    productCode: "FLR-GRN60",
    productName: "กระเบื้องปูพื้นแกรนิตโต้ นาโนไวท์ 60x60 ซม. (เกรด A)",
    category: "กระเบื้องและตกแต่ง",
    standardCost: "280.00",
    unit: "กล่อง",
    suppliers: [
      {
        id: "rel-04",
        supplierCode: "AP-004",
        supplierName: "บริษัท อิตาเลียนไทย บิลดิ้ง ซัพพลาย จำกัด",
        isPrimary: true,
        supplierPartNo: "CER-NANO-60W",
        lastPurchasePrice: "275.00",
        leadTimeDays: 4,
        minOrderQty: 30,
        contactPerson: "นางสาวรัตนา วาณิชย์",
        phone: "02-716-1234",
      },
    ],
  },
];

export interface ProductSupplierUpdateScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function ProductSupplierUpdateScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: ProductSupplierUpdateScreenProps) {
  const isTh = language === "th";
  const [products, setProducts] = useState<ProductSupplierMaster[]>(INITIAL_PRODUCTS);
  const [selectedId, setSelectedId] = useState<string>(INITIAL_PRODUCTS[0].id);
  const [categoryFilter, setCategoryFilter] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 38,
    min: 25,
    max: 60,
    storageKey: "bcai_split_product_supplier",
  });

  const selectedProduct = useMemo(() => {
    return products.find((p) => p.id === selectedId) || products[0];
  }, [products, selectedId]);

  const categories = useMemo(() => {
    const set = new Set<string>();
    products.forEach((p) => set.add(p.category));
    return ["all", ...Array.from(set)];
  }, [products]);

  const filteredProducts = useMemo(() => {
    return products.filter((p) => {
      const matchCat = categoryFilter === "all" || p.category === categoryFilter;
      const q = searchQuery.toLowerCase().trim();
      const matchSearch =
        !q ||
        p.productCode.toLowerCase().includes(q) ||
        p.productName.toLowerCase().includes(q) ||
        p.suppliers.some((s) => s.supplierName.toLowerCase().includes(q));
      return matchCat && matchSearch;
    });
  }, [products, categoryFilter, searchQuery]);

  // Set primary supplier
  const handleSetPrimary = (relId: string) => {
    setProducts((prev) =>
      prev.map((p) => {
        if (p.id === selectedProduct.id) {
          return {
            ...p,
            suppliers: p.suppliers.map((s) => ({
              ...s,
              isPrimary: s.id === relId,
            })),
          };
        }
        return p;
      })
    );
    setNotification(isTh ? "กำหนดผู้จำหน่ายหลักเรียบร้อย" : "Set primary supplier");
  };

  // Update supplier field
  const handleUpdateSupplierField = (
    relId: string,
    field: keyof ProductSupplierRelation,
    val: string | number
  ) => {
    setProducts((prev) =>
      prev.map((p) => {
        if (p.id === selectedProduct.id) {
          return {
            ...p,
            suppliers: p.suppliers.map((s) => (s.id === relId ? { ...s, [field]: val } : s)),
          };
        }
        return p;
      })
    );
  };

  const handleSaveAll = () => {
    setNotification(
      isTh
        ? `บันทึกข้อมูล Supplier ประจำสินค้า ${selectedProduct.productCode} เรียบร้อยแล้ว`
        : `Saved suppliers for ${selectedProduct.productCode}`
    );
  };

  return (
    <div
      ref={containerRef}
      className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 text-slate-800 dark:text-slate-100 overflow-hidden font-sans"
    >
      {/* Top Header Bar */}
      <div className="flex flex-wrap items-center justify-between px-4 py-2.5 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 gap-3">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-emerald-50 dark:bg-emerald-950/50 text-emerald-600 dark:text-emerald-400 border border-emerald-200/60 dark:border-emerald-800/40">
            <Truck className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold tracking-tight text-slate-900 dark:text-white">
                {isTh ? "ปรับปรุงข้อมูล Supplier ตามสินค้า (Product Supplier Mapping)" : "Product Supplier Mapping"}
              </h1>
              <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800">
                Champ Parity 50000+
              </span>
            </div>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {isTh
                ? "กำหนดผู้จำหน่ายหลัก/สำรอง ราคาซื้อล่าสุด ระยะเวลาจัดส่ง (Lead Time) และรหัสสินค้าของผู้ขาย"
                : "Manage primary/secondary suppliers, purchase cost, lead times & vendor part numbers"}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            size="sm"
            onClick={handleSaveAll}
            className="text-xs h-8 gap-1.5 bg-emerald-600 hover:bg-emerald-700 text-white font-medium shadow-sm"
          >
            <FileCheck className="w-3.5 h-3.5" />
            <span>{isTh ? "บันทึกการปรับปรุง" : "Save Changes"}</span>
          </Button>
        </div>
      </div>

      {/* Notification Banner */}
      {notification && (
        <div className="flex items-center justify-between px-4 py-1.5 bg-emerald-50 dark:bg-emerald-950/40 text-emerald-800 dark:text-emerald-200 text-xs border-b border-emerald-200 dark:border-emerald-900">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 text-emerald-600" />
            <span>{notification}</span>
          </div>
          <button onClick={() => setNotification(null)} className="text-slate-400 hover:text-slate-600">✕</button>
        </div>
      )}

      {/* Main Split Layout */}
      <div className="flex flex-1 overflow-hidden relative">
        {/* Left Pane: Product List & Category Filters */}
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
                placeholder={isTh ? "ค้นหารหัสสินค้า / ชื่อสินค้า / ผู้จำหน่าย..." : "Search code, name or supplier..."}
                className="pl-8 h-8 text-xs bg-white dark:bg-slate-950"
              />
            </div>

            {/* Category Pills */}
            <div className="flex flex-wrap gap-1 items-center">
              {categories.map((cat) => (
                <button
                  key={cat}
                  onClick={() => setCategoryFilter(cat)}
                  className={`text-[10px] px-2 py-0.5 rounded-full border transition-all ${
                    categoryFilter === cat
                      ? "bg-emerald-600 text-white border-emerald-600 font-semibold"
                      : "bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 border-slate-200 dark:border-slate-800 hover:bg-slate-100"
                  }`}
                >
                  {cat === "all" ? (isTh ? "ทุกหมวดหมู่" : "All") : cat}
                </button>
              ))}
            </div>
          </div>

          {/* List Items */}
          <div className="flex-1 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800/60 p-1.5 space-y-1">
            {filteredProducts.map((p) => {
              const isSelected = p.id === selectedProduct.id;
              const primarySupplier = p.suppliers.find((s) => s.isPrimary);
              return (
                <div
                  key={p.id}
                  onClick={() => setSelectedId(p.id)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") setSelectedId(p.id);
                  }}
                  tabIndex={0}
                  role="button"
                  className={`flex flex-col p-2.5 rounded-lg cursor-pointer transition-all border text-left ${
                    isSelected
                      ? "bg-emerald-50/90 dark:bg-emerald-950/40 border-emerald-300 dark:border-emerald-700 shadow-sm"
                      : "border-transparent hover:bg-slate-100/70 dark:hover:bg-slate-800/60"
                  }`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-mono text-xs font-bold text-slate-900 dark:text-slate-100">
                      {p.productCode}
                    </span>
                    <span className="text-[10px] px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                      {p.category}
                    </span>
                  </div>

                  <p className="text-xs font-semibold text-slate-800 dark:text-slate-200 mt-1 truncate">
                    {p.productName}
                  </p>

                  <div className="flex items-center justify-between mt-1 text-[11px] text-slate-500">
                    <span className="truncate max-w-[150px]">
                      {primarySupplier ? `★ ${primarySupplier.supplierName}` : isTh ? "ยังไม่มีผู้จำหน่าย" : "No supplier"}
                    </span>
                    <span className="font-mono font-bold text-emerald-600 dark:text-emerald-400">
                      ฿{Number(p.standardCost).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>

          {/* List Footer */}
          <div className="p-3 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 flex justify-between items-center">
            <span>{isTh ? `รวม ${filteredProducts.length} สินค้า` : `${filteredProducts.length} items`}</span>
            <span className="font-semibold text-slate-700 dark:text-slate-300">
              ผูกผู้จำหน่ายแล้ว {filteredProducts.reduce((acc, p) => acc + p.suppliers.length, 0)} คู่ค้า
            </span>
          </div>
        </div>

        {/* Resizable Splitter */}
        <ResizableSplitter
          isResizing={isResizing}
          onMouseDown={startResize}
          onKeyDown={adjustWithKeyboard}
        />

        {/* Right Pane: Supplier Mapping Table */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="flex flex-col h-full bg-slate-50 dark:bg-slate-950 overflow-y-auto p-4 space-y-4"
        >
          {/* Card 1: Product Header */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-3">
            <div className="flex items-center gap-2.5 pb-2 border-b border-slate-100 dark:border-slate-800">
              <Package className="w-5 h-5 text-emerald-600" />
              <div>
                <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <span>{selectedProduct.productCode}</span>
                  <span className="text-xs font-normal text-slate-500">({selectedProduct.category})</span>
                </h2>
                <p className="text-xs text-slate-600 dark:text-slate-300">{selectedProduct.productName}</p>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-3 gap-3 text-xs">
              <div>
                <span className="text-[11px] text-slate-400">{isTh ? "ราคาทุนมาตรฐาน:" : "Standard Cost:"}</span>
                <p className="font-mono font-bold text-slate-800 dark:text-slate-200">
                  ฿{Number(selectedProduct.standardCost).toLocaleString("th-TH", { minimumFractionDigits: 2 })} / {selectedProduct.unit}
                </p>
              </div>
              <div>
                <span className="text-[11px] text-slate-400">{isTh ? "จำนวนผู้จำหน่ายที่ผูก:" : "Mapped Suppliers:"}</span>
                <p className="font-bold text-emerald-600">
                  {selectedProduct.suppliers.length} ราย
                </p>
              </div>
            </div>
          </div>

          {/* Card 2: Mapped Suppliers Table */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-3">
            <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
              <div className="flex items-center gap-2">
                <Users className="w-4 h-4 text-emerald-600" />
                <h3 className="text-xs font-bold text-slate-900 dark:text-white">
                  {isTh ? "รายชื่อผู้จำหน่ายประจำสินค้านี้ (Suppliers for this product)" : "Mapped Suppliers"}
                </h3>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 bg-slate-50/70 dark:bg-slate-950/70">
                    <th className="py-2 px-3 font-semibold text-center">{isTh ? "ผู้ขายหลัก" : "Primary"}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "รหัส / ชื่อผู้จำหน่าย" : "Supplier"}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "รหัสสินค้าผู้ขาย (Part No.)" : "Vendor Part No."}</th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "ราคาซื้อล่าสุด (฿)" : "Last Price"}</th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "Lead Time (วัน)" : "Lead Time"}</th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "MOQ ขั้นต่ำ" : "MOQ"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono text-[11px]">
                  {selectedProduct.suppliers.map((s) => (
                    <tr key={s.id} className="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                      <td className="py-2 px-3 text-center">
                        <button
                          type="button"
                          onClick={() => handleSetPrimary(s.id)}
                          className={`p-1 rounded-md transition-all ${
                            s.isPrimary
                              ? "bg-amber-100 text-amber-600 dark:bg-amber-950/60"
                              : "text-slate-300 hover:text-amber-500"
                          }`}
                          title={isTh ? "คลิกเพื่อตั้งเป็นผู้ขายหลัก" : "Set as primary"}
                        >
                          <Star className="w-4 h-4 fill-current" />
                        </button>
                      </td>
                      <td className="py-2 px-3">
                        <span className="font-bold text-indigo-600 dark:text-indigo-400">{s.supplierCode}</span>
                        <p className="font-sans text-xs text-slate-800 dark:text-slate-200">{s.supplierName}</p>
                        <p className="text-[10px] text-slate-400 font-sans">{s.contactPerson} ({s.phone})</p>
                      </td>
                      <td className="py-2 px-3">
                        <Input
                          value={s.supplierPartNo}
                          onChange={(e) => handleUpdateSupplierField(s.id, "supplierPartNo", e.target.value)}
                          className="h-7 text-xs font-mono"
                        />
                      </td>
                      <td className="py-2 px-3 text-right">
                        <Input
                          type="number"
                          step="0.01"
                          value={s.lastPurchasePrice}
                          onChange={(e) => handleUpdateSupplierField(s.id, "lastPurchasePrice", e.target.value)}
                          className="h-7 text-xs font-mono text-right w-24 ml-auto"
                        />
                      </td>
                      <td className="py-2 px-3 text-right">
                        <Input
                          type="number"
                          value={s.leadTimeDays}
                          onChange={(e) => handleUpdateSupplierField(s.id, "leadTimeDays", Number(e.target.value))}
                          className="h-7 text-xs font-mono text-right w-16 ml-auto"
                        />
                      </td>
                      <td className="py-2 px-3 text-right">
                        <Input
                          type="number"
                          value={s.minOrderQty}
                          onChange={(e) => handleUpdateSupplierField(s.id, "minOrderQty", Number(e.target.value))}
                          className="h-7 text-xs font-mono text-right w-20 ml-auto"
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
