import { afterEach, describe, expect, it, vi } from "vitest";
import { DELETE, GET, POST, PUT } from "./route";

describe("product barcode proxy route", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("proxies product barcode detail to the main API", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void init;
      return Response.json({ success: true, data: { guidfixed: "PB-GUID", barcode: "8850001" } });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(
      new Request("http://localhost/api/product-barcode/PB-GUID", {
        headers: {
          "x-bc-backend-url": "http://localhost:8888",
          Authorization: "Bearer test-token",
        },
      }),
      { params: Promise.resolve({ barcodePath: ["PB-GUID"] }) },
    );

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0] as [string | URL | Request, RequestInit | undefined];
    expect(String(url)).toBe("http://localhost:8888/product/barcode/PB-GUID");
    expect(init?.method).toBe("GET");
  });

  it("proxies selected product barcode deletes as a legacy guid array", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void url;
      void init;
      return Response.json({ success: true });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await DELETE(
      new Request("http://localhost/api/product-barcode", {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer test-token",
        },
        body: JSON.stringify({
          backendUrl: "http://localhost:8888",
          guids: ["GUID-1", "GUID-2"],
        }),
      }),
      { params: Promise.resolve({ barcodePath: [] }) },
    );

    expect(response.status).toBe(200);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0] as [string | URL | Request, RequestInit | undefined];
    expect(String(url)).toBe("http://localhost:8888/product/barcode");
    expect(init?.method).toBe("DELETE");
    expect(init?.body).toBe(JSON.stringify(["GUID-1", "GUID-2"]));
  });

  it("proxies product barcode creates to the main API", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void url;
      void init;
      return Response.json({ success: true });
    });
    vi.stubGlobal("fetch", fetchMock);

    const payload = { holdingcode: "bctest01", barcode: "8850002", names: [{ code: "th", name: "สินค้า" }] };
    const response = await POST(
      new Request("http://localhost/api/product-barcode", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer test-token",
        },
        body: JSON.stringify({
          backendUrl: "http://localhost:8888",
          data: payload,
        }),
      }),
    );

    expect(response.status).toBe(200);
    const [url, init] = fetchMock.mock.calls[0] as [string | URL | Request, RequestInit | undefined];
    expect(String(url)).toBe("http://localhost:8888/product/barcode");
    expect(init?.method).toBe("POST");
    expect(init?.body).toBe(JSON.stringify(payload));
  });

  it("proxies product barcode updates to the main API", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      void url;
      void init;
      return Response.json({ success: true });
    });
    vi.stubGlobal("fetch", fetchMock);

    const payload = { holdingcode: "bctest01", guidfixed: "PB-GUID", barcode: "8850002" };
    const response = await PUT(
      new Request("http://localhost/api/product-barcode/PB-GUID", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer test-token",
        },
        body: JSON.stringify({
          backendUrl: "http://localhost:8888",
          data: payload,
        }),
      }),
      { params: Promise.resolve({ barcodePath: ["PB-GUID"] }) },
    );

    expect(response.status).toBe(200);
    const [url, init] = fetchMock.mock.calls[0] as [string | URL | Request, RequestInit | undefined];
    expect(String(url)).toBe("http://localhost:8888/product/barcode/PB-GUID");
    expect(init?.method).toBe("PUT");
    expect(init?.body).toBe(JSON.stringify(payload));
  });
});
