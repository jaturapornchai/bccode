// Thai Tax & Compliance Engine for Thai SMEs & Thai Accounting
// Covers VAT (ภ.พ. 30, ภ.พ. 36, รายงานภาษีขาย/ซื้อ) and WHT (ภ.ง.ด. 2, ภ.ง.ด. 3, ภ.ง.ด. 53, 50 ทวิ)

import { authFetch } from "@/lib/client-auth-session";

export interface ThaiTaxRecord {
  id: string;
  docdate: string;
  taxinvoiceno: string;
  counterpartyname: string;
  taxid: string;
  branchno: string; // "00000" for Head Office
  isheadoffice: boolean;
  amountbeforevat: number;
  vatamount: number;
  totalamount: number;
  incometype?: string;
  taxrate?: number;
  whtamount?: number;
  remark?: string;
  status: "active" | "cancelled" | "excluded";
}


export interface ThaiTaxConfig {
  route: string;
  code: string;
  title: { th: string; en: string };
  formType: "vat_sale" | "vat_buy" | "pp30" | "pp36" | "pnd2" | "pnd3" | "pnd53" | "50twi" | "wht_received" | "wht_summary" | "deferred_tax";
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
    route: "/report/unreceivedtaxinvoice",
    code: "unreceived_tax_invoice",
    title: { th: "ค่าใช้จ่ายยังไม่ได้รับใบกำกับภาษี", en: "Unreceived Tax Invoices" },
    formType: "vat_buy",
    description: { th: "ทะเบียนติดตามใบกำกับภาษีซื้อที่ยังค้างรับจากคู่ค้า", en: "Pending vendor tax invoices registry" },
    revenueDepartmentFormCode: "TAX-PENDING",
  },
  {
    route: "/report/vatpp30",
    code: "pp30",
    title: { th: "แบบยื่นภาษี ภ.พ.30", en: "PP.30 VAT Return" },
    formType: "pp30",
    description: { th: "แบบแสดงรายการภาษีมูลค่าเพิ่ม ภ.พ.30 นำส่งกรมสรรพากรประจำเดือน", en: "Monthly Value Added Tax Return Form PP.30" },
    revenueDepartmentFormCode: "ภ.พ.30",
  },
  {
    route: "/report/vatpp36",
    code: "pp36",
    title: { th: "แบบยื่น ภ.พ.36", en: "PP.36 Cross-Border VAT" },
    formType: "pp36",
    description: { th: "แบบนำส่งภาษีมูลค่าเพิ่มจากการจ่ายเงินค่าบริการไปต่างประเทศ", en: "Cross-border services VAT remittance form PP.36" },
    revenueDepartmentFormCode: "ภ.พ.36",
  },
  {
    route: "/report/vatpnd2",
    code: "pnd2",
    title: { th: "แบบยื่น ภ.ง.ด.2", en: "PND.2 Withholding Tax Return" },
    formType: "pnd2",
    description: { th: "ภาษีหัก ณ ที่จ่ายเงินได้พึงประเมิน 40(3) และ 40(4) ดอกเบี้ย เงินปันผล ค่าสิทธิ", en: "WHT return for Section 40(3), (4) royalties and interest" },
    revenueDepartmentFormCode: "ภ.ง.ด.2",
  },
  {
    route: "/report/vatpnd3",
    code: "pnd3",
    title: { th: "แบบยื่น ภ.ง.ด.3", en: "PND.3 Personal WHT Return" },
    formType: "pnd3",
    description: { th: "ภาษีหัก ณ ที่จ่ายบุคคลธรรมดา (ค่าเช่า 5%, ค่าบริการ 3%, ค่าวิชาชีพ)", en: "Personal withholding tax return for individuals" },
    revenueDepartmentFormCode: "ภ.ง.ด.3",
  },
  {
    route: "/report/vatpnd53",
    code: "pnd53",
    title: { th: "แบบยื่น ภ.ง.ด.53", en: "PND.53 Corporate WHT Return" },
    formType: "pnd53",
    description: { th: "ภาษีหัก ณ ที่จ่ายนิติบุคคล (ค่าบริการ 3%, ค่าขนส่ง 1%, ค่าเช่า 5%, โฆษณา 2%)", en: "Corporate withholding tax return" },
    revenueDepartmentFormCode: "ภ.ง.ด.53",
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
    title: { th: "ทะเบียนถูกหัก ณ ที่จ่าย", en: "Tax Withheld from Us Register" },
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
  {
    route: "/report/deferredtax",
    code: "deferred_tax",
    title: { th: "ภาษีเงินได้รอการตัดบัญชี", en: "Deferred Income Tax" },
    formType: "deferred_tax",
    description: { th: "การคำนวณและกระทบยอดสินทรัพย์และหนี้สินภาษีเงินได้รอการตัดบัญชี (TAS 12)", en: "Deferred tax assets and liabilities calculation under TAS 12" },
    revenueDepartmentFormCode: "TAS-12",
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

export interface Pp30Summary {
  year: number;
  month: number;
  salestaxable: number;
  saleszerorated: number;
  salesexempt: number;
  outputvat: number;
  purchasetaxable: number;
  inputvat: number;
  netvat: number;
  payable: number;
  creditable: number;
}

const VAT_REGISTER_PATH = "/api/goapi/api/report/tax/vat-register";
const PP30_SUMMARY_PATH = "/api/goapi/api/report/tax/pp30-summary";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function toNumber(value: unknown): number {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim() !== "") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return 0;
}

function toText(value: unknown): string {
  return typeof value === "string" ? value : "";
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
    amountbeforevat: toNumber(row.amountbeforevat),
    vatamount: toNumber(row.vatamount),
    totalamount: toNumber(row.totalamount),
    // backend กรองเอกสารที่ยกเลิก/ตัดออกแล้ว จึงแสดงเป็น active ได้
    status: "active",
  };
}

