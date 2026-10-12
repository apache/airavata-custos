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

import { useClusters } from "@/features/core/clusters/queries";
import type { ComputeAllocationResourceSummary } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { formatNumber, formatSU } from "@/shared/format";
import {
  replaceShallowSearchParams,
  useCanonicalParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { queryUuid } from "@/shared/layout/nav";
import { DataTable, type DataTableColumn, NameCell } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import * as React from "react";
import { useResourceSummaries } from "../queries";
import { CreateResourceDialog } from "./CreateResourceDialog";
import { ResourceDetailDrawer } from "./ResourceDetailDrawer";
import { ResourcesRatesDrawer } from "./ResourcesRatesDrawer";

export function ResourcesTab() {
  const ability = useAbility();
  const canReadClusters = ability.can("read", "Cluster");
  const canCreate = ability.can("write", "Allocation");
  const query = useResourceSummaries();
  const clusters = useClusters(canReadClusters);
  // Without Cluster read the cluster column and filter are omitted; a bare cluster id names nothing.
  const clusterName = (id: string) =>
    clusters.data?.find((c) => c.id === id)?.name ?? (clusters.isSuccess ? "Unknown cluster" : "—");
  const [search, setSearch] = React.useState("");
  const [clusterId, setClusterId] = React.useState("");
  const params = useShallowSearchParams();
  const detailId = queryUuid(params.get("resource")) ?? null;
  const ratesId = queryUuid(params.get("rates")) ?? null;
  // One drawer at a time: opening either closes the other.
  const openDrawer = (key: "resource" | "rates", id: string | null) => {
    const next = new URLSearchParams(params);
    next.delete("resource");
    next.delete("rates");
    if (id) next.set(key, id);
    replaceShallowSearchParams(next);
  };
  const setDetailId = (id: string | null) => openDrawer("resource", id);
  const setRatesId = (id: string | null) => openDrawer("rates", id);
  useCanonicalParams({ resource: detailId ?? undefined, rates: ratesId ?? undefined });

  const rows = query.data ?? [];
  const clusterOptions = [...new Set(rows.map((row) => row.compute_cluster_id))]
    .filter(Boolean)
    .map((id) => [id, clusterName(id)] as const)
    .sort(([, a], [, b]) => a.localeCompare(b));
  const filtered = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    return rows.filter((row) => {
      if (clusterId && row.compute_cluster_id !== clusterId) return false;
      if (!needle) return true;
      return `${row.name} ${row.resource_type}`.toLowerCase().includes(needle);
    });
  }, [rows, search, clusterId]);

  const columns: Array<DataTableColumn<ComputeAllocationResourceSummary>> = [
    {
      key: "name",
      header: "Resource",
      cell: (row) => (
        <div>
          <NameCell text={row.name} className="font-medium text-foreground" />
          <NameCell text={row.resource_type} className="text-xs text-muted-foreground" />
        </div>
      ),
    },
    // Beside the name, so an open rates drawer never covers the triggers it swaps between.
    {
      key: "rates",
      header: "Rates",
      interactive: true,
      cell: (row) => (
        <Button
          variant="outline"
          size="sm"
          onClick={() => setRatesId(row.id)}
          data-row-opener="rates"
          aria-label={`Rates (${row.rate_count}) of ${row.name}`}
        >
          Rates ({row.rate_count})
        </Button>
      ),
    },
    ...(canReadClusters
      ? [
          {
            key: "cluster",
            header: "Cluster",
            cell: (row: ComputeAllocationResourceSummary) => (
              <NameCell text={clusterName(row.compute_cluster_id)} />
            ),
          },
        ]
      : []),
    {
      key: "allocation_count",
      header: "Allocations",
      align: "right",
      cell: (row) => (
        <span className="text-sm tabular-nums text-foreground">{row.allocation_count}</span>
      ),
    },
    {
      key: "total_allocated",
      header: "Allocated",
      align: "right",
      cell: (row) => (
        <span className="text-sm tabular-nums text-foreground">
          {formatNumber(row.total_allocated)}
        </span>
      ),
    },
    {
      key: "total_used_su",
      header: "Used SU",
      align: "right",
      cell: (row) => (
        <span className="text-sm tabular-nums text-foreground">{formatSU(row.total_used_su)}</span>
      ),
    },
  ];

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4 sm:flex-row sm:items-center">
        {canReadClusters ? (
          <select
            value={clusterId}
            onChange={(e) => setClusterId(e.target.value)}
            aria-label="Filter by cluster"
            className="h-9 rounded-md border bg-background px-3 text-sm sm:w-56"
          >
            <option value="">All clusters</option>
            {clusterOptions.map(([id, name]) => (
              <option key={id} value={id}>
                {name}
              </option>
            ))}
          </select>
        ) : null}
        <Input
          type="search"
          placeholder="Search resources"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Search resources"
          className="sm:w-72"
        />
        {canCreate ? (
          <div className="sm:ml-auto">
            <CreateResourceDialog clusters={clusters.data} />
          </div>
        ) : null}
      </div>

      {query.isLoading ? (
        <CardSkeleton />
      ) : query.error ? (
        <ErrorState error={query.error} onRetry={() => query.refetch()} />
      ) : rows.length === 0 ? (
        <EmptyState heading="No resources registered" />
      ) : filtered.length === 0 ? (
        <EmptyState heading="No resources match these filters" />
      ) : (
        <DataTable
          caption="Resources"
          columns={columns}
          rows={filtered}
          rowKey={(row) => row.id}
          onRowClick={(row) => setDetailId(row.id)}
        />
      )}

      <ResourceDetailDrawer
        resourceId={detailId}
        onOpenChange={(open) => {
          if (!open) setDetailId(null);
        }}
      />
      <ResourcesRatesDrawer
        resourceId={ratesId}
        resourceName={rows.find((row) => row.id === ratesId)?.name ?? null}
        onOpenChange={(open) => {
          if (!open) setRatesId(null);
        }}
      />
    </div>
  );
}
