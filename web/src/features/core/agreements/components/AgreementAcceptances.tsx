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
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { formatDateTime } from "../format";
import { useAgreementAcceptances } from "../queries";
import type { AcceptanceStatus, Actor, AgreementDetail } from "../schemas";

const PAGE_SIZE = 50;

type Row = { user: Actor; at?: string | null };

export function AgreementAcceptances({ agreement }: { agreement: AgreementDetail }) {
  const [status, setStatus] = React.useState<AcceptanceStatus>("accepted");
  const [page, setPage] = React.useState(1);
  const query = useAgreementAcceptances(agreement.id, {
    status,
    limit: PAGE_SIZE,
    offset: (page - 1) * PAGE_SIZE,
  });

  const rows: Row[] =
    status === "accepted"
      ? (query.data?.accepted ?? []).map((r) => ({ user: r.user, at: r.accepted_at }))
      : (query.data?.pending ?? []).map((r) => ({ user: r.user, at: r.first_sign_in_at }));

  const columns: Array<DataTableColumn<Row>> = [
    {
      key: "user",
      header: "User",
      cell: (row) => (
        <div>
          <div className="font-medium">{row.user.name || row.user.id}</div>
          <div className="text-xs text-muted-foreground">{row.user.email}</div>
        </div>
      ),
    },
    {
      key: "at",
      header: status === "accepted" ? "Accepted at" : "First signed in",
      cell: (row) => (row.at ? formatDateTime(row.at) : "Never signed in"),
    },
  ];

  const segments: Array<[AcceptanceStatus, string, number]> = [
    ["accepted", "Accepted", agreement.accepted_count],
    ["pending", "Pending", agreement.pending_count],
  ];

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        {segments.map(([value, label, count]) => (
          <Button
            key={value}
            size="sm"
            variant={status === value ? "default" : "outline"}
            aria-pressed={status === value}
            onClick={() => {
              setStatus(value);
              setPage(1);
            }}
          >
            {label} {count}
          </Button>
        ))}
      </div>
      {query.isLoading ? (
        <CardSkeleton />
      ) : query.error ? (
        <ErrorState message={query.error.message} onRetry={() => query.refetch()} />
      ) : (
        <DataTable
          columns={columns}
          rows={rows}
          rowKey={(row) => row.user.id}
          empty={
            <span className="text-sm text-muted-foreground">
              {status === "accepted"
                ? "No one has accepted this version yet."
                : "No pending users."}
            </span>
          }
          pagination={{
            page,
            pageSize: PAGE_SIZE,
            total: status === "accepted" ? agreement.accepted_count : agreement.pending_count,
            onPageChange: setPage,
          }}
        />
      )}
    </div>
  );
}
