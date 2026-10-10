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

import type { TraceSummary } from "@/generated/core/types.gen";
import { cn } from "@/lib/utils";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { Skeleton } from "@/shared/ui/skeleton";
import { ArrowRight } from "lucide-react";
import { formatAbsoluteUtc, formatRelative, isCodeShaped, shortHex, traceTone } from "../utils";
import { CopyValue } from "./primitives/CopyValue";
import { SourcePill } from "./primitives/SourcePill";
import { StatusPill } from "./primitives/StatusPill";

export type TraceTableProps = {
  traces: TraceSummary[];
  total: number;
  page: number;
  pageSize: number;
  loading?: boolean;
  error?: Error | null;
  hasActiveFilters?: boolean;
  onView(traceId: string): void;
  onPageChange(next: number): void;
  onPageSizeChange(next: number): void;
  onRetry?: () => void;
};

const GRID_COLS = "120px 132px 1fr 96px 64px";
const HEADERS = ["Started", "Trace ID", "Root action", "Source", "Spans"];
const PAGE_SIZES = [25, 50, 100];
const SKELETON_KEYS = ["s1", "s2", "s3", "s4", "s5"];

function TraceRow({ trace, onView }: { trace: TraceSummary; onView(id: string): void }) {
  const { trace_id, root_operation, source, started_at, event_count } = trace;
  const tone = traceTone(trace);
  const isErr = tone === "error";
  const isRunning = tone === "in-progress";
  const actionMono = isCodeShaped(root_operation);

  return (
    // biome-ignore lint/a11y/useKeyWithClickEvents: keyboard nav routed through the inner <button>; row mouse handler is a pointer convenience
    <div
      data-testid={`trace-row-${trace_id}`}
      data-tone={tone}
      onClick={(e) => {
        if ((e.target as HTMLElement).closest("button")) return;
        onView(trace_id);
      }}
      className={cn(
        "group relative grid cursor-pointer items-center border-b border-[color:var(--border)]",
        "bg-card hover:bg-[color:var(--muted-2)] focus-within:bg-[color:var(--muted-2)]",
      )}
      style={{ gridTemplateColumns: GRID_COLS, padding: "12px 16px" }}
    >
      {isErr && (
        <span
          aria-hidden="true"
          data-testid="error-rail"
          className="absolute left-0 top-0 bottom-0 w-[3px] bg-[color:var(--custos-red-500)]"
        />
      )}

      <button
        type="button"
        onClick={() => onView(trace_id)}
        aria-label={`Open trace ${shortHex(trace_id, 8)}`}
        className="sr-only focus:not-sr-only focus:absolute focus:inset-0 focus:rounded-sm focus:outline-none focus:ring-2 focus:ring-ring"
      />

      <span
        className="text-[13px] text-foreground tabular-nums"
        title={formatAbsoluteUtc(started_at)}
      >
        {formatRelative(started_at)}
      </span>

      <span className="font-mono text-[13px] text-foreground">
        <CopyValue value={trace_id} label="trace ID">
          <span>{shortHex(trace_id, 8)}…</span>
        </CopyValue>
      </span>

      <div className="min-w-0 pr-3">
        <div className="flex min-w-0 items-center gap-2">
          <StatusPill tone={tone} dotOnly />
          <span
            className={cn("truncate text-[13px] text-foreground", actionMono ? "font-mono" : "")}
            title={root_operation}
          >
            {root_operation}
          </span>
        </div>
        {isRunning && (
          <div className="ml-4 mt-0.5 truncate text-[12.5px] italic text-[color:var(--tone-warn-fg)]">
            …still running
          </div>
        )}
      </div>

      <span>
        <SourcePill source={source} />
      </span>

      <div className="flex items-center justify-between gap-2">
        <span className="font-mono text-[13px] tabular-nums text-muted-foreground">
          {event_count}
        </span>
        <span
          aria-hidden="true"
          className={cn(
            "inline-flex items-center gap-0.5 text-[12.5px] font-semibold text-muted-foreground",
            "opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100",
            "group-hover:text-[color:var(--brand)] group-focus-within:text-[color:var(--brand)]",
          )}
        >
          open <ArrowRight className="h-3 w-3" />
        </span>
      </div>
    </div>
  );
}

