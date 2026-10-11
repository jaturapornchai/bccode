import type { ChequeDoc, ChequeDocInput, ChequeRegister, ChequeRegisterInput } from "./cheque-api";
import { CHEQUE_TYPE_ISSUED, CHEQUE_TYPE_RECEIVED, chequeDocBookPolicy, type ChequeBookPolicy, type ChequeDocConfig } from "./cheque-doc-config";
import { centsToMoney, isMoneyText, moneyToCents, subtractMoney, sumMoney } from "./cheque-forms";
import { formatDefaultDateTime } from "./date-time";
import type { LanguageCode } from "./i18n";

// สถานะฟอร์มของจอทะเบียนเช็คและจอเอกสารเปลี่ยนสถานะ — แยกจาก React เพื่อทดสอบได้
// ไม่มีกฎบัญชี: หน้าจอรวมยอดเพื่อแสดงผลและส่งให้ backend ตรวจซ้ำ (ไม่ตรงกัน = ปฏิเสธ); กติกาจริงอยู่ที่ backend cheque/*.go

/** ความยาวสูงสุดตามคอลัมน์ฝั่ง backend (cheque/register.go) */
export const CHEQUE_LIMITS = { docNo: 30, chequeNo: 20, description: 255, refCode: 50 } as const;

/** ช่องกรอกเงิน: รับเฉพาะตัวเลขกับจุดทศนิยมไม่เกิน 2 ตำแหน่ง (ตัดจุลภาคและช่องว่างที่วางมา); ค่าที่พิมพ์แล้วผิดรูปแบบ = คงค่าเดิม */
export function moneyInput(next: string, previous: string): string {
  const cleaned = next.replace(/[,\s]/g, "");
  return /^\d{0,16}(?:\.\d{0,2})?$/.test(cleaned) ? cleaned : previous;
}

/** "ผู้ใช้ — วันเวลา" ของการสร้าง/แก้ไขล่าสุด (เวลา Asia/Bangkok, พ.ศ. เมื่อเป็นภาษาไทย); ไม่มีข้อมูล = ข้อความว่าง */
export function chequeStamp(language: LanguageCode, by: string | undefined, at: string | undefined): string {
  const when = at ? formatDefaultDateTime(at, { language, yearType: language === "th" ? "buddhist" : "christian" }) : "";
  return [by ?? "", when].filter(Boolean).join(" — ");
}

// ---- ทะเบียนเช็ค ----

export type RegisterDraft = {
  doc_no: string;
  doc_date: string;
  partner_code: string;
  due_date: string;
  status: number;
  cheque_no: string;
  ext_status_code: string;
  bank_code: string;
  bank_branch_code: string;
  book_code: string;
  amount: string;
  description: string;
  sale_code: string;
  department_code: string;
};

/** เช็คใหม่: วันที่เอกสารและวันที่ครบกำหนดเป็นวันนี้ สถานะ "เช็คในมือ" (ค่าเริ่มต้นตาม Champ) */
export const emptyRegisterDraft = (today: string): RegisterDraft => ({
  doc_no: "",
  doc_date: today,
  partner_code: "",
  due_date: today,
  status: 0,
  cheque_no: "",
  ext_status_code: "",
  bank_code: "",
  bank_branch_code: "",
  book_code: "",
  amount: "",
  description: "",
  sale_code: "",
  department_code: "",
});

export const draftFromRegister = (register: ChequeRegister): RegisterDraft => ({
  doc_no: register.doc_no,
  doc_date: register.doc_date,
  partner_code: register.partner_code,
  due_date: register.due_date,
  status: register.status,
  cheque_no: register.cheque_no,
  ext_status_code: register.ext_status_code ?? "",
  bank_code: register.bank_code,
  bank_branch_code: register.bank_branch_code,
  book_code: register.book_code ?? "",
  amount: register.amount,
  description: register.description ?? "",
  sale_code: register.sale_code ?? "",
  department_code: register.department_code ?? "",
});

const trimRegister = (draft: RegisterDraft): RegisterDraft => ({
  ...draft,
  doc_no: draft.doc_no.trim(),
  partner_code: draft.partner_code.trim(),
  cheque_no: draft.cheque_no.trim(),
  amount: draft.amount.trim(),
  description: draft.description.trim(),
  sale_code: draft.sale_code.trim(),
  department_code: draft.department_code.trim(),
});

