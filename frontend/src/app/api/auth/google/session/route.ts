import { NextResponse } from "next/server";
import { getAuthBridgeUrl, getNumber, getRecord, getString, isRecord, readJsonOrText } from "@/lib/auth-bridge";

type GoogleSessionBody = {
  deviceInfo?: string;
};

export async function POST(request: Request) {
  let body: GoogleSessionBody;
  try {
    body = (await request.json()) as GoogleSessionBody;
  } catch {
    body = {};
  }

  try {
    const bridgeUrl = getAuthBridgeUrl();
    const response = await fetch(`${bridgeUrl}/api/google/session`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ device_info: body.deviceInfo ?? "web" }),
      cache: "no-store",
    });

    const payload = await readJsonOrText(response);
    if (!response.ok || !isRecord(payload)) {
      return NextResponse.json(
        { success: false, message: "ไม่สามารถสร้าง Google login session ได้" },
        { status: response.ok ? 502 : response.status },
      );
    }

    const data = getRecord(payload, "data") ?? payload;
    const sessionId = getString(data, "session_id") ?? getString(data, "sessionId");
    const loginUrl = getString(data, "login_url") ?? getString(data, "loginUrl");
    const pollInterval = getNumber(data, "poll_interval") ?? getNumber(data, "pollInterval") ?? 2000;

    if (!sessionId || !loginUrl) {
      return NextResponse.json({ success: false, message: "Google login session ตอบกลับไม่ครบ" }, { status: 502 });
    }

    return NextResponse.json({
      success: true,
      sessionId,
      loginUrl,
      pollInterval,
    });
  } catch {
    return NextResponse.json({ success: false, message: "ไม่สามารถเชื่อมต่อ Google login service ได้" }, { status: 504 });
  }
}
