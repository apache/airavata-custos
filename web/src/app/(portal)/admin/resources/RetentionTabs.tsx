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

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CornerDownRight, HardDrive, Pencil, Server, User } from "lucide-react";
import Link from "next/link";
import * as React from "react";
import { toast } from "sonner";
import type { z } from "zod";
import { zGetStoragesResponse, type zStorage } from "@/generated/core/zod.gen";
import { apiFetch } from "@/shared/api/client";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Label } from "@/shared/ui/label";
import { UsageBar } from "@/shared/ui/UsageBar";

type Folder = z.infer<typeof zStorage>;

const LIST = "storage";

// Rows an admin may remove.
const DISPOSABLE = "bg-[color:var(--tone-warn-bg)]/40";

const day = (iso: string) => new Date(iso).toLocaleDateString();

const mono = (text: string) => <span className="font-mono text-xs">{text}</span>;

const inherited = <span className="text-xs text-muted-foreground">Inherited</span>;

const bytes = (n: number) => {
  const i = n > 0 ? Math.min(5, Math.floor(Math.log2(n) / 10)) : 0;
  return `${(n / 1024 ** i).toFixed(i ? 1 : 0)} ${["B", "KiB", "MiB", "GiB", "TiB", "PiB"][i]}`;
};

// An admin's row button that deletes target's path, toasts the outcome and refreshes the list.
function useDeleteColumn<T>(
  list: string,
  variant: "destructive" | "outline",
  label: string,
  target: (row: T) => { path: string; done: string } | false,
): DataTableColumn<T> {
  const client = useQueryClient();
  const canManage = useAbility().can("manage", "Storage");
  const del = useMutation({
    mutationFn: ({ path }: { path: string; done: string }) => apiFetch(path, { method: "DELETE" }),
    onSuccess: (_, { done }) => {
      toast.success(done);
      return client.invalidateQueries({ queryKey: [list] });
    },
    onError: (err) => toast.error(err.message),
  });
  return {
    key: "actions",
    header: "",
    width: "6rem",
    align: "right",
    interactive: true,
    cell: (row) => {
      const t = canManage && target(row);
      return t ? (
        <Button size="xs" variant={variant} disabled={del.isPending} onClick={() => del.mutate(t)}>
          {label}
        </Button>
      ) : null;
    },
  };
}

const GIB = 1024 ** 3;

