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

const SHOTS =
  process.env.ANALYTICS_SHOT_DIR ??
  "/private/tmp/claude-501/-Users-lahiruj-Projects-dev-apache-airavata-custos/dbebf1bd-468c-4188-ac8f-665e07c7e37f/scratchpad";
const SEVERITIES = ["serious", "critical"] as const;

test.describe("analytics page", () => {
  test("member view: shared base, no member breakdown, jobs toggle, or role chip", async ({
    page,
  }) => {
    await signInAs(page, "viewer");
    // Explicit allocation: the Climate context where the caller is a MEMBER.
    await page.goto("/analytics?allocation=alloc-climate");

    await expect(page.getByRole("heading", { name: "Analytics", exact: true })).toBeVisible({
      timeout: 20_000,
    });
    await expect(page.getByText("Credits left")).toBeVisible();
    await expect(page.getByRole("heading", { name: "Usage over time" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Usage by resource" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Recent jobs" })).toBeVisible();
    // The manager-only affordances and the role chip are all absent for a member.
    await expect(page.getByRole("heading", { name: "Usage by member" })).toHaveCount(0);
    await expect(page.getByRole("button", { name: "Everyone" })).toHaveCount(0);
    await expect(page.getByText("PI on this project")).toHaveCount(0);

    await page.screenshot({ path: `${SHOTS}/analytics-member-light.png`, fullPage: true });
  });

  test("PI view: member breakdown, jobs toggle, role chip; switcher flips role", async ({
    page,
  }) => {
    await signInAs(page, "manager");
    // Explicit allocation: the GenAI context where the caller is a PI.
    await page.goto("/analytics?allocation=alloc-genai");

    await expect(page.getByText("PI on this project")).toBeVisible({ timeout: 20_000 });
    await expect(page.getByRole("heading", { name: "Usage by resource" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Usage by member" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Everyone" })).toBeVisible();

    await page.screenshot({ path: `${SHOTS}/analytics-pi-light.png`, fullPage: true });
    await page.evaluate(() => document.documentElement.classList.add("dark"));
    await page.screenshot({ path: `${SHOTS}/analytics-pi-dark.png`, fullPage: true });

    // Switcher: pick the Climate allocation, where the caller is only a member,
    // and the member-only cards should disappear.
    await page.getByRole("button", { name: /GenAI GPU/ }).click();
    const climate = page.getByRole("menuitem", { name: /Climate GPU/ });
    await climate.waitFor({ state: "visible" });
    await climate.click();
    await expect(page.getByRole("heading", { name: "Usage by member" })).toHaveCount(0);
  });

  test("axe: no serious or critical violations", async ({ page }) => {
    await signInAs(page, "viewer");
    await page.goto("/analytics");
    await expect(page.getByRole("heading", { name: "Usage by resource" })).toBeVisible({
      timeout: 20_000,
    });
    const results = await new AxeBuilder({ page }).analyze();
    const serious = results.violations.filter((v) =>
      SEVERITIES.includes(v.impact as (typeof SEVERITIES)[number]),
    );
    expect(serious).toEqual([]);
  });
});
