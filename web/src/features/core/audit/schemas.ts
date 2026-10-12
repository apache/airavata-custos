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

import {
  zDeliveryCounts,
  zEventDeliveryStatus,
  zGetAuditSourcesResponse,
  zGetAuditTracesByTraceIdResponse,
  zGetAuditTracesResponse,
  zTraceDelivery,
  zTraceEvent,
  zTraceSummary,
} from "@/generated/core/zod.gen";
import { z } from "zod";

const traceIdHexSchema = z
  .string()
  .regex(/^[0-9a-f]{32}$/, "trace_id must be 32-char lowercase hex");

const spanIdHexSchema = z.string().regex(/^[0-9a-f]{16}$/, "span_id must be 16-char lowercase hex");

export const traceStatusSchema = z.enum(["ok", "error", "in_progress"]);
export const stepStatusSchema = z.enum(["ok", "error"]);
export const deliveryStatusSchema = zEventDeliveryStatus;

export const deliveryCountsSchema = zDeliveryCounts;

export const traceSummarySchema = zTraceSummary.extend({
  trace_id: traceIdHexSchema,
  status: traceStatusSchema,
});

export const traceListSchema = zGetAuditTracesResponse.extend({
  traces: z.array(traceSummarySchema),
});

export const stepSchema = zTraceEvent.extend({
  span_id: spanIdHexSchema,
  parent_span_id: spanIdHexSchema.optional(),
  status: stepStatusSchema,
});

export type StepNode = z.infer<typeof stepSchema> & { children: StepNode[] };
export const stepNodeSchema: z.ZodType<StepNode> = stepSchema.extend({
  get children() {
    return z.array(stepNodeSchema);
  },
});

export const traceDeliverySchema = zTraceDelivery.extend({
  span_id: spanIdHexSchema,
});

export const traceDetailSchema = zGetAuditTracesByTraceIdResponse.extend({
  trace_id: traceIdHexSchema,
  status: traceStatusSchema,
  tree: z.array(stepNodeSchema),
  deliveries: z.array(traceDeliverySchema),
});

export const traceSourceListSchema = zGetAuditSourcesResponse;
