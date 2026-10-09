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

import { Dialog as DialogPrimitive } from "@base-ui/react/dialog";
import { AlertTriangle, RefreshCw, X } from "lucide-react";
import * as React from "react";
import { useDelivery, useRetryDelivery } from "@/features/core/events/queries";
import type { TraceEvent, TraceSummary } from "@/generated/core/types.gen";
import { cn } from "@/lib/utils";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { setSearchParam } from "@/shared/hooks/useShallowSearchParams";
import { Button } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Skeleton } from "@/shared/ui/skeleton";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { TabsRouter, type TabsRouterTab } from "@/shared/ui/TabsRouter";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/shared/ui/tooltip";
import { useTrace } from "../queries";
import { detectErrorPath, isCodeShaped, shortHex, traceTone } from "../utils";
import { CopyValue } from "./primitives/CopyValue";
import { StatusPill } from "./primitives/StatusPill";
import { TraceLinkedEntitiesTab } from "./TraceLinkedEntitiesTab";
import { TraceOverviewTab } from "./TraceOverviewTab";
import { TraceRawTab } from "./TraceRawTab";
import { TraceTreeTab } from "./TraceTreeTab";

export type TraceDetailDrawerTab = "overview" | "tree" | "raw" | "linked";

export type TraceDetailDrawerProps = {
  traceId: string | null;
  open: boolean;
  onClose: () => void;
  initialTab?: TraceDetailDrawerTab;
};


function HeaderSkeleton() {
  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <Skeleton className="h-3 w-12" />
        <Skeleton className="h-3 w-48" />
        <Skeleton className="h-5 w-16" />
      </div>
      <Skeleton className="h-6 w-72" />
    </div>
  );
}

function BodySkeleton() {
  return (
    <div className="space-y-3 p-6">
      <Skeleton className="h-4 w-1/3" />
      <Skeleton className="h-32 w-full" />
      <Skeleton className="h-4 w-1/4" />
      <Skeleton className="h-20 w-full" />
    </div>
  );
}

// The first error leaf: its description is the error summary, its action the
// "at span" suffix.
function failingSpanFor(spans: TraceEvent[]): TraceEvent | undefined {
  const [leafId] = detectErrorPath(spans).errorLeafIds;
  return spans.find((s) => s.span_id === leafId);
}

