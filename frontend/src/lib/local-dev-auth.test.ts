import { describe, expect, it } from "vitest";
import { getRequestHostName, isLocalLoginHost } from "./local-dev-auth";

describe("local dev auth guard", () => {
  it("allows only local login hosts", () => {
    expect(isLocalLoginHost("localhost")).toBe(true);
    expect(isLocalLoginHost("127.0.0.1")).toBe(true);
    expect(isLocalLoginHost("::1")).toBe(true);
    expect(isLocalLoginHost("example.com")).toBe(false);
  });

  it("extracts host names from common host headers", () => {
    expect(getRequestHostName("localhost:3000")).toBe("localhost");
    expect(getRequestHostName("127.0.0.1:3000")).toBe("127.0.0.1");
    expect(getRequestHostName("[::1]:3000")).toBe("::1");
    expect(getRequestHostName("localhost:3000, proxy.local")).toBe("localhost");
  });
});
