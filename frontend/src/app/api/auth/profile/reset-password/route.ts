import { NextResponse } from "next/server";
import { getBackendUrlFromRequest, getMainApiUrl, getString, isRecord, proxyMainApiJson, type ApiProxyBody } from "@/lib/workspace-api";

export async function PUT(request: Request) {
  let body: ApiProxyBody;
  try {
    body = await request.json() as ApiProxyBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const username = isRecord(body) ? getString(body, "username")?.trim() : "";
  if (!username) return NextResponse.json({ success: false, message: "username invalid" }, { status: 400 });

  const backendUrl = getBackendUrlFromRequest(request, body);
  if (!backendUrl) return NextResponse.json({ success: false, message: "ไม่พบ Backend URL" }, { status: 400 });

  return proxyMainApiJson(
    request,
    getMainApiUrl(backendUrl),
    `/profile/password/reset/${encodeURIComponent(username)}`,
    { method: "PUT", body: JSON.stringify({}) },
  );
}
