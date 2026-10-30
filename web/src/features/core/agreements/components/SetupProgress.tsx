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

import { relativeTime } from "@/features/core/events/utils";
import type { SetupStep } from "@/generated/core/zod.gen";
import { cn } from "@/lib/utils";
import { Button } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { ArrowRight, Check, Copy, SquareTerminal } from "lucide-react";
import Link from "next/link";
import * as React from "react";
import { useGateProfile, useSetupSteps } from "../queries";
import { GateLayout } from "./GateLayout";

const LABELS: Record<string, string> = {
  registry: "Identity registered",
  cluster_account: "Cluster account created",
  project_group: "Project group ready",
};

type Group = { key: string; label: string; entries: SetupStep[] };
type State = "done" | "active" | "waiting";

// The first step is the onboarding the user just finished, so it is always done.
function groupSteps(steps: SetupStep[]): Group[] {
  const groups: Group[] = [{ key: "onboarding", label: "Username and agreement", entries: [] }];
  for (const step of steps) {
    const group = groups.find((g) => g.key === step.key);
    if (group) group.entries.push(step);
    else groups.push({ key: step.key, label: LABELS[step.key] ?? step.key, entries: [step] });
  }
  return groups;
}

const isDone = (g: Group) => g.entries.every((e) => e.done_at);

function doneAt(g: Group): string | undefined {
  return g.entries
    .map((e) => e.done_at)
    .filter((t): t is string => Boolean(t))
    .sort()
    .at(-1);
}

function Marker({ state }: { state: State }) {
  if (state === "done") {
    return (
      <span className="absolute top-px left-0 flex size-5 items-center justify-center rounded-full bg-[color:var(--tone-ok-fg)]">
        <Check aria-hidden="true" className="size-3 text-background" strokeWidth={3} />
      </span>
    );
  }
  return (
    <span
      className={cn(
        "absolute top-px left-0 flex size-5 items-center justify-center rounded-full border-2 bg-background",
        state === "active" ? "border-[color:var(--brand)]" : "border-muted-foreground/40",
      )}
    >
      {state === "active" ? (
        <span className="size-2 animate-pulse rounded-full bg-[color:var(--brand)] motion-reduce:animate-none" />
      ) : null}
    </span>
  );
}

function StepStatus({ state, at }: { state: State; at?: string | undefined }) {
  if (state === "done") {
    return (
      <span className="flex items-center gap-2 text-xs">
        <span className="font-medium text-[color:var(--tone-ok-fg)]">Done</span>
        {at ? <span className="text-muted-foreground">{relativeTime(at)}</span> : null}
      </span>
    );
  }
  return (
    <span className={cn("text-xs", state === "active" ? "font-medium" : "text-muted-foreground")}>
      {state === "active" ? "In progress" : "Waiting"}
    </span>
  );
}

function Timeline({ groups }: { groups: Group[] }) {
  const active = groups.findIndex((g) => !isDone(g));
  const done = groups.filter(isDone).length;
  return (
    <div className="flex flex-col gap-3">
      <p className="text-xs text-muted-foreground">
        {active === -1 ? `All ${groups.length} steps done` : `${done} of ${groups.length} done`}
      </p>
      <ol aria-label="Setup steps" className="flex flex-col">
        {groups.map((g, i) => {
          const state: State = isDone(g) ? "done" : i === active ? "active" : "waiting";
          const named = g.entries.filter((e) => e.name);
          return (
            <li
              key={g.key}
              aria-current={state === "active" ? "step" : undefined}
              className="relative min-h-12 pb-3 pl-8"
            >
              {i < groups.length - 1 ? (
                <span
                  aria-hidden="true"
                  className={cn(
                    "absolute top-[21px] -bottom-px left-[9px] w-0.5",
                    state === "done" ? "bg-[color:var(--tone-ok-fg)]" : "bg-border",
                  )}
                />
              ) : null}
              <Marker state={state} />
              <div className="flex flex-col">
                <span className="font-medium">{g.label}</span>
                <StepStatus state={state} at={doneAt(g)} />
                {named.length > 1 ? (
                  <ul className="mt-1 flex flex-col gap-0.5 text-xs text-muted-foreground">
                    {named.map((e) => (
                      <li key={e.name} className="flex gap-2">
                        <span className="font-mono">{e.name}</span>
                        <span>{e.done_at ? "Done" : "In progress"}</span>
                      </li>
                    ))}
                  </ul>
                ) : named[0] ? (
                  <span className="font-mono text-xs text-muted-foreground">{named[0].name}</span>
                ) : null}
              </div>
            </li>
          );
        })}
      </ol>
    </div>
  );
}

