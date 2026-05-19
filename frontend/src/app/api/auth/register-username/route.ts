import { NextResponse } from "next/server";
import { validateBackendUrl } from "@/lib/backend-url";

type RegisterUsernameBody = {
  backendUrl?: string;
  username?: string;
  password?: string;
  name?: string;
};

export async function POST(request: Request) {
  let body: RegisterUsernameBody;
  try {
    body = (await request.json()) as RegisterUsernameBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const username = body.username?.trim() ?? "";
  const password = body.password ?? "";
  const name = body.name?.trim() ?? "";

  if (!username) {
    return NextResponse.json({ success: false, message: "กรุณากรอกชื่อผู้ใช้" }, { status: 400 });
  }
  if (!password) {
    return NextResponse.json({ success: false, message: "กรุณากรอกรหัสผ่าน" }, { status: 400 });
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
    const response = await fetch(`${mainApiUrl}/register-username`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password, name }),
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);

    if (!response.ok) {
      return NextResponse.json(
        { success: false, message: extractMessage(payload) ?? `Server ตอบกลับผิดปกติ (${response.status})` },
        { status: response.status },
      );
    }

    if (!isRecord(payload)) {
      return NextResponse.json({ success: false, message: "Server ตอบกลับไม่ถูกต้อง" }, { status: 502 });
    }

    return NextResponse.json({
      success: payload.success === true,
      id: getString(payload, "id") ?? getString(payload, "ID"),
      backendUrl: normalizedGoApiUrl,
      mainApiUrl,
    }, { status: 201 });
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
  if (contentType.includes("application/json")) return response.json();
  return response.text();
}

function extractMessage(payload: unknown): string | undefined {
  if (typeof payload === "string" && payload.trim()) return payload;
  if (!isRecord(payload)) return undefined;
  const error = payload.error;
  if (isRecord(error)) return getString(error, "message") ?? getString(error, "error");
  return getString(payload, "message") ?? getString(payload, "error");
}

function getString(payload: Record<string, unknown>, key: string): string | undefined {
  const value = payload[key];
  return typeof value === "string" ? value : undefined;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
