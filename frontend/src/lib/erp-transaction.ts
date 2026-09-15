import { authFetch } from "./client-auth-session";

export interface ErpDetailItem {
  linenumber: number;
  itemcode: string;
  itemnames?: { code: string; name: string }[];
  itemname?: string;
  barcode?: string;
  unitcode: string;
  unitname?: string;
  qty: number;
  price: number;
  discount?: string;
  discountamount?: number;
  sumamount: number;
  vatcal?: number;
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
  totalamount: number;
  totalvalue?: number;
  totaldiscount?: number;
  totalvatvalue?: number;
  totalbeforevat?: number;
  totalaftervat?: number;
  vattype?: number;
  vatrate?: number;
  status?: number; // 0=Draft, 1=Approved/Completed, 2=Cancelled
  isclosed?: boolean;
  iscancel?: boolean;
  branchcode?: string;
  remark?: string;
  details: ErpDetailItem[];
}

export interface ErpModuleConfig {
  route: string;
  moduleKey: string;
  apiPath: string;
  domain: "inventory" | "sales" | "purchase" | "ar" | "ap" | "cash-bank";
  title: { th: string; en: string };
  counterpartyLabel: { th: string; en: string };
  defaultDocPrefix: string;
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
    defaultDocPrefix: "STK-BAL",
    hasLineItems: true,
  },
  {
    route: "/transaction/stockreceiveproduct",
    moduleKey: "stock_receive",
    apiPath: "stock-receive-product",
    domain: "inventory",
    title: { th: "บันทึกรับสินค้าสำเร็จรูป", en: "Stock Receive" },
    counterpartyLabel: { th: "คลังปลายทาง", en: "Target Warehouse" },
    defaultDocPrefix: "STK-RCV",
    hasLineItems: true,
  },
  {
    route: "/transaction/stockpickupproduct",
    moduleKey: "stock_pickup",
    apiPath: "stock-prickup-product",
    domain: "inventory",
    title: { th: "บันทึกเบิกใช้สินค้า,วัตถุดิบ", en: "Stock Issue / Pickup" },
    counterpartyLabel: { th: "แผนกผู้ขอเบิก", en: "Requesting Dept" },
    defaultDocPrefix: "STK-OUT",
    hasLineItems: true,
  },
  {
    route: "/transaction/stockreturnproduct",
    moduleKey: "stock_return",
    apiPath: "stock-return-product",
    domain: "inventory",
    title: { th: "บันทึกรับคืนสินค้า,วัตถุดิบ", en: "Stock Return" },
    counterpartyLabel: { th: "คลังรับคืน", en: "Receiving Warehouse" },
    defaultDocPrefix: "STK-RET",
    hasLineItems: true,
  },
  {
    route: "/transaction/stocktransfer",
    moduleKey: "stock_transfer",
    apiPath: "stock-transfer",
    domain: "inventory",
    title: { th: "บันทึกโอนสินค้าระหว่างคลัง", en: "Stock Transfer" },
    counterpartyLabel: { th: "คลังปลายทาง", en: "Destination Warehouse" },
    defaultDocPrefix: "STK-XFR",
    hasLineItems: true,
  },
  {
    route: "/transaction/adjust",
    moduleKey: "stock_adjust",
    apiPath: "stock-adjustment",
    domain: "inventory",
    title: { th: "บันทึกปรับปรุงสินค้าหลังตรวจนับ", en: "Stock Adjustment" },
    counterpartyLabel: { th: "คลังที่ตรวจนับ", en: "Audited Warehouse" },
    defaultDocPrefix: "STK-ADJ",
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
    defaultDocPrefix: "QT",
    hasLineItems: true,
  },
  {
    route: "/transaction/saleorder",
    moduleKey: "sale_order",
    apiPath: "sale-order",
    domain: "sales",
    title: { th: "บันทึกใบสั่งขาย/สั่งจองสินค้า", en: "Sales Order" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "SO",
    hasLineItems: true,
  },
  {
    route: "/transaction/sale",
    moduleKey: "sale_invoice",
    apiPath: "sale-invoice",
    domain: "sales",
    title: { th: "บันทึกขายสินค้า,บริการ", en: "Sale & Dispatch" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "INV",
    hasLineItems: true,
  },
  {
    route: "/transaction/saleinvoice",
    moduleKey: "sale_invoice_billing",
    apiPath: "sale-invoice",
    domain: "sales",
    title: { th: "ใบแจ้งหนี้ / ใบกำกับภาษี", en: "Sales Invoice / Tax Invoice" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "INV",
    hasLineItems: true,
  },
  {
    route: "/transaction/taxinvoice",
    moduleKey: "tax_invoice_receipt",
    apiPath: "sale-invoice",
    domain: "sales",
    title: { th: "ใบเสร็จรับเงิน/ใบกำกับภาษี", en: "Tax Invoice / Receipt" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "TAX",
    hasLineItems: true,
  },
  {
    route: "/transaction/salereturn",
    moduleKey: "sale_return",
    apiPath: "sale-invoice-return",
    domain: "sales",
    title: { th: "บันทึกใบรับคืนสินค้า,ลดหนี้", en: "Sale Return" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "SR",
    hasLineItems: true,
  },
  {
    route: "/transaction/creditnote",
    moduleKey: "credit_note",
    apiPath: "sale-invoice-return",
    domain: "sales",
    title: { th: "ใบลดหนี้ (Credit Note)", en: "Credit Note" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "CN",
    hasLineItems: true,
  },
  {
    route: "/transaction/debitnote",
    moduleKey: "debit_note",
    apiPath: "bank/saledebitnote",
    domain: "sales",
    title: { th: "บันทึกใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้)", en: "Debit Note" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "DN",
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
    defaultDocPrefix: "PR",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchaseorder",
    moduleKey: "purchase_order",
    apiPath: "purchase-order",
    domain: "purchase",
    title: { th: "บันทึกใบสั่งซื้อสินค้า", en: "Purchase Order" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "PO",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchase",
    moduleKey: "purchase_invoice",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "บันทึกซื้อสินค้า,บริการ", en: "Purchase Invoice / Goods Receipt" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "PUR",
    hasLineItems: true,
  },
  {
    route: "/transaction/expense",
    moduleKey: "expense_record",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "บันทึกจ่ายเงินอื่นๆ", en: "Expense Record" },
    counterpartyLabel: { th: "ผู้รับเงิน / เจ้าหนี้", en: "Payee / Vendor" },
    defaultDocPrefix: "EXP",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasereturn",
    moduleKey: "purchase_return",
    apiPath: "purchase-return",
    domain: "purchase",
    title: { th: "บันทึกส่งคืนสินค้า", en: "Purchase Return" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "PRT",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasecreditnote",
    moduleKey: "purchase_credit_note",
    apiPath: "purchase-return",
    domain: "purchase",
    title: { th: "บันทึกเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้)", en: "Vendor Credit Note" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "VCN",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasedebitnote",
    moduleKey: "purchase_debit_note",
    apiPath: "bank/purchasedebitnote",
    domain: "purchase",
    title: { th: "บันทึกส่งคืนสินค้า/ลดหนี้", en: "Vendor Debit Note" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "VDN",
    hasLineItems: true,
  },

  // 4. ลูกหนี้ (AR)
  {
    route: "/debtorbeginningbalance",
    moduleKey: "debtor_beginning_balance",
    apiPath: "receivableother",
    domain: "ar",
    title: { th: "บันทึกเอกสารลูกหนี้ยกมา", en: "Debtor Beginning Balance" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    defaultDocPrefix: "AR-OB",
    hasLineItems: false,
  },
  {
    route: "/transaction/billingnote",
    moduleKey: "billing_note",
    apiPath: "billingnote",
    domain: "ar",
    title: { th: "บันทึกใบวางบิล", en: "Billing Note" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    defaultDocPrefix: "BN",
    hasLineItems: true,
  },
  {
    route: "/transaction/paid",
    moduleKey: "ar_receipt",
    apiPath: "paid",
    domain: "ar",
    title: { th: "บันทึกใบเสร็จรับเงิน/รับชำระหนี้", en: "Debtor Payment Receipt" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    defaultDocPrefix: "RCP",
    hasLineItems: true,
  },
  {
    route: "/transaction/arotherdebt",
    moduleKey: "ar_other_debt",
    apiPath: "receivableother",
    domain: "ar",
    title: { th: "บันทึกเอกสารตั้งลูกหนี้อื่นๆ", en: "Other Receivables" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    defaultDocPrefix: "AR-OTH",
    hasLineItems: false,
  },
  {
    route: "/transaction/arbaddebt",
    moduleKey: "ar_bad_debt",
    apiPath: "receivableother",
    domain: "ar",
    title: { th: "บันทึกตัดหนี้สูญ(ลูกหนี้)", en: "Write Off Receivables" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    defaultDocPrefix: "AR-BAD",
    hasLineItems: false,
  },

  // 5. เจ้าหนี้ (AP)
  {
    route: "/creditorbeginningbalance",
    moduleKey: "creditor_beginning_balance",
    apiPath: "purchase",
    domain: "ap",
    title: { th: "บันทึกเอกสารเจ้าหนี้ยกมา", en: "Creditor Beginning Balance" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    defaultDocPrefix: "AP-OB",
    hasLineItems: false,
  },
  {
    route: "/transaction/apbillingreceipt",
    moduleKey: "ap_billing_receipt",
    apiPath: "purchase",
    domain: "ap",
    title: { th: "บันทึกใบรับวางบิล", en: "Supplier Billing Receipt" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    defaultDocPrefix: "VBR",
    hasLineItems: true,
  },
  {
    route: "/transaction/paymentvoucher",
    moduleKey: "payment_voucher",
    apiPath: "pay",
    domain: "ap",
    title: { th: "ใบสำคัญจ่าย", en: "Payment Voucher" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    defaultDocPrefix: "PV",
    hasLineItems: true,
  },
  {
    route: "/transaction/pay",
    moduleKey: "ap_payment",
    apiPath: "pay",
    domain: "ap",
    title: { th: "บันทึกจ่ายชำระหนี้", en: "Creditor Settlement" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    defaultDocPrefix: "PAY",
    hasLineItems: true,
  },
  {
    route: "/transaction/apotherdebt",
    moduleKey: "ap_other_debt",
    apiPath: "purchase",
    domain: "ap",
    title: { th: "บันทึกเอกสารตั้งเจ้าหนี้อื่นๆ", en: "Other Payables" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    defaultDocPrefix: "AP-OTH",
    hasLineItems: false,
  },
  {
    route: "/transaction/apbaddebt",
    moduleKey: "ap_bad_debt",
    apiPath: "purchase",
    domain: "ap",
    title: { th: "บันทึกตัดหนี้สูญ(เจ้าหนี้)", en: "Write Off Payables" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    defaultDocPrefix: "AP-BAD",
    hasLineItems: false,
  },

  // 6. เงินสดและธนาคาร (Cash & Bank)
  {
    route: "/transaction/accounttransfer",
    moduleKey: "account_transfer",
    apiPath: "bank/banktransferrecord",
    domain: "cash-bank",
    title: { th: "บันทึกโอนเงินระหว่างธนาคาร", en: "Account Transfer" },
    counterpartyLabel: { th: "บัญชีปลายทาง", en: "Destination Account" },
    defaultDocPrefix: "TRF",
    hasLineItems: false,
  },
  {
    route: "/transaction/chequereceived",
    moduleKey: "cheque_received",
    apiPath: "chequereceive/chequedeposit",
    domain: "cash-bank",
    title: { th: "บันทึกเช็ครับ", en: "Cheques Received Register" },
    counterpartyLabel: { th: "ลูกหนี้ / ผู้ออกเช็ค", en: "Drawer / Debtor" },
    defaultDocPrefix: "CHQ-IN",
    hasLineItems: false,
  },
  {
    route: "/transaction/chequeissued",
    moduleKey: "cheque_issued",
    apiPath: "chequepayment/chequepaymentdeposit",
    domain: "cash-bank",
    title: { th: "บันทึกเช็คจ่าย", en: "Cheques Issued Register" },
    counterpartyLabel: { th: "เจ้าหนี้ / ผู้รับเงิน", en: "Payee / Creditor" },
    defaultDocPrefix: "CHQ-OUT",
    hasLineItems: false,
  },
  {
    route: "/transaction/directoradvance",
    moduleKey: "director_advance",
    apiPath: "paidadvance",
    domain: "cash-bank",
    title: { th: "บันทึกขอเบิกเงินทดรองจ่ายกรรมการ", en: "Director Advance" },
    counterpartyLabel: { th: "กรรมการผู้ทดรอง", en: "Director" },
    defaultDocPrefix: "ADV-DIR",
    hasLineItems: false,
  },
  {
    route: "/transaction/employeeadvance",
    moduleKey: "employee_advance",
    apiPath: "paidadvance",
    domain: "cash-bank",
    title: { th: "เงินทดรองจ่ายพนักงาน", en: "Employee Advance" },
    counterpartyLabel: { th: "พนักงานผู้ทดรอง", en: "Employee" },
    defaultDocPrefix: "ADV-EMP",
    hasLineItems: false,
  },
  {
    route: "/cashinginthedrawer",
    moduleKey: "pos_cash_drawer",
    apiPath: "bank/depositrecord",
    domain: "cash-bank",
    title: { th: "รับ-ส่งเงิน POS", en: "POS Cash Drawer" },
    counterpartyLabel: { th: "จุดขาย / แคชเชียร์", en: "POS / Cashier" },
    defaultDocPrefix: "POS-CSH",
    hasLineItems: false,
  },

  // 7. จัดซื้อเพิ่มเติมและต้นทุน (Procurement & Landed Cost)
  {
    route: "/transaction/rfq",
    moduleKey: "rfq",
    apiPath: "purchase-requisition",
    domain: "purchase",
    title: { th: "บันทึกใบสืบราคาสินค้ารวม", en: "Request for Quotation" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "RFQ",
    hasLineItems: true,
  },
  {
    route: "/transaction/recurringexpense",
    moduleKey: "recurring_expense",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "ค่าใช้จ่ายประจำ", en: "Recurring Expense" },
    counterpartyLabel: { th: "ผู้รับเงิน / เจ้าหนี้", en: "Payee / Vendor" },
    defaultDocPrefix: "RC-EXP",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasepartial",
    moduleKey: "purchase_partial",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "บันทึกรับสินค้าจากการซื้อ", en: "Partial Goods Receipt" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "GR-PART",
    hasLineItems: true,
  },
  {
    route: "/transaction/accrualreceive",
    moduleKey: "accrual_receive",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "บันทึกตั้งหนี้จากการซื้อ", en: "Accrued Payables from Receipt" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "AP-ACC",
    hasLineItems: true,
  },
  {
    route: "/transaction/landedcost",
    moduleKey: "landed_cost",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "บันทึกต้นทุนแฝง (Weight cost)", en: "Landed Cost Allocation" },
    counterpartyLabel: { th: "ใบรับสินค้า / เจ้าหนี้", en: "Goods Receipt / Vendor" },
    defaultDocPrefix: "LC",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasereceiptadjustment",
    moduleKey: "purchase_receipt_adjustment",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "ปรับปรุงใบรับสินค้า", en: "Purchase Receipt Adjustment" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "GR-ADJ",
    hasLineItems: true,
  },
  {
    route: "/transaction/advancepayment",
    moduleKey: "advance_payment",
    apiPath: "paidadvance",
    domain: "purchase",
    title: { th: "บันทึกใบจ่ายเงินล่วงหน้า", en: "Advance Payment" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "ADV-PAY",
    hasLineItems: false,
  },
  {
    route: "/transaction/advancepaymentrefund",
    moduleKey: "advance_payment_refund",
    apiPath: "paidadvance",
    domain: "purchase",
    title: { th: "บันทึกรับคืนเงินจ่ายล่วงหน้า", en: "Advance Payment Refund" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "ADV-REF",
    hasLineItems: false,
  },
  {
    route: "/transaction/deposit",
    moduleKey: "vendor_deposit",
    apiPath: "paidadvance",
    domain: "purchase",
    title: { th: "บันทึกใบจ่ายเงินมัดจำ", en: "Supplier Deposit" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "DEP-OUT",
    hasLineItems: false,
  },
  {
    route: "/transaction/depositrefund",
    moduleKey: "vendor_deposit_refund",
    apiPath: "paidadvance",
    domain: "purchase",
    title: { th: "รับคืนเงินมัดจำ", en: "Supplier Deposit Refund" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "DEP-REF",
    hasLineItems: false,
  },

  // 8. ขายเพิ่มเติมและมัดจำ (Sales & Deposits)
  {
    route: "/transaction/recurringinvoice",
    moduleKey: "recurring_invoice",
    apiPath: "sale-invoice",
    domain: "sales",
    title: { th: "เอกสารขายประจำ", en: "Recurring Invoice" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "RC-INV",
    hasLineItems: true,
  },
  {
    route: "/transaction/etax",
    moduleKey: "etax_invoice",
    apiPath: "sale-invoice",
    domain: "sales",
    title: { th: "ใบกำกับภาษีอิเล็กทรอนิกส์ (e-Tax)", en: "e-Tax Invoice" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "ETAX",
    hasLineItems: true,
  },
  {
    route: "/transaction/combinedreceipt",
    moduleKey: "combined_receipt",
    apiPath: "paid",
    domain: "sales",
    title: { th: "ใบเสร็จรวม", en: "Combined Sales Receipt" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "CRCP",
    hasLineItems: true,
  },
  {
    route: "/transaction/paidadvance",
    moduleKey: "customer_paid_advance",
    apiPath: "paid",
    domain: "sales",
    title: { th: "บันทึกใบรับเงินล่วงหน้า", en: "Customer Advance Payment" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "RCV-ADV",
    hasLineItems: false,
  },
  {
    route: "/transaction/paidadvancerefund",
    moduleKey: "customer_paid_advance_refund",
    apiPath: "paid",
    domain: "sales",
    title: { th: "บันทึกคืนเงินรับล่วงหน้า", en: "Customer Advance Refund" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "RFD-ADV",
    hasLineItems: false,
  },
  {
    route: "/transaction/receivedeposit",
    moduleKey: "customer_deposit",
    apiPath: "paid",
    domain: "sales",
    title: { th: "บันทึกใบรับเงินมัดจำ", en: "Customer Deposit" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "DEP-IN",
    hasLineItems: false,
  },
  {
    route: "/transaction/receivedepositrefund",
    moduleKey: "customer_deposit_refund",
    apiPath: "paid",
    domain: "sales",
    title: { th: "คืนเงินมัดจำให้ลูกค้า", en: "Customer Deposit Refund" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "DEP-OUTRFD",
    hasLineItems: false,
  },
  {
    route: "/transaction/temporaryreceipt",
    moduleKey: "temporary_receipt",
    apiPath: "paid",
    domain: "sales",
    title: { th: "บันทึกใบเสร็จชั่วคราว", en: "Temporary Receipt" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "TRCP",
    hasLineItems: true,
  },

  // 9. เจ้าหนี้เพิ่มเติม (Payables Extensions)
  {
    route: "/transaction/combinedpayment",
    moduleKey: "combined_payment",
    apiPath: "pay",
    domain: "ap",
    title: { th: "ใบรวมจ่าย", en: "Combined Payment Voucher" },
    counterpartyLabel: { th: "เจ้าหนี้ / ผู้รับเงิน", en: "Payee / Creditor" },
    defaultDocPrefix: "CPAY",
    hasLineItems: true,
  },

  // 10. สต็อกเพิ่มเติมและตรวจนับ (Inventory Operations & Requests)
  {
    route: "/transaction/stockcount",
    moduleKey: "stock_count",
    apiPath: "stock-adjustment",
    domain: "inventory",
    title: { th: "บันทึกผลการตรวจนับสินค้า", en: "Physical Stock Count" },
    counterpartyLabel: { th: "คลังที่ตรวจนับ", en: "Count Warehouse" },
    defaultDocPrefix: "CNT",
    hasLineItems: true,
  },
  {
    route: "/transaction/costadjustment",
    moduleKey: "cost_adjustment",
    apiPath: "stock-adjustment",
    domain: "inventory",
    title: { th: "ปรับปรุงต้นทุนสินค้า", en: "Cost Adjustment" },
    counterpartyLabel: { th: "คลัง / กลุ่มสินค้า", en: "Warehouse / Product Group" },
    defaultDocPrefix: "CADJ",
    hasLineItems: true,
  },
  {
    route: "/inventory/issue-request",
    moduleKey: "inventory_issue_request",
    apiPath: "stock-prickup-product",
    domain: "inventory",
    title: { th: "บันทึกใบขอเบิกใช้สินค้า,วัตถุดิบ", en: "Material Requisition" },
    counterpartyLabel: { th: "แผนกผู้ขอเบิก", en: "Requesting Dept" },
    defaultDocPrefix: "MRQ",
    hasLineItems: true,
  },
  {
    route: "/inventory/transfer-request",
    moduleKey: "inventory_transfer_request",
    apiPath: "stock-transfer",
    domain: "inventory",
    title: { th: "บันทึกใบขอโอนสินค้า", en: "Transfer Requisition" },
    counterpartyLabel: { th: "คลังปลายทาง", en: "Destination Warehouse" },
    defaultDocPrefix: "TRQ",
    hasLineItems: true,
  },
  {
    route: "/inventory/goods-inspection",
    moduleKey: "inventory_goods_inspection",
    apiPath: "stock-receive-product",
    domain: "inventory",
    title: { th: "ใบตรวจรับสินค้า (QC/QA)", en: "Goods Inspection" },
    counterpartyLabel: { th: "เจ้าหนี้ / ผู้ส่งมอบ", en: "Vendor / Carrier" },
    defaultDocPrefix: "QC",
    hasLineItems: true,
  },
  {
    route: "/inventory/count-sheet",
    moduleKey: "inventory_count_sheet",
    apiPath: "stock-adjustment",
    domain: "inventory",
    title: { th: "บันทึกเอกสารเพื่อตรวจนับสินค้า", en: "Count Sheet" },
    counterpartyLabel: { th: "คลัง / ที่เก็บ", en: "Warehouse / Zone" },
    defaultDocPrefix: "CSH",
    hasLineItems: true,
  },

  // 11. การเงิน เงินสด ธนาคาร บัตรเครดิต และเช็ค (Cash, Banking & Cheques)
  {
    route: "/pettycashscreen",
    moduleKey: "petty_cash",
    apiPath: "bank/banktransferrecord",
    domain: "cash-bank",
    title: { th: "เงินสดย่อย", en: "Petty Cash Management" },
    counterpartyLabel: { th: "ผู้ถือเงินสดย่อย", en: "Petty Cash Custodian" },
    defaultDocPrefix: "PC",
    hasLineItems: true,
  },
  {
    route: "/transaction/creditcardexpense",
    moduleKey: "credit_card_expense",
    apiPath: "purchase",
    domain: "cash-bank",
    title: { th: "บัตรเครดิตกิจการ", en: "Corporate Credit Card Expense" },
    counterpartyLabel: { th: "บัตร / ร้านค้า", en: "Card / Merchant" },
    defaultDocPrefix: "CC-EXP",
    hasLineItems: true,
  },
  {
    route: "/transaction/otherincomereceipt",
    moduleKey: "other_income_receipt",
    apiPath: "paid",
    domain: "cash-bank",
    title: { th: "บันทึกรับเงินอื่นๆ", en: "Other Income Receipt" },
    counterpartyLabel: { th: "ผู้ชำระเงิน", en: "Payer" },
    defaultDocPrefix: "OTH-INC",
    hasLineItems: true,
  },
  {
    route: "/banking/petty-cash-limit",
    moduleKey: "petty_cash_limit",
    apiPath: "bank/banktransferrecord",
    domain: "cash-bank",
    title: { th: "กำหนดวงเงินสดย่อย", en: "Petty Cash Limit Setting" },
    counterpartyLabel: { th: "กองเงินสดย่อย", en: "Petty Cash Fund" },
    defaultDocPrefix: "PC-LIM",
    hasLineItems: false,
  },
  {
    route: "/transaction/bankpaymentfile",
    moduleKey: "bank_payment_file",
    apiPath: "pay",
    domain: "cash-bank",
    title: { th: "ไฟล์โอนเงินจ่ายผ่านธนาคาร", en: "Bank Payment File" },
    counterpartyLabel: { th: "ธนาคารคู่ค้า", en: "Bank" },
    defaultDocPrefix: "B-PAY",
    hasLineItems: true,
  },
  {
    route: "/slipmoneyin",
    moduleKey: "slip_money_in",
    apiPath: "paid",
    domain: "cash-bank",
    title: { th: "รูปสลิปเงินเข้า", en: "Incoming Payment Slip Registry" },
    counterpartyLabel: { th: "ผู้โอนเงิน", en: "Transferor" },
    defaultDocPrefix: "SLP-IN",
    hasLineItems: false,
  },
  {
    route: "/slipmoneyout",
    moduleKey: "slip_money_out",
    apiPath: "pay",
    domain: "cash-bank",
    title: { th: "รูปสลิปเงินออก", en: "Outgoing Payment Slip Registry" },
    counterpartyLabel: { th: "ผู้รับเงิน", en: "Transferee" },
    defaultDocPrefix: "SLP-OUT",
    hasLineItems: false,
  },
  {
    route: "/banking/statements",
    moduleKey: "bank_statements",
    apiPath: "bank/banktransferrecord",
    domain: "cash-bank",
    title: { th: "รายการเดินบัญชีธนาคาร", en: "Bank Statements" },
    counterpartyLabel: { th: "ธนาคาร / เลขที่บัญชี", en: "Bank / Account" },
    defaultDocPrefix: "STM",
    hasLineItems: false,
  },
  {
    route: "/banking/reconciliation",
    moduleKey: "bank_reconciliation",
    apiPath: "bank/banktransferrecord",
    domain: "cash-bank",
    title: { th: "กระทบยอดเงินฝากธนาคาร", en: "Bank Reconciliation" },
    counterpartyLabel: { th: "บัญชีธนาคาร", en: "Bank Account" },
    defaultDocPrefix: "REC",
    hasLineItems: false,
  },
  {
    route: "/banking/cash-deposit",
    moduleKey: "cash_deposit",
    apiPath: "bank/depositrecord",
    domain: "cash-bank",
    title: { th: "บันทึกนำฝากเงินสด", en: "Cash Deposit" },
    counterpartyLabel: { th: "บัญชีธนาคารปลายทาง", en: "Target Bank Account" },
    defaultDocPrefix: "DEP",
    hasLineItems: false,
  },
  {
    route: "/banking/cash-withdrawal",
    moduleKey: "cash_withdrawal",
    apiPath: "bank/depositrecord",
    domain: "cash-bank",
    title: { th: "บันทึกถอนเงินสด", en: "Cash Withdrawal" },
    counterpartyLabel: { th: "บัญชีธนาคารถอน", en: "Source Bank Account" },
    defaultDocPrefix: "WTH",
    hasLineItems: false,
  },
  {
    route: "/banking/charges",
    moduleKey: "bank_charges",
    apiPath: "bank/banktransferrecord",
    domain: "cash-bank",
    title: { th: "บันทึกค่าใช้จ่ายธนาคาร", en: "Bank Charges" },
    counterpartyLabel: { th: "ธนาคาร", en: "Bank" },
    defaultDocPrefix: "BNK-FEE",
    hasLineItems: false,
  },
  {
    route: "/banking/income",
    moduleKey: "bank_income",
    apiPath: "bank/banktransferrecord",
    domain: "cash-bank",
    title: { th: "บันทึกรายได้จากธนาคาร", en: "Bank Income" },
    counterpartyLabel: { th: "ธนาคาร", en: "Bank" },
    defaultDocPrefix: "BNK-INC",
    hasLineItems: false,
  },
  {
    route: "/banking/cheques/deposit",
    moduleKey: "cheque_deposit",
    apiPath: "chequereceive/chequedeposit",
    domain: "cash-bank",
    title: { th: "บันทึกนำฝากเช็ครับ", en: "Cheque Deposit" },
    counterpartyLabel: { th: "ธนาคารที่นำฝาก", en: "Deposit Bank" },
    defaultDocPrefix: "CHQ-DEP",
    hasLineItems: false,
  },
  {
    route: "/banking/cheques/received-clear",
    moduleKey: "cheque_received_clear",
    apiPath: "chequereceive/chequedeposit",
    domain: "cash-bank",
    title: { th: "บันทึกเช็ครับผ่าน", en: "Cleared Cheques Received" },
    counterpartyLabel: { th: "ธนาคาร / บัญชี", en: "Bank / Account" },
    defaultDocPrefix: "CHQ-CLR",
    hasLineItems: false,
  },
  {
    route: "/banking/cheques/received-return",
    moduleKey: "cheque_received_return",
    apiPath: "chequereceive/chequedeposit",
    domain: "cash-bank",
    title: { th: "บันทึกเช็ครับคืน", en: "Returned Cheques Received" },
    counterpartyLabel: { th: "ลูกหนี้ / ธนาคาร", en: "Debtor / Bank" },
    defaultDocPrefix: "CHQ-RET",
    hasLineItems: false,
  },
  {
    route: "/banking/cheques/redeposit",
    moduleKey: "cheque_redeposit",
    apiPath: "chequereceive/chequedeposit",
    domain: "cash-bank",
    title: { th: "บันทึกนำเช็คเข้าใหม่", en: "Cheque Re-deposit" },
    counterpartyLabel: { th: "ธนาคาร", en: "Bank" },
    defaultDocPrefix: "CHQ-REDEP",
    hasLineItems: false,
  },
  {
    route: "/banking/cheques/received-cancel",
    moduleKey: "cheque_received_cancel",
    apiPath: "chequereceive/chequedeposit",
    domain: "cash-bank",
    title: { th: "บันทึกยกเลิกเช็ครับ", en: "Cancel Cheque Received" },
    counterpartyLabel: { th: "ผู้ออกเช็ค", en: "Drawer" },
    defaultDocPrefix: "CHQ-CAN",
    hasLineItems: false,
  },
  {
    route: "/banking/cheques/discount",
    moduleKey: "cheque_discount",
    apiPath: "chequereceive/chequedeposit",
    domain: "cash-bank",
    title: { th: "บันทึกขายลดเช็ครับ", en: "Discounted Cheques" },
    counterpartyLabel: { th: "สถาบันการเงิน / ธนาคาร", en: "Financial Institution" },
    defaultDocPrefix: "CHQ-DSC",
    hasLineItems: false,
  },
  {
    route: "/banking/cheques/issued-clear",
    moduleKey: "cheque_issued_clear",
    apiPath: "chequepayment/chequepaymentdeposit",
    domain: "cash-bank",
    title: { th: "บันทึกเช็คจ่ายผ่าน", en: "Cleared Cheques Issued" },
    counterpartyLabel: { th: "ธนาคารตัดบัญชี", en: "Drawn Bank" },
    defaultDocPrefix: "CHQ-OUTCLR",
    hasLineItems: false,
  },
  {
    route: "/banking/cheques/issued-cancel",
    moduleKey: "cheque_issued_cancel",
    apiPath: "chequepayment/chequepaymentdeposit",
    domain: "cash-bank",
    title: { th: "บันทึกยกเลิกเช็คจ่าย", en: "Cancel Cheque Issued" },
    counterpartyLabel: { th: "เจ้าหนี้ / ผู้รับเงิน", en: "Payee" },
    defaultDocPrefix: "CHQ-OUTCAN",
    hasLineItems: false,
  },
  {
    route: "/banking/cards/receipts",
    moduleKey: "card_receipts",
    apiPath: "paid",
    domain: "cash-bank",
    title: { th: "บันทึกบัตรเครดิต", en: "Credit Card Receipts Register" },
    counterpartyLabel: { th: "เครื่อง EDC / บัตร", en: "EDC Terminal / Card" },
    defaultDocPrefix: "CARD-RCV",
    hasLineItems: false,
  },
  {
    route: "/banking/cards/settlement",
    moduleKey: "card_settlement",
    apiPath: "bank/depositrecord",
    domain: "cash-bank",
    title: { th: "บันทึกขึ้นเงินบัตรเครดิต", en: "Credit Card Settlement" },
    counterpartyLabel: { th: "ธนาคารรับชำระ", en: "Acquirer Bank" },
    defaultDocPrefix: "CARD-SET",
    hasLineItems: false,
  },
  {
    route: "/banking/cards/cancellation",
    moduleKey: "card_cancellation",
    apiPath: "paid",
    domain: "cash-bank",
    title: { th: "บันทึกยกเลิกบัตรเครดิต", en: "Credit Card Receipt Void" },
    counterpartyLabel: { th: "ลูกค้า / บัตร", en: "Customer / Card" },
    defaultDocPrefix: "CARD-VOID",
    hasLineItems: false,
  },
  {
    route: "/banking/contacts",
    moduleKey: "bank_contacts",
    apiPath: "bank/banktransferrecord",
    domain: "cash-bank",
    title: { th: "บันทึกรายละเอียดผู้ติดต่อธนาคาร", en: "Bank Contacts Register" },
    counterpartyLabel: { th: "สาขาธนาคาร / เจ้าหน้าที่", en: "Branch / Officer" },
    defaultDocPrefix: "BNK-CON",
    hasLineItems: false,
  },

  // 12. ภาษีและทะเบียนภาษี (Tax Transactions)
  {
    route: "/transaction/purchasetaxinvoice",
    moduleKey: "purchase_tax_invoice",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "ทะเบียนใบกำกับภาษีซื้อ", en: "Purchase Tax Invoice Register" },
    counterpartyLabel: { th: "เจ้าหนี้ / ผู้ออกใบกำกับ", en: "Vendor / Tax Issuer" },
    defaultDocPrefix: "TX-IN",
    hasLineItems: true,
  },
  {
    route: "/transaction/withholdingtax",
    moduleKey: "withholding_tax_record",
    apiPath: "pay",
    domain: "purchase",
    title: { th: "บันทึกภาษีหัก ณ ที่จ่าย (50 ทวิ)", en: "Withholding Tax Record" },
    counterpartyLabel: { th: "ผู้ถูกหักภาษี", en: "Withholding Payee" },
    defaultDocPrefix: "WHT",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasevatadjustment",
    moduleKey: "purchase_vat_adjustment",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "ปรับปรุงภาษีซื้อ", en: "Purchase VAT Adjustment" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor" },
    defaultDocPrefix: "VAT-ADJ",
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

export async function fetchErpTransactions(
  config: ErpModuleConfig,
  params: { q?: string; offset?: number; limit?: number } = {},
): Promise<{ items: ErpTransactionDoc[]; total: number; error?: string }> {
  const query = new URLSearchParams();
  if (params.q) query.set("q", params.q);
  if (params.offset !== undefined) query.set("offset", String(params.offset));
  if (params.limit !== undefined) query.set("limit", String(params.limit));

  try {
    const res = await authFetch(`/api/erp-transaction/${config.apiPath}/list?${query.toString()}`);
    if (!res.ok) {
      const error = res.status === 401 || res.status === 403 ? "unauthorized" : "load_failed";
      return { items: [], total: 0, error };
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

export async function saveErpTransaction(
  config: ErpModuleConfig,
  doc: Partial<ErpTransactionDoc>,
  isEdit: boolean,
): Promise<{ success: boolean; data?: ErpTransactionDoc; message?: string }> {
  try {
    const method = isEdit ? "PUT" : "POST";
    const path = isEdit && doc.id ? `${config.apiPath}/${encodeURIComponent(doc.id)}` : config.apiPath;
    const res = await authFetch(`/api/erp-transaction/${path}`, {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(doc),
    });
    if (!res.ok) {
      return { success: false, message: "save_failed" };
    }
    const data = await res.json();
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
): Promise<{ success: boolean; message?: string }> {
  try {
    const res = await authFetch(`/api/erp-transaction/${config.apiPath}/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
    if (!res.ok) {
      return { success: false, message: "delete_failed" };
    }
    return { success: true, message: "delete_success" };
  } catch {
    return { success: false, message: "connection_error" };
  }
}
