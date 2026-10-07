"use client";

import { useEffect, useState } from "react";
import { fillText, type GLAccountReference, type GLLabel, type GLTextFn } from "@/lib/general-ledger";
import { glAccountReferences } from "@/lib/general-ledger-api";
import { Notice, useGLText } from "./gl-common";

// "รายการที่ใช้บัญชีนี้" ของจอผังบัญชี (Champ 5502, GLFrmChartofAccountPage1.cpp:269-483): Champ ไล่ 6 แหล่งของตัวเอง (รูปแบบการเชื่อม กลุ่มบัญชีสินค้า ลูกหนี้ เจ้าหนี้)
// ส่วนนี้นับชนิดข้อมูลของระบบบัญชีที่อ้างถึงบัญชี ชุดเดียวกับที่กันการลบบัญชี — นับที่ backend ในคำสั่งเดียว (account_references.go)

/** Kinds in the order the panel lists them; a kind the backend adds later still shows (last, generic label). */
const KIND_LABELS: Record<string, GLLabel> = {
  journals: ["gl_account_ref_journals", "ใบสำคัญรายวัน"],
  budgets: ["gl_account_ref_budgets", "งบประมาณ"],
  "account-groups": ["gl_account_ref_account_groups", "กลุ่มผังบัญชี"],
  "statement-templates": ["gl_account_ref_statement_templates", "รูปแบบงบการเงิน"],
  mappings: ["gl_account_ref_mappings", "รูปแบบการเชื่อมบัญชี"],
  "product-account-groups": ["gl_account_ref_product_account_groups", "กลุ่มบัญชีสินค้า"],
  "fiscal-years": ["gl_account_ref_fiscal_years", "ปีบัญชี (บัญชีกำไรขาดทุนสุทธิ / กำไรสะสม)"],
  accounts: ["gl_account_ref_accounts", "บัญชีลูกที่อยู่ใต้บัญชีนี้"],
  allocations: ["gl_account_ref_allocations", "การปันส่วนค่าใช้จ่าย"],
  "journal-templates": ["gl_account_ref_journal_templates", "รูปแบบรายการรายวัน"],
};
const KIND_ORDER = Object.keys(KIND_LABELS);

/** Rows of the panel: known kinds in a fixed order, then any other kind by name; zero counts are dropped. */
export function accountReferenceRows(items: readonly GLAccountReference[], tr: GLTextFn): { kind: string; label: string; count: number }[] {
  const rank = (kind: string) => { const index = KIND_ORDER.indexOf(kind); return index < 0 ? KIND_ORDER.length : index; };
  return items
    .filter((item) => item.count > 0)
    .sort((a, b) => rank(a.kind) - rank(b.kind) || a.kind.localeCompare(b.kind))
    .map((item) => {
      const label = KIND_LABELS[item.kind];
      return { kind: item.kind, count: item.count, label: label ? tr(...label) : fillText(tr("gl_account_ref_other", "ข้อมูลอื่น ({0})"), item.kind) };
    });
}

/** Counts of what uses one saved account; reloads when the account or its saved version changes. */
export function AccountReferencesPanel({ accountId, version }: { accountId: string; version?: number }) {
  const tr = useGLText();
  const [state, setState] = useState<{ key: string; items: GLAccountReference[]; error: string } | null>(null);
  const key = `${accountId}:${version ?? 0}`;
  useEffect(() => {
    const controller = new AbortController();
    glAccountReferences(accountId, controller.signal)
      .then((items) => setState({ key, items, error: "" }))
      .catch((cause: unknown) => { if (!controller.signal.aborted) setState({ key, items: [], error: cause instanceof Error ? cause.message : tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง") }); });
    return () => controller.abort();
    // tr only supplies the fallback text of an error; a language switch must not refetch the counts.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [accountId, key]);
  const current = state?.key === key ? state : null;
  const rows = current ? accountReferenceRows(current.items, tr) : [];
  const title = tr("gl_account_refs_title", "รายการที่ใช้บัญชีนี้");
  return (
    <section aria-label={title} aria-busy={!current} className="sm:col-span-2 grid gap-1.5 rounded-xl border border-border/60 bg-muted/20 p-2 text-[0.95rem] leading-normal" data-field="account-references">
      <h3 className="font-semibold text-foreground">{title}</h3>
      {!current && <p className="text-muted-foreground">{tr("gl_loading_data", "กำลังโหลดข้อมูล…")}</p>}
      {current?.error && <Notice error text={current.error} />}
      {current && !current.error && rows.length === 0 && <p className="text-muted-foreground">{tr("gl_account_refs_none", "ยังไม่มีรายการใดใช้บัญชีนี้")}</p>}
      {rows.length > 0 && (
        <ul className="grid gap-1">
          {rows.map((row) => (
            <li key={row.kind} className="flex items-baseline justify-between gap-3 border-b border-border/40 pb-1 last:border-b-0 last:pb-0">
              <span className="min-w-0 [overflow-wrap:anywhere]">{row.label}</span>
              <span className="shrink-0 font-semibold tabular-nums">{fillText(tr("gl_account_ref_count", "{0} รายการ"), row.count.toLocaleString("th-TH"))}</span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
