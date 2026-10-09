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

import type { TraceEvent, TraceSummary } from "@/generated/core/types.gen";
import { cn } from "@/lib/utils";
import { ArrowRight } from "lucide-react";
import type * as React from "react";
import { formatAbsoluteUtc, formatRelative, isCodeShaped } from "../utils";
import { CopyValue } from "./primitives/CopyValue";
import { SourcePill } from "./primitives/SourcePill";
import { StatusPill } from "./primitives/StatusPill";

export type TraceSpanDetailPanelProps = {
  span: TraceEvent | null;
  trace: TraceSummary;
  source: string;
  onOpenInRaw: () => void;
};

const SECTION_LABEL_CLASS =
  "mb-1.5 text-[11px] font-bold uppercase tracking-[0.05em] text-muted-foreground";

function FactRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-baseline gap-3 py-1.5">
      <span className="w-[120px] shrink-0 text-[12px] font-medium text-muted-foreground">
        {label}
      </span>
      <span className="min-w-0 flex-1 text-[13px] text-foreground break-words">{children}</span>
    </div>
  );
}

export function TraceSpanDetailPanel({
  span,
  trace,
  source,
  onOpenInRaw,
}: TraceSpanDetailPanelProps) {
  if (!span) {
    return (
      <div
        data-testid="trace-span-detail-empty"
        className="flex h-full shrink-0 items-center justify-center rounded-[10px] border border-[color:var(--border)] bg-[color:var(--card)] p-6 text-center text-sm text-muted-foreground"
        style={{ width: 360 }}
      >
        Select a row to see span details.
      </div>
    );
  }

  const { event_type, created_at, status } = span;
  const code = isCodeShaped(event_type);
  const attrs = Object.entries({ "entity.type": span.entity_type, "entity.id": span.entity_id });

  return (
    <div
      data-testid="trace-span-detail"
      className="flex shrink-0 flex-col overflow-auto rounded-[10px] border border-[color:var(--border)] bg-[color:var(--card)] p-4"
      style={{ width: 360 }}
    >
      <div className="flex items-start justify-between gap-2">
        <div
          className={cn(
            "min-w-0 break-words font-bold text-foreground leading-snug",
            code ? "font-mono text-sm" : "text-[15.5px]",
          )}
        >
          {event_type}
        </div>
      </div>
      <button
        type="button"
        onClick={onOpenInRaw}
        data-testid="trace-span-open-raw"
        className="inline-flex items-center gap-1 self-start border-none bg-transparent py-1 text-[12px] font-semibold text-foreground hover:underline"
      >
        Open in Raw tab <ArrowRight className="h-3 w-3" aria-hidden="true" />
      </button>

      <div className="mt-2 mb-3 flex items-center gap-2">
        <StatusPill tone={status} />
        <SourcePill source={source} />
      </div>

      <div className="border-t border-[color:var(--border)] pt-1.5">
        <FactRow label="Time">
          {formatAbsoluteUtc(created_at)}
          <span className="ml-1.5 text-muted-foreground">· {formatRelative(created_at)}</span>
        </FactRow>
        {span.description ? (
          <FactRow label="Description">
            <span
              className={cn(
                "font-mono text-[12px]",
                status === "error" && "font-semibold text-[color:var(--banner-error-fg)]",
              )}
            >
              {span.description}
            </span>
          </FactRow>
        ) : null}
      </div>

      {attrs.some(([, v]) => v) ? (
        <div className="mt-3.5">
          <div className={SECTION_LABEL_CLASS}>Attributes</div>
          <div className="overflow-hidden rounded-md border border-[color:var(--border)]">
            {attrs.map(([k, value = ""], i) => {
              return (
                <div
                  key={k}
                  className={cn(
                    "flex items-baseline gap-2.5 px-2.5 py-1.5",
                    i % 2 === 1 ? "bg-[color:var(--muted-2)]" : "bg-[color:var(--card)]",
                  )}
                >
                  <span className="w-[46%] shrink-0 break-all font-mono text-[11.5px] text-muted-foreground">
                    {k}
                  </span>
                  <span className="min-w-0 flex-1">
                    <CopyValue value={value} label={k} explicit />
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      ) : null}

      <div className="mt-3.5 flex flex-col gap-2 border-t border-[color:var(--border)] pt-2.5">
        <IdRow label="Trace ID" value={trace.trace_id} />
        <IdRow label="Span ID" value={span.span_id} />
        <IdRow label="Parent" value={span.parent_span_id} />
      </div>
    </div>
  );
}

function IdRow({ label, value }: { label: string; value?: string }) {
  return (
    <div className="flex items-center gap-3">
      <span className="w-[64px] shrink-0 text-[12px] font-medium text-muted-foreground">
        {label}
      </span>
      <span className="min-w-0 flex-1 truncate font-mono text-xs">
        {value ? <CopyValue value={value} label={label} explicit /> : "—"}
      </span>
    </div>
  );
}
