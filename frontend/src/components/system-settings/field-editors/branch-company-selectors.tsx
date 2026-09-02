"use client";

import { authFetch } from "@/lib/client-auth-session";
import { BadgeCheck, Building2, GitBranch, Loader2, MapPin, Search, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { MapPickerDialog } from "@/components/map-picker-dialog";
import type { SystemSettingConfig, SystemSettingField } from "@/lib/system-setting-screens";
import type { LanguageCode } from "@/lib/i18n";
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";
import type { MasterEntry } from "@/lib/product-barcode/api";
import { cn } from "@/lib/utils";
import {
  type FormState,
  type SettingRecord,
  isRecord,
  stringValue,
  requestHeaders,
  extractListRecords,
} from "../types";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type BranchOption = {
  guidfixed: string;
  code: string;
  names: Array<{ code?: string; name?: string }>;
  businesscode?: string;
  companyguid?: string;
  holdingcode?: string;
  shopname?: string;
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

export function branchOptionDisplayName(
  option: BranchOption,
  language: LanguageCode,
): string {
  const names = option.names ?? [];
  const localized =
    names.find((item) => item.code?.toLowerCase() === language && item.name)
      ?.name ??
    names.find((item) => item.code?.toLowerCase() === "th" && item.name)
      ?.name ??
    names.find((item) => item.name)?.name;
  const branchName = (localized ?? option.code ?? option.guidfixed).trim();
  if (option.shopname) {
    return `${option.shopname} - ${branchName}`;
  }
  return branchName;
}

export function recordToBranchOption(record: SettingRecord): BranchOption {
  const namesRaw = Array.isArray(record.names) ? record.names : [];
  const names: BranchOption["names"] = namesRaw
    .filter(isRecord)
    .map((entry) => ({
      code: stringValue(entry.code),
      name: stringValue(entry.name),
    }));
  return {
    guidfixed: stringValue(record.guidfixed ?? record.guidfixed ?? record.guid ?? record.guidFixed),
    code: stringValue(record.code),
    companyguid: stringValue(record.companyguid ?? record.companyguid ?? record.companyGuid),
    names,
  };
}

export function selectedBranchesFromValue(value: unknown): BranchOption[] {
  if (!Array.isArray(value)) return [];
  return value
    .map((item) => {
      if (typeof item === "string") {
        const trimmed = item.trim();
        if (!trimmed) return null;
        return { guidfixed: trimmed, code: "", names: [] } as BranchOption;
      }
      if (!isRecord(item)) return null;
      const namesRaw = Array.isArray(item.names) ? item.names : [];
      const names: BranchOption["names"] = namesRaw
        .filter(isRecord)
        .map((entry) => ({
          code: stringValue(entry.code),
          name: stringValue(entry.name),
        }));
      const guid = stringValue(item.guidfixed ?? item.guidfixed ?? item.guid);
      const code = stringValue(item.code);
      if (!guid && !code) return null;
      return { guidfixed: guid, code, names } as BranchOption;
    })
    .filter((item): item is BranchOption => item !== null);
}

export function branchKeyOf(option: { guidfixed?: string; code?: string; businesscode?: string; holdingcode?: string }): string {
  const shopPrefix = option.businesscode || option.holdingcode ? `${option.businesscode ?? option.holdingcode}_` : "";
  const coreKey = stringValue(option.guidfixed) || stringValue(option.code);
  return `${shopPrefix}${coreKey}`;
}

function companyOptionDisplayName(
  option: { guidfixed?: string; code?: string; names?: any },
  language: LanguageCode,
): string {
  const names = option.names ?? [];
  const localized =
    names.find((item: any) => item.code?.toLowerCase() === language && item.name)
      ?.name ??
    names.find((item: any) => item.code?.toLowerCase() === "th" && item.name)
      ?.name ??
    names.find((item: any) => item.name)?.name;
  return (localized ?? option.code ?? option.guidfixed ?? "").trim();
}

function parseCoordinateValue(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const text = value.trim();
    if (!text) return null;
    const parsed = Number(text);
    return Number.isFinite(parsed) ? parsed : null;
  }
  return null;
}

// ---------------------------------------------------------------------------
// BranchMultiSelectFieldEditor
// ---------------------------------------------------------------------------

export function BranchMultiSelectFieldEditor({
  auth,
  field,
  form,
  label,
  language,
  setForm,
  workspace,
}: {
  auth: AuthSession | null;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (update: FormState | ((current: FormState) => FormState)) => void;
  workspace: WorkspaceSession | null;
}) {
  const [options, setOptions] = useState<BranchOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const selected = useMemo(() => {
    const rawSelected = selectedBranchesFromValue(form[field.key]);
    return rawSelected.map((item) => {
      const match = options.find((opt) => branchKeyOf(opt) === branchKeyOf(item));
      if (match) {
        return {
          ...item,
          holdingcode: match.holdingcode,
          shopname: match.shopname,
          names: match.names,
        };
      }
      return item;
    });
  }, [field.key, form, options]);

  useEffect(() => {
    if (!auth || !workspace) return;
    const controller = new AbortController();
    let cancelled = false;
    setLoading(true);
    setError("");

    let holdingcodes: string[] = [];
    const formCompanies = form.businesscodes ?? form.companyguids;
    if (Array.isArray(formCompanies) && formCompanies.length > 0) {
      holdingcodes = formCompanies
        .map((s) => (typeof s === "string" ? s.trim() : stringValue((s as any)?.guidfixed ?? (s as any)?.holdingcode ?? "")))
        .filter(Boolean);
    }
    if (holdingcodes.length === 0) {
      holdingcodes = [workspace.shop.holdingcode];
    }

    void authFetch(`/api/workspace/holdings?management=true`, {
      headers: requestHeaders(auth),
      cache: "no-store",
      signal: controller.signal,
    })
      .then(async (res) => {
        if (!res.ok) throw new Error("Load shops failed");
        return res.json() as Promise<any>;
      })
      .then(async (shopsPayload) => {
        if (cancelled) return;
        const shopsList = Array.isArray(shopsPayload.data) ? shopsPayload.data : [];
        const shopNameMap = new Map<string, string>();
        for (const s of shopsList) {
          shopNameMap.set(s.holdingcode, s.name1 || s.name || s.holdingcode);
        }

        const fetchPromises = holdingcodes.map(async (sid) => {
          const params = new URLSearchParams({
            limit: "1000",
            offset: "0",
            holdingcode: sid,
          });
          const response = await authFetch(`/api/system-settings/branch?${params.toString()}`, {
            headers: requestHeaders(auth),
            cache: "no-store",
            signal: controller.signal,
          });
          if (!response.ok) throw new Error(`HTTP ${response.status}`);
          const payload = await response.json();
          const records = extractListRecords(payload);
          return records
            .map(recordToBranchOption)
            .filter((opt) => opt.guidfixed || opt.code)
            .map((opt) => ({
              ...opt,
              holdingcode: sid,
              shopname: shopNameMap.get(sid) || sid,
            }));
        });

        const allBranchesResults = await Promise.all(fetchPromises);
        const mergedBranches = allBranchesResults.flat();

        if (!cancelled) {
          setOptions(mergedBranches);
          setLoading(false);
        }
      })
      .catch((catchError: unknown) => {
        if (cancelled) return;
        if (
          catchError instanceof DOMException &&
          catchError.name === "AbortError"
        )
          return;
        setError(
          catchError instanceof Error && catchError.message
            ? catchError.message
            : language === "th"
              ? "โหลดสาขาไม่สำเร็จ"
              : "Failed to load branches",
        );
        setLoading(false);
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [auth, language, workspace, form.businesscodes, form.companyguids]);

  function commitSelection(next: BranchOption[]) {
    setForm((current) => ({ ...current, [field.key]: next }));
  }

  function removeBranch(option: BranchOption) {
    const key = branchKeyOf(option);
    commitSelection(selected.filter((item) => branchKeyOf(item) !== key));
  }

  const summary =
    language === "th"
      ? `เลือก ${selected.length} / ${options.length} สาขา`
      : `${selected.length} / ${options.length} branches selected`;
  const pickLabel =
    language === "th" ? "เลือกสาขา" : "Pick branches";

  return (
    <section className="grid w-full gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>
          {label}
          {field.required ? " *" : ""}
        </span>
        <div className="flex flex-wrap items-center gap-2 text-xs font-normal text-muted-foreground">
          <span>{summary}</span>
          <Button
            type="button"
            size="sm"
            onClick={() => setDialogOpen(true)}
            disabled={loading || (options.length === 0 && !error)}
          >
            {loading ? (
              <Loader2 className="animate-spin" />
            ) : (
              <Search />
            )}
            {pickLabel}
          </Button>
        </div>
      </div>
      {loading ? (
        <div className="flex items-center gap-2 px-1 text-xs text-muted-foreground">
          <Loader2 className="size-3 animate-spin" />
          {language === "th" ? "กำลังโหลดสาขา…" : "Loading branches…"}
        </div>
      ) : null}
      {error ? (
        <p className="text-xs font-semibold text-destructive">{error}</p>
      ) : null}
      {options.length === 0 && !loading && !error ? (
        <p className="text-xs font-normal text-muted-foreground">
          {language === "th"
            ? "ยังไม่มีสาขาให้เลือก — เพิ่มสาขาในหน้า \"สาขา\" ก่อน"
            : 'No branches to choose yet — add one on the "Branch" screen first.'}
        </p>
      ) : null}
      {selected.length > 0 ? (
        <ul className="flex flex-wrap gap-1.5">
          {selected.map((option) => {
            const key = branchKeyOf(option);
            return (
              <li key={key}>
                <span className="inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-2 py-1 text-xs font-medium">
                  <span className="max-w-[200px] truncate">
                    {branchOptionDisplayName(option, language)}
                  </span>
                  {option.code ? (
                    <span className="text-[10px] font-normal text-muted-foreground">
                      {option.code}
                    </span>
                  ) : null}
                  <button
                    type="button"
                    className="ml-1 grid size-4 place-items-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
                    onClick={() => removeBranch(option)}
                    aria-label={
                      language === "th" ? "ลบสาขา" : "Remove branch"
                    }
                  >
                    <X className="size-3" />
                  </button>
                </span>
              </li>
            );
          })}
        </ul>
      ) : !loading && !error && options.length > 0 ? (
        <p className="text-xs font-medium text-emerald-600 dark:text-emerald-400">
          {language === "th"
            ? `ไม่ได้เลือก = ใช้ได้ทุกบริษัท (กด "${pickLabel}" เพื่อจำกัดเฉพาะที่เลือก)`
            : `None selected = all companies (click "${pickLabel}" to limit).`}
        </p>
      ) : null}
      {dialogOpen ? (
        <BranchPickerDialog
          initialSelected={selected}
          language={language}
          onCancel={() => setDialogOpen(false)}
          onConfirm={(next) => {
            commitSelection(next);
            setDialogOpen(false);
          }}
          options={options}
        />
      ) : null}
    </section>
  );
}

// ---------------------------------------------------------------------------
// BranchPickerDialog
// ---------------------------------------------------------------------------

export function BranchPickerDialog({
  initialSelected,
  language,
  onCancel,
  onConfirm,
  options,
}: {
  initialSelected: BranchOption[];
  language: LanguageCode;
  onCancel: () => void;
  onConfirm: (selected: BranchOption[]) => void;
  options: BranchOption[];
}) {
  const [draft, setDraft] = useState<BranchOption[]>(initialSelected);
  const [query, setQuery] = useState("");
  const draftKeys = useMemo(
    () => new Set(draft.map((item) => branchKeyOf(item))),
    [draft],
  );
  const filteredOptions = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return options;
    return options.filter((option) => {
      const name = branchOptionDisplayName(option, language).toLowerCase();
      const code = option.code.toLowerCase();
      return name.includes(needle) || code.includes(needle);
    });
  }, [language, options, query]);

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") onCancel();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onCancel]);

  function toggleDraft(option: BranchOption, checked: boolean) {
    const key = branchKeyOf(option);
    const without = draft.filter((item) => branchKeyOf(item) !== key);
    setDraft(checked ? [...without, option] : without);
  }

  function selectAllVisible() {
    const map = new Map<string, BranchOption>();
    for (const item of draft) map.set(branchKeyOf(item), item);
    for (const item of filteredOptions) map.set(branchKeyOf(item), item);
    setDraft(Array.from(map.values()));
  }

  function clearVisible() {
    if (!query.trim()) {
      setDraft([]);
      return;
    }
    const visibleKeys = new Set(
      filteredOptions.map((item) => branchKeyOf(item)),
    );
    setDraft(draft.filter((item) => !visibleKeys.has(branchKeyOf(item))));
  }

  const title = language === "th" ? "เลือกสาขา" : "Pick branches";
  const searchPlaceholder =
    language === "th"
      ? "ค้นหารหัสหรือชื่อสาขา"
      : "Search branch code or name";
  const summary =
    language === "th"
      ? `เลือก ${draft.length} / ${options.length} สาขา (กรอง ${filteredOptions.length})`
      : `${draft.length} / ${options.length} selected (${filteredOptions.length} filtered)`;
  const selectAllLabel =
    language === "th"
      ? query.trim()
        ? "เลือกทั้งหมดที่กรอง"
        : "เลือกทุกสาขา"
      : query.trim()
        ? "Select all filtered"
        : "Select all";
  const clearLabel =
    language === "th"
      ? query.trim()
        ? "ล้างที่กรอง"
        : "ล้างทั้งหมด"
      : query.trim()
        ? "Clear filtered"
        : "Clear all";

  return (
    <div
      className="fixed inset-0 z-50 flex flex-col bg-card text-card-foreground"
      role="dialog"
      aria-modal="true"
      aria-label={title}
    >
      <header className="flex items-center justify-between gap-2 border-b border-border px-3 py-2">
        <div className="flex items-center gap-2 text-sm font-semibold">
          <GitBranch className="size-4" />
          {title}
        </div>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onCancel}
          aria-label={language === "th" ? "ยกเลิก" : "Cancel"}
        >
          <X />
        </Button>
      </header>
      <div className="flex flex-wrap items-center gap-2 border-b border-border px-3 py-2">
        <label className="relative flex min-w-0 flex-1 items-center">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            autoFocus
            className="h-9 !pl-10"
            placeholder={searchPlaceholder}
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
        </label>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={selectAllVisible}
          disabled={filteredOptions.length === 0}
        >
          {selectAllLabel}
        </Button>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={clearVisible}
          disabled={draft.length === 0}
        >
          {clearLabel}
        </Button>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto px-3 py-2">
        {filteredOptions.length === 0 ? (
          <p className="grid h-full place-items-center text-sm text-muted-foreground">
            {language === "th" ? "ไม่พบสาขา" : "No branches found"}
          </p>
        ) : (
          <ul className="grid gap-1 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
            {filteredOptions.map((option) => {
              const key = branchKeyOf(option);
              const checked = draftKeys.has(key);
              return (
                <li key={key}>
                  <label
                    className={cn(
                      "flex cursor-pointer items-center gap-2 rounded-md border px-2 py-1.5 text-xs font-medium transition-colors",
                      checked
                        ? "border-primary/40 bg-primary/5 hover:bg-primary/10"
                        : "border-border bg-card hover:border-primary/40 hover:bg-accent/40",
                    )}
                  >
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={(event) =>
                        toggleDraft(option, event.target.checked)
                      }
                      className="size-4 accent-primary"
                    />
                    <span className="min-w-0 flex-1 truncate">
                      {branchOptionDisplayName(option, language)}
                    </span>
                    {option.code ? (
                      <span className="shrink-0 text-[10px] text-muted-foreground">
                        {option.code}
                      </span>
                    ) : null}
                  </label>
                </li>
              );
            })}
          </ul>
        )}
      </div>
      <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border px-3 py-2">
        <span className="text-xs text-muted-foreground">{summary}</span>
        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline" onClick={onCancel}>
            {language === "th" ? "ยกเลิก" : "Cancel"}
          </Button>
          <Button type="button" onClick={() => onConfirm(draft)}>
            <BadgeCheck />
            {language === "th" ? "ยืนยัน" : "Confirm"}
          </Button>
        </div>
      </footer>
    </div>
  );
}

// ---------------------------------------------------------------------------
// BranchMultiSelectReadOnlyDetail
// ---------------------------------------------------------------------------

export function BranchMultiSelectReadOnlyDetail({
  label,
  language,
  value,
}: {
  label: string;
  language: LanguageCode;
  value: unknown;
}) {
  const selected = selectedBranchesFromValue(value);
  return (
    <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)] border-l-2 border-l-secondary">
      <span className="text-xs font-semibold text-muted-foreground">
        {label}
      </span>
      {selected.length === 0 ? (
        <b className="text-foreground font-medium">-</b>
      ) : (
        <ul className="flex flex-wrap gap-1">
          {selected.map((option) => {
            const key = branchKeyOf(option);
            return (
              <li
                key={key}
                className="rounded-md border border-border bg-background px-2 py-0.5 text-xs font-semibold"
              >
                {branchOptionDisplayName(option, language)}
                {option.code ? (
                  <span className="ml-1 text-[10px] font-normal text-muted-foreground">
                    {option.code}
                  </span>
                ) : null}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// CompanyMultiSelectFieldEditor
// ---------------------------------------------------------------------------

export function CompanyMultiSelectFieldEditor({
  auth,
  field,
  form,
  language,
  setForm,
  workspace,
}: {
  auth: AuthSession | null;
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  setForm: (update: FormState | ((current: FormState) => FormState)) => void;
  workspace: WorkspaceSession | null;
}) {
  const [shops, setShops] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const selectedHoldingCodes = useMemo(() => {
    const val = form[field.key] ?? form.companyguids;
    if (Array.isArray(val)) {
      return val.map((item) => {
        if (typeof item === "string") return item.trim();
        if (isRecord(item)) return stringValue(item.guidfixed ?? item.guidfixed ?? item.guid ?? "");
        return "";
      }).filter(Boolean);
    }
    return [];
  }, [field.key, form]);

  useEffect(() => {
    if (!auth || !workspace) return;
    const controller = new AbortController();
    let cancelled = false;
    setLoading(true);
    setError("");

    const params = new URLSearchParams();
    const activeHoldingCode = stringValue(workspace.shop.holdingcode);
    if (activeHoldingCode) params.set("activeholdingcode", activeHoldingCode);

    void authFetch(`/api/workspace/holdings?management=true${params.size > 0 ? `&${params.toString()}` : ""}`, {
      headers: requestHeaders(auth),
      cache: "no-store",
      signal: controller.signal,
    })
      .then(async (res) => {
        if (!res.ok) throw new Error("Load shops failed");
        return res.json() as Promise<any>;
      })
      .then(async (shopsPayload) => {
        if (cancelled) return;
        const shopsList = Array.isArray(shopsPayload.data) ? shopsPayload.data : [];
        setShops(shopsList);
        setLoading(false);
      })
      .catch((catchError: unknown) => {
        if (cancelled) return;
        setError(
          catchError instanceof Error && catchError.message
            ? catchError.message
            : language === "th"
              ? "โหลดข้อมูลบริษัทไม่สำเร็จ"
              : "Failed to load companies",
        );
        setLoading(false);
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [auth, language, workspace]);

  function handleToggleShop(holdingCode: string, checked: boolean) {
    let nextShops = [...selectedHoldingCodes];
    if (checked) {
      if (!nextShops.includes(holdingCode)) nextShops.push(holdingCode);
    } else {
      nextShops = nextShops.filter((id) => id !== holdingCode);
    }

    setForm({
      ...form,
      [field.key]: nextShops,
    });
  }

  return (
    <section className="grid w-full gap-3 rounded-2xl border border-border bg-background p-4 text-sm font-semibold shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border pb-2">
        <span>{language === "th" ? "สิทธิ์การเข้าถึงบริษัท" : "Company Access"}</span>
        {loading ? (
          <span className="flex items-center gap-1 text-xs font-normal text-muted-foreground animate-pulse">
            <Loader2 className="size-3 animate-spin" />
            {language === "th" ? "กำลังโหลดข้อมูลบริษัท…" : "Loading companies…"}
          </span>
        ) : null}
      </div>
      {error ? <p className="text-xs font-semibold text-destructive">{error}</p> : null}

      {!loading && !error && shops.length === 0 ? (
        <p className="text-xs font-normal text-muted-foreground py-2">
          {language === "th" ? "ไม่พบข้อมูลบริษัทในระบบ" : "No companies found."}
        </p>
      ) : null}

      {!loading && !error && shops.length > 0 ? (
        <div className="flex flex-col gap-2.5 py-1">
          {shops.map((shop) => {
            const sid = stringValue(shop.businesscode ?? shop.code ?? shop.holdingcode);
            const shopName = shop.names?.find((n: any) => n.code === language)?.name || shop.name1 || shop.name || sid;
            const isShopChecked = selectedHoldingCodes.includes(sid);

            return (
              <label
                key={sid}
                className="flex min-w-0 cursor-pointer items-center gap-2.5 rounded-xl border border-border/60 bg-card p-3 font-bold transition-colors hover:bg-muted/40"
              >
                <input
                  type="checkbox"
                  checked={isShopChecked}
                  onChange={(event) => handleToggleShop(sid, event.target.checked)}
                  className="size-4 cursor-pointer rounded border-border text-primary"
                />
                <Building2 className="size-4 shrink-0 text-primary" />
                <span className="min-w-0 truncate">{shopName}</span>
              </label>
            );
          })}
        </div>
      ) : null}
    </section>
  );
}

// ---------------------------------------------------------------------------
// CompanyMultiSelectReadOnlyDetail
// ---------------------------------------------------------------------------

export function CompanyMultiSelectReadOnlyDetail({
  label,
  language,
  value,
  auth,
}: {
  label: string;
  language: LanguageCode;
  value: unknown;
  auth: AuthSession | null;
}) {
  const [options, setOptions] = useState<MasterEntry[]>([]);
  const selectedGuids = useMemo(() => {
    if (Array.isArray(value)) {
      return value.map((item) => {
        if (typeof item === "string") return item.trim();
        if (isRecord(item)) return stringValue(item.guidfixed ?? item.guidfixed ?? item.guid ?? "");
        return "";
      }).filter(Boolean);
    }
    return [];
  }, [value]);

  useEffect(() => {
    if (!auth || selectedGuids.length === 0) return;
    const controller = new AbortController();
    void authFetch(`/api/workspace/holdings?management=true`, {
      headers: requestHeaders(auth),
      cache: "no-store",
      signal: controller.signal,
    })
      .then(async (response) => {
        if (response.ok) return response.json();
      })
      .then((payload) => {
        if (payload && payload.success && Array.isArray(payload.data)) {
          const parsed = payload.data.map((shop: any) => ({
            guidfixed: shop.holdingcode,
            code: "",
            names: shop.names || [{ code: "th", name: shop.name1 || shop.name || "" }]
          }));
          setOptions(parsed);
        }
      })
      .catch(() => {});
    return () => controller.abort();
  }, [auth, selectedGuids.length]);

  const selectedOptions = useMemo(() => {
    return selectedGuids.map((guid) => {
      const match = options.find((opt) => opt.guidfixed === guid);
      if (match) return match;
      return { guidfixed: guid, code: "", names: [] } as MasterEntry;
    });
  }, [selectedGuids, options]);

  return (
    <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)] border-l-2 border-l-secondary">
      <span className="text-xs font-semibold text-muted-foreground">
        {label}
      </span>
      {selectedGuids.length === 0 ? (
        <b className="font-medium text-emerald-600 dark:text-emerald-400">
          {language === "th" ? "ใช้ได้ทุกบริษัท" : "All companies"}
        </b>
      ) : (
        <ul className="flex flex-wrap gap-1">
          {selectedOptions.map((option) => {
            return (
              <li
                key={option.guidfixed}
                className="rounded-md border border-border bg-background px-2 py-0.5 text-xs font-semibold"
              >
                {companyOptionDisplayName(option, language)}
                {option.code ? (
                  <span className="ml-1 text-[10px] font-normal text-muted-foreground">
                    ({option.code})
                  </span>
                ) : null}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// BranchCoordinatePairEditor
// ---------------------------------------------------------------------------

export function BranchCoordinatePairEditor({
  config,
  form,
  language,
  setForm,
}: {
  config: SystemSettingConfig;
  form: FormState;
  language: LanguageCode;
  setForm: (update: FormState | ((current: FormState) => FormState)) => void;
}) {
  const [mapOpen, setMapOpen] = useState(false);
  const latitudeField = config.fields.find(
    (item) => item.key === "contact.latitude",
  );
  const longitudeField = config.fields.find(
    (item) => item.key === "contact.longitude",
  );
  const latitudeLabel = latitudeField
    ? (latitudeField.label[language] ?? latitudeField.label.en ?? "Latitude")
    : "Latitude";
  const longitudeLabel = longitudeField
    ? (longitudeField.label[language] ?? longitudeField.label.en ?? "Longitude")
    : "Longitude";
  const latitudeRaw = stringValue(form["contact.latitude"]);
  const longitudeRaw = stringValue(form["contact.longitude"]);
  const pickLabel =
    language === "th" ? "เลือกจากแผนที่" : "Pick on map";

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <div className="grid gap-2 md:grid-cols-[1fr_1fr_auto] md:items-end">
        <label className="grid gap-1 text-sm font-semibold">
          <span>{latitudeLabel}</span>
          <Input
            type="number"
            inputMode="decimal"
            step="any"
            value={latitudeRaw}
            placeholder="13.756300"
            onChange={(event) =>
              setForm({ ...form, "contact.latitude": event.target.value })
            }
          />
        </label>
        <label className="grid gap-1 text-sm font-semibold">
          <span>{longitudeLabel}</span>
          <Input
            type="number"
            inputMode="decimal"
            step="any"
            value={longitudeRaw}
            placeholder="100.501800"
            onChange={(event) =>
              setForm({ ...form, "contact.longitude": event.target.value })
            }
          />
        </label>
        <Button
          type="button"
          variant="outline"
          onClick={() => setMapOpen(true)}
        >
          <MapPin />
          {pickLabel}
        </Button>
      </div>
      <MapPickerDialog
        open={mapOpen}
        initialLat={parseCoordinateValue(latitudeRaw)}
        initialLng={parseCoordinateValue(longitudeRaw)}
        language={language}
        onCancel={() => setMapOpen(false)}
        onSelect={(lat, lng) => {
          setForm({
            ...form,
            "contact.latitude": lat.toFixed(6),
            "contact.longitude": lng.toFixed(6),
          });
          setMapOpen(false);
        }}
      />
    </section>
  );
}
