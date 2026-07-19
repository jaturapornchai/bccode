import { afterEach, describe, expect, it, vi } from "vitest";
import { GET, POST } from "./route";

describe("product route", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns PostgreSQL list search errors without falling back to another product source", async () => {
    const fetchMock = vi.fn(
      async (url: string | URL | Request, init?: RequestInit) => {
        const requestUrl = String(url);
        expect(init?.headers).toMatchObject({
          Authorization: "Bearer test-token",
        });

        if (requestUrl === "http://localhost:8888/goapi/api/product/search") {
          expect(init?.method).toBe("POST");
          expect(JSON.parse(String(init?.body))).toMatchObject({
            holdingcode: "bctest01",
            search: "",
            limit: 80,
            offset: 0,
            usecache: false,
          });
          return Response.json(
            { status: "error", error: "Query execution failed" },
            { status: 500 },
          );
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

    expect(response.status).toBe(500);
    expect(json).toMatchObject({
      success: false,
      message: "Query execution failed",
      source: "pgsql",
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("maps PostgreSQL list rows without inventing guidfixed values", async () => {
    const fetchMock = vi.fn(
      async (url: string | URL | Request, init?: RequestInit) => {
        expect(String(url)).toBe(
          "http://localhost:8888/goapi/api/product/search",
        );
        expect(JSON.parse(String(init?.body))).toMatchObject({
          usecache: false,
        });
        return Response.json({
          status: "success",
          count: 1,
          data: [
            {
              itemcode: "SKU001",
              itemname: "สินค้า",
              unitcode: "PCS",
              unitname: "ชิ้น",
            },
          ],
        });
      },
    );
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/product?holdingcode=bctest01", {
        headers: {
          Authorization: "Bearer test-token",
          "x-bc-backend-url": "http://localhost:8888/goapi",
        },
      }),
      productContext(),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json.data).toEqual([
      expect.objectContaining({
        code: "SKU001",
        guidfixed: "",
        _source: "pgsql",
      }),
    ]);
  });

  it("proxies the explicit product resync action", async () => {
    const fetchMock = vi.fn(
      async (url: string | URL | Request, init?: RequestInit) => {
        expect(String(url)).toBe("http://localhost:8888/product/resync");
        expect(init?.method).toBe("POST");
        expect(init?.body).toBeUndefined();
        return Response.json({
          success: true,
          data: { rebuilt: 2, published: 2 },
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
      data: { rebuilt: 2, published: 2 },
    });
  });
});

function productContext(productPath?: string[]) {
  return {
    params: Promise.resolve({ productPath }),
  };
}
