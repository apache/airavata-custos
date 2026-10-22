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

import type { ProjectResponse } from "@/generated/core/zod.gen";

type PI = Pick<ProjectResponse, "project_pi_id" | "project_pi_display_name" | "project_pi_email">;

export function ProjectPI({ project }: { project: PI }) {
  if (!project.project_pi_display_name)
    return <span className="font-mono">{project.project_pi_id}</span>;
  return (
    <div className="leading-tight">
      <div>{project.project_pi_display_name}</div>
      {project.project_pi_email ? (
        <div className="text-xs text-muted-foreground">{project.project_pi_email}</div>
      ) : null}
    </div>
  );
}
