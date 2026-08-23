import { NextResponse } from "next/server";
import { serverGoApiBase, validateBackendUrl } from "@/lib/backend-url";
import { verifyHs256Jwt } from "@/lib/server-jwt";
import { getSystemSettingConfig, type SystemSettingConfig } from "@/lib/system-setting-screens";
import { flattenMenuItems } from "@/lib/menu-data";
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
  const unsupportedResponse = rejectUnsupportedProxy(resolved.config, "GET");
  if (unsupportedResponse) return unsupportedResponse;

  const tenantResponse = await validateTenantAccess(request, resolved.config);
  if (tenantResponse) return tenantResponse;

  if (resolved.config.kind === "permission-catalog") {
    const authorization = requireBearerToken(request);
    if (typeof authorization !== "string") return authorization;
    return permissionCatalogResponse(request, resolved.id);
  }

  const base = resolveBaseUrl(request, resolved.config);
  if (base instanceof NextResponse) return base;

  return proxyJson(request, base, buildGetPath(request, resolved.config, resolved.id), buildGetInit(request, resolved.config, resolved.id));
}

export async function POST(request: Request, context: SystemSettingsProxyContext) {
  const body = await readBody(request);
  if (!isRecord(body) && !Array.isArray(body)) return badPayload();

  const resolved = await resolveProxy(context);
  if (resolved instanceof NextResponse) return resolved;
  const unsupportedResponse = rejectUnsupportedProxy(resolved.config, "POST");
  if (unsupportedResponse) return unsupportedResponse;

  const tenantResponse = await validateTenantAccess(request, resolved.config, Array.isArray(body) ? undefined : body);
  if (tenantResponse) return tenantResponse;

  const base = resolveBaseUrl(request, resolved.config, Array.isArray(body) ? undefined : body);
  if (base instanceof NextResponse) return base;

  if (Array.isArray(body)) {
    const path = resolved.config.basePath ?? "";
    return proxyJson(request, base, path, { method: "POST", body: JSON.stringify(body) });
  }

  const path = buildWritePath(request, resolved.config, "", "POST", body);
  const payload = buildWritePayload(request, resolved.config, "", body);
  return proxyJson(request, base, path, { method: resolved.config.slug === "user" ? "PUT" : "POST", body: JSON.stringify(payload) });
}

export async function PUT(request: Request, context: SystemSettingsProxyContext) {
  const body = await readBody(request);
  if (!isRecord(body) && !Array.isArray(body)) return badPayload();

  const resolved = await resolveProxy(context);
  if (resolved instanceof NextResponse) return resolved;
  const unsupportedResponse = rejectUnsupportedProxy(resolved.config, "PUT");
  if (unsupportedResponse) return unsupportedResponse;

  const tenantResponse = await validateTenantAccess(request, resolved.config, Array.isArray(body) ? undefined : body);
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
  const unsupportedResponse = rejectUnsupportedProxy(resolved.config, "DELETE");
  if (unsupportedResponse) return unsupportedResponse;

  const tenantResponse = await validateTenantAccess(request, resolved.config, isRecord(body) ? body : undefined);
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
    if (usesGoApi(config)) {
      validateBackendUrl(backendUrl);
      return serverGoApiBase();
    }
    return getMainApiUrl(backendUrl);
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }
}

function rejectUnsupportedProxy(config: SystemSettingConfig, method: "GET" | "POST" | "PUT" | "DELETE"): NextResponse | null {
  if (config.kind === "report") {
    return NextResponse.json({ success: false, message: "รายงานนี้ไม่มี CRUD API" }, { status: 405 });
  }
  if (config.kind === "permission-catalog" && method !== "GET") {
    return NextResponse.json({ success: false, message: "รายการสิทธิ์หน้าจอแก้ไขไม่ได้" }, { status: 405 });
  }
  return null;
}

function usesGoApi(config: SystemSettingConfig): boolean {
  return config.kind === "ai-provider" || config.kind === "atlas" || config.kind === "copy-uat" || config.kind === "goapi-crud";
}

async function validateTenantAccess(
  request: Request,
  config: SystemSettingConfig,
  body?: Record<string, unknown>,
): Promise<NextResponse | null> {
  const requestedHoldingCode = getRequestedHoldingCode(request, body);
  if (!requestedHoldingCode) return null;

  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  if (process.env.JWT_SECRET_KEY?.trim()) {
    const jwt = verifyHs256Jwt(authorization);
    if (!jwt.ok) return NextResponse.json({ success: false, message: jwt.message }, { status: jwt.status });
  }

  if (config.kind !== "atlas") return null;
  return ensureBackendHoldingAccess(request, body, authorization, requestedHoldingCode);
}

