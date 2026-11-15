import type { NextConfig } from "next";

const backendURL = process.env.BACKEND_URL ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  async rewrites() {
    return [{ source: "/api/auth/:path*", destination: `${backendURL}/auth/:path*` }, { source: "/avatars/:path*", destination: `${backendURL}/avatars/:path*` }, { source: "/files/:path*", destination: `${backendURL}/files/:path*` }];
  },
};

export default nextConfig;
