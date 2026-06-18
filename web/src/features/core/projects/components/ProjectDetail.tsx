"use client";

import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { TabsRouter } from "@/shared/ui/TabsRouter";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { useProject } from "../queries";
import { ProjectAllocationsTab } from "./ProjectAllocationsTab";
import { ProjectDetailHeader } from "./ProjectDetailHeader";
import { ProjectMembersTab } from "./ProjectMembersTab";
import { ProjectOverviewTab } from "./ProjectOverviewTab";

export type ProjectDetailProps = {
  projectId: string;
};

export function ProjectDetail({ projectId }: ProjectDetailProps) {
  const ability = useAbility();
  const projectQuery = useProject(projectId);

  if (projectQuery.isLoading) return <CardSkeleton />;
  if (projectQuery.error) {
    return (
      <ErrorState
        message={(projectQuery.error as Error).message}
        onRetry={() => projectQuery.refetch()}
      />
    );
  }
  const project = projectQuery.data;
  if (!project) return null;

  const canManage = ability.can("manage", "Project");

  return (
    <section className="space-y-6">
      <ProjectDetailHeader project={project} />
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
            content: <ProjectAllocationsTab projectId={project.id} />,
          },
          {
            value: "members",
            label: "Members",
            content: <ProjectMembersTab projectId={project.id} canManage={canManage} />,
          },
        ]}
      />
    </section>
  );
}
