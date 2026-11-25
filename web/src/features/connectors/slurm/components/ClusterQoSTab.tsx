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

import * as React from "react";
import {
  useClusterQoS,
  useCreateQoS,
  useDeleteQoS,
  useUpdateQoS,
} from "@/features/core/clusters/queries";
import type { ComputeClusterQoS } from "@/generated/core/zod.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { ConfirmDialog } from "@/shared/ui/ConfirmDialog";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { DropdownMenuItem } from "@/shared/ui/dropdown-menu";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import type { Conflict } from "../api";
import { asRecord, type Joined, joinByName, limitLines } from "../format";
import { conflictsOf, useSlurmCluster } from "../queries";
import { QoSForm } from "./QoSForm";
import { SlurmActions } from "./SlurmActions";
import { SlurmValue } from "./SlurmValue";

type Side = Partial<ComputeClusterQoS>;

// A definition's Slurm side exists only when it conflicts.
type Row = Joined<ComputeClusterQoS, Conflict>;

export function ClusterQoSTab({ clusterId }: { clusterId: string }) {
  const ability = useAbility();
  const conflicts = useSlurmCluster(clusterId, conflictsOf);
  const qos = useClusterQoS(clusterId);
  const remove = useDeleteQoS(clusterId);
  const create = useCreateQoS(clusterId);
  const update = useUpdateQoS(clusterId);
  const [editing, setEditing] = React.useState<ComputeClusterQoS | "new" | null>(null);
  const [deleting, setDeleting] = React.useState<ComputeClusterQoS | null>(null);
  const closeEditing = () => {
    conflicts.refetch();
    setEditing(null);
  };

  const columns: Array<DataTableColumn<Row>> = [
    {
      key: "name",
      header: "QoS",
      cell: (r) => (
        <span>
          <b>{r.name}</b>
          {r.custos?.is_default ? (
            <Badge variant="secondary" className="ml-2">
              default
            </Badge>
          ) : null}
          {r.custos ? null : (
            <Badge variant="secondary" className="ml-2">
              only on SLURM
            </Badge>
          )}
        </span>
      ),
    },
    { key: "priority", header: "Priority", cell: sideBySide((q) => q.priority) },
    { key: "limits", header: "Limits", cell: sideBySide(limitLines) },
    { key: "flags", header: "Flags", cell: sideBySide((q) => q.flags) },
    {
      key: "actions",
      header: "Actions",
      interactive: true,
      cell: (r) => (
        <SlurmActions
          clusterId={clusterId}
          scope="cluster"
          syncUnknown={!conflicts.isSuccess}
          conflicts={r.slurm ? [r.slurm] : []}
          items={
            r.custos ? (
              <>
                <DropdownMenuItem onClick={() => setEditing(r.custos as ComputeClusterQoS)}>
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => setDeleting(r.custos as ComputeClusterQoS)}>
                  Delete
                </DropdownMenuItem>
              </>
            ) : undefined
          }
        />
      ),
    },
  ];

  if (qos.isLoading) return <CardSkeleton />;
  if (qos.error) return <ErrorState message={qos.error.message} onRetry={() => qos.refetch()} />;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">QoS definitions Custos pushes to SLURM.</p>
        {ability.can("manage", "Cluster") ? (
          <Button size="sm" onClick={() => setEditing("new")}>
            + New QoS
          </Button>
        ) : null}
      </div>
      <DataTable
        columns={columns}
        rows={joinByName(
          qos.data ?? [],
          (conflicts.data ?? []).filter((c) => c.kind === "qos_definition"),
        )}
        rowKey={(r) => r.name}
        empty="No QoS definitions."
      />
      {editing === "new" ? (
        <QoSForm
          title="New QoS"
          pending={create.isPending}
          error={create.error?.message}
          onCancel={() => setEditing(null)}
          onSubmit={(body) => create.mutate(body, { onSuccess: closeEditing })}
        />
      ) : editing ? (
        <QoSForm
          title={`Edit QoS · ${editing.name}`}
          qos={editing}
          pending={update.isPending}
          error={update.error?.message}
          onCancel={() => setEditing(null)}
          onSubmit={(body) => update.mutate({ id: editing.id, body }, { onSuccess: closeEditing })}
        />
      ) : null}
      <ConfirmDialog
        open={deleting !== null}
        onOpenChange={(open) => !open && setDeleting(null)}
        title={`Delete ${deleting?.name ?? "QoS"}?`}
        confirmLabel="Delete"
        pending={remove.isPending}
        error={remove.error}
        onConfirm={() => {
          if (deleting)
            remove.mutate(deleting.id, {
              onSuccess: () => {
                conflicts.refetch();
                setDeleting(null);
              },
            });
        }}
      >
        It stays on SLURM as a conflict until you resolve it.
      </ConfirmDialog>
    </div>
  );
}

// Custos's value, else Slurm's; Slurm's shows below when both exist and differ.
const sideBySide = (read: (q: Side) => unknown) => (r: Row) => {
  const slurm = asRecord<Side>(r.slurm?.slurm);
  return (
    <SlurmValue
      custos={read(r.custos ?? slurm ?? {})}
      {...(r.custos && slurm ? { slurm: read(slurm) } : {})}
    />
  );
};
