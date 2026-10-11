import { apiFetch } from "./client-auth-session";
import type { BankCodeKind } from "./bank-code-kinds";
import type { LanguageCode } from "./i18n";

// ไคลเอนต์ของโมดูลเช็ค/ธนาคาร → BFF /api/cheque/* → backend /cheque/v1/*
// ข้อความผิดพลาดมาจาก backend ตามภาษาของคำขอ (message) และ code = key ใน languages.tsv

export type BankCode = {
  kind: BankCodeKind;
  code: string;
  name: string;
  bank_code: string;
  user_group: string;
  is_active: boolean;
  version: number;
};

export type BankCodePage = { items: BankCode[]; total: number; page: number; limit: number };

export type BankContact = {
  contact_code: string;
  name: string;
  position: string;
  address: string;
  telephone: string;
  is_contact: boolean;
};

export class ChequeApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly field: string;

  constructor(status: number, code: string, message: string, field = "") {
    super(message);
    this.name = "ChequeApiError";
    this.status = status;
    this.code = code;
    this.field = field;
  }
}

const isRecord = (value: unknown): value is Record<string, unknown> => typeof value === "object" && value !== null && !Array.isArray(value);

function toError(httpStatus: number, payload: unknown): ChequeApiError {
  const body = isRecord(payload) ? payload : {};
  // BFF ส่งความผิดของผู้ใช้เป็น HTTP 200 + success:false จึงใช้ statuscode ใน body ก่อน
  const status = typeof body.statuscode === "number" ? body.statuscode : httpStatus;
  const code = typeof body.code === "string" && body.code ? body.code : typeof body.errorcode === "string" && body.errorcode ? body.errorcode : "";
  const message = typeof body.message === "string" ? body.message : "";
  return new ChequeApiError(status, code, message, typeof body.field === "string" ? body.field : "");
}

async function request<T>(path: string, language: LanguageCode, init?: RequestInit): Promise<T> {
  const response = await apiFetch(path, {
    ...init,
    headers: { "Accept-Language": language, ...(init?.body ? { "Content-Type": "application/json" } : {}), ...init?.headers },
  });
  let payload: unknown = null;
  try {
    payload = await response.json();
  } catch {
    payload = null;
  }
  if (response.ok && isRecord(payload) && payload.success !== false) return payload as T;
  throw toError(response.status, payload);
}

function query(params: Record<string, string | number | boolean | undefined>): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === "" || value === false) continue;
    search.set(key, value === true ? "1" : String(value));
  }
  const text = search.toString();
  return text ? `?${text}` : "";
}

export type BankCodeListQuery = { q?: string; bankCode?: string; activeOnly?: boolean; page?: number; limit?: number };

export function listBankCodes(kind: BankCodeKind, filter: BankCodeListQuery, language: LanguageCode): Promise<BankCodePage> {
  const path = `/api/cheque/codes/${kind}${query({ q: filter.q?.trim(), bank_code: filter.bankCode, active: filter.activeOnly, page: filter.page, limit: filter.limit })}`;
  return request<BankCodePage>(path, language);
}

export type BankCodeInput = { kind: BankCodeKind; code: string; name: string; bank_code?: string; user_group?: string; is_active?: boolean };

function sendCodeCommand(action: "create" | "update" | "delete", body: Record<string, unknown>, language: LanguageCode): Promise<BankCode> {
  return request<{ data: BankCode }>("/api/cheque/command", language, {
    method: "POST",
    body: JSON.stringify({ resource: "codes", action, ...body }),
  }).then((result) => result.data);
}

export const createBankCode = (input: BankCodeInput, language: LanguageCode) => sendCodeCommand("create", input, language);
export const updateBankCode = (input: BankCodeInput & { version: number }, language: LanguageCode) => sendCodeCommand("update", input, language);
export const deleteBankCode = (target: { kind: BankCodeKind; code: string; version: number }, language: LanguageCode) =>
  sendCodeCommand("delete", target, language);

export async function listBankContacts(bankCode: string, branchCode: string, language: LanguageCode): Promise<BankContact[]> {
  const result = await request<{ items: BankContact[] }>(
    `/api/cheque/bank-contacts${query({ bank_code: bankCode, bank_branch_code: branchCode })}`,
    language,
  );
  return result.items ?? [];
}

/** บันทึกผู้ติดต่อ "ทั้งชุด" ของ ธนาคาร + สาขา (ชุดว่าง = ล้างทั้งสาขา) */
export async function saveBankContacts(bankCode: string, branchCode: string, contacts: BankContact[], language: LanguageCode): Promise<BankContact[]> {
  const result = await request<{ items: BankContact[] }>("/api/cheque/command", language, {
    method: "POST",
    body: JSON.stringify({ resource: "bank-contacts", action: "save", bank_code: bankCode, bank_branch_code: branchCode, contacts }),
  });
  return result.items ?? [];
}

