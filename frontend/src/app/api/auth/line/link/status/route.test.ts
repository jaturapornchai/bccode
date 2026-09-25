import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { POST as mintCode } from "../../code/route";
import { POST } from "./route";

const verifyUrl = "http://localhost:8888/verify-token";
const bindUrl = "http://localhost:8888/profile/link-line/code";
const checkUrl = "http://localhost:8888/profile/link-line/code/check";
const linkUrl = "http://localhost:8888/profile/link-line";
const bridgeCreateUrl = "https://bridge.example/liff/api/login?action=create";
const bridgePollUrl = "https://bridge.example/liff/api/login?code=123456";

function hrefOf(url: string | URL | Request): string {
  return typeof url === "string" ? url : url instanceof URL ? url.toString() : url.url;
}

function statusRequest(authorization: string | null, code = "123456"): Request {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (authorization) headers.Authorization = authorization;
  return new Request("http://localhost/api/auth/line/link/status", {
    method: "POST",
    headers,
    body: JSON.stringify({ backendUrl: "http://localhost:8888/goapi", code }),
  });
}

function mintRequest(authorization: string): Request {
  return new Request("http://localhost/api/auth/line/code", { method: "POST", headers: { Authorization: authorization } });
}

// Fake mainapi (live sessions + the code→uid binding kept in cache_entries) and a LINE bridge that has
// already confirmed code 123456. Mirrors backend/internal/authentication/line_link_code.go.
function fakeServers() {
  const sessions: Record<string, string> = { "Bearer token-a": "uid-a", "Bearer token-b": "uid-b" };
  const owners = new Map<string, string>();
  const linked: Array<{ uid: string; body: Record<string, string> }> = [];
  const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
    const href = hrefOf(url);
    const uid = sessions[new Headers(init?.headers).get("authorization") ?? ""];
    const body = init?.body ? (JSON.parse(String(init.body)) as Record<string, string>) : {};

    if (href === bridgeCreateUrl) {
      return Response.json({ data: { code: "123456", loginUrl: "https://bridge.example/liff/login/123456" } });
    }
    if (href === bridgePollUrl) {
      return Response.json({ confirmed: true, data: { userId: "U123", displayName: "Jead", pictureUrl: "https://line.example/p.png" } });
    }
    if (!uid) return Response.json({ success: false, message: "Token Invalid." }, { status: 401 });
    if (href === verifyUrl) return Response.json({ success: true, uid });
    if (href === bindUrl) {
      const owner = owners.get(body.code);
      if (owner && owner !== uid) return Response.json({ success: false, message: "bound to another account" }, { status: 409 });
      owners.set(body.code, uid);
      return Response.json({ success: true });
    }
    if (href === checkUrl || href === linkUrl) {
      if (owners.get(body.code) !== uid) {
        return Response.json({ success: false, message: "LINE link code is not yours or expired" }, { status: 404 });
      }
      if (href === linkUrl) {
        linked.push({ uid, body });
        owners.delete(body.code);
      }
      return Response.json({ success: true });
    }
    throw new Error(`unexpected fetch ${href}`);
  });
  vi.stubGlobal("fetch", fetchMock);
  const calls = () => fetchMock.mock.calls.map(([url]) => hrefOf(url));
  return { owners, linked, fetchMock, calls };
}

describe("LINE link status route", () => {
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
    const { fetchMock } = fakeServers();

    const response = await POST(statusRequest(null));

    expect(response.status).toBe(401);
    expect(await response.json()).toEqual({ success: false, message: "ไม่พบ token กรุณาเข้าสู่ระบบใหม่" });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects a forged bearer token with the same 401 body and never calls the bridge", async () => {
    const { calls } = fakeServers();

    const response = await POST(statusRequest("Bearer forged"));

    expect(response.status).toBe(401);
    expect(await response.json()).toEqual({ success: false, message: "ไม่พบ token กรุณาเข้าสู่ระบบใหม่" });
    expect(calls()).toEqual([verifyUrl]);
  });

  it("rejects user B polling a code minted by user A without asking the bridge or linking", async () => {
    const servers = fakeServers();
    const minted = await mintCode(mintRequest("Bearer token-a"));
    expect(minted.status).toBe(200);
    expect(servers.owners.get("123456")).toBe("uid-a");
    servers.fetchMock.mockClear();

    const response = await POST(statusRequest("Bearer token-b"));

    expect(response.status).toBe(404);
    expect(await response.json()).toMatchObject({ success: false, status: "failed" });
    expect(servers.calls()).toEqual([verifyUrl, checkUrl]);
    expect(servers.linked).toEqual([]);
    expect(servers.owners.get("123456")).toBe("uid-a");
  });

  it("links the confirmed LINE profile when the minter polls its own code", async () => {
    const servers = fakeServers();
    await mintCode(mintRequest("Bearer token-a"));
    servers.fetchMock.mockClear();

    const response = await POST(statusRequest("Bearer token-a"));

    expect(response.status).toBe(200);
    expect(await response.json()).toMatchObject({
      success: true,
      status: "success",
      backendUrl: "http://localhost:8888/goapi",
      mainApiUrl: "http://localhost:8888",
      user: { lineUserId: "U123", displayName: "Jead" },
    });
    expect(servers.calls()).toEqual([verifyUrl, checkUrl, bridgePollUrl, linkUrl]);
    expect(servers.linked).toEqual([
      {
        uid: "uid-a",
        body: { code: "123456", lineuserid: "U123", linedisplayname: "Jead", linepictureurl: "https://line.example/p.png" },
      },
    ]);
    expect(servers.owners.has("123456")).toBe(false);

    // The code is single-use: polling it again after the link is refused before the bridge.
    servers.fetchMock.mockClear();
    const again = await POST(statusRequest("Bearer token-a"));
    expect(again.status).toBe(404);
    expect(servers.calls()).toEqual([verifyUrl, checkUrl]);
  });

  it("rejects a code this user never minted (unknown or expired) without asking the bridge", async () => {
    const servers = fakeServers();

    const response = await POST(statusRequest("Bearer token-a"));

    expect(response.status).toBe(404);
    expect(servers.calls()).toEqual([verifyUrl, checkUrl]);
    expect(servers.linked).toEqual([]);
  });

  it("does not ask the bridge when mainapi cannot be reached", async () => {
    const fetchMock = vi.fn(async () => {
      throw new TypeError("fetch failed");
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(statusRequest("Bearer token-a"));

    expect(response.status).toBe(504);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledWith(verifyUrl, expect.anything());
  });
});
