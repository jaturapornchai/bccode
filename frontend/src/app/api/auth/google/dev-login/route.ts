import { NextResponse } from "next/server";
import { postMainApiAuth } from "@/lib/auth-bridge";
import { validateBackendUrl } from "@/lib/backend-url";
import {
  isLocalLoginRequest,
  LOCAL_GOOGLE_TEST_EMAIL,
  LOCAL_GOOGLE_TEST_NAME,
} from "@/lib/local-dev-auth";

type DevGoogleLoginBody = {
  backendUrl?: string;
};

export async function GET(request: Request) {
  const enabled = canUseLocalGoogleTestLogin(request);
  return NextResponse.json({
    enabled,
    email: enabled ? LOCAL_GOOGLE_TEST_EMAIL : undefined,
  });
}

export async function POST(request: Request) {
  if (!canUseLocalGoogleTestLogin(request)) {
    return NextResponse.json({ success: false, message: "ปุ่มทดสอบนี้ใช้ได้เฉพาะเครื่อง local เท่านั้น" }, { status: 403 });
  }

  let body: DevGoogleLoginBody;
  try {
    body = (await request.json()) as DevGoogleLoginBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
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
    const login = await postMainApiAuth(mainApiUrl, "/googlelogin", {
      google_user_id: `local-test:${LOCAL_GOOGLE_TEST_EMAIL}`,
      display_name: LOCAL_GOOGLE_TEST_NAME,
      picture_url: "",
      email: LOCAL_GOOGLE_TEST_EMAIL,
    });

    return NextResponse.json({
      success: true,
      status: "success",
      token: login.token,
      refresh: login.refresh,
      user: {
        username: login.username || LOCAL_GOOGLE_TEST_EMAIL,
        email: LOCAL_GOOGLE_TEST_EMAIL,
        name: LOCAL_GOOGLE_TEST_NAME,
        pictureUrl: "",
      },
      backendUrl: normalizedGoApiUrl,
      mainApiUrl,
    });
  } catch (error) {
    return NextResponse.json(
      { success: false, status: "failed", message: error instanceof Error ? error.message : "Google test login ไม่สำเร็จ" },
      { status: 504 },
    );
  }
}

function canUseLocalGoogleTestLogin(request: Request): boolean {
  return process.env.NODE_ENV !== "production" && isLocalLoginRequest(request);
}
