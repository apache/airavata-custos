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
import type { ProjectResponse } from "@/generated/core/zod.gen";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { StatusIcon } from "@/shared/ui/StatusIcon";
import { Button } from "@/shared/ui/button";
import { CheckCheck, CircleHelp, Server } from "lucide-react";
import Link from "next/link";

export type ApprovalGroupHeaderProps = {
  labelId: string;
  allocationId?: string | undefined;
  allocationName?: string | undefined;
  project?: ProjectResponse | undefined;
  cluster: string;
  count: number;
  onApproveAll?: (() => void) | undefined;
};

export function ApprovalGroupHeader({
  labelId,
  allocationId,
  allocationName,
  project,
  cluster,
  count,
  onApproveAll,
}: ApprovalGroupHeaderProps) {
  const pill = <StatusBadge variant="warning" label={`${count} waiting`} />;

  if (!allocationId) {
    return (
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
        <StatusIcon icon={CircleHelp} active={false} label="No allocation" />
        <span id={labelId} className="text-sm font-semibold text-muted-foreground">
          No allocation
        </span>
        <span className="min-w-0 flex-1 basis-48 text-xs text-muted-foreground">
          Not linked to an allocation. Check who this person is before approving.
        </span>
        {pill}
      </div>
    );
  }

  return (
    <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
      <StatusIcon
        icon={Server}
        active={project?.status === "ACTIVE"}
        label={project ? `Project ${project.status.toLowerCase()}` : "Allocation"}
      />
      <Link
        id={labelId}
        href={`/allocations/${allocationId}`}
        className="text-sm font-semibold text-foreground hover:underline"
      >
        {allocationName ?? allocationId}
      </Link>
      {project ? <ProjectRef project={project} /> : null}
      <span
        title={project?.title}
        className="min-w-0 flex-1 basis-48 truncate text-xs text-muted-foreground"
      >
        {project?.title}
      </span>
      <span className="shrink-0 text-xs text-muted-foreground">
        {project ? `PI ${project.project_pi_display_name ?? project.project_pi_id} · ` : null}
        <span className="font-mono">{cluster}</span>
      </span>
      {pill}
      {onApproveAll ? (
        <Button
          size="xs"
          variant="success"
          aria-label={`Approve all ${count} accounts in ${allocationName ?? allocationId}`}
          onClick={onApproveAll}
        >
          <CheckCheck data-icon="inline-start" />
          Approve all
        </Button>
      ) : null}
    </div>
  );
}
