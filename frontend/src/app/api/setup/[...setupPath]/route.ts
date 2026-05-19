import { NextResponse } from "next/server";
import { validateBackendUrl } from "@/lib/backend-url";

const allowedSetupPaths = new Set([
  "verify-password",
  "change-password",
  "config/get",
  "config/get-raw",
  "config/save",
  "config/seed",
  "test-connection",
  "create-clickhouse-database",
]);

type SetupProxyBody = Record<string, unknown> & {
  backendUrl?: unknown;
};

type SetupProxyContext = {
  params: Promise<{ setupPath: string[] }>;
};

export async function POST(request: Request, context: SetupProxyContext) {
  const { setupPath } = await context.params;
  const path = setupPath.join("/");
  if (!allowedSetupPaths.has(path)) {
    return NextResponse.json({ success: false, message: "ไม่พบ setup endpoint" }, { status: 404 });
  }

  let body: SetupProxyBody;
  try {
    body = (await request.json()) as SetupProxyBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  let normalizedGoApiUrl: string;
  try {
    normalizedGoApiUrl = validateBackendUrl(String(body.backendUrl ?? "")).normalizedGoApiUrl;
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const { backendUrl: _backendUrl, ...payload } = body;
  void _backendUrl;

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 20000);

  try {
    const response = await fetch(`${normalizedGoApiUrl}/api/setup/${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      signal: controller.signal,
      cache: "no-store",
    });
    const data = await readJsonOrText(response);

    if (isRecord(data)) {
      return NextResponse.json(data, { status: response.status });
    }

    return NextResponse.json({ success: response.ok, message: String(data ?? "") }, { status: response.status });
  } catch (error) {
    const message =
      error instanceof Error && error.name === "AbortError"
        ? "Server ไม่ตอบกลับทันเวลา"
        : "ไม่สามารถเชื่อมต่อ setup backend ได้";
    return NextResponse.json({ success: false, message }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}

async function readJsonOrText(response: Response): Promise<unknown> {
  const contentType = response.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) return response.json();
  return response.text();
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
