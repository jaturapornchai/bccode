import { NextResponse } from "next/server";
import { sanitizeBackendLanguageDictionary } from "@/lib/backend-language-sanitize";
import { serverGoApiBase, validateBackendUrl } from "@/lib/backend-url";

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

  try {
    validateBackendUrl(url.searchParams.get("backendUrl") ?? "");
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 8000);

  try {
    const response = await fetch(`${serverGoApiBase()}/api/language/${encodeURIComponent(language)}`, {
      cache: "no-store",
      signal: controller.signal,
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok || !data || typeof data !== "object" || Array.isArray(data)) {
      return NextResponse.json({ success: false, message: "โหลดภาษาไม่สำเร็จ" }, { status: 502 });
    }
    return NextResponse.json(sanitizeBackendLanguageDictionary(data as Record<string, string>), { status: 200 });
  } catch {
    return NextResponse.json({ success: false, message: "เชื่อมต่อ backend language API ไม่สำเร็จ" }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}
