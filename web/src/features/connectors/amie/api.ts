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
import { apiFetch } from "@/shared/api/client";
import {
  linkUnmappedPayloadSchema,
  packetEventSchema,
  packetListResponseSchema,
  packetSchema,
  packetStatsSchema,
  replyListResponseSchema,
  resolvePacketPayloadSchema,
} from "./schemas";
import type {
  Packet,
  PacketEvent,
  PacketListResponse,
  PacketStats,
  PacketStatus,
  ReplyListResponse,
} from "./types";

export type PacketListParams = {
  status?: PacketStatus | "all" | string;
  type?: string;
  source?: string;
  q?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
};

function qs(params: Record<string, string | number | undefined | null>): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === "") continue;
    search.set(key, String(value));
  }
  const str = search.toString();
  return str ? `?${str}` : "";
}

export async function listPackets(params: PacketListParams = {}): Promise<PacketListResponse> {
  const raw = await apiFetch(`/connectors/amie/packets${qs(params)}`);
  return packetListResponseSchema.parse(raw);
}

export async function getPacket(id: string): Promise<Packet> {
  const raw = await apiFetch(`/connectors/amie/packets/${encodeURIComponent(id)}`);
  return packetSchema.parse(raw);
}

export async function getPacketEvents(id: string): Promise<PacketEvent[]> {
  const raw = await apiFetch(`/connectors/amie/packets/${encodeURIComponent(id)}/events`);
  return z.array(packetEventSchema).parse(raw);
}

export async function retryPacket(id: string): Promise<{ queued: true; packet: Packet }> {
  const raw = await apiFetch(`/connectors/amie/packets/${encodeURIComponent(id)}/retry`, {
    method: "POST",
  });
  return z.object({ queued: z.literal(true), packet: packetSchema }).parse(raw);
}

export async function resolvePacket(id: string, payload: { reason: string }): Promise<Packet> {
  const validated = resolvePacketPayloadSchema.parse(payload);
  const raw = await apiFetch(`/connectors/amie/packets/${encodeURIComponent(id)}/resolve`, {
    method: "POST",
    body: validated,
  });
  return packetSchema.parse(raw);
}

export type ReplyListParams = {
  status?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
};

export async function listReplies(params: ReplyListParams = {}): Promise<ReplyListResponse> {
  const raw = await apiFetch(`/connectors/amie/replies${qs(params)}`);
  return replyListResponseSchema.parse(raw);
}

export async function retryReply(id: string): Promise<{ queued: true }> {
  const raw = await apiFetch(`/connectors/amie/replies/${encodeURIComponent(id)}/retry`, {
    method: "POST",
  });
  return z.object({ queued: z.literal(true) }).parse(raw);
}

export type UnmappedListParams = {
  limit?: number;
  offset?: number;
};

export async function listUnmapped(params: UnmappedListParams = {}): Promise<PacketListResponse> {
  const raw = await apiFetch(`/connectors/amie/unmapped${qs(params)}`);
  return packetListResponseSchema.parse(raw);
}

export async function linkUnmapped(
  id: string,
  payload: { entity_type: string; entity_id: string },
): Promise<Packet> {
  const validated = linkUnmappedPayloadSchema.parse(payload);
  const raw = await apiFetch(`/connectors/amie/unmapped/${encodeURIComponent(id)}/link`, {
    method: "POST",
    body: validated,
  });
  return packetSchema.parse(raw);
}

export type StatsParams = {
  window?: string;
};

export async function getPacketStats(
  params: StatsParams = { window: "30d" },
): Promise<PacketStats> {
  const raw = await apiFetch(`/connectors/amie/stats${qs(params)}`);
  return packetStatsSchema.parse(raw);
}
