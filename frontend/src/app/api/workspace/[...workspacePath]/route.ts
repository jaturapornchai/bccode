import { NextResponse } from "next/server";
import {
  extractMessage,
  getBackendUrlFromRequest,
  getArray,
  getMainApiUrl,
  isRecord,
  proxyMainApiJson,
  readJsonOrText,
  requireBearerToken,
  type ApiProxyBody,
} from "@/lib/workspace-api";

type WorkspaceProxyContext = {
  params: Promise<{ workspacePath: string[] }>;
};

type ProductUnitName = {
  code: string;
  name: string;
  isauto?: boolean;
  isdelete?: boolean;
};

type ProductUnit = {
  unitcode: string;
  names: ProductUnitName[];
};

type ShopName = {
  code?: string;
  name?: string;
  isauto?: boolean;
  isdelete?: boolean;
};

type MainApiResult = {
  ok: boolean;
  status: number;
  payload: unknown;
};

const UNIT_TEMPLATE_URL = "https://raw.githubusercontent.com/smlsoft/dedepos_template/main/unit.json";
const API_TIMEOUT_MS = 20000;

export async function GET(request: Request, context: WorkspaceProxyContext) {
  const { workspacePath } = await context.params;
  const path = workspacePath.join("/");

  let mainApiUrl: string;
  try {
    mainApiUrl = getMainApiUrl(getBackendUrlFromRequest(request));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const url = new URL(request.url);
  switch (path) {
    case "shops":
      return listShopsWithDisplayNames(request, mainApiUrl);
    case "shop-info": {
      const shopid = url.searchParams.get("shopid")?.trim() ?? "";
      if (!shopid) return NextResponse.json({ success: false, message: "ไม่พบรหัสบริษัท" }, { status: 400 });
      return proxyMainApiJson(request, mainApiUrl, `/shop/${encodeURIComponent(shopid)}`, { method: "GET" });
    }
    case "branches": {
      const offset = url.searchParams.get("offset") ?? "0";
      const limit = url.searchParams.get("limit") ?? "100";
      const query = url.searchParams.get("q") ?? "";
      const branchPath = `/organization/branch/list?offset=${encodeURIComponent(offset)}&limit=${encodeURIComponent(limit)}&q=${encodeURIComponent(query)}`;
      return proxyMainApiJson(request, mainApiUrl, branchPath, { method: "GET" });
    }
    case "product-units": {
      const offset = url.searchParams.get("offset") ?? "0";
      const limit = url.searchParams.get("limit") ?? "1";
      const query = url.searchParams.get("q") ?? "";
      const unitPath = `/unit/list?offset=${encodeURIComponent(offset)}&limit=${encodeURIComponent(limit)}&q=${encodeURIComponent(query)}&sort=unitcode:1`;
      return proxyMainApiJson(request, mainApiUrl, unitPath, { method: "GET" });
    }
    case "product-units/standard": {
      const mainShopId = url.searchParams.get("mainShopId")?.trim() ?? url.searchParams.get("main_shop_id")?.trim() ?? "";
      const query = url.searchParams.get("q")?.trim() ?? "";
      return listMissingStandardProductUnits(request, mainApiUrl, mainShopId, query);
    }
    default:
      return NextResponse.json({ success: false, message: "ไม่พบ workspace endpoint" }, { status: 404 });
  }
}

export async function POST(request: Request, context: WorkspaceProxyContext) {
  const { workspacePath } = await context.params;
  const path = workspacePath.join("/");

  let body: ApiProxyBody;
  try {
    body = (await request.json()) as ApiProxyBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  let mainApiUrl: string;
  try {
    mainApiUrl = getMainApiUrl(getBackendUrlFromRequest(request, body));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const { backendUrl: _backendUrl, ...payload } = body;
  void _backendUrl;

  switch (path) {
    case "select-shop": {
      const shopid = typeof payload.shopid === "string" ? payload.shopid.trim() : "";
      if (!shopid) return NextResponse.json({ success: false, message: "ไม่พบรหัสบริษัท" }, { status: 400 });
      return proxyMainApiJson(request, mainApiUrl, "/select-shop", {
        method: "POST",
        body: JSON.stringify({ shopid }),
      });
    }
    case "branch": {
      const branch = payload.branch;
      if (!branch || typeof branch !== "object") {
        return NextResponse.json({ success: false, message: "ไม่พบข้อมูลสาขา" }, { status: 400 });
      }
      return proxyMainApiJson(request, mainApiUrl, "/organization/branch", {
        method: "POST",
        body: JSON.stringify(branch),
      });
    }
    case "create-shop":
      return proxyMainApiJson(request, mainApiUrl, "/create-shop", {
        method: "POST",
        body: JSON.stringify(payload),
      });
    case "product-units/defaults": {
      const mainShopId = getPayloadString(payload, "mainShopId") ?? getPayloadString(payload, "main_shop_id") ?? getPayloadString(payload, "mainshopid") ?? "";
      return createDefaultProductUnits(request, mainApiUrl, mainShopId, getPayloadStringArray(payload, "unitcodes"));
    }
    default:
      return NextResponse.json({ success: false, message: "ไม่พบ workspace endpoint" }, { status: 404 });
  }
}

async function listMissingStandardProductUnits(request: Request, mainApiUrl: string, mainShopId: string, query: string): Promise<NextResponse> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  try {
    const existing = await loadExistingUnitCodes(request, mainApiUrl, authorization);
    if (!existing.ok) return mainApiError(existing.result, "ตรวจสอบหน่วยนับสินค้าไม่สำเร็จ");

    const source = await loadStandardUnits(request, mainApiUrl, authorization, mainShopId);
    const units = filterMissingUnits(source.units, existing.codes, query);

    return NextResponse.json({
      success: true,
      data: units,
      total: units.length,
      source: source.source,
    });
  } catch (error) {
    return NextResponse.json({ success: false, message: productUnitErrorMessage(error) }, { status: 504 });
  }
}

async function listShopsWithDisplayNames(request: Request, mainApiUrl: string): Promise<NextResponse> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  try {
    const result = await callMainApiJson(request, mainApiUrl, "/list-shop?limit=100", { method: "GET" }, authorization);
    if (!result.ok || isApiFailure(result.payload)) return mainApiError(result, "โหลดบริษัทไม่สำเร็จ");

    const shops = getArrayFromPayload(result.payload, "data");
    const enriched = await Promise.all(shops.map((shop) => enrichShopDisplayName(request, mainApiUrl, authorization, shop)));

    if (isRecord(result.payload)) {
      return NextResponse.json({ ...result.payload, data: enriched }, { status: result.status });
    }

    return NextResponse.json({ success: true, data: enriched, total: enriched.length }, { status: result.status });
  } catch (error) {
    return NextResponse.json({ success: false, message: workspaceErrorMessage(error, "โหลดบริษัทไม่สำเร็จ") }, { status: 504 });
  }
}

async function enrichShopDisplayName(request: Request, mainApiUrl: string, authorization: string, shop: unknown): Promise<unknown> {
  if (!isRecord(shop) || hasShopDisplayName(shop)) return shop;

  const shopid = getPayloadString(shop, "shopid")?.trim();
  if (!shopid) return shop;

  const result = await callMainApiJson(request, mainApiUrl, `/shop/${encodeURIComponent(shopid)}`, { method: "GET" }, authorization);
  if (!result.ok || isApiFailure(result.payload)) return shop;

  const shopInfo = payloadDataRecord(result.payload);
  if (!shopInfo) return shop;

  const names = getArray(shopInfo, "names")
    .map((name) => normalizeShopName(name))
    .filter((name): name is ShopName => Boolean(name));
  const name = firstPayloadString(shopInfo, ["name1", "companyname", "company_name", "name"]);

  return {
    ...shop,
    ...(name ? { name, name1: name } : {}),
    ...(names.length > 0 ? { names } : {}),
  };
}

async function createDefaultProductUnits(request: Request, mainApiUrl: string, mainShopId: string, selectedCodes: string[]): Promise<NextResponse> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  try {
    const existing = await loadExistingUnitCodes(request, mainApiUrl, authorization);
    if (!existing.ok) return mainApiError(existing.result, "ตรวจสอบหน่วยนับสินค้าไม่สำเร็จ");

    const source = await loadStandardUnits(request, mainApiUrl, authorization, mainShopId);
    const units = filterMissingUnits(source.units, existing.codes, "", selectedCodes);

    if (units.length === 0) {
      return NextResponse.json({ success: true, message: "ไม่มีหน่วยนับใหม่ที่ต้องเพิ่ม", data: { count: 0, source: source.source } });
    }

    const saved = await callMainApiJson(
      request,
      mainApiUrl,
      "/unit/bulk",
      {
        method: "POST",
        body: JSON.stringify(units),
      },
      authorization,
    );

    if (!saved.ok || isApiFailure(saved.payload)) return mainApiError(saved, "เพิ่มหน่วยนับสินค้าไม่สำเร็จ");

    return NextResponse.json({
      success: true,
      message: "เพิ่มหน่วยนับสินค้าเริ่มต้นแล้ว",
      data: { count: units.length, source: source.source },
    });
  } catch (error) {
    return NextResponse.json({ success: false, message: productUnitErrorMessage(error) }, { status: 504 });
  }
}

