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
import type { ComputeAllocationUsage } from "@/generated/core/types.gen";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import {
  useAllocationMembers,
  useAllocationResources,
  useDeleteUsage,
  useRecordUsage,
  useUsageRecord,
} from "../queries";
import { Mono, RecordDrawer } from "@/shared/ui/RecordDrawer";
import { Field, FormDialog, named, num, SelectField, text } from "@/shared/ui/FormDialog";

export type UsageRecordsProps = {
  allocationId: string;
  rows: ComputeAllocationUsage[];
  resourceName: (id?: string) => string;
  userName: (id?: string) => string;
  canRead: boolean;
  canManage: boolean;
};
export function UsageRecords({
  allocationId,
  rows,
  resourceName,
  userName,
  canRead,
  canManage,
}: UsageRecordsProps) {
  const [selectedId, setSelectedId] = React.useState<string>();
  const usage = useUsageRecord(selectedId);
  const remove = useDeleteUsage(allocationId);
  const close = () => setSelectedId(undefined);
  const columns: Array<DataTableColumn<ComputeAllocationUsage>> = [
    {
      key: "when",
      header: "Updated",
      cell: (row) => formatDateTime(row.last_updated),
    },
    { key: "user", header: "User", cell: (row) => userName(row.user_id) },
    { key: "job", header: "Job", cell: (row) => <Mono>{row.job_id}</Mono> },
    {
      key: "resource",
      header: "Resource",
      cell: (row) => resourceName(row.compute_allocation_resource_id),
    },
    {
      key: "su",
      header: "SUs",
      align: "right",
      cell: (row) => <span className="tabular-nums">{formatNumber(row.used_su_amount)}</span>,
    },
  ];

  return (
    <section className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-foreground">Usage records</h2>
        {canManage ? <RecordUsageDialog allocationId={allocationId} /> : null}
      </div>
      <DataTable
        columns={columns}
        rows={rows}
        rowKey={(row) => row.id ?? ""}
        onRowClick={canRead ? (row) => setSelectedId(row.id) : undefined}
      />
      <RecordDrawer
        title="Usage record"
        id={selectedId}
        onClose={close}
        query={usage}
        fields={(u) => ({
          "Record ID": <Mono>{u.id}</Mono>,
          User: userName(u.user_id),
          "Job ID": <Mono>{u.job_id}</Mono>,
          Resource: resourceName(u.compute_allocation_resource_id),
          "Raw amount": formatNumber(u.used_raw_amount),
          SUs: formatNumber(u.used_su_amount),
          Updated: formatDateTime(u.last_updated),
        })}
        remove={
          canManage
            ? {
                label: "Delete record",
                confirm: "Delete this usage record?",
                isPending: remove.isPending,
                onConfirm: () =>
                  remove.mutate(selectedId ?? "", toastOnSuccess("Usage record deleted", close)),
              }
            : undefined
        }
      />
    </section>
  );
}

function RecordUsageDialog({ allocationId }: { allocationId: string }) {
  const record = useRecordUsage(allocationId);
  const resources = useAllocationResources(allocationId);
  const members = useAllocationMembers(allocationId);
  return (
    <FormDialog
      trigger={<Button size="sm">+ Record usage</Button>}
      title="Record usage"
      description="Adds a usage record by hand, as the scheduler integration would."
      submitLabel="Record usage"
      isPending={record.isPending}
      onSubmit={(form, close) =>
        record.mutate(
          {
            body: {
              compute_allocation_id: allocationId,
              compute_allocation_resource_id: text(form, "resource"),
              user_id: text(form, "user"),
              job_id: text(form, "job"),
              used_raw_amount: num(form, "raw"),
              used_su_amount: num(form, "su"),
            },
          },
          toastOnSuccess("Usage recorded", close),
        )
      }
    >
      <SelectField
        label="Resource"
        name="resource"
        options={named(resources.data)}
      />
      <SelectField
        label="Member"
        name="user"
        options={(members.data ?? []).map((m) => ({
          value: m.user_id ?? "",
          label: m.display_name ?? m.user_id ?? "",
        }))}
      />
      <Field label="Job ID" name="job" required />
      <Field label="Raw amount" name="raw" type="number" min={0} step="any" required />
      <Field label="SUs" name="su" type="number" min={0} step="any" required />
    </FormDialog>
  );
}
