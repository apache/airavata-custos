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

import {
  getConnectorsAmiePackets,
  getConnectorsAmiePacketsById,
  getConnectorsAmiePacketsByIdEvents,
  getConnectorsAmiePacketsByPacketIdAudits,
  getConnectorsAmieReplies,
  getConnectorsAmieStats,
  getConnectorsAmieUnmapped,
  postConnectorsAmiePacketsByIdResolve,
  postConnectorsAmiePacketsByIdRetry,
  postConnectorsAmieRepliesByIdRetry,
  postConnectorsAmieUnmappedByIdLink,
} from "@/generated/amie/sdk.gen";
import type {
  GetConnectorsAmiePacketsData,
  GetConnectorsAmieRepliesData,
  GetConnectorsAmieStatsData,
} from "@/generated/amie/types.gen";
import { skipToken, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

export const amieKeys = {
  all: ["amie"] as const,
  packets: (query: GetConnectorsAmiePacketsData["query"]) =>
    [...amieKeys.all, "packets", "list", query] as const,
  packet: (id: string) => [...amieKeys.all, "packets", "detail", id] as const,
  events: (id: string) => [...amieKeys.all, "packets", "events", id] as const,
  audits: (id: string) => [...amieKeys.all, "packets", "audits", id] as const,
  replies: (query: GetConnectorsAmieRepliesData["query"]) =>
    [...amieKeys.all, "replies", query] as const,
  unmapped: () => [...amieKeys.all, "unmapped"] as const,
  stats: (query: GetConnectorsAmieStatsData["query"]) =>
    [...amieKeys.all, "stats", query] as const,
};

export function usePackets(query: GetConnectorsAmiePacketsData["query"]) {
  return useQuery({
    queryKey: amieKeys.packets(query),
    queryFn: () => getConnectorsAmiePackets({ query }),
  });
}

export function usePacket(id: string | undefined) {
  return useQuery({
    queryKey: amieKeys.packet(id ?? ""),
    queryFn: id ? () => getConnectorsAmiePacketsById({ path: { id } }) : skipToken,
  });
}

export function usePacketEvents(id: string | undefined) {
  return useQuery({
    queryKey: amieKeys.events(id ?? ""),
    queryFn: id ? () => getConnectorsAmiePacketsByIdEvents({ path: { id } }) : skipToken,
  });
}

export function usePacketAudits(id: string | undefined) {
  return useQuery({
    queryKey: amieKeys.audits(id ?? ""),
    queryFn: id
      ? () => getConnectorsAmiePacketsByPacketIdAudits({ path: { packet_id: id } })
      : skipToken,
  });
}

export function useReplies(query: GetConnectorsAmieRepliesData["query"]) {
  return useQuery({
    queryKey: amieKeys.replies(query),
    queryFn: () => getConnectorsAmieReplies({ query }),
  });
}

export function useUnmapped() {
  return useQuery({
    queryKey: amieKeys.unmapped(),
    queryFn: () => getConnectorsAmieUnmapped(),
  });
}

export function usePacketStats(query: GetConnectorsAmieStatsData["query"]) {
  return useQuery({
    queryKey: amieKeys.stats(query),
    queryFn: () => getConnectorsAmieStats({ query }),
  });
}

const AMIE_ACTIONS = {
  retryPacket: {
    done: "Retry queued",
    call: (id: string) => postConnectorsAmiePacketsByIdRetry({ path: { id } }),
  },
  resolvePacket: {
    done: "Marked processed",
    call: (id: string) => postConnectorsAmiePacketsByIdResolve({ path: { id } }),
  },
  retryReply: {
    done: "Reply retry queued",
    call: (id: string) => postConnectorsAmieRepliesByIdRetry({ path: { id } }),
  },
  linkUnmapped: {
    done: "Packet linked",
    call: (id: string) => postConnectorsAmieUnmappedByIdLink({ path: { id } }),
  },
};

// One mutation for single-item and bulk calls: every id is attempted, and any failure
// surfaces as an error toast carrying the backend message instead of a success.
export function useAmieAction(action: keyof typeof AMIE_ACTIONS) {
  const client = useQueryClient();
  const { done, call } = AMIE_ACTIONS[action];
  return useMutation({
    mutationFn: async (ids: string[]) => {
      const failures = (await Promise.allSettled(ids.map(call))).flatMap((r) =>
        r.status === "rejected" ? [r.reason] : [],
      );
      if (failures.length === 0) return ids.length;
      const reason = failures[0] instanceof Error ? failures[0].message : String(failures[0]);
      throw new Error(
        ids.length > 1 ? `${failures.length} of ${ids.length} failed: ${reason}` : reason,
      );
    },
    onSuccess: (count) => toast.success(count > 1 ? `${done} (${count})` : done),
    onSettled: () => client.invalidateQueries({ queryKey: amieKeys.all }),
  });
}
