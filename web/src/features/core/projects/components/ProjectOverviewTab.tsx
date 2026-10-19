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

import { Server } from "lucide-react";
import Link from "next/link";
import type * as React from "react";
import { useAllocationsByProject } from "@/features/core/allocations/queries";
import { ClusterName } from "@/features/core/clusters/components/ClusterName";
import type { ProjectResponse } from "@/generated/core/zod.gen";
import { usePermissions } from "@/shared/casl/permissions";
import { StatusIcon } from "@/shared/ui/StatusIcon";
import { ProjectPI } from "./ProjectPI";

export type ProjectOverviewTabProps = {
  project: ProjectResponse;
  addAllocation: React.ReactNode;
};

export function ProjectOverviewTab({ project, addAllocation }: ProjectOverviewTabProps) {
  const can = usePermissions();
  const allocations = useAllocationsByProject(project.id).data?.items ?? [];
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
        <ul className="space-y-0.5">
          {allocations.map((a) => (
            <li key={a.id} className="flex items-center gap-2">
              <StatusIcon icon={Server} active={a.status === "ACTIVE"} label={a.status.toLowerCase()} />
              <Link href={`/allocations/${a.id}`} className="hover:underline">
                {a.name}
              </Link>
              <span className="text-xs text-muted-foreground">
                {can.clustersRead ? (
                  <>
                    <ClusterName id={a.compute_cluster_id} />
                    {" · "}
                  </>
                ) : null}
                {a.initial_su_amount.toLocaleString()} SUs · ends{" "}
                {new Date(a.end_time).toLocaleDateString(undefined, { dateStyle: "medium" })}
              </span>
            </li>
          ))}
          <li className="flex gap-2">
            <span aria-hidden className="w-4 text-center">
              +
            </span>
            {addAllocation}
          </li>
        </ul>
      </dd>
    </dl>
  );
}
