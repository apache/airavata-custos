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
  getAuditEvents,
  getAuditSources,
  getAuditTraces,
  getAuditTracesByTraceId,
} from "@/generated/core/sdk.gen";
import type { GetAuditTracesData } from "@/generated/core/types.gen";
import { skipToken, useQuery } from "@tanstack/react-query";
import { type WindowPreset, bannerBounds, windowToFromTo } from "./components/traceListUrlState";
import { traceView } from "./utils";

export const traceKeys = {
  all: ["traces"] as const,
  list: (query: GetAuditTracesData["query"], window: TraceWindow) =>
    [...traceKeys.all, "list", query, window] as const,
  detail: (id: string) => [...traceKeys.all, "detail", id] as const,
  sources: () => [...traceKeys.all, "sources"] as const,
  audit: (id: string, spanId?: string) => [...traceKeys.all, "audit", id, spanId ?? null] as const,
};

// "failing24h" is the banner's 30d->24h-ago range.
type TraceWindow = WindowPreset | "failing24h";

// Bounds are computed per fetch so refetches slide the window forward.
export function useTraces(query: GetAuditTracesData["query"], window: TraceWindow) {
  return useQuery({
    queryKey: traceKeys.list(query, window),
    queryFn: () => {
      const now = Date.now();
      const bounds = window === "failing24h" ? bannerBounds(now) : windowToFromTo(window, now);
      return getAuditTraces({ query: { ...query, ...bounds } });
    },
  });
}

export function useTrace(id: string | undefined) {
  return useQuery({
    queryKey: traceKeys.detail(id ?? ""),
    queryFn: id ? () => getAuditTracesByTraceId({ path: { trace_id: id } }) : skipToken,
    select: traceView,
  });
}

export function useAuditSources() {
  return useQuery({
    queryKey: traceKeys.sources(),
    queryFn: () => getAuditSources(),
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
  });
}

export function useAuditEventsForTrace(id: string | undefined, spanId?: string) {
  return useQuery({
    queryKey: traceKeys.audit(id ?? "", spanId),
    queryFn: id ? () => getAuditEvents({ query: { trace_id: id, span_id: spanId } }) : skipToken,
  });
}
