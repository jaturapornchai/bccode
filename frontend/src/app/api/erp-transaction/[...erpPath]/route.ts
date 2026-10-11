import { NextResponse } from "next/server";
import { ERP_TRANSACTION_MODULES } from "@/lib/erp-transaction-modules";
import { getBackendUrlFromRequest, getMainApiUrl, isRecord, proxyMainApiJson, requireBearerToken } from "@/lib/workspace-api";

type Context = { params: Promise<{ erpPath: string[] }> };
const bad = (message: string, status = 400) => NextResponse.json({ success: false, message }, { status });

function sanitizeSegments(segments: string[]): boolean {
  // \p{M} = สระบน/ล่างและวรรณยุกต์ไทย (รหัสเช่น "ที่01") — ขาดแล้วรหัสไทยถูกปฏิเสธ (docs/kms/17-dev-gotchas.md)
  return segments.every((seg) => /^[\p{L}\p{M}\p{N}_.-]+$/u.test(seg) && seg !== "." && seg !== "..");
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

  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return bad("รูปแบบข้อมูลไม่ถูกต้อง");
  }

  if (!isRecord(body) && !Array.isArray(body)) {
    return bad("ข้อมูลต้องอยู่ในรูปแบบ JSON");
  }

  try {
    const targetPath = `/transaction/${erpPath.map(encodeURIComponent).join("/")}`;
    return await proxyMainApiJson(
      request,
      getMainApiUrl(getBackendUrlFromRequest(request)),
      targetPath,
      { method: "POST", body: JSON.stringify(body) },
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

  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return bad("รูปแบบข้อมูลไม่ถูกต้อง");
  }

  if (!isRecord(body)) {
    return bad("ข้อมูลต้องอยู่ในรูปแบบ JSON Object");
  }

  try {
    const targetPath = `/transaction/${erpPath.map(encodeURIComponent).join("/")}`;
    return await proxyMainApiJson(
      request,
      getMainApiUrl(getBackendUrlFromRequest(request)),
      targetPath,
      { method: "PUT", body: JSON.stringify(body) },
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
