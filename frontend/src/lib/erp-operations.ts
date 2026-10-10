import { apiFetch } from "./client-auth-session";
import {
  approveErpDoc,
  unapproveErpDoc,
  cancelErpDoc,
  approveBatchErpDocs,
  unapproveBatchErpDocs,
  cancelBatchErpDocs,
  type BatchApproveItemResult,
} from "./erp-transaction";
import { catalogText } from "@/lib/catalog-text";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";

export type OperationsCategory =
  | "approval"
  | "reservation"
  | "procurement"
  | "bom"
  | "pricing";

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
    title: { th: "บันทึกอนุมัติใบเสนอซื้อสินค้า", en: "Purchase Requisition Approval" },
    description: { th: "ตรวจทานและอนุมัติใบขอซื้อก่อนออกเป็นใบสั่งซื้อ (PO) ตามลำดับอำนาจอนุมัติ", en: "Review and approve purchase requisitions based on authorization hierarchy" },
    primaryActionLabel: { th: "อนุมัติรายการที่เลือก", en: "Approve Selected" },
    documentType: "PR",
  },
  {
    route: "/procurement/order-cancellation",
    code: "po_cancellation",
    category: "approval",
    title: { th: "ยกเลิกใบสั่งซื้อสินค้า", en: "Purchase Order Cancellation" },
    description: { th: "พิจารณายกเลิกใบสั่งซื้อและปลดภาระผูกพันทางการเงิน", en: "Cancel approved purchase orders and release commitments" },
    primaryActionLabel: { th: "ยืนยันการยกเลิก PO", en: "Confirm PO Cancellation" },
    documentType: "PO",
  },
  {
    route: "/sales/quotation-approval",
    code: "qt_approval",
    category: "approval",
    title: { th: "อนุมัติใบเสนอราคาสินค้า", en: "Quotation Approval" },
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
    title: { th: "อนุมัติใบสั่งขาย/สั่งจองสินค้า", en: "Sales Order Approval" },
    description: { th: "อนุมัติคำสั่งขาย ตรวจสอบวงเงินสินเชื่อคงเหลือ และกันสต็อกสินค้า", en: "Approve sales orders, verify credit limit and reserve physical inventory" },
    primaryActionLabel: { th: "อนุมัติใบสั่งขาย", en: "Approve Sales Order" },
    documentType: "SO",
  },
  {
    route: "/sales/order-cancellation",
    code: "so_cancellation",
    category: "approval",
    title: { th: "ยกเลิกใบสั่งขาย/สั่งจองสินค้า", en: "Sales Order Cancellation" },
    description: { th: "ยกเลิกคำสั่งขายและคืนยอดสต็อกที่จองไว้กลับสู่คลัง", en: "Cancel sales orders and release inventory reservations" },
    primaryActionLabel: { th: "ยืนยันยกเลิกคำสั่งขาย", en: "Cancel Sales Order" },
    documentType: "SO",
  },
  {
    route: "/procurement/requisition-cancellation",
    code: "pr_cancellation",
    category: "approval",
    title: { th: "ยกเลิกใบเสนอซื้อสินค้า", en: "Cancel Purchase Requisition" },
    description: { th: "ยกเลิกใบเสนอซื้อสินค้าที่ไม่ต้องการดำเนินการต่อ", en: "Cancel purchase requisitions that will not proceed" },
    primaryActionLabel: { th: "ยืนยันยกเลิกใบเสนอซื้อ", en: "Confirm Requisition Cancellation" },
    documentType: "PR",
  },
  {
    route: "/procurement/requisition-approval-cancellation",
    code: "pr_approval_cancellation",
    category: "approval",
    title: { th: "ยกเลิกใบอนุมัติเสนอซื้อสินค้า", en: "Cancel Requisition Approval" },
    description: { th: "ยกเลิกใบอนุมัติเสนอซื้อสินค้าที่บันทึกไว้แล้ว", en: "Cancel recorded purchase requisition approvals" },
    primaryActionLabel: { th: "ยืนยันยกเลิกใบอนุมัติ", en: "Confirm Approval Cancellation" },
    documentType: "PR",
  },
  {
    route: "/procurement/order-approval",
    code: "po_approval",
    category: "approval",
    title: { th: "อนุมัติใบสั่งซื้อสินค้า", en: "Approve Purchase Order" },
    description: { th: "ตรวจทานและอนุมัติใบสั่งซื้อสินค้าตามลำดับอำนาจอนุมัติ", en: "Review and approve purchase orders based on authorization hierarchy" },
    primaryActionLabel: { th: "อนุมัติใบสั่งซื้อ", en: "Approve Purchase Order" },
    documentType: "PO",
  },
  {
    route: "/inventory/issue-request-cancellation",
    code: "stock_issue_request_cancellation",
    category: "approval",
    title: { th: "ยกเลิกใบขอเบิกใช้สินค้า,วัตถุดิบ", en: "Cancel Stock Issue Request" },
    description: { th: "ยกเลิกใบขอเบิกใช้สินค้าและวัตถุดิบที่ไม่ต้องการดำเนินการต่อ", en: "Cancel stock and material issue requests that will not proceed" },
    primaryActionLabel: { th: "ยืนยันยกเลิกใบขอเบิก", en: "Confirm Issue Request Cancellation" },
    documentType: "SIR",
  },
  {
    route: "/inventory/transfer-request-cancellation",
    code: "stock_transfer_request_cancellation",
    category: "approval",
    title: { th: "ยกเลิกใบขอโอนสินค้า", en: "Cancel Stock Transfer Request" },
    description: { th: "ยกเลิกใบขอโอนสินค้าที่ไม่ต้องการดำเนินการต่อ", en: "Cancel stock transfer requests that will not proceed" },
    primaryActionLabel: { th: "ยืนยันยกเลิกใบขอโอน", en: "Confirm Transfer Request Cancellation" },
    documentType: "STR",
  },

  // --- 2. สั่งจองและกำหนดส่งสินค้า (Order Scheduling) ---
  {
    route: "/sales/reservations",
    code: "sales_reservations",
    category: "reservation",
    title: { th: "Flow ใบสั่งจองสินค้า", en: "Inventory Reservation Tracking" },
    description: { th: "ติดตามสถานะสินค้าที่ถูกจองตามใบสั่งขาย และกำหนดวันเบิกส่งมอบ", en: "Monitor reserved stock per sales order and delivery scheduling" },
    primaryActionLabel: { th: "อัปเดตการจองสต็อก", en: "Update Reservations" },
  },
  {
    route: "/sales/order-dates",
    code: "order_dates",
    category: "reservation",
    title: { th: "ตรวจสอบวันที่ใบสั่งขาย/สั่งจองสินค้า", en: "Order Dates Audit" },
    description: { th: "ตรวจสอบและปรับปรุงวันครบกำหนดและลำดับคิวการส่งมอบสินค้า", en: "Audit order delivery deadlines and customer dispatch queues" },
    primaryActionLabel: { th: "บันทึกการปรับวันที่", en: "Save Scheduled Dates" },
  },
  {
    route: "/sales/delivery-dates",
    code: "delivery_dates",
    category: "reservation",
    title: { th: "ปรับปรุงวันที่ส่งของให้ลูกค้า", en: "Delivery Promise Management" },
    description: { th: "จัดการตารางเวลารถขนส่งและนัดหมายวันส่งมอบสินค้าให้ลูกค้า", en: "Logistics dispatch scheduling and customer delivery commitment" },
    primaryActionLabel: { th: "อัปเดตกำหนดส่งมอบ", en: "Update Delivery Plan" },
  },

  // --- 3. ภาพรวมและจัดการจัดซื้อ (Procurement Intelligence) ---
  {
    route: "/procurement/delivery-dates",
    code: "purchase_receipt_dates",
    category: "procurement",
    title: { th: "ปรับปรุงวันที่รับสินค้าจากเจ้าหนี้", en: "Adjust Supplier Receipt Date" },
    description: { th: "ปรับปรุงวันที่รับสินค้าหรือวันที่วางบิล และคำนวณวันครบกำหนดชำระเงินใหม่", en: "Adjust receipt/billing date and recalculate payment due dates" },
    primaryActionLabel: { th: "บันทึกการปรับวันที่", en: "Save Scheduled Dates" },
  },
  {
    route: "/transaction/purchasereceiptadjustment",
    code: "purchase_receipt_adjustment",
    category: "procurement",
    title: { th: "ปรับปรุงใบรับสินค้า", en: "Adjust Purchase Receipt" },
    description: { th: "สืบค้นและปรับปรุงสถานะใบรับสินค้า (GR) ที่มีการตั้งหนี้หรือรับบิลครบแล้ว", en: "Search and adjust Goods Receipt (GR) status to closed after invoice receipt is matched" },
    primaryActionLabel: { th: "ปรับปรุงสถานะ GR", en: "Adjust GR Status" },
  },
  {
    route: "/procurement/price-comparison",
    code: "price_comparison",
    category: "procurement",
    title: { th: "รายงานเปรียบเทียบราคาซื้อ", en: "Vendor Price Comparison" },
    description: { th: "เปรียบเทียบใบเสนอราคาจากซัพพลายเออร์หลายรายเพื่อเลือกเงื่อนไขที่ดีที่สุด", en: "Side-by-side vendor quotation comparison for optimal procurement decisions" },
    primaryActionLabel: { th: "สร้างใบสั่งซื้อจากราคาที่เลือก", en: "Generate PO from Best Quote" },
  },
  {
    route: "/procurement/generate-orders",
    code: "generate_orders",
    category: "procurement",
    title: { th: "ประมวลผลใบสั่งซื้อสินค้าอัตโนมัติ", en: "Auto PO Generation" },
    description: { th: "สร้างใบสั่งซื้อ (PO) รวมอัตโนมัติจากใบขอซื้อที่ผ่านอนุมัติ หรือจากจุดสั่งซื้อซ้ำ", en: "Batch generate purchase orders from approved PRs and reorder point alerts" },
    primaryActionLabel: { th: "สร้างใบสั่งซื้อทันที", en: "Generate POs Now" },
  },

  // --- 4. สินค้าชุด ส่วนประกอบ และทะเบียนเลขเครื่อง (BOM, Sets & Serial) ---
  {
    route: "/inventory/set-assembly",
    code: "set_assembly",
    category: "bom",
    title: { th: "บันทึกการจัดชุดสินค้า", en: "Kit Assembly" },
    description: { th: "เบิกตัดสต็อกชิ้นส่วนย่อยเพื่อรวมเป็นสินค้าชุดสำเร็จรูปพร้อมจำหน่าย", en: "Deduct component stocks and assemble finished kits for sales" },
    primaryActionLabel: { th: "ทำการรวมสินค้าชุด", en: "Assemble Kits" },
  },
  {
    route: "/inventory/set-disassembly",
    code: "set_disassembly",
    category: "bom",
    title: { th: "บันทึกการแยกสินค้าออกจากชุด", en: "Kit Disassembly" },
    description: { th: "แยกสินค้าชุดกลับเป็นชิ้นส่วนย่อยคืนสต็อกตามสูตรการผลิต", en: "Disassemble finished sets back to individual components" },
    primaryActionLabel: { th: "ทำการแยกสินค้าชุด", en: "Disassemble Kits" },
  },
  {
    route: "/inventory/set-components",
    code: "set_components",
    category: "bom",
    title: { th: "บันทึกรายการย่อยสินค้าชุดแบบที่ 2", en: "BOM Component Recipes" },
    description: { th: "กำหนดโครงสร้างสูตรการผลิตและอัตราส่วนชิ้นส่วนประกอบหลายระดับ", en: "Manage multi-level Bill of Materials and packaging ratios" },
    primaryActionLabel: { th: "บันทึกสูตรสินค้าชุด", en: "Save Recipe" },
  },
  {
    route: "/productserialregistry",
    code: "serial_registry",
    category: "bom",
    title: { th: "บันทึก Serial Number", en: "Serial Number Registry" },
    description: { th: "ติดตามประวัติเครื่องจักร อุปกรณ์ เครื่องใช้ไฟฟ้า และการรับประกันรายชิ้น (Serial Tracking)", en: "Track individual equipment serial numbers, sales history, and warranties" },
    primaryActionLabel: { th: "ลงทะเบียน Serial Number", en: "Register Serial No." },
  },

  // --- 5. ราคาขาย โปรโมชั่น และชั้นต้นทุน (Pricing & Cost Layers) ---
  {
    route: "/inventory/selling-prices",
    code: "selling_prices",
    category: "pricing",
    title: { th: "กำหนดราคาขายสินค้า", en: "Selling Price Matrix" },
    description: { th: "ตั้งตารางราคาขายส่ง-ขายปลีก (Tier Pricing) แยกตามกลุ่มลูกค้า", en: "Configure wholesale/retail price tiers by customer category" },
    primaryActionLabel: { th: "บันทึกราคาขาย", en: "Save Price Matrix" },
  },
  {
    route: "/inventory/price-adjustment",
    code: "price_adjustment",
    category: "pricing",
    title: { th: "ปรับปรุงราคาขายสินค้า", en: "Mass Price Adjustment" },
    description: { th: "ปรับราคาขายตามเปอร์เซ็นต์หรือจำนวนเงินพร้อมกันทั้งหมวดหมู่", en: "Adjust pricing across product categories by percentage or fixed amount" },
    primaryActionLabel: { th: "นำการปรับราคาไปใช้", en: "Apply Price Adjustment" },
  },

  // --- 6. ระบบนำเข้าข้อมูล (Data Import Workbench) ---
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
  id?: string;
  docno: string;
  docdate: string;
  requestorname: string;
  counterpartyname: string;
  totalamount: number;
  requiredlevelname?: string;
  approval_status?: number;
  status?: number;
  iscancel?: boolean;
}

