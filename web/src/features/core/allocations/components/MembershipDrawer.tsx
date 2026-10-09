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
import { UserName } from "@/features/core/users/components/UserPicker";
import type { ComputeAllocationMembershipResourceOverride } from "@/generated/core/types.gen";
import { formatDate } from "@/shared/format";
import { Button } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import {
  useAllocationResources,
  useCreateOverride,
  useDeleteOverride,
  useMembership,
  useMembershipOverrides,
  useSetMembershipStatus,
  useUpdateMembership,
  useUpdateOverride,
} from "../queries";
import { GrantFields } from "./AllocationOverviewTab";
import { RecordDrawer } from "@/shared/ui/RecordDrawer";
import {
  dateInput,
  Field,
  FormDialog,
  isoDate,
  named,
  num,
  SelectField,
  text,
} from "@/shared/ui/FormDialog";

export type MembershipDrawerProps = {
  allocationId: string;
  membershipId: string | undefined;
  canManage: boolean;
  onClose: () => void;
};

export function MembershipDrawer({
  allocationId,
  membershipId,
  canManage,
  onClose,
}: MembershipDrawerProps) {
  const query = useMembership(membershipId);
  const update = useUpdateMembership(allocationId);
  const setStatus = useSetMembershipStatus(allocationId);
  const id = membershipId ?? "";

  return (
    <RecordDrawer
      title="Membership"
      id={membershipId}
      onClose={onClose}
      query={query}
      fields={(m) => ({
        User: <UserName id={m.user_id} />,
        Status: m.membership_status,
        Start: formatDate(m.start_time),
        End: formatDate(m.end_time),
      })}
    >
      {(membership) => {
        const nextStatus = membership.membership_status === "ACTIVE" ? "INACTIVE" : "ACTIVE";
        return (
          <>
            {canManage ? (
              <div className="space-y-3 border-t border-border pt-4">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={setStatus.isPending}
                  onClick={() =>
                    setStatus.mutate(
                      { path: { id }, body: { membership_status: nextStatus } },
                      toastOnSuccess(`Membership marked ${nextStatus}`),
                    )
                  }
                >
                  {nextStatus === "ACTIVE" ? "Activate" : "Deactivate"}
                </Button>
                <form
                  key={`${membership.start_time}-${membership.end_time}`}
                  className="flex items-end gap-2"
                  onSubmit={(e) => {
                    e.preventDefault();
                    const form = new FormData(e.currentTarget);
                    update.mutate(
                      {
                        path: { id },
                        body: {
                          start_time: isoDate(form, "start"),
                          end_time: isoDate(form, "end"),
                        },
                      },
                      toastOnSuccess("Membership updated"),
                    );
                  }}
                >
                  <Field
                    label="Start"
                    name="start"
                    type="date"
                    defaultValue={dateInput(membership.start_time)}
                    required
                  />
                  <Field
                    label="End"
                    name="end"
                    type="date"
                    defaultValue={dateInput(membership.end_time)}
                    required
                  />
                  <Button type="submit" size="sm" disabled={update.isPending}>
                    Save
                  </Button>
                </form>
              </div>
            ) : null}
            <ResourceOverrides
              allocationId={allocationId}
              membershipId={id}
              canManage={canManage}
            />
          </>
        );
      }}
    </RecordDrawer>
  );
}

function ResourceOverrides({
  allocationId,
  membershipId,
  canManage,
}: {
  allocationId: string;
  membershipId: string;
  canManage: boolean;
}) {
  const query = useMembershipOverrides(membershipId);
  const resources = useAllocationResources(allocationId);
  const create = useCreateOverride(membershipId);
  const remove = useDeleteOverride(membershipId);
  const [editing, setEditing] = React.useState<ComputeAllocationMembershipResourceOverride>();
  const resourceName = (id = "") => resources.data?.find((r) => r.id === id)?.name ?? id;
  const overrides = query.data ?? [];
  // One override per resource, so only offer the ones this member has none for.
  const available = (resources.data ?? []).filter(
    (r) => !overrides.some((o) => o.compute_allocation_resource_id === r.id),
  );

  return (
    <section className="space-y-2 border-t border-border pt-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">Resource overrides</h3>
        {canManage && query.data && available.length > 0 ? (
          <FormDialog
            trigger={<Button size="sm">+ Add override</Button>}
            title="Add resource override"
            submitLabel="Add override"
            isPending={create.isPending}
            onSubmit={(form, close) =>
              create.mutate(
                {
                  body: {
                    compute_allocation_membership_id: membershipId,
                    compute_allocation_resource_id: text(form, "resource"),
                    override_resource_amount: num(form, "amount"),
                    override_resource_time: num(form, "time"),
                  },
                },
                toastOnSuccess("Override added", close),
              )
            }
          >
            <SelectField label="Resource" name="resource" options={named(available)} />
            <GrantFields />
          </FormDialog>
        ) : null}
      </div>
      {query.error ? (
        <ErrorState message={query.error.message} onRetry={() => query.refetch()} />
      ) : query.isLoading ? (
        <p className="text-sm text-muted-foreground">Loading overrides…</p>
      ) : overrides.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No overrides; the allocation's resource grants apply.
        </p>
      ) : (
        <ul className="space-y-1">
          {overrides.map((o) => (
            <li
              key={o.id}
              className="flex items-center justify-between rounded-md border bg-card px-3 py-2 text-sm"
            >
              <div>
                <div className="font-medium">{resourceName(o.compute_allocation_resource_id)}</div>
                <div className="text-xs text-muted-foreground">
                  {o.override_resource_amount} units · {o.override_resource_time} min
                </div>
              </div>
              {canManage ? (
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" onClick={() => setEditing(o)}>
                    Edit
                  </Button>
                  <Button
                    size="sm"
                    variant="destructive"
                    disabled={remove.isPending}
                    onClick={() =>
                      confirmToast("Delete this resource override?", "Delete", () =>
                        remove.mutate(
                          { path: { id: o.id ?? "" } },
                          toastOnSuccess("Override deleted"),
                        ),
                      )
                    }
                  >
                    Delete
                  </Button>
                </div>
              ) : null}
            </li>
          ))}
        </ul>
      )}
      <EditOverrideDialog
        membershipId={membershipId}
        override={editing}
        onClose={() => setEditing(undefined)}
      />
    </section>
  );
}

function EditOverrideDialog({
  membershipId,
  override,
  onClose,
}: {
  membershipId: string;
  override: ComputeAllocationMembershipResourceOverride | undefined;
  onClose: () => void;
}) {
  const update = useUpdateOverride(membershipId);
  return (
    <FormDialog
      open={Boolean(override)}
      onOpenChange={(open) => (open ? null : onClose())}
      title="Edit resource override"
      submitLabel="Save"
      isPending={update.isPending}
      onSubmit={(form) =>
        update.mutate(
          {
            path: { id: override?.id ?? "" },
            body: {
              override_resource_amount: num(form, "amount"),
              override_resource_time: num(form, "time"),
            },
          },
          toastOnSuccess("Override updated", onClose),
        )
      }
    >
      {override ? (
        <GrantFields
          key={override.id}
          amount={override.override_resource_amount}
          time={override.override_resource_time}
        />
      ) : null}
    </FormDialog>
  );
}
