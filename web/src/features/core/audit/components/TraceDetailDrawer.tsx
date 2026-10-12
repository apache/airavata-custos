/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

"use client";

import { relativeTime } from "@/features/core/events/utils";
import { cn } from "@/lib/utils";
import { useCanonicalParams } from "@/shared/hooks/useShallowSearchParams";
import { QueryErrorState } from "@/shared/ui/ErrorState";
import { TabsRouter, type TabsRouterTab } from "@/shared/ui/TabsRouter";
import { Button, buttonVariants } from "@/shared/ui/button";
import { useFocusOrigin, wrapTab } from "@/shared/ui/focus";
import { Link } from "@/shared/ui/link";
import { Skeleton } from "@/shared/ui/skeleton";
import { Dialog as DialogPrimitive } from "@base-ui/react/dialog";
import { RefreshCw, X } from "lucide-react";
import * as React from "react";
import { useTrace } from "../queries";
import type { TraceDetail } from "../types";
import {
  buildFlow,
  countRows,
  detailStatus,
  entityLinks,
  isCodeShaped,
  originTitle,
  rootStep,
  shortHex,
} from "../utils";
import { DrawerLeaveContext, RelatedTracesLink } from "./RelatedTracesLink";
import { TraceFlowTab, useFlowParams } from "./TraceFlowTab";
import { TraceRawTab } from "./TraceRawTab";
import { CopyValue } from "./primitives/CopyValue";
import { SourcePill } from "./primitives/SourcePill";
import { StatusPill } from "./primitives/StatusPill";
import { Timestamp } from "./primitives/Timestamp";

export type TraceDetailDrawerProps = {
  traceId: string | null;
  open: boolean;
  onClose: () => void;
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
      <Skeleton className="h-4 w-96" />
    </div>
  );
}

function BodySkeleton() {
  return (
    <div className="space-y-3 p-6">
      <Skeleton className="h-9 w-full" />
      <Skeleton className="h-24 w-full" />
      <Skeleton className="h-9 w-full" />
      <Skeleton className="h-24 w-full" />
    </div>
  );
}

