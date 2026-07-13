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

import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { useClusterUsers } from "../queries";
import type { ComputeClusterUser } from "../schemas";

export type ClusterUsersDrawerProps = {
  clusterId: string | null;
  clusterName: string | null;
  onOpenChange: (open: boolean) => void;
};

export function ClusterUsersDrawer({
  clusterId,
  clusterName,
  onOpenChange,
}: ClusterUsersDrawerProps) {
  const usersQuery = useClusterUsers(clusterId ?? undefined);
  const users = usersQuery.data ?? [];

  return (
    <SideDrawer
      open={Boolean(clusterId)}
      onOpenChange={onOpenChange}
      title={clusterName ? `Cluster: ${clusterName}` : "Cluster"}
      width="lg"
    >
      {usersQuery.isLoading ? (
        <CardSkeleton />
      ) : usersQuery.error ? (
        <ErrorState
          message={(usersQuery.error as Error).message}
          onRetry={() => usersQuery.refetch()}
        />
      ) : users.length === 0 ? (
        <EmptyState heading="No local accounts on this cluster." />
      ) : (
        <table className="w-full text-left text-sm">
          <thead className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            <tr>
              <th className="py-2 pr-4 font-medium">Local username</th>
              <th className="py-2 font-medium">User ID</th>
            </tr>
          </thead>
          <tbody>
            {users.map((user: ComputeClusterUser) => (
              <tr key={user.id} className="border-t border-border/60">
                <td className="py-2 pr-4 font-mono text-xs text-foreground">
                  {user.local_username}
                </td>
                <td className="py-2 font-mono text-xs text-muted-foreground">{user.user_id}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </SideDrawer>
  );
}
