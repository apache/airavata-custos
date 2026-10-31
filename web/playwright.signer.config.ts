// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

import { defineConfig, devices } from "@playwright/test";

const port = Number(process.env.PORT ?? 3217);
const baseURL = `http://localhost:${port}`;
const sharedSecret = "test-secret-for-playwright-cookie-fixture-only-32chars";

export default defineConfig({
  testDir: "./tests",
  testMatch: "signer-certificates.e2e.ts",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: "line",
  use: {
    baseURL,
    trace: "retain-on-failure",
    // The signer suite installs deterministic Playwright routes. Blocking
    // service workers ensures browser MSW cannot consume or bypass a request.
    serviceWorkers: "block",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  webServer: {
    command: `corepack pnpm dev --port ${port}`,
    url: baseURL,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    env: {
      PORT: String(port),
      NEXT_PUBLIC_PORTAL_USE_MSW: "false",
      CUSTOS_E2E_FAIL_ON_SIGNER_PROXY_REQUEST: "true",
      NEXTAUTH_SECRET: sharedSecret,
      NEXTAUTH_URL: baseURL,
      OIDC_ISSUER_URL: "http://localhost:8081/realms/custos",
      OIDC_CLIENT_ID: "playwright-cookie-fixture",
      OIDC_CLIENT_SECRET: "playwright-cookie-fixture",
    },
  },
});

process.env.NEXTAUTH_SECRET = sharedSecret;
