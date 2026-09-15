// SME Operations, Approvals, BOM Kits, Serial Registry & Data Import Engine

import { authFetch } from "@/lib/client-auth-session";

export type OperationsCategory =
  | "approval"
  | "reservation"
  | "procurement"
  | "bom"
  | "pricing"
  | "import";

export interface OperationsConfig {
  route: string;
  code: string;
  category: OperationsCategory;
  title: { th: string; en: string };
  description: { th: string; en: string };
  primaryActionLabel: { th: string; en: string };
  documentType?: string;
}

export const OPERATIONS_CONFIGS: OperationsConfig[] = [
  // --- 1. ระบบอนุมัติและยกเลิกเอกสาร (Approval & Cancellation) ---
  {
    route: "/procurement/requisition-approval",
    code: "pr_approval",
    category: "approval",
    title: { th: "อนุมัติใบเสนอซื้อสินค้า (PR Approval)", en: "Purchase Requisition Approval" },
    description: { th: "ตรวจทานและอนุมัติใบขอซื้อก่อนออกเป็นใบสั่งซื้อ (PO) ตามลำดับอำนาจอนุมัติ", en: "Review and approve purchase requisitions based on authorization hierarchy" },
    primaryActionLabel: { th: "อนุมัติรายการที่เลือก", en: "Approve Selected" },
    documentType: "PR",
  },
  {
    route: "/procurement/order-cancellation",
    code: "po_cancellation",
    category: "approval",
    title: { th: "ยกเลิกใบสั่งซื้อสินค้า (PO Cancellation)", en: "Purchase Order Cancellation" },
    description: { th: "พิจารณายกเลิกใบสั่งซื้อและปลดภาระผูกพันทางการเงิน", en: "Cancel approved purchase orders and release commitments" },
    primaryActionLabel: { th: "ยืนยันการยกเลิก PO", en: "Confirm PO Cancellation" },
    documentType: "PO",
  },
  {
    route: "/sales/quotation-approval",
    code: "qt_approval",
    category: "approval",
    title: { th: "อนุมัติใบเสนอราคาสินค้า (Quotation Approval)", en: "Quotation Approval" },
    description: { th: "ตรวจสอบเงื่อนไขราคา ส่วนลดพิเศษ และเครดิตเทอม ก่อนส่งมอบให้ลูกค้า", en: "Review special pricing, discount overrides and credit terms before client delivery" },
    primaryActionLabel: { th: "อนุมัติใบเสนอราคา", en: "Approve Quotation" },
    documentType: "QT",
  },
  {
    route: "/sales/quotation-cancellation",
    code: "qt_cancellation",
    category: "approval",
    title: { th: "ยกเลิกใบเสนอราคาสินค้า", en: "Quotation Cancellation" },
    description: { th: "บันทึกยกเลิกใบเสนอราคาที่ลูกค้าปฏิเสธหรือหมดอายุ", en: "Void expired or rejected client quotations" },
    primaryActionLabel: { th: "ยกเลิกใบเสนอราคา", en: "Void Quotation" },
    documentType: "QT",
  },
  {
    route: "/sales/order-approval",
    code: "so_approval",
    category: "approval",
    title: { th: "อนุมัติใบสั่งขาย / สั่งจองสินค้า", en: "Sales Order Approval" },
    description: { th: "อนุมัติคำสั่งขาย ตรวจสอบวงเงินสินเชื่อคงเหลือ และกันสต็อกสินค้า", en: "Approve sales orders, verify credit limit and reserve physical inventory" },
    primaryActionLabel: { th: "อนุมัติใบสั่งขาย", en: "Approve Sales Order" },
    documentType: "SO",
  },
  {
    route: "/sales/order-cancellation",
    code: "so_cancellation",
    category: "approval",
    title: { th: "ยกเลิกใบสั่งขาย / สั่งจองสินค้า", en: "Sales Order Cancellation" },
    description: { th: "ยกเลิกคำสั่งขายและคืนยอดสต็อกที่จองไว้กลับสู่คลัง", en: "Cancel sales orders and release inventory reservations" },
    primaryActionLabel: { th: "ยืนยันยกเลิกคำสั่งขาย", en: "Cancel Sales Order" },
    documentType: "SO",
  },

  // --- 2. สั่งจองและกำหนดส่งสินค้า (Order Scheduling) ---
  {
    route: "/sales/reservations",
    code: "sales_reservations",
    category: "reservation",
    title: { th: "ติดตามใบสั่งจองสินค้า (Stock Reservation)", en: "Inventory Reservation Tracking" },
    description: { th: "ติดตามสถานะสินค้าที่ถูกจองตามใบสั่งขาย และกำหนดวันเบิกส่งมอบ", en: "Monitor reserved stock per sales order and delivery scheduling" },
    primaryActionLabel: { th: "อัปเดตการจองสต็อก", en: "Update Reservations" },
  },
  {
    route: "/sales/order-dates",
    code: "order_dates",
    category: "reservation",
    title: { th: "ตรวจสอบวันที่ใบสั่งขาย / สั่งจอง", en: "Order Dates Audit" },
    description: { th: "ตรวจสอบและปรับปรุงวันครบกำหนดและลำดับคิวการส่งมอบสินค้า", en: "Audit order delivery deadlines and customer dispatch queues" },
    primaryActionLabel: { th: "บันทึกการปรับวันที่", en: "Save Scheduled Dates" },
  },
  {
    route: "/sales/delivery-dates",
    code: "delivery_dates",
    category: "reservation",
    title: { th: "ปรับปรุงวันที่ส่งของให้ลูกค้า (Delivery Schedule)", en: "Delivery Promise Management" },
    description: { th: "จัดการตารางเวลารถขนส่งและนัดหมายวันส่งมอบสินค้าให้ลูกค้า", en: "Logistics dispatch scheduling and customer delivery commitment" },
    primaryActionLabel: { th: "อัปเดตกำหนดส่งมอบ", en: "Update Delivery Plan" },
  },

  // --- 3. ภาพรวมและจัดการจัดซื้อ (Procurement Intelligence) ---
  {
    route: "/procurement/dashboard",
    code: "procurement_dashboard",
    category: "procurement",
    title: { th: "ภาพรวมจัดซื้อ (Procurement Dashboard)", en: "Procurement Dashboard" },
    description: { th: "สรุปดัชนีชี้วัดงานจัดซื้อ: ยอดสั่งซื้อ, ระยะเวลาส่งมอบ (Lead Time), และประสิทธิภาพผู้จำหน่าย", en: "Key procurement KPIs: Spend volume, vendor lead times, and supplier performance" },
    primaryActionLabel: { th: "รีเฟรชข้อมูลจัดซื้อ", en: "Refresh Dashboard" },
  },
  {
    route: "/procurement/price-comparison",
    code: "price_comparison",
    category: "procurement",
    title: { th: "ตารางเปรียบเทียบราคาซื้อ (Price Comparison Matrix)", en: "Vendor Price Comparison" },
    description: { th: "เปรียบเทียบใบเสนอราคาจากซัพพลายเออร์หลายรายเพื่อเลือกเงื่อนไขที่ดีที่สุด", en: "Side-by-side vendor quotation comparison for optimal procurement decisions" },
    primaryActionLabel: { th: "สร้างใบสั่งซื้อจากราคาที่เลือก", en: "Generate PO from Best Quote" },
  },
  {
    route: "/procurement/generate-orders",
    code: "generate_orders",
    category: "procurement",
    title: { th: "ประมวลผลใบสั่งซื้ออัตโนมัติ (Auto-PO Generation)", en: "Auto PO Generation" },
    description: { th: "สร้างใบสั่งซื้อ (PO) รวมอัตโนมัติจากใบขอซื้อที่ผ่านอนุมัติ หรือจากจุดสั่งซื้อซ้ำ", en: "Batch generate purchase orders from approved PRs and reorder point alerts" },
    primaryActionLabel: { th: "สร้างใบสั่งซื้อทันที", en: "Generate POs Now" },
  },
  {
    route: "/transaction/documentvault",
    code: "document_vault",
    category: "procurement",
    title: { th: "คลังเอกสารและสแกนบิล (Document Vault & OCR)", en: "Document Vault & OCR" },
    description: { th: "จัดเก็บและอ่านข้อมูลใบเสร็จ/ใบกำกับภาษีด้วย OCR เพื่อสร้างธุรกรรมอัตโนมัติ", en: "Digital repository with OCR receipt scanning for instant transaction entry" },
    primaryActionLabel: { th: "อัปโหลดเอกสารเข้าคลัง", en: "Upload Document" },
  },
  {
    route: "/transaction/documentinbox",
    code: "document_inbox",
    category: "procurement",
    title: { th: "กล่องรับเอกสารระหว่างกิจการ (EDI / Inter-company Inbox)", en: "Inter-company Document Inbox" },
    description: { th: "รับบิลอิเล็กทรอนิกส์และเอกสารส่งตรงจากคู่ค้าหรือสาขาในเครือ", en: "Electronic invoice and document exchange between business affiliates" },
    primaryActionLabel: { th: "นำเข้าเอกสารสู่ระบบ", en: "Import Inbound Document" },
  },

  // --- 4. สินค้าชุด ส่วนประกอบ และทะเบียนเลขเครื่อง (BOM, Sets & Serial) ---
  {
    route: "/inventory/set-assembly",
    code: "set_assembly",
    category: "bom",
    title: { th: "ตรวจสอบและรวมสินค้าชุด (BOM Kit Assembly)", en: "Kit Assembly" },
    description: { th: "เบิกตัดสต็อกชิ้นส่วนย่อยเพื่อรวมเป็นสินค้าชุดสำเร็จรูปพร้อมจำหน่าย", en: "Deduct component stocks and assemble finished kits for sales" },
    primaryActionLabel: { th: "ทำการรวมสินค้าชุด", en: "Assemble Kits" },
  },
  {
    route: "/inventory/set-disassembly",
    code: "set_disassembly",
    category: "bom",
    title: { th: "ตรวจสอบและแยกสินค้าชุด (BOM Kit Disassembly)", en: "Kit Disassembly" },
    description: { th: "แยกสินค้าชุดกลับเป็นชิ้นส่วนย่อยคืนสต็อกตามสูตรการผลิต", en: "Disassemble finished sets back to individual components" },
    primaryActionLabel: { th: "ทำการแยกสินค้าชุด", en: "Disassemble Kits" },
  },
  {
    route: "/inventory/set-components",
    code: "set_components",
    category: "bom",
    title: { th: "สูตรส่วนประกอบสินค้าชุดแบบที่ 2 (Multi-level BOM)", en: "BOM Component Recipes" },
    description: { th: "กำหนดโครงสร้างสูตรการผลิตและอัตราส่วนชิ้นส่วนประกอบหลายระดับ", en: "Manage multi-level Bill of Materials and packaging ratios" },
    primaryActionLabel: { th: "บันทึกสูตรสินค้าชุด", en: "Save Recipe" },
  },
  {
    route: "/productserialregistry",
    code: "serial_registry",
    category: "bom",
    title: { th: "ทะเบียนเลขเครื่องสินค้า (Serial Number Registry)", en: "Serial Number Registry" },
    description: { th: "ติดตามประวัติเครื่องจักร อุปกรณ์ เครื่องใช้ไฟฟ้า และการรับประกันรายชิ้น (Serial Tracking)", en: "Track individual equipment serial numbers, sales history, and warranties" },
    primaryActionLabel: { th: "ลงทะเบียน Serial Number", en: "Register Serial No." },
  },

  // --- 5. ราคาขาย โปรโมชั่น และชั้นต้นทุน (Pricing & Cost Layers) ---
  {
    route: "/inventory/selling-prices",
    code: "selling_prices",
    category: "pricing",
    title: { th: "กำหนดราคาขายสินค้าและระดับราคา (Price Matrix)", en: "Selling Price Matrix" },
    description: { th: "ตั้งตารางราคาขายส่ง-ขายปลีก (Tier Pricing) แยกตามกลุ่มลูกค้า", en: "Configure wholesale/retail price tiers by customer category" },
    primaryActionLabel: { th: "บันทึกราคาขาย", en: "Save Price Matrix" },
  },
  {
    route: "/inventory/price-adjustment",
    code: "price_adjustment",
    category: "pricing",
    title: { th: "ปรับปรุงราคาขายสินค้าจำนวนมาก (Mass Price Update)", en: "Mass Price Adjustment" },
    description: { th: "ปรับราคาขายตามเปอร์เซ็นต์หรือจำนวนเงินพร้อมกันทั้งหมวดหมู่", en: "Adjust pricing across product categories by percentage or fixed amount" },
    primaryActionLabel: { th: "นำการปรับราคาไปใช้", en: "Apply Price Adjustment" },
  },
  {
    route: "/inventory/cost-layers",
    code: "cost_layers",
    category: "pricing",
    title: { th: "ชั้นต้นทุนสต็อกเข้าก่อนออกก่อน (FIFO Cost Layers)", en: "FIFO Cost Layers" },
    description: { th: "แสดงประวัติและมูลค่าของสต็อกแต่ละล็อตที่รับเข้าตามมาตรฐานบัญชี FIFO", en: "Inspect FIFO inventory cost layers and remaining valuations" },
    primaryActionLabel: { th: "ตรวจสอบชั้นต้นทุน", en: "Audit Layers" },
  },
  {
    route: "/inventory/lotexpiry",
    code: "lot_expiry_mgmt",
    category: "pricing",
    title: { th: "จัดการล็อตและวันหมดอายุ (Lot & Expiration Control)", en: "Lot & Expiration Control" },
    description: { th: "ควบคุมการเบิกจ่ายตามวันหมดอายุ (FEFO) เพื่อลดสินค้าสูญเสีย", en: "Manage lot shelf-life and First-Expired-First-Out dispatching" },
    primaryActionLabel: { th: "บันทึกการปรับล็อต", en: "Update Lot Status" },
  },

  // --- 6. ระบบนำเข้าข้อมูล (Data Import Workbench) ---
  {
    route: "/importdocuments",
    code: "import_documents",
    category: "import",
    title: { th: "นำเข้าเอกสารจากไฟล์ (Import Documents)", en: "Import Documents from File" },
    description: { th: "นำเข้าบิลขาย บิลซื้อ หรือใบกำกับภาษีจากไฟล์ Excel/CSV เข้าสู่ระบบโดยตรง", en: "Bulk import sales, purchases, or invoices from Excel/CSV" },
    primaryActionLabel: { th: "เลือกไฟล์เพื่อนำเข้า", en: "Upload File" },
  },
  {
    route: "/importpartner",
    code: "import_partner",
    category: "import",
    title: { th: "นำเข้ารายชื่อคู่ค้าและลูกค้า (Import Contacts)", en: "Import Business Partners" },
    description: { th: "นำเข้ารายชื่อลูกค้า ผู้จำหน่าย เลขประจำตัวผู้เสียภาษี และที่อยู่", en: "Import customers, suppliers, tax IDs and branches from spreadsheet" },
    primaryActionLabel: { th: "นำเข้าคู่ค้า", en: "Upload Contacts" },
  },
  {
    route: "/importproduct",
    code: "import_product",
    category: "import",
    title: { th: "นำเข้ารายการสินค้า (Import Products)", en: "Import Products Catalog" },
    description: { th: "นำเข้ารหัสสินค้า ชื่อ หมวดหมู่ หน่วยนับ และราคาจากไฟล์", en: "Bulk import item master catalog from spreadsheet" },
    primaryActionLabel: { th: "นำเข้ารายการสินค้า", en: "Upload Products" },
  },
  {
    route: "/importproductfromfile",
    code: "import_product_file",
    category: "import",
    title: { th: "นำเข้าสินค้าจากไฟล์รูปแบบพิเศษ", en: "Custom File Product Import" },
    description: { th: "รองรับไฟล์นำเข้าจากระบบ ERP อื่น, Shopee, Lazada หรือ POS", en: "Import product data mapped from external ERP or marketplace formats" },
    primaryActionLabel: { th: "เลือกไฟล์พิเศษ", en: "Select Custom File" },
  },
  {
    route: "/importproductimage",
    code: "import_product_image",
    category: "import",
    title: { th: "นำเข้ารูปภาพสินค้าจำนวนมาก (Bulk Product Images)", en: "Bulk Product Images Upload" },
    description: { th: "จับคู่รูปภาพสินค้ากับรหัสสินค้าโดยอัตโนมัติตามชื่อไฟล์", en: "Match and attach product images automatically by item code filename" },
    primaryActionLabel: { th: "เลือกรูปภาพเพื่อนำเข้า", en: "Upload Images" },
  },
];

