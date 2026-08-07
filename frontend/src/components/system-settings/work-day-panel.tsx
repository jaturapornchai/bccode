"use client";

// Work Day Panel - extracted from system-settings-screen.tsx
// Manages weekly work schedule with time ranges per day

import { Copy, Loader2, Plus, RefreshCcw, Save, Trash2 } from "lucide-react";
import type { LanguageCode } from "@/lib/i18n";
import type { WorkspaceSession } from "@/lib/workspace-models";
import {
  localTimeToUtcTime,
  normalizeTimeInput,
  timeToMinutes,
  type CalendarYearType,
} from "@/lib/date-time";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { TimeField } from "@/components/ui/date-time-field";
import { cn } from "@/lib/utils";
import {
  type SettingRecord,
  type DateTimeScope,
  dateTimeScopePayload,
  stringValue,
  getByPath,
  booleanLikeValue,
  isRecord,
} from "@/components/system-settings/types";
import { uiEn, type UiTextKey } from "@/components/system-settings/utils";

// ─── Types ───────────────────────────────────────────────────────────────────

export type WorkDayTime = {
  starttime?: string;
  endtime?: string;
  start?: string;
  end?: string;
  starttimeutc?: string;
  endtimeutc?: string;
  startdayoffsetutc?: number;
  enddayoffsetutc?: number;
  timezone?: string;
  timezoneoffset?: string;
};

export type WorkDay = {
  code: string;
  name: string;
  isactive: boolean;
  fullday: boolean;
  worktimes: WorkDayTime[];
};

// ─── Day Names ───────────────────────────────────────────────────────────────

export const dayNames: Record<LanguageCode, string[]> = {
  th: ["จันทร์", "อังคาร", "พุธ", "พฤหัสบดี", "ศุกร์", "เสาร์", "อาทิตย์"],
  en: [
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
    "Sunday",
  ],
  cn: ["星期一", "星期二", "星期三", "星期四", "星期五", "星期六", "星期日"],
  ja: ["月曜", "火曜", "水曜", "木曜", "金曜", "土曜", "日曜"],
  ko: ["월요일", "화요일", "수요일", "목요일", "금요일", "토요일", "일요일"],
  lo: ["ຈັນ", "ອັງຄານ", "ພຸດ", "ພະຫັດ", "ສຸກ", "ເສົາ", "ອາທິດ"],
  my: [
    "တနင်္လာ",
    "အင်္ဂါ",
    "ဗုဒ္ဓဟူး",
    "ကြာသပတေး",
    "သောကြာ",
    "စနေ",
    "တနင်္ဂနွေ",
  ],
  km: ["ចន្ទ", "អង្គារ", "ពុធ", "ព្រហស្បតិ៍", "សុក្រ", "សៅរ៍", "អាទិត្យ"],
  vi: [
    "Thứ hai",
    "Thứ ba",
    "Thứ tư",
    "Thứ năm",
    "Thứ sáu",
    "Thứ bảy",
    "Chủ nhật",
  ],
  ms: ["Isnin", "Selasa", "Rabu", "Khamis", "Jumaat", "Sabtu", "Ahad"],
  id: ["Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"],
  fil: [
    "Lunes",
    "Martes",
    "Miyerkules",
    "Huwebes",
    "Biyernes",
    "Sabado",
    "Linggo",
  ],
};

// ─── Helper Functions ────────────────────────────────────────────────────────

export function normalizeUtcOffset(value: string): string {
  const raw = value.trim().replace(/^UTC/i, "");
  const match = raw.match(/^([+-])(\d{1,2}):?(\d{2})$/);
  if (!match) return raw;
  return `${match[1]}${match[2].padStart(2, "0")}:${match[3]}`;
}

