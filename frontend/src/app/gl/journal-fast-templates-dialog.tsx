"use client";

import { useState, useMemo } from "react";
import {
  Sparkles,
  Search,
  CheckCircle2,
  X,
  FileSpreadsheet,
  ArrowRight,
  Info,
  Scale,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  THAI_JOURNAL_PATTERNS,
  THAI_BUSINESS_METADATA,
  type ThaiBusinessType,
  type ThaiJournalPatternTemplate,
  calculatePatternJournalLines,
} from "@/lib/thai-accounting-business-patterns";
import { formatAmount } from "@/lib/general-ledger";
import { useGLText } from "./gl-common";

export interface FastTemplateApplyData {
  lines: Array<{
    accountcode: string;
    description: string;
    debit: string;
    credit: string;
    departmentcode: string;
    projectcode: string;
    cashflow: string;
  }>;
  description: string;
}

export function JournalFastTemplatesDialog({
  open,
  onClose,
  onApply,
}: {
  open: boolean;
  onClose: () => void;
  onApply: (data: FastTemplateApplyData) => void;
}) {
  const tr = useGLText();
  const [selectedType, setSelectedType] = useState<ThaiBusinessType | "all">("all");
  const [search, setSearch] = useState("");
  const [baseAmount, setBaseAmount] = useState("10000");
  const [selectedPatternId, setSelectedPatternId] = useState<string>("service_office_rent");

  const businessTypesList: Array<{ id: ThaiBusinessType | "all"; label: string }> = [
    { id: "all", label: tr("gl_all_types", "ทั้งหมด") },
    { id: "trading", label: tr("gl_biz_trading", "ซื้อมาขายไป") },
    { id: "service", label: tr("gl_biz_service", "บริการ / วิชาชีพ") },
    { id: "manufacturing", label: tr("gl_biz_mfg", "โรงงาน / ผลิต") },
    { id: "restaurant_cafe", label: tr("gl_biz_fnb", "ร้านอาหาร / คาเฟ่") },
    { id: "construction", label: tr("gl_biz_construct", "รับเหมาก่อสร้าง") },
    { id: "ecommerce", label: tr("gl_biz_ecom", "ออนไลน์ / Social") },
    { id: "real_estate_rental", label: tr("gl_biz_rental", "อสังหาฯ / ค่าเช่า") },
    { id: "transport_logistics", label: tr("gl_biz_transport", "ขนส่ง / โลจิสติกส์") },
  ];

  const filteredPatterns = useMemo(() => {
    const q = search.toLowerCase().trim();
    return THAI_JOURNAL_PATTERNS.filter((pattern) => {
      const matchType = selectedType === "all" || pattern.businessType === selectedType;
      if (!matchType) return false;
      if (!q) return true;
      return (
        pattern.titleTh.toLowerCase().includes(q) ||
        pattern.titleEn.toLowerCase().includes(q) ||
        pattern.keywords.some((k) => k.toLowerCase().includes(q)) ||
        pattern.lines.some((l) => l.accountNameTh.toLowerCase().includes(q) || l.accountCode.includes(q))
      );
    });
  }, [selectedType, search]);

  const activePattern = useMemo(() => {
    return (
      filteredPatterns.find((p) => p.id === selectedPatternId) ||
      filteredPatterns[0] ||
      THAI_JOURNAL_PATTERNS[0]
    );
  }, [filteredPatterns, selectedPatternId]);

  const parsedBaseAmount = useMemo(() => {
    const num = parseFloat(baseAmount.replace(/,/g, ""));
    return isNaN(num) || num < 0 ? 0 : num;
  }, [baseAmount]);

  const calculatedLines = useMemo(() => {
    if (!activePattern) return [];
    return calculatePatternJournalLines(activePattern, parsedBaseAmount);
  }, [activePattern, parsedBaseAmount]);

  const totals = useMemo(() => {
    let debitSum = 0;
    let creditSum = 0;
    for (const line of calculatedLines) {
      if (line.side === "debit") debitSum += Math.round(line.amount * 100);
      else creditSum += Math.round(line.amount * 100);
    }
    return {
      debit: debitSum / 100,
      credit: creditSum / 100,
      balanced: debitSum === creditSum,
    };
  }, [calculatedLines]);

  if (!open) return null;

  const handleApply = () => {
    if (!activePattern) return;
    const lines = calculatedLines.map((line) => ({
      accountcode: line.accountCode,
      description: line.descriptionTh,
      debit: line.side === "debit" && line.amount > 0 ? line.amount.toFixed(2) : "0",
      credit: line.side === "credit" && line.amount > 0 ? line.amount.toFixed(2) : "0",
      departmentcode: "",
      projectcode: "",
      cashflow: "",
    }));

    onApply({
      lines,
      description: activePattern.titleTh,
    });
    onClose();
  };

  const quickAmounts = [1000, 5000, 10000, 20000, 50000, 100000];

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-3 sm:p-5 animate-in fade-in duration-200">
      <div className="max-w-5xl w-full max-h-[92vh] bg-card border border-border rounded-2xl shadow-2xl flex flex-col overflow-hidden text-card-foreground">
        {/* Header */}
        <header className="p-4 border-b border-border bg-muted/40 flex items-center justify-between gap-3 shrink-0">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-amber-500/10 text-amber-500 border border-amber-500/20 shadow-sm">
              <Sparkles className="size-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-bold text-foreground">
                  {tr("gl_fast_templates_title", "แม่แบบบันทึกบัญชีด่วน (Thai Accounting Fast Templates)")}
                </h2>
                <span className="text-xs px-2.5 py-0.5 rounded-full bg-primary/10 text-primary font-semibold border border-primary/20">
                  {tr("gl_thai_business_tfrs", "8 กลุ่มธุรกิจไทย TFRS")}
                </span>
              </div>
              <p className="text-xs text-muted-foreground mt-0.5">
                {tr(
                  "gl_fast_templates_desc",
                  "เลือกรูปแบบคู่บัญชีมาตรฐาน คำนวณภาษีมูลค่าเพิ่ม (VAT 7%) และหัก ณ ที่จ่าย (WHT) อัตโนมัติ สมดุล 100%"
                )}
              </p>
            </div>
          </div>
          <Button variant="ghost" size="icon" className="size-8 rounded-lg" onClick={onClose}>
            <X className="size-4" />
          </Button>
        </header>

        {/* Business Types Filter Bar */}
        <div className="px-4 py-2.5 border-b border-border bg-muted/20 flex items-center gap-1.5 overflow-x-auto shrink-0 scrollbar-thin">
          {businessTypesList.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => {
                setSelectedType(item.id);
                const firstMatch = THAI_JOURNAL_PATTERNS.find(
                  (p) => item.id === "all" || p.businessType === item.id
                );
                if (firstMatch) setSelectedPatternId(firstMatch.id);
              }}
              className={`whitespace-nowrap text-xs font-semibold px-3 py-1.5 rounded-lg border transition-all cursor-pointer ${
                selectedType === item.id
                  ? "bg-primary text-primary-foreground border-primary shadow-sm"
                  : "bg-card text-muted-foreground border-border hover:bg-muted hover:text-foreground"
              }`}
            >
              {item.label}
            </button>
          ))}
        </div>

        {/* Search & Base Amount Input Bar */}
        <div className="p-4 border-b border-border bg-background grid gap-3 sm:grid-cols-12 shrink-0 items-center">
          <div className="sm:col-span-6 relative">
            <Search className="size-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <input
              type="text"
              className="w-full h-10 pl-9 pr-3 rounded-xl border border-border bg-muted/30 text-sm focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary shadow-sm"
              placeholder={tr("gl_search_template_ph", "ค้นหาชื่อรายการ, ผังบัญชี, ภาษี (เช่น ค่าเช่า, น้ำมัน, WHT 3%)...")}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          <div className="sm:col-span-6 flex items-center gap-2">
            <span className="text-xs font-semibold text-muted-foreground whitespace-nowrap">
              {tr("gl_base_amount", "มูลค่าก่อนภาษี (บาท):")}
            </span>
            <div className="flex-1 relative">
              <input
                type="number"
                step="0.01"
                min="0"
                className="w-full h-10 px-3 text-right font-mono font-bold text-sm rounded-xl border border-border bg-muted/30 focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary shadow-sm"
                value={baseAmount}
                onChange={(e) => setBaseAmount(e.target.value)}
              />
            </div>
            <div className="hidden lg:flex items-center gap-1 shrink-0">
              {quickAmounts.map((amt) => (
                <button
                  key={amt}
                  type="button"
                  onClick={() => setBaseAmount(String(amt))}
                  className="text-[11px] px-2 py-1 rounded-md border border-border bg-card hover:bg-muted font-mono font-medium text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
                >
                  {amt.toLocaleString()}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Main Content: Left List & Right Preview */}
        <div className="flex-1 min-h-0 grid grid-cols-1 md:grid-cols-12 overflow-hidden">
          {/* Left Column: Template Selection List */}
          <div className="md:col-span-5 border-r border-border overflow-y-auto p-3 flex flex-col gap-2 bg-muted/10">
            <div className="text-xs font-semibold text-muted-foreground px-1 mb-1 flex items-center justify-between">
              <span>{tr("gl_template_list", "รายการแม่แบบมาตรฐาน")}</span>
              <span>{filteredPatterns.length} {tr("gl_items_count", "รายการ")}</span>
            </div>
            {filteredPatterns.map((p) => {
              const isSelected = p.id === activePattern?.id;
              const meta = THAI_BUSINESS_METADATA[p.businessType];
              return (
                <button
                  key={p.id}
                  type="button"
                  onClick={() => setSelectedPatternId(p.id)}
                  className={`text-left p-3 rounded-xl border transition-all cursor-pointer flex flex-col gap-1.5 ${
                    isSelected
                      ? "bg-primary/10 border-primary text-foreground shadow-sm ring-1 ring-primary/30"
                      : "bg-card border-border text-card-foreground hover:bg-muted/60 hover:border-primary/40"
                  }`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-[11px] px-2 py-0.5 rounded-md font-semibold bg-muted border border-border text-muted-foreground">
                      {meta?.titleTh || p.businessType}
                    </span>
                    {isSelected && (
                      <span className="flex items-center gap-1 text-[11px] font-bold text-primary">
                        <CheckCircle2 className="size-3.5" />
                        {tr("gl_selected", "เลือกอยู่")}
                      </span>
                    )}
                  </div>
                  <div className="font-bold text-sm leading-tight text-foreground">
                    {p.titleTh}
                  </div>
                  <div className="text-xs text-muted-foreground line-clamp-1">
                    {p.lines.map((l) => `${l.side === "debit" ? "Dr." : "Cr."} ${l.accountNameTh}`).join(" | ")}
                  </div>
                </button>
              );
            })}
            {filteredPatterns.length === 0 && (
              <div className="p-8 text-center text-muted-foreground text-sm">
                {tr("gl_no_templates_found", "ไม่พบแม่แบบที่ตรงกับคำค้นหา")}
              </div>
            )}
          </div>

          {/* Right Column: Active Pattern Details & Live Calculated Table */}
          <div className="md:col-span-7 flex flex-col min-h-0 bg-background overflow-y-auto p-4 gap-4">
            {activePattern ? (
              <>
                {/* Pattern Info Card */}
                <div className="p-3.5 rounded-xl border border-border bg-muted/30 flex flex-col gap-2 shadow-sm">
                  <div className="flex items-center justify-between gap-2">
                    <h3 className="text-base font-bold text-foreground">
                      {activePattern.titleTh}
                    </h3>
                    {activePattern.tfrsReference && (
                      <span className="text-[11px] px-2 py-0.5 rounded-md bg-muted border border-border font-medium text-muted-foreground">
                        {activePattern.tfrsReference}
                      </span>
                    )}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {activePattern.titleEn}
                  </p>
                  {activePattern.taxNotesTh && (
                    <div className="mt-1 flex items-start gap-2 p-2.5 rounded-lg bg-amber-500/10 border border-amber-500/20 text-xs text-amber-700 dark:text-amber-300">
                      <Info className="size-4 shrink-0 mt-0.5 text-amber-500" />
                      <span>{activePattern.taxNotesTh}</span>
                    </div>
                  )}
                </div>

                {/* Calculated Journal Entries Table */}
                <div className="flex-1 flex flex-col gap-2">
                  <div className="flex items-center justify-between text-xs font-semibold text-muted-foreground">
                    <span>{tr("gl_calculated_preview", "ตัวอย่างการลงบัญชีเดบิต-เครดิต (คำนวณอัตโนมัติ)")}</span>
                    <span className="flex items-center gap-1 font-bold text-emerald-600 dark:text-emerald-400">
                      <Scale className="size-3.5" />
                      {totals.balanced
                        ? tr("gl_perfect_balanced", "สมดุล 100% (Dr = Cr)")
                        : tr("gl_unbalanced", "ไม่สมดุล")}
                    </span>
                  </div>

                  <div className="rounded-xl border border-border overflow-hidden shadow-sm">
                    <table className="w-full text-left text-[0.95rem] leading-normal">
                      <thead className="bg-muted text-xs font-semibold uppercase text-muted-foreground border-b border-border">
                        <tr>
                          <th className="p-2.5">{tr("gl_account", "รหัส / ชื่อบัญชี")}</th>
                          <th className="p-2.5">{tr("gl_description", "คำอธิบาย")}</th>
                          <th className="p-2.5 text-right">{tr("gl_debit", "เดบิต (Dr)")}</th>
                          <th className="p-2.5 text-right">{tr("gl_credit", "เครดิต (Cr)")}</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border bg-card">
                        {calculatedLines.map((line, idx) => {
                          const isDebit = line.side === "debit";
                          return (
                            <tr key={idx} className="hover:bg-muted/30 transition-colors">
                              <td className="p-2.5">
                                <div className="font-mono text-xs font-bold text-primary">
                                  {line.accountCode}
                                </div>
                                <div className="font-medium text-xs text-foreground">
                                  {line.accountNameTh}
                                </div>
                              </td>
                              <td className="p-2.5 text-xs text-muted-foreground">
                                {line.descriptionTh}
                              </td>
                              <td className="p-2.5 text-right font-mono font-bold tabular-nums text-foreground">
                                {isDebit && line.amount > 0 ? formatAmount(line.amount.toFixed(2)) : "—"}
                              </td>
                              <td className="p-2.5 text-right font-mono font-bold tabular-nums text-foreground">
                                {!isDebit && line.amount > 0 ? formatAmount(line.amount.toFixed(2)) : "—"}
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                      <tfoot className="bg-muted/60 border-t-2 border-border font-bold text-foreground">
                        <tr>
                          <td colSpan={2} className="p-2.5 text-right text-xs">
                            {tr("gl_total", "ยอดรวม")}
                          </td>
                          <td className="p-2.5 text-right font-mono tabular-nums text-emerald-600 dark:text-emerald-400">
                            {formatAmount(totals.debit.toFixed(2))}
                          </td>
                          <td className="p-2.5 text-right font-mono tabular-nums text-emerald-600 dark:text-emerald-400">
                            {formatAmount(totals.credit.toFixed(2))}
                          </td>
                        </tr>
                      </tfoot>
                    </table>
                  </div>
                </div>
              </>
            ) : (
              <div className="flex-1 flex items-center justify-center text-muted-foreground">
                {tr("gl_select_template_prompt", "กรุณาเลือกแม่แบบทางซ้ายมือ")}
              </div>
            )}
          </div>
        </div>

        {/* Footer Actions */}
        <footer className="p-3.5 border-t border-border bg-muted/40 flex items-center justify-between gap-3 shrink-0">
          <div className="text-xs text-muted-foreground flex items-center gap-1.5">
            <CheckCircle2 className="size-4 text-emerald-500" />
            <span>
              {tr(
                "gl_fast_template_safe_note",
                "ยอดคำนวณถูกปัดเศษตามหลักสตางค์และตรวจสอบสมดุลก่อนนำเข้าสมุดรายวัน"
              )}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={onClose}>
              {tr("gl_cancel", "ยกเลิก")}
            </Button>
            <Button
              type="button"
              size="sm"
              disabled={!activePattern || !totals.balanced || parsedBaseAmount <= 0}
              onClick={handleApply}
              className="bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm font-semibold"
            >
              <Sparkles className="size-4 mr-1.5" />
              {tr("gl_apply_to_journal", "นำไปใช้ในใบสำคัญรายวัน")}
            </Button>
          </div>
        </footer>
      </div>
    </div>
  );
}
