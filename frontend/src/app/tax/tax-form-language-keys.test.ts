import { existsSync, readdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

// จอแบบยื่นกรมสรรพากร (tax-form-editor) — ข้อความบนจอมาจาก languages.tsv ทั้งหมด รวม key ที่ประกอบขึ้นตอนรัน
// (tax_form_title_<code>, tax_form_group_<group>) และ code/note ที่ backend ส่งกลับมา
const root = resolve(process.cwd(), "..");
const screens = ["tax-form-editor.tsx", "tax-form-fields.tsx", "tax-rdfile-panel.tsx"].map((name) => resolve(process.cwd(), "src", "app", "tax", name));
const specDir = resolve(root, "backend", "internal", "rdform", "specs");
const handlerDir = resolve(root, "backend", "internal", "goapi", "handlers");
// ไฟล์ยื่นกรมสรรพากร: key ของจุดที่ต้องแก้ (issues) มาจากตัวตรวจไฟล์และ handler — จอแปลด้วย tr(issue.key, …)
const rdFileDir = resolve(root, "backend", "internal", "rdfile");
const languageColumns = ["th", "en", "cn", "ja", "km", "ko", "lo", "my", "vi", "ms", "id", "fil"];

function languageRows(): Map<string, Record<string, string>> {
  const [headerLine = "", ...lines] = readFileSync(resolve(root, "backend", "assets", "language", "languages.tsv"), "utf8").split(/\r?\n/);
  const headers = headerLine.split("\t");
  const rows = new Map<string, Record<string, string>>();
  for (const line of lines) {
    if (!line) continue;
    const columns = line.split("\t");
    rows.set(columns[0] ?? "", Object.fromEntries(headers.map((header, index) => [header, columns[index] ?? ""])));
  }
  return rows;
}

function stripComments(source: string) {
  return source.replace(/\/\*[\s\S]*?\*\//g, "").replace(/^\s*\/\/.*$/gm, "");
}

function requiredKeys(): Set<string> {
  const keys = new Set<string>();
  for (const file of screens) {
    // tr("key", …) และข้อความสถานะ { key: "…", fallback: "…" } ที่แปลตอนแสดง
    for (const match of stripComments(readFileSync(file, "utf8")).matchAll(/\btr\(\s*"([a-z0-9_]+)"|\bkey: "([a-z0-9_]+)", fallback:/g)) keys.add(match[1] ?? match[2]);
  }
  const specs = readdirSync(specDir).filter((name) => name.endsWith(".json"));
  for (const name of specs) {
    const spec = JSON.parse(readFileSync(resolve(specDir, name), "utf8")) as { code?: string; fields?: { group?: string }[] };
    if (!name.includes("_attach")) keys.add(`tax_form_title_${name.replace(/\.json$/, "")}`);
    for (const field of spec.fields ?? []) keys.add(`tax_form_group_${field.group || "other"}`);
  }
  keys.add("tax_form_group_other");
  // ชื่อเดือนของตัวเลือกงวด: tr(`month_${MONTH_KEYS[m]}`)
  for (const month of ["january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"]) keys.add(`month_${month}`);
  for (const name of readdirSync(handlerDir).filter((n) => n.startsWith("tax_form") && n.endsWith(".go") && !n.endsWith("_test.go"))) {
    for (const match of readFileSync(resolve(handlerDir, name), "utf8").matchAll(/"(tax_form_[a-z0-9_]+)"/g)) keys.add(match[1]);
  }
  const rdSources = [
    ...readdirSync(handlerDir).filter((n) => n.startsWith("tax_rdfile")).map((n) => resolve(handlerDir, n)),
    ...(existsSync(rdFileDir) ? readdirSync(rdFileDir).map((n) => resolve(rdFileDir, n)) : []),
  ].filter((file) => file.endsWith(".go") && !file.endsWith("_test.go"));
  for (const file of rdSources) {
    for (const match of readFileSync(file, "utf8").matchAll(/"(tax_rdfile_[a-z0-9_]*[a-z0-9])"/g)) keys.add(match[1]);
  }
  return keys;
}

describe("tax form language keys", () => {
  it("ทุก key ของจอแบบยื่น + ชื่อแบบ + หมวด + code/note จาก backend มีแถวครบทุกภาษา", () => {
    const rows = languageRows();
    const missing: string[] = [];
    for (const key of requiredKeys()) {
      const row = rows.get(key);
      if (!row) { missing.push(`${key}:row`); continue; }
      for (const language of languageColumns) if (!row[language]?.trim()) missing.push(`${key}:${language}`);
    }
    expect(missing.sort()).toEqual([]);
  });

  it("จอแบบยื่นไม่มีข้อความไทยฝังนอก tr(key, fallback)", () => {
    const leftovers: string[] = [];
    for (const file of screens) {
      const source = stripComments(readFileSync(file, "utf8"))
        .replace(/\btr\(\s*"[a-z0-9_]+"\s*,\s*"(?:[^"\\]|\\.)*"\s*\)/g, "")
        .replace(/\bkey: "[a-z0-9_]+", fallback: "(?:[^"\\]|\\.)*"/g, "");
      source.split(/\r?\n/).forEach((line, index) => {
        if (/[฀-๿]/.test(line)) leftovers.push(`${file.split(/[\\/]/).pop()}:${index + 1}`);
      });
    }
    expect(leftovers).toEqual([]);
  });
});
