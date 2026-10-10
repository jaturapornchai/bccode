"use client";

import { useId, useMemo, useRef, useState } from "react";
import {
  CheckSquare,
  ChevronRight,
  Download,
  FileCheck,
  FileText,
  Filter,
  Layers,
  Pencil,
  Plus,
  Printer,
  RefreshCw,
  Search,
  Trash2,
  X,
  CreditCard,
  Building2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { thaiBahtText } from "@/lib/thai-baht-text";
import { type LanguageCode } from "@/lib/i18n";

export interface PreparedChequeItem {
  id: string;
  docNo: string;
  chequeNo: string;
  bookNo: string;
  bankCode: string;
  bankName: string;
  branchName: string;
  payDate: string;
  dueDate: string;
  payeeCode: string;
  payeeName: string;
  amount: string;
  description: string;
  department: string;
  refDocNo: string;
  status: "prepared" | "approved" | "issued" | "cancelled";
  signatory1: string;
  signatory2: string;
}

export interface ChequeBookRegister {
  id: string;
  bookNo: string;
  accountNo: string;
  bankCode: string;
  bankName: string;
  branchName: string;
  startNo: string;
  endNo: string;
  currentNo: string;
  remainingCount: number;
}

const INITIAL_BOOKS: ChequeBookRegister[] = [
  {
    id: "bk-bbl-01",
    bookNo: "เล่มที่ 001/2569 (BBL)",
    accountNo: "101-3-01928-1",
    bankCode: "BBL",
    bankName: "ธนาคารกรุงเทพ",
    branchName: "สาขาสีลม",
    startNo: "001001",
    endNo: "001050",
    currentNo: "001023",
    remainingCount: 28,
  },
  {
    id: "bk-kbank-01",
    bookNo: "เล่มที่ 002/2569 (KBANK)",
    accountNo: "029-1-98765-4",
    bankCode: "KBANK",
    bankName: "ธนาคารกสิกรไทย",
    branchName: "สาขาสยามพารากอน",
    startNo: "002001",
    endNo: "002050",
    currentNo: "002015",
    remainingCount: 36,
  },
  {
    id: "bk-scb-01",
    bookNo: "เล่มที่ 001/2569 (SCB)",
    accountNo: "012-3-45678-9",
    bankCode: "SCB",
    bankName: "ธนาคารไทยพาณิชย์",
    branchName: "สาขารัชโยธิน",
    startNo: "003001",
    endNo: "003050",
    currentNo: "003009",
    remainingCount: 42,
  },
  {
    id: "bk-ktb-01",
    bookNo: "เล่มที่ 001/2569 (KTB)",
    accountNo: "005-1-23456-7",
    bankCode: "KTB",
    bankName: "ธนาคารกรุงไทย",
    branchName: "สาขานานาเหนือ",
    startNo: "004001",
    endNo: "004050",
    currentNo: "004020",
    remainingCount: 31,
  },
];

const INITIAL_CHEQUES: PreparedChequeItem[] = [
  {
    id: "cq-01",
    docNo: "CQ6910-0001",
    chequeNo: "001021",
    bookNo: "101-3-01928-1",
    bankCode: "BBL",
    bankName: "ธนาคารกรุงเทพ",
    branchName: "สาขาสีลม",
    payDate: "2026-10-10",
    dueDate: "2026-10-15",
    payeeCode: "AP-001",
    payeeName: "บริษัท ซีเมนต์ไทย (ปูนใหญ่) มาร์เก็ตติ้ง จำกัด",
    amount: "185000.00",
    description: "จ่ายค่าปูนซีเมนต์และเหล็กเส้นก่อสร้าง งวดที่ 1",
    department: "ฝ่ายจัดซื้อและคลังสินค้า",
    refDocNo: "PO-6910-0042",
    status: "prepared",
    signatory1: "นายสมคิด รุ่งเรือง",
    signatory2: "นางวิภาดา จิตดี",
  },
  {
    id: "cq-02",
    docNo: "CQ6910-0002",
    chequeNo: "002014",
    bookNo: "029-1-98765-4",
    bankCode: "KBANK",
    bankName: "ธนาคารกสิกรไทย",
    branchName: "สาขาสยามพารากอน",
    payDate: "2026-10-10",
    dueDate: "2026-10-20",
    payeeCode: "AP-002",
    payeeName: "ห้างหุ้นส่วนจำกัด นครหลวงคอนกรีต",
    amount: "120000.00",
    description: "จ่ายค่าคอนกรีตผสมเสร็จ งานเทพื้นฐานราก",
    department: "ฝ่ายปฏิบัติการและจัดส่ง",
    refDocNo: "PO-6910-0048",
    status: "prepared",
    signatory1: "นายสมคิด รุ่งเรือง",
    signatory2: "นางสาวศิริพร บุญช่วย",
  },
  {
    id: "cq-03",
    docNo: "CQ6910-0003",
    chequeNo: "003008",
    bookNo: "012-3-45678-9",
    bankCode: "SCB",
    bankName: "ธนาคารไทยพาณิชย์",
    branchName: "สาขารัชโยธิน",
    payDate: "2026-10-09",
    dueDate: "2026-10-25",
    payeeCode: "AP-003",
    payeeName: "บริษัท ทีพีไอ โพลีน จำกัด (มหาชน)",
    amount: "95000.00",
    description: "จ่ายค่ากระเบื้องหลังคาลอนคู่และฉนวนกันความร้อน",
    department: "ฝ่ายจัดซื้อและคลังสินค้า",
    refDocNo: "PO-6910-0051",
    status: "approved",
    signatory1: "นายสมคิด รุ่งเรือง",
    signatory2: "นางวิภาดา จิตดี",
  },
  {
    id: "cq-04",
    docNo: "CQ6910-0004",
    chequeNo: "004019",
    bookNo: "005-1-23456-7",
    bankCode: "KTB",
    bankName: "ธนาคารกรุงไทย",
    branchName: "สาขานานาเหนือ",
    payDate: "2026-10-08",
    dueDate: "2026-10-30",
    payeeCode: "AP-004",
    payeeName: "บริษัท ไทยรวมสินโลหะการ จำกัด",
    amount: "145000.00",
    description: "จ่ายค่าท่อแป๊บน้ำ เหล็กกล่อง และอุปกรณ์สุขภัณฑ์",
    department: "ฝ่ายจัดซื้อและคลังสินค้า",
    refDocNo: "PO-6910-0055",
    status: "issued",
    signatory1: "นายสมคิด รุ่งเรือง",
    signatory2: "นางวิภาดา จิตดี",
  },
  {
    id: "cq-05",
    docNo: "CQ6910-0005",
    chequeNo: "001022",
    bookNo: "101-3-01928-1",
    bankCode: "BBL",
    bankName: "ธนาคารกรุงเทพ",
    branchName: "สาขาสีลม",
    payDate: "2026-10-10",
    dueDate: "2026-10-18",
    payeeCode: "AP-005",
    payeeName: "ห้างหุ้นส่วนจำกัด สยามวู้ดโปรดักส์",
    amount: "40000.00",
    description: "จ่ายค่าไม้อัดและไม้แบบก่อสร้าง",
    department: "ฝ่ายปฏิบัติการและจัดส่ง",
    refDocNo: "PO-6910-0059",
    status: "prepared",
    signatory1: "นายสมคิด รุ่งเรือง",
    signatory2: "นางสาวศิริพร บุญช่วย",
  },
];

interface ChequeIssuedPrepareScreenProps {
  language?: LanguageCode;
}

export function ChequeIssuedPrepareScreen({ language = "th" }: ChequeIssuedPrepareScreenProps) {
  const [items, setItems] = useState<PreparedChequeItem[]>(INITIAL_CHEQUES);
  const [books] = useState<ChequeBookRegister[]>(INITIAL_BOOKS);
  const [selectedBookNo, setSelectedBookNo] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [search, setSearch] = useState("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [activeItemId, setActiveItemId] = useState<string>(INITIAL_CHEQUES[0].id);

  // Dialog State
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<PreparedChequeItem | null>(null);

  // Form State
  const [formPayeeName, setFormPayeeName] = useState("");
  const [formChequeNo, setFormChequeNo] = useState("");
  const [formBookNo, setFormBookNo] = useState(INITIAL_BOOKS[0].accountNo);
  const [formAmount, setFormAmount] = useState("0.00");
  const [formDueDate, setFormDueDate] = useState("2026-10-20");
  const [formDescription, setFormDescription] = useState("");
  const [formRefDocNo, setFormRefDocNo] = useState("");
  const [formDepartment, setFormDepartment] = useState("ฝ่ายจัดซื้อและคลังสินค้า");

  const searchInputId = useId();
  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    defaultLeft: 32,
    min: 24,
    max: 45,
    storageKey: "bc_cheque_issued_prepare_split",
    containerRef,
  });

  // Filtered Cheques
  const filteredItems = useMemo(() => {
    return items.filter((item) => {
      const matchBook = selectedBookNo === "all" || item.bookNo === selectedBookNo;
      const matchStatus = statusFilter === "all" || item.status === statusFilter;
      const q = search.trim().toLowerCase();
      const matchSearch =
        !q ||
        item.chequeNo.toLowerCase().includes(q) ||
        item.docNo.toLowerCase().includes(q) ||
        item.payeeName.toLowerCase().includes(q) ||
        item.description.toLowerCase().includes(q) ||
        item.refDocNo.toLowerCase().includes(q);
      return matchBook && matchStatus && matchSearch;
    });
  }, [items, selectedBookNo, statusFilter, search]);

  const activeItem = useMemo(() => {
    return items.find((item) => item.id === activeItemId) || filteredItems[0] || null;
  }, [items, activeItemId, filteredItems]);

  // Statistics
  const stats = useMemo(() => {
    let totalAmt = 0;
    let prepCount = 0;
    let approvedCount = 0;
    let issuedCount = 0;

    for (const it of items) {
      const amt = parseFloat(it.amount) || 0;
      totalAmt += amt;
      if (it.status === "prepared") prepCount++;
      if (it.status === "approved") approvedCount++;
      if (it.status === "issued") issuedCount++;
    }

    return {
      totalAmountFormatted: totalAmt.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
      preparedCount: prepCount,
      approvedCount,
      issuedCount,
    };
  }, [items]);

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedIds(filteredItems.map((i) => i.id));
    } else {
      setSelectedIds([]);
    }
  };

  const handleToggleSelect = (id: string) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((i) => i !== id) : [...prev, id]
    );
  };

  const handleBatchApprove = () => {
    if (selectedIds.length === 0) return;
    setItems((prev) =>
      prev.map((it) => (selectedIds.includes(it.id) ? { ...it, status: "approved" } : it))
    );
    setSelectedIds([]);
  };

  const handleExportCSV = () => {
    const headers = "เลขที่เอกสาร,เลขที่เช็ค,เลขที่บัญชี,ธนาคาร,สั่งจ่ายให้,ยอดเงิน,วันที่จ่าย,วันที่ครบกำหนด,วัตถุประสงค์,สถานะ\n";
    const rows = filteredItems
      .map((i) =>
        `"${i.docNo}","${i.chequeNo}","${i.bookNo}","${i.bankName}","${i.payeeName}","${i.amount}","${i.payDate}","${i.dueDate}","${i.description}","${i.status}"`
      )
      .join("\n");
    const blob = new Blob(["\uFEFF" + headers + rows], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `cheques-prepared-${new Date().toISOString().slice(0, 10)}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  };

  const handlePrint = () => {
    window.print();
  };

  const handleOpenAdd = () => {
    setEditingItem(null);
    setFormPayeeName("");
    setFormChequeNo("001024");
    setFormBookNo(INITIAL_BOOKS[0].accountNo);
    setFormAmount("50000.00");
    setFormDueDate("2026-10-25");
    setFormDescription("ชำระค่าสินค้าและบริการตามใบสั่งซื้อ");
    setFormRefDocNo("PO-6910-0060");
    setFormDepartment("ฝ่ายจัดซื้อและคลังสินค้า");
    setIsDialogOpen(true);
  };

  const handleOpenEdit = (item: PreparedChequeItem) => {
    setEditingItem(item);
    setFormPayeeName(item.payeeName);
    setFormChequeNo(item.chequeNo);
    setFormBookNo(item.bookNo);
    setFormAmount(item.amount);
    setFormDueDate(item.dueDate);
    setFormDescription(item.description);
    setFormRefDocNo(item.refDocNo);
    setFormDepartment(item.department);
    setIsDialogOpen(true);
  };

  const handleDelete = (id: string) => {
    setItems((prev) => prev.filter((i) => i.id !== id));
    setSelectedIds((prev) => prev.filter((i) => i !== id));
  };

  const handleSaveForm = (e: React.FormEvent) => {
    e.preventDefault();
    const matchedBook = books.find((b) => b.accountNo === formBookNo) || books[0];

    if (editingItem) {
      setItems((prev) =>
        prev.map((i) =>
          i.id === editingItem.id
            ? {
                ...i,
                payeeName: formPayeeName.trim(),
                chequeNo: formChequeNo.trim(),
                bookNo: formBookNo,
                bankCode: matchedBook.bankCode,
                bankName: matchedBook.bankName,
                branchName: matchedBook.branchName,
                amount: parseFloat(formAmount).toFixed(2),
                dueDate: formDueDate,
                description: formDescription.trim(),
                refDocNo: formRefDocNo.trim(),
                department: formDepartment.trim(),
              }
            : i
        )
      );
    } else {
      const newItem: PreparedChequeItem = {
        id: `cq-${Date.now()}`,
        docNo: `CQ6910-${(items.length + 1).toString().padStart(4, "0")}`,
        chequeNo: formChequeNo.trim(),
        bookNo: formBookNo,
        bankCode: matchedBook.bankCode,
        bankName: matchedBook.bankName,
        branchName: matchedBook.branchName,
        payDate: new Date().toISOString().slice(0, 10),
        dueDate: formDueDate,
        payeeCode: `AP-${(items.length + 1).toString().padStart(3, "0")}`,
        payeeName: formPayeeName.trim(),
        amount: parseFloat(formAmount || "0").toFixed(2),
        description: formDescription.trim(),
        department: formDepartment.trim(),
        refDocNo: formRefDocNo.trim(),
        status: "prepared",
        signatory1: "นายสมคิด รุ่งเรือง",
        signatory2: "นางวิภาดา จิตดี",
      };
      setItems((prev) => [newItem, ...prev]);
      setActiveItemId(newItem.id);
    }
    setIsDialogOpen(false);
  };

  const getStatusBadge = (status: PreparedChequeItem["status"]) => {
    switch (status) {
      case "prepared":
        return <span className="rounded-full bg-amber-500/10 px-2.5 py-0.5 text-xs font-semibold text-amber-600 dark:text-amber-400 border border-amber-500/20">เตรียมจ่าย</span>;
      case "approved":
        return <span className="rounded-full bg-blue-500/10 px-2.5 py-0.5 text-xs font-semibold text-blue-600 dark:text-blue-400 border border-blue-500/20">อนุมัติลงนาม</span>;
      case "issued":
        return <span className="rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-xs font-semibold text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">จ่ายแล้ว</span>;
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
            <CreditCard className="h-6 w-6" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">บันทึกเช็คเตรียมจ่าย</h1>
              <span className="rounded-md bg-muted px-2 py-0.5 text-xs font-semibold text-muted-foreground">
                Champ Parity 60022
              </span>
            </div>
            <p className="text-sm text-muted-foreground mt-0.5">
              จัดทำ ตรวจสอบยอดเงิน และบริหารเล่มเช็คก่อนการส่งมอบหรือสั่งจ่ายจริง (Cheques Issued Preparation)
            </p>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {selectedIds.length > 0 && (
            <Button
              variant="default"
              size="sm"
              onClick={handleBatchApprove}
              className="gap-2 text-xs bg-blue-600 hover:bg-blue-700 text-white"
            >
              <FileCheck className="h-3.5 w-3.5" />
              อนุมัติลงนามที่เลือก ({selectedIds.length})
            </Button>
          )}

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
            onClick={() => setItems(INITIAL_CHEQUES)}
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
            เพิ่มเช็คเตรียมจ่าย
          </Button>
        </div>
      </div>

      {/* Main Content Area with ResizableSplitter */}
      <div ref={containerRef} className="flex-1 min-h-0 flex flex-row rounded-2xl border border-border bg-card shadow-xs overflow-hidden relative">
        <div style={{ width: `${splitPercent}%` }} className="h-full flex flex-col border-r border-border/80 bg-muted/20 min-w-[240px] max-w-[500px] overflow-hidden">
              {/* Search & Book Filter Header */}
              <div className="p-3.5 border-b border-border/80 flex flex-col gap-2.5 bg-card/60">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    id={searchInputId}
                    placeholder="ค้นหาเลขเช็ค, สั่งจ่าย, วัตถุประสงค์..."
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
                    { id: "prepared", label: `เตรียมจ่าย (${stats.preparedCount})` },
                    { id: "approved", label: `อนุมัติ (${stats.approvedCount})` },
                    { id: "issued", label: `จ่ายแล้ว (${stats.issuedCount})` },
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

              {/* Cheque Book Register Selection */}
              <div className="p-2 border-b border-border/60 bg-muted/40 text-xs">
                <div className="px-2 py-1 font-semibold text-muted-foreground flex items-center justify-between">
                  <span>สมุดเงินฝาก / เล่มเช็ค</span>
                  <span className="text-[10px] text-primary cursor-pointer hover:underline" onClick={() => setSelectedBookNo("all")}>ดูทุกเล่ม</span>
                </div>
                <div className="space-y-1 mt-1">
                  {books.map((b) => {
                    const isSelected = selectedBookNo === b.accountNo;
                    return (
                      <button
                        key={b.id}
                        onClick={() => setSelectedBookNo(isSelected ? "all" : b.accountNo)}
                        className={`w-full text-left px-2.5 py-1.5 rounded-lg flex items-center justify-between text-xs transition-colors ${
                          isSelected
                            ? "bg-primary/10 text-primary font-semibold border border-primary/30"
                            : "hover:bg-muted text-foreground"
                        }`}
                      >
                        <div className="truncate">
                          <div>{b.bankName} - {b.accountNo}</div>
                          <div className="text-[10px] text-muted-foreground">{b.bookNo} (คงเหลือ {b.remainingCount} ฉบับ)</div>
                        </div>
                        <ChevronRight className="h-3.5 w-3.5 opacity-40 shrink-0 ml-1" />
                      </button>
                    );
                  })}
                </div>
              </div>

              {/* Master List of Cheques */}
              <div className="flex-1 overflow-y-auto p-2 space-y-1.5">
                {filteredItems.map((item) => {
                  const isSelected = activeItemId === item.id;
                  const isChecked = selectedIds.includes(item.id);

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
                        <div className="flex items-center gap-2">
                          <input
                            type="checkbox"
                            checked={isChecked}
                            onChange={(e) => {
                              e.stopPropagation();
                              handleToggleSelect(item.id);
                            }}
                            className="rounded border-border text-primary focus:ring-primary h-3.5 w-3.5"
                          />
                          <span className="font-mono font-bold text-xs text-foreground">
                            เช็คเลขที่ {item.chequeNo}
                          </span>
                        </div>
                        {getStatusBadge(item.status)}
                      </div>

                      <div className="mt-1 text-xs font-semibold text-foreground truncate">
                        {item.payeeName}
                      </div>

                      <div className="mt-1 flex items-center justify-between text-xs text-muted-foreground">
                        <span>ครบกำหนด: {item.dueDate}</span>
                        <span className="font-mono font-bold text-primary">
                          {parseFloat(item.amount).toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                        </span>
                      </div>
                    </div>
                  );
                })}

                {filteredItems.length === 0 && (
                  <div className="p-8 text-center text-xs text-muted-foreground">
                    ไม่พบรายการเช็คเตรียมจ่ายที่ตรงกับเงื่อนไข
                  </div>
                )}
              </div>

              {/* Left Bottom Summary */}
              <div className="p-3 border-t border-border/80 bg-card flex items-center justify-between text-xs">
                <span className="text-muted-foreground">ยอดรวมเช็คในรายการ:</span>
                <span className="font-mono font-bold text-foreground">{stats.totalAmountFormatted} ฿</span>
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
                  {/* Active Cheque Detail Card */}
                  <div className="rounded-xl border border-border bg-card p-5 shadow-xs space-y-4">
                    <div className="flex items-start justify-between gap-3 border-b border-border/70 pb-3.5">
                      <div>
                        <div className="flex items-center gap-2">
                          <h2 className="text-lg font-bold text-foreground">เช็คสั่งจ่ายเลขที่ {activeItem.chequeNo}</h2>
                          {getStatusBadge(activeItem.status)}
                        </div>
                        <p className="text-xs text-muted-foreground mt-0.5">
                          เลขที่ใบสำคัญ: <span className="font-mono">{activeItem.docNo}</span> · อ้างอิง: <span className="font-mono">{activeItem.refDocNo}</span>
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

                    {/* Cheque Front Graphic Box */}
                    <div className="rounded-xl border-2 border-dashed border-primary/30 bg-primary/5 p-5 space-y-3 font-mono">
                      <div className="flex items-start justify-between text-xs">
                        <div>
                          <div className="font-bold text-sm text-foreground">{activeItem.bankName}</div>
                          <div className="text-muted-foreground">{activeItem.branchName} · บัญชี {activeItem.bookNo}</div>
                        </div>
                        <div className="text-right">
                          <div className="text-[10px] text-muted-foreground">วันที่หน้าเช็ค (Due Date)</div>
                          <div className="font-bold text-foreground text-sm">{activeItem.dueDate}</div>
                        </div>
                      </div>

                      <div className="border-t border-primary/20 pt-2 flex items-center justify-between text-sm">
                        <div className="text-xs font-sans">
                          <span className="text-muted-foreground">จ่าย (Pay to): </span>
                          <span className="font-bold text-foreground">{activeItem.payeeName}</span>
                        </div>
                      </div>

                      <div className="flex items-center justify-between bg-card/80 p-3 rounded-lg border border-border/80 text-xs">
                        <div className="font-sans text-muted-foreground">
                          ตัวอักษร: <span className="font-bold text-foreground font-sans">({thaiBahtText(parseFloat(activeItem.amount))})</span>
                        </div>
                        <div className="font-mono text-base font-bold text-primary">
                          ***{parseFloat(activeItem.amount).toLocaleString("en-US", { minimumFractionDigits: 2 })}*** ฿
                        </div>
                      </div>

                      <div className="grid grid-cols-2 gap-4 pt-2 text-[11px] font-sans text-muted-foreground border-t border-primary/20">
                        <div>
                          <span>ผู้มีอำนาจลงนาม 1: </span>
                          <span className="font-semibold text-foreground">{activeItem.signatory1}</span>
                        </div>
                        <div>
                          <span>ผู้มีอำนาจลงนาม 2: </span>
                          <span className="font-semibold text-foreground">{activeItem.signatory2}</span>
                        </div>
                      </div>
                    </div>

                    {/* Metadata & Accounting Grid */}
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                      <div className="p-3 rounded-lg bg-muted/40 border border-border/60">
                        <span className="text-muted-foreground">วันที่จัดทำ</span>
                        <div className="font-semibold text-foreground mt-0.5">{activeItem.payDate}</div>
                      </div>
                      <div className="p-3 rounded-lg bg-muted/40 border border-border/60">
                        <span className="text-muted-foreground">แผนกผู้เบิก</span>
                        <div className="font-semibold text-foreground mt-0.5">{activeItem.department}</div>
                      </div>
                      <div className="p-3 rounded-lg bg-muted/40 border border-border/60">
                        <span className="text-muted-foreground">รหัสเจ้าหนี้</span>
                        <div className="font-semibold text-foreground mt-0.5">{activeItem.payeeCode}</div>
                      </div>
                      <div className="p-3 rounded-lg bg-muted/40 border border-border/60">
                        <span className="text-muted-foreground">ยอดเงินสุทธิ</span>
                        <div className="font-semibold font-mono text-primary mt-0.5">
                          {parseFloat(activeItem.amount).toLocaleString("en-US", { minimumFractionDigits: 2 })} บาท
                        </div>
                      </div>
                    </div>

                    {/* Description Notes */}
                    <div className="text-xs space-y-1">
                      <span className="text-muted-foreground font-semibold">วัตถุประสงค์ในการสั่งจ่าย:</span>
                      <p className="p-3 rounded-lg bg-muted/20 border border-border/60 text-foreground leading-relaxed">
                        {activeItem.description}
                      </p>
                    </div>
                  </div>

                  {/* Cheque Register Table */}
                  <div className="rounded-xl border border-border bg-card p-4 shadow-xs space-y-3">
                    <div className="flex items-center justify-between">
                      <h3 className="text-sm font-bold text-foreground">
                        รายการเช็คเตรียมจ่ายทั้งหมด ({filteredItems.length} ฉบับ)
                      </h3>
                      <div className="text-xs text-muted-foreground">
                        คลิกเพื่อเลือกดูและตรวจเช็ครายละเอียด
                      </div>
                    </div>

                    <div className="overflow-x-auto rounded-lg border border-border">
                      <table className="w-full text-xs text-left">
                        <thead className="bg-muted/50 border-b border-border text-muted-foreground font-semibold">
                          <tr>
                            <th className="p-2.5 w-8 text-center">
                              <input
                                type="checkbox"
                                checked={filteredItems.length > 0 && selectedIds.length === filteredItems.length}
                                onChange={(e) => handleSelectAll(e.target.checked)}
                                className="rounded border-border text-primary focus:ring-primary h-3.5 w-3.5"
                              />
                            </th>
                            <th className="p-2.5">เลขที่เช็ค</th>
                            <th className="p-2.5">วันที่หน้าเช็ค</th>
                            <th className="p-2.5">สั่งจ่ายให้</th>
                            <th className="p-2.5">ธนาคาร</th>
                            <th className="p-2.5 text-right">ยอดเงิน (บาท)</th>
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
                              <td className="p-2.5 text-center" onClick={(e) => e.stopPropagation()}>
                                <input
                                  type="checkbox"
                                  checked={selectedIds.includes(item.id)}
                                  onChange={() => handleToggleSelect(item.id)}
                                  className="rounded border-border text-primary focus:ring-primary h-3.5 w-3.5"
                                />
                              </td>
                              <td className="p-2.5 font-mono font-bold">{item.chequeNo}</td>
                              <td className="p-2.5">{item.dueDate}</td>
                              <td className="p-2.5 font-medium max-w-[200px] truncate">{item.payeeName}</td>
                              <td className="p-2.5 text-muted-foreground">{item.bankCode}</td>
                              <td className="p-2.5 text-right font-mono font-bold text-primary">
                                {parseFloat(item.amount).toLocaleString("en-US", { minimumFractionDigits: 2 })}
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
                  เลือกรายการเช็คทางด้านซ้ายเพื่อดูรายละเอียด
                </div>
              )}
            </div>
      </div>

      {/* Modal Dialog for Add / Edit Cheque */}
      {isDialogOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-xs p-4 animate-in fade-in">
          <div className="w-full max-w-lg rounded-2xl border border-border bg-card p-6 shadow-xl space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <h3 className="text-base font-bold text-foreground">
                {editingItem ? "แก้ไขเช็คเตรียมจ่าย" : "เพิ่มเช็คเตรียมจ่ายใหม่"}
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
                  <label className="block text-muted-foreground font-semibold mb-1">สมุดเงินฝาก / เล่มเช็ค</label>
                  <select
                    value={formBookNo}
                    onChange={(e) => setFormBookNo(e.target.value)}
                    className="w-full h-9 rounded-md border border-input bg-background px-3 text-xs"
                  >
                    {books.map((b) => (
                      <option key={b.id} value={b.accountNo}>
                        {b.bankName} ({b.accountNo})
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-muted-foreground font-semibold mb-1">เลขที่เช็ค</label>
                  <Input
                    required
                    value={formChequeNo}
                    onChange={(e) => setFormChequeNo(e.target.value)}
                    className="h-9 font-mono"
                    placeholder="เช่น 001025"
                  />
                </div>
              </div>

              <div>
                <label className="block text-muted-foreground font-semibold mb-1">สั่งจ่ายให้ (ชื่อผู้รับเงิน / เจ้าหนี้)</label>
                <Input
                  required
                  value={formPayeeName}
                  onChange={(e) => setFormPayeeName(e.target.value)}
                  className="h-9"
                  placeholder="เช่น บริษัท ซีเมนต์ไทย (ปูนใหญ่) มาร์เก็ตติ้ง จำกัด"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-muted-foreground font-semibold mb-1">ยอดเงินสั่งจ่าย (บาท)</label>
                  <Input
                    required
                    type="number"
                    step="0.01"
                    value={formAmount}
                    onChange={(e) => setFormAmount(e.target.value)}
                    className="h-9 font-mono font-bold"
                  />
                </div>
                <div>
                  <label className="block text-muted-foreground font-semibold mb-1">วันที่หน้าเช็ค (Due Date)</label>
                  <Input
                    required
                    type="date"
                    value={formDueDate}
                    onChange={(e) => setFormDueDate(e.target.value)}
                    className="h-9"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-muted-foreground font-semibold mb-1">อ้างอิงเอกสาร</label>
                  <Input
                    value={formRefDocNo}
                    onChange={(e) => setFormRefDocNo(e.target.value)}
                    className="h-9"
                    placeholder="เช่น PO-6910-0060"
                  />
                </div>
                <div>
                  <label className="block text-muted-foreground font-semibold mb-1">แผนก</label>
                  <Input
                    value={formDepartment}
                    onChange={(e) => setFormDepartment(e.target.value)}
                    className="h-9"
                  />
                </div>
              </div>

              <div>
                <label className="block text-muted-foreground font-semibold mb-1">วัตถุประสงค์ในการสั่งจ่าย</label>
                <textarea
                  rows={2}
                  value={formDescription}
                  onChange={(e) => setFormDescription(e.target.value)}
                  className="w-full rounded-md border border-input bg-background p-2.5 text-xs"
                  placeholder="ระบุรายละเอียด เช่น ชำระค่างวดที่ 1 ตามสัญญา"
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
                  บันทึกรายการเช็ค
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
