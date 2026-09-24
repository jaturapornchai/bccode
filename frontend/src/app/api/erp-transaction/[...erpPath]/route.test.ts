import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET } from "./route";

const get = (path: string) =>
  GET(
    new Request(`http://localhost/api/erp-transaction/${path}`, {
      headers: { Authorization: "Bearer access-token", "x-bc-backend-url": "http://localhost:8888/goapi" },
    }),
    { params: Promise.resolve({ erpPath: path.split("/") }) },
  );

describe("erp-transaction BFF path check", () => {
  beforeEach(() => {
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
  });
  afterEach(() => {
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    vi.unstubAllGlobals();
  });

  // เลขเอกสารภาษาไทยที่มีสระบน/ล่างและวรรณยุกต์ (\p{M}) ต้องผ่าน — docs/kms/17-dev-gotchas.md
  it("proxies a Thai document code with vowel and tone marks", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request) => {
      expect(String(url)).toBe(`http://localhost:8888/transaction/sale-invoice/${encodeURIComponent("ใบกำกับที่๑")}?`);
      return Response.json({ success: true, data: {} });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await get("sale-invoice/ใบกำกับที่๑");

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("still rejects dot segments without calling mainapi", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await get("sale-invoice/..");

    expect(response.status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
