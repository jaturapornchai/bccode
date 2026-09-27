import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // อนุญาตให้เครื่องอื่นใน Tailscale เรียก dev server นี้ได้ (กัน cross-origin block ของ Next dev)
  allowedDevOrigins: ["100.118.122.7"],
  // No rewrites(): there is no public /backend proxy to mainapi. Every browser -> mainapi call is a BFF
  // route under src/app/api that uses serverMainApiBase()/serverGoApiBase() (src/lib/backend-url.ts) —
  // ADR docs/kms/decisions/2026-09-27-backend-proxy-allowlist.md; guarded by src/lib/next-config-rewrites.test.ts.
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
};


export default nextConfig;
