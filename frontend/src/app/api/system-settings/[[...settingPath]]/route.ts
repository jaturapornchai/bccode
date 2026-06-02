import { randomBytes } from "crypto";
import { NextResponse } from "next/server";
import { validateBackendUrl } from "@/lib/backend-url";
import { getJwtClaimShopId, verifyHs256Jwt } from "@/lib/server-jwt";
import { getSystemSettingConfig, type SystemSettingConfig } from "@/lib/system-setting-screens";
import {
  getBackendUrlFromRequest,
  extractMessage,
  getMainApiUrl,
  isRecord,
  readJsonOrText,
  requireBearerToken,
  type ApiProxyBody,
} from "@/lib/workspace-api";

type SystemSettingsProxyContext = {
  params: Promise<{ settingPath?: string[] }>;
};

type ResolvedProxy = {
  config: SystemSettingConfig;
  id: string;
};

export async function GET(request: Request, context: SystemSettingsProxyContext) {
  const resolved = await resolveProxy(context);
  if (resolved instanceof NextResponse) return resolved;

  const tenantResponse = validateTenantAccess(request);
  if (tenantResponse) return tenantResponse;

  const base = resolveBaseUrl(request, resolved.config);
  if (base instanceof NextResponse) return base;

  return proxyJson(request, base, buildGetPath(request, resolved.config, resolved.id), buildGetInit(request, resolved.config, resolved.id));
}

export async function POST(request: Request, context: SystemSettingsProxyContext) {
  const body = await readBody(request);
  if (!isRecord(body) && !Array.isArray(body)) return badPayload();

  const resolved = await resolveProxy(context);
  if (resolved instanceof NextResponse) return resolved;

  const tenantResponse = validateTenantAccess(request, Array.isArray(body) ? undefined : body);
  if (tenantResponse) return tenantResponse;

  const base = resolveBaseUrl(request, resolved.config, Array.isArray(body) ? undefined : body);
  if (base instanceof NextResponse) return base;

  if (Array.isArray(body)) {
    const path = resolved.config.basePath ?? "";
    return proxyJson(request, base, path, { method: "POST", body: JSON.stringify(body) });
  }

  const path = buildWritePath(request, resolved.config, "", "POST", body);
  const payload = buildWritePayload(request, resolved.config, "", body);
  if (resolved.config.slug === "user") {
    const ensureResponse = await ensureUsernameLoginUser(request, base, payload);
    if (ensureResponse) return ensureResponse;
  }
  return proxyJson(request, base, path, { method: resolved.config.slug === "user" ? "PUT" : "POST", body: JSON.stringify(payload) });
}

export async function PUT(request: Request, context: SystemSettingsProxyContext) {
  const body = await readBody(request);
  if (!isRecord(body) && !Array.isArray(body)) return badPayload();

  const resolved = await resolveProxy(context);
  if (resolved instanceof NextResponse) return resolved;

  const tenantResponse = validateTenantAccess(request, Array.isArray(body) ? undefined : body);
  if (tenantResponse) return tenantResponse;

  const base = resolveBaseUrl(request, resolved.config, Array.isArray(body) ? undefined : body);
  if (base instanceof NextResponse) return base;

  if (Array.isArray(body)) {
    const path = resolved.id ? `${resolved.config.basePath}/${encodeProxyPathId(resolved.config, resolved.id)}` : (resolved.config.basePath ?? "");
    return proxyJson(request, base, path, { method: "PUT", body: JSON.stringify(body) });
  }

  const path = buildWritePath(request, resolved.config, resolved.id, "PUT", body);
  const payload = buildWritePayload(request, resolved.config, resolved.id, body);
  return proxyJson(request, base, path, { method: resolved.config.kind === "atlas" || resolved.config.kind === "ai-provider" ? "POST" : "PUT", body: JSON.stringify(payload) });
}

export async function DELETE(request: Request, context: SystemSettingsProxyContext) {
  const body = await readBody(request);
  const resolved = await resolveProxy(context);
  if (resolved instanceof NextResponse) return resolved;

  const tenantResponse = validateTenantAccess(request, isRecord(body) ? body : undefined);
  if (tenantResponse) return tenantResponse;

  const base = resolveBaseUrl(request, resolved.config, isRecord(body) ? body : undefined);
  if (base instanceof NextResponse) return base;

  const path = buildDeletePath(request, resolved.config, resolved.id);
  const payload = buildDeletePayload(request, resolved.config, resolved.id, isRecord(body) ? body : {});
  return proxyJson(request, base, path, {
    method: resolved.config.kind === "atlas" || resolved.config.kind === "ai-provider" ? "POST" : "DELETE",
    body: payload === undefined ? undefined : JSON.stringify(payload),
  });
}