/** ฟอร์มถูกแก้จากค่าตั้งต้นหรือไม่ (เทียบหลังตัดช่องว่าง — พิมพ์เว้นวรรคเฉยๆ ไม่นับว่าแก้) */
export function isRegisterDraftDirty(draft: RegisterDraft, baseline: RegisterDraft): boolean {
  return JSON.stringify(trimRegister(draft)) !== JSON.stringify(trimRegister(baseline));
}

/** สมุดเงินฝากบังคับเมื่อ: เช็คจ่ายเสมอ / เช็ครับที่สถานะไม่ใช่ "เช็คในมือ" (0), "เช็คยกเลิก" (4) และ "ขายลด" (5: ขายลดไม่บังคับสมุด) — backend ตรวจซ้ำ */
export function registerBookRequired(chequeType: number, status: number): boolean {
  if (chequeType === CHEQUE_TYPE_ISSUED) return true;
  return status !== 0 && status !== 4 && status !== 5;
}

/** ชื่อช่องบังคับที่ยังว่างหรือผิดรูปแบบ เรียงตามลำดับช่องบนจอ (ว่าง = ครบ) */
export function registerMissingFields(chequeType: number, draft: RegisterDraft): string[] {
  const value = trimRegister(draft);
  const missing: string[] = [];
  if (!value.doc_date) missing.push("doc_date");
  if (!value.partner_code) missing.push("partner_code");
  if (!value.due_date) missing.push("due_date");
  if (!value.cheque_no) missing.push("cheque_no");
  if (!value.bank_code) missing.push("bank_code");
  if (!value.bank_branch_code) missing.push("bank_branch_code");
  if (registerBookRequired(chequeType, value.status) && !value.book_code.trim()) missing.push("book_code");
  const cents = moneyToCents(value.amount);
  if (cents === null || cents <= BigInt(0)) missing.push("amount");
  return missing;
}

/** ปุ่มบันทึกเปิดเมื่อกรอกช่องบังคับครบและมีการแก้จริง (กันบันทึกซ้ำโดยไม่ได้แก้) */
export function canSaveRegisterDraft(chequeType: number, draft: RegisterDraft, baseline: RegisterDraft): boolean {
  return registerMissingFields(chequeType, draft).length === 0 && isRegisterDraftDirty(draft, baseline);
}

/** จำนวนเงินบาท = จำนวนเงิน × อัตรา (อัตราเป็น 1 เสมอจนกว่าจะตัดสิน D15); จำนวนเงินยังไม่ถูกต้อง = ข้อความว่าง */
export function registerHomeAmount(amount: string): string {
  const cents = moneyToCents(amount);
  return cents === null ? "" : centsToMoney(cents);
}

/** คำสั่งเขียนทะเบียน: ตัดช่องว่าง บาท อัตรา 1; พนักงานขายมีเฉพาะเช็ครับ */
export function registerInputFromDraft(chequeType: number, draft: RegisterDraft): ChequeRegisterInput {
  const value = trimRegister(draft);
  return {
    cheque_type: chequeType,
    doc_no: value.doc_no,
    doc_date: value.doc_date,
    due_date: value.due_date,
    partner_code: value.partner_code,
    cheque_no: value.cheque_no,
    bank_code: value.bank_code,
    bank_branch_code: value.bank_branch_code,
    book_code: value.book_code,
    status: value.status,
    ext_status_code: value.ext_status_code,
    currency_code: "THB",
    exchange_rate: "1",
    amount: value.amount,
    description: value.description,
    sale_code: chequeType === CHEQUE_TYPE_RECEIVED ? value.sale_code : "",
    department_code: value.department_code,
  };
}

// ---- เอกสารเปลี่ยนสถานะ 419-426 ----

