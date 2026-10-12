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

import { useClusters } from "@/features/core/clusters/queries";
import { UserPicker, isAssignableUser } from "@/features/core/users/components/UserPicker";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Field, FormDialog, SelectField, idText, named, text } from "@/shared/ui/FormDialog";
import { Button } from "@/shared/ui/button";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { idInput } from "@/shared/users-admin/idInput";
import * as React from "react";
import { useCreateClusterAccount } from "../queries";

export function CreateClusterAccountDialog() {
  const [open, setOpen] = React.useState(false);
  const ability = useAbility();
  // Without Cluster or User read, or when the clusters fail to load, the record's id is typed.
  const canReadClusters = ability.can("read", "Cluster");
  const canPickUser = ability.can("read", "User");
  const [userId, setUserId] = React.useState("");
  const clustersQuery = useClusters(open && canReadClusters);
  const canPickCluster = canReadClusters && !clustersQuery.isError;
  const clusters = clustersQuery.data ?? [];
  const create = useCreateClusterAccount();

  return (
    <FormDialog
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (next) setUserId("");
      }}
      trigger={<Button>+ Add account</Button>}
      title="Add cluster account"
      description="An account you add is approved and recorded with you as the reviewer."
      submitLabel="Add account"
      isPending={create.isPending}
      submitDisabled={canPickUser && !userId}
      onSubmit={(form, close) => {
        const clusterId = idText(form, "cluster");
        const name = clusters.find((c) => c.id === clusterId)?.name;
        toastOnSuccess(
          create.mutateAsync({
            body: {
              compute_cluster_id: clusterId,
              user_id: canPickUser ? userId : idText(form, "user"),
              local_username: text(form, "username"),
            },
          }),
          name ? `Cluster account added on ${name}` : "Cluster account added",
          close,
        );
      }}
    >
      {canPickCluster ? (
        <SelectField label="Cluster" name="cluster" options={named(clusters)} />
      ) : (
        <Field label="Cluster ID" name="cluster" required {...idInput("cluster")} />
      )}
      {canPickUser ? (
        <UserPicker
          id="cluster-account-user"
          label="User"
          enabled={open}
          single
          selected={new Set([userId])}
          onToggle={setUserId}
          eligible={isAssignableUser}
        />
      ) : (
        <Field label="User ID" name="user" required {...idInput("user")} />
      )}
      <Field
        label="Local username"
        name="username"
        className="font-mono"
        maxLength={32}
        required
        title="Enter a username, not only spaces."
      />
    </FormDialog>
  );
}
