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
import { encode } from "next-auth/jwt";

test("an unlinked sign-in lands on a banner that asks an administrator to check the account", async ({
  page,
}) => {
  test.setTimeout(90_000);
  await page.goto("/sign-in?error=identity_not_linked", { waitUntil: "domcontentloaded" });
  await expect(page.locator("#auth-banner-text")).toHaveText(
    "Your sign-in is not linked to a portal account. Ask an administrator to check your account.",
    { timeout: 30_000 },
  );
});

test("a refused sign-in is signed out once, by the server", async ({ page }) => {
  test.setTimeout(90_000);
  const token = await encode({
    salt: "custos.session-token",
    secret: process.env.NEXTAUTH_SECRET ?? "",
    token: { sub: "test-disabled", email: "disabled@custos.local", authError: "user_disabled" },
  });
  await page.context().addCookies([
    {
      name: "custos.session-token",
      value: token,
      url: `http://localhost:${process.env.PORT ?? "3217"}`,
    },
  ]);
  const ends: string[] = [];
  const sessions: string[] = [];
  const errors: string[] = [];
  page.on("request", (r) => {
    if (r.url().includes("/api/auth/end-session")) ends.push(r.url());
    if (r.url().includes("/api/auth/session")) sessions.push(r.url());
  });
  page.on("console", (m) => {
    if (m.type() === "error") errors.push(m.text());
  });
  await page.goto("/projects");
  await expect(page.locator("#auth-banner-text")).toHaveText(
    "Your account is not active. Ask an administrator to reactivate it.",
    {
      timeout: 30_000,
    },
  );
  expect(ends).toHaveLength(1);
  expect(sessions).toEqual([]);
  expect(errors).toEqual([]);
  const cookies = await page.context().cookies();
  expect(cookies.find((c) => c.name === "custos.session-token")).toBeUndefined();
});
