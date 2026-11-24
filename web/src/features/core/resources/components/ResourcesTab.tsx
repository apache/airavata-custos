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

import { useClusters } from "@/features/core/clusters/queries";
import type { ComputeAllocationResourceSummary } from "@/generated/core/zod.gen";
import { requires, usePermissions } from "@/shared/casl/permissions";
import { ConfirmDialog } from "@/shared/ui/ConfirmDialog";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { ErrorState } from "@/shared/ui/ErrorState";
import { GuardedButton } from "@/shared/ui/GuardedButton";
import { CardSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import * as React from "react";
import { useDeleteResource, useEffectiveRate, useResourceSummaries } from "../queries";
import { ResourceDialog } from "./ResourceDialog";
import { ResourcesRatesDrawer } from "./ResourcesRatesDrawer";

// Rates are SU per billing unit-hour: usage is units × hours × rate.
export function EffectiveRateCell({ resource }: { resource: ComputeAllocationResourceSummary }) {
  const rate = useEffectiveRate(resource.rate_count ? resource.id : undefined);
  if (rate.isLoading) return <span className="text-sm text-muted-foreground">…</span>;
  if (!rate.data) return <span className="text-sm text-muted-foreground">No rate</span>;
  return (
    <span className="text-sm tabular-nums text-foreground">
      {rate.data.rate.toLocaleString()} SU/unit·h
    </span>
  );
}

export function ResourcesTab() {
  const can = usePermissions();
  const query = useResourceSummaries();
  const clusters = useClusters().data ?? [];
  const [search, setSearch] = React.useState("");
  const [editing, setEditing] = React.useState<ComputeAllocationResourceSummary | "new" | null>(
    null,
  );
  const [ratesFor, setRatesFor] = React.useState<ComputeAllocationResourceSummary | null>(null);
  const [deleting, setDeleting] = React.useState<ComputeAllocationResourceSummary | null>(null);
  const remove = useDeleteResource();

  const rows = query.data ?? [];
  const filtered = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    return rows.filter((row) => `${row.name} ${row.resource_type}`.toLowerCase().includes(needle));
  }, [rows, search]);

  const columns: Array<DataTableColumn<ComputeAllocationResourceSummary>> = [
    {
      key: "name",
      header: "Resource",
      cell: (row) => (
        <div>
          <div className="font-medium text-foreground">{row.name}</div>
          <div className="text-xs text-muted-foreground">{row.resource_type}</div>
        </div>
      ),
    },
    {
      key: "rate",
      header: "SU rate",
      align: "right",
      cell: (row) => <EffectiveRateCell resource={row} />,
    },
    {
      key: "allocation_count",
      header: "Allocations",
      align: "right",
      cell: (row) => (
        <span className="text-sm tabular-nums text-foreground">{row.allocation_count}</span>
      ),
    },
    {
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (row) => (
        <div className="flex justify-end gap-2">
          <GuardedButton
            variant="outline"
            size="sm"
            allowed={can.allocationsWrite}
            reason={requires("allocationsWrite")}
            onClick={() => setEditing(row)}
          >
            Edit
          </GuardedButton>
          <Button variant="outline" size="sm" onClick={() => setRatesFor(row)}>
            Rates ({row.rate_count})
          </Button>
          <GuardedButton
            variant="destructive"
            size="sm"
            allowed={can.allocationsWrite}
            reason={requires("allocationsWrite")}
            onClick={() => {
              remove.reset();
              setDeleting(row);
            }}
          >
            Delete
          </GuardedButton>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        <Input
          type="search"
          placeholder="Search resources"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Search resources"
          className="sm:w-72"
        />
        <GuardedButton
          allowed={can.allocationsWrite}
          reason={requires("allocationsWrite")}
          onClick={() => setEditing("new")}
        >
          + Add resource
        </GuardedButton>
      </div>

      {query.isLoading ? (
        <CardSkeleton />
      ) : query.error ? (
        <ErrorState message={query.error.message} onRetry={() => query.refetch()} />
      ) : (
        <DataTable
          columns={columns}
          rows={filtered}
          rowKey={(row) => row.id}
          empty={
            rows.length === 0 ? "No resources registered." : "No resources match these filters."
          }
        />
      )}

      <ResourceDialog
        open={editing !== null}
        onOpenChange={(open) => {
          if (!open) setEditing(null);
        }}
        clusters={clusters}
        resource={editing === "new" ? null : editing}
      />
      <ConfirmDialog
        open={deleting !== null}
        onOpenChange={(open) => {
          if (!open) setDeleting(null);
        }}
        title={`Delete ${deleting?.name ?? "resource"}?`}
        confirmLabel="Delete"
        pending={remove.isPending}
        error={remove.error}
        onConfirm={() => {
          if (deleting) remove.mutate(deleting.id, { onSuccess: () => setDeleting(null) });
        }}
      >
        Its rates go with it, and allocations lose access to it.
      </ConfirmDialog>
      <ResourcesRatesDrawer
        resourceId={ratesFor?.id ?? null}
        resourceName={ratesFor?.name ?? null}
        onOpenChange={(open) => {
          if (!open) setRatesFor(null);
        }}
      />
    </div>
  );
}
