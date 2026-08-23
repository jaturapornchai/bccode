import { NextResponse } from "next/server";
import { serverMainApiBase } from "@/lib/backend-url";
import {
  clearRefreshTokenCookie,
  getRefreshToken,
  setRefreshTokenCookie,
} from "@/lib/auth-session-server";
import { extractMessage, getString, isRecord, readJsonOrText } from "@/lib/workspace-api";

export async function POST(request: Request) {
  const refreshToken = getRefreshToken(request);
  if (!refreshToken) {
    return clearRefreshTokenCookie(
      NextResponse.json({ success: false, message: "กรุณาเข้าสู่ระบบอีกครั้ง" }, { status: 401 }),
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 15000);
  try {
    const response = await fetch(`${serverMainApiBase()}/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token: refreshToken }),
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);
    const token = isRecord(payload) ? getString(payload, "token") : undefined;
    const rotatedRefresh = isRecord(payload) ? getString(payload, "refresh") : undefined;

    if (!response.ok || !token || !rotatedRefresh) {
      const errorResponse = NextResponse.json(
        { success: false, message: extractMessage(payload) ?? "กรุณาเข้าสู่ระบบอีกครั้ง" },
        { status: response.ok ? 502 : response.status },
      );
      return response.status === 401 || response.status === 403
        ? clearRefreshTokenCookie(errorResponse)
        : errorResponse;
    }

    return setRefreshTokenCookie(
      NextResponse.json({ success: true, token }),
      rotatedRefresh,
    );
  } catch (error) {
    const message = error instanceof Error && error.name === "AbortError"
      ? "Server ไม่ตอบกลับทันเวลา"
      : "ไม่สามารถเชื่อมต่อ Server ได้";
    return NextResponse.json({ success: false, message }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}
