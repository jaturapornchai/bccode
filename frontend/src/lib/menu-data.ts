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
  "purchase-requisition": ["บันทึกใบเสนอซื้อสินค้า"],
  "rfq": ["บันทึกใบสืบราคาสินค้ารวม", "ระบบสืบราคาซื้อสินค้า"],
  "purchase-partial": ["บันทึกรับสินค้าจากการซื้อ"],
  "accrual-receive": ["บันทึกตั้งหนี้จากการซื้อ"],
  "purchase-landed-cost": ["บันทึกต้นทุนแฝง (Weight cost)"],
  "sale-order": ["บันทึกใบสั่งขาย/สั่งจองสินค้า", "ใบสั่งจอง"],
  "sale-reservation-flow": ["Flow ใบสั่งจองสินค้า"],
  "creditor": ["รายละเอียดเจ้าหนี้"],
  "debtor": ["รายละเอียดลูกหนี้"],
  "temporary-receipt": ["บันทึกใบเสร็จชั่วคราว"],
  "cheque-redeposit": ["บันทึกนำเช็คเข้าใหม่", "บันทึกนำเช็คฝากใหม่"],
  "product-set-assemble": ["ตรวจสอบสินค้าที่สามารถรวมเป็นสินค้าชุดได้"],
  "product-set-disassemble": ["ตรวจสอบชุดสินค้าที่สามารถแยกเป็นสินค้าได้"],
  "product-serial-registry": ["บันทึก Serial Number", "เลขซีเรียล"],
  "product-promotion": ["กำหนดราคาขาย/โปรโมชั่นสินค้า"],
  "asset-post-gl": ["โอนข้อมูลเข้าสู่ GL", "Post GL"],
  "gl-budget": ["กำหนดงบประมาณ", "Budgeting"],
  "gl-account-mapping": ["กำหนดรูปแบบการเชื่อม", "Account Mapping"],
  "gl-unpost": ["ยกเลิกเอกสารที่ผ่านรายการแล้ว"],
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
    title: { th: "ซื้อ/สั่งซื้อสินค้า", en: "Purchase Order (PO)" },
    groups: [
      {
        id: "po-procurement",
        title: { th: "งานจัดซื้อจัดหา", en: "Procurement Operations" },
        items: [
          tx("procurement-dashboard", "ภาพรวมจัดซื้อ", "Purchase Overview", "/procurement/dashboard"),
          tx("purchase-requisition", "ใบขอซื้อ", "Purchase Requisition", "/transaction/purchaserequisition"),
          tx("rfq", "สืบราคาและเจรจา", "Price Inquiry & Negotiation", "/transaction/rfq"),
          tx("purchase-order", "ใบสั่งซื้อ", "Purchase Order", "/transaction/purchaseorder"),
          tx("import-documents", "นำเข้าเอกสารจากไฟล์", "Import Documents from File", "/importdocuments", "master"),
          tx("purchase-price-comparison", "ตารางเปรียบเทียบราคาซื้อ", "Purchase Price Comparison", "/procurement/price-comparison", "transaction"),
          tx("purchase-order-generate", "ประมวลผลใบสั่งซื้ออัตโนมัติ", "Generate Purchase Orders", "/procurement/generate-orders", "transaction"),
        ],
      },
      {
        id: "po-approvals",
        title: ml("po-approvals", "อนุมัติและยกเลิกการซื้อ", "Purchase Approvals & Cancellations"),
        items: [
          tx("purchase-requisition-approve", "อนุมัติใบเสนอซื้อสินค้า", "Approve Purchase Requisition", "/procurement/requisition-approval", "approval"),
          tx("purchase-order-cancel", "ยกเลิกใบสั่งซื้อสินค้า", "Cancel Purchase Order", "/procurement/order-cancellation", "approval"),
        ],
      },
      {
        id: "po-transactions",
        title: { th: "บันทึกซื้อและค่าใช้จ่าย", en: "Purchase & Expense Transactions" },
        items: [
          tx("purchase", "ซื้อสินค้า", "Purchase", "/transaction/purchase"),
          tx("expense-record", "บันทึกค่าใช้จ่าย", "Expense Record", "/transaction/expense"),
          tx("recurring-expense", "ค่าใช้จ่ายประจำ", "Recurring Expenses", "/transaction/recurringexpense"),
          tx("purchase-partial", "รับสินค้าแบบทยอยรับ", "Gradual Receipt", "/transaction/purchasepartial"),
          tx("accrual-receive", "ตั้งหนี้จากทยอยรับ", "Set Debt from Receipt", "/transaction/accrualreceive"),
          tx("purchase-credit-note", "ใบเพิ่มหนี้เจ้าหนี้", "Purchase Credit Note", "/transaction/purchasecreditnote"),
          tx("purchase-debit-note", "ใบลดหนี้เจ้าหนี้", "Purchase Debit Note", "/transaction/purchasedebitnote"),
          tx("purchase-return", "คืนซื้อ", "Purchase Return", "/transaction/purchasereturn"),
          tx("document-vault", "คลังเอกสารและสแกนบิล", "Document Vault & Scan", "/transaction/documentvault"),
          tx("inter-company-inbox", "กล่องรับเอกสารระหว่างกิจการ", "Inter-Company Document Inbox", "/transaction/documentinbox"),
        ],
      },
      {
        id: "po-costs",
        title: ml("po-costs", "ต้นทุนแฝงและปรับปรุงใบรับสินค้า", "Landed Costs & Receipt Adjustments"),
        items: [
          tx("purchase-landed-cost", "บันทึกต้นทุนแฝง", "Landed Cost Allocation", "/transaction/landedcost", "transaction"),
          tx("purchase-receipt-adjustment", "ปรับปรุงใบรับสินค้า", "Adjust Purchase Receipt", "/transaction/purchasereceiptadjustment", "transaction"),
        ],
      },
      {
        id: "po-payment",
        title: { th: "เงินมัดจำและจ่ายล่วงหน้า", en: "Advances & Deposits" },
        items: [
          tx("advance-payment", "จ่ายเงินล่วงหน้า", "Advance Payment", "/transaction/advancepayment", "finance"),
          tx("advance-payment-refund", "รับคืนเงินล่วงหน้า", "Advance Refund", "/transaction/advancepaymentrefund", "finance"),
          tx("deposit", "จ่ายเงินมัดจำ", "Pay Deposit", "/transaction/deposit", "finance"),
          tx("deposit-refund", "รับคืนเงินมัดจำ", "Deposit Refund", "/transaction/depositrefund", "finance"),
        ],
      },
      {
        id: "po-reports",
        title: { th: "รายงานจัดซื้อ", en: "Purchase Reports" },
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
    title: { th: "ใบสั่งของ/ใบกำกับสินค้า", en: "Order Entry & Billing (BILL/OE)" },
    groups: [
      {
        id: "bill-transactions",
        title: { th: "งานขายและออกบิล", en: "Sales & Billing Operations" },
        items: [
          tx("quotation", "ใบเสนอราคา", "Quotation", "/transaction/quotation"),
          tx("sale-order", "ใบสั่งขาย", "Sale Order", "/transaction/saleorder"),
          tx("sale", "ขายสินค้า", "Sale", "/transaction/sale"),
          tx("recurring-invoice", "เอกสารขายประจำ", "Recurring Invoices", "/transaction/recurringinvoice"),
          tx("tax-invoice", "ใบเสร็จรับเงิน/ใบกำกับภาษี", "Tax Invoice / Receipt", "/transaction/taxinvoice"),
          tx("sale-return", "คืนขาย", "Sale Return", "/transaction/salereturn"),
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
          tx("sale-reservation-flow", "ติดตามใบสั่งจองสินค้า", "Sales Reservation Tracking", "/sales/reservations", "transaction"),
          tx("sale-order-date-check", "ตรวจสอบวันที่ใบสั่งขาย/สั่งจอง", "Check Sales Order / Reservation Dates", "/sales/order-dates", "transaction"),
          tx("sale-delivery-date", "ปรับปรุงวันที่ส่งของให้ลูกค้า", "Adjust Customer Delivery Date", "/sales/delivery-dates", "transaction"),
        ],
      },
      {
        id: "bill-adjustments",
        title: { th: "ใบลดหนี้และเพิ่มหนี้", en: "Credit & Debit Notes" },
        items: [
          tx("credit-note", "ใบลดหนี้", "Credit Note", "/transaction/creditnote"),
          tx("debit-note", "ใบเพิ่มหนี้", "Debit Note", "/transaction/debitnote"),
        ],
      },
      {
        id: "bill-payment",
        title: { th: "เงินมัดจำและรับล่วงหน้า", en: "Advances & Deposits Received" },
        items: [
          tx("receive-advance", "รับเงินล่วงหน้า", "Receive Advance", "/transaction/paidadvance", "finance"),
          tx("return-advance", "คืนเงินล่วงหน้า", "Return Advance", "/transaction/paidadvancerefund", "finance"),
          tx("receive-deposit", "รับเงินมัดจำ", "Receive Deposit", "/transaction/receivedeposit", "finance"),
          tx("return-deposit", "คืนเงินมัดจำ", "Return Deposit", "/transaction/receivedepositrefund", "finance"),
        ],
      },
      {
        id: "bill-reports",
        title: { th: "รายงานขายและวิเคราะห์การขาย", en: "Sales Reports & Analytics" },
        items: [
          tx("sales", "รายงานขาย", "Sales Report", "/report/reportdedebisales", "report"),
          tx("sales-daily", "รายงานขายรายวัน", "Daily Sales", "/report/reportdedebisalesdaily", "report"),
          tx("sales-by-seller", "ยอดขายตามพนักงาน", "Sales by Seller", "/report/salesbyseller", "report", "sales_by_seller"),
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
    title: { th: "เจ้าหนี้", en: "Account Payable (AP)" },
    groups: [
      {
        id: "ap-master",
        title: { th: "ข้อมูลหลักเจ้าหนี้", en: "Creditor Master Data" },
        items: [
          tx("creditor", "เจ้าหนี้", "Creditor", "/creditor", "master"),
          tx("creditor-group", "กลุ่มเจ้าหนี้", "Creditor Group", "/creditorgroup", "master"),
          tx("creditor-beginning-balance", "เจ้าหนี้ตั้งต้นรายเอกสาร", "Creditor Beginning Balance", "/creditorbeginningbalance", "master"),
          tx("import-partner", "นำเข้ารายชื่อคู่ค้า", "Import Trade Partners", "/importpartner", "master"),
        ],
      },
      {
        id: "ap-adjustments",
        title: ml("ap-adjustments", "ตั้งหนี้อื่น รับวางบิล และตัดหนี้", "Other Payables, Billing & Write-offs"),
        items: [
          tx("ap-other-debt", "ตั้งเจ้าหนี้อื่นๆ", "Other Payables", "/transaction/apotherdebt", "finance"),
          tx("ap-billing-receipt", "ใบรับวางบิลเจ้าหนี้", "Supplier Billing Receipt", "/transaction/apbillingreceipt", "finance"),
          tx("ap-bad-debt", "ตัดหนี้สูญเจ้าหนี้", "Write Off Payables", "/transaction/apbaddebt", "finance"),
        ],
      },
      {
        id: "ap-payment",
        title: { th: "การเงินและการจ่ายชำระหนี้", en: "Payment & Settlements" },
        items: [
          tx("payment-voucher", "ใบสำคัญจ่าย", "Payment Voucher", "/transaction/paymentvoucher", "finance"),
          tx("combined-payment", "ใบรวมจ่าย", "Combined Payment", "/transaction/combinedpayment", "finance", "combined_payment"),
          tx("pay", "จ่ายชำระ", "Pay", "/transaction/pay", "finance"),
        ],
      },
      {
        id: "ap-reports",
        title: { th: "รายงานเจ้าหนี้", en: "AP Reports" },
        items: [
          tx("ap-aging", "รายงานอายุเจ้าหนี้", "AP Aging Report", "/report/apaging", "report"),
        ],
      },
      {
        id: "ap-utilities",
        title: ml("ap-utilities", "ตรวจและประมวลผลเจ้าหนี้", "Payables Utilities"),
        items: [
          tx("ap-recalculate", "คำนวณยอดเจ้าหนี้ใหม่", "Recalculate Payables", "/tools/ap-recalculate", "finance"),
          tx("ap-bill-recalculate", "คำนวณยอดคงเหลือบิลเจ้าหนี้ใหม่", "Recalculate Payable Bill Balances", "/tools/ap-bill-balances", "finance"),
        ],
      },
    ],
  },
  {
    id: "ar",
    title: { th: "ลูกหนี้", en: "Account Receivable (AR)" },
    groups: [
      {
        id: "ar-master",
        title: { th: "ข้อมูลหลักลูกหนี้", en: "Debtor Master Data" },
        items: [
          tx("debtor", "ลูกหนี้", "Debtor", "/debtor", "master"),
          tx("debtor-group", "กลุ่มลูกหนี้", "Debtor Group", "/debtorgroup", "master"),
          tx("debtor-beginning-balance", "ลูกหนี้ตั้งต้นรายเอกสาร", "Debtor Beginning Balance", "/debtorbeginningbalance", "master"),
        ],
      },
      {
        id: "ar-adjustments",
        title: ml("ar-adjustments", "ตั้งหนี้อื่นและตัดหนี้ลูกหนี้", "Other Receivables & Write-offs"),
        items: [
          tx("ar-other-debt", "ตั้งลูกหนี้อื่นๆ", "Other Receivables", "/transaction/arotherdebt", "finance"),
          tx("ar-bad-debt", "ตัดหนี้สูญลูกหนี้", "Write Off Receivables", "/transaction/arbaddebt", "finance"),
        ],
      },
      {
        id: "ar-transactions",
        title: { th: "งานวางบิลและแจ้งหนี้", en: "Billing & Invoicing" },
        items: [
          tx("billing-note", "ใบวางบิล", "Billing Note", "/transaction/billingnote"),
          tx("sale-invoice", "ใบแจ้งหนี้", "Sale Invoice", "/transaction/saleinvoice"),
        ],
      },
      {
        id: "ar-payment",
        title: { th: "การเงินและการรับชำระหนี้", en: "Receipts & Collections" },
        items: [
          tx("paid", "รับชำระ", "Receive Payment", "/transaction/paid", "finance"),
          tx("temporary-receipt", "ใบเสร็จชั่วคราว", "Temporary Receipt", "/transaction/temporaryreceipt", "finance"),
        ],
      },
      {
        id: "ar-reports",
        title: { th: "รายงานลูกหนี้", en: "AR Reports" },
        items: [
          tx("ar-aging", "รายงานอายุลูกหนี้", "AR Aging Report", "/report/araging", "report"),
          tx("payment-daily", "รายงานรับชำระรายวัน", "Daily Payment", "/report/reportdedebipaymentdaily", "report"),
        ],
      },
      {
        id: "ar-utilities",
        title: ml("ar-utilities", "ตรวจและประมวลผลลูกหนี้", "Receivables Utilities"),
        items: [
          tx("ar-recalculate", "คำนวณยอดลูกหนี้ใหม่", "Recalculate Receivables", "/tools/ar-recalculate", "finance"),
          tx("ar-bill-recalculate", "คำนวณยอดคงเหลือบิลลูกหนี้ใหม่", "Recalculate Receivable Bill Balances", "/tools/ar-bill-balances", "finance"),
        ],
      },
    ],
  },
  {
    id: "cash-bank",
    title: { th: "เงินสดและธนาคาร", en: "Cash & Cheque" },
    groups: [
      {
        id: "cash-bank-master",
        title: { th: "บัญชีเงินฝากและสมุดบัญชี", en: "Bank Accounts" },
        items: [
          tx("book-bank", "สมุดบัญชี", "Bank Book", "/bookbankscreen", "master", "bank_book"),
          tx("bank-contact", "ผู้ติดต่อธนาคาร", "Bank Contacts", "/banking/contacts", "master"),
        ],
      },
      {
        id: "cash-management",
        title: { th: "การจัดการเงินสดและทดรอง", en: "Cash & Advances" },
        items: [
          tx("petty-cash", "เงินสดย่อย", "Petty Cash", "/pettycashscreen", "finance"),
          tx("director-advance", "เงินทดรองจ่ายกรรมการ", "Director Advance", "/transaction/directoradvance", "finance", "director_advance"),
          tx("employee-advance", "เงินทดรองจ่ายพนักงาน", "Employee Advance", "/transaction/employeeadvance", "finance"),
          tx("credit-card-expense", "บัตรเครดิตกิจการ", "Corporate Credit Card", "/transaction/creditcardexpense", "finance", "credit_card_expense"),
          tx("cash-drawer", "รับ-ส่งเงิน POS", "POS Cash Drawer", "/cashinginthedrawer", "finance"),
          tx("other-income-receipt", "ใบเสร็จรับเงินรายได้อื่น", "Other Income Receipt", "/transaction/otherincomereceipt", "finance"),
          tx("petty-cash-limit", "กำหนดวงเงินสดย่อย", "Petty Cash Limit", "/banking/petty-cash-limit", "finance"),
        ],
      },
      {
        id: "bank-transactions",
        title: { th: "ธุรกรรมธนาคาร", en: "Banking Transactions" },
        items: [
          tx("account-transfer", "โอนเงินระหว่างบัญชี", "Account Transfer", "/transaction/accounttransfer", "finance"),
          tx("bank-payment-file", "ไฟล์โอนเงินจ่ายผ่านธนาคาร", "Bank Payment File", "/transaction/bankpaymentfile", "finance"),
          tx("slip-in", "รูปสลิปเงินเข้า", "Money In Slip", "/slipmoneyin", "finance"),
          tx("slip-out", "รูปสลิปเงินออก", "Money Out Slip", "/slipmoneyout", "finance"),
          tx("bank-statement", "รายการเดินบัญชีธนาคาร", "Bank Statement", "/banking/statements", "finance"),
          tx("bank-reconcile", "กระทบยอดเงินฝากธนาคาร", "Bank Reconciliation", "/banking/reconciliation", "finance"),
          tx("cash-deposit", "นำฝากเงินสด", "Deposit Cash", "/banking/cash-deposit", "finance"),
          tx("cash-withdraw", "ถอนเงินสด", "Withdraw Cash", "/banking/cash-withdrawal", "finance"),
          tx("bank-charges", "บันทึกค่าใช้จ่ายธนาคาร", "Bank Charges", "/banking/charges", "finance"),
          tx("bank-income", "บันทึกรายได้จากธนาคาร", "Bank Income", "/banking/income", "finance"),
        ],
      },
      {
        id: "cheques-received",
        title: ml("cheques-received", "เช็ครับ", "Received Cheques"),
        items: [
          tx("cheque-received", "ทะเบียนเช็ครับ", "Cheques Received", "/transaction/chequereceived", "finance"),
          tx("cheque-deposit", "นำฝากเช็ครับ", "Deposit Received Cheques", "/banking/cheques/deposit", "finance"),
          tx("cheque-received-clear", "เช็ครับผ่าน", "Clear Received Cheques", "/banking/cheques/received-clear", "finance"),
          tx("cheque-received-return", "เช็ครับคืน", "Returned Received Cheques", "/banking/cheques/received-return", "finance"),
          tx("cheque-redeposit", "นำเช็คเข้าใหม่", "Redeposit Cheques", "/banking/cheques/redeposit", "finance"),
          tx("cheque-received-cancel", "ยกเลิกเช็ครับ", "Cancel Received Cheques", "/banking/cheques/received-cancel", "finance"),
          tx("cheque-discount", "ขายลดเช็ครับ", "Discount Received Cheques", "/banking/cheques/discount", "finance"),
        ],
      },
      {
        id: "cheques-issued",
        title: ml("cheques-issued", "เช็คจ่าย", "Issued Cheques"),
        items: [
          tx("cheque-issued", "ทะเบียนเช็คจ่าย", "Cheques Issued", "/transaction/chequeissued", "finance"),
          tx("cheque-issued-clear", "เช็คจ่ายผ่าน", "Clear Issued Cheques", "/banking/cheques/issued-clear", "finance"),
          tx("cheque-issued-cancel", "ยกเลิกเช็คจ่าย", "Cancel Issued Cheques", "/banking/cheques/issued-cancel", "finance"),
        ],
      },
      {
        id: "card-settlement",
        title: ml("card-settlement", "รับและขึ้นเงินบัตรเครดิต", "Credit Card Receipts & Settlement"),
        items: [
          tx("credit-card-receipts", "ทะเบียนรับชำระด้วยบัตรเครดิต", "Credit Card Receipts Register", "/banking/cards/receipts", "finance"),
          tx("credit-card-settle", "ขึ้นเงินบัตรเครดิต", "Settle Credit Card Receipts", "/banking/cards/settlement", "finance"),
          tx("credit-card-cancel", "ยกเลิกรายการรับบัตรเครดิต", "Cancel Credit Card Receipts", "/banking/cards/cancellation", "finance"),
        ],
      },
      {
        id: "bank-utilities",
        title: ml("bank-utilities", "ตรวจและประมวลผลเงินฝากและเช็ค", "Bank & Cheque Utilities"),
        items: [
          tx("cheque-recalculate", "คำนวณยอดคงเหลือเช็คใหม่", "Recalculate Cheque Balances", "/tools/cheque-balances", "finance"),
          tx("bank-recalculate", "คำนวณยอดสมุดบัญชีใหม่", "Recalculate Bank Book Balances", "/tools/bank-balances", "finance"),
        ],
      },
    ],
  },
  {
    id: "ic",
    title: { th: "สินค้าคงคลัง", en: "Inventory Control (IC)" },
    groups: [
      {
        id: "ic-master",
        title: { th: "ข้อมูลหลักสินค้า", en: "Product Catalog & Setup" },
        items: [
          tx("product", "สินค้า", "Product", "/product", "master"),
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
          tx("stock-issue-request", "ใบขอเบิกสินค้าและวัตถุดิบ", "Stock / Material Issue Request", "/inventory/issue-request", "transaction"),
          tx("stock-transfer-request", "ใบขอโอนสินค้า", "Stock Transfer Request", "/inventory/transfer-request", "transaction"),
          tx("goods-inspection", "ใบตรวจรับสินค้า", "Goods Inspection Note", "/inventory/goods-inspection", "transaction"),
        ],
      },
      {
        id: "ic-transactions",
        title: { th: "งานประจำสินค้าและคลัง", en: "Stock Operations" },
        items: [
          tx("stock-balance", "ยอดยกมาสินค้า", "Stock Balance", "/transaction/stockbalance"),
          tx("stock-receive", "รับสินค้า", "Stock Receive", "/transaction/stockreceiveproduct"),
          tx("stock-pickup", "เบิกสินค้า", "Stock Pickup", "/transaction/stockpickupproduct"),
          tx("stock-return", "คืนสินค้าเข้าคลัง", "Stock Return", "/transaction/stockreturnproduct"),
          tx("stock-transfer", "โอนสินค้า", "Stock Transfer", "/transaction/stocktransfer"),
          tx("fifo-cost-layers", "ชั้นต้นทุนสต็อกเข้าก่อนออกก่อน", "FIFO Cost Layers", "/inventory/cost-layers"),
          tx("stock-lot", "จัดการล็อตและวันหมดอายุ", "Lot & Expiry Management", "/inventory/lotexpiry", "transaction", "stock_lot"),
        ],
      },
      {
        id: "ic-counting",
        title: ml("ic-counting", "ตรวจนับและปรับปรุงสินค้า", "Stock Counting & Adjustments"),
        items: [
          tx("stock-count-sheet", "เอกสารเพื่อตรวจนับสินค้า", "Stock Count Sheet", "/inventory/count-sheet", "transaction"),
          tx("stock-count", "ตรวจนับสต็อก", "Stock Count", "/transaction/stockcount"),
          tx("stock-adjust", "ปรับปรุงสต็อก", "Stock Adjustment", "/transaction/adjust"),
          tx("cost-adjustment", "ปรับปรุงต้นทุนสินค้า", "Cost Adjustment", "/transaction/costadjustment"),
        ],
      },
      {
        id: "ic-sets",
        title: ml("ic-sets", "สินค้าชุดและส่วนประกอบ", "Product Sets & Components"),
        items: [
          tx("product-set", "สินค้าชุด", "Product Sets", "/inventory/product-sets", "master"),
          tx("product-set-assemble", "ตรวจสอบรวมสินค้าชุด", "Check Set Assembly", "/inventory/set-assembly", "transaction"),
          tx("product-set-disassemble", "ตรวจสอบแยกสินค้าชุด", "Check Set Disassembly", "/inventory/set-disassembly", "transaction"),
          tx("product-set-components", "รายการย่อยสินค้าชุดแบบที่ 2", "Product Set Components Type 2", "/inventory/set-components", "master"),
        ],
      },
      {
        id: "ic-identifiers",
        title: ml("ic-identifiers", "ทะเบียนเลขเครื่องสินค้า", "Product Serial Numbers"),
        items: [
          tx("product-serial-registry", "ทะเบียนเลขเครื่อง", "Serial Registry", "/productserialregistry", "master"),
        ],
      },
      {
        id: "ic-pricing",
        title: ml("ic-pricing", "ราคาขายและโปรโมชั่น", "Selling Prices & Promotions"),
        items: [
          tx("product-sale-price", "กำหนดราคาขายสินค้า", "Product Selling Prices", "/inventory/selling-prices", "master"),
          tx("product-promotion", "โปรโมชั่น", "Promotion", "/promotionscreen", "master"),
          tx("product-price-adjust", "ปรับปรุงราคาขายสินค้า", "Adjust Selling Prices", "/inventory/price-adjustment", "master"),
        ],
      },
      {
        id: "ic-reports",
        title: { th: "รายงานสินค้า", en: "Inventory Reports" },
        items: [
          tx("stock-balance-item", "คงเหลือตามสินค้า", "Stock by Item", "/report/stockbalanceitem", "report"),
          tx("stock-balance-warehouse", "คงเหลือตามคลัง", "Stock by Warehouse", "/report/stockbalancewarehouse", "report"),
          tx("stock-balance-location", "คงเหลือตามโซนเก็บสินค้า", "Stock by Storage Zone", "/report/stockbalancelocation", "report"),
          tx("inventory", "รายงานสินค้าคงเหลือ", "Inventory Report", "/report/reportdedebistockbalance", "report"),
          tx("low-stock-alert", "สินค้าใกล้หมดขั้นต่ำ", "Low Stock Alert", "/report/lowstock", "report"),
          tx("expiring-stock-alert", "สินค้าใกล้หมดอายุ", "Expiring Stock Alert", "/report/expiringstock", "report", "expiring_stock_alert"),
          tx("stock-movement-cost", "เคลื่อนไหวสินค้าพร้อมต้นทุน", "Stock Movement with Cost", "/report/stockmovementcost", "report"),
          tx("product-movement", "ความเคลื่อนไหวสินค้า", "Product Movement", "/report/reportstockmovement", "report"),
          tx("stock-lot-movement", "ความเคลื่อนไหวล็อตสินค้า", "Stock Lot Movement", "/report/stocklotmovement", "report", "stock_lot_movement"),
        ],
      },
      {
        id: "ic-utilities",
        title: ml("ic-utilities", "ตรวจและประมวลผลสินค้า", "Inventory Utilities"),
        items: [
          tx("stock-daily-sequence", "กำหนดลำดับรายวันสินค้า", "Inventory Daily Sequence", "/inventory/daily-sequence", "transaction"),
          tx("reprocess", "คำนวณข้อมูลใหม่", "Recalculate Data", "/rebuildstockscreen", "report"),
          tx("audit-data", "ตรวจข้อมูล", "Check Data", "/auditscreen", "report"),
          tx("rebuild-products", "สร้างรายการสินค้าใหม่", "Rebuild Product List", "/rebuildproductsscreen", "report"),
          tx("rebuild-product-balance", "สร้างยอดคงเหลือใหม่", "Rebuild Stock Balance", "/rebuildproductbalancescreen", "report"),
        ],
      },
    ],
  },
  {
    id: "fa",
    title: { th: "สินทรัพย์และค่าเสื่อมราคา", en: "Fixed Assets (FA)" },
    groups: [
      {
        id: "fa-transactions",
        title: { th: "งานสินทรัพย์ถาวร", en: "Fixed Asset Operations" },
        items: [
          tx("asset-registry", "ทะเบียนสินทรัพย์ถาวร", "Fixed Asset Registry", "/asset/registry", "master"),
          tx("asset-depreciation", "คำนวณค่าเสื่อมราคา", "Asset Depreciation", "/asset/depreciation", "master"),
          tx("asset-purchase", "บันทึกซื้อสินทรัพย์", "Asset Purchase Record", "/asset/purchase", "master", "asset_purchase"),
          tx("asset-construction-in-progress", "งานระหว่างก่อสร้าง", "Construction in Progress", "/asset/cip", "master", "asset_construction_in_progress"),
          tx("asset-disposal", "จำหน่ายและตัดสินทรัพย์", "Asset Disposal", "/asset/disposal", "master"),
          tx("asset-maintenance", "บันทึกซ่อมบำรุงสินทรัพย์", "Asset Maintenance", "/asset/maintenance", "transaction"),
          tx("asset-type", "ประเภทสินทรัพย์", "Asset Types", "/asset/types", "master"),
        ],
      },
      {
        id: "fa-posting",
        title: ml("fa-posting", "ผ่านค่าเสื่อมเข้าบัญชี", "Depreciation Posting"),
        items: [
          tx("asset-post-gl", "โอนค่าเสื่อมราคาเข้าบัญชีแยกประเภท", "Post Depreciation to General Ledger", "/asset/post-gl", "finance"),
        ],
      },
      {
        id: "fa-reports",
        title: { th: "รายงานสินทรัพย์", en: "Fixed Asset Reports" },
        items: [
          tx("fixed-asset-schedule", "ตารางค่าเสื่อมและสินทรัพย์", "Fixed Asset Schedule", "/report/assetschedule", "report", "fixed_asset_schedule"),
        ],
      },
    ],
  },
  {
    id: "vat",
    title: { th: "ภาษีมูลค่าเพิ่ม", en: "VAT & Taxes" },
    groups: [
      {
        id: "vat-transactions",
        title: { th: "บันทึกภาษีและทะเบียนภาษี", en: "Tax Registers" },
        items: [
          tx("purchase-tax-invoice-register", "ทะเบียนใบกำกับภาษีซื้อ", "Purchase Tax Invoice Register", "/transaction/purchasetaxinvoice"),
          tx("withholding-tax-deduction", "ภาษีหัก ณ ที่จ่าย", "Withholding Tax", "/transaction/withholdingtax", "finance"),
          tx("purchase-vat-adjust", "ปรับปรุงภาษีซื้อ", "Adjust Purchase VAT", "/transaction/purchasevatadjustment", "finance"),
        ],
      },
      {
        id: "vat-reports",
        title: { th: "รายงานภาษีและแบบยื่น", en: "Tax Reports & Returns" },
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
          tx("withholding-tax-received", "ทะเบียนถูกหัก ณ ที่จ่าย", "Withholding Tax Received", "/report/whtreceived", "report"),
          tx("withholding-tax-report", "รายงานภาษีหัก ณ ที่จ่าย", "Withholding Tax Report", "/report/wht-reports", "report"),
          tx("deferred-tax", "ภาษีเงินได้รอการตัดบัญชี", "Deferred Tax", "/report/deferredtax", "report", "deferred_tax"),
        ],
      },
    ],
  },
  {
    id: "gl",
    title: { th: "บัญชีแยกประเภท", en: "General Ledger (GL)" },
    groups: [
      {
        id: "gl-master",
        title: { th: "ข้อมูลหลักและยอดยกมา", en: "Master & Balances" },
        items: [
          tx("chart-of-accounts", "ผังบัญชี", "Chart of Accounts", "/gl/chartofaccounts", "finance"),
          tx("gl-opening-balance", "ยอดยกมาทางบัญชี", "GL Opening Balances", "/gl/openingbalance", "finance"),
          tx("gl-budget", "กำหนดงบประมาณประจำปี", "Annual Budget", "/gl/budget", "finance"),
          tx("gl-account-groups", "กลุ่มผังบัญชี", "Chart of Account Groups", "/gl/account-groups", "finance"),
          tx("gl-account-mapping", "รูปแบบการเชื่อมโยงบัญชีอัตโนมัติ", "Automatic Account Mapping", "/gl/account-mapping", "finance"),
          tx("gl-product-account-groups", "กลุ่มบัญชีสินค้า", "Product Account Groups", "/gl/product-account-groups", "finance"),
          tx("gl-annual-accumulated", "ยอดสะสมประจำปี", "Annual Accumulated Balances", "/gl/annual-balances", "finance"),
        ],
      },
      {
        id: "gl-journals",
        title: { th: "สมุดรายวันและงานประจำ", en: "Journals & Operations" },
        items: [
          tx("uv-journal", "สมุดรายวันขาย", "Sales Journal (UV)", "/gl/journal/uv", "finance"),
          tx("sv-journal", "สมุดรายวันซื้อ", "Purchase Journal (SV)", "/gl/journal/sv", "finance"),
          tx("rv-journal", "สมุดรายวันรับเงิน", "Receipt Voucher Journal (RV)", "/gl/journal/rv", "finance"),
          tx("pv-journal", "สมุดรายวันจ่ายเงิน", "Payment Voucher Journal (PV)", "/gl/journal/pv", "finance"),
          tx("jv-journal", "สมุดรายวันทั่วไป", "General Journal (JV)", "/gl/journal/jv", "finance"),
          tx("working-paper", "กระดาษทำการ", "Working Paper", "/gl/workingpaper", "finance"),
          tx("period-lock", "ล็อกงวดบัญชี", "Period Lock", "/gl/periodlock", "finance"),
          tx("financial-close", "ปิดงบบัญชีสิ้นงวด", "Financial Period Close", "/gl/financialclose", "finance"),
          tx("daily-info", "ตรวจสอบประจำวัน", "Daily Check", "/checkdaily/dailyinfoscreen", "finance"),
        ],
      },
      {
        id: "gl-posting",
        title: ml("gl-posting", "ผ่านบัญชีและประมวลผลสิ้นปี", "Ledger Posting & Year-End"),
        items: [
          tx("gl-post", "ผ่านรายการบัญชี", "Post General Ledger", "/gl/posting", "finance"),
          tx("gl-unpost", "ยกเลิกการผ่านรายการบัญชี", "Reverse General Ledger Posting", "/gl/unposting", "finance"),
          tx("gl-reprocess", "ประมวลผลข้อมูลบัญชีใหม่", "Reprocess Accounting Data", "/gl/reprocess", "finance"),
          tx("gl-recalculate-posted", "คำนวณยอดผ่านรายการใหม่", "Recalculate Posted Balances", "/gl/recalculate-posted", "finance"),
          tx("gl-year-end", "ประมวลผลสิ้นปี", "Year-End Processing", "/gl/year-end", "finance"),
        ],
      },
      {
        id: "gl-reports",
        title: { th: "รายงานการเงินและงบบัญชี", en: "Financial Statements & Reports" },
        items: [
          tx("general-ledger", "บัญชีแยกประเภท", "General Ledger", "/report/ledger", "report", "ledger"),
          tx("trial-balance", "งบทดลอง", "Trial Balance", "/report/trialbalance", "report"),
          tx("profit-loss", "งบกำไรขาดทุน", "Profit and Loss", "/report/pnl", "report"),
          tx("balance-sheet", "งบดุล", "Balance Sheet", "/report/balancesheet", "report"),
          tx("financial-statement-designer", "ออกแบบงบการเงิน", "Financial Statement Designer", "/gl/statement-designer", "finance"),
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
