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

import {
  ClusterAccountDetail,
  useAccountUser,
} from "@/features/core/cluster-accounts/components/ClusterAccountDetail";
import { ClusterAccountsTable } from "@/features/core/cluster-accounts/components/ClusterAccountsTable";
import { CreateClusterAccountDialog } from "@/features/core/cluster-accounts/components/CreateClusterAccountDialog";
import {
  type ReviewAction,
  ReviewClusterAccountDialog,
} from "@/features/core/cluster-accounts/components/ReviewClusterAccountDialog";
import {
  useApproveClusterAccount,
  useClusterAccount,
  useClusterAccounts,
  useDenyClusterAccount,
} from "@/features/core/cluster-accounts/queries";
import type { ClusterAccountApproval, ComputeClusterUser } from "@/generated/core/types.gen";
import { cn } from "@/lib/utils";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { useUrlFilteredPage } from "@/shared/hooks/usePageClamp";
import { setSearchParam, useCanonicalParams } from "@/shared/hooks/useShallowSearchParams";
import { queryUuid } from "@/shared/layout/nav";
import { NotPermitted } from "@/shared/ui/ErrorState";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { Button } from "@/shared/ui/button";
import { toastOnSuccess } from "@/shared/ui/sonner";
import * as React from "react";

const PAGE_SIZE = 25;

type Filter = ClusterAccountApproval | "ALL";

const FILTERS: Array<{ value: Filter; label: string; empty: string }> = [
  { value: "PENDING", label: "Pending approval", empty: "Nothing waiting for approval" },
  { value: "APPROVED", label: "Approved", empty: "No approved accounts yet" },
  { value: "DENIED", label: "Denied", empty: "No denied accounts" },
  { value: "ALL", label: "All", empty: "No cluster accounts yet" },
];

export function ClusterAccountsContainer() {
  const ability = useAbility();
  const canRead = ability.can("read", "Cluster");
  const canReview = ability.can("write", "Cluster");
  const { params, page, setPage, offset, setFilter } = useUrlFilteredPage(PAGE_SIZE);
  const current = FILTERS.find((f) => f.value === params.get("status"));
  const filter = current?.value ?? "PENDING";
  const [review, setReview] = React.useState<{
    account: ComputeClusterUser;
    action: ReviewAction;
  } | null>(null);
  // ?account= holds the open account, so traces can link straight to it.
  const detailId = queryUuid(params.get("account")) ?? null;
  const setDetailId = (id: string | null) => setSearchParam(params, "account", id);
  useCanonicalParams({
    status: filter === "PENDING" ? undefined : filter,
    account: detailId ?? undefined,
  });

  const query = useClusterAccounts(
    {
      approval_status: filter === "ALL" ? undefined : filter,
      limit: PAGE_SIZE,
      offset,
    },
    canRead,
  );
  const approve = useApproveClusterAccount();
  const deny = useDenyClusterAccount();
  const detail = useClusterAccount(canRead && detailId ? detailId : undefined).data;
  const detailUser = useAccountUser(detailId ?? "", detail?.user_id);
  const statusRow = React.useRef<HTMLSpanElement>(null);
  // A cancelled review returns to its Approve or Deny button; a finished one to the Status row.
  const reviewOpener = React.useRef<HTMLElement | null>(null);
  const reviewed = React.useRef(false);
  const startReview = (opener: HTMLElement, account: ComputeClusterUser, action: ReviewAction) => {
    reviewOpener.current = opener;
    reviewed.current = false;
    setReview({ account, action });
  };

  function handleConfirm(note: string) {
    if (!review) return;
    const { id } = review.account;
    const approving = review.action === "approve";
    toastOnSuccess(
      approving
        ? approve.mutateAsync({ path: { id } })
        : deny.mutateAsync({ path: { id }, body: { note } }),
      `Cluster account ${approving ? "approved" : "denied"} for ${review.account.local_username}`,
      () => {
        reviewed.current = true;
        setReview(null);
      },
    );
  }

  if (!canRead) {
    return (
      <div className="space-y-4">
        {canReview ? (
          <div className="flex justify-end">
            <CreateClusterAccountDialog />
          </div>
        ) : null}
        <NotPermitted />
      </div>
    );
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
        onRowClick={(row) => setDetailId(row.id)}
        page={page}
        pageSize={PAGE_SIZE}
        total={query.data?.total ?? 0}
        onPageChange={setPage}
        emptyHeading={current?.empty ?? "Nothing waiting for approval"}
      />

      <SideDrawer
        open={detailId !== null}
        recordKey={detailId ?? undefined}
        onOpenChange={(open) => {
          if (!open) setDetailId(null);
        }}
        title="Cluster account"
      >
        {detailId ? (
          <div className="space-y-4">
            <ClusterAccountDetail
              accountId={detailId}
              canWrite={canReview}
              onDeleted={() => setDetailId(null)}
              statusRef={statusRow}
            />
            {canReview && detail && detail.approval_status !== "APPROVED" ? (
              <div className="flex gap-2">
                {detail.approval_status === "PENDING" ? (
                  <Button
                    size="sm"
                    variant="destructive"
                    onClick={(e) => startReview(e.currentTarget, detail, "deny")}
                  >
                    Deny
                  </Button>
                ) : null}
                <Button
                  size="sm"
                  variant="success"
                  onClick={(e) => startReview(e.currentTarget, detail, "approve")}
                >
                  Approve
                </Button>
              </div>
            ) : null}
            {/* Inside the drawer, so a press in the dialog is not a press outside the drawer. */}
            {canReview ? (
              <ReviewClusterAccountDialog
                account={review?.account.id === detailId ? review.account : null}
                user={detailUser}
                action={review?.action ?? "approve"}
                onOpenChange={(open) => {
                  if (!open) setReview(null);
                }}
                onConfirm={handleConfirm}
                isPending={approve.isPending || deny.isPending}
                finalFocus={() =>
                  !reviewed.current && reviewOpener.current?.isConnected
                    ? reviewOpener.current
                    : statusRow.current
                }
              />
            ) : null}
          </div>
        ) : null}
      </SideDrawer>
    </div>
  );
}