// Sets a temporary hard limit on a folder's own quota; the change and its history are audited under cluster:path.
function QuotaDialog({ folder, onClose }: { folder: Folder; onClose: () => void }) {
  const client = useQueryClient();
  const [gib, setGib] = React.useState(String(folder.quota_bytes / GIB));
  const [until, setUntil] = React.useState("");
  const [reason, setReason] = React.useState("");
  const latest = new Date(Date.now() + 90 * 864e5).toISOString().slice(0, 10);
  const save = useMutation({
    mutationFn: () =>
      apiFetch(`/storages/${folder.id}`, {
        method: "PUT",
        body: {
          quota_bytes: Math.round(Number(gib) * GIB),
          quota_until: new Date(`${until}T23:59:59`).toISOString(),
          reason,
        },
      }),
    onSuccess: () => {
      toast.success(`${folder.path} quota set until ${until}`);
      onClose();
      return client.invalidateQueries({ queryKey: [LIST] });
    },
    onError: (err) => toast.error(err.message),
  });
  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Quota for {folder.path}</DialogTitle>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            save.mutate();
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="quota-gib">Hard limit (GiB)</Label>
            <Input
              id="quota-gib"
              type="number"
              min="1"
              step="any"
              value={gib}
              onChange={(e) => setGib(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="quota-until">Until, then back to the configured limit</Label>
            <Input
              id="quota-until"
              type="date"
              max={latest}
              value={until}
              onChange={(e) => setUntil(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="quota-reason">Reason</Label>
            <Input id="quota-reason" value={reason} onChange={(e) => setReason(e.target.value)} />
          </div>
          <DialogFooter>
            <Link
              href={`/admin/traces?q=${encodeURIComponent(`${folder.cluster_id}:${folder.path}`)}`}
              className="mr-auto self-center text-sm hover:underline"
            >
              History
            </Link>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={save.isPending}>
              Save
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// Folders a storage connector created. A deleted owner's folder stays here until an admin deletes it.
export function StorageFoldersTab({ clusterId }: { clusterId: string | undefined }) {
  const canManage = useAbility().can("manage", "Storage");
  const [editing, setEditing] = React.useState<Folder | null>(null);
  const del = useDeleteColumn<Folder>(
    LIST,
    "destructive",
    "Delete",
    (row) =>
      row.status === "pending_deletion" && {
        path: `/storages/${row.id}`,
        done: `${row.path} deleted`,
      },
  );
  const path = `/storages?cluster_id=${clusterId}`;
  const query = useQuery({
    queryKey: [LIST, path],
    enabled: Boolean(clusterId),
    queryFn: async () => {
      const items: Folder[] = [];
      for (let total = 1; items.length < total; ) {
        const page = zGetStoragesResponse.parse(
          await apiFetch(`${path}&limit=200&offset=${items.length}`),
        );
        if (!page.items?.length) break;
        items.push(...page.items);
        total = page.total ?? 0;
      }
      return items;
    },
  });
  const columns: Array<DataTableColumn<Folder>> = [
    {
      key: "path",
      header: "Folder",
      cell: (row) => {
        const Icon = row.inherited ? CornerDownRight : HardDrive;
        return (
          <span
            className="flex items-center gap-2"
            title={row.status === "active" ? "Active" : "Owner deleted"}
          >
            <Icon
              aria-hidden
              className={`size-3.5 shrink-0 ${row.status === "active" ? "text-[color:var(--tone-ok-fg)]" : "text-muted-foreground/40"}`}
            />
            {mono(row.path)}
          </span>
        );
      },
    },
    {
      key: "owner",
      header: "Owner",
      width: "16rem",
      interactive: true,
      cell: (row) => (
        <span
          className="flex items-center gap-3 whitespace-nowrap"
          title={`${row.owner}:${row.group}`.replace(/^0:/, "root:")}
        >
          {row.allocation_id ? (
            <Link
              href={`/allocations/${row.allocation_id}`}
              className="flex items-center gap-1 hover:underline"
            >
              <Server aria-hidden className="size-3.5 text-muted-foreground" />
              {row.allocation_name}
            </Link>
          ) : null}
          {row.username ? (
            <span className="flex items-center gap-1">
              <User aria-hidden className="size-3.5 text-muted-foreground" />
              {mono(row.username)}
            </span>
          ) : null}
        </span>
      ),
    },
    {
      key: "validity",
      header: "Validity",
      width: "13rem",
      cell: (row) =>
        row.start_time && row.end_time ? (
          <span className="whitespace-nowrap">
            {day(row.start_time)} – {day(row.end_time)}
          </span>
        ) : (
          "—"
        ),
    },
    {
      key: "quota",
      header: "Quota",
      width: "7rem",
      interactive: true,
      cell: (row) =>
        row.quota_bytes ? (
          <span className="flex items-center gap-1 whitespace-nowrap">
            {bytes(row.quota_bytes)}
            {row.quota_until && new Date(row.quota_until) > new Date() ? (
              <span className="text-xs text-muted-foreground">until {day(row.quota_until)}</span>
            ) : null}
            {canManage ? (
              <button type="button" aria-label="Edit quota" onClick={() => setEditing(row)}>
                <Pencil
                  aria-hidden
                  className="size-3 text-muted-foreground hover:text-foreground"
                />
              </button>
            ) : null}
          </span>
        ) : row.inherited ? (
          inherited
        ) : (
          "—"
        ),
    },
    {
      key: "used",
      header: "Used",
      width: "12rem",
      cell: (row) =>
        row.quota_bytes ? (
          <UsageBar
            size="sm"
            value={row.used_bytes}
            max={row.quota_bytes}
            label={bytes(row.used_bytes)}
            className="min-w-32"
          />
        ) : row.inherited ? (
          inherited
        ) : null,
    },
    del,
  ];
  if (!clusterId) return null;
  if (query.isLoading) return <TableSkeleton rows={6} columns={columns.length} />;
  if (query.error)
    return <ErrorState message={query.error.message} onRetry={() => void query.refetch()} />;
  return (
    <>
      <DataTable
        rows={query.data ?? []}
        columns={columns}
        rowKey={(row) => row.id}
        rowClassName={(row) => (row.status === "active" ? undefined : DISPOSABLE)}
        empty="No folders recorded."
      />
      {editing ? <QuotaDialog folder={editing} onClose={() => setEditing(null)} /> : null}
    </>
  );
}
