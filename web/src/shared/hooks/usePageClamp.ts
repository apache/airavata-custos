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

import type { UseQueryResult } from "@tanstack/react-query";
import * as React from "react";
import { useDebounce } from "./useDebounce";
import { setSearchParam, useShallowSearchParams } from "./useShallowSearchParams";

// A server-paged list filtered by URL params: the page returns to 1 whenever a filter changes.
export function useUrlFilteredPage(pageSize: number) {
  const params = useShallowSearchParams();
  const [page, setPage] = React.useState(1);
  const setFilter = (key: string, value?: string | null) => {
    setSearchParam(params, key, value);
    setPage(1);
  };
  return { params, page, setPage, offset: (page - 1) * pageSize, setFilter };
}

// Steps a server-paged list back to its last page once a mutation or filter empties it.
export function usePageClamp(
  page: number,
  setPage: (page: number) => void,
  total: number | undefined,
  pageSize: number,
) {
  React.useEffect(() => {
    const last = Math.max(1, Math.ceil((total ?? 0) / pageSize));
    if (total !== undefined && page > last) setPage(last);
  }, [page, setPage, total, pageSize]);
}

type StatusSearchQuery<S> = { limit: number; offset: number; status?: S; q?: string };

// A server-paged list searched by `q` and filtered by `status`, both held in the URL; returns list props.
export function useStatusSearchList<S extends string, T>(
  statuses: readonly S[],
  useList: (query: StatusSearchQuery<S>) => UseQueryResult<{ items: T[]; total: number }>,
) {
  const pageSize = 50;
  const { params, page, setPage, offset, setFilter } = useUrlFilteredPage(pageSize);
  const search = params.get("q") ?? "";
  const q = useDebounce(search.trim(), 300);
  const statusFilter: S | "all" = statuses.find((s) => s === params.get("status")) ?? "all";
  const query = useList({
    limit: pageSize,
    offset,
    status: statusFilter === "all" ? undefined : statusFilter,
    q: q || undefined,
  });
  usePageClamp(page, setPage, query.data?.total, pageSize);
  return {
    rows: query.data?.items ?? [],
    isLoading: query.isLoading,
    error: query.error,
    onRetry: () => query.refetch(),
    search,
    onSearchChange: (next: string) => setFilter("q", next),
    statusFilter,
    onStatusFilterChange: (next: S | "all") => setFilter("status", next === "all" ? null : next),
    pagination: { page, pageSize, total: query.data?.total ?? 0, onPageChange: setPage },
  };
}
