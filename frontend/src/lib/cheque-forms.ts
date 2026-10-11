import type { BankCodeKind } from "./bank-code-kinds";
import type { BankBook, BankBookInput, BankCode, BankCodeInput, BankContact } from "./cheque-api";

// สถานะฟอร์มของจอกำหนดรหัสและจอผู้ติดต่อธนาคาร — แยกจาก React เพื่อทดสอบได้ (ไม่มีกฎบัญชี: ตรวจจริงที่ backend)

export type CodeDraft = { code: string; name: string; bank_code: string; user_group: string; is_active: boolean };

export const emptyCodeDraft = (): CodeDraft => ({ code: "", name: "", bank_code: "", user_group: "", is_active: true });

export const draftFromCode = (code: BankCode): CodeDraft => ({
  code: code.code,
  name: code.name,
  bank_code: code.bank_code ?? "",
  user_group: code.user_group ?? "",
  is_active: code.is_active,
});

const trimmed = (draft: CodeDraft): CodeDraft => ({ ...draft, code: draft.code.trim(), name: draft.name.trim(), bank_code: draft.bank_code.trim(), user_group: draft.user_group.trim() });

/** ฟอร์มถูกแก้จากค่าตั้งต้นหรือไม่ (เทียบหลังตัดช่องว่าง — พิมพ์เว้นวรรคเฉยๆ ไม่นับว่าแก้) */
export function isCodeDraftDirty(draft: CodeDraft, baseline: CodeDraft): boolean {
  const a = trimmed(draft);
  const b = trimmed(baseline);
  return a.code !== b.code || a.name !== b.name || a.bank_code !== b.bank_code || a.user_group !== b.user_group || a.is_active !== b.is_active;
}

/** ปุ่มบันทึกเปิดเมื่อมีรหัสกับชื่อ และมีการแก้จริง (กันบันทึกซ้ำโดยไม่ได้แก้) */
export function canSaveCodeDraft(draft: CodeDraft, baseline: CodeDraft): boolean {
  return draft.code.trim() !== "" && draft.name.trim() !== "" && isCodeDraftDirty(draft, baseline);
}

/** คำสั่งเขียนรหัส: ธนาคารของสาขาส่งเฉพาะชนิดสาขา (ชนิดอื่นไม่มีฟิลด์นี้) */
export function codeInputFromDraft(kind: BankCodeKind, draft: CodeDraft): BankCodeInput {
  const value = trimmed(draft);
  const input: BankCodeInput = { kind, code: value.code, name: value.name, user_group: value.user_group, is_active: value.is_active };
  if (kind === "bank_branch") input.bank_code = value.bank_code;
  return input;
}

const normalizeCode = (code: string) => code.trim().toUpperCase();

export const blankContact = (): BankContact => ({ contact_code: "", name: "", position: "", address: "", telephone: "", is_contact: true });

export const isBlankContact = (row: BankContact): boolean =>
  !row.contact_code.trim() && !row.name.trim() && !row.position.trim() && !row.address.trim() && !row.telephone.trim();

const trimContact = (row: BankContact): BankContact => ({
  contact_code: row.contact_code.trim(),
  name: row.name.trim(),
  position: row.position.trim(),
  address: row.address.trim(),
  telephone: row.telephone.trim(),
  is_contact: row.is_contact,
});

/** แถวที่จะส่งบันทึก: ตัดช่องว่าง และทิ้งแถวที่เพิ่มมาแล้วไม่ได้กรอกอะไร */
export function contactsToSave(rows: BankContact[]): BankContact[] {
  return rows.filter((row) => !isBlankContact(row)).map(trimContact);
}

export function areContactsDirty(rows: BankContact[], baseline: BankContact[]): boolean {
  return JSON.stringify(contactsToSave(rows)) !== JSON.stringify(contactsToSave(baseline));
}

/** รหัสผู้ติดต่อที่มีอยู่เดิมแต่ถูกเอาออกจากชุด — การบันทึกจะลบรายการเหล่านี้ จึงต้องขอยืนยันก่อน */
export function removedContactCodes(rows: BankContact[], baseline: BankContact[]): string[] {
  const kept = new Set(contactsToSave(rows).map((row) => normalizeCode(row.contact_code)));
  return contactsToSave(baseline)
    .map((row) => row.contact_code)
    .filter((code) => !kept.has(normalizeCode(code)));
}

