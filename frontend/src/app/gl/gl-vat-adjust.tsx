"use client";

import { Fragment, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Calculator, ListChecks, RotateCcw, Save } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { fillText, formatAmount, type GLCommand, type GLPage, type GLVatAdjustFilter, type GLVatAdjustItem, type GLVatAdjustRow } from "@/lib/general-ledger";
import { vatClaimTiming, vatPeriodChoices, vatPeriodPatch, vatPeriodValue } from "@/lib/gl-journal-details";
import { GLCommandError, commandFailure, glRequest } from "@/lib/general-ledger-api";
import { formatAppDate } from "@/lib/date-time";
import { toast } from "@/lib/toast";
import { cn } from "@/lib/utils";
import { Field, Notice, Pager, UnsavedBadge, actionClass, control, panel, useDirtyGuard, useGLCommand, useGLLanguage, useGLText } from "./gl-common";
import { VatClaimHint } from "./gl-journal-details";

/** Filter of Champ 5498: document date, tax invoice date and creditor code, each a from/to pair (blank = not used). */
export type VatAdjustFilterForm = { docdatefrom: string; docdateto: string; invdatefrom: string; invdateto: string; partnerfrom: string; partnerto: string };
export const EMPTY_VAT_ADJUST_FILTER: VatAdjustFilterForm = { docdatefrom: "", docdateto: "", invdatefrom: "", invdateto: "", partnerfrom: "", partnerto: "" };
export const VAT_ADJUST_PAGE_SIZE = 100;
const MONTH_NAMES = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];

/** Filter sent to the backend: trimmed, blank fields are not sent. */
export function vatAdjustFilterPayload(filter: VatAdjustFilterForm): GLVatAdjustFilter {
  const payload: GLVatAdjustFilter = {};
  for (const key of Object.keys(EMPTY_VAT_ADJUST_FILTER) as (keyof VatAdjustFilterForm)[]) {
    const value = (filter[key] ?? "").trim();
    if (value) payload[key] = value;
  }
  return payload;
}
/** Champ refuses to load without a document-date or tax-invoice-date bound; the backend checks the same rule. */
export function vatAdjustHasDate(filter: VatAdjustFilterForm): boolean {
  return [filter.docdatefrom, filter.docdateto, filter.invdatefrom, filter.invdateto].some((value) => value.trim() !== "");
}
export function vatAdjustQuery(filter: VatAdjustFilterForm, page: number, limit = VAT_ADJUST_PAGE_SIZE): string {
  const query = new URLSearchParams({ ...vatAdjustFilterPayload(filter), page: String(page), limit: String(limit) });
  return query.toString();
}
export function sameVatAdjustFilter(a: VatAdjustFilterForm, b: VatAdjustFilterForm): boolean {
  return JSON.stringify(vatAdjustFilterPayload(a)) === JSON.stringify(vatAdjustFilterPayload(b));
}

/** A changed row: the row as loaded plus the claim period (year/month, 0/0 = none) and claim status the user chose. */
export type VatAdjustEdit = { row: GLVatAdjustRow; year: number; month: number; claim: number };
export const vatAdjustKey = (row: Pick<GLVatAdjustRow, "journalid" | "id">) => `${row.journalid}:${row.id}`;
const originalOf = (row: GLVatAdjustRow) => ({ year: row.tax_period_year ?? 0, month: row.tax_period_month ?? 0, claim: row.claim_status ?? 0 });

