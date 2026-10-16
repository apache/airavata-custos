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

import type { z } from "zod";
import type {
  linkedEntityRefSchema,
  packetEventSchema,
  packetListResponseSchema,
  packetSchema,
  packetStatBucketSchema,
  packetStatsSchema,
  packetStatusSchema,
  replyListResponseSchema,
  replySchema,
  replyStatusSchema,
} from "./schemas";

export type PacketStatus = z.infer<typeof packetStatusSchema>;

export const PACKET_STATUSES: PacketStatus[] = [
  "NEW",
  "DECODED",
  "PROCESSED",
  "FAILED",
  "WAITING_APPROVAL",
  "REFUSED",
];

export function packetStatusLabel(status: PacketStatus): string {
  return status.replaceAll("_", " ");
}

export const PACKET_TYPES = [
  "request_project_create",
  "request_project_inactivate",
  "request_project_reactivate",
  "request_account_create",
  "request_account_inactivate",
  "request_account_reactivate",
  "request_person_merge",
  "request_user_modify",
  "data_account_create",
  "data_project_create",
  "inform_transaction_complete",
  "notify_project_create",
  "notify_account_create",
] as const;

export type PacketType = (typeof PACKET_TYPES)[number];

export type LinkedEntityRef = z.infer<typeof linkedEntityRefSchema>;

export type LinkedEntityType = LinkedEntityRef["type"];

export type Packet = z.infer<typeof packetSchema>;

export type PacketEvent = z.infer<typeof packetEventSchema>;

export type PacketEventType = PacketEvent["event_type"];

export type ReplyStatus = z.infer<typeof replyStatusSchema>;

export type Reply = z.infer<typeof replySchema>;

export type PacketStatBucket = z.infer<typeof packetStatBucketSchema>;

export type PacketStats = z.infer<typeof packetStatsSchema>;

export type PacketListResponse = z.infer<typeof packetListResponseSchema>;

export type ReplyListResponse = z.infer<typeof replyListResponseSchema>;
