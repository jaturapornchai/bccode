import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

// Guard for the AGENTS.md rule (2026-09-14): screen text must follow the selected
// language — these screens carry English keys only, the texts live in languages.tsv.
const sources = [
  "src/app/menu/product-screen.tsx",
  "src/app/menu/product-barcode-screen.tsx",
  "src/app/menu/product-set-screen.tsx",
].map((file) => resolve(process.cwd(), file));
const languageColumns = ["th", "en", "cn", "ja", "km", "ko", "lo", "my", "vi", "ms", "id", "fil"];
// Thai letters/vowels/tone marks/digits — the baht sign ฿ (U+0E3F) is a currency symbol, not text.
const thaiText = /[ก-฾เ-๛]/;
// Allowed carriers of the Thai fallback: tr("key", "ไทย") / t("key", "ไทย") calls and ["key", "ไทย"] label tuples.
const keyedCall = /\b(?:tr|t)\(\s*"([a-z0-9_.]+)"\s*,\s*(?:"(?:[^"\\]|\\.)*"|`[^`]*`)\s*,?\s*\)/gs;
const keyedBackendText = /\bbackendText\(\s*[^,]+,\s*"([a-z0-9_.]+)"\s*,\s*(?:"(?:[^"\\]|\\.)*"|`[^`]*`)\s*,?\s*\)/gs;
const keyedTuple = /\[\s*"([a-z0-9_.]+)"\s*,\s*"[^"]*[ก-฾เ-๛][^"]*"\s*\]/g;
// Stored data defaults or internal predicates, not UI labels.
const dataDefault = /\bcode:\s*(?:"th"|'th'|language),\s*name:\s*|normalized === "ไม่ใช่"/;

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

describe("product screens language keys", () => {
  it("every key used exists in languages.tsv with all language cells", () => {
    const rows = backendLanguageRows();
    const missing = new Set<string>();
    for (const file of sources) {
      const source = stripComments(readFileSync(file, "utf8"));
      const matches = [
        ...source.matchAll(keyedCall),
        ...source.matchAll(keyedBackendText),
        ...source.matchAll(keyedTuple),
      ];
      for (const match of matches) {
        const key = match[1];
        if (!key) continue;
        const row = rows.get(key);
        if (!row) {
          missing.add(`${file.split(/[\\/]/).pop()}:${key}:row`);
          continue;
        }
        for (const language of languageColumns) {
          if (!row[language]?.trim()) {
            missing.add(`${file.split(/[\\/]/).pop()}:${key}:${language}`);
          }
        }
      }
    }
    expect([...missing].sort()).toEqual([]);
  });

  it("screens carry no hard-coded Thai outside keyed calls and label tuples", () => {
    const leftovers: string[] = [];
    for (const file of sources) {
      let source = stripComments(readFileSync(file, "utf8"));
      // Replace matching keyed calls preserving newlines
      source = source.replace(keyedCall, (match) => match.replace(/[^\n]/g, " "));
      source = source.replace(keyedBackendText, (match) => match.replace(/[^\n]/g, " "));
      source = source.replace(keyedTuple, (match) => match.replace(/[^\n]/g, " "));
      source.split(/\r?\n/).forEach((line, index) => {
        if (thaiText.test(line) && !dataDefault.test(line)) {
          leftovers.push(`${file.split(/[\\/]/).pop()}:${index + 1}`);
        }
      });
    }
    expect(leftovers).toEqual([]);
  });
});
