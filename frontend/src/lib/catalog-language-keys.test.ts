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
  "src/lib/permission-actions.ts",
  "src/app/menu/product-barcode-shelf-screen.tsx",
  "src/app/menu/product-price-history-screen.tsx",
  "src/app/menu/datamodel-graph-screen.tsx",
  "src/app/workspace/workspace-screen.tsx",
  "src/app/settings/settings-screen.tsx",
  "src/app/login-screen.tsx",
  "src/components/system-settings/field-editors/role-screen-matrix.tsx",
];
const screens = [
  "src/app/operations/operations-workbench.tsx",
  "src/app/tools/erp-tools-screen.tsx",
  "src/app/report/erp-report-viewer.tsx",
  "src/app/tax/tax-filing-workbench.tsx",
  "src/app/asset/fixed-assets-screen.tsx",
  "src/app/settings/settings-screen.tsx",
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

// Every catalog names its map `catalogKeys` or `<something>Keys`; the block runs
// from the declaration to the first `};`.
function catalogKeysOf(source: string): string[] {
  const keys: string[] = [];
  for (const match of source.matchAll(/const (?:catalogKeys|\w*(?:TextKeys|ACTION_KEYS|CatalogKeys))\b/g)) {
    const start = match.index ?? 0;
    const block = source.slice(start, source.indexOf("};", start));
    keys.push(...[...block.matchAll(/:\s*"([a-z0-9_]+)",/g)].map((entry) => entry[1] ?? ""));
  }
  return keys;
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

  it("no catalog file picks a language for the user any more", () => {
    // A locale argument (`isThai ? "th-TH" : "en-US"`) is a number format, not a
    // label, so only a quoted string that is itself rendered text is forbidden.
    const bilingualText = /(?:\bisThai|language === "th"|lang === "th")\s*\?\s*(?:\r?\n\s*)?["`](?!th-TH|en-US|en-GB|th"|en")/;
    for (const file of catalogs) {
      const source = read(file);
      expect(`${file}: ${bilingualText.test(source)}`).toBe(`${file}: false`);
    }
  });

  it("the screens read the dictionary instead of choosing between Thai and English", () => {
    // `language === "th" ? text.th : text.en` survives in catalog-text.ts only,
    // as the offline fallback when no dictionary has loaded yet.
    const bilingualText = /(?:\bisThai|language === "th"|lang === "th")\s*\?\s*(?:\r?\n\s*)?["`](?!th-TH|en-US|en-GB)/;
    for (const file of screens) {
      const source = read(file);
      // Either shape is fine: the dictionary itself, or the memoised `tr` built
      // from it — what matters is that the string comes from the provider.
      expect(`${file}: ${/useBackend(?:Dictionary|Text|Language)\(/.test(source)}`).toBe(`${file}: true`);
      expect(`${file}: ${bilingualText.test(source)}`).toBe(`${file}: false`);
      expect(`${file}: ${/language === "th" \? config\./.test(source)}`).toBe(`${file}: false`);
    }
  });
});
