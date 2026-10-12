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

const TEMP = ["temp-account:accounts:read", "temp-account:accounts:write"];

test.describe("temporary account memberships", () => {
  test("Edit loads the membership and Update saves it", async ({ page }) => {
    await signInAs(page, "admin", TEMP);
    await page.goto("/admin/temp-accounts?user=5f0c2a7e-3b9d-4e61-8a24-c1d7e9b0f342");
    const row = page.getByRole("row", { name: /Genomic Sequencing - GPU Pool/ });
    await row.getByRole("button", { name: "Edit" }).click();

    const dialog = page.getByRole("dialog", { name: "Update allocation" });
    await dialog.getByLabel("Status").selectOption("INACTIVE");
    await dialog.getByRole("button", { name: "Update" }).click();
    await expect(dialog).toBeHidden();
    await expect(row).toContainText(/inactive/i);
  });

  test("Edit on a membership deleted elsewhere says it no longer exists and drops the row", async ({
    page,
  }) => {
    await signInAs(page, "admin", TEMP);
    await page.goto("/admin/temp-accounts?user=5f0c2a7e-3b9d-4e61-8a24-c1d7e9b0f342");
    const row = page.getByRole("row", { name: /Genomic Sequencing - CPU Pool/ });
    await expect(row).toBeVisible();
    await page.evaluate(() =>
      fetch("/api/v1/compute-allocation-memberships/00000120-0000-4000-8000-099000000120", {
        method: "DELETE",
      }),
    );
    await row.getByRole("button", { name: "Edit" }).click();

    const dialog = page.getByRole("dialog", { name: "Update allocation" });
    await expect(dialog.getByText("This membership no longer exists")).toBeVisible();
    await expect(dialog.getByRole("button", { name: "Update" })).toBeDisabled();
    await expect(row).toBeHidden();
  });
});
