"use client";

import { useMemo, useRef, useState } from "react";
import {
  ArrowDownLeft,
  ArrowUpRight,
  CheckCircle2,
  CreditCard,
  History,
  Printer,
  RefreshCw,
  Search,
  XCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { thaiBahtText } from "@/lib/thai-baht-text";
import { type LanguageCode } from "@/lib/i18n";

export type ChequeDirection = "received" | "issued";
export type ChequeStatus = "on_hand" | "deposited" | "cleared" | "returned" | "cancelled";

export interface ChequeAuditEntry {
  id: string;
  timestamp: string;
  status: ChequeStatus;
  statusTh: string;
  note: string;
  updatedBy: string;
  referenceDoc?: string;
}

export interface ChequeMovementRecord {
  id: string;
  direction: ChequeDirection;
  chequeNo: string;
  bankCode: string;
  bankName: string;
  bankBranch: string;
  chequeDate: string;
  dueDate: string;
  contactCode: string;
  contactName: string;
  amount: string;
  currentStatus: ChequeStatus;
  depositDate?: string;
  clearedDate?: string;
  returnedDate?: string;
  returnReason?: string;
  associatedAccount: string;
  auditTrail: ChequeAuditEntry[];
}

const INITIAL_CHEQUES: ChequeMovementRecord[] = [
  {
    id: "chq-01",
    direction: "received",
    chequeNo: "40891234",
    bankCode: "KBANK",
    bankName: "ธนาคารกสิกรไทย",
    bankBranch: "สาขาสยามสแควร์",
    chequeDate: "2026-10-01",
    dueDate: "2026-10-15",
    contactCode: "AR-001",
    contactName: "บริษัท สยามคอนกรีตมิกซ์ จำกัด",
    amount: "77040.00",
    currentStatus: "deposited",
    depositDate: "2026-10-08",
    associatedAccount: "KBANK 045-1-88992-1 กระแสรายวัน",
    auditTrail: [
      {
        id: "aud-01",
        timestamp: "2026-10-01 10:30",
        status: "on_hand",
        statusTh: "รับเช็คในมือ (รอขึ้นเงิน)",
        note: "รับชำระตามใบเสร็จรับเงิน RC6910-0012",
        updatedBy: "นางสาวกาญจนา การเงินดี",
        referenceDoc: "RC6910-0012",
      },
      {
        id: "aud-02",
        timestamp: "2026-10-08 14:00",
        status: "deposited",
        statusTh: "นำฝากธนาคาร (รอเคลียริ่ง)",
        note: "นำฝากเข้าบัญชีกสิกรไทย สลิปฝาก #DEP6910-004",
        updatedBy: "นายสุวิทย์ บัญชีทุน",
        referenceDoc: "DEP6910-004",
      },
    ],
  },
  {
    id: "chq-02",
    direction: "issued",
    chequeNo: "90218845",
    bankCode: "BBL",
    bankName: "ธนาคารกรุงเทพ",
    bankBranch: "สาขาสีลม",
    chequeDate: "2026-10-05",
    dueDate: "2026-10-20",
    contactCode: "AP-002",
    contactName: "บริษัท สยามสตีลอินดัสตรี จำกัด (มหาชน)",
    amount: "209880.00",
    currentStatus: "on_hand",
    associatedAccount: "BBL 120-3-04987-5 กระแสรายวัน",
    auditTrail: [
      {
        id: "aud-03",
        timestamp: "2026-10-05 16:45",
        status: "on_hand",
        statusTh: "ออกเช็คแล้ว (รอส่งมอบ/ขึ้นเงิน)",
        note: "ออกเช็คเตรียมจ่ายชำระหนี้ค่าเหล็กเส้น PVP6910-0002",
        updatedBy: "นางสาวกาญจนา การเงินดี",
        referenceDoc: "PVP6910-0002",
      },
    ],
  },
  {
    id: "chq-03",
    direction: "received",
    chequeNo: "55120987",
    bankCode: "SCB",
    bankName: "ธนาคารไทยพาณิชย์",
    bankBranch: "สาขารัชโยธิน",
    chequeDate: "2026-09-25",
    dueDate: "2026-09-30",
    contactCode: "AR-002",
    contactName: "ห้างหุ้นส่วนจำกัด นครหลวงวิศวกรรม",
    amount: "125000.00",
    currentStatus: "cleared",
    depositDate: "2026-09-28",
    clearedDate: "2026-09-30",
    associatedAccount: "SCB 111-2-99881-0 ออมทรัพย์",
    auditTrail: [
      {
        id: "aud-04",
        timestamp: "2026-09-25 11:00",
        status: "on_hand",
        statusTh: "รับเช็คในมือ",
        note: "รับชำระบิลค่าวัสดุก่อสร้าง",
        updatedBy: "นางสาวกาญจนา การเงินดี",
      },
      {
        id: "aud-05",
        timestamp: "2026-09-28 10:15",
        status: "deposited",
        statusTh: "นำฝากธนาคาร",
        note: "นำฝาก SCB สาขารัชโยธิน",
        updatedBy: "นายสุวิทย์ บัญชีทุน",
      },
      {
        id: "aud-06",
        timestamp: "2026-09-30 17:00",
        status: "cleared",
        statusTh: "ผ่านเช็คเรียบร้อย (เงินเข้า)",
        note: "ยอดเงินเข้าบัญชีตาม Bank Statement 30 ก.ย. 69",
        updatedBy: "นายบัญชา บริหารกิจ (สมุห์บัญชี)",
      },
    ],
  },
  {
    id: "chq-04",
    direction: "received",
    chequeNo: "11029384",
    bankCode: "KTB",
    bankName: "ธนาคารกรุงไทย",
    bankBranch: "สาขาอโศก",
    chequeDate: "2026-09-20",
    dueDate: "2026-09-25",
    contactCode: "AR-003",
    contactName: "บริษัท ธนบุรีอินดัสเตรียล จำกัด",
    amount: "45000.00",
    currentStatus: "returned",
    depositDate: "2026-09-25",
    returnedDate: "2026-09-27",
    returnReason: "เงินในบัญชีของผู้สั่งจ่ายไม่พอจ่าย (Insufficient Funds)",
    associatedAccount: "KTB 002-1-55443-2 กระแสรายวัน",
    auditTrail: [
      {
        id: "aud-07",
        timestamp: "2026-09-20 09:30",
        status: "on_hand",
        statusTh: "รับเช็คในมือ",
        note: "รับชำระค่างวด",
        updatedBy: "นางสาวกาญจนา การเงินดี",
      },
      {
        id: "aud-08",
        timestamp: "2026-09-25 11:20",
        status: "deposited",
        statusTh: "นำฝากธนาคาร",
        note: "นำฝาก KTB",
        updatedBy: "นายสุวิทย์ บัญชีทุน",
      },
      {
        id: "aud-09",
        timestamp: "2026-09-27 15:30",
        status: "returned",
        statusTh: "เช็คคืน (ธนาคารปฏิเสธการจ่ายเงิน)",
        note: "ยอดเงินในบัญชีไม่พอจ่าย ประสานงานฝ่ายขายเพื่อขอเปลี่ยนเป็นเงินสด",
        updatedBy: "นายบัญชา บริหารกิจ (สมุห์บัญชี)",
      },
    ],
  },
];

export interface ChequeMovementScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function ChequeMovementScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: ChequeMovementScreenProps) {
  const isTh = language === "th";
  const [cheques, setCheques] = useState<ChequeMovementRecord[]>(INITIAL_CHEQUES);
  const [selectedId, setSelectedId] = useState<string>(INITIAL_CHEQUES[0].id);
  const [directionFilter, setDirectionFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 40,
    min: 28,
    max: 60,
    storageKey: "bcai_split_cheque_movement",
  });

  const selectedCheque = useMemo(() => {
    return cheques.find((c) => c.id === selectedId) || cheques[0];
  }, [cheques, selectedId]);

  const filteredCheques = useMemo(() => {
    return cheques.filter((c) => {
      const matchDir = directionFilter === "all" || c.direction === directionFilter;
      const matchStatus = statusFilter === "all" || c.currentStatus === statusFilter;
      const q = searchQuery.toLowerCase().trim();
      const matchSearch =
        !q ||
        c.chequeNo.toLowerCase().includes(q) ||
        c.contactName.toLowerCase().includes(q) ||
        c.bankName.toLowerCase().includes(q) ||
        c.bankCode.toLowerCase().includes(q);
      return matchDir && matchStatus && matchSearch;
    });
  }, [cheques, directionFilter, statusFilter, searchQuery]);

  // Status transition handlers
  const handleUpdateStatus = (newStatus: ChequeStatus, statusTh: string, defaultNote: string) => {
    const now = new Date();
    const timestamp = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")} ${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`;

    const newAudit: ChequeAuditEntry = {
      id: `aud-${Date.now()}`,
      timestamp,
      status: newStatus,
      statusTh,
      note: defaultNote,
      updatedBy: "ผู้ใช้งานปัจจุบัน / ลุงจืด",
    };

    setCheques((prev) =>
      prev.map((c) => {
        if (c.id === selectedCheque.id) {
          return {
            ...c,
            currentStatus: newStatus,
            depositDate: newStatus === "deposited" ? timestamp.slice(0, 10) : c.depositDate,
            clearedDate: newStatus === "cleared" ? timestamp.slice(0, 10) : c.clearedDate,
            returnedDate: newStatus === "returned" ? timestamp.slice(0, 10) : c.returnedDate,
            auditTrail: [newAudit, ...c.auditTrail],
          };
        }
        return c;
      })
    );

    setNotification(
      isTh
        ? `ปรับปรุงสถานะเช็ค ${selectedCheque.chequeNo} เป็น "${statusTh}" เรียบร้อยแล้ว`
        : `Updated cheque ${selectedCheque.chequeNo} status to ${newStatus}`
    );
  };

  const thaiText = useMemo(() => {
    return thaiBahtText(selectedCheque.amount);
  }, [selectedCheque.amount]);

  const totalFilteredAmount = useMemo(() => {
    return filteredCheques.reduce((acc, c) => acc + Number(c.amount), 0);
  }, [filteredCheques]);

  return (
    <div
      ref={containerRef}
      className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 text-slate-800 dark:text-slate-100 overflow-hidden font-sans"
    >
      {/* Top Header Bar */}
      <div className="flex flex-wrap items-center justify-between px-4 py-2.5 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 gap-3">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 border border-indigo-200/60 dark:border-indigo-800/40">
            <CreditCard className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold tracking-tight text-slate-900 dark:text-white">
                {isTh ? "ข้อมูลการเคลื่อนไหวเช็ค (Cheque Movement Inquiry)" : "Cheque Movement Inquiry"}
              </h1>
              <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800">
                Champ Parity 60000+
              </span>
            </div>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {isTh
                ? "ติดตามสถานะเช็ครับ/เช็คจ่าย การนำฝาก การผ่านเช็ค เช็คคืน และประวัติการเคลื่อนไหวทุกขั้นตอน"
                : "Track cheque lifecycle: receipt, deposit, clearing, return & audit timeline"}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setCheques(INITIAL_CHEQUES)}
            className="text-xs h-8 gap-1.5"
            title={isTh ? "รีเซ็ตข้อมูลตัวอย่าง" : "Reset sample"}
          >
            <RefreshCw className="w-3.5 h-3.5" />
            <span>{isTh ? "คืนค่าเดิม" : "Reset"}</span>
          </Button>
          <Button
            size="sm"
            onClick={() => window.print()}
            className="text-xs h-8 gap-1.5 bg-indigo-600 hover:bg-indigo-700 text-white font-medium shadow-sm"
          >
            <Printer className="w-3.5 h-3.5" />
            <span>{isTh ? "พิมพ์รายงานเช็ค" : "Print Report"}</span>
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
        {/* Left Pane: Filter & Cheque List */}
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
                placeholder={isTh ? "ค้นหาเลขที่เช็ค / ชื่อบริษัท / ธนาคาร..." : "Search cheque, party or bank..."}
                className="pl-8 h-8 text-xs bg-white dark:bg-slate-950"
              />
            </div>

            {/* Direction Filter Tabs (รับ / จ่าย / ทั้งหมด) */}
            <div className="grid grid-cols-3 gap-1 bg-slate-100 dark:bg-slate-800 p-0.5 rounded-lg text-xs">
              <button
                type="button"
                onClick={() => setDirectionFilter("all")}
                className={`py-1 rounded font-medium transition-all ${
                  directionFilter === "all"
                    ? "bg-white dark:bg-slate-900 text-slate-900 dark:text-white shadow-sm"
                    : "text-slate-500 hover:text-slate-900"
                }`}
              >
                {isTh ? "ทั้งหมด" : "All"}
              </button>
              <button
                type="button"
                onClick={() => setDirectionFilter("received")}
                className={`py-1 rounded font-medium transition-all flex items-center justify-center gap-1 ${
                  directionFilter === "received"
                    ? "bg-emerald-600 text-white shadow-sm font-semibold"
                    : "text-slate-500 hover:text-emerald-600"
                }`}
              >
                <ArrowDownLeft className="w-3 h-3" />
                <span>{isTh ? "เช็ครับ" : "Cheques In"}</span>
              </button>
              <button
                type="button"
                onClick={() => setDirectionFilter("issued")}
                className={`py-1 rounded font-medium transition-all flex items-center justify-center gap-1 ${
                  directionFilter === "issued"
                    ? "bg-rose-600 text-white shadow-sm font-semibold"
                    : "text-slate-500 hover:text-rose-600"
                }`}
              >
                <ArrowUpRight className="w-3 h-3" />
                <span>{isTh ? "เช็คจ่าย" : "Cheques Out"}</span>
              </button>
            </div>

            {/* Status Pills */}
            <div className="flex flex-wrap gap-1 items-center">
              {[
                { id: "all", th: "ทุกสถานะ", en: "All" },
                { id: "on_hand", th: "ในมือ", en: "On Hand" },
                { id: "deposited", th: "นำฝาก", en: "Deposited" },
                { id: "cleared", th: "ผ่านเช็ค", en: "Cleared" },
                { id: "returned", th: "เช็คคืน", en: "Returned" },
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

          {/* Cheque List Items */}
          <div className="flex-1 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800/60 p-1.5 space-y-1">
            {filteredCheques.map((c) => {
              const isSelected = c.id === selectedCheque.id;
              const isRecv = c.direction === "received";
              return (
                <div
                  key={c.id}
                  onClick={() => setSelectedId(c.id)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") setSelectedId(c.id);
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
                    <div className="flex items-center gap-1.5">
                      <span
                        className={`text-[10px] px-1.5 py-0.5 rounded font-bold ${
                          isRecv
                            ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/70 dark:text-emerald-300"
                            : "bg-rose-100 text-rose-800 dark:bg-rose-950/70 dark:text-rose-300"
                        }`}
                      >
                        {isRecv ? (isTh ? "เช็ครับ" : "IN") : (isTh ? "เช็คจ่าย" : "OUT")}
                      </span>
                      <span className="font-mono text-xs font-bold text-slate-900 dark:text-slate-100">
                        #{c.chequeNo}
                      </span>
                    </div>

                    <span
                      className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${
                        c.currentStatus === "cleared"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-300/50"
                          : c.currentStatus === "deposited"
                          ? "bg-blue-100 text-blue-800 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-300/50"
                          : c.currentStatus === "returned"
                          ? "bg-rose-100 text-rose-800 dark:bg-rose-950/80 dark:text-rose-300 border border-rose-300/50"
                          : "bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400"
                      }`}
                    >
                      {c.currentStatus === "cleared"
                        ? isTh ? "ผ่านเช็คแล้ว" : "Cleared"
                        : c.currentStatus === "deposited"
                        ? isTh ? "นำฝากแล้ว" : "Deposited"
                        : c.currentStatus === "returned"
                        ? isTh ? "เช็คคืน" : "Returned"
                        : isTh ? "ในมือ/รอขึ้นเงิน" : "On Hand"}
                    </span>
                  </div>

                  <p className="text-xs font-semibold text-slate-800 dark:text-slate-200 mt-1 truncate">
                    {c.contactName}
                  </p>

                  <div className="flex items-center justify-between mt-1 text-[11px] text-slate-500">
                    <span className="truncate max-w-[140px]">{c.bankName}</span>
                    <span className="font-mono font-bold text-indigo-600 dark:text-indigo-400">
                      ฿{Number(c.amount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>

          {/* List Footer Total */}
          <div className="p-3 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 flex justify-between items-center">
            <span>{isTh ? `รวม ${filteredCheques.length} ฉบับ` : `${filteredCheques.length} cheques`}</span>
            <span className="font-mono font-bold text-slate-800 dark:text-slate-200">
              ฿{totalFilteredAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </span>
          </div>
        </div>

        {/* Resizable Splitter */}
        <ResizableSplitter
          isResizing={isResizing}
          onMouseDown={startResize}
          onKeyDown={adjustWithKeyboard}
        />

        {/* Right Pane: Cheque Details & Timeline */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="flex flex-col h-full bg-slate-50 dark:bg-slate-950 overflow-y-auto p-4 space-y-4"
        >
          {/* Card 1: Cheque Master Card */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-4">
            <div className="flex flex-wrap items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800 gap-2">
              <div className="flex items-center gap-2.5">
                <div
                  className={`p-2 rounded-lg ${
                    selectedCheque.direction === "received"
                      ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300"
                      : "bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300"
                  }`}
                >
                  <CreditCard className="w-5 h-5" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h2 className="text-base font-bold text-slate-900 dark:text-white font-mono">
                      เช็คเลขที่ {selectedCheque.chequeNo}
                    </h2>
                    <span
                      className={`text-xs px-2 py-0.5 rounded font-bold ${
                        selectedCheque.direction === "received"
                          ? "bg-emerald-50 text-emerald-700 border border-emerald-200"
                          : "bg-rose-50 text-rose-700 border border-rose-200"
                      }`}
                    >
                      {selectedCheque.direction === "received" ? "เช็ครับ (Cheque In)" : "เช็คจ่าย (Cheque Out)"}
                    </span>
                  </div>
                  <p className="text-xs text-slate-500">{selectedCheque.bankName} ({selectedCheque.bankBranch})</p>
                </div>
              </div>

              {/* Status Action Buttons */}
              <div className="flex items-center gap-1.5 flex-wrap">
                {selectedCheque.currentStatus !== "cleared" && (
                  <Button
                    size="sm"
                    onClick={() => handleUpdateStatus("cleared", "ผ่านเช็คเรียบร้อย", "เงินเข้า/ตัดบัญชีตรงกับ Statement")}
                    className="text-xs h-7 gap-1 bg-emerald-600 hover:bg-emerald-700 text-white"
                  >
                    <CheckCircle2 className="w-3.5 h-3.5" />
                    <span>{isTh ? "ผ่านเช็ค" : "Clear"}</span>
                  </Button>
                )}
                {selectedCheque.currentStatus !== "deposited" && selectedCheque.direction === "received" && (
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleUpdateStatus("deposited", "นำฝากธนาคาร", "นำฝากรอเคลียริ่งเข้าบัญชี")}
                    className="text-xs h-7 gap-1"
                  >
                    <ArrowDownLeft className="w-3.5 h-3.5" />
                    <span>{isTh ? "นำฝาก" : "Deposit"}</span>
                  </Button>
                )}
                {selectedCheque.currentStatus !== "returned" && (
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleUpdateStatus("returned", "เช็คคืน", "ธนาคารปฏิเสธการจ่ายเงิน")}
                    className="text-xs h-7 gap-1 text-rose-600 hover:bg-rose-50"
                  >
                    <XCircle className="w-3.5 h-3.5" />
                    <span>{isTh ? "เช็คคืน" : "Return"}</span>
                  </Button>
                )}
              </div>
            </div>

            {/* Information Grid */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
              <div className="p-2.5 bg-slate-50 dark:bg-slate-950 rounded-lg border border-slate-100 dark:border-slate-800">
                <span className="text-[11px] text-slate-400">{isTh ? "ผู้สั่งจ่าย / ผู้รับเงิน:" : "Party:"}</span>
                <p className="font-bold text-slate-800 dark:text-slate-200 mt-0.5 truncate">{selectedCheque.contactName}</p>
                <p className="text-[10px] text-slate-400 font-mono">รหัส: {selectedCheque.contactCode}</p>
              </div>

              <div className="p-2.5 bg-slate-50 dark:bg-slate-950 rounded-lg border border-slate-100 dark:border-slate-800">
                <span className="text-[11px] text-slate-400">{isTh ? "วันที่บนเช็ค / วันครบกำหนด:" : "Cheque / Due Date:"}</span>
                <p className="font-mono font-semibold text-slate-800 dark:text-slate-200 mt-0.5">{selectedCheque.chequeDate}</p>
                <p className="text-[10px] text-indigo-600 dark:text-indigo-400 font-mono">กำหนด: {selectedCheque.dueDate}</p>
              </div>

              <div className="p-2.5 bg-slate-50 dark:bg-slate-950 rounded-lg border border-slate-100 dark:border-slate-800">
                <span className="text-[11px] text-slate-400">{isTh ? "บัญชีเงินฝากที่ผูก:" : "Bank Account:"}</span>
                <p className="font-mono text-slate-800 dark:text-slate-200 mt-0.5 truncate">{selectedCheque.associatedAccount}</p>
              </div>
            </div>

            {/* Financial Amount Banner */}
            <div className="p-3 bg-indigo-50/70 dark:bg-indigo-950/30 rounded-xl border border-indigo-200/80 dark:border-indigo-800/60 flex flex-wrap items-center justify-between gap-3">
              <div>
                <span className="text-[11px] text-indigo-800 dark:text-indigo-300 font-semibold">{isTh ? "จำนวนเงินตามหน้าเช็ค" : "Cheque Amount"}</span>
                <p className="font-mono text-xl font-black text-indigo-700 dark:text-indigo-300">
                  ฿{Number(selectedCheque.amount).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                </p>
              </div>
              <div className="text-right">
                <span className="text-[11px] text-slate-500">{isTh ? "ตัวอักษรไทย (Thai Baht Text)" : "Baht Text"}</span>
                <p className="text-xs font-semibold text-slate-700 dark:text-slate-300">({thaiText})</p>
              </div>
            </div>
          </div>

          {/* Card 2: Cheque Movement Timeline & History */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-4">
            <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
              <div className="flex items-center gap-2">
                <History className="w-4 h-4 text-indigo-500" />
                <h3 className="text-xs font-bold text-slate-900 dark:text-white">
                  {isTh ? "ประวัติการเคลื่อนไหวและลำดับสถานะเช็ค (Movement Timeline)" : "Cheque Lifecycle Timeline"}
                </h3>
              </div>
              <span className="text-[11px] text-slate-400">
                {isTh ? `${selectedCheque.auditTrail.length} ขั้นตอน` : `${selectedCheque.auditTrail.length} stages`}
              </span>
            </div>

            {/* Timeline Steps */}
            <div className="relative pl-6 space-y-4 before:absolute before:left-2 before:top-2 before:bottom-2 before:w-0.5 before:bg-slate-200 dark:before:bg-slate-800">
              {selectedCheque.auditTrail.map((audit) => (
                <div key={audit.id} className="relative">
                  {/* Timeline bullet dot */}
                  <div
                    className={`absolute -left-6 top-1 w-3.5 h-3.5 rounded-full border-2 border-white dark:border-slate-900 ${
                      audit.status === "cleared"
                        ? "bg-emerald-500"
                        : audit.status === "returned"
                        ? "bg-rose-500"
                        : audit.status === "deposited"
                        ? "bg-blue-500"
                        : "bg-slate-400"
                    }`}
                  />

                  <div className="bg-slate-50 dark:bg-slate-950 p-3 rounded-lg border border-slate-100 dark:border-slate-800 space-y-1">
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-xs font-bold text-slate-800 dark:text-slate-200">
                        {audit.statusTh}
                      </span>
                      <span className="text-[11px] font-mono text-slate-400">{audit.timestamp}</span>
                    </div>

                    <p className="text-xs text-slate-600 dark:text-slate-400">{audit.note}</p>

                    <div className="flex items-center justify-between pt-1 border-t border-slate-100 dark:border-slate-800/60 text-[10px] text-slate-400">
                      <span>โดย: {audit.updatedBy}</span>
                      {audit.referenceDoc && (
                        <span className="font-mono text-indigo-600 dark:text-indigo-400">
                          อ้างอิง: {audit.referenceDoc}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
