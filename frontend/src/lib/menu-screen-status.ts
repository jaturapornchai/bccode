import { getSystemSettingConfig } from "./system-setting-screens";
import { isGeneralLedgerRoute } from "./general-ledger";
import { isErpTransactionRoute } from "./erp-transaction";
import { isThaiTaxRoute } from "./thai-tax";
import { isErpReportRoute, getErpReportConfig, isErpReportApiReady } from "./erp-reports";
import { isErpToolsRoute, getErpToolConfig, isErpToolApiReady } from "./erp-tools";
import { isOperationsRoute, getOperationsConfig, isApprovalApiReady } from "./erp-operations";
import { getThaiTaxConfig } from "./thai-tax";

// Keep aligned with the explicit WorkTabPanel branches (checked by the test).
export const CUSTOM_MENU_SCREEN_ROUTES = [
  "/shortcuts", "/line-oa", "/product",
  "/productbarcode", "/productbarcodeshelf", "/pricehistory",
  "/datamodelgraph", "/inventory/product-sets", "/productset",
] as const;

export const FIXED_ASSET_ROUTES = [
  "/asset/registry", "/asset/depreciation", "/asset/purchase",
  "/asset/cip", "/asset/disposal", "/asset/maintenance",
  "/asset/types", "/asset/post-gl", "/report/assetschedule",
] as const;

export function isFixedAssetRoute(route: string): boolean {
  const clean = route.split("?")[0];
  return clean.startsWith("/asset/") || clean === "/report/assetschedule";
}

const customRoutes = new Set<string>(CUSTOM_MENU_SCREEN_ROUTES);

// Explicitly blocked routes (if any)
const pendingRoutes = new Set<string>([]);

/** A connected screen is not a guarantee that its business workflow is complete. */
export function isMenuScreenPending(route: string): boolean {
  return (
    pendingRoutes.has(route) ||
    (!customRoutes.has(route) &&
      !isGeneralLedgerRoute(route) &&
      !isFixedAssetRoute(route) &&
      !isErpTransactionRoute(route) &&
      !isThaiTaxRoute(route) &&
      !isErpReportRoute(route) &&
      !isErpToolsRoute(route) &&
      !isOperationsRoute(route) &&
      !getSystemSettingConfig(route))
  );
}


// แบบภาษีที่มี API จริงแล้ว (ที่เหลือยังไม่มีตารางภาษีหัก ณ ที่จ่ายใน backend)
const LIVE_TAX_FORMS = new Set(["vat_sale", "vat_buy", "pp30"]);

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
