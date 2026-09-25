import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { GET, POST } from "./route";
import { POST as mcpPOST } from "@/app/mcp/gl/route";
beforeEach(() => vi.stubEnv("BCAI_LOCAL_BACKEND_URL", "http://mainapi:8888"));
afterEach(() => { vi.unstubAllGlobals(); vi.unstubAllEnvs(); });
const context = (path: string[]) => ({ params: Promise.resolve({ path }) });
it("API and MCP credentials cannot be used interchangeably", async () => {
  const fetcher = vi.fn(); vi.stubGlobal("fetch", fetcher);
  expect((await GET(new Request("http://local/api/integration/gl/accounts", { headers: { authorization: "Bearer bcaimcp_test.id.secret" } }), context(["accounts"]))).status).toBe(401);
  expect((await mcpPOST(new Request("http://local/mcp/gl", { method: "POST", headers: { authorization: "Bearer bcaiapi_test.id.secret" }, body: "{}" }))).status).toBe(401);
  expect(fetcher).not.toHaveBeenCalled();
});
it("API forwards to fixed GL endpoint and strips tenant/backend overrides", async () => {
  const fetcher = vi.fn().mockResolvedValue(Response.json({ success: true, data: [] })); vi.stubGlobal("fetch", fetcher);
  const result = await GET(new Request("http://local/api/integration/gl/accounts?limit=20&company=OTHER&backendUrl=https://evil.example", { headers: { authorization: "Bearer bcaiapi_test.id.secret", "x-bc-company-codes": "C01,C02" } }), context(["accounts"]));
  expect(result.status).toBe(200);
  expect(fetcher.mock.calls[0][0]).toBe("http://mainapi:8888/integration/gl/v2/accounts?limit=20");
  expect(fetcher.mock.calls[0][1].headers.get("x-bc-company-codes")).toBe("C01,C02");
});
it("only the command write endpoint is exposed", async () => {
  const fetcher = vi.fn().mockResolvedValue(Response.json({ success: true })); vi.stubGlobal("fetch", fetcher);
  const req = () => new Request("http://local/api/integration/gl/command", { method: "POST", headers: { authorization: "Bearer bcaiapi_test.id.secret", "content-type": "application/json" }, body: '{"amount":"0.30"}' });
  expect((await POST(req(), context(["tokens", "revoke"]))).status).toBe(404);
  expect((await POST(req(), context(["command"]))).status).toBe(200);
  expect(new TextDecoder().decode(fetcher.mock.calls[0][1].body)).toBe('{"amount":"0.30"}');
});

it("API token can page accounting evidence at an as-of date without overriding its tenant", async () => {
  const fetcher = vi.fn().mockResolvedValue(Response.json({ success: true, data: { items: [] } })); vi.stubGlobal("fetch", fetcher);
  const request = new Request("http://local/api/integration/gl/journal-support?kind=documents&q=INV-1&page=2&limit=30&asof=2026-06-30&companycode=OTHER", { headers: { authorization: "Bearer bcaiapi_test.id.secret", "x-bc-company-code": "C01" } });
  expect((await GET(request, context(["journal-support"]))).status).toBe(200);
  expect(fetcher.mock.calls[0][0]).toBe("http://mainapi:8888/integration/gl/v2/journal-support?kind=documents&q=INV-1&page=2&limit=30&asof=2026-06-30");
  expect(fetcher.mock.calls[0][1].headers.get("x-bc-company-code")).toBe("C01");
  expect((await GET(request, context(["journal-support","id"]))).status).toBe(404);
  expect(fetcher).toHaveBeenCalledTimes(1);
});
it("API token keeps the budgetcode filter on the budget comparison report (same allowlist as /api/gl)", async () => {
  const fetcher = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetcher);
  const response = await GET(new Request("http://local/api/integration/gl/reports/budgetcomparison?fiscalyear=2569&budgetcode=BG-2569&holdingcode=other", { headers: { authorization: "Bearer bcaiapi_test.id.secret" } }), context(["reports", "budgetcomparison"]));
  expect(response.status).toBe(200);
  expect(fetcher.mock.calls[0][0]).toBe("http://mainapi:8888/integration/gl/v2/reports/budgetcomparison?fiscalyear=2569&budgetcode=BG-2569");
});
it.each(["ar-outstanding","ap-outstanding","bank-unmatched"])("API token exposes %s through the authenticated report route", async (report) => {
  const fetcher = vi.fn().mockResolvedValue(Response.json({ success: true, data: {} })); vi.stubGlobal("fetch", fetcher);
  const response = await GET(new Request(`http://local/api/integration/gl/reports/${report}?to=2026-06-30&companywide=true`, { headers: { authorization: "Bearer bcaiapi_test.id.secret" } }), context(["reports",report]));
  expect(response.status).toBe(200);
  expect(fetcher.mock.calls[0][0]).toContain(`/integration/gl/v2/reports/${report}?to=2026-06-30&companywide=true`);
});
