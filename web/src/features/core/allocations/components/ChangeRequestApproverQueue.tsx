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
import { AllocationName } from "./AllocationName";
import { formatDate, formatNumber } from "@/shared/format";
import Link from "next/link";
import * as React from "react";
import { toast } from "sonner";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge, statusBadgeVariantFromChangeRequest } from "@/shared/ui/StatusBadge";
import type { ComputeAllocationChangeRequest } from "@/generated/core/types.gen";
import { useChangeRequests, useCustosManaged, useDecideChangeRequest } from "../queries";

export type ChangeRequestApproverQueueProps = {
  canApprove: boolean;
  // Undefined until /me loads; decisions wait for it.
  approverId: string | undefined;
};

// The backend's ceiling; the endpoint has no offset to page past it.
const LIMIT = 200;

const STATUS_OPTIONS = [
  { value: "all", label: "All" },
  { value: "PENDING", label: "Pending" },
  { value: "APPROVED", label: "Approved" },
  { value: "REJECTED", label: "Rejected" },
] as const;

export function ChangeRequestApproverQueue({
  canApprove,
  approverId,
}: ChangeRequestApproverQueueProps) {
  const [statusFilter, setStatusFilter] = React.useState("PENDING");
  const [search, setSearch] = React.useState("");
  const [pendingIds, setPendingIds] = React.useState<Set<string>>(new Set());

  const query = useChangeRequests({
    status: statusFilter === "all" ? undefined : statusFilter,
    limit: LIMIT,
  });
  const decideMutation = useDecideChangeRequest();

  function markPending(id: string, on: boolean) {
    setPendingIds((prev) => {
      const next = new Set(prev);
      if (on) next.add(id);
      else next.delete(id);
      return next;
    });
  }

  function decide(row: ComputeAllocationChangeRequest, changeStatus: "APPROVED" | "REJECTED") {
    const id = row.id ?? "";
    markPending(id, true);
    decideMutation.mutate(
      { path: { id }, body: { ...row, change_status: changeStatus, approver_id: approverId } },
      {
        onSuccess: () =>
          toast.success(`${changeStatus === "APPROVED" ? "Approved" : "Rejected"} ${id}`),
        onSettled: () => markPending(id, false),
      },
    );
  }

  const rows = (query.data ?? []).filter((r) => {
    if (!search) return true;
    const needle = search.toLowerCase();
    return [r.id, r.compute_allocation_id, r.requester_id, r.reason].some((v) =>
      v?.toLowerCase().includes(needle),
    );
  });

  const columns: Array<DataTableColumn<ComputeAllocationChangeRequest>> = [
    {
      key: "submitted",
      header: "Submitted",
      sortable: true,
      sortValue: (row) => new Date(row.timestamp ?? ""),
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{formatDate(row.timestamp)}</span>
      ),
    },
    {
      key: "allocation",
      header: "Allocation",
      cell: (row) => (
        <AllocationName
          id={row.compute_allocation_id}
          className="text-sm font-medium text-foreground hover:underline"
        />
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
      sortable: true,
      sortValue: (row) => row.requested_su_amount,
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
      cell: (row) => {
        const disabled = pendingIds.has(row.id ?? "") || !approverId;
        const view = (
          <Button variant="ghost" size="sm" nativeButton={false} render={<Link href={`/change-requests/${row.id}`} />}>
            View
          </Button>
        );
        if (row.change_status !== "PENDING" || !canApprove) return view;
        return (
          <CustosManaged allocationId={row.compute_allocation_id} fallback={view}>
            <div className="flex justify-end gap-1.5">
              <Button
                variant="outline"
                size="sm"
                disabled={disabled}
                onClick={() => decide(row, "REJECTED")}
                aria-label={`Reject ${row.id}`}
              >
                Reject
              </Button>
              <Button
                size="sm"
                disabled={disabled}
                onClick={() => decide(row, "APPROVED")}
                aria-label={`Approve ${row.id}`}
              >
                Approve
              </Button>
            </div>
          </CustosManaged>
        );
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-display text-[28px] font-bold leading-tight">Change requests</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Review allocation change requests from PIs and project managers.
        </p>
      </div>

      <div className="flex flex-wrap items-end gap-3 rounded-md border bg-card p-4">
        <div className="space-y-1">
          <Label htmlFor="cr-status-filter">Status</Label>
          <select
            id="cr-status-filter"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="h-9 rounded-md border bg-background px-3 text-sm"
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-1">
          <Label htmlFor="cr-search">Search</Label>
          <Input
            id="cr-search"
            type="search"
            placeholder="Request, allocation or requester ID, reason…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="sm:w-72"
          />
        </div>
      </div>

      {query.isLoading ? (
        <TableSkeleton rows={5} columns={6} />
      ) : query.error ? (
        <ErrorState message={query.error.message} onRetry={() => query.refetch()} />
      ) : rows.length === 0 ? (
        <EmptyState
          heading="No change requests"
          description="No requests match the current filters."
        />
      ) : (
        <>
          <DataTable columns={columns} rows={rows} rowKey={(row) => row.id ?? ""} />
          {(query.data?.length ?? 0) >= LIMIT ? (
            <p className="text-xs text-muted-foreground">
              Showing the {LIMIT} most recent requests; narrow the status filter to see older ones.
            </p>
          ) : null}
        </>
      )}
    </div>
  );
}

// Rows carry only the allocation id, so origination resolves per row through cached queries.
function CustosManaged({
  allocationId,
  fallback,
  children,
}: {
  allocationId: string | undefined;
  fallback: React.ReactNode;
  children: React.ReactNode;
}) {
  return useCustosManaged(allocationId) ? children : fallback;
}
