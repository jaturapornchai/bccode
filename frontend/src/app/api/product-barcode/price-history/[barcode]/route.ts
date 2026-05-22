import { NextResponse } from "next/server";
import { getBackendUrlFromRequest, getMainApiUrl, proxyMainApiJson } from "@/lib/workspace-api";

type PriceHistoryContext = { params: Promise<{ barcode?: string }> };

export async function GET(request: Request, context: PriceHistoryContext) {
  const { barcode } = await context.params;
  const code = (barcode ?? "").trim();
  if (!code) {
    return NextResponse.json({ success: false, message: "ไม่พบบาร์โค้ด" }, { status: 400 });
  }

  let base: string;
  try {
    base = getMainApiUrl(getBackendUrlFromRequest(request));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const url = new URL(request.url);
  const qs = url.searchParams.toString();
  return proxyMainApiJson(
    request,
    base,
    `/product/barcode/price-history/${encodeURIComponent(code)}${qs ? `?${qs}` : ""}`,
    { method: "GET" },
  );
}
