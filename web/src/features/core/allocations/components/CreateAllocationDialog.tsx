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
import { useProjects } from "@/features/core/projects/queries";
import { listOf } from "@/shared/api/listOf";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Field, FormDialog, idText, isoDate, named, num, text } from "@/shared/ui/FormDialog";
import { Button } from "@/shared/ui/button";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { idInput } from "@/shared/users-admin/idInput";
import * as React from "react";
import { useCreateAllocation } from "../queries";
import { dayAfter } from "./AllocationStatusLabel";
import { ListSelect } from "./ListSelect";

const byLabel = (options: Array<{ value: string; label: string }>) =>
  options.sort((a, b) => a.label.localeCompare(b.label));

// Mounted with the open dialog, so the picker's list loads only then.
function ProjectSelect() {
  // 200 is the backend's page-size ceiling.
  const projects = useProjects({ limit: 200 }, true);
  return (
    <ListSelect
      label="Project"
      name="project"
      things="projects"
      query={projects}
      options={(data) => byLabel(listOf(data?.items).map((p) => ({ value: p.id, label: p.title })))}
    />
  );
}

function ClusterSelect() {
  return (
    <ListSelect
      label="Cluster"
      name="cluster"
      things="clusters"
      query={useClusters()}
      options={(data) => byLabel(named(data))}
    />
  );
}

// Without projectId the dialog asks for the project; a record the caller cannot list is entered by ID.
export function CreateAllocationDialog({ projectId }: { projectId?: string }) {
  const ability = useAbility();
  const [open, setOpen] = React.useState(false);
  const [start, setStart] = React.useState("");
  const pickProject = ability.can("read", "Project");
  const pickCluster = ability.can("read", "Cluster");
  const create = useCreateAllocation({
    project: !projectId && !pickProject,
    cluster: !pickCluster,
  });

  return (
    <FormDialog
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        setStart("");
      }}
      trigger={<Button size="sm">+ New allocation</Button>}
      title="New allocation"
      submitLabel="Create allocation"
      isPending={create.isPending}
      onSubmit={(form, close) =>
        toastOnSuccess(
          create.mutateAsync({
            body: {
              project_id: projectId ?? idText(form, "project"),
              name: text(form, "name"),
              compute_cluster_id: idText(form, "cluster"),
              initial_su_amount: num(form, "su"),
              start_time: isoDate(form, "start"),
              end_time: isoDate(form, "end"),
            },
          }),
          "Allocation created",
          close,
        )
      }
    >
      {projectId ? null : pickProject ? (
        <ProjectSelect />
      ) : (
        <Field label="Project ID" name="project" required {...idInput("project")} />
      )}
      <Field label="Name" name="name" maxLength={255} required />
      {pickCluster ? (
        <ClusterSelect />
      ) : (
        <Field label="Cluster ID" name="cluster" required {...idInput("cluster")} />
      )}
      <Field
        label="Initial SUs"
        name="su"
        type="number"
        min={0}
        max={Number.MAX_SAFE_INTEGER}
        step="any"
        required
      />
      <Field
        label="Start"
        name="start"
        type="date"
        required
        onChange={(e) => setStart(e.currentTarget.value)}
      />
      <Field label="End" name="end" type="date" min={dayAfter(start)} required />
    </FormDialog>
  );
}
