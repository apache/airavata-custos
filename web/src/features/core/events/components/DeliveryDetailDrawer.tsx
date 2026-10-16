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

import { ViewTraceLink } from "@/features/core/audit/components/ViewTraceLink";
import { JsonBlock } from "@/features/core/audit/components/primitives/JsonBlock";
import { Timestamp } from "@/features/core/audit/components/primitives/Timestamp";
import type { DeliveryHistory } from "@/generated/core/types.gen";
import { QueryErrorState } from "@/shared/ui/ErrorState";
import { CenteredSpinner } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { Button } from "@/shared/ui/button";
import { confirmToast } from "@/shared/ui/sonner";
import { CircleCheckIcon, CircleXIcon, RotateCcwIcon } from "lucide-react";
import * as React from "react";
import {
  type HistoryStep,
  MAX_ATTEMPTS,
  deliveryState,
  historyRuns,
  leftThisRun,
  relativeTime,
} from "../utils";
import { DeliveryStatusBadge } from "./DeliveryStatusBadge";

export type DeliveryDetailDrawerProps = {
  deliveryId: string | undefined;
  onOpenChange: (open: boolean) => void;
  delivery: DeliveryHistory | undefined;
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

function StepRow({ step }: { step: HistoryStep }) {
  if (step.kind === "retry") {
    return (
      <li className="flex gap-2 text-sm">
        <RotateCcwIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden />
        <div>
          <p>Retried by an admin</p>
          <Timestamp iso={step.at} className="text-xs text-muted-foreground" />
        </div>
      </li>
    );
  }
  if (step.kind === "other") {
    return (
      <li className="text-sm text-muted-foreground">
        {step.label} · <Timestamp iso={step.at} />
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
          Try {step.attempt} {step.ok ? "succeeded" : "failed"}
        </p>
        <Timestamp iso={step.at} className="text-xs text-muted-foreground" />
        {step.error ? (
          <p className="mt-1 break-words font-mono text-xs text-[color:var(--tone-error-fg)]">
            {step.error}
          </p>
        ) : null}
      </div>
    </li>
  );
}

function AttemptHistory({ delivery }: { delivery: DeliveryHistory }) {
  const runs = historyRuns(delivery.history);
  if (runs.length === 0) {
    return <p className="text-sm text-muted-foreground">No tries yet.</p>;
  }
  // Newest run and try first; runs are append-only, so the run number is a stable key.
  return (
    <div className="space-y-4">
      {runs.toReversed().map((run, back) => {
        const i = runs.length - 1 - back;
        return (
          <section key={i} aria-label={`Run ${i + 1}`} className="space-y-2">
            {runs.length > 1 ? (
              <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                Run {i + 1}
              </h3>
            ) : null}
            <ol className="space-y-3 border-l border-border pl-3">
              {run.toReversed().map((step, j) => (
                <StepRow key={step.at + String(j)} step={step} />
              ))}
            </ol>
          </section>
        );
      })}
    </div>
  );
}

// The Retry box sits in the history and unmounts once the delivery leaves FAILED.
const tryHistory = () =>
  document.querySelector<HTMLElement>('section[aria-labelledby="delivery-history-heading"]');

function RetryBox({
  confirmOnOpen,
  isRetrying,
  onRetry,
}: {
  confirmOnOpen: boolean;
  isRetrying: boolean;
  onRetry: () => void;
}) {
  const confirm = React.useCallback(
    () =>
      confirmToast(
        `Retry this delivery? It starts a new run of ${MAX_ATTEMPTS} tries. The history below is kept, and the retry is recorded as yours.`,
        "Retry now",
        onRetry,
        tryHistory,
      ),
    [onRetry],
  );
  // Deferred past the commit that opens the drawer, whose opening withdraws standing confirmations.
  // biome-ignore lint/correctness/useExhaustiveDependencies: raised once, when the box mounts
  React.useEffect(() => {
    if (!confirmOnOpen) return;
    const raise = setTimeout(confirm);
    return () => clearTimeout(raise);
  }, []);
  return (
    <div className="flex justify-end">
      <Button variant="outline" onClick={confirm} pending={isRetrying}>
        {isRetrying ? "Retrying…" : "Retry"}
      </Button>
    </div>
  );
}

export function DeliveryDetailDrawer({
  deliveryId,
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
      open={deliveryId != null}
      recordKey={deliveryId}
      onOpenChange={onOpenChange}
      width="lg"
      title={delivery?.event?.event_type ?? "Delivery"}
      description={delivery ? `Handled by ${delivery.subscriber}` : undefined}
    >
      {isLoading ? (
        <CenteredSpinner label="Loading delivery" />
      ) : error ? (
        <QueryErrorState error={error} what="delivery" onRetry={onRefresh} as="h3" />
      ) : !delivery ? null : (
        <div className="space-y-5">
          <div className="flex flex-wrap items-center gap-3">
            <DeliveryStatusBadge state={deliveryState(delivery, runningSubscribers)} />
            {delivery.status === "PENDING" ? (
              <span className="text-sm tabular-nums text-muted-foreground">
                {leftThisRun(delivery.attempts)}
              </span>
            ) : null}
          </div>

          <dl className="grid grid-cols-[max-content_1fr] gap-x-6 gap-y-2 text-sm">
            <dt className="text-muted-foreground">Connector</dt>
            <dd className="font-mono text-xs">{delivery.subscriber}</dd>
            <dt className="text-muted-foreground">Sent by</dt>
            <dd>{delivery.event.source}</dd>
            <dt className="text-muted-foreground">Created</dt>
            <dd className="tabular-nums">
              <Timestamp iso={delivery.created_at} />
            </dd>
            {delivery.status === "PENDING" ? (
              <>
                <dt className="text-muted-foreground">Next try</dt>
                <dd className="tabular-nums">
                  <Timestamp iso={delivery.next_run_at} /> ({relativeTime(delivery.next_run_at)})
                </dd>
              </>
            ) : (
              <>
                <dt className="text-muted-foreground">Finished</dt>
                <dd className="tabular-nums">
                  {delivery.finished_at ? <Timestamp iso={delivery.finished_at} /> : "—"}
                </dd>
              </>
            )}
            <dt className="text-muted-foreground">Delivery ID</dt>
            <dd className="break-all font-mono text-xs">{delivery.id}</dd>
            <dt className="text-muted-foreground">Event ID</dt>
            <dd className="break-all font-mono text-xs">{delivery.event_id}</dd>
            <dt className="text-muted-foreground">Trace</dt>
            <dd>
              {delivery.event.trace_id ? (
                <ViewTraceLink traceId={delivery.event.trace_id} variant="text" />
              ) : (
                "—"
              )}
            </dd>
          </dl>

          {delivery.last_error && delivery.status !== "SUCCEEDED" ? (
            <div className="rounded-md bg-[color:var(--tone-error-bg)] p-3 text-sm text-[color:var(--tone-error-fg)]">
              <p className="font-semibold">Last error</p>
              <p className="mt-1 break-words font-mono text-xs">{delivery.last_error}</p>
            </div>
          ) : null}

          <section aria-labelledby="delivery-history-heading" tabIndex={-1} className="space-y-3">
            <h3 id="delivery-history-heading" className="text-sm font-semibold">
              Try history
            </h3>
            {canRetry && delivery.status === "FAILED" ? (
              <RetryBox
                key={delivery.id}
                confirmOnOpen={confirmRetryOnOpen}
                isRetrying={isRetrying}
                onRetry={onRetry}
              />
            ) : null}
            <AttemptHistory delivery={delivery} />
          </section>

          <details className="rounded-md border border-border">
            <summary className="cursor-pointer px-3 py-2 text-sm font-semibold">Payload</summary>
            <JsonBlock
              text={JSON.stringify(delivery.event.payload, null, 2)}
              label="Payload"
              className="rounded-none border-x-0 border-b-0"
            />
          </details>
        </div>
      )}
    </SideDrawer>
  );
}
