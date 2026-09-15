import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  isRecord,
  proxyMainApiJson,
  requireBearerToken,
} from "@/lib/workspace-api";

type Context = { params: Promise<{ goPath: string[] }> };

// Allowlist matches full paths because goapi exposes many destructive endpoints that must never be reachable.
const GET_ALLOWED_EXACT = ["api/reports/inventory-valuation"];

const POST_ALLOWED_EXACT = [
  "api/report/tax/vat-register",
  "api/report/tax/pp30-summary",
  "api/report/sales/summary",
  "api/report/sales/by-document",
  "api/process/product-balance",
  "api/stockcost/check",
  "processstockcalccost",
  "api/approval/pr-status/approve",
  "api/approval/pr-status/reject",
  "api/approval/pr-status/pending",
  "api/approval/po-status/pending",
  "api/approval/rfq-status/pending",
  "api/approval/po-status/approve",
  "api/approval/po-status/reject",
  "api/approval/rfq-status/approve",
  "api/approval/rfq-status/reject",
];

const ALLOWED_QUERY_PARAMS = [
  "holdingcode",
  "businesscode",
  "warehousecode",
  "fromdate",
  "todate",
  "limit",
  "offset",
  "lang",
];

const bad = (message: string, status = 400) => NextResponse.json({ success: false, message }, { status });

function sanitizeSegments(segments: string[]): boolean {
  return segments.every((seg) => /^[\p{L}\p{N}_.-]+$/u.test(seg) && seg !== "." && seg !== "..");
}

function toGoApiPath(segments: string[]): string {
  return `/goapi/${segments.map(encodeURIComponent).join("/")}`;
}

function isAllowedGet(segments: string[]): boolean {
  if (GET_ALLOWED_EXACT.includes(segments.join("/"))) return true;
  // api/reports/stock-card/<itemcode>
  if (segments.length === 4 && segments[0] === "api" && segments[1] === "reports" && segments[2] === "stock-card") {
    return true;
  }
  // api/products/<itemcode>/cost-layers
  if (segments.length === 4 && segments[0] === "api" && segments[1] === "products" && segments[3] === "cost-layers") {
    return true;
  }
  return false;
}

export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { goPath } = await context.params;
  if (!goPath || goPath.length === 0) return bad("not_found", 404);
  if (!sanitizeSegments(goPath)) return bad("error_occurred", 400);
  if (!isAllowedGet(goPath)) return bad("not_found", 404);

  try {
    const query = new URLSearchParams();
    new URL(request.url).searchParams.forEach((value, key) => {
      if (ALLOWED_QUERY_PARAMS.includes(key)) query.set(key, value);
    });
    const queryString = query.toString();
    const targetPath = `${toGoApiPath(goPath)}${queryString ? `?${queryString}` : ""}`;

    return await proxyMainApiJson(request, getMainApiUrl(getBackendUrlFromRequest(request)), targetPath, {
      method: "GET",
    });
  } catch {
    return bad("connection_error", 500);
  }
}

export async function POST(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { goPath } = await context.params;
  if (!goPath || goPath.length === 0) return bad("not_found", 404);
  if (!sanitizeSegments(goPath)) return bad("error_occurred", 400);
  if (!POST_ALLOWED_EXACT.includes(goPath.join("/"))) return bad("not_found", 404);

  let parsedBody: unknown;
  try {
    parsedBody = await request.json();
  } catch {
    return bad("error_occurred", 400);
  }
  // isRecord() also returns true for arrays, so reject arrays explicitly.
  if (!isRecord(parsedBody) || Array.isArray(parsedBody)) return bad("error_occurred", 400);

  try {
    return await proxyMainApiJson(
      request,
      getMainApiUrl(getBackendUrlFromRequest(request)),
      toGoApiPath(goPath),
      { method: "POST", body: JSON.stringify(parsedBody) },
    );
  } catch {
    return bad("connection_error", 500);
  }
}
