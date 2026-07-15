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

import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

test.describe("ssh certificates", () => {
  test("lists certificates and navigates to a detail page", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/signer/certificates");
    await expect(page.getByRole("heading", { name: /^SSH Certificates$/ })).toBeVisible();

    const firstRow = page.locator('a[href^="/signer/certificates/"]').first();
    await expect(firstRow).toBeVisible({ timeout: 15_000 });
    await firstRow.click();
    await expect(page).toHaveURL(/\/signer\/certificates\/\d+/);
    await expect(page.getByRole("heading", { name: /^Certificate \d+$/ })).toBeVisible();
  });

  test("status filter updates URL state", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/signer/certificates");
    await page.getByLabel(/filter by status/i).selectOption("revoked");
    await expect(page).toHaveURL(/[?&]status=revoked/);
  });

  test("shows the revoked status on an already-revoked certificate", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/signer/certificates/44");
    await expect(page.getByRole("heading", { name: /^Certificate 44$/ })).toBeVisible();
    await expect(page.getByText(/Key compromised/).first()).toBeVisible({ timeout: 15_000 });
  });

  test("revokes an active certificate and reflects backend state", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/signer/certificates/42");
    await expect(page.getByRole("heading", { name: /^Certificate 42$/ })).toBeVisible();

    await page.getByRole("button", { name: /^Revoke$/ }).click();
    await page.getByLabel(/reason/i).fill("Compromised in e2e");
    await page.getByRole("button", { name: /^Confirm revoke$/ }).click();

    // The detail refetches after revoke and shows backend state (reason + no
    // remaining Revoke button), rather than a faked client-side update. The
    // button's absence also means a double-revoke can't be triggered from the UI.
    await expect(page.getByText(/Compromised in e2e/).first()).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("button", { name: /^Revoke$/ })).toHaveCount(0);
  });

  test("axe sweep on certificates list", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/signer/certificates");
    await expect(page.getByRole("heading", { name: /^SSH Certificates$/ })).toBeVisible();
    const results = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa"]).analyze();
    const blocking = results.violations.filter(
      (v) => v.impact === "serious" || v.impact === "critical",
    );
    expect(blocking, JSON.stringify(blocking, null, 2)).toEqual([]);
  });
});
