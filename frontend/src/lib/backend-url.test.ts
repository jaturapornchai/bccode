import { describe, expect, it } from "vitest";
import {
  deriveMainApiUrl,
  localGoApiUrlForOrigin,
  migrateRuntimeBackendUrl,
  migrateSameOriginLegacyGoApiUrl,
  normalizeBackendUrl,
  publicGoApiUrlForOrigin,
  runtimeGoApiUrlForOrigin,
  validateBackendUrl,
} from "./backend-url";

describe("backend URL helpers", () => {
  it("normalizes a host without scheme", () => {
    expect(normalizeBackendUrl("localhost:8888/goapi/")).toBe("http://localhost:8888/goapi");
  });

  it("derives MainAPI origin from GoAPI URL", () => {
    expect(deriveMainApiUrl("http://localhost:8888/goapi")).toBe("http://localhost:8888");
    expect(deriveMainApiUrl("https://dev-api.bcaicloud.com/goapi")).toBe("https://dev-api.bcaicloud.com");
    expect(deriveMainApiUrl("https://dev.bcaicloud.com/backend/goapi")).toBe("https://dev.bcaicloud.com/backend");
  });

  it("uses the public DEV router path for same-origin GoAPI URLs", () => {
    expect(publicGoApiUrlForOrigin("https://dev.bcaicloud.com")).toBe("https://dev.bcaicloud.com/backend/goapi");
    expect(runtimeGoApiUrlForOrigin("https://dev.bcaicloud.com")).toBe("https://dev.bcaicloud.com/backend/goapi");
    expect(migrateSameOriginLegacyGoApiUrl("https://dev.bcaicloud.com/goapi", "https://dev.bcaicloud.com")).toBe(
      "https://dev.bcaicloud.com/backend/goapi",
    );
    expect(migrateSameOriginLegacyGoApiUrl("http://localhost:8888/goapi", "http://localhost:3000")).toBe(
      "http://localhost:8888/goapi",
    );
  });

  it("uses the local backend that matches the local frontend URL", () => {
    expect(localGoApiUrlForOrigin("http://localhost:3000")).toBe("http://localhost:3000/backend/goapi");
    expect(runtimeGoApiUrlForOrigin("http://127.0.0.1:3000")).toBe("http://127.0.0.1:3000/backend/goapi");
    expect(migrateRuntimeBackendUrl("http://45.144.166.112:8888/goapi", "http://localhost:3000")).toBe(
      "http://localhost:3000/backend/goapi",
    );
    expect(migrateRuntimeBackendUrl("https://dev.bcaicloud.com/backend/goapi", "http://localhost:3000")).toBe(
      "http://localhost:3000/backend/goapi",
    );
    expect(migrateRuntimeBackendUrl("http://localhost:8888/goapi", "http://localhost:3000")).toBe("http://localhost:3000/backend/goapi");
    expect(migrateRuntimeBackendUrl("http://localhost:8888/goapi", "https://dev.bcaicloud.com")).toBe(
      "https://dev.bcaicloud.com/backend/goapi",
    );
    expect(migrateRuntimeBackendUrl("http://45.144.166.112:8888/goapi", "https://dev.bcaicloud.com")).toBe(
      "https://dev.bcaicloud.com/backend/goapi",
    );
  });

  it("blocks hosts outside the default allowlist", () => {
    expect(() => validateBackendUrl("http://example.com/goapi")).toThrow("allowlist");
  });

  it("blocks direct DEV server backend ports by default", () => {
    expect(() => validateBackendUrl("http://45.144.166.112:8888/goapi")).toThrow("allowlist");
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
