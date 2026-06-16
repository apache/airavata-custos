import { Suspense } from "react";
import { serverEnv } from "@/lib/env";
import { SignInForm } from "./SignInForm";

export const metadata = {
  title: "Sign in · Custos Portal",
};

export default function SignInPage() {
  const mode = serverEnv.PORTAL_AUTH_MODE;
  return (
    <div className="flex min-h-screen items-center justify-center bg-muted px-6 py-12">
      <div className="w-full max-w-md rounded-2xl border border-border bg-card p-8 shadow-sm">
        <div className="mb-6 flex flex-col items-start gap-1">
          <span className="font-display text-2xl font-extrabold uppercase tracking-tight text-brand">
            Custos
          </span>
          <h1 className="text-xl font-semibold tracking-tight">Sign in</h1>
          <p className="text-sm text-muted-foreground">
            {mode === "oidc" ? "Sign in with your Custos account." : "Pick a dev privilege level."}
          </p>
        </div>
        <Suspense fallback={null}>
          <SignInForm mode={mode} />
        </Suspense>
      </div>
    </div>
  );
}
