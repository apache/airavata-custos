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

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  type ClusterView,
  type Conflict,
  getClusterView,
  type ResolveRequest,
  resolveConflict,
} from "./api";

export const slurmKeys = {
  all: ["slurm"] as const,
  cluster: (clusterId: string) => [...slurmKeys.all, "cluster", clusterId] as const,
};

// Slurm is live state: never cache it, and a 503 is an answer, not a flake.
const LIVE = { staleTime: 0, retry: false, refetchOnWindowFocus: false } as const;

// One read per cluster; each consumer selects its part.
export function useSlurmCluster<T>(clusterId: string, select: (view: ClusterView) => T) {
  return useQuery({
    queryKey: slurmKeys.cluster(clusterId),
    queryFn: () => getClusterView(clusterId),
    select,
    ...LIVE,
  });
}

export const conflictsOf = (view: ClusterView) => view.conflicts ?? [];

export const MEMBER_KINDS: Conflict["kind"][] = [
  "member_qos",
  "member_limit",
  "partition_association",
  "access",
];

// An allocation's share of its cluster's conflicts; sync is unknown unless the last read succeeded.
export function allocationConflicts(
  query: { data?: Conflict[] | undefined; isSuccess: boolean },
  allocationId: string,
) {
  const own = (query.data ?? []).filter((c) => c.allocation_id === allocationId);
  return {
    syncUnknown: !query.isSuccess,
    // A null membership picks the conflicts no Custos membership matches.
    ofKinds: (kinds: Conflict["kind"][], membershipId?: string | null) =>
      own.filter(
        (c) =>
          kinds.includes(c.kind) &&
          (membershipId === undefined || c.membership_id === (membershipId ?? undefined)),
      ),
  };
}

// Refetches on success and on a stale 409; a resolve may also write Custos QoS and allocations.
export function useResolveConflict(clusterId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: ResolveRequest) => resolveConflict(clusterId, body),
    onSettled: () =>
      Promise.all(
        [slurmKeys.all, ["clusters"], ["allocations"]].map((queryKey) =>
          queryClient.invalidateQueries({ queryKey }),
        ),
      ),
  });
}
