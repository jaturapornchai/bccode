import { NextResponse } from "next/server";
import { extractMessage, getAuthBridgeUrl, getRecord, getString, isRecord, readJsonOrText } from "@/lib/auth-bridge";

export async function POST() {
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
