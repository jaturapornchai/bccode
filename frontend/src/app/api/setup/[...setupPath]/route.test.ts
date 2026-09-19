import { describe, expect, it } from "vitest";
import { POST } from "./route";

describe("setup handler", () => {
  it("verifies password with 12345", async () => {
    const response = await POST(
      new Request("http://localhost/api/setup/verify-password", {
        method: "POST",
        body: JSON.stringify({ password: "12345" }),
      }),
      { params: Promise.resolve({ setupPath: ["verify-password"] }) },
    );

    expect(response.status).toBe(200);
    const json = await response.json();
    expect(json.success).toBe(true);
    expect(json.message).toContain("สำเร็จ");
  });

  it("verifies password with admin", async () => {
    const response = await POST(
      new Request("http://localhost/api/setup/verify-password", {
        method: "POST",
        body: JSON.stringify({ password: "admin" }),
      }),
      { params: Promise.resolve({ setupPath: ["verify-password"] }) },
    );

    expect(response.status).toBe(200);
    const json = await response.json();
    expect(json.success).toBe(true);
  });

  it("rejects invalid password with 401", async () => {
    const response = await POST(
      new Request("http://localhost/api/setup/verify-password", {
        method: "POST",
        body: JSON.stringify({ password: "wrong-password-999" }),
      }),
      { params: Promise.resolve({ setupPath: ["verify-password"] }) },
    );

    expect(response.status).toBe(401);
    const json = await response.json();
    expect(json.success).toBe(false);
  });

  it("returns config on config/get", async () => {
    const response = await POST(
      new Request("http://localhost/api/setup/config/get", { method: "POST", body: "{}" }),
      { params: Promise.resolve({ setupPath: ["config", "get"] }) },
    );

    expect(response.status).toBe(200);
    const json = await response.json();
    expect(json.success).toBe(true);
    expect(Array.isArray(json.data)).toBe(true);
    expect(json.data.length).toBeGreaterThan(0);
  });
});