async function resolveProxy(context: SystemSettingsProxyContext): Promise<ResolvedProxy | NextResponse> {
  const { settingPath } = await context.params;
  const [rawSlug, ...rest] = settingPath ?? [];
  const slug = decodePathSegment(rawSlug ?? "");
  if (!slug) return NextResponse.json({ success: false, message: "ไม่พบหน้าจอตั้งค่า" }, { status: 404 });

  const config = getSystemSettingConfig(slug);
  if (!config) return NextResponse.json({ success: false, message: "หน้าจอตั้งค่านี้ยังไม่รองรับ" }, { status: 404 });

  return { config, id: rest.map(decodePathSegment).join("/") };
}

function decodePathSegment(value: string): string {
  try {
    return decodeURIComponent(value);
  } catch {
    return value;
  }
}

function resolveBaseUrl(request: Request, config: SystemSettingConfig, body?: ApiProxyBody): string | NextResponse {
  try {
    const backendUrl = getBackendUrlFromRequest(request, body);
    if (usesGoApi(config)) return validateBackendUrl(backendUrl).normalizedGoApiUrl;
    return getMainApiUrl(backendUrl);
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }
}

function usesGoApi(config: SystemSettingConfig): boolean {
  return config.kind === "ai-provider" || config.kind === "atlas" || config.kind === "copy-uat" || config.kind === "goapi-crud";
}

function validateTenantAccess(request: Request, body?: Record<string, unknown>): NextResponse | null {
  const requestedShopId = getRequestedShopId(request, body);
  if (!requestedShopId) return null;
  if (!process.env.JWT_SECRET_KEY?.trim()) return null;

  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const jwt = verifyHs256Jwt(authorization);
  if (!jwt.ok) return NextResponse.json({ success: false, message: jwt.message }, { status: jwt.status });

  const claimShopId = getJwtClaimShopId(jwt.claims);
  if (!claimShopId) return NextResponse.json({ success: false, message: "token ไม่มีรหัสบริษัท" }, { status: 401 });
  if (claimShopId !== requestedShopId) {
    return NextResponse.json({ success: false, message: "ไม่มีสิทธิ์เข้าถึงข้อมูลบริษัทนี้" }, { status: 403 });
  }
  return null;
}

function getRequestedShopId(request: Request, body?: Record<string, unknown>): string {
  const url = new URL(request.url);
  const value =
    body?.shopid ??
    body?.shop_id ??
    body?.holding_code ??
    url.searchParams.get("shopid") ??
    url.searchParams.get("shop_id") ??
    url.searchParams.get("holding_code") ??
    "";
  return String(value).trim();
}

function buildGetPath(request: Request, config: SystemSettingConfig, id: string): string {
  const url = new URL(request.url);
  const query = new URLSearchParams();
  const shopid = url.searchParams.get("shopid") ?? url.searchParams.get("shop_id") ?? "";

  if (config.kind === "company") {
    const targetShop = id || shopid;
    return `/shop/${encodeURIComponent(targetShop)}`;
  }

  if (config.kind === "restaurant-setting") {
    query.set("page", url.searchParams.get("page") ?? "1");
    query.set("limit", url.searchParams.get("limit") ?? "1000");
    const q = url.searchParams.get("q");
    if (q) query.set("q", q);
    return `/restaurant/settings/code/${encodeURIComponent(config.code ?? "")}?${query.toString()}`;
  }

  if (config.kind === "atlas") {
    return "/atlas/get";
  }

  if (config.kind === "goapi-crud") {
    query.set("shop_id", shopid);
    return `${config.basePath ?? ""}?${query.toString()}`;
  }

  if (config.kind === "ai-provider") {
    return "/api/v1/ai-provider/list";
  }

  if (config.kind === "copy-uat") {
    const sourceEnvironment = url.searchParams.get("source_environment") ?? url.searchParams.get("source_env") ?? "uat";
    return `/listsourceshops?source_environment=${encodeURIComponent(sourceEnvironment)}`;
  }

  let basePath = id ? `${config.basePath}/${encodeProxyPathId(config, id)}` : (config.listPath ?? config.basePath ?? "");
  const companyGuid = url.searchParams.get("company_guid");
  if (!id && config.slug === "branch" && companyGuid) {
    basePath = config.basePath ?? "";
  }
  forwardPagingParams(url, query);
  const queryText = query.toString();
  return queryText ? `${basePath}?${queryText}` : basePath;
}

