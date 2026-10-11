"use client";

// Central report screen: ONE screen for every report the backend registry (GET /report/v1/registry) offers.
// The form is built from the registry's parameter list, the result is drawn by the shared ReportGrid and
// printed / exported by the backend contract, so a new report needs no frontend code (docs/kms/champ/10-reports.md §0).
import { useEffect, useMemo, useRef, useState } from "react";
import { Download, Printer, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Combobox } from "@/components/ui/combobox";
import { journalBookName, type GLAccount, type GLBudget, type GLDimension, type GLPage } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import {
  REPORT_PAGE_SIZE, ReportError, buildRunRequest, defaultParamValues, downloadBlob, drillFiscalYear, exportReportCsv, fetchReportRegistry, fillBlankValues,
  printRowsPerPage, reportMenuIdForRoute, runReport, toGLReport, type ReportFiscalYear, type ReportRegistry, type ReportRun, type ReportView,
} from "@/lib/report-registry";
import { AccountLedgerModal } from "@/app/gl/account-ledger-drilldown";
import { Field, Notice, Pager, actionClass, panel, useGLLanguage, useGLText, useReferences } from "@/app/gl/gl-common";
import { useDimensionRegisters } from "@/app/gl/gl-dimension-select";
import { useGLPrint } from "@/app/gl/gl-print";
import { JournalDrillDownModal, ReportGrid } from "@/app/gl/gl-reports";
import { KeyHint, ReportFilterSummary, ReportFilterToggle, useReportShortcuts } from "@/app/gl/report-filter-bar";
import { ReportCenterPrint, splitConditions } from "./report-center-print";
import { EMPTY_REPORT_LOOKUPS, ReportParamField, type ReportBudgetOption, type ReportLookups } from "./report-param-field";

/** Largest print run the screen asks for (the backend's limit ceiling); a shorter answer than total_rows is refused rather than printed incomplete. */
export const PRINT_LIMIT = 1000;
/** Families whose rows are accounts: their account code opens that account's ledger. */
const ACCOUNT_DRILL_FAMILIES: ReadonlySet<string> = new Set(["trial_balance", "working_paper"]);
const FIELD_KEY = /^[a-z0-9_]+$/;

/** Values after the user edits one parameter; picking a fiscal year also moves the date-range parameters (…_from / …_to) to that year. */
export function applyParamChange(report: Pick<ReportView, "params">, values: Record<string, string>, key: string, value: string, years: readonly ReportFiscalYear[]): Record<string, string> {
  const next = { ...values, [key]: value };
  const edited = report.params.find((param) => param.key === key);
  const year = edited && (edited.type === "year" || edited.lookup === "fiscal_year") ? years.find((item) => item.code === value) : undefined;
  if (!year) return next;
  for (const param of report.params) {
    if (param.type !== "date") continue;
    if (/(^|_)from$/.test(param.key)) next[param.key] = year.startdate;
    else if (/(^|_)to$/.test(param.key)) next[param.key] = year.enddate;
  }
  return next;
}

/** True when the account code of a row may open its ledger: account-based family, a known year to open it for, and the account exists. */
export function canDrillAccount(report: Pick<ReportView, "family">, year: string): boolean {
  return ACCOUNT_DRILL_FAMILIES.has(report.family) && year !== "";
}

function isAbort(error: unknown): boolean {
  return typeof error === "object" && error !== null && (error as { name?: string }).name === "AbortError";
}

function budgetOptions(items: GLBudget[]): ReportBudgetOption[] {
  return items.map((item) => ({ code: item.code, name: item.name, fiscalyear: item.fiscalyear }));
}

type Applied = { values: Record<string, string>; sort: string };

