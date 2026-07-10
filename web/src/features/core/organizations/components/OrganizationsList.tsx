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

import type * as React from "react";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import type { Organization } from "../schemas";

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
}: OrganizationsListProps) {
  const columns: Array<DataTableColumn<Organization>> = [
    {
      key: "name",
      header: "Name",
      cell: (row) => <span className="font-medium text-foreground">{row.name}</span>,
    },
    {
      key: "originated_id",
      header: "Originated ID",
      cell: (row) => (
        <span className="font-mono text-xs text-muted-foreground">{row.originated_id}</span>
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
          <h1 className="font-display text-[28px] font-bold leading-tight">Organizations</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Institutions and resource providers registered in Custos.
          </p>
        </div>
        {headerCta ? <div>{headerCta}</div> : null}
      </div>

      {isLoading ? (
        <CardSkeleton />
      ) : error ? (
        <ErrorState message={error.message} onRetry={onRetry} />
      ) : rows.length === 0 ? (
        <EmptyState heading="No organizations yet." />
      ) : (
        <DataTable
          columns={columns}
          rows={rows}
          rowKey={(row) => row.id}
          pagination={{ page, pageSize, total, onPageChange }}
        />
      )}
    </div>
  );
}
