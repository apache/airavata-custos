import { execSync } from "node:child_process";
import type { NextConfig } from "next";

function detectBuildSha(): string {
  if (process.env.NEXT_PUBLIC_PORTAL_BUILD_SHA) return process.env.NEXT_PUBLIC_PORTAL_BUILD_SHA;
  try {
    return execSync("git rev-parse --short HEAD", {
      stdio: ["ignore", "pipe", "ignore"],
    })
      .toString()
      .trim();
  } catch {
    return "dev";
  }
}

const nextConfig: NextConfig = {
  env: {
    NEXT_PUBLIC_PORTAL_BUILD_SHA: detectBuildSha(),
  },
};

export default nextConfig;
