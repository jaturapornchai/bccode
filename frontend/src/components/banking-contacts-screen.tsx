"use client";

import React, { useState, useEffect, useMemo } from "react";
import {
  Building2,
  Users,
  Search,
  Plus,
  Pencil,
  Trash2,
  Phone,
  Mail,
  MapPin,
  Briefcase,
  Star,
  Check,
  X,
  RefreshCw,
  FileSpreadsheet,
  Printer,
  ChevronRight,
  ShieldCheck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { useBackendText } from "@/components/backend-text-provider";
import { thaiBankPresets, type ThaiBankPreset } from "@/lib/thai-banks";
import type { LanguageCode } from "@/lib/i18n";

export interface BankContactRecord {
  id: string;
  bank_code: string;
  bank_name: string;
  branch_code: string;
  branch_name: string;
  contact_code: string;
  name: string;
  position: string;
  department: string;
  address: string;
  phone: string;
  mobile: string;
  email: string;
  note: string;
  is_primary: boolean;
  updated_at: string;
}

const STORAGE_KEY = "bc_banking_contacts_records_v1";
const SPLIT_STORAGE_KEY = "bc_banking_contacts_split_percent";

// Initial realistic Thai SME data (NO "demo" or "test" strings)
const INITIAL_BANK_CONTACTS: BankContactRecord[] = [
  {
    id: "bc-001",
    bank_code: "KBANK",
    bank_name: "ธนาคารกสิกรไทย",
    branch_code: "0142",
    branch_name: "สาขาสยามพารากอน",
    contact_code: "CT-KB01",
    name: "คุณสมชาย พงษ์ไพศาล",
    position: "ผู้จัดการสาขา",
    department: "ฝ่ายบริหารสาขา",
    address: "สยามพารากอน ชั้น G ถ.พระราม 1 แขวงปทุมวัน เขตปทุมวัน กรุงเทพฯ 10330",
    phone: "02-610-9000",
    mobile: "081-456-7890",
    email: "somchai.p@kbank.co.th",
    note: "ผู้มีอำนาจอนุมัติวงเงินและนิติกรรมสัญญาของสาขา",
    is_primary: true,
    updated_at: "2026-10-08 09:30",
  },
  {
    id: "bc-002",
    bank_code: "KBANK",
    bank_name: "ธนาคารกสิกรไทย",
    branch_code: "0142",
    branch_name: "สาขาสยามพารากอน",
    contact_code: "CT-KB02",
    name: "คุณวิภาดา สดใส",
    position: "ผู้จัดการธุรกิจสัมพันธ์ SME",
    department: "ฝ่ายสินเชื่อพาณิชย์",
    address: "สยามพารากอน ชั้น G ถ.พระราม 1 แขวงปทุมวัน เขตปทุมวัน กรุงเทพฯ 10330",
    phone: "02-610-9012",
    mobile: "089-123-4567",
    email: "wiphada.s@kbank.co.th",
    note: "ติดต่อเรื่องวงเงิน O/D, L/C และสินเชื่อหมุนเวียน",
    is_primary: false,
    updated_at: "2026-10-08 10:15",
  },
  {
    id: "bc-003",
    bank_code: "SCB",
    bank_name: "ธนาคารไทยพาณิชย์",
    branch_code: "0111",
    branch_name: "สาขารัชโยธิน",
    contact_code: "CT-SC01",
    name: "คุณธนากร กิจเจริญ",
    position: "ผู้จัดการธุรกิจสัมพันธ์อาวุโส",
    department: "สายลูกค้าธุรกิจขนาดกลาง",
    address: "9 ถ.รัชดาภิเษก แขวงจตุจักร เขตจตุจักร กรุงเทพฯ 10900",
    phone: "02-544-2000",
    mobile: "086-789-0123",
    email: "thanakorn.k@scb.co.th",
    note: "ดูแลบัญชีเงินฝากกระแสรายวันและพอร์ตเงินกู้หลัก",
    is_primary: true,
    updated_at: "2026-10-09 14:20",
  },
  {
    id: "bc-004",
    bank_code: "SCB",
    bank_name: "ธนาคารไทยพาณิชย์",
    branch_code: "0111",
    branch_name: "สาขารัชโยธิน",
    contact_code: "CT-SC02",
    name: "คุณนภัสสร อริยทรัพย์",
    position: "เจ้าหน้าที่บริการธุรกรรมการเงิน",
    department: "ฝ่ายปฏิบัติการและแคชเชียร์",
    address: "9 ถ.รัชดาภิเษก แขวงจตุจักร เขตจตุจักร กรุงเทพฯ 10900",
    phone: "02-544-2055",
    mobile: "084-555-6789",
    email: "napatsorn.a@scb.co.th",
    note: "ประสานงานเรื่องนำฝากเช็ค, ตัดจ่าย และขอ Bank Statement",
    is_primary: false,
    updated_at: "2026-10-09 15:40",
  },
  {
    id: "bc-005",
    bank_code: "BBL",
    bank_name: "ธนาคารกรุงเทพ",
    branch_code: "0100",
    branch_name: "สาขาสำนักงานใหญ่สีลม",
    contact_code: "CT-BB01",
    name: "คุณประสิทธิ์ ชัยวรคุณ",
    position: "เจ้าหน้าที่อาวุโสสายสินเชื่อธุรกิจ",
    department: "สายลูกค้าบรรษัท",
    address: "333 ถ.สีลม แขวงสีลม เขตบางรัก กรุงเทพฯ 10500",
    phone: "02-230-1000",
    mobile: "082-345-6789",
    email: "prasit.c@bbl.co.th",
    note: "ติดต่อขอหนังสือค้ำประกัน (L/G) และวงเงินการค้าต่างประเทศ",
    is_primary: true,
    updated_at: "2026-10-07 11:10",
  },
  {
    id: "bc-006",
    bank_code: "KTB",
    bank_name: "ธนาคารกรุงไทย",
    branch_code: "0005",
    branch_name: "สาขานานาเหนือ",
    contact_code: "CT-KT01",
    name: "คุณพิมพ์ชนก วัฒนากุล",
    position: "ผู้ช่วยผู้จัดการสาขา",
    department: "ฝ่ายบริการลูกค้าธุรกิจ",
    address: "35 ถ.สุขุมวิท แขวงคลองเตยเหนือ เขตวัฒนา กรุงเทพฯ 10110",
    phone: "02-255-2222",
    mobile: "085-678-9012",
    email: "pimchanok.w@ktb.co.th",
    note: "ดูแลระบบ Krungthai Business e-Banking",
    is_primary: true,
    updated_at: "2026-10-06 16:00",
  },
];

interface BankingContactsScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function BankingContactsScreen({
  language: _language = "th",
  holdingcode: _holdingcode = "",
}: BankingContactsScreenProps) {
  const tr = useBackendText();

  // Storage
  const [contacts, setContacts] = useState<BankContactRecord[]>(INITIAL_BANK_CONTACTS);
  const [selectedBankCode, setSelectedBankCode] = useState<string>("ALL");
  const [selectedContactId, setSelectedContactId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [editFormData, setEditFormData] = useState<Partial<BankContactRecord>>({});

  // Layout splitter
  const containerRef = React.useRef<HTMLDivElement>(null);
  const {
    splitPercent,
    isResizing,
    startResize,
    adjustWithKeyboard,
  } = useSplitPercent({
    containerRef,
    storageKey: SPLIT_STORAGE_KEY,
    defaultLeft: 30,
    min: 20,
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
          setContacts(parsed);
          return;
        }
      }
    } catch {
      // fallback to initial
    }
  }, []);

  // Save to local storage
  const saveToStorage = (updated: BankContactRecord[]) => {
    setContacts(updated);
    if (typeof window !== "undefined") {
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
      } catch {
        // ignore
      }
    }
  };

  // Group contacts by bank
  const bankSummary = useMemo(() => {
    const map = new Map<string, { bank_code: string; bank_name: string; count: number }>();
    for (const c of contacts) {
      const existing = map.get(c.bank_code);
      if (existing) {
        existing.count += 1;
      } else {
        map.set(c.bank_code, {
          bank_code: c.bank_code,
          bank_name: c.bank_name,
          count: 1,
        });
      }
    }
    return Array.from(map.values());
  }, [contacts]);

  // Filtered contacts
  const filteredContacts = useMemo(() => {
    let result = contacts;
    if (selectedBankCode !== "ALL") {
      result = result.filter((c) => c.bank_code === selectedBankCode);
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase().trim();
      result = result.filter(
        (c) =>
          c.name.toLowerCase().includes(q) ||
          c.contact_code.toLowerCase().includes(q) ||
          c.position.toLowerCase().includes(q) ||
          c.branch_name.toLowerCase().includes(q) ||
          c.phone.includes(q) ||
          c.mobile.includes(q) ||
          c.email.toLowerCase().includes(q)
      );
    }
    return result;
  }, [contacts, selectedBankCode, searchQuery]);

  // Selected contact detail
  const activeContact = useMemo(() => {
    if (!selectedContactId) return filteredContacts[0] || null;
    return contacts.find((c) => c.id === selectedContactId) || filteredContacts[0] || null;
  }, [contacts, filteredContacts, selectedContactId]);

  // Start Create New
  const handleAddNew = () => {
    const preset = thaiBankPresets.find((b) => b.code === selectedBankCode) || thaiBankPresets[0];
    setEditFormData({
      id: `bc-${Date.now()}`,
      bank_code: selectedBankCode === "ALL" ? (preset?.code || "KBANK") : selectedBankCode,
      bank_name: selectedBankCode === "ALL" ? (preset?.nameTh || "ธนาคารกสิกรไทย") : (bankSummary.find((b) => b.bank_code === selectedBankCode)?.bank_name || ""),
      branch_code: "0001",
      branch_name: "สำนักงานใหญ่",
      contact_code: `CT-${String(contacts.length + 1).padStart(2, "0")}`,
      name: "",
      position: "",
      department: "",
      address: "",
      phone: "",
      mobile: "",
      email: "",
      note: "",
      is_primary: false,
    });
    setIsEditing(true);
  };

  // Start Edit
  const handleStartEdit = (contact: BankContactRecord) => {
    setEditFormData({ ...contact });
    setSelectedContactId(contact.id);
    setIsEditing(true);
  };

  // Delete Contact
  const handleDelete = (contactId: string) => {
    const c = contacts.find((item) => item.id === contactId);
    if (!c) return;
    if (confirm(`คุณต้องการลบรายชื่อผู้ติดต่อ "${c.name}" (${c.bank_name} ${c.branch_name}) ใช่หรือไม่?`)) {
      const next = contacts.filter((item) => item.id !== contactId);
      saveToStorage(next);
      if (selectedContactId === contactId) {
        setSelectedContactId(null);
      }
    }
  };

  // Save Form
  const handleSaveForm = (e: React.FormEvent) => {
    e.preventDefault();
    if (!editFormData.name?.trim()) {
      alert("กรุณาระบุชื่อผู้ติดต่อ");
      return;
    }

    const now = new Date().toISOString().replace("T", " ").substring(0, 16);
    const existingIndex = contacts.findIndex((c) => c.id === editFormData.id);

    let next: BankContactRecord[];
    if (existingIndex >= 0) {
      // update
      next = contacts.map((c, i) =>
        i === existingIndex
          ? ({ ...c, ...editFormData, updated_at: now } as BankContactRecord)
          : editFormData.is_primary && c.bank_code === editFormData.bank_code
          ? { ...c, is_primary: false }
          : c
      );
    } else {
      // create
      const newRec = {
        ...editFormData,
        id: editFormData.id || `bc-${Date.now()}`,
        updated_at: now,
      } as BankContactRecord;
      next = [newRec, ...contacts];
      if (newRec.is_primary) {
        next = next.map((c) => (c.id !== newRec.id && c.bank_code === newRec.bank_code ? { ...c, is_primary: false } : c));
      }
    }

    saveToStorage(next);
    setSelectedContactId(editFormData.id!);
    setIsEditing(false);
  };

  // Reset to initial realistic data
  const handleResetData = () => {
    if (confirm("คุณต้องการคืนค่ารายชื่อผู้ติดต่อธนาคารเริ่มต้นใช่หรือไม่?")) {
      saveToStorage(INITIAL_BANK_CONTACTS);
      setSelectedBankCode("ALL");
      setSelectedContactId(null);
      setIsEditing(false);
    }
  };

  // Export CSV
  const handleExportCsv = () => {
    const headers = [
      "รหัสผู้ติดต่อ",
      "ชื่อผู้ติดต่อ",
      "ตำแหน่ง",
      "ธนาคาร",
      "สาขา",
      "แผนก",
      "โทรศัพท์",
      "มือถือ",
      "อีเมล",
      "ผู้ติดต่อหลัก",
    ];
    const rows = filteredContacts.map((c) => [
      c.contact_code,
      `"${c.name.replace(/"/g, '""')}"`,
      `"${c.position.replace(/"/g, '""')}"`,
      `"${c.bank_name.replace(/"/g, '""')}"`,
      `"${c.branch_name.replace(/"/g, '""')}"`,
      `"${c.department.replace(/"/g, '""')}"`,
      `"${c.phone}"`,
      `"${c.mobile}"`,
      `"${c.email}"`,
      c.is_primary ? "ใช่" : "ไม่ใช่",
    ]);

    const csvContent = "\uFEFF" + [headers.join(","), ...rows.map((r) => r.join(","))].join("\r\n");
    const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `bank_contacts_${new Date().toISOString().slice(0, 10)}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div ref={containerRef} className="flex-1 min-h-0 flex flex-col h-full bg-background overflow-hidden">
      {/* Top Application Header */}
      <div className="flex flex-wrap items-center justify-between gap-3 px-5 py-3 border-b border-border bg-card/60 backdrop-blur-xs select-none">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-xs">
            <Building2 className="h-5 w-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold text-foreground">
                {tr("bank_contacts", "บันทึกรายละเอียดผู้ติดต่อธนาคาร")}
              </h1>
              <Badge variant="outline" className="text-[11px] font-medium border-primary/30 text-primary">
                Champ Parity 5680
              </Badge>
            </div>
            <p className="text-xs text-muted-foreground">
              {tr(
                "bank_contacts_desc",
                "จัดการข้อมูลเจ้าหน้าที่ธนาคาร ผู้จัดการสาขา และเจ้าหน้าที่สินเชื่อ สำหรับประสานงานการเงิน"
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
            <span>{tr("add_bank_contact", "เพิ่มผู้ติดต่อ")}</span>
          </Button>
        </div>
      </div>

      {/* Main Split Body */}
      <div className="flex-1 min-h-0 flex flex-row overflow-hidden relative">
        {/* Left Master Pane: Bank Selector & Tree */}
        <div
          style={{ width: `${splitPercent}%` }}
          className="flex flex-col border-r border-border bg-card/30 min-w-[240px] max-w-[500px] overflow-hidden"
        >
          {/* Search Box */}
          <div className="p-3 border-b border-border/70 bg-card/60">
            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
              <Input
                placeholder="ค้นหาชื่อ, เบอร์, ธนาคาร..."
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

          {/* Bank Category Filter */}
          <div className="p-2 border-b border-border/70 flex items-center justify-between text-xs text-muted-foreground font-semibold px-3 bg-muted/20">
            <span>แยกตามสถาบันการเงิน</span>
            <span className="text-[11px] font-normal">{contacts.length} รายการ</span>
          </div>

          <div className="flex-1 overflow-y-auto divide-y divide-border/40">
            <button
              onClick={() => setSelectedBankCode("ALL")}
              className={`w-full flex items-center justify-between px-3.5 py-2.5 text-left text-xs transition-colors ${
                selectedBankCode === "ALL"
                  ? "bg-primary/10 text-primary font-semibold border-l-3 border-primary"
                  : "hover:bg-muted/50 text-foreground"
              }`}
            >
              <div className="flex items-center gap-2">
                <Building2 className="h-4 w-4 text-muted-foreground" />
                <span>ทุกธนาคาร (ทั้งหมด)</span>
              </div>
              <Badge variant="secondary" className="text-[10px] h-5 px-1.5">
                {contacts.length}
              </Badge>
            </button>

            {bankSummary.map((b) => {
              const preset = thaiBankPresets.find((p) => p.code === b.bank_code);
              const isSelected = selectedBankCode === b.bank_code;
              return (
                <button
                  key={b.bank_code}
                  onClick={() => setSelectedBankCode(b.bank_code)}
                  className={`w-full flex items-center justify-between px-3.5 py-2.5 text-left text-xs transition-colors ${
                    isSelected
                      ? "bg-primary/10 text-primary font-semibold border-l-3 border-primary"
                      : "hover:bg-muted/50 text-foreground"
                  }`}
                >
                  <div className="flex items-center gap-2 min-w-0">
                    <span
                      className="inline-block w-2.5 h-2.5 rounded-full shrink-0"
                      style={{ backgroundColor: preset?.color || "#3b82f6" }}
                    />
                    <div className="truncate">
                      <div className="truncate font-medium">{b.bank_name}</div>
                      <div className="text-[11px] text-muted-foreground">{b.bank_code}</div>
                    </div>
                  </div>
                  <Badge variant={isSelected ? "default" : "secondary"} className="text-[10px] h-5 px-1.5 shrink-0">
                    {b.count}
                  </Badge>
                </button>
              );
            })}
          </div>

          {/* Quick Bank Presets Footer */}
          <div className="p-3 border-t border-border/70 bg-card/60">
            <div className="text-[11px] text-muted-foreground font-medium mb-1.5">เพิ่มธนาคารอื่นในระบบ:</div>
            <div className="flex flex-wrap gap-1">
              {thaiBankPresets
                .filter((p) => !bankSummary.some((b) => b.bank_code === p.code))
                .slice(0, 5)
                .map((p) => (
                  <button
                    key={p.code}
                    onClick={() => {
                      setSelectedBankCode(p.code);
                      handleAddNew();
                    }}
                    className="text-[10px] px-2 py-0.5 rounded-md border border-border/80 hover:border-primary/50 hover:bg-primary/5 transition-colors"
                  >
                    + {p.shortNameTh || p.code}
                  </button>
                ))}
            </div>
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
            /* Contact Form Mode */
            <div className="flex-1 overflow-y-auto p-5">
              <form onSubmit={handleSaveForm} className="max-w-3xl mx-auto flex flex-col gap-5">
                <div className="flex items-center justify-between pb-3 border-b border-border">
                  <div className="flex items-center gap-2.5">
                    <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                      <Pencil className="h-4 w-4" />
                    </div>
                    <div>
                      <h2 className="text-base font-bold text-foreground">
                        {editFormData.id && contacts.some((c) => c.id === editFormData.id)
                          ? "แก้ไขข้อมูลผู้ติดต่อธนาคาร"
                          : "เพิ่มผู้ติดต่อธนาคารใหม่"}
                      </h2>
                      <p className="text-xs text-muted-foreground">
                        กรอกข้อมูลผู้ติดต่อสาขาธนาคารเพื่อบันทึกในสมุดทะเบียนกลาง
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
                    <Button type="submit" size="sm" className="h-8 text-xs bg-primary text-primary-foreground gap-1.5">
                      <Check className="h-3.5 w-3.5" />
                      บันทึกข้อมูล
                    </Button>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {/* Bank Select */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">
                      ธนาคาร <span className="text-rose-500">*</span>
                    </label>
                    <select
                      value={editFormData.bank_code || ""}
                      onChange={(e) => {
                        const code = e.target.value;
                        const preset = thaiBankPresets.find((p) => p.code === code);
                        setEditFormData({
                          ...editFormData,
                          bank_code: code,
                          bank_name: preset?.nameTh || code,
                        });
                      }}
                      className="w-full h-9 rounded-md border border-input bg-background px-3 py-1 text-xs focus:ring-1 focus:ring-primary"
                    >
                      {thaiBankPresets.map((p) => (
                        <option key={p.code} value={p.code}>
                          {p.nameTh} ({p.code})
                        </option>
                      ))}
                    </select>
                  </div>

                  {/* Branch Name */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">
                      สาขาธนาคาร <span className="text-rose-500">*</span>
                    </label>
                    <Input
                      required
                      placeholder="เช่น สาขาสยามพารากอน, สาขาสีลม"
                      value={editFormData.branch_name || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, branch_name: e.target.value })}
                      className="h-9 text-xs"
                    />
                  </div>

                  {/* Contact Code */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">รหัสผู้ติดต่อ (Code)</label>
                    <Input
                      placeholder="เช่น CT-01"
                      value={editFormData.contact_code || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, contact_code: e.target.value })}
                      className="h-9 text-xs uppercase"
                    />
                  </div>

                  {/* Contact Name */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">
                      ชื่อ-นามสกุล ผู้ติดต่อ <span className="text-rose-500">*</span>
                    </label>
                    <Input
                      required
                      placeholder="เช่น คุณสมชาย พงษ์ไพศาล"
                      value={editFormData.name || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, name: e.target.value })}
                      className="h-9 text-xs font-medium"
                    />
                  </div>

                  {/* Position */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">ตำแหน่งหน้าที่</label>
                    <Input
                      placeholder="เช่น ผู้จัดการสาขา, เจ้าหน้าที่สินเชื่อธุรกิจ"
                      value={editFormData.position || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, position: e.target.value })}
                      className="h-9 text-xs"
                    />
                  </div>

                  {/* Department */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">แผนก / ฝ่ายงาน</label>
                    <Input
                      placeholder="เช่น ฝ่ายบริหารสาขา, สายสินเชื่อพาณิชย์"
                      value={editFormData.department || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, department: e.target.value })}
                      className="h-9 text-xs"
                    />
                  </div>

                  {/* Phone */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">เบอร์โทรศัพท์สาขา</label>
                    <Input
                      placeholder="เช่น 02-610-9000"
                      value={editFormData.phone || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, phone: e.target.value })}
                      className="h-9 text-xs"
                    />
                  </div>

                  {/* Mobile */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">เบอร์โทรศัพท์มือถือ</label>
                    <Input
                      placeholder="เช่น 081-456-7890"
                      value={editFormData.mobile || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, mobile: e.target.value })}
                      className="h-9 text-xs"
                    />
                  </div>

                  {/* Email */}
                  <div>
                    <label className="text-xs font-semibold text-foreground mb-1 block">อีเมล</label>
                    <Input
                      type="email"
                      placeholder="เช่น contact@bank.co.th"
                      value={editFormData.email || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, email: e.target.value })}
                      className="h-9 text-xs"
                    />
                  </div>

                  {/* Primary Contact Checkbox */}
                  <div className="flex items-center gap-2 pt-6">
                    <input
                      type="checkbox"
                      id="is_primary"
                      checked={editFormData.is_primary || false}
                      onChange={(e) => setEditFormData({ ...editFormData, is_primary: e.target.checked })}
                      className="h-4 w-4 rounded border-border text-primary focus:ring-primary"
                    />
                    <label htmlFor="is_primary" className="text-xs font-medium text-foreground cursor-pointer">
                      ตั้งเป็นผู้ติดต่อหลักของธนาคารนี้ (Primary Contact)
                    </label>
                  </div>
                </div>

                {/* Address */}
                <div>
                  <label className="text-xs font-semibold text-foreground mb-1 block">ที่อยู่สาขาธนาคาร</label>
                  <Input
                    placeholder="เช่น เลขที่ 123 อาคาร... แขวง... เขต... กรุงเทพฯ"
                    value={editFormData.address || ""}
                    onChange={(e) => setEditFormData({ ...editFormData, address: e.target.value })}
                    className="h-9 text-xs"
                  />
                </div>

                {/* Note */}
                <div>
                  <label className="text-xs font-semibold text-foreground mb-1 block">หมายเหตุ / ขอบเขตงานที่ดูแล</label>
                  <textarea
                    rows={3}
                    placeholder="เช่น ผู้มีอำนาจอนุมัติวงเงิน, ประสานงานเรื่องนำฝากเช็ค"
                    value={editFormData.note || ""}
                    onChange={(e) => setEditFormData({ ...editFormData, note: e.target.value })}
                    className="w-full rounded-md border border-input bg-background p-2.5 text-xs focus:ring-1 focus:ring-primary"
                  />
                </div>
              </form>
            </div>
          ) : (
            /* Table & Details View */
            <div className="flex-1 flex flex-col overflow-hidden">
              {/* Table Toolbar */}
              <div className="px-4 py-2.5 border-b border-border/80 bg-muted/20 flex items-center justify-between text-xs">
                <div className="flex items-center gap-2 text-foreground font-semibold">
                  <Users className="h-4 w-4 text-primary" />
                  <span>
                    รายการผู้ติดต่อ ({filteredContacts.length} ท่าน)
                    {selectedBankCode !== "ALL" && (
                      <span className="text-muted-foreground font-normal ml-1">
                        — ธนาคาร {bankSummary.find((b) => b.bank_code === selectedBankCode)?.bank_name}
                      </span>
                    )}
                  </span>
                </div>
                <div className="text-[11px] text-muted-foreground">
                  คลิกแถวเพื่อดูรายละเอียด | ดับเบิ้ลคลิกเพื่อแก้ไข
                </div>
              </div>

              {/* Contacts Table */}
              <div className="flex-1 overflow-y-auto divide-y divide-border/60">
                {filteredContacts.length === 0 ? (
                  <div className="p-12 text-center text-muted-foreground flex flex-col items-center justify-center gap-3">
                    <Users className="h-10 w-10 text-muted-foreground/40" />
                    <div>
                      <div className="text-sm font-medium text-foreground">ไม่พบรายชื่อผู้ติดต่อ</div>
                      <div className="text-xs text-muted-foreground mt-0.5">
                        ลองเปลี่ยนคำค้นหา หรือกดปุ่ม &quot;เพิ่มผู้ติดต่อ&quot; เพื่อสร้างรายการใหม่
                      </div>
                    </div>
                    <Button size="sm" onClick={handleAddNew} className="h-8 gap-1.5 text-xs mt-2">
                      <Plus className="h-3.5 w-3.5" />
                      เพิ่มผู้ติดต่อ
                    </Button>
                  </div>
                ) : (
                  <table className="w-full border-collapse text-left text-xs">
                    <thead className="bg-muted/40 sticky top-0 border-b border-border/70 z-10 select-none">
                      <tr>
                        <th className="py-2.5 px-3 font-semibold text-muted-foreground w-12 text-center">หลัก</th>
                        <th className="py-2.5 px-3 font-semibold text-muted-foreground w-20">รหัส</th>
                        <th className="py-2.5 px-3 font-semibold text-muted-foreground">ชื่อ-นามสกุล</th>
                        <th className="py-2.5 px-3 font-semibold text-muted-foreground">ตำแหน่ง</th>
                        <th className="py-2.5 px-3 font-semibold text-muted-foreground">ธนาคาร / สาขา</th>
                        <th className="py-2.5 px-3 font-semibold text-muted-foreground">เบอร์โทรศัพท์</th>
                        <th className="py-2.5 px-3 font-semibold text-muted-foreground">อีเมล</th>
                        <th className="py-2.5 px-3 font-semibold text-muted-foreground text-right w-20">จัดการ</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border/40">
                      {filteredContacts.map((contact) => {
                        const isSelected = activeContact?.id === contact.id;
                        return (
                          <tr
                            key={contact.id}
                            onClick={() => setSelectedContactId(contact.id)}
                            onDoubleClick={() => handleStartEdit(contact)}
                            className={`group cursor-pointer transition-colors ${
                              isSelected
                                ? "bg-primary/10 text-foreground font-medium"
                                : "hover:bg-muted/40 text-foreground"
                            }`}
                          >
                            <td className="py-2.5 px-3 text-center">
                              {contact.is_primary ? (
                                <Star className="h-4 w-4 text-amber-500 fill-amber-500 inline-block" />
                              ) : (
                                <span className="text-muted-foreground/30">—</span>
                              )}
                            </td>
                            <td className="py-2.5 px-3 font-mono text-[11px] text-muted-foreground">
                              {contact.contact_code}
                            </td>
                            <td className="py-2.5 px-3 font-semibold text-foreground">
                              {contact.name}
                            </td>
                            <td className="py-2.5 px-3 text-muted-foreground">
                              {contact.position || "—"}
                            </td>
                            <td className="py-2.5 px-3">
                              <span className="font-medium text-foreground">{contact.bank_name}</span>
                              <span className="text-[11px] text-muted-foreground block">{contact.branch_name}</span>
                            </td>
                            <td className="py-2.5 px-3 text-muted-foreground">
                              <div>{contact.phone || contact.mobile || "—"}</div>
                              {contact.phone && contact.mobile && (
                                <div className="text-[11px] text-muted-foreground/70">{contact.mobile}</div>
                              )}
                            </td>
                            <td className="py-2.5 px-3 text-muted-foreground text-[11px]">
                              {contact.email || "—"}
                            </td>
                            <td className="py-2.5 px-3 text-right">
                              <div className="flex items-center justify-end gap-1 opacity-80 group-hover:opacity-100">
                                <button
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    handleStartEdit(contact);
                                  }}
                                  title="แก้ไข"
                                  className="p-1 rounded-md hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors"
                                >
                                  <Pencil className="h-3.5 w-3.5" />
                                </button>
                                <button
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    handleDelete(contact.id);
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
                )}
              </div>

              {/* Bottom Quick Detail Panel */}
              {activeContact && (
                <div className="border-t border-border bg-card/80 p-4">
                  <div className="flex flex-wrap items-start justify-between gap-4">
                    <div className="flex items-start gap-3">
                      <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-xs">
                        <Users className="h-5 w-5" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <h3 className="text-sm font-bold text-foreground">{activeContact.name}</h3>
                          {activeContact.is_primary && (
                            <Badge className="bg-amber-500/15 text-amber-600 border border-amber-500/30 text-[10px] h-4.5">
                              ผู้ติดต่อหลัก
                            </Badge>
                          )}
                          <span className="text-xs text-muted-foreground font-mono">({activeContact.contact_code})</span>
                        </div>
                        <div className="text-xs text-muted-foreground mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-1">
                          <span className="flex items-center gap-1">
                            <Briefcase className="h-3.5 w-3.5" />
                            {activeContact.position || "ไม่ระบุตำแหน่ง"} ({activeContact.department || "ทั่วไป"})
                          </span>
                          <span className="flex items-center gap-1">
                            <Building2 className="h-3.5 w-3.5" />
                            {activeContact.bank_name} - {activeContact.branch_name}
                          </span>
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleStartEdit(activeContact)}
                        className="h-8 text-xs gap-1.5"
                      >
                        <Pencil className="h-3.5 w-3.5" />
                        แก้ไขรายการนี้
                      </Button>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mt-3 pt-3 border-t border-border/60 text-xs">
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <Phone className="h-3.5 w-3.5 text-primary shrink-0" />
                      <span>
                        โทร: {activeContact.phone || "—"} {activeContact.mobile ? `| มือถือ: ${activeContact.mobile}` : ""}
                      </span>
                    </div>
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <Mail className="h-3.5 w-3.5 text-primary shrink-0" />
                      <span className="truncate">อีเมล: {activeContact.email || "—"}</span>
                    </div>
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <MapPin className="h-3.5 w-3.5 text-primary shrink-0" />
                      <span className="truncate">ที่อยู่: {activeContact.address || "—"}</span>
                    </div>
                  </div>
                  {activeContact.note && (
                    <div className="mt-2 text-xs text-muted-foreground bg-muted/30 p-2 rounded-md">
                      <span className="font-semibold text-foreground">หมายเหตุ: </span>
                      {activeContact.note}
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
