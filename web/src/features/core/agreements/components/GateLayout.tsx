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

import { ArrowRight, LogOut } from "lucide-react";
import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import * as React from "react";
import { useSignOut } from "@/shared/auth/useSignOut";
import { brand } from "@/shared/brand";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import { Button } from "@/shared/ui/button";
import { Skeleton } from "@/shared/ui/skeleton";

function IdentityCard() {
  const { data: session } = useSession();
  const user = session?.user;
  const email = user?.email ?? "";
  const name = user?.name || email;
  return (
    <div className="flex items-center gap-3 rounded-xl border border-border bg-card px-4 py-3.5 shadow-xs">
      {user ? (
        <>
          <Avatar className="h-10 w-10 border-2 border-brand">
            <AvatarFallback>{name.slice(0, 1).toUpperCase()}</AvatarFallback>
          </Avatar>
          <div className="flex min-w-0 flex-col gap-0.5">
            <span className="font-semibold text-foreground">{name}</span>
            <span className="truncate text-[13px] text-muted-foreground">{email}</span>
          </div>
        </>
      ) : (
        <Skeleton className="h-10 w-full" />
      )}
    </div>
  );
}

// Runs a gate action, then re-reads the session so the server layout lets the user through.
export function useGateSubmit(next?: string) {
  const router = useRouter();
  const { update } = useSession();
  const [error, setError] = React.useState<string | null>(null);

  async function run(action: () => Promise<unknown>, handled?: (err: unknown) => boolean) {
    setError(null);
    try {
      await action();
      // Without an argument update() only re-reads the cookie and the flags stay stale.
      await update({});
      if (next) router.replace(next);
      else router.refresh();
    } catch (err) {
      if (!handled?.(err)) setError(err instanceof Error ? err.message : "Request failed");
    }
  }

  return { error, run };
}

export function GateLayout({
  heading,
  intro,
  aside,
  children,
}: {
  heading: string;
  intro: string;
  aside: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen flex-col bg-background text-sm text-foreground md:flex-row">
      <aside className="border-b border-border bg-muted px-4 py-5 md:flex md:w-2/5 md:shrink-0 md:justify-end md:border-r md:border-b-0 md:px-12 md:py-14">
        <div className="flex w-full flex-col gap-5 md:max-w-[360px] md:gap-7">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={brand.logo} alt={brand.name} className="h-8 w-auto self-start" />
          <div className="space-y-1.5">
            <p className="font-display text-[22px] font-bold leading-tight">{heading}</p>
            <p className="text-muted-foreground">{intro}</p>
          </div>
          <IdentityCard />
          {aside}
        </div>
      </aside>
      <main className="min-w-0 flex-1 px-4 pt-6 pb-10 md:px-12 md:py-14">{children}</main>
    </div>
  );
}

export function GateActions({
  submitLabel,
  canSubmit,
  isPending,
  disabledHelp,
  error,
}: {
  submitLabel: string;
  canSubmit: boolean;
  isPending: boolean;
  disabledHelp: string;
  error: string | null;
}) {
  const { signOut, isPending: signingOut } = useSignOut();
  return (
    <div className="space-y-2">
      {error ? (
        <p
          role="alert"
          className="mb-5 rounded-lg bg-[color:var(--tone-error-bg)] px-4 py-3 text-[color:var(--tone-error-fg)]"
        >
          {error}
        </p>
      ) : null}
      <div className="flex flex-col-reverse gap-3 sm:flex-row sm:items-center sm:justify-between">
        <Button
          type="button"
          variant="ghost"
          size="lg"
          className="text-muted-foreground"
          disabled={signingOut}
          onClick={() => signOut()}
        >
          <LogOut aria-hidden="true" />
          Sign out
        </Button>
        <Button
          type="submit"
          size="lg"
          disabled={!canSubmit || isPending}
          aria-describedby={canSubmit ? undefined : "gate-submit-help"}
        >
          {submitLabel}
          <ArrowRight aria-hidden="true" />
        </Button>
      </div>
      {canSubmit ? null : (
        <p id="gate-submit-help" className="text-xs text-muted-foreground sm:text-right">
          {disabledHelp}
        </p>
      )}
    </div>
  );
}
