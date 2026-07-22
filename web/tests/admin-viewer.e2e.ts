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

test.describe("admin viewer persona", () => {
  test("sees read-only surfaces with no mutation UI or dropped nav", async ({ page }) => {
    await signInAs(page, "member");
    await page.goto("/admin/organizations");

    await expect(page.getByRole("heading", { name: /^Organizations$/ })).toBeVisible({
      timeout: 20_000,
    });
    // Read-only viewer lacks manage Organization, so no create affordance.
    await expect(page.getByRole("button", { name: "+ Create organization" })).toHaveCount(0);

    // The dropped areas have no nav entry for anyone now.
    await expect(page.getByRole("link", { name: "Users" })).toHaveCount(0);
    await expect(page.getByRole("link", { name: "Roles & Privileges" })).toHaveCount(0);
    await expect(page.getByRole("link", { name: "Clusters" })).toHaveCount(0);
    // Read entries the viewer is entitled to remain.
    await expect(page.getByRole("link", { name: "Organizations" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Resources" })).toBeVisible();

    await page.goto("/admin/resources");
    await expect(page.getByRole("heading", { name: /Resources & Clusters/ })).toBeVisible();
    await expect(page.getByRole("tab", { name: /^clusters$/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /^resources$/i })).toBeVisible();

    // Viewer lacks manage Allocation, so the rates drawer offers no create action.
    await page.getByRole("tab", { name: /^resources$/i }).click();
    await page
      .getByRole("row", { name: /ClusterA CPU/ })
      .getByRole("button", { name: "Rates" })
      .click();
    await expect(page.getByText("Rates: ClusterA CPU")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("button", { name: "+ Add rate" })).toHaveCount(0);
  });
});
