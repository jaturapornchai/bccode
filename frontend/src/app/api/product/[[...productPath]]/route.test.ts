import { afterEach, describe, expect, it, vi } from "vitest";
import { GET, POST } from "./route";

describe("product route", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("uses the authenticated Company-scoped Product API for list reads", async () => {
    const fetchMock = vi.fn(
      async (url: string | URL | Request, init?: RequestInit) => {
        const requestUrl = String(url);
        expect(init?.headers).toMatchObject({
          Authorization: "Bearer test-token",
        });

        if (requestUrl === "http://localhost:8888/product?q=&page=1&limit=80") {
          expect(init?.method).toBe("GET");
          expect(init?.body).toBeUndefined();
          return Response.json({ success: true, data: [] });
        }

        throw new Error(`Unexpected URL ${requestUrl}`);
      },
    );
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request(
        "http://localhost/api/product?q=&limit=80&holdingcode=bctest01",
        {
          headers: {
            Authorization: "Bearer test-token",
            "x-bc-backend-url": "http://localhost:8888/goapi",
          },
        },
      ),
      productContext(),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({
      success: true,
      data: [],
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("proxies the explicit product resync action", async () => {
    const fetchMock = vi.fn(
      async (url: string | URL | Request, init?: RequestInit) => {
        expect(String(url)).toBe("http://localhost:8888/product/resync");
        expect(init?.method).toBe("POST");
        expect(init?.body).toBeUndefined();
        return Response.json({
          success: true,
          data: { rebuilt: 2, queued: 2, published: 0 },
        });
      },
    );
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/product/resync", {
        method: "POST",
        headers: {
          Authorization: "Bearer test-token",
          "x-bc-backend-url": "http://localhost:8888/goapi",
        },
      }),
      productContext(["resync"]),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({
      success: true,
      data: { rebuilt: 2, queued: 2, published: 0 },
    });
  });
});

function productContext(productPath?: string[]) {
  return {
    params: Promise.resolve({ productPath }),
  };
}
