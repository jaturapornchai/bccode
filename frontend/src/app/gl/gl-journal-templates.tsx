import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { BookmarkPlus, FolderOpen, Trash2, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { ConfirmDialogOptions } from "@/components/ui/confirm-dialog";
import type { GLJournal, GLJournalTemplate, GLJournalTemplateBody, GLPage } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { JOURNAL_TEMPLATE_NAME_MAX, journalTemplateName, journalTemplateNameProblem, templateBodyFromJournal, SME_PRESET_TEMPLATES, presetToJournalTemplate } from "@/lib/gl-journal-templates";
import { Notice, actionClass, control, useGLCommand, useGLText } from "./gl-common";

// Champ daily-entry templates (CGLFrmJournal OnRecur / OnSelRecure, glfrmjournal.cpp:1657 and :2704-2757): the first button names the
// entry on the screen and stores it as a template (Champ asks to confirm, then shows the Recur Name dialog); the second lists the
// templates and fills the screen from the chosen one. A template is a pattern — it never posts and never touches gl_lines.

type TemplateNotice = { error: boolean; text: string };

/** Native modal dialog in a portal: focus stays inside, Esc and Tab are handled by the browser, and the journal form around the
 *  trigger neither receives the dialog's submit/keyboard events nor disables its controls. */
export function TemplateModal({ label, onClose, children }: { label: string; onClose: () => void; children: ReactNode }) {
  const dialog = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const element = dialog.current;
    const trigger = document.activeElement as HTMLElement | null;
    element?.showModal();
    return () => { element?.close(); trigger?.focus(); };
  }, []);
  return createPortal(
    <dialog
      ref={dialog}
      aria-label={label}
      className="fixed inset-0 m-0 h-full max-h-none w-full max-w-none items-center justify-center border-0 bg-black/50 p-4 open:flex"
      // the journal screen listens for Esc/F-keys on window; a modal must keep them to itself
      onKeyDown={(event) => event.stopPropagation()}
      onCancel={(event) => { event.preventDefault(); onClose(); }}
    >
      <section className="grid max-h-[85vh] w-full max-w-2xl gap-3 overflow-auto rounded-xl border border-border bg-card p-4 text-[0.95rem] text-foreground shadow-xl">{children}</section>
    </dialog>,
    document.body,
  );
}

function NameDialog({ onSubmit, onClose }: { onSubmit: (name: string) => Promise<string | null>; onClose: () => void }) {
  const tr = useGLText();
  const [name, setName] = useState(""), [error, setError] = useState(""), [busy, setBusy] = useState(false);
  const submit = async () => {
    if (busy) return;
    const problem = journalTemplateNameProblem(name);
    if (problem) {
      setError(problem.problem === "required"
        ? tr("gl_journal_template_name_required", "กรุณาระบุชื่อรูปแบบรายการ")
        : tr("gl_journal_template_name_too_long", "ชื่อรูปแบบรายการยาวได้ไม่เกิน {0} ตัวอักษร (ตอนนี้ {1} ตัว) กรุณาย่อให้สั้นลง").replace("{0}", String(JOURNAL_TEMPLATE_NAME_MAX)).replace("{1}", String(problem.length)));
      return;
    }
    setBusy(true);
    const failure = await onSubmit(journalTemplateName(name));
    setBusy(false);
    if (failure) setError(failure);
  };
  const title = tr("gl_journal_template_name_title", "ชื่อรูปแบบรายการ");
  return (
    <TemplateModal label={title} onClose={onClose}>
      <h3 className="font-semibold">{title}</h3>
      <label className="grid gap-1">
        {tr("gl_journal_template_name_label", "ตั้งชื่อรูปแบบ (ไม่เกิน {0} ตัวอักษร)").replace("{0}", String(JOURNAL_TEMPLATE_NAME_MAX))}
        <input
          className={control}
          value={name}
          aria-invalid={error ? true : undefined}
          onChange={(event) => { setName(event.target.value); setError(""); }}
          onKeyDown={(event) => { if (event.key === "Enter") { event.preventDefault(); void submit(); } }}
        />
      </label>
      <Notice error text={error} />
      <div className="flex flex-wrap justify-end gap-2">
        <Button type="button" variant="outline" className={actionClass} disabled={busy} onClick={onClose}>{tr("gl_cancel", "ยกเลิก")}</Button>
        <Button type="button" className={actionClass} disabled={busy} onClick={() => void submit()}>{busy ? tr("gl_saving", "กำลังบันทึก…") : tr("gl_journal_template_save_name", "บันทึกรูปแบบ")}</Button>
      </div>
    </TemplateModal>
  );
}

