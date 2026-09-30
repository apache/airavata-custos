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

import type { Privilege } from "@/features/core/identity/types";
import { serverEnv } from "@/lib/env";
import { captureInitialOIDCLogin, loginCaptureBearer } from "@/shared/auth/login-capture";
import NextAuth, { type NextAuthConfig } from "next-auth";
import Keycloak from "next-auth/providers/keycloak";

const isProduction = serverEnv.NODE_ENV === "production";
const sessionCookieName = isProduction ? "__Secure-custos.session-token" : "custos.session-token";

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
        (token as { accessToken?: string }).accessToken = account.access_token;
        // Some IdPs hand out opaque access tokens; use the JWT-shaped
        // bearer for the backend call so the verifier accepts it.
        const bearer = loginCaptureBearer(account) ?? account.access_token;
        try {
          const res = await fetch(`${serverEnv.CUSTOS_CORE_API_BASE_URL}/me`, {
            headers: { authorization: `Bearer ${bearer}` },
            cache: "no-store",
          });
          if (res.ok) {
            const body = (await res.json()) as {
              user?: { id?: string };
              privileges?: Privilege[];
            };
            const t = token as { custosUserId?: string; privileges?: Privilege[] };
            t.custosUserId = body.user?.id;
            t.privileges = body.privileges ?? [];
          }
        } catch {
          // Leave token as-is; the portal shows the no-access notice for the empty case.
        }
        const recorded = await captureInitialOIDCLogin({
          account,
          coreApiBaseUrl: serverEnv.CUSTOS_CORE_API_BASE_URL,
        });
        if (!recorded) {
          console.error("OIDC login activity capture failed");
        }
      }
      if (account?.id_token) {
        // Needed for the Keycloak end-session endpoint's id_token_hint.
        (token as { idToken?: string }).idToken = account.id_token;
      }
      return token;
    },
    async session({ session, token }) {
      const t = token as {
        accessToken?: string | null;
        idToken?: string | null;
        privileges?: Privilege[];
        sub?: string;
        custosUserId?: string;
      };
      if (t.accessToken) session.accessToken = t.accessToken;
      if (t.idToken) session.idToken = t.idToken;
      session.privileges = t.privileges ?? [];
      if (session.user) {
        // Prefer the backend-resolved Custos user_id over the OIDC sub so
        // downstream API callers get a value that matches users.id.
        session.user.id = t.custosUserId ?? t.sub ?? session.user.id;
        session.user.privileges = t.privileges ?? [];
      }
      return session;
    },
  },
};

export const { auth, handlers, signIn, signOut } = NextAuth(authConfig);
