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
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { LastSyncedBadge } from "@/shared/ui/LastSyncedBadge";
import { AlertTriangle, ArrowRight } from "lucide-react";
import { useRouter } from "next/navigation";
import * as React from "react";
import { useAuditSources, useTraces } from "../queries";
import { listStatus } from "../utils";
import { TraceDetailDrawer } from "./TraceDetailDrawer";
import { TraceFilterStrip } from "./TraceFilterStrip";
import { TraceTable } from "./TraceTable";
import {
  DEFAULT_FILTERS,
  type ListFilters,
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
const DRAWER_PARAMS = ["trace", "step", "span", "tab"];
const WINDOW_LABEL = { "24h": "24 hours", "7d": "7 days", "30d": "30 days" } as const;

export function TraceListContainer({ initialTraceId }: TraceListContainerProps = {}) {
  const params = useShallowSearchParams();
  const router = useRouter();
  const filters = React.useMemo(() => parseFilters(params), [params]);
  const traceParam = params.get(TRACE_PARAM);
  const activeTraceId = traceParam ?? initialTraceId ?? null;
  const drawerOpen = traceParam !== null || initialTraceId != null;

  const updateFilters = React.useCallback(
    (next: ListFilters) => {
      const search = serializeFilters(next);
      for (const key of DRAWER_PARAMS) {
        const value = params.get(key);
        if (value) search.set(key, value);
      }
      replaceShallowSearchParams(search);
    },
    [params],
  );

  // One `now` per mount keeps the from/to window, and so the query key, stable
  // between renders.
  const nowRef = React.useRef<number>(Date.now());
  const { from, to } = React.useMemo(
    () => windowToFromTo(filters.window, nowRef.current),
    [filters.window],
  );
  const { apiStatus, keep } = React.useMemo(
    () => statusFiltersToApi(filters.status),
    [filters.status],
  );

  const apiFilters = React.useMemo(
    () => ({
      status: apiStatus.length ? apiStatus : undefined,
      source: filters.source.length ? filters.source : undefined,
      from,
      to,
      q: filters.q || undefined,
      limit: filters.pageSize,
      offset: (filters.page - 1) * filters.pageSize,
    }),
    [apiStatus, filters.source, filters.q, filters.page, filters.pageSize, from, to],
  );

  const tracesQuery = useTraces(apiFilters);
  const sourcesQuery = useAuditSources();
  const visibleTraces = React.useMemo(
    () => (tracesQuery.data?.traces ?? []).filter((t) => keep(listStatus(t))),
    [tracesQuery.data, keep],
  );

  // The banner counts failed traces in the window whatever the filters say.
  const failingQuery = useTraces({ status: ["error"], from, to, limit: 1 });
  const failingCount = failingQuery.data?.total ?? 0;
  const showBanner = failingCount > 0 && !filters.status.includes("failed");

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
    if (initialTraceId) {
      router.push(`/admin/traces${next.toString() ? `?${next.toString()}` : ""}`);
      return;
    }
    replaceShallowSearchParams(next);
  }, [filters, initialTraceId, router]);

  const dataUpdatedAt = tracesQuery.dataUpdatedAt;
  const syncedAt = dataUpdatedAt ? new Date(dataUpdatedAt) : new Date(nowRef.current);
  const refresh = () => {
    void tracesQuery.refetch();
    void failingQuery.refetch();
  };

  return (
    <div className="w-full pb-12 pt-2">
      <header className="mb-2 flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight tracking-[-0.01em] text-foreground">
            Tracing
          </h1>
          <p className="mt-1.5 max-w-[560px] text-sm text-muted-foreground">
            Follow a request from where it started through the event bus to each connector.
          </p>
        </div>
        <LastSyncedBadge syncedAt={syncedAt} onRefetch={refresh} />
      </header>

      {showBanner && (
        // biome-ignore lint/a11y/useSemanticElements: role="status" promotes the banner to a polite live region; section is the landmark.
        <section
          role="status"
          aria-label="Failed traces"
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
            <strong>{failingCount}</strong> {failingCount === 1 ? "trace needs" : "traces need"}{" "}
            attention
          </span>
          <button
            type="button"
            onClick={() => updateFilters({ ...filters, status: ["failed"], page: 1 })}
            className="ml-auto inline-flex items-center gap-1 text-[13px] font-semibold text-[color:var(--banner-error-fg)] hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            Show them <ArrowRight className="h-3.5 w-3.5" />
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
          total={filters.status.length ? visibleTraces.length : (tracesQuery.data?.total ?? 0)}
          page={filters.page}
          pageSize={filters.pageSize}
          loading={tracesQuery.isLoading}
          error={tracesQuery.error as Error | null}
          hasActiveFilters={hasActiveFilters(filters)}
          windowLabel={WINDOW_LABEL[filters.window]}
          onView={onView}
          onPageChange={(next) => updateFilters({ ...filters, page: Math.max(1, next) })}
          onPageSizeChange={(next) => updateFilters({ ...filters, pageSize: next, page: 1 })}
          onClearFilters={() => updateFilters({ ...DEFAULT_FILTERS, pageSize: filters.pageSize })}
          onRetry={() => tracesQuery.refetch()}
        />
      </div>

      <TraceDetailDrawer traceId={activeTraceId} open={drawerOpen} onClose={closeDrawer} />
    </div>
  );
}
