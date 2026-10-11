"use client";

import { useState, useMemo, useEffect, useCallback } from "react";
import {
  useBackendDictionary,
  useBackendText,
} from "@/components/backend-text-provider";
import {
  getErpReportConfig,
  isErpReportApiReady,
  reportColumnLabel,
  reportText,
  fetchErpReportData,
  agingColumnLabel,
  isAgingReport,
  DEFAULT_AGING_OVER_DAYS,
  DEFAULT_AGING_RANGES,
  type AgingCondition,
  type ErpReportRow,
  type ReportColumn,
} from "@/lib/erp-reports";
import { amountString, displayAmountUnits, formatAmount } from "@/lib/general-ledger";
import type { LanguageCode } from "@/lib/i18n";
import { useReportPreferences } from "@/hooks/use-report-preferences";
import { ReportDisplayToolbar } from "@/components/report-display-toolbar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { SmartBreadcrumb } from "@/components/smart-breadcrumb";
import {
  BarChart3,
  Download,
  Printer,
  Calendar,
  Search,
  ArrowUpDown,
  FileCode,
  Layers,
  Loader2,
  AlertCircle,
} from "lucide-react";

interface ErpReportViewerProps {
  route: string;
  embedded?: boolean;
  language?: LanguageCode;
  holdingcode?: string;
}

