import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
// The same helpers Next uses at runtime for rewrites (server/lib/router-utils/resolve-routes.js):
// request URL parsing, source matching (path-to-regexp, case-insensitive), the `has` check and
// destination compilation. Testing through them proves what Next would really proxy.
import { parseUrl } from "next/dist/shared/lib/router/utils/parse-url";
import { getPathMatch } from "next/dist/shared/lib/router/utils/path-match";
import { matchHas, prepareDestination } from "next/dist/shared/lib/router/utils/prepare-destination";
import nextConfig from "../../next.config";

type RouteHas = { type: string; key?: string; value?: string };
type Rewrite = { source: string; destination: string; has?: RouteHas[] };
type RewriteGroups = { beforeFiles?: Rewrite[]; afterFiles?: Rewrite[]; fallback?: Rewrite[] };
type RequestHeaders = Record<string, string>;
type Resolved = { url: string; query: Record<string, unknown> };

const localBackend = "http://mainapi.test:8888";
const withBearer: RequestHeaders = { authorization: "Bearer test-token" };

beforeEach(() => {
  // rewrites() reads BCAI_LOCAL_BACKEND_URL when called (Next bakes it in at build time).
  vi.stubEnv("BCAI_LOCAL_BACKEND_URL", localBackend);
});

afterEach(() => {
  vi.unstubAllEnvs();
});

async function loadRewrites(): Promise<RewriteGroups> {
  return (await nextConfig.rewrites?.()) as RewriteGroups;
}

// Every afterFiles rule, not only "/backend*" ones: a new road to mainapi under another prefix
// (e.g. "/goapi/:path*" or a root catch-all) must also fail the exact source list below.
async function backendRules(): Promise<Rewrite[]> {
  const rewrites = await loadRewrites();
  return rewrites.afterFiles ?? [];
}

// Mirrors resolve-routes: normalise the request URL, then collect every rule whose source AND `has` match.
async function resolve(rawPath: string, headers: RequestHeaders = withBearer): Promise<Resolved[]> {
  const parsed = parseUrl(rawPath);
  const results: Resolved[] = [];
  for (const rule of await backendRules()) {
    const params = getPathMatch(rule.source, { strict: true, removeUnnamedParams: true })(parsed.pathname);
    if (params === false) continue;
    const hasParams = matchHas({ headers } as never, parsed.query, rule.has as never);
    if (hasParams === false) continue;
    const { parsedDestination } = prepareDestination({
      appendParamsToQuery: true,
      destination: rule.destination,
      params: { ...params, ...hasParams },
      query: parsed.query,
    });
    results.push({
      url: `${parsedDestination.origin}${parsedDestination.pathname}`,
      query: parsedDestination.query as Record<string, unknown>,
    });
  }
  return results;
}

