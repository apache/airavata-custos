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

import type { GetConnectorsAmiePacketsData, PacketResponse } from "@/generated/amie/types.gen";
import * as React from "react";
import { zPacketResponse } from "@/generated/amie/zod.gen";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { ageHoursOf, formatDate, packetStatusLabel } from "../utils";
import { PacketStatusBadge } from "./PacketStatusBadge";
import { PacketActionButtons, type PacketActions, selectColumn } from "./packetActions";

const PACKET_TYPES = [
  "request_project_create",
  "request_project_inactivate",
  "request_project_reactivate",
  "request_account_create",
  "request_account_inactivate",
  "request_account_reactivate",
  "request_person_merge",
  "request_user_modify",
  "data_account_create",
  "data_project_create",
  "inform_transaction_complete",
];

export type PacketFilters = Required<
  Pick<NonNullable<GetConnectorsAmiePacketsData["query"]>, "status" | "type" | "q">
>;

export type PacketInboxTableProps = PacketActions & {
  rows: PacketResponse[];
  total: number;
  isLoading: boolean;
  error: Error | null;
  page: number;
  pageSize: number;
  filters: PacketFilters;
  selected: Set<string>;
  onSelectChange: (selected: Set<string>) => void;
  onFiltersChange: (filters: PacketFilters) => void;
  onPageChange: (page: number) => void;
  onRowClick: (packet: PacketResponse) => void;
  onBulkExport: () => void;
  onRetry: () => void;
};

export function PacketInboxTable({
  rows,
  total,
  isLoading,
  error,
  page,
  pageSize,
  filters,
  selected,
  onSelectChange,
  onFiltersChange,
  onPageChange,
  onRowClick,
  onBulkExport,
  onRetry,
  ...actions
}: PacketInboxTableProps) {
  const [searchDraft, setSearchDraft] = React.useState(filters.q);
  React.useEffect(() => setSearchDraft(filters.q), [filters.q]);

  const columns: DataTableColumn<PacketResponse>[] = [
    selectColumn(rows, selected, onSelectChange),
    {
      key: "received",
      header: "Received",
      sortable: true,
      sortValue: (row) => new Date(row.received_at),
      cell: (row) => (
        <span className="text-xs text-muted-foreground tabular-nums">
          {formatDate(row.received_at)}
        </span>
      ),
    },
    {
      key: "amie_id",
      header: "AMIE ID",
      sortable: true,
      sortValue: (row) => row.amie_id,
      cell: (row) => <span className="font-mono text-sm">{row.amie_id}</span>,
    },
    {
      key: "type",
      header: "Type",
      sortable: true,
      sortValue: (row) => row.type,
      cell: (row) => <span className="text-sm">{row.type}</span>,
    },
    {
      key: "status",
      header: "Status",
      sortable: true,
      sortValue: (row) => row.status,
      cell: (row) => (
        <PacketStatusBadge status={row.status} ageHours={ageHoursOf(row.received_at)} />
      ),
    },
    {
      key: "updated",
      header: "Last updated",
      sortable: true,
      sortValue: (row) => new Date(row.updated_at),
      cell: (row) => (
        <span className="text-xs text-muted-foreground tabular-nums">
          {formatDate(row.updated_at)}
        </span>
      ),
    },
    {
      key: "view",
      header: "",
      align: "right",
      interactive: true,
      cell: (row) => (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={() => onRowClick(row)}
        >
          View
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-4">
      <form
        className="flex flex-wrap items-end gap-3 rounded-md border bg-card p-4"
        onSubmit={(e) => {
          e.preventDefault();
          onFiltersChange({ ...filters, q: searchDraft });
        }}
      >
        <div className="flex flex-col gap-1">
          <Label htmlFor="amie-status">Status</Label>
          <select
            id="amie-status"
            value={filters.status}
            onChange={(e) => onFiltersChange({ ...filters, status: e.currentTarget.value })}
            className="rounded-md border bg-background px-3 py-1.5 text-sm"
          >
            <option value="all">All</option>
            {zPacketResponse.shape.status.options.map((s) => (
              <option key={s} value={s}>
                {packetStatusLabel(s)}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-1">
          <Label htmlFor="amie-type">Type</Label>
          <select
            id="amie-type"
            value={filters.type}
            onChange={(e) => onFiltersChange({ ...filters, type: e.currentTarget.value })}
            className="rounded-md border bg-background px-3 py-1.5 text-sm"
          >
            <option value="all">All types</option>
            {PACKET_TYPES.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-1">
          <Label htmlFor="amie-search">Search</Label>
          <Input
            id="amie-search"
            type="search"
            placeholder="amie id, packet id"
            value={searchDraft}
            onChange={(e) => setSearchDraft(e.currentTarget.value)}
            className="w-56"
          />
        </div>

        <Button type="submit" variant="outline">
          Apply
        </Button>

        <div className="ml-auto flex items-center gap-2" role="toolbar" aria-label="Bulk actions">
          <span className="text-xs text-muted-foreground">{selected.size} selected</span>
          <Button
            type="button"
            variant="outline"
            disabled={selected.size === 0}
            onClick={onBulkExport}
          >
            Export JSON
          </Button>
          <PacketActionButtons ids={[...selected]} bulk {...actions} />
        </div>
      </form>

      {error ? (
        <ErrorState message={error.message} onRetry={onRetry} />
      ) : isLoading ? (
        <TableSkeleton />
      ) : rows.length === 0 ? (
        <EmptyState
          heading="No packets in this view"
          description="Try clearing filters or widening the time window."
        />
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
