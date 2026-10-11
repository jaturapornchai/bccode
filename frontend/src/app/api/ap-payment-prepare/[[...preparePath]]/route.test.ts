import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET, POST, PUT } from "./route";

const headers = { Authorization: "Bearer access-token", "x-bc-backend-url": "http://localhost:8888/goapi" };
const params = (path: string) => ({ params: Promise.resolve({ preparePath: path ? path.split("/") : undefined }) });
const request = (path: string, init?: RequestInit) =>
  new Request(`http://localhost/api/ap-payment-prepare${path ? `/${path}` : ""}`, { ...init, headers });

function captureUpstream() {
  const calls: { url: string; method?: string; body?: unknown }[] = [];
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      calls.push({ url: String(url), method: init?.method, body: init?.body });
      return Response.json({ success: true, data: {} });
    }),
  );
  return calls;
}

describe("ap-payment-prepare BFF", () => {
  beforeEach(() => {
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
  });
  afterEach(() => {
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    vi.unstubAllGlobals();
  });

  it.each([
    ["list", "/ap-payment-prepare/list"],
    ["outstanding", "/ap-payment-prepare/outstanding"],
    ["creditors", "/ap-payment-prepare/creditors"],
    ["ตจ6910-0001", `/ap-payment-prepare/${encodeURIComponent("ตจ6910-0001")}`],
    ["PP6910-0001/outstanding", "/ap-payment-prepare/PP6910-0001/outstanding"],
    ["PP6910-0001/events", "/ap-payment-prepare/PP6910-0001/events"],
  ])("proxies GET %s to the backend route group", async (path, upstream) => {
    const calls = captureUpstream();
    const response = await GET(request(path), params(path));
    expect(response.status).toBe(200);
    expect(calls[0].url).toBe(`http://localhost:8888${upstream}`);
  });

  it("forwards only the filters the backend reads", async () => {
    const calls = captureUpstream();
    const names = ["q", "page", "limit", "fromdate", "todate", "partner", "status"];
    const query = names.map((name) => `${name}=v_${name}`).join("&");
    const response = await GET(
      new Request(`http://localhost/api/ap-payment-prepare/list?${query}&evil=1`, { headers }),
      params("list"),
    );
    expect(response.status).toBe(200);
    const forwarded = new URL(calls[0].url).searchParams;
    for (const name of names) expect(forwarded.get(name), name).toBe(`v_${name}`);
    expect(forwarded.has("evil")).toBe(false);
  });

  it.each([
    ["GET", "", GET],
    ["GET", "list/outstanding", GET],
    ["GET", "PP6910-0001/lines", GET],
    ["GET", "PP6910-0001/events/1", GET],
    ["POST", "PP6910-0001", POST],
    ["POST", "PP6910-0001/approve", POST],
    ["PUT", "", PUT],
    ["PUT", "list", PUT],
    ["PUT", "PP6910-0001/cancel", PUT],
  ] as const)("rejects %s %s without calling mainapi", async (method, path, handler) => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const response = await handler(request(path, method === "GET" ? undefined : { method, body: "{}" }), params(path));
    expect(response.status).toBe(404);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects dot segments", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const response = await GET(request("../events"), params("../events"));
    expect(response.status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  // ยอดเงินเป็นข้อความทศนิยม — BFF ส่งข้อความเดิมต่อตรงตัว ไม่ parse แล้ว stringify ใหม่ (Rule 20)
  it.each([
    ["POST", POST, "", "/ap-payment-prepare"],
    ["PUT", PUT, "PP6910-0001", "/ap-payment-prepare/PP6910-0001"],
    ["POST", POST, "PP6910-0001/cancel", "/ap-payment-prepare/PP6910-0001/cancel"],
  ] as const)("forwards %s body exactly to %s", async (method, handler, path, upstream) => {
    const raw = '{"docno":"PP6910-0001","lines":[{"debt_document_id":"AP-PUR6910-0003","payamount":"9999999999999999.99"}],"version":3}';
    const calls = captureUpstream();
    const response = await handler(request(path, { method, body: raw }), params(path));
    expect(response.status).toBe(200);
    expect(calls[0]).toMatchObject({ url: `http://localhost:8888${upstream}`, method, body: raw });
  });

  it("rejects a body that is not a JSON object", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    for (const body of ["{not json", "[1,2]"]) {
      const response = await POST(request("", { method: "POST", body }), params(""));
      expect(response.status).toBe(400);
    }
    expect(fetchMock).not.toHaveBeenCalled();
  });

  // ข้อผิดพลาดทางธุรกิจ (409 พร้อม code) ส่งต่อให้จอแสดงข้อความไทยของ backend
  it("passes a business rule error through with its code", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => Response.json({ success: false, code: "ap_prepare_paid_locked", message: "ใบเตรียมจ่ายนี้ถูกนำไปจ่ายแล้ว" }, { status: 409 })),
    );
    const response = await PUT(request("PP6910-0001", { method: "PUT", body: "{}" }), params("PP6910-0001"));
    const payload = await response.json();
    expect(payload).toMatchObject({ code: "ap_prepare_paid_locked" });
  });
});