async function ensureBackendHoldingAccess(
  request: Request,
  body: Record<string, unknown> | undefined,
  authorization: string,
  requestedHoldingCode: string,
): Promise<NextResponse | null> {
  let mainApiUrl = "";
  try {
    mainApiUrl = getMainApiUrl(getBackendUrlFromRequest(request, body));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 10000);

  try {
    const response = await fetch(`${mainApiUrl}/select-holding`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
      },
      body: JSON.stringify({ holdingcode: requestedHoldingCode }),
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);
    if (response.ok && (!isRecord(payload) || payload.success !== false)) return null;
    return NextResponse.json(
      {
        success: false,
        message: extractMessage(payload) ?? "ไม่มีสิทธิ์เข้าถึงข้อมูลกลุ่มกิจการนี้ กรุณาเลือกกลุ่มกิจการใหม่หรือเข้าสู่ระบบใหม่",
      },
      { status: response.status === 401 ? 401 : 403 },
    );
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

// Resolve the holdingcode from a record (body/payload) first, then the query string.
function holdingCodeFrom(record: Record<string, unknown> | undefined, url: URL): string {
  return String(record?.holdingcode ?? url.searchParams.get("holdingcode") ?? "");
}

function getRequestedHoldingCode(request: Request, body?: Record<string, unknown>): string {
  return holdingCodeFrom(body, new URL(request.url)).trim();
}

function buildGetPath(request: Request, config: SystemSettingConfig, id: string): string {
  const url = new URL(request.url);
  const query = new URLSearchParams();
  const holdingcode = holdingCodeFrom(undefined, url);

  if (config.kind === "company") {
    const targetShop = id || holdingcode;
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
    query.set("holdingcode", holdingcode);
    const basePath = id
      ? `${config.basePath ?? ""}/${encodeProxyPathId(config, id)}`
      : (config.listPath ?? config.basePath ?? "");
    return `${basePath}?${query.toString()}`;
  }

  if (config.kind === "ai-provider") {
    return "/api/v1/ai-provider/list";
  }

  if (config.kind === "copy-uat") {
    const sourceEnvironment = url.searchParams.get("sourceenvironment") ?? url.searchParams.get("sourceenv") ?? "uat";
    return `/listsourceshops?sourceenvironment=${encodeURIComponent(sourceEnvironment)}`;
  }

  let basePath = id ? `${config.basePath}/${encodeProxyPathId(config, id)}` : (config.listPath ?? config.basePath ?? "");
  const companyGuid = url.searchParams.get("companyguid");
  if (!id && config.slug === "branch" && companyGuid) {
    basePath = config.basePath ?? "";
  }
  forwardPagingParams(url, query);
  const queryText = query.toString();
  return queryText ? `${basePath}?${queryText}` : basePath;
}

function buildGetInit(request: Request, config: SystemSettingConfig, id = ""): RequestInit {
  const url = new URL(request.url);
  const holdingcode = holdingCodeFrom(undefined, url);
  const holdingCode = url.searchParams.get("holdingcode")?.trim() ?? "";

  if (config.kind === "atlas") {
    const body: Record<string, unknown> = {
      collection: config.collection,
      holdingcode,
      limit: Number(url.searchParams.get("limit") ?? "1000"),
      skip: Number(url.searchParams.get("offset") ?? "0"),
    };
    if (holdingCode) body.holdingcode = holdingCode;
    if (id) {
      body.guidfixed = id;
      body.email = id;
      body.cartid = id;
      if (config.collection === "employeepermissions") body.useruid = id;
    }
    return {
      method: "POST",
      body: JSON.stringify(body),
    };
  }

  if (config.kind === "ai-provider") {
    return { method: "POST", body: JSON.stringify({ holdingcode: holdingcode }) };
  }

  return { method: "GET" };
}

function permissionCatalogResponse(request: Request, id: string): NextResponse {
  const url = new URL(request.url);
  const query = (url.searchParams.get("q") ?? "").trim().toLowerCase();
  const catalog = flattenMenuItems().map((item) => ({
    _id: item.id,
    permissioncode: item.id,
    permissionname: item.label.th,
    names: Object.entries(item.label)
      .filter(([code, name]) => code !== "key" && typeof name === "string" && name.trim())
      .map(([code, name]) => ({ code, name })),
    description: item.route,
    category: item.category,
    isactive: true,
  }));
  const matched = catalog.filter((item) => {
    if (id && item.permissioncode !== id) return false;
    if (!query) return true;
    return `${item.permissioncode} ${item.permissionname} ${item.description} ${item.category}`
      .toLowerCase()
      .includes(query);
  });
  const offset = boundedCatalogNumber(url.searchParams.get("offset"), 0, 0, matched.length);
  const limit = boundedCatalogNumber(url.searchParams.get("limit"), 100, 1, 1000);
  const data = id ? matched.slice(0, 1) : matched.slice(offset, offset + limit);
  return NextResponse.json({ success: true, data: id ? (data[0] ?? null) : data, total: matched.length });
}

function boundedCatalogNumber(raw: string | null, fallback: number, minimum: number, maximum = Number.MAX_SAFE_INTEGER): number {
  const value = Number(raw);
  if (!Number.isInteger(value)) return fallback;
  return Math.min(maximum, Math.max(minimum, value));
}

function buildWritePath(request: Request, config: SystemSettingConfig, id: string, method: "POST" | "PUT", body: Record<string, unknown>): string {
  const url = new URL(request.url);
  const search = url.search;
  let path = "";

  if (config.kind === "company") {
    const holdingcode = String(body.holdingcode ?? url.searchParams.get("holdingcode") ?? id);
    path = `/holding/${encodeURIComponent(holdingcode)}`;
  } else if (config.kind === "restaurant-setting") {
    path = id ? `/restaurant/settings/${encodeURIComponent(id)}` : "/restaurant/settings";
  } else if (config.kind === "atlas") {
    path = "/atlas/update";
  } else if (config.kind === "goapi-crud") {
    path = id ? `${config.basePath}/${encodeProxyPathId(config, id)}` : (config.basePath ?? "");
  } else if (config.kind === "ai-provider") {
    path = "/api/v1/ai-provider/save";
  } else if (config.kind === "copy-uat") {
    path = body.action === "copy" ? "/copymongouattodev" : "/previewcopymongo";
  } else if (config.slug === "user") {
    path = "/holding/permission";
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
  else if (config.slug === "user") path = `/holding/permission/${encodeProxyPathId(config, id)}`;
  else path = `${config.basePath}/${encodeProxyPathId(config, id)}`;

  return search ? `${path}${search}` : path;
}

function buildWritePayload(request: Request, config: SystemSettingConfig, id: string, body: Record<string, unknown>): Record<string, unknown> {
  const payload = stripProxyKeys(body);
  normalizeAccessScopePayload(config.slug, payload);
  normalizeBusinessCodeFields(config, payload);
  const url = new URL(request.url);
  const holdingcode = holdingCodeFrom(payload, url);
  const holdingCode = holdingcode.trim();

  if (config.kind === "restaurant-setting") {
    const { guidfixed: _guidfixed, ...rest } = payload;
    void _guidfixed;
    return {
      code: config.code,
      body: typeof payload.body === "string" ? payload.body : JSON.stringify(rest),
    };
  }

  if (config.kind === "atlas") {
    const key = String(payload.guidfixed ?? payload[config.idField ?? ""] ?? id);
    const legacyKey = id && id !== key ? id : "";
    const {
      _id: _mongoId,
      holdingcode: _legacyHoldingCode,
      holdingcode: _legacyHoldingCodeAlt,
      updatedAt: _legacyUpdatedAt,
      updatedBy: _legacyUpdatedBy,
      createdAt: _legacyCreatedAt,
      createdBy: _legacyCreatedBy,
      ...atlasData
    } = payload;
    void _mongoId;
    void _legacyHoldingCode;
    void _legacyHoldingCodeAlt;
    void _legacyUpdatedAt;
    void _legacyUpdatedBy;
    void _legacyCreatedAt;
    void _legacyCreatedBy;
    if (!holdingCode) delete atlasData.holdingcode;
    return {
      collection: config.collection,
      holdingcode,
      ...(holdingCode ? { holdingcode: holdingCode } : {}),
      guidfixed: key,
      email: legacyKey,
      cartid: legacyKey,
      useruid: typeof payload.useruid === "string" ? payload.useruid : undefined,
      data: { ...atlasData, ...(holdingCode ? { holdingcode: holdingCode } : {}), guidfixed: key },
      upsert: true,
    };
  }

  if (config.kind === "goapi-crud") {
    return {
      ...payload,
      holdingcode: holdingcode,
    };
  }

  if (config.kind === "ai-provider") {
    return {
      ...payload,
      holdingcode: holdingcode,
    };
  }

  if (config.kind === "copy-uat") {
    return payload;
  }

  return payload;
}

function normalizeAccessScopePayload(slug: string, payload: Record<string, unknown>) {
  const field =
    slug === "user"
      ? "accessscopes"
      : slug === "approvalsetting"
        ? "approvalrules"
        : slug === "permissiondefinition" || slug === "permissiongroup" || slug === "permissionlink"
          ? "scoperules"
          : "";
  if (!field) return;
  const value = payload[field] ?? payload.accessscopes ?? payload.businesscodes ?? payload.companyguids;
  payload[field] = normalizeScopeRules(value);
}

function normalizeScopeRules(value: unknown): Record<string, unknown>[] {
  const raw = scopeRawArray(value);
  const seen = new Set<string>();
  const rules: Record<string, unknown>[] = [];
  for (const item of raw) {
    const rule = normalizeScopeRule(item);
    const key = `${rule.scopetype}|${rule.businesscode ?? ""}|${rule.branchcode ?? ""}`;
    if (seen.has(key)) continue;
    seen.add(key);
    rules.push(rule);
  }
  return rules;
}

function scopeRawArray(value: unknown): unknown[] {
  if (Array.isArray(value)) return value;
  if (typeof value !== "string") return [];
  const trimmed = value.trim();
  if (!trimmed) return [];
  try {
    const parsed = JSON.parse(trimmed) as unknown;
    if (Array.isArray(parsed)) return parsed;
  } catch {
    // Fall through to comma-separated legacy company list.
  }
  return trimmed.split(",").map((item) => item.trim()).filter(Boolean);
}

function normalizeScopeRule(value: unknown): Record<string, unknown> {
  if (typeof value === "string") {
    const businessCode = normalizeBusinessCode(value);
    return businessCode
      ? { scopetype: "company", businesscode: businessCode, allbranches: true }
      : { scopetype: "holding" };
  }
  const record = isRecord(value) ? value : {};
  const rawScope = String(record.scopetype ?? record.scopeType ?? "holding").trim().toLowerCase();
  const businessCode = normalizeBusinessCode(record.businesscode ?? record.businessCode ?? record.companycode ?? record.companyCode);
  if (rawScope === "holding" || !businessCode) return { scopetype: "holding", allbranches: false };
  const allBranches = booleanValue(record.allbranches ?? record.allBranches ?? record.useallbranches);
  if (rawScope === "company" || allBranches) return { scopetype: "company", businesscode: businessCode, allbranches: true };
  return {
    scopetype: "branch",
    businesscode: businessCode,
    branchcode: normalizeBranchCode(record.branchcode ?? record.branchCode ?? record.code),
    allbranches: false,
  };
}

function normalizeBusinessCode(value: unknown): string {
  // Uppercase + strip ALL whitespace (no spaces inside a business code).
  return String(value ?? "").toUpperCase().replace(/\s+/g, "");
}

// Server-side enforcement: re-normalize every business-code field for this screen
// (uppercase + no whitespace) so a code never persists with spaces even if the client
// is bypassed. Mirrors the frontend buildPayload normalization.
function normalizeBusinessCodeFields(config: SystemSettingConfig, payload: Record<string, unknown>) {
  for (const field of config.fields ?? []) {
    if (field.businessCode && typeof payload[field.key] === "string") {
      payload[field.key] = normalizeBusinessCode(payload[field.key]);
    }
  }
}

function normalizeBranchCode(value: unknown): string {
  const raw = String(value ?? "").trim();
  if (!raw) return "";
  return /^\d{1,5}$/.test(raw) ? raw.padStart(5, "0") : raw;
}

function booleanValue(value: unknown): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "number") return value !== 0;
  if (typeof value !== "string") return false;
  const text = value.trim().toLowerCase();
  return text === "true" || text === "1" || text === "yes" || text === "y";
}

