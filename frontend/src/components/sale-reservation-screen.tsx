"use client";

import { useMemo, useRef, useState } from "react";
import {
  CheckCircle2,
  FileCheck,
  Package,
  Plus,
  Printer,
  Search,
  ShoppingCart,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { thaiBahtText } from "@/lib/thai-baht-text";
import { type LanguageCode } from "@/lib/i18n";

export interface ReservationLineItem {
  id: string;
  itemCode: string;
  itemName: string;
  unit: string;
  qtyReserved: number;
  availableStock: number;
  unitPrice: string;
  discount: string;
  lineTotal: string;
}

export interface SalesReservationDoc {
  id: string;
  docNo: string;
  docDate: string;
  dueDate: string;
  customerCode: string;
  customerName: string;
  salesperson: string;
  department: string;
  depositRequired: string;
  depositReceived: string;
  status: "draft" | "reserved" | "converted_to_so" | "cancelled";
  items: ReservationLineItem[];
  subtotal: string;
  vatAmount: string;
  netTotal: string;
  remark: string;
}

const INITIAL_RESERVATIONS: SalesReservationDoc[] = [
  {
    id: "sr-01",
    docNo: "SR6910-0001",
    docDate: "2026-10-05",
    dueDate: "2026-10-25",
    customerCode: "AR-001",
    customerName: "บริษัท สยามคอนกรีตมิกซ์ จำกัด",
    salesperson: "นายสมศักดิ์ ขยันขาย",
    department: "ฝ่ายขายโครงการ",
    depositRequired: "20000.00",
    depositReceived: "20000.00",
    status: "reserved",
    subtotal: "128400.00",
    vatAmount: "8988.00",
    netTotal: "137388.00",
    remark: "จองปูนซีเมนต์และหินผสมสำหรับงานเทพื้นอาคาร C",
    items: [
      {
        id: "item-01",
        itemCode: "MAT-CM-01",
        itemName: "ปูนซีเมนต์ผสมสำเร็จ ปอร์ตแลนด์ประเภท 1 (ถุง 50 กก.)",
        unit: "ถุง",
        qtyReserved: 300,
        availableStock: 850,
        unitPrice: "160.00",
        discount: "0.00",
        lineTotal: "48000.00",
      },
      {
        id: "item-02",
        itemCode: "MAT-ST-02",
        itemName: "หินก่อสร้าง เบอร์ 1 ผสมคอนกรีต (คิว)",
        unit: "คิว",
        qtyReserved: 120,
        availableStock: 250,
        unitPrice: "670.00",
        discount: "0.00",
        lineTotal: "80400.00",
      },
    ],
  },
  {
    id: "sr-02",
    docNo: "SR6910-0002",
    docDate: "2026-10-08",
    dueDate: "2026-10-31",
    customerCode: "AR-002",
    customerName: "ห้างหุ้นส่วนจำกัด นครหลวงวิศวกรรม",
    salesperson: "นายสมศักดิ์ ขยันขาย",
    department: "ฝ่ายขายโครงการ",
    depositRequired: "50000.00",
    depositReceived: "0.00",
    status: "draft",
    subtotal: "214000.00",
    vatAmount: "14980.00",
    netTotal: "228980.00",
    remark: "จองเหล็กเส้นข้ออ้อย DB12 และ DB16 ล็อตใหญ่ รอวางเงินมัดจำ",
    items: [
      {
        id: "item-03",
        itemCode: "STL-DB12",
        itemName: "เหล็กเส้นข้ออ้อย มอก. SD40 ขนาด DB12 ยาว 10 ม.",
        unit: "เส้น",
        qtyReserved: 1000,
        availableStock: 1400,
        unitPrice: "214.00",
        discount: "0.00",
        lineTotal: "214000.00",
      },
    ],
  },
  {
    id: "sr-03",
    docNo: "SR6910-0003",
    docDate: "2026-09-28",
    dueDate: "2026-10-10",
    customerCode: "AR-004",
    customerName: "บริษัท สหรุ่งโรจน์ ดีเวลลอปเมนท์ จำกัด",
    salesperson: "นางสาวมณีรัตน์ วาจางาม",
    department: "ฝ่ายขายสาขา",
    depositRequired: "10000.00",
    depositReceived: "10000.00",
    status: "converted_to_so",
    subtotal: "56710.00",
    vatAmount: "3969.70",
    netTotal: "60679.70",
    remark: "จองกระเบื้องปูพื้นแกรนิตโต้ แปลงเป็นใบสั่งขาย SO6910-0045 เรียบร้อย",
    items: [
      {
        id: "item-04",
        itemCode: "FLR-GRN60",
        itemName: "กระเบื้องปูพื้นแกรนิตโต้ นาโนไวท์ 60x60 ซม. (เกรด A)",
        unit: "กล่อง",
        qtyReserved: 180,
        availableStock: 320,
        unitPrice: "315.00",
        discount: "0.00",
        lineTotal: "56710.00",
      },
    ],
  },
];

export interface SaleReservationScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SaleReservationScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: SaleReservationScreenProps) {
  const isTh = language === "th";
  const [docs, setDocs] = useState<SalesReservationDoc[]>(INITIAL_RESERVATIONS);
  const [selectedId, setSelectedId] = useState<string>(INITIAL_RESERVATIONS[0].id);
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 38,
    min: 25,
    max: 60,
    storageKey: "bcai_split_sale_reservation",
  });

  const selectedDoc = useMemo(() => {
    return docs.find((d) => d.id === selectedId) || docs[0];
  }, [docs, selectedId]);

  // Form states
  const [editDocDate, setEditDocDate] = useState(selectedDoc.docDate);
  const [editDueDate, setEditDueDate] = useState(selectedDoc.dueDate);
  const [editDepositReq, setEditDepositReq] = useState(selectedDoc.depositRequired);
  const [editRemark, setEditRemark] = useState(selectedDoc.remark);

  const handleSelectDoc = (doc: SalesReservationDoc) => {
    setSelectedId(doc.id);
    setEditDocDate(doc.docDate);
    setEditDueDate(doc.dueDate);
    setEditDepositReq(doc.depositRequired);
    setEditRemark(doc.remark);
    setNotification(null);
  };

  const filteredDocs = useMemo(() => {
    return docs.filter((d) => {
      const matchStatus = statusFilter === "all" || d.status === statusFilter;
      const q = searchQuery.toLowerCase().trim();
      const matchSearch =
        !q ||
        d.docNo.toLowerCase().includes(q) ||
        d.customerName.toLowerCase().includes(q) ||
        d.customerCode.toLowerCase().includes(q) ||
        d.salesperson.toLowerCase().includes(q);
      return matchStatus && matchSearch;
    });
  }, [docs, statusFilter, searchQuery]);

  const thaiText = useMemo(() => {
    return thaiBahtText(selectedDoc.netTotal);
  }, [selectedDoc.netTotal]);

  const handleSaveDoc = () => {
    setDocs((prev) =>
      prev.map((d) => {
        if (d.id === selectedDoc.id) {
          return {
            ...d,
            docDate: editDocDate,
            dueDate: editDueDate,
            depositRequired: editDepositReq,
            remark: editRemark,
          };
        }
        return d;
      })
    );
    setNotification(isTh ? "บันทึกใบสั่งจองสินค้าเรียบร้อย" : "Sales reservation saved");
  };

  const handleCreateNew = () => {
    const nextSeq = String(docs.length + 1).padStart(4, "0");
    const newDoc: SalesReservationDoc = {
      id: `sr-${Date.now()}`,
      docNo: `SR6910-${nextSeq}`,
      docDate: "2026-10-10",
      dueDate: "2026-10-31",
      customerCode: "AR-001",
      customerName: "บริษัท สยามคอนกรีตมิกซ์ จำกัด",
      salesperson: "นายสมศักดิ์ ขยันขาย",
      department: "ฝ่ายขายโครงการ",
      depositRequired: "15000.00",
      depositReceived: "0.00",
      status: "draft",
      subtotal: "45000.00",
      vatAmount: "3150.00",
      netTotal: "48150.00",
      remark: "ใบสั่งจองสินค้าใหม่",
      items: [
        {
          id: `item-${Date.now()}`,
          itemCode: "MAT-CM-01",
          itemName: "ปูนซีเมนต์ผสมสำเร็จ ปอร์ตแลนด์ประเภท 1 (ถุง 50 กก.)",
          unit: "ถุง",
          qtyReserved: 280,
          availableStock: 850,
          unitPrice: "160.00",
          discount: "0.00",
          lineTotal: "44800.00",
        },
      ],
    };

    setDocs([newDoc, ...docs]);
    handleSelectDoc(newDoc);
    setNotification(isTh ? `สร้างใบสั่งจองใหม่ ${newDoc.docNo}` : `Created reservation ${newDoc.docNo}`);
  };

  const handleConfirmReservation = () => {
    setDocs((prev) =>
      prev.map((d) => {
        if (d.id === selectedDoc.id) {
          return { ...d, status: "reserved" };
        }
        return d;
      })
    );
    setNotification(isTh ? `ยืนยันการจองสินค้า ${selectedDoc.docNo} (กันสต็อกสินค้าเรียบร้อย)` : `Confirmed reservation ${selectedDoc.docNo}`);
  };

  return (
    <div
      ref={containerRef}
      className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 text-slate-800 dark:text-slate-100 overflow-hidden font-sans"
    >
      {/* Top Header Bar */}
      <div className="flex flex-wrap items-center justify-between px-4 py-2.5 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 gap-3">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 border border-indigo-200/60 dark:border-indigo-800/40">
            <ShoppingCart className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold tracking-tight text-slate-900 dark:text-white">
                {isTh ? "ใบสั่งจองสินค้า (Sales Reservation)" : "Sales Reservation"}
              </h1>
              <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800">
                Champ Parity 30005
              </span>
            </div>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {isTh
                ? "บันทึกจองสินค้าล่วงหน้า กันยอดสต็อกคงเหลือ และติดตามเงินมัดจำก่อนแปลงเป็นใบสั่งขาย"
                : "Reserve inventory stock in advance, track deposits & convert to sales order"}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            size="sm"
            onClick={handleCreateNew}
            className="text-xs h-8 gap-1.5 bg-indigo-600 hover:bg-indigo-700 text-white font-medium shadow-sm"
          >
            <Plus className="w-3.5 h-3.5" />
            <span>{isTh ? "เพิ่มใบสั่งจอง" : "New Reservation"}</span>
          </Button>
          <Button
            size="sm"
            onClick={handleSaveDoc}
            className="text-xs h-8 gap-1.5 bg-slate-800 hover:bg-slate-900 text-white dark:bg-slate-200 dark:text-slate-900"
          >
            <FileCheck className="w-3.5 h-3.5" />
            <span>{isTh ? "บันทึก" : "Save"}</span>
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => window.print()}
            className="text-xs h-8 gap-1.5"
          >
            <Printer className="w-3.5 h-3.5" />
            <span>{isTh ? "พิมพ์ใบจอง" : "Print"}</span>
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
        {/* Left Pane: Filter & Reservation List */}
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
                placeholder={isTh ? "ค้นหาเลขที่จอง / ลูกค้า / พนักงาน..." : "Search doc, customer or sales..."}
                className="pl-8 h-8 text-xs bg-white dark:bg-slate-950"
              />
            </div>

            {/* Status Pills */}
            <div className="flex flex-wrap gap-1 items-center">
              {[
                { id: "all", th: "ทั้งหมด", en: "All" },
                { id: "draft", th: "ร่าง", en: "Draft" },
                { id: "reserved", th: "จองแล้ว", en: "Reserved" },
                { id: "converted_to_so", th: "แปลงเป็น SO", en: "To SO" },
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
            {filteredDocs.map((doc) => {
              const isSelected = doc.id === selectedDoc.id;
              return (
                <div
                  key={doc.id}
                  onClick={() => handleSelectDoc(doc)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") handleSelectDoc(doc);
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
                      {doc.docNo}
                    </span>
                    <span
                      className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${
                        doc.status === "reserved"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-300/50"
                          : doc.status === "converted_to_so"
                          ? "bg-blue-100 text-blue-800 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-300/50"
                          : "bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400"
                      }`}
                    >
                      {doc.status === "reserved"
                        ? isTh ? "จองสินค้าแล้ว" : "Reserved"
                        : doc.status === "converted_to_so"
                        ? isTh ? "เป็นใบสั่งขายแล้ว" : "SO Created"
                        : isTh ? "ร่าง" : "Draft"}
                    </span>
                  </div>

                  <p className="text-xs font-semibold text-slate-800 dark:text-slate-200 mt-1 truncate">
                    {doc.customerName}
                  </p>

                  <div className="flex items-center justify-between mt-1 text-[11px] text-slate-500">
                    <span>{isTh ? `นัดส่ง: ${doc.dueDate}` : `Due: ${doc.dueDate}`}</span>
                    <span className="font-mono font-bold text-indigo-600 dark:text-indigo-400">
                      ฿{Number(doc.netTotal).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>

          {/* List Footer Total */}
          <div className="p-3 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 flex justify-between items-center">
            <span>{isTh ? `รวม ${filteredDocs.length} ฉบับ` : `${filteredDocs.length} orders`}</span>
            <span className="font-mono font-bold text-slate-800 dark:text-slate-200">
              ฿{filteredDocs
                .reduce((acc, d) => acc + Number(d.netTotal), 0)
                .toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </span>
          </div>
        </div>

        {/* Resizable Splitter */}
        <ResizableSplitter
          isResizing={isResizing}
          onMouseDown={startResize}
          onKeyDown={adjustWithKeyboard}
        />

        {/* Right Pane: Reservation Detail Form & Line Items */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="flex flex-col h-full bg-slate-50 dark:bg-slate-950 overflow-y-auto p-4 space-y-4"
        >
          {/* Card 1: Header Information */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-4">
            <div className="flex flex-wrap items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800 gap-2">
              <div className="flex items-center gap-2">
                <ShoppingCart className="w-5 h-5 text-indigo-600" />
                <div>
                  <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                    <span>{selectedDoc.docNo}</span>
                    <span
                      className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${
                        selectedDoc.status === "reserved"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300"
                          : "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300"
                      }`}
                    >
                      {selectedDoc.status === "reserved" ? "จองสต็อกแล้ว" : "แบบร่าง"}
                    </span>
                  </h2>
                  <p className="text-xs text-slate-500">{selectedDoc.customerName} ({selectedDoc.customerCode})</p>
                </div>
              </div>

              {selectedDoc.status === "draft" && (
                <Button
                  size="sm"
                  onClick={handleConfirmReservation}
                  className="text-xs h-7 gap-1 bg-emerald-600 hover:bg-emerald-700 text-white"
                >
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  <span>{isTh ? "ยืนยันการจองสต็อก" : "Confirm Reservation"}</span>
                </Button>
              )}
            </div>

            {/* Form Fields */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "วันที่ใบจอง:" : "Doc Date:"}</label>
                <Input
                  type="date"
                  value={editDocDate}
                  onChange={(e) => setEditDocDate(e.target.value)}
                  className="h-8 text-xs font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "กำหนดนัดส่ง / สิ้นสุดการจอง:" : "Due Date:"}</label>
                <Input
                  type="date"
                  value={editDueDate}
                  onChange={(e) => setEditDueDate(e.target.value)}
                  className="h-8 text-xs font-mono text-indigo-600 font-bold"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "พนักงานขาย:" : "Salesperson:"}</label>
                <Input
                  value={selectedDoc.salesperson}
                  readOnly
                  className="h-8 text-xs bg-slate-100 dark:bg-slate-800 text-slate-500"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "เงินมัดจำที่ต้องวาง (฿):" : "Deposit Required:"}</label>
                <Input
                  type="number"
                  step="0.01"
                  value={editDepositReq}
                  onChange={(e) => setEditDepositReq(e.target.value)}
                  className="h-8 text-xs font-mono font-bold"
                />
              </div>

              <div className="md:col-span-2 space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "หมายเหตุการจอง:" : "Remark:"}</label>
                <Input
                  value={editRemark}
                  onChange={(e) => setEditRemark(e.target.value)}
                  placeholder="เช่น งานเทพื้นอาคาร C / กันสต็อกสินค้าไว้ 15 วัน"
                  className="h-8 text-xs"
                />
              </div>
            </div>
          </div>

          {/* Card 2: Reserved Items Table */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-3">
            <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
              <div className="flex items-center gap-2">
                <Package className="w-4 h-4 text-indigo-500" />
                <h3 className="text-xs font-bold text-slate-900 dark:text-white">
                  {isTh ? "รายการสินค้าที่จองสต็อกไว้ (Reserved Inventory Lines)" : "Reserved Inventory Lines"}
                </h3>
              </div>
              <span className="text-[11px] text-slate-500">
                {isTh ? `รวม ${selectedDoc.items.length} รายการ` : `${selectedDoc.items.length} items`}
              </span>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 bg-slate-50/70 dark:bg-slate-950/70">
                    <th className="py-2 px-3 font-semibold">{isTh ? "รหัสสินค้า" : "Item Code"}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "ชื่อสินค้า" : "Description"}</th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "จำนวนที่จอง" : "Qty Reserved"}</th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "สต็อกพร้อมใช้" : "Available"}</th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "ราคา/หน่วย" : "Unit Price"}</th>
                    <th className="py-2 px-3 text-right font-semibold text-indigo-700 dark:text-indigo-400">
                      {isTh ? "มูลค่ารวม" : "Line Total"}
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono text-[11px]">
                  {selectedDoc.items.map((item) => (
                    <tr key={item.id} className="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                      <td className="py-2 px-3 font-bold text-indigo-600 dark:text-indigo-400 whitespace-nowrap">
                        {item.itemCode}
                      </td>
                      <td className="py-2 px-3 font-sans text-slate-800 dark:text-slate-200 max-w-[260px] truncate" title={item.itemName}>
                        {item.itemName}
                      </td>
                      <td className="py-2 px-3 text-right font-bold text-slate-900 dark:text-white">
                        {item.qtyReserved} {item.unit}
                      </td>
                      <td className="py-2 px-3 text-right text-emerald-600 font-semibold">
                        {item.availableStock} {item.unit}
                      </td>
                      <td className="py-2 px-3 text-right text-slate-600 dark:text-slate-400">
                        {Number(item.unitPrice).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="py-2 px-3 text-right font-bold text-indigo-600 dark:text-indigo-400">
                        {Number(item.lineTotal).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Card 3: Summary Banner & Thai Baht Text */}
          <div className="bg-indigo-50/40 dark:bg-indigo-950/20 rounded-xl border border-indigo-200/80 dark:border-indigo-800/60 p-4 space-y-3">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <span className="text-[11px] text-slate-500">{isTh ? "ยอดรวมก่อนภาษี" : "Subtotal"}</span>
                <p className="font-mono text-sm font-semibold text-slate-800 dark:text-slate-200">
                  ฿{Number(selectedDoc.subtotal).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                </p>
              </div>
              <div>
                <span className="text-[11px] text-slate-500">{isTh ? "ภาษีมูลค่าเพิ่ม 7%" : "VAT 7%"}</span>
                <p className="font-mono text-sm font-semibold text-slate-800 dark:text-slate-200">
                  ฿{Number(selectedDoc.vatAmount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                </p>
              </div>
              <div>
                <span className="text-[11px] text-slate-500">{isTh ? "เงินมัดจำที่วางแล้ว" : "Deposit Paid"}</span>
                <p className="font-mono text-sm font-semibold text-emerald-600">
                  ฿{Number(selectedDoc.depositReceived).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                </p>
              </div>
              <div>
                <span className="text-[11px] font-bold text-indigo-800 dark:text-indigo-300">
                  {isTh ? "ยอดสุทธิใบสั่งจอง" : "Net Total"}
                </span>
                <p className="font-mono text-base font-extrabold text-indigo-700 dark:text-indigo-300">
                  ฿{Number(selectedDoc.netTotal).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                </p>
              </div>
            </div>

            <div className="p-2.5 bg-white dark:bg-slate-900 rounded-lg border border-indigo-200 dark:border-indigo-900 text-xs flex items-center justify-between">
              <span className="text-slate-500 font-semibold">{isTh ? "ตัวอักษรไทย:" : "Thai Baht Text:"}</span>
              <span className="font-medium text-slate-800 dark:text-slate-200">({thaiText})</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
