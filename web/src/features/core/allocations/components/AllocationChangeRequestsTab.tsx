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

import { UserName } from "@/features/core/users/components/UserPicker";
import { formatDate, formatNumber } from "@/shared/format";
import Link from "next/link";
import * as React from "react";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge, statusBadgeVariantFromChangeRequest } from "@/shared/ui/StatusBadge";
import { useAllocationChangeRequests, useCurrentSuAmount } from "../queries";
import type { ComputeAllocation, ComputeAllocationChangeRequest } from "@/generated/core/types.gen";
import { ChangeRequestSubmitDrawer } from "./ChangeRequestSubmitDrawer";

export type AllocationChangeRequestsTabProps = {
  allocation: ComputeAllocation;
  canSubmit: boolean;
  requesterId: string;
};

export function AllocationChangeRequestsTab({
  allocation,
  canSubmit,
  requesterId,
}: AllocationChangeRequestsTabProps) {
  const [drawerOpen, setDrawerOpen] = React.useState(false);
  const query = useAllocationChangeRequests(allocation.id);
  const currentSu = useCurrentSuAmount(allocation);

  if (query.isLoading) return <TableSkeleton rows={3} columns={4} />;
  if (query.error) {
    return <ErrorState message={query.error.message} onRetry={() => query.refetch()} />;
  }
  const rows = query.data ?? [];

  const columns: Array<DataTableColumn<ComputeAllocationChangeRequest>> = [
    {
      key: "submitted",
      header: "Submitted",
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{formatDate(row.timestamp)}</span>
      ),
    },
    {
      key: "requester",
      header: "Requester",
      cell: (row) => <UserName id={row.requester_id} />,
    },
    {
      key: "amount",
      header: "Requested SUs",
      align: "right",
      cell: (row) => (
        <span className="tabular-nums">
          {formatNumber(row.requested_su_amount)}
        </span>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <StatusBadge
          variant={statusBadgeVariantFromChangeRequest(row.change_status)}
          label={row.change_status}
        />
      ),
    },
    {
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (row) => (
        <Button variant="ghost" size="sm" nativeButton={false} render={<Link href={`/change-requests/${row.id}`} />}>
          View
        </Button>
      ),
    },
  ];

  const headerCta = canSubmit ? (
    <Button size="sm" disabled={currentSu.data === undefined} onClick={() => setDrawerOpen(true)}>
      + Submit change request
    </Button>
  ) : null;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {rows.length} request{rows.length === 1 ? "" : "s"}
        </p>
        {headerCta}
      </div>
      {rows.length === 0 ? (
        <EmptyState
          heading="No change requests"
          description={
            canSubmit
              ? "Submit a request when this allocation needs more SUs, more time, or a status change."
              : "There are no change requests for this allocation."
          }
        />
      ) : (
        <DataTable columns={columns} rows={rows} rowKey={(row) => row.id ?? ""} />
      )}
      <ChangeRequestSubmitDrawer
        open={drawerOpen}
        onOpenChange={setDrawerOpen}
        allocationId={allocation.id ?? ""}
        requesterId={requesterId}
        currentSuAmount={currentSu.data ?? 0}
        currentStatus={allocation.status}
      />
    </div>
  );
}
