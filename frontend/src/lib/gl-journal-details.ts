import { amountUnits } from "./general-ledger";
/** คู่ค้า — title_name/addr_district/addr_province/addr_postcode ไม่บังคับ (ไฟล์ยื่นด้วยสื่อ Format กลาง: คำนำหน้าทุกแบบ, ที่อยู่แยกช่องใน ภ.ง.ด.3) */
export type GLDetailPartner = {partner_code:string;name_th:string;title_name?:string;tax_id?:string;tax_branch_no?:string;address?:string;addr_district?:string;addr_province?:string;addr_postcode?:string;is_customer:boolean;is_supplier:boolean;is_active:boolean;version?:number};
/** ความยาวสูงสุด (ตัวอักษร) ของช่องคู่ค้าที่ไม่บังคับ — ตรงกับที่ backend ตรวจ (C:/tmp/rdfile/spec.md §3a) */
export const PARTNER_TEXT_LIMITS = {title_name: 100, addr_district: 50, addr_province: 50} as const;
export type GLDetailBankAccount = {bank_account_code:string;bank_name:string;account_number:string;account_name:string;gl_account_code:string;currency_code:string;is_active:boolean;version?:number};
export type GLDetailDocument = {id:string;ledger:string;partner_code:string;document_no:string;document_date:string;due_date?:string;branch_code:string;document_kind:number;balance_side:number;amount:string;currency_code:string;control_account_code:string;version?:number};
export type GLDetailAllocation = {id:string;ledger:string;document_id:string;journal_id?:string;line_number:number;amount:string};
export type GLDetailSettlement = {id:string;ledger:string;partner_code:string;debt_document_id:string;payment_document_id:string;settlement_date:string;amount:string};
export type GLDetailBankLine = {journal_id?:string;line_number:number;bank_account_code:string;direction:number};
export type GLDetailStatement = {id:string;bank_account_code:string;source_key:string;transaction_date:string;value_date?:string;bank_reference?:string;description?:string;direction:number;amount:string;balance_after?:string};
export type GLDetailMatch = {id:string;statement_line_id:string;journal_id?:string;line_number:number;amount:string};
export type GLDetailWithdrawal = {kind:string;id:string;reason:string};
/** ภาษีหัก ณ ที่จ่ายประกอบใบสำคัญ (ตาม mydocs wht.sql) — ฐานภาษีแก้ได้เสมอ; tax_amount ว่าง = backend คำนวณ ฐาน × อัตรา */
export type GLDetailWithholding = {id:string;wht_direction:number;form_type:string;partner_code:string;wht_cert_no?:string;certificate_date?:string;payment_date:string;income_tax_type:string;income_description?:string;condition_type:number;wht_rate:string;base_amount:string;tax_amount?:string;
  /** snapshot ผู้จ่าย(ผู้หัก)/ผู้รับเงิน(ผู้ถูกหัก) ณ วันบันทึก (wht.sql) — ไม่บังคับ; เลขภาษี 13 หลัก, สาขา 5 หลัก, ว่าง = ใช้ทะเบียนตอนพิมพ์ */
  payer_tax_id?:string;payer_branch_no?:string;payer_name?:string;payer_address?:string;
  payee_tax_id?:string;payee_branch_no?:string;payee_name?:string;payee_address?:string;
  /** หมายเหตุ ≤ 500 ตัวอักษร · เล่มที่ของหนังสือรับรอง 50 ทวิ */
  remark?:string;wht_book_no?:string};
/** ภาษีมูลค่าเพิ่มประกอบใบสำคัญ (ตาม mydocs vat.sql) — 1 แถว = 1 ใบกำกับภาษี; ฐานภาษีแก้ได้เสมอ; vat_amount ว่าง = backend คำนวณ ฐาน × อัตรา
 *  tax_type 1=ซื้อ 2=ขาย · document_type 1=ใบกำกับ 2=ใบเพิ่มหนี้ 3=ใบลดหนี้ · claim_status (เฉพาะซื้อ) 1=ใช้สิทธิ 2=ต้องห้าม 3=รอใช้สิทธิ 4=ไม่ใช้สิทธิ */
