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

test.describe("projects list", () => {
  test("renders the list and navigates to a detail page", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects");
    await expect(page.getByRole("heading", { name: /^Projects$/ })).toBeVisible();

    const firstRow = page.locator('a[href^="/projects/project-"]').first();
    await expect(firstRow).toBeVisible({ timeout: 15_000 });
    await firstRow.click();
    await expect(page).toHaveURL(/\/projects\/project-/);
  });

  test("status filter updates URL state", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects");
    await page.getByLabel(/filter by status/i).selectOption("INACTIVE");
    await expect(page).toHaveURL(/[?&]status=INACTIVE/);
  });

  test("axe sweep on projects list", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects");
    await expect(page.getByRole("heading", { name: /^Projects$/ })).toBeVisible();
    const results = await new AxeBuilder({ page })
      .withTags(["wcag2a", "wcag2aa"])
      .analyze();
    const blocking = results.violations.filter(
      (v) => v.impact === "serious" || v.impact === "critical",
    );
    expect(blocking, JSON.stringify(blocking, null, 2)).toEqual([]);
  });
});
