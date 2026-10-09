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
import { ClusterAccountsTable } from "@/features/core/cluster-accounts/components/ClusterAccountsTable";
import {
  type ReviewAction,
  ReviewClusterAccountDialog,
} from "@/features/core/cluster-accounts/components/ReviewClusterAccountDialog";
import { CreateClusterAccountDialog } from "@/features/core/cluster-accounts/components/CreateClusterAccountDialog";
import {
  useApproveClusterAccount,
  useClusterAccounts,
  useDenyClusterAccount,
} from "@/features/core/cluster-accounts/queries";
import type { ClusterAccountApproval, ClusterAccountResponse } from "@/generated/core/types.gen";
import { cn } from "@/lib/utils";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { usePageClamp, useUrlFilteredPage } from "@/shared/hooks/usePageClamp";
import { ErrorState } from "@/shared/ui/ErrorState";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { toastOnSuccess } from "@/shared/ui/sonner";

const PAGE_SIZE = 25;

type Filter = ClusterAccountApproval | "ALL";

const FILTERS: Array<{ value: Filter; label: string; empty: string }> = [
  { value: "PENDING", label: "Pending approval", empty: "Nothing waiting for approval." },
  { value: "APPROVED", label: "Approved", empty: "No approved accounts yet." },
  { value: "DENIED", label: "Denied", empty: "No denied accounts." },
  { value: "ALL", label: "All", empty: "No cluster accounts yet." },
];

export function ClusterAccountsContainer() {
  const ability = useAbility();
  const canRead = ability.can("read", "Cluster");
  const canReview = ability.can("write", "Cluster");
  const { params, page, setPage, offset, setFilter } = useUrlFilteredPage(PAGE_SIZE);
  const current = FILTERS.find((f) => f.value === params.get("status"));
  const filter = current?.value ?? "PENDING";
  const [review, setReview] = React.useState<{
    account: ClusterAccountResponse;
    action: ReviewAction;
  } | null>(null);
  // ?account= opens an account, so traces can link straight to it.
  const [detailId, setDetailId] = React.useState<string | null>(params.get("account"));

  const query = useClusterAccounts(
    {
      approval_status: filter === "ALL" ? undefined : filter,
      limit: PAGE_SIZE,
      offset,
    },
    canRead,
  );
  usePageClamp(page, setPage, query.data?.total, PAGE_SIZE);
  const approve = useApproveClusterAccount();
  const deny = useDenyClusterAccount();

  const openReview = (action: ReviewAction) => (account: ClusterAccountResponse) =>
    setReview({ account, action });

  function handleConfirm(note: string) {
    if (!review) return;
    const { id, local_username, cluster_name } = review.account;
    const approving = review.action === "approve";
    const done = toastOnSuccess(
      `Cluster account ${approving ? "approved" : "denied"} for ${local_username} on ${cluster_name}`,
      () => setReview(null),
    );
    if (approving) approve.mutate({ path: { id } }, done);
    else deny.mutate({ path: { id }, body: { note } }, done);
  }

  if (!canRead) {
    return <ErrorState heading="Not permitted" message="You cannot view cluster accounts." />;
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <fieldset
          aria-label="Filter by status"
          className="inline-flex overflow-hidden rounded-lg border border-border"
        >
          {FILTERS.map((f) => (
            <button
              key={f.value}
              type="button"
              aria-pressed={filter === f.value}
              onClick={() => setFilter("status", f.value === "PENDING" ? null : f.value)}
              className={cn(
                "px-3 py-1.5 text-sm font-medium transition-colors",
                filter === f.value
                  ? "bg-muted text-foreground"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              {f.label}
            </button>
          ))}
        </fieldset>
        <div className="flex items-center gap-3">
          <p className="text-sm text-muted-foreground">
            An account is created on the cluster only after an admin approves it.
          </p>
          {canReview ? <CreateClusterAccountDialog /> : null}
        </div>
      </div>

      <ClusterAccountsTable
        rows={query.data?.items ?? []}
        isLoading={query.isLoading}
        error={query.error}
        onRetry={() => void query.refetch()}
        canReview={canReview}
        onApprove={openReview("approve")}
        onDeny={openReview("deny")}
        onRowClick={(row) => setDetailId(row.id)}
        page={page}
        pageSize={PAGE_SIZE}
        total={query.data?.total ?? 0}
        onPageChange={setPage}
        emptyHeading={current?.empty ?? "Nothing waiting for approval."}
      />

      <SideDrawer
        open={detailId !== null}
        onOpenChange={(open) => {
          if (!open) setDetailId(null);
        }}
        title="Cluster account"
      >
        {detailId ? (
          <ClusterAccountDetail
            accountId={detailId}
            canWrite={canReview}
            onDeleted={() => setDetailId(null)}
          />
        ) : null}
      </SideDrawer>

      {canReview ? (
        <ReviewClusterAccountDialog
          account={review?.account ?? null}
          action={review?.action ?? "approve"}
          onOpenChange={(open) => {
            if (!open) setReview(null);
          }}
          onConfirm={handleConfirm}
          isPending={approve.isPending || deny.isPending}
        />
      ) : null}
    </div>
  );
}
