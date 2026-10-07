"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { replaceAccountGroupSingles, type GLAccount, type GLAccountEntry } from "@/lib/general-ledger";
import { AccountSearchDialog, AccountSelect, Field, actionClass, useGLText } from "./gl-common";

/** Entries editor of the chart-of-accounts group screen (Champ 5497 column "account code"): single codes from the picker, plus from-to range rows. */
export function GLAccountGroupEntries({ entries, accounts, groupLabel, onChange }: { entries: GLAccountEntry[]; accounts: GLAccount[]; groupLabel: string; onChange: (entries: GLAccountEntry[]) => void }) {
  const tr = useGLText();
  const [pickerOpen, setPickerOpen] = useState(false);
  const update = (index: number, patch: Partial<GLAccountEntry>) => onChange(entries.map((entry, i) => (i === index ? { ...entry, ...patch } : entry)));
  const singles = entries.filter((entry) => entry.to === undefined).map((entry) => entry.from);
  return (
    <section className="grid gap-2" aria-label={tr("gl_account_code", "รหัสบัญชี")}>
      <div className="flex flex-wrap items-center gap-2">
        <h3 className="font-semibold">{tr("gl_account_code", "รหัสบัญชี")}</h3>
        <Button type="button" variant="outline" className={actionClass} onClick={() => setPickerOpen(true)}>{tr("gl_select_chart_of_accounts", "+ เลือกผังบัญชี")}</Button>
        <Button type="button" variant="outline" className={actionClass} onClick={() => onChange([...entries, { from: "", to: "" }])}>{tr("gl_group_add_range", "+ เพิ่มช่วงรหัส")}</Button>
      </div>
      <p className="text-[0.9rem] leading-relaxed text-muted-foreground">{tr("gl_group_entries_hint", "เลือกบัญชีทีละรหัส หรือเพิ่มช่วงรหัส “จาก–ถึง” (เทียบรหัสตามตัวอักษร ช่วงจะรวมบัญชีที่เพิ่มภายหลังด้วย) บัญชีหนึ่งอยู่ได้หลายกลุ่ม")}</p>
      {entries.map((entry, index) => {
        const isRange = entry.to !== undefined;
        const label = isRange ? tr("gl_group_range_from", "จากรหัสบัญชี") : tr("gl_account_code", "รหัสบัญชี");
        return (
          <div key={index} className="grid items-end gap-2 rounded-xl border border-border bg-muted/20 p-2 shadow-sm sm:grid-cols-[1fr_1fr_auto]">
            <div className={isRange ? "" : "sm:col-span-2"}>
              <Field label={label}>
                <AccountSelect field={`accountentries[${index}].from`} label={label} all value={entry.from} accounts={accounts} onChange={(from) => update(index, { from })} />
              </Field>
            </div>
            {isRange && (
              <Field label={tr("gl_group_range_to", "ถึงรหัสบัญชี")}>
                <AccountSelect field={`accountentries[${index}].to`} label={tr("gl_group_range_to", "ถึงรหัสบัญชี")} all value={entry.to ?? ""} accounts={accounts} onChange={(to) => update(index, { to })} />
              </Field>
            )}
            <Button
              type="button"
              variant="outline"
              className={actionClass}
              aria-label={`${tr("gl_remove", "นำออก")} ${entry.from}${entry.to ? `:${entry.to}` : ""}`}
              onClick={() => onChange(entries.filter((_, i) => i !== index))}
            >
              {tr("gl_remove", "นำออก")}
            </Button>
          </div>
        );
      })}
      {pickerOpen && (
        <AccountSearchDialog
          open
          onClose={() => setPickerOpen(false)}
          accounts={accounts}
          multiSelect
          selectedCodes={singles}
          onSelectMultiple={(codes) => onChange(replaceAccountGroupSingles(entries, codes))}
          title={tr("gl_select_coa_for", "เลือกผังบัญชีสำหรับ \"{0}\"").replace("{0}", groupLabel)}
        />
      )}
    </section>
  );
}
