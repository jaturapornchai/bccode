import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { PUT } from "./route";

describe("change password route", () => {
  beforeEach(() => {
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
  });

  afterEach(() => {
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    vi.unstubAllGlobals();
  });

  it("clears the refresh cookie after backend confirms password change and session revocation", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Response.json({ success: true })));

    const response = await PUT(request());

    expect(response.status).toBe(200);
    expect(response.headers.get("set-cookie")).toMatch(/bc_refresh_token=;.*Max-Age=0/i);
  });

  it("preserves the refresh cookie when password change fails", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => Response.json({ success: false }, { status: 400 })));

    const response = await PUT(request());

    expect(response.status).toBe(400);
    expect(response.headers.get("set-cookie")).toBeNull();
  });
});

function request(): Request {
  return new Request("http://localhost/api/auth/profile", {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer access-current",
      "x-bc-backend-url": "http://localhost:8888/goapi",
      Cookie: "bc_refresh_token=refresh-current",
    },
    body: JSON.stringify({
      backendUrl: "http://localhost:8888/goapi",
      currentpassword: "current-password-safe",
      newpassword: "new-password-safe-2026",
    }),
  });
}
