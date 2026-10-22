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

import { ClusterAccountsTable } from "@/features/core/cluster-accounts/components/ClusterAccountsTable";
import {
  type ReviewAction,
  ReviewClusterAccountDialog,
} from "@/features/core/cluster-accounts/components/ReviewClusterAccountDialog";
import {
  useApproveClusterAccount,
  useApproveClusterAccounts,
  useClusterAccounts,
  useDenyClusterAccount,
} from "@/features/core/cluster-accounts/queries";
import type { ClusterAccountApproval, ClusterAccountResponse } from "@/generated/core/zod.gen";
import { cn } from "@/lib/utils";
import { ApiError } from "@/shared/api/client";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { useDebounce } from "@/shared/hooks/useDebounce";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Search } from "lucide-react";
import * as React from "react";
import { toast } from "sonner";

const PAGE_SIZE = 25;
// Pending is not paged so Approve all covers a whole allocation.
const PENDING_LIMIT = 200;

type Filter = ClusterAccountApproval | "ALL";

type Review = { accounts: ClusterAccountResponse[]; action: ReviewAction; bulk: boolean };

const FILTERS: Array<{ value: Filter; label: string; empty: string }> = [
  { value: "PENDING", label: "Pending approval", empty: "Nothing waiting for approval." },
  { value: "APPROVED", label: "Approved", empty: "No approved accounts yet." },
  { value: "DENIED", label: "Denied", empty: "No denied accounts." },
  { value: "ALL", label: "All", empty: "No cluster accounts yet." },
];

function readFilter(params: URLSearchParams): Filter {
  const raw = params.get("status");
  return FILTERS.find((f) => f.value === raw)?.value ?? "PENDING";
}

export function AccountApprovalsContainer() {
  const ability = useAbility();
  const canRead = ability.can("read", "Cluster");
  const canReview = ability.can("manage", "Cluster");
  const params = useShallowSearchParams();
  const filter = readFilter(params);
  const [page, setPage] = React.useState(1);
  const [search, setSearch] = React.useState(() => params.get("q") ?? "");
  const q = useDebounce(search.trim(), 300);
  const [review, setReview] = React.useState<Review | null>(null);
  const [reviewError, setReviewError] = React.useState<string | null>(null);
  const [failures, setFailures] = React.useState<Record<string, string>>({});

  React.useEffect(() => {
    const next = new URLSearchParams(window.location.search);
    if (q) next.set("q", q);
    else next.delete("q");
    replaceShallowSearchParams(next);
  }, [q]);

  const grouped = filter === "PENDING";
  const query = useClusterAccounts(
    {
      ...(filter === "ALL" ? {} : { approval_status: filter }),
      ...(q ? { q } : {}),
      ...(grouped
        ? { limit: PENDING_LIMIT }
        : { limit: PAGE_SIZE, offset: (page - 1) * PAGE_SIZE }),
    },
    { enabled: canRead },
  );
  const approve = useApproveClusterAccount();
  const approveMany = useApproveClusterAccounts();
  const deny = useDenyClusterAccount();

  function setFilter(next: Filter) {
    const nextParams = new URLSearchParams(params);
    if (next === "PENDING") nextParams.delete("status");
    else nextParams.set("status", next);
    replaceShallowSearchParams(nextParams);
    setPage(1);
  }

  function openReview(action: ReviewAction) {
    return (account: ClusterAccountResponse) => {
      setReviewError(null);
      setReview({ accounts: [account], action, bulk: false });
    };
  }

  function openApproveAll(accounts: ClusterAccountResponse[]) {
    setReviewError(null);
    setFailures({});
    setReview({ accounts, action: "approve", bulk: true });
  }

  async function handleApproveAll(accounts: ClusterAccountResponse[]) {
    const results = await approveMany.mutateAsync(accounts.map((a) => a.id));
    const failed: Record<string, string> = {};
    results.forEach((result, i) => {
      const account = accounts[i];
      if (result.status === "fulfilled" || !account) return;
      // 409: already approved by someone else.
      if (result.reason instanceof ApiError && result.reason.status === 409) return;
      failed[account.id] =
        result.reason instanceof Error ? result.reason.message : "Approval failed.";
    });
    const failedAccounts = accounts.filter((a) => failed[a.id]);
    if (failedAccounts.length === 0) {
      const n = accounts.length;
      toast.success(
        `Approved ${n} cluster account${n === 1 ? "" : "s"} on ${accounts[0]?.cluster_name}.`,
      );
      setReview(null);
      return;
    }
    toast.error(
      `Approved ${accounts.length - failedAccounts.length} of ${accounts.length}. ${failedAccounts.length} failed.`,
    );
    setFailures(failed);
    setReview({ accounts: failedAccounts, action: "approve", bulk: true });
  }

  async function handleConfirm(note: string) {
    if (!review) return;
    if (review.bulk) return handleApproveAll(review.accounts);
    const [account] = review.accounts;
    if (!account) return;
    setReviewError(null);
    try {
      if (review.action === "approve") {
        await approve.mutateAsync(account.id);
        toast.success(
          `Cluster account approved for ${account.local_username} on ${account.cluster_name}`,
        );
      } else {
        await deny.mutateAsync({ id: account.id, note });
        toast.success(
          `Cluster account denied for ${account.local_username} on ${account.cluster_name}`,
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
  const rows = query.data?.items ?? [];
  const total = query.data?.total ?? 0;
  const noMatch = q !== "" && !query.isLoading && !query.error && rows.length === 0;

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

      <div className="relative max-w-lg">
        <Search
          aria-hidden
          className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
        />
        <Input
          type="search"
          placeholder="Search by name, email, username, allocation or project"
          aria-label="Search requests"
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            setPage(1);
          }}
          className="h-9 pl-9"
        />
      </div>

      {noMatch ? (
        <EmptyState
          heading={`No requests match "${q}".`}
          cta={
            <Button variant="outline" size="sm" onClick={() => setSearch("")}>
              Clear search
            </Button>
          }
        />
      ) : (
        <>
          {grouped && total > PENDING_LIMIT ? (
            <p className="text-sm text-muted-foreground">
              Showing {PENDING_LIMIT} of {total}. Approve some to see the rest.
            </p>
          ) : null}
          <ClusterAccountsTable
            rows={rows}
            isLoading={query.isLoading}
            error={query.error}
            onRetry={() => void query.refetch()}
            canReview={canReview}
            onApprove={openReview("approve")}
            onDeny={openReview("deny")}
            onApproveAll={openApproveAll}
            grouped={grouped}
            page={page}
            pageSize={PAGE_SIZE}
            total={total}
            onPageChange={setPage}
            emptyHeading={emptyHeading}
          />
        </>
      )}

      {canReview ? (
        <ReviewClusterAccountDialog
          accounts={review?.accounts ?? []}
          action={review?.action ?? "approve"}
          bulk={review?.bulk ?? false}
          onOpenChange={(open) => {
            if (!open) setReview(null);
          }}
          onConfirm={handleConfirm}
          isPending={approve.isPending || approveMany.isPending || deny.isPending}
          error={reviewError}
          failures={failures}
        />
      ) : null}
    </div>
  );
}
