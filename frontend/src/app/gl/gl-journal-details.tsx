"use client";

import { useEffect, useState, useRef } from "react";
import { AlertTriangle, Info } from "lucide-react";
import { Button } from "@/components/ui/button";
import { glRequest } from "@/lib/general-ledger-api";
import { type GLAccount, type GLLine, amountUnits, formatAmount } from "@/lib/general-ledger";
import { PARTNER_TEXT_LIMITS, type GLJournalDetails, type GLSupportKind, type GLSupportRow, defaultVatTaxType, normalizeBranchNo, parseStatementCsv, supportLabel, taxDigits, validThaiTaxId, vatClaimTiming, vatPeriodChoices, vatPeriodPatch, vatPeriodValue } from "@/lib/gl-journal-details";
import { WHT_INCOME_OPTIONS } from "@/lib/thai-tax";
import { AccountSelect, AmountInput, Field, Notice, control, downloadText, useGLText } from "./gl-common";

type Key = keyof GLJournalDetails;
type Row = GLSupportRow;
type Input = {key:string;label:string;type?:"date"|"amount"|"check"|"account"|"line"|"period"|"taxid"|"branch"|"postcode";options?:[string,string][];lookup?:GLSupportKind;side?:number;placeholder?:string;maxLength?:number;show?:(row:Row)=>boolean};
// ชื่อเดือนอังกฤษเป็น fallback ของ key month_* ใน languages.tsv (ไทยมาจาก dictionary)
const MONTH_NAMES=["January","February","March","April","May","June","July","August","September","October","November","December"];
const isPurchaseVat=(row:Row)=>Number(row.tax_type)===1;
function detailFields(tr:(key:string,fallback:string)=>string) {
const money:Input={key:"amount",label:tr("gl_detail_matches_amount","ยอดจัดสรร / จำนวนเงิน"),type:"amount"};
const bank:Input={key:"bank_account_code",label:tr("gl_detail_statement_lines_bank_account_code","บัญชีธนาคาร"),lookup:"bank-accounts"};
const partner:Input={key:"partner_code",label:tr("gl_detail_documents_partner_code","คู่ค้า"),lookup:"partners"};
const direction:Input={key:"direction",label:tr("gl_detail_statement_lines_direction","ทิศทาง"),options:[["1",tr("gl_money_in","เงินเข้า")],["2",tr("gl_money_out","เงินออก")]]};
const line:Input={key:"line_number",label:tr("gl_detail_bank_lines_line_number","บรรทัดบัญชี"),type:"line"};
const fields:Record<Key,Input[]>={
  partners:[{key:"partner_code",label:tr("gl_detail_partners_partner_code","รหัสคู่ค้า")},{key:"title_name",label:tr("gl_detail_partners_title_name","คำนำหน้าชื่อ (ไม่บังคับ เช่น นาย, บริษัท)"),maxLength:PARTNER_TEXT_LIMITS.title_name},{key:"name_th",label:tr("gl_detail_partners_name_th","ชื่อคู่ค้า")},{key:"tax_id",label:tr("gl_detail_partners_tax_id","เลขผู้เสียภาษี"),type:"taxid"},{key:"tax_branch_no",label:tr("gl_detail_partners_tax_branch_no","สาขาภาษี"),type:"branch"},{key:"address",label:tr("gl_detail_partners_address","ที่อยู่")},{key:"addr_district",label:tr("gl_detail_partners_addr_district","อำเภอ/เขต (ไม่บังคับ)"),maxLength:PARTNER_TEXT_LIMITS.addr_district},{key:"addr_province",label:tr("gl_detail_partners_addr_province","จังหวัด (ไม่บังคับ)"),maxLength:PARTNER_TEXT_LIMITS.addr_province},{key:"addr_postcode",label:tr("gl_detail_partners_addr_postcode","รหัสไปรษณีย์ (ไม่บังคับ, 5 หลัก)"),type:"postcode"},{key:"is_customer",label:tr("gl_detail_partners_is_customer","ลูกหนี้"),type:"check"},{key:"is_supplier",label:tr("gl_detail_partners_is_supplier","เจ้าหนี้"),type:"check"},{key:"is_active",label:tr("gl_detail_partners_is_active","ใช้งานอยู่ (ยกเลิกติ๊ก = เลิกใช้คู่ค้านี้ในรายการใหม่)"),type:"check"}],
  bank_accounts:[{key:"bank_account_code",label:tr("gl_detail_bank_accounts_bank_account_code","รหัสบัญชีธนาคาร")},{key:"bank_name",label:tr("gl_detail_bank_accounts_bank_name","ธนาคาร")},{key:"account_number",label:tr("gl_detail_bank_accounts_account_number","เลขบัญชี")},{key:"account_name",label:tr("gl_detail_bank_accounts_account_name","ชื่อบัญชี")},{key:"gl_account_code",label:tr("gl_detail_bank_accounts_gl_account_code","บัญชี GL ธนาคาร"),type:"account"}],
  documents:[{key:"ledger",label:tr("gl_detail_documents_ledger","ประเภทบัญชี"),options:[["ar",tr("gl_detail_partners_is_customer","ลูกหนี้")],["ap",tr("gl_detail_partners_is_supplier","เจ้าหนี้")]]},partner,{key:"document_no",label:tr("gl_detail_documents_document_no","เลขเอกสาร")},{key:"document_date",label:tr("gl_detail_documents_document_date","วันที่เอกสาร"),type:"date"},{key:"due_date",label:tr("gl_detail_documents_due_date","วันครบกำหนด"),type:"date"},{key:"document_kind",label:tr("gl_detail_documents_document_kind","ประเภทเอกสาร"),options:[["1",tr("gl_details_ui_1","ตั้งหนี้")],["2",tr("gl_details_ui_2","ผลชำระที่เกิดแล้ว")],["3",tr("gl_details_ui_3","เพิ่มหนี้")],["4",tr("gl_details_ui_4","ลดหนี้")],["5",tr("gl_opening_balance_2","ยอดยกมา")]]},{key:"balance_side",label:tr("gl_detail_documents_balance_side","ด้านหนี้"),options:[["1",tr("gl_details_ui_3","เพิ่มหนี้")],["2",tr("gl_details_ui_4","ลดหนี้")]]},money,{key:"control_account_code",label:tr("gl_detail_documents_control_account_code","บัญชีคุม"),type:"account"},{key:"branch_code",label:tr("gl_detail_documents_branch_code","สาขา")}],
  allocations:[{key:"document_id",label:tr("gl_detail_allocations_document_id","เอกสารที่จัดสรร"),lookup:"documents"},line,money],
  settlements:[{key:"debt_document_id",label:tr("gl_detail_settlements_debt_document_id","บิลเพิ่มหนี้"),lookup:"documents",side:1},{key:"payment_document_id",label:tr("gl_detail_settlements_payment_document_id","เอกสารลดหนี้ / ผลชำระ"),lookup:"documents",side:2},{key:"settlement_date",label:tr("gl_detail_settlements_settlement_date","วันที่ตัดยอด"),type:"date"},money],
  bank_lines:[line,bank,direction],
  statement_lines:[bank,{key:"source_key",label:tr("gl_detail_statement_lines_source_key","รหัสรายการจากธนาคาร")},{key:"transaction_date",label:tr("gl_detail_statement_lines_transaction_date","วันที่ธนาคาร"),type:"date"},{key:"value_date",label:tr("gl_detail_statement_lines_value_date","วันที่มีผล"),type:"date"},{key:"bank_reference",label:tr("gl_detail_statement_lines_bank_reference","เลขอ้างอิงธนาคาร")},{key:"description",label:tr("gl_detail_statement_lines_description","รายละเอียด")},direction,money],
  matches:[{key:"statement_line_id",label:tr("gl_detail_matches_statement_line_id","รายการ Statement"),lookup:"statements"},{key:"line_number",label:tr("gl_detail_matches_line_number","รายการบัญชีธนาคารที่จะจับคู่"),lookup:"bank-lines"},money],
  withholdings:[{key:"wht_direction",label:tr("gl_detail_withholdings_wht_direction","ทิศทางภาษี"),options:[["1",tr("gl_wht_direction_paid","เราหักภาษีผู้รับเงิน (ยื่น ภ.ง.ด.)")],["2",tr("gl_wht_direction_received","ผู้จ่ายหักภาษีเรา (ภาษีถูกหัก)")]]},{key:"form_type",label:tr("gl_detail_withholdings_form_type","แบบยื่น"),options:[["PND53",tr("wht_cert_ui_form_53","ภ.ง.ด.53 (นิติบุคคล)")],["PND3",tr("wht_cert_ui_form_3","ภ.ง.ด.3 (บุคคลธรรมดา)")],["PND2",tr("wht_cert_ui_form_2","ภ.ง.ด.2 (ดอกเบี้ย/เงินปันผล)")]]},{...partner,label:tr("gl_detail_withholdings_partner_code","ผู้ถูกหักภาษี / ผู้หักภาษี")},{key:"payment_date",label:tr("gl_detail_withholdings_payment_date","วันที่จ่ายเงิน"),type:"date"},{key:"income_tax_type",label:tr("gl_detail_withholdings_income_tax_type","ประเภทเงินได้"),options:WHT_INCOME_OPTIONS.map(o=>[o.value,tr(o.key,o.th)] as [string,string])},{key:"income_description",label:tr("gl_detail_withholdings_income_description","รายละเอียดเงินได้")},{key:"condition_type",label:tr("gl_detail_withholdings_condition_type","ผู้จ่ายเงิน"),options:[["1",tr("wht_cert_ui_condition_withhold","(1) หัก ณ ที่จ่าย")],["2",tr("wht_cert_ui_condition_always","(2) ออกให้ตลอดไป")],["3",tr("wht_cert_ui_condition_once","(3) ออกให้ครั้งเดียว")]]},{key:"base_amount",label:tr("gl_detail_withholdings_base_amount","ฐานภาษี (จำนวนเงินที่จ่าย)"),type:"amount"},{key:"wht_rate",label:tr("gl_detail_withholdings_wht_rate","อัตราภาษี (%)"),type:"amount",placeholder:tr("gl_example_3","เช่น 3")},{key:"tax_amount",label:tr("gl_detail_withholdings_tax_amount","ภาษีที่หัก"),type:"amount",placeholder:tr("gl_wht_tax_auto","เว้นว่าง = คำนวณ ฐาน × อัตรา ตอนบันทึก")},{key:"wht_cert_no",label:tr("gl_detail_withholdings_wht_cert_no","เลขที่หนังสือรับรอง")},{key:"certificate_date",label:tr("gl_detail_withholdings_certificate_date","วันที่ออกหนังสือรับรอง"),type:"date"}],
  // ภาษีมูลค่าเพิ่ม 1 แถว = 1 ใบกำกับภาษี (vat.sql) — ช่องใบกำกับเดิมแสดงเฉพาะใบเพิ่ม/ลดหนี้, ช่องใช้สิทธิเฉพาะภาษีซื้อ
  vats:[
    {key:"tax_type",label:tr("gl_detail_vats_tax_type","ประเภทภาษี"),options:[["1",tr("gl_vat_tax_type_purchase","ภาษีซื้อ (ใบกำกับที่ได้รับ)")],["2",tr("gl_vat_tax_type_sale","ภาษีขาย (ใบกำกับที่ออก)")]]},
    {key:"document_type",label:tr("gl_detail_vats_document_type","ประเภทเอกสาร"),options:[["1",tr("gl_vat_document_invoice","ใบกำกับภาษี")],["2",tr("gl_vat_document_debit_note","ใบเพิ่มหนี้")],["3",tr("gl_vat_document_credit_note","ใบลดหนี้")]]},
    {key:"tax_invoice_no",label:tr("gl_detail_vats_tax_invoice_no","เลขที่ใบกำกับภาษี"),maxLength:50},
    {key:"tax_invoice_date",label:tr("gl_detail_vats_tax_invoice_date","วันที่ใบกำกับภาษี"),type:"date"},
    {key:"original_invoice_no",label:tr("gl_detail_vats_original_invoice_no","เลขที่ใบกำกับภาษีเดิม"),maxLength:50,show:row=>Number(row.document_type)!==1},
    {key:"original_invoice_date",label:tr("gl_detail_vats_original_invoice_date","วันที่ใบกำกับภาษีเดิม"),type:"date",show:row=>Number(row.document_type)!==1},
    {key:"tax_period",label:tr("gl_detail_vats_tax_period","งวดภาษี (เดือนที่ยื่น ภ.พ.30)"),type:"period"},
    {...partner,label:tr("gl_detail_vats_partner_code","คู่ค้า (ไม่บังคับ — เลือกแล้วเติมชื่อและเลขภาษีให้)")},
    {key:"partner_name",label:tr("gl_detail_vats_partner_name","ชื่อผู้ซื้อ / ผู้ขาย")},
    {key:"partner_tax_id",label:tr("gl_detail_vats_partner_tax_id","เลขประจำตัวผู้เสียภาษี (13 หลัก)"),type:"taxid"},
    {key:"partner_branch_no",label:tr("gl_detail_vats_partner_branch_no","สาขา (5 หลัก, 00000 = สำนักงานใหญ่)"),type:"branch"},
    {key:"base_amount",label:tr("gl_detail_vats_base_amount","ฐานภาษี (มูลค่าที่ต้องเสียภาษี)"),type:"amount"},
    {key:"zero_rate_amount",label:tr("gl_detail_vats_zero_rate_amount","มูลค่าอัตราร้อยละ 0"),type:"amount"},
    {key:"exempt_amount",label:tr("gl_detail_vats_exempt_amount","มูลค่าที่ได้รับยกเว้น"),type:"amount"},
    {key:"vat_rate",label:tr("gl_detail_vats_vat_rate","อัตราภาษี (%)"),type:"amount",placeholder:tr("gl_example_7","เช่น 7")},
    {key:"vat_amount",label:tr("gl_detail_vats_vat_amount","ภาษีมูลค่าเพิ่ม"),type:"amount",placeholder:tr("gl_wht_tax_auto","เว้นว่าง = คำนวณ ฐาน × อัตรา ตอนบันทึก")},
    {key:"claim_status",label:tr("gl_detail_vats_claim_status","การใช้สิทธิภาษีซื้อ"),options:[["1",tr("gl_vat_claim_used","ใช้สิทธิในงวดนี้")],["2",tr("gl_vat_claim_forbidden","ภาษีซื้อต้องห้าม")],["3",tr("gl_vat_claim_pending","รอใช้สิทธิ")],["4",tr("gl_vat_claim_not_used","ไม่ใช้สิทธิ")]],show:isPurchaseVat},
    {key:"claim_reason",label:tr("gl_detail_vats_claim_reason","เหตุผล (เลื่อนงวด / ต้องห้าม / ไม่ใช้สิทธิ)"),maxLength:500,show:isPurchaseVat},
    {key:"remark",label:tr("gl_detail_vats_remark","หมายเหตุ"),maxLength:500},
  ],
  withdrawals:[{key:"kind",label:tr("gl_detail_withdrawals_kind","ประเภทการถอน"),options:[["allocation",tr("gl_details_ui_5","จัดสรรเอกสาร")],["settlement",tr("gl_details_ui_6","ตัดชำระ")],["match",tr("gl_details_ui_7","จับคู่ธนาคาร")]]},{key:"id",label:tr("gl_detail_withdrawals_id","รายการเดิมที่จะถอน"),lookup:"matches"},{key:"reason",label:tr("gl_detail_withdrawals_reason","เหตุผลการถอน")}],
};
const sections:[Key,string][]=[["withholdings",tr("gl_details_section_withholdings","ภาษีหัก ณ ที่จ่าย")],["vats",tr("gl_details_section_vats","ภาษีมูลค่าเพิ่ม (ใบกำกับภาษี)")],["documents",tr("gl_details_section_documents","ลูกหนี้–เจ้าหนี้ / เอกสารค้าง")],["allocations",tr("gl_details_section_allocations","จัดสรรบรรทัดบัญชีกับเอกสาร")],["settlements",tr("gl_details_section_settlements","ตัดยอดหลายบิล / ชำระบางส่วน")],["bank_lines",tr("gl_details_section_bank_lines","บัญชีธนาคารของบรรทัด GL")],["statement_lines",tr("gl_details_section_statement_lines","หลักฐาน Statement ธนาคาร")],["matches",tr("gl_details_section_matches","จับคู่บัญชีกับ Statement")],["withdrawals",tr("gl_details_section_withdrawals","ถอนรายการที่ยืนยันแล้ว")],["partners",tr("gl_details_section_partners","คู่ค้าภายในห้องบัญชี")],["bank_accounts",tr("gl_details_section_bank_accounts","บัญชีธนาคาร")]];
return {fields,sections};
}
// ภาษีหัก ณ ที่จ่าย/ภาษีมูลค่าเพิ่มเป็นรายละเอียดประกอบ แก้ได้เสมอแม้ผ่านบัญชีแล้ว (ยอด GL ไม่เปลี่ยน, backend เก็บค่าเดิมใน audit)
const taxKeys:Key[]=["withholdings","vats"];
const postedKeys:Key[]=["allocations","statement_lines","settlements","matches","withdrawals",...taxKeys];
const identity=(kind:GLSupportKind,row:Row)=>String(kind==="partners"?row.partner_code:kind==="bank-accounts"?row.bank_account_code:kind==="bank-lines"?`${row.journal_id ?? ""}:${row.line_number}`:row.id ?? "");
const detailRows=(details:GLJournalDetails,key:Key)=> (details[key] ?? []) as unknown as Row[];
// ภาษีหัก: snapshot ผู้จ่าย/ผู้รับเงินของแถวที่เปลี่ยนคู่ค้าหรือทิศทางเป็นของคู่ค้าเดิม — ล้างให้ backend เติมจากทะเบียนคู่ค้าใหม่ (เดิม 50 ทวิ/ภ.ง.ด. ออกในนามคู่ค้าเดิม)
export const clearWithholdingParty=(side:"payer"|"payee"):Row=>({[`${side}_tax_id`]:undefined,[`${side}_branch_no`]:undefined,[`${side}_name`]:undefined,[`${side}_address`]:undefined});
function newRow(key:Key,date:string,branch:string,booktype?:number):Row {
  const id=crypto.randomUUID();
  const taxType=defaultVatTaxType(booktype);
  const defaults:Record<Key,Row>={partners:{partner_code:"",name_th:"",is_customer:true,is_supplier:false,is_active:true},bank_accounts:{bank_account_code:"",bank_name:"",account_name:"",account_number:"",gl_account_code:"",currency_code:"THB",is_active:true},documents:{id,ledger:"ar",partner_code:"",document_no:"",document_date:date,due_date:date,branch_code:branch,document_kind:1,balance_side:1,amount:"0",currency_code:"THB",control_account_code:""},allocations:{id,ledger:"ar",document_id:"",line_number:1,amount:"0"},settlements:{id,ledger:"ar",partner_code:"",debt_document_id:"",payment_document_id:"",settlement_date:date,amount:"0"},bank_lines:{line_number:1,bank_account_code:"",direction:1},statement_lines:{id,bank_account_code:"",source_key:"",transaction_date:date,direction:1,amount:"0"},matches:{id,statement_line_id:"",line_number:1,amount:"0"},withdrawals:{kind:"match",id:"",reason:""},withholdings:{id,wht_direction:1,form_type:"PND53",partner_code:"",payment_date:date,income_tax_type:"3_tres",income_description:"",condition_type:1,wht_rate:"3",base_amount:"0"},vats:{id,tax_type:taxType,document_type:1,tax_invoice_no:"",tax_invoice_date:date,...vatPeriodPatch(date.slice(0,7)),partner_name:"",partner_tax_id:"",partner_branch_no:"",base_amount:"0",zero_rate_amount:"0",exempt_amount:"0",vat_rate:"7.00",...(taxType===1?{claim_status:1}:{})}};
  return defaults[key];
}

