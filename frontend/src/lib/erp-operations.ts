// SME Operations, Approvals, BOM Kits, Serial Registry & Data Import Engine

import { apiFetch } from "./client-auth-session";
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
    title: { th: "ตรวจสอบสินค้าที่สามารถรวมเป็นสินค้าชุดได้", en: "Kit Assembly" },
    description: { th: "เบิกตัดสต็อกชิ้นส่วนย่อยเพื่อรวมเป็นสินค้าชุดสำเร็จรูปพร้อมจำหน่าย", en: "Deduct component stocks and assemble finished kits for sales" },
    primaryActionLabel: { th: "ทำการรวมสินค้าชุด", en: "Assemble Kits" },
  },
  {
    route: "/inventory/set-disassembly",
    code: "set_disassembly",
    category: "bom",
    title: { th: "ตรวจสอบชุดสินค้าที่สามารถแยกเป็นสินค้าได้", en: "Kit Disassembly" },
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
  if (status === 401 || status === 403) return "unauthorized";
  // 404 = this approval kind has no endpoint in the backend build, not a
  // transient failure the user should retry.
  if (status === 404) return "module_not_available";
  return "load_failed";
}

export async function fetchPendingApprovals(params: {
  code: string;
  holdingcode: string;
}): Promise<{ docs: PendingApprovalDoc[]; error?: string }> {
  const kind = APPROVAL_KINDS[params.code];
  if (!kind) return { docs: [], error: "approval_not_available" };
  if (!params.holdingcode) return { docs: [], error: "holding_required" };

  try {
    const res = await apiFetch(`/api/goapi/api/approval/${kind}-status/pending`, {
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
    const res = await apiFetch(`/api/goapi/api/approval/${kind}-status/${params.action}`, {
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
