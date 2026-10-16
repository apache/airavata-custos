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
import type { ComputeAllocationChangeRequest } from "@/generated/core/types.gen";
import { formatDate, formatDateTime, formatSU } from "@/shared/format";
import { useUrlFilteredPage } from "@/shared/hooks/usePageClamp";
import { useCanonicalParams } from "@/shared/hooks/useShallowSearchParams";
import { ROUTE_LABELS } from "@/shared/layout/nav";
import { DataTable, type DataTableColumn, PastLastPage } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { statusLabel } from "@/shared/ui/FormDialog";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge, statusBadgeVariantFromChangeRequest } from "@/shared/ui/StatusBadge";
import { Button } from "@/shared/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLinkItem,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";
import { keepFocus } from "@/shared/ui/focus";
import { Label } from "@/shared/ui/label";
import { Link } from "@/shared/ui/link";
import { toastOnSuccess } from "@/shared/ui/sonner";
import * as React from "react";
import { CHANGE_REQUEST_LIMIT, useChangeRequests, useDecideChangeRequest } from "../queries";
import { AllocationName } from "./AllocationName";
import { MemberName } from "./MemberName";

export type ChangeRequestApproverQueueProps = {
  canApprove: boolean;
  // Undefined until the session loads; decisions wait for it.
  approverId: string | undefined;
};

const PAGE_SIZE = 50;

const STATUS_OPTIONS = [
  { value: "all", label: "All" },
  { value: "PENDING", label: "Pending" },
  { value: "APPROVED", label: "Approved" },
  { value: "REJECTED", label: "Rejected" },
] as const;

export const decisionToast = (status: "APPROVED" | "REJECTED") =>
  `${status === "APPROVED" ? "Approved" : "Rejected"} the change request`;

// The requested change and its decision, shared with the allocation's Change requests tab.
export const requestedChangeColumns: Array<DataTableColumn<ComputeAllocationChangeRequest>> = [
  {
    key: "amount",
    header: "Requested SUs",
    align: "right",
    sortable: true,
    sortValue: (row) => row.requested_su_amount,
    cell: (row) => <span className="tabular-nums">{formatSU(row.requested_su_amount)}</span>,
  },
  {
    key: "requestedStatus",
    header: "Requested status",
    cell: (row) => (
      <span className="text-sm text-muted-foreground">{statusLabel(row.requested_status)}</span>
    ),
  },
  {
    key: "status",
    header: "Status",
    cell: (row) => (
      <StatusBadge
        variant={statusBadgeVariantFromChangeRequest(row.change_status)}
        label={statusLabel(row.change_status)}
      />
    ),
  },
];

