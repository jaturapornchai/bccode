import type { LanguageCode } from "@/lib/i18n";

export type CalendarYearType = "buddhist" | "christian";

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
  return yearType === "buddhist" ? `${locale}-u-ca-buddhist` : locale;
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
