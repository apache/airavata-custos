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

import { Popover as PopoverPrimitive } from "@base-ui/react/popover";
import * as React from "react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { useAccessRequests, useDecideAccessRequest } from "../queries";
import type { AccessRequest, AccessRequestStatus } from "../schemas";
import { AccessRequestDrawer } from "./AccessRequestDrawer";

const APPROVE_UNDO_MS = 5000;

const STATUS_TABS = [
  { value: "pending", label: "Pending" },
  { value: "approved", label: "Approved" },
  { value: "denied", label: "Denied" },
  { value: "all", label: "All" },
] as const;

type StatusTab = (typeof STATUS_TABS)[number]["value"];

function serverStatusFor(tab: StatusTab): string | undefined {
  return tab === "all" ? undefined : tab.toUpperCase();
}

export function formatDate(iso: string): string {
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

export function statusBadgeFor(status: AccessRequestStatus) {
  if (status === "APPROVED") return <StatusBadge variant="approved" />;
  if (status === "DENIED") return <StatusBadge variant="rejected" label="Denied" />;
  return <StatusBadge variant="pending" />;
}

function DenyPopover({
  requestName,
  disabled,
  onDeny,
}: {
  requestName: string;
  disabled?: boolean;
  onDeny: (reason: string) => void;
}) {
  const [open, setOpen] = React.useState(false);
  const [reason, setReason] = React.useState("");
  return (
    <PopoverPrimitive.Root
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (next) setReason("");
      }}
    >
      <PopoverPrimitive.Trigger
        render={
          <Button variant="outline" size="sm" disabled={disabled}>
            Deny
          </Button>
        }
      />
      <PopoverPrimitive.Portal>
        <PopoverPrimitive.Positioner sideOffset={4} className="z-50 outline-none">
          <PopoverPrimitive.Popup
            className={cn(
              "z-50 w-72 rounded-lg bg-popover p-3 text-popover-foreground shadow-md ring-1 ring-foreground/10",
              "data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0",
            )}
          >
            <label
              className="mb-1 block text-xs font-medium text-muted-foreground"
              htmlFor={`deny-reason-${requestName}`}
            >
              Reason (optional)
            </label>
            <textarea
              id={`deny-reason-${requestName}`}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={3}
              className="w-full rounded-md border border-border bg-background px-2 py-1 text-sm text-foreground"
            />
            <div className="mt-2 flex justify-end gap-2">
              <Button variant="ghost" size="sm" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button
                variant="destructive"
                size="sm"
                onClick={() => {
                  setOpen(false);
                  onDeny(reason.trim());
                }}
              >
                Deny request
              </Button>
            </div>
          </PopoverPrimitive.Popup>
        </PopoverPrimitive.Positioner>
      </PopoverPrimitive.Portal>
    </PopoverPrimitive.Root>
  );
}

