// ERP Tools & Integrity Recalculate Engine for Thai SMEs & Thai Accounting
// Covers GL Reprocess, Stock Rebuild, AR/AP Recalculate, Bank & Cheque Balance Recalculation

import { authFetch } from "@/lib/client-auth-session";
import { catalogText } from "@/lib/catalog-text";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";

export interface ErpToolConfig {
  route: string;
  code: string;
  domain: "gl" | "inventory" | "ar" | "ap" | "bank";
  title: { th: string; en: string };
  description: { th: string; en: string };
  actionLabel: { th: string; en: string };
  steps: { th: string; en: string }[];
  warningMessage?: { th: string; en: string };
}

export const ERP_TOOL_CONFIGS: ErpToolConfig[] = [
  {
    route: "/gl/reprocess",
    code: "gl_reprocess",
    domain: "gl",
    title: { th: "ประมวลผลข้อมูลบัญชีใหม่", en: "Reprocess Accounting Data" },
    description: { th: "ประมวลผลการลงสมุดรายวัน ยอดแยกประเภท และงบทดลองใหม่ทั้งหมดจากการบันทึกธุรกรรม", en: "Reprocess general ledger journals, ledger postings and trial balance from source documents" },
    actionLabel: { th: "เริ่มประมวลผลบัญชีใหม่", en: "Start GL Reprocessing" },
    steps: [
      { th: "ตรวจสอบความสมดุลเดบิต-เครดิตในสมุดรายวัน", en: "Audit Debit-Credit balance in journals" },
      { th: "ล้างยอดสรุปแยกประเภทชั่วคราว", en: "Clear temporary ledger summary projections" },
      { th: "คำนวณยอดสะสมยกมาและยอดเคลื่อนไหวตามงวด", en: "Recalculate period balances and cumulative totals" },
      { th: "ปรับปรุงงบทดลองและงบการเงินให้เป็นปัจจุบัน", en: "Update trial balance and financial statements" },
    ],
  },
  {
    route: "/tools/ar-recalculate",
    code: "ar_recalculate",
    domain: "ar",
    title: { th: "คำนวณยอดลูกหนี้ใหม่", en: "Recalculate Debtor Balances" },
    description: { th: "คำนวณยอดหนี้คงค้างของลูกหนี้ทุกรายจากเอกสารขาย ใบวางบิล ใบลดหนี้ และใบเสร็จรับเงิน", en: "Recalculate outstanding debtor balances from sales, billings, credit notes, and receipts" },
    actionLabel: { th: "คำนวณยอดหนี้ลูกหนี้ใหม่", en: "Recalculate AR Balances" },
    steps: [
      { th: "รวบรวมบิลขายและใบแจ้งหนี้ที่ยังค้างชำระ", en: "Aggregate unpaid sales invoices" },
      { th: "กระทบยอดการรับชำระเงินและใบลดหนี้", en: "Match payments and credit note applications" },
      { th: "ปรับปรุงยอดหนี้สุทธิและวงเงินสินเชื่อของลูกหนี้", en: "Update net balance and customer credit limits" },
    ],
  },
  {
    route: "/tools/ar-bill-balances",
    code: "ar_bill_balances",
    domain: "ar",
    title: { th: "คำนวณยอดคงเหลือของบิลใหม่", en: "Recalculate AR Invoice Balances" },
    description: { th: "ปรับยอดคงเหลือรายบิล (Bill-by-Bill Balance) เพื่อความถูกต้องในการตัดรับชำระ", en: "Reconcile individual invoice outstanding amounts for accurate receipt allocation" },
    actionLabel: { th: "คำนวณยอดบิลลูกหนี้ใหม่", en: "Recompute Bill Balances" },
    steps: [
      { th: "ตรวจนับยอดตัดจ่ายรายบิล", en: "Verify itemized bill allocations" },
      { th: "ปรับปรุงยอดค้างรับของแต่ละบิล", en: "Update outstanding balance per bill" },
    ],
  },
  {
    route: "/tools/ap-recalculate",
    code: "ap_recalculate",
    domain: "ap",
    title: { th: "คำนวณยอดเจ้าหนี้ใหม่", en: "Recalculate Creditor Balances" },
    description: { th: "คำนวณยอดหนี้คงค้างของเจ้าหนี้จากใบรับสินค้า ใบซื้อ ใบสำคัญจ่าย และใบลดหนี้", en: "Recalculate outstanding vendor payables from purchases, vouchers, and debit notes" },
    actionLabel: { th: "คำนวณยอดหนี้เจ้าหนี้ใหม่", en: "Recalculate AP Balances" },
    steps: [
      { th: "รวบรวมใบรับสินค้าและตั้งหนี้ค้างจ่าย", en: "Aggregate purchase invoices and accrued payables" },
      { th: "กระทบยอดใบสำคัญจ่ายและการชำระเงิน", en: "Reconcile payment vouchers and settlements" },
      { th: "ปรับปรุงยอดหนี้คงค้างสุทธิของเจ้าหนี้แต่ละราย", en: "Update net vendor balance" },
    ],
  },
  {
    route: "/tools/ap-bill-balances",
    code: "ap_bill_balances",
    domain: "ap",
    title: { th: "คำนวณยอดคงเหลือของบิลใหม่", en: "Recalculate AP Bill Balances" },
    description: { th: "คำนวณยอดค้างจ่ายรายใบกำกับภาษีซื้อ เพื่อการจ่ายชำระที่แม่นยำ", en: "Recompute outstanding balance for each vendor invoice" },
    actionLabel: { th: "คำนวณยอดบิลเจ้าหนี้ใหม่", en: "Recompute Vendor Bill Balances" },
    steps: [
      { th: "ตรวจสอบประวัติการจ่ายชำระรายบิลซื้อ", en: "Inspect payment history per purchase invoice" },
      { th: "ปรับปรุงยอดค้างจ่ายสุทธิรายเอกสาร", en: "Update remaining payable amount per document" },
    ],
  },
  {
    route: "/tools/cheque-balances",
    code: "cheque_balances",
    domain: "bank",
    title: { th: "คำนวณยอดคงเหลือของเช็คใหม่", en: "Recalculate Cheque Balances" },
    description: { th: "ปรับสถานะเช็ครับและเช็คจ่ายในมือ เช็คผ่าน เช็คคืน และเช็คยกเลิก", en: "Synchronize on-hand, cleared, returned, and voided cheque balances" },
    actionLabel: { th: "คำนวณยอดเช็คใหม่", en: "Recalculate Cheques" },
    steps: [
      { th: "ตรวจสอบสถานะทะเบียนเช็ครับและเช็คจ่าย", en: "Verify status of received and issued cheques" },
      { th: "ปรับปรุงยอดเช็คในมือที่ยังไม่นำฝาก/ยังไม่ตัดบัญชี", en: "Update pending cheque totals" },
    ],
  },
  {
    route: "/tools/bank-balances",
    code: "bank_balances",
    domain: "bank",
    title: { th: "คำนวณยอดคงเหลือของสมุดบัญชีใหม่", en: "Recalculate Bank Book Balances" },
    description: { th: "คำนวณยอดคงเหลือในสมุดบัญชีเงินฝากธนาคารทุกบัญชีจากการโอน ฝาก ถอน", en: "Recalculate ledger balances for all bank accounts from deposits, withdrawals and transfers" },
    actionLabel: { th: "คำนวณยอดสมุดบัญชีใหม่", en: "Recalculate Bank Books" },
    steps: [
      { th: "เรียงลำดับรายการเดินบัญชีตามวันที่และเวลา", en: "Sort transactions chronologically" },
      { th: "คำนวณยอดคงเหลือสะสมทีละรายการ", en: "Compute rolling balance for each account" },
      { th: "เทียบยอดกับสมุดบัญชีเงินฝากและ Statement", en: "Reconcile with passbook and bank statements" },
    ],
  },
  {
    route: "/rebuildstockscreen",
    code: "rebuild_stock",
    domain: "inventory",
    title: { th: "คำนวณยอดสินค้าใหม่", en: "Rebuild Stock Data" },
    description: { th: "คำนวณยอดสินค้าคงเหลือและต้นทุนเฉลี่ยใหม่ทั้งหมดจากประวัติเอกสารเข้า-ออก", en: "Reconstruct physical inventory on-hand balances and moving average costs from scratch" },
    actionLabel: { th: "เริ่มคำนวณสต็อกใหม่ทั้งหมด", en: "Start Full Stock Rebuild" },
    steps: [
      { th: "ตรวจสอบความต่อเนื่องของเอกสารคลังสินค้า", en: "Audit warehouse document sequence" },
      { th: "คำนวณยอดรับเข้า-จ่ายออกตามลำดับวันที่", en: "Recompute in-out movements chronologically" },
      { th: "คำนวณต้นทุนต่อหน่วยและมูลค่าคงเหลือใหม่", en: "Recompute unit costs and inventory valuation" },
      { th: "ปรับปรุงตารางยอดคงเหลือในฐานข้อมูล", en: "Update balance summary tables" },
    ],
  },
  {
    route: "/auditscreen",
    code: "audit_data",
    domain: "inventory",
    title: { th: "ตรวจข้อมูล", en: "Data Integrity Audit" },
    description: { th: "ตรวจสอบข้อผิดพลาด ยอดติดลบ เอกสารค้าง และความไม่สอดคล้องในระบบ", en: "Scan for anomalies, negative stocks, orphaned documents, and data inconsistencies" },
    actionLabel: { th: "เริ่มตรวจความสมบูรณ์ของระบบ", en: "Run Integrity Audit" },
    steps: [
      { th: "สแกนหาสินค้าที่ยอดคงเหลือติดลบ", en: "Scan for negative inventory balances" },
      { th: "ตรวจสอบเอกสารที่ไม่มีรายการย่อย", en: "Check for empty line item documents" },
      { th: "ตรวจสอบความถูกต้องของรหัสสาขาและคลัง", en: "Validate branch and warehouse codes" },
    ],
  },
  {
    route: "/rebuildproductsscreen",
    code: "rebuild_products",
    domain: "inventory",
    title: { th: "สร้างรายการสินค้าใหม่", en: "Rebuild Product Indexes" },
    description: { th: "สร้างดัชนีการค้นหา บาร์โค้ด และหน่วยนับของสินค้าใหม่เพื่อความรวดเร็ว", en: "Rebuild search index, barcode catalogs, and packaging units" },
    actionLabel: { th: "สร้างดัชนีสินค้าใหม่", en: "Rebuild Product Index" },
    steps: [
      { th: "รวบรวมรายการสินค้าและบาร์โค้ดทั้งหมด", en: "Collect all items and barcodes" },
      { th: "สร้าง Full-text search index ภาษาไทยและอังกฤษ", en: "Generate multilingual full-text search index" },
    ],
  },
  {
    route: "/rebuildproductbalancescreen",
    code: "rebuild_product_balance",
    domain: "inventory",
    title: { th: "สร้างยอดคงเหลือใหม่", en: "Rebuild Product Balances" },
    description: { th: "อัปเดตยอดคงเหลือระดับสินค้าและระดับคลังให้ตรงกับฐานข้อมูลจริง", en: "Synchronize item-level and warehouse-level stock balance snapshots" },
    actionLabel: { th: "สร้างยอดคงเหลือใหม่", en: "Rebuild Balance Snapshots" },
    steps: [
      { th: "สแกนยอดคงเหลือในทุกคลังสินค้า", en: "Scan on-hand quantities across all warehouses" },
      { th: "ปรับปรุงตารางแคชยอดคงเหลือ (Balance Snapshot)", en: "Update balance snapshot cache" },
    ],
  },
  {
    route: "/inventory/daily-sequence",
    code: "inventory_daily_sequence",
    domain: "inventory",
    title: { th: "กำหนดลำดับรายวัน", en: "Inventory Daily Sequence" },
    description: { th: "จัดลำดับเวลาการเกิดเอกสารรับ-จ่ายในแต่ละวัน เพื่อคำนวณต้นทุน FIFO ให้ถูกต้อง", en: "Order daily receipts and issues chronologically for precise FIFO costing" },
    actionLabel: { th: "จัดลำดับเอกสารรายวัน", en: "Sort Daily Transactions" },
    steps: [
      { th: "จัดคิวเอกสารรับเข้าก่อนเอกสารจ่ายออกในวันเดียวกัน", en: "Sequence receipts before issues within the same day" },
      { th: "ปรับปรุงลำดับการตัดสต็อก FIFO", en: "Update FIFO allocation sequence" },
    ],
  },
];

