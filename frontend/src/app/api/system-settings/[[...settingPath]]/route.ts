import { NextResponse } from "next/server";
import { serverGoApiBase, validateBackendUrl } from "@/lib/backend-url";
import { verifyHs256Jwt } from "@/lib/server-jwt";
import { getSystemSettingConfig, type SystemSettingConfig } from "@/lib/system-setting-screens";
import { flattenMenuItems } from "@/lib/menu-data";
import {
  getBackendUrlFromRequest,
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

  return proxyJson(request, base, buildGetPath(request, resolved.config, resolved.id), { method: "GET" });
}

export async function POST(request: Request, context: SystemSettingsProxyContext) {
  const body = await readBody(request);
  if (!isRecord(body) && !Array.isArray(body)) return badPayload();

  const resolved = await resolveProxy(context);
  if (resolved instanceof NextResponse) return resolved;
  const unsupportedResponse = rejectUnsupportedProxy(resolved.config, "POST");
  if (unsupportedResponse) return unsupportedResponse;

  const validationError = validateSystemSettingWrite(resolved.config, body);
  if (validationError) return validationError;

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

  const validationError = validateSystemSettingWrite(resolved.config, body);
  if (validationError) return validationError;

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
  return proxyJson(request, base, path, { method: "PUT", body: JSON.stringify(payload) });
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
  return proxyJson(request, base, path, { method: "DELETE" });
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
  return config.kind === "goapi-crud";
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

  return null;
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

  if (config.kind === "goapi-crud") {
    query.set("holdingcode", holdingcode);
    const basePath = id
      ? `${config.basePath ?? ""}/${encodeProxyPathId(config, id)}`
      : (config.listPath ?? config.basePath ?? "");
    return `${basePath}?${query.toString()}`;
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
  } else if (config.kind === "goapi-crud") {
    path = id ? `${config.basePath}/${encodeProxyPathId(config, id)}` : (config.basePath ?? "");
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

  if (config.kind === "goapi-crud") path = `${config.basePath}/${encodeProxyPathId(config, id)}`;
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

  if (config.kind === "restaurant-setting") {
    const { guidfixed: _guidfixed, ...rest } = payload;
    void _guidfixed;
    return {
      code: config.code,
      body: typeof payload.body === "string" ? payload.body : JSON.stringify(rest),
    };
  }

  if (config.kind === "goapi-crud") {
    return {
      ...payload,
      holdingcode: holdingcode,
    };
  }

  if (config.slug === "bookbankscreen" || config.slug === "bookbank") {
    if (payload.logo && (!payload.images || !Array.isArray(payload.images) || payload.images.length === 0)) {
      payload.images = [{ xorder: 0, uri: String(payload.logo) }];
    }
    if (!payload.bookcode && payload.code) {
      payload.bookcode = payload.code;
    }
    if (!payload.banknames || (Array.isArray(payload.banknames) && payload.banknames.length === 0)) {
      if (payload.names && Array.isArray(payload.names) && payload.names.length > 0) {
        payload.banknames = payload.names;
      }
    }
  }

  return payload;
}

function validateSystemSettingWrite(config: SystemSettingConfig, body: unknown): NextResponse | null {
  if (!isRecord(body)) return null;
  if (hasInvalidAccessScopeRules(config.slug, body)) {
    // A languages.tsv key: the screen shows it in the user's language.
    return NextResponse.json(
      { success: false, errorcode: "VALIDATION_FAILED", message: "ss_err_access_scope_invalid" },
      { status: 400 },
    );
  }
  if (config.slug === "bookbankscreen" || config.slug === "bookbank") {
    const hasBanknames = Array.isArray(body.banknames) && body.banknames.some((item) => isRecord(item) && typeof item.name === "string" && item.name.trim() !== "");
    const hasNames = Array.isArray(body.names) && body.names.some((item) => isRecord(item) && typeof item.name === "string" && item.name.trim() !== "");
    if (!hasBanknames && !hasNames) {
      return NextResponse.json({ success: false, message: "กรุณาระบุชื่อธนาคาร" }, { status: 400 });
    }
  }
  return null;
}

function accessScopeField(slug: string): string {
  if (slug === "user") return "accessscopes";
  if (slug === "permissiondefinition" || slug === "permissiongroup") return "scoperules";
  return "";
}

function submittedScopeRules(field: string, payload: Record<string, unknown>): unknown {
  return payload[field] ?? payload.accessscopes ?? payload.businesscodes ?? payload.companyguids;
}

// Every submitted rule must be valid: dropping a bad one silently is unsafe, because for an
// OWNER/ADMIN an all-invalid list would be saved as [] = the whole Holding.
function hasInvalidAccessScopeRules(slug: string, body: Record<string, unknown>): boolean {
  const field = accessScopeField(slug);
  if (!field) return false;
  const value = submittedScopeRules(field, body);
  if (value !== undefined && value !== null && typeof value !== "string" && !Array.isArray(value)) return true;
  return scopeRawArray(value).some((item) => !normalizeScopeRule(item));
}

function normalizeAccessScopePayload(slug: string, payload: Record<string, unknown>) {
  const field = accessScopeField(slug);
  if (!field) return;
  payload[field] = normalizeScopeRules(submittedScopeRules(field, payload));
}

function normalizeScopeRules(value: unknown): Record<string, unknown>[] {
  const raw = scopeRawArray(value);
  const seen = new Set<string>();
  const rules: Record<string, unknown>[] = [];
  for (const item of raw) {
    const rule = normalizeScopeRule(item);
    if (!rule) continue; // unreachable for writes: hasInvalidAccessScopeRules rejected the request
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

// Only an explicit scopetype "holding" grants the whole Holding (companies added later
// included); a rule with an unknown type, without a company, or a branch rule without a
// branch is invalid (null), never widened.
function normalizeScopeRule(value: unknown): Record<string, unknown> | null {
  if (typeof value === "string") {
    const businessCode = normalizeBusinessCode(value);
    return businessCode ? { scopetype: "company", businesscode: businessCode, allbranches: true } : null;
  }
  if (!isRecord(value)) return null;
  const rawScope = String(value.scopetype ?? value.scopeType ?? "").trim().toLowerCase();
  if (rawScope === "holding") return { scopetype: "holding" };
  if (rawScope !== "company" && rawScope !== "branch") return null;
  // The company uid is the company code (PostgreSQL central schema).
  const businessCode = normalizeBusinessCode(
    value.businesscode ?? value.businessCode ?? value.companycode ?? value.companyCode ?? value.companyuid ?? value.companyUID,
  );
  if (!businessCode) return null;
  const allBranches = booleanValue(value.allbranches ?? value.allBranches ?? value.useallbranches);
  if (rawScope === "company" || allBranches) return { scopetype: "company", businesscode: businessCode, allbranches: true };
  const branchCode = normalizeBranchCode(value.branchcode ?? value.branchCode ?? value.branchuid ?? value.code);
  if (!branchCode) return null;
  return { scopetype: "branch", businesscode: businessCode, branchcode: branchCode, allbranches: false };
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
