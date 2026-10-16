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
import { type Locator, type Page, expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

const VIEWPORTS = [
  { width: 390, height: 667 },
  { width: 1366, height: 768 },
];

const overlaps = (a: DOMRect, b: DOMRect) =>
  a.left < b.right && b.left < a.right && a.top < b.bottom && b.top < a.bottom;

// The confirmation stands inside the container that raised it, fully on screen and clear of its
// Close, which stays clickable; its action then sends exactly one request.
async function expectHostedConfirm(
  page: Page,
  container: Locator,
  message: RegExp,
  action: string,
  request: (method: string, url: string) => boolean,
) {
  const bar = container.getByRole("group", { name: message });
  await expect(bar).toBeInViewport({ ratio: 1 });
  await expect(bar.getByRole("button", { name: "Cancel" })).toBeFocused();
  const close = container.getByRole("button", { name: "Close" });
  await close.click({ trial: true });
  const [barBox, closeBox] = await Promise.all([
    bar.evaluate((el) => el.getBoundingClientRect().toJSON()),
    close.evaluate((el) => el.getBoundingClientRect().toJSON()),
  ]);
  expect(overlaps(barBox, closeBox)).toBe(false);

  const sent: string[] = [];
  page.on("request", (r) => {
    if (request(r.method(), r.url())) sent.push(r.url());
  });
  await bar.getByRole("button", { name: action, exact: true }).dblclick();
  await expect(bar).toBeHidden();
  await expect.poll(() => sent.length).toBe(1);
  await page.waitForTimeout(300);
  expect(sent).toHaveLength(1);
}

for (const viewport of VIEWPORTS) {
  test.describe(`confirmations at ${viewport.width}x${viewport.height}`, () => {
    test.use({ viewport });

    test("deleting a role confirms inside the role editor", async ({ page }) => {
      await signInAs(page, "admin");
      await page.goto("/admin/users/roles");
      await page.getByRole("button", { name: "Edit" }).last().click();
      const editor = page.getByRole("dialog", { name: "Edit role" });
      await editor.getByRole("button", { name: "Delete role" }).click();
      await expectHostedConfirm(
        page,
        editor,
        /^Delete role/,
        "Delete",
        (method, url) => method === "DELETE" && /\/roles\//.test(url),
      );
    });

    test("removing an identity confirms inside the user drawer", async ({ page }) => {
      await signInAs(page, "admin");
      await page.goto("/admin/users/management");
      await page.getByRole("button", { name: "Elena Vasquez" }).first().click();
      const drawer = page.getByRole("dialog", { name: "Elena Vasquez" });
      await drawer.getByRole("button", { name: "Remove CILogon identity" }).click();
      await expectHostedConfirm(
        page,
        drawer,
        /^Remove the CILogon identity/,
        "Remove",
        (method, url) => method === "DELETE" && /identities/.test(url),
      );
    });
  });
}

test("a reopened filter select leaves no focusable node under aria-hidden", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 900 });
  await signInAs(page, "admin");
  await page.goto("/admin/users/management");
  const filter = page.getByRole("combobox", { name: "Filter this page by external identity" });
  // The guards mount again on every open, so the check runs on a second one.
  await filter.click();
  await expect(page.getByRole("listbox")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("listbox")).toBeHidden();
  await filter.click();
  await expect(page.getByRole("listbox")).toBeVisible();
  const results = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa"]).analyze();
  expect(results.violations, JSON.stringify(results.violations, null, 2)).toEqual([]);
  // Tab leaves the list for the control after the trigger, as the guards it no longer uses did.
  const trigger = page.getByRole("combobox", { name: "Filter this page by external identity" });
  const next = await trigger.evaluate((el) => {
    const all = [...document.querySelectorAll<HTMLElement>("button, input, a[href], [tabindex]")];
    const tabbable = all.filter((e) => e.tabIndex >= 0 && e.offsetParent !== null);
    return tabbable[tabbable.indexOf(el as HTMLElement) + 1]?.outerHTML;
  });
  await page.keyboard.press("Tab");
  await expect(page.getByRole("listbox")).toBeHidden();
  expect(await page.evaluate(() => document.activeElement?.outerHTML)).toBe(next);
});

test("a page confirmation returns focus to its raiser, else to the raiser's list", async ({
  page,
}) => {
  await signInAs(page, "admin");
  await page.goto("/allocations/01000001-0000-4000-8000-000001000001?tab=members");
  const remove = page.getByRole("button", { name: /^remove /i }).first();
  await remove.click();
  await page.getByRole("button", { name: "Cancel" }).click();
  await expect(remove).toBeFocused();
  // A dismissed toast keeps its buttons until its exit animation ends.
  await expect(page.getByRole("button", { name: "Remove", exact: true })).toHaveCount(0);

  const label = await remove.getAttribute("aria-label");
  await remove.click();
  await page.getByRole("button", { name: "Remove", exact: true }).click();
  await expect(page.getByRole("button", { name: label ?? "" })).toHaveCount(0);
  const focused = () =>
    page.evaluate(() => {
      const active = document.activeElement;
      return active === document.body ? "body" : active?.closest('[role="region"]') ? "list" : "";
    });
  await expect.poll(focused).toBe("list");
});

test("a double press on a submit sends one write", async ({ page }) => {
  await signInAs(page, "admin");
  await page.goto("/allocations/01000004-0000-4000-8000-000001000004?tab=members");
  const sent: string[] = [];
  page.on("request", (r) => {
    if (r.method() === "POST" && /memberships/.test(r.url())) sent.push(r.url());
  });
  await page.getByRole("button", { name: /\+ add member/i }).click();
  await page.getByRole("radio", { name: /Elena Vasquez/ }).check();
  await page.getByRole("button", { name: /^add member$/i }).dblclick();
  await expect.poll(() => sent.length).toBe(1);
  await page.waitForTimeout(300);
  expect(sent).toHaveLength(1);
});
