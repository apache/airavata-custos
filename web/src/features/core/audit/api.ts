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

import { apiFetch } from "@/shared/api/client";
import { traceDetailSchema, traceListSchema, traceSourceListSchema } from "./schemas";
import type { TraceDetail, TraceListResponse, TraceStatus } from "./types";

export type TraceListFilters = {
  source?: string[];
  status?: TraceStatus[];
  from?: string;
  to?: string;
  q?: string;
  limit?: number;
  offset?: number;
};

// Sorted so the same filters make the same URL and the same query key.
function qs(filters: TraceListFilters): string {
  const search = new URLSearchParams();
  for (const s of [...(filters.source ?? [])].sort()) search.append("source", s);
  for (const s of [...(filters.status ?? [])].sort()) search.append("status", s);
  if (filters.from) search.set("from", filters.from);
  if (filters.to) search.set("to", filters.to);
  if (filters.q) search.set("q", filters.q);
  if (typeof filters.limit === "number") search.set("limit", String(filters.limit));
  if (typeof filters.offset === "number") search.set("offset", String(filters.offset));
  const str = search.toString();
  return str ? `?${str}` : "";
}

export async function listTraces(filters: TraceListFilters = {}): Promise<TraceListResponse> {
  return traceListSchema.parse(await apiFetch(`/audit/traces${qs(filters)}`));
}

export async function getTrace(traceId: string): Promise<TraceDetail> {
  return traceDetailSchema.parse(await apiFetch(`/audit/traces/${encodeURIComponent(traceId)}`));
}

export async function listAuditSources(): Promise<string[]> {
  return traceSourceListSchema.parse(await apiFetch("/audit/sources")).sources;
}