function formatDate(value: Date): string {
  const year = value.getFullYear();
  const month = String(value.getMonth() + 1).padStart(2, "0");
  const day = String(value.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

// ช่วงวันที่สำหรับส่งให้ API (คำนวณตามเวลาเครื่องผู้ใช้ ไม่ใช้ toISOString เพราะจะเพี้ยน timezone)
function computeDateRange(range: string): { fromdate: string; todate: string } {
  const now = new Date();
  const todate = formatDate(now);
  if (range === "today") return { fromdate: todate, todate };
  if (range === "week") {
    const from = new Date(now);
    from.setDate(from.getDate() - 6);
    return { fromdate: formatDate(from), todate };
  }
  if (range === "year") {
    return { fromdate: formatDate(new Date(now.getFullYear(), 0, 1)), todate };
  }
  return { fromdate: formatDate(new Date(now.getFullYear(), now.getMonth(), 1)), todate };
}

const ERROR_MESSAGES: Record<string, { key: string; th: string }> = {
  holding_required: { key: "holding_required", th: "ยังไม่ได้เลือกกิจการ" },
  unauthorized: { key: "rpt_msg_you_do_not_have_permission", th: "ไม่มีสิทธิ์เข้าถึงข้อมูลนี้" },
  load_failed: { key: "load_data_failed", th: "โหลดข้อมูลไม่สำเร็จ" },
  connection_error: { key: "ops_msg_unable_to_connect_to_the", th: "เชื่อมต่อระบบไม่ได้" },
  aging_ranges_invalid: { key: "aging_ranges_invalid", th: "ช่วงวันที่เกินกำหนดไม่ถูกต้อง" },
  aging_condition_invalid: { key: "aging_condition_invalid", th: "วันที่พิมพ์รายงานหรือวิธีนับอายุหนี้ไม่ถูกต้อง" },
};

// เงื่อนไขรายงานอายุหนี้ตามจอ Champ: ช่องกรอกเก็บเป็นข้อความระหว่างพิมพ์ แปลงเป็นตัวเลขตอนกดประมวลผล
interface AgingDraft {
  asofdate: string;
  agebasis: AgingCondition["agebasis"];
  ranges: { from: string; to: string }[];
  overdays: string;
}

const AGING_STORAGE_PREFIX = "bcai.report.aging.";

function agingDraftFrom(ranges: readonly { from: number; to: number }[], overdays: number): AgingDraft {
  return {
    asofdate: formatDate(new Date()),
    agebasis: "duedate",
    ranges: ranges.map((range) => ({ from: String(range.from), to: String(range.to) })),
    overdays: String(overdays),
  };
}

// ช่วงอายุที่ใช้ล่าสุดของแต่ละรายงาน (Champ เก็บใน ini ของเครื่อง) — ค่าเสียหาย/ไม่ครบใช้ค่าเริ่มต้นของ Champ
function loadAgingDraft(reportCode: string): AgingDraft {
  try {
    const saved: unknown = JSON.parse(window.localStorage.getItem(AGING_STORAGE_PREFIX + reportCode) ?? "null");
    if (saved && typeof saved === "object" && "ranges" in saved && "overdays" in saved) {
      const { ranges, overdays } = saved as { ranges: unknown; overdays: unknown };
      const isDay = (value: unknown): value is number => Number.isInteger(value) && (value as number) >= 0;
      if (
        Array.isArray(ranges) &&
        ranges.length === DEFAULT_AGING_RANGES.length &&
        ranges.every((range) => isDay(range?.from) && isDay(range?.to)) &&
        isDay(overdays)
      ) {
        return agingDraftFrom(ranges, overdays);
      }
    }
  } catch {
    // ค่าที่บันทึกไว้อ่านไม่ได้ ใช้ค่าเริ่มต้นแทน
  }
  return agingDraftFrom(DEFAULT_AGING_RANGES, DEFAULT_AGING_OVER_DAYS);
}

// ตรวจเบื้องต้นแค่ว่าเป็นจำนวนวันเต็ม ลำดับ/การทับกันของช่วง backend เป็นผู้ตรวจ
function agingConditionFrom(draft: AgingDraft): AgingCondition | null {
  const day = (text: string) => (/^\d{1,5}$/.test(text.trim()) ? Number.parseInt(text.trim(), 10) : null);
  const ranges = draft.ranges.map((range) => ({ from: day(range.from), to: day(range.to) }));
  const overdays = day(draft.overdays);
  if (overdays === null || ranges.some((range) => range.from === null || range.to === null)) return null;
  return {
    asofdate: draft.asofdate,
    agebasis: draft.agebasis,
    ranges: ranges.map((range) => ({ from: range.from ?? 0, to: range.to ?? 0 })),
    overdays,
  };
}

function amountUnits(value: unknown): bigint | null {
  if (typeof value !== "string") return null;
  try {
    return displayAmountUnits(value);
  } catch {
    return null;
  }
}

function formatTotal(column: ReportColumn, total: string | number): string {
  if (typeof total === "string") return formatAmount(total);
  return column.isCurrency
    ? total.toLocaleString("th-TH", { minimumFractionDigits: 2 })
    : total.toLocaleString("th-TH");
}

export function ErpReportViewer({
  route,
  embedded: _embedded = false,
  language = "th",
  holdingcode = "",
}: ErpReportViewerProps) {
  const tr = useBackendText();
  const dictionary = useBackendDictionary();
  const config = getErpReportConfig(route) || {
    route,
    code: "report_viewer",
    category: "inventory" as const,
    title: { th: "รายงานระบบ ERP", en: "ERP Report Viewer" },
    description: { th: "รายงานสรุปและวิเคราะห์ข้อมูลธุรกิจ", en: "Business Analytics & Report" },
    defaultSortKey: "",
    columns: [],
  };

  const [searchTerm, setSearchTerm] = useState<string>("");
  const [sortKey, setSortKey] = useState<string>(config.defaultSortKey);
  const [sortAsc, setSortAsc] = useState<boolean>(true);
  const [dateRange, setDateRange] = useState<string>("month");
  const reportPref = useReportPreferences();

  const [rawData, setRawData] = useState<ErpReportRow[]>([]);
  const [reportTotals, setReportTotals] = useState<ErpReportRow | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [errorKey, setErrorKey] = useState<string | null>(null);

  const reportCode = config.code;
  const isAging = isAgingReport(reportCode);
  const [agingDraft, setAgingDraft] = useState<AgingDraft>(() =>
    agingDraftFrom(DEFAULT_AGING_RANGES, DEFAULT_AGING_OVER_DAYS),
  );
  // null until the saved ranges are read, so the first fetch already uses them
  const [agingApplied, setAgingApplied] = useState<AgingCondition | null>(null);

  useEffect(() => {
    if (!isAging) return;
    const draft = loadAgingDraft(reportCode);
    setAgingDraft(draft);
    setAgingApplied(agingConditionFrom(draft));
  }, [isAging, reportCode]);

  const loadData = useCallback(async () => {
    if (!isErpReportApiReady(reportCode)) {
      setRawData([]);
      setReportTotals(null);
      setErrorKey("report_not_available");
      setLoading(false);
      return;
    }
    if (isAging && !agingApplied) return;
    setLoading(true);
    setErrorKey(null);
    const { fromdate, todate } = computeDateRange(dateRange);
    const result = await fetchErpReportData({
      code: reportCode,
      holdingcode,
      fromdate,
      todate,
      aging: agingApplied ?? undefined,
    });
    setRawData(result.rows);
    setReportTotals(result.totals ?? null);
    setErrorKey(result.error ?? null);
    setLoading(false);
    if (agingApplied && !result.error) {
      try {
        window.localStorage.setItem(
          AGING_STORAGE_PREFIX + reportCode,
          JSON.stringify({ ranges: agingApplied.ranges, overdays: agingApplied.overdays }),
        );
      } catch {
        // เก็บค่าไม่ได้ (เช่น storage เต็ม) ไม่กระทบรายงาน
      }
    }
  }, [reportCode, holdingcode, dateRange, isAging, agingApplied]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  function handleAgingProcess() {
    const condition = agingConditionFrom(agingDraft);
    if (!condition) {
      setRawData([]);
      setReportTotals(null);
      setErrorKey("aging_ranges_invalid");
      return;
    }
    setAgingApplied(condition);
  }

  function updateAgingRange(index: number, field: "from" | "to", value: string) {
    setAgingDraft((draft) => ({
      ...draft,
      ranges: draft.ranges.map((range, i) => (i === index ? { ...range, [field]: value } : range)),
    }));
  }

  function columnLabel(col: ReportColumn): string {
    return (
      (agingApplied && agingColumnLabel(col.key, agingApplied, language, dictionary)) ||
      reportColumnLabel(config, col, language, dictionary)
    );
  }

  // Filtered & Sorted Data
  const processedData = useMemo(() => {
    let result = [...rawData];

    if (searchTerm.trim()) {
      const term = searchTerm.toLowerCase();
      result = result.filter((row) =>
        Object.values(row).some((val) =>
          String(val ?? "").toLowerCase().includes(term),
        ),
      );
    }

    if (sortKey) {
      const sortsAmountText = config.columns.some((col) => col.key === sortKey && col.isCurrency);
      result.sort((a, b) => {
        const valA = a[sortKey];
        const valB = b[sortKey];
        if (typeof valA === "number" && typeof valB === "number") {
          return sortAsc ? valA - valB : valB - valA;
        }
        const unitsA = sortsAmountText ? amountUnits(valA) : null;
        const unitsB = sortsAmountText ? amountUnits(valB) : null;
        if (unitsA !== null && unitsB !== null) {
          const order = unitsA < unitsB ? -1 : unitsA > unitsB ? 1 : 0;
          return sortAsc ? order : -order;
        }
        return sortAsc
          ? String(valA ?? "").localeCompare(String(valB ?? ""))
          : String(valB ?? "").localeCompare(String(valA ?? ""));
      });
    }

    return result;
  }, [rawData, searchTerm, sortKey, sortAsc, config.columns]);

  // Totals for numeric columns: backend totals cover every row of the report; while searching,
  // amount text is summed exactly over the visible rows.
  const columnTotals = useMemo(() => {
    const totals: Record<string, string | number> = {};
    const filtered = searchTerm.trim() !== "";
    for (const col of config.columns) {
      if (!col.isNumeric && !col.isCurrency) continue;
      const backendTotal = filtered ? undefined : reportTotals?.[col.key];
      if (typeof backendTotal === "string" || typeof backendTotal === "number") {
        totals[col.key] = backendTotal;
      } else if (processedData.some((row) => typeof row[col.key] === "string")) {
        const units = processedData.reduce((sum, row) => sum + (amountUnits(row[col.key]) ?? 0n), 0n);
        totals[col.key] = amountString(units, 2);
      } else {
        totals[col.key] = processedData.reduce(
          (sum, row) => sum + (Number(row[col.key]) || 0),
          0,
        );
      }
    }
    return totals;
  }, [config.columns, processedData, reportTotals, searchTerm]);

  function handleSort(key: string) {
    if (sortKey === key) {
      setSortAsc(!sortAsc);
    } else {
      setSortKey(key);
      setSortAsc(true);
    }
  }

  function handleExportCsv() {
    const headers = config.columns.map((col) => columnLabel(col));
    const rows = processedData.map((row) =>
      config.columns.map((col) => `"${row[col.key] ?? ""}"`).join(","),
    );
    const csvContent = [headers.join(","), ...rows].join("\n");
    const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", `${config.code}_report.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  // ส่งออกได้เฉพาะเมื่อข้อมูลมาจาก API จริงเท่านั้น
  const canExport = !loading && !errorKey && processedData.length > 0;

  return (
    <div className="flex flex-col gap-4 p-4 lg:p-6">
      {!_embedded && <SmartBreadcrumb currentTitle={reportText(config, "title", language, dictionary)} />}

      {/* Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-4 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
            {config.category === "xbrl" ? <FileCode className="h-6 w-6" /> : <BarChart3 className="h-6 w-6" />}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">
                {reportText(config, "title", language, dictionary)}
              </h1>
              <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary uppercase">
                {config.category}
              </span>
            </div>
            <p className="text-sm text-muted-foreground">
              {reportText(config, "description", language, dictionary)}
            </p>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {config.category === "xbrl" && (
            <Button
              variant="outline"
              disabled
              className="gap-2"
              title={
                tr("ops_not_available_yet_requires_real", "ยังไม่เปิดใช้งาน — ต้องมีข้อมูลงบการเงินจริงและเลขประจำตัวผู้เสียภาษีของกิจการจากระบบก่อน")
              }
            >
              <FileCode className="h-4 w-4" />
              {tr("ops_export_dbd_xbrl", "ส่งออก XBRL (ยื่น DBD)")}
            </Button>
          )}

          <Button variant="outline" onClick={handleExportCsv} disabled={!canExport} className="gap-2">
            <Download className="h-4 w-4" />
            {tr("gl_export_csv", "ส่งออก CSV")}
          </Button>

          <Button variant="outline" onClick={() => window.print()} disabled={!canExport} className="gap-2">
            <Printer className="h-4 w-4" />
            {tr("print_report", "พิมพ์รายงาน")}
          </Button>
        </div>
      </div>

      {/* Filter and Date Control Bar */}
      <Card>
        <CardContent className="flex flex-wrap items-center justify-between gap-4 p-4">
          <div className="flex flex-wrap items-center gap-2">
            {!isAging && (
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground mr-2">
                <Calendar className="h-4 w-4 text-primary" />
                <span>{tr("ops_period", "ช่วงเวลา:")}</span>
              </div>
            )}
            {!isAging && (["today", "week", "month", "year"] as const).map((period) => (
              <Button
                key={period}
                size="sm"
                variant={dateRange === period ? "default" : "outline"}
                onClick={() => setDateRange(period)}
                className="h-8 text-xs"
              >
                {period === "today" && (tr("alert_today", "วันนี้"))}
                {period === "week" && (tr("alert_this_week", "สัปดาห์นี้"))}
                {period === "month" && (tr("alert_this_month", "เดือนนี้"))}
                {period === "year" && (tr("ops_this_year", "ปีนี้"))}
              </Button>
            ))}

            <div className="ml-3 flex items-center gap-1.5 rounded-lg bg-muted px-3 py-1 text-xs text-muted-foreground">
              <Layers className="h-3.5 w-3.5" />
              <span>{tr("gl_all_branches", "ทุกสาขา")}</span>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <ReportDisplayToolbar
              fontSize={reportPref.fontSize}
              onFontSizeChange={reportPref.setFontSize}
              highContrast={reportPref.highContrast}
              onToggleHighContrast={reportPref.toggleHighContrast}
            />

            <div className="relative w-full max-w-xs">
              <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder={tr("ops_search_in_report", "ค้นหาในรายงาน...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-9 h-9"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* เงื่อนไขรายงานอายุหนี้ (ตามจอ Champ): วันที่พิมพ์รายงาน วิธีนับอายุ และช่วงวันที่เกินกำหนด 4 ช่วง + เกินกว่า */}
      {isAging && (
        <Card>
          <CardContent className="flex flex-wrap items-end gap-x-6 gap-y-4 p-4 text-sm">
            <label className="flex flex-col gap-1.5">
              <span className="font-medium text-muted-foreground">{tr("aging_as_of_date", "พิมพ์รายงานประจำวันที่")}</span>
              <Input
                type="date"
                value={agingDraft.asofdate}
                onChange={(e) => setAgingDraft((draft) => ({ ...draft, asofdate: e.target.value }))}
                className="h-11 w-44"
              />
            </label>

            <fieldset className="flex flex-col gap-1.5">
              <legend className="mb-1.5 font-medium text-muted-foreground">{tr("aging_basis_label", "นับอายุหนี้")}</legend>
              <div className="flex flex-wrap items-center gap-4">
                {(["duedate", "docdate"] as const).map((basis) => (
                  <label key={basis} className="flex min-h-11 cursor-pointer items-center gap-2">
                    <input
                      type="radio"
                      name={`aging-basis-${reportCode}`}
                      value={basis}
                      checked={agingDraft.agebasis === basis}
                      onChange={() => setAgingDraft((draft) => ({ ...draft, agebasis: basis }))}
                      className="h-4 w-4 accent-primary"
                    />
                    <span>
                      {basis === "duedate"
                        ? tr("aging_basis_due_date", "เริ่มจากวันถึงกำหนด")
                        : tr("aging_basis_doc_date", "เริ่มจากวันที่เอกสาร")}
                    </span>
                  </label>
                ))}
              </div>
            </fieldset>

            <fieldset className="flex flex-col gap-1.5">
              <legend className="mb-1.5 font-medium text-muted-foreground">{tr("aging_ranges_title", "กำหนดช่วงวันที่เกินกำหนด")}</legend>
              <div className="flex flex-wrap items-center gap-x-5 gap-y-3">
                {agingDraft.ranges.map((range, index) => {
                  const rangeName = tr("aging_range_n", "ช่วงที่ {n}").replace("{n}", String(index + 1));
                  return (
                    <div key={index} className="flex items-center gap-1.5">
                      <span className="text-muted-foreground">{rangeName}</span>
                      <Input
                        inputMode="numeric"
                        value={range.from}
                        onChange={(e) => updateAgingRange(index, "from", e.target.value)}
                        aria-label={rangeName}
                        className="h-11 w-16 text-right"
                      />
                      <span>{tr("to", "ถึง")}</span>
                      <Input
                        inputMode="numeric"
                        value={range.to}
                        onChange={(e) => updateAgingRange(index, "to", e.target.value)}
                        aria-label={`${rangeName} ${tr("to", "ถึง")}`}
                        className="h-11 w-16 text-right"
                      />
                    </div>
                  );
                })}
                <div className="flex items-center gap-1.5">
                  <span className="text-muted-foreground">{tr("aging_over_than", "เกินกว่า")}</span>
                  <Input
                    inputMode="numeric"
                    value={agingDraft.overdays}
                    onChange={(e) => setAgingDraft((draft) => ({ ...draft, overdays: e.target.value }))}
                    aria-label={tr("aging_over_than", "เกินกว่า")}
                    className="h-11 w-16 text-right"
                  />
                  <span>{tr("days", "วัน")}</span>
                </div>
              </div>
            </fieldset>

            <Button onClick={handleAgingProcess} disabled={loading} className="h-11 min-w-28 gap-2">
              <BarChart3 className="h-4 w-4" />
              {tr("process", "ประมวลผล")}
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Data Status */}
      {loading && (
        <div className="flex items-center gap-2 rounded-xl border border-border bg-muted/50 px-4 py-3 text-sm text-muted-foreground" role="status">
          <Loader2 className="h-4 w-4 animate-spin" />
          <span>{tr("ops_loading_data", "กำลังโหลดข้อมูลจากระบบ...")}</span>
        </div>
      )}

      {!loading && errorKey === "report_not_available" && (
        <div className="flex items-center gap-2 rounded-xl border border-border bg-muted/50 px-4 py-3 text-sm text-muted-foreground" role="status">
          <AlertCircle className="h-4 w-4" />
          <span>
            {tr("ops_this_report_is_not_connected", "รายงานนี้ยังไม่เชื่อมกับข้อมูลจริง — อยู่ระหว่างเปิดใช้งาน API")}
          </span>
        </div>
      )}

      {!loading && errorKey && errorKey !== "report_not_available" && (
        <div className="flex items-center gap-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
          <AlertCircle className="h-4 w-4" />
          <span>
            {ERROR_MESSAGES[errorKey]
              ? tr(ERROR_MESSAGES[errorKey].key, ERROR_MESSAGES[errorKey].th)
              : errorKey}
          </span>
        </div>
      )}

      {!loading && !errorKey && rawData.length === 0 && (
        <div className="rounded-xl border border-border bg-muted/50 px-4 py-3 text-sm text-muted-foreground" role="status">
          {tr("ops_no_data_found_for_the", "ไม่พบข้อมูลในช่วงเวลาที่เลือก")}
        </div>
      )}

      {/* Report Data Table */}
      <Card className={`overflow-hidden ${reportPref.contrastClass}`}>
        <div className="overflow-x-auto">
          <table className={`w-full text-left leading-normal ${reportPref.fontSizeClass}`}>
            <thead className="border-b bg-muted/60 text-xs font-semibold text-muted-foreground uppercase">
              <tr>
                <th className="px-3 py-3 w-12 text-center">{tr("sequence", "ลำดับ")}</th>
                {config.columns.map((col) => (
                  <th
                    key={col.key}
                    className={`px-4 py-3 cursor-pointer select-none hover:bg-muted/80 transition-colors ${
                      col.align === "right"
                        ? "text-right"
                        : col.align === "center"
                          ? "text-center"
                          : "text-left"
                    }`}
                    onClick={() => handleSort(col.key)}
                  >
                    <div className={`flex items-center gap-1.5 ${col.align === "right" ? "justify-end" : col.align === "center" ? "justify-center" : "justify-start"}`}>
                      <span>{columnLabel(col)}</span>
                      <ArrowUpDown className="h-3 w-3 opacity-60" />
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {processedData.length === 0 ? (
                <tr>
                  <td colSpan={config.columns.length + 1} className="py-12 text-center text-muted-foreground">
                    {tr("ops_no_matching_records_found", "ไม่พบข้อมูลที่ตรงกับเงื่อนไขการค้นหา")}
                  </td>
                </tr>
              ) : (
                processedData.map((row, idx) => (
                  <tr key={idx} className="hover:bg-muted/30 transition-colors">
                    <td className="px-3 py-2.5 text-center text-muted-foreground font-mono text-xs">{idx + 1}</td>
                    {config.columns.map((col) => {
                      const val = row[col.key];
                      return (
                        <td
                          key={col.key}
                          className={`px-4 py-2.5 ${
                            col.align === "right"
                              ? "text-right font-mono"
                              : col.align === "center"
                                ? "text-center"
                                : "text-left"
                          } ${col.isCurrency ? "font-medium" : ""}`}
                        >
                          {col.isCurrency && typeof val === "string"
                            ? formatAmount(val)
                            : col.isCurrency && typeof val === "number"
                            ? val.toLocaleString("th-TH", { minimumFractionDigits: 2 })
                            : col.isNumeric && typeof val === "number"
                              ? val.toLocaleString("th-TH")
                              : String(val ?? "-")}
                        </td>
                      );
                    })}
                  </tr>
                ))
              )}
            </tbody>
            {processedData.length > 0 && (
              <tfoot className="border-t-2 bg-muted/40 font-bold">
                <tr>
                  <td className="px-3 py-3 text-center text-xs">{tr("total", "รวม")}</td>
                  {config.columns.map((col, idx) => {
                    const total = columnTotals[col.key];
                    if (idx === 0 && total === undefined) {
                      return (
                        <td key={col.key} className="px-4 py-3 text-xs text-muted-foreground">
                          {processedData.length} {tr("items", "รายการ")}
                        </td>
                      );
                    }
                    if (total !== undefined) {
                      return (
                        <td key={col.key} className="px-4 py-3 text-right font-mono text-primary">
                          {formatTotal(col, total)}
                        </td>
                      );
                    }
                    return <td key={col.key}></td>;
                  })}
                </tr>
              </tfoot>
            )}
          </table>
        </div>
      </Card>
    </div>
  );
}
