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

type GoogleStatusBody = {
  backendUrl?: string;
  sessionId?: string;
};

export async function POST(request: Request) {
  let body: GoogleStatusBody;
  try {
    body = (await request.json()) as GoogleStatusBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const sessionId = body.sessionId?.trim() ?? "";
  if (!sessionId) {
    return NextResponse.json({ success: false, message: "ไม่พบ Google session" }, { status: 400 });
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
    const response = await fetch(`${bridgeUrl}/api/google/session/${encodeURIComponent(sessionId)}/status`, {
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);

    if (!response.ok || !isRecord(payload)) {
      return NextResponse.json(
        { success: false, status: "failed", message: extractMessage(payload) ?? "ตรวจสอบ Google login ไม่สำเร็จ" },
        { status: response.ok ? 502 : response.status },
      );
    }

    const data = getRecord(payload, "data") ?? payload;
    const status = getString(data, "status") ?? "pending";
    if (status !== "success") {
      return NextResponse.json({ success: true, status, message: extractMessage(payload) });
    }

    const user = getRecord(data, "user") ?? {};
    const email = getString(user, "email");
    if (!email) {
      return NextResponse.json({ success: false, status: "failed", message: "Google login ไม่มี email" }, { status: 502 });
    }

    const login = await postMainApiAuth(serverMainApiBase(), "/googlelogin", {
      google_user_id: getString(user, "google_user_id") ?? getString(user, "googleUserId") ?? "",
      display_name: getString(user, "display_name") ?? getString(user, "displayName") ?? "",
      picture_url: getString(user, "picture_url") ?? getString(user, "pictureUrl") ?? "",
      email,
    });

    return NextResponse.json({
      success: true,
      status: "success",
      token: login.token,
      refresh: login.refresh,
      user: {
        username: login.username || email,
        email,
        name: getString(user, "display_name") ?? getString(user, "displayName") ?? "",
        pictureUrl: getString(user, "picture_url") ?? getString(user, "pictureUrl") ?? "",
      },
      backendUrl: normalizedGoApiUrl,
      mainApiUrl,
    });
  } catch (error) {
    return NextResponse.json(
      { success: false, status: "failed", message: error instanceof Error ? error.message : "Google login ไม่สำเร็จ" },
      { status: 504 },
    );
  }
}