function buildDeletePayload(
  request: Request,
  config: SystemSettingConfig,
  id: string,
  body: Record<string, unknown>,
): Record<string, unknown> | undefined {
  const url = new URL(request.url);
  const holdingcode = holdingCodeFrom(body, url);
  const holdingCode = holdingcode.trim();

  if (config.kind === "atlas") {
    return {
      collection: config.collection,
      holdingcode,
      ...(holdingCode ? { holdingcode: holdingCode } : {}),
      guidfixed: id,
      email: id,
      cartid: id,
      useruid: typeof body.useruid === "string" ? body.useruid : undefined,
      deletemany: false,
    };
  }

  if (config.kind === "ai-provider") {
    return {
      holdingcode: holdingcode,
      providername: id,
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
    ...payload
  } = body;
  void _action;
  void _backendUrl;
  return payload;
}

function forwardPagingParams(sourceUrl: URL, target: URLSearchParams) {
  for (const key of ["offset", "limit", "q", "page", "branchkey", "branchcode", "branchguid", "group-number", "companyguid", "sort"]) {
    const value = sourceUrl.searchParams.get(key);
    if (value) target.set(key, value);
  }
  if (!target.has("offset") && !target.has("page")) target.set("offset", "0");
  if (!target.has("limit")) target.set("limit", "1000");
}

function badPayload() {
  return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
}
