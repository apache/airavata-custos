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

export type PacketStatus = "NEW" | "DECODED" | "PROCESSED" | "FAILED";

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

export type LinkedEntityType = "project" | "account" | "person" | "user_merge";

export type LinkedEntityRef = {
  type: LinkedEntityType;
  id: string;
  display_id?: string;
};

export type Packet = {
  id: string;
  amie_id: string;
  type: string;
  status: PacketStatus;
  source: string;
  raw_json?: string;
  decoded_payload?: Record<string, unknown>;
  received_at: string;
  updated_at: string;
  decoded_at?: string;
  processed_at?: string;
  retries: number;
  last_error?: string;
  linked_entity?: LinkedEntityRef;
};

export type PacketEventType =
  | "RECEIVED"
  | "DECODED"
  | "HANDLED"
  | "RETRY_SCHEDULED"
  | "RETRY"
  | "FAILED"
  | "MANUAL_RESOLVE"
  | "MANUAL_LINK";

export type PacketEvent = {
  id: string;
  packet_id: string;
  event_type: PacketEventType;
  actor: string;
  status: "SUCCEEDED" | "FAILED" | "RUNNING";
  message?: string;
  timestamp: string;
  duration_ms?: number;
  trace_id?: string;
};

export type ReplyStatus = "PENDING" | "SENT" | "ACKED" | "FAILED";

export type Reply = {
  id: string;
  amie_id: string;
  type: string;
  status: ReplyStatus;
  in_reply_to_packet_id?: string;
  sent_at?: string;
  acked_at?: string;
  retries: number;
  last_error?: string;
  created_at: string;
};

export type PacketStatBucket = {
  date: string;
  status: PacketStatus;
  type: string;
  count: number;
};

export type PacketStats = {
  byDay: PacketStatBucket[];
};

export type PacketListResponse = {
  packets: Packet[];
  total: number;
  limit: number;
  offset: number;
};

export type ReplyListResponse = {
  replies: Reply[];
  total: number;
  limit: number;
  offset: number;
};
