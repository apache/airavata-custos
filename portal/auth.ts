import NextAuth, { type NextAuthConfig } from "next-auth";
import Credentials from "next-auth/providers/credentials";

const oidcConfigured = Boolean(process.env.OIDC_ISSUER_URL);

const providers: NextAuthConfig["providers"] = oidcConfigured
  ? [
      {
        id: "oidc",
        name: "OIDC",
        type: "oidc",
        issuer: process.env.OIDC_ISSUER_URL,
        clientId: process.env.OIDC_CLIENT_ID,
        clientSecret: process.env.OIDC_CLIENT_SECRET,
        authorization: { params: { scope: "openid profile email" } },
      },
    ]
  : [
      // Dev fallback: pairs with the signer's DEV_MODE so contributors can run
      // the portal without standing up a real OIDC provider. Never registers
      // when OIDC_ISSUER_URL is set.
      Credentials({
        id: "dev",
        name: "Dev",
        credentials: {},
        authorize: async () => ({
          id: "dev-user",
          name: "Dev User",
          email: process.env.DEV_DEFAULT_EMAIL ?? "dev@example.com",
        }),
      }),
    ];

export const { handlers, auth, signIn, signOut } = NextAuth({
  providers,
  callbacks: {
    async jwt({ token, account }) {
      if (account?.access_token) {
        token.accessToken = account.access_token;
      } else if (!token.accessToken && !oidcConfigured) {
        // Placeholder bearer accepted by the signer when DEV_MODE=true.
        token.accessToken = "dev-token";
      }
      return token;
    },
    async session({ session, token }) {
      if (token.accessToken) {
        session.accessToken = token.accessToken as string;
      }
      return session;
    },
  },
});
