import { ProjectDetail } from "@/features/core/projects/components/ProjectDetail";

export default async function ProjectDetailPage(props: {
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = await props.params;
  return <ProjectDetail projectId={projectId} />;
}
