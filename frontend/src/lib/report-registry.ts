/**
 * Central report registry client (backend/internal/report, BFF /api/report/*).
 * Every report menu opens ONE screen (app/report/report-center.tsx); which parameters it shows and how a report
 * runs comes from the registry the backend returns, so this file holds the contract types, the route -> menu id
 * table and the fetch helpers — no report-specific rule lives in the browser.
 */
import { apiFetch } from "./client-auth-session";
import type { GLReport } from "./general-ledger";
import { normalizeLanguage, type LanguageCode } from "./i18n";

export const REPORT_MODULE_GL = "gl";

/**
 * Menu route -> menu id of every item the central report screen opens (wave 1, GL). Keep in step with the variants in
 * backend/internal/report/defs_gl.go (golden list: backend/internal/report/testdata/registry.json, checked by
 * report-registry.test.ts). `financial-statement-designer` is planned in the registry but its route keeps opening the
 * statement designer, so it is deliberately absent here.
 */
export const REPORT_ROUTES: Readonly<Record<string, string>> = {
  "/report/gljournal": "gl-daily-report",
  "/report/gljournal-summary": "gl-daily-summary-report",
  "/report/ledger": "general-ledger",
  "/report/ledger-summary": "general-ledger-summary",
  "/report/trialbalance": "trial-balance",
  "/report/trialbalance-2": "trial-balance-2",
  "/gl/workingpaper": "working-paper",
  "/report/budgetcomparison": "budget-comparison-report",
  "/report/budget-detail": "gl-budget-detail",
  "/report/journal-books": "gl-journal-books-report",
  "/report/departments": "gl-department-report",
  "/report/projects": "gl-project-report",
  "/report/chart-of-accounts": "gl-chart-of-accounts-report",
};

/** Menu id the central screen opens for a menu route (query string ignored); undefined = not a registry report. */
export function reportMenuIdForRoute(route: string): string | undefined {
  const clean = route.split("?")[0];
  return Object.prototype.hasOwnProperty.call(REPORT_ROUTES, clean) ? REPORT_ROUTES[clean] : undefined;
}
export function isReportRegistryRoute(route: string): boolean {
  return reportMenuIdForRoute(route) !== undefined;
}

// ---- contract (GET /report/v1/registry, POST /report/v1/run) ----

export type ReportParamType = "date" | "code" | "codes" | "enum" | "bool" | "year";
export type ReportLookup = "account" | "branch" | "book" | "department" | "project" | "budget" | "fiscal_year" | "";
export type ReportOption = { value: string; label: string };
export type ReportParam = {
  key: string;
  type: ReportParamType;
  label: string;
  required: boolean;
  default: string;
  lookup: ReportLookup;
  options: ReportOption[];
};
export type ReportView = {
  menu_id: string;
  champ_menu_id: string;
  code: string;
  module: string;
  family: string;
  status: "ready" | "planned";
  name: string;
  route: string;
  landscape: boolean;
  exports: string[];
  drill: "journal" | "";
  params: ReportParam[];
  sorts: ReportOption[];
  default_sort: string;
};
export type ReportRegistry = { can_print: boolean; can_export: boolean; reports: ReportView[] };

export type ReportColumnType = "text" | "date" | "amount" | "qty" | "int";
export type ReportRunColumn = { key: string; label: string; type: ReportColumnType; scale: number; group: string };
/** Rows keep the engine's raw string shape; `_kind`, `_label`, `_bold` are the registry's own row markers. */
export type ReportRow = Record<string, string>;
export type ReportCondition = { key: string; label: string; value: string; type: string };
export type ReportPageTotal = { print_page: number; first_row: number; last_row: number; sums: Record<string, string>; carried: Record<string, string> };
export type ReportRun = {
  menu_id: string;
  code: string;
  columns: ReportRunColumn[];
  rows: ReportRow[];
  totals: Record<string, string>;
  totals_labels: Record<string, string>;
  footer: ReportRow[];
  page_totals: ReportPageTotal[];
  total_rows: number;
  page: number;
  limit: number;
  as_of: string;
  header: { company_name: string; branch_name: string; title: string; conditions: ReportCondition[]; printed_by: string; printed_at: string };
  warnings: string[];
};
export type ReportRunRequest = {
  menu_id: string;
  params: Record<string, string>;
  sort?: string;
  page?: number;
  limit?: number;
  print?: boolean;
  rows_per_page?: number;
};

