import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  isRecord,
  proxyMainApiJson,
  type ApiProxyBody,
} from "@/lib/workspace-api";

type CurrencyProxyContext = {
  params: Promise<{ currencyPath?: string[] }>;
};

export async function GET(request: Request, context: CurrencyProxyContext) {
  const mainApiUrl = resolveMainApiUrl(request);
  if (mainApiUrl instanceof NextResponse) return mainApiUrl;

  const { currencyPath } = await context.params;
  return proxyMainApiJson(request, mainApiUrl, buildCurrencyPath(request, currencyPath), { method: "GET" });
}

export async function POST(request: Request, context: CurrencyProxyContext) {
  const body = await readBody(request);
  if (!body || !isRecord(body)) {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const mainApiUrl = resolveMainApiUrl(request, body);
  if (mainApiUrl instanceof NextResponse) return mainApiUrl;

  const { currencyPath } = await context.params;
  const { backendUrl: _backendUrl, ...payload } = body;
  void _backendUrl;

  return proxyMainApiJson(request, mainApiUrl, buildCurrencyPath(request, currencyPath), {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function PUT(request: Request, context: CurrencyProxyContext) {
  const body = await readBody(request);
  if (!body || !isRecord(body)) {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const mainApiUrl = resolveMainApiUrl(request, body);
  if (mainApiUrl instanceof NextResponse) return mainApiUrl;

  const { currencyPath } = await context.params;
  const { backendUrl: _backendUrl, ...payload } = body;
  void _backendUrl;

  return proxyMainApiJson(request, mainApiUrl, buildCurrencyPath(request, currencyPath), {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export async function DELETE(request: Request, context: CurrencyProxyContext) {
  const body = await readBody(request);
  const mainApiUrl = resolveMainApiUrl(request, isRecord(body) ? body : undefined);
  if (mainApiUrl instanceof NextResponse) return mainApiUrl;

  const { currencyPath } = await context.params;
  const payload = Array.isArray(body)
    ? body
    : isRecord(body)
      ? stripBackendUrl(body)
      : undefined;

  return proxyMainApiJson(request, mainApiUrl, buildCurrencyPath(request, currencyPath), {
    method: "DELETE",
    body: payload === undefined ? undefined : JSON.stringify(payload),
  });
}

function resolveMainApiUrl(request: Request, body?: ApiProxyBody): string | NextResponse {
  try {
    return getMainApiUrl(getBackendUrlFromRequest(request, body));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }
}

function buildCurrencyPath(request: Request, segments: string[] = []): string {
  const safeSegments = segments.map((segment) => encodeURIComponent(segment));
  const path = ["/currency", ...safeSegments].join("/");
  const source = new URL(request.url).searchParams;
  const forwarded = new URLSearchParams();

  source.forEach((value, key) => {
    if (key !== "backendUrl") forwarded.append(key, value);
  });

  const query = forwarded.toString();
  return query ? `${path}?${query}` : path;
}

async function readBody(request: Request): Promise<unknown> {
  try {
    return await request.json();
  } catch {
    return undefined;
  }
}

function stripBackendUrl(body: Record<string, unknown>): Record<string, unknown> {
  const { backendUrl: _backendUrl, ...payload } = body;
  void _backendUrl;
  return payload;
}
