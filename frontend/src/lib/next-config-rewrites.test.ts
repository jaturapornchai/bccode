import { existsSync } from "node:fs";
import { resolve } from "node:path";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import nextConfig from "../../next.config";
import { getSystemSettingConfig } from "./system-setting-screens";

// There is no public /backend proxy to mainapi any more (ADR docs/kms/decisions/2026-09-27-backend-proxy-allowlist.md,
// stage 2): company/branch go through src/app/api/organization and private files through src/app/api/files.
// These checks fail if a rewrite, middleware/proxy or page route ever puts a road to mainapi back.

const localBackend = "http://mainapi.test:8888";
const frontendRoot = resolve(process.cwd());

beforeEach(() => {
  vi.stubEnv("BCAI_LOCAL_BACKEND_URL", localBackend);
});

afterEach(() => {
  vi.unstubAllEnvs();
});

describe("next.config has no proxy to mainapi", () => {
  it("defines no rewrites or redirects at all", () => {
    expect(nextConfig.rewrites).toBeUndefined();
    expect(nextConfig.redirects).toBeUndefined();
  });

  it("does not mention the backend host anywhere in the config", async () => {
    const headers = (await nextConfig.headers?.()) ?? [];
    const serialized = JSON.stringify({ ...nextConfig, headers });
    expect(serialized).not.toContain(localBackend);
    expect(serialized).not.toMatch(/mainapi|:8888|\/backend/i);
  });
});

describe("/backend/* is not served by the frontend", () => {
  it("has no middleware/proxy file that could rewrite requests", () => {
    for (const file of ["middleware.ts", "middleware.js", "proxy.ts", "proxy.js"]) {
      expect(existsSync(resolve(frontendRoot, file)), file).toBe(false);
      expect(existsSync(resolve(frontendRoot, "src", file)), `src/${file}`).toBe(false);
    }
  });

  it("has no app route named backend, and the [systemSetting] page 404s for it", () => {
    expect(existsSync(resolve(frontendRoot, "src/app/backend"))).toBe(false);
    // src/app/[systemSetting]/page.tsx calls notFound() when there is no setting config for the slug.
    expect(getSystemSettingConfig("backend")).toBeUndefined();
    expect(getSystemSettingConfig("/backend")).toBeUndefined();
  });
});
