/**
 * คีย์ประเภทเอกสารธุรกรรมที่ route /transaction/:module ของ backend รับ — ตรงกับทะเบียนเดียว
 * `backend/internal/doctype/doctype.go` (63 ประเภท) ห้ามเพิ่ม/ลบที่นี่ก่อนแก้ทะเบียน backend
 * (มี test เทียบสองฝั่งใน erp-transaction.test.ts)
 */
export const ERP_TRANSACTION_MODULES: readonly string[] = [
  // สินค้าคงคลัง
  "stock-balance",
  "stock-receive-product",
  "stock-pickup-product",
  "stock-return-product",
  "stock-transfer",
  "stock-adjustment",
  "stock-issue-request",
  "stock-transfer-request",
  "goods-inspection",
  "stock-count",
  // ขาย ลูกหนี้
  "quotation",
  "sale-order",
  "sale-invoice",
  "sale-invoice-return",
  "ardebitnote",
  "debtorbeginningbalance",
  "receivableother",
  "billingnote",
  "paid",
  "receivedeposit",
  "receiveadvancerefund",
  "chequereceived",
  // ซื้อ เจ้าหนี้
  "purchase-requisition",
  "purchase-order",
  "purchase",
  "purchase-return",
  "creditorbeginningbalance",
  "expense",
  "apbillingreceipt",
  "pay",
  "deposit",
  "advancepayment",
  "advancepaymentrefund",
  // ประเภทที่เพิ่มใหม่ (transflag 400-499)
  "rfq",
  "goodsreceive",
  "invoicereceive",
  "landedcost",
  "apdebitnote",
  "apcreditnote",
  "apotherdebt",
  "apbaddebt",
  "receiveadvance",
  "temporaryreceipt",
  "arbaddebt",
  "otherincome",
  "employeeadvance",
  "cashdeposit",
  "cashwithdrawal",
  "bankexpense",
  "bankincome",
  "banktransfer",
  "chequeissued",
  "chequedeposit",
  "chequereceivedclear",
  "chequereceivedreturn",
  "chequeredeposit",
  "chequereceivedcancel",
  "chequediscount",
  "chequeissuedclear",
  "chequeissuedcancel",
  "creditcard",
  "creditcardsettle",
  "creditcardcancel",
];

/**
 * ประเภทเอกสารที่ backend ลงบัญชี (หรือตัดสต็อก) อัตโนมัติตอนบันทึก — ประเภทอื่นทุกชนิด "เก็บเอกสารอย่างเดียว"
 * ตรงกับค่าคงที่ที่ตั้งชื่อไว้ใน doctype.go (Purchase, SaleInvoice, SaleInvoiceReturn, PurchaseReturn, Paid, Pay, Expense, StockPickup)
 */
export const ERP_AUTO_POSTING_MODULES: readonly string[] = [
  "purchase",
  "sale-invoice",
  "sale-invoice-return",
  "purchase-return",
  "paid",
  "pay",
  "expense",
  "stock-pickup-product",
];
