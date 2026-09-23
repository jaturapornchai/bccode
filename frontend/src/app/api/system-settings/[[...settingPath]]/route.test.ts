import { createHmac } from "crypto";
import { afterEach, describe, expect, it, vi } from "vitest";
import { DELETE, GET, POST, PUT } from "./route";

const SECRET = "test-secret";

describe("system settings API route security", () => {
  afterEach(() => {
    delete process.env.JWT_SECRET_KEY;
    vi.unstubAllGlobals();
  });

  it("delegates user login-account creation to the authorized backend save", async () => {
    process.env.JWT_SECRET_KEY = SECRET;
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void url;
      void init;
      return Response.json({ success: true, data: false });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/system-settings/user?holdingcode=SHOP001", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${signJwt({ username: "owner@example.com", holdingcode: "SHOP001" })}`,
        },
        body: JSON.stringify({
          backendUrl: "http://localhost:8888",
          holdingcode: "SHOP001",
          username: "new-user",
          name: "New User",
        }),
      }),
      { params: Promise.resolve({ settingPath: ["user"] }) },
    );

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(String(fetchMock.mock.calls[0][0])).toBe("http://localhost:8888/holding/permission?holdingcode=SHOP001");
    const saveBody = JSON.parse(String(fetchMock.mock.calls[0][1]?.body));
    expect(saveBody).toMatchObject({ username: "new-user", name: "New User" });
    expect(saveBody).not.toHaveProperty("password");
  });

  it("proxies product unit deletes to the legacy unit guid endpoint", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void url;
      void init;
      return Response.json({ success: true });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await DELETE(
      new Request("http://localhost/api/system-settings/productunit/UNIT-GUID?holdingcode=SHOP001", {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer test-token",
        },
        body: JSON.stringify({
          backendUrl: "http://localhost:8888",
          holdingcode: "SHOP001",
        }),
      }),
      { params: Promise.resolve({ settingPath: ["productunit", "UNIT-GUID"] }) },
    );

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [proxiedUrl, proxiedInit] = fetchMock.mock.calls[0] as [string | URL | Request, RequestInit | undefined];
    expect(String(proxiedUrl)).toBe("http://localhost:8888/unit/UNIT-GUID?holdingcode=SHOP001");
    expect(proxiedInit?.method).toBe("DELETE");
    expect(proxiedInit?.body).toBeUndefined();
  });

  it("keeps email user ids readable by the holding permission endpoint", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void url;
      void init;
      return Response.json({ success: true, data: { username: "demo.admin01@example.com" } });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/system-settings/user/demo.admin01%40example.com?holdingcode=SHOP001", {
        headers: {
          Authorization: "Bearer test-token",
          "x-bc-backend-url": "http://localhost:8888/goapi",
        },
      }),
      { params: Promise.resolve({ settingPath: ["user", "demo.admin01%40example.com"] }) },
    );

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [proxiedUrl, proxiedInit] = fetchMock.mock.calls[0] as [string | URL | Request, RequestInit | undefined];
    expect(String(proxiedUrl)).toBe("http://localhost:8888/holding/permission/demo.admin01@example.com?offset=0&limit=1000");
    expect(proxiedInit?.method).toBe("GET");
  });

  it.each([
    ["activelanguages", "GET"],
    ["company", "GET"],
    ["activelanguages", "PUT"],
    ["company", "PUT"],
  ])("supports company setup CRUD proxy for %s %s", async (slug, method) => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      if (method === "GET") {
        expect(String(url)).toBe("http://localhost:8888/shop/SHOP001");
        expect(init?.method).toBe("GET");
        return Response.json({ success: true, data: { holdingcode: "SHOP001" } });
      }

      expect(String(url)).toBe("http://localhost:8888/holding/SHOP001?holdingcode=SHOP001");
      expect(init?.method).toBe("PUT");
      expect(JSON.parse(String(init?.body))).toMatchObject({ holdingcode: "SHOP001" });
      return Response.json({ success: true });
    });
    vi.stubGlobal("fetch", fetchMock);

    const request = new Request(`http://localhost/api/system-settings/${slug}?holdingcode=SHOP001`, {
      method,
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer test-token",
        "x-bc-backend-url": "http://localhost:8888/goapi",
      },
      body: method === "GET" ? undefined : JSON.stringify({ holdingcode: "SHOP001", names: [{ code: "th", name: "บริษัททดสอบ" }] }),
    });
    const response =
      method === "GET"
        ? await GET(request, { params: Promise.resolve({ settingPath: [slug] }) })
        : await PUT(request, { params: Promise.resolve({ settingPath: [slug] }) });

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("supports user setup list, create, update, and delete proxy paths", async () => {
    const calls: Array<[string, RequestInit | undefined]> = [];
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      calls.push([String(url), init]);
      const requestUrl = String(url);
      return Response.json({ success: true, data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);

    const headers = {
      "Content-Type": "application/json",
      Authorization: "Bearer test-token",
      "x-bc-backend-url": "http://localhost:8888/goapi",
    };

    await GET(
      new Request("http://localhost/api/system-settings/user?holdingcode=SHOP001&limit=100&offset=0", { headers }),
      { params: Promise.resolve({ settingPath: ["user"] }) },
    );
    await POST(
      new Request("http://localhost/api/system-settings/user?holdingcode=SHOP001", {
        method: "POST",
        headers,
        body: JSON.stringify({ holdingcode: "SHOP001", username: "newuser@example.com", userprofilename: "New User" }),
      }),
      { params: Promise.resolve({ settingPath: ["user"] }) },
    );
    await PUT(
      new Request("http://localhost/api/system-settings/user/newuser@example.com?holdingcode=SHOP001", {
        method: "PUT",
        headers,
        body: JSON.stringify({ holdingcode: "SHOP001", username: "newuser@example.com", userprofilename: "Updated User" }),
      }),
      { params: Promise.resolve({ settingPath: ["user", "newuser@example.com"] }) },
    );
    await DELETE(
      new Request("http://localhost/api/system-settings/user/newuser@example.com?holdingcode=SHOP001", {
        method: "DELETE",
        headers,
        body: JSON.stringify({ holdingcode: "SHOP001" }),
      }),
      { params: Promise.resolve({ settingPath: ["user", "newuser@example.com"] }) },
    );

    expect(calls.map(([url]) => url)).toEqual([
      "http://localhost:8888/holding/users?offset=0&limit=100",
      "http://localhost:8888/holding/permission?holdingcode=SHOP001",
      "http://localhost:8888/holding/permission?holdingcode=SHOP001",
      "http://localhost:8888/holding/permission/newuser@example.com?holdingcode=SHOP001",
    ]);
    expect(calls[1][1]?.method).toBe("PUT");
    expect(calls[2][1]?.method).toBe("PUT");
    expect(calls[3][1]?.method).toBe("DELETE");
  });

  it("serves the screen permission catalog locally and keeps it read-only", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const headers = {
      "Content-Type": "application/json",
      Authorization: "Bearer test-token",
      "x-bc-backend-url": "http://localhost:8888/goapi",
    };

    const getResponse = await GET(
      new Request("http://localhost/api/system-settings/permissiondefinition?q=sale&limit=10&offset=0", { headers }),
      { params: Promise.resolve({ settingPath: ["permissiondefinition"] }) },
    );
    const getPayload = await getResponse.json() as { success: boolean; data: Array<Record<string, unknown>>; total: number };
    expect(getResponse.status).toBe(200);
    expect(getPayload.success).toBe(true);
    expect(getPayload.total).toBeGreaterThan(0);
    expect(getPayload.data.every((item) => typeof item.permissioncode === "string" && item.isactive === true)).toBe(true);

    const postResponse = await POST(
      new Request("http://localhost/api/system-settings/permissiondefinition", {
        method: "POST",
        headers,
        body: JSON.stringify({ permissioncode: "CUSTOM" }),
      }),
      { params: Promise.resolve({ settingPath: ["permissiondefinition"] }) },
    );
    expect(postResponse.status).toBe(405);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("proxies role permission CRUD to the dedicated organization API", async () => {
    const calls: Array<[string, RequestInit | undefined]> = [];
    vi.stubGlobal("fetch", vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      calls.push([String(url), init]);
      return Response.json({ success: true, data: [] });
    }));
    const headers = {
      "Content-Type": "application/json",
      Authorization: "Bearer test-token",
      "x-bc-backend-url": "http://localhost:8888/goapi",
    };
    const body = {
      rolecode: "USER",
      names: [{ code: "th", name: "ผู้ใช้งาน" }],
      permissions: ["sale-order"],
      isactive: true,
      __v: 0,
    };

    await GET(
      new Request("http://localhost/api/system-settings/permissiongroup?holdingcode=SHOP001&limit=100&offset=0", { headers }),
      { params: Promise.resolve({ settingPath: ["permissiongroup"] }) },
    );
    await POST(
      new Request("http://localhost/api/system-settings/permissiongroup?holdingcode=SHOP001", {
        method: "POST",
        headers,
        body: JSON.stringify(body),
      }),
      { params: Promise.resolve({ settingPath: ["permissiongroup"] }) },
    );
    await PUT(
      new Request("http://localhost/api/system-settings/permissiongroup/507f1f77bcf86cd799439011?holdingcode=SHOP001", {
        method: "PUT",
        headers,
        body: JSON.stringify(body),
      }),
      { params: Promise.resolve({ settingPath: ["permissiongroup", "507f1f77bcf86cd799439011"] }) },
    );
    await DELETE(
      new Request("http://localhost/api/system-settings/permissiongroup/507f1f77bcf86cd799439011?holdingcode=SHOP001", {
        method: "DELETE",
        headers,
      }),
      { params: Promise.resolve({ settingPath: ["permissiongroup", "507f1f77bcf86cd799439011"] }) },
    );

    expect(calls.map(([url]) => url)).toEqual([
      "http://localhost:8888/organization/role-permission?offset=0&limit=100",
      "http://localhost:8888/organization/role-permission?holdingcode=SHOP001",
      "http://localhost:8888/organization/role-permission/507f1f77bcf86cd799439011?holdingcode=SHOP001",
      "http://localhost:8888/organization/role-permission/507f1f77bcf86cd799439011?holdingcode=SHOP001",
    ]);
    expect(calls.map(([, init]) => init?.method)).toEqual(["GET", "POST", "PUT", "DELETE"]);
    expect(JSON.parse(String(calls[1][1]?.body))).toMatchObject(body);
    expect(JSON.parse(String(calls[2][1]?.body))).toMatchObject(body);
    expect(calls[3][1]?.body).toBeUndefined();
  });

  it("keeps user access audit read-only at the API proxy layer", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const headers = {
      "Content-Type": "application/json",
      Authorization: "Bearer test-token",
      "x-bc-backend-url": "http://localhost:8888/goapi",
    };

    const getResponse = await GET(
      new Request("http://localhost/api/system-settings/useraccessaudit?holdingcode=SHOP001", { headers }),
      { params: Promise.resolve({ settingPath: ["useraccessaudit"] }) },
    );
    const postResponse = await POST(
      new Request("http://localhost/api/system-settings/useraccessaudit?holdingcode=SHOP001", {
        method: "POST",
        headers,
        body: JSON.stringify({ holdingcode: "SHOP001" }),
      }),
      { params: Promise.resolve({ settingPath: ["useraccessaudit"] }) },
    );

    expect(getResponse.status).toBe(405);
    expect(postResponse.status).toBe(405);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("proxies book bank master CRUD operations to the payment/bookbank backend endpoints", async () => {
    const fetchMock = vi.fn(async () => Response.json({ success: true }));
    vi.stubGlobal("fetch", fetchMock);
    const headers = {
      "Content-Type": "application/json",
      Authorization: "Bearer test-token",
      "x-bc-backend-url": "http://localhost:8888",
    };

    // List
    const listRes = await GET(
      new Request("http://localhost/api/system-settings/bookbankscreen?offset=0&limit=100", { headers }),
      { params: Promise.resolve({ settingPath: ["bookbankscreen"] }) },
    );
    const calls = fetchMock.mock.calls as unknown as [string | URL | Request, RequestInit | undefined][];
    expect(listRes.status).toBe(200);
    expect(String(calls[0][0])).toBe("http://localhost:8888/payment/bookbank/list?offset=0&limit=100");

    // Create
    const createRes = await POST(
      new Request("http://localhost/api/system-settings/bookbankscreen", {
        method: "POST",
        headers,
        body: JSON.stringify({
          backendUrl: "http://localhost:8888",
          bookcode: "kbank-01",
          passbook: "123-4-56789-0",
          bankbranch: "Siam",
          accountname: "BC Corp",
          bankcode: "kbank",
          names: [{ code: "th", name: "กสิกรไทย สยาม" }],
          logo: "https://example.com/kbank.png",
        }),
      }),
      { params: Promise.resolve({ settingPath: ["bookbankscreen"] }) },
    );
    expect(createRes.status).toBe(200);
    expect(String(calls[1][0])).toBe("http://localhost:8888/payment/bookbank");
    const createBody = JSON.parse(String(calls[1][1]?.body));
    expect(createBody.bookcode).toBe("KBANK-01"); // Normalized uppercase businessCode
    expect(createBody.passbook).toBe("123-4-56789-0");
    expect(createBody.images).toEqual([{ xorder: 0, uri: "https://example.com/kbank.png" }]);

    // Update
    const updateRes = await PUT(
      new Request("http://localhost/api/system-settings/bookbankscreen/BANK-GUID", {
        method: "PUT",
        headers,
        body: JSON.stringify({
          backendUrl: "http://localhost:8888",
          bookcode: "kbank-01",
          passbook: "123-4-56789-0",
          names: [{ code: "th", name: "กสิกรไทย สยาม (แก้ไข)" }],
        }),
      }),
      { params: Promise.resolve({ settingPath: ["bookbankscreen", "BANK-GUID"] }) },
    );
    expect(updateRes.status).toBe(200);
    expect(String(calls[2][0])).toBe("http://localhost:8888/payment/bookbank/BANK-GUID");

    // Delete
    const deleteRes = await DELETE(
      new Request("http://localhost/api/system-settings/bookbankscreen/BANK-GUID", {
        method: "DELETE",
        headers,
        body: JSON.stringify({
          backendUrl: "http://localhost:8888",
        }),
      }),
      { params: Promise.resolve({ settingPath: ["bookbankscreen", "BANK-GUID"] }) },
    );
    expect(deleteRes.status).toBe(200);
    expect(String(calls[3][0])).toBe("http://localhost:8888/payment/bookbank/BANK-GUID");
    expect(calls[3][1]?.method).toBe("DELETE");
  });

  it("rejects book bank creation with 400 when bank name is missing instead of fabricating banknames from bankcode", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const headers = {
      "Content-Type": "application/json",
      Authorization: "Bearer test-token",
      "x-bc-backend-url": "http://localhost:8888",
    };

    const res = await POST(
      new Request("http://localhost/api/system-settings/bookbankscreen", {
        method: "POST",
        headers,
        body: JSON.stringify({
          backendUrl: "http://localhost:8888",
          bookcode: "kbank-01",
          bankcode: "KBANK",
        }),
      }),
      { params: Promise.resolve({ settingPath: ["bookbankscreen"] }) },
    );

    expect(res.status).toBe(400);
    const data = await res.json();
    expect(data).toEqual({ success: false, message: "กรุณาระบุชื่อธนาคาร" });
    expect(fetchMock).not.toHaveBeenCalled();
  });
});

function signJwt(payload: Record<string, unknown>): string {
  const encodedHeader = encodeBase64Url(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const encodedPayload = encodeBase64Url(JSON.stringify({ exp: Math.floor(Date.now() / 1000) + 60, ...payload }));
  const signature = createHmac("sha256", SECRET).update(`${encodedHeader}.${encodedPayload}`).digest("base64url");
  return `${encodedHeader}.${encodedPayload}.${signature}`;
}

function encodeBase64Url(value: string): string {
  return Buffer.from(value).toString("base64url");
}
