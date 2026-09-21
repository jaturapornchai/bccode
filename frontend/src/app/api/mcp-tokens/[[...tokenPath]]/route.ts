import { NextResponse } from "next/server";
import { getBackendUrlFromRequest, getMainApiUrl, isRecord, proxyMainApiJson, requireBearerToken } from "@/lib/workspace-api";

type Context = { params: Promise<{ tokenPath?: string[] }> };
const bad = (message: string, status = 400) => NextResponse.json({ success: false, message }, { status, headers: { "Cache-Control": "no-store" } });
async function proxy(request: Request, path: string, init: RequestInit) {
  try {
    const response = await proxyMainApiJson(request, getMainApiUrl(getBackendUrlFromRequest(request)), path, init);
    response.headers.set("Cache-Control", "no-store, private");
    return response;
  } catch { return bad("การเชื่อมต่อระบบไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่"); }
}
export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;
  const path = (await context.params).tokenPath ?? [];
  if (path.length === 1 && path[0] === "companies") return proxy(request, "/mcp-tokens/companies", { method: "GET" });
  if (path.length) return bad("ไม่พบรายการที่ต้องการ", 404);
  return proxy(request, "/mcp-tokens", { method: "GET" });
}
export async function POST(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;
  const path = (await context.params).tokenPath ?? [];
  if (path.length === 2 && /^[0-9a-f]{32}$/.test(path[0]) && path[1] === "revoke") {
    return proxy(request, `/mcp-tokens/${encodeURIComponent(path[0])}/revoke`, { method: "POST", body: "{}" });
  }
  if (path.length) return bad("ไม่พบรายการที่ต้องการ", 404);
  let body: unknown;
  try { body = await request.json(); } catch { return bad("รูปแบบข้อมูลไม่ถูกต้อง"); }
  if (!isRecord(body) || Array.isArray(body) || typeof body.name !== "string" || !body.name.trim() || body.name.trim().length > 100 || !["api", "mcp"].includes(String(body.kind)) || !["readonly", "readwrite"].includes(String(body.mode)) || typeof body.expiresAt !== "string" || !Number.isFinite(Date.parse(body.expiresAt))) return bad("กรุณาระบุชื่อ ประเภท สิทธิ์ และวันหมดอายุให้ถูกต้อง");
  if (!Array.isArray(body.companyCodes) || !body.companyCodes.length || body.companyCodes.some((code) => typeof code !== "string" || !code.trim() || code !== code.trim() || code === "*") || new Set(body.companyCodes).size !== body.companyCodes.length) return bad("กรุณาเลือกบริษัทที่อนุญาตอย่างน้อยหนึ่งบริษัท");
  // Holding/admin access and every allowed company are validated by the backend.
  return proxy(request, "/mcp-tokens", { method: "POST", body: JSON.stringify({ name: body.name.trim(), kind: body.kind, mode: body.mode, companyCodes: body.companyCodes, expiresAt: body.expiresAt }) });
}
