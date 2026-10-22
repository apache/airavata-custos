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

import eventsFixture from "@/features/connectors/amie/__fixtures__/events.json";
import packetsFixture from "@/features/connectors/amie/__fixtures__/packets.json";
import repliesFixture from "@/features/connectors/amie/__fixtures__/replies.json";
import statsFixture from "@/features/connectors/amie/__fixtures__/stats.json";
import {
  linkedEntityRefSchema,
  packetEventSchema,
  packetListResponseSchema,
  replyListResponseSchema,
} from "@/features/connectors/amie/schemas";
import type { Packet } from "@/features/connectors/amie/types";
import { http, HttpResponse } from "msw";
import { z } from "zod";

const allPackets = packetListResponseSchema.parse(packetsFixture).packets;
const eventsByPacket = z.record(z.string(), z.array(packetEventSchema)).parse(eventsFixture);
const allReplies = replyListResponseSchema.parse(repliesFixture).replies;

type IdParams = { id: string };
const resolveBodySchema = z.object({ reason: z.string().optional() });
const linkBodySchema = z.object({
  entity_type: linkedEntityRefSchema.shape.type.optional(),
  entity_id: z.string().optional(),
});

function filterPackets(url: URL): {
  rows: Packet[];
  total: number;
  limit: number;
  offset: number;
} {
  const status = url.searchParams.get("status");
  const type = url.searchParams.get("type");
  const source = url.searchParams.get("source");
  const q = url.searchParams.get("q")?.toLowerCase() ?? "";
  const limit = Number(url.searchParams.get("limit") ?? "100");
  const offset = Number(url.searchParams.get("offset") ?? "0");

  let filtered = allPackets;
  if (status) filtered = filtered.filter((p) => p.status === status);
  if (type) filtered = filtered.filter((p) => p.type === type);
  if (source) filtered = filtered.filter((p) => p.source === source);
  if (q) {
    filtered = filtered.filter((p) => {
      const hay = `${p.amie_id} ${p.id} ${p.linked_entity?.display_id ?? ""}`.toLowerCase();
      return hay.includes(q);
    });
  }
  const total = filtered.length;
  const sliced = filtered.slice(offset, offset + limit);
  return { rows: sliced, total, limit, offset };
}

export const amieHandlers = [
  http.get("/api/v1/connectors/amie/packets", ({ request }) => {
    const url = new URL(request.url);
    const { rows, total, limit, offset } = filterPackets(url);
    return HttpResponse.json({ packets: rows, total, limit, offset });
  }),

  http.get<IdParams>("/api/v1/connectors/amie/packets/:id", ({ params }) => {
    const id = String(params.id);
    const packet = allPackets.find((p) => p.id === id);
    if (!packet) return HttpResponse.json({ error: "not_found" }, { status: 404 });
    return HttpResponse.json(packet);
  }),

  http.get<IdParams>("/api/v1/connectors/amie/packets/:id/events", ({ params }) => {
    const id = String(params.id);
    const rows = eventsByPacket[id] ?? [];
    return HttpResponse.json(rows);
  }),

  http.post<IdParams>("/api/v1/connectors/amie/packets/:id/retry", ({ params }) => {
    const id = String(params.id);
    const packet = allPackets.find((p) => p.id === id);
    if (!packet) return HttpResponse.json({ error: "not_found" }, { status: 404 });
    packet.retries += 1;
    packet.status = "DECODED";
    packet.updated_at = new Date().toISOString();
    return HttpResponse.json({ queued: true, packet });
  }),

  http.post<IdParams>(
    "/api/v1/connectors/amie/packets/:id/resolve",
    async ({ params, request }) => {
      const id = String(params.id);
      const packet = allPackets.find((p) => p.id === id);
      if (!packet) return HttpResponse.json({ error: "not_found" }, { status: 404 });
      const body = resolveBodySchema.parse(await request.json().catch(() => ({})));
      if (!body.reason || body.reason.trim().length < 3) {
        return HttpResponse.json({ error: "reason_required" }, { status: 400 });
      }
      packet.status = "PROCESSED";
      packet.processed_at = new Date().toISOString();
      packet.updated_at = packet.processed_at;
      return HttpResponse.json(packet);
    },
  ),

  http.get("/api/v1/connectors/amie/replies", ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get("status");
    const limit = Number(url.searchParams.get("limit") ?? "200");
    const offset = Number(url.searchParams.get("offset") ?? "0");
    let filtered = allReplies;
    if (status) filtered = filtered.filter((r) => r.status === status);
    return HttpResponse.json({
      replies: filtered.slice(offset, offset + limit),
      total: filtered.length,
      limit,
      offset,
    });
  }),

  http.post<IdParams>("/api/v1/connectors/amie/replies/:id/retry", ({ params }) => {
    const id = String(params.id);
    const reply = allReplies.find((r) => r.id === id);
    if (!reply) return HttpResponse.json({ error: "not_found" }, { status: 404 });
    reply.retries += 1;
    reply.status = "PENDING";
    return HttpResponse.json({ queued: true });
  }),

  http.get("/api/v1/connectors/amie/unmapped", ({ request }) => {
    const url = new URL(request.url);
    const limit = Number(url.searchParams.get("limit") ?? "100");
    const offset = Number(url.searchParams.get("offset") ?? "0");
    const filtered = allPackets.filter((p) => p.status === "DECODED" && !p.linked_entity);
    return HttpResponse.json({
      packets: filtered.slice(offset, offset + limit),
      total: filtered.length,
      limit,
      offset,
    });
  }),

  http.post<IdParams>("/api/v1/connectors/amie/unmapped/:id/link", async ({ params, request }) => {
    const id = String(params.id);
    const packet = allPackets.find((p) => p.id === id);
    if (!packet) return HttpResponse.json({ error: "not_found" }, { status: 404 });
    const body = linkBodySchema.parse(await request.json().catch(() => ({})));
    if (!body.entity_type || !body.entity_id) {
      return HttpResponse.json({ error: "entity_required" }, { status: 400 });
    }
    packet.linked_entity = {
      type: body.entity_type,
      id: body.entity_id,
      display_id: body.entity_id,
    };
    packet.status = "PROCESSED";
    packet.processed_at = new Date().toISOString();
    packet.updated_at = packet.processed_at;
    return HttpResponse.json(packet);
  }),

  http.get("/api/v1/connectors/amie/stats", () => {
    return HttpResponse.json(statsFixture);
  }),
];
