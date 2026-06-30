// Interactive sign-in form for the custom /signin page.
//
// Composes the same brand vocabulary as PortalLayout — CUSTOS wordmark,
// blue accent, neutral surfaces, the support-card pattern — into a
// pre-auth layout (no sidebar or header, since neither applies to a
// signed-out user). The actual provider hand-off goes through NextAuth's
// signIn() with the configured provider id (`cilogon` in production, the
// `dev` Credentials fallback when OIDC_ISSUER_URL is unset).
"use client";

import { useEffect, useState } from "react";
import { signIn, useSession } from "next-auth/react";
import { useRouter, useSearchParams } from "next/navigation";
import { Headphones, ShieldCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ORG_NAME, PORTAL_NAME, SUPPORT_EMAIL } from "../../lib/config";

// Friendly copy for the Auth.js error codes that surface as ?error=... on
// the sign-in URL. Anything not in this map falls back to a generic line.
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
};

// Default provider id when OIDC is configured. Mirrors auth.ts so this page
// works whether the runtime registers the OIDC or the dev fallback.
const DEFAULT_PROVIDER = "cilogon";

export function SignInForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { status } = useSession();
  const [submitting, setSubmitting] = useState(false);

  const callbackUrl = searchParams.get("callbackUrl") ?? "/";
  const errorCode = searchParams.get("error");
  const errorMessage = errorCode
    ? ERROR_MESSAGES[errorCode] ??
      "Something went wrong during sign-in. Please try again."
    : null;

  // If the visitor is already authenticated (e.g. they hit /signin directly
  // after a previous session), skip the form and send them to their target.
  useEffect(() => {
    if (status === "authenticated") {
      router.replace(callbackUrl);
    }
  }, [status, callbackUrl, router]);

  async function handleSignIn() {
    setSubmitting(true);
    await signIn(DEFAULT_PROVIDER, { callbackUrl });
  }

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
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-blue-50">
              <ShieldCheck className="h-6 w-6 text-blue-700" />
            </div>

            <h1 className="mt-6 text-2xl font-semibold leading-tight tracking-normal">
              Sign in to {PORTAL_NAME}
            </h1>
            <p className="mt-2 text-sm leading-snug text-neutral-600">
              {ORG_NAME}. Use your institutional identity to manage SSH
              certificates, view allocations, and access portal tools.
            </p>

            {errorMessage && (
              <div
                className="mt-6 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700"
                role="alert"
              >
                {errorMessage}
              </div>
            )}

            <Button
              className="mt-8 h-11 w-full bg-blue-700 text-white hover:bg-blue-800"
              size="lg"
              onClick={handleSignIn}
              disabled={submitting || status === "loading"}
            >
              {submitting ? "Redirecting…" : "Continue with CILogon"}
            </Button>

            <p className="mt-4 text-center text-xs text-neutral-500">
              You&apos;ll be redirected to your identity provider to
              complete sign-in.
            </p>
          </div>

          <div className="flex items-center gap-3 rounded-xl border border-neutral-200 bg-white px-5 py-4 text-sm shadow-sm">
            <div className="flex h-9 w-9 items-center justify-center rounded-md bg-neutral-100">
              <Headphones className="h-5 w-5" />
            </div>
            <div className="flex-1">
              <div className="font-semibold">Need help signing in?</div>
              <p className="text-neutral-500">
                Reach out to portal support and we&apos;ll help you get
                access.
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
