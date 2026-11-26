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
import { conflictsOf, useSlurmCluster } from "../queries";
import { useAllocations } from "@/features/core/allocations/queries";
import { Badge } from "@/shared/ui/badge";

// Links each out-of-sync allocation to its page, with its conflict count.
export function ClusterSyncNotice({ clusterId }: { clusterId: string }) {
  const conflicts = useSlurmCluster(clusterId, conflictsOf);
  const allocations = useAllocations();
  const nameOf = (id: string) => allocations.data?.items.find((a) => a.id === id)?.name ?? id;

  const perAllocation = Map.groupBy(
    (conflicts.data ?? []).filter((c) => c.allocation_id),
    (c) => c.allocation_id ?? "",
  );

  if (perAllocation.size === 0) return null;
  return (
    <div className="rounded-md border border-border p-3 text-sm">
      Allocations out of sync:{" "}
      {[...perAllocation].map(([id, list]) => (
        <Badge
          key={id}
          variant="secondary"
          className="mr-2"
          render={<Link href={`/allocations/${id}`} />}
        >
          {nameOf(id)} · {list.length}
        </Badge>
      ))}
    </div>
  );
}