const operationsRouteMap = new Map<string, OperationsConfig>(
  OPERATIONS_CONFIGS.map((config) => [config.route, config]),
);

export function isOperationsRoute(route: string): boolean {
  const clean = route.split("?")[0];
  return operationsRouteMap.has(clean);
}

export function getOperationsConfig(route: string): OperationsConfig | undefined {
  const clean = route.split("?")[0];
  return operationsRouteMap.get(clean);
}

// --- การอนุมัติเอกสารจริงผ่าน backend (ไม่มีข้อมูลจำลอง) ---

export interface PendingApprovalDoc {
  docno: string;
  docdate: string;
  requestorname: string;
  counterpartyname: string;
  totalamount: number;
  requiredlevelname: string;
}

// จอที่ backend มี endpoint อนุมัติจริงรองรับแล้วเท่านั้น
// (จอยกเลิกเอกสาร/ใบเสนอราคา/ใบสั่งขาย ยังไม่มี API จึงไม่ผูกไว้ ห้ามเดาว่าใช้ชุดเดียวกัน)
const APPROVAL_KINDS: Record<string, string> = {
  pr_approval: "pr",
};

export function isApprovalApiReady(code: string): boolean {
  return code in APPROVAL_KINDS;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function toNumber(value: unknown): number {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return 0;
}

function toText(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function errorKeyFor(status: number): string {
  return status === 401 || status === 403 ? "unauthorized" : "load_failed";
}

export async function fetchPendingApprovals(params: {
  code: string;
  holdingcode: string;
}): Promise<{ docs: PendingApprovalDoc[]; error?: string }> {
  const kind = APPROVAL_KINDS[params.code];
  if (!kind) return { docs: [], error: "approval_not_available" };
  if (!params.holdingcode) return { docs: [], error: "holding_required" };

  try {
    const res = await authFetch(`/api/goapi/api/approval/${kind}-status/pending`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ holdingcode: params.holdingcode }),
    });
    if (!res.ok) return { docs: [], error: errorKeyFor(res.status) };
    const payload: unknown = await res.json();
    if (!isRecord(payload) || !Array.isArray(payload.data)) return { docs: [], error: "load_failed" };
    const docs = payload.data.filter(isRecord).map((item) => ({
      docno: toText(item.docno),
      docdate: toText(item.docdatetime) || toText(item.createdat),
      requestorname: toText(item.createdbyname) || toText(item.createdby),
      counterpartyname: toText(item.custname) || toText(item.custcode),
      totalamount: toNumber(item.totalamount),
      requiredlevelname: toText(item.requiredlevelname),
    }));
    return { docs };
  } catch {
    return { docs: [], error: "connection_error" };
  }
}

