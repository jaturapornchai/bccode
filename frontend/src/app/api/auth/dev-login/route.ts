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
  const backendBase = devLoginBackendBase();
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
  // The browser must be on a loopback origin (docs: Dev Login is loopback-only)
  // and the Origin header must match the Host actually served — i.e. a
  // same-origin call. Both loopback spellings (localhost / 127.0.0.1 / ::1)
  // are accepted so the origin registered in the Google Cloud Console for
  // Google Login works for Dev Login too.
  const originHeader = request.headers.get("origin");
  if (!originHeader) return false;
  try {
    const origin = new URL(originHeader);
    if (origin.protocol !== "http:" && origin.protocol !== "https:") return false;
    if (!isLoopbackHostname(origin.hostname)) return false;
    return request.headers.get("host") === origin.host;
  } catch {
    return false;
  }
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

// Server-side-only override for container deployments where the backend is
// reachable by docker-network hostname instead of host loopback. Explicitly
// set per environment; absent by default so the loopback rule still applies.
function devLoginBackendBase(): string | undefined {
  const override = process.env.BCAI_DEV_LOGIN_BACKEND_URL?.trim();
  if (!override) {
    return loopbackBackendBase();
  }
  try {
    const parsed = new URL(override);
    if (
      (parsed.protocol !== "http:" && parsed.protocol !== "https:") ||
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
