"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AlertTriangle, CheckCircle2, ListChecks, Send, SquareCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { activeJournalBooks, amountString, displayAmountUnits, fillText, findJournalBook, formatAmount, journalBookName, type GLPostBatchInput, type GLPostCheck, type GLPostCheckRow, type GLTextFn } from "@/lib/general-ledger";
import { commandFailure, glRequest } from "@/lib/general-ledger-api";
import { formatAppDate } from "@/lib/date-time";
import { toast } from "@/lib/toast";
import { cn } from "@/lib/utils";
import { Combobox, Field, Notice, actionClass, control, panel, useGLCommand, useGLLanguage, useGLText, useReferences } from "./gl-common";
import { GLJournals } from "./gl-journals";
import { JournalDrillDownModal } from "./gl-reports";

/** ตัวกรองของจอผ่านรายการแบบชุด (Champ 5512): สมุดรายวัน วันที่ จาก–ถึง เลขที่ จาก–ถึง — ช่องว่าง = ไม่จำกัด */
export type PostBatchFilter = { bookcode: string; datefrom: string; dateto: string; docnofrom: string; docnoto: string };
export const EMPTY_POST_BATCH_FILTER: PostBatchFilter = { bookcode: "", datefrom: "", dateto: "", docnofrom: "", docnoto: "" };

/** ตัวกรองที่ส่งให้ backend (post_batches.go action "check"): ตัดช่องว่างหัวท้าย และไม่ส่งช่องที่ว่าง */
export function postBatchFilterPayload(filter: PostBatchFilter): GLPostBatchInput {
  const payload: GLPostBatchInput = {};
  for (const key of Object.keys(EMPTY_POST_BATCH_FILTER) as (keyof PostBatchFilter)[]) {
    const value = (filter[key] ?? "").trim();
    if (value) payload[key] = value;
  }
  return payload;
}
/** ผู้ตรวจสอบ 1/2 (ไม่บังคับ) และผู้อนุมัติ (บังคับ) ของชุดผ่านรายการ ตามกล่องรายละเอียดชุดของ Champ — ชื่อยาวสุด 50 ตัวอักษร (BatchCheck1/2/Approve VARCHAR(50)) */
export type PostBatchSignatories = { check1: string; check2: string; approver: string };
export const EMPTY_POST_BATCH_SIGNATORIES: PostBatchSignatories = { check1: "", check2: "", approver: "" };
export const POST_BATCH_PERSON_MAX_LENGTH = 50;

/** คำสั่งผ่านรายการ (post_batches.go action "create"): ตัดช่องว่างหัวท้ายทุกช่อง ไม่ส่งผู้ตรวจสอบที่ว่าง */
export function postBatchCreatePayload(items: { id: string; version: number }[], header: { remark: string } & PostBatchSignatories): GLPostBatchInput {
  const payload: GLPostBatchInput = { items, remark: header.remark.trim(), approver: header.approver.trim() };
  const check1 = header.check1.trim(), check2 = header.check2.trim();
  if (check1) payload.check1 = check1;
  if (check2) payload.check2 = check2;
  return payload;
}

/** ใบที่ผ่านการตรวจ (เลือกได้) ตามลำดับในตาราง */
export function readyRowIds(rows: readonly GLPostCheckRow[]): string[] {
  return rows.filter((row) => row.ok).map((row) => row.id);
}
/** F2: เลือกใบที่พร้อมผ่านทุกใบ — ถ้าเลือกครบอยู่แล้วให้ล้างการเลือก; ใบที่ติดปัญหาเลือกไม่ได้เสมอ */
export function toggleAllReady(rows: readonly GLPostCheckRow[], selected: ReadonlySet<string>): Set<string> {
  const ready = readyRowIds(rows);
  return ready.length > 0 && ready.every((id) => selected.has(id)) ? new Set() : new Set(ready);
}
/** รายการที่ส่งผ่านรายการ: เฉพาะใบที่พร้อมและเลือกไว้ เรียงตามตาราง พร้อม version ที่ตรวจ (กันใบที่ถูกแก้ระหว่างนั้น) */
export function postBatchItems(rows: readonly GLPostCheckRow[], selected: ReadonlySet<string>): { id: string; version: number }[] {
  return rows.filter((row) => row.ok && selected.has(row.id)).map(({ id, version }) => ({ id, version }));
}
/** สรุปท้ายตาราง: จำนวนใบ + ยอดเดบิตของใบที่เลือก (คิดด้วย BigInt ไม่ใช้ float — Rule 20); ยอดอ่านไม่ได้ = "" (แสดง "-") */
export function postBatchSummary(rows: readonly GLPostCheckRow[], selected: ReadonlySet<string>) {
  let ready = 0, chosen = 0, debit: bigint | null = 0n;
  for (const row of rows) {
    if (!row.ok) continue;
    ready++;
    if (!selected.has(row.id)) continue;
    chosen++;
    try { if (debit !== null) debit += displayAmountUnits(row.debit || "0"); } catch { debit = null; }
  }
  return { checked: rows.length, ready, issues: rows.length - ready, selected: chosen, debit: debit === null ? "" : amountString(debit, 2) };
}
/** ช่อง "ที่มา" (Champ 5512): ระบบที่นำเข้าใบสำคัญ — ใบที่คีย์ในบัญชีแยกประเภทเองบอกว่าคีย์เอง */
export function postBatchSourceText(row: Pick<GLPostCheckRow, "sourcesystem">, tr: GLTextFn): string {
  return row.sourcesystem?.trim() || tr("gl_post_batch_source_manual", "คีย์ในบัญชีแยกประเภท");
}
export type PostBatchShortcut = "check" | "select-all" | "post";
/** ปุ่มลัดแบบ Champ: F7 ตรวจสอบ · F2 เลือกทั้งหมด · F11 ผ่านรายการ */
export function postBatchShortcut(key: string): PostBatchShortcut | null {
  return key === "F7" ? "check" : key === "F2" ? "select-all" : key === "F11" ? "post" : null;
}

