import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { POST } from "./route";

describe("storage health route", () => {
  beforeEach(() => {
    vi.stubEnv("BCAI_LOCAL_BACKEND_URL", "http://127.0.0.1:8888");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("returns 400 when endpoint is empty", async () => {
    const request = new Request("http://localhost/api/storage/health", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ endpoint: "" }),
    });

    const response = await POST(request);
    expect(response.status).toBe(400);
    const data = await response.json();
    expect(data.success).toBe(false);
    expect(data.message).toBe("กรุณาระบุ S3 Endpoint");
  });

  it("returns 403 when endpoint is blocked private IP (SSRF guard)", async () => {
    const request = new Request("http://localhost/api/storage/health", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ endpoint: "http://169.254.169.254/latest/meta-data" }),
    });

    const response = await POST(request);
    expect(response.status).toBe(403);
    const data = await response.json();
    expect(data.success).toBe(false);
    expect(data.message).toContain("SSRF Guard");
  });

  it("verifies internal minio endpoint via backend health check", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ status: "healthy" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const request = new Request("http://localhost/api/storage/health", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ endpoint: "http://minio:9000" }),
    });

    const response = await POST(request);
    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data.success).toBe(true);
    expect(data.message).toContain("MinIO Storage");
    expect(data.httpStatus).toBe(200);
  });

  it("verifies server host port 9100 via backend health check", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ status: "healthy" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const request = new Request("http://localhost/api/storage/health", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ endpoint: "http://account.bcaicloud.com:9100" }),
    });

    const response = await POST(request);
    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data.success).toBe(true);
    expect(data.message).toContain("MinIO Storage");
    expect(data.httpStatus).toBe(200);
  });

  it("handles internal minio endpoint when backend is down", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValueOnce(new Error("Connection refused"));

    const request = new Request("http://localhost/api/storage/health", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ endpoint: "http://minio:9000" }),
    });

    const response = await POST(request);
    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data.success).toBe(false);
    expect(data.message).toContain("Connection refused");
  });

  it("probes external S3 endpoint directly", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(null, { status: 403, statusText: "Forbidden" }),
    );

    const request = new Request("http://localhost/api/storage/health", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ endpoint: "https://mybucket.r2.cloudflarestorage.com" }),
    });

    const response = await POST(request);
    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data.success).toBe(true);
    expect(data.httpStatus).toBe(403);
  });
});
