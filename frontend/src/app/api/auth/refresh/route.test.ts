import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
let POST: typeof import("./route").POST;

describe.each(["development", "test", "production"] as const)("refresh route (%s)", (environment) => {
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

  it("rotates the HttpOnly cookie and returns only an access token", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/refresh");
      expect(JSON.parse(String(init?.body))).toEqual({ token: "refresh-old" });
      return Response.json({ success: true, token: "access-new", refresh: "refresh-new" });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/refresh", {
      method: "POST",
      headers: { Cookie: "bc_refresh_token=refresh-old" },
    }));
    const json = await response.json();

    expect(json).toEqual({ success: true, token: "access-new" });
    const cookie = response.headers.get("set-cookie") ?? "";
    expect(cookie).toContain("bc_refresh_token=refresh-new");
    expect(cookie).toContain("Max-Age=43200");
    expect(cookie).toContain("Path=/");
    expect(cookie).toContain("HttpOnly");
    expect(cookie.includes("; Secure")).toBe(environment === "production");
    expect(cookie).toContain("SameSite=lax");
  });

  it("rejects a request without a refresh cookie", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/refresh", { method: "POST" }));

    expect(response.status).toBe(401);
    expect(response.headers.get("set-cookie")).toMatch(/bc_refresh_token=;.*Max-Age=0/i);
    expect(response.headers.get("set-cookie")?.includes("; Secure")).toBe(environment === "production");
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("preserves the refresh cookie on a transient backend failure", async () => {
    vi.stubGlobal("fetch", vi.fn(async () =>
      Response.json({ success: false, message: "temporary" }, { status: 503 }),
    ));

    const response = await POST(new Request("http://localhost/api/auth/refresh", {
      method: "POST",
      headers: { Cookie: "bc_refresh_token=refresh-old" },
    }));

    expect(response.status).toBe(503);
    expect(response.headers.get("set-cookie")).toBeNull();
  });
});
