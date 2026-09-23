// Thai Tax & Compliance Engine for Thai SMEs & Thai Accounting
// Covers VAT registers (รายงานภาษีขาย/ซื้อ) and WHT reports + 50 ทวิ; แบบยื่นกรมสรรพากรอยู่ที่ tax-forms.ts

import { apiFetch } from "./client-auth-session";
import { catalogText } from "@/lib/catalog-text";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";

export interface ThaiTaxRecord {
  id: string;
  docdate: string;
  taxinvoiceno: string;
  counterpartyname: string;
  taxid: string;
  branchno: string; // "00000" for Head Office
  isheadoffice: boolean;
  amountbeforevat: string;
  vatamount: string;
  totalamount: string;
  incometype?: string;
  taxrate?: string;
  whtamount?: string;
  remark?: string;
  status: "active" | "cancelled" | "excluded";
}


export interface ThaiTaxConfig {
  route: string;
  code: string;
  title: { th: string; en: string };
  formType: "vat_sale" | "vat_buy" | "50twi" | "wht_received" | "wht_summary";
  description: { th: string; en: string };
  revenueDepartmentFormCode: string;
}

export const THAI_TAX_CONFIGS: ThaiTaxConfig[] = [
  {
    route: "/report/reportvatsale",
    code: "vat_sale",
    title: { th: "รายงานภาษีขาย", en: "Sales Tax Report" },
    formType: "vat_sale",
    description: { th: "รายงานภาษีขายตามมาตรา 87(1) แห่งประมวลรัษฎากร", en: "Sales VAT report pursuant to Section 87(1)" },
    revenueDepartmentFormCode: "ภ.พ. 87(1)",
  },
  {
    route: "/report/reportvatbuy",
    code: "vat_buy",
    title: { th: "รายงานภาษีซื้อ", en: "Purchase Tax Report" },
    formType: "vat_buy",
    description: { th: "รายงานภาษีซื้อตามมาตรา 87(2) แห่งประมวลรัษฎากร", en: "Purchase VAT report pursuant to Section 87(2)" },
    revenueDepartmentFormCode: "ภ.พ. 87(2)",
  },
  {
    route: "/report/whtcertificate",
    code: "50twi",
    title: { th: "หนังสือรับรองหัก ณ ที่จ่าย (50 ทวิ)", en: "Withholding Tax Certificate (50 Twi)" },
    formType: "50twi",
    description: { th: "หนังสือรับรองการหักภาษี ณ ที่จ่าย ตามมาตรา 50 ทวิ แห่งประมวลรัษฎากร", en: "Certificate of withholding tax deduction under Section 50 Twi" },
    revenueDepartmentFormCode: "มาตรา 50 ทวิ",
  },
  {
    route: "/report/whtreceived",
    code: "wht_received",
    title: { th: "รายงานภาษีถูกหัก ณ ที่จ่าย", en: "Tax Withheld from Us Register" },
    formType: "wht_received",
    description: { th: "ทะเบียนเอกสารภาษีที่กิจการถูกลูกค้าหัก ณ ที่จ่ายไว้", en: "Register of withholding tax deducted by customers" },
    revenueDepartmentFormCode: "WHT-RCV",
  },
  {
    route: "/report/wht-reports",
    code: "wht_summary",
    title: { th: "รายงานภาษีหัก ณ ที่จ่าย", en: "Withholding Tax Summary" },
    formType: "wht_summary",
    description: { th: "สรุปภาพรวมภาษีหัก ณ ที่จ่าย ทุกประเภทพร้อมนำส่ง", en: "Comprehensive withholding tax summary across all filing types" },
    revenueDepartmentFormCode: "WHT-ALL",
  },
];

const taxRouteMap = new Map<string, ThaiTaxConfig>(
  THAI_TAX_CONFIGS.map((config) => [config.route, config]),
);

export function isThaiTaxRoute(route: string): boolean {
  const clean = route.split("?")[0];
  return taxRouteMap.has(clean);
}

export function getThaiTaxConfig(route: string): ThaiTaxConfig | undefined {
  const clean = route.split("?")[0];
  return taxRouteMap.get(clean);
}

// ยอดเงินทุกช่องเป็น string ทศนิยม 2 ตำแหน่งจาก backend (decimal) — จอแสดงผลอย่างเดียว ห้ามแปลงเป็น number มาคำนวณ
export interface CompanyHeader {
  code: string;
  name: string;
  taxid: string;
}

export interface VatRegisterSummary {
  amountbeforevat: string;
  vatamount: string;
  totalamount: string;
}

const VAT_REGISTER_PATH = "/api/goapi/api/report/tax/vat-register";
const WHT_REPORT_PATH = "/api/goapi/api/report/tax/wht";

