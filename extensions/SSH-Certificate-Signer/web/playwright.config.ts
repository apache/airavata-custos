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

const portalPort = Number(process.env.PORT ?? 3216);
const zonePort = 3217;
const manifestPort = 3218;
const baseURL = `http://localhost:${portalPort}`;
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
  webServer: [
    {
      command: "node tests/fixtures/manifest-server.mjs",
      url: `http://localhost:${manifestPort}/.well-known/custos-extension.json`,
      reuseExistingServer: false,
      timeout: 30_000,
    },
    {
      command: `corepack pnpm dev --port ${zonePort}`,
      url: `http://localhost:${zonePort}/signer/api/v1/health`,
      reuseExistingServer: false,
      timeout: 120_000,
      env: {
        NEXTAUTH_SECRET: sharedSecret,
        CUSTOS_SIGNER_API_BASE_URL: "http://127.0.0.1:1",
        CUSTOS_PORTAL_BASE_URL: baseURL,
      },
    },
    {
      command: `node tests/fixtures/wait-for-manifest.mjs && corepack pnpm --dir ../../../web dev --port ${portalPort}`,
      url: baseURL,
      reuseExistingServer: false,
      timeout: 120_000,
      env: {
        NEXTAUTH_SECRET: sharedSecret,
        NEXTAUTH_URL: baseURL,
        OIDC_ISSUER_URL: "http://localhost:8081/realms/custos",
        OIDC_CLIENT_ID: "playwright-cookie-fixture",
        OIDC_CLIENT_SECRET: "playwright-cookie-fixture",
        CUSTOS_EXTENSION_MANIFEST_URLS: `http://localhost:${manifestPort}/.well-known/custos-extension.json`,
      },
    },
  ],
});

process.env.NEXTAUTH_SECRET = sharedSecret;
process.env.PORT = String(portalPort);