export function resolveDateTimeScope(
  workspace: WorkspaceSession | null,
): DateTimeScope {
  const branch = workspace?.branch ?? null;
  const shopInfo = workspace?.shopInfo ?? null;
  const branchcode = stringValue(branch?.code ?? workspace?.shop.branchcode);
  const branchguid = stringValue(branch?.guidfixed);
  const shopTimezone = stringValue(getByPath(shopInfo, "settings.timezone"));
  const shopTimezoneLabel = stringValue(
    getByPath(shopInfo, "settings.timezonelabel"),
  );
  const shopTimezoneOffset = stringValue(
    getByPath(shopInfo, "settings.timezoneoffset"),
  );
  const branchYear = stringValue(branch?.yeartype).toLowerCase();
  const useBuddhistCalendar = booleanLikeValue(
    getByPath(shopInfo, "settings.usebuddhistcalendar"),
  );
  const calendarYearType: CalendarYearType =
    branchYear === "buddhist" || branchYear === "be" || branchYear === "พ.ศ."
      ? "buddhist"
      : branchYear === "christian" ||
          branchYear === "ce" ||
          branchYear === "ค.ศ."
        ? "christian"
        : useBuddhistCalendar
          ? "buddhist"
          : "christian";
  return {
    key: branchguid || branchcode || "company",
    branchcode,
    branchguid,
    timezone: stringValue(branch?.timezone) || shopTimezone,
    timezonelabel:
      stringValue(branch?.timezonelabel) ||
      shopTimezoneLabel ||
      stringValue(branch?.timezone) ||
      shopTimezone,
    timezoneoffset: normalizeUtcOffset(
      stringValue(branch?.timezoneoffset) || shopTimezoneOffset,
    ),
    calendarYearType,
  };
}

function toLegacyTime(value: string): string {
  const normalized = normalizeTimeInput(value);
  return normalized ? normalized.replace(":", "") : "";
}

export function workTimeStart(time: WorkDayTime): string {
  return normalizeTimeInput(time.starttime ?? time.start);
}

export function workTimeEnd(time: WorkDayTime): string {
  return normalizeTimeInput(time.endtime ?? time.end);
}

export function applyUtcFields(time: WorkDayTime, scope?: DateTimeScope): WorkDayTime {
  const next: WorkDayTime = { ...time };
  const starttime = workTimeStart(next);
  const endtime = workTimeEnd(next);
  next.starttime = starttime;
  next.endtime = endtime;
  next.start = toLegacyTime(starttime);
  next.end = toLegacyTime(endtime);
  if (scope) {
    next.timezone = scope.timezone;
    next.timezoneoffset = scope.timezoneoffset;
  }
  const startUtc = scope?.timezoneoffset
    ? localTimeToUtcTime(starttime, scope.timezoneoffset)
    : null;
  const endUtc = scope?.timezoneoffset
    ? localTimeToUtcTime(endtime, scope.timezoneoffset)
    : null;
  if (startUtc) {
    next.starttimeutc = startUtc.time;
    next.startdayoffsetutc = startUtc.dayOffset;
  } else {
    delete next.starttimeutc;
    delete next.startdayoffsetutc;
  }
  if (endUtc) {
    next.endtimeutc = endUtc.time;
    next.enddayoffsetutc = endUtc.dayOffset;
  } else {
    delete next.endtimeutc;
    delete next.enddayoffsetutc;
  }
  return next;
}

export function createWorkTime(
  starttime = "",
  endtime = "",
  scope?: DateTimeScope,
): WorkDayTime {
  const start = normalizeTimeInput(starttime);
  const end = normalizeTimeInput(endtime);
  return applyUtcFields(
    {
      starttime: start,
      endtime: end,
      start: toLegacyTime(start),
      end: toLegacyTime(end),
    },
    scope,
  );
}

export function withWorkTimePatch(
  time: WorkDayTime,
  key: keyof WorkDayTime,
  value: string,
  scope: DateTimeScope,
): WorkDayTime {
  const normalized =
    key === "starttime" || key === "endtime" || key === "start" || key === "end"
      ? normalizeTimeInput(value)
      : value;
  const next: WorkDayTime = { ...time, [key]: normalized };
  if (key === "starttime" || key === "start") {
    next.starttime = normalized;
    next.start = toLegacyTime(normalized);
  }
  if (key === "endtime" || key === "end") {
    next.endtime = normalized;
    next.end = toLegacyTime(normalized);
  }
  return applyUtcFields(next, scope);
}

