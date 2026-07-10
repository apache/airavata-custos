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

test.describe("admin traces detail drawer", () => {
  test("clicking a row opens the drawer, reload persists it, Esc closes it", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");
    await expect(page.getByRole("heading", { name: /^Traces$/ })).toBeVisible({
      timeout: 20_000,
    });

    const firstRow = page.locator('[data-testid^="trace-row-"]').first();
    await expect(firstRow).toBeVisible({ timeout: 20_000 });
    await firstRow.click();

    await expect(page).toHaveURL(/[?&]trace=[0-9a-f]{32}\b/, { timeout: 10_000 });
    await expect(page.getByTestId("trace-detail-drawer")).toBeVisible({ timeout: 10_000 });

    await page.reload();
    await expect(page.getByTestId("trace-detail-drawer")).toBeVisible({ timeout: 20_000 });

    await page.keyboard.press("Escape");
    await expect(page.getByTestId("trace-detail-drawer")).toHaveCount(0, { timeout: 10_000 });
    await expect(page).not.toHaveURL(/[?&]trace=[0-9a-f]{32}\b/);
  });

  test("Tree is the default tab and renders the span tree", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");
    await expect(page.getByRole("heading", { name: /^Traces$/ })).toBeVisible({
      timeout: 20_000,
    });
    await page.locator('[data-testid^="trace-row-"]').first().click();
    await expect(page.getByTestId("trace-detail-drawer")).toBeVisible({ timeout: 10_000 });

    await expect(page.getByRole("tree", { name: /trace span tree/i })).toBeVisible({
      timeout: 10_000,
    });
  });

  test("Tree tab on a failed trace shows the error chip", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");
    await expect(page.getByRole("heading", { name: /^Traces$/ })).toBeVisible({
      timeout: 20_000,
    });

    await page.locator('[data-testid^="trace-row-"][data-tone="error"]').first().click();
    await expect(page.getByTestId("trace-detail-drawer")).toBeVisible({ timeout: 10_000 });
    await expect(page.getByTestId("trace-error-chip")).toBeVisible({ timeout: 10_000 });
  });

  test("switching tabs surfaces Overview, Raw, and Linked entities content", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");
    await expect(page.getByRole("heading", { name: /^Traces$/ })).toBeVisible({
      timeout: 20_000,
    });
    await page.locator('[data-testid^="trace-row-"]').first().click();
    await expect(page.getByTestId("trace-detail-drawer")).toBeVisible({ timeout: 10_000 });

    const drawer = page.getByTestId("trace-detail-drawer");

    await page.getByRole("tab", { name: /^Overview$/ }).click();
    await expect(drawer.getByText(/^Trace ID$/)).toBeVisible();
    await expect(drawer.getByText(/^Span count$/)).toBeVisible();

    await page.getByRole("tab", { name: /^Raw$/ }).click();
    await expect(drawer.getByRole("button", { name: /copy trace JSON/i })).toBeVisible();

    await page.getByRole("tab", { name: /^Linked entities$/ }).click();
    await expect(drawer.getByText(/Entities referenced by spans/)).toBeVisible();
  });

  test("Retry button is aria-disabled and surfaces the coming-soon tooltip", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");
    await expect(page.getByRole("heading", { name: /^Traces$/ })).toBeVisible({
      timeout: 20_000,
    });
    await page.locator('[data-testid^="trace-row-"]').first().click();
    await expect(page.getByTestId("trace-detail-drawer")).toBeVisible({ timeout: 10_000 });

    const retry = page.getByTestId("retry-tooltip-anchor");
    await expect(retry).toHaveAttribute("aria-disabled", "true");
    await retry.hover();
    await expect(page.getByText(/Retry coming soon/i)).toBeVisible({ timeout: 5_000 });
  });

  test("axe: no serious or critical violations on the drawer-open page", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");
    await expect(page.getByRole("heading", { name: /^Traces$/ })).toBeVisible({
      timeout: 20_000,
    });
    await page.locator('[data-testid^="trace-row-"]').first().click();
    await expect(page.getByTestId("trace-detail-drawer")).toBeVisible({ timeout: 10_000 });
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