export function ChangeRequestApproverQueue({
  canApprove,
  approverId,
}: ChangeRequestApproverQueueProps) {
  const { params, page, setPage, offset, setFilter } = useUrlFilteredPage(PAGE_SIZE);
  // No ?status= means the pending queue.
  const statusFilter =
    STATUS_OPTIONS.find((o) => o.value === params.get("status"))?.value ?? "PENDING";
  const [pendingIds, setPendingIds] = React.useState<Set<string>>(new Set());
  useCanonicalParams({ status: statusFilter === "PENDING" ? undefined : statusFilter });

  const query = useChangeRequests(statusFilter === "all" ? undefined : statusFilter);
  const decideMutation = useDecideChangeRequest();
  const queue = React.useRef<HTMLElement>(null);

  function markPending(id: string, on: boolean) {
    setPendingIds((prev) => {
      const next = new Set(prev);
      if (on) next.add(id);
      else next.delete(id);
      return next;
    });
  }

  function decide(row: ComputeAllocationChangeRequest, changeStatus: "APPROVED" | "REJECTED") {
    const id = row.id;
    if (!approverId) return;
    // A decided row can leave the filter, taking its control (or the menu's trigger) with it.
    keepFocus(() => queue.current);
    markPending(id, true);
    toastOnSuccess(
      decideMutation
        .mutateAsync({ id, change_status: changeStatus, approver_id: approverId })
        .finally(() => markPending(id, false)),
      decisionToast(changeStatus),
    );
  }

  const all = query.data ?? [];
  const rows = all.slice(offset, offset + PAGE_SIZE);

  const columns: Array<DataTableColumn<ComputeAllocationChangeRequest>> = [
    {
      key: "submitted",
      header: "Submitted",
      sortable: true,
      sortValue: (row) => new Date(row.timestamp),
      cell: (row) => (
        <Timestamp iso={row.timestamp} className="text-sm text-muted-foreground">
          {formatDate(row.timestamp)}
        </Timestamp>
      ),
    },
    {
      key: "allocation",
      header: "Allocation",
      cell: (row) => (
        <AllocationName
          id={row.compute_allocation_id}
          className="block max-w-[24ch] text-sm font-medium text-foreground hover:underline"
        />
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
      sticky: true,
      cell: (row) => {
        const disabled = pendingIds.has(row.id) || !approverId;
        const name = requestSummary(row);
        const view = <ViewRequest id={row.id} name={name} />;
        if (row.change_status !== "PENDING" || !canApprove) return view;
        return (
          <>
            <span className="hidden items-center gap-2 sm:inline-flex">
              <DecisionButtons
                name={name}
                disabled={disabled}
                onDecide={(status) => decide(row, status)}
              />
              {view}
            </span>
            <DecideMenu
              name={name}
              href={`/change-requests/${row.id}`}
              disabled={disabled}
              onDecide={(status) => decide(row, status)}
            />
          </>
        );
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-display text-[28px] font-bold leading-tight">
          {ROUTE_LABELS["/change-requests"]}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          The newest {CHANGE_REQUEST_LIMIT} change requests across every allocation, filtered by
          status. Holders of allocation write access approve or reject pending requests.
        </p>
      </div>

      <div className="flex flex-wrap items-end gap-3 rounded-md border bg-card p-4">
        <div className="space-y-1">
          <Label htmlFor="cr-status-filter">Status</Label>
          <select
            id="cr-status-filter"
            value={statusFilter}
            onChange={(e) =>
              setFilter("status", e.target.value === "PENDING" ? null : e.target.value)
            }
            className="h-9 rounded-md border bg-background px-3 text-sm"
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      <section
        ref={queue}
        aria-label="Change request queue"
        tabIndex={-1}
        className="rounded-md outline-none focus-visible:ring-2 focus-visible:ring-ring"
      >
        {query.isLoading ? (
          <TableSkeleton rows={5} columns={6} />
        ) : query.error ? (
          <ErrorState onRetry={() => query.refetch()} />
        ) : rows.length === 0 && page > 1 ? (
          <PastLastPage things="change requests" page={page} onPageChange={setPage} />
        ) : rows.length === 0 ? (
          <EmptyState
            heading="No change requests"
            description="No requests match the current filter."
          />
        ) : (
          <DataTable
            caption="Change requests"
            columns={columns}
            rows={rows}
            rowKey={(row) => row.id}
            pagination={{ page, pageSize: PAGE_SIZE, total: all.length, onPageChange: setPage }}
          />
        )}
      </section>
    </div>
  );
}

// Tells apart requests in the row controls' labels; the short id separates any two left alike.
export function requestSummary(row: ComputeAllocationChangeRequest): string {
  return [
    `${formatSU(row.requested_su_amount)} SU`,
    statusLabel(row.requested_status),
    `submitted ${formatDateTime(row.timestamp)}`,
    `request ${row.id.slice(0, 8)}`,
  ].join(", ");
}

export const ViewRequest = ({ id, name }: { id: string; name: string }) => (
  <Button
    variant="ghost"
    size="sm"
    nativeButton={false}
    render={<Link href={`/change-requests/${id}`} />}
    aria-label={`View change request for ${name}`}
  >
    View
  </Button>
);

// Below sm the three actions fold into one menu, so the row keeps room for what is decided.
function DecideMenu({
  name,
  href,
  disabled,
  onDecide,
}: {
  name: string;
  href: string;
  disabled: boolean;
  onDecide: (status: "APPROVED" | "REJECTED") => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            variant="outline"
            size="sm"
            className="sm:hidden"
            disabled={disabled}
            focusableWhenDisabled
            aria-label={`Decide change request for ${name}`}
          />
        }
      >
        Decide
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => onDecide("APPROVED")}>Approve</DropdownMenuItem>
        <DropdownMenuItem onClick={() => onDecide("REJECTED")}>Reject</DropdownMenuItem>
        <DropdownMenuLinkItem href={href} closeOnClick>
          View
        </DropdownMenuLinkItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function DecisionButtons({
  name,
  disabled,
  onDecide,
}: {
  name: string;
  disabled: boolean;
  onDecide: (status: "APPROVED" | "REJECTED") => void;
}) {
  return (
    <div className="flex justify-end gap-1.5">
      <Button
        variant="outline"
        size="sm"
        disabled={disabled}
        focusableWhenDisabled
        onClick={() => onDecide("REJECTED")}
        aria-label={`Reject change request for ${name}`}
      >
        Reject
      </Button>
      <Button
        size="sm"
        disabled={disabled}
        focusableWhenDisabled
        onClick={() => onDecide("APPROVED")}
        aria-label={`Approve change request for ${name}`}
      >
        Approve
      </Button>
    </div>
  );
}
