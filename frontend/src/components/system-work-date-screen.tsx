"use client";

import { useMemo, useState } from "react";
import {
  AlertTriangle,
  Calendar,
  CalendarCheck,
  CheckCircle2,
  Clock,
  History,
  Lock,
  RotateCcw,
  Save,
  ShieldAlert,
  Sliders,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { type LanguageCode } from "@/lib/i18n";

export type WorkDateMode = "realtime" | "fixed" | "last_document";

export interface SystemWorkDateSettings {
  mode: WorkDateMode;
  fixedDate: string; // YYYY-MM-DD
  warnOutOfPeriod: boolean;
  blockClosedPeriod: boolean;
  maxBackdateDays: number;
  maxFutureDays: number;
  currentWorkDate: string; // Active session date
}

const DEFAULT_SETTINGS: SystemWorkDateSettings = {
  mode: "realtime",
  fixedDate: new Date().toISOString().split("T")[0],
  warnOutOfPeriod: true,
  blockClosedPeriod: true,
  maxBackdateDays: 30,
  maxFutureDays: 7,
  currentWorkDate: new Date().toISOString().split("T")[0],
};

export interface SystemWorkDateScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemWorkDateScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: SystemWorkDateScreenProps) {
  const isTh = language === "th";
  const [settings, setSettings] = useState<SystemWorkDateSettings>(DEFAULT_SETTINGS);
  const [notification, setNotification] = useState<string | null>(null);

  const fiscalPeriodInfo = useMemo(() => {
    const parts = settings.currentWorkDate.split("-");
    if (parts.length < 3) return { yearCE: 2026, yearBE: 2569, month: 10, periodStr: "10/2569" };
    const y = parseInt(parts[0], 10);
    const m = parseInt(parts[1], 10);
    return {
      yearCE: y,
      yearBE: y + 543,
      month: m,
      periodStr: `งวดที่ ${m}/${y + 543} (สถานะ: เปิดให้บันทึกรายการได้ปกติ)`,
    };
  }, [settings.currentWorkDate]);

  const handleUpdate = <K extends keyof SystemWorkDateSettings>(
    key: K,
    val: SystemWorkDateSettings[K]
  ) => {
    setSettings((prev) => ({ ...prev, [key]: val }));
  };

  const handleQuickSetDate = (type: "today" | "first_day" | "last_day") => {
    const now = new Date();
    const y = now.getFullYear();
    const m = now.getMonth();

    let targetDate = new Date();
    if (type === "today") {
      targetDate = now;
    } else if (type === "first_day") {
      targetDate = new Date(y, m, 1);
    } else if (type === "last_day") {
      targetDate = new Date(y, m + 1, 0);
    }

    const iso = targetDate.toISOString().split("T")[0];
    setSettings((prev) => ({
      ...prev,
      currentWorkDate: iso,
      ...(prev.mode === "fixed" ? { fixedDate: iso } : {}),
    }));
  };

  const handleSave = () => {
    setNotification(
      isTh
        ? "บันทึกการตั้งค่าวันที่ทำงานเริ่มต้นและมาตรการควบคุมงวดบัญชีเรียบร้อยแล้ว"
        : "Work date configuration and fiscal period safeguards saved successfully."
    );
    setTimeout(() => setNotification(null), 3500);
  };

  const handleReset = () => {
    setSettings(DEFAULT_SETTINGS);
    setNotification(isTh ? "คืนค่าการตั้งค่าเริ่มต้นเรียบร้อย" : "Reset settings to defaults");
    setTimeout(() => setNotification(null), 3000);
  };

