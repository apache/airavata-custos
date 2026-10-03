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

import { CircleCheckIcon, CircleXIcon, RotateCcwIcon } from "lucide-react";
import * as React from "react";
import { ViewTraceLink } from "@/features/core/audit/components/ViewTraceLink";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CenteredSpinner } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { Button } from "@/shared/ui/button";
import { useUserName } from "../queries";
import type { DeliveryDetail } from "../types";
import {
  type HistoryStep,
  MAX_ATTEMPTS,
  absoluteTime,
  deliveryState,
  historyRuns,
  relativeTime,
} from "../utils";
import { DeliveryStatusBadge } from "./DeliveryStatusBadge";

export type DeliveryDetailDrawerProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  delivery: DeliveryDetail | undefined;
  runningSubscribers?: Set<string>;
  isLoading: boolean;
  error: Error | null;
  canRetry: boolean;
  // Opens with the retry confirmation already showing, for the Retry button in the list.
  confirmRetryOnOpen: boolean;
  isRetrying: boolean;
  onRetry: () => void;
  onRefresh: () => void;
};

function RetriedBy({ actorId }: { actorId: string }) {
  const name = useUserName(actorId);
  return <span className="font-medium">{name.data ?? (actorId || "an admin")}</span>;
}

function StepRow({ step }: { step: HistoryStep }) {
  if (step.kind === "retry") {
    return (
      <li className="flex gap-2 text-sm">
        <RotateCcwIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden />
        <div>
          <p>
            Retried by <RetriedBy actorId={step.actorId} /> after {step.previousAttempts}{" "}
            {step.previousAttempts === 1 ? "attempt" : "attempts"}
          </p>
          <time className="text-xs text-muted-foreground" dateTime={step.at}>
            {absoluteTime(step.at)}
          </time>
        </div>
      </li>
    );
  }
  if (step.kind === "other") {
    return (
      <li className="text-sm text-muted-foreground">
        {step.label} · {absoluteTime(step.at)}
      </li>
    );
  }
  const Icon = step.ok ? CircleCheckIcon : CircleXIcon;
  return (
    <li className="flex gap-2 text-sm">
      <Icon
        className={
          step.ok
            ? "mt-0.5 size-4 shrink-0 text-[color:var(--tone-ok-fg)]"
            : "mt-0.5 size-4 shrink-0 text-[color:var(--tone-error-fg)]"
        }
        aria-hidden
      />
      <div className="min-w-0">
        <p>
          Attempt {step.attempt} {step.ok ? "succeeded" : "failed"}
        </p>
        <time className="text-xs text-muted-foreground" dateTime={step.at}>
          {absoluteTime(step.at)}
        </time>
        {step.error ? (
          <p className="mt-1 break-words font-mono text-xs text-[color:var(--tone-error-fg)]">
            {step.error}
          </p>
        ) : null}
      </div>
    </li>
  );
}

function AttemptHistory({ delivery }: { delivery: DeliveryDetail }) {
  const runs = historyRuns(delivery.history);
  if (runs.length === 0) {
    return <p className="text-sm text-muted-foreground">No attempts yet.</p>;
  }
  return (
    <div className="space-y-4">
      {runs.map((run, i) => (
        // Runs never reorder, so the position is a stable key.
        // biome-ignore lint/suspicious/noArrayIndexKey: runs are append-only history
        <section key={i} aria-label={`Run ${i + 1}`} className="space-y-2">
          {runs.length > 1 ? (
            <h4 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Run {i + 1}
            </h4>
          ) : null}
          <ol className="space-y-3 border-l border-border pl-3">
            {run.map((step, j) => (
              <StepRow key={step.at + String(j)} step={step} />
            ))}
          </ol>
        </section>
      ))}
    </div>
  );
}

function RetryBox({
  confirmOnOpen,
  isRetrying,
  onRetry,
}: {
  confirmOnOpen: boolean;
  isRetrying: boolean;
  onRetry: () => void;
}) {
  const [confirming, setConfirming] = React.useState(confirmOnOpen);
  if (!confirming) {
    return (
      <div className="flex justify-end">
        <Button variant="outline" onClick={() => setConfirming(true)}>
          Retry
        </Button>
      </div>
    );
  }
  return (
    <section
      aria-label="Confirm retry"
      className="space-y-3 rounded-md border border-border bg-muted/30 p-3"
    >
      <p className="text-sm">
        Retry this delivery? It starts over with {MAX_ATTEMPTS} new attempts. The attempt history
        below is kept, and the retry is recorded with your name.
      </p>
      <div className="flex justify-end gap-2">
        <Button variant="ghost" onClick={() => setConfirming(false)} disabled={isRetrying}>
          Cancel
        </Button>
        <Button onClick={onRetry} disabled={isRetrying}>
          {isRetrying ? "Retrying..." : "Retry now"}
        </Button>
      </div>
    </section>
  );
}

