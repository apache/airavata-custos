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
import Credentials from "next-auth/providers/credentials";
import Keycloak from "next-auth/providers/keycloak";
import { z } from "zod";
import { serverEnv } from "@/lib/env";
import type { Privilege } from "@/features/core/identity/types";
import { DEV_LEVEL_NAMES, DEV_LEVEL_PRIVILEGES } from "./devLevels";

export { DEV_LEVEL_PRIVILEGES, DEV_LEVEL_NAMES, type DevLevel } from "./devLevels";

const credentialsSchema = z.object({
  level: z.enum(["viewer", "manager", "admin"]),
});

const oidcEnabled = serverEnv.PORTAL_AUTH_MODE === "oidc";
const devEnabled = serverEnv.PORTAL_AUTH_MODE === "dev";

// Map dev levels to the seeded backend user IDs in
// dev-ops/compose/seeds/dev_users_and_roles.sql so X-Custos-User-Id resolves.
const DEV_LEVEL_BACKEND_USER_ID: Record<"viewer" | "manager" | "admin", string> = {
  viewer: "dev-researcher",
  manager: "dev-operator",
  admin: "dev-admin",
};

const credentialsProvider = Credentials({
  id: "credentials",
  name: "Dev credentials",
  credentials: { level: { label: "Level", type: "text" } },
  authorize: async (raw) => {
    const parsed = credentialsSchema.safeParse(raw);
    if (!parsed.success) return null;
    const level = parsed.data.level;
    const email = `${level}@custos.local`;
    return {
      id: DEV_LEVEL_BACKEND_USER_ID[level],
      email,
      name: DEV_LEVEL_NAMES[level],
      privileges: DEV_LEVEL_PRIVILEGES[level],
    };
  },
});

const providers: NextAuthConfig["providers"] = [
  ...(devEnabled ? [credentialsProvider] : []),
  ...(oidcEnabled
    ? [
        Keycloak({
          id: "oidc",
          issuer: serverEnv.OIDC_ISSUER_URL ?? "",
          clientId: serverEnv.OIDC_CLIENT_ID ?? "",
          clientSecret: serverEnv.OIDC_CLIENT_SECRET ?? "",
          authorization: { params: { scope: "openid email profile" } },
        }),
      ]
    : []),
];

export const authConfig: NextAuthConfig = {
  trustHost: true,
  secret: serverEnv.NEXTAUTH_SECRET,
  session: { strategy: "jwt" },
  pages: { signIn: "/sign-in" },
  providers,
  callbacks: {
    async jwt({ token, user, account }) {
      if (user) {
        (token as { privileges?: Privilege[] }).privileges = user.privileges ?? [];
      }
      if (account?.access_token) {
        (token as { accessToken?: string }).accessToken = account.access_token;
      } else if (!token.accessToken && devEnabled) {
        // Dev mode has no upstream token; the proxy uses this as a sentinel
        // so user-path forwarding succeeds without an IdP.
        (token as { accessToken?: string }).accessToken = "dev-token";
      }
      return token;
    },
    async session({ session, token }) {
      const t = token as { accessToken?: string | null; privileges?: Privilege[]; sub?: string };
      if (t.accessToken) session.accessToken = t.accessToken;
      session.privileges = t.privileges ?? [];
      if (session.user) {
        if (typeof t.sub === "string") session.user.id = t.sub;
        session.user.privileges = t.privileges ?? [];
      }
      return session;
    },
  },
};

export const { auth, handlers, signIn, signOut } = NextAuth(authConfig);