function templateMeta(template: GLJournalTemplate, tr: (key: string, fallback: string) => string): string {
  const body = template.template;
  return [body.bookcode, body.description, tr("gl_journal_template_line_count", "{0} บรรทัด").replace("{0}", String(body.lines.length))].filter(Boolean).join(" · ");
}

function PickerDialog({ onPick, onDelete, onClose }: { onPick: (template: GLJournalTemplate) => void; onDelete: (template: GLJournalTemplate) => void; onClose: () => void }) {
  const tr = useGLText();
  const [activeTab, setActiveTab] = useState<"preset" | "custom">("preset");
  const [items, setItems] = useState<GLJournalTemplate[]>([]), [query, setQuery] = useState(""), [loading, setLoading] = useState(true), [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    glRequest<GLPage<GLJournalTemplate>>(`journal-templates?${new URLSearchParams({ q: "", page: "1", limit: "1000" })}`, { signal: controller.signal })
      .then((page) => { if (!controller.signal.aborted) setItems(page.items ?? []); })
      .catch((cause) => { if (!controller.signal.aborted) setError(cause instanceof Error ? cause.message : tr("gl_journal_template_load_failed", "โหลดรูปแบบรายการไม่สำเร็จ")); })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
    // tr only words the fallback error; a dictionary that arrives later must not restart the request
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const needle = query.trim().toLowerCase();
  const customRows = useMemo(() => {
    return items
      .filter((item) => !needle || `${item.name} ${item.template.bookcode ?? ""} ${item.template.description ?? ""}`.toLowerCase().includes(needle))
      .sort((a, b) => a.name.localeCompare(b.name, "th"));
  }, [items, needle]);

  const presetRows = useMemo(() => {
    return SME_PRESET_TEMPLATES.filter((item) => !needle || `${item.name} ${item.bookcode} ${item.description}`.toLowerCase().includes(needle));
  }, [needle]);

  const title = tr("gl_journal_template_picker_title", "เลือกรูปแบบรายการรายวัน");
  return (
    <TemplateModal label={title} onClose={onClose}>
      <div className="flex items-center justify-between gap-2 border-b border-border pb-2">
        <h3 className="font-semibold text-lg">{title}</h3>
        <Button type="button" variant="outline" size="sm" className="h-8" onClick={onClose}>{tr("gl_close", "ปิด")}</Button>
      </div>

      {/* Tabs: แม่แบบมาตรฐาน SMEs vs แม่แบบของฉัน */}
      <div className="flex items-center gap-1.5 p-1 rounded-xl bg-muted/60 border border-border">
        <button
          type="button"
          className={`flex-1 py-1.5 px-3 rounded-lg text-sm font-medium transition-colors ${activeTab === "preset" ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"}`}
          onClick={() => setActiveTab("preset")}
        >
          <span className="flex items-center justify-center gap-1.5">
            <Sparkles className="size-4 text-primary" />
            <span>{tr("gl_template_preset_sme", "แม่แบบมาตรฐาน SMEs")}</span>
            <span className="ml-1 text-xs px-1.5 py-0.2 rounded-full bg-primary/10 text-primary">{presetRows.length}</span>
          </span>
        </button>
        <button
          type="button"
          className={`flex-1 py-1.5 px-3 rounded-lg text-sm font-medium transition-colors ${activeTab === "custom" ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"}`}
          onClick={() => setActiveTab("custom")}
        >
          <span className="flex items-center justify-center gap-1.5">
            <FolderOpen className="size-4" />
            <span>{tr("gl_template_custom_saved", "แม่แบบที่บันทึกไว้")}</span>
            <span className="ml-1 text-xs px-1.5 py-0.2 rounded-full bg-muted text-muted-foreground">{customRows.length}</span>
          </span>
        </button>
      </div>

      <div className="flex gap-2">
        <input className={control} aria-label={tr("gl_journal_template_search", "ค้นหาชื่อรูปแบบ")} placeholder={tr("gl_journal_template_search", "ค้นหาชื่อรูปแบบ")} value={query} onChange={(event) => setQuery(event.target.value)} />
      </div>

      <Notice error text={error} />

      {activeTab === "preset" ? (
        <ul className="grid gap-2 max-h-[50vh] overflow-y-auto pr-1">
          {presetRows.map((preset) => (
            <li key={preset.id}>
              <Button
                type="button"
                variant="outline"
                className="w-full min-h-14 h-auto justify-start p-3 whitespace-normal text-left hover:border-primary/50 hover:bg-accent/50"
                onClick={() => onPick(presetToJournalTemplate(preset))}
              >
                <div className="grid w-full gap-1">
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-semibold text-foreground text-[0.95rem]">{preset.name}</span>
                    <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-primary/10 text-primary border border-primary/20 shrink-0">
                      {preset.badge} · {tr("gl_preset_book_badge", "สมุด {0}").replace("{0}", preset.bookcode)}
                    </span>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {preset.description} ({tr("gl_preset_lines_count", "{0} บรรทัดบัญชี:").replace("{0}", String(preset.lines.length))} {preset.lines.map((l) => `${l.accountcode} ${l.description}`).join(" · ")})
                  </div>
                </div>
              </Button>
            </li>
          ))}
          {presetRows.length === 0 && <p className="text-center text-muted-foreground py-4">{tr("gl_details_ui_11", "ไม่พบรายการตามเงื่อนไข")}</p>}
        </ul>
      ) : (
        <>
          {loading && <p role="status">{tr("gl_journal_template_loading", "กำลังโหลดรูปแบบรายการ…")}</p>}
          <ul className="grid gap-2 max-h-[50vh] overflow-y-auto pr-1">
            {customRows.map((template) => (
              <li key={template.id ?? template.code} className="flex flex-wrap items-stretch gap-2">
                <Button type="button" variant="outline" className="min-h-11 h-auto min-w-0 flex-1 justify-start whitespace-normal text-left" onClick={() => onPick(template)}>
                  <span className="grid min-w-0 gap-0.5 leading-snug [overflow-wrap:anywhere]">
                    <span className="font-medium">{template.name}</span>
                    <span className="text-[0.9rem] font-normal text-muted-foreground">{templateMeta(template, tr)}</span>
                  </span>
                </Button>
                <Button type="button" variant="outline" className={`${actionClass} h-auto min-h-11 border-destructive/30 text-destructive hover:border-destructive/50 hover:bg-destructive/10`} aria-label={`${tr("gl_delete", "ลบ")} ${template.name}`} onClick={() => onDelete(template)}>
                  <Trash2 className="mr-1.5 size-4" />{tr("gl_delete", "ลบ")}
                </Button>
              </li>
            ))}
          </ul>
          {!loading && !error && customRows.length === 0 && <p className="text-center text-muted-foreground py-4">{items.length === 0 ? tr("gl_journal_template_empty", "ยังไม่มีรูปแบบรายการ — เปิดใบสำคัญที่ต้องการแล้วกด “ตั้งเป็นรูปแบบ”") : tr("gl_details_ui_11", "ไม่พบรายการตามเงื่อนไข")}</p>}
        </>
      )}
    </TemplateModal>
  );
}

export function JournalTemplateTools({ journal, disabled, confirm, onApply, onNotice }: {
  journal: GLJournal | null;
  disabled: boolean;
  confirm: (options: ConfirmDialogOptions) => Promise<boolean>;
  onApply: (template: GLJournalTemplate) => void;
  onNotice: (notice: TemplateNotice) => void;
}) {
  const tr = useGLText();
  const { busy, execute } = useGLCommand();
  const [naming, setNaming] = useState<GLJournalTemplateBody | null>(null), [picking, setPicking] = useState(false);

  async function startSave() {
    if (!journal || disabled || busy) return;
    const body = templateBodyFromJournal(journal);
    if (body.lines.length === 0) {
      onNotice({ error: true, text: tr("gl_journal_template_no_lines", "ใส่รหัสบัญชีอย่างน้อย 1 บรรทัดก่อนตั้งเป็นรูปแบบรายการ") });
      return;
    }
    // Champ: "Do you want to set this transaction to Recur" → Recur Name dialog (mybasedbformview.cpp BaseRecur)
    if (!await confirm({
      title: tr("gl_journal_template_confirm_title", "ตั้งรายการนี้เป็นรูปแบบรายการรายวัน?"),
      description: tr("gl_journal_template_confirm_desc", "ระบบจะจำสมุด เอกสารอ้างอิง คำอธิบาย ประเภทภาษี และรหัสบัญชีของทุกบรรทัดไว้ให้เลือกใช้ซ้ำ — ไม่เก็บเลขที่เอกสาร วันที่ ยอดเดบิต/เครดิต และเลขที่ใบกำกับ รูปแบบไม่ใช่ใบสำคัญและไม่ลงบัญชี"),
      confirmLabel: tr("gl_journal_template_confirm_ok", "ตั้งเป็นรูปแบบ"),
      tone: "info",
    })) return;
    setNaming(body);
  }

  async function saveTemplate(name: string): Promise<string | null> {
    if (!naming) return null;
    try {
      await execute({ resource: "journal-templates", action: "create", master: { name, isactive: true, template: naming } });
    } catch (cause) {
      return cause instanceof Error ? cause.message : tr("gl_journal_template_save_failed", "บันทึกรูปแบบรายการไม่สำเร็จ");
    }
    setNaming(null);
    onNotice({ error: false, text: tr("gl_journal_template_saved", "บันทึกรูปแบบรายการ “{0}” แล้ว").replace("{0}", name) });
    return null;
  }

  async function removeTemplate(template: GLJournalTemplate) {
    if (!template.id || busy) return;
    // the confirm dialog is a plain overlay and would sit behind the native modal picker, so the picker closes first
    setPicking(false);
    const ok = await confirm({
      title: tr("gl_journal_template_delete_title", "ลบรูปแบบรายการ “{0}”?").replace("{0}", template.name),
      description: tr("gl_journal_template_delete_desc", "ลบเฉพาะรูปแบบที่จำไว้ ใบสำคัญที่เคยบันทึกไม่เปลี่ยนแปลง"),
      confirmLabel: tr("gl_journal_template_delete", "ลบรูปแบบ"),
      tone: "danger",
    });
    if (ok) {
      try {
        await execute({ resource: "journal-templates", action: "delete", id: template.id, version: template.version });
        onNotice({ error: false, text: tr("gl_journal_template_deleted", "ลบรูปแบบรายการ “{0}” แล้ว").replace("{0}", template.name) });
      } catch (cause) {
        onNotice({ error: true, text: cause instanceof Error ? cause.message : tr("gl_journal_template_delete_failed", "ลบรูปแบบรายการไม่สำเร็จ") });
      }
    }
    setPicking(true);
  }

  const group = tr("gl_journal_template_group", "รูปแบบรายการรายวัน");
  return (
    <div className="flex flex-wrap items-center gap-2" role="group" aria-label={group}>
      <span className="font-medium">{group}</span>
      <Button type="button" variant="outline" className={actionClass} disabled={disabled || busy || !journal} title={tr("gl_journal_template_save_hint", "จำสมุด คำอธิบาย และรหัสบัญชีของรายการนี้ไว้ใช้ซ้ำ (ไม่เก็บเลขที่ วันที่ และยอดเงิน)")} onClick={() => void startSave()}>
        <BookmarkPlus className="mr-1.5 size-4" />{tr("gl_journal_template_save", "ตั้งเป็นรูปแบบ")}
      </Button>
      <Button type="button" variant="outline" className={actionClass} disabled={disabled || busy} title={tr("gl_journal_template_pick_hint", "เลือกรูปแบบที่จำไว้ แล้วเติมสมุด คำอธิบาย และรหัสบัญชีลงหน้าจอ (ยอดเงินเว้นว่าง)")} onClick={() => setPicking(true)}>
        <FolderOpen className="mr-1.5 size-4" />{tr("gl_journal_template_pick", "เลือกรูปแบบ")}
      </Button>
      {naming && <NameDialog onSubmit={saveTemplate} onClose={() => setNaming(null)} />}
      {picking && <PickerDialog onPick={(template) => { setPicking(false); onApply(template); }} onDelete={(template) => void removeTemplate(template)} onClose={() => setPicking(false)} />}
    </div>
  );
}
