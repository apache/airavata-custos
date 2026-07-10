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
      NEXTAUTH_SECRET: sharedSecret,
      NEXTAUTH_URL: baseURL,
      // Required by env schema. Tests inject session cookies directly so this
      // OIDC config is never actually invoked — the real OIDC flow is in the
      // live config against the compose Keycloak.
      OIDC_ISSUER_URL: "http://localhost:8081/realms/custos",
      OIDC_CLIENT_ID: "playwright-cookie-fixture",
      OIDC_CLIENT_SECRET: "playwright-cookie-fixture",
    },
  },
});

// The fixture reads NEXTAUTH_SECRET from process.env, but Playwright workers
// run in their own process — propagate it explicitly here.
process.env.NEXTAUTH_SECRET = sharedSecret;
