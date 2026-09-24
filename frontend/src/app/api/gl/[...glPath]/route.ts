import { NextResponse } from "next/server";
import { getBackendUrlFromRequest, getMainApiUrl, isRecord, proxyMainApiJson, requireBearerToken } from "@/lib/workspace-api";
import { GL_REPORTS, GL_RESOURCES } from "@/lib/general-ledger";

type Context = { params: Promise<{ glPath: string[] }> };
const bad = (message: string, status = 400) => NextResponse.json({ success: false, message }, { status });
export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;
  const { glPath } = await context.params;
  const valid = glPath[0] === "reports"
    ? glPath.length === 2 && (GL_REPORTS as readonly string[]).includes(glPath[1])
    : glPath[0] === "journal-support" ? glPath.length === 1
    : glPath[0] === "journal-reviews" ? glPath.length === 2
    : (GL_RESOURCES as readonly string[]).includes(glPath[0]) && glPath.length <= 2;
  if (!valid || glPath.some((segment) => !/^[\p{L}\p{M}\p{N}_.-]+$/u.test(segment) || segment === "." || segment === "..")) return bad("ไม่พบรายการที่ต้องการ", 404);
  try {
    const query = new URLSearchParams();
    const allowed = ["q", "page", "limit", "from", "to", "fiscalyear", "accountcode", "branchcode", "departmentcode", "projectcode", "bookcode", "status", "kind", "snapshot", "asof", "companywide"];
    new URL(request.url).searchParams.forEach((value, key) => { if (allowed.includes(key)) query.set(key, value); });
    return proxyMainApiJson(request, getMainApiUrl(getBackendUrlFromRequest(request)), `/gl/v2/${glPath.map(encodeURIComponent).join("/")}?${query}`, { method: "GET" });
  } catch { return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่"); }
}
export async function POST(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;
  if ((await context.params).glPath.join("/") !== "command") return bad("ไม่พบรายการที่ต้องการ", 404);
  let body: unknown;
  try { body = await request.json(); } catch { return bad("รูปแบบข้อมูลไม่ถูกต้อง"); }
  if (!isRecord(body) || Array.isArray(body) || typeof body.requestid !== "string" || !/^[0-9a-f-]{36}$/i.test(body.requestid)) return bad("กรุณาระบุหมายเลขคำขอที่ถูกต้อง");
  if (![...GL_RESOURCES, "processes"].includes(body.resource as never) || !["create", "update", "delete", "post", "reverse", "lock", "unlock", "close", "year-end", "recalculate", "reprocess", "review", "reconcile"].includes(String(body.action))) return bad("ประเภทการทำรายการไม่ถูกต้อง");
  if (["review", "reconcile"].includes(String(body.action)) && body.resource !== "journals") return bad("ประเภทการทำรายการไม่ถูกต้อง");
  if (hasNumericAmount(body)) return bad("จำนวนเงินต้องส่งเป็นข้อความทศนิยม");
  // Scope is resolved from the authenticated backend session. Never forward a caller's company/holding override.
  const payload: Record<string, unknown> = {};
  for (const key of ["resource", "id", "action", "requestid", "version", "reason", "date", "docno", "targetyear", "account", "fiscalyear", "master", "journal", "statementtemplate", "review"]) {
    if (!(key in body)) continue;
    if (isRecord(body[key])) {
      const record = { ...body[key] };
      for (const field of ["holdingcode", "businesscode", "companycode", "createdby", "updatedby", "createdat", "updatedat", "isdeleted"]) delete record[field];
      payload[key] = record;
    } else payload[key] = body[key];
  }
  try { return proxyMainApiJson(request, getMainApiUrl(getBackendUrlFromRequest(request)), "/gl/v2/command", { method: "POST", body: JSON.stringify(payload) }, { userErrorStatusOk: true }); }
  catch { return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่"); }
}

function hasNumericAmount(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(hasNumericAmount);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, child]) => ["amount", "debit", "credit", "balance_after"].includes(key) ? typeof child !== "string" : hasNumericAmount(child));
}
