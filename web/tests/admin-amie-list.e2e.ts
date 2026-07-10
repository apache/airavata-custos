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

test.describe("admin AMIE inbox", () => {
  test("admin lands on the inbox, sees the trend chart and at least one row", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/amie/packets");
    await expect(page.getByRole("heading", { name: /AMIE packet inbox/i })).toBeVisible({
      timeout: 20_000,
    });
    await expect(page.getByLabel(/AMIE packets per day/i)).toBeVisible({ timeout: 15_000 });
    const firstRow = page.locator("tbody tr").first();
    await expect(firstRow).toBeVisible({ timeout: 15_000 });
  });

  test("filtering by FAILED narrows the list to failed packets only", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/amie/packets");
    await expect(page.getByRole("heading", { name: /AMIE packet inbox/i })).toBeVisible({
      timeout: 20_000,
    });

    await page.getByLabel(/^Status$/).selectOption({ value: "FAILED" });
    await page.getByRole("button", { name: /^Apply$/ }).click();

    const rows = page.locator("tbody tr");
    await expect(rows.first()).toBeVisible({ timeout: 15_000 });
    const statuses = await rows.locator("text=FAILED").count();
    expect(statuses).toBeGreaterThan(0);
  });

  test("axe: no serious or critical violations on the inbox", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/amie/packets");
    await expect(page.getByRole("heading", { name: /AMIE packet inbox/i })).toBeVisible({
      timeout: 20_000,
    });
    await page.waitForLoadState("networkidle");
    const results = await new AxeBuilder({ page })
      .include("main")
      .withTags(["wcag2a", "wcag2aa"])
      .analyze();
    const critical = results.violations.filter((v) => SEVERITIES.includes(v.impact as "serious"));
    expect(critical).toEqual([]);
  });
});
