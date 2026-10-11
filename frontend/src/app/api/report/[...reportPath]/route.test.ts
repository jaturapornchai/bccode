import { afterEach, describe, expect, it, vi } from "vitest";
import { GET, POST } from "./route";

const headers = { Authorization: "Bearer test-token", "x-bc-backend-url": "http://localhost:8888", "Content-Type": "application/json", "Accept-Language": "en" };
const context = (...reportPath: string[]) => ({ params: Promise.resolve({ reportPath }) });
const post = (action: string, body: unknown, init: RequestInit = {}) => POST(new Request(`http://localhost/api/report/${action}`, { method: "POST", headers, body: typeof body === "string" ? body : JSON.stringify(body), ...init }), context(action));
const sentBody = (fetchMock: ReturnType<typeof vi.fn>, call = 0) => JSON.parse(String((fetchMock.mock.calls[call][1] as RequestInit).body)) as Record<string, unknown>;

describe("/api/report relay (central report registry)", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("registry: forwards module and language with the caller's bearer token", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: { can_print: true, can_export: true, reports: [] } }));
    vi.stubGlobal("fetch", fetchMock);
    const response = await GET(new Request("http://localhost/api/report/registry?module=gl&lang=th", { headers }), context("registry"));
    expect(response.status).toBe(200);
    expect((await response.json()).data.can_print).toBe(true);
    expect(String(fetchMock.mock.calls[0][0])).toContain("/report/v1/registry?module=gl&lang=th");
    expect((fetchMock.mock.calls[0][1] as RequestInit).headers).toMatchObject({ Authorization: "Bearer test-token" });
  });

  it("registry: needs a token, rejects odd module / language values and unknown paths", async () => {
    const fetchMock = vi.fn(); vi.stubGlobal("fetch", fetchMock);
    expect((await GET(new Request("http://localhost/api/report/registry"), context("registry"))).status).toBe(401);
    expect((await GET(new Request("http://localhost/api/report/registry?module=gl%2F..", { headers }), context("registry"))).status).toBe(400);
    expect((await GET(new Request("http://localhost/api/report/registry?lang=thai-language", { headers }), context("registry"))).status).toBe(400);
    expect((await GET(new Request("http://localhost/api/report/other", { headers }), context("other"))).status).toBe(404);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("run: forwards only the documented keys; scope keys in the body never reach the backend", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: { menu_id: "general-ledger", rows: [] } }));
    vi.stubGlobal("fetch", fetchMock);
    const response = await post("run", { menu_id: "general-ledger", params: { fiscal_year: "2569", account_from: "11000" }, sort: "date", page: 2, limit: 50, print: false, rows_per_page: 28, holdingcode: "other", businesscode: "99", branchcode: "9" });
    expect(response.status).toBe(200);
    expect(String(fetchMock.mock.calls[0][0])).toContain("/report/v1/run");
    expect(sentBody(fetchMock)).toEqual({ menu_id: "general-ledger", params: { fiscal_year: "2569", account_from: "11000" }, sort: "date", page: 2, limit: 50, print: false, rows_per_page: 28 });
  });

  it("run: rejects malformed bodies without calling the backend", async () => {
    const fetchMock = vi.fn(); vi.stubGlobal("fetch", fetchMock);
    const cases: unknown[] = [
      "not json",
      [],
      {},
      { menu_id: "../admin" },
      { menu_id: "general-ledger", params: { fiscal_year: 2569 } },
      { menu_id: "general-ledger", params: { "Bad Key": "x" } },
      { menu_id: "general-ledger", params: { note: "x".repeat(12_001) } },
      { menu_id: "general-ledger", page: "2" },
      { menu_id: "general-ledger", limit: 1.5 },
      { menu_id: "general-ledger", print: "yes" },
      { menu_id: "general-ledger", sort: "date; drop" },
    ];
    for (const body of cases) expect((await post("run", body)).status, JSON.stringify(body)).toBe(400);
    expect((await post("unknown", { menu_id: "general-ledger" })).status).toBe(404);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("run: a parameter error from the backend is relayed as HTTP 200 + success:false with the field to highlight", async () => {
    const error = { success: false, errorcode: "REPORT_PARAM_REQUIRED", code: "REPORT_PARAM_REQUIRED", message: "กรุณาเลือกปีบัญชี", message_th: "กรุณาเลือกปีบัญชี", field: "fiscal_year", statuscode: 422 };
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async () => Response.json(error, { status: 422 })));
    const response = await post("run", { menu_id: "general-ledger", params: {} });
    expect(response.status).toBe(200);
    expect(await response.json()).toMatchObject({ success: false, errorcode: "REPORT_PARAM_REQUIRED", field: "fiscal_year", message: "กรุณาเลือกปีบัญชี" });
  });

  it("run: permission and server errors keep their real status", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async () => Response.json({ success: false, errorcode: "REPORT_FORBIDDEN", code: "REPORT_FORBIDDEN", message: "ไม่มีสิทธิ์" }, { status: 403 })));
    expect((await post("run", { menu_id: "general-ledger" })).status).toBe(403);
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async () => Response.json({ success: false, errorcode: "REPORT_ENGINE_FAILED", code: "REPORT_ENGINE_FAILED", message: "x" }, { status: 500 })));
    expect((await post("run", { menu_id: "general-ledger" })).status).toBe(500);
  });

  it("export: always asks for csv and streams the file with its name and BOM untouched", async () => {
    const csv = "\uFEFFรหัสบัญชี,ชื่อบัญชี\r\n11000,เงินสด\r\n";
    const fetchMock = vi.fn().mockImplementation(async () => new Response(csv, { status: 200, headers: { "Content-Type": "text/csv; charset=utf-8", "Content-Disposition": "attachment; filename*=UTF-8''general-ledger.csv" } }));
    vi.stubGlobal("fetch", fetchMock);
    const response = await post("export", { menu_id: "general-ledger", params: { fiscal_year: "2569" }, format: "csv" });
    expect(response.status).toBe(200);
    expect(response.headers.get("content-type")).toContain("text/csv");
    expect(response.headers.get("content-disposition")).toContain("general-ledger.csv");
    // text() would strip the BOM: compare the raw bytes
    expect(new TextDecoder("utf-8", { ignoreBOM: true }).decode(await response.arrayBuffer())).toBe(csv);
    expect(String(fetchMock.mock.calls[0][0])).toContain("/report/v1/export");
    expect(sentBody(fetchMock)).toMatchObject({ menu_id: "general-ledger", format: "csv" });
    // no format given: still csv (the only export format the owner approved)
    await post("export", { menu_id: "general-ledger" });
    expect(sentBody(fetchMock, 1)).toMatchObject({ format: "csv" });
  });

  it("export: refuses any other format and relays the backend error as JSON", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: false, errorcode: "REPORT_EXPORT_TOO_LARGE", code: "REPORT_EXPORT_TOO_LARGE", message: "ข้อมูลมากเกินไป" }, { status: 422 }));
    vi.stubGlobal("fetch", fetchMock);
    expect((await post("export", { menu_id: "general-ledger", format: "xlsx" })).status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
    const response = await post("export", { menu_id: "general-ledger", format: "csv" });
    expect(response.status).toBe(200);
    expect(await response.json()).toMatchObject({ success: false, errorcode: "REPORT_EXPORT_TOO_LARGE" });
  });
});
