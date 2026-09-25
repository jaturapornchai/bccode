import { NextResponse } from "next/server";
import { extractMessage, getAuthBridgeUrl, getRecord, getString, isRecord, readJsonOrText } from "@/lib/auth-bridge";
import { serverMainApiBase } from "@/lib/backend-url";
import { requireBearerToken } from "@/lib/workspace-api";

// Mints a LINE link code for the signed-in user (ปุ่ม "เชื่อมต่อ LINE" in menu/workspace).
// The route takes no input: the bridge URL comes only from BC_AUTH_BRIDGE_URL, the request body is
// never read or forwarded, and the user's token goes to mainapi only — never to the bridge.
export async function POST(request: Request) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const sessionError = await verifySession(request, authorization);
  if (sessionError) return sessionError;

  try {
    const bridgeUrl = getAuthBridgeUrl();
    const response = await fetch(`${bridgeUrl}/api/login?action=create`, {
      headers: { "Content-Type": "application/json" },
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);

    if (!response.ok || !isRecord(payload)) {
      return NextResponse.json(
        { success: false, message: extractMessage(payload) ?? "ไม่สามารถสร้าง LINE login code ได้" },
        { status: response.ok ? 502 : response.status },
      );
    }

    const data = getRecord(payload, "data") ?? payload;
    const code = getString(data, "code");
    const loginUrl = getString(data, "loginUrl") ?? getString(data, "login_url");
    const expiresAt = getString(data, "expiresAt") ?? getString(data, "expiresat");

    if (!code || !loginUrl) {
      return NextResponse.json({ success: false, message: "LINE login service ตอบกลับไม่ครบ" }, { status: 502 });
    }

    return NextResponse.json({ success: true, code, loginUrl, expiresAt });
  } catch {
    return NextResponse.json({ success: false, message: "ไม่สามารถเชื่อมต่อ LINE login service ได้" }, { status: 504 });
  }
}

// requireBearerToken only checks the header shape and this route has no other mainapi call, so a
// forged "Bearer x" would pass. Ask mainapi GET /verify-token (the same session middleware as every
// protected API: the opaque token is looked up in cache_entries, then the live session/access is
// checked) before minting a code.
async function verifySession(request: Request, authorization: string): Promise<NextResponse | null> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 15000);

  try {
    const response = await fetch(`${serverMainApiBase()}/verify-token`, {
      headers: {
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
      },
      signal: controller.signal,
      cache: "no-store",
    });
    if (response.ok) return null;
    if (response.status === 401 || response.status === 403) {
      // Same body as requireBearerToken so authFetch refreshes the token and retries once.
      return NextResponse.json({ success: false, message: "ไม่พบ token กรุณาเข้าสู่ระบบใหม่" }, { status: 401 });
    }
    return NextResponse.json({ success: false, message: "ไม่สามารถเชื่อมต่อ Server ได้" }, { status: 502 });
  } catch {
    return NextResponse.json({ success: false, message: "ไม่สามารถเชื่อมต่อ Server ได้" }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}