describe("next.config public /backend proxy allowlist", () => {
  it("has no catch-all and no blocklist: the only rewrites are exact /backend -> mainapi paths behind a Bearer check", async () => {
    const rewrites = await loadRewrites();
    expect(rewrites.beforeFiles ?? []).toEqual([]);
    expect(rewrites.fallback ?? []).toEqual([]);

    const rules = await backendRules();
    expect(rules.map((rule) => rule.source)).toEqual([
      "/backend/goapi/s3/file/:path+",
      "/backend/organization/company",
      "/backend/organization/company/:id",
      "/backend/organization/branch",
      "/backend/organization/branch/:id",
    ]);
    for (const rule of rules) {
      expect(rule.source, rule.source).not.toMatch(/\*/);
      // Destination = the same path minus "/backend": a crafted suffix cannot land on another mainapi route.
      expect(rule.destination, rule.source).toBe(`${localBackend}${rule.source.slice("/backend".length)}`);
      expect(rule.has, rule.source).toEqual([{ type: "header", key: "authorization", value: "Bearer .+" }]);
    }
  });

  it("proxies exactly the routes the browser calls, each to one fixed destination", async () => {
    const cases: Array<[string, string]> = [
      ["/backend/goapi/s3/file/tenant/a.png", "/goapi/s3/file/tenant/a.png"],
      ["/backend/goapi/s3/file/a/b/c.webp", "/goapi/s3/file/a/b/c.webp"],
      ["/backend/organization/company", "/organization/company"],
      ["/backend/organization/company/01", "/organization/company/01"],
      ["/backend/organization/branch", "/organization/branch"],
      ["/backend/organization/branch/00000", "/organization/branch/00000"],
    ];
    for (const [path, destination] of cases) {
      const resolved = await resolve(path);
      expect(resolved.map((result) => result.url), path).toEqual([`${localBackend}${destination}`]);
    }
  });

  it("passes the query string through but never copies the Authorization header into it", async () => {
    for (const path of [
      "/backend/organization/company?management=true",
      "/backend/goapi/s3/file/tenant/a.png?variant=thumbnail",
    ]) {
      const [result] = await resolve(path);
      expect(result, path).toBeDefined();
      expect(Object.keys(result.query), path).not.toContain("authorization");
      expect(JSON.stringify(result.query), path).not.toContain("test-token");
    }
    const [company] = await resolve("/backend/organization/company?management=true");
    expect(company.query).toEqual({ management: "true" });
  });

  it("does not proxy anything without a Bearer Authorization header", async () => {
    expect(await resolve("/backend/organization/company", {})).toEqual([]);
    expect(await resolve("/backend/goapi/s3/file/tenant/a.png", { authorization: "Basic dXNlcjpwYXNz" })).toEqual([]);
  });

  it("denies every other /backend path, including the /v1 alias and encoded bypasses of the old blocklist", async () => {
    for (const path of [
      "/backend",
      "/backend/",
      "/backend/v1/profile/link-line",
      "/backend/v1/profile/link-line/code",
      "/backend/profile/link-line",
      "/backend/v1/login",
      "/backend/login",
      "/backend/googlelogin",
      "/backend/goapi/%67et",
      "/backend/goapi/get",
      "/backend/metrics",
      "/backend/healthz",
      "/backend/mcp/gl",
      "/backend/integration/gl/v2/accounts",
      "/backend/gl/v2/accounts",
      "/backend/v1/gl/v2/accounts",
      "/backend/mcp-tokens",
      "/backend/goapi/api/language/th",
      "/backend/api/language/th",
      "/backend/goapi/api/report/tax/wht",
      "/backend/goapi/image/upload",
      "/backend/holding/permission",
      "/backend/warehouse/tree",
      "/backend/organization/company/01/extra",
      "/backend/organization/department",
      "/backend/v1/organization/company",
      "/backend/goapi/s3/file",
      // Literal and %2e dot segments are resolved by Next's URL parsing BEFORE matching,
      // so they cannot climb out of an allowed prefix (both become /backend/goapi/get).
      "/backend/goapi/s3/file/../../get",
      "/backend/goapi/s3/file/%2e%2e/%2e%2e/get",
      // path-to-regexp in Next is case-insensitive: other casings of denied paths still match nothing.
      "/BACKEND/Metrics",
      "/Backend/V1/Login",
      "/Backend/GoApi/S3/File",
    ]) {
      expect(await resolve(path), path).toEqual([]);
    }
  });

  it("keeps a case-variant of an allowed path on the same fixed lowercase destination", async () => {
    // Case-insensitive matching can only select an allowlisted rule; the destination is fixed text.
    const resolved = await resolve("/BACKEND/ORGANIZATION/COMPANY");
    expect(resolved.map((result) => result.url)).toEqual([`${localBackend}/organization/company`]);
  });

  it("keeps an encoded ..%2f inside the allowed route prefix (one segment, fixed destination prefix)", async () => {
    expect((await resolve("/backend/organization/company/..%2f..%2fv1%2flogin")).map((result) => result.url)).toEqual([
      `${localBackend}/organization/company/..%2f..%2fv1%2flogin`,
    ]);
    expect((await resolve("/backend/goapi/s3/file/..%2f..%2fget")).map((result) => result.url)).toEqual([
      `${localBackend}/goapi/s3/file/..%2f..%2fget`,
    ]);
  });
});
