import { NextResponse } from "next/server";
import { validateBackendUrl } from "@/lib/backend-url";

type LanguageProxyContext = {
  params: Promise<{ lang: string }>;
};

export async function GET(request: Request, context: LanguageProxyContext) {
  const { lang } = await context.params;
  const language = lang.trim().toLowerCase();
  const url = new URL(request.url);

  if (!language) {
    return NextResponse.json({ success: false, message: "กรุณาระบุภาษา" }, { status: 400 });
  }

  let normalizedGoApiUrl: string;
  try {
    normalizedGoApiUrl = validateBackendUrl(url.searchParams.get("backendUrl") ?? "").normalizedGoApiUrl;
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 8000);

  try {
    const response = await fetch(`${normalizedGoApiUrl}/api/language/${encodeURIComponent(language)}`, {
      cache: "no-store",
      signal: controller.signal,
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok || !data || typeof data !== "object" || Array.isArray(data)) {
      return NextResponse.json({ success: false, message: "โหลดภาษาไม่สำเร็จ" }, { status: 502 });
    }
    return NextResponse.json(data, { status: 200 });
  } catch {
    return NextResponse.json({ success: false, message: "เชื่อมต่อ backend language API ไม่สำเร็จ" }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}
