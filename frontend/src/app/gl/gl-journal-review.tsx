"use client";

import { useEffect, useState } from "react";
import { ClipboardCheck, Save } from "lucide-react";
import { Button } from "@/components/ui/button";
import { commandFailure, glRequest } from "@/lib/general-ledger-api";
import type { GLJournalReview, GLReviewStatus } from "@/lib/general-ledger";
import { control, Notice, useGLCommand, useGLText } from "./gl-common";

export function GLJournalReviewPanel({ journalId, version, onDirtyChange, onBusyChange, onReload }: {
  journalId: string; version: number; onDirtyChange: (dirty: boolean) => void;
  onBusyChange: (busy: boolean) => void; onReload: () => void;
}) {
  const tr = useGLText();
  const [data, setData] = useState<GLJournalReview | null>(null);
  const [status, setStatus] = useState<GLReviewStatus>(1);
  const [note, setNote] = useState("");
  const [editing, setEditing] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);
  const { busy, execute } = useGLCommand();
  const dirty = editing && (note !== "" || status !== data?.status);
  const labels = {
    1: tr("gl_review_pending", "รอตรวจ"),
    2: tr("gl_review_difference", "พบข้อแตกต่าง"),
    3: tr("gl_review_checked", "ตรวจแล้ว"),
  };
  useEffect(() => {
    const controller = new AbortController();
    glRequest<GLJournalReview>(`journal-reviews/${encodeURIComponent(journalId)}`, { signal: controller.signal })
      .then((value) => { setData(value); setStatus(value.status); })
      .catch((cause) => { if (!controller.signal.aborted) setError(commandFailure(cause).message); });
    return () => controller.abort();
  }, [journalId]);
  useEffect(() => { onDirtyChange(dirty); return () => onDirtyChange(false); }, [dirty, onDirtyChange]);
  useEffect(() => { onBusyChange(busy); return () => onBusyChange(false); }, [busy, onBusyChange]);

  async function saveReview() {
    if (!data || busy || data.version !== version || (status === 2 && !note.trim())) return;
    setError(""); setSaved(false);
    try {
      await execute({ resource: "journals", action: "review", id: journalId, version,
        review: { status, note: note.trim(), expectedEventNo: data.eventno } });
      // A successful write must not be mistaken for a failed write if refreshing fails.
      setEditing(false); setNote(""); setSaved(true);
      const latest = await glRequest<GLJournalReview>(`journal-reviews/${encodeURIComponent(journalId)}`);
      setData(latest); setStatus(latest.status);
    } catch (cause) {
      setError(commandFailure(cause, tr("gl_review_save_failed", "บันทึกหรือโหลดผลตรวจไม่สำเร็จ กรุณาโหลดเอกสารล่าสุด")).message);
    }
  }

  return <section className="rounded-xl border border-border bg-card p-4 shadow-sm text-[0.95rem] leading-relaxed" aria-label={tr("gl_review_title", "ผลตรวจและข้อแตกต่าง")}>
    <div className="flex flex-wrap items-center justify-between gap-2">
      <h3 className="flex items-center gap-2 font-semibold"><ClipboardCheck className="size-5 text-primary" />{tr("gl_review_title", "ผลตรวจและข้อแตกต่าง")}</h3>
      {data && <span className="rounded-lg border border-border bg-muted px-3 py-1">{data.version === version ? labels[data.status] : tr("gl_review_reload_required", "เอกสารเปลี่ยนแล้ว กรุณาโหลดใหม่")}</span>}
    </div>
    <p className="my-2 text-muted-foreground">{tr("gl_review_separate_posting", "ผลตรวจแยกจากการผ่านรายการ เมื่อแก้ไขเอกสารต้องตรวจใหม่")}</p>
    <Notice error text={error} />
    <Notice text={saved ? tr("gl_review_saved", "บันทึกผลตรวจแล้ว") : ""} />
    {!data && !error && <p role="status">{tr("gl_review_loading", "กำลังโหลดผลตรวจ")}</p>}
    {error || (data && data.version !== version) ? <Button type="button" variant="outline" className="min-h-11" disabled={busy} onClick={onReload}>{tr("gl_review_reload", "โหลดเอกสารล่าสุด")}</Button> : null}
    {data && data.version === version && <>
      {!editing ? <Button type="button" variant="outline" className="min-h-11" onClick={() => { setEditing(true); setSaved(false); }}>{tr("gl_review_record", "บันทึกผลตรวจ")}</Button> : <div className="grid gap-3">
        <label className="grid gap-1">{tr("gl_review_status", "สถานะผลตรวจ")}
          <select className={control} value={status} disabled={busy} onChange={(event) => setStatus(Number(event.target.value) as GLReviewStatus)}>
            {([1, 2, 3] as const).map((value) => <option key={value} value={value}>{labels[value]}</option>)}
          </select>
        </label>
        <label className="grid gap-1">{tr("gl_review_note", "หมายเหตุการตรวจ / ข้อแตกต่าง")}
          <textarea className={`${control} min-h-24`} value={note} maxLength={2000} disabled={busy} onChange={(event) => setNote(event.target.value)} />
        </label>
        {status === 2 && !note.trim() && <p className="text-destructive">{tr("gl_review_note_required", "กรุณาระบุข้อแตกต่างที่พบ")}</p>}
        <div className="flex flex-wrap gap-2">
          <Button type="button" className="min-h-11" disabled={busy || (status === 2 && !note.trim())} onClick={() => void saveReview()}><Save className="mr-2 size-4" />{tr("gl_review_record", "บันทึกผลตรวจ")}</Button>
          <Button type="button" variant="outline" className="min-h-11" disabled={busy} onClick={() => { setEditing(false); setNote(""); setStatus(data.status); }}>{tr("gl_cancel", "ยกเลิก")}</Button>
        </div>
      </div>}
      <details className="mt-3">
        <summary className="cursor-pointer py-2 font-medium">{tr("gl_review_history", "ประวัติผลตรวจ")}</summary>
        {(data.events ?? []).length === 0 && <p className="text-muted-foreground">{tr("gl_review_no_history", "ยังไม่มีประวัติผลตรวจ")}</p>}
        <ol className="grid gap-2">
          {(data.events ?? []).map((event) => <li key={event.eventno} className="rounded-lg border border-border p-3">
            <div className="font-medium">{labels[event.status]} · {tr("gl_review_version", "รุ่นเอกสาร {0}").replace("{0}", String(event.version))}</div>
            <div className="text-muted-foreground break-words">{event.reviewedby} · <time dateTime={event.reviewedat}>{Number.isNaN(Date.parse(event.reviewedat)) ? "—" : new Date(event.reviewedat).toLocaleString("th-TH", { timeZone: "Asia/Bangkok" })}</time></div>
            {event.note && <p className="whitespace-pre-wrap break-words">{event.note}</p>}
          </li>)}
        </ol>
      </details>
    </>}
  </section>;
}