/** Rows per printed page: kept below what the print CSS fits so one server page = one paper page (portrait / landscape). */
export const PRINT_ROWS_PORTRAIT = 28;
export const PRINT_ROWS_LANDSCAPE = 18;
export const REPORT_PAGE_SIZE = 50;
export function printRowsPerPage(report: Pick<ReportView, "landscape">): number {
  return report.landscape ? PRINT_ROWS_LANDSCAPE : PRINT_ROWS_PORTRAIT;
}

// ---- errors ----

/** A failed registry/run/export call: `message` is already translated by the backend, `field` names the parameter to highlight. */
export class ReportError extends Error {
  readonly errorcode: string;
  readonly field: string;
  readonly status: number;
  constructor(message: string, errorcode = "", field = "", status = 0) {
    super(message);
    this.name = "ReportError";
    this.errorcode = errorcode;
    this.field = field;
    this.status = status;
  }
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
function text(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

/** Reads the backend error body ({success:false, errorcode, message, message_th, field}); `fallback` covers an empty message. */
export function reportErrorFrom(payload: unknown, status: number, fallback: string): ReportError {
  const body = isObject(payload) ? payload : {};
  const nested = isObject(body.error) ? body.error : {};
  const message = text(body.message) || text(body.message_th) || text(nested.message) || fallback;
  return new ReportError(message, text(body.errorcode) || text(body.code), text(body.field), status);
}

// ---- requests ----

function languageOf(language: LanguageCode | string): LanguageCode {
  return normalizeLanguage(language);
}

async function relay(path: string, init: RequestInit, language: LanguageCode | string): Promise<Response> {
  const lang = languageOf(language);
  return apiFetch(path, {
    ...init,
    headers: { "Content-Type": "application/json", "Accept-Language": lang, ...init.headers },
  });
}

async function readBody(response: Response): Promise<unknown> {
  return response.json().catch(() => ({}));
}

function failed(payload: unknown, response: Response): boolean {
  return !response.ok || (isObject(payload) && payload.success === false);
}

export async function fetchReportRegistry(language: LanguageCode | string, fallback: string, signal?: AbortSignal, module = REPORT_MODULE_GL): Promise<ReportRegistry> {
  const lang = languageOf(language);
  const response = await relay(`/api/report/registry?${new URLSearchParams({ module, lang })}`, { method: "GET", signal }, lang);
  const payload = await readBody(response);
  if (failed(payload, response)) throw reportErrorFrom(payload, response.status, fallback);
  const data = isObject(payload) && isObject(payload.data) ? payload.data : {};
  return {
    can_print: data.can_print === true,
    can_export: data.can_export === true,
    reports: Array.isArray(data.reports) ? (data.reports as ReportView[]).map(normalizeView) : [],
  };
}

/** The backend may omit empty lists (JSON null): the screen always iterates arrays. */
function normalizeView(view: ReportView): ReportView {
  return {
    ...view,
    exports: view.exports ?? [],
    params: (view.params ?? []).map((param) => ({ ...param, options: param.options ?? [], default: param.default ?? "", lookup: param.lookup ?? "" })),
    sorts: view.sorts ?? [],
    default_sort: view.default_sort ?? "",
    drill: view.drill ?? "",
  };
}

export async function runReport(request: ReportRunRequest, language: LanguageCode | string, fallback: string, signal?: AbortSignal): Promise<ReportRun> {
  const response = await relay("/api/report/run", { method: "POST", body: JSON.stringify(request), signal }, language);
  const payload = await readBody(response);
  if (failed(payload, response)) throw reportErrorFrom(payload, response.status, fallback);
  const data = isObject(payload) && isObject(payload.data) ? (payload.data as Partial<ReportRun>) : {};
  return normalizeRun(data);
}

export function normalizeRun(data: Partial<ReportRun>): ReportRun {
  const header = data.header ?? { company_name: "", branch_name: "", title: "", conditions: [], printed_by: "", printed_at: "" };
  return {
    menu_id: data.menu_id ?? "",
    code: data.code ?? "",
    columns: data.columns ?? [],
    rows: data.rows ?? [],
    totals: data.totals ?? {},
    totals_labels: data.totals_labels ?? {},
    footer: data.footer ?? [],
    page_totals: data.page_totals ?? [],
    total_rows: data.total_rows ?? 0,
    page: data.page ?? 1,
    limit: data.limit ?? 0,
    as_of: data.as_of ?? "",
    header: { ...header, conditions: header.conditions ?? [] },
    warnings: data.warnings ?? [],
  };
}

/** CSV export (UTF-8 BOM, streamed by the backend): returns the file, or throws the backend's translated error. */
export async function exportReportCsv(request: ReportRunRequest, language: LanguageCode | string, fallback: string, signal?: AbortSignal): Promise<{ blob: Blob; filename: string }> {
  const response = await relay("/api/report/export", { method: "POST", body: JSON.stringify({ ...request, format: "csv" }), signal }, language);
  const type = response.headers.get("content-type") ?? "";
  if (!response.ok || !type.includes("text/csv")) throw reportErrorFrom(await readBody(response), response.status, fallback);
  return { blob: await response.blob(), filename: filenameFromDisposition(response.headers.get("content-disposition")) || `${request.menu_id}.csv` };
}

/** Content-Disposition filename (RFC 5987 `filename*` first, then plain `filename`); "" when absent. */
export function filenameFromDisposition(header: string | null): string {
  if (!header) return "";
  const star = /filename\*\s*=\s*(?:UTF-8|utf-8)''([^;]+)/.exec(header);
  if (star) {
    try { return decodeURIComponent(star[1].trim()); } catch { /* fall through to the plain name */ }
  }
  const plain = /filename\s*=\s*"?([^";]+)"?/.exec(header);
  return plain ? plain[1].trim() : "";
}

