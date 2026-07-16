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

import { type Page, expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

// Seam read by the MSW /access-requests/me handler: the caller already has a
// denied request.
async function seedDeniedRequest(page: Page) {
  const port = process.env.PORT ?? "3217";
  await page.context().addCookies([
    {
      name: "custos.test-my-access-request",
      value: "DENIED",
      url: `http://localhost:${port}`,
      httpOnly: false,
      secure: false,
      sameSite: "Lax",
    },
  ]);
}

test.describe("access requests — requester", () => {
  test("entry state, event code validation, submit, pending state", async ({ page }) => {
    await signInAs(page, "viewer");
    await page.goto("/no-access");
    await expect(page.getByRole("heading", { name: /^No portal access$/ })).toBeVisible({
      timeout: 15_000,
    });

    await page.getByRole("button", { name: /request trial access/i }).click();
    await expect(page.getByRole("heading", { name: /request trial access/i })).toBeVisible();

    // Identity comes from the session and is not editable.
    await expect(page.getByLabel("Name")).toHaveValue("Test Viewer");
    await expect(page.getByLabel("Name")).toBeDisabled();
    await expect(page.getByLabel("Email")).toHaveValue("viewer@custos.local");
    await expect(page.getByLabel("Email")).toBeDisabled();

    await page.getByLabel("Institution").fill("Example University");
    await page.getByLabel("Event code").fill("NOPE99");
    await expect(page.getByText(/unknown event code/i)).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("button", { name: /submit request/i })).toBeDisabled();

    await page.getByLabel("Event code").fill("PEARC26");
    await expect(page.getByText(/event: pearc26-tutorial/i)).toBeVisible({ timeout: 15_000 });

    await page.getByRole("button", { name: /submit request/i }).click();
    await expect(page.getByRole("heading", { name: /request received/i })).toBeVisible({
      timeout: 15_000,
    });
  });

  test("?event= prefills the event code", async ({ page }) => {
    await signInAs(page, "viewer");
    await page.goto("/no-access?event=PEARC26");
    await page.getByRole("button", { name: /request trial access/i }).click();

    await expect(page.getByLabel("Event code")).toHaveValue("PEARC26");
    await expect(page.getByText(/event: pearc26-tutorial/i)).toBeVisible({ timeout: 15_000 });
  });

  test("declined state shows the deny reason and allows a new request", async ({ page }) => {
    await signInAs(page, "viewer");
    await seedDeniedRequest(page);
    await page.goto("/no-access");

    await expect(page.getByRole("heading", { name: /request declined/i })).toBeVisible({
      timeout: 15_000,
    });
    await expect(
      page.getByText("Could not verify event registration for this attendee."),
    ).toBeVisible();

    await page.getByRole("button", { name: /submit a new request/i }).click();
    await expect(page.getByRole("heading", { name: /request trial access/i })).toBeVisible();
  });
});

test.describe("access requests — admin queue", () => {
  test("queue defaults to pending; deny via popover records the reason", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/access-requests");
    await expect(page.getByRole("heading", { name: /^Access Requests$/ })).toBeVisible();

    // Default filter is pending: the seeded approved row stays out.
    await expect(page.getByRole("row", { name: /Amara Osei/ })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("row", { name: /Li Wei/ })).toBeVisible();
    await expect(page.getByRole("row", { name: /Marcus Rivera/ })).toHaveCount(0);

    await page
      .getByRole("row", { name: /Amara Osei/ })
      .getByRole("button", { name: /^Deny$/ })
      .click();
    await page.getByLabel(/reason \(optional\)/i).fill("Not a registered attendee.");
    await page.getByRole("button", { name: /deny request/i }).click();

    // The decided row leaves the pending view and lands under Denied.
    await expect(page.getByRole("row", { name: /Amara Osei/ })).toHaveCount(0, {
      timeout: 15_000,
    });
    await page.getByRole("button", { name: "Denied", exact: true }).click();
    const deniedRow = page.getByRole("row", { name: /Amara Osei/ });
    await expect(deniedRow).toBeVisible({ timeout: 15_000 });
    await expect(deniedRow.getByText("Denied", { exact: true })).toBeVisible();

    await deniedRow.click();
    await expect(page.getByText("Not a registered attendee.")).toBeVisible({ timeout: 15_000 });
  });

  test("approve shows an undo toast, persists after the window, drawer has the decision", async ({
    page,
  }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/access-requests");
    const pendingRow = page.getByRole("row", { name: /Li Wei/ });
    await expect(pendingRow).toBeVisible({ timeout: 15_000 });

    await pendingRow.getByRole("button", { name: /^Approve$/ }).click();
    await expect(page.getByText("Approved Li Wei")).toBeVisible();
    await expect(page.getByRole("button", { name: /^Undo$/ })).toBeVisible();
    // The row flips optimistically while the undo window runs.
    await expect(pendingRow.getByText("Approved", { exact: true })).toBeVisible();

    // Let the 5s undo window elapse; the PUT lands and the row moves tabs.
    await page.getByRole("button", { name: "Approved", exact: true }).click();
    const approvedRow = page.getByRole("row", { name: /Li Wei/ });
    await expect(approvedRow).toBeVisible({ timeout: 15_000 });

    await approvedRow.click();
    await expect(page.getByRole("heading", { name: /^Decision$/ })).toBeVisible({
      timeout: 15_000,
    });
    await expect(page.getByText("Decided at")).toBeVisible();
    await expect(page.getByText("pearc26-tutorial")).toBeVisible();
    await expect(page.getByText(/5,000 credits/)).toBeVisible();
  });

  test("bulk select approves all selected pending rows", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/access-requests");
    await expect(page.getByRole("row", { name: /Amara Osei/ })).toBeVisible({ timeout: 15_000 });

    await page.getByRole("checkbox", { name: "Select Amara Osei" }).check();
    await page.getByRole("checkbox", { name: "Select Li Wei" }).check();
    await page.getByRole("button", { name: /^Approve 2 selected$/ }).click();

    // Bulk approve fires immediately; the pending view drains after refetch.
    await expect(page.getByText(/no access requests match/i)).toBeVisible({ timeout: 15_000 });
    await page.getByRole("button", { name: "Approved", exact: true }).click();
    await expect(page.getByRole("row", { name: /Amara Osei/ })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("row", { name: /Li Wei/ })).toBeVisible();
  });
});
