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

import { formatDateTime, formatNumber } from "@/shared/format";
import * as React from "react";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge, statusBadgeVariantFromAllocationStatus } from "@/shared/ui/StatusBadge";
import { useAllocationDiffs, useDeleteDiff, useDiff, useRecordDiff } from "../queries";
import type { ComputeAllocation, ComputeAllocationDiff } from "@/generated/core/types.gen";
import { RecordDrawer, Mono } from "@/shared/ui/RecordDrawer";
import { Field, FormDialog, num, SelectField, STATUS_OPTIONS, status, text } from "@/shared/ui/FormDialog";

export type AllocationHistoryTabProps = {
  allocation: ComputeAllocation;
  canManage: boolean;
};

function RecordDiffDialog({ allocation }: { allocation: ComputeAllocation }) {
  const record = useRecordDiff();
  return (
    <FormDialog
      trigger={<Button size="sm">+ Record diff</Button>}
      title="Record diff"
      description="The latest diff sets the allocation's current SU amount and status."
      submitLabel="Record diff"
      isPending={record.isPending}
      onSubmit={(form, close) =>
        record.mutate(
          {
            body: {
              compute_allocation_id: allocation.id,
              diff_type: text(form, "type"),
              new_su_amount: num(form, "su"),
              status: status(form, "status"),
              description: text(form, "description"),
            },
          },
          toastOnSuccess("Diff recorded", close),
        )
      }
    >
      <Field label="Type" name="type" placeholder="MANUAL_ADJUSTMENT" required />
      <Field label="New SU amount" name="su" type="number" min={0} required />
      <SelectField
        label="Status"
        name="status"
        options={STATUS_OPTIONS}
        defaultValue={allocation.status}
      />
      <Field label="Description" name="description" />
    </FormDialog>
  );
}

export function AllocationHistoryTab({ allocation, canManage }: AllocationHistoryTabProps) {
  const query = useAllocationDiffs(allocation.id);
  const [selectedId, setSelectedId] = React.useState<string>();
  const diff = useDiff(selectedId);
  const remove = useDeleteDiff(allocation.id ?? "");
  const close = () => setSelectedId(undefined);

  if (query.isLoading) return <TableSkeleton rows={3} columns={5} />;
  if (query.error) {
    return <ErrorState message={query.error.message} onRetry={() => query.refetch()} />;
  }
  const rows = [...(query.data ?? [])].sort(
    (a, b) => Date.parse(b.timestamp ?? "") - Date.parse(a.timestamp ?? ""),
  );

  const cta = canManage ? (
    <div className="flex justify-end">
      <RecordDiffDialog allocation={allocation} />
    </div>
  ) : null;
  if (rows.length === 0) {
    return (
      <div className="space-y-3">
        {cta}
        <EmptyState heading="No history recorded for this allocation." />
      </div>
    );
  }

  const columns: Array<DataTableColumn<ComputeAllocationDiff>> = [
    {
      key: "when",
      header: "When",
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{formatDateTime(row.timestamp)}</span>
      ),
    },
    {
      key: "type",
      header: "Type",
      cell: (row) => <span className="text-sm">{row.diff_type}</span>,
    },
    {
      key: "su",
      header: "New SU amount",
      align: "right",
      cell: (row) => (
        <span className="tabular-nums">{formatNumber(row.new_su_amount)}</span>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <StatusBadge
          variant={statusBadgeVariantFromAllocationStatus(row.status)}
          label={row.status}
        />
      ),
    },
    {
      key: "description",
      header: "Description",
      cell: (row) =>
        row.description ? (
          <span className="text-sm text-muted-foreground">{row.description}</span>
        ) : null,
    },
  ];

  return (
    <div className="space-y-3">
      {cta}
      <DataTable
        columns={columns}
        rows={rows}
        rowKey={(row) => row.id ?? ""}
        onRowClick={(row) => setSelectedId(row.id)}
      />
      <RecordDrawer
        title="Allocation diff"
        id={selectedId}
        onClose={close}
        query={diff}
        fields={(d) => ({
          "Diff ID": <Mono>{d.id}</Mono>,
          When: formatDateTime(d.timestamp),
          Type: d.diff_type,
          "New SU amount": formatNumber(d.new_su_amount),
          Status: d.status,
          Description: <span className="whitespace-pre-wrap">{d.description}</span>,
        })}
        remove={
          canManage
            ? {
                label: "Delete diff",
                confirm: "Delete this diff?",
                isPending: remove.isPending,
                onConfirm: () =>
                  remove.mutate(selectedId ?? "", toastOnSuccess("Diff deleted", close)),
              }
            : undefined
        }
      />
    </div>
  );
}
