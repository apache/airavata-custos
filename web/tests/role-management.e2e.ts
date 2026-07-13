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

test.describe("role management", () => {
  test("lists roles, creates a role, and updates its privileges", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/users/roles");

    await expect(page.getByRole("heading", { name: "Users & Permissions" })).toBeVisible();
    await expect(page.getByText("Super Admin")).toBeVisible();
    await expect(page.getByText("Operator")).toBeVisible();
    await expect(page.getByText("Auditor")).toBeVisible();

    const roleName = `Billing Reviewer ${Date.now()}`;
    await page.getByRole("button", { name: "Create role" }).click();
    await page.getByLabel("Name").fill(roleName);
    await page.getByLabel("Description").fill("Reviews allocation billing reports.");
    await page.getByRole("button", { name: "read", exact: true }).first().click();
    await page.getByRole("button", { name: "Create role" }).click();

    await expect(page.getByText("Role created")).toBeVisible();
    await expect(page.getByText(roleName)).toBeVisible();

    await page.getByRole("button", { name: "Edit" }).last().click();
    await expect(page.getByRole("dialog", { name: "Edit role" })).toBeVisible();
    await page.getByRole("button", { name: "write", exact: true }).first().click();
    await page.getByRole("button", { name: "Save changes" }).click();

    await expect(page.getByText("Role updated")).toBeVisible();
  });
});
