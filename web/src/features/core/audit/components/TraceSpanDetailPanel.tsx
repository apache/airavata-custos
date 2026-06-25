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

import { ArrowRight } from "lucide-react";
import type * as React from "react";
import { cn } from "@/lib/utils";
import type { Trace, UISpan } from "../types";
import {
  durationBetween,
  formatAbsoluteUtc,
  formatDurationMs,
  formatRelative,
  isCodeShaped,
  rowTone,
} from "../utils";
import { CopyValue } from "./primitives/CopyValue";
import { SourcePill } from "./primitives/SourcePill";
import { StatusPill } from "./primitives/StatusPill";

export type TraceSpanDetailPanelProps = {
  span: UISpan | null;
  trace: Trace;
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

export function TraceSpanDetailPanel({ span, trace, source, onOpenInRaw }: TraceSpanDetailPanelProps) {
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

  const tone = rowTone(span);
  const code = isCodeShaped(span.name);
  const dur = durationBetween(span.start_time, span.end_time ?? null);
  const attrs =
    span.attributes && typeof span.attributes === "object" && !Array.isArray(span.attributes)
      ? (span.attributes as Record<string, unknown>)
      : null;
  const attrKeys = attrs ? Object.keys(attrs).sort() : [];

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
          {span.name}
        </div>
      </div>
      <button
        type="button"
        onClick={onOpenInRaw}
        data-testid="trace-span-open-raw"
        className="inline-flex items-center gap-1 self-start border-none bg-transparent py-1 text-[12px] font-semibold text-brand hover:underline"
      >
        Open in Raw tab <ArrowRight className="h-3 w-3" aria-hidden="true" />
      </button>

      <div className="mt-2 mb-3 flex items-center gap-2">
        <StatusPill tone={tone} />
        <SourcePill source={source} />
        <span className="font-mono text-[11.5px] text-muted-foreground">kind={span.kind}</span>
      </div>

      <div className="border-t border-[color:var(--border)] pt-1.5">
        <FactRow label="Time">
          {formatAbsoluteUtc(span.start_time)}
          <span className="ml-1.5 text-muted-foreground">· {formatRelative(span.start_time)}</span>
        </FactRow>
        <FactRow label="Duration">
          {dur == null ? "—" : formatDurationMs(dur)}
        </FactRow>
        {span.status_message ? (
          <FactRow label="Status message">
            <span className="font-mono text-[12px] font-semibold text-[color:var(--banner-error-fg)]">
              {span.status_message}
            </span>
          </FactRow>
        ) : null}
      </div>

      {attrs && typeof attrs.summary === "string" && attrs.summary ? (
        <div className="mt-3">
          <div className={SECTION_LABEL_CLASS}>Summary</div>
          <div className="text-[13px] leading-relaxed text-foreground">{attrs.summary}</div>
        </div>
      ) : null}

      {attrKeys.length > 0 ? (
        <div className="mt-3.5">
          <div className={SECTION_LABEL_CLASS}>Attributes</div>
          <div className="overflow-hidden rounded-md border border-[color:var(--border)]">
            {attrKeys.map((k, i) => {
              const raw = attrs?.[k];
              const value =
                raw == null
                  ? ""
                  : typeof raw === "string"
                    ? raw
                    : typeof raw === "number" || typeof raw === "boolean"
                      ? String(raw)
                      : JSON.stringify(raw);
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
        {span.parent_span_id ? <IdRow label="Parent" value={span.parent_span_id} /> : null}
      </div>
    </div>
  );
}

function IdRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center gap-3">
      <span className="w-[64px] shrink-0 text-[12px] font-medium text-muted-foreground">
        {label}
      </span>
      <span className="min-w-0 flex-1 truncate font-mono text-xs">
        <CopyValue value={value} label={label} explicit />
      </span>
    </div>
  );
}
