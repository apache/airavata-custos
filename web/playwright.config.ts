import { defineConfig, devices } from "@playwright/test";

const port = Number(process.env.PORT ?? 3217);
const baseURL = `http://localhost:${port}`;

export default defineConfig({
  testDir: "./tests",
  testMatch: /.*\.e2e\.ts/,
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: "line",
  use: {
    baseURL,
    trace: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  webServer: {
    command: `pnpm dev --port ${port}`,
    url: baseURL,
    reuseExistingServer: false,
    timeout: 120_000,
    env: {
      PORT: String(port),
      NEXT_PUBLIC_PORTAL_USE_MSW: "true",
      // 32-char filler so the schema accepts boot; MSW intercepts every
      // /api/v1 call, so OIDC values are placeholders the IdP never sees.
      NEXTAUTH_SECRET: "test-secret-test-secret-test-secret",
      NEXTAUTH_URL: baseURL,
      OIDC_ISSUER_URL: "https://issuer.test",
      OIDC_CLIENT_ID: "test-client",
      OIDC_CLIENT_SECRET: "test-secret",
    },
  },
});
