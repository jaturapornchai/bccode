import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  proxyMainApiJson,
} from "@/lib/workspace-api";

type BomContext = { params: Promise<{ barcode?: string }> };

export async function GET(request: Request, context: BomContext) {
  const { barcode } = await context.params;
  const code = (barcode ?? "").trim();
  if (!code) {
    return NextResponse.json(
      { success: false, message: "ไม่พบบาร์โค้ด" },
      { status: 400 },
    );
  }

  let base: string;
  try {
    base = getMainApiUrl(getBackendUrlFromRequest(request));
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

  const itemCode =
    new URL(request.url).searchParams.get("itemcode")?.trim() ?? "";
  const query = itemCode ? `?itemcode=${encodeURIComponent(itemCode)}` : "";
  return proxyMainApiJson(
    request,
    base,
    `/product/barcode/bom/${encodeURIComponent(code)}${query}`,
    { method: "GET" },
  );
}
