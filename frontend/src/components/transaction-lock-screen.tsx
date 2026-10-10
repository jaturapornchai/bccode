"use client";

import { useMemo, useRef, useState } from "react";
import {
  AlertTriangle,
  Calendar,
  CheckCircle2,
  Clock,
  History,
  Info,
  Lock,
  RefreshCw,
  Save,
  Search,
  ShieldCheck,
  Unlock,
  UserCheck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export type ModuleLockKey = "all" | "purchase" | "sale" | "ar" | "ap" | "inventory" | "gl";

export interface ModuleLockConfig {
  key: ModuleLockKey;
  nameTh: string;
  nameEn: string;
  descTh: string;
  descEn: string;
  isLocked: boolean;
  lockedThroughDate: string; // YYYY-MM-DD
  lockedBy: string;
  updatedAt: string;
  allowAdminBypass: boolean;
  remark: string;
}

export interface LockAuditRecord {
  id: string;
  timestamp: string;
  moduleKey: ModuleLockKey;
  moduleNameTh: string;
  action: "lock" | "unlock" | "extend_date" | "bypass_toggle";
  performedBy: string;
  previousDate: string;
  newDate: string;
  reason: string;
}

const INITIAL_MODULE_LOCKS: ModuleLockConfig[] = [
  {
    key: "all",
    nameTh: "ล็อกทุกระบบพร้อมกัน (Global Lock)",
    nameEn: "All Modules (Global Lock)",
    descTh: "ป้องกันการสร้าง แก้ไข หรือลบเอกสารทุกประเภทในระบบทั้งหมด",
    descEn: "Prevent create, edit, or delete for all document types across the system",
    isLocked: false,
    lockedThroughDate: "2026-09-30",
    lockedBy: "นายบัญชา บริหารกิจ (สมุห์บัญชี)",
    updatedAt: "2026-10-01 09:30:15",
    allowAdminBypass: true,
    remark: "ล็อกงวดสิ้นไตรมาส 3/2569 ตรวจสอบรายงานภาษีเรียบร้อย",
  },
  {
    key: "purchase",
    nameTh: "ระบบจัดซื้อ (Procurement / PO / AP)",
    nameEn: "Purchase System",
    descTh: "ใบสั่งซื้อ ใบรับสินค้า ใบกำกับภาษีซื้อ ใบลดหนี้/เพิ่มหนี้ซื้อ",
    descEn: "Purchase Orders, Goods Receipts, Tax Invoices, Debit/Credit Notes",
    isLocked: true,
    lockedThroughDate: "2026-09-30",
    lockedBy: "นางสาวสมร งามตา (หัวหน้าแผนกจัดซื้อ)",
    updatedAt: "2026-10-02 11:15:20",
    allowAdminBypass: true,
    remark: "ปิดงวดภาษีซื้อ ภ.พ.30 ประจำเดือนกันยายน 2569",
  },
  {
    key: "sale",
    nameTh: "ระบบจัดขาย (Sales / SO / AR)",
    nameEn: "Sales System",
    descTh: "ใบเสนอราคา ใบสั่งขาย ใบส่งของ ใบกำกับภาษีขาย ใบลดหนี้ขาย",
    descEn: "Quotations, Sales Orders, Delivery Orders, Tax Invoices, Credit Notes",
    isLocked: true,
    lockedThroughDate: "2026-09-30",
    lockedBy: "นายสุรชัย มุ่งมั่น (ผู้จัดการฝ่ายขาย)",
    updatedAt: "2026-10-02 11:20:45",
    allowAdminBypass: true,
    remark: "ปิดงวดภาษีขาย ภ.พ.30 ประจำเดือนกันยายน 2569 ส่งสรรพากรแล้ว",
  },
  {
    key: "ar",
    nameTh: "ระบบลูกหนี้และการรับเงิน (Receivables / AR)",
    nameEn: "Account Receivables & Receipt",
    descTh: "ใบวางบิล ใบเสร็จรับเงิน ใบเสร็จรับเงินชั่วคราว การรับชำระหนี้",
    descEn: "Billing Notes, Official Receipts, Temporary Receipts, Collections",
    isLocked: true,
    lockedThroughDate: "2026-09-30",
    lockedBy: "นายบัญชา บริหารกิจ (สมุห์บัญชี)",
    updatedAt: "2026-10-02 14:00:10",
    allowAdminBypass: false,
    remark: "กระทบยอดเงินฝากธนาคาร (Bank Reconcile) ก.ย. 69 ตรงกับ Statement",
  },
  {
    key: "ap",
    nameTh: "ระบบเจ้าหนี้และการจ่ายเงิน (Payables / AP)",
    nameEn: "Account Payables & Payment",
    descTh: "ใบรับวางบิล ใบสำคัญจ่าย ทะเบียนเช็คจ่าย ภาษีหัก ณ ที่จ่าย ภ.ง.ด.3/53",
    descEn: "Supplier Billing Receipts, Payment Vouchers, Issued Cheques, WHT 3/53",
    isLocked: true,
    lockedThroughDate: "2026-09-30",
    lockedBy: "นายบัญชา บริหารกิจ (สมุห์บัญชี)",
    updatedAt: "2026-10-02 14:15:30",
    allowAdminBypass: false,
    remark: "ยื่น ภ.ง.ด.3, 53 ประจำเดือน ก.ย. 69 ผ่านอินเทอร์เน็ตแล้ว",
  },
  {
    key: "inventory",
    nameTh: "ระบบคลังสินค้าและการควบคุมสต็อก (Inventory / Stock)",
    nameEn: "Inventory & Stock Control",
    descTh: "ใบเบิกสินค้า ใบรับเข้าสินค้า ใบโอนย้ายคลัง และการตรวจนับสินค้า",
    descEn: "Stock Issue, Stock Receive, Inter-warehouse Transfer, Physical Count",
    isLocked: true,
    lockedThroughDate: "2026-09-30",
    lockedBy: "นายเกรียงไกร คลังดี (หัวหน้าคลังสินค้า)",
    updatedAt: "2026-10-01 18:00:00",
    allowAdminBypass: true,
    remark: "ตรวจนับสต็อกสิ้นไตรมาส Q3 เรียบร้อย ยอดคงเหลือตรงบัญชีคุม",
  },
  {
    key: "gl",
    nameTh: "ระบบบัญชีแยกประเภท (General Ledger / GL)",
    nameEn: "General Ledger System",
    descTh: "สมุดรายวันทั่วไป ซื้อ ขาย รับ จ่าย ปิดงวดบัญชี และงบทดลอง",
    descEn: "General Journals, Cash Journals, Closing Entries, Trial Balance",
    isLocked: true,
    lockedThroughDate: "2026-08-31",
    lockedBy: "นายบัญชา บริหารกิจ (สมุห์บัญชี)",
    updatedAt: "2026-09-15 16:30:00",
    allowAdminBypass: false,
    remark: "ปิดงบการเงินประจำเดือนสิงหาคม 2569 อยู่ระหว่างปิดงวดกันยายน",
  },
];

const INITIAL_AUDIT_LOGS: LockAuditRecord[] = [
  {
    id: "log-001",
    timestamp: "2026-10-02 14:15:30",
    moduleKey: "ap",
    moduleNameTh: "ระบบเจ้าหนี้และการจ่ายเงิน",
    action: "lock",
    performedBy: "นายบัญชา บริหารกิจ",
    previousDate: "2026-08-31",
    newDate: "2026-09-30",
    reason: "ยื่น ภ.ง.ด.3, 53 ประจำเดือน ก.ย. 69 ผ่านอินเทอร์เน็ตแล้ว",
  },
  {
    id: "log-002",
    timestamp: "2026-10-02 14:00:10",
    moduleKey: "ar",
    moduleNameTh: "ระบบลูกหนี้และการรับเงิน",
    action: "lock",
    performedBy: "นายบัญชา บริหารกิจ",
    previousDate: "2026-08-31",
    newDate: "2026-09-30",
    reason: "กระทบยอดเงินฝากธนาคาร (Bank Reconcile) ก.ย. 69 เรียบร้อย",
  },
  {
    id: "log-003",
    timestamp: "2026-10-02 11:20:45",
    moduleKey: "sale",
    moduleNameTh: "ระบบจัดขาย",
    action: "lock",
    performedBy: "นายสุรชัย มุ่งมั่น",
    previousDate: "2026-08-31",
    newDate: "2026-09-30",
    reason: "ปิดงวดภาษีขาย ภ.พ.30 ประจำเดือนกันยายน 2569",
  },
  {
    id: "log-004",
    timestamp: "2026-10-02 11:15:20",
    moduleKey: "purchase",
    moduleNameTh: "ระบบจัดซื้อ",
    action: "lock",
    performedBy: "นางสาวสมร งามตา",
    previousDate: "2026-08-31",
    newDate: "2026-09-30",
    reason: "ปิดงวดภาษีซื้อ ภ.พ.30 ประจำเดือนกันยายน 2569",
  },
  {
    id: "log-005",
    timestamp: "2026-10-01 18:00:00",
    moduleKey: "inventory",
    moduleNameTh: "ระบบคลังสินค้า",
    action: "lock",
    performedBy: "นายเกรียงไกร คลังดี",
    previousDate: "2026-08-31",
    newDate: "2026-09-30",
    reason: "ตรวจนับสต็อกสิ้นไตรมาส Q3 เรียบร้อย ยอดคงเหลือตรงบัญชีคุม",
  },
];

export interface TransactionLockScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function TransactionLockScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: TransactionLockScreenProps) {
  const isTh = language === "th";
  const [moduleLocks, setModuleLocks] = useState<ModuleLockConfig[]>(INITIAL_MODULE_LOCKS);
  const [selectedKey, setSelectedKey] = useState<ModuleLockKey>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [auditLogs, setAuditLogs] = useState<LockAuditRecord[]>(INITIAL_AUDIT_LOGS);
  const [notification, setNotification] = useState<{ type: "success" | "warning"; message: string } | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 42,
    min: 28,
    max: 65,
    storageKey: "bcai_split_transaction_lock",
  });

  const selectedModule = useMemo(() => {
    return moduleLocks.find((m) => m.key === selectedKey) || moduleLocks[0];
  }, [moduleLocks, selectedKey]);

  // Form edit state for currently selected module
  const [editLocked, setEditLocked] = useState(selectedModule.isLocked);
  const [editDate, setEditDate] = useState(selectedModule.lockedThroughDate);
  const [editBypass, setEditBypass] = useState(selectedModule.allowAdminBypass);
  const [editRemark, setEditRemark] = useState(selectedModule.remark);

  // Sync edit state when selection changes
  const handleSelectModule = (mod: ModuleLockConfig) => {
    setSelectedKey(mod.key);
    setEditLocked(mod.isLocked);
    setEditDate(mod.lockedThroughDate);
    setEditBypass(mod.allowAdminBypass);
    setEditRemark(mod.remark);
    setNotification(null);
  };

  const filteredModules = useMemo(() => {
    if (!searchQuery.trim()) return moduleLocks;
    const q = searchQuery.toLowerCase().trim();
    return moduleLocks.filter(
      (m) =>
        m.nameTh.toLowerCase().includes(q) ||
        m.nameEn.toLowerCase().includes(q) ||
        m.descTh.toLowerCase().includes(q) ||
        m.remark.toLowerCase().includes(q)
    );
  }, [moduleLocks, searchQuery]);

  const activeLockedCount = useMemo(() => {
    return moduleLocks.filter((m) => m.key !== "all" && m.isLocked).length;
  }, [moduleLocks]);

  // Save changes handler
  const handleSaveChanges = () => {
    const now = new Date();
    const timestamp = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")} ${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}:${String(now.getSeconds()).padStart(2, "0")}`;

    let updatedList = [...moduleLocks];
    const prevDate = selectedModule.lockedThroughDate;

    if (selectedKey === "all") {
      // Global lock applies to all modules
      updatedList = updatedList.map((m) => ({
        ...m,
        isLocked: editLocked,
        lockedThroughDate: editDate,
        allowAdminBypass: editBypass,
        remark: editRemark,
        updatedAt: timestamp,
        lockedBy: "ผู้ดูแลระบบ (Admin) / ลุงจืด",
      }));
    } else {
      updatedList = updatedList.map((m) => {
        if (m.key === selectedKey) {
          return {
            ...m,
            isLocked: editLocked,
            lockedThroughDate: editDate,
            allowAdminBypass: editBypass,
            remark: editRemark,
            updatedAt: timestamp,
            lockedBy: "ผู้ดูแลระบบ (Admin) / ลุงจืด",
          };
        }
        return m;
      });
    }

    setModuleLocks(updatedList);

    // Add Audit Log
    const newLog: LockAuditRecord = {
      id: `log-${Date.now()}`,
      timestamp,
      moduleKey: selectedKey,
      moduleNameTh: selectedModule.nameTh,
      action: editLocked ? (editDate !== prevDate ? "extend_date" : "lock") : "unlock",
      performedBy: "ผู้ดูแลระบบ (Admin) / ลุงจืด",
      previousDate: prevDate,
      newDate: editDate,
      reason: editRemark || (editLocked ? "ปรับปรุงวันล็อกเอกสาร" : "ปลดล็อกเอกสารชั่วคราว"),
    };

    setAuditLogs([newLog, ...auditLogs]);
    setNotification({
      type: "success",
      message: isTh
        ? `บันทึกการตั้งค่าล็อก ${selectedModule.nameTh} เรียบร้อยแล้ว (ถึงวันที่ ${editDate})`
        : `Transaction lock settings saved for ${selectedModule.nameEn} (Through ${editDate})`,
    });
  };

  // Quick preset dates
  const handleSetPresetDate = (daysAgo: number) => {
    const d = new Date();
    d.setDate(d.getDate() - daysAgo);
    const dateStr = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
    setEditDate(dateStr);
    setEditLocked(true);
  };

  const handleSetMonthEnd = () => {
    const now = new Date();
    // End of previous month
    const endPrevMonth = new Date(now.getFullYear(), now.getMonth(), 0);
    const dateStr = `${endPrevMonth.getFullYear()}-${String(endPrevMonth.getMonth() + 1).padStart(2, "0")}-${String(endPrevMonth.getDate()).padStart(2, "0")}`;
    setEditDate(dateStr);
    setEditLocked(true);
  };

  const handleSetToday = () => {
    const now = new Date();
    const dateStr = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
    setEditDate(dateStr);
    setEditLocked(true);
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
            <Lock className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold tracking-tight text-slate-900 dark:text-white">
                {isTh ? "Transaction Lock (กำหนดสิทธิ์และล็อกเอกสารย้อนหลัง)" : "Transaction Lock Manager"}
              </h1>
              <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800">
                Champ Parity 80000+
              </span>
            </div>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {isTh
                ? `ป้องกันการแก้ไข/ลบ/เพิ่มเอกสารย้อนหลังเกินกำหนดเวลา • ล็อกอยู่ ${activeLockedCount} จาก 6 ระบบย่อย`
                : `Prevent backdated transaction modifications • Active locks: ${activeLockedCount} of 6 modules`}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setModuleLocks(INITIAL_MODULE_LOCKS);
              setAuditLogs(INITIAL_AUDIT_LOGS);
              setNotification({
                type: "warning",
                message: isTh ? "รีเซ็ตสถานะเป็นค่าตั้งต้นเรียบร้อย" : "Reset to default state",
              });
            }}
            className="text-xs h-8 gap-1.5"
            title={isTh ? "รีเซ็ตสถานะทั้งหมด" : "Reset to initial"}
          >
            <RefreshCw className="w-3.5 h-3.5" />
            <span>{isTh ? "คืนค่าเดิม" : "Reset"}</span>
          </Button>
          <Button
            size="sm"
            onClick={handleSaveChanges}
            className="text-xs h-8 gap-1.5 bg-indigo-600 hover:bg-indigo-700 text-white font-medium shadow-sm"
          >
            <Save className="w-3.5 h-3.5" />
            <span>{isTh ? "บันทึกการตั้งค่า" : "Save Changes"}</span>
          </Button>
        </div>
      </div>

      {/* Notification Toast Bar if any */}
      {notification && (
        <div
          className={`flex items-center justify-between px-4 py-2 text-xs font-medium border-b ${
            notification.type === "success"
              ? "bg-emerald-50 dark:bg-emerald-950/40 text-emerald-800 dark:text-emerald-200 border-emerald-200 dark:border-emerald-900"
              : "bg-amber-50 dark:bg-amber-950/40 text-amber-800 dark:text-amber-200 border-amber-200 dark:border-amber-900"
          }`}
        >
          <div className="flex items-center gap-2">
            {notification.type === "success" ? (
              <CheckCircle2 className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
            ) : (
              <AlertTriangle className="w-4 h-4 text-amber-600 dark:text-amber-400" />
            )}
            <span>{notification.message}</span>
          </div>
          <button
            onClick={() => setNotification(null)}
            className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
          >
            ✕
          </button>
        </div>
      )}

      {/* Main Split Layout: Left Modules List / Right Lock Settings Form */}
      <div className="flex flex-1 overflow-hidden relative">
        {/* Left Side: Module List & Status */}
        <div
          style={{ width: `${splitPercent}%` }}
          className="flex flex-col h-full bg-white dark:bg-slate-900 border-r border-slate-200 dark:border-slate-800 overflow-hidden"
        >
          {/* Search Box */}
          <div className="p-3 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
            <div className="relative">
              <Search className="w-3.5 h-3.5 absolute left-2.5 top-2.5 text-slate-400" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={isTh ? "ค้นหาโมดูล / ระบบงาน..." : "Filter module..."}
                className="pl-8 h-8 text-xs bg-white dark:bg-slate-950"
              />
            </div>
          </div>

          {/* Module List Items */}
          <div className="flex-1 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800/60 p-1.5 space-y-1">
            {filteredModules.map((mod) => {
              const isSelected = mod.key === selectedKey;
              return (
                <div
                  key={mod.key}
                  onClick={() => handleSelectModule(mod)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      handleSelectModule(mod);
                    }
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
                    <div className="flex items-center gap-2">
                      <div
                        className={`p-1.5 rounded-md ${
                          mod.isLocked
                            ? "bg-rose-100 text-rose-700 dark:bg-rose-950/70 dark:text-rose-300"
                            : "bg-slate-200 text-slate-600 dark:bg-slate-800 dark:text-slate-400"
                        }`}
                      >
                        {mod.isLocked ? <Lock className="w-3.5 h-3.5" /> : <Unlock className="w-3.5 h-3.5" />}
                      </div>
                      <span className="text-xs font-semibold text-slate-900 dark:text-slate-100">
                        {isTh ? mod.nameTh : mod.nameEn}
                      </span>
                    </div>

                    <span
                      className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${
                        mod.isLocked
                          ? "bg-rose-100 text-rose-800 dark:bg-rose-950/80 dark:text-rose-300 border border-rose-300/50"
                          : "bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400"
                      }`}
                    >
                      {mod.isLocked
                        ? isTh
                          ? `ล็อกถึง ${mod.lockedThroughDate}`
                          : `Locked thru ${mod.lockedThroughDate}`
                        : isTh
                        ? "ไม่ได้ล็อก"
                        : "Unlocked"}
                    </span>
                  </div>

                  <p className="text-[11px] text-slate-500 dark:text-slate-400 mt-1 line-clamp-1">
                    {isTh ? mod.descTh : mod.descEn}
                  </p>

                  <div className="flex items-center justify-between mt-2 pt-1.5 border-t border-slate-100 dark:border-slate-800/40 text-[10px] text-slate-400">
                    <span className="truncate max-w-[160px]">{mod.lockedBy}</span>
                    <span>{mod.updatedAt.slice(0, 10)}</span>
                  </div>
                </div>
              );
            })}
          </div>

          {/* Quick Lock Summary Footer */}
          <div className="p-3 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 space-y-1">
            <div className="flex justify-between items-center">
              <span>{isTh ? "สถานะการล็อกรวม:" : "Overall lock:"}</span>
              <span className="font-semibold text-slate-800 dark:text-slate-200">
                {activeLockedCount === 6
                  ? isTh
                    ? "ล็อกครบทุกโมดูล (ปลอดภัย 100%)"
                    : "Fully Locked"
                  : isTh
                  ? `ล็อก ${activeLockedCount} / 6 โมดูล`
                  : `${activeLockedCount}/6 Modules Locked`}
              </span>
            </div>
            <div className="w-full bg-slate-200 dark:bg-slate-800 rounded-full h-1.5 overflow-hidden">
              <div
                className="bg-indigo-600 h-1.5 rounded-full transition-all"
                style={{ width: `${(activeLockedCount / 6) * 100}%` }}
              />
            </div>
          </div>
        </div>

        {/* Resizable Splitter */}
        <ResizableSplitter
          isResizing={isResizing}
          onMouseDown={startResize}
          onKeyDown={adjustWithKeyboard}
        />

        {/* Right Side: Lock Settings Details & Audit Trail */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="flex flex-col h-full bg-slate-50 dark:bg-slate-950 overflow-y-auto p-4 space-y-4"
        >
          {/* Detail Card 1: Current Module Settings */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-4">
            <div className="flex items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800">
              <div className="flex items-center gap-2.5">
                <div
                  className={`p-2 rounded-lg ${
                    editLocked
                      ? "bg-rose-100 text-rose-700 dark:bg-rose-950/80 dark:text-rose-300"
                      : "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300"
                  }`}
                >
                  {editLocked ? <Lock className="w-5 h-5" /> : <Unlock className="w-5 h-5" />}
                </div>
                <div>
                  <h2 className="text-sm font-bold text-slate-900 dark:text-white">
                    {isTh ? selectedModule.nameTh : selectedModule.nameEn}
                  </h2>
                  <p className="text-xs text-slate-500 dark:text-slate-400">
                    {isTh ? selectedModule.descTh : selectedModule.descEn}
                  </p>
                </div>
              </div>

              {selectedKey === "all" && (
                <span className="text-[11px] px-2.5 py-1 bg-amber-100 dark:bg-amber-950/60 text-amber-800 dark:text-amber-300 font-semibold rounded-md border border-amber-300/60">
                  {isTh ? "⚡ มีผลต่อทุกระบบย่อย" : "Global Master Switch"}
                </span>
              )}
            </div>

            {/* Lock Activation Switch */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="p-3 bg-slate-50 dark:bg-slate-950 rounded-lg border border-slate-200 dark:border-slate-800 space-y-2">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300 flex items-center justify-between">
                  <span>{isTh ? "สถานะการล็อกเอกสาร:" : "Lock Activation:"}</span>
                  <span
                    className={`text-xs font-bold ${
                      editLocked ? "text-rose-600 dark:text-rose-400" : "text-slate-500"
                    }`}
                  >
                    {editLocked ? (isTh ? "🔒 เปิดใช้งานล็อก (Locked)" : "Active") : (isTh ? "🔓 ปลดล็อก (Unlocked)" : "Inactive")}
                  </span>
                </label>
                <div className="flex gap-2">
                  <Button
                    type="button"
                    size="sm"
                    variant={editLocked ? "default" : "outline"}
                    onClick={() => setEditLocked(true)}
                    className={`flex-1 text-xs h-8 gap-1.5 ${
                      editLocked ? "bg-rose-600 hover:bg-rose-700 text-white font-semibold" : ""
                    }`}
                  >
                    <Lock className="w-3.5 h-3.5" />
                    <span>{isTh ? "เปิดล็อกเอกสาร" : "Lock"}</span>
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant={!editLocked ? "default" : "outline"}
                    onClick={() => setEditLocked(false)}
                    className={`flex-1 text-xs h-8 gap-1.5 ${
                      !editLocked ? "bg-slate-700 text-white font-semibold" : ""
                    }`}
                  >
                    <Unlock className="w-3.5 h-3.5" />
                    <span>{isTh ? "ปลดล็อก" : "Unlock"}</span>
                  </Button>
                </div>
                <p className="text-[11px] text-slate-500">
                  {isTh
                    ? "หากเปิดล็อก ผู้ใช้งานทั่วไปจะไม่สามารถสร้าง แก้ไข หรือลบเอกสารที่ลงวันที่ก่อนหรือเท่ากับวันล็อกได้"
                    : "When active, regular users cannot add, edit, or delete transactions dated on or before lock date."}
                </p>
              </div>

              {/* Lock Date Picker & Presets */}
              <div className="p-3 bg-slate-50 dark:bg-slate-950 rounded-lg border border-slate-200 dark:border-slate-800 space-y-2">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                  <Calendar className="w-3.5 h-3.5 text-indigo-500" />
                  <span>{isTh ? "ล็อกเอกสารจนถึงวันที่ (Lock Through Date):" : "Lock Through Date:"}</span>
                </label>
                <Input
                  type="date"
                  value={editDate}
                  onChange={(e) => setEditDate(e.target.value)}
                  className="h-8 text-xs bg-white dark:bg-slate-900 font-mono"
                />
                <div className="flex flex-wrap gap-1.5 pt-1">
                  <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    onClick={handleSetMonthEnd}
                    className="text-[10px] h-6 px-2 py-0 bg-slate-200/80 dark:bg-slate-800"
                  >
                    {isTh ? "สิ้นเดือนก่อนหน้า" : "Prev Month-End"}
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    onClick={() => handleSetPresetDate(7)}
                    className="text-[10px] h-6 px-2 py-0 bg-slate-200/80 dark:bg-slate-800"
                  >
                    {isTh ? "ย้อนหลัง 7 วัน" : "7 Days Ago"}
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    onClick={handleSetToday}
                    className="text-[10px] h-6 px-2 py-0 bg-slate-200/80 dark:bg-slate-800"
                  >
                    {isTh ? "ถึงวันปัจจุบัน" : "Through Today"}
                  </Button>
                </div>
              </div>
            </div>

            {/* Admin Bypass Option */}
            <div className="p-3 bg-slate-50/70 dark:bg-slate-950/70 rounded-lg border border-slate-200 dark:border-slate-800 flex items-center justify-between gap-3">
              <div className="flex items-center gap-2.5">
                <ShieldCheck className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
                <div>
                  <h3 className="text-xs font-semibold text-slate-800 dark:text-slate-200">
                    {isTh ? "อนุญาตให้ผู้ดูแลระบบ (Admin) แก้ไขเอกสารที่ล็อกได้" : "Allow Admin Bypass"}
                  </h3>
                  <p className="text-[11px] text-slate-500">
                    {isTh
                      ? "เมื่อเปิด ผู้ใช้ระดับ Admin/สมุห์บัญชี สามารถแก้ไขได้ แต่ระบบจะบันทึก Audit Trail ทุกครั้ง"
                      : "Supervisors/Admins can edit with compulsory audit trail logging."}
                  </p>
                </div>
              </div>
              <input
                type="checkbox"
                checked={editBypass}
                onChange={(e) => setEditBypass(e.target.checked)}
                className="w-4 h-4 rounded text-indigo-600 focus:ring-indigo-500 border-slate-300"
              />
            </div>

            {/* Remark / Reason Input */}
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                <Info className="w-3.5 h-3.5 text-slate-400" />
                <span>{isTh ? "เหตุผล / หมายเหตุการล็อกงวด:" : "Reason / Closing Note:"}</span>
              </label>
              <Input
                value={editRemark}
                onChange={(e) => setEditRemark(e.target.value)}
                placeholder={isTh ? "ระบุเหตุผล เช่น ปิดงวดภาษีมูลค่าเพิ่ม / ตรวจนับสต็อกสิ้นไตรมาส" : "Enter reason..."}
                className="h-8 text-xs bg-white dark:bg-slate-900"
              />
            </div>

            {/* Last Updated Meta info */}
            <div className="flex items-center justify-between text-[11px] text-slate-400 pt-2 border-t border-slate-100 dark:border-slate-800">
              <div className="flex items-center gap-1.5">
                <UserCheck className="w-3.5 h-3.5 text-slate-500" />
                <span>
                  {isTh ? "ผู้ปรับปรุงล่าสุด:" : "Last updated by:"} {selectedModule.lockedBy}
                </span>
              </div>
              <div className="flex items-center gap-1.5">
                <Clock className="w-3.5 h-3.5 text-slate-500" />
                <span>{selectedModule.updatedAt}</span>
              </div>
            </div>
          </div>

          {/* Detail Card 2: Audit Trail Log */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-3">
            <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
              <div className="flex items-center gap-2">
                <History className="w-4 h-4 text-indigo-500" />
                <h3 className="text-xs font-bold text-slate-900 dark:text-white">
                  {isTh ? "ประวัติการดำเนินการล็อกและปลดล็อก (Lock Audit Trail)" : "Lock History & Audit Trail"}
                </h3>
              </div>
              <span className="text-[11px] text-slate-500">
                {isTh ? `รวม ${auditLogs.length} รายการ` : `${auditLogs.length} records`}
              </span>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-200 dark:border-slate-800 text-[11px] text-slate-500 bg-slate-50/70 dark:bg-slate-950/70">
                    <th className="py-2 px-3 font-semibold">{isTh ? "วันเวลา" : "Timestamp"}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "ระบบงาน" : "Module"}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "การกระทำ" : "Action"}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "ช่วงวันที่ล็อกใหม่" : "New Lock Date"}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "ผู้ดำเนินการ" : "User"}</th>
                    <th className="py-2 px-3 font-semibold">{isTh ? "เหตุผล" : "Reason"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono text-[11px]">
                  {auditLogs.map((log) => (
                    <tr key={log.id} className="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                      <td className="py-2 px-3 text-slate-500 whitespace-nowrap">{log.timestamp}</td>
                      <td className="py-2 px-3 font-sans font-medium text-slate-800 dark:text-slate-200">
                        {log.moduleNameTh}
                      </td>
                      <td className="py-2 px-3 whitespace-nowrap">
                        <span
                          className={`px-2 py-0.5 rounded-full text-[10px] font-semibold ${
                            log.action === "lock" || log.action === "extend_date"
                              ? "bg-rose-100 text-rose-800 dark:bg-rose-950/60 dark:text-rose-300"
                              : "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300"
                          }`}
                        >
                          {log.action === "lock"
                            ? isTh ? "ล็อก" : "Lock"
                            : log.action === "extend_date"
                            ? isTh ? "ขยายวันล็อก" : "Extend"
                            : isTh ? "ปลดล็อก" : "Unlock"}
                        </span>
                      </td>
                      <td className="py-2 px-3 text-indigo-600 dark:text-indigo-400 font-semibold">
                        {log.newDate}
                      </td>
                      <td className="py-2 px-3 font-sans text-slate-600 dark:text-slate-300 whitespace-nowrap">
                        {log.performedBy}
                      </td>
                      <td className="py-2 px-3 font-sans text-slate-500 max-w-[200px] truncate" title={log.reason}>
                        {log.reason}
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
