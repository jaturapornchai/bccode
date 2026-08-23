"use client";

import { authFetch } from "@/lib/client-auth-session";
import { Check, Loader2, UserRound } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { LogoAvatar } from "@/components/logo-avatar";
import { Input } from "@/components/ui/input";
import {
  backendText,
  type BackendLanguageDictionary,
} from "@/lib/backend-language";
import {
  getSystemSettingConfig,
  type SystemSettingConfig,
  type SystemSettingField,
} from "@/lib/system-setting-screens";
import { MENU_SECTIONS, menuText } from "@/lib/menu-data";
import type { LanguageCode } from "@/lib/i18n";
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";
import { cn } from "@/lib/utils";
import {
  type DateTimeScope,
  type FormState,
  type SettingRecord,
  applyWorkspaceTenantParams,
  dateTimeScopePayload,
  extractMessage,
  isFailed,
  isRecord,
  localeOf,
  requestHeaders,
  safeJsonParse,
  stringValue,
  uniqueStrings,
} from "../types";

// ---------------------------------------------------------------------------
// Field detection helpers
// ---------------------------------------------------------------------------

export function isPermissionAccessRulesField(field: SystemSettingField): boolean {
  return field.key === "accessrules" || field.key === "branches";
}

export function isPermissionCodesField(field: SystemSettingField): boolean {
  return field.key === "permissioncodes" || field.key === "permissionCodes";
}

export function isApprovalCodesField(field: SystemSettingField): boolean {
  return field.key === "approvalcodes" || field.key === "approvalCodes";
}

export function isEmployeeCodeField(field: SystemSettingField): boolean {
  return field.key === "employeecode" || field.key === "employeeCode";
}

export function isEmployeeNameField(field: SystemSettingField): boolean {
  return field.key === "employeename" || field.key === "employeeName";
}

// ---------------------------------------------------------------------------
// Permission-specific constants
// ---------------------------------------------------------------------------

export type MenuPermissionAction =
  | "access"
  | "create"
  | "update"
  | "delete"
  | "own_only";

const permissionActionDefs: {
  key: MenuPermissionAction;
  textKey: string;
}[] = [
  { key: "access", textKey: "readAccess" },
  { key: "create", textKey: "writeAccess" },
  { key: "update", textKey: "updateAccess" },
  { key: "delete", textKey: "delete" },
  { key: "own_only", textKey: "selfOnly" },
];

const permissionUiEn: Record<string, string> = {
  readAccess: "Access",
  writeAccess: "Add",
  updateAccess: "Edit",
  delete: "Delete",
  selfOnly: "Own data only",
  allBranches: "All branches",
  branch: "Branch",
};

const permissionUiTh: Record<string, string> = {
  readAccess: "เข้าถึง",
  writeAccess: "เพิ่มได้",
  updateAccess: "แก้ไขได้",
  delete: "ลบ",
  selfOnly: "เห็นข้อมูลตัวเองเท่านั้น",
  allBranches: "ใช้กับทุกสาขา",
  branch: "สาขา",
};

function permissionActionText(
  dictionary: BackendLanguageDictionary,
  language: LanguageCode,
  key: string,
): string {
  const fallback = language === "th"
    ? (permissionUiTh[key] ?? permissionUiEn[key] ?? key)
    : (permissionUiEn[key] ?? key);
  const backendKeyMap: Record<string, string> = {
    readAccess: "access",
    writeAccess: "can_add",
    updateAccess: "can_edit",
    delete: "delete",
    selfOnly: "own_data_only",
    allBranches: "allbranches",
  };
  const backendKey = backendKeyMap[key] ?? key;
  const value = backendText(dictionary, backendKey, fallback);
  return value === backendKey || value === key ? fallback : value;
}

// Backend keys for permissionlink fields
const permissionFieldBackendKeys: Record<string, string> = {
  "permissionlink.employeecode": "user_employeecode",
  "permissionlink.employeename": "name",
  "permissionlink.scoperules": "scoperules",
  "permissionlink.permissioncodes": "permissioncodes",
  "permissionlink.approvalcodes": "approvalcodes",
};

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type PermissionLinkUserOption = {
  code: string;
  userUid: string;
  name: string;
  subtitle: string;
  avatar: string;
  isDisabled: boolean;
};

