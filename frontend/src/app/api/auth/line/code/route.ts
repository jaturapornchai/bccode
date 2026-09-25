import { NextResponse } from "next/server";
import { extractMessage, getAuthBridgeUrl, getRecord, getString, isRecord, readJsonOrText } from "@/lib/auth-bridge";
import { bindLineLinkCode, verifyLineSession } from "@/lib/line-link-session";
import { requireBearerToken } from "@/lib/workspace-api";

// Mints a LINE link code for the signed-in user (ปุ่ม "เชื่อมต่อ LINE" in menu/workspace).
// The route takes no input: the bridge URL comes only from BC_AUTH_BRIDGE_URL, the request body is
// never read or forwarded, and the user's token goes to mainapi only — never to the bridge. The code is
// bound to this user in mainapi before it is returned, so only this user can poll or complete it.
export async function POST(request: Request) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const sessionError = await verifyLineSession(request, authorization);
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

    const bindError = await bindLineLinkCode(request, authorization, code);
    if (bindError) return bindError;

    return NextResponse.json({ success: true, code, loginUrl, expiresAt });
  } catch {
    return NextResponse.json({ success: false, message: "ไม่สามารถเชื่อมต่อ LINE login service ได้" }, { status: 504 });
  }
}
