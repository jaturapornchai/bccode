import { readdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

// Guard for the AGENTS.md rule (2026-09-14): screen text must follow the selected
// language — GL code carries English keys only, the texts live in languages.tsv.
const glDir = resolve(process.cwd(), "src", "app", "gl");
const languageColumns = ["th", "en", "cn", "ja", "km", "ko", "lo", "my", "vi", "ms", "id", "fil"];
const glSources = [
  ...readdirSync(glDir).filter((name) => name.endsWith(".tsx")).map((name) => resolve(glDir, name)),
  resolve(process.cwd(), "src", "lib", "general-ledger.ts"),
  resolve(process.cwd(), "src", "lib", "gl-journal-details.ts"),
];

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
  return source.replace(/\/\*[\s\S]*?\*\//g, "").replace(/^\s*\/\/.*$/gm, "");
}

describe("general ledger language keys", () => {
  // vat_* = VAT detail hints (vat_ui_*), 2026-09-24; the claim-window hint reuses the backend error keys gl_err_vat_claim_*.
  it("every gl_* key used in GL code exists in languages.tsv with all language cells", () => {
    const rows = backendLanguageRows();
    const missing = new Set<string>();
    for (const file of glSources) {
      const source = stripComments(readFileSync(file, "utf8"));
      for (const match of source.matchAll(/["'`](gl_[a-z0-9_]+|common_[a-z0-9_]+|menu_[a-z0-9_]+|vat_[a-z0-9_]+)["'`]/g)) {
        const key = match[1];
        const row = rows.get(key);
        if (!row) { missing.add(`${key}:row`); continue; }
        for (const language of languageColumns) if (!row[language]?.trim()) missing.add(`${key}:${language}`);
      }
    }
    expect([...missing].sort()).toEqual([]);
  });

  it("GL screens carry no hard-coded Thai outside tr(key, fallback) calls", () => {
    // Financial statement starter templates in lib/general-ledger.ts are user-editable data, not UI labels.
    const screens = glSources.filter((file) => file.endsWith(".tsx"));
    const leftovers: string[] = [];
    for (const file of screens) {
      // Allowed carriers of Thai fallback text: tr("gl_key", "ไทย") calls and ["gl_key", "ไทย"] GLLabel tuples.
      // menu_* keys are the shared main-menu texts (planned-workflow card reused by GLPendingPanel, 2026-09-19).
      // wht_* keys are the shared 50 Tawi form/income/condition labels reused by the GL withholding details (2026-09-23).
      // vat_* keys are the VAT claim-window hints; out-of-window text reuses the backend error keys (2026-09-24).
      const source = stripComments(readFileSync(file, "utf8"))
        .replace(/\btr\(\s*"(?:gl|common|menu|wht|vat)_[a-z0-9_]+"\s*,\s*"(?:[^"\\]|\\.)*"\s*\)/g, "")
        .replace(/\[\s*"gl_[a-z0-9_]+"\s*,\s*"(?:[^"\\]|\\.)*"\s*\]/g, "");
      source.split(/\r?\n/).forEach((line, index) => {
        if (/[฀-๿]/.test(line)) leftovers.push(`${file.split(/[\\/]/).pop()}:${index + 1}`);
      });
    }
    expect(leftovers).toEqual([]);
  });

  // Champ wording (screens in D:\project-champ\bcspec\หน้าจอเมนูบัญชี) — 2026-10-07 label pass: the shared keys (gl_code, gl_name, gl_document_date, tax_register_col_date ...) keep their text.
  it("Champ labels of the group, mapping, budget, close and WHT-register screens are in languages.tsv (th = Champ wording, en = English)", () => {
    const champTh: Record<string, string> = {
      gl_group_code: "รหัสกลุ่ม", gl_group_name: "ชื่อกลุ่ม", gl_mapping_code: "รหัสรูปแบบ", gl_mapping_name: "ชื่อรูปแบบ",
      gl_inventory_account: "รหัสผังบัญชีสินค้า", gl_cogs_account: "รหัสผังบัญชีต้นทุน", gl_sales_revenue_account: "รหัสผังบัญชีรายได้",
      gl_accounts_in_rule: "รหัสผังบัญชี", gl_accounting_side_rule: "ประเภท", gl_amount_source_rule: "เงื่อนไข",
      gl_budget_list_name: "ชื่อ/รายละเอียด", gl_budget_list_amount: "จำนวน", gl_budget_line_account_code: "รหัสบัญชี",
      gl_close_document_date: "วันที่เอกสาร", gl_close_description: "คำอธิบาย", gl_close_verify: "ประมวลผล", gl_close_save_entry: "บันทึกรายการ",
      tax_register_col_payee: "ชื่อผู้มีเงินได้", tax_register_col_wht_date: "วันที่จ่าย", tax_register_col_wht_docno: "เลขที่เอกสาร",
      tax_register_col_wht_taxid: "เลขประจำตัวผู้เสียภาษี", tax_register_col_wht_base: "ฐานภาษี", tax_register_col_wht_tax: "ภาษีหัก ณ ที่จ่าย",
    };
    const rows = backendLanguageRows();
    for (const [key, th] of Object.entries(champTh)) {
      expect(rows.get(key)?.th, key).toBe(th);
      expect(rows.get(key)?.en ?? "", key).not.toMatch(/[\u0E00-\u0E7F]/);
    }
    // shared keys that other screens still use must not be re-worded by the Champ pass
    expect(rows.get("gl_code")?.th).toBe("รหัส");
    expect(rows.get("gl_name")?.th).toBe("ชื่อ");
    expect(rows.get("tax_register_col_date")?.th).toBe("วันที่");
    expect(rows.get("tax_register_col_base")?.th).toBe("มูลค่าก่อนภาษี");
    expect(rows.get("tax_register_col_tax")?.th).toBe("ภาษี");
  });
});
