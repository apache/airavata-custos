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
import type { ComputeAllocationUsage } from "@/generated/core/types.gen";
import { listOf } from "@/shared/api/listOf";
import { formatNumber, formatSU } from "@/shared/format";
import { DataTable, type DataTableColumn, NameCell } from "@/shared/ui/DataTable";
import { Field, FormDialog, named, num, text } from "@/shared/ui/FormDialog";
import { Mono, RecordDrawer } from "@/shared/ui/RecordDrawer";
import { Button } from "@/shared/ui/button";
import { toastOnSuccess } from "@/shared/ui/sonner";
import type * as React from "react";
import {
  useAllocationMembers,
  useAllocationResources,
  useDeleteUsage,
  useRecordUsage,
  useUsageRecord,
} from "../queries";
import { ListSelect } from "./ListSelect";
import { memberName } from "./MemberName";
import { useDrawerParam } from "./useDrawerParam";

export type UsageRecordsProps = {
  allocationId: string;
  rows: ComputeAllocationUsage[];
  userName: (id: string) => React.ReactNode;
  resourceName: (id: string) => string;
  canRead: boolean;
  canManage: boolean;
};
export function UsageRecords({
  allocationId,
  rows,
  userName,
  resourceName,
  canRead,
  canManage,
}: UsageRecordsProps) {
  const [param, setSelectedId] = useDrawerParam("usage");
  // Record detail requires the allocations read privilege.
  const selectedId = canRead ? param : undefined;
  const usage = useUsageRecord(selectedId);
  const remove = useDeleteUsage(allocationId);
  const close = () => setSelectedId();
  const columns: Array<DataTableColumn<ComputeAllocationUsage>> = [
    {
      key: "when",
      header: "Updated",
      cell: (row) => <Timestamp iso={row.last_updated} />,
    },
    {
      key: "user",
      header: "User",
      cell: (row) => (
        <span className="block min-w-0 max-w-[40ch] truncate">{userName(row.user_id)}</span>
      ),
    },
    {
      key: "job",
      header: "Job",
      cell: (row) => <NameCell text={row.job_id} className="font-mono text-xs" />,
    },
    {
      key: "resource",
      header: "Resource",
      cell: (row) => <NameCell text={resourceName(row.compute_allocation_resource_id)} />,
    },
    {
      key: "su",
      header: "SUs",
      align: "right",
      cell: (row) => <span className="tabular-nums">{formatSU(row.used_su_amount)}</span>,
    },
  ];

  return (
    <>
      {rows.length > 0 || canManage ? (
        <section className="space-y-2">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <h2 className="text-sm font-semibold text-foreground">Usage records</h2>
            {canManage ? <RecordUsageDialog allocationId={allocationId} /> : null}
          </div>
          {rows.length > 0 ? (
            <DataTable
              caption="Usage records"
              columns={columns}
              rows={[...rows].sort(
                (a, b) => Date.parse(b.last_updated) - Date.parse(a.last_updated),
              )}
              rowKey={(row) => row.id}
              onRowClick={canRead ? (row) => setSelectedId(row.id) : undefined}
            />
          ) : null}
        </section>
      ) : null}
      <RecordDrawer
        title="Usage record"
        id={selectedId}
        onClose={close}
        query={usage}
        owned={(u) => u.compute_allocation_id === allocationId}
        fields={(u) => ({
          "Record ID": <Mono>{u.id}</Mono>,
          User: userName(u.user_id),
          "Job ID": <Mono>{u.job_id}</Mono>,
          Resource: resourceName(u.compute_allocation_resource_id),
          "Raw amount": formatNumber(u.used_raw_amount),
          SUs: formatSU(u.used_su_amount),
          Updated: <Timestamp iso={u.last_updated} />,
        })}
        remove={
          canManage
            ? {
                label: "Delete record",
                confirm: "Delete this usage record?",
                isPending: remove.isPending,
                onConfirm: () =>
                  toastOnSuccess(
                    remove.mutateAsync(selectedId ?? ""),
                    "Usage record deleted",
                    close,
                  ),
              }
            : undefined
        }
      />
    </>
  );
}

function RecordUsageDialog({ allocationId }: { allocationId: string }) {
  const record = useRecordUsage(allocationId);
  const resources = useAllocationResources(allocationId);
  const members = useAllocationMembers(allocationId);
  const blocker =
    resources.data?.length === 0
      ? "Attach a resource to this allocation before recording usage"
      : members.data?.length === 0
        ? "Add a member to this allocation before recording usage"
        : undefined;
  if (blocker) {
    return (
      <div className="flex flex-wrap items-center justify-end gap-2">
        <p className="text-xs text-muted-foreground">{blocker}.</p>
        <Button size="sm" disabled focusableWhenDisabled>
          + Record usage
        </Button>
      </div>
    );
  }
  return (
    <FormDialog
      trigger={<Button size="sm">+ Record usage</Button>}
      title="Record usage"
      description="Adds a usage record by hand, as the scheduler integration would."
      submitLabel="Record usage"
      isPending={record.isPending}
      onSubmit={(form, close) =>
        toastOnSuccess(
          record.mutateAsync({
            body: {
              compute_allocation_id: allocationId,
              compute_allocation_resource_id: text(form, "resource"),
              user_id: text(form, "user"),
              job_id: text(form, "job"),
              used_raw_amount: num(form, "raw"),
              used_su_amount: num(form, "su"),
            },
          }),
          "Usage recorded",
          close,
        )
      }
    >
      <ListSelect
        label="Resource"
        name="resource"
        things="resources"
        query={resources}
        options={named}
      />
      <ListSelect
        label="Member"
        name="user"
        things="members"
        query={members}
        options={(data) => listOf(data).map((m) => ({ value: m.user_id, label: memberName(m) }))}
      />
      <Field label="Job ID" name="job" maxLength={255} required />
      <Field label="Raw amount" name="raw" type="number" min={0} step="any" required />
      <Field label="SUs" name="su" type="number" min={0} step="any" required />
    </FormDialog>
  );
}
