import { NextResponse } from "next/server";
import { serverGoApiBase, validateBackendUrl } from "@/lib/backend-url";
import { getJwtClaimHoldingCode, verifyHs256Jwt } from "@/lib/server-jwt";
import {
  extractMessage,
  isRecord,
  readJsonOrText,
  requireBearerToken,
} from "@/lib/workspace-api";

type LineOaUserBody = {
  action?: "link" | "profile";
  backendUrl?: string;
  holdingcode?: string;
  holdingCode?: string;
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

  const holdingCode = (body.holdingcode ?? body.holdingCode ?? "").trim();
  const username = body.username?.trim() ?? "";
  if (!holdingCode || !username) {
    return NextResponse.json({ success: false, message: "ไม่พบข้อมูลบริษัทหรือผู้ใช้" }, { status: 400 });
  }

  const tenantResponse = validateTenantAccess(authorization, holdingCode);
  if (tenantResponse) return tenantResponse;

  try {
    validateBackendUrl(body.backendUrl ?? "");
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 20000);

  try {
    const response = await fetch(`${serverGoApiBase()}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
      },
      body: JSON.stringify({ holdingcode: holdingCode, username }),
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

function validateTenantAccess(authorization: string, requestedHoldingCode: string): NextResponse | null {
  if (!process.env.JWT_SECRET_KEY?.trim()) return null;

  const jwt = verifyHs256Jwt(authorization);
  if (!jwt.ok) return NextResponse.json({ success: false, message: jwt.message }, { status: jwt.status });

  const claimHoldingCode = getJwtClaimHoldingCode(jwt.claims);
  if (!claimHoldingCode) return NextResponse.json({ success: false, message: "token ไม่มีรหัส holding" }, { status: 401 });
  if (claimHoldingCode !== requestedHoldingCode) {
    return NextResponse.json({ success: false, message: "ไม่มีสิทธิ์เข้าถึงข้อมูลบริษัทนี้" }, { status: 403 });
  }
  return null;
}
