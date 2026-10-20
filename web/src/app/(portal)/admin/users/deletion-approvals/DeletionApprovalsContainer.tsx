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

import { toast } from "sonner";
import type { DeletionReview } from "@/features/core/deletion-requests/api";
import { useDeletionRequests, useReviewDeletion } from "@/features/core/deletion-requests/queries";
import type { DeletionRequest } from "@/generated/core/zod.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";

const KINDS: Record<string, string> = {
  compute_cluster_user: "Cluster account",
  compute_allocation: "Allocation",
};

export function DeletionApprovalsContainer() {
  const ability = useAbility();
  const canReview = ability.can("manage", "Cluster");
  const query = useDeletionRequests(canReview);
  const review = useReviewDeletion();

  function handle(row: DeletionRequest, action: DeletionReview) {
    review.mutate(
      { row, action },
      {
        onSuccess: () =>
          toast.success(`Deletion of ${row.name} ${action === "approve" ? "approved" : "denied"}`),
        onError: (err) => toast.error(err.message),
      },
    );
  }

  if (!canReview) {
    return <ErrorState heading="Not permitted" message="You cannot review deletions." />;
  }

  const columns: Array<DataTableColumn<DeletionRequest>> = [
    { key: "kind", header: "Kind", cell: (row) => KINDS[row.entity_type] ?? row.entity_type },
    {
      key: "name",
      header: "Name",
      cell: (row) => <span className="font-mono text-xs">{row.name}</span>,
    },
    {
      key: "cluster",
      header: "Cluster",
      cell: (row) => <span className="font-mono text-xs">{row.cluster_name}</span>,
    },
    {
      key: "requested",
      header: "Requested",
      cell: (row) => (
        <span className="text-xs text-muted-foreground">
          {new Date(row.requested_at).toLocaleString()}
        </span>
      ),
    },
    {
      key: "review",
      header: "Review",
      align: "right",
      interactive: true,
      cell: (row) => (
        <div className="flex justify-end gap-1">
          <Button
            size="xs"
            variant="success"
            disabled={review.isPending}
            onClick={() => handle(row, "approve")}
          >
            Approve
          </Button>
          <Button
            size="xs"
            variant="outline"
            disabled={review.isPending}
            onClick={() => handle(row, "deny")}
          >
            Deny
          </Button>
        </div>
      ),
    },
  ];

  if (query.isLoading) return <TableSkeleton rows={6} columns={columns.length} />;
  if (query.error)
    return <ErrorState message={query.error.message} onRetry={() => void query.refetch()} />;

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        Approving removes the account or allocation from the cluster. Files are kept.
      </p>
      <DataTable
        columns={columns}
        rows={query.data ?? []}
        rowKey={(row) => `${row.entity_type}/${row.entity_id}`}
        empty="No deletions waiting for approval."
      />
    </div>
  );
}