function SshCommand({ command }: { command: string }) {
  const [copied, setCopied] = React.useState(false);
  async function copy() {
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);
    } catch {
      setCopied(false);
    }
  }
  return (
    <div className="flex items-center gap-2">
      <input
        id="ssh-command"
        readOnly
        value={command}
        spellCheck={false}
        className="h-11 min-w-0 flex-1 rounded-lg border border-border bg-muted px-3 font-mono text-[13px] outline-none"
      />
      <Button type="button" variant="outline" size="lg" onClick={() => void copy()}>
        {copied ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
        {copied ? "Copied" : "Copy"}
      </Button>
    </div>
  );
}

export function SetupProgress() {
  const { data: steps, error, refetch } = useSetupSteps();
  const profile = useGateProfile();
  const username = profile.data?.onboarding?.username ?? "";
  const groups = groupSteps(steps ?? []);
  const ready = steps !== undefined && groups.every(isDone);
  const host = process.env.NEXT_PUBLIC_CLUSTER_LOGIN_HOST;

  const chosen = username ? (
    <div className="flex flex-col gap-3 rounded-xl border border-border bg-card px-5 py-4 shadow-xs">
      <span className="font-semibold">What you chose</span>
      <dl className="flex flex-col gap-0.5">
        <dt className="text-xs text-muted-foreground">Username</dt>
        <dd className="font-mono text-[13px]">{username}</dd>
      </dl>
    </div>
  ) : null;

  const connect = ready ? (
    <div className="flex flex-col gap-2.5 rounded-xl border border-border bg-card px-5 py-4 shadow-xs">
      <label htmlFor="ssh-command" className="flex items-center gap-2 font-semibold">
        <SquareTerminal aria-hidden="true" className="size-4" />
        Connect with SSH
      </label>
      {host && username ? (
        <>
          <SshCommand command={`ssh ${username}@${host}`} />
          <p className="text-xs text-muted-foreground">
            When asked, finish signing in with your institution's login.
          </p>
        </>
      ) : (
        <p className="text-[13px] text-muted-foreground">
          Sign in to the cluster as <span className="font-mono text-foreground">{username}</span>.
        </p>
      )}
    </div>
  ) : (
    <div className="flex flex-col gap-1.5 rounded-xl border border-border bg-muted px-5 py-4">
      <span className="flex items-center gap-2 font-semibold text-muted-foreground">
        <SquareTerminal aria-hidden="true" className="size-4" />
        Connect
      </span>
      <p className="text-[13px] text-muted-foreground">
        The SSH command will appear here when your cluster account is ready.
      </p>
    </div>
  );

  return (
    <GateLayout
      heading="Welcome to Custos"
      intro={
        ready
          ? "Every step finished. Your account is ready to use."
          : "Your account is being set up. Each step is ticked off as it finishes."
      }
      aside={error ? null : <Timeline groups={groups} />}
    >
      <div className="flex w-full max-w-[760px] flex-col gap-7">
        <div className="space-y-1.5">
          <h1 className="font-display text-[28px] font-bold leading-tight">
            {ready ? "Your account is ready" : "Setting up your account"}
          </h1>
          <p className="text-muted-foreground">
            {ready ? "Here is how to reach the cluster." : "This usually takes under a minute."}
          </p>
        </div>
        {error ? <ErrorState message={error.message} onRetry={() => refetch()} /> : null}
        {ready ? connect : chosen}
        {ready ? chosen : connect}
        {ready ? (
          <div className="flex justify-end">
            <Button size="lg" nativeButton={false} render={<Link href="/" />}>
              Continue to the portal
              <ArrowRight aria-hidden="true" />
            </Button>
          </div>
        ) : null}
      </div>
    </GateLayout>
  );
}
