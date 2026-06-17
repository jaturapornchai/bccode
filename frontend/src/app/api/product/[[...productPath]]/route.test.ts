import { afterEach, describe, expect, it, vi } from "vitest";
import { GET } from "./route";

describe("product route", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns PostgreSQL list search errors without falling back to another product source", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      expect(init?.headers).toMatchObject({ Authorization: "Bearer test-token" });

      if (requestUrl === "http://localhost:8888/goapi/api/product/search") {
        expect(init?.method).toBe("POST");
        expect(JSON.parse(String(init?.body))).toMatchObject({
          holdingcode: "bctest01",
          search: "",
          limit: 80,
          offset: 0,
        });
        return Response.json({ status: "error", error: "Query execution failed" }, { status: 500 });
      }

      throw new Error(`Unexpected URL ${requestUrl}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/product?q=&limit=80&holdingcode=bctest01", {
        headers: {
          Authorization: "Bearer test-token",
          "x-bc-backend-url": "http://localhost:8888/goapi",
        },
      }),
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
});

function productContext(productPath?: string[]) {
  return {
    params: Promise.resolve({ productPath }),
  };
}
