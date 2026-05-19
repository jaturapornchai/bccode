import { NextResponse } from "next/server";
import { getBackendUrlFromRequest, getMainApiUrl, isRecord, readJsonOrText, requireBearerToken } from "@/lib/workspace-api";

export async function POST(request: Request) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  let mainApiUrl: string;
  try {
    mainApiUrl = getMainApiUrl(getBackendUrlFromRequest(request));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const form = await request.formData().catch(() => null);
  const file = form?.get("file");
  if (!(file instanceof File)) {
    return NextResponse.json({ success: false, message: "ไม่พบไฟล์รูป" }, { status: 400 });
  }

  const uploadForm = new FormData();
  uploadForm.append("file", file, file.name || "image.webp");

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 30000);

  try {
    const response = await fetch(`${mainApiUrl}/upload/images`, {
      method: "POST",
      headers: {
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
      },
      body: uploadForm,
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);
    if (isRecord(payload)) return NextResponse.json(payload, { status: response.status });
    return NextResponse.json({ success: response.ok, message: String(payload ?? "") }, { status: response.status });
  } catch (error) {
    const message =
      error instanceof Error && error.name === "AbortError"
        ? "Server ไม่ตอบกลับทันเวลา"
        : "อัปโหลดรูปไม่สำเร็จ";
    return NextResponse.json({ success: false, message }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}