export type GLDetailVat = {id:string;tax_type:number;document_type:number;tax_invoice_no:string;tax_invoice_date:string;original_invoice_no?:string;original_invoice_date?:string;tax_period_year?:number;tax_period_month?:number;partner_code?:string;partner_tax_id?:string;partner_branch_no?:string;partner_name:string;base_amount:string;zero_rate_amount:string;exempt_amount:string;vat_rate:string;vat_amount?:string;claim_status?:number;claim_reason?:string;remark?:string};
export type GLJournalDetails = {partners?:GLDetailPartner[];bank_accounts?:GLDetailBankAccount[];documents?:GLDetailDocument[];allocations?:GLDetailAllocation[];settlements?:GLDetailSettlement[];bank_lines?:GLDetailBankLine[];statement_lines?:GLDetailStatement[];matches?:GLDetailMatch[];withdrawals?:GLDetailWithdrawal[];withholdings?:GLDetailWithholding[];vats?:GLDetailVat[]};
export type GLSupportRow = Record<string,string|number|boolean|undefined>;
export type GLSupportKind = "partners"|"bank-accounts"|"documents"|"statements"|"bank-lines"|"allocations"|"settlements"|"matches";
// ข้อความทุกคำผ่าน key (ภาษาตามที่เลือก); ไม่ส่ง tr = ใช้ข้อความไทยสำรอง
export function supportLabel(kind: GLSupportKind, row: GLSupportRow, tr: (key: string, fallback: string) => string = (_key, fallback) => fallback): string {
  const remaining = row.remaining_amount === undefined ? "" : ` · ${tr("gl_support_remaining", "คงเหลือ {0}").replace("{0}", String(row.remaining_amount))}`;
  if(kind === "partners") return `${row.partner_code ?? ""} · ${row.name_th ?? ""}`;
  if(kind === "bank-accounts") return `${row.bank_account_code ?? ""} · ${row.bank_name ?? ""} ${row.account_number ?? ""}`;
  if(kind === "documents") return `${row.ledger === "ap" ? tr("gl_support_ledger_ap", "เจ้าหนี้") : tr("gl_support_ledger_ar", "ลูกหนี้")} · ${row.document_no ?? ""} · ${row.partner_code ?? ""} · ${row.amount ?? ""}${remaining}`;
  if(kind === "statements") return `${row.transaction_date ?? ""} · ${row.bank_account_code ?? ""} · ${row.bank_reference || row.description || row.source_key || ""} · ${row.amount ?? ""}${remaining}`;
  if(kind === "bank-lines") return `${row.docno ?? tr("gl_support_this_journal", "ใบสำคัญนี้")} · ${tr("gl_line_nth", "บรรทัดที่ {0}").replace("{0}", String(row.line_number ?? ""))} · ${row.bank_account_code ?? ""} · ${row.amount ?? ""}${remaining}`;
  if(kind === "settlements") return `${row.debt_document_no ?? tr("gl_support_bill", "บิล")} ↔ ${row.payment_document_no ?? tr("gl_support_payment", "ผลชำระ")} · ${row.amount ?? ""}`;
  return `${row.document_no || row.docno || row.bank_reference || row.partner_code || tr("gl_support_item", "รายการ")} · ${row.settlement_date || row.created_at || ""} · ${row.amount ?? ""}`;
}
export function reconciliationChanges(before: GLJournalDetails = {}, after: GLJournalDetails = {}): GLJournalDetails {
  const result: GLJournalDetails = {};
  for(const key of ["allocations","statement_lines","settlements","matches","withdrawals"] as const) {
    const identity = (row: {id:string;kind?:string}) => key === "withdrawals" ? `${row.kind}:${row.id}` : row.id;
    const oldIDs = new Set((before[key] ?? []).map(identity));
    const rows = (after[key] ?? []).filter(row => !oldIDs.has(identity(row)));
    Object.assign(result, {[key]: rows});
  }
  // ภาษีหัก ณ ที่จ่าย/ภาษีมูลค่าเพิ่มส่งทั้งชุดเมื่อมีการแก้ (backend แทนทั้งชุดและเก็บค่าเดิมใน audit)
  // ชุดว่าง [] = ล้างทุกแถวของใบที่ผ่านบัญชีแล้ว; ไม่ส่ง key = ไม่เปลี่ยน
  for(const key of ["withholdings","vats"] as const) {
    if(JSON.stringify(before[key] ?? []) !== JSON.stringify(after[key] ?? [])) Object.assign(result, {[key]: after[key] ?? []});
  }
  return result;
}
/** ประเภทภาษีเริ่มต้นตามประเภทสมุด (booktype) ไม่ใช่รหัสสมุด: 4 = สมุดรายวันขาย → ภาษีขาย (2); อื่น ๆ → ภาษีซื้อ (1) ผู้ใช้เปลี่ยนได้ */
export function defaultVatTaxType(booktype?: number): 1 | 2 { return booktype === 4 ? 2 : 1; }

