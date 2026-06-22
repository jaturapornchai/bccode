import { NextResponse } from "next/server";
import { serverGoApiBase, validateBackendUrl } from "@/lib/backend-url";

export async function POST(request: Request) {
  const body = (await request.json().catch(() => ({}))) as { backendUrl?: string };

  try {
    validateBackendUrl(body.backendUrl ?? "");
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 8000);

  try {
    const result = await probeBackend(controller.signal);
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

async function probeBackend(signal: AbortSignal): Promise<string> {
  // Probe the backend the app actually proxies to (the next.config rewrite target),
  // not the public same-origin URL — this server-side process cannot resolve the public host.
  const base = serverGoApiBase();
  const candidates = [`${base}/api/health`, `${base}/version`];

  for (const url of candidates) {
    const response = await fetch(url, { signal, cache: "no-store" }).catch(() => null);
    if (response?.ok) return url;
  }

  throw new Error("backend probe failed");
}
