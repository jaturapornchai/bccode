import type { LanguageCode } from "./i18n";
import { isBackendLanguageReady, type BackendLanguageDictionary } from "./backend-language";

export type MenuCategory = "transaction" | "report" | "master" | "settings" | "finance" | "restaurant" | "approval";

export type MenuLabel = {
  key?: string;
  /** Familiar names from the previous program, searched without duplicating menu entries. */
  aliases?: string[];
  th: string;
  en: string;
} & Partial<Record<LanguageCode, string>>;

export type MenuItem = {
  id: string;
  label: MenuLabel;
  route: string;
  category: MenuCategory;
};

export type MenuGroup = {
  id: string;
  title: MenuLabel;
  items: MenuItem[];
};

export type MenuSection = {
  id: string;
  title: MenuLabel;
  groups: MenuGroup[];
};

const menuKey = (id: string) => id.replace(/-/g, "_");
const ml = (id: string, th: string, en: string): MenuLabel => ({ key: menuKey(id), th, en });

// Verified against Champ menuconfig.xml; preserve approved visible labels and route IDs.
const CHAMP_SEARCH_ALIASES: Record<string, string[]> = {
  "account-transfer": ["โอนเงินระหว่างบัญชี"],
  "accrual-receive": ["ตั้งหนี้จากทยอยรับ"],
  "advance-payment": ["จ่ายเงินล่วงหน้า"],
  "advance-payment-refund": ["รับคืนเงินล่วงหน้า"],
  "ap-bad-debt": ["ตัดหนี้สูญเจ้าหนี้"],
  "ap-bill-recalculate": ["คำนวณยอดคงเหลือบิลเจ้าหนี้ใหม่"],
  "ap-billing-receipt": ["ใบรับวางบิลเจ้าหนี้"],
  "ap-other-debt": ["ตั้งเจ้าหนี้อื่นๆ"],
  "ar-bad-debt": ["ตัดหนี้สูญลูกหนี้"],
  "ar-bill-recalculate": ["คำนวณยอดคงเหลือบิลลูกหนี้ใหม่"],
  "ar-other-debt": ["ตั้งลูกหนี้อื่นๆ"],
  "asset-construction-in-progress": ["งานระหว่างก่อสร้าง"],
  "asset-depreciation": ["คำนวณค่าเสื่อมราคา"],
  "asset-maintenance": ["บันทึกซ่อมบำรุงสินทรัพย์"],
  "asset-post-gl": ["โอนค่าเสื่อมราคาเข้าบัญชีแยกประเภท", "Post GL"],
  "asset-registry": ["ทะเบียนสินทรัพย์ถาวร"],
  "asset-type": ["ประเภทสินทรัพย์"],
  "bank-contact": ["ผู้ติดต่อธนาคาร"],
  "bank-recalculate": ["คำนวณยอดสมุดบัญชีใหม่"],
  "billing-note": ["ใบวางบิล"],
  "book-bank": ["สมุดบัญชี"],
  "cash-deposit": ["นำฝากเงินสด"],
  "cash-withdraw": ["ถอนเงินสด"],
  "chart-of-accounts": ["ผังบัญชี"],
  "cheque-deposit": ["นำฝากเช็ครับ"],
  "cheque-discount": ["ขายลดเช็ครับ"],
  "cheque-issued": ["ทะเบียนเช็คจ่าย"],
  "cheque-issued-cancel": ["ยกเลิกเช็คจ่าย"],
  "cheque-issued-clear": ["เช็คจ่ายผ่าน"],
  "cheque-recalculate": ["คำนวณยอดคงเหลือเช็คใหม่"],
  "cheque-received": ["ทะเบียนเช็ครับ"],
  "cheque-received-cancel": ["ยกเลิกเช็ครับ"],
  "cheque-received-clear": ["เช็ครับผ่าน"],
  "cheque-received-return": ["เช็ครับคืน"],
  "cheque-redeposit": ["นำเช็คเข้าใหม่", "บันทึกนำเช็คฝากใหม่"],
  "credit-card-cancel": ["ยกเลิกรายการรับบัตรเครดิต"],
  "credit-card-receipts": ["ทะเบียนรับชำระด้วยบัตรเครดิต"],
  "credit-card-settle": ["ขึ้นเงินบัตรเครดิต"],
  "creditor": ["เจ้าหนี้"],
  "creditor-beginning-balance": ["เจ้าหนี้ตั้งต้นรายเอกสาร"],
  "debit-note": ["ใบเพิ่มหนี้"],
  "debtor": ["ลูกหนี้"],
  "debtor-beginning-balance": ["ลูกหนี้ตั้งต้นรายเอกสาร"],
  "deposit": ["จ่ายเงินมัดจำ"],
  "employee-advance": ["เงินทดรองจ่ายพนักงาน"],
  "expense-record": ["บันทึกค่าใช้จ่าย"],
  "expense-summary-report": ["รายงานสรุปรายจ่าย"],
  "financial-close": ["ปิดงบบัญชีสิ้นงวด"],
  "financial-statement-designer": ["ออกแบบงบการเงิน"],
  "general-ledger": ["บัญชีแยกประเภท"],
  "gl-account-mapping": ["รูปแบบการเชื่อมโยงบัญชีอัตโนมัติ", "Account Mapping"],
  "gl-budget": ["กำหนดงบประมาณประจำปี", "Budgeting"],
  "gl-opening-balance": ["บันทึกข้อมูลรายวันยกมา", "ยอดยกมาทางบัญชี"],
  "gl-post": ["ผ่านรายการบัญชี", "ยกเลิกการผ่านรายการบัญชี", "ยกเลิกเอกสารที่ผ่านรายการแล้ว", "unpost", "posting", "กลับรายการ", "ผ่านรายการและยกเลิก"],
  "gl-product-account-groups": ["กลุ่มบัญชีสินค้า"],
  "gl-journals": ["สมุดรายวันขาย", "สมุดรายวันซื้อ", "สมุดรายวันรับเงิน", "สมุดรายวันจ่ายเงิน", "สมุดรายวันทั่วไป", "UV", "SV", "RV", "PV", "JV", "บันทึกสมุดรายวัน", "รายวัน"],
  "low-stock-alert": ["สินค้าใกล้หมดขั้นต่ำ"],
  "other-income-receipt": ["ใบเสร็จรับเงินรายได้อื่น"],
  "paid": ["รับชำระ"],
  "pay": ["จ่ายชำระ"],
  "payment-daily": ["รายงานรับชำระรายวัน"],
  "period-lock": ["ล็อกงวดบัญชี"],
  "product": ["สินค้า"],
  "product-movement": ["ความเคลื่อนไหวสินค้า"],
  "product-promotion": ["โปรโมชั่น"],
  "product-serial-registry": ["ทะเบียนเลขเครื่อง", "เลขซีเรียล"],
  "product-set": ["สินค้าชุด"],
  "product-set-assemble": ["ตรวจสอบรวมสินค้าชุด"],
  "product-set-components": ["รายการย่อยสินค้าชุดแบบที่ 2"],
  "product-set-disassemble": ["ตรวจสอบแยกสินค้าชุด"],
  "purchase": ["ซื้อสินค้า"],
  "purchase-credit-note": ["ใบเพิ่มหนี้เจ้าหนี้"],
  "purchase-debit-note": ["ใบลดหนี้เจ้าหนี้"],
  "purchase-landed-cost": ["บันทึกต้นทุนแฝง"],
  "purchase-order": ["ใบสั่งซื้อ"],
  "purchase-order-generate": ["ประมวลผลใบสั่งซื้ออัตโนมัติ"],
  "purchase-partial": ["รับสินค้าแบบทยอยรับ"],
  "purchase-price-comparison": ["ตารางเปรียบเทียบราคาซื้อ"],
  "purchase-requisition": ["ใบขอซื้อ"],
  "purchase-requisition-approve": ["อนุมัติใบเสนอซื้อสินค้า"],
  "purchase-return": ["คืนซื้อ"],
  "quotation": ["ใบเสนอราคา"],
  "receive-advance": ["รับเงินล่วงหน้า"],
  "receive-deposit": ["รับเงินมัดจำ"],
  "reprocess": ["คำนวณข้อมูลใหม่"],
  "return-advance": ["คืนเงินล่วงหน้า"],
  "rfq": ["สืบราคาและเจรจา", "ระบบสืบราคาซื้อสินค้า"],
  "sale": ["ขายสินค้า"],
  "sale-order": ["ใบสั่งขาย", "ใบสั่งจอง"],
  "sale-order-date-check": ["ตรวจสอบวันที่ใบสั่งขาย/สั่งจอง"],
  "sale-reservation-flow": ["ติดตามใบสั่งจองสินค้า"],
  "sale-return": ["คืนขาย"],
  "sales-by-seller": ["ยอดขายตามพนักงาน"],
  "sales-daily": ["รายงานขายรายวัน"],
  "stock-adjust": ["ปรับปรุงสต็อก"],
  "stock-balance": ["ยอดยกมาสินค้า"],
  "stock-balance-item": ["คงเหลือตามสินค้า"],
  "stock-balance-warehouse": ["คงเหลือตามคลัง"],
  "stock-count": ["ตรวจนับสต็อก"],
  "stock-count-sheet": ["เอกสารเพื่อตรวจนับสินค้า"],
  "stock-daily-sequence": ["กำหนดลำดับรายวันสินค้า"],
  "stock-issue-request": ["ใบขอเบิกสินค้าและวัตถุดิบ"],
  "stock-movement-cost": ["เคลื่อนไหวสินค้าพร้อมต้นทุน", "รายงานเคลื่อนไหวสินค้าต้นทุนมาตรฐาน"],
  "stock-pickup": ["เบิกสินค้า"],
  "stock-receive": ["รับสินค้า"],
  "stock-return": ["คืนสินค้าเข้าคลัง"],
  "stock-transfer": ["โอนสินค้า"],
  "stock-transfer-request": ["ใบขอโอนสินค้า"],
  "temporary-receipt": ["ใบเสร็จชั่วคราว"],
  "trial-balance": ["งบทดลอง"],
  "withholding-tax-deduction": ["พิมพ์ภาษีหัก ณ ที่จ่าย (ภ.ง.ด. 3,ภ.ง.ด. 53)"],
  "withholding-tax-received": ["ทะเบียนถูกหัก ณ ที่จ่าย"],
  "withholding-tax-report": ["รายงานภาษีหัก ณ ที่จ่าย (ภ.ง.ด. 3)", "รายงานภาษีหัก ณ ที่จ่าย (ภ.ง.ด. 53)"],
  "working-paper": ["กระดาษทำการ"],
};

