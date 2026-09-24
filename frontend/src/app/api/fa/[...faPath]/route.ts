import { NextResponse } from "next/server";
import { getBackendUrlFromRequest, getMainApiUrl, isRecord, proxyMainApiJson, requireBearerToken } from "@/lib/workspace-api";

type Context = { params: Promise<{ faPath: string[] }> };
const bad = (message: string, status = 400) => NextResponse.json({ success: false, message }, { status });

const ALLOWED_RESOURCES = ["assets", "types", "reports", "command", "mcp"];
const ALLOWED_REPORTS = ["schedule", "tax-reconciliation"];

export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { faPath } = await context.params;
  if (!faPath || faPath.length === 0) return bad("ไม่พบรายการที่ต้องการ", 404);

  const resource = faPath[0];
  if (!ALLOWED_RESOURCES.includes(resource)) return bad("ไม่พบรายการที่ต้องการ", 404);

  if (resource === "reports" && (!faPath[1] || !ALLOWED_REPORTS.includes(faPath[1]))) {
    return bad("ไม่พบรายงานที่ต้องการ", 404);
  }

  // Segment safety check
  // \p{M} = สระบน/ล่างและวรรณยุกต์ไทย (รหัสทรัพย์สินไทย) — docs/kms/17-dev-gotchas.md
  if (faPath.some((segment) => !/^[\p{L}\p{M}\p{N}_.-]+$/u.test(segment) || segment === "." || segment === "..")) {
    return bad("ไม่พบรายการที่ต้องการ", 404);
  }

  try {
    const query = new URLSearchParams();
    const allowedParams = ["q", "page", "limit", "type", "status", "fiscalyear", "period"];
    new URL(request.url).searchParams.forEach((val, key) => {
      if (allowedParams.includes(key)) query.set(key, val);
    });

    const targetPath = `/fa/v2/${faPath.map(encodeURIComponent).join("/")}?${query.toString()}`;
    return proxyMainApiJson(request, getMainApiUrl(getBackendUrlFromRequest(request)), targetPath, { method: "GET" });
  } catch {
    return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่");
  }
}

export async function POST(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { faPath } = await context.params;
  const pathStr = faPath.join("/");

  if (pathStr !== "command" && pathStr !== "mcp") {
    return bad("ไม่พบรายการที่ต้องการ", 404);
  }

  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return bad("รูปแบบข้อมูลไม่ถูกต้อง");
  }

  if (!isRecord(body) || Array.isArray(body)) {
    return bad("ข้อมูลต้องอยู่ในรูปแบบ JSON Object");
  }

  // Handle MCP RPC Call
  if (pathStr === "mcp") {
    try {
      return proxyMainApiJson(
        request,
        getMainApiUrl(getBackendUrlFromRequest(request)),
        "/fa/v2/mcp",
        { method: "POST", body: JSON.stringify(body) },
        { userErrorStatusOk: true }
      );
    } catch {
      return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่");
    }
  }

  // Strip tenant overrides
  const payload: Record<string, unknown> = {};
  for (const [key, val] of Object.entries(body)) {
    if (isRecord(val)) {
      const record = { ...val };
      for (const field of ["holdingcode", "businesscode", "companycode", "createdby", "updatedby", "createdat", "updatedat", "isdeleted"]) {
        delete record[field];
      }
      payload[key] = record;
    } else {
      payload[key] = val;
    }
  }

  try {
    return proxyMainApiJson(
      request,
      getMainApiUrl(getBackendUrlFromRequest(request)),
      "/fa/v2/command",
      { method: "POST", body: JSON.stringify(payload) },
      { userErrorStatusOk: true }
    );
  } catch {
    return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่");
  }
}
