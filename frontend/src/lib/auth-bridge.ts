export const DEFAULT_AUTH_BRIDGE_URL = "https://dev-api.bcaicloud.com/liff";

type AuthBridgeEnv = {
  [key: string]: string | undefined;
  BC_AUTH_BRIDGE_URL?: string;
};

export function getAuthBridgeUrl(env: AuthBridgeEnv = process.env): string {
  return normalizeAuthBridgeUrl(env.BC_AUTH_BRIDGE_URL ?? DEFAULT_AUTH_BRIDGE_URL);
}

export function normalizeAuthBridgeUrl(rawUrl: string): string {
  const value = rawUrl.trim();
  if (!value) throw new Error("Auth bridge URL ไม่ถูกต้อง");

  let parsed: URL;
  try {
    parsed = new URL(value);
  } catch {
    throw new Error("Auth bridge URL ไม่ถูกต้อง");
  }

  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    throw new Error("Auth bridge URL ต้องเป็น http หรือ https เท่านั้น");
  }

  if (parsed.username || parsed.password) {
    throw new Error("Auth bridge URL ห้ามมี username หรือ password");
  }

  parsed.hash = "";
  parsed.search = "";
  parsed.pathname = parsed.pathname.replace(/\/+$/, "");
  return parsed.toString().replace(/\/$/, "");
}

export async function readJsonOrText(response: Response): Promise<unknown> {
  const contentType = response.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) {
    return response.json();
  }
  return response.text();
}

export function extractMessage(payload: unknown): string | undefined {
  if (typeof payload === "string" && payload.trim()) return payload;
  if (!isRecord(payload)) return undefined;

  const error = payload.error;
  if (isRecord(error)) {
    return getString(error, "message") ?? getString(error, "error");
  }

  return getString(payload, "message") ?? getString(payload, "error");
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export function getRecord(payload: Record<string, unknown>, key: string): Record<string, unknown> | undefined {
  const value = payload[key];
  return isRecord(value) ? value : undefined;
}

export function getString(payload: Record<string, unknown>, key: string): string | undefined {
  const value = payload[key];
  return typeof value === "string" ? value : undefined;
}

export function getNumber(payload: Record<string, unknown>, key: string): number | undefined {
  const value = payload[key];
  return typeof value === "number" && Number.isFinite(value) ? value : undefined;
}

export async function postMainApiAuth(mainApiUrl: string, path: string, payload: Record<string, unknown>) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 15000);

  try {
    const response = await fetch(`${mainApiUrl}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      signal: controller.signal,
      cache: "no-store",
    });

    const data = await readJsonOrText(response);
    if (!response.ok) {
      throw new Error(extractMessage(data) ?? `Server ตอบกลับผิดปกติ (${response.status})`);
    }

    if (!isRecord(data)) {
      throw new Error("Server ตอบกลับไม่ถูกต้อง");
    }

    const token = getString(data, "token") ?? getString(getRecord(data, "data") ?? {}, "token");
    const refresh = getString(data, "refresh") ?? getString(getRecord(data, "data") ?? {}, "refresh");
    const username = getString(data, "username") ?? getString(getRecord(data, "data") ?? {}, "username");

    if (!token) {
      throw new Error(extractMessage(data) ?? "เข้าสู่ระบบไม่สำเร็จ");
    }

    return { token, refresh: refresh ?? "", username: username ?? "" };
  } catch (error) {
    if (error instanceof Error && error.name === "AbortError") {
      throw new Error("Server ไม่ตอบกลับทันเวลา");
    }
    throw error;
  } finally {
    clearTimeout(timeout);
  }
}
