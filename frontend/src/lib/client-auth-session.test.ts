import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

class MemoryStorage implements Storage {
  private readonly values = new Map<string, string>();

  get length() { return this.values.size; }
  clear() { this.values.clear(); }
  getItem(key: string) { return this.values.get(key) ?? null; }
  key(index: number) { return [...this.values.keys()][index] ?? null; }
  removeItem(key: string) { this.values.delete(key); }
  setItem(key: string, value: string) { this.values.set(key, value); }
}

describe("client auth session", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.stubGlobal("localStorage", new MemoryStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("keeps the access token in memory and stores metadata only", async () => {
    const { getAuthSession, setAuthSession } = await import("./client-auth-session");

    setAuthSession({
      token: "access-secret",
      username: "uat_user",
      backendUrl: "http://localhost:8888/goapi",
      method: "password",
    });

    expect(getAuthSession()?.token).toBe("access-secret");
    expect(localStorage.getItem("bc_auth")).toBe(
      JSON.stringify({
        username: "uat_user",
        backendUrl: "http://localhost:8888/goapi",
        method: "password",
      }),
    );
  });

  it("deduplicates reload refresh and scrubs legacy stored tokens", async () => {
    localStorage.setItem("bc_auth", JSON.stringify({
      token: "legacy-access",
      refresh: "legacy-refresh",
      username: "uat_user",
      backendUrl: "http://localhost:8888/goapi",
    }));
    const fetchMock = vi.fn(async () => Response.json({ success: true, token: "access-new" }));
    vi.stubGlobal("fetch", fetchMock);
    const { restoreAuthSession } = await import("./client-auth-session");

    const [first, second] = await Promise.all([restoreAuthSession(), restoreAuthSession()]);

    expect(first?.token).toBe("access-new");
    expect(second?.token).toBe("access-new");
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(localStorage.getItem("bc_auth")).toBe(
      JSON.stringify({ username: "uat_user", backendUrl: "http://localhost:8888/goapi" }),
    );
  });

  it("refreshes one 401 and retries once with the rotated access token", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(async (_input: RequestInfo | URL, init?: RequestInit) => {
        expect(new Headers(init?.headers).get("authorization")).toBe("Bearer access-old");
        return Response.json({ success: false }, { status: 401 });
      })
      .mockImplementationOnce(async (input: RequestInfo | URL) => {
        expect(String(input)).toBe("/api/auth/refresh");
        return Response.json({ success: true, token: "access-new" });
      })
      .mockImplementationOnce(async (_input: RequestInfo | URL, init?: RequestInit) => {
        expect(new Headers(init?.headers).get("authorization")).toBe("Bearer access-new");
        return Response.json({ success: true });
      });
    vi.stubGlobal("fetch", fetchMock);
    const { authFetch, getAuthSession, setAuthSession } = await import("./client-auth-session");
    setAuthSession({
      token: "access-old",
      username: "uat_user",
      backendUrl: "http://localhost:8888/goapi",
    });

    const response = await authFetch("/api/protected", {
      headers: { Authorization: "Bearer access-old" },
    });

    expect(response.ok).toBe(true);
    expect(getAuthSession()?.token).toBe("access-new");
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });

  it("refreshes configured direct-backend requests", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(Response.json({}, { status: 401 }))
      .mockResolvedValueOnce(Response.json({ success: true, token: "access-new" }))
      .mockResolvedValueOnce(Response.json({ success: true }));
    vi.stubGlobal("fetch", fetchMock);
    const { authFetch, setAuthSession } = await import("./client-auth-session");
    setAuthSession({
      token: "access-old",
      username: "uat_user",
      backendUrl: "http://localhost:8888/goapi",
    });

    const response = await authFetch("http://localhost:8888/organization/company", {
      headers: { Authorization: "Bearer access-old" },
    });

    expect(response.ok).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(new Headers(fetchMock.mock.calls[2]?.[1]?.headers).get("authorization")).toBe("Bearer access-new");
  });

  it("does not enter a refresh loop when the retried request is still unauthorized", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(Response.json({}, { status: 401 }))
      .mockResolvedValueOnce(Response.json({ success: true, token: "access-new" }))
      .mockResolvedValueOnce(Response.json({}, { status: 401 }));
    vi.stubGlobal("fetch", fetchMock);
    const { authFetch, setAuthSession } = await import("./client-auth-session");
    setAuthSession({
      token: "access-old",
      username: "uat_user",
      backendUrl: "http://localhost:8888/goapi",
    });

    const response = await authFetch("/api/protected", {
      headers: { Authorization: "Bearer access-old" },
    });

    expect(response.status).toBe(401);
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });

  it("coordinates concurrent 401 responses through one refresh", async () => {
    let protectedCalls = 0;
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      if (String(input) === "/api/auth/refresh") {
        return Response.json({ success: true, token: "access-new" });
      }
      protectedCalls += 1;
      return protectedCalls <= 2
        ? Response.json({}, { status: 401 })
        : Response.json({ success: true });
    });
    vi.stubGlobal("fetch", fetchMock);
    const { authFetch, setAuthSession } = await import("./client-auth-session");
    setAuthSession({
      token: "access-old",
      username: "uat_user",
      backendUrl: "http://localhost:8888/goapi",
    });

    const responses = await Promise.all([
      authFetch("/api/first", { headers: { Authorization: "Bearer access-old" } }),
      authFetch("/api/second", { headers: { Authorization: "Bearer access-old" } }),
    ]);

    expect(responses.every((response) => response.ok)).toBe(true);
    expect(fetchMock.mock.calls.filter(([input]) => String(input) === "/api/auth/refresh")).toHaveLength(1);
  });

  it("keeps session metadata when refresh fails transiently", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(Response.json({}, { status: 401 }))
      .mockResolvedValueOnce(Response.json({ success: false }, { status: 503 }));
    vi.stubGlobal("fetch", fetchMock);
    const { authFetch, getAuthSession, setAuthSession } = await import("./client-auth-session");
    setAuthSession({
      token: "access-old",
      username: "uat_user",
      backendUrl: "http://localhost:8888/goapi",
    });

    await authFetch("/api/protected", { headers: { Authorization: "Bearer access-old" } });

    expect(getAuthSession()?.token).toBe("access-old");
    expect(localStorage.getItem("bc_auth")).not.toBeNull();
  });

  it("clears the session only when refresh authentication is definitively rejected", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(Response.json({}, { status: 401 }))
      .mockResolvedValueOnce(Response.json({ success: false }, { status: 401 }));
    vi.stubGlobal("fetch", fetchMock);
    const { authFetch, getAuthSession, setAuthSession } = await import("./client-auth-session");
    setAuthSession({
      token: "access-old",
      username: "uat_user",
      backendUrl: "http://localhost:8888/goapi",
    });

    await authFetch("/api/protected", { headers: { Authorization: "Bearer access-old" } });

    expect(getAuthSession()).toBeNull();
    expect(localStorage.getItem("bc_auth")).toBeNull();
  });
});
