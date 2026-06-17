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
};

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
