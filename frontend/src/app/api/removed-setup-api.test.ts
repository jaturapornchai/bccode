import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join, resolve } from "node:path";
import { describe, expect, it } from "vitest";

// 2026-09-24: the /settings "setup center" and its BFF routes were deleted.
// They took no session, accepted an empty/"12345"/"admin" password, returned
// internal hostnames, and let anyone on the public URL make the frontend
// server fetch any URL or open a TCP socket to any host:port (blind SSRF).
// Nothing read the config they saved. See
// docs/kms/bugs/2026-09-24-unauthenticated-setup-api-ssrf.md before adding
// any of this back.
const apiDir = resolve(process.cwd(), "src", "app", "api");

function routeFiles(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) return routeFiles(path);
    return entry.name === "route.ts" ? [path] : [];
  });
}

describe("removed setup API", () => {
  it.each([
    ["src/app/settings"],
    ["src/app/api/setup"],
    ["src/app/api/storage/health"],
  ])("%s stays deleted", (path) => {
    expect(existsSync(resolve(process.cwd(), path))).toBe(false);
  });

  it("no BFF route opens raw TCP sockets", () => {
    const offenders = routeFiles(apiDir).filter((file) =>
      /from ["'](node:)?net["']/.test(readFileSync(file, "utf8")),
    );
    expect(offenders).toEqual([]);
  });
});
