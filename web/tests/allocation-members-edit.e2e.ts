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

test.describe("allocation members tab", () => {
  test("admin adds and removes a member", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/allocations/alloc-004");
    await expect(
      page.getByRole("heading", { name: /Catalyst Discovery - Legacy/i }),
    ).toBeVisible();

    await page.getByRole("tab", { name: /^members$/i }).click();

    // Add
    await page.getByRole("button", { name: /\+ add member/i }).click();
    await expect(page.getByRole("heading", { name: /^Add member$/i })).toBeVisible();
    await page.getByLabel(/user id/i).fill("user-new");
    await page.getByRole("button", { name: /^add member$/i }).click();
    await expect(page.getByText(/user-new/i).first()).toBeVisible({ timeout: 15_000 });

    // Remove
    const removeButtons = page.getByRole("button", { name: /^remove /i });
    const initialCount = await removeButtons.count();
    expect(initialCount).toBeGreaterThan(0);
    await removeButtons.first().click();
    await page.getByRole("button", { name: "Remove", exact: true }).click();
    await expect
      .poll(async () => removeButtons.count(), { timeout: 15_000 })
      .toBeLessThan(initialCount);
  });
});
