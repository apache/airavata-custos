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

    const approveBtn = page.getByRole("button", { name: /^Approve cr-001$/ });
    await expect(approveBtn).toBeVisible({ timeout: 15_000 });
    await approveBtn.click();
    await expect(page.getByText(/approved cr-001/i)).toBeVisible({ timeout: 15_000 });
  });

  test("admin can reject a pending request", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/change-requests");
    await expect(page.getByRole("heading", { name: /^Change requests$/ })).toBeVisible();

    const rejectBtn = page.getByRole("button", { name: /^Reject cr-002$/ });
    await expect(rejectBtn).toBeVisible({ timeout: 15_000 });
    await rejectBtn.click();
    await expect(page.getByText(/rejected cr-002/i)).toBeVisible({ timeout: 15_000 });
  });
});
