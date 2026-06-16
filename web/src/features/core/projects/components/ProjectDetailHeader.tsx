"use client";

import { Building2, Calendar, UserSquare } from "lucide-react";
import { MetaItem, MetaRow } from "@/shared/ui/MetaRow";
import type { Project } from "../schemas";

export type ProjectDetailHeaderProps = {
  project: Project;
};

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  } catch {
    return iso;
  }
}

export function ProjectDetailHeader({ project }: ProjectDetailHeaderProps) {
  const tone: "success" | "warning" | "danger" =
    project.status === "ACTIVE" ? "success" : project.status === "INACTIVE" ? "warning" : "danger";
  return (
    <header className="space-y-4">
      <h1 className="font-display text-[28px] font-bold leading-tight text-foreground">
        {project.title}
      </h1>
      <MetaRow>
        <MetaItem variant="status" tone={tone} value={project.status} />
        <MetaItem icon={UserSquare} label="PI" value={project.project_pi_display_name ?? project.project_pi_id} />
        <MetaItem icon={Building2} label="Origination" value={project.origination} />
        <MetaItem icon={Calendar} label="Created" value={formatDate(project.created_time)} />
      </MetaRow>
    </header>
  );
}
