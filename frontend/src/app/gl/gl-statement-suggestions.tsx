"use client";
// แนะนำบัญชีให้บรรทัด/คอลัมน์ของแม่แบบงบ (ADR docs/kms/decisions/2026-09-27-gl-statement-account-suggestions.md):
// รายการมาจาก backend (statement_suggestions.go) ซึ่งดูเฉพาะข้อมูลที่ผู้ใช้บันทึกไว้ ไม่ดูรหัส/ชื่อบัญชี — จอนี้ไม่เลือกให้เอง
// ทุกบัญชีเริ่มแบบไม่ติ๊ก ผู้ใช้อ่านเหตุผลแล้วติ๊กเอง; บัญชีที่อยู่ในบรรทัดอื่นของแม่แบบนี้แล้วแสดงแยกโดยไม่มีช่องติ๊ก (ถ้าเพิ่มจะนับยอดซ้ำ)
import { useEffect, useId, useRef, useState, type KeyboardEvent } from "react";
import { Plus, Sparkles, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  accountTypeLabels,
  fillText,
  labelText,
  statementSuggestReasonLabels,
  type GLStatementAccountFix,
  type GLStatementSuggestionReason,
  type GLStatementSuggestionTarget,
  type GLTextFn,
} from "@/lib/general-ledger";
import { actionClass } from "./gl-common";

export type StatementTargetRef = { target: "row" | "column"; rowno?: number; title?: string };

/** "บรรทัด 30 “เงินสด…”" / "คอลัมน์ “กำไร (ขาดทุน) สะสม”" */
export function statementTargetLabel(ref: StatementTargetRef, tr: GLTextFn): string {
  const title = ref.title?.trim() || "—";
  return ref.target === "column"
    ? fillText(tr("gl_statement_target_column", "คอลัมน์ “{0}”"), title)
    : fillText(tr("gl_statement_target_row", "บรรทัด {0} “{1}”"), ref.rowno ?? "", title);
}

/** เหตุผลหนึ่งข้อ: {0} = จำนวน, {1} = รหัสแม่แบบงบอื่น (backend ส่งไม่เกิน 5 รหัส; มีมากกว่านั้นต่อท้ายด้วย …) */
export function statementSuggestReasonText(reason: GLStatementSuggestionReason, tr: GLTextFn): string {
  const label = statementSuggestReasonLabels[reason.source];
  if (!label) return "";
  const templates = reason.templates ?? [];
  const list = templates.join(", ") + (templates.length && reason.count > templates.length ? ", …" : "");
  return fillText(tr(label[0], label[1]), reason.count, list);
}

/** ข้อความหลังแทนบัญชีหัวข้อด้วยบัญชีย่อย/นำรหัสที่ไม่มีในผังออก — ต่อบรรทัด: บรรทัดไหน แทนอะไรด้วยกี่บัญชี ข้ามบัญชีไหน */
export function statementFixMessage(fixes: GLStatementAccountFix[], labelOf: (fix: GLStatementAccountFix) => string, tr: GLTextFn): string {
  const parts: string[] = [];
  for (const fix of fixes) {
    const sentences: string[] = [];
    for (const header of fix.headers ?? []) {
      const headerText = [header.accountcode, header.accountname].filter((part) => part?.trim()).join(" ");
      const descendants = header.descendants ?? [];
      const skipped = header.skipped ?? [];
      if (descendants.length || skipped.length) {
        sentences.push(fillText(tr("gl_statement_header_expanded", "บัญชีหัวข้อ {0} ลงรายการไม่ได้ จึงแทนด้วยบัญชีย่อยที่ลงรายการได้ {1} บัญชี"), headerText, descendants.length));
      } else {
        sentences.push(fillText(tr("gl_statement_header_no_children", "บัญชีหัวข้อ {0} ยังไม่มีบัญชีย่อยที่ลงรายการได้ จึงไม่ได้เพิ่มบัญชีใด"), headerText));
      }
      if (skipped.length) {
        sentences.push(fillText(tr("gl_statement_header_skipped_inuse", "บัญชีย่อย {0} อยู่ในบรรทัดอื่นของแม่แบบนี้แล้ว จึงไม่เพิ่มซ้ำ (ถ้าเพิ่มจะนับยอดสองครั้ง)"), skipped.map((item) => item.accountcode).join(", ")));
      }
    }
    if (fix.removed?.length) sentences.push(fillText(tr("gl_statement_code_removed", "นำรหัสบัญชี {0} ออก เพราะไม่มีในผังบัญชีหรือถูกลบแล้ว"), fix.removed.join(", ")));
    if (sentences.length) parts.push(`${labelOf(fix)}: ${sentences.join(" ")}`);
  }
  return parts.join(" · ");
}

