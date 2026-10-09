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
import { formatNumber } from "@/shared/format";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { Mono, RecordDrawer } from "@/shared/ui/RecordDrawer";
import { StatusBadge, statusBadgeVariantFromAllocationStatus } from "@/shared/ui/StatusBadge";
import { useResource, useResourceAllocations, useResourceOverrides } from "../queries";

export type ResourceDetailDrawerProps = {
  resourceId: string | null;
  clusterName: (id?: string) => string;
  onOpenChange: (open: boolean) => void;
};

export function ResourceDetailDrawer({
  resourceId,
  clusterName,
  onOpenChange,
}: ResourceDetailDrawerProps) {
  const id = resourceId ?? undefined;
  const resource = useResource(id);
  const allocations = useResourceAllocations(id);
  const overrides = useResourceOverrides(id);

  return (
    <RecordDrawer
      title={resource.data?.name ?? "Resource"}
      id={id}
      onClose={() => onOpenChange(false)}
      query={resource}
      width="lg"
      fields={(r) => ({
        "Resource ID": <Mono>{r.id}</Mono>,
        Type: r.resource_type,
        Amount: <span className="tabular-nums">{formatNumber(r.resource_amount)}</span>,
        Cluster: clusterName(r.compute_cluster_id),
      })}
    >
      {() => (
        <>
          <section className="space-y-2">
            <h3 className="text-sm font-semibold text-foreground">Allocations</h3>
            {allocations.isLoading ? (
              <CardSkeleton />
            ) : allocations.error ? (
              <ErrorState
                message={allocations.error.message}
                onRetry={() => allocations.refetch()}
              />
            ) : !allocations.data?.length ? (
              <p className="text-sm text-muted-foreground">No allocations use this resource.</p>
            ) : (
              <ul className="space-y-1">
                {allocations.data.map((a) => (
                  <li
                    key={a.id}
                    className="flex items-center justify-between rounded-md border bg-card px-3 py-2 text-sm"
                  >
                    <Link
                      href={`/allocations/${a.id}`}
                      className="text-brand underline-offset-4 hover:underline"
                    >
                      {a.name}
                    </Link>
                    <StatusBadge
                      variant={statusBadgeVariantFromAllocationStatus(a.status)}
                      label={a.status}
                    />
                  </li>
                ))}
              </ul>
            )}
          </section>

          <section className="space-y-2">
            <h3 className="text-sm font-semibold text-foreground">Membership overrides</h3>
            {overrides.isLoading ? (
              <CardSkeleton />
            ) : overrides.error ? (
              <ErrorState message={overrides.error.message} onRetry={() => overrides.refetch()} />
            ) : !overrides.data?.length ? (
              <p className="text-sm text-muted-foreground">
                No membership overrides on this resource.
              </p>
            ) : (
              <table className="w-full text-left text-sm">
                <thead className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  <tr>
                    <th className="py-2 pr-4 font-medium">Membership ID</th>
                    <th className="py-2 pr-4 text-right font-medium">Amount</th>
                    <th className="py-2 text-right font-medium">Time (min)</th>
                  </tr>
                </thead>
                <tbody>
                  {overrides.data.map((o) => (
                    <tr key={o.id} className="border-t border-border/60">
                      <td className="py-2 pr-4 font-mono text-xs">
                        {o.compute_allocation_membership_id}
                      </td>
                      <td className="py-2 pr-4 text-right tabular-nums">
                        {formatNumber(o.override_resource_amount)}
                      </td>
                      <td className="py-2 text-right tabular-nums">
                        {formatNumber(o.override_resource_time)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </section>
        </>
      )}
    </RecordDrawer>
  );
}
