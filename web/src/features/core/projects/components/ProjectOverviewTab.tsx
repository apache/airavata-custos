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
import { ProjectPI } from "./ProjectPI";

export type ProjectOverviewTabProps = {
  project: ProjectResponse;
};

export function ProjectOverviewTab({ project }: ProjectOverviewTabProps) {
  return (
    <dl className="grid gap-x-8 gap-y-3 rounded-lg border border-border bg-muted/40 p-4 text-sm sm:grid-cols-[max-content_1fr]">
      <dt className="text-muted-foreground">Ref</dt>
      <dd className="text-foreground">{project.ref}</dd>

      <dt className="text-muted-foreground">PI</dt>
      <dd className="text-foreground">
        <ProjectPI project={project} />
      </dd>

      <dt className="text-muted-foreground">Allocations</dt>
      <dd className="text-foreground">
        <AllocationList project={project} />
      </dd>
    </dl>
  );
}
