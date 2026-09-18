"use client";

import { useState } from "react";
import {
  useBackendDictionary,
  useBackendText,
  type BackendTextFn,
} from "@/components/backend-text-provider";
import {
  getErpToolConfig,
  isErpToolApiReady,
  runErpTool,
  toolStepText,
  toolText,
} from "@/lib/erp-tools";
import type { LanguageCode } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { ChoiceSelect } from "@/components/ui/select";
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

const RESULT_MESSAGES: Record<string, { key: string; th: string }> = {
  process_success: {
    key: "tool_msg_processing_completed_successfully",
    th: "ระบบประมวลผลเสร็จเรียบร้อยแล้ว",
  },
  process_failed: {
    key: "tool_msg_processing_failed_please_try_again",
    th: "ประมวลผลไม่สำเร็จ กรุณาลองใหม่อีกครั้ง",
  },
  tool_not_available: {
    key: "tool_msg_this_tool_is_not_connected",
    th: "เครื่องมือนี้ยังไม่เชื่อมกับระบบประมวลผลจริง — อยู่ระหว่างเปิดใช้งาน API",
  },
  holding_required: {
    key: "holding_required",
    th: "ยังไม่ได้เลือกกิจการ",
  },
  unauthorized: {
    key: "tool_msg_you_do_not_have_permission",
    th: "ไม่มีสิทธิ์สั่งประมวลผลรายการนี้",
  },
  connection_error: {
    key: "ops_msg_unable_to_connect_to_the",
    th: "เชื่อมต่อระบบไม่ได้",
  },
};

function resultText(key: string, tr: BackendTextFn): string {
  const message = RESULT_MESSAGES[key];
  if (!message) return key;
  return tr(message.key, message.th);
}

export function ErpToolsScreen({
  route,
  embedded: _embedded = false,
  language = "th",
  holdingcode = "",
  businesscode = "",
}: ErpToolsScreenProps) {
  const tr = useBackendText();
  const dictionary = useBackendDictionary();
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
  const [logs, setLogs] = useState<string[]>([]);
  const [selectedYear, setSelectedYear] = useState<number>(currentYear);

  const apiReady = isErpToolApiReady(config.code);
  const canRun = apiReady && !isRunning && holdingcode !== "";

  async function handleRunProcess() {
    setIsRunning(true);
    setResultKey(null);
    setSucceeded(false);
    const title = toolText(config, "title", language, dictionary);
    setLogs([`[${new Date().toLocaleTimeString()}] ${tr("ops_sending_command", "ส่งคำสั่งไปยังระบบ")}: ${title}`]);

    const result = await runErpTool({
      code: config.code,
      holdingcode,
      businesscode,
      year: selectedYear,
    });

    const nextLogs: string[] = [
      `[${new Date().toLocaleTimeString()}] ${tr("ops_response", "ระบบตอบกลับ")}: ${resultText(result.messageKey, tr)}`,
    ];
    setLogs((prev) => [...prev, ...nextLogs]);
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
                {toolText(config, "title", language, dictionary)}
              </h1>
              <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary uppercase">
                {config.domain}
              </span>
            </div>
            <p className="text-sm text-muted-foreground">
              {toolText(config, "description", language, dictionary)}
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
              : tr("ops_this_tool_is_not_connected", "เครื่องมือนี้ยังไม่เชื่อมกับระบบประมวลผลจริง")
          }
        >
          {isRunning ? (
            <>
              <RotateCw className="h-5 w-5 animate-spin" />
              {tr("ops_processing", "กำลังประมวลผล...")}
            </>
          ) : (
            <>
              <Play className="h-5 w-5 fill-current" />
              {toolText(config, "actionLabel", language, dictionary)}
            </>
          )}
        </Button>
      </div>

      {/* Availability Notice */}
      {!apiReady && (
        <div className="flex items-center gap-2 rounded-xl border border-border bg-muted/50 px-4 py-3 text-sm text-muted-foreground" role="status">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{resultText("tool_not_available", tr)}</span>
        </div>
      )}
      {apiReady && holdingcode === "" && (
        <div className="flex items-center gap-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{resultText("holding_required", tr)}</span>
        </div>
      )}

      {/* Scope and Parameters Card */}
      <Card>
        <CardContent className="p-5 space-y-4">
          <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
            <Building2 className="h-4 w-4 text-primary" />
            {tr("ops_execution_scope", "ขอบเขตการประมวลผล (Execution Scope)")}
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-xs font-medium text-muted-foreground block mb-1.5">
                {tr("ops_target_business", "กิจการที่ประมวลผล:")}
              </label>
              <div className="w-full rounded-lg border border-border bg-muted/40 px-3 py-2 text-sm font-medium text-foreground">
                {holdingcode
                  ? `${holdingcode}${businesscode ? ` / ${businesscode}` : ""}`
                  : tr("holding_required", "ยังไม่ได้เลือกกิจการ")}
              </div>
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground block mb-1.5" htmlFor="tool-fiscal-year">
                {tr("ops_fiscal_year", "รอบปีบัญชี:")}
              </label>
              <ChoiceSelect
                id="tool-fiscal-year"
                value={selectedYear}
                onChange={(val) => setSelectedYear(Number(val))}
                disabled={isRunning}
                options={[currentYear, currentYear - 1, currentYear - 2].map((year) => ({
                  value: year,
                  label: tr("tool_fiscal_year_option", "ปี {0} ({1})")
                    .replace("{0}", String(year + 543))
                    .replace("{1}", String(year)),
                }))}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Checklist of Steps Card */}
      <Card>
        <CardContent className="p-5 space-y-3">
          <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
            <ShieldCheck className="h-4 w-4 text-primary" />
            {tr("ops_what_this_process_will_do", "ขั้นตอนที่ระบบจะทำเมื่อสั่งประมวลผล:")}
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
                <span className="font-medium">{toolStepText(config, idx, language, dictionary)}</span>
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
                  {tr("ops_waiting_for_the_server_please", "กำลังรอผลจากระบบ กรุณาอย่าปิดหน้าจอนี้")}
                </span>
              </div>
            )}

            {!isRunning && resultKey && succeeded && (
              <div className="flex items-center gap-2.5 rounded-xl border border-primary/30 bg-primary/10 p-3 text-sm font-semibold text-primary" role="status">
                <CheckCircle2 className="h-5 w-5" />
                <span>{resultText(resultKey, tr)}</span>
              </div>
            )}

            {!isRunning && resultKey && !succeeded && (
              <div className="flex items-center gap-2.5 rounded-xl border border-destructive/30 bg-destructive/10 p-3 text-sm font-semibold text-destructive" role="alert">
                <AlertCircle className="h-5 w-5" />
                <span>{resultText(resultKey, tr)}</span>
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
          <span className="font-semibold">{tr("ops_safety_notice", "ข้อแนะนำความปลอดภัย:")} </span>
          {tr("ops_reprocessing_overwrites_previously_calculated_balances", "การประมวลผลใหม่จะเขียนทับยอดที่คำนวณไว้เดิม และอาจใช้เวลานานเมื่อข้อมูลมีจำนวนมาก แนะนำให้สั่งประมวลผลนอกเวลาทำการ")}
        </div>
      </div>
    </div>
  );
}
