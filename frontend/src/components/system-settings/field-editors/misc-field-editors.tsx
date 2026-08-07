"use client";

import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { Check, ChevronsUpDown, Plus, Search, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { AuthSession } from "@/lib/workspace-models";
import type { LanguageCode } from "@/lib/i18n";
import type {
  SystemSettingField,
  SystemSettingOption,
} from "@/lib/system-setting-screens";
import { MasterPicker } from "@/components/product-barcode/master-picker";
import type { MasterEntry, MasterName } from "@/lib/product-barcode/api";
import {
  type FormState,
  type SettingRecord,
  getLocalizedNameArray,
  isRecord,
  localeOf,
  localizedNameForLanguage,
  stringValue,
  uniqueStrings,
} from "../types";
import { normalizeStringListValue } from "./permission-editors";
import { optionLabel } from "../utils";

/* ------------------------------------------------------------------ */
/* StringListFieldEditor                                               */
/* ------------------------------------------------------------------ */

export function StringListFieldEditor({
  field,
  form,
  label,
  language,
  setForm,
}: {
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const values = normalizeStringListValue(form[field.key]);
  const [draft, setDraft] = useState("");
  const helper =
    field.helper?.[language] ?? field.helper?.en ?? field.helper?.th;
  const addLabel = language === "th" ? "เพิ่ม" : "Add";
  const emptyLabel =
    language === "th" ? "ยังไม่มีชื่อเรียกอื่น" : "No aliases yet";
  const inputPlaceholder =
    field.placeholder ??
    (language === "th"
      ? "พิมพ์แล้วกด Enter"
      : "Type and press Enter");

  const commit = (nextValues: string[]) => {
    setForm({
      ...form,
      [field.key]: uniqueStrings(
        nextValues.map((item) => item.trim()).filter(Boolean),
      ),
    });
  };

  const addDraft = () => {
    const next = draft.trim();
    if (!next) return;
    commit([...values, next]);
    setDraft("");
  };

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-3 md:col-span-2">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="text-sm font-semibold">
            {label}
            {field.required ? " *" : ""}
          </div>
          {helper ? (
            <p className="mt-0.5 text-xs font-medium text-muted-foreground">
              {helper}
            </p>
          ) : null}
        </div>
        <Badge variant="outline" className="shrink-0 text-xs">
          {values.length}
        </Badge>
      </div>
      <div className="flex flex-col gap-2 sm:flex-row">
        <Input
          className="min-h-10 flex-1"
          value={draft}
          placeholder={inputPlaceholder}
          onChange={(event) => setDraft(event.target.value)}
          onKeyDown={(event) => {
            if (event.key !== "Enter") return;
            event.preventDefault();
            addDraft();
          }}
        />
        <Button type="button" variant="outline" onClick={addDraft}>
          <Plus className="size-4" />
          {addLabel}
        </Button>
      </div>
      {values.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-xs font-medium text-muted-foreground">
          {emptyLabel}
        </div>
      ) : (
        <div className="flex flex-wrap gap-2">
          {values.map((item) => (
            <span
              key={item}
              className="inline-flex max-w-full items-center gap-1 rounded-full border border-border bg-muted/40 px-3 py-1 text-xs font-semibold"
            >
              <span className="break-words">{item}</span>
              <button
                type="button"
                className="rounded-full p-0.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                aria-label={
                  language === "th" ? `ลบ ${item}` : `Remove ${item}`
                }
                onClick={() =>
                  commit(values.filter((value) => value !== item))
                }
              >
                <X className="size-3.5" />
              </button>
            </span>
          ))}
        </div>
      )}
    </section>
  );
}

/* ------------------------------------------------------------------ */
/* MasterPickerFieldEditor                                             */
/* ------------------------------------------------------------------ */

