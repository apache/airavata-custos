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

// Test-only: mints a NextAuth session cookie so e2e tests skip the OIDC
// handshake. The app has no auth bypass; real OIDC runs in playwright.live.

import type { Page } from "@playwright/test";
import { encode } from "next-auth/jwt";

export type Persona = "viewer" | "manager" | "admin";

const PRIVILEGES: Record<Persona, string[]> = {
  viewer: ["hpc:read"],
  manager: ["hpc:read", "hpc:write", "amie:read"],
  admin: [
    "amie:read",
    "amie:write",
    "hpc:read",
    "hpc:write",
    "signer:read",
    "signer:write",
    "privileges:grant",
    "roles:manage",
  ],
};

const NAMES: Record<Persona, string> = {
  viewer: "Test Viewer",
  manager: "Test Manager",
  admin: "Test Admin",
};

export async function signInAs(page: Page, persona: Persona = "admin") {
  const secret = process.env.NEXTAUTH_SECRET;
  if (!secret) {
    throw new Error("NEXTAUTH_SECRET must be set for the cookie-injection fixture");
  }

  const token = await encode({
    salt: "custos.session-token",
    secret,
    token: {
      sub: `test-${persona}`,
      name: NAMES[persona],
      email: `${persona}@custos.local`,
      privileges: PRIVILEGES[persona],
      accessToken: "test-mock-bearer",
    },
  });

  const port = process.env.PORT ?? "3217";
  await page.context().addCookies([
    {
      name: "custos.session-token",
      value: token,
      url: `http://localhost:${port}`,
      httpOnly: true,
      secure: false,
      sameSite: "Lax",
    },
  ]);
}
