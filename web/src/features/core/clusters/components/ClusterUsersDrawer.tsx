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

import * as React from "react";
import { ClusterAccountDetail } from "@/features/core/cluster-accounts/components/ClusterAccountDetail";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { CardSkeleton } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { useCluster, useClusterUser, useClusterUsers } from "../queries";

export type ClusterUsersDrawerProps = {
  clusterId: string | null;
  canWrite: boolean;
  onOpenChange: (open: boolean) => void;
};

// Cluster detail: its local accounts, a lookup by user, and one account opened in place.
export function ClusterUsersDrawer({ clusterId, canWrite, onOpenChange }: ClusterUsersDrawerProps) {
  const clusterQuery = useCluster(clusterId ?? undefined);
  const usersQuery = useClusterUsers(clusterId ?? undefined);
  const [lookupInput, setLookupInput] = React.useState("");
  const [lookupUserId, setLookupUserId] = React.useState<string>();
  const lookup = useClusterUser(clusterId ?? undefined, lookupUserId);
  const [accountId, setAccountId] = React.useState<string | null>(null);
  const users = usersQuery.data ?? [];
  const cluster = clusterQuery.data;
  const found = lookup.data;

  React.useEffect(() => {
    if (!clusterId) {
      setAccountId(null);
      setLookupInput("");
      setLookupUserId(undefined);
    }
  }, [clusterId]);

  return (
    <SideDrawer
      open={Boolean(clusterId)}
      onOpenChange={onOpenChange}
      title={cluster?.name ? `Cluster: ${cluster.name}` : "Cluster"}
      description={cluster?.id ? <span className="font-mono text-xs">{cluster.id}</span> : null}
      width="lg"
    >
      {accountId ? (
        <div className="space-y-4">
          <Button variant="outline" size="sm" onClick={() => setAccountId(null)}>
            Back to accounts
          </Button>
          <ClusterAccountDetail
            accountId={accountId}
            canWrite={canWrite}
            onDeleted={() => setAccountId(null)}
          />
        </div>
      ) : (
        <div className="space-y-4">
          <form
            className="flex flex-wrap items-center gap-2"
            onSubmit={(e) => {
              e.preventDefault();
              setLookupUserId(lookupInput.trim() || undefined);
            }}
          >
            <Input
              aria-label="Find account by user ID"
              placeholder="User ID"
              value={lookupInput}
              onChange={(e) => setLookupInput(e.target.value)}
              className="w-56"
              required
            />
            <Button type="submit" variant="outline" disabled={lookup.isFetching}>
              Find account
            </Button>
            {lookup.error ? (
              <span role="alert" className="text-sm text-muted-foreground">
                No account for that user on this cluster.
              </span>
            ) : found ? (
              <Button type="button" variant="link" onClick={() => setAccountId(found.id)}>
                {found.local_username}
              </Button>
            ) : null}
          </form>

          {usersQuery.isLoading ? (
            <CardSkeleton />
          ) : usersQuery.error ? (
            <ErrorState message={usersQuery.error.message} onRetry={() => usersQuery.refetch()} />
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
                {users.map((user) => (
                  <tr key={user.id} className="border-t border-border/60">
                    <td className="py-2 pr-4">
                      <button
                        type="button"
                        className="font-mono text-xs text-foreground hover:underline"
                        onClick={() => setAccountId(user.id)}
                      >
                        {user.local_username}
                      </button>
                    </td>
                    <td className="py-2 font-mono text-xs text-muted-foreground">{user.user_id}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </SideDrawer>
  );
}
