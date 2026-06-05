import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  isRecord,
  proxyMainApiJson,
  type ApiProxyBody,
} from "@/lib/workspace-api";

type ProductPriceHistoryContext = {
  params: Promise<{ historyPath?: string[] }>;
};

export async function GET(request: Request, context: ProductPriceHistoryContext) {
  const mainApiUrl = resolveMainApiUrl(request);
  if (mainApiUrl instanceof NextResponse) return mainApiUrl;

  const { historyPath } = await context.params;
  return proxyMainApiJson(request, mainApiUrl, buildPriceHistoryPath(request, historyPath), { method: "GET" });
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

function buildPriceHistoryPath(request: Request, segments: string[] = []): string {
  const safeSegments = segments.map((segment) => encodeURIComponent(segment));
  const path = ["/product/barcode/price-history", ...safeSegments].join("/");
  const source = new URL(request.url).searchParams;
  const forwarded = new URLSearchParams();

  source.forEach((value, key) => {
    if (key !== "backendUrl" && key !== "holdingcode" && key !== "holdingcode") forwarded.append(key, value);
  });

  const query = forwarded.toString();
  return query ? `${path}?${query}` : path;
}

export async function POST(request: Request) {
  const body = await readBody(request);
  if (!isRecord(body)) {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }
  return NextResponse.json({ success: false, message: "ประวัติราคาเป็นข้อมูลอ่านอย่างเดียว" }, { status: 405 });
}

async function readBody(request: Request): Promise<unknown> {
  try {
    return await request.json();
  } catch {
    return undefined;
  }
}
