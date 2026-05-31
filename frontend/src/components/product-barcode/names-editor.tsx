"use client";

import { useMemo } from "react";
import Image from "next/image";
import { Input } from "@/components/ui/input";
import { LANGUAGES, normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { type NameX } from "@/lib/product-barcode/types";
import { isRecord, setNameXEntry } from "@/lib/product-barcode/utils";
import { cn } from "@/lib/utils";
import { type WorkspaceSession } from "@/lib/workspace-models";

function booleanFromUnknown(value: unknown): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "number") return value !== 0;
  if (typeof value === "string") {
    const normalized = value.trim().toLowerCase();
    return normalized === "true" || normalized === "1" || normalized === "yes" || normalized === "y";
  }
  return false;
}

function supportedLanguageCode(
  value: unknown,
  fallback: LanguageCode | "",
): LanguageCode | "" {
  const raw = typeof value === "string" ? value.trim().toLowerCase() : "";
  if (!raw) return fallback;
  const normalized = normalizeLanguage(raw);
  return LANGUAGES.some((item) => item.code === normalized)
    ? (normalized as LanguageCode)
    : fallback;
}

function languageName(code: string, language: LanguageCode | string): string {
  const item = LANGUAGES.find((entry) => entry.code === code);
  if (!item) return code.toUpperCase();
  if (language === "th") return item.name;
  return `${item.name} (${item.code.toUpperCase()})`;
}

export function languageCodesFromWorkspace(workspace: WorkspaceSession | null): string[] {
  const shopInfo = isRecord(workspace?.shopInfo) ? workspace.shopInfo : {};
  const settings = isRecord(shopInfo.settings) ? shopInfo.settings : {};

  const rawConfigs = settings.languageconfigs ?? shopInfo["settings.languageconfigs"] ?? shopInfo.languageconfigs;
  const rawLang = settings.language ?? shopInfo["settings.language"] ?? shopInfo.language;

  const rows = Array.isArray(rawConfigs)
    ? rawConfigs
        .filter(isRecord)
        .map((item) => ({
          code: supportedLanguageCode(item.code, ""),
          isUse: item.is_use === undefined && item.isuse === undefined ? true : booleanFromUnknown(item.is_use ?? item.isuse),
          isDefault: booleanFromUnknown(item.isdefault),
        }))
        .filter((item) => item.code && item.isUse)
    : [];
  const configuredDefault = supportedLanguageCode(rawLang, "");
  const primary = rows.find((item) => item.isDefault)?.code || (configuredDefault as string) || rows[0]?.code || "th";
  const ordered = [primary, ...rows.map((item) => item.code).filter((code) => code && code !== primary)];
  return Array.from(new Set(ordered));
}

export function LanguageFlag({ code }: { code: string }) {
  const normalized = supportedLanguageCode(code, "th");
  return (
    <span className="grid size-7 shrink-0 place-items-center overflow-hidden rounded border border-border bg-card">
      <Image alt="" height={18} src={`/flags/${normalized}.png`} width={27} />
    </span>
  );
}

export function NamesEditor({
  names,
  onChange,
  languages,
  label,
  firstRequired,
  error,
  language = "th",
  disabled,
}: {
  names: NameX[];
  onChange: (next: NameX[]) => void;
  languages: string[];
  label: string;
  firstRequired?: boolean;
  error?: string;
  language?: string;
  disabled?: boolean;
}) {
  const allLanguages = useMemo(() => {
    const codes = new Set<string>();
    const primary = languages.find((code) => code.trim())?.trim() || "th";
    codes.add(primary);
    languages.forEach((code) => {
      if (code.trim()) codes.add(code.trim());
    });
    names.forEach((entry) => {
      const code = entry.code?.trim();
      if (code) codes.add(code);
    });
    return Array.from(codes);
  }, [languages, names]);
  const primaryLanguage = languages.find((code) => code.trim())?.trim() || allLanguages[0] || "th";

  return (
    <div className="space-y-2">
      <div className="text-sm font-medium">
        {label}
        {firstRequired ? <span className="ml-1 text-destructive">*</span> : null}
      </div>
      <div className={cn("grid gap-2", allLanguages.length > 1 && "grid-cols-2")}>
        {allLanguages.map((code, index) => {
          const entry = names.find((n) => n.code === code);
          return (
            <label className="grid gap-1 text-sm font-semibold" key={code}>
              <span className="flex min-w-0 items-center gap-2 text-xs text-muted-foreground">
                <LanguageFlag code={code} />
                <span className="truncate">
                  {index === 0
                    ? language === "th"
                      ? "ภาษาแรก"
                      : "Primary"
                    : languageName(code, language)}
                </span>
                <span className="uppercase">{code}</span>
              </span>
              <Input
                value={entry?.name ?? ""}
                onChange={(event) => onChange(setNameXEntry(names, code, event.target.value))}
                placeholder={code}
                disabled={disabled}
                aria-invalid={firstRequired && code === primaryLanguage && !entry?.name ? true : undefined}
              />
            </label>
          );
        })}
      </div>
      {error ? <p className="text-xs text-destructive">{error}</p> : null}
    </div>
  );
}
