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
import { pluralize } from "@/shared/format";
import { PastLastPage } from "@/shared/ui/DataTable";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { useScrollStop } from "@/shared/ui/focus";
import { Skeleton } from "@/shared/ui/skeleton";
import { ArrowRight } from "lucide-react";
import type * as React from "react";
import type { TraceSummary } from "../types";
import { isCodeShaped, listStatus, listStatusLabel, plainName, shortHex } from "../utils";
import { CopyValue } from "./primitives/CopyValue";
import { SourcePill } from "./primitives/SourcePill";
import { StatusPill } from "./primitives/StatusPill";
import { Timestamp } from "./primitives/Timestamp";

export type TraceTableProps = {
  traces: TraceSummary[];
  total: number;
  page: number;
  pageSize: number;
  loading?: boolean;
  error?: Error | null;
  hasActiveFilters?: boolean;
  windowLabel: string;
  onView(traceId: string): void;
  onPageChange(next: number): void;
  onPageSizeChange(next: number): void;
  onClearFilters(): void;
  onRetry?: () => void;
  regionRef?: React.RefObject<HTMLElement | null>;
};

const GRID_COLS = "210px minmax(220px, 1fr) 130px 112px 140px";
const HEADERS = ["Status", "Started by", "Deliveries", "Last activity", "Trace"];
const PAGE_SIZES = [25, 50, 100];
const SKELETON_KEYS = ["s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8"];

function deliveriesText(t: TraceSummary): { text: string; className: string } {
  const { pending, succeeded, failed } = t.deliveries;
  const total = pending + succeeded + failed;
  if (total === 0) return { text: "—", className: "text-muted-foreground" };
  const text = `${succeeded} of ${total} done`;
  if (failed > 0) return { text, className: "text-[color:var(--tone-error-fg)] font-semibold" };
  if (pending > 0) return { text, className: "text-[color:var(--tone-warn-fg)]" };
  return { text, className: "text-muted-foreground" };
}