function TableHeader() {
  return (
    <div
      className="grid border-b border-[color:var(--border)] bg-[color:var(--muted-2)]"
      style={{ gridTemplateColumns: GRID_COLS, padding: "10px 16px" }}
    >
      {HEADERS.map((h) => (
        <span
          key={h}
          className="text-[11.5px] font-semibold uppercase tracking-[0.04em] text-muted-foreground"
        >
          {h}
        </span>
      ))}
    </div>
  );
}

function SkeletonRow() {
  return (
    <div
      className="grid items-center border-b border-[color:var(--border)]"
      style={{ gridTemplateColumns: GRID_COLS, padding: "12px 16px" }}
    >
      <Skeleton className="h-3.5 w-16" />
      <Skeleton className="h-3.5 w-20" />
      <Skeleton className="h-3.5 w-3/5" />
      <Skeleton className="h-4 w-14" />
      <Skeleton className="h-3.5 w-6" />
    </div>
  );
}

export function TraceTable({
  traces,
  total,
  page,
  pageSize,
  loading,
  error,
  hasActiveFilters,
  onView,
  onPageChange,
  onPageSizeChange,
  onRetry,
}: TraceTableProps) {
  // total counts traces before the status filter, so a page can hold fewer
  // rows than pageSize while later pages still have matches.
  const prevDisabled = page <= 1;
  const nextDisabled = page * pageSize >= total;

  return (
    <div className="flex flex-col gap-3" data-testid="trace-table-region">
      <Card className="overflow-hidden rounded-xl shadow-sm">
        <TableHeader />
        {loading ? (
          <div data-testid="trace-table-loading">
            {SKELETON_KEYS.map((k) => (
              <SkeletonRow key={k} />
            ))}
          </div>
        ) : error ? (
          <div className="p-4">
            <ErrorState
              message={error.message ?? "Could not load traces."}
              onRetry={onRetry}
              retryLabel="Retry"
            />
          </div>
        ) : traces.length === 0 ? (
          <div
            className="px-4 py-12 text-center text-sm text-muted-foreground"
            data-testid="trace-table-empty"
          >
            {hasActiveFilters
              ? "No traces match these filters."
              : "No traces yet. Once activity starts, flows will appear here."}
          </div>
        ) : (
          traces.map((t) => <TraceRow key={t.trace_id} trace={t} onView={onView} />)
        )}
      </Card>

      <div className="flex items-center justify-between gap-3 text-[13px] text-muted-foreground">
        <Button
          size="sm"
          variant="outline"
          disabled={prevDisabled}
          onClick={() => onPageChange(page - 1)}
          aria-label="Previous page"
        >
          ‹ Prev
        </Button>
        <div className="flex items-center gap-4">
          <span>
            Page <strong className="text-foreground">{page}</strong> ·{" "}
            <strong className="text-foreground">{traces.length}</strong>{" "}
            {traces.length === 1 ? "trace" : "traces"} shown
          </span>
          <label className="inline-flex items-center gap-2">
            <span className="text-xs">Rows</span>
            <select
              value={pageSize}
              onChange={(e) => onPageSizeChange(Number.parseInt(e.target.value, 10))}
              aria-label="Rows per page"
              className="h-8 rounded-md border border-[color:var(--border-strong)] bg-card px-2 text-[13px] text-foreground"
            >
              {PAGE_SIZES.map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </label>
        </div>
        <Button
          size="sm"
          variant="outline"
          disabled={nextDisabled}
          onClick={() => onPageChange(page + 1)}
          aria-label="Next page"
        >
          Next ›
        </Button>
      </div>
    </div>
  );
}
