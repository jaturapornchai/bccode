import { serverMainApiBase } from "@/lib/backend-url";
import { proxyMainApiJson } from "@/lib/workspace-api";

// Holding admin list (server-side proxy to mainapi /holding-member/list) — read-only; adding/removing
// members goes through System Settings › Users (/holding/permission). holdingcode comes from the request
// so it works from the holding-selection screen; the backend resolves the caller's role per-holding.
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
