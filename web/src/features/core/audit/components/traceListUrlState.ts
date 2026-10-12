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

// URL <-> filter mapping for the trace list. Stray values collapse to defaults
// so a pasted URL never widens the query cache key.

import { pageParam } from "@/shared/hooks/usePageClamp";
import { normalizeQueryId } from "@/shared/layout/nav";
import type { TraceStatus } from "../types";

export type WindowPreset = "24h" | "7d" | "30d";

export type ListFilters = {
  status: TraceStatus[];
  source: string[];
  window: WindowPreset;
  q: string;
  page: number;
  pageSize: number;
};

export const DEFAULT_FILTERS: ListFilters = {
  status: [],
  source: [],
  window: "7d",
  q: "",
  page: 1,
  pageSize: 50,
};

export const STATUS_FILTERS: ReadonlyArray<TraceStatus> = ["error", "in_progress", "ok"];
const TRACE_ID = /^[0-9a-f]{32}$/;

// A ?trace= id: 32 hex, lowercased, else undefined, so no request is made for it.
export function queryTraceId(value: string | null | undefined): string | undefined {
  const id = normalizeQueryId(value);
  return id && TRACE_ID.test(id) ? id : undefined;
}

const VALID_WINDOWS: ReadonlyArray<WindowPreset> = ["24h", "7d", "30d"];
const VALID_PAGE_SIZES: ReadonlyArray<number> = [25, 50, 100];

type SearchParamsLike = {
  getAll: (key: string) => string[];
  get: (key: string) => string | null;
};

const DAY_MS = 24 * 60 * 60 * 1000;

// sources is GET /audit/sources; until it loads, any named source is kept.
export function parseFilters(
  params: SearchParamsLike,
  sources?: ReadonlyArray<string>,
): ListFilters {
  const status = params
    .getAll("status")
    .filter((s): s is TraceStatus => STATUS_FILTERS.includes(s as TraceStatus));
  const source = params
    .getAll("source")
    .filter((s) => (sources ? sources.includes(s) : s.trim() !== ""));
  const winRaw = params.get("window");
  const window: WindowPreset = VALID_WINDOWS.includes(winRaw as WindowPreset)
    ? (winRaw as WindowPreset)
    : DEFAULT_FILTERS.window;
  const q = normalizeQueryId(params.get("q")) ?? "";
  const pageSizeRaw = params.get("pageSize") ?? "";
  const pageSize =
    VALID_PAGE_SIZES.find((n) => String(n) === pageSizeRaw) ?? DEFAULT_FILTERS.pageSize;
  return { status, source, window, q, page: pageParam(params), pageSize };
}

export function serializeFilters(filters: ListFilters): URLSearchParams {
  const params = new URLSearchParams();
  for (const s of [...filters.status].sort()) params.append("status", s);
  for (const s of [...filters.source].sort()) params.append("source", s);
  if (filters.window !== DEFAULT_FILTERS.window) params.set("window", filters.window);
  if (filters.q) params.set("q", filters.q);
  if (filters.page !== DEFAULT_FILTERS.page) params.set("page", String(filters.page));
  if (filters.pageSize !== DEFAULT_FILTERS.pageSize) {
    params.set("pageSize", String(filters.pageSize));
  }
  return params;
}

export function hasActiveFilters(filters: ListFilters): boolean {
  return (
    filters.status.length > 0 ||
    filters.source.length > 0 ||
    filters.window !== DEFAULT_FILTERS.window ||
    filters.q.length > 0
  );
}

// Presets send no `to`, so the server reads up to now and a refresh shows new traces.
export function windowFrom(win: WindowPreset, now: number): string {
  const days = win === "24h" ? 1 : win === "7d" ? 7 : 30;
  return new Date(now - days * DAY_MS).toISOString();
}