/** ตัวเลขล้วนจากเลขภาษี/เลขสาขาที่พิมพ์หรือวาง เช่น "0-1055-12345-67-8" → "0105512345678" (เลขไทยแปลงเป็นอารบิก, ไม่ตัดความยาว) */
export function taxDigits(value: unknown): string {
  return String(value ?? "").replace(/[\u0E50-\u0E59]/g, digit => String(digit.charCodeAt(0) - 0x0e50)).replace(/\D/g, "");
}
/** เลขสาขาเก็บ 5 หลัก เติม 0 ข้างหน้า ("0" → "00000"); ว่างคงว่าง; เกิน 5 หลักคืนตามที่พิมพ์ให้ตรวจแจ้งผู้ใช้ */
export function normalizeBranchNo(value: unknown): string {
  const digits = taxDigits(value);
  return digits && digits.length <= 5 ? digits.padStart(5, "0") : digits;
}

/** เลขประจำตัวผู้เสียภาษี 13 หลักผ่านหลักตรวจสอบ: หลักที่ 13 = (11 − (Σ หลักที่ p × (14 − p), p = 1..12) mod 11) mod 10
 *  — สูตรเดียวกับ backend whtcert.ValidThaiTaxID (ผู้ตัดสินตอนบันทึก); จอใช้บอกผู้ใช้ทันทีที่พิมพ์ครบ 13 หลัก */
export function validThaiTaxId(value: unknown): boolean {
  const digits = taxDigits(value);
  if(digits.length !== 13) return false;
  let sum = 0;
  for(let i = 0; i < 12; i++) sum += Number(digits[i]) * (13 - i);
  return (11 - (sum % 11)) % 10 === Number(digits[12]);
}
const TAX_ID_FIELDS = ["partner_tax_id","payer_tax_id","payee_tax_id","tax_id"];
/** แถวแรกที่ช่องเลขภาษี field (ชื่อช่องที่ backend ส่งมากับ error เช่น tax_id, partner_tax_id) ผิดรูปแบบหรือหลักตรวจสอบ
 *  — backend บอกช่องแต่ไม่บอกแถว จอจึงหาแถวเองเพื่อพาไปที่ช่องนั้น (UAT S3 2026-09-24); row นับจาก 1 เหมือน GLDetailProblem */
export function detailTaxIdTarget(details: GLJournalDetails|undefined, field: string): {section: "withholdings"|"vats"|"partners"; row: number; field: string}|null {
  if(!TAX_ID_FIELDS.includes(field)) return null;
  for(const section of ["partners","vats","withholdings"] as const) {
    const index = ((details?.[section] ?? []) as Record<string, unknown>[]).findIndex(row => taxDigits(row[field]) !== "" && !validThaiTaxId(row[field]));
    if(index >= 0) return {section, row: index + 1, field};
  }
  return null;
}
const BRANCH_FIELDS = ["partner_branch_no","payer_branch_no","payee_branch_no","tax_branch_no"];
const ZERO_WHEN_BLANK = ["base_amount","zero_rate_amount","exempt_amount","amount"];
const OMIT_WHEN_BLANK = ["tax_amount","vat_amount","balance_after","wht_rate","vat_rate"];
// ช่องคู่ค้าที่ไม่บังคับ: ตัดช่องว่างหัวท้าย ว่าง = ไม่ส่ง; รหัสไปรษณีย์เหลือตัวเลขล้วน (เลขไทยแปลงเป็นอารบิก)
const OPTIONAL_TEXT = ["title_name","addr_district","addr_province"];
const POSTCODE_FIELDS = ["addr_postcode"];
function normalizeDetailRow<T extends object>(row: T): T {
  const next: Record<string, unknown> = {...(row as Record<string, unknown>)};
  for(const [field, current] of Object.entries(next)) {
    const blank = current === undefined || current === null || String(current).trim() === "";
    if(TAX_ID_FIELDS.includes(field)) next[field] = taxDigits(current);
    else if(BRANCH_FIELDS.includes(field)) next[field] = normalizeBranchNo(current);
    else if(blank && ZERO_WHEN_BLANK.includes(field)) next[field] = "0";
    else if(blank && OMIT_WHEN_BLANK.includes(field)) delete next[field];
    else if(OPTIONAL_TEXT.includes(field) || POSTCODE_FIELDS.includes(field)) {
      const value = POSTCODE_FIELDS.includes(field) ? taxDigits(current) : String(current ?? "").trim();
      if(value) next[field] = value; else delete next[field];
    }
  }
  return next as T;
}
/** เตรียมรายละเอียดก่อนส่ง backend: ช่องเงินว่าง → "0" (backend ไม่รับ ""), ยอดภาษี/อัตราว่าง → ไม่ส่ง (ยอดภาษีว่าง = คำนวณให้),
 *  เลขภาษีเหลือตัวเลขล้วน, เลขสาขาเติม 0 เป็น 5 หลัก — ไม่เปลี่ยนค่าที่ผู้ใช้พิมพ์ครบแล้ว */
