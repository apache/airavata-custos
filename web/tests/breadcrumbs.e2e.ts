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

const trail = (page: Page) => page.getByRole("navigation", { name: "Breadcrumb" }).locator("ol");

// The current crumb stays inside the trail and ends in its label or an ellipsis; it and every visible
// ancestor show their whole label or at least four characters before the ellipsis.
async function expectFits(page: Page, label: string) {
  const current = trail(page).locator('[aria-current="page"]');
  await expect(current).toHaveAttribute("title", label);
  const fit = await current.evaluate((el) => {
    const ol = el.closest("ol") as HTMLElement;
    const shown = (a: Element) => {
      const text = a.textContent ?? "";
      const { scrollWidth, clientWidth } = a as HTMLElement;
      return scrollWidth <= clientWidth
        ? text.length
        : Math.floor(clientWidth / (scrollWidth / text.length)) - 1;
    };
    const ancestors = [
      ...ol.querySelectorAll(":scope > li:not([aria-hidden]) > :not([aria-current])"),
    ]
      .filter((a) => (a as HTMLElement).offsetParent !== null)
      .map(shown);
    return {
      inside: el.getBoundingClientRect().right <= ol.getBoundingClientRect().right + 0.5,
      clipped: el.scrollWidth > el.clientWidth,
      ellipsis: getComputedStyle(el).textOverflow === "ellipsis",
      current: shown(el),
      ancestors,
    };
  });
  expect(fit.inside).toBe(true);
  expect(fit.current).toBeGreaterThanOrEqual(Math.min(4, label.length));
  if (fit.clipped) expect(fit.ellipsis).toBe(true);
  else await expect(current).toHaveText(label);
  for (const shown of fit.ancestors) expect(shown).toBeGreaterThanOrEqual(4);
}

const pages = [
  {
    path: "/change-requests/05000001-0000-4000-8000-000005000001",
    label: "Genomic Sequencing - GPU Pool",
  },
  { path: "/admin/users/management", label: "User management" },
  { path: "/admin/users/cluster-accounts", label: "Cluster accounts" },
  { path: "/admin/users/roles", label: "Role management" },
  { path: "/admin/resources", label: "Resources & clusters" },
];

for (const width of [390, 768, 820, 1024]) {
  test(`the current crumb stays inside the trail at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 });
    await signInAs(page, "admin");
    for (const { path, label } of pages) {
      await page.goto(path);
      await expectFits(page, label);
    }
  });
}

const LONG = "N".repeat(255);

test("a 255-character current label leaves its parent crumb readable", async ({ page }) => {
  // Lengthen the names the detail pages register before the app's first fetch.
  await page.addInitScript((long) => {
    const original = window.fetch;
    window.fetch = async (input, init) => {
      const response = await original(input, init);
      const url = typeof input === "string" ? input : input instanceof URL ? input.href : input.url;
      if (
        !/\/api\/v1\/(projects|compute-allocations|compute-allocation-change-requests)\/[^/?]+$/.test(
          new URL(url, location.href).pathname,
        ) ||
        !response.ok
      )
        return response;
      const body = await response.clone().json();
      for (const key of ["title", "name", "compute_allocation_name"])
        if (key in body) body[key] = long;
      return Response.json(body, { status: response.status, headers: response.headers });
    };
  }, LONG);
  await signInAs(page, "admin");
  for (const width of [390, 768, 1440]) {
    await page.setViewportSize({ width, height: 900 });
    for (const path of [
      "/allocations/01000001-0000-4000-8000-000001000001",
      "/projects/19000001-0000-4000-8000-000019000001",
      "/change-requests/05000001-0000-4000-8000-000005000001",
    ]) {
      await page.goto(path);
      await expectFits(page, LONG);
    }
  }
});
