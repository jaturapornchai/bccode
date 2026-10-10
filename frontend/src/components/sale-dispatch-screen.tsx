"use client";

import { useMemo, useRef, useState } from "react";
import {
  CheckCircle2,
  FileCheck,
  Package,
  Plus,
  Printer,
  Search,
  Truck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export interface DispatchLineItem {
  id: string;
  itemCode: string;
  itemName: string;
  warehouse: string;
  shelf: string;
  orderedQty: number;
  dispatchQty: number;
  remainingQty: number;
  unit: string;
}

export interface GoodsDispatchDoc {
  id: string;
  docNo: string;
  docDate: string;
  deliveryDate: string;
  soRefNo: string;
  customerCode: string;
  customerName: string;
  shippingAddress: string;
  driverName: string;
  vehiclePlate: string;
  status: "draft" | "picking" | "dispatched" | "delivered";
  items: DispatchLineItem[];
  totalQty: number;
  remark: string;
  dispatchedBy: string;
}

const INITIAL_DISPATCHES: GoodsDispatchDoc[] = [
  {
    id: "do-01",
    docNo: "DO6910-0001",
    docDate: "2026-10-09",
    deliveryDate: "2026-10-10",
    soRefNo: "SO6910-0038",
    customerCode: "AR-001",
    customerName: "บริษัท สยามคอนกรีตมิกซ์ จำกัด",
    shippingAddress: "โครงการก่อสร้างสะพานข้ามแยก ถ.พระราม 2 ซอย 44 แขวงบางมด กรุงเทพฯ",
    driverName: "นายสมพงษ์ รถพร้อม",
    vehiclePlate: "70-8845 กทม. (รถบรรทุก 10 ล้อ)",
    status: "dispatched",
    totalQty: 250,
    remark: "จัดส่งรอบเช้า 08:30 น. โทรแจ้งผู้ควบคุมงานหน้าไซต์ก่อนถึง 30 นาที",
    dispatchedBy: "นายเกรียงไกร คลังดี",
    items: [
      {
        id: "ditem-01",
        itemCode: "MAT-CM-01",
        itemName: "ปูนซีเมนต์ผสมสำเร็จ ปอร์ตแลนด์ประเภท 1 (ถุง 50 กก.)",
        warehouse: "WH-01 คลังใหญ่พระประแดง",
        shelf: "A-01-04",
        orderedQty: 250,
        dispatchQty: 250,
        remainingQty: 0,
        unit: "ถุง",
      },
    ],
  },
  {
    id: "do-02",
    docNo: "DO6910-0002",
    docDate: "2026-10-10",
    deliveryDate: "2026-10-11",
    soRefNo: "SO6910-0042",
    customerCode: "AR-002",
    customerName: "ห้างหุ้นส่วนจำกัด นครหลวงวิศวกรรม",
    shippingAddress: "เลขที่ 88/1 หมู่ 3 ต.คลองหนึ่ง อ.คลองหลวง จ.ปทุมธานี",
    driverName: "นายวิเชียร มั่นคง",
    vehiclePlate: "82-9912 นนทบุรี (รถพ่วง 18 ล้อ)",
    status: "picking",
    totalQty: 600,
    remark: "เหล็กเส้นข้ออ้อย มีใบเซอร์โรงงานแนบพร้อมเอกสารส่งของ",
    dispatchedBy: "นายเกรียงไกร คลังดี",
    items: [
      {
        id: "ditem-02",
        itemCode: "STL-DB12",
        itemName: "เหล็กเส้นข้ออ้อย มอก. SD40 ขนาด DB12 ยาว 10 ม.",
        warehouse: "WH-02 คลังเหล็กและโครงสร้าง",
        shelf: "BAY-03",
        orderedQty: 1000,
        dispatchQty: 600,
        remainingQty: 400,
        unit: "เส้น",
      },
    ],
  },
  {
    id: "do-03",
    docNo: "DO6910-0003",
    docDate: "2026-10-08",
    deliveryDate: "2026-10-08",
    soRefNo: "SO6910-0035",
    customerCode: "AR-004",
    customerName: "บริษัท สหรุ่งโรจน์ ดีเวลลอปเมนท์ จำกัด",
    shippingAddress: "โครงการหมู่บ้านเดอะพาร์ควิว ถ.ราชพฤกษ์ นนทบุรี",
    driverName: "นายประเสริฐ ดีเลิศ",
    vehiclePlate: "3ฒภ-4411 กทม. (รถกระบะ 4 ล้อ)",
    status: "delivered",
    totalQty: 120,
    remark: "จัดส่งและลูกค้าตรวจรับสินค้าเรียบร้อย มีลายเซ็นกำกับ",
    dispatchedBy: "นายเกรียงไกร คลังดี",
    items: [
      {
        id: "ditem-03",
        itemCode: "FLR-GRN60",
        itemName: "กระเบื้องปูพื้นแกรนิตโต้ นาโนไวท์ 60x60 ซม. (เกรด A)",
        warehouse: "WH-03 คลังกระเบื้องและสุขภัณฑ์",
        shelf: "PALLET-12",
        orderedQty: 120,
        dispatchQty: 120,
        remainingQty: 0,
        unit: "กล่อง",
      },
    ],
  },
];

export interface SaleDispatchScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SaleDispatchScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: SaleDispatchScreenProps) {
  const isTh = language === "th";
  const [docs, setDocs] = useState<GoodsDispatchDoc[]>(INITIAL_DISPATCHES);
  const [selectedId, setSelectedId] = useState<string>(INITIAL_DISPATCHES[0].id);
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 38,
    min: 25,
    max: 60,
    storageKey: "bcai_split_sale_dispatch",
  });

  const selectedDoc = useMemo(() => {
    return docs.find((d) => d.id === selectedId) || docs[0];
  }, [docs, selectedId]);

  // Form states
  const [editDocDate, setEditDocDate] = useState(selectedDoc.docDate);
  const [editDeliveryDate, setEditDeliveryDate] = useState(selectedDoc.deliveryDate);
  const [editDriver, setEditDriver] = useState(selectedDoc.driverName);
  const [editPlate, setEditPlate] = useState(selectedDoc.vehiclePlate);
  const [editAddress, setEditAddress] = useState(selectedDoc.shippingAddress);
  const [editRemark, setEditRemark] = useState(selectedDoc.remark);

  const handleSelectDoc = (doc: GoodsDispatchDoc) => {
    setSelectedId(doc.id);
    setEditDocDate(doc.docDate);
    setEditDeliveryDate(doc.deliveryDate);
    setEditDriver(doc.driverName);
    setEditPlate(doc.vehiclePlate);
    setEditAddress(doc.shippingAddress);
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
        d.soRefNo.toLowerCase().includes(q) ||
        d.customerName.toLowerCase().includes(q) ||
        d.driverName.toLowerCase().includes(q) ||
        d.vehiclePlate.toLowerCase().includes(q);
      return matchStatus && matchSearch;
    });
  }, [docs, statusFilter, searchQuery]);

  const handleSaveDoc = () => {
    setDocs((prev) =>
      prev.map((d) => {
        if (d.id === selectedDoc.id) {
          return {
            ...d,
            docDate: editDocDate,
            deliveryDate: editDeliveryDate,
            driverName: editDriver,
            vehiclePlate: editPlate,
            shippingAddress: editAddress,
            remark: editRemark,
          };
        }
        return d;
      })
    );
    setNotification(isTh ? "บันทึกใบสั่งจ่ายสินค้าเรียบร้อย" : "Dispatch order saved");
  };

  const handleCreateNew = () => {
    const nextSeq = String(docs.length + 1).padStart(4, "0");
    const newDoc: GoodsDispatchDoc = {
      id: `do-${Date.now()}`,
      docNo: `DO6910-${nextSeq}`,
      docDate: "2026-10-10",
      deliveryDate: "2026-10-12",
      soRefNo: "SO6910-0048",
      customerCode: "AR-001",
      customerName: "บริษัท สยามคอนกรีตมิกซ์ จำกัด",
      shippingAddress: "สำนักงานใหญ่ พระราม 3 กรุงเทพฯ",
      driverName: "นายประเสริฐ ดีเลิศ",
      vehiclePlate: "70-1122 นนทบุรี",
      status: "draft",
      totalQty: 100,
      remark: "ใบสั่งจ่ายสินค้าล็อตใหม่",
      dispatchedBy: "ผู้ใช้งานปัจจุบัน",
      items: [
        {
          id: `ditem-${Date.now()}`,
          itemCode: "MAT-CM-01",
          itemName: "ปูนซีเมนต์ผสมสำเร็จ ปอร์ตแลนด์ประเภท 1 (ถุง 50 กก.)",
          warehouse: "WH-01 คลังใหญ่พระประแดง",
          shelf: "A-01-04",
          orderedQty: 100,
          dispatchQty: 100,
          remainingQty: 0,
          unit: "ถุง",
        },
      ],
    };

    setDocs([newDoc, ...docs]);
    handleSelectDoc(newDoc);
    setNotification(isTh ? `สร้างใบสั่งจ่ายใหม่ ${newDoc.docNo}` : `Created dispatch order ${newDoc.docNo}`);
  };

  const handleUpdateStatus = (newStatus: GoodsDispatchDoc["status"], labelTh: string) => {
    setDocs((prev) =>
      prev.map((d) => {
        if (d.id === selectedDoc.id) {
          return { ...d, status: newStatus };
        }
        return d;
      })
    );
    setNotification(isTh ? `ปรับสถานะเป็น "${labelTh}" เรียบร้อย` : `Updated status to ${newStatus}`);
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
                {isTh ? "ใบสั่งจ่ายสินค้า (Goods Dispatch Order)" : "Goods Dispatch Order"}
              </h1>
              <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800">
                Champ Parity 30005
              </span>
            </div>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {isTh
                ? "คำสั่งเบิกจ่ายและเตรียมจัดส่งสินค้าตามใบสั่งขาย ตรวจสอบคลัง/ที่เก็บ และมอบหมายพนักงานขนส่ง"
                : "Manage inventory picking and transport dispatch for approved sales orders"}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            size="sm"
            onClick={handleCreateNew}
            className="text-xs h-8 gap-1.5 bg-emerald-600 hover:bg-emerald-700 text-white font-medium shadow-sm"
          >
            <Plus className="w-3.5 h-3.5" />
            <span>{isTh ? "เพิ่มใบสั่งจ่าย" : "New Dispatch"}</span>
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
            <span>{isTh ? "พิมพ์ใบสั่งจ่าย" : "Print"}</span>
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
        {/* Left Pane: Filter & Dispatch List */}
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
                placeholder={isTh ? "ค้นหาเลขที่สั่งจ่าย / SO / ทะเบียนรถ..." : "Search doc, SO or plate..."}
                className="pl-8 h-8 text-xs bg-white dark:bg-slate-950"
              />
            </div>

            {/* Status Pills */}
            <div className="flex flex-wrap gap-1 items-center">
              {[
                { id: "all", th: "ทั้งหมด", en: "All" },
                { id: "draft", th: "ร่าง", en: "Draft" },
                { id: "picking", th: "รอจัดของ", en: "Picking" },
                { id: "dispatched", th: "ส่งแล้ว", en: "Dispatched" },
                { id: "delivered", th: "ถึงแล้ว", en: "Delivered" },
              ].map((st) => (
                <button
                  key={st.id}
                  onClick={() => setStatusFilter(st.id)}
                  className={`text-[10px] px-2 py-0.5 rounded-full border transition-all ${
                    statusFilter === st.id
                      ? "bg-emerald-600 text-white border-emerald-600 font-semibold"
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
                      ? "bg-emerald-50/90 dark:bg-emerald-950/40 border-emerald-300 dark:border-emerald-700 shadow-sm"
                      : "border-transparent hover:bg-slate-100/70 dark:hover:bg-slate-800/60"
                  }`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-mono text-xs font-bold text-slate-900 dark:text-slate-100">
                      {doc.docNo}
                    </span>
                    <span
                      className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${
                        doc.status === "delivered"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-300/50"
                          : doc.status === "dispatched"
                          ? "bg-blue-100 text-blue-800 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-300/50"
                          : doc.status === "picking"
                          ? "bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-300 border border-amber-300/50"
                          : "bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400"
                      }`}
                    >
                      {doc.status === "delivered"
                        ? isTh ? "ส่งมอบแล้ว" : "Delivered"
                        : doc.status === "dispatched"
                        ? isTh ? "กำลังจัดส่ง" : "Dispatched"
                        : doc.status === "picking"
                        ? isTh ? "รอจัดของ" : "Picking"
                        : isTh ? "ร่าง" : "Draft"}
                    </span>
                  </div>

                  <p className="text-xs font-semibold text-slate-800 dark:text-slate-200 mt-1 truncate">
                    {doc.customerName}
                  </p>

                  <div className="flex items-center justify-between mt-1 text-[11px] text-slate-500">
                    <span className="font-mono text-indigo-600 dark:text-indigo-400">อ้างอิง: {doc.soRefNo}</span>
                    <span className="font-medium">{isTh ? `ส่ง: ${doc.deliveryDate}` : `Date: ${doc.deliveryDate}`}</span>
                  </div>
                </div>
              );
            })}
          </div>

          {/* List Footer Total */}
          <div className="p-3 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 flex justify-between items-center">
            <span>{isTh ? `รวม ${filteredDocs.length} ฉบับ` : `${filteredDocs.length} orders`}</span>
            <span className="font-mono font-bold text-slate-800 dark:text-slate-200">
              {filteredDocs.reduce((acc, d) => acc + d.totalQty, 0).toLocaleString()} ชิ้น/หน่วย
            </span>
          </div>
        </div>

        {/* Resizable Splitter */}
        <ResizableSplitter
          isResizing={isResizing}
          onMouseDown={startResize}
          onKeyDown={adjustWithKeyboard}
        />

        {/* Right Pane: Dispatch Detail Form & Picking Table */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="flex flex-col h-full bg-slate-50 dark:bg-slate-950 overflow-y-auto p-4 space-y-4"
        >
          {/* Card 1: Document Header */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-4">
            <div className="flex flex-wrap items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800 gap-2">
              <div className="flex items-center gap-2">
                <Truck className="w-5 h-5 text-emerald-600" />
                <div>
                  <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                    <span>{selectedDoc.docNo}</span>
                    <span className="font-mono text-xs text-indigo-600 font-semibold">(SO: {selectedDoc.soRefNo})</span>
                  </h2>
                  <p className="text-xs text-slate-500">{selectedDoc.customerName} ({selectedDoc.customerCode})</p>
                </div>
              </div>

              {/* Status Actions */}
              <div className="flex items-center gap-1.5 flex-wrap">
                {selectedDoc.status === "draft" && (
                  <Button
                    size="sm"
                    onClick={() => handleUpdateStatus("picking", "รอจัดของ")}
                    className="text-xs h-7 gap-1 bg-amber-600 hover:bg-amber-700 text-white"
                  >
                    <Package className="w-3.5 h-3.5" />
                    <span>{isTh ? "ส่งใบจัดของเข้าคลัง" : "Start Picking"}</span>
                  </Button>
                )}
                {selectedDoc.status === "picking" && (
                  <Button
                    size="sm"
                    onClick={() => handleUpdateStatus("dispatched", "กำลังจัดส่ง")}
                    className="text-xs h-7 gap-1 bg-blue-600 hover:bg-blue-700 text-white"
                  >
                    <Truck className="w-3.5 h-3.5" />
                    <span>{isTh ? "ปล่อยรถจัดส่ง" : "Dispatch"}</span>
                  </Button>
                )}
                {selectedDoc.status === "dispatched" && (
                  <Button
                    size="sm"
                    onClick={() => handleUpdateStatus("delivered", "ส่งมอบแล้ว")}
                    className="text-xs h-7 gap-1 bg-emerald-600 hover:bg-emerald-700 text-white"
                  >
                    <CheckCircle2 className="w-3.5 h-3.5" />
                    <span>{isTh ? "ยืนยันส่งมอบสำเร็จ" : "Mark Delivered"}</span>
                  </Button>
                )}
              </div>
            </div>

            {/* Form Fields */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "วันที่สั่งจ่าย:" : "Dispatch Date:"}</label>
                <Input
                  type="date"
                  value={editDocDate}
                  onChange={(e) => setEditDocDate(e.target.value)}
                  className="h-8 text-xs font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "วันที่กำหนดส่งถึง:" : "Delivery Date:"}</label>
                <Input
                  type="date"
                  value={editDeliveryDate}
                  onChange={(e) => setEditDeliveryDate(e.target.value)}
                  className="h-8 text-xs font-mono text-emerald-600 font-bold"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "พนักงานขับรถ:" : "Driver Name:"}</label>
                <Input
                  value={editDriver}
                  onChange={(e) => setEditDriver(e.target.value)}
                  className="h-8 text-xs"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "ทะเบียนยานพาหนะ:" : "Vehicle Plate:"}</label>
                <Input
                  value={editPlate}
                  onChange={(e) => setEditPlate(e.target.value)}
                  className="h-8 text-xs font-mono"
                />
              </div>

              <div className="md:col-span-2 space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "สถานที่จัดส่งสินค้า:" : "Shipping Destination:"}</label>
                <Input
                  value={editAddress}
                  onChange={(e) => setEditAddress(e.target.value)}
                  className="h-8 text-xs"
                />
              </div>

              <div className="md:col-span-3 space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "หมายเหตุการจัดส่ง:" : "Dispatch Note:"}</label>
                <Input
                  value={editRemark}
                  onChange={(e) => setEditRemark(e.target.value)}
                  className="h-8 text-xs"
                />
              </div>
            </div>
          </div>

          {/* Card 2: Items Breakdown */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-3">
            <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
              <div className="flex items-center gap-2">
                <Package className="w-4 h-4 text-emerald-600" />
                <h3 className="text-xs font-bold text-slate-900 dark:text-white">
                  {isTh ? "รายการสินค้าที่สั่งจ่ายและที่เก็บ (Picking & Shelf Items)" : "Picking & Shelf Items"}
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
                    <th className="py-2 px-3 font-semibold">{isTh ? "คลัง / Shelf" : "Location"}</th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "สั่งตาม SO" : "Ordered"}</th>
                    <th className="py-2 px-3 text-right font-semibold text-emerald-700 dark:text-emerald-400">
                      {isTh ? "จ่ายรอบนี้" : "Dispatch Qty"}
                    </th>
                    <th className="py-2 px-3 text-right font-semibold text-slate-400">{isTh ? "ค้างจ่าย" : "Remaining"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono text-[11px]">
                  {selectedDoc.items.map((item) => (
                    <tr key={item.id} className="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                      <td className="py-2 px-3 font-bold text-indigo-600 dark:text-indigo-400 whitespace-nowrap">
                        {item.itemCode}
                      </td>
                      <td className="py-2 px-3 font-sans text-slate-800 dark:text-slate-200 max-w-[240px] truncate" title={item.itemName}>
                        {item.itemName}
                      </td>
                      <td className="py-2 px-3 font-sans text-slate-500">
                        <span className="font-semibold">{item.warehouse}</span> ({item.shelf})
                      </td>
                      <td className="py-2 px-3 text-right text-slate-600">
                        {item.orderedQty} {item.unit}
                      </td>
                      <td className="py-2 px-3 text-right font-bold text-emerald-600 dark:text-emerald-400">
                        {item.dispatchQty} {item.unit}
                      </td>
                      <td className="py-2 px-3 text-right text-slate-400">
                        {item.remainingQty} {item.unit}
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
