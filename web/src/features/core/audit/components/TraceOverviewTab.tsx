"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import type { Span, Trace } from "../types";
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
  trace: Trace;
  spans: Span[];
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
  const tone = traceTone(trace);
  const actionMono = isCodeShaped(trace.root_name);
  const durationMs = durationBetween(trace.started_at, trace.ended_at ?? null);

  // ROOT span entity refs only — first kind=1 entry, falling back to first
  // span when no root marker is present.
  const rootEntityAttrs = React.useMemo(() => {
    const root = spans.find((s) => s.kind === 1) ?? spans[0];
    return root ? getEntityRefs([root]) : [];
  }, [spans]);

  const facts: Array<{ label: string; value: React.ReactNode }> = [
    {
      label: "Trace ID",
      value: (
        <span className="font-mono text-xs">
          <CopyValue value={trace.trace_id} label="trace ID" explicit />
        </span>
      ),
    },
    { label: "Source", value: <SourcePill source={String(trace.source)} /> },
    {
      label: "Root action",
      value: (
        <span className={actionMono ? "font-mono text-[12.5px]" : "text-[13.5px]"}>
          {trace.root_name}
        </span>
      ),
    },
    { label: "Status", value: <StatusPill tone={tone} /> },
    {
      label: "Started",
      value: (
        <span>
          {formatAbsoluteUtc(trace.started_at)}{" "}
          <span className="text-muted-foreground">· {formatRelative(trace.started_at)}</span>
        </span>
      ),
    },
    {
      label: "Ended",
      value: trace.ended_at ? (
        <span>
          {formatAbsoluteUtc(trace.ended_at)}{" "}
          <span className="text-muted-foreground">· {formatRelative(trace.ended_at)}</span>
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
      value: <span className="tabular-nums">{trace.span_count}</span>,
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
      {rootEntityAttrs.length === 0 ? (
        <div className="mb-6 rounded-[10px] border border-[color:var(--border)] bg-[color:var(--card)] p-4 text-[13px] text-muted-foreground">
          No root entity attributes captured.
        </div>
      ) : (
        <div className="mb-6 flex flex-wrap gap-2.5">
          {rootEntityAttrs.map((ref) => {
            const entry = Object.entries(ref.attrs)[0];
            if (!entry) return null;
            const [key, value] = entry;
            return (
              <div
                key={`${ref.kind}::${ref.primaryId}::${key}`}
                className="rounded-lg border border-[color:var(--border)] bg-[color:var(--card)] px-3 py-2"
              >
                <div className="mb-1 font-mono text-[11px] text-muted-foreground">{key}</div>
                <CopyValue value={value} label={key} />
              </div>
            );
          })}
        </div>
      )}

      <div className={SECTION_LABEL_CLASS}>ATTEMPTS</div>
      <div className="rounded-[10px] border border-dashed border-[color:var(--border-strong)] px-4 py-3.5 text-[13px] text-muted-foreground">
        No retry attempts yet. When retry ships, each attempt appears here as a linked
        sub-trace.
      </div>
    </div>
  );
}