export type DocLineDraft = {
  cheque_id: string;
  cheque_no: string;
  partner_name: string;
  bank_code: string;
  bank_branch_code: string;
  due_date: string;
  /** จำนวนเงินบนเช็ค (ตามสกุลเงินของเช็ค) */
  cheque_amount: string;
  /** จำนวนเงินบาท — ฐานที่ใช้ตัดค่าใช้จ่าย */
  amount: string;
  /** คำอธิบายของทะเบียนเช็ค (คอลัมน์หมายเหตุของจอนำฝาก) */
  description: string;
  fee_amount: string;
};

export type DocDraft = {
  doc_no: string;
  doc_date: string;
  book_code: string;
  /** ใช้เฉพาะ 423; ค่าตั้งต้น 0 = ยกเลิกเช็คในมือ */
  cancel_type: number;
  /** ใช้เฉพาะ 426 */
  cancel_passed: boolean;
  description: string;
  lines: DocLineDraft[];
};

export const emptyDocDraft = (today: string): DocDraft => ({
  doc_no: "",
  doc_date: today,
  book_code: "",
  cancel_type: 0,
  cancel_passed: false,
  description: "",
  lines: [],
});

export const docLineFromRegister = (register: ChequeRegister): DocLineDraft => ({
  cheque_id: register.id,
  cheque_no: register.cheque_no,
  partner_name: register.partner_name ?? "",
  bank_code: register.bank_code,
  bank_branch_code: register.bank_branch_code,
  due_date: register.due_date,
  cheque_amount: register.amount,
  amount: register.home_amount,
  description: register.description ?? "",
  fee_amount: "",
});

/** แถวของเอกสารที่บันทึกแล้ว: ใช้ยอดบาทเป็นทั้งยอดบนเช็คและยอดบาท (รับเฉพาะบาท อัตรา 1 จนกว่าจะตัดสิน D15) */
const docLineFromSaved = (line: ChequeDoc["lines"][number]): DocLineDraft => ({
  cheque_id: line.cheque_id,
  cheque_no: line.cheque_no,
  partner_name: line.partner_name,
  bank_code: line.bank_code,
  bank_branch_code: line.bank_branch_code,
  due_date: line.due_date,
  cheque_amount: line.amount,
  amount: line.amount,
  description: line.description,
  fee_amount: line.fee_amount,
});

/** เอกสารที่บันทึกแล้ว → ข้อมูลที่แสดงบนฟอร์ม (อ่านอย่างเดียว) */
export const draftFromDoc = (doc: ChequeDoc): DocDraft => ({
  doc_no: doc.doc_no,
  doc_date: doc.doc_date,
  book_code: doc.book_code ?? "",
  cancel_type: doc.cancel_type ?? 0,
  cancel_passed: doc.cancel_passed,
  description: doc.description ?? "",
  lines: (doc.lines ?? []).map(docLineFromSaved),
});

/** เพิ่มเช็คที่เลือกจากรายการ — เช็คที่อยู่ในเอกสารแล้วไม่ถูกเพิ่มซ้ำ (backend ปฏิเสธแถวซ้ำ) */
export function addDocLines(lines: readonly DocLineDraft[], picked: readonly ChequeRegister[]): DocLineDraft[] {
  const existing = new Set(lines.map((line) => line.cheque_id));
  const added = picked.filter((register) => !existing.has(register.id)).map(docLineFromRegister);
  return [...lines, ...added];
}

export const removeDocLine = (lines: readonly DocLineDraft[], chequeId: string): DocLineDraft[] => lines.filter((line) => line.cheque_id !== chequeId);

export const setDocLineFee = (lines: readonly DocLineDraft[], chequeId: string, fee: string): DocLineDraft[] =>
  lines.map((line) => (line.cheque_id === chequeId ? { ...line, fee_amount: fee } : line));

/** ยอดที่หักค่าใช้จ่ายแล้วของแถว (ค่าใช้จ่ายว่างหรือไม่ถูกต้องนับเป็น 0) */
export const docLineNet = (line: DocLineDraft): string => subtractMoney(line.amount, line.fee_amount);

/** ยอดท้ายเอกสาร: จอนำฝาก (ไม่มีค่าใช้จ่าย) มีแค่ยอดรวม — ค่าใช้จ่ายเป็น 0 และยอดสุทธิเท่ายอดรวม */
export function docDraftTotals(config: ChequeDocConfig, lines: readonly DocLineDraft[]): { total: string; fee: string; net: string } {
  const total = sumMoney(lines.map((line) => line.amount));
  if (!config.hasFee) return { total, fee: "0.00", net: total };
  const fee = sumMoney(lines.map((line) => line.fee_amount));
  return { total, fee, net: subtractMoney(total, fee) };
}

