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
import { AllocationsList } from "@/features/core/allocations/components/AllocationsList";
import { CreateAllocationDialog } from "@/features/core/allocations/components/CreateAllocationDialog";
import { useAllocations } from "@/features/core/allocations/queries";
import type { AllocationStatus } from "@/generated/core/zod.gen";
import { requires, usePermissions } from "@/shared/casl/permissions";
import { GuardedButton } from "@/shared/ui/GuardedButton";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";

function statusFilterFromUrl(raw: string | null): AllocationStatus | "all" {
  if (raw === "ACTIVE" || raw === "INACTIVE" || raw === "DELETED") return raw;
  return "all";
}

export function AllocationsListContainer() {
  const can = usePermissions();
  const searchParams = useShallowSearchParams();
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

  const query = useAllocations({});
  const rows = query.data?.items ?? [];
  const [creating, setCreating] = React.useState(false);

  return (
    <AllocationsList
      rows={rows}
      isLoading={query.isLoading}
      error={query.error}
      onRetry={() => query.refetch()}
      search={search}
      onSearchChange={(next) => updateParam("q", next)}
      statusFilter={statusFilter}
      onStatusFilterChange={(next) => updateParam("status", next === "all" ? null : next)}
      canEditResources={can.allocationsWrite}
      headerCta={
        <>
          <GuardedButton
            allowed={can.allocationsWrite && can.clustersRead}
            reason={requires("allocationsWrite", "clustersRead")}
            onClick={() => setCreating(true)}
          >
            + Add allocation
          </GuardedButton>
          {creating ? <CreateAllocationDialog onOpenChange={setCreating} /> : null}
        </>
      }
    />
  );
}
