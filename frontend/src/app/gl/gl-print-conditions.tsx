"use client";

import { useState, type ReactNode } from "react";
import { Printer } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Combobox } from "@/components/ui/combobox";
import { accountTypeLabels, fillText, normalizeDimensionCodes, reportCodeList, type GLAccount, type GLJournalBook, type GLPage, type GLReport, type GLResource, type GLTextFn } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { Field, Notice, actionClass, control, useGLLanguage, useGLText } from "./gl-common";
import { DimensionCodeMultiSelect, useDimensionRegisters } from "./gl-dimension-select";
import { TemplateModal as Modal } from "./gl-journal-templates";
import { GLBudgetsPrint, GLChartOfAccountsPrint, GLJournalBooksPrint, budgetPrintColumnCount, printCompanyName, printOrientation, type ChartNameMode, type useGLPrint } from "./gl-print";
import { emptyReportFilters, fetchReport, type ReportFilters } from "./gl-reports";

// "เงื่อนไขการพิมพ์" ของรายงานข้อมูลหลัก 3 รายการ ที่ Champ ถามก่อนพิมพ์ (docs/kms/champ/09-gl.md §5.1-5.2):
// 5520 ผังบัญชี (ช่วงรหัสบัญชี · ช่วงหมวด · แสดงชื่อบัญชี) · 5531 สมุดรายวัน (ช่วงรหัสสมุด) · 5525 งบประมาณ (ช่วงรหัสงบ · ช่วงบัญชี · สาขา แผนก โครงการ)
// ช่วงรหัสแบบ Champ: ใส่ช่อง "จาก" อย่างเดียว = เฉพาะรหัสนั้น, "จาก"+"ถึง" = ช่วง (GLRepChartOfAccountDlg.cpp:152-166, GLRepBudgetDlg.cpp:244-251)

type Print = ReturnType<typeof useGLPrint>["print"];

const PRINT_PAGE_SIZE = 500;
const PRINT_MAX_PAGES = 20;
const BUDGET_PRINT_PAGE_ROWS = 1000;

/** Rows of a master in print order, filtered and ordered by the backend list endpoint (codefrom/codeto); purpose=print lets the backend refuse without report:print. `request` is injectable for tests; `complete` is false when the backend reported more rows than were read. */
export async function codeRangePrintRows<T>(resource: GLResource, from: string, to: string, request: <R>(path: string) => Promise<R> = glRequest): Promise<{ rows: T[]; complete: boolean }> {
  const codefrom = from.trim(), codeto = codefrom ? to.trim() : "";
  const rows: T[] = [];
  let total = 0;
  for (let page = 1; page <= PRINT_MAX_PAGES; page += 1) {
    const query = new URLSearchParams({ page: String(page), limit: String(PRINT_PAGE_SIZE), ...(codefrom ? { codefrom } : {}), ...(codeto ? { codeto } : {}), purpose: "print" });
    const result = await request<GLPage<T>>(`${resource}?${query}`);
    const items = result.items ?? [];
    total = result.total;
    rows.push(...items);
    if (rows.length >= total || items.length === 0) break;
  }
  return { rows, complete: rows.length >= total };
}

/** Champ checks สิทธิ์เฉพาะ on click before it prints / exports (BaseReportView.cpp:733-746, 2402-2416). Data already on screen is not read again: one purpose-marked read of the register lets the backend refuse with its own message (thrown). */
export async function ensureReportRight(resource: GLResource, purpose: "print" | "export", request: <R>(path: string) => Promise<R> = glRequest): Promise<void> {
  await request(`${resource}?${new URLSearchParams({ page: "1", limit: "1", purpose })}`);
}

/** "from" empty = no condition (a lone "to" is ignored, like Champ); "from" alone = equals; both = range. A reversed range matches nothing in Champ, so the dialogs refuse it. */
export function codeRangeReversed(from: string, to: string): boolean {
  const codefrom = from.trim(), codeto = to.trim();
  return Boolean(codefrom && codeto && codefrom > codeto);
}

/** Print-condition line of one code range: nothing without "from", "equals" for one code, a range for from + to. */
export function codeRangeLines(noun: string, from: string, to: string, tr: GLTextFn): string[] {
  const codefrom = from.trim(), codeto = to.trim();
  if (!codefrom) return [];
  return [codeto
    ? fillText(tr("gl_dimension_cond_range", "จาก{0} {1} ถึง {2}"), noun, codefrom, codeto)
    : fillText(tr("gl_dimension_cond_equal", "จาก{0} เท่ากับ {1}"), noun, codefrom)];
}

