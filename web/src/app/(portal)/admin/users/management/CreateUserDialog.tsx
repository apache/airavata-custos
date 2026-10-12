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

import { useAllocations } from "@/features/core/allocations/queries";
import { useClusters } from "@/features/core/clusters/queries";
import { useCreateUser } from "@/features/core/users/queries";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Field, FormDialog, SelectField, idText, named, text } from "@/shared/ui/FormDialog";
import { Button } from "@/shared/ui/button";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { idInput } from "@/shared/users-admin/idInput";
import * as React from "react";

const NO_FLAGS = { portal_admin: false, cluster_admin: false };
const NO_ALLOCATION = "none";

// Onboarding approves the user's cluster account; with one cluster registered the backend picks it.
// A researcher's allocation, which only a researcher may join on creation, fixes the cluster.
export function CreateUserDialog() {
  const ability = useAbility();
  const [open, setOpen] = React.useState(false);
  const [flags, setFlags] = React.useState(NO_FLAGS);
  const create = useCreateUser();
  const clustersQuery = useClusters(open && ability.can("read", "Cluster"));
  const canReadClusters = ability.can("read", "Cluster") && !clustersQuery.isError;
  const clusters = clustersQuery.data ?? [];
  const researcher = !flags.portal_admin && !flags.cluster_admin;
  const allocationsQuery = useAllocations(
    { status: "ACTIVE", limit: 200 },
    open && researcher && ability.can("read", "Allocation"),
  );
  // A list that fails to load falls back to its typed ID.
  const canReadAllocations = ability.can("read", "Allocation") && !allocationsQuery.isError;
  const allocations = allocationsQuery.data?.items?.toSorted((a, b) =>
    a.name.localeCompare(b.name),
  );
  const portalOnly = flags.portal_admin && !flags.cluster_admin;

  if (!ability.can("write", "User")) return null;

  const flag = (key: keyof typeof NO_FLAGS, label: string) => (
    <label className="flex items-center gap-2">
      <input
        type="checkbox"
        className="size-4 rounded border-input"
        checked={flags[key]}
        onChange={(e) => setFlags((prev) => ({ ...prev, [key]: e.target.checked }))}
      />
      {label}
    </label>
  );

  return (
    <FormDialog
      open={open}
      onOpenChange={(next) => {
        if (next) setFlags(NO_FLAGS);
        setOpen(next);
      }}
      trigger={<Button>Create user</Button>}
      title="Create user"
      description="Onboards the user with an approved cluster account unless they are a portal-only admin."
      submitLabel="Create user"
      isPending={create.isPending}
      onSubmit={(form, close) => {
        const picked = text(form, "allocation");
        const allocationId =
          researcher && picked !== NO_ALLOCATION ? idText(form, "allocation") : "";
        toastOnSuccess(
          create.mutateAsync({
            ...flags,
            email: text(form, "email"),
            first_name: text(form, "first_name"),
            last_name: text(form, "last_name"),
            username: portalOnly ? undefined : text(form, "username") || undefined,
            allocation_id: allocationId || undefined,
            compute_cluster_id:
              portalOnly || allocationId ? undefined : idText(form, "cluster") || undefined,
          }),
          `Created ${text(form, "email")}`,
          close,
        );
      }}
    >
      <Field label="Email" name="email" type="email" maxLength={255} required autoFocus />
      <div className="grid grid-cols-2 gap-3">
        <Field
          label="First name"
          name="first_name"
          maxLength={255}
          required
          title="Enter a name, not only spaces."
        />
        <Field label="Last name" name="last_name" maxLength={255} />
      </div>
      {/* Onboarding an admin takes core:roles:manage. */}
      {ability.can("write", "Role") ? (
        <div className="flex gap-4 text-sm">
          {flag("portal_admin", "Portal admin")}
          {flag("cluster_admin", "Cluster admin")}
        </div>
      ) : null}
      {portalOnly ? null : (
        <>
          <Field
            label="Cluster username"
            name="username"
            maxLength={32}
            placeholder="Generated when blank"
          />
          {!canReadClusters ? (
            <Field
              label="Cluster ID"
              name="cluster"
              placeholder="The only cluster when blank"
              {...idInput("cluster")}
            />
          ) : clusters.length > 1 ? (
            <SelectField
              label="Cluster"
              name="cluster"
              options={named(clusters)}
              description={researcher ? "An allocation's own cluster takes precedence." : undefined}
            />
          ) : null}
        </>
      )}
      {researcher ? (
        canReadAllocations ? (
          <SelectField
            label="Allocation"
            name="allocation"
            options={[
              { value: NO_ALLOCATION, label: "No allocation, assign later" },
              ...named(allocations),
            ]}
            description="Needed before they can run jobs."
          />
        ) : (
          <Field
            label="Allocation ID"
            name="allocation"
            placeholder="Assign later when blank"
            {...idInput("allocation")}
          />
        )
      ) : null}
    </FormDialog>
  );
}
