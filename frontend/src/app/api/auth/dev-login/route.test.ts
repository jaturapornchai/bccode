import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
let POST: typeof import("./route").POST;

const DEV_SECRET = "d".repeat(32);

function devRequest(origin = "http://localhost:3000", headers: Record<string, string> = {}) {
  const url = new URL("/api/auth/dev-login", origin);
  return new Request(url, {
    method: "POST",
    headers: { host: url.host, origin: url.origin, ...headers },
  });
}

describe.each(["development", "test", "production"] as const)("Dev Login BFF route (%s)", (environment) => {
  beforeEach(async () => {
    vi.resetModules();
    vi.stubEnv("NODE_ENV", environment);
    process.env.BCAI_DEV_LOGIN_ENABLED = "true";
    process.env.BCAI_DEV_LOGIN_SECRET = DEV_SECRET;
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
    ({ POST } = await import("./route"));
  });

  afterEach(() => {
    delete process.env.BCAI_DEV_LOGIN_ENABLED;
    delete process.env.BCAI_DEV_LOGIN_SECRET;
    delete process.env.BCAI_DEV_LOGIN_BACKEND_URL;
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("forwards only the server secret and keeps refresh in an HttpOnly cookie", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/dev-login");
      expect(init?.method).toBe("POST");
      expect(init?.headers).toEqual({ "X-BC-Dev-Login-Secret": DEV_SECRET });
      expect(init?.body).toBeUndefined();
      return Response.json({ success: true, token: "access-token", refresh: "refresh-token" });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(devRequest());

    expect(response.status).toBe(200);
    expect(await response.json()).toEqual({ success: true, token: "access-token", user: "dev" });
    const cookie = response.headers.get("set-cookie") ?? "";
    expect(cookie).toContain("bc_refresh_token=refresh-token");
    expect(cookie).toContain("HttpOnly");
    expect(cookie.includes("; Secure")).toBe(environment === "production");
    expect(cookie).toContain("SameSite=lax");
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it.each(["http://127.0.0.1:3000", "http://[::1]:3000"])(
    "accepts the exact loopback Host and Origin at %s",
    async (origin) => {
      vi.stubGlobal("fetch", vi.fn(async () => Response.json({ token: "access", refresh: "refresh" })));
      expect((await POST(devRequest(origin))).status).toBe(200);
    },
  );

  it("accepts a 127.0.0.1 browser even when the route's own URL spells localhost", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Response.json({ token: "access", refresh: "refresh" })));
    const request = new Request(new URL("/api/auth/dev-login", "http://127.0.0.1:3000"), {
      method: "POST",
      headers: { host: "127.0.0.1:3000", origin: "http://127.0.0.1:3000" },
    });
    expect((await POST(request)).status).toBe(200);
  });

  it("returns 404 while the server-side feature flag is disabled", async () => {
    process.env.BCAI_DEV_LOGIN_ENABLED = "false";
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(devRequest());
    expect(response.status).toBe(404);
    expect(await response.json()).toMatchObject({ errorcode: "DEV_LOGIN_DISABLED" });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it.each([
    ["a non-loopback URL", devRequest("http://dev.bcaicloud.com")],
    ["a mismatched Host", devRequest("http://localhost:3000", { host: "localhost:3001" })],
    ["a mismatched Origin", devRequest("http://localhost:3000", { origin: "http://127.0.0.1:3000" })],
  ])("rejects %s", async (_case, request) => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    expect((await POST(request)).status).toBe(403);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it.each([
    ["a short secret", "short", "http://localhost:8888"],
    ["a non-loopback backend", DEV_SECRET, "http://192.168.2.202:8888"],
  ])("fails closed for %s", async (_case, secret, backendUrl) => {
    process.env.BCAI_DEV_LOGIN_SECRET = secret;
    process.env.BCAI_LOCAL_BACKEND_URL = backendUrl;
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    expect((await POST(devRequest())).status).toBe(503);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("uses the explicit BCAI_DEV_LOGIN_BACKEND_URL override for container deployments", async () => {
    process.env.BCAI_DEV_LOGIN_BACKEND_URL = "http://mainapi:8888";
    process.env.BCAI_LOCAL_BACKEND_URL = "http://192.168.2.202:8888";
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://mainapi:8888/dev-login");
      expect(init?.headers).toEqual({ "X-BC-Dev-Login-Secret": DEV_SECRET });
      return Response.json({ token: "access", refresh: "refresh" });
    });
    vi.stubGlobal("fetch", fetchMock);

    expect((await POST(devRequest())).status).toBe(200);
  });

  it.each([
    ["a backend path", "http://mainapi:8888/goapi"],
    ["a non-http override", "ftp://mainapi:8888"],
    ["a url with credentials", "http://user:pass@mainapi:8888"],
  ])("fails closed for %s in BCAI_DEV_LOGIN_BACKEND_URL", async (_case, override) => {
    process.env.BCAI_DEV_LOGIN_BACKEND_URL = override;
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    expect((await POST(devRequest())).status).toBe(503);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("does not create a session from an incomplete backend response", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Response.json({ success: true, token: "access-only" })));

    const response = await POST(devRequest());

    expect(response.status).toBe(502);
    expect(response.headers.get("set-cookie")).toBeNull();
  });

  it("does not expose backend account details when authorization fails", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Response.json(
      { errorcode: "DISABLED", message: "user is disabled" },
      { status: 403 },
    )));

    const response = await POST(devRequest());

    expect(response.status).toBe(403);
    expect(await response.json()).toEqual({
      success: false,
      errorcode: "DEV_LOGIN_FORBIDDEN",
      message: "ไม่อนุญาตให้ใช้ Dev Login",
    });
  });
});
