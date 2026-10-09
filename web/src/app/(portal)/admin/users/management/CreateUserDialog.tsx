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
import { useClusters } from "@/features/core/clusters/queries";
import { useCreateUser } from "@/features/core/users/queries";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";
import { Field, FormDialog, named, SelectField, text } from "@/shared/ui/FormDialog";
import { toastOnSuccess } from "@/shared/ui/sonner";

const NO_FLAGS = { portal_admin: false, cluster_admin: false };

// Onboarding approves the user's cluster account; with one cluster registered the backend picks it.
export function CreateUserDialog() {
  const ability = useAbility();
  const [open, setOpen] = React.useState(false);
  const [flags, setFlags] = React.useState(NO_FLAGS);
  const create = useCreateUser();
  const clusters = useClusters(open && ability.can("read", "Cluster")).data ?? [];
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
      onSubmit={(form, close) =>
        create.mutate(
          {
            ...flags,
            email: text(form, "email"),
            first_name: text(form, "first_name"),
            last_name: text(form, "last_name"),
            username: portalOnly ? undefined : text(form, "username") || undefined,
            compute_cluster_id: portalOnly ? undefined : text(form, "cluster") || undefined,
          },
          toastOnSuccess(`Created ${text(form, "email")}`, close),
        )
      }
    >
      <Field label="Email" name="email" type="email" required autoFocus />
      <div className="grid grid-cols-2 gap-3">
        <Field label="First name" name="first_name" />
        <Field label="Last name" name="last_name" />
      </div>
      {ability.can("write", "Role") ? (
        <div className="flex gap-4 text-sm">
          {flag("portal_admin", "Portal admin")}
          {flag("cluster_admin", "Cluster admin")}
        </div>
      ) : null}
      {portalOnly ? null : (
        <>
          <Field label="Cluster username" name="username" placeholder="Generated when blank" />
          {clusters.length > 1 ? (
            <SelectField
              label="Cluster"
              name="cluster"
              options={named(clusters)}
            />
          ) : null}
        </>
      )}
    </FormDialog>
  );
}
