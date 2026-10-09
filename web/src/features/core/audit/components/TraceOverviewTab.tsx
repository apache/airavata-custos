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
import type * as React from "react";
import {
  durationBetween,
  formatAbsoluteUtc,
  formatDurationMs,
  formatRelative,
  getEntityRefs,
  isCodeShaped,
  traceTone,
} from "../utils";
import { CopyValue } from "./primitives/CopyValue";
import { SourcePill } from "./primitives/SourcePill";
import { StatusPill } from "./primitives/StatusPill";

export type TraceOverviewTabProps = {
  trace: TraceSummary;
  spans: TraceEvent[];
};

const SECTION_LABEL_CLASS =
  "mb-2 text-[11.5px] font-bold uppercase tracking-[0.04em] text-muted-foreground";

function FactRow({
  label,
  value,
  even,
}: {
  label: string;
  value: React.ReactNode;
  even: boolean;
}) {
  return (
    <div
      className={cn(
        "flex items-center gap-4 px-4 py-2.5",
        even ? "bg-[color:var(--muted-2)]" : "bg-[color:var(--card)]",
      )}
    >
      <span className="w-[150px] shrink-0 text-[12.5px] font-medium text-muted-foreground">
        {label}
      </span>
      <span className="min-w-0 flex-1 text-[13.5px] text-foreground">{value}</span>
    </div>
  );
}

export function TraceOverviewTab({ trace, spans }: TraceOverviewTabProps) {
  const { trace_id, root_operation, source, started_at } = trace;
  const tone = traceTone(trace);
  const endedAt = tone === "in-progress" ? undefined : trace.ended_at;
  const actionMono = isCodeShaped(root_operation);
  const durationMs = durationBetween(started_at, endedAt);
  const rootEntities = getEntityRefs(spans.slice(0, 1));

  const facts: Array<{ label: string; value: React.ReactNode }> = [
    {
      label: "Trace ID",
      value: (
        <span className="font-mono text-xs">
          <CopyValue value={trace_id} label="trace ID" explicit />
        </span>
      ),
    },
    { label: "Source", value: <SourcePill source={source} /> },
    {
      label: "Root action",
      value: (
        <span className={actionMono ? "font-mono text-[12.5px]" : "text-[13.5px]"}>
          {root_operation}
        </span>
      ),
    },
    { label: "Status", value: <StatusPill tone={tone} /> },
    {
      label: "Started",
      value: (
        <span>
          {formatAbsoluteUtc(started_at)}{" "}
          <span className="text-muted-foreground">· {formatRelative(started_at)}</span>
        </span>
      ),
    },
    {
      label: "Ended",
      value: endedAt ? (
        <span>
          {formatAbsoluteUtc(endedAt)}{" "}
          <span className="text-muted-foreground">· {formatRelative(endedAt)}</span>
        </span>
      ) : (
        <span className="italic text-[color:var(--custos-amber-700)]">still running</span>
      ),
    },
    {
      label: "Duration",
      value: durationMs == null ? "—" : formatDurationMs(durationMs),
    },
    {
      label: "Span count",
      value: <span className="tabular-nums">{trace.event_count}</span>,
    },
  ];

  return (
    <div className="max-w-[760px]">
      <div className={SECTION_LABEL_CLASS}>TRACE FACTS</div>
      <div className="mb-6 overflow-hidden rounded-[10px] border border-[color:var(--border)]">
        {facts.map((row, i) => (
          <FactRow key={row.label} label={row.label} value={row.value} even={i % 2 === 1} />
        ))}
      </div>

      <div className={SECTION_LABEL_CLASS}>ROOT ENTITY</div>
      {rootEntities.length === 0 ? (
        <div className="mb-6 rounded-[10px] border border-[color:var(--border)] bg-[color:var(--card)] p-4 text-[13px] text-muted-foreground">
          No root entity attributes captured.
        </div>
      ) : (
        <div className="mb-6 flex flex-wrap gap-2.5">
          {rootEntities.map(({ entity_type = "", entity_id = "" }) => (
            <div
              key={`${entity_type}::${entity_id}`}
              className="rounded-lg border border-[color:var(--border)] bg-[color:var(--card)] px-3 py-2"
            >
              <div className="mb-1 font-mono text-[11px] text-muted-foreground">{entity_type}</div>
              <CopyValue value={entity_id} label={entity_type} />
            </div>
          ))}
        </div>
      )}

      <div className={SECTION_LABEL_CLASS}>ATTEMPTS</div>
      <div className="rounded-[10px] border border-dashed border-[color:var(--border-strong)] px-4 py-3.5 text-[13px] text-muted-foreground">
        No retry attempts yet. When retry ships, each attempt appears here as a linked sub-trace.
      </div>
    </div>
  );
}
