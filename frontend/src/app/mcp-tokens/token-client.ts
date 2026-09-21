import { apiFetch } from "@/lib/client-auth-session";

export type MCPToken = {
  id: string; name: string; kind: "api" | "mcp"; mode: "readonly" | "readwrite";
  holdingCode: string; companyCodes: string[]; companyCode?: string; branchCode?: string;
  createdBy: string; createdAt: string; expiresAt: string;
  revokedAt?: string | null; lastUsedAt?: string | null;
};
export type TokenCompany = { code: string; name: string };
export class TokenRequestError extends Error {
  constructor(readonly status: number, message: string) { super(message); }
}
export async function tokenRequest<T>(path = "", init?: RequestInit): Promise<T> {
  const response = await apiFetch(`/api/mcp-tokens${path}`, { ...init, cache: "no-store", headers: { "Content-Type": "application/json" } });
  const payload = await response.json();
  if (!response.ok || payload.success !== true) throw new TokenRequestError(response.status, typeof payload.message === "string" ? payload.message : "ไม่สามารถจัดการ token ได้ กรุณาลองใหม่");
  return payload.data as T;
}
export function tokenStatus(token: Pick<MCPToken, "revokedAt" | "expiresAt">, now = Date.now()) {
  if (token.revokedAt) return "revoked";
  return Date.parse(token.expiresAt) <= now ? "expired" : "active";
}
