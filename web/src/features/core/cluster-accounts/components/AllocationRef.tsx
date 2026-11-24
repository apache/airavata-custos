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

import { ProjectRef } from "@/features/core/projects/components/ProjectRef";
import { useProject } from "@/features/core/projects/queries";
import type { ClusterAccountResponse } from "@/generated/core/zod.gen";

export function AllocationRef({
  account,
  withTitle = false,
}: {
  account: ClusterAccountResponse;
  withTitle?: boolean;
}) {
  const { data: project } = useProject(account.project_id);
  if (!account.allocation_id) return <span className="text-xs text-muted-foreground">None</span>;
  return (
    <div className="min-w-40 whitespace-normal">
      <div className="flex flex-wrap items-center gap-x-2 gap-y-0.5">
        <span className="font-medium">{account.allocation_name}</span>
        {project ? <ProjectRef project={project} /> : null}
      </div>
      {withTitle ? (
        <div title={project?.title} className="max-w-[32ch] truncate text-xs text-muted-foreground">
          {project?.title}
        </div>
      ) : null}
    </div>
  );
}
