import { NextResponse } from "next/server";

export const REFRESH_TOKEN_COOKIE = "bc_refresh_token";
export const REFRESH_TOKEN_MAX_AGE_SECONDS = 12 * 60 * 60;

export function getRefreshToken(request: Request): string {
  const cookieHeader = request.headers.get("cookie") ?? "";
  for (const part of cookieHeader.split(";")) {
    const [rawName, ...rawValue] = part.trim().split("=");
    if (rawName === REFRESH_TOKEN_COOKIE) {
      try {
        return decodeURIComponent(rawValue.join("=")).trim();
      } catch {
        return "";
      }
    }
  }
  return "";
}

export function setRefreshTokenCookie(response: NextResponse, token: string): NextResponse {
  response.cookies.set(REFRESH_TOKEN_COOKIE, token, {
    httpOnly: true,
    secure: true,
    sameSite: "lax",
    path: "/",
    maxAge: REFRESH_TOKEN_MAX_AGE_SECONDS,
  });
  return response;
}

export function clearRefreshTokenCookie(response: NextResponse): NextResponse {
  response.cookies.set(REFRESH_TOKEN_COOKIE, "", {
    httpOnly: true,
    secure: true,
    sameSite: "lax",
    path: "/",
    maxAge: 0,
  });
  return response;
}
