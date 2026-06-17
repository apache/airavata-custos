import { Suspense } from "react";
import { SignInForm } from "./SignInForm";

export const metadata = {
  title: "Sign in · Custos Portal",
};

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-muted px-6 py-12">
      <div className="w-full max-w-md rounded-2xl border border-border bg-card p-8 shadow-sm">
        <div className="mb-6 flex flex-col items-start gap-1">
          <span className="font-display text-2xl font-extrabold uppercase tracking-tight text-brand">
            Custos
          </span>
          <h1 className="text-xl font-semibold tracking-tight">Sign in</h1>
          <p className="text-sm text-muted-foreground">
            Sign in with your Custos account.
          </p>
        </div>
        <Suspense fallback={null}>
          <SignInForm />
        </Suspense>
      </div>
    </div>
  );
}
