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

import type { Account } from "next-auth";
import type { JWT } from "next-auth/jwt";
import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("next-auth", () => ({ default: () => ({}) }));
vi.mock("next-auth/providers/keycloak", () => ({ default: () => ({}) }));
vi.mock("@/lib/env", () => ({
  serverEnv: {
    NODE_ENV: "test",
    OIDC_ISSUER_URL: "https://idp.example.org",
    OIDC_CLIENT_ID: "client-123",
    OIDC_CLIENT_SECRET: "secret",
    CUSTOS_CORE_API_BASE_URL: "https://core.example.org",
  },
}));

import { authConfig } from "../auth";

const fetchMock = vi.fn<typeof fetch>();
vi.stubGlobal("fetch", fetchMock);
afterEach(() => fetchMock.mockReset());

const ID_TOKEN = "h.id.s";
const meResponse = () =>
  new Response(
    JSON.stringify({
      user: { id: "u-1" },
      onboarding: { required: false },
      consent: { pending: [] },
    }),
  );

async function signIn(account: Partial<Account>): Promise<JWT> {
  fetchMock.mockResolvedValueOnce(meResponse());
  const jwt = authConfig.callbacks?.jwt;
  if (!jwt) throw new Error("jwt callback missing");
  return (await jwt({
    token: {},
    account: { provider: "oidc", type: "oidc", providerAccountId: "sub", ...account } as Account,
    user: { id: "sub" },
  } as Parameters<typeof jwt>[0])) as JWT;
}

describe("jwt callback at sign-in", () => {
  it("drops an opaque access token and calls the backend with the id_token", async () => {
    const token = await signIn({ access_token: "opaque-access", id_token: ID_TOKEN });
    expect(token.accessToken).toBeUndefined();
    expect(token.idToken).toBe(ID_TOKEN);
    expect(token.custosUserId).toBe("u-1");
    expect(fetchMock.mock.calls[0]?.[1]?.headers).toEqual({ authorization: `Bearer ${ID_TOKEN}` });
  });

  it("keeps a JWT access token as the backend bearer", async () => {
    const token = await signIn({ access_token: "h.access.s", id_token: ID_TOKEN });
    expect(token.accessToken).toBe("h.access.s");
    expect(token.idToken).toBe(ID_TOKEN);
  });
});
