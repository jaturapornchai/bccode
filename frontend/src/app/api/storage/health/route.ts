import { NextResponse } from "next/server";

// Server-side S3/MinIO endpoint health probe.
// The browser cannot fetch cross-origin S3 endpoints directly (CORS), so this
// route performs the probe server-to-server and returns the result.
type Body = {
  endpoint?: string;
};

export async function POST(request: Request) {
  let body: Body;
  try {
    body = (await request.json()) as Body;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  const endpoint = (body.endpoint ?? "").trim();
  if (!endpoint) {
    return NextResponse.json(
      { success: false, message: "กรุณาระบุ S3 Endpoint" },
      { status: 400 },
    );
  }

  const url = /^https?:\/\//i.test(endpoint) ? endpoint : `http://${endpoint}`;
  const start = Date.now();

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 8000);

  try {
    const response = await fetch(url, {
      method: "GET",
      redirect: "manual",
      signal: controller.signal,
      cache: "no-store",
    });
    clearTimeout(timeout);
    const latencyMs = Date.now() - start;
    // For S3/MinIO any HTTP response (even 403/400 from the root path) means the
    // endpoint is reachable. Only network errors (catch block) count as failure.
    return NextResponse.json({
      success: true,
      message: "เชื่อมต่อได้",
      latencyMs,
      httpStatus: response.status,
    });
  } catch (error) {
    clearTimeout(timeout);
    const latencyMs = Date.now() - start;
    const message = error instanceof Error ? error.message : "เชื่อมต่อไม่ได้";
    return NextResponse.json(
      { success: false, message, latencyMs },
      { status: 200 }, // 200 with success:false so the client can read the message
    );
  }
}
