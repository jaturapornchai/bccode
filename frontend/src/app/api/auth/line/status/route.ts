import { NextResponse } from "next/server";
import {
  extractMessage,
  getAuthBridgeUrl,
  getRecord,
  getString,
  isRecord,
  postMainApiAuth,
  readJsonOrText,
} from "@/lib/auth-bridge";
import { serverMainApiBase, validateBackendUrl } from "@/lib/backend-url";

type LineStatusBody = {
  backendUrl?: string;
  code?: string;
};

export async function POST(request: Request) {
  let body: LineStatusBody;
  try {
    body = (await request.json()) as LineStatusBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const code = body.code?.trim() ?? "";
  if (!code) {
    return NextResponse.json({ success: false, message: "ไม่พบ LINE login code" }, { status: 400 });
  }

  let mainApiUrl: string;
  let normalizedGoApiUrl: string;
  try {
    const checked = validateBackendUrl(body.backendUrl ?? "");
    mainApiUrl = checked.mainApiUrl;
    normalizedGoApiUrl = checked.normalizedGoApiUrl;
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  try {
    const bridgeUrl = getAuthBridgeUrl();
    const response = await fetch(`${bridgeUrl}/api/login?code=${encodeURIComponent(code)}`, {
      headers: { "Content-Type": "application/json" },
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);

    if (!response.ok || !isRecord(payload)) {
      return NextResponse.json(
        { success: false, status: "failed", message: extractMessage(payload) ?? "ตรวจสอบ LINE login ไม่สำเร็จ" },
        { status: response.ok ? 502 : response.status },
      );
    }

    const confirmed = payload.confirmed === true;
    if (!confirmed) {
      return NextResponse.json({ success: true, status: "pending" });
    }

    const data = getRecord(payload, "data") ?? {};
    const lineUserId = getString(data, "userId") ?? getString(data, "lineuserid") ?? getString(data, "lineUserId");
    if (!lineUserId) {
      return NextResponse.json({ success: false, status: "failed", message: "LINE login ไม่มี user id" }, { status: 502 });
    }

    const displayName = getString(data, "displayName") ?? getString(data, "displayname") ?? "";
    const pictureUrl = getString(data, "pictureUrl") ?? getString(data, "pictureurl") ?? "";
    const email = getString(data, "email") ?? "";
    const login = await postMainApiAuth(serverMainApiBase(), "/linelogin", {
      lineuserid: lineUserId,
      displayname: displayName,
      pictureurl: pictureUrl,
      email,
    });

    return NextResponse.json({
      success: true,
      status: "success",
      token: login.token,
      refresh: login.refresh,
      user: {
        username: login.username || email || displayName,
        email,
        name: displayName,
        pictureUrl,
      },
      backendUrl: normalizedGoApiUrl,
      mainApiUrl,
    });
  } catch (error) {
    return NextResponse.json(
      { success: false, status: "failed", message: error instanceof Error ? error.message : "LINE login ไม่สำเร็จ" },
      { status: 504 },
    );
  }
}
