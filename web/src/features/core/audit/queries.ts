/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

"use client";

import { paged } from "@/shared/api/paged";
import { skipToken, useQuery } from "@tanstack/react-query";
import * as React from "react";
import { type TraceListFilters, getTrace, listAuditSources, listTraces } from "./api";

const DEFAULT_STALE_TIME = 30_000;
const DEFAULT_GC_TIME = 5 * 60_000;

export const traceKeys = {
  all: ["traces"] as const,
  lists: () => [...traceKeys.all, "list"] as const,
  list: (filters: TraceListFilters) => [...traceKeys.lists(), filters] as const,
  details: () => [...traceKeys.all, "detail"] as const,
  detail: (id: string) => [...traceKeys.details(), id] as const,
  sources: () => [...traceKeys.all, "sources"] as const,
};

// Lists are never served from cache; the shown page stays while the next loads. refetchOn joins the
// key, so a query refetches whenever it changes.
export function useTraces(filters: TraceListFilters = {}, refetchOn?: unknown) {
  return useQuery({
    queryKey: [...traceKeys.list(filters), refetchOn],
    queryFn: () => listTraces(filters),
    ...paged,
    staleTime: 0,
    gcTime: 0,
    refetchOnWindowFocus: false,
  });
}

export function useTrace(id: string | undefined) {
  return useQuery({
    queryKey: id ? traceKeys.detail(id) : [...traceKeys.details(), "none"],
    queryFn: id ? () => getTrace(id) : skipToken,
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
