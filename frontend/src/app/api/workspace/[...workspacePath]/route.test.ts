import { afterEach, describe, expect, it, vi } from "vitest";
import { GET, POST } from "./route";

describe("workspace product unit setup route", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("enriches company cards with names from holding info when list-holding only returns ids", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      expect(init?.headers).toMatchObject({ Authorization: "Bearer test-token" });

      if (requestUrl === "http://localhost:8888/list-holding?limit=100") {
        expect(init?.method).toBe("GET");
        return Response.json({
          success: true,
          data: [{ holdingcode: "SHOP001", name: "SHOP001", names: null }],
          total: 1,
        });
      }

      if (requestUrl === "http://localhost:8888/holding/SHOP001") {
        expect(init?.method).toBe("GET");
        return Response.json({
          success: true,
          data: {
            name1: "บริษัท ทดสอบ จำกัด",
            names: [{ code: "th", name: "บริษัท ทดสอบ จำกัด" }],
          },
        });
      }
      if (requestUrl === "http://localhost:8888/select-holding") {
        expect(init?.method).toBe("POST");
        expect(JSON.parse(String(init?.body))).toEqual({ holdingcode: "SHOP001" });
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
      new Request("http://localhost/api/workspace/holdings?backendUrl=http://localhost:8888/goapi", {
        headers: { Authorization: "Bearer test-token" },
      }),
      workspaceContext("holdings"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json.data[0]).toMatchObject({
      holdingcode: "SHOP001",
      name: "บริษัท ทดสอบ จำกัด",
      names: [{ code: "th", name: "บริษัท ทดสอบ จำกัด" }],
      companies: [],
      branches: [],
    });
    expect(fetchMock).toHaveBeenCalledTimes(5);
  });

  it("uses company metadata returned by list-holding without requiring selected holding detail", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      expect(init?.headers).toMatchObject({ Authorization: "Bearer test-token" });

      if (requestUrl === "http://localhost:8888/list-holding?limit=100") {
        return Response.json({
          success: true,
          data: [{
            holdingcode: "SHOP001",
            names: [{ code: "th", name: "บริษัท ทดสอบ จำกัด" }],
            language: "th",
            languageconfigs: [
              { code: "th", name: "ภาษาไทย", isuse: true, isdefault: true },
              { code: "en", name: "English", isuse: true, isdefault: false },
              { code: "lo", name: "ພາສາລາວ", isuse: true, isdefault: false },
            ],
            basecurrency: "THB",
            currencies: ["THB", "USD"],
            dateformat: "dd/MM/yyyy",
            timezone: "Asia/Bangkok",
            usebuddhistcalendar: true,
          }],
          total: 1,
        });
      }
      if (requestUrl === "http://localhost:8888/select-holding") {
        expect(init?.method).toBe("POST");
        expect(JSON.parse(String(init?.body))).toEqual({ holdingcode: "SHOP001" });
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
      new Request("http://localhost/api/workspace/holdings?backendUrl=http://localhost:8888/goapi", {
        headers: { Authorization: "Bearer test-token" },
      }),
      workspaceContext("holdings"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json.data[0]).toMatchObject({
      holdingcode: "SHOP001",
      activelanguages: ["th", "en", "lo"],
      basecurrency: "THB",
      currencies: ["THB", "USD"],
      dateformat: "dd/MM/yyyy",
      timezone: "Asia/Bangkok",
      yeartype: "buddhist",
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

      if (requestUrl === "http://localhost:8888/list-holding?limit=100") {
        return Response.json({
          success: true,
          data: [
            {
              holdingcode: "SHOP_EMPTY",
              names: [{ code: "th", name: "กิจการว่าง" }],
              languageconfigs: [{ code: "th", name: "ภาษาไทย", isuse: true, isdefault: true }],
            },
            {
              holdingcode: "SHOP_WITH_ORG",
              names: [{ code: "th", name: "กิจการมีบริษัท" }],
              languageconfigs: [{ code: "th", name: "ภาษาไทย", isuse: true, isdefault: true }],
            },
          ],
          total: 2,
        });
      }
      if (requestUrl === "http://localhost:8888/select-holding") {
        selectedShop = JSON.parse(String(init?.body)).holdingcode;
        return Response.json({ success: true });
      }
      if (requestUrl === "http://localhost:8888/organization/company") {
        return Response.json({
          success: true,
          data: selectedShop === "SHOP_WITH_ORG"
            ? [{ guidfixed: "COMP001", code: "001", names: [{ code: "th", name: "บริษัท A" }] }]
            : [],
        });
      }
      if (requestUrl === "http://localhost:8888/organization/branch") {
        return Response.json({
          success: true,
          data: selectedShop === "SHOP_WITH_ORG"
            ? [{ guidfixed: "BR001", companyguid: "COMP001", code: "00000", names: [{ code: "th", name: "สำนักงานใหญ่" }] }]
            : [],
        });
      }

      throw new Error(`Unexpected URL ${requestUrl}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/workspace/holdings?backendUrl=http://localhost:8888/goapi", {
        headers: { Authorization: "Bearer test-token" },
      }),
      workspaceContext("holdings"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json.data[0]).toMatchObject({ holdingcode: "SHOP_EMPTY", companies: [], branches: [] });
    expect(json.data[1]).toMatchObject({
      holdingcode: "SHOP_WITH_ORG",
      companies: [{ guidfixed: "COMP001", code: "001" }],
      branches: [{ guidfixed: "BR001", companyguid: "COMP001", code: "00000" }],
    });
    expect(fetchMock).toHaveBeenCalledTimes(7);
  });

  it("filters deleted organization records from workspace holdings", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      expect(init?.headers).toMatchObject({ Authorization: "Bearer test-token" });

      if (requestUrl === "http://localhost:8888/list-holding?limit=100") {
        return Response.json({
          success: true,
          data: [{
            holdingcode: "SHOP_WITH_DELETED_ORG",
            names: [{ code: "th", name: "กิจการมีข้อมูลถูกลบ" }],
            languageconfigs: [{ code: "th", name: "ภาษาไทย", isuse: true, isdefault: true }],
          }],
          total: 1,
        });
      }
      if (requestUrl === "http://localhost:8888/select-holding") {
        expect(JSON.parse(String(init?.body))).toEqual({ holdingcode: "SHOP_WITH_DELETED_ORG" });
        return Response.json({ success: true });
      }
      if (requestUrl === "http://localhost:8888/organization/company") {
        return Response.json({
          success: true,
          data: [
            { guidfixed: "COMP_ACTIVE", code: "001", names: [{ code: "th", name: "บริษัทใช้งาน" }] },
            { guidfixed: "COMP_DELETED", code: "002", names: [{ code: "th", name: "บริษัทลบแล้ว" }], deletedat: "2026-06-03T00:00:00Z" },
          ],
        });
      }
      if (requestUrl === "http://localhost:8888/organization/branch") {
        return Response.json({
          success: true,
          data: [
            { guidfixed: "BR_ACTIVE", companyguid: "COMP_ACTIVE", code: "00000", names: [{ code: "th", name: "สำนักงานใหญ่" }] },
            { guidfixed: "BR_DELETED", companyguid: "COMP_ACTIVE", code: "00001", names: [{ code: "th", name: "สาขาลบแล้ว" }], deletedat: "2026-06-03T00:00:00Z" },
            { guidfixed: "BR_ORPHAN", companyguid: "COMP_DELETED", code: "00002", names: [{ code: "th", name: "สาขาของบริษัทลบแล้ว" }] },
          ],
        });
      }

      throw new Error(`Unexpected URL ${requestUrl}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/workspace/holdings?backendUrl=http://localhost:8888/goapi", {
        headers: { Authorization: "Bearer test-token" },
      }),
      workspaceContext("holdings"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json.data[0]).toMatchObject({
      holdingcode: "SHOP_WITH_DELETED_ORG",
      companies: [{ guidfixed: "COMP_ACTIVE", code: "001" }],
      branches: [{ guidfixed: "BR_ACTIVE", companyguid: "COMP_ACTIVE", code: "00000" }],
    });
    expect(fetchMock).toHaveBeenCalledTimes(4);
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

  it("normalizes holdingcode before creating a Holding", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/create-holding");
      expect(init?.method).toBe("POST");
      expect(JSON.parse(String(init?.body))).toMatchObject({
        holdingcode: "bcnew1",
        name1: "New Holding",
      });
      return Response.json({ success: true, ID: "bcnew1" });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/workspace/create-holding", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: "Bearer test-token" },
        body: JSON.stringify({
          backendUrl: "http://localhost:8888/goapi",
          holdingcode: "BCNew1",
          name1: "New Holding",
          names: [{ code: "th", name: "New Holding" }],
        }),
      }),
      workspaceContext("create-holding"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({ success: true, ID: "bcnew1" });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("rejects invalid holdingcode before creating a Holding", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/workspace/create-holding", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: "Bearer test-token" },
        body: JSON.stringify({
          backendUrl: "http://localhost:8888/goapi",
          holdingcode: "bc-demo",
          name1: "New Holding",
        }),
      }),
      workspaceContext("create-holding"),
    );
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json).toMatchObject({
      success: false,
      message: "holdingcode ต้องใช้ a-z และ 0-9 เท่านั้น ยาว 3-30 ตัว และขึ้นต้นด้วย a-z ห้ามใช้ _ หรือสัญลักษณ์",
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("requires a Holding name before creating a Holding", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/workspace/create-holding", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: "Bearer test-token" },
        body: JSON.stringify({
          backendUrl: "http://localhost:8888/goapi",
          holdingcode: "bcnew1",
          name1: "   ",
        }),
      }),
      workspaceContext("create-holding"),
    );
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json).toMatchObject({ success: false, message: "กรุณากรอกชื่อ Holding" });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("updates a Holding display name through the owner-only holding update endpoint", async () => {
    let selected = false;
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      if (requestUrl === "http://localhost:8888/select-holding") {
        expect(init?.method).toBe("POST");
        expect(JSON.parse(String(init?.body))).toEqual({ holdingcode: "bcnew1" });
        selected = true;
        return Response.json({ success: true });
      }
      if (requestUrl === "http://localhost:8888/holding/bcnew1" && init?.method === "GET") {
        if (!selected) {
          return Response.json({ success: false, message: "Holding not selected." }, { status: 401 });
        }
        return Response.json({
          success: true,
          data: {
            holdingcode: "bcnew1",
            name1: "Old Holding Name",
            names: [
              { code: "th", name: "Old Holding Name" },
              { code: "en", name: "Old Holding EN" },
            ],
            settings: { language: "th", emailowners: ["owner@example.com"] },
            telephone: "020000000",
          },
        });
      }
      if (requestUrl === "http://localhost:8888/holding/bcnew1" && init?.method === "PUT") {
        expect(JSON.parse(String(init?.body))).toMatchObject({
          holdingcode: "bcnew1",
          name1: "New Holding Name",
          names: [
            { code: "th", name: "New Holding Name" },
            { code: "en", name: "Old Holding EN" },
          ],
          settings: { language: "th", emailowners: ["owner@example.com"] },
          telephone: "020000000",
        });
        return Response.json({ success: true, ID: "bcnew1" });
      }
      throw new Error(`Unexpected URL ${requestUrl}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/workspace/update-holding", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: "Bearer owner-token" },
        body: JSON.stringify({
          backendUrl: "http://localhost:8888/goapi",
          holdingcode: "bcnew1",
          name1: "New Holding Name",
        }),
      }),
      workspaceContext("update-holding"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({ success: true, ID: "bcnew1" });
    expect(fetchMock).toHaveBeenCalledTimes(3);
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

  it("normalizes product unit code aliases before filtering and seeding", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      const requestUrl = String(url);
      if (requestUrl.endsWith("/unit/list?offset=0&limit=10000&q=&sort=unitcode:1")) {
        return Response.json({ success: true, data: [{ unitcode: "PCE" }], total: 1 });
      }
      if (requestUrl === "https://raw.githubusercontent.com/smlsoft/dedepos_template/main/unit.json") {
        return Response.json([
          { code: "pce", names: [{ code: "th", name: "ชิ้น" }] },
          { unitcode: "kg", names: [{ code: "th", name: "กิโลกรัม" }] },
        ]);
      }
      if (requestUrl.endsWith("/unit/bulk")) {
        expect(init?.method).toBe("POST");
        expect(JSON.parse(String(init?.body))).toEqual([
          {
            unitcode: "KG",
            names: [{ code: "th", name: "กิโลกรัม", isauto: false, isdelete: false }],
          },
        ]);
        return Response.json({ success: true, bulk_import: { created: ["KG"] } }, { status: 201 });
      }
      throw new Error(`Unexpected URL ${requestUrl}`);
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/workspace/product-units/defaults", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: "Bearer test-token" },
        body: JSON.stringify({ backendUrl: "http://localhost:8888/goapi", unitcodes: ["KG", "PCE"] }),
      }),
      workspaceContext("product-units", "defaults"),
    );
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({ success: true, data: { count: 1, source: "template" } });
    expect(fetchMock).toHaveBeenCalledTimes(3);
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
