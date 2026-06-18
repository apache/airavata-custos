import type { ReactNode } from "react";
import { Breadcrumbs } from "@/shared/ui/Breadcrumbs";

export default async function ProjectDetailLayout(props: {
  children: ReactNode;
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = await props.params;
  return (
    <div className="space-y-6">
      <Breadcrumbs items={[{ label: "Projects", href: "/projects" }, { label: projectId }]} />
      {props.children}
    </div>
  );
}
