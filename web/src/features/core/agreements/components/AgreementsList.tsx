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

import Link from "next/link";
import type * as React from "react";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { formatDate } from "../format";
import type { AgreementListItem } from "../schemas";
import { AgreementStatusBadge } from "./AgreementStatusBadge";

export type AgreementsListProps = {
  rows: AgreementListItem[];
  isLoading: boolean;
  error: Error | null;
  onRetry?: () => void;
  headerCta?: React.ReactNode;
};

const columns: Array<DataTableColumn<AgreementListItem>> = [
  {
    key: "version",
    header: "Version",
    cell: (row) => (
      <Link
        href={`/admin/agreements/${row.id}`}
        className="font-medium text-brand underline-offset-4 hover:underline"
      >
        {row.version_label}
      </Link>
    ),
  },
  { key: "title", header: "Title", cell: (row) => row.title },
  { key: "status", header: "Status", cell: (row) => <AgreementStatusBadge status={row.status} /> },
  {
    key: "published",
    header: "Published",
    cell: (row) => (row.published_at ? formatDate(row.published_at) : "Not yet"),
  },
  {
    key: "accepted",
    header: "Accepted",
    align: "right",
    cell: (row) => (row.status === "DRAFT" ? "—" : row.accepted_count),
  },
  {
    key: "pending",
    header: "Pending",
    align: "right",
    // The backend counts pending users only for the current version.
    cell: (row) =>
      row.status !== "PUBLISHED" ? (
        "—"
      ) : row.pending_count > 0 ? (
        <StatusBadge variant="warning" label={String(row.pending_count)} />
      ) : (
        0
      ),
  },
];

export function AgreementsList({
  rows,
  isLoading,
  error,
  onRetry,
  headerCta,
}: AgreementsListProps) {
  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">Agreements</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            The user agreement every user must accept before using the portal.
          </p>
        </div>
        {headerCta ? <div>{headerCta}</div> : null}
      </div>

      {isLoading ? (
        <CardSkeleton />
      ) : error ? (
        <ErrorState message={error.message} onRetry={onRetry} />
      ) : rows.length === 0 ? (
        <EmptyState
          heading="No agreement versions yet."
          description="Users are not asked to accept anything until a version is published."
        />
      ) : (
        <DataTable columns={columns} rows={rows} rowKey={(row) => row.id} />
      )}
    </div>
  );
}
