"use client";

import { Building2, GitBranch, Loader2, Plus, Trash2 } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
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

function normalizeHoldingScopeRule(value: unknown): HoldingScopeRule {
  if (typeof value === "string") {
    const businessCode = normalizeBusinessCode(value);
    return businessCode
      ? { scopetype: "company", businesscode: businessCode, allbranches: true }
      : { scopetype: "holding" };
  }
  const record = isRecord(value) ? value : {};
  const scopeType = normalizeHoldingScopeType(record.scopetype ?? record.scopeType);
  const businessCode = normalizeBusinessCode(record.businesscode ?? record.businessCode ?? record.companycode ?? record.companyCode);
  const branchCode = normalizeScopeBranchCode(record.branchcode ?? record.branchCode ?? record.code);
  if (scopeType === "holding" || !businessCode) return { scopetype: "holding", allbranches: false };
  if (scopeType === "company" || booleanLikeValue(record.allbranches ?? record.allBranches)) {
    return { scopetype: "company", businesscode: businessCode, allbranches: true };
  }
  return { scopetype: "branch", businesscode: businessCode, branchcode: branchCode, allbranches: false };
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

function normalizeHoldingScopeType(value: unknown): HoldingScopeType {
  const text = stringValue(value).toLowerCase();
  if (text === "company" || text === "business") return "company";
  if (text === "branch") return "branch";
  return "holding";
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
  setForm?: (form: FormState) => void;
  workspace: WorkspaceSession | null;
}) {
  const [companies, setCompanies] = useState<CompanyScopeOption[]>([]);
  const [branches, setBranches] = useState<BranchOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [companyAddCode, setCompanyAddCode] = useState("");
  const [activeBusinessCode, setActiveBusinessCode] = useState("");
  const [branchAddCodes, setBranchAddCodes] = useState<Record<string, string>>({});
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

    void fetch(`/api/workspace/holdings${params.size ? `?${params.toString()}` : ""}`, {
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
            : language === "th"
              ? "โหลดบริษัท/สาขาไม่สำเร็จ"
              : "Failed to load companies and branches",
        );
        setLoading(false);
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [auth, language, workspace]);

  useEffect(() => {
    if (holdingSelected || selectedCompanyScopes.length === 0) {
      if (activeBusinessCode) setActiveBusinessCode("");
      return;
    }
    if (!selectedCompanyScopes.some((scope) => scope.company.businesscode === activeBusinessCode)) {
      setActiveBusinessCode(selectedCompanyScopes[0].company.businesscode);
    }
  }, [activeBusinessCode, holdingSelected, selectedCompanyScopes]);

  function commit(nextRules: HoldingScopeRule[]) {
    if (readOnly || !setForm) return;
    setForm({ ...form, [field.key]: normalizeHoldingScopeRules(nextRules) });
  }

  function setHoldingScope(enabled: boolean) {
    if (enabled) {
      commit(holdingSelected ? rules : [{ scopetype: "holding", allbranches: false }, ...rules]);
      return;
    }
    commit(rules.filter((rule) => rule.scopetype !== "holding"));
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
    setCompanyAddCode("");
    setActiveBusinessCode(normalizedBusinessCode);
  }

  function removeCompanyScope(businessCode: string) {
    const normalizedBusinessCode = normalizeBusinessCode(businessCode);
    commit(
      rules.filter(
        (rule) => rule.scopetype === "holding" || rule.businesscode !== normalizedBusinessCode,
      ),
    );
    setBranchAddCodes((current) => {
      const next = { ...current };
      delete next[normalizedBusinessCode];
      return next;
    });
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
    setBranchAddCodes((current) => ({ ...current, [normalizedBusinessCode]: "" }));
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
    language === "th"
      ? "ติ๊กทั้งกลุ่มกิจการ หรือค้นหาบริษัทเพื่อเพิ่มเข้า list แล้วเลือกบริษัทเพื่อกำหนดสาขา"
      : "Select the whole business group, or search and add companies, then pick a company to configure branches.";
  const summary =
    holdingSelected
      ? language === "th"
        ? "ทั้งกลุ่มกิจการ"
        : "Whole business group"
      : language === "th"
        ? `เลือก ${selectedCompanyCount()} บริษัท / ${selectedBranchCount()} สาขา`
        : `${selectedCompanyCount()} companies / ${selectedBranchCount()} branches selected`;

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
          {language === "th" ? "กำลังโหลดบริษัท/สาขา…" : "Loading companies and branches…"}
        </p>
      ) : null}
      {error ? <p className="text-xs font-semibold text-destructive">{error}</p> : null}
      {!loading && !error && companies.length === 0 ? (
        <p className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
          {language === "th" ? "ยังไม่มีบริษัทให้เลือก" : "No companies available."}
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
            <span>{language === "th" ? "ใช้ได้ทั้งกลุ่มกิจการ" : "Apply to whole business group"}</span>
            <span className="text-xs font-normal text-muted-foreground">
              {language === "th"
                ? "ถ้าเลือกข้อนี้ ผู้ใช้งานหรือสิทธิ์นี้ใช้ได้ทุกบริษัทและทุกสาขา"
                : "When checked, this user or permission applies to every company and branch."}
            </span>
          </span>
        </label>
        <div className="grid gap-3">
            <div className="grid gap-3 rounded-xl border border-border bg-background p-3">
              <div className="grid gap-2">
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <p className="text-sm font-bold">{language === "th" ? "บริษัทที่เลือก" : "Selected companies"}</p>
                    <p className="text-xs text-muted-foreground">
                      {language === "th" ? "ค้นหาบริษัท แล้วกดเพิ่มเข้า list" : "Search a company, then add it to the list."}
                    </p>
                  </div>
                  <Badge variant="outline">{selectedCompanyScopes.length}</Badge>
                </div>
                {!readOnly ? (
                  <div className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
                    <CompanyScopeSearchPicker
                      companies={companies}
                      disabled={readOnly || loading || companies.length === 0}
                      language={language}
                      value={companyAddCode}
                      onChange={setCompanyAddCode}
                    />
                    <Button
                      type="button"
                      size="sm"
                      onClick={() => addCompanyScope(companyAddCode)}
                      disabled={!companyAddCode}
                    >
                      <Plus className="size-3.5" />
                      {language === "th" ? "เพิ่ม" : "Add"}
                    </Button>
                  </div>
                ) : null}
              </div>
              {selectedCompanyScopes.length === 0 ? (
                <p className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
                  {language === "th" ? "ยังไม่ได้เลือกบริษัท" : "No companies selected."}
                </p>
              ) : (
                <div className="grid gap-2">
                  {selectedCompanyScopes.map((scope) => {
                    const active = scope.company.businesscode === activeCompanyScope?.company.businesscode;
                    const branchLabel =
                      scope.allbranches
                        ? language === "th"
                          ? "ทุกสาขา"
                          : "All branches"
                        : language === "th"
                          ? `${scope.branches.length} สาขา`
                          : `${scope.branches.length} branches`;
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
                        {language === "th"
                          ? "เลือกทุกสาขา หรือค้นหาแล้วเพิ่มเฉพาะสาขาที่ใช้ได้"
                          : "Select all branches, or search and add specific branches."}
                      </p>
                    </div>
                    <Badge variant={activeCompanyScope.allbranches ? "success" : "outline"}>
                      {activeCompanyScope.allbranches
                        ? language === "th"
                          ? "ทุกสาขา"
                          : "All branches"
                        : language === "th"
                          ? `${activeCompanyScope.branches.length} สาขา`
                          : `${activeCompanyScope.branches.length} branches`}
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
                    {language === "th" ? "ใช้กับทุกสาขาในบริษัทนี้" : "Apply to all branches in this company"}
                  </label>
                  <div className="grid gap-2">
                    {activeCompanyScope.allbranches ? (
                      <p className="rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-xs font-semibold text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950 dark:text-emerald-300">
                        {language === "th"
                          ? "เลือกใช้ได้ทุกสาขาในบริษัทนี้แล้ว ไม่ต้องเลือกสาขาทีละสาขา"
                          : "All branches in this company are enabled. No need to pick branches one by one."}
                      </p>
                    ) : (
                      <>
                        {!readOnly ? (
                          <div className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
                            <BranchScopeSearchPicker
                              branches={branches.filter(
                                (branch) => branch.businesscode === activeCompanyScope.company.businesscode,
                              )}
                              disabled={readOnly || loading}
                              language={language}
                              value={branchAddCodes[activeCompanyScope.company.businesscode] ?? ""}
                              onChange={(branchCode) =>
                                setBranchAddCodes((current) => ({
                                  ...current,
                                  [activeCompanyScope.company.businesscode]: branchCode,
                                }))
                              }
                            />
                            <Button
                              type="button"
                              size="sm"
                              onClick={() =>
                                addBranchScope(
                                  activeCompanyScope.company.businesscode,
                                  branchAddCodes[activeCompanyScope.company.businesscode] ?? "",
                                )
                              }
                              disabled={!branchAddCodes[activeCompanyScope.company.businesscode]}
                            >
                              <Plus className="size-3.5" />
                              {language === "th" ? "เพิ่มสาขา" : "Add branch"}
                            </Button>
                          </div>
                        ) : null}
                        {activeCompanyScope.branches.length === 0 ? (
                          <p className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
                            {language === "th" ? "ยังไม่ได้เลือกสาขา" : "No branches selected."}
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
                                    aria-label={language === "th" ? "เอาสาขาออก" : "Remove branch"}
                                    title={language === "th" ? "เอาสาขาออก" : "Remove branch"}
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
                  {language === "th"
                    ? "เลือกบริษัทจาก list ด้านซ้ายก่อน แล้วระบบจะแสดงสาขาที่กำหนด"
                    : "Pick a company from the list to configure its branches."}
                </p>
              )}
            </div>
        </div>
      </div>
    </section>
  );
}

// ---------------------------------------------------------------------------
// CompanyScopeSearchPicker
// ---------------------------------------------------------------------------

export function CompanyScopeSearchPicker({
  companies,
  disabled,
  language,
  onChange,
  value,
}: {
  companies: CompanyScopeOption[];
  disabled?: boolean;
  language: LanguageCode;
  onChange: (businessCode: string) => void;
  value: string;
}) {
  const selected = companies.find((company) => company.businesscode === normalizeBusinessCode(value));
  const selectedLabel = selected ? `${selected.businesscode} - ${selected.name}` : value;
  const [query, setQuery] = useState(selectedLabel);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    setQuery(selectedLabel);
  }, [selectedLabel]);

  const filteredCompanies = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    const searchAll = !normalizedQuery || normalizedQuery === selectedLabel.toLowerCase();
    return companies
      .filter((company) => {
        if (searchAll) return true;
        return `${company.businesscode} ${company.name}`.toLowerCase().includes(normalizedQuery);
      })
      .slice(0, 30);
  }, [companies, query, selectedLabel]);

  return (
    <label className="relative grid gap-1 text-xs font-semibold">
      <span>{language === "th" ? "บริษัท" : "Company"}</span>
      <Input
        className="h-9 text-sm"
        disabled={disabled}
        placeholder={language === "th" ? "ค้นหารหัสหรือชื่อบริษัท" : "Search company code or name"}
        value={query}
        onBlur={() => window.setTimeout(() => setOpen(false), 120)}
        onChange={(event) => {
          setQuery(event.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
      />
      {open && !disabled ? (
        <div className="absolute left-0 right-0 top-full z-30 mt-1 max-h-56 overflow-y-auto rounded-xl border border-border bg-card p-1 shadow-lg">
          {filteredCompanies.length === 0 ? (
            <p className="px-2 py-1.5 text-xs text-muted-foreground">
              {language === "th" ? "ไม่พบบริษัท" : "No companies found"}
            </p>
          ) : (
            filteredCompanies.map((company) => (
              <button
                className={cn(
                  "flex w-full min-w-0 items-center gap-2 rounded-lg px-2 py-1.5 text-left text-xs font-semibold hover:bg-muted",
                  company.businesscode === selected?.businesscode && "bg-primary/10 text-primary",
                )}
                key={company.businesscode}
                type="button"
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => {
                  onChange(company.businesscode);
                  setQuery(`${company.businesscode} - ${company.name}`);
                  setOpen(false);
                }}
              >
                <Building2 className="size-3.5 shrink-0" />
                <span className="min-w-0 flex-1 truncate">
                  {company.businesscode} - {company.name}
                </span>
              </button>
            ))
          )}
        </div>
      ) : null}
    </label>
  );
}

// ---------------------------------------------------------------------------
// BranchScopeSearchPicker
// ---------------------------------------------------------------------------

export function BranchScopeSearchPicker({
  branches,
  disabled,
  language,
  onChange,
  value,
}: {
  branches: BranchOption[];
  disabled?: boolean;
  language: LanguageCode;
  onChange: (branchCode: string) => void;
  value: string;
}) {
  const normalizedValue = normalizeScopeBranchCode(value);
  const selected = branches.find((branch) => normalizeScopeBranchCode(branch.code) === normalizedValue);
  const selectedLabel = selected ? `${selected.code} - ${branchOptionDisplayName(selected, language)}` : value;
  const [query, setQuery] = useState(selectedLabel);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    setQuery(selectedLabel);
  }, [selectedLabel]);

  const filteredBranches = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    const searchAll = !normalizedQuery || normalizedQuery === selectedLabel.toLowerCase();
    return branches
      .filter((branch) => {
        if (searchAll) return true;
        return `${branch.code} ${branchOptionDisplayName(branch, language)}`.toLowerCase().includes(normalizedQuery);
      })
      .slice(0, 30);
  }, [branches, language, query, selectedLabel]);

  return (
    <label className="relative grid gap-1 text-xs font-semibold">
      <span>{language === "th" ? "สาขา" : "Branch"}</span>
      <Input
        className="h-9 text-sm"
        disabled={disabled}
        placeholder={language === "th" ? "ค้นหารหัสหรือชื่อสาขา" : "Search branch code or name"}
        value={query}
        onBlur={() => window.setTimeout(() => setOpen(false), 120)}
        onChange={(event) => {
          setQuery(event.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
      />
      {open && !disabled ? (
        <div className="absolute left-0 right-0 top-full z-30 mt-1 max-h-56 overflow-y-auto rounded-xl border border-border bg-card p-1 shadow-lg">
          {filteredBranches.length === 0 ? (
            <p className="px-2 py-1.5 text-xs text-muted-foreground">
              {language === "th" ? "ไม่พบสาขา" : "No branches found"}
            </p>
          ) : (
            filteredBranches.map((branch) => (
              <button
                className={cn(
                  "flex w-full min-w-0 items-center gap-2 rounded-lg px-2 py-1.5 text-left text-xs font-semibold hover:bg-muted",
                  normalizeScopeBranchCode(branch.code) === normalizedValue && "bg-primary/10 text-primary",
                )}
                key={branchKeyOf(branch)}
                type="button"
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => {
                  onChange(normalizeScopeBranchCode(branch.code));
                  setQuery(`${branch.code} - ${branchOptionDisplayName(branch, language)}`);
                  setOpen(false);
                }}
              >
                <GitBranch className="size-3.5 shrink-0" />
                <span className="min-w-0 flex-1 truncate">
                  {branch.code} - {branchOptionDisplayName(branch, language)}
                </span>
              </button>
            ))
          )}
        </div>
      ) : null}
    </label>
  );
}