/** เมนู "ผ่านรายการและยกเลิก": แท็บผ่านรายการทีละชุด (ใหม่) + แท็บกลับรายการ (จอเดิม ไม่เปลี่ยน) */
export function GLPostBatch({ route }: { route: string }) {
  const tr = useGLText();
  const [tab, setTab] = useState<"post" | "reverse">("post");
  const tabs = [
    { id: "post" as const, label: tr("gl_post_batch_tab_post", "ผ่านรายการทีละชุด") },
    { id: "reverse" as const, label: tr("gl_post_batch_tab_reverse", "กลับรายการที่ผ่านแล้ว") },
  ];
  return <div className="flex min-h-0 flex-1 flex-col gap-2" data-gl-post-batch="true">
    <div role="tablist" aria-label={tr("gl_post_and_reverse", "ผ่านรายการและยกเลิก")} className="flex flex-wrap gap-2">
      {tabs.map((item) => <button key={item.id} type="button" role="tab" aria-selected={tab === item.id} onClick={() => setTab(item.id)}
        className={cn("min-h-11 rounded-xl border px-4 py-2 font-semibold transition-colors", tab === item.id ? "border-primary bg-primary text-primary-foreground shadow-sm" : "border-border bg-card text-foreground hover:border-primary/60 hover:bg-primary/5")}>
        {item.label}
      </button>)}
    </div>
    {tab === "post" ? <PostBatchPanel /> : <GLJournals key="reverse" route={route} mode="reverse" />}
  </div>;
}

