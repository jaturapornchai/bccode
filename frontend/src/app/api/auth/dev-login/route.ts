import { NextResponse } from "next/server";
import { setRefreshTokenCookie } from "@/lib/auth-session-server";
import { isLoopbackHostname, serverMainApiBase } from "@/lib/backend-url";
import { getString, isRecord, readJsonOrText } from "@/lib/workspace-api";

const DEV_LOGIN_HEADER = "X-BC-Dev-Login-Secret";

export async function POST(request: Request) {
  if (process.env.BCAI_DEV_LOGIN_ENABLED !== "true") {
    return NextResponse.json(
      { success: false, errorcode: "DEV_LOGIN_DISABLED", message: "ไม่พบเส้นทาง" },
      { status: 404 },
    );
  }
  if (!isExactLoopbackRequest(request)) {
    return NextResponse.json(
      { success: false, errorcode: "DEV_LOGIN_FORBIDDEN", message: "Dev Login ใช้ได้เฉพาะ localhost" },
      { status: 403 },
    );
  }

  const secret = process.env.BCAI_DEV_LOGIN_SECRET ?? "";
  const backendBase = loopbackBackendBase();
  if (secret.length < 32 || !backendBase) {
    return NextResponse.json(
      { success: false, errorcode: "DEV_LOGIN_DISABLED", message: "Dev Login ยังไม่พร้อมใช้งาน" },
      { status: 503 },
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 15000);
  try {
    const response = await fetch(`${backendBase}/dev-login`, {
      method: "POST",
      headers: { [DEV_LOGIN_HEADER]: secret },
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);
    if (!response.ok) {
      const forbidden = response.status === 401 || response.status === 403;
      return NextResponse.json(
        {
          success: false,
          errorcode: forbidden ? "DEV_LOGIN_FORBIDDEN" : "DEV_LOGIN_DISABLED",
          message: forbidden ? "ไม่อนุญาตให้ใช้ Dev Login" : "Dev Login ยังไม่พร้อมใช้งาน",
        },
        { status: response.status },
      );
    }

    const token = isRecord(payload) ? getString(payload, "token") : undefined;
    const refresh = isRecord(payload) ? getString(payload, "refresh") : undefined;
    if (!token || !refresh) {
      return NextResponse.json(
        { success: false, errorcode: "DEV_LOGIN_DISABLED", message: "Server ตอบกลับไม่ครบ" },
        { status: 502 },
      );
    }

    return setRefreshTokenCookie(
      NextResponse.json({ success: true, token, user: "dev" }),
      refresh,
    );
  } catch (error) {
    const message = error instanceof Error && error.name === "AbortError"
      ? "Server ไม่ตอบกลับทันเวลา"
      : "ไม่สามารถเชื่อมต่อ Server ได้";
    return NextResponse.json(
      { success: false, errorcode: "DEV_LOGIN_DISABLED", message },
      { status: 504 },
    );
  } finally {
    clearTimeout(timeout);
  }
}

function isExactLoopbackRequest(request: Request): boolean {
  const requestUrl = new URL(request.url);
  return isLoopbackHostname(requestUrl.hostname) &&
    request.headers.get("host") === requestUrl.host &&
    request.headers.get("origin") === requestUrl.origin;
}

function loopbackBackendBase(): string | undefined {
  try {
    const raw = serverMainApiBase();
    const parsed = new URL(raw);
    if (
      (parsed.protocol !== "http:" && parsed.protocol !== "https:") ||
      !isLoopbackHostname(parsed.hostname) ||
      parsed.username ||
      parsed.password ||
      parsed.pathname !== "/" ||
      parsed.search ||
      parsed.hash
    ) {
      return undefined;
    }
    return parsed.origin;
  } catch {
    return undefined;
  }
}
