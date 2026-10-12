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

test.describe("admin AMIE packet detail", () => {
  test("clicking a row opens the drawer with its three tabs", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/amie/packets");
    await expect(page.getByRole("heading", { name: "AMIE", level: 1 })).toBeVisible({
      timeout: 20_000,
    });

    const firstRow = page.locator("tbody tr").first();
    await expect(firstRow).toBeVisible({ timeout: 15_000 });
    await firstRow.click();

    await expect(page.getByRole("tab", { name: /Overview/i })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("tab", { name: /Raw JSON/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Timeline/i })).toBeVisible();
  });

  test("a Failed-tab drawer is kept in ?packet and reopens on reload", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/amie/failed");
    const row = page.locator("tbody tr", { hasText: "1234600" });
    await expect(row).toBeVisible({ timeout: 20_000 });
    await row.click();
    await expect(page).toHaveURL(
      /\/admin\/amie\/failed\?packet=16000005-0000-4000-8000-000016000005$/,
    );
    await page.reload();
    await expect(page.getByRole("heading", { name: "Packet 1234600" })).toBeVisible({
      timeout: 20_000,
    });
  });

  test("Timeline tab shows event rows", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/amie/packets/16000005-0000-4000-8000-000016000005");

    await expect(page.getByRole("tab", { name: /Timeline/i })).toBeVisible({ timeout: 15_000 });
    await page.getByRole("tab", { name: /Timeline/i }).click();
    await expect(page.getByTestId("packet-events-table")).toBeVisible({ timeout: 10_000 });
  });

  test("axe: no serious or critical violations on the open drawer", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/amie/packets/16000005-0000-4000-8000-000016000005");
    await expect(page.getByRole("tab", { name: /Overview/i })).toBeVisible({ timeout: 20_000 });
    await page.waitForLoadState("networkidle");
    const results = await new AxeBuilder({ page })
      .include("main")
      .withTags(["wcag2a", "wcag2aa"])
      .analyze();
    const critical = results.violations.filter((v) => SEVERITIES.includes(v.impact as "serious"));
    expect(critical).toEqual([]);
  });
});