export interface ApprovalOperationMeta {
  module: string;
  actionType: "approve" | "unapprove" | "cancel";
  statusFilter: "pending" | "approved" | "active";
}

export const APPROVAL_OPERATIONS: Record<string, ApprovalOperationMeta> = {
  pr_approval: { module: "purchase-requisition", actionType: "approve", statusFilter: "pending" },
  po_approval: { module: "purchase-order", actionType: "approve", statusFilter: "pending" },
  qt_approval: { module: "quotation", actionType: "approve", statusFilter: "pending" },
  so_approval: { module: "sale-order", actionType: "approve", statusFilter: "pending" },
  pr_approval_cancellation: { module: "purchase-requisition", actionType: "unapprove", statusFilter: "approved" },
  pr_cancellation: { module: "purchase-requisition", actionType: "cancel", statusFilter: "active" },
  po_cancellation: { module: "purchase-order", actionType: "cancel", statusFilter: "active" },
  qt_cancellation: { module: "quotation", actionType: "cancel", statusFilter: "active" },
  so_cancellation: { module: "sale-order", actionType: "cancel", statusFilter: "active" },
  stock_issue_request_cancellation: { module: "stock-issue-request", actionType: "cancel", statusFilter: "active" },
  stock_transfer_request_cancellation: { module: "stock-transfer-request", actionType: "cancel", statusFilter: "active" },
};

