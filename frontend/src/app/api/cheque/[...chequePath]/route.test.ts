import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET, POST } from "./route";

const headers = { Authorization: "Bearer access-token", "x-bc-backend-url": "http://localhost:8888/goapi" };

const get = (path: string) =>
  GET(new Request(`http://localhost/api/cheque/${path}`, { headers }), { params: Promise.resolve({ chequePath: path.split("?")[0].split("/") }) });

const post = (path: string, body: unknown, init: { raw?: string; auth?: boolean } = {}) =>
  POST(
    new Request(`http://localhost/api/cheque/${path}`, {
      method: "POST",
      headers: init.auth === false ? {} : headers,
      body: init.raw ?? JSON.stringify(body),
    }),
    { params: Promise.resolve({ chequePath: path.split("/") }) },
  );

describe("cheque BFF", () => {
  beforeEach(() => {
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
  });
  afterEach(() => {
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    vi.unstubAllGlobals();
  });

  it("forwards code lists with only the allowed query parameters", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/cheque/v1/codes/bank_branch?q=silom&bank_code=BBL&active=1");
      expect((init?.headers as Record<string, string>).Authorization).toBe("Bearer access-token");
      return Response.json({ success: true, items: [], total: 0, page: 1, limit: 50 });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await get("codes/bank_branch?q=silom&bank_code=BBL&active=1&holdingcode=OTHER&company=99");

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("forwards the bank contact list", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request) => {
      expect(String(url)).toBe("http://localhost:8888/cheque/v1/bank-contacts?bank_code=BBL&bank_branch_code=0123");
      return Response.json({ success: true, items: [], total: 0 });
    });
    vi.stubGlobal("fetch", fetchMock);

    expect((await get("bank-contacts?bank_code=BBL&bank_branch_code=0123")).status).toBe(200);
  });

  it.each(["codes/branch", "codes", "codes/bank/extra", "cheques", "command", "codes/..", "bank-contacts/1"])(
    "rejects GET %s without calling the backend",
    async (path) => {
      const fetchMock = vi.fn();
      vi.stubGlobal("fetch", fetchMock);

      const response = await get(path);

      expect(response.status).toBe(404);
      expect(await response.json()).toMatchObject({ success: false, code: "cheque_err_resource_invalid" });
      expect(fetchMock).not.toHaveBeenCalled();
    },
  );

  it("rejects requests without a bearer token", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    expect((await post("command", { resource: "codes" }, { auth: false })).status).toBe(401);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("forwards a code command body unchanged to /cheque/v1/command", async () => {
    const command = { resource: "codes", action: "create", kind: "bank", code: "BBL", name: "ธนาคารกรุงเทพ" };
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/cheque/v1/command");
      expect(init?.method).toBe("POST");
      expect(JSON.parse(String(init?.body))).toEqual(command);
      return Response.json({ success: true, data: { code: "BBL" } });
    });
    vi.stubGlobal("fetch", fetchMock);

    expect((await post("command", command)).status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("relays a user error (409 with a code) as HTTP 200 + success:false so the console stays clean", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Response.json({ success: false, code: "bank_code_err_duplicate", message: "x", statuscode: 409 }, { status: 409 })));

    const response = await post("command", { resource: "codes", action: "create" });

    expect(response.status).toBe(200);
    expect(await response.json()).toMatchObject({ success: false, code: "bank_code_err_duplicate", statuscode: 409 });
  });

  it("keeps 403 as a real status", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Response.json({ success: false, code: "cheque_err_action_forbidden", statuscode: 403 }, { status: 403 })));

    expect((await post("command", { resource: "codes", action: "delete" })).status).toBe(403);
  });

  it.each([
    ["not json", { raw: "{" }, "cheque_err_invalid_payload"],
    ["an array body", { raw: "[]" }, "cheque_err_resource_invalid"],
    ["an unknown resource", { raw: JSON.stringify({ resource: "cheques" }) }, "cheque_err_resource_invalid"],
    ["a missing resource", { raw: JSON.stringify({ action: "create" }) }, "cheque_err_resource_invalid"],
  ])("rejects a command with %s", async (_name, init, code) => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await post("command", null, init);

    expect(response.status).toBe(400);
    expect(await response.json()).toMatchObject({ success: false, code, message: code });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects POST to any path other than command", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    expect((await post("codes/bank", { resource: "codes" })).status).toBe(404);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