/** หน้าต่างตรวจคำแนะนำ: ไม่ติ๊กให้, ปุ่มเพิ่มกดได้เมื่อเลือกอย่างน้อย 1 บัญชี, Escape/ยกเลิกปิด, Tab วนอยู่ในหน้าต่าง, ปิดแล้วโฟกัสกลับปุ่มที่เปิด */
export function GLStatementSuggestionDialog({ open, target, targetLabel, inUseLabel, onClose, onAccept, tr }: {
  open: boolean;
  target: GLStatementSuggestionTarget | null;
  targetLabel: string;
  inUseLabel: (ref: { target: "row" | "column"; id: string; rowno?: number; title: string }) => string;
  onClose: () => void;
  onAccept: (codes: string[]) => void;
  tr: GLTextFn;
}) {
  const [checked, setChecked] = useState<string[]>([]);
  const dialogRef = useRef<HTMLDivElement>(null);
  const titleId = useId();
  const inUseId = useId();

  // เปิดแล้วโฟกัสช่องติ๊กแรก (ไม่มี = ตัวหน้าต่าง); ปิดแล้วคืนโฟกัสให้ปุ่ม "แนะนำ N บัญชี" ที่เปิด (แบบเดียวกับ gl-statement-set.tsx)
  useEffect(() => {
    if (!open) return;
    const opener = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const first = dialogRef.current?.querySelector<HTMLElement>("input[type='checkbox']:not(:disabled)");
    (first ?? dialogRef.current)?.focus();
    return () => { if (opener?.isConnected) opener.focus(); };
  }, [open]);

  if (!open || !target) return null;
  const accounts = target.accounts ?? [];
  const inUse = target.inuse ?? [];

  function toggle(code: string, on: boolean) {
    setChecked((current) => (on ? (current.includes(code) ? current : [...current, code]) : current.filter((item) => item !== code)));
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === "Escape" && !event.defaultPrevented) {
      event.preventDefault();
      onClose();
      return;
    }
    if (event.key !== "Tab" || !dialogRef.current) return;
    const focusable = Array.from(dialogRef.current.querySelectorAll<HTMLElement>("button:not(:disabled), input:not(:disabled), select:not(:disabled), [tabindex]:not([tabindex='-1'])"));
    if (!focusable.length) { event.preventDefault(); return; }
    const first = focusable[0], last = focusable[focusable.length - 1];
    if (event.shiftKey && (document.activeElement === first || document.activeElement === dialogRef.current)) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  const selected = accounts.filter((account) => checked.includes(account.accountcode)).map((account) => account.accountcode);
  const cancelLabel = tr("gl_cancel", "ยกเลิก");

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-xs">
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        tabIndex={-1}
        onKeyDown={onKeyDown}
        className="flex max-h-[90vh] w-full max-w-2xl flex-col rounded-2xl border border-border bg-card text-[0.95rem] leading-[1.5] text-foreground shadow-[0_18px_48px_color-mix(in_srgb,var(--primary)_18%,transparent)] outline-none"
      >
        <div className="flex items-start justify-between gap-3 border-b border-border p-4">
          <h2 id={titleId} className="flex items-start gap-2 text-lg font-bold leading-[1.45] [overflow-wrap:anywhere]">
            <Sparkles aria-hidden className="mt-1 size-5 shrink-0 text-primary" />
            {fillText(tr("gl_statement_suggest_dialog_title", "บัญชีที่แนะนำสำหรับ “{0}”"), targetLabel)}
          </h2>
          <Button type="button" variant="ghost" className={actionClass} onClick={onClose} aria-label={cancelLabel} title={cancelLabel}>
            <X aria-hidden className="size-5" />
          </Button>
        </div>

        <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-auto p-4">
          <p className="text-muted-foreground [overflow-wrap:anywhere]">{tr("gl_statement_suggest_dialog_desc", "ระบบแนะนำจากข้อมูลที่บันทึกไว้ในระบบ ไม่ได้ดูจากรหัสหรือชื่อบัญชี — ติ๊กเฉพาะบัญชีที่ต้องการ แล้วกด “เพิ่มบัญชีที่เลือก” (ยังไม่บันทึกจนกว่าจะกดบันทึกแม่แบบ)")}</p>
          <p className="font-medium">{fillText(tr("gl_statement_suggest_expected_type", "แนะนำเฉพาะบัญชีหมวด{0}ที่ลงรายการได้"), labelText(accountTypeLabels, target.accounttype ?? "", tr))}</p>

          <ul className="flex flex-col gap-2">
            {accounts.map((account) => (
              <li key={account.accountcode}>
                <label className="flex min-h-12 cursor-pointer items-start gap-3 rounded-xl border border-border bg-background px-3 py-2.5 shadow-[0_2px_8px_rgba(0,0,0,0.06)] transition-colors hover:border-primary/60 has-[:checked]:border-primary/60 has-[:checked]:bg-primary/5 has-[:focus-visible]:ring-2 has-[:focus-visible]:ring-ring">
                  <Checkbox className="mt-0.5" checked={checked.includes(account.accountcode)} onCheckedChange={(on) => toggle(account.accountcode, on)} />
                  <span className="min-w-0 flex-1 [overflow-wrap:anywhere]">
                    <span className="font-semibold tabular-nums">{account.accountcode}</span> {account.accountname}
                    {!account.isactive && (
                      <span className="ml-2 inline-flex items-center rounded-md border border-border bg-muted px-1.5 font-medium text-muted-foreground">
                        {tr("gl_statement_suggest_inactive", "ปิดใช้งาน")}
                      </span>
                    )}
                    <span className="mt-0.5 block text-muted-foreground">
                      {(account.reasons ?? []).map((reason) => statementSuggestReasonText(reason, tr)).filter(Boolean).join(" · ")}
                    </span>
                  </span>
                </label>
              </li>
            ))}
          </ul>

          {inUse.length > 0 && (
            <section aria-labelledby={inUseId} className="rounded-xl border border-border bg-muted/20 p-3">
              <h3 id={inUseId} className="font-semibold leading-[1.45]">{tr("gl_statement_suggest_inuse_title", "อยู่ในบรรทัดอื่นของแม่แบบนี้แล้ว จึงไม่แนะนำซ้ำ")}</h3>
              <ul className="mt-1.5 list-disc space-y-1 pl-6 [overflow-wrap:anywhere]">
                {inUse.map((item, index) => (
                  <li key={`${item.accountcode}-${item.target}-${item.id}-${index}`}>
                    <span className="font-semibold tabular-nums">{item.accountcode}</span> {item.accountname} — {inUseLabel(item)}
                  </li>
                ))}
              </ul>
            </section>
          )}
        </div>

        <div className="flex flex-wrap items-center justify-end gap-2 border-t border-border p-4">
          <Button type="button" variant="outline" className={actionClass} onClick={onClose}>
            {cancelLabel}
          </Button>
          <Button type="button" className={actionClass} disabled={selected.length === 0} onClick={() => onAccept(selected)}>
            <Plus aria-hidden className="mr-1.5 size-4" /> {fillText(tr("gl_statement_suggest_add", "เพิ่มบัญชีที่เลือก ({0})"), selected.length)}
          </Button>
        </div>
      </div>
    </div>
  );
}