export interface WhtReportRow {
  journalid: string;
  docno: string;
  docdate: string;
  partnercode: string;
  partnername: string;
  taxid: string;
  address: string;
  description: string;
  baseamount: string;
  whtamount: string;
  whtamounttext: string;
  netamount: string;
  ratepercent: string;
  /** recorded = ผู้ใช้บันทึกฐานภาษีในใบสำคัญ, inferred = ระบบประมาณจากบรรทัดบัญชี */
  taxbasesource: string;
  incometype: string;
  condition: number;
  paiddate: string;
  certificateno: string;
}

export interface WhtRateGroup {
  ratepercent: string;
  count: number;
  baseamount: string;
  whtamount: string;
}

export interface WhtReportSummary {
  basetotal: string;
  whttotal: string;
  whttotaltext: string;
  nettotal: string;
  payeecount: number;
  byrate: WhtRateGroup[];
}

export interface WhtReportParams {
  holdingcode: string;
  businesscode: string;
  year: number;
  month: number;
  direction: "paid" | "received";
  forms?: string[];
  limit?: number;
  offset?: number;
}

export interface WhtReportResult {
  rows: WhtReportRow[];
  total: number;
  summary: WhtReportSummary;
  company: CompanyHeader;
  note: string;
  error?: string;
}

const EMPTY_COMPANY: CompanyHeader = { code: "", name: "", taxid: "" };
const EMPTY_WHT_SUMMARY: WhtReportSummary = { basetotal: "0.00", whttotal: "0.00", whttotaltext: "", nettotal: "0.00", payeecount: 0, byrate: [] };
const EMPTY_VAT_SUMMARY: VatRegisterSummary = { amountbeforevat: "0.00", vatamount: "0.00", totalamount: "0.00" };

