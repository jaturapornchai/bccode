import { NextResponse } from "next/server";
import { validateBackendUrl } from "@/lib/backend-url";

export async function POST(request: Request) {
  const body = (await request.json().catch(() => ({}))) as { backendUrl?: string };

  let normalizedGoApiUrl: string;
  try {
    normalizedGoApiUrl = validateBackendUrl(body.backendUrl ?? "").normalizedGoApiUrl;
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 8000);

  try {
    const result = await probeBackend(normalizedGoApiUrl, controller.signal);
    return NextResponse.json({ success: true, endpoint: result });
  } catch {
    return NextResponse.json(
      { success: false, message: "ไม่สามารถเชื่อมต่อ Server ได้ กรุณาตรวจสอบ Backend URL" },
      { status: 504 },
    );
  } finally {
    clearTimeout(timeout);
  }
}

async function probeBackend(goApiUrl: string, signal: AbortSignal): Promise<string> {
  const candidates = [`${goApiUrl}/api/health`, `${goApiUrl}/version`];

  for (const url of candidates) {
    const response = await fetch(url, { signal, cache: "no-store" }).catch(() => null);
    if (response?.ok) return url;
  }

  throw new Error("backend probe failed");
}
