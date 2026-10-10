"use client";

import { useMemo, useRef, useState } from "react";
import {
  CheckCircle2,
  FileCheck,
  Landmark,
  Plus,
  Search,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { thaiBahtText } from "@/lib/thai-baht-text";
import { type LanguageCode } from "@/lib/i18n";

export type BankServiceType = "fee" | "interest_income" | "interest_expense" | "wht_deduction";

export interface BankServiceDoc {
  id: string;
  docNo: string;
  docDate: string;
  serviceType: BankServiceType;
  bankBookCode: string;
  bankName: string;
  accountNo: string;
  glAccountCode: string;
  glAccountName: string;
  amount: string;
  vatRate: number; // 0 or 7
  vatAmount: string;
  netTotal: string;
  remark: string;
  referenceNo: string;
  status: "draft" | "posted" | "cancelled";
  createdBy: string;
}

const INITIAL_BANK_SERVICES: BankServiceDoc[] = [
  {
    id: "bs-01",
    docNo: "BS6910-0001",
    docDate: "2026-10-02",
    serviceType: "fee",
    bankBookCode: "BB-01",
    bankName: "ธนาคารกสิกรไทย",
    accountNo: "045-1-88992-1 (กระแสรายวัน)",
    glAccountCode: "53900",
    glAccountName: "ค่าธรรมเนียมธนาคาร",
    amount: "300.00",
    vatRate: 7,
    vatAmount: "21.00",
    netTotal: "321.00",
    remark: "ค่าธรรมเนียมสั่งซื้อสมุดเช็คเล่มใหม่ 1 เล่ม (25 ฉบับ)",
    referenceNo: "FE-KBANK-9912",
    status: "posted",
    createdBy: "นางสาวกาญจนา การเงินดี",
  },
  {
    id: "bs-02",
    docNo: "BS6910-0002",
    docDate: "2026-10-05",
    serviceType: "fee",
    bankBookCode: "BB-02",
    bankName: "ธนาคารกรุงเทพ",
    accountNo: "120-3-04987-5 (กระแสรายวัน)",
    glAccountCode: "53900",
    glAccountName: "ค่าธรรมเนียมธนาคาร",
    amount: "150.00",
    vatRate: 0,
    vatAmount: "0.00",
    netTotal: "150.00",
    remark: "ค่าธรรมเนียมการโอนเงินต่างธนาคารระบบ SMART/BAHTNET",
    referenceNo: "TRF-88210",
    status: "posted",
    createdBy: "นางสาวกาญจนา การเงินดี",
  },
  {
    id: "bs-03",
    docNo: "BS6910-0003",
    docDate: "2026-09-30",
    serviceType: "interest_income",
    bankBookCode: "BB-03",
    bankName: "ธนาคารไทยพาณิชย์",
    accountNo: "111-2-99881-0 (ออมทรัพย์)",
    glAccountCode: "42100",
    glAccountName: "ดอกเบี้ยรับ",
    amount: "8450.00",
    vatRate: 0,
    vatAmount: "0.00",
    netTotal: "8450.00",
    remark: "ดอกเบี้ยเงินฝากออมทรัพย์ประจำงวดสิ้นสุด 30 ก.ย. 2569",
    referenceNo: "INT-SCB-6909",
    status: "posted",
    createdBy: "นายสุวิทย์ บัญชีทุน",
  },
  {
    id: "bs-04",
    docNo: "BS6910-0004",
    docDate: "2026-09-30",
    serviceType: "wht_deduction",
    bankBookCode: "BB-03",
    bankName: "ธนาคารไทยพาณิชย์",
    accountNo: "111-2-99881-0 (ออมทรัพย์)",
    glAccountCode: "11920",
    glAccountName: "ภาษีเงินได้ถูกหัก ณ ที่จ่าย",
    amount: "84.50",
    vatRate: 0,
    vatAmount: "0.00",
    netTotal: "84.50",
    remark: "ภาษีเงินได้หัก ณ ที่จ่าย 1% จากดอกเบี้ยรับธนาคาร SCB",
    referenceNo: "WHT-SCB-6909",
    status: "posted",
    createdBy: "นายสุวิทย์ บัญชีทุน",
  },
  {
    id: "bs-05",
    docNo: "BS6910-0005",
    docDate: "2026-10-10",
    serviceType: "fee",
    bankBookCode: "BB-01",
    bankName: "ธนาคารกสิกรไทย",
    accountNo: "045-1-88992-1 (กระแสรายวัน)",
    glAccountCode: "53900",
    glAccountName: "ค่าธรรมเนียมธนาคาร",
    amount: "500.00",
    vatRate: 7,
    vatAmount: "35.00",
    netTotal: "535.00",
    remark: "ค่าบริการรายปีระบบ Corporate Internet Banking",
    referenceNo: "K-BIZ-2026",
    status: "draft",
    createdBy: "นางสาวกาญจนา การเงินดี",
  },
];

export interface BankServiceScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function BankServiceScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: BankServiceScreenProps) {
  const isTh = language === "th";
  const [docs, setDocs] = useState<BankServiceDoc[]>(INITIAL_BANK_SERVICES);
  const [selectedId, setSelectedId] = useState<string>(INITIAL_BANK_SERVICES[0].id);
  const [serviceTypeFilter, setServiceTypeFilter] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 38,
    min: 25,
    max: 60,
    storageKey: "bcai_split_bank_service",
  });

  const selectedDoc = useMemo(() => {
    return docs.find((d) => d.id === selectedId) || docs[0];
  }, [docs, selectedId]);

  // Form edit states
  const [editDate, setEditDate] = useState(selectedDoc.docDate);
  const [editAmount, setEditAmount] = useState(selectedDoc.amount);
  const [editVatRate, setEditVatRate] = useState(selectedDoc.vatRate);
  const [editRemark, setEditRemark] = useState(selectedDoc.remark);
  const [editRefNo, setEditRefNo] = useState(selectedDoc.referenceNo);

  const handleSelectDoc = (doc: BankServiceDoc) => {
    setSelectedId(doc.id);
    setEditDate(doc.docDate);
    setEditAmount(doc.amount);
    setEditVatRate(doc.vatRate);
    setEditRemark(doc.remark);
    setEditRefNo(doc.referenceNo);
    setNotification(null);
  };

  const filteredDocs = useMemo(() => {
    return docs.filter((d) => {
      const matchType = serviceTypeFilter === "all" || d.serviceType === serviceTypeFilter;
      const q = searchQuery.toLowerCase().trim();
      const matchSearch =
        !q ||
        d.docNo.toLowerCase().includes(q) ||
        d.bankName.toLowerCase().includes(q) ||
        d.accountNo.toLowerCase().includes(q) ||
        d.remark.toLowerCase().includes(q);
      return matchType && matchSearch;
    });
  }, [docs, serviceTypeFilter, searchQuery]);

  // Dynamic calculations
  const calcVat = useMemo(() => {
    const amt = parseFloat(editAmount) || 0;
    return editVatRate === 7 ? (amt * 0.07).toFixed(2) : "0.00";
  }, [editAmount, editVatRate]);

  const calcNet = useMemo(() => {
    const amt = parseFloat(editAmount) || 0;
    const vat = parseFloat(calcVat) || 0;
    return (amt + vat).toFixed(2);
  }, [editAmount, calcVat]);

  const thaiText = useMemo(() => {
    return thaiBahtText(selectedDoc.netTotal);
  }, [selectedDoc.netTotal]);

  const handleSaveDoc = () => {
    setDocs((prev) =>
      prev.map((d) => {
        if (d.id === selectedDoc.id) {
          return {
            ...d,
            docDate: editDate,
            amount: Number(editAmount || 0).toFixed(2),
            vatRate: editVatRate,
            vatAmount: calcVat,
            netTotal: calcNet,
            remark: editRemark,
            referenceNo: editRefNo,
          };
        }
        return d;
      })
    );
    setNotification(isTh ? "บันทึกรายการบริการธนาคารเรียบร้อย" : "Bank service record saved");
  };

  const handleCreateNew = () => {
    const nextSeq = String(docs.length + 1).padStart(4, "0");
    const newDoc: BankServiceDoc = {
      id: `bs-${Date.now()}`,
      docNo: `BS6910-${nextSeq}`,
      docDate: "2026-10-10",
      serviceType: "fee",
      bankBookCode: "BB-01",
      bankName: "ธนาคารกสิกรไทย",
      accountNo: "045-1-88992-1 (กระแสรายวัน)",
      glAccountCode: "53900",
      glAccountName: "ค่าธรรมเนียมธนาคาร",
      amount: "200.00",
      vatRate: 7,
      vatAmount: "14.00",
      netTotal: "214.00",
      remark: "ค่าธรรมเนียมรายการใหม่",
      referenceNo: `REF-${Date.now().toString().slice(-4)}`,
      status: "draft",
      createdBy: "ผู้ใช้งานปัจจุบัน",
    };

    setDocs([newDoc, ...docs]);
    handleSelectDoc(newDoc);
    setNotification(isTh ? `สร้างรายการบริการธนาคารใหม่ ${newDoc.docNo}` : `Created record ${newDoc.docNo}`);
  };

  const handlePostGL = () => {
    setDocs((prev) =>
      prev.map((d) => {
        if (d.id === selectedDoc.id) {
          return { ...d, status: "posted" };
        }
        return d;
      })
    );
    setNotification(isTh ? `ผ่านรายการ ${selectedDoc.docNo} ลงสมุดรายวันทั่วไปแล้ว` : `Posted ${selectedDoc.docNo} to GL`);
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
            <Landmark className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold tracking-tight text-slate-900 dark:text-white">
                {isTh ? "บริการธนาคาร (Bank Services & Fees)" : "Bank Services & Fees"}
              </h1>
              <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800">
                Champ Parity 40000+
              </span>
            </div>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {isTh
                ? "บันทึกค่าธรรมเนียมธนาคาร ดอกเบี้ยรับ ภาษีหัก ณ ที่จ่าย และดอกเบี้ยจ่าย พร้อมบันทึกบัญชีแยกประเภท"
                : "Record bank service fees, interest income/expense & WHT with GL integration"}
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
            <span>{isTh ? "เพิ่มรายการ" : "New Service"}</span>
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
        {/* Left Pane: Filter & Records List */}
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
                placeholder={isTh ? "ค้นหาเลขที่เอกสาร / บัญชี / ธนาคาร..." : "Search doc, account or bank..."}
                className="pl-8 h-8 text-xs bg-white dark:bg-slate-950"
              />
            </div>

            {/* Service Type Filter Tabs */}
            <div className="flex flex-wrap gap-1 items-center">
              {[
                { id: "all", th: "ทั้งหมด", en: "All" },
                { id: "fee", th: "ค่าธรรมเนียม", en: "Fees" },
                { id: "interest_income", th: "ดอกเบี้ยรับ", en: "Interest In" },
                { id: "wht_deduction", th: "หัก ณ ที่จ่าย", en: "WHT" },
              ].map((t) => (
                <button
                  key={t.id}
                  onClick={() => setServiceTypeFilter(t.id)}
                  className={`text-[10px] px-2 py-0.5 rounded-full border transition-all ${
                    serviceTypeFilter === t.id
                      ? "bg-emerald-600 text-white border-emerald-600 font-semibold"
                      : "bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 border-slate-200 dark:border-slate-800 hover:bg-slate-100"
                  }`}
                >
                  {isTh ? t.th : t.en}
                </button>
              ))}
            </div>
          </div>

          {/* List Items */}
          <div className="flex-1 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800/60 p-1.5 space-y-1">
            {filteredDocs.map((doc) => {
              const isSelected = doc.id === selectedDoc.id;
              const isFee = doc.serviceType === "fee";
              const isInt = doc.serviceType === "interest_income";
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
                        isFee
                          ? "bg-rose-100 text-rose-800 dark:bg-rose-950/70 dark:text-rose-300"
                          : isInt
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/70 dark:text-emerald-300"
                          : "bg-blue-100 text-blue-800 dark:bg-blue-950/70 dark:text-blue-300"
                      }`}
                    >
                      {isFee
                        ? isTh ? "ค่าธรรมเนียม" : "Bank Fee"
                        : isInt
                        ? isTh ? "ดอกเบี้ยรับ" : "Interest Income"
                        : isTh ? "หัก ณ ที่จ่าย" : "WHT"}
                    </span>
                  </div>

                  <p className="text-xs font-semibold text-slate-800 dark:text-slate-200 mt-1 truncate">
                    {doc.remark}
                  </p>

                  <div className="flex items-center justify-between mt-1 text-[11px] text-slate-500">
                    <span className="truncate max-w-[140px]">{doc.bankName}</span>
                    <span className="font-mono font-bold text-emerald-600 dark:text-emerald-400">
                      ฿{Number(doc.netTotal).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>

          {/* List Footer Total */}
          <div className="p-3 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 flex justify-between items-center">
            <span>{isTh ? `รวม ${filteredDocs.length} รายการ` : `${filteredDocs.length} records`}</span>
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

        {/* Right Pane: Document Details & GL Accounting Form */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="flex flex-col h-full bg-slate-50 dark:bg-slate-950 overflow-y-auto p-4 space-y-4"
        >
          {/* Card 1: Document Master Details */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-4">
            <div className="flex flex-wrap items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800 gap-2">
              <div className="flex items-center gap-2">
                <Landmark className="w-5 h-5 text-emerald-600" />
                <div>
                  <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                    <span>{selectedDoc.docNo}</span>
                    <span
                      className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${
                        selectedDoc.status === "posted"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300"
                          : "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300"
                      }`}
                    >
                      {selectedDoc.status === "posted" ? "ผ่านรายการ GL แล้ว" : "แบบร่าง (Draft)"}
                    </span>
                  </h2>
                  <p className="text-xs text-slate-500">{selectedDoc.bankName} - {selectedDoc.accountNo}</p>
                </div>
              </div>

              {selectedDoc.status !== "posted" && (
                <Button
                  size="sm"
                  onClick={handlePostGL}
                  className="text-xs h-7 gap-1 bg-indigo-600 hover:bg-indigo-700 text-white"
                >
                  <FileCheck className="w-3.5 h-3.5" />
                  <span>{isTh ? "ผ่านรายการ GL" : "Post to GL"}</span>
                </Button>
              )}
            </div>

            {/* Form Fields Grid */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "วันที่เอกสาร:" : "Doc Date:"}</label>
                <Input
                  type="date"
                  value={editDate}
                  onChange={(e) => setEditDate(e.target.value)}
                  className="h-8 text-xs font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "เลขที่อ้างอิงธนาคาร:" : "Bank Ref No:"}</label>
                <Input
                  value={editRefNo}
                  onChange={(e) => setEditRefNo(e.target.value)}
                  placeholder="เช่น สลิปค่าธรรมเนียม"
                  className="h-8 text-xs font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "บัญชีเงินฝาก:" : "Bank Book:"}</label>
                <Input
                  value={`${selectedDoc.bankName} (${selectedDoc.bankBookCode})`}
                  readOnly
                  className="h-8 text-xs bg-slate-100 dark:bg-slate-800 text-slate-500"
                />
              </div>

              <div className="md:col-span-2 space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "ผังบัญชีแยกประเภท (GL):" : "GL Account:"}</label>
                <Input
                  value={`${selectedDoc.glAccountCode} - ${selectedDoc.glAccountName}`}
                  readOnly
                  className="h-8 text-xs font-mono bg-slate-100 dark:bg-slate-800"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "อัตราภาษีมูลค่าเพิ่ม (VAT):" : "VAT Rate:"}</label>
                <select
                  value={editVatRate}
                  onChange={(e) => setEditVatRate(Number(e.target.value))}
                  className="w-full h-8 px-2 text-xs rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900"
                >
                  <option value={0}>{isTh ? "ไม่มี VAT (0%)" : "No VAT (0%)"}</option>
                  <option value={7}>{isTh ? "VAT 7% (มีใบกำกับภาษี)" : "VAT 7%"}</option>
                </select>
              </div>

              <div className="md:col-span-3 space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "รายละเอียด / วัตถุประสงค์:" : "Remark:"}</label>
                <Input
                  value={editRemark}
                  onChange={(e) => setEditRemark(e.target.value)}
                  placeholder="เช่น ค่าธรรมเนียมสั่งซื้อสมุดเช็ค"
                  className="h-8 text-xs"
                />
              </div>
            </div>
          </div>

          {/* Card 2: Calculation Breakdown & Thai Baht Text */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-4">
            <h3 className="text-xs font-bold text-slate-900 dark:text-white pb-2 border-b border-slate-100 dark:border-slate-800">
              {isTh ? "สรุปยอดเงินและภาษี (Amounts & VAT Calculation)" : "Financial Calculation"}
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "ยอดเงินก่อนภาษี (฿):" : "Amount (Gross):"}</label>
                <Input
                  type="number"
                  step="0.01"
                  value={editAmount}
                  onChange={(e) => setEditAmount(e.target.value)}
                  className="h-8 text-xs font-mono font-bold"
                />
              </div>

              <div className="space-y-1">
                <label className="font-semibold text-slate-700 dark:text-slate-300">{isTh ? "ภาษีมูลค่าเพิ่ม (VAT):" : "VAT Amount:"}</label>
                <Input
                  value={`฿${calcVat}`}
                  readOnly
                  className="h-8 text-xs font-mono bg-slate-100 dark:bg-slate-800"
                />
              </div>

              <div className="space-y-1">
                <label className="font-bold text-emerald-800 dark:text-emerald-300">{isTh ? "ยอดสุทธิ (Net Total):" : "Net Total:"}</label>
                <Input
                  value={`฿${calcNet}`}
                  readOnly
                  className="h-8 text-xs font-mono font-bold text-emerald-600 bg-emerald-50/50 dark:bg-emerald-950/50 border-emerald-300"
                />
              </div>
            </div>

            <div className="p-2.5 bg-slate-50 dark:bg-slate-950 rounded-lg border border-slate-200 dark:border-slate-800 text-xs flex items-center justify-between">
              <span className="text-slate-500 font-semibold">{isTh ? "ตัวอักษรไทย:" : "Thai Baht Text:"}</span>
              <span className="font-medium text-slate-800 dark:text-slate-200">({thaiText})</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
