"use client";

import { useMemo, useState } from "react";
import { Eye, FileText, Pencil, Plus, RefreshCw, Save, Search, Trash2, X, ClipboardPaste, Scale, Sparkles, CheckCircle2, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { amountString, bookLabels, labelText, type GLLabel, emptyJournal, emptyLine, formatAmount, journalTotals, localDate, validateJournal, type GLJournal, type GLLine } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { AccountSelect, AmountInput, Combobox, Field, Notice, Pager, SearchInput, SplitWorkbench, UnsavedBadge, YearSelect, actionClass, control, useDebouncedSearch, useDirtyGuard, useGLCommand, useGLList, useReferences, useRowDensity, useGLText } from "./gl-common";
import { useFormShortcuts } from "@/hooks/use-form-shortcuts";
import { parseClipboardJournalLines } from "@/lib/clipboard-journal-parser";
import { useTabularEnterNav } from "@/hooks/use-tabular-enter-nav";
import { analyzeGLTaxAndBalance, autoBalanceJournalLines, setExactVatLine, appendVatLine } from "@/lib/gl-smart-guard";

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
  const isJournalDirty = journal !== null && JSON.stringify(journal) !== original;
  const isReasonDirty = reason.trim() !== "";
  const dirty = isEditing && journal !== null && (isJournalDirty || isReasonDirty);

  useDirtyGuard(route, dirty);
  const year = refs.years.find((item) => item.code === journal?.fiscalyear);
  const totals = useMemo(() => { try { return journal ? journalTotals(journal.lines) : null; } catch { return null; } }, [journal]);
  const smartGuard = useMemo(() => {
    try {
      return journal ? analyzeGLTaxAndBalance(journal.lines, refs.accounts, year?.scale ?? 2) : null;
    } catch {
      return null;
    }
  }, [journal, refs.accounts, year?.scale]);
  const patch = (value: Partial<GLJournal>) => setJournal((current) => current ? { ...current, ...value } : current);
  const patchLine = (index: number, value: Partial<GLLine>) => patch({ lines: journal!.lines.map((line, i) => i === index ? { ...line, ...value } : line) });

  const tabularEnterNav = useTabularEnterNav({
    onAddNewRow: () => {
      if (journal && journal.lines.length < 500) {
        patch({ lines: [...journal.lines, emptyLine()] });
      }
    },
  });

  const applyPastedLines = (clipboardText: string) => {
    if (!journal || !clipboardText.trim()) return;
    const parsed = parseClipboardJournalLines(clipboardText);
    if (parsed.length === 0) return;

    const isEmptyJournal = journal.lines.every(
      (l) => !l.accountcode && !l.description && (!l.debit || l.debit === "0") && (!l.credit || l.credit === "0")
    );
    const combined = isEmptyJournal ? parsed : [...journal.lines, ...parsed];
    const finalLines = combined.slice(0, 500);
    patch({ lines: finalLines });
    setMessage(`Pasted ${parsed.length} row(s)`);
  };

  const handlePasteClick = async () => {
    try {
      if (!navigator?.clipboard?.readText) {
        setError("Clipboard API not supported in this browser");
        return;
      }
      const text = await navigator.clipboard.readText();
      if (text) {
        applyPastedLines(text);
      }
    } catch {
      setError("Clipboard read permission denied");
    }
  };

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

  // Global Keyboard Shortcuts (Ctrl+S, Alt+N, Esc)
  useFormShortcuts({
    onSave: () => void save(),
    onNew: mode === "edit" ? () => void openCreate() : undefined,
    onCancel: isEditing ? () => void cancelEdit() : undefined,
    canSave: isEditing && dirty && !busy,
    disabled: busy,
  });

  async function act(action: "post" | "reverse" | "delete") {
    if (!journal?.id || busy || dirty) return;
    if (action === "post") { const problem = validateJournal(journal, year, refs.accounts, tr); if (problem) { setError(problem); return; } }
    if (action !== "post" && !reason.trim()) { setError(tr("gl_enter_reason_before_posting", "กรุณาระบุเหตุผลก่อนทำรายการ")); return; }
    if (action === "reverse" && (!reverseDate || !reverseDocno.trim())) { setError(tr("gl_enter_date_and_reversal_doc_no", "กรุณาระบุวันที่และเลขที่ใบกลับรายการ")); return; }
    const label = action === "post" ? tr("gl_post_accounting_entry", "ผ่านรายการ") : action === "reverse" ? tr("gl_create_reversing_entry", "สร้างรายการกลับบัญชี") : tr("gl_delete_draft", "ลบฉบับร่าง");
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
            <div className="shrink-0 pb-3 border-b border-border/70 flex flex-col gap-2">
              <form className="flex flex-wrap items-center gap-2 w-full" onSubmit={(event) => { event.preventDefault(); searchDebounce.searchNow(); }}>
                <SearchInput
                  className="min-w-44 flex-1"
                  ariaLabel={tr("gl_search_number_or_desc", "ค้นหาเลขที่หรือคำอธิบาย")}
                  placeholder={tr("gl_number_or_desc", "เลขที่หรือคำอธิบาย")}
                  value={searchDebounce.query}
                  onChange={searchDebounce.setQuery}
                  onClear={searchDebounce.clear}
                  onSearch={searchDebounce.searchNow}
                />
                <div className="flex flex-wrap items-center gap-2 shrink-0">
                  <Button type="submit" variant="outline" className={actionClass}><Search className="size-4 mr-1.5" />{tr("gl_search", "ค้นหา")}</Button>
                  <Button type="button" variant="outline" className={actionClass} onClick={() => { list.reload(); refs.reload(); }} disabled={list.loading}><RefreshCw className="size-4 mr-1.5" />{tr("gl_reload", "โหลดใหม่")}</Button>
                  {mode === "edit" && (
                    <Button type="button" className={actionClass} disabled={busy} onClick={() => void openCreate()}>
                      <Plus className="size-4 mr-1.5" />
                      <span>{tr("gl_add_row", "เพิ่มรายการ")}</span>
                      <kbd className="ml-1.5 hidden sm:inline-block rounded border border-primary-foreground/30 bg-primary-foreground/15 px-1.5 py-0.5 text-[10px] font-mono text-primary-foreground">
                        Alt+N
                      </kbd>
                    </Button>
                  )}
                  <Button type="button" variant="outline" className={actionClass} aria-pressed={density.compact} onClick={density.toggle}>{density.compact ? tr("gl_expand_row", "ขยายบรรทัด") : tr("gl_collapse_row", "ย่อบรรทัด")}</Button>
                </div>
              </form>
              <div className="flex items-center justify-between px-1 text-xs text-muted-foreground">
                <span className="font-medium text-foreground/80">{tr("gl_x_items", "{0} รายการ").replace("{0}", String(list.data.total.toLocaleString("th-TH")))}</span>
                {density.compact && <span className="inline-flex items-center rounded-md bg-muted px-2 py-0.5 text-[11px] font-medium text-muted-foreground">{tr("gl_collapse_row_mode", "โหมดย่อบรรทัด")}</span>}
              </div>
            </div>
            <div className="flex-1 min-h-[300px] overflow-auto rounded-xl border border-border shadow-sm" aria-busy={list.loading}>
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
                    {([ [tr("gl_total_debit", "รวมเดบิต"), totals?.debit], [tr("gl_total_credit", "รวมเครดิต"), totals?.credit], [tr("gl_difference", "ผลต่าง"), totals?.difference] ] as const).map(([label, units]) => (
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
                        {tr("gl_post_accounting_entry", "ผ่านรายการ")}
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
                <header className="flex items-center justify-between gap-2 border-b border-border pb-3 shrink-0">
                  <div className="flex min-w-0 items-center gap-2">
                    <div className="grid size-8 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
                      {journal.id ? <Pencil className="size-4" /> : <Plus className="size-4" />}
                    </div>
                    <h2 className="truncate text-base font-semibold">
                      {journal.id ? tr("gl_edit_draft", "แก้ไขฉบับร่าง {0}").replace("{0}", String(journal.docno)) : tr("gl_save_new_journal", "บันทึกรายวันใหม่")}
                    </h2>
                  </div>
                  <div className="flex shrink-0 items-center gap-2">
                    <UnsavedBadge dirty={dirty} />
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
                  </div>
                </header>
                <div className="overflow-y-auto p-1 grid min-w-0 gap-3 content-start">
                  <fieldset disabled={busy} className="grid min-w-0 gap-3">
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label={tr("gl_document_no", "เลขที่เอกสาร")}><input className={control} required value={journal.docno} maxLength={60} onChange={(e) => patch({ docno: e.target.value })} /></Field>
                      <Field label={tr("gl_document_date", "วันที่เอกสาร")}><input className={control} required type="date" value={journal.date} onChange={(e) => patch({ date: e.target.value })} /></Field>
                      <Field label={tr("gl_fiscal_year", "ปีบัญชี")}><YearSelect years={refs.years} value={journal.fiscalyear} onChange={(fiscalyear) => patch({ fiscalyear })} /></Field>
                      <Field label={tr("gl_journal", "สมุดรายวัน")}>
                        <Combobox
                          disabled={!!book}
                          value={journal.bookcode}
                          onChange={(val) => patch({ bookcode: String(val) })}
                          placeholder={tr("gl_journal", "สมุดรายวัน")}
                        >
                          {Object.entries(bookLabels).map(([code, name]) => <option key={code} value={code}>{tr(...name)}</option>)}
                        </Combobox>
                      </Field>
                      <Field label={tr("gl_entry_description", "คำอธิบายรายการ")}><input className={control} required value={journal.description} onChange={(e) => patch({ description: e.target.value })} maxLength={500} /></Field>
                      <Field label={tr("gl_reference_document", "เอกสารอ้างอิง")}><input className={control} value={journal.reference} onChange={(e) => patch({ reference: e.target.value })} /></Field>
                      <Field label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} value={journal.branchcode} onChange={(e) => patch({ branchcode: e.target.value })} /></Field>
                    </div>
                    <div
                      className="overflow-x-auto rounded-xl border border-border"
                      onKeyDown={tabularEnterNav.onKeyDown}
                      onPaste={(e) => {
                        const text = e.clipboardData.getData("text");
                        if (text && (text.includes("\t") || text.includes("\n"))) {
                          e.preventDefault();
                          applyPastedLines(text);
                        }
                      }}
                    >
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
                                <Combobox
                                  aria-label={tr("gl_cash_flow_line", "กระแสเงินสดบรรทัด {0}").replace("{0}", String(index + 1))}
                                  value={line.cashflow}
                                  onChange={(val) => patchLine(index, { cashflow: String(val) })}
                                  placeholder={tr("gl_not_specified_2", "ไม่ระบุ")}
                                >
                                  <option value="">{tr("gl_not_specified_2", "ไม่ระบุ")}</option>
                                  <option value="operating">{tr("gl_operating", "ดำเนินงาน")}</option>
                                  <option value="investing">{tr("gl_investing", "ลงทุน")}</option>
                                  <option value="financing">{tr("gl_raise_funds", "จัดหาเงิน")}</option>
                                </Combobox>
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
                    <div className="flex flex-wrap items-center gap-2">
                      <Button type="button" variant="outline" className={actionClass} disabled={journal.lines.length >= 500} onClick={() => patch({ lines: [...journal.lines, emptyLine()] })}>
                        <Plus className="size-4 mr-1.5" />{tr("gl_add_line", "เพิ่มบรรทัด")}
                      </Button>
                      <Button type="button" variant="outline" className={actionClass} disabled={journal.lines.length >= 500} onClick={() => void handlePasteClick()} title="Paste rows from Excel or Google Sheets (Ctrl+V)">
                        <ClipboardPaste className="size-4 mr-1.5" />
                        <span>Excel Paste</span>
                      </Button>
                      {smartGuard && !smartGuard.isBalanced && (
                        <Button
                          type="button"
                          variant="outline"
                          className={`${actionClass} border-amber-500/40 text-amber-700 bg-amber-50/80 hover:bg-amber-100 dark:bg-amber-950/30 dark:text-amber-300 dark:border-amber-700/50 shadow-sm`}
                          onClick={() => patch({ lines: autoBalanceJournalLines(journal.lines, year?.scale ?? 2) })}
                          title={tr("gl_auto_balance_hint", "ปรับยอดให้เดบิตและเครดิตสมดุลกันอัตโนมัติ")}
                        >
                          <Scale className="size-4 mr-1.5 text-amber-600 dark:text-amber-400" />
                          <span>{tr("gl_auto_balance", "ปรับยอดให้ดุล (Auto-Balance)")}</span>
                        </Button>
                      )}
                      {smartGuard && !smartGuard.vat.hasVatLine && smartGuard.vat.suggestedVatUnits > 0n && (
                        <Button
                          type="button"
                          variant="outline"
                          className={`${actionClass} border-primary/40 text-primary bg-primary/5 hover:bg-primary/10 shadow-sm`}
                          onClick={() => patch({ lines: appendVatLine(journal.lines, refs.accounts, smartGuard.vat.suggestedVatType!, smartGuard.vat.suggestedVatUnits, year?.scale ?? 2) })}
                          title={tr("gl_add_vat_line_hint", "คำนวณและเพิ่มบรรทัดภาษีมูลค่าเพิ่ม 7% จากฐาน")}
                        >
                          <Sparkles className="size-4 mr-1.5 text-primary" />
                          <span>{tr("gl_add_vat_7", "+ ภาษี 7% ({0})").replace("{0}", smartGuard.vat.suggestedVatFormatted)}</span>
                        </Button>
                      )}
                    </div>
                  </fieldset>

                  {/* Totals & Balance Guard Card */}
                  <div
                    className={`grid gap-2 rounded-xl border p-3 sm:grid-cols-3 transition-colors ${
                      smartGuard?.isBalanced
                        ? "border-emerald-500/30 bg-emerald-50/30 dark:bg-emerald-950/10"
                        : "border-destructive/30 bg-destructive/5"
                    }`}
                    aria-live="polite"
                  >
                    <div>
                      <div className="text-[0.9rem] text-muted-foreground">{tr("gl_total_debit", "รวมเดบิต")}</div>
                      <strong className="text-lg tabular-nums">
                        {totals?.debit === undefined ? tr("gl_verify_amount", "ตรวจจำนวนเงิน") : formatAmount(amountString(totals.debit), year?.scale)}
                      </strong>
                    </div>
                    <div>
                      <div className="text-[0.9rem] text-muted-foreground">{tr("gl_total_credit", "รวมเครดิต")}</div>
                      <strong className="text-lg tabular-nums">
                        {totals?.credit === undefined ? tr("gl_verify_amount", "ตรวจจำนวนเงิน") : formatAmount(amountString(totals.credit), year?.scale)}
                      </strong>
                    </div>
                    <div>
                      <div className="flex items-center justify-between gap-1">
                        <span className="text-[0.9rem] text-muted-foreground">{tr("gl_difference", "ผลต่าง")}</span>
                        {smartGuard?.isBalanced ? (
                          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-700 dark:text-emerald-400 bg-emerald-100/70 dark:bg-emerald-900/40 px-2 py-0.5 rounded-full">
                            <CheckCircle2 className="size-3" />
                            {tr("gl_balanced_100", "สมดุล 100%")}
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-destructive bg-destructive/10 px-2 py-0.5 rounded-full">
                            <AlertTriangle className="size-3" />
                            {smartGuard?.balanceStatus === "debit_surplus"
                              ? tr("gl_short_credit", "ขาดเครดิต")
                              : tr("gl_short_debit", "ขาดเดบิต")}
                          </span>
                        )}
                      </div>
                      <strong className={`text-lg tabular-nums ${smartGuard?.isBalanced ? "text-emerald-700 dark:text-emerald-400" : "text-destructive"}`}>
                        {totals?.difference === undefined ? tr("gl_verify_amount", "ตรวจจำนวนเงิน") : formatAmount(amountString(totals.difference), year?.scale)}
                      </strong>
                      {!smartGuard?.isBalanced && smartGuard && (
                        <div className="text-[11px] text-destructive/90 mt-0.5 font-medium">
                          {smartGuard.balanceStatus === "debit_surplus"
                            ? tr("gl_debit_over_credit", "เดบิตมากกว่าเครดิต {0}").replace("{0}", smartGuard.balanceDifferenceFormatted)
                            : tr("gl_credit_over_debit", "เครดิตมากกว่าเดบิต {0}").replace("{0}", smartGuard.balanceDifferenceFormatted)}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* VAT 7% Smart Checker Card */}
                  {smartGuard?.vat.hasVatLine && (
                    <div
                      className={`flex flex-wrap items-center justify-between gap-2 p-2.5 rounded-xl border text-xs ${
                        smartGuard.vat.isExactVat
                          ? "bg-emerald-50/50 border-emerald-500/20 text-emerald-800 dark:bg-emerald-950/20 dark:text-emerald-300"
                          : smartGuard.vat.isCloseVat
                          ? "bg-amber-50/50 border-amber-500/30 text-amber-800 dark:bg-amber-950/20 dark:text-amber-300"
                          : "bg-rose-50/50 border-rose-500/30 text-rose-800 dark:bg-rose-950/20 dark:text-rose-300"
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        {smartGuard.vat.isExactVat ? (
                          <CheckCircle2 className="size-4 shrink-0 text-emerald-600 dark:text-emerald-400" />
                        ) : (
                          <AlertTriangle className="size-4 shrink-0 text-amber-600 dark:text-amber-400" />
                        )}
                        <span>
                          {smartGuard.vat.isExactVat
                            ? tr("gl_vat_verified_exact", "ตรวจสอบภาษีมูลค่าเพิ่ม 7%: ยอด {0} ตรงตามฐานคำนวณเป๊ะ").replace("{0}", smartGuard.vat.actualVatFormatted)
                            : smartGuard.vat.isCloseVat
                            ? tr("gl_vat_close_satang", "ภาษีมูลค่าเพิ่มในเอกสาร {0} (ต่างจาก 7% คำนวณปกติ {1} อยู่ {2} สตางค์ — สรรพากรยอมรับได้ตามใบกำกับภาษีจริง)")
                                .replace("{0}", smartGuard.vat.actualVatFormatted)
                                .replace("{1}", smartGuard.vat.expectedVatFormatted)
                                .replace("{2}", formatAmount(amountString(smartGuard.vat.varianceUnits < 0n ? -smartGuard.vat.varianceUnits : smartGuard.vat.varianceUnits), 2))
                            : tr("gl_vat_mismatch", "ภาษีมูลค่าเพิ่มในเอกสาร {0} ต่างจาก 7% ของฐานภาษี ({1}) โปรดตรวจสอบ")
                                .replace("{0}", smartGuard.vat.actualVatFormatted)
                                .replace("{1}", smartGuard.vat.expectedVatFormatted)}
                        </span>
                      </div>
                      {!smartGuard.vat.isExactVat && (
                        <Button
                          type="button"
                          size="sm"
                          variant="outline"
                          className="h-6 px-2.5 text-[11px] font-medium bg-background border-border hover:bg-muted"
                          onClick={() => patch({ lines: setExactVatLine(journal.lines, smartGuard.vat.vatLineIndex, smartGuard.vat.expectedVatUnits, year?.scale ?? 2) })}
                        >
                          {tr("gl_adjust_exact_vat", "ปรับเป็น 7% พอดี ({0})").replace("{0}", smartGuard.vat.expectedVatFormatted)}
                        </Button>
                      )}
                    </div>
                  )}
                  {journal.id && (
                    <Field label={tr("gl_reason", "เหตุผล")}><input className={control} value={reason} disabled={busy} onChange={(e) => setReason(e.target.value)} placeholder={tr("gl_edit_reason_hint", "ระบุเหตุผลการแก้ไข (ถ้ามี)")} /></Field>
                  )}
                </div>
                <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-3 shrink-0 mt-auto">
                  <div className="flex flex-wrap gap-2">
                    <Button type="submit" className={actionClass} disabled={busy || !dirty}>
                      <Save className="size-4 mr-1.5" />
                      <span>{busy ? tr("gl_saving", "กำลังบันทึก…") : tr("gl_save_draft", "บันทึกฉบับร่าง")}</span>
                      <kbd className="ml-1.5 hidden sm:inline-block rounded border border-primary-foreground/30 bg-primary-foreground/15 px-1.5 py-0.5 text-[10px] font-mono text-primary-foreground">
                        Ctrl+S
                      </kbd>
                    </Button>
                    <Button type="button" variant="outline" className={actionClass} onClick={() => void cancelEdit()}>
                      <span>{tr("gl_cancel", "ยกเลิก")}</span>
                      <kbd className="ml-1.5 hidden sm:inline-block rounded border border-border bg-muted/60 px-1.5 py-0.5 text-[10px] font-mono text-muted-foreground">
                        Esc
                      </kbd>
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
