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

test.describe("admin agreements", () => {
  // Checks that an admin can open the published version from the list and see its Remove action.
  test("opens a version from the list", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/agreements");

    await expect(page.getByRole("heading", { name: /^Agreements$/ })).toBeVisible({
      timeout: 20_000,
    });

    await page.getByRole("link", { name: "1.0", exact: true }).click();
    await expect(page.getByText("Version 1.0")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("button", { name: "Remove" })).toBeVisible();
  });
});