/** เช็คที่ค่าใช้จ่ายไม่ถูกต้อง: ไม่ใช่จำนวนเงิน หรือมากกว่ายอดเช็ค (ช่องว่าง = 0 ใช้ได้) */
export function invalidFeeChequeIds(config: ChequeDocConfig, lines: readonly DocLineDraft[]): string[] {
  if (!config.hasFee) return [];
  return lines
    .filter((line) => {
      const fee = line.fee_amount.trim();
      if (!fee) return false;
      const feeCents = moneyToCents(fee);
      const amountCents = moneyToCents(line.amount);
      return !isMoneyText(fee) || feeCents === null || amountCents === null || feeCents > amountCents;
    })
    .map((line) => line.cheque_id);
}

export const docBookPolicy = (config: ChequeDocConfig, draft: DocDraft): ChequeBookPolicy => chequeDocBookPolicy(config, draft.cancel_type);

/** ชื่อช่องบังคับที่ยังว่าง เรียงตามลำดับบนจอ (ว่าง = ครบ) */
export function docMissingFields(config: ChequeDocConfig, draft: DocDraft): string[] {
  const missing: string[] = [];
  if (!draft.doc_date.trim()) missing.push("doc_date");
  if (docBookPolicy(config, draft).required && !draft.book_code.trim()) missing.push("book_code");
  if (draft.lines.length === 0) missing.push("lines");
  return missing;
}

/** ปุ่มบันทึกเปิดเมื่อมีวันที่ สมุด (ถ้าบังคับ) อย่างน้อยหนึ่งเช็ค และค่าใช้จ่ายทุกแถวถูกต้อง */
export function canSaveDocDraft(config: ChequeDocConfig, draft: DocDraft): boolean {
  return docMissingFields(config, draft).length === 0 && invalidFeeChequeIds(config, draft.lines).length === 0;
}

/** เอกสารมีข้อมูลที่ผู้ใช้กรอกไว้หรือไม่ (ใช้เตือนก่อนทิ้ง) — วันที่ตั้งต้นอย่างเดียวไม่นับ */
export function isDocDraftDirty(draft: DocDraft, baseline: DocDraft): boolean {
  return JSON.stringify({ ...draft, doc_no: draft.doc_no.trim(), description: draft.description.trim() }) !==
    JSON.stringify({ ...baseline, doc_no: baseline.doc_no.trim(), description: baseline.description.trim() });
}

/** คำสั่งสร้างเอกสาร: ส่งเฉพาะ id เช็ค + ค่าใช้จ่ายต่อแถว กับยอดท้ายเอกสารให้ backend ตรวจซ้ำ; ไม่ส่ง cancel_type/cancel_passed กับเอกสารที่ไม่มี */
export function docInputFromDraft(config: ChequeDocConfig, draft: DocDraft): ChequeDocInput {
  const policy = docBookPolicy(config, draft);
  return {
    transflag: config.flag,
    doc_no: draft.doc_no.trim(),
    doc_date: draft.doc_date,
    book_code: policy.none ? "" : draft.book_code.trim(),
    cancel_type: config.cancelTypeCombo ? draft.cancel_type : undefined,
    cancel_passed: config.cancelPassedBox ? draft.cancel_passed : undefined,
    description: draft.description.trim(),
    lines: draft.lines.map((line) => ({ cheque_id: line.cheque_id, fee_amount: config.hasFee ? line.fee_amount.trim() : "" })),
    totals: docDraftTotals(config, draft.lines),
  };
}

/** สถานะที่เช็คในเอกสารจะกลับไปเป็นเมื่อยกเลิกเอกสาร (ไม่ซ้ำ เรียงจากน้อยไปมาก) — ใช้บอกผลในหน้าต่างยืนยัน */
export const docRevertStatuses = (lines: readonly { from_status: number }[]): number[] =>
  [...new Set(lines.map((line) => line.from_status))].sort((a, b) => a - b);
