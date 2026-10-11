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

test.describe("admin tracing list", () => {
  // The banner is the on-call entry point: it counts failed traces and one
  // click narrows the list to them.
  test("banner counts failed traces and Show them applies the Failed filter", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");
    await expect(page.getByRole("heading", { name: /^Tracing$/ })).toBeVisible({
      timeout: 20_000,
    });

    const banner = page.getByTestId("failing-banner");
    await expect(banner).toContainText(/traces need attention/, { timeout: 20_000 });
    await banner.getByRole("button", { name: /Show them/ }).click();

    await expect(page).toHaveURL(/[?&]status=failed\b/, { timeout: 10_000 });
    await expect(page.getByRole("button", { name: "Failed", pressed: true })).toBeVisible();
    await expect(
      page.locator('[data-testid^="trace-row-"][data-status="failed"]').first(),
    ).toBeVisible();
    await expect(page.locator('[data-testid^="trace-row-"][data-status="done"]')).toHaveCount(0);
  });

  test("axe: no serious or critical violations on the list page", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/traces");
    await expect(page.getByRole("heading", { name: /^Tracing$/ })).toBeVisible({
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
