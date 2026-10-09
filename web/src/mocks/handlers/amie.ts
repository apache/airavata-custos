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

import { http, HttpResponse } from "msw";
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
import { z } from "zod";
import { notFound, page } from "../paging";

const { packets } = zPacketListResponse.parse(packetsFixture);
const eventsByPacket = z.record(z.string(), z.array(zPacketEventResponse)).parse(eventsFixture);
const { byDay } = zPacketStatsResponse.parse(statsFixture);

// Shift fixture days so the latest lands on today, as a live 30d window would.
const lastDay = Date.parse(byDay.at(-1)?.date ?? "");
const dayShift = Number.isNaN(lastDay)
  ? 0
  : Date.parse(new Date().toISOString().slice(0, 10)) - lastDay;
const stats = {
  byDay: byDay.map((b) => ({
    ...b,
    date: new Date(Date.parse(b.date) + dayShift).toISOString().slice(0, 10),
  })),
};

function filterPackets(url: URL): PacketResponse[] {
  const status = url.searchParams.get("status");
  const type = url.searchParams.get("type");
  const q = url.searchParams.get("q")?.toLowerCase() ?? "";
  const to = Date.parse(url.searchParams.get("to") ?? "");
  return packets
    .filter(
      (p) =>
        (!status || p.status === status) &&
        (!type || p.type === type) &&
        (!q || `${p.amie_id} ${p.id}`.toLowerCase().includes(q)) &&
        (Number.isNaN(to) || Date.parse(p.received_at) <= to),
    )
    .sort((a, b) => b.received_at.localeCompare(a.received_at));
}

export const amieHandlers = [
  http.get("*/api/v1/connectors/amie/packets", ({ request }) => {
    const url = new URL(request.url);
    const { items, ...rest } = page(url, filterPackets(url));
    return HttpResponse.json({ packets: items, ...rest });
  }),

  http.get<GetConnectorsAmiePacketsByIdData["path"]>(
    "*/api/v1/connectors/amie/packets/:id",
    ({ params }) => {
      const packet = packets.find((p) => p.id === params.id);
      return packet ? HttpResponse.json(packet) : notFound("packet");
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

  // The backend has no reply or unmapped store yet and always returns an empty page.
  http.get("*/api/v1/connectors/amie/replies", ({ request }) => {
    const { items, ...rest } = page(new URL(request.url), []);
    return HttpResponse.json({ replies: items, ...rest });
  }),

  http.get("*/api/v1/connectors/amie/unmapped", ({ request }) => {
    const { items, ...rest } = page(new URL(request.url), []);
    return HttpResponse.json({ packets: items, ...rest });
  }),

  http.get("*/api/v1/connectors/amie/stats", () => HttpResponse.json(stats)),
];
