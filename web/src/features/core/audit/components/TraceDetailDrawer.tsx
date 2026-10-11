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

import { cn } from "@/lib/utils";
import { ApiError } from "@/shared/api/client";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TabsRouter, type TabsRouterTab } from "@/shared/ui/TabsRouter";
import { Button, buttonVariants } from "@/shared/ui/button";
import { Skeleton } from "@/shared/ui/skeleton";
import { Dialog as DialogPrimitive } from "@base-ui/react/dialog";
import { RefreshCw, X } from "lucide-react";
import Link from "next/link";
import * as React from "react";
import { useTrace } from "../queries";
import type { TraceDetail } from "../types";
import {
  countRows,
  entityLinks,
  formatAbsoluteUtc,
  formatRelative,
  isCodeShaped,
  originTitle,
  rootStep,
  shortHex,
} from "../utils";
import { TraceFlowTab } from "./TraceFlowTab";
import { TraceRawTab } from "./TraceRawTab";
import { CopyValue } from "./primitives/CopyValue";
import { SourcePill } from "./primitives/SourcePill";
import { StatusPill } from "./primitives/StatusPill";

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
  const notFound = error instanceof ApiError && error.status === 404;

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
            width: "min(1080px, calc(100vw - 48px))",
            boxShadow: "var(--shadow-drawer)",
            transitionDuration: reducedMotion ? "0ms" : "220ms",
          }}
        >
          <DialogPrimitive.Title className="sr-only">
            {data ? `Trace ${shortHex(data.trace_id, 12)}` : "Trace detail"}
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
                {notFound ? (
                  <ErrorState
                    heading="No trace with this id"
                    message="It may be older than what we keep."
                    onRetry={onClose}
                    retryLabel="Back to traces"
                  />
                ) : (
                  <ErrorState
                    message={(error as Error).message ?? "Could not load trace."}
                    onRetry={() => refetch()}
                    retryLabel="Retry"
                  />
                )}
              </div>
            </div>
          ) : data ? (
            <DrawerContent
              detail={data}
              onClose={onClose}
              onRefresh={() => refetch()}
              refreshing={isFetching}
            />
          ) : null}
        </DialogPrimitive.Popup>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}

type DrawerContentProps = {
  detail: TraceDetail;
  onClose: () => void;
  onRefresh: () => void;
  refreshing: boolean;
};

function DrawerContent({ detail, onClose, onRefresh, refreshing }: DrawerContentProps) {
  const root = rootStep(detail);
  const title = root ? originTitle(root) : "Trace";
  const stepCount = React.useMemo(() => countRows(detail.tree), [detail.tree]);
  const entities = React.useMemo(() => entityLinks(detail), [detail]);
  const related = root?.entity_id;

  const tabs: TabsRouterTab[] = React.useMemo(
    () => [
      { value: "flow", label: "Flow", content: <TraceFlowTab detail={detail} /> },
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
    [detail, stepCount],
  );

  return (
    <div className="flex h-full flex-col">
      <div className="flex-shrink-0 px-6 pt-[18px]">
        <div className="flex items-start gap-3">
          <div className="min-w-0 flex-1">
            <div className="mb-1 flex flex-wrap items-center gap-2.5">
              <span className="text-[11.5px] font-bold uppercase tracking-[0.04em] text-muted-foreground">
                TRACE
              </span>
              <span className="min-w-0 truncate font-mono text-xs text-foreground">
                <CopyValue value={detail.trace_id} label="trace ID" explicit />
              </span>
              <StatusPill
                status={
                  detail.status === "error"
                    ? "failed"
                    : detail.status === "in_progress"
                      ? "retrying"
                      : "done"
                }
                label={detail.status === "in_progress" ? "In progress" : undefined}
              />
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
              {root && (
                <span title={formatAbsoluteUtc(root.created_at)}>
                  Started {formatRelative(root.created_at)}
                </span>
              )}
            </div>
          </div>
          <div className="flex shrink-0 items-center gap-2">
            {related && (
              <Link
                href={`/admin/traces?q=${encodeURIComponent(related)}&window=30d`}
                className={buttonVariants({ variant: "outline", size: "sm" })}
              >
                Related traces
              </Link>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={onRefresh}
              aria-label="Refresh trace"
              disabled={refreshing}
            >
              <RefreshCw className={cn("h-3.5 w-3.5", refreshing && "animate-spin")} />
              <span>Refresh</span>
            </Button>
            <CloseButton onClose={onClose} />
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
