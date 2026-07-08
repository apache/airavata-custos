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

test.describe("admin resources & clusters read-only sweep", () => {
  test("clusters tab opens a drawer with local users", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/resources");
    await expect(page.getByRole("heading", { name: /Resources & Clusters/ })).toBeVisible({
      timeout: 20_000,
    });

    // Clusters is the default tab.
    await page.getByRole("button", { name: "Anvil", exact: true }).click();
    await expect(page.getByText("Cluster: Anvil")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("ghopper")).toBeVisible();
  });

  test("resources tab opens a rates drawer with rate rows", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/resources");
    await expect(page.getByRole("heading", { name: /Resources & Clusters/ })).toBeVisible({
      timeout: 20_000,
    });

    await page.getByRole("tab", { name: /^resources$/i }).click();
    await page
      .getByRole("row", { name: /Anvil CPU/ })
      .getByRole("button", { name: "Rates" })
      .click();

    await expect(page.getByText("Rates: Anvil CPU")).toBeVisible({ timeout: 15_000 });
    // History carries the current rate plus the superseded one (client-derived).
    await expect(page.getByText("EXPIRED")).toBeVisible();
    await expect(page.getByText("0.04")).toBeVisible();
  });
});
