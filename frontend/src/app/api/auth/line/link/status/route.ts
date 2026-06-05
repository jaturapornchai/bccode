import { NextResponse } from "next/server";
import {
  extractMessage as extractBridgeMessage,
  getAuthBridgeUrl,
  getRecord,
  getString,
  isRecord as isBridgeRecord,
  readJsonOrText as readBridgeJsonOrText,
} from "@/lib/auth-bridge";
import { validateBackendUrl } from "@/lib/backend-url";
import {
  extractMessage,
  isRecord,
  readJsonOrText,
  requireBearerToken,
} from "@/lib/workspace-api";

type LineLinkStatusBody = {
  backendUrl?: string;
  code?: string;
};

export async function POST(request: Request) {
  let body: LineLinkStatusBody;
  try {
    body = (await request.json()) as LineLinkStatusBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const code = body.code?.trim() ?? "";
  if (!code) {
    return NextResponse.json({ success: false, message: "ไม่พบ LINE login code" }, { status: 400 });
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
    const bridgeUrl = getAuthBridgeUrl();
    const response = await fetch(`${bridgeUrl}/api/login?code=${encodeURIComponent(code)}`, {
      headers: { "Content-Type": "application/json" },
      cache: "no-store",
    });
    const payload = await readBridgeJsonOrText(response);

    if (!response.ok || !isBridgeRecord(payload)) {
      return NextResponse.json(
        { success: false, status: "failed", message: extractBridgeMessage(payload) ?? "ตรวจสอบ LINE ไม่สำเร็จ" },
        { status: response.ok ? 502 : response.status },
      );
    }

    if (payload.confirmed !== true) {
      return NextResponse.json({ success: true, status: "pending" });
    }

    const data = getRecord(payload, "data") ?? {};
    const lineUserId = getString(data, "userId") ?? getString(data, "lineuserid") ?? getString(data, "lineUserId");
    if (!lineUserId) {
      return NextResponse.json({ success: false, status: "failed", message: "LINE login ไม่มี user id" }, { status: 502 });
    }

    const displayName = getString(data, "displayName") ?? getString(data, "displayname") ?? "";
    const pictureUrl = getString(data, "pictureUrl") ?? getString(data, "pictureurl") ?? "";
    const linkResponse = await putLineProfile(request, mainApiUrl, authorization, {
      lineuserid: lineUserId,
      linedisplayname: displayName,
      linepictureurl: pictureUrl,
    });

    if (!linkResponse.success) {
      return NextResponse.json(
        { success: false, status: "failed", message: linkResponse.message },
        { status: linkResponse.status },
      );
    }

    return NextResponse.json({
      success: true,
      status: "success",
      backendUrl: normalizedGoApiUrl,
      mainApiUrl,
      user: {
        lineUserId,
        displayName,
        pictureUrl,
      },
    });
  } catch (error) {
    return NextResponse.json(
      { success: false, status: "failed", message: error instanceof Error ? error.message : "เชื่อมต่อ LINE ไม่สำเร็จ" },
      { status: 504 },
    );
  }
}

async function putLineProfile(
  request: Request,
  mainApiUrl: string,
  authorization: string,
  payload: Record<string, string>,
): Promise<{ success: boolean; message?: string; status: number }> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 15000);

  try {
    const response = await fetch(`${mainApiUrl}/profile/link-line`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
      },
      body: JSON.stringify(payload),
      signal: controller.signal,
      cache: "no-store",
    });
    const responsePayload = await readJsonOrText(response);
    if (!response.ok) {
      return {
        success: false,
        message: extractMessage(responsePayload) ?? `Server ตอบกลับผิดปกติ (${response.status})`,
        status: response.status,
      };
    }

    if (isRecord(responsePayload) && responsePayload.success === false) {
      return {
        success: false,
        message: extractMessage(responsePayload) ?? "เชื่อมต่อ LINE ไม่สำเร็จ",
        status: 400,
      };
    }

    return { success: true, status: response.status };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error && error.name === "AbortError" ? "Server ไม่ตอบกลับทันเวลา" : "ไม่สามารถเชื่อมต่อ Server ได้",
      status: 504,
    };
  } finally {
    clearTimeout(timeout);
  }
}
