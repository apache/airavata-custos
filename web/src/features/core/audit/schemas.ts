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

// TODO(openapi): adopt zTrace* from generated/core/zod.gen.

import { z } from "zod";

const traceIdHexSchema = z
  .string()
  .regex(/^[0-9a-f]{32}$/, "trace_id must be 32-char lowercase hex");

const spanIdHexSchema = z
  .string()
  .regex(/^[0-9a-f]{16}$/, "span_id must be 16-char lowercase hex");

// Widen with `.or(string)` so a future backend status doesn't crash list parsing.
export const wireStatusSchema = z.enum(["ok", "error", "in_progress"]).or(z.string());

export const traceSourceSchema = z.enum(["amie", "comanage", "slurm", "http", "core"]).or(z.string());

export const traceSummaryWireSchema = z.object({
  trace_id: traceIdHexSchema,
  root_operation: z.string(),
  source: traceSourceSchema,
  status: wireStatusSchema,
  started_at: z.string(),
  ended_at: z.string(),
  event_count: z.number().int().nonnegative(),
});

export const traceListWireSchema = z.object({
  traces: z.array(traceSummaryWireSchema),
  total: z.number().int().nonnegative(),
  limit: z.number().int().nonnegative(),
  offset: z.number().int().nonnegative(),
});

// parent_span_id is .optional() not .nullable() because Go omitempty never emits null.
const traceEventBaseSchema = z.object({
  span_id: spanIdHexSchema,
  parent_span_id: spanIdHexSchema.optional(),
  source: traceSourceSchema,
  event_type: z.string(),
  entity_type: z.string().optional(),
  entity_id: z.string().optional(),
  description: z.string().optional(),
  status: wireStatusSchema,
  created_at: z.string(),
});

export type TraceNodeWire = z.infer<typeof traceEventBaseSchema> & {
  children: TraceNodeWire[];
};
export const traceNodeWireSchema: z.ZodType<TraceNodeWire> = traceEventBaseSchema.extend({
  get children() {
    return z.array(traceNodeWireSchema);
  },
});

export const traceDetailWireSchema = z.object({
  trace_id: traceIdHexSchema,
  tree: z.array(traceNodeWireSchema),
  truncated: z.boolean(),
});

export const traceEventListWireSchema = z.object({
  events: z.array(traceEventBaseSchema),
});

export const traceSourceListWireSchema = z.object({
  sources: z.array(z.string()),
});

export const auditErrorEnvelopeSchema = z.object({
  error: z.string(),
});

// ---------- UI shape (adapted in api.ts; no wire parsing here) ----------

// Numeric status codes preserve tone-precedence (0=ok, 1=error,
// 2=no-status, 3=orphaned). Trace-level in-progress comes from null ended_at.
const uiStatusCodeSchema = z.number().int().nonnegative();
const uiSpanKindSchema = z.number().int().nonnegative();
const nullableJsonSchema = z.unknown().nullish();

export const traceSchema = z.object({
  trace_id: traceIdHexSchema,
  root_name: z.string(),
  source: traceSourceSchema,
  status: uiStatusCodeSchema,
  started_at: z.string(),
  ended_at: z.string().nullish(),
  span_count: z.number().int().nonnegative(),
  root_event: nullableJsonSchema,
});

export const spanSchema = z.object({
  span_id: spanIdHexSchema,
  parent_span_id: spanIdHexSchema.optional(),
  name: z.string(),
  kind: uiSpanKindSchema,
  status: uiStatusCodeSchema,
  status_message: z.string().nullish(),
  start_time: z.string(),
  end_time: z.string().nullish(),
  attributes: nullableJsonSchema,
});

export const traceListResponseSchema = z.object({
  traces: z.array(traceSchema),
  total: z.number().int().nonnegative(),
  limit: z.number().int().nonnegative(),
  offset: z.number().int().nonnegative(),
});

export const traceDetailResponseSchema = z.object({
  trace: traceSchema,
  spans: z.array(spanSchema),
});

// Linked-entities tab consumes the merged shape. Backend has no AMIE-extras
// table on this branch so amie_audit_log stays empty until the connector wires it.
export const coreAuditEventSchema = z.object({
  id: z.string(),
  event_type: z.string(),
  event_time: z.string(),
  entity_id: z.string(),
  details: z.string(),
  trace_id: traceIdHexSchema,
  span_id: spanIdHexSchema,
});

export const amieAuditLogSchema = z.object({
  id: z.number().int(),
  packet_id: z.string(),
  event_id: z.string().nullish(),
  action: z.string(),
  entity_type: z.string(),
  entity_id: z.string().nullish(),
  summary: z.string().nullish(),
  created_at: z.string(),
  trace_id: traceIdHexSchema,
  span_id: spanIdHexSchema,
});

export const auditEventsResponseSchema = z.object({
  audit_events: z.array(coreAuditEventSchema),
  amie_audit_log: z.array(amieAuditLogSchema),
});
