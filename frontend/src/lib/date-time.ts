import type { LanguageCode } from "@/lib/i18n";
import type { WorkspaceSession } from "@/lib/workspace-models";

export type CalendarYearType = "buddhist" | "christian";
export type DateTimeDisplayOptions = {
  language?: LanguageCode;
  yearType?: CalendarYearType;
  timeZone?: string;
  fallback?: string;
  includeSeconds?: boolean;
};

type DateTimeInput = string | number | Date | null | undefined;

export const DEFAULT_DATE_LANGUAGE: LanguageCode = "th";
export const DEFAULT_CALENDAR_YEAR_TYPE: CalendarYearType = "buddhist";
export const DEFAULT_TIME_ZONE = "Asia/Bangkok";

const languageLocales: Record<LanguageCode, string> = {
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

export function localeForDate(language: LanguageCode, yearType: CalendarYearType): string {
  const locale = languageLocales[language] ?? languageLocales.en;
  return `${locale}-u-ca-${yearType === "buddhist" ? "buddhist" : "gregory"}`;
}

export function formatLocalDate(value: string, language: LanguageCode, yearType: CalendarYearType): string {
  if (!value) return "";
  const [year, month, day] = value.split("-").map(Number);
  if (!year || !month || !day) return value;
  const date = new Date(Date.UTC(year, month - 1, day, 12, 0, 0));
  return new Intl.DateTimeFormat(localeForDate(language, yearType), {
    day: "2-digit",
    month: "short",
    year: "numeric",
    timeZone: "UTC",
  }).format(date);
}

export function formatDefaultDate(value: DateTimeInput, options: DateTimeDisplayOptions = {}): string {
  return formatDateTimeValue(value, { ...options, includeSeconds: false }, false);
}

export function formatDefaultDateTime(value: DateTimeInput, options: DateTimeDisplayOptions = {}): string {
  return formatDateTimeValue(value, { includeSeconds: true, ...options }, true);
}

export function resolveWorkspaceDateTimeDisplayOptions(
  workspace: WorkspaceSession | null | undefined,
  language: LanguageCode,
): Required<Pick<DateTimeDisplayOptions, "language" | "yearType" | "timeZone">> {
  const branch = workspace?.branch ?? null;
  const shopInfo = isRecord(workspace?.shopInfo) ? workspace.shopInfo : {};
  const branchYear = stringValue(branch?.yeartype).toLowerCase();
  const useBuddhistCalendar = booleanLikeValue(getByPath(shopInfo, "settings.usebuddhistcalendar"));
  return {
    language,
    yearType:
      branchYear === "buddhist" || branchYear === "be" || branchYear === "พ.ศ."
        ? "buddhist"
        : branchYear === "christian" || branchYear === "ce" || branchYear === "ค.ศ."
          ? "christian"
          : useBuddhistCalendar,
    timeZone: stringValue(branch?.timezone) || stringValue(getByPath(shopInfo, "settings.timezone")) || DEFAULT_TIME_ZONE,
  };
}

export function normalizeTimeInput(value: unknown, fallback = ""): string {
  const raw = String(value ?? "").trim();
  if (!raw) return fallback;
  const hhmm = raw.match(/^(\d{1,2}):(\d{2})$/);
  if (hhmm) return `${hhmm[1].padStart(2, "0")}:${hhmm[2]}`;
  const compact = raw.match(/^(\d{3,4})$/);
  if (compact) {
    const padded = compact[1].padStart(4, "0");
    return `${padded.slice(0, 2)}:${padded.slice(2, 4)}`;
  }
  return fallback;
}

export function parseUtcOffsetMinutes(offset: string): number | null {
  const match = offset.trim().match(/^([+-])(\d{2}):?(\d{2})$/);
  if (!match) return null;
  const sign = match[1] === "-" ? -1 : 1;
  return sign * (Number(match[2]) * 60 + Number(match[3]));
}

export function localTimeToUtcTime(value: string, utcOffset: string): { time: string; dayOffset: number } | null {
  const normalized = normalizeTimeInput(value);
  const offsetMinutes = parseUtcOffsetMinutes(utcOffset);
  if (!normalized || offsetMinutes === null) return null;
  const minutes = timeToMinutes(normalized);
  if (!Number.isFinite(minutes)) return null;
  let utcMinutes = minutes - offsetMinutes;
  let dayOffset = 0;
  while (utcMinutes < 0) {
    utcMinutes += 1440;
    dayOffset -= 1;
  }
  while (utcMinutes >= 1440) {
    utcMinutes -= 1440;
    dayOffset += 1;
  }
  return { time: minutesToTime(utcMinutes), dayOffset };
}

export function timeToMinutes(value: string): number {
  const normalized = normalizeTimeInput(value);
  if (!normalized) return Number.NaN;
  const [hours, minutes] = normalized.split(":").map(Number);
  if (!Number.isFinite(hours) || !Number.isFinite(minutes)) return Number.NaN;
  return hours * 60 + minutes;
}

export function minutesToTime(totalMinutes: number): string {
  const hours = Math.floor(totalMinutes / 60).toString().padStart(2, "0");
  const minutes = (totalMinutes % 60).toString().padStart(2, "0");
  return `${hours}:${minutes}`;
}

function formatDateTimeValue(value: DateTimeInput, options: DateTimeDisplayOptions, includeTime: boolean): string {
  const fallback = options.fallback ?? "-";
  const date = parseDateInput(value);
  if (!date) return typeof value === "string" && value ? value : fallback;

  const language = options.language ?? DEFAULT_DATE_LANGUAGE;
  const yearType = options.yearType ?? DEFAULT_CALENDAR_YEAR_TYPE;
  const timeZone = options.timeZone?.trim() || DEFAULT_TIME_ZONE;
  const parts = formatParts(date, language, yearType, timeZone, includeTime, options.includeSeconds ?? includeTime)
    ?? formatParts(date, language, yearType, DEFAULT_TIME_ZONE, includeTime, options.includeSeconds ?? includeTime);
  if (!parts) return fallback;

  const day = partValue(parts, "day").padStart(2, "0");
  const month = partValue(parts, "month").padStart(2, "0");
  const year = partValue(parts, "year");
  if (!includeTime) return `${day}/${month}/${year}`;

  const hour = normalizeHour(partValue(parts, "hour"));
  const minute = partValue(parts, "minute").padStart(2, "0");
  const second = partValue(parts, "second").padStart(2, "0");
  return `${day}/${month}/${year} ${hour}:${minute}${options.includeSeconds === false ? "" : `:${second}`}`;
}

function parseDateInput(value: DateTimeInput): Date | null {
  if (value === null || value === undefined || value === "") return null;
  const date = value instanceof Date ? value : new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

function formatParts(
  date: Date,
  language: LanguageCode,
  yearType: CalendarYearType,
  timeZone: string,
  includeTime: boolean,
  includeSeconds: boolean,
): Intl.DateTimeFormatPart[] | null {
  try {
    return new Intl.DateTimeFormat(localeForDate(language, yearType), {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      ...(includeTime
        ? {
            hour: "2-digit",
            minute: "2-digit",
            second: includeSeconds ? "2-digit" : undefined,
            hourCycle: "h23",
          }
        : {}),
      timeZone,
    }).formatToParts(date);
  } catch {
    return null;
  }
}

function partValue(parts: Intl.DateTimeFormatPart[], type: Intl.DateTimeFormatPartTypes): string {
  return parts.find((part) => part.type === type)?.value ?? "";
}

function normalizeHour(hour: string): string {
  return (hour === "24" ? "00" : hour).padStart(2, "0");
}

function booleanLikeValue(value: unknown): CalendarYearType {
  if (typeof value === "boolean") return value ? "buddhist" : "christian";
  const raw = stringValue(value).toLowerCase();
  return raw === "false" || raw === "0" || raw === "christian" || raw === "ce" || raw === "ค.ศ."
    ? "christian"
    : "buddhist";
}

function stringValue(value: unknown): string {
  return typeof value === "string" ? value.trim() : value === null || value === undefined ? "" : String(value).trim();
}

function getByPath(record: Record<string, unknown>, path: string): unknown {
  return path.split(".").reduce<unknown>((current, key) => (isRecord(current) ? current[key] : undefined), record);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
