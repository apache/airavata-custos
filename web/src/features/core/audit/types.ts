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
