"use client";

import { useMemo, useState } from "react";
import { Eye, FileText, Pencil, Plus, RefreshCw, Save, Search, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { amountString, books, emptyJournal, emptyLine, formatAmount, journalTotals, localDate, validateJournal, type GLJournal, type GLLine } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { AccountSelect, AmountInput, Field, Notice, Pager, SearchInput, SplitWorkbench, YearSelect, actionClass, control, useDebouncedSearch, useDirtyGuard, useGLCommand, useGLList, useReferences, useRowDensity } from "./gl-common";

const statusLabel: Record<string, string> = { draft: "ฉบับร่าง", posted: "ผ่านรายการแล้ว", reversed: "กลับรายการแล้ว" };

export function GLJournals({ route, book = "", kind = "", mode = "edit" }: { route: string; book?: string; kind?: string; mode?: "edit" | "post" | "reverse" }) {
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
  const { busy, execute } = useGLCommand(), { confirm, confirmationDialog } = useConfirmDialog();
  const dirty = isEditing && journal !== null && JSON.stringify(journal) !== original;
  useDirtyGuard(route, dirty);
  const year = refs.years.find((item) => item.code === journal?.fiscalyear);
  const totals = useMemo(() => { try { return journal ? journalTotals(journal.lines) : null; } catch { return null; } }, [journal]);
  const patch = (value: Partial<GLJournal>) => setJournal((current) => current ? { ...current, ...value } : current);
  const patchLine = (index: number, value: Partial<GLLine>) => patch({ lines: journal!.lines.map((line, i) => i === index ? { ...line, ...value } : line) });

  async function openView(item: GLJournal) {
    if (dirty && !await confirm({ title: "ละทิ้งรายการที่ยังไม่บันทึก?", description: "รายการที่กรอกอยู่จะไม่ถูกบันทึก", tone: "warning", confirmLabel: "ละทิ้งการแก้ไข" })) return;
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
    if (dirty && !await confirm({ title: "ละทิ้งรายการที่ยังไม่บันทึก?", description: "รายการที่กรอกอยู่จะไม่ถูกบันทึก", tone: "warning", confirmLabel: "ละทิ้งการแก้ไข" })) return;
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
    if (dirty && !await confirm({ title: "ละทิ้งรายการที่ยังไม่บันทึก?", description: "รายการที่กรอกอยู่จะไม่ถูกบันทึก", tone: "warning", confirmLabel: "ละทิ้งการแก้ไข" })) return;
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
    if (dirty && !await confirm({ title: "ละทิ้งรายการที่ยังไม่บันทึก?", description: "รายการที่กรอกอยู่จะไม่ถูกบันทึก", tone: "warning", confirmLabel: "ละทิ้งการแก้ไข" })) return;
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
    const problem = validateJournal(journal, year, refs.accounts);
    if (problem) { setError(problem); return; }
    if (journal.id && !await confirm({ title: "บันทึกการแก้ไขฉบับร่าง?", description: journal.docno, confirmLabel: "บันทึกฉบับร่าง", tone: "info" })) return;
    try {
      const result = await execute({ resource: "journals", id: journal.id, version: journal.version, action: journal.id ? "update" : "create", journal });
      const saved = { ...journal, id: result.id, version: result.version };
      setJournal(saved);
      setOriginal(JSON.stringify(saved));
      setIsEditing(false);
      list.reload();
      setError("");
      setMessage(result.projectionpending ? "บันทึกฉบับร่างแล้ว กำลังปรับปรุงข้อมูลสำหรับรายงาน" : "บันทึกฉบับร่างแล้ว ตรวจสอบและกดผ่านรายการเมื่อพร้อม");
    } catch (e) { setError((e as Error).message); }
  }

  async function act(action: "post" | "reverse" | "delete") {
    if (!journal?.id || busy || dirty) return;
    if (action === "post") { const problem = validateJournal(journal, year, refs.accounts); if (problem) { setError(problem); return; } }
    if (action !== "post" && !reason.trim()) { setError("กรุณาระบุเหตุผลก่อนทำรายการ"); return; }
    if (action === "reverse" && (!reverseDate || !reverseDocno.trim())) { setError("กรุณาระบุวันที่และเลขที่ใบกลับรายการ"); return; }
    const label = action === "post" ? "ผ่านรายการบัญชี" : action === "reverse" ? "สร้างรายการกลับบัญชี" : "ลบฉบับร่าง";
    if (!await confirm({ title: `${label}?`, description: action === "post" ? "หลังผ่านรายการจะไม่สามารถแก้ไขหรือลบได้ การแก้ไขต้องสร้างรายการกลับบัญชีพร้อมเหตุผล" : action === "reverse" ? `สร้างเอกสาร ${reverseDocno} วันที่ ${reverseDate} กลับเดบิตและเครดิตของ ${journal.docno} โดยเก็บรายการเดิมไว้` : `ลบฉบับร่าง ${journal.docno} พร้อมเก็บประวัติ`, details: `เดบิต ${totals ? formatAmount(amountString(totals.debit), year?.scale) : "—"} · เครดิต ${totals ? formatAmount(amountString(totals.credit), year?.scale) : "—"}`, confirmLabel: label, tone: action === "delete" ? "danger" : "warning" })) return;
    try {
      await execute({ resource: "journals", action, id: journal.id, version: journal.version, reason, ...(action === "reverse" ? { date: reverseDate, docno: reverseDocno } : {}) });
      setJournal(null);
      setOriginal("");
      setIsEditing(false);
      list.reload();
      setError("");
      setMessage(`${label}เรียบร้อยแล้ว`);
    } catch (e) { setError((e as Error).message); }
  }

  async function deleteDraftDirect(item: GLJournal) {
    if (!item.id || busy || item.status !== "draft") return;
    if (!await confirm({
      title: "ลบฉบับร่าง?",
      description: `ลบฉบับร่าง ${item.docno} พร้อมเก็บประวัติ`,
      confirmLabel: "ลบฉบับร่าง",
      tone: "danger",
    })) return;
    try {
      await execute({ resource: "journals", action: "delete", id: item.id, version: item.version, reason: "ลบฉบับร่าง" });
      if (journal?.id === item.id) {
        setJournal(null);
        setOriginal("");
        setIsEditing(false);
      }
      list.reload();
      setError("");
      setMessage(`ลบฉบับร่าง ${item.docno} เรียบร้อยแล้ว`);
    } catch (e) {
      setError((e as Error).message);
    }
  }

  return (
    <div className="flex flex-col flex-1 min-h-0 gap-2">
      <div className="shrink-0 flex flex-col gap-2">
        <Notice error text={error || list.error || refs.error} />
        <Notice text={message} />
        {kind === "opening" && <Notice text="บันทึกยอดยกมาเป็นรายการเดบิตและเครดิตที่สมดุล เลือกปีบัญชี วันที่ และบัญชีคู่รายการตามยอดปิดที่ตรวจสอบแล้ว" />}
      </div>
      <SplitWorkbench
        list={
          <div className="flex flex-col flex-1 min-h-0 gap-2">
            <form className="flex flex-wrap gap-2 shrink-0" onSubmit={(event) => { event.preventDefault(); searchDebounce.searchNow(); }}>
              <SearchInput
                className="min-w-28 flex-1"
                ariaLabel="ค้นหาเลขที่หรือคำอธิบาย"
                placeholder="เลขที่หรือคำอธิบาย"
                value={searchDebounce.query}
                onChange={searchDebounce.setQuery}
                onClear={searchDebounce.clear}
                onSearch={searchDebounce.searchNow}
              />
              <Button type="submit" variant="outline" className={actionClass}><Search className="size-4 mr-1.5" />ค้นหา</Button>
              <Button type="button" variant="outline" className={actionClass} onClick={() => { list.reload(); refs.reload(); }} disabled={list.loading}><RefreshCw className="size-4 mr-1.5" />โหลดใหม่</Button>
              {mode === "edit" && <Button type="button" className={actionClass} disabled={busy} onClick={() => void openCreate()}><Plus className="size-4 mr-1.5" />เพิ่มรายการ</Button>}
              <Button type="button" variant="outline" className={actionClass} aria-pressed={density.compact} onClick={density.toggle}>{density.compact ? "ขยายบรรทัด" : "ย่อบรรทัด"}</Button>
            </form>
            <div className="flex items-center justify-between px-1 text-xs text-muted-foreground shrink-0">
              <span>{list.data.total.toLocaleString("th-TH")} รายการ</span>
              {density.compact && <span className="text-[11px]">โหมดย่อบรรทัด</span>}
            </div>
            <div className="flex-1 min-h-[300px] overflow-auto rounded-xl border border-border" aria-busy={list.loading}>
              <table className={`w-full text-left text-[0.95rem] leading-normal ${density.tableClass}`}>
                <thead className="sticky top-0 bg-muted z-10">
                  <tr>
                    <th className="p-2.5">วันที่ / เลขที่</th>
                    <th className="p-2.5 min-w-44">คำอธิบาย</th>
                    <th className="p-2.5 text-center w-28">สถานะ</th>
                    <th className="p-2.5 text-right pr-3 w-24">จัดการ</th>
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
                            ? "bg-amber-100/70 hover:bg-amber-100/90 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 ring-1 ring-inset ring-amber-500/50 font-medium"
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
                          <span className={`font-mono font-bold ${isRowEditing ? "text-amber-900 dark:text-amber-300" : "text-primary"}`}>{item.docno}</span>
                        </td>
                        <td className="max-w-60 truncate p-2" title={item.description}>{item.description}</td>
                        <td className="whitespace-nowrap p-2 text-center">
                          <span className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ${
                            isDraft
                              ? "bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/20"
                              : item.status === "posted"
                                ? "bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20"
                                : "bg-muted text-muted-foreground border border-border"
                          }`}>
                            {statusLabel[item.status] ?? "ตรวจสอบสถานะ"}
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
                                  aria-label="แก้ไข"
                                  title="แก้ไขฉบับร่าง (Edit)"
                                >
                                  <Pencil className="size-3.5 shrink-0" />
                                </Button>
                                <Button
                                  type="button"
                                  size="icon"
                                  variant="outline"
                                  className="size-7 rounded-md bg-background text-red-600 border-red-200/70 hover:bg-red-50 hover:border-red-300 dark:text-red-400 dark:border-red-900/50 dark:hover:bg-red-950/40 shadow-none transition-colors shrink-0"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    void deleteDraftDirect(item);
                                  }}
                                  aria-label="ลบ"
                                  title="ลบฉบับร่าง (Delete)"
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
                                aria-label="แสดงข้อมูล"
                                title="แสดงข้อมูล (View)"
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
                        {list.loading ? "กำลังโหลดข้อมูล…" : "ยังไม่มีรายการในสมุดนี้"}
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
                        <h2 className="text-base font-semibold">แสดงข้อมูล: {journal.docno}</h2>
                        <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-semibold bg-muted text-muted-foreground border border-border">
                          {statusLabel[journal.status]} · โหมดแสดงข้อมูล
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
                        title="แก้ไขฉบับร่างนี้ (Edit)"
                      >
                        <Pencil className="size-4 mr-1.5" />แก้ไขฉบับร่าง
                      </Button>
                    )}
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="size-8 rounded-lg text-muted-foreground hover:text-foreground"
                      onClick={() => { setJournal(null); setOriginal(""); setIsEditing(false); }}
                      title="ปิดหน้าต่างแสดงข้อมูล"
                    >
                      <X className="size-4" />
                    </Button>
                  </div>
                </header>
                <div className="flex items-center gap-2 rounded-lg border border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground shrink-0">
                  <Eye className="size-3.5 text-primary shrink-0" />
                  <span>
                    โหมดแสดงข้อมูล (Read-only) — {journal.status === "draft" ? "หากต้องการแก้ไข ให้กดปุ่ม \"แก้ไขฉบับร่าง\"" : "เอกสารนี้ผ่านรายการแล้ว ไม่สามารถแก้ไขได้โดยตรง"}
                  </span>
                </div>
                <div className="overflow-y-auto p-1 grid min-w-0 gap-3 content-start">
                  <fieldset disabled className="grid min-w-0 gap-3 opacity-95">
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label="เลขที่เอกสาร"><input className={control} value={journal.docno} readOnly /></Field>
                      <Field label="วันที่เอกสาร"><input className={control} type="date" value={journal.date} readOnly /></Field>
                      <Field label="ปีบัญชี"><input className={control} value={journal.fiscalyear} readOnly /></Field>
                      <Field label="สมุดรายวัน"><input className={control} value={books[journal.bookcode as keyof typeof books] ?? journal.bookcode} readOnly /></Field>
                      <Field label="คำอธิบายรายการ"><input className={control} value={journal.description} readOnly /></Field>
                      <Field label="เอกสารอ้างอิง"><input className={control} value={journal.reference || "-"} readOnly /></Field>
                      <Field label="รหัสสาขา"><input className={control} value={journal.branchcode || "-"} readOnly /></Field>
                      <Field label="สกุลเงิน"><input className={control} value={journal.currency || "THB"} readOnly /></Field>
                    </div>
                    <div className="overflow-x-auto rounded-xl border border-border">
                      <table className="w-full min-w-[760px] text-left text-[0.95rem]">
                        <thead className="bg-muted">
                          <tr>
                            <th className="p-2">บัญชี / คำอธิบาย</th>
                            <th className="p-2 text-right">เดบิต</th>
                            <th className="p-2 text-right">เครดิต</th>
                            <th className="p-2">แผนก / โครงการ</th>
                            <th className="p-2">กระแสเงินสด</th>
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
                                <span className="text-xs text-muted-foreground">{line.cashflow || "ไม่ระบุ"}</span>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </fieldset>
                  <div className="grid gap-2 rounded-xl border border-primary/20 bg-primary/5 p-3 sm:grid-cols-3" aria-live="polite">
                    {([ ["เดบิตรวม", totals?.debit], ["เครดิตรวม", totals?.credit], ["ผลต่าง", totals?.difference] ] as const).map(([label, units]) => (
                      <div key={label}>
                        <div className="text-[0.9rem] text-muted-foreground">{label}</div>
                        <strong className="text-lg tabular-nums">{units === undefined ? "ตรวจจำนวนเงิน" : formatAmount(amountString(units), year?.scale)}</strong>
                      </div>
                    ))}
                  </div>
                  {journal.id && journal.status === "draft" && (
                    <Field label="เหตุผล"><input className={control} value={reason} disabled={busy} onChange={(e) => setReason(e.target.value)} placeholder="จำเป็นสำหรับลบฉบับร่าง" /></Field>
                  )}
                  {journal.status === "posted" && (
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label="เลขที่ใบกลับรายการ"><input className={control} value={reverseDocno} disabled={busy} onChange={(e) => setReverseDocno(e.target.value)} placeholder="ระบุเลขที่ใบกลับรายการ" /></Field>
                      <Field label="วันที่กลับรายการ"><input className={control} type="date" value={reverseDate} disabled={busy} onChange={(e) => setReverseDate(e.target.value)} /></Field>
                    </div>
                  )}
                  {journal.status === "posted" && (
                    <Field label="เหตุผลการกลับรายการ"><input className={control} value={reason} disabled={busy} onChange={(e) => setReason(e.target.value)} placeholder="จำเป็นสำหรับสร้างรายการกลับบัญชี" /></Field>
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
                        <Pencil className="size-4 mr-1.5" />แก้ไขฉบับร่าง
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
                        ผ่านรายการบัญชี
                      </Button>
                    )}
                    {journal.status === "posted" && (
                      <Button
                        type="button"
                        className={actionClass}
                        disabled={busy}
                        onClick={() => void act("reverse")}
                      >
                        สร้างรายการกลับบัญชี
                      </Button>
                    )}
                    <Button
                      type="button"
                      variant="outline"
                      className={actionClass}
                      onClick={() => { setJournal(null); setOriginal(""); setIsEditing(false); }}
                    >
                      ปิด
                    </Button>
                  </div>
                  {journal.status === "draft" && journal.id && (
                    <Button
                      type="button"
                      variant="outline"
                      className={`${actionClass} text-red-600 border-red-200/70 hover:bg-red-50 dark:text-red-400 dark:border-red-900/50 dark:hover:bg-red-950/40`}
                      disabled={busy}
                      onClick={() => void act("delete")}
                    >
                      <Trash2 className="size-4 mr-1.5" />ลบฉบับร่าง
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
                      <h2 className="text-base font-semibold">{journal.id ? `แก้ไขฉบับร่าง ${journal.docno}` : "บันทึกรายวันใหม่"}</h2>
                      {dirty && <span className="text-xs text-amber-600 dark:text-amber-400">● มีการเปลี่ยนแปลงที่ยังไม่บันทึก</span>}
                    </div>
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="size-8 rounded-lg text-muted-foreground hover:text-foreground"
                    onClick={() => void cancelEdit()}
                    title="ปิดหน้าต่างแก้ไข"
                  >
                    <X className="size-4" />
                  </Button>
                </header>
                <div className="overflow-y-auto p-1 grid min-w-0 gap-3 content-start">
                  <fieldset disabled={busy} className="grid min-w-0 gap-3">
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label="เลขที่เอกสาร"><input className={control} required value={journal.docno} maxLength={60} onChange={(e) => patch({ docno: e.target.value })} /></Field>
                      <Field label="วันที่เอกสาร"><input className={control} required type="date" value={journal.date} onChange={(e) => patch({ date: e.target.value })} /></Field>
                      <Field label="ปีบัญชี"><YearSelect years={refs.years} value={journal.fiscalyear} onChange={(fiscalyear) => patch({ fiscalyear, currency: refs.years.find((item) => item.code === fiscalyear)?.currency ?? "" })} /></Field>
                      <Field label="สมุดรายวัน"><select className={control} disabled={!!book} value={journal.bookcode} onChange={(e) => patch({ bookcode: e.target.value })}>{Object.entries(books).map(([code, name]) => <option key={code} value={code}>{name}</option>)}</select></Field>
                      <Field label="คำอธิบายรายการ"><input className={control} required value={journal.description} onChange={(e) => patch({ description: e.target.value })} maxLength={500} /></Field>
                      <Field label="เอกสารอ้างอิง"><input className={control} value={journal.reference} onChange={(e) => patch({ reference: e.target.value })} /></Field>
                      <Field label="รหัสสาขา"><input className={control} value={journal.branchcode} onChange={(e) => patch({ branchcode: e.target.value })} /></Field>
                      <Field label="สกุลเงิน"><input className={control} readOnly value={journal.currency} /></Field>
                    </div>
                    <div className="overflow-x-auto rounded-xl border border-border">
                      <table className="w-full min-w-[760px] text-left text-[0.95rem]">
                        <thead className="bg-muted">
                          <tr>
                            <th className="p-2">บัญชี / คำอธิบาย</th>
                            <th className="p-2">เดบิต</th>
                            <th className="p-2">เครดิต</th>
                            <th className="p-2">แผนก / โครงการ</th>
                            <th className="p-2">กระแสเงินสด</th>
                            <th className="p-2">จัดการ</th>
                          </tr>
                        </thead>
                        <tbody>
                          {journal.lines.map((line, index) => (
                            <tr key={index} className="border-t border-border">
                              <td className="min-w-56 p-2">
                                <div className="grid gap-1">
                                  <AccountSelect label={`บัญชีบรรทัด ${index + 1}`} value={line.accountcode} accounts={refs.accounts} onChange={(accountcode) => patchLine(index, { accountcode })} disabled={busy} />
                                  <input className={control} aria-label={`คำอธิบายบรรทัด ${index + 1}`} placeholder="คำอธิบาย" value={line.description} onChange={(e) => patchLine(index, { description: e.target.value })} />
                                </div>
                              </td>
                              <td className="min-w-32 p-2">
                                <AmountInput ariaLabel={`เดบิตบรรทัด ${index + 1}`} disabled={busy} scale={year?.scale ?? 2} value={line.debit} onChange={(debit) => patchLine(index, { debit })} />
                              </td>
                              <td className="min-w-32 p-2">
                                <AmountInput ariaLabel={`เครดิตบรรทัด ${index + 1}`} disabled={busy} scale={year?.scale ?? 2} value={line.credit} onChange={(credit) => patchLine(index, { credit })} />
                              </td>
                              <td className="min-w-32 p-2">
                                <div className="grid gap-1">
                                  <input className={control} aria-label={`แผนกบรรทัด ${index + 1}`} placeholder="แผนก" value={line.departmentcode} onChange={(e) => patchLine(index, { departmentcode: e.target.value })} />
                                  <input className={control} aria-label={`โครงการบรรทัด ${index + 1}`} placeholder="โครงการ" value={line.projectcode} onChange={(e) => patchLine(index, { projectcode: e.target.value })} />
                                </div>
                              </td>
                              <td className="min-w-36 p-2">
                                <select className={control} aria-label={`กระแสเงินสดบรรทัด ${index + 1}`} value={line.cashflow} onChange={(e) => patchLine(index, { cashflow: e.target.value })}>
                                  <option value="">ไม่ระบุ</option>
                                  <option value="operating">ดำเนินงาน</option>
                                  <option value="investing">ลงทุน</option>
                                  <option value="financing">จัดหาเงิน</option>
                                </select>
                              </td>
                              <td className="p-2">
                                <Button type="button" variant="outline" className={actionClass} disabled={busy || journal.lines.length <= 2} onClick={() => patch({ lines: journal.lines.filter((_, i) => i !== index) })}>
                                  นำออก
                                </Button>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                    <Button type="button" variant="outline" className={actionClass} disabled={journal.lines.length >= 500} onClick={() => patch({ lines: [...journal.lines, emptyLine()] })}>
                      <Plus className="size-4 mr-1.5" />เพิ่มบรรทัด
                    </Button>
                  </fieldset>
                  <div className="grid gap-2 rounded-xl border border-primary/20 bg-primary/5 p-3 sm:grid-cols-3" aria-live="polite">
                    {([ ["เดบิตรวม", totals?.debit], ["เครดิตรวม", totals?.credit], ["ผลต่าง", totals?.difference] ] as const).map(([label, units]) => (
                      <div key={label}>
                        <div className="text-[0.9rem] text-muted-foreground">{label}</div>
                        <strong className="text-lg tabular-nums">{units === undefined ? "ตรวจจำนวนเงิน" : formatAmount(amountString(units), year?.scale)}</strong>
                      </div>
                    ))}
                  </div>
                  {journal.id && (
                    <Field label="เหตุผล"><input className={control} value={reason} disabled={busy} onChange={(e) => setReason(e.target.value)} placeholder="ระบุเหตุผลการแก้ไข (ถ้ามี)" /></Field>
                  )}
                </div>
                <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-3 shrink-0 mt-auto">
                  <div className="flex flex-wrap gap-2">
                    <Button type="submit" className={actionClass} disabled={busy || !dirty}>
                      <Save className="size-4 mr-1.5" />{busy ? "กำลังบันทึก…" : "บันทึกฉบับร่าง"}
                    </Button>
                    <Button type="button" variant="outline" className={actionClass} onClick={() => void cancelEdit()}>
                      ยกเลิก
                    </Button>
                  </div>
                  {journal.id && (
                    <Button
                      type="button"
                      variant="outline"
                      className={`${actionClass} text-red-600 border-red-200/70 hover:bg-red-50 dark:text-red-400 dark:border-red-900/50 dark:hover:bg-red-950/40`}
                      disabled={busy}
                      onClick={() => void act("delete")}
                    >
                      <Trash2 className="size-4 mr-1.5" />ลบฉบับร่าง
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
              <h2 className="text-base font-semibold">เลือกรายการเพื่อแสดงข้อมูลบัญชี</h2>
              <p className="text-sm text-muted-foreground">คลิกที่แถวในตารางเพื่อแสดงข้อมูล หรือกดปุ่ม &ldquo;+ เพิ่มรายการ&rdquo; เพื่อบันทึกรายวันใหม่</p>
              {mode === "edit" && (
                <div className="pt-2">
                  <Button type="button" className={actionClass} onClick={() => void openCreate()} disabled={busy}>
                    <Plus className="size-4 mr-1.5" />บันทึกรายวันใหม่
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