export function DeliveryDetailDrawer({
  open,
  onOpenChange,
  delivery,
  runningSubscribers,
  isLoading,
  error,
  canRetry,
  confirmRetryOnOpen,
  isRetrying,
  onRetry,
  onRefresh,
}: DeliveryDetailDrawerProps) {
  return (
    <SideDrawer
      open={open}
      onOpenChange={onOpenChange}
      width="lg"
      title={delivery ? delivery.event.event_type : "Delivery"}
      description={delivery ? `Handled by ${delivery.subscriber}` : undefined}
    >
      {isLoading ? (
        <CenteredSpinner label="Loading delivery" />
      ) : error ? (
        <ErrorState message={error.message || "Could not load the delivery."} onRetry={onRefresh} />
      ) : !delivery ? (
        <p className="text-sm text-muted-foreground">Delivery not found.</p>
      ) : (
        <div className="space-y-5">
          <div className="flex flex-wrap items-center gap-3">
            <DeliveryStatusBadge state={deliveryState(delivery, runningSubscribers)} />
            <span className="text-sm tabular-nums text-muted-foreground">
              {delivery.attempts} of {MAX_ATTEMPTS} attempts
            </span>
          </div>

          <dl className="grid grid-cols-[max-content_1fr] gap-x-6 gap-y-2 text-sm">
            <dt className="text-muted-foreground">Connector</dt>
            <dd className="font-mono text-xs">{delivery.subscriber}</dd>
            <dt className="text-muted-foreground">Sent by</dt>
            <dd>{delivery.event.source}</dd>
            <dt className="text-muted-foreground">Created</dt>
            <dd className="tabular-nums">{absoluteTime(delivery.created_at)}</dd>
            {delivery.status === "PENDING" ? (
              <>
                <dt className="text-muted-foreground">Next try</dt>
                <dd className="tabular-nums">
                  {absoluteTime(delivery.next_run_at)} ({relativeTime(delivery.next_run_at)})
                </dd>
              </>
            ) : (
              <>
                <dt className="text-muted-foreground">Finished</dt>
                <dd className="tabular-nums">{absoluteTime(delivery.finished_at)}</dd>
              </>
            )}
            <dt className="text-muted-foreground">Delivery id</dt>
            <dd className="break-all font-mono text-xs">{delivery.id}</dd>
            <dt className="text-muted-foreground">Event id</dt>
            <dd className="break-all font-mono text-xs">{delivery.event_id}</dd>
            <dt className="text-muted-foreground">Trace</dt>
            <dd>
              {delivery.event.trace_id ? (
                <ViewTraceLink traceId={delivery.event.trace_id} variant="text" />
              ) : (
                "-"
              )}
            </dd>
          </dl>

          {delivery.last_error ? (
            <div className="rounded-md bg-[color:var(--tone-error-bg)] p-3 text-sm text-[color:var(--tone-error-fg)]">
              <p className="font-semibold">Last error</p>
              <p className="mt-1 break-words font-mono text-xs">{delivery.last_error}</p>
            </div>
          ) : null}

          {/* Above the history, so it stays in view when the list Retry button opens the drawer. */}
          {canRetry && delivery.status === "FAILED" ? (
            <RetryBox
              key={delivery.id}
              confirmOnOpen={confirmRetryOnOpen}
              isRetrying={isRetrying}
              onRetry={onRetry}
            />
          ) : null}

          <section aria-labelledby="delivery-history-heading" className="space-y-3">
            <h3 id="delivery-history-heading" className="text-sm font-semibold">
              Attempt history
            </h3>
            <AttemptHistory delivery={delivery} />
          </section>

          <details className="rounded-md border border-border">
            <summary className="cursor-pointer px-3 py-2 text-sm font-semibold">Payload</summary>
            <pre className="overflow-x-auto border-t border-border bg-muted/20 p-3 text-xs">
              {JSON.stringify(delivery.event.payload, null, 2)}
            </pre>
          </details>
        </div>
      )}
    </SideDrawer>
  );
}
