import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
let POST: typeof import("./route").POST;

describe.each(["development", "test", "production"] as const)("password login route (%s)", (environment) => {
  beforeEach(async () => {
    vi.resetModules();
    vi.stubEnv("NODE_ENV", environment);
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
    ({ POST } = await import("./route"));
  });

  afterEach(() => {
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("allows password login without holdingcode", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/login");
      expect(JSON.parse(String(init?.body))).toEqual({
        username: "demo",
        password: "secret",
        holdingcode: "",
      });
      return Response.json({ success: true, token: "token-0", refresh: "refresh-0" });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        backendUrl: "http://localhost:8888/goapi",
        username: "demo",
        password: "secret",
      }),
    }));
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({
      success: true,
      token: "token-0",
    });
    expect(json).not.toHaveProperty("refresh");
    const cookie = response.headers.get("set-cookie") ?? "";
    expect(cookie).toContain("bc_refresh_token=refresh-0");
    expect(cookie).toContain("Max-Age=43200");
    expect(cookie).toContain("Path=/");
    expect(cookie).toContain("HttpOnly");
    expect(cookie.includes("; Secure")).toBe(environment === "production");
    expect(cookie).toContain("SameSite=lax");
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("forwards normalized holdingcode to mainapi login", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/login");
      expect(init?.method).toBe("POST");
      expect(JSON.parse(String(init?.body))).toEqual({
        username: "demo",
        password: "secret",
        holdingcode: "bcdemo",
      });
      return Response.json({ success: true, token: "token-1", refresh: "refresh-1", mustchangepassword: true });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        backendUrl: "http://localhost:8888/goapi",
        username: "demo",
        password: "secret",
        holdingcode: "BCDemo",
      }),
    }));
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({
      success: true,
      token: "token-1",
      backendUrl: "http://localhost:8888/goapi",
    });
    expect(json).not.toHaveProperty("mustchangepassword");
    expect(json).not.toHaveProperty("refresh");
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("rejects holdingcode with underscore", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        backendUrl: "http://localhost:8888/goapi",
        username: "demo",
        password: "secret",
        holdingcode: "bc_demo",
      }),
    }));
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json).toMatchObject({
      success: false,
      message: "holdingcode ต้องใช้ a-z และ 0-9 เท่านั้น ยาว 3-30 ตัว และขึ้นต้นด้วย a-z ห้ามใช้ _ หรือสัญลักษณ์",
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects a backend login response without a refresh token", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Response.json({ success: true, token: "access-only" })));

    const response = await POST(new Request("http://localhost/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        backendUrl: "http://localhost:8888/goapi",
        username: "demo",
        password: "secret",
      }),
    }));

    expect(response.status).toBe(502);
    expect(await response.json()).toEqual({ success: false, message: "Server ตอบกลับไม่ครบ" });
    expect(response.headers.get("set-cookie")).toBeNull();
  });
});
