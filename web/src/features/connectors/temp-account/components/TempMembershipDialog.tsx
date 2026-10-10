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

import type { ComputeAllocationMembership } from "@/generated/temp-account/types.gen";
import { toastOnSuccess } from "@/shared/ui/sonner";
import {
  dateInput,
  Field,
  FormDialog,
  isoDate,
  SelectField,
  STATUS_OPTIONS,
  status,
  text,
} from "@/shared/ui/FormDialog";
import { useAssignTempAllocation, useUpdateTempAllocation } from "../queries";

export type TempMembershipDialogProps = {
  userId: string;
  // undefined: closed; null: assigning a new allocation; a membership: editing it.
  membership: ComputeAllocationMembership | null | undefined;
  onClose: () => void;
};

export function TempMembershipDialog({ userId, membership, onClose }: TempMembershipDialogProps) {
  const assign = useAssignTempAllocation();
  const update = useUpdateTempAllocation();
  return (
    <FormDialog
      open={membership !== undefined}
      onOpenChange={(open) => (open ? null : onClose())}
      title={membership ? "Update allocation" : "Assign allocation"}
      description={
        membership
          ? "Change this temporary account's membership window or status."
          : "Grant this temporary account a membership on a compute allocation."
      }
      submitLabel={membership ? "Update" : "Assign"}
      isPending={assign.isPending || update.isPending}
      onSubmit={(form) =>
        (membership ? update : assign).mutate(
          {
            body: {
              id: membership?.id,
              user_id: userId,
              compute_allocation_id: text(form, "allocation"),
              start_time: isoDate(form, "start"),
              end_time: isoDate(form, "end"),
              membership_status: status(form, "status"),
            },
          },
          toastOnSuccess(membership ? "Allocation updated" : "Allocation assigned", onClose),
        )
      }
    >
      <Field
        label="Compute allocation ID"
        name="allocation"
        defaultValue={membership?.compute_allocation_id}
        readOnly={membership != null}
        required
      />
      <div className="grid grid-cols-2 gap-4">
        <Field
          label="Start"
          name="start"
          type="date"
          defaultValue={dateInput(membership?.start_time)}
          required
        />
        <Field
          label="End"
          name="end"
          type="date"
          defaultValue={dateInput(membership?.end_time)}
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
