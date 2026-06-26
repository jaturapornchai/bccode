import { serverMainApiBase } from "@/lib/backend-url";
import { proxyMainApiJson } from "@/lib/workspace-api";

// Bulk user import proxy (server-side) to mainapi /holding/users/import.
// Body: {holdingcode, filename, contentbase64, commit}. The backend parses the .csv/.xlsx,
// validates each row, and either previews (commit=false) or writes (commit=true) via the
// per-holding-authorized SaveUserFullProfile path. Forwards the bearer token.
export async function POST(request: Request): Promise<Response> {
  const body = (await request.json().catch(() => ({}))) as Record<string, unknown>;
  return proxyMainApiJson(request, serverMainApiBase(), "/holding/users/import", {
    method: "POST",
    body: JSON.stringify(body),
  });
}