export function TraceDetailDrawer({
  traceId,
  open,
  onClose,
  initialTab = "tree",
}: TraceDetailDrawerProps) {
  const reducedMotion = useReducedMotion();
  const { data, isLoading, error, refetch } = useTrace(open && traceId ? traceId : undefined);

  const trace = data?.trace;

  return (
    <DialogPrimitive.Root open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Backdrop
          className={cn(
            "fixed inset-0 z-50",
            reducedMotion
              ? null
              : "data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0",
          )}
          style={{
            background: "var(--drawer-scrim)",
            transitionDuration: reducedMotion ? "0ms" : "180ms",
          }}
        />
        <DialogPrimitive.Popup
          data-testid="trace-detail-drawer"
          className={cn(
            "fixed inset-y-0 right-0 z-50 flex flex-col bg-[color:var(--card)] outline-none",
            reducedMotion
              ? null
              : "data-open:animate-in data-open:slide-in-from-right-12 data-closed:animate-out data-closed:slide-out-to-right-12",
          )}
          style={{
            width: "min(1080px, calc(100vw - 96px))",
            boxShadow: "var(--shadow-drawer)",
            transitionDuration: reducedMotion ? "0ms" : "220ms",
          }}
        >
          <DialogPrimitive.Title className="sr-only">
            {trace ? `Trace ${shortHex(trace.trace_id, 12)}` : "Trace detail"}
          </DialogPrimitive.Title>

          {isLoading ? (
            <>
              <div className="px-6 pt-[18px]">
                <HeaderSkeleton />
              </div>
              <BodySkeleton />
            </>
          ) : error ? (
            <div className="flex h-full flex-col">
              <div className="flex items-center justify-end px-6 pt-[18px]">
                <CloseButton onClose={onClose} />
              </div>
              <div className="flex-1 px-6 py-10">
                <ErrorState
                  message={error.message || "Could not load trace."}
                  onRetry={() => refetch()}
                  retryLabel="Retry"
                />
              </div>
            </div>
          ) : data?.trace ? (
            <DrawerContent
              trace={data.trace}
              spans={data.spans}
              truncated={data.truncated}
              onClose={onClose}
              onRefresh={() => refetch()}
              initialTab={initialTab}
            />
          ) : null}
        </DialogPrimitive.Popup>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}

type DrawerContentProps = {
  trace: TraceSummary;
  spans: TraceEvent[];
  truncated: boolean;
  onClose: () => void;
  onRefresh: () => void;
  initialTab: TraceDetailDrawerTab;
};

function DrawerContent({
  trace,
  spans,
  truncated,
  onClose,
  onRefresh,
  initialTab,
}: DrawerContentProps) {
  const tone = traceTone(trace);
  const isInProgress = tone === "in-progress";
  const isError = tone === "error";
  const failingSpan = React.useMemo(
    () => (isError ? failingSpanFor(spans) : undefined),
    [isError, spans],
  );
  const failingAction = failingSpan?.event_type || null;
  // Retrying a trace re-runs a delivery whose latest attempt failed; attempts are time-ordered siblings.
  const lastAttempts = new Map(
    spans.filter((s) => s.entity_type === "event_delivery").map((s) => [s.entity_id, s]),
  );
  const failedDelivery = [...lastAttempts.values()].find((s) => s.status === "error")?.entity_id;
  const canRetry = useAbility().can("write", "EventDelivery");
  // Between automatic attempts the delivery is not FAILED yet, and the backend refuses a manual retry.
  const deliveryFailed =
    useDelivery(canRetry ? failedDelivery : undefined).data?.status === "FAILED";
  const retry = useRetryDelivery();
  const retryBlocked = !failedDelivery
    ? "Nothing to retry: this trace has no failed event delivery."
    : !canRetry
      ? "You cannot retry event deliveries."
      : !deliveryFailed
        ? "The delivery is still retrying on its own."
        : null;
  const retryButton = (
    <Button
      variant="default"
      size="sm"
      aria-disabled={retryBlocked ? true : undefined}
      disabled={retry.isPending}
      className={retryBlocked ? "opacity-50" : undefined}
      onClick={() =>
        !retryBlocked &&
        failedDelivery &&
        confirmToast("Retry the failed delivery in this trace?", "Retry", () =>
          retry.mutate(
            { path: { id: failedDelivery } },
            toastOnSuccess("Retry started", onRefresh),
          ),
        )
      }
    >
      <RefreshCw className="h-3.5 w-3.5" />
      <span>Retry</span>
    </Button>
  );
  const errorSummary = failingSpan?.description || null;
  const actionMono = isCodeShaped(trace.root_operation);
  const showErrorLine = isError && (errorSummary || failingAction);

  const treeLabel = spans.length > 0 ? `Tree (${spans.length})` : "Tree";
  // The span panel's "Open in Raw tab" writes ?tab, which TabsRouter reads.
  const switchTab = React.useCallback(
    (value: "overview" | "raw" | "linked") =>
      setSearchParam(new URLSearchParams(window.location.search), "tab", value),
    [],
  );

  const tabs: TabsRouterTab[] = React.useMemo(
    () => [
      {
        value: "overview",
        label: "Overview",
        content: (
          <div className="h-full overflow-auto">
            <TraceOverviewTab trace={trace} spans={spans} />
          </div>
        ),
      },
      {
        value: "tree",
        label: treeLabel,
        content: <TraceTreeTab trace={trace} spans={spans} onSwitchToTab={switchTab} />,
      },
      {
        value: "raw",
        label: "Raw",
        content: (
          <div className="h-full overflow-auto">
            <TraceRawTab trace={trace} spans={spans} />
          </div>
        ),
      },
      {
        value: "linked",
        label: "Linked entities",
        content: (
          <div className="h-full overflow-auto">
            <TraceLinkedEntitiesTab trace={trace} spans={spans} />
          </div>
        ),
      },
    ],
    [trace, spans, treeLabel, switchTab],
  );

  return (
    <TooltipProvider>
      <div className="flex h-full flex-col">
        <div className="flex-shrink-0 px-6 pt-[18px]">
          <div className="flex items-start gap-3">
            <div className="min-w-0 flex-1">
              <div className="mb-1 flex items-center gap-2.5">
                <span className="text-[11.5px] font-bold uppercase tracking-[0.04em] text-muted-foreground">
                  TRACE
                </span>
                <span className="min-w-0 truncate font-mono text-xs text-foreground">
                  <CopyValue value={trace.trace_id} label="trace ID" explicit />
                </span>
                <StatusPill tone={tone} />
              </div>
              <div
                className={cn(
                  "text-[19px] font-bold text-foreground break-words",
                  actionMono ? "font-mono" : "font-display",
                )}
              >
                {trace.root_operation}
              </div>
              {showErrorLine && (
                <div className="mt-1.5 flex items-center gap-1.5 text-[13.5px] font-semibold text-[color:var(--banner-error-fg)]">
                  <AlertTriangle
                    className="h-3.5 w-3.5 shrink-0 text-[color:var(--banner-error-icon)]"
                    aria-hidden="true"
                  />
                  <span>
                    {errorSummary}
                    {failingAction ? (
                      <span className="font-medium">
                        {errorSummary ? " " : ""}at span{" "}
                        <span className="font-mono text-[12.5px]">{failingAction}</span>
                      </span>
                    ) : null}
                  </span>
                </div>
              )}
              {truncated && (
                <div
                  data-testid="trace-truncated"
                  className="mt-1.5 text-[12.5px] text-[color:var(--tone-warn-fg)]"
                >
                  Showing the first {spans.length} spans; this trace has more.
                </div>
              )}
            </div>
            <div className="flex shrink-0 items-center gap-2">
              {isInProgress && (
                <Tooltip>
                  <TooltipTrigger
                    render={
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={onRefresh}
                        aria-label="Refresh trace"
                      >
                        <RefreshCw className="h-3.5 w-3.5" />
                        <span>Refresh</span>
                      </Button>
                    }
                  />
                  <TooltipContent>Refresh — manual polling, no auto-refetch.</TooltipContent>
                </Tooltip>
              )}
              {retryBlocked ? (
                <Tooltip>
                  {/* aria-disabled keeps the button focusable so the tooltip stays reachable. */}
                  <TooltipTrigger render={retryButton} />
                  <TooltipContent className="max-w-[260px]">{retryBlocked}</TooltipContent>
                </Tooltip>
              ) : (
                retryButton
              )}
              <CloseButton onClose={onClose} />
            </div>
          </div>
        </div>

        <div className="flex min-h-0 flex-1 flex-col overflow-hidden px-6 pb-6">
          <TabsRouter
            tabs={tabs}
            defaultValue={initialTab}
            className="flex min-h-0 flex-1 flex-col"
            panelClassName="min-h-0 flex-1 overflow-hidden"
          />
        </div>
      </div>
    </TooltipProvider>
  );
}

function CloseButton({ onClose }: { onClose: () => void }) {
  return (
    <Button
      variant="outline"
      size="icon-sm"
      onClick={onClose}
      aria-label="Close"
      autoFocus
      data-testid="drawer-close"
    >
      <X className="h-4 w-4" />
    </Button>
  );
}

function useReducedMotion(): boolean {
  const [reduced, setReduced] = React.useState(false);
  React.useEffect(() => {
    if (typeof window === "undefined" || !window.matchMedia) return;
    const mql = window.matchMedia("(prefers-reduced-motion: reduce)");
    setReduced(mql.matches);
    const handler = (e: MediaQueryListEvent) => setReduced(e.matches);
    mql.addEventListener("change", handler);
    return () => mql.removeEventListener("change", handler);
  }, []);
  return reduced;
}
