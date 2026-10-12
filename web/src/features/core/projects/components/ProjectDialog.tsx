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
import { normalizeQueryId } from "@/shared/layout/nav";
import { Field, FormDialog, SelectField, edited, idText, text } from "@/shared/ui/FormDialog";
import { Button } from "@/shared/ui/button";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { idInput } from "@/shared/users-admin/idInput";
import { useState } from "react";
import { useCreateProject, useForgetProject, useProjectRecord, useUpdateProject } from "../queries";
import { roleLabel } from "./ProjectMembersTab";

// The roles PUT /projects/{id} accepts in previous_pi_role; MEMBER removes the project role.
const PREVIOUS_PI_ROLES = ["CO_PI", "ALLOCATION_MANAGER", "MEMBER"].map((value) => ({
  value,
  label: value === "MEMBER" ? "No project role" : roleLabel(value),
}));

// The PI field; the previous PI's new role is asked only once the PI changes.
function PiFields({ current }: { current?: string }) {
  const [pi, setPi] = useState(current ?? "");
  return (
    <>
      <Field
        label="PI user ID"
        name="pi"
        defaultValue={current}
        onChange={(e) => setPi(e.target.value)}
        required
        {...idInput("user")}
      />
      {current !== undefined && (normalizeQueryId(pi) ?? "") !== current ? (
        <SelectField
          label="Previous PI becomes"
          name="previous_pi_role"
          options={PREVIOUS_PI_ROLES}
          defaultValue="CO_PI"
          description="No project role: they stay listed only through allocation memberships."
        />
      ) : null}
    </>
  );
}

// Creates a project, or edits the given one.
export function ProjectDialog({ project: opened }: { project?: ProjectResponse }) {
  const create = useCreateProject();
  const update = useUpdateProject();
  const [open, setOpen] = useState(false);
  const record = useProjectRecord(open ? opened?.id : undefined);
  const forget = useForgetProject(opened?.id);
  const project = opened && (record.data ?? opened);
  const label = project ? "Edit project" : "Create project";

  return (
    <FormDialog
      open={open}
      onOpenChange={setOpen}
      trigger={<Button variant={project ? "outline" : "default"}>{label}</Button>}
      title={label}
      submitLabel={label}
      isPending={(project ? update : create).isPending}
      record={opened ? record : undefined}
      gone="This project no longer exists"
      // The page swaps to its unavailable state, whose heading takes the focus.
      onGone={forget}
      onSubmit={(form, close) => {
        const fields = { title: text(form, "title"), project_pi_id: idText(form, "pi") };
        if (project && !edited(project, fields)) return close();
        const body =
          project && fields.project_pi_id !== project.project_pi_id
            ? { ...fields, previous_pi_role: text(form, "previous_pi_role") }
            : fields;
        toastOnSuccess(
          project
            ? update.mutateAsync({ path: { id: project.id }, body })
            : create.mutateAsync({ body: fields }),
          project ? "Project updated" : "Project created",
          close,
        );
      }}
    >
      <Field
        label="Title"
        name="title"
        defaultValue={project?.title}
        maxLength={255}
        required
        title="Enter a title, not only spaces."
      />
      <PiFields current={project?.project_pi_id} />
    </FormDialog>
  );
}