export function cloneWorkTimes(times: WorkDayTime[]): WorkDayTime[] {
  return times.map((time) => ({ ...time }));
}

export function getWorkTimeIssue(
  day: WorkDay,
  time: WorkDayTime,
  timeIndex: number,
): "timeInvalid" | "timeOverlap" | null {
  if (!day.isactive || day.fullday) return null;
  const start = timeToMinutes(workTimeStart(time));
  const end = timeToMinutes(workTimeEnd(time));
  if (!Number.isFinite(start) || !Number.isFinite(end) || start >= end)
    return "timeInvalid";
  const overlaps = day.worktimes.some((other, otherIndex) => {
    if (otherIndex === timeIndex) return false;
    const otherStart = timeToMinutes(workTimeStart(other));
    const otherEnd = timeToMinutes(workTimeEnd(other));
    if (
      !Number.isFinite(otherStart) ||
      !Number.isFinite(otherEnd) ||
      otherStart >= otherEnd
    )
      return false;
    return start < otherEnd && end > otherStart;
  });
  return overlaps ? "timeOverlap" : null;
}

export function formatUtcPreview(
  time: string | undefined,
  dayOffset: number | undefined,
): string {
  if (!time) return "";
  const suffix = dayOffset ? ` ${dayOffset > 0 ? "+" : ""}${dayOffset}d` : "";
  return `UTC+0 ${time}${suffix}`;
}

export function buildBranchScopedWorkDayBody(
  record: SettingRecord | undefined,
  workspace: WorkspaceSession,
  workDays: WorkDay[],
): SettingRecord {
  const scope = resolveDateTimeScope(workspace);
  const base = record ?? {};
  const branches = isRecord(base.branches) ? { ...base.branches } : {};
  const scopedWorkDays = workDays.map((day) => ({
    ...day,
    worktimes: day.worktimes.map((time) => applyUtcFields(time, scope)),
  }));
  const branchPayload = {
    ...(isRecord(branches[scope.key])
      ? (branches[scope.key] as SettingRecord)
      : {}),
    ...dateTimeScopePayload(scope),
    workdays: scopedWorkDays,
  };
  branches[scope.key] = branchPayload;
  return {
    ...base,
    ...dateTimeScopePayload(scope),
    workdays: scopedWorkDays,
    branches,
  };
}

// ─── WorkDayPanel Component ──────────────────────────────────────────────────

