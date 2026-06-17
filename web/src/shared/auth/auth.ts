import NextAuth, { type NextAuthConfig } from "next-auth";
import { serverEnv } from "@/lib/env";
import type { Privilege } from "@/features/core/identity/types";

const oidcProvider = {
  id: "oidc",
  name: "Sign in",
  type: "oidc" as const,
  issuer: serverEnv.OIDC_ISSUER_URL,
  clientId: serverEnv.OIDC_CLIENT_ID,
  clientSecret: serverEnv.OIDC_CLIENT_SECRET,
  authorization: { params: { scope: serverEnv.OIDC_SCOPES } },
  // PKCE protects code interception, state binds the response to this browser
  // session, nonce binds the id_token. id_token is the bearer the backend
  // accepts, so replay protection matters here.
  checks: ["pkce", "state", "nonce"] as ("pkce" | "state" | "nonce")[],
};

// Privileges aren't an OIDC claim — fetch from the backend so the layout gate sees them.
async function fetchPrivileges(bearer: string): Promise<Privilege[]> {
  try {
    const res = await fetch(`${serverEnv.CUSTOS_CORE_API_BASE_URL}/user/privileges`, {
      headers: { Authorization: `Bearer ${bearer}` },
      cache: "no-store",
    });
    if (!res.ok) {
      console.error("fetchPrivileges: backend returned", res.status, res.statusText);
      return [];
    }
    const data = (await res.json()) as { privileges?: Privilege[] };
    return data.privileges ?? [];
  } catch (err) {
    console.error("fetchPrivileges: request failed", err);
    return [];
  }
}

export const authConfig: NextAuthConfig = {
  trustHost: true,
  secret: serverEnv.NEXTAUTH_SECRET,
  session: { strategy: "jwt" },
  pages: { signIn: "/sign-in" },
  providers: [oidcProvider],
  callbacks: {
    async jwt({ token, user, account }) {
      if (user) {
        (token as { privileges?: Privilege[] }).privileges = user.privileges ?? [];
      }
      if (account?.id_token) {
        (token as { accessToken?: string }).accessToken = account.id_token;
        (token as { privileges?: Privilege[] }).privileges = await fetchPrivileges(account.id_token);
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
