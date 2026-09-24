import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

// จอ 50 ทวิ (ออกใบ + ดูใบที่ได้รับ) — key ที่เรียกผ่าน tr("…") และ map { key: "…" } / ["key", "ไทย"] ต้องมีแถวใน languages.tsv
// แถวที่ขาดทำให้จอภาษาอังกฤษค้างเป็นไทย (fallback) โดยไม่มีใครเห็น
const panel = resolve(process.cwd(), "src", "app", "tax", "wht-certificate-panel.tsx");
const languageColumns = ["th", "en", "cn", "ja", "km", "ko", "lo", "my", "vi", "ms", "id", "fil"];

function languageRows(): Map<string, Record<string, string>> {
  const path = resolve(process.cwd(), "..", "backend", "assets", "language", "languages.tsv");
  const [headerLine = "", ...lines] = readFileSync(path, "utf8").split(/\r?\n/);
  const headers = headerLine.split("\t");
  const rows = new Map<string, Record<string, string>>();
  for (const line of lines) {
    if (!line) continue;
    const columns = line.split("\t");
    rows.set(columns[0] ?? "", Object.fromEntries(headers.map((header, index) => [header, columns[index] ?? ""])));
  }
  return rows;
}

describe("wht certificate panel language keys", () => {
  it("ทุก key ของจอ 50 ทวิ มีแถวครบทุกภาษา", () => {
    const source = readFileSync(panel, "utf8");
    const keys = new Set<string>();
    for (const match of source.matchAll(/\btr\(\s*"([a-z0-9_]+)"|\bkey: "([a-z0-9_]+)"|\[\s*"([a-z][a-z0-9_]*)",\s*"/g)) {
      keys.add(match[1] ?? match[2] ?? match[3] ?? "");
    }
    expect(keys.size).toBeGreaterThan(20);
    const rows = languageRows();
    const missing: string[] = [];
    for (const key of keys) {
      const row = rows.get(key);
      if (!row) { missing.push(`${key}:row`); continue; }
      for (const language of languageColumns) if (!row[language]?.trim()) missing.push(`${key}:${language}`);
    }
    expect(missing.sort()).toEqual([]);
  });
});
