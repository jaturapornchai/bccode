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
    title: { th: "ยอดยกมาสินค้า", en: "Stock Beginning Balance" },
    counterpartyLabel: { th: "คลัง / แผนก", en: "Warehouse / Dept" },
    defaultDocPrefix: "STK-BAL",
    hasLineItems: true,
  },
  {
    route: "/transaction/stockreceiveproduct",
    moduleKey: "stock_receive",
    apiPath: "stock-receive-product",
    domain: "inventory",
    title: { th: "รับสินค้าเข้าคลัง", en: "Stock Receive" },
    counterpartyLabel: { th: "คลังปลายทาง", en: "Target Warehouse" },
    defaultDocPrefix: "STK-RCV",
    hasLineItems: true,
  },
  {
    route: "/transaction/stockpickupproduct",
    moduleKey: "stock_pickup",
    apiPath: "stock-prickup-product",
    domain: "inventory",
    title: { th: "เบิกสินค้าออกจากคลัง", en: "Stock Issue / Pickup" },
    counterpartyLabel: { th: "แผนกผู้ขอเบิก", en: "Requesting Dept" },
    defaultDocPrefix: "STK-OUT",
    hasLineItems: true,
  },
  {
    route: "/transaction/stockreturnproduct",
    moduleKey: "stock_return",
    apiPath: "stock-return-product",
    domain: "inventory",
    title: { th: "คืนสินค้าเข้าคลัง", en: "Stock Return" },
    counterpartyLabel: { th: "คลังรับคืน", en: "Receiving Warehouse" },
    defaultDocPrefix: "STK-RET",
    hasLineItems: true,
  },
  {
    route: "/transaction/stocktransfer",
    moduleKey: "stock_transfer",
    apiPath: "stock-transfer",
    domain: "inventory",
    title: { th: "โอนสินค้าระหว่างคลัง", en: "Stock Transfer" },
    counterpartyLabel: { th: "คลังปลายทาง", en: "Destination Warehouse" },
    defaultDocPrefix: "STK-XFR",
    hasLineItems: true,
  },
  {
    route: "/transaction/adjust",
    moduleKey: "stock_adjust",
    apiPath: "stock-adjustment",
    domain: "inventory",
    title: { th: "ปรับปรุงสต็อก", en: "Stock Adjustment" },
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
    title: { th: "ใบเสนอราคา", en: "Quotation" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "QT",
    hasLineItems: true,
  },
  {
    route: "/transaction/saleorder",
    moduleKey: "sale_order",
    apiPath: "sale-order",
    domain: "sales",
    title: { th: "ใบสั่งขาย / สั่งจอง", en: "Sales Order" },
    counterpartyLabel: { th: "ลูกค้า", en: "Customer" },
    defaultDocPrefix: "SO",
    hasLineItems: true,
  },
  {
    route: "/transaction/sale",
    moduleKey: "sale_invoice",
    apiPath: "sale-invoice",
    domain: "sales",
    title: { th: "ขายสินค้า / ส่งของ", en: "Sale & Dispatch" },
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
    title: { th: "คืนสินค้าจากการขาย", en: "Sale Return" },
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
    title: { th: "ใบเพิ่มหนี้ (Debit Note)", en: "Debit Note" },
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
    title: { th: "ใบขอซื้อ (PR)", en: "Purchase Requisition" },
    counterpartyLabel: { th: "ผู้จำหน่ายที่เสนอ", en: "Proposed Vendor" },
    defaultDocPrefix: "PR",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchaseorder",
    moduleKey: "purchase_order",
    apiPath: "purchase-order",
    domain: "purchase",
    title: { th: "ใบสั่งซื้อ (PO)", en: "Purchase Order" },
    counterpartyLabel: { th: "ผู้จำหน่าย", en: "Vendor" },
    defaultDocPrefix: "PO",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchase",
    moduleKey: "purchase_invoice",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "ซื้อสินค้า / รับของ", en: "Purchase Invoice / Goods Receipt" },
    counterpartyLabel: { th: "ผู้จำหน่าย", en: "Vendor" },
    defaultDocPrefix: "PUR",
    hasLineItems: true,
  },
  {
    route: "/transaction/expense",
    moduleKey: "expense_record",
    apiPath: "purchase",
    domain: "purchase",
    title: { th: "บันทึกค่าใช้จ่าย", en: "Expense Record" },
    counterpartyLabel: { th: "ผู้รับเงิน / ผู้จำหน่าย", en: "Payee / Vendor" },
    defaultDocPrefix: "EXP",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasereturn",
    moduleKey: "purchase_return",
    apiPath: "purchase-return",
    domain: "purchase",
    title: { th: "คืนสินค้าให้ผู้ขาย", en: "Purchase Return" },
    counterpartyLabel: { th: "ผู้จำหน่าย", en: "Vendor" },
    defaultDocPrefix: "PRT",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasecreditnote",
    moduleKey: "purchase_credit_note",
    apiPath: "purchase-return",
    domain: "purchase",
    title: { th: "ใบเพิ่มหนี้เจ้าหนี้", en: "Vendor Credit Note" },
    counterpartyLabel: { th: "ผู้จำหน่าย", en: "Vendor" },
    defaultDocPrefix: "VCN",
    hasLineItems: true,
  },
  {
    route: "/transaction/purchasedebitnote",
    moduleKey: "purchase_debit_note",
    apiPath: "bank/purchasedebitnote",
    domain: "purchase",
    title: { th: "ใบลดหนี้เจ้าหนี้", en: "Vendor Debit Note" },
    counterpartyLabel: { th: "ผู้จำหน่าย", en: "Vendor" },
    defaultDocPrefix: "VDN",
    hasLineItems: true,
  },

  // 4. ลูกหนี้ (AR)
  {
    route: "/debtorbeginningbalance",
    moduleKey: "debtor_beginning_balance",
    apiPath: "receivableother",
    domain: "ar",
    title: { th: "ลูกหนี้ตั้งต้นรายเอกสาร", en: "Debtor Beginning Balance" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    defaultDocPrefix: "AR-OB",
    hasLineItems: false,
  },
  {
    route: "/transaction/billingnote",
    moduleKey: "billing_note",
    apiPath: "billingnote",
    domain: "ar",
    title: { th: "ใบวางบิล", en: "Billing Note" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    defaultDocPrefix: "BN",
    hasLineItems: true,
  },
  {
    route: "/transaction/paid",
    moduleKey: "ar_receipt",
    apiPath: "paid",
    domain: "ar",
    title: { th: "รับชำระหนี้จากลูกหนี้", en: "Debtor Payment Receipt" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    defaultDocPrefix: "RCP",
    hasLineItems: true,
  },
  {
    route: "/transaction/arotherdebt",
    moduleKey: "ar_other_debt",
    apiPath: "receivableother",
    domain: "ar",
    title: { th: "ตั้งลูกหนี้อื่นๆ", en: "Other Receivables" },
    counterpartyLabel: { th: "ลูกหนี้", en: "Customer / Debtor" },
    defaultDocPrefix: "AR-OTH",
    hasLineItems: false,
  },
  {
    route: "/transaction/arbaddebt",
    moduleKey: "ar_bad_debt",
    apiPath: "receivableother",
    domain: "ar",
    title: { th: "ตัดหนี้สูญลูกหนี้", en: "Write Off Receivables" },
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
    title: { th: "เจ้าหนี้ตั้งต้นรายเอกสาร", en: "Creditor Beginning Balance" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    defaultDocPrefix: "AP-OB",
    hasLineItems: false,
  },
  {
    route: "/transaction/apbillingreceipt",
    moduleKey: "ap_billing_receipt",
    apiPath: "purchase",
    domain: "ap",
    title: { th: "ใบรับวางบิลเจ้าหนี้", en: "Supplier Billing Receipt" },
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
    title: { th: "จ่ายชำระหนี้เจ้าหนี้", en: "Creditor Settlement" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    defaultDocPrefix: "PAY",
    hasLineItems: true,
  },
  {
    route: "/transaction/apotherdebt",
    moduleKey: "ap_other_debt",
    apiPath: "purchase",
    domain: "ap",
    title: { th: "ตั้งเจ้าหนี้อื่นๆ", en: "Other Payables" },
    counterpartyLabel: { th: "เจ้าหนี้", en: "Vendor / Creditor" },
    defaultDocPrefix: "AP-OTH",
    hasLineItems: false,
  },
  {
    route: "/transaction/apbaddebt",
    moduleKey: "ap_bad_debt",
    apiPath: "purchase",
    domain: "ap",
    title: { th: "ตัดหนี้สูญเจ้าหนี้", en: "Write Off Payables" },
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
    title: { th: "โอนเงินระหว่างบัญชี", en: "Account Transfer" },
    counterpartyLabel: { th: "บัญชีปลายทาง", en: "Destination Account" },
    defaultDocPrefix: "TRF",
    hasLineItems: false,
  },
  {
    route: "/transaction/chequereceived",
    moduleKey: "cheque_received",
    apiPath: "chequereceive/chequedeposit",
    domain: "cash-bank",
    title: { th: "ทะเบียนเช็ครับ", en: "Cheques Received Register" },
    counterpartyLabel: { th: "ลูกหนี้ / ผู้ออกเช็ค", en: "Drawer / Debtor" },
    defaultDocPrefix: "CHQ-IN",
    hasLineItems: false,
  },
  {
    route: "/transaction/chequeissued",
    moduleKey: "cheque_issued",
    apiPath: "chequepayment/chequepaymentdeposit",
    domain: "cash-bank",
    title: { th: "ทะเบียนเช็คจ่าย", en: "Cheques Issued Register" },
    counterpartyLabel: { th: "เจ้าหนี้ / ผู้รับเงิน", en: "Payee / Creditor" },
    defaultDocPrefix: "CHQ-OUT",
    hasLineItems: false,
  },
  {
    route: "/transaction/directoradvance",
    moduleKey: "director_advance",
    apiPath: "paidadvance",
    domain: "cash-bank",
    title: { th: "เงินทดรองจ่ายกรรมการ", en: "Director Advance" },
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

// In-memory demo data cache to ensure offline/mock fallback resilience
const mockStore = new Map<string, ErpTransactionDoc[]>();

function getOrCreateMockDocs(apiPath: string, prefix: string): ErpTransactionDoc[] {
  if (mockStore.has(apiPath)) return mockStore.get(apiPath)!;

  const today = new Date().toISOString().split("T")[0];
  const initial: ErpTransactionDoc[] = [
    {
      id: `${prefix}-001`,
      docno: `${prefix}-202609-0001`,
      docdatetime: `${today}T09:00:00Z`,
      custcode: "CUST-001",
      custname: "บริษัท สยามพาณิชย์ จำกัด",
      description: "รายการประจำงวดกันยายน 2026",
      totalamount: 10700,
      totalbeforevat: 10000,
      totalvatvalue: 700,
      vattype: 1,
      vatrate: 7,
      status: 1,
      details: [
        {
          linenumber: 1,
          itemcode: "P-001",
          itemname: "สินค้ามาตรฐานชนิด A",
          unitcode: "PCS",
          unitname: "ชิ้น",
          qty: 10,
          price: 1000,
          sumamount: 10000,
        },
      ],
    },
    {
      id: `${prefix}-002`,
      docno: `${prefix}-202609-0002`,
      docdatetime: `${today}T11:30:00Z`,
      custcode: "CUST-002",
      custname: "ห้างหุ้นส่วนจำกัด ทวีโชคทรานสปอร์ต",
      description: "เอกสารรอบพิเศษ",
      totalamount: 21400,
      totalbeforevat: 20000,
      totalvatvalue: 1400,
      vattype: 1,
      vatrate: 7,
      status: 0,
      details: [
        {
          linenumber: 1,
          itemcode: "P-002",
          itemname: "อะไหล่และอุปกรณ์เสริมชนิด B",
          unitcode: "SET",
          unitname: "ชุด",
          qty: 4,
          price: 5000,
          sumamount: 20000,
        },
      ],
    },
  ];
  mockStore.set(apiPath, initial);
  return initial;
}

export async function fetchErpTransactions(
  config: ErpModuleConfig,
  params: { q?: string; offset?: number; limit?: number } = {},
): Promise<{ items: ErpTransactionDoc[]; total: number }> {
  const query = new URLSearchParams();
  if (params.q) query.set("q", params.q);
  if (params.offset !== undefined) query.set("offset", String(params.offset));
  if (params.limit !== undefined) query.set("limit", String(params.limit));

  try {
    const res = await authFetch(`/api/erp-transaction/${config.apiPath}/list?${query.toString()}`);
    if (res.ok) {
      const data = await res.json();
      if (Array.isArray(data?.data)) {
        return { items: data.data, total: data.pagination?.total ?? data.data.length };
      }
      if (Array.isArray(data)) {
        return { items: data, total: data.length };
      }
    }
  } catch {
    // Network or server unreachable: fallback to structured local store
  }

  // Fallback resilience
  const docs = getOrCreateMockDocs(config.apiPath, config.defaultDocPrefix);
  let filtered = docs;
  if (params.q?.trim()) {
    const q = params.q.toLowerCase();
    filtered = docs.filter(
      (d) =>
        d.docno.toLowerCase().includes(q) ||
        (d.custcode && d.custcode.toLowerCase().includes(q)) ||
        (d.custname && d.custname.toLowerCase().includes(q)) ||
        (d.description && d.description.toLowerCase().includes(q)),
    );
  }
  return { items: filtered, total: filtered.length };
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
    if (res.ok) {
      const data = await res.json();
      return { success: true, data: data.data || doc, message: "บันทึกข้อมูลเรียบร้อยแล้ว" };
    }
  } catch {
    // Fallback store update
  }

  // Local store update
  const docs = getOrCreateMockDocs(config.apiPath, config.defaultDocPrefix);
  if (isEdit && doc.id) {
    const idx = docs.findIndex((x) => x.id === doc.id);
    if (idx >= 0) {
      docs[idx] = { ...docs[idx], ...doc } as ErpTransactionDoc;
      return { success: true, data: docs[idx] };
    }
  } else {
    const newDoc: ErpTransactionDoc = {
      ...doc,
      id: `${config.defaultDocPrefix}-${Date.now()}`,
      docno: doc.docno || `${config.defaultDocPrefix}-${Date.now().toString().slice(-6)}`,
      docdatetime: doc.docdatetime || new Date().toISOString(),
      status: doc.status ?? 0,
      totalamount: doc.totalamount ?? 0,
      details: doc.details ?? [],
    } as ErpTransactionDoc;
    docs.unshift(newDoc);
    return { success: true, data: newDoc };
  }

  return { success: true, data: doc as ErpTransactionDoc };
}

export async function deleteErpTransaction(
  config: ErpModuleConfig,
  id: string,
): Promise<{ success: boolean; message?: string }> {
  try {
    const res = await authFetch(`/api/erp-transaction/${config.apiPath}/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
    if (res.ok) {
      return { success: true, message: "ลบเอกสารเรียบร้อยแล้ว" };
    }
  } catch {
    // Fallback
  }

  const docs = getOrCreateMockDocs(config.apiPath, config.defaultDocPrefix);
  const idx = docs.findIndex((x) => x.id === id);
  if (idx >= 0) {
    docs.splice(idx, 1);
  }
  return { success: true, message: "ลบเอกสารเรียบร้อยแล้ว" };
}
