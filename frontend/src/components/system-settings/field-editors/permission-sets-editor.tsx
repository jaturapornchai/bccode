"use client";

import { authFetch } from "@/lib/client-auth-session";
import { Loader2, Search } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { backendText, type BackendLanguageDictionary } from "@/lib/backend-language";
import { getSystemSettingConfig, type SystemSettingField } from "@/lib/system-setting-screens";
import type { LanguageCode } from "@/lib/i18n";
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";
import { cn } from "@/lib/utils";
import {
  type FormState,
  applyWorkspaceTenantParams,
  extractMessage,
  isFailed,
  isRecord,
  localizedNameForLanguage,
  localeOf,
  requestHeaders,
  stringValue,
  uniqueStrings,
} from "../types";
import { stringArrayFromForm } from "./permission-editors";

type PermissionSetOption = {
  code: string;
  name: string;
  isActive: boolean;
  permissionCount: number;
  allScreens: boolean;
};

const BUILT_IN_SETS = new Set(["USER", "ADMIN", "OWNER"]);

function optionFromRecord(record: Record<string, unknown>, language: LanguageCode): PermissionSetOption {
  const permissions = Array.isArray(record.permissions) ? record.permissions : [];
  const names = Array.isArray(record.names) ? record.names.filter(isRecord) : [];
  return {
    code: stringValue(record.rolecode).toUpperCase(),
    name: localizedNameForLanguage(names as Array<{ code?: string; name?: string }>, language),
    isActive: record.isactive !== false,
    permissionCount: permissions.length,
    allScreens: permissions.includes("*"),
  };
}

/**
 * ชุดสิทธิ์เพิ่มเติมของคนในองค์กร: เลือกได้หลายชุดจาก role_permission ของ holding
 * (ขั้น 3 ตั้งค่าระบบ). สิทธิ์จริง = รวมทุกชุด + ชุดมาตรฐานของระดับสิทธิ์.
 */
