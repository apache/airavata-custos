"use client";

import * as React from "react";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import {
  ProjectAutocomplete,
  type ProjectAutocompletePick,
} from "@/shared/ui/ProjectAutocomplete";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import type { Packet } from "../types";
import { formatDate } from "../utils";

type Row = Packet;

export type ReconciliationQueueProps = {
  rows: Packet[];
  total: number;
  isLoading: boolean;
  error: Error | null;
  searchProjects: (q: string) => Promise<ProjectAutocompletePick[]>;
  onLink: (packet: Packet, entity_type: string, entity_id: string) => void;
  onSkip: (packet: Packet, reason: string) => void;
  onRefresh: () => void;
};

export function ReconciliationQueue({
  rows,
  total,
  isLoading,
  error,
  searchProjects,
  onLink,
  onSkip,
  onRefresh,
}: ReconciliationQueueProps) {
  const [picks, setPicks] = React.useState<Record<string, ProjectAutocompletePick | null>>({});
  const [skipDraft, setSkipDraft] = React.useState<Record<string, string>>({});

  const columns: DataTableColumn<Row>[] = [
    {
      key: "amie_id",
      header: "AMIE ID",
      cell: (r) => <span className="font-mono text-sm">{r.amie_id}</span>,
    },
    {
      key: "type",
      header: "Type",
      cell: (r) => <span className="text-sm">{r.type}</span>,
    },
    {
      key: "received",
      header: "Received",
      cell: (r) => (
        <span className="text-xs tabular-nums text-muted-foreground">
          {formatDate(r.received_at)}
        </span>
      ),
    },
    {
      key: "link",
      header: "Link to existing project",
      cell: (r) => {
        const pick = picks[r.id] ?? null;
        return (
          <form
            className="flex items-end gap-2"
            onSubmit={(e) => {
              e.preventDefault();
              if (!pick) return;
              onLink(r, "project", pick.originated_id);
            }}
          >
            <ProjectAutocomplete
              value={pick}
              onChange={(next) => setPicks((prev) => ({ ...prev, [r.id]: next }))}
              search={searchProjects}
              ariaLabel={`Search projects to link ${r.amie_id}`}
            />
            <Button type="submit" variant="outline" size="sm" disabled={!pick}>
              Link
            </Button>
          </form>
        );
      },
    },
    {
      key: "skip",
      header: "Or skip…",
      align: "right",
      cell: (r) => (
        <form
          className="flex items-end justify-end gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            const reason = (skipDraft[r.id] ?? "").trim();
            if (reason.length < 3) return;
            onSkip(r, reason);
            setSkipDraft((prev) => ({ ...prev, [r.id]: "" }));
          }}
        >
          <Input
            type="text"
            aria-label={`Skip reason for ${r.amie_id}`}
            placeholder="reason (≥3 chars)"
            value={skipDraft[r.id] ?? ""}
            onChange={(e) => setSkipDraft((prev) => ({ ...prev, [r.id]: e.currentTarget.value }))}
            className="w-44"
          />
          <Button
            type="submit"
            variant="ghost"
            size="sm"
            disabled={(skipDraft[r.id] ?? "").trim().length < 3}
          >
            Skip
          </Button>
        </form>
      ),
    },
  ];

  return (
    <div className="space-y-4">
      <div className="rounded-md border bg-card p-4">
        <div className="flex items-end justify-between gap-3">
          <p className="text-xs text-muted-foreground">
            Decoded packets that couldn't be auto-linked to a domain entity.
          </p>
          <Label className="text-xs text-muted-foreground">{total} unmapped</Label>
        </div>
      </div>

      {error ? (
        <ErrorState message={error.message ?? "Failed to load queue"} onRetry={onRefresh} />
      ) : isLoading ? (
        <TableSkeleton />
      ) : rows.length === 0 ? (
        <EmptyState
          heading="Inbox clean"
          description="Every decoded packet is mapped to a domain entity."
        />
      ) : (
        <DataTable columns={columns} rows={rows} rowKey={(r) => r.id} />
      )}
    </div>
  );
}