// fetchWhtReport - รายการภาษีหัก ณ ที่จ่ายจากบัญชีแยกประเภทที่ผ่านรายการจริง
// (backend โยงคู่ค้า/เลขผู้เสียภาษีจากหลักฐานประกอบ กรองตามแบบยื่น และรวมยอดทั้งงวดด้วย decimal)
export async function fetchWhtReport(params: WhtReportParams): Promise<WhtReportResult> {
  const empty = { rows: [], total: 0, summary: EMPTY_WHT_SUMMARY, company: EMPTY_COMPANY, note: "" };
  if (!params.holdingcode || !params.businesscode) {
    return { ...empty, error: "company_required" };
  }
  const result = await postApi(WHT_REPORT_PATH, params);
  if (!result.ok) {
    return { ...empty, error: result.error };
  }
  const payload = result.payload;
  if (!isRecord(payload) || !Array.isArray(payload.data)) {
    return { ...empty, error: "load_failed" };
  }
  const rows = payload.data.filter(isRecord).map((rec): WhtReportRow => ({
    journalid: toText(rec.journalid),
    docno: toText(rec.docno),
    docdate: toText(rec.docdate),
    partnercode: toText(rec.partnercode),
    partnername: toText(rec.partnername),
    taxid: toText(rec.taxid),
    address: toText(rec.address),
    description: toText(rec.description),
    baseamount: toMoney(rec.baseamount),
    whtamount: toMoney(rec.whtamount),
    whtamounttext: toText(rec.whtamounttext),
    netamount: toMoney(rec.netamount),
    ratepercent: toText(rec.ratepercent),
    taxbasesource: toText(rec.taxbasesource),
    incometype: toText(rec.incometype),
    condition: toCount(rec.condition, 0),
    paiddate: toText(rec.paiddate),
    certificateno: toText(rec.certificateno),
  }));
  const summary = isRecord(payload.summary) ? payload.summary : {};
  return {
    rows,
    total: toCount(payload.total, rows.length),
    summary: {
      basetotal: toMoney(summary.basetotal),
      whttotal: toMoney(summary.whttotal),
      whttotaltext: toText(summary.whttotaltext),
      nettotal: toMoney(summary.nettotal),
      payeecount: toCount(summary.payeecount, 0),
      byrate: (Array.isArray(summary.byrate) ? summary.byrate : []).filter(isRecord).map((group) => ({
        ratepercent: toText(group.ratepercent),
        count: toCount(group.count, 0),
        baseamount: toMoney(group.baseamount),
        whtamount: toMoney(group.whtamount),
      })),
    },
    company: toCompany(payload.company),
    note: toText(payload.note),
  };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

const MONEY_TEXT = /^-?\d+(\.\d+)?$/;

// toMoney - รับเฉพาะ string ทศนิยมจาก backend; ค่าอื่น (null/number/ข้อความแปลก) แสดงเป็น 0.00 แทนการเดา
function toMoney(value: unknown): string {
  return typeof value === "string" && MONEY_TEXT.test(value.trim()) ? value.trim() : "0.00";
}

// จำนวนแถว/จำนวนราย (ไม่ใช่เงิน) — เป็น integer จึงใช้ number ได้
function toCount(value: unknown, fallback: number): number {
  return typeof value === "number" && Number.isInteger(value) && value >= 0 ? value : fallback;
}

function toText(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function toCompany(value: unknown): CompanyHeader {
  if (!isRecord(value)) return EMPTY_COMPANY;
  return { code: toText(value.code), name: toText(value.name), taxid: toText(value.taxid) };
}

function toTaxRecord(row: Record<string, unknown>, index: number): ThaiTaxRecord {
  const taxinvoiceno = toText(row.taxinvoiceno);
  const branchno = toText(row.branchno);

  return {
    id: `${taxinvoiceno}#${index}`,
    docdate: toText(row.docdate),
    taxinvoiceno,
    counterpartyname: toText(row.counterpartyname),
    taxid: toText(row.taxid),
    branchno,
    isheadoffice: branchno === "" || branchno === "00000",
    amountbeforevat: toMoney(row.amountbeforevat),
    vatamount: toMoney(row.vatamount),
    totalamount: toMoney(row.totalamount),
    // backend กรองเอกสารที่ยกเลิก/ตัดออกแล้ว จึงแสดงเป็น active ได้
    status: "active",
  };
}

type PostResult = { ok: true; payload: unknown } | { ok: false; error: string };

const WHT_CERTIFICATE_PATH = "/api/goapi/api/report/tax/wht/certificate";

// ข้อมูลใบ 50 ทวิ — ตรงกับ backend/internal/whtcert.Certificate (ยอดเงินเป็น string ทศนิยม ห้าม number)
export interface WhtCertificateParty {
  name: string;
  address: string;
  taxid: string;
}

export interface WhtCertificateIncome {
  type: string;
  paiddate: string;
  amount: string;
  tax: string;
  note: string;
}

export interface WhtCertificateInput {
  bookno: string;
  runno: string;
  sequenceno: string;
  form: string;
  condition: string;
  conditionnote: string;
  issuedate: string;
  archivecopy: boolean;
  replacement: boolean;
  payer: WhtCertificateParty;
  payee: WhtCertificateParty;
  incomes: WhtCertificateIncome[];
}

export type WhtCertificateResult =
  | { ok: true; pdf: Blob }
  | { ok: false; error: string; message?: string; field?: string };

// requestWhtCertificatePdf - backend ตรวจข้อมูล คำนวณยอดรวม/ตัวอักษร และสร้าง PDF บนแบบฟอร์มกรมสรรพากร — จอแค่แสดง
export async function requestWhtCertificatePdf(
  holdingcode: string,
  businesscode: string,
  certificate: WhtCertificateInput,
): Promise<WhtCertificateResult> {
  const res = await apiFetch(WHT_CERTIFICATE_PATH, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ holdingcode, businesscode, certificate }),
  }).catch(() => null);
  if (res === null) return { ok: false, error: "connection_error" };
  if ((res.headers.get("content-type") ?? "").startsWith("application/pdf")) {
    return { ok: true, pdf: await res.blob() };
  }
  if (res.status === 401 || res.status === 403) return { ok: false, error: "unauthorized" };
  const payload: unknown = await res.json().catch(() => null);
  if (isRecord(payload)) {
    return { ok: false, error: toText(payload.code) || "load_failed", message: toText(payload.message) || undefined, field: toText(payload.field) || undefined };
  }
  return { ok: false, error: "load_failed" };
}

async function postApi(path: string, body: unknown): Promise<PostResult> {
  const res = await apiFetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  }).catch(() => null);

  if (res === null) return { ok: false, error: "connection_error" };

  if (!res.ok) {
    return {
      ok: false,
      error: res.status === 401 || res.status === 403 ? "unauthorized" : "load_failed",
    };
  }

  try {
    return { ok: true, payload: await res.json() };
  } catch {
    return { ok: false, error: "load_failed" };
  }
}

export async function fetchVatRegister(params: {
  holdingcode: string;
  businesscode: string;
  year: number;
  month: number;
  type: "sale" | "purchase";
  limit?: number;
  offset?: number;
}): Promise<{ records: ThaiTaxRecord[]; total: number; summary: VatRegisterSummary; error?: string }> {
  if (!params.holdingcode || !params.businesscode) {
    return { records: [], total: 0, summary: EMPTY_VAT_SUMMARY, error: "company_required" };
  }

  const result = await postApi(VAT_REGISTER_PATH, params);
  if (!result.ok) {
    return { records: [], total: 0, summary: EMPTY_VAT_SUMMARY, error: result.error };
  }

  const payload = result.payload;
  if (!isRecord(payload) || !Array.isArray(payload.data)) {
    return { records: [], total: 0, summary: EMPTY_VAT_SUMMARY, error: "load_failed" };
  }

  const records = payload.data.filter(isRecord).map(toTaxRecord);
  const summary = isRecord(payload.summary) ? payload.summary : {};
  return {
    records,
    total: toCount(payload.total, records.length),
    summary: {
      amountbeforevat: toMoney(summary.amountbeforevat),
      vatamount: toMoney(summary.vatamount),
      totalamount: toMoney(summary.totalamount),
    },
  };
}