export function PermissionSetsEditor({
  auth,
  dictionary,
  field,
  form,
  language,
  readOnly = false,
  setForm,
  workspace,
}: {
  auth: AuthSession | null;
  dictionary: BackendLanguageDictionary;
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  readOnly?: boolean;
  setForm?: (update: FormState | ((current: FormState) => FormState)) => void;
  workspace: WorkspaceSession | null;
}) {
  const sourceConfig = useMemo(() => getSystemSettingConfig("permissiongroup"), []);
  const [options, setOptions] = useState<PermissionSetOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");
  const selected = stringArrayFromForm(form[field.key]).map((code) => code.toUpperCase());
  const label = field.label[language] ?? field.label.en ?? field.label.th;
  const isThai = language === "th";

  useEffect(() => {
    let cancelled = false;
    async function load() {
      if (!auth || !workspace || !sourceConfig) {
        setOptions([]);
        return;
      }
      setLoading(true);
      setError("");
      try {
        const searchParams = new URLSearchParams({ limit: "1000", offset: "0", page: "1", q: "" });
        applyWorkspaceTenantParams(searchParams, workspace);
        const response = await authFetch(`/api/system-settings/${sourceConfig.slug}?${searchParams.toString()}`, {
          headers: requestHeaders(auth),
          cache: "no-store",
        });
        const payload = (await response.json()) as unknown;
        if (!response.ok || isFailed(payload)) {
          throw new Error(extractMessage(payload) ?? backendText(dictionary, "request_failed", "Request failed."));
        }
        if (cancelled) return;
        const data = isRecord(payload) ? payload.data : payload;
        const records = Array.isArray(data) ? data.filter(isRecord) : [];
        setOptions(
          records
            .map((record) => optionFromRecord(record, language))
            // built-in sets come from the access level radio, not from this picker
            .filter((option) => option.code && !BUILT_IN_SETS.has(option.code)),
        );
      } catch (loadError) {
        if (!cancelled) setError(loadError instanceof Error && loadError.message ? loadError.message : backendText(dictionary, "request_failed", "Request failed."));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [auth, dictionary, language, sourceConfig, workspace]);

  const query = search.trim().toLowerCase();
  const visible = query
    ? options.filter((option) => option.code.toLowerCase().includes(query) || option.name.toLowerCase().includes(query))
    : options;

  function toggle(code: string, checked: boolean) {
    if (readOnly || !setForm) return;
    const next = checked ? uniqueStrings([...selected, code]) : selected.filter((item) => item !== code);
    setForm((current) => ({ ...current, [field.key]: next }));
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-3 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>{label}</span>
        <Badge variant="outline">{selected.length.toLocaleString(localeOf(language))}</Badge>
      </div>
      <p className="text-xs font-normal text-muted-foreground">
        {isThai
          ? "เลือกได้หลายชุด สิทธิ์ที่ได้ = รวมทุกชุดที่เลือก + ชุดมาตรฐานของระดับสิทธิ์ · สร้าง/แก้ชุดได้ที่ขั้น 3 ชุดสิทธิ์"
          : "Pick any number of sets. Effective rights = union of the chosen sets + the access-level default. Manage sets in step 3."}
      </p>
      {options.length > 6 ? (
        <label className="relative block">
          <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="pl-9"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder={isThai ? "ค้นหาชุดสิทธิ์ (รหัส/ชื่อ)" : "Search permission sets"}
          />
        </label>
      ) : null}
      {error ? (
        <p className="rounded-xl border border-destructive/40 bg-destructive/10 px-2 py-1 text-xs text-destructive">{error}</p>
      ) : null}
      <div className="grid gap-2 md:grid-cols-2 xl:grid-cols-3">
        {loading ? (
          <div className="flex min-h-20 items-center gap-2 rounded-xl border border-border bg-card p-3 text-muted-foreground md:col-span-2 xl:col-span-3">
            <Loader2 className="animate-spin" />
            {backendText(dictionary, "loading", "Loading data")}
          </div>
        ) : visible.length ? (
          visible.map((option) => {
            const checked = selected.includes(option.code);
            return (
              <label
                key={option.code}
                className={cn(
                  "grid cursor-pointer gap-1 rounded-xl border p-2 transition-colors",
                  checked ? "border-primary bg-primary/10 text-primary" : "border-border bg-card text-foreground hover:bg-muted/60",
                  !option.isActive && "opacity-60",
                )}
              >
                <span className="flex min-w-0 items-center gap-2">
                  <input
                    className="size-4 shrink-0 accent-primary"
                    type="checkbox"
                    checked={checked}
                    disabled={readOnly}
                    onChange={(event) => toggle(option.code, event.target.checked)}
                  />
                  <span className="min-w-0 truncate font-semibold">{option.name || option.code}</span>
                </span>
                <span className="truncate text-xs font-normal text-muted-foreground">
                  {option.code} ·{" "}
                  {option.allScreens
                    ? isThai ? "ทุกจอ" : "All screens"
                    : isThai
                      ? `${option.permissionCount.toLocaleString("th-TH")} รายการสิทธิ์`
                      : `${option.permissionCount.toLocaleString("en-US")} entries`}
                  {!option.isActive ? (isThai ? " · ปิดใช้งาน" : " · inactive") : ""}
                </span>
              </label>
            );
          })
        ) : (
          <div className="rounded-xl border border-dashed border-border bg-card p-3 text-sm font-normal text-muted-foreground md:col-span-2 xl:col-span-3">
            {options.length
              ? isThai ? "ไม่พบชุดสิทธิ์ที่ค้นหา" : "No matching permission set"
              : isThai
                ? "ยังไม่มีชุดสิทธิ์เพิ่มเติม — สร้างได้ที่ขั้น 3 ชุดสิทธิ์ (เช่น บัญชี, ขาย, คลัง)"
                : "No custom permission sets yet — create them in step 3 (e.g. Accounting, Sales, Stock)."}
          </div>
        )}
      </div>
    </section>
  );
}
