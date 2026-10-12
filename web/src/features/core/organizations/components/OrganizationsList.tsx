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

import type { Organization } from "@/generated/core/types.gen";
import { ROUTE_LABELS } from "@/shared/layout/nav";
import { DataTable, type DataTableColumn, NameCell, PastLastPage } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState, NotPermitted } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import type * as React from "react";

export type OrganizationsListProps = {
  rows: Organization[];
  isLoading: boolean;
  error: Error | null;
  onRetry?: () => void;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  headerCta?: React.ReactNode;
  onRowClick?: (row: Organization) => void;
  // A write-only caller sees the header and its create control, not the list.
  notPermitted?: boolean;
};

export function OrganizationsList({
  rows,
  isLoading,
  error,
  onRetry,
  page,
  pageSize,
  total,
  onPageChange,
  headerCta,
  onRowClick,
  notPermitted = false,
}: OrganizationsListProps) {
  const columns: Array<DataTableColumn<Organization>> = [
    {
      key: "name",
      header: "Name",
      cell: (row) => <NameCell text={row.name} className="font-medium text-foreground" />,
    },
    {
      key: "originated_id",
      header: "Originated ID",
      cell: (row) => (
        <NameCell
          text={row.originated_id || "—"}
          className="max-w-[24ch] font-mono text-xs text-muted-foreground"
        />
      ),
    },
    {
      key: "id",
      header: "ID",
      cell: (row) => <span className="font-mono text-xs text-muted-foreground">{row.id}</span>,
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">
            {ROUTE_LABELS["/admin/organizations"]}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Institutions and resource providers registered in Custos.
          </p>
        </div>
        {headerCta ? <div>{headerCta}</div> : null}
      </div>

      {notPermitted ? (
        <NotPermitted />
      ) : isLoading ? (
        <CardSkeleton />
      ) : error ? (
        <ErrorState error={error} onRetry={onRetry} />
      ) : rows.length === 0 && page > 1 ? (
        <PastLastPage things="organizations" page={page} onPageChange={onPageChange} />
      ) : rows.length === 0 ? (
        <EmptyState heading="No organizations yet" />
      ) : (
        <DataTable
          caption="Organizations"
          columns={columns}
          rows={rows}
          rowKey={(row) => row.id}
          onRowClick={onRowClick}
          pagination={{ page, pageSize, total, onPageChange }}
        />
      )}
    </div>
  );
}
