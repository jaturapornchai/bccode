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

// ---- ทะเบียนเช็ค (registers) และเอกสารเปลี่ยนสถานะ 419-426 (docs): backend cheque/register.go, cheque/docs.go ----
// เงินทุกช่องเป็นสตริงทศนิยมจาก backend ห้ามแปลงเป็น number

export type ChequePage<T> = { items: T[]; total: number; page: number; limit: number };

export type ChequeRegister = {
  id: string;
  cheque_type: number;
  doc_no: string;
  doc_date: string;
  due_date: string;
  partner_code: string;
  partner_name: string;
  cheque_no: string;
  bank_code: string;
  bank_branch_code: string;
  book_code: string;
  status: number;
  ext_status_code: string;
  currency_code: string;
  exchange_rate: string;
  amount: string;
  home_amount: string;
  balance_amount: string;
  description: string;
  sale_code: string;
  department_code: string;
  source_kind: string;
  prepare_doc_no: string;
  print_count: number;
  version: number;
  created_at: string;
  created_by: string;
  updated_at: string;
  updated_by: string;
};

/** แถวในตารางเอกสารอ้างอิงใต้หัวเช็ค; kind_key = key ของ doctype (หรือชนิดต้นทางถ้าไม่มี doctype) */
export type ChequeReference = { kind_key: string; doc_no: string; doc_date: string; amount: string; description: string };

export type ChequeEvent = {
  id: number;
  event_kind: string;
  doc_flag?: number;
  doc_no?: string;
  event_date: string;
  from_status: number | null;
  to_status: number | null;
  book_code: string;
  reason: string;
  actor: string;
  created_at: string;
};

export type ChequeRegisterDetail = ChequeRegister & { last_event: ChequeEvent | null; references: ChequeReference[] };

export type ChequeRegisterQuery = {
  type: number;
  status?: number | null;
  partner?: string;
  book?: string;
  dueFrom?: string;
  dueTo?: string;
  q?: string;
  page?: number;
  limit?: number;
};

export function listChequeRegisters(filter: ChequeRegisterQuery, language: LanguageCode): Promise<ChequePage<ChequeRegister>> {
  const path = `/api/cheque/registers${query({
    type: filter.type,
    status: filter.status ?? undefined,
    partner: filter.partner?.trim(),
    book: filter.book,
    due_from: filter.dueFrom,
    due_to: filter.dueTo,
    q: filter.q?.trim(),
    page: filter.page,
    limit: filter.limit,
  })}`;
  return request<ChequePage<ChequeRegister>>(path, language);
}

export async function getChequeRegister(id: string, language: LanguageCode): Promise<ChequeRegisterDetail> {
  const result = await request<{ data: ChequeRegisterDetail }>(`/api/cheque/registers/${encodeURIComponent(id)}`, language);
  return { ...result.data, references: result.data.references ?? [] };
}

export type ChequeSelectableQuery = {
  transflag: number;
  docDate?: string;
  book?: string;
  cancelType?: number | null;
  cancelPassed?: boolean;
  q?: string;
  page?: number;
  limit?: number;
};

/** เช็คที่เอกสารสถานะชนิดนั้นเลือกได้ ณ วันที่เอกสาร (backend ตรวจซ้ำตอนบันทึก) */
export function listSelectableCheques(filter: ChequeSelectableQuery, language: LanguageCode): Promise<ChequePage<ChequeRegister>> {
  const path = `/api/cheque/registers/selectable${query({
    transflag: filter.transflag,
    doc_date: filter.docDate,
    book: filter.book,
    cancel_type: filter.cancelType ?? undefined,
    cancel_passed: filter.cancelPassed,
    q: filter.q?.trim(),
    page: filter.page,
    limit: filter.limit,
  })}`;
  return request<ChequePage<ChequeRegister>>(path, language);
}

/** ค่าที่ผู้ใช้กรอกในจอทะเบียน (สกุลเงินบาท อัตรา 1 เท่านั้นจนกว่าจะตัดสิน D15) */
export type ChequeRegisterInput = {
  cheque_type: number;
  doc_no: string;
  doc_date: string;
  due_date: string;
  partner_code: string;
  cheque_no: string;
  bank_code: string;
  bank_branch_code: string;
  book_code: string;
  status: number;
  ext_status_code: string;
  currency_code: "THB";
  exchange_rate: "1";
  amount: string;
  description: string;
  sale_code: string;
  department_code: string;
};

/** ทุกคำสั่งเขียนของเช็คใช้ requestid ใหม่ต่อการกดหนึ่งครั้ง (ส่งซ้ำหลัง timeout ได้ผลเดิม ไม่เขียนซ้ำ) */
function sendChequeCommand<T>(resource: string, action: string, body: Record<string, unknown>, language: LanguageCode): Promise<T> {
  return request<{ data: T }>("/api/cheque/command", language, {
    method: "POST",
    body: JSON.stringify({ resource, action, requestid: newRequestId(), ...body }),
  }).then((result) => result.data);
}

