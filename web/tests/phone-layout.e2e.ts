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

import { expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

test.use({ viewport: { width: 390, height: 844 } });

test.describe("phone layout", () => {
  test("keeps the account menu, navigation and membership Save on screen", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/allocations/01000004-0000-4000-8000-000001000004");
    await expect(page.getByRole("heading", { name: /Catalyst Discovery - Legacy/i })).toBeVisible();

    await expect(page.getByRole("button", { name: "Account menu" })).toBeInViewport({ ratio: 1 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
      390,
    );

    await page.getByRole("button", { name: "Open navigation" }).click();
    await page
      .getByRole("dialog", { name: "Navigation" })
      .getByRole("link", { name: "Projects" })
      .click();
    await expect(page).toHaveURL(/\/projects$/, { timeout: 15_000 });
    await expect(page.getByRole("dialog", { name: "Navigation" })).toBeHidden();

    await page.goto("/allocations/01000004-0000-4000-8000-000001000004");
    await page.getByRole("tab", { name: /^members$/i }).click();
    await page.getByRole("button", { name: /\+ add member/i }).click();
    await page.getByRole("radio", { name: /Elena Vasquez/ }).check();
    await page.getByRole("button", { name: /^add member$/i }).click();
    await page.getByRole("row").filter({ hasText: "3a8e1f64-9c2b-4d57-b0e3-6f4a2c8d1b95" }).click();
    await expect(page.getByRole("dialog").getByRole("button", { name: "Save" })).toBeInViewport({
      ratio: 1,
    });
  });
});
