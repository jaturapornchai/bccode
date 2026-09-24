"use client";

import { useState } from "react";
import { Plus, Trash2 } from "lucide-react";
import { useBackendText } from "@/components/backend-text-provider";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { ChoiceSelect } from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { formatAmount } from "@/lib/general-ledger";
import { acceptsTyping, groupFields, type TaxFormAttachment, type TaxFormField, type TaxFormValues } from "@/lib/tax-forms";
import { cn } from "@/lib/utils";

// ช่องกรอกของแบบยื่นภาษี: ป้ายช่องเป็นถ้อยคำบนแบบฟอร์มทางการของกรมสรรพากร (มาจากสเปกแบบใน backend)
// ยอดเงินเก็บเป็น string ทศนิยม 2 ตำแหน่งไม่มีคอมมา ("125000.00") — จอแสดงคอมมาเมื่อไม่ได้แก้ช่องนั้นอยู่
// (ตัดต่อข้อความผ่าน formatAmount ไม่แปลงเป็นตัวเลข)

const MONEY_FORMATTED = /^-?[\d,]+\.\d{2}$/;
const displayMoney = (value: string) => {
  const formatted = formatAmount(value, 2);
  return MONEY_FORMATTED.test(formatted) ? formatted : value;
};

export type FieldInvalid = { key: string; row: number } | null;

// taxFormFieldId - id ของช่องในจอ (ใช้พาไปช่องที่ต้องแก้): หัวแบบ = <prefix>-<key>, แถวใบแนบที่ n = <prefix>-row<n>-<key>
export const taxFormFieldId = (prefix: string, key: string, row = 0) => (row > 0 ? `${prefix}-row${row}-${key}` : `${prefix}-${key}`);

// focusTaxFormField - เปิดหมวดที่พับอยู่ เลื่อนไปช่องตาม id แล้วโฟกัส (ผู้ใช้กดเองจึงเลื่อนได้ — นุ่มนวล ยกเว้นตั้งลดการเคลื่อนไหว)
// ช่องตัวเลือกแบบปุ่มไม่มี id → หา label[for] แล้วโฟกัสตัวเลือกแรกในช่องนั้น (id ที่เป็นกล่องครอบก็โฟกัสตัวเลือกแรกข้างใน)
const FOCUSABLE = "input, select, textarea, button";
export function focusTaxFormField(id: string): boolean {
  if (typeof document === "undefined") return false;
  const target = document.getElementById(id) ?? document.querySelector(`label[for="${CSS.escape(id)}"]`)?.parentElement ?? null;
  if (!target) return false;
  for (let d = target.closest("details"); d; d = d.parentElement?.closest("details") ?? null) d.open = true;
  const reduce = window.matchMedia?.("(prefers-reduced-motion: reduce)").matches;
  target.scrollIntoView({ block: "center", behavior: reduce ? "auto" : "smooth" });
  (target.matches(FOCUSABLE) ? target : target.querySelector<HTMLElement>(FOCUSABLE))?.focus({ preventScroll: true });
  return true;
}

type FieldInputProps = {
  field: TaxFormField;
  value: string;
  onChange: (value: string) => void;
  invalid?: boolean;
  compact?: boolean;
  id?: string;
};

export function TaxFormFieldInput({ field, value, onChange, invalid, compact, id }: FieldInputProps) {
  const tr = useBackendText();
  const [editing, setEditing] = useState(false);
  if (field.type === "check") {
    return (
      <Checkbox
        id={id}
        checked={value !== ""}
        onCheckedChange={(checked) => onChange(checked ? "1" : "")}
        aria-label={field.label}
        aria-invalid={invalid || undefined}
      />
    );
  }
  if (field.type === "choice") {
    const options = [{ value: "", label: tr("tax_form_no_choice", "ไม่เลือก") }, ...(field.options ?? []).map((o) => ({ value: o.value, label: o.label }))];
    return (
      <ChoiceSelect
        id={id}
        value={value}
        onChange={(next) => onChange(String(next))}
        options={options}
        radioThreshold={compact ? 0 : 4}
        aria-label={field.label}
      />
    );
  }
  const money = field.type === "money";
  return (
    <Input
      id={id}
      value={money && !editing ? displayMoney(value) : value}
      inputMode={money ? "decimal" : field.type === "text" ? undefined : "numeric"}
      aria-label={field.label}
      aria-invalid={invalid || undefined}
      title={field.note || field.label}
      className={cn("min-h-[2.4em]", money && "text-right tabular-nums", invalid && "border-destructive ring-2 ring-destructive/25")}
      onChange={(e) => {
        if (acceptsTyping(field.type, e.target.value)) onChange(e.target.value);
      }}
      onFocus={() => setEditing(true)}
      onBlur={() => {
        setEditing(false);
        if (money && value.trim()) {
          const formatted = formatAmount(value, 2);
          if (MONEY_FORMATTED.test(formatted)) onChange(formatted.replace(/,/g, ""));
        }
      }}
    />
  );
}

