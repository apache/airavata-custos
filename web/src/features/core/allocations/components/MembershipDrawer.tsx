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

import { useResources } from "@/features/core/resources/queries";
import type {
  ComputeAllocationMembership,
  ComputeAllocationMembershipResourceOverride,
} from "@/generated/core/types.gen";
import { formatNumber } from "@/shared/format";
import { ErrorState } from "@/shared/ui/ErrorState";
import {
  Field,
  FormDialog,
  SelectField,
  dateInput,
  edited,
  firstField,
  isoDate,
  named,
  num,
  statusLabel,
  text,
  useFreshOnOpen,
} from "@/shared/ui/FormDialog";
import { RecordDrawer } from "@/shared/ui/RecordDrawer";
import { Button } from "@/shared/ui/button";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import * as React from "react";
import {
  useAllocationResources,
  useCreateOverride,
  useDeleteOverride,
  useMembership,
  useMembershipOverrides,
  useRefreshAllocationMembers,
  useSetMembershipStatus,
  useUpdateMembership,
  useUpdateOverride,
} from "../queries";
import { GrantFields } from "./AllocationOverviewTab";
import { dayAfter, knownDay, knownTime, membershipStatusText } from "./AllocationStatusLabel";

export type MembershipDrawerProps = {
  allocationId: string;
  membershipId: string | undefined;
  // The member's name from the allocation's member list, the membership record having none;
  // undefined for a membership the list does not show.
  name: string | undefined;
  canManage: boolean;
  onClose: () => void;
};

export function MembershipDrawer({
  allocationId,
  membershipId,
  name,
  canManage,
  onClose,
}: MembershipDrawerProps) {
  const query = useMembership(membershipId);
  // The window form prefills from the membership as the server holds it when the drawer opens.
  const refreshMembers = useRefreshAllocationMembers(allocationId);
  const fresh = useFreshOnOpen(query, Boolean(membershipId), { onGone: refreshMembers });
  const setStatus = useSetMembershipStatus(allocationId);
  const windowForm = React.useRef<HTMLFormElement>(null);
  // The form takes focus once per opening; a row swap keeps focus where the press left it.
  const focused = React.useRef(false);
  React.useEffect(() => {
    if (!membershipId) focused.current = false;
    else if (fresh.ready && !focused.current) {
      focused.current = true;
      firstField(windowForm.current)?.focus();
    }
  }, [membershipId, fresh.ready]);

  return (
    <RecordDrawer
      title="Membership"
      id={membershipId}
      onClose={onClose}
      initialFocus={() => firstField(windowForm.current) ?? true}
      query={query}
      owned={(m) => m.compute_allocation_id === allocationId}
      fields={(m) => ({
        ...(name ? { User: name } : {}),
        Status: membershipStatusText(m.membership_status),
        Start: knownDay(m.start_time),
        End: knownDay(m.end_time),
      })}
    >
      {(membership) => (
        <>
          {canManage ? (
            <div className="space-y-3 border-t border-border pt-4">
              <Button
                variant="outline"
                size="sm"
                pending={setStatus.isPending}
                onClick={() => {
                  const next = membership.membership_status === "ACTIVE" ? "INACTIVE" : "ACTIVE";
                  toastOnSuccess(
                    setStatus.mutateAsync({
                      path: { id: membership.id },
                      body: { membership_status: next },
                    }),
                    `Membership marked ${statusLabel(next).toLowerCase()}`,
                  );
                }}
              >
                {membership.membership_status === "ACTIVE" ? "Deactivate" : "Activate"}
              </Button>
              <MembershipWindow
                key={membership.id}
                membership={membership}
                allocationId={allocationId}
                ready={fresh.ready}
                formRef={windowForm}
              />
            </div>
          ) : null}
          <ResourceOverrides
            key={membership.id}
            allocationId={allocationId}
            membershipId={membership.id}
            canManage={canManage}
          />
        </>
      )}
    </RecordDrawer>
  );
}

