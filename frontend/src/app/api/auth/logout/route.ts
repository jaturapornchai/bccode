import { NextResponse } from "next/server";
import { serverMainApiBase } from "@/lib/backend-url";
import { clearRefreshTokenCookie, getRefreshToken, setRefreshTokenCookie } from "@/lib/auth-session-server";
import { getString, isRecord, readJsonOrText } from "@/lib/workspace-api";

export async function POST(request: Request) {
  const authorization = request.headers.get("authorization") ?? "";
  let rotatedRefreshToken = "";
  const refreshToken = getRefreshToken(request);
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 15000);

  try {
    const bearerAvailable = authorization.toLowerCase().startsWith("bearer ");
    if (bearerAvailable) {
      const logoutResponse = await revoke(authorization, controller.signal);
      if (logoutResponse.ok) {
        return clearRefreshTokenCookie(NextResponse.json({ success: true }));
      }
      if ((logoutResponse.status !== 401 && logoutResponse.status !== 403) || !refreshToken) {
        return failureResponse(logoutResponse.status);
      }
    }

    if (refreshToken) {
      const refreshResponse = await fetch(`${serverMainApiBase()}/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token: refreshToken }),
        signal: controller.signal,
        cache: "no-store",
      });
      const refreshPayload = await readJsonOrText(refreshResponse);
      const accessToken = isRecord(refreshPayload) ? getString(refreshPayload, "token") : undefined;
      rotatedRefreshToken = isRecord(refreshPayload) ? getString(refreshPayload, "refresh") ?? "" : "";
      if (!refreshResponse.ok || !accessToken || !rotatedRefreshToken) {
        return failureResponse(refreshResponse.ok ? 502 : refreshResponse.status);
      }

      const logoutResponse = await revoke(`Bearer ${accessToken}`, controller.signal);
      if (!logoutResponse.ok) {
        return setRefreshTokenCookie(failureResponse(logoutResponse.status), rotatedRefreshToken);
      }
      return clearRefreshTokenCookie(NextResponse.json({ success: true }));
    }

    return failureResponse(401);
  } catch {
    const failure = failureResponse(504);
    return rotatedRefreshToken ? setRefreshTokenCookie(failure, rotatedRefreshToken) : failure;
  } finally {
    clearTimeout(timeout);
  }
}

function revoke(authorization: string, signal: AbortSignal): Promise<Response> {
  return fetch(`${serverMainApiBase()}/logout`, {
    method: "POST",
    headers: { Authorization: authorization },
    signal,
    cache: "no-store",
  });
}

function failureResponse(status: number): NextResponse {
  return NextResponse.json(
    { success: false, message: "ไม่สามารถเพิกถอน Session ได้" },
    { status },
  );
}
