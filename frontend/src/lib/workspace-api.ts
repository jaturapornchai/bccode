import { NextResponse } from "next/server";
import { validateBackendUrl } from "./backend-url";

export type ApiProxyBody = Record<string, unknown> & {
  backendUrl?: unknown;
};

export function requireBearerToken(request: Request): string | NextResponse {
  const authorization = request.headers.get("authorization") ?? "";
  if (!authorization.toLowerCase().startsWith("bearer ") || authorization.slice(7).trim().length === 0) {
    return NextResponse.json({ success: false, message: "ไม่พบ token กรุณาเข้าสู่ระบบใหม่" }, { status: 401 });
  }
  return authorization;
}

export function getBackendUrlFromRequest(request: Request, body?: ApiProxyBody): string {
  const url = new URL(request.url);
  const fromQuery = url.searchParams.get("backendUrl");
  const fromHeader = request.headers.get("x-bc-backend-url");
  const fromBody = typeof body?.backendUrl === "string" ? body.backendUrl : undefined;
  return fromBody ?? fromHeader ?? fromQuery ?? "";
}

export function getMainApiUrl(rawBackendUrl: string): string {
  return validateBackendUrl(rawBackendUrl).mainApiUrl;
}

export async function readJsonOrText(response: Response): Promise<unknown> {
  const contentType = response.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) return response.json();
  return response.text();
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export function extractMessage(payload: unknown): string | undefined {
  if (typeof payload === "string" && payload.trim()) return payload;
  if (!isRecord(payload)) return undefined;
  const error = payload.error;
  if (isRecord(error)) return getString(error, "message") ?? getString(error, "error");
  return getString(payload, "message") ?? getString(payload, "error");
}

export function getString(payload: Record<string, unknown>, key: string): string | undefined {
  const value = payload[key];
  return typeof value === "string" ? value : undefined;
}

export function getArray(payload: Record<string, unknown>, key: string): unknown[] {
  const value = payload[key];
  return Array.isArray(value) ? value : [];
}

export async function proxyMainApiJson(
  request: Request,
  mainApiUrl: string,
  path: string,
  init: RequestInit,
): Promise<NextResponse> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 20000);

  try {
    const response = await fetch(`${mainApiUrl}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
        ...(init.headers ?? {}),
      },
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);
    if (isRecord(payload)) return NextResponse.json(payload, { status: response.status });
    return NextResponse.json({ success: response.ok, message: String(payload ?? "") }, { status: response.status });
  } catch (error) {
    const message =
      error instanceof Error && error.name === "AbortError"
        ? "Server ไม่ตอบกลับทันเวลา"
        : "ไม่สามารถเชื่อมต่อ Server ได้";
    return NextResponse.json({ success: false, message }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}
