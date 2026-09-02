import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          { key: "Referrer-Policy", value: "no-referrer-when-downgrade" },
          { key: "Cross-Origin-Opener-Policy", value: "same-origin-allow-popups" },
        ],
      },
    ];
  },
  async rewrites() {
    const localBackendUrl = process.env.BCAI_LOCAL_BACKEND_URL?.trim();
    if (!localBackendUrl) {
      throw new Error("BCAI_LOCAL_BACKEND_URL is required");
    }
    // The browser legitimately calls many AUTHENTICATED mainapi paths via this same-origin proxy
    // (e.g. /backend/organization/*, /backend/goapi/*, /backend/assets/*), all requiring a Bearer token.
    // So we proxy /backend/* broadly — but BLOCK the unauthenticated, identity-trusting token routes:
    // /googlelogin trusts the posted email with no verification, so exposing it publicly is an auth-bypass
    // (anyone could mint an owner token). These login/register flows are only ever called server-side
    // (Next /api/auth/* -> local backend), never by the browser, so blocking them here breaks nothing.
    const blockedAuthRoutes = [
      "/backend/login",
      "/backend/login/email",
      "/backend/login/phone-number",
      "/backend/login/line",
      "/backend/poslogin",
      "/backend/linelogin",
      "/backend/googlelogin",
      "/backend/dev-login",
      "/backend/v1/dev-login",
      "/backend/demo-login",
      "/backend/v1/demo-login",
      "/backend/tokenlogin",
      "/backend/register",
      "/backend/register-username",
      "/backend/register-phonenumber",
    ];
    // Defense-in-depth stopgap (2026-06-21 security audit): these backend routes are powerful and
    // were reachable from the public internet through this /backend/* catch-all. Verified that NO
    // browser code calls them — legitimate uses go server-side via Next /api/* (which fetches the
    // backend directly, NOT through this rewrite), so blocking the public /backend/* path here
    // breaks nothing while closing the audited holes:
    //  - /goapi/get,/exec,/getdoc : raw-SQL / raw-doc query handlers (SQLi + cross-tenant)
    //  - /reportm/*               : report query engine (NoSQL injection + cross-tenant playground)
    //  - /goapi/api/setup/*       : unauthenticated bootstrap.json secret dump (default pw "12345")
    //  - /goapi/api/mcp/*,/mcp/*  : MCP key mgmt + SSE/invoke + dev raw-SQL tools (weak keys, IDOR,
    //                               holdingcode injection). Re-open intentionally only after the MCP
    //                               holes are fixed. (LINE webhook stays public by design — LINE must
    //                               reach it; its fix is HMAC signature verification, not blocking.)
    const blockedDangerousRoutes = [
      "/backend/goapi/get",
      "/backend/goapi/exec",
      "/backend/goapi/getdoc",
      "/backend/reportm/:path*",
      "/backend/goapi/api/setup/:path*",
      "/backend/goapi/api/mcp/:path*",
      "/backend/goapi/mcp/:path*",
      "/backend/reload-config",
    ];
    return {
      beforeFiles: [...blockedAuthRoutes, ...blockedDangerousRoutes].map((source) => ({
        source,
        destination: "/_blocked-auth-route",
      })),
      afterFiles: [
        { source: "/backend/:path*", destination: `${localBackendUrl}/:path*` },
      ],
    };
  },
};


export default nextConfig;