// Keyed by membership, so a save keeps the form and its focus while a row swap starts afresh.
function MembershipWindow({
  membership,
  allocationId,
  ready,
  formRef,
}: {
  membership: ComputeAllocationMembership;
  allocationId: string;
  ready: boolean;
  formRef: React.RefObject<HTMLFormElement | null>;
}) {
  const update = useUpdateMembership(allocationId);
  const start = dateInput(knownTime(membership.start_time));
  const end = dateInput(knownTime(membership.end_time));
  const [startDay, setStartDay] = React.useState(start);
  return (
    <form
      ref={formRef}
      className="flex flex-wrap items-end gap-2"
      onSubmit={(e) => {
        e.preventDefault();
        if (!ready) return;
        const form = new FormData(e.currentTarget);
        if (text(form, "start") === start && text(form, "end") === end) return;
        toastOnSuccess(
          update.mutateAsync({
            path: { id: membership.id },
            // The server keeps the stored status, which another tab may have changed.
            body: {
              compute_allocation_id: membership.compute_allocation_id,
              user_id: membership.user_id,
              start_time: isoDate(form, "start"),
              end_time: isoDate(form, "end"),
            },
          }),
          "Membership updated",
        );
      }}
    >
      <Field
        label="Start"
        name="start"
        type="date"
        defaultValue={start}
        required
        onChange={(e) => setStartDay(e.currentTarget.value)}
      />
      <Field
        label="End"
        name="end"
        type="date"
        defaultValue={end}
        min={dayAfter(startDay)}
        required
      />
      <Button
        type="submit"
        size="sm"
        pending={update.isPending}
        disabled={!ready}
        focusableWhenDisabled
      >
        Save
      </Button>
    </form>
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
  // Only allocation readers open this drawer, and they may list the catalog.
  const catalog = useResources();
  const create = useCreateOverride(allocationId, membershipId);
  const remove = useDeleteOverride(membershipId);
  const [editing, setEditing] = React.useState<ComputeAllocationMembershipResourceOverride>();
  // An override outlives its resource's detachment, so it is named from the catalog as well.
  const resourceName = (id: string) =>
    [...(resources.data ?? []), ...(catalog.data ?? [])].find((r) => r.id === id)?.name ??
    (resources.isSuccess && catalog.isSuccess ? "Unknown resource" : "—");
  const headingId = React.useId();
  const overrides = query.data ?? [];
  // One override per attached resource, so only offer the ones this member has none for.
  const available = (resources.data ?? []).filter(
    (r) => !overrides.some((o) => o.compute_allocation_resource_id === r.id),
  );

  return (
    <section
      aria-labelledby={headingId}
      className="space-y-2 border-t border-border pt-4 outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 id={headingId} className="text-sm font-semibold">
          Resource overrides
        </h3>
        {canManage && query.data && resources.data?.length ? (
          <div className="flex flex-wrap items-center justify-end gap-2">
            {available.length === 0 ? (
              <p className="text-xs text-muted-foreground">
                Every attached resource has an override.
              </p>
            ) : null}
            <FormDialog
              trigger={
                <Button size="sm" disabled={available.length === 0} focusableWhenDisabled>
                  + Add override
                </Button>
              }
              title="Add resource override"
              submitLabel="Add override"
              isPending={create.isPending}
              onSubmit={(form, close) =>
                toastOnSuccess(
                  create.mutateAsync({
                    body: {
                      compute_allocation_membership_id: membershipId,
                      compute_allocation_resource_id: text(form, "resource"),
                      override_resource_amount: num(form, "amount"),
                      override_resource_time: num(form, "time"),
                    },
                  }),
                  "Override added",
                  close,
                )
              }
            >
              <SelectField label="Resource" name="resource" options={named(available)} />
              <GrantFields />
            </FormDialog>
          </div>
        ) : null}
      </div>
      {query.error ? (
        <ErrorState onRetry={() => query.refetch()} />
      ) : query.isLoading ? (
        <output aria-label="Loading" className="block text-sm text-muted-foreground">
          Loading overrides…
        </output>
      ) : overrides.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No overrides; the allocation's resource grants apply.
        </p>
      ) : (
        <ul className="space-y-1">
          {overrides.map((o) => {
            const resource = resourceName(o.compute_allocation_resource_id);
            return (
              <li
                key={o.id}
                className="flex items-center justify-between gap-3 rounded-md border bg-card px-3 py-2 text-sm"
              >
                <div className="min-w-0 [overflow-wrap:anywhere]">
                  <div className="font-medium">{resource}</div>
                  <div className="text-xs text-muted-foreground">
                    {formatNumber(o.override_resource_amount)} units ·{" "}
                    {formatNumber(o.override_resource_time)} min
                  </div>
                </div>
                {canManage ? (
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      aria-label={`Edit override for ${resource}`}
                      onClick={() => setEditing(o)}
                    >
                      Edit
                    </Button>
                    <Button
                      size="sm"
                      variant="destructive"
                      aria-label={`Delete override for ${resource}`}
                      pending={remove.isPending && remove.variables?.path.id === o.id}
                      onClick={() =>
                        confirmToast("Delete this resource override?", "Delete", () =>
                          toastOnSuccess(
                            remove.mutateAsync({ path: { id: o.id } }),
                            "Override deleted",
                          ),
                        )
                      }
                    >
                      Delete
                    </Button>
                  </div>
                ) : null}
              </li>
            );
          })}
        </ul>
      )}
      <EditOverrideDialog
        allocationId={allocationId}
        membershipId={membershipId}
        override={editing}
        onClose={() => setEditing(undefined)}
      />
    </section>
  );
}

function EditOverrideDialog({
  allocationId,
  membershipId,
  override,
  onClose,
}: {
  allocationId: string;
  membershipId: string;
  override: ComputeAllocationMembershipResourceOverride | undefined;
  onClose: () => void;
}) {
  const update = useUpdateOverride(allocationId, membershipId);
  const record = useMembershipOverrides(membershipId);
  const fresh = override && record.data?.find((o) => o.id === override.id);
  const refreshMembers = useRefreshAllocationMembers(allocationId);
  return (
    <FormDialog
      open={Boolean(override)}
      onOpenChange={(open) => (open ? null : onClose())}
      title="Edit resource override"
      submitLabel="Save"
      isPending={update.isPending}
      record={override ? record : undefined}
      gone="This override no longer exists"
      missing={!fresh}
      onGone={refreshMembers}
      onSubmit={(form) => {
        const body =
          fresh &&
          edited(fresh, {
            override_resource_amount: num(form, "amount"),
            override_resource_time: num(form, "time"),
          });
        if (!body) return onClose();
        toastOnSuccess(
          update.mutateAsync({ path: { id: body.id }, body }),
          "Override updated",
          onClose,
        );
      }}
    >
      {fresh ? (
        <GrantFields
          key={fresh.id}
          amount={fresh.override_resource_amount}
          time={fresh.override_resource_time}
        />
      ) : null}
    </FormDialog>
  );
}
