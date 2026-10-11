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

import { ApiError, apiFetch } from "@/shared/api/client";
import {
  auditErrorEnvelopeSchema,
  type TraceNodeWire,
  traceDetailWireSchema,
  traceEventListWireSchema,
  traceListWireSchema,
  traceSourceListWireSchema,
} from "./schemas";
import type {
  AuditEventsResponse,
  RetryErrorEnvelope,
  Span,
  Trace,
  TraceDetailResponse,
  TraceListResponse,
} from "./types";

export type TraceListFilters = {
  source?: string[];
  status?: number[];
  from?: string;
  to?: string;
  q?: string;
  limit?: number;
  offset?: number;
};

export class RetryApiError extends ApiError {
  constructor(
    status: number,
    path: string,
    rawBody: unknown,
    public readonly body: RetryErrorEnvelope,
  ) {
    super(status, path, rawBody, body.error);
    this.name = "RetryApiError";
  }
}

// UI uses numeric status codes; wire is strings.
const STATUS_STR_TO_NUM: Record<string, number> = {
  ok: 0,
  error: 1,
  in_progress: 2,
};

const STATUS_NUM_TO_STR: Record<number, string> = {
  0: "ok",
  1: "error",
  2: "in_progress",
};

function uiStatus(wire: string): number {
  return STATUS_STR_TO_NUM[wire] ?? 2;
}

// Go zero-time marshals as "0001-01-01T00:00:00Z"; treat as not-ended so
// the in-progress tone fires correctly.
function endedAt(raw: string, status: string): string | null {
  if (status === "in_progress") return null;
  if (!raw || raw.startsWith("0001-01-01")) return null;
  return raw;
}

function qs(filters: TraceListFilters): string {
  const search = new URLSearchParams();
  if (filters.source?.length) {
    for (const s of [...filters.source].sort()) search.append("source", s);
  }
  if (filters.status?.length) {
    for (const code of [...filters.status].sort((a, b) => a - b)) {
      const wire = STATUS_NUM_TO_STR[code];
      if (wire) search.append("status", wire);
    }
  }
  if (filters.from) search.set("from", filters.from);
  if (filters.to) search.set("to", filters.to);
  if (filters.q) search.set("q", filters.q);
  if (typeof filters.limit === "number") search.set("limit", String(filters.limit));
  if (typeof filters.offset === "number") search.set("offset", String(filters.offset));
  const str = search.toString();
  return str ? `?${str}` : "";
}

export async function listTraces(filters: TraceListFilters = {}): Promise<TraceListResponse> {
  const raw = await apiFetch(`/audit/traces${qs(filters)}`);
  const wire = traceListWireSchema.parse(raw);
  const traces: Trace[] = wire.traces.map((t) => ({
    trace_id: t.trace_id,
    root_name: t.root_operation,
    source: t.source,
    status: uiStatus(t.status),
    started_at: t.started_at,
    ended_at: endedAt(t.ended_at, t.status),
    span_count: t.event_count,
    root_event: null,
  }));
  return { traces, total: wire.total, limit: wire.limit, offset: wire.offset };
}

// Backend ships a recursive tree; the UI's buildTree expects a flat array
// joined by parent_span_id. Walk the tree once, emit one span per node.
function flattenTree(nodes: TraceNodeWire[], rootTraceStart: string): Span[] {
  const out: Span[] = [];
  const walk = (n: TraceNodeWire) => {
    out.push({
      span_id: n.span_id,
      parent_span_id: n.parent_span_id,
      name: n.event_type,
      kind: out.length === 0 && n.parent_span_id == null ? 1 : 0,
      status: uiStatus(n.status),
      status_message: n.description,
      start_time: n.created_at || rootTraceStart,
      end_time: n.status === "in_progress" ? null : n.created_at || rootTraceStart,
      attributes: n.entity_id || n.entity_type
        ? { ...(n.entity_id ? { "entity.id": n.entity_id } : {}), ...(n.entity_type ? { "entity.type": n.entity_type } : {}) }
        : null,
    });
    for (const c of n.children) walk(c);
  };
  for (const r of nodes) walk(r);
  return out;
}

export async function getTrace(traceId: string): Promise<TraceDetailResponse> {
  const raw = await apiFetch(`/audit/traces/${encodeURIComponent(traceId)}`);
  const wire = traceDetailWireSchema.parse(raw);
  const root = wire.tree[0];
  const spans = flattenTree(wire.tree, root?.created_at ?? "");
  const inProgress = root?.status === "in_progress";
  const trace: Trace = {
    trace_id: wire.trace_id,
    root_name: root?.event_type ?? "",
    source: root?.source ?? "core",
    status: uiStatus(root?.status ?? "in_progress"),
    started_at: root?.created_at ?? "",
    ended_at: inProgress ? null : (spans[spans.length - 1]?.end_time ?? root?.created_at ?? null),
    span_count: spans.length,
    root_event: root?.description ? { description: root.description } : null,
  };
  return { trace, spans };
}

export async function listAuditSources(): Promise<string[]> {
  const raw = await apiFetch("/audit/sources");
  return traceSourceListWireSchema.parse(raw).sources;
}

// 202 Accepted on success. Backend doesn't expose retry on this branch; kept
// for cross-feature shape parity so the drawer's tooltip can remain unchanged.
export async function retryTrace(traceId: string): Promise<void> {
  try {
    await apiFetch(`/audit/traces/${encodeURIComponent(traceId)}/retry`, { method: "POST" });
  } catch (err) {
    if (err instanceof ApiError) {
      const parsed = auditErrorEnvelopeSchema.safeParse(err.body);
      if (parsed.success) {
        throw new RetryApiError(err.status, err.path, err.body, parsed.data);
      }
    }
    throw err;
  }
}

export async function getAuditEventsForTrace(
  traceId: string,
  spanId?: string,
): Promise<AuditEventsResponse> {
  const search = new URLSearchParams();
  search.set("trace_id", traceId);
  if (spanId) search.set("span_id", spanId);
  const raw = await apiFetch(`/audit/events?${search.toString()}`);
  const wire = traceEventListWireSchema.parse(raw);
  return {
    audit_events: wire.events.map((e, i) => ({
      id: `${traceId}::${e.span_id}::${i}`,
      event_type: e.event_type,
      event_time: e.created_at,
      entity_id: e.entity_id ?? "",
      details: e.description ?? "",
      trace_id: traceId,
      span_id: e.span_id,
    })),
    amie_audit_log: [],
  };
}
