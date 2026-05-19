import { NextResponse } from "next/server";
import { validateBackendUrl } from "@/lib/backend-url";
import {
  extractMessage,
  isRecord,
  readJsonOrText,
  requireBearerToken,
} from "@/lib/workspace-api";

type LineOaUserBody = {
  action?: "link" | "profile";
  backendUrl?: string;
  shopId?: string;
  shop_id?: string;
  username?: string;
};

const actionPaths = {
  link: "/api/user/lineoa/link",
  profile: "/api/user/lineoa/profile",
} as const;

export async function POST(request: Request) {
  let body: LineOaUserBody;
  try {
    body = (await request.json()) as LineOaUserBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const action = body.action ?? "profile";
  const path = actionPaths[action];
  if (!path) {
    return NextResponse.json({ success: false, message: "LINE OA action ไม่ถูกต้อง" }, { status: 400 });
  }

  const shopId = (body.shopId ?? body.shop_id ?? "").trim();
  const username = body.username?.trim() ?? "";
  if (!shopId || !username) {
    return NextResponse.json({ success: false, message: "ไม่พบข้อมูลกิจการหรือผู้ใช้" }, { status: 400 });
  }

  let goApiUrl: string;
  try {
    goApiUrl = validateBackendUrl(body.backendUrl ?? "").normalizedGoApiUrl;
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 20000);

  try {
    const response = await fetch(`${goApiUrl}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
      },
      body: JSON.stringify({ shop_id: shopId, username }),
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);
    if (!isRecord(payload)) {
      return NextResponse.json({ success: response.ok, message: String(payload ?? "") }, { status: response.status });
    }

    if (!response.ok) {
      return NextResponse.json(
        { ...payload, success: false, message: extractMessage(payload) ?? `Server ตอบกลับผิดปกติ (${response.status})` },
        { status: response.status },
      );
    }

    return NextResponse.json({ success: payload.status !== "error", ...payload }, { status: response.status });
  } catch (error) {
    const message =
      error instanceof Error && error.name === "AbortError"
        ? "Server ไม่ตอบกลับทันเวลา"
        : "ไม่สามารถเชื่อมต่อ Server ได้";
    return NextResponse.json({ success: false, message }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}
