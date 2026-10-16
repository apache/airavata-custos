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

import { ClusterAccountDetail } from "@/features/core/cluster-accounts/components/ClusterAccountDetail";
import { UserName } from "@/features/core/users/components/UserPicker";
import { isGone } from "@/shared/api/useOnGone";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { setSearchParam, useShallowSearchParams } from "@/shared/hooks/useShallowSearchParams";
import { queryUuid } from "@/shared/layout/nav";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import * as React from "react";
import { useCluster, useClusterUsers } from "../queries";

const ACCOUNTS_REGION = '[aria-label="Local accounts"]';

export type ClusterUsersDrawerProps = {
  clusterId: string | null;
  canWrite: boolean;
  onOpenChange: (open: boolean) => void;
};

// Cluster detail: its local accounts, a lookup by user, and one account opened in place.
export function ClusterUsersDrawer({ clusterId, canWrite, onOpenChange }: ClusterUsersDrawerProps) {
  const clusterQuery = useCluster(clusterId ?? undefined);
  const cluster = clusterQuery.data;
  return (
    <SideDrawer
      open={Boolean(clusterId)}
      onOpenChange={onOpenChange}
      recordKey={clusterId ?? undefined}
      title={cluster?.name ? `Cluster: ${cluster.name}` : "Cluster"}
      description={cluster ? <span className="font-mono text-xs">{cluster.id}</span> : null}
      width="lg"
    >
      {clusterQuery.isLoading ? (
        <CardSkeleton />
      ) : isGone(clusterQuery.error) ? (
        <ErrorState heading="This cluster no longer exists" as="h3" />
      ) : clusterQuery.error ? (
        <ErrorState error={clusterQuery.error} onRetry={() => clusterQuery.refetch()} as="h3" />
      ) : clusterId ? (
        // Keyed, so the lookup and the opened account never carry over to another cluster.
        <ClusterAccounts key={clusterId} clusterId={clusterId} canWrite={canWrite} />
      ) : null}
    </SideDrawer>
  );
}

function ClusterAccounts({ clusterId, canWrite }: { clusterId: string; canWrite: boolean }) {
  // Without users:read every name would read "Unknown user", so the column is omitted.
  const canReadUsers = useAbility().can("read", "User");
  const usersQuery = useClusterUsers(clusterId);
  const [lookupInput, setLookupInput] = React.useState("");
  // undefined until a lookup runs; userId stays unset when the input is not a user ID.
  const [lookup, setLookup] = React.useState<{ userId?: string }>();
  const params = useShallowSearchParams();
  const accountId = queryUuid(params.get("account")) ?? null;
  const users = usersQuery.data ?? [];
  const found = lookup?.userId ? users.find((u) => u.user_id === lookup.userId) : undefined;
  const body = React.useRef<HTMLDivElement>(null);
  // Focus once the view swaps: a selector inside the drawer, else the accounts region.
  const focusNext = React.useRef<string | null>(null);
  const show = (next: string | null, focus: string) => {
    focusNext.current = focus;
    setSearchParam(params, "account", next);
  };

  React.useEffect(() => {
    const target = focusNext.current;
    focusNext.current = null;
    if (!target || !body.current) return;
    (
      body.current.querySelector<HTMLElement>(target) ??
      body.current.querySelector<HTMLElement>(ACCOUNTS_REGION)
    )?.focus();
  });

  return (
    <div ref={body}>
      {accountId ? (
        <div className="space-y-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => show(null, `[data-account-id="${accountId}"]`)}
          >
            Back to accounts
          </Button>
          {usersQuery.isLoading ? (
            <CardSkeleton />
          ) : usersQuery.error ? (
            <ErrorState error={usersQuery.error} onRetry={() => usersQuery.refetch()} as="h3" />
          ) : users.some((u) => u.id === accountId) ? (
            <ClusterAccountDetail
              accountId={accountId}
              canWrite={canWrite}
              onDeleted={() => show(null, ACCOUNTS_REGION)}
            />
          ) : (
            <ErrorState heading="This cluster account no longer exists" as="h3" />
          )}
        </div>
      ) : (
        <section className="space-y-4 outline-none" aria-label="Local accounts" tabIndex={-1}>
          <form
            className="flex flex-wrap items-center gap-2"
            onSubmit={(e) => {
              e.preventDefault();
              setLookup({ userId: queryUuid(lookupInput) });
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
            <Button
              type="submit"
              variant="outline"
              disabled={!usersQuery.data}
              focusableWhenDisabled
            >
              Find account
            </Button>
            {found ? (
              <Button type="button" variant="link" onClick={() => show(found.id, "button")}>
                {found.local_username}
              </Button>
            ) : lookup ? (
              <output className="text-sm text-muted-foreground">
                {lookup.userId
                  ? "No account for that user on this cluster."
                  : "Enter a valid user ID."}
              </output>
            ) : null}
          </form>

          {usersQuery.isLoading ? (
            <CardSkeleton />
          ) : usersQuery.error ? (
            <ErrorState error={usersQuery.error} onRetry={() => usersQuery.refetch()} />
          ) : users.length === 0 ? (
            <EmptyState heading="No local accounts on this cluster" />
          ) : (
            <table className="w-full text-left text-sm">
              <caption className="sr-only">Local accounts</caption>
              <thead className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                <tr>
                  <th className="py-2 pr-4 font-medium">Local username</th>
                  {canReadUsers ? <th className="py-2 font-medium">User</th> : null}
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id} className="border-t border-border/60">
                    <td className="py-2 pr-4 [overflow-wrap:anywhere]">
                      <button
                        type="button"
                        className="text-left font-mono text-xs text-foreground hover:underline"
                        data-account-id={user.id}
                        onClick={() => show(user.id, "button")}
                      >
                        {user.local_username}
                      </button>
                    </td>
                    {canReadUsers ? (
                      <td className="py-2 text-foreground [overflow-wrap:anywhere]">
                        <UserName id={user.user_id} />
                      </td>
                    ) : null}
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </section>
      )}
    </div>
  );
}
