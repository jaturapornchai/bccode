import { NextResponse } from "next/server";
import { getBackendUrlFromRequest, getMainApiUrl, isRecord, proxyMainApiJson, readJsonOrText, requireBearerToken } from "@/lib/workspace-api";

// Central report registry relay (backend/internal/report, GET /report/v1/registry, POST /report/v1/run|export).
// Scope (holding/company/branch) always comes from the authenticated backend session: no body key can override it.
type Context = { params: Promise<{ reportPath: string[] }> };
const bad = (message: string, status = 400) => NextResponse.json({ success: false, message }, { status });
// A report may scan a whole year of journals; the backend enforces its own timeout (REPORT_TIMEOUT), this only bounds the relay
const REPORT_TIMEOUT_MS = 120_000;
const MENU_ID = /^[a-z0-9][a-z0-9_-]{0,99}$/i;
const PARAM_KEY = /^[a-z][a-z0-9_]{0,49}$/;
const SORT_VALUE = /^[a-z0-9_.-]{0,50}$/i;
const MODULE = /^[a-z][a-z_]{0,19}$/;
const LANG = /^[a-z]{2,3}$/;
// the backend accepts up to 200 comma-separated codes of 50 characters (report params.go); anything longer is malformed
const MAX_PARAM_LENGTH = 12_000;

export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;
  const { reportPath } = await context.params;
  if (reportPath.length !== 1 || reportPath[0] !== "registry") return bad("ไม่พบรายการที่ต้องการ", 404);
  const incoming = new URL(request.url).searchParams;
  const query = new URLSearchParams();
  const moduleCode = incoming.get("module");
  const lang = incoming.get("lang");
  if (moduleCode !== null) { if (!MODULE.test(moduleCode)) return bad("รูปแบบข้อมูลไม่ถูกต้อง"); query.set("module", moduleCode); }
  if (lang !== null) { if (!LANG.test(lang)) return bad("รูปแบบข้อมูลไม่ถูกต้อง"); query.set("lang", lang); }
  try {
    return await proxyMainApiJson(request, getMainApiUrl(getBackendUrlFromRequest(request)), `/report/v1/registry?${query}`, { method: "GET" }, { userErrorStatusOk: true, timeoutMs: REPORT_TIMEOUT_MS });
  } catch { return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่"); }
}

export async function POST(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;
  const { reportPath } = await context.params;
  const action = reportPath.length === 1 ? reportPath[0] : "";
  if (action !== "run" && action !== "export") return bad("ไม่พบรายการที่ต้องการ", 404);
  let body: unknown;
  try { body = await request.json(); } catch { return bad("รูปแบบข้อมูลไม่ถูกต้อง"); }
  const payload = runPayload(body, action);
  if (typeof payload === "string") return bad(payload);
  try {
    const mainApiUrl = getMainApiUrl(getBackendUrlFromRequest(request));
    if (action === "run") return await proxyMainApiJson(request, mainApiUrl, "/report/v1/run", { method: "POST", body: JSON.stringify(payload) }, { userErrorStatusOk: true, timeoutMs: REPORT_TIMEOUT_MS });
    return await relayExport(request, authorization, mainApiUrl, payload);
  } catch { return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่"); }
}

/** Keeps only the documented run keys with the documented types; returns a Thai error text when the body is malformed. */
function runPayload(body: unknown, action: "run" | "export"): Record<string, unknown> | string {
  const invalid = "รูปแบบข้อมูลไม่ถูกต้อง";
  if (!isRecord(body) || Array.isArray(body)) return invalid;
  if (typeof body.menu_id !== "string" || !MENU_ID.test(body.menu_id)) return invalid;
  const payload: Record<string, unknown> = { menu_id: body.menu_id };
  if (body.params !== undefined) {
    if (!isRecord(body.params) || Array.isArray(body.params)) return invalid;
    const params: Record<string, string> = {};
    for (const [key, value] of Object.entries(body.params)) {
      // every parameter travels as a string ("true"/"false", comma-separated codes); a JS number would lose decimal/leading-zero meaning
      if (!PARAM_KEY.test(key) || typeof value !== "string" || value.length > MAX_PARAM_LENGTH) return invalid;
      params[key] = value;
    }
    payload.params = params;
  }
  if (body.sort !== undefined) {
    if (typeof body.sort !== "string" || !SORT_VALUE.test(body.sort)) return invalid;
    payload.sort = body.sort;
  }
  for (const key of ["page", "limit", "rows_per_page"]) {
    if (body[key] === undefined) continue;
    if (typeof body[key] !== "number" || !Number.isInteger(body[key])) return invalid;
    payload[key] = body[key];
  }
  if (body.print !== undefined) {
    if (typeof body.print !== "boolean") return invalid;
    payload.print = body.print;
  }
  if (action === "export") {
    if (body.format !== undefined && body.format !== "csv") return invalid;
    payload.format = "csv";
  }
  return payload;
}

/** CSV is streamed untouched (Content-Type / Content-Disposition / UTF-8 BOM); a JSON error is relayed like proxyMainApiJson does. */
async function relayExport(request: Request, authorization: string, mainApiUrl: string, payload: Record<string, unknown>): Promise<Response> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), REPORT_TIMEOUT_MS);
  try {
    const response = await fetch(`${mainApiUrl}/report/v1/export`, {
      method: "POST",
      headers: { "Content-Type": "application/json", "Accept-Language": request.headers.get("accept-language") ?? "th", Authorization: authorization },
      body: JSON.stringify(payload),
      signal: controller.signal,
      cache: "no-store",
    });
    const type = response.headers.get("content-type") ?? "";
    if (response.ok && type.includes("text/csv") && response.body) {
      const headers = new Headers({ "Content-Type": type, "Cache-Control": "no-store" });
      const disposition = response.headers.get("content-disposition");
      if (disposition) headers.set("Content-Disposition", disposition);
      return new Response(response.body, { status: 200, headers });
    }
    const error = await readJsonOrText(response);
    if (isRecord(error)) {
      // same rule as proxyMainApiJson(userErrorStatusOk): a 4xx with a machine code is a form error, not a console error
      const userError = response.status >= 400 && response.status < 500 && response.status !== 401 && response.status !== 403 && typeof error.code === "string" && error.code.length > 0;
      return NextResponse.json(userError ? { ...error, success: false } : error, { status: userError ? 200 : response.status });
    }
    return NextResponse.json({ success: false, message: String(error ?? "") }, { status: response.ok ? 502 : response.status });
  } catch (cause) {
    const message = cause instanceof Error && cause.name === "AbortError" ? "Server ไม่ตอบกลับทันเวลา" : "ไม่สามารถเชื่อมต่อ Server ได้";
    return NextResponse.json({ success: false, message }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}
