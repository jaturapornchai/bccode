import { describe, expect, it } from "vitest";
// Same matcher Next uses for rewrite sources (path-to-regexp, case-insensitive).
import { getPathMatch } from "next/dist/shared/lib/router/utils/path-match";
import nextConfig from "../../next.config";

type Rewrite = { source: string; destination: string };

async function blockedSources(): Promise<string[]> {
  const rewrites = (await nextConfig.rewrites?.()) as { beforeFiles: Rewrite[] };
  return rewrites.beforeFiles
    .filter((rule) => rule.destination === "/_blocked-auth-route")
    .map((rule) => rule.source);
}

function isBlocked(sources: string[], path: string): boolean {
  return sources.some((source) => getPathMatch(source, { strict: true })(path) !== false);
}

describe("next.config /backend/* blocklist", () => {
  // mainapi trusts the client-supplied LINE code + lineuserid; only the BFF (/api/auth/line/*,
  // server-side) may reach these, so the public /backend/* proxy must not.
  it("blocks every LINE link endpoint on the public /backend proxy", async () => {
    const sources = await blockedSources();
    for (const path of [
      "/backend/profile/link-line",
      "/backend/profile/link-line/code",
      "/backend/profile/link-line/code/check",
      "/backend/Profile/Link-Line/code",
    ]) {
      expect(isBlocked(sources, path), path).toBe(true);
    }
  });

  it("keeps ordinary profile routes reachable", async () => {
    const sources = await blockedSources();
    for (const path of ["/backend/profile", "/backend/profile/password", "/backend/profile/link-liner"]) {
      expect(isBlocked(sources, path), path).toBe(false);
    }
  });
});
