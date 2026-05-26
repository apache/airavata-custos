// Next.js configuration. `allowedDevOrigins` permits both localhost and
// 127.0.0.1 (over the portal's dev port 5173) to suppress the App Router's
// cross-origin warning when contributors reach the dev server via either
// hostname.
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  allowedDevOrigins: [
    "http://127.0.0.1:5173",
    "http://localhost:5173",
    "127.0.0.1",
    "localhost",
  ],
};

export default nextConfig;
