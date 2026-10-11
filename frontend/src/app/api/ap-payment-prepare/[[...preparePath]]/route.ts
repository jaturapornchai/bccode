import { NextResponse } from "next/server";
import { getBackendUrlFromRequest, getMainApiUrl, isRecord, proxyMainApiJson, requireBearerToken } from "@/lib/workspace-api";

// BFF ของใบเตรียมจ่ายชำระหนี้ (design P5: route group ของตัวเอง ไม่ใช่ /transaction/:module) — backend
// backend/internal/transaction/httpapi/ap_payment_prepare_http.go registerPrepare
type Context = { params: Promise<{ preparePath?: string[] }> };
const bad = (message: string, status = 400) => NextResponse.json({ success: false, message }, { status });

// คำค้น/ตัวกรอง/หน้า ที่ backend อ่านจริง (listPrepares, prepareOutstanding, listPrepareCreditors)
const ALLOWED_GET_PARAMS = ["q", "page", "limit", "fromdate", "todate", "partner", "status"];
const COLLECTIONS = ["list", "outstanding", "creditors"];
const DOC_VIEWS = ["outstanding", "events"];

function validSegments(segments: string[]): boolean {
  // \p{M} = สระบน/ล่างและวรรณยุกต์ไทย (เลขที่เช่น "ตจ6910-0001") — docs/kms/17-dev-gotchas.md
  return segments.every((seg) => /^[\p{L}\p{M}\p{N}_.-]+$/u.test(seg) && seg !== "." && seg !== "..");
}

// GET: list | outstanding | creditors | <docno> | <docno>/outstanding | <docno>/events
function isReadPath(path: string[]): boolean {
  if (path.length === 1) return true;
  return path.length === 2 && !COLLECTIONS.includes(path[0]) && DOC_VIEWS.includes(path[1]);
}

// ส่งข้อความ JSON เดิมต่อ ไม่ parse แล้ว stringify ใหม่ (ยอดเงินไม่ถูกปัดเป็น double ระหว่างทาง, Rule 20); null = ไม่ใช่ JSON object
async function readJsonObject(request: Request): Promise<string | null> {
  try {
    const text = await request.text();
    const body: unknown = JSON.parse(text);
    return isRecord(body) && !Array.isArray(body) ? text : null;
  } catch {
    return null;
  }
}

function target(path: string[], query = ""): string {
  const tail = path.map(encodeURIComponent).join("/");
  return `/ap-payment-prepare${tail ? `/${tail}` : ""}${query ? `?${query}` : ""}`;
}

async function forward(request: Request, path: string[], method: "POST" | "PUT"): Promise<Response> {
  const body = await readJsonObject(request);
  if (body === null) return bad("ข้อมูลต้องอยู่ในรูปแบบ JSON Object");
  try {
    return await proxyMainApiJson(request, getMainApiUrl(getBackendUrlFromRequest(request)), target(path), { method, body }, { userErrorStatusOk: true });
  } catch {
    return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่");
  }
}

export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;
  const path = (await context.params).preparePath ?? [];
  if (!isReadPath(path)) return bad("ไม่พบรายการที่ต้องการ", 404);
  if (!validSegments(path)) return bad("พารามิเตอร์ไม่ถูกต้อง");

  const query = new URLSearchParams();
  new URL(request.url).searchParams.forEach((value, key) => {
    if (ALLOWED_GET_PARAMS.includes(key)) query.set(key, value);
  });
  try {
    return await proxyMainApiJson(request, getMainApiUrl(getBackendUrlFromRequest(request)), target(path, query.toString()), { method: "GET" }, { userErrorStatusOk: true });
  } catch {
    return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่");
  }
}

// POST: (ว่าง) = บันทึกใบใหม่ | <docno>/cancel = ยกเลิกทั้งใบพร้อมเหตุผล
export async function POST(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;
  const path = (await context.params).preparePath ?? [];
  if (!(path.length === 0 || (path.length === 2 && path[1] === "cancel"))) return bad("ไม่พบรายการที่ต้องการ", 404);
  if (!validSegments(path)) return bad("พารามิเตอร์ไม่ถูกต้อง");
  return forward(request, path, "POST");
}

// PUT: <docno> = แก้ใบเดิม (ส่ง version มาด้วย)
export async function PUT(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;
  const path = (await context.params).preparePath ?? [];
  if (path.length !== 1 || COLLECTIONS.includes(path[0])) return bad("ไม่พบรายการที่ต้องการ", 404);
  if (!validSegments(path)) return bad("พารามิเตอร์ไม่ถูกต้อง");
  return forward(request, path, "PUT");
}
