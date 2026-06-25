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

import { getLastTraceId, subscribeLastTraceId } from "@/shared/api/last-trace-id";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import * as React from "react";
import {
  type TraceListFilters,
  getAuditEventsForTrace,
  getTrace,
  listAuditSources,
  listTraces,
  retryTrace,
} from "./api";
import { useLastTraceContext } from "./components/LastTraceProvider";

const DEFAULT_STALE_TIME = 30_000;
const DEFAULT_GC_TIME = 5 * 60_000;

export const traceKeys = {
  all: ["traces"] as const,
  lists: () => [...traceKeys.all, "list"] as const,
  list: (filters: TraceListFilters) => [...traceKeys.lists(), filters] as const,
  details: () => [...traceKeys.all, "detail"] as const,
  detail: (id: string) => [...traceKeys.details(), id] as const,
  sources: () => [...traceKeys.all, "sources"] as const,
  audit: (id: string, spanId?: string) => [...traceKeys.all, "audit", id, spanId ?? null] as const,
};

export function useTraces(filters: TraceListFilters = {}, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: traceKeys.list(filters),
    queryFn: () => listTraces(filters),
    staleTime: DEFAULT_STALE_TIME,
    gcTime: DEFAULT_GC_TIME,
    refetchOnWindowFocus: false,
    enabled: options?.enabled ?? true,
  });
}

export function useTrace(id: string | undefined) {
  return useQuery({
    queryKey: id ? traceKeys.detail(id) : [...traceKeys.details(), "none"],
    queryFn: () => getTrace(id as string),
    enabled: Boolean(id),
    staleTime: DEFAULT_STALE_TIME,
    gcTime: DEFAULT_GC_TIME,
    refetchOnWindowFocus: false,
  });
}

export function useAuditSources() {
  return useQuery({
    queryKey: traceKeys.sources(),
    queryFn: listAuditSources,
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
    refetchOnWindowFocus: false,
  });
}

export function useAuditEventsForTrace(id: string | undefined, spanId?: string) {
  return useQuery({
    queryKey: id ? traceKeys.audit(id, spanId) : [...traceKeys.all, "audit", "none"],
    queryFn: () => getAuditEventsForTrace(id as string, spanId),
    enabled: Boolean(id),
    staleTime: DEFAULT_STALE_TIME,
    gcTime: DEFAULT_GC_TIME,
    refetchOnWindowFocus: false,
  });
}

export function useRetryTrace() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => retryTrace(id),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: traceKeys.all });
    },
  });
}

// Prefers LastTraceProvider's context value; falls back to a singleton
// subscription so consumers rendered outside the provider still observe updates.
export function useLastTraceId(): string | null {
  const fromContext = useLastTraceContext();
  const [traceId, setTraceId] = React.useState<string | null>(() => fromContext ?? null);
  React.useEffect(() => {
    if (fromContext !== undefined) return;
    setTraceId(getLastTraceId());
    return subscribeLastTraceId(setTraceId);
  }, [fromContext]);
  return fromContext ?? traceId;
}