function PostBatchPanel() {
  const tr = useGLText(), language = useGLLanguage();
  const refs = useReferences();
  const books = useMemo(() => activeJournalBooks(refs.books), [refs.books]);
  const { busy, execute } = useGLCommand();
  const { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const [filter, setFilter] = useState<PostBatchFilter>(EMPTY_POST_BATCH_FILTER);
  const [remark, setRemark] = useState("");
  const [signatories, setSignatories] = useState<PostBatchSignatories>(EMPTY_POST_BATCH_SIGNATORIES);
  const approverRef = useRef<HTMLInputElement>(null);
  const [check, setCheck] = useState<GLPostCheck | null>(null);
  const [selected, setSelected] = useState<Set<string>>(() => new Set());
  const [checking, setChecking] = useState(false);
  const [error, setError] = useState("");
  const [viewing, setViewing] = useState<{ id: string; docno: string } | null>(null);
  const rootRef = useRef<HTMLElement>(null), checkSeq = useRef(0);
  const rows = useMemo(() => check?.rows ?? [], [check]);
  const summary = useMemo(() => postBatchSummary(rows, selected), [rows, selected]);
  const working = busy || checking;

  // เปลี่ยนเงื่อนไขแล้วผลตรวจเดิมไม่ตรงกับเงื่อนไขใหม่ → ล้างทิ้ง ให้ตรวจใหม่ก่อนผ่านรายการ
  const changeFilter = (patch: Partial<PostBatchFilter>) => { setFilter((value) => ({ ...value, ...patch })); setCheck(null); setSelected(new Set()); };

  const runCheck = useCallback(async () => {
    const seq = ++checkSeq.current;
    setChecking(true); setError("");
    try {
      const data = await glRequest<{ postcheck?: Partial<GLPostCheck> | null }>("command", { method: "POST", body: JSON.stringify({ resource: "post-batches", action: "check", requestid: crypto.randomUUID(), postbatch: postBatchFilterPayload(filter) }) });
      if (seq !== checkSeq.current) return;
      const result: GLPostCheck = { rows: data?.postcheck?.rows ?? [], more: Boolean(data?.postcheck?.more), limit: data?.postcheck?.limit ?? 0 };
      setCheck(result);
      setSelected(new Set(readyRowIds(result.rows)));
    } catch (e) {
      if (seq !== checkSeq.current) return;
      setCheck(null); setSelected(new Set());
      setError(commandFailure(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง")).message);
    } finally { if (seq === checkSeq.current) setChecking(false); }
  }, [filter, tr]);

  const selectAll = useCallback(() => setSelected((value) => toggleAllReady(rows, value)), [rows]);

  const runPost = useCallback(async () => {
    if (working) return;
    const items = postBatchItems(rows, selected);
    if (!items.length) { setError(tr("gl_post_batch_select_first", "กรุณาเลือกใบที่พร้อมผ่านรายการอย่างน้อย 1 ใบ")); return; }
    if (!signatories.approver.trim()) { setError(tr("gl_err_post_batch_approver_required", "กรุณาระบุชื่อผู้อนุมัติก่อนผ่านรายการ")); approverRef.current?.focus(); return; }
    if (!await confirm({
      title: fillText(tr("gl_post_batch_confirm_title", "ผ่านรายการ {0} ใบ?"), items.length.toLocaleString("th-TH")),
      description: tr("gl_post_batch_confirm_detail", "ระบบจะผ่านรายการทุกใบที่เลือกพร้อมกันเป็นชุดเดียว ถ้ามีใบใดผ่านไม่ได้ จะไม่ผ่านเลยสักใบ หลังผ่านรายการแล้วแก้ไขได้ด้วยการกลับรายการเท่านั้น"),
      confirmLabel: tr("gl_post_accounting_entry", "ผ่านรายการ"),
      tone: "warning",
    })) return;
    try {
      setError("");
      const result = await execute({ resource: "post-batches", action: "create", postbatch: postBatchCreatePayload(items, { remark, ...signatories }) });
      toast.success(fillText(tr("gl_post_batch_done", "ผ่านรายการชุดที่ {0} แล้ว {1} ใบ"), result?.batchno ?? "", (result?.postedjournals ?? items.length).toLocaleString("th-TH")));
      setRemark(""); setSignatories(EMPTY_POST_BATCH_SIGNATORIES);
    } catch (e) {
      setError(commandFailure(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง")).message);
    }
    // ทั้งสำเร็จและไม่สำเร็จ: ตรวจใหม่ให้ตารางตรงกับข้อมูลจริง (ใบที่ผ่านแล้วหายไป / ใบที่มีปัญหาขึ้นเหตุผล)
    await runCheck();
  }, [working, rows, selected, confirm, execute, remark, signatories, runCheck, tr]);

  // ปุ่มลัด F7/F2/F11 ทำงานเฉพาะเมื่อจอนี้มองเห็นอยู่ (แท็บเมนูอื่นที่ซ่อนอยู่ไม่รับ) และไม่มีกล่องยืนยันเปิดอยู่
  const dialogOpen = confirmationDialog !== null || viewing !== null;
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const shortcut = postBatchShortcut(event.key);
      if (!shortcut || event.repeat || dialogOpen || !rootRef.current?.getClientRects().length) return;
      event.preventDefault();
      if (working) return;
      if (shortcut === "check") void runCheck();
      else if (shortcut === "select-all") selectAll();
      else void runPost();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [dialogOpen, working, runCheck, selectAll, runPost]);

  const bookLabel = (code: string) => journalBookName(findJournalBook(refs.books, code), code, language);
  const toggleRow = (id: string, checked: boolean) => setSelected((value) => { const next = new Set(value); if (checked) next.add(id); else next.delete(id); return next; });
  const allReadySelected = summary.ready > 0 && summary.selected === summary.ready;

  return <section ref={rootRef} className={cn(panel, "flex min-h-0 flex-1 flex-col gap-3")} aria-busy={working}>
    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
      <Field label={tr("gl_journal", "สมุดรายวัน")}>
        <Combobox aria-label={tr("gl_journal", "สมุดรายวัน")} value={filter.bookcode} disabled={working} onChange={(value) => changeFilter({ bookcode: String(value) })}
          options={[{ value: "", label: tr("gl_post_batch_all_books", "ทุกสมุดรายวัน") }, ...books.map((book) => ({ value: book.code, label: `${book.code} · ${journalBookName(book, book.code, language)}` }))]} />
      </Field>
      <Field label={tr("gl_post_batch_date_from", "จากวันที่")}><input className={control} type="date" value={filter.datefrom} disabled={working} onChange={(event) => changeFilter({ datefrom: event.target.value })} /></Field>
      <Field label={tr("gl_to_date", "ถึงวันที่")}><input className={control} type="date" value={filter.dateto} disabled={working} onChange={(event) => changeFilter({ dateto: event.target.value })} /></Field>
      <Field label={tr("gl_post_batch_docno_from", "จากเลขที่เอกสาร")}><input className={control} value={filter.docnofrom} disabled={working} maxLength={50} onChange={(event) => changeFilter({ docnofrom: event.target.value })} /></Field>
      <Field label={tr("gl_post_batch_docno_to", "ถึงเลขที่เอกสาร")}><input className={control} value={filter.docnoto} disabled={working} maxLength={50} onChange={(event) => changeFilter({ docnoto: event.target.value })} /></Field>
    </div>
    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <Field label={tr("gl_post_batch_check1", "ผู้ตรวจสอบ 1")}><input className={control} value={signatories.check1} disabled={working} maxLength={POST_BATCH_PERSON_MAX_LENGTH} onChange={(event) => setSignatories((value) => ({ ...value, check1: event.target.value }))} /></Field>
      <Field label={tr("gl_post_batch_check2", "ผู้ตรวจสอบ 2")}><input className={control} value={signatories.check2} disabled={working} maxLength={POST_BATCH_PERSON_MAX_LENGTH} onChange={(event) => setSignatories((value) => ({ ...value, check2: event.target.value }))} /></Field>
      <Field label={`${tr("gl_post_batch_approver", "ผู้อนุมัติ")} *`}><input ref={approverRef} className={control} value={signatories.approver} disabled={working} maxLength={POST_BATCH_PERSON_MAX_LENGTH} aria-required="true" onChange={(event) => setSignatories((value) => ({ ...value, approver: event.target.value }))} /></Field>
      <Field label={tr("gl_post_batch_remark", "หมายเหตุชุดผ่านรายการ")}><input className={control} value={remark} disabled={working} maxLength={500} onChange={(event) => setRemark(event.target.value)} /></Field>
    </div>
    <div className="flex flex-wrap items-center gap-2">
      <Button className={actionClass} variant="outline" disabled={working} onClick={() => void runCheck()}><ListChecks className="mr-1.5 size-4" aria-hidden="true" />{tr("gl_post_batch_check", "ประมวลผล")} (F7)</Button>
      <Button className={actionClass} variant="outline" disabled={working || summary.ready === 0} onClick={selectAll}><SquareCheck className="mr-1.5 size-4" aria-hidden="true" />{allReadySelected ? tr("gl_post_batch_clear_selection", "ไม่เลือกทั้งหมด") : tr("gl_post_batch_select_all", "เลือกทั้งหมด")} (F2)</Button>
      <Button className={actionClass} disabled={working || summary.selected === 0} onClick={() => void runPost()}><Send className="mr-1.5 size-4" aria-hidden="true" />{working ? tr("gl_processing", "กำลังประมวลผล…") : tr("gl_post_batch_save", "บันทึกรายการ")} (F11)</Button>
      <span className="text-[0.9rem] text-muted-foreground">{tr("gl_post_batch_shortcuts", "ปุ่มลัด: F7 ประมวลผล · F2 เลือก/ไม่เลือกทั้งหมด · F11 บันทึกรายการ")}</span>
    </div>
    <Notice text={error} error />
    {refs.error && <Notice text={refs.error} error />}
    {!check && !checking && <Notice text={tr("gl_post_batch_hint", "การผ่านรายการรายวัน: กำหนดเงื่อนไข แล้วกด ประมวลผล (F7) เพื่อค้นหารายการ ทำเครื่องหมายถูกในแถวที่ต้องการ จากนั้นกด บันทึกรายการ (F11) เพื่อยืนยันการผ่านรายการ (ระบบตรวจ เดบิตเท่ากับเครดิต งวดยังไม่ปิด และผังบัญชีลงรายการได้ ให้ทุกใบ)")} />}
    {check?.more && <Notice text={fillText(tr("gl_post_batch_more", "แสดงเพียง {0} ใบแรก กรุณากำหนดช่วงวันที่หรือเลขที่ให้แคบลง แล้วตรวจสอบอีกครั้ง"), check.limit.toLocaleString("th-TH"))} />}
    {check && <div className="min-h-0 flex-1 overflow-auto rounded-xl border border-border">
      <table className="w-full min-w-[960px] border-collapse text-[0.95rem]">
        <thead className="sticky top-0 z-10 bg-muted text-left">
          <tr className="[&_th]:px-3 [&_th]:py-2 [&_th]:font-semibold">
            <th className="w-12"><span className="sr-only">{tr("gl_select", "เลือก")}</span></th>
            <th>{tr("gl_journal", "สมุดรายวัน")}</th>
            <th>{tr("gl_date", "วันที่")}</th>
            <th>{tr("gl_document_no", "เลขที่เอกสาร")}</th>
            <th>{tr("gl_description", "คำอธิบาย")}</th>
            <th className="text-right">{tr("gl_amount", "จำนวนเงิน")}</th>
            <th>{tr("gl_post_batch_source", "ที่มา")}</th>
            <th className="min-w-[16rem]">{tr("gl_status", "สถานะ")}</th>
          </tr>
        </thead>
        <tbody>
          {rows.length === 0 && <tr><td colSpan={8} className="px-3 py-6 text-center text-muted-foreground">{tr("gl_post_batch_none", "ไม่พบใบร่างตามเงื่อนไขนี้")}</td></tr>}
          {rows.map((row) => <tr key={row.id} className={cn("border-t border-border align-top [&_td]:px-3 [&_td]:py-2", row.ok ? "hover:bg-muted/40" : "bg-destructive/5")} data-post-check={row.ok ? "ok" : "issue"}
            onDoubleClick={(event) => { if (!(event.target as HTMLElement).closest("button,input,[role=checkbox]")) setViewing({ id: row.id, docno: row.docno }); }}>
            <td><Checkbox checked={row.ok && selected.has(row.id)} disabled={!row.ok || working} onCheckedChange={(checked) => toggleRow(row.id, checked)} aria-label={fillText(tr("gl_post_batch_select_row", "เลือกเอกสาร {0}"), row.docno)} /></td>
            <td className="whitespace-nowrap">{bookLabel(row.bookcode)}</td>
            <td className="whitespace-nowrap">{row.date ? formatAppDate(row.date, language) : "-"}</td>
            <td className="whitespace-nowrap font-medium"><button type="button" className="min-h-0! rounded-md p-0! font-medium text-primary underline-offset-2 hover:underline focus-visible:ring-2 focus-visible:ring-ring" title={fillText(tr("gl_post_batch_view_voucher", "ดูรายการบัญชีของเอกสาร {0}"), row.docno)} aria-label={fillText(tr("gl_post_batch_view_voucher", "ดูรายการบัญชีของเอกสาร {0}"), row.docno)} onClick={() => setViewing({ id: row.id, docno: row.docno })}>{row.docno}</button></td>
            <td className="[overflow-wrap:anywhere]">{row.description || "-"}</td>
            <td className="whitespace-nowrap text-right tabular-nums">{formatAmount(row.debit ?? "")}</td>
            <td className="[overflow-wrap:anywhere]">{postBatchSourceText(row, tr)}</td>
            <td>{row.ok
              ? <span className="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2.5 py-1 font-semibold text-primary"><CheckCircle2 className="size-4 shrink-0" aria-hidden="true" />{tr("gl_post_batch_ready", "พร้อมผ่านรายการ")}</span>
              : <span className="inline-flex items-start gap-1.5 text-destructive"><AlertTriangle className="mt-0.5 size-4 shrink-0" aria-hidden="true" /><span className="leading-relaxed"><span className="font-semibold">{tr("gl_post_batch_issue", "ติดปัญหา")}:</span> {row.message || row.code || "-"}</span></span>}</td>
          </tr>)}
        </tbody>
      </table>
    </div>}
    {check && <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-2 text-[0.95rem]" data-post-batch-summary="true">
      <span>{fillText(tr("gl_post_batch_summary", "ตรวจแล้ว {0} ใบ · พร้อมผ่าน {1} ใบ · ติดปัญหา {2} ใบ · เลือกไว้ {3} ใบ"), summary.checked, summary.ready, summary.issues, summary.selected)}</span>
      <span className="font-semibold">{tr("gl_post_batch_selected_debit", "ยอดเดบิตที่เลือก")}: <span className="tabular-nums">{summary.debit ? formatAmount(summary.debit) : "-"}</span></span>
    </div>}
    <JournalDrillDownModal docno={viewing?.docno ?? null} journalId={viewing?.id ?? null} open={viewing !== null} onClose={() => setViewing(null)} books={refs.books} />
    {confirmationDialog}
  </section>;
}
