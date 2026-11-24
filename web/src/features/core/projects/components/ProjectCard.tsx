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

import { AllocationList } from "@/features/core/allocations/components/AllocationList";
import type { ProjectResponse } from "@/generated/core/zod.gen";
import { cn } from "@/lib/utils";
import { Badge } from "@/shared/ui/badge";
import { ChevronRight } from "lucide-react";
import Link from "next/link";
import { ProjectPI } from "./ProjectPI";

// Colours per origination, so ACCESS, NAIRR and internal projects read apart at a glance.
const ORIGINATION_COLOR: Record<string, { outline: string; pill: string }> = {
  access: {
    outline: "border-custos-blue-300 border-l-custos-blue-500",
    pill: "border-custos-blue-300 bg-custos-blue-50 text-custos-blue-700",
  },
  nairr: {
    outline: "border-custos-green-400 border-l-custos-green-600",
    pill: "border-custos-green-400 bg-custos-green-50 text-custos-green-800",
  },
  internal: {
    outline: "border-custos-amber-500/60 border-l-custos-amber-600",
    pill: "border-custos-amber-500/60 bg-custos-amber-50 text-custos-amber-800",
  },
};

export function ProjectCard({ project }: { project: ProjectResponse }) {
  const color = ORIGINATION_COLOR[project.origination.toLowerCase()];
  return (
    <section
      className={cn(
        "relative overflow-hidden rounded-lg border border-l-4 bg-card",
        color?.outline ?? "border-foreground/20",
      )}
    >
      <Link
        href={`/projects/${project.id}`}
        className="group flex items-center gap-3 border-b bg-muted px-4 py-2 hover:bg-brand/10"
      >
        <div className="min-w-0 flex-1">
          <h2 className="truncate font-display text-sm font-semibold leading-tight text-foreground">
            {project.ref.slice(project.ref.indexOf("-") + 1)}: {project.title}
          </h2>
          <Badge variant="outline" className={cn("mt-1", color?.pill ?? "bg-card")}>
            {project.origination}
          </Badge>
        </div>
        <span className="inline-flex shrink-0 items-center gap-0.5 text-xs font-medium text-muted-foreground group-hover:text-foreground">
          Go to project
          <ChevronRight className="size-3.5" aria-hidden />
        </span>
      </Link>
      <dl className="grid grid-cols-[max-content_1fr] gap-x-6 gap-y-1 px-4 py-2 text-sm">
        <dt className="text-muted-foreground">PI</dt>
        <dd>
          <ProjectPI project={project} />
        </dd>
        <dt className="text-muted-foreground">Allocations</dt>
        <dd>
          <AllocationList project={project} />
        </dd>
      </dl>
      {project.status === "ACTIVE" ? null : (
        <div className="pointer-events-none absolute inset-0 grid select-none place-items-center bg-[repeating-linear-gradient(-45deg,transparent_0_3px,rgb(0_0_0/0.06)_3px_4px)] font-display text-4xl font-bold tracking-[0.3em] text-foreground/15">
          {project.status}
        </div>
      )}
    </section>
  );
}
