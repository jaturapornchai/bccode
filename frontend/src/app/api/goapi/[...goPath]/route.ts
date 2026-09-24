import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  isRecord,
  proxyMainApiJson,
  readJsonOrText,
  requireBearerToken,
} from "@/lib/workspace-api";

type Context = { params: Promise<{ goPath: string[] }> };

// Allowlist matches full paths because goapi exposes many destructive endpoints that must never be reachable.
const GET_ALLOWED_EXACT = ["api/reports/inventory-valuation"];

const POST_ALLOWED_EXACT = [
  "api/report/tax/vat-register",
  "api/report/tax/wht",
  "api/report/tax/wht/certificate",
  "api/report/tax/form/catalog",
  "api/report/tax/form/schema",
  "api/report/tax/form/prefill",
  "api/report/tax/form/compute",
  "api/report/tax/form/pdf",
  "api/report/tax/form/save",
  "api/report/tax/form/list",
  "api/report/tax/form/load",
  "api/report/tax/form/delete",
  "api/report/tax/form/rdfile",
  "api/report/debt/query",
  "api/report/sales/by-document",
];

// ปลายทางที่ backend ตอบเป็นไฟล์ (PDF ใบ 50 ทวิ/แบบยื่นภาษี, .txt ไฟล์ยื่นด้วยสื่อ Format กลาง) — ส่งต่อ byte ตรง ๆ
// ห้ามแปลงเป็น text/JSON: ไฟล์ .txt ต้องคง UTF-8 BOM และ CRLF ไว้ทุกไบต์ ไม่งั้นโปรแกรม SWC-UI ของกรมสรรพากรไม่รับ
const FILE_PATHS = ["api/report/tax/wht/certificate", "api/report/tax/form/pdf", "api/report/tax/form/rdfile"];
const FILE_CONTENT_TYPES = ["application/pdf", "text/plain"];

const ALLOWED_QUERY_PARAMS = [
  "holdingcode",
  "businesscode",
  "warehousecode",
  "fromdate",
  "todate",
  "limit",
  "offset",
  "lang",
];

const bad = (message: string, status = 400) => NextResponse.json({ success: false, message }, { status });

function sanitizeSegments(segments: string[]): boolean {
  // \p{M} = สระบน/ล่างและวรรณยุกต์ไทย — docs/kms/17-dev-gotchas.md
  return segments.every((seg) => /^[\p{L}\p{M}\p{N}_.-]+$/u.test(seg) && seg !== "." && seg !== "..");
}

function toGoApiPath(segments: string[]): string {
  return `/goapi/${segments.map(encodeURIComponent).join("/")}`;
}

function isAllowedGet(segments: string[]): boolean {
  return GET_ALLOWED_EXACT.includes(segments.join("/"));
}

export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { goPath } = await context.params;
  if (!goPath || goPath.length === 0) return bad("not_found", 404);
  if (!sanitizeSegments(goPath)) return bad("error_occurred", 400);
  if (!isAllowedGet(goPath)) return bad("not_found", 404);

  try {
    const query = new URLSearchParams();
    new URL(request.url).searchParams.forEach((value, key) => {
      if (ALLOWED_QUERY_PARAMS.includes(key)) query.set(key, value);
    });
    const queryString = query.toString();
    const targetPath = `${toGoApiPath(goPath)}${queryString ? `?${queryString}` : ""}`;

    return await proxyMainApiJson(request, getMainApiUrl(getBackendUrlFromRequest(request)), targetPath, {
      method: "GET",
    });
  } catch {
    return bad("connection_error", 500);
  }
}

export async function POST(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { goPath } = await context.params;
  if (!goPath || goPath.length === 0) return bad("not_found", 404);
  if (!sanitizeSegments(goPath)) return bad("error_occurred", 400);
  if (!POST_ALLOWED_EXACT.includes(goPath.join("/"))) return bad("not_found", 404);

  let parsedBody: unknown;
  try {
    parsedBody = await request.json();
  } catch {
    return bad("error_occurred", 400);
  }
  // isRecord() also returns true for arrays, so reject arrays explicitly.
  if (!isRecord(parsedBody) || Array.isArray(parsedBody)) return bad("error_occurred", 400);

  try {
    const mainApiUrl = getMainApiUrl(getBackendUrlFromRequest(request));
    if (FILE_PATHS.includes(goPath.join("/"))) {
      return await proxyFile(request, authorization, `${mainApiUrl}${toGoApiPath(goPath)}`, parsedBody);
    }
    return await proxyMainApiJson(
      request,
      mainApiUrl,
      toGoApiPath(goPath),
      { method: "POST", body: JSON.stringify(parsedBody) },
    );
  } catch {
    return bad("connection_error", 500);
  }
}

// proxyFile - สำเร็จ = ไฟล์จาก backend (PDF หรือ text/plain) ส่งต่อทั้ง byte + Content-Type/Content-Disposition เดิม;
// ผิดพลาด = JSON ของ backend ทั้งก้อน (code/field/row/message และ issues/total ของไฟล์ยื่นด้วยสื่อ)
// user error (4xx มี code) ส่งเป็น 200 + success:false เหมือน proxyMainApiJson เพื่อไม่ให้ browser log error
async function proxyFile(request: Request, authorization: string, url: string, body: unknown): Promise<NextResponse> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 20000);
  try {
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
      },
      body: JSON.stringify(body),
      signal: controller.signal,
      cache: "no-store",
    });
    const contentType = response.headers.get("content-type") ?? "";
    if (response.ok && FILE_CONTENT_TYPES.some((type) => contentType.toLowerCase().startsWith(type))) {
      // arrayBuffer เท่านั้น — .text() จะตัด BOM ทิ้ง
      return new NextResponse(await response.arrayBuffer(), {
        status: 200,
        headers: {
          "Content-Type": contentType,
          "Content-Disposition": response.headers.get("content-disposition") ?? "inline",
          "Cache-Control": "no-store",
        },
      });
    }
    const payload = await readJsonOrText(response);
    const record = isRecord(payload) ? payload : { message: String(payload ?? "") };
    const userError = response.status >= 400 && response.status < 500 && response.status !== 401 && response.status !== 403 && typeof record.code === "string";
    return NextResponse.json({ ...record, success: false }, { status: userError ? 200 : response.status });
  } catch {
    return bad("connection_error", 504);
  } finally {
    clearTimeout(timeout);
  }
}
