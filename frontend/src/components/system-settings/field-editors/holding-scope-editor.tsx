"use client";

import { authFetch } from "@/lib/client-auth-session";
import { useBackendText } from "@/components/backend-text-provider";
import { Building2, GitBranch, Loader2, Plus, Trash2 } from "lucide-react";
import { useEffect, useMemo, useRef, useState, type CSSProperties, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { normalizeBusinessCode } from "@/lib/business-code";
import { normalizeThaiTaxBranchCode } from "@/lib/thai-branch-code";
import type { SystemSettingField } from "@/lib/system-setting-screens";
import type { LanguageCode } from "@/lib/i18n";
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";
import { cn } from "@/lib/utils";
import {
  type FormState,
  type SettingRecord,
  booleanLikeValue,
  extractListRecords,
  isRecord,
  localizedValue,
  requestHeaders,
  safeJsonParse,
  stringValue,
} from "../types";
import {
  type BranchOption,
  branchKeyOf,
  branchOptionDisplayName,
  recordToBranchOption,
} from "./branch-company-selectors";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type HoldingScopeType = "holding" | "company" | "branch";

export type HoldingScopeRule = {
  scopetype: HoldingScopeType;
  businesscode?: string;
  branchcode?: string;
  allbranches?: boolean;
};

export type CompanyScopeOption = {
  businesscode: string;
  guidfixed: string;
  name: string;
};

export type SelectedCompanyScope = {
  company: CompanyScopeOption;
  allbranches: boolean;
  branches: BranchOption[];
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

export function normalizeScopeBranchCode(value: unknown): string {
  const text = stringValue(value);
  if (!text) return "";
  try {
    return normalizeThaiTaxBranchCode(text);
  } catch {
    return text.trim();
  }
}

export function recordToCompanyScopeOption(
  record: SettingRecord,
  language: LanguageCode,
): CompanyScopeOption | null {
  const businessCode = normalizeBusinessCode(record.businesscode ?? record.code);
  const guidFixed = stringValue(record.guidfixed ?? record.guidfixed ?? record.guid ?? record.guidFixed);
  if (!businessCode) return null;
  return {
    businesscode: businessCode,
    guidfixed: guidFixed,
    name:
      localizedValue(record.names, language) ||
      stringValue(record.name1 ?? record.name ?? record.company_name ?? record.companyname) ||
      businessCode,
  };
}

export function selectedCompanyScopesFromRules(
  rules: HoldingScopeRule[],
  companies: CompanyScopeOption[],
  branches: BranchOption[],
): SelectedCompanyScope[] {
  const companyByCode = new Map(companies.map((company) => [company.businesscode, company]));
  const branchByKey = new Map(
    branches.map((branch) => [
      `${normalizeBusinessCode(branch.businesscode)}|${normalizeScopeBranchCode(branch.code)}`,
      branch,
    ]),
  );
  const scopes = new Map<string, SelectedCompanyScope>();

  function ensureScope(businessCode: string): SelectedCompanyScope {
    const normalizedBusinessCode = normalizeBusinessCode(businessCode);
    const existing = scopes.get(normalizedBusinessCode);
    if (existing) return existing;
    const company =
      companyByCode.get(normalizedBusinessCode) ?? {
        businesscode: normalizedBusinessCode,
        guidfixed: "",
        name: normalizedBusinessCode,
      };
    const scope: SelectedCompanyScope = { company, allbranches: false, branches: [] };
    scopes.set(normalizedBusinessCode, scope);
    return scope;
  }

  for (const rule of rules) {
    if (rule.scopetype === "holding") continue;
    const businessCode = normalizeBusinessCode(rule.businesscode);
    if (!businessCode) continue;
    const scope = ensureScope(businessCode);
    if (rule.scopetype === "company" || rule.allbranches) {
      scope.allbranches = true;
      scope.branches = branches.filter(
        (branch) => normalizeBusinessCode(branch.businesscode) === businessCode,
      );
    } else if (rule.branchcode) {
      const branchKey = `${businessCode}|${normalizeScopeBranchCode(rule.branchcode)}`;
      const branch = branchByKey.get(branchKey);
      if (branch && !scope.branches.some((b) => branchKeyOf(b) === branchKeyOf(branch))) {
        scope.branches.push(branch);
      } else if (!branch) {
        scope.branches.push({
          guidfixed: "",
          code: rule.branchcode,
          names: [],
          businesscode: businessCode,
        });
      }
    }
  }

  return Array.from(scopes.values());
}

export function normalizeHoldingScopeRules(value: unknown, fallbackCompanies?: unknown): HoldingScopeRule[] {
  const raw = holdingScopeRawArray(value);
  const source = raw.length ? raw : holdingScopeRawArray(fallbackCompanies).map((item) => ({ scopetype: "company", businesscode: item }));
  const seen = new Set<string>();
  const result: HoldingScopeRule[] = [];
  for (const item of source) {
    const normalized = normalizeHoldingScopeRule(item);
    if (!normalized) continue;
    const key = `${normalized.scopetype}|${normalized.businesscode ?? ""}|${normalized.branchcode ?? ""}`;
    if (seen.has(key)) continue;
    seen.add(key);
    result.push(normalized);
  }
  return result;
}

function holdingScopeRawArray(value: unknown): unknown[] {
  if (Array.isArray(value)) return value;
  if (typeof value !== "string") return [];
  const trimmed = value.trim();
  if (!trimmed) return [];
  const parsed = safeJsonParse(trimmed, undefined);
  if (Array.isArray(parsed)) return parsed;
  return trimmed.split(",").map((item) => item.trim()).filter(Boolean);
}

// Only an explicit holding rule grants the whole group; a rule with an unknown type or
// without a company is dropped, never widened to holding.
function normalizeHoldingScopeRule(value: unknown): HoldingScopeRule | null {
  if (typeof value === "string") {
    const businessCode = normalizeBusinessCode(value);
    return businessCode ? { scopetype: "company", businesscode: businessCode, allbranches: true } : null;
  }
  const record = isRecord(value) ? value : {};
  const scopeType = normalizeHoldingScopeType(record.scopetype ?? record.scopeType);
  if (scopeType === "holding") return { scopetype: "holding" };
  // The company uid is the company code (PostgreSQL central schema).
  const businessCode = normalizeBusinessCode(
    record.businesscode ?? record.businessCode ?? record.companycode ?? record.companyCode ?? record.companyuid,
  );
  if (!scopeType || !businessCode) return null;
  const branchCode = normalizeScopeBranchCode(record.branchcode ?? record.branchCode ?? record.branchuid ?? record.code);
  if (scopeType === "company" || booleanLikeValue(record.allbranches ?? record.allBranches)) {
    return { scopetype: "company", businesscode: businessCode, allbranches: true };
  }
  return { scopetype: "branch", businesscode: businessCode, branchcode: branchCode, allbranches: false };
}

/**
 * Ticking "whole holding" replaces the company/branch picks with the single holding rule;
 * unticking brings back the picks it replaced (remembered by the editor), so a mis-tick does
 * not wipe them.
 */
export function toggleHoldingScopeRules(
  rules: HoldingScopeRule[],
  enabled: boolean,
  picksBeforeHolding: HoldingScopeRule[],
): { rules: HoldingScopeRule[]; picksBeforeHolding: HoldingScopeRule[] } {
  const picks = rules.filter((rule) => rule.scopetype !== "holding");
  if (enabled) return { rules: [{ scopetype: "holding" }], picksBeforeHolding: picks };
  return { rules: picks.length ? picks : picksBeforeHolding, picksBeforeHolding: [] };
}

export function hasInvalidHoldingScopeRules(value: unknown, required: boolean): boolean {
  const rules = normalizeHoldingScopeRules(value);
  if (required && rules.length === 0) return true;
  if (rules.some((rule) => rule.scopetype === "holding")) return false;
  return rules.some((rule) => {
    if (rule.scopetype === "holding") return false;
    if (!rule.businesscode) return true;
    return rule.scopetype === "branch" && !rule.branchcode;
  });
}

function normalizeHoldingScopeType(value: unknown): HoldingScopeType | "" {
  const text = stringValue(value).toLowerCase();
  if (text === "holding") return "holding";
  if (text === "company" || text === "business") return "company";
  if (text === "branch") return "branch";
  return "";
}

function workspaceHoldingCode(workspace: WorkspaceSession): string {
  return workspace.shop.holdingcode?.trim() || "";
}

// ---------------------------------------------------------------------------
// HoldingScopeRulesEditor
// ---------------------------------------------------------------------------

export function HoldingScopeRulesEditor({
  auth,
  field,
  form,
  label,
  language,
  readOnly = false,
  setForm,
  workspace,
}: {
  auth: AuthSession | null;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  readOnly?: boolean;
  setForm?: (update: FormState | ((current: FormState) => FormState)) => void;
  workspace: WorkspaceSession | null;
}) {
  const tr = useBackendText();
  const [companies, setCompanies] = useState<CompanyScopeOption[]>([]);
  const [branches, setBranches] = useState<BranchOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [activeBusinessCode, setActiveBusinessCode] = useState("");
  const rules = useMemo(
    () => normalizeHoldingScopeRules(form[field.key], form.businesscodes ?? form.companyguids),
    [field.key, form],
  );
  const holdingSelected = rules.some((rule) => rule.scopetype === "holding");
  const selectedCompanyScopes = useMemo(
    () => selectedCompanyScopesFromRules(rules, companies, branches),
    [branches, companies, rules],
  );
  const activeCompanyScope =
    selectedCompanyScopes.find((scope) => scope.company.businesscode === activeBusinessCode) ??
    selectedCompanyScopes[0] ??
    null;

  useEffect(() => {
    if (!auth || !workspace) return;
    const controller = new AbortController();
    let cancelled = false;
    setLoading(true);
    setError("");

    const activeHoldingCode = workspaceHoldingCode(workspace);
    const params = new URLSearchParams();
    if (activeHoldingCode) params.set("activeholdingcode", activeHoldingCode);

    void authFetch(`/api/workspace/holdings?management=true${params.size ? `&${params.toString()}` : ""}`, {
      headers: requestHeaders(auth),
      cache: "no-store",
      signal: controller.signal,
    })
      .then(async (response) => {
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return response.json() as Promise<unknown>;
      })
      .then(async (payload) => {
        if (cancelled) return;
        const holdingRecords = extractListRecords(payload);
        const activeHolding = activeHoldingCode.toLowerCase();
        const currentHolding =
          holdingRecords.find((record) => stringValue(record.holdingcode).toLowerCase() === activeHolding) ??
          holdingRecords[0] ??
          {};
        const rawCompanies = Array.isArray(currentHolding.companies)
          ? currentHolding.companies.filter(isRecord)
          : [];
        const rawBranches = Array.isArray(currentHolding.branches)
          ? currentHolding.branches.filter(isRecord)
          : [];
        const companyOptions = rawCompanies
          .map((record) => recordToCompanyScopeOption(record, language))
          .filter((item): item is CompanyScopeOption => item !== null);
        setCompanies(companyOptions);

        const companyByGuid = new Map(companyOptions.map((company) => [company.guidfixed, company]));
        const branchOptions = rawBranches
          .flatMap((record) => {
            const option = recordToBranchOption(record);
            const company = companyByGuid.get(option.companyguid ?? "");
            if (!company || (!option.code && !option.guidfixed)) return [];
            const scopedBranch: BranchOption = {
              ...option,
              businesscode: company.businesscode,
            };
            return [scopedBranch];
          })
          .sort((a, b) => {
            const companyCompare = stringValue(a.businesscode).localeCompare(stringValue(b.businesscode));
            if (companyCompare !== 0) return companyCompare;
            return stringValue(a.code).localeCompare(stringValue(b.code), undefined, { numeric: true, sensitivity: "base" });
          });
        if (!cancelled) {
          setBranches(branchOptions);
          setLoading(false);
        }
      })
      .catch((catchError: unknown) => {
        if (cancelled) return;
        if (catchError instanceof DOMException && catchError.name === "AbortError") return;
        setError(
          catchError instanceof Error && catchError.message
            ? catchError.message
            : tr("ss_failed_to_load_companies_and", "โหลดบริษัท/สาขาไม่สำเร็จ"),
        );
        setLoading(false);
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [auth, language, tr, workspace]);

  useEffect(() => {
    if (holdingSelected || selectedCompanyScopes.length === 0) {
      if (activeBusinessCode) setActiveBusinessCode("");
      return;
    }
    if (!selectedCompanyScopes.some((scope) => scope.company.businesscode === activeBusinessCode)) {
      setActiveBusinessCode(selectedCompanyScopes[0].company.businesscode);
    }
  }, [activeBusinessCode, holdingSelected, selectedCompanyScopes]);

  function commit(nextRules: HoldingScopeRule[]): HoldingScopeRule[] {
    const normalized = normalizeHoldingScopeRules(nextRules);
    if (readOnly || !setForm) return normalized;
    setForm((current) => ({ ...current, [field.key]: normalized }));
    return normalized;
  }

  // The picks replaced by the holding rule, tied to the exact value committed: if the form now
  // holds another record's rules, nothing is restored into it.
  const picksBeforeHoldingRef = useRef<{ picks: HoldingScopeRule[]; committed: unknown }>({ picks: [], committed: null });

  function setHoldingScope(enabled: boolean) {
    const remembered = picksBeforeHoldingRef.current;
    const picks = remembered.committed === form[field.key] ? remembered.picks : [];
    const next = toggleHoldingScopeRules(rules, enabled, picks);
    // The holding rule stands alone: it already covers every company and branch.
    const committed = commit(next.rules);
    picksBeforeHoldingRef.current = { picks: next.picksBeforeHolding, committed };
  }

  function addCompanyScope(businessCode: string) {
    const normalizedBusinessCode = normalizeBusinessCode(businessCode);
    if (!normalizedBusinessCode) return;
    const nextRules = [...rules];
    const hasCompany =
      nextRules.some(
        (rule) =>
          rule.businesscode === normalizedBusinessCode &&
          (rule.scopetype === "company" || rule.scopetype === "branch"),
      );
    if (!hasCompany) {
      nextRules.push({
        scopetype: "branch",
        businesscode: normalizedBusinessCode,
        branchcode: "",
        allbranches: false,
      });
    }
    commit(nextRules);
    setActiveBusinessCode(normalizedBusinessCode);
  }

  function removeCompanyScope(businessCode: string) {
    const normalizedBusinessCode = normalizeBusinessCode(businessCode);
    commit(
      rules.filter(
        (rule) => rule.scopetype === "holding" || rule.businesscode !== normalizedBusinessCode,
      ),
    );
  }

  function setCompanyAllBranches(businessCode: string, enabled: boolean) {
    const normalizedBusinessCode = normalizeBusinessCode(businessCode);
    if (!normalizedBusinessCode) return;
    const nextRules = rules.filter(
      (rule) => !(rule.scopetype === "company" && rule.businesscode === normalizedBusinessCode),
    );
    if (enabled) {
      nextRules.push({
        scopetype: "company",
        businesscode: normalizedBusinessCode,
        allbranches: true,
      });
    } else if (!nextRules.some((rule) => rule.scopetype === "branch" && rule.businesscode === normalizedBusinessCode)) {
      nextRules.push({
        scopetype: "branch",
        businesscode: normalizedBusinessCode,
        branchcode: "",
        allbranches: false,
      });
    }
    commit(nextRules);
  }

  function addBranchScope(businessCode: string, branchCode: string) {
    const normalizedBusinessCode = normalizeBusinessCode(businessCode);
    const normalizedBranchCode = normalizeScopeBranchCode(branchCode);
    if (!normalizedBusinessCode || !normalizedBranchCode) return;
    const nextRules = rules.filter(
      (rule) =>
        !(
          rule.scopetype === "branch" &&
          rule.businesscode === normalizedBusinessCode &&
          (!rule.branchcode || rule.branchcode === normalizedBranchCode)
        ),
    );
    nextRules.push({
      scopetype: "branch",
      businesscode: normalizedBusinessCode,
      branchcode: normalizedBranchCode,
      allbranches: false,
    });
    commit(nextRules);
  }

  function removeBranchScope(businessCode: string, branchCode: string) {
    const normalizedBusinessCode = normalizeBusinessCode(businessCode);
    const normalizedBranchCode = normalizeScopeBranchCode(branchCode);
    const nextRules = rules.filter(
      (rule) =>
        !(
          rule.scopetype === "branch" &&
          rule.businesscode === normalizedBusinessCode &&
          rule.branchcode === normalizedBranchCode
        ),
    );
    if (!nextRules.some((rule) => rule.businesscode === normalizedBusinessCode)) {
      nextRules.push({
        scopetype: "branch",
        businesscode: normalizedBusinessCode,
        branchcode: "",
        allbranches: false,
      });
    }
    commit(nextRules);
  }

  function selectedBranchCount() {
    if (holdingSelected) return branches.length;
    return selectedCompanyScopes.reduce((count, scope) => {
      if (scope.allbranches) {
        return count + branches.filter((branch) => branch.businesscode === scope.company.businesscode).length;
      }
      return count + scope.branches.length;
    }, 0);
  }

  function selectedCompanyCount() {
    if (holdingSelected) return companies.length;
    return new Set(
      rules
        .filter((rule) => rule.businesscode)
        .map((rule) => rule.businesscode as string),
    ).size;
  }

  const hint =
    tr("ss_select_the_whole_business_group", "ติ๊กทั้งกลุ่มกิจการ หรือค้นหาบริษัทเพื่อเพิ่มเข้า list แล้วเลือกบริษัทเพื่อกำหนดสาขา");
  const summary =
    holdingSelected
      ? tr("ss_whole_business_group", "ทั้งกลุ่มกิจการ")
      : tr("ss_selected_companies_branches", "เลือก {0} บริษัท / {1} สาขา")
          .replace("{0}", String(selectedCompanyCount()))
          .replace("{1}", String(selectedBranchCount()));

  return (
    <section className="grid gap-3 rounded-2xl border border-border bg-background p-3 md:col-span-2">
      <div className="flex flex-wrap items-start justify-between gap-2 border-b border-border pb-2">
        <div className="grid gap-1">
          <h3 className="text-sm font-bold">{label}{field.required ? " *" : ""}</h3>
          <p className="text-xs text-muted-foreground">{hint}</p>
        </div>
        <Badge variant={rules.length > 0 ? "success" : "outline"}>{summary}</Badge>
      </div>
      {loading ? (
        <p className="flex items-center gap-2 text-xs text-muted-foreground">
          <Loader2 className="size-3 animate-spin" />
          {tr("ss_loading_companies_and_branches", "กำลังโหลดบริษัท/สาขา…")}
        </p>
      ) : null}
      {error ? <p className="text-xs font-semibold text-destructive">{error}</p> : null}
      {!loading && !error && companies.length === 0 ? (
        <p className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
          {tr("ss_no_companies_available", "ยังไม่มีบริษัทให้เลือก")}
        </p>
      ) : null}
      <div className="grid gap-3 rounded-xl border border-border bg-muted/20 p-3">
        <label className="flex items-start gap-3 rounded-xl border border-border bg-background p-3 text-sm font-bold">
          <input
            className="mt-1 size-4 accent-primary"
            type="checkbox"
            checked={holdingSelected}
            disabled={readOnly}
            onChange={(event) => setHoldingScope(event.target.checked)}
          />
          <span className="grid gap-1">
            <span>{tr("ss_apply_to_whole_business_group", "ใช้ได้ทั้งกลุ่มกิจการ")}</span>
            <span className="text-xs font-normal text-muted-foreground">
              {tr("ss_when_checked_this_user_or", "ถ้าเลือกข้อนี้ ผู้ใช้งานหรือสิทธิ์นี้ใช้ได้ทุกบริษัทและทุกสาขา รวมถึงบริษัทและสาขาที่เพิ่มภายหลังโดยอัตโนมัติ")}
            </span>
          </span>
        </label>
        <div className="grid gap-3">
            <div className="grid gap-3 rounded-xl border border-border bg-background p-3">
              <div className="grid gap-2">
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <p className="text-sm font-bold">{tr("ss_selected_companies", "บริษัทที่เลือก")}</p>
                    <p className="text-xs text-muted-foreground">
                      {tr("ss_search_a_company_then_add", "ค้นหาบริษัท แล้วกดเพิ่มเข้า list")}
                    </p>
                  </div>
                  <Badge variant="outline">{selectedCompanyScopes.length}</Badge>
                </div>
                {!readOnly ? (
                  <div className="grid gap-2">
                    <CompanyScopeSearchPicker
                      companies={companies}
                      disabled={readOnly || loading || companies.length === 0}
                      onPick={(businessCode) => addCompanyScope(businessCode)}
                    />
                  </div>
                ) : null}
              </div>
              {selectedCompanyScopes.length === 0 ? (
                <p className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
                  {tr("ss_no_companies_selected", "ยังไม่ได้เลือกบริษัท")}
                </p>
              ) : (
                <div className="grid gap-2">
                  {selectedCompanyScopes.map((scope) => {
                    const active = scope.company.businesscode === activeCompanyScope?.company.businesscode;
                    const branchLabel =
                      scope.allbranches
                        ? tr("ss_all_branches", "ทุกสาขา")
                        : tr("ss_branch_count", "{0} สาขา").replace("{0}", String(scope.branches.length));
                    return (
                      <button
                        className={cn(
                          "grid w-full min-w-0 gap-1 rounded-xl border p-3 text-left text-xs transition hover:border-primary/50 hover:bg-primary/5",
                          active ? "border-primary bg-primary/10 text-primary" : "border-border bg-card",
                        )}
                        key={scope.company.businesscode}
                        type="button"
                        onClick={() => setActiveBusinessCode(scope.company.businesscode)}
                      >
                        <span className="flex min-w-0 items-center gap-2">
                          <Building2 className="size-4 shrink-0" />
                          <span className="min-w-0 flex-1 truncate text-sm font-bold">
                            {scope.company.businesscode} - {scope.company.name}
                          </span>
                          {!readOnly ? (
                            <span
                              role="button"
                              tabIndex={0}
                              className="inline-flex size-8 shrink-0 items-center justify-center rounded-lg border border-border bg-background text-destructive"
                              onClick={(event) => {
                                event.stopPropagation();
                                removeCompanyScope(scope.company.businesscode);
                              }}
                              onKeyDown={(event) => {
                                if (event.key === "Enter" || event.key === " ") {
                                  event.preventDefault();
                                  event.stopPropagation();
                                  removeCompanyScope(scope.company.businesscode);
                                }
                              }}
                            >
                              <Trash2 className="size-4" />
                            </span>
                          ) : null}
                        </span>
                        <span className="text-muted-foreground">{branchLabel}</span>
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
            <div className="grid gap-3 rounded-xl border border-border bg-background p-3">
              {activeCompanyScope ? (
                <>
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div className="min-w-0">
                      <p className="truncate text-sm font-bold">
                        {activeCompanyScope.company.businesscode} - {activeCompanyScope.company.name}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {tr("ss_select_all_branches_or_search", "เลือกทุกสาขา หรือค้นหาแล้วเพิ่มเฉพาะสาขาที่ใช้ได้")}
                      </p>
                    </div>
                    <Badge variant={activeCompanyScope.allbranches ? "success" : "outline"}>
                      {activeCompanyScope.allbranches
                        ? tr("ss_all_branches", "ทุกสาขา")
                        : tr("ss_branch_count", "{0} สาขา").replace("{0}", String(activeCompanyScope.branches.length))}
                    </Badge>
                  </div>
                  <label className="flex items-center gap-2 rounded-xl border border-border bg-card p-3 text-sm font-bold">
                    <input
                      className="size-4 accent-primary"
                      type="checkbox"
                      checked={activeCompanyScope.allbranches}
                      disabled={readOnly}
                      onChange={(event) =>
                        setCompanyAllBranches(activeCompanyScope.company.businesscode, event.target.checked)
                      }
                    />
                    {tr("ss_apply_to_all_branches_in", "ใช้กับทุกสาขาในบริษัทนี้")}
                  </label>
                  <div className="grid gap-2">
                    {activeCompanyScope.allbranches ? (
                      <p className="rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-xs font-semibold text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950 dark:text-emerald-300">
                        {tr("ss_all_branches_in_this_company", "เลือกใช้ได้ทุกสาขาในบริษัทนี้แล้ว ไม่ต้องเลือกสาขาทีละสาขา")}
                      </p>
                    ) : (
                      <>
                        {!readOnly ? (
                          <div className="grid gap-2">
                            <BranchScopeSearchPicker
                              branches={branches.filter(
                                (branch) => branch.businesscode === activeCompanyScope.company.businesscode,
                              )}
                              disabled={readOnly || loading}
                              language={language}
                              onPick={(branchCode) =>
                                addBranchScope(activeCompanyScope.company.businesscode, branchCode)
                              }
                            />
                          </div>
                        ) : null}
                        {activeCompanyScope.branches.length === 0 ? (
                          <p className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
                            {tr("ss_no_branches_selected", "ยังไม่ได้เลือกสาขา")}
                          </p>
                        ) : (
                          <div className="grid gap-2">
                            {activeCompanyScope.branches.map((branch) => (
                              <div
                                className="flex min-w-0 items-center gap-2 rounded-xl border border-border bg-card px-3 py-2 text-xs font-semibold"
                                key={branchKeyOf(branch)}
                              >
                                <GitBranch className="size-4 shrink-0 text-primary" />
                                <span className="min-w-0 flex-1 break-words">
                                  {branch.code} - {branchOptionDisplayName(branch, language)}
                                </span>
                                {!readOnly ? (
                                  <Button
                                    type="button"
                                    size="icon"
                                    variant="outline"
                                    aria-label={tr("ss_remove_branch", "เอาสาขาออก")}
                                    title={tr("ss_remove_branch", "เอาสาขาออก")}
                                    className="size-8 shrink-0 text-destructive"
                                    onClick={() => removeBranchScope(activeCompanyScope.company.businesscode, branch.code)}
                                  >
                                    <Trash2 className="size-4" />
                                  </Button>
                                ) : null}
                              </div>
                            ))}
                          </div>
                        )}
                      </>
                    )}
                  </div>
                </>
              ) : (
                <p className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
                  {tr("ss_pick_a_company_from_the", "เลือกบริษัทจาก list ด้านซ้ายก่อน แล้วระบบจะแสดงสาขาที่กำหนด")}
                </p>
              )}
            </div>
        </div>
      </div>
    </section>
  );
}

// ---------------------------------------------------------------------------
// PortalDropdownList — รายการ dropdown ที่ render ผ่าน portal ไปที่ body ด้วย
// position:fixed เพื่อไม่ให้ container overflow-y-auto (เช่น ส่วน "รายละเอียด"
// บนจอเมนู สูง ~440px) ตัดครึ่งล่างของ dropdown ทิ้ง — เคสจริง 2026-08-31
// พื้นที่ด้านล่างไม่พอจะเปิด "ขึ้นบน" (drop-up) แทน และตำแหน่งตาม scroll/resize
// ---------------------------------------------------------------------------

const PORTAL_DROPDOWN_MAX_H = 224; // max-h-56 + ระยะ mt-1

export function PortalDropdownList({ anchor, children }: { anchor: HTMLElement | null; children: ReactNode }) {
  const [style, setStyle] = useState<CSSProperties>({ visibility: "hidden" });

  useEffect(() => {
    if (!anchor) return;
    const place = () => {
      const rect = anchor.getBoundingClientRect();
      const below = window.innerHeight - rect.bottom;
      const openUp = below < PORTAL_DROPDOWN_MAX_H && rect.top > below;
      const left = Math.max(8, Math.min(rect.left, window.innerWidth - rect.width - 8));
      setStyle({
        position: "fixed",
        left,
        width: rect.width,
        visibility: "visible",
        ...(openUp ? { bottom: window.innerHeight - rect.top + 4 } : { top: rect.bottom + 4 }),
      });
    };
    place();
    window.addEventListener("scroll", place, true);
    window.addEventListener("resize", place);
    return () => {
      window.removeEventListener("scroll", place, true);
      window.removeEventListener("resize", place);
    };
  }, [anchor]);

  if (typeof document === "undefined") return null;
  return createPortal(
    <div
      className="z-50 max-h-56 overflow-y-auto rounded-xl border border-border bg-card p-1 shadow-lg"
      role="listbox"
      style={style}
    >
      {children}
    </div>,
    document.body,
  );
}

// ---------------------------------------------------------------------------
// CompanyScopeSearchPicker
// ---------------------------------------------------------------------------

// ScopePickerPopup — popup เลือกบริษัท/สาขา (แทน combobox แบบ dropdown —
// ลุงจืดขอ 2026-09-01): dialog มีช่องค้นหา + รายการใหญ่กดง่าย กดรายการ = เลือกทันที
function ScopePickerPopup({
  closeLabel,
  noMatchLabel,
  title,
  placeholder,
  options,
  open,
  onClose,
  onPick,
  selectedKey,
}: {
  closeLabel: string;
  noMatchLabel: string;
  title: string;
  placeholder: string;
  options: { key: string; label: string }[];
  open: boolean;
  onClose: () => void;
  onPick: (key: string) => void;
  selectedKey?: string;
}) {
  const [query, setQuery] = useState("");
  useEffect(() => {
    if (open) setQuery("");
  }, [open]);
  if (!open) return null;
  const normalizedQuery = query.trim().toLowerCase();
  const filtered = options
    .filter((option) => !normalizedQuery || option.label.toLowerCase().includes(normalizedQuery))
    .slice(0, 30);
  return (
    <div className="dialog-backdrop" role="presentation">
      <section className="line-login-dialog" aria-label={title} role="dialog" aria-modal="true">
        <div className="dialog-header">
          <div>
            <h2>{title}</h2>
          </div>
          <button className="icon-button dialog-close" type="button" onClick={onClose} aria-label={closeLabel}>
            ×
          </button>
        </div>
        <Input
          autoFocus
          className="h-9 text-sm"
          placeholder={placeholder}
          value={query}
          onChange={(event) => setQuery(event.target.value)}
        />
        <div className="grid gap-1 overflow-y-auto scrollbar-thin mt-2 max-h-[55vh] pr-1">
          {filtered.length === 0 ? (
            <p className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-sm text-muted-foreground">
              {noMatchLabel}
            </p>
          ) : (
            filtered.map((option) => (
              <button
                className={cn(
                  "flex w-full min-w-0 items-center gap-2 rounded-xl border px-3 py-2.5 text-left text-sm font-semibold transition hover:border-primary/50 hover:bg-primary/5",
                  option.key === selectedKey ? "border-primary bg-primary/10 text-primary" : "border-border bg-card",
                )}
                key={option.key}
                type="button"
                onClick={() => {
                  onPick(option.key);
                  onClose();
                }}
              >
                <Building2 className="size-4 shrink-0" />
                <span className="min-w-0 flex-1 truncate">{option.label}</span>
              </button>
            ))
          )}
        </div>
        <div className="line-dialog-actions">
          <button className="secondary-button" type="button" onClick={onClose}>
            {closeLabel}
          </button>
        </div>
      </section>
    </div>
  );
}

export function CompanyScopeSearchPicker({
  companies,
  disabled,
  onPick,
}: {
  companies: CompanyScopeOption[];
  disabled?: boolean;
  onPick: (businessCode: string) => void;
}) {
  const tr = useBackendText();
  const [popupOpen, setPopupOpen] = useState(false);

  return (
    <div className="grid gap-1 text-xs font-semibold">
      <Button type="button" variant="outline" className="h-9 justify-start text-sm" disabled={disabled} onClick={() => setPopupOpen(true)}>
        <Plus />
        {tr("st_add_company", "เพิ่มบริษัท")}
      </Button>
      <ScopePickerPopup
        closeLabel={tr("ss_close", "ปิด")}
        noMatchLabel={tr("ss_no_company_matched", "ไม่พบบริษัท")}
        title={tr("select_shop", "เลือกบริษัท")}
        placeholder={tr("ss_search_company_code_or_name", "ค้นหารหัสหรือชื่อบริษัท")}
        open={popupOpen}
        onClose={() => setPopupOpen(false)}
        onPick={(key) => onPick(key)}
        options={companies.map((company) => ({
          key: company.businesscode,
          label: `${company.businesscode} - ${company.name}`,
        }))}
      />
    </div>
  );
}

// ---------------------------------------------------------------------------
// BranchScopeSearchPicker
// ---------------------------------------------------------------------------

export function BranchScopeSearchPicker({
  branches,
  disabled,
  language,
  onPick,
}: {
  branches: BranchOption[];
  disabled?: boolean;
  language: LanguageCode;
  onPick: (branchCode: string) => void;
}) {
  const tr = useBackendText();
  const [popupOpen, setPopupOpen] = useState(false);

  return (
    <div className="grid gap-1 text-xs font-semibold">
      <Button type="button" variant="outline" className="h-9 justify-start text-sm" disabled={disabled} onClick={() => setPopupOpen(true)}>
        <Plus />
        {tr("add_branch", "เพิ่มสาขา")}
      </Button>
      <ScopePickerPopup
        closeLabel={tr("ss_close", "ปิด")}
        noMatchLabel={tr("ss_no_branches_found", "ไม่พบสาขา")}
        title={tr("ss_pick_branches", "เลือกสาขา")}
        placeholder={tr("ss_search_branch_code_or_name", "ค้นหารหัสหรือชื่อสาขา")}
        open={popupOpen}
        onClose={() => setPopupOpen(false)}
        onPick={(key) => onPick(key)}
        options={branches.map((branch) => ({
          key: normalizeScopeBranchCode(branch.code),
          label: `${branch.code} - ${branchOptionDisplayName(branch, language)}`,
        }))}
      />
    </div>
  );
}
