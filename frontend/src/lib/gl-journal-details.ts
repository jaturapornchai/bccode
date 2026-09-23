import { amountUnits } from "./general-ledger";
export type GLDetailPartner = {partner_code:string;name_th:string;tax_id?:string;tax_branch_no?:string;address?:string;is_customer:boolean;is_supplier:boolean;is_active:boolean;version?:number};
export type GLDetailBankAccount = {bank_account_code:string;bank_name:string;account_number:string;account_name:string;gl_account_code:string;currency_code:string;is_active:boolean;version?:number};
export type GLDetailDocument = {id:string;ledger:string;partner_code:string;document_no:string;document_date:string;due_date?:string;branch_code:string;document_kind:number;balance_side:number;amount:string;currency_code:string;control_account_code:string;version?:number};
export type GLDetailAllocation = {id:string;ledger:string;document_id:string;journal_id?:string;line_number:number;amount:string};
export type GLDetailSettlement = {id:string;ledger:string;partner_code:string;debt_document_id:string;payment_document_id:string;settlement_date:string;amount:string};
export type GLDetailBankLine = {journal_id?:string;line_number:number;bank_account_code:string;direction:number};
export type GLDetailStatement = {id:string;bank_account_code:string;source_key:string;transaction_date:string;value_date?:string;bank_reference?:string;description?:string;direction:number;amount:string;balance_after?:string};
export type GLDetailMatch = {id:string;statement_line_id:string;journal_id?:string;line_number:number;amount:string};
export type GLDetailWithdrawal = {kind:string;id:string;reason:string};
/** ภาษีหัก ณ ที่จ่ายประกอบใบสำคัญ (ตาม mydocs wht.sql) — ฐานภาษีแก้ได้เสมอ; tax_amount ว่าง = backend คำนวณ ฐาน × อัตรา */
export type GLDetailWithholding = {id:string;wht_direction:number;form_type:string;partner_code:string;wht_cert_no?:string;certificate_date?:string;payment_date:string;income_tax_type:string;income_description?:string;condition_type:number;wht_rate:string;base_amount:string;tax_amount?:string};
/** ภาษีมูลค่าเพิ่มประกอบใบสำคัญ (ตาม mydocs vat.sql) — 1 แถว = 1 ใบกำกับภาษี; ฐานภาษีแก้ได้เสมอ; vat_amount ว่าง = backend คำนวณ ฐาน × อัตรา
 *  tax_type 1=ซื้อ 2=ขาย · document_type 1=ใบกำกับ 2=ใบเพิ่มหนี้ 3=ใบลดหนี้ · claim_status (เฉพาะซื้อ) 1=ใช้สิทธิ 2=ต้องห้าม 3=รอใช้สิทธิ 4=ไม่ใช้สิทธิ */
export type GLDetailVat = {id:string;tax_type:number;document_type:number;tax_invoice_no:string;tax_invoice_date:string;original_invoice_no?:string;original_invoice_date?:string;tax_period_year?:number;tax_period_month?:number;partner_code?:string;partner_tax_id?:string;partner_branch_no?:string;partner_name:string;base_amount:string;zero_rate_amount:string;exempt_amount:string;vat_rate:string;vat_amount?:string;claim_status?:number;claim_reason?:string;remark?:string};
export type GLJournalDetails = {partners?:GLDetailPartner[];bank_accounts?:GLDetailBankAccount[];documents?:GLDetailDocument[];allocations?:GLDetailAllocation[];settlements?:GLDetailSettlement[];bank_lines?:GLDetailBankLine[];statement_lines?:GLDetailStatement[];matches?:GLDetailMatch[];withdrawals?:GLDetailWithdrawal[];withholdings?:GLDetailWithholding[];vats?:GLDetailVat[]};
export type GLSupportRow = Record<string,string|number|boolean|undefined>;
export type GLSupportKind = "partners"|"bank-accounts"|"documents"|"statements"|"bank-lines"|"allocations"|"settlements"|"matches";
export function supportLabel(kind: GLSupportKind, row: GLSupportRow): string {
  const remaining = row.remaining_amount === undefined ? "" : ` · คงเหลือ ${row.remaining_amount}`;
  if(kind === "partners") return `${row.partner_code ?? ""} · ${row.name_th ?? ""}`;
  if(kind === "bank-accounts") return `${row.bank_account_code ?? ""} · ${row.bank_name ?? ""} ${row.account_number ?? ""}`;
  if(kind === "documents") return `${row.ledger === "ap" ? "เจ้าหนี้" : "ลูกหนี้"} · ${row.document_no ?? ""} · ${row.partner_code ?? ""} · ${row.amount ?? ""}${remaining}`;
  if(kind === "statements") return `${row.transaction_date ?? ""} · ${row.bank_account_code ?? ""} · ${row.bank_reference || row.description || row.source_key || ""} · ${row.amount ?? ""}${remaining}`;
  if(kind === "bank-lines") return `${row.docno ?? "ใบสำคัญนี้"} · บรรทัด ${row.line_number ?? ""} · ${row.bank_account_code ?? ""} · ${row.amount ?? ""}${remaining}`;
  if(kind === "settlements") return `${row.debt_document_no ?? "บิล"} ↔ ${row.payment_document_no ?? "ผลชำระ"} · ${row.amount ?? ""}`;
  return `${row.document_no || row.docno || row.bank_reference || row.partner_code || "รายการ"} · ${row.settlement_date || row.created_at || ""} · ${row.amount ?? ""}`;
}
export function reconciliationChanges(before: GLJournalDetails = {}, after: GLJournalDetails = {}): GLJournalDetails {
  const result: GLJournalDetails = {};
  for(const key of ["allocations","statement_lines","settlements","matches","withdrawals"] as const) {
    const identity = (row: {id:string;kind?:string}) => key === "withdrawals" ? `${row.kind}:${row.id}` : row.id;
    const oldIDs = new Set((before[key] ?? []).map(identity));
    const rows = (after[key] ?? []).filter(row => !oldIDs.has(identity(row)));
    Object.assign(result, {[key]: rows});
  }
  // ภาษีหัก ณ ที่จ่ายส่งทั้งชุดเมื่อมีการแก้ (backend แทนทั้งชุดและเก็บค่าเดิมใน audit)
  if((after.withholdings ?? []).length > 0 && JSON.stringify(before.withholdings ?? []) !== JSON.stringify(after.withholdings)) result.withholdings = after.withholdings;
  // ภาษีมูลค่าเพิ่มเช่นเดียวกัน: ส่งทั้งชุดเมื่อแก้ (backend แทนทั้งชุดและเก็บค่าเดิมใน audit vat_replace)
  if((after.vats ?? []).length > 0 && JSON.stringify(before.vats ?? []) !== JSON.stringify(after.vats)) result.vats = after.vats;
  return result;
}
/** ประเภทภาษีเริ่มต้นตามสมุด: UV = สมุดรายวันขาย → ภาษีขาย (2); อื่น ๆ → ภาษีซื้อ (1) ผู้ใช้เปลี่ยนได้ */
export function defaultVatTaxType(bookcode?: string): 1 | 2 { return bookcode === "UV" ? 2 : 1; }
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
