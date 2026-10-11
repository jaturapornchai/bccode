import { apiFetch } from "./client-auth-session";
import { blankAmountAsZero } from "./general-ledger";

// ยอดเงิน/จำนวน/ราคา/ภาษีเป็นข้อความทศนิยม (backend shopspring/decimal, Rule 20) เช่น "1070" หรือ "0.125";
// แสดงผลด้วย formatAmount และคำนวณตัวอย่างด้วย erpLineAmount/erpDocTotals เท่านั้น — ห้าม Number()/parseFloat
export interface ErpDetailItem {
  linenumber: number;
  itemcode: string;
  itemnames?: { code: string; name: string }[];
  itemname?: string;
  barcode?: string;
  unitcode: string;
  unitname?: string;
  qty: string;
  price: string;
  discount?: string;
  discountamount?: string;
  sumamount: string;
  vatcal?: number;
  docref?: string;
  // ใบจ่ายชำระหนี้ตัดหนี้รายบรรทัด: id แถวหนี้ในบัญชีย่อยเจ้าหนี้ + เลขที่ใบเตรียมจ่ายที่ดึงมา (reftype backend ตั้งให้)
  debt_document_id?: string;
  paybillno?: string;
  reftype?: number;
}

export interface ErpTransactionDoc {
  id?: string;
  docno: string;
  docdatetime: string;
  transflag?: number;
  custcode?: string;
  custname?: string;
  custnames?: { code: string; name: string }[];
  description?: string;
  totalamount: string;
  totalvalue?: string;
  totaldiscount?: string;
  totalvatvalue?: string;
  totalbeforevat?: string;
  totalaftervat?: string;
  vattype?: number;
  vatrate?: string;
  status?: number; // 0=Draft, 1=Approved/Completed, 2=Cancelled
  // คีย์ตรงกับ backend/internal/transaction/httpapi/models.go: approvalstatus เป็น 'approved' หรือ '' (ว่าง = รออนุมัติ)
  approvalstatus?: string;
  approvedby?: string;
  approvedat?: string;
  cancelledby?: string;
  cancelledat?: string;
  cancelreason?: string;
  createdby?: string;
  isclosed?: boolean;
  iscancel?: boolean;
  branchcode?: string;
  remark?: string;
  // เอกสารหนี้เจ้าหนี้ (ERP_DUE_DATE_MODULES): ไม่ส่ง creditday = ใช้วันเครดิตของเจ้าหนี้ในทะเบียน, ไม่ส่ง duedate = วันที่เอกสาร + วันเครดิต
  creditday?: number;
  duedate?: string;
  // ผลลงบัญชีตอนบันทึก (backend ตั้งให้): posted | no_mapping | no_amount; gldocno = เลขที่ใบสำคัญเมื่อ posted
  glstatus?: string;
  gldocno?: string;
  details: ErpDetailItem[];
}

// ทศนิยมตรงตัวด้วย BigInt (ไม่มีไลบรารีทศนิยมใน package.json): units × 10^-scale
type ExactDecimal = { units: bigint; scale: number };
const ZERO: ExactDecimal = { units: 0n, scale: 0 };
const decimalText = /^([+-]?)([0-9]*)(?:[.]([0-9]*))?(?:e([+-]?[0-9]+))?$/i;

// ข้อความว่าง/ผิดรูป = ศูนย์ (ช่องที่ผู้ใช้กำลังพิมพ์) — รับรูป 1e5 แบบที่ backend รับ
function exact(value: string | undefined): ExactDecimal {
  const match = decimalText.exec((value ?? "").trim());
  if (!match || (!match[2] && !match[3])) return ZERO;
  const exponent = Number(match[4] ?? 0);
  if (Math.abs(exponent) > 30) return ZERO;
  const fraction = match[3] ?? "";
  let units = BigInt((match[2] || "0") + fraction);
  let scale = fraction.length - exponent;
  if (scale < 0) {
    units *= 10n ** BigInt(-scale);
    scale = 0;
  }
  return { units: match[1] === "-" ? -units : units, scale };
}
function addExact(a: ExactDecimal, b: ExactDecimal): ExactDecimal {
  const scale = Math.max(a.scale, b.scale);
  return { units: a.units * 10n ** BigInt(scale - a.scale) + b.units * 10n ** BigInt(scale - b.scale), scale };
}
function negate(a: ExactDecimal): ExactDecimal {
  return { units: -a.units, scale: a.scale };
}
function mulExact(a: ExactDecimal, b: ExactDecimal): ExactDecimal {
  return { units: a.units * b.units, scale: a.scale + b.scale };
}
// ปัดครึ่งออกจากศูนย์เหมือน decimal.Round ของ backend: 1.005 → 1.01
function roundExact(a: ExactDecimal, places: number): ExactDecimal {
  if (a.scale <= places) return a;
  const divisor = 10n ** BigInt(a.scale - places);
  const abs = a.units < 0n ? -a.units : a.units;
  const rounded = abs / divisor + ((abs % divisor) * 2n >= divisor ? 1n : 0n);
  return { units: a.units < 0n ? -rounded : rounded, scale: places };
}
// ข้อความทศนิยมแบบที่ backend ตอบ (ตัดศูนย์ท้าย): 1070.00 → "1070"
function exactString(a: ExactDecimal): string {
  const abs = a.units < 0n ? -a.units : a.units;
  const digits = abs.toString().padStart(a.scale + 1, "0");
  const whole = digits.slice(0, digits.length - a.scale);
  const fraction = digits.slice(digits.length - a.scale).replace(/0+$/, "");
  return `${a.units < 0n ? "-" : ""}${whole}${fraction ? `.${fraction}` : ""}`;
}