export function TraceDetailDrawer({ traceId, open, onClose }: TraceDetailDrawerProps) {
  const reducedMotion = useReducedMotion();
  const { data, isLoading, error, refetch, isFetching } = useTrace(
    open && traceId ? traceId : undefined,
  );
  const returnTo = useFocusOrigin(open, traceId ?? undefined);
  const leaving = React.useRef(false);
  const leave = React.useCallback(() => {
    leaving.current = true;
  }, []);

  return (
    // Non-modal, so a press on another trace row swaps the drawer; Tab wraps within it.
    <DialogPrimitive.Root
      open={open}
      modal={false}
      onOpenChange={(next, details) => {
        if (next) return;
        const target = details.event?.target;
        if (
          details.reason === "outside-press" &&
          target instanceof Element &&
          target.closest("[data-drawer-trigger]")
        )
          return details.cancel();
        onClose();
      }}
    >
      <DialogPrimitive.Portal>
        <DialogPrimitive.Popup
          data-testid="trace-detail-drawer"
          finalFocus={() => {
            const left = leaving.current;
            leaving.current = false;
            return left ? false : returnTo();
          }}
          onKeyDown={wrapTab}
          className={cn(
            "fixed inset-y-0 right-0 z-50 flex flex-col bg-[color:var(--card)] outline-none",
            reducedMotion
              ? null
              : "data-open:animate-in data-open:slide-in-from-right-12 data-closed:animate-out data-closed:slide-out-to-right-12",
          )}
          style={{
            width: "min(1080px, calc(100vw - 48px))",
            boxShadow: "var(--shadow-drawer)",
            transitionDuration: reducedMotion ? "0ms" : "220ms",
          }}
        >
          <DialogPrimitive.Title className="sr-only">
            {data ? `Trace ${shortHex(data.trace_id, 12)}` : "Trace detail"}
          </DialogPrimitive.Title>

          {isLoading ? (
            <output aria-label="Loading" className="block">
              <div className="px-6 pt-[18px]">
                <HeaderSkeleton />
              </div>
              <BodySkeleton />
            </output>
          ) : error ? (
            <DrawerError>
              <div className="flex items-center justify-end px-6 pt-[18px]">
                <CloseButton />
              </div>
              <div className="flex-1 px-6 py-10">
                <QueryErrorState error={error} what="trace" onRetry={() => refetch()} />
              </div>
            </DrawerError>
          ) : data ? (
            <DrawerLeaveContext.Provider value={leave}>
              <DrawerContent detail={data} onRefresh={() => refetch()} refreshing={isFetching} />
            </DrawerLeaveContext.Provider>
          ) : null}
        </DialogPrimitive.Popup>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}

// A trace that failed to load renders no tabs and no steps, so their params are dropped.
function DrawerError({ children }: { children: React.ReactNode }) {
  useCanonicalParams({ tab: undefined, step: undefined, span: undefined });
  return <div className="flex h-full flex-col">{children}</div>;
}

type DrawerContentProps = {
  detail: TraceDetail;
  onRefresh: () => void;
  refreshing: boolean;
};

function DrawerContent({ detail, onRefresh, refreshing }: DrawerContentProps) {
  const root = rootStep(detail);
  const title = root ? originTitle(root) : "Trace";
  const stepCount = React.useMemo(() => countRows(detail.tree), [detail.tree]);
  const model = React.useMemo(() => buildFlow(detail), [detail]);
  const { stepParam, spanParam } = useFlowParams(model);
  const entities = React.useMemo(() => entityLinks(detail), [detail]);
  const related = root?.entity_id;
  const started = detail.tree[0]?.created_at;

  const tabs: TabsRouterTab[] = React.useMemo(
    () => [
      {
        value: "flow",
        label: "Flow",
        content: (
          <TraceFlowTab detail={detail} model={model} stepParam={stepParam} spanParam={spanParam} />
        ),
      },
      {
        value: "raw",
        label: "Raw",
        content: (
          <div className="h-full overflow-auto">
            <TraceRawTab detail={detail} stepCount={stepCount} />
          </div>
        ),
      },
    ],
    [detail, model, stepParam, spanParam, stepCount],
  );

  return (
    <div className="flex h-full flex-col">
      <div className="flex-shrink-0 px-6 pt-[18px]">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start">
          <div className="min-w-0 flex-1">
            <div className="mb-1 flex flex-wrap items-center gap-2.5">
              <span className="text-[11.5px] font-bold uppercase tracking-[0.04em] text-muted-foreground">
                TRACE
              </span>
              <span className="min-w-0 truncate font-mono text-xs text-foreground">
                <CopyValue value={detail.trace_id} label="trace ID" explicit />
              </span>
              <StatusPill status={detailStatus(detail)} live />
              {root && <SourcePill source={root.source} size="sm" />}
            </div>
            <h2
              className={cn(
                "text-[19px] font-bold text-foreground break-words",
                isCodeShaped(title) ? "font-mono text-[16px]" : "font-display",
              )}
            >
              {title}
            </h2>
            <div className="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-[12.5px] text-muted-foreground">
              {entities.slice(0, 3).map((e) => (
                <span key={`${e.kind}:${e.id}`} className="inline-flex items-center gap-1">
                  <span>{e.kind}</span>
                  {e.href ? (
                    <Link href={e.href} className="font-mono text-foreground hover:underline">
                      {shortHex(e.id, 8)}
                    </Link>
                  ) : (
                    <CopyValue value={e.id} label={e.kind}>
                      <span>{shortHex(e.id, 8)}</span>
                    </CopyValue>
                  )}
                </span>
              ))}
              {started && <Timestamp iso={started}>Started {relativeTime(started)}</Timestamp>}
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2 sm:shrink-0">
            {related && (
              <RelatedTracesLink
                entityId={related}
                entityType={root?.entity_type}
                className={buttonVariants({ variant: "outline", size: "sm" })}
              />
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={onRefresh}
              aria-label="Refresh trace"
              disabled={refreshing}
              focusableWhenDisabled
            >
              <RefreshCw className={cn("h-3.5 w-3.5", refreshing && "animate-spin")} />
              <span>Refresh</span>
            </Button>
            <CloseButton />
          </div>
        </div>
      </div>

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden px-6 pb-6">
        <TabsRouter
          tabs={tabs}
          defaultValue="flow"
          className="flex min-h-0 flex-1 flex-col"
          panelClassName="min-h-0 flex-1 overflow-hidden"
        />
      </div>
    </div>
  );
}

function CloseButton() {
  return (
    <DialogPrimitive.Close
      render={
        <Button
          variant="outline"
          size="icon-sm"
          aria-label="Close"
          autoFocus
          data-testid="drawer-close"
        />
      }
    >
      <X className="h-4 w-4" />
    </DialogPrimitive.Close>
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
