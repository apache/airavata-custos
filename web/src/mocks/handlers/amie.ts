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
import statsFixture from "@/features/connectors/amie/__fixtures__/stats.json";
import type {
  GetConnectorsAmiePacketsByIdData,
  GetConnectorsAmiePacketsByIdEventsData,
  GetConnectorsAmiePacketsByPacketIdAuditsData,
  PacketResponse,
} from "@/generated/amie/types.gen";
import {
  zPacketEventResponse,
  zPacketListResponse,
  zPacketStatsResponse,
} from "@/generated/amie/zod.gen";
import { http, HttpResponse } from "msw";
import { z } from "zod";

const packets = zPacketListResponse.parse(packetsFixture).packets ?? [];
const eventsByPacket = z.record(z.string(), z.array(zPacketEventResponse)).parse(eventsFixture);
const byDay = zPacketStatsResponse.parse(statsFixture).byDay ?? [];

// Shift fixture days so the latest lands on today, as a live 30d window would.
const dayShift =
  Date.parse(new Date().toISOString().slice(0, 10)) - Date.parse(byDay.at(-1)?.date ?? "");
const stats = {
  byDay: byDay.map((b) => ({
    ...b,
    date: new Date(Date.parse(b.date) + dayShift).toISOString().slice(0, 10),
  })),
};

function pageOf(url: URL) {
  return {
    limit: Number(url.searchParams.get("limit")) || 50,
    offset: Number(url.searchParams.get("offset")) || 0,
  };
}

function filterPackets(url: URL): PacketResponse[] {
  const status = url.searchParams.get("status");
  const type = url.searchParams.get("type");
  const q = url.searchParams.get("q")?.toLowerCase() ?? "";
  const to = Date.parse(url.searchParams.get("to") ?? "");
  const at = (p: PacketResponse) => Date.parse(p.received_at);
  return packets
    .filter(
      (p) =>
        (!status || p.status === status) &&
        (!type || p.type === type) &&
        (!q || `${p.amie_id} ${p.id}`.toLowerCase().includes(q)) &&
        (Number.isNaN(to) || at(p) <= to),
    )
    .sort((a, b) => at(b) - at(a));
}

export const amieHandlers = [
  http.get("*/api/v1/connectors/amie/packets", ({ request }) => {
    const url = new URL(request.url);
    const rows = filterPackets(url);
    const { limit, offset } = pageOf(url);
    return HttpResponse.json({
      packets: rows.slice(offset, offset + limit),
      total: rows.length,
      limit,
      offset,
    });
  }),

  http.get<GetConnectorsAmiePacketsByIdData["path"]>(
    "*/api/v1/connectors/amie/packets/:id",
    ({ params }) => {
      const packet = packets.find((p) => p.id === params.id);
      return packet
        ? HttpResponse.json(packet)
        : HttpResponse.json({ error: "not found" }, { status: 404 });
    },
  ),

  http.get<GetConnectorsAmiePacketsByIdEventsData["path"]>(
    "*/api/v1/connectors/amie/packets/:id/events",
    ({ params }) => HttpResponse.json(eventsByPacket[params.id] ?? []),
  ),

  http.get<GetConnectorsAmiePacketsByPacketIdAuditsData["path"]>(
    "*/api/v1/connectors/amie/packets/:packet_id/audits",
    ({ params }) => HttpResponse.json({ packet_id: params.packet_id, events: [] }),
  ),

  http.get("*/api/v1/connectors/amie/stats", () => HttpResponse.json(stats)),
];
