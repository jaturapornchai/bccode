import { NextResponse } from "next/server";
import { serverGoApiBase } from "@/lib/backend-url";
import { requireBearerToken } from "@/lib/workspace-api";

// Private image/file stream for the browser. Stored URIs stay "/goapi/s3/file/<key>"; the display
// builders (fileBffUrl in src/lib/image-upload-proxy.ts) turn them into "/api/files/<key>" and fetch
// with the Bearer token (authenticated-image.tsx, logo-avatar.tsx). This replaces the public
// /backend/goapi/s3/file/* rewrite (ADR docs/kms/decisions/2026-09-27-backend-proxy-allowlist.md).
// mainapi (S3FileProxyHandler, backend/internal/goapi/handlers/s3_proxy.go) checks the token and that
// the key belongs to the caller's holding; this route only rebuilds a clean key path and streams back.

type Context = { params: Promise<{ key: string[] }> };

// What mainapi sets on a file response (storageObjectNotModified): caching is private + Vary: Authorization.
const PASSED_RESPONSE_HEADERS = [
  "content-type",
  "content-length",
  "cache-control",
  "etag",
  "last-modified",
  "vary",
  "x-content-type-options",
];
const HEADER_TIMEOUT_MS = 60_000;
const MAX_SEGMENT_LENGTH = 255;
const UNSAFE_SEGMENT = /[\u0000-\u001f\u007f/\\]/;
const CLIENT_IP = /^[0-9A-Fa-f:.]{2,45}$/;

const bad = (message: string, status: number) => NextResponse.json({ success: false, message }, { status });

export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { key } = await context.params;
  // Next hands over decoded segments, so "%2F" arrives as "/" inside one segment: reject it together
  // with empty/dot segments, backslashes (mainapi turns "\" into "/") and control characters.
  if (!key || key.length === 0 || !key.every(isSafeSegment)) return bad("error_occurred", 400);

  const variant = new URL(request.url).searchParams.get("variant");
  if (variant !== null && variant.trim().toLowerCase() !== "thumbnail") return bad("error_occurred", 400);

  let upstreamUrl: string;
  try {
    upstreamUrl = `${serverGoApiBase()}/s3/file/${key.map(encodeURIComponent).join("/")}${variant === null ? "" : "?variant=thumbnail"}`;
  } catch {
    return bad("connection_error", 502);
  }

  const controller = new AbortController();
  const abortUpstream = () => controller.abort();
  request.signal?.addEventListener("abort", abortUpstream);
  const timeout = setTimeout(abortUpstream, HEADER_TIMEOUT_MS);

  let upstream: Response;
  try {
    const ifNoneMatch = request.headers.get("if-none-match");
    const clientIp = forwardedClientIp(request);
    upstream = await fetch(upstreamUrl, {
      method: "GET",
      headers: {
        Authorization: authorization,
        // Images are already compressed; identity keeps Content-Length true for the streamed body.
        "Accept-Encoding": "identity",
        ...(ifNoneMatch ? { "If-None-Match": ifNoneMatch } : {}),
        ...(clientIp ? { "X-Forwarded-For": clientIp } : {}),
      },
      signal: controller.signal,
      cache: "no-store",
      redirect: "manual",
    });
  } catch (error) {
    request.signal?.removeEventListener("abort", abortUpstream);
    const timedOut = error instanceof Error && error.name === "AbortError";
    return bad("connection_error", timedOut ? 504 : 502);
  } finally {
    clearTimeout(timeout);
  }

  const headers = new Headers();
  for (const name of PASSED_RESPONSE_HEADERS) {
    const value = upstream.headers.get(name);
    if (value !== null) headers.set(name, value);
  }
  if (upstream.status === 304) return new Response(null, { status: 304, headers });
  // A redirect (or anything outside 2xx/4xx/5xx) from mainapi is never followed or relayed to the browser.
  if (upstream.status < 200 || (upstream.status >= 300 && upstream.status < 400)) {
    await upstream.body?.cancel();
    return bad("connection_error", 502);
  }
  return new Response(upstream.body, { status: upstream.status, headers });
}

// mainapi's /goapi rate limiter counts per ctx.RealIP() (backend/internal/goapi/bootstrap.go
// createTieredRateLimiter), which reads X-Forwarded-For first; without it every user's images share the
// frontend container's bucket. Caddy (deploy/account/Caddyfile.account, no trusted_proxies) replaces any
// client-sent X-Forwarded-For with the connecting IP and the frontend port is bound to 127.0.0.1
// (deploy/account/compose.yml), so the first entry is the real client. Only that address is passed on.
function forwardedClientIp(request: Request): string | null {
  const first = request.headers.get("x-forwarded-for")?.split(",")[0]?.trim() ?? "";
  return CLIENT_IP.test(first) ? first : null;
}

function isSafeSegment(segment: string): boolean {
  return (
    typeof segment === "string" &&
    segment.length > 0 &&
    segment.length <= MAX_SEGMENT_LENGTH &&
    segment !== "." &&
    segment !== ".." &&
    !UNSAFE_SEGMENT.test(segment)
  );
}