function buildGetInit(request: Request, config: SystemSettingConfig, id = ""): RequestInit {
  const url = new URL(request.url);
  const shopid = url.searchParams.get("shopid") ?? url.searchParams.get("shop_id") ?? "";

  if (config.kind === "atlas") {
    const body: Record<string, unknown> = {
      collection: config.collection,
      shopid,
      holding_code: shopid,
      limit: Number(url.searchParams.get("limit") ?? "1000"),
      skip: Number(url.searchParams.get("offset") ?? "0"),
    };
    if (id) {
      body.guid_fixed = id;
      body.email = id;
      body.cartid = id;
      if (config.collection === "employee_permissions") body.user_uid = id;
    }
    return {
      method: "POST",
      body: JSON.stringify(body),
    };
  }

  if (config.kind === "ai-provider") {
    return { method: "POST", body: JSON.stringify({ shop_id: shopid }) };
  }

  return { method: "GET" };
}

function buildWritePath(request: Request, config: SystemSettingConfig, id: string, method: "POST" | "PUT", body: Record<string, unknown>): string {
  const url = new URL(request.url);
  const search = url.search;
  let path = "";

  if (config.kind === "company") {
    const shopid = String(body.shopid ?? url.searchParams.get("shopid") ?? id);
    path = `/shop/${encodeURIComponent(shopid)}`;
  } else if (config.kind === "restaurant-setting") {
    path = id ? `/restaurant/settings/${encodeURIComponent(id)}` : "/restaurant/settings";
  } else if (config.kind === "atlas") {
    path = "/atlas/update";
  } else if (config.kind === "goapi-crud") {
    const createWithExport = Boolean(body.createWithExport);
    if (method === "POST" && createWithExport) path = "/api/mcp/keys/create-with-export";
    else path = id ? `${config.basePath}/${encodeProxyPathId(config, id)}` : (config.basePath ?? "");
  } else if (config.kind === "ai-provider") {
    path = "/api/v1/ai-provider/save";
  } else if (config.kind === "copy-uat") {
    path = body.action === "copy" ? "/copymongouattodev" : "/previewcopymongo";
  } else if (config.slug === "user") {
    path = "/shop/permission";
  } else {
    path = id ? `${config.basePath}/${encodeProxyPathId(config, id)}` : (config.basePath ?? "");
  }

  return search ? `${path}${search}` : path;
}

function buildDeletePath(request: Request, config: SystemSettingConfig, id: string): string {
  const url = new URL(request.url);
  const search = url.search;
  let path = "";

  if (config.kind === "atlas") path = "/atlas/delete";
  else if (config.kind === "ai-provider") path = "/api/v1/ai-provider/delete";
  else if (config.kind === "goapi-crud") path = `${config.basePath}/${encodeProxyPathId(config, id)}`;
  else if (config.slug === "user") path = `/shop/permission/${encodeProxyPathId(config, id)}`;
  else path = `${config.basePath}/${encodeProxyPathId(config, id)}`;

  return search ? `${path}${search}` : path;
}

function buildWritePayload(request: Request, config: SystemSettingConfig, id: string, body: Record<string, unknown>): Record<string, unknown> {
  const payload = stripProxyKeys(body);
  const url = new URL(request.url);
  const shopid = String(payload.shopid ?? payload.shop_id ?? payload.holding_code ?? url.searchParams.get("shopid") ?? url.searchParams.get("holding_code") ?? "");

  if (config.kind === "restaurant-setting") {
    const { guid_fixed: _guidfixed, ...rest } = payload;
    void _guidfixed;
    return {
      code: config.code,
      body: typeof payload.body === "string" ? payload.body : JSON.stringify(rest),
    };
  }

  if (config.kind === "atlas") {
    const key = String(payload.guid_fixed ?? payload[config.idField ?? ""] ?? id);
    const legacyKey = id && id !== key ? id : "";
    const {
      _id: _mongoId,
      shopid: _legacyShopID,
      shop_id: _legacyShopIDAlt,
      updatedAt: _legacyUpdatedAt,
      updatedBy: _legacyUpdatedBy,
      createdAt: _legacyCreatedAt,
      createdBy: _legacyCreatedBy,
      ...atlasData
    } = payload;
    void _mongoId;
    void _legacyShopID;
    void _legacyShopIDAlt;
    void _legacyUpdatedAt;
    void _legacyUpdatedBy;
    void _legacyCreatedAt;
    void _legacyCreatedBy;
    return {
      collection: config.collection,
      shopid,
      holding_code: String(payload.holding_code ?? shopid),
      guid_fixed: key,
      email: legacyKey,
      cartid: legacyKey,
      user_uid: typeof payload.user_uid === "string" ? payload.user_uid : undefined,
      data: { ...atlasData, holding_code: String(payload.holding_code ?? shopid), guid_fixed: key },
      upsert: true,
    };
  }

  if (config.kind === "goapi-crud") {
    return {
      ...payload,
      shop_id: shopid,
    };
  }

  if (config.kind === "ai-provider") {
    return {
      ...payload,
      shop_id: shopid,
    };
  }

  if (config.kind === "copy-uat") {
    return payload;
  }

  return payload;
}

