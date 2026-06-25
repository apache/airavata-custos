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

test.describe("AMIE → Traces cross-link", () => {
  test("clicking View trace on an AMIE event navigates into the trace drawer", async ({
    page,
  }) => {
    await signInAs(page, "admin");
    // pkt-100 has an event with trace_id a3b1c92d3f4e5a6b7c8d9e0f12345678.
    await page.goto("/admin/amie/packets/pkt-100");
    await expect(page.getByRole("tab", { name: /Timeline/i })).toBeVisible({ timeout: 20_000 });
    await page.getByRole("tab", { name: /Timeline/i }).click();

    const link = page.getByRole("link", { name: /View trace/i }).first();
    await expect(link).toBeVisible({ timeout: 10_000 });
    await link.click();

    await expect(page).toHaveURL(/\/admin\/traces\/[0-9a-f]{32}/, { timeout: 10_000 });
    await expect(page.getByTestId("trace-detail-drawer")).toBeVisible({ timeout: 20_000 });
  });
});
