import { afterEach, describe, expect, it, vi } from "vitest";
import { POST } from "./route";

describe("LINE OA user route", () => {
  afterEach(() => {
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
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/goapi/api/user/lineoa/link");
      expect(init?.method).toBe("POST");
      expect((init?.headers as Record<string, string>).Authorization).toBe("Bearer token");
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
        Authorization: "Bearer token",
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
});
