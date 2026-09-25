import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { POST } from "./route";

const verifyUrl = "http://localhost:8888/verify-token";
const bridgeCreateUrl = "https://bridge.example/liff/api/login?action=create";
const bindUrl = "http://localhost:8888/profile/link-line/code";

function hrefOf(url: string | URL | Request): string {
  return typeof url === "string" ? url : url instanceof URL ? url.toString() : url.url;
}

function codeRequest(init: RequestInit = {}): Request {
  return new Request("http://localhost/api/auth/line/code", { method: "POST", ...init });
}

describe("LINE link code route", () => {
  beforeEach(() => {
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
    process.env.BC_AUTH_BRIDGE_URL = "https://bridge.example/liff";
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    delete process.env.BC_AUTH_BRIDGE_URL;
  });

  it("rejects a request without a bearer token before any outbound call", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(codeRequest());

    expect(response.status).toBe(401);
    expect(await response.json()).toMatchObject({ success: false });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects a bearer token that mainapi does not accept and never calls the bridge", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request) => {
      if (hrefOf(url) === verifyUrl) {
        return Response.json({ success: false, message: "Token Invalid." }, { status: 401 });
      }
      throw new Error(`unexpected fetch ${hrefOf(url)}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(codeRequest({ headers: { Authorization: "Bearer forged" } }));

    expect(response.status).toBe(401);
    expect(await response.json()).toMatchObject({ success: false });
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(hrefOf(fetchMock.mock.calls[0][0])).toBe(verifyUrl);
  });

  it("mints a code once for a live session without sending the user token to the bridge", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const href = hrefOf(url);
      const headers = new Headers(init?.headers);
      if (href === verifyUrl) {
        expect(headers.get("authorization")).toBe("Bearer token");
        return Response.json({ success: true, uid: "u-1", username: "somchai" });
      }
      if (href === bridgeCreateUrl) {
        expect(headers.get("authorization")).toBeNull();
        expect(init?.body).toBeUndefined();
        return Response.json({
          data: { code: "123456", loginUrl: "https://bridge.example/liff/login/123456", expiresAt: "2026-09-25T10:05:00Z" },
        });
      }
      if (href === bindUrl) {
        expect(init?.method).toBe("POST");
        expect(headers.get("authorization")).toBe("Bearer token");
        expect(JSON.parse(String(init?.body))).toEqual({ code: "123456" });
        return Response.json({ success: true });
      }
      throw new Error(`unexpected fetch ${href}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(codeRequest({ headers: { Authorization: "Bearer token" } }));

    expect(response.status).toBe(200);
    expect(await response.json()).toEqual({
      success: true,
      code: "123456",
      loginUrl: "https://bridge.example/liff/login/123456",
      expiresAt: "2026-09-25T10:05:00Z",
    });
    expect(fetchMock.mock.calls.map(([url]) => hrefOf(url))).toEqual([verifyUrl, bridgeCreateUrl, bindUrl]);
  });

  it("ignores a caller-supplied bridge URL or extra fields in the body", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const href = hrefOf(url);
      if (href === verifyUrl) return Response.json({ success: true });
      if (href === bridgeCreateUrl) {
        expect(init?.body).toBeUndefined();
        return Response.json({ code: "654321", loginUrl: "https://bridge.example/liff/login/654321" });
      }
      if (href === bindUrl) return Response.json({ success: true });
      throw new Error(`unexpected fetch ${href}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(codeRequest({
      headers: { Authorization: "Bearer token", "Content-Type": "application/json" },
      body: '{"bridgeUrl":"https://attacker.example","action":"delete",',
    }));

    expect(response.status).toBe(200);
    expect(await response.json()).toMatchObject({ success: true, code: "654321" });
    expect(fetchMock.mock.calls.map(([url]) => hrefOf(url))).toEqual([verifyUrl, bridgeCreateUrl, bindUrl]);
  });

  it("does not hand out a code that mainapi refuses to bind to this user", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request) => {
      const href = hrefOf(url);
      if (href === verifyUrl) return Response.json({ success: true, uid: "u-1" });
      if (href === bridgeCreateUrl) return Response.json({ code: "123456", loginUrl: "https://bridge.example/liff/login/123456" });
      if (href === bindUrl) return Response.json({ success: false, message: "bound to another account" }, { status: 409 });
      throw new Error(`unexpected fetch ${href}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(codeRequest({ headers: { Authorization: "Bearer token" } }));
    const json = await response.json();

    expect(response.status).toBe(409);
    expect(json).toMatchObject({ success: false, message: "bound to another account" });
    expect(json).not.toHaveProperty("code");
    expect(json).not.toHaveProperty("loginUrl");
  });

  it("does not mint a code when mainapi cannot be reached", async () => {
    const fetchMock = vi.fn(async () => {
      throw new TypeError("fetch failed");
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(codeRequest({ headers: { Authorization: "Bearer token" } }));

    expect(response.status).toBe(504);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
