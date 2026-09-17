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
  "bank-reconcile": ["กระทบยอดสมุดเงินฝากธนาคาร"],
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
  "director-advance": ["เงินทดรองจ่ายกรรมการ", "บันทึกขอเบิกเงินทดรองจ่าย"],
  "employee-advance": ["เงินทดรองจ่ายพนักงาน"],
  "expense-record": ["บันทึกค่าใช้จ่าย"],
  "financial-close": ["ปิดงบบัญชีสิ้นงวด"],
  "financial-statement-designer": ["ออกแบบงบการเงิน"],
  "general-ledger": ["บัญชีแยกประเภท"],
  "gl-account-groups": ["กลุ่มผังบัญชี"],
  "gl-account-mapping": ["รูปแบบการเชื่อมโยงบัญชีอัตโนมัติ", "Account Mapping"],
  "gl-annual-accumulated": ["ยอดสะสมประจำปี"],
  "gl-budget": ["กำหนดงบประมาณประจำปี", "Budgeting"],
  "gl-opening-balance": ["บันทึกข้อมูลรายวันยกมา"],
  "gl-post": ["ผ่านรายการบัญชี"],
  "gl-product-account-groups": ["กลุ่มบัญชีสินค้า"],
  "gl-unpost": ["ยกเลิกการผ่านรายการบัญชี"],
  "jv-journal": ["บันทึกข้อมูลรายวัน"],
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
  "stock-balance-location": ["คงเหลือตามโซนเก็บสินค้า", "รายงานยอดคงเหลือสินค้า-ตามที่เก็บ"],
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

