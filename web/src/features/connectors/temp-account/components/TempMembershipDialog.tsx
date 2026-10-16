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

import { dayAfter } from "@/features/core/allocations/components/AllocationStatusLabel";
import { useAllocation, useAllocations } from "@/features/core/allocations/queries";
import type { ComputeAllocationMembership } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import {
  Field,
  FormDialog,
  STATUS_OPTIONS,
  SelectField,
  dateInput,
  edited,
  idText,
  isoDate,
  named,
  status,
  text,
} from "@/shared/ui/FormDialog";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { idInput } from "@/shared/users-admin/idInput";
import type { UseQueryResult } from "@tanstack/react-query";
import * as React from "react";
import { useAssignTempAllocation, useUpdateTempAllocation } from "../queries";

export type TempMembershipDialogProps = {
  userId: string;
  // undefined: closed; null: assigning a new allocation; a membership: editing it.
  membership: ComputeAllocationMembership | null | undefined;
  // The page's memberships query; undefined without TempAccount read.
  memberships: UseQueryResult<ComputeAllocationMembership[]> | undefined;
  onClose: () => void;
};

function membershipWindow(form: FormData) {
  return {
    start_time: isoDate(form, "start"),
    end_time: isoDate(form, "end"),
    membership_status: status(form, "status"),
  };
}

export function TempMembershipDialog({
  userId,
  membership: opened,
  memberships,
  onClose,
}: TempMembershipDialogProps) {
  const assign = useAssignTempAllocation();
  const update = useUpdateTempAllocation();
  const record = opened ? memberships : undefined;
  const found = opened ? record?.data?.find((m) => m.id === opened.id) : undefined;
  const membership = opened && (found ?? opened);
  const [start, setStart] = React.useState<string>();
  const canReadAllocations = useAbility().can("read", "Allocation");
  const allocations = useAllocations(
    { status: "ACTIVE", limit: 200 },
    opened === null && canReadAllocations,
  );
  const allocationName = useAllocation(
    canReadAllocations ? membership?.compute_allocation_id : undefined,
  ).data?.name;
  const close = () => {
    setStart(undefined);
    onClose();
  };
  return (
    <FormDialog
      open={membership !== undefined}
      onOpenChange={(open) => (open ? null : close())}
      title={membership ? "Update allocation" : "Assign allocation"}
      description={
        membership
          ? "Change this temporary account's membership window or status."
          : `Grant this temporary account a membership on a compute allocation.${memberships ? "" : " Its current memberships are not shown without temporary-account read."}`
      }
      submitLabel={membership ? "Update" : "Assign"}
      isPending={assign.isPending || update.isPending}
      record={record}
      gone="This membership no longer exists"
      missing={Boolean(record) && !found}
      onSubmit={(form) => {
        // The connector replaces the membership, so an update sends all of it.
        const body =
          membership &&
          edited(
            {
              ...membership,
              start_time: `${dateInput(membership.start_time)}T00:00:00Z`,
              end_time: `${dateInput(membership.end_time)}T00:00:00Z`,
            },
            membershipWindow(form),
          );
        if (membership && !body) return close();
        toastOnSuccess(
          body
            ? update.mutateAsync(body)
            : assign.mutateAsync({
                body: {
                  user_id: userId,
                  compute_allocation_id: idText(form, "allocation"),
                  ...membershipWindow(form),
                },
              }),
          membership ? "Allocation updated" : "Allocation assigned",
          close,
        );
      }}
    >
      {membership && canReadAllocations ? (
        <Field
          label="Compute allocation"
          name="allocation-name"
          value={allocationName ?? "—"}
          readOnly
        />
      ) : !membership && canReadAllocations && !allocations.isError ? (
        <SelectField
          label="Compute allocation"
          name="allocation"
          options={named(allocations.data?.items?.toSorted((a, b) => a.name.localeCompare(b.name)))}
        />
      ) : (
        <Field
          label="Compute allocation ID"
          name="allocation"
          defaultValue={membership?.compute_allocation_id}
          readOnly={membership != null}
          required
          {...idInput("compute allocation")}
        />
      )}
      <div className="grid grid-cols-2 gap-4">
        <Field
          label="Start"
          name="start"
          type="date"
          defaultValue={dateInput(membership?.start_time)}
          required
          onChange={(e) => setStart(e.currentTarget.value)}
        />
        <Field
          label="End"
          name="end"
          type="date"
          defaultValue={dateInput(membership?.end_time)}
          min={dayAfter(start ?? dateInput(membership?.start_time))}
          required
        />
      </div>
      <SelectField
        label="Status"
        name="status"
        options={STATUS_OPTIONS}
        defaultValue={membership?.membership_status}
      />
    </FormDialog>
  );
}
