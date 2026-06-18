"use client";

import Link from "next/link";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import {
  StatusBadge,
  statusBadgeVariantFromAllocationStatus,
} from "@/shared/ui/StatusBadge";
// Sanctioned cross-feature import per ADR-0004.
import { useAllocationsByProject } from "@/features/core/allocations/queries";

export type ProjectAllocationsTabProps = {
  projectId: string;
};

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  } catch {
    return iso;
  }
}

export function ProjectAllocationsTab({ projectId }: ProjectAllocationsTabProps) {
  const query = useAllocationsByProject(projectId);

  if (query.isLoading) return <TableSkeleton rows={3} columns={4} />;
  if (query.error) {
    return <ErrorState message={(query.error as Error).message} onRetry={() => query.refetch()} />;
  }
  const rows = query.data?.items ?? [];
  if (rows.length === 0) {
    return (
      <EmptyState
        heading="No allocations on this project"
        description="Allocations granted to this project will appear here."
      />
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
      cell: (row) => <span className="text-sm">{row.compute_cluster_id}</span>,
    },
    {
      key: "initial",
      header: "Initial SUs",
      align: "right",
      cell: (row) => (
        <span className="tabular-nums">{new Intl.NumberFormat().format(row.initial_su_amount)}</span>
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

  return <DataTable columns={columns} rows={rows} rowKey={(row) => row.id} />;
}
