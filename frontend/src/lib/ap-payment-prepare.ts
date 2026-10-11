import { apiFetch } from "./client-auth-session";
import { erpDecimalSign, erpSignedSum, type ErpDetailItem } from "./erp-transaction";

// ใบเตรียมจ่ายชำระหนี้ (Champ "บันทึกเตรียมจ่ายชำระหนี้") — โครงตรงกับ JSON ของ
// backend/internal/transaction/httpapi/ap_payment_prepare.go; ยอดเงินทุกช่องเป็นข้อความทศนิยม (Rule 20)

export const AP_PREPARE_API = "/api/ap-payment-prepare";

export type PrepareResult<T> = { ok: true; data: T; total: number } | { ok: false; message: string; field?: string };

/**
 * เรียก BFF ของใบเตรียมจ่าย: ข้อความปฏิเสธเป็นภาษาที่เลือกบนจอ (Accept-Language) — ข้อผิดพลาดทางธุรกิจ BFF ตอบ 200 + success:false
 * message ว่าง = เชื่อมต่อไม่ได้/ไม่มีข้อความ (จอแสดงข้อความกลางเอง)
 */
export async function prepareRequest<T>(path: string, language: string, init?: RequestInit): Promise<PrepareResult<T>> {
  try {
    const response = await apiFetch(`${AP_PREPARE_API}${path}`, {
      ...init,
      headers: { "Accept-Language": language, ...(init?.body ? { "Content-Type": "application/json" } : {}) },
    });
    const payload = (await response.json().catch(() => null)) as
      | { success?: boolean; data?: T; message?: unknown; field?: unknown; pagination?: { total?: number } }
      | null;
    if (response.ok && payload && payload.success !== false) {
      return { ok: true, data: payload.data as T, total: payload.pagination?.total ?? 0 };
    }
    return {
      ok: false,
      message: typeof payload?.message === "string" ? payload.message : "",
      field: typeof payload?.field === "string" ? payload.field : undefined,
    };
  } catch (error) {
    return { ok: false, message: error instanceof Error ? error.message : "" };
  }
}

/** บรรทัดของใบที่บันทึกแล้ว: payamount = ยอดตัดจ่าย, used_amount = ใบจ่ายใช้ไปแล้ว, open_amount = ยังดึงไปจ่ายได้ */
export type ApPrepareLine = {
  linenumber: number;
  debt_document_id: string;
  document_no: string;
  document_date: string;
  debt_amount: string;
  billtype: number;
  billtype_key: string;
  sign: number;
  description: string;
  invbalance: string;
  payamount: string;
  used_amount: string;
  open_amount: string;
};

export type ApPrepareDoc = {
  docno: string;
  paydate: string;
  partner_code: string;
  partner_name: string;
  branch_code: string;
  remark: string;
  projectcode: string;
  allocatecode: string;
  billgroupcode: string;
  departmentcode: string;
  total_amount: string;
  total_home_amount: string;
  discount_amount: string;
  net_amount: string;
  exchange_profit: string;
  billstatus: number;
  iscancel: boolean;
  cancel_reason: string;
  version: number;
  lines: ApPrepareLine[] | null;
};

export type ApPrepareListRow = Pick<
  ApPrepareDoc,
  "docno" | "paydate" | "partner_code" | "partner_name" | "total_amount" | "discount_amount" | "net_amount" | "billstatus" | "iscancel" | "version"
>;

/** เอกสารหนี้ค้างจ่าย: available_amount = คงเหลือ − ยอดที่ใบเตรียมจ่ายอื่นจองไว้ */
export type ApOutstandingRow = {
  debt_document_id: string;
  document_no: string;
  document_date: string;
  due_date: string;
  billtype: number;
  billtype_key: string;
  sign: number;
  description: string;
  debt_amount: string;
  remaining_amount: string;
  reserved_amount: string;
  available_amount: string;
};

export type ApCreditor = { partner_code: string; name: string; credit_days: number };

/** บรรทัดบนจอ: invbalance = มูลค่าคงเหลือ (อ่านอย่างเดียว), used_amount = ยอดที่ใบจ่ายใช้แล้ว (ลดต่ำกว่านี้ไม่ได้) */
export type ApPrepareFormLine = {
  debt_document_id: string;
  document_no: string;
  document_date: string;
  description: string;
  billtype: number;
  billtype_key: string;
  sign: number;
  invbalance: string;
  payamount: string;
  used_amount: string;
};

export type ApPrepareForm = {
  docno: string;
  paydate: string;
  partner_code: string;
  partner_name: string;
  remark: string;
  projectcode: string;
  departmentcode: string;
  discount_amount: string;
  version: number;
  lines: ApPrepareFormLine[];
};

/** คำขอบันทึก/แก้ (apPrepareInput) — ไม่มีช่องการจัดสรร/กลุ่มของบิล (D8: เก็บค่าว่างจนกว่า master จะเสร็จ) */
export type ApPrepareInput = {
  docno: string;
  paydate: string;
  partner_code: string;
  remark: string;
  projectcode: string;
  departmentcode: string;
  discount_amount: string;
  version: number;
  lines: { debt_document_id: string; payamount: string; description: string }[];
};

// รูปแบบเดียวกับ apPrepareAmountPattern ของ backend: ไม่ติดลบ ทศนิยมไม่เกิน 2 ตำแหน่ง
const AMOUNT_TEXT = /^[0-9]{1,16}(\.[0-9]{1,2})?$/;

