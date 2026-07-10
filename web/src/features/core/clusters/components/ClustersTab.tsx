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
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { CardSkeleton } from "@/shared/ui/Loading";
import { useClusters } from "../queries";
import type { ComputeCluster } from "../schemas";
import { ClusterUsersDrawer } from "./ClusterUsersDrawer";

export function ClustersTab() {
  const query = useClusters();
  const [search, setSearch] = React.useState("");
  const [selected, setSelected] = React.useState<ComputeCluster | null>(null);

  const rows = query.data ?? [];
  const filtered = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    if (!needle) return rows;
    return rows.filter((row) => `${row.name} ${row.id}`.toLowerCase().includes(needle));
  }, [rows, search]);

  const columns: Array<DataTableColumn<ComputeCluster>> = [
    {
      key: "name",
      header: "Cluster",
      cell: (row) => <span className="font-medium text-foreground">{row.name}</span>,
    },
    {
      key: "id",
      header: "ID",
      cell: (row) => <span className="font-mono text-xs text-muted-foreground">{row.id}</span>,
    },
  ];

  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-border bg-card p-4">
        <Input
          type="search"
          placeholder="Search clusters"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Search clusters"
          className="sm:w-72"
        />
      </div>

      {query.isLoading ? (
        <CardSkeleton />
      ) : query.error ? (
        <ErrorState message={(query.error as Error).message} onRetry={() => query.refetch()} />
      ) : rows.length === 0 ? (
        <EmptyState heading="No clusters registered." />
      ) : filtered.length === 0 ? (
        <EmptyState heading="No clusters match this search." />
      ) : (
        <DataTable
          columns={columns}
          rows={filtered}
          rowKey={(row) => row.id}
          onRowClick={(row) => setSelected(row)}
        />
      )}

      <ClusterUsersDrawer
        clusterId={selected?.id ?? null}
        clusterName={selected?.name ?? null}
        onOpenChange={(open) => {
          if (!open) setSelected(null);
        }}
      />
    </div>
  );
}
