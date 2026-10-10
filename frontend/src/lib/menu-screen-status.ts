import { getSystemSettingConfig } from "./system-setting-screens";
import { isGeneralLedgerRoute } from "./general-ledger";
import { isErpTransactionRoute } from "./erp-transaction";
import { isThaiTaxRoute } from "./thai-tax";
import { isErpReportRoute, getErpReportConfig, isErpReportApiReady } from "./erp-reports";
import { isErpToolsRoute, getErpToolConfig, isErpToolApiReady } from "./erp-tools";
import { isOperationsRoute, getOperationsConfig, isApprovalApiReady } from "./erp-operations";
import { getThaiTaxConfig } from "./thai-tax";
import { taxFormRouteCode } from "./tax-forms";

// Keep aligned with the explicit WorkTabPanel branches (checked by the test).
export const CUSTOM_MENU_SCREEN_ROUTES = [
  "/shortcuts", "/product",
  "/productbarcode", "/productbarcodeshelf", "/pricehistory",
  "/datamodelgraph", "/inventory/product-sets", "/productset",
  "/report/vatsummary", "/customer/label-print",
  "/ar/billing-note-auto-print",
  "/settings/font", "/banking/contacts",
  "/settings/font-report", "/settings/font-report-title",
  "/settings/font-report-column", "/settings/font-report-detail", "/settings/font-report-company",
  "/banking/petty-cash-limit",
  "/banking/cheque-issued-prepare",
  "/transaction/receivedeposit",
  "/ar/billing-note-auto",
  "/tools/transaction-lock",
  "/ap/payment-prepare",
  "/banking/cheque-movement",
  "/banking/bank-service",
  "/sale/reservation",
  "/sale/dispatch",
  "/inventory/stock-count-handheld",
  "/inventory/product-supplier-update",
  "/settings/journal-format",
  "/settings/work-date",
  "/settings/product-by-warehouse",
  "/settings/default-warehouse",
  "/settings/currency",
  "/settings/payment-terms",
  "/settings/credit-debit-reasons",
  "/settings/main-map",
  "/settings/customer-inquiry",
  "/settings/product-inquiry",
  "/settings/about",
  "/settings/standard-software",
  "/transaction/etax",
] as const;

export const FIXED_ASSET_ROUTES = [
  "/asset/registry", "/asset/depreciation",
  "/asset/cip", "/asset/maintenance",
  "/asset/types", "/asset/post-gl", "/report/assetschedule",
  "/report/depreciationmonthly", "/report/depreciationyearly",
  "/report/depreciationpnd50", "/report/assetdisposal",
] as const;

export function isFixedAssetRoute(route: string): boolean {
  const clean = route.split("?")[0];
  return clean.startsWith("/asset/") || clean === "/report/assetschedule" ||
    clean === "/report/depreciationmonthly" || clean === "/report/depreciationyearly" ||
    clean === "/report/depreciationpnd50" || clean === "/report/assetdisposal";
}

const customRoutes = new Set<string>(CUSTOM_MENU_SCREEN_ROUTES);

/** A connected screen is not a guarantee that its business workflow is complete. */
// 2026-09-23: ระบบใช้ PostgreSQL อย่างเดียว (ถอด MongoDB/Kafka/Redis/ClickHouse) — จอเหล่านี้เคยอ่าน/เขียนผ่าน
// API บน MongoDB ซึ่งถูกลบไปแล้ว จึงเป็น "ยังไม่พร้อม" จนกว่าจะสร้าง API บน PostgreSQL ให้ (แล้วค่อยเอาออกจากรายการนี้)
const RETIRED_BACKEND_ROUTES = new Set<string>([
  "/line-oa", "/product", "/productbarcode", "/productbarcodeshelf",
  "/pricehistory", "/inventory/product-sets", "/productset",
  // เมนู จัดการระบบ › ที่เก็บสินค้าตามคลัง / กลุ่มลูกหนี้ (2026-10-09) — config ตั้งค่ายังอยู่แต่ API เดิมถูกถอดแล้ว
  "/productwarehousescreen", "/debtorgroup",
]);

// ข้อมูลตั้งค่าที่ backend PostgreSQL ยังให้บริการ
const POSTGRES_SETTING_BASE_PATHS = new Set<string>([
  "/organization/business-type", "/organization/branch",
  "/holding/employee", "/holding/permission", "/organization/role-permission",
]);

/** จอที่ไม่มี API จริงรองรับเลย — WorkTabPanel ต้องแสดงการ์ด "ยังไม่พร้อม" แทนจอที่ยิง API ที่ไม่มีอยู่ */
export function isRetiredBackendScreen(route: string): boolean {
  return RETIRED_BACKEND_ROUTES.has(route.split("?")[0]);
}

// จอเอกสารธุรกรรม ERP ที่ยังมี config (ERP_MODULE_CONFIGS) ทำงานจริงกับตาราง doc ผ่าน /transaction/:module
// จึงไม่ใช่ "retired"; เมนูที่ไม่มี config จะตกเป็น pending ใน isMenuScreenPending เอง
export function isMenuBackendRetired(route: string): boolean {
  const clean = route.split("?")[0];
  if (isRetiredBackendScreen(clean)) return true;
  const op = getOperationsConfig(clean);
  if (op) {
    if (op.code === "delivery_dates" || op.code === "purchase_receipt_dates" || op.code === "purchase_receipt_adjustment") return false;
    return !isApprovalApiReady(op.code);
  }
  const setting = getSystemSettingConfig(clean);
  return Boolean(setting?.basePath && !POSTGRES_SETTING_BASE_PATHS.has(setting.basePath));
}

export function isMenuScreenPending(route: string): boolean {
  return (
    isMenuBackendRetired(route) ||
    (!customRoutes.has(route) &&
      !isGeneralLedgerRoute(route) &&
      !isFixedAssetRoute(route) &&
      !isErpTransactionRoute(route) &&
      !isThaiTaxRoute(route) &&
      taxFormRouteCode(route) === undefined &&
      !isErpReportRoute(route) &&
      !isErpToolsRoute(route) &&
      !isOperationsRoute(route) &&
      !getSystemSettingConfig(route))
  );
}


// แบบภาษีที่มี API จริงแล้ว
// 2026-09-23: ภาษีหัก ณ ที่จ่ายอ่านจากบัญชีแยกประเภท (gl_lines) จริง; ภาษีซื้อ/ขาย/ภ.พ.30 อ่านจากรายละเอียดภาษีมูลค่าเพิ่ม
// ของใบสำคัญ GL ที่ผ่านบัญชี (details.vats); ภ.พ.36 ยังไม่มีข้อมูลต้นทาง → "รอเชื่อมข้อมูล"
const LIVE_TAX_FORMS = new Set(["vat_sale", "vat_buy", "50twi", "wht_received", "wht_summary"]);

/**
 * จอเปิดใช้งานแล้ว แต่ยังไม่มี API จริงป้อนข้อมูลให้
 * (ต่างจาก isMenuScreenPending ที่แปลว่า "ยังไม่มีจอเลย")
 */
export function isMenuDataPending(route: string): boolean {
  const clean = route.split("?")[0];

  const report = getErpReportConfig(clean);
  if (report) return !isErpReportApiReady(report.code);

  const tool = getErpToolConfig(clean);
  if (tool) return !isErpToolApiReady(tool.code);

  const operation = getOperationsConfig(clean);
  if (operation) return !isApprovalApiReady(operation.code);

  const tax = getThaiTaxConfig(clean);
  if (tax) return !LIVE_TAX_FORMS.has(tax.formType);

  return false;
}
