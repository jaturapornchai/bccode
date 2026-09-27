import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // อนุญาตให้เครื่องอื่นใน Tailscale เรียก dev server นี้ได้ (กัน cross-origin block ของ Next dev)
  allowedDevOrigins: ["100.118.122.7"],
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
    const localBackendUrl = process.env.BCAI_LOCAL_BACKEND_URL?.trim() || (process.env.NODE_ENV === "production" ? "http://mainapi:8888" : "http://localhost:8888");
    // Public /backend/* proxy = explicit ALLOWLIST (ADR docs/kms/decisions/2026-09-27-backend-proxy-allowlist.md).
    // It used to be a "/backend/:path*" catch-all guarded by a blocklist of literal paths, which did not hold:
    // mainapi registers every route a second time under /v1, so /backend/v1/profile/link-line reached that
    // handler for any logged-in user; URL-encoded paths (/backend/goapi/%67et) slipped past the matcher and
    // reached mainapi (Echo routes on the raw path, so they ended in goapi auth/404, but the list proved leaky);
    // and routes mainapi itself treats as public (/metrics, /healthz, /goapi/version) were open to the internet.
    // Each entry is a route the browser really calls with a Bearer token and maps to ONE fixed mainapi path,
    // so a crafted suffix (..%2f, /v1, other casing) cannot land on a different mainapi route.
    // The `has` check only trims unauthenticated scanner noise; it is NOT the security boundary (mainapi
    // still verifies the token). Its value is a regex WITHOUT named groups on purpose: a bare `has` copies
    // the header into the rewrite params and Next then appends it to the query of param-less destinations
    // (?authorization=Bearer...), leaking the token into upstream URLs and logs.
    // Any other browser -> mainapi call must go through a BFF route under src/app/api that uses
    // serverMainApiBase()/serverGoApiBase() (src/lib/backend-url.ts) — do not widen this list without an ADR.
    const withBearer = [{ type: "header" as const, key: "authorization", value: "Bearer .+" }];
    return {
      afterFiles: [
        // Stored image/file URIs are "/goapi/s3/file/<key>" (authenticated-image.tsx, logo-avatar.tsx).
        { source: "/backend/goapi/s3/file/:path+", destination: `${localBackendUrl}/goapi/s3/file/:path+`, has: withBearer },
        // Company/branch management (company-branch-tree-view.tsx, currency-screen.tsx).
        { source: "/backend/organization/company", destination: `${localBackendUrl}/organization/company`, has: withBearer },
        { source: "/backend/organization/company/:id", destination: `${localBackendUrl}/organization/company/:id`, has: withBearer },
        { source: "/backend/organization/branch", destination: `${localBackendUrl}/organization/branch`, has: withBearer },
        { source: "/backend/organization/branch/:id", destination: `${localBackendUrl}/organization/branch/:id`, has: withBearer },
      ],
    };
  },
};


export default nextConfig;