async function loadExistingUnitCodes(
  request: Request,
  mainApiUrl: string,
  authorization: string,
): Promise<{ ok: true; codes: Set<string> } | { ok: false; result: MainApiResult }> {
  const result = await callMainApiJson(request, mainApiUrl, "/unit/list?offset=0&limit=10000&q=&sort=unitcode:1", { method: "GET" }, authorization);
  if (!result.ok) return { ok: false, result };

  const codes = new Set<string>();
  for (const item of getArrayFromPayload(result.payload, "data")) {
    if (!isRecord(item)) continue;
    const unitcode = getPayloadString(item, "unitcode")?.trim();
    if (unitcode) codes.add(normalizeUnitCode(unitcode));
  }
  return { ok: true, codes };
}

async function loadStandardUnits(
  request: Request,
  mainApiUrl: string,
  authorization: string,
  mainShopId: string,
): Promise<{ source: string; units: ProductUnit[] }> {
  if (mainShopId) {
    const mainShopUnits = await callMainApiJson(
      request,
      mainApiUrl,
      `/unit?limit=1000&shopsid=${encodeURIComponent(mainShopId)}`,
      { method: "GET" },
      authorization,
    );
    if (mainShopUnits.ok) {
      const units = normalizeUnitList(getArrayFromPayload(mainShopUnits.payload, "data"));
      if (units.length > 0) return { source: "main-shop", units };
    }
  }

  return { source: "template", units: await loadTemplateUnits() };
}

