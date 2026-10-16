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

import { AllocationStatusLabel } from "@/features/core/allocations/components/AllocationStatusLabel";
// Sanctioned cross-feature import per ADR-0004.
import { CreateAllocationDialog } from "@/features/core/allocations/components/CreateAllocationDialog";
import { useAllocationsByProject } from "@/features/core/allocations/queries";
import { useClusters } from "@/features/core/clusters/queries";
import { formatDay, formatSU } from "@/shared/format";
import { DataTable, type DataTableColumn, NameCell } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";

export type ProjectAllocationsTabProps = {
  projectId: string;
  canCreate: boolean;
  // Without Allocation read the list holds only the caller's own allocations.
  canReadAll: boolean;
  canReadClusters: boolean;
};

export function ProjectAllocationsTab({
  projectId,
  canCreate,
  canReadAll,
  canReadClusters,
}: ProjectAllocationsTabProps) {
  const query = useAllocationsByProject(projectId);
  const clusters = useClusters(canReadClusters);

  if (query.isLoading) return <TableSkeleton rows={3} columns={4} />;
  if (query.error) {
    return <ErrorState error={query.error} onRetry={() => query.refetch()} />;
  }
  const rows = query.data ?? [];
  const cta = canCreate ? (
    <div className="flex justify-end">
      <CreateAllocationDialog projectId={projectId} />
    </div>
  ) : null;
  if (rows.length === 0) {
    return (
      <div className="space-y-3">
        {cta}
        <EmptyState
          heading={
            canReadAll
              ? "No allocations on this project"
              : "No allocations you can see on this project"
          }
          description={
            canReadAll
              ? "Allocations granted to this project will appear here."
              : "Allocations of this project you belong to will appear here."
          }
        />
      </div>
    );
  }

  type Row = (typeof rows)[number];
  const columns: Array<DataTableColumn<Row>> = [
    {
      key: "name",
      header: "Allocation",
      cell: (row) => (
        <NameCell
          href={`/allocations/${row.id}`}
          text={row.name}
          className="font-medium text-foreground hover:underline"
        />
      ),
    },
    ...(canReadClusters
      ? [
          {
            key: "cluster",
            header: "Cluster",
            cell: (row: Row) => (
              <NameCell
                text={
                  clusters.data?.find((c) => c.id === row.compute_cluster_id)?.name ??
                  (clusters.isSuccess ? "Unknown cluster" : "—")
                }
              />
            ),
          },
        ]
      : []),
    {
      key: "initial",
      header: "Initial SUs",
      align: "right",
      cell: (row) => <span className="tabular-nums">{formatSU(row.initial_su_amount)}</span>,
    },
    {
      key: "endDate",
      header: "End date",
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{formatDay(row.end_time)}</span>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <AllocationStatusLabel
          status={row.status}
          startTime={row.start_time}
          endTime={row.end_time}
        />
      ),
    },
  ];

  return (
    <div className="space-y-3">
      {cta}
      <DataTable caption="Allocations" columns={columns} rows={rows} rowKey={(row) => row.id} />
    </div>
  );
}