// ---- 5520 chart of accounts -------------------------------------------------------------------------------------------------------

export type AccountCategory = GLAccount["accounttype"];
export const CHART_CATEGORY_ORDER: readonly AccountCategory[] = ["asset", "liability", "equity", "income", "expense"];
export type ChartPrintConditions = { codefrom: string; codeto: string; categoryfrom: AccountCategory; categoryto: AccountCategory; namemode: ChartNameMode };
/** Champ default: every account, categories asset … expense, both names. */
export const DEFAULT_CHART_PRINT: ChartPrintConditions = { codefrom: "", codeto: "", categoryfrom: "asset", categoryto: "expense", namemode: "both" };

const categoryRank = (category: AccountCategory) => CHART_CATEGORY_ORDER.indexOf(category);

export function chartPrintReversed(c: ChartPrintConditions): boolean {
  return codeRangeReversed(c.codefrom, c.codeto) || categoryRank(c.categoryfrom) > categoryRank(c.categoryto);
}

/** Accounts the print covers. The chart is read in full by the screen (useReferences), so the filter runs here; the hierarchy and totals stay in the print component. */
export function filterChartForPrint(accounts: readonly GLAccount[], c: ChartPrintConditions): GLAccount[] {
  const codefrom = c.codefrom.trim(), codeto = codefrom ? c.codeto.trim() : "";
  const lowest = categoryRank(c.categoryfrom), highest = categoryRank(c.categoryto);
  return accounts.filter((account) => {
    if (account.isdeleted) return false;
    const code = account.accountcode.trim();
    if (codefrom && (codeto ? code < codefrom || code > codeto : code !== codefrom)) return false;
    const rank = categoryRank(account.accounttype);
    return rank >= lowest && rank <= highest;
  });
}

const nameModeLabels: Record<ChartNameMode, [string, string]> = {
  both: ["gl_coa_name_mode_both", "ทั้งชื่อบัญชี 1 และ 2"],
  name1: ["gl_account_name_th", "ชื่อบัญชี 1"],
  name2: ["gl_account_name_en", "ชื่อบัญชี 2"],
};

/** Conditions footer like Champ 5520: the code range (when given), the category range and the name display are always listed. */
export function chartPrintConditionLines(c: ChartPrintConditions, tr: GLTextFn): string[] {
  const category = (value: AccountCategory) => tr(...accountTypeLabels[value]);
  return [
    ...codeRangeLines(tr("gl_account_code", "รหัสบัญชี"), c.codefrom, c.codeto, tr),
    fillText(tr("gl_coa_cond_category", "จากหมวด {0} ถึง {1}"), category(c.categoryfrom), category(c.categoryto)),
    fillText(tr("gl_coa_cond_name", "แสดงชื่อบัญชี: {0}"), tr(...nameModeLabels[c.namemode])),
  ];
}

function RangeFields({ fromLabel, toLabel, from, to, onFrom, onTo, onEnter }: { fromLabel: string; toLabel: string; from: string; to: string; onFrom: (value: string) => void; onTo: (value: string) => void; onEnter: () => void }) {
  const submitOnEnter = (event: { key: string; preventDefault: () => void }) => { if (event.key === "Enter") { event.preventDefault(); onEnter(); } };
  return <>
    <Field label={fromLabel}><input className={control} data-field="codefrom" value={from} onChange={(event) => onFrom(event.target.value)} onKeyDown={submitOnEnter} /></Field>
    <Field label={toLabel}><input className={control} data-field="codeto" disabled={!from.trim()} value={to} onChange={(event) => onTo(event.target.value)} onKeyDown={submitOnEnter} /></Field>
  </>;
}

