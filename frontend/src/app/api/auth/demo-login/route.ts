import { NextResponse } from "next/server";
import { setRefreshTokenCookie } from "@/lib/auth-session-server";
import { serverMainApiBase } from "@/lib/backend-url";
import { getString, isRecord, readJsonOrText } from "@/lib/workspace-api";

/**
 * Demo login (ปุ่ม "ทดลองใช้ระบบ"): signs in as the shared demo account that only
 * holds sample data. Works on local and the public server alike; the backend
 * decides availability (BCAI_DEMO_LOGIN_ENABLED) so nothing here is secret.
 */
export async function POST() {
  let backendBase = "";
  try {
    backendBase = serverMainApiBase();
  } catch {
    backendBase = "";
  }
  if (!backendBase) {
    return NextResponse.json(
      { success: false, errorcode: "DEMO_LOGIN_DISABLED", message: "Demo ยังไม่พร้อมใช้งาน" },
      { status: 503 },
    );
  }
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 15000);
  try {
    const response = await fetch(`${backendBase}/demo-login`, {
      method: "POST",
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);
    if (!response.ok) {
      return NextResponse.json(
        {
          success: false,
          errorcode: "DEMO_LOGIN_DISABLED",
          message: response.status === 404 ? "ระบบนี้ไม่ได้เปิดบัญชี Demo" : "Demo ยังไม่พร้อมใช้งาน",
        },
        { status: response.status },
      );
    }
    const token = isRecord(payload) ? getString(payload, "token") : undefined;
    const refresh = isRecord(payload) ? getString(payload, "refresh") : undefined;
    if (!token || !refresh) {
      return NextResponse.json(
        { success: false, errorcode: "DEMO_LOGIN_DISABLED", message: "Server ตอบกลับไม่ครบ" },
        { status: 502 },
      );
    }
    return setRefreshTokenCookie(NextResponse.json({ success: true, token, user: "demo" }), refresh);
  } catch (error) {
    const message =
      error instanceof Error && error.name === "AbortError" ? "Server ไม่ตอบกลับทันเวลา" : "ไม่สามารถเชื่อมต่อ Server ได้";
    return NextResponse.json({ success: false, errorcode: "DEMO_LOGIN_DISABLED", message }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}
