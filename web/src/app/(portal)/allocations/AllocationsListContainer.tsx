"use client";

import * as React from "react";
import { AllocationsList } from "@/features/core/allocations/components/AllocationsList";
import { useAllocations } from "@/features/core/allocations/queries";
import type { AllocationStatus } from "@/features/core/allocations/schemas";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";

function statusFilterFromUrl(raw: string | null): AllocationStatus | "all" {
  if (raw === "ACTIVE" || raw === "INACTIVE" || raw === "DELETED") return raw;
  return "all";
}

export function AllocationsListContainer() {
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

  return (
    <AllocationsList
      rows={rows}
      isLoading={query.isLoading}
      error={(query.error as Error | null) ?? null}
      onRetry={() => query.refetch()}
      search={search}
      onSearchChange={(next) => updateParam("q", next)}
      statusFilter={statusFilter}
      onStatusFilterChange={(next) => updateParam("status", next === "all" ? null : next)}
    />
  );
}