export function isPrepareAmountText(value: string): boolean {
  return AMOUNT_TEXT.test(value.trim());
}

export function emptyPrepareForm(today: string): ApPrepareForm {
  return {
    docno: "",
    paydate: today,
    partner_code: "",
    partner_name: "",
    remark: "",
    projectcode: "",
    departmentcode: "",
    discount_amount: "",
    version: 0,
    lines: [],
  };
}

export function prepareFormFromDoc(doc: ApPrepareDoc): ApPrepareForm {
  return {
    docno: doc.docno,
    paydate: doc.paydate,
    partner_code: doc.partner_code,
    partner_name: doc.partner_name,
    remark: doc.remark ?? "",
    projectcode: doc.projectcode ?? "",
    departmentcode: doc.departmentcode ?? "",
    discount_amount: erpDecimalSign(doc.discount_amount) === 0 ? "" : doc.discount_amount,
    version: doc.version,
    lines: (doc.lines ?? []).map((line) => ({
      debt_document_id: line.debt_document_id,
      document_no: line.document_no,
      document_date: line.document_date,
      description: line.description ?? "",
      billtype: line.billtype,
      billtype_key: line.billtype_key,
      sign: line.sign,
      invbalance: line.invbalance,
      payamount: line.payamount,
      used_amount: line.used_amount,
    })),
  };
}

/** เลือกเอกสารค้างจ่ายเข้าบรรทัด: ยอดตัดจ่ายตั้งต้น = ยอดที่จ่ายได้ (Champ InvBalance = PayAmount = PayBalance) */
export function prepareLineFromOutstanding(row: ApOutstandingRow): ApPrepareFormLine {
  return {
    debt_document_id: row.debt_document_id,
    document_no: row.document_no,
    document_date: row.document_date,
    description: row.description ?? "",
    billtype: row.billtype,
    billtype_key: row.billtype_key,
    sign: row.sign,
    invbalance: row.available_amount,
    payamount: row.available_amount,
    used_amount: "0",
  };
}

/** ยอดรวมตัวอย่างระหว่างแก้ (ยอดจริงคือที่ backend ตอบหลังบันทึก): รวม = Σ sign × ยอดตัดจ่าย, ยอดจ่าย = รวม − ส่วนลด */
export function prepareTotals(lines: readonly Pick<ApPrepareFormLine, "payamount" | "sign">[], discount: string): { total: string; net: string } {
  const total = erpSignedSum(lines.map((line) => ({ amount: line.payamount, sign: line.sign })));
  return { total, net: erpSignedSum([{ amount: total, sign: 1 }, { amount: discount, sign: -1 }]) };
}

export function preparePayload(form: ApPrepareForm): ApPrepareInput {
  return {
    docno: form.docno.trim(),
    paydate: form.paydate,
    partner_code: form.partner_code.trim(),
    remark: form.remark,
    projectcode: form.projectcode.trim(),
    departmentcode: form.departmentcode.trim(),
    discount_amount: form.discount_amount.trim(),
    version: form.version,
    lines: form.lines.map((line) => ({
      debt_document_id: line.debt_document_id,
      payamount: line.payamount.trim(),
      description: line.description,
    })),
  };
}

export type ApPrepareStatus = "cancelled" | "paid" | "unpaid";

export function prepareStatus(doc: Pick<ApPrepareDoc, "iscancel" | "billstatus">): ApPrepareStatus {
  if (doc.iscancel) return "cancelled";
  return doc.billstatus === 1 ? "paid" : "unpaid";
}

/**
 * Ctrl+F3 บนใบจ่ายชำระหนี้: บรรทัดของใบเตรียมจ่ายที่ยังเหลือยอด (open_amount > 0) เป็นบรรทัดใบจ่าย
 * (design §4.5 ข้อ 1: docref = เลขที่เอกสารหนี้, debt_document_id, paybillno = เลขที่ใบเตรียมจ่าย, ยอด = open_amount; reftype backend ตั้งให้).
 * บรรทัดลดหนี้ (sign −1) ไม่ดึง เพราะใบจ่ายตัดหนี้ได้เฉพาะยอดบวกต่อบรรทัด (settleLines) — ผู้ใช้หักในยอดจ่ายเอง
 * ข้ามหนี้ที่มีบรรทัดจากใบเตรียมจ่ายนี้อยู่แล้ว (ดึงซ้ำไม่เพิ่มซ้ำ)
 */
export function payLinesFromPrepare(doc: ApPrepareDoc, existing: readonly ErpDetailItem[]): ErpDetailItem[] {
  const taken = new Set(existing.filter((item) => item.paybillno === doc.docno).map((item) => item.debt_document_id));
  return (doc.lines ?? [])
    .filter((line) => line.sign > 0 && erpDecimalSign(line.open_amount) > 0 && !taken.has(line.debt_document_id))
    .map((line, index) => ({
      linenumber: existing.length + index + 1,
      itemcode: line.document_no,
      itemname: line.description || line.document_no,
      unitcode: "",
      qty: "1",
      price: line.open_amount,
      sumamount: line.open_amount,
      docref: line.document_no,
      debt_document_id: line.debt_document_id,
      paybillno: doc.docno,
    }));
}