export function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

// ---- parameter values ----

export type ReportFiscalYear = { code: string; startdate: string; enddate: string };

export function localDateText(now = new Date()): string {
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
}

/** Fiscal year to start from: the one that contains today, else the latest by start date. */
export function currentFiscalYear(years: readonly ReportFiscalYear[], today = localDateText()): ReportFiscalYear | undefined {
  return years.find((year) => year.startdate <= today && today <= year.enddate)
    ?? [...years].sort((left, right) => right.startdate.localeCompare(left.startdate))[0];
}

/**
 * Starting value of every parameter the variant offers: the registry default; year parameters take the current
 * fiscal year; date_from / date_to-style parameters take that year's first / last day. Nothing is invented for
 * other parameters (an empty value means "all").
 */
export function defaultParamValues(report: Pick<ReportView, "params">, years: readonly ReportFiscalYear[], today = localDateText()): Record<string, string> {
  const year = currentFiscalYear(years, today);
  const values: Record<string, string> = {};
  for (const param of report.params) {
    let value = param.default ?? "";
    if (!value && year) {
      if (param.type === "year") value = year.code;
      else if (param.type === "date") value = /(^|_)from$/.test(param.key) ? year.startdate : /(^|_)to$/.test(param.key) ? year.enddate : "";
    }
    values[param.key] = value;
  }
  return values;
}

/** Fills only the blank values of `current` (years load after the screen opens); typed values are never overwritten. */
export function fillBlankValues(current: Record<string, string>, defaults: Record<string, string>): Record<string, string> {
  const next = { ...current };
  for (const [key, value] of Object.entries(defaults)) if (!next[key]) next[key] = value;
  return next;
}

/** Request params: offered, non-empty values as strings; booleans always travel as "true" / "false". */
export function buildRunParams(report: Pick<ReportView, "params">, values: Record<string, string>): Record<string, string> {
  const params: Record<string, string> = {};
  for (const param of report.params) {
    const value = (values[param.key] ?? "").trim();
    if (param.type === "bool") params[param.key] = value === "true" ? "true" : "false";
    else if (value) params[param.key] = value;
  }
  return params;
}