// ---- สมุดเงินฝาก: GL resource "bank-accounts" ผ่าน BFF /api/gl/* (backend generalledger/bank_accounts.go) ----

export type BankBook = {
  bank_account_code: string;
  bank_name: string;
  account_number: string;
  account_name: string;
  gl_account_code: string;
  currency_code: string;
  is_active: boolean;
  version: number;
  bank_code?: string;
  bank_branch_code?: string;
  account_type_code?: string;
  /** ยอดคงเหลือของบัญชี GL ที่ผูกกับสมุด — ข้อความทศนิยมจาก backend ห้ามแปลงเป็น number */
  gl_balance?: string;
};

export type BankBookPage = { items: BankBook[]; total: number; page: number; limit: number };

export type BankBookInput = {
  bank_account_code: string;
  account_number: string;
  account_name: string;
  bank_code: string;
  bank_branch_code: string;
  account_type_code: string;
  gl_account_code: string;
  currency_code: "THB";
  is_active: boolean;
};

export type BankGLAccount = { code: string; name: string };

export async function listBankBooks(filter: { q?: string; page?: number; limit?: number }, language: LanguageCode): Promise<BankBookPage> {
  const result = await request<{ data: BankBookPage }>(`/api/gl/bank-accounts${query({ q: filter.q?.trim(), page: filter.page, limit: filter.limit })}`, language);
  return result.data;
}

type GLAccountRow = { accountcode: string; names?: { code: string; name: string }[]; accounttype: string; allowposting: boolean; isactive: boolean };

/** บัญชีผังบัญชีที่สมุดเงินฝากผูกได้: สินทรัพย์ เปิดใช้งาน และลงรายการได้ (backend ตรวจซ้ำตอนบันทึก) — เรียงตามรหัส บัญชีสินทรัพย์มาก่อน */
export async function listBankGLAccounts(language: LanguageCode): Promise<BankGLAccount[]> {
  const result = await request<{ data: { items?: GLAccountRow[] } }>("/api/gl/accounts?limit=1000", language);
  return (result.data.items ?? [])
    .filter((account) => account.isactive && account.allowposting && account.accounttype === "asset")
    .map((account) => ({
      code: account.accountcode,
      name: account.names?.find((name) => name.code === language)?.name || account.names?.find((name) => name.code === "th")?.name || "",
    }));
}

/** UUID สำหรับ requestid ของคำสั่ง GL (BFF รับเฉพาะรูป UUID); randomUUID ใช้ไม่ได้บน HTTP ธรรมดา จึงมีทางสำรองด้วย getRandomValues */
export function newRequestId(): string {
  if (typeof crypto.randomUUID === "function") return crypto.randomUUID();
  const bytes = crypto.getRandomValues(new Uint8Array(16));
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join("");
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

function sendBookCommand(action: string, body: Record<string, unknown>, language: LanguageCode): Promise<{ id: string; version: number }> {
  return request<{ data: { id: string; version: number } }>("/api/gl/command", language, {
    method: "POST",
    body: JSON.stringify({ resource: "bank-accounts", action, requestid: newRequestId(), ...body }),
  }).then((result) => result.data);
}

export const createBankBook = (book: BankBookInput, language: LanguageCode) => sendBookCommand("create", { bankaccount: book }, language);
export const updateBankBook = (book: BankBookInput, version: number, language: LanguageCode) =>
  sendBookCommand("update", { id: book.bank_account_code, version, bankaccount: book }, language);
export const deleteBankBook = (code: string, version: number, language: LanguageCode) => sendBookCommand("delete", { id: code, version }, language);
/** ปุ่ม "เปลี่ยนเลขที่บัญชี": แก้เฉพาะเลขที่บัญชีธนาคาร รหัสสมุดและเอกสารที่อ้างสมุดไม่เปลี่ยน */
export const changeBankBookNumber = (code: string, version: number, newNumber: string, language: LanguageCode) =>
  sendBookCommand("change-account-number", { id: code, version, newaccountnumber: newNumber }, language);

/**
 * ข้อความผิดพลาดที่จะแสดงผู้ใช้: ข้อความจาก backend ก่อน (ตรงภาษาและแทนค่าให้แล้ว);
 * ความผิดที่ BFF สร้างเอง (message = code) หรือเชื่อมต่อไม่ได้ ใช้ข้อความจาก dictionary ตาม code
 */
export function chequeErrorText(error: unknown, tr: (key: string, fallback: string) => string): string {
  if (error instanceof ChequeApiError) {
    if (error.message && error.message !== error.code) return error.message;
    if (error.code) return tr(error.code, tr("cheque_err_internal", "ระบบขัดข้อง ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง ถ้ายังไม่หายให้แจ้งผู้ดูแลระบบ"));
  }
  return tr("cheque_request_failed", "เชื่อมต่อระบบไม่สำเร็จ กรุณาตรวจสอบอินเทอร์เน็ตแล้วลองใหม่อีกครั้ง");
}