export function MasterPickerFieldEditor({
  auth,
  field,
  form,
  label,
  language,
  setForm,
}: {
  auth: AuthSession | null;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const [open, setOpen] = useState(false);
  const anchorRef = useRef<HTMLButtonElement | null>(null);
  const rawValue = form[field.key];
  const value: SettingRecord = isRecord(rawValue) ? rawValue : {};
  const names = getLocalizedNameArray(value.names);
  const displayName =
    localizedNameForLanguage(names, language) || stringValue(value.name);
  const code = stringValue(value.code);
  const master = (field.master ?? "businesstype") as MasterName;

  function select(entry: MasterEntry) {
    setForm({
      ...form,
      [field.key]: {
        guidfixed: entry.guidfixed,
        code: entry.code,
        names: entry.names,
      },
    });
    setOpen(false);
  }

  return (
    <section className="grid gap-1 text-sm font-semibold md:col-span-2">
      <span>
        {label}
        {field.required ? " *" : ""}
      </span>
      <div className="flex min-w-0 gap-2">
        <button
          ref={anchorRef}
          type="button"
          className="flex min-h-10 min-w-0 flex-1 items-center justify-between gap-2 rounded-2xl border border-input bg-background px-3 text-left text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          onClick={() => setOpen(true)}
        >
          <span
            className={cn(
              "min-w-0 truncate",
              !code && !displayName && "text-muted-foreground",
            )}
          >
            {code || displayName
              ? `${code}${code && displayName ? " - " : ""}${displayName}`
              : language === "th"
                ? "เลือกข้อมูล"
                : "Select"}
          </span>
          <Search className="size-4 shrink-0 text-muted-foreground" />
        </button>
        {code || displayName ? (
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={() => setForm({ ...form, [field.key]: {} })}
            aria-label={language === "th" ? "ล้างค่า" : "Clear"}
          >
            <X />
          </Button>
        ) : null}
      </div>
      <MasterPicker
        open={open}
        onClose={() => setOpen(false)}
        auth={auth}
        language={language}
        master={master}
        placement="field"
        anchorRef={anchorRef}
        title={label}
        onSelect={select}
      />
    </section>
  );
}

/* ------------------------------------------------------------------ */
/* MasterMultiPickerFieldEditor                                        */
/* ------------------------------------------------------------------ */

export function masterMultiPickerEntries(value: unknown): SettingRecord[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord);
}

export function normalizeMasterMultiPickerValue(value: unknown): SettingRecord[] {
  return masterMultiPickerEntries(value).map((item) => ({
    guidfixed: stringValue(item.guidfixed),
    code: stringValue(item.code ?? item.groupcode),
    names: item.names,
  }));
}

export function MasterMultiPickerFieldEditor({
  auth,
  field,
  form,
  label,
  language,
  setForm,
}: {
  auth: AuthSession | null;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const [open, setOpen] = useState(false);
  const anchorRef = useRef<HTMLButtonElement | null>(null);
  const entries = masterMultiPickerEntries(form[field.key]);
  const master = (field.master ?? "businesstype") as MasterName;

  function addEntry(entry: MasterEntry) {
    if (entries.some((item) => stringValue(item.guidfixed) === entry.guidfixed)) {
      setOpen(false);
      return;
    }
    setForm({
      ...form,
      [field.key]: [
        ...entries,
        { guidfixed: entry.guidfixed, code: entry.code, names: entry.names },
      ],
    });
    setOpen(false);
  }

  function removeEntry(guidfixed: string) {
    setForm({
      ...form,
      [field.key]: entries.filter((item) => stringValue(item.guidfixed) !== guidfixed),
    });
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-3 md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="text-sm font-semibold">
          {label}
          {field.required ? " *" : ""}
        </span>
        <Button
          ref={anchorRef}
          type="button"
          variant="outline"
          size="sm"
          onClick={() => setOpen(true)}
        >
          <Plus className="size-4" />
          {language === "th" ? "เพิ่ม" : "Add"}
        </Button>
      </div>
      {entries.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-xs font-medium text-muted-foreground">
          {language === "th" ? "ยังไม่ได้เลือกกลุ่ม" : "No groups selected"}
        </div>
      ) : (
        <div className="flex flex-wrap gap-2">
          {entries.map((item) => {
            const guid = stringValue(item.guidfixed);
            const names = getLocalizedNameArray(item.names);
            const displayName =
              localizedNameForLanguage(names, language) || stringValue(item.code);
            return (
              <span
                key={guid}
                className="inline-flex max-w-full items-center gap-1 rounded-full border border-border bg-muted/40 px-3 py-1 text-xs font-semibold"
              >
                <span className="break-words">
                  {stringValue(item.code)}
                  {stringValue(item.code) && displayName ? " - " : ""}
                  {displayName}
                </span>
                <button
                  type="button"
                  className="rounded-full p-0.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                  aria-label={language === "th" ? "ลบ" : "Remove"}
                  onClick={() => removeEntry(guid)}
                >
                  <X className="size-3.5" />
                </button>
              </span>
            );
          })}
        </div>
      )}
      <MasterPicker
        open={open}
        onClose={() => setOpen(false)}
        auth={auth}
        language={language}
        master={master}
        placement="field"
        anchorRef={anchorRef}
        title={label}
        onSelect={addEntry}
      />
    </section>
  );
}

/* ------------------------------------------------------------------ */
/* ComboFieldEditor + helpers                                          */
/* ------------------------------------------------------------------ */

export type ComboOption = SystemSettingOption;
type ComboPlacement = {
  left: number;
  top: number;
  width: number;
  maxHeight: number;
};

const COMBO_VIEWPORT_MARGIN = 12;
const COMBO_GAP = 4;
const COMBO_MAX_HEIGHT = 360;
const COMBO_MIN_USABLE_HEIGHT = 80;


