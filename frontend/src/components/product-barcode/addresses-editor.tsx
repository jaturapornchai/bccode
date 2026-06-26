"use client";

import { useMemo } from "react";
import { Textarea } from "@/components/ui/textarea";
import { LANGUAGES } from "@/lib/i18n";
import { LanguageFlag } from "@/components/product-barcode/names-editor";
import { cn } from "@/lib/utils";

function languageLabel(code: string, language: string): string {
  const item = LANGUAGES.find((entry) => entry.code === code);
  if (!item) return code.toUpperCase();
  if (language === "th") return item.name;
  return `${item.name} (${item.code.toUpperCase()})`;
}

// Per-language multiline address editor. Mirrors NamesEditor layout but renders a
// <Textarea> per language because document addresses span multiple lines.
// Data shape is a { [languagecode]: addressText } map (text keeps user newlines).
export function AddressesEditor({
  addresses,
  onChange,
  languages,
  label,
  language = "th",
  disabled,
}: {
  addresses: Record<string, string>;
  onChange: (next: Record<string, string>) => void;
  languages: string[];
  label: string;
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
    Object.keys(addresses).forEach((code) => {
      if (code.trim()) codes.add(code.trim());
    });
    return Array.from(codes);
  }, [languages, addresses]);

  return (
    <div className="space-y-2">
      <div className="text-sm font-medium">{label}</div>
      <div className={cn("grid gap-2", allLanguages.length > 1 && "grid-cols-2")}>
        {allLanguages.map((code, index) => (
          <label className="grid gap-1 text-sm font-semibold" key={code}>
            <span className="flex min-w-0 items-center gap-2 text-xs text-muted-foreground">
              <LanguageFlag code={code} />
              <span className="truncate">
                {index === 0
                  ? language === "th"
                    ? "ภาษาแรก"
                    : "Primary"
                  : languageLabel(code, language)}
              </span>
              <span className="uppercase">{code}</span>
            </span>
            <Textarea
              value={addresses[code] ?? ""}
              onChange={(event) => onChange({ ...addresses, [code]: event.target.value })}
              placeholder={language === "th" ? "ที่อยู่สำหรับออกเอกสาร" : "Address for documents"}
              rows={4}
              disabled={disabled}
            />
          </label>
        ))}
      </div>
    </div>
  );
}
