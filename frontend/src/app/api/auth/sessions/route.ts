import { NextResponse } from "next/server";
import { getBackendUrlFromRequest, getMainApiUrl, proxyMainApiJson } from "@/lib/workspace-api";

/** GET /api/auth/sessions — จำนวนเซสชันที่กำลังใช้งานระบบ (อ่านจาก Redis ฝั่ง backend) */
export async function GET(request: Request) {
  const backendUrl = getBackendUrlFromRequest(request);
  if (!backendUrl) return NextResponse.json({ success: false, message: "ไม่พบ Backend URL" }, { status: 400 });
  return proxyMainApiJson(request, getMainApiUrl(backendUrl), "/sessions/active-count", { method: "GET" });
}
