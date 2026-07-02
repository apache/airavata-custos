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

// Live smoke against the dev-ops/compose Keycloak + the backend on :8080.
// Picked up only by playwright.live.config.ts (`.spec.ts` pattern).
test.describe("OIDC smoke", () => {
  test("admin signs in, lands on portal, /users/dev-admin returns 200", async ({ page, request }) => {
    await page.goto("/sign-in");
    await page.getByRole("button", { name: /sign in with custos/i }).click();
    await page.waitForURL(/\/realms\/custos\/protocol\/openid-connect\/auth/);
    await page.getByLabel(/username or email/i).fill("admin");
    await page.getByLabel(/password/i).fill("admin");
    await page.getByRole("button", { name: /^sign in$/i }).click();
    await page.waitForURL((url) => !url.pathname.startsWith("/sign-in") && !url.host.includes("8081"));
    await expect(page.getByText("admin@custos.local")).toBeVisible();

    const cookies = await page.context().cookies();
    const cookieHeader = cookies.map((c) => `${c.name}=${c.value}`).join("; ");
    const meRes = await request.get("/api/v1/users/dev-admin", {
      headers: { cookie: cookieHeader },
    });
    expect(meRes.status()).toBe(200);
    const body = await meRes.json();
    expect(body.email).toBe("admin@custos.local");
  });
});
