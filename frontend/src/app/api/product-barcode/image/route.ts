import { NextResponse } from "next/server";
import { getBackendUrlFromRequest, getMainApiUrl, requireBearerToken } from "@/lib/workspace-api";

/**
 * Image upload proxy.
 *
 * Frontend sends `multipart/form-data` with field `file`.
 * Backend mainapi route: `POST /upload/images`.
 * Pass-through the multipart body to backend with Authorization preserved.
 *
 * Note: this proxy treats the file as an opaque blob — it doesn't read content
 * into memory beyond `body` streaming. Multipart Content-Type header includes
 * the boundary, so we forward `request.headers.get("content-type")` directly.
 */
export async function POST(request: Request) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  let base: string;
  try {
    base = getMainApiUrl(getBackendUrlFromRequest(request));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 60000);

  try {
    const upstreamResponse = await fetch(`${base}/upload/images`, {
      method: "POST",
      headers: {
        Authorization: authorization,
        ...(request.headers.get("content-type")
          ? { "Content-Type": request.headers.get("content-type") as string }
          : {}),
      },
      body: request.body,
      // @ts-expect-error -- Node-only fetch option needed to stream FormData passthrough
      duplex: "half",
      signal: controller.signal,
    });

    const responseText = await upstreamResponse.text();
    try {
      const json = JSON.parse(responseText) as Record<string, unknown>;
      return NextResponse.json(json, { status: upstreamResponse.status });
    } catch {
      return NextResponse.json(
        { success: upstreamResponse.ok, message: responseText },
        { status: upstreamResponse.status },
      );
    }
  } catch (error) {
    const message =
      error instanceof Error && error.name === "AbortError"
        ? "Server ไม่ตอบกลับทันเวลา (60 วินาที)"
        : "ไม่สามารถเชื่อมต่อ Server ได้";
    return NextResponse.json({ success: false, message }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}
