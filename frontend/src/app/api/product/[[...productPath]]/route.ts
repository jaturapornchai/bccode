import { NextResponse } from "next/server";
import { validateBackendUrl } from "@/lib/backend-url";
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
    const holdingCode = url.searchParams.get("holdingcode") ?? "";
    const qs = new URLSearchParams({ q, page, limit });
    for (const key of ["itemtype", "materialtype"]) {
      const value = url.searchParams.get(key);
      if (value) qs.set(key, value);
    }
    if (holdingCode) {
      return proxyProductPgListJson(request, holdingCode, q, Number(limit) || 50, pageToOffset(page, limit));
    }
    return proxyProductJson(request, base, `/product?${qs.toString()}`, { method: "GET" });
  }

  return proxyProductJson(request, base, `/product/${encodeURIComponent(id)}`, { method: "GET" });
}

async function proxyProductPgListJson(
  request: Request,
  holdingCode: string,
  search: string,
  limit: number,
  offset: number,
): Promise<NextResponse> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  let goApiUrl: string;
  try {
    goApiUrl = validateBackendUrl(getBackendUrlFromRequest(request)).normalizedGoApiUrl;
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 30000);
  try {
    const response = await fetch(`${goApiUrl}/api/product/search`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
      },
      body: JSON.stringify({
        holdingcode: holdingCode,
        search,
        limit,
        offset,
        use_cache: false,
      }),
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);
    if (!isRecord(payload)) {
      const message = String(payload ?? "");
      return NextResponse.json({ success: response.ok, message, source: "pgsql" }, { status: response.status });
    }
    if (!response.ok || payload.status === "error") {
      return NextResponse.json(productPgErrorPayload(payload), { status: response.status });
    }
    const rawRows = Array.isArray(payload.data) ? payload.data : [];
    const rows = rawRows.map(productPgListRowToProduct);
    return NextResponse.json(
      {
        success: response.ok && payload.status !== "error",
        data: rows,
        total: typeof payload.count === "number" ? payload.count : rows.length,
        source: "pgsql",
        message: typeof payload.message === "string" ? payload.message : typeof payload.error === "string" ? payload.error : undefined,
      },
      { status: response.status },
    );
  } catch (error) {
    const message =
      error instanceof Error && error.name === "AbortError"
        ? "Server ไม่ตอบกลับทันเวลา"
        : "ไม่สามารถเชื่อมต่อ Server ได้";
    return NextResponse.json({ success: false, message, source: "pgsql" }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}

function productPgErrorPayload(payload: Record<string, unknown>): Record<string, unknown> {
  const message =
    stringFromRecord(payload, "message") ||
    stringFromRecord(payload, "error") ||
    "โหลดรายการสินค้าจากฐานข้อมูลไม่สำเร็จ";
  const code = stringFromRecord(payload, "code");
  return {
    success: false,
    message,
    ...(code ? { code } : {}),
    source: "pgsql",
  };
}

function pageToOffset(page: string, limit: string): number {
  const pageNumber = Math.max(1, Number(page) || 1);
  const limitNumber = Math.max(1, Number(limit) || 50);
  return (pageNumber - 1) * limitNumber;
}

function productPgListRowToProduct(row: unknown): Record<string, unknown> {
  const r = isRecord(row) ? row : {};
  const code = stringFromRecord(r, "itemcode");
  const name = stringFromRecord(r, "itemname");
  const unitCode = stringFromRecord(r, "unitcode");
  const unitName = stringFromRecord(r, "unitname");
  const unitCount = numberFromRecord(r, "unit_count");
  const balanceQty = numberFromRecord(r, "balanceqty");
  return {
    guidfixed: code,
    code,
    names: name ? [{ code: "th", name }] : [],
    itemtype: 0,
    materialtype: 0,
    categorycode: stringFromRecord(r, "categorycode"),
    vattype: numberFromRecord(r, "vattype"),
    unitcode: unitCode,
    unitnames: unitName ? [{ code: "th", name: unitName }] : [],
    itemunitcode: unitCode,
    itemunitnames: unitName ? [{ code: "th", name: unitName }] : [],
    qty: balanceQty,
    barcodes: [
      {
        barcode: stringFromRecord(r, "barcode"),
        itemunitcode: unitCode,
        itemunitnames: unitName ? [{ code: "th", name: unitName }] : [],
        qty: Math.max(1, numberFromRecord(r, "unitstand")),
        standvalue: Math.max(1, numberFromRecord(r, "unitstand")),
        dividevalue: Math.max(1, numberFromRecord(r, "unitdivide")),
      },
    ].filter((item) => item.barcode),
    _unit_count: unitCount,
    _source: "pgsql",
  };
}

function stringFromRecord(record: Record<string, unknown>, key: string): string {
  const value = record[key];
  return typeof value === "string" ? value : value == null ? "" : String(value);
}

function numberFromRecord(record: Record<string, unknown>, key: string): number {
  const value = record[key];
  if (typeof value === "number") return Number.isFinite(value) ? value : 0;
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

export async function POST(request: Request) {
  const body = await readBody(request);
  const bodyRecord = isRecord(body) ? body : undefined;
  const base = resolveMainApiUrl(request, bodyRecord);
  if (base instanceof NextResponse) return base;

  const payload = getPayload(body);
  if (!isRecord(payload)) {
    return NextResponse.json({ success: false, message: "ไม่พบข้อมูลสินค้า" }, { status: 400 });
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
  if (!id) return NextResponse.json({ success: false, message: "ไม่พบรหัสสินค้า" }, { status: 400 });

  const base = resolveMainApiUrl(request, bodyRecord);
  if (base instanceof NextResponse) return base;

  const payload = getPayload(body);
  if (!isRecord(payload)) {
    return NextResponse.json({ success: false, message: "ไม่พบข้อมูลสินค้า" }, { status: 400 });
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
    return proxyProductJson(request, base, `/product/${encodeURIComponent(id)}`, { method: "DELETE" });
  }

  return NextResponse.json({ success: false, message: "ไม่พบรายการที่ต้องการลบ" }, { status: 400 });
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

async function proxyProductJson(request: Request, baseUrl: string, path: string, init: RequestInit): Promise<NextResponse> {
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

function getPayload(value: unknown): unknown {
  if (!isRecord(value)) return value;
  return value.data ?? value.payload ?? value.product ?? value;
}
