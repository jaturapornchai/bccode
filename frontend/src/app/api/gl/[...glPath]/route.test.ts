import { afterEach, describe, expect, it, vi } from "vitest";
import { GET, POST } from "./route";
const headers = { Authorization: "Bearer test-token", "Content-Type": "application/json", "x-bc-backend-url": "http://localhost:8888" };
const context = (...glPath: string[]) => ({ params: Promise.resolve({ glPath }) });
describe("GL authenticated proxy", () => {
  afterEach(() => vi.unstubAllGlobals());
  it("forwards only supported report filters and snapshot", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    const response = await GET(new Request("http://localhost/api/gl/reports/ledger?fiscalyear=FY&snapshot=5&companywide=true&holdingcode=other&backendUrl=http://localhost:8888", { headers }), context("reports", "ledger"));
    expect(response.status).toBe(200);
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toContain("/gl/v2/reports/ledger?fiscalyear=FY&snapshot=5&companywide=true");
    expect(url).not.toContain("holdingcode");
    expect(init.headers.Authorization).toBe("Bearer test-token");
  });
  it("rejects unknown/traversal paths and unauthenticated reads", async () => {
    vi.stubGlobal("fetch", vi.fn());
    expect((await GET(new Request("http://localhost/api/gl/accounts"), context("accounts"))).status).toBe(401);
    expect((await GET(new Request("http://localhost/api/gl/secrets", { headers }), context("secrets"))).status).toBe(404);
    expect((await GET(new Request("http://localhost/api/gl/accounts/..", { headers }), context("accounts", ".."))).status).toBe(404);
    expect(fetch).not.toHaveBeenCalled();
  });
  it("removes caller scope and preserves decimal strings in commands", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: { id: "1", version: 1 } })); vi.stubGlobal("fetch", fetchMock);
    const body = { resource: "journals", action: "create", requestid: "12345678-1234-1234-1234-123456789012", holdingcode: "other", journal: { holdingcode: "other", businesscode: "other", createdby: "other", lines: [{ debit: "0.10", credit: "0" }] } };
    expect((await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"))).status).toBe(200);
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ resource: "journals", action: "create", requestid: body.requestid, journal: { lines: [{ debit: "0.10", credit: "0" }] } });
  });
  it("rejects numeric money before forwarding", async () => {
    vi.stubGlobal("fetch", vi.fn());
    const response = await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify({ resource: "journals", action: "create", requestid: "12345678-1234-1234-1234-123456789012", journal: { lines: [{ debit: 0.1, credit: "0" }] } }) }), context("command"));
    expect(response.status).toBe(400); expect(fetch).not.toHaveBeenCalled();
  });
  it("accepts Thai codes with vowel signs and tone marks as path segments", async () => {
    // สระ/วรรณยุกต์ไทยเป็น combining mark (\p{M}) ไม่ใช่ \p{L} — regex เดิมตอบ 404 กับรหัส "สมุดซื้อ"
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    expect((await GET(new Request("http://localhost/api/gl/journal-books/สมุดซื้อ", { headers }), context("journal-books", "สมุดซื้อ"))).status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
  it("forwards journal review reads but rejects missing IDs and nested paths", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    expect((await GET(new Request("http://localhost/api/gl/journal-reviews/j-1?holdingcode=other", { headers }), context("journal-reviews", "j-1"))).status).toBe(200);
    expect(fetchMock.mock.calls[0][0]).toContain("/gl/v2/journal-reviews/j-1?");
    expect(fetchMock.mock.calls[0][0]).not.toContain("holdingcode");
    expect((await GET(new Request("http://localhost/api/gl/journal-reviews", { headers }), context("journal-reviews"))).status).toBe(404);
    expect((await GET(new Request("http://localhost/api/gl/journal-reviews/j-1/events", { headers }), context("journal-reviews", "j-1", "events"))).status).toBe(404);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
  it("preserves review concurrency fields and strips caller company scope", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    const body = { resource: "journals", action: "review", id: "j-1", requestid: "12345678-1234-1234-1234-123456789012", version: 4,
      holdingcode: "other", reviewedby: "forged", review: { status: 2, note: "ยอดไม่ตรง", expectedEventNo: 7, companycode: "other", createdby: "forged" } };
    expect((await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"))).status).toBe(200);
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ resource: "journals", action: "review", id: "j-1", requestid: body.requestid, version: 4,
      review: { status: 2, note: "ยอดไม่ตรง", expectedEventNo: 7 } });
    const response = await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify({ ...body, resource: "accounts" }) }), context("command"));
    expect(response.status).toBe(400); expect(fetchMock).toHaveBeenCalledTimes(1);
  });
  it("pages journal support and preserves as-of while discarding caller scope", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: { items: [], total: 0 } })); vi.stubGlobal("fetch", fetchMock);
    const response = await GET(new Request("http://localhost/api/gl/journal-support?kind=documents&q=INV-01&page=2&limit=25&asof=2026-06-30&companycode=OTHER&holdingcode=OTHER", { headers }), context("journal-support"));
    expect(response.status).toBe(200);
    const query = new URL(fetchMock.mock.calls[0][0]).searchParams;
    expect(Object.fromEntries(query)).toEqual({ kind: "documents", q: "INV-01", page: "2", limit: "25", asof: "2026-06-30" });
    expect((await GET(new Request("http://localhost/api/gl/journal-support/id", { headers }), context("journal-support", "id"))).status).toBe(404);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
  it.each(["ar-outstanding", "ap-outstanding", "bank-unmatched"])("allows the %s evidence report", async (report) => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    expect((await GET(new Request(`http://localhost/api/gl/reports/${report}?to=2026-06-30`, { headers }), context("reports", report))).status).toBe(200);
    expect(fetchMock.mock.calls[0][0]).toContain(`/gl/v2/reports/${report}?to=2026-06-30`);
  });
  it("preserves stable source identity and exact nested evidence amounts", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    const journal = { source_type: 2, source_system: "accounting-import", source_record_id: "INV-2569-01", details: { documents: [{ id: "D1", amount: "90071992547409.91" }], statement_lines: [{ id: "S1", amount: "0.30", balance_after: "90071992547410.21" }] } };
    const body = { resource: "journals", action: "create", requestid: "12345678-1234-1234-1234-123456789012", journal };
    expect((await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"))).status).toBe(200);
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual(body);
  });
  it.each(["documents", "allocations", "settlements", "statement_lines", "matches"])("rejects numeric amounts nested under %s", async (kind) => {
    vi.stubGlobal("fetch", vi.fn());
    const body = { resource: "journals", action: "reconcile", requestid: "12345678-1234-1234-1234-123456789012", journal: { details: { [kind]: [{ amount: 1 }] } } };
    expect((await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"))).status).toBe(400);
    expect(fetch).not.toHaveBeenCalled();
  });
  it("rejects numeric statement balance and limits reconcile to journals", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    const body = { resource: "journals", action: "reconcile", id: "J1", version: 3, reason: "statement matching", requestid: "12345678-1234-1234-1234-123456789012", journal: { details: { statement_lines: [{ amount: "0.30", balance_after: "10.30" }] } } };
    expect((await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"))).status).toBe(200);
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual(body);
    expect((await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify({ ...body, resource: "accounts" }) }), context("command"))).status).toBe(400);
    expect((await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify({ ...body, journal: { details: { statement_lines: [{ amount: "0.30", balance_after: 10 }] } } }) }), context("command"))).status).toBe(400);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
  it("forwards monthly budget commands, spread and the budgetcode report filter", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    const periods = Array.from({ length: 12 }, () => "8333.33");
    const budget = { code: "BG-2569", name: "งบค่าเช่าหน้าร้านและคลังสินค้า ปี 2569", fiscalyear: "2569", branchcode: "00000", status: "open", lines: [{ accountcode: "53110", periods }] };
    const body = { resource: "budgets", action: "create", requestid: "12345678-1234-1234-1234-123456789012", budget: { ...budget, holdingcode: "other", createdby: "forged" } };
    expect((await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"))).status).toBe(200);
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ resource: "budgets", action: "create", requestid: body.requestid, budget });
    const spread = { resource: "budgets", action: "spread", requestid: body.requestid, budget: { lines: [{ accountcode: "53110", total: "100000" }] } };
    expect((await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(spread) }), context("command"))).status).toBe(200);
    expect((await GET(new Request("http://localhost/api/gl/reports/budgetcomparison?fiscalyear=2569&budgetcode=BG-2569&holdingcode=other", { headers }), context("reports", "budgetcomparison"))).status).toBe(200);
    expect(fetchMock.mock.calls[2][0]).toContain("/gl/v2/reports/budgetcomparison?fiscalyear=2569&budgetcode=BG-2569");
    expect(fetchMock.mock.calls[2][0]).not.toContain("holdingcode");
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });
  it("rejects numeric budget money and spread outside budgets", async () => {
    vi.stubGlobal("fetch", vi.fn());
    const requestid = "12345678-1234-1234-1234-123456789012";
    const post = (body: unknown) => POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"));
    expect((await post({ resource: "budgets", action: "create", requestid, budget: { lines: [{ accountcode: "53110", periods: [100, "0"] }] } })).status).toBe(400);
    expect((await post({ resource: "budgets", action: "create", requestid, budget: { lines: [{ accountcode: "53110", periods: "100" }] } })).status).toBe(400);
    expect((await post({ resource: "budgets", action: "spread", requestid, budget: { lines: [{ accountcode: "53110", total: 100000 }] } })).status).toBe(400);
    expect((await post({ resource: "journals", action: "spread", requestid, budget: { lines: [{ accountcode: "53110", total: "100000" }] } })).status).toBe(400);
    expect(fetch).not.toHaveBeenCalled();
  });
});
