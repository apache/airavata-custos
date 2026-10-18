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

import Link from "next/link";
import { ClusterName } from "@/features/core/clusters/components/ClusterName";
import { ProjectPI } from "@/features/core/projects/components/ProjectPI";
import { useProject } from "@/features/core/projects/queries";
import { useEffectiveRate } from "@/features/core/resources/queries";
import { usePermissions } from "@/shared/casl/permissions";
import { ErrorState } from "@/shared/ui/ErrorState";
import { useAllocationResources } from "../queries";
import type { ComputeAllocation, ComputeAllocationResource } from "@/generated/core/zod.gen";

export type AllocationOverviewTabProps = {
  allocation: ComputeAllocation;
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

function formatNumber(n: number): string {
  return new Intl.NumberFormat().format(n);
}

function AllocationResourceRow({ resource }: { resource: ComputeAllocationResource }) {
  const rateQuery = useEffectiveRate(resource.id);
  return (
    <li className="flex gap-2">
      <span aria-hidden className="w-3 text-center">
        •
      </span>
      <span>
        {resource.name}
        <span className="ml-2 text-xs text-muted-foreground">
          {resource.resource_type}
          {rateQuery.data ? ` · ${formatNumber(rateQuery.data.rate)} SU/unit` : ""}
        </span>
      </span>
    </li>
  );
}

export function AllocationOverviewTab({ allocation }: AllocationOverviewTabProps) {
  const resourcesQuery = useAllocationResources(allocation.id);
  const project = useProject(allocation.project_id).data;
  const can = usePermissions();

  return (
    <div className="space-y-6">
      <dl className="grid gap-x-8 gap-y-3 rounded-lg border border-border bg-muted/40 p-4 text-sm sm:grid-cols-[max-content_1fr]">
        {project ? (
          <>
            <dt className="text-muted-foreground">Project</dt>
            <dd className="text-foreground">
              <Link href={`/projects/${project.id}`} className="hover:underline">
                {project.title}
              </Link>
            </dd>

            <dt className="text-muted-foreground">PI</dt>
            <dd className="text-foreground">
              <ProjectPI project={project} />
            </dd>
          </>
        ) : (
          <>
            <dt className="text-muted-foreground">Project ID</dt>
            <dd className="font-mono text-foreground">{allocation.project_id}</dd>
          </>
        )}

        {can.clustersRead ? (
          <>
            <dt className="text-muted-foreground">Cluster</dt>
            <dd className="text-foreground">
              <ClusterName id={allocation.compute_cluster_id} />
            </dd>
          </>
        ) : (
          <>
            <dt className="text-muted-foreground">Cluster ID</dt>
            <dd className="font-mono text-foreground">{allocation.compute_cluster_id}</dd>
          </>
        )}

        <dt className="text-muted-foreground">Start</dt>
        <dd className="text-foreground">{formatDate(allocation.start_time)}</dd>

        <dt className="text-muted-foreground">End</dt>
        <dd className="text-foreground">{formatDate(allocation.end_time)}</dd>

        <dt className="text-muted-foreground">Group provisioned</dt>
        <dd className="text-foreground">
          {allocation.group_provisioned_at ? formatDate(allocation.group_provisioned_at) : "Pending first member"}
        </dd>

        <dt className="text-muted-foreground">Resources</dt>
        <dd className="text-foreground">
          {resourcesQuery.error ? (
            <ErrorState message={resourcesQuery.error.message} onRetry={() => resourcesQuery.refetch()} />
          ) : (
            <ul className="space-y-0.5">
              {(resourcesQuery.data ?? []).map((resource) => (
                <AllocationResourceRow key={resource.id} resource={resource} />
              ))}
              {resourcesQuery.data?.length === 0 ? (
                <li className="text-muted-foreground">No resources attached.</li>
              ) : null}
            </ul>
          )}
        </dd>
      </dl>
    </div>
  );
}
