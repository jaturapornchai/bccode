import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { POST } from "./route";

describe("Google verify route", () => {
  beforeEach(() => {
    process.env.GOOGLE_CLIENT_ID = "google-client-id";
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
  });

  afterEach(() => {
    delete process.env.GOOGLE_CLIENT_ID;
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    vi.unstubAllGlobals();
  });

  it("keeps the refresh token in an HttpOnly cookie", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request) => {
      if (String(url).startsWith("https://oauth2.googleapis.com/tokeninfo")) {
        return Response.json({
          aud: "google-client-id",
          iss: "https://accounts.google.com",
          exp: String(Math.floor(Date.now() / 1000) + 300),
          email: "uat@example.com",
          email_verified: true,
          sub: "google-subject",
        });
      }
      expect(String(url)).toBe("http://localhost:8888/googlelogin");
      return Response.json({ token: "access-google", refresh: "refresh-google", username: "uat_user" });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/google/verify", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ credential: "google-id-token" }),
    }));
    const json = await response.json();

    expect(json).toMatchObject({ success: true, token: "access-google" });
    expect(json).not.toHaveProperty("refresh");
    const cookie = response.headers.get("set-cookie") ?? "";
    expect(cookie).toContain("bc_refresh_token=refresh-google");
    expect(cookie).toContain("HttpOnly");
    expect(cookie).toContain("Secure");
    expect(cookie).toContain("SameSite=lax");
  });
});
