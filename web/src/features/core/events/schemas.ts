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

import { z } from "zod";

export const deliveryStatusSchema = z.enum(["PENDING", "SUCCEEDED", "FAILED"]);

// The trace id column is fixed width, so a missing trace comes back as spaces.
const traceIdSchema = z.string().transform((s) => s.trim());

export const busEventSchema = z.object({
  id: z.string(),
  event_type: z.string(),
  payload: z.unknown(),
  source: z.string(),
  trace_id: traceIdSchema,
  created_at: z.string(),
});

export const deliverySchema = z.object({
  id: z.string(),
  event_id: z.string(),
  subscriber: z.string(),
  status: deliveryStatusSchema,
  attempts: z.number().int().nonnegative(),
  next_run_at: z.string(),
  last_error: z.string().optional(),
  created_at: z.string(),
  finished_at: z.string().optional(),
  event: busEventSchema,
});

export const deliveryListSchema = z.object({ items: z.array(deliverySchema) });

export const historyEntrySchema = z.object({
  id: z.string(),
  event_type: z.string(),
  event_time: z.string(),
  details: z.string(),
  source: z.string(),
  trace_id: traceIdSchema,
});

export const userNameSchema = z.object({
  first_name: z.string().optional(),
  last_name: z.string().optional(),
  email: z.string().optional(),
});

export const deliveryDetailSchema = deliverySchema.extend({
  history: z.array(historyEntrySchema),
});

export const subscriptionListSchema = z.object({
  items: z.array(
    z.object({
      subscriber: z.string(),
      event_type: z.string(),
      loaded: z.boolean(),
    }),
  ),
});