/** ยอดของบรรทัด = จำนวน × ราคา − ส่วนลด (ติดลบให้เป็นศูนย์) ไม่ปัดรายบรรทัด */
export function erpLineAmount(item: Pick<ErpDetailItem, "qty" | "price" | "discountamount">): string {
  const amount = addExact(mulExact(exact(item.qty), exact(item.price)), negate(exact(item.discountamount)));
  return exactString(amount.units < 0n ? ZERO : amount);
}

/**
 * ตัวอย่างยอดรวมระหว่างแก้เอกสาร (ยอดจริงหลังบันทึกคือที่ backend ตอบกลับ) — กติกาปัดเศษเดิมของหน้าจอ:
 * ก่อนภาษี = ผลรวมยอดบรรทัดที่ไม่ปัด; ภาษี (vattype 1) = ก่อนภาษี × อัตรา ÷ 100 ปัดที่ระดับเอกสาร 2 ตำแหน่ง ครึ่งออกจากศูนย์;
 * รวมทั้งสิ้น = ก่อนภาษี + ภาษี — ทศนิยมตรงตัว 0.1 + 0.2 = 0.3
 */
export function erpDocTotals(
  details: Pick<ErpDetailItem, "sumamount">[],
  vattype?: number,
  vatrate = "7",
): Pick<ErpTransactionDoc, "totalbeforevat" | "totalvatvalue" | "totalamount"> {
  const before = details.reduce((sum, item) => addExact(sum, exact(item.sumamount)), ZERO);
  const rate = exact(vatrate);
  const vat = vattype === 1 ? roundExact(mulExact(before, { units: rate.units, scale: rate.scale + 2 }), 2) : ZERO;
  return { totalbeforevat: exactString(before), totalvatvalue: exactString(vat), totalamount: exactString(addExact(before, vat)) };
}

/** มูลค่าก่อนภาษีของเอกสาร: ยอดที่บันทึกไว้ ถ้าไม่มี (backend ละช่องที่เป็นศูนย์) = ยอดรวม − ภาษี แบบตรงตัว */
export function erpBeforeVatAmount(doc: Pick<ErpTransactionDoc, "totalamount" | "totalbeforevat" | "totalvatvalue">): string {
  if (doc.totalbeforevat !== undefined) return doc.totalbeforevat;
  return exactString(addExact(exact(doc.totalamount), negate(exact(doc.totalvatvalue))));
}

/**
 * ยอดเงินสำหรับแสดงในตัวอย่างระหว่างแก้ (แสดงผลอย่างเดียว ไม่เปลี่ยนค่าที่ส่งบันทึก): ปัด 2 ตำแหน่งครึ่งออกจากศูนย์
 * แบบเดียวกับที่ PostgreSQL ปัดยอดบรรทัดลง docdetail.sumamount numeric(18,2) — 152415.7763907942 → "152415.78"
 */
export function erpMoneyDisplay(value: string | undefined): string {
  return exactString(roundExact(exact(value), 2));
}

/** ผลรวมทศนิยมตรงตัวของยอดที่มีเครื่องหมาย (sign −1 = หักออก) ไม่ปัด — ยอดรวมตัวอย่างของใบเตรียมจ่ายชำระหนี้ */
export function erpSignedSum(items: readonly { amount: string | undefined; sign: number }[]): string {
  return exactString(items.reduce((sum, item) => addExact(sum, item.sign < 0 ? negate(exact(item.amount)) : exact(item.amount)), ZERO));
}

/** เครื่องหมายของข้อความทศนิยม (ว่าง/ผิดรูป = 0) โดยไม่แปลงเป็น number */
export function erpDecimalSign(value: string | undefined): -1 | 0 | 1 {
  const units = exact(value).units;
  return units > 0n ? 1 : units < 0n ? -1 : 0;
}

/** วันครบกำหนด = วันที่เอกสาร (YYYY-MM-DD) + วันเครดิต แบบเดียวกับ backend resolveDueDate — นับวันปฏิทิน ไม่ขึ้นกับเขตเวลาเครื่อง */
export function erpDueDate(docDate: string, creditDays: number): string {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(docDate);
  if (!match || !Number.isInteger(creditDays) || creditDays < 0) return "";
  return new Date(Date.UTC(Number(match[1]), Number(match[2]) - 1, Number(match[3]) + creditDays)).toISOString().slice(0, 10);
}

/** ช่องตัวเลขที่ผู้ใช้ลบจนว่างส่งเป็น "0" — backend รับเฉพาะข้อความทศนิยมที่ถูกต้อง; วันครบกำหนดว่าง = ไม่ส่ง (backend คำนวณ) */
export function erpDocForSave(doc: Partial<ErpTransactionDoc>): Partial<ErpTransactionDoc> {
  const duedate = doc.duedate?.trim() || undefined;
  if (!doc.details) return { ...doc, duedate };
  return {
    ...doc,
    duedate,
    details: doc.details.map((item) => ({
      ...item,
      qty: blankAmountAsZero(item.qty),
      price: blankAmountAsZero(item.price),
      sumamount: blankAmountAsZero(item.sumamount),
    })),
  };
}

export type ErpDocLifecycle = "draft" | "approved" | "cancelled";

// สถานะของเอกสารจากคีย์ที่ backend ส่ง — ยกเลิกชนะอนุมัติ; เอกสารที่อนุมัติหรือยกเลิกแล้วแก้ไข/ลบไม่ได้ (docLockedError ฝั่ง backend)
export function erpDocLifecycle(doc: Pick<ErpTransactionDoc, "status" | "iscancel" | "approvalstatus">): ErpDocLifecycle {
  if (doc.iscancel || doc.status === 2) return "cancelled";
  if (doc.approvalstatus === "approved" || doc.status === 1) return "approved";
  return "draft";
}

