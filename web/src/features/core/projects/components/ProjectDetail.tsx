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

import { CreateAllocationDialog } from "@/features/core/allocations/components/CreateAllocationDialog";
import { requires, usePermissions } from "@/shared/casl/permissions";
import { useBreadcrumbLabel } from "@/shared/layout/BreadcrumbLabelsProvider";
import { DeleteButton } from "@/shared/ui/ConfirmDialog";
import { ErrorState } from "@/shared/ui/ErrorState";
import { GuardedButton } from "@/shared/ui/GuardedButton";
import { CardSkeleton } from "@/shared/ui/Loading";
import * as React from "react";
import { useDeleteProject, useProject } from "../queries";
import { ProjectDetailHeader } from "./ProjectDetailHeader";
import { ProjectMembersTab } from "./ProjectMembersTab";
import { ProjectOverviewTab } from "./ProjectOverviewTab";

export type ProjectDetailProps = {
  projectId: string;
};

export function ProjectDetail({ projectId }: ProjectDetailProps) {
  const can = usePermissions();
  const projectQuery = useProject(projectId);
  const [addingAllocation, setAddingAllocation] = React.useState(false);
  const remove = useDeleteProject();
  useBreadcrumbLabel(projectId, projectQuery.data?.title);

  if (projectQuery.isLoading) return <CardSkeleton />;
  if (projectQuery.error) {
    return (
      <ErrorState message={projectQuery.error.message} onRetry={() => projectQuery.refetch()} />
    );
  }
  const project = projectQuery.data;
  if (!project) return null;

  return (
    <section className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <ProjectDetailHeader project={project} />
        <DeleteButton
          noun="Project"
          id={project.id}
          name={project.title}
          mutation={remove}
          redirectTo="/projects"
          allowed={can.projectsWrite}
          reason={requires("projectsWrite")}
        >
          A project can be deleted only once it has no allocations. This cannot be undone.
        </DeleteButton>
      </div>
      <ProjectOverviewTab
        project={project}
        addAllocation={
          <GuardedButton
            variant="link"
            className="h-auto border-0 p-0"
            allowed={can.allocationsWrite && can.clustersRead}
            reason={requires("allocationsWrite", "clustersRead")}
            onClick={() => setAddingAllocation(true)}
          >
            Add allocation
          </GuardedButton>
        }
      />
      <ProjectMembersTab
        projectId={project.id}
        canEditRoles={can.projectsWrite}
        canEditAccess={can.allocationsWrite}
        canSearchUsers={can.usersRead}
      />
      {addingAllocation ? (
        <CreateAllocationDialog onOpenChange={setAddingAllocation} project={project} />
      ) : null}
    </section>
  );
}