/** Applies one cell change; a row that goes back to its loaded values leaves the edit set (so it is not saved). */
export function setVatAdjustEdit(edits: ReadonlyMap<string, VatAdjustEdit>, row: GLVatAdjustRow, patch: { period?: string; claim?: number }): Map<string, VatAdjustEdit> {
  const key = vatAdjustKey(row);
  const base = edits.get(key) ?? { row, ...originalOf(row) };
  const period = patch.period === undefined ? null : vatPeriodPatch(patch.period);
  const next: VatAdjustEdit = { row: base.row, year: period ? period.tax_period_year ?? 0 : base.year, month: period ? period.tax_period_month ?? 0 : base.month, claim: patch.claim ?? base.claim };
  const original = originalOf(row);
  const out = new Map(edits);
  if (next.year === original.year && next.month === original.month && next.claim === original.claim) out.delete(key); else out.set(key, next);
  return out;
}
/** The row as the user sees it now (loaded values overlaid with the edit). */
export function vatAdjustMergedRow(row: GLVatAdjustRow, edit?: VatAdjustEdit): GLVatAdjustRow {
  return edit ? { ...row, tax_period_year: edit.year || undefined, tax_period_month: edit.month || undefined, claim_status: edit.claim } : row;
}
export function vatAdjustItems(edits: ReadonlyMap<string, VatAdjustEdit>): GLVatAdjustItem[] {
  return [...edits.values()].map(({ row, year, month, claim }) => ({ journalid: row.journalid, version: row.version, vatid: row.id, taxperiodyear: year, taxperiodmonth: month, claimstatus: claim }));
}
export type VatAdjustProblem = { key: string; row: GLVatAdjustRow; kind: "missing_period" | "before" | "exceeded" | "credit_note_before" | "buddhist_year" };
/** Rows the backend would refuse (claimed without a period, outside the 6-month window, Buddhist year) — shown before saving. */
export function vatAdjustProblems(edits: ReadonlyMap<string, VatAdjustEdit>): VatAdjustProblem[] {
  const problems: VatAdjustProblem[] = [];
  for (const [key, edit] of edits) {
    const merged = vatAdjustMergedRow(edit.row, edit);
    const timing = vatClaimTiming(merged);
    if (timing && timing.kind !== "late") problems.push({ key, row: edit.row, kind: timing.kind });
    else if (edit.claim === 1 && !(edit.year && edit.month)) problems.push({ key, row: edit.row, kind: "missing_period" });
  }
  return problems;
}
/** Command saving the changed rows (backend resource vat-adjustments, mode rows): one transaction, all or nothing. */
export function vatAdjustRowsCommand(edits: ReadonlyMap<string, VatAdjustEdit>, reason: string): Omit<GLCommand, "requestid"> {
  return { resource: "vat-adjustments", action: "update", reason: reason.trim(), vatadjust: { mode: "rows", items: vatAdjustItems(edits) } };
}
/** Command of the button that sets the claim period to the invoice month: only the rows of the filter, counted as shown. */
export function vatAdjustInvoiceDateCommand(filter: VatAdjustFilterForm, expectedCount: number, reason: string): Omit<GLCommand, "requestid"> {
  return { resource: "vat-adjustments", action: "update", reason: reason.trim(), vatadjust: { mode: "invoice_date", filter: vatAdjustFilterPayload(filter), expectedcount: expectedCount } };
}