export interface ErpModuleConfig {
  route: string;
  moduleKey: string;
  apiPath: string;
  domain: "inventory" | "sales" | "purchase" | "ar" | "ap" | "cash-bank";
  title: { th: string; en: string };
  counterpartyLabel: { th: string; en: string };
  counterpartyKey: string;
  hasLineItems: boolean;
}

export const ERP_MODULE_CONFIGS: ErpModuleConfig[] = [
  // 1. สินค้า (Inventory / IC)
  {
    route: "/transaction/stockbalance",
    moduleKey: "stock_balance",
    apiPath: "stock-balance",
    domain: "inventory",
    title: { th: "บันทึกสินค้ายกมา", en: "Stock Beginning Balance" },
    counterpartyLabel: { th: "คลัง / แผนก", en: "Warehouse / Dept" },
    counterpartyKey: "cp_warehouse_dept",
    hasLineItems: true,
  },
  {
    route: "/transaction/stockreceiveproduct",
    moduleKey: "stock_receive",
    apiPath: "stock-receive-product",
    domain: "inventory",
    title: { th: "บันทึกรับสินค้าสำเร็จรูป", en: "Stock Receive" },
    counterpartyLabel: { th: "คลังปลายทาง", en: "Target Warehouse" },
    counterpartyKey: "cp_destination_warehouse",
    hasLineItems: true,
  },
  {
    route: "/transaction/stockpickupproduct",
    moduleKey: "stock_pickup",
    apiPath: "stock-pickup-product",
    domain: "inventory",
    title: { th: "บันทึกเบิกใช้สินค้า,วัตถุดิบ", en: "Stock Issue / Pickup" },
    counterpartyLabel: { th: "แผนกผู้ขอเบิก", en: "Requesting Dept" },
    counterpartyKey: "cp_requesting_dept",
    hasLineItems: true,
  },
  {
    route: "/transaction/stockreturnproduct",
    moduleKey: "stock_return",
    apiPath: "stock-return-product",
    domain: "inventory",
    title: { th: "บันทึกรับคืนสินค้า,วัตถุดิบ", en: "Stock Return" },
    counterpartyLabel: { th: "คลังรับคืน", en: "Receiving Warehouse" },
    counterpartyKey: "cp_receiving_warehouse",
    hasLineItems: true,
  },
  {
    route: "/transaction/stocktransfer",
    moduleKey: "stock_transfer",
    apiPath: "stock-transfer",
    domain: "inventory",
    title: { th: "บันทึกโอนสินค้าระหว่างคลัง", en: "Stock Transfer" },
    counterpartyLabel: { th: "คลังปลายทาง", en: "Destination Warehouse" },
    counterpartyKey: "cp_destination_warehouse",
    hasLineItems: true,
  },
  {
    route: "/transaction/adjust",
    moduleKey: "stock_adjust",
    apiPath: "stock-adjustment",
    domain: "inventory",
    title: { th: "บันทึกปรับปรุงสินค้าหลังตรวจนับ", en: "Stock Adjustment" },
    counterpartyLabel: { th: "คลังที่ตรวจนับ", en: "Audited Warehouse" },
    counterpartyKey: "cp_count_warehouse",
    hasLineItems: true,
  },

  // 2. ขาย (Sales / BILL / OE)
  {
    route: "/transaction/quotation",
    moduleKey: "quotation",
    apiPath: "quotation",
    domain: "sales",
    title: { th: "บันทึกใบเสนอราคาสินค้า", en: "Quotation" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    counterpartyKey: "customer",
    hasLineItems: true,
  },
  {
    route: "/transaction/saleorder",
    moduleKey: "sale_order",
    apiPath: "sale-order",
    domain: "sales",
    title: { th: "บันทึกใบสั่งขาย/สั่งจองสินค้า", en: "Sales Order" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    counterpartyKey: "customer",
    hasLineItems: true,
  },
  {
    route: "/transaction/sale",
    moduleKey: "sale_invoice",
    apiPath: "sale-invoice",
    domain: "sales",
    title: { th: "บันทึกขายสินค้า,บริการ", en: "Sale & Dispatch" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    counterpartyKey: "customer",
    hasLineItems: true,
  },
  {
    route: "/transaction/salereturn",
    moduleKey: "sale_return",
    apiPath: "sale-invoice-return",
    domain: "sales",
    title: { th: "บันทึกใบรับคืนสินค้า,ลดหนี้", en: "Sale Return" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    counterpartyKey: "customer",
    hasLineItems: true,
  },
  {
    route: "/transaction/debitnote",
    moduleKey: "debit_note",
    apiPath: "ardebitnote",
    domain: "sales",
    title: { th: "บันทึกใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้)", en: "Debit Note" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    counterpartyKey: "customer",
    hasLineItems: true,
  },

  // 3. ซื้อ (Purchase / PO)
  {
    route: "/transaction/purchaserequisition",
    moduleKey: "purchase_requisition",
    apiPath: "purchase-requisition",
    domain: "purchase",
    title: { th: "บันทึกใบเสนอซื้อสินค้า", en: "Purchase Requisition" },
    counterpartyLabel: { th: "เจ้าหนี้ที่เสนอราคา", en: "Proposed Vendor" },
    counterpartyKey: "cp_proposed_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchaseorder",
    moduleKey: "purchase_order",
    apiPath: "purchase-order",
    domain: "purchase",
    title: { th: "บันทึกใบสั่งซื้อสินค้า", en: "Purchase Order" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchase",
    moduleKey: "purchase_invoice",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "บันทึกซื้อสินค้า,บริการ", en: "Purchase Invoice / Goods Receipt" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/expense",
    moduleKey: "expense_record",
    apiPath: "expense",
    domain: "purchase",
    title: { th: "บันทึกจ่ายเงินอื่นๆ", en: "Expense Record" },
    counterpartyLabel: { th: "ผู้รับเงิน / เจ้าหนี้", en: "Payee / Vendor" },
    counterpartyKey: "cp_payee_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasereturn",
    moduleKey: "purchase_return",
    apiPath: "purchase-return",
    domain: "purchase",
    title: { th: "บันทึกส่งคืนสินค้า", en: "Purchase Return" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasecreditnote",
    moduleKey: "purchase_credit_note",
    apiPath: "apdebitnote",
    domain: "purchase",
    title: { th: "บันทึกเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้)", en: "Vendor Credit Note" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasedebitnote",
    moduleKey: "purchase_debit_note",
    apiPath: "apcreditnote",
    domain: "purchase",
    title: { th: "บันทึกส่งคืนสินค้า/ลดหนี้", en: "Vendor Debit Note" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasereducedebt",
    moduleKey: "purchase_reduce_debt",
    apiPath: "apcreditnote",
    domain: "purchase",
    title: { th: "บันทึกลดหนี้", en: "Purchase Debt Reduction" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },

  // 4. ลูกหนี้ (AR)
  {
    route: "/debtorbeginningbalance",
    moduleKey: "debtor_beginning_balance",
    apiPath: "debtorbeginningbalance",
    domain: "ar",
    title: { th: "บันทึกเอกสารลูกหนี้ยกมา", en: "Debtor Beginning Balance" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    counterpartyKey: "cp_debtor",
    hasLineItems: false,
  },
  {
    route: "/transaction/billingnote",
    moduleKey: "billing_note",
    apiPath: "billingnote",
    domain: "ar",
    title: { th: "บันทึกใบวางบิล", en: "Billing Note" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    counterpartyKey: "cp_debtor",
    hasLineItems: true,
  },
  {
    route: "/transaction/paid",
    moduleKey: "ar_receipt",
    apiPath: "paid",
    domain: "ar",
    title: { th: "บันทึกใบเสร็จรับเงิน/รับชำระหนี้", en: "Debtor Payment Receipt" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    counterpartyKey: "cp_debtor",
    hasLineItems: true,
  },
  {
    route: "/transaction/arotherdebt",
    moduleKey: "ar_other_debt",
    apiPath: "receivableother",
    domain: "ar",
    title: { th: "บันทึกเอกสารตั้งลูกหนี้อื่นๆ", en: "Other Receivables" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    counterpartyKey: "cp_debtor",
    hasLineItems: false,
  },
  {
    route: "/transaction/arbaddebt",
    moduleKey: "ar_bad_debt",
    apiPath: "arbaddebt",
    domain: "ar",
    title: { th: "บันทึกตัดหนี้สูญ(ลูกหนี้)", en: "Write Off Receivables" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    counterpartyKey: "cp_debtor",
    hasLineItems: false,
  },

  // 5. เจ้าหนี้ (AP)
  {
    route: "/creditorbeginningbalance",
    moduleKey: "creditor_beginning_balance",
    apiPath: "creditorbeginningbalance",
    domain: "ap",
    title: { th: "บันทึกเอกสารเจ้าหนี้ยกมา", en: "Creditor Beginning Balance" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: false,
  },
  {
    route: "/transaction/apbillingreceipt",
    moduleKey: "ap_billing_receipt",
    apiPath: "apbillingreceipt",
    domain: "ap",
    title: { th: "บันทึกใบรับวางบิล", en: "Supplier Billing Receipt" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/pay",
    moduleKey: "ap_payment",
    apiPath: "pay",
    domain: "ap",
    title: { th: "บันทึกจ่ายชำระหนี้", en: "Creditor Settlement" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/apotherdebt",
    moduleKey: "ap_other_debt",
    apiPath: "apotherdebt",
    domain: "ap",
    title: { th: "บันทึกเอกสารตั้งเจ้าหนี้อื่นๆ", en: "Other Payables" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: false,
  },
  {
    route: "/transaction/apbaddebt",
    moduleKey: "ap_bad_debt",
    apiPath: "apbaddebt",
    domain: "ap",
    title: { th: "บันทึกตัดหนี้สูญ(เจ้าหนี้)", en: "Write Off Payables" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: false,
  },

  // 6. เงินสดและธนาคาร (Cash & Bank)
  {
    route: "/transaction/accounttransfer",
    moduleKey: "account_transfer",
    apiPath: "banktransfer",
    domain: "cash-bank",
    title: { th: "บันทึกโอนเงินระหว่างธนาคาร", en: "Account Transfer" },
    counterpartyLabel: { th: "บัญชีปลายทาง", en: "Destination Account" },
    counterpartyKey: "cp_destination_account",
    hasLineItems: false,
  },
  {
    route: "/transaction/employeeadvance",
    moduleKey: "employee_advance",
    apiPath: "employeeadvance",
    domain: "cash-bank",
    title: { th: "บันทึกขอเบิกเงินทดรองจ่าย", en: "Employee Advance" },
    counterpartyLabel: { th: "พนักงานผู้ทดรอง", en: "Employee" },
    counterpartyKey: "cp_employee",
    hasLineItems: false,
  },

  // 7. จัดซื้อเพิ่มเติมและต้นทุน (Procurement & Landed Cost)
  {
    route: "/transaction/rfq",
    moduleKey: "rfq",
    apiPath: "rfq",
    domain: "purchase",
    title: { th: "บันทึกใบสืบราคาสินค้ารวม", en: "Request for Quotation" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/rfqpricetable",
    moduleKey: "rfq_price_table",
    apiPath: "rfq",
    domain: "purchase",
    title: { th: "บันทึกใบตารางราคารวม", en: "Consolidated Price Table" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasepartial",
    moduleKey: "purchase_partial",
    apiPath: "goodsreceive",
    domain: "purchase",
    title: { th: "บันทึกรับสินค้าจากการซื้อ", en: "Partial Goods Receipt" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/accrualreceive",
    moduleKey: "accrual_receive",
    apiPath: "invoicereceive",
    domain: "purchase",
    title: { th: "บันทึกตั้งหนี้จากการซื้อ", en: "Accrued Payables from Receipt" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/landedcost",
    moduleKey: "landed_cost",
    apiPath: "landedcost",
    domain: "purchase",
    title: { th: "บันทึกต้นทุนแฝง (Weight cost)", en: "Landed Cost Allocation" },
    counterpartyLabel: { th: "ใบรับสินค้า / เจ้าหนี้", en: "Goods Receipt / Vendor" },
    counterpartyKey: "cp_goods_receipt_vendor",
    hasLineItems: true,
  },
  {
    route: "/transaction/advancepayment",
    moduleKey: "advance_payment",
    apiPath: "advancepayment",
    domain: "purchase",
    title: { th: "บันทึกใบจ่ายเงินล่วงหน้า", en: "Advance Payment" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: false,
  },
  {
    route: "/transaction/advancepaymentrefund",
    moduleKey: "advance_payment_refund",
    apiPath: "advancepaymentrefund",
    domain: "purchase",
    title: { th: "บันทึกรับคืนเงินจ่ายล่วงหน้า", en: "Advance Payment Refund" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: false,
  },
  {
    route: "/transaction/deposit",
    moduleKey: "vendor_deposit",
    apiPath: "deposit",
    domain: "purchase",
    title: { th: "บันทึกใบจ่ายเงินมัดจำ", en: "Supplier Deposit" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    counterpartyKey: "cp_vendor",
    hasLineItems: false,
  },

  // 8. ขายเพิ่มเติมและมัดจำ (Sales & Deposits)
  {
    route: "/transaction/paidadvance",
    moduleKey: "customer_paid_advance",
    apiPath: "receiveadvance",
    domain: "sales",
    title: { th: "บันทึกใบรับเงินล่วงหน้า", en: "Customer Advance Payment" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    counterpartyKey: "customer",
    hasLineItems: false,
  },
  {
    route: "/transaction/paidadvancerefund",
    moduleKey: "customer_paid_advance_refund",
    apiPath: "receiveadvancerefund",
    domain: "sales",
    title: { th: "บันทึกคืนเงินรับล่วงหน้า", en: "Customer Advance Refund" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    counterpartyKey: "customer",
    hasLineItems: false,
  },
  {
    route: "/transaction/receivedeposit",
    moduleKey: "customer_deposit",
    apiPath: "receivedeposit",
    domain: "sales",
    title: { th: "บันทึกใบรับเงินมัดจำ", en: "Customer Deposit" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    counterpartyKey: "customer",
    hasLineItems: false,
  },
  {
    route: "/transaction/temporaryreceipt",
    moduleKey: "temporary_receipt",
    apiPath: "temporaryreceipt",
    domain: "sales",
    title: { th: "บันทึกใบเสร็จชั่วคราว", en: "Temporary Receipt" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    counterpartyKey: "customer",
    hasLineItems: true,
  },

  // 9. เจ้าหนี้เพิ่มเติม (Payables Extensions)

  // 10. สต็อกเพิ่มเติมและตรวจนับ (Inventory Operations & Requests)
  {
    route: "/transaction/stockcount",
    moduleKey: "stock_count",
    apiPath: "stock-count",
    domain: "inventory",
    title: { th: "บันทึกผลการตรวจนับสินค้า", en: "Physical Stock Count" },
    counterpartyLabel: { th: "คลังที่ตรวจนับ", en: "Count Warehouse" },
    counterpartyKey: "cp_count_warehouse",
    hasLineItems: true,
  },
  {
    route: "/inventory/issue-request",
    moduleKey: "inventory_issue_request",
    apiPath: "stock-issue-request",
    domain: "inventory",
    title: { th: "บันทึกใบขอเบิกใช้สินค้า,วัตถุดิบ", en: "Material Requisition" },
    counterpartyLabel: { th: "แผนกผู้ขอเบิก", en: "Requesting Dept" },
    counterpartyKey: "cp_requesting_dept",
    hasLineItems: true,
  },
  {
    route: "/inventory/transfer-request",
    moduleKey: "inventory_transfer_request",
    apiPath: "stock-transfer-request",
    domain: "inventory",
    title: { th: "บันทึกใบขอโอนสินค้า", en: "Transfer Requisition" },
    counterpartyLabel: { th: "คลังปลายทาง", en: "Destination Warehouse" },
    counterpartyKey: "cp_destination_warehouse",
    hasLineItems: true,
  },
  {
    route: "/inventory/goods-inspection",
    moduleKey: "inventory_goods_inspection",
    apiPath: "goods-inspection",
    domain: "inventory",
    title: { th: "ใบตรวจรับสินค้า", en: "Goods Inspection" },
    counterpartyLabel: { th: "เจ้าหนี้ / ผู้ส่งมอบ", en: "Vendor / Carrier" },
    counterpartyKey: "cp_vendor_carrier",
    hasLineItems: true,
  },
  {
    route: "/inventory/count-sheet",
    moduleKey: "inventory_count_sheet",
    apiPath: "stock-count",
    domain: "inventory",
    title: { th: "บันทึกเอกสารเพื่อตรวจนับสินค้า", en: "Count Sheet" },
    counterpartyLabel: { th: "คลัง / ที่เก็บ", en: "Warehouse / Zone" },
    counterpartyKey: "cp_warehouse_location",
    hasLineItems: true,
  },

  // 11. การเงิน เงินสด ธนาคาร บัตรเครดิต และเช็ค (Cash, Banking & Cheques)
  {
    route: "/transaction/otherincomereceipt",
    moduleKey: "other_income_receipt",
    apiPath: "otherincome",
    domain: "cash-bank",
    title: { th: "บันทึกรับเงินอื่นๆ", en: "Other Income Receipt" },
    counterpartyLabel: { th: "ผู้ชำระเงิน", en: "Payer" },
    counterpartyKey: "cp_payer",
    hasLineItems: true,
  },
  {
    route: "/banking/cash-deposit",
    moduleKey: "cash_deposit",
    apiPath: "cashdeposit",
    domain: "cash-bank",
    title: { th: "บันทึกนำฝากเงินสด", en: "Cash Deposit" },
    counterpartyLabel: { th: "บัญชีธนาคารปลายทาง", en: "Target Bank Account" },
    counterpartyKey: "cp_target_bank_account",
    hasLineItems: false,
  },
  {
    route: "/banking/cash-withdrawal",
    moduleKey: "cash_withdrawal",
    apiPath: "cashwithdrawal",
    domain: "cash-bank",
    title: { th: "บันทึกถอนเงินสด", en: "Cash Withdrawal" },
    counterpartyLabel: { th: "บัญชีธนาคารถอน", en: "Source Bank Account" },
    counterpartyKey: "cp_source_bank_account",
    hasLineItems: false,
  },
  {
    route: "/banking/charges",
    moduleKey: "bank_charges",
    apiPath: "bankexpense",
    domain: "cash-bank",
    title: { th: "บันทึกค่าใช้จ่ายธนาคาร", en: "Bank Charges" },
    counterpartyLabel: { th: "ธนาคาร", en: "Bank" },
    counterpartyKey: "bank",
    hasLineItems: false,
  },
  {
    route: "/banking/income",
    moduleKey: "bank_income",
    apiPath: "bankincome",
    domain: "cash-bank",
    title: { th: "บันทึกรายได้จากธนาคาร", en: "Bank Income" },
    counterpartyLabel: { th: "ธนาคาร", en: "Bank" },
    counterpartyKey: "bank",
    hasLineItems: false,
  },
  {
    route: "/banking/cards/receipts",
    moduleKey: "card_receipts",
    apiPath: "creditcard",
    domain: "cash-bank",
    title: { th: "บันทึกบัตรเครดิต", en: "Credit Card Receipts Register" },
    counterpartyLabel: { th: "เครื่อง EDC / บัตร", en: "EDC Terminal / Card" },
    counterpartyKey: "cp_edc_terminal_card",
    hasLineItems: false,
  },
  {
    route: "/banking/cards/settlement",
    moduleKey: "card_settlement",
    apiPath: "creditcardsettle",
    domain: "cash-bank",
    title: { th: "บันทึกขึ้นเงินบัตรเครดิต", en: "Credit Card Settlement" },
    counterpartyLabel: { th: "ธนาคารรับชำระ", en: "Acquirer Bank" },
    counterpartyKey: "cp_acquirer_bank",
    hasLineItems: false,
  },
  {
    route: "/banking/cards/cancellation",
    moduleKey: "card_cancellation",
    apiPath: "creditcardcancel",
    domain: "cash-bank",
    title: { th: "บันทึกยกเลิกบัตรเครดิต", en: "Credit Card Receipt Void" },
    counterpartyLabel: { th: "ลูกค้า / บัตร", en: "Customer / Card" },
    counterpartyKey: "cp_customer_card",
    hasLineItems: false,
  },
];

const routeMap = new Map<string, ErpModuleConfig>(
  ERP_MODULE_CONFIGS.map((config) => [config.route, config]),
);

export function isErpTransactionRoute(route: string): boolean {
  const clean = route.split("?")[0];
  return routeMap.has(clean);
}

export function getErpModuleConfig(route: string): ErpModuleConfig | undefined {
  const clean = route.split("?")[0];
  return routeMap.get(clean);
}

/**
 * A 404 here is not a failure the user can retry: the document type has no
 * endpoint in this backend build, so the screen is a dead end today. Saying
 * "โหลดข้อมูลไม่สำเร็จ" sends people looking for a problem that is not theirs —
 * `node tools/probe-endpoints.mjs` lists which modules are in this state.
 */
function statusErrorKey(status: number, fallback: string, forbiddenKey = "unauthorized"): string {
  if (status === 401) return "unauthorized";
  if (status === 403) return forbiddenKey;
  if (status === 404) return "module_not_available";
  return fallback;
}

function userDetail(body: { message?: unknown }): string | undefined {
  return typeof body.message === "string" && body.message.trim() ? body.message : undefined;
}

export async function fetchErpTransactions(
  config: ErpModuleConfig,
  params: { q?: string; offset?: number; limit?: number; status?: string } = {},
): Promise<{ items: ErpTransactionDoc[]; total: number; error?: string }> {
  const query = new URLSearchParams();
  if (params.q) query.set("q", params.q);
  if (params.offset !== undefined) query.set("offset", String(params.offset));
  if (params.limit !== undefined) query.set("limit", String(params.limit));
  if (params.status) query.set("status", params.status);

  try {
    const res = await apiFetch(`/api/erp-transaction/${config.apiPath}/list?${query.toString()}`);
    if (!res.ok) {
      return { items: [], total: 0, error: statusErrorKey(res.status, "load_failed", "erp_err_read_forbidden") };
    }
    const data = await res.json();
    if (Array.isArray(data?.data)) {
      return { items: data.data, total: data.pagination?.total ?? data.data.length };
    }
    if (Array.isArray(data)) {
      return { items: data, total: data.length };
    }
    return { items: [], total: 0, error: "load_failed" };
  } catch {
    return { items: [], total: 0, error: "connection_error" };
  }
}

// ข้อความปฏิเสธจาก backend แปลตาม Accept-Language ของคำขอ — ส่งภาษาที่ผู้ใช้เลือกในหน้าจอ ไม่ใช่ภาษาของเบราว์เซอร์
const languageHeader = (language: string) => ({ "Accept-Language": language });

export async function saveErpTransaction(
  config: ErpModuleConfig,
  doc: Partial<ErpTransactionDoc>,
  isEdit: boolean,
  language = "th",
): Promise<{ success: boolean; data?: ErpTransactionDoc; message?: string; detail?: string }> {
  try {
    const method = isEdit ? "PUT" : "POST";
    const path = isEdit && doc.id ? `${config.apiPath}/${encodeURIComponent(doc.id)}` : config.apiPath;
    const res = await apiFetch(`/api/erp-transaction/${path}`, {
      method,
      headers: { "Content-Type": "application/json", ...languageHeader(language) },
      body: JSON.stringify(erpDocForSave(doc)),
    });
    if (!res.ok) {
      return { success: false, message: statusErrorKey(res.status, "save_failed", "erp_err_action_forbidden") };
    }
    const data = await res.json();
    // เหตุผลทางบัญชี/ข้อมูล (ปีบัญชีปิด เลขที่ซ้ำ ถูกตัดหนี้แล้ว) backend ตอบ success:false พร้อมข้อความแปลแล้ว — ห้ามนับเป็นสำเร็จ
    if (data?.success === false) return { success: false, message: "save_failed", detail: userDetail(data) };
    return {
      success: true,
      data: data?.data ?? (doc as ErpTransactionDoc),
      message: "save_success",
    };
  } catch {
    return { success: false, message: "connection_error" };
  }
}

export async function deleteErpTransaction(
  config: ErpModuleConfig,
  id: string,
  language = "th",
): Promise<{ success: boolean; message?: string; detail?: string }> {
  try {
    const res = await apiFetch(`/api/erp-transaction/${config.apiPath}/${encodeURIComponent(id)}`, {
      method: "DELETE",
      headers: languageHeader(language),
    });
    if (!res.ok) {
      return { success: false, message: statusErrorKey(res.status, "delete_failed", "erp_err_action_forbidden") };
    }
    const data = await res.json().catch(() => null);
    if (data?.success === false) return { success: false, message: "delete_failed", detail: userDetail(data) };
    return { success: true, message: "delete_success" };
  } catch {
    return { success: false, message: "connection_error" };
  }
}

export interface ErpDocEvent {
  id: number;
  docno: string;
  transflag?: number;
  action: string;
  from_status?: string | number;
  to_status?: string | number;
  reason?: string;
  actor?: string;
  created_at?: string;
  createdby?: string;
  createdat?: string;
}

// ผลของเอกสารแต่ละใบในงานเป็นชุด — id คือเลขที่เอกสารที่ส่งไป, code = คีย์ข้อความ (gl_err_<code>) เมื่อไม่สำเร็จ
export interface BatchApproveItemResult {
  id: string;
  success: boolean;
  code?: string;
  message?: string;
}

export async function approveErpDoc(
  module: string,
  id: string,
  language = "th",
): Promise<{ success: boolean; message?: string; detail?: string }> {
  try {
    const res = await apiFetch(`/api/erp-transaction/${encodeURIComponent(module)}/${encodeURIComponent(id)}/approve`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...languageHeader(language) },
      body: JSON.stringify({}),
    });
    if (!res.ok) {
      return { success: false, message: statusErrorKey(res.status, "action_failed", "erp_err_action_forbidden") };
    }
    const data = await res.json().catch(() => null);
    if (data?.success === false) return { success: false, message: "action_failed", detail: userDetail(data) };
    return { success: true, message: "approve_success" };
  } catch {
    return { success: false, message: "connection_error" };
  }
}

export async function unapproveErpDoc(
  module: string,
  id: string,
  reason = "",
  language = "th",
): Promise<{ success: boolean; message?: string; detail?: string }> {
  try {
    const res = await apiFetch(`/api/erp-transaction/${encodeURIComponent(module)}/${encodeURIComponent(id)}/unapprove`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...languageHeader(language) },
      body: JSON.stringify({ reason }),
    });
    if (!res.ok) {
      return { success: false, message: statusErrorKey(res.status, "action_failed", "erp_err_action_forbidden") };
    }
    const data = await res.json().catch(() => null);
    if (data?.success === false) return { success: false, message: "action_failed", detail: userDetail(data) };
    return { success: true, message: "unapprove_success" };
  } catch {
    return { success: false, message: "connection_error" };
  }
}

export async function cancelErpDoc(
  module: string,
  id: string,
  reason: string,
  language = "th",
): Promise<{ success: boolean; message?: string; detail?: string }> {
  try {
    const res = await apiFetch(`/api/erp-transaction/${encodeURIComponent(module)}/${encodeURIComponent(id)}/cancel`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...languageHeader(language) },
      body: JSON.stringify({ reason }),
    });
    if (!res.ok) {
      return { success: false, message: statusErrorKey(res.status, "action_failed", "erp_err_action_forbidden") };
    }
    const data = await res.json().catch(() => null);
    if (data?.success === false) return { success: false, message: "action_failed", detail: userDetail(data) };
    return { success: true, message: "cancel_success" };
  } catch {
    return { success: false, message: "connection_error" };
  }
}

interface BatchResult {
  success: boolean;
  items?: BatchApproveItemResult[];
  total?: number;
  message?: string;
  detail?: string;
}

// backend ตอบ data เป็น array ผลทีละใบ {id, success, code, message} (BatchApproveResponse) — HTTP 200 ไม่ได้แปลว่าทุกใบสำเร็จ
// จึงนับสำเร็จจาก success ของแต่ละใบ และคืน items ให้หน้าจอแสดงใบที่ไม่สำเร็จ
async function postBatchErpDocs(
  module: string,
  action: "approve" | "unapprove" | "cancel",
  body: { docnos: string[]; reason?: string },
  language: string,
): Promise<BatchResult & { succeeded?: number }> {
  try {
    const res = await apiFetch(`/api/erp-transaction/${encodeURIComponent(module)}/${action}-batch`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...languageHeader(language) },
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      return { success: false, message: statusErrorKey(res.status, "action_failed", "erp_err_action_forbidden") };
    }
    const data = await res.json().catch(() => null);
    if (data?.success === false) return { success: false, message: "action_failed", detail: userDetail(data) };
    const items: BatchApproveItemResult[] = Array.isArray(data?.data) ? data.data : [];
    return {
      success: true,
      items,
      total: items.length || body.docnos.length,
      succeeded: items.filter((item) => item.success).length,
      message: `${action}_success`,
    };
  } catch {
    return { success: false, message: "connection_error" };
  }
}

export async function approveBatchErpDocs(
  module: string,
  docnos: string[],
  language = "th",
): Promise<BatchResult & { approved?: number }> {
  const { succeeded, ...result } = await postBatchErpDocs(module, "approve", { docnos }, language);
  return { ...result, approved: succeeded };
}

export async function unapproveBatchErpDocs(
  module: string,
  docnos: string[],
  reason?: string,
  language = "th",
): Promise<BatchResult & { unapproved?: number }> {
  const { succeeded, ...result } = await postBatchErpDocs(module, "unapprove", { docnos, reason }, language);
  return { ...result, unapproved: succeeded };
}

export async function cancelBatchErpDocs(
  module: string,
  docnos: string[],
  reason: string,
  language = "th",
): Promise<BatchResult & { cancelled?: number }> {
  const { succeeded, ...result } = await postBatchErpDocs(module, "cancel", { docnos, reason }, language);
  return { ...result, cancelled: succeeded };
}

export async function fetchErpDocEvents(
  module: string,
  id: string,
  language = "th",
): Promise<{ success: boolean; events: ErpDocEvent[]; error?: string }> {
  try {
    const res = await apiFetch(`/api/erp-transaction/${encodeURIComponent(module)}/${encodeURIComponent(id)}/events`, {
      headers: languageHeader(language),
    });
    if (!res.ok) {
      return { success: false, events: [], error: statusErrorKey(res.status, "load_failed", "erp_err_read_forbidden") };
    }
    const data = await res.json().catch(() => null);
    if (data?.success === false) return { success: false, events: [], error: "load_failed" };
    const list = Array.isArray(data?.data) ? data.data : Array.isArray(data?.events) ? data.events : (data?.data?.events ?? []);
    return { success: true, events: list };
  } catch {
    return { success: false, events: [], error: "connection_error" };
  }
}

export type DeliveryDateItem = {
  docno: string;
  docdate: string;
  custcode: string;
  custname: string;
  credit_of_bill: number;
  old_duedate: string;
  delivery_date: string;
  credit_delivery: number;
  new_duedate: string;
};

export async function fetchDeliveryDateInvoices(
  module: string,
  partner = "",
  language = "th",
): Promise<{ success: boolean; items: DeliveryDateItem[]; error?: string }> {
  try {
    const query = partner ? `?partner=${encodeURIComponent(partner)}` : "";
    const res = await apiFetch(`/api/erp-transaction/${encodeURIComponent(module)}/delivery-dates${query}`, {
      headers: languageHeader(language),
    });
    if (!res.ok) {
      return { success: false, items: [], error: statusErrorKey(res.status, "load_failed", "erp_err_read_forbidden") };
    }
    const data = await res.json().catch(() => null);
    if (data?.success === false) return { success: false, items: [], error: "load_failed" };
    const list = Array.isArray(data?.data) ? data.data : [];
    return { success: true, items: list };
  } catch {
    return { success: false, items: [], error: "connection_error" };
  }
}

export async function saveDeliveryDates(
  module: string,
  items: Array<{
    docno: string;
    delivery_date: string;
    credit_delivery: number;
    new_duedate: string;
  }>,
  language = "th",
): Promise<{ success: boolean; count?: number; message?: string; detail?: string }> {
  try {
    const res = await apiFetch(`/api/erp-transaction/${encodeURIComponent(module)}/delivery-dates`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...languageHeader(language) },
      body: JSON.stringify({ items }),
    });
    if (!res.ok) {
      return { success: false, message: statusErrorKey(res.status, "action_failed", "erp_err_action_forbidden") };
    }
    const data = await res.json().catch(() => null);
    if (data?.success === false) return { success: false, message: "action_failed", detail: userDetail(data) };
    return { success: true, count: data?.count ?? items.length, message: "ops_save_delivery_dates_success" };
  } catch {
    return { success: false, message: "connection_error" };
  }
}