export function isApprovalApiReady(code: string): boolean {
  return code in APPROVAL_OPERATIONS;
}

export function getApprovalOperationMeta(code: string): ApprovalOperationMeta | undefined {
  return APPROVAL_OPERATIONS[code];
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
  if (status === 401 || status === 403) return "unauthorized";
  if (status === 404) return "module_not_available";
  return "load_failed";
}

function mapDocItem(item: Record<string, unknown>): PendingApprovalDoc {
  return {
    id: toText(item.id || item.docno),
    docno: toText(item.docno),
    docdate: toText(item.docdatetime) || toText(item.createdat),
    requestorname: toText(item.createdbyname) || toText(item.createdby),
    counterpartyname: toText(item.custname) || toText(item.custcode),
    totalamount: toNumber(item.totalamount),
    requiredlevelname: toText(item.requiredlevelname),
    approval_status: toNumber(item.approval_status),
    status: toNumber(item.status),
    iscancel: Boolean(item.iscancel),
  };
}

export async function fetchPendingApprovals(params: {
  code: string;
  holdingcode?: string;
  q?: string;
  offset?: number;
  limit?: number;
}): Promise<{ docs: PendingApprovalDoc[]; total: number; error?: string }> {
  const meta = APPROVAL_OPERATIONS[params.code];
  if (!meta) return { docs: [], total: 0, error: "approval_not_available" };

  try {
    const query = new URLSearchParams();
    query.set("status", meta.statusFilter);
    if (params.q) query.set("q", params.q);
    if (params.offset !== undefined) query.set("offset", String(params.offset));
    if (params.limit !== undefined) query.set("limit", String(params.limit));

    const res = await apiFetch(`/api/erp-transaction/${encodeURIComponent(meta.module)}/list?${query.toString()}`);
    if (!res.ok) return { docs: [], total: 0, error: errorKeyFor(res.status) };

    const payload: unknown = await res.json();
    if (!isRecord(payload) || !Array.isArray(payload.data)) {
      if (Array.isArray(payload)) {
        const docs = payload.filter(isRecord).map(mapDocItem);
        return { docs, total: docs.length };
      }
      return { docs: [], total: 0, error: "load_failed" };
    }

    const docs = payload.data.filter(isRecord).map(mapDocItem);
    const pagination = isRecord(payload.pagination) ? payload.pagination : undefined;
    const total = typeof pagination?.total === "number" ? pagination.total : docs.length;
    return { docs, total };
  } catch {
    return { docs: [], total: 0, error: "connection_error" };
  }
}

