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

import { Timestamp } from "@/features/core/audit/components/primitives/Timestamp";
import type { ComputeAllocation, ComputeAllocationChangeRequest } from "@/generated/core/types.gen";
import { formatDate, pluralize } from "@/shared/format";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import * as React from "react";
import { useAllocationChangeRequests } from "../queries";
import { ViewRequest, requestSummary, requestedChangeColumns } from "./ChangeRequestApproverQueue";
import { ChangeRequestSubmitDrawer } from "./ChangeRequestSubmitDrawer";
import { MemberName } from "./MemberName";

export type AllocationChangeRequestsTabProps = {
  allocation: ComputeAllocation;
  canSubmit: boolean;
};

export function AllocationChangeRequestsTab({
  allocation,
  canSubmit,
}: AllocationChangeRequestsTabProps) {
  const [drawerOpen, setDrawerOpen] = React.useState(false);
  const query = useAllocationChangeRequests(allocation.id);

  if (query.isLoading) return <TableSkeleton rows={3} columns={4} />;
  if (query.error) {
    return <ErrorState onRetry={() => query.refetch()} />;
  }
  const rows = [...(query.data ?? [])].sort(
    (a, b) => Date.parse(b.timestamp) - Date.parse(a.timestamp),
  );

  const columns: Array<DataTableColumn<ComputeAllocationChangeRequest>> = [
    {
      key: "submitted",
      header: "Submitted",
      cell: (row) => (
        <Timestamp iso={row.timestamp} className="text-sm text-muted-foreground">
          {formatDate(row.timestamp)}
        </Timestamp>
      ),
    },
    {
      key: "requester",
      header: "Requester",
      cell: (row) => <MemberName id={row.requester_id} allocationId={row.compute_allocation_id} />,
    },
    ...requestedChangeColumns,
    {
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (row) => <ViewRequest id={row.id} name={requestSummary(row)} />,
    },
  ];

  const headerCta = canSubmit ? (
    <Button size="sm" onClick={() => setDrawerOpen(true)}>
      + Submit change request
    </Button>
  ) : null;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {rows.length} {pluralize("request", rows.length)}
        </p>
        {headerCta}
      </div>
      {rows.length === 0 ? (
        <EmptyState
          heading="No change requests"
          description={
            canSubmit
              ? "Submit a request when this allocation needs a new SU total or a status change."
              : "There are no change requests for this allocation."
          }
        />
      ) : (
        <DataTable
          caption="Change requests"
          columns={columns}
          rows={rows}
          rowKey={(row) => row.id}
        />
      )}
      <ChangeRequestSubmitDrawer
        open={drawerOpen}
        onOpenChange={setDrawerOpen}
        allocation={allocation}
      />
    </div>
  );
}