type FieldGroupsProps = {
  fields: TaxFormField[];
  values: TaxFormValues;
  onChange: (key: string, value: string) => void;
  invalid: FieldInvalid;
  search: string;
  idPrefix: string;
};

const matches = (f: TaxFormField, q: string) => !q || f.label.toLowerCase().includes(q) || f.key.includes(q) || (f.note ?? "").toLowerCase().includes(q);

// TaxFormFieldGroups - ช่องของแบบแบ่งตามหมวดบนกระดาษ; ค้นหาแล้วเปิดเฉพาะหมวดที่เจอ
export function TaxFormFieldGroups({ fields, values, onChange, invalid, search, idPrefix }: FieldGroupsProps) {
  const tr = useBackendText();
  const q = search.trim().toLowerCase();
  const groups = groupFields(fields.filter((f) => matches(f, q)));
  const openAll = q !== "" || fields.length <= 120;
  return (
    <div className="grid gap-3">
      {groups.map(({ group, fields: groupFieldsList }, gi) => (
        <details
          key={group}
          open={openAll || gi < 3 || groupFieldsList.some((f) => invalid?.row === 0 && invalid.key === f.key)}
          className="rounded-2xl border border-border bg-card shadow-[var(--shadow-card)] [&[open]>summary]:border-b"
        >
          <summary className="flex min-h-[2.8em] cursor-pointer items-center justify-between gap-2 border-border px-4 text-[0.95rem] font-semibold text-foreground">
            <span>{tr(`tax_form_group_${group}`, group)}</span>
            <span className="text-xs font-normal text-muted-foreground">{groupFieldsList.length}</span>
          </summary>
          <div className="grid gap-x-4 gap-y-3 p-4 [grid-template-columns:repeat(auto-fill,minmax(min(100%,17rem),1fr))]">
            {groupFieldsList.map((f) => {
              const id = taxFormFieldId(idPrefix, f.key);
              const bad = invalid?.row === 0 && invalid.key === f.key;
              return (
                <div key={f.key} className={cn("grid content-start gap-1", f.type === "choice" && (f.options?.length ?? 0) <= 3 && "col-span-full")}>
                  <label htmlFor={id} className="text-sm leading-[1.5] text-foreground [overflow-wrap:anywhere]">
                    {f.label}
                    {f.note ? <span className="block text-xs text-muted-foreground">{f.note}</span> : null}
                  </label>
                  <TaxFormFieldInput id={id} field={f} value={values[f.key] ?? ""} onChange={(v) => onChange(f.key, v)} invalid={bad} />
                </div>
              );
            })}
          </div>
        </details>
      ))}
    </div>
  );
}

type RowsProps = {
  attachment: TaxFormAttachment;
  rows: TaxFormValues[];
  onChange: (rows: TaxFormValues[]) => void;
  invalid: FieldInvalid;
  idPrefix?: string;
};

