import { afterEach, describe, expect, it, vi } from "vitest";
import { GET } from "./route";

const context = (master: string) => ({ params: Promise.resolve({ master }) });

describe("master picker proxy", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("reads business types from the organization endpoint and flattens entries", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/organization/business-type?q=%E0%B8%84%E0%B9%89%E0%B8%B2&limit=20");
      expect(init?.method).toBe("GET");
      return Response.json({
        success: true,
        data: [{ guidfixed: "bt-1", code: "RT", names: [{ code: "th", name: "ค้าปลีก" }], isdefault: true }],
        total: 1,
      });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/product-barcode/master/businesstype?q=ค้า&limit=20&backendUrl=http://localhost:8888/goapi", {
        headers: { Authorization: "Bearer test-token" },
      }),
      context("businesstype"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toEqual({ success: true, data: [{ guidfixed: "bt-1", code: "RT", names: [{ code: "th", name: "ค้าปลีก" }] }] });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("rejects masters that are not on the allowlist", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    // brand/company were removed 2026-09-23 (no caller, brand backend gone with MongoDB)
    for (const master of ["users", "brand", "company"]) {
      const response = await GET(new Request(`http://localhost/api/product-barcode/master/${master}`), context(master));
      expect(response.status).toBe(404);
    }
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