export async function submitApprovalAction(params: {
  code: string;
  holdingcode?: string;
  docno: string;
  action: "approve" | "unapprove" | "cancel" | "reject";
  actionby?: string;
  comment?: string;
  language?: string;
}): Promise<{ success: boolean; messageKey: string; detail?: string }> {
  const meta = APPROVAL_OPERATIONS[params.code];
  if (!meta) return { success: false, messageKey: "approval_not_available" };
  if (!params.docno) return { success: false, messageKey: "docno_required" };

  const lang = params.language || "th";

  if (params.action === "approve") {
    const res = await approveErpDoc(meta.module, params.docno, lang);
    return {
      success: res.success,
      messageKey: res.success ? "approve_success" : res.message || "action_failed",
      detail: res.detail,
    };
  }

  if (params.action === "unapprove") {
    const res = await unapproveErpDoc(meta.module, params.docno, params.comment || "", lang);
    return {
      success: res.success,
      messageKey: res.success ? "unapprove_success" : res.message || "action_failed",
      detail: res.detail,
    };
  }

  if (params.action === "cancel") {
    const res = await cancelErpDoc(meta.module, params.docno, params.comment || "", lang);
    return {
      success: res.success,
      messageKey: res.success ? "cancel_success" : res.message || "action_failed",
      detail: res.detail,
    };
  }

  // legacy "reject" maps to unapprove or cancel depending on meta.actionType
  if (meta.actionType === "unapprove") {
    const res = await unapproveErpDoc(meta.module, params.docno, params.comment || "", lang);
    return {
      success: res.success,
      messageKey: res.success ? "unapprove_success" : res.message || "action_failed",
      detail: res.detail,
    };
  }

  const res = await cancelErpDoc(meta.module, params.docno, params.comment || "", lang);
  return {
    success: res.success,
    messageKey: res.success ? "cancel_success" : res.message || "action_failed",
    detail: res.detail,
  };
}

