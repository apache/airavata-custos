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

import * as React from "react";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { CertificateList } from "@/features/core/signer/components/CertificateList";
import { useCertificates } from "@/features/core/signer/queries";
import type { CertificateStatus } from "@/features/core/signer/status";

function statusFilterFromUrl(raw: string | null): CertificateStatus | "all" {
  if (raw === "active" || raw === "expired" || raw === "revoked") return raw;
  return "all";
}

export function CertificateListContainer() {
  const searchParams = useShallowSearchParams();
  const [page, setPage] = React.useState(1);
  const [pageSize, setPageSize] = React.useState(20);

  const search = searchParams.get("q") ?? "";
  const statusFilter = statusFilterFromUrl(searchParams.get("status"));

  const updateParam = React.useCallback(
    (key: string, value: string | null) => {
      const params = new URLSearchParams(searchParams.toString());
      if (!value) params.delete(key);
      else params.set(key, value);
      replaceShallowSearchParams(params);
    },
    [searchParams],
  );

  const query = useCertificates({ limit: pageSize, offset: (page - 1) * pageSize });
  const rows = query.data?.certificates ?? [];
  const total = query.data?.total ?? 0;

  return (
    <CertificateList
      rows={rows}
      isLoading={query.isLoading}
      error={(query.error as Error | null) ?? null}
      onRetry={() => query.refetch()}
      search={search}
      onSearchChange={(next) => updateParam("q", next)}
      statusFilter={statusFilter}
      onStatusFilterChange={(next) => updateParam("status", next === "all" ? null : next)}
      pagination={{
        page,
        pageSize,
        total,
        onPageChange: setPage,
        onPageSizeChange: setPageSize,
      }}
    />
  );
}
