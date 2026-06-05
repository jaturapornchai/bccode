import { afterEach, describe, expect, it, vi } from "vitest";
import { POST } from "./route";

describe("LINE link status route", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    delete process.env.BC_AUTH_BRIDGE_URL;
  });

  it("requires an authenticated database user session", async () => {
    const response = await POST(new Request("http://localhost/api/auth/line/link/status", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ backendUrl: "http://localhost:8888/goapi", code: "123456" }),
    }));

    expect(response.status).toBe(401);
  });

  it("links confirmed LINE profile to the current user", async () => {
    process.env.BC_AUTH_BRIDGE_URL = "https://bridge.example/liff";

    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const href = typeof url === "string" ? url : url instanceof URL ? url.toString() : url.url;
      if (href === "https://bridge.example/liff/api/login?code=123456") {
        return Response.json({
          confirmed: true,
          data: {
            userId: "U123",
            displayName: "Jead",
            pictureUrl: "https://line.example/picture.png",
          },
        });
      }

      if (href === "http://localhost:8888/profile/link-line") {
        expect(init?.method).toBe("PUT");
        expect((init?.headers as Record<string, string>).Authorization).toBe("Bearer token");
        expect(JSON.parse(String(init?.body))).toEqual({
          lineuserid: "U123",
          linedisplayname: "Jead",
          linepictureurl: "https://line.example/picture.png",
        });
        return Response.json({ success: true });
      }

      throw new Error(`unexpected fetch ${href}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/line/link/status", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer token",
      },
      body: JSON.stringify({ backendUrl: "http://localhost:8888/goapi", code: "123456" }),
    }));
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({
      success: true,
      status: "success",
      backendUrl: "http://localhost:8888/goapi",
      mainApiUrl: "http://localhost:8888",
      user: { lineUserId: "U123", displayName: "Jead" },
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});
