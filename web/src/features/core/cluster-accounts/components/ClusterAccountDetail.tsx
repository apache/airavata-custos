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

import { Timestamp } from "@/features/core/audit/components/primitives/Timestamp";
import { ClusterName } from "@/features/core/clusters/components/ClusterName";
import { UserName } from "@/features/core/users/components/UserPicker";
import { isGone, useOnGone } from "@/shared/api/useOnGone";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { enumLabel, formatDate } from "@/shared/format";
import { QueryErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { FieldList, Mono } from "@/shared/ui/RecordDrawer";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { Button } from "@/shared/ui/button";
import { focusIsLost } from "@/shared/ui/focus";
import { Input } from "@/shared/ui/input";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { useQueryClient } from "@tanstack/react-query";
import * as React from "react";
import {
  clusterAccountListKeys,
  useClusterAccount,
  useDeleteClusterAccount,
  useListedAccountName,
  useUpdateClusterAccount,
} from "../queries";
import { clusterAccountStatus } from "./ClusterAccountsTable";

// The account's user by name, or undefined when this caller could only say "Unknown user".
export function useAccountUser(accountId: string, userId: string | undefined): React.ReactNode {
  const listed = useListedAccountName(accountId);
  const canReadUsers = useAbility().can("read", "User");
  return listed ?? (canReadUsers ? <UserName id={userId} /> : undefined);
}

export type ClusterAccountDetailProps = {
  accountId: string;
  canWrite: boolean;
  onDeleted: () => void;
  // A review moves focus here once its buttons are gone.
  statusRef?: React.Ref<HTMLSpanElement>;
};

export function ClusterAccountDetail({
  accountId,
  canWrite,
  onDeleted,
  statusRef,
}: ClusterAccountDetailProps) {
  const query = useClusterAccount(accountId);
  const update = useUpdateClusterAccount();
  const remove = useDeleteClusterAccount();
  const [username, setUsername] = React.useState("");
  const usernameInput = React.useRef<HTMLInputElement>(null);
  const renameButton = React.useRef<HTMLButtonElement>(null);
  const account = query.data;
  const user = useAccountUser(accountId, account?.user_id);
  const client = useQueryClient();
  useOnGone(isGone(query.error), () => {
    for (const queryKey of clusterAccountListKeys()) void client.invalidateQueries({ queryKey });
  });

  React.useEffect(() => {
    setUsername(account?.local_username ?? "");
  }, [account?.local_username]);

  if (query.isLoading) return <CardSkeleton />;
  if (query.error) {
    return (
      <QueryErrorState
        error={query.error}
        what="cluster account"
        onRetry={() => query.refetch()}
        as="h3"
      />
    );
  }
  if (!account) return null;

  const status = clusterAccountStatus(account);
  const trimmed = username.trim();

  return (
    <div className="space-y-4">
      <FieldList
        fields={{
          Status: (
            <span ref={statusRef} tabIndex={-1}>
              <StatusBadge variant={status.variant} label={status.label} />
            </span>
          ),
          "Local username": <Mono>{account.local_username}</Mono>,
          ...(user ? { User: user } : {}),
          Cluster: <ClusterName id={account.compute_cluster_id} />,
          "Access level": enumLabel(account.access_level),
          Reviewed: (
            <>
              {account.reviewed_at ? (
                <Timestamp iso={account.reviewed_at}>{formatDate(account.reviewed_at)}</Timestamp>
              ) : (
                "Not yet"
              )}
              {account.review_note ? (
                <span className="text-muted-foreground"> · {account.review_note}</span>
              ) : null}
            </>
          ),
          Provisioned: account.provisioned_at ? (
            <Timestamp iso={account.provisioned_at}>{formatDate(account.provisioned_at)}</Timestamp>
          ) : (
            "Not yet"
          ),
        }}
      />

      {canWrite ? (
        <div className="space-y-3 border-t border-border pt-4">
          <form
            className="flex items-center gap-2"
            onSubmit={(e) => {
              e.preventDefault();
              // PUT replaces the mapping, so the fresh cluster and user go with the new name.
              const { compute_cluster_id, user_id } = account;
              toastOnSuccess(
                update.mutateAsync({
                  path: { id: accountId },
                  body: { compute_cluster_id, user_id, local_username: trimmed },
                }),
                "Local username updated",
                // Rename is unavailable once the name matches; focus moves on only from it.
                () => {
                  if (focusIsLost() || document.activeElement === renameButton.current) {
                    usernameInput.current?.focus();
                  }
                },
              );
            }}
          >
            <Input
              ref={usernameInput}
              aria-label="Local username"
              maxLength={32}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-48 font-mono"
              required
              title="Enter a username, not only spaces."
            />
            <Button
              ref={renameButton}
              type="submit"
              variant="outline"
              pending={update.isPending}
              disabled={!trimmed || trimmed === account.local_username}
              focusableWhenDisabled
            >
              Rename
            </Button>
          </form>
          <Button
            variant="destructive"
            pending={remove.isPending}
            onClick={() =>
              confirmToast(
                `Delete the account ${account.local_username}? Custos stops tracking it on the cluster.`,
                "Delete",
                () =>
                  toastOnSuccess(
                    remove.mutateAsync({ path: { id: accountId } }),
                    "Cluster account deleted",
                    onDeleted,
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
