/**
 * NextAuth (Auth.js v5) configuration for the Custos portal.
 *
 * Two provider configurations are possible at startup:
 *  - When `OIDC_ISSUER_URL` is set, a single OIDC provider is registered
 *    (provider id `cilogon`) and used as the only sign-in option.
 *  - When `OIDC_ISSUER_URL` is unset, a Credentials-based "Dev" provider is
 *    registered so contributors can exercise the portal locally against a
 *    signer service started with `DEV_MODE=true`.
 *
 * Anything downstream that needs the issuer's access token reads it off the
 * session as `session.accessToken`, populated by the `jwt` and `session`
 * callbacks below. The `/api/v1` proxy is the primary consumer.
 */
import NextAuth, { type NextAuthConfig } from "next-auth";
import Credentials from "next-auth/providers/credentials";

const oidcConfigured = Boolean(process.env.OIDC_ISSUER_URL);

// A public OIDC client (no shared secret) authenticates the token exchange
// with PKCE instead. Detect it from the absence of `OIDC_CLIENT_SECRET` so
// the same code supports both confidential and public clients.
const isPublicClient = !process.env.OIDC_CLIENT_SECRET;

// CILogon exposes its custom userinfo claims behind the `org.cilogon.userinfo`
// scope. Other OIDC issuers (notably Keycloak realms) reject unknown scopes
// with `invalid_scope`, so the scope set is overridable via env.
const oidcScope =
  process.env.OIDC_SCOPE ?? "openid profile email org.cilogon.userinfo";

const providers: NextAuthConfig["providers"] = oidcConfigured
  ? [
      {
        // The provider id is part of the callback path
        // (/api/auth/callback/<id>) that must be registered at the OIDC
        // provider, so changing it is a breaking change for deployments.
        id: "cilogon",
        name: "CILogon",
        type: "oidc",
        issuer: process.env.OIDC_ISSUER_URL,
        clientId: process.env.OIDC_CLIENT_ID,
        clientSecret: process.env.OIDC_CLIENT_SECRET,
        authorization: { params: { scope: oidcScope } },
        // Tells openid-client to send no client credentials at the token
        // endpoint. The matching client at the IdP must be configured as
        // "public" (PKCE-only).
        client: isPublicClient
          ? { token_endpoint_auth_method: "none" }
          : undefined,
      },
    ]
  : [
      // Dev fallback. Pairs with the signer's DEV_MODE so contributors can
      // run the portal without standing up a real OIDC provider. Never
      // registers when OIDC_ISSUER_URL is set.
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
  // The portal is typically deployed behind a reverse proxy and run locally
  // on a non-standard host; trust the inbound Host header rather than fixing
  // a Vercel-style URL.
  trustHost: true,
  // Replace Auth.js's stock provider chooser with a branded portal page.
  // Both the unauthenticated redirect in PortalLayout and any explicit
  // signIn() call without a provider id route here.
  pages: {
    signIn: "/signin",
  },
  callbacks: {
    // The `account` object is only present on the first JWT pass right after
    // a successful sign-in. Persist the issuer's access token onto the JWT
    // then so subsequent requests can read it from the session.
    async jwt({ token, account }) {
      if (account?.access_token) {
        token.accessToken = account.access_token;
      } else if (!token.accessToken && !oidcConfigured) {
        // Placeholder bearer the signer accepts when DEV_MODE=true.
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