export type PermissionLinkOption = {
  code: string;
  description: string;
  isActive: boolean;
  name: string;
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function permissionLinkUserOption(
  record: SettingRecord,
): PermissionLinkUserOption {
  const userUid = stringValue(record.useruid ?? record.uid);
  const code = stringValue(record.user_name ?? record.username ?? record.employeecode ?? userUid);
  const name =
    stringValue(record.userprofilename ?? record.name ?? record.email ?? record.username) || code;
  const subtitle = stringValue(record.email) || userUid;
  const avatar = stringValue(record.avatarthumb) || stringValue(record.avatar);
  return {
    code,
    userUid,
    name,
    subtitle,
    avatar,
    isDisabled: Boolean(
      record.isaccessdisabled ?? record.isaccessdisabled ?? false,
    ),
  };
}

function permissionLinkOption(
  record: SettingRecord,
  isApproval: boolean,
): PermissionLinkOption {
  const code = stringValue(
    isApproval
      ? record.approvalcode ?? record.approvalCode
      : record.permissioncode ?? record.permissionCode,
  );
  const name =
    stringValue(
      isApproval
        ? record.approvalname ?? record.approvalName
        : record.permissionname ?? record.permissionName,
    ) ||
    code;
  const description = stringValue(record.description);
  return {
    code,
    description,
    isActive: Boolean(record.isactive ?? record.isActive ?? record.isactive ?? true),
    name,
  };
}

export function stringArrayFromForm(value: unknown): string[] {
  if (Array.isArray(value))
    return uniqueStrings(value.map(stringValue).filter(Boolean));
  if (typeof value !== "string") return [];
  const trimmed = value.trim();
  if (!trimmed) return [];
  try {
    const parsed = JSON.parse(trimmed) as unknown;
    return Array.isArray(parsed)
      ? uniqueStrings(parsed.map(stringValue).filter(Boolean))
      : [];
  } catch {
    return uniqueStrings(
      trimmed
        .split(/[,\n]/)
        .map((item) => item.trim())
        .filter(Boolean),
    );
  }
}

export function normalizeStringListValue(value: unknown): string[] {
  if (Array.isArray(value))
    return uniqueStrings(value.map(stringValue).map((item) => item.trim()).filter(Boolean));
  if (typeof value !== "string") return [];
  const trimmed = value.trim();
  if (!trimmed) return [];
  try {
    const parsed = JSON.parse(trimmed) as unknown;
    if (Array.isArray(parsed)) return normalizeStringListValue(parsed);
  } catch {
    // plain text input; split below
  }
  return uniqueStrings(
    trimmed
      .split(/[,\n]/)
      .map((item) => item.trim())
      .filter(Boolean),
  );
}

function permissionBranchesFromForm(value: unknown): SettingRecord {
  if (isRecord(value)) return { ...value };
  if (typeof value !== "string" || !value.trim()) return {};
  const parsed = safeJsonParse(value, {});
  return isRecord(parsed) ? { ...parsed } : {};
}

function permissionBranchValue(
  branches: SettingRecord,
  branchKey: string,
): SettingRecord {
  const branch = branches[branchKey];
  return isRecord(branch) ? { ...branch } : {};
}

// ---------------------------------------------------------------------------
// PermissionLinkUserSelector
// ---------------------------------------------------------------------------

export function PermissionLinkUserSelector({
  auth,
  dictionary,
  field,
  form,
  language,
  normalizeRecords,
  setForm,
  workspace,
}: {
  auth: AuthSession | null;
  dictionary: BackendLanguageDictionary;
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  normalizeRecords: (payload: unknown, config: SystemSettingConfig) => SettingRecord[];
  setForm: (form: FormState) => void;
  workspace: WorkspaceSession | null;
}) {
  const userConfig = useMemo(() => getSystemSettingConfig("user"), []);
  const [users, setUsers] = useState<PermissionLinkUserOption[]>([]);
  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [selectedAvatar, setSelectedAvatar] = useState("");
  const selectedCode = stringValue(form.employeecode ?? form.employeeCode);
  const selectedName = stringValue(form.employeename ?? form.employeeName);
  const fallbackLabel =
    field.label[language] ?? field.label.en ?? field.label.th;
  const labelKey =
    permissionFieldBackendKeys[`permissionlink.${field.key}`] ?? field.key;
  const translatedLabel = backendText(dictionary, labelKey, fallbackLabel);
  const label =
    translatedLabel === labelKey || translatedLabel === field.key
      ? fallbackLabel
      : translatedLabel;
  const searchText = debouncedQuery.trim();
  const canSearch = searchText.length >= 2;

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), 250);
    return () => window.clearTimeout(timer);
  }, [query]);

  useEffect(() => {
    let cancelled = false;
    async function loadUsers() {
      if (!auth || !workspace || !userConfig || !canSearch) {
        setUsers([]);
        setLoading(false);
        setError("");
        return;
      }
      setLoading(true);
      setError("");
      try {
        const searchParams = new URLSearchParams({
          limit: "20",
          offset: "0",
          q: searchText,
        });
        applyWorkspaceTenantParams(searchParams, workspace);
        const response = await authFetch(
          `/api/system-settings/${userConfig.slug}?${searchParams.toString()}`,
          {
            headers: requestHeaders(auth),
            cache: "no-store",
          },
        );
        const payload = (await response.json()) as unknown;
        if (!response.ok || isFailed(payload))
          throw new Error(
            extractMessage(payload) ??
              backendText(dictionary, "request_failed", "Request failed."),
          );
        if (cancelled) return;
        setUsers(
          normalizeRecords(payload, userConfig)
            .map(permissionLinkUserOption)
            .filter((user) => user.code && user.userUid),
        );
      } catch (loadError) {
        if (!cancelled)
          setError(
            loadError instanceof Error && loadError.message
              ? loadError.message
              : backendText(dictionary, "request_failed", "Request failed."),
          );
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void loadUsers();
    return () => {
      cancelled = true;
    };
  }, [auth, canSearch, dictionary, normalizeRecords, searchText, userConfig, workspace]);

  function choose(user: PermissionLinkUserOption) {
    if (user.isDisabled) return;
    setForm({ ...form, employeecode: user.code, employeename: user.name, useruid: user.userUid });
    setSelectedAvatar(user.avatar);
    setQuery("");
    setUsers([]);
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-3 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>
          {label}
          {field.required ? " *" : ""}
        </span>
        {selectedCode ? (
          <Badge variant="success">{selectedCode}</Badge>
        ) : (
          <Badge variant="outline">
            {backendText(dictionary, "select_user", "Select user")}
          </Badge>
        )}
      </div>
      <Input
        value={query}
        onChange={(event) => setQuery(event.target.value)}
        placeholder={backendText(dictionary, "search", "Search")}
      />
      {selectedCode ? (
        <div className="flex items-center gap-2 rounded-xl border border-primary/40 bg-primary/10 p-2 text-primary">
          <LogoAvatar
            uri={selectedAvatar}
            auth={auth}
            alt={selectedName || selectedCode}
            sizeClass="size-9 rounded-full shrink-0"
            iconSize={18}
            width={64}
            fallbackIcon={UserRound}
          />
          <span className="grid min-w-0 gap-1">
            <span className="break-words font-semibold">
              {selectedName || selectedCode}
            </span>
            <span className="break-words text-xs">
              {language === "th" ? "รหัสผู้ใช้" : "User code"}: {selectedCode}
            </span>
          </span>
        </div>
      ) : null}
      {error ? (
        <p className="rounded-xl border border-destructive/40 bg-destructive/10 px-2 py-1 text-xs text-destructive">
          {error}
        </p>
      ) : null}
      <div className="grid max-h-60 gap-2 overflow-y-auto pr-1 md:grid-cols-2">
        {!canSearch ? (
          <div className="rounded-xl border border-border bg-card p-3 text-sm text-muted-foreground md:col-span-2">
            {backendText(
              dictionary,
              "search_user_hint",
              "Type at least 2 characters to search users.",
            )}
          </div>
        ) : loading ? (
          <div className="flex min-h-16 items-center gap-2 rounded-xl border border-border bg-card p-3 text-muted-foreground md:col-span-2">
            <Loader2 className="animate-spin" />
            {backendText(dictionary, "loading", "Loading data")}
          </div>
        ) : users.length ? (
          users.map((user) => {
            const checked = selectedCode === user.code;
            return (
              <button
                className={cn(
                  "grid min-h-16 gap-1 rounded-xl border p-2 text-left transition-colors",
                  checked
                    ? "border-primary bg-primary/10 text-primary"
                    : "border-border bg-card text-foreground hover:bg-muted/60",
                  user.isDisabled && "cursor-not-allowed opacity-50",
                )}
                disabled={user.isDisabled}
                key={user.code}
                onClick={() => choose(user)}
                type="button"
              >
                <span className="flex min-w-0 items-center gap-2">
                  {checked ? (
                    <Check className="size-4 shrink-0" />
                  ) : (
                    <LogoAvatar
                      uri={user.avatar}
                      auth={auth}
                      alt={user.name}
                      sizeClass="size-7 rounded-full shrink-0"
                      iconSize={14}
                      width={48}
                      fallbackIcon={UserRound}
                    />
                  )}
                  <span className="min-w-0 break-words font-semibold">
                    {user.name}
                  </span>
                </span>
                <span className="break-words text-xs text-muted-foreground">
                  {language === "th" ? "รหัสผู้ใช้" : "User code"}: {user.code}
                </span>
                {user.subtitle ? (
                  <span className="truncate text-xs font-normal text-muted-foreground">
                    {user.subtitle}
                  </span>
                ) : null}
              </button>
            );
          })
        ) : (
          <div className="rounded-xl border border-border bg-card p-3 text-sm text-muted-foreground md:col-span-2">
            {backendText(dictionary, "empty_data", "No data")}
          </div>
        )}
      </div>
    </section>
  );
}

// ---------------------------------------------------------------------------
// PermissionLinkMultiSelectEditor
// ---------------------------------------------------------------------------

export function PermissionLinkMultiSelectEditor({
  config,
  auth,
  dictionary,
  field,
  form,
  language,
  normalizeRecords,
  readOnly = false,
  setForm,
  workspace,
}: {
  config?: SystemSettingConfig;
  auth: AuthSession | null;
  dictionary: BackendLanguageDictionary;
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  normalizeRecords: (payload: unknown, config: SystemSettingConfig) => SettingRecord[];
  readOnly?: boolean;
  setForm?: (form: FormState) => void;
  workspace: WorkspaceSession | null;
}) {
  const isApproval = isApprovalCodesField(field);
  const sourceConfig = useMemo(
    () =>
      getSystemSettingConfig(
        isApproval ? "approvalsetting" : "permissiondefinition",
      ),
    [isApproval],
  );
  const [options, setOptions] = useState<PermissionLinkOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const selectedCodes = stringArrayFromForm(form[field.key]);
  const isGroup = config?.slug === "permissiongroup";
  const selectedUserCode = isGroup ? "group" : stringValue(form.employeecode ?? form.employeeCode);
  const fallbackLabel =
    field.label[language] ?? field.label.en ?? field.label.th;
  const labelKey =
    permissionFieldBackendKeys[`permissionlink.${field.key}`] ?? field.key;
  const translatedLabel = backendText(dictionary, labelKey, fallbackLabel);
  const label =
    translatedLabel === labelKey || translatedLabel === field.key
      ? fallbackLabel
      : translatedLabel;

  useEffect(() => {
    let cancelled = false;
    async function loadOptions() {
      if (!auth || !workspace || !sourceConfig) {
        setOptions([]);
        return;
      }
      setLoading(true);
      setError("");
      try {
        const searchParams = new URLSearchParams({
          limit: "1000",
          offset: "0",
        });
        applyWorkspaceTenantParams(searchParams, workspace);
        const response = await authFetch(
          `/api/system-settings/${sourceConfig.slug}?${searchParams.toString()}`,
          {
            headers: requestHeaders(auth),
            cache: "no-store",
          },
        );
        const payload = (await response.json()) as unknown;
        if (!response.ok || isFailed(payload))
          throw new Error(
            extractMessage(payload) ??
              backendText(dictionary, "request_failed", "Request failed."),
          );
        if (cancelled) return;
        setOptions(
          normalizeRecords(payload, sourceConfig)
            .map((record) => permissionLinkOption(record, isApproval))
            .filter((option) => option.code),
        );
      } catch (loadError) {
        if (!cancelled)
          setError(
            loadError instanceof Error && loadError.message
              ? loadError.message
              : backendText(dictionary, "request_failed", "Request failed."),
          );
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void loadOptions();
    return () => {
      cancelled = true;
    };
  }, [auth, dictionary, isApproval, normalizeRecords, sourceConfig, workspace]);

  function toggle(code: string, checked: boolean) {
    if (readOnly || !setForm || !selectedUserCode) return;
    const next = checked
      ? uniqueStrings([...selectedCodes, code])
      : selectedCodes.filter((item) => item !== code);
    setForm({ ...form, [field.key]: next });
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-3 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>
          {label}
          {field.required ? " *" : ""}
        </span>
        <Badge variant="outline">
          {selectedCodes.length.toLocaleString(localeOf(language))}
        </Badge>
      </div>
      {!selectedUserCode ? (
        <p className="rounded-xl border border-amber-300 bg-amber-50 px-2 py-1 text-xs text-amber-800 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-200">
          {backendText(dictionary, "select_user_first", "Select user first")}
        </p>
      ) : null}
      {error ? (
        <p className="rounded-xl border border-destructive/40 bg-destructive/10 px-2 py-1 text-xs text-destructive">
          {error}
        </p>
      ) : null}
      <div className="grid gap-2 md:grid-cols-2">
        {loading ? (
          <div className="flex min-h-20 items-center gap-2 rounded-xl border border-border bg-card p-3 text-muted-foreground md:col-span-2">
            <Loader2 className="animate-spin" />
            {backendText(dictionary, "loading", "Loading data")}
          </div>
        ) : options.length ? (
          options.map((option) => {
            const checked = selectedCodes.includes(option.code);
            return (
              <label
                className={cn(
                  "grid cursor-pointer gap-1 rounded-xl border p-2 transition-colors",
                  checked
                    ? "border-primary bg-primary/10 text-primary"
                    : "border-border bg-card text-foreground hover:bg-muted/60",
                  !option.isActive && "opacity-60",
                )}
                key={option.code}
              >
                <span className="flex min-w-0 items-center gap-2">
                  <input
                    className="size-4 shrink-0 accent-primary"
                    type="checkbox"
                    checked={checked}
                    disabled={readOnly || !selectedUserCode}
                    onChange={(event) =>
                      toggle(option.code, event.target.checked)
                    }
                  />
                  <span className="min-w-0 truncate font-semibold">
                    {option.name || option.code}
                  </span>
                </span>
                <span className="truncate text-xs text-muted-foreground">
                  {language === "th" ? "รหัสสิทธิ์" : "Permission code"}: {option.code}
                </span>
                {option.description ? (
                  <span className="line-clamp-2 text-xs font-normal text-muted-foreground">
                    {option.description}
                  </span>
                ) : null}
              </label>
            );
          })
        ) : (
          <div className="rounded-xl border border-border bg-card p-3 text-sm text-muted-foreground md:col-span-2">
            {backendText(dictionary, "empty_data", "No data")}
          </div>
        )}
      </div>
    </section>
  );
}

// ---------------------------------------------------------------------------
// PermissionMatrixEditor
// ---------------------------------------------------------------------------

export function PermissionMatrixEditor({
  dateTimeScope,
  dictionary,
  form,
  language,
  readOnly = false,
  setForm,
}: {
  dateTimeScope: DateTimeScope;
  dictionary: BackendLanguageDictionary;
  form: FormState;
  language: LanguageCode;
  readOnly?: boolean;
  setForm?: (form: FormState) => void;
}) {
  const branchKey = dateTimeScope.key || "company";
  const branches = permissionBranchesFromForm(form.accessrules ?? form.branches);
  const branchPermission = permissionBranchValue(branches, branchKey);
  const menus = isRecord(branchPermission.menus) ? branchPermission.menus : {};
  const branchLabel = dateTimeScope.branchcode || dateTimeScope.branchguid || branchKey;
  const scopeHint =
    language === "th"
      ? "ติ๊กใช้กับทุกสาขา = สิทธิ์เมนูนี้ใช้ได้ทุกสาขา; ไม่ติ๊ก = ใช้เฉพาะสาขาปัจจุบัน"
      : "Checked all branches = this menu permission applies to every branch; unchecked = current branch only.";

  function updateMenuRule(menuId: string, patch: SettingRecord) {
    if (readOnly || !setForm) return;
    const nextBranches = permissionBranchesFromForm(form.accessrules ?? form.branches);
    const nextBranch = permissionBranchValue(nextBranches, branchKey);
    const nextMenus = isRecord(nextBranch.menus) ? { ...nextBranch.menus } : {};
    const currentMenu = isRecord(nextMenus[menuId])
      ? { ...nextMenus[menuId] }
      : {};
    nextMenus[menuId] = { ...currentMenu, ...patch };
    nextBranches[branchKey] = {
      ...nextBranch,
      ...dateTimeScopePayload(dateTimeScope),
      menus: nextMenus,
    };
    setForm({ ...form, accessrules: nextBranches });
  }

  function updateMenuPermission(
    menuId: string,
    action: MenuPermissionAction,
    checked: boolean,
  ) {
    updateMenuRule(menuId, { [action]: checked });
  }

  function updateMenuAllBranches(menuId: string, checked: boolean) {
    updateMenuRule(menuId, { allbranches: checked });
  }

  const allBranchesLabel = language === "th"
    ? permissionUiTh.allBranches
    : permissionUiEn.allBranches;
  const branchWord = language === "th"
    ? permissionUiTh.branch
    : permissionUiEn.branch;

  return (
    <section className="grid gap-3 rounded-2xl border border-border bg-background p-3 md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h3 className="text-base font-semibold">
            {backendText(
              dictionary,
              "permissiondefinition",
              "Permission Definition",
            )}
          </h3>
          <p className="text-xs text-muted-foreground">
            {branchWord}: {branchLabel}
          </p>
          <p className="text-xs text-muted-foreground">{scopeHint}</p>
        </div>
        <Badge variant="outline">
          {
            MENU_SECTIONS.flatMap((section) =>
              section.groups.flatMap((group) => group.items),
            ).length
          }{" "}
          {language === "th" ? "เมนู" : "menus"}
        </Badge>
      </div>

      <div className="grid gap-3">
        {MENU_SECTIONS.map((section) => (
          <section
            className="grid gap-2 rounded-2xl border border-border bg-card p-2"
            key={section.id}
          >
            <h4 className="text-sm font-semibold">
              {menuText(section.title, language, dictionary)}
            </h4>
            {section.groups.map((group) => (
              <div
                className="grid gap-1 rounded-xl border border-border bg-background p-2"
                key={group.id}
              >
                <p className="text-xs font-semibold text-muted-foreground">
                  {menuText(group.title, language, dictionary)}
                </p>
                {group.items.map((item) => {
                  const permission: SettingRecord = isRecord(menus[item.id])
                    ? (menus[item.id] as SettingRecord)
                    : {};
                  const allBranches = Boolean(
                    permission.allbranches ??
                      permission.allBranches ??
                      permission.useallbranches,
                  );
                  return (
                    <div
                      className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2.5 xl:grid-cols-[minmax(180px,1fr)_auto] items-center shadow-[0_1px_2px_rgba(0,0,0,0.01)] hover:bg-secondary/5 transition-colors duration-150"
                      key={item.id}
                    >
                      <div className="min-w-0">
                        <p className="break-words text-sm font-bold text-foreground">
                          {menuText(item.label, language, dictionary)}
                        </p>
                        <p className="break-words text-[10px] text-muted-foreground mt-0.5">
                          <span className="font-mono bg-secondary/35 px-1 py-0.5 rounded text-primary">
                            {language === "th" ? "รหัสเมนู" : "Menu code"}: {item.id}
                          </span>
                        </p>
                      </div>
                      <div className="grid grid-cols-2 gap-1.5 sm:grid-cols-3 lg:grid-cols-6 md:gap-2">
                        <label
                          className="inline-flex h-8 w-auto max-w-none items-center gap-1.5 rounded-lg border border-border/80 bg-background/50 hover:bg-secondary/5 px-2 text-xs font-semibold cursor-pointer transition-colors shadow-sm"
                        >
                          <input
                            className="size-3.5 accent-primary shrink-0"
                            type="checkbox"
                            checked={allBranches}
                            disabled={readOnly}
                            onChange={(event) =>
                              updateMenuAllBranches(
                                item.id,
                                event.target.checked,
                              )
                            }
                          />
                          <span className="truncate text-foreground/85">
                            {allBranchesLabel}
                          </span>
                        </label>
                        {permissionActionDefs.map((action) => (
                          <label
                            className="inline-flex h-8 w-auto max-w-none items-center gap-1.5 rounded-lg border border-border/80 bg-background/50 hover:bg-secondary/5 px-2 text-xs font-semibold cursor-pointer transition-colors shadow-sm"
                            key={action.key}
                          >
                            <input
                              className="size-3.5 accent-primary shrink-0"
                              type="checkbox"
                              checked={Boolean(permission[action.key])}
                              disabled={readOnly}
                              onChange={(event) =>
                                updateMenuPermission(
                                  item.id,
                                  action.key,
                                  event.target.checked,
                                )
                              }
                            />
                            <span className="truncate text-foreground/85">
                              {permissionActionText(
                                dictionary,
                                language,
                                action.textKey,
                              )}
                            </span>
                          </label>
                        ))}
                      </div>
                    </div>
                  );
                })}
              </div>
            ))}
          </section>
        ))}
      </div>
    </section>
  );
}
