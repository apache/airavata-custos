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
import { toastOnSuccess } from "@/shared/ui/sonner";
import { useClusters } from "@/features/core/clusters/queries";
import { Button } from "@/shared/ui/button";
import { useCreateAllocation } from "../queries";
import { Field, FormDialog, isoDate, named, num, SelectField, text } from "@/shared/ui/FormDialog";

export function CreateAllocationDialog({ projectId }: { projectId: string }) {
  const [open, setOpen] = React.useState(false);
  const clusters = useClusters(open);
  const create = useCreateAllocation();

  return (
    <FormDialog
      open={open}
      onOpenChange={setOpen}
      trigger={<Button size="sm">+ New allocation</Button>}
      title="New allocation"
      submitLabel="Create allocation"
      isPending={create.isPending}
      onSubmit={(form, close) =>
        create.mutate(
          {
            body: {
              project_id: projectId,
              name: text(form, "name"),
              compute_cluster_id: text(form, "cluster"),
              initial_su_amount: num(form, "su"),
              start_time: isoDate(form, "start"),
              end_time: isoDate(form, "end"),
            },
          },
          toastOnSuccess("Allocation created", close),
        )
      }
    >
      <Field label="Name" name="name" required />
      <SelectField
        label="Cluster"
        name="cluster"
        options={named(clusters.data)}
      />
      <Field label="Initial SUs" name="su" type="number" min={0} required />
      <Field label="Start" name="start" type="date" required />
      <Field label="End" name="end" type="date" required />
    </FormDialog>
  );
}
