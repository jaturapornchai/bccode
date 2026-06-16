import { afterEach, describe, expect, it, vi } from "vitest";
import { POST } from "./route";

describe("password login route", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("requires holdingcode before password login", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        backendUrl: "http://localhost:8888/goapi",
        username: "demo",
        password: "secret",
      }),
    }));
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json).toMatchObject({
      success: false,
      message: "กรุณากรอกรหัสกลุ่มกิจการก่อนเข้าสู่ระบบด้วย User , Password",
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("forwards normalized holdingcode to mainapi login", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/login");
      expect(init?.method).toBe("POST");
      expect(JSON.parse(String(init?.body))).toEqual({
        username: "demo",
        password: "secret",
        holdingcode: "bcdemo",
      });
      return Response.json({ success: true, token: "token-1", refresh: "refresh-1" });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        backendUrl: "http://localhost:8888/goapi",
        username: "demo",
        password: "secret",
        holdingcode: "BCDemo",
      }),
    }));
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json).toMatchObject({
      success: true,
      token: "token-1",
      refresh: "refresh-1",
      backendUrl: "http://localhost:8888/goapi",
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("rejects holdingcode with underscore", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        backendUrl: "http://localhost:8888/goapi",
        username: "demo",
        password: "secret",
        holdingcode: "bc_demo",
      }),
    }));
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json).toMatchObject({
      success: false,
      message: "holdingcode ต้องใช้ a-z และ 0-9 เท่านั้น ยาว 3-30 ตัว และขึ้นต้นด้วย a-z ห้ามใช้ _ หรือสัญลักษณ์",
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
