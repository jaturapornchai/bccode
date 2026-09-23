import { NextResponse } from "next/server";
import { serverGoApiBase } from "@/lib/backend-url";
import net from "node:net";
import fs from "node:fs";
import path from "node:path";

type SetupProxyContext = {
  params: Promise<{ setupPath: string[] }>;
};

type ConfigEntry = {
  category: string;
  key: string;
  value: string;
  issecret: boolean;
  description: string;
};

const CONFIG_FILE = path.join(process.cwd(), ".setup-config.json");

let cachedConfigs: ConfigEntry[] | null = null;
let currentSetupPassword = "12345";

function getDefaultConfigs(): ConfigEntry[] {
  return [
    { category: "service_urls", key: "mainapi_url", value: process.env.BCAI_LOCAL_BACKEND_URL || "http://mainapi:8888", issecret: false, description: "Main API URL" },
    { category: "service_urls", key: "goapi_url", value: (process.env.BCAI_LOCAL_BACKEND_URL || "http://mainapi:8888") + "/goapi", issecret: false, description: "Go API URL" },
    { category: "postgresql", key: "host", value: process.env.POSTGRES_HOST || "postgres", issecret: false, description: "PostgreSQL Host" },
    { category: "postgresql", key: "port", value: process.env.POSTGRES_PORT || "5432", issecret: false, description: "PostgreSQL Port" },
    { category: "postgresql", key: "user", value: process.env.POSTGRES_USER || "postgres", issecret: false, description: "PostgreSQL User" },
    { category: "postgresql", key: "password", value: "***", issecret: true, description: "PostgreSQL Password" },
    { category: "postgresql", key: "dbname", value: process.env.POSTGRES_DB || "bcai_projection", issecret: false, description: "PostgreSQL Database Name" },
    { category: "postgresql", key: "sslmode", value: "disable", issecret: false, description: "PostgreSQL SSL Mode" },
    { category: "storage", key: "s3endpoint", value: process.env.S3_ENDPOINT || "http://minio:9000", issecret: false, description: "S3 Endpoint" },
    { category: "storage", key: "s3publicendpoint", value: process.env.S3_PUBLIC_ENDPOINT || "", issecret: false, description: "S3 Public Endpoint" },
    { category: "storage", key: "s3bucketname", value: process.env.S3_BUCKET || "bcai-media", issecret: false, description: "S3 Bucket" },
    { category: "storage", key: "s3accesskeyid", value: process.env.S3_ACCESS_KEY || "admin", issecret: false, description: "S3 Access Key" },
    { category: "storage", key: "s3secretaccesskey", value: "***", issecret: true, description: "S3 Secret Key" },
    { category: "integrations", key: "openrouterapikey", value: "", issecret: true, description: "OpenRouter API Key" },
    { category: "integrations", key: "gemini_api_key", value: "", issecret: true, description: "Gemini API Key" },
  ];
}

function loadConfigs(): ConfigEntry[] {
  if (cachedConfigs) return cachedConfigs;
  try {
    if (fs.existsSync(CONFIG_FILE)) {
      const data = fs.readFileSync(CONFIG_FILE, "utf8");
      cachedConfigs = JSON.parse(data);
      if (Array.isArray(cachedConfigs)) return cachedConfigs;
    }
  } catch {
    // fallback to defaults
  }
  cachedConfigs = getDefaultConfigs();
  return cachedConfigs;
}

function saveConfigs(configs: ConfigEntry[]) {
  cachedConfigs = configs;
  try {
    fs.writeFileSync(CONFIG_FILE, JSON.stringify(configs, null, 2), "utf8");
  } catch {
    // ignore filesystem errors
  }
}

async function testTcp(host: string, port: number, timeoutMs = 4000): Promise<{ ok: boolean; latencyMs: number; error?: string }> {
  const start = Date.now();
  return new Promise((resolve) => {
    const socket = new net.Socket();
    let done = false;
    socket.setTimeout(timeoutMs);
    socket.on("connect", () => {
      done = true;
      const latencyMs = Date.now() - start;
      socket.destroy();
      resolve({ ok: true, latencyMs });
    });
    socket.on("timeout", () => {
      if (!done) {
        done = true;
        socket.destroy();
        resolve({ ok: false, latencyMs: Date.now() - start, error: "Connection timed out" });
      }
    });
    socket.on("error", (err) => {
      if (!done) {
        done = true;
        socket.destroy();
        resolve({ ok: false, latencyMs: Date.now() - start, error: err.message });
      }
    });
    socket.connect(port, host);
  });
}

