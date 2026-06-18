"use client";

import * as React from "react";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { PACKET_TYPES, type Packet, type PacketStatus } from "../types";
import { ageHoursOf, formatDate } from "../utils";
import { PacketStatusBadge } from "./PacketStatusBadge";

export type PacketFilters = {
  status: PacketStatus | "all";
  type: string;
  source: string;
  q: string;
};

export type PacketInboxTableProps = {
  rows: Packet[];
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
  onRowClick: (packet: Packet) => void;
  onBulkRetry: () => void;
  onBulkMarkProcessed: () => void;
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
  onBulkRetry,
  onBulkMarkProcessed,
  onBulkExport,
  onRetry,
}: PacketInboxTableProps) {
  const [searchDraft, setSearchDraft] = React.useState(filters.q);
  React.useEffect(() => setSearchDraft(filters.q), [filters.q]);

  const allSelectable = rows.map((r) => r.id);
  const allSelected = allSelectable.length > 0 && allSelectable.every((id) => selected.has(id));

  function toggleAll() {
    if (allSelected) {
      onSelectChange(new Set());
    } else {
      onSelectChange(new Set(allSelectable));
    }
  }

  function toggleOne(id: string) {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    onSelectChange(next);
  }

  const columns: DataTableColumn<Packet>[] = [
    {
      key: "select",
      header: (
        <input
          type="checkbox"
          aria-label="Select all packets on this page"
          checked={allSelected}
          onChange={toggleAll}
        />
      ),
      width: "32px",
      interactive: true,
      cell: (row) => (
        <input
          type="checkbox"
          aria-label={`Select ${row.amie_id}`}
          checked={selected.has(row.id)}
          onChange={() => toggleOne(row.id)}
          onClick={(e) => e.stopPropagation()}
        />
      ),
    },
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
      key: "source",
      header: "Source",
      sortable: true,
      sortValue: (row) => row.source,
      cell: (row) => <span className="text-xs text-muted-foreground">{row.source}</span>,
    },
    {
      key: "linked",
      header: "Linked entity",
      sortable: true,
      sortValue: (row) =>
        row.linked_entity
          ? `${row.linked_entity.type}:${row.linked_entity.display_id ?? row.linked_entity.id}`
          : null,
      cell: (row) =>
        row.linked_entity ? (
          <span className="text-xs text-muted-foreground">
            {row.linked_entity.type} · {row.linked_entity.display_id ?? row.linked_entity.id}
          </span>
        ) : (
          <span className="text-xs text-muted-foreground">—</span>
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
          onClick={(e) => {
            e.stopPropagation();
            onRowClick(row);
          }}
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
            onChange={(e) =>
              onFiltersChange({
                ...filters,
                status: e.currentTarget.value as PacketStatus | "all",
              })
            }
            className="rounded-md border bg-background px-3 py-1.5 text-sm"
          >
            <option value="all">All</option>
            <option value="NEW">NEW</option>
            <option value="DECODED">DECODED</option>
            <option value="PROCESSED">PROCESSED</option>
            <option value="FAILED">FAILED</option>
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
          <Label htmlFor="amie-source">Source</Label>
          <select
            id="amie-source"
            value={filters.source}
            onChange={(e) => onFiltersChange({ ...filters, source: e.currentTarget.value })}
            className="rounded-md border bg-background px-3 py-1.5 text-sm"
          >
            <option value="all">All sources</option>
            <option value="access">access</option>
          </select>
        </div>

        <div className="flex flex-col gap-1">
          <Label htmlFor="amie-search">Search</Label>
          <Input
            id="amie-search"
            type="search"
            placeholder="amie id, packet id, entity"
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
            onClick={onBulkRetry}
          >
            Retry
          </Button>
          <Button
            type="button"
            variant="outline"
            disabled={selected.size === 0}
            onClick={onBulkMarkProcessed}
          >
            Mark processed
          </Button>
          <Button
            type="button"
            variant="outline"
            disabled={selected.size === 0}
            onClick={onBulkExport}
          >
            Export JSON
          </Button>
        </div>
      </form>

      {error ? (
        <ErrorState message={error.message ?? "Failed to load packets"} onRetry={onRetry} />
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
