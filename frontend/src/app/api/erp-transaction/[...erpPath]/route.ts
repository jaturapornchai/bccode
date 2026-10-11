import { NextResponse } from "next/server";
import { ERP_TRANSACTION_MODULES } from "@/lib/erp-transaction-modules";
import { getBackendUrlFromRequest, getMainApiUrl, isRecord, proxyMainApiJson, requireBearerToken } from "@/lib/workspace-api";

type Context = { params: Promise<{ erpPath: string[] }> };
const bad = (message: string, status = 400) => NextResponse.json({ success: false, message }, { status });

function sanitizeSegments(segments: string[]): boolean {
  // \p{M} = สระบน/ล่างและวรรณยุกต์ไทย (รหัสเช่น "ที่01") — ขาดแล้วรหัสไทยถูกปฏิเสธ (docs/kms/17-dev-gotchas.md)
  return segments.every((seg) => /^[\p{L}\p{M}\p{N}_.-]+$/u.test(seg) && seg !== "." && seg !== "..");
}

// ส่งข้อความ JSON เดิมต่อ ไม่ parse แล้ว stringify ใหม่: ยอดเงินที่ส่งเป็นตัวเลข JSON ความละเอียดสูง (เช่น 9999999999.99999999)
// จะถูกปัดเป็น double ระหว่างทาง (Rule 20) — parse เพื่อตรวจรูปแบบอย่างเดียว; null = ไม่ใช่ JSON
async function readJsonText(request: Request): Promise<{ text: string; body: unknown } | null> {
  try {
    const text = await request.text();
    return { text, body: JSON.parse(text) };
  } catch {
    return null;
  }
}

// partner = จอวันที่ส่งของ; from_/to_*, sale_code = จอปรับปรุงสถานะใบรับสินค้าและพิมพ์ใบวางบิลอัตโนมัติ (ชื่อตรงกับที่ backend อ่านจริง)
const ALLOWED_GET_PARAMS = [
  "q",
  "page",
  "limit",
  "offset",
  "fromdate",
  "todate",
  "custcode",
  "branchcode",
  "status",
  "lang",
  "partner",
  "from_date",
  "to_date",
  "from_docno",
  "to_docno",
  "from_vendor",
  "to_vendor",
  "from_dept",
  "to_dept",
  "from_cust",
  "to_cust",
  "sale_code",
];

export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { erpPath } = await context.params;
  if (!erpPath || erpPath.length === 0) return bad("ไม่พบรายการที่ต้องการ", 404);

  if (!ERP_TRANSACTION_MODULES.includes(erpPath[0])) {
    return bad("โมดูลไม่ถูกต้อง", 404);
  }

  if (!sanitizeSegments(erpPath)) {
    return bad("พารามิเตอร์ไม่ถูกต้อง", 400);
  }

  try {
    const query = new URLSearchParams();
    new URL(request.url).searchParams.forEach((val, key) => {
      if (ALLOWED_GET_PARAMS.includes(key)) query.set(key, val);
    });

    const targetPath = `/transaction/${erpPath.map(encodeURIComponent).join("/")}?${query.toString()}`;
    return await proxyMainApiJson(request, getMainApiUrl(getBackendUrlFromRequest(request)), targetPath, { method: "GET" });
  } catch {
    return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่");
  }
}

export async function POST(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { erpPath } = await context.params;
  if (!erpPath || erpPath.length === 0) return bad("ไม่พบรายการที่ต้องการ", 404);

  if (!ERP_TRANSACTION_MODULES.includes(erpPath[0])) {
    return bad("โมดูลไม่ถูกต้อง", 404);
  }

  if (!sanitizeSegments(erpPath)) {
    return bad("พารามิเตอร์ไม่ถูกต้อง", 400);
  }

  const payload = await readJsonText(request);
  if (!payload) return bad("รูปแบบข้อมูลไม่ถูกต้อง");

  if (!isRecord(payload.body) && !Array.isArray(payload.body)) {
    return bad("ข้อมูลต้องอยู่ในรูปแบบ JSON");
  }

  try {
    const targetPath = `/transaction/${erpPath.map(encodeURIComponent).join("/")}`;
    return await proxyMainApiJson(
      request,
      getMainApiUrl(getBackendUrlFromRequest(request)),
      targetPath,
      { method: "POST", body: payload.text },
      { userErrorStatusOk: true }
    );
  } catch {
    return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่");
  }
}

export async function PUT(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { erpPath } = await context.params;
  if (!erpPath || erpPath.length === 0) return bad("ไม่พบรายการที่ต้องการ", 404);

  if (!ERP_TRANSACTION_MODULES.includes(erpPath[0])) {
    return bad("โมดูลไม่ถูกต้อง", 404);
  }

  if (!sanitizeSegments(erpPath)) {
    return bad("พารามิเตอร์ไม่ถูกต้อง", 400);
  }

  const payload = await readJsonText(request);
  if (!payload) return bad("รูปแบบข้อมูลไม่ถูกต้อง");

  if (!isRecord(payload.body)) {
    return bad("ข้อมูลต้องอยู่ในรูปแบบ JSON Object");
  }

  try {
    const targetPath = `/transaction/${erpPath.map(encodeURIComponent).join("/")}`;
    return await proxyMainApiJson(
      request,
      getMainApiUrl(getBackendUrlFromRequest(request)),
      targetPath,
      { method: "PUT", body: payload.text },
      { userErrorStatusOk: true }
    );
  } catch {
    return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่");
  }
}

export async function DELETE(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { erpPath } = await context.params;
  if (!erpPath || erpPath.length === 0) return bad("ไม่พบรายการที่ต้องการ", 404);

  if (!ERP_TRANSACTION_MODULES.includes(erpPath[0])) {
    return bad("โมดูลไม่ถูกต้อง", 404);
  }

  if (!sanitizeSegments(erpPath)) {
    return bad("พารามิเตอร์ไม่ถูกต้อง", 400);
  }

  try {
    const targetPath = `/transaction/${erpPath.map(encodeURIComponent).join("/")}`;
    return await proxyMainApiJson(
      request,
      getMainApiUrl(getBackendUrlFromRequest(request)),
      targetPath,
      { method: "DELETE" },
      { userErrorStatusOk: true }
    );
  } catch {
    return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่");
  }
}
