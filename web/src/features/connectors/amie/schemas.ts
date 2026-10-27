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

import {
  zPacketEventResponse,
  zPacketListResponse,
  zPacketResponse,
  zPacketStatBucketResponse,
  zPacketStatsResponse,
  zReplyListResponse,
} from "@/generated/amie/zod.gen";
import { z } from "zod";

export const packetStatusSchema = z.enum([
  "NEW",
  "DECODED",
  "PROCESSED",
  "FAILED",
  "WAITING_APPROVAL",
  "REFUSED",
]);

export const linkedEntityRefSchema = z.object({
  type: z.enum(["project", "account", "person", "user_merge"]),
  id: z.string().min(1),
  display_id: z.string().optional(),
});

// The backend sends status as a free string, and neither decoded_payload nor linked_entity yet.
export const packetSchema = zPacketResponse.extend({
  status: packetStatusSchema,
  decoded_payload: z.record(z.string(), z.unknown()).optional(),
  linked_entity: linkedEntityRefSchema.optional(),
});

export const packetEventSchema = zPacketEventResponse.extend({
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
  status: z.enum(["SUCCEEDED", "FAILED", "RUNNING"]),
  // Backend gap: amie_processing_events has no trace_id column yet. The schema
  // accepts it as optional so the UI can surface ViewTraceLink when present.
  trace_id: z
    .string()
    .regex(/^[0-9a-f]{32}$/)
    .optional(),
});

export const replyStatusSchema = z.enum(["PENDING", "SENT", "ACKED", "FAILED"]);

// The backend lists replies as untyped values and always an empty list; this is the shape the UI renders.
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

export const packetListResponseSchema = zPacketListResponse.extend({
  packets: z.array(packetSchema),
});

export const replyListResponseSchema = zReplyListResponse.extend({
  replies: z.array(replySchema),
});

export const packetStatBucketSchema = zPacketStatBucketResponse.extend({
  status: packetStatusSchema,
});

export const packetStatsSchema = zPacketStatsResponse.extend({
  byDay: z.array(packetStatBucketSchema),
});

export const resolvePacketPayloadSchema = z.object({
  reason: z.string().min(3).max(500),
});

export const linkUnmappedPayloadSchema = z.object({
  entity_type: z.enum(["project", "account", "person", "user_merge"]),
  entity_id: z.string().min(1),
});
