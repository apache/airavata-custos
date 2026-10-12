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

import { useAllocationsByProject } from "@/features/core/allocations/queries";
import type { ProjectResponse, ProjectStatus } from "@/generated/core/types.gen";
import { isGone, useOnGone } from "@/shared/api/useOnGone";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { useBreadcrumbLabel } from "@/shared/layout/BreadcrumbLabelsProvider";
import { ROUTE_LABELS } from "@/shared/layout/nav";
import { QueryErrorState } from "@/shared/ui/ErrorState";
import { STATUSES, statusLabel } from "@/shared/ui/FormDialog";
import { DetailLoading } from "@/shared/ui/Loading";
import { TabsRouter } from "@/shared/ui/TabsRouter";
import { Button } from "@/shared/ui/button";
import { captureOrigin, keepFocus, originTarget, useLeaveTo } from "@/shared/ui/focus";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { confirmToast, dismissConfirm, toastOnSuccess } from "@/shared/ui/sonner";
import {
  useDeleteProject,
  useProject,
  useRefreshProjectLists,
  useUpdateProjectStatus,
} from "../queries";
import { ProjectAllocationsTab } from "./ProjectAllocationsTab";
import { ProjectDetailHeader } from "./ProjectDetailHeader";
import { ProjectDialog } from "./ProjectDialog";
import { ProjectMembersTab } from "./ProjectMembersTab";
import { ProjectOverviewTab } from "./ProjectOverviewTab";

export type ProjectDetailProps = {
  projectId: string;
};

type ProjectProps = { project: ProjectResponse };

function DeleteProjectButton({ project }: ProjectProps) {
  const leaveTo = useLeaveTo();
  const remove = useDeleteProject();
  // Shown only once the full list is known empty; master refuses the delete while allocations remain.
  if (useAllocationsByProject(project.id).data?.length !== 0) return null;
  return (
    <Button
      variant="destructive"
      pending={remove.isPending}
      onClick={() => {
        const origin = captureOrigin();
        confirmToast(`Delete project "${project.title}"? This cannot be undone.`, "Delete", () => {
          const deleting = remove.mutateAsync({ path: { id: project.id } });
          // A 409 removes this button once the allocations refetch.
          deleting.catch(() => keepFocus(() => originTarget(origin)));
          toastOnSuccess(deleting, "Project deleted", () => leaveTo("/projects"));
        });
      }}
    >
      Delete project
    </Button>
  );
}

function ProjectStatusSelect({ project }: ProjectProps) {
  const update = useUpdateProjectStatus();
  function apply(status: ProjectStatus) {
    toastOnSuccess(
      update.mutateAsync({ path: { id: project.id }, body: { status } }),
      `Project marked ${statusLabel(status).toLowerCase()}`,
    );
  }
  return (
    <Select
      value={project.status}
      // readOnly, not disabled: a disabled trigger drops focus to <body> mid-request.
      readOnly={update.isPending}
      onValueChange={(value) => {
        dismissConfirm();
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
  useOnGone(isGone(projectQuery.error), useRefreshProjectLists());
  useBreadcrumbLabel(
    projectId,
    projectQuery.data?.title ?? (projectQuery.error ? "Project" : undefined),
  );

  if (projectQuery.isLoading) return <DetailLoading title={ROUTE_LABELS["/projects"]} />;
  if (projectQuery.error) {
    return (
      <QueryErrorState
        error={projectQuery.error}
        what="project"
        onRetry={() => projectQuery.refetch()}
        as="h1"
        page
      />
    );
  }
  const project = projectQuery.data;
  if (!project) return null;

  const canManage = ability.can("write", "Project");

  return (
    <section className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <ProjectDetailHeader project={project} />
        {canManage ? (
          <div className="flex flex-wrap gap-2">
            <ProjectStatusSelect project={project} />
            <ProjectDialog project={project} />
            {ability.can("read", "Allocation") ? <DeleteProjectButton project={project} /> : null}
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
                canCreate={ability.can("write", "Allocation")}
                canReadAll={ability.can("read", "Allocation")}
                canReadClusters={ability.can("read", "Cluster")}
              />
            ),
          },
          {
            value: "members",
            label: "Members",
            content: (
              <ProjectMembersTab
                projectId={projectId}
                canManage={canManage}
                canReadAllocations={ability.can("read", "Allocation")}
              />
            ),
          },
        ].filter((tab) => tab.value !== "members" || ability.can("read", "Project"))}
      />
    </section>
  );
}
