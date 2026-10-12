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

test.describe("project detail", () => {
  test("renders header, overview, and members tab", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects/19000001-0000-4000-8000-000019000001");
    await expect(page.getByRole("heading", { name: /Genomic Sequencing Pipeline/i })).toBeVisible();

    await page.getByRole("tab", { name: /members/i }).click();
    await expect(page.getByText(/Ada Lovelace/i).first()).toBeVisible({ timeout: 15_000 });
  });

  test("replaces a padded route id with the normalised one, keeping the query", async ({
    page,
  }) => {
    await signInAs(page, "admin");
    await page.goto("/projects/%2019000001-0000-4000-8000-000019000001%20?tab=members");
    await expect(page).toHaveURL(/\/projects\/19000001-0000-4000-8000-000019000001\?tab=members$/);
    await expect(page.getByRole("heading", { name: /Genomic Sequencing Pipeline/i })).toBeVisible();
  });

  test("an edit sends the full project record", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects/19000002-0000-4000-8000-000019000002");
    await page.getByRole("button", { name: "Edit project" }).click();
    const pi = await page.getByLabel("PI user ID").inputValue();
    await page.getByLabel("Title").fill("Renamed project");

    const put = page.waitForRequest(
      (r) =>
        r.method() === "PUT" &&
        r.url().endsWith("/api/v1/projects/19000002-0000-4000-8000-000019000002"),
    );
    await page.getByRole("dialog").getByRole("button", { name: "Edit project" }).click();
    expect((await put).postDataJSON()).toEqual({ title: "Renamed project", project_pi_id: pi });
    await expect(page.getByText("Project updated")).toBeVisible();
  });

  test("marks a project deleted through the confirmation", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects/19000002-0000-4000-8000-000019000002");
    await page.getByRole("combobox", { name: "Project status" }).click();
    await page.getByRole("option", { name: "Deleted" }).click();

    const put = page.waitForRequest(
      (r) =>
        r.method() === "PUT" &&
        r.url().endsWith("/api/v1/projects/19000002-0000-4000-8000-000019000002/status"),
    );
    await page.getByRole("button", { name: "Mark deleted", exact: true }).click();
    expect((await put).postDataJSON()).toEqual({ status: "DELETED" });
    await expect(page.getByText("Project marked deleted")).toBeVisible();
  });

  test("axe sweep on projects detail", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects/19000001-0000-4000-8000-000019000001");
    await expect(page.getByRole("heading", { name: /Genomic Sequencing Pipeline/i })).toBeVisible();
    const results = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa"]).analyze();
    const blocking = results.violations.filter(
      (v) => v.impact === "serious" || v.impact === "critical",
    );
    expect(blocking, JSON.stringify(blocking, null, 2)).toEqual([]);
  });
});