const tx = (
  id: string,
  th: string,
  en: string,
  route: string,
  category: MenuCategory = "transaction",
  languageKey: string = menuKey(id),
): MenuItem => ({
  id,
  label: { key: languageKey, th, en, aliases: CHAMP_SEARCH_ALIASES[id] },
  route,
  category,
});

// Ordered per module like Champ menuconfig.xml (2026-09-19 Champ-parity pass).
export const MENU_SECTIONS: MenuSection[] = [
  {
    id: "po",
    title: { key: "menu_purchase_order_po", th: "ซื้อ/สั่งซื้อสินค้า", en: "Purchase Order (PO)" },
    groups: [
      {
        id: "po-transactions",
        title: { key: "menu_purchase_expense_transactions", th: "งานซื้อสินค้า", en: "Purchase Transactions" },
        items: [
          tx("purchase-order", "บันทึกใบสั่งซื้อสินค้า", "Purchase Order", "/transaction/purchaseorder"),
          tx("purchase-order-cancel", "ยกเลิกใบสั่งซื้อสินค้า", "Cancel Purchase Order", "/procurement/order-cancellation", "approval"),
          tx("purchase-partial", "บันทึกรับสินค้าจากการซื้อ", "Gradual Receipt", "/transaction/purchasepartial"),
          tx("accrual-receive", "บันทึกตั้งหนี้จากการซื้อ", "Set Debt from Receipt", "/transaction/accrualreceive"),
          tx("purchase-return", "บันทึกส่งคืนสินค้า", "Purchase Return", "/transaction/purchasereturn"),
          tx("purchase-reduce-debt", "บันทึกลดหนี้", "Purchase Debt Reduction", "/transaction/purchasereducedebt"),
          tx("purchase", "บันทึกซื้อสินค้า,บริการ", "Purchase", "/transaction/purchase"),
          tx("purchase-debit-note", "บันทึกส่งคืนสินค้า/ลดหนี้", "Purchase Debit Note", "/transaction/purchasedebitnote"),
          tx("purchase-credit-note", "บันทึกเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้)", "Purchase Credit Note", "/transaction/purchasecreditnote"),
        ],
      },
      {
        id: "po-procurement",
        title: { key: "menu_procurement_operations", th: "ใบเสนอซื้อและสืบราคา", en: "Requisitions & Price Inquiry" },
        items: [
          tx("purchase-requisition", "บันทึกใบเสนอซื้อสินค้า", "Purchase Requisition", "/transaction/purchaserequisition"),
          tx("purchase-requisition-approve", "บันทึกอนุมัติใบเสนอซื้อสินค้า", "Approve Purchase Requisition", "/procurement/requisition-approval", "approval"),
          tx("rfq", "บันทึกใบสืบราคาสินค้ารวม", "Price Inquiry & Negotiation", "/transaction/rfq"),
          tx("rfq-price-table", "บันทึกใบตารางราคารวม", "Consolidated Price Table", "/transaction/rfqpricetable"),
          tx("purchase-order-generate", "ประมวลผลใบสั่งซื้อสินค้าอัตโนมัติ", "Generate Purchase Orders", "/procurement/generate-orders", "transaction"),
        ],
      },
      {
        id: "po-payment",
        title: { key: "menu_advances_deposits", th: "เงินมัดจำและจ่ายล่วงหน้า", en: "Advances & Deposits" },
        items: [
          tx("deposit", "บันทึกใบจ่ายเงินมัดจำ", "Pay Deposit", "/transaction/deposit", "finance"),
          tx("advance-payment", "บันทึกใบจ่ายเงินล่วงหน้า", "Advance Payment", "/transaction/advancepayment", "finance"),
          tx("advance-payment-refund", "บันทึกรับคืนเงินจ่ายล่วงหน้า", "Advance Refund", "/transaction/advancepaymentrefund", "finance"),
        ],
      },
      {
        id: "po-costs",
        title: ml("po-costs", "ต้นทุนแฝงและปรับปรุงใบรับสินค้า", "Landed Costs & Receipt Adjustments"),
        items: [
          tx("purchase-landed-cost", "บันทึกต้นทุนแฝง (Weight cost)", "Landed Cost Allocation", "/transaction/landedcost", "transaction"),
          tx("purchase-receipt-adjustment", "ปรับปรุงใบรับสินค้า", "Adjust Purchase Receipt", "/transaction/purchasereceiptadjustment", "transaction"),
        ],
      },
      {
        id: "po-reports",
        title: { key: "menu_purchase_reports", th: "รายงานจัดซื้อ", en: "Purchase Reports" },
        items: [
          tx("purchase-report", "รายงานซื้อ", "Purchase Report", "/report/reportdedebipurchase", "report"),
          tx("purchase-by-product", "รายงานซื้อตามสินค้า", "Purchase by Product", "/report/purchasebyproduct", "report"),
          tx("gradual-receipt", "รายงานทยอยรับ", "Gradual Receipt Report", "/report/reportdedebipurchasepartial", "report"),
          tx("purchase-price-comparison", "รายงานเปรียบเทียบราคาซื้อ", "Purchase Price Comparison", "/procurement/price-comparison", "report"),
        ],
      },
    ],
  },
  {
    id: "bill",
    title: { key: "menu_order_entry_billing", th: "ใบสั่งของ/ใบกำกับสินค้า", en: "Order Entry & Billing (BILL/OE)" },
    groups: [
      {
        id: "bill-orders",
        title: ml("bill-orders", "ใบเสนอราคาและใบสั่งขาย", "Quotations & Sales Orders"),
        items: [
          tx("quotation", "บันทึกใบเสนอราคาสินค้า", "Quotation", "/transaction/quotation"),
          tx("quotation-approve", "อนุมัติใบเสนอราคาสินค้า", "Approve Quotation", "/sales/quotation-approval", "approval"),
          tx("quotation-cancel", "ยกเลิกใบเสนอราคาสินค้า", "Cancel Quotation", "/sales/quotation-cancellation", "approval"),
          tx("sale-order", "บันทึกใบสั่งขาย/สั่งจองสินค้า", "Sale Order", "/transaction/saleorder"),
          tx("sale-order-approve", "อนุมัติใบสั่งขาย/สั่งจองสินค้า", "Approve Sales Order / Reservation", "/sales/order-approval", "approval"),
          tx("sale-order-cancel", "ยกเลิกใบสั่งขาย/สั่งจองสินค้า", "Cancel Sales Order / Reservation", "/sales/order-cancellation", "approval"),
          tx("sale-reservation-flow", "Flow ใบสั่งจองสินค้า", "Sales Reservation Tracking", "/sales/reservations", "transaction"),
          tx("sale-order-date-check", "ตรวจสอบวันที่ใบสั่งขาย/สั่งจองสินค้า", "Check Sales Order / Reservation Dates", "/sales/order-dates", "transaction"),
        ],
      },
      {
        id: "bill-payment",
        title: { key: "menu_advances_deposits_received", th: "เงินมัดจำและรับล่วงหน้า", en: "Advances & Deposits Received" },
        items: [
          tx("receive-deposit", "บันทึกใบรับเงินมัดจำ", "Receive Deposit", "/transaction/receivedeposit", "finance"),
          tx("receive-advance", "บันทึกใบรับเงินล่วงหน้า", "Receive Advance", "/transaction/paidadvance", "finance"),
          tx("return-advance", "บันทึกคืนเงินรับล่วงหน้า", "Return Advance", "/transaction/paidadvancerefund", "finance"),
        ],
      },
      {
        id: "bill-transactions",
        title: { key: "menu_sales_billing_operations", th: "งานขายสินค้า", en: "Sales Transactions" },
        items: [
          tx("sale", "บันทึกขายสินค้า,บริการ", "Sale", "/transaction/sale"),
          tx("sale-return", "บันทึกใบรับคืนสินค้า,ลดหนี้", "Sale Return", "/transaction/salereturn"),
          tx("debit-note", "บันทึกใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้)", "Debit Note", "/transaction/debitnote"),
          tx("sale-delivery-date", "ปรับปรุงวันที่ส่งของให้ลูกค้า", "Adjust Customer Delivery Date", "/sales/delivery-dates", "transaction"),
          tx("etax-invoice", "ใบกำกับภาษีอิเล็กทรอนิกส์", "e-Tax Invoice & e-Receipt", "/transaction/etax"),
        ],
      },
      {
        id: "bill-reports",
        title: { key: "menu_sales_reports_analytics", th: "รายงานขาย", en: "Sales Reports" },
        items: [
          tx("sales", "รายงานขาย", "Sales Report", "/report/reportdedebisales", "report"),
          tx("sales-daily", "รายงานรายวันขายสินค้าและบริการ", "Daily Sales", "/report/reportdedebisalesdaily", "report"),
          tx("sales-by-seller", "รายงานยอดขายสุทธิ-เรียงตามพนักงานขาย", "Sales by Seller", "/report/salesbyseller", "report", "sales_by_seller"),
          tx("sales-by-document", "รายงานขายตามเอกสาร", "Sales by Document", "/report/salesreportbydocument", "report", "sales_by_document"),
          tx("gross-profit-by-document", "กำไรขั้นต้นตามเอกสาร", "Gross Profit by Document", "/report/reportgrossprofitbydocument", "report"),
          tx("gross-profit-by-product", "กำไรขั้นต้นตามสินค้า", "Gross Profit by Product", "/report/reportgrossprofitbyproduct", "report"),
          tx("sale-return-report", "รายงานคืนขาย", "Sale Return Report", "/report/reportdedebisalereturn", "report"),
        ],
      },
    ],
  },
  {
    id: "ap",
    title: { key: "menu_account_payable_ap", th: "เจ้าหนี้", en: "Account Payable (AP)" },
    groups: [
      {
        id: "ap-master",
        title: { key: "menu_creditor_master_data", th: "ข้อมูลหลักเจ้าหนี้", en: "Creditor Master Data" },
        items: [
          tx("creditor", "รายละเอียดเจ้าหนี้", "Creditor", "/creditor", "master"),
          tx("creditor-beginning-balance", "บันทึกเอกสารเจ้าหนี้ยกมา", "Creditor Beginning Balance", "/creditorbeginningbalance", "master"),
        ],
      },
      {
        id: "ap-adjustments",
        title: ml("ap-adjustments", "งานเจ้าหนี้", "Payables Transactions"),
        items: [
          tx("ap-other-debt", "บันทึกเอกสารตั้งเจ้าหนี้อื่นๆ", "Other Payables", "/transaction/apotherdebt", "finance"),
          tx("ap-billing-receipt", "บันทึกใบรับวางบิล", "Supplier Billing Receipt", "/transaction/apbillingreceipt", "finance"),
          tx("pay", "บันทึกจ่ายชำระหนี้", "Pay", "/transaction/pay", "finance"),
          tx("ap-bad-debt", "บันทึกตัดหนี้สูญ(เจ้าหนี้)", "Write Off Payables", "/transaction/apbaddebt", "finance"),
        ],
      },
      {
        id: "ap-reports",
        title: { key: "menu_ap_reports", th: "รายงานเจ้าหนี้", en: "AP Reports" },
        items: [
          tx("ap-aging", "รายงานอายุเจ้าหนี้", "AP Aging Report", "/report/apaging", "report"),
          tx("ap-movement", "รายงานเคลื่อนไหวเจ้าหนี้", "Creditor Movement Report", "/report/apmovement", "report"),
          tx("ap-status", "รายงานสถานะเจ้าหนี้", "Creditor Status Report", "/report/apstatus", "report"),
          tx("ap-outstanding", "รายงานใบส่งของค้างจ่ายชำระ", "Unpaid Invoices Report", "/report/apoutstanding", "report"),
          tx("ap-daily-payment", "รายงานการจ่ายเงินประจำวัน", "Daily Payment Report", "/report/apdailypayment", "report"),
        ],
      },
      {
        id: "ap-utilities",
        title: ml("ap-utilities", "ตรวจและประมวลผลเจ้าหนี้", "Payables Utilities"),
        items: [
          tx("ap-recalculate", "คำนวณยอดเจ้าหนี้ใหม่", "Recalculate Payables", "/tools/ap-recalculate", "finance"),
          tx("ap-bill-recalculate", "คำนวณยอดคงเหลือของบิลใหม่", "Recalculate Payable Bill Balances", "/tools/ap-bill-balances", "finance"),
        ],
      },
    ],
  },
  {
    id: "ar",
    title: { key: "menu_account_receivable_ar", th: "ลูกหนี้", en: "Account Receivable (AR)" },
    groups: [
      {
        id: "ar-master",
        title: { key: "menu_debtor_master_data", th: "ข้อมูลหลักลูกหนี้", en: "Debtor Master Data" },
        items: [
          tx("debtor", "รายละเอียดลูกหนี้", "Debtor", "/debtor", "master"),
          tx("debtor-beginning-balance", "บันทึกเอกสารลูกหนี้ยกมา", "Debtor Beginning Balance", "/debtorbeginningbalance", "master"),
        ],
      },
      {
        id: "ar-transactions",
        title: { key: "menu_billing_invoicing", th: "งานลูกหนี้", en: "Receivables Transactions" },
        items: [
          tx("ar-other-debt", "บันทึกเอกสารตั้งลูกหนี้อื่นๆ", "Other Receivables", "/transaction/arotherdebt", "finance"),
          tx("billing-note", "บันทึกใบวางบิล", "Billing Note", "/transaction/billingnote"),
          tx("temporary-receipt", "บันทึกใบเสร็จชั่วคราว", "Temporary Receipt", "/transaction/temporaryreceipt", "finance"),
          tx("paid", "บันทึกใบเสร็จรับเงิน/รับชำระหนี้", "Receive Payment", "/transaction/paid", "finance"),
          tx("ar-bad-debt", "บันทึกตัดหนี้สูญ(ลูกหนี้)", "Write Off Receivables", "/transaction/arbaddebt", "finance"),
        ],
      },
      {
        id: "ar-reports",
        title: { key: "menu_ar_reports", th: "รายงานลูกหนี้", en: "AR Reports" },
        items: [
          tx("ar-aging", "รายงานอายุลูกหนี้", "AR Aging Report", "/report/araging", "report"),
          tx("payment-daily", "รายงานการรับชำระหนี้ประจำวัน", "Daily Payment", "/report/reportdedebipaymentdaily", "report"),
          tx("ar-movement", "รายงานเคลื่อนไหวลูกหนี้", "Debtor Movement Report", "/report/armovement", "report"),
          tx("ar-status", "รายงานสถานะลูกหนี้", "Debtor Status Report", "/report/arstatus", "report"),
          tx("ar-outstanding", "รายงานใบส่งของค้างชำระ", "Outstanding Invoices Report", "/report/aroutstanding", "report"),
          tx("ar-credit-limit", "รายงานตรวจสอบยอดวงเงิน", "Credit Limit Check Report", "/report/arcreditlimit", "report"),
        ],
      },
      {
        id: "ar-utilities",
        title: ml("ar-utilities", "ตรวจและประมวลผลลูกหนี้", "Receivables Utilities"),
        items: [
          tx("ar-recalculate", "คำนวณยอดลูกหนี้ใหม่", "Recalculate Receivables", "/tools/ar-recalculate", "finance"),
          tx("ar-bill-recalculate", "คำนวณยอดคงเหลือของบิลใหม่", "Recalculate Receivable Bill Balances", "/tools/ar-bill-balances", "finance"),
        ],
      },
    ],
  },
  {
    id: "cash-bank",
    title: { key: "menu_cash_cheque", th: "เงินสดและธนาคาร", en: "Cash & Cheque" },
    groups: [
      {
        id: "cheques-received",
        title: ml("cheques-received", "เช็ครับ", "Received Cheques"),
        items: [
          tx("cheque-received", "บันทึกเช็ครับ", "Cheques Received", "/transaction/chequereceived", "finance"),
          tx("cheque-deposit", "บันทึกนำฝากเช็ครับ", "Deposit Received Cheques", "/banking/cheques/deposit", "finance"),
          tx("cheque-received-clear", "บันทึกเช็ครับผ่าน", "Clear Received Cheques", "/banking/cheques/received-clear", "finance"),
          tx("cheque-received-return", "บันทึกเช็ครับคืน", "Returned Received Cheques", "/banking/cheques/received-return", "finance"),
          tx("cheque-redeposit", "บันทึกนำเช็คเข้าใหม่", "Redeposit Cheques", "/banking/cheques/redeposit", "finance"),
          tx("cheque-received-cancel", "บันทึกยกเลิกเช็ครับ", "Cancel Received Cheques", "/banking/cheques/received-cancel", "finance"),
          tx("cheque-discount", "บันทึกขายลดเช็ครับ", "Discount Received Cheques", "/banking/cheques/discount", "finance"),
        ],
      },
      {
        id: "cheques-issued",
        title: ml("cheques-issued", "เช็คจ่าย", "Issued Cheques"),
        items: [
          tx("cheque-issued", "บันทึกเช็คจ่าย", "Cheques Issued", "/transaction/chequeissued", "finance"),
          tx("cheque-issued-clear", "บันทึกเช็คจ่ายผ่าน", "Clear Issued Cheques", "/banking/cheques/issued-clear", "finance"),
          tx("cheque-issued-cancel", "บันทึกยกเลิกเช็คจ่าย", "Cancel Issued Cheques", "/banking/cheques/issued-cancel", "finance"),
        ],
      },
      {
        id: "card-settlement",
        title: ml("card-settlement", "รับและขึ้นเงินบัตรเครดิต", "Credit Card Receipts & Settlement"),
        items: [
          tx("credit-card-receipts", "บันทึกบัตรเครดิต", "Credit Card Receipts Register", "/banking/cards/receipts", "finance"),
          tx("credit-card-settle", "บันทึกขึ้นเงินบัตรเครดิต", "Settle Credit Card Receipts", "/banking/cards/settlement", "finance"),
          tx("credit-card-cancel", "บันทึกยกเลิกบัตรเครดิต", "Cancel Credit Card Receipts", "/banking/cards/cancellation", "finance"),
        ],
      },
      {
        id: "cash-management",
        title: { key: "menu_cash_advances", th: "เงินสดและเงินทดรองจ่าย", en: "Cash & Advances" },
        items: [
          tx("other-income-receipt", "บันทึกรับเงินอื่นๆ", "Other Income Receipt", "/transaction/otherincomereceipt", "finance"),
          tx("employee-advance", "บันทึกขอเบิกเงินทดรองจ่าย", "Employee Advance", "/transaction/employeeadvance", "finance"),
          tx("expense-record", "บันทึกจ่ายเงินอื่นๆ", "Expense Record", "/transaction/expense"),
          tx("petty-cash-limit", "กำหนดวงเงินสดย่อย", "Petty Cash Limit", "/banking/petty-cash-limit", "finance"),
        ],
      },
      {
        id: "cash-bank-master",
        title: { key: "menu_bank_accounts", th: "สมุดเงินฝากธนาคาร", en: "Bank Accounts" },
        items: [
          tx("bank-contact", "บันทึกรายละเอียดผู้ติดต่อธนาคาร", "Bank Contacts", "/banking/contacts", "master"),
          tx("book-bank", "บันทึกสมุดเงินฝากธนาคาร", "Bank Book", "/bookbankscreen", "master", "book_bank"),
        ],
      },
      {
        id: "bank-transactions",
        title: { key: "menu_banking_transactions", th: "ธุรกรรมธนาคาร", en: "Banking Transactions" },
        items: [
          tx("cash-deposit", "บันทึกนำฝากเงินสด", "Deposit Cash", "/banking/cash-deposit", "finance"),
          tx("cash-withdraw", "บันทึกถอนเงินสด", "Withdraw Cash", "/banking/cash-withdrawal", "finance"),
          tx("bank-charges", "บันทึกค่าใช้จ่ายธนาคาร", "Bank Charges", "/banking/charges", "finance"),
          tx("bank-income", "บันทึกรายได้จากธนาคาร", "Bank Income", "/banking/income", "finance"),
          tx("account-transfer", "บันทึกโอนเงินระหว่างธนาคาร", "Account Transfer", "/transaction/accounttransfer", "finance"),
        ],
      },
      {
        id: "bank-utilities",
        title: ml("bank-utilities", "ตรวจและประมวลผลเงินฝากและเช็ค", "Bank & Cheque Utilities"),
        items: [
          tx("cheque-recalculate", "คำนวณยอดคงเหลือของเช็คใหม่", "Recalculate Cheque Balances", "/tools/cheque-balances", "finance"),
          tx("bank-recalculate", "คำนวณยอดคงเหลือของสมุดบัญชีใหม่", "Recalculate Bank Book Balances", "/tools/bank-balances", "finance"),
        ],
      },
      {
        id: "bank-reports",
        title: ml("bank-reports", "รายงานเช็ค เงินสด และธนาคาร", "Cheque, Cash & Bank Reports"),
        items: [
          tx("cheque-received-report", "รายงานเช็ครับ", "Received Cheques Report", "/report/chequereceived", "report"),
          tx("cheque-issued-report", "รายงานเช็คจ่าย", "Issued Cheques Report", "/report/chequeissued", "report"),
          tx("credit-card-report", "รายงานบัตรเครดิต", "Credit Card Report", "/report/creditcard", "report"),
          tx("bank-statement-report", "รายงาน Bank Statement", "Bank Statement Report", "/report/bankstatement", "report"),
          tx("cash-movement-report", "รายงานเคลื่อนไหวเงินสด", "Cash Movement Report", "/report/cashmovement", "report"),
          tx("petty-cash-movement-report", "รายงานเคลื่อนไหวเงินสดย่อย", "Petty Cash Movement Report", "/report/pettycashmovement", "report"),
          tx("expense-summary-report", "รายงานการจ่ายเงินอื่นๆ", "Other Payments Report", "/report/expensesummary", "report"),
          tx("monthly-payment-book-report", "รายงานสมุดจ่ายเงินประจำเดือน", "Monthly Payment Book", "/report/monthlypaymentbook", "report"),
        ],
      },
    ],
  },
  {
    id: "ic",
    title: { key: "menu_inventory_control_ic", th: "สินค้าคงคลัง", en: "Inventory Control (IC)" },
    groups: [
      {
        id: "ic-master",
        title: { key: "menu_product_catalog_setup", th: "ข้อมูลหลักสินค้า", en: "Product Catalog & Setup" },
        items: [
          tx("product", "รายละเอียดสินค้า", "Product", "/product", "master"),
          { ...tx("barcode", "บาร์โค้ด", "Barcode", "/productbarcode", "master"), label: ml("barcode", "บาร์โค้ด", "Barcode") },
          tx("product-unit", "หน่วยนับสินค้า", "Product Unit", "/productunit", "master"),
          tx("product-group", "กลุ่มสินค้า", "Product Group", "/productgroup", "master"),
          tx("brand", "ยี่ห้อสินค้า", "Brand", "/masterbrandscreen", "master"),
          tx("warehouse", "คลัง", "Warehouse", "/productwarehousescreen", "master"),
        ],
      },
      {
        id: "ic-pricing",
        title: ml("ic-pricing", "ราคาขายและโปรโมชั่น", "Selling Prices & Promotions"),
        items: [
          tx("product-sale-price", "กำหนดราคาขายสินค้า", "Product Selling Prices", "/inventory/selling-prices", "master"),
          tx("product-promotion", "กำหนดราคาขาย/โปรโมชั่นสินค้า", "Promotion", "/promotionscreen", "master"),
          tx("product-price-adjust", "ปรับปรุงราคาขายสินค้า", "Adjust Selling Prices", "/inventory/price-adjustment", "master"),
        ],
      },
      {
        id: "ic-transactions",
        title: { key: "menu_stock_operations", th: "งานประจำสินค้าและคลัง", en: "Stock Operations" },
        items: [
          tx("stock-balance", "บันทึกสินค้ายกมา", "Stock Balance", "/transaction/stockbalance"),
          tx("stock-issue-request", "บันทึกใบขอเบิกใช้สินค้า,วัตถุดิบ", "Stock / Material Issue Request", "/inventory/issue-request", "transaction"),
          tx("stock-pickup", "บันทึกเบิกใช้สินค้า,วัตถุดิบ", "Stock Pickup", "/transaction/stockpickupproduct"),
          tx("stock-return", "บันทึกรับคืนสินค้า,วัตถุดิบ", "Stock Return", "/transaction/stockreturnproduct"),
          tx("stock-receive", "บันทึกรับสินค้าสำเร็จรูป", "Stock Receive", "/transaction/stockreceiveproduct"),
          tx("stock-transfer-request", "บันทึกใบขอโอนสินค้า", "Stock Transfer Request", "/inventory/transfer-request", "transaction"),
          tx("stock-transfer", "บันทึกโอนสินค้าระหว่างคลัง", "Stock Transfer", "/transaction/stocktransfer"),
          tx("goods-inspection", "ใบตรวจรับสินค้า", "Goods Inspection Note", "/inventory/goods-inspection", "transaction"),
        ],
      },
      {
        id: "ic-counting",
        title: ml("ic-counting", "ตรวจนับและปรับปรุงสินค้า", "Stock Counting & Adjustments"),
        items: [
          tx("stock-count-sheet", "บันทึกเอกสารเพื่อตรวจนับสินค้า", "Stock Count Sheet", "/inventory/count-sheet", "transaction"),
          tx("stock-count", "บันทึกผลการตรวจนับสินค้า", "Stock Count", "/transaction/stockcount"),
          tx("stock-adjust", "บันทึกปรับปรุงสินค้าหลังตรวจนับ", "Stock Adjustment", "/transaction/adjust"),
        ],
      },
      {
        id: "ic-sets",
        title: ml("ic-sets", "สินค้าชุดและส่วนประกอบ", "Product Sets & Components"),
        items: [
          tx("product-set", "บันทึกสินค้าชุด", "Product Sets", "/inventory/product-sets", "master"),
          tx("product-set-assemble", "ตรวจสอบสินค้าที่สามารถรวมเป็นสินค้าชุดได้", "Check Set Assembly", "/inventory/set-assembly", "transaction"),
          tx("product-set-disassemble", "ตรวจสอบชุดสินค้าที่สามารถแยกเป็นสินค้าได้", "Check Set Disassembly", "/inventory/set-disassembly", "transaction"),
          tx("product-set-components", "บันทึกรายการย่อยสินค้าชุดแบบที่ 2", "Product Set Components Type 2", "/inventory/set-components", "master"),
        ],
      },
      {
        id: "ic-identifiers",
        title: ml("ic-identifiers", "ทะเบียนเลขเครื่องสินค้า", "Product Serial Numbers"),
        items: [
          tx("product-serial-registry", "บันทึก Serial Number", "Serial Registry", "/productserialregistry", "master"),
        ],
      },
      {
        id: "ic-reports",
        title: { key: "menu_inventory_reports", th: "รายงานสินค้า", en: "Inventory Reports" },
        items: [
          tx("stock-balance-item", "รายงานยอดคงเหลือสินค้า-ตามสินค้า", "Stock by Item", "/report/stockbalanceitem", "report"),
          tx("stock-balance-warehouse", "รายงานยอดคงเหลือสินค้า-ตามคลัง,สินค้า", "Stock by Warehouse", "/report/stockbalancewarehouse", "report"),
          tx("inventory", "รายงานสินค้าคงเหลือ", "Inventory Report", "/report/reportdedebistockbalance", "report"),
          tx("low-stock-alert", "รายงานยอดคงเหลือสินค้าที่ถึงจุดสั่งซื้อ", "Low Stock Alert", "/report/lowstock", "report"),
          tx("max-stock-report", "รายงานยอดคงเหลือสินค้าที่ถึงจุดสูงสุด", "Maximum Stock Level Report", "/report/maxstock", "report"),
          tx("product-movement", "รายงานเคลื่อนไหวสินค้า", "Product Movement", "/report/reportstockmovement", "report"),
          tx("stock-movement-cost", "รายงานบัญชีคุมพิเศษสินค้า", "Stock Movement with Cost", "/report/stockmovementcost", "report"),
          tx("no-movement-stock-report", "รายงานสินค้าที่ไม่มีการเคลื่อนไหว", "Non-moving Stock Report", "/report/nomovementstock", "report"),
          tx("stock-count-variance-report", "รายงานผลต่างจากการตรวจนับ", "Stock Count Variance Report", "/report/stockcountvariance", "report"),
          tx("pending-receive-report", "รายงานสรุปยอดสินค้าค้างรับ", "Pending Receipt Summary", "/report/pendingreceive", "report"),
          tx("pending-delivery-report", "รายงานสรุปยอดสินค้าค้างส่ง", "Pending Delivery Summary", "/report/pendingdelivery", "report"),
          tx("serial-movement-report", "รายงานเคลื่อนไหว Serial Number", "Serial Number Movement Report", "/report/serialmovement", "report"),
        ],
      },
      {
        id: "ic-utilities",
        title: ml("ic-utilities", "ตรวจและประมวลผลสินค้า", "Inventory Utilities"),
        items: [
          tx("stock-daily-sequence", "กำหนดลำดับรายวัน", "Inventory Daily Sequence", "/inventory/daily-sequence", "transaction"),
          tx("reprocess", "คำนวณยอดสินค้าใหม่", "Recalculate Data", "/rebuildstockscreen", "report"),
        ],
      },
    ],
  },
  {
    id: "fa",
    title: { key: "menu_fixed_assets_fa", th: "สินทรัพย์และค่าเสื่อมราคา", en: "Fixed Assets (FA)" },
    groups: [
      {
        id: "fa-transactions",
        title: { key: "menu_fixed_asset_operations", th: "งานสินทรัพย์ถาวร", en: "Fixed Asset Operations" },
        items: [
          tx("asset-registry", "รายละเอียดสินทรัพย์", "Fixed Asset Registry", "/asset/registry", "master"),
          tx("asset-maintenance", "บันทึกรายการซ่อมบำรุง", "Asset Maintenance", "/asset/maintenance", "transaction"),
          tx("asset-construction-in-progress", "บันทึกสินทรัพย์ระหว่างก่อสร้าง", "Construction in Progress", "/asset/cip", "master", "asset_construction_in_progress"),
          tx("asset-type", "กำหนดประเภทสินทรัพย์", "Asset Types", "/asset/types", "master"),
          tx("asset-depreciation", "ประมวลผลสินทรัพย์", "Asset Depreciation", "/asset/depreciation", "master"),
        ],
      },
      {
        id: "fa-posting",
        title: ml("fa-posting", "ผ่านค่าเสื่อมเข้าบัญชี", "Depreciation Posting"),
        items: [
          tx("asset-post-gl", "โอนข้อมูลเข้าสู่ GL", "Post Depreciation to General Ledger", "/asset/post-gl", "finance"),
        ],
      },
      {
        id: "fa-reports",
        title: { key: "menu_fixed_asset_reports", th: "รายงานสินทรัพย์", en: "Fixed Asset Reports" },
        items: [
          tx("fixed-asset-schedule", "ตารางค่าเสื่อมและสินทรัพย์", "Fixed Asset Schedule", "/report/assetschedule", "report", "fixed_asset_schedule"),
          tx("depreciation-monthly-report", "รายงานค่าเสื่อมราคาเป็นเดือน", "Monthly Depreciation Report", "/report/depreciationmonthly", "report"),
          tx("depreciation-yearly-report", "รายงานค่าเสื่อมราคาเป็นปี", "Yearly Depreciation Report", "/report/depreciationyearly", "report"),
          tx("depreciation-pnd50-report", "รายงานค่าเสื่อมตาม ภ.ง.ด.50", "Depreciation Report (PND.50)", "/report/depreciationpnd50", "report"),
          tx("asset-disposal-report", "รายงานสินทรัพย์ที่ถูกขายพร้อมกำไรขั้นต้น", "Disposed Assets & Gain Report", "/report/assetdisposal", "report"),
        ],
      },
    ],
  },
  {
    id: "gl",
    title: { key: "menu_general_ledger_gl", th: "บัญชีแยกประเภท", en: "General Ledger (GL)" },
    groups: [
      {
        id: "gl-master",
        title: { key: "menu_master_balances", th: "ข้อมูลหลักและยอดยกมา", en: "Master & Balances" },
        items: [
          tx("chart-of-accounts", "รายละเอียดผังบัญชี", "Chart of Accounts", "/gl/chartofaccounts", "finance"),
          tx("gl-account-groups", "บันทึกกลุ่มผังบัญชี", "Chart of Account Groups", "/gl/account-groups", "finance"),
          tx("gl-opening-balance", "บันทึกยอดสะสมประจำปี", "Annual Accumulated Balances", "/gl/openingbalance", "finance"),
          tx("gl-budget", "กำหนดงบประมาณ", "Annual Budget", "/gl/budget", "finance"),
          tx("gl-account-mapping", "กำหนดรูปแบบการเชื่อม", "Automatic Account Mapping", "/gl/account-mapping", "finance"),
          tx("gl-product-account-groups", "กำหนดกลุ่มบัญชีสินค้า", "Product Account Groups", "/gl/product-account-groups", "finance"),
          tx("gl-fiscal-years", "ปีบัญชีและบัญชีปิดปี", "Fiscal Year and Closing Accounts", "/gl/fiscal-years", "finance", "gl_fiscal_year_closing_acct"),
        ],
      },
      {
        id: "gl-journals",
        title: { key: "menu_journals_operations", th: "สมุดรายวันและงานประจำ", en: "Journals & Operations" },
        items: [
          tx("gl-journals", "สมุดรายวัน", "Journals", "/gl/journals", "finance", "gl_journals"),
          tx("gl-journal-books", "กำหนดสมุดรายวัน", "Journal Books", "/gl/journal-books", "finance"),
          tx("period-lock", "กำหนดงวดบัญชี", "Period Lock", "/gl/periodlock", "finance"),
          tx("financial-close", "ปิดงวดบัญชี", "Financial Period Close", "/gl/financialclose", "finance"),
        ],
      },
      {
        id: "gl-posting",
        title: ml("gl-posting", "ผ่านบัญชีและประมวลผลสิ้นปี", "Ledger Posting & Year-End"),
        items: [
          tx("gl-post", "ผ่านรายการและยกเลิก", "Posting & Reversal", "/gl/posting", "finance", "gl_post_and_reverse"),
          tx("gl-recalculate-posted", "คำนวณยอดผ่านรายการใหม่", "Recalculate Posted Balances", "/gl/recalculate-posted", "finance"),
          tx("gl-reprocess", "ประมวลผลข้อมูลบัญชีใหม่", "Reprocess Accounting Data", "/gl/reprocess", "finance"),
          tx("gl-year-end", "ประมวลผลสิ้นปี", "Year-End Processing", "/gl/year-end", "finance"),
        ],
      },
      {
        id: "gl-reports",
        title: { key: "menu_financial_statements_reports", th: "รายงานการเงินและงบบัญชี", en: "Financial Statements & Reports" },
        items: [
          tx("gl-daily-report", "รายงานข้อมูลรายวัน", "Journal Entries Report", "/report/gljournal", "report"),
          tx("general-ledger", "รายงานบัญชีแยกประเภท", "General Ledger", "/report/ledger", "report", "general_ledger"),
          tx("trial-balance", "รายงานงบทดลอง", "Trial Balance", "/report/trialbalance", "report"),
          tx("working-paper", "รายงานกระดาษทำการ", "Working Paper", "/gl/workingpaper", "report"),
          tx("financial-statement-designer", "รูปแบบงบการเงิน", "Financial Statement Designer", "/gl/statement-designer", "finance"),
          tx("profit-loss", "งบกำไรขาดทุน", "Profit and Loss", "/report/pnl", "report"),
          tx("balance-sheet", "งบดุล", "Balance Sheet", "/report/balancesheet", "report"),
          tx("budget-comparison-report", "รายงานเปรียบเทียบงบประมาณ", "Budget Comparison Report", "/report/budgetcomparison", "report"),
        ],
      },
      // 2026-09-23 ลุงจืด: ทุกอย่างอยู่ในระบบบัญชีแยกประเภท (ผู้ใช้ GL ตัวเดียว เช่น สำนักงานบัญชี ใช้ได้เลย) — ตรงกับ Champ ที่เมนูภาษีอยู่ใต้ GL
      {
        id: "vat-transactions",
        title: { key: "menu_tax_registers", th: "บันทึกภาษี", en: "Tax Entries" },
        items: [
          tx("purchase-vat-adjust", "ปรับปรุงภาษีซื้อ", "Adjust Purchase VAT", "/transaction/purchasevatadjustment", "finance"),
          tx("withholding-tax-deduction", "ภาษีหัก ณ ที่จ่าย", "Withholding Tax", "/transaction/withholdingtax", "finance"),
        ],
      },
      {
        id: "vat-reports",
        title: { key: "menu_tax_reports_returns", th: "รายงานภาษีและแบบยื่น", en: "Tax Reports & Returns" },
        items: [
          tx("vat-buy", "รายงานภาษีซื้อ", "VAT Purchase", "/report/reportvatbuy", "report"),
          tx("vat-sale", "รายงานภาษีขาย", "VAT Sale", "/report/reportvatsale", "report"),
          tx("vat-summary-report", "รายงานสรุปยอดภาษี", "VAT Summary Report", "/report/vatsummary", "report"),
          tx("vat-pp30", "แบบยื่นภาษี ภ.พ.30", "VAT Return (P.P.30)", "/report/vatpp30", "report"),
          tx("vat-pp36", "แบบยื่น ภ.พ.36", "P.P.36 Return", "/report/vatpp36", "report"),
          tx("withholding-tax-report", "รายงานภาษีหัก ณ ที่จ่าย", "Withholding Tax Report", "/report/wht-reports", "report"),
          tx("withholding-tax-received", "รายงานภาษีถูกหัก ณ ที่จ่าย", "Withholding Tax Received", "/report/whtreceived", "report"),
          tx("vat-pnd2", "แบบยื่น ภ.ง.ด.2", "P.N.D.2 Return", "/report/vatpnd2", "report"),
          tx("vat-pnd3", "แบบยื่น ภ.ง.ด.3", "P.N.D.3 Return", "/report/vatpnd3", "report"),
          tx("vat-pnd53", "แบบยื่น ภ.ง.ด.53", "P.N.D.53 Return", "/report/vatpnd53", "report"),
          // 2026-09-23 ลุงจืดสั่ง "เอาให้ครบ" แบบใน mydocs/sample: ภ.ง.ด.2ก/50/51/93/94, ภ.ธ.40 เปิดจากจอเดียวนี้ (ไม่แตกเมนูรายแบบ)
          tx("tax-forms", "แบบยื่นภาษีทั้งหมด", "All Tax Returns", "/report/taxforms", "report"),
          tx("wht-certificate", "หนังสือรับรองหัก ณ ที่จ่าย (50 ทวิ)", "Withholding Tax Certificate", "/report/whtcertificate", "report"),
        ],
      },
    ],
  },
];
export function menuText(label: MenuLabel, language: LanguageCode, dictionary?: BackendLanguageDictionary): string {
  // Thai menu wording is approved in this catalog. A cached backend dictionary
  // from an older release must not rename a business action after loading.
  if (language === "th" && label.th) return label.th;
  const fromBackend = label.key ? dictionary?.[label.key] : "";
  if (fromBackend) return fromBackend;
  if (label.key && dictionary && !isBackendLanguageReady(dictionary)) return label[language] || label.en || label.th || "";
  return label[language] || label.en || label.th || label.key || "";
}

