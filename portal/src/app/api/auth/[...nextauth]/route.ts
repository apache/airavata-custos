/**
 * NextAuth route mount.
 *
 * The `[...nextauth]` catch-all under /api/auth/ delegates every NextAuth
 * endpoint (/signin, /callback/<provider>, /csrf, /session, /signout, ...)
 * to the handlers built from auth.ts. The actual sign-in logic, code
 * exchange, and session cookie handling all live inside `next-auth`; this
 * file only attaches them to the App Router.
 */
import { handlers } from "../../../../../auth";

export const { GET, POST } = handlers;
