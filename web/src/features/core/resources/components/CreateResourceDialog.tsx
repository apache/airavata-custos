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

import { Field, FormDialog, named, num, SelectField, text } from "@/shared/ui/FormDialog";
import type { ComputeCluster } from "@/generated/core/types.gen";
import { Button } from "@/shared/ui/button";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { useCreateResource } from "../queries";

export function CreateResourceDialog({ clusters }: { clusters: ComputeCluster[] }) {
  const create = useCreateResource();
  return (
    <FormDialog
      trigger={<Button disabled={clusters.length === 0}>+ New resource</Button>}
      title="New resource"
      description="A partition on a cluster that allocations can be attached to."
      submitLabel="Create"
      isPending={create.isPending}
      onSubmit={(form, close) =>
        create.mutate(
          {
            body: {
              compute_cluster_id: text(form, "cluster"),
              name: text(form, "name"),
              resource_type: text(form, "type"),
              resource_amount: num(form, "amount"),
            },
          },
          toastOnSuccess(`Resource ${text(form, "name")} created`, close),
        )
      }
    >
      <SelectField
        label="Cluster"
        name="cluster"
        options={named(clusters)}
      />
      <Field label="Name" name="name" placeholder="e.g. gpu-01" required />
      <Field label="Type (TRES)" name="type" placeholder="e.g. cpu, gres/gpu" required />
      <Field label="Amount" name="amount" type="number" min={0} required />
    </FormDialog>
  );
}
