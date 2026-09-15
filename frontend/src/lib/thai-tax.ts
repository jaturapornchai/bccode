// Thai Tax & Compliance Engine for Thai SMEs & Thai Accounting
// Covers VAT (ภ.พ. 30, ภ.พ. 36, รายงานภาษีขาย/ซื้อ) and WHT (ภ.ง.ด. 2, ภ.ง.ด. 3, ภ.ง.ด. 53, 50 ทวิ)

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

export interface Pp30Summary {
  taxyear: number;
  taxmonth: number;
  totalsales: number;
  zeroratedsales: number;
  exemptsales: number;
  taxablesales: number;
  outputvat: number; // ภาษีขาย
  totalpurchases: number;
  claimablepurchases: number;
  inputvat: number; // ภาษีซื้อ
  nettaxpayable: number; // outputvat - inputvat (บวกคือต้องชำระ, ลบคือชำระเกิน)
  previousoverpayment: number; // ภาษีชำระเกินยกมา
  finaltaxpayable: number; // ภาษีที่ต้องชำระสุทธิ (หรือขอคืน)
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

export function calculatePp30Summary(
  taxYear: number,
  taxMonth: number,
  sales: { taxable: number; zeroRated: number; exempt: number },
  purchases: { claimable: number; exempt: number },
  previousOverpayment: number = 0,
): Pp30Summary {
  const outputvat = Math.round(sales.taxable * 0.07 * 100) / 100;
  const inputvat = Math.round(purchases.claimable * 0.07 * 100) / 100;
  const nettaxpayable = Math.round((outputvat - inputvat) * 100) / 100;
  const finaltaxpayable = Math.round((nettaxpayable - previousOverpayment) * 100) / 100;

  return {
    taxyear: taxYear,
    taxmonth: taxMonth,
    totalsales: sales.taxable + sales.zeroRated + sales.exempt,
    zeroratedsales: sales.zeroRated,
    exemptsales: sales.exempt,
    taxablesales: sales.taxable,
    outputvat,
    totalpurchases: purchases.claimable + purchases.exempt,
    claimablepurchases: purchases.claimable,
    inputvat,
    nettaxpayable,
    previousoverpayment: previousOverpayment,
    finaltaxpayable,
  };
}

export function getSampleTaxRecords(formType: ThaiTaxConfig["formType"]): ThaiTaxRecord[] {
  if (formType === "vat_sale") {
    return [
      {
        id: "tx-s-01",
        docdate: "2026-09-02",
        taxinvoiceno: "INV-202609-001",
        counterpartyname: "บริษัท สยามการค้าปลีก จำกัด (มหาชน)",
        taxid: "0107536000123",
        branchno: "00000",
        isheadoffice: true,
        amountbeforevat: 50000,
        vatamount: 3500,
        totalamount: 53500,
        status: "active",
      },
      {
        id: "tx-s-02",
        docdate: "2026-09-05",
        taxinvoiceno: "INV-202609-002",
        counterpartyname: "ห้างหุ้นส่วนจำกัด ไทยเจริญการช่าง",
        taxid: "0103554009876",
        branchno: "00001",
        isheadoffice: false,
        amountbeforevat: 28000,
        vatamount: 1960,
        totalamount: 29960,
        status: "active",
      },
      {
        id: "tx-s-03",
        docdate: "2026-09-10",
        taxinvoiceno: "INV-202609-003",
        counterpartyname: "บริษัท กรุงเทพพรีเมียม ซัพพลาย จำกัด",
        taxid: "0105558004567",
        branchno: "00000",
        isheadoffice: true,
        amountbeforevat: 125000,
        vatamount: 8750,
        totalamount: 133750,
        status: "active",
      },
    ];
  }

  if (formType === "vat_buy") {
    return [
      {
        id: "tx-b-01",
        docdate: "2026-09-03",
        taxinvoiceno: "TI-889012",
        counterpartyname: "บริษัท นภาอุตสาหกรรม จำกัด",
        taxid: "0105551007890",
        branchno: "00000",
        isheadoffice: true,
        amountbeforevat: 40000,
        vatamount: 2800,
        totalamount: 42800,
        status: "active",
      },
      {
        id: "tx-b-02",
        docdate: "2026-09-08",
        taxinvoiceno: "TAX-54129",
        counterpartyname: "บริษัท แอดวานซ์ ดิจิทัล เซอร์วิส จำกัด",
        taxid: "0105556012345",
        branchno: "00002",
        isheadoffice: false,
        amountbeforevat: 15000,
        vatamount: 1050,
        totalamount: 16050,
        status: "active",
      },
    ];
  }

  // WHT Forms (PND 53, PND 3, 50 Twi)
  return [
    {
      id: "tx-wht-01",
      docdate: "2026-09-04",
      taxinvoiceno: "50TWI-202609-01",
      counterpartyname: "บริษัท โปรคอนซัลแตนท์ แอนด์ เซอร์วิส จำกัด",
      taxid: "0105559011223",
      branchno: "00000",
      isheadoffice: true,
      amountbeforevat: 50000,
      vatamount: 3500,
      totalamount: 53500,
      incometype: "ค่าบริการ (มาตรา 40(8))",
      taxrate: 3,
      whtamount: 1500,
      status: "active",
    },
    {
      id: "tx-wht-02",
      docdate: "2026-09-07",
      taxinvoiceno: "50TWI-202609-02",
      counterpartyname: "นายสมชาย ใจดี",
      taxid: "3100600123456",
      branchno: "00000",
      isheadoffice: true,
      amountbeforevat: 20000,
      vatamount: 0,
      totalamount: 20000,
      incometype: "ค่าเช่าอาคาร (มาตรา 40(5))",
      taxrate: 5,
      whtamount: 1000,
      status: "active",
    },
    {
      id: "tx-wht-03",
      docdate: "2026-09-12",
      taxinvoiceno: "50TWI-202609-03",
      counterpartyname: "บริษัท เอ็กซ์เพรส ทรานสปอร์ต จำกัด",
      taxid: "0105557008899",
      branchno: "00000",
      isheadoffice: true,
      amountbeforevat: 35000,
      vatamount: 0,
      totalamount: 35000,
      incometype: "ค่าขนส่ง (มาตรา 40(8))",
      taxrate: 1,
      whtamount: 350,
      status: "active",
    },
  ];
}
