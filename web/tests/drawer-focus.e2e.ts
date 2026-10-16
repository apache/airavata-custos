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

// Where focus lands when a list's drawer closes, in real Chromium: a pointer press moves focus
// before the drawer reacts, which synthetic events in unit tests do not.

import { type Locator, type Page, expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

type List = {
  url: string;
  ready: (page: Page) => Locator;
  // The first row's opener, pressed with Enter; the second row, its cell pressed by pointer.
  opener: (page: Page) => Locator;
  swapCell: (page: Page) => Locator;
  swapOpener: (page: Page) => Locator;
};

const lists: Record<string, List> = {
  traces: {
    url: "/admin/traces",
    ready: (page) => page.locator('[data-testid^="trace-row-"]').nth(1),
    opener: (page) =>
      page.locator('[data-testid^="trace-row-"]').nth(0).locator("[data-row-opener]"),
    swapCell: (page) => page.locator('[data-testid^="trace-row-"]').nth(1),
    swapOpener: (page) =>
      page.locator('[data-testid^="trace-row-"]').nth(1).locator("[data-row-opener]"),
  },
  events: {
    url: "/admin/events",
    ready: (page) => page.getByRole("region", { name: "Event deliveries" }).getByRole("row").nth(2),
    opener: (page) => rows(page, "Event deliveries").nth(0).locator("[data-row-opener]"),
    swapCell: (page) => rows(page, "Event deliveries").nth(1).getByRole("cell").first(),
    swapOpener: (page) => rows(page, "Event deliveries").nth(1).locator("[data-row-opener]"),
  },
  "cluster accounts": {
    url: "/admin/users/cluster-accounts",
    ready: (page) => page.getByRole("region", { name: "Cluster accounts" }).getByRole("row").nth(2),
    opener: (page) => rows(page, "Cluster accounts").nth(0).locator("[data-row-opener]"),
    swapCell: (page) => rows(page, "Cluster accounts").nth(1).getByRole("cell").first(),
    swapOpener: (page) => rows(page, "Cluster accounts").nth(1).locator("[data-row-opener]"),
  },
};

const rows = (page: Page, region: string) =>
  page.getByRole("region", { name: region }).locator("tbody tr");

const drawer = (page: Page) => page.getByRole("dialog");

test.describe("drawer focus origin", () => {
  test.describe.configure({ timeout: 90_000 });
  test.use({ viewport: { width: 1920, height: 1080 } });

  for (const [name, list] of Object.entries(lists)) {
    for (const close of ["Escape", "Close"] as const) {
      test(`${name}: keyboard open, pointer swap, ${close} returns to the swapped row`, async ({
        page,
      }) => {
        await signInAs(page, "admin");
        await page.goto(list.url);
        await expect(list.ready(page)).toBeVisible({ timeout: 60_000 });

        await list.opener(page).focus();
        await page.keyboard.press("Enter");
        await expect(drawer(page)).toBeVisible();
        await list.swapCell(page).click({ position: { x: 4, y: 4 } });
        await expect(drawer(page).locator(":focus")).toHaveCount(1);

        if (close === "Escape") await page.keyboard.press("Escape");
        else await drawer(page).getByRole("button", { name: "Close" }).first().click();
        await expect(drawer(page)).toHaveCount(0);
        await expect(list.swapOpener(page)).toBeFocused();
      });
    }
  }

  test("rates: pointer swap, the row re-renders, Escape returns to that row's Rates", async ({
    page,
  }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/resources?tab=resources");
    const resources = rows(page, "Resources");
    await expect(resources.nth(1)).toBeVisible({ timeout: 60_000 });

    await resources
      .nth(0)
      .getByRole("button", { name: /^Rates/ })
      .focus();
    await page.keyboard.press("Enter");
    await expect(drawer(page)).toBeVisible();
    const second = resources.nth(1);
    const key = await second.getAttribute("data-row-key");
    await second.getByRole("button", { name: /^Rates/ }).click();
    await expect(drawer(page).getByRole("button", { name: "Close" })).toBeFocused();

    const before = await second.getByRole("button", { name: /^Rates/ }).textContent();
    await drawer(page).getByRole("button", { name: "+ Add rate" }).click();
    await drawer(page).getByRole("spinbutton", { name: "Rate" }).fill("0.42");
    await drawer(page).getByLabel("Ends").fill("2099-12-31");
    await drawer(page).getByRole("button", { name: "Add rate", exact: true }).click();
    await expect(second.getByRole("button", { name: /^Rates/ })).not.toHaveText(before ?? "");

    await page.keyboard.press("Escape");
    await expect(drawer(page)).toHaveCount(0);
    await expect(
      page.locator(`tr[data-row-key="${key}"]`).getByRole("button", { name: /^Rates/ }),
    ).toBeFocused();
  });

  test("a drawer opened from the address bar closes to the list region", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/users/cluster-accounts?account=03000012-0000-4000-8000-000003000012");
    await expect(drawer(page)).toBeVisible({ timeout: 60_000 });
    await expect(rows(page, "Cluster accounts").nth(1)).toBeVisible();
    await page.keyboard.press("Escape");

    await expect(drawer(page)).toHaveCount(0);
    await expect(page.getByRole("region", { name: "Cluster accounts" })).toBeFocused();
  });

  test("a drawer whose row left the list closes to the list region", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/events?view=failed");
    const failed = rows(page, "Event deliveries");
    await expect(failed.getByRole("button", { name: /^Retry/ })).toBeVisible({ timeout: 60_000 });

    await failed.getByRole("button", { name: /^Retry/ }).focus();
    await page.keyboard.press("Enter");
    await drawer(page).getByRole("button", { name: "Retry now" }).click();
    await expect(page.getByText("Retry started")).toBeVisible();
    await expect(failed.getByRole("button", { name: /^Retry/ })).toHaveCount(0);

    await page.keyboard.press("Escape");
    await expect(drawer(page)).toHaveCount(0);
    await expect(page.getByRole("region", { name: "Event deliveries" })).toBeFocused();
  });
});
