import { describe, expect, it } from "vitest";
import { getRequestHostName, isConfiguredDevLoginRequest, isDevLoginHost, isLocalLoginHost } from "./local-dev-auth";

describe("local dev auth guard", () => {
  it("allows only local login hosts", () => {
    expect(isLocalLoginHost("localhost")).toBe(true);
    expect(isLocalLoginHost("127.0.0.1")).toBe(true);
    expect(isLocalLoginHost("::1")).toBe(true);
    expect(isLocalLoginHost("example.com")).toBe(false);
  });

  it("allows configured DEV Google login hosts only when explicitly enabled", () => {
    expect(isDevLoginHost("dev.bcaicloud.com")).toBe(true);
    expect(isDevLoginHost("example.com")).toBe(false);

    const request = new Request("https://dev.bcaicloud.com/api/auth/google/dev-login", {
      headers: { host: "dev.bcaicloud.com" },
    });

    expect(isConfiguredDevLoginRequest(request, {})).toBe(false);
    expect(isConfiguredDevLoginRequest(request, { BC_ENABLE_DEV_GOOGLE_LOGIN: "true" })).toBe(true);
  });

  it("extracts host names from common host headers", () => {
    expect(getRequestHostName("localhost:3000")).toBe("localhost");
    expect(getRequestHostName("127.0.0.1:3000")).toBe("127.0.0.1");
    expect(getRequestHostName("[::1]:3000")).toBe("::1");
    expect(getRequestHostName("localhost:3000, proxy.local")).toBe("localhost");
  });
});
