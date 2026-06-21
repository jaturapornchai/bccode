import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  async rewrites() {
    // Default to the on-prem backend server. Override with BCAI_LOCAL_BACKEND_URL
    // for other environments. Note: next start only reads .env at build time,
    // so the fallback below must be the real default.
    const localBackendUrl = process.env.BCAI_LOCAL_BACKEND_URL ?? "http://192.168.2.202:8888";
    return [
      {
        source: "/backend/:path*",
        destination: `${localBackendUrl}/:path*`,
      },
    ];
  },
};


export default nextConfig;
