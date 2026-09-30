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
import { toast } from "sonner";
import { ClusterAccountsTable } from "@/features/core/cluster-accounts/components/ClusterAccountsTable";
import {
  type ReviewAction,
  ReviewClusterAccountDialog,
} from "@/features/core/cluster-accounts/components/ReviewClusterAccountDialog";
import {
  useApproveClusterAccount,
  useClusterAccounts,
  useDenyClusterAccount,
} from "@/features/core/cluster-accounts/queries";
import type { ClusterAccount, ClusterAccountApproval } from "@/features/core/cluster-accounts/schemas";
import { cn } from "@/lib/utils";
import { useAbility } from "@/shared/casl/AbilityProvider";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { ErrorState } from "@/shared/ui/ErrorState";

const PAGE_SIZE = 25;

type Filter = ClusterAccountApproval | "ALL";

const FILTERS: Array<{ value: Filter; label: string; empty: string }> = [
  { value: "PENDING", label: "Pending approval", empty: "Nothing waiting for approval." },
  { value: "APPROVED", label: "Approved", empty: "No approved accounts yet." },
  { value: "DENIED", label: "Denied", empty: "No denied accounts." },
  { value: "ALL", label: "All", empty: "No cluster accounts yet." },
];

function readFilter(params: URLSearchParams): Filter {
  const raw = params.get("status");
  return FILTERS.some((f) => f.value === raw) ? (raw as Filter) : "PENDING";
}

export function ClusterAccountsContainer() {
  const ability = useAbility();
  const canRead = ability.can("read", "Cluster");
  const canReview = ability.can("manage", "Cluster");
  const params = useShallowSearchParams();
  const filter = readFilter(params);
  const [page, setPage] = React.useState(1);
  const [review, setReview] = React.useState<{ account: ClusterAccount; action: ReviewAction } | null>(
    null,
  );
  const [reviewError, setReviewError] = React.useState<string | null>(null);

  const query = useClusterAccounts(
    {
      approval_status: filter === "ALL" ? undefined : filter,
      limit: PAGE_SIZE,
      offset: (page - 1) * PAGE_SIZE,
    },
    { enabled: canRead },
  );
  const approve = useApproveClusterAccount();
  const deny = useDenyClusterAccount();

  function setFilter(next: Filter) {
    const nextParams = new URLSearchParams(params);
    if (next === "PENDING") nextParams.delete("status");
    else nextParams.set("status", next);
    replaceShallowSearchParams(nextParams);
    setPage(1);
  }

  function openReview(action: ReviewAction) {
    return (account: ClusterAccount) => {
      setReviewError(null);
      setReview({ account, action });
    };
  }

  async function handleConfirm(note: string) {
    if (!review) return;
    setReviewError(null);
    try {
      if (review.action === "approve") {
        await approve.mutateAsync(review.account.id);
        toast.success(
          `Cluster account approved for ${review.account.local_username} on ${review.account.cluster_name}`,
        );
      } else {
        await deny.mutateAsync({ id: review.account.id, note });
        toast.success(
          `Cluster account denied for ${review.account.local_username} on ${review.account.cluster_name}`,
        );
      }
      setReview(null);
    } catch (err) {
      setReviewError(err instanceof Error ? err.message : "The review did not go through.");
    }
  }

  if (!canRead) {
    return <ErrorState heading="Not permitted" message="You cannot view cluster accounts." />;
  }

  const emptyHeading =
    FILTERS.find((f) => f.value === filter)?.empty ?? "Nothing waiting for approval.";

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
              onClick={() => setFilter(f.value)}
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
        <p className="text-sm text-muted-foreground">
          An account is created on the cluster only after an admin approves it.
        </p>
      </div>

      <ClusterAccountsTable
        rows={query.data?.items ?? []}
        isLoading={query.isLoading}
        error={(query.error as Error | null) ?? null}
        onRetry={() => void query.refetch()}
        canReview={canReview}
        onApprove={openReview("approve")}
        onDeny={openReview("deny")}
        page={page}
        pageSize={PAGE_SIZE}
        total={query.data?.total ?? 0}
        onPageChange={setPage}
        emptyHeading={emptyHeading}
      />

      {canReview ? (
        <ReviewClusterAccountDialog
          account={review?.account ?? null}
          action={review?.action ?? "approve"}
          onOpenChange={(open) => {
            if (!open) setReview(null);
          }}
          onConfirm={handleConfirm}
          isPending={approve.isPending || deny.isPending}
          error={reviewError}
        />
      ) : null}
    </div>
  );
}