export const MENU_SECTIONS: MenuSection[] = [
  {
    id: "po",
    title: { key: "menu_purchase_order_po", th: "ซื้อ/สั่งซื้อสินค้า", en: "Purchase Order (PO)" },
    groups: [
      {
        id: "po-procurement",
        title: { key: "menu_procurement_operations", th: "งานจัดซื้อจัดหา", en: "Procurement Operations" },
        items: [
          tx("procurement-dashboard", "ภาพรวมจัดซื้อ", "Purchase Overview", "/procurement/dashboard"),
          tx("purchase-requisition", "บันทึกใบเสนอซื้อสินค้า", "Purchase Requisition", "/transaction/purchaserequisition"),
          tx("rfq", "บันทึกใบสืบราคาสินค้ารวม", "Price Inquiry & Negotiation", "/transaction/rfq"),
          tx("purchase-order", "บันทึกใบสั่งซื้อสินค้า", "Purchase Order", "/transaction/purchaseorder"),
          tx("import-documents", "นำเข้าเอกสารจากไฟล์", "Import Documents from File", "/importdocuments", "master"),
          tx("purchase-price-comparison", "รายงานเปรียบเทียบราคาซื้อ", "Purchase Price Comparison", "/procurement/price-comparison", "transaction"),
          tx("purchase-order-generate", "ประมวลผลใบสั่งซื้อสินค้าอัตโนมัติ", "Generate Purchase Orders", "/procurement/generate-orders", "transaction"),
        ],
      },
      {
        id: "po-approvals",
        title: ml("po-approvals", "อนุมัติและยกเลิกการซื้อ", "Purchase Approvals & Cancellations"),
        items: [
          tx("purchase-requisition-approve", "บันทึกอนุมัติใบเสนอซื้อสินค้า", "Approve Purchase Requisition", "/procurement/requisition-approval", "approval"),
          tx("purchase-order-cancel", "ยกเลิกใบสั่งซื้อสินค้า", "Cancel Purchase Order", "/procurement/order-cancellation", "approval"),
        ],
      },
      {
        id: "po-transactions",
        title: { key: "menu_purchase_expense_transactions", th: "บันทึกซื้อและค่าใช้จ่าย", en: "Purchase & Expense Transactions" },
        items: [
          tx("purchase", "บันทึกซื้อสินค้า,บริการ", "Purchase", "/transaction/purchase"),
          tx("expense-record", "บันทึกจ่ายเงินอื่นๆ", "Expense Record", "/transaction/expense"),
          tx("recurring-expense", "ค่าใช้จ่ายประจำ", "Recurring Expenses", "/transaction/recurringexpense"),
          tx("purchase-partial", "บันทึกรับสินค้าจากการซื้อ", "Gradual Receipt", "/transaction/purchasepartial"),
          tx("accrual-receive", "บันทึกตั้งหนี้จากการซื้อ", "Set Debt from Receipt", "/transaction/accrualreceive"),
          tx("purchase-credit-note", "บันทึกเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้)", "Purchase Credit Note", "/transaction/purchasecreditnote"),
          tx("purchase-debit-note", "บันทึกส่งคืนสินค้า/ลดหนี้", "Purchase Debit Note", "/transaction/purchasedebitnote"),
          tx("purchase-return", "บันทึกส่งคืนสินค้า", "Purchase Return", "/transaction/purchasereturn"),
          tx("document-vault", "คลังเอกสารและสแกนบิล", "Document Vault & Scan", "/transaction/documentvault"),
          tx("inter-company-inbox", "กล่องรับเอกสารระหว่างกิจการ", "Inter-Company Document Inbox", "/transaction/documentinbox"),
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
        id: "po-payment",
        title: { key: "menu_advances_deposits", th: "เงินมัดจำและจ่ายล่วงหน้า", en: "Advances & Deposits" },
        items: [
          tx("advance-payment", "บันทึกใบจ่ายเงินล่วงหน้า", "Advance Payment", "/transaction/advancepayment", "finance"),
          tx("advance-payment-refund", "บันทึกรับคืนเงินจ่ายล่วงหน้า", "Advance Refund", "/transaction/advancepaymentrefund", "finance"),
          tx("deposit", "บันทึกใบจ่ายเงินมัดจำ", "Pay Deposit", "/transaction/deposit", "finance"),
          tx("deposit-refund", "รับคืนเงินมัดจำ", "Deposit Refund", "/transaction/depositrefund", "finance"),
        ],
      },
      {
        id: "po-reports",
        title: { key: "menu_purchase_reports", th: "รายงานจัดซื้อ", en: "Purchase Reports" },
        items: [
          tx("purchase-report", "รายงานซื้อ", "Purchase Report", "/report/reportdedebipurchase", "report"),
          tx("purchase-by-product", "รายงานซื้อตามสินค้า", "Purchase by Product", "/report/purchasebyproduct", "report"),
          tx("gradual-receipt", "รายงานทยอยรับ", "Gradual Receipt Report", "/report/reportdedebipurchasepartial", "report"),
          tx("expense-summary-report", "รายงานสรุปรายจ่าย", "Expense Summary Report", "/report/expensesummary", "report"),
        ],
      },
    ],
  },
  {
    id: "bill",
    title: { key: "menu_order_entry_billing", th: "ใบสั่งของ/ใบกำกับสินค้า", en: "Order Entry & Billing (BILL/OE)" },
    groups: [
      {
        id: "bill-transactions",
        title: { key: "menu_sales_billing_operations", th: "งานขายและออกบิล", en: "Sales & Billing Operations" },
        items: [
          tx("quotation", "บันทึกใบเสนอราคาสินค้า", "Quotation", "/transaction/quotation"),
          tx("sale-order", "บันทึกใบสั่งขาย/สั่งจองสินค้า", "Sale Order", "/transaction/saleorder"),
          tx("sale", "บันทึกขายสินค้า,บริการ", "Sale", "/transaction/sale"),
          tx("recurring-invoice", "เอกสารขายประจำ", "Recurring Invoices", "/transaction/recurringinvoice"),
          tx("tax-invoice", "ใบเสร็จรับเงิน/ใบกำกับภาษี", "Tax Invoice / Receipt", "/transaction/taxinvoice"),
          tx("sale-return", "บันทึกใบรับคืนสินค้า,ลดหนี้", "Sale Return", "/transaction/salereturn"),
          tx("etax-invoice", "ใบกำกับภาษีอิเล็กทรอนิกส์", "e-Tax Invoice & e-Receipt", "/transaction/etax"),
          tx("combined-receipt", "ใบเสร็จรวม", "Combined Receipt", "/transaction/combinedreceipt", "finance", "combined_receipt"),
        ],
      },
      {
        id: "bill-approvals",
        title: ml("bill-approvals", "อนุมัติและยกเลิกการขาย", "Sales Approvals & Cancellations"),
        items: [
          tx("quotation-approve", "อนุมัติใบเสนอราคาสินค้า", "Approve Quotation", "/sales/quotation-approval", "approval"),
          tx("quotation-cancel", "ยกเลิกใบเสนอราคาสินค้า", "Cancel Quotation", "/sales/quotation-cancellation", "approval"),
          tx("sale-order-approve", "อนุมัติใบสั่งขาย/สั่งจองสินค้า", "Approve Sales Order / Reservation", "/sales/order-approval", "approval"),
          tx("sale-order-cancel", "ยกเลิกใบสั่งขาย/สั่งจองสินค้า", "Cancel Sales Order / Reservation", "/sales/order-cancellation", "approval"),
        ],
      },
      {
        id: "bill-delivery",
        title: ml("bill-delivery", "สั่งจองและกำหนดส่งสินค้า", "Reservations & Delivery Dates"),
        items: [
          tx("sale-reservation-flow", "Flow ใบสั่งจองสินค้า", "Sales Reservation Tracking", "/sales/reservations", "transaction"),
          tx("sale-order-date-check", "ตรวจสอบวันที่ใบสั่งขาย/สั่งจองสินค้า", "Check Sales Order / Reservation Dates", "/sales/order-dates", "transaction"),
          tx("sale-delivery-date", "ปรับปรุงวันที่ส่งของให้ลูกค้า", "Adjust Customer Delivery Date", "/sales/delivery-dates", "transaction"),
        ],
      },
      {
        id: "bill-adjustments",
        title: { key: "menu_credit_debit_notes", th: "ใบลดหนี้และเพิ่มหนี้", en: "Credit & Debit Notes" },
        items: [
          tx("credit-note", "ใบลดหนี้", "Credit Note", "/transaction/creditnote"),
          tx("debit-note", "บันทึกใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้)", "Debit Note", "/transaction/debitnote"),
        ],
      },
      {
        id: "bill-payment",
        title: { key: "menu_advances_deposits_received", th: "เงินมัดจำและรับล่วงหน้า", en: "Advances & Deposits Received" },
        items: [
          tx("receive-advance", "บันทึกใบรับเงินล่วงหน้า", "Receive Advance", "/transaction/paidadvance", "finance"),
          tx("return-advance", "บันทึกคืนเงินรับล่วงหน้า", "Return Advance", "/transaction/paidadvancerefund", "finance"),
          tx("receive-deposit", "บันทึกใบรับเงินมัดจำ", "Receive Deposit", "/transaction/receivedeposit", "finance"),
          tx("return-deposit", "คืนเงินมัดจำ", "Return Deposit", "/transaction/receivedepositrefund", "finance"),
        ],
      },
      {
        id: "bill-reports",
        title: { key: "menu_sales_reports_analytics", th: "รายงานขายและวิเคราะห์การขาย", en: "Sales Reports & Analytics" },
        items: [
          tx("sales", "รายงานขาย", "Sales Report", "/report/reportdedebisales", "report"),
          tx("sales-daily", "รายงานรายวันขายสินค้าและบริการ", "Daily Sales", "/report/reportdedebisalesdaily", "report"),
          tx("sales-by-seller", "รายงานยอดขายสุทธิ-เรียงตามพนักงานขาย", "Sales by Seller", "/report/salesbyseller", "report", "sales_by_seller"),
          tx("sales-by-document", "รายงานขายตามเอกสาร", "Sales by Document", "/report/salesreportbydocument", "report", "sales_by_document"),
          tx("gross-profit-by-document", "กำไรขั้นต้นตามเอกสาร", "Gross Profit by Document", "/report/reportgrossprofitbydocument", "report"),
          tx("gross-profit-by-product", "กำไรขั้นต้นตามสินค้า", "Gross Profit by Product", "/report/reportgrossprofitbyproduct", "report"),
          tx("sale-return-report", "รายงานคืนขาย", "Sale Return Report", "/report/reportdedebisalereturn", "report"),
          tx("sales-by-customer", "ยอดขายตามลูกค้า", "Sales by Customer", "/report/salesbycustomer", "report"),
          tx("sales-by-channel", "ยอดขายตามช่องทางขาย", "Sales by Channel", "/report/salesbychannel", "report"),
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
          tx("creditor-group", "กลุ่มเจ้าหนี้", "Creditor Group", "/creditorgroup", "master"),
          tx("creditor-beginning-balance", "บันทึกเอกสารเจ้าหนี้ยกมา", "Creditor Beginning Balance", "/creditorbeginningbalance", "master"),
          tx("import-partner", "นำเข้ารายชื่อคู่ค้า", "Import Trade Partners", "/importpartner", "master"),
        ],
      },
      {
        id: "ap-adjustments",
        title: ml("ap-adjustments", "ตั้งหนี้อื่น รับวางบิล และตัดหนี้", "Other Payables, Billing & Write-offs"),
        items: [
          tx("ap-other-debt", "บันทึกเอกสารตั้งเจ้าหนี้อื่นๆ", "Other Payables", "/transaction/apotherdebt", "finance"),
          tx("ap-billing-receipt", "บันทึกใบรับวางบิล", "Supplier Billing Receipt", "/transaction/apbillingreceipt", "finance"),
          tx("ap-bad-debt", "บันทึกตัดหนี้สูญ(เจ้าหนี้)", "Write Off Payables", "/transaction/apbaddebt", "finance"),
        ],
      },
      {
        id: "ap-payment",
        title: { key: "menu_payment_settlements", th: "การเงินและการจ่ายชำระหนี้", en: "Payment & Settlements" },
        items: [
          tx("payment-voucher", "ใบสำคัญจ่าย", "Payment Voucher", "/transaction/paymentvoucher", "finance"),
          tx("combined-payment", "ใบรวมจ่าย", "Combined Payment", "/transaction/combinedpayment", "finance", "combined_payment"),
          tx("pay", "บันทึกจ่ายชำระหนี้", "Pay", "/transaction/pay", "finance"),
        ],
      },
      {
        id: "ap-reports",
        title: { key: "menu_ap_reports", th: "รายงานเจ้าหนี้", en: "AP Reports" },
        items: [
          tx("ap-aging", "รายงานอายุเจ้าหนี้", "AP Aging Report", "/report/apaging", "report"),
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
          tx("debtor-group", "กลุ่มลูกหนี้", "Debtor Group", "/debtorgroup", "master"),
          tx("debtor-beginning-balance", "บันทึกเอกสารลูกหนี้ยกมา", "Debtor Beginning Balance", "/debtorbeginningbalance", "master"),
        ],
      },
      {
        id: "ar-adjustments",
        title: ml("ar-adjustments", "ตั้งหนี้อื่นและตัดหนี้ลูกหนี้", "Other Receivables & Write-offs"),
        items: [
          tx("ar-other-debt", "บันทึกเอกสารตั้งลูกหนี้อื่นๆ", "Other Receivables", "/transaction/arotherdebt", "finance"),
          tx("ar-bad-debt", "บันทึกตัดหนี้สูญ(ลูกหนี้)", "Write Off Receivables", "/transaction/arbaddebt", "finance"),
        ],
      },
      {
        id: "ar-transactions",
        title: { key: "menu_billing_invoicing", th: "งานวางบิลและแจ้งหนี้", en: "Billing & Invoicing" },
        items: [
          tx("billing-note", "บันทึกใบวางบิล", "Billing Note", "/transaction/billingnote"),
          tx("sale-invoice", "ใบแจ้งหนี้", "Sale Invoice", "/transaction/saleinvoice"),
        ],
      },
      {
        id: "ar-payment",
        title: { key: "menu_receipts_collections", th: "การเงินและการรับชำระหนี้", en: "Receipts & Collections" },
        items: [
          tx("paid", "บันทึกใบเสร็จรับเงิน/รับชำระหนี้", "Receive Payment", "/transaction/paid", "finance"),
          tx("temporary-receipt", "บันทึกใบเสร็จชั่วคราว", "Temporary Receipt", "/transaction/temporaryreceipt", "finance"),
        ],
      },
      {
        id: "ar-reports",
        title: { key: "menu_ar_reports", th: "รายงานลูกหนี้", en: "AR Reports" },
        items: [
          tx("ar-aging", "รายงานอายุลูกหนี้", "AR Aging Report", "/report/araging", "report"),
          tx("payment-daily", "รายงานการรับชำระหนี้ประจำวัน", "Daily Payment", "/report/reportdedebipaymentdaily", "report"),
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
        id: "cash-bank-master",
        title: { key: "menu_bank_accounts", th: "บัญชีเงินฝากและสมุดบัญชี", en: "Bank Accounts" },
        items: [
          tx("book-bank", "บันทึกสมุดเงินฝากธนาคาร", "Bank Book", "/bookbankscreen", "master", "book_bank"),
          tx("bank-contact", "บันทึกรายละเอียดผู้ติดต่อธนาคาร", "Bank Contacts", "/banking/contacts", "master"),
        ],
      },
      {
        id: "cash-management",
        title: { key: "menu_cash_advances", th: "การจัดการเงินสดและทดรอง", en: "Cash & Advances" },
        items: [
          tx("petty-cash", "เงินสดย่อย", "Petty Cash", "/pettycashscreen", "finance"),
          tx("director-advance", "บันทึกขอเบิกเงินทดรองจ่ายกรรมการ", "Director Advance", "/transaction/directoradvance", "finance", "director_advance"),
          tx("employee-advance", "บันทึกขอเบิกเงินทดรองจ่าย", "Employee Advance", "/transaction/employeeadvance", "finance"),
          tx("credit-card-expense", "บัตรเครดิตกิจการ", "Corporate Credit Card", "/transaction/creditcardexpense", "finance", "credit_card_expense"),
          tx("cash-drawer", "รับ-ส่งเงิน POS", "POS Cash Drawer", "/cashinginthedrawer", "finance"),
          tx("other-income-receipt", "บันทึกรับเงินอื่นๆ", "Other Income Receipt", "/transaction/otherincomereceipt", "finance"),
          tx("petty-cash-limit", "กำหนดวงเงินสดย่อย", "Petty Cash Limit", "/banking/petty-cash-limit", "finance"),
        ],
      },
      {
        id: "bank-transactions",
        title: { key: "menu_banking_transactions", th: "ธุรกรรมธนาคาร", en: "Banking Transactions" },
        items: [
          tx("account-transfer", "บันทึกโอนเงินระหว่างธนาคาร", "Account Transfer", "/transaction/accounttransfer", "finance"),
          tx("bank-payment-file", "ไฟล์โอนเงินจ่ายผ่านธนาคาร", "Bank Payment File", "/transaction/bankpaymentfile", "finance"),
          tx("slip-in", "รูปสลิปเงินเข้า", "Money In Slip", "/slipmoneyin", "finance"),
          tx("slip-out", "รูปสลิปเงินออก", "Money Out Slip", "/slipmoneyout", "finance"),
          tx("bank-statement", "รายการเดินบัญชีธนาคาร", "Bank Statement", "/banking/statements", "finance"),
          tx("bank-reconcile", "กระทบยอดเงินฝากธนาคาร", "Bank Reconciliation", "/banking/reconciliation", "finance"),
          tx("cash-deposit", "บันทึกนำฝากเงินสด", "Deposit Cash", "/banking/cash-deposit", "finance"),
          tx("cash-withdraw", "บันทึกถอนเงินสด", "Withdraw Cash", "/banking/cash-withdrawal", "finance"),
          tx("bank-charges", "บันทึกค่าใช้จ่ายธนาคาร", "Bank Charges", "/banking/charges", "finance"),
          tx("bank-income", "บันทึกรายได้จากธนาคาร", "Bank Income", "/banking/income", "finance"),
        ],
      },
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
        id: "bank-utilities",
        title: ml("bank-utilities", "ตรวจและประมวลผลเงินฝากและเช็ค", "Bank & Cheque Utilities"),
        items: [
          tx("cheque-recalculate", "คำนวณยอดคงเหลือของเช็คใหม่", "Recalculate Cheque Balances", "/tools/cheque-balances", "finance"),
          tx("bank-recalculate", "คำนวณยอดคงเหลือของสมุดบัญชีใหม่", "Recalculate Bank Book Balances", "/tools/bank-balances", "finance"),
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
          tx("import-product", "นำเข้ารายการสินค้า", "Import Product List", "/importproduct", "master"),
          tx("import-product-file", "นำเข้าสินค้าจากไฟล์", "Import Product File", "/importproductfromfile", "master"),
          tx("import-product-image", "นำเข้ารูปสินค้า", "Import Product Image", "/importproductimage", "master"),
        ],
      },
      {
        id: "ic-requests",
        title: ml("ic-requests", "ขอเบิก ขอโอน และตรวจรับสินค้า", "Stock Requests & Inspection"),
        items: [
          tx("stock-issue-request", "บันทึกใบขอเบิกใช้สินค้า,วัตถุดิบ", "Stock / Material Issue Request", "/inventory/issue-request", "transaction"),
          tx("stock-transfer-request", "บันทึกใบขอโอนสินค้า", "Stock Transfer Request", "/inventory/transfer-request", "transaction"),
          tx("goods-inspection", "ใบตรวจรับสินค้า", "Goods Inspection Note", "/inventory/goods-inspection", "transaction"),
        ],
      },
      {
        id: "ic-transactions",
        title: { key: "menu_stock_operations", th: "งานประจำสินค้าและคลัง", en: "Stock Operations" },
        items: [
          tx("stock-balance", "บันทึกสินค้ายกมา", "Stock Balance", "/transaction/stockbalance"),
          tx("stock-receive", "บันทึกรับสินค้าสำเร็จรูป", "Stock Receive", "/transaction/stockreceiveproduct"),
          tx("stock-pickup", "บันทึกเบิกใช้สินค้า,วัตถุดิบ", "Stock Pickup", "/transaction/stockpickupproduct"),
          tx("stock-return", "บันทึกรับคืนสินค้า,วัตถุดิบ", "Stock Return", "/transaction/stockreturnproduct"),
          tx("stock-transfer", "บันทึกโอนสินค้าระหว่างคลัง", "Stock Transfer", "/transaction/stocktransfer"),
          tx("fifo-cost-layers", "ชั้นต้นทุนสต็อกเข้าก่อนออกก่อน", "FIFO Cost Layers", "/inventory/cost-layers"),
          tx("stock-lot", "จัดการล็อตและวันหมดอายุ", "Lot & Expiry Management", "/inventory/lotexpiry", "transaction", "stock_lot"),
        ],
      },
      {
        id: "ic-counting",
        title: ml("ic-counting", "ตรวจนับและปรับปรุงสินค้า", "Stock Counting & Adjustments"),
        items: [
          tx("stock-count-sheet", "บันทึกเอกสารเพื่อตรวจนับสินค้า", "Stock Count Sheet", "/inventory/count-sheet", "transaction"),
          tx("stock-count", "บันทึกผลการตรวจนับสินค้า", "Stock Count", "/transaction/stockcount"),
          tx("stock-adjust", "บันทึกปรับปรุงสินค้าหลังตรวจนับ", "Stock Adjustment", "/transaction/adjust"),
          tx("cost-adjustment", "ปรับปรุงต้นทุนสินค้า", "Cost Adjustment", "/transaction/costadjustment"),
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
        id: "ic-pricing",
        title: ml("ic-pricing", "ราคาขายและโปรโมชั่น", "Selling Prices & Promotions"),
        items: [
          tx("product-sale-price", "กำหนดราคาขายสินค้า", "Product Selling Prices", "/inventory/selling-prices", "master"),
          tx("product-promotion", "กำหนดราคาขาย/โปรโมชั่นสินค้า", "Promotion", "/promotionscreen", "master"),
          tx("product-price-adjust", "ปรับปรุงราคาขายสินค้า", "Adjust Selling Prices", "/inventory/price-adjustment", "master"),
        ],
      },
      {
        id: "ic-reports",
        title: { key: "menu_inventory_reports", th: "รายงานสินค้า", en: "Inventory Reports" },
        items: [
          tx("stock-balance-item", "รายงานยอดคงเหลือสินค้า-ตามสินค้า", "Stock by Item", "/report/stockbalanceitem", "report"),
          tx("stock-balance-warehouse", "รายงานยอดคงเหลือสินค้า-ตามคลัง,สินค้า", "Stock by Warehouse", "/report/stockbalancewarehouse", "report"),
          tx("stock-balance-location", "คงเหลือตามที่เก็บสินค้า", "Stock by Storage Zone", "/report/stockbalancelocation", "report"),
          tx("inventory", "รายงานสินค้าคงเหลือ", "Inventory Report", "/report/reportdedebistockbalance", "report"),
          tx("low-stock-alert", "รายงานยอดคงเหลือสินค้าที่ถึงจุดสั่งซื้อ", "Low Stock Alert", "/report/lowstock", "report"),
          tx("expiring-stock-alert", "สินค้าใกล้หมดอายุ", "Expiring Stock Alert", "/report/expiringstock", "report", "expiring_stock_alert"),
          tx("stock-movement-cost", "รายงานบัญชีคุมพิเศษสินค้า", "Stock Movement with Cost", "/report/stockmovementcost", "report"),
          tx("product-movement", "รายงานเคลื่อนไหวสินค้า", "Product Movement", "/report/reportstockmovement", "report"),
          tx("stock-lot-movement", "ความเคลื่อนไหวล็อตสินค้า", "Stock Lot Movement", "/report/stocklotmovement", "report", "stock_lot_movement"),
        ],
      },
      {
        id: "ic-utilities",
        title: ml("ic-utilities", "ตรวจและประมวลผลสินค้า", "Inventory Utilities"),
        items: [
          tx("stock-daily-sequence", "กำหนดลำดับรายวัน", "Inventory Daily Sequence", "/inventory/daily-sequence", "transaction"),
          tx("reprocess", "คำนวณยอดสินค้าใหม่", "Recalculate Data", "/rebuildstockscreen", "report"),
          tx("audit-data", "ตรวจข้อมูล", "Check Data", "/auditscreen", "report"),
          tx("rebuild-products", "สร้างรายการสินค้าใหม่", "Rebuild Product List", "/rebuildproductsscreen", "report"),
          tx("rebuild-product-balance", "สร้างยอดคงเหลือใหม่", "Rebuild Stock Balance", "/rebuildproductbalancescreen", "report"),
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
          tx("asset-depreciation", "ประมวลผลสินทรัพย์", "Asset Depreciation", "/asset/depreciation", "master"),
          tx("asset-purchase", "บันทึกซื้อสินทรัพย์", "Asset Purchase Record", "/asset/purchase", "master", "asset_purchase"),
          tx("asset-construction-in-progress", "บันทึกสินทรัพย์ระหว่างก่อสร้าง", "Construction in Progress", "/asset/cip", "master", "asset_construction_in_progress"),
          tx("asset-disposal", "จำหน่ายและตัดสินทรัพย์", "Asset Disposal", "/asset/disposal", "master"),
          tx("asset-maintenance", "บันทึกรายการซ่อมบำรุง", "Asset Maintenance", "/asset/maintenance", "transaction"),
          tx("asset-type", "กำหนดประเภทสินทรัพย์", "Asset Types", "/asset/types", "master"),
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
        ],
      },
    ],
  },
  {
    id: "vat",
    title: { key: "menu_vat_taxes", th: "ภาษีมูลค่าเพิ่ม", en: "VAT & Taxes" },
    groups: [
      {
        id: "vat-transactions",
        title: { key: "menu_tax_registers", th: "บันทึกภาษีและทะเบียนภาษี", en: "Tax Registers" },
        items: [
          tx("purchase-tax-invoice-register", "ทะเบียนใบกำกับภาษีซื้อ", "Purchase Tax Invoice Register", "/transaction/purchasetaxinvoice"),
          tx("withholding-tax-deduction", "ภาษีหัก ณ ที่จ่าย", "Withholding Tax", "/transaction/withholdingtax", "finance"),
          tx("purchase-vat-adjust", "ปรับปรุงภาษีซื้อ", "Adjust Purchase VAT", "/transaction/purchasevatadjustment", "finance"),
        ],
      },
      {
        id: "vat-reports",
        title: { key: "menu_tax_reports_returns", th: "รายงานภาษีและแบบยื่น", en: "Tax Reports & Returns" },
        items: [
          tx("vat-sale", "รายงานภาษีขาย", "VAT Sale", "/report/reportvatsale", "report"),
          tx("vat-buy", "รายงานภาษีซื้อ", "VAT Purchase", "/report/reportvatbuy", "report"),
          tx("unreceived-tax-invoice", "ค่าใช้จ่ายยังไม่ได้รับใบกำกับ", "Unreceived Tax Invoices", "/report/unreceivedtaxinvoice", "report"),
          tx("vat-pp30", "แบบยื่นภาษี ภ.พ.30", "VAT Return (P.P.30)", "/report/vatpp30", "report"),
          tx("vat-pp36", "แบบยื่น ภ.พ.36", "P.P.36 Return", "/report/vatpp36", "report"),
          tx("vat-pnd2", "แบบยื่น ภ.ง.ด.2", "P.N.D.2 Return", "/report/vatpnd2", "report"),
          tx("vat-pnd3", "แบบยื่น ภ.ง.ด.3", "P.N.D.3 Return", "/report/vatpnd3", "report"),
          tx("vat-pnd53", "แบบยื่น ภ.ง.ด.53", "P.N.D.53 Return", "/report/vatpnd53", "report"),
          tx("wht-certificate", "หนังสือรับรองหัก ณ ที่จ่าย (50 ทวิ)", "Withholding Tax Certificate", "/report/whtcertificate", "report"),
          tx("withholding-tax-received", "รายงานภาษีถูกหัก ณ ที่จ่าย", "Withholding Tax Received", "/report/whtreceived", "report"),
          tx("withholding-tax-report", "รายงานภาษีหัก ณ ที่จ่าย", "Withholding Tax Report", "/report/wht-reports", "report"),
          tx("deferred-tax", "ภาษีเงินได้รอการตัดบัญชี", "Deferred Tax", "/report/deferredtax", "report", "deferred_tax"),
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
          tx("gl-fiscal-years", "ปีบัญชีและบัญชีปิดปี", "Fiscal Year and Closing Accounts", "/gl/fiscal-years", "finance", "gl_fiscal_year_closing_acct"),
          tx("gl-opening-balance", "ยอดยกมาทางบัญชี", "GL Opening Balances", "/gl/openingbalance", "finance"),
          tx("gl-allocation", "ปันส่วนค่าใช้จ่าย", "Cost Allocation", "/gl/allocations", "finance"),
          tx("gl-budget", "กำหนดงบประมาณ", "Annual Budget", "/gl/budget", "finance"),
          tx("gl-account-groups", "บันทึกกลุ่มผังบัญชี", "Chart of Account Groups", "/gl/account-groups", "finance"),
          tx("gl-account-mapping", "กำหนดรูปแบบการเชื่อม", "Automatic Account Mapping", "/gl/account-mapping", "finance"),
          tx("gl-product-account-groups", "กำหนดกลุ่มบัญชีสินค้า", "Product Account Groups", "/gl/product-account-groups", "finance"),
          tx("gl-annual-accumulated", "บันทึกยอดสะสมประจำปี", "Annual Accumulated Balances", "/gl/annual-balances", "finance"),
        ],
      },
      {
        id: "gl-journals",
        title: { key: "menu_journals_operations", th: "สมุดรายวันและงานประจำ", en: "Journals & Operations" },
        items: [
          tx("uv-journal", "สมุดรายวันขาย", "Sales Journal (UV)", "/gl/journal/uv", "finance"),
          tx("sv-journal", "สมุดรายวันซื้อ", "Purchase Journal (SV)", "/gl/journal/sv", "finance"),
          tx("rv-journal", "สมุดรายวันรับเงิน", "Receipt Voucher Journal (RV)", "/gl/journal/rv", "finance"),
          tx("pv-journal", "สมุดรายวันจ่ายเงิน", "Payment Voucher Journal (PV)", "/gl/journal/pv", "finance"),
          tx("jv-journal", "สมุดรายวันทั่วไป", "General Journal (JV)", "/gl/journal/jv", "finance"),
          tx("working-paper", "รายงานกระดาษทำการ", "Working Paper", "/gl/workingpaper", "finance"),
          tx("period-lock", "กำหนดงวดบัญชี", "Period Lock", "/gl/periodlock", "finance"),
          tx("financial-close", "ปิดงวดบัญชี", "Financial Period Close", "/gl/financialclose", "finance"),
          tx("daily-info", "ตรวจสอบประจำวัน", "Daily Check", "/checkdaily/dailyinfoscreen", "finance"),
        ],
      },
      {
        id: "gl-posting",
        title: ml("gl-posting", "ผ่านบัญชีและประมวลผลสิ้นปี", "Ledger Posting & Year-End"),
        items: [
          tx("gl-post", "ผ่านรายการ", "Post General Ledger", "/gl/posting", "finance"),
          tx("gl-unpost", "ยกเลิกเอกสารที่ผ่านรายการแล้ว", "Reverse General Ledger Posting", "/gl/unposting", "finance"),
          tx("gl-reprocess", "ประมวลผลข้อมูลบัญชีใหม่", "Reprocess Accounting Data", "/gl/reprocess", "finance"),
          tx("gl-recalculate-posted", "คำนวณยอดผ่านรายการใหม่", "Recalculate Posted Balances", "/gl/recalculate-posted", "finance"),
          tx("gl-year-end", "ประมวลผลสิ้นปี", "Year-End Processing", "/gl/year-end", "finance"),
        ],
      },
      {
        id: "gl-reports",
        title: { key: "menu_financial_statements_reports", th: "รายงานการเงินและงบบัญชี", en: "Financial Statements & Reports" },
        items: [
          tx("general-ledger", "รายงานบัญชีแยกประเภท", "General Ledger", "/report/ledger", "report", "general_ledger"),
          tx("trial-balance", "รายงานงบทดลอง", "Trial Balance", "/report/trialbalance", "report"),
          tx("profit-loss", "งบกำไรขาดทุน", "Profit and Loss", "/report/pnl", "report"),
          tx("balance-sheet", "งบดุล", "Balance Sheet", "/report/balancesheet", "report"),
          tx("financial-statement-designer", "รูปแบบงบการเงิน", "Financial Statement Designer", "/gl/statement-designer", "finance"),
          tx("cash-flow", "งบกระแสเงินสด", "Cash Flow Statement", "/report/cashflow", "report"),
          tx("cash-flow-forecast", "ประมาณการกระแสเงินสด", "Cash Flow Forecast", "/report/cashflowforecast", "report"),
          tx("financial-graphs", "กราฟประกอบงบการเงิน", "Financial Statement Graphs", "/report/financialgraphs", "report"),
          tx("project-pnl", "กำไรขาดทุนตามโครงการ", "Project P&L", "/report/project-pnl", "report"),
          tx("dimension-pnl", "กำไรขาดทุนตามสาขาและแผนก", "P&L by Branch and Department", "/report/dimensionpnl", "report"),
          tx("project-summary-report", "สรุปภาพรวมโครงการ", "Project Summary Report", "/report/projectsummary", "report"),
          tx("business-dashboard", "ภาพรวมธุรกิจ", "Business Dashboard", "/report/dashboard", "report"),
          tx("executive-summary", "วิเคราะห์ธุรกิจสำหรับผู้บริหาร", "Executive Business Analysis", "/report/executivesummary", "report"),
          tx("xbrl-export", "ส่งออกงบการเงิน XBRL (ยื่น DBD)", "XBRL Export for DBD", "/report/xbrl", "report"),
          tx("data-backup-export", "สำรองและส่งออกข้อมูล", "Backup & Export Data", "/tools/databackup", "master", "data_backup_export"),
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