export const createChequeRegister = (input: ChequeRegisterInput, language: LanguageCode) =>
  sendChequeCommand<ChequeRegister>("registers", "create", input, language);

/** confirm = ผู้ใช้ยืนยันแล้ว (D8): แก้ฟิลด์หลักของเช็คที่ถูกอ้างแล้ว หรือแก้สถานะตรงในจอทะเบียน */
export const updateChequeRegister = (input: ChequeRegisterInput & { id: string; version: number }, confirm: boolean, language: LanguageCode) =>
  sendChequeCommand<ChequeRegister>("registers", "update", confirm ? { ...input, confirm: true } : input, language);

/** ลบเช็คออกจากทะเบียน (เก็บประวัติไว้); ถูกเอกสารใช้อยู่หรือมาจากเอกสารต้นทาง backend ปฏิเสธ */
export const cancelChequeRegister = (target: { chequeType: number; id: string; version: number }, language: LanguageCode) =>
  sendChequeCommand<ChequeRegister>("registers", "cancel", { cheque_type: target.chequeType, id: target.id, version: target.version }, language);

export type ChequeDocLine = {
  line_no: number;
  cheque_id: string;
  cheque_no: string;
  partner_code: string;
  partner_name: string;
  bank_code: string;
  bank_branch_code: string;
  cheque_date: string;
  due_date: string;
  description: string;
  from_status: number;
  amount: string;
  fee_amount: string;
  net_amount: string;
};

/** gl_status: none / posted / no_mapping / not_required; status: active / cancelled */
export type ChequeDoc = {
  transflag: number;
  doc_no: string;
  doc_date: string;
  branch_code: string;
  book_code: string;
  cancel_type: number | null;
  cancel_passed: boolean;
  description: string;
  total_amount: string;
  fee_amount: string;
  net_amount: string;
  gl_status: string;
  journal_docno: string;
  status: string;
  cancel_reason: string;
  cancelled_at: string | null;
  cancelled_by: string;
  version: number;
  created_at: string;
  created_by: string;
  updated_at: string;
  updated_by: string;
  lines: ChequeDocLine[];
};

export type ChequeDocQuery = { q?: string; status?: "active" | "cancelled"; dateFrom?: string; dateTo?: string; page?: number; limit?: number };

export function listChequeDocs(flag: number, filter: ChequeDocQuery, language: LanguageCode): Promise<ChequePage<ChequeDoc>> {
  const path = `/api/cheque/docs/${flag}${query({ q: filter.q?.trim(), status: filter.status, date_from: filter.dateFrom, date_to: filter.dateTo, page: filter.page, limit: filter.limit })}`;
  return request<ChequePage<ChequeDoc>>(path, language);
}

export async function getChequeDoc(flag: number, docNo: string, language: LanguageCode): Promise<ChequeDoc> {
  const result = await request<{ data: ChequeDoc }>(`/api/cheque/docs/${flag}/${encodeURIComponent(docNo)}`, language);
  return { ...result.data, lines: result.data.lines ?? [] };
}

export type ChequeDocInput = {
  transflag: number;
  doc_no: string;
  doc_date: string;
  book_code: string;
  /** ใช้เฉพาะ 423 (ชนิดอื่นต้องไม่ส่ง) */
  cancel_type?: number;
  /** ใช้เฉพาะ 426 */
  cancel_passed?: boolean;
  description: string;
  lines: { cheque_id: string; fee_amount: string }[];
  /** ยอดท้ายเอกสารที่หน้าจอรวมมา — backend คำนวณซ้ำแล้วต้องตรงกัน */
  totals: { total: string; fee: string; net: string };
};

export const createChequeDoc = (input: ChequeDocInput, language: LanguageCode) => sendChequeCommand<ChequeDoc>("docs", "create", input, language);

/** ยกเลิกเอกสาร: ต้องมีเหตุผล; เช็คกลับสถานะเดิมและใบสำคัญ GL ถูกกลับรายการ */
export const cancelChequeDoc = (target: { transflag: number; docNo: string; version: number; reason: string }, language: LanguageCode) =>
  sendChequeCommand<ChequeDoc>("docs", "cancel", { transflag: target.transflag, doc_no: target.docNo, version: target.version, reason: target.reason }, language);

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

// ---- กำหนดวงเงินสดย่อย (Petty Cash Limit; Champ 5177 CASHFrmPettyCash) ----

export type PettyCashItem = {
  code: string;
  name: string;
  amount: string;
  begin_amount: string;
  balance: string;
  version?: number;
};

export async function listPettyCash(language: LanguageCode): Promise<PettyCashItem[]> {
  const result = await request<{ items: PettyCashItem[] }>("/api/cheque/petty-cash", language);
  return result.items ?? [];
}

export async function savePettyCash(items: PettyCashItem[], language: LanguageCode): Promise<PettyCashItem[]> {
  const result = await request<{ items: PettyCashItem[] }>("/api/cheque/command", language, {
    method: "POST",
    body: JSON.stringify({ resource: "petty-cash", action: "save", petty_cash_items: items }),
  });
  return result.items ?? [];
}

