import { createHmac } from "crypto";
import { afterEach, describe, expect, it, vi } from "vitest";
import { DELETE, GET, POST, PUT } from "./route";

const SECRET = "test-secret";

describe("system settings API route security", () => {
  afterEach(() => {
    delete process.env.JWT_SECRET_KEY;
    vi.unstubAllGlobals();
  });

  it("asks the backend to validate selected holding before atlas proxying", async () => {
    process.env.JWT_SECRET_KEY = SECRET;
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void url;
      void init;
      return Response.json({ success: false, message: "holdingcode invalid" }, { status: 400 });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/system-settings/permissiondefinition?holdingcode=SHOP002", {
        headers: {
          Authorization: `Bearer ${signJwt({ username: "user@example.com", holdingcode: "SHOP001" })}`,
          "x-bc-backend-url": "http://localhost:8888/goapi",
        },
      }),
      { params: Promise.resolve({ settingPath: ["permissiondefinition"] }) },
    );

    expect(response.status).toBe(403);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [selectUrl, selectInit] = fetchMock.mock.calls[0] as [string | URL | Request, RequestInit | undefined];
    expect(String(selectUrl)).toBe("http://localhost:8888/select-holding");
    expect(JSON.parse(String(selectInit?.body))).toMatchObject({ holdingcode: "SHOP002" });
  });

  it("uses a generated password when creating a username login user", async () => {
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
    expect(fetchMock).toHaveBeenCalledTimes(3);
    const registerBody = JSON.parse(String(fetchMock.mock.calls[1][1]?.body));
    expect(registerBody.password).not.toBe("12345");
    expect(registerBody.password).toMatch(/^[0-9a-f]{24}$/);
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

  it("passes atlas detail ids as email and cartid filters", async () => {
    process.env.JWT_SECRET_KEY = SECRET;
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void url;
      void init;
      return Response.json({ status: "success", code: 200, data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/system-settings/permissionlink/demo.admin01%40example.com?holdingcode=SHOP001", {
        headers: {
          Authorization: `Bearer ${signJwt({ username: "owner@example.com", holdingcode: "SHOP001" })}`,
          "x-bc-backend-url": "http://localhost:8888/goapi",
        },
      }),
      { params: Promise.resolve({ settingPath: ["permissionlink", "demo.admin01%40example.com"] }) },
    );

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    const [, proxiedInit] = fetchMock.mock.calls[1] as [string | URL | Request, RequestInit | undefined];
    const body = JSON.parse(String(proxiedInit?.body));
    expect(body).toMatchObject({
      collection: "employeepermissions",
      holdingcode: "SHOP001",
      email: "demo.admin01@example.com",
      cartid: "demo.admin01@example.com",
    });
  });

  it("passes real holdingcode separately from legacy holdingcode for atlas reads", async () => {
    process.env.JWT_SECRET_KEY = SECRET;
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void url;
      void init;
      return Response.json({ status: "success", code: 200, data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/system-settings/permissionlink?holdingcode=SHOP001", {
        headers: {
          Authorization: `Bearer ${signJwt({ username: "owner@example.com", holdingcode: "SHOP001" })}`,
          "x-bc-backend-url": "http://localhost:8888/goapi",
        },
      }),
      { params: Promise.resolve({ settingPath: ["permissionlink"] }) },
    );

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    const [, proxiedInit] = fetchMock.mock.calls[1] as [string | URL | Request, RequestInit | undefined];
    const body = JSON.parse(String(proxiedInit?.body));
    expect(body).toMatchObject({
      collection: "employeepermissions",
      holdingcode: "SHOP001",
    });
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
      if (requestUrl.includes("/register/exists-username")) return Response.json({ success: true, data: true });
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
      "http://localhost:8888/register/exists-username",
      "http://localhost:8888/holding/permission?holdingcode=SHOP001",
      "http://localhost:8888/holding/permission?holdingcode=SHOP001",
      "http://localhost:8888/holding/permission/newuser@example.com?holdingcode=SHOP001",
    ]);
    expect(calls[2][1]?.method).toBe("PUT");
    expect(calls[3][1]?.method).toBe("PUT");
    expect(calls[4][1]?.method).toBe("DELETE");
  });

  it.each([
    ["permissiondefinition", "permissiondefinitions", "permissioncode"],
    ["permissiongroup", "permissiongroups", "groupcode"],
    ["permissionlink", "employeepermissions", "employeecode"],
    ["approvalsetting", "approvalsettings", "approvalcode"],
  ])("supports atlas CRUD proxy for %s", async (slug, collection, codeKey) => {
    const calls: Array<[string, RequestInit | undefined]> = [];
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      calls.push([String(url), init]);
      return Response.json({ success: true, data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    const headers = {
      "Content-Type": "application/json",
      Authorization: "Bearer test-token",
      "x-bc-backend-url": "http://localhost:8888/goapi",
    };
    const body = {
      holdingcode: "SHOP001",
      guidfixed: "GUID001",
      [codeKey]: "CODE001",
      accessscopes: ["B001"],
    };

    await GET(
      new Request(`http://localhost/api/system-settings/${slug}?holdingcode=SHOP001&limit=100&offset=0`, { headers }),
      { params: Promise.resolve({ settingPath: [slug] }) },
    );
    await POST(
      new Request(`http://localhost/api/system-settings/${slug}?holdingcode=SHOP001`, {
        method: "POST",
        headers,
        body: JSON.stringify(body),
      }),
      { params: Promise.resolve({ settingPath: [slug] }) },
    );
    await PUT(
      new Request(`http://localhost/api/system-settings/${slug}/GUID001?holdingcode=SHOP001`, {
        method: "PUT",
        headers,
        body: JSON.stringify(body),
      }),
      { params: Promise.resolve({ settingPath: [slug, "GUID001"] }) },
    );
    await DELETE(
      new Request(`http://localhost/api/system-settings/${slug}/GUID001?holdingcode=SHOP001`, {
        method: "DELETE",
        headers,
        body: JSON.stringify({ holdingcode: "SHOP001" }),
      }),
      { params: Promise.resolve({ settingPath: [slug, "GUID001"] }) },
    );

    expect(calls.map(([url]) => url)).toEqual([
      "http://localhost:8888/select-holding",
      "http://localhost:8888/goapi/atlas/get",
      "http://localhost:8888/select-holding",
      "http://localhost:8888/goapi/atlas/update?holdingcode=SHOP001",
      "http://localhost:8888/select-holding",
      "http://localhost:8888/goapi/atlas/update?holdingcode=SHOP001",
      "http://localhost:8888/select-holding",
      "http://localhost:8888/goapi/atlas/delete?holdingcode=SHOP001",
    ]);
    const readBody = JSON.parse(String(calls[1][1]?.body));
    const createBody = JSON.parse(String(calls[3][1]?.body));
    const updateBody = JSON.parse(String(calls[5][1]?.body));
    const deleteBody = JSON.parse(String(calls[7][1]?.body));
    expect(readBody).toMatchObject({ collection, holdingcode: "SHOP001" });
    expect(createBody).toMatchObject({ collection, holdingcode: "SHOP001", guidfixed: "GUID001" });
    expect(updateBody).toMatchObject({ collection, holdingcode: "SHOP001", guidfixed: "GUID001" });
    expect(deleteBody).toMatchObject({ collection, holdingcode: "SHOP001", guidfixed: "GUID001", deletemany: false });
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