function buildDeletePayload(
  request: Request,
  config: SystemSettingConfig,
  id: string,
  body: Record<string, unknown>,
): Record<string, unknown> | undefined {
  const url = new URL(request.url);
  const shopid = String(body.shopid ?? body.shop_id ?? body.holding_code ?? url.searchParams.get("shopid") ?? url.searchParams.get("holding_code") ?? "");

  if (config.kind === "atlas") {
    return {
      collection: config.collection,
      shopid,
      holding_code: String(body.holding_code ?? shopid),
      guid_fixed: id,
      email: id,
      cartid: id,
      user_uid: typeof body.user_uid === "string" ? body.user_uid : undefined,
      delete_many: false,
    };
  }

  if (config.kind === "ai-provider") {
    return {
      shop_id: shopid,
      provider_name: id,
    };
  }

  return undefined;
}

function encodeProxyPathId(config: SystemSettingConfig, id: string): string {
  if (config.slug === "user") {
    return encodeURIComponent(id).replace(/%40/gi, "@").replace(/%2B/gi, "+");
  }
  return encodeURIComponent(id);
}

async function ensureUsernameLoginUser(request: Request, baseUrl: string, payload: Record<string, unknown>): Promise<NextResponse | null> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const username = String(payload.username ?? "").trim();
  if (!username) return NextResponse.json({ success: false, message: "username invalid" }, { status: 400 });

  const existsResponse = await fetch(`${baseUrl}/register/exists-username`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Accept-Language": request.headers.get("accept-language") ?? "th" },
    body: JSON.stringify({ username }),
    cache: "no-store",
  });
  const existsPayload = await readJsonOrText(existsResponse);
  if (!existsResponse.ok || !isRecord(existsPayload) || existsPayload.success === false) {
    return NextResponse.json({ success: false, message: extractMessage(existsPayload) ?? "ตรวจสอบรหัสผู้ใช้ไม่สำเร็จ" }, { status: existsResponse.status });
  }

  if (existsPayload.data === true) return null;

  const registerResponse = await fetch(`${baseUrl}/register-username`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Accept-Language": request.headers.get("accept-language") ?? "th" },
    body: JSON.stringify({
      username,
      password: randomBytes(12).toString("hex"),
      name: String(payload.name ?? payload.user_profile_name ?? username),
    }),
    cache: "no-store",
  });
  const registerPayload = await readJsonOrText(registerResponse);
  if (!registerResponse.ok || !isRecord(registerPayload) || registerPayload.success === false) {
    return NextResponse.json({ success: false, message: extractMessage(registerPayload) ?? "สร้างผู้ใช้เข้าสู่ระบบไม่สำเร็จ" }, { status: registerResponse.status });
  }

  return null;
}

async function proxyJson(request: Request, baseUrl: string, path: string, init: RequestInit): Promise<NextResponse> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 30000);

  try {
    const response = await fetch(`${baseUrl}${path}`, {
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
    if (isRecord(payload)) return NextResponse.json(payload, { status: response.status });
    return NextResponse.json({ success: response.ok, message: String(payload ?? "") }, { status: response.status });
  } catch (error) {
    const message =
      error instanceof Error && error.name === "AbortError"
        ? "Server ไม่ตอบกลับทันเวลา"
        : "ไม่สามารถเชื่อมต่อ Server ได้";
    return NextResponse.json({ success: false, message }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}

async function readBody(request: Request): Promise<unknown> {
  try {
    return await request.json();
  } catch {
    return undefined;
  }
}

function stripProxyKeys(body: Record<string, unknown>): Record<string, unknown> {
  const {
    action: _action,
    backendUrl: _backendUrl,
    createWithExport: _createWithExport,
    ...payload
  } = body;
  void _action;
  void _backendUrl;
  void _createWithExport;
  return payload;
}

function forwardPagingParams(sourceUrl: URL, target: URLSearchParams) {
  for (const key of ["offset", "limit", "q", "page", "branch_key", "branchcode", "branchguid", "group-number", "company_guid", "sort"]) {
    const value = sourceUrl.searchParams.get(key);
    if (value) target.set(key, value);
  }
  if (!target.has("offset") && !target.has("page")) target.set("offset", "0");
  if (!target.has("limit")) target.set("limit", "1000");
}

function badPayload() {
  return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
}
