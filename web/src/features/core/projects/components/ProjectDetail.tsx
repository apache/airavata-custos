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

import type { ProjectResponse, ProjectStatus } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { useBreadcrumbLabel } from "@/shared/layout/BreadcrumbLabelsProvider";
import { QueryErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { TabsRouter } from "@/shared/ui/TabsRouter";
import { Button } from "@/shared/ui/button";
import { STATUSES, statusLabel } from "@/shared/ui/FormDialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { useRouter } from "next/navigation";
import { useDeleteProject, useProject, useUpdateProjectStatus } from "../queries";
import { ProjectAllocationsTab } from "./ProjectAllocationsTab";
import { ProjectDetailHeader } from "./ProjectDetailHeader";
import { ProjectDialog } from "./ProjectDialog";
import { ProjectMembersTab } from "./ProjectMembersTab";
import { ProjectOverviewTab } from "./ProjectOverviewTab";

export type ProjectDetailProps = {
  projectId: string;
};

function DeleteProjectButton({ project }: { project: ProjectResponse }) {
  const router = useRouter();
  const remove = useDeleteProject();
  return (
    <Button
      variant="destructive"
      disabled={remove.isPending}
      onClick={() =>
        confirmToast(`Delete project "${project.title}"? This cannot be undone.`, "Delete", () =>
          remove.mutate(
            { path: { id: project.id } },
            toastOnSuccess("Project deleted", () => router.push("/projects")),
          ),
        )
      }
    >
      Delete project
    </Button>
  );
}

function ProjectStatusSelect({ project }: { project: ProjectResponse }) {
  const update = useUpdateProjectStatus();
  function apply(status: ProjectStatus) {
    update.mutate(
      { path: { id: project.id }, body: { status } },
      toastOnSuccess(`Project marked ${statusLabel(status).toLowerCase()}`),
    );
  }
  return (
    <Select
      value={project.status}
      disabled={update.isPending}
      onValueChange={(value) => {
        const status = STATUSES.find((s) => s === value);
        if (!status || status === project.status) return;
        if (status === "DELETED") {
          confirmToast(`Mark project "${project.title}" as deleted?`, "Mark deleted", () =>
            apply(status),
          );
        } else apply(status);
      }}
    >
      <SelectTrigger aria-label="Project status" className="h-9 w-32 px-3">
        <SelectValue>{statusLabel}</SelectValue>
      </SelectTrigger>
      <SelectContent>
        {STATUSES.map((status) => (
          <SelectItem key={status} value={status}>
            {statusLabel(status)}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

export function ProjectDetail({ projectId }: ProjectDetailProps) {
  const ability = useAbility();
  const projectQuery = useProject(projectId);
  useBreadcrumbLabel(projectId, projectQuery.data?.title);

  if (projectQuery.isLoading) return <CardSkeleton />;
  if (projectQuery.error) {
    return (
      <QueryErrorState error={projectQuery.error} what="project" onRetry={() => projectQuery.refetch()} />
    );
  }
  const project = projectQuery.data;
  if (!project) return null;

  const canManage = ability.can("write", "Project") && project.origination === "internal";

  return (
    <section className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <ProjectDetailHeader project={project} />
        {canManage ? (
          <div className="flex gap-2">
            <ProjectStatusSelect project={project} />
            <ProjectDialog project={project} />
            <DeleteProjectButton project={project} />
          </div>
        ) : null}
      </div>
      <TabsRouter
        defaultValue="overview"
        tabs={[
          {
            value: "overview",
            label: "Overview",
            content: <ProjectOverviewTab project={project} />,
          },
          {
            value: "allocations",
            label: "Allocations",
            content: (
              <ProjectAllocationsTab
                projectId={projectId}
                canCreate={ability.can("write", "Allocation") && project.origination === "internal"}
              />
            ),
          },
          {
            value: "members",
            label: "Members",
            content: <ProjectMembersTab projectId={projectId} canManage={canManage} />,
          },
        ].filter((tab) => tab.value !== "members" || ability.can("read", "Project"))}
      />
    </section>
  );
}
