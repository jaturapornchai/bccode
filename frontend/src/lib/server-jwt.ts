import { createHmac, timingSafeEqual } from "crypto";

export type JwtVerificationResult =
  | { ok: true; claims: Record<string, unknown> }
  | { ok: false; status: 401 | 500; message: string };

export function verifyHs256Jwt(authorization: string): JwtVerificationResult {
  const secret = process.env.JWT_SECRET_KEY?.trim();
  if (!secret) return { ok: false, status: 500, message: "JWT_SECRET_KEY is not configured" };

  const token = authorization.slice(7).trim();
  const parts = token.split(".");
  if (parts.length !== 3) return { ok: false, status: 401, message: "token ไม่ถูกต้อง" };

  try {
    const [encodedHeader, encodedPayload, encodedSignature] = parts;
    const header = JSON.parse(base64UrlToBuffer(encodedHeader).toString("utf8")) as Record<string, unknown>;
    if (header.alg !== "HS256") return { ok: false, status: 401, message: "token ไม่ถูกต้อง" };

    const expectedSignature = createHmac("sha256", secret).update(`${encodedHeader}.${encodedPayload}`).digest();
    const actualSignature = base64UrlToBuffer(encodedSignature);
    if (actualSignature.length !== expectedSignature.length || !timingSafeEqual(actualSignature, expectedSignature)) {
      return { ok: false, status: 401, message: "token ไม่ถูกต้อง" };
    }

    const payload = JSON.parse(base64UrlToBuffer(encodedPayload).toString("utf8")) as Record<string, unknown>;
    if (typeof payload.exp === "number" && payload.exp <= Math.floor(Date.now() / 1000)) {
      return { ok: false, status: 401, message: "token หมดอายุ" };
    }
    return { ok: true, claims: payload };
  } catch {
    return { ok: false, status: 401, message: "token ไม่ถูกต้อง" };
  }
}

export function getJwtClaimShopId(claims: Record<string, unknown>): string {
  const value = claims.shopid ?? claims.shop_id ?? claims.shopId;
  return typeof value === "string" ? value.trim() : "";
}

function base64UrlToBuffer(value: string): Buffer {
  const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
  return Buffer.from(padded, "base64");
}
