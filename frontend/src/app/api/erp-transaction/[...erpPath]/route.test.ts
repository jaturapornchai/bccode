import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET, POST, PUT } from "./route";

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

  // จอปรับปรุงสถานะใบรับสินค้า / พิมพ์ใบวางบิลอัตโนมัติ / วันที่ส่งของ อ่านพารามิเตอร์ชื่อเหล่านี้ที่ backend — BFF ต้องส่งต่อครบ
  it("forwards the filter params of the tool screens and drops unknown ones", async () => {
    let upstream = "";
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string | URL | Request) => {
        upstream = String(url);
        return Response.json({ success: true, data: [] });
      }),
    );

    const names = [
      "partner",
      "from_date",
      "to_date",
      "from_docno",
      "to_docno",
      "from_vendor",
      "to_vendor",
      "from_dept",
      "to_dept",
      "from_cust",
      "to_cust",
      "sale_code",
      "limit",
      "offset",
      "status",
    ];
    const query = names.map((name) => `${name}=v_${name}`).join("&");

    const response = await GET(
      new Request(`http://localhost/api/erp-transaction/goodsreceive/receipt-adjustments?${query}&evil=1&__proto__=2`, {
        headers: { Authorization: "Bearer access-token", "x-bc-backend-url": "http://localhost:8888/goapi" },
      }),
      { params: Promise.resolve({ erpPath: ["goodsreceive", "receipt-adjustments"] }) },
    );

    expect(response.status).toBe(200);
    const forwarded = new URL(upstream).searchParams;
    for (const name of names) expect(forwarded.get(name), name).toBe(`v_${name}`);
    expect(forwarded.has("evil")).toBe(false);
    expect(forwarded.has("__proto__")).toBe(false);
  });

  it("still rejects dot segments without calling mainapi", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await get("sale-invoice/..");

    expect(response.status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  // ยอดเงินที่ลูกค้า API ส่งเป็นตัวเลข JSON ความละเอียดสูงต้องถึง backend ตรงตัว (เดิม parse แล้ว stringify ใหม่ = ปัดเป็น double)
  it.each([
    ["POST", POST, "sale-invoice"],
    ["PUT", PUT, "sale-invoice/IV6910-0001"],
  ] as const)("forwards %s amounts exactly as sent", async (method, handler, path) => {
    const raw = '{"docno":"IV6910-0001","totalamount":9999999999.99999999,"details":[{"qty":0.1,"price":"0.2"}]}';
    let forwarded: unknown;
    vi.stubGlobal(
      "fetch",
      vi.fn(async (_url: string | URL | Request, init?: RequestInit) => {
        forwarded = init?.body;
        return Response.json({ success: true, data: {} });
      }),
    );

    const response = await handler(
      new Request(`http://localhost/api/erp-transaction/${path}`, {
        method,
        body: raw,
        headers: { Authorization: "Bearer access-token", "x-bc-backend-url": "http://localhost:8888/goapi" },
      }),
      { params: Promise.resolve({ erpPath: path.split("/") }) },
    );

    expect(response.status).toBe(200);
    expect(forwarded).toBe(raw);
  });

  it("still rejects a body that is not JSON without calling mainapi", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/erp-transaction/sale-invoice", {
        method: "POST",
        body: "{not json",
        headers: { Authorization: "Bearer access-token", "x-bc-backend-url": "http://localhost:8888/goapi" },
      }),
      { params: Promise.resolve({ erpPath: ["sale-invoice"] }) },
    );

    expect(response.status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
