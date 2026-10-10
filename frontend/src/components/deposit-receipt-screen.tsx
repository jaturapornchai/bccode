"use client";

import { useId, useMemo, useRef, useState } from "react";
import {
  Building2,
  Calendar,
  CheckCircle2,
  Clock,
  CreditCard,
  DollarSign,
  Download,
  FileCheck,
  FileText,
  Filter,
  Layers,
  Pencil,
  Percent,
  Plus,
  Printer,
  Receipt,
  RefreshCw,
  Search,
  Tag,
  Trash2,
  User,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { thaiBahtText } from "@/lib/thai-baht-text";
import { type LanguageCode } from "@/lib/i18n";

export interface DepositReceiptItem {
  id: string;
  docNo: string;
  docDate: string;
  customerCode: string;
  customerName: string;
  taxId: string;
  address: string;
  refOrderNo: string;
  depositAmount: string;
  vatRate: string;
  vatAmount: string;
  whtRate: string;
  whtAmount: string;
  netTotal: string;
  allocatedAmount: string;
  remainingAmount: string;
  paymentMethod: "bank_transfer" | "cheque" | "cash";
  bankAccount: string;
  glAccount: string;
  status: "unallocated" | "partially_allocated" | "fully_allocated" | "cancelled";
  notes: string;
}

const INITIAL_DEPOSITS: DepositReceiptItem[] = [
  {
    id: "dep-01",
    docNo: "RDEP6910-0001",
    docDate: "2026-10-10",
    customerCode: "AR-001",
    customerName: "บริษัท สยามคอนกรีตมิกซ์ จำกัด",
    taxId: "0105558123456",
    address: "88/1 ถ.พระราม 3 แขวงบางคอแหลม เขตบางคอแหลม กรุงเทพมหานคร 10120",
    refOrderNo: "SO-6910-0012",
    depositAmount: "50000.00",
    vatRate: "7.00",
    vatAmount: "3500.00",
    whtRate: "0.00",
    whtAmount: "0.00",
    netTotal: "53500.00",
    allocatedAmount: "0.00",
    remainingAmount: "50000.00",
    paymentMethod: "bank_transfer",
    bankAccount: "BBL กระแสรายวัน 101-3-01928-1",
    glAccount: "21310 - เงินมัดจำรับล่วงหน้าจากลูกค้า",
    status: "unallocated",
    notes: "เงินมัดจำสั่งซื้อปูนซีเมนต์และเหล็กเส้นก่อสร้าง 20%",
  },
  {
    id: "dep-02",
    docNo: "RDEP6910-0002",
    docDate: "2026-10-09",
    customerCode: "AR-002",
    customerName: "ห้างหุ้นส่วนจำกัด นครหลวงวิศวกรรม",
    taxId: "0103554098765",
    address: "124/5 ถ.สุขุมวิท 71 แขวงพระโขนงเหนือ เขตวัฒนา กรุงเทพฯ 10110",
    refOrderNo: "SO-6910-0015",
    depositAmount: "80000.00",
    vatRate: "7.00",
    vatAmount: "5600.00",
    whtRate: "0.00",
    whtAmount: "0.00",
    netTotal: "85600.00",
    allocatedAmount: "30000.00",
    remainingAmount: "50000.00",
    paymentMethod: "bank_transfer",
    bankAccount: "KBANK กระแสรายวัน 029-1-98765-4",
    glAccount: "21310 - เงินมัดจำรับล่วงหน้าจากลูกค้า",
    status: "partially_allocated",
    notes: "เงินมัดจำงานโครงสร้างเหล็กรูปพรรณ (ตัดชำระบิลส่งของงวดแรกแล้ว 30,000 บาท)",
  },
  {
    id: "dep-03",
    docNo: "RDEP6910-0003",
    docDate: "2026-10-08",
    customerCode: "AR-003",
    customerName: "บริษัท ธนบุรีอินดัสเตรียล จำกัด",
    taxId: "0105559012389",
    address: "45 หมู่ 2 ถ.เพชรเกษม แขวงบางแคเหนือ เขตบางแค กรุงเทพฯ 10160",
    refOrderNo: "SO-6910-0008",
    depositAmount: "120000.00",
    vatRate: "7.00",
    vatAmount: "8400.00",
    whtRate: "0.00",
    whtAmount: "0.00",
    netTotal: "128400.00",
    allocatedAmount: "120000.00",
    remainingAmount: "0.00",
    paymentMethod: "cheque",
    bankAccount: "เช็ครับ SCB เลขที่ 0892341",
    glAccount: "21310 - เงินมัดจำรับล่วงหน้าจากลูกค้า",
    status: "fully_allocated",
    notes: "เงินมัดจำสั่งผลิตเสาเข็มคอนกรีตอัดแรง (ตัดส่งมอบครบถ้วนแล้ว)",
  },
  {
    id: "dep-04",
    docNo: "RDEP6910-0004",
    docDate: "2026-10-07",
    customerCode: "AR-004",
    customerName: "บริษัท สหรุ่งโรจน์ ดีเวลลอปเมนท์ จำกัด",
    taxId: "0105561023451",
    address: "500 ถ.พหลโยธิน แขวงลาดยาว เขตจตุจักร กรุงเทพฯ 10900",
    refOrderNo: "SO-6910-0019",
    depositAmount: "35000.00",
    vatRate: "7.00",
    vatAmount: "2450.00",
    whtRate: "0.00",
    whtAmount: "0.00",
    netTotal: "37450.00",
    allocatedAmount: "0.00",
    remainingAmount: "35000.00",
    paymentMethod: "cash",
    bankAccount: "เงินสดในมือ (แผนกการเงินสำนักงานใหญ่)",
    glAccount: "21310 - เงินมัดจำรับล่วงหน้าจากลูกค้า",
    status: "unallocated",
    notes: "มัดจำล่วงหน้างานกระเบื้องปูพื้นและสุขภัณฑ์",
  },
];

interface DepositReceiptScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function DepositReceiptScreen({ language = "th", holdingcode = "" }: DepositReceiptScreenProps) {
  const [items, setItems] = useState<DepositReceiptItem[]>(INITIAL_DEPOSITS);
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [search, setSearch] = useState("");
  const [activeItemId, setActiveItemId] = useState<string>(INITIAL_DEPOSITS[0].id);

  // Modal dialog state
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<DepositReceiptItem | null>(null);

  // Form Fields
  const [formCustomerCode, setFormCustomerCode] = useState("AR-001");
  const [formCustomerName, setFormCustomerName] = useState("บริษัท สยามคอนกรีตมิกซ์ จำกัด");
  const [formTaxId, setFormTaxId] = useState("0105558123456");
  const [formAddress, setFormAddress] = useState("88/1 ถ.พระราม 3 แขวงบางคอแหลม เขตบางคอแหลม กรุงเทพฯ 10120");
  const [formRefOrderNo, setFormRefOrderNo] = useState("SO-6910-0022");
  const [formDepositAmount, setFormDepositAmount] = useState("30000.00");
  const [formPaymentMethod, setFormPaymentMethod] = useState<"bank_transfer" | "cheque" | "cash">("bank_transfer");
  const [formBankAccount, setFormBankAccount] = useState("BBL กระแสรายวัน 101-3-01928-1");
  const [formNotes, setFormNotes] = useState("เงินมัดจำตามใบสั่งขาย");

  const searchInputId = useId();
  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    defaultLeft: 34,
    min: 24,
    max: 45,
    storageKey: "bc_deposit_receipt_split",
    containerRef,
  });

  // Filtered Items
  const filteredItems = useMemo(() => {
    return items.filter((item) => {
      const matchStatus =
        statusFilter === "all" ||
        (statusFilter === "unallocated" && (item.status === "unallocated" || item.status === "partially_allocated")) ||
        item.status === statusFilter;
      const q = search.trim().toLowerCase();
      const matchSearch =
        !q ||
        item.docNo.toLowerCase().includes(q) ||
        item.customerName.toLowerCase().includes(q) ||
        item.customerCode.toLowerCase().includes(q) ||
        item.refOrderNo.toLowerCase().includes(q) ||
        item.notes.toLowerCase().includes(q);
      return matchStatus && matchSearch;
    });
  }, [items, statusFilter, search]);

  const activeItem = useMemo(() => {
    return items.find((item) => item.id === activeItemId) || filteredItems[0] || null;
  }, [items, activeItemId, filteredItems]);

  // Statistics
  const stats = useMemo(() => {
    let totalDeposit = 0;
    let totalRemaining = 0;
    let totalAllocated = 0;

    for (const it of items) {
      if (it.status !== "cancelled") {
        totalDeposit += parseFloat(it.depositAmount) || 0;
        totalRemaining += parseFloat(it.remainingAmount) || 0;
        totalAllocated += parseFloat(it.allocatedAmount) || 0;
      }
    }

    return {
      totalDepositFormatted: totalDeposit.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
      totalRemainingFormatted: totalRemaining.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
      totalAllocatedFormatted: totalAllocated.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
    };
  }, [items]);

  const handleExportCSV = () => {
    const headers = "เลขที่ใบรับมัดจำ,วันที่,รหัสลูกค้า,ชื่อลูกค้า,อ้างอิงใบสั่งขาย,ยอดเงินมัดจำ,VAT,ยอดสุทธิ,ยอดตัดแล้ว,ยอดคงเหลือ,สถานะ\n";
    const rows = filteredItems
      .map((i) =>
        `"${i.docNo}","${i.docDate}","${i.customerCode}","${i.customerName}","${i.refOrderNo}","${i.depositAmount}","${i.vatAmount}","${i.netTotal}","${i.allocatedAmount}","${i.remainingAmount}","${i.status}"`
      )
      .join("\n");
    const blob = new Blob(["\uFEFF" + headers + rows], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `deposit-receipts-${new Date().toISOString().slice(0, 10)}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  };

  const handlePrint = () => {
    window.print();
  };

  const handleOpenAdd = () => {
    setEditingItem(null);
    setFormCustomerCode("AR-001");
    setFormCustomerName("บริษัท สยามคอนกรีตมิกซ์ จำกัด");
    setFormTaxId("0105558123456");
    setFormAddress("88/1 ถ.พระราม 3 แขวงบางคอแหลม เขตบางคอแหลม กรุงเทพฯ 10120");
    setFormRefOrderNo("SO-6910-0022");
    setFormDepositAmount("30000.00");
    setFormPaymentMethod("bank_transfer");
    setFormBankAccount("BBL กระแสรายวัน 101-3-01928-1");
    setFormNotes("เงินมัดจำค่าสินค้าตามใบสั่งขาย 20%");
    setIsDialogOpen(true);
  };

  const handleOpenEdit = (item: DepositReceiptItem) => {
    setEditingItem(item);
    setFormCustomerCode(item.customerCode);
    setFormCustomerName(item.customerName);
    setFormTaxId(item.taxId);
    setFormAddress(item.address);
    setFormRefOrderNo(item.refOrderNo);
    setFormDepositAmount(item.depositAmount);
    setFormPaymentMethod(item.paymentMethod);
    setFormBankAccount(item.bankAccount);
    setFormNotes(item.notes);
    setIsDialogOpen(true);
  };

  const handleDelete = (id: string) => {
    setItems((prev) => prev.filter((i) => i.id !== id));
  };

  const handleSaveForm = (e: React.FormEvent) => {
    e.preventDefault();
    const baseAmt = parseFloat(formDepositAmount) || 0;
    const vatAmt = baseAmt * 0.07;
    const net = baseAmt + vatAmt;

    if (editingItem) {
      setItems((prev) =>
        prev.map((i) =>
          i.id === editingItem.id
            ? {
                ...i,
                customerCode: formCustomerCode.trim(),
                customerName: formCustomerName.trim(),
                taxId: formTaxId.trim(),
                address: formAddress.trim(),
                refOrderNo: formRefOrderNo.trim(),
                depositAmount: baseAmt.toFixed(2),
                vatAmount: vatAmt.toFixed(2),
                netTotal: net.toFixed(2),
                remainingAmount: (baseAmt - parseFloat(i.allocatedAmount)).toFixed(2),
                paymentMethod: formPaymentMethod,
                bankAccount: formBankAccount,
                notes: formNotes.trim(),
              }
            : i
        )
      );
    } else {
      const newItem: DepositReceiptItem = {
        id: `dep-${Date.now()}`,
        docNo: `RDEP6910-${(items.length + 1).toString().padStart(4, "0")}`,
        docDate: new Date().toISOString().slice(0, 10),
        customerCode: formCustomerCode.trim(),
        customerName: formCustomerName.trim(),
        taxId: formTaxId.trim(),
        address: formAddress.trim(),
        refOrderNo: formRefOrderNo.trim(),
        depositAmount: baseAmt.toFixed(2),
        vatRate: "7.00",
        vatAmount: vatAmt.toFixed(2),
        whtRate: "0.00",
        whtAmount: "0.00",
        netTotal: net.toFixed(2),
        allocatedAmount: "0.00",
        remainingAmount: baseAmt.toFixed(2),
        paymentMethod: formPaymentMethod,
        bankAccount: formBankAccount,
        glAccount: "21310 - เงินมัดจำรับล่วงหน้าจากลูกค้า",
        status: "unallocated",
        notes: formNotes.trim(),
      };
      setItems((prev) => [newItem, ...prev]);
      setActiveItemId(newItem.id);
    }
    setIsDialogOpen(false);
  };

  const getStatusBadge = (status: DepositReceiptItem["status"]) => {
    switch (status) {
      case "unallocated":
        return <span className="rounded-full bg-amber-500/10 px-2.5 py-0.5 text-xs font-semibold text-amber-600 dark:text-amber-400 border border-amber-500/20">รอตัดชำระ</span>;
      case "partially_allocated":
        return <span className="rounded-full bg-blue-500/10 px-2.5 py-0.5 text-xs font-semibold text-blue-600 dark:text-blue-400 border border-blue-500/20">ตัดบางส่วน</span>;
      case "fully_allocated":
        return <span className="rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-xs font-semibold text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">ตัดครบแล้ว</span>;
      case "cancelled":
        return <span className="rounded-full bg-rose-500/10 px-2.5 py-0.5 text-xs font-semibold text-rose-600 dark:text-rose-400 border border-rose-500/20">ยกเลิก</span>;
    }
  };

  return (
    <div className="flex-1 min-h-0 flex flex-col gap-4 p-4 lg:p-6 overflow-hidden">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-5 shadow-xs">
        <div className="flex items-center gap-3.5">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-inner">
            <Receipt className="h-6 w-6" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">บันทึกใบรับเงินมัดจำ</h1>
              <span className="rounded-md bg-muted px-2 py-0.5 text-xs font-semibold text-muted-foreground">
                Champ Parity 30006
              </span>
            </div>
            <p className="text-sm text-muted-foreground mt-0.5">
              บันทึกรับเงินมัดจำล่วงหน้า ออกใบเสร็จรับเงินมัดจำ และติดตามการตัดชำระหนี้กับใบกำกับภาษีขาย (Customer Deposit)
            </p>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleExportCSV}
            className="gap-2 text-xs"
          >
            <Download className="h-3.5 w-3.5" />
            ส่งออก CSV
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={handlePrint}
            className="gap-2 text-xs"
          >
            <Printer className="h-3.5 w-3.5" />
            พิมพ์รายงาน
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={() => setItems(INITIAL_DEPOSITS)}
            className="gap-2 text-xs"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            คืนค่าเริ่มต้น
          </Button>

          <Button
            variant="default"
            size="sm"
            onClick={handleOpenAdd}
            className="gap-2 text-xs"
          >
            <Plus className="h-3.5 w-3.5" />
            สร้างใบรับเงินมัดจำ
          </Button>
        </div>
      </div>

      {/* Main Content Area with ResizableSplitter */}
      <div ref={containerRef} className="flex-1 min-h-0 flex flex-row rounded-2xl border border-border bg-card shadow-xs overflow-hidden relative">
        <div style={{ width: `${splitPercent}%` }} className="h-full flex flex-col border-r border-border/80 bg-muted/20 min-w-[250px] max-w-[500px] overflow-hidden">
              {/* Search & Filter Header */}
              <div className="p-3.5 border-b border-border/80 flex flex-col gap-2.5 bg-card/60">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    id={searchInputId}
                    placeholder="ค้นหาเลขที่, ลูกค้า, ใบสั่งขาย..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="!pl-9 h-9 text-xs"
                  />
                  {search && (
                    <button
                      onClick={() => setSearch("")}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    >
                      <X className="h-3.5 w-3.5" />
                    </button>
                  )}
                </div>

                {/* Status Tabs */}
                <div className="flex items-center gap-1 overflow-x-auto pb-1 text-xs">
                  {[
                    { id: "all", label: "ทั้งหมด" },
                    { id: "unallocated", label: "รอตัดชำระ" },
                    { id: "fully_allocated", label: "ตัดครบแล้ว" },
                  ].map((tab) => (
                    <button
                      key={tab.id}
                      onClick={() => setStatusFilter(tab.id)}
                      className={`px-2.5 py-1 rounded-lg font-medium whitespace-nowrap transition-colors ${
                        statusFilter === tab.id
                          ? "bg-primary text-primary-foreground shadow-xs"
                          : "text-muted-foreground hover:bg-muted/70 hover:text-foreground"
                      }`}
                    >
                      {tab.label}
                    </button>
                  ))}
                </div>
              </div>

              {/* Master List of Deposits */}
              <div className="flex-1 overflow-y-auto p-2 space-y-1.5">
                {filteredItems.map((item) => {
                  const isSelected = activeItemId === item.id;

                  return (
                    <div
                      key={item.id}
                      onClick={() => setActiveItemId(item.id)}
                      className={`group relative rounded-xl border p-3 cursor-pointer transition-all ${
                        isSelected
                          ? "border-primary bg-primary/5 shadow-xs ring-1 ring-primary/20"
                          : "border-border/60 bg-card hover:border-primary/40 hover:bg-muted/30"
                      }`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <span className="font-mono font-bold text-xs text-foreground">
                            {item.docNo}
                          </span>
                          <span className="text-[10px] text-muted-foreground ml-2">
                            {item.docDate}
                          </span>
                        </div>
                        {getStatusBadge(item.status)}
                      </div>

                      <div className="mt-1 text-xs font-semibold text-foreground truncate">
                        {item.customerName}
                      </div>

                      <div className="mt-1 flex items-center justify-between text-xs text-muted-foreground">
                        <span>อ้างอิง: {item.refOrderNo}</span>
                        <span className="font-mono font-bold text-primary">
                          {parseFloat(item.depositAmount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                        </span>
                      </div>

                      <div className="mt-1 pt-1.5 border-t border-border/40 flex items-center justify-between text-[11px] text-muted-foreground">
                        <span>คงเหลือค้างตัด:</span>
                        <span className="font-mono font-medium text-amber-600 dark:text-amber-400">
                          {parseFloat(item.remainingAmount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                        </span>
                      </div>
                    </div>
                  );
                })}

                {filteredItems.length === 0 && (
                  <div className="p-8 text-center text-xs text-muted-foreground">
                    ไม่พบรายการใบรับเงินมัดจำที่ตรงกับเงื่อนไข
                  </div>
                )}
              </div>

              {/* Left Bottom Summary */}
              <div className="p-3 border-t border-border/80 bg-card space-y-1 text-xs">
                <div className="flex items-center justify-between text-muted-foreground">
                  <span>ยอดรับมัดจำรวม:</span>
                  <span className="font-mono font-bold text-foreground">{stats.totalDepositFormatted} ฿</span>
                </div>
                <div className="flex items-center justify-between text-muted-foreground">
                  <span>ยอดคงเหลือรอตัด:</span>
                  <span className="font-mono font-bold text-amber-600 dark:text-amber-400">{stats.totalRemainingFormatted} ฿</span>
                </div>
              </div>
            </div>

            <ResizableSplitter
              value={splitPercent}
              isResizing={isResizing}
              onPointerDown={startResize}
              onKeyDown={adjustWithKeyboard}
            />

            <div className="flex-1 min-w-0 h-full flex flex-col overflow-y-auto">
              {activeItem ? (
                <div className="p-5 flex flex-col gap-5">
                  {/* Active Deposit Detail Card */}
                  <div className="rounded-xl border border-border bg-card p-5 shadow-xs space-y-4">
                    <div className="flex items-start justify-between gap-3 border-b border-border/70 pb-3.5">
                      <div>
                        <div className="flex items-center gap-2">
                          <h2 className="text-lg font-bold text-foreground">ใบรับเงินมัดจำ {activeItem.docNo}</h2>
                          {getStatusBadge(activeItem.status)}
                        </div>
                        <p className="text-xs text-muted-foreground mt-0.5">
                          วันที่เอกสาร: {activeItem.docDate} · อ้างอิงใบสั่งขาย: <span className="font-mono">{activeItem.refOrderNo}</span>
                        </p>
                      </div>

                      <div className="flex items-center gap-1.5">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleOpenEdit(activeItem)}
                          className="h-8 gap-1.5 text-xs"
                        >
                          <Pencil className="h-3.5 w-3.5" />
                          แก้ไข
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(activeItem.id)}
                          className="h-8 text-destructive hover:bg-destructive/10"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    </div>

                    {/* Customer Info Card */}
                    <div className="rounded-xl border border-border/80 bg-muted/20 p-4 space-y-2 text-xs">
                      <div className="flex items-center justify-between">
                        <div className="font-bold text-sm text-foreground flex items-center gap-2">
                          <Building2 className="h-4 w-4 text-primary" />
                          <span>{activeItem.customerName}</span>
                          <span className="font-mono text-xs text-muted-foreground">({activeItem.customerCode})</span>
                        </div>
                        <div className="text-muted-foreground">
                          เลขผู้เสียภาษี: <span className="font-mono">{activeItem.taxId}</span>
                        </div>
                      </div>
                      <div className="text-muted-foreground">{activeItem.address}</div>
                    </div>

                    {/* Accounting Amounts Grid */}
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                      <div className="p-3 rounded-lg bg-card border border-border/80">
                        <span className="text-muted-foreground">ยอดเงินมัดจำ (ฐานภาษี)</span>
                        <div className="font-bold font-mono text-foreground text-sm mt-0.5">
                          {parseFloat(activeItem.depositAmount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                        </div>
                      </div>
                      <div className="p-3 rounded-lg bg-card border border-border/80">
                        <span className="text-muted-foreground">ภาษีมูลค่าเพิ่ม (7%)</span>
                        <div className="font-bold font-mono text-foreground text-sm mt-0.5">
                          {parseFloat(activeItem.vatAmount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                        </div>
                      </div>
                      <div className="p-3 rounded-lg bg-card border border-border/80">
                        <span className="text-muted-foreground">ยอดรับสุทธิ</span>
                        <div className="font-bold font-mono text-primary text-sm mt-0.5">
                          {parseFloat(activeItem.netTotal).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                        </div>
                      </div>
                      <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20">
                        <span className="text-amber-700 dark:text-amber-400 font-semibold">ยอดคงเหลือค้างตัด</span>
                        <div className="font-bold font-mono text-amber-600 dark:text-amber-400 text-sm mt-0.5">
                          {parseFloat(activeItem.remainingAmount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                        </div>
                      </div>
                    </div>

                    {/* Thai Baht Text Box */}
                    <div className="rounded-lg border border-border/60 bg-muted/30 p-3 text-xs flex items-center justify-between">
                      <span className="text-muted-foreground">จำนวนเงินตัวอักษร:</span>
                      <span className="font-semibold text-foreground">
                        ({thaiBahtText(parseFloat(activeItem.netTotal))})
                      </span>
                    </div>

                    {/* Payment & GL Account Details */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                      <div className="p-3 rounded-lg bg-muted/20 border border-border/60 space-y-1">
                        <span className="text-muted-foreground font-semibold">ช่องทางรับเงิน / บัญชีเข้า:</span>
                        <div className="text-foreground">{activeItem.bankAccount}</div>
                      </div>
                      <div className="p-3 rounded-lg bg-muted/20 border border-border/60 space-y-1">
                        <span className="text-muted-foreground font-semibold">ผังบัญชีแยกประเภท (GL Account):</span>
                        <div className="text-foreground">{activeItem.glAccount}</div>
                      </div>
                    </div>

                    {/* Description Notes */}
                    <div className="text-xs space-y-1">
                      <span className="text-muted-foreground font-semibold">หมายเหตุ / วัตถุประสงค์:</span>
                      <p className="p-3 rounded-lg bg-muted/20 border border-border/60 text-foreground leading-relaxed">
                        {activeItem.notes}
                      </p>
                    </div>

                    {/* Cut Allocation Action Bar */}
                    <div className="pt-2 border-t border-border/70 flex items-center justify-between">
                      <div className="text-xs text-muted-foreground">
                        ตัดชำระแล้ว: <span className="font-mono font-semibold text-foreground">{parseFloat(activeItem.allocatedAmount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿</span>
                      </div>
                      <Button
                        size="sm"
                        variant="default"
                        className="gap-2 text-xs"
                        onClick={() => alert(`ตัดชำระเงินมัดจำ ${activeItem.docNo} กับใบแจ้งหนี้/ใบกำกับภาษีสำเร็จ`)}
                      >
                        <FileCheck className="h-3.5 w-3.5" />
                        ตัดชำระกับใบกำกับภาษี
                      </Button>
                    </div>
                  </div>

                  {/* Summary Table */}
                  <div className="rounded-xl border border-border bg-card p-4 shadow-xs space-y-3">
                    <h3 className="text-sm font-bold text-foreground">
                      ประวัติใบรับเงินมัดจำทั้งหมด ({filteredItems.length} รายการ)
                    </h3>

                    <div className="overflow-x-auto rounded-lg border border-border">
                      <table className="w-full text-xs text-left">
                        <thead className="bg-muted/50 border-b border-border text-muted-foreground font-semibold">
                          <tr>
                            <th className="p-2.5">เลขที่เอกสาร</th>
                            <th className="p-2.5">วันที่</th>
                            <th className="p-2.5">ชื่อลูกค้า</th>
                            <th className="p-2.5">อ้างอิง SO</th>
                            <th className="p-2.5 text-right">ยอดมัดจำ (฿)</th>
                            <th className="p-2.5 text-right">คงเหลือรอตัด (฿)</th>
                            <th className="p-2.5 text-center">สถานะ</th>
                            <th className="p-2.5 text-center">จัดการ</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                          {filteredItems.map((item) => (
                            <tr
                              key={item.id}
                              onClick={() => setActiveItemId(item.id)}
                              className={`cursor-pointer transition-colors ${
                                activeItemId === item.id ? "bg-primary/10" : "hover:bg-muted/40"
                              }`}
                            >
                              <td className="p-2.5 font-mono font-bold">{item.docNo}</td>
                              <td className="p-2.5">{item.docDate}</td>
                              <td className="p-2.5 font-medium max-w-[200px] truncate">{item.customerName}</td>
                              <td className="p-2.5 font-mono text-muted-foreground">{item.refOrderNo}</td>
                              <td className="p-2.5 text-right font-mono font-bold text-foreground">
                                {parseFloat(item.depositAmount).toLocaleString("en-US", { minimumFractionDigits: 2 })}
                              </td>
                              <td className="p-2.5 text-right font-mono font-bold text-amber-600 dark:text-amber-400">
                                {parseFloat(item.remainingAmount).toLocaleString("en-US", { minimumFractionDigits: 2 })}
                              </td>
                              <td className="p-2.5 text-center">{getStatusBadge(item.status)}</td>
                              <td className="p-2.5 text-center" onClick={(e) => e.stopPropagation()}>
                                <div className="flex items-center justify-center gap-1">
                                  <button
                                    onClick={() => handleOpenEdit(item)}
                                    className="p-1 text-muted-foreground hover:text-foreground"
                                  >
                                    <Pencil className="h-3.5 w-3.5" />
                                  </button>
                                  <button
                                    onClick={() => handleDelete(item.id)}
                                    className="p-1 text-destructive hover:bg-destructive/10 rounded"
                                  >
                                    <Trash2 className="h-3.5 w-3.5" />
                                  </button>
                                </div>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="flex-1 flex items-center justify-center p-8 text-xs text-muted-foreground">
                  เลือกรายการใบรับเงินมัดจำทางด้านซ้ายเพื่อดูรายละเอียด
                </div>
              )}
            </div>
      </div>

      {/* Modal Dialog for Add / Edit Deposit */}
      {isDialogOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-xs p-4 animate-in fade-in">
          <div className="w-full max-w-lg rounded-2xl border border-border bg-card p-6 shadow-xl space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <h3 className="text-base font-bold text-foreground">
                {editingItem ? "แก้ไขใบรับเงินมัดจำ" : "สร้างใบรับเงินมัดจำใหม่"}
              </h3>
              <button
                onClick={() => setIsDialogOpen(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleSaveForm} className="space-y-3.5 text-xs">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-muted-foreground font-semibold mb-1">รหัสลูกค้า</label>
                  <Input
                    required
                    value={formCustomerCode}
                    onChange={(e) => setFormCustomerCode(e.target.value)}
                    className="h-9 font-mono"
                    placeholder="เช่น AR-001"
                  />
                </div>
                <div>
                  <label className="block text-muted-foreground font-semibold mb-1">เลขผู้เสียภาษี 13 หลัก</label>
                  <Input
                    required
                    value={formTaxId}
                    onChange={(e) => setFormTaxId(e.target.value)}
                    className="h-9 font-mono"
                    placeholder="เช่น 0105558123456"
                  />
                </div>
              </div>

              <div>
                <label className="block text-muted-foreground font-semibold mb-1">ชื่อลูกค้า / บริษัท</label>
                <Input
                  required
                  value={formCustomerName}
                  onChange={(e) => setFormCustomerName(e.target.value)}
                  className="h-9"
                  placeholder="เช่น บริษัท สยามคอนกรีตมิกซ์ จำกัด"
                />
              </div>

              <div>
                <label className="block text-muted-foreground font-semibold mb-1">ที่อยู่</label>
                <Input
                  value={formAddress}
                  onChange={(e) => setFormAddress(e.target.value)}
                  className="h-9"
                  placeholder="ที่อยู่ตาม ภ.พ.20"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-muted-foreground font-semibold mb-1">ยอดเงินมัดจำก่อน VAT (บาท)</label>
                  <Input
                    required
                    type="number"
                    step="0.01"
                    value={formDepositAmount}
                    onChange={(e) => setFormDepositAmount(e.target.value)}
                    className="h-9 font-mono font-bold"
                  />
                </div>
                <div>
                  <label className="block text-muted-foreground font-semibold mb-1">อ้างอิงใบสั่งขาย (SO)</label>
                  <Input
                    value={formRefOrderNo}
                    onChange={(e) => setFormRefOrderNo(e.target.value)}
                    className="h-9 font-mono"
                    placeholder="เช่น SO-6910-0022"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-muted-foreground font-semibold mb-1">ช่องทางรับชำระ</label>
                  <select
                    value={formPaymentMethod}
                    onChange={(e) => setFormPaymentMethod(e.target.value as any)}
                    className="w-full h-9 rounded-md border border-input bg-background px-3 text-xs"
                  >
                    <option value="bank_transfer">โอนเงินธนาคาร</option>
                    <option value="cheque">เช็ครับ</option>
                    <option value="cash">เงินสด</option>
                  </select>
                </div>
                <div>
                  <label className="block text-muted-foreground font-semibold mb-1">บัญชีรับเงิน / รายละเอียด</label>
                  <Input
                    value={formBankAccount}
                    onChange={(e) => setFormBankAccount(e.target.value)}
                    className="h-9"
                    placeholder="เช่น BBL 101-3-01928-1"
                  />
                </div>
              </div>

              <div>
                <label className="block text-muted-foreground font-semibold mb-1">หมายเหตุ</label>
                <textarea
                  rows={2}
                  value={formNotes}
                  onChange={(e) => setFormNotes(e.target.value)}
                  className="w-full rounded-md border border-input bg-background p-2.5 text-xs"
                  placeholder="ระบุวัตถุประสงค์เงินมัดจำ"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setIsDialogOpen(false)}
                >
                  ยกเลิก
                </Button>
                <Button type="submit" variant="default" size="sm">
                  บันทึกใบรับเงินมัดจำ
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
