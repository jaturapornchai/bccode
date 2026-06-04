import { NextResponse } from "next/server";
import { validateBackendUrl } from "@/lib/backend-url";
import { getBackendUrlFromRequest, isRecord, readJsonOrText, requireBearerToken } from "@/lib/workspace-api";

type ProductBarcodeListBody = Record<string, unknown> & {
  backendUrl?: string;
  holding_code?: string;
};

export async function POST(request: Request) {
  const body = await readBody(request);
  if (!isRecord(body)) {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const holdingCode = String(body.holding_code ?? "").trim();
  if (!holdingCode) return NextResponse.json({ success: false, message: "กรุณาเลือก holding ก่อน" }, { status: 400 });

  let goApiUrl: string;
  try {
    goApiUrl = validateBackendUrl(getBackendUrlFromRequest(request, body as ProductBarcodeListBody)).normalizedGoApiUrl;
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const { backendUrl: _backendUrl, holding_code: _holdingCode, ...payload } = body;
  void _backendUrl;
  void _holdingCode;

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 20000);

  try {
    const response = await fetch(`${goApiUrl}/api/product/barcode/list`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
      },
      body: JSON.stringify({ ...payload, holding_code: holdingCode }),
      signal: controller.signal,
      cache: "no-store",
    });
    const responsePayload = await readJsonOrText(response);
    if (isRecord(responsePayload)) return NextResponse.json(responsePayload, { status: response.status });
    return NextResponse.json({ success: response.ok, message: String(responsePayload ?? "") }, { status: response.status });
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
