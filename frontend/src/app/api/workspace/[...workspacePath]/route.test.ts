import { afterEach, describe, expect, it, vi } from "vitest";
import { GET, POST } from "./route";

describe("workspace product unit setup route", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("enriches company cards with names from shop info when list-shop only returns ids", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      expect(init?.headers).toMatchObject({ Authorization: "Bearer test-token" });

      if (requestUrl === "http://localhost:8888/list-shop?limit=100") {
        expect(init?.method).toBe("GET");
        return Response.json({
          success: true,
          data: [{ shopid: "SHOP001", name: "SHOP001", names: null }],
          total: 1,
        });
      }

      if (requestUrl === "http://localhost:8888/shop/SHOP001") {
        expect(init?.method).toBe("GET");
        return Response.json({
          success: true,
          data: {
            name1: "บริษัท ทดสอบ จำกัด",
            names: [{ code: "th", name: "บริษัท ทดสอบ จำกัด" }],
          },
        });
      }
      if (requestUrl === "http://localhost:8888/select-shop") {
        expect(init?.method).toBe("POST");
        expect(JSON.parse(String(init?.body))).toEqual({ shopid: "SHOP001" });
        return Response.json({ success: true });
      }
      if (requestUrl === "http://localhost:8888/organization/company") {
        expect(init?.method).toBe("GET");
        return Response.json({ success: true, data: [] });
      }
      if (requestUrl === "http://localhost:8888/organization/branch") {
        expect(init?.method).toBe("GET");
        return Response.json({ success: true, data: [] });
      }

      throw new Error(`Unexpected URL ${requestUrl}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/workspace/shops?backendUrl=http://localhost:8888/goapi", {
        headers: { Authorization: "Bearer test-token" },
      }),
      workspaceContext("shops"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json.data[0]).toMatchObject({
      shopid: "SHOP001",
      name: "บริษัท ทดสอบ จำกัด",
      names: [{ code: "th", name: "บริษัท ทดสอบ จำกัด" }],
      companies: [],
      branches: [],
    });
    expect(fetchMock).toHaveBeenCalledTimes(5);
  });

  it("uses company metadata returned by list-shop without requiring selected shop detail", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      expect(init?.headers).toMatchObject({ Authorization: "Bearer test-token" });

      if (requestUrl === "http://localhost:8888/list-shop?limit=100") {
        return Response.json({
          success: true,
          data: [{
            shopid: "SHOP001",
            names: [{ code: "th", name: "บริษัท ทดสอบ จำกัด" }],
            language: "th",
            languageconfigs: [
              { code: "th", name: "ภาษาไทย", is_use: true, isdefault: true },
              { code: "en", name: "English", is_use: true, isdefault: false },
              { code: "lo", name: "ພາສາລາວ", is_use: true, isdefault: false },
            ],
            base_currency: "THB",
            currencies: ["THB", "USD"],
            date_format: "dd/MM/yyyy",
            timezone: "Asia/Bangkok",
            usebuddhistcalendar: true,
          }],
          total: 1,
        });
      }
      if (requestUrl === "http://localhost:8888/select-shop") {
        expect(init?.method).toBe("POST");
        expect(JSON.parse(String(init?.body))).toEqual({ shopid: "SHOP001" });
        return Response.json({ success: true });
      }
      if (requestUrl === "http://localhost:8888/organization/company") {
        expect(init?.method).toBe("GET");
        return Response.json({ success: true, data: [] });
      }
      if (requestUrl === "http://localhost:8888/organization/branch") {
        expect(init?.method).toBe("GET");
        return Response.json({ success: true, data: [] });
      }

      throw new Error(`Unexpected URL ${requestUrl}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/workspace/shops?backendUrl=http://localhost:8888/goapi", {
        headers: { Authorization: "Bearer test-token" },
      }),
      workspaceContext("shops"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json.data[0]).toMatchObject({
      shopid: "SHOP001",
      active_languages: ["th", "en", "lo"],
      base_currency: "THB",
      currencies: ["THB", "USD"],
      date_format: "dd/MM/yyyy",
      timezone: "Asia/Bangkok",
      year_type: "buddhist",
      companies: [],
      branches: [],
    });
    expect(fetchMock).toHaveBeenCalledTimes(4);
  });

  it("attaches organization companies and branches to the shop selected for each lookup", async () => {
    let selectedShop = "";
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      expect(init?.headers).toMatchObject({ Authorization: "Bearer test-token" });

      if (requestUrl === "http://localhost:8888/list-shop?limit=100") {
        return Response.json({
          success: true,
          data: [
            {
              shopid: "SHOP_EMPTY",
              names: [{ code: "th", name: "กิจการว่าง" }],
              languageconfigs: [{ code: "th", name: "ภาษาไทย", is_use: true, isdefault: true }],
            },
            {
              shopid: "SHOP_WITH_ORG",
              names: [{ code: "th", name: "กิจการมีบริษัท" }],
              languageconfigs: [{ code: "th", name: "ภาษาไทย", is_use: true, isdefault: true }],
            },
          ],
          total: 2,
        });
      }
      if (requestUrl === "http://localhost:8888/select-shop") {
        selectedShop = JSON.parse(String(init?.body)).shopid;
        return Response.json({ success: true });
      }
      if (requestUrl === "http://localhost:8888/organization/company") {
        return Response.json({
          success: true,
          data: selectedShop === "SHOP_WITH_ORG"
            ? [{ guid_fixed: "COMP001", code: "001", names: [{ code: "th", name: "บริษัท A" }] }]
            : [],
        });
      }
      if (requestUrl === "http://localhost:8888/organization/branch") {
        return Response.json({
          success: true,
          data: selectedShop === "SHOP_WITH_ORG"
            ? [{ guid_fixed: "BR001", company_guid: "COMP001", code: "00000", names: [{ code: "th", name: "สำนักงานใหญ่" }] }]
            : [],
        });
      }

      throw new Error(`Unexpected URL ${requestUrl}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/workspace/shops?backendUrl=http://localhost:8888/goapi", {
        headers: { Authorization: "Bearer test-token" },
      }),
      workspaceContext("shops"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json.data[0]).toMatchObject({ shopid: "SHOP_EMPTY", companies: [], branches: [] });
    expect(json.data[1]).toMatchObject({
      shopid: "SHOP_WITH_ORG",
      companies: [{ guid_fixed: "COMP001", code: "001" }],
      branches: [{ guid_fixed: "BR001", company_guid: "COMP001", code: "00000" }],
    });
    expect(fetchMock).toHaveBeenCalledTimes(7);
  });

  it("proxies the product unit existence check to mainapi", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/unit/list?offset=0&limit=1&q=&sort=unitcode:1");
      expect(init?.method).toBe("GET");
      return Response.json({ success: true, data: [], total: 0 });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/workspace/product-units?backendUrl=http://localhost:8888/goapi", {
        headers: { Authorization: "Bearer test-token" },
      }),
      workspaceContext("product-units"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({ success: true, total: 0 });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("lists only standard product units that are not in the current company", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      if (requestUrl.endsWith("/unit/list?offset=0&limit=10000&q=&sort=unitcode:1")) {
        expect(init?.method).toBe("GET");
        return Response.json({ success: true, data: [{ unitcode: "PCE" }], total: 1 });
      }
      if (requestUrl === "https://raw.githubusercontent.com/smlsoft/dedepos_template/main/unit.json") {
        return Response.json([
          { unitcode: "PCE", names: [{ code: "th", name: "ชิ้น" }] },
          { unitcode: "KG", names: [{ code: "th", name: "กิโลกรัม" }] },
        ]);
      }
      throw new Error(`Unexpected URL ${requestUrl}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/workspace/product-units/standard?backendUrl=http://localhost:8888/goapi", {
        headers: { Authorization: "Bearer test-token" },
      }),
      workspaceContext("product-units", "standard"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({ success: true, total: 1, data: [{ unitcode: "KG" }] });
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("does not seed anything when selected product units already exist", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      if (requestUrl.endsWith("/unit/list?offset=0&limit=10000&q=&sort=unitcode:1")) {
        expect(init?.method).toBe("GET");
        return Response.json({ success: true, data: [{ unitcode: "PCE" }], total: 1 });
      }
      if (requestUrl === "https://raw.githubusercontent.com/smlsoft/dedepos_template/main/unit.json") {
        return Response.json([{ unitcode: "PCE", names: [{ code: "th", name: "ชิ้น" }] }]);
      }
      throw new Error(`Unexpected URL ${requestUrl}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/workspace/product-units/defaults", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: "Bearer test-token" },
        body: JSON.stringify({ backendUrl: "http://localhost:8888/goapi", unitcodes: ["PCE"] }),
      }),
      workspaceContext("product-units", "defaults"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({ success: true, data: { count: 0, source: "template" } });
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("seeds selected template product units through unit bulk API when none exist", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      if (requestUrl.endsWith("/unit/list?offset=0&limit=10000&q=&sort=unitcode:1")) {
        return Response.json({ success: true, data: [], total: 0 });
      }
      if (requestUrl === "https://raw.githubusercontent.com/smlsoft/dedepos_template/main/unit.json") {
        return Response.json([
          { unitcode: "PCE", names: [{ code: "th", name: "ชิ้น" }, { code: "en", name: "Piece" }] },
          { unitcode: "KG", names: [{ code: "th", name: "กิโลกรัม" }] },
        ]);
      }
      if (requestUrl.endsWith("/unit/bulk")) {
        expect(init?.method).toBe("POST");
        expect(JSON.parse(String(init?.body))).toEqual([
          {
            unitcode: "PCE",
            names: [
              { code: "th", name: "ชิ้น", isauto: false, isdelete: false },
              { code: "en", name: "Piece", isauto: false, isdelete: false },
            ],
          },
        ]);
        return Response.json({ success: true, bulk_import: { created: ["PCE"] } }, { status: 201 });
      }
      throw new Error(`Unexpected URL ${requestUrl}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/workspace/product-units/defaults", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: "Bearer test-token" },
        body: JSON.stringify({ backendUrl: "http://localhost:8888/goapi", unitcodes: ["PCE"] }),
      }),
      workspaceContext("product-units", "defaults"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({ success: true, data: { count: 1, source: "template" } });
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });
});

function workspaceContext(...workspacePath: string[]) {
  return { params: Promise.resolve({ workspacePath }) };
}
