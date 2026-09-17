"use client";

import {
  CircleDollarSign,
  Edit3,
  Loader2,
  Plus,
  RefreshCcw,
  Save,
  Search,
  Trash2,
  X,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { backendText, useBackendLanguage, type BackendLanguageDictionary } from "@/lib/backend-language";
import { deriveMainApiUrl } from "@/lib/backend-url";
import { applyCurrencySymbolPreset, currencyPresetSource, filterCurrencySymbolPresets, findCurrencySymbolPreset } from "@/lib/currency-presets";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { pushNotice } from "@/lib/toast";
import { authFetch, getAuthSession } from "@/lib/client-auth-session";
import {
  type AuthSession,
  type WorkspaceSession,
  workspaceStorageKeys,
} from "@/lib/workspace-models";
import { AppHeaderControls } from "../app-header-controls";
import { ManualLink } from "../manual-link";

type CurrencyScreenProps = {
  embedded?: boolean;
  initialBackendLanguage?: BackendLanguageDictionary;
  initialBackendUrl?: string;
  initialLanguage?: LanguageCode;
  language?: LanguageCode;
};

type CurrencyRecord = {
  guidfixed: string;
  duplicate_guidfixeds?: string[];
  code: string;
  name: string;
  symbol: string;
  isdisabled: boolean;
  exchangerates?: ExchangeRateEntry[];
};

type ExchangeRateEntry = {
  guidfixed?: string;
  date?: string;
  rate?: number;
};

type CurrencyForm = {
  code: string;
  name: string;
  symbol: string;
  isdisabled: boolean;
  rate: number;
};

type ApiResponse<T> = {
  success?: boolean;
  message?: string;
  data?: T;
  id?: string;
  total?: number;
  pagination?: {
    total?: number;
    totalitem?: number;
    totalItem?: number;
  };
};

const currencyTextEn = {
  title: "Currency",
  subtitle: "Manage document currencies, symbols, and active status for the selected company.",
  refresh: "Refresh",
  total: "Total",
  active: "Active",
  disabled: "Disabled",
  baseCurrency: "Base currency",
  search: "Search currency",
  add: "Add currency",
  edit: "Edit currency",
  delete: "Delete",
  save: "Save",
  cancel: "Cancel",
  saving: "Saving",
  loading: "Loading currency data",
  noData: "No currency data",
  emptyHint: "Add the first currency or change the search term.",
  code: "Currency code",
  name: "Currency name",
  symbol: "Symbol",
  status: "Status",
  enabled: "Enabled",
  baseBadge: "Base",
  formNew: "New currency",
  formEdit: "Edit currency",
  editLocked: "The code is locked in edit mode to match the legacy workflow.",
  requiredCode: "Please enter currency code.",
  requiredName: "Please enter currency name.",
  codeLength: "Currency code must be 2-5 characters.",
  selectSymbol: "Please select currency symbol.",
  symbolSearch: "Search code, name, or symbol",
  currencyVerified: "Verified",
  currencyNotVerified: "Not verified",
  currencyVerifySource: "Code/name: ISO 4217 / SIX. Symbol/abbreviation: Wikipedia.",
  currencyVerifyRequired: "Select a currency from the verified list.",
  duplicateCode: "This currency code already exists.",
  noSymbolResult: "No matching currency",
  loadFailed: "Could not load currencies.",
  requestFailed: "Request failed.",
  deleteConfirm: "Confirm delete this currency?",
  addSuccess: "Currency added.",
  editSuccess: "Currency updated.",
  deleteSuccess: "Currency deleted.",
  apiRequired: "Please login and select a company before opening this screen.",
  tenant: "Company",
  branch: "Branch",
  exchangeRates: "Exchange rates",
  none: "None",
  setBase: "Set as base currency",
  baseSetSuccess: "Base currency of this branch updated.",
  baseSetFailed: "Could not set base currency.",
  lastActiveGuard: "At least one active currency is required — add or enable another currency first.",
  baseGuardDelete: "This currency is the base of this branch — set another base currency first.",
  autoSeedNotice: "No currency found — THB was added and set as the base currency automatically.",
  rateVsBase: "Exchange rate (vs base)",
  quickAdd: "Quick add (ISO 4217)",
  baseNoneHint: "No base currency set for this branch yet.",
} as const;

type CurrencyTextKey = keyof typeof currencyTextEn;

const currencyBackendKeys: Record<CurrencyTextKey, string> = {
  add: "add_currency",
  addSuccess: "add_currency_success",
  active: "active",
  apiRequired: "currency_api_required",
  autoSeedNotice: "currency_auto_seed_notice",
  baseBadge: "currency_base_badge",
  baseCurrency: "basecurrency",
  baseGuardDelete: "currency_base_guard_delete",
  baseNoneHint: "currency_base_none_hint",
  baseSetFailed: "currency_base_set_failed",
  baseSetSuccess: "currency_base_set_success",
  branch: "branch",
  cancel: "cancel",
  code: "currency_code",
  codeLength: "currency_code_length_error",
  currencyNotVerified: "currency_not_verified",
  currencyVerified: "currency_verified",
  currencyVerifyRequired: "currency_verify_required",
  currencyVerifySource: "currency_verify_source",
  delete: "delete",
  deleteConfirm: "confirm_delete_currency",
  deleteSuccess: "delete_currency_success",
  disabled: "disabled",
  duplicateCode: "duplicate_currency_code",
  edit: "edit_currency",
  editLocked: "currency_edit_locked",
  editSuccess: "edit_currency_success",
  emptyHint: "currency_empty_hint",
  enabled: "enabled",
  exchangeRates: "currency_exchange_rates",
  formEdit: "currency_form_edit",
  formNew: "currency_form_new",
  lastActiveGuard: "currency_last_active_guard",
  loadFailed: "currency_load_failed",
  loading: "currency_loading",
  name: "currency_name",
  noData: "no_currency_data",
  none: "none",
  noSymbolResult: "currency_no_symbol_result",
  quickAdd: "currency_quick_add",
  rateVsBase: "currency_rate_vs_base",
  refresh: "refresh",
  requestFailed: "request_failed",
  requiredCode: "please_enter_currency_code",
  requiredName: "please_enter_currency_name",
  save: "save",
  saving: "saving",
  search: "search_currency",
  selectSymbol: "please_select_currency_symbol",
  setBase: "currency_set_base",
  status: "status",
  subtitle: "currency_subtitle",
  symbol: "symbol",
  symbolSearch: "search_currency_symbol",
  tenant: "company",
  title: "currency",
  total: "total",
};

const emptyForm: CurrencyForm = {
  code: "",
  name: "",
  symbol: "",
  isdisabled: false,
  rate: 1,
};

/** เพิ่มด่วนจากมาตรฐาน ISO 4217 — สกุลที่ธุรกิจไทยใช้บ่อยที่สุด */
const QUICK_ADD_CODES = ["THB", "USD", "EUR", "JPY", "CNY", "GBP", "MYR", "SGD"] as const;

export function CurrencyScreen({ embedded = false, initialBackendLanguage, initialBackendUrl, initialLanguage = "th", language: externalLanguage }: CurrencyScreenProps) {
  const router = useRouter();
  const [language, setLanguage] = useState<LanguageCode>(externalLanguage ?? initialLanguage);
  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [currencies, setCurrencies] = useState<CurrencyRecord[]>([]);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const setNotice = pushNotice;
  const [editing, setEditing] = useState<CurrencyRecord | null>(null);
  const [form, setForm] = useState<CurrencyForm>(emptyForm);
  const [formOpen, setFormOpen] = useState(false);
  const activeBackendUrl = auth?.backendUrl ?? initialBackendUrl;
  const backendLanguage = useBackendLanguage(language, activeBackendUrl, language === initialLanguage ? initialBackendLanguage : undefined);
  const { confirm, confirmationDialog } = useConfirmDialog({
    defaultConfirmLabel: backendText(backendLanguage, "common_confirm", "ยืนยัน"),
    defaultCancelLabel: backendText(backendLanguage, "common_cancel", "ยกเลิก"),
  });

  const text = useCallback(
    (key: CurrencyTextKey) => {
      const fallback = currencyTextEn[key] ?? key;
      return backendText(backendLanguage, currencyBackendKeys[key], fallback);
    },
    [backendLanguage],
  );

  const loadCurrencies = useCallback(async (currentAuth: AuthSession | null) => {
    if (!currentAuth) return;
    setLoading(true);
    setNotice(null);
    try {
      const response = await authFetch("/api/currency?page=1&limit=1000", {
        headers: {
          "x-bc-backend-url": currentAuth.backendUrl,
          Authorization: `Bearer ${currentAuth.token}`,
        },
        cache: "no-store",
      });
      const payload = await response.json() as ApiResponse<unknown>;
      if (!response.ok || payload.success === false) {
        throw new Error(payload.message || currencyTextEn.loadFailed);
      }
      setCurrencies(normalizeCurrencies(payload.data));
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : currencyTextEn.loadFailed });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (externalLanguage) setLanguage(externalLanguage);
  }, [externalLanguage]);

  useEffect(() => {
    const savedLanguage = normalizeLanguage(localStorage.getItem("user_language") ?? externalLanguage ?? initialLanguage);
    if (!externalLanguage) setLanguage(savedLanguage);

    const nextAuth = readAuth();
    const nextWorkspace = readWorkspace();
    if (!nextAuth || !nextWorkspace) {
      setNotice({ type: "info", text: text("apiRequired") });
      if (!embedded) router.replace("/");
      return;
    }

    setAuth(nextAuth);
    setWorkspace(nextWorkspace);
    void loadCurrencies(nextAuth);
  }, [embedded, externalLanguage, initialLanguage, loadCurrencies, router, text]);

  useEffect(() => {
    document.documentElement.lang = language;
    if (!externalLanguage) localStorage.setItem("user_language", language);
  }, [externalLanguage, language]);

  const [baseCurrencyOverride, setBaseCurrencyOverride] = useState("");
  const baseCurrency = (baseCurrencyOverride || (workspace?.branch as { basecurrency?: string } | undefined)?.basecurrency || "").trim().toUpperCase();
  const activeCount = currencies.filter((item) => !item.isdisabled).length;
  const disabledCount = currencies.length - activeCount;
  const visibleCurrencies = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return currencies;
    return currencies.filter((currency) =>
      `${currency.code} ${currency.name} ${currency.symbol}`.toLowerCase().includes(needle),
    );
  }, [currencies, query]);

  function openCreate() {
    setEditing(null);
    setForm(emptyForm);
    setFormOpen(true);
    setNotice(null);
  }

  function openEdit(currency: CurrencyRecord) {
    setEditing(currency);
    setForm({
      code: currency.code,
      name: currency.name,
      symbol: currency.symbol || "฿",
      isdisabled: currency.isdisabled,
      rate: latestRateVsBase(currency, baseCurrency) ?? 1,
    });
    setFormOpen(true);
    setNotice(null);
  }

  function isLastActive(currency: CurrencyRecord): boolean {
    return activeCount === 1 && !currency.isdisabled;
  }

  function isBaseCurrency(currency: CurrencyRecord): boolean {
    return Boolean(baseCurrency) && currency.code.toUpperCase() === baseCurrency;
  }

  /** ตั้งสกุลหลักของสาขาปัจจุบันจากหน้านี้เลย (เดิมซ่อนอยู่ในฟอร์มแก้ไขสาขา) */
  async function setBaseCurrency(code: string) {
    if (!auth) return;
    const branchGuid = workspace?.branch?.guidfixed ?? "";
    if (!branchGuid) {
      setNotice({ type: "error", text: text("baseSetFailed") });
      return;
    }
    setSaving(true);
    try {
      const mainApiUrl = deriveMainApiUrl(auth.backendUrl);
      const headers = { "Content-Type": "application/json", Authorization: `Bearer ${auth.token}` };
      const listRes = await authFetch(`${mainApiUrl}/organization/branch?management=true&_=${Date.now()}`, { headers, cache: "no-store" });
      const listPayload = await listRes.json() as ApiResponse<unknown>;
      if (!listRes.ok || listPayload.success === false) throw new Error(String(listPayload.message || text("requestFailed")));
      const branchDoc = (Array.isArray(listPayload.data) ? listPayload.data : []).find(
        (item) => (item as { guidfixed?: string })?.guidfixed === branchGuid,
      ) as Record<string, unknown> | undefined;
      if (!branchDoc) throw new Error(text("requestFailed"));
      const putRes = await authFetch(`${mainApiUrl}/organization/branch/${encodeURIComponent(branchGuid)}`, {
        method: "PUT",
        headers,
        body: JSON.stringify({ ...branchDoc, basecurrency: code }),
      });
      const putPayload = await putRes.json() as ApiResponse<unknown>;
      if (!putRes.ok || putPayload.success === false) throw new Error(String(putPayload.message || text("requestFailed")));
      setBaseCurrencyOverride(code);
      setNotice({ type: "success", text: `${text("baseSetSuccess")} (${code})` });
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text("baseSetFailed") });
    } finally {
      setSaving(false);
    }
  }

  /** ระบบต้องมีสกุลเงินอย่างน้อย 1 สกุลเสมอ — ถ้าว่างเปล่าให้ seed THB + ตั้งเป็นสกุลหลักอัตโนมัติ */
  const seedAttemptedRef = useRef(false);
  useEffect(() => {
    if (seedAttemptedRef.current || loading || !auth || currencies.length > 0) return;
    seedAttemptedRef.current = true;
    void (async () => {
      try {
        const response = await authFetch("/api/currency", {
          method: "POST",
          headers: { "Content-Type": "application/json", "x-bc-backend-url": auth.backendUrl, Authorization: `Bearer ${auth.token}` },
          body: JSON.stringify({ backendUrl: auth.backendUrl, guidfixed: "", code: "THB", name: "Thai Baht", symbol: "฿", isdisabled: false }),
        });
        const payload = await response.json() as ApiResponse<unknown>;
        if (!response.ok || payload.success === false) throw new Error(String(payload.message || ""));
        if (!baseCurrency) await setBaseCurrency("THB");
        setNotice({ type: "success", text: text("autoSeedNotice") });
        await loadCurrencies(auth);
      } catch {
        seedAttemptedRef.current = false;
      }
    })();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [auth, currencies.length, loading]);

  /** เพิ่มด่วนจาก preset ISO 4217 คลิกเดียวจบ */
  async function quickAdd(code: string) {
    if (!auth) return;
    if (currencies.some((item) => item.code === code)) {
      setNotice({ type: "error", text: text("duplicateCode") });
      return;
    }
    // หา preset ด้วยรหัส — ต้องได้ name+symbol มาตรฐาน ไม่ใช่ fallback ตัวรหัสเอง
    // (ชื่อ/สัญลักษณ์ไม่ตรง preset จะทำให้ validateForm ปฏิเสธตอนแก้ไขภายหลัง)
    const preset = filterCurrencySymbolPresets(code).find((item) => item.code === code);
    const name = preset?.name || preset?.isoName || code;
    const symbol = preset?.symbol || code;
    setSaving(true);
    try {
      const response = await authFetch("/api/currency", {
        method: "POST",
        headers: { "Content-Type": "application/json", "x-bc-backend-url": auth.backendUrl, Authorization: `Bearer ${auth.token}` },
        body: JSON.stringify({ backendUrl: auth.backendUrl, guidfixed: "", code, name, symbol, isdisabled: false }),
      });
      const payload = await response.json() as ApiResponse<unknown>;
      if (!response.ok || payload.success === false) throw new Error(String(payload.message || text("requestFailed")));
      if (!baseCurrency) await setBaseCurrency(code);
      setNotice({ type: "success", text: text("addSuccess") });
      await loadCurrencies(auth);
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text("requestFailed") });
    } finally {
      setSaving(false);
    }
  }

  async function saveCurrency(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!auth) return;

    const payload = normalizeForm(form);
    const validation = validateForm(payload, text);
    if (validation) {
      setNotice({ type: "error", text: validation });
      return;
    }
    const duplicate = currencies.some((currency) =>
      currency.code === payload.code &&
      (!editing || currency.guidfixed !== editing.guidfixed),
    );
    if (duplicate) {
      setNotice({ type: "error", text: text("duplicateCode") });
      return;
    }

    // Guard (กฎ 2026-08-30): ห้ามปิดใช้งาน active ตัวสุดท้าย / ห้ามปิดสกุลที่เป็นสกุลหลัก
    if (editing && !editing.isdisabled && payload.isdisabled) {
      if (isBaseCurrency(editing)) {
        setNotice({ type: "error", text: text("baseGuardDelete") });
        return;
      }
      if (isLastActive(editing)) {
        setNotice({ type: "error", text: text("lastActiveGuard") });
        return;
      }
    }

    setSaving(true);
    setNotice(null);
    try {
      const path = editing ? `/api/currency/${encodeURIComponent(editing.guidfixed)}` : "/api/currency";
      const response = await authFetch(path, {
        method: editing ? "PUT" : "POST",
        headers: {
          "Content-Type": "application/json",
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify({
          backendUrl: auth.backendUrl,
          guidfixed: editing?.guidfixed ?? "",
          code: payload.code,
          name: payload.name,
          symbol: payload.symbol,
          isdisabled: payload.isdisabled,
          ...(editing && Number.isFinite(payload.rate) && payload.rate > 0 && payload.code !== baseCurrency
            ? { exchangerates: upsertRateEntry(editing.exchangerates, payload.rate) }
            : {}),
        }),
      });
      const data = await response.json() as ApiResponse<unknown>;
      if (!response.ok || data.success === false) {
        throw new Error(data.message || text("requestFailed"));
      }

      setFormOpen(false);
      setEditing(null);
      setNotice({ type: "success", text: editing ? text("editSuccess") : text("addSuccess") });
      await loadCurrencies(auth);
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text("requestFailed") });
    } finally {
      setSaving(false);
    }
  }

  async function deleteCurrency(currency: CurrencyRecord) {
    if (!auth) return;
    // Guard (กฎ 2026-08-30): ห้ามลบสกุลหลัก / ห้ามลบ active ตัวสุดท้าย
    if (isBaseCurrency(currency)) {
      setNotice({ type: "error", text: text("baseGuardDelete") });
      return;
    }
    if (isLastActive(currency)) {
      setNotice({ type: "error", text: text("lastActiveGuard") });
      return;
    }
    const deleteIds = uniqueStrings([currency.guidfixed, ...(currency.duplicate_guidfixeds ?? [])]);
    if (!deleteIds.length) {
      setNotice({ type: "error", text: text("requestFailed") });
      return;
    }
    const confirmed = await confirm({
      title: text("deleteConfirm"),
      description: `${currency.code} ${currency.name}`.trim(),
      details: currency.duplicate_guidfixeds && currency.duplicate_guidfixeds.length > 1
        ? `${backendText(backendLanguage, "cur_all_duplicate_entries_will_be", "จะลบรายการซ้ำทั้งหมด")}: ${currency.duplicate_guidfixeds.length}`
        : undefined,
      confirmLabel: text("delete"),
      cancelLabel: text("cancel"),
      tone: "danger",
    });
    if (!confirmed) return;

    setLoading(true);
    setNotice(null);
    try {
      const response = await authFetch(deleteIds.length === 1 ? `/api/currency/${encodeURIComponent(deleteIds[0])}` : "/api/currency", {
        method: "DELETE",
        headers: {
          ...(deleteIds.length > 1 ? { "Content-Type": "application/json" } : {}),
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: deleteIds.length > 1 ? JSON.stringify(deleteIds) : undefined,
      });
      const data = await response.json() as ApiResponse<unknown>;
      if (!response.ok || data.success === false) {
        throw new Error(data.message || text("requestFailed"));
      }
      setNotice({ type: "success", text: text("deleteSuccess") });
      await loadCurrencies(auth);
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text("requestFailed") });
    } finally {
      setLoading(false);
    }
  }

  const content = (
    <div className="grid w-full min-w-0 gap-3">
      <header className="rounded-2xl border border-border bg-card p-3 shadow-sm">
        <div className="flex w-full flex-wrap items-start justify-between gap-2">
          <div className="flex min-w-0 items-start gap-2">
            <span className="grid size-10 shrink-0 place-items-center rounded-2xl bg-primary/10 text-primary">
              <CircleDollarSign size={22} />
            </span>
            <div className="min-w-0">
              <p className="text-xs font-semibold uppercase text-muted-foreground">BC Ai Account</p>
              <h1 className="truncate text-xl font-semibold sm:text-2xl">{text("title")}</h1>
              <p className="max-w-[92ch] text-sm leading-6 text-muted-foreground">{text("subtitle")}</p>
            </div>
          </div>
          <div className="flex min-w-0 flex-wrap justify-end gap-2">
            {embedded ? null : <AppHeaderControls language={language} onLanguageChange={setLanguage} showSettings={false} />}
            <ManualLink compact language={language} screen="currency" />
          </div>
        </div>
        {baseCurrency ? (
          <div className="mt-2 flex flex-wrap gap-2 text-xs text-muted-foreground">
            <Badge variant="success">{text("baseCurrency")}: {baseCurrency}</Badge>
          </div>
        ) : null}
      </header>

      <section className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label={text("total")} value={formatCount(currencies.length, language)} />
        <StatCard label={text("active")} value={formatCount(activeCount, language)} tone="success" />
        <StatCard label={text("disabled")} value={formatCount(disabledCount, language)} tone="warning" />
        <StatCard label={text("baseCurrency")} value={baseCurrency || "-"} />
      </section>

      <Card>
        <CardContent className="grid gap-2 p-3">
          <div className="grid gap-2 lg:grid-cols-[minmax(0,1fr)_auto_auto]">
            <label className="relative block min-w-0">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input className="!pl-10" value={query} onChange={(event) => setQuery(event.target.value)} placeholder={text("search")} />
            </label>
            <Button type="button" variant="outline" onClick={() => void loadCurrencies(auth)} disabled={loading || !auth}>
              {loading ? <Loader2 className="animate-spin" /> : <RefreshCcw />}
              {text("refresh")}
            </Button>
            <Button type="button" onClick={openCreate} disabled={!auth}>
              <Plus />
              {text("add")}
            </Button>
          </div>
          {formOpen ? null : (
            <div className="flex min-w-0 flex-wrap items-center gap-1.5">
              <span className="text-xs font-semibold text-muted-foreground">{text("quickAdd")}:</span>
              {QUICK_ADD_CODES.map((code) => (
                <Button
                  key={code}
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-7 px-2 text-xs font-bold"
                  disabled={saving || !auth || currencies.some((item) => item.code === code)}
                  onClick={() => void quickAdd(code)}
                >
                  {code}
                </Button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {formOpen ? (
        <CurrencyFormPanel
          baseCurrency={baseCurrency}
          editing={editing}
          form={form}
          onClose={() => {
            if (!saving) setFormOpen(false);
          }}
          onFormChange={setForm}
          onSubmit={saveCurrency}
          saving={saving}
          text={text}
        />
      ) : null}

      {loading && currencies.length === 0 ? (
        <Card>
          <CardContent className="flex min-h-40 items-center justify-center gap-2 p-4 text-sm text-muted-foreground">
            <Loader2 className="animate-spin" />
            {text("loading")}
          </CardContent>
        </Card>
      ) : visibleCurrencies.length ? (
        <section className="grid min-w-0 gap-2 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4" aria-label={text("title")}>
          {visibleCurrencies.map((currency) => (
            <CurrencyCard
              baseCurrency={baseCurrency}
              currency={currency}
              isBaseFlag={Boolean(baseCurrency) && currency.code.toUpperCase() === baseCurrency}
              isLastActiveFlag={activeCount === 1 && !currency.isdisabled}
              key={currency.guidfixed || currency.code}
              onDelete={deleteCurrency}
              onEdit={openEdit}
              onSetBase={(code) => void setBaseCurrency(code)}
              text={text}
            />
          ))}
        </section>
      ) : (
        <Card>
          <CardContent className="grid min-h-44 place-items-center p-4 text-center">
            <div className="grid gap-2">
              <CircleDollarSign className="mx-auto size-10 text-muted-foreground" />
              <h2 className="text-base font-semibold">{text("noData")}</h2>
              <p className="text-sm text-muted-foreground">{text("emptyHint")}</p>
              <Button type="button" onClick={openCreate} disabled={!auth}>
                <Plus />
                {text("add")}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {confirmationDialog}
    </div>
  );

  if (embedded) return <section className="grid w-full min-w-0 gap-3">{content}</section>;
  return <main className="min-h-dvh w-full overflow-x-hidden bg-background p-2 text-foreground sm:p-3">{content}</main>;
}

function StatCard({ label, tone, value }: { label: string; tone?: "success" | "warning"; value: string }) {
  return (
    <Card className="shadow-sm">
      <CardHeader className="p-3 pb-1">
        <CardDescription>{label}</CardDescription>
        <CardTitle className={tone === "success" ? "text-emerald-600 dark:text-emerald-300" : tone === "warning" ? "text-amber-600 dark:text-amber-300" : ""}>
          {value}
        </CardTitle>
      </CardHeader>
    </Card>
  );
}

function CurrencyCard({
  baseCurrency,
  currency,
  isBaseFlag,
  isLastActiveFlag,
  onDelete,
  onEdit,
  onSetBase,
  text,
}: {
  baseCurrency: string;
  currency: CurrencyRecord;
  isBaseFlag: boolean;
  isLastActiveFlag: boolean;
  onDelete: (currency: CurrencyRecord) => void;
  onEdit: (currency: CurrencyRecord) => void;
  onSetBase: (code: string) => void;
  text: (key: CurrencyTextKey) => string;
}) {
  const isBase = baseCurrency && currency.code.toUpperCase() === baseCurrency;
  const verifiedPreset = findCurrencySymbolPreset(currency);
  const deleteBlockedReason = isBaseFlag
    ? text("baseGuardDelete")
    : isLastActiveFlag
      ? text("lastActiveGuard")
      : "";
  const rate = latestRateVsBase(currency, baseCurrency);

  return (
    <Card className="min-w-0 shadow-sm">
      <CardContent className="grid gap-2 p-3">
        <div className="flex min-w-0 items-start justify-between gap-2">
          <div className="flex min-w-0 items-center gap-2">
            <span className="grid size-11 shrink-0 place-items-center rounded-2xl bg-primary/10 text-xl font-semibold text-primary">
              {currency.symbol || "?"}
            </span>
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-1.5">
                <h2 className="truncate text-base font-semibold">{currency.code}</h2>
                <Badge variant={verifiedPreset ? "success" : "warning"}>
                  {verifiedPreset ? text("currencyVerified") : text("currencyNotVerified")}
                </Badge>
                {isBase ? <Badge variant="success">{text("baseBadge")}</Badge> : null}
                {currency.isdisabled ? <Badge variant="warning">{text("disabled")}</Badge> : null}
              </div>
              <p className="truncate text-sm text-muted-foreground">{currency.name}</p>
              {verifiedPreset ? (
                <p className="truncate text-xs text-muted-foreground">
                  {verifiedPreset.isoName} · {currencyPresetSource.standard} · {currencyPresetSource.symbolSource}
                </p>
              ) : null}
            </div>
          </div>
          <div className="flex shrink-0 gap-1">
            <Button type="button" variant="outline" size="icon" onClick={() => onEdit(currency)} aria-label={text("edit")} title={text("edit")}>
              <Edit3 />
            </Button>
            <Button
              type="button"
              variant="outline"
              size="icon"
              onClick={() => onDelete(currency)}
              disabled={Boolean(deleteBlockedReason)}
              aria-label={deleteBlockedReason || text("delete")}
              title={deleteBlockedReason || text("delete")}
            >
              <Trash2 />
            </Button>
          </div>
        </div>
        <div className="grid gap-1 text-xs text-muted-foreground">
          <span className="flex items-center justify-between gap-2 rounded-xl border border-border bg-background px-2 py-1.5">
            <span>{text("status")}</span>
            <b className="text-foreground">{currency.isdisabled ? text("disabled") : text("enabled")}</b>
          </span>
          <span className="flex items-center justify-between gap-2 rounded-xl border border-border bg-background px-2 py-1.5">
            <span>{text("rateVsBase")}</span>
            <b className="text-foreground">
              {isBase || !baseCurrency
                ? text("baseBadge")
                : rate
                  ? `1 ${currency.code} = ${rate.toLocaleString()} ${baseCurrency}`
                  : text("none")}
            </b>
          </span>
        </div>
        {!isBase ? (
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="w-full font-semibold"
            disabled={currency.isdisabled}
            onClick={() => onSetBase(currency.code)}
            title={text("setBase")}
          >
            <CircleDollarSign className="size-4" />
            {text("setBase")}
          </Button>
        ) : null}
      </CardContent>
    </Card>
  );
}

function CurrencyFormPanel({
  baseCurrency,
  editing,
  form,
  onClose,
  onFormChange,
  onSubmit,
  saving,
  text,
}: {
  baseCurrency: string;
  editing: CurrencyRecord | null;
  form: CurrencyForm;
  onClose: () => void;
  onFormChange: (form: CurrencyForm) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  saving: boolean;
  text: (key: CurrencyTextKey) => string;
}) {
  const [symbolQuery, setSymbolQuery] = useState("");
  const filteredSymbols = useMemo(() => filterCurrencySymbolPresets(symbolQuery), [symbolQuery]);
  const verifiedPreset = useMemo(() => findCurrencySymbolPreset(form), [form]);

  return (
    <Card className="shadow-sm">
      <form className="grid gap-3 p-3" onSubmit={onSubmit} aria-label={editing ? text("formEdit") : text("formNew")}>
        <header className="flex min-w-0 items-center justify-between gap-2">
          <div className="min-w-0">
            <h2 className="truncate text-lg font-semibold">{editing ? text("formEdit") : text("formNew")}</h2>
            {editing ? <p className="text-xs text-muted-foreground">{text("editLocked")}</p> : null}
          </div>
          <Button type="button" variant="outline" size="icon" onClick={onClose} disabled={saving} aria-label={text("cancel")}>
            <X />
          </Button>
        </header>

        <div className="grid gap-2">
          <div className="grid gap-2 sm:grid-cols-2">
            <label className="grid gap-1 text-sm font-semibold">
              <span>{text("code")} *</span>
              <Input
                autoFocus
                disabled={Boolean(editing)}
                maxLength={5}
                value={form.code}
                onChange={(event) => onFormChange({ ...form, code: event.target.value.toUpperCase() })}
                placeholder="USD"
              />
            </label>
            <label className="grid gap-1 text-sm font-semibold">
              <span>{text("name")} *</span>
              <Input
                value={form.name}
                onChange={(event) => onFormChange({ ...form, name: event.target.value })}
                placeholder="US Dollar"
              />
            </label>
          </div>

          <section className="grid gap-2 rounded-2xl border border-border bg-background p-2">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="min-w-0 flex-1 text-sm font-semibold">{text("symbol")} *</h3>
              <Badge variant={verifiedPreset ? "success" : "warning"}>
                {verifiedPreset ? text("currencyVerified") : text("currencyNotVerified")}
              </Badge>
            </div>
            <p className="text-xs text-muted-foreground">
              {text("currencyVerifySource")} {currencyPresetSource.list} ({currencyPresetSource.published})
            </p>
            <label className="relative block min-w-0">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                className="!pl-10"
                value={symbolQuery}
                onChange={(event) => setSymbolQuery(event.target.value)}
                placeholder={text("symbolSearch")}
              />
            </label>
            <div className="grid gap-1.5 sm:grid-cols-2 md:grid-cols-3 min-[900px]:grid-cols-4">
              {filteredSymbols.map((item) => {
                const selected = form.symbol === item.symbol && form.code.toUpperCase() === item.code;
                return (
                  <button
                    className={`grid min-h-11 grid-cols-[32px_minmax(0,1fr)] items-center gap-2 rounded-xl border px-2 text-left text-sm transition ${selected ? "border-primary bg-primary/10 text-primary" : "border-border bg-card text-foreground hover:bg-accent"}`}
                    key={item.code}
                    onClick={() => onFormChange(applyCurrencySymbolPreset(form, item, Boolean(editing)))}
                    type="button"
                  >
                    <span className="text-center text-xl font-semibold">{item.symbol}</span>
                    <span className="min-w-0 truncate text-xs text-muted-foreground">{item.name} ({item.code})</span>
                  </button>
                );
              })}
              {filteredSymbols.length ? null : (
                <div className="rounded-xl border border-dashed border-border bg-card px-3 py-4 text-center text-sm font-semibold text-muted-foreground sm:col-span-2 md:col-span-3 min-[900px]:col-span-4">
                  {text("noSymbolResult")}
                </div>
              )}
            </div>
          </section>

          {editing && baseCurrency && form.code !== baseCurrency ? (
            <label className="grid gap-1 text-sm font-semibold">
              <span>{text("rateVsBase")} (1 {form.code} = ? {baseCurrency})</span>
              <Input
                type="number"
                min="0"
                step="0.0001"
                value={form.rate || ""}
                onChange={(event) => onFormChange({ ...form, rate: Number(event.target.value) })}
                placeholder="เช่น 35.5"
              />
            </label>
          ) : null}
        </div>

        <footer className="flex flex-wrap justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose} disabled={saving}>
            {text("cancel")}
          </Button>
          <Button type="submit" disabled={saving}>
            {saving ? <Loader2 className="animate-spin" /> : <Save />}
            {saving ? text("saving") : text("save")}
          </Button>
        </footer>
      </form>
    </Card>
  );
}

function readAuth(): AuthSession | null {
  return getAuthSession();
}

function readWorkspace(): WorkspaceSession | null {
  try {
    const raw = localStorage.getItem(workspaceStorageKeys.workspace);
    if (!raw) return null;
    const workspace = JSON.parse(raw) as WorkspaceSession;
    return workspace?.shop?.holdingcode ? workspace : null;
  } catch {
    return null;
  }
}

function normalizeCurrencies(value: unknown): CurrencyRecord[] {
  if (!Array.isArray(value)) return [];
  const currencies = value.map(normalizeCurrency).filter((currency): currency is CurrencyRecord => Boolean(currency));
  const byCode = new Map<string, CurrencyRecord & { duplicate_guidfixeds: string[] }>();
  for (const currency of currencies) {
    const existing = byCode.get(currency.code);
    if (!existing) {
      byCode.set(currency.code, { ...currency, duplicate_guidfixeds: uniqueStrings([currency.guidfixed]) });
      continue;
    }

    existing.duplicate_guidfixeds = uniqueStrings([...existing.duplicate_guidfixeds, currency.guidfixed]);
    if ((existing.isdisabled && !currency.isdisabled) || (!existing.guidfixed && currency.guidfixed)) {
      byCode.set(currency.code, { ...currency, duplicate_guidfixeds: existing.duplicate_guidfixeds });
    }
  }
  return [...byCode.values()];
}

function normalizeCurrency(value: unknown): CurrencyRecord | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) return null;
  const item = value as Record<string, unknown>;
  const code = toStringValue(item.code).trim().toUpperCase();
  if (!code) return null;

  return {
    guidfixed: toStringValue(item.guidfixed || item.guidfixed || item.guidFixed || item.id || item._id),
    code,
    name: toStringValue(item.name),
    symbol: toStringValue(item.symbol),
    isdisabled: Boolean(item.isdisabled),
    exchangerates: Array.isArray(item.exchangerates) ? item.exchangerates as ExchangeRateEntry[] : [],
  };
}

function uniqueStrings(values: unknown[]): string[] {
  return Array.from(new Set(values.map((value) => toStringValue(value).trim()).filter(Boolean)));
}

function toStringValue(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function normalizeForm(form: CurrencyForm): CurrencyForm {
  return {
    code: form.code.trim().toUpperCase(),
    name: form.name.trim(),
    symbol: form.symbol.trim(),
    isdisabled: form.isdisabled,
    rate: Number(form.rate) || 0,
  };
}

/** อัตราล่าสุดเทียบสกุลหลักจากประวัติ exchangerates (เอา entry วันที่มากสุด) */
function latestRateVsBase(currency: CurrencyRecord, baseCurrency: string): number | null {
  if (!baseCurrency || currency.code.toUpperCase() === baseCurrency) return null;
  const rates = (currency.exchangerates ?? []).filter((entry) => Number(entry.rate) > 0);
  if (!rates.length) return null;
  const sorted = [...rates].sort((a, b) => String(b.date ?? "").localeCompare(String(a.date ?? "")));
  return Number(sorted[0].rate) || null;
}

/** append rate วันนี้เข้าประวัติ (ไม่ทับของเก่า — เก็บ history ตามแบบ backend) */
function upsertRateEntry(existing: ExchangeRateEntry[] | undefined, rate: number): ExchangeRateEntry[] {
  const today = new Date().toISOString().slice(0, 10);
  const history = (existing ?? []).filter((entry) => entry.date !== today);
  return [...history, { date: today, rate }];
}

function validateForm(form: CurrencyForm, text: (key: CurrencyTextKey) => string): string | null {
  if (!form.code) return text("requiredCode");
  if (form.code.length < 2 || form.code.length > 5) return text("codeLength");
  if (!form.name) return text("requiredName");
  if (!form.symbol) return text("selectSymbol");
  if (!findCurrencySymbolPreset(form)) return text("currencyVerifyRequired");
  return null;
}

function formatCount(value: number, language: LanguageCode): string {
  const localeByLanguage: Record<LanguageCode, string> = {
    th: "th-TH",
    en: "en-US",
    cn: "zh-CN",
    ja: "ja-JP",
    ko: "ko-KR",
    lo: "lo-LA",
    my: "my-MM",
    km: "km-KH",
    vi: "vi-VN",
    ms: "ms-MY",
    id: "id-ID",
    fil: "fil-PH",
  };
  return value.toLocaleString(localeByLanguage[language]);
}
