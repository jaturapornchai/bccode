import { NextResponse } from "next/server";
import { validateBackendUrl } from "@/lib/backend-url";
import { holdingCodeValidationMessageTh, isValidHoldingCode, normalizeHoldingCode } from "@/lib/holding-code";

type LoginBody = {
  backendUrl?: string;
  username?: string;
  password?: string;
  holdingcode?: string;
};

export async function POST(request: Request) {
  let body: LoginBody;
  try {
    body = (await request.json()) as LoginBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const username = body.username?.trim() ?? "";
  const password = body.password ?? "";
  // holdingcode is optional — backend /login accepts it but does not require it.
  // When omitted, the user will pick a holding on the holding selection screen.
  const holdingCode = body.holdingcode ? normalizeHoldingCode(body.holdingcode ?? "") : "";

  if (!username) {
    return NextResponse.json({ success: false, message: "กรุณากรอกชื่อผู้ใช้" }, { status: 400 });
  }
  if (!password) {
    return NextResponse.json({ success: false, message: "กรุณากรอกรหัสผ่าน" }, { status: 400 });
  }
  if (body.holdingcode && !holdingCode) {
    return NextResponse.json(
      { success: false, message: "รูปแบบรหัสกลุ่มกิจการไม่ถูกต้อง" },
      { status: 400 },
    );
  }
  if (holdingCode && !isValidHoldingCode(holdingCode)) {
    return NextResponse.json(
      { success: false, message: holdingCodeValidationMessageTh },
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
      body: JSON.stringify({ username, password, holdingcode: holdingCode }),
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
