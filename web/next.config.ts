import type { NextConfig } from "next";

const API_INTERNAL = process.env.API_INTERNAL_URL || "http://127.0.0.1:8080";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  poweredByHeader: false,
  output: "standalone",
  // Прокси /api/* на Go-API (на сервере — напрямую, на клиенте — через этот же путь).
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${API_INTERNAL}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