export function normalizeJournalDetails(details?: GLJournalDetails): GLJournalDetails|undefined {
  if(!details) return details;
  const result: GLJournalDetails = {...details};
  for(const key of Object.keys(result) as (keyof GLJournalDetails)[]) {
    const rows = result[key];
    if(Array.isArray(rows)) Object.assign(result, {[key]: (rows as object[]).map(normalizeDetailRow)});
  }
  return result;
}

export type GLDetailProblem = {section: "withholdings"|"vats"|"partners"; row: number; field: string; message: string};
/** ตรวจรายละเอียดภาษีก่อนบันทึก บอกหมวด แถว ช่อง และวิธีแก้ — backend ตรวจซ้ำเสมอ */
export function journalDetailsProblem(details: GLJournalDetails|undefined, tr: (key: string, fallback: string) => string): GLDetailProblem|null {
  const sectionName = {
    withholdings: tr("gl_details_section_withholdings", "ภาษีหัก ณ ที่จ่าย"),
    vats: tr("gl_details_section_vats", "ภาษีมูลค่าเพิ่ม (ใบกำกับภาษี)"),
    // noun for the message prefix — the partners screen heading ("เพิ่มคู่ค้า…") reads as an instruction here
    partners: tr("gl_details_error_section_partners", "คู่ค้า"),
  };
  const fieldName: Record<string, string> = {
    partner_tax_id: tr("gl_detail_vats_partner_tax_id", "เลขประจำตัวผู้เสียภาษี (13 หลัก)"),
    partner_branch_no: tr("gl_detail_vats_partner_branch_no", "สาขา (5 หลัก, 00000 = สำนักงานใหญ่)"),
    tax_id: tr("gl_detail_partners_tax_id", "เลขผู้เสียภาษี"),
    tax_branch_no: tr("gl_detail_partners_tax_branch_no", "สาขาภาษี"),
    payer_tax_id: tr("wht_cert_ui_payer_taxid", "เลขประจำตัวผู้เสียภาษีอากรผู้มีหน้าที่หักภาษี (13 หลัก)"),
    payee_tax_id: tr("wht_cert_ui_payee_taxid", "เลขประจำตัวผู้เสียภาษีอากรผู้ถูกหัก (13 หลัก)"),
    payer_branch_no: tr("wht_cert_ui_payer_branch", "สาขาของผู้มีหน้าที่หักภาษี (5 หลัก, 00000 = สำนักงานใหญ่)"),
    payee_branch_no: tr("wht_cert_ui_payee_branch", "สาขาของผู้ถูกหักภาษี (5 หลัก, 00000 = สำนักงานใหญ่)"),
    title_name: tr("gl_detail_partners_title_name", "คำนำหน้าชื่อ (ไม่บังคับ เช่น นาย, บริษัท)"),
    addr_district: tr("gl_detail_partners_addr_district", "อำเภอ/เขต (ไม่บังคับ)"),
    addr_province: tr("gl_detail_partners_addr_province", "จังหวัด (ไม่บังคับ)"),
    addr_postcode: tr("gl_detail_partners_addr_postcode", "รหัสไปรษณีย์ (ไม่บังคับ, 5 หลัก)"),
  };
  const problem = (section: GLDetailProblem["section"], index: number, field: string, key: string, fallback: string, ...values: (string|number)[]) => ({
    section, row: index + 1, field,
    message: values.reduce<string>((text, value, position) => text.replace(`{${position + 2}}`, String(value)),
      tr(key, fallback).replace("{0}", sectionName[section]).replace("{1}", String(index + 1))),
  });
  const checkParty = (section: GLDetailProblem["section"], index: number, row: Record<string, unknown>) => {
    for(const field of TAX_ID_FIELDS) {
      const digits = taxDigits(row[field]);
      if(digits && digits.length !== 13) return problem(section, index, field, "gl_detail_tax_id_invalid", "{0} แถวที่ {1}: ช่อง {2} ต้องเป็นตัวเลข 13 หลัก (ตอนนี้ {3} หลัก) — วางแบบมีขีดได้ ระบบตัดขีดให้เอง", fieldName[field], digits.length);
    }
    for(const field of BRANCH_FIELDS) {
      if(taxDigits(row[field]).length > 5) return problem(section, index, field, "gl_detail_branch_invalid", "{0} แถวที่ {1}: ช่อง {2} ต้องเป็นตัวเลขไม่เกิน 5 หลัก เช่น 0 หรือ 00000 = สำนักงานใหญ่", fieldName[field]);
    }
    return null;
  };
  for(const [index, row] of (details?.withholdings ?? []).entries()) {
    if(!String(row.wht_rate ?? "").trim()) return problem("withholdings", index, "wht_rate", "gl_detail_wht_rate_required", "{0} แถวที่ {1}: กรุณาใส่อัตราภาษี (%) เช่น 1, 3 หรือ 5 — ระบบไม่ใส่ 0% ให้เอง");
    const party = checkParty("withholdings", index, row);
    if(party) return party;
    if([...(row.remark ?? "")].length > 500) return problem("withholdings", index, "remark", "gl_detail_remark_too_long", "{0} แถวที่ {1}: หมายเหตุต้องไม่เกิน 500 ตัวอักษร (ตอนนี้ {2}) — กรุณาย่อข้อความ", [...(row.remark ?? "")].length);
  }
  for(const [index, row] of (details?.vats ?? []).entries()) {
    if(!String(row.vat_rate ?? "").trim()) return problem("vats", index, "vat_rate", "gl_detail_vat_rate_required", "{0} แถวที่ {1}: กรุณาใส่อัตราภาษี (%) เช่น 7 หรือ 0 สำหรับอัตราร้อยละ 0");
    const party = checkParty("vats", index, row);
    if(party) return party;
  }
  for(const [index, row] of (details?.partners ?? []).entries()) {
    const party = checkParty("partners", index, row);
    if(party) return party;
    const postcode = taxDigits(row.addr_postcode);
    if(postcode && postcode.length !== 5) return problem("partners", index, "addr_postcode", "gl_detail_postcode_invalid", "{0} แถวที่ {1}: ช่อง {2} ต้องเป็นตัวเลข 5 หลัก (ตอนนี้ {3} หลัก) หรือเว้นว่าง", fieldName.addr_postcode, postcode.length);
    for(const [field, limit] of Object.entries(PARTNER_TEXT_LIMITS)) {
      const length = [...String(row[field as keyof typeof PARTNER_TEXT_LIMITS] ?? "").trim()].length;
      if(length > limit) return problem("partners", index, field, "gl_detail_text_too_long", "{0} แถวที่ {1}: ช่อง {2} ยาวเกิน {3} ตัวอักษร (ตอนนี้ {4}) — กรุณาย่อข้อความ", fieldName[field], limit, length);
    }
  }
  return null;
}
/** งวดภาษีของแถวเป็นค่า "YYYY-MM" (ค.ศ.) — ว่าง = ยังไม่กำหนดงวด */
export function vatPeriodValue(row: {tax_period_year?: unknown; tax_period_month?: unknown}): string {
  const year = Number(row.tax_period_year), month = Number(row.tax_period_month);
  return Number.isInteger(year) && year >= 1900 && Number.isInteger(month) && month >= 1 && month <= 12 ? `${year}-${String(month).padStart(2, "0")}` : "";
}
/** "YYYY-MM" → ปี/เดือนของ payload; ว่าง = ไม่มีงวดทั้งคู่ (vat.sql: ว่างทั้งคู่หรือระบุทั้งคู่) */
export function vatPeriodPatch(value: string): {tax_period_year?: number; tax_period_month?: number} {
  const match = /^(\d{4})-(\d{2})$/.exec(value);
  return match ? {tax_period_year: Number(match[1]), tax_period_month: Number(match[2])} : {tax_period_year: undefined, tax_period_month: undefined};
}
/** เดือนที่ใช้สิทธิภาษีซื้อเทียบเดือนที่ออกใบกำกับ — ม.82/3 วรรคสอง (https://www.rd.go.th/5206.html) ให้ใช้สิทธิภายหลังได้ตามที่อธิบดีกำหนด;
 *  ประกาศอธิบดีฯ เกี่ยวกับ VAT ฉบับที่ 4 ข้อ 2 แก้โดยฉบับที่ 76 (https://www.rd.go.th/3417.html): ไม่เกิน 6 เดือนนับแต่เดือนถัดจากเดือนที่ออกใบกำกับ
 *  และต้องเขียน "ถือเป็นภาษีซื้อในเดือนภาษี ..." ในใบกำกับ → ออกเดือน M ใช้สิทธิได้งวด M..M+6 (ม.ค. → ก.พ.–ก.ค. คือเลื่อน)
 *  ตรวจเฉพาะภาษีซื้อที่ใช้สิทธิ (claim_status 1) ตรงกับ checkPurchaseClaimWindow ของ backend (ผู้ตัดสินตอนบันทึก):
 *  ใบกำกับ/ใบเพิ่มหนี้ (ม.77/1(22) นับใบเพิ่มหนี้เป็นใบกำกับ) ใช้ช่วง 6 เดือน; ใบลดหนี้ (document_type 3) ต้องลดภาษีซื้อในเดือนที่ได้รับ
 *  (ม.82/10 วรรคท้าย https://www.rd.go.th/5206.html) → ไม่มีเพดาน 6 เดือน ตรวจแค่งวดต้องไม่ก่อนเดือนของใบลดหนี้ (credit_note_before)
 *  ปีงวด ≥ 2400 = กรอกเป็น พ.ศ. (buddhist_year) ทุกประเภทแถว เหมือน backend vat_period_year_buddhist — ไม่แปลงปีให้เอง
 *  คำแนะนำให้เขียนข้อความในใบกำกับ (late) บอกเฉพาะใบกำกับภาษี (document_type 1) ที่ประกาศข้างต้นพูดถึง
 *  คิดจากเลขปี/เดือนล้วน ไม่ใช้ Date (ไม่มีปัญหาเขตเวลา) */