type PostResult = { ok: true; payload: unknown } | { ok: false; error: string };

async function postApi(path: string, body: unknown): Promise<PostResult> {
  const res = await authFetch(path, {
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
}): Promise<{ records: ThaiTaxRecord[]; total: number; error?: string }> {
  if (!params.holdingcode || !params.businesscode) {
    return { records: [], total: 0, error: "company_required" };
  }

  const result = await postApi(VAT_REGISTER_PATH, params);
  if (!result.ok) {
    return { records: [], total: 0, error: result.error };
  }

  const payload = result.payload;
  if (!isRecord(payload) || !Array.isArray(payload.data)) {
    return { records: [], total: 0, error: "load_failed" };
  }

  const records = payload.data.filter(isRecord).map(toTaxRecord);
  const total = payload.count === undefined ? records.length : toNumber(payload.count);

  return { records, total };
}

export async function fetchPp30Summary(params: {
  holdingcode: string;
  businesscode: string;
  year: number;
  month: number;
}): Promise<{ summary: Pp30Summary | null; error?: string }> {
  if (!params.holdingcode || !params.businesscode) {
    return { summary: null, error: "company_required" };
  }

  const result = await postApi(PP30_SUMMARY_PATH, params);
  if (!result.ok) {
    return { summary: null, error: result.error };
  }

  const payload = result.payload;
  if (!isRecord(payload) || !isRecord(payload.data)) {
    return { summary: null, error: "load_failed" };
  }

  const data = payload.data;

  return {
    summary: {
      year: toNumber(data.year),
      month: toNumber(data.month),
      salestaxable: toNumber(data.salestaxable),
      saleszerorated: toNumber(data.saleszerorated),
      salesexempt: toNumber(data.salesexempt),
      outputvat: toNumber(data.outputvat),
      purchasetaxable: toNumber(data.purchasetaxable),
      inputvat: toNumber(data.inputvat),
      netvat: toNumber(data.netvat),
      payable: toNumber(data.payable),
      creditable: toNumber(data.creditable),
    },
  };
}
