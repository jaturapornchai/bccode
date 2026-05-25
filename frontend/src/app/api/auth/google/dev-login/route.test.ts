import { afterEach, describe, expect, it, vi } from "vitest";
import { GET, POST } from "./route";

describe("dev Google login route", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.unstubAllGlobals();
  });

  it("is disabled in production even for localhost", async () => {
    vi.stubEnv("NODE_ENV", "production");

    const response = await GET(new Request("http://localhost/api/auth/google/dev-login", {
      headers: { host: "localhost:3000" },
    }));
    const json = await response.json();

    expect(json).toEqual({ enabled: false });
  });

  it("blocks production login POST before contacting mainapi", async () => {
    vi.stubEnv("NODE_ENV", "production");
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/google/dev-login", {
      method: "POST",
      headers: { host: "localhost:3000", "Content-Type": "application/json" },
      body: JSON.stringify({ backendUrl: "http://localhost:8888/goapi" }),
    }));
    const json = await response.json();

    expect(response.status).toBe(403);
    expect(json).toMatchObject({ success: false });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("allows configured DEV host in production", async () => {
    vi.stubEnv("NODE_ENV", "production");
    vi.stubEnv("BC_ENABLE_DEV_GOOGLE_LOGIN", "true");

    const response = await GET(new Request("https://dev.bcaicloud.com/api/auth/google/dev-login", {
      headers: { host: "dev.bcaicloud.com" },
    }));
    const json = await response.json();

    expect(json).toEqual({
      enabled: true,
      email: "jaturapornchai@gmail.com",
    });
  });

  it("is enabled for localhost outside production", async () => {
    vi.stubEnv("NODE_ENV", "test");

    const response = await GET(new Request("http://localhost/api/auth/google/dev-login", {
      headers: { host: "localhost:3000" },
    }));
    const json = await response.json();

    expect(json).toEqual({
      enabled: true,
      email: "jaturapornchai@gmail.com",
    });
  });
});
