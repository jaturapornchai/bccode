import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

// Guard for the AGENTS.md rule (2026-09-14). The work-tab catalogs carry their
// own `{ th, en }` literals; those are only a fallback now, and every string is
// also a row in languages.tsv. A missing row silently drops the other ten
// languages back to English, which is exactly the bug this round closed.
const catalogs = [
  "src/lib/erp-operations.ts",
  "src/lib/erp-tools.ts",
  "src/lib/erp-reports.ts",
  "src/lib/thai-tax.ts",
];
const screens = [
  "src/app/operations/operations-workbench.tsx",
  "src/app/tools/erp-tools-screen.tsx",
  "src/app/report/erp-report-viewer.tsx",
  "src/app/tax/tax-filing-workbench.tsx",
];
const languageColumns = ["th", "en", "cn", "ja", "km", "ko", "lo", "my", "vi", "ms", "id", "fil"];

function backendLanguageRows(): Map<string, Record<string, string>> {
  const path = resolve(process.cwd(), "..", "backend", "assets", "language", "languages.tsv");
  const [headerLine = "", ...lines] = readFileSync(path, "utf8").split(/\r?\n/);
  const headers = headerLine.split("\t");
  const rows = new Map<string, Record<string, string>>();
  for (const line of lines) {
    if (!line || line.startsWith("#")) continue;
    const columns = line.split("\t");
    rows.set(columns[0] ?? "", Object.fromEntries(headers.map((header, index) => [header, columns[index] ?? ""])));
  }
  return rows;
}

function read(file: string): string {
  return readFileSync(resolve(process.cwd(), file), "utf8");
}

function catalogKeysOf(source: string): string[] {
  const start = source.indexOf("const catalogKeys");
  if (start < 0) return [];
  const block = source.slice(start, source.indexOf("};", start));
  return [...block.matchAll(/:\s*"([a-z0-9_]+)",/g)].map((match) => match[1] ?? "");
}

describe("work-tab catalog language keys", () => {
  it("every catalog key exists in languages.tsv with all language cells", () => {
    const rows = backendLanguageRows();
    const missing = new Set<string>();
    for (const file of catalogs) {
      const keys = catalogKeysOf(read(file));
      expect(`${file}: ${keys.length > 0}`).toBe(`${file}: true`);
      for (const key of keys) {
        const row = rows.get(key);
        if (!row) {
          missing.add(`${file}:${key}:row`);
          continue;
        }
        for (const language of languageColumns) {
          if (!row[language]?.trim()) missing.add(`${file}:${key}:${language}`);
        }
      }
    }
    expect([...missing].sort()).toEqual([]);
  });

  it("every message-table and tr() key on these screens exists too", () => {
    const rows = backendLanguageRows();
    const missing = new Set<string>();
    for (const file of screens) {
      const source = read(file);
      const keys = [
        ...[...source.matchAll(/\btr\(\s*"([a-z0-9_.]+)"\s*,/g)].map((m) => m[1] ?? ""),
        ...[...source.matchAll(/\bkey:\s*"([a-z0-9_]+)"/g)].map((m) => m[1] ?? ""),
      ];
      for (const key of keys) {
        const row = rows.get(key);
        if (!row) {
          missing.add(`${file}:${key}:row`);
          continue;
        }
        for (const language of languageColumns) {
          if (!row[language]?.trim()) missing.add(`${file}:${key}:${language}`);
        }
      }
    }
    expect([...missing].sort()).toEqual([]);
  });

  it("the screens read the dictionary instead of choosing between Thai and English", () => {
    // `language === "th" ? text.th : text.en` survives in catalog-text.ts only,
    // as the offline fallback when no dictionary has loaded yet.
    const bilingualText = /(?:\bisThai|language === "th"|lang === "th")\s*\?\s*(?:\r?\n\s*)?["`](?!th-TH|en-US|en-GB)/;
    for (const file of screens) {
      const source = read(file);
      expect(`${file}: ${source.includes("useBackendDictionary")}`).toBe(`${file}: true`);
      expect(`${file}: ${bilingualText.test(source)}`).toBe(`${file}: false`);
      expect(`${file}: ${/language === "th" \? config\./.test(source)}`).toBe(`${file}: false`);
    }
  });
});
