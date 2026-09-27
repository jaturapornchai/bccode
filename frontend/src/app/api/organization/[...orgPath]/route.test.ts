import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET, POST, PUT } from "./route";

const context = (path: string) => ({ params: Promise.resolve({ orgPath: path.split("/") }) });
const authHeaders = { Authorization: "Bearer access-token", "Accept-Language": "en" };

function request(method: string, path: string, body?: unknown, headers: Record<string, string> = authHeaders) {
  return new Request(`http://localhost/api/organization/${path}`, {
    method,
    headers: { "Content-Type": "application/json", ...headers },
    body: body === undefined ? undefined : typeof body === "string" ? body : JSON.stringify(body),
  });
}

describe("organization BFF (company/branch management)", () => {
  beforeEach(() => {
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
  });
  afterEach(() => {
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    vi.unstubAllGlobals();
  });

  it("lists companies and branches as the management structure, forwarding Authorization and Accept-Language", async () => {
    const calls: Array<{ url: string; init?: RequestInit }> = [];
    vi.stubGlobal("fetch", vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      calls.push({ url: String(url), init });
      return Response.json({ success: true, data: [{ guidfixed: "01" }] });
    }));

    const company = await GET(request("GET", "company?management=false&holdingcode=other"), context("company"));
    const branch = await GET(request("GET", "branch"), context("branch"));

    expect(company.status).toBe(200);
    expect(await company.json()).toEqual({ success: true, data: [{ guidfixed: "01" }] });
    expect(branch.status).toBe(200);
    // Client query parameters are not forwarded; the target is fixed.
    expect(calls.map((call) => call.url)).toEqual([
      "http://localhost:8888/organization/company?management=true",
      "http://localhost:8888/organization/branch?management=true",
    ]);
    for (const call of calls) {
      const headers = call.init?.headers as Record<string, string>;
      expect(call.init?.method).toBe("GET");
      expect(headers.Authorization).toBe("Bearer access-token");
      expect(headers["Accept-Language"]).toBe("en");
    }
  });

  it("creates a company and updates a company or branch by code, stripping backendUrl from the body", async () => {
    const calls: Array<{ url: string; init?: RequestInit }> = [];
    vi.stubGlobal("fetch", vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      calls.push({ url: String(url), init });
      return Response.json({ success: true });
    }));

    await POST(request("POST", "company", { code: "02", backendUrl: "http://localhost:3000/backend/goapi" }), context("company"));
    await PUT(request("PUT", "company/01", { code: "01", names: [] }), context("company/01"));
    await PUT(request("PUT", "branch/00000", { code: "00000" }), context("branch/00000"));
    await PUT(request("PUT", "company/บริษัท01", { code: "บริษัท01" }), context("company/บริษัท01"));

    expect(calls.map((call) => [call.init?.method, call.url])).toEqual([
      ["POST", "http://localhost:8888/organization/company"],
      ["PUT", "http://localhost:8888/organization/company/01"],
      ["PUT", "http://localhost:8888/organization/branch/00000"],
      ["PUT", `http://localhost:8888/organization/company/${encodeURIComponent("บริษัท01")}`],
    ]);
    expect(JSON.parse(String(calls[0].init?.body))).toEqual({ code: "02" });
    expect((calls[1].init?.headers as Record<string, string>).Authorization).toBe("Bearer access-token");
  });

  it("sends every code a company can be created with in Go's canonical path escaping", async () => {
    // Codes are free text (backend NormalizeCompanyCode only trims + uppercases). Echo does not unescape :id,
    // so the path must be exactly what Go's net/url would produce: & + , ; = @ : $ stay raw, the rest is %XX.
    const calls: string[] = [];
    vi.stubGlobal("fetch", vi.fn(async (url: string | URL | Request) => {
      calls.push(String(url));
      return Response.json({ success: true });
    }));
    const put = (id: string) =>
      PUT(request("PUT", "company/x", { code: id }), { params: Promise.resolve({ orgPath: ["company", id] }) });

    for (const id of ["A&B", "K+1(ก)!", "a%2Fb", "a b", "X,Y;Z=1@2:$"]) {
      expect((await put(id)).status, id).toBe(200);
    }

    expect(calls).toEqual([
      "http://localhost:8888/organization/company/A&B",
      "http://localhost:8888/organization/company/K+1%28%E0%B8%81%29%21",
      "http://localhost:8888/organization/company/a%252Fb",
      "http://localhost:8888/organization/company/a%20b",
      "http://localhost:8888/organization/company/X,Y;Z=1@2:$",
    ]);
  });

  it("passes the mainapi status and JSON through", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Response.json({ success: false, code: "FORBIDDEN", message: "forbidden" }, { status: 403 })));

    const response = await PUT(request("PUT", "branch/00000", { code: "00000" }), context("branch/00000"));

    expect(response.status).toBe(403);
    expect(await response.json()).toEqual({ success: false, code: "FORBIDDEN", message: "forbidden" });
  });

  it("rejects paths and methods outside the allowlist without calling mainapi", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const responses = await Promise.all([
      GET(request("GET", "department"), context("department")),
      GET(request("GET", "company/01"), context("company/01")),
      GET(request("GET", "branch/list"), context("branch/list")),
      POST(request("POST", "branch", { code: "00001" }), context("branch")),
      POST(request("POST", "company/01", { code: "01" }), context("company/01")),
      PUT(request("PUT", "company", { code: "01" }), context("company")),
      PUT(request("PUT", "company/01/extra", { code: "01" }), context("company/01/extra")),
      PUT(request("PUT", "holding/01", { code: "01" }), context("holding/01")),
    ]);

    expect(responses.map((response) => response.status)).toEqual([404, 404, 404, 404, 404, 404, 404, 404]);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects an unsafe id and a non-object body", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    for (const id of ["..", ".", "", "01/../x", "..\\x", "a\u0000b", "x".repeat(129)]) {
      const response = await PUT(request("PUT", "company/x", { code: "x" }), { params: Promise.resolve({ orgPath: ["company", id] }) });
      expect(response.status, id).toBe(400);
    }
    expect((await PUT(request("PUT", "branch/00000", "[1]"), context("branch/00000"))).status).toBe(400);
    expect((await PUT(request("PUT", "branch/00000", "not json"), context("branch/00000"))).status).toBe(400);
    expect((await POST(request("POST", "company", "[]"), context("company"))).status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("requires a Bearer token", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(request("GET", "company", undefined, {}), context("company"));

    expect(response.status).toBe(401);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("never fetches a host named by the client", async () => {
    const urls: string[] = [];
    vi.stubGlobal("fetch", vi.fn(async (url: string | URL | Request) => {
      urls.push(String(url));
      return Response.json({ success: true });
    }));

    await GET(request("GET", "company", undefined, { ...authHeaders, "x-bc-backend-url": "https://evil.example/goapi" }), context("company"));
    await POST(request("POST", "company", { code: "02", backendUrl: "https://evil.example/goapi" }), context("company"));
    const garbage = await GET(request("GET", "company", undefined, { ...authHeaders, "x-bc-backend-url": "ftp://evil.example" }), context("company"));

    expect(urls).toEqual(["http://localhost:8888/organization/company?management=true", "http://localhost:8888/organization/company"]);
    expect(garbage.status).toBe(400);
  });
});