export function AccessRequestsQueue() {
  const searchParams = useShallowSearchParams();
  const statusParam = searchParams.get("status") ?? "pending";
  const statusTab: StatusTab = STATUS_TABS.some((t) => t.value === statusParam)
    ? (statusParam as StatusTab)
    : "pending";
  const search = searchParams.get("q") ?? "";
  const eventFilter = searchParams.get("event") ?? "all";

  function updateParam(key: string, value: string | null) {
    const params = new URLSearchParams(searchParams.toString());
    if (!value || value === "all") params.delete(key);
    else params.set(key, value);
    replaceShallowSearchParams(params);
  }

  const { data, isLoading, error, refetch } = useAccessRequests(
    { status: serverStatusFor(statusTab) },
    { refetchInterval: 20_000 },
  );
  const decide = useDecideAccessRequest();

  // Rows the admin just decided, shown in place until the next refetch swaps
  // the server data out from under them.
  const [optimistic, setOptimistic] = React.useState<Record<string, AccessRequest>>({});
  const [selected, setSelected] = React.useState<Set<string>>(new Set());
  const [drawerId, setDrawerId] = React.useState<string | null>(null);
  const [bulkProgress, setBulkProgress] = React.useState<string | null>(null);
  const undoTimers = React.useRef(new Map<string, ReturnType<typeof setTimeout>>());

  const rows = React.useMemo(() => {
    const base = (data ?? []).map((r) => optimistic[r.id] ?? r);
    const needle = search.trim().toLowerCase();
    return base.filter((r) => {
      if (needle) {
        const hay = `${r.name} ${r.email} ${r.institution}`.toLowerCase();
        if (!hay.includes(needle)) return false;
      }
      if (eventFilter !== "all" && r.event_code !== eventFilter) return false;
      return true;
    });
  }, [data, optimistic, search, eventFilter]);

  const eventCodes = React.useMemo(
    () => [...new Set((data ?? []).map((r) => r.event_code))].sort(),
    [data],
  );

  const pendingRows = rows.filter((r) => r.status === "PENDING");
  const allPendingSelected = pendingRows.length > 0 && pendingRows.every((r) => selected.has(r.id));

  function toggleAll() {
    setSelected(allPendingSelected ? new Set() : new Set(pendingRows.map((r) => r.id)));
  }

  function toggleOne(id: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function deselect(id: string) {
    setSelected((prev) => {
      if (!prev.has(id)) return prev;
      const next = new Set(prev);
      next.delete(id);
      return next;
    });
  }

  function revertOptimistic(id: string) {
    setOptimistic((prev) => {
      const next = { ...prev };
      delete next[id];
      return next;
    });
  }

  // Single-click approve is optimistic: the row flips to APPROVED at once and
  // the PUT fires only after the undo window closes. Closing the tab or
  // leaving the app inside the window drops the approval — accepted PoC risk.
  function approveWithUndo(row: AccessRequest) {
    setOptimistic((prev) => ({ ...prev, [row.id]: { ...row, status: "APPROVED" } }));
    deselect(row.id);
    const timer = setTimeout(() => {
      undoTimers.current.delete(row.id);
      decide.mutate(
        { id: row.id, body: { status: "APPROVED" } },
        {
          onError: (err) => {
            revertOptimistic(row.id);
            toast.error(`Approving ${row.name} failed: ${err.message}`);
          },
        },
      );
    }, APPROVE_UNDO_MS);
    undoTimers.current.set(row.id, timer);
    toast(`Approved ${row.name}`, {
      duration: APPROVE_UNDO_MS,
      action: {
        label: "Undo",
        onClick: () => {
          const pending = undoTimers.current.get(row.id);
          if (pending) {
            clearTimeout(pending);
            undoTimers.current.delete(row.id);
          }
          revertOptimistic(row.id);
        },
      },
    });
  }

  function deny(row: AccessRequest, reason: string) {
    decide.mutate(
      { id: row.id, body: { status: "DENIED", deny_reason: reason || undefined } },
      {
        onSuccess: (updated) => {
          setOptimistic((prev) => ({ ...prev, [row.id]: updated }));
          deselect(row.id);
        },
        onError: (err) => toast.error(`Denying ${row.name} failed: ${err.message}`),
      },
    );
  }

  // Bulk approve is immediate (no undo): N sequential PUTs with per-row errors.
  async function bulkApprove() {
    const targets = pendingRows.filter((r) => selected.has(r.id));
    for (const [i, row] of targets.entries()) {
      setBulkProgress(`Approving ${i + 1} of ${targets.length}…`);
      try {
        const updated = await decide.mutateAsync({ id: row.id, body: { status: "APPROVED" } });
        setOptimistic((prev) => ({ ...prev, [row.id]: updated }));
        deselect(row.id);
      } catch (err) {
        toast.error(
          `Approving ${row.name} failed: ${err instanceof Error ? err.message : String(err)}`,
        );
      }
    }
    setBulkProgress(null);
  }

  const columns: Array<DataTableColumn<AccessRequest>> = [
    {
      key: "select",
      header: (
        <input
          type="checkbox"
          aria-label="Select all pending"
          checked={allPendingSelected}
          onChange={toggleAll}
        />
      ),
      width: "32px",
      interactive: true,
      cell: (row) =>
        row.status === "PENDING" ? (
          <input
            type="checkbox"
            aria-label={`Select ${row.name}`}
            checked={selected.has(row.id)}
            onChange={() => toggleOne(row.id)}
            onClick={(e) => e.stopPropagation()}
          />
        ) : null,
    },
    {
      key: "name",
      header: "Name",
      cell: (row) => <span className="font-medium text-foreground">{row.name}</span>,
    },
    {
      key: "email",
      header: "Email",
      cell: (row) => <span className="text-muted-foreground">{row.email}</span>,
    },
    {
      key: "institution",
      header: "Institution",
      cell: (row) => <span className="text-muted-foreground">{row.institution}</span>,
    },
    {
      key: "event",
      header: "Event",
      cell: (row) => <span className="font-mono text-xs">{row.event_code}</span>,
    },
    {
      key: "submitted",
      header: "Submitted",
      sortable: true,
      sortValue: (row) => new Date(row.timestamp),
      cell: (row) => (
        <span className="text-xs tabular-nums text-muted-foreground">
          {formatDate(row.timestamp)}
        </span>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => statusBadgeFor(row.status),
    },
    {
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (row) =>
        row.status === "PENDING" ? (
          <span className="flex justify-end gap-2">
            <Button
              type="button"
              size="sm"
              disabled={bulkProgress !== null}
              onClick={(e) => {
                e.stopPropagation();
                approveWithUndo(row);
              }}
            >
              Approve
            </Button>
            <DenyPopover
              requestName={row.name}
              disabled={bulkProgress !== null}
              onDeny={(reason) => deny(row, reason)}
            />
          </span>
        ) : null,
    },
  ];

  const drawerRow = rows.find((r) => r.id === drawerId) ?? null;

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 rounded-md border bg-card p-4 sm:flex-row sm:items-center">
        <fieldset aria-label="Filter by status" className="flex gap-1">
          {STATUS_TABS.map((tab) => (
            <Button
              key={tab.value}
              variant={statusTab === tab.value ? "default" : "ghost"}
              size="sm"
              onClick={() => updateParam("status", tab.value === "pending" ? null : tab.value)}
            >
              {tab.label}
            </Button>
          ))}
        </fieldset>
        <Input
          type="search"
          placeholder="Search by name, email, or institution"
          value={search}
          onChange={(e) => updateParam("q", e.target.value)}
          aria-label="Search requests"
          className="sm:ml-2 sm:w-72"
        />
        <select
          value={eventFilter}
          onChange={(e) => updateParam("event", e.target.value)}
          aria-label="Filter by event"
          className="h-9 rounded-md border bg-background px-3 text-sm"
        >
          <option value="all">All events</option>
          {eventCodes.map((code) => (
            <option key={code} value={code}>
              {code}
            </option>
          ))}
        </select>
      </div>

      {selected.size > 0 ? (
        <div className="flex items-center justify-between rounded-md border bg-card p-4">
          <span className="text-sm text-muted-foreground">
            {bulkProgress ?? `${selected.size} selected`}
          </span>
          <Button disabled={bulkProgress !== null} onClick={bulkApprove}>
            Approve {selected.size} selected
          </Button>
        </div>
      ) : null}

      {error ? (
        <ErrorState message={error.message ?? "Failed to load requests"} onRetry={refetch} />
      ) : isLoading ? (
        <TableSkeleton />
      ) : (
        <DataTable
          columns={columns}
          rows={rows}
          rowKey={(row) => row.id}
          onRowClick={(row) => setDrawerId((prev) => (prev === row.id ? null : row.id))}
          rowClassName={(row) =>
            row.id === drawerId
              ? "bg-[color:var(--brand-tint)] hover:bg-[color:var(--brand-tint)]"
              : undefined
          }
          caption="Access requests"
          empty={
            <span className="text-sm text-muted-foreground">
              No access requests match the current filters.
            </span>
          }
        />
      )}

      <AccessRequestDrawer request={drawerRow} onClose={() => setDrawerId(null)} />
    </div>
  );
}
