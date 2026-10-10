"use client";

import { useMemo, useRef, useState } from "react";
import {
  AlertTriangle,
  ArrowDownRight,
  ArrowUpRight,
  Barcode,
  Boxes,
  CheckCircle2,
  Clock,
  Coins,
  DollarSign,
  FileSpreadsheet,
  Filter,
  Layers,
  Package,
  RefreshCw,
  Search,
  Tag,
  Warehouse,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export type ProductCategoryType = "finished_goods" | "raw_material" | "spare_parts" | "service";
export type StockStatusType = "normal" | "low_stock" | "out_of_stock";

export interface WarehouseStockDetail {
  warehouseCode: string;
  warehouseName: string;
  shelfLocation: string;
  onHand: number;
  reserved: number;
  available: number;
  minStock: number;
}

export interface ProductMovementRecord {
  id: string;
  docNo: string;
  docType: string;
  date: string;
  warehouseCode: string;
  movementType: "in" | "out";
  qty: number;
  balanceAfter: number;
}

export interface ProductInquiryItem {
  id: string;
  code: string;
  barcode: string;
  nameTh: string;
  nameEn: string;
  category: ProductCategoryType;
  categoryName: string;
  unit: string;
  standardPrice: string; // minor unit formatted string
  wholesalePrice: string;
  contractorPrice: string;
  averageCost: string;
  lastPurchaseCost: string;
  totalOnHand: number;
  totalReserved: number;
  totalAvailable: number;
  status: StockStatusType;
  supplierName: string;
  warehouses: WarehouseStockDetail[];
  movements: ProductMovementRecord[];
}

const DEFAULT_PRODUCTS: ProductInquiryItem[] = [
  {
    id: "prod-001",
    code: "PROD-0001",
    barcode: "8851234567890",
    nameTh: "ปูนซีเมนต์ปอร์ตแลนด์ ตราเพชร ชนิดถุง 50 กก.",
    nameEn: "Portland Cement Diamond Brand 50kg",
    category: "finished_goods",
    categoryName: "สินค้าสำเร็จรูป / วัสดุก่อสร้าง",
    unit: "ถุง",
    standardPrice: "145.00",
    wholesalePrice: "135.00",
    contractorPrice: "138.00",
    averageCost: "118.50",
    lastPurchaseCost: "120.00",
    totalOnHand: 1450,
    totalReserved: 120,
    totalAvailable: 1330,
    status: "normal",
    supplierName: "บริษัท ปูนซีเมนต์ไทยอุตสาหกรรม จำกัด",
    warehouses: [
      { warehouseCode: "WH-01", warehouseName: "คลังสินค้าหลัก (Main Warehouse)", shelfLocation: "A-01-02", onHand: 1100, reserved: 100, available: 1000, minStock: 300 },
      { warehouseCode: "WH-02", warehouseName: "คลังสาขาบางนา (Bangna Branch)", shelfLocation: "B-03-01", onHand: 350, reserved: 20, available: 330, minStock: 100 },
    ],
    movements: [
      { id: "mov-1", docNo: "RC-202610-0011", docType: "รับสินค้าจากการซื้อ (PO)", date: "2026-10-05", warehouseCode: "WH-01", movementType: "in", qty: 500, balanceAfter: 1450 },
      { id: "mov-2", docNo: "IV-202610-0082", docType: "จ่ายสินค้าตามใบกำกับภาษี", date: "2026-10-04", warehouseCode: "WH-01", movementType: "out", qty: 150, balanceAfter: 950 },
      { id: "mov-3", docNo: "TR-202610-0005", docType: "โอนย้ายสินค้าระหว่างคลัง", date: "2026-10-02", warehouseCode: "WH-02", movementType: "in", qty: 100, balanceAfter: 1100 },
    ],
  },
  {
    id: "prod-002",
    code: "PROD-0002",
    barcode: "8851234567891",
    nameTh: "เหล็กเส้นข้ออ้อย มอก. SD40 ขนาด 16 มม. (ยาว 10 ม.)",
    nameEn: "Deformed Bar SD40 Dia 16mm x 10m",
    category: "finished_goods",
    categoryName: "สินค้าสำเร็จรูป / เหล็กก่อสร้าง",
    unit: "เส้น",
    standardPrice: "320.00",
    wholesalePrice: "295.00",
    contractorPrice: "305.00",
    averageCost: "272.00",
    lastPurchaseCost: "275.00",
    totalOnHand: 820,
    totalReserved: 150,
    totalAvailable: 670,
    status: "normal",
    supplierName: "บริษัท สยามสตีลคอร์ปอเรชั่น จำกัด (มหาชน)",
    warehouses: [
      { warehouseCode: "WH-01", warehouseName: "คลังสินค้าหลัก (Main Warehouse)", shelfLocation: "YARD-01", onHand: 820, reserved: 150, available: 670, minStock: 200 },
    ],
    movements: [
      { id: "mov-4", docNo: "IV-202610-0071", docType: "จ่ายสินค้าตามใบกำกับภาษี", date: "2026-10-05", warehouseCode: "WH-01", movementType: "out", qty: 80, balanceAfter: 820 },
      { id: "mov-5", docNo: "RC-202609-0120", docType: "รับสินค้าจากการซื้อ (PO)", date: "2026-09-28", warehouseCode: "WH-01", movementType: "in", qty: 400, balanceAfter: 900 },
    ],
  },
  {
    id: "prod-003",
    code: "PROD-0003",
    barcode: "8851234567892",
    nameTh: "กระเบื้องหลังคาลอนคู่ 50x120 ซม. สีเทาโมเดิร์น",
    nameEn: "Double Wave Roofing Tile 50x120cm Modern Grey",
    category: "finished_goods",
    categoryName: "สินค้าสำเร็จรูป / หลังคาและผนัง",
    unit: "แผ่น",
    standardPrice: "65.00",
    wholesalePrice: "56.00",
    contractorPrice: "59.00",
    averageCost: "48.50",
    lastPurchaseCost: "50.00",
    totalOnHand: 3200,
    totalReserved: 400,
    totalAvailable: 2800,
    status: "normal",
    supplierName: "บริษัท กระเบื้องไทยพาณิชย์ จำกัด",
    warehouses: [
      { warehouseCode: "WH-01", warehouseName: "คลังสินค้าหลัก (Main Warehouse)", shelfLocation: "R-02-04", onHand: 2400, reserved: 300, available: 2100, minStock: 500 },
      { warehouseCode: "WH-02", warehouseName: "คลังสาขาบางนา (Bangna Branch)", shelfLocation: "R-01-01", onHand: 800, reserved: 100, available: 700, minStock: 200 },
    ],
    movements: [
      { id: "mov-6", docNo: "IV-202610-0045", docType: "จ่ายสินค้าตามใบส่งของ", date: "2026-10-02", warehouseCode: "WH-01", movementType: "out", qty: 250, balanceAfter: 3200 },
    ],
  },
  {
    id: "prod-004",
    code: "PROD-0004",
    barcode: "8851234567893",
    nameTh: "สีน้ำอะคริลิกทาภายนอก ชนิดกึ่งเงา ขาวสะอาด ถัง 5 แกลลอน",
    nameEn: "Acrylic Exterior Paint Semi-Gloss White 5 Gal",
    category: "finished_goods",
    categoryName: "เคมีภัณฑ์และสีทาอาคาร",
    unit: "ถัง",
    standardPrice: "1,750.00",
    wholesalePrice: "1,550.00",
    contractorPrice: "1,600.00",
    averageCost: "1,380.00",
    lastPurchaseCost: "1,400.00",
    totalOnHand: 25,
    totalReserved: 18,
    totalAvailable: 7,
    status: "low_stock",
    supplierName: "บริษัท ทีโอเอ คัลเลอร์ เวิลด์ จำกัด",
    warehouses: [
      { warehouseCode: "WH-01", warehouseName: "คลังสินค้าหลัก (Main Warehouse)", shelfLocation: "P-04-02", onHand: 25, reserved: 18, available: 7, minStock: 30 },
    ],
    movements: [
      { id: "mov-7", docNo: "IV-202610-0012", docType: "จ่ายสินค้าโครงการ", date: "2026-10-01", warehouseCode: "WH-01", movementType: "out", qty: 20, balanceAfter: 25 },
    ],
  },
  {
    id: "prod-005",
    code: "PROD-0005",
    barcode: "8851234567894",
    nameTh: "ท่อ PVC ชนิดหนา ขนาด 2 นิ้ว ชั้น 8.5 (ยาว 4 ม.) มอก.",
    nameEn: "PVC Pipe Class 8.5 2-Inch x 4m",
    category: "finished_goods",
    categoryName: "อุปกรณ์ระบบประปาและสุขาภิบาล",
    unit: "ท่อน",
    standardPrice: "175.00",
    wholesalePrice: "148.00",
    contractorPrice: "155.00",
    averageCost: "128.00",
    lastPurchaseCost: "130.00",
    totalOnHand: 410,
    totalReserved: 30,
    totalAvailable: 380,
    status: "normal",
    supplierName: "บริษัท ท่อน้ำไทยภัณฑ์ จำกัด",
    warehouses: [
      { warehouseCode: "WH-01", warehouseName: "คลังสินค้าหลัก (Main Warehouse)", shelfLocation: "PIPE-01", onHand: 350, reserved: 20, available: 330, minStock: 80 },
      { warehouseCode: "WH-02", warehouseName: "คลังสาขาบางนา (Bangna Branch)", shelfLocation: "PIPE-02", onHand: 60, reserved: 10, available: 50, minStock: 20 },
    ],
    movements: [
      { id: "mov-8", docNo: "RC-202609-0099", docType: "รับสินค้าจากการซื้อ", date: "2026-09-22", warehouseCode: "WH-01", movementType: "in", qty: 150, balanceAfter: 410 },
    ],
  },
  {
    id: "prod-006",
    code: "PROD-0006",
    barcode: "8851234567895",
    nameTh: "ข้อต่อตรง PVC หนา ชนิดสวม ขนาด 2 นิ้ว ชั้น 13.5",
    nameEn: "PVC Coupling Socket 2-Inch Class 13.5",
    category: "finished_goods",
    categoryName: "อุปกรณ์ระบบประปาและสุขาภิบาล",
    unit: "ตัว",
    standardPrice: "18.00",
    wholesalePrice: "13.50",
    contractorPrice: "15.00",
    averageCost: "10.20",
    lastPurchaseCost: "10.50",
    totalOnHand: 0,
    totalReserved: 0,
    totalAvailable: 0,
    status: "out_of_stock",
    supplierName: "บริษัท ท่อน้ำไทยภัณฑ์ จำกัด",
    warehouses: [
      { warehouseCode: "WH-01", warehouseName: "คลังสินค้าหลัก (Main Warehouse)", shelfLocation: "BIN-22", onHand: 0, reserved: 0, available: 0, minStock: 50 },
    ],
    movements: [
      { id: "mov-9", docNo: "IV-202609-0215", docType: "จ่ายสินค้าตามใบส่งของ", date: "2026-09-25", warehouseCode: "WH-01", movementType: "out", qty: 45, balanceAfter: 0 },
    ],
  },
];

export interface SystemProductInquiryScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemProductInquiryScreen({
  language = "th",
  holdingcode = "rungrueng",
}: SystemProductInquiryScreenProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 35,
    min: 25,
    max: 50,
    storageKey: "bc_system_product_inquiry_split",
  });

  const [products, setProducts] = useState<ProductInquiryItem[]>(DEFAULT_PRODUCTS);
  const [selectedId, setSelectedId] = useState<string>(DEFAULT_PRODUCTS[0].id);
  const [search, setSearch] = useState("");
  const [filterCategory, setFilterCategory] = useState<string>("all");
  const [filterStockStatus, setFilterStockStatus] = useState<string>("all");
  const [activeTab, setActiveTab] = useState<"warehouses" | "pricing" | "movements" | "specs">("warehouses");

  const filteredProducts = useMemo(() => {
    const q = search.trim().toLowerCase();
    return products.filter((p) => {
      const matchSearch =
        !q ||
        p.code.toLowerCase().includes(q) ||
        p.barcode.includes(q) ||
        p.nameTh.toLowerCase().includes(q) ||
        p.nameEn.toLowerCase().includes(q) ||
        p.categoryName.toLowerCase().includes(q) ||
        p.supplierName.toLowerCase().includes(q);

      const matchCategory = filterCategory === "all" || p.category === filterCategory;
      const matchStatus = filterStockStatus === "all" || p.status === filterStockStatus;

      return matchSearch && matchCategory && matchStatus;
    });
  }, [products, search, filterCategory, filterStockStatus]);

  const selectedProduct = useMemo(
    () => products.find((p) => p.id === selectedId) ?? filteredProducts[0] ?? products[0],
    [products, selectedId, filteredProducts],
  );

  return (
    <div className="flex h-full flex-col overflow-hidden bg-background text-foreground" ref={containerRef}>
      {/* Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border bg-card px-4 py-2.5">
        <div className="flex items-center gap-2.5">
          <div className="grid h-9 w-9 place-items-center rounded-lg bg-primary/10 text-primary">
            <Package className="h-5 w-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold text-foreground">
                {language === "th" ? "แสดงข้อมูลสินค้า" : "Product Information Inquiry"}
              </h1>
              <span className="rounded-md bg-emerald-500/10 px-2 py-0.5 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-xs text-muted-foreground">
              {language === "th"
                ? "จัดการระบบ › แสดงข้อมูลสินค้า: ค้นหา รหัส บาร์โค้ด ยอดสต็อกคงเหลือแต่ละคลัง ราคาขาย และประวัติเคลื่อนไหว"
                : "System Management > Product Inquiry: Check barcodes, stock by warehouse, pricing tiers, and movement"}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setProducts(DEFAULT_PRODUCTS)}
            className="h-8 gap-1.5 text-xs"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span>{language === "th" ? "รีเฟรชข้อมูล" : "Refresh"}</span>
          </Button>
        </div>
      </div>

      {/* Main Content Pane */}
      <div className="relative flex flex-1 overflow-hidden">
        {/* Left Pane: Product Catalog List */}
        <div
          className="flex flex-col border-r border-border bg-card/40"
          style={{ width: `${splitPercent}%` }}
        >
          {/* Search & Filters */}
          <div className="border-b border-border p-2.5 space-y-2">
            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
              <Input
                placeholder={language === "th" ? "ค้นหารหัส บาร์โค้ด ชื่อสินค้า..." : "Search code, barcode, name..."}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="h-8 pl-8 text-xs"
              />
            </div>
            <div className="flex items-center gap-1.5">
              <select
                value={filterCategory}
                onChange={(e) => setFilterCategory(e.target.value)}
                className="h-7 flex-1 rounded-md border border-input bg-background px-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
              >
                <option value="all">{language === "th" ? "ทุกหมวดสินค้า" : "All Categories"}</option>
                <option value="finished_goods">{language === "th" ? "สินค้าสำเร็จรูป" : "Finished Goods"}</option>
                <option value="raw_material">{language === "th" ? "วัตถุดิบ" : "Raw Materials"}</option>
                <option value="spare_parts">{language === "th" ? "อะไหล่/อุปกรณ์" : "Spare Parts"}</option>
              </select>

              <select
                value={filterStockStatus}
                onChange={(e) => setFilterStockStatus(e.target.value)}
                className="h-7 w-28 rounded-md border border-input bg-background px-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
              >
                <option value="all">{language === "th" ? "ทุกสถานะสต็อก" : "All Stock"}</option>
                <option value="normal">{language === "th" ? "สต็อกปกติ" : "In Stock"}</option>
                <option value="low_stock">{language === "th" ? "สต็อกใกล้หมด" : "Low Stock"}</option>
                <option value="out_of_stock">{language === "th" ? "สินค้าหมด" : "Out of Stock"}</option>
              </select>
            </div>
          </div>

          {/* List count summary */}
          <div className="flex items-center justify-between border-b border-border/60 bg-muted/30 px-3 py-1.5 text-[0.72rem] text-muted-foreground">
            <span>
              {language === "th" ? `พบสินค้า ${filteredProducts.length} รายการ` : `Found ${filteredProducts.length} items`}
            </span>
            <span>{holdingcode}</span>
          </div>

          {/* Product Items */}
          <div className="flex-1 overflow-y-auto divide-y divide-border/60">
            {filteredProducts.length === 0 ? (
              <div className="p-8 text-center text-xs text-muted-foreground">
                {language === "th" ? "ไม่พบข้อมูลสินค้าที่ตรงกับเงื่อนไข" : "No products found"}
              </div>
            ) : (
              filteredProducts.map((p) => {
                const isSelected = p.id === selectedProduct?.id;
                return (
                  <button
                    key={p.id}
                    type="button"
                    onClick={() => setSelectedId(p.id)}
                    className={`w-full p-2.5 text-left transition-colors ${
                      isSelected
                        ? "bg-primary/10 border-l-4 border-primary"
                        : "hover:bg-muted/40"
                    }`}
                  >
                    <div className="flex items-start justify-between gap-1.5">
                      <span className="font-semibold text-xs text-primary">{p.code}</span>
                      {p.status === "normal" && (
                        <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.2 text-[0.68rem] font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                          มีสินค้า
                        </span>
                      )}
                      {p.status === "low_stock" && (
                        <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.2 text-[0.68rem] font-medium bg-amber-500/10 text-amber-600 dark:text-amber-400">
                          ใกล้หมด
                        </span>
                      )}
                      {p.status === "out_of_stock" && (
                        <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.2 text-[0.68rem] font-medium bg-rose-500/10 text-rose-600 dark:text-rose-400">
                          หมด
                        </span>
                      )}
                    </div>
                    <p className="mt-0.5 line-clamp-1 text-xs font-medium text-foreground">
                      {p.nameTh}
                    </p>
                    <div className="mt-1 flex items-center justify-between text-[0.7rem] text-muted-foreground">
                      <span className="font-mono">{p.barcode}</span>
                      <span className="font-medium text-foreground">
                        คงเหลือ: {p.totalOnHand.toLocaleString()} {p.unit}
                      </span>
                    </div>
                  </button>
                );
              })
            )}
          </div>
        </div>

        {/* Resizable Splitter */}
        <ResizableSplitter
          onMouseDown={startResize}
          onKeyDown={adjustWithKeyboard}
          isResizing={isResizing}
        />

        {/* Right Pane: Detailed Product Info */}
        <div
          className="flex-1 overflow-y-auto bg-background/60 p-4"
          style={{ width: `${100 - splitPercent}%` }}
        >
          {selectedProduct ? (
            <div className="space-y-4">
              {/* Product Header Card */}
              <Card className="border-border shadow-xs">
                <CardHeader className="p-4 pb-3">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="rounded-md bg-primary/15 px-2 py-0.5 text-xs font-bold text-primary">
                          {selectedProduct.code}
                        </span>
                        <h2 className="text-base font-bold text-foreground">
                          {selectedProduct.nameTh}
                        </h2>
                      </div>
                      <p className="mt-1 text-xs text-muted-foreground">
                        {selectedProduct.nameEn} · บาร์โค้ด: <span className="font-mono">{selectedProduct.barcode}</span> · {selectedProduct.categoryName}
                      </p>
                    </div>

                    <div className="text-right">
                      <div className="text-[0.7rem] text-muted-foreground">ราคาขายมาตรฐาน (Standard Price)</div>
                      <div className="text-lg font-bold text-foreground">
                        {selectedProduct.standardPrice} <span className="text-xs font-normal">บาท/{selectedProduct.unit}</span>
                      </div>
                    </div>
                  </div>
                </CardHeader>

                {/* KPI Ribbon */}
                <CardContent className="grid grid-cols-2 gap-2 border-t border-border/80 bg-muted/20 p-3 sm:grid-cols-4">
                  <div className="rounded-lg border border-border/60 bg-card p-2 text-center">
                    <span className="text-[0.68rem] text-muted-foreground">ยอดคงเหลือจริง (On Hand)</span>
                    <p className="text-sm font-bold text-foreground">
                      {selectedProduct.totalOnHand.toLocaleString()} {selectedProduct.unit}
                    </p>
                  </div>
                  <div className="rounded-lg border border-border/60 bg-card p-2 text-center">
                    <span className="text-[0.68rem] text-muted-foreground">ยอดถูกจอง (Reserved)</span>
                    <p className="text-sm font-bold text-amber-600 dark:text-amber-400">
                      {selectedProduct.totalReserved.toLocaleString()} {selectedProduct.unit}
                    </p>
                  </div>
                  <div className="rounded-lg border border-border/60 bg-card p-2 text-center">
                    <span className="text-[0.68rem] text-muted-foreground">ยอดพร้อมขาย (Available)</span>
                    <p className="text-sm font-bold text-emerald-600 dark:text-emerald-400">
                      {selectedProduct.totalAvailable.toLocaleString()} {selectedProduct.unit}
                    </p>
                  </div>
                  <div className="rounded-lg border border-border/60 bg-card p-2 text-center">
                    <span className="text-[0.68rem] text-muted-foreground">ต้นทุนเฉลี่ย (Avg Cost)</span>
                    <p className="text-sm font-bold text-primary">
                      {selectedProduct.averageCost} ฿
                    </p>
                  </div>
                </CardContent>
              </Card>

              {/* Sub-tabs Navigation */}
              <div className="flex border-b border-border gap-2">
                <button
                  type="button"
                  onClick={() => setActiveTab("warehouses")}
                  className={`px-3 py-2 text-xs font-semibold border-b-2 transition-colors ${
                    activeTab === "warehouses"
                      ? "border-primary text-primary"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {language === "th" ? "ยอดสต็อกแยกตามคลัง" : "Stock by Warehouse"}
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("pricing")}
                  className={`px-3 py-2 text-xs font-semibold border-b-2 transition-colors ${
                    activeTab === "pricing"
                      ? "border-primary text-primary"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {language === "th" ? "โครงสร้างราคาและต้นทุน" : "Pricing & Cost"}
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("movements")}
                  className={`px-3 py-2 text-xs font-semibold border-b-2 transition-colors ${
                    activeTab === "movements"
                      ? "border-primary text-primary"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {language === "th" ? "ประวัติเคลื่อนไหวล่าสุด" : "Stock Movement"}
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("specs")}
                  className={`px-3 py-2 text-xs font-semibold border-b-2 transition-colors ${
                    activeTab === "specs"
                      ? "border-primary text-primary"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {language === "th" ? "ข้อมูลจัดซื้อและผู้จำหน่าย" : "Supplier & Specs"}
                </button>
              </div>

              {/* Tab 1: Stock by Warehouse */}
              {activeTab === "warehouses" && (
                <Card className="border-border">
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-sm font-bold">ยอดสินค้าคงคลังแยกตามที่เก็บและคลังสินค้า</CardTitle>
                    <CardDescription className="text-xs">
                      แสดงตำแหน่งที่จัดเก็บ ยอดคงเหลือจริง และยอดที่พร้อมจำหน่ายในแต่ละคลัง
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="p-4 pt-1">
                    <div className="overflow-x-auto rounded-lg border border-border">
                      <table className="w-full text-left text-xs">
                        <thead className="bg-muted/50 text-[0.7rem] uppercase text-muted-foreground font-semibold">
                          <tr>
                            <th className="px-3 py-2">รหัสคลัง</th>
                            <th className="px-3 py-2">ชื่อคลังสินค้า</th>
                            <th className="px-3 py-2">ตำแหน่งที่เก็บ (Shelf)</th>
                            <th className="px-3 py-2 text-right">ยอดคงเหลือจริง</th>
                            <th className="px-3 py-2 text-right">ยอดติดจอง</th>
                            <th className="px-3 py-2 text-right">ยอดพร้อมขาย</th>
                            <th className="px-3 py-2 text-right">จุดสั่งซื้อ (Min)</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                          {selectedProduct.warehouses.map((wh) => (
                            <tr key={wh.warehouseCode} className="hover:bg-muted/30">
                              <td className="px-3 py-2.5 font-mono font-medium text-primary">{wh.warehouseCode}</td>
                              <td className="px-3 py-2.5 font-medium">{wh.warehouseName}</td>
                              <td className="px-3 py-2.5 font-mono text-muted-foreground">{wh.shelfLocation}</td>
                              <td className="px-3 py-2.5 text-right font-medium">
                                {wh.onHand.toLocaleString()} {selectedProduct.unit}
                              </td>
                              <td className="px-3 py-2.5 text-right font-medium text-amber-600 dark:text-amber-400">
                                {wh.reserved.toLocaleString()} {selectedProduct.unit}
                              </td>
                              <td className="px-3 py-2.5 text-right font-bold text-emerald-600 dark:text-emerald-400">
                                {wh.available.toLocaleString()} {selectedProduct.unit}
                              </td>
                              <td className="px-3 py-2.5 text-right text-muted-foreground">
                                {wh.minStock.toLocaleString()} {selectedProduct.unit}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Tab 2: Pricing & Cost */}
              {activeTab === "pricing" && (
                <Card className="border-border">
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-sm font-bold">โครงสร้างราคาขายและต้นทุนสินค้า</CardTitle>
                    <CardDescription className="text-xs">
                      ราคากลางที่ระบบนำไปใช้ในการเปิดใบเสนอราคาและใบกำกับสินค้า
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="p-4 pt-1 space-y-3">
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                      <div className="rounded-lg border border-border p-3 space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">ราคาขายปลีกมาตรฐาน (Retail)</span>
                        <p className="text-sm font-bold text-foreground">{selectedProduct.standardPrice} บาท/{selectedProduct.unit}</p>
                      </div>
                      <div className="rounded-lg border border-border p-3 space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">ราคาขายส่ง (Wholesale)</span>
                        <p className="text-sm font-bold text-primary">{selectedProduct.wholesalePrice} บาท/{selectedProduct.unit}</p>
                      </div>
                      <div className="rounded-lg border border-border p-3 space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">ราคาผู้รับเหมา/โครงการ (Contractor)</span>
                        <p className="text-sm font-bold text-foreground">{selectedProduct.contractorPrice} บาท/{selectedProduct.unit}</p>
                      </div>
                      <div className="rounded-lg border border-border p-3 space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">ต้นทุนถัวเฉลี่ยเคลื่อนที่ (Avg Cost)</span>
                        <p className="text-sm font-bold text-foreground">{selectedProduct.averageCost} บาท</p>
                      </div>
                      <div className="rounded-lg border border-border p-3 space-y-1 sm:col-span-2">
                        <span className="text-[0.72rem] text-muted-foreground">ต้นทุนซื้อล่าสุด (Last Purchase Cost)</span>
                        <p className="text-sm font-bold text-emerald-600 dark:text-emerald-400">
                          {selectedProduct.lastPurchaseCost} บาท ({selectedProduct.supplierName})
                        </p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Tab 3: Stock Movement */}
              {activeTab === "movements" && (
                <Card className="border-border">
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-sm font-bold">ประวัติการเคลื่อนไหวสต็อกสินค้า (Cardex Summary)</CardTitle>
                    <CardDescription className="text-xs">
                      รายการรับเข้า-จ่ายออก และโอนย้ายสินค้าล่าสุด
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="p-4 pt-1">
                    <div className="overflow-x-auto rounded-lg border border-border">
                      <table className="w-full text-left text-xs">
                        <thead className="bg-muted/50 text-[0.7rem] uppercase text-muted-foreground font-semibold">
                          <tr>
                            <th className="px-3 py-2">เลขที่เอกสาร</th>
                            <th className="px-3 py-2">ประเภทรายการ</th>
                            <th className="px-3 py-2">วันที่</th>
                            <th className="px-3 py-2">คลัง</th>
                            <th className="px-3 py-2 text-right">จำนวน</th>
                            <th className="px-3 py-2 text-right">ยอดคงเหลือหลังทำรายการ</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                          {selectedProduct.movements.map((m) => (
                            <tr key={m.id} className="hover:bg-muted/30">
                              <td className="px-3 py-2 font-mono font-medium text-primary">{m.docNo}</td>
                              <td className="px-3 py-2 text-muted-foreground">{m.docType}</td>
                              <td className="px-3 py-2">{m.date}</td>
                              <td className="px-3 py-2 font-mono">{m.warehouseCode}</td>
                              <td className="px-3 py-2 text-right font-medium">
                                {m.movementType === "in" ? (
                                  <span className="text-emerald-600 dark:text-emerald-400 font-semibold">
                                    +{m.qty.toLocaleString()}
                                  </span>
                                ) : (
                                  <span className="text-rose-600 dark:text-rose-400 font-semibold">
                                    -{m.qty.toLocaleString()}
                                  </span>
                                )}
                              </td>
                              <td className="px-3 py-2 text-right font-bold text-foreground">
                                {m.balanceAfter.toLocaleString()} {selectedProduct.unit}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Tab 4: Specs & Supplier */}
              {activeTab === "specs" && (
                <Card className="border-border">
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-sm font-bold">ข้อมูลจำเพาะและผู้จำหน่ายหลัก</CardTitle>
                    <CardDescription className="text-xs">
                      ข้อมูลซัพพลายเออร์สำหรับการสั่งซื้อซ้ำ
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="p-4 pt-1 space-y-3">
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                      <div className="space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">ผู้จำหน่ายหลัก (Primary Supplier)</span>
                        <p className="text-xs font-semibold text-foreground">{selectedProduct.supplierName}</p>
                      </div>
                      <div className="space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">หน่วยนับหลัก (Primary Unit)</span>
                        <p className="text-xs font-semibold text-foreground">{selectedProduct.unit}</p>
                      </div>
                      <div className="space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">รหัสหมวดสินค้า</span>
                        <p className="text-xs font-medium text-foreground">{selectedProduct.categoryName}</p>
                      </div>
                      <div className="space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">รหัสบาร์โค้ดสากล (EAN-13)</span>
                        <p className="text-xs font-mono font-medium text-foreground">{selectedProduct.barcode}</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}
            </div>
          ) : (
            <div className="grid h-full place-items-center text-xs text-muted-foreground">
              กรุณาเลือกสินค้าจากรายการทางด้านซ้าย
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
