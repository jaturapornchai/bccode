"use client";

import { useState } from "react";
import { getErpToolConfig, isErpToolApiReady, runErpTool, type ErpToolStockCheckStats } from "@/lib/erp-tools";
import type { LanguageCode } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Wrench,
  Play,
  CheckCircle2,
  AlertTriangle,
  AlertCircle,
  RotateCw,
  Terminal,
  ShieldCheck,
  Building2,
} from "lucide-react";

interface ErpToolsScreenProps {
  route: string;
  embedded?: boolean;
  language?: LanguageCode;
  holdingcode?: string;
  businesscode?: string;
}

const RESULT_MESSAGES: Record<string, { th: string; en: string }> = {
  process_success: {
    th: "ระบบประมวลผลเสร็จเรียบร้อยแล้ว",
    en: "Processing completed successfully",
  },
  process_failed: {
    th: "ประมวลผลไม่สำเร็จ กรุณาลองใหม่อีกครั้ง",
    en: "Processing failed, please try again",
  },
  tool_not_available: {
    th: "เครื่องมือนี้ยังไม่เชื่อมกับระบบประมวลผลจริง — อยู่ระหว่างเปิดใช้งาน API",
    en: "This tool is not connected to the processing API yet",
  },
  holding_required: {
    th: "ยังไม่ได้เลือกกิจการ",
    en: "No business selected",
  },
  unauthorized: {
    th: "ไม่มีสิทธิ์สั่งประมวลผลรายการนี้",
    en: "You do not have permission to run this process",
  },
  connection_error: {
    th: "เชื่อมต่อระบบไม่ได้",
    en: "Unable to connect to the system",
  },
};

function resultText(key: string, language: LanguageCode): string {
  const message = RESULT_MESSAGES[key];
  if (!message) return key;
  return language === "th" ? message.th : message.en;
}

