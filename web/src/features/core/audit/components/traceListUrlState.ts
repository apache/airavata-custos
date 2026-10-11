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

import type { ListStatus, TraceStatus } from "../types";

export type WindowPreset = "24h" | "7d" | "30d";

export type ListFilters = {
  status: ListStatus[];
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

export const STATUS_FILTERS: ReadonlyArray<ListStatus> = ["failed", "retrying", "waiting", "done"];
const VALID_WINDOWS: ReadonlyArray<WindowPreset> = ["24h", "7d", "30d"];
const VALID_PAGE_SIZES: ReadonlyArray<number> = [25, 50, 100];

type SearchParamsLike = {
  getAll: (key: string) => string[];
  get: (key: string) => string | null;
};

const DAY_MS = 24 * 60 * 60 * 1000;

export function parseFilters(params: SearchParamsLike): ListFilters {
  const status = params
    .getAll("status")
    .filter((s): s is ListStatus => STATUS_FILTERS.includes(s as ListStatus));
  const source = params.getAll("source").filter((s) => /^[a-z][a-z0-9-]*$/.test(s));
  const winRaw = params.get("window");
  const window: WindowPreset = VALID_WINDOWS.includes(winRaw as WindowPreset)
    ? (winRaw as WindowPreset)
    : DEFAULT_FILTERS.window;
  const q = (params.get("q") ?? "").trim();
  const pageRaw = Number.parseInt(params.get("page") ?? "", 10);
  const page = Number.isFinite(pageRaw) && pageRaw >= 1 ? pageRaw : DEFAULT_FILTERS.page;
  const pageSizeRaw = Number.parseInt(params.get("pageSize") ?? "", 10);
  const pageSize = VALID_PAGE_SIZES.includes(pageSizeRaw) ? pageSizeRaw : DEFAULT_FILTERS.pageSize;
  return { status, source, window, q, page, pageSize };
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

export function windowToFromTo(win: WindowPreset, now: number): { from: string; to: string } {
  const days = win === "24h" ? 1 : win === "7d" ? 7 : 30;
  const to = new Date(now).toISOString();
  const from = new Date(now - days * DAY_MS).toISOString();
  return { from, to };
}

// Retrying and Waiting are both in_progress on the wire. When only one of them
// is picked the page filters the rows it got.
export function statusFiltersToApi(status: ListStatus[]): {
  apiStatus: TraceStatus[];
  keep: (s: ListStatus) => boolean;
} {
  const apiStatus = new Set<TraceStatus>();
  for (const s of status) {
    if (s === "failed") apiStatus.add("error");
    else if (s === "done") apiStatus.add("ok");
    else apiStatus.add("in_progress");
  }
  const keep = (s: ListStatus) => status.length === 0 || status.includes(s);
  return { apiStatus: [...apiStatus], keep };
}
