"use client";

import { useState, useMemo } from "react";
import { useBackendText } from "@/components/backend-text-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  ShieldAlert,
  ShieldCheck,
  CheckCircle2,
  AlertTriangle,
  AlertCircle,
  FileText,
  Scan,
  Sparkles,
  Search,
  ExternalLink,
  ChevronRight,
  ArrowRight,
  Check,
} from "lucide-react";
import {
  runPreClosingChecklist,
  detectDuplicateInvoices,
  parseSmartSlipOrReceipt,
  type PreClosingChecklistResult,
  type DuplicateInvoiceAlert,
} from "@/lib/ai-audit-guard";

interface AIAuditGuardModalProps {
  open: boolean;
  onClose: () => void;
}

export function AIAuditGuardModal({ open, onClose }: AIAuditGuardModalProps) {
  const tr = useBackendText();
  const [activeTab, setActiveTab] = useState<"checklist" | "duplicates" | "slip_ocr">("checklist");
  const [slipInputText, setSlipInputText] = useState("");
  const [slipResultMsg, setSlipResultMsg] = useState<string | null>(null);

  // Pre-closing sample data
  const checklistResult: PreClosingChecklistResult = useMemo(() => {
    return runPreClosingChecklist({
      cashOnHandBalance: 45200,
      isBankReconciled: true,
      unbalancedJournalsCount: 0,
      draftVouchersCount: 1, // warning
      depreciationPosted: true,
      stockCountAdjustmentDone: true,
      vatClosingDone: true,
      pendingWhtCertificatesCount: 2, // warning
      unclearedSuspenseCount: 0,
      overdueArDays90Count: 1, // warning
      branchIntercompanyMatched: true,
      taxProvisionEstimated: true,
    });
  }, []);

  // Sample duplicate invoices
  const duplicateAlerts: DuplicateInvoiceAlert[] = useMemo(() => {
    return detectDuplicateInvoices([
      {
        docNo: "PV2026-09012",
        taxId: "0105558000121",
        vendorOrCustomer: "Office Depot Co., Ltd.",
        invoiceNo: "INV-8877",
        date: "2026-09-02",
        totalAmount: 14200,
      },
      {
        docNo: "PV2026-09088",
        taxId: "0105558000121",
        vendorOrCustomer: "Office Depot Co., Ltd.",
        invoiceNo: "INV-8877",
        date: "2026-09-12",
        totalAmount: 14200,
      },
      {
        docNo: "PV2026-09015",
        taxId: "0105559000999",
        vendorOrCustomer: "Siam Logistics Co., Ltd.",
        invoiceNo: "LOG-01",
        date: "2026-09-05",
        totalAmount: 25000,
      },
      {
        docNo: "PV2026-09022",
        taxId: "0105559000999",
        vendorOrCustomer: "Siam Logistics Co., Ltd.",
        invoiceNo: "LOG-02",
        date: "2026-09-09",
        totalAmount: 25000,
      },
    ]);
  }, []);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-3 backdrop-blur-xs">
      <Card className="flex max-h-[90vh] w-full max-w-4xl flex-col border-border/80 bg-background shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b p-4 sm:p-5">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-inner">
              <ShieldAlert className="h-6 w-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-bold text-foreground">
                  {tr("gl_ai_audit_guard_title", "AI Audit Copilot & ระบบตรวจจับเงินรั่วไหล")}
                </h2>
                <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary">
                  Audit & Fraud Guard
                </span>
              </div>
              <p className="text-xs text-muted-foreground">
                {tr("gl_ai_audit_guard_subtitle", "ตรวจจับบิลซ้ำ การตั้งเบิกผิดปกติ สแกนสลิปโอนเงิน และเช็คลิสต์ 12 ข้อก่อนปิดงบตามมาตรฐานการสอบบัญชี")}
              </p>
            </div>
          </div>

          <Button variant="ghost" size="sm" onClick={onClose} className="h-8 w-8 p-0">
            ✕
          </Button>
        </div>

        {/* Tab Selection */}
        <div className="flex border-b bg-muted/30 px-4 pt-2">
          <button
            onClick={() => setActiveTab("checklist")}
            className={`border-b-2 px-4 py-2 text-xs font-semibold transition-all ${
              activeTab === "checklist"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            {tr("gl_ai_tab_checklist", "เช็คลิสต์ 12 ข้อก่อนปิดงบ")} ({checklistResult.readinessScore}%)
          </button>
          <button
            onClick={() => setActiveTab("duplicates")}
            className={`border-b-2 px-4 py-2 text-xs font-semibold transition-all flex items-center gap-1.5 ${
              activeTab === "duplicates"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            {tr("gl_ai_tab_duplicates", "ตรวจจับบิลซ้ำ / เงินรั่วไหล")}
            {duplicateAlerts.length > 0 && (
              <span className="rounded-full bg-rose-500 px-1.5 py-0.2 text-[10px] text-white">
                {duplicateAlerts.length}
              </span>
            )}
          </button>
          <button
            onClick={() => setActiveTab("slip_ocr")}
            className={`border-b-2 px-4 py-2 text-xs font-semibold transition-all ${
              activeTab === "slip_ocr"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            {tr("gl_ai_tab_slip_ocr", "AI สแกนสลิปโอนเงิน / ใบเสร็จ")}
          </button>
        </div>

        {/* Body Content */}
        <CardContent className="overflow-y-auto p-5">
          {/* Tab 1: Checklist */}
          {activeTab === "checklist" && (
            <div className="space-y-4">
              {/* Score Meter */}
              <div className="flex flex-col sm:flex-row items-center justify-between gap-4 rounded-xl border bg-muted/20 p-4">
                <div className="flex items-center gap-4">
                  <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary text-primary-foreground font-bold text-xl shadow-md">
                    {checklistResult.readinessScore}%
                  </div>
                  <div>
                    <h3 className="font-bold text-sm text-foreground">
                      {tr("gl_ai_readiness_title", "ความพร้อมในการปิดงบบัญชี (Closing Readiness)")}
                    </h3>
                    <p className="text-xs text-muted-foreground">
                      {tr("gl_ai_readiness_pass_prefix", "ผ่านแล้ว")} {checklistResult.passCount} {tr("gl_ai_readiness_items_unit", "ข้อ")} • {tr("gl_ai_readiness_warning_prefix", "ข้อควรระวัง")} {checklistResult.warningCount} {tr("gl_ai_readiness_items_unit", "ข้อ")} • {tr("gl_ai_readiness_critical_prefix", "วิกฤติ")} {checklistResult.criticalCount} {tr("gl_ai_readiness_items_unit", "ข้อ")}
                    </p>
                  </div>
                </div>

                <span
                  className={`rounded-full px-3 py-1 text-xs font-bold ${
                    checklistResult.overallStatus === "PASS"
                      ? "bg-emerald-500/15 text-emerald-700"
                      : checklistResult.overallStatus === "WARNING"
                        ? "bg-amber-500/15 text-amber-700"
                        : "bg-rose-500/15 text-rose-700"
                  }`}
                >
                  {checklistResult.overallStatus === "PASS"
                    ? tr("gl_ai_status_ready_all", "✓ พร้อมปิดงบการเงิน 100%")
                    : checklistResult.overallStatus === "WARNING"
                      ? tr("gl_ai_status_has_warnings", "มีรายการควรตรวจสอบก่อนปิดงบ")
                      : tr("gl_ai_status_has_critical", "ต้องแก้ไขรายการวิกฤติก่อน")}
                </span>
              </div>

              {/* 12 Checklist Items */}
              <div className="space-y-2">
                {checklistResult.items.map((item) => (
                  <div
                    key={item.id}
                    className="flex items-start justify-between gap-3 rounded-lg border p-3 text-xs bg-card transition-colors hover:border-primary/40"
                  >
                    <div className="flex items-start gap-2.5">
                      {item.status === "PASS" ? (
                        <CheckCircle2 className="h-4 w-4 shrink-0 text-emerald-600 mt-0.5" />
                      ) : item.status === "WARNING" ? (
                        <AlertTriangle className="h-4 w-4 shrink-0 text-amber-500 mt-0.5" />
                      ) : (
                        <AlertCircle className="h-4 w-4 shrink-0 text-rose-600 mt-0.5" />
                      )}
                      <div>
                        <span className="font-semibold text-foreground">{item.titleTh}</span>
                        <p className="text-muted-foreground mt-0.5 leading-relaxed">{item.detailTh}</p>
                      </div>
                    </div>

                    <span
                      className={`shrink-0 rounded px-2 py-0.5 text-[10px] font-bold ${
                        item.status === "PASS"
                          ? "bg-emerald-500/15 text-emerald-700"
                          : item.status === "WARNING"
                            ? "bg-amber-500/15 text-amber-700"
                            : "bg-rose-500/15 text-rose-700"
                      }`}
                    >
                      {item.status}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Tab 2: Duplicate Invoices */}
          {activeTab === "duplicates" && (
            <div className="space-y-4">
              <div className="rounded-lg border bg-rose-500/10 border-rose-500/20 p-3 text-xs text-rose-800">
                <span className="font-bold block">{tr("gl_ai_dup_alert_title", "ระบบตรวจจับความเสี่ยงการจ่ายเงินซ้ำและการตั้งเบิกซ้ำซ้อน:")}</span>
                {tr("gl_ai_dup_alert_desc_prefix", "ตรวจพบคู่รายการที่ต้องสงสัย")} {duplicateAlerts.length} {tr("gl_ai_dup_alert_desc_suffix", "รายการ กรุณาตรวจสอบก่อนอนุมัติจ่ายเงิน")}
              </div>

              <div className="space-y-3">
                {duplicateAlerts.map((alert) => (
                  <div
                    key={alert.id}
                    className="rounded-xl border border-border bg-card p-4 space-y-2 shadow-xs"
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-bold text-sm text-foreground">{alert.vendorName}</span>
                      <span
                        className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${
                          alert.severity === "CRITICAL"
                            ? "bg-rose-500/15 text-rose-700"
                            : "bg-amber-500/15 text-amber-700"
                        }`}
                      >
                        {alert.severity}
                      </span>
                    </div>

                    <p className="text-xs text-muted-foreground leading-relaxed">{alert.reasonTh}</p>

                    <div className="grid grid-cols-2 gap-2 border-t pt-2 text-xs">
                      <div>
                        <span className="text-muted-foreground">{tr("gl_ai_dup_doc_1", "เอกสารใบที่ 1")}:</span>{" "}
                        <strong className="text-primary">{alert.docNo}</strong> ({alert.date})
                      </div>
                      <div>
                        <span className="text-muted-foreground">{tr("gl_ai_dup_doc_2", "เอกสารใบที่ 2")}:</span>{" "}
                        <strong className="text-rose-600">{alert.conflictingDocNo}</strong> ({alert.conflictingDate})
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Tab 3: Smart Slip OCR */}
          {activeTab === "slip_ocr" && (
            <div className="space-y-4">
              <div className="rounded-lg border bg-muted/20 p-3 text-xs text-muted-foreground">
                <span className="font-bold text-foreground block">
                  AI Smart Slip Scanner (PromptPay / Mobile Banking):
                </span>
                {tr("gl_ai_slip_scanner_desc", "คัดลอกข้อความจากสลิปโอนเงิน หรือข้อความแจ้งเตือน SMS ธนาคาร วางลงในช่องด้านล่าง AI จะสกัด วันที่, ยอดเงิน, ธนาคาร, รหัสอ้างอิง และสร้างร่างใบสำคัญรับ/จ่ายให้ทันที")}
              </div>

              <textarea
                value={slipInputText}
                onChange={(e) => setSlipInputText(e.target.value)}
                placeholder="KBANK / SCB / BBL / PromptPay transfer slip text..."
                rows={5}
                className="w-full rounded-xl border border-input bg-background p-3 font-mono text-xs focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
              />

              <div className="flex justify-end gap-2">
                <Button
                  size="sm"
                  onClick={() => {
                    if (!slipInputText.trim()) return;
                    const res = parseSmartSlipOrReceipt(slipInputText);
                    setSlipResultMsg(
                      `${tr("gl_ai_slip_scan_success", "ถอดรหัสสำเร็จ")}: ${res.bankName} ${res.amount.toLocaleString()} (${res.transactionDate}) [${tr("gl_ai_slip_draft_voucher", "ร่างใบสำคัญ")} ${res.draftJournal.type}]`,
                    );
                  }}
                  className="gap-1.5"
                >
                  <Sparkles className="h-4 w-4" /> {tr("gl_ai_slip_scan_btn", "สแกนและสร้างใบสำคัญร่าง")}
                </Button>
              </div>

              {slipResultMsg && (
                <div className="flex items-center gap-2 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-3 text-xs text-emerald-800 font-semibold">
                  <Check className="h-4 w-4 text-emerald-600" />
                  <span>{slipResultMsg}</span>
                </div>
              )}
            </div>
          )}
        </CardContent>

        {/* Footer */}
        <div className="flex justify-end border-t p-4">
          <Button variant="outline" size="sm" onClick={onClose}>
            {tr("gl_close", "ปิด")}
          </Button>
        </div>
      </Card>
    </div>
  );
}