// 2026-09-16: every user-visible string above also lives in languages.tsv,
// keyed by `<code>.<part>`. The literals stay as the offline fallback.
const catalogKeys: Record<string, string> = {
  "vat_sale.title": "report_vat_sale",
  "vat_sale.description": "tax_sales_vat_report_pursuant_to",
  "vat_buy.title": "report_vat_buy",
  "vat_buy.description": "tax_purchase_vat_report_pursuant_to",
  "50twi.title": "wht_certificate",
  "50twi.description": "tax_certificate_of_withholding_tax_deduction",
  "wht_received.title": "withholding_tax_received",
  "wht_received.description": "tax_register_of_withholding_tax_deducted",
  "wht_summary.title": "withholding_tax_report",
  "wht_summary.description": "tax_comprehensive_withholding_tax_summary_across",
};

export function taxText(
  config: ThaiTaxConfig,
  part: "title" | "description",
  language: LanguageCode,
  dictionary?: BackendLanguageDictionary,
): string {
  return catalogText(catalogKeys, `${config.code}.${part}`, config[part], language, dictionary);
}

// ประเภทเงินได้ของแบบ 50 ทวิ (รหัสตรงกับ backend/internal/whtcert) — ใช้ทั้งจอ 50 ทวิ และรายละเอียดภาษีในใบสำคัญ GL
export const WHT_INCOME_OPTIONS: { value: string; key: string; th: string }[] = [
  { value: "3_tres", key: "wht_cert_ui_income_3_tres", th: "5. ตามมาตรา 3 เตรส (ค่าบริการ ค่าเช่า ค่าขนส่ง ค่าโฆษณา ค่าจ้างทำของ ฯลฯ)" },
  { value: "40_2", key: "wht_cert_ui_income_40_2", th: "2. ค่าธรรมเนียม ค่านายหน้า ฯลฯ 40 (2)" },
  { value: "40_3", key: "wht_cert_ui_income_40_3", th: "3. ค่าแห่งลิขสิทธิ์ ฯลฯ 40 (3)" },
  { value: "40_4a", key: "wht_cert_ui_income_40_4a", th: "4. (ก) ดอกเบี้ย ฯลฯ 40 (4) (ก)" },
  { value: "40_4b_1_1", key: "wht_cert_ui_income_div_1_1", th: "4. (ข) เงินปันผล ได้เครดิตภาษี — กำไรเสียภาษีร้อยละ 30" },
  { value: "40_4b_1_2", key: "wht_cert_ui_income_div_1_2", th: "4. (ข) เงินปันผล ได้เครดิตภาษี — กำไรเสียภาษีร้อยละ 25" },
  { value: "40_4b_1_3", key: "wht_cert_ui_income_div_1_3", th: "4. (ข) เงินปันผล ได้เครดิตภาษี — กำไรเสียภาษีร้อยละ 20" },
  { value: "40_4b_1_4", key: "wht_cert_ui_income_div_1_4", th: "4. (ข) เงินปันผล ได้เครดิตภาษี — อัตราอื่น (ระบุอัตรา)" },
  { value: "40_4b_2_1", key: "wht_cert_ui_income_div_2_1", th: "4. (ข) เงินปันผล ไม่ได้เครดิต — กิจการได้รับยกเว้นภาษี" },
  { value: "40_4b_2_2", key: "wht_cert_ui_income_div_2_2", th: "4. (ข) เงินปันผล ไม่ได้เครดิต — เงินปันผลที่ได้รับยกเว้น" },
  { value: "40_4b_2_3", key: "wht_cert_ui_income_div_2_3", th: "4. (ข) เงินปันผล ไม่ได้เครดิต — หักผลขาดทุนยกมาไม่เกิน 5 ปี" },
  { value: "40_4b_2_4", key: "wht_cert_ui_income_div_2_4", th: "4. (ข) เงินปันผล ไม่ได้เครดิต — วิธีส่วนได้เสีย (equity method)" },
  { value: "40_4b_2_5", key: "wht_cert_ui_income_div_2_5", th: "4. (ข) เงินปันผล ไม่ได้เครดิต — อื่น ๆ (ระบุ)" },
  { value: "other", key: "wht_cert_ui_income_other", th: "6. อื่น ๆ (ระบุ)" },
];
