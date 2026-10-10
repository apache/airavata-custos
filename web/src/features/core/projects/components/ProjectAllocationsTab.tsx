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

import { formatDate, formatNumber } from "@/shared/format";
// Sanctioned cross-feature import per ADR-0004.
import { CreateAllocationDialog } from "@/features/core/allocations/components/CreateAllocationDialog";
import { useAllocationsByProject } from "@/features/core/allocations/queries";
import { useClusterName } from "@/features/core/clusters/queries";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge, statusBadgeVariantFromAllocationStatus } from "@/shared/ui/StatusBadge";
import Link from "next/link";

export type ProjectAllocationsTabProps = {
  projectId: string;
  canCreate: boolean;
};

export function ProjectAllocationsTab({ projectId, canCreate }: ProjectAllocationsTabProps) {
  const query = useAllocationsByProject(projectId);
  const clusterName = useClusterName();

  if (query.isLoading) return <TableSkeleton rows={3} columns={4} />;
  if (query.error) {
    return <ErrorState message={query.error.message} onRetry={() => query.refetch()} />;
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
          heading="No allocations on this project"
          description="Allocations granted to this project will appear here."
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
        <Link
          href={`/allocations/${row.id}`}
          className="font-medium text-foreground hover:underline"
        >
          {row.name}
        </Link>
      ),
    },
    {
      key: "cluster",
      header: "Cluster",
      cell: (row) => <span className="text-sm">{clusterName(row.compute_cluster_id)}</span>,
    },
    {
      key: "initial",
      header: "Initial SUs",
      align: "right",
      cell: (row) => (
        <span className="tabular-nums">
          {formatNumber(row.initial_su_amount)}
        </span>
      ),
    },
    {
      key: "endDate",
      header: "End date",
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{formatDate(row.end_time)}</span>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <StatusBadge
          variant={statusBadgeVariantFromAllocationStatus(row.status)}
          label={row.status}
        />
      ),
    },
  ];

  return (
    <div className="space-y-3">
      {cta}
      <DataTable columns={columns} rows={rows} rowKey={(row) => row.id ?? ""} />
    </div>
  );
}
