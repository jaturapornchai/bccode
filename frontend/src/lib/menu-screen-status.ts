import { getSystemSettingConfig } from "./system-setting-screens";
import { isGeneralLedgerRoute } from "./general-ledger";

// Keep aligned with the explicit WorkTabPanel branches (checked by the test).
export const CUSTOM_MENU_SCREEN_ROUTES = [
  "/shortcuts", "/line-oa", "/product",
  "/productbarcode", "/productbarcodeshelf", "/pricehistory",
  "/datamodelgraph", "/inventory/product-sets", "/productset",
] as const;

const customRoutes = new Set<string>(CUSTOM_MENU_SCREEN_ROUTES);

// The serial registry has a UI config, but its backing endpoint returns 404
// (verified 2026-09-11). Keep navigation pending until that workflow is connected.
const pendingRoutes = new Set(["/productserialregistry", "/gl/reprocess", "/report/xbrl"]);

/** A connected screen is not a guarantee that its business workflow is complete. */
export function isMenuScreenPending(route: string): boolean {
  return pendingRoutes.has(route) || (!customRoutes.has(route) && !isGeneralLedgerRoute(route) && !getSystemSettingConfig(route));
}
