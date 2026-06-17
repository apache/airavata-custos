"use client";

import { Button } from "@/shared/ui/button";
import { signIn } from "next-auth/react";
import { useSearchParams } from "next/navigation";
import { useState } from "react";

export function SignInForm() {
  const params = useSearchParams();
  const callbackUrl = params.get("callbackUrl") ?? "/";
  const [submitting, setSubmitting] = useState(false);

  const onClick = async () => {
    setSubmitting(true);
    try {
      await signIn("oidc", { callbackUrl });
    } catch {
      // signIn normally redirects on success; reset on failure so the button
      // doesn't stay disabled forever.
      setSubmitting(false);
    }
  };

  return (
    <Button
      onClick={onClick}
      disabled={submitting}
      aria-busy={submitting}
      className="w-full"
    >
      {submitting ? "Signing in…" : "Sign in"}
    </Button>
  );
}