export async function submitBatchApprove(params: {
  code: string;
  docnos: string[];
  language?: string;
}): Promise<{
  success: boolean;
  approved: number;
  total: number;
  messageKey: string;
  detail?: string;
  items?: BatchApproveItemResult[];
}> {
  const meta = APPROVAL_OPERATIONS[params.code];
  if (!meta) return { success: false, approved: 0, total: 0, messageKey: "approval_not_available" };
  const res = await approveBatchErpDocs(meta.module, params.docnos, params.language || "th");
  return {
    success: res.success,
    approved: res.approved ?? 0,
    total: res.total ?? params.docnos.length,
    messageKey: res.success ? "approve_success" : res.message || "action_failed",
    detail: res.detail,
    items: res.items,
  };
}

export async function submitBatchUnapprove(params: {
  code: string;
  docnos: string[];
  comment?: string;
  language?: string;
}): Promise<{
  success: boolean;
  unapproved: number;
  total: number;
  messageKey: string;
  detail?: string;
  items?: BatchApproveItemResult[];
}> {
  const meta = APPROVAL_OPERATIONS[params.code];
  if (!meta) return { success: false, unapproved: 0, total: 0, messageKey: "approval_not_available" };
  const res = await unapproveBatchErpDocs(meta.module, params.docnos, params.comment, params.language || "th");
  return {
    success: res.success,
    unapproved: res.unapproved ?? 0,
    total: res.total ?? params.docnos.length,
    messageKey: res.success ? "unapprove_success" : res.message || "action_failed",
    detail: res.detail,
    items: res.items,
  };
}

