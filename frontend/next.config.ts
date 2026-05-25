import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  output: "standalone",
  async rewrites() {
    const localBackendUrl = process.env.BCAI_LOCAL_BACKEND_URL ?? "http://127.0.0.1:8888";
    return [
      {
        source: "/backend/:path*",
        destination: `${localBackendUrl}/:path*`,
      },
    ];
  },
};


export default nextConfig;
