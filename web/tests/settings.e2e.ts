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

test.describe("settings", () => {
  test("account menu opens the settings page and theme persists across reload", async ({
    page,
  }) => {
    await signInAs(page, "admin");
    await page.goto("/");

    // Open the account menu and navigate to Settings.
    await page.getByRole("button", { name: /account menu/i }).click();
    await page.getByRole("menuitem", { name: /settings/i }).click();
    await expect(page).toHaveURL(/\/settings$/);
    await expect(page.getByRole("heading", { name: /^Settings$/ })).toBeVisible();

    // The identity cards render against the seeded MSW fixtures.
    await expect(page.getByRole("heading", { name: "Roles" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Effective privileges" })).toBeVisible();

    // Switch to the dark theme and confirm it applies.
    await page.getByRole("button", { name: /^dark$/i }).click();
    await expect(page.locator("html")).toHaveClass(/dark/);

    // Reload and confirm the choice persisted (next-themes localStorage).
    await page.reload();
    await expect(page.getByRole("button", { name: /^dark$/i })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    await expect(page.locator("html")).toHaveClass(/dark/);
  });
});
