import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { POST } from "./route";

describe("logout route", () => {
  beforeEach(() => {
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
  });

  afterEach(() => {
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    vi.unstubAllGlobals();
  });

  it("revokes with the supplied access token without rotating refresh, and clears the cookie", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/logout");
      expect((init?.headers as Record<string, string>).Authorization).toBe("Bearer access-old");
      return Response.json({ success: true });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/logout", {
      method: "POST",
      headers: {
        Authorization: "Bearer access-old",
        Cookie: "bc_refresh_token=refresh-current",
      },
    }));

    expect(await response.json()).toEqual({ success: true });
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(response.headers.get("set-cookie")).toMatch(/bc_refresh_token=;.*Max-Age=0/i);
  });

  it("keeps the refresh cookie when backend revocation fails", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Response.json({ success: false }, { status: 503 })));

    const response = await POST(new Request("http://localhost/api/auth/logout", {
      method: "POST",
      headers: {
        Authorization: "Bearer access-old",
        Cookie: "bc_refresh_token=refresh-current",
      },
    }));

    expect(response.status).toBe(503);
    expect(await response.json()).toEqual({ success: false, message: "ไม่สามารถเพิกถอน Session ได้" });
    expect(response.headers.get("set-cookie")).toBeNull();
  });

  it.each([401, 403])("rotates and retries once when the supplied access token returns %i", async (status) => {
    const calls: string[] = [];
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      calls.push(String(url));
      if (calls.length === 1) {
        expect((init?.headers as Record<string, string>).Authorization).toBe("Bearer access-old");
        return Response.json({ success: false }, { status });
      }
      if (calls.length === 2) {
        expect(JSON.parse(String(init?.body))).toEqual({ token: "refresh-current" });
        return Response.json({ success: true, token: "access-new", refresh: "refresh-rotated" });
      }
      expect((init?.headers as Record<string, string>).Authorization).toBe("Bearer access-new");
      return Response.json({ success: true });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/logout", {
      method: "POST",
      headers: {
        Authorization: "Bearer access-old",
        Cookie: "bc_refresh_token=refresh-current",
      },
    }));

    expect(await response.json()).toEqual({ success: true });
    expect(calls).toEqual([
      "http://localhost:8888/logout",
      "http://localhost:8888/refresh",
      "http://localhost:8888/logout",
    ]);
    expect(response.headers.get("set-cookie")).toMatch(/bc_refresh_token=;.*Max-Age=0/i);
  });

  it("does not rotate after a non-auth revocation failure", async () => {
    const fetchMock = vi.fn(async () => Response.json({ success: false }, { status: 503 }));
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/logout", {
      method: "POST",
      headers: {
        Authorization: "Bearer access-old",
        Cookie: "bc_refresh_token=refresh-current",
      },
    }));

    expect(response.status).toBe(503);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(response.headers.get("set-cookie")).toBeNull();
  });

  it("does not claim success when refresh cannot produce a revocable access token", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Response.json({ success: false }, { status: 503 })));

    const response = await POST(new Request("http://localhost/api/auth/logout", {
      method: "POST",
      headers: { Cookie: "bc_refresh_token=refresh-current" },
    }));

    expect(response.status).toBe(503);
    expect(await response.json()).toEqual({ success: false, message: "ไม่สามารถเพิกถอน Session ได้" });
    expect(response.headers.get("set-cookie")).toBeNull();
  });

  it("returns the rotated refresh cookie when revocation fails after rotation", async () => {
    const fetchMock = vi.fn(async (url: string | URL) => {
      if (String(url).endsWith("/refresh")) {
        return Response.json({ success: true, token: "access-current", refresh: "refresh-rotated" });
      }
      return Response.json({ success: false }, { status: 503 });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/logout", {
      method: "POST",
      headers: { Cookie: "bc_refresh_token=refresh-current" },
    }));

    expect(response.status).toBe(503);
    expect(response.headers.get("set-cookie")).toMatch(/bc_refresh_token=refresh-rotated/i);
  });

  it("returns the rotated refresh cookie when the retry throws", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request) => {
      if (String(url).endsWith("/refresh")) {
        return Response.json({ success: true, token: "access-current", refresh: "refresh-rotated" });
      }
      throw new Error("backend unavailable");
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/logout", {
      method: "POST",
      headers: { Cookie: "bc_refresh_token=refresh-current" },
    }));

    expect(response.status).toBe(504);
    expect(await response.json()).toEqual({ success: false, message: "ไม่สามารถเพิกถอน Session ได้" });
    expect(response.headers.get("set-cookie")).toMatch(/bc_refresh_token=refresh-rotated/i);
  });

  it("does not report success without a bearer or refresh credential", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/logout", { method: "POST" }));

    expect(response.status).toBe(401);
    expect(await response.json()).toEqual({ success: false, message: "ไม่สามารถเพิกถอน Session ได้" });
    expect(response.headers.get("set-cookie")).toBeNull();
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