export function comboOptionsForField(
  field: SystemSettingField,
  language: LanguageCode,
  currentValue: string,
): ComboOption[] {
  const sourceOptions =
    field.optionSource === "timezones"
      ? timezoneOptions(language)
      : (field.options ?? []);
  if (
    !currentValue ||
    sourceOptions.some((option) => option.value === currentValue)
  )
    return sourceOptions;
  return [{ value: currentValue, label: currentValue }, ...sourceOptions];
}

function timezoneOptions(language: LanguageCode): ComboOption[] {
  return supportedTimeZones()
    .map((timeZone) => {
      const meta = timezoneMeta(timeZone);
      return {
        value: timeZone,
        label: `${meta.offset ? `UTC${meta.offset} ` : ""}${timeZone}`,
      };
    })
    .sort((first, second) =>
      first.label.localeCompare(second.label, localeOf(language)),
    );
}

function supportedTimeZones(): string[] {
  const intl = Intl as typeof Intl & {
    supportedValuesOf?: (key: "timeZone") => string[];
  };
  if (typeof intl.supportedValuesOf !== "function") return [];
  try {
    return intl.supportedValuesOf("timeZone");
  } catch {
    return [];
  }
}

export function timezoneMeta(timeZone: string): { label: string; offset: string } {
  const offset = timezoneUtcOffset(timeZone);
  return {
    label: offset ? `(UTC${offset}) ${timeZone}` : timeZone,
    offset,
  };
}

function timezoneUtcOffset(timeZone: string): string {
  try {
    const parts = new Intl.DateTimeFormat("en-US", {
      hour: "2-digit",
      minute: "2-digit",
      timeZone,
      timeZoneName: "longOffset",
    }).formatToParts(new Date());
    const zone =
      parts.find((part) => part.type === "timeZoneName")?.value ?? "";
    const rawOffset = zone.replace(/^GMT/i, "").trim();
    return rawOffset ? normalizeUtcOffset(rawOffset) : "+00:00";
  } catch {
    return "";
  }
}

function normalizeUtcOffset(value: string): string {
  const raw = value.trim().replace(/^UTC/i, "");
  const match = raw.match(/^([+-])(\d{1,2}):?(\d{2})$/);
  if (!match) return raw;
  return `${match[1]}${match[2].padStart(2, "0")}:${match[3]}`;
}

