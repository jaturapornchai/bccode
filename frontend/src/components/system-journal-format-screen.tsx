"use client";

import { useMemo, useRef, useState } from "react";
import {
  CheckCircle2,
  FileCheck,
  FileCog,
  Hash,
  HelpCircle,
  RefreshCw,
  RotateCcw,
  Save,
  Sliders,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export interface JournalBookFormat {
  id: string;
  bookCode: string;
  bookName: string;
  prefix: string;
  datePattern: "YYYYMM" | "YYMM" | "BE-YYYYMM" | "BE-YYMM" | "YYYY" | "NONE";
  separator: "-" | "/" | "." | "";
  runningDigits: 4 | 5 | 6;
  resetPolicy: "monthly" | "yearly" | "never";
  autoGenerate: boolean;
  allowManualOverride: boolean;
  lastSequence: number;
}

const DEFAULT_FORMATS: JournalBookFormat[] = [
  {
    id: "book-gj",
    bookCode: "GJ",
    bookName: "สมุดรายวันทั่วไป (General Journal)",
    prefix: "GJ",
    datePattern: "YYYYMM",
    separator: "-",
    runningDigits: 4,
    resetPolicy: "monthly",
    autoGenerate: true,
    allowManualOverride: true,
    lastSequence: 12,
  },
  {
    id: "book-cr",
    bookCode: "CR",
    bookName: "สมุดรายวันรับเงิน (Cash Receipts Journal)",
    prefix: "CR",
    datePattern: "YYYYMM",
    separator: "-",
    runningDigits: 4,
    resetPolicy: "monthly",
    autoGenerate: true,
    allowManualOverride: true,
    lastSequence: 45,
  },
  {
    id: "book-cp",
    bookCode: "CP",
    bookName: "สมุดรายวันจ่ายเงิน (Cash Payments Journal)",
    prefix: "CP",
    datePattern: "YYYYMM",
    separator: "-",
    runningDigits: 4,
    resetPolicy: "monthly",
    autoGenerate: true,
    allowManualOverride: true,
    lastSequence: 38,
  },
  {
    id: "book-pj",
    bookCode: "PJ",
    bookName: "สมุดรายวันซื้อ (Purchase Journal)",
    prefix: "PJ",
    datePattern: "YYYYMM",
    separator: "-",
    runningDigits: 5,
    resetPolicy: "monthly",
    autoGenerate: true,
    allowManualOverride: true,
    lastSequence: 64,
  },
  {
    id: "book-sj",
    bookCode: "SJ",
    bookName: "สมุดรายวันขาย (Sales Journal)",
    prefix: "SJ",
    datePattern: "YYYYMM",
    separator: "-",
    runningDigits: 5,
    resetPolicy: "monthly",
    autoGenerate: true,
    allowManualOverride: true,
    lastSequence: 89,
  },
  {
    id: "book-aj",
    bookCode: "AJ",
    bookName: "สมุดรายวันปรับปรุง (Adjusting Journal)",
    prefix: "AJ",
    datePattern: "YYYYMM",
    separator: "-",
    runningDigits: 4,
    resetPolicy: "yearly",
    autoGenerate: true,
    allowManualOverride: true,
    lastSequence: 5,
  },
];

export interface SystemJournalFormatScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemJournalFormatScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: SystemJournalFormatScreenProps) {
  const isTh = language === "th";
  const [formats, setFormats] = useState<JournalBookFormat[]>(DEFAULT_FORMATS);
  const [selectedId, setSelectedId] = useState<string>(DEFAULT_FORMATS[0].id);
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 38,
    min: 25,
    max: 60,
    storageKey: "bcai_split_journal_format",
  });

  const selectedFormat = useMemo(() => {
    return formats.find((f) => f.id === selectedId) || formats[0];
  }, [formats, selectedId]);

  // Generate live sample number based on current date
  const sampleNumber = useMemo(() => {
    const now = new Date();
    const ceYear = now.getFullYear();
    const beYear = ceYear + 543;
    const month = String(now.getMonth() + 1).padStart(2, "0");

    let datePart = "";
    switch (selectedFormat.datePattern) {
      case "YYYYMM":
        datePart = `${ceYear}${month}`;
        break;
      case "YYMM":
        datePart = `${String(ceYear).slice(-2)}${month}`;
        break;
      case "BE-YYYYMM":
        datePart = `${beYear}${month}`;
        break;
      case "BE-YYMM":
        datePart = `${String(beYear).slice(-2)}${month}`;
        break;
      case "YYYY":
        datePart = `${ceYear}`;
        break;
      case "NONE":
        datePart = "";
        break;
    }

    const nextSeq = selectedFormat.lastSequence + 1;
    const seqPart = String(nextSeq).padStart(selectedFormat.runningDigits, "0");

    const parts = [selectedFormat.prefix];
    if (datePart) parts.push(datePart);
    parts.push(seqPart);

    return parts.filter(Boolean).join(selectedFormat.separator);
  }, [selectedFormat]);

  const handleUpdateField = <K extends keyof JournalBookFormat>(
    field: K,
    value: JournalBookFormat[K]
  ) => {
    setFormats((prev) =>
      prev.map((f) => (f.id === selectedFormat.id ? { ...f, [field]: value } : f))
    );
  };

  const handleSave = () => {
    setNotification(
      isTh
        ? `บันทึกการตั้งค่ารูปแบบเลขที่สมุดรายวัน (${selectedFormat.bookCode}) เรียบร้อยแล้ว`
        : `Journal format for (${selectedFormat.bookCode}) saved successfully.`
    );
    setTimeout(() => setNotification(null), 3500);
  };

  const handleResetToDefault = () => {
    const original = DEFAULT_FORMATS.find((d) => d.id === selectedFormat.id);
    if (original) {
      setFormats((prev) => prev.map((f) => (f.id === original.id ? { ...original } : f)));
      setNotification(
        isTh
          ? `คืนค่าเริ่มต้นสำหรับ ${original.bookCode} เรียบร้อย`
          : `Reset format to default for ${original.bookCode}`
      );
      setTimeout(() => setNotification(null), 3000);
    }
  };

  return (
    <div className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 font-sans text-xs">
      {/* Header Bar */}
      <div className="flex-none p-3.5 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-wrap items-center justify-between gap-3 shadow-xs">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-indigo-100 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
            <FileCog className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-semibold text-slate-900 dark:text-slate-100">
                {isTh ? "กำหนดรูปแบบรายวัน" : "Journal Format Configuration"}
              </h1>
              <span className="px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-[11px] text-slate-500 dark:text-slate-400">
              {isTh
                ? "จัดการระบบ › กำหนดรูปแบบเลขที่เอกสารสมุดรายวัน รหัสคำนำหน้า รูปแบบปี/เดือน และจำนวนหลักเลขรัน"
                : "System Settings › Setup document numbering formats, prefixes, and running sequence per journal book"}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={handleResetToDefault}
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
        {/* Left Column: Journal Books List */}
        <div
          style={{ width: `${splitPercent}%` }}
          className="h-full flex flex-col border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900"
        >
          <div className="p-3 border-b border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/50 flex items-center justify-between">
            <span className="font-semibold text-slate-700 dark:text-slate-200">
              {isTh ? "สมุดรายวันทั้งหมด" : "Journal Books"} ({formats.length})
            </span>
            <span className="text-[11px] text-slate-500 font-mono">Select to Edit</span>
          </div>

          <div className="flex-1 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800">
            {formats.map((book) => {
              const isSelected = book.id === selectedFormat.id;
              return (
                <div
                  key={book.id}
                  onClick={() => setSelectedId(book.id)}
                  className={`p-3 cursor-pointer transition-colors ${
                    isSelected
                      ? "bg-indigo-50/80 dark:bg-indigo-950/40 border-l-4 border-indigo-600"
                      : "hover:bg-slate-50 dark:hover:bg-slate-800/40"
                  }`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-mono font-bold text-indigo-600 dark:text-indigo-400 text-sm">
                      {book.bookCode}
                    </span>
                    <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
                      {book.prefix}{book.separator}{book.datePattern === "NONE" ? "" : book.datePattern + book.separator}{"0".repeat(book.runningDigits)}
                    </span>
                  </div>
                  <p className="text-slate-800 dark:text-slate-200 font-medium">{book.bookName}</p>
                  <div className="mt-2 flex items-center justify-between text-[11px] text-slate-500">
                    <span>
                      {isTh ? "ลำดับล่าสุด: " : "Last: "}
                      <strong className="text-slate-700 dark:text-slate-300 font-mono">
                        {book.lastSequence}
                      </strong>
                    </span>
                    <span className="flex items-center gap-1 text-emerald-600 dark:text-emerald-400">
                      <CheckCircle2 className="w-3 h-3" />
                      {book.autoGenerate ? (isTh ? "รันอัตโนมัติ" : "Auto") : (isTh ? "คีย์เอง" : "Manual")}
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

        {/* Right Column: Format Detail & Preview */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="h-full flex flex-col bg-slate-50/50 dark:bg-slate-950 overflow-y-auto p-4 gap-4"
        >
          {/* Live Preview Card */}
          <div className="bg-gradient-to-r from-indigo-900 to-slate-900 text-white p-4 rounded-xl shadow-md border border-indigo-800/60">
            <div className="flex items-center justify-between mb-2">
              <span className="text-indigo-200 font-medium flex items-center gap-1.5 text-xs">
                <FileCheck className="w-4 h-4 text-emerald-400" />
                {isTh ? "ตัวอย่างเลขที่เอกสารจริงถัดไป (Live Preview)" : "Next Document Number Preview"}
              </span>
              <span className="text-[11px] text-indigo-300 font-mono">
                {selectedFormat.bookCode} • {selectedFormat.bookName.split(" ")[0]}
              </span>
            </div>
            <div className="py-2">
              <div className="text-2xl sm:text-3xl font-mono font-bold tracking-wider text-amber-300">
                {sampleNumber}
              </div>
            </div>
            <div className="mt-2 pt-2 border-t border-indigo-800/60 flex items-center justify-between text-[11px] text-indigo-200">
              <span>{isTh ? "ลำดับเลขที่ถัดไป: " : "Next Sequence: "} #{selectedFormat.lastSequence + 1}</span>
              <span>{isTh ? "รีเซ็ตเลขรัน: " : "Reset: "} {selectedFormat.resetPolicy === "monthly" ? "ทุกเดือน" : selectedFormat.resetPolicy === "yearly" ? "ทุกปี" : "ไม่รีเซ็ต"}</span>
            </div>
          </div>

          {/* Configuration Form Card */}
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-4 shadow-xs">
            <div className="flex items-center gap-2 pb-3 mb-4 border-b border-slate-200 dark:border-slate-800">
              <Sliders className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
              <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
                {isTh ? "โครงสร้างเลขที่เอกสาร" : "Document Number Structure"}
              </h2>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {/* Prefix */}
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "รหัสคำนำหน้า (Prefix Code)" : "Prefix Code"}
                </label>
                <Input
                  value={selectedFormat.prefix}
                  onChange={(e) => handleUpdateField("prefix", e.target.value.toUpperCase())}
                  className="font-mono uppercase font-bold text-indigo-600 dark:text-indigo-400"
                  maxLength={6}
                />
                <p className="mt-1 text-[10px] text-slate-400">
                  {isTh ? "เช่น GJ, CR, CP, PJ, SJ (สูงสุด 6 ตัวอักษร)" : "e.g. GJ, CR, CP (max 6 chars)"}
                </p>
              </div>

              {/* Separator */}
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "เครื่องหมายคั่น (Separator)" : "Separator Symbol"}
                </label>
                <select
                  value={selectedFormat.separator}
                  onChange={(e) => handleUpdateField("separator", e.target.value as JournalBookFormat["separator"])}
                  className="w-full h-8 px-2.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-mono"
                >
                  <option value="-">- (ขีดกลาง / Hyphen)</option>
                  <option value="/">/ (ทับ / Slash)</option>
                  <option value=".">. (จุด / Dot)</option>
                  <option value="">ไม่มีเครื่องหมายคั่น (None)</option>
                </select>
                <p className="mt-1 text-[10px] text-slate-400">
                  {isTh ? "ตัวคั่นระหว่างรหัส วันที่ และเลขรัน" : "Delimiter between prefix, date, and sequence"}
                </p>
              </div>

              {/* Date Pattern */}
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "รูปแบบปีและเดือน (Date Pattern)" : "Date Pattern"}
                </label>
                <select
                  value={selectedFormat.datePattern}
                  onChange={(e) =>
                    handleUpdateField("datePattern", e.target.value as JournalBookFormat["datePattern"])
                  }
                  className="w-full h-8 px-2.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-mono"
                >
                  <option value="YYYYMM">YYYYMM (ค.ศ. 4 หลัก + เดือน 2 หลัก เช่น 202610)</option>
                  <option value="YYMM">YYMM (ค.ศ. 2 หลัก + เดือน 2 หลัก เช่น 2610)</option>
                  <option value="BE-YYYYMM">BE-YYYYMM (พ.ศ. 4 หลัก + เดือน 2 หลัก เช่น 256910)</option>
                  <option value="BE-YYMM">BE-YYMM (พ.ศ. 2 หลัก + เดือน 2 หลัก เช่น 6910)</option>
                  <option value="YYYY">YYYY (เฉพาะปี ค.ศ. เช่น 2026)</option>
                  <option value="NONE">NONE (ไม่มีวันที่ในเลขที่)</option>
                </select>
              </div>

              {/* Running Digits */}
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "จำนวนหลักเลขรัน (Running Digits)" : "Running Digits"}
                </label>
                <select
                  value={selectedFormat.runningDigits}
                  onChange={(e) =>
                    handleUpdateField("runningDigits", Number(e.target.value) as JournalBookFormat["runningDigits"])
                  }
                  className="w-full h-8 px-2.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-mono"
                >
                  <option value={4}>4 หลัก (0001 - 9999)</option>
                  <option value={5}>5 หลัก (00001 - 99999)</option>
                  <option value={6}>6 หลัก (000001 - 999999)</option>
                </select>
              </div>

              {/* Reset Frequency */}
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "การเริ่มนับเลขใหม่ (Reset Frequency)" : "Sequence Reset Policy"}
                </label>
                <select
                  value={selectedFormat.resetPolicy}
                  onChange={(e) =>
                    handleUpdateField("resetPolicy", e.target.value as JournalBookFormat["resetPolicy"])
                  }
                  className="w-full h-8 px-2.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs"
                >
                  <option value="monthly">{isTh ? "เริ่มนับ 1 ใหม่ทุกสิ้นเดือน (Monthly Reset)" : "Monthly Reset"}</option>
                  <option value="yearly">{isTh ? "เริ่มนับ 1 ใหม่ทุกสิ้นปีบัญชี (Yearly Reset)" : "Yearly Reset"}</option>
                  <option value="never">{isTh ? "รันเลขต่อเนื่องไม่รีเซ็ต (Continuous)" : "Continuous"}</option>
                </select>
              </div>

              {/* Last Sequence */}
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ลำดับล่าสุดในงวดปัจจุบัน" : "Current Last Sequence"}
                </label>
                <Input
                  type="number"
                  min={0}
                  value={selectedFormat.lastSequence}
                  onChange={(e) => handleUpdateField("lastSequence", Math.max(0, parseInt(e.target.value) || 0))}
                  className="font-mono text-xs"
                />
              </div>
            </div>

            {/* Checkbox Options */}
            <div className="mt-4 pt-4 border-t border-slate-100 dark:border-slate-800 space-y-2.5">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedFormat.autoGenerate}
                  onChange={(e) => handleUpdateField("autoGenerate", e.target.checked)}
                  className="rounded text-indigo-600 focus:ring-indigo-500 w-4 h-4"
                />
                <span className="text-slate-700 dark:text-slate-300">
                  {isTh
                    ? "สร้างเลขที่เอกสารอัตโนมัติเมื่อเริ่มบันทึกใบสำคัญใหม่ (Auto Number Generation)"
                    : "Automatically assign document number on new voucher creation"}
                </span>
              </label>

              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedFormat.allowManualOverride}
                  onChange={(e) => handleUpdateField("allowManualOverride", e.target.checked)}
                  className="rounded text-indigo-600 focus:ring-indigo-500 w-4 h-4"
                />
                <span className="text-slate-700 dark:text-slate-300">
                  {isTh
                    ? "อนุญาตให้ผู้ใช้ที่มีสิทธิ์พิมพ์แก้เลขที่เอกสารเองได้ (Allow Manual Override)"
                    : "Allow authorized users to manually override voucher numbers"}
                </span>
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
