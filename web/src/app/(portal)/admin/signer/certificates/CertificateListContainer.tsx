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
  replaceShallowSearchParam,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { CertificateList } from "@/features/core/signer/components/CertificateList";
import { useCertificates, useCursorPaging } from "@/features/core/signer/queries";
import type { CertificateStatus } from "@/features/core/signer/status";
import { useUrlSearch } from "@/shared/hooks/useUrlSearch";

function statusFilterFromUrl(raw: string | null): CertificateStatus | "all" {
  if (raw === "active" || raw === "not-yet-valid" || raw === "expired" || raw === "revoked") {
    return raw;
  }
  return "all";
}

export function CertificateListContainer() {
  const searchParams = useShallowSearchParams();
  const search = useUrlSearch();
  const statusFilter = statusFilterFromUrl(searchParams.get("status"));
  const { params: pageParams, paging } = useCursorPaging({
    ...(search.settledQuery ? { q: search.settledQuery } : {}),
    ...(statusFilter === "all" ? {} : { status: statusFilter }),
  });

  const query = useCertificates(pageParams);
  const rows = query.data?.certificates ?? [];

  return (
    <CertificateList
      rows={rows}
      isLoading={query.isLoading}
      error={(query.error as Error | null) ?? null}
      onRetry={() => query.refetch()}
      search={search.text}
      onSearchChange={search.setText}
      statusFilter={statusFilter}
      onStatusFilterChange={(next) =>
        replaceShallowSearchParam("status", next === "all" ? null : next)
      }
      paging={paging(query.data?.next_cursor)}
    />
  );
}
