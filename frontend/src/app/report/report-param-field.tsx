"use client";

import { useMemo } from "react";
import { Combobox, type ComboboxOption } from "@/components/ui/combobox";
import { normalizeDimensionCodes, reportCodeList, type GLAccount, type GLDimension, type GLFiscalYear, type GLTextFn } from "@/lib/general-ledger";
import type { ReportParam } from "@/lib/report-registry";
import { AccountSelect, Check, Field, YearSelect, control, useGLText } from "@/app/gl/gl-common";
import { DimensionCodeMultiSelect } from "@/app/gl/gl-dimension-select";

export type ReportBudgetOption = { code: string; name: string; fiscalyear: string };
/** Registers the parameter editors pick from (loaded once by the screen). Books are the journal books shaped as register entries. */
export type ReportLookups = {
  accounts: GLAccount[];
  years: GLFiscalYear[];
  books: GLDimension[];
  departments: GLDimension[];
  projects: GLDimension[];
  budgets: ReportBudgetOption[];
};
export const EMPTY_REPORT_LOOKUPS: ReportLookups = { accounts: [], years: [], books: [], departments: [], projects: [], budgets: [] };

type Props = {
  param: ReportParam;
  value: string;
  /** The backend named this parameter in its last error: the editor is outlined (the screen also focuses it). */
  invalid: boolean;
  lookups: ReportLookups;
  /** Fiscal year chosen on the form: budgets of other years are not offered. */
  fiscalYear: string;
  onChange: (value: string) => void;
};

const INVALID_CLASS = "rounded-xl ring-2 ring-destructive/70 ring-offset-2 ring-offset-background";

/** One editor of a registry parameter: the editor follows the parameter type and its lookup, never the report. */
export function ReportParamField({ param, value, invalid, lookups, fiscalYear, onChange }: Props) {
  const tr = useGLText();
  const pickable = useMemo(() => singleCodeOptions(param, lookups, fiscalYear, tr), [param, lookups, fiscalYear, tr]);

  function editor() {
    if (param.type === "date") return <input className={control} type="date" required={param.required} value={value} onChange={(event) => onChange(event.target.value)} />;
    if (param.type === "year" || param.lookup === "fiscal_year") return <YearSelect years={lookups.years} label={param.label} value={value} onChange={onChange} />;
    if (param.type === "enum") return <Combobox<string> aria-label={param.label} value={value} options={param.options.map((option) => ({ value: option.value, label: option.label }))} onChange={onChange} />;
    if (param.lookup === "account") return <AccountSelect accounts={lookups.accounts} all label={param.label} value={value} onChange={onChange} />;
    if (param.type === "codes" && (param.lookup === "department" || param.lookup === "project" || param.lookup === "book")) {
      const resource = param.lookup === "department" ? "departments" : param.lookup === "project" ? "projects" : "books";
      const register = param.lookup === "department" ? lookups.departments : param.lookup === "project" ? lookups.projects : lookups.books;
      return <DimensionCodeMultiSelect report resource={resource} value={reportCodeList(value)} register={register} onChange={(codes) => onChange(normalizeDimensionCodes(codes).join(","))} />;
    }
    if (pickable.length) return <Combobox<string> aria-label={param.label} value={value} options={pickable} onChange={onChange} searchable={pickable.length > 8} placeholder={tr("gl_select", "เลือก")} />;
    if (param.lookup === "branch") return <input className={control} value={value} maxLength={20} onChange={(event) => onChange(event.target.value)} placeholder={tr("gl_all_branches", "ทุกสาขา")} />;
    return <input className={control} value={value} maxLength={param.type === "codes" ? 200 : 30} required={param.required} onChange={(event) => onChange(event.target.value)} />;
  }

  const body = param.type === "bool"
    ? <Check label={param.label} checked={value === "true"} onChange={(on) => onChange(on ? "true" : "false")} />
    : <Field label={param.label}>{editor()}</Field>;
  return <div data-field={param.key} data-invalid={invalid ? "true" : undefined} className={`min-w-0 ${invalid ? INVALID_CLASS : ""}`}>{body}</div>;
}

/** Single-code pickers over a register (department / project / book / budget); the first option "" = no limit. Empty list = not a register lookup. */
export function singleCodeOptions(param: Pick<ReportParam, "type" | "lookup">, lookups: ReportLookups, fiscalYear: string, tr: GLTextFn): ComboboxOption<string>[] {
  if (param.type !== "code") return [];
  const named = (items: { code: string; name: string }[]) => items.map((item) => ({ value: item.code, label: item.name ? `${item.code} · ${item.name}` : item.code }));
  if (param.lookup === "department") return [{ value: "", label: tr("gl_all_departments", "ทุกแผนก") }, ...named(lookups.departments)];
  if (param.lookup === "project") return [{ value: "", label: tr("gl_all_projects", "ทุกโครงการ") }, ...named(lookups.projects)];
  if (param.lookup === "book") return [{ value: "", label: tr("gl_report_all_books", "ทุกสมุดรายวัน") }, ...named(lookups.books)];
  if (param.lookup === "budget") return [{ value: "", label: tr("gl_all_budgets", "ทุกงบประมาณ") }, ...named(lookups.budgets.filter((budget) => !fiscalYear || budget.fiscalyear === fiscalYear))];
  return [];
}
