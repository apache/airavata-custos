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

import { AlertTriangle, ArrowRight } from "lucide-react";
import { useRouter } from "next/navigation";
import * as React from "react";
import { cn } from "@/lib/utils";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { LastSyncedBadge } from "@/shared/ui/LastSyncedBadge";
import { useAuditSources, useTraces } from "../queries";
import type { Trace } from "../types";
import { traceTone } from "../utils";
import { TraceDetailDrawer } from "./TraceDetailDrawer";
import { TraceFilterStrip } from "./TraceFilterStrip";
import { TraceTable } from "./TraceTable";
import {
  DEFAULT_FILTERS,
  type ListFilters,
  bannerBounds,
  hasActiveFilters,
  parseFilters,
  serializeFilters,
  statusFiltersToApi,
  windowToFromTo,
} from "./traceListUrlState";

export type TraceListContainerProps = {
  initialTraceId?: string;
};

const TRACE_PARAM = "trace";

function syncUrl(filters: ListFilters, traceId: string | null) {
  const next = serializeFilters(filters);
  if (traceId) next.set(TRACE_PARAM, traceId);
  replaceShallowSearchParams(next);
}

export function TraceListContainer({ initialTraceId }: TraceListContainerProps = {}) {
  const params = useShallowSearchParams();
  const router = useRouter();
  const filters = React.useMemo(() => parseFilters(params), [params]);
  const traceParam = params.get(TRACE_PARAM);
  const activeTraceId = traceParam ?? initialTraceId ?? null;
  const drawerOpen = traceParam !== null || initialTraceId != null;

  const updateFilters = React.useCallback(
    (next: ListFilters) => syncUrl(next, activeTraceId),
    [activeTraceId],
  );

  // Stable `now` per-mount keeps the from/to window from drifting between
  // re-renders (and changing the TanStack cache key).
  const nowRef = React.useRef<number>(Date.now());
  const failing24h = React.useMemo(() => bannerBounds(nowRef.current), []);
  const { from, to } = React.useMemo(
    () => windowToFromTo(filters.window, nowRef.current),
    [filters.window],
  );

  const { apiStatus, inProgressOnly } = React.useMemo(
    () => statusFiltersToApi(filters.status),
    [filters.status],
  );

  // failingOver24h overrides status/window and pins the 30d->24h window so the
  // banner click lands on exactly the rows the count came from.
  const apiFilters = React.useMemo(
    () =>
      filters.failingOver24h
        ? {
            status: [1] as number[],
            source: filters.source.length ? filters.source : undefined,
            from: failing24h.from,
            to: failing24h.to,
            q: filters.q || undefined,
            limit: filters.pageSize,
            offset: (filters.page - 1) * filters.pageSize,
          }
        : {
            status: apiStatus.length ? apiStatus : undefined,
            source: filters.source.length ? filters.source : undefined,
            from,
            to,
            q: filters.q || undefined,
            limit: filters.pageSize,
            offset: (filters.page - 1) * filters.pageSize,
          },
    [
      filters.failingOver24h,
      apiStatus,
      filters.source,
      filters.q,
      filters.page,
      filters.pageSize,
      from,
      to,
      failing24h,
    ],
  );

  const tracesQuery = useTraces(apiFilters);
  const sourcesQuery = useAuditSources();
  const visibleTraces: Trace[] = React.useMemo(() => {
    const rows = tracesQuery.data?.traces ?? [];
    if (inProgressOnly) return rows.filter((t) => t.ended_at == null);
    return rows;
  }, [tracesQuery.data, inProgressOnly]);

  const total = inProgressOnly
    ? visibleTraces.length
    : (tracesQuery.data?.total ?? 0);

  // 24h-failing banner — separate query so it survives any active filter.
  const failingQuery = useTraces({
    status: [1],
    from: failing24h.from,
    to: failing24h.to,
    limit: 1,
  });
  const failingCount = failingQuery.data?.total ?? 0;
  const showBanner =
    !failingQuery.isLoading && failingCount > 0 && !filters.failingOver24h;

  const onView = React.useCallback(
    (traceId: string) => {
      const next = serializeFilters(filters);
      next.set(TRACE_PARAM, traceId);
      replaceShallowSearchParams(next);
    },
    [filters],
  );

  const closeDrawer = React.useCallback(() => {
    const next = serializeFilters(filters);
    next.delete(TRACE_PARAM);
    next.delete("span");
    next.delete("tab");
    if (initialTraceId) {
      router.push(`/admin/traces${next.toString() ? `?${next.toString()}` : ""}`);
      return;
    }
    replaceShallowSearchParams(next);
  }, [filters, initialTraceId, router]);

  const onPageChange = React.useCallback(
    (next: number) => updateFilters({ ...filters, page: Math.max(1, next) }),
    [filters, updateFilters],
  );
  const onPageSizeChange = React.useCallback(
    (next: number) => updateFilters({ ...filters, pageSize: next, page: 1 }),
    [filters, updateFilters],
  );

  const applyFailingPreset = React.useCallback(() => {
    updateFilters({
      ...DEFAULT_FILTERS,
      failingOver24h: true,
      pageSize: filters.pageSize,
    });
  }, [filters.pageSize, updateFilters]);

  const dataUpdatedAt = tracesQuery.dataUpdatedAt;
  const syncedAt = dataUpdatedAt ? new Date(dataUpdatedAt) : new Date(nowRef.current);

  // Sanity check — confirm tone derivation stays loaded.
  void traceTone;

  return (
    <div className="w-full pb-12 pt-2">
      <header className="mb-2 flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight tracking-[-0.01em] text-foreground">
            Traces
          </h1>
          <p className="mt-1.5 max-w-[560px] text-sm text-muted-foreground">
            Investigate where a flow broke and retry from the failed step.
          </p>
        </div>
        <LastSyncedBadge syncedAt={syncedAt} onRefetch={() => tracesQuery.refetch()} />
      </header>

      {showBanner && (
        // biome-ignore lint/a11y/useSemanticElements: role="status" promotes the banner to a polite live region; section is the landmark.
        <section
          role="status"
          aria-label="Failing traces alert"
          data-testid="failing-banner"
          className={cn(
            "mt-4 flex items-center gap-3 rounded-[10px] border px-4 py-3",
            "border-[color:var(--banner-error-border)] bg-[color:var(--banner-error-bg)] text-[color:var(--banner-error-fg)]",
          )}
        >
          <AlertTriangle
            className="h-4 w-4 shrink-0 text-[color:var(--banner-error-icon)]"
            aria-hidden="true"
          />
          <span className="text-[13.5px]">
            <strong>{failingCount}</strong>{" "}
            {failingCount === 1 ? "trace" : "traces"} failing for over 24h
          </span>
          <button
            type="button"
            onClick={applyFailingPreset}
            className="ml-auto inline-flex items-center gap-1 text-[13px] font-semibold text-[color:var(--banner-error-fg)] hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            Investigate <ArrowRight className="h-3.5 w-3.5" />
          </button>
        </section>
      )}

      <div className="mt-4">
        <TraceFilterStrip
          value={filters}
          onChange={updateFilters}
          sourceOptions={sourcesQuery.data}
        />
      </div>

      <div className="mt-4">
        <TraceTable
          traces={visibleTraces}
          total={total}
          page={filters.page}
          pageSize={filters.pageSize}
          loading={tracesQuery.isLoading}
          error={tracesQuery.error as Error | null}
          hasActiveFilters={hasActiveFilters(filters)}
          onView={onView}
          onPageChange={onPageChange}
          onPageSizeChange={onPageSizeChange}
          onRetry={() => tracesQuery.refetch()}
        />
      </div>

      <TraceDetailDrawer traceId={activeTraceId} open={drawerOpen} onClose={closeDrawer} />
    </div>
  );
}
