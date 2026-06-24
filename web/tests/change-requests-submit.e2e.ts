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

test.describe("change request submit", () => {
  test("PI persona opens drawer, fills, submits", async ({ page }) => {
    await signInAs(page, "manager");
    await page.goto("/allocations/alloc-001");
    await expect(
      page.getByRole("heading", { name: /Genomic Sequencing - GPU Pool/i }),
    ).toBeVisible();

    await page.getByRole("tab", { name: /^change requests$/i }).click();
    await page.getByRole("button", { name: /submit change request/i }).click();

    await expect(
      page.getByRole("heading", { name: /submit change request/i }),
    ).toBeVisible({ timeout: 15_000 });

    await page.getByLabel(/additional sus requested/i).fill("10000");
    await page
      .getByLabel(/reason/i)
      .fill("Need additional SUs for the final modeling run before paper submission deadline.");

    await page.getByRole("button", { name: /submit request/i }).click();

    await expect(page.getByText(/change request submitted/i)).toBeVisible({ timeout: 15_000 });
  });
});
