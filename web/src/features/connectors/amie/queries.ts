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

"use client";

import { lastReviewAt } from "@/features/core/cluster-accounts/review";
import {
  getConnectorsAmiePackets,
  getConnectorsAmiePacketsById,
  getConnectorsAmiePacketsByIdEvents,
  getConnectorsAmiePacketsByPacketIdAudits,
  getConnectorsAmieStats,
} from "@/generated/amie/sdk.gen";
import type {
  GetConnectorsAmiePacketsData,
  GetConnectorsAmieStatsData,
} from "@/generated/amie/types.gen";
import { paged } from "@/shared/api/paged";
import { skipToken, useQuery } from "@tanstack/react-query";

export const amieKeys = {
  all: ["amie"] as const,
  packets: (query: GetConnectorsAmiePacketsData["query"]) =>
    [...amieKeys.all, "packets", "list", query] as const,
  packet: (id: string) => [...amieKeys.all, "packets", "detail", id] as const,
  events: (id: string) => [...amieKeys.all, "packets", "events", id] as const,
  audits: (id: string) => [...amieKeys.all, "packets", "audits", id] as const,
  stats: (query: GetConnectorsAmieStatsData["query"]) => [...amieKeys.all, "stats", query] as const,
};

// Master replies to ACCESS seconds after a cluster-account review, so packets re-read every 2 s for 30 s after one.
const replyPoll = () => (Date.now() < lastReviewAt() + 30_000 ? 2_000 : false);

export function usePackets(query: GetConnectorsAmiePacketsData["query"]) {
  return useQuery({
    queryKey: amieKeys.packets(query),
    queryFn: () => getConnectorsAmiePackets({ query }),
    ...paged,
    refetchInterval: replyPoll,
  });
}

export function usePacket(id: string | undefined) {
  return useQuery({
    queryKey: amieKeys.packet(id ?? ""),
    queryFn: id ? () => getConnectorsAmiePacketsById({ path: { id } }) : skipToken,
    refetchInterval: replyPoll,
  });
}

export function usePacketEvents(id: string | undefined) {
  return useQuery({
    queryKey: amieKeys.events(id ?? ""),
    queryFn: id ? () => getConnectorsAmiePacketsByIdEvents({ path: { id } }) : skipToken,
    refetchInterval: replyPoll,
  });
}

export function usePacketAudits(id: string | undefined) {
  return useQuery({
    queryKey: amieKeys.audits(id ?? ""),
    queryFn: id
      ? () => getConnectorsAmiePacketsByPacketIdAudits({ path: { packet_id: id } })
      : skipToken,
    refetchInterval: replyPoll,
  });
}

export function usePacketStats(query: GetConnectorsAmieStatsData["query"]) {
  return useQuery({
    queryKey: amieKeys.stats(query),
    queryFn: () => getConnectorsAmieStats({ query }),
  });
}
