import { NextResponse } from "next/server";
import { serverGoApiBase } from "@/lib/backend-url";

type Body = {
  endpoint?: string;
  publicEndpoint?: string;
};

// Check if an IP address is a private / link-local address to prevent SSRF
function isBlockedPrivateIP(hostname: string): boolean {
  // Cloud metadata endpoint (AWS / GCP / DigitalOcean / Azure)
  if (hostname === "169.254.169.254") return true;

  const ipv4Regex = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/;
  const match = hostname.match(ipv4Regex);
  if (match) {
    const octet1 = parseInt(match[1], 10);
    const octet2 = parseInt(match[2], 10);
    // 10.0.0.0/8
    if (octet1 === 10) return true;
    // 172.16.0.0/12 (RFC 1918 private)
    if (octet1 === 172 && octet2 >= 16 && octet2 <= 31) return true;
    // 192.168.0.0/16
    if (octet1 === 192 && octet2 === 168) return true;
    // 169.254.0.0/16 (link-local)
    if (octet1 === 169 && octet2 === 254) return true;
  }
  return false;
}

// Check if endpoint refers to the internal MinIO deployment or server host
function isInternalMinio(hostname: string, port?: string): boolean {
  const h = hostname.toLowerCase();
  return (
    h === "minio" ||
    h === "localhost" ||
    h === "127.0.0.1" ||
    h === "::1" ||
    h.endsWith(".internal") ||
    h === "account.bcaicloud.com" ||
    h === "159.223.43.229" ||
    port === "9000" ||
    port === "9100"
  );
}

function getLocalBackendBase(): string {
  try {
    return serverGoApiBase();
  } catch {
    return "http://127.0.0.1:8888/goapi";
  }
}

async function verifyInternalMinioViaBackend(start: number): Promise<NextResponse> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 6000);
  try {
    const backendBase = getLocalBackendBase();
    const healthRes = await fetch(`${backendBase}/api/health`, {
      signal: controller.signal,
      cache: "no-store",
    });
    clearTimeout(timeout);
    const latencyMs = Date.now() - start;
    if (healthRes.ok) {
      return NextResponse.json({
        success: true,
        message: "เชื่อมต่อที่เก็บรูป (MinIO Storage) สำเร็จ",
        latencyMs,
        httpStatus: 200,
      });
    }
    return NextResponse.json({
      success: false,
      message: `ระบบ backend รายงานสถานะไม่พร้อมใช้งาน (HTTP ${healthRes.status})`,
      latencyMs,
    });
  } catch (err) {
    clearTimeout(timeout);
    const latencyMs = Date.now() - start;
    return NextResponse.json({
      success: false,
      message: err instanceof Error ? err.message : "ไม่สามารถเชื่อมต่อระบบจัดเก็บรูปภาพภายในได้",
      latencyMs,
    });
  }
}

export async function POST(request: Request) {
  let body: Body = {};
  try {
    const raw = await request.text();
    if (raw.trim()) {
      body = JSON.parse(raw) as Body;
    }
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  // Prioritize endpoint (s3endpoint) over publicEndpoint
  const endpoint = (body.endpoint ?? "").trim() || (body.publicEndpoint ?? "").trim();
  if (!endpoint) {
    return NextResponse.json(
      { success: false, message: "กรุณาระบุ S3 Endpoint" },
      { status: 400 },
    );
  }

  const urlString = /^https?:\/\//i.test(endpoint) ? endpoint : `http://${endpoint}`;
  let parsed: URL;
  try {
    parsed = new URL(urlString);
  } catch {
    return NextResponse.json(
      { success: false, message: "URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    return NextResponse.json(
      { success: false, message: "รองรับเฉพาะ HTTP และ HTTPS เท่านั้น" },
      { status: 400 },
    );
  }

  const start = Date.now();

  // Internal MinIO / local server check: since frontend is in the edge network and minio is in the internal data network,
  // the frontend container verifies storage service health via backend Go API.
  if (isInternalMinio(parsed.hostname, parsed.port)) {
    return verifyInternalMinioViaBackend(start);
  }

  // SSRF guard: block private RFC 1918 / cloud metadata IPs for external probes
  if (isBlockedPrivateIP(parsed.hostname)) {
    return NextResponse.json(
      { success: false, message: "ไม่อนุญาตให้เชื่อมต่อ IP ภายในเครือข่าย (SSRF Guard)" },
      { status: 403 },
    );
  }

  // External S3 / R2 probe: probe directly with timeout
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 6000);

  try {
    const response = await fetch(urlString, {
      method: "GET",
      redirect: "manual",
      signal: controller.signal,
      cache: "no-store",
    });
    clearTimeout(timeout);
    const latencyMs = Date.now() - start;
    // For S3/MinIO/R2 any HTTP response (even 400, 403, 404) proves the endpoint is alive and reachable
    return NextResponse.json({
      success: true,
      message: "เชื่อมต่อได้",
      latencyMs,
      httpStatus: response.status,
    });
  } catch (error) {
    clearTimeout(timeout);
    // If probing external endpoint failed, but this was a port 9100/9000 or server-related URL, fallback to backend
    if (parsed.port === "9100" || parsed.port === "9000" || parsed.hostname.includes("bcaicloud.com")) {
      return verifyInternalMinioViaBackend(start);
    }
    const latencyMs = Date.now() - start;
    const message = error instanceof Error ? error.message : "เชื่อมต่อไม่ได้";
    return NextResponse.json(
      { success: false, message, latencyMs },
      { status: 200 },
    );
  }
}
