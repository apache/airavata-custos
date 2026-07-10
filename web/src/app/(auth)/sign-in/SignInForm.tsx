// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

"use client";

import { Button } from "@/shared/ui/button";
import { signIn } from "next-auth/react";
import { useSearchParams } from "next/navigation";
import { useState } from "react";
import { AuthErrorAlert } from "./AuthErrorAlert";

export function SignInForm() {
  const params = useSearchParams();
  const callbackUrl = params.get("callbackUrl") ?? "/";
  const externalError = params.get("error");
  const [submitting, setSubmitting] = useState(false);

  const submitOidc = async () => {
    setSubmitting(true);
    await signIn("oidc", { callbackUrl });
  };

  return (
    <div className="flex flex-col gap-4">
      <AuthErrorAlert code={externalError} />
      <Button onClick={submitOidc} disabled={submitting} className="w-full">
        {submitting ? "Redirecting…" : "Sign in with Custos"}
      </Button>
    </div>
  );
}
