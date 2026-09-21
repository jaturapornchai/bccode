import { afterEach, describe, expect, it, vi } from "vitest";
import { GET, POST } from "./route";
const headers = { Authorization: "Bearer session", "x-bc-backend-url": "http://localhost:8888", "Content-Type": "application/json" };
const context = (...tokenPath: string[]) => ({ params: Promise.resolve({ tokenPath }) });
const request = (body: unknown) => new Request("http://localhost/api/mcp-tokens", { method: "POST", headers, body: JSON.stringify(body) });

describe("MCP token management proxy", () => {
  afterEach(() => vi.unstubAllGlobals());
  it("requires a session and disallows arbitrary routes", async () => {
    const fetchMock = vi.fn(); vi.stubGlobal("fetch", fetchMock);
    expect((await GET(new Request("http://localhost/api/mcp-tokens"), context())).status).toBe(401);
    expect((await POST(request({}), context("..", "revoke"))).status).toBe(404);
    expect(fetchMock).not.toHaveBeenCalled();
  });
  it("preserves administrator rejection and prevents caching", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(Response.json({ success: false, message: "Forbidden" }, { status: 403 })));
    const response = await GET(new Request("http://localhost/api/mcp-tokens?company=other", { headers }), context());
    expect(response.status).toBe(403); expect(response.headers.get("cache-control")).toContain("no-store");
    expect(vi.mocked(fetch).mock.calls[0][0]).toMatch(/\/mcp-tokens$/);
  });
  it("lists allowed active companies within the authenticated Holding", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: [{ code: "C01", name: "Company" }] })); vi.stubGlobal("fetch", fetchMock);
    const response = await GET(new Request("http://localhost/api/mcp-tokens/companies?holding=forged", { headers }), context("companies"));
    expect(response.status).toBe(200);
    expect(fetchMock.mock.calls[0][0]).toMatch(/\/mcp-tokens\/companies$/);
  });
  it("rejects missing, empty, wildcard and duplicate company allowlists", async () => {
    const fetchMock = vi.fn(); vi.stubGlobal("fetch", fetchMock);
    const base = { name: "Agent", kind: "mcp", mode: "readonly", expiresAt: "2026-10-20T00:00:00Z" };
    for (const companyCodes of [undefined, [], ["*"], [""], ["C01", "C01"], [42]]) {
      expect((await POST(request({ ...base, companyCodes }), context())).status).toBe(400);
    }
    expect(fetchMock).not.toHaveBeenCalled();
  });
  it("forwards only allowed creation fields and does not cache returned secrets", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true, data: { token: "test-secret" } })); vi.stubGlobal("fetch", fetchMock);
    const response = await POST(request({ name: " Agent ", kind: "mcp", mode: "readonly", companyCodes: ["C01", "C02"], expiresAt: "2026-10-20T00:00:00Z", companyCode: "forged", createdBy: "forged", token: "forged" }), context());
    expect(response.status).toBe(200); expect(response.headers.get("cache-control")).toContain("no-store");
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ name: "Agent", kind: "mcp", mode: "readonly", companyCodes: ["C01", "C02"], expiresAt: "2026-10-20T00:00:00Z" });
  });
  it("rejects invalid modes, names and expiry before forwarding", async () => {
    const fetchMock = vi.fn(); vi.stubGlobal("fetch", fetchMock);
    for (const body of [{ name: "Agent", mode: "admin", expiresAt: "2026-10-20" }, { name: "", kind: "mcp", mode: "readonly", expiresAt: "2026-10-20" }, { name: "Agent", kind: "api", mode: "readwrite", expiresAt: "invalid" }]) expect((await POST(request(body), context())).status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });
  it("revokes only the specified token ID with no caller-controlled scope", async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: true })); vi.stubGlobal("fetch", fetchMock);
    const id = "12345678123412341234123456789012";
    expect((await POST(request({ companyCode: "other" }), context(id, "revoke"))).status).toBe(200);
    expect(fetchMock.mock.calls[0][0]).toMatch(new RegExp(`/mcp-tokens/${id}/revoke$`)); expect(fetchMock.mock.calls[0][1].body).toBe("{}");
  });
});
