import { NextResponse } from "next/server";
import { validateBackendUrl } from "@/lib/backend-url";

type LoginBody = {
  backendUrl?: string;
  username?: string;
  password?: string;
  holding_code?: string;
};

const holdingCodePattern = /^[a-z][a-z0-9_]{2,29}$/;

export async function POST(request: Request) {
  let body: LoginBody;
  try {
    body = (await request.json()) as LoginBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const username = body.username?.trim() ?? "";
  const password = body.password ?? "";
  const holdingCode = body.holding_code?.trim().toLowerCase() ?? "";

  if (!username) {
    return NextResponse.json({ success: false, message: "กรุณากรอกชื่อผู้ใช้" }, { status: 400 });
  }
  if (!password) {
    return NextResponse.json({ success: false, message: "กรุณากรอกรหัสผ่าน" }, { status: 400 });
  }
  if (!holdingCode) {
    return NextResponse.json(
      { success: false, message: "กรุณากรอกรหัส Holding ก่อนเข้าสู่ระบบด้วย User , Password" },
      { status: 400 },
    );
  }
  if (!holdingCodePattern.test(holdingCode)) {
    return NextResponse.json(
      { success: false, message: "holding_code ต้องเป็น a-z, 0-9, _ ยาว 3-30 ตัว และขึ้นต้นด้วย a-z" },
      { status: 400 },
    );
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

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 15000);

  try {
    const response = await fetch(`${mainApiUrl}/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password, holding_code: holdingCode }),
      signal: controller.signal,
      cache: "no-store",
    });

    const payload = await readJsonOrText(response);
    if (!response.ok) {
      return NextResponse.json(
        {
          success: false,
          message: extractMessage(payload) ?? `Server ตอบกลับผิดปกติ (${response.status})`,
        },
        { status: response.status },
      );
    }

    if (!isRecord(payload)) {
      return NextResponse.json({ success: false, message: "Server ตอบกลับไม่ถูกต้อง" }, { status: 502 });
    }

    const token = getString(payload, "token") ?? getNestedString(payload, "data", "token");
    const refresh = getString(payload, "refresh") ?? getNestedString(payload, "data", "refresh");
    const success = payload.success === true || Boolean(token);

    if (!success) {
      return NextResponse.json(
        { success: false, message: extractMessage(payload) ?? "เข้าสู่ระบบไม่สำเร็จ" },
        { status: 401 },
      );
    }

    return NextResponse.json({
      success: true,
      user: username,
      token,
      refresh,
      backendUrl: normalizedGoApiUrl,
      mainApiUrl,
    });
  } catch (error) {
    const message = error instanceof Error && error.name === "AbortError"
      ? "Server ไม่ตอบกลับทันเวลา"
      : "ไม่สามารถเชื่อมต่อ Server ได้";
    return NextResponse.json({ success: false, message }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}

async function readJsonOrText(response: Response): Promise<unknown> {
  const contentType = response.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) {
    return response.json();
  }
  return response.text();
}

function extractMessage(payload: unknown): string | undefined {
  if (typeof payload === "string" && payload.trim()) return payload;
  if (!isRecord(payload)) return undefined;
  return getString(payload, "message") ?? getNestedString(payload, "error", "message");
}

function getNestedString(payload: Record<string, unknown>, key: string, nestedKey: string): string | undefined {
  const nested = payload[key];
  if (!isRecord(nested)) return undefined;
  return getString(nested, nestedKey);
}

function getString(payload: Record<string, unknown>, key: string): string | undefined {
  const value = payload[key];
  return typeof value === "string" ? value : undefined;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