/** Content of one registry report (route -> menu_id); rendered inside the GL workbench, which supplies the language provider. */
export function ReportCenter({ route }: { route: string }) {
  const tr = useGLText();
  const language = useGLLanguage();
  const refs = useReferences();
  const dims = useDimensionRegisters();
  const reportPrint = useGLPrint();
  const menuId = reportMenuIdForRoute(route);

  const [registry, setRegistry] = useState<ReportRegistry | null>(null);
  const [registryError, setRegistryError] = useState("");
  const [attempt, setAttempt] = useState(0);
  const [values, setValues] = useState<Record<string, string>>({});
  const [sort, setSort] = useState("");
  const [applied, setApplied] = useState<Applied | null>(null);
  const [run, setRun] = useState<ReportRun | null>(null);
  const [page, setPage] = useState(1);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [errorField, setErrorField] = useState("");
  const [budgets, setBudgets] = useState<ReportBudgetOption[]>([]);
  const [drillDocument, setDrillDocument] = useState<{ docno: string; journalId: string } | null>(null);
  const [drillAccount, setDrillAccount] = useState<GLAccount | null>(null);
  const [filterCollapsed, setFilterCollapsed] = useState(false);
  const formRef = useRef<HTMLFormElement>(null);
  const requests = useRef<AbortController | null>(null);

  // kept in a ref so a dictionary that finishes loading does not restart the registry request
  const genericError = tr("report_center_error_generic", "ไม่สามารถทำรายการรายงานได้ กรุณาลองใหม่");
  const genericErrorRef = useRef(genericError);
  useEffect(() => { genericErrorRef.current = genericError; }, [genericError]);

  const report = registry?.reports.find((item) => item.menu_id === menuId);
  const usesBudgets = !!report?.params.some((param) => param.lookup === "budget");
  const exportsCsv = !!report?.exports.includes("csv");

  useEffect(() => {
    const controller = new AbortController();
    fetchReportRegistry(language, genericErrorRef.current, controller.signal)
      .then(setRegistry)
      .catch((reason: unknown) => { if (!isAbort(reason)) setRegistryError(reason instanceof Error && reason.message ? reason.message : genericErrorRef.current); });
    return () => controller.abort();
  }, [language, attempt]);

  useEffect(() => () => requests.current?.abort(), []);

  // starting values: registry defaults, the current fiscal year and its first / last day; years load after the screen opens, so only blanks are filled
  useEffect(() => {
    if (!report) return;
    setValues((current) => fillBlankValues(current, defaultParamValues(report, refs.years)));
    setSort((current) => current || report.default_sort || report.sorts[0]?.value || "");
  }, [report, refs.years]);

  useEffect(() => {
    if (!usesBudgets) return;
    let active = true;
    glRequest<GLPage<GLBudget>>(`budgets?${new URLSearchParams({ q: "", page: "1", limit: "500" })}`).then((result) => { if (active) setBudgets(budgetOptions(result.items ?? [])); }).catch(() => { if (active) setBudgets([]); });
    return () => { active = false; };
  }, [usesBudgets]);

  // the backend names the offending parameter (REPORT_PARAM_*): move the cursor there
  useEffect(() => {
    if (!errorField || !FIELD_KEY.test(errorField)) return;
    formRef.current?.querySelector<HTMLElement>(`[data-field="${errorField}"]`)?.querySelector<HTMLElement>("input:not([disabled]), button:not([disabled]), select, textarea")?.focus();
  }, [errorField, error]);

  // the book filter is picked like departments / projects: the company's journal books shaped as register entries (reports read closed books too)
  const bookRegister = useMemo<GLDimension[]>(() => (refs.books ?? []).map((book) => ({ ...book, name: journalBookName(book, book.code, language), isactive: true })), [refs.books, language]);
  const lookups = useMemo<ReportLookups>(() => ({ ...EMPTY_REPORT_LOOKUPS, accounts: refs.accounts, years: refs.years, books: bookRegister, departments: dims.departments, projects: dims.projects, budgets }), [refs.accounts, refs.years, bookRegister, dims.departments, dims.projects, budgets]);
  const glReport = useMemo(() => (run ? toGLReport(run) : null), [run]);
  const formYear = report?.params.find((param) => param.type === "year")?.key ?? "";

  function nextSignal(): AbortSignal {
    requests.current?.abort();
    requests.current = new AbortController();
    return requests.current.signal;
  }
  function clearError() {
    setError("");
    setErrorField("");
  }
  function showError(reason: unknown) {
    if (isAbort(reason)) return;
    if (reason instanceof ReportError) {
      setError(reason.message);
      setErrorField(reason.field);
      if (reason.field) setFilterCollapsed(false);
    } else {
      setError(reason instanceof Error && reason.message ? reason.message : genericErrorRef.current);
    }
  }
  function change(key: string, value: string) {
    if (!report) return;
    if (errorField === key) setErrorField("");
    setValues((current) => applyParamChange(report, current, key, value, refs.years));
  }

  async function load(nextPage = 1, source: Applied = { values, sort }) {
    if (!report || busy) return;
    const signal = nextSignal();
    setBusy(true);
    clearError();
    try {
      const result = await runReport(buildRunRequest(report, source.values, source.sort, nextPage), language, genericErrorRef.current, signal);
      if (signal.aborted) return;
      setRun(result);
      setApplied({ values: { ...source.values }, sort: source.sort });
      setPage(nextPage);
    } catch (reason) {
      showError(reason);
    } finally {
      if (!signal.aborted) setBusy(false);
    }
  }

  // without the right the button still answers (Champ BaseReportView.cpp:736-745, 2406-2415), so a role missing it is told how to get it
  function refused(allowed: boolean | undefined, key: string, fallback: string): boolean {
    if (allowed) return false;
    showError(new Error(tr(key, fallback)));
    return true;
  }

  async function exportCsv() {
    if (!report || !applied || busy) return;
    if (refused(registry?.can_export, "report_error_export_forbidden", "คุณไม่มีสิทธิ์ส่งออกรายงาน กรุณาให้ผู้ดูแลระบบเปิดสิทธิ์ (ส่งออกข้อมูลของรายงานได้) ที่เมนูกำหนดสิทธิ์การใช้งาน หัวข้อสิทธิ์เฉพาะ")) return;
    const signal = nextSignal();
    setBusy(true);
    clearError();
    try {
      // the export covers the whole result: no page / limit
      const file = await exportReportCsv({ ...buildRunRequest(report, applied.values, applied.sort, 1), page: undefined, limit: undefined }, language, genericErrorRef.current, signal);
      downloadBlob(file.blob, file.filename);
    } catch (reason) {
      showError(reason);
    } finally {
      if (!signal.aborted) setBusy(false);
    }
  }

  async function printReport() {
    if (!report || !applied || busy) return;
    if (refused(registry?.can_print, "report_error_print_forbidden", "คุณไม่มีสิทธิ์พิมพ์รายงาน กรุณาให้ผู้ดูแลระบบเปิดสิทธิ์ (พิมพ์รายงานได้) ที่เมนูกำหนดสิทธิ์การใช้งาน หัวข้อสิทธิ์เฉพาะ")) return;
    // print mode returns every row (backend cap 100,000); refuse before fetching when the screen run already shows too many
    const tooMany = () => new Error(tr("gl_print_too_many", "รายการเกิน {0} รายการ กรุณาจำกัดช่วงวันที่ หรือเลือกสมุดรายวัน/บัญชีก่อนพิมพ์").replace("{0}", PRINT_LIMIT.toLocaleString("en-US")));
    if (run && run.total_rows > PRINT_LIMIT) { showError(tooMany()); return; }
    const signal = nextSignal();
    setBusy(true);
    clearError();
    try {
      const complete = await runReport(buildRunRequest(report, applied.values, applied.sort, 1, { print: true, limit: PRINT_LIMIT, rows_per_page: printRowsPerPage(report) }), language, genericErrorRef.current, signal);
      if (signal.aborted) return;
      if (complete.total_rows > PRINT_LIMIT || complete.rows.length < complete.total_rows) throw tooMany();
      reportPrint.print(<ReportCenterPrint run={complete} report={report} books={refs.books ?? []} tr={tr} language={language} />, report.landscape ? "landscape" : "portrait");
    } catch (reason) {
      showError(reason);
    } finally {
      if (!signal.aborted) setBusy(false);
    }
  }

  const modalOpen = drillDocument !== null || drillAccount !== null;
  const resultReady = !modalOpen && !!run && !!applied && !busy;
  useReportShortcuts({
    print: resultReady ? () => void printReport() : null,
    exportCsv: resultReady && exportsCsv ? () => void exportCsv() : null,
    refresh: !modalOpen && !busy && report?.status === "ready" ? () => void load() : null,
  });

  if (!menuId || (registry && !report)) return <section className={`${panel} flex flex-col gap-3`}><Notice error text={tr("report_center_not_available", "ไม่พบรายงานนี้ในระบบ กรุณาเลือกรายงานจากเมนูอีกครั้ง")} /></section>;
  if (registryError) {
    return <section className={`${panel} flex flex-col items-start gap-3`}>
      <Notice error text={registryError} />
      <Button type="button" variant="outline" className={actionClass} onClick={() => { setRegistryError(""); setAttempt((value) => value + 1); }}><RefreshCw />{tr("common_retry", "ลองใหม่")}</Button>
    </section>;
  }
  if (!registry || !report) return <section className={`${panel} flex flex-col gap-3`}><div role="status" className="rounded-xl border border-dashed border-border p-6 text-center text-muted-foreground">{tr("report_center_loading", "กำลังโหลดรายงาน…")}</div></section>;
  if (report.status === "planned") {
    return <section className={`${panel} flex flex-col gap-2`}>
      <h2 className="text-lg font-semibold">{tr("menu_planned_workflow", "เมนูในแผนพัฒนา")}</h2>
      <p className="leading-relaxed text-muted-foreground">{tr("menu_planned_description", "หน้าจอนี้ยังอยู่ระหว่างเตรียมพัฒนา จึงยังบันทึกหรือประมวลผลข้อมูลไม่ได้ เลือกใช้งานเมนูอื่นจากแถบเมนูได้ตามปกติ")}</p>
    </section>;
  }

  const fieldParams = report.params.filter((param) => param.type !== "bool");
  const boolParams = report.params.filter((param) => param.type === "bool");
  const drillYear = applied ? drillFiscalYear(report, applied.values, refs.years) : "";
  const accountDrill = canDrillAccount(report, drillYear);
  const fiscalYear = formYear ? values[formYear] ?? "" : "";
  const runConditions = run ? splitConditions(run.header.conditions, language) : null;

  return <section className={`${panel} flex flex-col flex-1 min-h-0 gap-3`}>
    <div className="shrink-0 flex items-center justify-between gap-3">
      <div className="flex items-center gap-2.5 min-w-0">
        <h2 className="text-lg font-semibold truncate">{report.name}</h2>
        {run && <ReportFilterToggle collapsed={filterCollapsed} onToggle={() => setFilterCollapsed(!filterCollapsed)} tr={tr} />}
      </div>
    </div>
    <Notice error text={error || refs.error || dims.error} />
    {run && filterCollapsed ? (
      <ReportFilterSummary
        summary={runConditions ? [...runConditions.description, ...runConditions.conditions].join(" · ") : ""}
        busy={busy}
        onExpand={() => setFilterCollapsed(false)}
        onRun={() => void load()}
        onExport={exportsCsv ? () => void exportCsv() : undefined}
        onPrint={() => void printReport()}
        tr={tr}
      />
    ) : <form ref={formRef} className="shrink-0 grid gap-3" aria-busy={busy} onSubmit={(event) => { event.preventDefault(); void load(); }}>
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {fieldParams.map((param) => <ReportParamField key={param.key} param={param} value={values[param.key] ?? ""} invalid={errorField === param.key} lookups={lookups} fiscalYear={fiscalYear} onChange={(value) => change(param.key, value)} />)}
        {report.sorts.length > 1 && <Field label={tr("gl_report_order_by", "เรียงตาม")}><Combobox<string> aria-label={tr("gl_report_order_by", "เรียงตาม")} value={sort} options={report.sorts.map((option) => ({ value: option.value, label: option.label }))} onChange={setSort} /></Field>}
      </div>
      {boolParams.length > 0 && <div role="group" aria-label={tr("gl_report_options", "ตัวเลือกรายงาน")} className="flex flex-wrap gap-2">
        {boolParams.map((param) => <ReportParamField key={param.key} param={param} value={values[param.key] ?? ""} invalid={errorField === param.key} lookups={lookups} fiscalYear={fiscalYear} onChange={(value) => change(param.key, value)} />)}
      </div>}
      <div className="flex flex-wrap items-center gap-2">
        <Button type="submit" className={actionClass} title="Enter / Alt+R" disabled={busy}>
          <RefreshCw />
          {busy ? tr("gl_processing", "กำลังประมวลผล…") : tr("gl_show_report", "แสดงรายงาน")}
          <KeyHint keys="Enter" onPrimary />
        </Button>
        {exportsCsv && <Button type="button" variant="outline" className={actionClass} title="Alt+E" disabled={busy || !run} onClick={() => void exportCsv()}>
          <Download />
          {tr("gl_export_table", "ส่งออกตาราง")}
          <KeyHint keys="Alt+E" />
        </Button>}
        <Button type="button" variant="outline" className={actionClass} title="Alt+P" disabled={busy || !run} onClick={() => void printReport()}>
          <Printer />
          {tr("gl_print", "พิมพ์")}
          <KeyHint keys="Alt+P" />
        </Button>
      </div>
    </form>}
    {glReport && run ? <>
      <ReportGrid
        report={glReport}
        books={refs.books ?? []}
        onDrillDocNo={report.drill === "journal" ? (docno, journalId) => setDrillDocument({ docno, journalId }) : undefined}
        onDrillAccount={accountDrill ? (code) => setDrillAccount(refs.accounts.find((account) => account.accountcode === code) ?? null) : undefined}
      />
      <div className="shrink-0">
        <Pager page={page} total={run.total_rows} onPage={(next) => { if (applied) void load(next, applied); }} loading={busy} limit={REPORT_PAGE_SIZE} />
      </div>
    </> : (
      <div className="rounded-xl border border-dashed border-border p-6 text-center text-muted-foreground">
        {tr("gl_select_fy_and_conditions_show_report", "เลือกปีบัญชีและเงื่อนไข แล้วกดแสดงรายงาน")}
      </div>
    )}
    {report.drill === "journal" && <JournalDrillDownModal docno={drillDocument?.docno ?? null} journalId={drillDocument?.journalId ?? null} open={drillDocument !== null} onClose={() => setDrillDocument(null)} books={refs.books ?? []} />}
    {accountDrill && <AccountLedgerModal open={drillAccount !== null} onClose={() => setDrillAccount(null)} account={drillAccount} fiscalYear={drillYear} books={refs.books ?? []} />}
    {reportPrint.portal}
  </section>;
}
