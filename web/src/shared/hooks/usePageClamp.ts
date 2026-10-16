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

import { normalizeQueryId } from "@/shared/layout/nav";
import type { UseQueryResult } from "@tanstack/react-query";
import * as React from "react";
import { useDebounce } from "./useDebounce";
import {
  setSearchParam,
  useCanonicalParams,
  useShallowSearchParams,
} from "./useShallowSearchParams";

// ?page= of a paged list: a whole number from 1, else page 1.
export function pageParam(params: Pick<URLSearchParams, "get">): number {
  const raw = params.get("page") ?? "";
  return /^[1-9]\d*$/.test(raw) ? Number(raw) : 1;
}

// Rewrites only keys the URL carries, so params not yet read from the URL never erase it.
function onlyPresent(params: URLSearchParams, applied: Record<string, string | undefined>) {
  return Object.fromEntries(Object.entries(applied).filter(([key]) => params.has(key)));
}

export function useUrlPage() {
  const params = useShallowSearchParams();
  const page = pageParam(params);
  const setPage = (page: number) => setSearchParam(params, "page", page > 1 ? String(page) : null);
  useCanonicalParams(onlyPresent(params, { page: page > 1 ? String(page) : undefined }));
  return { params, page, setPage };
}

// A server-paged list filtered by URL params: the page returns to 1 whenever a filter changes.
export function useUrlFilteredPage(pageSize: number) {
  const { params, page, setPage } = useUrlPage();
  const setFilter = (key: string, value?: string | null) => {
    const next = new URLSearchParams(params);
    next.delete("page");
    setSearchParam(next, key, value);
  };
  return { params, page, setPage, offset: (page - 1) * pageSize, setFilter };
}

// A list's search box: the input keeps what is typed, the URL gets it trimmed once typing pauses,
// and a URL change from elsewhere (back, Clear filters) replaces the draft.
export function useSearchDraft(value: string, commit: (q: string) => void) {
  const [draft, setDraft] = React.useState(value);
  // biome-ignore lint/correctness/useExhaustiveDependencies: follows the URL only
  React.useEffect(() => {
    if (value !== draft.trim()) setDraft(value);
  }, [value]);
  const settled = useDebounce(draft, 300).trim();
  // biome-ignore lint/correctness/useExhaustiveDependencies: commits each pause once
  React.useEffect(() => {
    if (settled !== value) commit(settled);
  }, [settled]);
  return [draft, setDraft] as const;
}

type StatusSearchQuery<S> = { limit: number; offset: number; status?: S; q?: string };

// A server-paged list searched by `q` and filtered by `status`, all held in the URL with its page;
// a page past the end stays, for the list to offer First page. Returns list props.
export function useStatusSearchList<S extends string, T>(
  statuses: readonly S[],
  useList: (query: StatusSearchQuery<S>) => UseQueryResult<{ items: T[] | null; total: number }>,
) {
  const pageSize = 50;
  const { params, page, setPage, offset, setFilter } = useUrlFilteredPage(pageSize);
  const q = normalizeQueryId(params.get("q")) ?? "";
  const [search, setSearch] = useSearchDraft(q, (next) => setFilter("q", next));
  const statusFilter: S | "all" = statuses.find((s) => s === params.get("status")) ?? "all";
  useCanonicalParams(
    onlyPresent(params, {
      q: q || undefined,
      status: statusFilter === "all" ? undefined : statusFilter,
    }),
  );
  const query = useList({
    limit: pageSize,
    offset,
    status: statusFilter === "all" ? undefined : statusFilter,
    q: q || undefined,
  });
  return {
    rows: query.data?.items ?? [],
    isLoading: query.isLoading,
    error: query.error,
    onRetry: () => query.refetch(),
    search,
    onSearchChange: setSearch,
    statusFilter,
    onStatusFilterChange: (next: S | "all") => setFilter("status", next === "all" ? null : next),
    pagination: { page, pageSize, total: query.data?.total ?? 0, onPageChange: setPage },
  };
}