export function ComboFieldEditor({
  field,
  form,
  label,
  language,
  setForm,
  applyCountryDefaults,
}: {
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
  applyCountryDefaults?: (form: FormState, fieldKey: string, countryCode: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const buttonRef = useRef<HTMLButtonElement | null>(null);
  const panelRef = useRef<HTMLDivElement | null>(null);
  const [placement, setPlacement] = useState<ComboPlacement | null>(null);
  const [query, setQuery] = useState("");
  const value = String(form[field.key] ?? "");
  const options = useMemo(
    () => comboOptionsForField(field, language, value),
    [field, language, value],
  );
  const currentOption = options.find((option) => option.value === value);
  const needle = query.trim().toLowerCase();
  const visibleOptions = needle
    ? options.filter((option) =>
        `${option.value} ${optionLabel(option, language)}`
          .toLowerCase()
          .includes(needle),
      )
    : options;

  const updatePlacement = useCallback(() => {
    if (!open || typeof window === "undefined") return;
    const rect = buttonRef.current?.getBoundingClientRect();
    if (!rect) return;

    const margin = COMBO_VIEWPORT_MARGIN;
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    const availableWidth = Math.max(120, viewportWidth - margin * 2);
    const width = Math.min(
      Math.max(rect.width, Math.min(240, availableWidth)),
      availableWidth,
    );
    const maxLeft = Math.max(margin, viewportWidth - margin - width);
    const left = Math.min(Math.max(rect.left, margin), maxLeft);
    const below = Math.max(
      0,
      viewportHeight - rect.bottom - COMBO_GAP - margin,
    );
    const above = Math.max(0, rect.top - COMBO_GAP - margin);
    const bestSpace = Math.max(below, above);

    let top: number;
    let maxHeight: number;
    if (bestSpace < COMBO_MIN_USABLE_HEIGHT) {
      top = margin;
      maxHeight = Math.max(
        COMBO_MIN_USABLE_HEIGHT,
        viewportHeight - margin * 2,
      );
    } else if (below >= COMBO_MIN_USABLE_HEIGHT || below >= above) {
      maxHeight = Math.min(COMBO_MAX_HEIGHT, below);
      top = rect.bottom + COMBO_GAP;
    } else {
      maxHeight = Math.min(COMBO_MAX_HEIGHT, above);
      top = Math.max(margin, rect.top - COMBO_GAP - maxHeight);
    }

    setPlacement((current) => {
      if (
        current &&
        current.left === left &&
        current.top === top &&
        current.width === width &&
        current.maxHeight === maxHeight
      ) {
        return current;
      }
      return { left, top, width, maxHeight };
    });
  }, [open]);

  useLayoutEffect(() => {
    if (!open) {
      setPlacement(null);
      return;
    }

    updatePlacement();
    let frameId = 0;
    const scheduleUpdate = () => {
      window.cancelAnimationFrame(frameId);
      frameId = window.requestAnimationFrame(updatePlacement);
    };
    const resizeObserver =
      typeof ResizeObserver !== "undefined" && buttonRef.current
        ? new ResizeObserver(scheduleUpdate)
        : null;

    window.addEventListener("resize", scheduleUpdate);
    window.addEventListener("scroll", scheduleUpdate, true);
    resizeObserver?.observe(buttonRef.current as Element);

    return () => {
      window.cancelAnimationFrame(frameId);
      window.removeEventListener("resize", scheduleUpdate);
      window.removeEventListener("scroll", scheduleUpdate, true);
      resizeObserver?.disconnect();
    };
  }, [open, updatePlacement]);

  useEffect(() => {
    if (!open) return;
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const handler = (event: PointerEvent) => {
      const target = event.target;
      if (!(target instanceof Node)) return;
      if (panelRef.current?.contains(target)) return;
      if (buttonRef.current?.contains(target)) return;
      setOpen(false);
    };
    window.addEventListener("pointerdown", handler, true);
    return () => window.removeEventListener("pointerdown", handler, true);
  }, [open]);

  function choose(option: ComboOption) {
    const nextForm = { ...form, [field.key]: option.value };
    if (field.optionSource === "timezones") {
      const meta = timezoneMeta(option.value);
      const prefix = field.key.includes(".")
        ? `${field.key.split(".").slice(0, -1).join(".")}.`
        : "";
      nextForm[`${prefix}timezonelabel`] = meta.label;
      nextForm[`${prefix}timezoneoffset`] = meta.offset;
    }
    if (field.optionSource === "countries" && applyCountryDefaults)
      applyCountryDefaults(nextForm, field.key, option.value);
    setForm(nextForm);
    setQuery("");
    setOpen(false);
  }

  return (
    <div className="grid gap-1 text-sm font-semibold">
      <span>
        {label}
        {field.required ? " *" : ""}
      </span>
      <button
        ref={buttonRef}
        aria-expanded={open}
        className="flex min-h-10 w-full items-center justify-between gap-2 rounded-2xl border border-input bg-background px-3 text-left text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onClick={() => setOpen((current) => !current)}
        type="button"
      >
        <span
          className={cn("min-w-0 truncate", !value && "text-muted-foreground")}
        >
          {currentOption
            ? optionLabel(currentOption, language)
            : value || field.placeholder || ""}
        </span>
        <ChevronsUpDown className="size-4 shrink-0 text-muted-foreground" />
      </button>
      {open ? (
        <div
          ref={panelRef}
          className="fixed z-50 grid grid-rows-[auto_minmax(0,1fr)] gap-1 overflow-hidden rounded-2xl border border-border bg-popover p-2 text-popover-foreground shadow-lg"
          style={{
            left: placement?.left ?? COMBO_VIEWPORT_MARGIN,
            top: placement?.top ?? COMBO_VIEWPORT_MARGIN,
            width: placement?.width ?? 240,
            maxHeight: placement?.maxHeight,
            visibility: placement ? "visible" : "hidden",
          }}
        >
          <Input
            autoFocus
            className="h-9 min-h-9 rounded-xl px-3 text-sm font-normal"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={language === "th" ? "ค้นหา" : "Search"}
          />
          <div className="grid min-h-0 gap-1 overflow-y-auto">
            {visibleOptions.length ? (
              visibleOptions.map((option) => (
                <button
                  className={cn(
                    "flex min-h-8 w-full items-center justify-between gap-2 rounded-xl px-2 py-1 text-left text-sm font-medium hover:bg-accent hover:text-accent-foreground",
                    option.value === value && "bg-primary/10 text-primary",
                  )}
                  key={option.value}
                  onClick={() => choose(option)}
                  type="button"
                >
                  <span className="min-w-0 truncate">
                    {optionLabel(option, language)}
                  </span>
                  {option.value === value ? (
                    <Check className="size-4 shrink-0" />
                  ) : null}
                </button>
              ))
            ) : (
              <div className="rounded-xl px-2 py-3 text-sm text-muted-foreground">
                {language === "th" ? "ไม่พบข้อมูล" : "No options found"}
              </div>
            )}
          </div>
        </div>
      ) : null}
    </div>
  );
}
