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
  it("forwards the VAT calculator with its decimal-string inputs and drops caller scope", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: { base_amount: "100.00", vat_amount: "7.00", total_amount: "107.00" } })); vi.stubGlobal("fetch", fetchMock);
    const response = await GET(new Request("http://localhost/api/gl/vat-calc?mode=inclusive&amount=107.00&rate=7&fiscalyear=FY2569&holdingcode=other", { headers }), context("vat-calc"));
    expect(response.status).toBe(200);
    const url = new URL(fetchMock.mock.calls[0][0]);
    expect(url.pathname).toBe("/gl/v2/vat-calc");
    expect(Object.fromEntries(url.searchParams)).toEqual({ mode: "inclusive", amount: "107.00", rate: "7", fiscalyear: "FY2569" });
    expect((await GET(new Request("http://localhost/api/gl/vat-calc/x", { headers }), context("vat-calc", "x"))).status).toBe(404);
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
  it("forwards statement notes reads and commands with multi-line Thai text intact", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    expect((await GET(new Request("http://localhost/api/gl/statement-notes?q=2569&page=1&limit=100&holdingcode=other", { headers }), context("statement-notes"))).status).toBe(200);
    expect(fetchMock.mock.calls[0][0]).toContain("/gl/v2/statement-notes?q=2569&page=1&limit=100");
    expect(fetchMock.mock.calls[0][0]).not.toContain("holdingcode");
    const master = { code: "2569", name: "", isactive: true, notes: [{ id: "n1", noteno: "1", title: "ข้อมูลทั่วไป", body: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด\nสำนักงานใหญ่" }] };
    const body = { resource: "statement-notes", action: "create", requestid: "12345678-1234-1234-1234-123456789012", master };
    expect((await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"))).status).toBe(200);
    expect(JSON.parse(fetchMock.mock.calls[1][1].body)).toEqual(body);
  });
  it("forwards the one-request statement set with its template list and notes flag, including an empty list", async () => {
    // ชุดงบการเงิน (backend statement_set.go): templates ว่าง = หมายเหตุอย่างเดียว ต้องส่งต่อเป็นค่าว่าง ไม่หายไป (ไม่ส่ง = ชุดเริ่มต้น)
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: { sections: [], notes: [] } })); vi.stubGlobal("fetch", fetchMock);
    expect((await GET(new Request("http://localhost/api/gl/reports/statement-set?fiscalyear=2569&from=2569-01-01&to=2569-12-31&templates=BS-1,PL-1&notes=false&holdingcode=other", { headers }), context("reports", "statement-set"))).status).toBe(200);
    const query = new URL(fetchMock.mock.calls[0][0]).searchParams;
    expect(new URL(fetchMock.mock.calls[0][0]).pathname).toBe("/gl/v2/reports/statement-set");
    expect(Object.fromEntries(query)).toEqual({ fiscalyear: "2569", from: "2569-01-01", to: "2569-12-31", templates: "BS-1,PL-1", notes: "false" });
    expect((await GET(new Request("http://localhost/api/gl/reports/statement-set?fiscalyear=2569&templates=&notes=true", { headers }), context("reports", "statement-set"))).status).toBe(200);
    const notesOnly = new URL(fetchMock.mock.calls[1][0]).searchParams;
    expect(notesOnly.has("templates")).toBe(true);
    expect(notesOnly.get("templates")).toBe("");
    expect((await GET(new Request("http://localhost/api/gl/reports/statement-set/extra", { headers }), context("reports", "statement-set", "extra"))).status).toBe(404);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
  it("forwards account suggestions for statement templates only (calculation, nothing saved)", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: { suggestions: { targets: [], fixes: [] } } })); vi.stubGlobal("fetch", fetchMock);
    const requestid = "12345678-1234-1234-1234-123456789012";
    const master = { code: "BS-DBD", name: "งบฐานะการเงิน", statementtype: "balance_sheet", isactive: true, rows: [{ id: "bs-30", rowno: 30, rowtype: "account", title: "เงินสดและรายการเทียบเท่าเงินสด", accountcodes: [], suggestkey: "cash_and_equivalents" }], holdingcode: "other" };
    const post = (body: unknown) => POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"));
    expect((await post({ resource: "statement-templates", action: "suggest", id: "tpl-1", requestid, master })).status).toBe(200);
    const { holdingcode: _scope, ...forwarded } = master;
    expect(_scope).toBe("other");
    expect(fetchMock.mock.calls[0][0]).toContain("/gl/v2/command");
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ resource: "statement-templates", action: "suggest", id: "tpl-1", requestid, master: forwarded });
    for (const resource of ["accounts", "journals", "budgets", "statement-notes"]) expect((await post({ resource, action: "suggest", requestid, master })).status).toBe(400);
    expect(fetchMock).toHaveBeenCalledTimes(1);
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
  it("forwards period generation and batch posting (check + create) only for their resources", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    const requestid = "12345678-1234-1234-1234-123456789012";
    const post = (body: unknown) => POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"));
    expect((await post({ resource: "periods", action: "generate", requestid, targetyear: "2569" })).status).toBe(200);
    const filter = { bookcode: "JV", datefrom: "2026-01-01", dateto: "2026-01-31", docnofrom: "JV001", docnoto: "JV099" };
    expect((await post({ resource: "post-batches", action: "check", requestid, postbatch: filter })).status).toBe(200);
    const batch = { items: [{ id: "j1", version: 3 }], remark: "ผ่านรายการประจำเดือนมกราคม 2569", holdingcode: "other" };
    expect((await post({ resource: "post-batches", action: "create", requestid, postbatch: batch })).status).toBe(200);
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ resource: "periods", action: "generate", requestid, targetyear: "2569" });
    expect(JSON.parse(fetchMock.mock.calls[1][1].body)).toEqual({ resource: "post-batches", action: "check", requestid, postbatch: filter });
    const { holdingcode: _scope, ...forwarded } = batch;
    expect(_scope).toBe("other");
    expect(JSON.parse(fetchMock.mock.calls[2][1].body)).toEqual({ resource: "post-batches", action: "create", requestid, postbatch: forwarded });
    for (const body of [{ resource: "journals", action: "generate", requestid }, { resource: "journals", action: "check", requestid }, { resource: "periods", action: "check", requestid }]) expect((await post(body)).status).toBe(400);
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });
  it("forwards the closing-entries preview only as processes/close-preview (calculation only)", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: { closepreview: { branches: [] } } })); vi.stubGlobal("fetch", fetchMock);
    const requestid = "12345678-1234-1234-1234-123456789012";
    const post = (body: unknown) => POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"));
    expect((await post({ resource: "processes", action: "close-preview", id: "2026", version: 3, date: "2026-02-28", docno: "CLOSE-FEB", holdingcode: "other", requestid })).status).toBe(200);
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ resource: "processes", action: "close-preview", id: "2026", version: 3, date: "2026-02-28", docno: "CLOSE-FEB", requestid });
    for (const resource of ["journals", "periods", "post-batches"]) expect((await post({ resource, action: "close-preview", requestid })).status).toBe(400);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
  it("forwards the department and project code range, ignoring unknown filters and caller scope", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    for (const resource of ["departments", "projects"]) {
      expect((await GET(new Request(`http://localhost/api/gl/${resource}?codefrom=D01&codeto=D09&page=1&limit=500&holdingcode=other&sql=x`, { headers }), context(resource))).status).toBe(200);
    }
    const urls = fetchMock.mock.calls.map((call) => new URL(call[0]));
    expect(urls.map((url) => url.pathname)).toEqual(["/gl/v2/departments", "/gl/v2/projects"]);
    for (const url of urls) expect(Object.fromEntries(url.searchParams)).toEqual({ codefrom: "D01", codeto: "D09", page: "1", limit: "500" });
    const requestid = "12345678-1234-1234-1234-123456789012";
    const master = { id: "d1", version: 2, code: "D01", name: "ฝ่ายบัญชีและการเงิน", isactive: true, holdingcode: "other" };
    expect((await POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify({ resource: "departments", action: "update", requestid, reason: "แก้ชื่อ", master }) }), context("command"))).status).toBe(200);
    const { holdingcode: _scope, ...forwarded } = master;
    expect(_scope).toBe("other");
    expect(JSON.parse(fetchMock.mock.calls[2][1].body)).toEqual({ resource: "departments", action: "update", requestid, reason: "แก้ชื่อ", master: forwarded });
  });
  it("forwards the purchase VAT adjustment list filters and both command modes, dropping caller scope", async () => {
    const fetchMock = vi.fn().mockImplementation(async () => Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    const filters = "docdatefrom=2026-01-01&docdateto=2026-01-31&invdatefrom=2026-01-01&invdateto=2026-01-31&partnerfrom=AP001&partnerto=AP999&page=2&limit=100";
    expect((await GET(new Request(`http://localhost/api/gl/vat-adjustments?${filters}&holdingcode=other`, { headers }), context("vat-adjustments"))).status).toBe(200);
    const url = new URL(fetchMock.mock.calls[0][0]);
    expect(url.pathname).toBe("/gl/v2/vat-adjustments");
    expect(url.searchParams.get("holdingcode")).toBeNull();
    expect(Object.fromEntries(url.searchParams)).toEqual(Object.fromEntries(new URLSearchParams(filters)));
    const requestid = "12345678-1234-1234-1234-123456789012";
    const post = (body: unknown) => POST(new Request("http://localhost/api/gl/command", { method: "POST", headers, body: JSON.stringify(body) }), context("command"));
    const rows = { mode: "rows", items: [{ journalid: "j1", version: 3, vatid: "v1", taxperiodyear: 2026, taxperiodmonth: 2, claimstatus: 1 }], holdingcode: "other" };
    expect((await post({ resource: "vat-adjustments", action: "update", requestid, reason: "ย้ายงวด", vatadjust: rows })).status).toBe(200);
    const bulk = { mode: "invoice_date", filter: { invdatefrom: "2026-01-01", invdateto: "2026-01-31" }, expectedcount: 12 };
    expect((await post({ resource: "vat-adjustments", action: "update", requestid, reason: "ตั้งงวด", vatadjust: bulk })).status).toBe(200);
    const { holdingcode: _scope, ...forwarded } = rows;
    expect(_scope).toBe("other");
    expect(JSON.parse(fetchMock.mock.calls[1][1].body)).toEqual({ resource: "vat-adjustments", action: "update", requestid, reason: "ย้ายงวด", vatadjust: forwarded });
    expect(JSON.parse(fetchMock.mock.calls[2][1].body)).toEqual({ resource: "vat-adjustments", action: "update", requestid, reason: "ตั้งงวด", vatadjust: bulk });
    expect((await post({ resource: "vat-adjustments", action: "generate", requestid })).status).toBe(400);
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });
});
