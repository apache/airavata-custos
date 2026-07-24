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
      <div className="flex flex-wrap items-center gap-3">
        <MetaItem variant="status" tone={tone} value={project.status} className="py-1.5" />
        <MetaRow>
          <MetaItem icon={UserSquare} label="PI" value={project.project_pi_display_name ?? project.project_pi_id} />
          <MetaItem icon={Building2} label="Origination" value={project.origination} />
          <MetaItem icon={Calendar} label="Created" value={formatDate(project.created_time)} />
        </MetaRow>
      </div>
    </header>
  );
}