// ข้อความในใบกำกับภาษีเป็นภาษาไทยเสมอ: "{เดือนไทย} {ปี พ.ศ.}" เช่น "กันยายน 2569" — ชื่อเดือนจาก locale th ของเบราว์เซอร์ (ไม่ขึ้นกับภาษาจอ)
// ปี พ.ศ. คิดเอง (ค.ศ. + 543) ไม่ให้ Intl ใส่คำว่า "พ.ศ." หรือรูปแบบปีต่างกันตามเบราว์เซอร์
export const vatClaimPeriodTh=(year:number,month:number)=>`${new Intl.DateTimeFormat("th-TH",{month:"long",timeZone:"UTC"}).format(new Date(Date.UTC(year,month-1,1)))} ${year+543}`;
// ภาษีซื้อที่ใช้สิทธิหลังเดือนที่ออกใบกำกับ: บอกข้อความที่ต้องเขียนในใบกำกับ (1–6 เดือน) หรือบอกทางแก้ก่อนกดบันทึก (นอกกำหนด)
// ข้อความนอกกำหนดใช้ key เดียวกับ error ของ backend เพื่อให้ผู้ใช้เห็นคำแนะนำเดียวกันก่อนและหลังบันทึก (backend ยังตรวจซ้ำเสมอ)
export function VatClaimHint({row,tr}:{row:Row;tr:(key:string,fallback:string)=>string}) {
  const timing=vatClaimTiming(row);
  if(!timing)return null;
  // {period} = เดือนตามภาษาจอ; {period_th} = ข้อความภาษาไทยที่ต้องเขียนบนใบกำกับ (ไม่ขึ้นกับภาษาจอ); แทนทุกตำแหน่ง (replaceAll)
  if(timing.kind==="late"){const name=MONTH_NAMES[timing.periodMonth-1];const period=`${tr(`month_${name.toLowerCase()}`,name)} ${timing.periodYear+543}`;const periodTh=vatClaimPeriodTh(timing.periodYear,timing.periodMonth);
    return <p role="status" className="flex items-start gap-2 rounded-lg border border-primary/30 bg-primary/5 p-2 leading-relaxed text-foreground" data-field="vat-claim-late"><Info className="mt-1 size-4 shrink-0 text-primary" aria-hidden/><span>{tr("vat_ui_claim_late_hint","ใช้สิทธิภาษีซื้อหลังเดือนที่ออกใบกำกับ {months} เดือน — ใบกำกับภาษีฉบับนี้ต้องมีข้อความ “ถือเป็นภาษีซื้อในเดือนภาษี {period_th}” (ประกาศอธิบดีฯ ภาษีมูลค่าเพิ่ม ฉบับที่ 4 ข้อ 2)").replaceAll("{months}",String(timing.months)).replaceAll("{period_th}",periodTh).replaceAll("{period}",period)}</span></p>;}
  const message=timing.kind==="before"?tr("gl_err_vat_claim_before_invoice_month","งวดที่ใช้สิทธิภาษีซื้อต้องไม่ก่อนเดือนที่ออกใบกำกับภาษีหรือใบเพิ่มหนี้ — เลือกงวดตั้งแต่เดือนของวันที่เอกสารเป็นต้นไป หรือแก้วันที่เอกสารให้ถูกต้อง")
    :timing.kind==="credit_note_before"?tr("gl_err_vat_credit_note_period_before_note","ใบลดหนี้ต้องนำไปลดภาษีซื้อในงวดของเดือนที่ได้รับใบลดหนี้ — เลือกงวดตั้งแต่เดือนของวันที่ใบลดหนี้เป็นต้นไป หรือแก้วันที่ใบลดหนี้ให้ถูกต้อง")
    :timing.kind==="buddhist_year"?tr("gl_err_vat_period_year_buddhist","ปีงวดภาษีให้ใช้ ค.ศ. เช่น 2026 (ถ้ากรอกปี พ.ศ. ให้ลบ 543)")
    :tr("gl_err_vat_claim_window_exceeded","ภาษีซื้อตามใบกำกับภาษีหรือใบเพิ่มหนี้ใช้สิทธิได้ไม่เกิน 6 เดือนนับแต่เดือนถัดจากเดือนที่ออกเอกสาร — ให้บันทึกในงวดของเดือนที่ออกเอกสาร แล้วยื่น ภ.พ.30 เพิ่มเติมของเดือนนั้น");
  return <p role="alert" className="flex items-start gap-2 rounded-lg border border-destructive/40 bg-destructive/5 p-2 leading-relaxed text-foreground" data-field="vat-claim-window"><AlertTriangle className="mt-1 size-4 shrink-0 text-destructive" aria-hidden/><span>{message}</span></p>;
}

