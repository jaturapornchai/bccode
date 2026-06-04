import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  isRecord,
  proxyMainApiJson,
} from "@/lib/workspace-api";

/**
 * Master-data picker proxy.
 *
 * Frontend: GET /api/product-barcode/master/{master}?q=&page=&limit=&lang=
 * Backend:  GET <mainapi>/{masterPath} where masterPath depends on `master` name.
 *
 * Whitelist of accepted master names → backend path. Anything else returns 404
 * so we don't accidentally proxy arbitrary mainapi endpoints.
 */
const MASTER_PATHS: Record<string, string> = {
  group: "/product/group",
  groupsubone: "/aicloud/groupsubone",
  groupsubtwo: "/aicloud/groupsubtwo",
  brand: "/aicloud/brand",
  category: "/aicloud/category",
  class: "/aicloud/class",
  design: "/aicloud/design",
  grade: "/aicloud/grade",
  model: "/aicloud/model",
  pattern: "/aicloud/pattern",
  unit: "/unit",
  producttype: "/product/type",
  ordertype: "/product/order-type",
  businesstype: "/product-section/business-type",
  branch: "/list-holding",
  company: "/organization/company",
  creditor: "/debtaccount/creditor",
  product: "/product",
};

type MasterContext = { params: Promise<{ master?: string }> };

export async function GET(request: Request, context: MasterContext) {
  const { master } = await context.params;
  const masterKey = (master ?? "").toLowerCase().trim();
  const path = MASTER_PATHS[masterKey];
  if (!path) {
    return NextResponse.json({ success: false, message: "ไม่รองรับ master นี้" }, { status: 404 });
  }

  let base: string;
  try {
    base = getMainApiUrl(getBackendUrlFromRequest(request));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const url = new URL(request.url);
  const qs = new URLSearchParams();
  for (const key of ["q", "page", "limit", "lang", "company_guid", "item_type", "materialtype"] as const) {
    const value = url.searchParams.get(key);
    if (value) qs.set(key, value);
  }
  const queryString = qs.toString();
  const proxyResponse = await proxyMainApiJson(request, base, `${path}${queryString ? `?${queryString}` : ""}`, {
    method: "GET",
  });

  // Standardize response: backend returns `{ success, data, pagination }` — flatten to MasterEntry[].
  const text = await proxyResponse.clone().text();
  try {
    const payload = JSON.parse(text) as unknown;
    if (!isRecord(payload)) return proxyResponse;
    const rawData = Array.isArray(payload.data) ? payload.data : [];
    const entries = rawData.flatMap((entry: unknown) => {
      if (!isRecord(entry)) return [];
      let guid = String(entry.guidfixed ?? entry.guid_fixed ?? "");
      let code = String(entry.code ?? entry.unitcode ?? entry.itemunitcode ?? entry.groupcode ?? entry.brand_code ?? entry.categorycode ?? "");
      if (masterKey === "branch") {
        guid = String(entry.holding_code ?? "");
        code = String(entry.branchcode && entry.branchcode !== "" ? entry.branchcode : (entry.holding_code ?? ""));
      }
      const namesSource = entry.names ?? entry.unitnames ?? entry.unit_names ?? entry.itemunitnames ?? entry.item_unit_names;
      const names = Array.isArray(namesSource) ? namesSource : [];
      return [{ guidfixed: guid, code, names }];
    });
    return NextResponse.json(
      { success: payload.success !== false, data: entries, message: payload.message ?? undefined },
      { status: proxyResponse.status },
    );
  } catch {
    return proxyResponse;
  }
}
