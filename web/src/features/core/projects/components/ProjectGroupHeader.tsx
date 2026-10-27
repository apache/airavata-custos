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

import { ChevronRight } from "lucide-react";
import Link from "next/link";
import { useProject } from "../queries";
import { ProjectRef } from "./ProjectRef";
import { ProjectStatusIcon } from "./ProjectStatusIcon";

// Header of a project's card for rows that carry only its id; fetched once and cached per project.
export function ProjectGroupHeader({ id }: { id: string }) {
  const { data: project, isLoading } = useProject(id);
  if (!project) return <>{isLoading ? "—" : "Unknown project"}</>;
  return (
    <div className="flex items-center gap-3">
      <ProjectStatusIcon status={project.status} />
      <div className="flex min-w-0 flex-1 items-center gap-2">
        <span className="truncate text-xs font-semibold uppercase tracking-wide text-foreground">
          {project.title}
        </span>
        <ProjectRef project={project} />
      </div>
      <Link
        href={`/projects/${id}`}
        className="inline-flex shrink-0 items-center gap-0.5 text-xs font-medium text-muted-foreground hover:text-foreground"
      >
        Go to project
        <ChevronRight className="size-3.5" aria-hidden />
      </Link>
    </div>
  );
}
