"use client";

import { useMemo, useRef, useState } from "react";
import {
  AlertCircle,
  Building2,
  Calendar,
  CheckCircle2,
  Clock,
  CreditCard,
  DollarSign,
  FileCheck,
  FileSpreadsheet,
  FileText,
  Filter,
  Phone,
  RefreshCw,
  Search,
  ShieldAlert,
  UserCheck,
  Users,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export type CustomerStatus = "active" | "credit_hold" | "inactive";
export type DebtorGroupType = "corporate" | "retail" | "wholesale" | "project" | "government";

export interface CustomerTransactionRecord {
  id: string;
  docNo: string;
  docType: string;
  date: string;
  amount: string; // minor unit formatted string
  status: "paid" | "pending" | "overdue";
  dueDate: string;
}

export interface CustomerInquiryItem {
  id: string;
  code: string;
  nameTh: string;
  nameEn: string;
  taxId: string;
  branchCode: string;
  branchName: string;
  debtorGroup: DebtorGroupType;
  debtorGroupName: string;
  contactPerson: string;
  phone: string;
  email: string;
  address: string;
  paymentTermDays: number;
  creditLimit: string;
  outstandingBalance: string;
  availableCredit: string;
  status: CustomerStatus;
  priceTier: string;
  tradeDiscountPct: string;
  aging: {
    current: string;
    days1_30: string;
    days31_60: string;
    days61_90: string;
    over90: string;
  };
  transactions: CustomerTransactionRecord[];
}

const DEFAULT_CUSTOMERS: CustomerInquiryItem[] = [
  {
    id: "cust-001",
    code: "CUST-0001",
    nameTh: "บริษัท สยามก่อสร้างรุ่งเรือง จำกัด",
    nameEn: "Siam Construction Rungrueng Co., Ltd.",
    taxId: "0105556012345",
    branchCode: "00000",
    branchName: "สำนักงานใหญ่",
    debtorGroup: "corporate",
    debtorGroupName: "ลูกหนี้นิติบุคคลทั่วไป",
    contactPerson: "คุณสมศักดิ์ มั่งมี (ฝ่ายจัดซื้อ)",
    phone: "02-589-4321, 081-456-7890",
    email: "procurement@siamconstruction.co.th",
    address: "128/45 ถนนวิภาวดีรังสิต แขวงจอมพล เขตจตุจักร กรุงเทพมหานคร 10900",
    paymentTermDays: 30,
    creditLimit: "1,500,000.00",
    outstandingBalance: "425,000.00",
    availableCredit: "1,075,000.00",
    status: "active",
    priceTier: "ราคาขายส่งระดับ 1 (Wholesale Tier 1)",
    tradeDiscountPct: "5.00%",
    aging: {
      current: "250,000.00",
      days1_30: "175,000.00",
      days31_60: "0.00",
      days61_90: "0.00",
      over90: "0.00",
    },
    transactions: [
      { id: "tx-101", docNo: "IV-202610-0082", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-10-04", amount: "125,000.00", status: "pending", dueDate: "2026-11-03" },
      { id: "tx-102", docNo: "IV-202609-0215", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-09-22", amount: "175,000.00", status: "pending", dueDate: "2026-10-22" },
      { id: "tx-103", docNo: "RC-202609-0044", docType: "ใบเสร็จรับเงิน/ใบกำกับภาษี", date: "2026-09-15", amount: "320,000.00", status: "paid", dueDate: "2026-09-15" },
      { id: "tx-104", docNo: "QT-202610-0019", docType: "ใบเสนอราคา", date: "2026-10-08", amount: "89,000.00", status: "pending", dueDate: "2026-11-07" },
    ],
  },
  {
    id: "cust-002",
    code: "CUST-0002",
    nameTh: "ห้างหุ้นส่วนจำกัด บางกอกวัสดุก่อสร้างภัณฑ์",
    nameEn: "Bangkok Construction Supplies Ltd., Part.",
    taxId: "0103554098765",
    branchCode: "00001",
    branchName: "สาขาบางนา",
    debtorGroup: "wholesale",
    debtorGroupName: "ตัวแทนจำหน่าย/ค้าส่ง",
    contactPerson: "คุณวิภาวรรณ สดใส (ผู้จัดการ)",
    phone: "02-744-8900, 089-123-4567",
    email: "contact@bangkoksupplies.com",
    address: "55/8 หมู่ 12 ถนนบางนา-ตราด กม.8 ต.บางพลีใหญ่ อ.บางพลี จ.สมุทรปราการ 10540",
    paymentTermDays: 45,
    creditLimit: "800,000.00",
    outstandingBalance: "180,500.00",
    availableCredit: "619,500.00",
    status: "active",
    priceTier: "ราคาตัวแทนจำหน่าย (Dealer Price)",
    tradeDiscountPct: "7.50%",
    aging: {
      current: "110,500.00",
      days1_30: "70,000.00",
      days31_60: "0.00",
      days61_90: "0.00",
      over90: "0.00",
    },
    transactions: [
      { id: "tx-201", docNo: "IV-202610-0045", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-10-02", amount: "110,500.00", status: "pending", dueDate: "2026-11-16" },
      { id: "tx-202", docNo: "IV-202609-0112", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-09-10", amount: "70,000.00", status: "pending", dueDate: "2026-10-25" },
      { id: "tx-203", docNo: "RC-202608-0199", docType: "ใบเสร็จรับเงิน", date: "2026-08-30", amount: "150,000.00", status: "paid", dueDate: "2026-08-30" },
    ],
  },
  {
    id: "cust-003",
    code: "CUST-0003",
    nameTh: "บริษัท ธนบูรณ์วิศวกรรมการช่าง จำกัด",
    nameEn: "Thanaboon Engineering & Construction Co., Ltd.",
    taxId: "0105561044556",
    branchCode: "00000",
    branchName: "สำนักงานใหญ่",
    debtorGroup: "project",
    debtorGroupName: "ลูกค้าโครงการ (Project)",
    contactPerson: "วิศวกรโครงการ คุณพิชัย (ไซต์งานพระราม 9)",
    phone: "02-930-1122, 086-778-9900",
    email: "site.rama9@thanaboon.co.th",
    address: "99 อาคารธนบูรณ์ ทาวเวอร์ ชั้น 14 ถนนพระราม 9 แขวงห้วยขวาง เขตห้วยขวาง กรุงเทพมหานคร 10310",
    paymentTermDays: 60,
    creditLimit: "2,000,000.00",
    outstandingBalance: "1,980,000.00",
    availableCredit: "20,000.00",
    status: "credit_hold",
    priceTier: "ราคาโครงการพิเศษ (Project Tier A)",
    tradeDiscountPct: "8.00%",
    aging: {
      current: "600,000.00",
      days1_30: "580,000.00",
      days31_60: "500,000.00",
      days61_90: "300,000.00",
      over90: "0.00",
    },
    transactions: [
      { id: "tx-301", docNo: "IV-202608-0091", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-08-15", amount: "300,000.00", status: "overdue", dueDate: "2026-10-14" },
      { id: "tx-302", docNo: "IV-202609-0022", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-09-05", amount: "500,000.00", status: "pending", dueDate: "2026-11-04" },
      { id: "tx-303", docNo: "IV-202609-0199", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-09-28", amount: "580,000.00", status: "pending", dueDate: "2026-11-27" },
      { id: "tx-304", docNo: "IV-202610-0012", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-10-01", amount: "600,000.00", status: "pending", dueDate: "2026-11-30" },
    ],
  },
  {
    id: "cust-004",
    code: "CUST-0004",
    nameTh: "บริษัท พัฒนายนต์ เอ็กซ์เพรส ทรานสปอร์ต จำกัด",
    nameEn: "Phatthanayont Express Transport Co., Ltd.",
    taxId: "0105559077889",
    branchCode: "00000",
    branchName: "สำนักงานใหญ่",
    debtorGroup: "corporate",
    debtorGroupName: "ลูกหนี้นิติบุคคลทั่วไป",
    contactPerson: "คุณกมลวรรณ (ฝ่ายบัญชี)",
    phone: "02-318-7700",
    email: "account@phatthanayont.co.th",
    address: "214 ถนนพัฒนาการ แขวงสวนหลวง เขตสวนหลวง กรุงเทพมหานคร 10250",
    paymentTermDays: 30,
    creditLimit: "500,000.00",
    outstandingBalance: "0.00",
    availableCredit: "500,000.00",
    status: "active",
    priceTier: "ราคามาตรฐาน (Standard Retail)",
    tradeDiscountPct: "0.00%",
    aging: {
      current: "0.00",
      days1_30: "0.00",
      days31_60: "0.00",
      days61_90: "0.00",
      over90: "0.00",
    },
    transactions: [
      { id: "tx-401", docNo: "RC-202609-0188", docType: "ใบเสร็จรับเงิน/ใบกำกับภาษี", date: "2026-09-25", amount: "48,500.00", status: "paid", dueDate: "2026-09-25" },
    ],
  },
  {
    id: "cust-005",
    code: "CUST-0005",
    nameTh: "สหกรณ์การเกษตรเมืองทองเพื่อการพัฒนา จำกัด",
    nameEn: "Muangthong Agricultural Cooperative Ltd.",
    taxId: "0994000123456",
    branchCode: "00000",
    branchName: "สำนักงานใหญ่",
    debtorGroup: "government",
    debtorGroupName: "หน่วยงานรัฐ/สหกรณ์",
    contactPerson: "คุณประเสริฐ (หัวหน้าฝ่ายจัดซื้อ)",
    phone: "035-241-890, 081-998-1122",
    email: "procure@muangthong-coop.or.th",
    address: "44 หมู่ 3 ถนนสายเอเชีย ต.บ้านกรด อ.บางปะอิน จ.พระนครศรีอยุธยา 13160",
    paymentTermDays: 60,
    creditLimit: "1,000,000.00",
    outstandingBalance: "120,000.00",
    availableCredit: "880,000.00",
    status: "active",
    priceTier: "ราคาหน่วยงานราชการ/สหกรณ์",
    tradeDiscountPct: "4.00%",
    aging: {
      current: "120,000.00",
      days1_30: "0.00",
      days31_60: "0.00",
      days61_90: "0.00",
      over90: "0.00",
    },
    transactions: [
      { id: "tx-501", docNo: "IV-202610-0071", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-10-05", amount: "120,000.00", status: "pending", dueDate: "2026-12-04" },
    ],
  },
  {
    id: "cust-006",
    code: "CUST-0006",
    nameTh: "ห้างหุ้นส่วนจำกัด รัตนวิศวกรรมการค้า",
    nameEn: "Rattana Engineering Trade Ltd., Part.",
    taxId: "0103557088990",
    branchCode: "00000",
    branchName: "สำนักงานใหญ่",
    debtorGroup: "retail",
    debtorGroupName: "ผู้รับเหมาและร้านค้าย่อย",
    contactPerson: "ช่างสมหมาย รัตนชัย",
    phone: "084-556-7788",
    email: "rattana.eng@gmail.com",
    address: "88/12 ซอยรามคำแหง 53 แขวงพลับพลา เขตวังทองหลาง กรุงเทพมหานคร 10310",
    paymentTermDays: 15,
    creditLimit: "300,000.00",
    outstandingBalance: "310,000.00",
    availableCredit: "0.00",
    status: "credit_hold",
    priceTier: "ราคาผู้รับเหมา (Contractor Tier)",
    tradeDiscountPct: "3.00%",
    aging: {
      current: "50,000.00",
      days1_30: "110,000.00",
      days31_60: "150,000.00",
      days61_90: "0.00",
      over90: "0.00",
    },
    transactions: [
      { id: "tx-601", docNo: "IV-202608-0145", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-08-20", amount: "150,000.00", status: "overdue", dueDate: "2026-09-04" },
      { id: "tx-602", docNo: "IV-202609-0099", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-09-18", amount: "110,000.00", status: "overdue", dueDate: "2026-10-03" },
      { id: "tx-603", docNo: "IV-202610-0033", docType: "ใบกำกับภาษี/ใบส่งของ", date: "2026-10-01", amount: "50,000.00", status: "pending", dueDate: "2026-10-16" },
    ],
  },
];

export interface SystemCustomerInquiryScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemCustomerInquiryScreen({
  language = "th",
  holdingcode = "rungrueng",
}: SystemCustomerInquiryScreenProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 34,
    min: 24,
    max: 50,
    storageKey: "bc_system_customer_inquiry_split",
  });

  const [customers, setCustomers] = useState<CustomerInquiryItem[]>(DEFAULT_CUSTOMERS);
  const [selectedId, setSelectedId] = useState<string>(DEFAULT_CUSTOMERS[0].id);
  const [search, setSearch] = useState("");
  const [filterGroup, setFilterGroup] = useState<string>("all");
  const [filterStatus, setFilterStatus] = useState<string>("all");
  const [activeTab, setActiveTab] = useState<"general" | "credit" | "transactions" | "terms">("general");

  const filteredCustomers = useMemo(() => {
    const q = search.trim().toLowerCase();
    return customers.filter((c) => {
      const matchSearch =
        !q ||
        c.code.toLowerCase().includes(q) ||
        c.nameTh.toLowerCase().includes(q) ||
        c.nameEn.toLowerCase().includes(q) ||
        c.taxId.includes(q) ||
        c.phone.includes(q) ||
        c.contactPerson.toLowerCase().includes(q);

      const matchGroup = filterGroup === "all" || c.debtorGroup === filterGroup;
      const matchStatus = filterStatus === "all" || c.status === filterStatus;

      return matchSearch && matchGroup && matchStatus;
    });
  }, [customers, search, filterGroup, filterStatus]);

  const selectedCustomer = useMemo(
    () => customers.find((c) => c.id === selectedId) ?? filteredCustomers[0] ?? customers[0],
    [customers, selectedId, filteredCustomers],
  );

  return (
    <div className="flex h-full flex-col overflow-hidden bg-background text-foreground" ref={containerRef}>
      {/* Header bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border bg-card px-4 py-2.5">
        <div className="flex items-center gap-2.5">
          <div className="grid h-9 w-9 place-items-center rounded-lg bg-primary/10 text-primary">
            <Users className="h-5 w-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold text-foreground">
                {language === "th" ? "แสดงข้อมูลลูกค้า" : "Customer Information Inquiry"}
              </h1>
              <span className="rounded-md bg-emerald-500/10 px-2 py-0.5 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-xs text-muted-foreground">
              {language === "th"
                ? "จัดการระบบ › แสดงข้อมูลลูกค้า: ตรวจสอบวงเงิน ยอดลูกหนี้ค้างชำระ เงื่อนไขราคา และประวัติเอกสาร"
                : "System Management > Customer Inquiry: View credit limits, aging balances, pricing, and history"}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setCustomers(DEFAULT_CUSTOMERS)}
            className="h-8 gap-1.5 text-xs"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span>{language === "th" ? "รีเฟรชข้อมูล" : "Refresh"}</span>
          </Button>
        </div>
      </div>

      {/* Main Content Pane */}
      <div className="relative flex flex-1 overflow-hidden">
        {/* Left Pane: Customer List */}
        <div
          className="flex flex-col border-r border-border bg-card/40"
          style={{ width: `${splitPercent}%` }}
        >
          {/* Search & Filters */}
          <div className="border-b border-border p-2.5 space-y-2">
            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
              <Input
                placeholder={language === "th" ? "ค้นหารหัส ชื่อ เลขผู้เสียภาษี เบอร์โทร..." : "Search code, name, tax ID..."}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="h-8 pl-8 text-xs"
              />
            </div>
            <div className="flex items-center gap-1.5">
              <select
                value={filterGroup}
                onChange={(e) => setFilterGroup(e.target.value)}
                className="h-7 flex-1 rounded-md border border-input bg-background px-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
              >
                <option value="all">{language === "th" ? "ทุกลูกค้า/กลุ่ม" : "All Groups"}</option>
                <option value="corporate">{language === "th" ? "นิติบุคคลทั่วไป" : "Corporate"}</option>
                <option value="wholesale">{language === "th" ? "ค้าส่ง/ตัวแทน" : "Wholesale"}</option>
                <option value="project">{language === "th" ? "งานโครงการ" : "Project"}</option>
                <option value="retail">{language === "th" ? "ผู้รับเหมา/ย่อย" : "Contractor/Retail"}</option>
                <option value="government">{language === "th" ? "สหกรณ์/ราชการ" : "Coop/Gov"}</option>
              </select>

              <select
                value={filterStatus}
                onChange={(e) => setFilterStatus(e.target.value)}
                className="h-7 w-28 rounded-md border border-input bg-background px-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
              >
                <option value="all">{language === "th" ? "ทุกสถานะ" : "All Status"}</option>
                <option value="active">{language === "th" ? "ปกติ" : "Active"}</option>
                <option value="credit_hold">{language === "th" ? "ระงับวงเงิน" : "Credit Hold"}</option>
              </select>
            </div>
          </div>

          {/* List count summary */}
          <div className="flex items-center justify-between border-b border-border/60 bg-muted/30 px-3 py-1.5 text-[0.72rem] text-muted-foreground">
            <span>
              {language === "th" ? `พบข้อมูล ${filteredCustomers.length} รายการ` : `Found ${filteredCustomers.length} customers`}
            </span>
            <span>{holdingcode}</span>
          </div>

          {/* Customer list items */}
          <div className="flex-1 overflow-y-auto divide-y divide-border/60">
            {filteredCustomers.length === 0 ? (
              <div className="p-8 text-center text-xs text-muted-foreground">
                {language === "th" ? "ไม่พบข้อมูลลูกค้าที่ตรงกับเงื่อนไข" : "No customers found"}
              </div>
            ) : (
              filteredCustomers.map((cust) => {
                const isSelected = cust.id === selectedCustomer?.id;
                return (
                  <button
                    key={cust.id}
                    type="button"
                    onClick={() => setSelectedId(cust.id)}
                    className={`w-full p-2.5 text-left transition-colors ${
                      isSelected
                        ? "bg-primary/10 border-l-4 border-primary"
                        : "hover:bg-muted/40"
                    }`}
                  >
                    <div className="flex items-start justify-between gap-1.5">
                      <span className="font-semibold text-xs text-primary">{cust.code}</span>
                      {cust.status === "credit_hold" ? (
                        <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.2 text-[0.68rem] font-medium bg-rose-500/10 text-rose-600 dark:text-rose-400">
                          ระงับวงเงิน
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.2 text-[0.68rem] font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                          ปกติ
                        </span>
                      )}
                    </div>
                    <p className="mt-0.5 line-clamp-1 text-xs font-medium text-foreground">
                      {cust.nameTh}
                    </p>
                    <div className="mt-1 flex items-center justify-between text-[0.7rem] text-muted-foreground">
                      <span>{cust.debtorGroupName}</span>
                      <span className="font-medium text-foreground">
                        หนี้: {cust.outstandingBalance} ฿
                      </span>
                    </div>
                  </button>
                );
              })
            )}
          </div>
        </div>

        {/* Resizable splitter handle */}
        <ResizableSplitter
          onMouseDown={startResize}
          onKeyDown={adjustWithKeyboard}
          isResizing={isResizing}
        />

        {/* Right Pane: Detailed Customer Profile */}
        <div
          className="flex-1 overflow-y-auto bg-background/60 p-4"
          style={{ width: `${100 - splitPercent}%` }}
        >
          {selectedCustomer ? (
            <div className="space-y-4">
              {/* Header Profile Card */}
              <Card className="border-border shadow-xs">
                <CardHeader className="p-4 pb-3">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="rounded-md bg-primary/15 px-2 py-0.5 text-xs font-bold text-primary">
                          {selectedCustomer.code}
                        </span>
                        <h2 className="text-base font-bold text-foreground">
                          {selectedCustomer.nameTh}
                        </h2>
                        {selectedCustomer.status === "credit_hold" ? (
                          <Badge variant="warning" className="gap-1 text-[0.7rem] bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30">
                            <ShieldAlert className="h-3 w-3" />
                            {language === "th" ? "ระงับวงเงินสินเชื่อ" : "Credit Hold"}
                          </Badge>
                        ) : (
                          <Badge variant="outline" className="gap-1 text-[0.7rem] border-emerald-500 text-emerald-600 dark:text-emerald-400">
                            <CheckCircle2 className="h-3 w-3" />
                            {language === "th" ? "สถานะเปิดขายปกติ" : "Active"}
                          </Badge>
                        )}
                      </div>
                      <p className="mt-1 text-xs text-muted-foreground">
                        {selectedCustomer.nameEn} · {selectedCustomer.debtorGroupName}
                      </p>
                    </div>

                    <div className="text-right">
                      <div className="text-[0.7rem] text-muted-foreground">วงเงินสินเชื่อรวม (Credit Limit)</div>
                      <div className="text-lg font-bold text-foreground">
                        {selectedCustomer.creditLimit} <span className="text-xs font-normal">บาท</span>
                      </div>
                    </div>
                  </div>
                </CardHeader>

                {/* KPI Ribbon */}
                <CardContent className="grid grid-cols-2 gap-2 border-t border-border/80 bg-muted/20 p-3 sm:grid-cols-4">
                  <div className="rounded-lg border border-border/60 bg-card p-2 text-center">
                    <span className="text-[0.68rem] text-muted-foreground">ยอดหนี้ค้างชำระ</span>
                    <p className="text-sm font-bold text-rose-600 dark:text-rose-400">
                      {selectedCustomer.outstandingBalance} ฿
                    </p>
                  </div>
                  <div className="rounded-lg border border-border/60 bg-card p-2 text-center">
                    <span className="text-[0.68rem] text-muted-foreground">วงเงินคงเหลือ</span>
                    <p className="text-sm font-bold text-emerald-600 dark:text-emerald-400">
                      {selectedCustomer.availableCredit} ฿
                    </p>
                  </div>
                  <div className="rounded-lg border border-border/60 bg-card p-2 text-center">
                    <span className="text-[0.68rem] text-muted-foreground">เครดิตเทอม</span>
                    <p className="text-sm font-bold text-foreground">
                      {selectedCustomer.paymentTermDays} วัน
                    </p>
                  </div>
                  <div className="rounded-lg border border-border/60 bg-card p-2 text-center">
                    <span className="text-[0.68rem] text-muted-foreground">ส่วนลดการค้า</span>
                    <p className="text-sm font-bold text-primary">
                      {selectedCustomer.tradeDiscountPct}
                    </p>
                  </div>
                </CardContent>
              </Card>

              {/* Sub-tabs Navigation */}
              <div className="flex border-b border-border gap-2">
                <button
                  type="button"
                  onClick={() => setActiveTab("general")}
                  className={`px-3 py-2 text-xs font-semibold border-b-2 transition-colors ${
                    activeTab === "general"
                      ? "border-primary text-primary"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {language === "th" ? "ข้อมูลทั่วไปและที่อยู่" : "General Info"}
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("credit")}
                  className={`px-3 py-2 text-xs font-semibold border-b-2 transition-colors ${
                    activeTab === "credit"
                      ? "border-primary text-primary"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {language === "th" ? "การวิเคราะห์อายุหนี้ (Aging)" : "Credit Aging"}
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("transactions")}
                  className={`px-3 py-2 text-xs font-semibold border-b-2 transition-colors ${
                    activeTab === "transactions"
                      ? "border-primary text-primary"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {language === "th" ? "ประวัติเอกสารล่าสุด" : "Recent Documents"}
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("terms")}
                  className={`px-3 py-2 text-xs font-semibold border-b-2 transition-colors ${
                    activeTab === "terms"
                      ? "border-primary text-primary"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {language === "th" ? "เงื่อนไขราคาขาย" : "Pricing Terms"}
                </button>
              </div>

              {/* Tab 1: General Info */}
              {activeTab === "general" && (
                <Card className="border-border">
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-sm font-bold">ข้อมูลนิติบุคคลและผู้ติดต่อ</CardTitle>
                    <CardDescription className="text-xs">
                      ข้อมูลสำหรับออกใบกำกับภาษี ใบส่งของ และการประสานงาน
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="p-4 pt-1 space-y-3">
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                      <div className="space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">เลขประจำตัวผู้เสียภาษีอากร (Tax ID)</span>
                        <p className="text-xs font-mono font-medium text-foreground">{selectedCustomer.taxId}</p>
                      </div>
                      <div className="space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">รหัสสาขา/ชื่อสาขา</span>
                        <p className="text-xs font-medium text-foreground">
                          {selectedCustomer.branchCode} ({selectedCustomer.branchName})
                        </p>
                      </div>
                      <div className="space-y-1 sm:col-span-2">
                        <span className="text-[0.72rem] text-muted-foreground">ที่อยู่ตามหนังสือรับรอง (สำนักงานใหญ่)</span>
                        <p className="text-xs leading-relaxed text-foreground">{selectedCustomer.address}</p>
                      </div>
                      <div className="space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">บุคคลติดต่อ (Contact Person)</span>
                        <p className="text-xs font-medium text-foreground">{selectedCustomer.contactPerson}</p>
                      </div>
                      <div className="space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">เบอร์โทรศัพท์</span>
                        <p className="text-xs font-medium text-foreground">{selectedCustomer.phone}</p>
                      </div>
                      <div className="space-y-1 sm:col-span-2">
                        <span className="text-[0.72rem] text-muted-foreground">อีเมลติดต่อเอกสารบัญชี</span>
                        <p className="text-xs font-medium text-foreground">{selectedCustomer.email}</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Tab 2: Credit Aging */}
              {activeTab === "credit" && (
                <Card className="border-border">
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-sm font-bold">ตารางแจกแจงอายุหนี้ (Accounts Receivable Aging)</CardTitle>
                    <CardDescription className="text-xs">
                      ยอดหนี้แบ่งตามช่วงระยะเวลาครบกำหนดชำระเพื่อการบริหารความเสี่ยง
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="p-4 pt-1">
                    <div className="overflow-x-auto rounded-lg border border-border">
                      <table className="w-full text-left text-xs">
                        <thead className="bg-muted/50 text-[0.7rem] uppercase text-muted-foreground font-semibold">
                          <tr>
                            <th className="px-3 py-2">ยังไม่ถึงกำหนด (Current)</th>
                            <th className="px-3 py-2">เกิน 1-30 วัน</th>
                            <th className="px-3 py-2">เกิน 31-60 วัน</th>
                            <th className="px-3 py-2">เกิน 61-90 วัน</th>
                            <th className="px-3 py-2">เกิน 90 วัน</th>
                            <th className="px-3 py-2 font-bold text-foreground">รวมยอดค้างชำระ</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                          <tr>
                            <td className="px-3 py-2.5 font-medium">{selectedCustomer.aging.current} ฿</td>
                            <td className="px-3 py-2.5 font-medium text-amber-600 dark:text-amber-400">
                              {selectedCustomer.aging.days1_30} ฿
                            </td>
                            <td className="px-3 py-2.5 font-medium text-orange-600 dark:text-orange-400">
                              {selectedCustomer.aging.days31_60} ฿
                            </td>
                            <td className="px-3 py-2.5 font-medium text-rose-600 dark:text-rose-400">
                              {selectedCustomer.aging.days61_90} ฿
                            </td>
                            <td className="px-3 py-2.5 font-medium text-rose-700 dark:text-rose-300">
                              {selectedCustomer.aging.over90} ฿
                            </td>
                            <td className="px-3 py-2.5 font-bold text-foreground bg-muted/20">
                              {selectedCustomer.outstandingBalance} ฿
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Tab 3: Recent Transactions */}
              {activeTab === "transactions" && (
                <Card className="border-border">
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-sm font-bold">เอกสารและประวัติรายการค้าล่าสุด</CardTitle>
                    <CardDescription className="text-xs">
                      รายการใบส่งของ/ใบกำกับภาษี และใบเสร็จรับเงินที่เกี่ยวข้อง
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="p-4 pt-1">
                    <div className="overflow-x-auto rounded-lg border border-border">
                      <table className="w-full text-left text-xs">
                        <thead className="bg-muted/50 text-[0.7rem] uppercase text-muted-foreground font-semibold">
                          <tr>
                            <th className="px-3 py-2">เลขที่เอกสาร</th>
                            <th className="px-3 py-2">ประเภทเอกสาร</th>
                            <th className="px-3 py-2">วันที่เอกสาร</th>
                            <th className="px-3 py-2">วันครบกำหนด</th>
                            <th className="px-3 py-2 text-right">จำนวนเงิน (บาท)</th>
                            <th className="px-3 py-2 text-center">สถานะ</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                          {selectedCustomer.transactions.map((tx) => (
                            <tr key={tx.id} className="hover:bg-muted/30">
                              <td className="px-3 py-2 font-mono font-medium text-primary">{tx.docNo}</td>
                              <td className="px-3 py-2 text-muted-foreground">{tx.docType}</td>
                              <td className="px-3 py-2">{tx.date}</td>
                              <td className="px-3 py-2">{tx.dueDate}</td>
                              <td className="px-3 py-2 text-right font-medium">{tx.amount}</td>
                              <td className="px-3 py-2 text-center">
                                {tx.status === "paid" && (
                                  <span className="rounded bg-emerald-500/10 px-1.5 py-0.5 text-[0.68rem] text-emerald-600 dark:text-emerald-400">
                                    ชำระแล้ว
                                  </span>
                                )}
                                {tx.status === "pending" && (
                                  <span className="rounded bg-amber-500/10 px-1.5 py-0.5 text-[0.68rem] text-amber-600 dark:text-amber-400">
                                    รอชำระ
                                  </span>
                                )}
                                {tx.status === "overdue" && (
                                  <span className="rounded bg-rose-500/10 px-1.5 py-0.5 text-[0.68rem] text-rose-600 dark:text-rose-400">
                                    เกินกำหนด
                                  </span>
                                )}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Tab 4: Terms & Pricing */}
              {activeTab === "terms" && (
                <Card className="border-border">
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-sm font-bold">เงื่อนไขการค้าและระดับราคา</CardTitle>
                    <CardDescription className="text-xs">
                      การตั้งค่าระดับราคาขายและนโยบายการให้เครดิตสำหรับลูกค้ารายนี้
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="p-4 pt-1 space-y-3">
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                      <div className="rounded-lg border border-border p-3 space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">ระดับราคาขายที่กำหนด (Price Tier)</span>
                        <p className="text-xs font-semibold text-foreground">{selectedCustomer.priceTier}</p>
                      </div>
                      <div className="rounded-lg border border-border p-3 space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">ส่วนลดการค้าอัตโนมัติ (Trade Discount)</span>
                        <p className="text-xs font-semibold text-primary">{selectedCustomer.tradeDiscountPct}</p>
                      </div>
                      <div className="rounded-lg border border-border p-3 space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">ระยะเวลาให้สินเชื่อ (Payment Term)</span>
                        <p className="text-xs font-semibold text-foreground">{selectedCustomer.paymentTermDays} วัน</p>
                      </div>
                      <div className="rounded-lg border border-border p-3 space-y-1">
                        <span className="text-[0.72rem] text-muted-foreground">การควบคุมวงเงิน (Credit Control)</span>
                        <p className="text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                          {selectedCustomer.status === "credit_hold" ? "ระงับการเปิดบิลใหม่" : "เปิดบิลได้ตามวงเงินคงเหลือ"}
                        </p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}
            </div>
          ) : (
            <div className="grid h-full place-items-center text-xs text-muted-foreground">
              กรุณาเลือกลูกค้าจากรายการทางด้านซ้าย
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
