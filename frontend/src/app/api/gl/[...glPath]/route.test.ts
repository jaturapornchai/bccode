import { afterEach, describe, expect, it, vi } from "vitest";
import { GET, POST } from "./route";
const headers = { Authorization: "Bearer test-token", "Content-Type": "application/json", "x-bc-backend-url": "http://localhost:8888" };
const context = (...glPath: string[]) => ({ params: Promise.resolve({ glPath }) });
describe("GL authenticated proxy", () => {
  afterEach(() => vi.unstubAllGlobals());
  it("forwards only supported report filters and snapshot", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetchMock);
    const response = await GET(new Request("http://localhost/api/gl/reports/ledger?fiscalyear=FY&snapshot=5&holdingcode=other&backendUrl=http://localhost:8888", { headers }), context("reports", "ledger"));
    expect(response.status).toBe(200);
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toContain("/gl/v2/reports/ledger?fiscalyear=FY&snapshot=5");
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
});
