"use client";

import { Check, ChevronsUpDown, Plus, Trash2, X } from "lucide-react";
import Image from "next/image";
import { type DragEvent as ReactDragEvent, useEffect, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { SystemSettingConfig, SystemSettingField } from "@/lib/system-setting-screens";
import { LANGUAGES, normalizeLanguage, type LanguageCode } from "@/lib/i18n";
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
          {LANGUAGES.map((item) => {
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
  const usedCodes = new Set(rows.map((row) => row.code));
  const availableLanguages = LANGUAGES.filter(
    (item) => !usedCodes.has(item.code),
  );
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [draggingCode, setDraggingCode] = useState("");

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

  function reorderByCode(sourceCode: string, targetCode: string) {
    if (!sourceCode || sourceCode === targetCode) return;
    const sourceIndex = rows.findIndex((row) => row.code === sourceCode);
    const targetIndex = rows.findIndex((row) => row.code === targetCode);
    if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex)
      return;
    commit(moveArrayItem(rows, sourceIndex, targetIndex));
  }

  function handleDragStart(event: ReactDragEvent<HTMLElement>, code: string) {
    if (rows.length <= 1) return;
    setDraggingCode(code);
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", code);
  }

  function handleDragOver(
    event: ReactDragEvent<HTMLElement>,
    targetCode: string,
  ) {
    const sourceCode = draggingCode || event.dataTransfer.getData("text/plain");
    if (!sourceCode || sourceCode === targetCode) return;
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
    reorderByCode(sourceCode, targetCode);
  }

  function handleDrop(event: ReactDragEvent<HTMLElement>, targetCode: string) {
    event.preventDefault();
    reorderByCode(
      draggingCode || event.dataTransfer.getData("text/plain"),
      targetCode,
    );
    setDraggingCode("");
  }

  const textPrimary = language === "th" ? "ภาษาแรก" : "Primary language";
  const textAdd = language === "th" ? "เพิ่มภาษา" : "Add language";
  const textRemove = language === "th" ? "เอาออก" : "Remove";
  const textNoMore =
    language === "th"
      ? "เพิ่มครบทุกภาษาที่รองรับแล้ว"
      : "All supported languages are already added.";
  const textDrag =
    language === "th" ? "ลากเพื่อจัดลำดับภาษา" : "Drag to reorder language";

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="font-semibold">{label}</div>
          <div className="text-xs text-muted-foreground">
            {language === "th"
              ? "ลำดับแรกคือภาษาแรกของบริษัท"
              : "The first row is the company primary language."}
          </div>
        </div>
        {availableLanguages.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              size="sm"
              onClick={() => setAddDialogOpen(true)}
            >
              <Plus />
              {textAdd}
            </Button>
          </div>
        ) : null}
      </div>
      <div className="grid gap-1">
        {rows.map((row, index) => (
          <div
            key={row.code}
            draggable={rows.length > 1}
            onDragEnd={() => setDraggingCode("")}
            onDragOver={(event) => handleDragOver(event, row.code)}
            onDragStart={(event) => handleDragStart(event, row.code)}
            onDrop={(event) => handleDrop(event, row.code)}
            title={textDrag}
            className={cn(
              "grid cursor-grab gap-2 rounded-xl border border-border bg-card px-2 py-1.5 transition-[transform,box-shadow,border-color,background-color,opacity] duration-150 ease-out hover:-translate-y-0.5 hover:shadow-sm active:cursor-grabbing sm:grid-cols-[84px_minmax(0,1fr)_auto] sm:items-center",
              index === 0 && "border-primary/40 bg-primary/5",
              draggingCode === row.code &&
                "scale-[0.99] opacity-60 ring-2 ring-primary/30",
            )}
          >
            <div className="flex items-center gap-1">
              <span className="grid size-8 place-items-center rounded-lg border border-border bg-background text-muted-foreground">
                <ChevronsUpDown className="size-4" aria-hidden="true" />
              </span>
              <div
                className={cn(
                  "grid size-10 place-items-center rounded-full font-semibold",
                  index === 0
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-foreground",
                )}
              >
                {index + 1}
              </div>
            </div>
            <div className="flex min-w-0 items-center gap-2">
              <LanguageFlag code={row.code} />
              <div className="min-w-0">
                <div className="truncate font-semibold">
                  {languageName(row.code, language)}
                </div>
                <div className="flex flex-wrap items-center gap-2 text-xs uppercase text-muted-foreground">
                  <span>{row.code}</span>
                  {index === 0 ? (
                    <Badge variant="outline">{textPrimary}</Badge>
                  ) : null}
                </div>
              </div>
            </div>
            <div className="flex flex-wrap justify-end gap-1">
              {rows.length > 1 ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() =>
                    commit(rows.filter((item) => item.code !== row.code))
                  }
                  title={textRemove}
                >
                  <Trash2 />
                  <span className="sr-only">{textRemove}</span>
                </Button>
              ) : null}
            </div>
          </div>
        ))}
      </div>
      {availableLanguages.length === 0 ? (
        <p className="text-xs text-muted-foreground">{textNoMore}</p>
      ) : null}
      <LanguageAddDialog
        activeCodes={rows.map((row) => row.code)}
        language={language}
        open={addDialogOpen}
        onClose={() => setAddDialogOpen(false)}
        onToggle={(code) => {
          const isActive = rows.some((row) => row.code === code);
          if (isActive) {
            if (rows[0]?.code === code) return;
            commit(rows.filter((row) => row.code !== code));
          } else {
            commit([...rows, languageConfigRow(code, false)]);
          }
        }}
      />
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
  const usedCodes = new Set(rows);
  const availableLanguages = LANGUAGES.filter(
    (item) => !usedCodes.has(item.code),
  );
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [draggingCode, setDraggingCode] = useState("");
  const textPrimary = language === "th" ? "ภาษาแรก" : "Primary";
  const textAdd = language === "th" ? "เพิ่มภาษา" : "Add language";
  const textRemove = language === "th" ? "เอาออก" : "Remove";
  const textNoMore =
    language === "th"
      ? "เพิ่มครบทุกภาษาที่รองรับแล้ว"
      : "All supported languages are already added.";
  const textDrag =
    language === "th" ? "ลากเพื่อจัดลำดับภาษา" : "Drag to reorder language";

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

  function reorderByCode(sourceCode: string, targetCode: string) {
    if (readOnly) return;
    if (!sourceCode || sourceCode === targetCode) return;
    const sourceIndex = rows.findIndex((code) => code === sourceCode);
    const targetIndex = rows.findIndex((code) => code === targetCode);
    if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex)
      return;
    commit(moveArrayItem(rows, sourceIndex, targetIndex));
  }

  function handleDragStart(event: ReactDragEvent<HTMLElement>, code: string) {
    if (readOnly) return;
    if (rows.length <= 1) return;
    setDraggingCode(code);
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", code);
  }

  function handleDragOver(
    event: ReactDragEvent<HTMLElement>,
    targetCode: string,
  ) {
    if (readOnly) return;
    const sourceCode = draggingCode || event.dataTransfer.getData("text/plain");
    if (!sourceCode || sourceCode === targetCode) return;
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
    reorderByCode(sourceCode, targetCode);
  }

  function handleDrop(event: ReactDragEvent<HTMLElement>, targetCode: string) {
    if (readOnly) return;
    event.preventDefault();
    reorderByCode(
      draggingCode || event.dataTransfer.getData("text/plain"),
      targetCode,
    );
    setDraggingCode("");
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="font-semibold">{label}</div>
          <div className="text-xs text-muted-foreground">
            {language === "th"
              ? "ลำดับแรกคือภาษาแรกของข้อมูลนี้"
              : "The first row is the primary language for this record."}
          </div>
        </div>
        {!readOnly && availableLanguages.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              size="sm"
              onClick={() => setAddDialogOpen(true)}
            >
              <Plus />
              {textAdd}
            </Button>
          </div>
        ) : null}
      </div>
      <div className="grid gap-1">
        {rows.map((code, index) => (
          <div
            key={code}
            draggable={!readOnly && rows.length > 1}
            onDragEnd={() => setDraggingCode("")}
            onDragOver={(event) => handleDragOver(event, code)}
            onDragStart={(event) => handleDragStart(event, code)}
            onDrop={(event) => handleDrop(event, code)}
            title={textDrag}
            className={cn(
              "grid cursor-grab gap-2 rounded-xl border border-border bg-card px-2 py-1.5 transition-[transform,box-shadow,border-color,background-color,opacity] duration-150 ease-out hover:-translate-y-0.5 hover:shadow-sm active:cursor-grabbing sm:grid-cols-[84px_minmax(0,1fr)_auto] sm:items-center",
              readOnly &&
                "cursor-default hover:translate-y-0 active:cursor-default",
              index === 0 && "border-primary/40 bg-primary/5",
              draggingCode === code &&
                "scale-[0.99] opacity-60 ring-2 ring-primary/30",
            )}
          >
            <div className="flex items-center gap-1">
              <span className="grid size-8 place-items-center rounded-lg border border-border bg-background text-muted-foreground">
                <ChevronsUpDown className="size-4" aria-hidden="true" />
              </span>
              <div
                className={cn(
                  "grid size-10 place-items-center rounded-full font-semibold",
                  index === 0
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-foreground",
                )}
              >
                {index + 1}
              </div>
            </div>
            <div className="flex min-w-0 items-center gap-2">
              <LanguageFlag code={code} />
              <div className="min-w-0">
                <div className="truncate font-semibold">
                  {languageName(code, language)}
                </div>
                <div className="flex flex-wrap items-center gap-2 text-xs uppercase text-muted-foreground">
                  <span>{code}</span>
                  {index === 0 ? (
                    <Badge variant="outline">{textPrimary}</Badge>
                  ) : null}
                </div>
              </div>
            </div>
            <div className="flex flex-wrap justify-end gap-1">
              {!readOnly && rows.length > 1 ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => commit(rows.filter((item) => item !== code))}
                  title={textRemove}
                >
                  <Trash2 />
                  <span className="sr-only">{textRemove}</span>
                </Button>
              ) : null}
            </div>
          </div>
        ))}
      </div>
      {!readOnly && availableLanguages.length === 0 ? (
        <p className="text-xs text-muted-foreground">{textNoMore}</p>
      ) : null}
      {!readOnly ? (
        <LanguageAddDialog
          activeCodes={rows}
          language={language}
          open={addDialogOpen}
          onClose={() => setAddDialogOpen(false)}
          onToggle={(code) => {
            const isActive = rows.includes(code);
            if (isActive) {
              if (rows[0] === code) return;
              commit(rows.filter((item) => item !== code));
            } else {
              commit([...rows, code]);
            }
          }}
        />
      ) : null}
    </section>
  );
}
