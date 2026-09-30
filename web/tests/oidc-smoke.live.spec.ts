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

// Live smoke against Keycloak and the backend configured for the portal.
// Picked up only by playwright.live.config.ts (`.spec.ts` pattern).
test.describe("OIDC smoke", () => {
  test("admin signs in and sees captured activity from the backend", async ({ page, request }) => {
    await page.goto("/sign-in");
    await page.waitForURL(/\/realms\/custos\/protocol\/openid-connect\/auth/);
    await page.getByLabel(/username or email/i).fill("admin");
    await page.getByLabel("Password", { exact: true }).fill("admin");
    await page.getByRole("button", { name: /^sign in$/i }).click();
    const portalOrigin = new URL(process.env.LIVE_PORTAL_URL ?? "http://localhost:3001").origin;
    await page.waitForURL(
      (url) => url.origin === portalOrigin && !url.pathname.startsWith("/sign-in"),
    );
    await expect(page.getByText("admin@custos.local")).toBeVisible();

    const cookies = await page.context().cookies();
    const cookieHeader = cookies.map((c) => `${c.name}=${c.value}`).join("; ");
    const meRes = await request.get("/api/v1/users/dev-admin", {
      headers: { cookie: cookieHeader },
    });
    expect(meRes.status()).toBe(200);
    const body = await meRes.json();
    expect(body.email).toBe("admin@custos.local");

    await page.goto("/admin/users/activity");
    await expect(page.getByRole("heading", { name: "User activity overview" })).toBeVisible();
    await expect(
      page
        .getByRole("table", { name: "User login activity" })
        .getByRole("row")
        .filter({ hasText: "admin@custos.local" }),
    ).toBeVisible();
    const activityRes = await request.get(
      "/api/v1/users/activity?window=30&status=active&query=admin%40custos.local&limit=10&offset=0&sort=last_login&direction=desc",
      { headers: { cookie: cookieHeader } },
    );
    expect(activityRes.status()).toBe(200);
    const activity = await activityRes.json();
    expect(activity.total).toBe(1);
    expect(activity.items[0].login_count).toBe(1);
  });
});
