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

import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import * as React from "react";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { useAcceptAgreement, useGateProfile } from "../queries";
import type { PendingAgreement } from "../schemas";
import { AgreementBlock, versionLine } from "./AgreementBlock";
import { GateActions, GateLayout, useGateSubmit } from "./GateLayout";

function WhatChanged({ agreement }: { agreement: PendingAgreement }) {
  return (
    <div className="space-y-2.5">
      <p className="text-xs font-medium text-muted-foreground">What changed</p>
      <div className="space-y-1">
        <p className="font-semibold">{agreement.title}</p>
        <p className="text-xs text-muted-foreground">{versionLine(agreement)}</p>
      </div>
      {agreement.change_note ? <p className="leading-relaxed">{agreement.change_note}</p> : null}
    </div>
  );
}

function AcceptForm({ agreement }: { agreement: PendingAgreement }) {
  const accept = useAcceptAgreement();
  const { error, run } = useGateSubmit();
  const [agreed, setAgreed] = React.useState(false);

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    void run(() => accept.mutateAsync(agreement.id));
  }

  return (
    <form onSubmit={onSubmit} noValidate className="flex w-full max-w-[760px] flex-col gap-7">
      <div className="space-y-1.5">
        <h1 className="font-display text-[28px] font-bold leading-tight">
          Accept the updated agreement
        </h1>
        <p className="text-muted-foreground">
          The user agreement changed. Accept the new version to keep using Custos.
        </p>
      </div>
      <AgreementBlock agreement={agreement} checked={agreed} onCheckedChange={setAgreed} />
      <GateActions
        submitLabel="Accept and continue"
        canSubmit={agreed}
        isPending={accept.isPending}
        error={error}
        disabledHelp="Accept and continue turns on once you check the box."
      />
    </form>
  );
}

export function PolicyUpdateGate() {
  const router = useRouter();
  const { update } = useSession();
  const profile = useGateProfile();
  const agreement = profile.data?.pending;
  const stale = profile.isSuccess && !agreement;
  const [tried, setTried] = React.useState(false);
  const [refreshing, startRefresh] = React.useTransition();

  const reload = React.useCallback(() => {
    setTried(true);
    startRefresh(async () => {
      await update({});
      startRefresh(() => router.refresh());
    });
  }, [update, router]);

  // Nothing pending means the session flag is stale. Re-read it once to reach the portal.
  React.useEffect(() => {
    if (stale && !tried) reload();
  }, [stale, tried, reload]);

  return (
    <GateLayout
      heading="Policy update"
      intro="The user agreement you accepted has a new version."
      aside={agreement ? <WhatChanged agreement={agreement} /> : null}
    >
      {profile.isLoading || (stale && (!tried || refreshing)) ? (
        <CardSkeleton />
      ) : stale ? (
        <ErrorState message="Couldn't load the updated agreement." onRetry={reload} />
      ) : profile.error ? (
        <ErrorState message={profile.error.message} onRetry={() => profile.refetch()} />
      ) : agreement ? (
        <AcceptForm agreement={agreement} />
      ) : null}
    </GateLayout>
  );
}
