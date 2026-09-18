"use client";

import { useMemo, useState } from "react";
import {
  ShieldCheck,
  AlertOctagon,
  AlertTriangle,
  Info,
  CheckCircle2,
  X,
  ExternalLink,
  Activity,
  ArrowRight,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { type GLAccount, type GLReport } from "@/lib/general-ledger";
import { auditGLHealth, type AnomalySeverity } from "@/lib/gl-health-check";
import { useGLText } from "./gl-common";

export function GLHealthAuditModal({
  open,
  onClose,
  report,
  accounts = [],
  onDrillAccount,
}: {
  open: boolean;
  onClose: () => void;
  report: GLReport | null;
  accounts?: GLAccount[];
  onDrillAccount?: (accountCode: string) => void;
}) {
  const tr = useGLText();
  const [filterSeverity, setFilterSeverity] = useState<AnomalySeverity | "all">("all");

  const auditResult = useMemo(() => {
    if (!report) return null;
    return auditGLHealth({
      reportRows: report.rows ?? [],
      totals: report.totals ?? {},
      accounts,
    });
  }, [report, accounts]);

  if (!open || !auditResult) return null;

  const { score, status, statusTextTh, summary, anomalies } = auditResult;

  const filteredAnomalies = anomalies.filter(
    (a) => filterSeverity === "all" || a.severity === filterSeverity
  );

  const getStatusColor = (s: string) => {
    switch (s) {
      case "excellent":
        return "text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 border-emerald-500/30";
      case "good":
        return "text-blue-600 dark:text-blue-400 bg-blue-500/10 border-blue-500/30";
      case "warning":
        return "text-amber-600 dark:text-amber-400 bg-amber-500/10 border-amber-500/30";
      default:
        return "text-rose-600 dark:text-rose-400 bg-rose-500/10 border-rose-500/30";
    }
  };

  const getScoreCircleClass = (s: number) => {
    if (s >= 90) return "text-emerald-500 border-emerald-500/30 bg-emerald-500/5";
    if (s >= 75) return "text-blue-500 border-blue-500/30 bg-blue-500/5";
    if (s >= 50) return "text-amber-500 border-amber-500/30 bg-amber-500/5";
    return "text-rose-500 border-rose-500/30 bg-rose-500/5";
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-3 sm:p-5 animate-in fade-in duration-200">
      <div className="max-w-4xl w-full max-h-[92vh] bg-card border border-border rounded-2xl shadow-2xl flex flex-col overflow-hidden text-card-foreground">
        {/* Modal Header */}
        <header className="p-4 border-b border-border bg-muted/40 flex items-center justify-between gap-3 shrink-0">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20 shadow-sm">
              <ShieldCheck className="size-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-bold text-foreground">
                  {tr("gl_health_audit_title", "ระบบตรวจสุขภาพและกระทบยอดบัญชีแยกประเภท (GL Health Audit)")}
                </h2>
                <span className={`text-xs px-2.5 py-0.5 rounded-full font-bold border ${getStatusColor(status)}`}>
                  {statusTextTh}
                </span>
              </div>
              <p className="text-xs text-muted-foreground mt-0.5">
                {tr(
                  "gl_health_audit_subtitle",
                  "ตรวจจับยอดผิดฝั่งตามธรรมชาติผังบัญชี (Normal Balance), เงินสดติดลบ, ผลต่างงบทดลอง และบัญชีพักรอเคลียร์"
                )}
              </p>
            </div>
          </div>
          <Button variant="ghost" size="icon" className="size-8 rounded-lg" onClick={onClose}>
            <X className="size-4" />
          </Button>
        </header>

        {/* Score & Summary KPI Section */}
        <div className="p-4 border-b border-border bg-background grid grid-cols-2 sm:grid-cols-5 gap-3 shrink-0 items-center">
          {/* Health Score Gauge */}
          <div className="col-span-2 sm:col-span-1 flex flex-col items-center justify-center p-3 rounded-xl border border-border bg-muted/30">
            <div className="text-[11px] font-semibold text-muted-foreground mb-1">
              {tr("gl_health_score", "คะแนนสุขภาพ")}
            </div>
            <div
              className={`size-16 rounded-full border-4 flex items-center justify-center font-mono font-black text-xl shadow-inner ${getScoreCircleClass(
                score
              )}`}
            >
              {score}%
            </div>
          </div>

          {/* Accounts Analyzed */}
          <div className="p-3 rounded-xl border border-border bg-muted/20 flex flex-col justify-center">
            <span className="text-xs text-muted-foreground">{tr("gl_accounts_analyzed", "บัญชีที่ตรวจสอบ")}</span>
            <strong className="text-lg font-mono font-bold mt-1 text-foreground">
              {summary.totalAccountsAnalyzed}
            </strong>
            <span className="text-[11px] text-muted-foreground">{tr("gl_active_accounts", "บัญชีในงวด")}</span>
          </div>

          {/* Critical Count */}
          <div className="p-3 rounded-xl border border-rose-500/20 bg-rose-500/5 flex flex-col justify-center">
            <span className="text-xs text-rose-600 dark:text-rose-400 flex items-center gap-1 font-semibold">
              <AlertOctagon className="size-3.5" />
              {tr("gl_severity_critical", "วิกฤติ (Critical)")}
            </span>
            <strong className="text-lg font-mono font-bold mt-1 text-rose-600 dark:text-rose-400">
              {summary.criticalCount}
            </strong>
            <span className="text-[11px] text-muted-foreground">{tr("gl_must_fix", "ต้องแก้ไขทันที")}</span>
          </div>

          {/* Warning Count */}
          <div className="p-3 rounded-xl border border-amber-500/20 bg-amber-500/5 flex flex-col justify-center">
            <span className="text-xs text-amber-600 dark:text-amber-400 flex items-center gap-1 font-semibold">
              <AlertTriangle className="size-3.5" />
              {tr("gl_severity_warning", "ควรระวัง (Warning)")}
            </span>
            <strong className="text-lg font-mono font-bold mt-1 text-amber-600 dark:text-amber-400">
              {summary.warningCount}
            </strong>
            <span className="text-[11px] text-muted-foreground">{tr("gl_check_balance", "ตรวจยอดผิดฝั่ง")}</span>
          </div>

          {/* Info Count */}
          <div className="p-3 rounded-xl border border-blue-500/20 bg-blue-500/5 flex flex-col justify-center">
            <span className="text-xs text-blue-600 dark:text-blue-400 flex items-center gap-1 font-semibold">
              <Info className="size-3.5" />
              {tr("gl_severity_info", "ข้อสังเกต (Info)")}
            </span>
            <strong className="text-lg font-mono font-bold mt-1 text-blue-600 dark:text-blue-400">
              {summary.infoCount}
            </strong>
            <span className="text-[11px] text-muted-foreground">{tr("gl_suspense_dormant", "พัก/ไม่มีเคลื่อนไหว")}</span>
          </div>
        </div>

        {/* Filter Tabs */}
        <div className="px-4 py-2 border-b border-border bg-muted/20 flex items-center gap-2 overflow-x-auto shrink-0">
          {[
            { id: "all", label: `${tr("gl_all", "ทั้งหมด")} (${anomalies.length})` },
            { id: "critical", label: `${tr("gl_severity_critical", "วิกฤติ")} (${summary.criticalCount})` },
            { id: "warning", label: `${tr("gl_severity_warning", "ควรระวัง")} (${summary.warningCount})` },
            { id: "info", label: `${tr("gl_severity_info", "ข้อสังเกต")} (${summary.infoCount})` },
          ].map((tab) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => setFilterSeverity(tab.id as AnomalySeverity | "all")}
              className={`text-xs font-semibold px-3 py-1.5 rounded-lg border transition-all cursor-pointer ${
                filterSeverity === tab.id
                  ? "bg-primary text-primary-foreground border-primary shadow-sm"
                  : "bg-card text-muted-foreground border-border hover:bg-muted hover:text-foreground"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* Anomalies List */}
        <div className="flex-1 min-h-0 overflow-y-auto p-4 flex flex-col gap-3">
          {filteredAnomalies.map((item) => {
            const isCrit = item.severity === "critical";
            const isWarn = item.severity === "warning";
            return (
              <div
                key={item.id}
                className={`p-3.5 rounded-xl border flex flex-col gap-2 transition-all ${
                  isCrit
                    ? "bg-rose-500/5 border-rose-500/30 text-card-foreground shadow-sm"
                    : isWarn
                    ? "bg-amber-500/5 border-amber-500/30 text-card-foreground shadow-sm"
                    : "bg-blue-500/5 border-blue-500/30 text-card-foreground shadow-sm"
                }`}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span
                      className={`text-[11px] font-bold px-2 py-0.5 rounded-md uppercase border ${
                        isCrit
                          ? "bg-rose-500/20 text-rose-700 dark:text-rose-300 border-rose-500/40"
                          : isWarn
                          ? "bg-amber-500/20 text-amber-700 dark:text-amber-300 border-amber-500/40"
                          : "bg-blue-500/20 text-blue-700 dark:text-blue-300 border-blue-500/40"
                      }`}
                    >
                      {item.severity}
                    </span>
                    {item.accountCode && (
                      <span className="font-mono text-xs font-bold px-2 py-0.5 rounded bg-muted border border-border text-foreground">
                        {item.accountCode}
                      </span>
                    )}
                    {item.accountName && (
                      <span className="text-xs font-semibold text-foreground">
                        {item.accountName}
                      </span>
                    )}
                  </div>
                  {item.amount && (
                    <div className="font-mono text-sm font-bold tabular-nums text-foreground shrink-0">
                      {item.amount} {tr("gl_baht", "บาท")}
                    </div>
                  )}
                </div>

                <div className="text-xs text-foreground font-medium">
                  {item.messageTh}
                </div>

                <div className="p-2.5 rounded-lg bg-background/80 border border-border/80 flex items-start gap-2 text-xs text-muted-foreground">
                  <span className="font-semibold text-primary shrink-0">
                    {tr("gl_expert_recommendation", "คำแนะนำจากผู้เชี่ยวชาญ:")}
                  </span>
                  <span>{item.recommendationTh}</span>
                </div>

                {item.canDrillLedger && item.accountCode && onDrillAccount && (
                  <div className="flex justify-end pt-1">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="text-xs h-7 px-2.5 bg-primary/10 text-primary border-primary/20 hover:bg-primary/20"
                      onClick={() => {
                        onClose();
                        onDrillAccount(item.accountCode!);
                      }}
                    >
                      <ExternalLink className="size-3 mr-1" />
                      {tr("gl_drill_to_ledger", "ดูรายงานแยกประเภท")}
                    </Button>
                  </div>
                )}
              </div>
            );
          })}

          {filteredAnomalies.length === 0 && (
            <div className="flex flex-col items-center justify-center p-12 text-center gap-3">
              <div className="size-16 rounded-full bg-emerald-500/10 text-emerald-500 flex items-center justify-center">
                <CheckCircle2 className="size-8" />
              </div>
              <h3 className="text-base font-bold text-foreground">
                {tr("gl_no_anomalies_found", "ไม่พบข้อผิดพลาดตามเงื่อนไขที่เลือก")}
              </h3>
              <p className="text-xs text-muted-foreground max-w-md">
                {tr(
                  "gl_clean_health_audit",
                  "ยอดคงเหลือและรายการในงบการเงินมีความสมบูรณ์ตามมาตรฐานการบัญชี TFRS"
                )}
              </p>
            </div>
          )}
        </div>

        {/* Footer */}
        <footer className="p-3 border-t border-border bg-muted/30 flex items-center justify-between shrink-0">
          <span className="text-xs text-muted-foreground">
            {tr("gl_imbalance_check_note", "การตรวจสุขภาพบัญชีอิงตามผังบัญชีและข้อมูลสรุปของงวดปัจจุบัน")}
          </span>
          <Button variant="outline" size="sm" onClick={onClose}>
            {tr("gl_close", "ปิด")}
          </Button>
        </footer>
      </div>
    </div>
  );
}
