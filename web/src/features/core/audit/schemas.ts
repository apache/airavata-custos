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

// Wire shapes of the audit trace endpoints. The components render these as
// they are; there is no adapted UI shape.

import { z } from "zod";

const traceIdHexSchema = z
  .string()
  .regex(/^[0-9a-f]{32}$/, "trace_id must be 32-char lowercase hex");

const spanIdHexSchema = z.string().regex(/^[0-9a-f]{16}$/, "span_id must be 16-char lowercase hex");

export const traceStatusSchema = z.enum(["ok", "error", "in_progress"]);
export const stepStatusSchema = z.enum(["ok", "error"]);
export const deliveryStatusSchema = z.enum(["PENDING", "SUCCEEDED", "FAILED"]);

export const deliveryCountsSchema = z.object({
  pending: z.number().int().nonnegative(),
  succeeded: z.number().int().nonnegative(),
  failed: z.number().int().nonnegative(),
  attempts: z.number().int().nonnegative(),
});

export const traceSummarySchema = z.object({
  trace_id: traceIdHexSchema,
  root_operation: z.string(),
  source: z.string(),
  status: traceStatusSchema,
  started_at: z.string(),
  ended_at: z.string(),
  event_count: z.number().int().nonnegative(),
  deliveries: deliveryCountsSchema,
});

export const traceListSchema = z.object({
  traces: z.array(traceSummarySchema),
  total: z.number().int().nonnegative(),
  limit: z.number().int().nonnegative(),
  offset: z.number().int().nonnegative(),
});

// parent_span_id is .optional() not .nullable() because Go omitempty never emits null.
export const stepSchema = z.object({
  id: z.string(),
  span_id: spanIdHexSchema,
  parent_span_id: spanIdHexSchema.optional(),
  source: z.string(),
  event_type: z.string(),
  entity_type: z.string().optional(),
  entity_id: z.string().optional(),
  description: z.string().optional(),
  status: stepStatusSchema,
  created_at: z.string(),
});

export type StepNode = z.infer<typeof stepSchema> & { children: StepNode[] };
export const stepNodeSchema: z.ZodType<StepNode> = stepSchema.extend({
  get children() {
    return z.array(stepNodeSchema);
  },
});

export const traceDeliverySchema = z.object({
  id: z.string(),
  event_type: z.string(),
  subscriber: z.string(),
  status: deliveryStatusSchema,
  attempts: z.number().int().nonnegative(),
  next_run_at: z.string(),
  last_error: z.string().optional(),
  span_id: spanIdHexSchema,
});

export const traceDetailSchema = z.object({
  trace_id: traceIdHexSchema,
  status: traceStatusSchema,
  tree: z.array(stepNodeSchema),
  deliveries: z.array(traceDeliverySchema),
  truncated: z.boolean(),
});

export const traceSourceListSchema = z.object({
  sources: z.array(z.string()),
});
