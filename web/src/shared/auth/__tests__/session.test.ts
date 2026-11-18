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

import { type JWT, encode } from "next-auth/jwt";
import { describe, expect, it, vi } from "vitest";

const SECRET = "s".repeat(32);

vi.mock("server-only", () => ({}));
vi.mock("next-auth", () => ({ default: () => ({}) }));
vi.mock("next-auth/providers/keycloak", () => ({ default: () => ({}) }));
vi.mock("@/lib/env", () => ({
  serverEnv: {
    NODE_ENV: "production",
    NEXTAUTH_SECRET: "s".repeat(32),
    OIDC_ISSUER_URL: "https://idp.example.org",
    OIDC_CLIENT_ID: "client-123",
    OIDC_CLIENT_SECRET: "secret",
  },
}));

import { sessionCookieName } from "../auth";
import { freshBackendBearer } from "../session";

const fetchMock = vi.fn<typeof fetch>();
vi.stubGlobal("fetch", fetchMock);

// The proxy sees plain http behind the reverse proxy, with the __Secure- cookie.
async function requestWith(token: JWT): Promise<Request> {
  const value = await encode({ token, secret: SECRET, salt: sessionCookieName });
  return new Request("http://localhost:3000/api/v1/me", {
    headers: { cookie: `${sessionCookieName}=${value}` },
  });
}

describe("freshBackendBearer", () => {
  it("reads the bearer from the session cookie", async () => {
    const req = await requestWith({ idToken: "h.id.s", expiresAt: Date.now() / 1000 + 3600 });
    expect(await freshBackendBearer(req)).toBe("h.id.s");
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("returns nothing without a session cookie", async () => {
    expect(
      await freshBackendBearer(new Request("http://localhost:3000/api/v1/me")),
    ).toBeUndefined();
  });
});
