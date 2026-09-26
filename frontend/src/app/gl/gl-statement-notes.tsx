"use client";
// หมายเหตุประกอบงบการเงิน (ประกาศกรมพัฒนาธุรกิจการค้า เรื่อง กำหนดรายการย่อที่ต้องมีในงบการเงิน พ.ศ. 2566 แบบ 2 ข้อ 5; TFRS for NPAEs 4.1)
// อยู่ในจอออกแบบงบการเงิน (ไม่มีเมนูใหม่): หนึ่งรายการต่อปีบัญชี เก็บเป็นข้อมูลหลัก statement-notes (code = รหัสปีบัญชี);
// ตรวจเลขที่/ความยาวที่ backend (backend/internal/generalledger/statement_notes.go) — จอนี้แค่แก้ข้อความและพิมพ์
import { useEffect, useMemo, useRef, useState } from "react";
import { ArrowDown, ArrowUp, FileText, Plus, Printer, RotateCw, Save, Sparkles, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { GLCommandError, glRequest } from "@/lib/general-ledger-api";
import { newStatementNote, starterStatementNotes, statementNoteHasText, statementNotesAfterReload, statementNotesFromRecord, statementNotesNeedReload, type GLPage, type GLStatementNotes, type StatementNote } from "@/lib/general-ledger";
import { Field, Notice, UnsavedBadge, YearSelect, actionClass, control, useGLCommand, useGLLanguage, useGLText, useReferences } from "./gl-common";
import { GLNotesPrint, printCompanyName, statementPeriodText, useGLPrint } from "./gl-print";

/** อ่านหมายเหตุของปีบัญชี: ค้นจาก list แล้วเลือกรายการที่ code ตรงปีพอดี จากนั้นอ่านฉบับล่าสุดด้วย id (ได้ version สำหรับแก้ไข) */
async function loadStatementNotes(year: string): Promise<{ record: GLStatementNotes; notes: StatementNote[] } | null> {
  const page = await glRequest<GLPage<GLStatementNotes>>(`statement-notes?${new URLSearchParams({ q: year, page: "1", limit: "100" })}`);
  const found = (page.items ?? []).find((item) => item.code === year);
  if (!found?.id) return null;
  const record = await glRequest<GLStatementNotes>(`statement-notes/${encodeURIComponent(found.id)}`);
  return { record, notes: statementNotesFromRecord(record) };
}

export function GLStatementNotesEditor({ onDirtyChange }: { onDirtyChange: (dirty: boolean) => void }) {
  const tr = useGLText();
  const language = useGLLanguage();
  const refs = useReferences();
  const notesPrint = useGLPrint();
  const { busy, execute } = useGLCommand();
  const { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const [year, setYear] = useState("");
  const [revision, setRevision] = useState(0);
  const [record, setRecord] = useState<GLStatementNotes | null>(null);
  // draft = null: ปีนี้ยังไม่มีหมายเหตุและยังไม่เริ่มเขียน; saved = สำเนาที่บันทึกแล้ว ใช้เทียบว่ามีการแก้ไขค้างอยู่
  const [draft, setDraft] = useState<StatementNote[] | null>(null);
  const [saved, setSaved] = useState("null");
  const [loading, setLoading] = useState(false);
  // โหลดปีนี้ไม่สำเร็จ: ห้ามแสดงว่า "ยังไม่มีหมายเหตุ" และห้ามเริ่มเขียน/บันทึก (จะสร้างซ้ำทับปีที่มีหมายเหตุอยู่แล้ว)
  const [loadFailed, setLoadFailed] = useState(false);
  // สำเนาที่ส่งบันทึกล่าสุด: ถ้าผู้ใช้พิมพ์ต่อระหว่างรอบันทึก ผลอ่านกลับจาก backend ต้องไม่ทับข้อความที่พิมพ์เพิ่ม
  const submittedRef = useRef<string | null>(null);
  const [error, setError] = useState("");
  // บันทึก/ลบชนกับฉบับที่บันทึกไว้ (version เก่า / ปีนี้มีคนสร้างก่อน / ถูกลบไปแล้ว): ส่งซ้ำไม่มีวันผ่าน ต้องมีปุ่มโหลดฉบับล่าสุด
  const [reloadNeeded, setReloadNeeded] = useState(false);
  const [message, setMessage] = useState("");
  const dirty = JSON.stringify(draft) !== saved;

  useEffect(() => { onDirtyChange(dirty); }, [dirty, onDirtyChange]);
  useEffect(() => () => onDirtyChange(false), [onDirtyChange]);

  // ปีตั้งต้น = ปีบัญชีที่เปิดใช้และยังไม่ปิด (กติกาเดียวกับพรีวิวงบ)
  useEffect(() => {
    if (year || !refs.years.length) return;
    const active = refs.years.find((item) => item.isactive && !item.closed) ?? refs.years[0];
    if (active) setYear(active.code);
  }, [refs.years, year]);

  useEffect(() => {
    if (!year) return;
    let active = true;
    setLoading(true);
    setLoadFailed(false);
    setError("");
    setReloadNeeded(false);
    loadStatementNotes(year)
      .then((result) => {
        if (!active) return;
        const notes = result ? result.notes : null;
        const submitted = submittedRef.current;
        submittedRef.current = null;
        setRecord(result?.record ?? null);
        setSaved(JSON.stringify(notes));
        setDraft((current) => statementNotesAfterReload(current, submitted, notes));
      })
      .catch((e: Error) => { if (active) { setLoadFailed(true); setError(e.message); } })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [year, revision]);

  const fiscalYear = useMemo(() => refs.years.find((item) => item.code === year), [refs.years, year]);
  const period = fiscalYear ? statementPeriodText({ from: fiscalYear.startdate, to: fiscalYear.enddate }, false, true, tr, language) : "";
  const notes = draft ?? [];

  async function confirmDiscard() {
    return !dirty || await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") });
  }

  async function changeYear(next: string) {
    if (next === year || !await confirmDiscard()) return;
    submittedRef.current = null;
    setMessage("");
    setRecord(null);
    setDraft(null);
    setSaved("null");
    setYear(next);
  }

  async function reloadLatest() {
    if (busy || !await confirmDiscard()) return;
    submittedRef.current = null;
    setMessage("");
    setError("");
    setReloadNeeded(false);
    setRecord(null);
    setDraft(null);
    setSaved("null");
    setRevision((value) => value + 1);
  }

  function saveFailed(e: unknown) {
    setError((e as Error).message);
    setReloadNeeded(statementNotesNeedReload(e instanceof GLCommandError ? e.code : ""));
  }

  function updateNote(id: string, patch: Partial<StatementNote>) {
    setDraft((current) => (current ?? []).map((note) => (note.id === id ? { ...note, ...patch } : note)));
  }

  // ลบหมายเหตุที่มีข้อความต้องถามก่อน (เนื้อหายาวได้ถึง 20,000 ตัวอักษร ไม่มีปุ่มย้อนกลับ); ข้อที่ว่างทั้งข้อลบได้เลย
  async function removeNote(note: StatementNote, label: string) {
    if (statementNoteHasText(note) && !await confirm({
      title: tr("gl_statement_note_delete_confirm_title", "ลบหมายเหตุ {0}?").replace("{0}", label),
      description: tr("gl_statement_note_delete_confirm_desc", "หัวข้อและเนื้อหาของหมายเหตุข้อนี้จะหายไป และจะลบถาวรเมื่อกดบันทึกหมายเหตุ"),
      confirmLabel: tr("gl_statement_note_delete", "ลบหมายเหตุข้อนี้"),
      tone: "danger",
    })) return;
    setDraft((current) => (current ?? []).filter((item) => item.id !== note.id));
  }

  function moveNote(index: number, offset: -1 | 1) {
    setDraft((current) => {
      const next = [...(current ?? [])];
      const target = index + offset;
      if (target < 0 || target >= next.length) return current;
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  }

  async function save() {
    if (!year || draft === null || busy || loading) return;
    const submitted = draft;
    try {
      setError("");
      setReloadNeeded(false);
      setMessage("");
      const result = await execute({
        resource: "statement-notes",
        action: record?.id ? "update" : "create",
        id: record?.id,
        version: record?.version,
        master: { code: year, name: record?.name ?? "", isactive: true, notes: submitted },
      });
      // id/version จากผลบันทึกทันที: ถ้าอ่านกลับไม่สำเร็จ การบันทึกครั้งถัดไปยังเป็น update ด้วย version ที่ถูกต้อง
      setRecord({ ...(record ?? {}), id: result.id, version: result.version, code: year, name: record?.name ?? "", isactive: true, notes: submitted } as GLStatementNotes);
      setSaved(JSON.stringify(submitted));
      setMessage(tr("gl_statement_notes_saved", "บันทึกหมายเหตุประกอบงบการเงินปี {0} เรียบร้อยแล้ว").replace("{0}", year));
      // อ่านฉบับที่ backend บันทึกจริง (ตัดช่องว่างแล้ว) — ไม่ทับข้อความที่ผู้ใช้พิมพ์ต่อระหว่างรอ
      submittedRef.current = JSON.stringify(submitted);
      setRevision((value) => value + 1);
    } catch (e) {
      saveFailed(e);
    }
  }

  async function removeRecord() {
    if (!record?.id || busy) return;
    if (!await confirm({
      title: tr("gl_statement_notes_delete_confirm_title", "ลบหมายเหตุประกอบงบการเงินปี {0}?").replace("{0}", year),
      description: tr("gl_statement_notes_delete_confirm_desc", "หมายเหตุทุกข้อของปีนี้จะถูกลบ และงบการเงินที่อ้างเลขที่หมายเหตุจะมีคำเตือน"),
      confirmLabel: tr("gl_statement_notes_delete_all", "ลบหมายเหตุทั้งปี"),
      tone: "danger",
    })) return;
    try {
      setError("");
      setReloadNeeded(false);
      await execute({ resource: "statement-notes", action: "delete", id: record.id, version: record.version, reason: tr("gl_statement_notes_delete_all", "ลบหมายเหตุทั้งปี") });
      setRecord(null);
      setDraft(null);
      setSaved("null");
      setMessage(tr("gl_statement_notes_deleted", "ลบหมายเหตุประกอบงบการเงินปี {0} เรียบร้อยแล้ว").replace("{0}", year));
    } catch (e) {
      saveFailed(e);
    }
  }

  const title = tr("gl_statement_notes_title", "หมายเหตุประกอบงบการเงิน");

  return (
    <div className="flex flex-col flex-1 min-h-0 gap-3">
      <div className="shrink-0 flex flex-col gap-2">
        <Notice error text={error || refs.error} />
        {reloadNeeded && (
          <div>
            <Button type="button" variant="outline" className={actionClass} disabled={busy || loading} onClick={() => void reloadLatest()}>
              <RotateCw className="mr-1.5 h-4 w-4" /> {tr("gl_statement_notes_reload_latest", "โหลดหมายเหตุล่าสุด")}
            </Button>
          </div>
        )}
        <Notice text={message} />
      </div>

      <div className="flex flex-wrap items-end justify-between gap-3 rounded-xl border border-border bg-muted/20 p-3 shrink-0">
        <div className="flex flex-wrap items-end gap-3 min-w-0">
          <div className="w-48">
            <Field label={tr("gl_fiscal_year", "ปีบัญชี")}>
              <YearSelect years={refs.years} value={year} onChange={(value) => void changeYear(value)} disabled={busy} />
            </Field>
          </div>
          <div className="grid min-w-0 gap-0.5 text-[0.9rem] leading-relaxed text-muted-foreground [overflow-wrap:anywhere]">
            <span className="font-medium text-foreground">{tr("gl_statement_notes_form_hint", "หัวข้อตามแบบ 2 ของประกาศกรมพัฒนาธุรกิจการค้า พ.ศ. 2566")}</span>
            <span>{tr("gl_statement_notes_noteno_hint", "เลขที่หมายเหตุต้องตรงกับช่องหมายเหตุของบรรทัดในรูปแบบงบ งบการเงินจะเตือนถ้าอ้างเลขที่ที่ยังไม่มี")}</span>
          </div>
          <UnsavedBadge dirty={dirty} className="text-[0.9rem]" />
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button type="button" variant="outline" className={actionClass} disabled={!notes.length} onClick={() => notesPrint.print(<GLNotesPrint notes={notes} company={printCompanyName()} period={period} tr={tr} />, "portrait")}>
            <Printer className="mr-1.5 h-4 w-4" /> {tr("gl_statement_notes_print", "พิมพ์หมายเหตุประกอบงบ")}
          </Button>
          {record?.id && (
            <Button type="button" variant="outline" className={`${actionClass} text-destructive hover:bg-destructive/10`} onClick={() => void removeRecord()} disabled={busy}>
              <Trash2 className="mr-1.5 h-4 w-4" /> {tr("gl_statement_notes_delete_all", "ลบหมายเหตุทั้งปี")}
            </Button>
          )}
          <Button type="button" className={actionClass} onClick={() => void save()} disabled={busy || loading || draft === null || !dirty || (loadFailed && !record?.id)}>
            <Save className="mr-1.5 h-4 w-4" /> {busy ? tr("gl_saving_2", "กำลังบันทึก...") : tr("gl_statement_notes_save", "บันทึกหมายเหตุ")}
          </Button>
        </div>
      </div>

      <div className="flex-1 min-h-[300px] overflow-auto" aria-busy={loading}>
        {!year ? (
          <p className="py-8 text-center text-[0.95rem] text-muted-foreground">{tr("gl_statement_notes_select_year", "เลือกปีบัญชีเพื่อเขียนหมายเหตุประกอบงบการเงิน")}</p>
        ) : loading && draft === null ? (
          <p className="py-8 text-center text-[0.95rem] text-muted-foreground">{tr("gl_loading_data_2", "กำลังโหลดข้อมูล...")}</p>
        ) : loadFailed && draft === null ? (
          <div className="mx-auto max-w-2xl rounded-2xl border border-amber-500/50 bg-amber-500/10 p-8 text-center" role="alert">
            <p className="text-[0.95rem] leading-relaxed text-foreground">{tr("gl_statement_notes_load_failed", "โหลดหมายเหตุประกอบงบการเงินของปี {0} ไม่สำเร็จ กรุณากดลองใหม่ก่อนเขียนหมายเหตุ เพื่อไม่ให้เขียนทับหมายเหตุที่มีอยู่แล้ว").replace("{0}", year)}</p>
            <Button type="button" variant="outline" className={`${actionClass} mt-4`} onClick={() => setRevision((value) => value + 1)}>
              <RotateCw className="mr-1.5 h-4 w-4" /> {tr("common_retry", "ลองใหม่")}
            </Button>
          </div>
        ) : !notes.length ? (
          <div className="mx-auto max-w-2xl rounded-2xl border border-dashed border-border p-10 text-center">
            <FileText aria-hidden className="mx-auto mb-3 h-10 w-10 text-muted-foreground/60" />
            <h3 className="text-base font-semibold text-foreground">{tr("gl_statement_notes_empty", "ยังไม่มีหมายเหตุประกอบงบการเงินของปี {0}").replace("{0}", year)}</h3>
            <p className="mt-1 text-[0.95rem] leading-relaxed text-muted-foreground">{tr("gl_statement_notes_form_hint", "หัวข้อตามแบบ 2 ของประกาศกรมพัฒนาธุรกิจการค้า พ.ศ. 2566")}</p>
            <div className="mt-4 flex flex-wrap justify-center gap-2">
              <Button type="button" className={actionClass} onClick={() => setDraft(starterStatementNotes())}>
                <Sparkles className="mr-1.5 h-4 w-4" /> {tr("gl_statement_notes_start_form2", "เริ่มจากหัวข้อตามแบบ 2")}
              </Button>
              <Button type="button" variant="outline" className={actionClass} onClick={() => setDraft([newStatementNote([])])}>
                <Plus className="mr-1.5 h-4 w-4" /> {tr("gl_statement_note_add", "เพิ่มหมายเหตุ")}
              </Button>
            </div>
          </div>
        ) : (
          <div className="mx-auto flex max-w-5xl flex-col gap-3 pb-2">
            <h2 className="text-base font-semibold">{title} · {tr("gl_fiscal_year", "ปีบัญชี")} {year}</h2>
            <ol className="flex flex-col gap-3">
              {notes.map((note, index) => {
                const label = note.noteno.trim() || String(index + 1);
                const locked = busy || loading;
                return (
                  <li key={note.id} className="rounded-xl border border-border bg-card p-3 shadow-sm">
                    <div className="flex flex-wrap items-end gap-2">
                      <div className="w-28">
                        <Field label={tr("gl_statement_note_no", "เลขที่หมายเหตุ")}>
                          <input className={control} value={note.noteno} maxLength={10} readOnly={locked} aria-label={tr("gl_statement_note_no", "เลขที่หมายเหตุ")} onChange={(event) => updateNote(note.id, { noteno: event.target.value })} />
                        </Field>
                      </div>
                      <div className="min-w-48 flex-1">
                        <Field label={tr("gl_statement_note_title", "หัวข้อ")}>
                          <input className={control} value={note.title} maxLength={200} readOnly={locked} aria-label={tr("gl_statement_note_title", "หัวข้อ")} onChange={(event) => updateNote(note.id, { title: event.target.value })} />
                        </Field>
                      </div>
                      <div className="flex flex-wrap items-center gap-1.5">
                        <Button type="button" size="icon" variant="outline" className="size-10 rounded-xl" disabled={locked || index === 0} onClick={() => moveNote(index, -1)}
                          aria-label={tr("gl_statement_note_move_up", "เลื่อนหมายเหตุ {0} ขึ้น").replace("{0}", label)} title={tr("gl_statement_note_move_up", "เลื่อนหมายเหตุ {0} ขึ้น").replace("{0}", label)}>
                          <ArrowUp className="h-4 w-4" />
                        </Button>
                        <Button type="button" size="icon" variant="outline" className="size-10 rounded-xl" disabled={locked || index === notes.length - 1} onClick={() => moveNote(index, 1)}
                          aria-label={tr("gl_statement_note_move_down", "เลื่อนหมายเหตุ {0} ลง").replace("{0}", label)} title={tr("gl_statement_note_move_down", "เลื่อนหมายเหตุ {0} ลง").replace("{0}", label)}>
                          <ArrowDown className="h-4 w-4" />
                        </Button>
                        <Button type="button" variant="outline" className={`${actionClass} text-destructive hover:bg-destructive/10`} disabled={locked} onClick={() => void removeNote(note, label)}>
                          <Trash2 className="mr-1.5 h-4 w-4" /> {tr("gl_statement_note_delete", "ลบหมายเหตุข้อนี้")}
                        </Button>
                      </div>
                    </div>
                    <div className="mt-2">
                      <Field label={tr("gl_statement_note_body", "เนื้อหา")}>
                        <textarea
                          className={`${control} resize-y`}
                          style={{ fontSize: "0.95rem", lineHeight: 1.6 }}
                          rows={Math.min(30, Math.max(6, note.body.split("\n").length + 1))}
                          maxLength={20000}
                          readOnly={locked}
                          value={note.body}
                          aria-label={tr("gl_statement_note_body", "เนื้อหา")}
                          placeholder={tr("gl_statement_note_body_placeholder", "พิมพ์เนื้อหาของหมายเหตุ กด Enter เพื่อขึ้นบรรทัดใหม่")}
                          onChange={(event) => updateNote(note.id, { body: event.target.value })}
                        />
                      </Field>
                    </div>
                  </li>
                );
              })}
            </ol>
            <div>
              <Button type="button" variant="outline" className={actionClass} onClick={() => setDraft([...notes, newStatementNote(notes)])} disabled={busy || loading || notes.length >= 100}>
                <Plus className="mr-1.5 h-4 w-4" /> {tr("gl_statement_note_add", "เพิ่มหมายเหตุ")}
              </Button>
            </div>
          </div>
        )}
      </div>

      {confirmationDialog}
      {notesPrint.portal}
    </div>
  );
}
