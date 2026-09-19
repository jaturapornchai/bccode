"use client";

import { Check, ChevronsUpDown, Plus, Trash2, X } from "lucide-react";
import Image from "next/image";
import { type DragEvent as ReactDragEvent, useEffect, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { SystemSettingConfig, SystemSettingField } from "@/lib/system-setting-screens";
import { LANGUAGES, SYSTEM_LANGUAGES, normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import type { WorkspaceSession } from "@/lib/workspace-models";
import { cn } from "@/lib/utils";
import {
  type FormState,
  type SettingRecord,
  booleanLikeValue,
  isRecord,
  moveArrayItem,
  safeJsonParse,
  stringValue,
} from "../types";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type LanguageConfigFormRow = {
  code: LanguageCode;
  codetranslator: string;
  name: string;
  isuse: boolean;
  isdefault: boolean;
};

type NormalizeLanguageOptions = {
  forcePrimaryFirst?: boolean;
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

export function supportedLanguageCode(
  value: unknown,
  fallback: LanguageCode,
): LanguageCode;
export function supportedLanguageCode(value: unknown, fallback: ""): LanguageCode | "";
export function supportedLanguageCode(
  value: unknown,
  fallback: LanguageCode | "",
): LanguageCode | "" {
  const raw = stringValue(value).toLowerCase();
  if (!raw) return fallback;
  const normalized = normalizeLanguage(raw);
  return LANGUAGES.some((item) => item.code === normalized)
    ? normalized
    : fallback;
}

export function languageName(code: string, language: LanguageCode): string {
  const item = LANGUAGES.find((entry) => entry.code === code);
  if (!item) return code.toUpperCase();
  if (language === "th") return item.name;
  return `${item.name} (${item.code.toUpperCase()})`;
}

export function languageConfigRow(
  code: LanguageCode,
  isDefault: boolean,
): LanguageConfigFormRow {
  return {
    code,
    codetranslator: code,
    name: languageName(code, "en"),
    isuse: true,
    isdefault: isDefault,
  };
}

export function defaultLanguageConfigs(defaultCode: unknown): LanguageConfigFormRow[] {
  return [languageConfigRow(supportedLanguageCode(defaultCode, "th"), true)];
}

export function normalizeLanguageConfigs(
  value: unknown,
  defaultCode: unknown,
  options: NormalizeLanguageOptions = {},
): LanguageConfigFormRow[] {
  const requestedPrimaryCode = supportedLanguageCode(defaultCode, "");
  const source = typeof value === "string" ? safeJsonParse(value, []) : value;
  const input = Array.isArray(source) ? source : [];
  const seen = new Set<string>();
  const rows: LanguageConfigFormRow[] = [];

  for (const item of input) {
    if (!isRecord(item)) continue;
    const code = supportedLanguageCode(item.code, "");
    if (!code || seen.has(code)) continue;
    const isDefault = booleanLikeValue(item.isdefault);
    const isUse =
      item.isuse === undefined && item.isuse === undefined
        ? true
        : booleanLikeValue(item.isuse ?? item.isuse);
    if (!isUse && !isDefault) continue;
    seen.add(code);
    rows.push({
      code,
      codetranslator:
        stringValue(item.codetranslator ?? item.codeTranslator) || code,
      name: stringValue(item.name) || languageName(code, "en"),
      isuse: true,
      isdefault: isDefault,
    });
  }

  if (!rows.length)
    return [languageConfigRow(requestedPrimaryCode || "th", true)];

  const requestedExists = Boolean(
    requestedPrimaryCode && seen.has(requestedPrimaryCode),
  );
  const primaryCode = supportedLanguageCode(
    options.forcePrimaryFirst && requestedPrimaryCode
      ? requestedPrimaryCode
      : rows.find((row) => row.isdefault)?.code ||
          rows[0]?.code ||
          requestedPrimaryCode ||
          "th",
    "th",
  );
  if (!seen.has(primaryCode))
    rows.unshift(languageConfigRow(primaryCode, true));
  const ordered =
    options.forcePrimaryFirst && requestedExists
      ? [
          rows.find((row) => row.code === primaryCode) ??
            languageConfigRow(primaryCode, true),
          ...rows.filter((row) => row.code !== primaryCode),
        ]
      : rows;
  return ordered.map((row) => ({
    ...row,
    codetranslator: row.codetranslator || row.code,
    isdefault: row.code === primaryCode,
    isuse: true,
    name: row.name || languageName(row.code, "en"),
  }));
}

export function setDefaultLanguageConfig(
  value: unknown,
  defaultCode: unknown,
): LanguageConfigFormRow[] {
  return normalizeLanguageConfigs(value, defaultCode, {
    forcePrimaryFirst: true,
  });
}

export function normalizeLanguageList(
  value: unknown,
  defaultCode: unknown,
  options: NormalizeLanguageOptions = {},
): string[] {
  const requestedPrimaryCode = supportedLanguageCode(defaultCode, "");
  const source =
    typeof value === "string"
      ? value.trim().startsWith("[")
        ? safeJsonParse(value, [])
        : value.split(",")
      : value;
  const input = Array.isArray(source) ? source : [];
  const seen = new Set<string>();
  const rows: string[] = [];

  for (const item of input) {
    const code = supportedLanguageCode(item, "");
    if (!code || seen.has(code)) continue;
    seen.add(code);
    rows.push(code);
  }

  if (!rows.length) return [requestedPrimaryCode || "th"];
  if (!options.forcePrimaryFirst || !requestedPrimaryCode) return rows;
  if (!seen.has(requestedPrimaryCode)) return [requestedPrimaryCode, ...rows];
  return [
    requestedPrimaryCode,
    ...rows.filter((code) => code !== requestedPrimaryCode),
  ];
}

export function nameEditorLanguageCodes(
  form: FormState,
  config: SystemSettingConfig,
  language: LanguageCode,
  workspace?: WorkspaceSession | null,
): string[] {
  if (config.kind === "company") {
    return normalizeLanguageConfigs(
      form["settings.languageconfigs"] ?? form["settings.languageconfigs"],
      form["settings.language"],
    ).map((row) => row.code);
  }
  if (config.slug === "branch") {
    return normalizeLanguageList(form.languages, form.language);
  }
  if (workspace) {
    const shopInfo = isRecord(workspace.shopInfo) ? workspace.shopInfo : {};
    const settings = isRecord(shopInfo.settings) ? shopInfo.settings : {};

    const rawConfigs = settings.languageconfigs ?? settings.languageconfigs ?? shopInfo["settings.languageconfigs"] ?? shopInfo["settings.languageconfigs"] ?? shopInfo.languageconfigs;
    const rawLang = settings.language ?? shopInfo["settings.language"] ?? shopInfo.language;

    if (Array.isArray(rawConfigs)) {
      const activeCodes = rawConfigs
        .filter(isRecord)
        .map((item) => {
          const isUse = item.isuse === undefined && item.isuse === undefined ? true : (item.isuse === true || item.isuse === "true" || item.isuse === true || item.isuse === "true");
          return {
            code: supportedLanguageCode(item.code, ""),
            isUse,
          };
        })
        .filter((item) => item.code && item.isUse)
        .map((item) => item.code);
      if (activeCodes.length > 0) {
        const primary = supportedLanguageCode(rawLang, language);
        return [primary, ...activeCodes.filter((code) => code !== primary)];
      }
    }
  }
  return [language];
}

// ---------------------------------------------------------------------------
// LanguageFlag
// ---------------------------------------------------------------------------

export function LanguageFlag({ code }: { code: string }) {
  const normalized = supportedLanguageCode(code, "th");
  return (
    <span className="grid size-7 shrink-0 place-items-center overflow-hidden rounded border border-border bg-card">
      <Image alt="" className="h-auto w-full object-contain" height={18} src={`/flags/${normalized}.png`} width={27} />
    </span>
  );
}

// ---------------------------------------------------------------------------
// LanguageAddDialog
// ---------------------------------------------------------------------------

export function LanguageAddDialog({
  activeCodes,
  language,
  onClose,
  onToggle,
  open,
}: {
  activeCodes: string[];
  language: LanguageCode;
  onClose: () => void;
  onToggle: (code: LanguageCode) => void;
  open: boolean;
}) {
  useEffect(() => {
    if (!open) return;
    const previousOverflow = document.body.style.overflow;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    document.body.style.overflow = "hidden";
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [onClose, open]);

  if (!open) return null;

  const title = language === "th" ? "เลือกภาษา" : "Select Language";
  const closeText = language === "th" ? "ปิด" : "Close";

  return (
    <div className="dialog-backdrop" onClick={onClose} role="presentation">
      <section
        aria-label={title}
        aria-modal="true"
        className="language-dialog"
        onClick={(event) => event.stopPropagation()}
        role="dialog"
      >
        <div className="dialog-header">
          <div className="min-w-0">
            <p className="eyebrow">{language === "th" ? "ภาษา" : "Language"}</p>
            <h2>{title}</h2>
          </div>
          <Button
            aria-label={closeText}
            type="button"
            variant="outline"
            size="icon"
            onClick={onClose}
          >
            <X aria-hidden="true" />
          </Button>
        </div>

        <div className="language-grid">
          {SYSTEM_LANGUAGES.map((item) => {
            const isPrimary = activeCodes[0] === item.code;
            const isActive = activeCodes.includes(item.code);
            return (
              <button
                className={cn(
                  "language-choice transition-all",
                  isPrimary &&
                    "border-primary/30 bg-primary/10 text-primary cursor-default hover:bg-primary/10 hover:border-primary/30 shadow-none",
                  isActive &&
                    !isPrimary &&
                    "border-primary text-primary bg-background hover:bg-primary/5",
                  !isActive &&
                    "border-border bg-card text-foreground hover:border-primary/50 hover:bg-accent/50",
                )}
                key={item.code}
                onClick={() => {
                  if (isPrimary) return;
                  onToggle(item.code);
                }}
                type="button"
              >
                <LanguageFlag code={item.code} />
                <span className="font-semibold">
                  {languageName(item.code, language)}
                </span>
                {isPrimary && (
                  <Check
                    className="ml-auto size-4 text-primary"
                    aria-hidden="true"
                  />
                )}
              </button>
            );
          })}
        </div>
      </section>
    </div>
  );
}

// ---------------------------------------------------------------------------
// LanguageConfigsEditor
// ---------------------------------------------------------------------------

export function LanguageConfigsEditor({
  form,
  language,
  label,
  setForm,
}: {
  form: FormState;
  language: LanguageCode;
  label: string;
  setForm: (update: FormState | ((current: FormState) => FormState)) => void;
}) {
  const defaultCode = supportedLanguageCode(form["settings.language"], "th");
  const rows = normalizeLanguageConfigs(
    form["settings.languageconfigs"] ?? form["settings.languageconfigs"],
    defaultCode,
  );

  function commit(
    nextRows: LanguageConfigFormRow[],
    nextDefault = nextRows[0]?.code ?? defaultCode,
  ) {
    const normalized = normalizeLanguageConfigs(nextRows, nextDefault, {
      forcePrimaryFirst: true,
    });
    const primary =
      normalized[0]?.code ?? supportedLanguageCode(nextDefault, "th");
    setForm({
      ...form,
      "settings.language": primary,
      "settings.languageconfigs": normalized,
    });
  }

  const primaryCode = rows[0]?.code ?? defaultCode ?? "th";
  const isEnActive = rows.some((row) => row.code === "en");

  function setPrimaryLanguage(code: LanguageCode) {
    if (code === "th") {
      if (isEnActive) {
        commit([languageConfigRow("th", true), languageConfigRow("en", false)], "th");
      } else {
        commit([languageConfigRow("th", true)], "th");
      }
    } else {
      commit([languageConfigRow("en", true), languageConfigRow("th", false)], "en");
    }
  }

  function toggleEnglish() {
    if (isEnActive) {
      commit([languageConfigRow("th", true)], "th");
    } else {
      commit([languageConfigRow(primaryCode, true), languageConfigRow(primaryCode === "en" ? "th" : "en", false)], primaryCode);
    }
  }

  const textPrimary = language === "th" ? "ภาษาหลัก" : "Primary language";
  const textSetPrimary = language === "th" ? "ตั้งเป็นภาษาหลัก" : "Set as primary";

  return (
    <section className="grid gap-2.5 rounded-2xl border border-border bg-background p-3 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="font-bold text-foreground">{label}</div>
          <div className="text-xs text-muted-foreground">
            {language === "th"
              ? "เลือกภาษาที่ต้องการใช้งานในระบบ (ไทย / อังกฤษ)"
              : "Select languages to use in system (Thai / English)"}
          </div>
        </div>
      </div>

      <div className="grid gap-2 sm:grid-cols-2">
        {/* การ์ดภาษาไทย */}
        <div
          className={cn(
            "flex flex-col justify-between gap-3 rounded-xl border p-3 transition-all",
            primaryCode === "th"
              ? "border-primary/50 bg-primary/5 ring-1 ring-primary/20 shadow-xs"
              : "border-border bg-card hover:border-border/80",
          )}
        >
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-3">
              <LanguageFlag code="th" />
              <div>
                <div className="font-bold text-foreground">ภาษาไทย</div>
                <div className="text-xs text-muted-foreground font-mono">TH</div>
              </div>
            </div>
            {primaryCode === "th" ? (
              <Badge variant="default" className="gap-1 bg-primary text-primary-foreground font-bold text-xs py-0.5 px-2">
                <Check className="size-3" />
                {textPrimary}
              </Badge>
            ) : (
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="text-xs h-7 font-bold"
                onClick={() => setPrimaryLanguage("th")}
              >
                {textSetPrimary}
              </Button>
            )}
          </div>
          <div className="text-xs text-muted-foreground">
            {primaryCode === "th"
              ? (language === "th" ? "ภาษาเริ่มต้นของระบบ เปิดใช้งานเสมอ" : "Default system language, always enabled")
              : (language === "th" ? "เปิดใช้งานคู่กับภาษาอังกฤษ" : "Enabled alongside English")}
          </div>
        </div>

        {/* การ์ด English */}
        <div
          className={cn(
            "flex flex-col justify-between gap-3 rounded-xl border p-3 transition-all",
            primaryCode === "en"
              ? "border-primary/50 bg-primary/5 ring-1 ring-primary/20 shadow-xs"
              : isEnActive
                ? "border-border bg-card"
                : "border-dashed border-border/80 bg-muted/20 opacity-75 hover:opacity-100",
          )}
        >
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-3">
              <LanguageFlag code="en" />
              <div>
                <div className="font-bold text-foreground">English</div>
                <div className="text-xs text-muted-foreground font-mono">EN</div>
              </div>
            </div>
            {primaryCode === "en" ? (
              <Badge variant="default" className="gap-1 bg-primary text-primary-foreground font-bold text-xs py-0.5 px-2">
                <Check className="size-3" />
                {textPrimary}
              </Badge>
            ) : isEnActive ? (
              <div className="flex items-center gap-1.5">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="text-xs h-7 font-bold"
                  onClick={() => setPrimaryLanguage("en")}
                >
                  {textSetPrimary}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="text-xs h-7 text-muted-foreground hover:text-destructive"
                  onClick={toggleEnglish}
                  title={language === "th" ? "ปิดใช้งาน English" : "Disable English"}
                >
                  <X className="size-3.5" />
                </Button>
              </div>
            ) : (
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="text-xs h-7 font-bold border-primary/40 text-primary hover:bg-primary/10"
                onClick={toggleEnglish}
              >
                <Plus className="size-3 mr-1" />
                {language === "th" ? "เปิดใช้งาน" : "Enable"}
              </Button>
            )}
          </div>
          <div className="text-xs text-muted-foreground">
            {primaryCode === "en"
              ? (language === "th" ? "ภาษาหลักของระบบ" : "Primary system language")
              : isEnActive
                ? (language === "th" ? "เปิดใช้งานร่วมกับภาษาไทย" : "Enabled alongside Thai")
                : (language === "th" ? "ยังไม่ได้เปิดใช้งาน คลิกเปิดใช้งานเพื่อกรอกข้อมูลภาษาอังกฤษได้" : "Not enabled. Click to enable English.")}
          </div>
        </div>
      </div>
    </section>
  );
}

// ---------------------------------------------------------------------------
// LanguageListEditor
// ---------------------------------------------------------------------------

export function LanguageListEditor({
  field,
  form,
  language,
  label,
  readOnly = false,
  setForm,
}: {
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  label: string;
  readOnly?: boolean;
  setForm?: (update: FormState | ((current: FormState) => FormState)) => void;
}) {
  const rows = normalizeLanguageList(form[field.key], form.language);

  function commit(nextCodes: string[]) {
    if (readOnly || !setForm) return;
    const normalized = normalizeLanguageList(
      nextCodes,
      nextCodes[0] ?? form.language,
      { forcePrimaryFirst: true },
    );
    setForm({
      ...form,
      [field.key]: normalized,
      language: normalized[0] ?? "th",
    });
  }

  const primaryCode = rows[0] ?? form.language ?? "th";
  const isEnActive = rows.includes("en");

  function setPrimaryLanguage(code: LanguageCode) {
    if (readOnly) return;
    if (code === "th") {
      commit(isEnActive ? ["th", "en"] : ["th"]);
    } else {
      commit(["en", "th"]);
    }
  }

  function toggleEnglish() {
    if (readOnly) return;
    if (isEnActive) {
      commit(["th"]);
    } else {
      commit([primaryCode, primaryCode === "en" ? "th" : "en"]);
    }
  }

  const textPrimary = language === "th" ? "ภาษาหลัก" : "Primary";
  const textSetPrimary = language === "th" ? "ตั้งเป็นภาษาหลัก" : "Set as primary";

  return (
    <section className="grid gap-2.5 rounded-2xl border border-border bg-background p-3 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="font-bold text-foreground">{label}</div>
          <div className="text-xs text-muted-foreground">
            {language === "th"
              ? "เลือกภาษาที่ใช้งานสำหรับข้อมูลนี้ (ไทย / อังกฤษ)"
              : "Select languages for this record (Thai / English)"}
          </div>
        </div>
      </div>

      <div className="grid gap-2 sm:grid-cols-2">
        {/* การ์ดภาษาไทย */}
        <div
          className={cn(
            "flex flex-col justify-between gap-3 rounded-xl border p-3 transition-all",
            primaryCode === "th"
              ? "border-primary/50 bg-primary/5 ring-1 ring-primary/20 shadow-xs"
              : "border-border bg-card",
          )}
        >
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-3">
              <LanguageFlag code="th" />
              <div>
                <div className="font-bold text-foreground">ภาษาไทย</div>
                <div className="text-xs text-muted-foreground font-mono">TH</div>
              </div>
            </div>
            {primaryCode === "th" ? (
              <Badge variant="default" className="gap-1 bg-primary text-primary-foreground font-bold text-xs py-0.5 px-2">
                <Check className="size-3" />
                {textPrimary}
              </Badge>
            ) : !readOnly ? (
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="text-xs h-7 font-bold"
                onClick={() => setPrimaryLanguage("th")}
              >
                {textSetPrimary}
              </Button>
            ) : null}
          </div>
          <div className="text-xs text-muted-foreground">
            {primaryCode === "th"
              ? (language === "th" ? "ภาษาเริ่มต้นของระบบ" : "Default system language")
              : (language === "th" ? "เปิดใช้งานร่วมกับภาษาอังกฤษ" : "Enabled alongside English")}
          </div>
        </div>

        {/* การ์ด English */}
        <div
          className={cn(
            "flex flex-col justify-between gap-3 rounded-xl border p-3 transition-all",
            primaryCode === "en"
              ? "border-primary/50 bg-primary/5 ring-1 ring-primary/20 shadow-xs"
              : isEnActive
                ? "border-border bg-card"
                : "border-dashed border-border/80 bg-muted/20 opacity-75 hover:opacity-100",
          )}
        >
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-3">
              <LanguageFlag code="en" />
              <div>
                <div className="font-bold text-foreground">English</div>
                <div className="text-xs text-muted-foreground font-mono">EN</div>
              </div>
            </div>
            {primaryCode === "en" ? (
              <Badge variant="default" className="gap-1 bg-primary text-primary-foreground font-bold text-xs py-0.5 px-2">
                <Check className="size-3" />
                {textPrimary}
              </Badge>
            ) : isEnActive ? (
              !readOnly ? (
                <div className="flex items-center gap-1.5">
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="text-xs h-7 font-bold"
                    onClick={() => setPrimaryLanguage("en")}
                  >
                    {textSetPrimary}
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="text-xs h-7 text-muted-foreground hover:text-destructive"
                    onClick={toggleEnglish}
                    title={language === "th" ? "ปิดใช้งาน English" : "Disable English"}
                  >
                    <X className="size-3.5" />
                  </Button>
                </div>
              ) : null
            ) : !readOnly ? (
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="text-xs h-7 font-bold border-primary/40 text-primary hover:bg-primary/10"
                onClick={toggleEnglish}
              >
                <Plus className="size-3 mr-1" />
                {language === "th" ? "เปิดใช้งาน" : "Enable"}
              </Button>
            ) : null}
          </div>
          <div className="text-xs text-muted-foreground">
            {primaryCode === "en"
              ? (language === "th" ? "ภาษาหลักของข้อมูลนี้" : "Primary language for this record")
              : isEnActive
                ? (language === "th" ? "เปิดใช้งานร่วมกับภาษาไทย" : "Enabled alongside Thai")
                : (language === "th" ? "ยังไม่ได้เปิดใช้งาน" : "Not enabled")}
          </div>
        </div>
      </div>
    </section>
  );
}
