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

test.describe("change request approver queue", () => {
  test("admin sees the queue and can approve a pending request", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/change-requests");
    await expect(page.getByRole("heading", { name: /^Change requests$/ })).toBeVisible();

    const approveBtn = page.getByRole("button", {
      name: /^Approve change request .* request 05000002$/,
    });
    await expect(approveBtn).toBeVisible({ timeout: 15_000 });
    // Request 05000001 belongs to an ACCESS project; origination never hides a decision.
    await expect(
      page.getByRole("button", {
        name: /^Approve change request .* request 05000001$/,
      }),
    ).toBeVisible();
    await approveBtn.click();
    await expect(page.getByText("Approved the change request")).toBeVisible({ timeout: 15_000 });
  });

  test("admin can reject a pending request", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/change-requests");
    await expect(page.getByRole("heading", { name: /^Change requests$/ })).toBeVisible();

    const rejectBtn = page.getByRole("button", {
      name: /^Reject change request .* request 05000002$/,
    });
    await expect(rejectBtn).toBeVisible({ timeout: 15_000 });
    await rejectBtn.click();
    await expect(page.getByText("Rejected the change request")).toBeVisible({ timeout: 15_000 });
  });
});
