"use client";

import { useState, useMemo } from "react";
import { useBackendText } from "@/components/backend-text-provider";
import { type LanguageCode } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Building2,
  UploadCloud,
  Sparkles,
  CheckCircle2,
  AlertCircle,
  Clock,
  ArrowDownLeft,
  ArrowUpRight,
  FileSpreadsheet,
  Check,
  RefreshCw,
  PlusCircle,
  FileText,
  Printer,
  ChevronRight,
  CreditCard,
  Wallet,
} from "lucide-react";
import {
  parseThaiBankStatement,
  autoReconcileBankStatement,
  calculateBankReconciliation,
  generateBankFastVoucher,
  type BankStatementTransaction,
  type BookTransaction,
  type BankReconciliationSummary,
} from "@/lib/bank-reconciliation";
import { ChequeManagementView } from "./cheque-management-view";
import { PettyCashView } from "./petty-cash-view";

interface BankingWorkbenchProps {
  route?: string;
  embedded?: boolean;
  language?: LanguageCode;
  holdingcode?: string;
  businesscode?: string;
}

export function BankingWorkbench({
  route: _route = "/banking/reconciliation",
  embedded: _embedded = false,
  language: _language = "th",
}: BankingWorkbenchProps) {
  const tr = useBackendText();

  // Module Navigation Tab
  const defaultTab = _route.includes("cheque")
    ? "cheque"
    : _route.includes("petty")
      ? "petty"
      : "recon";
  const [activeTab, setActiveTab] = useState<"recon" | "cheque" | "petty">(defaultTab);

  // Selected Bank Account
  const [selectedBank, setSelectedBank] = useState<string>("111201");
  const [statementEndBal, setStatementEndBal] = useState<number>(154820.5);
  const [bookEndBal, setBookEndBal] = useState<number>(152680.5);

  // Initial Sample Statements
  const [statements, setStatements] = useState<BankStatementTransaction[]>([
    {
      id: "stmt-1",
      date: "2026-09-10",
      type: "deposit",
      amount: 25000,
      description: "PromptPay Transfer from Customer A",
      reference: "INV2026001",
      channel: "promptpay",
      status: "unmatched",
    },
    {
      id: "stmt-2",
      date: "2026-09-12",
      type: "withdrawal",
      amount: 4500,
      description: "Direct Debit - Internet TrueOnline",
      channel: "transfer",
      status: "unmatched",
    },
    {
      id: "stmt-3",
      date: "2026-09-14",
      type: "withdrawal",
      amount: 140,
      description: "Bank Monthly Transfer Fee",
      channel: "fee",
      status: "unmatched",
    },
    {
      id: "stmt-4",
      date: "2026-09-15",
      type: "deposit",
      amount: 8500,
      description: "Transfer in from KBank EDC",
      reference: "EDC9942",
      channel: "edc",
      status: "unmatched",
    },
  ]);

  // Initial Sample Book Transactions
  const [books, setBooks] = useState<BookTransaction[]>([
    {
      docno: "RV2026-09001",
      date: "2026-09-10",
      type: "receipt",
      amount: 25000,
      accountCode: "111201",
      accountName: "ธนาคารกสิกรไทย ออมทรัพย์",
      description: "รับชำระหนี้จากลูกหนี้การค้า INV2026001",
      reference: "INV2026001",
      isReconciled: false,
    },
    {
      docno: "PV2026-09042",
      date: "2026-09-11",
      type: "payment",
      amount: 4500,
      accountCode: "111201",
      accountName: "ธนาคารกสิกรไทย ออมทรัพย์",
      description: "จ่ายค่าบริการอินเทอร์เน็ตประจำเดือน",
      isReconciled: false,
    },
    {
      docno: "RV2026-09099",
      date: "2026-09-18",
      type: "receipt",
      amount: 12000,
      accountCode: "111201",
      accountName: "ธนาคารกสิกรไทย ออมทรัพย์",
      description: "เงินโอนรับชำระระหว่างทาง (ส่งฝากสิ้นวัน)",
      isReconciled: false,
    },
  ]);

  // Modals & UI Controls
  const [showImportModal, setShowImportModal] = useState(false);
  const [showStatementModal, setShowStatementModal] = useState(false);
  const [importText, setImportText] = useState("");
  const [fastVoucherMessage, setFastVoucherMessage] = useState<string | null>(null);

  // Recalculate Bank Reconciliation Summary
  const reconSummary: BankReconciliationSummary = useMemo(() => {
    return calculateBankReconciliation({
      statementEndingBalance: statementEndBal,
      statementTransactions: statements,
      bookEndingBalance: bookEndBal,
      bookTransactions: books,
    });
  }, [statementEndBal, statements, bookEndBal, books]);

  // Run AI Auto Reconciliation
  const handleAutoReconcile = () => {
    const res = autoReconcileBankStatement(statements, books, { dateWindowDays: 3 });
    setStatements(res.matchedStatements);
    setBooks(res.matchedBooks);
  };

  // Import Statement Handler
  const handleImportSubmit = () => {
    if (!importText.trim()) return;
    const parsed = parseThaiBankStatement(importText);
    if (parsed.length > 0) {
      setStatements((prev) => [...prev, ...parsed]);
      setImportText("");
      setShowImportModal(false);
    }
  };

  // Fast Voucher Generator
  const handleFastVoucher = (stmt: BankStatementTransaction) => {
    const voucher = generateBankFastVoucher(stmt, selectedBank);
    setFastVoucherMessage(
      `สร้างใบสำคัญ ${voucher.journalType} สำเร็จ: ${voucher.description} ยอด ${stmt.amount.toLocaleString()} บาท`,
    );
    // Mark as cleared
    setStatements((prev) =>
      prev.map((s) =>
        s.id === stmt.id ? { ...s, status: "cleared", matchedJournalDocNo: `${voucher.journalType}-NEW` } : s,
      ),
    );
    setTimeout(() => setFastVoucherMessage(null), 5000);
  };

  return (
    <div className="flex flex-col gap-5 p-2">
      {/* Module Navigation Tabs */}
      <div className="flex flex-wrap items-center gap-2 border-b border-border/70 pb-3">
        <button
          type="button"
          onClick={() => setActiveTab("recon")}
          className={`flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-xl transition-all shadow-xs ${
            activeTab === "recon"
              ? "bg-primary text-primary-foreground shadow-md shadow-primary/20"
              : "bg-background hover:bg-muted/70 text-muted-foreground border border-border/60"
          }`}
        >
          <Building2 className="size-4" />
          <span>กระทบยอดเงินฝากธนาคาร (Bank Feeds & Reconciliation)</span>
        </button>

        <button
          type="button"
          onClick={() => setActiveTab("cheque")}
          className={`flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-xl transition-all shadow-xs ${
            activeTab === "cheque"
              ? "bg-primary text-primary-foreground shadow-md shadow-primary/20"
              : "bg-background hover:bg-muted/70 text-muted-foreground border border-border/60"
          }`}
        >
          <CreditCard className="size-4" />
          <span>ทะเบียนเช็ครับ-จ่าย (Cheque Flow)</span>
        </button>

        <button
          type="button"
          onClick={() => setActiveTab("petty")}
          className={`flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-xl transition-all shadow-xs ${
            activeTab === "petty"
              ? "bg-primary text-primary-foreground shadow-md shadow-primary/20"
              : "bg-background hover:bg-muted/70 text-muted-foreground border border-border/60"
          }`}
        >
          <Wallet className="size-4" />
          <span>เงินสดย่อย (Petty Cash Imprest)</span>
        </button>
      </div>

      {activeTab === "recon" && (
        <>
          {/* Top Header Card */}
          <Card className="border-border/60 shadow-sm">
        <CardContent className="flex flex-col gap-4 p-5 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex items-center gap-3">
            <span className="grid h-12 w-12 place-items-center rounded-xl bg-primary/10 text-primary">
              <Building2 className="h-6 w-6" />
            </span>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-xl font-bold tracking-tight text-foreground">
                  {tr("bank_reconciliation_title", "ระบบกระทบยอดเงินฝากธนาคารอัจฉริยะ (Bank Feeds & Smart Reconciliation)")}
                </h1>
                <span className="rounded-md bg-emerald-500/10 px-2 py-0.5 text-xs font-semibold text-emerald-600">
                  TFRS & Tax Compliant
                </span>
              </div>
              <p className="text-xs text-muted-foreground">
                {tr(
                  "bank_reconciliation_subtitle",
                  "เชื่อมต่อ Statement ธนาคารไทย (KBank, SCB, BBL, KTB) กระทบยอด 2 ทางอัตโนมัติ และออกงบพิสูจน์ยอด",
                )}
              </p>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowImportModal(true)}
              className="gap-1.5"
            >
              <UploadCloud className="h-4 w-4 text-primary" />
              {tr("bank_import_statement", "นำเข้า Statement")}
            </Button>

            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowStatementModal(true)}
              className="gap-1.5"
            >
              <FileSpreadsheet className="h-4 w-4 text-blue-600" />
              {tr("bank_reconciliation_statement", "งบพิสูจน์ยอดเงินฝาก")}
            </Button>

            <Button
              size="sm"
              onClick={handleAutoReconcile}
              className="gap-1.5 bg-primary text-primary-foreground shadow-sm"
            >
              <Sparkles className="h-4 w-4" />
              {tr("bank_auto_match_btn", "AI กระทบยอดอัตโนมัติ")}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Alert Banner if Fast Voucher was created */}
      {fastVoucherMessage && (
        <div className="flex items-center justify-between rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-3 text-sm text-emerald-800">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="h-4 w-4 text-emerald-600" />
            <span>{fastVoucherMessage}</span>
          </div>
        </div>
      )}

      {/* Metric Cards Row */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card className="border-border/60">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-muted-foreground">
                {tr("bank_statement_ending_balance", "ยอดตามใบแจ้งยอดธนาคาร")}
              </span>
              <Building2 className="h-4 w-4 text-blue-600" />
            </div>
            <div className="mt-2 text-xl font-bold">
              ฿{statementEndBal.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </div>
            <p className="mt-1 text-[11px] text-muted-foreground">
              {tr("bank_acct_kbank", "111201 ธนาคารกสิกรไทย ออมทรัพย์")}
            </p>
          </CardContent>
        </Card>

        <Card className="border-border/60">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-muted-foreground">
                {tr("bank_book_ending_balance", "ยอดตามสมุดบัญชีกิจการ")}
              </span>
              <FileText className="h-4 w-4 text-indigo-600" />
            </div>
            <div className="mt-2 text-xl font-bold">
              ฿{bookEndBal.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </div>
            <p className="mt-1 text-[11px] text-muted-foreground">
              {tr("bank_general_ledger_balance", "ยอดคงเหลือในผังบัญชีแยกประเภท")}
            </p>
          </CardContent>
        </Card>

        <Card className="border-border/60">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-muted-foreground">
                {tr("bank_matched_count", "กระทบยอดสำเร็จแล้ว")}
              </span>
              <CheckCircle2 className="h-4 w-4 text-emerald-600" />
            </div>
            <div className="mt-2 text-xl font-bold text-emerald-600">
              {statements.filter((s) => s.status === "matched" || s.status === "cleared").length} / {statements.length}
            </div>
            <p className="mt-1 text-[11px] text-muted-foreground">
              {tr("bank_matched_status_hint", "จับคู่ตรงกันทั้งยอดเงินและช่วงวันที่")}
            </p>
          </CardContent>
        </Card>

        <Card className="border-border/60">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-muted-foreground">
                {tr("bank_adjusted_difference", "ผลต่างหลังพิสูจน์ยอด")}
              </span>
              {reconSummary.isBalanced ? (
                <Check className="h-4 w-4 text-emerald-600" />
              ) : (
                <AlertCircle className="h-4 w-4 text-amber-600" />
              )}
            </div>
            <div
              className={`mt-2 text-xl font-bold ${
                reconSummary.isBalanced ? "text-emerald-600" : "text-amber-600"
              }`}
            >
              ฿{reconSummary.difference.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </div>
            <p className="mt-1 text-[11px] text-muted-foreground">
              {reconSummary.isBalanced
                ? tr("bank_fully_reconciled", "✓ ยอดสมดุล 100% ไม่มีผลต่างค้าง")
                : tr("bank_has_differences", "มีรายการค้างกระทบยอด")}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 2-Way Reconciliation Split Canvas */}
      <div className="grid grid-cols-1 gap-5 lg:grid-cols-2">
        {/* Left Panel: Bank Statement Feeds */}
        <Card className="border-border/60 shadow-sm">
          <CardContent className="p-4">
            <div className="flex items-center justify-between border-b pb-3">
              <div className="flex items-center gap-2">
                <Building2 className="h-5 w-5 text-blue-600" />
                <h2 className="font-semibold text-foreground">
                  {tr("bank_statement_panel_title", "1. รายการเดินบัญชีธนาคาร (Bank Statement)")}
                </h2>
              </div>
              <span className="rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground">
                {statements.length} รายการ
              </span>
            </div>

            <div className="mt-3 flex flex-col gap-2.5">
              {statements.map((stmt) => (
                <div
                  key={stmt.id}
                  className={`flex flex-col justify-between gap-2 rounded-lg border p-3 transition-colors ${
                    stmt.status === "matched" || stmt.status === "cleared"
                      ? "border-emerald-500/20 bg-emerald-50/40 dark:bg-emerald-950/10"
                      : "border-border/70 bg-card hover:border-primary/40"
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-semibold text-muted-foreground">{stmt.date}</span>
                        {stmt.type === "deposit" ? (
                          <span className="inline-flex items-center gap-1 rounded bg-emerald-500/15 px-1.5 py-0.5 text-[10px] font-semibold text-emerald-700">
                            <ArrowDownLeft className="h-3 w-3" /> เงินเข้า (ฝาก)
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 rounded bg-rose-500/15 px-1.5 py-0.5 text-[10px] font-semibold text-rose-700">
                            <ArrowUpRight className="h-3 w-3" /> เงินออก (ถอน)
                          </span>
                        )}
                        {stmt.channel && (
                          <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] uppercase text-muted-foreground">
                            {stmt.channel}
                          </span>
                        )}
                      </div>
                      <p className="mt-1 text-sm font-medium text-foreground">{stmt.description}</p>
                      {stmt.reference && (
                        <p className="text-xs text-muted-foreground">Ref: {stmt.reference}</p>
                      )}
                    </div>

                    <div className="text-right">
                      <span
                        className={`text-base font-bold ${
                          stmt.type === "deposit" ? "text-emerald-600" : "text-foreground"
                        }`}
                      >
                        {stmt.type === "deposit" ? "+" : "-"}฿
                        {stmt.amount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center justify-between border-t pt-2 text-xs">
                    {stmt.status === "matched" || stmt.status === "cleared" ? (
                      <span className="flex items-center gap-1 font-semibold text-emerald-600">
                        <CheckCircle2 className="h-3.5 w-3.5" />
                        {tr("bank_matched_with", "จับคู่สำเร็จ:")} {stmt.matchedJournalDocNo}
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-amber-600">
                        <Clock className="h-3.5 w-3.5" />
                        {tr("bank_status_unmatched", "ยังไม่ได้กระทบยอด")}
                      </span>
                    )}

                    {stmt.status === "unmatched" && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleFastVoucher(stmt)}
                        className="h-7 gap-1 px-2 text-xs text-primary hover:bg-primary/10"
                      >
                        <PlusCircle className="h-3.5 w-3.5" />
                        {tr("bank_fast_voucher_btn", "สร้างใบสำคัญ 1 คลิก")}
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Right Panel: Book Transactions (Vouchers) */}
        <Card className="border-border/60 shadow-sm">
          <CardContent className="p-4">
            <div className="flex items-center justify-between border-b pb-3">
              <div className="flex items-center gap-2">
                <FileText className="h-5 w-5 text-indigo-600" />
                <h2 className="font-semibold text-foreground">
                  {tr("bank_book_panel_title", "2. สมุดบัญชีกิจการ (Book Vouchers / GL)")}
                </h2>
              </div>
              <span className="rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground">
                {books.length} รายการ
              </span>
            </div>

            <div className="mt-3 flex flex-col gap-2.5">
              {books.map((book) => (
                <div
                  key={book.docno}
                  className={`flex flex-col justify-between gap-2 rounded-lg border p-3 transition-colors ${
                    book.isReconciled
                      ? "border-emerald-500/20 bg-emerald-50/40 dark:bg-emerald-950/10"
                      : "border-border/70 bg-card hover:border-indigo-400"
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-xs font-bold text-primary">{book.docno}</span>
                        <span className="text-xs text-muted-foreground">{book.date}</span>
                        <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] uppercase font-semibold text-muted-foreground">
                          {book.type}
                        </span>
                      </div>
                      <p className="mt-1 text-sm font-medium text-foreground">{book.description}</p>
                      <p className="text-xs text-muted-foreground">{book.accountName}</p>
                    </div>

                    <div className="text-right">
                      <span className="text-base font-bold text-foreground">
                        ฿{book.amount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center justify-between border-t pt-2 text-xs">
                    {book.isReconciled ? (
                      <span className="flex items-center gap-1 font-semibold text-emerald-600">
                        <CheckCircle2 className="h-3.5 w-3.5" />
                        {tr("bank_reconciled_tag", "กระทบยอดแล้ว")}
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-muted-foreground">
                        <Clock className="h-3.5 w-3.5 text-amber-500" />
                        {tr("bank_pending_statement_tag", "รอ Statement ธนาคาร")}
                      </span>
                    )}

                    {book.reference && (
                      <span className="text-[11px] text-muted-foreground">Ref: {book.reference}</span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Modal 1: Import Statement */}
      {showImportModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <Card className="w-full max-w-2xl border-border/80 bg-background shadow-xl">
            <CardContent className="p-6">
              <div className="flex items-center justify-between border-b pb-3">
                <div className="flex items-center gap-2">
                  <UploadCloud className="h-5 w-5 text-primary" />
                  <h3 className="text-lg font-bold text-foreground">
                    {tr("bank_import_modal_title", "นำเข้าข้อมูล Statement ธนาคาร (KBank, SCB, BBL)")}
                  </h3>
                </div>
                <Button variant="ghost" size="sm" onClick={() => setShowImportModal(false)}>
                  ✕
                </Button>
              </div>

              <div className="mt-4 flex flex-col gap-3">
                <p className="text-xs text-muted-foreground">
                  {tr(
                    "bank_import_instructions",
                    "คัดลอก (Copy) ตารางจากไฟล์ Excel หรือ Statement แล้ววาง (Paste) ลงในช่องด้านล่าง ระบบจะจัดรูปแบบให้อัตโนมัติ",
                  )}
                </p>
                <textarea
                  className="h-44 w-full rounded-md border border-input bg-background p-3 font-mono text-xs focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                  placeholder="Date,Description,Withdrawal,Deposit&#10;2026-09-15,PromptPay payment to Supplier,1200.00,0.00&#10;2026-09-16,Transfer in from customer,0.00,8500.00"
                  value={importText}
                  onChange={(e) => setImportText(e.target.value)}
                />
              </div>

              <div className="mt-6 flex justify-end gap-2">
                <Button variant="outline" size="sm" onClick={() => setShowImportModal(false)}>
                  {tr("cancel", "ยกเลิก")}
                </Button>
                <Button size="sm" onClick={handleImportSubmit} className="gap-1.5">
                  <Check className="h-4 w-4" />
                  {tr("confirm_import", "ยืนยันนำเข้าข้อมูล")}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Modal 2: Official Bank Reconciliation Statement */}
      {showStatementModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <Card className="w-full max-w-3xl border-border/80 bg-background shadow-2xl">
            <CardContent className="p-6">
              <div className="flex items-center justify-between border-b pb-4">
                <div>
                  <h3 className="text-lg font-bold text-foreground">
                    {tr("bank_recon_statement_modal_title", "งบพิสูจน์ยอดเงินฝากธนาคาร (Bank Reconciliation Statement)")}
                  </h3>
                  <p className="text-xs text-muted-foreground">
                    ณ วันที่ 30 กันยายน 2569 • บัญชีธนาคารกสิกรไทย (111201)
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <Button variant="outline" size="sm" onClick={() => window.print()} className="gap-1">
                    <Printer className="h-4 w-4" /> พิมพ์
                  </Button>
                  <Button variant="ghost" size="sm" onClick={() => setShowStatementModal(false)}>
                    ✕
                  </Button>
                </div>
              </div>

              {/* TFRS Reconciliation Table */}
              <div className="mt-5 space-y-4 font-sans text-sm">
                <div className="rounded-lg border bg-muted/20 p-4 space-y-2.5">
                  <div className="flex justify-between font-semibold">
                    <span>ยอดคงเหลือตามใบแจ้งยอดธนาคาร (Statement Balance)</span>
                    <span>฿{reconSummary.statementEndingBalance.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                  </div>

                  <div className="flex justify-between pl-4 text-xs text-emerald-700">
                    <span>บวก: เงินฝากระหว่างทาง (Deposits in Transit) ({reconSummary.depositsInTransitCount} รายการ)</span>
                    <span>+฿{reconSummary.depositsInTransit.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                  </div>

                  <div className="flex justify-between pl-4 text-xs text-rose-700">
                    <span>หัก: เช็คสั่งจ่ายที่ยังไม่ได้นำไปขึ้นเงิน (Outstanding Cheques) ({reconSummary.outstandingChequesCount} รายการ)</span>
                    <span>-฿{reconSummary.outstandingCheques.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                  </div>

                  <div className="border-t pt-2 flex justify-between font-bold text-primary">
                    <span>= ยอดคงเหลือที่ถูกต้องตามธนาคาร (Adjusted Bank Balance)</span>
                    <span>฿{reconSummary.adjustedBankBalance.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                  </div>
                </div>

                <div className="rounded-lg border bg-muted/20 p-4 space-y-2.5">
                  <div className="flex justify-between font-semibold">
                    <span>ยอดคงเหลือตามสมุดบัญชีกิจการ (Book Balance)</span>
                    <span>฿{reconSummary.bookEndingBalance.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                  </div>

                  <div className="flex justify-between pl-4 text-xs text-emerald-700">
                    <span>บวก: ดอกเบี้ยรับ / เงินโอนรับที่ยังไม่ได้ลงบัญชี</span>
                    <span>+฿{(reconSummary.unrecordedBankInterest + reconSummary.unrecordedTransfersIn).toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                  </div>

                  <div className="flex justify-between pl-4 text-xs text-rose-700">
                    <span>หัก: ค่าธรรมเนียมธนาคารที่ยังไม่ได้ลงบัญชี</span>
                    <span>-฿{reconSummary.unrecordedBankCharges.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                  </div>

                  <div className="border-t pt-2 flex justify-between font-bold text-primary">
                    <span>= ยอดคงเหลือที่ถูกต้องตามสมุดบัญชี (Adjusted Book Balance)</span>
                    <span>฿{reconSummary.adjustedBookBalance.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                  </div>
                </div>

                {/* Audit Status */}
                <div
                  className={`flex items-center justify-between rounded-lg p-3 font-semibold ${
                    reconSummary.isBalanced
                      ? "bg-emerald-500/15 text-emerald-800"
                      : "bg-amber-500/15 text-amber-800"
                  }`}
                >
                  <span className="flex items-center gap-2">
                    {reconSummary.isBalanced ? <CheckCircle2 className="h-5 w-5" /> : <AlertCircle className="h-5 w-5" />}
                    {reconSummary.isBalanced
                      ? "ยอดปรับปรุงทั้งสองฝั่งลงตัว 100% (Balanced)"
                      : "มียอดผลต่างที่ต้องตรวจสอบเพิ่มเติม"}
                  </span>
                  <span>ผลต่าง: ฿{reconSummary.difference.toFixed(2)}</span>
                </div>
              </div>

              <div className="mt-6 flex justify-end">
                <Button size="sm" onClick={() => setShowStatementModal(false)}>
                  {tr("close", "ปิด")}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
        </>
      )}

      {activeTab === "cheque" && <ChequeManagementView />}

      {activeTab === "petty" && <PettyCashView />}
    </div>
  );
}
