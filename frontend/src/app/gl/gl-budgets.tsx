"use client";

// จอกำหนดงบประมาณรายเดือน (Champ 5500 "กำหนดงบประมาณ") บน API resource "budgets"
// ADR docs/kms/decisions/2026-09-25-gl-monthly-budget.md — เงินเป็นข้อความทศนิยม, ยอดรวมคำนวณด้วย BigInt,
// การแบ่งยอดทั้งปีเป็นงวดทำที่ backend (action "spread") ไม่คำนวณเองในเบราว์เซอร์
import { useEffect, useMemo, useState } from "react";
import { Columns3, Plus, RefreshCw, Save, Search, Trash2, Undo2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { BUDGET_PERIODS, amountString, blankAmountAsZero, budgetPeriodStarts, budgetPeriodsTotal, formatAmount, type GLBudget, type GLBudgetLine, type GLPage } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { AccountSelect, AmountInput, Field, Notice, Pager, SearchInput, SplitWorkbench, UnsavedBadge, YearSelect, actionClass, control, useDebouncedSearch, useDirtyGuard, useGLCommand, useGLLanguage, useGLText, useReferences } from "./gl-common";

type BudgetDraft = Omit<GLBudget, "total">;
const LIST_LIMIT = 30;
const emptyPeriods = () => Array.from({ length: BUDGET_PERIODS }, () => "");
const newLine = (): GLBudgetLine => ({ accountcode: "", periods: emptyPeriods() });
function newBudget(fiscalyear: string): BudgetDraft {
  return { code: "", name: "", fiscalyear, branchcode: "", departmentcode: "", projectcode: "", status: "open", remark: "", lines: [newLine()] };
}
function toDraft(value: GLBudget): BudgetDraft {
  return {
    id: value.id, version: value.version, code: value.code ?? "", name: value.name ?? "", fiscalyear: value.fiscalyear ?? "",
    branchcode: value.branchcode ?? "", departmentcode: value.departmentcode ?? "", projectcode: value.projectcode ?? "",
    status: value.status === "closed" ? "closed" : "open", remark: value.remark ?? "",
    lines: (value.lines ?? []).map((line) => ({ accountcode: line.accountcode ?? "", accountname: line.accountname, periods: Array.from({ length: BUDGET_PERIODS }, (_, index) => line.periods?.[index] ?? "") })),
  };
}
// ยอดที่ยังพิมพ์ไม่ครบรูปแบบไม่ทำให้จอพัง — นับเป็นศูนย์จนกว่าจะเป็นทศนิยมที่ถูกต้อง (backend ตรวจซ้ำตอนบันทึก)
function safeTotal(periods: string[]) { try { return budgetPeriodsTotal(periods); } catch { return 0n; } }
const money = (units: bigint) => formatAmount(amountString(units, 2));
// งวดที่อยู่นอกปีบัญชี (ปีสั้น) ต้องส่งศูนย์ ตาม validateBudget ใน backend
function budgetPayload(draft: BudgetDraft, periodCount: number) {
  return {
    code: draft.code.trim(), name: draft.name.trim(), fiscalyear: draft.fiscalyear, branchcode: draft.branchcode.trim(),
    departmentcode: draft.departmentcode.trim(), projectcode: draft.projectcode.trim(), status: draft.status, remark: draft.remark.trim(),
    lines: (draft.lines ?? []).map((line) => ({ accountcode: line.accountcode, periods: line.periods.map((value, index) => index < periodCount ? blankAmountAsZero(value) : "0") })),
  };
}

export function GLBudgets({ route }: { route: string }) {
  const tr = useGLText();
  const language = useGLLanguage();
  const refs = useReferences();
  const [search, setSearch] = useState(""), [page, setPage] = useState(1), [revision, setRevision] = useState(0);
  const [list, setList] = useState<GLPage<GLBudget>>({ items: [], total: 0, page: 1, limit: LIST_LIMIT, sequence: 0 });
  const [loading, setLoading] = useState(false);
  const [draft, setDraft] = useState<BudgetDraft | null>(null), [original, setOriginal] = useState("");
  const [annual, setAnnual] = useState<string[]>([]);
  const [message, setMessage] = useState(""), [error, setError] = useState("");
  const { busy, execute } = useGLCommand();
  const { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const searchDebounce = useDebouncedSearch({ onSearch: (value) => { setPage(1); setSearch(value); }, debounceMs: 2000 });
  const dirty = draft !== null && JSON.stringify(draft) !== original;
  useDirtyGuard(route, dirty);

  useEffect(() => {
    let active = true;
    setLoading(true);
    glRequest<GLPage<GLBudget>>(`budgets?${new URLSearchParams({ q: search, page: String(page), limit: String(LIST_LIMIT) })}`)
      .then((result) => { if (active) setList({ ...result, items: result.items ?? [] }); })
      .catch((e: Error) => { if (active) setError(e.message); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [search, page, revision]);

  const year = refs.years.find((item) => item.code === draft?.fiscalyear);
  const starts = useMemo(() => budgetPeriodStarts(year), [year]);
  const periodCount = starts.length || BUDGET_PERIODS;
  const monthFormat = useMemo(() => new Intl.DateTimeFormat(language === "th" ? "th-TH" : "en-GB", { month: "short", year: "2-digit" }), [language]);
  const periodLabel = (index: number) => starts[index] ? monthFormat.format(new Date(`${starts[index]}T00:00:00`)) : tr("gl_budget_period_n", "งวด {0}").replace("{0}", String(index + 1));
  const postingAccounts = useMemo(() => refs.accounts.filter((account) => account.allowposting && account.isactive), [refs.accounts]);
  const lines = draft?.lines ?? [];
  const lineTotals = lines.map((line) => safeTotal(line.periods));
  const periodTotals = Array.from({ length: BUDGET_PERIODS }, (_, period) => lines.reduce((sum, line) => sum + safeTotal([line.periods[period] ?? ""]), 0n));
  const grandTotal = lineTotals.reduce((sum, value) => sum + value, 0n);

  function load(next: BudgetDraft) {
    setDraft(next);
    setOriginal(JSON.stringify(next));
    setAnnual((next.lines ?? []).map(() => ""));
  }
  async function discardOk() {
    return !dirty || await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning" });
  }
  async function openBudget(code: string) {
    if (busy || !await discardOk()) return;
    setError(""); setMessage("");
    try { load(toDraft(await glRequest<GLBudget>(`budgets/${encodeURIComponent(code)}`))); }
    catch (e) { setError((e as Error).message); }
  }
  async function openCreate() {
    if (busy || !await discardOk()) return;
    const now = new Date(), today = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
    const current = refs.years.find((item) => item.isactive && !item.closed && item.startdate <= today && today <= item.enddate) ?? refs.years.find((item) => item.isactive && !item.closed);
    setError(""); setMessage("");
    load(newBudget(current?.code ?? ""));
  }
  async function closeEditor() {
    if (!await discardOk()) return;
    setDraft(null); setOriginal(""); setAnnual([]); setError("");
  }
  const setHeader = (patch: Partial<BudgetDraft>) => setDraft((current) => current ? { ...current, ...patch } : current);
  const setLine = (index: number, patch: Partial<GLBudgetLine>) => setDraft((current) => current ? { ...current, lines: (current.lines ?? []).map((line, i) => i === index ? { ...line, ...patch } : line) } : current);
  const setPeriod = (index: number, period: number, value: string) => setDraft((current) => current ? { ...current, lines: (current.lines ?? []).map((line, i) => i === index ? { ...line, periods: line.periods.map((old, p) => p === period ? value : old) } : line) } : current);
  const addLine = () => { setDraft((current) => current ? { ...current, lines: [...(current.lines ?? []), newLine()] } : current); setAnnual((values) => [...values, ""]); };
  const removeLine = (index: number) => { setDraft((current) => current ? { ...current, lines: (current.lines ?? []).filter((_, i) => i !== index) } : current); setAnnual((values) => values.filter((_, i) => i !== index)); };

  function validate(value: BudgetDraft) {
    if (!value.code.trim()) return tr("gl_budget_code_required", "กรุณาระบุรหัสงบประมาณ");
    if (!value.name.trim()) return tr("gl_budget_name_required", "กรุณาระบุชื่องบประมาณ");
    if (!value.fiscalyear) return tr("gl_budget_year_required", "กรุณาเลือกปีบัญชี");
    if (!value.lines?.length) return tr("gl_budget_lines_required", "กรุณาเพิ่มบัญชีในงบประมาณอย่างน้อย 1 บัญชี");
    if (value.lines.some((line) => !line.accountcode)) return tr("gl_budget_account_required", "กรุณาเลือกรหัสบัญชีในทุกบรรทัดของงบประมาณ");
    return "";
  }
  async function spreadLine(index: number) {
    if (!draft || busy) return;
    const total = (annual[index] ?? "").trim();
    if (!draft.fiscalyear) { setError(tr("gl_budget_year_required", "กรุณาเลือกปีบัญชี")); return; }
    if (!total) { setError(tr("gl_budget_annual_required", "กรุณาใส่ยอดทั้งปีก่อนกดแบ่งงวด")); return; }
    try {
      setError("");
      const result = await execute({ resource: "budgets", action: "spread", budget: { ...budgetPayload(draft, periodCount), lines: [{ accountcode: lines[index]?.accountcode ?? "", periods: [], total }] } }) as { lines?: GLBudgetLine[] };
      const periods = result.lines?.[0]?.periods;
      if (periods?.length === BUDGET_PERIODS) setLine(index, { periods: [...periods] });
    } catch (e) { setError((e as Error).message); }
  }
  async function save() {
    if (!draft || busy) return;
    const problem = validate(draft);
    if (problem) { setError(problem); return; }
    const isNew = !draft.id;
    try {
      setError(""); setMessage("");
      const result = await execute({ resource: "budgets", action: isNew ? "create" : "update", id: isNew ? undefined : draft.id, version: draft.version, reason: isNew ? tr("gl_add_row", "เพิ่มรายการ") : tr("gl_edit_data", "แก้ไขข้อมูล"), budget: budgetPayload(draft, periodCount) });
      const saved = toDraft(await glRequest<GLBudget>(`budgets/${encodeURIComponent(result.id || draft.code.trim())}`));
      load(saved);
      setRevision((value) => value + 1);
      setMessage(tr("gl_budget_saved", "บันทึกงบประมาณ {0} เรียบร้อยแล้ว").replace("{0}", saved.code));
    } catch (e) { setError((e as Error).message); }
  }
  async function remove() {
    if (!draft?.id || busy) return;
    if (!await confirm({ title: tr("gl_confirm_delete", "ยืนยันลบรายการ?"), description: `${draft.code} · ${draft.name}`, details: tr("gl_budget_delete_detail", "ลบงบประมาณพร้อมยอดทุกบัญชีทุกงวด รายงานเปรียบเทียบงบประมาณจะไม่แสดงงบนี้อีก"), confirmLabel: tr("gl_delete_item", "ลบรายการ"), tone: "danger" })) return;
    try {
      setError("");
      await execute({ resource: "budgets", action: "delete", id: draft.id, version: draft.version, reason: tr("gl_delete_item", "ลบรายการ") });
      setMessage(tr("gl_delete_success", "ลบ {0} เรียบร้อยแล้ว").replace("{0}", draft.code));
      setDraft(null); setOriginal(""); setAnnual([]);
      setRevision((value) => value + 1);
    } catch (e) { setError((e as Error).message); }
  }
  function revert() {
    if (!original) return;
    load(JSON.parse(original) as BudgetDraft);
  }

  const listPane = <div className="flex flex-col flex-1 min-h-0 gap-2">
    <form className="shrink-0 flex flex-wrap items-center gap-2 border-b border-border/70 pb-3" onSubmit={(event) => { event.preventDefault(); searchDebounce.searchNow(); }}>
      <SearchInput className="min-w-36 flex-1 basis-44" ariaLabel={tr("gl_search_code_name", "ค้นหารหัสหรือชื่อ")} placeholder={tr("gl_search_code_name", "ค้นหารหัสหรือชื่อ")} value={searchDebounce.query} onChange={searchDebounce.setQuery} onClear={searchDebounce.clear} onSearch={searchDebounce.searchNow} />
      <Button type="submit" variant="outline" className={actionClass}><Search className="size-4 mr-1.5" />{tr("gl_search", "ค้นหา")}</Button>
      <Button type="button" variant="outline" className={actionClass} disabled={loading} onClick={() => { setRevision((value) => value + 1); refs.reload(); }}><RefreshCw className="size-4 mr-1.5" />{tr("gl_reload", "โหลดใหม่")}</Button>
      <Button type="button" className={actionClass} disabled={busy} onClick={() => void openCreate()} data-field="budget-add"><Plus className="size-4 mr-1.5" />{tr("gl_add_row", "เพิ่มรายการ")}</Button>
    </form>
    <div className="min-h-0 flex-1 overflow-auto rounded-xl border border-border">
      <table className="w-full text-[0.95rem]" data-field="budget-list">
        <thead className="sticky top-0 z-10 bg-muted/90 text-left"><tr>
          <th className="p-2">{tr("gl_code", "รหัส")}</th><th className="p-2">{tr("gl_name", "ชื่อ")}</th><th className="p-2">{tr("gl_fiscal_year", "ปีบัญชี")}</th><th className="p-2">{tr("gl_status", "สถานะ")}</th><th className="p-2 text-right">{tr("gl_total", "ยอดรวม")}</th>
        </tr></thead>
        <tbody>
          {list.items.map((item) => <tr key={item.code} className={`border-t border-border cursor-pointer hover:bg-accent/60 ${draft?.id === item.code ? "bg-primary/10" : ""}`} onClick={() => void openBudget(item.code)}>
            <td className="p-2"><button type="button" className="font-mono font-semibold text-primary hover:underline" onClick={(event) => { event.stopPropagation(); void openBudget(item.code); }}>{item.code}</button></td>
            <td className="p-2 [overflow-wrap:anywhere]">{item.name}</td>
            <td className="p-2">{item.fiscalyear}</td>
            <td className="p-2">{item.status === "closed" ? tr("gl_closed", "ปิดแล้ว") : tr("gl_open", "เปิด")}</td>
            <td className="p-2 text-right tabular-nums">{formatAmount(item.total ?? "0")}</td>
          </tr>)}
          {!list.items.length && !loading && <tr><td colSpan={5} className="p-4 text-center text-muted-foreground">{search ? tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก") : tr("gl_no_entries_add", "ยังไม่มีรายการ กดเพิ่มรายการเพื่อเริ่มต้น")}</td></tr>}
        </tbody>
      </table>
    </div>
    <Pager page={page} total={list.total} onPage={setPage} loading={loading} limit={LIST_LIMIT} />
  </div>;

  const editorPane = !draft ? <div className="flex flex-1 items-center justify-center rounded-xl border border-dashed border-border p-6 text-center text-muted-foreground">{tr("gl_budget_select_or_add", "เลือกงบประมาณจากรายการ หรือกดเพิ่มรายการเพื่อเริ่มกำหนดงบประมาณ")}</div>
    : <form className="flex flex-col flex-1 min-h-0 gap-3" onSubmit={(event) => { event.preventDefault(); void save(); }} data-field="budget-editor">
      <div className="shrink-0 flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2 min-w-0">
          <h2 className="text-lg font-semibold [overflow-wrap:anywhere]">{draft.id ? `${draft.code} · ${draft.name}` : tr("gl_budget_new", "งบประมาณใหม่")}</h2>
          <UnsavedBadge dirty={dirty} />
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button type="submit" className={actionClass} disabled={busy || !dirty}><Save className="size-4 mr-1.5" />{busy ? tr("gl_saving_please_wait", "กำลังบันทึก กรุณารอสักครู่") : tr("gl_save", "บันทึก")}</Button>
          <Button type="button" variant="outline" className={actionClass} disabled={busy || !dirty} onClick={revert}><Undo2 className="size-4 mr-1.5" />{tr("gl_budget_revert", "คืนค่าที่บันทึกไว้")}</Button>
          {draft.id && <Button type="button" variant="outline" className={`${actionClass} text-destructive`} disabled={busy} onClick={() => void remove()}><Trash2 className="size-4 mr-1.5" />{tr("gl_delete_item", "ลบรายการ")}</Button>}
          <Button type="button" variant="ghost" className={actionClass} disabled={busy} onClick={() => void closeEditor()} aria-label={tr("gl_budget_close", "ปิดหน้าแก้ไข")} title={tr("gl_budget_close", "ปิดหน้าแก้ไข")}><X className="size-4" /></Button>
        </div>
      </div>
      <Notice text={message} />
      <Notice error text={error || refs.error} />
      <fieldset disabled={busy} className="shrink-0 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <Field label={tr("gl_budget_code", "รหัสงบประมาณ")}><input className={control} data-field="code" value={draft.code} disabled={Boolean(draft.id)} maxLength={50} onChange={(e) => setHeader({ code: e.target.value })} /></Field>
        <Field label={tr("gl_budget_name", "ชื่องบประมาณ")}><input className={control} data-field="name" value={draft.name} maxLength={200} onChange={(e) => setHeader({ name: e.target.value })} /></Field>
        <Field label={tr("gl_fiscal_year", "ปีบัญชี")}><YearSelect years={refs.years} value={draft.fiscalyear} onChange={(value) => setHeader({ fiscalyear: value })} /></Field>
        <Field label={tr("gl_status", "สถานะ")}><select className={control} data-field="status" value={draft.status} onChange={(e) => setHeader({ status: e.target.value === "closed" ? "closed" : "open" })}><option value="open">{tr("gl_open", "เปิด")}</option><option value="closed">{tr("gl_closed", "ปิดแล้ว")}</option></select></Field>
        <Field label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} data-field="branchcode" value={draft.branchcode} placeholder={tr("gl_all_branches", "ทุกสาขา")} onChange={(e) => setHeader({ branchcode: e.target.value })} /></Field>
        <Field label={tr("gl_department_code", "รหัสแผนก")}><input className={control} data-field="departmentcode" value={draft.departmentcode} placeholder={tr("gl_all_departments", "ทุกแผนก")} onChange={(e) => setHeader({ departmentcode: e.target.value })} /></Field>
        <Field label={tr("gl_project_code", "รหัสโครงการ")}><input className={control} data-field="projectcode" value={draft.projectcode} placeholder={tr("gl_all_projects", "ทุกโครงการ")} onChange={(e) => setHeader({ projectcode: e.target.value })} /></Field>
        <Field label={tr("gl_note", "หมายเหตุ")}><input className={control} data-field="remark" value={draft.remark} onChange={(e) => setHeader({ remark: e.target.value })} /></Field>
      </fieldset>
      <p className="shrink-0 text-[0.9rem] text-muted-foreground">{tr("gl_budget_spread_hint", "ใส่ยอดทั้งปีแล้วกด แบ่งงวด ระบบจะแบ่งเท่ากันทุกงวดของปีบัญชี (เศษสตางค์อยู่งวดสุดท้าย) หรือพิมพ์ยอดรายเดือนเองได้")}</p>
      <div className="min-h-0 flex-1 overflow-auto rounded-xl border border-border">
        <table className="w-max min-w-full text-[0.9rem]" data-field="budget-lines">
          <thead className="sticky top-0 z-10 bg-muted/95"><tr>
            <th className="sticky left-0 z-20 bg-muted p-2 text-left min-w-64">{tr("gl_account", "บัญชี")}</th>
            <th className="p-2 text-right min-w-36">{tr("gl_budget_annual", "ยอดทั้งปี")}</th>
            <th className="p-2" aria-label={tr("gl_budget_spread", "แบ่งงวด")} />
            {Array.from({ length: BUDGET_PERIODS }, (_, period) => <th key={period} className={`p-2 text-right min-w-32 ${period >= periodCount ? "text-muted-foreground" : ""}`}>{periodLabel(period)}</th>)}
            <th className="p-2 text-right min-w-36">{tr("gl_budget_line_total", "รวมทั้งปี")}</th>
            <th className="p-2" aria-label={tr("gl_budget_remove_line", "ลบบัญชีนี้")} />
          </tr></thead>
          <tbody>
            {lines.map((line, index) => <tr key={index} className="border-t border-border align-top">
              <td className="sticky left-0 z-10 bg-card p-1.5"><AccountSelect accounts={postingAccounts} value={line.accountcode} onChange={(value) => setLine(index, { accountcode: value })} field={`line-${index}-account`} /></td>
              <td className="p-1.5"><AmountInput value={annual[index] ?? ""} onChange={(value) => setAnnual((values) => values.map((old, i) => i === index ? value : old))} ariaLabel={tr("gl_budget_annual", "ยอดทั้งปี")} disabled={busy} /></td>
              <td className="p-1.5"><Button type="button" variant="outline" className={actionClass} disabled={busy} onClick={() => void spreadLine(index)} title={tr("gl_budget_spread_title", "แบ่งยอดทั้งปีเท่ากันทุกงวด")}><Columns3 className="size-4 mr-1.5" />{tr("gl_budget_spread", "แบ่งงวด")}</Button></td>
              {line.periods.map((value, period) => <td key={period} className="p-1.5"><AmountInput value={period < periodCount ? value : ""} onChange={(next) => setPeriod(index, period, next)} disabled={busy || period >= periodCount} ariaLabel={`${line.accountcode || tr("gl_account", "บัญชี")} ${periodLabel(period)}`} /></td>)}
              <td className="p-2 text-right font-semibold tabular-nums">{money(lineTotals[index] ?? 0n)}</td>
              <td className="p-1.5"><Button type="button" variant="ghost" className={actionClass} disabled={busy} onClick={() => removeLine(index)} aria-label={tr("gl_budget_remove_line", "ลบบัญชีนี้")} title={tr("gl_budget_remove_line", "ลบบัญชีนี้")}><Trash2 className="size-4 text-destructive" /></Button></td>
            </tr>)}
          </tbody>
          <tfoot className="sticky bottom-0 bg-muted/95 font-semibold"><tr className="border-t-2 border-border">
            <td className="sticky left-0 bg-muted p-2">{tr("gl_total", "ยอดรวม")}</td><td /><td />
            {periodTotals.map((value, period) => <td key={period} className="p-2 text-right tabular-nums">{period < periodCount ? money(value) : ""}</td>)}
            <td className="p-2 text-right tabular-nums" data-field="budget-grand-total">{money(grandTotal)}</td><td />
          </tr></tfoot>
        </table>
      </div>
      <div className="shrink-0"><Button type="button" variant="outline" className={actionClass} disabled={busy} onClick={addLine}><Plus className="size-4 mr-1.5" />{tr("gl_budget_add_account", "เพิ่มบัญชี")}</Button></div>
    </form>;

  return <>
    {!draft && <Notice text={message} />}
    {!draft && <Notice error text={error || refs.error} />}
    <SplitWorkbench list={listPane} editor={editorPane} />
    {confirmationDialog}
  </>;
}
