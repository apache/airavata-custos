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
import { useClusterName } from "@/features/core/clusters/queries";
import { UserName } from "@/features/core/users/components/UserPicker";
import { formatDate } from "@/shared/format";
import { Button } from "@/shared/ui/button";
import { QueryErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { CardSkeleton } from "@/shared/ui/Loading";
import { FieldList, Mono } from "@/shared/ui/RecordDrawer";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { useClusterAccount, useDeleteClusterAccount, useUpdateClusterAccount } from "../queries";
import { clusterAccountStatus } from "./ClusterAccountsTable";

export type ClusterAccountDetailProps = {
  accountId: string;
  canWrite: boolean;
  onDeleted: () => void;
};

export function ClusterAccountDetail({
  accountId,
  canWrite,
  onDeleted,
}: ClusterAccountDetailProps) {
  const query = useClusterAccount(accountId);
  const update = useUpdateClusterAccount();
  const remove = useDeleteClusterAccount();
  const clusterName = useClusterName();
  const [username, setUsername] = React.useState("");
  const account = query.data;

  React.useEffect(() => {
    setUsername(account?.local_username ?? "");
  }, [account?.local_username]);

  if (query.isLoading) return <CardSkeleton />;
  if (query.error) {
    return (
      <QueryErrorState error={query.error} what="cluster account" onRetry={() => query.refetch()} />
    );
  }
  if (!account) return null;

  const status = clusterAccountStatus(account);
  const trimmed = username.trim();

  return (
    <div className="space-y-4">
      <FieldList
        fields={{
          Status: <StatusBadge variant={status.variant} label={status.label} />,
          "Local username": <Mono>{account.local_username}</Mono>,
          User: <UserName id={account.user_id} />,
          Cluster: clusterName(account.compute_cluster_id),
          "Access level": account.access_level,
          Reviewed: (
            <>
              {formatDate(account.reviewed_at) || "Not yet"}
              {account.review_note ? (
                <span className="text-muted-foreground"> · {account.review_note}</span>
              ) : null}
            </>
          ),
          Provisioned: formatDate(account.provisioned_at) || "Not yet",
        }}
      />

      {canWrite ? (
        <div className="space-y-3 border-t border-border pt-4">
          <form
            className="flex items-center gap-2"
            onSubmit={(e) => {
              e.preventDefault();
              update.mutate(
                { path: { id: accountId }, body: { local_username: trimmed } },
                toastOnSuccess("Local username updated"),
              );
            }}
          >
            <Input
              aria-label="Local username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-48 font-mono"
              required
            />
            <Button
              type="submit"
              variant="outline"
              disabled={update.isPending || !trimmed || trimmed === account.local_username}
            >
              Rename
            </Button>
          </form>
          <Button
            variant="destructive"
            disabled={remove.isPending}
            onClick={() =>
              confirmToast(
                `Delete the account ${account.local_username}? Custos stops tracking it on the cluster.`,
                "Delete",
                () =>
                  remove.mutate(
                    { path: { id: accountId } },
                    toastOnSuccess("Cluster account deleted", onDeleted),
                  ),
              )
            }
          >
            Delete account
          </Button>
        </div>
      ) : null}
    </div>
  );
}
