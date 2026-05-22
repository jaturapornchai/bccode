import { createHmac } from "crypto";
import { afterEach, describe, expect, it, vi } from "vitest";
import { POST } from "./route";

const SECRET = "test-secret";

describe("LINE OA user route", () => {
  afterEach(() => {
    delete process.env.JWT_SECRET_KEY;
    vi.unstubAllGlobals();
  });

  it("requires an authenticated user session", async () => {
    const response = await POST(new Request("http://localhost/api/line-oa/user", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        action: "profile",
        backendUrl: "http://localhost:8888/goapi",
        shopId: "SHOP001",
        username: "user@example.com",
      }),
    }));

    expect(response.status).toBe(401);
  });

  it("generates a LINE OA user link through goapi", async () => {
    process.env.JWT_SECRET_KEY = SECRET;
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/goapi/api/user/lineoa/link");
      expect(init?.method).toBe("POST");
      expect(JSON.parse(String(init?.body))).toEqual({
        shop_id: "SHOP001",
        username: "user@example.com",
      });
      return Response.json({ status: "success", link: "https://liff.line.me/123?token=abc" });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/line-oa/user", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${signJwt({ username: "user@example.com", shopid: "SHOP001" })}`,
      },
      body: JSON.stringify({
        action: "link",
        backendUrl: "http://localhost:8888/goapi",
        shopId: "SHOP001",
        username: "user@example.com",
      }),
    }));
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({
      success: true,
      status: "success",
      link: "https://liff.line.me/123?token=abc",
    });
  });

  it("rejects a LINE OA user request for another company", async () => {
    process.env.JWT_SECRET_KEY = SECRET;
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/line-oa/user", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${signJwt({ username: "user@example.com", shopid: "SHOP001" })}`,
      },
      body: JSON.stringify({
        action: "profile",
        backendUrl: "http://localhost:8888/goapi",
        shopId: "SHOP002",
        username: "user@example.com",
      }),
    }));

    expect(response.status).toBe(403);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});

function signJwt(payload: Record<string, unknown>): string {
  const encodedHeader = encodeBase64Url(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const encodedPayload = encodeBase64Url(JSON.stringify({ exp: Math.floor(Date.now() / 1000) + 60, ...payload }));
  const signature = createHmac("sha256", SECRET).update(`${encodedHeader}.${encodedPayload}`).digest("base64url");
  return `${encodedHeader}.${encodedPayload}.${signature}`;
}

function encodeBase64Url(value: string): string {
  return Buffer.from(value).toString("base64url");
}
