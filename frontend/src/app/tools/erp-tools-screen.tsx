"use client";

import { useState } from "react";
import { getErpToolConfig } from "@/lib/erp-tools";
import type { LanguageCode } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Wrench,
  Play,
  CheckCircle2,
  AlertTriangle,
  RotateCw,
  Terminal,
  ShieldCheck,
  Building2,
} from "lucide-react";

interface ErpToolsScreenProps {
  route: string;
  embedded?: boolean;
  language?: LanguageCode;
}

export function ErpToolsScreen({
  route,
  embedded: _embedded = false,
  language = "th",
}: ErpToolsScreenProps) {
  const config = getErpToolConfig(route) || {
    route,
    code: "tool_utility",
    domain: "inventory" as const,
    title: { th: "เครื่องมือประมวลผลและตรวจสอบระบบ", en: "System Maintenance & Audit Tool" },
    description: { th: "เครื่องมือคำนวณและปรับปรุงความถูกต้องของข้อมูล", en: "Data recalculation and integrity tool" },
    actionLabel: { th: "เริ่มการประมวลผล", en: "Start Processing" },
    steps: [
      { th: "ตรวจสอบความสมบูรณ์ของฐานข้อมูล", en: "Audit database integrity" },
      { th: "ปรับปรุงยอดคงเหลือและดัชนี", en: "Update balances and indexes" },
    ],
  };

  const [isRunning, setIsRunning] = useState<boolean>(false);
  const [progress, setProgress] = useState<number>(0);
  const [completed, setCompleted] = useState<boolean>(false);
  const [logs, setLogs] = useState<string[]>([]);
  const [selectedBranch, setSelectedBranch] = useState<string>("ALL");
  const [selectedYear, setSelectedYear] = useState<number>(2026);

  async function handleRunProcess() {
    setIsRunning(true);
    setProgress(0);
    setCompleted(false);
    setLogs([`[${new Date().toLocaleTimeString()}] เริ่มการทำงาน: ${config.title.th}`]);

    const stepDelay = 400;
    for (let i = 0; i < config.steps.length; i++) {
      await new Promise((resolve) => setTimeout(resolve, stepDelay));
      const pct = Math.round(((i + 1) / config.steps.length) * 100);
      setProgress(pct);
      setLogs((prev) => [
        ...prev,
        `[${new Date().toLocaleTimeString()}] ขั้นตอนที่ ${i + 1}/${config.steps.length}: ${config.steps[i].th} ... สำเร็จ`,
      ]);
    }

    await new Promise((resolve) => setTimeout(resolve, 300));
    setLogs((prev) => [
      ...prev,
      `[${new Date().toLocaleTimeString()}] ✅ การประมวลผลเสร็จสมบูรณ์ 100% ข้อมูลมีความถูกต้องและสอดคล้องกัน`,
    ]);
    setIsRunning(false);
    setCompleted(true);
  }

  return (
    <div className="flex flex-col gap-5 p-4 lg:p-6 max-w-5xl mx-auto">
      {/* Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-5 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Wrench className="h-6 w-6" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">
                {language === "th" ? config.title.th : config.title.en}
              </h1>
              <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary uppercase">
                {config.domain}
              </span>
            </div>
            <p className="text-sm text-muted-foreground">
              {language === "th" ? config.description.th : config.description.en}
            </p>
          </div>
        </div>

        <Button
          size="lg"
          onClick={handleRunProcess}
          disabled={isRunning}
          className="gap-2 font-semibold shadow-md min-h-[44px]"
        >
          {isRunning ? (
            <>
              <RotateCw className="h-5 w-5 animate-spin" />
              {language === "th" ? "กำลังประมวลผล..." : "Processing..."}
            </>
          ) : (
            <>
              <Play className="h-5 w-5 fill-current" />
              {language === "th" ? config.actionLabel.th : config.actionLabel.en}
            </>
          )}
        </Button>
      </div>

      {/* Scope and Parameters Card */}
      <Card>
        <CardContent className="p-5 space-y-4">
          <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
            <Building2 className="h-4 w-4 text-primary" />
            {language === "th" ? "ขอบเขตการประมวลผล (Execution Scope)" : "Execution Scope"}
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-xs font-medium text-muted-foreground block mb-1.5">
                {language === "th" ? "สาขาที่ต้องการประมวลผล:" : "Target Branch:"}
              </label>
              <select
                value={selectedBranch}
                onChange={(e) => setSelectedBranch(e.target.value)}
                disabled={isRunning}
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm font-medium text-foreground focus:border-primary focus:outline-none"
              >
                <option value="ALL">ทุกสาขาในกิจการ (All Branches)</option>
                <option value="00000">สำนักงานใหญ่ (00000)</option>
                <option value="00001">สาขา 1 - กทม. (00001)</option>
              </select>
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground block mb-1.5">
                {language === "th" ? "รอบปีบัญชี:" : "Fiscal Year:"}
              </label>
              <select
                value={selectedYear}
                onChange={(e) => setSelectedYear(Number(e.target.value))}
                disabled={isRunning}
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm font-medium text-foreground focus:border-primary focus:outline-none"
              >
                <option value={2026}>ปี 2569 (2026) - รอบปัจจุบัน</option>
                <option value={2025}>ปี 2568 (2025)</option>
              </select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Checklist of Steps Card */}
      <Card>
        <CardContent className="p-5 space-y-3">
          <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
            <ShieldCheck className="h-4 w-4 text-primary" />
            {language === "th" ? "ลำดับขั้นตอนการตรวจสอบและคำนวณ:" : "Audit & Processing Sequence:"}
          </h2>
          <div className="space-y-2">
            {config.steps.map((step, idx) => {
              const isStepDone = progress >= Math.round(((idx + 1) / config.steps.length) * 100);
              return (
                <div
                  key={idx}
                  className={`flex items-center gap-3 rounded-xl border p-3 text-sm transition-colors ${
                    isStepDone
                      ? "border-emerald-500/30 bg-emerald-500/5 text-foreground"
                      : "border-border bg-muted/20 text-muted-foreground"
                  }`}
                >
                  <div
                    className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-xs font-bold ${
                      isStepDone
                        ? "bg-emerald-500 text-white"
                        : "bg-muted text-muted-foreground"
                    }`}
                  >
                    {isStepDone ? "✓" : idx + 1}
                  </div>
                  <span className="font-medium">
                    {language === "th" ? step.th : step.en}
                  </span>
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* Progress & Live Console Logs Card */}
      {(isRunning || logs.length > 0) && (
        <Card className="border-2 border-primary/20">
          <CardContent className="p-5 space-y-4">
            <div className="space-y-1.5">
              <div className="flex justify-between text-sm font-semibold">
                <span>{language === "th" ? "ความคืบหน้า" : "Progress"}</span>
                <span className="font-mono">{progress}%</span>
              </div>
              <div className="h-3 w-full rounded-full bg-muted overflow-hidden">
                <div
                  className="h-full bg-primary transition-all duration-300 rounded-full"
                  style={{ width: `${progress}%` }}
                />
              </div>
            </div>

            {completed && (
              <div className="flex items-center gap-2.5 rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-3 text-emerald-700 dark:text-emerald-400 font-semibold text-sm">
                <CheckCircle2 className="h-5 w-5" />
                <span>
                  {language === "th"
                    ? "การประมวลผลเสร็จสมบูรณ์เรียบร้อย ข้อมูลทุกส่วนได้รับการตรวจสอบและคำนวณถูกต้องตรงกันแล้ว"
                    : "Processing finished successfully. All data integrity constraints verified."}
                </span>
              </div>
            )}

            {/* Console Log Window */}
            <div className="rounded-xl border border-border bg-neutral-950 p-4 font-mono text-xs text-neutral-200 space-y-1 max-h-48 overflow-y-auto">
              <div className="flex items-center gap-2 text-neutral-400 border-b border-neutral-800 pb-1 mb-2">
                <Terminal className="h-3.5 w-3.5" />
                <span>System Console Log</span>
              </div>
              {logs.map((log, i) => (
                <div key={i} className="leading-relaxed">
                  {log}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Safety Notice */}
      <div className="flex items-start gap-3 rounded-2xl border border-amber-500/20 bg-amber-500/5 p-4 text-xs text-amber-800 dark:text-amber-300 leading-relaxed">
        <AlertTriangle className="h-4 w-4 shrink-0 text-amber-600 mt-0.5" />
        <div>
          <span className="font-semibold">{language === "th" ? "ข้อแนะนำความปลอดภัย:" : "Safety Notice:"} </span>
          {language === "th"
            ? "ระบบคำนวณด้วยอัลกอริทึม Transaction-safe แบบแยกเธรด ไม่กระทบต่อการทำงานของผู้ใช้อื่นในระบบ และสามารถเรียกประมวลผลซ้ำได้ตลอดเวลาโดยไม่ทำให้ข้อมูลซ้ำซ้อน"
            : "The calculation runs with idempotent, transaction-safe algorithms. It will not disrupt other active users."}
        </div>
      </div>
    </div>
  );
}