/** รหัสผู้ติดต่อที่ซ้ำกันในชุด (ไม่แยกตัวพิมพ์ใหญ่เล็ก) — backend ปฏิเสธ จึงเตือนก่อนส่ง */
export function duplicateContactCode(rows: BankContact[]): string | null {
  const seen = new Set<string>();
  for (const row of contactsToSave(rows)) {
    const key = normalizeCode(row.contact_code);
    if (key && seen.has(key)) return row.contact_code;
    seen.add(key);
  }
  return null;
}

// ---- จอสมุดเงินฝาก ----

/** ความยาวสูงสุดตามฝั่ง backend (generalledger/bank_accounts.go: รหัสสมุด = AccountCodeMaxRunes, เลขที่บัญชี 50, ชื่อบัญชี 255) */
export const BANK_BOOK_LIMITS = { code: 20, accountNumber: 50, accountName: 255 } as const;

export type BookDraft = {
  bank_account_code: string;
  account_number: string;
  account_name: string;
  bank_code: string;
  bank_branch_code: string;
  account_type_code: string;
  gl_account_code: string;
  is_active: boolean;
};

export const emptyBookDraft = (): BookDraft => ({
  bank_account_code: "",
  account_number: "",
  account_name: "",
  bank_code: "",
  bank_branch_code: "",
  account_type_code: "",
  gl_account_code: "",
  is_active: true,
});

export const draftFromBook = (book: BankBook): BookDraft => ({
  bank_account_code: book.bank_account_code,
  account_number: book.account_number,
  account_name: book.account_name,
  bank_code: book.bank_code ?? "",
  bank_branch_code: book.bank_branch_code ?? "",
  account_type_code: book.account_type_code ?? "",
  gl_account_code: book.gl_account_code,
  is_active: book.is_active,
});

const trimBook = (draft: BookDraft): BookDraft => ({
  ...draft,
  bank_account_code: draft.bank_account_code.trim(),
  account_number: draft.account_number.trim(),
  account_name: draft.account_name.trim(),
  bank_code: draft.bank_code.trim(),
  bank_branch_code: draft.bank_branch_code.trim(),
  account_type_code: draft.account_type_code.trim(),
  gl_account_code: draft.gl_account_code.trim(),
});

export function isBookDraftDirty(draft: BookDraft, baseline: BookDraft): boolean {
  return JSON.stringify(trimBook(draft)) !== JSON.stringify(trimBook(baseline));
}

/** ปุ่มบันทึกเปิดเมื่อกรอกฟิลด์บังคับครบ (รหัสสมุด เลขที่บัญชี ชื่อบัญชี ธนาคาร บัญชีผังบัญชี) และมีการแก้จริง */
export function canSaveBookDraft(draft: BookDraft, baseline: BookDraft): boolean {
  const value = trimBook(draft);
  const complete = [value.bank_account_code, value.account_number, value.account_name, value.bank_code, value.gl_account_code].every(Boolean);
  return complete && isBookDraftDirty(draft, baseline);
}

/** สมุดเงินฝากรับเฉพาะบาท (THB) จนกว่าจะตัดสินเรื่องสกุลเงินต่างประเทศ */
export function bookInputFromDraft(draft: BookDraft): BankBookInput {
  return { ...trimBook(draft), currency_code: "THB" };
}

/**
 * แสดงจำนวนเงินจากข้อความทศนิยมของ backend โดยไม่ผ่าน number: คั่นหลักพัน ขั้นต่ำ minFraction หลัก
 * และไม่ปัดเศษ (ถ้ามีทศนิยมจริงเกิน minFraction หลักก็แสดงครบ); ข้อความที่ไม่ใช่ตัวเลข = ข้อความว่าง
 */
export function formatDecimalText(value: string | undefined, minFraction = 2): string {
  const match = /^(-?)(\d+)(?:\.(\d+))?$/.exec((value ?? "").trim());
  if (!match) return "";
  const [, sign = "", integer = "0", fraction = ""] = match;
  const decimals = fraction.replace(/0+$/, "").padEnd(minFraction, "0");
  const grouped = integer.replace(/^0+(?=\d)/, "").replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  const isZero = !/[1-9]/.test(integer + decimals);
  return `${isZero ? "" : sign}${grouped}${decimals ? `.${decimals}` : ""}`;
}
