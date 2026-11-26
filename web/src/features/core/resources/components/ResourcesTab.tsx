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

import type { Partition } from "@/features/connectors/slurm/api";
import { formatMinutes, type Joined, joinByName } from "@/features/connectors/slurm/format";
import { useSlurmCluster } from "@/features/connectors/slurm/queries";
import type { ComputeAllocationResourceSummary } from "@/generated/core/zod.gen";
import { requires, usePermissions } from "@/shared/casl/permissions";
import { ConfirmDialog } from "@/shared/ui/ConfirmDialog";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { ErrorState } from "@/shared/ui/ErrorState";
import { GuardedButton } from "@/shared/ui/GuardedButton";
import { CardSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
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

// A partition's billing record in Custos joined to its SLURM partition; either may be missing.
type Row = Joined<ComputeAllocationResourceSummary, Partition>;

export function ResourcesTab({ clusterId }: { clusterId: string }) {
  const can = usePermissions();
  const partitions = useSlurmCluster(clusterId, (v) => v.partitions ?? []);
  const query = useResourceSummaries();
  const [editing, setEditing] = React.useState<Row | null>(null);
  const [ratesFor, setRatesFor] = React.useState<ComputeAllocationResourceSummary | null>(null);
  const [deleting, setDeleting] = React.useState<ComputeAllocationResourceSummary | null>(null);
  const remove = useDeleteResource();

  if (partitions.isLoading || query.isLoading) return <CardSkeleton />;
  if (query.error) {
    return <ErrorState message={query.error.message} onRetry={() => query.refetch()} />;
  }
  const billing = (query.data ?? []).filter((b) => b.compute_cluster_id === clusterId);
  const rows = joinByName(billing, partitions.data ?? []);
  const slurmKnown = !partitions.error;

  const columns: Array<DataTableColumn<Row>> = [
    { key: "name", header: "Resource", cell: (r) => <b>{r.name}</b> },
    { key: "nodes", header: "Nodes", cell: (r) => r.slurm?.nodes ?? "–" },
    {
      key: "state",
      header: "State",
      cell: (r) =>
        r.slurm ? (
          r.slurm.state
        ) : slurmKnown ? (
          <span className="text-[color:var(--tone-warn-fg)]">Not in SLURM</span>
        ) : (
          "–"
        ),
    },
    {
      key: "max",
      header: "MaxTime",
      cell: (r) => (r.slurm ? formatMinutes(r.slurm.max_time_minutes) : "–"),
    },
    {
      key: "default",
      header: "DefaultTime",
      cell: (r) => (r.slurm ? formatMinutes(r.slurm.default_time_minutes) : "–"),
    },
    {
      key: "tres",
      header: "TRES",
      cell: (r) => r.custos?.resource_type ?? "–",
    },
    {
      key: "rate",
      header: "SU rate",
      align: "right",
      cell: (r) =>
        r.custos ? (
          <EffectiveRateCell resource={r.custos} />
        ) : (
          <span className="text-sm text-muted-foreground">Not billed</span>
        ),
    },
    {
      key: "allocation_count",
      header: "Allocations",
      align: "right",
      cell: (r) => (
        <span className="text-sm tabular-nums text-foreground">
          {r.custos ? r.custos.allocation_count : "–"}
        </span>
      ),
    },
    {
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (r) => {
        const b = r.custos;
        return (
          <div className="flex justify-end gap-2">
            <GuardedButton
              variant="outline"
              size="sm"
              allowed={can.allocationsWrite}
              reason={requires("allocationsWrite")}
              onClick={() => setEditing(r)}
            >
              {b ? "Edit" : "Set rate"}
            </GuardedButton>
            {b ? (
              <Button variant="outline" size="sm" onClick={() => setRatesFor(b)}>
                Rates ({b.rate_count})
              </Button>
            ) : null}
            {b && !r.slurm && slurmKnown ? (
              <GuardedButton
                variant="destructive"
                size="sm"
                allowed={can.allocationsWrite}
                reason={requires("allocationsWrite")}
                onClick={() => {
                  remove.reset();
                  setDeleting(b);
                }}
              >
                Delete
              </GuardedButton>
            ) : null}
          </div>
        );
      },
    },
  ];

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        SLURM partitions and their billing. SLURM columns are read-only.
      </p>
      {partitions.error ? (
        <p className="text-sm text-[color:var(--tone-warn-fg)]">
          SLURM unreachable; showing billing only.
        </p>
      ) : null}
      <DataTable
        columns={columns}
        rows={rows}
        rowKey={(r) => r.name}
        empty="No resources registered."
      />

      <ResourceDialog
        key={editing?.name}
        open={editing !== null}
        onOpenChange={(open) => {
          if (!open) setEditing(null);
        }}
        clusterId={clusterId}
        partition={editing?.name ?? ""}
        resource={editing?.custos ?? null}
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
