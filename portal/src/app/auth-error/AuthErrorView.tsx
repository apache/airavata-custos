// Recovery UI shown when a CILogon/OIDC sign-in attempt fails. Decodes
// Auth.js's `?error=<code>` query param into a friendly explanation and
// offers two unambiguous user-driven actions:
//   - "Try signing in again" → restart the OIDC flow directly with the
//     same provider id.
//   - "Clear session and start over" → sign out (drops any partial JWT
//     cookie that may be in a bad state) and bounce back to /signin.
//
// Both actions are explicit button clicks; the page does NOT auto-redirect,
// so refreshing it cannot trap the user in a redirect loop the way the
// stock /api/auth/error page did.
"use client";

import { useState } from "react";
import { signIn, signOut } from "next-auth/react";
import { useSearchParams } from "next/navigation";
import { AlertTriangle, Headphones } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PORTAL_NAME, SUPPORT_EMAIL } from "../../lib/config";

// Friendly copy for the Auth.js error codes that surface as ?error=... on
// the auth-error URL. Kept in lockstep with the same map in SignInForm so
// users see consistent messaging regardless of which page Auth.js routes
// them to.
const ERROR_MESSAGES: Record<string, string> = {
  Configuration:
    "The sign-in service is misconfigured. Please contact the portal administrator.",
  AccessDenied:
    "Your identity provider declined the sign-in request. Try again or use a different account.",
  Verification:
    "That sign-in link has expired or already been used. Please try again.",
  OAuthSignin:
    "We couldn't reach the identity provider. Check your connection and try again.",
  OAuthCallback:
    "The identity provider rejected the sign-in. Try again or contact support.",
  OAuthAccountNotLinked:
    "An account with this email already exists with a different provider.",
  CallbackRouteError:
    "Something went wrong finishing the sign-in. Try again, and if the problem persists clear your session and start over.",
  SessionTokenError:
    "Your session token couldn't be validated. Clear your session and sign in again.",
};

// Match SignInForm so the OIDC handoff goes straight to the provider
// instead of round-tripping through /signin first.
const DEFAULT_PROVIDER = "cilogon";
const DEFAULT_CALLBACK = "/signer/certificates";

export function AuthErrorView() {
  const searchParams = useSearchParams();
  const [retrying, setRetrying] = useState(false);
  const [clearing, setClearing] = useState(false);

  const errorCode = searchParams.get("error");
  const errorMessage =
    (errorCode && ERROR_MESSAGES[errorCode]) ??
    "Something went wrong during sign-in. Please try again.";

  async function handleRetry() {
    setRetrying(true);
    await signIn(DEFAULT_PROVIDER, { callbackUrl: DEFAULT_CALLBACK });
  }

  async function handleClearSession() {
    setClearing(true);
    // signOut clears the (possibly partial) JWT cookie before returning to
    // /signin where the user can start a fresh OIDC handshake.
    await signOut({ callbackUrl: "/signin" });
  }

  const busy = retrying || clearing;

  return (
    <div className="flex min-h-screen flex-col bg-[#f1f1f1] text-neutral-950">
      <header className="flex h-20 items-center border-b border-neutral-200 bg-white px-8">
        <span className="text-2xl font-extrabold uppercase tracking-normal text-blue-800">
          {PORTAL_NAME}
        </span>
      </header>

      <main className="flex flex-1 items-center justify-center px-6 py-10">
        <div className="w-full max-w-md space-y-6">
          <div className="rounded-2xl bg-white p-10 shadow-[0_8px_24px_rgba(0,0,0,0.08)]">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-rose-50">
              <AlertTriangle className="h-6 w-6 text-rose-600" />
            </div>

            <h1 className="mt-6 text-2xl font-semibold leading-tight tracking-normal">
              Sign-in didn&apos;t complete
            </h1>
            <p className="mt-2 text-sm leading-snug text-neutral-600">
              We couldn&apos;t finish signing you in to {PORTAL_NAME}. You can
              try again or clear your session and start over.
            </p>

            <div
              className="mt-6 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700"
              role="alert"
            >
              {errorMessage}
              {errorCode && (
                <span className="ml-2 text-xs text-rose-600/80">
                  (code: {errorCode})
                </span>
              )}
            </div>

            <Button
              className="mt-8 h-11 w-full bg-blue-700 text-white hover:bg-blue-800"
              size="lg"
              onClick={handleRetry}
              disabled={busy}
            >
              {retrying ? "Redirecting…" : "Try signing in again"}
            </Button>

            <Button
              className="mt-3 h-11 w-full"
              size="lg"
              variant="outline"
              onClick={handleClearSession}
              disabled={busy}
            >
              {clearing ? "Clearing session…" : "Clear session and start over"}
            </Button>

            <p className="mt-4 text-center text-xs text-neutral-500">
              Retrying restarts the CILogon login flow. Clearing your session
              also drops any partial sign-in cookie.
            </p>
          </div>

          <div className="flex items-center gap-3 rounded-xl border border-neutral-200 bg-white px-5 py-4 text-sm shadow-sm">
            <div className="flex h-9 w-9 items-center justify-center rounded-md bg-neutral-100">
              <Headphones className="h-5 w-5" />
            </div>
            <div className="flex-1">
              <div className="font-semibold">Still stuck?</div>
              <p className="text-neutral-500">
                Contact portal support and include the error code shown above.
              </p>
            </div>
            <a
              className="inline-flex h-9 items-center justify-center rounded-lg bg-black px-4 text-xs font-semibold text-white hover:bg-neutral-800"
              href={`mailto:${SUPPORT_EMAIL}`}
            >
              Contact
            </a>
          </div>
        </div>
      </main>
    </div>
  );
}