type Picker = {kind:GLSupportKind;accept:(row:Row)=>void;filter?:(row:Row)=>boolean};
function SupportPicker({picker,local,onClose,onCache}:{picker:Picker;local:Row[];onClose:()=>void;onCache:(kind:GLSupportKind,rows:Row[])=>void}) {
  const tr=useGLText();
  const dialog=useRef<HTMLDialogElement>(null);
  useEffect(()=>{const element=dialog.current;const trigger=document.activeElement as HTMLElement|null;element?.showModal();return()=>{element?.close();trigger?.focus();};},[]);
  const [asof,setAsOf]=useState("");
  const [query,setQuery]=useState(""),[search,setSearch]=useState(""),[page,setPage]=useState(1);
  const [rows,setRows]=useState<Row[]>([]),[total,setTotal]=useState(0),[error,setError]=useState(""),[busy,setBusy]=useState(false);
  useEffect(()=>{const controller=new AbortController();setBusy(true);setError("");
    glRequest<{items:Row[];total:number}>(`journal-support?${new URLSearchParams({kind:picker.kind,q:search,page:String(page),limit:"30",...(asof?{asof}:{})})}`,{signal:controller.signal})
      .then(data=>{if(!controller.signal.aborted){setRows(data.items ?? []);setTotal(data.total);onCache(picker.kind,data.items ?? []);}})
      .catch(cause=>{if(!controller.signal.aborted)setError(cause instanceof Error?cause.message:tr("gl_details_ui_8","โหลดรายการไม่สำเร็จ"));})
      .finally(()=>{if(!controller.signal.aborted)setBusy(false);});return()=>controller.abort();
  // onCache updates labels only; it must not restart a request.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  },[picker.kind,search,page,asof]);
  const merged=[...local.filter(row=>!search || supportLabel(picker.kind,row,tr).toLowerCase().includes(search.toLowerCase())),...rows].filter((row,index,all)=>all.findIndex(candidate=>identity(picker.kind,candidate)===identity(picker.kind,row))===index).filter(row=>picker.filter?.(row) ?? true);
  return <dialog ref={dialog} aria-label={tr("gl_details_select_evidence","เลือกหลักฐานบัญชี")} className="fixed inset-0 m-0 h-full max-h-none w-full max-w-none items-center justify-center border-0 bg-black/50 p-4 open:flex" onCancel={event=>{event.preventDefault();event.stopPropagation();onClose();}}>
    <section className="grid max-h-[85vh] w-full max-w-3xl gap-3 overflow-auto rounded-xl border border-border bg-card p-4 text-[0.95rem] shadow-xl">
      <h3 className="font-semibold">{tr("gl_details_select_evidence","เลือกหลักฐานบัญชี")}</h3><div className="flex gap-2"><input autoFocus className={control} aria-label={tr("gl_details_ui_9","ค้นหาหลักฐาน")} value={query} onChange={e=>setQuery(e.target.value)} onKeyDown={e=>{if(e.key==="Enter"){e.preventDefault();setSearch(query);setPage(1);}}}/><Button type="button" disabled={busy} onClick={()=>{setSearch(query);setPage(1);}}>{tr("gl_details_search","ค้นหา")}</Button><Button type="button" variant="outline" onClick={onClose}>{tr("gl_details_close","ปิด")}</Button></div>
      <label className="grid gap-1">{tr("gl_details_asof","ยอดคงเหลือ ณ วันที่")}<input type="date" className={control} value={asof} onChange={event=>{setAsOf(event.target.value);setPage(1);}}/></label>
      <Notice error text={error}/>{busy&&<p role="status">{tr("gl_details_ui_10","กำลังโหลดหลักฐาน")}</p>}
      <ul className="grid gap-2">{merged.map(row=><li key={identity(picker.kind,row)}><Button type="button" variant="outline" className="min-h-11 h-auto w-full justify-start whitespace-normal text-left" onClick={()=>{picker.accept(row);onClose();}}>{supportLabel(picker.kind,row,tr)}</Button></li>)}</ul>
      {!busy&&!merged.length&&<p>{tr("gl_details_ui_11","ไม่พบรายการตามเงื่อนไข")}</p>}
      <div className="flex items-center justify-between gap-2"><Button type="button" variant="outline" disabled={busy||page===1} onClick={()=>setPage(page-1)}>{tr("gl_details_previous","ก่อนหน้า")}</Button><span>{tr("gl_details_ui_12","หน้า")}{page} · {total}{tr("gl_items","รายการ")}</span><Button type="button" variant="outline" disabled={busy||page*30>=total} onClick={()=>setPage(page+1)}>{tr("gl_details_next","ถัดไป")}</Button></div>
    </section>
  </dialog>;
}

function EvidenceLabel({kind,id,fallback}:{kind:GLSupportKind;id:string;fallback:string}) {
  const tr=useGLText();
  const [found,setFound]=useState<Row|null>(null);
  useEffect(()=>{if(!id)return;const controller=new AbortController();setFound(null);
    glRequest<{items:Row[]}>(`journal-support?${new URLSearchParams({kind,q:kind==="bank-lines"?id.split(":")[0]:id,page:"1",limit:"30"})}`,{signal:controller.signal})
      .then(data=>{const match=data.items?.find(row=>identity(kind,row)===id);if(match&&!controller.signal.aborted)setFound(match);})
      .catch(()=>{});return()=>controller.abort();
  },[kind,id]);
  return <>{found?supportLabel(kind,found,tr):fallback}</>;
}

export function GLJournalDetailsPanel({value={},original={},onChange,lines,accounts,date,branch,booktype,journalId,editable,posted,onBusyChange,scale=2}:{value?:GLJournalDetails;original?:GLJournalDetails;onChange:(value:GLJournalDetails)=>void;lines:GLLine[];accounts:GLAccount[];date:string;branch:string;booktype?:number;journalId?:string;editable:boolean;posted:boolean;onBusyChange?:(busy:boolean)=>void;scale?:number}) {
  const tr=useGLText();
  const {fields,sections}=detailFields(tr);
  const [picker,setPicker]=useState<Picker|null>(null),[cache,setCache]=useState<Partial<Record<GLSupportKind,Row[]>>>({}),[error,setError]=useState("");
  const [importBank,setImportBank]=useState<Row|null>(null),[importing,setImporting]=useState(false);
  useEffect(()=>{onBusyChange?.(importing);return()=>onBusyChange?.(false);},[importing,onBusyChange]);
  function local(kind:GLSupportKind):Row[] {const key=({partners:"partners","bank-accounts":"bank_accounts",documents:"documents",statements:"statement_lines","bank-lines":"bank_lines",allocations:"allocations",settlements:"settlements",matches:"matches"} as const)[kind];return detailRows(value,key).map(row=>kind==="bank-lines"?{...row,journal_id:row.journal_id||journalId,amount:amountUnits(lines[Number(row.line_number)-1]?.debit||"0")!==0n?lines[Number(row.line_number)-1]?.debit:lines[Number(row.line_number)-1]?.credit}:row);}
  function update(key:Key,index:number,patch:Row) {onChange({...value,[key]:detailRows(value,key).map((row,i)=>i===index?{...row,...patch}:row)});}
  function remember(kind:GLSupportKind,rows:Row[]) {setCache(current=>({...current,[kind]:[...(current[kind]??[]),...rows].filter((row,index,all)=>all.findIndex(candidate=>identity(kind,candidate)===identity(kind,row))===index)}));}
  function pick(key:Key,index:number,field:Input,row:Row) {
    const kind=key==="withdrawals"?({allocation:"allocations",settlement:"settlements",match:"matches"} as const)[String(row.kind) as "allocation"|"settlement"|"match"]:field.lookup!;
    // เลือกคู่ค้าใหม่: ซ่อนคู่ค้าที่เลิกใช้แล้ว (backend ปฏิเสธคู่ค้าที่ไม่ใช้งานในรายการใหม่อยู่แล้ว)
    setPicker({kind,filter:candidate=>!candidate.reversed_at && !(kind==="partners"&&candidate.is_active===false) && (field.side===undefined||candidate.balance_side===field.side) && (field.key!=="payment_document_id"||(!row.partner_code||(candidate.partner_code===row.partner_code&&candidate.ledger===row.ledger))),accept:candidate=>{
      remember(kind,[candidate]);let patch:Row={[field.key]:kind==="bank-lines"?candidate.line_number:identity(kind,candidate)};
      if(kind==="documents") patch={...patch,ledger:candidate.ledger,...(key==="settlements"?{partner_code:candidate.partner_code}: {})};
      if(kind==="bank-lines") patch={...patch,journal_id:candidate.journal_id};
      // ภาษีมูลค่าเพิ่ม: เลือกคู่ค้าแล้วเติมชื่อ/เลขภาษี/สาขาจากทะเบียนคู่ค้า (ยังแก้ต่อได้ตามใบกำกับจริง)
      if(key==="vats"&&kind==="partners") patch={...patch,partner_name:candidate.name_th??row.partner_name,partner_tax_id:candidate.tax_id??"",partner_branch_no:candidate.tax_branch_no??""};
      if(key==="withholdings"&&kind==="partners"&&patch.partner_code!==row.partner_code) patch={...patch,...clearWithholdingParty(Number(row.wht_direction)===2?"payer":"payee")};
      update(key,index,patch);
    }});
  }
  function renderInput(key:Key,index:number,row:Row,field:Input,canEdit:boolean) {
    const current=row[field.key] ?? "";
    const label=field.label;
    if(field.lookup) {const kind=key==="withdrawals"?({allocation:"allocations",settlement:"settlements",match:"matches"} as const)[String(row.kind) as "allocation"|"settlement"|"match"]:field.lookup;const id=kind==="bank-lines"?`${row.journal_id||journalId||""}:${current}`:String(current);const found=[...local(kind),...(cache[kind]??[])].find(item=>identity(kind,item)===id);return <Button type="button" variant="outline" className="min-h-11 h-auto justify-start whitespace-normal text-left" aria-label={label} disabled={!canEdit && !current} onClick={()=>canEdit?pick(key,index,field,row):setPicker({kind,filter:item=>identity(kind,item)===id,accept:()=>{}})}>{found?supportLabel(kind,found):current?<EvidenceLabel kind={kind} id={id} fallback={tr("gl_details_ui_13","หลักฐานที่บันทึกไว้ · เปิดดูรายการ")}/>:tr("gl_select","เลือก")}</Button>;}
    if(field.type==="account")return <AccountSelect label={label} accounts={accounts} value={String(current)} onChange={v=>update(key,index,{[field.key]:v})} disabled={!canEdit}/>;
    if(field.type==="period"){const period=vatPeriodValue(row);return <select aria-label={label} className={control} value={period} disabled={!canEdit} onChange={e=>update(key,index,vatPeriodPatch(e.target.value))}>{(isPurchaseVat(row)||!period)&&<option value="">{tr("gl_vat_period_none","ยังไม่กำหนดงวด (รอใช้สิทธิ)")}</option>}{vatPeriodChoices(date,period).map(choice=>{const [year,month]=choice.split("-").map(Number);const name=MONTH_NAMES[month-1];return <option key={choice} value={choice}>{tr("gl_vat_period_option","{0} {2}").replace("{0}",tr(`month_${name.toLowerCase()}`,name)).replace("{1}",String(year)).replace("{2}",String(year+543))}</option>;})}</select>;}
    // ยอดภาษีว่าง = backend คำนวณ ฐาน × อัตรา; แก้ฐาน/อัตราแล้วกลับเป็นอัตโนมัติ
    // ช่องเงินอื่นว่าง = "0" (backend ไม่รับ ""); อัตราภาษีว่างคงว่างให้ระบบเตือนก่อนบันทึก ไม่ใส่ 0% ให้เอง
    if(field.type==="amount"){const optional=["tax_amount","vat_amount","balance_after"].includes(field.key),rate=field.key==="wht_rate"||field.key==="vat_rate",missingRate=rate&&canEdit&&!String(current).trim();
      return <span className="grid gap-1"><AmountInput value={String(current)} allowEmpty={optional||rate} onChange={v=>update(key,index,{[field.key]:v===""?(optional?undefined:rate?"":"0"):v,...(key==="withholdings"&&(field.key==="base_amount"||field.key==="wht_rate")?{tax_amount:undefined}:{}),...(key==="vats"&&(field.key==="base_amount"||field.key==="vat_rate")?{vat_amount:undefined}:{})})} scale={field.key==="wht_rate"||field.key==="vat_rate"?2:scale} allowNegative={false} disabled={!canEdit} ariaLabel={label} placeholder={field.placeholder}/>{missingRate&&<span role="alert" className="text-[0.9rem] leading-snug text-destructive">{tr("gl_detail_rate_required_inline","กรุณาใส่อัตราภาษี (%) — ระบบไม่ใส่ 0% ให้เอง")}</span>}</span>;}
    // เลขภาษี/สาขา: รับค่าที่วางแบบมีขีด/เว้นวรรค ตัดเหลือตัวเลขก่อน (ไม่ตัดความยาวก่อน) แล้วบอกทันทีถ้าจำนวนหลักไม่ถูก
    // เลขภาษีครบ 13 หลักแต่หลักตรวจสอบไม่ตรง: เตือนทันที (backend ตัดสินตอนบันทึก — คู่ค้าเดิมในทะเบียนที่ไม่ได้แก้ยังใช้ได้ จึงไม่บล็อกที่จอ)
    if(field.type==="taxid"||field.type==="branch"){const digits=taxDigits(current),taxId=field.type==="taxid",checksum=taxId&&digits.length===13&&!validThaiTaxId(digits),wrong=(taxId?digits.length>0&&digits.length!==13:digits.length>5)||checksum;
      return <span className="grid gap-1"><input aria-label={label} type="text" inputMode="numeric" autoComplete="off" className={`${control}${wrong?" border-destructive":""}`} aria-invalid={wrong||undefined} value={String(current)} disabled={!canEdit} maxLength={32} onChange={e=>update(key,index,{[field.key]:taxDigits(e.target.value)})} onBlur={taxId?undefined:e=>{const padded=normalizeBranchNo(e.target.value);if(padded!==String(current))update(key,index,{[field.key]:padded});}}/>
        {wrong&&<span role="alert" className="text-[0.9rem] leading-snug text-destructive">{checksum?tr("gl_detail_tax_id_checksum_hint","เลขไม่ถูกต้อง — หลักสุดท้าย (หลักตรวจสอบ) ไม่ตรงกับ 12 หลักแรก มักพิมพ์ผิดหรือสลับตัวเลข ตรวจกับหนังสือรับรอง บัตรประชาชน หรือใบกำกับภาษีแล้วพิมพ์ใหม่"):taxId?tr("gl_detail_tax_id_hint","ต้องเป็นตัวเลข 13 หลัก (ตอนนี้ {0} หลัก) — วางแบบมีขีดได้ ระบบตัดขีดให้เอง").replace("{0}",String(digits.length)):tr("gl_detail_branch_hint","ใส่ได้ไม่เกิน 5 หลัก เช่น 0 หรือ 00000 = สำนักงานใหญ่")}</span>}</span>;}
    // รหัสไปรษณีย์ (ไม่บังคับ): ตัวเลขล้วน 5 หลัก — รับเลขไทย/ค่าที่วางแบบมีช่องว่าง แล้วบอกทันทีถ้าจำนวนหลักไม่ถูก (ไม่ตัดความยาวเงียบ ๆ)
    if(field.type==="postcode"){const digits=taxDigits(current),wrong=digits.length>0&&digits.length!==5;
      return <span className="grid gap-1"><input aria-label={label} type="text" inputMode="numeric" autoComplete="off" className={`${control}${wrong?" border-destructive":""}`} aria-invalid={wrong||undefined} value={String(current)} disabled={!canEdit} maxLength={16} onChange={e=>update(key,index,{[field.key]:taxDigits(e.target.value)})}/>
        {wrong&&<span role="alert" className="text-[0.9rem] leading-snug text-destructive">{tr("gl_detail_postcode_hint","รหัสไปรษณีย์ต้องเป็นตัวเลข 5 หลัก (ตอนนี้ {0} หลัก) หรือเว้นว่าง").replace("{0}",String(digits.length))}</span>}</span>;}
    if(field.type==="check")return <input aria-label={label} type="checkbox" checked={Boolean(current)} disabled={!canEdit} onChange={e=>update(key,index,{[field.key]:e.target.checked})}/>;
    if(field.type==="line")return <select aria-label={label} className={control} value={String(current)} disabled={!canEdit} onChange={e=>update(key,index,{[field.key]:Number(e.target.value)})}>{lines.map((entry,i)=><option key={i} value={i+1}>{i+1} · {entry.accountcode} · {entry.description}{tr("gl_details_ui_14","· เดบิต")}{formatAmount(entry.debit)}{tr("gl_details_ui_15","/ เครดิต")}{formatAmount(entry.credit)}</option>)}</select>;
    if(field.options)return <select aria-label={label} className={control} value={String(current)} disabled={!canEdit} onChange={e=>{const numeric=["document_kind","balance_side","direction","wht_direction","condition_type","tax_type","document_type","claim_status"].includes(field.key);update(key,index,{[field.key]:numeric?Number(e.target.value):e.target.value,...(field.key==="document_kind"?{balance_side:["2","4"].includes(e.target.value)?2:1}:{}),...(key==="withholdings"&&field.key==="wht_direction"&&e.target.value!==String(row.wht_direction)?{...clearWithholdingParty("payer"),...clearWithholdingParty("payee")}:{}),...(key==="vats"&&field.key==="tax_type"?(e.target.value==="1"?{claim_status:1}:{claim_status:undefined,claim_reason:undefined,...(vatPeriodValue(row)?{}:vatPeriodPatch(date.slice(0,7)))}):{})});}}>{field.options.map(([v,label])=><option key={v} value={v}>{label}</option>)}</select>;
    return <input aria-label={label} type={field.type==="date"?"date":"text"} className={control} value={String(current)} disabled={!canEdit} maxLength={field.maxLength??(field.key==="reason"?500:255)} onChange={e=>update(key,index,{[field.key]:e.target.value})}/>;
  }
  async function importCsv(file:File) {setError("");setImporting(true);try {if(file.size>2*1024*1024)throw new Error(tr("gl_details_ui_16","ไฟล์ต้องไม่เกิน 2 MB"));const rows=await parseStatementCsv(await file.text(),String(importBank?.bank_account_code ?? ""));onChange({...value,statement_lines:[...(value.statement_lines??[]),...rows.filter(row=>!(value.statement_lines??[]).some(old=>old.id===row.id))]});}catch(cause){setError(cause instanceof Error?cause.message:tr("gl_details_ui_17","นำเข้าไม่สำเร็จ"));}finally{setImporting(false);}}
  return <section aria-label={tr("gl_details_ui_18","รายละเอียดตรวจสอบลูกหนี้ เจ้าหนี้ และธนาคาร")} className="grid gap-3 rounded-xl border border-border bg-card p-3 text-[0.95rem] leading-relaxed shadow-sm">
    <h3 className="font-semibold">{tr("gl_details_title","หลักฐาน ลูกหนี้–เจ้าหนี้ และธนาคาร")}</h3><p className="text-muted-foreground">{posted?tr("gl_details_posted_hint","ยอด GL ผ่านรายการแล้ว เพิ่มผลกระทบยอดได้โดยเก็บประวัติ"):tr("gl_details_draft_hint","รายละเอียดทั้งหมดบันทึกพร้อมใบสำคัญ ไม่สร้างคำสั่งรับ–จ่ายเงินจริง")}</p><Notice error text={error}/>
    <div className="flex flex-wrap gap-2">{([ ["documents",tr("gl_details_ui_19","ดูเอกสารและยอดค้าง")],["statements",tr("gl_details_ui_20","ดู Statement คงเหลือ")],["bank-lines",tr("gl_details_ui_21","ดูบัญชีธนาคารรอจับคู่")] ] as [GLSupportKind,string][]).map(([kind,label])=><Button key={kind} type="button" variant="outline" onClick={()=>setPicker({kind,accept:()=>{}})}>{label}</Button>)}</div>
    {sections.map(([key,title])=>{const rows=detailRows(value,key),canAdd=editable&&(!posted||postedKeys.includes(key));return <details key={key} className="rounded-lg border border-border p-3" open={rows.length>0||undefined}>
      <summary className="cursor-pointer py-2 font-semibold">{title} · {rows.length}</summary>
      <div className="grid gap-3">{rows.map((row,index)=>{const saved=detailRows(original,key).some(old=>row.id?old.id===row.id:old.line_number===row.line_number&&old.bank_account_code===row.bank_account_code);const canEdit=canAdd&&(!posted||!saved||taxKeys.includes(key));return <div key={String(row.id||index)} className="grid gap-2 rounded-lg border border-border bg-muted/20 p-3">
        <div className="flex justify-between gap-2"><span>{tr(`gl_details_section_${key}`,title)} {index+1}</span>{canEdit&&<Button type="button" variant="outline" onClick={()=>onChange({...value,[key]:rows.filter((_,i)=>i!==index)})}>{tr("gl_details_remove_pending","นำรายการที่ยังไม่บันทึกออก")}</Button>}</div>
        <div className="grid gap-3 sm:grid-cols-2">{fields[key].filter(field=>field.show?.(row)??true).map(field=><div key={field.key} className="contents" data-detail-field={`${key}.${index+1}.${field.key}`}><Field label={tr(`gl_detail_${key}_${field.key}`,field.label)}>{renderInput(key,index,row,field,canEdit)}</Field></div>)}</div>
        {key==="vats"&&<VatClaimHint row={row} tr={tr}/>}
      </div>;})}</div>
      {key==="withholdings"&&<p className="text-muted-foreground">{tr("gl_details_withholdings_hint","ฐานภาษีแก้ได้เสมอ แม้ผ่านบัญชีแล้ว — ยอดบัญชีไม่เปลี่ยน ระบบเก็บค่าเดิมไว้ในประวัติ; ภาษีเว้นว่าง = คำนวณ ฐาน × อัตรา ตอนบันทึก")}</p>}
      {key==="vats"&&<p className="text-muted-foreground">{tr("gl_details_vats_hint","ฐานภาษีและภาษีแก้ได้เสมอ แม้ผ่านบัญชีแล้ว — ยอดบัญชีไม่เปลี่ยน ระบบเก็บค่าเดิมไว้ในประวัติ; ภาษีเว้นว่าง = คำนวณ ฐาน × อัตรา ตอนบันทึก; ใบลดหนี้กรอกยอดเป็นบวก รายงาน ภ.พ.30 หักออกให้")}</p>}
      {posted&&key==="allocations"&&<p>{tr("gl_details_reallocate_hint","ถอนและจัดสรรใหม่ในครั้งเดียว ยอด GL เดิมไม่เปลี่ยน")}</p>}
      {canAdd&&<Button type="button" variant="outline" className="mt-3 min-h-11" onClick={()=>onChange({...value,[key]:[...rows,newRow(key,date,branch,booktype)]})}>{tr("gl_details_add","เพิ่ม")}{tr(`gl_details_section_${key}`,title)}</Button>}
      {key==="statement_lines"&&canAdd&&<div className="mt-3 grid gap-2 border-t border-border pt-3"><p>{tr("gl_details_ui_22","CSV: transaction_date (YYYY-MM-DD), direction (1 เข้า / 2 ออก), amount, bank_reference, description")}</p><Button type="button" variant="outline" onClick={()=>setPicker({kind:"bank-accounts",accept:row=>{setImportBank(row);remember("bank-accounts",[row]);}})}>{importBank?supportLabel("bank-accounts",importBank):tr("gl_details_ui_23","เลือกบัญชีสำหรับนำเข้า Statement")}</Button><input type="file" accept=".csv,text/csv" aria-label={tr("gl_details_ui_24","นำเข้า Statement CSV")} disabled={!importBank||importing} onChange={e=>{const file=e.target.files?.[0];if(file)void importCsv(file);e.target.value="";}}/><Button type="button" variant="outline" onClick={()=>downloadText("statement-template.csv","transaction_date,direction,amount,bank_reference,description\n","text/csv;charset=utf-8")}>{tr("gl_details_csv_header","ดาวน์โหลดหัวตาราง CSV")}</Button></div>}
    </details>;})}
    {picker&&<SupportPicker picker={picker} local={local(picker.kind)} onClose={()=>setPicker(null)} onCache={remember}/>}
  </section>;
}