export function ErpToolsScreen({
  route,
  embedded: _embedded = false,
  language = "th",
  holdingcode = "",
  businesscode = "",
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

  const currentYear = new Date().getFullYear();
  const [isRunning, setIsRunning] = useState<boolean>(false);
  const [resultKey, setResultKey] = useState<string | null>(null);
  const [succeeded, setSucceeded] = useState<boolean>(false);
  const [stats, setStats] = useState<ErpToolStockCheckStats | null>(null);
  const [logs, setLogs] = useState<string[]>([]);
  const [selectedYear, setSelectedYear] = useState<number>(currentYear);

  const apiReady = isErpToolApiReady(config.code);
  const canRun = apiReady && !isRunning && holdingcode !== "";

  async function handleRunProcess() {
    setIsRunning(true);
    setResultKey(null);
    setStats(null);
    setSucceeded(false);
    const title = language === "th" ? config.title.th : config.title.en;
    setLogs([`[${new Date().toLocaleTimeString()}] ${language === "th" ? "ส่งคำสั่งไปยังระบบ" : "Sending command"}: ${title}`]);

    const result = await runErpTool({
      code: config.code,
      holdingcode,
      businesscode,
      year: selectedYear,
    });

    const nextLogs: string[] = [
      `[${new Date().toLocaleTimeString()}] ${language === "th" ? "ระบบตอบกลับ" : "Response"}: ${resultText(result.messageKey, language)}`,
    ];
    if (result.stats) {
      nextLogs.push(
        `[${new Date().toLocaleTimeString()}] ${language === "th" ? "รายการต้นทุนที่ตรวจพบ" : "Cost rows found"}: ${result.stats.totalrows.toLocaleString("th-TH")}`,
      );
    }
    setLogs((prev) => [...prev, ...nextLogs]);
    setStats(result.stats ?? null);
    setResultKey(result.messageKey);
    setSucceeded(result.success);
    setIsRunning(false);
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
          disabled={!canRun}
          className="gap-2 font-semibold shadow-md min-h-[44px]"
          title={
            apiReady
              ? undefined
              : language === "th"
                ? "เครื่องมือนี้ยังไม่เชื่อมกับระบบประมวลผลจริง"
                : "This tool is not connected to the processing API yet"
          }
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

      {/* Availability Notice */}
      {!apiReady && (
        <div className="flex items-center gap-2 rounded-xl border border-border bg-muted/50 px-4 py-3 text-sm text-muted-foreground" role="status">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{resultText("tool_not_available", language)}</span>
        </div>
      )}
      {apiReady && holdingcode === "" && (
        <div className="flex items-center gap-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{resultText("holding_required", language)}</span>
        </div>
      )}

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
                {language === "th" ? "กิจการที่ประมวลผล:" : "Target Business:"}
              </label>
              <div className="w-full rounded-lg border border-border bg-muted/40 px-3 py-2 text-sm font-medium text-foreground">
                {holdingcode
                  ? `${holdingcode}${businesscode ? ` / ${businesscode}` : ""}`
                  : language === "th"
                    ? "ยังไม่ได้เลือกกิจการ"
                    : "No business selected"}
              </div>
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground block mb-1.5" htmlFor="tool-fiscal-year">
                {language === "th" ? "รอบปีบัญชี:" : "Fiscal Year:"}
              </label>
              <select
                id="tool-fiscal-year"
                value={selectedYear}
                onChange={(e) => setSelectedYear(Number(e.target.value))}
                disabled={isRunning}
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm font-medium text-foreground focus:border-primary focus:outline-none"
              >
                {[currentYear, currentYear - 1, currentYear - 2].map((year) => (
                  <option key={year} value={year}>
                    {language === "th" ? `ปี ${year + 543} (${year})` : `Year ${year}`}
                  </option>
                ))}
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
            {language === "th" ? "ขั้นตอนที่ระบบจะทำเมื่อสั่งประมวลผล:" : "What this process will do:"}
          </h2>
          <div className="space-y-2">
            {config.steps.map((step, idx) => (
              <div
                key={idx}
                className="flex items-center gap-3 rounded-xl border border-border bg-muted/20 p-3 text-sm text-foreground"
              >
                <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-bold text-muted-foreground">
                  {idx + 1}
                </div>
                <span className="font-medium">{language === "th" ? step.th : step.en}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Result & Console Logs Card */}
      {(isRunning || logs.length > 0) && (
        <Card className="border-2 border-primary/20">
          <CardContent className="p-5 space-y-4">
            {isRunning && (
              <div className="flex items-center gap-2.5 rounded-xl border border-border bg-muted/50 p-3 text-sm text-muted-foreground" role="status">
                <RotateCw className="h-5 w-5 animate-spin" />
                <span>
                  {language === "th"
                    ? "กำลังรอผลจากระบบ กรุณาอย่าปิดหน้าจอนี้"
                    : "Waiting for the server, please keep this screen open"}
                </span>
              </div>
            )}

            {!isRunning && resultKey && succeeded && (
              <div className="flex items-center gap-2.5 rounded-xl border border-primary/30 bg-primary/10 p-3 text-sm font-semibold text-primary" role="status">
                <CheckCircle2 className="h-5 w-5" />
                <span>{resultText(resultKey, language)}</span>
              </div>
            )}

            {!isRunning && resultKey && !succeeded && (
              <div className="flex items-center gap-2.5 rounded-xl border border-destructive/30 bg-destructive/10 p-3 text-sm font-semibold text-destructive" role="alert">
                <AlertCircle className="h-5 w-5" />
                <span>{resultText(resultKey, language)}</span>
              </div>
            )}

            {stats && (
              <div className="grid grid-cols-1 gap-2 rounded-xl border border-border bg-muted/30 p-3 text-sm text-foreground sm:grid-cols-3">
                <div>
                  <div className="text-xs text-muted-foreground">{language === "th" ? "จำนวนรายการต้นทุน" : "Cost rows"}</div>
                  <div className="font-mono font-semibold">{stats.totalrows.toLocaleString("th-TH")}</div>
                </div>
                <div>
                  <div className="text-xs text-muted-foreground">{language === "th" ? "จำนวนเอกสาร" : "Documents"}</div>
                  <div className="font-mono font-semibold">{stats.totaldocuments.toLocaleString("th-TH")}</div>
                </div>
                <div>
                  <div className="text-xs text-muted-foreground">{language === "th" ? "จำนวนสินค้า" : "Products"}</div>
                  <div className="font-mono font-semibold">{stats.totalproducts.toLocaleString("th-TH")}</div>
                </div>
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
            ? "การประมวลผลใหม่จะเขียนทับยอดที่คำนวณไว้เดิม และอาจใช้เวลานานเมื่อข้อมูลมีจำนวนมาก แนะนำให้สั่งประมวลผลนอกเวลาทำการ"
            : "Reprocessing overwrites previously calculated balances and may take a long time on large datasets. Run it outside business hours."}
        </div>
      </div>
    </div>
  );
}
