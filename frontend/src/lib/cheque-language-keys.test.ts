import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

// กฎ AGENTS.md "Language: key+th/en": จอเช็ค/ธนาคารห้ามฝังข้อความไทยในโค้ด — ใช้ key + tr(key, ข้อความไทยสำรอง)
// และทุก key ต้องมีแถวครบ 13 คอลัมน์ใน backend/assets/language/languages.tsv (ไม่มีแถว = ผู้ใช้เห็น key ดิบ)
const sources = [
  "src/components/bank-code-screen.tsx",
  "src/components/bank-contact-screen.tsx",
  "src/components/book-bank-screen.tsx",
  "src/lib/cheque-api.ts",
  "src/lib/cheque-forms.ts",
  "src/lib/bank-code-kinds.ts",
  "src/app/api/cheque/[...chequePath]/route.ts",
].map((file) => resolve(process.cwd(), file));

const thaiText = /[ก-฾เ-๛]/;
const keyedCall = /\btr\(\s*"([a-z0-9_]+)"\s*,\s*"(?:[^"\\]|\\.)*"\s*,?\s*\)/g;
const keyedTuple = /\[\s*"([a-z0-9_]+)"\s*,\s*"[^"]*[ก-฾เ-๛][^"]*"\s*\]/g;
// key ข้อความผิดพลาดที่ BFF/ไลบรารีส่งเป็น literal แล้วให้จอแปลผ่าน tr(code, …)
const errorKey = /"((?:cheque_err|bank_code_err|bank_contact_err|bank_book_err)_[a-z0-9_]+|cheque_request_failed)"/g;

function stripComments(source: string) {
  return source
    .replace(/\{?\/\*[\s\S]*?\*\/\}?/g, (comment) => comment.replace(/[^\n]/g, ""))
    .replace(/(?:^|[ \t])\/\/[^\n]*$/gm, "");
}

function languageRows(): { columns: number; rows: Map<string, string[]> } {
  const path = resolve(process.cwd(), "..", "backend", "assets", "language", "languages.tsv");
  const [headerLine = "", ...lines] = readFileSync(path, "utf8").split(/\r?\n/);
  const rows = new Map<string, string[]>();
  for (const line of lines) {
    if (!line || line.startsWith("#")) continue;
    const cells = line.split("\t");
    rows.set(cells[0] ?? "", cells);
  }
  return { columns: headerLine.split("\t").length, rows };
}

describe("cheque / bank screens language keys", () => {
  it("every key used has a full 13-column row with no empty cell", () => {
    const { columns, rows } = languageRows();
    expect(columns).toBe(13);
    const problems = new Set<string>();
    let used = 0;
    for (const file of sources) {
      const source = stripComments(readFileSync(file, "utf8"));
      for (const match of [...source.matchAll(keyedCall), ...source.matchAll(keyedTuple), ...source.matchAll(errorKey)]) {
        const key = match[1];
        if (!key) continue;
        used += 1;
        const cells = rows.get(key);
        if (!cells) problems.add(`${key}: no row`);
        else if (cells.length !== columns || cells.some((cell) => !cell.trim())) problems.add(`${key}: needs ${columns} non-empty cells`);
      }
    }
    expect(used).toBeGreaterThan(40);
    expect([...problems].sort()).toEqual([]);
  });

  it("keeps Thai only in the th column fallback of keyed calls, never loose in the code", () => {
    const loose: string[] = [];
    for (const file of sources) {
      const source = stripComments(readFileSync(file, "utf8")).replace(keyedCall, "").replace(keyedTuple, "");
      source.split("\n").forEach((line, index) => {
        if (thaiText.test(line)) loose.push(`${file.split(/[\\/]/).slice(-2).join("/")}:${index + 1}`);
      });
    }
    expect(loose).toEqual([]);
  });

  it("the Thai fallback in code equals the th column, so the screen reads the same before the dictionary loads", () => {
    const { rows } = languageRows();
    const drift: string[] = [];
    for (const file of sources) {
      const source = stripComments(readFileSync(file, "utf8"));
      for (const match of source.matchAll(/\btr\(\s*"([a-z0-9_]+)"\s*,\s*"((?:[^"\\]|\\.)*)"\s*,?\s*\)/g)) {
        const key = match[1] ?? "";
        const fallback = (match[2] ?? "").replace(/\\"/g, "\"");
        const th = rows.get(key)?.[1];
        // key เดิมของระบบ (save, cancel, …) มีคำไทยของตัวเองอยู่แล้ว — ตรวจเฉพาะ key ของงานนี้
        if (th !== undefined && /^(bank_code_|bank_contact_|cheque_|book_bank_)/.test(key) && th.replace(/[“”]/g, "\"") !== fallback) drift.push(key);
      }
    }
    expect(drift).toEqual([]);
  });
});
