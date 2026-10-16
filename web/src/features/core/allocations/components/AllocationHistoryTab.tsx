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
import { DIFF_LABELS } from "@/features/core/audit/eventLabels";
import type { ComputeAllocation, ComputeAllocationDiff } from "@/generated/core/types.gen";
import { enumLabel, formatSU } from "@/shared/format";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import {
  Field,
  FormDialog,
  STATUS_OPTIONS,
  SelectField,
  num,
  status,
  statusLabel,
  text,
} from "@/shared/ui/FormDialog";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Mono, RecordDrawer } from "@/shared/ui/RecordDrawer";
import { StatusBadge, statusBadgeVariantFromAllocationStatus } from "@/shared/ui/StatusBadge";
import { Button } from "@/shared/ui/button";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { useAllocationDiffs, useDeleteDiff, useDiff, useRecordDiff } from "../queries";
import { useDrawerParam } from "./useDrawerParam";

const diffLabel = (type: string) => DIFF_LABELS[type] ?? enumLabel(type);
// An approval's diff names the change request by id, which the History tab does not link.
const diffText = (description?: string) =>
  description?.replace(/^(Applied approved change request) \S+$/, "$1");

export type AllocationHistoryTabProps = {
  allocation: ComputeAllocation;
  canManage: boolean;
};

// The form starts from the latest recorded SU total and status, else the allocation's own.
function RecordDiffDialog({
  allocation,
  latest,
}: {
  allocation: ComputeAllocation;
  latest: ComputeAllocationDiff | undefined;
}) {
  const record = useRecordDiff();
  return (
    <FormDialog
      trigger={<Button size="sm">+ Record diff</Button>}
      title="Record diff"
      description="Records the allocation's new SU total and status."
      submitLabel="Record diff"
      isPending={record.isPending}
      onSubmit={(form, close) =>
        toastOnSuccess(
          record.mutateAsync({
            body: {
              compute_allocation_id: allocation.id,
              diff_type: text(form, "type"),
              new_su_amount: num(form, "su"),
              status: status(form, "status"),
              description: text(form, "description"),
            },
          }),
          "Diff recorded",
          close,
        )
      }
    >
      <Field label="Type" name="type" placeholder="MANUAL_ADJUSTMENT" maxLength={64} required />
      <Field
        label="New SU total"
        name="su"
        type="number"
        min={0}
        max={Number.MAX_SAFE_INTEGER}
        step="any"
        defaultValue={latest?.new_su_amount ?? allocation.initial_su_amount}
        required
      />
      <SelectField
        label="Status"
        name="status"
        options={STATUS_OPTIONS}
        defaultValue={latest?.status ?? allocation.status}
      />
      <Field label="Description" name="description" />
    </FormDialog>
  );
}

export function AllocationHistoryTab({ allocation, canManage }: AllocationHistoryTabProps) {
  const query = useAllocationDiffs(allocation.id);
  const [selectedId, setSelectedId] = useDrawerParam("diff");
  const diff = useDiff(selectedId);
  const remove = useDeleteDiff(allocation.id);
  const close = () => setSelectedId();

  if (query.isLoading) return <TableSkeleton rows={3} columns={5} />;
  if (query.error) {
    return <ErrorState onRetry={() => query.refetch()} />;
  }
  const rows = [...(query.data ?? [])].sort(
    (a, b) => Date.parse(b.timestamp) - Date.parse(a.timestamp),
  );

  const cta = canManage ? (
    <div className="flex justify-end">
      <RecordDiffDialog allocation={allocation} latest={rows[0]} />
    </div>
  ) : null;
  const columns: Array<DataTableColumn<ComputeAllocationDiff>> = [
    {
      key: "when",
      header: "When",
      cell: (row) => <Timestamp iso={row.timestamp} className="text-sm text-muted-foreground" />,
    },
    {
      key: "type",
      header: "Type",
      cell: (row) => <span className="text-sm">{diffLabel(row.diff_type)}</span>,
    },
    {
      key: "su",
      header: "SU total",
      align: "right",
      cell: (row) => <span className="tabular-nums">{formatSU(row.new_su_amount)}</span>,
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <StatusBadge
          variant={statusBadgeVariantFromAllocationStatus(row.status)}
          label={statusLabel(row.status)}
        />
      ),
    },
    {
      key: "description",
      header: "Description",
      cell: (row) => (
        <span className="block max-w-[40ch] text-sm text-muted-foreground [overflow-wrap:anywhere]">
          {diffText(row.description) || "—"}
        </span>
      ),
    },
  ];

  return (
    <div className="space-y-3">
      {cta}
      {rows.length === 0 ? (
        <EmptyState heading="No history recorded for this allocation" />
      ) : (
        <DataTable
          caption="Allocation history"
          columns={columns}
          rows={rows}
          rowKey={(row) => row.id}
          onRowClick={(row) => setSelectedId(row.id)}
        />
      )}
      <RecordDrawer
        title="Allocation diff"
        id={selectedId}
        onClose={close}
        query={diff}
        owned={(d) => d.compute_allocation_id === allocation.id}
        fields={(d) => ({
          "Diff ID": <Mono>{d.id}</Mono>,
          When: <Timestamp iso={d.timestamp} />,
          Type: diffLabel(d.diff_type),
          "SU total": formatSU(d.new_su_amount),
          Status: statusLabel(d.status),
          Description: (
            <span className="whitespace-pre-wrap">{diffText(d.description) || "—"}</span>
          ),
        })}
        remove={
          canManage
            ? {
                label: "Delete diff",
                confirm: "Delete this diff?",
                isPending: remove.isPending,
                onConfirm: () =>
                  toastOnSuccess(remove.mutateAsync(selectedId ?? ""), "Diff deleted", close),
              }
            : undefined
        }
      />
    </div>
  );
}
