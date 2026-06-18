// UI shapes consumed by the ported tracing components. Wire shapes live in
// schemas.ts; api.ts adapts wire -> UI so component code stays unchanged.

import type { z } from "zod";
import type {
  auditEventsResponseSchema,
  auditErrorEnvelopeSchema,
  spanSchema,
  traceDetailResponseSchema,
  traceListResponseSchema,
  traceSchema,
} from "./schemas";

export type Trace = z.infer<typeof traceSchema>;
export type Span = z.infer<typeof spanSchema>;
export type TraceListResponse = z.infer<typeof traceListResponseSchema>;
export type TraceDetailResponse = z.infer<typeof traceDetailResponseSchema>;
export type AuditEventsResponse = z.infer<typeof auditEventsResponseSchema>;
export type RetryErrorEnvelope = z.infer<typeof auditErrorEnvelopeSchema>;

export type TraceSource = "amie" | "http" | "comanage" | "slurm" | "core" | (string & {});

export const TRACE_SOURCES: ReadonlyArray<TraceSource> = [
  "amie",
  "http",
  "comanage",
  "slurm",
  "core",
];

export type RowTone = "ok" | "error" | "in-progress" | "orphaned" | "no-status";

// Wire span augmented with derived run-state flags. `running` = unfinished;
// `notRun` = skipped because a parent failed; `orphan` = parent_span_id
// doesn't resolve in this trace's span set.
export type UISpan = Span & {
  running?: boolean;
  notRun?: boolean;
  orphan?: boolean;
};

export type EntityRef = {
  kind: string;
  primaryId: string;
  attrs: Record<string, string>;
};
