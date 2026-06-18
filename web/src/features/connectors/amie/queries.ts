"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  type PacketListParams,
  type ReplyListParams,
  type StatsParams,
  type UnmappedListParams,
  getPacket,
  getPacketEvents,
  getPacketStats,
  linkUnmapped,
  listPackets,
  listReplies,
  listUnmapped,
  resolvePacket,
  retryPacket,
  retryReply,
} from "./api";

const DEFAULT_STALE_TIME = 30_000;
const DEFAULT_GC_TIME = 5 * 60_000;

export const amieKeys = {
  all: ["amie"] as const,
  packets: (params: PacketListParams = {}) =>
    [...amieKeys.all, "packets", "list", params] as const,
  packet: (id: string) => [...amieKeys.all, "packets", "detail", id] as const,
  events: (id: string) => [...amieKeys.all, "packets", "events", id] as const,
  replies: (params: ReplyListParams = {}) => [...amieKeys.all, "replies", params] as const,
  unmapped: (params: UnmappedListParams = {}) => [...amieKeys.all, "unmapped", params] as const,
  stats: (params: StatsParams = {}) => [...amieKeys.all, "stats", params] as const,
};

export function usePackets(params: PacketListParams = {}) {
  return useQuery({
    queryKey: amieKeys.packets(params),
    queryFn: () => listPackets(params),
    staleTime: DEFAULT_STALE_TIME,
    gcTime: DEFAULT_GC_TIME,
    refetchOnWindowFocus: false,
  });
}

export function usePacket(id: string | undefined) {
  return useQuery({
    queryKey: id ? amieKeys.packet(id) : [...amieKeys.all, "packets", "detail", "none"],
    queryFn: () => getPacket(id as string),
    enabled: Boolean(id),
    staleTime: DEFAULT_STALE_TIME,
    gcTime: DEFAULT_GC_TIME,
    refetchOnWindowFocus: false,
  });
}

export function usePacketEvents(id: string | undefined) {
  return useQuery({
    queryKey: id ? amieKeys.events(id) : [...amieKeys.all, "packets", "events", "none"],
    queryFn: () => getPacketEvents(id as string),
    enabled: Boolean(id),
    staleTime: DEFAULT_STALE_TIME,
    gcTime: DEFAULT_GC_TIME,
    refetchOnWindowFocus: false,
  });
}

export function useRetryPacket() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => retryPacket(id),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: amieKeys.all });
    },
  });
}

export function useResolvePacket() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => resolvePacket(id, { reason }),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: amieKeys.all });
    },
  });
}

export function useReplies(params: ReplyListParams = {}) {
  return useQuery({
    queryKey: amieKeys.replies(params),
    queryFn: () => listReplies(params),
    staleTime: DEFAULT_STALE_TIME,
    gcTime: DEFAULT_GC_TIME,
    refetchOnWindowFocus: false,
  });
}

export function useRetryReply() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => retryReply(id),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: amieKeys.all });
    },
  });
}

export function useUnmapped(params: UnmappedListParams = {}) {
  return useQuery({
    queryKey: amieKeys.unmapped(params),
    queryFn: () => listUnmapped(params),
    staleTime: DEFAULT_STALE_TIME,
    gcTime: DEFAULT_GC_TIME,
    refetchOnWindowFocus: false,
  });
}

export function useLinkUnmapped() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      entity_type,
      entity_id,
    }: { id: string; entity_type: string; entity_id: string }) =>
      linkUnmapped(id, { entity_type, entity_id }),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: amieKeys.all });
    },
  });
}

export function usePacketStats(params: StatsParams = { window: "30d" }) {
  return useQuery({
    queryKey: amieKeys.stats(params),
    queryFn: () => getPacketStats(params),
    staleTime: DEFAULT_STALE_TIME,
    gcTime: DEFAULT_GC_TIME,
    refetchOnWindowFocus: false,
  });
}
