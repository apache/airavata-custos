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

import type { ProjectResponse } from "@/generated/core/types.gen";
import { Button } from "@/shared/ui/button";
import { Field, FormDialog, text } from "@/shared/ui/FormDialog";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { useCreateProject, useUpdateProject } from "../queries";

// Creates a project, or edits the given one.
export function ProjectDialog({ project }: { project?: ProjectResponse }) {
  const create = useCreateProject();
  const update = useUpdateProject();
  const label = project ? "Edit project" : "Create project";

  return (
    <FormDialog
      trigger={<Button variant={project ? "outline" : "default"}>{label}</Button>}
      title={label}
      submitLabel={label}
      isPending={(project ? update : create).isPending}
      onSubmit={(form, close) => {
        const body = { title: text(form, "title"), project_pi_id: text(form, "pi") };
        const callbacks = toastOnSuccess(project ? "Project updated" : "Project created", close);
        if (project) update.mutate({ path: { id: project.id }, body }, callbacks);
        else create.mutate({ body }, callbacks);
      }}
    >
      <Field label="Title" name="title" defaultValue={project?.title} required />
      <Field label="PI user ID" name="pi" defaultValue={project?.project_pi_id} required />
    </FormDialog>
  );
}