// TaxFormRowsTable - ใบแนบแบบตาราง (ภ.ง.ด.3/53/2/2ก, ภ.พ.30 หลายสาขา): ระบบแบ่งแผ่นและใส่ลำดับ/ยอดรวมแผ่นเอง
export function TaxFormRowsTable({ attachment, rows, onChange, invalid, idPrefix }: RowsProps) {
  const tr = useBackendText();
  const setCell = (i: number, key: string, value: string) => onChange(rows.map((r, j) => (j === i ? { ...r, [key]: value } : r)));
  return (
    <div className="grid gap-2">
      <p className="text-sm leading-[1.5] text-muted-foreground">
        {tr("tax_form_rows_hint", "ระบบแบ่งใบแนบแผ่นละ {n} รายการ และใส่ลำดับที่/ยอดรวมของแต่ละแผ่นให้เอง").replace("{n}", String(attachment.rowspersheet ?? 0))}
      </p>
      <div className="max-h-[60dvh] overflow-auto rounded-xl border border-border">
        <table className="w-max min-w-full border-collapse text-sm">
          <thead className="sticky top-0 z-10 bg-muted">
            <tr>
              <th className="px-2 py-2 text-left font-semibold">#</th>
              {attachment.columns.map((c) => (
                <th key={c.key} className="min-w-[9rem] max-w-[16rem] px-2 py-2 text-left align-bottom font-semibold leading-[1.45]" title={c.note || c.label}>
                  {c.label}
                </th>
              ))}
              <th className="px-2 py-2" aria-hidden="true" />
            </tr>
          </thead>
          <tbody>
            {rows.map((r, i) => (
              <tr key={i} className="border-t border-border align-top">
                <td className="px-2 py-1.5 text-muted-foreground tabular-nums">{i + 1}</td>
                {attachment.columns.map((c) => (
                  <td key={c.key} className={cn("px-1.5 py-1", c.type === "choice" ? "min-w-[12rem]" : c.type === "money" ? "min-w-[9rem]" : "min-w-[8rem]")}>
                    <TaxFormFieldInput
                      id={idPrefix ? taxFormFieldId(idPrefix, c.key, i + 1) : undefined}
                      field={c}
                      value={r[c.key] ?? ""}
                      onChange={(v) => setCell(i, c.key, v)}
                      invalid={invalid?.row === i + 1 && invalid.key === c.key}
                      compact
                    />
                  </td>
                ))}
                <td className="px-1.5 py-1">
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="gap-1 text-destructive"
                    onClick={() => onChange(rows.filter((_, j) => j !== i))}
                    title={tr("tax_form_remove_row", "ลบรายการ")}
                  >
                    <Trash2 />
                    {tr("tax_form_remove_row", "ลบรายการ")}
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div>
        <Button type="button" variant="outline" className="gap-2" onClick={() => onChange([...rows, {}])}>
          <Plus />
          {tr("tax_form_add_row", "เพิ่มรายการ")}
        </Button>
      </div>
    </div>
  );
}

type SheetsProps = {
  attachment: TaxFormAttachment;
  sheets: TaxFormValues[];
  onChange: (sheets: TaxFormValues[]) => void;
  invalid: FieldInvalid;
};

// TaxFormSheets - ใบแนบหนึ่งแผ่นต่อหนึ่งรายการ (ภ.ธ.40 สถานประกอบการที่ยื่นรวม)
export function TaxFormSheets({ attachment, sheets, onChange, invalid }: SheetsProps) {
  const tr = useBackendText();
  return (
    <div className="grid gap-3">
      {sheets.map((s, i) => (
        <section key={i} className="grid gap-2 rounded-2xl border border-border bg-background p-3">
          <header className="flex items-center justify-between gap-2">
            <h4 className="font-semibold">{tr("tax_form_sheet_no", "แผ่นที่ {n}").replace("{n}", String(i + 1))}</h4>
            <Button type="button" variant="outline" size="sm" className="gap-1 text-destructive" onClick={() => onChange(sheets.filter((_, j) => j !== i))}>
              <Trash2 />
              {tr("tax_form_remove_sheet", "ลบแผ่นนี้")}
            </Button>
          </header>
          <TaxFormFieldGroups
            fields={attachment.columns}
            values={s}
            onChange={(key, value) => onChange(sheets.map((x, j) => (j === i ? { ...x, [key]: value } : x)))}
            invalid={invalid?.row === i + 1 ? { key: invalid.key, row: 0 } : null}
            search=""
            idPrefix={`sheet${i}`}
          />
        </section>
      ))}
      <div>
        <Button type="button" variant="outline" className="gap-2" onClick={() => onChange([...sheets, {}])}>
          <Plus />
          {tr("tax_form_add_sheet", "เพิ่มแผ่นใบแนบ")}
        </Button>
      </div>
    </div>
  );
}
