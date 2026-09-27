import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  isRecord,
  proxyMainApiJson,
  requireBearerToken,
  type ApiProxyBody,
} from "@/lib/workspace-api";

// Company/branch management for the settings screens (company-branch-tree-view.tsx, currency-screen.tsx).
// Replaces the public /backend/organization/* rewrite that next.config.ts no longer has
// (ADR docs/kms/decisions/2026-09-27-backend-proxy-allowlist.md). Exact allowlist:
//   GET  company | branch        -> mainapi GET /organization/<resource>?management=true (whole structure)
//   POST company                 -> mainapi POST /organization/company
//   PUT  company/<id> | branch/<id> -> mainapi PUT /organization/<resource>/<id>
// Creating a branch uses the existing POST /api/workspace/branch. mainapi checks the token and the
// holding-admin rights itself; this route only fixes which mainapi paths the browser can reach.

type Context = { params: Promise<{ orgPath: string[] }> };
type Resource = "company" | "branch";

const RESOURCES = new Set<string>(["company", "branch"]);
// The id is the company/branch code (mainapi setBranchIdentity: guidfixed = code). Codes are free text
// (backend NormalizeCompanyCode only trims + uppercases), so every character a code can hold is allowed
// except what could change the mainapi path: "/", "\", control characters, "." and "..".
const ID_PATTERN = /^[^/\\\p{Cc}]{1,128}$/u;

const bad = (message: string, status: number) => NextResponse.json({ success: false, message }, { status });

export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { orgPath } = await context.params;
  const resource = listResource(orgPath);
  if (!resource) return bad("not_found", 404);

  const mainApiUrl = resolveMainApiUrl(request);
  if (mainApiUrl instanceof NextResponse) return mainApiUrl;
  return proxyMainApiJson(request, mainApiUrl, `/organization/${resource}?management=true`, { method: "GET" });
}

export async function POST(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { orgPath } = await context.params;
  if (listResource(orgPath) !== "company") return bad("not_found", 404);

  const body = await readObjectBody(request);
  if (!body) return bad("error_occurred", 400);

  const mainApiUrl = resolveMainApiUrl(request, body);
  if (mainApiUrl instanceof NextResponse) return mainApiUrl;
  return proxyMainApiJson(request, mainApiUrl, "/organization/company", {
    method: "POST",
    body: JSON.stringify(withoutBackendUrl(body)),
  });
}

export async function PUT(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { orgPath } = await context.params;
  if (!orgPath || orgPath.length !== 2 || !RESOURCES.has(orgPath[0])) return bad("not_found", 404);
  const [resource, id] = orgPath;
  if (!ID_PATTERN.test(id) || id === "." || id === "..") return bad("error_occurred", 400);

  const body = await readObjectBody(request);
  if (!body) return bad("error_occurred", 400);

  const mainApiUrl = resolveMainApiUrl(request, body);
  if (mainApiUrl instanceof NextResponse) return mainApiUrl;
  return proxyMainApiJson(request, mainApiUrl, `/organization/${resource}/${goPathEscape(id)}`, {
    method: "PUT",
    body: JSON.stringify(withoutBackendUrl(body)),
  });
}

// mainapi (Echo v4) routes on r.URL.RawPath when Go sets it and never unescapes :id (echo.GetPath,
// backend/pkg/microservice/context_http.go Param). Go leaves RawPath empty only when the request path is
// byte-for-byte its own canonical escaping (net/url setPath + shouldEscape, encodePath mode); :id is then
// the decoded code. This reproduces that escaping so codes such as A&B, K+1 or Thai text reach mainapi
// exactly as stored (encodeURIComponent escapes & + , ; = @ : $, which Go keeps, and keeps ! ' ( ) *).
function goPathEscape(segment: string): string {
  let escaped = "";
  for (const byte of new TextEncoder().encode(segment)) {
    const char = String.fromCharCode(byte);
    escaped += /[A-Za-z0-9\-_.~$&+,:;=@]/.test(char) ? char : `%${byte.toString(16).toUpperCase().padStart(2, "0")}`;
  }
  return escaped;
}

function listResource(orgPath: string[] | undefined): Resource | null {
  if (!orgPath || orgPath.length !== 1 || !RESOURCES.has(orgPath[0])) return null;
  return orgPath[0] as Resource;
}

// Validates an x-bc-backend-url / body.backendUrl identifier if one was sent; the fetch always goes to
// serverMainApiBase() (getMainApiUrl), never to a host the client names.
function resolveMainApiUrl(request: Request, body?: ApiProxyBody): string | NextResponse {
  try {
    return getMainApiUrl(getBackendUrlFromRequest(request, body));
  } catch {
    return bad("error_occurred", 400);
  }
}

async function readObjectBody(request: Request): Promise<ApiProxyBody | null> {
  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return null;
  }
  // isRecord() is also true for arrays.
  return isRecord(body) && !Array.isArray(body) ? (body as ApiProxyBody) : null;
}

function withoutBackendUrl(body: ApiProxyBody): Record<string, unknown> {
  const { backendUrl: _backendUrl, ...payload } = body;
  void _backendUrl;
  return payload;
}
