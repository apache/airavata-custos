import { defineConfig, devices } from "@playwright/test";

const baseURL = process.env.LIVE_PORTAL_URL ?? "http://localhost:3001";

export default defineConfig({
  testDir: "./tests",
  testMatch: /.*\.spec\.ts/,
  fullyParallel: false,
  forbidOnly: false,
  retries: 0,
  workers: 1,
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
});
