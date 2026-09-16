import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

// Guard for the AGENTS.md rule (2026-09-14): a tr("key", …) whose row is missing
// prints the raw key on screen, and a row with an empty cell falls back to the
// key as well. These editors sit under the settings screen's BackendTextProvider.
const sources = [
  "src/app/system-settings/product-category-items-editor.tsx",
  "src/components/system-settings/field-editors/holding-scope-editor.tsx",
  "src/components/system-settings/field-editors/branch-company-selectors.tsx",
  "src/components/system-settings/field-editors/permission-editors.tsx",
  "src/components/system-settings/field-editors/permission-sets-editor.tsx",
].map((file) => resolve(process.cwd(), file));
const languageColumns = ["th", "en", "cn", "ja", "km", "ko", "lo", "my", "vi", "ms", "id", "fil"];
const keyedCall = /\btr\(\s*"([a-z0-9_.]+)"\s*,/g;

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

describe("settings field editors language keys", () => {
  it("every key used exists in languages.tsv with all language cells", () => {
    const rows = backendLanguageRows();
    const missing = new Set<string>();
    for (const file of sources) {
      const source = readFileSync(file, "utf8");
      for (const match of source.matchAll(keyedCall)) {
        const key = match[1];
        if (!key) continue;
        const row = rows.get(key);
        if (!row) {
          missing.add(`${file.split(/[\\/]/).pop()}:${key}:row`);
          continue;
        }
        for (const language of languageColumns) {
          if (!row[language]?.trim()) missing.add(`${file.split(/[\\/]/).pop()}:${key}:${language}`);
        }
      }
    }
    expect([...missing].sort()).toEqual([]);
  });

  it("reads the dictionary from context instead of choosing between Thai and English", () => {
    for (const file of sources) {
      const source = readFileSync(file, "utf8");
      const label = file.split(/[\\/]/).pop();
      expect(`${label}: ${source.includes("useBackendText")}`).toBe(`${label}: true`);
      // Only text is forbidden: `isThai ? "th-TH" : "en-US"` for toLocaleString
      // is a number format, not a label, so the pattern looks for a quoted
      // string that is itself the rendered text.
      const bilingualText = /(?:\bisThai|language === "th"|lang === "th")\s*\?\s*(?:\r?\n\s*)?["`](?!th-TH|en-US|en-GB)/;
      expect(`${label}: ${bilingualText.test(source)}`).toBe(`${label}: false`);
    }
  });
});
