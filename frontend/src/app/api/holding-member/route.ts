import { serverMainApiBase } from "@/lib/backend-url";
import { proxyMainApiJson } from "@/lib/workspace-api";

// Holding admin management (server-side proxy to mainapi /holding-member/*).
// holdingcode comes from the request so it works from the holding-selection screen;
// the backend resolves the caller's role per-holding and enforces owner/admin.
// Uses serverMainApiBase (local backend) per the server-side hairpin rule and forwards the bearer token.

export async function GET(request: Request): Promise<Response> {
  const holdingcode = new URL(request.url).searchParams.get("holdingcode") ?? "";
  return proxyMainApiJson(
    request,
    serverMainApiBase(),
    `/holding-member/list?holdingcode=${encodeURIComponent(holdingcode)}`,
    { method: "GET" },
  );
}

export async function POST(request: Request): Promise<Response> {
  const body = (await request.json().catch(() => ({}))) as Record<string, unknown>;
  return proxyMainApiJson(request, serverMainApiBase(), "/holding-member/add", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export async function DELETE(request: Request): Promise<Response> {
  const body = (await request.json().catch(() => ({}))) as Record<string, unknown>;
  return proxyMainApiJson(request, serverMainApiBase(), "/holding-member/remove", {
    method: "POST",
    body: JSON.stringify(body),
  });
}
