"use client";

import { useMemo, useRef, useState } from "react";
import {
  CheckCircle2,
  FileCheck,
  FileText,
  Layers,
  Plus,
  Printer,
  Search,
  Wallet,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { thaiBahtText } from "@/lib/thai-baht-text";
import { type LanguageCode } from "@/lib/i18n";

export interface ApInvoiceItem {
  id: string;
  invoiceNo: string;
  invoiceDate: string;
  dueDate: string;
  description: string;
  grossAmount: string;
  discountAmount: string;
  whtRate: number; // 0, 1, 2, 3, 5 %
  whtAmount: string;
  netPayAmount: string;
  isSelected: boolean;
}

export interface PaymentPrepareDoc {
  id: string;
  docNo: string;
  docDate: string;
  paymentDueDate: string;
  vendorCode: string;
  vendorName: string;
  taxId: string;
  branch: string;
  paymentMethod: "bank_transfer" | "cheque" | "cash";
  bankAccount: string;
  department: string;
  status: "draft" | "pending_approval" | "approved" | "completed";
  items: ApInvoiceItem[];
  totalGross: string;
  totalDiscount: string;
  totalWht: string;
  netPayment: string;
  remark: string;
  preparedBy: string;
  approvedBy?: string;
}

const INITIAL_PREPARE_DOCS: PaymentPrepareDoc[] = [
  {
    id: "prep-001",
    docNo: "PVP6910-0001",
    docDate: "2026-10-08",
    paymentDueDate: "2026-10-15",
    vendorCode: "AP-001",
    vendorName: "บริษัท ปูนซีเมนต์ไทย (ลำปาง) จำกัด",
    taxId: "0105530012345",
    branch: "สำนักงานใหญ่",
    paymentMethod: "bank_transfer",
    bankAccount: "KBANK กระแสรายวัน 045-1-88992-1",
    department: "ฝ่ายจัดซื้อและคลัง",
    status: "approved",
    totalGross: "128400.00",
    totalDiscount: "1000.00",
    totalWht: "3822.00",
    netPayment: "123578.00",
    remark: "เตรียมจ่ายตามรอบวันที่ 15 ต.ค. 2569 บิลคอนกรีตผสมเสร็จ",
    preparedBy: "นางสาวกาญจนา การเงินดี",
    approvedBy: "นายบัญชา บริหารกิจ (สมุห์บัญชี)",
    items: [
      {
        id: "item-01",
        invoiceNo: "TX-6909-0891",
        invoiceDate: "2026-09-15",
        dueDate: "2026-10-15",
        description: "ปูนซีเมนต์ผสมสำเร็จ ปอร์ตแลนด์ประเภท 1 (100 ถุง)",
        grossAmount: "64200.00",
        discountAmount: "500.00",
        whtRate: 3,
        whtAmount: "1911.00",
        netPayAmount: "61789.00",
        isSelected: true,
      },
      {
        id: "item-02",
        invoiceNo: "TX-6909-0945",
        invoiceDate: "2026-09-22",
        dueDate: "2026-10-15",
        description: "หินก่อสร้าง เบอร์ 1 และเบอร์ 2 รวม 40 คิว",
        grossAmount: "64200.00",
        discountAmount: "500.00",
        whtRate: 3,
        whtAmount: "1911.00",
        netPayAmount: "61789.00",
        isSelected: true,
      },
    ],
  },
  {
    id: "prep-002",
    docNo: "PVP6910-0002",
    docDate: "2026-10-09",
    paymentDueDate: "2026-10-20",
    vendorCode: "AP-002",
    vendorName: "บริษัท สยามสตีลอินดัสตรี จำกัด (มหาชน)",
    taxId: "0107537000542",
    branch: "สำนักงานใหญ่",
    paymentMethod: "cheque",
    bankAccount: "BBL กระแสรายวัน 120-3-04987-5",
    department: "ฝ่ายผลิตและคลังสินค้า",
    status: "pending_approval",
    totalGross: "214000.00",
    totalDiscount: "2000.00",
    totalWht: "2120.00",
    netPayment: "209880.00",
    remark: "สั่งซื้อเหล็กเส้นข้ออ้อย มอก. ล็อตงานโครงการอาคารพาณิชย์",
    preparedBy: "นางสาวกาญจนา การเงินดี",
    items: [
      {
        id: "item-03",
        invoiceNo: "INV-ST-99120",
        invoiceDate: "2026-09-20",
        dueDate: "2026-10-20",
        description: "เหล็กเส้นกลม SR24 ขนาด 9 มม. และ DB12",
        grossAmount: "214000.00",
        discountAmount: "2000.00",
        whtRate: 1,
        whtAmount: "2120.00",
        netPayAmount: "209880.00",
        isSelected: true,
      },
    ],
  },
  {
    id: "prep-003",
    docNo: "PVP6910-0003",
    docDate: "2026-10-10",
    paymentDueDate: "2026-10-31",
    vendorCode: "AP-003",
    vendorName: "บริษัท โตโยต้าทรานสปอร์ต แอนด์ โลจิสติกส์ จำกัด",
    taxId: "0105548099881",
    branch: "สาขา 00001",
    paymentMethod: "bank_transfer",
    bankAccount: "KBANK กระแสรายวัน 045-1-88992-1",
    department: "ฝ่ายจัดส่ง",
    status: "draft",
    totalGross: "48150.00",
    totalDiscount: "0.00",
    totalWht: "450.00",
    netPayment: "47700.00",
    remark: "ค่าขนส่งวัสดุก่อสร้างประจำเดือนกันยายน",
    preparedBy: "นายสุวิทย์ บัญชีทุน",
    items: [
      {
        id: "item-04",
        invoiceNo: "LOG-2026-8812",
        invoiceDate: "2026-09-30",
        dueDate: "2026-10-31",
        description: "ค่าบริการขนส่งและกระจายสินค้า 15 เที่ยว",
        grossAmount: "48150.00",
        discountAmount: "0.00",
        whtRate: 1,
        whtAmount: "450.00",
        netPayAmount: "47700.00",
        isSelected: true,
      },
    ],
  },
];

export interface ApPaymentPrepareScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function ApPaymentPrepareScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: ApPaymentPrepareScreenProps) {
  const isTh = language === "th";
  const [docs, setDocs] = useState<PaymentPrepareDoc[]>(INITIAL_PREPARE_DOCS);
  const [selectedId, setSelectedId] = useState<string>(INITIAL_PREPARE_DOCS[0].id);
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [showPrintModal, setShowPrintModal] = useState(false);
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 38,
    min: 25,
    max: 60,
    storageKey: "bcai_split_ap_payment_prepare",
  });

  const selectedDoc = useMemo(() => {
    return docs.find((d) => d.id === selectedId) || docs[0];
  }, [docs, selectedId]);

  // Form edit states
  const [editDocDate, setEditDocDate] = useState(selectedDoc.docDate);
  const [editDueDate, setEditDueDate] = useState(selectedDoc.paymentDueDate);
  const [editMethod, setEditMethod] = useState(selectedDoc.paymentMethod);
  const [editBankAccount, setEditBankAccount] = useState(selectedDoc.bankAccount);
  const [editRemark, setEditRemark] = useState(selectedDoc.remark);

  const handleSelectDoc = (doc: PaymentPrepareDoc) => {
    setSelectedId(doc.id);
    setEditDocDate(doc.docDate);
    setEditDueDate(doc.paymentDueDate);
    setEditMethod(doc.paymentMethod);
    setEditBankAccount(doc.bankAccount);
    setEditRemark(doc.remark);
    setNotification(null);
  };

  const filteredDocs = useMemo(() => {
    return docs.filter((d) => {
      const matchSearch =
        !searchQuery.trim() ||
        d.docNo.toLowerCase().includes(searchQuery.toLowerCase()) ||
        d.vendorName.toLowerCase().includes(searchQuery.toLowerCase()) ||
        d.vendorCode.toLowerCase().includes(searchQuery.toLowerCase());
      const matchStatus = statusFilter === "all" || d.status === statusFilter;
      return matchSearch && matchStatus;
    });
  }, [docs, searchQuery, statusFilter]);

  // Save changes handler
  const handleSaveDoc = () => {
    setDocs((prev) =>
      prev.map((d) => {
        if (d.id === selectedDoc.id) {
          return {
            ...d,
            docDate: editDocDate,
            paymentDueDate: editDueDate,
            paymentMethod: editMethod,
            bankAccount: editBankAccount,
            remark: editRemark,
          };
        }
        return d;
      })
    );
    setNotification(isTh ? "บันทึกเอกสารเตรียมจ่ายชำระหนี้เรียบร้อย" : "Payment preparation saved");
  };

  // Status transitions
  const handleApprove = () => {
    setDocs((prev) =>
      prev.map((d) => {
        if (d.id === selectedDoc.id) {
          return {
            ...d,
            status: "approved",
            approvedBy: "นายบัญชา บริหารกิจ (สมุห์บัญชี)",
          };
        }
        return d;
      })
    );
    setNotification(isTh ? "อนุมัติเอกสารเตรียมจ่ายชำระหนี้เรียบร้อย" : "Preparation doc approved");
  };

  const handleCreateNew = () => {
    const nextSeq = String(docs.length + 1).padStart(4, "0");
    const newDoc: PaymentPrepareDoc = {
      id: `prep-${Date.now()}`,
      docNo: `PVP6910-${nextSeq}`,
      docDate: "2026-10-10",
      paymentDueDate: "2026-10-25",
      vendorCode: "AP-004",
      vendorName: "บริษัท อิตาเลียนไทย บิลดิ้ง ซัพพลาย จำกัด",
      taxId: "0105541099234",
      branch: "สำนักงานใหญ่",
      paymentMethod: "bank_transfer",
      bankAccount: "KBANK กระแสรายวัน 045-1-88992-1",
      department: "ฝ่ายจัดซื้อ",
      status: "draft",
      totalGross: "53500.00",
      totalDiscount: "500.00",
      totalWht: "1590.00",
      netPayment: "51410.00",
      remark: "เอกสารเตรียมจ่ายค่าวัสดุก่อสร้างงานโครงสร้าง",
      preparedBy: "ผู้ใช้งานปัจจุบัน",
      items: [
        {
          id: `item-${Date.now()}`,
          invoiceNo: "INV-ITB-0988",
          invoiceDate: "2026-10-01",
          dueDate: "2026-10-25",
          description: "กระเบื้องปูพื้นแกรนิตโต้ 60x60 ซม. (150 กล่อง)",
          grossAmount: "53500.00",
          discountAmount: "500.00",
          whtRate: 3,
          whtAmount: "1590.00",
          netPayAmount: "51410.00",
          isSelected: true,
        },
      ],
    };

    setDocs([newDoc, ...docs]);
    handleSelectDoc(newDoc);
    setNotification(isTh ? `สร้างเอกสารเตรียมจ่ายใหม่ ${newDoc.docNo}` : `Created doc ${newDoc.docNo}`);
  };

  const thaiText = useMemo(() => {
    return thaiBahtText(selectedDoc.netPayment);
  }, [selectedDoc.netPayment]);

  return (
    <div
      ref={containerRef}
      className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 text-slate-800 dark:text-slate-100 overflow-hidden font-sans"
    >
      {/* Top Header Bar */}
      <div className="flex flex-wrap items-center justify-between px-4 py-2.5 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 gap-3">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-emerald-50 dark:bg-emerald-950/50 text-emerald-600 dark:text-emerald-400 border border-emerald-200/60 dark:border-emerald-800/40">
            <Wallet className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold tracking-tight text-slate-900 dark:text-white">
                {isTh ? "บันทึกเตรียมจ่ายชำระหนี้ (Payment Preparation)" : "Prepare Debt Payment"}
              </h1>
              <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800">
                Champ Parity 20000+
              </span>
            </div>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {isTh
                ? "รวบรวมใบรับวางบิล/ใบแจ้งหนี้เพื่อจัดเตรียมการจ่ายเงิน คำนวณภาษีหัก ณ ที่จ่าย และขออนุมัติก่อนออกเช็ค"
                : "Consolidate supplier invoices for payment approval, WHT deduction & cheque preparation"}
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
            <span>{isTh ? "เพิ่มใบเตรียมจ่าย" : "New Preparation"}</span>
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowPrintModal(true)}
            className="text-xs h-8 gap-1.5"
          >
            <Printer className="w-3.5 h-3.5" />
            <span>{isTh ? "พิมพ์ใบเตรียมจ่าย" : "Print Voucher"}</span>
          </Button>
          <Button
            size="sm"
            onClick={handleSaveDoc}
            className="text-xs h-8 gap-1.5 bg-slate-800 hover:bg-slate-900 text-white dark:bg-slate-200 dark:text-slate-900"
          >
            <FileCheck className="w-3.5 h-3.5" />
            <span>{isTh ? "บันทึก" : "Save"}</span>
          </Button>
        </div>
      </div>

      {/* Notification Banner */}
      {notification && (
        <div className="flex items-center justify-between px-4 py-1.5 bg-emerald-50 dark:bg-emerald-950/50 text-emerald-800 dark:text-emerald-200 text-xs border-b border-emerald-200 dark:border-emerald-900">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 text-emerald-600" />
            <span>{notification}</span>
          </div>
          <button onClick={() => setNotification(null)} className="text-slate-400 hover:text-slate-600">✕</button>
        </div>
      )}

      {/* Main Split Layout */}
      <div className="flex flex-1 overflow-hidden relative">
        {/* Left List Pane: Document List */}
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
                placeholder={isTh ? "ค้นหาเลขที่เอกสาร / เจ้าหนี้..." : "Search doc or vendor..."}
                className="pl-8 h-8 text-xs bg-white dark:bg-slate-950"
              />
            </div>
            <div className="flex items-center gap-1.5">
              {["all", "draft", "pending_approval", "approved"].map((st) => (
                <button
                  key={st}
                  onClick={() => setStatusFilter(st)}
                  className={`text-[11px] px-2 py-0.5 rounded-full border transition-all ${
                    statusFilter === st
                      ? "bg-emerald-600 text-white border-emerald-600 font-semibold"
                      : "bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 border-slate-200 dark:border-slate-800 hover:bg-slate-100"
                  }`}
                >
                  {st === "all"
                    ? isTh ? "ทั้งหมด" : "All"
                    : st === "draft"
                    ? isTh ? "ร่าง" : "Draft"
                    : st === "pending_approval"
                    ? isTh ? "รออนุมัติ" : "Pending"
                    : isTh ? "อนุมัติแล้ว" : "Approved"}
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
                        doc.status === "approved"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-300/50"
                          : doc.status === "pending_approval"
                          ? "bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-300 border border-amber-300/50"
                          : "bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400"
                      }`}
                    >
                      {doc.status === "approved"
                        ? isTh ? "อนุมัติแล้ว" : "Approved"
                        : doc.status === "pending_approval"
                        ? isTh ? "รออนุมัติ" : "Pending"
                        : isTh ? "ร่าง" : "Draft"}
                    </span>
                  </div>

                  <p className="text-xs font-semibold text-slate-800 dark:text-slate-200 mt-1 truncate">
                    {doc.vendorName}
                  </p>

                  <div className="flex items-center justify-between mt-1 text-[11px] text-slate-500">
                    <span>{isTh ? `กำหนดจ่าย: ${doc.paymentDueDate}` : `Due: ${doc.paymentDueDate}`}</span>
                    <span className="font-mono font-bold text-emerald-600 dark:text-emerald-400">
                      ฿{Number(doc.netPayment).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>

          {/* List Footer Total */}
          <div className="p-3 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 flex justify-between items-center">
            <span>{isTh ? `รวม ${filteredDocs.length} เอกสาร` : `${filteredDocs.length} docs`}</span>
            <span className="font-mono font-bold text-slate-800 dark:text-slate-200">
              ฿{filteredDocs
                .reduce((acc, d) => acc + Number(d.netPayment), 0)
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

        {/* Right Pane: Document Details, Invoice Items & Payment Setup */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="flex flex-col h-full bg-slate-50 dark:bg-slate-950 overflow-y-auto p-4 space-y-4"
        >
          {/* Card 1: Document Header Information */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-4">
            <div className="flex flex-wrap items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800 gap-2">
              <div className="flex items-center gap-2">
                <FileText className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
                <div>
                  <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                    <span>{selectedDoc.docNo}</span>
                    <span
                      className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${
                        selectedDoc.status === "approved"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300"
                          : "bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-300"
                      }`}
                    >
                      {selectedDoc.status === "approved" ? "อนุมัติแล้ว" : "รออนุมัติ"}
                    </span>
                  </h2>
                  <p className="text-xs text-slate-500">{selectedDoc.vendorName} ({selectedDoc.vendorCode})</p>
                </div>
              </div>

              {selectedDoc.status !== "approved" && (
                <Button
                  size="sm"
                  onClick={handleApprove}
                  className="text-xs h-7 gap-1 bg-emerald-600 hover:bg-emerald-700 text-white"
                >
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  <span>{isTh ? "อนุมัติเตรียมจ่าย" : "Approve"}</span>
                </Button>
              )}
            </div>

            {/* Form Fields Grid */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">
                  {isTh ? "วันที่เอกสาร:" : "Doc Date:"}
                </label>
                <Input
                  type="date"
                  value={editDocDate}
                  onChange={(e) => setEditDocDate(e.target.value)}
                  className="h-8 text-xs font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">
                  {isTh ? "กำหนดวันที่จ่ายเงิน:" : "Payment Due Date:"}
                </label>
                <Input
                  type="date"
                  value={editDueDate}
                  onChange={(e) => setEditDueDate(e.target.value)}
                  className="h-8 text-xs font-mono text-emerald-600 font-bold"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">
                  {isTh ? "วิธีการจ่ายเงิน:" : "Payment Method:"}
                </label>
                <select
                  value={editMethod}
                  onChange={(e) => setEditMethod(e.target.value as PaymentPrepareDoc["paymentMethod"])}
                  className="w-full h-8 px-2 text-xs rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900"
                >
                  <option value="bank_transfer">{isTh ? "โอนเงินผ่านธนาคาร (Bank Transfer)" : "Bank Transfer"}</option>
                  <option value="cheque">{isTh ? "จ่ายด้วยเช็คธนาคาร (Cheque)" : "Cheque"}</option>
                  <option value="cash">{isTh ? "เงินสด / สดย่อย (Cash)" : "Cash"}</option>
                </select>
              </div>

              <div className="md:col-span-2 space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">
                  {isTh ? "บัญชีธนาคารต้นทางที่จะตัดจ่าย:" : "Paying Bank Account:"}
                </label>
                <Input
                  value={editBankAccount}
                  onChange={(e) => setEditBankAccount(e.target.value)}
                  className="h-8 text-xs font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">
                  {isTh ? "แผนก / โครงการ:" : "Department / Project:"}
                </label>
                <Input
                  value={selectedDoc.department}
                  readOnly
                  className="h-8 text-xs bg-slate-100 dark:bg-slate-800 text-slate-500"
                />
              </div>

              <div className="md:col-span-3 space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">
                  {isTh ? "หมายเหตุการเตรียมจ่าย:" : "Remark:"}
                </label>
                <Input
                  value={editRemark}
                  onChange={(e) => setEditRemark(e.target.value)}
                  placeholder={isTh ? "ระบุเหตุผล เช่น รอบจ่ายวันที่ 15 ต.ค. / บิลสั่งซื้อปูน" : "Remark..."}
                  className="h-8 text-xs"
                />
              </div>
            </div>
          </div>

          {/* Card 2: AP Invoices Breakdown Table */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-3">
            <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
              <div className="flex items-center gap-2">
                <Layers className="w-4 h-4 text-emerald-600" />
                <h3 className="text-xs font-bold text-slate-900 dark:text-white">
                  {isTh ? "รายการใบแจ้งหนี้ / ใบรับวางบิลที่เตรียมจ่าย" : "Attached AP Invoices"}
                </h3>
              </div>
              <span className="text-[11px] text-slate-500">
                {isTh ? `รวม ${selectedDoc.items.length} รายการบิล` : `${selectedDoc.items.length} items`}
              </span>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 bg-slate-50/70 dark:bg-slate-950/70">
                    <th className="py-2 px-3 font-semibold">{isTh ? "เลขที่บิล" : "Invoice No."}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "วันที่บิล" : "Inv Date"}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "รายละเอียดสินค้า/บริการ" : "Description"}</th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "ยอดหนี้เต็ม" : "Gross"}</th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "ส่วนลด" : "Discount"}</th>
                    <th className="py-2 px-3 text-right font-semibold">{isTh ? "หัก ณ ที่จ่าย" : "WHT"}</th>
                    <th className="py-2 px-3 text-right font-semibold text-emerald-700 dark:text-emerald-400">
                      {isTh ? "ยอดเตรียมจ่าย" : "Net Payable"}
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono text-[11px]">
                  {selectedDoc.items.map((item) => (
                    <tr key={item.id} className="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                      <td className="py-2 px-3 font-bold text-indigo-600 dark:text-indigo-400 whitespace-nowrap">
                        {item.invoiceNo}
                      </td>
                      <td className="py-2 px-3 text-slate-500 whitespace-nowrap">{item.invoiceDate}</td>
                      <td className="py-2 px-3 font-sans text-slate-700 dark:text-slate-300 max-w-[240px] truncate" title={item.description}>
                        {item.description}
                      </td>
                      <td className="py-2 px-3 text-right text-slate-700 dark:text-slate-300">
                        {Number(item.grossAmount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="py-2 px-3 text-right text-amber-600">
                        {Number(item.discountAmount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="py-2 px-3 text-right text-rose-600">
                        {Number(item.whtAmount).toLocaleString("th-TH", { minimumFractionDigits: 2 })} ({item.whtRate}%)
                      </td>
                      <td className="py-2 px-3 text-right font-bold text-emerald-600 dark:text-emerald-400">
                        {Number(item.netPayAmount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Card 3: Financial Summary & Thai Baht Text */}
          <div className="bg-emerald-50/50 dark:bg-emerald-950/20 rounded-xl border border-emerald-200/80 dark:border-emerald-800/60 p-4 space-y-3">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <span className="text-[11px] text-slate-500">{isTh ? "ยอดหนี้ตามบิลรวม" : "Gross Total"}</span>
                <p className="font-mono text-sm font-semibold text-slate-800 dark:text-slate-200">
                  ฿{Number(selectedDoc.totalGross).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                </p>
              </div>
              <div>
                <span className="text-[11px] text-slate-500">{isTh ? "ส่วนลดจ่ายรวม" : "Settlement Discount"}</span>
                <p className="font-mono text-sm font-semibold text-amber-600">
                  -฿{Number(selectedDoc.totalDiscount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                </p>
              </div>
              <div>
                <span className="text-[11px] text-slate-500">{isTh ? "ภาษีหัก ณ ที่จ่ายรวม" : "Withholding Tax"}</span>
                <p className="font-mono text-sm font-semibold text-rose-600">
                  -฿{Number(selectedDoc.totalWht).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                </p>
              </div>
              <div>
                <span className="text-[11px] font-bold text-emerald-800 dark:text-emerald-300">
                  {isTh ? "ยอดสุทธิที่ต้องเตรียมจ่าย" : "Net Payment Total"}
                </span>
                <p className="font-mono text-base font-extrabold text-emerald-700 dark:text-emerald-300">
                  ฿{Number(selectedDoc.netPayment).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                </p>
              </div>
            </div>

            {/* Thai Baht Text Banner */}
            <div className="p-2.5 bg-white dark:bg-slate-900 rounded-lg border border-emerald-200 dark:border-emerald-900 text-xs text-slate-700 dark:text-slate-300 flex items-center gap-2">
              <span className="font-semibold text-emerald-800 dark:text-emerald-400">
                {isTh ? "ตัวอักษร:" : "Baht Text:"}
              </span>
              <span className="font-medium tracking-wide">({thaiText})</span>
            </div>
          </div>
        </div>
      </div>

      {/* Print Preview Modal */}
      {showPrintModal && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white dark:bg-slate-900 rounded-xl shadow-2xl max-w-2xl w-full border border-slate-200 dark:border-slate-800 overflow-hidden">
            <div className="flex items-center justify-between p-4 border-b border-slate-200 dark:border-slate-800">
              <div className="flex items-center gap-2">
                <Printer className="w-5 h-5 text-emerald-600" />
                <h3 className="font-bold text-sm text-slate-900 dark:text-white">
                  {isTh ? `พิมพ์ใบเตรียมจ่ายชำระหนี้ - ${selectedDoc.docNo}` : `Print Voucher - ${selectedDoc.docNo}`}
                </h3>
              </div>
              <button
                onClick={() => setShowPrintModal(false)}
                className="p-1 rounded-md text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="p-6 space-y-4 text-xs font-sans">
              <div className="text-center space-y-1 border-b pb-4">
                <h2 className="text-base font-bold text-slate-900 dark:text-white">ใบขออนุมัติเตรียมจ่ายชำระหนี้ (Payment Voucher Request)</h2>
                <p className="text-slate-500">บริษัท รุ่งเรืองกิจการค้า จำกัด (สำนักงานใหญ่)</p>
              </div>

              <div className="grid grid-cols-2 gap-2 text-[11px]">
                <div><span className="font-semibold">เลขที่เอกสาร:</span> {selectedDoc.docNo}</div>
                <div><span className="font-semibold">วันที่:</span> {selectedDoc.docDate}</div>
                <div><span className="font-semibold">ผู้รับเงิน (เจ้าหนี้):</span> {selectedDoc.vendorName}</div>
                <div><span className="font-semibold">กำหนดจ่าย:</span> {selectedDoc.paymentDueDate}</div>
                <div><span className="font-semibold">วิธีชำระ:</span> {selectedDoc.paymentMethod}</div>
                <div><span className="font-semibold">บัญชีที่จ่าย:</span> {selectedDoc.bankAccount}</div>
              </div>

              <div className="border rounded-md p-3 bg-slate-50 dark:bg-slate-950 font-mono text-[11px] space-y-1">
                <div className="flex justify-between">
                  <span>ยอดหนี้รวม:</span>
                  <span>฿{Number(selectedDoc.totalGross).toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="flex justify-between text-amber-600">
                  <span>หัก ส่วนลดจ่าย:</span>
                  <span>-฿{Number(selectedDoc.totalDiscount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="flex justify-between text-rose-600">
                  <span>หัก ภาษี ณ ที่จ่าย:</span>
                  <span>-฿{Number(selectedDoc.totalWht).toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="flex justify-between font-bold text-emerald-600 border-t pt-1">
                  <span>ยอดสุทธิที่ต้องจ่าย:</span>
                  <span>฿{Number(selectedDoc.netPayment).toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="font-sans text-[11px] text-slate-600 pt-1">({thaiText})</div>
              </div>

              <div className="grid grid-cols-2 gap-8 pt-6 text-center text-[11px]">
                <div className="border-t pt-2">
                  <p className="font-semibold">{selectedDoc.preparedBy}</p>
                  <p className="text-slate-400">ผู้จัดทำ (Prepared By)</p>
                </div>
                <div className="border-t pt-2">
                  <p className="font-semibold">{selectedDoc.approvedBy || "..................................................."}</p>
                  <p className="text-slate-400">ผู้อนุมัติ (Approved By)</p>
                </div>
              </div>
            </div>

            <div className="p-3 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 flex justify-end gap-2">
              <Button variant="outline" size="sm" onClick={() => setShowPrintModal(false)} className="text-xs h-8">
                {isTh ? "ปิด" : "Close"}
              </Button>
              <Button
                size="sm"
                onClick={() => {
                  window.print();
                  setShowPrintModal(false);
                }}
                className="text-xs h-8 gap-1.5 bg-emerald-600 hover:bg-emerald-700 text-white"
              >
                <Printer className="w-3.5 h-3.5" />
                <span>{isTh ? "สั่งพิมพ์" : "Print Now"}</span>
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
