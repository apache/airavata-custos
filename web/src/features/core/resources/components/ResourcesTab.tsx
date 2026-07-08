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
import { useClusters } from "@/features/core/clusters/queries";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { CardSkeleton } from "@/shared/ui/Loading";
import { useResources } from "../queries";
import type { ComputeAllocationResource } from "../schemas";
import { ResourcesRatesDrawer } from "./ResourcesRatesDrawer";

export function ResourcesTab() {
  const query = useResources();
  const clustersQuery = useClusters();
  const [search, setSearch] = React.useState("");
  const [clusterId, setClusterId] = React.useState("");
  const [ratesFor, setRatesFor] = React.useState<ComputeAllocationResource | null>(null);

  const clusters = clustersQuery.data ?? [];
  const clusterName = (id: string) => clusters.find((c) => c.id === id)?.name ?? id;

  const rows = query.data ?? [];
  const filtered = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    return rows.filter((row) => {
      if (clusterId && row.compute_cluster_id !== clusterId) return false;
      if (!needle) return true;
      return `${row.name} ${row.resource_type}`.toLowerCase().includes(needle);
    });
  }, [rows, search, clusterId]);

  const columns: Array<DataTableColumn<ComputeAllocationResource>> = [
    {
      key: "name",
      header: "Resource",
      cell: (row) => (
        <div>
          <div className="font-medium text-foreground">{row.name}</div>
          <div className="text-xs text-muted-foreground">{row.resource_type}</div>
        </div>
      ),
    },
    {
      key: "cluster",
      header: "Cluster",
      cell: (row) => (
        <span className="text-sm text-foreground">{clusterName(row.compute_cluster_id)}</span>
      ),
    },
    {
      key: "resource_amount",
      header: "Amount",
      align: "right",
      cell: (row) => (
        <span className="text-sm tabular-nums text-foreground">{row.resource_amount}</span>
      ),
    },
    {
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (row) => (
        <Button variant="outline" size="sm" onClick={() => setRatesFor(row)}>
          Rates
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4 sm:flex-row sm:items-center">
        <select
          value={clusterId}
          onChange={(e) => setClusterId(e.target.value)}
          aria-label="Filter by cluster"
          className="h-9 rounded-md border bg-background px-3 text-sm sm:w-56"
        >
          <option value="">All clusters</option>
          {clusters.map((cluster) => (
            <option key={cluster.id} value={cluster.id}>
              {cluster.name}
            </option>
          ))}
        </select>
        <Input
          type="search"
          placeholder="Search resources"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Search resources"
          className="sm:w-72"
        />
      </div>

      {query.isLoading ? (
        <CardSkeleton />
      ) : query.error ? (
        <ErrorState message={(query.error as Error).message} onRetry={() => query.refetch()} />
      ) : rows.length === 0 ? (
        <EmptyState heading="No resources registered." />
      ) : filtered.length === 0 ? (
        <EmptyState heading="No resources match these filters." />
      ) : (
        <DataTable columns={columns} rows={filtered} rowKey={(row) => row.id} />
      )}

      <ResourcesRatesDrawer
        resourceId={ratesFor?.id ?? null}
        resourceName={ratesFor?.name ?? null}
        onOpenChange={(open) => {
          if (!open) setRatesFor(null);
        }}
      />
    </div>
  );
}
