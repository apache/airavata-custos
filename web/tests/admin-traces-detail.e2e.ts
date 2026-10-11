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

const SEVERITIES = ["serious", "critical"] as const;

async function openFailedTrace(page: import("@playwright/test").Page) {
  await signInAs(page, "admin");
  await page.goto("/admin/traces");
  const row = page.locator('[data-testid^="trace-row-"][data-status="failed"]').first();
  await expect(row).toBeVisible({ timeout: 20_000 });
  await row.click();
  await expect(page.getByTestId("trace-detail-drawer")).toBeVisible({ timeout: 10_000 });
}

test.describe("admin tracing drawer", () => {
  test("opens on the failing step with its error, survives reload, Esc closes it", async ({
    page,
  }) => {
    await openFailedTrace(page);
    await expect(page).toHaveURL(/[?&]trace=[0-9a-f]{32}\b/);

    const failing = page.getByTestId("failing-step");
    await expect(failing).toHaveAttribute("aria-selected", "true");
    await expect(page.getByTestId("step-error")).toContainText(/401 Unauthorized/);

    await page.reload();
    await expect(page.getByTestId("trace-detail-drawer")).toBeVisible({ timeout: 20_000 });
    await expect(page.getByTestId("failing-step")).toBeVisible();

    await page.keyboard.press("Escape");
    await expect(page.getByTestId("trace-detail-drawer")).toHaveCount(0, { timeout: 10_000 });
    await expect(page).not.toHaveURL(/[?&]trace=/);
  });

  // The delivery header is where an admin goes to retry; it must say how the
  // delivery ended and hand over to the Events page with the delivery open.
  test("the failed delivery links to its delivery in Events", async ({ page }) => {
    await openFailedTrace(page);
    const delivery = page
      .getByRole("treeitem")
      .filter({ hasText: "comanage-identity-provisioner" })
      .first();
    await expect(delivery).toContainText("Failed after 10 tries");
    await delivery.click();
    await expect(page).toHaveURL(/[?&]step=d%3A|[?&]step=d:/);
    const link = page
      .getByTestId("step-panel")
      .getByRole("link", { name: /Open delivery in Events/ });
    await expect(link).toHaveAttribute("href", /\/admin\/events\?delivery=/);
  });

  test("selecting a step and switching to Raw keeps the selection in the URL", async ({ page }) => {
    await openFailedTrace(page);
    const published = page
      .getByRole("treeitem")
      .filter({ hasText: /^Published/ })
      .first();
    await published.click();
    await expect(page).toHaveURL(/[?&]step=h/);
    await expect(page.getByTestId("step-panel")).toContainText(/Sent to 2 connectors/);

    await page.getByRole("tab", { name: /^Raw$/ }).click();
    await expect(page.getByRole("button", { name: /copy trace JSON/i })).toBeVisible();
    await expect(page).toHaveURL(/[?&]step=h/);
  });

  test("axe: no serious or critical violations with the drawer open", async ({ page }) => {
    await openFailedTrace(page);
    await page.waitForLoadState("networkidle");

    const results = await new AxeBuilder({ page })
      .options({ resultTypes: ["violations"] })
      .analyze();
    const blocking = results.violations.filter((v) =>
      SEVERITIES.includes((v.impact ?? "minor") as (typeof SEVERITIES)[number]),
    );
    expect(blocking, JSON.stringify(blocking, null, 2)).toEqual([]);
  });
});
