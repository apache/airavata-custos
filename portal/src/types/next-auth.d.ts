/**
 * Type augmentations for next-auth.
 *
 * The OIDC access token is persisted onto the JWT in the `jwt()` callback
 * and copied to the session in the `session()` callback (see auth.ts).
 * NextAuth's stock types don't know about that field, so we extend them
 * here once for the whole app.
 */
import "next-auth";
import "next-auth/jwt";

declare module "next-auth" {
  interface Session {
    accessToken?: string;
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    accessToken?: string;
  }
}
