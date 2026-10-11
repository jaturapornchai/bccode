/**
 * คีย์ประเภทเอกสารธุรกรรมที่ route /transaction/:module ของ backend รับ — ตรงกับทะเบียนเดียว
 * `backend/internal/doctype/doctype.go` (63 ประเภท ลบ 10 ประเภทเช็คที่ย้ายไปโมดูล /cheque/v1 = 53) ห้ามเพิ่ม/ลบที่นี่ก่อนแก้ทะเบียน backend
 * (มี test เทียบสองฝั่งใน erp-transaction.test.ts; เช็คถูกปฏิเสธที่ /transaction/:module ด้วย erp_err_cheque_module_moved)
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
  "creditcard",
  "creditcardsettle",
  "creditcardcancel",
];

/**
 * เอกสารหนี้เจ้าหนี้ที่มีช่องวันเครดิต/วันครบกำหนดบนหัวเอกสาร (Champ CreditDay/DueDate) — ตรงกับ dueDateDocs ใน
 * backend/internal/transaction/httpapi/due_date.go (มีเทสต์เทียบ)
 */
export const ERP_DUE_DATE_MODULES: readonly string[] = ["purchase", "creditorbeginningbalance", "apotherdebt", "apdebitnote"];

/**
 * ประเภทเอกสารที่ backend ทำอะไรต่อให้ตอนบันทึก — ลงบัญชี (10 ประเภท) หรือเคลื่อนไหวสต็อก (รับเข้า/จ่ายออกตามทะเบียน)
 * ประเภทอื่นทุกชนิด "เก็บเอกสารอย่างเดียว"; ตรงกับ doctype.go (postingFlags + Stock = StockIn/StockOut) — มีเทสต์เทียบ
 */
export const ERP_AUTO_POSTING_MODULES: readonly string[] = [
  "purchase",
  "purchase-return",
  "expense",
  "paid",
  "pay",
  "sale-invoice",
  "sale-invoice-return",
  "apdebitnote",
  "apcreditnote",
  "apotherdebt",
  "goodsreceive",
  "stock-balance",
  "stock-receive-product",
  "stock-pickup-product",
  "stock-return-product",
];
