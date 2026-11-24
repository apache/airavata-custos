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

import { FileCheck, Server, User } from "lucide-react";
import * as React from "react";
import { ApiError } from "@/shared/api/client";
import { useDebounce } from "@/shared/hooks/useDebounce";
import { Button } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { CardSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { useCompleteOnboarding, useGateProfile, useUsernameCheck } from "../queries";
import type { PendingAgreement } from "../schemas";
import { AgreementBlock } from "./AgreementBlock";
import { GateActions, GateLayout, useGateSubmit } from "./GateLayout";

const STEPS = [
  { icon: User, label: "Choose your username" },
  { icon: FileCheck, label: "Accept the agreement" },
  { icon: Server, label: "We set up your account" },
];

function Steps() {
  return (
    <div className="space-y-3">
      <ul className="space-y-2.5">
        {STEPS.map(({ icon: Icon, label }) => (
          <li key={label} className="flex items-center gap-2.5">
            <Icon aria-hidden="true" className="size-4 shrink-0 text-muted-foreground" />
            {label}
          </li>
        ))}
      </ul>
      <p className="text-xs text-muted-foreground">Setup starts as soon as you continue.</p>
    </div>
  );
}

function UsernameStatus({
  username,
  checking,
  failed,
  valid,
  available,
  suggestion,
  onUse,
}: {
  username: string;
  checking: boolean;
  failed: boolean;
  valid?: boolean | undefined;
  available?: boolean | undefined;
  suggestion?: string | undefined;
  onUse: (name: string) => void;
}) {
  let status: React.ReactNode = null;
  if (!username) status = null;
  else if (checking) status = <span className="text-xs text-muted-foreground">Checking…</span>;
  else if (failed)
    status = (
      <span className="text-xs text-destructive">Couldn't check this username. Try again.</span>
    );
  else if (valid === false)
    status = (
      <span className="text-xs text-destructive">
        Use lowercase letters, digits, - or _, starting with a letter.
      </span>
    );
  else if (available === false) status = <StatusBadge variant="rejected" label="Taken" />;
  else if (available) status = <StatusBadge variant="approved" label="Available" />;

  return (
    <div className="flex min-h-5 flex-wrap items-center gap-2">
      <output id="username-status">{status}</output>
      {!checking && available === false && suggestion ? (
        <Button
          type="button"
          variant="link"
          size="xs"
          className="px-0"
          onClick={() => onUse(suggestion)}
        >
          Use {suggestion}
        </Button>
      ) : null}
    </div>
  );
}

function OnboardingForm({
  initialUsername,
  editable,
  agreement,
}: {
  initialUsername: string;
  editable: boolean;
  agreement: PendingAgreement | null;
}) {
  const submit = useCompleteOnboarding();
  const { error, run } = useGateSubmit("/setup");
  const [username, setUsername] = React.useState(initialUsername);
  const [agreed, setAgreed] = React.useState(false);
  const debounced = useDebounce(username, 300);
  const check = useUsernameCheck(debounced, editable);
  const checking = editable && (username !== debounced || check.isFetching);
  const usernameOk =
    !editable ||
    (Boolean(username) && !checking && check.data?.valid === true && check.data.available);
  const canSubmit = usernameOk && (agreed || !agreement);

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    void run(
      () => submit.mutateAsync({ username, accepted: agreement ? [agreement.id] : [] }),
      (err) => {
        // Someone took the name between the check and the submit; re-check to show it.
        if (!(err instanceof ApiError && err.status === 409)) return false;
        void check.refetch();
        return true;
      },
    );
  }

  return (
    <form onSubmit={onSubmit} noValidate className="flex w-full max-w-[760px] flex-col gap-7">
      <div className="space-y-1.5">
        <h1 className="font-display text-[28px] font-bold leading-tight">Set up your account</h1>
        <p className="text-muted-foreground">
          {initialUsername
            ? "Choose your username and accept the user agreement."
            : agreement
              ? "Accept the user agreement to continue."
              : "Continue to finish setting up your account."}
        </p>
      </div>

      {/* Users without a cluster account have no username to choose. */}
      {initialUsername ? (
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="username">Username</Label>
          <Input
            id="username"
            className="h-11"
            value={username}
            readOnly={!editable}
            autoComplete="off"
            spellCheck={false}
            aria-describedby="username-status username-help"
            aria-invalid={check.data?.valid === false || check.data?.available === false}
            onChange={(e) => setUsername(e.target.value)}
          />
          {editable ? (
            <UsernameStatus
              username={username}
              checking={checking}
              failed={check.isError}
              valid={check.data?.valid}
              available={check.data?.available}
              suggestion={check.data?.suggestion}
              onUse={setUsername}
            />
          ) : null}
          <p id="username-help" className="text-xs text-muted-foreground">
            Used for your cluster account. It cannot be changed later.
          </p>
        </div>
      ) : null}

      {agreement ? (
        <AgreementBlock agreement={agreement} checked={agreed} onCheckedChange={setAgreed} />
      ) : null}

      <GateActions
        submitLabel="Continue"
        canSubmit={canSubmit}
        isPending={submit.isPending}
        error={error}
        disabledHelp={
          !editable
            ? "Continue turns on once you check the box."
            : agreement
              ? "Continue turns on once your username is available and you check the box."
              : "Continue turns on once your username is available."
        }
      />
    </form>
  );
}

export function OnboardingGate() {
  const profile = useGateProfile();
  return (
    <GateLayout
      heading="Welcome to Custos"
      intro="Your sign-in worked. One short step before you reach your dashboard."
      aside={<Steps />}
    >
      {profile.isLoading ? (
        <CardSkeleton />
      ) : profile.error ? (
        <ErrorState message={profile.error.message} onRetry={() => profile.refetch()} />
      ) : profile.data ? (
        <OnboardingForm
          initialUsername={profile.data.onboarding?.username ?? ""}
          editable={profile.data.onboarding?.username_editable !== false}
          agreement={profile.data.pending}
        />
      ) : null}
    </GateLayout>
  );
}