function TraceRow({ trace, onView }: { trace: TraceSummary; onView(id: string): void }) {
  const status = listStatus(trace.status, trace.deliveries.attempts);
  const name = plainName(trace.root_operation) ?? trace.root_operation;
  const deliveries = deliveriesText(trace);

  return (
    // biome-ignore lint/a11y/useKeyWithClickEvents: keyboard nav routed through the inner <button>; row mouse handler is a pointer convenience
    <div
      data-testid={`trace-row-${trace.trace_id}`}
      data-row
      data-row-key={trace.trace_id}
      data-drawer-trigger
      data-status={status}
      onClick={(e) => {
        if ((e.target as HTMLElement).closest("button")) return;
        onView(trace.trace_id);
      }}
      className={cn(
        "group relative grid min-h-11 cursor-pointer items-center gap-x-3 border-b border-[color:var(--border)]",
        "bg-card hover:bg-[color:var(--muted-2)] focus-within:bg-[color:var(--muted-2)]",
      )}
      style={{ gridTemplateColumns: GRID_COLS, padding: "10px 16px" }}
    >
      {status === "failed" && (
        <span
          aria-hidden="true"
          data-testid="error-rail"
          className="absolute left-0 top-0 bottom-0 w-[3px] bg-[color:var(--fill-error)]"
        />
      )}

      <button
        type="button"
        onClick={() => onView(trace.trace_id)}
        aria-label={`Open trace ${shortHex(trace.trace_id, 8)}`}
        data-row-opener
        className="sr-only focus:not-sr-only focus:absolute focus:inset-0 focus:rounded-sm focus:outline-none focus:ring-2 focus:ring-ring"
      />

      <span>
        <StatusPill status={status} label={listStatusLabel(trace)} />
      </span>

      <div className="flex min-w-0 items-center gap-2">
        <span
          className={cn(
            "truncate text-[13px] text-foreground",
            isCodeShaped(name) ? "font-mono text-[12.5px]" : "",
          )}
          title={trace.root_operation}
        >
          {name}
        </span>
        <SourcePill source={trace.source} size="sm" />
      </div>

      <span className={cn("text-[13px] tabular-nums", deliveries.className)}>
        {deliveries.text}
      </span>

      <Timestamp iso={trace.ended_at} className="text-[13px] text-foreground tabular-nums">
        {relativeTime(trace.ended_at)}
      </Timestamp>

      <div className="flex items-center justify-between gap-2">
        <span className="font-mono text-[12.5px] text-muted-foreground">
          <CopyValue value={trace.trace_id} label="trace ID">
            <span>{shortHex(trace.trace_id, 8)}</span>
          </CopyValue>
        </span>
        <span
          aria-hidden="true"
          className={cn(
            "inline-flex items-center gap-0.5 text-[12.5px] font-semibold text-muted-foreground",
            "opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100",
            "group-hover:text-foreground group-focus-within:text-foreground",
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
      className="grid gap-x-3 border-b border-[color:var(--border)] bg-[color:var(--muted-2)]"
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
      className="grid items-center gap-x-3 border-b border-[color:var(--border)]"
      style={{ gridTemplateColumns: GRID_COLS, padding: "12px 16px" }}
    >
      <Skeleton className="h-4 w-20" />
      <Skeleton className="h-3.5 w-3/5" />
      <Skeleton className="h-3.5 w-16" />
      <Skeleton className="h-3.5 w-12" />
      <Skeleton className="h-3.5 w-16" />
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
  windowLabel,
  onView,
  onPageChange,
  onPageSizeChange,
  onClearFilters,
  onRetry,
  regionRef,
}: TraceTableProps) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const prevDisabled = page <= 1;
  const nextDisabled = page >= totalPages || total === 0;
  const scrollStop = useScrollStop<HTMLElement>();

  return (
    <div className="flex flex-col gap-3" data-testid="trace-table-region">
      <Card className="rounded-xl shadow-sm">
        <section
          ref={(el) => {
            scrollStop.ref(el);
            if (regionRef) regionRef.current = el;
          }}
          className="overflow-x-auto focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
          aria-label="Traces"
          data-table-region
          // Focused by the list after a filter change even when it is not a Tab stop.
          tabIndex={scrollStop.tabIndex ?? -1}
        >
          <div className="min-w-[892px]">
            <TableHeader />
            {loading ? (
              <output aria-label="Loading" className="block" data-testid="trace-table-loading">
                {SKELETON_KEYS.map((k) => (
                  <SkeletonRow key={k} />
                ))}
              </output>
            ) : error ? (
              <div className="p-4">
                <ErrorState error={error} onRetry={onRetry} retryLabel="Retry" />
              </div>
            ) : traces.length === 0 ? (
              <div
                className="flex flex-col items-center gap-3 px-4 py-12 text-center text-sm text-muted-foreground"
                data-testid="trace-table-empty"
              >
                {page > 1 ? (
                  <PastLastPage things="traces" page={page} onPageChange={onPageChange} />
                ) : hasActiveFilters ? (
                  <>
                    <span>No traces match these filters.</span>
                    <Button size="sm" variant="outline" onClick={onClearFilters}>
                      Clear filters
                    </Button>
                  </>
                ) : (
                  <span>No traces in the last {windowLabel}.</span>
                )}
              </div>
            ) : (
              traces.map((t) => <TraceRow key={t.trace_id} trace={t} onView={onView} />)
            )}
          </div>
        </section>
      </Card>

      {!loading && traces.length > 0 && (
        <div className="flex items-center justify-between gap-3 text-[13px] text-muted-foreground">
          <Button
            size="sm"
            variant="outline"
            disabled={prevDisabled}
            focusableWhenDisabled
            onClick={() => onPageChange(page - 1)}
            aria-label="Previous page"
          >
            Prev
          </Button>
          <div className="flex items-center gap-4">
            <span>
              Page <strong className="text-foreground">{page}</strong> of {totalPages},{" "}
              <strong className="text-foreground">{total}</strong> {pluralize("trace", total)}
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
            focusableWhenDisabled
            onClick={() => onPageChange(page + 1)}
            aria-label="Next page"
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}