function filterMissingUnits(units: ProductUnit[], existingCodes: Set<string>, query: string, selectedCodes: string[] = []): ProductUnit[] {
  const needle = query.trim().toLowerCase();
  const selected = new Set(selectedCodes.map(normalizeUnitCode));

  return units.filter((unit) => {
    const normalizedCode = normalizeUnitCode(unit.unitcode);
    if (existingCodes.has(normalizedCode)) return false;
    if (selected.size > 0 && !selected.has(normalizedCode)) return false;
    if (!needle) return true;
    const haystack = `${unit.unitcode} ${unit.names.map((name) => name.name).join(" ")}`.toLowerCase();
    return haystack.includes(needle);
  });
}

async function callMainApiJson(
  request: Request,
  mainApiUrl: string,
  path: string,
  init: RequestInit,
  authorization: string,
): Promise<MainApiResult> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), API_TIMEOUT_MS);

  try {
    const response = await fetch(`${mainApiUrl}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
        ...(init.headers ?? {}),
      },
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);
    return { ok: response.ok, status: response.status, payload };
  } finally {
    clearTimeout(timeout);
  }
}

async function loadTemplateUnits(): Promise<ProductUnit[]> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), API_TIMEOUT_MS);

  try {
    const response = await fetch(UNIT_TEMPLATE_URL, {
      headers: { Accept: "application/json" },
      signal: controller.signal,
      cache: "no-store",
    });

    if (!response.ok) throw new Error("โหลด template หน่วยนับสินค้าไม่สำเร็จ");
    const payload = await response.json();
    return normalizeUnitList(Array.isArray(payload) ? payload : []);
  } finally {
    clearTimeout(timeout);
  }
}

function normalizeUnitList(value: unknown[]): ProductUnit[] {
  const units: ProductUnit[] = [];
  const seen = new Set<string>();

  for (const item of value) {
    const unit = normalizeUnit(item);
    if (!unit || seen.has(unit.unitcode)) continue;
    seen.add(unit.unitcode);
    units.push(unit);
  }

  return units;
}

function normalizeUnit(value: unknown): ProductUnit | null {
  if (!isRecord(value)) return null;
  const unitcode = getPayloadString(value, "unitcode")?.trim();
  const names = getArray(value, "names")
    .map((name) => normalizeUnitName(name))
    .filter((name): name is ProductUnitName => Boolean(name));

  if (!unitcode || names.length === 0) return null;
  return { unitcode, names };
}

function normalizeUnitName(value: unknown): ProductUnitName | null {
  if (!isRecord(value)) return null;
  const code = getPayloadString(value, "code")?.trim();
  if (!code) return null;
  return {
    code,
    name: getPayloadString(value, "name")?.trim() ?? "",
    isauto: false,
    isdelete: false,
  };
}

function normalizeShopName(value: unknown): ShopName | null {
  if (!isRecord(value)) return null;
  const name = getPayloadString(value, "name")?.trim();
  if (!name) return null;
  return {
    code: getPayloadString(value, "code")?.trim(),
    name,
    isauto: value.isauto === true,
    isdelete: value.isdelete === true,
  };
}

function getPayloadString(payload: Record<string, unknown>, key: string): string | undefined {
  const value = payload[key];
  return typeof value === "string" ? value : undefined;
}

function firstPayloadString(payload: Record<string, unknown>, keys: string[]): string | undefined {
  for (const key of keys) {
    const value = getPayloadString(payload, key)?.trim();
    if (value) return value;
  }
  return undefined;
}

function getPayloadStringArray(payload: Record<string, unknown>, key: string): string[] {
  const value = payload[key];
  if (!Array.isArray(value)) return [];
  return value.map((item) => stringFromUnknown(item)).filter(Boolean);
}

function getArrayFromPayload(payload: unknown, key: string): unknown[] {
  return isRecord(payload) ? getArray(payload, key) : [];
}

function isApiFailure(payload: unknown): boolean {
  return isRecord(payload) && payload.success === false;
}

function payloadDataRecord(payload: unknown): Record<string, unknown> | null {
  if (!isRecord(payload)) return null;
  return isRecord(payload.data) ? payload.data : payload;
}

function hasShopDisplayName(shop: Record<string, unknown>): boolean {
  const shopid = getPayloadString(shop, "shopid")?.trim();
  const directName = firstPayloadString(shop, ["name1", "companyname", "company_name", "name"]);
  if (directName && directName !== shopid) return true;
  return getArray(shop, "names").some((name) => {
    const displayName = isRecord(name) ? getPayloadString(name, "name")?.trim() : "";
    return Boolean(displayName && displayName !== shopid);
  });
}

function mainApiError(result: MainApiResult, fallback: string): NextResponse {
  return NextResponse.json(
    { success: false, message: extractMessage(result.payload) ?? fallback },
    { status: result.status >= 400 ? result.status : 400 },
  );
}

function normalizeUnitCode(value: string): string {
  return value.trim().toUpperCase();
}

function stringFromUnknown(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function productUnitErrorMessage(error: unknown): string {
  return error instanceof Error && error.name === "AbortError"
    ? "Server ไม่ตอบกลับทันเวลา"
    : error instanceof Error && error.message
      ? error.message
      : "ไม่สามารถจัดการหน่วยนับสินค้าเริ่มต้นได้";
}

function workspaceErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error && error.name === "AbortError"
    ? "Server ไม่ตอบกลับทันเวลา"
    : error instanceof Error && error.message
      ? error.message
      : fallback;
}