export function GLVatAdjust({ route }: { route: string }) {
  const tr = useGLText(), language = useGLLanguage();
  const { busy, execute } = useGLCommand();
  const { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const [filter, setFilter] = useState<VatAdjustFilterForm>(EMPTY_VAT_ADJUST_FILTER);
  const [applied, setApplied] = useState<VatAdjustFilterForm | null>(null);
  const [data, setData] = useState<GLPage<GLVatAdjustRow> | null>(null);
  const [edits, setEdits] = useState<Map<string, VatAdjustEdit>>(() => new Map());
  const [reason, setReason] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const rootRef = useRef<HTMLElement>(null), reasonRef = useRef<HTMLInputElement>(null), loadSeq = useRef(0);
  const dirty = edits.size > 0;
  useDirtyGuard(route, dirty);
  const working = busy || loading;
  const rows = useMemo(() => data?.items ?? [], [data]);
  const total = data?.total ?? 0;
  const problems = useMemo(() => vatAdjustProblems(edits), [edits]);
  const filterChanged = applied !== null && !sameVatAdjustFilter(filter, applied);

  const load = useCallback(async (source: VatAdjustFilterForm, page: number) => {
    const seq = ++loadSeq.current;
    setLoading(true);
    try {
      const result = await glRequest<GLPage<GLVatAdjustRow>>(`vat-adjustments?${vatAdjustQuery(source, page)}`);
      if (seq !== loadSeq.current) return;
      setData({ ...result, items: result?.items ?? [], total: result?.total ?? 0 });
    } catch (e) {
      if (seq !== loadSeq.current) return;
      setData(null);
      setError(commandFailure(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง")).message);
    } finally { if (seq === loadSeq.current) setLoading(false); }
  }, [tr]);

  const discardOk = useCallback(async () => !dirty || await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning" }), [dirty, confirm, tr]);

  const runProcess = useCallback(async () => {
    if (working) return;
    if (!vatAdjustHasDate(filter)) { setError(tr("gl_err_vat_adjust_filter_required", "กรุณาระบุช่วงวันที่เอกสารหรือวันที่ใบกำกับภาษีอย่างน้อย 1 ช่วง ก่อนประมวลผล")); return; }
    if (!await discardOk()) return;
    setError(""); setEdits(new Map()); setApplied(filter);
    await load(filter, 1);
  }, [working, filter, discardOk, load, tr]);

  const goPage = useCallback((page: number) => { if (applied) void load(applied, page); }, [applied, load]);
  const reloadCurrent = useCallback(async () => { if (applied) await load(applied, data?.page ?? 1); }, [applied, data?.page, load]);

  const needReason = useCallback(() => {
    if (reason.trim()) return false;
    setError(tr("gl_err_tax_edit_reason_required", "กรุณาระบุเหตุผลการแก้รายการภาษีหลังผ่านบัญชี (เก็บไว้ในประวัติการแก้ไข)"));
    reasonRef.current?.focus();
    return true;
  }, [reason, tr]);

  // The save and the bulk button both end the same way: show the backend answer, then re-read the rows (versions moved).
  const finish = useCallback(async (run: () => Promise<{ adjustedjournals?: number } | undefined>, doneText: (journals: string) => string) => {
    try {
      setError("");
      const result = await run();
      toast.success(doneText((result?.adjustedjournals ?? 0).toLocaleString("th-TH")));
      setEdits(new Map()); setReason("");
    } catch (e) {
      setError(commandFailure(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง")).message);
      // Rows mode keeps the user edits unless the rows moved under them; the bulk button has no edits, so it always re-reads.
      if (!(e instanceof GLCommandError && e.code === "stale_version") && edits.size > 0) return;
      setEdits(new Map());
    }
    await reloadCurrent();
  }, [reloadCurrent, edits.size, tr]);

  const runSave = useCallback(async () => {
    if (working || !dirty) return;
    if (problems.length) { setError(tr("gl_vat_adjust_problem_rows", "ตรวจรายการที่มีข้อความเตือนสีแดงใต้แถวก่อนบันทึก")); return; }
    if (needReason()) return;
    if (!await confirm({
      title: fillText(tr("gl_vat_adjust_confirm_save_title", "บันทึกการแก้ไข {0} รายการ?"), edits.size.toLocaleString("th-TH")),
      description: tr("gl_vat_adjust_confirm_save_detail", "ระบบจะแก้งวดที่ใช้สิทธิและสถานะของรายการเหล่านี้พร้อมกัน ถ้ามีรายการใดไม่ผ่านกฎ จะไม่บันทึกเลยสักรายการ ประวัติการแก้ไขและเหตุผลถูกเก็บไว้ รายงานภาษีซื้อของงวดที่ยื่นไปแล้วอาจเปลี่ยนตาม"),
      confirmLabel: tr("gl_save", "บันทึก"),
      tone: "warning",
    })) return;
    await finish(() => execute(vatAdjustRowsCommand(edits, reason)), (journals) => fillText(tr("gl_vat_adjust_saved", "บันทึกการแก้ไขภาษีซื้อแล้ว {0} ใบสำคัญ"), journals));
  }, [working, dirty, problems, needReason, confirm, edits, reason, finish, execute, tr]);

  const runBulk = useCallback(async () => {
    if (working || dirty || !applied || filterChanged || total < 1) return;
    if (needReason()) return;
    if (!await confirm({
      title: fillText(tr("gl_vat_adjust_confirm_bulk_title", "ตั้งวันที่ยื่นภาษี = วันที่ใบกำกับภาษี {0} รายการ?"), total.toLocaleString("th-TH")),
      description: tr("gl_vat_adjust_confirm_bulk_detail", "ระบบจะตั้งงวดที่ยื่นภาษีของเฉพาะรายการตามเงื่อนไขที่แสดงอยู่ (ทุกหน้า) ให้เป็นเดือนของวันที่ใบกำกับภาษี โดยคงสถานะการใช้สิทธิเดิม รายการนอกเงื่อนไขไม่ถูกแก้ ถ้ามีรายการใดไม่ผ่านกฎ จะไม่แก้เลยสักรายการ"),
      confirmLabel: tr("gl_vat_adjust_set_by_invoice", "ตั้งวันที่ยื่นภาษี = วันที่ใบกำกับภาษี"),
      tone: "warning",
    })) return;
    await finish(() => execute(vatAdjustInvoiceDateCommand(applied, total, reason)), (journals) => fillText(tr("gl_vat_adjust_bulk_done", "ตั้งวันที่ยื่นภาษีตามวันที่ใบกำกับภาษีแล้ว {0} ใบสำคัญ"), journals));
  }, [working, dirty, applied, filterChanged, total, needReason, confirm, finish, execute, reason, tr]);

  // F7 process / F11 save as on the other GL entry screens; ignored while a dialog is open or another tab is showing.
  const dialogOpen = confirmationDialog !== null;
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if ((event.key !== "F7" && event.key !== "F11") || event.repeat || dialogOpen || !rootRef.current?.getClientRects().length) return;
      event.preventDefault();
      if (working) return;
      if (event.key === "F7") void runProcess(); else void runSave();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [dialogOpen, working, runProcess, runSave]);

  const change = (patch: Partial<VatAdjustFilterForm>) => { setFilter((value) => ({ ...value, ...patch })); setError(""); };
  const periodLabel = (choice: string) => {
    const [year, month] = choice.split("-").map(Number);
    const name = MONTH_NAMES[month - 1];
    return tr("gl_vat_period_option", "{0} {2}").replace("{0}", tr(`month_${name.toLowerCase()}`, name)).replace("{1}", String(year)).replace("{2}", String(year + 543));
  };
  const range = (label: string, fromKey: keyof VatAdjustFilterForm, toKey: keyof VatAdjustFilterForm, kind: "date" | "text") => {
    const from = tr("gl_vat_adjust_from", "จาก"), to = tr("gl_vat_adjust_to", "ถึง");
    return <fieldset className="grid min-w-0 gap-1 text-[0.95rem] leading-normal" disabled={working}>
      <legend className="mb-1 font-medium text-foreground">{label}</legend>
      <div className="grid grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-2">
        <input className={control} type={kind} aria-label={`${label} ${from}`} value={filter[fromKey]} maxLength={kind === "text" ? 60 : undefined} onChange={(event) => change({ [fromKey]: event.target.value })} />
        <span className="text-muted-foreground">{to}</span>
        <input className={control} type={kind} aria-label={`${label} ${to}`} value={filter[toKey]} maxLength={kind === "text" ? 60 : undefined} onChange={(event) => change({ [toKey]: event.target.value })} />
      </div>
    </fieldset>;
  };
  const problemText = (problem: VatAdjustProblem["kind"]) => problem === "missing_period" ? tr("gl_err_vat_claim_period_required", "ภาษีซื้อที่ใช้สิทธิต้องระบุงวดภาษี") : "";
  const claimOptions: [number, string][] = [[1, tr("gl_vat_claim_used", "ใช้สิทธิในงวดนี้")], [2, tr("gl_vat_claim_forbidden", "ภาษีซื้อต้องห้าม")], [3, tr("gl_vat_claim_pending", "รอใช้สิทธิ")], [4, tr("gl_vat_claim_not_used", "ไม่ใช้สิทธิ")]];
  const problemKeys = new Set(problems.map((problem) => problem.key));

  return <section ref={rootRef} className={cn(panel, "flex min-h-0 flex-1 flex-col gap-3")} aria-busy={working} data-gl-vat-adjust="true">
    <div className="grid gap-3 lg:grid-cols-3">
      {range(tr("gl_document_date", "วันที่เอกสาร"), "docdatefrom", "docdateto", "date")}
      {range(tr("gl_detail_vats_tax_invoice_date", "วันที่ใบกำกับภาษี"), "invdatefrom", "invdateto", "date")}
      {range(tr("gl_vat_adjust_creditor_code", "รหัสเจ้าหนี้"), "partnerfrom", "partnerto", "text")}
    </div>
    <p className="text-[0.9rem] leading-relaxed text-muted-foreground">{tr("gl_vat_adjust_filter_hint", "ต้องระบุช่วงวันที่เอกสารหรือวันที่ใบกำกับภาษีอย่างน้อย 1 ช่วง · ใส่เฉพาะช่องจาก = ค้นเฉพาะค่านั้นค่าเดียว · ใส่ทั้งจากและถึง = ค้นเป็นช่วง")}</p>
    <div className="grid gap-3 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
      <Field label={`${tr("gl_reason", "เหตุผล")} *`}>
        <input ref={reasonRef} className={control} value={reason} disabled={working} maxLength={500} aria-required="true" placeholder={tr("gl_vat_adjust_reason_hint", "เช่น ย้ายงวดตามใบกำกับภาษีที่ได้รับ (เก็บไว้ในประวัติการแก้ไข)")} onChange={(event) => setReason(event.target.value)} />
      </Field>
      <div className="flex flex-wrap items-center gap-2">
        <Button className={actionClass} variant="outline" disabled={working} onClick={() => void runProcess()}><ListChecks className="mr-1.5 size-4" aria-hidden="true" />{tr("gl_post_batch_check", "ประมวลผล")} (F7)</Button>
        <Button className={actionClass} variant="outline" disabled={working || dirty || !applied || filterChanged || total < 1} title={dirty ? tr("gl_vat_adjust_save_first", "มีการแก้ไขที่ยังไม่บันทึก — บันทึกหรือล้างการแก้ไขก่อนใช้ปุ่มนี้") : undefined} onClick={() => void runBulk()}><Calculator className="mr-1.5 size-4" aria-hidden="true" />{tr("gl_vat_adjust_set_by_invoice", "ตั้งวันที่ยื่นภาษี = วันที่ใบกำกับภาษี")}</Button>
        <Button className={actionClass} variant="outline" disabled={working || !dirty} onClick={() => setEdits(new Map())}><RotateCcw className="mr-1.5 size-4" aria-hidden="true" />{tr("gl_vat_adjust_discard", "ล้างการแก้ไข")}</Button>
        <Button className={actionClass} disabled={working || !dirty} onClick={() => void runSave()}><Save className="mr-1.5 size-4" aria-hidden="true" />{working ? tr("gl_processing", "กำลังประมวลผล…") : tr("gl_save", "บันทึก")} (F11)</Button>
        <UnsavedBadge dirty={dirty} />
      </div>
    </div>
    <Notice text={error} error />
    {filterChanged && <Notice text={tr("gl_vat_adjust_filter_changed_note", "เงื่อนไขเปลี่ยนไปแล้ว กด ประมวลผล (F7) เพื่อโหลดรายการใหม่")} />}
    {!data && !loading && !error && <Notice text={tr("gl_vat_adjust_hint", "กำหนดเงื่อนไข แล้วกด ประมวลผล (F7) เพื่อค้นหารายการภาษีซื้อ จากนั้นเลือกการใช้สิทธิหรือวันที่ยื่นภาษีในแถวที่ต้องการแก้ แล้วกด บันทึก (F11) หรือกดตั้งวันที่ยื่นภาษี = วันที่ใบกำกับภาษี เพื่อตั้งให้ทุกรายการตามเงื่อนไข")} />}
    {data && <div className="min-h-0 flex-1 overflow-auto rounded-xl border border-border">
      <table className="w-full min-w-[1180px] border-collapse text-[0.95rem]">
        <thead className="sticky top-0 z-10 bg-muted text-left">
          <tr className="[&_th]:px-3 [&_th]:py-2 [&_th]:font-semibold">
            <th>{tr("gl_document_no", "เลขที่เอกสาร")}</th>
            <th>{tr("gl_vat_adjust_col_invoice_no", "เลขที่ใบกำกับ")}</th>
            <th>{tr("gl_detail_vats_tax_invoice_date", "วันที่ใบกำกับภาษี")}</th>
            <th>{tr("gl_description", "คำอธิบาย")}</th>
            <th className="text-right">{tr("gl_vat_adjust_col_rate", "อัตรา")}</th>
            <th className="text-right">{tr("gl_vat_adjust_col_base", "ยอดก่อนภาษี")}</th>
            <th className="text-right">{tr("gl_vat_adjust_col_vat", "ยอดภาษี")}</th>
            <th className="min-w-[12rem]">{tr("gl_detail_vats_claim_status", "การใช้สิทธิภาษีซื้อ")}</th>
            <th className="min-w-[12rem]">{tr("gl_vat_adjust_col_period", "วันที่ยื่นภาษี (งวด)")}</th>
          </tr>
        </thead>
        <tbody>
          {rows.length === 0 && <tr><td colSpan={9} className="px-3 py-6 text-center text-muted-foreground">{tr("gl_vat_adjust_none", "ไม่พบรายการภาษีซื้อตามเงื่อนไขนี้")}</td></tr>}
          {rows.map((row) => {
            const key = vatAdjustKey(row), edit = edits.get(key), merged = vatAdjustMergedRow(row, edit);
            const period = vatPeriodValue(merged);
            const showHint = edit !== undefined && (vatClaimTiming(merged) !== null || problemKeys.has(key));
            const missingPeriod = problems.find((problem) => problem.key === key)?.kind === "missing_period";
            const bg = edit ? "bg-primary/15 hover:bg-primary/20" : "hover:bg-muted/40";
            return <Fragment key={key}>
              <tr className={cn("border-t border-border align-top [&_td]:px-3 [&_td]:py-2", bg)} data-vat-adjust-row={edit ? "edited" : "loaded"}>
                <td className="whitespace-nowrap"><span className="font-medium">{row.docno}</span><span className="block text-[0.9rem] text-muted-foreground">{row.docdate ? formatAppDate(row.docdate, language) : "-"}</span></td>
                <td className="whitespace-nowrap font-medium">{row.tax_invoice_no}</td>
                <td className="whitespace-nowrap">{row.tax_invoice_date ? formatAppDate(row.tax_invoice_date, language) : "-"}</td>
                <td className="[overflow-wrap:anywhere]">{row.partner_name || "-"}{row.remark && <span className="block text-[0.9rem] text-muted-foreground">{row.remark}</span>}</td>
                <td className="whitespace-nowrap text-right tabular-nums">{formatAmount(row.vat_rate ?? "")}</td>
                <td className="whitespace-nowrap text-right tabular-nums">{formatAmount(row.base_amount ?? "")}</td>
                <td className="whitespace-nowrap text-right tabular-nums">{formatAmount(row.vat_amount ?? "")}</td>
                <td>
                  <select className={control} aria-label={fillText(tr("gl_vat_adjust_claim_aria", "การใช้สิทธิของใบกำกับ {0}"), row.tax_invoice_no)} value={merged.claim_status ?? 0} disabled={working}
                    onChange={(event) => setEdits((value) => setVatAdjustEdit(value, row, { claim: Number(event.target.value) }))}>
                    {claimOptions.map(([value, label]) => <option key={value} value={value}>{label}</option>)}
                  </select>
                </td>
                <td>
                  <select className={control} aria-label={fillText(tr("gl_vat_adjust_period_aria", "วันที่ยื่นภาษีของใบกำกับ {0}"), row.tax_invoice_no)} value={period} disabled={working}
                    onChange={(event) => setEdits((value) => setVatAdjustEdit(value, row, { period: event.target.value }))}>
                    <option value="">{tr("gl_vat_period_none", "ยังไม่กำหนดงวด (รอใช้สิทธิ)")}</option>
                    {vatPeriodChoices(row.tax_invoice_date, period).map((choice) => <option key={choice} value={choice}>{periodLabel(choice)}</option>)}
                  </select>
                </td>
              </tr>
              {showHint && <tr className={bg} data-vat-adjust-hint="true"><td colSpan={9} className="px-3 pb-2">
                <VatClaimHint row={merged} tr={tr} />
                {missingPeriod && <p role="alert" className="rounded-lg border border-destructive/40 bg-destructive/5 p-2 leading-relaxed text-foreground">{problemText("missing_period")}</p>}
              </td></tr>}
            </Fragment>;
          })}
        </tbody>
      </table>
    </div>}
    {data && <Pager page={data.page || 1} total={total} onPage={goPage} loading={working} limit={VAT_ADJUST_PAGE_SIZE} />}
    {data && <div className="flex flex-wrap items-center justify-between gap-2 text-[0.95rem]" data-vat-adjust-summary="true">
      <span>{fillText(tr("gl_vat_adjust_summary", "พบ {0} รายการ · แก้ไขแล้ว {1} รายการ (ยังไม่บันทึก)"), total.toLocaleString("th-TH"), edits.size.toLocaleString("th-TH"))}</span>
    </div>}
    {confirmationDialog}
  </section>;
}
