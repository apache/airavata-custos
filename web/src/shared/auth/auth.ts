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

import NextAuth, { type NextAuthConfig } from "next-auth";
import Keycloak from "next-auth/providers/keycloak";
import { serverEnv } from "@/lib/env";
import { getMe } from "@/generated/core/sdk.gen";

function looksLikeJwt(token: string | null | undefined): boolean {
  return typeof token === "string" && token.split(".").length === 3;
}

const isProduction = serverEnv.NODE_ENV === "production";
const sessionCookieName = isProduction
  ? "__Secure-custos.session-token"
  : "custos.session-token";

// Match the access_token's natural lifetime; without refresh-token handling a
// longer session is misleading — the inner bearer dies first.
const SESSION_MAX_AGE_SECONDS = 60 * 60;

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
    async jwt({ token, account }) {
      if (account?.access_token) {
        token.accessToken = account.access_token;
        // Some IdPs hand out opaque access tokens; use the JWT-shaped
        // bearer for the backend call so the verifier accepts it.
        const bearer = looksLikeJwt(account.access_token)
          ? account.access_token
          : (account.id_token ?? account.access_token);
        try {
          const body = await getMe({
            baseUrl: serverEnv.CUSTOS_CORE_API_BASE_URL,
            headers: { authorization: `Bearer ${bearer}` },
            cache: "no-store",
          });
          token.custosUserId = body.user?.id;
          token.privileges = body.privileges ?? [];
        } catch {
          // The token stays without a Custos user, so the portal shows the no-access notice.
        }
      }
      if (account?.id_token) {
        // Needed for the Keycloak end-session endpoint's id_token_hint.
        token.idToken = account.id_token;
      }
      return token;
    },
    async session({ session, token }) {
      if (token.accessToken) session.accessToken = token.accessToken;
      if (token.idToken) session.idToken = token.idToken;
      session.privileges = token.privileges ?? [];
      session.custosUserId = token.custosUserId;
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
