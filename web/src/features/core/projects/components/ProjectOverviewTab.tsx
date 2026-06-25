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

import type { Project } from "../schemas";

export type ProjectOverviewTabProps = {
  project: Project;
};

export function ProjectOverviewTab({ project }: ProjectOverviewTabProps) {
  return (
    <dl className="grid gap-x-8 gap-y-3 sm:grid-cols-[max-content_1fr] text-sm">
      <dt className="text-muted-foreground">Project ID</dt>
      <dd className="font-mono text-foreground">{project.id}</dd>

      <dt className="text-muted-foreground">Originated ID</dt>
      <dd className="font-mono text-foreground">{project.originated_id || "—"}</dd>

      <dt className="text-muted-foreground">Title</dt>
      <dd className="text-foreground">{project.title}</dd>

      <dt className="text-muted-foreground">Origination</dt>
      <dd className="text-foreground">{project.origination}</dd>

      <dt className="text-muted-foreground">PI</dt>
      <dd className="text-foreground">
        {project.project_pi_display_name ? (
          <>
            {project.project_pi_display_name}
            {project.project_pi_email ? (
              <span className="ml-2 text-xs text-muted-foreground">{project.project_pi_email}</span>
            ) : null}
          </>
        ) : (
          <span className="font-mono">{project.project_pi_id}</span>
        )}
      </dd>

      <dt className="text-muted-foreground">Status</dt>
      <dd className="text-foreground">{project.status}</dd>

      <dt className="text-muted-foreground">Created</dt>
      <dd className="text-foreground">{project.created_time}</dd>
    </dl>
  );
}
