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
import { toast } from "sonner";
import { ApiError } from "@/shared/api/client";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { useCursorPaging, useMyCertificates, useRevokeMyCertificate } from "../queries";
import type { Certificate } from "../schemas";
import {
  CERTIFICATE_STATUS_LABELS,
  type CertificateStatus,
  formatUnixSeconds,
  getCertificateStatus,
  statusBadgeVariantFromCertificateStatus,
} from "../status";
import { revokeErrorMessage } from "./CertificateDetail";
import { CursorPager } from "./CertificateList";
import { GenerateCertificateDialog } from "./GenerateCertificateDialog";
import { RevokeCertificateDialog } from "./RevokeCertificateDialog";

export function loadErrorMessage(error: unknown): string {
  const message = error instanceof ApiError ? (error.body as { message?: string } | null)?.message : undefined;
  return message || "Your certificates could not be loaded. Please try again.";
}

export function MyCertificates() {
  const { params, paging } = useCursorPaging();
  const query = useMyCertificates(params);
  const revoke = useRevokeMyCertificate();
  const [generateOpen, setGenerateOpen] = React.useState(false);
  const [statusFilter, setStatusFilter] = React.useState<CertificateStatus | "all">("all");
  const [revoking, setRevoking] = React.useState<Certificate | null>(null);
  const [revokeError, setRevokeError] = React.useState<string | null>(null);

  const rows = query.data?.certificates ?? [];
  const filtered = rows.filter(
    (row) => statusFilter === "all" || getCertificateStatus(row) === statusFilter,
  );

  const columns: Array<DataTableColumn<Certificate>> = [
    {
      key: "issued",
      header: "Issued",
      cell: (row) => (
        <Link href={`/ssh-access/${row.serial_number}`} className="text-sm font-medium hover:underline">
          {formatUnixSeconds(row.issued_at)}
        </Link>
      ),
    },
    {
      key: "validBefore",
      header: "Valid until",
      cell: (row) => <span className="text-sm">{formatUnixSeconds(row.valid_before)}</span>,
    },
    {
      key: "status",
      header: "Status",
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
    {
      key: "actions",
      header: "",
      cell: (row) =>
        getCertificateStatus(row) === "active" ? (
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setRevokeError(null);
              setRevoking(row);
            }}
          >
            Revoke
          </Button>
        ) : null,
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">SSH access</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Your SSH certificates for logging in to the cluster.
          </p>
        </div>
        <Button onClick={() => setGenerateOpen(true)}>Generate certificate</Button>
      </div>

      <p className="rounded-md border bg-card p-3 text-sm text-muted-foreground">
        A revoked certificate is still accepted by login nodes until it expires.
      </p>

      <div className="flex items-center gap-3">
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as CertificateStatus | "all")}
          aria-label="Filter by status"
          className="h-9 rounded-md border bg-background px-3 text-sm"
        >
          <option value="all">All statuses</option>
          <option value="active">Active</option>
          <option value="expired">Expired</option>
          <option value="revoked">Revoked</option>
        </select>
        <span className="text-xs text-muted-foreground">Status filter applies to this page.</span>
      </div>

      {query.isLoading ? (
        <TableSkeleton rows={4} columns={5} />
      ) : query.error ? (
        <ErrorState message={loadErrorMessage(query.error)} onRetry={() => query.refetch()} />
      ) : filtered.length === 0 ? (
        <EmptyState
          heading={rows.length === 0 ? "No certificates yet" : "No matches on this page"}
          description={
            rows.length === 0
              ? "Generate a certificate to log in to the cluster."
              : "No certificates on this page have that status."
          }
        />
      ) : (
        <DataTable columns={columns} rows={filtered} rowKey={(row) => String(row.serial_number)} />
      )}
      {!query.isLoading && (query.error || rows.length > 0) ? (
        <CursorPager {...paging(query.data?.next_cursor)} />
      ) : null}

      <GenerateCertificateDialog open={generateOpen} onOpenChange={setGenerateOpen} />
      <RevokeCertificateDialog
        serialNumber={revoking?.serial_number ?? 0}
        open={revoking !== null}
        onOpenChange={(open) => {
          if (!open) setRevoking(null);
        }}
        isPending={revoke.isPending}
        error={revokeError}
        onSubmit={(reason) => {
          if (!revoking) return;
          setRevokeError(null);
          revoke.mutate(
            { serial: revoking.serial_number, reason },
            {
              onSuccess: () => {
                setRevoking(null);
                toast.success("Certificate revoked");
              },
              onError: (err) => setRevokeError(revokeErrorMessage(err)),
            },
          );
        }}
      />
    </div>
  );
}
