import { NextResponse } from "next/server";
import { getBackendUrlFromRequest, getMainApiUrl, isRecord, proxyMainApiJson, requireBearerToken } from "@/lib/workspace-api";

type Context = { params: Promise<{ erpPath: string[] }> };
const bad = (message: string, status = 400) => NextResponse.json({ success: false, message }, { status });

const ALLOWED_MODULES = [
  "sale-invoice",
  "sale-order",
  "quotation",
  "sale-invoice-return",
  "purchase",
  "purchase-order",
  "purchase-requisition",
  "purchase-return",
  "paid",
  "pay",
  "stock-balance",
  "stock-receive-product",
  "stock-prickup-product",
  "stock-return-product",
  "stock-transfer",
  "stock-adjustment",
  "receivableother",
  "billingnote",
  "deposit",
  "depositrefund",
  "paidadvance",
  "paidadvancerefund",
  "receivedeposit",
  "receivedepositrefund",
  "bank",
  "chequereceive",
  "chequepayment",
];

function sanitizeSegments(segments: string[]): boolean {
  return segments.every((seg) => /^[\p{L}\p{N}_.-]+$/u.test(seg) && seg !== "." && seg !== "..");
}

export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { erpPath } = await context.params;
  if (!erpPath || erpPath.length === 0) return bad("ไม่พบรายการที่ต้องการ", 404);

  if (!ALLOWED_MODULES.includes(erpPath[0])) {
    return bad("โมดูลไม่ถูกต้อง", 404);
  }

  if (!sanitizeSegments(erpPath)) {
    return bad("พารามิเตอร์ไม่ถูกต้อง", 400);
  }

  try {
    const query = new URLSearchParams();
    const allowedParams = ["q", "page", "limit", "offset", "fromdate", "todate", "custcode", "branchcode", "status", "lang"];
    new URL(request.url).searchParams.forEach((val, key) => {
      if (allowedParams.includes(key)) query.set(key, val);
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

  if (!ALLOWED_MODULES.includes(erpPath[0])) {
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

  if (!ALLOWED_MODULES.includes(erpPath[0])) {
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

  if (!ALLOWED_MODULES.includes(erpPath[0])) {
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
