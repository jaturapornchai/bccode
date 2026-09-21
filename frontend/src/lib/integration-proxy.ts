import { serverMainApiBase } from "@/lib/backend-url";

// API and MCP use separate bearer credentials. Never forward browser session cookies or
// accept a caller-selected backend URL on this public machine-to-machine route.
export async function proxyIntegration(request: Request, kind: "api" | "mcp", path: string): Promise<Response> {
  const authorization = request.headers.get("authorization") ?? "";
  if (!new RegExp(`^Bearer bcai${kind}_[A-Za-z0-9_.-]+$`, "i").test(authorization) || authorization.length > 210) {
    return Response.json({ error: "invalid_token" }, { status: 401, headers: { "WWW-Authenticate": `Bearer realm="BC GL ${kind.toUpperCase()}"`, "Cache-Control": "no-store" } });
  }
  const headers = new Headers({ authorization });
  if (kind === "api") {
    for (const key of ["x-bc-company-code", "x-bc-company-codes"]) {
      const company = request.headers.get(key);
      if (company) headers.set(key, company);
    }
  }
  for (const key of ["content-type", "accept", "origin", "mcp-protocol-version"]) {
    const value = request.headers.get(key);
    if (value) headers.set(key, value);
  }
  let body: Uint8Array | undefined;
  if (request.method === "POST") {
    const reader = request.body?.getReader();
    const chunks: Uint8Array[] = [];
    let size = 0;
    if (reader) {
      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;
        size += value.byteLength;
        if (size > 2 * 1024 * 1024) {
          await reader.cancel();
          return Response.json({ error: "request_too_large" }, { status: 413 });
        }
        chunks.push(value);
      }
    }
    body = new Uint8Array(size);
    let offset = 0;
    for (const chunk of chunks) { body.set(chunk, offset); offset += chunk.length; }
  }
  try {
    const result = await fetch(`${serverMainApiBase()}${path}`, {
      method: request.method, headers, body: body as BodyInit | undefined,
      cache: "no-store", redirect: "error", signal: AbortSignal.timeout(40_000),
    });
    const responseHeaders = new Headers({ "Cache-Control": "no-store", "X-Content-Type-Options": "nosniff" });
    for (const key of ["content-type", "www-authenticate", "allow"]) {
      const value = result.headers.get(key);
      if (value) responseHeaders.set(key, value);
    }
    return new Response(result.body, { status: result.status, headers: responseHeaders });
  } catch {
    return Response.json({ error: "integration_unavailable" }, { status: 503, headers: { "Cache-Control": "no-store" } });
  }
}

