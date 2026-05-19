import { NextResponse } from "next/server";
import { validateBackendUrl } from "@/lib/backend-url";
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

  const base = resolveBaseUrl(request, resolved.config);
  if (base instanceof NextResponse) return base;

  return proxyJson(request, base, buildGetPath(request, resolved.config, resolved.id), buildGetInit(request, resolved.config));
}

export async function POST(request: Request, context: SystemSettingsProxyContext) {
  const body = await readBody(request);
  if (!isRecord(body)) return badPayload();

  const resolved = await resolveProxy(context);
  if (resolved instanceof NextResponse) return resolved;

  const base = resolveBaseUrl(request, resolved.config, body);
  if (base instanceof NextResponse) return base;

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
  if (!isRecord(body)) return badPayload();

  const resolved = await resolveProxy(context);
  if (resolved instanceof NextResponse) return resolved;

  const base = resolveBaseUrl(request, resolved.config, body);
  if (base instanceof NextResponse) return base;

  const path = buildWritePath(request, resolved.config, resolved.id, "PUT", body);
  const payload = buildWritePayload(request, resolved.config, resolved.id, body);
  return proxyJson(request, base, path, { method: resolved.config.kind === "atlas" || resolved.config.kind === "ai-provider" ? "POST" : "PUT", body: JSON.stringify(payload) });
}

export async function DELETE(request: Request, context: SystemSettingsProxyContext) {
  const body = await readBody(request);
  const resolved = await resolveProxy(context);
  if (resolved instanceof NextResponse) return resolved;

  const base = resolveBaseUrl(request, resolved.config, isRecord(body) ? body : undefined);
  if (base instanceof NextResponse) return base;

  const path = buildDeletePath(resolved.config, resolved.id);
  const payload = buildDeletePayload(request, resolved.config, resolved.id, isRecord(body) ? body : {});
  return proxyJson(request, base, path, {
    method: resolved.config.kind === "atlas" || resolved.config.kind === "ai-provider" ? "POST" : "DELETE",
    body: payload === undefined ? undefined : JSON.stringify(payload),
  });
}

async function resolveProxy(context: SystemSettingsProxyContext): Promise<ResolvedProxy | NextResponse> {
  const { settingPath } = await context.params;
  const [slug, ...rest] = settingPath ?? [];
  if (!slug) return NextResponse.json({ success: false, message: "ไม่พบหน้าจอตั้งค่า" }, { status: 404 });

  const config = getSystemSettingConfig(slug);
  if (!config) return NextResponse.json({ success: false, message: "หน้าจอตั้งค่านี้ยังไม่รองรับ" }, { status: 404 });

  return { config, id: rest.join("/") };
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
    return "/listsourceshops";
  }

  const basePath = id ? `${config.basePath}/${encodeURIComponent(id)}` : (config.listPath ?? config.basePath ?? "");
  forwardPagingParams(url, query);
  const queryText = query.toString();
  return queryText ? `${basePath}?${queryText}` : basePath;
}

function buildGetInit(request: Request, config: SystemSettingConfig): RequestInit {
  const url = new URL(request.url);
  const shopid = url.searchParams.get("shopid") ?? url.searchParams.get("shop_id") ?? "";

  if (config.kind === "atlas") {
    return {
      method: "POST",
      body: JSON.stringify({
        collection: config.collection,
        shopid,
        limit: Number(url.searchParams.get("limit") ?? "1000"),
        skip: Number(url.searchParams.get("offset") ?? "0"),
      }),
    };
  }

  if (config.kind === "ai-provider") {
    return { method: "POST", body: JSON.stringify({ shop_id: shopid }) };
  }

  return { method: "GET" };
}

function buildWritePath(request: Request, config: SystemSettingConfig, id: string, method: "POST" | "PUT", body: Record<string, unknown>): string {
  if (config.kind === "company") {
    const shopid = String(body.shopid ?? new URL(request.url).searchParams.get("shopid") ?? id);
    return `/shop/${encodeURIComponent(shopid)}`;
  }

  if (config.kind === "restaurant-setting") {
    return id ? `/restaurant/settings/${encodeURIComponent(id)}` : "/restaurant/settings";
  }

  if (config.kind === "atlas") {
    return "/atlas/update";
  }

  if (config.kind === "goapi-crud") {
    const createWithExport = Boolean(body.createWithExport);
    if (method === "POST" && createWithExport) return "/api/mcp/keys/create-with-export";
    return id ? `${config.basePath}/${encodeURIComponent(id)}` : (config.basePath ?? "");
  }

  if (config.kind === "ai-provider") {
    return "/api/v1/ai-provider/save";
  }

  if (config.kind === "copy-uat") {
    return body.action === "copy" ? "/copymongouattodev" : "/previewcopymongo";
  }

  if (config.slug === "user") return "/shop/permission";
  return id ? `${config.basePath}/${encodeURIComponent(id)}` : (config.basePath ?? "");
}

function buildDeletePath(config: SystemSettingConfig, id: string): string {
  if (config.kind === "atlas") return "/atlas/delete";
  if (config.kind === "ai-provider") return "/api/v1/ai-provider/delete";
  if (config.kind === "goapi-crud") return `${config.basePath}/${encodeURIComponent(id)}`;
  if (config.slug === "user") return `/shop/permission/${encodeURIComponent(id)}`;
  return `${config.basePath}/${encodeURIComponent(id)}`;
}

function buildWritePayload(request: Request, config: SystemSettingConfig, id: string, body: Record<string, unknown>): Record<string, unknown> {
  const payload = stripProxyKeys(body);
  const url = new URL(request.url);
  const shopid = String(payload.shopid ?? payload.shop_id ?? url.searchParams.get("shopid") ?? "");

  if (config.kind === "restaurant-setting") {
    const { guidfixed: _guidfixed, ...rest } = payload;
    void _guidfixed;
    return {
      code: config.code,
      body: typeof payload.body === "string" ? payload.body : JSON.stringify(rest),
    };
  }

  if (config.kind === "atlas") {
    const key = String(payload[config.idField ?? ""] ?? id);
    const { _id: _mongoId, ...atlasData } = payload;
    void _mongoId;
    return {
      collection: config.collection,
      shopid,
      email: key,
      cartid: key,
      data: { ...atlasData, shopid },
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
  const shopid = String(body.shopid ?? body.shop_id ?? url.searchParams.get("shopid") ?? "");

  if (config.kind === "atlas") {
    return {
      collection: config.collection,
      shopid,
      email: id,
      cartid: id,
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
      password: "12345",
      name: String(payload.name ?? payload.userprofilename ?? username),
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
  for (const key of ["offset", "limit", "q", "page", "branch_key", "branchcode", "branchguid"]) {
    const value = sourceUrl.searchParams.get(key);
    if (value) target.set(key, value);
  }
  if (!target.has("offset") && !target.has("page")) target.set("offset", "0");
  if (!target.has("limit")) target.set("limit", "1000");
}

function badPayload() {
  return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
}
