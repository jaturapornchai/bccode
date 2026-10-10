"use client";

import { useId, useMemo, useRef, useState } from "react";
import {
  CheckCircle2,
  Download,
  FileSpreadsheet,
  Printer,
  RefreshCw,
  Search,
  Sparkles,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export interface PendingInvoiceItem {
  id: string;
  invoiceNo: string;
  invoiceDate: string;
  dueDate: string;
  customerCode: string;
  customerName: string;
  amount: string;
  vatAmount: string;
  netTotal: string;
  salesperson: string;
  billingCycle: string;
  creditTermDays: number;
}

export interface GeneratedBillingNote {
  docNo: string;
  docDate: string;
  dueDate: string;
  customerCode: string;
  customerName: string;
  invoiceCount: number;
  totalAmount: string;
  status: "created" | "printed";
}

const INITIAL_PENDING_INVOICES: PendingInvoiceItem[] = [
  {
    id: "inv-01",
    invoiceNo: "INV6910-0012",
    invoiceDate: "2026-10-01",
    dueDate: "2026-10-31",
    customerCode: "AR-001",
    customerName: "บริษัท สยามคอนกรีตมิกซ์ จำกัด",
    amount: "45000.00",
    vatAmount: "3150.00",
    netTotal: "48150.00",
    salesperson: "นายสมศักดิ์ ขยันขาย",
    billingCycle: "สิ้นเดือน",
    creditTermDays: 30,
  },
  {
    id: "inv-02",
    invoiceNo: "INV6910-0015",
    invoiceDate: "2026-10-05",
    dueDate: "2026-10-31",
    customerCode: "AR-001",
    customerName: "บริษัท สยามคอนกรีตมิกซ์ จำกัด",
    amount: "72000.00",
    vatAmount: "5040.00",
    netTotal: "77040.00",
    salesperson: "นายสมศักดิ์ ขยันขาย",
    billingCycle: "สิ้นเดือน",
    creditTermDays: 30,
  },
  {
    id: "inv-03",
    invoiceNo: "INV6910-0018",
    invoiceDate: "2026-10-08",
    dueDate: "2026-10-31",
    customerCode: "AR-001",
    customerName: "บริษัท สยามคอนกรีตมิกซ์ จำกัด",
    amount: "38500.00",
    vatAmount: "2695.00",
    netTotal: "41195.00",
    salesperson: "นายสมศักดิ์ ขยันขาย",
    billingCycle: "สิ้นเดือน",
    creditTermDays: 30,
  },
  {
    id: "inv-04",
    invoiceNo: "INV6910-0021",
    invoiceDate: "2026-10-02",
    dueDate: "2026-10-17",
    customerCode: "AR-002",
    customerName: "ห้างหุ้นส่วนจำกัด นครหลวงวิศวกรรม",
    amount: "115000.00",
    vatAmount: "8050.00",
    netTotal: "123050.00",
    salesperson: "นางสาวศิริพร บุญช่วย",
    billingCycle: "ทุกวันที่ 15",
    creditTermDays: 15,
  },
  {
    id: "inv-05",
    invoiceNo: "INV6910-0025",
    invoiceDate: "2026-10-06",
    dueDate: "2026-10-21",
    customerCode: "AR-002",
    customerName: "ห้างหุ้นส่วนจำกัด นครหลวงวิศวกรรม",
    amount: "64000.00",
    vatAmount: "4480.00",
    netTotal: "68480.00",
    salesperson: "นางสาวศิริพร บุญช่วย",
    billingCycle: "ทุกวันที่ 15",
    creditTermDays: 15,
  },
  {
    id: "inv-06",
    invoiceNo: "INV6910-0028",
    invoiceDate: "2026-10-04",
    dueDate: "2026-11-03",
    customerCode: "AR-003",
    customerName: "บริษัท ธนบุรีอินดัสเตรียล จำกัด",
    amount: "89000.00",
    vatAmount: "6230.00",
    netTotal: "95230.00",
    salesperson: "นายมานพ มั่นคง",
    billingCycle: "สิ้นเดือน",
    creditTermDays: 30,
  },
  {
    id: "inv-07",
    invoiceNo: "INV6910-0032",
    invoiceDate: "2026-10-09",
    dueDate: "2026-11-08",
    customerCode: "AR-004",
    customerName: "บริษัท สหรุ่งโรจน์ ดีเวลลอปเมนท์ จำกัด",
    amount: "53000.00",
    vatAmount: "3710.00",
    netTotal: "56710.00",
    salesperson: "นายมานพ มั่นคง",
    billingCycle: "สิ้นเดือน",
    creditTermDays: 30,
  },
];

interface BillingNoteAutoScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function BillingNoteAutoScreen({ language: _language = "th", holdingcode: _holdingcode = "" }: BillingNoteAutoScreenProps) {
  const [invoices, setInvoices] = useState<PendingInvoiceItem[]>(INITIAL_PENDING_INVOICES);
  const [selectedCustomerCode, setSelectedCustomerCode] = useState<string>("AR-001");
  const [cycleFilter, setCycleFilter] = useState<string>("all");
  const [search, setSearch] = useState("");
  const [selectedInvoiceIds, setSelectedInvoiceIds] = useState<string[]>([]);
  const [generatedNotes, setGeneratedNotes] = useState<GeneratedBillingNote[]>([
    {
      docNo: "BN6910-0001",
      docDate: "2026-10-01",
      dueDate: "2026-10-31",
      customerCode: "AR-001",
      customerName: "บริษัท สยามคอนกรีตมิกซ์ จำกัด",
      invoiceCount: 2,
      totalAmount: "125000.00",
      status: "printed",
    },
  ]);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  // Form Fields for Generation
  const [billingDate, setBillingDate] = useState("2026-10-15");
  const [paymentDueDate, setPaymentDueDate] = useState("2026-10-31");
  const [billingRemark, setBillingRemark] = useState("วางบิลค่าสินค้าตามรอบประจำงวด");

  const searchInputId = useId();
  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    defaultLeft: 34,
    min: 25,
    max: 45,
    storageKey: "bc_billing_note_auto_split",
    containerRef,
  });

  // Grouped Customers with Pending Invoices
  const customerSummaries = useMemo(() => {
    const map = new Map<string, { code: string; name: string; cycle: string; count: number; total: number }>();

    for (const inv of invoices) {
      if (cycleFilter !== "all" && inv.billingCycle !== cycleFilter) continue;
      const q = search.trim().toLowerCase();
      if (q && !inv.customerName.toLowerCase().includes(q) && !inv.customerCode.toLowerCase().includes(q)) continue;

      const curr = map.get(inv.customerCode) || {
        code: inv.customerCode,
        name: inv.customerName,
        cycle: inv.billingCycle,
        count: 0,
        total: 0,
      };
      curr.count += 1;
      curr.total += parseFloat(inv.netTotal) || 0;
      map.set(inv.customerCode, curr);
    }

    return Array.from(map.values());
  }, [invoices, cycleFilter, search]);

  // Pending Invoices for Currently Selected Customer
  const currentCustomerInvoices = useMemo(() => {
    return invoices.filter((i) => i.customerCode === selectedCustomerCode);
  }, [invoices, selectedCustomerCode]);

  const activeCustomer = useMemo(() => {
    return customerSummaries.find((c) => c.code === selectedCustomerCode) || customerSummaries[0] || null;
  }, [customerSummaries, selectedCustomerCode]);

  // Overall Statistics
  const overallStats = useMemo(() => {
    let count = 0;
    let sum = 0;
    for (const inv of invoices) {
      count++;
      sum += parseFloat(inv.netTotal) || 0;
    }
    return {
      totalCount: count,
      totalSumFormatted: sum.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
    };
  }, [invoices]);

  const handleSelectAllCurrent = (checked: boolean) => {
    if (checked) {
      setSelectedInvoiceIds(currentCustomerInvoices.map((i) => i.id));
    } else {
      setSelectedInvoiceIds([]);
    }
  };

  const handleToggleSelectInvoice = (id: string) => {
    setSelectedInvoiceIds((prev) =>
      prev.includes(id) ? prev.filter((i) => i !== id) : [...prev, id]
    );
  };

  const handleGenerateBillingNote = () => {
    if (selectedInvoiceIds.length === 0 || !activeCustomer) return;

    const selectedList = currentCustomerInvoices.filter((i) => selectedInvoiceIds.includes(i.id));
    const totalAmount = selectedList.reduce((sum, item) => sum + parseFloat(item.netTotal), 0);

    const newDocNo = `BN6910-${(generatedNotes.length + 1).toString().padStart(4, "0")}`;
    const newNote: GeneratedBillingNote = {
      docNo: newDocNo,
      docDate: billingDate,
      dueDate: paymentDueDate,
      customerCode: activeCustomer.code,
      customerName: activeCustomer.name,
      invoiceCount: selectedList.length,
      totalAmount: totalAmount.toFixed(2),
      status: "created",
    };

    setGeneratedNotes((prev) => [newNote, ...prev]);
    // Remove generated invoices from pending
    setInvoices((prev) => prev.filter((i) => !selectedInvoiceIds.includes(i.id)));
    setSelectedInvoiceIds([]);

    setSuccessMessage(`สร้างใบวางบิล ${newDocNo} สำหรับ ${activeCustomer.name} สำเร็จ (${selectedList.length} รายการ, ${totalAmount.toLocaleString("en-US", { minimumFractionDigits: 2 })} บาท)`);
    setTimeout(() => setSuccessMessage(null), 5000);
  };

  const handleExportCSV = () => {
    const headers = "เลขที่ใบแจ้งหนี้,วันที่,วันครบกำหนด,รหัสลูกค้า,ชื่อลูกค้า,ยอดก่อนภาษี,VAT,ยอดสุทธิ,รอบวางบิล,พนักงานขาย\n";
    const rows = invoices
      .map((i) =>
        `"${i.invoiceNo}","${i.invoiceDate}","${i.dueDate}","${i.customerCode}","${i.customerName}","${i.amount}","${i.vatAmount}","${i.netTotal}","${i.billingCycle}","${i.salesperson}"`
      )
      .join("\n");
    const blob = new Blob(["\uFEFF" + headers + rows], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `pending-invoices-for-billing-${new Date().toISOString().slice(0, 10)}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  };

  const handlePrint = () => {
    window.print();
  };

  return (
    <div className="flex-1 min-h-0 flex flex-col gap-4 p-4 lg:p-6 overflow-hidden">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-5 shadow-xs">
        <div className="flex items-center gap-3.5">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-inner">
            <FileSpreadsheet className="h-6 w-6" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">บันทึกใบวางบิลอัตโนมัติ</h1>
              <span className="rounded-md bg-muted px-2 py-0.5 text-xs font-semibold text-muted-foreground">
                Champ Parity 30005
              </span>
            </div>
            <p className="text-sm text-muted-foreground mt-0.5">
              ประมวลผลและสร้างใบวางบิลอัตโนมัติจากใบกำกับภาษีที่ยังไม่ได้วางบิล แยกตามรอบวางบิลของลูกค้าแต่ละราย (Automatic Billing Note)
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
            onClick={() => setInvoices(INITIAL_PENDING_INVOICES)}
            className="gap-2 text-xs"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            คืนค่าเริ่มต้น
          </Button>

          <Button
            variant="default"
            size="sm"
            disabled={selectedInvoiceIds.length === 0}
            onClick={handleGenerateBillingNote}
            className="gap-2 text-xs bg-emerald-600 hover:bg-emerald-700 text-white"
          >
            <Sparkles className="h-3.5 w-3.5" />
            สร้างใบวางบิล ({selectedInvoiceIds.length} บิล)
          </Button>
        </div>
      </div>

      {successMessage && (
        <div className="flex items-center gap-2.5 rounded-xl border border-emerald-300 dark:border-emerald-800 bg-emerald-50 dark:bg-emerald-950/40 p-4 text-xs font-semibold text-emerald-800 dark:text-emerald-300 animate-in fade-in">
          <CheckCircle2 className="h-4 w-4 shrink-0" />
          <span>{successMessage}</span>
        </div>
      )}

      {/* Main Content Area with ResizableSplitter */}
      <div ref={containerRef} className="flex-1 min-h-0 flex flex-row rounded-2xl border border-border bg-card shadow-xs overflow-hidden relative">
        {/* Left Column: Customer Summaries */}
        <div
          style={{ width: `${splitPercent}%` }}
          className="h-full flex flex-col border-r border-border/80 bg-muted/20 min-w-[250px] max-w-[500px] overflow-hidden"
        >
          {/* Search Box */}
          <div className="p-3.5 border-b border-border/80 flex flex-col gap-2.5 bg-card/60">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                id={searchInputId}
                placeholder="ค้นหาชื่อลูกค้า, รหัส..."
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

            {/* Cycle Filter */}
            <div className="flex items-center gap-1 overflow-x-auto pb-1 text-xs">
              {[
                { id: "all", label: "ทุกรอบ" },
                { id: "สิ้นเดือน", label: "สิ้นเดือน" },
                { id: "ทุกวันที่ 15", label: "ทุกวันที่ 15" },
              ].map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setCycleFilter(tab.id)}
                  className={`px-2.5 py-1 rounded-lg font-medium whitespace-nowrap transition-colors ${
                    cycleFilter === tab.id
                      ? "bg-primary text-primary-foreground shadow-xs"
                      : "text-muted-foreground hover:bg-muted/70 hover:text-foreground"
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {/* Master List of Customers */}
          <div className="flex-1 overflow-y-auto p-2 space-y-1.5">
            {customerSummaries.map((cust) => {
              const isSelected = selectedCustomerCode === cust.code;

              return (
                <div
                  key={cust.code}
                  onClick={() => {
                    setSelectedCustomerCode(cust.code);
                    setSelectedInvoiceIds([]);
                  }}
                  className={`group relative rounded-xl border p-3 cursor-pointer transition-all ${
                    isSelected
                      ? "border-primary bg-primary/5 shadow-xs ring-1 ring-primary/20"
                      : "border-border/60 bg-card hover:border-primary/40 hover:bg-muted/30"
                  }`}
                >
                  <div className="flex items-start justify-between gap-2">
                    <span className="font-mono font-bold text-xs text-foreground">
                      {cust.code}
                    </span>
                    <span className="rounded-full bg-primary/10 text-primary px-2 py-0.5 text-[10px] font-semibold border border-primary/20">
                      รอบ: {cust.cycle}
                    </span>
                  </div>

                  <div className="mt-1 text-xs font-semibold text-foreground truncate">
                    {cust.name}
                  </div>

                  <div className="mt-1 flex items-center justify-between text-xs text-muted-foreground">
                    <span>{cust.count} บิลรอวาง</span>
                    <span className="font-mono font-bold text-primary">
                      {cust.total.toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                    </span>
                  </div>
                </div>
              );
            })}

            {customerSummaries.length === 0 && (
              <div className="p-8 text-center text-xs text-muted-foreground">
                ไม่มีบิลค้างวางบิลสำหรับรอบที่เลือก
              </div>
            )}
          </div>

          {/* Left Bottom Summary */}
          <div className="p-3 border-t border-border/80 bg-card flex items-center justify-between text-xs">
            <span className="text-muted-foreground">บิลค้างวางทั้งหมด:</span>
            <span className="font-mono font-bold text-foreground">
              {overallStats.totalCount} บิล ({overallStats.totalSumFormatted} ฿)
            </span>
          </div>
        </div>

        {/* Resizable Splitter Handle */}
        <ResizableSplitter
          value={splitPercent}
          isResizing={isResizing}
          onPointerDown={startResize}
          onKeyDown={adjustWithKeyboard}
        />

        {/* Right Column: Invoice Details & Generation Form */}
        <div className="flex-1 min-w-0 h-full flex flex-col overflow-y-auto">
          {activeCustomer ? (
            <div className="p-5 flex flex-col gap-5">
              {/* Customer Header Box */}
              <div className="rounded-xl border border-border bg-card p-5 shadow-xs space-y-4">
                <div className="flex items-start justify-between gap-3 border-b border-border/70 pb-3.5">
                  <div>
                    <div className="flex items-center gap-2">
                      <h2 className="text-lg font-bold text-foreground">{activeCustomer.name}</h2>
                      <span className="font-mono text-xs text-muted-foreground">({activeCustomer.code})</span>
                    </div>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      รอบวางบิล: <span className="font-semibold text-foreground">{activeCustomer.cycle}</span> · บิลค้างวาง: <span className="font-semibold text-primary">{currentCustomerInvoices.length} รายการ</span>
                    </p>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleSelectAllCurrent(selectedInvoiceIds.length !== currentCustomerInvoices.length)}
                      className="text-xs h-8"
                    >
                      {selectedInvoiceIds.length === currentCustomerInvoices.length ? "ยกเลิกเลือกทั้งหมด" : "เลือกทั้งหมด"}
                    </Button>
                  </div>
                </div>

                {/* Generation Parameters */}
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs bg-muted/20 p-3.5 rounded-xl border border-border/70">
                  <div>
                    <label className="block text-muted-foreground font-semibold mb-1">วันที่ในใบวางบิล</label>
                    <Input
                      type="date"
                      value={billingDate}
                      onChange={(e) => setBillingDate(e.target.value)}
                      className="h-8 text-xs bg-background"
                    />
                  </div>
                  <div>
                    <label className="block text-muted-foreground font-semibold mb-1">วันนัดชำระเงิน</label>
                    <Input
                      type="date"
                      value={paymentDueDate}
                      onChange={(e) => setPaymentDueDate(e.target.value)}
                      className="h-8 text-xs bg-background"
                    />
                  </div>
                  <div>
                    <label className="block text-muted-foreground font-semibold mb-1">หมายเหตุกำกับ</label>
                    <Input
                      value={billingRemark}
                      onChange={(e) => setBillingRemark(e.target.value)}
                      className="h-8 text-xs bg-background"
                    />
                  </div>
                </div>

                {/* Table of Invoices */}
                <div className="overflow-x-auto rounded-lg border border-border">
                  <table className="w-full text-xs text-left">
                    <thead className="bg-muted/50 border-b border-border text-muted-foreground font-semibold">
                      <tr>
                        <th className="p-2.5 w-8 text-center">
                          <input
                            type="checkbox"
                            checked={currentCustomerInvoices.length > 0 && selectedInvoiceIds.length === currentCustomerInvoices.length}
                            onChange={(e) => handleSelectAllCurrent(e.target.checked)}
                            className="rounded border-border text-primary focus:ring-primary h-3.5 w-3.5"
                          />
                        </th>
                        <th className="p-2.5">เลขที่ใบกำกับภาษี</th>
                        <th className="p-2.5">วันที่ใบกำกับ</th>
                        <th className="p-2.5">วันครบกำหนด</th>
                        <th className="p-2.5">เครดิต</th>
                        <th className="p-2.5 text-right">ยอดก่อนภาษี</th>
                        <th className="p-2.5 text-right">VAT 7%</th>
                        <th className="p-2.5 text-right">ยอดสุทธิ (บาท)</th>
                        <th className="p-2.5">พนักงานขาย</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                      {currentCustomerInvoices.map((inv) => {
                        const isChecked = selectedInvoiceIds.includes(inv.id);
                        return (
                          <tr
                            key={inv.id}
                            onClick={() => handleToggleSelectInvoice(inv.id)}
                            className={`cursor-pointer transition-colors ${
                              isChecked ? "bg-primary/10" : "hover:bg-muted/40"
                            }`}
                          >
                            <td className="p-2.5 text-center" onClick={(e) => e.stopPropagation()}>
                              <input
                                type="checkbox"
                                checked={isChecked}
                                onChange={() => handleToggleSelectInvoice(inv.id)}
                                className="rounded border-border text-primary focus:ring-primary h-3.5 w-3.5"
                              />
                            </td>
                            <td className="p-2.5 font-mono font-bold text-foreground">{inv.invoiceNo}</td>
                            <td className="p-2.5">{inv.invoiceDate}</td>
                            <td className="p-2.5">{inv.dueDate}</td>
                            <td className="p-2.5">{inv.creditTermDays} วัน</td>
                            <td className="p-2.5 text-right font-mono">
                              {parseFloat(inv.amount).toLocaleString("en-US", { minimumFractionDigits: 2 })}
                            </td>
                            <td className="p-2.5 text-right font-mono text-muted-foreground">
                              {parseFloat(inv.vatAmount).toLocaleString("en-US", { minimumFractionDigits: 2 })}
                            </td>
                            <td className="p-2.5 text-right font-mono font-bold text-primary">
                              {parseFloat(inv.netTotal).toLocaleString("en-US", { minimumFractionDigits: 2 })}
                            </td>
                            <td className="p-2.5 text-muted-foreground truncate max-w-[120px]">{inv.salesperson}</td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>

                {/* Selected Summary and Generate Action */}
                <div className="flex flex-wrap items-center justify-between gap-4 p-4 rounded-xl border border-primary/20 bg-primary/5">
                  <div className="text-xs">
                    <span className="text-muted-foreground">รายการที่เลือกเพื่อวางบิล: </span>
                    <span className="font-bold text-foreground">{selectedInvoiceIds.length} จาก {currentCustomerInvoices.length} บิล</span>
                    <span className="mx-2 text-muted-foreground">|</span>
                    <span className="text-muted-foreground">ยอดรวมทั้งสิ้น: </span>
                    <span className="font-mono text-base font-bold text-primary">
                      {currentCustomerInvoices
                        .filter((i) => selectedInvoiceIds.includes(i.id))
                        .reduce((sum, item) => sum + parseFloat(item.netTotal), 0)
                        .toLocaleString("en-US", { minimumFractionDigits: 2 })} ฿
                    </span>
                  </div>

                  <Button
                    size="sm"
                    disabled={selectedInvoiceIds.length === 0}
                    onClick={handleGenerateBillingNote}
                    className="gap-2 text-xs bg-primary hover:bg-primary/90 text-primary-foreground font-semibold px-4"
                  >
                    <CheckCircle2 className="h-4 w-4" />
                    ยืนยันสร้างใบวางบิล
                  </Button>
                </div>
              </div>

              {/* History of Generated Billing Notes */}
              <div className="rounded-xl border border-border bg-card p-4 shadow-xs space-y-3">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-bold text-foreground">
                    ใบวางบิลที่สร้างแล้วล่าสุด ({generatedNotes.length} ฉบับ)
                  </h3>
                  <span className="text-xs text-muted-foreground">
                    สามารถส่งต่อไปพิมพ์ที่หน้าจอ พิมพ์ใบวางบิลอัตโนมัติ
                  </span>
                </div>

                <div className="overflow-x-auto rounded-lg border border-border">
                  <table className="w-full text-xs text-left">
                    <thead className="bg-muted/50 border-b border-border text-muted-foreground font-semibold">
                      <tr>
                        <th className="p-2.5">เลขที่ใบวางบิล</th>
                        <th className="p-2.5">วันที่วางบิล</th>
                        <th className="p-2.5">วันนัดชำระ</th>
                        <th className="p-2.5">ลูกค้า</th>
                        <th className="p-2.5 text-center">จำนวนบิล</th>
                        <th className="p-2.5 text-right">ยอดรวม (บาท)</th>
                        <th className="p-2.5 text-center">สถานะ</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                      {generatedNotes.map((note) => (
                        <tr key={note.docNo} className="hover:bg-muted/40">
                          <td className="p-2.5 font-mono font-bold text-foreground">{note.docNo}</td>
                          <td className="p-2.5">{note.docDate}</td>
                          <td className="p-2.5">{note.dueDate}</td>
                          <td className="p-2.5 font-medium">{note.customerName}</td>
                          <td className="p-2.5 text-center">{note.invoiceCount}</td>
                          <td className="p-2.5 text-right font-mono font-bold text-primary">
                            {parseFloat(note.totalAmount).toLocaleString("en-US", { minimumFractionDigits: 2 })}
                          </td>
                          <td className="p-2.5 text-center">
                            <span className="rounded-full bg-emerald-500/10 text-emerald-600 px-2.5 py-0.5 text-xs font-semibold border border-emerald-500/20">
                              สร้างสำเร็จ
                            </span>
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
              เลือกลูกค้าทางด้านซ้ายเพื่อดูรายการใบกำกับภาษีที่ค้างวางบิล
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
