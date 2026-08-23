import { afterEach, describe, expect, it, vi } from "vitest";
import { PUT } from "./route";

describe("password reset link route", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("fails closed while reset-link delivery is unavailable", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await PUT(new Request("http://localhost/api/auth/profile/reset-password", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username: "uat_user" }),
    }));

    expect(response.status).toBe(501);
    expect(await response.json()).toEqual({
      success: false,
      message: "ระบบส่งลิงก์รีเซ็ตรหัสผ่านยังไม่พร้อมใช้งาน",
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
