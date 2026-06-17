"use client";

import { Button } from "@/shared/ui/button";
import { signIn } from "next-auth/react";
import { useSearchParams } from "next/navigation";
import { useState } from "react";

export function SignInForm() {
  const params = useSearchParams();
  const callbackUrl = params.get("callbackUrl") ?? "/";
  const [submitting, setSubmitting] = useState(false);

  const onSubmit = async () => {
    setSubmitting(true);
    await signIn("oidc", { callbackUrl });
  };

  return (
    <Button onClick={onSubmit} disabled={submitting} className="w-full">
      Sign in
    </Button>
  );
}
