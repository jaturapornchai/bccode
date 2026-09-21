import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET, POST } from "./route";

describe("MCP public bearer proxy", () => {
  beforeEach(() => { vi.stubEnv("BCAI_LOCAL_BACKEND_URL", "http://mainapi:8888"); });
  afterEach(() => { vi.unstubAllEnvs(); vi.unstubAllGlobals(); });
  const request = (extra: Record<string, string> = {}, body = "{}") => new Request("https://account.bcaicloud.com/mcp/gl?backendUrl=https://evil.example", {
    method: "POST", headers: { authorization: "Bearer bcaimcp_test.id.secret", "content-type": "application/json", accept: "application/json, text/event-stream", ...extra }, body,
  });
  it("rejects browser/session credentials and URL tokens", async () => {
    const fetcher = vi.fn(); vi.stubGlobal("fetch", fetcher);
    expect((await POST(request({ authorization: "Bearer session-token", cookie: "session=secret" }))).status).toBe(401);
    expect((await GET(new Request("https://account.bcaicloud.com/mcp/gl?access_token=bcaimcp_test.id.secret"))).status).toBe(401);
    expect(fetcher).not.toHaveBeenCalled();
  });
  it("forwards only MCP headers to the fixed server, without cookies", async () => {
    const fetcher = vi.fn().mockResolvedValue(Response.json({ jsonrpc: "2.0", id: 1, result: {} })); vi.stubGlobal("fetch", fetcher);
    const response = await POST(request({ cookie: "private=session", "x-bc-backend-url": "https://evil.example", "mcp-protocol-version": "2025-11-25" }));
    expect(response.status).toBe(200);
    const [url, init] = fetcher.mock.calls[0];
    expect(url).toBe("http://mainapi:8888/mcp/gl");
    expect(init.headers.get("cookie")).toBeNull();
    expect(init.headers.get("x-bc-backend-url")).toBeNull();
    expect(init.headers.get("mcp-protocol-version")).toBe("2025-11-25");
    expect(init.redirect).toBe("error");
    expect(response.headers.get("cache-control")).toBe("no-store");
  });
  it("preserves bearer challenges and notification empty responses", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 401, headers: { "www-authenticate": "Bearer" } })));
    expect((await POST(request())).headers.get("www-authenticate")).toBe("Bearer");
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 202 })));
    expect(await (await POST(request())).text()).toBe("");
  });
  it("rejects oversize bodies without forwarding and hides upstream errors", async () => {
    const fetcher = vi.fn(); vi.stubGlobal("fetch", fetcher);
    expect((await POST(request({}, "x".repeat(2 * 1024 * 1024 + 1)))).status).toBe(413);
    expect(fetcher).not.toHaveBeenCalled();
    fetcher.mockRejectedValue(new Error("private credentials"));
    const response = await POST(request());
    expect(response.status).toBe(503); expect(await response.text()).not.toContain("private");
  });
});