export function buildRunRequest(report: ReportView, values: Record<string, string>, sort: string, page: number, extra: Partial<ReportRunRequest> = {}): ReportRunRequest {
  return {
    menu_id: report.menu_id,
    params: buildRunParams(report, values),
    ...(report.sorts.length > 1 && sort ? { sort } : {}),
    page,
    limit: REPORT_PAGE_SIZE,
    ...extra,
  };
}

/** Fiscal year to open an account's ledger for: the variant's year parameter, else the year that holds the report's end date. */
export function drillFiscalYear(report: Pick<ReportView, "params">, values: Record<string, string>, years: readonly ReportFiscalYear[]): string {
  const yearParam = report.params.find((param) => param.type === "year");
  const chosen = yearParam ? (values[yearParam.key] ?? "") : "";
  if (chosen) return chosen;
  const end = Object.entries(values).find(([key, value]) => /(^|_)to$/.test(key) && /^\d{4}-\d{2}-\d{2}$/.test(value))?.[1] ?? "";
  return end ? years.find((year) => year.startdate <= end && end <= year.enddate)?.code ?? "" : "";
}

// ---- run -> screen / print shapes ----

/** Totals that are plain counts (master lists, budget lines) arrive as integer strings and are not columns; amounts always carry decimals. */
export function countTotalKeys(run: Pick<ReportRun, "columns" | "totals">): string[] {
  const amountColumns = new Set(run.columns.filter((column) => column.type === "amount" || column.type === "qty").map((column) => column.key));
  return Object.entries(run.totals).filter(([key, value]) => !amountColumns.has(key) && /^-?\d+$/.test(value)).map(([key]) => key);
}

/** Adapts a registry run to the shape ReportGrid draws (the one grid every GL report shares). */
export function toGLReport(run: ReportRun): GLReport {
  return {
    columns: run.columns.map((column) => ({
      key: column.key,
      label: column.label,
      amount: column.type === "amount" || column.type === "qty",
      scale: column.type === "amount" || column.type === "qty" ? column.scale : undefined,
      ...(column.group ? { group: column.group } : {}),
    })),
    rows: run.rows,
    totals: run.totals,
    totalrows: run.total_rows,
    warnings: run.warnings,
    asof: run.as_of,
    sequence: 0,
    footer: run.footer,
    totalLabels: run.totals_labels,
    countTotals: countTotalKeys(run),
  };
}

export type ReportPrintPage = {
  rows: ReportRow[];
  /** Sums of this printed page and the running total through it (server page_totals); empty when the report has none. */
  sums: Record<string, string>;
  carried: Record<string, string>;
  /** Running total through the previous page ("ยอดยกมา"); empty on the first page. */
  brought: Record<string, string>;
};

/**
 * Splits a print run into printed pages by the server's page_totals (1-based inclusive row indexes). A run without
 * page totals is one page of all rows. Indexes outside the rows are clamped, rows missing from every range are kept
 * on the last page so nothing is silently dropped from the printout.
 */
export function printPages(run: Pick<ReportRun, "rows" | "page_totals">): ReportPrintPage[] {
  const totals = run.page_totals;
  if (!totals.length) return [{ rows: run.rows, sums: {}, carried: {}, brought: {} }];
  const pages: ReportPrintPage[] = [];
  let previous: Record<string, string> = {};
  let next = 0;
  for (const total of totals) {
    const start = Math.min(Math.max(total.first_row - 1, next), run.rows.length);
    const end = Math.min(Math.max(total.last_row, start), run.rows.length);
    pages.push({ rows: run.rows.slice(start, end), sums: total.sums ?? {}, carried: total.carried ?? {}, brought: previous });
    previous = total.carried ?? {};
    next = end;
  }
  if (next < run.rows.length) pages[pages.length - 1].rows = pages[pages.length - 1].rows.concat(run.rows.slice(next));
  return pages;
}
