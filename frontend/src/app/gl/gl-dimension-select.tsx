"use client";

import { useEffect, useMemo, useState } from "react";
import { X } from "lucide-react";
import { Combobox, type ComboboxOption } from "@/components/ui/combobox";
import { glAllRecords } from "@/lib/general-ledger-api";
import { activeDimensions, dimensionOptions, normalizeDimensionCodes, type GLDimension, type GLDimensionResource, type GLTextFn } from "@/lib/general-ledger";
import { useGLText } from "./gl-common";

type Registers = { departments: GLDimension[]; projects: GLDimension[] };

// The registers are small lookups shared by the journal and the chart of accounts. A short-lived module cache keeps a screen
// that mounts often (every record opened in the chart of accounts) from reloading them; the register screen drops it after a save.
const FRESH_MS = 30_000;
let cache: (Registers & { at: number }) | null = null;
let inFlight: Promise<Registers> | null = null;

export function invalidateDimensionRegisters() {
  cache = null;
  inFlight = null;
}

function loadRegisters(): Promise<Registers> {
  if (!inFlight) {
    const request = Promise.all([glAllRecords<GLDimension>("departments", "", 5000), glAllRecords<GLDimension>("projects", "", 5000)])
      .then(([departments, projects]) => {
        if (inFlight === request) cache = { departments, projects, at: Date.now() };
        return { departments, projects };
      })
      .finally(() => { if (inFlight === request) inFlight = null; });
    inFlight = request;
  }
  return inFlight;
}

export function useDimensionRegisters() {
  const [registers, setRegisters] = useState<Registers>(() => cache ?? { departments: [], projects: [] });
  const [error, setError] = useState("");
  useEffect(() => {
    let active = true;
    if (cache && Date.now() - cache.at < FRESH_MS) return;
    loadRegisters()
      .then((next) => { if (active) { setRegisters(next); setError(""); } })
      .catch((e: Error) => { if (active) setError(e.message); });
    return () => { active = false; };
  }, []);
  return { departments: registers.departments, projects: registers.projects, error };
}

export function dimensionLabel(item: GLDimension, tr: GLTextFn): string {
  if (!item.isactive) return `${item.code} · ${tr("gl_dim_code_not_offered", "ไม่อยู่ในรายการที่เลือกได้")}`;
  return item.name ? `${item.code} · ${item.name}` : item.code;
}

/** A report reads history, so a closed code is not flagged — just "code · name". */
function plainLabel(item: Pick<GLDimension, "code" | "name">): string {
  return item.name ? `${item.code} · ${item.name}` : item.code;
}

const resourceTexts = {
  departments: { all: ["gl_all_departments", "ทุกแผนก"], label: ["gl_depart_codes_label", "แผนกที่ใช้ได้"], pick: ["gl_depart_codes_pick", "เลือกแผนกที่อนุญาต"], reportLabel: ["gl_department_code", "รหัสแผนก"], reportPick: ["gl_report_depart_pick", "เลือกแผนก"] },
  projects: { all: ["gl_all_projects", "ทุกโครงการ"], label: ["gl_project_codes_label", "โครงการที่ใช้ได้"], pick: ["gl_project_codes_pick", "เลือกโครงการที่อนุญาต"], reportLabel: ["gl_project_code", "รหัสโครงการ"], reportPick: ["gl_report_project_pick", "เลือกโครงการ"] },
} as const;

/** One journal line department / project: a searchable list of the register codes the line account allows (empty list = all). */
export function DimensionSelect({ value, onChange, register, allowed, label, disabled = false, title, field }: {
  value: string; onChange: (code: string) => void; register: GLDimension[]; allowed: string[]; label: string; disabled?: boolean; title?: string; field?: string;
}) {
  const tr = useGLText();
  const offered = useMemo(() => dimensionOptions(register, allowed, value), [register, allowed, value]);
  const options = useMemo<ComboboxOption<string>[]>(() => [
    ...(value ? [{ value: "", label: tr("gl_dim_clear_value", "ไม่ระบุ (ล้างค่า)") }] : []),
    ...offered.map((item) => ({ value: item.code, label: dimensionLabel(item, tr) })),
  ], [offered, value, tr]);
  const placeholder = disabled && !value ? "—" : offered.length === 0 ? tr("gl_dim_register_empty", "ยังไม่มีรหัสในทะเบียน") : tr("gl_select", "เลือก");
  return (
    <div title={title} className="min-w-0">
      <Combobox<string> aria-label={label} data-field={field} value={value} options={options} onChange={onChange} disabled={disabled} placeholder={placeholder} searchable={offered.length > 6} />
    </div>
  );
}