const toolRouteMap = new Map<string, ErpToolConfig>(
  ERP_TOOL_CONFIGS.map((config) => [config.route, config]),
);

export function isErpToolsRoute(route: string): boolean {
  const clean = route.split("?")[0];
  return toolRouteMap.has(clean);
}

export function getErpToolConfig(route: string): ErpToolConfig | undefined {
  const clean = route.split("?")[0];
  return toolRouteMap.get(clean);
}

// --- การประมวลผลจริงผ่าน backend (ไม่มีข้อมูลจำลอง) ---

export interface ErpToolStockCheckStats {
  totalrows: number;
  totaldocuments: number;
  totalproducts: number;
  earliestdate: string;
  latestdate: string;
}

export interface ErpToolRunResult {
  success: boolean;
  messageKey: string;
  stats?: ErpToolStockCheckStats;
}

// เครื่องมือที่มี API จริงรองรับแล้วเท่านั้น (ตัวอื่นยังไม่มี endpoint ห้ามแสร้งว่าทำงานสำเร็จ)
const TOOL_ENDPOINTS: Record<string, string> = {
  rebuild_product_balance: "/api/goapi/api/process/product-balance",
  audit_data: "/api/goapi/api/stockcost/check",
};

export function isErpToolApiReady(code: string): boolean {
  return code in TOOL_ENDPOINTS;
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

export async function runErpTool(params: {
  code: string;
  holdingcode: string;
  businesscode: string;
  year: number;
}): Promise<ErpToolRunResult> {
  const { code, holdingcode, businesscode, year } = params;
  const endpoint = TOOL_ENDPOINTS[code];
  if (!endpoint) return { success: false, messageKey: "tool_not_available" };
  if (!holdingcode) return { success: false, messageKey: "holding_required" };

  const body: Record<string, unknown> =
    code === "audit_data"
      ? { holdingcode, businesscode, fromdate: `${year}-01-01`, todate: `${year}-12-31` }
      : { holdingcode, businesscode };

  try {
    const res = await authFetch(endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      return {
        success: false,
        messageKey: res.status === 401 || res.status === 403 ? "unauthorized" : "process_failed",
      };
    }
    const payload: unknown = await res.json();
    if (!isRecord(payload)) return { success: false, messageKey: "process_failed" };
    if (payload.success === false || payload.status === "error") {
      return { success: false, messageKey: "process_failed" };
    }
    if (code !== "audit_data") return { success: true, messageKey: "process_success" };
    return {
      success: true,
      messageKey: "process_success",
      stats: {
        totalrows: toNumber(payload.total_rows),
        totaldocuments: toNumber(payload.total_documents),
        totalproducts: toNumber(payload.total_products),
        earliestdate: toText(payload.earliest_date),
        latestdate: toText(payload.latest_date),
      },
    };
  } catch {
    return { success: false, messageKey: "connection_error" };
  }
}

// 2026-09-16: every user-visible string above also lives in languages.tsv,
// keyed by `<code>.<part>`. The literals stay as the offline fallback.
const catalogKeys: Record<string, string> = {
  "gl_reprocess.title": "gl_reprocess",
  "gl_reprocess.description": "tool_reprocess_general_ledger_journals_ledger",
  "gl_reprocess.actionLabel": "tool_start_gl_reprocessing",
  "gl_reprocess.step.0": "tool_audit_debit_credit_balance_in",
  "gl_reprocess.step.1": "tool_clear_temporary_ledger_summary_projections",
  "gl_reprocess.step.2": "tool_recalculate_period_balances_and_cumulative",
  "gl_reprocess.step.3": "tool_update_trial_balance_and_financial",
  "ar_recalculate.title": "ar_recalculate",
  "ar_recalculate.description": "tool_recalculate_outstanding_debtor_balances_from",
  "ar_recalculate.actionLabel": "tool_recalculate_ar_balances",
  "ar_recalculate.step.0": "tool_aggregate_unpaid_sales_invoices",
  "ar_recalculate.step.1": "tool_match_payments_and_credit_note",
  "ar_recalculate.step.2": "tool_update_net_balance_and_customer",
  "ar_bill_balances.title": "ap_bill_recalculate",
  "ar_bill_balances.description": "tool_reconcile_individual_invoice_outstanding_amounts",
  "ar_bill_balances.actionLabel": "tool_recompute_bill_balances",
  "ar_bill_balances.step.0": "tool_verify_itemized_bill_allocations",
  "ar_bill_balances.step.1": "tool_update_outstanding_balance_per_bill",
  "ap_recalculate.title": "ap_recalculate",
  "ap_recalculate.description": "tool_recalculate_outstanding_vendor_payables_from",
  "ap_recalculate.actionLabel": "tool_recalculate_ap_balances",
  "ap_recalculate.step.0": "tool_aggregate_purchase_invoices_and_accrued",
  "ap_recalculate.step.1": "tool_reconcile_payment_vouchers_and_settlements",
  "ap_recalculate.step.2": "tool_update_net_vendor_balance",
  "ap_bill_balances.title": "ap_bill_recalculate",
  "ap_bill_balances.description": "tool_recompute_outstanding_balance_for_each",
  "ap_bill_balances.actionLabel": "tool_recompute_vendor_bill_balances",
  "ap_bill_balances.step.0": "tool_inspect_payment_history_per_purchase",
  "ap_bill_balances.step.1": "tool_update_remaining_payable_amount_per",
  "cheque_balances.title": "cheque_recalculate",
  "cheque_balances.description": "tool_synchronize_on_hand_cleared_returned",
  "cheque_balances.actionLabel": "tool_recalculate_cheques",
  "cheque_balances.step.0": "tool_verify_status_of_received_and",
  "cheque_balances.step.1": "tool_update_pending_cheque_totals",
  "bank_balances.title": "bank_recalculate",
  "bank_balances.description": "tool_recalculate_ledger_balances_for_all",
  "bank_balances.actionLabel": "tool_recalculate_bank_books",
  "bank_balances.step.0": "tool_sort_transactions_chronologically",
  "bank_balances.step.1": "tool_compute_rolling_balance_for_each",
  "bank_balances.step.2": "tool_reconcile_with_passbook_and_bank",
  "rebuild_stock.title": "reprocess",
  "rebuild_stock.description": "tool_reconstruct_physical_inventory_on_hand",
  "rebuild_stock.actionLabel": "tool_start_full_stock_rebuild",
  "rebuild_stock.step.0": "tool_audit_warehouse_document_sequence",
  "rebuild_stock.step.1": "tool_recompute_in_out_movements_chronologically",
  "rebuild_stock.step.2": "tool_recompute_unit_costs_and_inventory",
  "rebuild_stock.step.3": "tool_update_balance_summary_tables",
  "audit_data.title": "audit_data",
  "audit_data.description": "tool_scan_for_anomalies_negative_stocks",
  "audit_data.actionLabel": "tool_run_integrity_audit",
  "audit_data.step.0": "tool_scan_for_negative_inventory_balances",
  "audit_data.step.1": "tool_check_for_empty_line_item",
  "audit_data.step.2": "tool_validate_branch_and_warehouse_codes",
  "rebuild_products.title": "rebuild_products",
  "rebuild_products.description": "tool_rebuild_search_index_barcode_catalogs",
  "rebuild_products.actionLabel": "tool_rebuild_product_index",
  "rebuild_products.step.0": "tool_collect_all_items_and_barcodes",
  "rebuild_products.step.1": "tool_generate_multilingual_full_text_search",
  "rebuild_product_balance.title": "rebuild_product_balance",
  "rebuild_product_balance.description": "tool_synchronize_item_level_and_warehouse",
  "rebuild_product_balance.actionLabel": "rebuild_product_balance",
  "rebuild_product_balance.step.0": "tool_scan_on_hand_quantities_across",
  "rebuild_product_balance.step.1": "tool_update_balance_snapshot_cache",
  "inventory_daily_sequence.title": "stock_daily_sequence",
  "inventory_daily_sequence.description": "tool_order_daily_receipts_and_issues",
  "inventory_daily_sequence.actionLabel": "tool_sort_daily_transactions",
  "inventory_daily_sequence.step.0": "tool_sequence_receipts_before_issues_within",
  "inventory_daily_sequence.step.1": "tool_update_fifo_allocation_sequence",
};

export function toolText(
  config: ErpToolConfig,
  part: "title" | "description" | "actionLabel" | "warningMessage",
  language: LanguageCode,
  dictionary?: BackendLanguageDictionary,
): string {
  return catalogText(catalogKeys, `${config.code}.${part}`, config[part], language, dictionary);
}

export function toolStepText(
  config: ErpToolConfig,
  index: number,
  language: LanguageCode,
  dictionary?: BackendLanguageDictionary,
): string {
  return catalogText(
    catalogKeys,
    `${config.code}.step.${index}`,
    config.steps[index],
    language,
    dictionary,
  );
}
