// Shared types for system-settings extracted components
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";
import type { LanguageCode } from "@/lib/i18n";
import type { CalendarYearType } from "@/lib/date-time";

export type SettingRecord = Record<string, unknown>;
export type FormState = Record<string, unknown>;

export type ProductUnitOption = {
  unitcode?: string;
  names?: { code?: string; name?: string }[];
};

// Subkeys copied between a `thai-address` field's flat form state (`${prefix}.${sub}`) and the
// nested record path (`field.key` = the Address object prefix, e.g. "addressforbilling").
export const THAI_ADDRESS_SUBKEYS = [
  "countrycode",
  "provincecode",
  "districtcode",
  "subdistrictcode",
  "zipcode",
] as const;

// Type guard for plain objects
export function isRecord(value: unknown): value is SettingRecord {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

// Generic string coercion used across editors
export function stringValue(value: unknown): string {
  return typeof value === "string"
    ? value.trim()
    : value === null || value === undefined
      ? ""
      : String(value).trim();
}

// Navigate a dot-separated path in a nested record
export function getByPath(record: unknown, path: string): unknown {
  if (!isRecord(record)) return undefined;
  return path
    .split(".")
    .reduce<unknown>(
      (current, key) => (isRecord(current) ? current[key] : undefined),
      record,
    );
}

// Read a potentially nested path (e.g. "contact.latitude") or a flat key from a record
export function getPathOrFlatValue(
  record: SettingRecord,
  path: string,
): unknown {
  const nested = getByPath(record, path);
  return nested ?? record[path];
}

// Standard auth headers for backend API calls
export function requestHeaders(auth: AuthSession): HeadersInit {
  return {
    "Content-Type": "application/json",
    "x-bc-backend-url": auth.backendUrl,
    Authorization: `Bearer ${auth.token}`,
  };
}

// Extract array of records from various API response shapes
export function extractListRecords(payload: unknown): SettingRecord[] {
  if (!isRecord(payload)) return Array.isArray(payload) ? payload.filter(isRecord) : [];
  const data = payload.data;
  if (Array.isArray(data)) return data.filter(isRecord);
  if (isRecord(data) && Array.isArray(data.data)) return data.data.filter(isRecord);
  return [];
}

// Check if API response indicates failure
export function isFailed(payload: unknown): boolean {
  if (!isRecord(payload)) return false;
  return payload.success === false || payload.status === "error";
}

// Extract error message from API response
export function extractMessage(payload: unknown): string | undefined {
  if (typeof payload === "string") return payload;
  if (!isRecord(payload)) return undefined;
  return typeof payload.message === "string"
    ? payload.message
    : typeof payload.error === "string"
      ? payload.error
      : undefined;
}

// Safe JSON parse with fallback
export function safeJsonParse(value: string, fallback: unknown): unknown {
  try {
    return JSON.parse(value);
  } catch {
    return fallback;
  }
}

// Map language code to locale string
export function localeOf(language: LanguageCode): string {
  const map: Record<LanguageCode, string> = {
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
  return map[language];
}

// Unique string deduplication
export function uniqueStrings(values: string[]): string[] {
  return Array.from(new Set(values.filter(Boolean)));
}

// Apply workspace tenant params to URL search params
export function applyWorkspaceTenantParams(params: URLSearchParams, workspace: WorkspaceSession): URLSearchParams {
  params.set("holdingcode", workspace.shop.holdingcode);
  const holdingCode = workspace.shop.holdingcode?.trim() || "";
  if (holdingCode) params.set("holdingcode", holdingCode);
  else params.delete("holdingcode");
  return params;
}

// DateTime scope for branch-specific settings
export type DateTimeScope = {
  key: string;
  branchcode: string;
  branchguid: string;
  timezone: string;
  timezonelabel: string;
  timezoneoffset: string;
  calendarYearType: CalendarYearType;
};

// Build payload from DateTimeScope
export function dateTimeScopePayload(scope: DateTimeScope): SettingRecord {
  return {
    branchkey: scope.key,
    branchcode: scope.branchcode,
    branchguid: scope.branchguid,
    timezone: scope.timezone,
    timezonelabel: scope.timezonelabel,
    timezoneoffset: scope.timezoneoffset,
    calendaryeartype: scope.calendarYearType,
  };
}

// Coerce various truthy representations to boolean
export function booleanLikeValue(value: unknown): boolean {
  if (typeof value === "boolean") return value;
  const normalized = stringValue(value).toLowerCase();
  return (
    normalized === "true" ||
    normalized === "1" ||
    normalized === "yes" ||
    normalized === "y" ||
    normalized === "buddhist" ||
    normalized === "be" ||
    normalized === "พ.ศ."
  );
}

// Extract localized name from a names array or string
export function localizedValue(value: unknown, language: LanguageCode): string {
  if (Array.isArray(value)) {
    const found = value.find(
      (item) =>
        isRecord(item) &&
        String(item.code).toLowerCase() === language &&
        typeof item.name === "string",
    );
    if (isRecord(found) && typeof found.name === "string") return found.name;
    const fallback = value.find(
      (item) => isRecord(item) && typeof item.name === "string",
    );
    return isRecord(fallback) && typeof fallback.name === "string"
      ? fallback.name
      : "";
  }
  if (typeof value === "string") return value;
  return "";
}

// Extract names array from unknown value
export function getLocalizedNameArray(
  value: unknown,
): Array<{ code?: string; name?: string }> {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord).map((item) => ({
    code: stringValue(item.code),
    name: stringValue(item.name),
  }));
}

// Get localized name for a specific language from names array
export function localizedNameForLanguage(
  names: Array<{ code?: string; name?: string }>,
  language: LanguageCode,
): string {
  return (
    names.find((item) => item.code === language && item.name)?.name ??
    names.find((item) => item.code === "th" && item.name)?.name ??
    names.find((item) => item.name)?.name ??
    ""
  );
}

// Move an array item from one index to another
export function moveArrayItem<T>(
  items: readonly T[],
  fromIndex: number,
  toIndex: number,
): T[] {
  const next = [...items];
  const [item] = next.splice(fromIndex, 1);
  if (item === undefined) return next;
  next.splice(toIndex, 0, item);
  return next;
}
