import { describe, expect, it } from "vitest";
import { POST } from "./route";

describe("setup proxy", () => {
  it("fails closed until a separate control-plane authentication exists", async () => {
    const response = await POST(
      new Request("http://localhost/api/setup/config/get", { method: "POST", body: "{}" }),
      { params: Promise.resolve({ setupPath: ["config", "get"] }) },
    );

    expect(response.status).toBe(410);
    expect(await response.json()).toEqual({
      success: false,
      message: "Setup API ถูกปิดจนกว่าจะมี Control Plane Authentication ที่แยกจาก Tenant",
    });
  });
});
