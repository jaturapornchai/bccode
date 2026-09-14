"use client";

import { useMemo, useState } from "react";
import { Eye, FileText, Pencil, Plus, RefreshCw, Save, Search, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { amountString, bookLabels, labelText, type GLLabel, emptyJournal, emptyLine, formatAmount, journalTotals, localDate, validateJournal, type GLJournal, type GLLine } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { AccountSelect, AmountInput, Field, Notice, Pager, SearchInput, SplitWorkbench, YearSelect, actionClass, control, useDebouncedSearch, useDirtyGuard, useGLCommand, useGLList, useReferences, useRowDensity, useGLText } from "./gl-common";

const statusLabel: Record<string, GLLabel> = { draft: ["gl_draft", "ฉบับร่าง"], posted: ["gl_posted", "ผ่านรายการแล้ว"], reversed: ["gl_reversed", "กลับรายการแล้ว"] };

export function GLJournals({ route, book = "", kind = "", mode = "edit" }: { route: string; book?: string; kind?: string; mode?: "edit" | "post" | "reverse" }) {
  const tr = useGLText();
  const [search, setSearch] = useState("");
  const filters = new URLSearchParams({ ...(book ? { bookcode: book } : {}), ...(kind ? { kind } : {}), ...(mode === "post" ? { status: "draft" } : mode === "reverse" ? { status: "posted" } : {}) }).toString();
  const list = useGLList<GLJournal>("journals", search, filters), refs = useReferences();
  const searchDebounce = useDebouncedSearch({
    onSearch: (val) => {
      list.setPage(1);
      setSearch(val);
    },
    debounceMs: 2000,
  });
  const density = useRowDensity();
  const [journal, setJournal] = useState<GLJournal | null>(null), [original, setOriginal] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [message, setMessage] = useState(""), [error, setError] = useState("");
  const [reason, setReason] = useState(""), [reverseDate, setReverseDate] = useState(localDate()), [reverseDocno, setReverseDocno] = useState("");
  const { busy, execute } = useGLCommand(), { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const dirty = isEditing && journal !== null && JSON.stringify(journal) !== original;

  useDirtyGuard(route, dirty);
  const year = refs.years.find((item) => item.code === journal?.fiscalyear);
  const totals = useMemo(() => { try { return journal ? journalTotals(journal.lines) : null; } catch { return null; } }, [journal]);
  const patch = (value: Partial<GLJournal>) => setJournal((current) => current ? { ...current, ...value } : current);
  const patchLine = (index: number, value: Partial<GLLine>) => patch({ lines: journal!.lines.map((line, i) => i === index ? { ...line, ...value } : line) });

  async function openView(item: GLJournal) {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_entries", "ละทิ้งรายการที่ยังไม่บันทึก?"), description: tr("gl_unsaved_entries_not_saved", "รายการที่กรอกอยู่จะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    try {
      const value = item?.id ? await glRequest<GLJournal>(`journals/${encodeURIComponent(item.id)}`) : emptyJournal(book || "JV", kind || "manual");
      setJournal(value);
      setOriginal(JSON.stringify(value));
      setIsEditing(false);
      setError("");
      setMessage("");
      setReason("");
      setReverseDocno("");
    } catch (e) {
      setError((e as Error).message);
    }
  }

  async function openEdit(item: GLJournal) {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_entries", "ละทิ้งรายการที่ยังไม่บันทึก?"), description: tr("gl_unsaved_entries_not_saved", "รายการที่กรอกอยู่จะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    try {
      const value = item?.id ? await glRequest<GLJournal>(`journals/${encodeURIComponent(item.id)}`) : emptyJournal(book || "JV", kind || "manual");
      setJournal(value);
      setOriginal(JSON.stringify(value));
      setIsEditing(true);
      setError("");
      setMessage("");
      setReason("");
      setReverseDocno("");
    } catch (e) {
      setError((e as Error).message);
    }
  }

  async function openCreate() {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_entries", "ละทิ้งรายการที่ยังไม่บันทึก?"), description: tr("gl_unsaved_entries_not_saved", "รายการที่กรอกอยู่จะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    const value = emptyJournal(book || "JV", kind || "manual");
    setJournal(value);
    setOriginal(JSON.stringify(value));
    setIsEditing(true);
    setError("");
    setMessage("");
    setReason("");
    setReverseDocno("");
  }

  async function cancelEdit() {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_entries", "ละทิ้งรายการที่ยังไม่บันทึก?"), description: tr("gl_unsaved_entries_not_saved", "รายการที่กรอกอยู่จะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    if (journal?.id) {
      setJournal(JSON.parse(original));
      setIsEditing(false);
    } else {
      setJournal(null);
      setOriginal("");
      setIsEditing(false);
    }
    setError("");
    setReason("");
  }

  async function save() {
    if (!journal || busy || !isEditing) return;
    const problem = validateJournal(journal, year, refs.accounts, tr);
    if (problem) { setError(problem); return; }
    if (journal.id && !await confirm({ title: tr("gl_save_draft_changes", "บันทึกการแก้ไขฉบับร่าง?"), description: journal.docno, confirmLabel: tr("gl_save_draft", "บันทึกฉบับร่าง"), tone: "info" })) return;
    try {
      const result = await execute({ resource: "journals", id: journal.id, version: journal.version, action: journal.id ? "update" : "create", journal });
      const saved = { ...journal, id: result.id, version: result.version };
      setJournal(saved);
      setOriginal(JSON.stringify(saved));
      setIsEditing(false);
      list.reload();
      setError("");
      setMessage(result.projectionpending ? tr("gl_draft_saved_updating_report_data", "บันทึกฉบับร่างแล้ว กำลังปรับปรุงข้อมูลสำหรับรายงาน") : tr("gl_draft_saved_review_and_post", "บันทึกฉบับร่างแล้ว ตรวจสอบและกดผ่านรายการเมื่อพร้อม"));
    } catch (e) { setError((e as Error).message); }
  }

  async function act(action: "post" | "reverse" | "delete") {
    if (!journal?.id || busy || dirty) return;
    if (action === "post") { const problem = validateJournal(journal, year, refs.accounts, tr); if (problem) { setError(problem); return; } }
    if (action !== "post" && !reason.trim()) { setError(tr("gl_enter_reason_before_posting", "กรุณาระบุเหตุผลก่อนทำรายการ")); return; }
    if (action === "reverse" && (!reverseDate || !reverseDocno.trim())) { setError(tr("gl_enter_date_and_reversal_doc_no", "กรุณาระบุวันที่และเลขที่ใบกลับรายการ")); return; }
    const label = action === "post" ? tr("gl_post_accounting_entry", "ผ่านรายการบัญชี") : action === "reverse" ? tr("gl_create_reversing_entry", "สร้างรายการกลับบัญชี") : tr("gl_delete_draft", "ลบฉบับร่าง");
    if (!await confirm({ title: `${label}?`, description: action === "post" ? tr("gl_after_posting_edit_requires_reversal", "หลังผ่านรายการจะไม่สามารถแก้ไขหรือลบได้ การแก้ไขต้องสร้างรายการกลับบัญชีพร้อมเหตุผล") : action === "reverse" ? tr("gl_create_reversal_doc", "สร้างเอกสาร {0} วันที่ {1} กลับเดบิตและเครดิตของ {2} โดยเก็บรายการเดิมไว้").replace("{0}", String(reverseDocno)).replace("{1}", String(reverseDate)).replace("{2}", String(journal.docno)) : tr("gl_delete_draft_keep_history", "ลบฉบับร่าง {0} พร้อมเก็บประวัติ").replace("{0}", String(journal.docno)), details: tr("gl_debit_credit", "เดบิต {0} · เครดิต {1}").replace("{0}", String(totals ? formatAmount(amountString(totals.debit), year?.scale) : "—")).replace("{1}", String(totals ? formatAmount(amountString(totals.credit), year?.scale) : "—")), confirmLabel: label, tone: action === "delete" ? "danger" : "warning" })) return;
    try {
      await execute({ resource: "journals", action, id: journal.id, version: journal.version, reason, ...(action === "reverse" ? { date: reverseDate, docno: reverseDocno } : {}) });
      setJournal(null);
      setOriginal("");
      setIsEditing(false);
      list.reload();
      setError("");
      setMessage(tr("gl_success_message", "{0}เรียบร้อยแล้ว").replace("{0}", String(label)));
    } catch (e) { setError((e as Error).message); }
  }

  async function deleteDraftDirect(item: GLJournal) {
    if (!item.id || busy || item.status !== "draft") return;
    if (!await confirm({
      title: tr("gl_delete_draft_confirm", "ลบฉบับร่าง?"),
      description: tr("gl_delete_draft_keep_history", "ลบฉบับร่าง {0} พร้อมเก็บประวัติ").replace("{0}", String(item.docno)),
      confirmLabel: tr("gl_delete_draft", "ลบฉบับร่าง"),
      tone: "danger",
    })) return;
    try {
      await execute({ resource: "journals", action: "delete", id: item.id, version: item.version, reason: tr("gl_delete_draft", "ลบฉบับร่าง") });
      if (journal?.id === item.id) {
        setJournal(null);
        setOriginal("");
        setIsEditing(false);
      }
      list.reload();
      setError("");
      setMessage(tr("gl_delete_draft_success", "ลบฉบับร่าง {0} เรียบร้อยแล้ว").replace("{0}", String(item.docno)));
    } catch (e) {
      setError((e as Error).message);
    }
  }

  return (
    <div className="flex flex-col flex-1 min-h-0 gap-2">
      <div className="shrink-0 flex flex-col gap-2">
        <Notice error text={error || list.error || refs.error} />
        <Notice text={message} />
        {kind === "opening" && <Notice text={tr("gl_opening_balance_dr_cr_entries", "บันทึกยอดยกมาเป็นรายการเดบิตและเครดิตที่สมดุล เลือกปีบัญชี วันที่ และบัญชีคู่รายการตามยอดปิดที่ตรวจสอบแล้ว")} />}
      </div>
      <SplitWorkbench
        list={
          <div className="flex flex-col flex-1 min-h-0 gap-2">
            <form className="flex flex-wrap gap-2 shrink-0" onSubmit={(event) => { event.preventDefault(); searchDebounce.searchNow(); }}>
              <SearchInput
                className="min-w-28 flex-1"
                ariaLabel={tr("gl_search_number_or_desc", "ค้นหาเลขที่หรือคำอธิบาย")}
                placeholder={tr("gl_number_or_desc", "เลขที่หรือคำอธิบาย")}
                value={searchDebounce.query}
                onChange={searchDebounce.setQuery}
                onClear={searchDebounce.clear}
                onSearch={searchDebounce.searchNow}
              />
              <Button type="submit" variant="outline" className={actionClass}><Search className="size-4 mr-1.5" />{tr("gl_search", "ค้นหา")}</Button>
              <Button type="button" variant="outline" className={actionClass} onClick={() => { list.reload(); refs.reload(); }} disabled={list.loading}><RefreshCw className="size-4 mr-1.5" />{tr("gl_reload", "โหลดใหม่")}</Button>
              {mode === "edit" && <Button type="button" className={actionClass} disabled={busy} onClick={() => void openCreate()}><Plus className="size-4 mr-1.5" />{tr("gl_add_row", "เพิ่มรายการ")}</Button>}
              <Button type="button" variant="outline" className={actionClass} aria-pressed={density.compact} onClick={density.toggle}>{density.compact ? tr("gl_expand_row", "ขยายบรรทัด") : tr("gl_collapse_row", "ย่อบรรทัด")}</Button>
            </form>
            <div className="flex items-center justify-between px-1 text-xs text-muted-foreground shrink-0">
              <span>{tr("gl_x_items", "{0} รายการ").replace("{0}", String(list.data.total.toLocaleString("th-TH")))}</span>
              {density.compact && <span className="text-[11px]">{tr("gl_collapse_row_mode", "โหมดย่อบรรทัด")}</span>}
            </div>
            <div className="flex-1 min-h-[300px] overflow-auto rounded-xl border border-border" aria-busy={list.loading}>
              <table className={`w-full text-left text-[0.95rem] leading-normal ${density.tableClass}`}>
                <thead className="sticky top-0 bg-muted z-10">
                  <tr>
                    <th className="p-2.5">{tr("gl_date_number", "วันที่ / เลขที่")}</th>
                    <th className="p-2.5 min-w-44">{tr("gl_description", "คำอธิบาย")}</th>
                    <th className="p-2.5 text-center w-28">{tr("gl_status", "สถานะ")}</th>
                    <th className="p-2.5 text-right pr-3 w-24">{tr("gl_manage", "จัดการ")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {list.data.items.map((item, index) => {
                    const isSelected = journal?.id === item.id;
                    const isRowEditing = isSelected && isEditing;
                    const isDraft = item.status === "draft";
                    return (
                      <tr
                        key={item.id}
                        className={`group cursor-pointer transition-colors ${
                          isRowEditing
                            ? "bg-primary/15 hover:bg-primary/20 text-foreground ring-1 ring-inset ring-primary/50 font-medium"
                            : isSelected
                              ? "bg-primary/10 ring-1 ring-inset ring-primary/40 font-medium"
                              : index % 2 === 0
                                ? "bg-background hover:bg-accent/60"
                                : "bg-muted/20 hover:bg-accent/60"
                        }`}
                        onClick={() => void openView(item)}
                      >
                        <td className="whitespace-nowrap p-2">
                          <div className="text-[0.85rem] text-muted-foreground">{item.date}</div>
                          <span className="font-mono font-bold text-primary">{item.docno}</span>
                        </td>
                        <td className="max-w-60 truncate p-2" title={item.description}>{item.description}</td>
                        <td className="whitespace-nowrap p-2 text-center">
                          <span className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ${
                            isDraft
                              ? "bg-muted text-muted-foreground border border-border"
                              : item.status === "posted"
                                ? "bg-primary/10 text-primary border border-primary/20"
                                : "bg-muted text-muted-foreground border border-border"
                          }`}>
                            {labelText(statusLabel, item.status, tr, tr("gl_check_status", "ตรวจสอบสถานะ"))}
                          </span>
                        </td>
                        <td className="p-2 text-right whitespace-nowrap pr-2" onClick={(e) => e.stopPropagation()}>
                          <div className="flex items-center justify-end gap-1">
                            {isDraft && mode === "edit" ? (
                              <>
                                <Button
                                  type="button"
                                  size="icon"
                                  variant="outline"
                                  className="size-7 rounded-md bg-background text-primary border-primary/30 hover:bg-primary/15 hover:border-primary/50 shadow-none transition-colors shrink-0"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    void openEdit(item);
                                  }}
                                  aria-label={tr("gl_edit", "แก้ไข")}
                                  title={tr("gl_edit_draft_2", "แก้ไขฉบับร่าง (Edit)")}
                                >
                                  <Pencil className="size-3.5 shrink-0" />
                                </Button>
                                <Button
                                  type="button"
                                  size="icon"
                                  variant="outline"
                                  className="size-7 rounded-md bg-background text-destructive border-destructive/30 hover:bg-destructive/10 hover:border-destructive/50 shadow-none transition-colors shrink-0"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    void deleteDraftDirect(item);
                                  }}
                                  aria-label={tr("gl_delete", "ลบ")}
                                  title={tr("gl_delete_draft_2", "ลบฉบับร่าง (Delete)")}
                                >
                                  <Trash2 className="size-3.5 shrink-0" />
                                </Button>
                              </>
                            ) : (
                              <Button
                                type="button"
                                size="icon"
                                variant="outline"
                                className="size-7 rounded-md bg-background text-muted-foreground hover:text-foreground border-border shadow-none transition-colors shrink-0"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  void openView(item);
                                }}
                                aria-label={tr("gl_view_data", "แสดงข้อมูล")}
                                title={tr("gl_view", "แสดงข้อมูล (View)")}
                              >
                                <Eye className="size-3.5 shrink-0" />
                              </Button>
                            )}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                  {!list.data.items.length && (
                    <tr>
                      <td colSpan={4} className="p-6 text-center text-muted-foreground">
                        {list.loading ? tr("gl_loading_data", "กำลังโหลดข้อมูล…") : tr("gl_no_entries_in_journal", "ยังไม่มีรายการในสมุดนี้")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
            <Pager page={list.page} total={list.data.total} onPage={list.setPage} loading={list.loading} />
          </div>
        }
        editor={
          journal ? (
            !isEditing ? (
              <div className="flex flex-col h-full min-h-0 gap-3">
                <header className="flex flex-wrap items-center justify-between gap-2 border-b border-border pb-3 shrink-0">
                  <div className="flex items-center gap-2">
                    <div className="grid size-8 place-items-center rounded-lg bg-primary/10 text-primary">
                      <Eye className="size-4" />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <h2 className="text-base font-semibold">{tr("gl_show_data", "แสดงข้อมูล: {0}").replace("{0}", String(journal.docno))}</h2>
                        <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-semibold bg-muted text-muted-foreground border border-border">
                          {labelText(statusLabel, journal.status, tr)} · {tr("gl_display_mode", tr("gl_display_mode", "โหมดแสดงข้อมูล"))}
                        </span>
                      </div>
                      <p className="text-xs text-muted-foreground">{journal.description}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-1.5">
                    {journal.status === "draft" && mode === "edit" && (
                      <Button
                        type="button"
                        className={actionClass}
                        onClick={() => setIsEditing(true)}
                        disabled={busy}
                        title={tr("gl_edit_this_draft", "แก้ไขฉบับร่างนี้ (Edit)")}
                      >
                        <Pencil className="size-4 mr-1.5" />{tr("gl_edit_draft_3", "แก้ไขฉบับร่าง")}
                      </Button>
                    )}
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="size-8 rounded-lg text-muted-foreground hover:text-foreground"
                      onClick={() => { setJournal(null); setOriginal(""); setIsEditing(false); }}
                      aria-label={tr("gl_close_view_dialog", "ปิดหน้าต่างแสดงข้อมูล")}
                      title={tr("gl_close_view_dialog", "ปิดหน้าต่างแสดงข้อมูล")}
                    >
                      <X className="size-4" />
                    </Button>
                  </div>
                </header>
                <div className="flex items-center gap-2 rounded-lg border border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground shrink-0">
                  <Eye className="size-3.5 text-primary shrink-0" />
                  <span>
                    {tr("gl_read_only_mode", "โหมดแสดงข้อมูล (Read-only) — {0}").replace("{0}", String(journal.status === "draft" ? tr("gl_edit_draft_instruction", "หากต้องการแก้ไข ให้กดปุ่ม \"แก้ไขฉบับร่าง\"") : tr("gl_posted_cannot_edit_directly", "เอกสารนี้ผ่านรายการแล้ว ไม่สามารถแก้ไขได้โดยตรง")))}
                  </span>
                </div>
                <div className="overflow-y-auto p-1 grid min-w-0 gap-3 content-start">
                  <fieldset disabled className="grid min-w-0 gap-3 opacity-95">
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label={tr("gl_document_no", "เลขที่เอกสาร")}><input className={control} value={journal.docno} readOnly /></Field>
                      <Field label={tr("gl_document_date", "วันที่เอกสาร")}><input className={control} type="date" value={journal.date} readOnly /></Field>
                      <Field label={tr("gl_fiscal_year", "ปีบัญชี")}><input className={control} value={journal.fiscalyear} readOnly /></Field>
                      <Field label={tr("gl_journal", "สมุดรายวัน")}><input className={control} value={labelText(bookLabels, journal.bookcode, tr)} readOnly /></Field>
                      <Field label={tr("gl_entry_description", "คำอธิบายรายการ")}><input className={control} value={journal.description} readOnly /></Field>
                      <Field label={tr("gl_reference_document", "เอกสารอ้างอิง")}><input className={control} value={journal.reference || "-"} readOnly /></Field>
                      <Field label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} value={journal.branchcode || "-"} readOnly /></Field>
                    </div>
                    <div className="overflow-x-auto rounded-xl border border-border">
                      <table className="w-full min-w-[760px] text-left text-[0.95rem]">
                        <thead className="bg-muted">
                          <tr>
                            <th className="p-2">{tr("gl_account_description", "บัญชี / คำอธิบาย")}</th>
                            <th className="p-2 text-right">{tr("gl_debit", "เดบิต")}</th>
                            <th className="p-2 text-right">{tr("gl_credit", "เครดิต")}</th>
                            <th className="p-2">{tr("gl_department_project", "แผนก / โครงการ")}</th>
                            <th className="p-2">{tr("gl_cash_flow", "กระแสเงินสด")}</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                          {journal.lines.map((line, index) => (
                            <tr key={index}>
                              <td className="min-w-56 p-2">
                                <div className="font-mono font-semibold text-primary">{line.accountcode}</div>
                                {line.description && <div className="text-xs text-muted-foreground">{line.description}</div>}
                              </td>
                              <td className="min-w-32 p-2 text-right font-mono tabular-nums whitespace-nowrap">
                                {line.debit ? formatAmount(line.debit, year?.scale) : "—"}
                              </td>
                              <td className="min-w-32 p-2 text-right font-mono tabular-nums whitespace-nowrap">
                                {line.credit ? formatAmount(line.credit, year?.scale) : "—"}
                              </td>
                              <td className="min-w-32 p-2">
                                <span className="text-sm">{line.departmentcode || line.projectcode ? `${line.departmentcode || "-"}${line.projectcode ? ` / ${line.projectcode}` : ""}` : "—"}</span>
                              </td>
                              <td className="min-w-36 p-2">
                                <span className="text-xs text-muted-foreground">{line.cashflow || tr("gl_not_specified_2", "ไม่ระบุ")}</span>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </fieldset>
                  <div className="grid gap-2 rounded-xl border border-primary/20 bg-primary/5 p-3 sm:grid-cols-3" aria-live="polite">
                    {([ [tr("gl_total_debit", "เดบิตรวม"), totals?.debit], [tr("gl_total_credit", "เครดิตรวม"), totals?.credit], [tr("gl_difference", "ผลต่าง"), totals?.difference] ] as const).map(([label, units]) => (
                      <div key={label}>
                        <div className="text-[0.9rem] text-muted-foreground">{label}</div>
                        <strong className="text-lg tabular-nums">{units === undefined ? tr("gl_verify_amount", "ตรวจจำนวนเงิน") : formatAmount(amountString(units), year?.scale)}</strong>
                      </div>
                    ))}
                  </div>
                  {journal.id && journal.status === "draft" && (
                    <Field label={tr("gl_reason", "เหตุผล")}><input className={control} value={reason} disabled={busy} onChange={(e) => setReason(e.target.value)} placeholder={tr("gl_required_for_draft_deletion", "จำเป็นสำหรับลบฉบับร่าง")} /></Field>
                  )}
                  {journal.status === "posted" && (
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label={tr("gl_reversal_document_number", "เลขที่ใบกลับรายการ")}><input className={control} value={reverseDocno} disabled={busy} onChange={(e) => setReverseDocno(e.target.value)} placeholder={tr("gl_enter_reversal_document_number", "ระบุเลขที่ใบกลับรายการ")} /></Field>
                      <Field label={tr("gl_reversal_date", "วันที่กลับรายการ")}><input className={control} type="date" value={reverseDate} disabled={busy} onChange={(e) => setReverseDate(e.target.value)} /></Field>
                    </div>
                  )}
                  {journal.status === "posted" && (
                    <Field label={tr("gl_reversal_reason", "เหตุผลการกลับรายการ")}><input className={control} value={reason} disabled={busy} onChange={(e) => setReason(e.target.value)} placeholder={tr("gl_required_for_reversal_entry", "จำเป็นสำหรับสร้างรายการกลับบัญชี")} /></Field>
                  )}
                </div>
                <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-3 shrink-0 mt-auto">
                  <div className="flex flex-wrap gap-2">
                    {journal.status === "draft" && mode === "edit" && (
                      <Button
                        type="button"
                        className={actionClass}
                        onClick={() => setIsEditing(true)}
                        disabled={busy}
                      >
                        <Pencil className="size-4 mr-1.5" />{tr("gl_edit_draft_3", "แก้ไขฉบับร่าง")}
                      </Button>
                    )}
                    {journal.status === "draft" && (
                      <Button
                        type="button"
                        variant="outline"
                        className={actionClass}
                        disabled={busy}
                        onClick={() => void act("post")}
                      >
                        {tr("gl_post_accounting_entry", "ผ่านรายการบัญชี")}
                      </Button>
                    )}
                    {journal.status === "posted" && (
                      <Button
                        type="button"
                        className={actionClass}
                        disabled={busy}
                        onClick={() => void act("reverse")}
                      >
                        {tr("gl_create_reversing_entry", "สร้างรายการกลับบัญชี")}
                      </Button>
                    )}
                    <Button
                      type="button"
                      variant="outline"
                      className={actionClass}
                      onClick={() => { setJournal(null); setOriginal(""); setIsEditing(false); }}
                    >
                      {tr("gl_close", "ปิด")}
                    </Button>
                  </div>
                  {journal.status === "draft" && journal.id && (
                    <Button
                      type="button"
                      variant="outline"
                      className={`${actionClass} text-destructive border-destructive/30 hover:bg-destructive/10 hover:border-destructive/50`}
                      disabled={busy}
                      onClick={() => void act("delete")}
                    >
                      <Trash2 className="size-4 mr-1.5" />{tr("gl_delete_draft", "ลบฉบับร่าง")}
                    </Button>
                  )}
                </footer>
              </div>
            ) : (
              <form className="flex flex-col h-full min-h-0 gap-3" onSubmit={(event) => { event.preventDefault(); void save(); }}>
                <header className="flex flex-wrap items-center justify-between gap-2 border-b border-border pb-3 shrink-0">
                  <div className="flex items-center gap-2">
                    <div className="grid size-8 place-items-center rounded-lg bg-primary/10 text-primary">
                      {journal.id ? <Pencil className="size-4" /> : <Plus className="size-4" />}
                    </div>
                    <div>
                      <h2 className="text-base font-semibold">{journal.id ? tr("gl_edit_draft", "แก้ไขฉบับร่าง {0}").replace("{0}", String(journal.docno)) : tr("gl_save_new_journal", "บันทึกรายวันใหม่")}</h2>
                      {dirty && <span className="text-xs text-primary font-medium">{tr("gl_unsaved_changes", "● มีการเปลี่ยนแปลงที่ยังไม่บันทึก")}</span>}
                    </div>
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="size-8 rounded-lg text-muted-foreground hover:text-foreground"
                    onClick={() => void cancelEdit()}
                    aria-label={tr("gl_close_edit_dialog", "ปิดหน้าต่างแก้ไข")}
                    title={tr("gl_close_edit_dialog", "ปิดหน้าต่างแก้ไข")}
                  >
                    <X className="size-4" />
                  </Button>
                </header>
                <div className="overflow-y-auto p-1 grid min-w-0 gap-3 content-start">
                  <fieldset disabled={busy} className="grid min-w-0 gap-3">
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label={tr("gl_document_no", "เลขที่เอกสาร")}><input className={control} required value={journal.docno} maxLength={60} onChange={(e) => patch({ docno: e.target.value })} /></Field>
                      <Field label={tr("gl_document_date", "วันที่เอกสาร")}><input className={control} required type="date" value={journal.date} onChange={(e) => patch({ date: e.target.value })} /></Field>
                      <Field label={tr("gl_fiscal_year", "ปีบัญชี")}><YearSelect years={refs.years} value={journal.fiscalyear} onChange={(fiscalyear) => patch({ fiscalyear })} /></Field>
                      <Field label={tr("gl_journal", "สมุดรายวัน")}><select className={control} disabled={!!book} value={journal.bookcode} onChange={(e) => patch({ bookcode: e.target.value })}>{Object.entries(bookLabels).map(([code, name]) => <option key={code} value={code}>{tr(...name)}</option>)}</select></Field>
                      <Field label={tr("gl_entry_description", "คำอธิบายรายการ")}><input className={control} required value={journal.description} onChange={(e) => patch({ description: e.target.value })} maxLength={500} /></Field>
                      <Field label={tr("gl_reference_document", "เอกสารอ้างอิง")}><input className={control} value={journal.reference} onChange={(e) => patch({ reference: e.target.value })} /></Field>
                      <Field label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} value={journal.branchcode} onChange={(e) => patch({ branchcode: e.target.value })} /></Field>
                    </div>
                    <div className="overflow-x-auto rounded-xl border border-border">
                      <table className="w-full min-w-[760px] text-left text-[0.95rem]">
                        <thead className="bg-muted">
                          <tr>
                            <th className="p-2">{tr("gl_account_description", "บัญชี / คำอธิบาย")}</th>
                            <th className="p-2">{tr("gl_debit", "เดบิต")}</th>
                            <th className="p-2">{tr("gl_credit", "เครดิต")}</th>
                            <th className="p-2">{tr("gl_department_project", "แผนก / โครงการ")}</th>
                            <th className="p-2">{tr("gl_cash_flow", "กระแสเงินสด")}</th>
                            <th className="p-2">{tr("gl_manage", "จัดการ")}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {journal.lines.map((line, index) => (
                            <tr key={index} className="border-t border-border">
                              <td className="min-w-56 p-2">
                                <div className="grid gap-1">
                                  <AccountSelect label={tr("gl_line_account", "บัญชีบรรทัด {0}").replace("{0}", String(index + 1))} value={line.accountcode} accounts={refs.accounts} onChange={(accountcode) => patchLine(index, { accountcode })} disabled={busy} />
                                  <input className={control} aria-label={tr("gl_line_description", "คำอธิบายบรรทัด {0}").replace("{0}", String(index + 1))} placeholder={tr("gl_description", "คำอธิบาย")} value={line.description} onChange={(e) => patchLine(index, { description: e.target.value })} />
                                </div>
                              </td>
                              <td className="min-w-32 p-2">
                                <AmountInput ariaLabel={tr("gl_line_debit", "เดบิตบรรทัด {0}").replace("{0}", String(index + 1))} disabled={busy} scale={year?.scale ?? 2} value={line.debit} onChange={(debit) => patchLine(index, { debit })} />
                              </td>
                              <td className="min-w-32 p-2">
                                <AmountInput ariaLabel={tr("gl_line_credit", "เครดิตบรรทัด {0}").replace("{0}", String(index + 1))} disabled={busy} scale={year?.scale ?? 2} value={line.credit} onChange={(credit) => patchLine(index, { credit })} />
                              </td>
                              <td className="min-w-32 p-2">
                                <div className="grid gap-1">
                                  <input className={control} aria-label={tr("gl_line_department", "แผนกบรรทัด {0}").replace("{0}", String(index + 1))} placeholder={tr("gl_department", "แผนก")} value={line.departmentcode} onChange={(e) => patchLine(index, { departmentcode: e.target.value })} />
                                  <input className={control} aria-label={tr("gl_line_project", "โครงการบรรทัด {0}").replace("{0}", String(index + 1))} placeholder={tr("gl_project", "โครงการ")} value={line.projectcode} onChange={(e) => patchLine(index, { projectcode: e.target.value })} />
                                </div>
                              </td>
                              <td className="min-w-36 p-2">
                                <select className={control} aria-label={tr("gl_cash_flow_line", "กระแสเงินสดบรรทัด {0}").replace("{0}", String(index + 1))} value={line.cashflow} onChange={(e) => patchLine(index, { cashflow: e.target.value })}>
                                  <option value="">{tr("gl_not_specified_2", "ไม่ระบุ")}</option>
                                  <option value="operating">{tr("gl_operating", "ดำเนินงาน")}</option>
                                  <option value="investing">{tr("gl_investing", "ลงทุน")}</option>
                                  <option value="financing">{tr("gl_raise_funds", "จัดหาเงิน")}</option>
                                </select>
                              </td>
                              <td className="p-2">
                                <Button type="button" variant="outline" className={actionClass} disabled={busy || journal.lines.length <= 2} onClick={() => patch({ lines: journal.lines.filter((_, i) => i !== index) })}>
                                  {tr("gl_remove", "นำออก")}
                                </Button>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                    <Button type="button" variant="outline" className={actionClass} disabled={journal.lines.length >= 500} onClick={() => patch({ lines: [...journal.lines, emptyLine()] })}>
                      <Plus className="size-4 mr-1.5" />{tr("gl_add_line", "เพิ่มบรรทัด")}
                    </Button>
                  </fieldset>
                  <div className="grid gap-2 rounded-xl border border-primary/20 bg-primary/5 p-3 sm:grid-cols-3" aria-live="polite">
                    {([ [tr("gl_total_debit", "เดบิตรวม"), totals?.debit], [tr("gl_total_credit", "เครดิตรวม"), totals?.credit], [tr("gl_difference", "ผลต่าง"), totals?.difference] ] as const).map(([label, units]) => (
                      <div key={label}>
                        <div className="text-[0.9rem] text-muted-foreground">{label}</div>
                        <strong className="text-lg tabular-nums">{units === undefined ? tr("gl_verify_amount", "ตรวจจำนวนเงิน") : formatAmount(amountString(units), year?.scale)}</strong>
                      </div>
                    ))}
                  </div>
                  {journal.id && (
                    <Field label={tr("gl_reason", "เหตุผล")}><input className={control} value={reason} disabled={busy} onChange={(e) => setReason(e.target.value)} placeholder={tr("gl_edit_reason_hint", "ระบุเหตุผลการแก้ไข (ถ้ามี)")} /></Field>
                  )}
                </div>
                <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-3 shrink-0 mt-auto">
                  <div className="flex flex-wrap gap-2">
                    <Button type="submit" className={actionClass} disabled={busy || !dirty}>
                      <Save className="size-4 mr-1.5" />{busy ? tr("gl_saving", "กำลังบันทึก…") : tr("gl_save_draft", "บันทึกฉบับร่าง")}
                    </Button>
                    <Button type="button" variant="outline" className={actionClass} onClick={() => void cancelEdit()}>
                      {tr("gl_cancel", "ยกเลิก")}
                    </Button>
                  </div>
                  {journal.id && (
                    <Button
                      type="button"
                      variant="outline"
                      className={`${actionClass} text-destructive border-destructive/30 hover:bg-destructive/10 hover:border-destructive/50`}
                      disabled={busy}
                      onClick={() => void act("delete")}
                    >
                      <Trash2 className="size-4 mr-1.5" />{tr("gl_delete_draft", "ลบฉบับร่าง")}
                    </Button>
                  )}
                </footer>
              </form>
            )
          ) : (
            <div className="flex flex-col flex-1 h-full min-h-64 items-center justify-center content-center gap-3 text-center p-6">
              <div className="mx-auto grid size-12 place-items-center rounded-2xl bg-muted text-muted-foreground">
                <FileText className="size-6" />
              </div>
              <h2 className="text-base font-semibold">{tr("gl_select_item_show_acct", "เลือกรายการเพื่อแสดงข้อมูลบัญชี")}</h2>
              <p className="text-sm text-muted-foreground">{tr("gl_click_row_or_add_journal", "คลิกที่แถวในตารางเพื่อแสดงข้อมูล หรือกดปุ่ม &ldquo;+ เพิ่มรายการ&rdquo; เพื่อบันทึกรายวันใหม่")}</p>
              {mode === "edit" && (
                <div className="pt-2">
                  <Button type="button" className={actionClass} onClick={() => void openCreate()} disabled={busy}>
                    <Plus className="size-4 mr-1.5" />{tr("gl_save_new_journal", "บันทึกรายวันใหม่")}
                  </Button>
                </div>
              )}
            </div>
          )
        }
      />
      {confirmationDialog}
    </div>
  );
}
