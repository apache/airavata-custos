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

test.describe("admin events", () => {
  // Make sure an admin can retry a failed delivery, and the retry shows up in its history.
  test("retrying a failed delivery adds the retry to its history", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/events?view=failed");
    await expect(page.getByRole("heading", { name: /^Events$/ })).toBeVisible({ timeout: 20_000 });

    await page.getByRole("button", { name: /Retry compute_cluster_user::approve/ }).click();
    const drawer = page.getByRole("dialog");
    await expect(drawer.getByRole("heading", { name: "Attempt history" })).toBeVisible();
    await drawer.getByRole("button", { name: "Retry now" }).click();

    await expect(page.getByText("Retry started")).toBeVisible();
    await expect(drawer.getByRole("region", { name: "Run 3" })).toBeVisible();
  });
});
