import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  isRecord,
  readJsonOrText,
  requireBearerToken,
  type ApiProxyBody,
} from "@/lib/workspace-api";

type ProductBarcodeProxyContext = {
  params: Promise<{ barcodePath?: string[] }>;
};

export async function GET(request: Request, context: ProductBarcodeProxyContext) {
  const id = await resolveBarcodeId(context);
  if (!id) return NextResponse.json({ success: false, message: "ไม่พบรหัสบาร์โค้ดสินค้า" }, { status: 400 });

  const base = resolveMainApiUrl(request);
  if (base instanceof NextResponse) return base;

  return proxyProductBarcodeJson(request, base, `/product/barcode/${encodeURIComponent(id)}`, { method: "GET" });
}

export async function POST(request: Request) {
  const body = await readBody(request);
  const bodyRecord = isRecord(body) ? body : undefined;
  const base = resolveMainApiUrl(request, bodyRecord);
  if (base instanceof NextResponse) return base;

  const payload = getBarcodePayload(body);
  if (!isRecord(payload)) {
    return NextResponse.json({ success: false, message: "ไม่พบข้อมูลบาร์โค้ดสินค้า" }, { status: 400 });
  }

  return proxyProductBarcodeJson(request, base, "/product/barcode", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function PUT(request: Request, context: ProductBarcodeProxyContext) {
  const body = await readBody(request);
  const bodyRecord = isRecord(body) ? body : undefined;
  const id = await resolveBarcodeId(context);
  if (!id) return NextResponse.json({ success: false, message: "ไม่พบรหัสบาร์โค้ดสินค้า" }, { status: 400 });

  const base = resolveMainApiUrl(request, bodyRecord);
  if (base instanceof NextResponse) return base;

  const payload = getBarcodePayload(body);
  if (!isRecord(payload)) {
    return NextResponse.json({ success: false, message: "ไม่พบข้อมูลบาร์โค้ดสินค้า" }, { status: 400 });
  }

  return proxyProductBarcodeJson(request, base, `/product/barcode/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export async function DELETE(request: Request, context: ProductBarcodeProxyContext) {
  const body = await readBody(request);
  const bodyRecord = isRecord(body) ? body : undefined;
  const id = await resolveBarcodeId(context);

  const base = resolveMainApiUrl(request, bodyRecord);
  if (base instanceof NextResponse) return base;

  if (id) {
    return proxyProductBarcodeJson(request, base, `/product/barcode/${encodeURIComponent(id)}`, { method: "DELETE" });
  }

  const guids = getGuidList(body);
  if (guids.length === 0) {
    return NextResponse.json({ success: false, message: "ไม่พบรายการที่ต้องการลบ" }, { status: 400 });
  }

  return proxyProductBarcodeJson(request, base, "/product/barcode", {
    method: "DELETE",
    body: JSON.stringify(guids),
  });
}

async function resolveBarcodeId(context: ProductBarcodeProxyContext): Promise<string> {
  const { barcodePath } = await context.params;
  return (barcodePath ?? []).join("/").trim();
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

async function proxyProductBarcodeJson(request: Request, baseUrl: string, path: string, init: RequestInit): Promise<NextResponse> {
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

function getGuidList(value: unknown): string[] {
  if (Array.isArray(value)) return cleanGuidList(value);
  if (!isRecord(value)) return [];
  return cleanGuidList(value.guids ?? value.guidfixeds ?? value.ids ?? value.data);
}

function getBarcodePayload(value: unknown): unknown {
  if (!isRecord(value)) return value;
  return value.data ?? value.payload ?? value.productBarcode ?? value;
}

function cleanGuidList(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.map((item) => String(item ?? "").trim()).filter(Boolean);
}
