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

// Closeout-only live walk: one-off verification against a running backend.
// Excluded from the default `pnpm test:e2e` suite via testMatch in
// playwright.config.ts (this file uses .spec.ts, not .e2e.ts).
import { type Page, expect, test } from "@playwright/test";

const BASE = process.env.LIVE_PORTAL_URL ?? "http://localhost:3001";

async function signInAdmin(page: Page) {
  await page.goto(`${BASE}/sign-in`);
  await page.getByLabel(/dev privilege level/i).selectOption("admin");
  await page.getByRole("button", { name: /^sign in$/i }).click();
  await page.waitForURL((url) => !url.pathname.startsWith("/sign-in"), { timeout: 20_000 });
}

async function gotoAndAssertNoRuntimeError(page: Page, path: string) {
  const pageErrors: string[] = [];
  page.on("pageerror", (err) => pageErrors.push(err.message));
  const consoleErrors: string[] = [];
  page.on("console", (msg) => {
    if (msg.type() === "error") consoleErrors.push(msg.text());
  });

  await page.goto(`${BASE}${path}`, { waitUntil: "domcontentloaded" });
  // Give React Query a beat to settle before asserting.
  await page.waitForLoadState("networkidle", { timeout: 15_000 }).catch(() => {});

  // No uncaught runtime exceptions in the page.
  expect(pageErrors, pageErrors.join("\n")).toEqual([]);
  // Next.js renders a visible error boundary card if a route component throws.
  await expect(page.getByText(/Application error: a client-side exception/i)).toHaveCount(0);

  return { consoleErrors };
}

test.describe("closeout live walk", () => {
  test.describe.configure({ mode: "serial" });

  test("projects renders against live backend", async ({ page }) => {
    await signInAdmin(page);
    await gotoAndAssertNoRuntimeError(page, "/projects");
    await expect(page.getByRole("heading", { name: /^Projects$/ })).toBeVisible();
  });

  test("allocations renders against live backend", async ({ page }) => {
    await signInAdmin(page);
    await gotoAndAssertNoRuntimeError(page, "/allocations");
    await expect(page.getByRole("heading", { name: /^Allocations$/ })).toBeVisible();
  });

  test("admin/traces renders against live backend (graceful empty/404)", async ({ page }) => {
    await signInAdmin(page);
    await gotoAndAssertNoRuntimeError(page, "/admin/traces");
    // Page mounts even when backend has no /audit/* routes.
    await expect(page.locator("main")).toBeVisible();
  });

  test("admin/amie renders against live backend (graceful empty/404)", async ({ page }) => {
    await signInAdmin(page);
    await gotoAndAssertNoRuntimeError(page, "/admin/amie/packets");
    await expect(page.locator("main")).toBeVisible();
  });
});
