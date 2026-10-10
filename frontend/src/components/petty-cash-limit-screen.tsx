"use client";

import React, { useState, useEffect, useMemo, useRef } from "react";
import {
  Wallet,
  Search,
  Plus,
  Pencil,
  Trash2,
  DollarSign,
  User,
  Building,
  Check,
  X,
  RefreshCw,
  FileSpreadsheet,
  AlertCircle,
  TrendingDown,
  Layers,
  ArrowRight,
  ShieldCheck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { useBackendText } from "@/components/backend-text-provider";
import type { LanguageCode } from "@/lib/i18n";

export interface PettyCashLimitRecord {
  id: string;
  code: string;
  name: string;
  limit_amount: string; // TFRS Rule 20: No float, format "0.00"
  begin_amount: string;
  balance: string;
  custodian_name: string;
  gl_account_code: string;
  department: string;
  description: string;
  is_active: boolean;
  updated_at: string;
}

const STORAGE_KEY = "bc_petty_cash_limits_v1";
const SPLIT_STORAGE_KEY = "bc_petty_cash_split_percent";

// Initial realistic Thai SME data (NO "demo" or "test" strings)
export const INITIAL_PETTY_CASH_LIMITS: PettyCashLimitRecord[] = [
  {
    id: "pc-001",
    code: "PC-01",
    name: "เงินสดย่อยฝ่ายบริหารและสำนักงานใหญ่",
    limit_amount: "30000.00",
    begin_amount: "25000.00",
    balance: "22450.00",
    custodian_name: "คุณวราภรณ์ มั่นคง",
    gl_account_code: "11111",
    department: "ฝ่ายบริหารและบัญชีการเงิน",
    description: "สำหรับค่าใช้จ่ายเบ็ดเตล็ดสำนักงาน ค่าไปรษณีย์ และเครื่องเขียน",
    is_active: true,
    updated_at: "2026-10-08 11:20",
  },
  {
    id: "pc-002",
    code: "PC-02",
    name: "เงินสดย่อยแผนกจัดซื้อและคลังสินค้า",
    limit_amount: "50000.00",
    begin_amount: "40000.00",
    balance: "38500.00",
    custodian_name: "คุณกิตติศักดิ์ พรหมมา",
    gl_account_code: "11112",
    department: "ฝ่ายจัดซื้อและบริหารคลังสินค้า",
    description: "สำหรับค่าขนส่งด่วน ค่าซ่อมแซมฉุกเฉิน และวัสดุแพ็คเกจจิ้ง",
    is_active: true,
    updated_at: "2026-10-09 09:45",
  },
  {
    id: "pc-003",
    code: "PC-03",
    name: "เงินสดย่อยฝ่ายขายและการตลาด",
    limit_amount: "20000.00",
    begin_amount: "15000.00",
    balance: "14200.00",
    custodian_name: "คุณมนัสวี อัครเดช",
    gl_account_code: "11113",
    department: "ฝ่ายขายและการตลาด",
    description: "สำหรับค่าเดินทางพบลูกค้า ค่าที่จอดรถ และค่ารับรองเบื้องต้น",
    is_active: true,
    updated_at: "2026-10-07 14:15",
  },
  {
    id: "pc-004",
    code: "PC-04",
    name: "เงินสดย่อยส่วนโรงงานและซ่อมบำรุง",
    limit_amount: "40000.00",
    begin_amount: "35000.00",
    balance: "31800.00",
    custodian_name: "คุณสมศักดิ์ วิชัยดิษฐ์",
    gl_account_code: "11114",
    department: "ฝ่ายผลิตและซ่อมบำรุง",
    description: "สำหรับอะไหล่ด่วน น้ำมันหล่อลื่น และอุปกรณ์ช่าง",
    is_active: true,
    updated_at: "2026-10-06 16:30",
  },
];

interface PettyCashLimitScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function PettyCashLimitScreen({
  language: _language = "th",
  holdingcode: _holdingcode = "",
}: PettyCashLimitScreenProps) {
  const tr = useBackendText();

  // Storage state
  const [limits, setLimits] = useState<PettyCashLimitRecord[]>(INITIAL_PETTY_CASH_LIMITS);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [editFormData, setEditFormData] = useState<Partial<PettyCashLimitRecord>>({});

  // Layout splitter
  const containerRef = useRef<HTMLDivElement>(null);
  const {
    splitPercent,
    isResizing,
    startResize,
    adjustWithKeyboard,
  } = useSplitPercent({
    containerRef,
    storageKey: SPLIT_STORAGE_KEY,
    defaultLeft: 32,
    min: 22,
    max: 50,
  });

  // Load from local storage
  useEffect(() => {
    if (typeof window === "undefined") return;
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        if (Array.isArray(parsed) && parsed.length > 0) {
          setLimits(parsed);
          return;
        }
      }
    } catch {
      // fallback to initial
    }
  }, []);

  // Save to local storage
  const saveToStorage = (updated: PettyCashLimitRecord[]) => {
    setLimits(updated);
    if (typeof window !== "undefined") {
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
      } catch {
        // ignore
      }
    }
  };

  // Filtered limits
  const filteredLimits = useMemo(() => {
    if (!searchQuery.trim()) return limits;
    const q = searchQuery.toLowerCase().trim();
    return limits.filter(
      (item) =>
        item.name.toLowerCase().includes(q) ||
        item.code.toLowerCase().includes(q) ||
        item.custodian_name.toLowerCase().includes(q) ||
        item.department.toLowerCase().includes(q) ||
        item.gl_account_code.includes(q)
    );
  }, [limits, searchQuery]);

  // Selected item detail
  const activeItem = useMemo(() => {
    if (!selectedId) return filteredLimits[0] || null;
    return limits.find((item) => item.id === selectedId) || filteredLimits[0] || null;
  }, [limits, filteredLimits, selectedId]);

  // Total summary calculations
  const totalSummary = useMemo(() => {
    let totalLimit = 0;
    let totalBalance = 0;
    for (const item of limits) {
      totalLimit += parseFloat(item.limit_amount) || 0;
      totalBalance += parseFloat(item.balance) || 0;
    }
    return {
      totalLimit: totalLimit.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
      totalBalance: totalBalance.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
      count: limits.length,
    };
  }, [limits]);

  // Start Create New
  const handleAddNew = () => {
    setEditFormData({
      id: `pc-${Date.now()}`,
      code: `PC-${String(limits.length + 1).padStart(2, "0")}`,
      name: "",
      limit_amount: "20000.00",
      begin_amount: "20000.00",
      balance: "20000.00",
      custodian_name: "",
      gl_account_code: "11111",
      department: "",
      description: "",
      is_active: true,
    });
    setIsEditing(true);
  };

  // Start Edit
  const handleStartEdit = (record: PettyCashLimitRecord) => {
    setEditFormData({ ...record });
    setSelectedId(record.id);
    setIsEditing(true);
  };

  // Delete Record
  const handleDelete = (id: string) => {
    const item = limits.find((r) => r.id === id);
    if (!item) return;
    if (confirm(`คุณต้องการลบวงเงินสดย่อย "${item.name}" (${item.code}) ใช่หรือไม่?`)) {
      const next = limits.filter((r) => r.id !== id);
      saveToStorage(next);
      if (selectedId === id) setSelectedId(null);
    }
  };

  // Save Form
  const handleSaveForm = (e: React.FormEvent) => {
    e.preventDefault();
    if (!editFormData.name?.trim()) {
      alert("กรุณาระบุชื่อวงเงินสดย่อย");
      return;
    }
    if (!editFormData.code?.trim()) {
      alert("กรุณาระบุรหัสวงเงินสดย่อย");
      return;
    }

    // Clean monetary amounts
    const limitNum = parseFloat(editFormData.limit_amount || "0") || 0;
    const beginNum = parseFloat(editFormData.begin_amount || "0") || 0;
    const balanceNum = parseFloat(editFormData.balance || editFormData.begin_amount || "0") || 0;

    const formattedLimit = limitNum.toFixed(2);
    const formattedBegin = beginNum.toFixed(2);
    const formattedBalance = balanceNum.toFixed(2);

    const now = new Date().toISOString().replace("T", " ").substring(0, 16);
    const existingIndex = limits.findIndex((r) => r.id === editFormData.id);

    let next: PettyCashLimitRecord[];
    if (existingIndex >= 0) {
      next = limits.map((r, i) =>
        i === existingIndex
          ? ({
              ...r,
              ...editFormData,
              limit_amount: formattedLimit,
              begin_amount: formattedBegin,
              balance: formattedBalance,
              updated_at: now,
            } as PettyCashLimitRecord)
          : r
      );
    } else {
      const newRec = {
        ...editFormData,
        id: editFormData.id || `pc-${Date.now()}`,
        limit_amount: formattedLimit,
        begin_amount: formattedBegin,
        balance: formattedBalance,
        updated_at: now,
      } as PettyCashLimitRecord;
      next = [newRec, ...limits];
    }

    saveToStorage(next);
    setSelectedId(editFormData.id!);
    setIsEditing(false);
  };

  // Reset Data
  const handleResetData = () => {
    if (confirm("คุณต้องการคืนค่าวัดวงเงินสดย่อยเริ่มต้นใช่หรือไม่?")) {
      saveToStorage(INITIAL_PETTY_CASH_LIMITS);
      setSelectedId(null);
      setIsEditing(false);
    }
  };

  // Export CSV
  const handleExportCsv = () => {
    const headers = [
      "รหัสวงเงิน",
      "ชื่อวงเงินสดย่อย",
      "วงเงินที่กำหนด",
      "ยอดยกมา",
      "ยอดคงเหลือ",
      "ผู้ดูแลรับผิดชอบ",
      "รหัสบัญชีคุม",
      "แผนก",
      "สถานะ",
    ];
    const rows = filteredLimits.map((r) => [
      r.code,
      `"${r.name.replace(/"/g, '""')}"`,
      r.limit_amount,
      r.begin_amount,
      r.balance,
      `"${r.custodian_name.replace(/"/g, '""')}"`,
      r.gl_account_code,
      `"${r.department.replace(/"/g, '""')}"`,
      r.is_active ? "ใช้งาน" : "ระงับ",
    ]);

    const csvContent = "\uFEFF" + [headers.join(","), ...rows.map((r) => r.join(","))].join("\r\n");
    const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `petty_cash_limits_${new Date().toISOString().slice(0, 10)}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div ref={containerRef} className="flex-1 min-h-0 flex flex-col h-full bg-background overflow-hidden">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-3 px-5 py-3 border-b border-border bg-card/60 backdrop-blur-xs select-none">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-xs">
            <Wallet className="h-5 w-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold text-foreground">
                {tr("petty_cash_limit", "กำหนดวงเงินสดย่อย")}
              </h1>
              <Badge variant="outline" className="text-[11px] font-medium border-primary/30 text-primary">
                Champ Parity 70001
              </Badge>
            </div>
            <p className="text-xs text-muted-foreground">
              {tr(
                "petty_cash_limit_desc",
                "กำหนดและควบคุมวงเงินสดย่อยประจำแผนก ยอดยกมา บัญชีคุม และผู้รับผิดชอบถือเงินสดย่อย"
              )}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleExportCsv}
            className="h-8 gap-1.5 text-xs text-muted-foreground hover:text-foreground"
          >
            <FileSpreadsheet className="h-3.5 w-3.5 text-emerald-600" />
            <span>ส่งออก CSV</span>
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={handleResetData}
            title="คืนค่าข้อมูลเริ่มต้น"
            className="h-8 gap-1.5 text-xs text-muted-foreground hover:text-foreground"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span>คืนค่าเริ่มต้น</span>
          </Button>

          <Button
            size="sm"
            onClick={handleAddNew}
            className="h-8 gap-1.5 text-xs bg-primary text-primary-foreground shadow-xs font-semibold"
          >
            <Plus className="h-4 w-4" />
            <span>{tr("add_petty_cash_limit", "เพิ่มวงเงินสดย่อย")}</span>
          </Button>
        </div>
      </div>

      {/* Main Split Body */}
      <div className="flex-1 min-h-0 flex flex-row overflow-hidden relative">
        {/* Left Master Pane: List of Limits */}
        <div
          style={{ width: `${splitPercent}%` }}
          className="flex flex-col border-r border-border bg-card/30 min-w-[240px] max-w-[500px] overflow-hidden"
        >
          {/* Search Box */}
          <div className="p-3 border-b border-border/70 bg-card/60">
            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
              <Input
                placeholder="ค้นหารหัส, ชื่อวงเงิน, ผู้ถือเงิน..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="h-8 pl-8 text-xs bg-background/80"
              />
              {searchQuery && (
                <button
                  onClick={() => setSearchQuery("")}
                  className="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground"
                >
                  <X className="h-3 w-3" />
                </button>
              )}
            </div>
          </div>

          {/* List Summary Toolbar */}
          <div className="p-2 border-b border-border/70 flex items-center justify-between text-xs text-muted-foreground font-semibold px-3 bg-muted/20">
            <span>รายการวงเงินสดย่อย ({filteredLimits.length})</span>
            <span className="text-[11px] font-mono text-foreground font-medium">
              รวม: {totalSummary.totalLimit} ฿
            </span>
          </div>

          {/* Items List */}
          <div className="flex-1 overflow-y-auto divide-y divide-border/40">
            {filteredLimits.map((item) => {
              const isSelected = activeItem?.id === item.id;
              const limitFloat = parseFloat(item.limit_amount) || 0;
              const balanceFloat = parseFloat(item.balance) || 0;
              const percentUsed = limitFloat > 0 ? Math.round(((limitFloat - balanceFloat) / limitFloat) * 100) : 0;

              return (
                <button
                  key={item.id}
                  onClick={() => setSelectedId(item.id)}
                  className={`w-full flex flex-col gap-1.5 p-3.5 text-left text-xs transition-colors ${
                    isSelected
                      ? "bg-primary/10 text-primary font-semibold border-l-3 border-primary"
                      : "hover:bg-muted/50 text-foreground"
                  }`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2 min-w-0">
                      <span className="font-mono text-[11px] text-muted-foreground">{item.code}</span>
                      <span className="truncate font-semibold text-foreground">{item.name}</span>
                    </div>
                    {item.is_active ? (
                      <Badge variant="outline" className="text-[10px] h-4.5 px-1.5 text-emerald-600 border-emerald-500/30">
                        ใช้งาน
                      </Badge>
                    ) : (
                      <Badge variant="secondary" className="text-[10px] h-4.5 px-1.5 text-muted-foreground">
                        ระงับ
                      </Badge>
                    )}
                  </div>

                  <div className="flex items-center justify-between text-xs text-muted-foreground">
                    <span>วงเงิน: {parseFloat(item.limit_amount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿</span>
                    <span className="font-mono text-emerald-600 font-semibold">
                      คงเหลือ: {parseFloat(item.balance).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                    </span>
                  </div>

                  <div className="flex items-center justify-between text-[11px] text-muted-foreground/80 mt-0.5">
                    <span className="truncate">ผู้ดูแล: {item.custodian_name || "—"}</span>
                    <span>ใช้ไป {percentUsed}%</span>
                  </div>
                </button>
              );
            })}
          </div>

          {/* Quick Summary Footer */}
          <div className="p-3 border-t border-border/70 bg-card/60 flex items-center justify-between text-xs">
            <span className="text-muted-foreground">ยอดคงเหลือรวมทุกวงเงิน:</span>
            <span className="font-mono font-bold text-foreground text-sm">{totalSummary.totalBalance} ฿</span>
          </div>
        </div>

        {/* Resizable Splitter */}
        <ResizableSplitter
          value={splitPercent}
          isResizing={isResizing}
          onPointerDown={startResize}
          onKeyDown={adjustWithKeyboard}
        />

        {/* Right Detail Pane */}
        <div className="flex-1 min-w-0 flex flex-col bg-background overflow-hidden">
          {isEditing ? (
            /* Edit Form */
            <div className="flex-1 overflow-y-auto p-5">
              <form onSubmit={handleSaveForm} className="max-w-2xl mx-auto flex flex-col gap-5">
                <div className="flex items-center justify-between pb-3 border-b border-border">
                  <div className="flex items-center gap-2.5">
                    <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                      <Pencil className="h-4 w-4" />
                    </div>
                    <div>
                      <h2 className="text-base font-bold text-foreground">
                        {editFormData.id && limits.some((r) => r.id === editFormData.id)
                          ? "แก้ไขวงเงินสดย่อย"
                          : "เพิ่มวงเงินสดย่อยใหม่"}
                      </h2>
                      <p className="text-xs text-muted-foreground">
                        กรอกข้อมูลรายละเอียดวงเงินสดย่อย บัญชีคุม และผู้รับผิดชอบ
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => setIsEditing(false)}
                      className="h-8 text-xs"
                    >
                      ยกเลิก
                    </Button>
                    <Button
                      type="submit"
                      size="sm"
                      className="h-8 text-xs bg-primary text-primary-foreground gap-1.5"
                    >
                      <Check className="h-3.5 w-3.5" />
                      บันทึกข้อมูล
                    </Button>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {/* Code */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">
                      รหัสวงเงินสดย่อย <span className="text-rose-500">*</span>
                    </label>
                    <Input
                      required
                      placeholder="เช่น PC-01"
                      value={editFormData.code || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, code: e.target.value })}
                      className="h-9 text-xs uppercase font-mono font-medium"
                    />
                  </div>

                  {/* GL Account Code */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">
                      รหัสบัญชีแยกประเภท (GL) <span className="text-rose-500">*</span>
                    </label>
                    <Input
                      required
                      placeholder="เช่น 11111 เงินสดย่อย"
                      value={editFormData.gl_account_code || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, gl_account_code: e.target.value })}
                      className="h-9 text-xs font-mono"
                    />
                  </div>

                  {/* Name */}
                  <div className="md:col-span-2">
                    <label className="text-xs font-semibold text-foreground mb-1 block">
                      ชื่อวงเงินสดย่อย <span className="text-rose-500">*</span>
                    </label>
                    <Input
                      required
                      placeholder="เช่น เงินสดย่อยฝ่ายบริหารและสำนักงาน"
                      value={editFormData.name || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, name: e.target.value })}
                      className="h-9 text-xs font-medium"
                    />
                  </div>

                  {/* Limit Amount */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">
                      วงเงินที่ได้รับอนุมัติ (บาท) <span className="text-rose-500">*</span>
                    </label>
                    <Input
                      type="number"
                      step="0.01"
                      min="0"
                      required
                      placeholder="30000.00"
                      value={editFormData.limit_amount || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, limit_amount: e.target.value })}
                      className="h-9 text-xs font-mono font-semibold text-primary"
                    />
                  </div>

                  {/* Begin Amount */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">
                      ยอดยกมาต้นงวด (บาท) <span className="text-rose-500">*</span>
                    </label>
                    <Input
                      type="number"
                      step="0.01"
                      min="0"
                      required
                      placeholder="25000.00"
                      value={editFormData.begin_amount || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, begin_amount: e.target.value })}
                      className="h-9 text-xs font-mono"
                    />
                  </div>

                  {/* Custodian Name */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">ผู้ถือเงินสดย่อย / ผู้ดูแล</label>
                    <Input
                      placeholder="เช่น คุณวราภรณ์ มั่นคง"
                      value={editFormData.custodian_name || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, custodian_name: e.target.value })}
                      className="h-9 text-xs"
                    />
                  </div>

                  {/* Department */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">แผนก / สังกัด</label>
                    <Input
                      placeholder="เช่น ฝ่ายบริหารและบัญชีการเงิน"
                      value={editFormData.department || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, department: e.target.value })}
                      className="h-9 text-xs"
                    />
                  </div>

                  {/* Active Toggle */}
                  <div className="flex items-center gap-2 pt-4">
                    <input
                      type="checkbox"
                      id="is_active"
                      checked={editFormData.is_active ?? true}
                      onChange={(e) => setEditFormData({ ...editFormData, is_active: e.target.checked })}
                      className="h-4 w-4 rounded border-border text-primary focus:ring-primary"
                    />
                    <label htmlFor="is_active" className="text-xs font-medium text-foreground cursor-pointer">
                      เปิดใช้งานวงเงินสดย่อยนี้ (Active)
                    </label>
                  </div>
                </div>

                {/* Description */}
                <div>
                  <label className="text-xs font-semibold text-foreground mb-1 block">วัตถุประสงค์ / หมายเหตุ</label>
                  <textarea
                    rows={3}
                    placeholder="ระบุเงื่อนไขการเบิกจ่าย หรือวงเงินสูงสุดต่อครั้ง..."
                    value={editFormData.description || ""}
                    onChange={(e) => setEditFormData({ ...editFormData, description: e.target.value })}
                    className="w-full rounded-md border border-input bg-background p-2.5 text-xs focus:ring-1 focus:ring-primary"
                  />
                </div>
              </form>
            </div>
          ) : (
            /* Table & Detail View */
            <div className="flex-1 flex flex-col overflow-hidden">
              {/* Toolbar */}
              <div className="px-4 py-2.5 border-b border-border/80 bg-muted/20 flex items-center justify-between text-xs">
                <div className="flex items-center gap-2 text-foreground font-semibold">
                  <Wallet className="h-4 w-4 text-primary" />
                  <span>ตารางแสดงสถานะวงเงินสดย่อย</span>
                </div>
                <div className="text-[11px] text-muted-foreground">
                  คลิกแถวเพื่อดูรายละเอียด | ดับเบิ้ลคลิกเพื่อแก้ไข
                </div>
              </div>

              {/* Table */}
              <div className="flex-1 overflow-y-auto divide-y divide-border/60">
                <table className="w-full border-collapse text-left text-xs">
                  <thead className="bg-muted/40 sticky top-0 border-b border-border/70 z-10 select-none">
                    <tr>
                      <th className="py-2.5 px-3 font-semibold text-muted-foreground w-20">รหัส</th>
                      <th className="py-2.5 px-3 font-semibold text-muted-foreground">ชื่อวงเงินสดย่อย</th>
                      <th className="py-2.5 px-3 font-semibold text-muted-foreground text-right">วงเงินที่กำหนด</th>
                      <th className="py-2.5 px-3 font-semibold text-muted-foreground text-right">ยอดยกมา</th>
                      <th className="py-2.5 px-3 font-semibold text-muted-foreground text-right">ยอดคงเหลือ</th>
                      <th className="py-2.5 px-3 font-semibold text-muted-foreground">ผู้ดูแล</th>
                      <th className="py-2.5 px-3 font-semibold text-muted-foreground">ผังบัญชี (GL)</th>
                      <th className="py-2.5 px-3 font-semibold text-muted-foreground text-center w-16">สถานะ</th>
                      <th className="py-2.5 px-3 font-semibold text-muted-foreground text-right w-20">จัดการ</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border/40">
                    {filteredLimits.map((record) => {
                      const isSelected = activeItem?.id === record.id;
                      return (
                        <tr
                          key={record.id}
                          onClick={() => setSelectedId(record.id)}
                          onDoubleClick={() => handleStartEdit(record)}
                          className={`group cursor-pointer transition-colors ${
                            isSelected
                              ? "bg-primary/10 text-foreground font-medium"
                              : "hover:bg-muted/40 text-foreground"
                          }`}
                        >
                          <td className="py-2.5 px-3 font-mono text-[11px] font-bold text-primary">
                            {record.code}
                          </td>
                          <td className="py-2.5 px-3 font-semibold text-foreground">
                            {record.name}
                            <span className="text-[11px] text-muted-foreground block">{record.department}</span>
                          </td>
                          <td className="py-2.5 px-3 text-right font-mono font-medium">
                            {parseFloat(record.limit_amount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                          </td>
                          <td className="py-2.5 px-3 text-right font-mono text-muted-foreground">
                            {parseFloat(record.begin_amount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                          </td>
                          <td className="py-2.5 px-3 text-right font-mono font-bold text-emerald-600">
                            {parseFloat(record.balance).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                          </td>
                          <td className="py-2.5 px-3 text-muted-foreground">{record.custodian_name || "—"}</td>
                          <td className="py-2.5 px-3 font-mono text-muted-foreground text-[11px]">
                            {record.gl_account_code}
                          </td>
                          <td className="py-2.5 px-3 text-center">
                            {record.is_active ? (
                              <Badge variant="outline" className="text-[10px] text-emerald-600 border-emerald-500/30">
                                ใช้งาน
                              </Badge>
                            ) : (
                              <Badge variant="secondary" className="text-[10px] text-muted-foreground">
                                ระงับ
                              </Badge>
                            )}
                          </td>
                          <td className="py-2.5 px-3 text-right">
                            <div className="flex items-center justify-end gap-1 opacity-80 group-hover:opacity-100">
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleStartEdit(record);
                                }}
                                title="แก้ไข"
                                className="p-1 rounded-md hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors"
                              >
                                <Pencil className="h-3.5 w-3.5" />
                              </button>
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleDelete(record.id);
                                }}
                                title="ลบ"
                                className="p-1 rounded-md hover:bg-rose-500/20 text-muted-foreground hover:text-rose-600 transition-colors"
                              >
                                <Trash2 className="h-3.5 w-3.5" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>

              {/* Bottom Quick Detail Panel */}
              {activeItem && (
                <div className="border-t border-border bg-card/80 p-4">
                  <div className="flex flex-wrap items-start justify-between gap-4">
                    <div className="flex items-start gap-3">
                      <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-xs">
                        <Wallet className="h-5 w-5" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <h3 className="text-sm font-bold text-foreground">{activeItem.name}</h3>
                          <Badge className="font-mono text-[10px]">{activeItem.code}</Badge>
                          {activeItem.is_active ? (
                            <Badge className="bg-emerald-500/15 text-emerald-600 border border-emerald-500/30 text-[10px] h-4.5">
                              สถานะเปิดใช้งาน
                            </Badge>
                          ) : (
                            <Badge variant="secondary" className="text-[10px] h-4.5">
                              ระงับใช้งาน
                            </Badge>
                          )}
                        </div>
                        <div className="text-xs text-muted-foreground mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-1">
                          <span className="flex items-center gap-1">
                            <User className="h-3.5 w-3.5" />
                            ผู้ถือเงิน: {activeItem.custodian_name || "ไม่ระบุ"}
                          </span>
                          <span className="flex items-center gap-1">
                            <Building className="h-3.5 w-3.5" />
                            แผนก: {activeItem.department || "ทั่วไป"}
                          </span>
                          <span className="flex items-center gap-1">
                            <Layers className="h-3.5 w-3.5" />
                            บัญชีคุม GL: {activeItem.gl_account_code}
                          </span>
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleStartEdit(activeItem)}
                        className="h-8 text-xs gap-1.5"
                      >
                        <Pencil className="h-3.5 w-3.5" />
                        แก้ไขวงเงินนี้
                      </Button>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mt-3 pt-3 border-t border-border/60">
                    <div className="bg-background/80 p-2.5 rounded-lg border border-border/60">
                      <div className="text-[11px] text-muted-foreground">วงเงินอนุมัติสูงสุด</div>
                      <div className="text-sm font-bold font-mono text-primary mt-0.5">
                        {parseFloat(activeItem.limit_amount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                      </div>
                    </div>
                    <div className="bg-background/80 p-2.5 rounded-lg border border-border/60">
                      <div className="text-[11px] text-muted-foreground">ยอดยกมาต้นงวด</div>
                      <div className="text-sm font-bold font-mono text-muted-foreground mt-0.5">
                        {parseFloat(activeItem.begin_amount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                      </div>
                    </div>
                    <div className="bg-background/80 p-2.5 rounded-lg border border-border/60">
                      <div className="text-[11px] text-muted-foreground">ยอดเงินสดคงเหลือปัจจุบัน</div>
                      <div className="text-sm font-bold font-mono text-emerald-600 mt-0.5">
                        {parseFloat(activeItem.balance).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                      </div>
                    </div>
                  </div>

                  {activeItem.description && (
                    <div className="mt-2 text-xs text-muted-foreground bg-muted/30 p-2 rounded-md">
                      <span className="font-semibold text-foreground">วัตถุประสงค์: </span>
                      {activeItem.description}
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
