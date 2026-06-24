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

import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { useAllocationResources } from "../queries";
import type { ComputeAllocation } from "../schemas";

export type AllocationOverviewTabProps = {
  allocation: ComputeAllocation;
};

function formatNumber(n: number): string {
  return new Intl.NumberFormat().format(n);
}

export function AllocationOverviewTab({ allocation }: AllocationOverviewTabProps) {
  const resourcesQuery = useAllocationResources(allocation.id);

  return (
    <div className="space-y-6">
      <dl className="grid gap-x-8 gap-y-3 sm:grid-cols-[max-content_1fr] text-sm">
        <dt className="text-muted-foreground">Allocation ID</dt>
        <dd className="font-mono text-foreground">{allocation.id}</dd>

        <dt className="text-muted-foreground">Project</dt>
        <dd className="font-mono text-foreground">{allocation.project_id}</dd>

        <dt className="text-muted-foreground">Name</dt>
        <dd className="text-foreground">{allocation.name}</dd>

        <dt className="text-muted-foreground">Status</dt>
        <dd className="text-foreground">{allocation.status}</dd>

        <dt className="text-muted-foreground">Cluster</dt>
        <dd className="font-mono text-foreground">{allocation.compute_cluster_id}</dd>

        <dt className="text-muted-foreground">Initial SUs</dt>
        <dd className="tabular-nums text-foreground">{formatNumber(allocation.initial_su_amount)}</dd>

        <dt className="text-muted-foreground">Start</dt>
        <dd className="text-foreground">{allocation.start_time}</dd>

        <dt className="text-muted-foreground">End</dt>
        <dd className="text-foreground">{allocation.end_time}</dd>
      </dl>

      <section className="space-y-2">
        <h2 className="text-sm font-semibold text-foreground">Resources</h2>
        {resourcesQuery.isLoading ? (
          <TableSkeleton rows={2} columns={3} />
        ) : resourcesQuery.error ? (
          <ErrorState
            message={(resourcesQuery.error as Error).message}
            onRetry={() => resourcesQuery.refetch()}
          />
        ) : !resourcesQuery.data || resourcesQuery.data.length === 0 ? (
          <p className="text-sm text-muted-foreground">No resources attached.</p>
        ) : (
          <ul className="space-y-1">
            {resourcesQuery.data.map((r) => (
              <li
                key={r.id}
                className="flex items-center justify-between rounded-md border bg-card px-3 py-2 text-sm"
              >
                <div>
                  <span className="font-medium">{r.name}</span>
                  <span className="ml-2 text-xs uppercase tracking-wide text-muted-foreground">
                    {r.resource_type}
                  </span>
                </div>
                <span className="tabular-nums">{formatNumber(r.resource_amount)}</span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
