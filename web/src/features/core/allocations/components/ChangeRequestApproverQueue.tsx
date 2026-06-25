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
import {
  StatusBadge,
  statusBadgeVariantFromChangeRequest,
} from "@/shared/ui/StatusBadge";
import {
  useApproveChangeRequest,
  useChangeRequests,
  useRejectChangeRequest,
} from "../queries";
import type { ChangeRequest, ChangeRequestStatus } from "../schemas";

export type ChangeRequestApproverQueueProps = {
  canApprove: boolean;
  approverId: string;
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

const STATUS_OPTIONS: Array<{ value: "all" | ChangeRequestStatus; label: string }> = [
  { value: "all", label: "All" },
  { value: "PENDING", label: "Pending" },
  { value: "APPROVED", label: "Approved" },
  { value: "REJECTED", label: "Rejected" },
];

export function ChangeRequestApproverQueue({
  canApprove,
  approverId,
}: ChangeRequestApproverQueueProps) {
  const [statusFilter, setStatusFilter] = React.useState<"all" | ChangeRequestStatus>("PENDING");
  const [search, setSearch] = React.useState("");
  const [pendingIds, setPendingIds] = React.useState<Set<string>>(new Set());

  const query = useChangeRequests(
    statusFilter === "all" ? {} : { status: statusFilter },
  );
  const approveMutation = useApproveChangeRequest();
  const rejectMutation = useRejectChangeRequest();

  function markPending(id: string, on: boolean) {
    setPendingIds((prev) => {
      const next = new Set(prev);
      if (on) next.add(id);
      else next.delete(id);
      return next;
    });
  }

  async function handleApprove(row: ChangeRequest) {
    markPending(row.id, true);
    try {
      await approveMutation.mutateAsync({ id: row.id, approverId });
      toast.success(`Approved ${row.id}`);
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      markPending(row.id, false);
    }
  }

  async function handleReject(row: ChangeRequest) {
    markPending(row.id, true);
    try {
      await rejectMutation.mutateAsync({ id: row.id, approverId });
      toast.success(`Rejected ${row.id}`);
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      markPending(row.id, false);
    }
  }

  const rows = (query.data ?? []).filter((r) => {
    if (!search) return true;
    const needle = search.toLowerCase();
    return (
      r.id.toLowerCase().includes(needle) ||
      r.compute_allocation_id.toLowerCase().includes(needle) ||
      r.requester_id.toLowerCase().includes(needle) ||
      r.reason.toLowerCase().includes(needle)
    );
  });

  const columns: Array<DataTableColumn<ChangeRequest>> = [
    {
      key: "submitted",
      header: "Submitted",
      sortable: true,
      sortValue: (row) => new Date(row.timestamp),
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{formatDate(row.timestamp)}</span>
      ),
    },
    {
      key: "allocation",
      header: "Allocation",
      cell: (row) => (
        <Link
          href={`/allocations/${row.compute_allocation_id}`}
          className="text-sm font-medium text-foreground hover:underline"
        >
          {row.compute_allocation_id}
        </Link>
      ),
    },
    {
      key: "requester",
      header: "Requester",
      cell: (row) => <span className="font-mono text-xs">{row.requester_id}</span>,
    },
    {
      key: "amount",
      header: "Requested SUs",
      align: "right",
      sortable: true,
      sortValue: (row) => row.requested_su_amount,
      cell: (row) => (
        <span className="tabular-nums">{new Intl.NumberFormat().format(row.requested_su_amount)}</span>
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
        const pending = pendingIds.has(row.id);
        if (row.change_status !== "PENDING" || !canApprove) {
          return (
            <Button variant="ghost" size="sm" render={<Link href={`/change-requests/${row.id}`} />}>
              View
            </Button>
          );
        }
        return (
          <div className="flex justify-end gap-1.5">
            <Button
              variant="outline"
              size="sm"
              disabled={pending}
              onClick={() => handleReject(row)}
              aria-label={`Reject ${row.id}`}
            >
              Reject
            </Button>
            <Button
              size="sm"
              disabled={pending}
              onClick={() => handleApprove(row)}
              aria-label={`Approve ${row.id}`}
            >
              Approve
            </Button>
          </div>
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
            onChange={(e) => setStatusFilter(e.target.value as "all" | ChangeRequestStatus)}
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
            placeholder="Requester, allocation, reason…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="sm:w-72"
          />
        </div>
      </div>

      {query.isLoading ? (
        <TableSkeleton rows={5} columns={6} />
      ) : query.error ? (
        <ErrorState
          message={(query.error as Error).message}
          onRetry={() => query.refetch()}
        />
      ) : rows.length === 0 ? (
        <EmptyState
          heading="No change requests"
          description="No requests match the current filters."
        />
      ) : (
        <DataTable columns={columns} rows={rows} rowKey={(row) => row.id} />
      )}
    </div>
  );
}
