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

import { zUser } from "@/generated/core/zod.gen";
import { serverEnv } from "@/lib/env";
import NextAuth, { type NextAuthConfig } from "next-auth";
import type { JWT } from "next-auth/jwt";
import Keycloak from "next-auth/providers/keycloak";
import { z } from "zod";
import { discoverOidcEndpoints } from "./oidc";

// Only what the token keeps from GET /me.
const meTokenSchema = z.object({
  user: zUser.pick({ id: true }).nullable(),
  onboarding: z.object({ required: z.boolean() }),
  consent: z.object({ pending: z.array(z.unknown()) }),
});

function looksLikeJwt(token: string | null | undefined): token is string {
  return typeof token === "string" && token.split(".").length === 3;
}

const tokenSetSchema = z.object({
  access_token: z.string().optional(),
  id_token: z.string().optional(),
  refresh_token: z.string().optional(),
});
type TokenSet = z.infer<typeof tokenSetSchema>;

function jwtExpiry(jwt: string): number | undefined {
  try {
    const { exp } = JSON.parse(Buffer.from(jwt.split(".")[1] ?? "", "base64url").toString());
    return typeof exp === "number" ? exp : undefined;
  } catch {
    return undefined;
  }
}

// The cookie keeps only the backend bearer, the id_token (the logout hint) and the refresh token;
// an opaque access token the backend never sees would only push response headers past proxy limits.
function storeTokens(token: JWT, issued: TokenSet): void {
  delete token.accessToken;
  if (looksLikeJwt(issued.access_token)) token.accessToken = issued.access_token;
  if (issued.id_token) token.idToken = issued.id_token;
  if (issued.refresh_token) token.refreshToken = issued.refresh_token;
  const bearer = backendBearerOf(token);
  const expiresAt = bearer && jwtExpiry(bearer);
  if (expiresAt) token.expiresAt = expiresAt;
}

async function refreshTokens(token: JWT): Promise<JWT> {
  const failed: JWT = { ...token, error: "RefreshAccessTokenError" };
  if (!token.refreshToken) return failed;
  const endpoint = (await discoverOidcEndpoints())?.token_endpoint;
  if (!endpoint) return failed;
  try {
    const res = await fetch(endpoint, {
      method: "POST",
      headers: { "content-type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        grant_type: "refresh_token",
        refresh_token: token.refreshToken,
        client_id: serverEnv.OIDC_CLIENT_ID,
        client_secret: serverEnv.OIDC_CLIENT_SECRET,
      }),
      cache: "no-store",
    });
    if (!res.ok) return failed;
    const issued = tokenSetSchema.parse(await res.json());
    // Without a new JWT bearer the old id_token would be sent after it expired.
    if (!issued.id_token && !looksLikeJwt(issued.access_token)) return failed;
    storeTokens(token, issued);
    return token;
  } catch {
    return failed;
  }
}

async function loadProfile(token: JWT, bearer: string | undefined) {
  try {
    const res = await fetch(`${serverEnv.CUSTOS_CORE_API_BASE_URL}/me`, {
      headers: { authorization: `Bearer ${bearer}` },
      cache: "no-store",
    });
    if (!res.ok) return;
    const body = meTokenSchema.parse(await res.json());
    if (body.user) token.custosUserId = body.user.id;
    token.needsOnboarding = body.onboarding.required;
    token.consentPending = body.consent.pending.length > 0;
  } catch {
    // The token stays without a Custos user, so the portal shows the no-access notice.
  }
}

const isProduction = serverEnv.NODE_ENV === "production";
export const sessionCookieName = isProduction
  ? "__Secure-custos.session-token"
  : "custos.session-token";

// Refresh tokens keep the bearer alive; the IdP's refresh-token lifetime ends the session sooner.
const SESSION_MAX_AGE_SECONDS = 12 * 60 * 60;
// Twice the session poll interval in Providers.tsx, so a poll persists the refresh.
const REFRESH_BEFORE_EXPIRY_SECONDS = 120;

// The bearer the backend accepts: the JWT access token when the IdP issues one, else the id_token.
export function backendBearerOf(token: JWT): string | undefined {
  return token.accessToken ?? token.idToken;
}

// shortcut: a refresh outside the session poll cannot rewrite the cookie, so it repeats until a poll persists it.
export async function refreshIfDue(token: JWT): Promise<JWT> {
  const refreshDue =
    token.expiresAt !== undefined &&
    Date.now() / 1000 >= token.expiresAt - REFRESH_BEFORE_EXPIRY_SECONDS;
  return refreshDue && !token.error ? refreshTokens(token) : token;
}

export const authConfig: NextAuthConfig = {
  trustHost: true,
  secret: serverEnv.NEXTAUTH_SECRET,
  session: { strategy: "jwt", maxAge: SESSION_MAX_AGE_SECONDS },
  pages: { signIn: "/sign-in" },
  cookies: {
    sessionToken: {
      name: sessionCookieName,
      options: {
        httpOnly: true,
        sameSite: "lax",
        path: "/",
        secure: isProduction,
      },
    },
  },
  providers: [
    Keycloak({
      id: "oidc",
      issuer: serverEnv.OIDC_ISSUER_URL,
      clientId: serverEnv.OIDC_CLIENT_ID,
      clientSecret: serverEnv.OIDC_CLIENT_SECRET,
      authorization: { params: { scope: "openid email profile" } },
    }),
  ],
  callbacks: {
    async jwt({ token, account, trigger }) {
      if (account) {
        console.info("OIDC sign-in", { refreshTokenIssued: Boolean(account.refresh_token) });
        storeTokens(token, account);
        await loadProfile(token, backendBearerOf(token));
        return token;
      }
      const current = await refreshIfDue(token);
      if (trigger === "update") await loadProfile(current, backendBearerOf(current));
      return current;
    },
    // Tokens stay in the encrypted cookie; the session reaches browser JS through /api/auth/session.
    async session({ session, token }) {
      if (token.error) session.error = token.error;
      if (token.custosUserId) session.custosUserId = token.custosUserId;
      session.needsOnboarding = token.needsOnboarding === true;
      session.consentPending = token.consentPending === true;
      if (session.user) {
        // Prefer the backend-resolved Custos user_id over the OIDC sub so
        // downstream API callers get a value that matches users.id.
        session.user.id = token.custosUserId ?? token.sub ?? session.user.id;
      }
      return session;
    },
  },
};

export const { auth, handlers, signIn, signOut } = NextAuth(authConfig);
