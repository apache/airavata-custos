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
import * as React from "react";
import { DataTable, type DataTableColumn, type DataTablePagination } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import type { Certificate } from "../schemas";
import {
  CERTIFICATE_STATUS_LABELS,
  type CertificateStatus,
  formatUnixDate,
  getCertificateStatus,
  statusBadgeVariantFromCertificateStatus,
} from "../status";

export type CertificateListProps = {
  rows: Certificate[];
  isLoading: boolean;
  error: Error | null;
  onRetry?: () => void;
  search: string;
  onSearchChange: (next: string) => void;
  statusFilter: CertificateStatus | "all";
  onStatusFilterChange: (next: CertificateStatus | "all") => void;
  pagination?: DataTablePagination;
};

export function CertificateList({
  rows,
  isLoading,
  error,
  onRetry,
  search,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  pagination,
}: CertificateListProps) {
  // Status/search narrow only the current server page — the signer API paginates
  // by limit/offset and has no status filter, so deeper filtering is client-side.
  const filtered = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    return rows.filter((row) => {
      if (statusFilter !== "all" && getCertificateStatus(row) !== statusFilter) return false;
      if (needle) {
        const hay = `${row.serial_number} ${row.principal} ${row.key_id}`.toLowerCase();
        if (!hay.includes(needle)) return false;
      }
      return true;
    });
  }, [rows, search, statusFilter]);

  const columns: Array<DataTableColumn<Certificate>> = [
    {
      key: "serial",
      header: "Serial",
      sortable: true,
      sortValue: (row) => row.serial_number,
      cell: (row) => (
        <Link
          href={`/signer/certificates/${row.serial_number}`}
          className="font-mono font-medium text-foreground hover:underline"
        >
          {row.serial_number}
        </Link>
      ),
    },
    {
      key: "principal",
      header: "Principal",
      sortable: true,
      sortValue: (row) => row.principal,
      cell: (row) => <span className="text-sm">{row.principal}</span>,
    },
    {
      key: "issued",
      header: "Issued",
      sortable: true,
      sortValue: (row) => row.issued_at,
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{formatUnixDate(row.issued_at)}</span>
      ),
    },
    {
      key: "validBefore",
      header: "Valid until",
      sortable: true,
      sortValue: (row) => row.valid_before,
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{formatUnixDate(row.valid_before)}</span>
      ),
    },
    {
      key: "status",
      header: "Status",
      sortable: true,
      sortValue: (row) => getCertificateStatus(row),
      cell: (row) => {
        const status = getCertificateStatus(row);
        return (
          <StatusBadge
            variant={statusBadgeVariantFromCertificateStatus(status)}
            label={CERTIFICATE_STATUS_LABELS[status]}
          />
        );
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-display text-[28px] font-bold leading-tight">SSH Certificates</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          SSH certificates issued to you by the certificate signer.
        </p>
      </div>

      <div className="flex flex-col gap-3 rounded-md border bg-card p-4 sm:flex-row sm:items-center">
        <Input
          type="search"
          placeholder="Search by serial, principal, or key ID"
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          aria-label="Search certificates"
          className="sm:w-72"
        />
        <select
          value={statusFilter}
          onChange={(e) => onStatusFilterChange(e.target.value as CertificateStatus | "all")}
          aria-label="Filter by status"
          className="h-9 rounded-md border bg-background px-3 text-sm"
        >
          <option value="all">All statuses</option>
          <option value="active">Active</option>
          <option value="expired">Expired</option>
          <option value="revoked">Revoked</option>
        </select>
      </div>

      {isLoading ? (
        <TableSkeleton rows={6} columns={5} />
      ) : error ? (
        <ErrorState message={error.message} onRetry={onRetry} />
      ) : filtered.length === 0 ? (
        <EmptyState
          heading="No certificates yet"
          description="No certificates match the current filters."
        />
      ) : (
        <DataTable
          columns={columns}
          rows={filtered}
          rowKey={(row) => String(row.serial_number)}
          pagination={pagination}
        />
      )}
    </div>
  );
}
