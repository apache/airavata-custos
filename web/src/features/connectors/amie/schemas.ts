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

// TODO(openapi): replace with generated from amie.openapi.yaml.

import { z } from "zod";

export const packetStatusSchema = z.enum(["NEW", "DECODED", "PROCESSED", "FAILED"]);

export const linkedEntityRefSchema = z.object({
  type: z.enum(["project", "account", "person", "user_merge"]),
  id: z.string().min(1),
  display_id: z.string().optional(),
});

export const packetSchema = z.object({
  id: z.string(),
  amie_id: z.string(),
  type: z.string(),
  status: packetStatusSchema,
  source: z.string(),
  raw_json: z.string().optional(),
  decoded_payload: z.record(z.string(), z.unknown()).optional(),
  received_at: z.string(),
  updated_at: z.string(),
  decoded_at: z.string().optional(),
  processed_at: z.string().optional(),
  retries: z.number().int().nonnegative(),
  last_error: z.string().optional(),
  linked_entity: linkedEntityRefSchema.optional(),
});

export const packetEventSchema = z.object({
  id: z.string(),
  packet_id: z.string(),
  event_type: z.enum([
    "RECEIVED",
    "DECODED",
    "HANDLED",
    "RETRY_SCHEDULED",
    "RETRY",
    "FAILED",
    "MANUAL_RESOLVE",
    "MANUAL_LINK",
  ]),
  actor: z.string(),
  status: z.enum(["SUCCEEDED", "FAILED", "RUNNING"]),
  message: z.string().optional(),
  timestamp: z.string(),
  duration_ms: z.number().int().nonnegative().optional(),
  // Backend gap: amie_processing_events has no trace_id column yet. The schema
  // accepts it as optional so the UI can surface ViewTraceLink when present.
  trace_id: z
    .string()
    .regex(/^[0-9a-f]{32}$/)
    .optional(),
});

export const replyStatusSchema = z.enum(["PENDING", "SENT", "ACKED", "FAILED"]);

export const replySchema = z.object({
  id: z.string(),
  amie_id: z.string(),
  type: z.string(),
  status: replyStatusSchema,
  in_reply_to_packet_id: z.string().optional(),
  sent_at: z.string().optional(),
  acked_at: z.string().optional(),
  retries: z.number().int().nonnegative(),
  last_error: z.string().optional(),
  created_at: z.string(),
});

export const packetListResponseSchema = z.object({
  packets: z.array(packetSchema),
  total: z.number().int().nonnegative(),
  limit: z.number().int().nonnegative(),
  offset: z.number().int().nonnegative(),
});

export const replyListResponseSchema = z.object({
  replies: z.array(replySchema),
  total: z.number().int().nonnegative(),
  limit: z.number().int().nonnegative(),
  offset: z.number().int().nonnegative(),
});

export const packetStatBucketSchema = z.object({
  date: z.string(),
  status: packetStatusSchema,
  type: z.string(),
  count: z.number().int().nonnegative(),
});

export const packetStatsSchema = z.object({
  byDay: z.array(packetStatBucketSchema),
});

export const resolvePacketPayloadSchema = z.object({
  reason: z.string().min(3).max(500),
});

export const linkUnmappedPayloadSchema = z.object({
  entity_type: z.enum(["project", "account", "person", "user_merge"]),
  entity_id: z.string().min(1),
});