export async function submitBatchCancel(params: {
  code: string;
  docnos: string[];
  comment: string;
  language?: string;
}): Promise<{
  success: boolean;
  cancelled: number;
  total: number;
  messageKey: string;
  detail?: string;
  items?: BatchApproveItemResult[];
}> {
  const meta = APPROVAL_OPERATIONS[params.code];
  if (!meta) return { success: false, cancelled: 0, total: 0, messageKey: "approval_not_available" };
  const res = await cancelBatchErpDocs(meta.module, params.docnos, params.comment, params.language || "th");
  return {
    success: res.success,
    cancelled: res.cancelled ?? 0,
    total: res.total ?? params.docnos.length,
    messageKey: res.success ? "cancel_success" : res.message || "action_failed",
    detail: res.detail,
    items: res.items,
  };
}

// 2026-09-16: every user-visible string above also lives in languages.tsv,
// keyed by `<code>.<part>`. The literals stay as the offline fallback.
const catalogKeys: Record<string, string> = {
  "pr_approval.title": "purchase_requisition_approve",
  "pr_approval.description": "ops_review_and_approve_purchase_requisitions",
  "po_cancellation.title": "purchase_order_cancel",
  "po_cancellation.description": "ops_cancel_approved_purchase_orders_and",
  "qt_approval.title": "quotation_approve",
  "qt_approval.description": "ops_review_special_pricing_discount_overrides",
  "qt_cancellation.title": "quotation_cancel",
  "qt_cancellation.description": "ops_void_expired_or_rejected_client",
  "so_approval.title": "sale_order_approve",
  "so_approval.description": "ops_approve_sales_orders_verify_credit",
  "so_cancellation.title": "sale_order_cancel",
  "so_cancellation.description": "ops_cancel_sales_orders_and_release",
  "pr_cancellation.title": "purchase_requisition_cancel",
  "pr_cancellation.description": "ops_cancel_purchase_requisitions_not_proceeding",
  "pr_approval_cancellation.title": "purchase_requisition_approval_cancel",
  "pr_approval_cancellation.description": "ops_cancel_recorded_purchase_requisition_approvals",
  "po_approval.title": "purchase_order_approve",
  "po_approval.description": "ops_review_and_approve_purchase_orders",
  "stock_issue_request_cancellation.title": "stock_issue_request_cancel",
  "stock_issue_request_cancellation.description": "ops_cancel_stock_issue_requests_not_proceeding",
  "stock_transfer_request_cancellation.title": "stock_transfer_request_cancel",
  "stock_transfer_request_cancellation.description": "ops_cancel_stock_transfer_requests_not_proceeding",
  "sales_reservations.title": "sale_reservation_flow",
  "sales_reservations.description": "ops_monitor_reserved_stock_per_sales",
  "order_dates.title": "sale_order_date_check",
  "order_dates.description": "ops_audit_order_delivery_deadlines_and",
  "delivery_dates.title": "sale_delivery_date",
  "delivery_dates.description": "ops_logistics_dispatch_scheduling_and_customer",
  "price_comparison.title": "purchase_price_comparison",
  "price_comparison.description": "ops_side_by_side_vendor_quotation",
  "generate_orders.title": "purchase_order_generate",
  "generate_orders.description": "ops_batch_generate_purchase_orders_from",
  "set_assembly.title": "product_set_assemble",
  "set_assembly.description": "ops_deduct_component_stocks_and_assemble",
  "set_disassembly.title": "product_set_disassemble",
  "set_disassembly.description": "ops_disassemble_finished_sets_back_to",
  "set_components.title": "product_set_components",
  "set_components.description": "ops_manage_multi_level_bill_of",
  "serial_registry.title": "product_serial_registry",
  "serial_registry.description": "ops_track_individual_equipment_serial_numbers",
  "selling_prices.title": "product_sale_price",
  "selling_prices.description": "ops_configure_wholesale_retail_price_tiers",
  "price_adjustment.title": "product_price_adjust",
  "price_adjustment.description": "ops_adjust_pricing_across_product_categories",
};

export function operationsText(
  config: OperationsConfig,
  part: "title" | "description",
  language: LanguageCode,
  dictionary?: BackendLanguageDictionary,
): string {
  return catalogText(catalogKeys, `${config.code}.${part}`, config[part], language, dictionary);
}
