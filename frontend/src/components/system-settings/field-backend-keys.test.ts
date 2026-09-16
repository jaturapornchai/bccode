import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

import { fieldBackendKeys } from "./utils";

// Guard for the AGENTS.md rule (2026-09-14): a settings field that maps to a
// language key renders that row, and fieldLabel() falls back to the ENGLISH
// label when the row is missing — so a missing row means every non-Thai user
// reads English. 2026-09-16: 26 mappings pointed at rows that did not exist.
const languageColumns = ["th", "en", "cn", "ja", "km", "ko", "lo", "my", "vi", "ms", "id", "fil"];

// Mapped keys that intentionally have no row yet, with the reason.
const knownGaps: Record<string, string> = {
  // The permissiondefinition screen has no declared "accessrules" field; the
  // permission editor builds it from form data, so there is no label to mint.
  accessrules: "no declared field to take the label from",
};

function languageRows(): Map<string, Record<string, string>> {
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

describe("fieldBackendKeys", () => {
  const rows = languageRows();

  it("every mapped key has a row in languages.tsv", () => {
    const missing = [...new Set(Object.values(fieldBackendKeys))]
      .filter((key) => !rows.has(key) && !(key in knownGaps))
      .sort();
    expect(missing).toEqual([]);
  });

  it("every mapped row carries all twelve languages", () => {
    const incomplete: string[] = [];
    for (const key of new Set(Object.values(fieldBackendKeys))) {
      const row = rows.get(key);
      if (!row) continue;
      const blanks = languageColumns.filter((column) => !row[column]?.trim());
      if (blanks.length) incomplete.push(`${key}: ${blanks.join(",")}`);
    }
    expect(incomplete).toEqual([]);
  });
});