  return (
    <div className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 font-sans text-xs">
      {/* Header Bar */}
      <div className="flex-none p-3.5 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-wrap items-center justify-between gap-3 shadow-xs">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-indigo-100 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
            <CalendarCheck className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-semibold text-slate-900 dark:text-slate-100">
                {isTh ? "กำหนดวันที่ทำงานเริ่มต้น" : "Default Work Date Configuration"}
              </h1>
              <span className="px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-[11px] text-slate-500 dark:text-slate-400">
              {isTh
                ? "จัดการระบบ › กำหนดค่าเริ่มต้นวันที่ทำงานเมื่อเปิดระบบ การควบคุมงวดบัญชี และจำกัดการคีย์ย้อนหลัง"
                : "System Settings › Configure initial system work date mode, period bounds, and backdate restrictions"}
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

      {/* Content Container */}
      <div className="flex-1 overflow-y-auto p-4 sm:p-6 max-w-4xl mx-auto w-full space-y-5">
        {/* Active Working Date Status Card */}
        <div className="bg-gradient-to-r from-indigo-900 to-slate-900 text-white p-5 rounded-xl shadow-md border border-indigo-800/60">
          <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
            <span className="text-indigo-200 font-medium flex items-center gap-2 text-xs">
              <Calendar className="w-4 h-4 text-amber-400" />
              {isTh ? "วันที่ทำงานปัจจุบันของระบบ (Active System Work Date)" : "Current Active Work Date"}
            </span>
            <span className="px-2.5 py-0.5 rounded-full text-[11px] font-mono bg-indigo-800/80 text-indigo-200">
              {fiscalPeriodInfo.periodStr}
            </span>
          </div>

          <div className="flex flex-wrap items-center gap-4 py-2">
            <div className="text-3xl font-mono font-bold text-amber-300 tracking-wider">
              {settings.currentWorkDate}
            </div>
            <div className="flex flex-wrap items-center gap-1.5">
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleQuickSetDate("today")}
                className="h-7 text-[11px] bg-white/10 hover:bg-white/20 border-white/20 text-white"
              >
                {isTh ? "วันนี้" : "Today"}
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleQuickSetDate("first_day")}
                className="h-7 text-[11px] bg-white/10 hover:bg-white/20 border-white/20 text-white"
              >
                {isTh ? "ต้นเดือน" : "Month Start"}
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleQuickSetDate("last_day")}
                className="h-7 text-[11px] bg-white/10 hover:bg-white/20 border-white/20 text-white"
              >
                {isTh ? "สิ้นเดือน" : "Month End"}
              </Button>
            </div>
          </div>
          <p className="mt-1 text-[11px] text-indigo-300">
            {isTh
              ? "วันที่นี้จะถูกใช้เป็นวันที่ตั้งต้นอัตโนมัติในทุกหน้าจอสร้างเอกสาร (สมุดรายวัน, ใบสำคัญรับ/จ่าย, ใบแจ้งหนี้ ฯลฯ)"
              : "This date is prepopulated by default when creating any new document in the system"}
          </p>
        </div>

        {/* Section 1: Startup Work Date Mode */}
        <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5 shadow-xs">
          <div className="flex items-center gap-2 pb-3 mb-4 border-b border-slate-100 dark:border-slate-800">
            <Clock className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
            <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
              {isTh ? "นโยบายวันที่เริ่มต้นเมื่อเปิดโปรแกรม" : "Startup Work Date Policy"}
            </h2>
          </div>

          <div className="space-y-3">
            {/* Realtime Mode */}
            <label
              className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-all ${
                settings.mode === "realtime"
                  ? "border-indigo-600 bg-indigo-50/50 dark:bg-indigo-950/30"
                  : "border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/40"
              }`}
            >
              <input
                type="radio"
                name="workDateMode"
                value="realtime"
                checked={settings.mode === "realtime"}
                onChange={() => handleUpdate("mode", "realtime")}
                className="mt-0.5 text-indigo-600 focus:ring-indigo-500"
              />
              <div className="flex-1">
                <span className="font-semibold text-slate-900 dark:text-slate-100">
                  {isTh ? "ใช้วันที่ปัจจุบันตามปฏิทินจริง (Real-time Clock)" : "Current Calendar Date"}
                </span>
                <p className="text-slate-500 dark:text-slate-400 text-[11px] mt-0.5">
                  {isTh
                    ? "เหมาะสำหรับการทำงานประจำวันทั่วไป ระบบจะใช้วันที่ของเครื่องเซิร์ฟเวอร์เสมอเมื่อเข้าใช้งาน"
                    : "Recommended for general operations. Uses server system clock on login"}
                </p>
              </div>
            </label>

            {/* Fixed Date Mode */}
            <label
              className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-all ${
                settings.mode === "fixed"
                  ? "border-indigo-600 bg-indigo-50/50 dark:bg-indigo-950/30"
                  : "border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/40"
              }`}
            >
              <input
                type="radio"
                name="workDateMode"
                value="fixed"
                checked={settings.mode === "fixed"}
                onChange={() => handleUpdate("mode", "fixed")}
                className="mt-0.5 text-indigo-600 focus:ring-indigo-500"
              />
              <div className="flex-1">
                <span className="font-semibold text-slate-900 dark:text-slate-100">
                  {isTh ? "ใช้วันที่คงที่ที่กำหนดไว้ (Fixed Accounting Date)" : "Fixed Specified Date"}
                </span>
                <p className="text-slate-500 dark:text-slate-400 text-[11px] mt-0.5">
                  {isTh
                    ? "เหมาะสำหรับสำนักงานบัญชีที่กำลังปิดงวดบัญชีย้อนหลัง หรือคีย์เอกสารจำนวนมากของงวดก่อน"
                    : "Ideal for batch retroactive data entry or closing a past accounting period"}
                </p>
                {settings.mode === "fixed" && (
                  <div className="mt-3 flex items-center gap-2">
                    <span className="text-[11px] text-slate-600 dark:text-slate-300 font-medium">
                      {isTh ? "ระบุวันที่คงที่:" : "Specify Date:"}
                    </span>
                    <Input
                      type="date"
                      value={settings.fixedDate}
                      onChange={(e) => {
                        handleUpdate("fixedDate", e.target.value);
                        handleUpdate("currentWorkDate", e.target.value);
                      }}
                      className="h-8 w-44 font-mono text-xs"
                    />
                  </div>
                )}
              </div>
            </label>

            {/* Last Document Mode */}
            <label
              className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-all ${
                settings.mode === "last_document"
                  ? "border-indigo-600 bg-indigo-50/50 dark:bg-indigo-950/30"
                  : "border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/40"
              }`}
            >
              <input
                type="radio"
                name="workDateMode"
                value="last_document"
                checked={settings.mode === "last_document"}
                onChange={() => handleUpdate("mode", "last_document")}
                className="mt-0.5 text-indigo-600 focus:ring-indigo-500"
              />
              <div className="flex-1">
                <span className="font-semibold text-slate-900 dark:text-slate-100">
                  {isTh
                    ? "ใช้วันที่ของเอกสารล่าสุดที่บันทึก (Last Entered Document Date)"
                    : "Last Document Entered Date"}
                </span>
                <p className="text-slate-500 dark:text-slate-400 text-[11px] mt-0.5">
                  {isTh
                    ? "ระบบจะจำวันที่ของเอกสารใบสุดท้ายที่ผู้ใช้บันทึกไว้ในเซสชัน เพื่อความต่อเนื่องในการคีย์งาน"
                    : "Continuously retains the date from the most recently recorded document"}
                </p>
              </div>
            </label>
          </div>
        </div>

        {/* Section 2: Fiscal Period Guard & Restrictions */}
        <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5 shadow-xs">
          <div className="flex items-center gap-2 pb-3 mb-4 border-b border-slate-100 dark:border-slate-800">
            <Lock className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
            <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
              {isTh ? "การควบคุมงวดบัญชีและการคีย์ย้อนหลัง" : "Fiscal Period Safeguards & Constraints"}
            </h2>
          </div>

          <div className="space-y-4">
            <label className="flex items-start gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={settings.warnOutOfPeriod}
                onChange={(e) => handleUpdate("warnOutOfPeriod", e.target.checked)}
                className="mt-0.5 rounded text-indigo-600 focus:ring-indigo-500 w-4 h-4"
              />
              <div>
                <span className="font-medium text-slate-800 dark:text-slate-200">
                  {isTh
                    ? "เตือนผู้ใช้เมื่อวันที่ในเอกสารไม่อยู่ในงวดบัญชีปัจจุบัน (Warning on Out-of-Period)"
                    : "Warn when document date falls outside active fiscal period"}
                </span>
                <p className="text-[11px] text-slate-400">
                  {isTh
                    ? "จะแสดงหน้าต่างแจ้งเตือนยืนยันก่อนบันทึก ป้องกันการลงวันที่ผิดพลาดข้ามเดือน/ปี"
                    : "Displays a confirmation dialog before saving to prevent accidental wrong-month entries"}
                </p>
              </div>
            </label>

            <label className="flex items-start gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={settings.blockClosedPeriod}
                onChange={(e) => handleUpdate("blockClosedPeriod", e.target.checked)}
                className="mt-0.5 rounded text-indigo-600 focus:ring-indigo-500 w-4 h-4"
              />
              <div>
                <span className="font-medium text-slate-800 dark:text-slate-200">
                  {isTh
                    ? "ห้ามบันทึกหรือแก้ไขเอกสารในงวดบัญชีที่ปิดไปแล้ว (Block Closed Period Modifications)"
                    : "Strictly block transaction creation or edits in closed periods"}
                </span>
                <p className="text-[11px] text-slate-400">
                  {isTh
                    ? "ล็อกความถูกต้องของงบการเงินที่ส่งตรวจสอบแล้ว ห้ามมีรายการค้าใหม่ย้อนหลัง"
                    : "Enforces accounting lock on historical audited periods"}
                </p>
              </div>
            </label>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2 border-t border-slate-100 dark:border-slate-800">
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "จำกัดการคีย์เอกสารย้อนหลังได้ไม่เกิน (วัน)" : "Max Backdate Allowance (Days)"}
                </label>
                <Input
                  type="number"
                  min={0}
                  value={settings.maxBackdateDays}
                  onChange={(e) => handleUpdate("maxBackdateDays", Math.max(0, parseInt(e.target.value) || 0))}
                  className="font-mono text-xs w-32"
                />
                <p className="mt-1 text-[10px] text-slate-400">
                  {isTh ? "ระบุ 0 หากไม่ต้องการจำกัดวันย้อนหลัง" : "Enter 0 for unlimited backdating"}
                </p>
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "จำกัดการลงวันที่ล่วงหน้าได้ไม่เกิน (วัน)" : "Max Future Date Allowance (Days)"}
                </label>
                <Input
                  type="number"
                  min={0}
                  value={settings.maxFutureDays}
                  onChange={(e) => handleUpdate("maxFutureDays", Math.max(0, parseInt(e.target.value) || 0))}
                  className="font-mono text-xs w-32"
                />
                <p className="mt-1 text-[10px] text-slate-400">
                  {isTh ? "ป้องกันการใส่วันที่ล่วงหน้าเกินกำหนด" : "Restricts document dating too far in advance"}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
