import { NextResponse } from "next/server";
import { postMainApiAuth } from "@/lib/auth-bridge";
import { setRefreshTokenCookie } from "@/lib/auth-session-server";
import { serverMainApiBase } from "@/lib/backend-url";

// Real Google Sign-In (Google Identity Services).
// The browser obtains a Google ID token (JWT) from the "Sign in with Google" button,
// posts it here, and this route VERIFIES it with Google before trusting any email.
// Only after verification do we call mainapi /googlelogin. Mainapi verifies the same raw
// credential again and binds the account by issuer+subject, so neither layer trusts email
// supplied by the browser.

type VerifyBody = {
  credential?: string;
};

type GoogleTokenInfo = {
  aud?: string;
  iss?: string;
  exp?: string;
  email?: string;
  email_verified?: string | boolean;
  name?: string;
  picture?: string;
  sub?: string;
};

function configuredClientId(): string | undefined {
  return process.env.GOOGLE_CLIENT_ID ?? process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID;
}

export async function POST(request: Request) {
  let body: VerifyBody;
  try {
    body = (await request.json()) as VerifyBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const credential = body.credential?.trim();
  if (!credential) {
    return NextResponse.json({ success: false, message: "ไม่พบ Google credential" }, { status: 400 });
  }

  const clientId = configuredClientId();
  if (!clientId) {
    return NextResponse.json({ success: false, message: "ระบบยังไม่ได้ตั้งค่า Google Client ID" }, { status: 500 });
  }

  // Verify the ID token with Google. tokeninfo checks the signature for us; we then
  // enforce audience, issuer, expiry and verified-email ourselves.
  let info: GoogleTokenInfo;
  try {
    const response = await fetch(
      `https://oauth2.googleapis.com/tokeninfo?id_token=${encodeURIComponent(credential)}`,
      { cache: "no-store" },
    );
    if (!response.ok) {
      return NextResponse.json({ success: false, message: "ตรวจสอบ Google token ไม่สำเร็จ" }, { status: 401 });
    }
    info = (await response.json()) as GoogleTokenInfo;
  } catch {
    return NextResponse.json({ success: false, message: "เชื่อมต่อ Google เพื่อตรวจ token ไม่ได้" }, { status: 504 });
  }

  if (info.aud !== clientId) {
    return NextResponse.json({ success: false, message: "Google token ไม่ตรงกับแอปนี้" }, { status: 401 });
  }
  if (info.iss !== "accounts.google.com" && info.iss !== "https://accounts.google.com") {
    return NextResponse.json({ success: false, message: "ผู้ออก Google token ไม่ถูกต้อง" }, { status: 401 });
  }
  const expSeconds = Number(info.exp);
  if (!Number.isFinite(expSeconds) || expSeconds * 1000 <= Date.now()) {
    return NextResponse.json({ success: false, message: "Google token หมดอายุ" }, { status: 401 });
  }
  const emailVerified = info.email_verified === true || info.email_verified === "true";
  if (!info.email || !emailVerified) {
    return NextResponse.json({ success: false, message: "อีเมล Google ยังไม่ได้ยืนยัน" }, { status: 401 });
  }

  try {
    const login = await postMainApiAuth(serverMainApiBase(), "/googlelogin", {
      email: info.email,
      displayname: info.name ?? "",
      pictureurl: info.picture ?? "",
      googleuserid: info.sub ?? "",
      // Forward the raw Google ID token so mainapi re-verifies it server-side
      // (it no longer trusts the email field — defense in depth).
      credential,
    });

    if (!login.refresh) {
      return NextResponse.json({ success: false, status: "failed", message: "เข้าสู่ระบบไม่สำเร็จ" }, { status: 502 });
    }

    const response = NextResponse.json({
      success: true,
      status: "success",
      token: login.token,
      user: {
        username: login.username || info.email,
        email: info.email,
        name: info.name ?? "",
        pictureUrl: info.picture ?? "",
      },
    });
    return setRefreshTokenCookie(response, login.refresh);
  } catch (error) {
    return NextResponse.json(
      { success: false, status: "failed", message: error instanceof Error ? error.message : "เข้าสู่ระบบไม่สำเร็จ" },
      { status: 502 },
    );
  }
}