function PrintDialogShell({ title, onClose, busy = false, onRun, children, error, hint }: { title: string; onClose: () => void; busy?: boolean; onRun: () => void; children: ReactNode; error: string; hint?: string }) {
  const tr = useGLText();
  return (
    <Modal label={title} onClose={onClose}>
      <h3 className="font-semibold">{title}</h3>
      <div className="grid gap-3 sm:grid-cols-2">{children}</div>
      {hint && <p className="text-muted-foreground">{hint}</p>}
      <Notice error text={error} />
      <div className="flex flex-wrap justify-end gap-2">
        <Button type="button" variant="outline" className={actionClass} disabled={busy} onClick={onClose}>{tr("gl_cancel", "ยกเลิก")}</Button>
        <Button type="button" className={actionClass} disabled={busy} onClick={onRun}><Printer className="size-4 mr-1.5" />{tr("gl_print", "พิมพ์")}</Button>
      </div>
    </Modal>
  );
}

export function ChartPrintDialog({ accounts, print, onClose }: { accounts: readonly GLAccount[]; print: Print; onClose: () => void }) {
  const tr = useGLText(), language = useGLLanguage();
  const [c, setC] = useState<ChartPrintConditions>(DEFAULT_CHART_PRINT), [error, setError] = useState(""), [busy, setBusy] = useState(false);
  const patch = (next: Partial<ChartPrintConditions>) => { setC((value) => ({ ...value, ...next })); setError(""); };
  async function run() {
    if (busy) return;
    if (chartPrintReversed(c)) { setError(tr("gl_print_range_reversed", "ค่า “จาก” ต้องไม่มากกว่า “ถึง”")); return; }
    const rows = filterChartForPrint(accounts, c);
    if (!rows.length) { setError(tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")); return; }
    setBusy(true);
    try {
      await ensureReportRight("accounts", "print");
      print(<GLChartOfAccountsPrint accounts={rows} company={printCompanyName()} tr={tr} language={language} conditions={chartPrintConditionLines(c, tr)} nameMode={c.namemode} />, "portrait");
      onClose();
    } catch (cause) {
      setError(cause instanceof Error && cause.message ? cause.message : tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
    } finally {
      setBusy(false);
    }
  }
  const categoryOptions = CHART_CATEGORY_ORDER.map((value) => ({ value, label: tr(...accountTypeLabels[value]) }));
  return (
    <PrintDialogShell title={tr("gl_coa_report_title", "รายงานรายละเอียดผังบัญชี")} onClose={onClose} busy={busy} onRun={() => void run()} error={error} hint={tr("gl_dimension_print_hint", "เว้นว่างทั้งสองช่อง = พิมพ์ทุกรายการ · ใส่เฉพาะช่อง “จาก” = พิมพ์เฉพาะรหัสนั้น")}>
      <RangeFields fromLabel={tr("gl_print_from_account_code", "จากรหัสบัญชี")} toLabel={tr("gl_print_to_account_code", "ถึงรหัสบัญชี")} from={c.codefrom} to={c.codeto} onFrom={(codefrom) => patch({ codefrom })} onTo={(codeto) => patch({ codeto })} onEnter={() => void run()} />
      <Field label={tr("gl_print_from_category", "จากหมวด")}><Combobox<AccountCategory> aria-label={tr("gl_print_from_category", "จากหมวด")} value={c.categoryfrom} options={categoryOptions} onChange={(categoryfrom) => patch({ categoryfrom })} /></Field>
      <Field label={tr("gl_print_to_category", "ถึงหมวด")}><Combobox<AccountCategory> aria-label={tr("gl_print_to_category", "ถึงหมวด")} value={c.categoryto} options={categoryOptions} onChange={(categoryto) => patch({ categoryto })} /></Field>
      <Field label={tr("gl_print_name_display", "แสดงชื่อบัญชี")}>
        <Combobox<ChartNameMode> aria-label={tr("gl_print_name_display", "แสดงชื่อบัญชี")} value={c.namemode} options={(Object.keys(nameModeLabels) as ChartNameMode[]).map((value) => ({ value, label: tr(...nameModeLabels[value]) }))} onChange={(namemode) => patch({ namemode })} />
      </Field>
    </PrintDialogShell>
  );
}

// ---- 5531 journal books -----------------------------------------------------------------------------------------------------------

export function BookPrintDialog({ print, onClose }: { print: Print; onClose: () => void }) {
  const tr = useGLText(), language = useGLLanguage();
  const [from, setFrom] = useState(""), [to, setTo] = useState(""), [busy, setBusy] = useState(false), [error, setError] = useState("");
  const noun = tr("gl_journal", "สมุดรายวัน");
  async function run() {
    if (busy) return;
    if (codeRangeReversed(from, to)) { setError(tr("gl_print_range_reversed", "ค่า “จาก” ต้องไม่มากกว่า “ถึง”")); return; }
    setBusy(true);
    setError("");
    try {
      const { rows, complete } = await codeRangePrintRows<GLJournalBook>("journal-books", from, to);
      if (!complete) setError(tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
      else if (!rows.length) setError(tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก"));
      else {
        print(<GLJournalBooksPrint books={rows} conditions={codeRangeLines(noun, from, to, tr)} company={printCompanyName()} tr={tr} language={language} />, "portrait");
        onClose();
      }
    } catch (cause) {
      setError(cause instanceof Error && cause.message ? cause.message : tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
    } finally {
      setBusy(false);
    }
  }
  return (
    <PrintDialogShell title={tr("gl_book_report_title", "รายงานรายละเอียดสมุดบัญชี")} onClose={onClose} busy={busy} onRun={() => void run()} error={error} hint={tr("gl_dimension_print_hint", "เว้นว่างทั้งสองช่อง = พิมพ์ทุกรายการ · ใส่เฉพาะช่อง “จาก” = พิมพ์เฉพาะรหัสนั้น")}>
      <RangeFields fromLabel={fillText(tr("gl_dimension_filter_from", "จาก{0}"), noun)} toLabel={fillText(tr("gl_dimension_filter_to", "ถึง{0}"), noun)} from={from} to={to}
        onFrom={(value) => { setFrom(value); setError(""); }} onTo={(value) => { setTo(value); setError(""); }} onEnter={() => void run()} />
    </PrintDialogShell>
  );
}

// ---- 5525 budget detail -----------------------------------------------------------------------------------------------------------

export type BudgetPrintConditions = { budgetfrom: string; budgetto: string; accountfrom: string; accountto: string; branchcode: string; departments: string[]; projects: string[] };
export const DEFAULT_BUDGET_PRINT: BudgetPrintConditions = { budgetfrom: "", budgetto: "", accountfrom: "", accountto: "", branchcode: "", departments: [], projects: [] };

export function budgetPrintReversed(c: BudgetPrintConditions): boolean {
  return codeRangeReversed(c.budgetfrom, c.budgetto) || codeRangeReversed(c.accountfrom, c.accountto);
}

/** Filters of the backend report "budgetdetail". Champ's "from" alone means that exact code: for accounts the backend takes from=to, for budgets "to" stays empty (exact). */
export function budgetPrintFilters(c: BudgetPrintConditions): ReportFilters {
  const budgetcode = c.budgetfrom.trim(), accountfrom = c.accountfrom.trim();
  return {
    ...emptyReportFilters,
    budgetcode, budgetcodeto: budgetcode ? c.budgetto.trim() : "",
    accountfrom, accountto: accountfrom ? c.accountto.trim() || accountfrom : "",
    branchcode: c.branchcode.trim(),
    departmentcode: normalizeDimensionCodes(c.departments).join(","),
    projectcode: normalizeDimensionCodes(c.projects).join(","),
  };
}

export function budgetPrintConditionLines(c: BudgetPrintConditions, tr: GLTextFn): string[] {
  const lines = [...codeRangeLines(tr("gl_print_budget_code_noun", "รหัสงบประมาณ"), c.budgetfrom, c.budgetto, tr), ...codeRangeLines(tr("gl_account_code", "รหัสบัญชี"), c.accountfrom, c.accountto, tr)];
  const branch = c.branchcode.trim(), departments = normalizeDimensionCodes(c.departments), projects = normalizeDimensionCodes(c.projects);
  if (branch) lines.push(`${tr("gl_branch_code", "รหัสสาขา")} ${branch}`);
  if (departments.length) lines.push(`${tr("gl_department_code", "รหัสแผนก")} ${departments.join(", ")}`);
  if (projects.length) lines.push(`${tr("gl_project_code", "รหัสโครงการ")} ${projects.join(", ")}`);
  return lines;
}

/** Every row of the report at one snapshot (pages of BUDGET_PRINT_PAGE_ROWS); a change between pages aborts instead of printing a mixed report. */
export async function budgetPrintReport(filters: ReportFilters, changedMessage: string, request: typeof fetchReport = fetchReport): Promise<Pick<GLReport, "rows" | "totals">> {
  const first = await request("budgetdetail", filters, 1, BUDGET_PRINT_PAGE_ROWS, undefined, false, "print");
  const rows = [...(first.rows ?? [])];
  for (let next = 2; rows.length < first.totalrows; next++) {
    const chunk = await request("budgetdetail", filters, next, BUDGET_PRINT_PAGE_ROWS, first.sequence, false, "print");
    if (chunk.totalrows !== first.totalrows || chunk.sequence !== first.sequence || !chunk.rows?.length) throw new Error(changedMessage);
    rows.push(...chunk.rows);
  }
  return { rows, totals: first.totals };
}

export function BudgetPrintDialog({ print, onClose }: { print: Print; onClose: () => void }) {
  const tr = useGLText(), language = useGLLanguage();
  const dims = useDimensionRegisters();
  const [c, setC] = useState<BudgetPrintConditions>(DEFAULT_BUDGET_PRINT), [busy, setBusy] = useState(false), [error, setError] = useState("");
  const patch = (next: Partial<BudgetPrintConditions>) => { setC((value) => ({ ...value, ...next })); setError(""); };
  async function run() {
    if (busy) return;
    if (budgetPrintReversed(c)) { setError(tr("gl_print_range_reversed", "ค่า “จาก” ต้องไม่มากกว่า “ถึง”")); return; }
    setBusy(true);
    setError("");
    try {
      const report = await budgetPrintReport(budgetPrintFilters(c), tr("gl_data_changed_during_export_retry", "ข้อมูลเปลี่ยนแปลงระหว่างส่งออก กรุณาลองใหม่"));
      if (!report.rows.length) setError(tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก"));
      else {
        print(<GLBudgetsPrint report={report} conditions={budgetPrintConditionLines(c, tr)} company={printCompanyName()} tr={tr} language={language} />, printOrientation(budgetPrintColumnCount(report.rows)));
        onClose();
      }
    } catch (cause) {
      setError(cause instanceof Error && cause.message ? cause.message : tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
    } finally {
      setBusy(false);
    }
  }
  const submitOnEnter = (event: { key: string; preventDefault: () => void }) => { if (event.key === "Enter") { event.preventDefault(); void run(); } };
  return (
    <PrintDialogShell title={tr("gl_budget_report_title", "รายงานรายละเอียดงบประมาณ")} onClose={onClose} busy={busy} onRun={() => void run()} error={error} hint={tr("gl_dimension_print_hint", "เว้นว่างทั้งสองช่อง = พิมพ์ทุกรายการ · ใส่เฉพาะช่อง “จาก” = พิมพ์เฉพาะรหัสนั้น")}>
      <RangeFields fromLabel={tr("gl_print_from_budget_code", "จากรหัสงบประมาณ")} toLabel={tr("gl_print_to_budget_code", "ถึงรหัสงบประมาณ")} from={c.budgetfrom} to={c.budgetto} onFrom={(budgetfrom) => patch({ budgetfrom })} onTo={(budgetto) => patch({ budgetto })} onEnter={() => void run()} />
      <Field label={tr("gl_print_from_account_code", "จากรหัสบัญชี")}><input className={control} data-field="accountfrom" value={c.accountfrom} onChange={(event) => patch({ accountfrom: event.target.value })} onKeyDown={submitOnEnter} /></Field>
      <Field label={tr("gl_print_to_account_code", "ถึงรหัสบัญชี")}><input className={control} data-field="accountto" disabled={!c.accountfrom.trim()} value={c.accountto} onChange={(event) => patch({ accountto: event.target.value })} onKeyDown={submitOnEnter} /></Field>
      <Field label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} data-field="branchcode" value={c.branchcode} placeholder={tr("gl_all_branches", "ทุกสาขา")} onChange={(event) => patch({ branchcode: event.target.value })} onKeyDown={submitOnEnter} /></Field>
      <div className="sm:col-span-2 grid gap-3 sm:grid-cols-2">
        <DimensionCodeMultiSelect report resource="departments" value={reportCodeList(c.departments.join(","))} register={dims.departments} disabled={busy} onChange={(departments) => patch({ departments })} />
        <DimensionCodeMultiSelect report resource="projects" value={reportCodeList(c.projects.join(","))} register={dims.projects} disabled={busy} onChange={(projects) => patch({ projects })} />
      </div>
    </PrintDialogShell>
  );
}