export const VAT_CLAIM_MAX_LATE_MONTHS = 6;
export type VatClaimTiming = {kind:"late";months:number;periodYear:number;periodMonth:number}|{kind:"before"|"exceeded"|"credit_note_before"|"buddhist_year"};
export function vatClaimTiming(row: {tax_type?: unknown; document_type?: unknown; claim_status?: unknown; tax_invoice_date?: unknown; tax_period_year?: unknown; tax_period_month?: unknown}): VatClaimTiming|null {
  if(Number(row.tax_period_year) >= 2400) return {kind: "buddhist_year"};
  if(Number(row.tax_type) !== 1 || Number(row.claim_status) !== 1) return null;
  const invoice = /^(\d{4})-(\d{2})-\d{2}$/.exec(String(row.tax_invoice_date ?? ""));
  const period = vatPeriodValue(row);
  if(!invoice || !period) return null;
  const invoiceMonth = Number(invoice[2]);
  if(invoiceMonth < 1 || invoiceMonth > 12) return null;
  const [periodYear, periodMonth] = period.split("-").map(Number);
  const months = periodYear * 12 + periodMonth - (Number(invoice[1]) * 12 + invoiceMonth);
  if(Number(row.document_type) === 3) return months < 0 ? {kind: "credit_note_before"} : null;
  if(months < 0) return {kind: "before"};
  if(months > VAT_CLAIM_MAX_LATE_MONTHS) return {kind: "exceeded"};
  return months === 0 || Number(row.document_type) !== 1 ? null : {kind: "late", months, periodYear, periodMonth};
}
/** ตัวเลือกงวดภาษีรอบวันที่ใบสำคัญ: ย้อนหลัง 12 เดือนถึงล่วงหน้า 6 เดือน (ภาษีซื้อใช้สิทธิภายหลังได้) + งวดที่บันทึกไว้เดิมเสมอ */
export function vatPeriodChoices(date: string, current = ""): string[] {
  const match = /^(\d{4})-(\d{2})/.exec(date);
  const today = new Date();
  const baseYear = match ? Number(match[1]) : today.getFullYear(), baseMonth = match ? Number(match[2]) : today.getMonth() + 1;
  const choices: string[] = [];
  for(let offset = -12; offset <= 6; offset++) {
    const index = baseYear * 12 + baseMonth - 1 + offset;
    choices.push(`${Math.floor(index / 12)}-${String(index % 12 + 1).padStart(2, "0")}`);
  }
  if(current && !choices.includes(current)) choices.push(current);
  return choices.sort();
}
/** CSV supports quoted commas/newlines; source identity uses SHA-256 of the original file + row. */
export async function parseStatementCsv(text:string, bankAccountCode:string):Promise<GLDetailStatement[]> {
  if(!bankAccountCode) throw new Error("เลือกบัญชีธนาคารก่อนนำเข้า");
  const rows:string[][]=[]; let row:string[]=[], cell="", quoted=false;
  const source=text.replace(/^\uFEFF/, "");
  for(let i=0;i<source.length;i++) {const ch=source[i]; if(ch==='"') {if(quoted && source[i+1]==='"') {cell+='"';i++;} else quoted=!quoted;} else if(!quoted && (ch===',' || ch==='\n')) {row.push(cell);cell="";if(ch==='\n') {rows.push(row);row=[];}} else if(ch!=='\r' || quoted) cell+=ch;}
  if(quoted) throw new Error("เครื่องหมายคำพูดใน CSV ไม่ครบ");
  if(cell || row.length) {row.push(cell);rows.push(row);}
  const headers=rows.shift()?.map(value=>value.trim()) ?? [];
  const required=["transaction_date","direction","amount"];
  if(required.some(name=>!headers.includes(name))) throw new Error("CSV ต้องมี transaction_date,direction,amount และเลือกใส่ bank_reference,description ได้");
  if(rows.length>1000) throw new Error("นำเข้าได้ไม่เกิน 1,000 รายการต่อครั้ง");
  const hash=Array.from(new Uint8Array(await crypto.subtle.digest("SHA-256",new TextEncoder().encode(text))),b=>b.toString(16).padStart(2,"0")).join("");
  const bankHash=Array.from(new Uint8Array(await crypto.subtle.digest("SHA-256",new TextEncoder().encode(bankAccountCode+":"+hash))),b=>b.toString(16).padStart(2,"0")).join("");
  return rows.flatMap((values,index)=>{
    if(!values.some(value=>value.trim())) return [];
    const value=(key:string)=>values[headers.indexOf(key)]?.trim() ?? "";
    const date=value("transaction_date"), amount=value("amount"), direction=value("direction");
    if(!/^\d{4}-\d{2}-\d{2}$/.test(date) || !["1","2"].includes(direction) || amountUnits(amount)<=0n) throw new Error(`ตรวจวันที่ ทิศทาง หรือยอดเงินแถว ${index+2}`);
    const valueDate=value("value_date"), balanceAfter=value("balance_after");
    if(valueDate && !/^\d{4}-\d{2}-\d{2}$/.test(valueDate)) throw new Error(`ตรวจวันที่มีผลแถว ${index+2}`);
    if(balanceAfter) amountUnits(balanceAfter);
    const hex=bankHash.slice(0,24)+(index+2).toString(16).padStart(8,"0");
    const id=`${hex.slice(0,8)}-${hex.slice(8,12)}-5${hex.slice(13,16)}-a${hex.slice(17,20)}-${hex.slice(20)}`;
    return {id,bank_account_code:bankAccountCode,source_key:`sha256:${hash}:${index+2}`,transaction_date:date,direction:Number(direction),amount,bank_reference:value("bank_reference"),description:value("description"),...(valueDate?{value_date:valueDate}:{}),...(balanceAfter?{balance_after:balanceAfter}:{})};
  });
}