export function WorkDayPanel({
  dateTimeScope,
  language,
  loading,
  onRefresh,
  onSave,
  saving,
  setWorkDays,
  text,
  workDays,
}: {
  dateTimeScope: DateTimeScope;
  language: LanguageCode;
  loading: boolean;
  onRefresh: () => void;
  onSave: () => void;
  saving: boolean;
  setWorkDays: (workDays: WorkDay[]) => void;
  text: (key: UiTextKey) => string;
  workDays: WorkDay[];
}) {
  const activeDays = workDays.filter((day) => day.isactive).length;
  const totalRanges = workDays.reduce(
    (sum, day) =>
      sum +
      (day.isactive
        ? day.fullday
          ? 1
          : Math.max(day.worktimes.length, 1)
        : 0),
    0,
  );
  const invalidCount = workDays.reduce(
    (sum, day) =>
      sum +
      day.worktimes.filter((time, timeIndex) =>
        getWorkTimeIssue(day, time, timeIndex),
      ).length,
    0,
  );

  function updateDay(index: number, patch: Partial<WorkDay>) {
    setWorkDays(
      workDays.map((day, dayIndex) =>
        dayIndex === index ? { ...day, ...patch } : day,
      ),
    );
  }

  function updateTime(
    dayIndex: number,
    timeIndex: number,
    key: keyof WorkDayTime,
    value: string,
  ) {
    const day = workDays[dayIndex];
    const nextTimes = day.worktimes.length
      ? [...day.worktimes]
      : [createWorkTime("", "", dateTimeScope)];
    const currentTime =
      nextTimes[timeIndex] ?? createWorkTime("", "", dateTimeScope);
    nextTimes[timeIndex] = withWorkTimePatch(
      currentTime,
      key,
      value,
      dateTimeScope,
    );
    updateDay(dayIndex, { worktimes: nextTimes });
  }

  function addTimeRange(dayIndex: number) {
    const day = workDays[dayIndex];
    updateDay(dayIndex, {
      fullday: false,
      isactive: true,
      worktimes: [
        ...(day.worktimes.length ? day.worktimes : []),
        createWorkTime("", "", dateTimeScope),
      ],
    });
  }

  function removeTimeRange(dayIndex: number, timeIndex: number) {
    const day = workDays[dayIndex];
    const nextTimes = day.worktimes.filter((_, index) => index !== timeIndex);
    updateDay(dayIndex, {
      worktimes: nextTimes.length
        ? nextTimes
        : [createWorkTime("", "", dateTimeScope)],
    });
  }

  function copyMondaySchedule() {
    const monday = workDays[0];
    if (!monday) return;
    setWorkDays(
      workDays.map((day, index) =>
        index === 0 || !day.isactive
          ? day
          : {
              ...day,
              fullday: monday.fullday,
              worktimes: cloneWorkTimes(monday.worktimes),
            },
      ),
    );
  }

  function toggleFullDay(index: number, checked: boolean) {
    const day = workDays[index];
    updateDay(index, {
      fullday: checked,
      worktimes: day.worktimes.length
        ? day.worktimes
        : [createWorkTime("", "", dateTimeScope)],
    });
  }

  return (
    <Card className="w-full overflow-hidden">
      <CardHeader className="p-3 pb-1">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div className="min-w-0">
            <CardTitle>{text("workTime")}</CardTitle>
            <CardDescription className="truncate">
              {text("updated")}: restaurant settings code workDay
            </CardDescription>
          </div>
          <div className="flex w-full flex-wrap gap-2 sm:w-auto">
            <Badge
              variant={invalidCount ? "warning" : "success"}
              className="min-h-8 justify-center px-3"
            >
              {text("active")}: {activeDays}
            </Badge>
            <Badge variant="outline" className="min-h-8 justify-center px-3">
              {text("totalRanges")}: {totalRanges}
            </Badge>
            <Button
              className="w-full sm:w-auto"
              size="sm"
              type="button"
              variant="outline"
              onClick={copyMondaySchedule}
              disabled={loading || !workDays[0]}
            >
              <Copy />
              {text("copyMondaySchedule")}
            </Button>
            <Button
              className="w-full sm:w-auto"
              size="sm"
              type="button"
              variant="outline"
              onClick={onRefresh}
              disabled={loading}
            >
              {loading ? <Loader2 className="animate-spin" /> : <RefreshCcw />}
              {text("refresh")}
            </Button>
            <Button
              className="w-full sm:w-auto"
              size="sm"
              type="button"
              onClick={onSave}
              disabled={saving || invalidCount > 0}
            >
              {saving ? <Loader2 className="animate-spin" /> : <Save />}
              {text("save")}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="grid gap-2 p-3">
        {workDays.length ? (
          workDays.map((day, index) => {
            const times = day.worktimes.length
              ? day.worktimes
              : [createWorkTime("", "", dateTimeScope)];
            return (
              <div
                className={cn(
                  "grid gap-2 rounded-2xl border border-border bg-background p-2",
                  !day.isactive && "bg-muted/30",
                )}
                key={day.code}
              >
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <label className="flex min-w-40 flex-1 items-center gap-2 text-sm font-semibold">
                    <input
                      className="size-4 accent-primary"
                      type="checkbox"
                      checked={day.isactive}
                      onChange={(event) =>
                        updateDay(index, { isactive: event.target.checked })
                      }
                    />
                    <span className="truncate">
                      {dayNames[language]?.[index] ?? day.name}
                    </span>
                    <Badge variant={day.isactive ? "success" : "outline"}>
                      {day.isactive ? text("active") : text("off")}
                    </Badge>
                  </label>
                  {day.isactive ? (
                    <div className="flex w-full flex-wrap items-center gap-2 sm:w-auto">
                      <label className="flex h-8 items-center gap-2 rounded-xl border border-border bg-card px-2 text-xs font-semibold">
                        <input
                          className="size-4 accent-primary"
                          type="checkbox"
                          checked={day.fullday}
                          onChange={(event) =>
                            toggleFullDay(index, event.target.checked)
                          }
                        />
                        {text("fullDay")}
                      </label>
                      {!day.fullday ? (
                        <Button
                          size="sm"
                          type="button"
                          variant="outline"
                          onClick={() => addTimeRange(index)}
                        >
                          <Plus />
                          {text("add")} {text("range")}
                        </Button>
                      ) : null}
                    </div>
                  ) : null}
                </div>
                {day.isactive && !day.fullday ? (
                  <div className="grid gap-2 md:grid-cols-2 xl:grid-cols-3">
                    {times.map((time, timeIndex) => {
                      const issue = getWorkTimeIssue(day, time, timeIndex);
                      return (
                        <div
                          className={cn(
                            "grid gap-1 rounded-xl border border-border bg-card p-2",
                            issue &&
                              "border-amber-300 bg-amber-50/70 dark:border-amber-900 dark:bg-amber-950/30",
                          )}
                          key={`${day.code}-${timeIndex}`}
                        >
                          <div className="flex items-center justify-between gap-2">
                            <b className="text-xs">
                              {text("range")} {timeIndex + 1}
                            </b>
                            <Button
                              aria-label={`${text("removeRange")} ${timeIndex + 1}`}
                              disabled={times.length <= 1}
                              onClick={() => removeTimeRange(index, timeIndex)}
                              size="icon"
                              type="button"
                              variant="ghost"
                              className="h-7 w-7 rounded-xl"
                            >
                              <Trash2 />
                            </Button>
                          </div>
                          <div className="grid grid-cols-2 gap-2">
                            <TimeField
                              aria-label={`${dayNames[language]?.[index] ?? day.name} ${text("range")} ${timeIndex + 1} ${text("workTime")} start`}
                              label={text("startTime")}
                              timezoneLabel={dateTimeScope.timezoneoffset}
                              utcPreview={formatUtcPreview(
                                time.starttimeutc,
                                time.startdayoffsetutc,
                              )}
                              value={workTimeStart(time)}
                              onChange={(event) =>
                                updateTime(
                                  index,
                                  timeIndex,
                                  "starttime",
                                  event.target.value,
                                )
                              }
                            />
                            <TimeField
                              aria-label={`${dayNames[language]?.[index] ?? day.name} ${text("range")} ${timeIndex + 1} ${text("workTime")} end`}
                              label={text("endTime")}
                              timezoneLabel={dateTimeScope.timezoneoffset}
                              utcPreview={formatUtcPreview(
                                time.endtimeutc,
                                time.enddayoffsetutc,
                              )}
                              value={workTimeEnd(time)}
                              onChange={(event) =>
                                updateTime(
                                  index,
                                  timeIndex,
                                  "endtime",
                                  event.target.value,
                                )
                              }
                            />
                          </div>
                          {issue ? (
                            <span className="text-xs font-semibold text-amber-700 dark:text-amber-300">
                              {text(issue)}
                            </span>
                          ) : null}
                        </div>
                      );
                    })}
                    <Button
                      type="button"
                      variant="outline"
                      className="min-h-20 w-full border-dashed"
                      onClick={() => addTimeRange(index)}
                    >
                      <Plus />
                      {text("add")} {text("range")}
                    </Button>
                  </div>
                ) : null}
                {day.isactive && day.fullday ? (
                  <div className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-sm font-semibold text-muted-foreground">
                    {text("fullDay")}
                  </div>
                ) : null}
              </div>
            );
          })
        ) : (
          <div className="rounded-2xl border border-dashed border-border p-4 text-sm text-muted-foreground">
            {text("empty")}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
