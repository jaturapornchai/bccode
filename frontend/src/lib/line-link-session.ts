import { NextResponse } from "next/server";
import { extractMessage, readJsonOrText } from "@/lib/auth-bridge";
import { serverMainApiBase } from "@/lib/backend-url";

// Mainapi guards for the LINE link BFF routes (api/auth/line/code, api/auth/line/link/status).
// requireBearerToken only checks the header shape, so a forged "Bearer x" would pass; these calls let
// mainapi's session middleware (opaque token looked up in cache_entries, then live session/access checked)
// decide before the routes touch the auth bridge. A LINE link code is bound to the user who minted it
// (mainapi POST /profile/link-line/code) and checked before every bridge poll (/code/check).

/** Live session check via mainapi GET /verify-token; null = session OK, otherwise the response to return. */
export function verifyLineSession(request: Request, authorization: string): Promise<NextResponse | null> {
  return callMainApi(request, authorization, "/verify-token");
}

/** Binds a freshly minted bridge code to the signed-in user (first binding wins). */
export function bindLineLinkCode(request: Request, authorization: string, code: string): Promise<NextResponse | null> {
  return callMainApi(request, authorization, "/profile/link-line/code", { code });
}

/** Passes only when the code was minted by the signed-in user and has not expired or been used. */
export function checkLineLinkCode(request: Request, authorization: string, code: string): Promise<NextResponse | null> {
  return callMainApi(request, authorization, "/profile/link-line/code/check", { code });
}

async function callMainApi(
  request: Request,
  authorization: string,
  path: string,
  body?: Record<string, string>,
): Promise<NextResponse | null> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 15000);

  try {
    const response = await fetch(`${serverMainApiBase()}${path}`, {
      method: body ? "POST" : "GET",
      headers: {
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
        ...(body ? { "Content-Type": "application/json" } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
      signal: controller.signal,
      cache: "no-store",
    });
    if (response.ok) return null;
    if (response.status === 401 || response.status === 403) {
      // Same body as requireBearerToken so authFetch refreshes the token and retries once.
      return NextResponse.json({ success: false, message: "ไม่พบ token กรุณาเข้าสู่ระบบใหม่" }, { status: 401 });
    }
    if (body && (response.status === 400 || response.status === 404 || response.status === 409)) {
      // The code itself was refused (invalid, another user's, expired or already used): mainapi's message
      // is already translated from Accept-Language.
      const payload = await readJsonOrText(response).catch(() => undefined);
      return NextResponse.json(
        { success: false, status: "failed", message: extractMessage(payload) ?? "เชื่อมต่อ LINE ไม่สำเร็จ" },
        { status: response.status },
      );
    }
    return NextResponse.json({ success: false, message: "ไม่สามารถเชื่อมต่อ Server ได้" }, { status: 502 });
  } catch {
    return NextResponse.json({ success: false, message: "ไม่สามารถเชื่อมต่อ Server ได้" }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}