/**
 * Menu full-text search (2026-08-30): "ค้นหาไม่เจอ ต้องแบบ full text search
 * ด้วย ได้ทุกภาษา" — matching used to compare only the CURRENT UI language,
 * so typing English in a Thai UI (or vice versa) found nothing, and a Thai
 * query with a wrong tone mark missed exact labels.
 */
const SEARCH_NOISE = /[\u200B-\u200D\uFEFF\u2060]/g;
const THAI_COMBINING = /[\u0E31\u0E34-\u0E3A\u0E47-\u0E4E]/g;

export function normalizeMenuSearchText(value: string): string {
  return value
    .replace(SEARCH_NOISE, "")
    .replace(THAI_COMBINING, "")
    .toLocaleLowerCase()
    .replace(/\s+/g, " ")
    .trim();
}

/** Every searchable variant of one label: all shipped languages, the key and
    any backend dictionary override — a query in ANY language must match. */
export function menuSearchHaystack(label: MenuLabel, dictionary?: BackendLanguageDictionary): string {
  const parts: string[] = [label.th, label.en, ...(label.aliases ?? [])];
  for (const code of Object.keys(label)) {
    if (code === "th" || code === "en" || code === "key") continue;
    const value = (label as Record<string, unknown>)[code];
    if (typeof value === "string") parts.push(value);
  }
  if (label.key) {
    parts.push(label.key);
    const fromBackend = dictionary?.[label.key];
    if (fromBackend) parts.push(fromBackend);
  }
  return normalizeMenuSearchText(parts.filter(Boolean).join(" "));
}

export function menuSearchMatches(label: MenuLabel, needle: string, dictionary?: BackendLanguageDictionary): boolean {
  const normalizedNeedle = normalizeMenuSearchText(needle);
  return Boolean(normalizedNeedle) && menuSearchHaystack(label, dictionary).includes(normalizedNeedle);
}

export function flattenMenuItems(): MenuItem[] {
  return MENU_SECTIONS.flatMap((section) => section.groups.flatMap((group) => group.items));
}
