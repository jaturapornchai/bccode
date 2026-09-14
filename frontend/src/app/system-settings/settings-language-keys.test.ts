import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

// Guard for the AGENTS.md rule (2026-09-14): screen text must follow the selected
// language — these screens carry English keys only, the texts live in languages.tsv.
const sources = [
  "src/app/system-settings/warehouse-tree-view.tsx",
  "src/app/system-settings/company-branch-tree-view.tsx",
  "src/app/system-settings/product-category-tree-view.tsx",
  "src/app/system-settings/product-group-tree-view.tsx",
  "src/app/system-settings/product-bom-editor.tsx",
  "src/app/menu/manage-shortcuts-screen.tsx",
].map((file) => resolve(process.cwd(), file));
const languageColumns = ["th", "en", "cn", "ja", "km", "ko", "lo", "my", "vi", "ms", "id", "fil"];
// Thai letters/vowels/tone marks/digits — the baht sign ฿ (U+0E3F) is a currency symbol, not text.
const thaiText = /[ก-฾เ-๛]/;
// Allowed carriers of the Thai fallback: tr("key", "ไทย") / t("key", "ไทย") calls and ["key", "ไทย"] label tuples.
const keyedCall = /\b(?:tr|t)\(\s*"([a-z0-9_.]+)"\s*,\s*"(?:[^"\\]|\\.)*"\s*\)/g;
const keyedTuple = /\[\s*"([a-z0-9_.]+)"\s*,\s*"[^"]*[ก-฾เ-๛][^"]*"\s*\]/g;
// Stored data defaults, not UI labels: the Thai unit name of a recipe ({ code: "th", name: "สูตร" }).
const dataDefault = /\bcode: (?:"th"|language), name: /;

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

function stripComments(source: string) {
  // Keep line breaks so reported line numbers match the file.
  return source
    .replace(/\{?\/\*[\s\S]*?\*\/\}?/g, (comment) => comment.replace(/[^\n]/g, ""))
    .replace(/(?:^|[ \t])\/\/[^\n]*$/gm, "");
}

describe("system settings / shortcuts language keys", () => {
  it("every key used exists in languages.tsv with all language cells", () => {
    const rows = backendLanguageRows();
    const missing = new Set<string>();
    for (const file of sources) {
      const source = stripComments(readFileSync(file, "utf8"));
      for (const match of [...source.matchAll(keyedCall), ...source.matchAll(keyedTuple)]) {
        const key = match[1];
        if (!key) continue;
        const row = rows.get(key);
        if (!row) { missing.add(`${key}:row`); continue; }
        for (const language of languageColumns) if (!row[language]?.trim()) missing.add(`${key}:${language}`);
      }
    }
    expect([...missing].sort()).toEqual([]);
  });

  it("screens carry no hard-coded Thai outside keyed calls and label tuples", () => {
    const leftovers: string[] = [];
    for (const file of sources) {
      const source = stripComments(readFileSync(file, "utf8")).replace(keyedCall, "").replace(keyedTuple, "");
      source.split(/\r?\n/).forEach((line, index) => {
        if (thaiText.test(line) && !dataDefault.test(line)) leftovers.push(`${file.split(/[\\/]/).pop()}:${index + 1}`);
      });
    }
    expect(leftovers).toEqual([]);
  });
});