export async function submitApprovalAction(params: {
  code: string;
  holdingcode: string;
  docno: string;
  action: "approve" | "reject";
  actionby: string;
  comment?: string;
}): Promise<{ success: boolean; messageKey: string }> {
  const kind = APPROVAL_KINDS[params.code];
  if (!kind) return { success: false, messageKey: "approval_not_available" };
  if (!params.holdingcode) return { success: false, messageKey: "holding_required" };
  if (!params.docno) return { success: false, messageKey: "docno_required" };
  if (!params.actionby) return { success: false, messageKey: "unauthorized" };

  try {
    const res = await authFetch(`/api/goapi/api/approval/${kind}-status/${params.action}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        holdingcode: params.holdingcode,
        docno: params.docno,
        actionby: params.actionby,
        actionbyname: params.actionby,
        comment: params.comment ?? "",
        source: "app",
      }),
    });
    if (!res.ok) {
      return {
        success: false,
        messageKey: res.status === 401 || res.status === 403 ? "unauthorized" : "action_failed",
      };
    }
    const payload: unknown = await res.json();
    if (!isRecord(payload) || payload.success === false) {
      return { success: false, messageKey: "action_failed" };
    }
    return { success: true, messageKey: params.action === "approve" ? "approve_success" : "reject_success" };
  } catch {
    return { success: false, messageKey: "connection_error" };
  }
}
