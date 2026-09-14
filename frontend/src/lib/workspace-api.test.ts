import { afterEach, describe, expect, it, vi } from "vitest";
import { proxyMainApiJson } from "./workspace-api";

const request = new Request("http://localhost/api/gl/command", { method: "POST", headers: { Authorization: "Bearer test-token", "x-bc-backend-url": "http://localhost:8888" } });
const upstream = (body: unknown, status: number) => vi.stubGlobal("fetch", vi.fn().mockResolvedValue(Response.json(body, { status })));

afterEach(() => vi.unstubAllGlobals());

describe("proxyMainApiJson relays expected user errors without a console error", () => {
  it("turns 4xx + code into HTTP 200 with success:false and keeps the Thai message", async () => {
    upstream({ success: false, code: "duplicate_code", message: "รหัสบัญชีนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น" }, 409);
    const response = await proxyMainApiJson(request, "http://localhost:8888", "/gl/v2/command", { method: "POST" }, { userErrorStatusOk: true });
    expect(response.status).toBe(200);
    await expect(response.json()).resolves.toMatchObject({ success: false, code: "duplicate_code", message: "รหัสบัญชีนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น" });
  });

  it("keeps the upstream status when the caller did not opt in", async () => {
    upstream({ success: false, code: "duplicate_code", message: "รหัสบัญชีนี้ถูกใช้แล้ว" }, 409);
    const response = await proxyMainApiJson(request, "http://localhost:8888", "/gl/v2/command", { method: "POST" });
    expect(response.status).toBe(409);
  });

  it("never hides auth failures (401/403) or server errors (5xx)", async () => {
    for (const status of [401, 403, 500, 503]) {
      upstream({ success: false, code: "unauthorized", message: "ไม่ได้รับสิทธิ์" }, status);
      const response = await proxyMainApiJson(request, "http://localhost:8888", "/gl/v2/command", { method: "POST" }, { userErrorStatusOk: true });
      expect(response.status).toBe(status);
    }
  });

  it("keeps 4xx without a machine code as-is", async () => {
    upstream({ success: false, message: "bad request" }, 400);
    const response = await proxyMainApiJson(request, "http://localhost:8888", "/gl/v2/command", { method: "POST" }, { userErrorStatusOk: true });
    expect(response.status).toBe(400);
  });

  it("requires a bearer token before calling the API", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const response = await proxyMainApiJson(new Request("http://localhost/api/gl/command"), "http://localhost:8888", "/gl/v2/command", { method: "POST" }, { userErrorStatusOk: true });
    expect(response.status).toBe(401);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
