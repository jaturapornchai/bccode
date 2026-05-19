import { describe, expect, it } from "vitest";
import { deriveMainApiUrl, normalizeBackendUrl, validateBackendUrl } from "./backend-url";

describe("backend URL helpers", () => {
  it("normalizes a host without scheme", () => {
    expect(normalizeBackendUrl("localhost:8888/goapi/")).toBe("http://localhost:8888/goapi");
  });

  it("derives MainAPI origin from GoAPI URL", () => {
    expect(deriveMainApiUrl("http://localhost:8888/goapi")).toBe("http://localhost:8888");
    expect(deriveMainApiUrl("https://dev-api.bcaicloud.com/goapi")).toBe("https://dev-api.bcaicloud.com");
  });

  it("blocks hosts outside the default allowlist", () => {
    expect(() => validateBackendUrl("http://example.com/goapi")).toThrow("allowlist");
  });

  it("allows configured private backend hosts when explicitly enabled", () => {
    expect(
      validateBackendUrl("http://192.168.1.20:8888/goapi", {
        BC_ALLOW_PRIVATE_BACKENDS: "true",
      }).mainApiUrl,
    ).toBe("http://192.168.1.20:8888");
  });

  it("allows Docker Desktop host gateway by default", () => {
    expect(validateBackendUrl("http://host.docker.internal:8888/goapi").mainApiUrl).toBe("http://host.docker.internal:8888");
  });
});
