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
export type GLJournalDetails = {partners?:GLDetailPartner[];bank_accounts?:GLDetailBankAccount[];documents?:GLDetailDocument[];allocations?:GLDetailAllocation[];settlements?:GLDetailSettlement[];bank_lines?:GLDetailBankLine[];statement_lines?:GLDetailStatement[];matches?:GLDetailMatch[];withdrawals?:GLDetailWithdrawal[]};
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
  return result;
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
