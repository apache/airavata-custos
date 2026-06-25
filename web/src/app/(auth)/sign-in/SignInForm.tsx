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
import { Label } from "@/shared/ui/label";
import { signIn } from "next-auth/react";
import { useSearchParams } from "next/navigation";
import { useState } from "react";

type Mode = "dev" | "oidc";
type Level = "viewer" | "manager" | "admin";

const LEVELS: Array<{ id: Level; label: string; hint: string }> = [
  { id: "viewer", label: "Viewer", hint: "Read-only (hpc:read)" },
  { id: "manager", label: "Manager", hint: "hpc:read/write, amie:read" },
  { id: "admin", label: "Admin", hint: "All privileges" },
];

export function SignInForm({ mode }: { mode: Mode }) {
  const params = useSearchParams();
  const callbackUrl = params.get("callbackUrl") ?? "/";
  const [level, setLevel] = useState<Level>("admin");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submitDev = async () => {
    setSubmitting(true);
    setError(null);
    const res = await signIn("credentials", {
      level,
      redirect: false,
      callbackUrl,
    });
    setSubmitting(false);
    if (!res || res.error) {
      setError(res?.error ?? "Sign in failed");
      return;
    }
    window.location.assign(res.url ?? callbackUrl);
  };

  const submitOidc = async () => {
    setSubmitting(true);
    await signIn("oidc", { callbackUrl });
  };

  if (mode === "oidc") {
    return (
      <Button onClick={submitOidc} disabled={submitting} className="w-full">
        Sign in with Custos
      </Button>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-2">
        <Label htmlFor="dev-level">Dev privilege level</Label>
        <select
          id="dev-level"
          value={level}
          onChange={(e) => setLevel(e.target.value as Level)}
          className="h-10 rounded-md border border-input bg-background px-3 text-sm"
        >
          {LEVELS.map((l) => (
            <option key={l.id} value={l.id}>
              {l.label} — {l.hint}
            </option>
          ))}
        </select>
      </div>
      <Button onClick={submitDev} disabled={submitting} className="w-full">
        Sign in
      </Button>
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  );
}