/** Allowed codes of an account for one register (Champ BCAccSeparate): chips of the chosen codes plus a picker; nothing chosen = every code.
 * report = the department/project filter of a report (Champ multi-select DEPART_LIST): every registered code can be picked, closed ones included, since reports read history. */
export function DimensionCodeMultiSelect({ resource, value, onChange, register, disabled = false, report = false }: {
  resource: GLDimensionResource; value: string[]; onChange: (codes: string[]) => void; register: GLDimension[]; disabled?: boolean; report?: boolean;
}) {
  const tr = useGLText();
  const texts = resourceTexts[resource];
  const chosen = useMemo(() => normalizeDimensionCodes(value), [value]);
  const byCode = useMemo(() => new Map(register.map((item) => [item.code, item])), [register]);
  const remaining = useMemo(() => activeDimensions(register, report).filter((item) => !chosen.includes(item.code)), [register, chosen, report]);
  const options = useMemo<ComboboxOption<string>[]>(() => remaining.map((item) => ({ value: item.code, label: report ? plainLabel(item) : dimensionLabel(item, tr) })), [remaining, tr, report]);
  const [labelKey, labelText] = report ? texts.reportLabel : texts.label;
  const [pickKey, pickText] = report ? texts.reportPick : texts.pick;
  const label = tr(labelKey, labelText);
  return (
    <div className="grid min-w-0 gap-1.5 text-[0.95rem] leading-normal">
      <span className="font-medium text-foreground">{label}</span>
      <div className="flex min-w-0 flex-wrap items-center gap-1.5" role="group" aria-label={label}>
        {chosen.length === 0 ? (
          <span className="rounded-lg border border-dashed border-border bg-muted/30 px-2.5 py-1 text-muted-foreground">{tr(texts.all[0], texts.all[1])}</span>
        ) : chosen.map((code) => {
          const item = byCode.get(code);
          const offered = report ? Boolean(item) : Boolean(item?.isactive);
          const text = report ? plainLabel(item ?? { code, name: "" }) : dimensionLabel(item ?? { code, name: "", isactive: false }, tr);
          return (
            <span key={code} className={`inline-flex max-w-full items-center gap-1 rounded-lg border px-2.5 py-1 ${offered ? "border-primary/40 bg-primary/10 text-foreground" : "border-dashed border-destructive/50 bg-destructive/5 text-foreground"}`}>
              <span className="min-w-0 [overflow-wrap:anywhere]">{text}</span>
              <button
                type="button"
                disabled={disabled}
                className="inline-flex size-7 min-h-0! min-w-0! shrink-0 items-center justify-center rounded-full p-0! text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
                aria-label={tr("gl_dim_remove_code", "เอารหัส {0} ออก").replace("{0}", code)}
                title={tr("gl_dim_remove_code", "เอารหัส {0} ออก").replace("{0}", code)}
                onClick={() => onChange(chosen.filter((item) => item !== code))}
              >
                <X aria-hidden className="size-3.5" />
              </button>
            </span>
          );
        })}
      </div>
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        <div className="min-w-48 flex-1">
          <Combobox<string>
            aria-label={tr(pickKey, pickText)}
            value=""
            options={options}
            onChange={(code) => { if (code) onChange(normalizeDimensionCodes([...chosen, code])); }}
            disabled={disabled || options.length === 0}
            placeholder={options.length === 0 && chosen.length === 0 ? tr("gl_dim_register_empty", "ยังไม่มีรหัสในทะเบียน") : tr(pickKey, pickText)}
            searchable={options.length > 6}
          />
        </div>
        {chosen.length > 0 && (
          <button
            type="button"
            disabled={disabled}
            className="rounded-xl border border-input bg-background px-3 py-1.5 font-medium text-foreground hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
            onClick={() => onChange([])}
          >
            {report ? tr("gl_report_dim_clear", "ล้างที่เลือก (ทุกรหัส)") : tr("gl_dim_clear_codes", "ล้างทั้งหมด (ใช้ได้ทุกรหัส)")}
          </button>
        )}
      </div>
    </div>
  );
}
