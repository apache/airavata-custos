// Playwright end-to-end test configuration. Tests spawn their own Next
// dev server on a dedicated port (3105) to avoid colliding with a running
// `npm run dev` on 5173. Only files matching *.e2e.ts under tests/ are
// picked up.
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  testMatch: "**/*.e2e.ts",
  timeout: 30_000,
  use: {
    baseURL: "http://127.0.0.1:3105",
    trace: "on-first-retry",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
