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

test.describe("agreement gates", () => {
  // Checks that a first sign-in sees the setup page with the suggested username instead of the portal.
  test("shows the onboarding page", async ({ page }) => {
    await signInAs(page, "researcher", "onboarding");
    await page.goto("/allocations");

    await expect(page.getByRole("heading", { name: "Set up your account" })).toBeVisible({
      timeout: 20_000,
    });
    await expect(page.getByLabel("Username")).toHaveValue("custos-jdoe");
  });

  // Checks that a user with a newer version pending sees the policy update page.
  test("shows the policy update page", async ({ page }) => {
    await signInAs(page, "researcher", "consent");
    await page.goto("/allocations");

    await expect(page.getByRole("heading", { name: "Accept the updated agreement" })).toBeVisible({
      timeout: 20_000,
    });
  });
});
