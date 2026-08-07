import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  isRecord,
  readJsonOrText,
  requireBearerToken,
  type ApiProxyBody,
} from "@/lib/workspace-api";

type ProductProxyContext = {
  params: Promise<{ productPath?: string[] }>;
};

export async function GET(request: Request, context: ProductProxyContext) {
  const { productPath } = await context.params;
  const id = (productPath ?? []).join("/").trim();

  const base = resolveMainApiUrl(request);
  if (base instanceof NextResponse) return base;

  // If there is no ID, search products
  if (!id) {
    const url = new URL(request.url);
    const q = url.searchParams.get("q") ?? "";
    const page = url.searchParams.get("page") ?? "1";
    const limit = url.searchParams.get("limit") ?? "50";
    const qs = new URLSearchParams({ q, page, limit });
    for (const key of ["itemtype", "materialtype"]) {
      const value = url.searchParams.get(key);
      if (value) qs.set(key, value);
    }
    return proxyProductJson(request, base, `/product?${qs.toString()}`, {
      method: "GET",
    });
  }

  return proxyProductJson(request, base, `/product/${encodeURIComponent(id)}`, {
    method: "GET",
  });
}

export async function POST(request: Request, context: ProductProxyContext) {
  const body = await readBody(request);
  const bodyRecord = isRecord(body) ? body : undefined;
  const base = resolveMainApiUrl(request, bodyRecord);
  if (base instanceof NextResponse) return base;
  const { productPath } = await context.params;
  const action = (productPath ?? []).join("/").trim();
  if (action === "resync") {
    return proxyProductJson(request, base, "/product/resync", {
      method: "POST",
    });
  }
  if (action) {
    return NextResponse.json(
      { success: false, message: "ไม่พบคำสั่งสินค้า" },
      { status: 404 },
    );
  }

  const payload = getPayload(body);
  if (!isRecord(payload)) {
    return NextResponse.json(
      { success: false, message: "ไม่พบข้อมูลสินค้า" },
      { status: 400 },
    );
  }

  return proxyProductJson(request, base, "/product", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function PUT(request: Request, context: ProductProxyContext) {
  const body = await readBody(request);
  const bodyRecord = isRecord(body) ? body : undefined;
  const { productPath } = await context.params;
  const id = (productPath ?? []).join("/").trim();
  if (!id)
    return NextResponse.json(
      { success: false, message: "ไม่พบรหัสสินค้า" },
      { status: 400 },
    );

  const base = resolveMainApiUrl(request, bodyRecord);
  if (base instanceof NextResponse) return base;

  const payload = getPayload(body);
  if (!isRecord(payload)) {
    return NextResponse.json(
      { success: false, message: "ไม่พบข้อมูลสินค้า" },
      { status: 400 },
    );
  }

  return proxyProductJson(request, base, `/product/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export async function DELETE(request: Request, context: ProductProxyContext) {
  const body = await readBody(request);
  const bodyRecord = isRecord(body) ? body : undefined;
  const { productPath } = await context.params;
  const id = (productPath ?? []).join("/").trim();

  const base = resolveMainApiUrl(request, bodyRecord);
  if (base instanceof NextResponse) return base;

  if (id) {
    return proxyProductJson(
      request,
      base,
      `/product/${encodeURIComponent(id)}`,
      { method: "DELETE" },
    );
  }

  return NextResponse.json(
    { success: false, message: "ไม่พบรายการที่ต้องการลบ" },
    { status: 400 },
  );
}

function resolveMainApiUrl(
  request: Request,
  body?: ApiProxyBody,
): string | NextResponse {
  try {
    return getMainApiUrl(getBackendUrlFromRequest(request, body));
  } catch (error) {
    return NextResponse.json(
      {
        success: false,
        message:
          error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง",
      },
      { status: 400 },
    );
  }
}

async function proxyProductJson(
  request: Request,
  baseUrl: string,
  path: string,
  init: RequestInit,
): Promise<NextResponse> {
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
    if (isRecord(payload))
      return NextResponse.json(payload, { status: response.status });
    return NextResponse.json(
      { success: response.ok, message: String(payload ?? "") },
      { status: response.status },
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

async function readBody(request: Request): Promise<unknown> {
  try {
    return await request.json();
  } catch {
    return undefined;
  }
}

function getPayload(value: unknown): unknown {
  if (!isRecord(value)) return value;
  return value.data ?? value.payload ?? value.product ?? value;
}
