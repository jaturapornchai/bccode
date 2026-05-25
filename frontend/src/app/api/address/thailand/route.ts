import { NextResponse } from "next/server";

import { validateBackendUrl } from "@/lib/backend-url";
import type { ThailandAddressData } from "@/lib/thailand-addresses";

const addressCacheSeconds = 60 * 60 * 24;

export async function GET(request: Request) {
  const url = new URL(request.url);

  let normalizedGoApiUrl: string;
  try {
    normalizedGoApiUrl = validateBackendUrl(url.searchParams.get("backendUrl") ?? "").normalizedGoApiUrl;
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 8000);

  try {
    const response = await fetch(`${normalizedGoApiUrl}/api/address/thailand`, {
      cache: "force-cache",
      next: { revalidate: addressCacheSeconds },
      signal: controller.signal,
    });
    const data = (await response.json().catch(() => null)) as unknown;
    if (!response.ok || !isThailandAddressData(data)) {
      return NextResponse.json({ success: false, message: "โหลดข้อมูลที่อยู่ไทยไม่สำเร็จ" }, { status: 502 });
    }
    return NextResponse.json(data, {
      status: 200,
      headers: { "Cache-Control": `public, max-age=${addressCacheSeconds}` },
    });
  } catch {
    return NextResponse.json({ success: false, message: "เชื่อมต่อ backend address API ไม่สำเร็จ" }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}

function isThailandAddressData(payload: unknown): payload is ThailandAddressData {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) return false;
  const data = payload as Partial<ThailandAddressData>;
  return (
    data.version === 1 &&
    Array.isArray(data.countries) &&
    data.countries.some((country) => country?.code === "TH" && Array.isArray(country.provinces))
  );
}
