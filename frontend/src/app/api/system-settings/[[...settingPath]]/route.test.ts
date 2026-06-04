import { createHmac } from "crypto";
import { afterEach, describe, expect, it, vi } from "vitest";
import { DELETE, GET, POST } from "./route";

const SECRET = "test-secret";

describe("system settings API route security", () => {
  afterEach(() => {
    delete process.env.JWT_SECRET_KEY;
    vi.unstubAllGlobals();
  });

  it("rejects cross-tenant requests before proxying", async () => {
    process.env.JWT_SECRET_KEY = SECRET;

    const response = await GET(
      new Request("http://localhost/api/system-settings/permission_definition?holding_code=SHOP002", {
        headers: { Authorization: `Bearer ${signJwt({ username: "user@example.com", holding_code: "SHOP001" })}` },
      }),
      { params: Promise.resolve({ settingPath: ["permission_definition"] }) },
    );

    expect(response.status).toBe(403);
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
      new Request("http://localhost/api/system-settings/user?holding_code=SHOP001", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${signJwt({ username: "owner@example.com", holding_code: "SHOP001" })}`,
        },
        body: JSON.stringify({
          backendUrl: "http://localhost:8888",
          holding_code: "SHOP001",
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
      new Request("http://localhost/api/system-settings/productunit/UNIT-GUID?holding_code=SHOP001", {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer test-token",
        },
        body: JSON.stringify({
          backendUrl: "http://localhost:8888",
          holding_code: "SHOP001",
        }),
      }),
      { params: Promise.resolve({ settingPath: ["productunit", "UNIT-GUID"] }) },
    );

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [proxiedUrl, proxiedInit] = fetchMock.mock.calls[0] as [string | URL | Request, RequestInit | undefined];
    expect(String(proxiedUrl)).toBe("http://localhost:8888/unit/UNIT-GUID?holding_code=SHOP001");
    expect(proxiedInit?.method).toBe("DELETE");
    expect(proxiedInit?.body).toBeUndefined();
  });

  it("keeps email user ids readable by the legacy shop permission endpoint", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void url;
      void init;
      return Response.json({ success: true, data: { username: "demo.admin01@example.com" } });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/system-settings/user/demo.admin01%40example.com?holding_code=SHOP001", {
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
    expect(String(proxiedUrl)).toBe("http://localhost:8888/shop/permission/demo.admin01@example.com?offset=0&limit=1000");
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
      new Request("http://localhost/api/system-settings/permission_link/demo.admin01%40example.com?holding_code=SHOP001", {
        headers: {
          Authorization: `Bearer ${signJwt({ username: "owner@example.com", holding_code: "SHOP001" })}`,
          "x-bc-backend-url": "http://localhost:8888/goapi",
        },
      }),
      { params: Promise.resolve({ settingPath: ["permission_link", "demo.admin01%40example.com"] }) },
    );

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [, proxiedInit] = fetchMock.mock.calls[0] as [string | URL | Request, RequestInit | undefined];
    const body = JSON.parse(String(proxiedInit?.body));
    expect(body).toMatchObject({
      collection: "employee_permissions",
      holding_code: "SHOP001",
      email: "demo.admin01@example.com",
      cartid: "demo.admin01@example.com",
    });
  });

  it("passes real holding_code separately from legacy holding_code for atlas reads", async () => {
    process.env.JWT_SECRET_KEY = SECRET;
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void url;
      void init;
      return Response.json({ status: "success", code: 200, data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/system-settings/permission_link?holding_code=SHOP001", {
        headers: {
          Authorization: `Bearer ${signJwt({ username: "owner@example.com", holding_code: "SHOP001" })}`,
          "x-bc-backend-url": "http://localhost:8888/goapi",
        },
      }),
      { params: Promise.resolve({ settingPath: ["permission_link"] }) },
    );

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [, proxiedInit] = fetchMock.mock.calls[0] as [string | URL | Request, RequestInit | undefined];
    const body = JSON.parse(String(proxiedInit?.body));
    expect(body).toMatchObject({
      collection: "employee_permissions",
      holding_code: "SHOP001",
    });
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
