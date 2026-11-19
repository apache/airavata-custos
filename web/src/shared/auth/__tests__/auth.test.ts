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

const jwtExpiringAt = (exp: number) =>
  `h.${Buffer.from(JSON.stringify({ exp })).toString("base64url")}.s`;
const nowSeconds = () => Math.floor(Date.now() / 1000);

async function laterRequest(token: JWT): Promise<JWT> {
  const jwt = authConfig.callbacks?.jwt;
  if (!jwt) throw new Error("jwt callback missing");
  return (await jwt({ token } as Parameters<typeof jwt>[0])) as JWT;
}

describe("jwt callback on later requests", () => {
  it("leaves a live bearer alone", async () => {
    const token = await laterRequest({ idToken: "x", expiresAt: nowSeconds() + 3600 });
    expect(fetchMock).not.toHaveBeenCalled();
    expect(token.error).toBeUndefined();
  });

  it("refreshes an expiring bearer at the discovered token endpoint", async () => {
    const exp = nowSeconds() + 3600;
    const fresh = jwtExpiringAt(exp);
    fetchMock
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ token_endpoint: "https://idp.example.org/token" })),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ id_token: fresh })));
    const token = await laterRequest({
      idToken: "old",
      refreshToken: "refresh-1",
      expiresAt: nowSeconds() + 10,
    });
    const [url, init] = fetchMock.mock.calls[1] ?? [];
    expect(url).toBe("https://idp.example.org/token");
    expect(String(init?.body)).toContain("grant_type=refresh_token&refresh_token=refresh-1");
    expect(token.idToken).toBe(fresh);
    expect(token.refreshToken).toBe("refresh-1");
    expect(token.expiresAt).toBe(exp);
    expect(token.error).toBeUndefined();
  });

  it("flags the session when the bearer expires with no refresh token", async () => {
    const token = await laterRequest({ idToken: "old", expiresAt: nowSeconds() - 1 });
    expect(token.error).toBe("RefreshAccessTokenError");
  });

  it("flags the session when the issuer refuses the refresh", async () => {
    fetchMock
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ token_endpoint: "https://idp.example.org/token" })),
      )
      .mockResolvedValueOnce(new Response("{}", { status: 400 }));
    const token = await laterRequest({ refreshToken: "r", expiresAt: nowSeconds() - 1 });
    expect(token.error).toBe("RefreshAccessTokenError");
  });
});

describe("session callback", () => {
  it("keeps tokens out of the browser-visible session", async () => {
    const session = authConfig.callbacks?.session;
    if (!session) throw new Error("session callback missing");
    const result = await session({
      session: { user: {}, expires: "" },
      token: {
        accessToken: "h.access.s",
        idToken: ID_TOKEN,
        refreshToken: "r",
        custosUserId: "u-1",
      },
    } as Parameters<typeof session>[0]);
    expect(JSON.stringify(result)).not.toMatch(/h\.access\.s|h\.id\.s|"r"/);
    expect(result).toMatchObject({ custosUserId: "u-1" });
  });
});
