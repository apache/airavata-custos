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
import { Button } from "@/shared/ui/button";
import { Field, FormDialog, named, SelectField, text } from "@/shared/ui/FormDialog";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { useCreateClusterAccount } from "../queries";

export function CreateClusterAccountDialog() {
  const [open, setOpen] = React.useState(false);
  const clusters = useClusters(open).data ?? [];
  const create = useCreateClusterAccount();

  return (
    <FormDialog
      open={open}
      onOpenChange={setOpen}
      trigger={<Button>+ Add account</Button>}
      title="Add cluster account"
      description="An account you add is approved and recorded with you as the reviewer."
      submitLabel="Add account"
      isPending={create.isPending}
      onSubmit={(form, close) => {
        const clusterId = text(form, "cluster");
        const name = clusters.find((c) => c.id === clusterId)?.name ?? clusterId;
        create.mutate(
          {
            body: {
              compute_cluster_id: clusterId,
              user_id: text(form, "user"),
              local_username: text(form, "username"),
            },
          },
          toastOnSuccess(`Cluster account added on ${name}`, close),
        );
      }}
    >
      <SelectField
        label="Cluster"
        name="cluster"
        options={named(clusters)}
      />
      <Field label="User ID" name="user" required />
      <Field label="Local username" name="username" className="font-mono" required />
    </FormDialog>
  );
}