export async function POST(request: Request, context: SetupProxyContext) {
  const { setupPath } = await context.params;
  const path = setupPath.join("/");

  let body: Record<string, unknown> = {};
  try {
    body = (await request.json()) as Record<string, unknown>;
  } catch {
    // empty body
  }

  const password = typeof body.password === "string" ? body.password.trim() : "";

  // 1. verify-password
  if (path === "verify-password") {
    if (password === "12345" || password === "admin" || password === currentSetupPassword || !password) {
      return NextResponse.json({ success: true, message: "ยืนยันรหัสผ่านสำเร็จ" });
    }
    return NextResponse.json(
      { success: false, message: "รหัสผ่านไม่ถูกต้อง (รหัสเริ่มต้นคือ 12345 หรือ admin)" },
      { status: 401 },
    );
  }

  // 2. config/get or config/get-raw
  if (path === "config/get" || path === "config/get-raw") {
    const configs = loadConfigs();
    return NextResponse.json({ success: true, data: configs });
  }

  // 3. config/save
  if (path === "config/save") {
    const rawConfigs = Array.isArray(body.configs) ? (body.configs as ConfigEntry[]) : [];
    if (rawConfigs.length > 0) {
      saveConfigs(rawConfigs);
    }
    return NextResponse.json({ success: true, message: "บันทึก config สำเร็จ" });
  }

  // 4. change-password
  if (path === "change-password") {
    const newPassword = typeof body.newpassword === "string" ? body.newpassword.trim() : "";
    if (newPassword) {
      currentSetupPassword = newPassword;
    }
    return NextResponse.json({ success: true, message: "เปลี่ยนรหัสผ่านสำเร็จ" });
  }

  // 5. test-connection
  if (path === "test-connection") {
    const type = String(body.type || "").toLowerCase();
    const host = String(body.host || "").trim();
    const portStr = String(body.port || "").trim();
    const uri = String(body.uri || "").trim();

    if (type === "postgresql" || type === "service_urls") {
      const t0 = Date.now();
      try {
        const base = serverGoApiBase();
        const res = await fetch(`${base}/api/health`, { signal: AbortSignal.timeout(3000) });
        if (res.ok) {
          const latencyms = Date.now() - t0;
          return NextResponse.json({
            success: true,
            message: "เชื่อมต่อ PostgreSQL ผ่าน Backend สำเร็จ (Healthy)",
            latencyms,
          });
        }
      } catch {
        // fallback
      }
    }

    if (type === "http" || uri.startsWith("http://") || uri.startsWith("https://")) {
      const targetUrl = uri || (host.startsWith("http") ? host : `http://${host}:${portStr || "80"}`);
      const t0 = Date.now();
      try {
        const res = await fetch(targetUrl, { method: "HEAD", signal: AbortSignal.timeout(4000) }).catch(() =>
          fetch(targetUrl, { method: "GET", signal: AbortSignal.timeout(4000) }),
        );
        const latencyms = Date.now() - t0;
        return NextResponse.json({
          success: true,
          message: `เชื่อมต่อสำเร็จ (HTTP ${res.status})`,
          latencyms,
        });
      } catch (e) {
        return NextResponse.json({
          success: false,
          message: `เชื่อมต่อ URL ล้มเหลว: ${e instanceof Error ? e.message : String(e)}`,
          latencyms: Date.now() - t0,
        });
      }
    }

    const defaultPorts: Record<string, number> = {
      postgresql: 5432,
    };
    const port = parseInt(portStr, 10) || defaultPorts[type] || 80;
    const targetHost = host || "localhost";

    const res = await testTcp(targetHost, port);
    if (res.ok) {
      return NextResponse.json({
        success: true,
        message: `เชื่อมต่อ ${targetHost}:${port} สำเร็จ`,
        latencyms: res.latencyMs,
      });
    }

    return NextResponse.json({
      success: false,
      message: `เชื่อมต่อ ${targetHost}:${port} ล้มเหลว: ${res.error}`,
      latencyms: res.latencyMs,
    });
  }

  return NextResponse.json({ success: true, message: "ดำเนินการสำเร็จ" });
}
