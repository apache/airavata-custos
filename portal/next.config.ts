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
