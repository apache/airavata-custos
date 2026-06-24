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

import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

const SEVERITIES = ["serious", "critical"] as const;

test.describe("admin traces list", () => {
  test("renders the page, default error filter, and at least one error row", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");

    await expect(page.getByRole("heading", { name: /^Traces$/ })).toBeVisible({
      timeout: 20_000,
    });

    await expect(page.getByRole("button", { name: /^error$/i })).toHaveAttribute(
      "aria-pressed",
      "true",
    );

    const errorRows = page.locator('[data-testid^="trace-row-"][data-tone="error"]');
    await expect(errorRows.first()).toBeVisible({ timeout: 20_000 });
  });

  test("toggling to the ok pill updates the URL and re-fetches", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");
    await expect(page.getByRole("heading", { name: /^Traces$/ })).toBeVisible({
      timeout: 20_000,
    });

    await page.getByRole("button", { name: /^error$/i }).click();
    await page.getByRole("button", { name: /^ok$/i }).click();

    await expect(page).toHaveURL(/[?&]status=ok\b/, { timeout: 10_000 });

    await expect(page.getByTestId("trace-table-loading")).toHaveCount(0, { timeout: 15_000 });
  });

  test("clicking a row pushes ?trace=<id>", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");
    await expect(page.getByRole("heading", { name: /^Traces$/ })).toBeVisible({
      timeout: 20_000,
    });

    const firstRow = page.locator('[data-testid^="trace-row-"]').first();
    await expect(firstRow).toBeVisible({ timeout: 20_000 });
    await firstRow.click();

    await expect(page).toHaveURL(/[?&]trace=[0-9a-f]{32}\b/, { timeout: 10_000 });
  });

  test("axe: no serious or critical violations on the list page", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");
    await expect(page.getByRole("heading", { name: /^Traces$/ })).toBeVisible({
      timeout: 20_000,
    });
    await page.waitForLoadState("networkidle");

    const results = await new AxeBuilder({ page })
      .options({ resultTypes: ["violations"] })
      .analyze();
    const blocking = results.violations.filter((v) =>
      SEVERITIES.includes((v.impact ?? "minor") as (typeof SEVERITIES)[number]),
    );
    expect(blocking, JSON.stringify(blocking, null, 2)).toEqual([]);
  });
});
